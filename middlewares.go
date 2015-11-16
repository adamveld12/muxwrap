package muxwrap

import (
	"log"
	"net/http"
	"time"
)

// Count counts the number of times this middleware has run
func Count(next http.Handler) http.Handler {
	counter := make(chan int)
	count := 0

	go func() {
		for {
			select {
			case <-counter:
				count++
			case counter <- count:
			}
		}
	}()

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		counter <- 0
		next.ServeHTTP(res, req)
		<-counter
	})
}

// Elapsed time shows the total elapsed time of a request
func ElapsedTime(next http.Handler, elapsedCb func(res http.ResponseWriter, req *http.Request, elapsed time.Duration)) http.Handler {
	if elapsedCb == nil {
		elapsedCb = func(res http.ResponseWriter, req *http.Request, elapsed time.Duration) {
			log.Printf("[DEBUG] %s took %s \n", req.URL.RawPath, elapsed)
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
