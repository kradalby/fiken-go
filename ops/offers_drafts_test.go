package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestOfferDraftsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsList(context.Background(), OfferDraftsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpOffersDraftsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpOffersDraftsList)
	}
}

func TestOfferDraftsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OfferDraftsGet(context.Background(), OfferDraftsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OfferDraftsGet(context.Background(), OfferDraftsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestOfferDraftsListAgainstMock exercises the default empty-list path.
func TestOfferDraftsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsList(context.Background(), OfferDraftsListIn{Company: "acme"})
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

// TestOfferDraftsListAgainstMockOverride asserts the shared
// invoice-draft translation flows back through the offers path.
func TestOfferDraftsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOffersDraftsList, &fiken.GetOfferDraftsOKHeaders{
		Response: []fiken.InvoiceishDraftResult{{
			DraftId: fiken.OptInt64{Value: 5, Set: true},
			UUID:    fiken.OptString{Value: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", Set: true},
			Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeOffer, Set: true},
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
	res := c.OfferDraftsList(context.Background(), OfferDraftsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.DraftID != 5 || got.UUID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" ||
		got.Type != "offer" || got.Gross != 750000 || got.Net != 600000 {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestOfferDraftsGetAgainstMock asserts the single-resource path.
func TestOfferDraftsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOffersDraftsGet, &fiken.InvoiceishDraftResult{
		DraftId: fiken.OptInt64{Value: 42, Set: true},
		UUID:    fiken.OptString{Value: "00000000-1111-2222-3333-444444444444", Set: true},
		Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeOffer, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsGet(context.Background(), OfferDraftsGetIn{
		Company: "acme",
		DraftID: 42,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 42 || res.Ok.Type != "offer" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOfferDraftsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OfferDraftsCreate(context.Background(), OfferDraftsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OfferDraftsCreate(context.Background(), OfferDraftsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestOfferDraftsCreateAgainstMock asserts the create path renders
// the Location header through InvoiceDraftOut.Location.
func TestOfferDraftsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/offers/drafts/55")
	mock.Set(OpOffersDraftsCreate, &fiken.CreateOfferDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsCreate(context.Background(), OfferDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeOffer},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestOfferDraftsCreateError verifies upstream non-2xx maps through
// MapErr via mockfiken.SetError.
func TestOfferDraftsCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpOffersDraftsCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsCreate(context.Background(), OfferDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeOffer},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestOfferDraftsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/offers/drafts/55")
	mock.Set(OpOffersDraftsUpdate, &fiken.UpdateOfferDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsUpdate(context.Background(), OfferDraftsUpdateIn{
		Company: "acme", DraftID: 55,
		Body: &fiken.InvoiceishDraftRequest{Type: fiken.InvoiceishDraftRequestTypeOffer},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 55 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOfferDraftsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsDelete(context.Background(), OfferDraftsDeleteIn{
		Company: "acme", DraftID: 55,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOfferDraftsCreateFromAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/offers/8")
	mock.Set(OpOffersDraftsCreateFrom, &fiken.CreateOfferFromDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsCreateFrom(context.Background(), OfferDraftsCreateFromIn{
		Company: "acme", DraftID: 55,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOfferDraftsAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OfferDraftsAttachmentsList(context.Background(), OfferDraftsAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OfferDraftsAttachmentsList(context.Background(), OfferDraftsAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

// TestOfferDraftsAttachmentsListAgainstMock asserts the bare-array
// happy path.
func TestOfferDraftsAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOffersDraftsAttachmentsList, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "OFATT-1", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/offer-draft.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OfferDraftsAttachmentsList(context.Background(), OfferDraftsAttachmentsListIn{
		Company: "acme",
		DraftID: 1,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].Identifier != "OFATT-1" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOfferDraftsAttachmentsAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OfferDraftsAttachmentsAttach(context.Background(), OfferDraftsAttachmentsAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OfferDraftsAttachmentsAttach(context.Background(), OfferDraftsAttachmentsAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
	if got := c.OfferDraftsAttachmentsAttach(context.Background(), OfferDraftsAttachmentsAttachIn{Company: "acme", DraftID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestOfferDraftsAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/offers/drafts/1/attachments/7")
	mock.Set(OpOffersDraftsAttachmentsAttach, &fiken.AddAttachmentToOfferDraftCreated{
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
	res := c.OfferDraftsAttachmentsAttach(context.Background(), OfferDraftsAttachmentsAttachIn{
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
