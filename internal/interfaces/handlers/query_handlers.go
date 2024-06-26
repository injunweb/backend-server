package handlers

import "github.com/gin-gonic/gin"

func wrapQueryHandler(handler QueryHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		err := handler(query)
		handleResponse(c, err, nil)
	}
}

func wrapQueryResultHandler(handler QueryResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		result, err := handler(query)
		handleResponse(c, err, result)
	}
}
