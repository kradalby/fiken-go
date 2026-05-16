package mockfiken

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kradalby/fiken-go/fiken"
)

// TestUnauthenticated asserts the ogen security gate rejects
// requests that omit a Bearer token. ogen handles this before our
// handler runs, so we only need to assert non-200.
func TestUnauthenticated(t *testing.T) {
	mock := New(t)
	resp, err := http.Get(mock.URL() + "/companies")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Errorf("got 200, want non-200 (unauthorized)")
	}
}

// TestAuthenticatedDefault hits the mock with a Bearer token and
// expects the default empty-list response (no override registered).
func TestAuthenticatedDefault(t *testing.T) {
	mock := New(t)
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, mock.URL()+"/companies", nil,
	)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got %d want 200", resp.StatusCode)
	}
	var body []fiken.Company
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("default response should be empty, got %d items", len(body))
	}
}

// TestSetOverridesCompaniesList registers a success-override for
// ops.OpCompaniesList and asserts the mock returns it.
func TestSetOverridesCompaniesList(t *testing.T) {
	mock := New(t)
	want := fiken.OptString{Value: "acme", Set: true}
	mock.Set("companies_list", &fiken.GetCompaniesOKHeaders{
		Response: []fiken.Company{{Slug: want}},
	})

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, mock.URL()+"/companies", nil,
	)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got %d want 200", resp.StatusCode)
	}
	var body []fiken.Company
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) != 1 || body[0].Slug.Value != "acme" {
		t.Errorf("override not honored, got %+v", body)
	}
}
