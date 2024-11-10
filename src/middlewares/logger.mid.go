package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("duration", duration),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("[INSCRIPTION-API] Error en la solicitud", append(fields, zap.Error(e.Err))...)
			}
		} else {
			logger.Info("[INSCRIPTION-API] Solicitud procesada", fields...)
		}
	}
}
