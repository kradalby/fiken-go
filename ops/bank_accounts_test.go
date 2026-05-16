package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestBankAccountsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankAccountsList(context.Background(), BankAccountsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpBankAccountsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpBankAccountsList)
	}
}

func TestBankAccountsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.BankAccountsGet(context.Background(), BankAccountsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.BankAccountsGet(context.Background(), BankAccountsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing bank_account_id: want validation error, got %+v", got)
	}
}

// TestBankAccountsListAgainstMock exercises the default empty-list
// path through ops + ogen + httptest.
func TestBankAccountsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankAccountsList(context.Background(), BankAccountsListIn{Company: "acme"})
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

// TestBankAccountsListAgainstMockOverride asserts the success-override
// flows back through the translation layer, including the int64 øre
// balance and enum type.
func TestBankAccountsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpBankAccountsList, &fiken.GetBankAccountsOKHeaders{
		Response: []fiken.BankAccountResult{{
			BankAccountId:     fiken.OptInt64{Value: 2747365, Set: true},
			Name:              fiken.OptString{Value: "Utgiftskonto DNB", Set: true},
			AccountCode:       fiken.OptString{Value: "1920:10007", Set: true},
			BankAccountNumber: fiken.OptString{Value: "15035646830", Set: true},
			Type: fiken.OptBankAccountResultType{
				Value: fiken.BankAccountResultTypeNormal,
				Set:   true,
			},
			ReconciledBalance: fiken.OptInt64{Value: 10050, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankAccountsList(context.Background(), BankAccountsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.BankAccountID != 2747365 || got.Name != "Utgiftskonto DNB" ||
		got.AccountCode != "1920:10007" || got.BankAccountNumber != "15035646830" ||
		got.Type != "normal" || got.ReconciledBalance != 10050 {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestBankAccountsGetAgainstMock asserts the single-resource happy
// path.
func TestBankAccountsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpBankAccountsGet, &fiken.BankAccountResult{
		BankAccountId: fiken.OptInt64{Value: 99, Set: true},
		Name:          fiken.OptString{Value: "Driftskonto", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankAccountsGet(context.Background(), BankAccountsGetIn{
		Company:       "acme",
		BankAccountID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.BankAccountID != 99 || res.Ok.Name != "Driftskonto" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestBankAccountsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.BankAccountsCreate(context.Background(), BankAccountsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.BankAccountsCreate(context.Background(), BankAccountsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestBankAccountsCreateAgainstMock asserts the create path returns
// the Location header through BankAccountOut.Location.
func TestBankAccountsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/bankAccounts/42")
	mock.Set(OpBankAccountsCreate, &fiken.CreateBankAccountCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankAccountsCreate(context.Background(), BankAccountsCreateIn{
		Company: "acme",
		Body: &fiken.BankAccountRequest{
			Name:              "Drift",
			BankAccountNumber: "15035646830",
			Type:              fiken.BankAccountRequestTypeNormal,
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestBankAccountsCreateError verifies upstream non-2xx is mapped
// through MapErr via mockfiken.SetError.
func TestBankAccountsCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpBankAccountsCreate, 400, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.BankAccountsCreate(context.Background(), BankAccountsCreateIn{
		Company: "acme",
		Body:    &fiken.BankAccountRequest{Name: "Drift"},
	})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected CodeValidation, got %+v", res.Error)
	}
}
