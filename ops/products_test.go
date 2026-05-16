package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestProductsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsList(context.Background(), ProductsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpProductsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpProductsList)
	}
}

func TestProductsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ProductsGet(context.Background(), ProductsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ProductsGet(context.Background(), ProductsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing product_id: want validation error, got %+v", got)
	}
}

// TestProductsListAgainstMock exercises the default empty-list path.
func TestProductsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsList(context.Background(), ProductsListIn{Company: "acme"})
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

// TestProductsListAgainstMockOverride asserts translation of the
// upstream scalar fields and the unit-price → int64 øre convention.
func TestProductsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	created := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(OpProductsList, &fiken.GetProductsOKHeaders{
		Response: []fiken.Product{{
			ProductId:     fiken.OptInt64{Value: 7, Set: true},
			Name:          "Spade",
			ProductNumber: fiken.OptString{Value: "125-1", Set: true},
			UnitPrice:     fiken.OptInt64{Value: 300000, Set: true},
			IncomeAccount: "3000",
			VatType:       "HIGH",
			Active:        true,
			Stock:         fiken.OptFloat32{Value: 5.5, Set: true},
			CreatedDate:   fiken.OptDate{Value: created, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsList(context.Background(), ProductsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.ProductID != 7 || got.Name != "Spade" || got.ProductNumber != "125-1" ||
		got.UnitPrice != 300000 || got.IncomeAccount != "3000" ||
		got.VatType != "high" || !got.Active || got.CreatedDate != "2024-01-15" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestProductsGetAgainstMock asserts the single-resource happy path.
func TestProductsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpProductsGet, &fiken.Product{
		ProductId:     fiken.OptInt64{Value: 7, Set: true},
		Name:          "Spade",
		IncomeAccount: "3000",
		VatType:       "HIGH",
		Active:        true,
		UnitPrice:     fiken.OptInt64{Value: 300000, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsGet(context.Background(), ProductsGetIn{Company: "acme", ProductID: 7})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ProductID != 7 || res.Ok.Name != "Spade" || res.Ok.UnitPrice != 300000 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestProductsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ProductsCreate(context.Background(), ProductsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ProductsCreate(context.Background(), ProductsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestProductsCreateAgainstMock asserts the create path renders the
// Location header through ProductOut.Location.
func TestProductsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/products/9001")
	mock.Set(OpProductsCreate, &fiken.CreateProductCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsCreate(context.Background(), ProductsCreateIn{
		Company: "acme",
		Body:    &fiken.Product{Name: "Spade", IncomeAccount: "3000", VatType: "HIGH", Active: true},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestProductsCreateError verifies upstream non-2xx maps through
// MapErr via mockfiken.SetError.
func TestProductsCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpProductsCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsCreate(context.Background(), ProductsCreateIn{
		Company: "acme",
		Body:    &fiken.Product{Name: "Spade", IncomeAccount: "3000", VatType: "HIGH", Active: true},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

func TestProductsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/products/7")
	mock.Set(OpProductsUpdate, &fiken.UpdateProductOK{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsUpdate(context.Background(), ProductsUpdateIn{
		Company: "acme", ProductID: 7,
		Body: &fiken.Product{Name: "Spade", IncomeAccount: "3000", VatType: "HIGH", Active: true},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ProductID != 7 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestProductsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	_ = mock

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsDelete(context.Background(), ProductsDeleteIn{Company: "acme", ProductID: 7})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("unexpected nil ok")
	}
}

func TestProductsSalesReportCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ProductsSalesReportCreate(context.Background(), ProductsSalesReportCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ProductsSalesReportCreate(context.Background(), ProductsSalesReportCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing from: want validation error, got %+v", got)
	}
	if got := c.ProductsSalesReportCreate(context.Background(), ProductsSalesReportCreateIn{Company: "acme", From: "2024-01-01"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing to: want validation error, got %+v", got)
	}
	if got := c.ProductsSalesReportCreate(context.Background(), ProductsSalesReportCreateIn{Company: "acme", From: "not-a-date", To: "2024-12-31"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("bad from: want validation error, got %+v", got)
	}
}

// TestProductsSalesReportCreateAgainstMock asserts a populated row
// flows through the translation layer correctly.
func TestProductsSalesReportCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpProductsSalesReportCreate, []fiken.ProductSalesReportResult{{
		Product: fiken.OptProduct{Value: fiken.Product{
			ProductId:     fiken.OptInt64{Value: 7, Set: true},
			Name:          "Spade",
			IncomeAccount: "3000",
			VatType:       "HIGH",
			Active:        true,
		}, Set: true},
		Sold: fiken.OptProductSalesLineInfo{Value: fiken.ProductSalesLineInfo{
			Count:       fiken.OptInt64{Value: 10, Set: true},
			Sales:       fiken.OptInt64{Value: 3, Set: true},
			NetAmount:   fiken.OptInt64{Value: 3000000, Set: true},
			VatAmount:   fiken.OptInt64{Value: 750000, Set: true},
			GrossAmount: fiken.OptInt64{Value: 3750000, Set: true},
		}, Set: true},
		Sum: fiken.OptProductSalesLineInfo{Value: fiken.ProductSalesLineInfo{
			Count:       fiken.OptInt64{Value: 10, Set: true},
			GrossAmount: fiken.OptInt64{Value: 3750000, Set: true},
			NetAmount:   fiken.OptInt64{Value: 3000000, Set: true},
		}, Set: true},
	}})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProductsSalesReportCreate(context.Background(), ProductsSalesReportCreateIn{
		Company: "acme",
		From:    "2024-01-01",
		To:      "2024-12-31",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 row, got %+v", res.Ok)
	}
	row := res.Ok.Items[0]
	if row.Product.ProductID != 7 || row.Product.Name != "Spade" {
		t.Fatalf("product translation mismatch: %+v", row.Product)
	}
	if row.Sold.Count != 10 || row.Sold.GrossAmount != 3750000 {
		t.Fatalf("sold translation mismatch: %+v", row.Sold)
	}
	if row.Sum.GrossAmount != 3750000 || row.Sum.NetAmount != 3000000 {
		t.Fatalf("sum translation mismatch: %+v", row.Sum)
	}
}
