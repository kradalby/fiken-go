package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"

	"github.com/kradalby/fiken-go/fiken"
)

// InvoiceDraftsListIn carries paged-list input for invoice drafts.
// The upstream endpoint exposes only a narrow filter surface
// (order_reference + uuid), so this struct stays slim.
type InvoiceDraftsListIn struct {
	Company        string `json:"company"`
	Page           int    `json:"page,omitempty"`
	PageSize       int    `json:"page_size,omitempty"`
	OrderReference string `json:"order_reference,omitempty"`
	UUID           string `json:"uuid,omitempty"`
}

// InvoiceDraftLineOut is the canonical invoice-draft-line shape.
// Quantity stays float64 since drafts allow fractional units.
type InvoiceDraftLineOut struct {
	InvoiceishDraftLineID int64   `json:"invoiceish_draft_line_id,omitempty"`
	LastModifiedDate      Date    `json:"last_modified_date,omitempty"`
	ProductID             int64   `json:"product_id,omitempty"`
	Description           string  `json:"description,omitempty"`
	UnitPrice             int64   `json:"unit_price,omitempty"`
	VatType               string  `json:"vat_type,omitempty"`
	Quantity              float64 `json:"quantity,omitempty"`
	Discount              float64 `json:"discount,omitempty"`
	Comment               string  `json:"comment,omitempty"`
	IncomeAccount         string  `json:"income_account,omitempty"`
}

