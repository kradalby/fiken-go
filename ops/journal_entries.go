package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// JournalEntriesListIn carries paged-list input for journal entries.
// Company is required because the endpoint is
// /companies/{slug}/journalEntries. The date/lastModified filters
// mirror the upstream knobs verbatim — exposed as Date strings so the
// CLI/MCP flag surface stays declarative.
type JournalEntriesListIn struct {
	Company        string `json:"company"`
	PageSize       int    `json:"page_size,omitempty"`
	Page           int    `json:"page,omitempty"`
	Date           Date   `json:"date,omitempty"`
	DateLe         Date   `json:"date_le,omitempty"`
	DateLt         Date   `json:"date_lt,omitempty"`
	DateGe         Date   `json:"date_ge,omitempty"`
	DateGt         Date   `json:"date_gt,omitempty"`
	LastModified   Date   `json:"last_modified,omitempty"`
	LastModifiedLe Date   `json:"last_modified_le,omitempty"`
	LastModifiedLt Date   `json:"last_modified_lt,omitempty"`
	LastModifiedGe Date   `json:"last_modified_ge,omitempty"`
	LastModifiedGt Date   `json:"last_modified_gt,omitempty"`
	CreatedDate    Date   `json:"created_date,omitempty"`
	CreatedDateLe  Date   `json:"created_date_le,omitempty"`
	CreatedDateLt  Date   `json:"created_date_lt,omitempty"`
	CreatedDateGe  Date   `json:"created_date_ge,omitempty"`
	CreatedDateGt  Date   `json:"created_date_gt,omitempty"`
}

// JournalEntryLineOut is the canonical journal-entry line shape.
// Amount stays int64 øre per the global money convention.
type JournalEntryLineOut struct {
	Amount           int64   `json:"amount"`
	Account          string  `json:"account,omitempty"`
	VatCode          string  `json:"vat_code,omitempty"`
	DebitAccount     string  `json:"debit_account,omitempty"`
	DebitVatCode     int64   `json:"debit_vat_code,omitempty"`
	CreditAccount    string  `json:"credit_account,omitempty"`
	CreditVatCode    int64   `json:"credit_vat_code,omitempty"`
	ProjectID        []int64 `json:"project_id,omitempty"`
	LastModifiedDate Date    `json:"last_modified_date,omitempty"`
}

