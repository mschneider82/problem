# problem
![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)

[![PkgGoDev](https://pkg.go.dev/badge/schneider.vip/problem)](https://pkg.go.dev/schneider.vip/problem)
[![Go Report Card](https://goreportcard.com/badge/schneider.vip/problem)](https://goreportcard.com/report/schneider.vip/problem)

A golang library that implements `application/problem+json` and `application/problem+xml`

<img align="right" width="60px" title="houston we have a problem" src="https://raw.githubusercontent.com/egonelbre/gophers/master/.thumb/vector/science/rocket.png">

## Features

* compatible with `application/problem+json`
* inspired by https://github.com/zalando/problem
* RFC link https://tools.ietf.org/html/rfc7807
* a Problem implements the Error interface and can be compared with errors.Is()
* Wrap an error to a Problem
* `application/problem+xml` is also supported using `xml.Unmarshal` and `xml.Marshal`
* Auto-Title based on StatusCode with `problem.Of(statusCode)`

## Install

```bash
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

You can also autofill the title based on the StatusCode:

```go
problem.Of(404)
```

Will produce the same problem as above!

You can also append some more options:

```go
p := problem.Of(http.StatusNotFound)
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
		problem.Status(404),
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
			if c.Request().Method == http.MethodHead {
				prb.WriteHeaderTo(c.Response())
			} else if _, err := prb.WriteTo(c.Response()); err != nil {
				c.Logger().Error(err)
			}
		}
	} else {
		c.Echo().DefaultHTTPErrorHandler(err, c)
	}
}

...
// e is an instance of echo.Echo
e.HTTPErrorHandler = ProblemHandler
```
