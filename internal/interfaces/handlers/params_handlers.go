package handlers

import "github.com/gin-gonic/gin"

func wrapParamsHandler(handler ParamsHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		err := handler(params)
		handleResponse(c, err, nil)
	}
}

func wrapParamsResultHandler(handler ParamsResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		result, err := handler(params)
		handleResponse(c, err, result)
	}
}

func wrapParamsQueryHandler(handler ParamsQueryHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		err := handler(params, query)
		handleResponse(c, err, nil)
	}
}

func wrapParamsQueryResultHandler(handler ParamsQueryResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		result, err := handler(params, query)
		handleResponse(c, err, result)
	}
}
