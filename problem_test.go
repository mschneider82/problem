package problem_test

import (
	"encoding/json"
	"encoding/xml"
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

	p = problem.Of(http.StatusAccepted)
	str = p.JSONString()
	if str != `{"status":202,"title":"Accepted"}` {
		t.Fatalf("unexpected reply: %s", str)
	}
}

func TestProblemHTTP(t *testing.T) {
	p := problem.New(problem.Title("titlestring"), problem.Status(404), problem.Custom("x", "value"))
	p.Append(problem.Detail("some more details"), problem.Instance("https://example.com/details"))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.Append(problem.Type("https://example.com/404"))
		if r.Method == "HEAD" {
			p.WriteHeaderTo(w)
		} else {
			p.WriteTo(w)
		}
	}))
	defer ts.Close()

	// Try GET request
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
	if res.Header.Get("Content-Type") != problem.ContentTypeJSON {
		t.Fatalf("unexpected ContentType %s", res.Header.Get("Content-Type"))
	}

	if string(bodyBytes) != `{"detail":"some more details","instance":"https://example.com/details","status":404,"title":"titlestring","type":"https://example.com/404","x":"value"}` {
		t.Fatalf("unexpected reply: %s", bodyBytes)
	}

	// Try HEAD request
	res, err = http.Head(ts.URL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	bodyBytes, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(bodyBytes) != 0 {
		t.Fatal("expected empty body")
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected statuscode: %d expected 404", res.StatusCode)
	}
	if res.Header.Get("Content-Type") != problem.ContentTypeJSON {
		t.Fatalf("unexpected ContentType %s", res.Header.Get("Content-Type"))
	}
}

func TestXMLProblem(t *testing.T) {
	p := problem.New(problem.Status(404))
	xmlstr := p.XMLString()
	expected := `<problem xmlns="urn:ietf:rfc:7807"><status>404</status></problem>`
	if xmlstr != expected {
		t.Fatalf("unexpected reply: \ngot: %s\nexpected: %s", xmlstr, expected)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.Append(problem.Type("https://example.com/404"))
		if r.Method == "HEAD" {
			p.WriteXMLHeaderTo(w)
		} else {
			p.WriteXMLTo(w)
		}
	}))
	defer ts.Close()

	// Try GET request
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
	if res.Header.Get("Content-Type") != problem.ContentTypeXML {
		t.Fatalf("unexpected ContentType %s", res.Header.Get("Content-Type"))
	}

	if string(bodyBytes) != `<problem xmlns="urn:ietf:rfc:7807"><status>404</status><type>https://example.com/404</type></problem>` && string(bodyBytes) != `<problem xmlns="urn:ietf:rfc:7807"><type>https://example.com/404</type><status>404</status></problem>` {
		t.Fatalf("unexpected reply: %s", bodyBytes)
	}

	// Try HEAD request
	res, err = http.Head(ts.URL)
	if err != nil {
		t.Fatalf("%v", err)
	}
	bodyBytes, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if len(bodyBytes) != 0 {
		t.Fatal("expected empty body")
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected statuscode: %d expected 404", res.StatusCode)
	}
	if res.Header.Get("Content-Type") != problem.ContentTypeXML {
		t.Fatalf("unexpected Content-Type %s", res.Header.Get("Content-Type"))
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

func TestXMLMarshalUnmarshal(t *testing.T) {
	p := problem.New(problem.Status(500), problem.Title("StrangeXML"))

	xmlProblem, err := xml.Marshal(&p)
	if err != nil {
		t.Fatalf("no error expected in Marshal: %s", err.Error())
	}
	if string(xmlProblem) != `<problem xmlns="urn:ietf:rfc:7807"><status>500</status><title>StrangeXML</title></problem>` &&
		string(xmlProblem) != `<problem xmlns="urn:ietf:rfc:7807"><title>StrangeXML</title><status>500</status></problem>` {
		t.Fatalf("not expected xml output: %s", string(xmlProblem))
	}

	newProblem := problem.New()
	err = xml.Unmarshal(xmlProblem, &newProblem)
	if err != nil {
		t.Fatalf("no error expected in Unmarshal: %s", err.Error())
	}

	if newProblem.JSONString() != `{"status":500,"title":"StrangeXML"}` && newProblem.JSONString() != `{"title":"StrangeXML","status":500}` {
		t.Fatalf(`Expected {"status":500,"title":"StrangeXML"} got: %s`, newProblem.JSONString())
	}

	wrongProblem := []byte(`<problem xmlns="unknown"><status>123</status><title>xxx</title></problem>`)
	err = xml.Unmarshal(wrongProblem, &newProblem)
	if err == nil {
		t.Fatalf("Expected an error in Unmarshal")
	}

	wrongProblem = []byte(`<problem xmlns="urn:ietf:rfc:7807"><status>abc</status><title>StrangeXML</title></problem>`)
	err = xml.Unmarshal(wrongProblem, &newProblem)
	if err == nil {
		t.Fatalf("Expected an error in Unmarshal: status is not an int")
	}
}

func TestErrors(t *testing.T) {
	var knownProblem = problem.New(problem.Status(404), problem.Title("Go 1.13 Error"))

	var responseFromExternalService = http.Response{
		StatusCode: 404,
		Header: map[string][]string{
			"Content-Type": {"application/problem+json"},
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
	// See wrapped error in 'reason'
	if p.JSONString() != `{"reason":"{\"status\":404,\"title\":\"Root Problem\"}","title":"high level error msg"}` {
		t.Fatalf("Unexpected contents %s in problem", p.JSONString())
	}

	p = problem.New(problem.WrapSilent(rootProblem), problem.Title("high level error msg"))
	// We should not see a "reason" here
	if p.JSONString() != `{"title":"high level error msg"}` {
		t.Fatalf("Unexpected contents %s in problem", p.JSONString())
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

func TestTitlef(t *testing.T) {
	expected := "{\"title\":\"this is a test\"}"
	toTest := problem.New(problem.Titlef("this is a %s", "test")).JSONString()

	if !strings.Contains(expected, toTest) {
		t.Fatalf("expected problem %s to match %s", toTest, expected)
	}
}

func TestDetailf(t *testing.T) {
	expected := "{\"detail\":\"this is a test\"}"
	toTest := problem.New(problem.Detailf("this is a %s", "test")).JSONString()

	if !strings.Contains(expected, toTest) {
		t.Fatalf("expected problem %s to match %s", toTest, expected)
	}
}

func TestInstancef(t *testing.T) {
	expected := "{\"instance\":\"this is a test\"}"
	toTest := problem.New(problem.Instancef("this is a %s", "test")).JSONString()

	if !strings.Contains(expected, toTest) {
		t.Fatalf("expected problem %s to match %s", toTest, expected)
	}
}