// InvoiceDraftOut is the canonical invoice-draft shape exposed to
// CLI/MCP. Monetary fields stay int64 øre. Type surfaces the upstream
// "invoice" / "cash_invoice" / "credit_note" / "offer" enum verbatim.
// Location surfaces the upstream Location header for Create / Update
// responses (empty for List / Get rows).
type InvoiceDraftOut struct {
	DraftID              int64                 `json:"draft_id,omitempty"`
	UUID                 string                `json:"uuid,omitempty"`
	Type                 string                `json:"type,omitempty"`
	LastModifiedDate     Date                  `json:"last_modified_date,omitempty"`
	IssueDate            Date                  `json:"issue_date,omitempty"`
	DueDays              int32                 `json:"due_days,omitempty"`
	InvoiceText          string                `json:"invoice_text,omitempty"`
	Currency             string                `json:"currency,omitempty"`
	YourReference        string                `json:"your_reference,omitempty"`
	OurReference         string                `json:"our_reference,omitempty"`
	OrderReference       string                `json:"order_reference,omitempty"`
	Net                  int64                 `json:"net,omitempty"`
	Gross                int64                 `json:"gross,omitempty"`
	BankAccountNumber    string                `json:"bank_account_number,omitempty"`
	Iban                 string                `json:"iban,omitempty"`
	Bic                  string                `json:"bic,omitempty"`
	PaymentAccount       string                `json:"payment_account,omitempty"`
	CreatedFromInvoiceID int64                 `json:"created_from_invoice_id,omitempty"`
	ProjectID            int64                 `json:"project_id,omitempty"`
	Lines                []InvoiceDraftLineOut `json:"lines,omitempty"`
	Attachments          []AttachmentOut       `json:"attachments,omitempty"`
	CustomerIDs          []int64               `json:"customer_ids,omitempty"`
	Location             string                `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoiceDraftOut) TableHeader() []string {
	return []string{"ID", "UUID", "TYPE", "ISSUE", "GROSS"}
}

// TableRow implements output.tableRow.
func (d InvoiceDraftOut) TableRow() []string {
	return []string{
		strconv.FormatInt(d.DraftID, 10),
		d.UUID,
		d.Type,
		string(d.IssueDate),
		strconv.FormatInt(d.Gross, 10),
	}
}

// InvoiceDraftsListOut is the paged response.
type InvoiceDraftsListOut = ListOut[InvoiceDraftOut]

// InvoiceDraftsList returns invoice drafts for the specified company.
func (c *Client) InvoiceDraftsList(ctx context.Context, in InvoiceDraftsListIn) Result[InvoiceDraftsListOut] {
	if in.Company == "" {
		return Err[InvoiceDraftsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsList,
		})
	}
	params := fiken.GetInvoiceDraftsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.OrderReference != "" {
		params.OrderReference.SetTo(in.OrderReference)
	}
	if in.UUID != "" {
		if u, err := uuid.Parse(in.UUID); err == nil {
			params.UUID.SetTo(u)
		}
	}
	resp, err := c.gen.GetInvoiceDrafts(ctx, params)
	if err != nil {
		return Err[InvoiceDraftsListOut](MapErr(OpInvoicesDraftsList, err))
	}
	return Ok[InvoiceDraftsListOut](translateInvoiceDraftsList(resp))
}

// translateInvoiceDraftsList converts the ogen response into the
// canonical ListOut[InvoiceDraftOut] envelope.
func translateInvoiceDraftsList(resp *fiken.GetInvoiceDraftsOKHeaders) InvoiceDraftsListOut {
	if resp == nil {
		return InvoiceDraftsListOut{Items: []InvoiceDraftOut{}, Meta: ListMeta{}}
	}
	items := make([]InvoiceDraftOut, 0, len(resp.Response))
	for _, d := range resp.Response {
		items = append(items, invoiceDraftToOut(d))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return InvoiceDraftsListOut{Items: items, Meta: meta}
}

// invoiceDraftToOut maps fiken.InvoiceishDraftResult into the canonical
// InvoiceDraftOut. Type is rendered as its enum string ("invoice",
// "cash_invoice", ...). CustomerIDs flattens the upstream customers[]
// to contact ids — the full Contact remains reachable via the
// underlying ogen Result.
func invoiceDraftToOut(d fiken.InvoiceishDraftResult) InvoiceDraftOut {
	out := InvoiceDraftOut{
		DraftID:              d.DraftId.Or(0),
		UUID:                 d.UUID.Or(""),
		DueDays:              d.DaysUntilDueDate.Or(0),
		InvoiceText:          d.InvoiceText.Or(""),
		Currency:             d.Currency.Or(""),
		YourReference:        d.YourReference.Or(""),
		OurReference:         d.OurReference.Or(""),
		OrderReference:       d.OrderReference.Or(""),
		Net:                  d.Net.Or(0),
		Gross:                d.Gross.Or(0),
		BankAccountNumber:    d.BankAccountNumber.Or(""),
		Iban:                 d.Iban.Or(""),
		Bic:                  d.Bic.Or(""),
		PaymentAccount:       d.PaymentAccount.Or(""),
		CreatedFromInvoiceID: d.CreatedFromInvoiceId.Or(0),
		ProjectID:            d.ProjectId.Or(0),
	}
	if t, ok := d.Type.Get(); ok {
		out.Type = string(t)
	}
	if dt, ok := d.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(dt.Format("2006-01-02"))
	}
	if dt, ok := d.IssueDate.Get(); ok {
		out.IssueDate = Date(dt.Format("2006-01-02"))
	}
	if len(d.Lines) > 0 {
		out.Lines = make([]InvoiceDraftLineOut, 0, len(d.Lines))
		for _, ln := range d.Lines {
			out.Lines = append(out.Lines, invoiceDraftLineToOut(ln))
		}
	}
	if len(d.Attachments) > 0 {
		out.Attachments = make([]AttachmentOut, 0, len(d.Attachments))
		for _, a := range d.Attachments {
			out.Attachments = append(out.Attachments, attachmentToOut(a))
		}
	}
	if len(d.Customers) > 0 {
		out.CustomerIDs = make([]int64, 0, len(d.Customers))
		for _, c := range d.Customers {
			if id, ok := c.ContactId.Get(); ok {
				out.CustomerIDs = append(out.CustomerIDs, id)
			}
		}
	}
	return out
}

// invoiceDraftLineToOut maps fiken.InvoiceishDraftLine into the
// canonical InvoiceDraftLineOut.
func invoiceDraftLineToOut(ln fiken.InvoiceishDraftLine) InvoiceDraftLineOut {
	out := InvoiceDraftLineOut{
		InvoiceishDraftLineID: ln.InvoiceishDraftLineId.Or(0),
		ProductID:             ln.ProductId.Or(0),
		Description:           ln.Description.Or(""),
		UnitPrice:             ln.UnitPrice.Or(0),
		VatType:               canonicalVatType(ln.VatType.Or("")),
		Quantity:              ln.Quantity,
		Discount:              ln.Discount.Or(0),
		Comment:               ln.Comment.Or(""),
		IncomeAccount:         ln.IncomeAccount.Or(""),
	}
	if dt, ok := ln.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(dt.Format("2006-01-02"))
	}
	return out
}

// InvoiceDraftsGetIn requires company + draft id.
type InvoiceDraftsGetIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// InvoiceDraftsGet returns a single invoice draft by id.
func (c *Client) InvoiceDraftsGet(ctx context.Context, in InvoiceDraftsGetIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsGet,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpInvoicesDraftsGet,
		})
	}
	resp, err := c.gen.GetInvoiceDraft(ctx, fiken.GetInvoiceDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpInvoicesDraftsGet, err))
	}
	if resp == nil {
		return Ok[InvoiceDraftOut](InvoiceDraftOut{})
	}
	return Ok[InvoiceDraftOut](invoiceDraftToOut(*resp))
}

// InvoiceDraftsCreateIn carries the create-draft payload. Body mirrors
// the upstream InvoiceishDraftRequest so the field surface stays in
// lock-step with the spec.
type InvoiceDraftsCreateIn struct {
	Company string                        `json:"company"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// InvoiceDraftsCreate posts a new invoice draft. Fiken returns 201
// with a Location header pointing at the new resource; surfaced via
// InvoiceDraftOut.Location.
func (c *Client) InvoiceDraftsCreate(ctx context.Context, in InvoiceDraftsCreateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsCreate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpInvoicesDraftsCreate,
		})
	}
	resp, err := c.gen.CreateInvoiceDraft(ctx, in.Body, fiken.CreateInvoiceDraftParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpInvoicesDraftsCreate, err))
	}
	out := InvoiceDraftOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// InvoiceDraftsUpdateIn carries the update-draft payload.
