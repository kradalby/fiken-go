package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestInvoiceDraftsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsList(context.Background(), InvoiceDraftsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpInvoicesDraftsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpInvoicesDraftsList)
	}
}

func TestInvoiceDraftsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoiceDraftsGet(context.Background(), InvoiceDraftsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoiceDraftsGet(context.Background(), InvoiceDraftsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestInvoiceDraftsListAgainstMock exercises the default empty-list
// path.
func TestInvoiceDraftsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsList(context.Background(), InvoiceDraftsListIn{Company: "acme"})
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

// TestInvoiceDraftsListAgainstMockOverride asserts the
// translation layer for draft fields including the type enum,
// int64 øre Gross, and the flattened CustomerIDs.
func TestInvoiceDraftsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	issued := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	mock.Set(OpInvoicesDraftsList, &fiken.GetInvoiceDraftsOKHeaders{
		Response: []fiken.InvoiceishDraftResult{{
			DraftId:          fiken.OptInt64{Value: 1, Set: true},
			UUID:             fiken.OptString{Value: "11111111-2222-3333-4444-555555555555", Set: true},
			Type:             fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeInvoice, Set: true},
			IssueDate:        fiken.OptDate{Value: issued, Set: true},
			DaysUntilDueDate: fiken.OptInt32{Value: 14, Set: true},
			Net:              fiken.OptInt64{Value: 800000, Set: true},
			Gross:            fiken.OptInt64{Value: 1000000, Set: true},
			Customers: []fiken.Contact{{
				ContactId: fiken.OptInt64{Value: 7, Set: true},
				Name:      "Acme Buyer",
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
	res := c.InvoiceDraftsList(context.Background(), InvoiceDraftsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.DraftID != 1 || got.UUID != "11111111-2222-3333-4444-555555555555" ||
		got.Type != "invoice" || got.IssueDate != "2024-03-01" ||
		got.DueDays != 14 || got.Gross != 1000000 || got.Net != 800000 {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.CustomerIDs) != 1 || got.CustomerIDs[0] != 7 {
		t.Fatalf("customer mismatch: %+v", got.CustomerIDs)
	}
}

// TestInvoiceDraftsGetAgainstMock asserts the single-resource path.
func TestInvoiceDraftsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpInvoicesDraftsGet, &fiken.InvoiceishDraftResult{
		DraftId: fiken.OptInt64{Value: 99, Set: true},
		UUID:    fiken.OptString{Value: "deadbeef-1111-2222-3333-444444444444", Set: true},
		Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeInvoice, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsGet(context.Background(), InvoiceDraftsGetIn{
		Company: "acme",
		DraftID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 99 || res.Ok.UUID != "deadbeef-1111-2222-3333-444444444444" ||
		res.Ok.Type != "invoice" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoiceDraftsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoiceDraftsCreate(context.Background(), InvoiceDraftsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoiceDraftsCreate(context.Background(), InvoiceDraftsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestInvoiceDraftsCreateAgainstMock asserts the create path renders
// the Location header through InvoiceDraftOut.Location.
func TestInvoiceDraftsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/drafts/99")
	mock.Set(OpInvoicesDraftsCreate, &fiken.CreateInvoiceDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsCreate(context.Background(), InvoiceDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeInvoice},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoiceDraftsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/drafts/99")
	mock.Set(OpInvoicesDraftsUpdate, &fiken.UpdateInvoiceDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsUpdate(context.Background(), InvoiceDraftsUpdateIn{
		Company: "acme", DraftID: 99,
		Body: &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeInvoice},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 99 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoiceDraftsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsDelete(context.Background(), InvoiceDraftsDeleteIn{
		Company: "acme", DraftID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted {
		t.Fatalf("expected Deleted=true, got %+v", res.Ok)
	}
}

func TestInvoiceDraftsCreateFromAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/4242")
	mock.Set(OpInvoicesDraftsCreateFrom, &fiken.CreateInvoiceFromDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsCreateFrom(context.Background(), InvoiceDraftsCreateFromIn{
		Company: "acme", DraftID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestInvoiceDraftsCreateError verifies upstream non-2xx is mapped
// through MapErr via mockfiken.SetError.
func TestInvoiceDraftsCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpInvoicesDraftsCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsCreate(context.Background(), InvoiceDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeInvoice},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestInvoiceDraftsAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoiceDraftsAttachmentsList(context.Background(), InvoiceDraftsAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoiceDraftsAttachmentsList(context.Background(), InvoiceDraftsAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestInvoiceDraftsAttachmentsListAgainstMock asserts the bare-array
// happy path.
func TestInvoiceDraftsAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpInvoicesDraftsAttachmentsList, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "DATT-1", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/draft.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoiceDraftsAttachmentsList(context.Background(), InvoiceDraftsAttachmentsListIn{
		Company: "acme",
		DraftID: 1,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].Identifier != "DATT-1" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoiceDraftsAttachmentsAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoiceDraftsAttachmentsAttach(context.Background(), InvoiceDraftsAttachmentsAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoiceDraftsAttachmentsAttach(context.Background(), InvoiceDraftsAttachmentsAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
	if got := c.InvoiceDraftsAttachmentsAttach(context.Background(), InvoiceDraftsAttachmentsAttachIn{Company: "acme", DraftID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestInvoiceDraftsAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/drafts/1/attachments/7")
	mock.Set(OpInvoicesDraftsAttachmentsAttach, &fiken.AddAttachmentToInvoiceDraftCreated{
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
	res := c.InvoiceDraftsAttachmentsAttach(context.Background(), InvoiceDraftsAttachmentsAttachIn{
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
