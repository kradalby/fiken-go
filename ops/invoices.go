package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"

	"github.com/kradalby/fiken-go/fiken"
)

// InvoicesListIn carries paged-list input for invoices. The upstream
// endpoint exposes a richer filter surface than journal-entries —
// issue/due-date bounds, customer + reference filters — so the
// struct stays wider than its peers.
type InvoicesListIn struct {
	Company          string `json:"company"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"page_size,omitempty"`
	IssueDate        Date   `json:"issue_date,omitempty"`
	IssueDateLe      Date   `json:"issue_date_le,omitempty"`
	IssueDateLt      Date   `json:"issue_date_lt,omitempty"`
	IssueDateGe      Date   `json:"issue_date_ge,omitempty"`
	IssueDateGt      Date   `json:"issue_date_gt,omitempty"`
	LastModified     Date   `json:"last_modified,omitempty"`
	LastModifiedLe   Date   `json:"last_modified_le,omitempty"`
	LastModifiedLt   Date   `json:"last_modified_lt,omitempty"`
	LastModifiedGe   Date   `json:"last_modified_ge,omitempty"`
	LastModifiedGt   Date   `json:"last_modified_gt,omitempty"`
	DueDate          Date   `json:"due_date,omitempty"`
	DueDateLe        Date   `json:"due_date_le,omitempty"`
	DueDateLt        Date   `json:"due_date_lt,omitempty"`
	DueDateGe        Date   `json:"due_date_ge,omitempty"`
	DueDateGt        Date   `json:"due_date_gt,omitempty"`
	CustomerID       int64  `json:"customer_id,omitempty"`
	Settled          *bool  `json:"settled,omitempty"`
	OrderReference   string `json:"order_reference,omitempty"`
	InvoiceDraftUUID string `json:"invoice_draft_uuid,omitempty"`
	InvoiceNumber    string `json:"invoice_number,omitempty"`
}

// InvoiceLineOut is the canonical invoice-line shape. All monetary
// fields stay int64 øre per the global convention; vat_type surfaces
// the upstream enum verbatim ("HIGH", "LOW", ...). VatRate carries
// the percentage as basis-points (25% = 2500), translated from
// Fiken's 0–1 fractional vatInPercent.
type InvoiceLineOut struct {
	Net           int64   `json:"net,omitempty"`
	Vat           int64   `json:"vat,omitempty"`
	Gross         int64   `json:"gross,omitempty"`
	NetInNok      int64   `json:"net_in_nok,omitempty"`
	VatInNok      int64   `json:"vat_in_nok,omitempty"`
	GrossInNok    int64   `json:"gross_in_nok,omitempty"`
	VatType       string  `json:"vat_type,omitempty"`
	VatRate       int     `json:"vat_rate,omitempty"`
	UnitPrice     int64   `json:"unit_price,omitempty"`
	Quantity      float64 `json:"quantity,omitempty"`
	Discount      float64 `json:"discount,omitempty"`
	ProductID     int64   `json:"product_id,omitempty"`
	ProductName   string  `json:"product_name,omitempty"`
	Description   string  `json:"description,omitempty"`
	Comment       string  `json:"comment,omitempty"`
	IncomeAccount string  `json:"income_account,omitempty"`
}

