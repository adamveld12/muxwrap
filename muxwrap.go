/*
Package muxwrap implements helper functions for fluently building http APIs
*/
package muxwrap

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	// Get is for HTTP method GET
	Get = httpMethod("GET")
	// Post is for HTTP method POST
	Post = httpMethod("POST")
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
	return &builder{http.NewServeMux(), middlewares, map[string]multiMethodHandler{}}
}

type builder struct {
	*http.ServeMux
	middlewares []Middleware
	mmHandlers  map[string]multiMethodHandler
}

func (b *builder) Handle(pattern string, handler http.HandlerFunc) {
	b.ServeMux.HandleFunc(pattern, handler)
}

func (b *builder) Push(middleware Middleware) {
	b.middlewares = append(b.middlewares, middleware)
}

func (b *builder) Get(pattern string, handler http.HandlerFunc) {
	b.addStrictHandler(Get, pattern, handler)
}

func (b *builder) Post(pattern string, handler http.HandlerFunc) {
	b.addStrictHandler(Post, pattern, handler)
}

func (b *builder) Head(pattern string, handler http.HandlerFunc) {
	b.addStrictHandler(Head, pattern, handler)
}

func (b *builder) Put(pattern string, handler http.HandlerFunc) {
	b.addStrictHandler(Put, pattern, handler)
}

func (b *builder) Delete(pattern string, handler http.HandlerFunc) {
	b.addStrictHandler(Delete, pattern, handler)
}

// Embed embeds an http.ServeMux rooted at the pattern sepcified
func (b *builder) Embed(pattern string, handler http.Handler) {
	strip := pattern
	if strip != "/" {
		strip = strings.TrimSuffix(strip, "/")
	}

	b.ServeMux.Handle(pattern, http.StripPrefix(strip, handler))
}

func (b *builder) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler, _ := b.ServeMux.Handler(req)
	Use(handler.ServeHTTP, b.middlewares...).ServeHTTP(res, req)
}

func (b *builder) addStrictHandler(method httpMethod, pattern string, handler http.HandlerFunc) {
	methodStr := string(method)

	mmHandler, exists := b.mmHandlers[pattern]

	if !exists {
		mmHandler = multiMethodHandler{&map[string]http.Handler{}}
		b.mmHandlers[pattern] = mmHandler
		b.ServeMux.Handle(pattern, mmHandler)
	}

	strictHandlerMap := *mmHandler.handlers

	if _, handlerExists := strictHandlerMap[methodStr]; handlerExists {
		panic(fmt.Sprintf("multiple registrations for %s %s", methodStr, pattern))
	}

	strictHandlerMap[methodStr] = StrictMethod(method)(handler)
}

type multiMethodHandler struct {
	handlers *map[string]http.Handler
}

func (m multiMethodHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	method := req.Method
	handler, ok := (*m.handlers)[method]

	if !ok {
		res.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		handler.ServeHTTP(res, req)
	}
}
