package ops

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/mockfiken"
)

func TestCompaniesGetMissingCompanyArg(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}
	res := c.CompaniesGet(context.Background(), CompaniesGetIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error for missing Company, got %+v", res)
	}
	if res.Error.Op != OpCompaniesGet {
		t.Fatalf("expected Op=%q got %q", OpCompaniesGet, res.Error.Op)
	}
}

// startMockForTest spins up a mockfiken server scoped to the test.
// Returned Server has Set / SetError for per-op overrides.
func startMockForTest(t *testing.T) *mockfiken.Server {
	t.Helper()
	return mockfiken.New(t)
}

// TestCompaniesListAgainstMock exercises the default (empty) mock
// path end-to-end through ops.Client + ogen client + httptest.
func TestCompaniesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CompaniesList(context.Background(), CompaniesListIn{})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("nil Ok")
	}
}

// TestCompaniesListAgainstMockOverride registers a success-override
// and asserts the response flows back through the translation layer.
func TestCompaniesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpCompaniesList, &fiken.GetCompaniesOKHeaders{
		Response: []fiken.Company{{
			Slug: fiken.OptString{Value: "acme", Set: true},
			Name: fiken.OptString{Value: "Acme Co", Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CompaniesList(context.Background(), CompaniesListIn{})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	if got := res.Ok.Items[0].Slug; got != "acme" {
		t.Errorf("Slug: got %q want %q", got, "acme")
	}
}

// TestCompaniesGetAgainstMock asserts CompaniesGet's happy path
// against a Company override.
func TestCompaniesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpCompaniesGet, &fiken.Company{
		Slug: fiken.OptString{Value: "acme", Set: true},
		Name: fiken.OptString{Value: "Acme Co", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CompaniesGet(context.Background(), CompaniesGetIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Slug != "acme" {
		t.Fatalf("expected slug=acme, got %+v", res.Ok)
	}
}
