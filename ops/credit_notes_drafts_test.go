package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestCreditNoteDraftsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsList(context.Background(), CreditNoteDraftsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpCreditNotesDraftsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpCreditNotesDraftsList)
	}
}

func TestCreditNoteDraftsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNoteDraftsGet(context.Background(), CreditNoteDraftsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNoteDraftsGet(context.Background(), CreditNoteDraftsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestCreditNoteDraftsListAgainstMock exercises the default
// empty-list path.
func TestCreditNoteDraftsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsList(context.Background(), CreditNoteDraftsListIn{Company: "acme"})
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

// TestCreditNoteDraftsListAgainstMockOverride asserts the shared
// invoice-draft translation flows back through the credit-note path.
func TestCreditNoteDraftsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpCreditNotesDraftsList, &fiken.GetCreditNoteDraftsOKHeaders{
		Response: []fiken.InvoiceishDraftResult{{
			DraftId: fiken.OptInt64{Value: 1, Set: true},
			UUID:    fiken.OptString{Value: "11111111-2222-3333-4444-555555555555", Set: true},
			Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeCreditNote, Set: true},
			Net:     fiken.OptInt64{Value: 800000, Set: true},
			Gross:   fiken.OptInt64{Value: 1000000, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsList(context.Background(), CreditNoteDraftsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.DraftID != 1 || got.UUID != "11111111-2222-3333-4444-555555555555" ||
		got.Type != "credit_note" || got.Gross != 1000000 || got.Net != 800000 {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestCreditNoteDraftsGetAgainstMock asserts the single-resource path.
func TestCreditNoteDraftsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpCreditNotesDraftsGet, &fiken.InvoiceishDraftResult{
		DraftId: fiken.OptInt64{Value: 42, Set: true},
		UUID:    fiken.OptString{Value: "deadbeef-1111-2222-3333-444444444444", Set: true},
		Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeCreditNote, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsGet(context.Background(), CreditNoteDraftsGetIn{
		Company: "acme",
		DraftID: 42,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 42 || res.Ok.Type != "credit_note" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNoteDraftsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNoteDraftsCreate(context.Background(), CreditNoteDraftsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNoteDraftsCreate(context.Background(), CreditNoteDraftsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestCreditNoteDraftsCreateAgainstMock asserts the create path
// renders the Location header through InvoiceDraftOut.Location.
func TestCreditNoteDraftsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/creditNotes/drafts/77")
	mock.Set(OpCreditNotesDraftsCreate, &fiken.CreateCreditNoteDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsCreate(context.Background(), CreditNoteDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeCreditNote},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestCreditNoteDraftsCreateError verifies upstream non-2xx is mapped
// through MapErr via mockfiken.SetError.
func TestCreditNoteDraftsCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpCreditNotesDraftsCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsCreate(context.Background(), CreditNoteDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeCreditNote},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestCreditNoteDraftsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/creditNotes/drafts/77")
	mock.Set(OpCreditNotesDraftsUpdate, &fiken.UpdateCreditNoteDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsUpdate(context.Background(), CreditNoteDraftsUpdateIn{
		Company: "acme", DraftID: 77,
		Body: &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeCreditNote},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 77 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNoteDraftsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsDelete(context.Background(), CreditNoteDraftsDeleteIn{
		Company: "acme", DraftID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNoteDraftsCreateFromAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/creditNotes/9")
	mock.Set(OpCreditNotesDraftsCreateFrom, &fiken.CreateCreditNoteFromDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsCreateFrom(context.Background(), CreditNoteDraftsCreateFromIn{
		Company: "acme", DraftID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNoteDraftsAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNoteDraftsAttachmentsList(context.Background(), CreditNoteDraftsAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNoteDraftsAttachmentsList(context.Background(), CreditNoteDraftsAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestCreditNoteDraftsAttachmentsListAgainstMock asserts the
// bare-array happy path.
func TestCreditNoteDraftsAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpCreditNotesDraftsAttachmentsList, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "CDATT-1", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/cn-draft.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNoteDraftsAttachmentsList(context.Background(), CreditNoteDraftsAttachmentsListIn{
		Company: "acme",
		DraftID: 1,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].Identifier != "CDATT-1" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNoteDraftsAttachmentsAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNoteDraftsAttachmentsAttach(context.Background(), CreditNoteDraftsAttachmentsAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNoteDraftsAttachmentsAttach(context.Background(), CreditNoteDraftsAttachmentsAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
	if got := c.CreditNoteDraftsAttachmentsAttach(context.Background(), CreditNoteDraftsAttachmentsAttachIn{Company: "acme", DraftID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestCreditNoteDraftsAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/creditNotes/drafts/1/attachments/7")
	mock.Set(OpCreditNotesDraftsAttachmentsAttach, &fiken.AddAttachmentToCreditNoteDraftCreated{
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
	res := c.CreditNoteDraftsAttachmentsAttach(context.Background(), CreditNoteDraftsAttachmentsAttachIn{
		Company:  "acme",
		DraftID:  1,
		FilePath: path,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
