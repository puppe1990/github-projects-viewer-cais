package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	inertia "github.com/romsar/gonertia/v3"
)

const testInertiaRoot = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8" />{{ .inertiaHead }}</head>
<body>{{ .inertia }}</body>
</html>`

func setupTestInertia(t *testing.T) *inertia.Inertia {
	t.Helper()
	i, err := inertia.New(testInertiaRoot)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func inertiaRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("X-Inertia", "true")
	return req
}

func parseInertiaJSON(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("not json: %v body=%s", err, rr.Body.String())
	}
	return payload
}

func assertInertiaComponent(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	if payload["component"] != want {
		t.Errorf("component = %v, want %s", payload["component"], want)
	}
}

func assertInertiaProp(t *testing.T, rr *httptest.ResponseRecorder, key string) any {
	t.Helper()
	payload := parseInertiaJSON(t, rr)
	props, ok := payload["props"].(map[string]any)
	if !ok {
		t.Fatalf("missing props: %v", payload)
	}
	v, ok := props[key]
	if !ok {
		t.Fatalf("props missing %q: %v", key, props)
	}
	return v
}
