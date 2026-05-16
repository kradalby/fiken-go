package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// SalesListIn carries paged-list input for sales. The upstream
// endpoint exposes a rich filter surface — date / last-modified
// bounds, sale-number lookup, settled flag, and customer filter.
// Money-axis filters are not exposed upstream so the struct stays
// declarative.
type SalesListIn struct {
	Company        string `json:"company"`
	Page           int    `json:"page,omitempty"`
	PageSize       int    `json:"page_size,omitempty"`
	Date           Date   `json:"date,omitempty"`
	DateLe         Date   `json:"date_le,omitempty"`
	DateLt         Date   `json:"date_lt,omitempty"`
	DateGe         Date   `json:"date_ge,omitempty"`
	DateGt         Date   `json:"date_gt,omitempty"`
	LastModified   Date   `json:"last_modified,omitempty"`
	LastModifiedLe Date   `json:"last_modified_le,omitempty"`
	LastModifiedLt Date   `json:"last_modified_lt,omitempty"`
	LastModifiedGe Date   `json:"last_modified_ge,omitempty"`
	LastModifiedGt Date   `json:"last_modified_gt,omitempty"`
	SaleNumber     string `json:"sale_number,omitempty"`
	Settled        *bool  `json:"settled,omitempty"`
	ContactID      int64  `json:"contact_id,omitempty"`
}

// OrderLineOut is the canonical order-line shape used by sales and
// purchases. Monetary fields stay int64 øre per the global money
// convention; VatType surfaces the upstream enum verbatim.
type OrderLineOut struct {
	Description        string `json:"description,omitempty"`
	NetPrice           int64  `json:"net_price,omitempty"`
	Vat                int64  `json:"vat,omitempty"`
	Account            string `json:"account,omitempty"`
	VatType            string `json:"vat_type,omitempty"`
	NetPriceInCurrency int64  `json:"net_price_in_currency,omitempty"`
	VatInCurrency      int64  `json:"vat_in_currency,omitempty"`
	ProjectID          int64  `json:"project_id,omitempty"`
}

// orderLineToOut maps fiken.OrderLine into the canonical OrderLineOut.
func orderLineToOut(ln fiken.OrderLine) OrderLineOut {
	return OrderLineOut{
		Description:        ln.Description,
		NetPrice:           ln.NetPrice.Or(0),
		Vat:                ln.Vat.Or(0),
		Account:            ln.Account.Or(""),
		VatType:            canonicalVatType(ln.VatType),
		NetPriceInCurrency: ln.NetPriceInCurrency.Or(0),
		VatInCurrency:      ln.VatInCurrency.Or(0),
		ProjectID:          ln.ProjectId.Or(0),
	}
}

// PaymentOut is the canonical payment shape exposed to CLI/MCP for the
// sale-payment + purchase-payment endpoints. Amount stays int64 øre.
type PaymentOut struct {
	PaymentID   int64  `json:"payment_id,omitempty"`
	Date        Date   `json:"date,omitempty"`
	Account     string `json:"account,omitempty"`
	Amount      int64  `json:"amount,omitempty"`
	AmountInNok int64  `json:"amount_in_nok,omitempty"`
	Currency    string `json:"currency,omitempty"`
	Fee         int64  `json:"fee,omitempty"`
}

// TableHeader implements output.tableRow.
func (PaymentOut) TableHeader() []string {
	return []string{"PAYMENT_ID", "DATE", "AMOUNT", "ACCOUNT"}
}

// TableRow implements output.tableRow.
func (p PaymentOut) TableRow() []string {
	return []string{
		strconv.FormatInt(p.PaymentID, 10),
		string(p.Date),
		strconv.FormatInt(p.Amount, 10),
		p.Account,
	}
}

// paymentToOut maps fiken.Payment into the canonical PaymentOut.
func paymentToOut(p fiken.Payment) PaymentOut {
	return PaymentOut{
		PaymentID:   p.PaymentId.Or(0),
		Date:        Date(p.Date.Format("2006-01-02")),
		Account:     p.Account,
		Amount:      p.Amount,
		AmountInNok: p.AmountInNok.Or(0),
		Currency:    p.Currency.Or(""),
		Fee:         p.Fee.Or(0),
	}
}