// InvoiceOut is the canonical invoice shape exposed to CLI/MCP.
// Monetary totals stay int64 øre. Customer is flattened to id +
// name so the table renderer has a meaningful one-liner; the full
// Contact shape is reachable via JSON output.
type InvoiceOut struct {
	InvoiceID         int64            `json:"invoice_id,omitempty"`
	InvoiceNumber     int64            `json:"invoice_number,omitempty"`
	CreatedDate       Date             `json:"created_date,omitempty"`
	LastModifiedDate  Date             `json:"last_modified_date,omitempty"`
	IssueDate         Date             `json:"issue_date,omitempty"`
	DueDate           Date             `json:"due_date,omitempty"`
	OriginalDueDate   Date             `json:"original_due_date,omitempty"`
	Kid               string           `json:"kid,omitempty"`
	Net               int64            `json:"net,omitempty"`
	Vat               int64            `json:"vat,omitempty"`
	Gross             int64            `json:"gross,omitempty"`
	NetInNok          int64            `json:"net_in_nok,omitempty"`
	VatInNok          int64            `json:"vat_in_nok,omitempty"`
	GrossInNok        int64            `json:"gross_in_nok,omitempty"`
	Cash              bool             `json:"cash,omitempty"`
	InvoiceText       string           `json:"invoice_text,omitempty"`
	YourReference     string           `json:"your_reference,omitempty"`
	OurReference      string           `json:"our_reference,omitempty"`
	OrderReference    string           `json:"order_reference,omitempty"`
	InvoiceDraftUUID  string           `json:"invoice_draft_uuid,omitempty"`
	Currency          string           `json:"currency,omitempty"`
	BankAccountNumber string           `json:"bank_account_number,omitempty"`
	SentManually      bool             `json:"sent_manually,omitempty"`
	CustomerID        int64            `json:"customer_id,omitempty"`
	CustomerName      string           `json:"customer_name,omitempty"`
	Lines             []InvoiceLineOut `json:"lines,omitempty"`
	Attachments       []AttachmentOut  `json:"attachments,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoiceOut) TableHeader() []string {
	return []string{"ID", "NUMBER", "ISSUE", "DUE", "GROSS", "CUSTOMER"}
}

// TableRow implements output.tableRow.
func (i InvoiceOut) TableRow() []string {
	return []string{
		strconv.FormatInt(i.InvoiceID, 10),
		strconv.FormatInt(i.InvoiceNumber, 10),
		string(i.IssueDate),
		string(i.DueDate),
		strconv.FormatInt(i.Gross, 10),
		i.CustomerName,
	}
}

// InvoicesListOut is the paged response.
type InvoicesListOut = ListOut[InvoiceOut]

// InvoicesList returns invoices for the specified company.
func (c *Client) InvoicesList(ctx context.Context, in InvoicesListIn) Result[InvoicesListOut] {
	if in.Company == "" {
		return Err[InvoicesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesList,
		})
	}
	params := fiken.GetInvoicesParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	setDateParam(&params.IssueDate, in.IssueDate)
	setDateParam(&params.IssueDateLe, in.IssueDateLe)
	setDateParam(&params.IssueDateLt, in.IssueDateLt)
	setDateParam(&params.IssueDateGe, in.IssueDateGe)
	setDateParam(&params.IssueDateGt, in.IssueDateGt)
	setDateParam(&params.LastModified, in.LastModified)
	setDateParam(&params.LastModifiedLe, in.LastModifiedLe)
	setDateParam(&params.LastModifiedLt, in.LastModifiedLt)
	setDateParam(&params.LastModifiedGe, in.LastModifiedGe)
	setDateParam(&params.LastModifiedGt, in.LastModifiedGt)
	setDateParam(&params.DueDate, in.DueDate)
	setDateParam(&params.DueDateLe, in.DueDateLe)
	setDateParam(&params.DueDateLt, in.DueDateLt)
	setDateParam(&params.DueDateGe, in.DueDateGe)
	setDateParam(&params.DueDateGt, in.DueDateGt)
	if in.CustomerID != 0 {
		params.CustomerId.SetTo(in.CustomerID)
	}
	if in.Settled != nil {
		params.Settled.SetTo(*in.Settled)
	}
	if in.OrderReference != "" {
		params.OrderReference.SetTo(in.OrderReference)
	}
	if in.InvoiceDraftUUID != "" {
		if u, err := uuid.Parse(in.InvoiceDraftUUID); err == nil {
			params.InvoiceDraftUuid.SetTo(u)
		}
	}
	if in.InvoiceNumber != "" {
		params.InvoiceNumber.SetTo(in.InvoiceNumber)
	}
	resp, err := c.gen.GetInvoices(ctx, params)
	if err != nil {
		return Err[InvoicesListOut](MapErr(OpInvoicesList, err))
	}
	return Ok[InvoicesListOut](translateInvoicesList(resp))
}

// translateInvoicesList converts the ogen response into the canonical
// ListOut[InvoiceOut] envelope, including paging meta.
func translateInvoicesList(resp *fiken.GetInvoicesOKHeaders) InvoicesListOut {
	if resp == nil {
		return InvoicesListOut{Items: []InvoiceOut{}, Meta: ListMeta{}}
	}
	items := make([]InvoiceOut, 0, len(resp.Response))
	for _, i := range resp.Response {
		items = append(items, invoiceToOut(i))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return InvoicesListOut{Items: items, Meta: meta}
}

// invoiceToOut maps fiken.InvoiceResult into the canonical InvoiceOut.
// Customer is flattened to id + name; the full contact remains
// reachable through the underlying ogen Result for callers that
// peek at the upstream payload.
func invoiceToOut(i fiken.InvoiceResult) InvoiceOut {
	out := InvoiceOut{
		InvoiceID:         i.InvoiceId.Or(0),
		InvoiceNumber:     i.InvoiceNumber.Or(0),
		Kid:               i.Kid.Or(""),
		Net:               i.Net.Or(0),
		Vat:               i.Vat.Or(0),
		Gross:             i.Gross.Or(0),
		NetInNok:          i.NetInNok.Or(0),
		VatInNok:          i.VatInNok.Or(0),
		GrossInNok:        i.GrossInNok.Or(0),
		Cash:              i.Cash.Or(false),
		InvoiceText:       i.InvoiceText.Or(""),
		YourReference:     i.YourReference.Or(""),
		OurReference:      i.OurReference.Or(""),
		OrderReference:    i.OrderReference.Or(""),
		InvoiceDraftUUID:  i.InvoiceDraftUuid.Or(""),
		Currency:          i.Currency.Or(""),
		BankAccountNumber: i.BankAccountNumber.Or(""),
		SentManually:      i.SentManually.Or(false),
	}
	if d, ok := i.CreatedDate.Get(); ok {
		out.CreatedDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := i.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := i.IssueDate.Get(); ok {
		out.IssueDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := i.DueDate.Get(); ok {
		out.DueDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := i.OriginalDueDate.Get(); ok {
		out.OriginalDueDate = Date(d.Format("2006-01-02"))
	}
	if cust, ok := i.Customer.Get(); ok {
		out.CustomerID = cust.ContactId.Or(0)
		out.CustomerName = cust.Name
	}
	if len(i.Lines) > 0 {
		out.Lines = make([]InvoiceLineOut, 0, len(i.Lines))
		for _, ln := range i.Lines {
			out.Lines = append(out.Lines, invoiceLineToOut(ln))
		}
	}
	if len(i.Attachments) > 0 {
		out.Attachments = make([]AttachmentOut, 0, len(i.Attachments))
		for _, a := range i.Attachments {
			out.Attachments = append(out.Attachments, attachmentToOut(a))
		}
	}
	return out
}

// invoiceLineToOut maps fiken.InvoiceLineResult into the canonical
// InvoiceLineOut. VatInPercent (0–1 fraction) is rescaled to
// basis-points (0–10000) to satisfy the global int rate convention.
func invoiceLineToOut(ln fiken.InvoiceLineResult) InvoiceLineOut {
	return InvoiceLineOut{
		Net:           ln.Net.Or(0),
		Vat:           ln.Vat.Or(0),
		Gross:         ln.Gross.Or(0),
		NetInNok:      ln.NetInNok.Or(0),
		VatInNok:      ln.VatInNok.Or(0),
		GrossInNok:    ln.GrossInNok.Or(0),
		VatType:       canonicalVatType(ln.VatType.Or("")),
		VatRate:       int(ln.VatInPercent.Or(0) * 10000),
		UnitPrice:     ln.UnitPrice.Or(0),
		Quantity:      ln.Quantity.Or(0),
		Discount:      ln.Discount.Or(0),
		ProductID:     ln.ProductId.Or(0),
		ProductName:   ln.ProductName.Or(""),
		Description:   ln.Description.Or(""),
		Comment:       ln.Comment.Or(""),
		IncomeAccount: ln.IncomeAccount.Or(""),
	}
}

// InvoicesGetIn requires company + invoice id.
type InvoicesGetIn struct {
	Company   string `json:"company"`
	InvoiceID int64  `json:"invoice_id"`
}

// InvoicesGet returns a single invoice by id.
func (c *Client) InvoicesGet(ctx context.Context, in InvoicesGetIn) Result[InvoiceOut] {
	if in.Company == "" {
		return Err[InvoiceOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesGet,
		})
	}
	if in.InvoiceID == 0 {
		return Err[InvoiceOut](&Error{
			Code: CodeValidation, Message: "invoice_id is required", Op: OpInvoicesGet,
		})
	}
	resp, err := c.gen.GetInvoice(ctx, fiken.GetInvoiceParams{
		CompanySlug: in.Company,
		InvoiceId:   in.InvoiceID,
	})
	if err != nil {
		return Err[InvoiceOut](MapErr(OpInvoicesGet, err))
	}
	if resp == nil {
		return Ok[InvoiceOut](InvoiceOut{})
	}
	return Ok[InvoiceOut](invoiceToOut(*resp))
}

// InvoicesSendIn carries the send-invoice payload. Body mirrors the
// upstream SendInvoiceRequest so the field surface stays in lock-step
// with the spec.
type InvoicesSendIn struct {
	Company string                    `json:"company"`
	Body    *fiken.SendInvoiceRequest `json:"body"`
}

// InvoicesSendOut is the canonical success shape for sendInvoice.
// The upstream endpoint returns 200 with no body; surfacing the
// invoice_id keeps the success/error envelope meaningful.
type InvoicesSendOut struct {
	InvoiceID int64 `json:"invoice_id,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoicesSendOut) TableHeader() []string { return []string{"INVOICE_ID"} }

// TableRow implements output.tableRow.
func (o InvoicesSendOut) TableRow() []string {
	return []string{strconv.FormatInt(o.InvoiceID, 10)}
}

// InvoicesSend dispatches a finalized invoice through Fiken (email,
// EHF, etc.) using the request body to choose the delivery method and
// recipients. Upstream returns 200 with no body; we surface the
// invoice id echoed from the request for the table renderer.
func (c *Client) InvoicesSend(ctx context.Context, in InvoicesSendIn) Result[InvoicesSendOut] {
	if in.Company == "" {
		return Err[InvoicesSendOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesSend,
		})
	}
	if in.Body == nil {
		return Err[InvoicesSendOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpInvoicesSend,
		})
	}
	if err := c.gen.SendInvoice(ctx, in.Body, fiken.SendInvoiceParams{
		CompanySlug: in.Company,
	}); err != nil {
		return Err[InvoicesSendOut](MapErr(OpInvoicesSend, err))
	}
	return Ok[InvoicesSendOut](InvoicesSendOut{InvoiceID: in.Body.InvoiceId})
}

