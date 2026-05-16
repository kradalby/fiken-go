package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// PurchasesListIn carries paged-list input for purchases. The upstream
// endpoint exposes date + settled-date filters and a paid flag; sort
// is hardcoded to "date asc" by Fiken. Money-axis filters are not
// exposed upstream so the struct stays declarative.
type PurchasesListIn struct {
	Company       string `json:"company"`
	Page          int    `json:"page,omitempty"`
	PageSize      int    `json:"page_size,omitempty"`
	Date          Date   `json:"date,omitempty"`
	DateLe        Date   `json:"date_le,omitempty"`
	DateLt        Date   `json:"date_lt,omitempty"`
	DateGe        Date   `json:"date_ge,omitempty"`
	DateGt        Date   `json:"date_gt,omitempty"`
	SettledDate   Date   `json:"settled_date,omitempty"`
	SettledDateLe Date   `json:"settled_date_le,omitempty"`
	SettledDateLt Date   `json:"settled_date_lt,omitempty"`
	SettledDateGe Date   `json:"settled_date_ge,omitempty"`
	SettledDateGt Date   `json:"settled_date_gt,omitempty"`
	Paid          *bool  `json:"paid,omitempty"`
}

// PurchaseOut is the canonical purchase shape exposed to CLI/MCP.
// Monetary fields stay int64 øre. Supplier is flattened to id + name
// so the table renderer has a meaningful one-liner. Lines reuse
// OrderLineOut (shared with sales) since the upstream type is
// `orderLine`. Payments + attachments are surfaced inline since the
// upstream payload includes them.
type PurchaseOut struct {
	PurchaseID     int64           `json:"purchase_id,omitempty"`
	TransactionID  int64           `json:"transaction_id,omitempty"`
	Identifier     string          `json:"identifier,omitempty"`
	Date           Date            `json:"date,omitempty"`
	DueDate        Date            `json:"due_date,omitempty"`
	PaymentDate    Date            `json:"payment_date,omitempty"`
	SettledDate    Date            `json:"settled_date,omitempty"`
	Kind           string          `json:"kind,omitempty"`
	Paid           bool            `json:"paid,omitempty"`
	Settled        bool            `json:"settled,omitempty"`
	Deleted        bool            `json:"deleted,omitempty"`
	Currency       string          `json:"currency,omitempty"`
	PaymentAccount string          `json:"payment_account,omitempty"`
	Kid            string          `json:"kid,omitempty"`
	SupplierID     int64           `json:"supplier_id,omitempty"`
	SupplierName   string          `json:"supplier_name,omitempty"`
	Lines          []OrderLineOut  `json:"lines,omitempty"`
	Payments       []PaymentOut    `json:"payments,omitempty"`
	Attachments    []AttachmentOut `json:"attachments,omitempty"`
}

// TableHeader implements output.tableRow.
func (PurchaseOut) TableHeader() []string {
	return []string{"ID", "IDENTIFIER", "DATE", "KIND", "PAID", "SUPPLIER"}
}

// TableRow implements output.tableRow.
func (p PurchaseOut) TableRow() []string {
	return []string{
		strconv.FormatInt(p.PurchaseID, 10),
		p.Identifier,
		string(p.Date),
		p.Kind,
		strconv.FormatBool(p.Paid),
		p.SupplierName,
	}
}

// PurchasesListOut is the paged response.
type PurchasesListOut = ListOut[PurchaseOut]

// PurchasesList returns purchases for the specified company.
func (c *Client) PurchasesList(ctx context.Context, in PurchasesListIn) Result[PurchasesListOut] {
	if in.Company == "" {
		return Err[PurchasesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesList,
		})
	}
	params := fiken.GetPurchasesParams{CompanySlug: in.Company}
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
	setDateParam(&params.SettledDate, in.SettledDate)
	setDateParam(&params.SettledDateLe, in.SettledDateLe)
	setDateParam(&params.SettledDateLt, in.SettledDateLt)
	setDateParam(&params.SettledDateGe, in.SettledDateGe)
	setDateParam(&params.SettledDateGt, in.SettledDateGt)
	if in.Paid != nil {
		params.Paid.SetTo(*in.Paid)
	}
	resp, err := c.gen.GetPurchases(ctx, params)
	if err != nil {
		return Err[PurchasesListOut](MapErr(OpPurchasesList, err))
	}
	return Ok[PurchasesListOut](translatePurchasesList(resp))
}

