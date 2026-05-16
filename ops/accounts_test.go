package ops

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestAccountsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.AccountsList(context.Background(), AccountsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpAccountsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpAccountsList)
	}
}

func TestAccountsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.AccountsGet(context.Background(), AccountsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.AccountsGet(context.Background(), AccountsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing account_code: want validation error, got %+v", got)
	}
}

// TestAccountsListAgainstMock exercises the default empty-list path
// through ops + ogen + httptest.
func TestAccountsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.AccountsList(context.Background(), AccountsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("nil Ok")
	}
	if len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %d", len(res.Ok.Items))
	}
}

// TestAccountsListAgainstMockOverride asserts the success-override
// flows back through the translation layer.
func TestAccountsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpAccountsList, &fiken.GetAccountsOKHeaders{
		Response: []fiken.Account{{
			Code: fiken.OptString{Value: "3020", Set: true},
			Name: fiken.OptString{Value: "Salgsinntekt", Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.AccountsList(context.Background(), AccountsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.Code != "3020" || got.Name != "Salgsinntekt" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestAccountsGetAgainstMock asserts the single-resource happy path.
func TestAccountsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpAccountsGet, &fiken.Account{
		Code: fiken.OptString{Value: "1500:10001", Set: true},
		Name: fiken.OptString{Value: "Kunde reskontro", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.AccountsGet(context.Background(), AccountsGetIn{
		Company:     "acme",
		AccountCode: "1500:10001",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Code != "1500:10001" || res.Ok.Name != "Kunde reskontro" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
