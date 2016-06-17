package muxwrap

import "net/http"

// Use applies a list of middlewares onto a HandlerFunc
func Use(handler http.HandlerFunc, m ...Middleware) http.Handler {
	mc := len(m) - 1

	for i := range m {
		m := m[mc-i]
		handler = m(handler).ServeHTTP
	}

	return handler
}