// translatePurchasesList converts the ogen response into the canonical
// ListOut[PurchaseOut] envelope, including paging meta.
func translatePurchasesList(resp *fiken.GetPurchasesOKHeaders) PurchasesListOut {
	if resp == nil {
		return PurchasesListOut{Items: []PurchaseOut{}, Meta: ListMeta{}}
	}
	items := make([]PurchaseOut, 0, len(resp.Response))
	for _, p := range resp.Response {
		items = append(items, purchaseToOut(p))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return PurchasesListOut{Items: items, Meta: meta}
}

// purchaseToOut maps fiken.PurchaseResult into the canonical
// PurchaseOut. Supplier is flattened to id + name; payments +
// attachments are surfaced inline since they round-trip with the
// purchase. Kind is non-Optional upstream and is always set.
func purchaseToOut(p fiken.PurchaseResult) PurchaseOut {
	out := PurchaseOut{
		PurchaseID:     p.PurchaseId.Or(0),
		TransactionID:  p.TransactionId.Or(0),
		Identifier:     p.Identifier.Or(""),
		Date:           Date(p.Date.Format("2006-01-02")),
		Kind:           string(p.Kind),
		Paid:           p.Paid,
		Settled:        p.Settled.Or(false),
		Deleted:        p.Deleted.Or(false),
		Currency:       p.Currency,
		PaymentAccount: p.PaymentAccount.Or(""),
		Kid:            p.Kid.Or(""),
	}
	if d, ok := p.DueDate.Get(); ok {
		out.DueDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := p.PaymentDate.Get(); ok {
		out.PaymentDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := p.SettledDate.Get(); ok {
		out.SettledDate = Date(d.Format("2006-01-02"))
	}
	if supp, ok := p.Supplier.Get(); ok {
		out.SupplierID = supp.ContactId.Or(0)
		out.SupplierName = supp.Name
	}
	if len(p.Lines) > 0 {
		out.Lines = make([]OrderLineOut, 0, len(p.Lines))
		for _, ln := range p.Lines {
			out.Lines = append(out.Lines, orderLineToOut(ln))
		}
	}
	if len(p.Payments) > 0 {
		out.Payments = make([]PaymentOut, 0, len(p.Payments))
		for _, pay := range p.Payments {
			out.Payments = append(out.Payments, paymentToOut(pay))
		}
	}
	if len(p.PurchaseAttachments) > 0 {
		out.Attachments = make([]AttachmentOut, 0, len(p.PurchaseAttachments))
		for _, a := range p.PurchaseAttachments {
			out.Attachments = append(out.Attachments, attachmentToOut(a))
		}
	}
	return out
}

// PurchasesGetIn requires company + purchase id.
type PurchasesGetIn struct {
	Company    string `json:"company"`
	PurchaseID int64  `json:"purchase_id"`
}

// PurchasesGet returns a single purchase by id.
func (c *Client) PurchasesGet(ctx context.Context, in PurchasesGetIn) Result[PurchaseOut] {
	if in.Company == "" {
		return Err[PurchaseOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesGet,
		})
	}
	if in.PurchaseID == 0 {
		return Err[PurchaseOut](&Error{
			Code: CodeValidation, Message: "purchase_id is required", Op: OpPurchasesGet,
		})
	}
	resp, err := c.gen.GetPurchase(ctx, fiken.GetPurchaseParams{
		CompanySlug: in.Company,
		PurchaseId:  in.PurchaseID,
	})
	if err != nil {
		return Err[PurchaseOut](MapErr(OpPurchasesGet, err))
	}
	if resp == nil {
		return Ok[PurchaseOut](PurchaseOut{})
	}
	return Ok[PurchaseOut](purchaseToOut(*resp))
}

// PurchasesCreateIn carries the create payload alongside the company.
// Body mirrors the upstream PurchaseRequest. Currently stubbed — the
// CLI/MCP surfaces register the op so help and discovery stay
// complete; mutating wiring lands in a follow-up task.
type PurchasesCreateIn struct {
	Company string                 `json:"company"`
	Body    *fiken.PurchaseRequest `json:"body"`
}

// PurchasesCreateOut surfaces the Location header pointing at the
// newly-created purchase.
type PurchasesCreateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (PurchasesCreateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o PurchasesCreateOut) TableRow() []string { return []string{o.Location} }

// PurchasesCreate posts a new purchase. Upstream returns 201 with a
// Location header pointing at the new resource; surfaced via
// PurchasesCreateOut.Location.
func (c *Client) PurchasesCreate(ctx context.Context, in PurchasesCreateIn) Result[PurchasesCreateOut] {
	if in.Company == "" {
		return Err[PurchasesCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesCreate,
		})
	}
	if in.Body == nil {
		return Err[PurchasesCreateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpPurchasesCreate,
		})
	}
	resp, err := c.gen.CreatePurchase(ctx, in.Body, fiken.CreatePurchaseParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[PurchasesCreateOut](MapErr(OpPurchasesCreate, err))
	}
	out := PurchasesCreateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[PurchasesCreateOut](out)
}

