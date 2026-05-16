package ops

import (
	"encoding/json"
	"testing"
)

// TestResultMarshalSuccess: success envelope serializes with ok set
// and error explicitly null (both keys always present).
func TestResultMarshalSuccess(t *testing.T) {
	type out struct {
		Name string `json:"name"`
	}
	r := Result[out]{Ok: &out{Name: "acme"}}
	got, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"ok":{"name":"acme"},"error":null}`
	if string(got) != want {
		t.Fatalf("mismatch:\n got %s\nwant %s", got, want)
	}
}

// TestResultMarshalError: error envelope serializes with ok null
// and error populated.
func TestResultMarshalError(t *testing.T) {
	r := Result[struct{}]{Error: &Error{Code: "not_found", Message: "no such company", HTTPStatus: 404}}
	got, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"ok":null,"error":{"code":"not_found","message":"no such company","http_status":404}}`
	if string(got) != want {
		t.Fatalf("mismatch:\n got %s\nwant %s", got, want)
	}
}

// TestErrorImplementsErrorInterface — ops.Error satisfies error.
func TestErrorImplementsErrorInterface(t *testing.T) {
	var _ error = (*Error)(nil)
	e := &Error{Code: "validation", Message: "bad input"}
	if got := e.Error(); got != "validation: bad input" {
		t.Fatalf("error string mismatch: %q", got)
	}
}
