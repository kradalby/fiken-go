package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// PurchaseDraftsListIn carries paged-list input for purchase drafts.
// The upstream endpoint exposes only page/page-size.
type PurchaseDraftsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// PurchaseDraftLineOut is the canonical purchase-draft-line shape.
// Monetary fields stay int64 øre. VatType surfaces the upstream enum
// verbatim — note the purchase vat-type set is broader than sales.
type PurchaseDraftLineOut struct {
	Text          string `json:"text,omitempty"`
	VatType       string `json:"vat_type,omitempty"`
	IncomeAccount string `json:"income_account,omitempty"`
	Net           int64  `json:"net,omitempty"`
	Gross         int64  `json:"gross,omitempty"`
	ProjectID     int64  `json:"project_id,omitempty"`
}

// PurchaseDraftOut is the canonical purchase-draft shape exposed to
// CLI/MCP. Monetary fields stay int64 øre. Contact is flattened to id +
// name for the table renderer.
type PurchaseDraftOut struct {
	DraftID          int64                  `json:"draft_id,omitempty"`
	UUID             string                 `json:"uuid,omitempty"`
	InvoiceIssueDate Date                   `json:"invoice_issue_date,omitempty"`
	DueDate          Date                   `json:"due_date,omitempty"`
	InvoiceNumber    string                 `json:"invoice_number,omitempty"`
	ContactID        int64                  `json:"contact_id,omitempty"`
	ContactName      string                 `json:"contact_name,omitempty"`
	ProjectID        int64                  `json:"project_id,omitempty"`
	Cash             bool                   `json:"cash,omitempty"`
	Currency         string                 `json:"currency,omitempty"`
	Kid              string                 `json:"kid,omitempty"`
	Paid             bool                   `json:"paid,omitempty"`
	Lines            []PurchaseDraftLineOut `json:"lines,omitempty"`
	Attachments      []AttachmentOut        `json:"attachments,omitempty"`
	Location         string                 `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (PurchaseDraftOut) TableHeader() []string {
	return []string{"ID", "UUID", "ISSUE", "CONTACT", "CURRENCY"}
}

// TableRow implements output.tableRow.
func (d PurchaseDraftOut) TableRow() []string {
	return []string{
		strconv.FormatInt(d.DraftID, 10),
		d.UUID,
		string(d.InvoiceIssueDate),
		d.ContactName,
		d.Currency,
	}
}

// PurchaseDraftsListOut is the paged response.
type PurchaseDraftsListOut = ListOut[PurchaseDraftOut]

// PurchaseDraftsList returns purchase drafts for the specified company.
func (c *Client) PurchaseDraftsList(ctx context.Context, in PurchaseDraftsListIn) Result[PurchaseDraftsListOut] {
	if in.Company == "" {
		return Err[PurchaseDraftsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsList,
		})
	}
	params := fiken.GetPurchaseDraftsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetPurchaseDrafts(ctx, params)
	if err != nil {
		return Err[PurchaseDraftsListOut](MapErr(OpPurchasesDraftsList, err))
	}
	return Ok[PurchaseDraftsListOut](translatePurchaseDraftsList(resp))
}

// translatePurchaseDraftsList converts the ogen response into the
// canonical ListOut[PurchaseDraftOut] envelope.
func translatePurchaseDraftsList(resp *fiken.GetPurchaseDraftsOKHeaders) PurchaseDraftsListOut {
	if resp == nil {
		return PurchaseDraftsListOut{Items: []PurchaseDraftOut{}, Meta: ListMeta{}}
	}
	items := make([]PurchaseDraftOut, 0, len(resp.Response))
	for _, d := range resp.Response {
		items = append(items, purchaseDraftToOut(d))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return PurchaseDraftsListOut{Items: items, Meta: meta}
}

// purchaseDraftToOut maps fiken.DraftResult into the canonical
// PurchaseDraftOut. Same upstream shape as sale drafts.
func purchaseDraftToOut(d fiken.DraftResult) PurchaseDraftOut {
	out := PurchaseDraftOut{
		DraftID:       d.DraftId.Or(0),
		UUID:          d.UUID.Or(""),
		InvoiceNumber: d.InvoiceNumber.Or(""),
		Cash:          d.Cash.Or(false),
		Currency:      d.Currency.Or(""),
		Kid:           d.Kid.Or(""),
		Paid:          d.Paid.Or(false),
	}
	if dt, ok := d.InvoiceIssueDate.Get(); ok {
		out.InvoiceIssueDate = Date(dt.Format("2006-01-02"))
	}
	if dt, ok := d.DueDate.Get(); ok {
		out.DueDate = Date(dt.Format("2006-01-02"))
	}
	if cust, ok := d.Contact.Get(); ok {
		out.ContactID = cust.ContactId.Or(0)
		out.ContactName = cust.Name
	}
	if p, ok := d.Project.Get(); ok {
		out.ProjectID = p.ProjectId.Or(0)
	}
	if len(d.Lines) > 0 {
		out.Lines = make([]PurchaseDraftLineOut, 0, len(d.Lines))
		for _, ln := range d.Lines {
			out.Lines = append(out.Lines, purchaseDraftLineToOut(ln))
		}
	}
	if len(d.Attachments) > 0 {
		out.Attachments = make([]AttachmentOut, 0, len(d.Attachments))
		for _, a := range d.Attachments {
			out.Attachments = append(out.Attachments, attachmentToOut(a))
		}
	}
	return out
}

// purchaseDraftLineToOut maps fiken.DraftLineResult into
// PurchaseDraftLineOut. Identical wire-shape to sales but kept as
// a separate type so the field-set can drift if the spec splits
// the schema in the future.
func purchaseDraftLineToOut(ln fiken.DraftLineResult) PurchaseDraftLineOut {
	out := PurchaseDraftLineOut{
		Text:          ln.Text.Or(""),
		VatType:       canonicalVatType(ln.VatType.Or("")),
		IncomeAccount: ln.IncomeAccount.Or(""),
		Net:           ln.Net.Or(0),
		Gross:         ln.Gross.Or(0),
	}
	if p, ok := ln.Project.Get(); ok {
		out.ProjectID = p.ProjectId.Or(0)
	}
	return out
}

// PurchaseDraftsGetIn requires company + draft id.
type PurchaseDraftsGetIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// PurchaseDraftsGet returns a single purchase draft by id.
func (c *Client) PurchaseDraftsGet(ctx context.Context, in PurchaseDraftsGetIn) Result[PurchaseDraftOut] {
	if in.Company == "" {
		return Err[PurchaseDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsGet,
		})
	}
	if in.DraftID == 0 {
		return Err[PurchaseDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpPurchasesDraftsGet,
		})
	}
	resp, err := c.gen.GetPurchaseDraft(ctx, fiken.GetPurchaseDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[PurchaseDraftOut](MapErr(OpPurchasesDraftsGet, err))
	}
	if resp == nil {
		return Ok[PurchaseDraftOut](PurchaseDraftOut{})
	}
	return Ok[PurchaseDraftOut](purchaseDraftToOut(*resp))
}

// PurchaseDraftsCreateIn carries the create-draft payload.
type PurchaseDraftsCreateIn struct {
	Company string              `json:"company"`
	Body    *fiken.DraftRequest `json:"body"`
}

// PurchaseDraftsCreate posts a new purchase draft. Upstream returns
// 201 with the new draft URL in Location.
func (c *Client) PurchaseDraftsCreate(ctx context.Context, in PurchaseDraftsCreateIn) Result[PurchaseDraftOut] {
	if in.Company == "" {
		return Err[PurchaseDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsCreate,
		})
	}
	if in.Body == nil {
		return Err[PurchaseDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpPurchasesDraftsCreate,
		})
	}
	resp, err := c.gen.CreatePurchaseDraft(ctx, in.Body, fiken.CreatePurchaseDraftParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[PurchaseDraftOut](MapErr(OpPurchasesDraftsCreate, err))
	}
	out := PurchaseDraftOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[PurchaseDraftOut](out)
}

// PurchaseDraftsUpdateIn carries the update-draft payload.
type PurchaseDraftsUpdateIn struct {
	Company string              `json:"company"`
	DraftID int64               `json:"draft_id"`
	Body    *fiken.DraftRequest `json:"body"`
}

// PurchaseDraftsUpdate replaces an existing purchase draft in place.
func (c *Client) PurchaseDraftsUpdate(ctx context.Context, in PurchaseDraftsUpdateIn) Result[PurchaseDraftOut] {
	if in.Company == "" {
		return Err[PurchaseDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsUpdate,
		})
	}
	if in.DraftID == 0 {
		return Err[PurchaseDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpPurchasesDraftsUpdate,
		})
	}
	if in.Body == nil {
		return Err[PurchaseDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpPurchasesDraftsUpdate,
		})
	}
	resp, err := c.gen.UpdatePurchaseDraft(ctx, in.Body, fiken.UpdatePurchaseDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[PurchaseDraftOut](MapErr(OpPurchasesDraftsUpdate, err))
	}
	out := PurchaseDraftOut{DraftID: in.DraftID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[PurchaseDraftOut](out)
}

// PurchaseDraftsDeleteIn requires company + draft id.
type PurchaseDraftsDeleteIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// PurchaseDraftsDeleteOut signals a successful delete.
type PurchaseDraftsDeleteOut struct {
	Deleted bool `json:"deleted"`
}

// TableHeader implements output.tableRow.
func (PurchaseDraftsDeleteOut) TableHeader() []string { return []string{"DELETED"} }

// TableRow implements output.tableRow.
func (o PurchaseDraftsDeleteOut) TableRow() []string {
	return []string{strconv.FormatBool(o.Deleted)}
}

// PurchaseDraftsDelete removes a purchase draft.
func (c *Client) PurchaseDraftsDelete(ctx context.Context, in PurchaseDraftsDeleteIn) Result[PurchaseDraftsDeleteOut] {
	if in.Company == "" {
		return Err[PurchaseDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsDelete,
		})
	}
	if in.DraftID == 0 {
		return Err[PurchaseDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpPurchasesDraftsDelete,
		})
	}
	if err := c.gen.DeletePurchaseDraft(ctx, fiken.DeletePurchaseDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	}); err != nil {
		return Err[PurchaseDraftsDeleteOut](MapErr(OpPurchasesDraftsDelete, err))
	}
	return Ok[PurchaseDraftsDeleteOut](PurchaseDraftsDeleteOut{Deleted: true})
}

// PurchaseDraftsCreateFromIn turns a draft into a posted purchase.
type PurchaseDraftsCreateFromIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// PurchaseDraftsCreateFromOut surfaces the Location header pointing
// at the newly-created purchase resource.
type PurchaseDraftsCreateFromOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (PurchaseDraftsCreateFromOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o PurchaseDraftsCreateFromOut) TableRow() []string { return []string{o.Location} }

// PurchaseDraftsCreateFrom promotes a purchase draft into a posted
// purchase. Upstream returns 201 with the new purchase URL in Location.
func (c *Client) PurchaseDraftsCreateFrom(ctx context.Context, in PurchaseDraftsCreateFromIn) Result[PurchaseDraftsCreateFromOut] {
	if in.Company == "" {
		return Err[PurchaseDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsCreateFrom,
		})
	}
	if in.DraftID == 0 {
		return Err[PurchaseDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpPurchasesDraftsCreateFrom,
		})
	}
	resp, err := c.gen.CreatePurchaseFromDraft(ctx, fiken.CreatePurchaseFromDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[PurchaseDraftsCreateFromOut](MapErr(OpPurchasesDraftsCreateFrom, err))
	}
	out := PurchaseDraftsCreateFromOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[PurchaseDraftsCreateFromOut](out)
}

// PurchaseDraftsAttachmentsListIn requires company + draft id.
type PurchaseDraftsAttachmentsListIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// PurchaseDraftsAttachmentsList returns all attachments for a purchase
// draft.
func (c *Client) PurchaseDraftsAttachmentsList(ctx context.Context, in PurchaseDraftsAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsAttachmentsList,
		})
	}
	if in.DraftID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpPurchasesDraftsAttachmentsList,
		})
	}
	resp, err := c.gen.GetPurchaseDraftAttachments(ctx, fiken.GetPurchaseDraftAttachmentsParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpPurchasesDraftsAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// PurchaseDraftsAttachmentsAttachIn carries the multipart upload payload.
type PurchaseDraftsAttachmentsAttachIn struct {
	Company  string `json:"company"`
	DraftID  int64  `json:"draft_id"`
	Filename string `json:"filename"`
	FilePath string `json:"file_path"`
}

// PurchaseDraftsAttachmentsAttachOut mirrors the upstream Created
// response.
type PurchaseDraftsAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (PurchaseDraftsAttachmentsAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a PurchaseDraftsAttachmentsAttachOut) TableRow() []string { return []string{a.Location} }

// PurchaseDraftsAttachmentsAttach uploads a file to a purchase draft.
func (c *Client) PurchaseDraftsAttachmentsAttach(ctx context.Context, in PurchaseDraftsAttachmentsAttachIn) Result[PurchaseDraftsAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[PurchaseDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpPurchasesDraftsAttachmentsAttach,
		})
	}
	if in.DraftID == 0 {
		return Err[PurchaseDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpPurchasesDraftsAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[PurchaseDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpPurchasesDraftsAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[PurchaseDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpPurchasesDraftsAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToPurchaseDraftReq(fiken.AddAttachmentToPurchaseDraftReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToPurchaseDraft(ctx, req, fiken.AddAttachmentToPurchaseDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[PurchaseDraftsAttachmentsAttachOut](MapErr(OpPurchasesDraftsAttachmentsAttach, err))
	}
	out := PurchaseDraftsAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[PurchaseDraftsAttachmentsAttachOut](out)
}
