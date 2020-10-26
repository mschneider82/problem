package problem

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
)

const (
	// ContentTypeJSON https://tools.ietf.org/html/rfc7807#section-6.1
	ContentTypeJSON = "application/problem+json"
	// ContentTypeXML https://tools.ietf.org/html/rfc7807#section-6.2
	ContentTypeXML = "application/problem+xml"
)

// An Option configures a Problem using the functional options paradigm
// popularized by Rob Pike.
type Option interface {
	apply(*Problem)
}

type optionFunc func(*Problem)

func (f optionFunc) apply(problem *Problem) { f(problem) }

type Problem struct {
	data   map[string]interface{}
	reason error
}

// JSON returns the Problem as json bytes
func (p Problem) JSON() []byte {
	b, _ := p.MarshalJSON()
	return b
}

// XML returns the Problem as json bytes
func (p Problem) XML() []byte {
	b, _ := xml.Marshal(p)
	return b
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (p Problem) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &p.data)
}

// MarshalJSON implements the json.Marshaler interface
func (p Problem) MarshalJSON() ([]byte, error) {
	return json.Marshal(&p.data)
}

// UnmarshalXML implements the xml.Unmarshaler interface
func (p *Problem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	lastElem := ""
	for {
		if tok, err := d.Token(); err == nil {
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Space != "urn:ietf:rfc:7807" {
					return fmt.Errorf("Expected namespace urn:ietf:rfc:7807")
				}
				lastElem = t.Name.Local
			case xml.CharData:
				if lastElem != "" {
					if lastElem == "status" {
						i, err := strconv.Atoi(string(t))
						if err != nil {
							return err
						}
						p = p.Append(Status(i))
					} else {
						p = p.Append(Custom(lastElem, string(t)))
					}
				}
			}
		} else {
			break
		}
	}
	return nil
}

// MarshalXML implements the xml.Marshaler interface
func (p Problem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "problem"}
	start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: "urn:ietf:rfc:7807"})
	tokens := []xml.Token{start}
	for k, v := range p.data {
		v := fmt.Sprintf("%v", v)
		t := xml.StartElement{Name: xml.Name{Space: "", Local: k}}
		tokens = append(tokens, t, xml.CharData(v), xml.EndElement{Name: t.Name})
	}
	tokens = append(tokens, xml.EndElement{Name: start.Name})
	for _, t := range tokens {
		e.EncodeToken(t)
	}
	return e.Flush()
}

// XMLString returns the Problem as xml
func (p Problem) XMLString() string {
	return string(p.XML())
}

// JSONString returns the Problem as json string
func (p Problem) JSONString() string {
	return string(p.JSON())
}

// Error implements the error interface, so a Problem can be used as an error
func (p Problem) Error() string {
	return p.JSONString()
}

// Is compares Problem.Error() with err.Error()
func (p Problem) Is(err error) bool {
	return p.Error() == err.Error()
}

// Unwrap returns the result of calling the Unwrap method on err, if err implements Unwrap.
// Otherwise, Unwrap returns nil.
func (p Problem) Unwrap() error {
	return p.reason
}

// WriteTo writes the JSON Problem to a http Response Writer using the correct
// Content-Type and the problem's http statuscode
func (p Problem) WriteTo(w http.ResponseWriter) (int, error) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	if statuscode, ok := p.data["status"]; ok {
		if statusint, ok := statuscode.(int); ok {
			w.WriteHeader(statusint)
		}
	}
	return w.Write(p.JSON())
}

// WriteXMLTo writes the XML Problem to a http Response Writer using the correct
// Content-Type and the problem's http statuscode
func (p Problem) WriteXMLTo(w http.ResponseWriter) (int, error) {
	w.Header().Set("Content-Type", ContentTypeXML)
	if statuscode, ok := p.data["status"]; ok {
		if statusint, ok := statuscode.(int); ok {
			w.WriteHeader(statusint)
		}
	}
	return w.Write(p.XML())
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

// Of creates a Problem based on StatusCode with Title automatically set
func Of(statusCode int) *Problem {
	return New(Status(statusCode), Title(http.StatusText(statusCode)))
}

// Append an Option to a existing Problem
func (p *Problem) Append(opts ...Option) *Problem {
	for _, opt := range opts {
		opt.apply(p)
	}
	return p
}

// Wrap an error to the Problem
func Wrap(err error) Option {
	return optionFunc(func(problem *Problem) {
		problem.reason = err
		problem.data["reason"] = err.Error()
	})
}

// Type sets the type URI (typically, with the "http" or "https" scheme) that identifies the problem type.
// When dereferenced, it SHOULD provide human-readable documentation for the problem type
func Type(uri string) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["type"] = uri
	})
}

// Title sets a title that appropriately describes it (think short)
// Written in english and readable for engineers (usually not suited for
// non technical stakeholders and not localized); example: Service Unavailable
func Title(title string) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["title"] = title
	})
}

// Status sets the HTTP status code generated by the origin server for this
// occurrence of the problem.
func Status(status int) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["status"] = status
	})
}

// Detail A human readable explanation specific to this occurrence of the problem.
func Detail(detail string) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["detail"] = detail
	})
}

// Instance an absolute URI that identifies the specific occurrence of the
// problem.
func Instance(uri string) Option {
	return optionFunc(func(problem *Problem) {
		problem.data["instance"] = uri
	})
}

// Custom sets a custom key value
func Custom(key string, value interface{}) Option {
	return optionFunc(func(problem *Problem) {
		problem.data[key] = value
	})
}