// PaymentsListOut is the bare-array response for payment lists.
type PaymentsListOut = ListOut[PaymentOut]

// SaleOut is the canonical sale shape exposed to CLI/MCP. Monetary
// fields stay int64 øre. Customer is flattened to id + name so the
// table renderer has a meaningful one-liner. Lines use OrderLineOut
// (shared with purchases) since the upstream type is `orderLine`,
// distinct from invoiceLineResult. Payments + attachments are
// surfaced inline since the upstream payload includes them.
type SaleOut struct {
	SaleID                       int64           `json:"sale_id,omitempty"`
	SaleNumber                   string          `json:"sale_number,omitempty"`
	Date                         Date            `json:"date,omitempty"`
	DueDate                      Date            `json:"due_date,omitempty"`
	PaymentDate                  Date            `json:"payment_date,omitempty"`
	SettledDate                  Date            `json:"settled_date,omitempty"`
	LastModifiedDate             Date            `json:"last_modified_date,omitempty"`
	Kind                         string          `json:"kind,omitempty"`
	NetAmount                    int64           `json:"net_amount,omitempty"`
	VatAmount                    int64           `json:"vat_amount,omitempty"`
	TotalPaid                    int64           `json:"total_paid,omitempty"`
	TotalPaidInCurrency          int64           `json:"total_paid_in_currency,omitempty"`
	OutstandingBalance           int64           `json:"outstanding_balance,omitempty"`
	OutstandingBalanceInCurrency int64           `json:"outstanding_balance_in_currency,omitempty"`
	Settled                      bool            `json:"settled,omitempty"`
	WriteOff                     bool            `json:"write_off,omitempty"`
	Deleted                      bool            `json:"deleted,omitempty"`
	Currency                     string          `json:"currency,omitempty"`
	Kid                          string          `json:"kid,omitempty"`
	TransactionID                int64           `json:"transaction_id,omitempty"`
	CustomerID                   int64           `json:"customer_id,omitempty"`
	CustomerName                 string          `json:"customer_name,omitempty"`
	Lines                        []OrderLineOut  `json:"lines,omitempty"`
	Payments                     []PaymentOut    `json:"payments,omitempty"`
	Attachments                  []AttachmentOut `json:"attachments,omitempty"`
}

// TableHeader implements output.tableRow.
func (SaleOut) TableHeader() []string {
	return []string{"ID", "NUMBER", "DATE", "KIND", "NET", "OUTSTANDING", "CUSTOMER"}
}

// TableRow implements output.tableRow.
func (s SaleOut) TableRow() []string {
	return []string{
		strconv.FormatInt(s.SaleID, 10),
		s.SaleNumber,
		string(s.Date),
		s.Kind,
		strconv.FormatInt(s.NetAmount, 10),
		strconv.FormatInt(s.OutstandingBalance, 10),
		s.CustomerName,
	}
}

// SalesListOut is the paged response.
type SalesListOut = ListOut[SaleOut]

// SalesList returns sales for the specified company.
func (c *Client) SalesList(ctx context.Context, in SalesListIn) Result[SalesListOut] {
	if in.Company == "" {
		return Err[SalesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesList,
		})
	}
	params := fiken.GetSalesParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	setDateParam(&params.Date, in.Date)
	setDateParam(&params.DateLe, in.DateLe)
	setDateParam(&params.DateLt, in.DateLt)
	setDateParam(&params.DateGe, in.DateGe)
	setDateParam(&params.DateGt, in.DateGt)
	setDateParam(&params.LastModified, in.LastModified)
	setDateParam(&params.LastModifiedLe, in.LastModifiedLe)
	setDateParam(&params.LastModifiedLt, in.LastModifiedLt)
	setDateParam(&params.LastModifiedGe, in.LastModifiedGe)
	setDateParam(&params.LastModifiedGt, in.LastModifiedGt)
	if in.SaleNumber != "" {
		params.SaleNumber.SetTo(in.SaleNumber)
	}
	if in.Settled != nil {
		params.Settled.SetTo(*in.Settled)
	}
	if in.ContactID != 0 {
		params.ContactId.SetTo(in.ContactID)
	}
	resp, err := c.gen.GetSales(ctx, params)
	if err != nil {
		return Err[SalesListOut](MapErr(OpSalesList, err))
	}
	return Ok[SalesListOut](translateSalesList(resp))
}

