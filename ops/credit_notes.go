package ops

import (
	"context"
	"net/url"
	"strconv"

	"github.com/google/uuid"

	"github.com/kradalby/fiken-go/fiken"
)

// CreditNotesListIn carries paged-list input for credit notes. The
// filter surface mirrors invoices on the issue-date / last-modified /
// customer / settled axes; due-date is not exposed upstream so it is
// absent here.
type CreditNotesListIn struct {
	Company             string `json:"company"`
	Page                int    `json:"page,omitempty"`
	PageSize            int    `json:"page_size,omitempty"`
	IssueDate           Date   `json:"issue_date,omitempty"`
	IssueDateLe         Date   `json:"issue_date_le,omitempty"`
	IssueDateLt         Date   `json:"issue_date_lt,omitempty"`
	IssueDateGe         Date   `json:"issue_date_ge,omitempty"`
	IssueDateGt         Date   `json:"issue_date_gt,omitempty"`
	LastModified        Date   `json:"last_modified,omitempty"`
	LastModifiedLe      Date   `json:"last_modified_le,omitempty"`
	LastModifiedLt      Date   `json:"last_modified_lt,omitempty"`
	LastModifiedGe      Date   `json:"last_modified_ge,omitempty"`
	LastModifiedGt      Date   `json:"last_modified_gt,omitempty"`
	CustomerID          int64  `json:"customer_id,omitempty"`
	Settled             *bool  `json:"settled,omitempty"`
	CreditNoteDraftUUID string `json:"credit_note_draft_uuid,omitempty"`
}

// CreditNoteOut is the canonical credit-note shape exposed to CLI/MCP.
// Monetary fields stay int64 øre; customer is flattened to id + name
// so the table renderer has a meaningful one-liner. Reuses
// InvoiceLineOut for the lines slice since the upstream schema points
// at invoiceLineResult.
type CreditNoteOut struct {
	CreditNoteID        int64            `json:"credit_note_id,omitempty"`
	CreditNoteNumber    int64            `json:"credit_note_number,omitempty"`
	Kid                 string           `json:"kid,omitempty"`
	Net                 int64            `json:"net,omitempty"`
	Vat                 int64            `json:"vat,omitempty"`
	Gross               int64            `json:"gross,omitempty"`
	NetInNok            int64            `json:"net_in_nok,omitempty"`
	VatInNok            int64            `json:"vat_in_nok,omitempty"`
	GrossInNok          int64            `json:"gross_in_nok,omitempty"`
	CreditNoteText      string           `json:"credit_note_text,omitempty"`
	YourReference       string           `json:"your_reference,omitempty"`
	OurReference        string           `json:"our_reference,omitempty"`
	OrderReference      string           `json:"order_reference,omitempty"`
	Currency            string           `json:"currency,omitempty"`
	IssueDate           Date             `json:"issue_date,omitempty"`
	LastModifiedDate    Date             `json:"last_modified_date,omitempty"`
	Settled             bool             `json:"settled,omitempty"`
	AssociatedInvoiceID int64            `json:"associated_invoice_id,omitempty"`
	CreditNoteDraftUUID string           `json:"credit_note_draft_uuid,omitempty"`
	CustomerID          int64            `json:"customer_id,omitempty"`
	CustomerName        string           `json:"customer_name,omitempty"`
	Lines               []InvoiceLineOut `json:"lines,omitempty"`
	CreditNotePdf       *AttachmentOut   `json:"credit_note_pdf,omitempty"`
}

// TableHeader implements output.tableRow.
func (CreditNoteOut) TableHeader() []string {
	return []string{"ID", "NUMBER", "ISSUE", "GROSS", "CUSTOMER"}
}

// TableRow implements output.tableRow.
func (c CreditNoteOut) TableRow() []string {
	return []string{
		strconv.FormatInt(c.CreditNoteID, 10),
		strconv.FormatInt(c.CreditNoteNumber, 10),
		string(c.IssueDate),
		strconv.FormatInt(c.Gross, 10),
		c.CustomerName,
	}
}

