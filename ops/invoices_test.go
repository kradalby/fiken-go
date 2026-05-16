package ops

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

// writeTempAttachment writes a tiny payload to a per-test temp file
// and returns the absolute path. Used by multipart-attach happy-path
// tests so OpenMultipartFile has a real *os.File to stream.
func writeTempAttachment(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "attach.pdf")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write temp attachment: %v", err)
	}
	return path
}

func TestInvoicesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesList(context.Background(), InvoicesListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpInvoicesList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpInvoicesList)
	}
}

func TestInvoicesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesGet(context.Background(), InvoicesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoicesGet(context.Background(), InvoicesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing invoice_id: want validation error, got %+v", got)
	}
}

// TestInvoicesListAgainstMock exercises the default empty-list path
// through ops + ogen + httptest.
func TestInvoicesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesList(context.Background(), InvoicesListIn{Company: "acme"})
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

// TestInvoicesListAgainstMockOverride asserts the success-override flows
// back through the translation layer, including int64 øre monetary
// fields, the YYYY-MM-DD issue/due dates, and the basis-points
// VatRate rescale.
func TestInvoicesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	issued := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	mock.Set(OpInvoicesList, &fiken.GetInvoicesOKHeaders{
		Response: []fiken.InvoiceResult{{
			InvoiceId:     fiken.OptInt64{Value: 1234, Set: true},
			InvoiceNumber: fiken.OptInt64{Value: 42, Set: true},
			IssueDate:     fiken.OptDate{Value: issued, Set: true},
			DueDate:       fiken.OptDate{Value: due, Set: true},
			Gross:         fiken.OptInt64{Value: 1250000, Set: true},
			Net:           fiken.OptInt64{Value: 1000000, Set: true},
			Vat:           fiken.OptInt64{Value: 250000, Set: true},
			Customer: fiken.OptContact{Value: fiken.Contact{
				ContactId: fiken.OptInt64{Value: 99, Set: true},
				Name:      "Acme Buyer",
			}, Set: true},
			Lines: []fiken.InvoiceLineResult{{
				Net:          fiken.OptInt64{Value: 1000000, Set: true},
				Vat:          fiken.OptInt64{Value: 250000, Set: true},
				Gross:        fiken.OptInt64{Value: 1250000, Set: true},
				VatType:      fiken.OptString{Value: "HIGH", Set: true},
				VatInPercent: fiken.OptFloat64{Value: 0.25, Set: true},
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
	res := c.InvoicesList(context.Background(), InvoicesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.InvoiceID != 1234 || got.InvoiceNumber != 42 ||
		got.IssueDate != "2024-03-01" || got.DueDate != "2024-04-01" ||
		got.Gross != 1250000 || got.Net != 1000000 || got.Vat != 250000 ||
		got.CustomerID != 99 || got.CustomerName != "Acme Buyer" {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].VatType != "high" || got.Lines[0].VatRate != 2500 {
		t.Fatalf("line translation mismatch: %+v", got.Lines)
	}
}

// TestInvoicesGetAgainstMock asserts the single-resource happy path.
func TestInvoicesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpInvoicesGet, &fiken.InvoiceResult{
		InvoiceId:     fiken.OptInt64{Value: 99, Set: true},
		InvoiceNumber: fiken.OptInt64{Value: 7, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesGet(context.Background(), InvoicesGetIn{
		Company:   "acme",
		InvoiceID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.InvoiceID != 99 || res.Ok.InvoiceNumber != 7 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoicesSendValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesSend(context.Background(), InvoicesSendIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoicesSend(context.Background(), InvoicesSendIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestInvoicesSendAgainstMock asserts the send path round-trips the
// caller-supplied InvoiceID into the success envelope.
func TestInvoicesSendAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesSend(context.Background(), InvoicesSendIn{
		Company: "acme",
		Body: &fiken.SendInvoiceRequest{
			Method:                     []fiken.SendInvoiceRequestMethodItem{fiken.SendInvoiceRequestMethodItemEmail},
			IncludeDocumentAttachments: false,
			InvoiceId:                  4242,
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.InvoiceID != 4242 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoicesCounterCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesCounterCreate(context.Background(), InvoicesCounterCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoicesCounterCreate(context.Background(), InvoicesCounterCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing value: want validation error, got %+v", got)
	}
}

// TestInvoicesCounterCreateAgainstMock asserts the counter-create path
// echoes the requested starting value back through the success
// envelope.
func TestInvoicesCounterCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesCounterCreate(context.Background(), InvoicesCounterCreateIn{
		Company: "acme",
		Value:   1001,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Value != 1001 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestInvoicesSendError verifies upstream non-2xx is mapped through
// MapErr via mockfiken.SetError.
func TestInvoicesSendError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpInvoicesSend, 403, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesSend(context.Background(), InvoicesSendIn{
		Company: "acme",
		Body:    &fiken.SendInvoiceRequest{InvoiceId: 42},
	})
	if res.Error == nil || res.Error.Code != CodeAuthForbidden {
		t.Fatalf("expected CodeAuthForbidden, got %+v", res.Error)
	}
}

func TestInvoicesAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesAttachmentsList(context.Background(), InvoicesAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoicesAttachmentsList(context.Background(), InvoicesAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing invoice_id: want validation error, got %+v", got)
	}
}

// TestInvoicesAttachmentsListAgainstMock asserts the bare-array happy
// path against an override.
func TestInvoicesAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpInvoicesAttachmentsList, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "ATT-1", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/invoice.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesAttachmentsList(context.Background(), InvoicesAttachmentsListIn{
		Company:   "acme",
		InvoiceID: 1,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].Identifier != "ATT-1" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoicesAttachmentsAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesAttachmentsAttach(context.Background(), InvoicesAttachmentsAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoicesAttachmentsAttach(context.Background(), InvoicesAttachmentsAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing invoice_id: want validation error, got %+v", got)
	}
	if got := c.InvoicesAttachmentsAttach(context.Background(), InvoicesAttachmentsAttachIn{Company: "acme", InvoiceID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
	if got := c.InvoicesAttachmentsAttach(context.Background(), InvoicesAttachmentsAttachIn{
		Company: "acme", InvoiceID: 1, FilePath: "/does/not/exist",
	}); got.Error == nil || got.Error.Code != CodeValidation {
		t.Fatalf("missing file on disk: want validation error, got %+v", got)
	}
}

func TestInvoicesAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/1/attachments/7")
	mock.Set(OpInvoicesAttachmentsAttach, &fiken.AddAttachmentToInvoiceCreated{
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
	res := c.InvoicesAttachmentsAttach(context.Background(), InvoicesAttachmentsAttachIn{
		Company:   "acme",
		InvoiceID: 1,
		FilePath:  path,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoicesAttachmentsAttachError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpInvoicesAttachmentsAttach, 404, []byte(`{"validationErrors":[]}`))

	path := writeTempAttachment(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesAttachmentsAttach(context.Background(), InvoicesAttachmentsAttachIn{
		Company:   "acme",
		InvoiceID: 1,
		FilePath:  path,
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestInvoicesCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesCreate(context.Background(), InvoicesCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoicesCreate(context.Background(), InvoicesCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestInvoicesCreateAgainstMock asserts the create path surfaces the
// Location header through InvoicesCreateOut.Location.
func TestInvoicesCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/4242")
	mock.Set(OpInvoicesCreate, &fiken.CreateInvoiceCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesCreate(context.Background(), InvoicesCreateIn{
		Company: "acme",
		Body: &fiken.InvoiceRequest{
			IssueDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DueDate:         time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			CustomerId:      1,
			BankAccountCode: "1920:10001",
			Lines: []fiken.InvoiceLineRequest{
				{
					Quantity:  1,
					UnitPrice: fiken.OptInt64{Value: 100, Set: true},
					VatType:   fiken.OptString{Value: "HIGH", Set: true},
				},
			},
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoicesUpdateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesUpdate(context.Background(), InvoicesUpdateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.InvoicesUpdate(context.Background(), InvoicesUpdateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing invoice_id: want validation error, got %+v", got)
	}
	if got := c.InvoicesUpdate(context.Background(), InvoicesUpdateIn{Company: "acme", InvoiceID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestInvoicesUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/4242")
	mock.Set(OpInvoicesUpdate, &fiken.UpdateInvoiceOK{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesUpdate(context.Background(), InvoicesUpdateIn{
		Company:   "acme",
		InvoiceID: 4242,
		Body:      &fiken.UpdateInvoiceRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestInvoicesCounterGetValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.InvoicesCounterGet(context.Background(), CounterGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
}

func TestInvoicesCounterGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpInvoicesCounterGet, &fiken.Counter{
		Value: fiken.OptInt32{Value: 4242, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.InvoicesCounterGet(context.Background(), CounterGetIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Value != 4242 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
