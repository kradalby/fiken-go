package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestCreditNotesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesList(context.Background(), CreditNotesListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpCreditNotesList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpCreditNotesList)
	}
}

func TestCreditNotesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNotesGet(context.Background(), CreditNotesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNotesGet(context.Background(), CreditNotesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing credit_note_id: want validation error, got %+v", got)
	}
}

// TestCreditNotesListAgainstMock exercises the default empty-list path.
func TestCreditNotesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesList(context.Background(), CreditNotesListIn{Company: "acme"})
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

// TestCreditNotesListAgainstMockOverride asserts translation of
// monetary fields, dates, the flattened customer, and the
// invoice-line VAT-rate rescale (shared with invoices).
func TestCreditNotesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	issued := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	mock.Set(OpCreditNotesList, &fiken.GetCreditNotesOKHeaders{
		Response: []fiken.CreditNoteResult{{
			CreditNoteId:        7,
			CreditNoteNumber:    10001,
			Net:                 1000000,
			Vat:                 250000,
			Gross:               1250000,
			IssueDate:           fiken.OptDate{Value: issued, Set: true},
			AssociatedInvoiceId: fiken.OptInt64{Value: 4321, Set: true},
			Customer: fiken.Contact{
				ContactId: fiken.OptInt64{Value: 99, Set: true},
				Name:      "Acme Buyer",
			},
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
	res := c.CreditNotesList(context.Background(), CreditNotesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.CreditNoteID != 7 || got.CreditNoteNumber != 10001 ||
		got.IssueDate != "2024-03-01" ||
		got.Gross != 1250000 || got.Net != 1000000 || got.Vat != 250000 ||
		got.AssociatedInvoiceID != 4321 ||
		got.CustomerID != 99 || got.CustomerName != "Acme Buyer" {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].VatType != "high" || got.Lines[0].VatRate != 2500 {
		t.Fatalf("line translation mismatch: %+v", got.Lines)
	}
}

// TestCreditNotesGetAgainstMock asserts the single-resource happy path.
func TestCreditNotesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpCreditNotesGet, &fiken.CreditNoteResult{
		CreditNoteId:     7,
		CreditNoteNumber: 10001,
		Net:              500,
		Vat:              125,
		Gross:            625,
		Customer:         fiken.Contact{Name: "Acme Buyer"},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesGet(context.Background(), CreditNotesGetIn{
		Company:      "acme",
		CreditNoteID: "7",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.CreditNoteID != 7 || res.Ok.CreditNoteNumber != 10001 || res.Ok.Gross != 625 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNotesSendValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNotesSend(context.Background(), CreditNotesSendIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNotesSend(context.Background(), CreditNotesSendIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestCreditNotesSendAgainstMock asserts the send path round-trips the
// caller-supplied CreditNoteId into the success envelope.
func TestCreditNotesSendAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesSend(context.Background(), CreditNotesSendIn{
		Company: "acme",
		Body: &fiken.SendCreditNoteRequest{
			Method:                     []fiken.SendCreditNoteRequestMethodItem{fiken.SendCreditNoteRequestMethodItemEmail},
			IncludeDocumentAttachments: false,
			CreditNoteId:               4242,
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.CreditNoteID != 4242 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestCreditNotesSendError verifies upstream non-2xx maps through
// MapErr via mockfiken.SetError.
func TestCreditNotesSendError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpCreditNotesSend, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesSend(context.Background(), CreditNotesSendIn{
		Company: "acme",
		Body:    &fiken.SendCreditNoteRequest{CreditNoteId: 7},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestCreditNotesCounterCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNotesCounterCreate(context.Background(), CreditNotesCounterCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNotesCounterCreate(context.Background(), CreditNotesCounterCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing value: want validation error, got %+v", got)
	}
}

// TestCreditNotesCounterCreateAgainstMock asserts the counter-create
// path echoes the requested starting value back through the envelope.
func TestCreditNotesCounterCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesCounterCreate(context.Background(), CreditNotesCounterCreateIn{
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

func TestCreditNotesFullCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNotesFullCreate(context.Background(), CreditNotesFullCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNotesFullCreate(context.Background(), CreditNotesFullCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestCreditNotesFullCreateAgainstMock asserts the full-create path
// surfaces the upstream Location header through the success envelope.
func TestCreditNotesFullCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/creditNotes/9")
	mock.Set(OpCreditNotesFullCreate, &fiken.CreateFullCreditNoteCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesFullCreate(context.Background(), CreditNotesFullCreateIn{
		Company: "acme",
		Body:    &fiken.FullCreditNoteRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNotesPartialCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNotesPartialCreate(context.Background(), CreditNotesPartialCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.CreditNotesPartialCreate(context.Background(), CreditNotesPartialCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestCreditNotesPartialCreateAgainstMock asserts the partial-create
// path surfaces the upstream Location header.
func TestCreditNotesPartialCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/creditNotes/11")
	mock.Set(OpCreditNotesPartialCreate, &fiken.CreatePartialCreditNoteCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesPartialCreate(context.Background(), CreditNotesPartialCreateIn{
		Company: "acme",
		Body:    &fiken.PartialCreditNoteRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestCreditNotesCounterGetValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.CreditNotesCounterGet(context.Background(), CounterGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
}

func TestCreditNotesCounterGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpCreditNotesCounterGet, &fiken.Counter{
		Value: fiken.OptInt32{Value: 17, Set: true},
	})
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CreditNotesCounterGet(context.Background(), CounterGetIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Value != 17 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
