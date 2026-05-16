package ops

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/kradalby/fiken-go/fiken"
)

// CreditNoteDraftsListIn carries paged-list input for credit-note
// drafts. The upstream endpoint only supports page/page-size; no
// content-axis filters are exposed.
type CreditNoteDraftsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// CreditNoteDraftsListOut is the paged response. The draft payload is
// the shared invoiceish shape so it reuses InvoiceDraftOut verbatim —
// the upstream type is identical and renaming would add no signal.
type CreditNoteDraftsListOut = ListOut[InvoiceDraftOut]

// CreditNoteDraftsList returns credit-note drafts for the specified
// company.
func (c *Client) CreditNoteDraftsList(ctx context.Context, in CreditNoteDraftsListIn) Result[CreditNoteDraftsListOut] {
	if in.Company == "" {
		return Err[CreditNoteDraftsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsList,
		})
	}
	params := fiken.GetCreditNoteDraftsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetCreditNoteDrafts(ctx, params)
	if err != nil {
		return Err[CreditNoteDraftsListOut](MapErr(OpCreditNotesDraftsList, err))
	}
	return Ok[CreditNoteDraftsListOut](translateCreditNoteDraftsList(resp))
}

// translateCreditNoteDraftsList converts the ogen response into the
// canonical ListOut[InvoiceDraftOut] envelope.
func translateCreditNoteDraftsList(resp *fiken.GetCreditNoteDraftsOKHeaders) CreditNoteDraftsListOut {
	if resp == nil {
		return CreditNoteDraftsListOut{Items: []InvoiceDraftOut{}, Meta: ListMeta{}}
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
	return CreditNoteDraftsListOut{Items: items, Meta: meta}
}

// CreditNoteDraftsGetIn requires company + draft id.
type CreditNoteDraftsGetIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// CreditNoteDraftsGet returns a single credit-note draft by id.
func (c *Client) CreditNoteDraftsGet(ctx context.Context, in CreditNoteDraftsGetIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsGet,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpCreditNotesDraftsGet,
		})
	}
	resp, err := c.gen.GetCreditNoteDraft(ctx, fiken.GetCreditNoteDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpCreditNotesDraftsGet, err))
	}
	if resp == nil {
		return Ok[InvoiceDraftOut](InvoiceDraftOut{})
	}
	return Ok[InvoiceDraftOut](invoiceDraftToOut(*resp))
}

// CreditNoteDraftsCreateIn carries the create-draft payload. Body
// mirrors the upstream InvoiceishDraftRequest so the field surface
// stays in lock-step with the spec.
type CreditNoteDraftsCreateIn struct {
	Company string                        `json:"company"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// CreditNoteDraftsCreate posts a new credit-note draft. Fiken returns
// 201 with a Location header pointing at the new resource; surfaced
// via InvoiceDraftOut.Location.
func (c *Client) CreditNoteDraftsCreate(ctx context.Context, in CreditNoteDraftsCreateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsCreate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpCreditNotesDraftsCreate,
		})
	}
	resp, err := c.gen.CreateCreditNoteDraft(ctx, in.Body, fiken.CreateCreditNoteDraftParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpCreditNotesDraftsCreate, err))
	}
	out := InvoiceDraftOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// CreditNoteDraftsUpdateIn carries the update-draft payload.
type CreditNoteDraftsUpdateIn struct {
	Company string                        `json:"company"`
	DraftID int64                         `json:"draft_id"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// CreditNoteDraftsUpdate replaces an existing credit-note draft in
// place. Upstream returns 201 with a Location header pointing back at
// the draft.
func (c *Client) CreditNoteDraftsUpdate(ctx context.Context, in CreditNoteDraftsUpdateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsUpdate,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpCreditNotesDraftsUpdate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpCreditNotesDraftsUpdate,
		})
	}
	resp, err := c.gen.UpdateCreditNoteDraft(ctx, in.Body, fiken.UpdateCreditNoteDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpCreditNotesDraftsUpdate, err))
	}
	out := InvoiceDraftOut{DraftID: in.DraftID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// CreditNoteDraftsDeleteIn requires company + draft id.
type CreditNoteDraftsDeleteIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// CreditNoteDraftsDeleteOut signals a successful delete. Reuses the
// existing InvoiceDraftsDeleteOut shape would couple the two tags;
// keep this distinct so future evolutions remain free.
type CreditNoteDraftsDeleteOut struct {
	Deleted bool `json:"deleted"`
}

// TableHeader implements output.tableRow.
func (CreditNoteDraftsDeleteOut) TableHeader() []string { return []string{"DELETED"} }

// TableRow implements output.tableRow.
func (o CreditNoteDraftsDeleteOut) TableRow() []string {
	if o.Deleted {
		return []string{"true"}
	}
	return []string{"false"}
}

// CreditNoteDraftsDelete removes a credit-note draft. Upstream
// returns 204 NoContent; we surface the success as Deleted=true.
func (c *Client) CreditNoteDraftsDelete(ctx context.Context, in CreditNoteDraftsDeleteIn) Result[CreditNoteDraftsDeleteOut] {
	if in.Company == "" {
		return Err[CreditNoteDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsDelete,
		})
	}
	if in.DraftID == 0 {
		return Err[CreditNoteDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpCreditNotesDraftsDelete,
		})
	}
	if err := c.gen.DeleteCreditNoteDraft(ctx, fiken.DeleteCreditNoteDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	}); err != nil {
		return Err[CreditNoteDraftsDeleteOut](MapErr(OpCreditNotesDraftsDelete, err))
	}
	return Ok[CreditNoteDraftsDeleteOut](CreditNoteDraftsDeleteOut{Deleted: true})
}

