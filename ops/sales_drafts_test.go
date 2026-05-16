package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestSaleDraftsListValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SaleDraftsList(context.Background(), SaleDraftsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
}

func TestSaleDraftsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesDraftsList, &fiken.GetSaleDraftsOKHeaders{
		Response: []fiken.DraftResult{
			{
				DraftId:  fiken.OptInt64{Value: 1, Set: true},
				UUID:     fiken.OptString{Value: "u-1", Set: true},
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
	res := c.SaleDraftsList(context.Background(), SaleDraftsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].DraftID != 1 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSaleDraftsGetValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SaleDraftsGet(context.Background(), SaleDraftsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SaleDraftsGet(context.Background(), SaleDraftsGetIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing draft_id: want validation error, got %+v", got)
	}
}

func TestSaleDraftsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesDraftsGet, &fiken.DraftResult{
		DraftId: fiken.OptInt64{Value: 99, Set: true},
		UUID:    fiken.OptString{Value: "u-99", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SaleDraftsGet(context.Background(), SaleDraftsGetIn{Company: "acme", DraftID: 99})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 99 || res.Ok.UUID != "u-99" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSaleDraftsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SaleDraftsCreate(context.Background(), SaleDraftsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SaleDraftsCreate(context.Background(), SaleDraftsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestSaleDraftsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/sales/drafts/77")
	mock.Set(OpSalesDraftsCreate, &fiken.CreateSaleDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SaleDraftsCreate(context.Background(), SaleDraftsCreateIn{
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

func TestSaleDraftsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/sales/drafts/77")
	mock.Set(OpSalesDraftsUpdate, &fiken.UpdateSaleDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SaleDraftsUpdate(context.Background(), SaleDraftsUpdateIn{
		Company: "acme", DraftID: 77,
		Body: &fiken.DraftRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 77 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSaleDraftsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SaleDraftsDelete(context.Background(), SaleDraftsDeleteIn{
		Company: "acme", DraftID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSaleDraftsCreateFromAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/sales/88")
	mock.Set(OpSalesDraftsCreateFrom, &fiken.CreateSaleFromDraftCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SaleDraftsCreateFrom(context.Background(), SaleDraftsCreateFromIn{
		Company: "acme", DraftID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSaleDraftsAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesDraftsAttachmentsList, []fiken.Attachment{
		{Identifier: fiken.OptString{Value: "a-1", Set: true}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SaleDraftsAttachmentsList(context.Background(), SaleDraftsAttachmentsListIn{
		Company: "acme", DraftID: 77,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSaleDraftsAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/sales/drafts/77/attachments/9")
	mock.Set(OpSalesDraftsAttachmentsAttach, &fiken.AddAttachmentToSaleDraftCreated{
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
	res := c.SaleDraftsAttachmentsAttach(context.Background(), SaleDraftsAttachmentsAttachIn{
		Company: "acme", DraftID: 77, FilePath: path,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