// translateSalesList converts the ogen response into the canonical
// ListOut[SaleOut] envelope, including paging meta.
func translateSalesList(resp *fiken.GetSalesOKHeaders) SalesListOut {
	if resp == nil {
		return SalesListOut{Items: []SaleOut{}, Meta: ListMeta{}}
	}
	items := make([]SaleOut, 0, len(resp.Response))
	for _, s := range resp.Response {
		items = append(items, saleToOut(s))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return SalesListOut{Items: items, Meta: meta}
}

// saleToOut maps fiken.SaleResult into the canonical SaleOut. Customer
// is flattened to id + name; payments + attachments are surfaced
// inline since they round-trip with the sale.
func saleToOut(s fiken.SaleResult) SaleOut {
	out := SaleOut{
		SaleID:                       s.SaleId.Or(0),
		SaleNumber:                   s.SaleNumber.Or(""),
		NetAmount:                    s.NetAmount.Or(0),
		VatAmount:                    s.VatAmount.Or(0),
		TotalPaid:                    s.TotalPaid.Or(0),
		TotalPaidInCurrency:          s.TotalPaidInCurrency.Or(0),
		OutstandingBalance:           s.OutstandingBalance.Or(0),
		OutstandingBalanceInCurrency: s.OutstandingBalanceInCurrency.Or(0),
		Settled:                      s.Settled.Or(false),
		WriteOff:                     s.WriteOff.Or(false),
		Deleted:                      s.Deleted.Or(false),
		Currency:                     s.Currency.Or(""),
		Kid:                          s.Kid.Or(""),
		TransactionID:                s.TransactionId.Or(0),
	}
	if k, ok := s.Kind.Get(); ok {
		out.Kind = string(k)
	}
	if d, ok := s.Date.Get(); ok {
		out.Date = Date(d.Format("2006-01-02"))
	}
	if d, ok := s.DueDate.Get(); ok {
		out.DueDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := s.PaymentDate.Get(); ok {
		out.PaymentDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := s.SettledDate.Get(); ok {
		out.SettledDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := s.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(d.Format("2006-01-02"))
	}
	if cust, ok := s.Customer.Get(); ok {
		out.CustomerID = cust.ContactId.Or(0)
		out.CustomerName = cust.Name
	}
	if len(s.Lines) > 0 {
		out.Lines = make([]OrderLineOut, 0, len(s.Lines))
		for _, ln := range s.Lines {
			out.Lines = append(out.Lines, orderLineToOut(ln))
		}
	}
	if len(s.SalePayments) > 0 {
		out.Payments = make([]PaymentOut, 0, len(s.SalePayments))
		for _, p := range s.SalePayments {
			out.Payments = append(out.Payments, paymentToOut(p))
		}
	}
	if len(s.SaleAttachments) > 0 {
		out.Attachments = make([]AttachmentOut, 0, len(s.SaleAttachments))
		for _, a := range s.SaleAttachments {
			out.Attachments = append(out.Attachments, attachmentToOut(a))
		}
	}
	return out
}

// SalesGetIn requires company + sale id.
type SalesGetIn struct {
	Company string `json:"company"`
	SaleID  int64  `json:"sale_id"`
}

// SalesGet returns a single sale by id.
func (c *Client) SalesGet(ctx context.Context, in SalesGetIn) Result[SaleOut] {
	if in.Company == "" {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesGet,
		})
	}
	if in.SaleID == 0 {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesGet,
		})
	}
	resp, err := c.gen.GetSale(ctx, fiken.GetSaleParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
	})
	if err != nil {
		return Err[SaleOut](MapErr(OpSalesGet, err))
	}
	if resp == nil {
		return Ok[SaleOut](SaleOut{})
	}
	return Ok[SaleOut](saleToOut(*resp))
}

