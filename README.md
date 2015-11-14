# Muxwrap
[![wercker status](https://app.wercker.com/status/ceefa97401e7f0740fa094ffbde40dba/m "wercker status")](https://app.wercker.com/project/bykey/ceefa97401e7f0740fa094ffbde40dba)

A simple http.ServeMux wrapper adding shortcuts and middleware support.

## Usage

```go
func main(){
  mux := stratus.New()

  mux.Get("/content/", http.FileServer("static"))
  mux.Embed("/api/v1", apiRoutes())
  mux.Handle("/", NotFound)

  http.listenAndServe(":8080", mux)
}

func apiRoutes() *stratus.Mux {
  mux := stratus.New(ElapsedRequestTime)

  mux.Get("/users", getUsers)
  mux.Post("/users", createUser)

  return mux
}

func ElapsedRequestTime(next http.Handler) http.Handler {
  return func(res http.ResponseWriter, req *http.Request){
    // before handler
    start := time.Now()
    next(res, req)
    // after
    elapsed := time.Since(start)
    log.Printf("%s took %s.", req.URL.RawPath, elapsed)
  }
}
```
