package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kradalby/fiken-go/fiken"
)

// InboxListIn carries paged-list input for inbox documents. The
// upstream endpoint exposes sortBy + status enums plus a case-
// insensitive name substring filter. Sort/status are surfaced as
// raw strings so callers (CLI, MCP) round-trip the upstream literals
// ("createdDate asc", "unused", ...) without translation.
type InboxListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`
	Status   string `json:"status,omitempty"`
	Name     string `json:"name,omitempty"`
}

// InboxDocumentOut is the canonical inbox-document shape. CreatedAt
// is serialised as RFC 3339 to match the upstream date-time format
// and stay consistent with how table renderers handle timestamps.
// Status is the upstream "used" flag (true == used as documentation).
type InboxDocumentOut struct {
	DocumentID  int64  `json:"document_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Status      bool   `json:"status,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	DocumentURL string `json:"document_url,omitempty"`
}

// TableHeader implements output.tableRow.
func (InboxDocumentOut) TableHeader() []string {
	return []string{"ID", "NAME", "FILENAME", "USED", "CREATED"}
}

// TableRow implements output.tableRow.
func (d InboxDocumentOut) TableRow() []string {
	return []string{
		strconv.FormatInt(d.DocumentID, 10),
		d.Name,
		d.Filename,
		strconv.FormatBool(d.Status),
		d.CreatedAt,
	}
}

// InboxListOut is the paged response.
type InboxListOut = ListOut[InboxDocumentOut]

// InboxList returns inbox documents for the specified company.
func (c *Client) InboxList(ctx context.Context, in InboxListIn) Result[InboxListOut] {
	if in.Company == "" {
		return Err[InboxListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInboxList,
		})
	}
	params := fiken.GetInboxParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.SortBy != "" {
		params.SortBy.SetTo(fiken.GetInboxSortBy(in.SortBy))
	}
	if in.Status != "" {
		params.Status.SetTo(fiken.GetInboxStatus(in.Status))
	}
	if in.Name != "" {
		params.Name.SetTo(in.Name)
	}
	resp, err := c.gen.GetInbox(ctx, params)
	if err != nil {
		return Err[InboxListOut](MapErr(OpInboxList, err))
	}
	return Ok[InboxListOut](translateInboxList(resp))
}

// translateInboxList converts the ogen response into the canonical
// ListOut[InboxDocumentOut] envelope, including paging meta.
func translateInboxList(resp *fiken.GetInboxOKHeaders) InboxListOut {
	if resp == nil {
		return InboxListOut{Items: []InboxDocumentOut{}, Meta: ListMeta{}}
	}
	items := make([]InboxDocumentOut, 0, len(resp.Response))
	for _, d := range resp.Response {
		items = append(items, inboxDocumentToOut(d))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return InboxListOut{Items: items, Meta: meta}
}

// inboxDocumentToOut maps fiken.InboxResult into the canonical
// InboxDocumentOut. CreatedAt is rendered as RFC 3339 to preserve the
// upstream timestamp without locale-specific formatting.
func inboxDocumentToOut(d fiken.InboxResult) InboxDocumentOut {
	out := InboxDocumentOut{
		DocumentID:  d.DocumentId.Or(0),
		Name:        d.Name.Or(""),
		Description: d.Description.Or(""),
		Filename:    d.Filename.Or(""),
		Status:      d.Status.Or(false),
		DocumentURL: d.DocumentUrl.Or(""),
	}
	if t, ok := d.CreatedAt.Get(); ok {
		out.CreatedAt = t.Format(time.RFC3339)
	}
	return out
}

// InboxGetIn requires company + inbox document id.
type InboxGetIn struct {
	Company    string `json:"company"`
	DocumentID int64  `json:"document_id"`
}

// InboxGet returns a single inbox document by id.
func (c *Client) InboxGet(ctx context.Context, in InboxGetIn) Result[InboxDocumentOut] {
	if in.Company == "" {
		return Err[InboxDocumentOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInboxGet,
		})
	}
	if in.DocumentID == 0 {
		return Err[InboxDocumentOut](&Error{
			Code: CodeValidation, Message: "document_id is required", Op: OpInboxGet,
		})
	}
	resp, err := c.gen.GetInboxDocument(ctx, fiken.GetInboxDocumentParams{
		CompanySlug:     in.Company,
		InboxDocumentId: in.DocumentID,
	})
	if err != nil {
		return Err[InboxDocumentOut](MapErr(OpInboxGet, err))
	}
	if resp == nil {
		return Ok[InboxDocumentOut](InboxDocumentOut{})
	}
	return Ok[InboxDocumentOut](inboxDocumentToOut(*resp))
}

// InboxSendIn carries the multipart upload payload for the inbox.
// Stays declarative — multipart wiring lands in Plan D once
// EnableAttachments is fully threaded through both frontends. Name
// + Description are the optional upstream form fields; Filename and
// FilePath identify the local file to upload.
type InboxSendIn struct {
	Company     string `json:"company"`
	Name        string `json:"name,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Description string `json:"description,omitempty"`
	FilePath    string `json:"file_path"`
}

// InboxSendOut mirrors the upstream Created response — a Location URL
// pointing at the newly-uploaded inbox document.
type InboxSendOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (InboxSendOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (s InboxSendOut) TableRow() []string { return []string{s.Location} }

// InboxSend uploads a document to the inbox as multipart form data.
// Name surfaces in the inbox listing; Filename overrides the form-
// field filename (defaults to the basename of FilePath). Description
// is an optional free-text field.
func (c *Client) InboxSend(ctx context.Context, in InboxSendIn) Result[InboxSendOut] {
	if in.Company == "" {
		return Err[InboxSendOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpInboxSend,
		})
	}
	if in.FilePath == "" {
		return Err[InboxSendOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpInboxSend,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[InboxSendOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpInboxSend,
		})
	}
	defer func() { _ = closeFn() }()
	filename := in.Filename
	if filename == "" {
		filename = filepath.Base(in.FilePath)
	}
	name := in.Name
	if name == "" {
		name = filename
	}
	req := &fiken.CreateInboxDocumentReq{
		Name:     fiken.NewOptString(name),
		Filename: fiken.NewOptString(filename),
		File:     file,
	}
	if in.Description != "" {
		req.Description = fiken.NewOptString(in.Description)
	}
	resp, err := c.gen.CreateInboxDocument(ctx, req, fiken.CreateInboxDocumentParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[InboxSendOut](MapErr(OpInboxSend, err))
	}
	out := InboxSendOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[InboxSendOut](out)
}
