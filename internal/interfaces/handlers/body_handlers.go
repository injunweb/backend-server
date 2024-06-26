package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func wrapBodyHandler(handler BodyHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := handler(body)
		handleResponse(c, err, nil)
	}
}

func wrapBodyResultHandler(handler BodyResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := handler(body)
		handleResponse(c, err, result)
	}
}

func wrapBodyParamsHandler(handler BodyParamsHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		err := handler(body, params)
		handleResponse(c, err, nil)
	}
}

func wrapBodyParamsResultHandler(handler BodyParamsResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		result, err := handler(body, params)
		handleResponse(c, err, result)
	}
}

func wrapBodyQueryHandler(handler BodyQueryHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		err := handler(body, query)
		handleResponse(c, err, nil)
	}
}

func wrapBodyQueryResultHandler(handler BodyQueryResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		result, err := handler(body, query)
		handleResponse(c, err, result)
	}
}

func wrapBodyParamsQueryHandler(handler BodyParamsQueryHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		err := handler(body, params, query)
		handleResponse(c, err, nil)
	}
}

func wrapBodyParamsQueryResultHandler(handler BodyParamsQueryResultHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Body
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		params := make(Params)
		for _, p := range c.Params {
			params[p.Key] = p.Value
		}
		query := make(Query)
		for key, values := range c.Request.URL.Query() {
			query[key] = values[0]
		}
		result, err := handler(body, params, query)
		handleResponse(c, err, result)
	}
}
