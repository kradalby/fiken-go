package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestJournalEntriesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.JournalEntriesList(context.Background(), JournalEntriesListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpJournalEntriesList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpJournalEntriesList)
	}
}

func TestJournalEntriesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.JournalEntriesGet(context.Background(), JournalEntriesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.JournalEntriesGet(context.Background(), JournalEntriesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing journal_entry_id: want validation error, got %+v", got)
	}
}

// TestJournalEntriesListAgainstMock exercises the default empty-list
// path through ops + ogen + httptest.
func TestJournalEntriesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.JournalEntriesList(context.Background(), JournalEntriesListIn{Company: "acme"})
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

// TestJournalEntriesListAgainstMockOverride asserts the success-override
// flows back through the translation layer, including int64 øre line
// amounts and the YYYY-MM-DD posting date.
func TestJournalEntriesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	posted := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	mock.Set(OpJournalEntriesList, &fiken.GetJournalEntriesOKHeaders{
		Response: []fiken.JournalEntry{{
			JournalEntryId:     fiken.OptInt64{Value: 1234, Set: true},
			JournalEntryNumber: fiken.OptInt32{Value: 42, Set: true},
			Description:        "Test posting",
			Date:               posted,
			TransactionId:      fiken.OptInt64{Value: 9988, Set: true},
			Lines: []fiken.JournalEntryLine{{
				Amount:  310000,
				Account: fiken.OptString{Value: "3020", Set: true},
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
	res := c.JournalEntriesList(context.Background(), JournalEntriesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.JournalEntryID != 1234 || got.JournalEntryNumber != 42 ||
		got.Description != "Test posting" || got.TransactionID != 9988 ||
		got.TransactionDate != "2024-03-01" {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].Amount != 310000 || got.Lines[0].Account != "3020" {
		t.Fatalf("line translation mismatch: %+v", got.Lines)
	}
}

// TestJournalEntriesGetAgainstMock asserts the single-resource happy
// path.
func TestJournalEntriesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpJournalEntriesGet, &fiken.JournalEntry{
		JournalEntryId: fiken.OptInt64{Value: 99, Set: true},
		Description:    "Single fetch",
		Date:           time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.JournalEntriesGet(context.Background(), JournalEntriesGetIn{
		Company:        "acme",
		JournalEntryID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.JournalEntryID != 99 || res.Ok.Description != "Single fetch" ||
		res.Ok.TransactionDate != "2024-01-15" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestJournalEntriesCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.JournalEntriesCreate(context.Background(), JournalEntriesCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.JournalEntriesCreate(context.Background(), JournalEntriesCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestJournalEntriesCreateAgainstMock asserts the create path renders
// the Location header through JournalEntryOut.Location.
func TestJournalEntriesCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/generalJournalEntries/77")
	mock.Set(OpJournalEntriesCreate, &fiken.CreateGeneralJournalEntryCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.JournalEntriesCreate(context.Background(), JournalEntriesCreateIn{
		Company: "acme",
		Body: &fiken.GeneralJournalEntryRequest{
			Description: fiken.OptString{Value: "Opening balance", Set: true},
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestJournalEntriesCreateError verifies upstream non-2xx is mapped
// through MapErr via mockfiken.SetError.
func TestJournalEntriesCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpJournalEntriesCreate, 401, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.JournalEntriesCreate(context.Background(), JournalEntriesCreateIn{
		Company: "acme",
		Body:    &fiken.GeneralJournalEntryRequest{},
	})
	if res.Error == nil || res.Error.Code != CodeAuthInvalid {
		t.Fatalf("expected CodeAuthInvalid, got %+v", res.Error)
	}
}

func TestJournalEntriesAttachmentsAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.JournalEntriesAttachmentsAttach(context.Background(), JournalEntriesAttachmentsAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.JournalEntriesAttachmentsAttach(context.Background(), JournalEntriesAttachmentsAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing journal_entry_id: want validation error, got %+v", got)
	}
	if got := c.JournalEntriesAttachmentsAttach(context.Background(), JournalEntriesAttachmentsAttachIn{Company: "acme", JournalEntryID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestJournalEntriesAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/journalEntries/1/attachments/7")
	mock.Set(OpJournalEntriesAttachmentsAttach, &fiken.AddAttachmentToJournalEntryCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	path := writeTempAttachment(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.JournalEntriesAttachmentsAttach(context.Background(), JournalEntriesAttachmentsAttachIn{
		Company:        "acme",
		JournalEntryID: 1,
		FilePath:       path,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestJournalEntriesAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.JournalEntriesAttachmentsList(context.Background(), JournalEntriesAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.JournalEntriesAttachmentsList(context.Background(), JournalEntriesAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing journal_entry_id: want validation error, got %+v", got)
	}
}

// TestJournalEntriesAttachmentsListAgainstMock asserts the bare-array
// happy path against an override.
func TestJournalEntriesAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpJournalEntriesAttachmentsList, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "24760", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/file.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.JournalEntriesAttachmentsList(context.Background(), JournalEntriesAttachmentsListIn{
		Company:        "acme",
		JournalEntryID: 1,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 attachment, got %+v", res.Ok)
	}
	if res.Ok.Items[0].Identifier != "24760" {
		t.Fatalf("identifier mismatch: %+v", res.Ok.Items[0])
	}
}