// CreditNotesListOut is the paged response.
type CreditNotesListOut = ListOut[CreditNoteOut]

// CreditNotesList returns credit notes for the specified company.
func (c *Client) CreditNotesList(ctx context.Context, in CreditNotesListIn) Result[CreditNotesListOut] {
	if in.Company == "" {
		return Err[CreditNotesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesList,
		})
	}
	params := fiken.GetCreditNotesParams{CompanySlug: in.Company}
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
	if in.CustomerID != 0 {
		params.CustomerId.SetTo(in.CustomerID)
	}
	if in.Settled != nil {
		params.Settled.SetTo(*in.Settled)
	}
	if in.CreditNoteDraftUUID != "" {
		if u, err := uuid.Parse(in.CreditNoteDraftUUID); err == nil {
			params.CreditNoteDraftUuid.SetTo(u)
		}
	}
	resp, err := c.gen.GetCreditNotes(ctx, params)
	if err != nil {
		return Err[CreditNotesListOut](MapErr(OpCreditNotesList, err))
	}
	return Ok[CreditNotesListOut](translateCreditNotesList(resp))
}

// translateCreditNotesList converts the ogen response into the
// canonical ListOut[CreditNoteOut] envelope, including paging meta.
func translateCreditNotesList(resp *fiken.GetCreditNotesOKHeaders) CreditNotesListOut {
	if resp == nil {
		return CreditNotesListOut{Items: []CreditNoteOut{}, Meta: ListMeta{}}
	}
	items := make([]CreditNoteOut, 0, len(resp.Response))
	for _, cn := range resp.Response {
		items = append(items, creditNoteToOut(cn))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return CreditNotesListOut{Items: items, Meta: meta}
}

// creditNoteToOut maps fiken.CreditNoteResult into the canonical
// CreditNoteOut. Customer is flattened to id + name; CreditNotePdf
// reuses AttachmentOut. Lines reuse InvoiceLineOut because the
// upstream type is invoiceLineResult.
func creditNoteToOut(cn fiken.CreditNoteResult) CreditNoteOut {
	out := CreditNoteOut{
		CreditNoteID:        cn.CreditNoteId,
		CreditNoteNumber:    cn.CreditNoteNumber,
		Kid:                 cn.Kid.Or(""),
		Net:                 cn.Net,
		Vat:                 cn.Vat,
		Gross:               cn.Gross,
		NetInNok:            cn.NetInNok,
		VatInNok:            cn.VatInNok,
		GrossInNok:          cn.GrossInNok,
		CreditNoteText:      cn.CreditNoteText.Or(""),
		YourReference:       cn.YourReference.Or(""),
		OurReference:        cn.OurReference.Or(""),
		OrderReference:      cn.OrderReference.Or(""),
		Currency:            cn.Currency.Or(""),
		Settled:             cn.Settled.Or(false),
		AssociatedInvoiceID: cn.AssociatedInvoiceId.Or(0),
		CreditNoteDraftUUID: cn.CreditNoteDraftUuid.Or(""),
	}
	if d, ok := cn.IssueDate.Get(); ok {
		out.IssueDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := cn.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(d.Format("2006-01-02"))
	}
	out.CustomerID = cn.Customer.ContactId.Or(0)
	out.CustomerName = cn.Customer.Name
	if len(cn.Lines) > 0 {
		out.Lines = make([]InvoiceLineOut, 0, len(cn.Lines))
		for _, ln := range cn.Lines {
			out.Lines = append(out.Lines, invoiceLineToOut(ln))
		}
	}
	if pdf, ok := cn.CreditNotePdf.Get(); ok {
		a := attachmentToOut(pdf)
		out.CreditNotePdf = &a
	}
	return out
}

// CreditNotesGetIn requires company + credit-note id. CreditNoteID is
// a string here because the upstream path-param is `string` (not int)
// in the spec — keeps the call-site faithful to the OAS shape.
type CreditNotesGetIn struct {
	Company      string `json:"company"`
	CreditNoteID string `json:"credit_note_id"`
}

// CreditNotesGet returns a single credit note by id.
func (c *Client) CreditNotesGet(ctx context.Context, in CreditNotesGetIn) Result[CreditNoteOut] {
	if in.Company == "" {
		return Err[CreditNoteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesGet,
		})
	}
	if in.CreditNoteID == "" {
		return Err[CreditNoteOut](&Error{
			Code: CodeValidation, Message: "credit_note_id is required", Op: OpCreditNotesGet,
		})
	}
	resp, err := c.gen.GetCreditNote(ctx, fiken.GetCreditNoteParams{
		CompanySlug:  in.Company,
		CreditNoteId: in.CreditNoteID,
	})
	if err != nil {
		return Err[CreditNoteOut](MapErr(OpCreditNotesGet, err))
	}
	if resp == nil {
		return Ok[CreditNoteOut](CreditNoteOut{})
	}
	return Ok[CreditNoteOut](creditNoteToOut(*resp))
}

