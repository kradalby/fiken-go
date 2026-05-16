package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestSalesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesList(context.Background(), SalesListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpSalesList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpSalesList)
	}
}

func TestSalesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SalesGet(context.Background(), SalesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SalesGet(context.Background(), SalesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing sale_id: want validation error, got %+v", got)
	}
}

// TestSalesListAgainstMock exercises the default empty-list path.
func TestSalesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesList(context.Background(), SalesListIn{Company: "acme"})
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

// TestSalesListAgainstMockOverride asserts translation of monetary
// fields, dates, the flattened customer, the kind enum, and the
// payments + attachments inline children.
func TestSalesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	saleDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	payDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(OpSalesList, &fiken.GetSalesOKHeaders{
		Response: []fiken.SaleResult{{
			SaleId:             fiken.OptInt64{Value: 42, Set: true},
			SaleNumber:         fiken.OptString{Value: "XK455L", Set: true},
			Date:               fiken.OptDate{Value: saleDate, Set: true},
			DueDate:            fiken.OptDate{Value: dueDate, Set: true},
			Kind:               fiken.OptSaleResultKind{Value: fiken.SaleResultKindExternalInvoice, Set: true},
			NetAmount:          fiken.OptInt64{Value: 1000000, Set: true},
			VatAmount:          fiken.OptInt64{Value: 250000, Set: true},
			Settled:            fiken.OptBool{Value: true, Set: true},
			TotalPaid:          fiken.OptInt64{Value: 1250000, Set: true},
			OutstandingBalance: fiken.OptInt64{Value: 0, Set: true},
			Currency:           fiken.OptString{Value: "NOK", Set: true},
			Customer: fiken.OptContact{Value: fiken.Contact{
				ContactId: fiken.OptInt64{Value: 88, Set: true},
				Name:      "Acme Buyer",
			}, Set: true},
			Lines: []fiken.OrderLine{{
				Description: "Spade",
				NetPrice:    fiken.OptInt64{Value: 1000000, Set: true},
				Vat:         fiken.OptInt64{Value: 250000, Set: true},
				VatType:     "HIGH",
			}},
			SalePayments: []fiken.Payment{{
				PaymentId: fiken.OptInt64{Value: 7, Set: true},
				Date:      payDate,
				Account:   "1920:10001",
				Amount:    1250000,
			}},
			SaleAttachments: []fiken.Attachment{{
				Identifier:  fiken.OptString{Value: "INV-42", Set: true},
				DownloadUrl: fiken.OptString{Value: "https://example.test/inv.pdf", Set: true},
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
	res := c.SalesList(context.Background(), SalesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.SaleID != 42 || got.SaleNumber != "XK455L" || got.Kind != "external_invoice" ||
		got.Date != "2024-06-01" || got.DueDate != "2024-07-01" ||
		got.NetAmount != 1000000 || got.VatAmount != 250000 ||
		got.TotalPaid != 1250000 || !got.Settled ||
		got.CustomerID != 88 || got.CustomerName != "Acme Buyer" {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].VatType != "high" || got.Lines[0].NetPrice != 1000000 {
		t.Fatalf("line translation mismatch: %+v", got.Lines)
	}
	if len(got.Payments) != 1 || got.Payments[0].PaymentID != 7 ||
		got.Payments[0].Amount != 1250000 || got.Payments[0].Date != "2024-06-15" {
		t.Fatalf("payment translation mismatch: %+v", got.Payments)
	}
	if len(got.Attachments) != 1 || got.Attachments[0].Identifier != "INV-42" {
		t.Fatalf("attachment translation mismatch: %+v", got.Attachments)
	}
}

// TestSalesGetAgainstMock asserts the single-resource happy path.
func TestSalesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesGet, &fiken.SaleResult{
		SaleId:     fiken.OptInt64{Value: 42, Set: true},
		SaleNumber: fiken.OptString{Value: "XK455L", Set: true},
		NetAmount:  fiken.OptInt64{Value: 500, Set: true},
		VatAmount:  fiken.OptInt64{Value: 125, Set: true},
		Customer: fiken.OptContact{Value: fiken.Contact{
			ContactId: fiken.OptInt64{Value: 88, Set: true},
			Name:      "Acme Buyer",
		}, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesGet(context.Background(), SalesGetIn{Company: "acme", SaleID: 42})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.SaleID != 42 || res.Ok.SaleNumber != "XK455L" ||
		res.Ok.NetAmount != 500 || res.Ok.CustomerName != "Acme Buyer" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SalesCreate(context.Background(), SalesCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SalesCreate(context.Background(), SalesCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestSalesCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/sales/55")
	mock.Set(OpSalesCreate, &fiken.CreateSaleCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesCreate(context.Background(), SalesCreateIn{
		Company: "acme",
		Body:    &fiken.SaleRequest{Kind: fiken.SaleRequestKindCashSale, Currency: "NOK"},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpSalesCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesCreate(context.Background(), SalesCreateIn{
		Company: "acme",
		Body:    &fiken.SaleRequest{Kind: fiken.SaleRequestKindCashSale, Currency: "NOK"},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestSalesDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesDelete, &fiken.SaleResult{
		SaleId:  fiken.OptInt64{Value: 42, Set: true},
		Deleted: fiken.OptBool{Value: true, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesDelete(context.Background(), SalesDeleteIn{
		Company: "acme", SaleID: 42, Description: "duplicate",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted || res.Ok.SaleID != 42 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesSettleAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesSettle, &fiken.SaleResult{
		SaleId:  fiken.OptInt64{Value: 42, Set: true},
		Settled: fiken.OptBool{Value: true, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesSettle(context.Background(), SalesSettleIn{
		Company: "acme", SaleID: 42, SettledDate: "2024-06-01",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Settled || res.Ok.SaleID != 42 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesWriteOffAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesWriteOff, &fiken.SaleResult{
		SaleId:   fiken.OptInt64{Value: 42, Set: true},
		WriteOff: fiken.OptBool{Value: true, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesWriteOff(context.Background(), SalesWriteOffIn{
		Company: "acme", SaleID: 42,
		Body: &fiken.WriteOffRequest{
			Type: fiken.WriteOffRequestTypeCOLLECTIONFAILED,
			Date: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.WriteOff || res.Ok.SaleID != 42 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SalesAttachmentsList(context.Background(), SalesAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SalesAttachmentsList(context.Background(), SalesAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing sale_id: want validation error, got %+v", got)
	}
}

// TestSalesAttachmentsListAgainstMock asserts the bare-array happy path.
func TestSalesAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpSalesAttachments, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "ATT-1", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/sale.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesAttachmentsList(context.Background(), SalesAttachmentsListIn{
		Company: "acme",
		SaleID:  42,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].Identifier != "ATT-1" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SalesAttach(context.Background(), SalesAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SalesAttach(context.Background(), SalesAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing sale_id: want validation error, got %+v", got)
	}
	if got := c.SalesAttach(context.Background(), SalesAttachIn{Company: "acme", SaleID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestSalesAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/sales/1/attachments/7")
	mock.Set(OpSalesAttach, &fiken.AddAttachmentToSaleCreated{
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
	res := c.SalesAttach(context.Background(), SalesAttachIn{
		Company:      "acme",
		SaleID:       1,
		FilePath:     path,
		AttachToSale: true,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesPaymentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SalesPaymentsList(context.Background(), SalesPaymentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SalesPaymentsList(context.Background(), SalesPaymentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing sale_id: want validation error, got %+v", got)
	}
}

// TestSalesPaymentsListAgainstMock asserts translation of payment
// fields including the date format and the int64 øre amounts.
func TestSalesPaymentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	payDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(OpSalesPaymentsList, []fiken.Payment{{
		PaymentId: fiken.OptInt64{Value: 7, Set: true},
		Date:      payDate,
		Account:   "1920:10001",
		Amount:    1250000,
		Currency:  fiken.OptString{Value: "NOK", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesPaymentsList(context.Background(), SalesPaymentsListIn{
		Company: "acme",
		SaleID:  42,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.PaymentID != 7 || got.Date != "2024-06-15" ||
		got.Amount != 1250000 || got.Account != "1920:10001" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

func TestSalesPaymentsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SalesPaymentsGet(context.Background(), SalesPaymentsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.SalesPaymentsGet(context.Background(), SalesPaymentsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing sale_id: want validation error, got %+v", got)
	}
	if got := c.SalesPaymentsGet(context.Background(), SalesPaymentsGetIn{Company: "x", SaleID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing payment_id: want validation error, got %+v", got)
	}
}

// TestSalesPaymentsGetAgainstMock asserts the single-resource happy path.
func TestSalesPaymentsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	payDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(OpSalesPaymentsGet, &fiken.Payment{
		PaymentId: fiken.OptInt64{Value: 7, Set: true},
		Date:      payDate,
		Account:   "1920:10001",
		Amount:    1250000,
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesPaymentsGet(context.Background(), SalesPaymentsGetIn{
		Company:   "acme",
		SaleID:    42,
		PaymentID: 7,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.PaymentID != 7 || res.Ok.Amount != 1250000 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestSalesPaymentsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.SalesPaymentsCreate(context.Background(), SalesPaymentsCreateIn{Company: "acme", SaleID: 42}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestSalesPaymentsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/sales/42/payments/9")
	mock.Set(OpSalesPaymentsCreate, &fiken.CreateSalePaymentCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.SalesPaymentsCreate(context.Background(), SalesPaymentsCreateIn{
		Company: "acme", SaleID: 42,
		Body: &fiken.Payment{Date: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), Account: "1920", Amount: 1250000},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
