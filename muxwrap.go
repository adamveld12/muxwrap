/*
Package muxwrap implements helper functions for fluently building http APIs
*/
package muxwrap

import (
	"log"
	"net/http"
	"time"
)

// Constants for HTTP methods
const (
	Get    = httpMethod("GET")
	Post   = httpMethod("POSt")
	Put    = httpMethod("PUT")
	Head   = httpMethod("HEAD")
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

	// Registers a handler for the specified pattern
	Handle(pattern string, handler http.HandlerFunc)
}

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

// Elapsed time shows the total elapsed time of a request
func ElapsedTime(next http.Handler, elapsedCb func(res http.ResponseWriter, req *http.Request, elapsed time.Duration)) http.Handler {
	if elapsedCb == nil {
		elapsedCb = func(res http.ResponseWriter, req *http.Request, elapsed time.Duration) {
			log.Printf("%s took %s \n", req.URL.RawPath, elapsed)
		}
	}

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(res, req)
		elapsed := time.Since(start)
		elapsedCb(res, req, elapsed)
	})
}

// StrictMethod forces a handler to only accept a set of defined HTTP Methods
func StrictMethod(methods ...httpMethod) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			match := false

			for _, method := range methods {
				if string(method) == req.Method {
					match = true
					break
				}
			}

			if !match {
				res.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			next.ServeHTTP(res, req)
		})
	}
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
