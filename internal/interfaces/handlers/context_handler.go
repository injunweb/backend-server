package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func wrapContextBasedHandler(handler ContextBasedHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := handler(c)
		handleResponse(c, err, nil)
	}
}

func handleResponse(c *gin.Context, err error, result Result) {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result != nil {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}
}
