package problem_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"schneider.vip/problem"
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

func TestMarshalUnmarshal(t *testing.T) {
	p := problem.New(problem.Status(500), problem.Title("Strange"))

	newProblem := problem.New()
	err := json.Unmarshal(p.JSON(), &newProblem)
	if err != nil {
		t.Fatalf("no error expected in unmarshal")
	}
	if p.Error() != newProblem.Error() {
		t.Fatalf("expected equal problems, got %s - %s", p.Error(), newProblem.Error())
	}
}

func TestErrors(t *testing.T) {
	var knownProblem = problem.New(problem.Status(404), problem.Title("Go 1.13 Error"))

	var responseFromExternalService = http.Response{
		StatusCode: 404,
		Header: map[string][]string{
			"Content-Type": []string{"application/problem+json"},
		},
		Body: ioutil.NopCloser(strings.NewReader(`{"status":404,"title":"Go 1.13 Error"}`)),
	}
	// useless here but if you copy paste dont forget:
	defer responseFromExternalService.Body.Close()

	if responseFromExternalService.Header.Get("Content-Type") == problem.ContentTypeJSON {
		problemDecoder := json.NewDecoder(responseFromExternalService.Body)

		problemFromExternalService := problem.New()
		problemDecoder.Decode(&problemFromExternalService)

		if !errors.Is(problemFromExternalService, knownProblem) {
			t.Fatalf("Expected the same problem! %v, %v", problemFromExternalService, knownProblem)
		}
	}
}

func TestNestedErrors(t *testing.T) {
	rootProblem := problem.New(problem.Status(404), problem.Title("Root Problem"))
	p := problem.New(problem.Wrap(rootProblem), problem.Title("high level error msg"))

	unwrappedProblem := errors.Unwrap(p)
	if !errors.Is(unwrappedProblem, rootProblem) {
		t.Fatalf("Expected the same problem! %v, %v", unwrappedProblem, rootProblem)
	}

	if errors.Unwrap(unwrappedProblem) != nil {
		t.Fatalf("Expected unwrappedProblem has no reason")
	}
}

func TestOsErrorInProblem(t *testing.T) {
	_, err := ioutil.ReadFile("non-existing")
	if err != nil {
		p := problem.New(problem.Wrap(err), problem.Title("Internal Error"), problem.Status(404))
		if !errors.Is(p, os.ErrNotExist) {
			t.Fatalf("problem contains os.ErrNotExist")
		}

		if errors.Is(p, os.ErrPermission) {
			t.Fatalf("should not be a permission problem")
		}

		var o *os.PathError
		if !errors.As(p, &o) {
			t.Fatalf("expected error is in PathError")
		}

		newErr := errors.New("New Error")
		p = problem.New(problem.Wrap(newErr), problem.Title("NewProblem"))

		if !errors.Is(p, newErr) {
			t.Fatalf("problem should contain newErr")
		}

	}
}
