package middlewares

import (
	"search-courses-api/src/errors"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if customErr, ok := err.(*errors.Error); ok {
				logger.Error("Error de aplicación", zap.Error(customErr))
				c.JSON(customErr.HTTPStatusCode, gin.H{"error": customErr.Message})
			} else {
				logger.Error("Error no manejado", zap.Error(err))
				c.JSON(errors.ErrInternalServer.HTTPStatusCode, gin.H{"error": errors.ErrInternalServer.Message})
			}
		}
	}
}
