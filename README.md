# problem

[![GoDoc](https://godoc.org/schneider.vip/problem?status.svg)](https://godoc.org/schneider.vip/problem)
![https://github.com/jpoles1/gopherbadger](./coverage_badge.png)

A golang library that implements `application/problem+json` and `application/problem+xml`

<img align="right" width="60px" title="houston we have a problem" src="https://raw.githubusercontent.com/egonelbre/gophers/master/.thumb/vector/science/rocket.png">

## Features

* compatible with `application/problem+json`
* inspired by https://github.com/zalando/problem
* RFC link https://tools.ietf.org/html/rfc7807
* a Problem implements the Error interface and can be compared with errors.Is()
* Wrap an error to a Problem
* `application/problem+xml` is also supported using the `xml.Unmarshal` and `xml.Marshaler`

## Install

```
go get -u schneider.vip/problem
```

## Usage

```go
problem.New(problem.Title("Not Found"), problem.Status(404)).JSONString()
```

Will produce this:

```json
{
  "status": 404,
  "title": "Not Found"
}
```

You can also append some more options:

```go
p := problem.New(problem.Title("Not Found"), problem.Status(404))
p.Append(problem.Detail("some more details"))

// Use the Marshaler interface to get the problem json as []byte
jsonBytes, err := json.Marshal(p)

// or simpler (ignores the error)
jsonBytes = p.JSON()

```

Custom key/values are also supported:

```go
problem.New(problem.Title("Not Found"), problem.Custom("key", "value"))
```

To write the Problem directly to a http.ResponseWriter:

```go
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    problem.New(
      problem.Type("https://example.com/404"),
      problem.Status(404)
    ).WriteTo(w)
  })
```

Create a Problem from an existing error

```go
_, err := ioutil.ReadFile("non-existing")
if err != nil {
  p := problem.New(
    problem.Wrap(err),
    problem.Title("Internal Error"),
    problem.Status(404),
    )
  if !errors.Is(p, os.ErrNotExist) {
    t.Fatalf("expected not existing error")
  }
}
```


### [Gin](https://github.com/gin-gonic/gin) Framework
If you are using gin you can simply reply the problem to the client:

```go
func(c *gin.Context) {
  problem.New(
    problem.Title("houston! we have a problem"),
    problem.Status(http.StatusNotFound),
  ).WriteTo(c.Writer)
}
```

### [Echo](https://github.com/labstack/echo) Framework
If you are using echo you can use the following error handler to handle Problems and return them to client.

```go
func ProblemHandler(err error, c echo.Context) {
  if prb, ok := err.(*problem.Problem); ok {
    if !c.Response().Committed {
            if _, err := prb.WriteTo(c.Response().Writer); err != nil {
                    e.Logger.Error(err)
            }
    }
  } else {
        // e is an instance of echo.Echo
        e.DefaultHTTPErrorHandler(err, c)
  }
}
```