// SalesCreateIn carries the create payload alongside the company.
// Body mirrors the upstream SaleRequest. Currently stubbed — the
// CLI/MCP surfaces register the op so help and discovery stay
// complete; mutating wiring lands in a follow-up task.
type SalesCreateIn struct {
	Company string             `json:"company"`
	Body    *fiken.SaleRequest `json:"body"`
}

// SalesCreateOut surfaces the Location header pointing at the new sale.
type SalesCreateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (SalesCreateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o SalesCreateOut) TableRow() []string { return []string{o.Location} }

// SalesCreate posts a new sale. Upstream returns 201 with a Location
// header pointing at the new resource; surfaced via
// SalesCreateOut.Location.
func (c *Client) SalesCreate(ctx context.Context, in SalesCreateIn) Result[SalesCreateOut] {
	if in.Company == "" {
		return Err[SalesCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesCreate,
		})
	}
	if in.Body == nil {
		return Err[SalesCreateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpSalesCreate,
		})
	}
	resp, err := c.gen.CreateSale(ctx, in.Body, fiken.CreateSaleParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[SalesCreateOut](MapErr(OpSalesCreate, err))
	}
	out := SalesCreateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[SalesCreateOut](out)
}

// SalesDeleteIn identifies the sale to mark deleted. Description is a
// required query parameter upstream (audit trail).
type SalesDeleteIn struct {
	Company     string `json:"company"`
	SaleID      int64  `json:"sale_id"`
	Description string `json:"description"`
}

// SalesDeleteOut surfaces the post-delete sale snapshot.
type SalesDeleteOut = SaleOut

// SalesDelete soft-archives a sale. Despite the verb, the upstream
// route is a PATCH that takes a required description query param
// (audit trail) and returns the post-delete sale snapshot.
func (c *Client) SalesDelete(ctx context.Context, in SalesDeleteIn) Result[SalesDeleteOut] {
	if in.Company == "" {
		return Err[SalesDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDelete,
		})
	}
	if in.SaleID == 0 {
		return Err[SalesDeleteOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesDelete,
		})
	}
	if in.Description == "" {
		return Err[SalesDeleteOut](&Error{
			Code: CodeValidation, Message: "description is required", Op: OpSalesDelete,
		})
	}
	resp, err := c.gen.DeleteSale(ctx, fiken.DeleteSaleParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
		Description: in.Description,
	})
	if err != nil {
		return Err[SalesDeleteOut](MapErr(OpSalesDelete, err))
	}
	if resp == nil {
		return Ok[SalesDeleteOut](SalesDeleteOut{})
	}
	return Ok[SalesDeleteOut](saleToOut(*resp))
}

// SalesSettleIn identifies the sale to mark settled-without-payment.
// SettledDate is a required query param upstream.
type SalesSettleIn struct {
	Company     string `json:"company"`
	SaleID      int64  `json:"sale_id"`
	SettledDate Date   `json:"settled_date"`
}

// SalesSettle marks a sale settled-without-payment. The upstream route
// is PATCH with a required settledDate query param and returns the
// post-settle sale snapshot.
func (c *Client) SalesSettle(ctx context.Context, in SalesSettleIn) Result[SaleOut] {
	if in.Company == "" {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesSettle,
		})
	}
	if in.SaleID == 0 {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesSettle,
		})
	}
	if in.SettledDate == "" {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "settled_date is required (YYYY-MM-DD)", Op: OpSalesSettle,
		})
	}
	settled, err := parseDate(string(in.SettledDate))
	if err != nil {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "settled_date must be YYYY-MM-DD", Op: OpSalesSettle,
		})
	}
	resp, err := c.gen.SettledSale(ctx, fiken.SettledSaleParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
		SettledDate: settled,
	})
	if err != nil {
		return Err[SaleOut](MapErr(OpSalesSettle, err))
	}
	if resp == nil {
		return Ok[SaleOut](SaleOut{})
	}
	return Ok[SaleOut](saleToOut(*resp))
}

