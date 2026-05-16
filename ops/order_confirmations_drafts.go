package ops

import (
	"context"
	"net/url"
	"path/filepath"

	"github.com/kradalby/fiken-go/fiken"
)

// OrderConfirmationDraftsListIn carries paged-list input for order
// confirmation drafts. The upstream endpoint only supports
// page/page-size; no content-axis filters are exposed.
type OrderConfirmationDraftsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// OrderConfirmationDraftsListOut is the paged response. The draft
// payload is the shared invoiceish shape so it reuses InvoiceDraftOut
// verbatim — the upstream type is identical and renaming would add no
// signal.
type OrderConfirmationDraftsListOut = ListOut[InvoiceDraftOut]

// OrderConfirmationDraftsList returns order confirmation drafts for
// the specified company.
func (c *Client) OrderConfirmationDraftsList(ctx context.Context, in OrderConfirmationDraftsListIn) Result[OrderConfirmationDraftsListOut] {
	if in.Company == "" {
		return Err[OrderConfirmationDraftsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsList,
		})
	}
	params := fiken.GetOrderConfirmationDraftsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetOrderConfirmationDrafts(ctx, params)
	if err != nil {
		return Err[OrderConfirmationDraftsListOut](MapErr(OpOrderConfirmationsDraftsList, err))
	}
	return Ok[OrderConfirmationDraftsListOut](translateOrderConfirmationDraftsList(resp))
}

// translateOrderConfirmationDraftsList converts the ogen response into
// the canonical ListOut[InvoiceDraftOut] envelope.
func translateOrderConfirmationDraftsList(resp *fiken.GetOrderConfirmationDraftsOKHeaders) OrderConfirmationDraftsListOut {
	if resp == nil {
		return OrderConfirmationDraftsListOut{Items: []InvoiceDraftOut{}, Meta: ListMeta{}}
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
	return OrderConfirmationDraftsListOut{Items: items, Meta: meta}
}

// OrderConfirmationDraftsGetIn requires company + draft id.
type OrderConfirmationDraftsGetIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OrderConfirmationDraftsGet returns a single order confirmation draft
// by id.
func (c *Client) OrderConfirmationDraftsGet(ctx context.Context, in OrderConfirmationDraftsGetIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsGet,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOrderConfirmationsDraftsGet,
		})
	}
	resp, err := c.gen.GetOrderConfirmationDraft(ctx, fiken.GetOrderConfirmationDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpOrderConfirmationsDraftsGet, err))
	}
	if resp == nil {
		return Ok[InvoiceDraftOut](InvoiceDraftOut{})
	}
	return Ok[InvoiceDraftOut](invoiceDraftToOut(*resp))
}

