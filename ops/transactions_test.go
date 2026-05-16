package ops

import (
	"context"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestTransactionsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TransactionsList(context.Background(), TransactionsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpTransactionsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpTransactionsList)
	}
}

func TestTransactionsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TransactionsGet(context.Background(), TransactionsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TransactionsGet(context.Background(), TransactionsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing transaction_id: want validation error, got %+v", got)
	}
}

// TestTransactionsListAgainstMock exercises the default empty-list
// path through ops + ogen + httptest.
func TestTransactionsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TransactionsList(context.Background(), TransactionsListIn{Company: "acme"})
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

// TestTransactionsListAgainstMockOverride asserts the success-override
// flows back through the translation layer, including nested
// JournalEntry lines (int64 øre amounts) and YYYY-MM-DD dates.
func TestTransactionsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	created := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	posted := time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC)
	mock.Set(OpTransactionsList, &fiken.GetTransactionsOKHeaders{
		Response: []fiken.Transaction{{
			TransactionId: fiken.OptInt64{Value: 9988, Set: true},
			Description:   fiken.OptString{Value: "Salgskvittering", Set: true},
			Type:          fiken.OptString{Value: "Cash Sale", Set: true},
			CreatedDate:   fiken.OptDate{Value: created, Set: true},
			Entries: []fiken.JournalEntry{{
				JournalEntryId: fiken.OptInt64{Value: 1234, Set: true},
				Description:    "Salg",
				Date:           posted,
				Lines: []fiken.JournalEntryLine{{
					Amount:  450000,
					Account: fiken.OptString{Value: "3020", Set: true},
				}},
			}},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TransactionsList(context.Background(), TransactionsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.TransactionID != 9988 || got.Description != "Salgskvittering" ||
		got.Type != "Cash Sale" || got.CreatedDate != "2024-03-01" {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Entries) != 1 || got.Entries[0].JournalEntryID != 1234 ||
		got.Entries[0].TransactionDate != "2024-03-02" {
		t.Fatalf("nested entry translation mismatch: %+v", got.Entries)
	}
	if len(got.Entries[0].Lines) != 1 || got.Entries[0].Lines[0].Amount != 450000 ||
		got.Entries[0].Lines[0].Account != "3020" {
		t.Fatalf("line translation mismatch: %+v", got.Entries[0].Lines)
	}
}

// TestTransactionsGetAgainstMock asserts the single-resource happy
// path.
func TestTransactionsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpTransactionsGet, &fiken.Transaction{
		TransactionId: fiken.OptInt64{Value: 77, Set: true},
		Description:   fiken.OptString{Value: "Single fetch", Set: true},
		Type:          fiken.OptString{Value: "General Journal Entry", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TransactionsGet(context.Background(), TransactionsGetIn{
		Company:       "acme",
		TransactionID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.TransactionID != 77 ||
		res.Ok.Description != "Single fetch" ||
		res.Ok.Type != "General Journal Entry" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestTransactionsDeleteValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TransactionsDelete(context.Background(), TransactionsDeleteIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TransactionsDelete(context.Background(), TransactionsDeleteIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing transaction_id: want validation error, got %+v", got)
	}
	if got := c.TransactionsDelete(context.Background(), TransactionsDeleteIn{Company: "acme", TransactionID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing description: want validation error, got %+v", got)
	}
}

func TestTransactionsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpTransactionsDelete, &fiken.Transaction{
		TransactionId: fiken.OptInt64{Value: 1234, Set: true},
		Description:   fiken.OptString{Value: "duplicate", Set: true},
		Deleted:       fiken.OptBool{Value: true, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TransactionsDelete(context.Background(), TransactionsDeleteIn{
		Company:       "acme",
		TransactionID: 1234,
		Description:   "duplicate",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.TransactionID != 1234 || !res.Ok.Deleted {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
