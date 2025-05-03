package main

import (
	"context"
	"e-meetingproject/internal/database"
	"e-meetingproject/internal/handlers"
	"e-meetingproject/internal/middleware"
	"e-meetingproject/internal/services"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func initConfig() error {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	return nil
}

func gracefulShutdown(server *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")
	done <- true
}

func main() {
	// Parse command line flags
	seedOnly := flag.Bool("seed-only", false, "Run database seeder and exit")
	flag.Parse()

	// Initialize configuration
	if err := initConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection using environment variables
	err := database.InitDB(
		viper.GetString("BLUEPRINT_DB_HOST"),
		viper.GetString("BLUEPRINT_DB_PORT"),
		viper.GetString("BLUEPRINT_DB_USERNAME"),
		viper.GetString("BLUEPRINT_DB_PASSWORD"),
		viper.GetString("BLUEPRINT_DB_DATABASE"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run seeders
	if err := database.SeedUsers(); err != nil {
		log.Printf("Warning: Failed to seed users: %v", err)
	}
	if err := database.SeedRooms(); err != nil {
		log.Printf("Warning: Failed to seed rooms: %v", err)
	}

	// If seed-only flag is set, exit after seeding
	if *seedOnly {
		log.Println("Database seeding completed. Exiting...")
		os.Exit(0)
	}

	// Initialize services and handlers
	authService := services.NewAuthService()
	authHandler := handlers.NewAuthHandler(authService)

	userService := services.NewUserService()
	userHandler := handlers.NewUserHandler(userService)

	dashboardService := services.NewDashboardService()
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	reservationService := services.NewReservationService()
	reservationHandler := handlers.NewReservationHandler(reservationService)

	roomService := services.NewRoomService()
	roomHandler := handlers.NewRoomHandler(roomService)

	snackService := services.NewSnackService()
	snackHandler := handlers.NewSnackHandler(snackService)

	// Setup Gin router
	router := gin.Default()

	// Public routes
	router.POST("/register", authHandler.Register)
	router.POST("/password/reset_request", authHandler.RequestPasswordReset)
	router.POST("/password/reset", authHandler.ResetPassword)

	// Regular user login
	router.POST("/login", authHandler.Login)

	// Protected routes (requires authentication)
	protected := router.Group("")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/users/:id", userHandler.GetProfile)
		protected.POST("/users/:id", userHandler.UpdateProfile)
		protected.GET("/dashboard", dashboardHandler.GetDashboardStats)
		protected.GET("/rooms", roomHandler.GetRooms)
		protected.GET("/rooms/:id/schedule", roomHandler.GetRoomSchedule)
		protected.GET("/snacks", snackHandler.GetSnacks)
		protected.POST("/reservation/calculation", reservationHandler.CalculateReservationCost)
		protected.POST("/reservation", reservationHandler.CreateReservation)
		protected.GET("/reservation/history", reservationHandler.GetReservationHistory)
		protected.GET("/reservation/:id", reservationHandler.GetReservationByID)
	}

	// Admin routes group
	adminRoutes := router.Group("/admin")
	{
		// Admin login - public admin route
		adminRoutes.POST("/login", authHandler.Login)

		// Protected admin routes - requires admin role
		adminProtected := adminRoutes.Group("")
		adminProtected.Use(middleware.JWTAuthMiddleware())
		adminProtected.Use(middleware.AdminOnlyMiddleware())
		{
			// Dashboard routes
			adminProtected.GET("/dashboard", dashboardHandler.GetDashboardStats)

			// Reservation management
			adminProtected.GET("/reservations/history", reservationHandler.GetReservationHistory)
			adminProtected.POST("/reservation/status", reservationHandler.UpdateReservationStatus)

			// Room management
			adminProtected.POST("/rooms", roomHandler.CreateRoom)       // Create room
			adminProtected.PUT("/rooms/:id", roomHandler.UpdateRoom)    // Update room
			adminProtected.DELETE("/rooms/:id", roomHandler.DeleteRoom) // Delete room

			// Snack management
			adminProtected.POST("/snacks", snackHandler.CreateSnack) // Create snack
		}
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", viper.GetString("PORT")),
		Handler: router,
	}

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
