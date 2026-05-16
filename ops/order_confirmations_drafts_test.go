package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestOrderConfirmationDraftsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsList(context.Background(), OrderConfirmationDraftsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpOrderConfirmationsDraftsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpOrderConfirmationsDraftsList)
	}
}

func TestOrderConfirmationDraftsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationDraftsGet(context.Background(), OrderConfirmationDraftsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationDraftsGet(context.Background(), OrderConfirmationDraftsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestOrderConfirmationDraftsListAgainstMock exercises the default
// empty-list path.
func TestOrderConfirmationDraftsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsList(context.Background(), OrderConfirmationDraftsListIn{Company: "acme"})
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

// TestOrderConfirmationDraftsListAgainstMockOverride asserts the
// shared invoice-draft translation flows back through the order
// confirmation path.
func TestOrderConfirmationDraftsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOrderConfirmationsDraftsList, &fiken.GetOrderConfirmationDraftsOKHeaders{
		Response: []fiken.InvoiceishDraftResult{{
			DraftId: fiken.OptInt64{Value: 5, Set: true},
			UUID:    fiken.OptString{Value: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", Set: true},
			Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeOrderConfirmation, Set: true},
			Net:     fiken.OptInt64{Value: 600000, Set: true},
			Gross:   fiken.OptInt64{Value: 750000, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsList(context.Background(), OrderConfirmationDraftsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.DraftID != 5 || got.UUID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" ||
		got.Type != "order_confirmation" || got.Gross != 750000 || got.Net != 600000 {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestOrderConfirmationDraftsGetAgainstMock asserts the single-resource
// path.
func TestOrderConfirmationDraftsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOrderConfirmationsDraftsGet, &fiken.InvoiceishDraftResult{
		DraftId: fiken.OptInt64{Value: 42, Set: true},
		UUID:    fiken.OptString{Value: "00000000-1111-2222-3333-444444444444", Set: true},
		Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeOrderConfirmation, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsGet(context.Background(), OrderConfirmationDraftsGetIn{
		Company: "acme",
		DraftID: 42,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 42 || res.Ok.Type != "order_confirmation" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOrderConfirmationDraftsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationDraftsCreate(context.Background(), OrderConfirmationDraftsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationDraftsCreate(context.Background(), OrderConfirmationDraftsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestOrderConfirmationDraftsCreateAgainstMock asserts the create path
// renders the Location header through InvoiceDraftOut.Location.
func TestOrderConfirmationDraftsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/orderConfirmations/drafts/77")
	mock.Set(OpOrderConfirmationsDraftsCreate, &fiken.CreateOrderConfirmationDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsCreate(context.Background(), OrderConfirmationDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeOrderConfirmation},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestOrderConfirmationDraftsCreateError verifies upstream non-2xx
// maps through MapErr via mockfiken.SetError.
func TestOrderConfirmationDraftsCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpOrderConfirmationsDraftsCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsCreate(context.Background(), OrderConfirmationDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeOrderConfirmation},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestOrderConfirmationDraftsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/orderConfirmations/drafts/77")
	mock.Set(OpOrderConfirmationsDraftsUpdate, &fiken.UpdateOrderConfirmationDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsUpdate(context.Background(), OrderConfirmationDraftsUpdateIn{
		Company: "acme", DraftID: 77,
		Body: &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeOrderConfirmation},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 77 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOrderConfirmationDraftsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	_ = mock

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsDelete(context.Background(), OrderConfirmationDraftsDeleteIn{
		Company: "acme", DraftID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOrderConfirmationDraftsCreateFromAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/orderConfirmations/2001")
	mock.Set(OpOrderConfirmationsDraftsCreateFrom, &fiken.CreateOrderConfirmationFromDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsCreateFrom(context.Background(), OrderConfirmationDraftsCreateFromIn{
		Company: "acme", DraftID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOrderConfirmationDraftsAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationDraftsAttachmentsList(context.Background(), OrderConfirmationDraftsAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationDraftsAttachmentsList(context.Background(), OrderConfirmationDraftsAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestOrderConfirmationDraftsAttachmentsListAgainstMock asserts the
// bare-array happy path.
func TestOrderConfirmationDraftsAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOrderConfirmationsDraftsAttachmentsList, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "OCATT-1", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/order-confirmation-draft.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationDraftsAttachmentsList(context.Background(), OrderConfirmationDraftsAttachmentsListIn{
		Company: "acme",
		DraftID: 1,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].Identifier != "OCATT-1" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOrderConfirmationDraftsAttachmentsAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationDraftsAttachmentsAttach(context.Background(), OrderConfirmationDraftsAttachmentsAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationDraftsAttachmentsAttach(context.Background(), OrderConfirmationDraftsAttachmentsAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationDraftsAttachmentsAttach(context.Background(), OrderConfirmationDraftsAttachmentsAttachIn{Company: "acme", DraftID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestOrderConfirmationDraftsAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/orderConfirmations/drafts/1/attachments/7")
	mock.Set(OpOrderConfirmationsDraftsAttachmentsAttach, &fiken.AddAttachmentToOrderConfirmationDraftCreated{
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
	res := c.OrderConfirmationDraftsAttachmentsAttach(context.Background(), OrderConfirmationDraftsAttachmentsAttachIn{
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
