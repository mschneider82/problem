# problem

[![GoDoc](https://godoc.org/github.com/mschneider82/problem?status.svg)](https://godoc.org/github.com/mschneider82/problem)
![https://github.com/jpoles1/gopherbadger](./coverage_badge.png)

A golang library that implements application/problem+json

<img align="right" width="60px" title="houston we have a problem" src="https://raw.githubusercontent.com/egonelbre/gophers/master/.thumb/vector/science/rocket.png">

## Features

* compatible with application/problem+json
* inspired by https://github.com/zalando/problem
* RFC link https://tools.ietf.org/html/rfc7807

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

// create json as []byte
jsonBytes := p.JSON()
```

Custom key/values are also supported:

```go
problem.New(problem.Title("Not Found"), problem.Custom("key", "value"))
```

To write the Problem directly to a http.ResponseWriter:

```go
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.New(problem.Type("https://example.com/404"), problem.Status(404)).ToWriter(w)
	})
```