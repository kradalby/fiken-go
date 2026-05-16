package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// SaleDraftsListIn carries paged-list input for sale drafts. The
// upstream endpoint exposes only page/page-size; no content-axis
// filters are surfaced.
type SaleDraftsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// SaleDraftLineOut is the canonical sale-draft-line shape. Monetary
// fields stay int64 øre. VatType surfaces the upstream enum verbatim.
type SaleDraftLineOut struct {
	Text          string `json:"text,omitempty"`
	VatType       string `json:"vat_type,omitempty"`
	IncomeAccount string `json:"income_account,omitempty"`
	Net           int64  `json:"net,omitempty"`
	Gross         int64  `json:"gross,omitempty"`
	ProjectID     int64  `json:"project_id,omitempty"`
}

// SaleDraftOut is the canonical sale-draft shape exposed to CLI/MCP.
// Monetary fields stay int64 øre. Customer is flattened to id + name
// for the table renderer.
type SaleDraftOut struct {
	DraftID          int64              `json:"draft_id,omitempty"`
	UUID             string             `json:"uuid,omitempty"`
	InvoiceIssueDate Date               `json:"invoice_issue_date,omitempty"`
	DueDate          Date               `json:"due_date,omitempty"`
	InvoiceNumber    string             `json:"invoice_number,omitempty"`
	ContactID        int64              `json:"contact_id,omitempty"`
	ContactName      string             `json:"contact_name,omitempty"`
	ProjectID        int64              `json:"project_id,omitempty"`
	Cash             bool               `json:"cash,omitempty"`
	Currency         string             `json:"currency,omitempty"`
	Kid              string             `json:"kid,omitempty"`
	Paid             bool               `json:"paid,omitempty"`
	Lines            []SaleDraftLineOut `json:"lines,omitempty"`
	Attachments      []AttachmentOut    `json:"attachments,omitempty"`
	Location         string             `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (SaleDraftOut) TableHeader() []string {
	return []string{"ID", "UUID", "ISSUE", "CONTACT", "CURRENCY"}
}

// TableRow implements output.tableRow.
func (d SaleDraftOut) TableRow() []string {
	return []string{
		strconv.FormatInt(d.DraftID, 10),
		d.UUID,
		string(d.InvoiceIssueDate),
		d.ContactName,
		d.Currency,
	}
}

// SaleDraftsListOut is the paged response.
type SaleDraftsListOut = ListOut[SaleDraftOut]

// SaleDraftsList returns sale drafts for the specified company.
func (c *Client) SaleDraftsList(ctx context.Context, in SaleDraftsListIn) Result[SaleDraftsListOut] {
	if in.Company == "" {
		return Err[SaleDraftsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsList,
		})
	}
	params := fiken.GetSaleDraftsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetSaleDrafts(ctx, params)
	if err != nil {
		return Err[SaleDraftsListOut](MapErr(OpSalesDraftsList, err))
	}
	return Ok[SaleDraftsListOut](translateSaleDraftsList(resp))
}

// translateSaleDraftsList converts the ogen response into the
// canonical ListOut[SaleDraftOut] envelope.
func translateSaleDraftsList(resp *fiken.GetSaleDraftsOKHeaders) SaleDraftsListOut {
	if resp == nil {
		return SaleDraftsListOut{Items: []SaleDraftOut{}, Meta: ListMeta{}}
	}
	items := make([]SaleDraftOut, 0, len(resp.Response))
	for _, d := range resp.Response {
		items = append(items, saleDraftToOut(d))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return SaleDraftsListOut{Items: items, Meta: meta}
}

// saleDraftToOut maps fiken.DraftResult into the canonical SaleDraftOut.
// Customer / project are flattened to ids + names; the full upstream
// structs remain reachable via the underlying ogen Result.
func saleDraftToOut(d fiken.DraftResult) SaleDraftOut {
	out := SaleDraftOut{
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
		out.Lines = make([]SaleDraftLineOut, 0, len(d.Lines))
		for _, ln := range d.Lines {
			out.Lines = append(out.Lines, draftLineToOut(ln))
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

// draftLineToOut maps a fiken.DraftLineResult into SaleDraftLineOut.
// Used by both sale and purchase draft translators since the upstream
// shape is identical (the spec's `draftLineResult` is shared).
func draftLineToOut(ln fiken.DraftLineResult) SaleDraftLineOut {
	out := SaleDraftLineOut{
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

// SaleDraftsGetIn requires company + draft id.
type SaleDraftsGetIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// SaleDraftsGet returns a single sale draft by id.
func (c *Client) SaleDraftsGet(ctx context.Context, in SaleDraftsGetIn) Result[SaleDraftOut] {
	if in.Company == "" {
		return Err[SaleDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsGet,
		})
	}
	if in.DraftID == 0 {
		return Err[SaleDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpSalesDraftsGet,
		})
	}
	resp, err := c.gen.GetSaleDraft(ctx, fiken.GetSaleDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[SaleDraftOut](MapErr(OpSalesDraftsGet, err))
	}
	if resp == nil {
		return Ok[SaleDraftOut](SaleDraftOut{})
	}
	return Ok[SaleDraftOut](saleDraftToOut(*resp))
}

// SaleDraftsCreateIn carries the create-draft payload. Body mirrors
// the upstream DraftRequest.
type SaleDraftsCreateIn struct {
	Company string              `json:"company"`
	Body    *fiken.DraftRequest `json:"body"`
}

// SaleDraftsCreate posts a new sale draft. Upstream returns 201 with
// the new draft URL in Location.
func (c *Client) SaleDraftsCreate(ctx context.Context, in SaleDraftsCreateIn) Result[SaleDraftOut] {
	if in.Company == "" {
		return Err[SaleDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsCreate,
		})
	}
	if in.Body == nil {
		return Err[SaleDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpSalesDraftsCreate,
		})
	}
	resp, err := c.gen.CreateSaleDraft(ctx, in.Body, fiken.CreateSaleDraftParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[SaleDraftOut](MapErr(OpSalesDraftsCreate, err))
	}
	out := SaleDraftOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[SaleDraftOut](out)
}

// SaleDraftsUpdateIn carries the update-draft payload.
type SaleDraftsUpdateIn struct {
	Company string              `json:"company"`
	DraftID int64               `json:"draft_id"`
	Body    *fiken.DraftRequest `json:"body"`
}

// SaleDraftsUpdate replaces an existing sale draft in place. Upstream
// returns 201 with a Location header pointing back at the draft.
func (c *Client) SaleDraftsUpdate(ctx context.Context, in SaleDraftsUpdateIn) Result[SaleDraftOut] {
	if in.Company == "" {
		return Err[SaleDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsUpdate,
		})
	}
	if in.DraftID == 0 {
		return Err[SaleDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpSalesDraftsUpdate,
		})
	}
	if in.Body == nil {
		return Err[SaleDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpSalesDraftsUpdate,
		})
	}
	resp, err := c.gen.UpdateSaleDraft(ctx, in.Body, fiken.UpdateSaleDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[SaleDraftOut](MapErr(OpSalesDraftsUpdate, err))
	}
	out := SaleDraftOut{DraftID: in.DraftID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[SaleDraftOut](out)
}

// SaleDraftsDeleteIn requires company + draft id.
type SaleDraftsDeleteIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// SaleDraftsDeleteOut signals a successful delete.
type SaleDraftsDeleteOut struct {
	Deleted bool `json:"deleted"`
}

// TableHeader implements output.tableRow.
func (SaleDraftsDeleteOut) TableHeader() []string { return []string{"DELETED"} }

// TableRow implements output.tableRow.
func (o SaleDraftsDeleteOut) TableRow() []string {
	return []string{strconv.FormatBool(o.Deleted)}
}

// SaleDraftsDelete removes a sale draft. Upstream returns 204
// NoContent; we surface success as Deleted=true.
func (c *Client) SaleDraftsDelete(ctx context.Context, in SaleDraftsDeleteIn) Result[SaleDraftsDeleteOut] {
	if in.Company == "" {
		return Err[SaleDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsDelete,
		})
	}
	if in.DraftID == 0 {
		return Err[SaleDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpSalesDraftsDelete,
		})
	}
	if err := c.gen.DeleteSaleDraft(ctx, fiken.DeleteSaleDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	}); err != nil {
		return Err[SaleDraftsDeleteOut](MapErr(OpSalesDraftsDelete, err))
	}
	return Ok[SaleDraftsDeleteOut](SaleDraftsDeleteOut{Deleted: true})
}

// SaleDraftsCreateFromIn turns a draft into a posted sale.
type SaleDraftsCreateFromIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// SaleDraftsCreateFromOut surfaces the Location header pointing at
// the newly-created sale resource.
type SaleDraftsCreateFromOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (SaleDraftsCreateFromOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o SaleDraftsCreateFromOut) TableRow() []string { return []string{o.Location} }

// SaleDraftsCreateFrom promotes a sale draft into a posted sale.
// Upstream returns 201 with the new sale URL in Location.
func (c *Client) SaleDraftsCreateFrom(ctx context.Context, in SaleDraftsCreateFromIn) Result[SaleDraftsCreateFromOut] {
	if in.Company == "" {
		return Err[SaleDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsCreateFrom,
		})
	}
	if in.DraftID == 0 {
		return Err[SaleDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpSalesDraftsCreateFrom,
		})
	}
	resp, err := c.gen.CreateSaleFromDraft(ctx, fiken.CreateSaleFromDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[SaleDraftsCreateFromOut](MapErr(OpSalesDraftsCreateFrom, err))
	}
	out := SaleDraftsCreateFromOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[SaleDraftsCreateFromOut](out)
}

// SaleDraftsAttachmentsListIn requires company + draft id.
type SaleDraftsAttachmentsListIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// SaleDraftsAttachmentsList returns all attachments for a sale draft.
func (c *Client) SaleDraftsAttachmentsList(ctx context.Context, in SaleDraftsAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsAttachmentsList,
		})
	}
	if in.DraftID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpSalesDraftsAttachmentsList,
		})
	}
	resp, err := c.gen.GetSaleDraftAttachments(ctx, fiken.GetSaleDraftAttachmentsParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpSalesDraftsAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// SaleDraftsAttachmentsAttachIn carries the multipart upload payload.
type SaleDraftsAttachmentsAttachIn struct {
	Company  string `json:"company"`
	DraftID  int64  `json:"draft_id"`
	Filename string `json:"filename"`
	FilePath string `json:"file_path"`
}

// SaleDraftsAttachmentsAttachOut mirrors the upstream Created response.
type SaleDraftsAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (SaleDraftsAttachmentsAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a SaleDraftsAttachmentsAttachOut) TableRow() []string { return []string{a.Location} }

// SaleDraftsAttachmentsAttach uploads a file to a sale draft as
// multipart form data. Filename defaults to the basename of FilePath.
func (c *Client) SaleDraftsAttachmentsAttach(ctx context.Context, in SaleDraftsAttachmentsAttachIn) Result[SaleDraftsAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[SaleDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpSalesDraftsAttachmentsAttach,
		})
	}
	if in.DraftID == 0 {
		return Err[SaleDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpSalesDraftsAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[SaleDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpSalesDraftsAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[SaleDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpSalesDraftsAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToSaleDraftReq(fiken.AddAttachmentToSaleDraftReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToSaleDraft(ctx, req, fiken.AddAttachmentToSaleDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[SaleDraftsAttachmentsAttachOut](MapErr(OpSalesDraftsAttachmentsAttach, err))
	}
	out := SaleDraftsAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[SaleDraftsAttachmentsAttachOut](out)
}
