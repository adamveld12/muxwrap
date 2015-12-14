package muxwrap

import (
	"log"
	"net/http"
	"time"
)

// ElapsedTime shows the total elapsed time of a request
func ElapsedTime(elapsedCb func(res http.ResponseWriter, req *http.Request, elapsed time.Duration)) Middleware {
	if elapsedCb == nil {
		elapsedCb = func(res http.ResponseWriter, req *http.Request, elapsed time.Duration) {
			log.Printf("[DEBUG] %s took %s \n", req.URL.Path, elapsed)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(res, req)
			elapsed := time.Since(start)
			elapsedCb(res, req, elapsed)
		})
	}
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
