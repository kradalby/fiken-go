package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestInboxListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InboxList(context.Background(), InboxListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpInboxList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpInboxList)
	}
}

func TestInboxGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InboxGet(context.Background(), InboxGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InboxGet(context.Background(), InboxGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing document_id: want validation error, got %+v", got)
	}
}

// TestInboxListAgainstMock exercises the default empty-list path.
func TestInboxListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InboxList(context.Background(), InboxListIn{Company: "acme"})
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

// TestInboxListAgainstMockOverride asserts translation of the inbox
// document fields including the createdAt timestamp and the Status
// boolean (used / unused).
func TestInboxListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	createdAt := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)
	mock.Set(OpInboxList, &fiken.GetInboxOKHeaders{
		Response: []fiken.InboxResult{{
			DocumentId:  fiken.OptInt64{Value: 42, Set: true},
			Name:        fiken.OptString{Value: "Invoice for August", Set: true},
			Description: fiken.OptString{Value: "Uploaded with API", Set: true},
			Filename:    fiken.OptString{Value: "invoice.pdf", Set: true},
			Status:      fiken.OptBool{Value: false, Set: true},
			CreatedAt:   fiken.OptDateTime{Value: createdAt, Set: true},
			DocumentUrl: fiken.OptString{Value: "https://api.fiken.test/inbox/42/file", Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InboxList(context.Background(), InboxListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.DocumentID != 42 || got.Name != "Invoice for August" ||
		got.Filename != "invoice.pdf" || got.Status ||
		got.CreatedAt != "2024-06-01T12:30:00Z" ||
		got.DocumentURL != "https://api.fiken.test/inbox/42/file" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestInboxGetAgainstMock asserts the single-resource happy path.
func TestInboxGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	createdAt := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)
	mock.Set(OpInboxGet, &fiken.InboxResult{
		DocumentId: fiken.OptInt64{Value: 42, Set: true},
		Name:       fiken.OptString{Value: "Invoice for August", Set: true},
		Filename:   fiken.OptString{Value: "invoice.pdf", Set: true},
		Status:     fiken.OptBool{Value: true, Set: true},
		CreatedAt:  fiken.OptDateTime{Value: createdAt, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InboxGet(context.Background(), InboxGetIn{Company: "acme", DocumentID: 42})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DocumentID != 42 || res.Ok.Name != "Invoice for August" ||
		!res.Ok.Status || res.Ok.CreatedAt != "2024-06-01T12:30:00Z" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInboxSendValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InboxSend(context.Background(), InboxSendIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InboxSend(context.Background(), InboxSendIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
	if got := c.InboxSend(context.Background(), InboxSendIn{Company: "acme", FilePath: "/does/not/exist"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file on disk: want validation error, got %+v", got)
	}
}

func TestInboxSendAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/inbox/42")
	mock.Set(OpInboxSend, &fiken.CreateInboxDocumentCreated{
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
	res := c.InboxSend(context.Background(), InboxSendIn{
		Company:     "acme",
		Name:        "August invoice",
		Description: "Uploaded via test",
		FilePath:    path,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
