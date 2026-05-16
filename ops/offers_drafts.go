package ops

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/kradalby/fiken-go/fiken"
)

// OfferDraftsListIn carries paged-list input for offer drafts. The
// upstream endpoint only supports page/page-size; no content-axis
// filters are exposed.
type OfferDraftsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// OfferDraftsListOut is the paged response. The draft payload is the
// shared invoiceish shape so it reuses InvoiceDraftOut verbatim — the
// upstream type is identical and renaming would add no signal.
type OfferDraftsListOut = ListOut[InvoiceDraftOut]

// OfferDraftsList returns offer drafts for the specified company.
func (c *Client) OfferDraftsList(ctx context.Context, in OfferDraftsListIn) Result[OfferDraftsListOut] {
	if in.Company == "" {
		return Err[OfferDraftsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsList,
		})
	}
	params := fiken.GetOfferDraftsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetOfferDrafts(ctx, params)
	if err != nil {
		return Err[OfferDraftsListOut](MapErr(OpOffersDraftsList, err))
	}
	return Ok[OfferDraftsListOut](translateOfferDraftsList(resp))
}

// translateOfferDraftsList converts the ogen response into the
// canonical ListOut[InvoiceDraftOut] envelope.
func translateOfferDraftsList(resp *fiken.GetOfferDraftsOKHeaders) OfferDraftsListOut {
	if resp == nil {
		return OfferDraftsListOut{Items: []InvoiceDraftOut{}, Meta: ListMeta{}}
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
	return OfferDraftsListOut{Items: items, Meta: meta}
}

// OfferDraftsGetIn requires company + draft id.
type OfferDraftsGetIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OfferDraftsGet returns a single offer draft by id.
func (c *Client) OfferDraftsGet(ctx context.Context, in OfferDraftsGetIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsGet,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOffersDraftsGet,
		})
	}
	resp, err := c.gen.GetOfferDraft(ctx, fiken.GetOfferDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpOffersDraftsGet, err))
	}
	if resp == nil {
		return Ok[InvoiceDraftOut](InvoiceDraftOut{})
	}
	return Ok[InvoiceDraftOut](invoiceDraftToOut(*resp))
}

// OfferDraftsCreateIn carries the create-draft payload. Body mirrors
// the upstream InvoiceishDraftRequest so the field surface stays in
// lock-step with the spec.
type OfferDraftsCreateIn struct {
	Company string                        `json:"company"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// OfferDraftsCreate posts a new offer draft. Fiken returns 201 with a
// Location header pointing at the new resource; surfaced via
// InvoiceDraftOut.Location.
func (c *Client) OfferDraftsCreate(ctx context.Context, in OfferDraftsCreateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsCreate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpOffersDraftsCreate,
		})
	}
	resp, err := c.gen.CreateOfferDraft(ctx, in.Body, fiken.CreateOfferDraftParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpOffersDraftsCreate, err))
	}
	out := InvoiceDraftOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// OfferDraftsUpdateIn carries the update-draft payload.
type OfferDraftsUpdateIn struct {
	Company string                        `json:"company"`
	DraftID int64                         `json:"draft_id"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// OfferDraftsUpdate replaces an existing offer draft in place.
// Upstream returns 201 with a Location header pointing back at the
// draft.
func (c *Client) OfferDraftsUpdate(ctx context.Context, in OfferDraftsUpdateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsUpdate,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOffersDraftsUpdate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpOffersDraftsUpdate,
		})
	}
	resp, err := c.gen.UpdateOfferDraft(ctx, in.Body, fiken.UpdateOfferDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpOffersDraftsUpdate, err))
	}
	out := InvoiceDraftOut{DraftID: in.DraftID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// OfferDraftsDeleteIn requires company + draft id.
type OfferDraftsDeleteIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OfferDraftsDeleteOut signals a successful delete. Distinct from the
// invoice / credit-note delete-out shapes so future evolutions stay
// independent.
type OfferDraftsDeleteOut struct {
	Deleted bool `json:"deleted"`
}

// TableHeader implements output.tableRow.
func (OfferDraftsDeleteOut) TableHeader() []string { return []string{"DELETED"} }

// TableRow implements output.tableRow.
func (o OfferDraftsDeleteOut) TableRow() []string {
	if o.Deleted {
		return []string{"true"}
	}
	return []string{"false"}
}

// OfferDraftsDelete removes an offer draft. Upstream returns 204
// NoContent; we surface the success as Deleted=true.
func (c *Client) OfferDraftsDelete(ctx context.Context, in OfferDraftsDeleteIn) Result[OfferDraftsDeleteOut] {
	if in.Company == "" {
		return Err[OfferDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsDelete,
		})
	}
	if in.DraftID == 0 {
		return Err[OfferDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOffersDraftsDelete,
		})
	}
	if err := c.gen.DeleteOfferDraft(ctx, fiken.DeleteOfferDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	}); err != nil {
		return Err[OfferDraftsDeleteOut](MapErr(OpOffersDraftsDelete, err))
	}
	return Ok[OfferDraftsDeleteOut](OfferDraftsDeleteOut{Deleted: true})
}

