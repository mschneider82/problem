package problem_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mschneider82/problem"
)

func TestProblem(t *testing.T) {
	p := problem.New(problem.Title("titlestring"), problem.Custom("x", "value"))

	if p.JSONString() != `{"title":"titlestring","x":"value"}` {
		t.Fatalf("unexpected reply")
	}

	p = problem.New(problem.Title("titlestring"), problem.Status(404), problem.Custom("x", "value"))
	str := p.JSONString()
	if str != `{"status":404,"title":"titlestring","x":"value"}` {
		t.Fatalf("unexpected reply: %s", str)
	}

	p.Append(problem.Detail("some more details"))
	str = p.JSONString()
	if str != `{"detail":"some more details","status":404,"title":"titlestring","x":"value"}` {
		t.Fatalf("unexpected reply: %s", str)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.Append(problem.Type("https://example.com/404")).ToWriter(w)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if string(bodyBytes) != `{"detail":"some more details","status":404,"title":"titlestring","type":"https://example.com/404","x":"value"}` {
		t.Fatalf("unexpected reply: %s", bodyBytes)
	}
}