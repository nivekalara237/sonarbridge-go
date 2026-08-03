package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler
type Builder struct {
	handler     http.Handler
	middlewares []Middleware
}

func NewBuilder(h http.Handler) *Builder {
	return &Builder{
		handler: h,
	}
}

func (b *Builder) Add(m Middleware) *Builder {
	b.middlewares = append(b.middlewares, m)
	return b
}

func (b *Builder) Build() *http.Handler {
	h := b.handler
	for i := len(b.middlewares) - 1; i >= 0; i-- {
		h = b.middlewares[i](h)
	}
	return &h
}