// CreditNotesSendIn carries the send-credit-note payload. Body mirrors
// the upstream SendCreditNoteRequest so the field surface stays in
// lock-step with the spec.
type CreditNotesSendIn struct {
	Company string                       `json:"company"`
	Body    *fiken.SendCreditNoteRequest `json:"body"`
}

// CreditNotesSendOut is the canonical success shape for sendCreditNote.
// The upstream endpoint returns 200 with no body; surfacing the
// credit_note_id keeps the success/error envelope meaningful.
type CreditNotesSendOut struct {
	CreditNoteID int64 `json:"credit_note_id,omitempty"`
}

// TableHeader implements output.tableRow.
func (CreditNotesSendOut) TableHeader() []string { return []string{"CREDIT_NOTE_ID"} }

// TableRow implements output.tableRow.
func (o CreditNotesSendOut) TableRow() []string {
	return []string{strconv.FormatInt(o.CreditNoteID, 10)}
}

// CreditNotesSend posts a send-credit-note request. Upstream returns
// 200 with no body; we echo the body's CreditNoteId back so the
// success envelope stays meaningful.
func (c *Client) CreditNotesSend(ctx context.Context, in CreditNotesSendIn) Result[CreditNotesSendOut] {
	if in.Company == "" {
		return Err[CreditNotesSendOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesSend,
		})
	}
	if in.Body == nil {
		return Err[CreditNotesSendOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpCreditNotesSend,
		})
	}
	if err := c.gen.SendCreditNote(ctx, in.Body, fiken.SendCreditNoteParams{
		CompanySlug: in.Company,
	}); err != nil {
		return Err[CreditNotesSendOut](MapErr(OpCreditNotesSend, err))
	}
	return Ok[CreditNotesSendOut](CreditNotesSendOut{CreditNoteID: in.Body.CreditNoteId})
}

// CreditNotesCounterCreateIn carries the create-counter payload. The
// upstream endpoint sets the credit-note-counter starting value for
// the fiscal year.
type CreditNotesCounterCreateIn struct {
	Company string `json:"company"`
	Value   int32  `json:"value"`
}

// CreditNotesCounterCreateOut surfaces the new counter value.
type CreditNotesCounterCreateOut struct {
	Value int32 `json:"value,omitempty"`
}

// TableHeader implements output.tableRow.
func (CreditNotesCounterCreateOut) TableHeader() []string { return []string{"VALUE"} }

// TableRow implements output.tableRow.
func (o CreditNotesCounterCreateOut) TableRow() []string {
	return []string{strconv.Itoa(int(o.Value))}
}

// CreditNotesCounterCreate sets the credit-note-counter starting
// value for the current fiscal year. Upstream returns 201 with no
// body; we echo the requested value back as confirmation.
func (c *Client) CreditNotesCounterCreate(ctx context.Context, in CreditNotesCounterCreateIn) Result[CreditNotesCounterCreateOut] {
	if in.Company == "" {
		return Err[CreditNotesCounterCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesCounterCreate,
		})
	}
	if in.Value <= 0 {
		return Err[CreditNotesCounterCreateOut](&Error{
			Code: CodeValidation, Message: "value must be > 0", Op: OpCreditNotesCounterCreate,
		})
	}
	counter := fiken.NewOptCounter(fiken.Counter{Value: fiken.NewOptInt32(in.Value)})
	if err := c.gen.CreateCreditNoteCounter(ctx, counter, fiken.CreateCreditNoteCounterParams{
		CompanySlug: in.Company,
	}); err != nil {
		return Err[CreditNotesCounterCreateOut](MapErr(OpCreditNotesCounterCreate, err))
	}
	return Ok[CreditNotesCounterCreateOut](CreditNotesCounterCreateOut{Value: in.Value})
}