// OfferDraftsCreateFromIn turns a draft into a posted offer via
// POST /offers/drafts/{draftId}/createOffer.
type OfferDraftsCreateFromIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OfferDraftsCreateFromOut surfaces the Location header pointing at
// the newly-created offer resource.
type OfferDraftsCreateFromOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (OfferDraftsCreateFromOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o OfferDraftsCreateFromOut) TableRow() []string { return []string{o.Location} }

// OfferDraftsCreateFrom promotes a draft into a posted offer.
// Upstream returns 201 with the new offer URL in Location.
func (c *Client) OfferDraftsCreateFrom(ctx context.Context, in OfferDraftsCreateFromIn) Result[OfferDraftsCreateFromOut] {
	if in.Company == "" {
		return Err[OfferDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsCreateFrom,
		})
	}
	if in.DraftID == 0 {
		return Err[OfferDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOffersDraftsCreateFrom,
		})
	}
	resp, err := c.gen.CreateOfferFromDraft(ctx, fiken.CreateOfferFromDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[OfferDraftsCreateFromOut](MapErr(OpOffersDraftsCreateFrom, err))
	}
	out := OfferDraftsCreateFromOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[OfferDraftsCreateFromOut](out)
}

// OfferDraftsAttachmentsListIn requires company + draft id.
type OfferDraftsAttachmentsListIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OfferDraftsAttachmentsList returns all attachments for an offer
// draft. The upstream endpoint returns a bare array; reuses
// AttachmentOut from journal_entries.go.
func (c *Client) OfferDraftsAttachmentsList(ctx context.Context, in OfferDraftsAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsAttachmentsList,
		})
	}
	if in.DraftID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOffersDraftsAttachmentsList,
		})
	}
	resp, err := c.gen.GetOfferDraftAttachments(ctx, fiken.GetOfferDraftAttachmentsParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpOffersDraftsAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// OfferDraftsAttachmentsAttachIn carries the multipart upload payload
// for offer drafts. Multipart wiring lands in Plan D.
type OfferDraftsAttachmentsAttachIn struct {
	Company  string `json:"company"`
	DraftID  int64  `json:"draft_id"`
	Filename string `json:"filename"`
	FilePath string `json:"file_path"`
}

// OfferDraftsAttachmentsAttachOut mirrors the upstream Created
// response.
type OfferDraftsAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (OfferDraftsAttachmentsAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a OfferDraftsAttachmentsAttachOut) TableRow() []string { return []string{a.Location} }

// OfferDraftsAttachmentsAttach uploads a file to an offer draft as
// multipart form data. Filename defaults to the basename of FilePath;
// pass Filename to override.
func (c *Client) OfferDraftsAttachmentsAttach(ctx context.Context, in OfferDraftsAttachmentsAttachIn) Result[OfferDraftsAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[OfferDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOffersDraftsAttachmentsAttach,
		})
	}
	if in.DraftID == 0 {
		return Err[OfferDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOffersDraftsAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[OfferDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpOffersDraftsAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[OfferDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpOffersDraftsAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToOfferDraftReq(fiken.AddAttachmentToOfferDraftReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToOfferDraft(ctx, req, fiken.AddAttachmentToOfferDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[OfferDraftsAttachmentsAttachOut](MapErr(OpOffersDraftsAttachmentsAttach, err))
	}
	out := OfferDraftsAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[OfferDraftsAttachmentsAttachOut](out)
}
