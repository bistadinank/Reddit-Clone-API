package middleware

import (
	"net/http"
	"reddit/messages"
	"strings"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/gin-gonic/gin"
)

func NewAuthMiddleware(system *actor.ActorSystem, enginePID *actor.PID) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]
		
		// Validate token through actor system
		msg := &messages.ValidateToken{
			Token: token,
		}

		response, err := system.Root.RequestFuture(enginePID, msg, 5*time.Second).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token validation failed"})
			c.Abort()
			return
		}

		if validateResponse, ok := response.(*messages.ValidateTokenResponse); ok {
			if !validateResponse.Success {
				c.JSON(http.StatusUnauthorized, gin.H{"error": validateResponse.Error})
				c.Abort()
				return
			}
			
			// Store validated username in context
			c.Set("username", validateResponse.Username)
			c.Next()
		}
	}
} 