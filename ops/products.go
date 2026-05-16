package ops

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// ProductsListIn carries paged-list input for products. The upstream
// endpoint supports name / productNumber / active filters in addition
// to standard paging; only paging is surfaced through the canonical Out
// shape — narrower filters round-trip through the underlying ogen call
// for callers that need them.
type ProductsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// ProductOut is the canonical product shape exposed to CLI/MCP. The
// unit price stays int64 øre per the global convention. Stock is a
// float upstream (decimals allowed) and round-trips as such. Location
// surfaces the upstream Location header for Create/Update responses.
type ProductOut struct {
	ProductID        int64   `json:"product_id,omitempty"`
	Name             string  `json:"name"`
	ProductNumber    string  `json:"product_number,omitempty"`
	UnitPrice        int64   `json:"unit_price,omitempty"`
	IncomeAccount    string  `json:"income_account,omitempty"`
	VatType          string  `json:"vat_type,omitempty"`
	Active           bool    `json:"active"`
	Stock            float64 `json:"stock,omitempty"`
	Note             string  `json:"note,omitempty"`
	CreatedDate      Date    `json:"created_date,omitempty"`
	LastModifiedDate Date    `json:"last_modified_date,omitempty"`
	Location         string  `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (ProductOut) TableHeader() []string {
	return []string{"ID", "NAME", "NUMBER", "UNIT_PRICE", "ACTIVE"}
}

// TableRow implements output.tableRow.
func (p ProductOut) TableRow() []string {
	return []string{
		strconv.FormatInt(p.ProductID, 10),
		p.Name,
		p.ProductNumber,
		strconv.FormatInt(p.UnitPrice, 10),
		strconv.FormatBool(p.Active),
	}
}

// ProductsListOut is the paged response.
type ProductsListOut = ListOut[ProductOut]

// ProductsList returns products for the specified company.
func (c *Client) ProductsList(ctx context.Context, in ProductsListIn) Result[ProductsListOut] {
	if in.Company == "" {
		return Err[ProductsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProductsList,
		})
	}
	params := fiken.GetProductsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetProducts(ctx, params)
	if err != nil {
		return Err[ProductsListOut](MapErr(OpProductsList, err))
	}
	return Ok[ProductsListOut](translateProductsList(resp))
}

// translateProductsList converts the ogen response into the canonical
// ListOut[ProductOut] envelope, including paging meta.
func translateProductsList(resp *fiken.GetProductsOKHeaders) ProductsListOut {
	if resp == nil {
		return ProductsListOut{Items: []ProductOut{}, Meta: ListMeta{}}
	}
	items := make([]ProductOut, 0, len(resp.Response))
	for _, p := range resp.Response {
		items = append(items, productToOut(p))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return ProductsListOut{Items: items, Meta: meta}
}

// productToOut maps fiken.Product into the canonical ProductOut. Stock
// rounds float32 → float64 verbatim; callers that need higher precision
// can drop down to the underlying type via the JSON envelope.
func productToOut(p fiken.Product) ProductOut {
	out := ProductOut{
		ProductID:     p.ProductId.Or(0),
		Name:          p.Name,
		ProductNumber: p.ProductNumber.Or(""),
		UnitPrice:     p.UnitPrice.Or(0),
		IncomeAccount: p.IncomeAccount,
		VatType:       canonicalVatType(p.VatType),
		Active:        p.Active,
		Stock:         float64(p.Stock.Or(0)),
		Note:          p.Note.Or(""),
	}
	if d, ok := p.CreatedDate.Get(); ok {
		out.CreatedDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := p.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(d.Format("2006-01-02"))
	}
	return out
}

// ProductsGetIn requires company + product id.
type ProductsGetIn struct {
	Company   string `json:"company"`
	ProductID int64  `json:"product_id"`
}

// ProductsGet returns a single product by id.
func (c *Client) ProductsGet(ctx context.Context, in ProductsGetIn) Result[ProductOut] {
	if in.Company == "" {
		return Err[ProductOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProductsGet,
		})
	}
	if in.ProductID == 0 {
		return Err[ProductOut](&Error{
			Code: CodeValidation, Message: "product_id is required", Op: OpProductsGet,
		})
	}
	resp, err := c.gen.GetProduct(ctx, fiken.GetProductParams{
		CompanySlug: in.Company,
		ProductId:   in.ProductID,
	})
	if err != nil {
		return Err[ProductOut](MapErr(OpProductsGet, err))
	}
	if resp == nil {
		return Ok[ProductOut](ProductOut{})
	}
	return Ok[ProductOut](productToOut(*resp))
}

// ProductsCreateIn carries the create payload alongside the company.
// Body is the upstream fiken.Product shape so the field surface stays
// in lock-step with the spec.
type ProductsCreateIn struct {
	Company string         `json:"company"`
	Body    *fiken.Product `json:"body"`
}

// ProductsCreate posts a new product. Upstream returns 201 with a
// Location header pointing at the new resource; surfaced via
// ProductOut.Location.
func (c *Client) ProductsCreate(ctx context.Context, in ProductsCreateIn) Result[ProductOut] {
	if in.Company == "" {
		return Err[ProductOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProductsCreate,
		})
	}
	if in.Body == nil {
		return Err[ProductOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpProductsCreate,
		})
	}
	resp, err := c.gen.CreateProduct(ctx, in.Body, fiken.CreateProductParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[ProductOut](MapErr(OpProductsCreate, err))
	}
	out := ProductOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ProductOut](out)
}

// ProductsUpdateIn carries the update payload + identifiers.
type ProductsUpdateIn struct {
	Company   string         `json:"company"`
	ProductID int64          `json:"product_id"`
	Body      *fiken.Product `json:"body"`
}

// ProductsUpdate replaces an existing product in place. Upstream
// returns 200 with the Location header pointing back at the resource.
func (c *Client) ProductsUpdate(ctx context.Context, in ProductsUpdateIn) Result[ProductOut] {
	if in.Company == "" {
		return Err[ProductOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProductsUpdate,
		})
	}
	if in.ProductID == 0 {
		return Err[ProductOut](&Error{
			Code: CodeValidation, Message: "product_id is required", Op: OpProductsUpdate,
		})
	}
	if in.Body == nil {
		return Err[ProductOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpProductsUpdate,
		})
	}
	resp, err := c.gen.UpdateProduct(ctx, in.Body, fiken.UpdateProductParams{
		CompanySlug: in.Company,
		ProductId:   in.ProductID,
	})
	if err != nil {
		return Err[ProductOut](MapErr(OpProductsUpdate, err))
	}
	out := ProductOut{ProductID: in.ProductID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ProductOut](out)
}

// ProductsDeleteIn identifies the product to remove.
type ProductsDeleteIn struct {
	Company   string `json:"company"`
	ProductID int64  `json:"product_id"`
}

// ProductsDeleteOut is intentionally empty — DELETE returns NoContent.
type ProductsDeleteOut struct{}

// TableHeader implements output.tableRow.
func (ProductsDeleteOut) TableHeader() []string { return []string{"STATUS"} }

// TableRow implements output.tableRow.
func (ProductsDeleteOut) TableRow() []string { return []string{"deleted"} }

// ProductsDelete removes a product. Upstream returns 204 NoContent; we
// surface the success as a zero-value ProductsDeleteOut.
func (c *Client) ProductsDelete(ctx context.Context, in ProductsDeleteIn) Result[ProductsDeleteOut] {
	if in.Company == "" {
		return Err[ProductsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProductsDelete,
		})
	}
	if in.ProductID == 0 {
		return Err[ProductsDeleteOut](&Error{
			Code: CodeValidation, Message: "product_id is required", Op: OpProductsDelete,
		})
	}
	if err := c.gen.DeleteProduct(ctx, fiken.DeleteProductParams{
		CompanySlug: in.Company,
		ProductId:   in.ProductID,
	}); err != nil {
		return Err[ProductsDeleteOut](MapErr(OpProductsDelete, err))
	}
	return Ok[ProductsDeleteOut](ProductsDeleteOut{})
}

// ProductsSalesReportCreateIn carries the report-window payload.
// From/To are inclusive date strings (YYYY-MM-DD). The underlying
// endpoint is POST despite reading data, so the op is classified as
// mutating in the OAS — callers in read-only MCP mode see the op
// filtered out.
type ProductsSalesReportCreateIn struct {
	Company string `json:"company"`
	From    Date   `json:"from"`
	To      Date   `json:"to"`
}

// ProductSalesLineInfoOut mirrors fiken.ProductSalesLineInfo. Monetary
// fields stay int64 øre per global convention.
type ProductSalesLineInfoOut struct {
	Count       int64 `json:"count,omitempty"`
	Sales       int64 `json:"sales,omitempty"`
	NetAmount   int64 `json:"net_amount,omitempty"`
	VatAmount   int64 `json:"vat_amount,omitempty"`
	GrossAmount int64 `json:"gross_amount,omitempty"`
}

// ProductSalesReportRowOut is the canonical sales-report row exposed to
// CLI/MCP. Each row pairs a product with its sold / credited / sum
// aggregates over the requested window.
type ProductSalesReportRowOut struct {
	Product  ProductOut              `json:"product"`
	Sold     ProductSalesLineInfoOut `json:"sold"`
	Credited ProductSalesLineInfoOut `json:"credited"`
	Sum      ProductSalesLineInfoOut `json:"sum"`
}

// TableHeader implements output.tableRow. We surface the most
// useful aggregate columns; deeper detail stays in the JSON envelope.
func (ProductSalesReportRowOut) TableHeader() []string {
	return []string{"PRODUCT_ID", "NAME", "COUNT", "GROSS", "NET"}
}

// TableRow implements output.tableRow. Uses the sum aggregate for
// the monetary columns since "total over the window" is what callers
// most often want at a glance.
func (r ProductSalesReportRowOut) TableRow() []string {
	return []string{
		strconv.FormatInt(r.Product.ProductID, 10),
		r.Product.Name,
		strconv.FormatInt(r.Sum.Count, 10),
		strconv.FormatInt(r.Sum.GrossAmount, 10),
		strconv.FormatInt(r.Sum.NetAmount, 10),
	}
}

// ProductsSalesReportCreateOut wraps the report rows in a list-style
// envelope so the renderer can table-print them with shared paging
// metadata semantics (Returned only; the endpoint is not paged).
type ProductsSalesReportCreateOut = ListOut[ProductSalesReportRowOut]

// ProductsSalesReportCreate posts a sales-report window and returns
// the per-product aggregates. The op is POST upstream; in read-only
// MCP mode it is filtered out by AllowOp.
func (c *Client) ProductsSalesReportCreate(ctx context.Context, in ProductsSalesReportCreateIn) Result[ProductsSalesReportCreateOut] {
	if in.Company == "" {
		return Err[ProductsSalesReportCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProductsSalesReportCreate,
		})
	}
	if in.From == "" {
		return Err[ProductsSalesReportCreateOut](&Error{
			Code: CodeValidation, Message: "from is required (YYYY-MM-DD)", Op: OpProductsSalesReportCreate,
		})
	}
	if in.To == "" {
		return Err[ProductsSalesReportCreateOut](&Error{
			Code: CodeValidation, Message: "to is required (YYYY-MM-DD)", Op: OpProductsSalesReportCreate,
		})
	}
	from, err := parseDate(string(in.From))
	if err != nil {
		return Err[ProductsSalesReportCreateOut](&Error{
			Code: CodeValidation, Message: "from must be YYYY-MM-DD", Op: OpProductsSalesReportCreate,
		})
	}
	to, err := parseDate(string(in.To))
	if err != nil {
		return Err[ProductsSalesReportCreateOut](&Error{
			Code: CodeValidation, Message: "to must be YYYY-MM-DD", Op: OpProductsSalesReportCreate,
		})
	}
	body := &fiken.ProductSalesReportRequest{From: from, To: to}
	resp, err := c.gen.CreateProductSalesReport(ctx, body, fiken.CreateProductSalesReportParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[ProductsSalesReportCreateOut](MapErr(OpProductsSalesReportCreate, err))
	}
	return Ok[ProductsSalesReportCreateOut](translateProductSalesReport(resp))
}

// translateProductSalesReport maps the upstream slice into the
// canonical envelope. Nil rows are skipped defensively.
func translateProductSalesReport(rows []fiken.ProductSalesReportResult) ProductsSalesReportCreateOut {
	items := make([]ProductSalesReportRowOut, 0, len(rows))
	for _, r := range rows {
		items = append(items, productSalesReportRowToOut(r))
	}
	return ProductsSalesReportCreateOut{Items: items, Meta: ListMeta{Returned: len(items)}}
}

func productSalesReportRowToOut(r fiken.ProductSalesReportResult) ProductSalesReportRowOut {
	out := ProductSalesReportRowOut{}
	if p, ok := r.Product.Get(); ok {
		out.Product = productToOut(p)
	}
	if s, ok := r.Sold.Get(); ok {
		out.Sold = productSalesLineToOut(s)
	}
	if s, ok := r.Credited.Get(); ok {
		out.Credited = productSalesLineToOut(s)
	}
	if s, ok := r.Sum.Get(); ok {
		out.Sum = productSalesLineToOut(s)
	}
	return out
}

func productSalesLineToOut(l fiken.ProductSalesLineInfo) ProductSalesLineInfoOut {
	return ProductSalesLineInfoOut{
		Count:       l.Count.Or(0),
		Sales:       l.Sales.Or(0),
		NetAmount:   l.NetAmount.Or(0),
		VatAmount:   l.VatAmount.Or(0),
		GrossAmount: l.GrossAmount.Or(0),
	}
}
