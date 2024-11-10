// middlewares/api_key_auth.go
package middlewares

import (
	"net/http"
	"search-courses-api/src/config/envs"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func APIKeyAuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	KEY := envs.LoadEnvs(".env").Get("INSCRIPTION_API_KEY")
	return func(c *gin.Context) {
		apiKey := c.GetHeader("Authorization")

		if apiKey != KEY {
			logger.Warn("API Key inválida", zap.String("ip", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