// SalesWriteOffIn carries the write-off (tapsføring) payload.
type SalesWriteOffIn struct {
	Company string                 `json:"company"`
	SaleID  int64                  `json:"sale_id"`
	Body    *fiken.WriteOffRequest `json:"body"`
}

// SalesWriteOff registers a write-off (tapsføring) for a sale. The
// upstream route is PATCH with a WriteOffRequest body and returns the
// post-write-off sale snapshot.
func (c *Client) SalesWriteOff(ctx context.Context, in SalesWriteOffIn) Result[SaleOut] {
	if in.Company == "" {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesWriteOff,
		})
	}
	if in.SaleID == 0 {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesWriteOff,
		})
	}
	if in.Body == nil {
		return Err[SaleOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpSalesWriteOff,
		})
	}
	resp, err := c.gen.WriteOffSale(ctx, in.Body, fiken.WriteOffSaleParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
	})
	if err != nil {
		return Err[SaleOut](MapErr(OpSalesWriteOff, err))
	}
	if resp == nil {
		return Ok[SaleOut](SaleOut{})
	}
	return Ok[SaleOut](saleToOut(*resp))
}

// SalesAttachmentsListIn requires company + sale id.
type SalesAttachmentsListIn struct {
	Company string `json:"company"`
	SaleID  int64  `json:"sale_id"`
}

// SalesAttachmentsList returns all attachments for a sale. The upstream
// endpoint returns a bare array; Meta.Returned is the only field set.
func (c *Client) SalesAttachmentsList(ctx context.Context, in SalesAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesAttachments,
		})
	}
	if in.SaleID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesAttachments,
		})
	}
	resp, err := c.gen.GetSaleAttachments(ctx, fiken.GetSaleAttachmentsParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpSalesAttachments, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// SalesAttachIn carries the multipart upload payload. Stays declarative
// for Plan C; multipart wiring lands in Plan D once EnableAttachments
// is fully threaded. AttachToSale / AttachToPayment mirror the upstream
// query knobs — at least one must be true server-side.
type SalesAttachIn struct {
	Company         string `json:"company"`
	SaleID          int64  `json:"sale_id"`
	Filename        string `json:"filename"`
	FilePath        string `json:"file_path"`
	AttachToSale    bool   `json:"attach_to_sale,omitempty"`
	AttachToPayment bool   `json:"attach_to_payment,omitempty"`
}

// SalesAttachOut mirrors the upstream Created response — a Location URL.
type SalesAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (SalesAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a SalesAttachOut) TableRow() []string { return []string{a.Location} }

// SalesAttach uploads a file to a sale as multipart form data. At
// least one of AttachToSale / AttachToPayment must be true upstream;
// the server rejects the request otherwise. Filename defaults to the
// basename of FilePath; pass Filename to override.
func (c *Client) SalesAttach(ctx context.Context, in SalesAttachIn) Result[SalesAttachOut] {
	if in.Company == "" {
		return Err[SalesAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesAttach,
		})
	}
	if in.SaleID == 0 {
		return Err[SalesAttachOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesAttach,
		})
	}
	if in.FilePath == "" {
		return Err[SalesAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpSalesAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[SalesAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpSalesAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToSaleReq(fiken.AddAttachmentToSaleReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	params := fiken.AddAttachmentToSaleParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
	}
	if in.AttachToSale {
		params.AttachToSale = fiken.NewOptBool(true)
	}
	if in.AttachToPayment {
		params.AttachToPayment = fiken.NewOptBool(true)
	}
	resp, err := c.gen.AddAttachmentToSale(ctx, req, params)
	if err != nil {
		return Err[SalesAttachOut](MapErr(OpSalesAttach, err))
	}
	out := SalesAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[SalesAttachOut](out)
}

// SalesPaymentsListIn requires company + sale id.
type SalesPaymentsListIn struct {
	Company string `json:"company"`
	SaleID  int64  `json:"sale_id"`
}

// SalesPaymentsList returns all payments for a sale. The upstream
// endpoint returns a bare array; Meta.Returned is the only field set.
func (c *Client) SalesPaymentsList(ctx context.Context, in SalesPaymentsListIn) Result[PaymentsListOut] {
	if in.Company == "" {
		return Err[PaymentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesPaymentsList,
		})
	}
	if in.SaleID == 0 {
		return Err[PaymentsListOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesPaymentsList,
		})
	}
	resp, err := c.gen.GetSalePayments(ctx, fiken.GetSalePaymentsParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
	})
	if err != nil {
		return Err[PaymentsListOut](MapErr(OpSalesPaymentsList, err))
	}
	items := make([]PaymentOut, 0, len(resp))
	for _, p := range resp {
		items = append(items, paymentToOut(p))
	}
	return Ok[PaymentsListOut](PaymentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// SalesPaymentsGetIn requires company + sale id + payment id.
type SalesPaymentsGetIn struct {
	Company   string `json:"company"`
	SaleID    int64  `json:"sale_id"`
	PaymentID int64  `json:"payment_id"`
}

// SalesPaymentsGet returns a single payment by id.
func (c *Client) SalesPaymentsGet(ctx context.Context, in SalesPaymentsGetIn) Result[PaymentOut] {
	if in.Company == "" {
		return Err[PaymentOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesPaymentsGet,
		})
	}
	if in.SaleID == 0 {
		return Err[PaymentOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesPaymentsGet,
		})
	}
	if in.PaymentID == 0 {
		return Err[PaymentOut](&Error{
			Code: CodeValidation, Message: "payment_id is required", Op: OpSalesPaymentsGet,
		})
	}
	resp, err := c.gen.GetSalePayment(ctx, fiken.GetSalePaymentParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
		PaymentId:   in.PaymentID,
	})
	if err != nil {
		return Err[PaymentOut](MapErr(OpSalesPaymentsGet, err))
	}
	if resp == nil {
		return Ok[PaymentOut](PaymentOut{})
	}
	return Ok[PaymentOut](paymentToOut(*resp))
}

