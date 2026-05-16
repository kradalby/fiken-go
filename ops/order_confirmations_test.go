package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestOrderConfirmationsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationsList(context.Background(), OrderConfirmationsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpOrderConfirmationsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpOrderConfirmationsList)
	}
}

func TestOrderConfirmationsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationsGet(context.Background(), OrderConfirmationsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationsGet(context.Background(), OrderConfirmationsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing confirmation_id: want validation error, got %+v", got)
	}
}

// TestOrderConfirmationsListAgainstMock exercises the default
// empty-list path.
func TestOrderConfirmationsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationsList(context.Background(), OrderConfirmationsListIn{Company: "acme"})
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

// TestOrderConfirmationsListAgainstMockOverride asserts translation of
// monetary fields, date, the flat contact_id, and the shared
// invoice-line VAT-rate rescale.
func TestOrderConfirmationsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	issued := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	mock.Set(OpOrderConfirmationsList, &fiken.GetOrderConfirmationsOKHeaders{
		Response: []fiken.OrderConfirmation{{
			ConfirmationId:     fiken.OptInt64{Value: 11, Set: true},
			ConfirmationNumber: fiken.OptInt32{Value: 10001, Set: true},
			Date:               fiken.OptDate{Value: issued, Set: true},
			Net:                fiken.OptInt64{Value: 1000000, Set: true},
			Vat:                fiken.OptInt64{Value: 250000, Set: true},
			Gross:              fiken.OptInt64{Value: 1250000, Set: true},
			ContactId:          fiken.OptInt64{Value: 88, Set: true},
			CreatedInvoice:     fiken.OptInt64{Value: 4242, Set: true},
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
	res := c.OrderConfirmationsList(context.Background(), OrderConfirmationsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.ConfirmationID != 11 || got.ConfirmationNumber != 10001 ||
		got.Date != "2024-05-01" ||
		got.Gross != 1250000 || got.Net != 1000000 || got.Vat != 250000 ||
		got.ContactID != 88 || got.CreatedInvoice != 4242 {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].VatType != "high" || got.Lines[0].VatRate != 2500 {
		t.Fatalf("line translation mismatch: %+v", got.Lines)
	}
}

// TestOrderConfirmationsGetAgainstMock asserts the single-resource
// happy path.
func TestOrderConfirmationsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOrderConfirmationsGet, &fiken.OrderConfirmation{
		ConfirmationId:     fiken.OptInt64{Value: 11, Set: true},
		ConfirmationNumber: fiken.OptInt32{Value: 10001, Set: true},
		Net:                fiken.OptInt64{Value: 500, Set: true},
		Vat:                fiken.OptInt64{Value: 125, Set: true},
		Gross:              fiken.OptInt64{Value: 625, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationsGet(context.Background(), OrderConfirmationsGetIn{
		Company:        "acme",
		ConfirmationID: "11",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ConfirmationID != 11 || res.Ok.ConfirmationNumber != 10001 || res.Ok.Gross != 625 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOrderConfirmationsCounterCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationsCounterCreate(context.Background(), OrderConfirmationsCounterCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationsCounterCreate(context.Background(), OrderConfirmationsCounterCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing value: want validation error, got %+v", got)
	}
}

// TestOrderConfirmationsCounterCreateAgainstMock asserts the counter
// create path echoes the requested starting value back through the
// envelope.
func TestOrderConfirmationsCounterCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationsCounterCreate(context.Background(), OrderConfirmationsCounterCreateIn{
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

func TestOrderConfirmationsCreateInvoiceDraftValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationsCreateInvoiceDraft(context.Background(), OrderConfirmationsCreateInvoiceDraftIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OrderConfirmationsCreateInvoiceDraft(context.Background(), OrderConfirmationsCreateInvoiceDraftIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing confirmation_id: want validation error, got %+v", got)
	}
}

// TestOrderConfirmationsCreateInvoiceDraftAgainstMock asserts the
// promote-to-draft path surfaces the upstream Location header.
func TestOrderConfirmationsCreateInvoiceDraftAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/drafts/99")
	mock.Set(OpOrderConfirmationsCreateInvoiceDraft, &fiken.CreateInvoiceDraftFromOrderConfirmationCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationsCreateInvoiceDraft(context.Background(), OrderConfirmationsCreateInvoiceDraftIn{
		Company:        "acme",
		ConfirmationID: "11",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestOrderConfirmationsCreateInvoiceDraftError verifies upstream
// non-2xx maps through MapErr via mockfiken.SetError.
func TestOrderConfirmationsCreateInvoiceDraftError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpOrderConfirmationsCreateInvoiceDraft, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationsCreateInvoiceDraft(context.Background(), OrderConfirmationsCreateInvoiceDraftIn{
		Company:        "acme",
		ConfirmationID: "11",
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestOrderConfirmationsCounterGetValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OrderConfirmationsCounterGet(context.Background(), CounterGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
}

func TestOrderConfirmationsCounterGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOrderConfirmationsCounterGet, &fiken.Counter{
		Value: fiken.OptInt32{Value: 5, Set: true},
	})
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OrderConfirmationsCounterGet(context.Background(), CounterGetIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Value != 5 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