// InvoicesCounterCreateIn carries the create-counter payload. The
// upstream endpoint sets the invoice-counter starting value for the
// fiscal year.
type InvoicesCounterCreateIn struct {
	Company string `json:"company"`
	Value   int32  `json:"value"`
}

// InvoicesCounterCreateOut surfaces the new counter value.
type InvoicesCounterCreateOut struct {
	Value int32 `json:"value,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoicesCounterCreateOut) TableHeader() []string { return []string{"VALUE"} }

// TableRow implements output.tableRow.
func (o InvoicesCounterCreateOut) TableRow() []string {
	return []string{strconv.Itoa(int(o.Value))}
}

// InvoicesCounterCreate sets the invoice-number counter for the
// current fiscal year. Upstream returns 201 with no body; we echo the
// requested value back as success confirmation.
func (c *Client) InvoicesCounterCreate(ctx context.Context, in InvoicesCounterCreateIn) Result[InvoicesCounterCreateOut] {
	if in.Company == "" {
		return Err[InvoicesCounterCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesCounterCreate,
		})
	}
	if in.Value <= 0 {
		return Err[InvoicesCounterCreateOut](&Error{
			Code: CodeValidation, Message: "value must be > 0", Op: OpInvoicesCounterCreate,
		})
	}
	counter := fiken.NewOptCounter(fiken.Counter{Value: fiken.NewOptInt32(in.Value)})
	if err := c.gen.CreateInvoiceCounter(ctx, counter, fiken.CreateInvoiceCounterParams{
		CompanySlug: in.Company,
	}); err != nil {
		return Err[InvoicesCounterCreateOut](MapErr(OpInvoicesCounterCreate, err))
	}
	return Ok[InvoicesCounterCreateOut](InvoicesCounterCreateOut{Value: in.Value})
}

// InvoicesAttachmentsListIn requires company + invoice id.
type InvoicesAttachmentsListIn struct {
	Company   string `json:"company"`
	InvoiceID int64  `json:"invoice_id"`
}

// InvoicesAttachmentsList returns all attachments for an invoice. The
// upstream endpoint returns a bare array; Meta.Returned is the only
// field that gets set. Reuses AttachmentOut from journal_entries.go.
func (c *Client) InvoicesAttachmentsList(ctx context.Context, in InvoicesAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesAttachmentsList,
		})
	}
	if in.InvoiceID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "invoice_id is required", Op: OpInvoicesAttachmentsList,
		})
	}
	resp, err := c.gen.GetInvoiceAttachments(ctx, fiken.GetInvoiceAttachmentsParams{
		CompanySlug: in.Company,
		InvoiceId:   in.InvoiceID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpInvoicesAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// InvoicesAttachmentsAttachIn carries the multipart upload payload.
// Filename overrides the form-field filename; FilePath is the local
// path read by ops.OpenMultipartFile.
type InvoicesAttachmentsAttachIn struct {
	Company   string `json:"company"`
	InvoiceID int64  `json:"invoice_id"`
	Filename  string `json:"filename"`
	FilePath  string `json:"file_path"`
}

// InvoicesAttachmentsAttachOut mirrors the upstream Created response.
type InvoicesAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoicesAttachmentsAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a InvoicesAttachmentsAttachOut) TableRow() []string { return []string{a.Location} }

// InvoicesCreateIn carries the create-invoice payload. Body mirrors
// the upstream InvoiceRequest so the field surface stays in lock-step
// with the spec.
type InvoicesCreateIn struct {
	Company string                `json:"company"`
	Body    *fiken.InvoiceRequest `json:"body"`
}

// InvoicesCreateOut surfaces the Location header pointing at the
// newly-created invoice resource.
type InvoicesCreateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoicesCreateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o InvoicesCreateOut) TableRow() []string { return []string{o.Location} }

// InvoicesCreate posts a finalized invoice directly (no draft step).
// Upstream returns 201 with the new invoice URL in Location.
func (c *Client) InvoicesCreate(ctx context.Context, in InvoicesCreateIn) Result[InvoicesCreateOut] {
	if in.Company == "" {
		return Err[InvoicesCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesCreate,
		})
	}
	if in.Body == nil {
		return Err[InvoicesCreateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpInvoicesCreate,
		})
	}
	resp, err := c.gen.CreateInvoice(ctx, in.Body, fiken.CreateInvoiceParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[InvoicesCreateOut](MapErr(OpInvoicesCreate, err))
	}
	out := InvoicesCreateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoicesCreateOut](out)
}

// InvoicesUpdateIn carries the update-invoice payload. Body mirrors
// the upstream UpdateInvoiceRequest.
type InvoicesUpdateIn struct {
	Company   string                      `json:"company"`
	InvoiceID int64                       `json:"invoice_id"`
	Body      *fiken.UpdateInvoiceRequest `json:"body"`
}

// InvoicesUpdateOut surfaces the Location header pointing back at
// the updated invoice resource.
type InvoicesUpdateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoicesUpdateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o InvoicesUpdateOut) TableRow() []string { return []string{o.Location} }

// InvoicesUpdate patches mutable fields on a posted invoice (KID,
// payment account, references, etc). Upstream returns 200 with the
// invoice URL in Location.
func (c *Client) InvoicesUpdate(ctx context.Context, in InvoicesUpdateIn) Result[InvoicesUpdateOut] {
	if in.Company == "" {
		return Err[InvoicesUpdateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesUpdate,
		})
	}
	if in.InvoiceID == 0 {
		return Err[InvoicesUpdateOut](&Error{
			Code: CodeValidation, Message: "invoice_id is required", Op: OpInvoicesUpdate,
		})
	}
	if in.Body == nil {
		return Err[InvoicesUpdateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpInvoicesUpdate,
		})
	}
	resp, err := c.gen.UpdateInvoice(ctx, in.Body, fiken.UpdateInvoiceParams{
		CompanySlug: in.Company,
		InvoiceId:   in.InvoiceID,
	})
	if err != nil {
		return Err[InvoicesUpdateOut](MapErr(OpInvoicesUpdate, err))
	}
	out := InvoicesUpdateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoicesUpdateOut](out)
}

// CounterGetIn carries the company slug to look up a counter. Shared
// by invoice / offer / order_confirmation / credit_note GETs.
type CounterGetIn struct {
	Company string `json:"company"`
}

// CounterGetOut surfaces the current counter value. Reused by all
// four counter-get ops; the field is int32 to mirror the upstream
// schema.
type CounterGetOut struct {
	Value int32 `json:"value"`
}

// TableHeader implements output.tableRow.
func (CounterGetOut) TableHeader() []string { return []string{"VALUE"} }

// TableRow implements output.tableRow.
func (o CounterGetOut) TableRow() []string {
	return []string{strconv.Itoa(int(o.Value))}
}

// InvoicesCounterGet returns the current invoice-number counter
// value for the fiscal year. Read-only.
func (c *Client) InvoicesCounterGet(ctx context.Context, in CounterGetIn) Result[CounterGetOut] {
	if in.Company == "" {
		return Err[CounterGetOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesCounterGet,
		})
	}
	resp, err := c.gen.GetInvoiceCounter(ctx, fiken.GetInvoiceCounterParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[CounterGetOut](MapErr(OpInvoicesCounterGet, err))
	}
	if resp == nil {
		return Ok[CounterGetOut](CounterGetOut{})
	}
	return Ok[CounterGetOut](CounterGetOut{Value: resp.Value.Or(0)})
}

// InvoicesAttachmentsAttach uploads a file to an invoice as multipart
// form data. The form-field filename defaults to the basename of
// FilePath; pass Filename to override.
func (c *Client) InvoicesAttachmentsAttach(ctx context.Context, in InvoicesAttachmentsAttachIn) Result[InvoicesAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[InvoicesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesAttachmentsAttach,
		})
	}
	if in.InvoiceID == 0 {
		return Err[InvoicesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "invoice_id is required", Op: OpInvoicesAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[InvoicesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpInvoicesAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[InvoicesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpInvoicesAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToInvoiceReq(fiken.AddAttachmentToInvoiceReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToInvoice(ctx, req, fiken.AddAttachmentToInvoiceParams{
		CompanySlug: in.Company,
		InvoiceId:   in.InvoiceID,
	})
	if err != nil {
		return Err[InvoicesAttachmentsAttachOut](MapErr(OpInvoicesAttachmentsAttach, err))
	}
	out := InvoicesAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoicesAttachmentsAttachOut](out)
}
