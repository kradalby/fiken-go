package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestPurchasesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesList(context.Background(), PurchasesListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpPurchasesList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpPurchasesList)
	}
}

func TestPurchasesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchasesGet(context.Background(), PurchasesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchasesGet(context.Background(), PurchasesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing purchase_id: want validation error, got %+v", got)
	}
}

// TestPurchasesListAgainstMock exercises the default empty-list path.
func TestPurchasesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesList(context.Background(), PurchasesListIn{Company: "acme"})
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

// TestPurchasesListAgainstMockOverride asserts translation of monetary
// fields, dates, the flattened supplier, the kind enum, and the
// payments + attachments inline children.
func TestPurchasesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	purchaseDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	payDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(OpPurchasesList, &fiken.GetPurchasesOKHeaders{
		Response: []fiken.PurchaseResult{{
			PurchaseId: fiken.OptInt64{Value: 99, Set: true},
			Identifier: fiken.OptString{Value: "INV-99", Set: true},
			Date:       purchaseDate,
			DueDate:    fiken.OptDate{Value: dueDate, Set: true},
			Kind:       fiken.PurchaseResultKindSupplier,
			Paid:       true,
			Currency:   "NOK",
			Supplier: fiken.OptContact{Value: fiken.Contact{
				ContactId: fiken.OptInt64{Value: 77, Set: true},
				Name:      "Acme Supplier",
			}, Set: true},
			Lines: []fiken.OrderLine{{
				Description: "Pallet of widgets",
				NetPrice:    fiken.OptInt64{Value: 800000, Set: true},
				Vat:         fiken.OptInt64{Value: 200000, Set: true},
				VatType:     "HIGH",
			}},
			Payments: []fiken.Payment{{
				PaymentId: fiken.OptInt64{Value: 11, Set: true},
				Date:      payDate,
				Account:   "2400:10001",
				Amount:    1000000,
			}},
			PurchaseAttachments: []fiken.Attachment{{
				Identifier:  fiken.OptString{Value: "ATT-99", Set: true},
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
	res := c.PurchasesList(context.Background(), PurchasesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.PurchaseID != 99 || got.Identifier != "INV-99" || got.Kind != "supplier" ||
		got.Date != "2024-06-01" || got.DueDate != "2024-07-01" ||
		!got.Paid || got.Currency != "NOK" ||
		got.SupplierID != 77 || got.SupplierName != "Acme Supplier" {
		t.Fatalf("translation mismatch: %+v", got)
	}
	if len(got.Lines) != 1 || got.Lines[0].VatType != "high" || got.Lines[0].NetPrice != 800000 {
		t.Fatalf("line translation mismatch: %+v", got.Lines)
	}
	if len(got.Payments) != 1 || got.Payments[0].PaymentID != 11 ||
		got.Payments[0].Amount != 1000000 || got.Payments[0].Date != "2024-06-15" {
		t.Fatalf("payment translation mismatch: %+v", got.Payments)
	}
	if len(got.Attachments) != 1 || got.Attachments[0].Identifier != "ATT-99" {
		t.Fatalf("attachment translation mismatch: %+v", got.Attachments)
	}
}

// TestPurchasesGetAgainstMock asserts the single-resource happy path.
func TestPurchasesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	purchaseDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	mock.Set(OpPurchasesGet, &fiken.PurchaseResult{
		PurchaseId: fiken.OptInt64{Value: 99, Set: true},
		Identifier: fiken.OptString{Value: "INV-99", Set: true},
		Date:       purchaseDate,
		Kind:       fiken.PurchaseResultKindCashPurchase,
		Currency:   "NOK",
		Supplier: fiken.OptContact{Value: fiken.Contact{
			ContactId: fiken.OptInt64{Value: 77, Set: true},
			Name:      "Acme Supplier",
		}, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesGet(context.Background(), PurchasesGetIn{Company: "acme", PurchaseID: 99})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.PurchaseID != 99 || res.Ok.Identifier != "INV-99" ||
		res.Ok.Kind != "cash_purchase" || res.Ok.SupplierName != "Acme Supplier" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchasesCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchasesCreate(context.Background(), PurchasesCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchasesCreate(context.Background(), PurchasesCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestPurchasesCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/purchases/99")
	mock.Set(OpPurchasesCreate, &fiken.CreatePurchaseCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesCreate(context.Background(), PurchasesCreateIn{
		Company: "acme",
		Body: &fiken.PurchaseRequest{
			Date:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			Kind:     fiken.PurchaseRequestKindCashPurchase,
			Currency: "NOK",
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchasesCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpPurchasesCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesCreate(context.Background(), PurchasesCreateIn{
		Company: "acme",
		Body: &fiken.PurchaseRequest{
			Date:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			Kind:     fiken.PurchaseRequestKindCashPurchase,
			Currency: "NOK",
		},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestPurchasesDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpPurchasesDelete, &fiken.PurchaseResult{
		PurchaseId: fiken.OptInt64{Value: 99, Set: true},
		Date:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Kind:       fiken.PurchaseResultKindCashPurchase,
		Currency:   "NOK",
		Deleted:    fiken.OptBool{Value: true, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesDelete(context.Background(), PurchasesDeleteIn{
		Company: "acme", PurchaseID: 99, Description: "duplicate",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || !res.Ok.Deleted || res.Ok.PurchaseID != 99 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchasesAttachmentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchasesAttachmentsList(context.Background(), PurchasesAttachmentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchasesAttachmentsList(context.Background(), PurchasesAttachmentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing purchase_id: want validation error, got %+v", got)
	}
}

// TestPurchasesAttachmentsListAgainstMock asserts the bare-array happy path.
func TestPurchasesAttachmentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpPurchasesAttachments, []fiken.Attachment{{
		Identifier:  fiken.OptString{Value: "ATT-1", Set: true},
		DownloadUrl: fiken.OptString{Value: "https://example.test/purchase.pdf", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesAttachmentsList(context.Background(), PurchasesAttachmentsListIn{
		Company:    "acme",
		PurchaseID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 || res.Ok.Items[0].Identifier != "ATT-1" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchasesAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchasesAttach(context.Background(), PurchasesAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchasesAttach(context.Background(), PurchasesAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing purchase_id: want validation error, got %+v", got)
	}
	if got := c.PurchasesAttach(context.Background(), PurchasesAttachIn{Company: "acme", PurchaseID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestPurchasesAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/purchases/99/attachments/7")
	mock.Set(OpPurchasesAttach, &fiken.AddAttachmentToPurchaseCreated{
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
	res := c.PurchasesAttach(context.Background(), PurchasesAttachIn{
		Company:         "acme",
		PurchaseID:      99,
		FilePath:        path,
		AttachToPayment: true,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchasesPaymentsListMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchasesPaymentsList(context.Background(), PurchasesPaymentsListIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchasesPaymentsList(context.Background(), PurchasesPaymentsListIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing purchase_id: want validation error, got %+v", got)
	}
}

// TestPurchasesPaymentsListAgainstMock asserts translation of payment
// fields including the date format and the int64 øre amounts.
func TestPurchasesPaymentsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	payDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(OpPurchasesPaymentsList, []fiken.Payment{{
		PaymentId: fiken.OptInt64{Value: 11, Set: true},
		Date:      payDate,
		Account:   "2400:10001",
		Amount:    1000000,
		Currency:  fiken.OptString{Value: "NOK", Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesPaymentsList(context.Background(), PurchasesPaymentsListIn{
		Company:    "acme",
		PurchaseID: 99,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.PaymentID != 11 || got.Date != "2024-06-15" ||
		got.Amount != 1000000 || got.Account != "2400:10001" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

func TestPurchasesPaymentsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchasesPaymentsGet(context.Background(), PurchasesPaymentsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.PurchasesPaymentsGet(context.Background(), PurchasesPaymentsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing purchase_id: want validation error, got %+v", got)
	}
	if got := c.PurchasesPaymentsGet(context.Background(), PurchasesPaymentsGetIn{Company: "x", PurchaseID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing payment_id: want validation error, got %+v", got)
	}
}

// TestPurchasesPaymentsGetAgainstMock asserts the single-resource happy path.
func TestPurchasesPaymentsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	payDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(OpPurchasesPaymentsGet, &fiken.Payment{
		PaymentId: fiken.OptInt64{Value: 11, Set: true},
		Date:      payDate,
		Account:   "2400:10001",
		Amount:    1000000,
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesPaymentsGet(context.Background(), PurchasesPaymentsGetIn{
		Company:    "acme",
		PurchaseID: 99,
		PaymentID:  11,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.PaymentID != 11 || res.Ok.Amount != 1000000 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestPurchasesPaymentsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.PurchasesPaymentsCreate(context.Background(), PurchasesPaymentsCreateIn{Company: "acme", PurchaseID: 99}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestPurchasesPaymentsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/purchases/99/payments/11")
	mock.Set(OpPurchasesPaymentsCreate, &fiken.CreatePurchasePaymentCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.PurchasesPaymentsCreate(context.Background(), PurchasesPaymentsCreateIn{
		Company: "acme", PurchaseID: 99,
		Body: &fiken.Payment{Date: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), Account: "1920", Amount: 1000000},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