// JournalEntryOut is the canonical single journal-entry shape exposed
// to CLI/MCP. TransactionID + OffsetTransactionID surface the
// transaction wiring; the line array carries the postings. Location
// surfaces the upstream Location header for Create responses.
type JournalEntryOut struct {
	JournalEntryID      int64                 `json:"journal_entry_id,omitempty"`
	JournalEntryNumber  int32                 `json:"journal_entry_number,omitempty"`
	Description         string                `json:"description"`
	TransactionDate     Date                  `json:"transaction_date,omitempty"`
	CreatedDate         Date                  `json:"created_date,omitempty"`
	LastModifiedDate    Date                  `json:"last_modified_date,omitempty"`
	TransactionID       int64                 `json:"transaction_id,omitempty"`
	OffsetTransactionID int64                 `json:"offset_transaction_id,omitempty"`
	Lines               []JournalEntryLineOut `json:"lines,omitempty"`
	Attachments         []AttachmentOut       `json:"attachments,omitempty"`
	Location            string                `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (j JournalEntryOut) TableHeader() []string {
	return []string{"ID", "DATE", "NUMBER", "DESCRIPTION"}
}

// TableRow implements output.tableRow.
func (j JournalEntryOut) TableRow() []string {
	return []string{
		strconv.FormatInt(j.JournalEntryID, 10),
		string(j.TransactionDate),
		strconv.FormatInt(int64(j.JournalEntryNumber), 10),
		j.Description,
	}
}

// JournalEntriesListOut is the paged response.
type JournalEntriesListOut = ListOut[JournalEntryOut]

// JournalEntriesList returns journal entries for the specified company.
func (c *Client) JournalEntriesList(ctx context.Context, in JournalEntriesListIn) Result[JournalEntriesListOut] {
	if in.Company == "" {
		return Err[JournalEntriesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpJournalEntriesList,
		})
	}
	params := fiken.GetJournalEntriesParams{CompanySlug: in.Company}
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
	setDateParam(&params.LastModified, in.LastModified)
	setDateParam(&params.LastModifiedLe, in.LastModifiedLe)
	setDateParam(&params.LastModifiedLt, in.LastModifiedLt)
	setDateParam(&params.LastModifiedGe, in.LastModifiedGe)
	setDateParam(&params.LastModifiedGt, in.LastModifiedGt)
	setDateParam(&params.CreatedDate, in.CreatedDate)
	setDateParam(&params.CreatedDateLe, in.CreatedDateLe)
	setDateParam(&params.CreatedDateLt, in.CreatedDateLt)
	setDateParam(&params.CreatedDateGe, in.CreatedDateGe)
	setDateParam(&params.CreatedDateGt, in.CreatedDateGt)
	resp, err := c.gen.GetJournalEntries(ctx, params)
	if err != nil {
		return Err[JournalEntriesListOut](MapErr(OpJournalEntriesList, err))
	}
	return Ok[JournalEntriesListOut](translateJournalEntriesList(resp))
}

// translateJournalEntriesList converts the ogen response into the
// canonical ListOut[JournalEntryOut] envelope, including paging meta.
func translateJournalEntriesList(resp *fiken.GetJournalEntriesOKHeaders) JournalEntriesListOut {
	if resp == nil {
		return JournalEntriesListOut{Items: []JournalEntryOut{}, Meta: ListMeta{}}
	}
	items := make([]JournalEntryOut, 0, len(resp.Response))
	for _, je := range resp.Response {
		items = append(items, journalEntryToOut(je))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return JournalEntriesListOut{Items: items, Meta: meta}
}

// journalEntryToOut maps fiken.JournalEntry into the canonical
// JournalEntryOut. transactionDate is the upstream `date` field (the
// posting date); CreatedDate / LastModifiedDate flow through verbatim.
func journalEntryToOut(j fiken.JournalEntry) JournalEntryOut {
	out := JournalEntryOut{
		JournalEntryID:      j.JournalEntryId.Or(0),
		JournalEntryNumber:  j.JournalEntryNumber.Or(0),
		Description:         j.Description,
		TransactionDate:     Date(j.Date.Format("2006-01-02")),
		TransactionID:       j.TransactionId.Or(0),
		OffsetTransactionID: j.OffsetTransactionId.Or(0),
	}
	if d, ok := j.CreatedDate.Get(); ok {
		out.CreatedDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := j.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(d.Format("2006-01-02"))
	}
	if len(j.Lines) > 0 {
		out.Lines = make([]JournalEntryLineOut, 0, len(j.Lines))
		for _, ln := range j.Lines {
			out.Lines = append(out.Lines, journalEntryLineToOut(ln))
		}
	}
	if len(j.Attachments) > 0 {
		out.Attachments = make([]AttachmentOut, 0, len(j.Attachments))
		for _, a := range j.Attachments {
			out.Attachments = append(out.Attachments, attachmentToOut(a))
		}
	}
	return out
}

// journalEntryLineToOut maps fiken.JournalEntryLine into the canonical
// JournalEntryLineOut.
func journalEntryLineToOut(ln fiken.JournalEntryLine) JournalEntryLineOut {
	out := JournalEntryLineOut{
		Amount:        ln.Amount,
		Account:       ln.Account.Or(""),
		VatCode:       ln.VatCode.Or(""),
		DebitAccount:  ln.DebitAccount.Or(""),
		DebitVatCode:  ln.DebitVatCode.Or(0),
		CreditAccount: ln.CreditAccount.Or(""),
		CreditVatCode: ln.CreditVatCode.Or(0),
		ProjectID:     ln.ProjectId,
	}
	if d, ok := ln.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(d.Format("2006-01-02"))
	}
	return out
}

// JournalEntriesGetIn requires company + journal entry id.
type JournalEntriesGetIn struct {
	Company        string `json:"company"`
	JournalEntryID int64  `json:"journal_entry_id"`
}

// JournalEntriesGet returns a single journal entry by id.
func (c *Client) JournalEntriesGet(ctx context.Context, in JournalEntriesGetIn) Result[JournalEntryOut] {
	if in.Company == "" {
		return Err[JournalEntryOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpJournalEntriesGet,
		})
	}
	if in.JournalEntryID == 0 {
		return Err[JournalEntryOut](&Error{
			Code: CodeValidation, Message: "journal_entry_id is required", Op: OpJournalEntriesGet,
		})
	}
	resp, err := c.gen.GetJournalEntry(ctx, fiken.GetJournalEntryParams{
		CompanySlug:    in.Company,
		JournalEntryId: in.JournalEntryID,
	})
	if err != nil {
		return Err[JournalEntryOut](MapErr(OpJournalEntriesGet, err))
	}
	if resp == nil {
		return Ok[JournalEntryOut](JournalEntryOut{})
	}
	return Ok[JournalEntryOut](journalEntryToOut(*resp))
}

// JournalEntriesCreateIn carries the create payload alongside the
// company. Body is the upstream fiken.GeneralJournalEntryRequest so
// the field surface stays in lock-step with the spec.
type JournalEntriesCreateIn struct {
	Company string                            `json:"company"`
	Body    *fiken.GeneralJournalEntryRequest `json:"body"`
}

// JournalEntriesCreate posts a new general journal entry. Fiken
// returns 201 with a Location header pointing at the new resource;
// surfaced via JournalEntryOut.Location.
func (c *Client) JournalEntriesCreate(ctx context.Context, in JournalEntriesCreateIn) Result[JournalEntryOut] {
	if in.Company == "" {
		return Err[JournalEntryOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpJournalEntriesCreate,
		})
	}
	if in.Body == nil {
		return Err[JournalEntryOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpJournalEntriesCreate,
		})
	}
	resp, err := c.gen.CreateGeneralJournalEntry(ctx, in.Body, fiken.CreateGeneralJournalEntryParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[JournalEntryOut](MapErr(OpJournalEntriesCreate, err))
	}
	out := JournalEntryOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[JournalEntryOut](out)
}

// AttachmentOut is the canonical attachment shape (re-used by other
// document tags down the line). Identifier is the user-facing tag
// Fiken displays alongside the file.
type AttachmentOut struct {
	Identifier  string `json:"identifier,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Type        string `json:"type,omitempty"`
}