// CreditNotesFullCreateIn turns an invoice into a full credit note via
// POST /creditNotes/full. Body mirrors the upstream FullCreditNoteRequest.
type CreditNotesFullCreateIn struct {
	Company string                       `json:"company"`
	Body    *fiken.FullCreditNoteRequest `json:"body"`
}

// CreditNotesFullCreateOut surfaces the Location header pointing at
// the newly-created credit note resource.
type CreditNotesFullCreateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (CreditNotesFullCreateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o CreditNotesFullCreateOut) TableRow() []string { return []string{o.Location} }

// CreditNotesFullCreate posts a full credit-note request against an
// existing invoice. Upstream returns 201 with the new credit note URL
// in the Location header.
func (c *Client) CreditNotesFullCreate(ctx context.Context, in CreditNotesFullCreateIn) Result[CreditNotesFullCreateOut] {
	if in.Company == "" {
		return Err[CreditNotesFullCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesFullCreate,
		})
	}
	if in.Body == nil {
		return Err[CreditNotesFullCreateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpCreditNotesFullCreate,
		})
	}
	resp, err := c.gen.CreateFullCreditNote(ctx, in.Body, fiken.CreateFullCreditNoteParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[CreditNotesFullCreateOut](MapErr(OpCreditNotesFullCreate, err))
	}
	out := CreditNotesFullCreateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[CreditNotesFullCreateOut](out)
}

// CreditNotesPartialCreateIn turns an invoice into a partial credit
// note via POST /creditNotes/partial.
type CreditNotesPartialCreateIn struct {
	Company string                          `json:"company"`
	Body    *fiken.PartialCreditNoteRequest `json:"body"`
}

// CreditNotesPartialCreateOut surfaces the Location header pointing
// at the newly-created credit note resource.
type CreditNotesPartialCreateOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (CreditNotesPartialCreateOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o CreditNotesPartialCreateOut) TableRow() []string { return []string{o.Location} }

// CreditNotesPartialCreate posts a partial credit-note request. The
// upstream endpoint returns 201 with the new credit note URL in
// Location.
func (c *Client) CreditNotesPartialCreate(ctx context.Context, in CreditNotesPartialCreateIn) Result[CreditNotesPartialCreateOut] {
	if in.Company == "" {
		return Err[CreditNotesPartialCreateOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesPartialCreate,
		})
	}
	if in.Body == nil {
		return Err[CreditNotesPartialCreateOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpCreditNotesPartialCreate,
		})
	}
	resp, err := c.gen.CreatePartialCreditNote(ctx, in.Body, fiken.CreatePartialCreditNoteParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[CreditNotesPartialCreateOut](MapErr(OpCreditNotesPartialCreate, err))
	}
	out := CreditNotesPartialCreateOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[CreditNotesPartialCreateOut](out)
}

// CreditNotesCounterGet returns the current credit-note counter
// value for the fiscal year. Read-only. Reuses CounterGetIn /
// CounterGetOut from invoices.go.
func (c *Client) CreditNotesCounterGet(ctx context.Context, in CounterGetIn) Result[CounterGetOut] {
	if in.Company == "" {
		return Err[CounterGetOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesCounterGet,
		})
	}
	resp, err := c.gen.GetCreditNoteCounter(ctx, fiken.GetCreditNoteCounterParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[CounterGetOut](MapErr(OpCreditNotesCounterGet, err))
	}
	if resp == nil {
		return Ok[CounterGetOut](CounterGetOut{})
	}
	return Ok[CounterGetOut](CounterGetOut{Value: resp.Value.Or(0)})
}
