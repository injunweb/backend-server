package handlers

import "github.com/gin-gonic/gin"

func WrapHandler(handler interface{}) gin.HandlerFunc {
	switch h := handler.(type) {
	case SimpleHandler:
		return wrapSimpleHandler(h)
	case SimpleResultHandler:
		return wrapSimpleResultHandler(h)
	case BodyHandler:
		return wrapBodyHandler(h)
	case BodyResultHandler:
		return wrapBodyResultHandler(h)
	case ParamsHandler:
		return wrapParamsHandler(h)
	case ParamsResultHandler:
		return wrapParamsResultHandler(h)
	case BodyParamsHandler:
		return wrapBodyParamsHandler(h)
	case BodyParamsResultHandler:
		return wrapBodyParamsResultHandler(h)
	case QueryHandler:
		return wrapQueryHandler(h)
	case QueryResultHandler:
		return wrapQueryResultHandler(h)
	case BodyQueryHandler:
		return wrapBodyQueryHandler(h)
	case BodyQueryResultHandler:
		return wrapBodyQueryResultHandler(h)
	case ParamsQueryHandler:
		return wrapParamsQueryHandler(h)
	case ParamsQueryResultHandler:
		return wrapParamsQueryResultHandler(h)
	case BodyParamsQueryHandler:
		return wrapBodyParamsQueryHandler(h)
	case BodyParamsQueryResultHandler:
		return wrapBodyParamsQueryResultHandler(h)
	case ContextBasedHandler:
		return wrapContextBasedHandler(h)
	default:
		return nil
	}
}