// CreditNoteDraftsCreateFromIn turns a draft into a posted credit note
// via POST /creditNotes/drafts/{draftId}/createCreditNote.
type CreditNoteDraftsCreateFromIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// CreditNoteDraftsCreateFromOut surfaces the Location header pointing
// at the newly-created credit note resource.
type CreditNoteDraftsCreateFromOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (CreditNoteDraftsCreateFromOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o CreditNoteDraftsCreateFromOut) TableRow() []string { return []string{o.Location} }

// CreditNoteDraftsCreateFrom promotes a draft into a posted credit
// note. Upstream returns 201 with the new credit-note URL in Location.
func (c *Client) CreditNoteDraftsCreateFrom(ctx context.Context, in CreditNoteDraftsCreateFromIn) Result[CreditNoteDraftsCreateFromOut] {
	if in.Company == "" {
		return Err[CreditNoteDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsCreateFrom,
		})
	}
	if in.DraftID == 0 {
		return Err[CreditNoteDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpCreditNotesDraftsCreateFrom,
		})
	}
	resp, err := c.gen.CreateCreditNoteFromDraft(ctx, fiken.CreateCreditNoteFromDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[CreditNoteDraftsCreateFromOut](MapErr(OpCreditNotesDraftsCreateFrom, err))
	}
	out := CreditNoteDraftsCreateFromOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[CreditNoteDraftsCreateFromOut](out)
}

// CreditNoteDraftsAttachmentsListIn requires company + draft id.
type CreditNoteDraftsAttachmentsListIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// CreditNoteDraftsAttachmentsList returns all attachments for a
// credit-note draft. The upstream endpoint returns a bare array;
// reuses AttachmentOut from journal_entries.go.
func (c *Client) CreditNoteDraftsAttachmentsList(ctx context.Context, in CreditNoteDraftsAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsAttachmentsList,
		})
	}
	if in.DraftID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpCreditNotesDraftsAttachmentsList,
		})
	}
	resp, err := c.gen.GetCreditNoteDraftAttachments(ctx, fiken.GetCreditNoteDraftAttachmentsParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpCreditNotesDraftsAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// CreditNoteDraftsAttachmentsAttachIn carries the multipart upload
// payload for credit-note drafts. Multipart wiring lands in Plan D.
type CreditNoteDraftsAttachmentsAttachIn struct {
	Company  string `json:"company"`
	DraftID  int64  `json:"draft_id"`
	Filename string `json:"filename"`
	FilePath string `json:"file_path"`
}

// CreditNoteDraftsAttachmentsAttachOut mirrors the upstream Created
// response.
type CreditNoteDraftsAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (CreditNoteDraftsAttachmentsAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a CreditNoteDraftsAttachmentsAttachOut) TableRow() []string { return []string{a.Location} }

// CreditNoteDraftsAttachmentsAttach uploads a file to a credit-note
// draft as multipart form data. Filename defaults to the basename of
// FilePath; pass Filename to override.
func (c *Client) CreditNoteDraftsAttachmentsAttach(ctx context.Context, in CreditNoteDraftsAttachmentsAttachIn) Result[CreditNoteDraftsAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[CreditNoteDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpCreditNotesDraftsAttachmentsAttach,
		})
	}
	if in.DraftID == 0 {
		return Err[CreditNoteDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpCreditNotesDraftsAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[CreditNoteDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpCreditNotesDraftsAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[CreditNoteDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpCreditNotesDraftsAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToCreditNoteDraftReq(fiken.AddAttachmentToCreditNoteDraftReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToCreditNoteDraft(ctx, req, fiken.AddAttachmentToCreditNoteDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[CreditNoteDraftsAttachmentsAttachOut](MapErr(OpCreditNotesDraftsAttachmentsAttach, err))
	}
	out := CreditNoteDraftsAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[CreditNoteDraftsAttachmentsAttachOut](out)
}
