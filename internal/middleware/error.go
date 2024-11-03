package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/injunweb/backend-server/pkg/errors"
)

type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			if customErr, ok := err.(errors.CustomError); ok {
				c.JSON(customErr.GetStatus(), ErrorResponse{
					Type:    customErr.GetType(),
					Message: customErr.GetMessage(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Type:    "Internal",
				Message: err.Error(),
			})
		}
	}
}
