package middlewares

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type LogEntry struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Latency   string `json:"latency"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()
		latency := endTime.Sub(startTime)

		logEntry := LogEntry{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    c.Writer.Status(),
			Latency:   latency.String(),
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}

		log.Printf("%+v", logEntry)
	}
}
