package muxwrap

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPanicsWhenMultipleStrictHandlersForSameMethod(t *testing.T) {
	mux := New()

	mux.Get("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("GET"))
	})

	func() {
		errStr := "http: multiple registrations for /"
		defer func() {
			if err := recover(); err == nil || err != errStr {
				t.Errorf("\nexpected\n%s\nactual\n%s", errStr, err)
			}
		}()

		mux.Handle("/", func(res http.ResponseWriter, req *http.Request) {})
	}()

	func() {
		errStr := "multiple registrations for GET /"
		defer func() {
			if err := recover(); err == nil || err != errStr {
				t.Errorf("\nexpected\n%s\nactual\n%s", errStr, err)
			}
		}()

		mux.Get("/", func(res http.ResponseWriter, req *http.Request) {
			res.Write([]byte("GET"))
		})
	}()
}

func TestCanRegisterMultipleStrictHandlersToSameRoute(t *testing.T) {
	mux := New()

	mux.Get("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("GET"))
	})

	mux.Post("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("POST"))
	})

	getResponse := httptest.NewRecorder()
	getReq, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte{}))
	mux.ServeHTTP(getResponse, getReq)

	output, err := getResponse.Body.ReadString('\n')
	if err != nil && err != io.EOF && string(output) != "GET" {
		t.Errorf("expected GET actual %s", string(output))
	}

	postResponse := httptest.NewRecorder()
	postReq, _ := http.NewRequest("GET", "/", bytes.NewBuffer([]byte{}))
	mux.ServeHTTP(postResponse, postReq)

	output, err = postResponse.Body.ReadString('\n')
	if err != nil && err != io.EOF && string(output) != "POST" {
		t.Errorf("expected POST actual %s", string(output))
	}
}

func TestMiddlewareHandlersExecuteInCorrectOrder(t *testing.T) {

	hBuilder := func(payload string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.Write([]byte(payload))
				next.ServeHTTP(res, req)
			})
		}
	}

	b := New(hBuilder("1"), hBuilder("2"), hBuilder("3"))

	b.Handle("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("4"))
	})
	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", bytes.NewBufferString(""))

	if err != nil {
		t.Fatal(err)
	}

	b.ServeHTTP(res, req)

	res.Flush()

	str, err := res.Body.ReadString('\n')
	if err != nil && err != io.EOF {
		t.Error(err)
	}
	if str != "1234" {
		t.Errorf("expected 1234 actual %s", str)
	}
}

func TestDeleteHandleOnlyAcceptsDeletes(t *testing.T) {
	b := New()

	b.Delete("/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != string(Delete) {
			t.Errorf("expected Delete actual %s", req.Method)
		}
	})

	executeCases(b, string(Delete), t)
}

func TestPutHandleOnlyAcceptsPuts(t *testing.T) {
	b := New()

	b.Put("/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != string(Put) {
			t.Errorf("expected Put actual %s", req.Method)
		}
	})

	executeCases(b, string(Put), t)

}

func TestPostHandleOnlyAcceptsPosts(t *testing.T) {
	b := New()

	b.Post("/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != string(Post) {
			t.Errorf("expected Post actual %s", req.Method)
		}
	})

	executeCases(b, string(Post), t)

}

func TestGetHandleOnlyAcceptsGet(t *testing.T) {
	b := New()

	b.Get("/", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != string(Get) {
			t.Errorf("expected GET actual %s", req.Method)
		}
	})

	executeCases(b, string(Get), t)
}

func executeCases(b Mux, method string, t *testing.T) {
	cases := []HttpCase{
		makeHttpCase(string(Get), "/", ""),
		makeHttpCase(string(Put), "/", ""),
		makeHttpCase(string(Head), "/", ""),
		makeHttpCase(string(Post), "/", ""),
		makeHttpCase(string(Delete), "/", ""),
	}

	for _, testCase := range cases {
		b.ServeHTTP(testCase.Response, testCase.Request)
		if method != testCase.Request.Method && testCase.Response.Code != http.StatusMethodNotAllowed {
			t.Error(errors.New(fmt.Sprintf("expected http.StatusMethodNotAllowed actual %s", testCase.Response.Code)))
		} else if method == testCase.Request.Method && testCase.Response.Code != http.StatusOK {
			t.Error(errors.New(fmt.Sprintf("expected http.StatusOk actual %s", testCase.Response.Code)))
		}
	}

}

func makeHttpCase(method, url, data string) HttpCase {
	res := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, bytes.NewBufferString(data))

	if err != nil {
		log.Fatal(err)
	}

	return HttpCase{res, req}
}

type HttpCase struct {
	Response *httptest.ResponseRecorder
	Request  *http.Request
}
