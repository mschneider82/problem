/*
Package problem implements RFC7807 `application/problem+json` and
`application/problem+xml` using the functional options paradigm.

# Features

- Compatible with `application/problem+json`.
- Inspired by https://github.com/zalando/problem.
- RFC link https://tools.ietf.org/html/rfc7807.
- A Problem implements the Error interface and can be compared with errors.Is().
- Wrap an error to a Problem.
- `application/problem+xml` is also supported using `xml.Unmarshal` and `xml.Marshal`.
- Auto-Title based on StatusCode with `problem.Of(statusCode)`.

# Installation

To install the package, run:

go get -u schneider.vip/problem
*/
package problem
