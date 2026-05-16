package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestPurchaseDraftsListValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchaseDraftsList(context.Background(), PurchaseDraftsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
}

func TestPurchaseDraftsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpPurchasesDraftsList, &fiken.GetPurchaseDraftsOKHeaders{
		Response: []fiken.DraftResult{
			{
				DraftId:  fiken.OptInt64{Value: 2, Set: true},
				UUID:     fiken.OptString{Value: "p-2", Set: true},
				Currency: fiken.OptString{Value: "NOK", Set: true},
			},
		},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchaseDraftsList(context.Background(), PurchaseDraftsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].DraftID != 2 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchaseDraftsGetValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchaseDraftsGet(context.Background(), PurchaseDraftsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchaseDraftsGet(context.Background(), PurchaseDraftsGetIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

func TestPurchaseDraftsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpPurchasesDraftsGet, &fiken.DraftResult{
		DraftId: fiken.OptInt64{Value: 88, Set: true},
		UUID:    fiken.OptString{Value: "p-88", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchaseDraftsGet(context.Background(), PurchaseDraftsGetIn{Company: "acme", DraftID: 88})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 88 || res.Ok.UUID != "p-88" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchaseDraftsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchaseDraftsCreate(context.Background(), PurchaseDraftsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchaseDraftsCreate(context.Background(), PurchaseDraftsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestPurchaseDraftsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/purchases/drafts/55")
	mock.Set(OpPurchasesDraftsCreate, &fiken.CreatePurchaseDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchaseDraftsCreate(context.Background(), PurchaseDraftsCreateIn{
		Company: "acme",
		Body:    &fiken.DraftRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchaseDraftsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/purchases/drafts/55")
	mock.Set(OpPurchasesDraftsUpdate, &fiken.UpdatePurchaseDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchaseDraftsUpdate(context.Background(), PurchaseDraftsUpdateIn{
		Company: "acme", DraftID: 55,
		Body: &fiken.DraftRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 55 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchaseDraftsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchaseDraftsDelete(context.Background(), PurchaseDraftsDeleteIn{
		Company: "acme", DraftID: 55,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchaseDraftsCreateFromAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/purchases/66")
	mock.Set(OpPurchasesDraftsCreateFrom, &fiken.CreatePurchaseFromDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchaseDraftsCreateFrom(context.Background(), PurchaseDraftsCreateFromIn{
		Company: "acme", DraftID: 55,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchaseDraftsAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpPurchasesDraftsAttachmentsList, []fiken.Attachment{
		{Identifier: fiken.OptString{Value: "pa-1", Set: true}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchaseDraftsAttachmentsList(context.Background(), PurchaseDraftsAttachmentsListIn{
		Company: "acme", DraftID: 55,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchaseDraftsAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/purchases/drafts/55/attachments/9")
	mock.Set(OpPurchasesDraftsAttachmentsAttach, &fiken.AddAttachmentToPurchaseDraftCreated{
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
	res := c.PurchaseDraftsAttachmentsAttach(context.Background(), PurchaseDraftsAttachmentsAttachIn{
		Company: "acme", DraftID: 55, FilePath: path,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