// TableHeader implements output.tableRow.
func (a AttachmentOut) TableHeader() []string {
	return []string{"IDENTIFIER", "TYPE", "URL"}
}

// TableRow implements output.tableRow.
func (a AttachmentOut) TableRow() []string {
	return []string{a.Identifier, a.Type, a.DownloadURL}
}

// attachmentToOut maps fiken.Attachment into the canonical
// AttachmentOut. Type is rendered as its enum string ("invoice",
// "ocr", etc.).
func attachmentToOut(a fiken.Attachment) AttachmentOut {
	out := AttachmentOut{
		Identifier:  a.Identifier.Or(""),
		DownloadURL: a.DownloadUrl.Or(""),
		Comment:     a.Comment.Or(""),
	}
	if t, ok := a.Type.Get(); ok {
		out.Type = string(t)
	}
	return out
}

// AttachmentsListOut is the bare-array response for attachment lists
// (no pagination headers on these endpoints).
type AttachmentsListOut = ListOut[AttachmentOut]

// JournalEntriesAttachmentsListIn requires company + journal entry id.
type JournalEntriesAttachmentsListIn struct {
	Company        string `json:"company"`
	JournalEntryID int64  `json:"journal_entry_id"`
}

// JournalEntriesAttachmentsList returns all attachments for a journal
// entry. The upstream endpoint returns a bare array; Meta.Returned is
// the only field that gets set.
func (c *Client) JournalEntriesAttachmentsList(ctx context.Context, in JournalEntriesAttachmentsListIn) Result[AttachmentsListOut] {
	if in.Company == "" {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpJournalEntriesAttachmentsList,
		})
	}
	if in.JournalEntryID == 0 {
		return Err[AttachmentsListOut](&Error{
			Code: CodeValidation, Message: "journal_entry_id is required", Op: OpJournalEntriesAttachmentsList,
		})
	}
	resp, err := c.gen.GetJournalEntryAttachments(ctx, fiken.GetJournalEntryAttachmentsParams{
		CompanySlug:    in.Company,
		JournalEntryId: in.JournalEntryID,
	})
	if err != nil {
		return Err[AttachmentsListOut](MapErr(OpJournalEntriesAttachmentsList, err))
	}
	items := make([]AttachmentOut, 0, len(resp))
	for _, a := range resp {
		items = append(items, attachmentToOut(a))
	}
	return Ok[AttachmentsListOut](AttachmentsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// JournalEntriesAttachmentsAttachIn carries the multipart upload
// payload. Stays declarative for Plan C; the multipart wiring lands
// in Plan D once the EnableAttachments toggle is fully threaded.
type JournalEntriesAttachmentsAttachIn struct {
	Company        string `json:"company"`
	JournalEntryID int64  `json:"journal_entry_id"`
	Filename       string `json:"filename"`
	FilePath       string `json:"file_path"`
}

// JournalEntriesAttachmentsAttachOut mirrors the upstream Created
// response — a Location URL. Stays empty in the stub.
type JournalEntriesAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (JournalEntriesAttachmentsAttachOut) TableHeader() []string {
	return []string{"LOCATION"}
}

// TableRow implements output.tableRow.
func (a JournalEntriesAttachmentsAttachOut) TableRow() []string {
	return []string{a.Location}
}

// JournalEntriesAttachmentsAttach uploads a file to a journal entry
// as multipart form data. Filename defaults to the basename of
// FilePath; pass Filename to override.
func (c *Client) JournalEntriesAttachmentsAttach(ctx context.Context, in JournalEntriesAttachmentsAttachIn) Result[JournalEntriesAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[JournalEntriesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpJournalEntriesAttachmentsAttach,
		})
	}
	if in.JournalEntryID == 0 {
		return Err[JournalEntriesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "journal_entry_id is required", Op: OpJournalEntriesAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[JournalEntriesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpJournalEntriesAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[JournalEntriesAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpJournalEntriesAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	req := fiken.NewOptAddAttachmentToJournalEntryReq(fiken.AddAttachmentToJournalEntryReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	})
	resp, err := c.gen.AddAttachmentToJournalEntry(ctx, req, fiken.AddAttachmentToJournalEntryParams{
		CompanySlug:    in.Company,
		JournalEntryId: in.JournalEntryID,
	})
	if err != nil {
		return Err[JournalEntriesAttachmentsAttachOut](MapErr(OpJournalEntriesAttachmentsAttach, err))
	}
	out := JournalEntriesAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[JournalEntriesAttachmentsAttachOut](out)
}

// setDateParam copies a (possibly-empty) ops.Date into an ogen
// OptDate filter param. Empty Date is a no-op so the param stays
// unset and Fiken returns the unfiltered range.
func setDateParam(out *fiken.OptDate, in Date) {
	if in == "" {
		return
	}
	t, err := parseDate(string(in))
	if err != nil {
		return
	}
	out.SetTo(t)
}