// PurchasesDeleteIn identifies the purchase to mark deleted.
// Description is a required query parameter upstream (audit trail).
type PurchasesDeleteIn struct {
	Company     string `json:"company"`
	PurchaseID  int64  `json:"purchase_id"`
	Description string `json:"description"`
}

// PurchasesDeleteOut surfaces the post-delete purchase snapshot.
type PurchasesDeleteOut = PurchaseOut

// PurchasesDelete soft-archives a purchase. Despite the verb, the
// upstream route is a PATCH that takes a required description query
// param (audit trail) and returns the post-delete purchase snapshot.
func (c *Client) PurchasesDelete(ctx context.Context, in PurchasesDeleteIn) Result[PurchasesDeleteOut] {
	if in.Company == "" {
		return Err[PurchasesDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDelete,
		})
	}
	if in.PurchaseID == 0 {
		return Err[PurchasesDeleteOut](&Error{
			Code: CodeValidation, Message: "purchase_id is required", Op: OpPurchasesDelete,
		})
	}
	if in.Description == "" {
		return Err[PurchasesDeleteOut](&Error{
			Code: CodeValidation, Message: "description is required", Op: OpPurchasesDelete,
		})
	}
	resp, err := c.gen.DeletePurchase(ctx, fiken.DeletePurchaseParams{
		CompanySlug: in.Company,
		PurchaseId:  in.PurchaseID,
		Description: in.Description,
	})
	if err != nil {
		return Err[PurchasesDeleteOut](MapErr(OpPurchasesDelete, err))
	}
	if resp == nil {
		return Ok[PurchasesDeleteOut](PurchasesDeleteOut{})
	}
	return Ok[PurchasesDeleteOut](purchaseToOut(*resp))
}

// PurchasesAttachmentsListIn requires company + purchase id.
type PurchasesAttachmentsListIn struct {
	Company    string `json:"company"`
	PurchaseID int64  `json:"purchase_id"`
}

