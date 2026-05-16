package ops

import (
	"context"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestBankBalancesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankBalancesList(context.Background(), BankBalancesListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpBankBalancesList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpBankBalancesList)
	}
}

func TestBankBalancesListBadDate(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankBalancesList(context.Background(), BankBalancesListIn{Company: "x", Date: "bad"})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
}

// TestBankBalancesListAgainstMock exercises the default empty path.
func TestBankBalancesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankBalancesList(context.Background(), BankBalancesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %+v", res.Ok)
	}
}

// TestBankBalancesListAgainstMockOverride asserts translation.
func TestBankBalancesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	d := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	mock.Set(OpBankBalancesList, &fiken.GetBankBalancesOKHeaders{
		Response: []fiken.BankBalanceResult{{
			Source:          fiken.OptString{Value: "BANK", Set: true},
			BankAccountId:   fiken.OptInt64{Value: 7, Set: true},
			BankAccountCode: fiken.OptString{Value: "1920:10001", Set: true},
			Date:            fiken.OptDate{Value: d, Set: true},
			Amount:          fiken.OptInt64{Value: 250000, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankBalancesList(context.Background(), BankBalancesListIn{
		Company: "acme", Date: "2024-06-30",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.BankAccountID != 7 || got.BankAccountCode != "1920:10001" ||
		got.Date != "2024-06-30" || got.Amount != 250000 || got.Source != "BANK" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}
