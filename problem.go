package problem

import (
	"encoding/json"
	"net/http"
)

// An Option configures a Problem using the functional options paradigm
// popularized by Rob Pike.
type Option interface {
	apply(*Problem)
}

type optionFunc func(*Problem)

func (f optionFunc) apply(problem *Problem) { f(problem) }

type Problem struct {
	data map[string]interface{}
}

// JSON returns the Problem as json bytes
func (p Problem) JSON() []byte {
	b, _ := json.Marshal(&p.data)
	return b
}

// JSONString returns the Problem as json string
func (p Problem) JSONString() string {
	return string(p.JSON())
}

// ToWriter writes the Problem to a http Response Writer
func (p Problem) ToWriter(w http.ResponseWriter) (int, error) {
	w.Header().Set("Content-Type", "application/problem+json")
	if statuscode, ok := p.data["status"]; ok {
		if statusint, ok := statuscode.(int); ok {
			w.WriteHeader(statusint)
		}
	}
	return w.Write(p.JSON())
}

// New generates a new Problem
func New(opts ...Option) *Problem {
	problem := &Problem{}
	problem.data = make(map[string]interface{})
	for _, opt := range opts {
		opt.apply(problem)
	}
	return problem
}

// Append an Option to a existing Problem
func (p *Problem) Append(opts ...Option) *Problem {
	for _, opt := range opts {
		opt.apply(p)
	}
	return p
}

// Type sets the type URI (typically, with the "http" or "https" scheme),
func Type(typeuri string) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["type"] = typeuri
	})
}

// Title sets a title that appropriately describes it (think short)
func Title(title string) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["title"] = title
	})
}

// Status sets the HTTP status code for the problem
func Status(status int) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["status"] = status
	})
}

// Detail sets an Detail
func Detail(detail string) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["detail"] = detail
	})
}

// Custom sets a custom key value
func Custom(key string, value interface{}) Option {
	return optionFunc(func(problem *Problem) {
		problem.data[key] = value
	})
}