// PurchasesAttachmentsList returns all attachments for a purchase. The
// upstream endpoint returns a bare array; Meta.Returned is the only
// field set.
func (c *Client) PurchasesAttachmentsList(ctx context.Context, in PurchasesAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesAttachments,
		})
	}
	if in.PurchaseID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "purchase_id is required", Op: OpPurchasesAttachments,
		})
	}
	resp, err := c.gen.GetPurchaseAttachments(ctx, fiken.GetPurchaseAttachmentsParams{
		CompanySlug: in.Company,
		PurchaseId:  in.PurchaseID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpPurchasesAttachments, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// PurchasesAttachIn carries the multipart upload payload. Stays
// declarative for Plan C; multipart wiring lands in Plan D once
// EnableAttachments is fully threaded. AttachToPurchase /
// AttachToPayment mirror the upstream query knobs — at least one must
// be true server-side (the OpenAPI calls them attachToSale /
// attachToPayment even though the resource is a purchase).
type PurchasesAttachIn struct {
	Company         string `json:"company"`
	PurchaseID      int64  `json:"purchase_id"`
	Filename        string `json:"filename"`
	FilePath        string `json:"file_path"`
	AttachToSale    bool   `json:"attach_to_sale,omitempty"`
	AttachToPayment bool   `json:"attach_to_payment,omitempty"`
}

// PurchasesAttachOut mirrors the upstream Created response — a
// Location URL.
type PurchasesAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (PurchasesAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a PurchasesAttachOut) TableRow() []string { return []string{a.Location} }

// PurchasesAttach uploads a file to a purchase as multipart form
// data. The upstream query params are named attachToSale /
// attachToPayment even though the resource is a purchase; at least
// one must be true server-side. Filename defaults to the basename of
// FilePath; pass Filename to override.
func (c *Client) PurchasesAttach(ctx context.Context, in PurchasesAttachIn) Result[PurchasesAttachOut] {
	if in.Company == "" {
		return Err[PurchasesAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesAttach,
		})
	}
	if in.PurchaseID == 0 {
		return Err[PurchasesAttachOut](&Error{
			Code: CodeValidation, Message: "purchase_id is required", Op: OpPurchasesAttach,
		})
	}
	if in.FilePath == "" {
		return Err[PurchasesAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpPurchasesAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[PurchasesAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpPurchasesAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToPurchaseReq(fiken.AddAttachmentToPurchaseReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	params := fiken.AddAttachmentToPurchaseParams{
		CompanySlug: in.Company,
		PurchaseId:  in.PurchaseID,
	}
	if in.AttachToSale {
		params.AttachToSale = fiken.NewOptBool(true)
	}
	if in.AttachToPayment {
		params.AttachToPayment = fiken.NewOptBool(true)
	}
	resp, err := c.gen.AddAttachmentToPurchase(ctx, req, params)
	if err != nil {
		return Err[PurchasesAttachOut](MapErr(OpPurchasesAttach, err))
	}
	out := PurchasesAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[PurchasesAttachOut](out)
}

// PurchasesPaymentsListIn requires company + purchase id.
type PurchasesPaymentsListIn struct {
	Company    string `json:"company"`
	PurchaseID int64  `json:"purchase_id"`
}

// PurchasesPaymentsList returns all payments for a purchase. The
// upstream endpoint returns a bare array; Meta.Returned is the only
// field set.
func (c *Client) PurchasesPaymentsList(ctx context.Context, in PurchasesPaymentsListIn) Result[PaymentsListOut] {
	if in.Company == "" {
		return Err[PaymentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesPaymentsList,
		})
	}
	if in.PurchaseID == 0 {
		return Err[PaymentsListOut](&Error{
			Code: CodeValidation, Message: "purchase_id is required", Op: OpPurchasesPaymentsList,
		})
	}
	resp, err := c.gen.GetPurchasePayments(ctx, fiken.GetPurchasePaymentsParams{
		CompanySlug: in.Company,
		PurchaseId:  in.PurchaseID,
	})
	if err != nil {
		return Err[PaymentsListOut](MapErr(OpPurchasesPaymentsList, err))
	}
	items := make([]PaymentOut, 0, len(resp))
	for _, p := range resp {
		items = append(items, paymentToOut(p))
	}
	return Ok[PaymentsListOut](PaymentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// PurchasesPaymentsGetIn requires company + purchase id + payment id.
type PurchasesPaymentsGetIn struct {
	Company    string `json:"company"`
	PurchaseID int64  `json:"purchase_id"`
	PaymentID  int64  `json:"payment_id"`
}

// PurchasesPaymentsGet returns a single payment by id.
func (c *Client) PurchasesPaymentsGet(ctx context.Context, in PurchasesPaymentsGetIn) Result[PaymentOut] {
	if in.Company == "" {
		return Err[PaymentOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesPaymentsGet,
		})
	}
	if in.PurchaseID == 0 {
		return Err[PaymentOut](&Error{
			Code: CodeValidation, Message: "purchase_id is required", Op: OpPurchasesPaymentsGet,
		})
	}
	if in.PaymentID == 0 {
		return Err[PaymentOut](&Error{
			Code: CodeValidation, Message: "payment_id is required", Op: OpPurchasesPaymentsGet,
		})
	}
	resp, err := c.gen.GetPurchasePayment(ctx, fiken.GetPurchasePaymentParams{
		CompanySlug: in.Company,
		PurchaseId:  in.PurchaseID,
		PaymentId:   in.PaymentID,
	})
	if err != nil {
		return Err[PaymentOut](MapErr(OpPurchasesPaymentsGet, err))
	}
	if resp == nil {
		return Ok[PaymentOut](PaymentOut{})
	}
	return Ok[PaymentOut](paymentToOut(*resp))
}

// PurchasesPaymentsCreateIn carries the create-payment payload
// alongside the purchase identifier. Body mirrors the upstream Payment
// shape.
type PurchasesPaymentsCreateIn struct {
	Company    string         `json:"company"`
	PurchaseID int64          `json:"purchase_id"`
	Body       *fiken.Payment `json:"body"`
}

// PurchasesPaymentsCreateOut surfaces the Location header for the new
// payment.
type PurchasesPaymentsCreateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (PurchasesPaymentsCreateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o PurchasesPaymentsCreateOut) TableRow() []string { return []string{o.Location} }

// PurchasesPaymentsCreate registers a payment against a purchase.
// Upstream returns 201 with a Location header pointing at the new
// payment; surfaced via PurchasesPaymentsCreateOut.Location.
func (c *Client) PurchasesPaymentsCreate(ctx context.Context, in PurchasesPaymentsCreateIn) Result[PurchasesPaymentsCreateOut] {
	if in.Company == "" {
		return Err[PurchasesPaymentsCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesPaymentsCreate,
		})
	}
	if in.PurchaseID == 0 {
		return Err[PurchasesPaymentsCreateOut](&Error{
			Code: CodeValidation, Message: "purchase_id is required", Op: OpPurchasesPaymentsCreate,
		})
	}
	if in.Body == nil {
		return Err[PurchasesPaymentsCreateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpPurchasesPaymentsCreate,
		})
	}
	resp, err := c.gen.CreatePurchasePayment(ctx, in.Body, fiken.CreatePurchasePaymentParams{
		CompanySlug: in.Company,
		PurchaseId:  in.PurchaseID,
	})
	if err != nil {
		return Err[PurchasesPaymentsCreateOut](MapErr(OpPurchasesPaymentsCreate, err))
	}
	out := PurchasesPaymentsCreateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[PurchasesPaymentsCreateOut](out)
}
