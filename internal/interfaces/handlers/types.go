package handlers

import "github.com/gin-gonic/gin"

type Params map[string]string
type Query map[string]string
type Body interface{}
type Result interface{}

func (p Params) Get(key string) string {
	return p[key]
}

func (q Query) Get(key string) string {
	return q[key]
}

type SimpleHandler func() error
type SimpleResultHandler func() (Result, error)

type BodyHandler func(body Body) error
type BodyResultHandler func(body Body) (Result, error)

type ParamsHandler func(params Params) error
type ParamsResultHandler func(params Params) (Result, error)

type BodyParamsHandler func(body Body, params Params) error
type BodyParamsResultHandler func(body Body, params Params) (Result, error)

type QueryHandler func(query Query) error
type QueryResultHandler func(query Query) (Result, error)

type BodyQueryHandler func(body Body, query Query) error
type BodyQueryResultHandler func(body Body, query Query) (Result, error)

type ParamsQueryHandler func(params Params, query Query) error
type ParamsQueryResultHandler func(params Params, query Query) (Result, error)

type BodyParamsQueryHandler func(body Body, params Params, query Query) error
type BodyParamsQueryResultHandler func(body Body, params Params, query Query) (Result, error)

type ContextBasedHandler func(ctx *gin.Context) error
