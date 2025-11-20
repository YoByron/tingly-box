package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// authenticateMiddleware validates JWT tokens
func (s *Server) authenticateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := s.jwtManager.ValidateToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Add client ID to context
		c.Set("client_id", claims.ClientID)
		c.Next()
	}
}