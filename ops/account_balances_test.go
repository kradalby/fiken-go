package ops

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestAccountBalancesListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.AccountBalancesList(context.Background(), AccountBalancesListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.AccountBalancesList(context.Background(), AccountBalancesListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing date: want validation error, got %+v", got)
	}
	if got := c.AccountBalancesList(context.Background(), AccountBalancesListIn{Company: "x", Date: "nope"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("invalid date: want validation error, got %+v", got)
	}
}

func TestAccountBalancesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.AccountBalancesGet(context.Background(), AccountBalancesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.AccountBalancesGet(context.Background(), AccountBalancesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing account_code: want validation error, got %+v", got)
	}
	if got := c.AccountBalancesGet(context.Background(), AccountBalancesGetIn{Company: "x", AccountCode: "3020"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing date: want validation error, got %+v", got)
	}
}

// TestAccountBalancesListAgainstMock exercises the default empty-list path.
func TestAccountBalancesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.AccountBalancesList(context.Background(), AccountBalancesListIn{
		Company: "acme", Date: "2024-01-31",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %+v", res.Ok)
	}
}

// TestAccountBalancesListAgainstMockOverride asserts translation.
func TestAccountBalancesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpAccountBalancesList, &fiken.GetAccountBalancesOKHeaders{
		Response: []fiken.AccountBalance{{
			Code:    fiken.OptString{Value: "3020", Set: true},
			Name:    fiken.OptString{Value: "Salgsinntekt", Set: true},
			Balance: fiken.OptInt64{Value: 50050, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.AccountBalancesList(context.Background(), AccountBalancesListIn{
		Company: "acme", Date: "2024-01-31",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.Code != "3020" || got.Name != "Salgsinntekt" || got.Balance != 50050 {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestAccountBalancesGetAgainstMock asserts the single-resource path.
func TestAccountBalancesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpAccountBalancesGet, &fiken.AccountBalance{
		Code:    fiken.OptString{Value: "1500:10001", Set: true},
		Name:    fiken.OptString{Value: "Kunde reskontro", Set: true},
		Balance: fiken.OptInt64{Value: 12345, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.AccountBalancesGet(context.Background(), AccountBalancesGetIn{
		Company: "acme", AccountCode: "1500:10001", Date: "2024-01-31",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Code != "1500:10001" || res.Ok.Balance != 12345 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
