package middleware

import (
	"github.com/julienschmidt/httprouter"
)

type Middleware func(next httprouter.Handle) httprouter.Handle

func Chain(middlewares ...Middleware) Middleware {
	return func(next httprouter.Handle) httprouter.Handle {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}