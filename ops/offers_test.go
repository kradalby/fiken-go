package ops

import (
	"context"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestOffersListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OffersList(context.Background(), OffersListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpOffersList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpOffersList)
	}
}

func TestOffersGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OffersGet(context.Background(), OffersGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OffersGet(context.Background(), OffersGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing offer_id: want validation error, got %+v", got)
	}
}

// TestOffersListAgainstMock exercises the default empty-list path.
func TestOffersListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OffersList(context.Background(), OffersListIn{Company: "acme"})
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

// TestOffersListAgainstMockOverride asserts translation of monetary
// fields, dates, the flat contact_id, and the shared invoice-line
// VAT-rate rescale.
func TestOffersListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	issued := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	accepted := time.Date(2024, 5, 7, 0, 0, 0, 0, time.UTC)
	mock.Set(OpOffersList, &fiken.GetOffersOKHeaders{
		Response: []fiken.Offer{{
			OfferId:     fiken.OptInt64{Value: 11, Set: true},
			OfferNumber: fiken.OptInt32{Value: 10001, Set: true},
			Date:        fiken.OptDate{Value: issued, Set: true},
			Accepted:    fiken.OptDate{Value: accepted, Set: true},
			Net:         fiken.OptInt64{Value: 1000000, Set: true},
			Vat:         fiken.OptInt64{Value: 250000, Set: true},
			Gross:       fiken.OptInt64{Value: 1250000, Set: true},
			ContactId:   fiken.OptInt64{Value: 88, Set: true},
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
	res := c.OffersList(context.Background(), OffersListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.OfferID != 11 || got.OfferNumber != 10001 ||
		got.Date != "2024-05-01" || got.Accepted != "2024-05-07" ||
		got.Gross != 1250000 || got.Net != 1000000 || got.Vat != 250000 ||
		got.ContactID != 88 {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].VatType != "high" || got.Lines[0].VatRate != 2500 {
		t.Fatalf("line translation mismatch: %+v", got.Lines)
	}
}

// TestOffersGetAgainstMock asserts the single-resource happy path.
func TestOffersGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOffersGet, &fiken.Offer{
		OfferId:     fiken.OptInt64{Value: 11, Set: true},
		OfferNumber: fiken.OptInt32{Value: 10001, Set: true},
		Net:         fiken.OptInt64{Value: 500, Set: true},
		Vat:         fiken.OptInt64{Value: 125, Set: true},
		Gross:       fiken.OptInt64{Value: 625, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OffersGet(context.Background(), OffersGetIn{
		Company: "acme",
		OfferID: "11",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.OfferID != 11 || res.Ok.OfferNumber != 10001 || res.Ok.Gross != 625 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestOffersSendValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OffersSend(context.Background(), OffersSendIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OffersSend(context.Background(), OffersSendIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestOffersSendAgainstMock asserts the send path round-trips the
// caller-supplied OfferId into the success envelope.
func TestOffersSendAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OffersSend(context.Background(), OffersSendIn{
		Company: "acme",
		Body: &fiken.SendOfferRequest{
			Method:                     []fiken.SendOfferRequestMethodItem{fiken.SendOfferRequestMethodItemEmail},
			IncludeDocumentAttachments: false,
			OfferId:                    4242,
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.OfferID != 4242 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestOffersSendError verifies upstream non-2xx maps through MapErr.
func TestOffersSendError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpOffersSend, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OffersSend(context.Background(), OffersSendIn{
		Company: "acme",
		Body:    &fiken.SendOfferRequest{OfferId: 7},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestOffersCounterCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OffersCounterCreate(context.Background(), OffersCounterCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.OffersCounterCreate(context.Background(), OffersCounterCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing value: want validation error, got %+v", got)
	}
}

// TestOffersCounterCreateAgainstMock asserts the counter-create path
// echoes the requested starting value back through the envelope.
func TestOffersCounterCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OffersCounterCreate(context.Background(), OffersCounterCreateIn{
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

func TestOffersCounterGetValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.OffersCounterGet(context.Background(), CounterGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
}

func TestOffersCounterGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpOffersCounterGet, &fiken.Counter{
		Value: fiken.OptInt32{Value: 31, Set: true},
	})
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.OffersCounterGet(context.Background(), CounterGetIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Value != 31 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
