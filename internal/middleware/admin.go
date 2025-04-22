package middleware

import (
	"e-meetingproject/internal/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminOnlyMiddleware ensures that only users with admin role can access the protected routes
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get claims from the context (set by AuthMiddleware)
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: no claims found"})
			c.Abort()
			return
		}

		// Type assert claims to *auth.Claims
		userClaims, ok := claims.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error: invalid claims type"})
			c.Abort()
			return
		}

		// Check if user has admin role
		if userClaims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