type InvoiceDraftsUpdateIn struct {
	Company string                        `json:"company"`
	DraftID int64                         `json:"draft_id"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// InvoiceDraftsUpdate replaces an existing draft in place. Upstream
// returns 201 with a Location header pointing back at the draft.
func (c *Client) InvoiceDraftsUpdate(ctx context.Context, in InvoiceDraftsUpdateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsUpdate,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpInvoicesDraftsUpdate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpInvoicesDraftsUpdate,
		})
	}
	resp, err := c.gen.UpdateInvoiceDraft(ctx, in.Body, fiken.UpdateInvoiceDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpInvoicesDraftsUpdate, err))
	}
	out := InvoiceDraftOut{DraftID: in.DraftID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// InvoiceDraftsDeleteIn requires company + draft id.
type InvoiceDraftsDeleteIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// InvoiceDraftsDeleteOut signals a successful delete.
type InvoiceDraftsDeleteOut struct {
	Deleted bool `json:"deleted"`
}

// TableHeader implements output.tableRow.
func (InvoiceDraftsDeleteOut) TableHeader() []string { return []string{"DELETED"} }

// TableRow implements output.tableRow.
func (o InvoiceDraftsDeleteOut) TableRow() []string {
	return []string{strconv.FormatBool(o.Deleted)}
}

// InvoiceDraftsDelete removes an invoice draft. Upstream returns
// 204 NoContent; we surface the success as Deleted=true.
func (c *Client) InvoiceDraftsDelete(ctx context.Context, in InvoiceDraftsDeleteIn) Result[InvoiceDraftsDeleteOut] {
	if in.Company == "" {
		return Err[InvoiceDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsDelete,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpInvoicesDraftsDelete,
		})
	}
	if err := c.gen.DeleteInvoiceDraft(ctx, fiken.DeleteInvoiceDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	}); err != nil {
		return Err[InvoiceDraftsDeleteOut](MapErr(OpInvoicesDraftsDelete, err))
	}
	return Ok[InvoiceDraftsDeleteOut](InvoiceDraftsDeleteOut{Deleted: true})
}

// InvoiceDraftsCreateFromIn turns a draft into a posted invoice.
type InvoiceDraftsCreateFromIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// InvoiceDraftsCreateFromOut surfaces the Location header pointing
// at the newly-created invoice resource.
type InvoiceDraftsCreateFromOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoiceDraftsCreateFromOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o InvoiceDraftsCreateFromOut) TableRow() []string { return []string{o.Location} }

// InvoiceDraftsCreateFrom promotes a draft into a posted invoice.
// Upstream returns 201 with the new invoice URL in Location.
func (c *Client) InvoiceDraftsCreateFrom(ctx context.Context, in InvoiceDraftsCreateFromIn) Result[InvoiceDraftsCreateFromOut] {
	if in.Company == "" {
		return Err[InvoiceDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsCreateFrom,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpInvoicesDraftsCreateFrom,
		})
	}
	resp, err := c.gen.CreateInvoiceFromDraft(ctx, fiken.CreateInvoiceFromDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftsCreateFromOut](MapErr(OpInvoicesDraftsCreateFrom, err))
	}
	out := InvoiceDraftsCreateFromOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftsCreateFromOut](out)
}

// InvoiceDraftsAttachmentsListIn requires company + draft id.
type InvoiceDraftsAttachmentsListIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// InvoiceDraftsAttachmentsList returns all attachments for an invoice
// draft. The upstream endpoint returns a bare array; reuses
// AttachmentOut from journal_entries.go.
func (c *Client) InvoiceDraftsAttachmentsList(ctx context.Context, in InvoiceDraftsAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsAttachmentsList,
		})
	}
	if in.DraftID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpInvoicesDraftsAttachmentsList,
		})
	}
	resp, err := c.gen.GetInvoiceDraftAttachments(ctx, fiken.GetInvoiceDraftAttachmentsParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpInvoicesDraftsAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// InvoiceDraftsAttachmentsAttachIn carries the multipart upload
// payload for invoice drafts. Multipart wiring lands in Plan D.
type InvoiceDraftsAttachmentsAttachIn struct {
	Company  string `json:"company"`
	DraftID  int64  `json:"draft_id"`
	Filename string `json:"filename"`
	FilePath string `json:"file_path"`
}

// InvoiceDraftsAttachmentsAttachOut mirrors the upstream Created
// response.
type InvoiceDraftsAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (InvoiceDraftsAttachmentsAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a InvoiceDraftsAttachmentsAttachOut) TableRow() []string { return []string{a.Location} }

// InvoiceDraftsAttachmentsAttach uploads a file to an invoice draft
// as multipart form data. Filename defaults to the basename of
// FilePath; pass Filename to override.
func (c *Client) InvoiceDraftsAttachmentsAttach(ctx context.Context, in InvoiceDraftsAttachmentsAttachIn) Result[InvoiceDraftsAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[InvoiceDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInvoicesDraftsAttachmentsAttach,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpInvoicesDraftsAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[InvoiceDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpInvoicesDraftsAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[InvoiceDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpInvoicesDraftsAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToInvoiceDraftReq(fiken.AddAttachmentToInvoiceDraftReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToInvoiceDraft(ctx, req, fiken.AddAttachmentToInvoiceDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftsAttachmentsAttachOut](MapErr(OpInvoicesDraftsAttachmentsAttach, err))
	}
	out := InvoiceDraftsAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftsAttachmentsAttachOut](out)
}