// OrderConfirmationDraftsCreateIn carries the create-draft payload.
// Body mirrors the upstream InvoiceishDraftRequest so the field
// surface stays in lock-step with the spec.
type OrderConfirmationDraftsCreateIn struct {
	Company string                        `json:"company"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// OrderConfirmationDraftsCreate posts a new order confirmation draft.
// Upstream returns 201 with a Location header pointing at the new
// draft; surfaced via InvoiceDraftOut.Location.
func (c *Client) OrderConfirmationDraftsCreate(ctx context.Context, in OrderConfirmationDraftsCreateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsCreate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpOrderConfirmationsDraftsCreate,
		})
	}
	resp, err := c.gen.CreateOrderConfirmationDraft(ctx, in.Body, fiken.CreateOrderConfirmationDraftParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpOrderConfirmationsDraftsCreate, err))
	}
	out := InvoiceDraftOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// OrderConfirmationDraftsUpdateIn carries the update-draft payload.
type OrderConfirmationDraftsUpdateIn struct {
	Company string                        `json:"company"`
	DraftID int64                         `json:"draft_id"`
	Body    *fiken.InvoiceishDraftRequest `json:"body"`
}

// OrderConfirmationDraftsUpdate replaces an existing order confirmation
// draft in place. Upstream returns 201 with a Location header pointing
// back at the draft.
func (c *Client) OrderConfirmationDraftsUpdate(ctx context.Context, in OrderConfirmationDraftsUpdateIn) Result[InvoiceDraftOut] {
	if in.Company == "" {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsUpdate,
		})
	}
	if in.DraftID == 0 {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOrderConfirmationsDraftsUpdate,
		})
	}
	if in.Body == nil {
		return Err[InvoiceDraftOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpOrderConfirmationsDraftsUpdate,
		})
	}
	resp, err := c.gen.UpdateOrderConfirmationDraft(ctx, in.Body, fiken.UpdateOrderConfirmationDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[InvoiceDraftOut](MapErr(OpOrderConfirmationsDraftsUpdate, err))
	}
	out := InvoiceDraftOut{DraftID: in.DraftID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InvoiceDraftOut](out)
}

// OrderConfirmationDraftsDeleteIn requires company + draft id.
type OrderConfirmationDraftsDeleteIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OrderConfirmationDraftsDeleteOut signals a successful delete.
// Distinct from sibling delete-out shapes so future evolutions stay
// independent.
type OrderConfirmationDraftsDeleteOut struct {
	Deleted bool `json:"deleted"`
}

// TableHeader implements output.tableRow.
func (OrderConfirmationDraftsDeleteOut) TableHeader() []string { return []string{"DELETED"} }

// TableRow implements output.tableRow.
func (o OrderConfirmationDraftsDeleteOut) TableRow() []string {
	if o.Deleted {
		return []string{"true"}
	}
	return []string{"false"}
}

// OrderConfirmationDraftsDelete removes an order confirmation draft.
// Upstream returns 204 NoContent; we surface the success as
// Deleted=true.
func (c *Client) OrderConfirmationDraftsDelete(ctx context.Context, in OrderConfirmationDraftsDeleteIn) Result[OrderConfirmationDraftsDeleteOut] {
	if in.Company == "" {
		return Err[OrderConfirmationDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsDelete,
		})
	}
	if in.DraftID == 0 {
		return Err[OrderConfirmationDraftsDeleteOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOrderConfirmationsDraftsDelete,
		})
	}
	if err := c.gen.DeleteOrderConfirmationDraft(ctx, fiken.DeleteOrderConfirmationDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	}); err != nil {
		return Err[OrderConfirmationDraftsDeleteOut](MapErr(OpOrderConfirmationsDraftsDelete, err))
	}
	return Ok[OrderConfirmationDraftsDeleteOut](OrderConfirmationDraftsDeleteOut{Deleted: true})
}

// OrderConfirmationDraftsCreateFromIn turns a draft into a posted
// order confirmation via POST
// /orderConfirmations/drafts/{draftId}/createOrderConfirmation.
type OrderConfirmationDraftsCreateFromIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OrderConfirmationDraftsCreateFromOut surfaces the Location header
// pointing at the newly-created order confirmation resource.
type OrderConfirmationDraftsCreateFromOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (OrderConfirmationDraftsCreateFromOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (o OrderConfirmationDraftsCreateFromOut) TableRow() []string { return []string{o.Location} }

// OrderConfirmationDraftsCreateFrom promotes a draft into a posted
// order confirmation. Upstream returns 201 with the new resource URL
// in Location.
func (c *Client) OrderConfirmationDraftsCreateFrom(ctx context.Context, in OrderConfirmationDraftsCreateFromIn) Result[OrderConfirmationDraftsCreateFromOut] {
	if in.Company == "" {
		return Err[OrderConfirmationDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsCreateFrom,
		})
	}
	if in.DraftID == 0 {
		return Err[OrderConfirmationDraftsCreateFromOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOrderConfirmationsDraftsCreateFrom,
		})
	}
	resp, err := c.gen.CreateOrderConfirmationFromDraft(ctx, fiken.CreateOrderConfirmationFromDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[OrderConfirmationDraftsCreateFromOut](MapErr(OpOrderConfirmationsDraftsCreateFrom, err))
	}
	out := OrderConfirmationDraftsCreateFromOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[OrderConfirmationDraftsCreateFromOut](out)
}

// OrderConfirmationDraftsAttachmentsListIn requires company + draft id.
type OrderConfirmationDraftsAttachmentsListIn struct {
	Company string `json:"company"`
	DraftID int64  `json:"draft_id"`
}

// OrderConfirmationDraftsAttachmentsList returns all attachments for
// an order confirmation draft. The upstream endpoint returns a bare
// array; reuses AttachmentOut from journal_entries.go.
func (c *Client) OrderConfirmationDraftsAttachmentsList(ctx context.Context, in OrderConfirmationDraftsAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsAttachmentsList,
		})
	}
	if in.DraftID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOrderConfirmationsDraftsAttachmentsList,
		})
	}
	resp, err := c.gen.GetOrderConfirmationDraftAttachments(ctx, fiken.GetOrderConfirmationDraftAttachmentsParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpOrderConfirmationsDraftsAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// OrderConfirmationDraftsAttachmentsAttachIn carries the multipart
// upload payload for order confirmation drafts. Multipart wiring lands
// in Plan D.
type OrderConfirmationDraftsAttachmentsAttachIn struct {
	Company  string `json:"company"`
	DraftID  int64  `json:"draft_id"`
	Filename string `json:"filename"`
	FilePath string `json:"file_path"`
}

// OrderConfirmationDraftsAttachmentsAttachOut mirrors the upstream
// Created response.
type OrderConfirmationDraftsAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (OrderConfirmationDraftsAttachmentsAttachOut) TableHeader() []string {
	return []string{"LOCATION"}
}

// TableRow implements output.tableRow.
func (a OrderConfirmationDraftsAttachmentsAttachOut) TableRow() []string {
	return []string{a.Location}
}

// OrderConfirmationDraftsAttachmentsAttach uploads a file to an
// order-confirmation draft as multipart form data. Filename defaults
// to the basename of FilePath; pass Filename to override.
func (c *Client) OrderConfirmationDraftsAttachmentsAttach(ctx context.Context, in OrderConfirmationDraftsAttachmentsAttachIn) Result[OrderConfirmationDraftsAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[OrderConfirmationDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpOrderConfirmationsDraftsAttachmentsAttach,
		})
	}
	if in.DraftID == 0 {
		return Err[OrderConfirmationDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "draft_id is required", Op: OpOrderConfirmationsDraftsAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[OrderConfirmationDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpOrderConfirmationsDraftsAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[OrderConfirmationDraftsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpOrderConfirmationsDraftsAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToOrderConfirmationDraftReq(fiken.AddAttachmentToOrderConfirmationDraftReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToOrderConfirmationDraft(ctx, req, fiken.AddAttachmentToOrderConfirmationDraftParams{
		CompanySlug: in.Company,
		DraftId:     in.DraftID,
	})
	if err != nil {
		return Err[OrderConfirmationDraftsAttachmentsAttachOut](MapErr(OpOrderConfirmationsDraftsAttachmentsAttach, err))
	}
	out := OrderConfirmationDraftsAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[OrderConfirmationDraftsAttachmentsAttachOut](out)
}
