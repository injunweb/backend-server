package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/injunweb/backend-server/internal/global/security"
)

func AuthMiddleware(requireRole security.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
			c.Abort()
			return
		}

		id, roles, err := security.ValidateToken(token)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		if !roles.HasRole(requireRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		security.SetContext(&security.SecurityContext{ID: id})

		c.Next()
	}
}