// SalesPaymentsCreateIn carries the create-payment payload alongside
// the sale identifier. Body mirrors the upstream Payment shape.
type SalesPaymentsCreateIn struct {
	Company string         `json:"company"`
	SaleID  int64          `json:"sale_id"`
	Body    *fiken.Payment `json:"body"`
}

// SalesPaymentsCreateOut surfaces the Location header for the new payment.
type SalesPaymentsCreateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (SalesPaymentsCreateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o SalesPaymentsCreateOut) TableRow() []string { return []string{o.Location} }

// SalesPaymentsCreate registers a payment against a sale. Upstream
// returns 201 with a Location header pointing at the new payment;
// surfaced via SalesPaymentsCreateOut.Location.
func (c *Client) SalesPaymentsCreate(ctx context.Context, in SalesPaymentsCreateIn) Result[SalesPaymentsCreateOut] {
	if in.Company == "" {
		return Err[SalesPaymentsCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesPaymentsCreate,
		})
	}
	if in.SaleID == 0 {
		return Err[SalesPaymentsCreateOut](&Error{
			Code: CodeValidation, Message: "sale_id is required", Op: OpSalesPaymentsCreate,
		})
	}
	if in.Body == nil {
		return Err[SalesPaymentsCreateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpSalesPaymentsCreate,
		})
	}
	resp, err := c.gen.CreateSalePayment(ctx, in.Body, fiken.CreateSalePaymentParams{
		CompanySlug: in.Company,
		SaleId:      in.SaleID,
	})
	if err != nil {
		return Err[SalesPaymentsCreateOut](MapErr(OpSalesPaymentsCreate, err))
	}
	out := SalesPaymentsCreateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[SalesPaymentsCreateOut](out)
}
