package handlers

import "github.com/gin-gonic/gin"

func wrapSimpleHandler(handler SimpleHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := handler()
		handleResponse(c, err, nil)
	}
}

func wrapSimpleResultHandler(handler SimpleResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := handler()
		handleResponse(c, err, result)
	}
}
