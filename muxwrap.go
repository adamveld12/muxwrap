/*
Package muxwrap implements helper functions for fluently building http APIs
*/
package muxwrap

import (
	"log"
	"net/http"
	"time"
)

const (
	// Get is for HTTP method GET
	Get = httpMethod("GET")
	// Post is for HTTP method POST
	Post = httpMethod("POSt")
	// Put is a constant for HTTP method PUT
	Put = httpMethod("PUT")
	// Head is a constant for HTTP method HEAD
	Head = httpMethod("HEAD")
	// Delete is a constant for HTTP method DELETE
	Delete = httpMethod("DELETE")
)

type httpMethod string

// Middleware is an interface for plugging in handler behaviors
type Middleware func(next http.Handler) http.Handler

// Mux wraps http.ServeMux with conveniance functions
type Mux interface {
	http.Handler

	// Get only accepts requests with the GET http method
	Get(pattern string, handler http.HandlerFunc)

	// Post only accepts requests with the POSt http method
	Post(pattern string, handler http.HandlerFunc)

	// Put only accepts requests with the PUT http method
	Put(pattern string, handler http.HandlerFunc)

	// Head only accepts requests with the HEAD http method
	Head(pattern string, handler http.HandlerFunc)

	// Delete only accepts requests with the DELETE http method
	Delete(pattern string, handler http.HandlerFunc)

	// Pushes a middleware adapter onto the mux
	Push(middleware Middleware)

	// Embed will place a handler rooted under the specified pattern using http.StripPrefix
	Embed(pattern string, handler http.Handler)

	// Handle registers a handler for the specified pattern
	Handle(pattern string, handler http.HandlerFunc)
}

// New initializes a returns a new instance of Mux
func New(middlewares ...Middleware) Mux {
	return &builder{http.NewServeMux(), middlewares}
}

type builder struct {
	*http.ServeMux
	middlewareStack []Middleware
}

func (b *builder) Handle(pattern string, handler http.HandlerFunc) {
	b.ServeMux.HandleFunc(pattern, handler)
}

func (b *builder) Push(middleware Middleware) {
	b.middlewareStack = append(b.middlewareStack, middleware)
}

func (b *builder) Get(pattern string, handler http.HandlerFunc) {
	b.strictMethodForHandler(Get, pattern, handler)
}

func (b *builder) Post(pattern string, handler http.HandlerFunc) {
	b.strictMethodForHandler(Post, pattern, handler)
}

func (b *builder) Head(pattern string, handler http.HandlerFunc) {
	b.strictMethodForHandler(Head, pattern, handler)
}

func (b *builder) Put(pattern string, handler http.HandlerFunc) {
	b.strictMethodForHandler(Put, pattern, handler)
}

func (b *builder) Delete(pattern string, handler http.HandlerFunc) {
	b.strictMethodForHandler(Delete, pattern, handler)
}

// Embed embeds an http.ServeMux rooted at the pattern sepcified
func (b *builder) Embed(pattern string, handler http.Handler) {
	b.ServeMux.Handle(pattern, http.StripPrefix(pattern, handler))
}

func (b *builder) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler, _ := b.ServeMux.Handler(req)
	wrap(handler, b).ServeHTTP(res, req)
}

func (b *builder) strictMethodForHandler(method httpMethod, pattern string, handler http.HandlerFunc) {
	b.ServeMux.Handle(pattern, StrictMethod(method)(handler))
}

func wrap(handler http.Handler, b *builder) http.Handler {
	mw := b.middlewareStack
	mc := len(mw) - 1

	if mc >= 0 {
		for i := range mw {
			m := mw[mc-i]
			handler = m(handler)
		}
	}

	return handler
}

}


	}
}
