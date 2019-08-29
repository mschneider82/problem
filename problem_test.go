package problem_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/mschneider82/problem"
)

func TestProblem(t *testing.T) {
	p := problem.New(problem.Title("titlestring"), problem.Custom("x", "value"))

	if p.JSONString() != `{"title":"titlestring","x":"value"}` {
		t.Fatalf("unexpected reply")
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("unexpected json marshal is fine: %v", err)
	}
	if string(b) != `{"title":"titlestring","x":"value"}` {
		t.Fatalf("unexpected reply")
	}

	p = problem.New(problem.Title("titlestring"), problem.Status(404), problem.Custom("x", "value"))
	str := p.JSONString()
	if str != `{"status":404,"title":"titlestring","x":"value"}` {
		t.Fatalf("unexpected reply: %s", str)
	}

	p.Append(problem.Detail("some more details"), problem.Instance("https://example.com/details"))
	str = p.JSONString()
	expected := `{"detail":"some more details","instance":"https://example.com/details","status":404,"title":"titlestring","x":"value"}`
	if str != expected {
		t.Fatalf("unexpected reply: \ngot: %s\nexpected: %s", str, expected)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.Append(problem.Type("https://example.com/404")).WriteTo(w)
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

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected statuscode: %d expected 404", res.StatusCode)
	}

	if string(bodyBytes) != `{"detail":"some more details","instance":"https://example.com/details","status":404,"title":"titlestring","type":"https://example.com/404","x":"value"}` {
		t.Fatalf("unexpected reply: %s", bodyBytes)
	}

}
