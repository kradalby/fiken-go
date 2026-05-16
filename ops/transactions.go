package ops

import (
	"context"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// TransactionsListIn carries paged-list input for transactions. The
// upstream endpoint exposes only last-modified + created-date filters
// (no posting-date filter), so this struct stays narrower than
// JournalEntriesListIn.
type TransactionsListIn struct {
	Company        string `json:"company"`
	PageSize       int    `json:"page_size,omitempty"`
	Page           int    `json:"page,omitempty"`
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

// TransactionOut is the canonical transaction shape. A transaction is
// a posted ledger event; Entries surfaces the underlying journal
// entries via the existing JournalEntryOut shape so line amounts stay
// int64 øre throughout.
type TransactionOut struct {
	TransactionID    int64             `json:"transaction_id,omitempty"`
	Description      string            `json:"description,omitempty"`
	Type             string            `json:"type,omitempty"`
	CreatedDate      Date              `json:"created_date,omitempty"`
	LastModifiedDate Date              `json:"last_modified_date,omitempty"`
	Deleted          bool              `json:"deleted,omitempty"`
	Entries          []JournalEntryOut `json:"entries,omitempty"`
}

// TableHeader implements output.tableRow.
func (TransactionOut) TableHeader() []string {
	return []string{"ID", "TYPE", "DESCRIPTION", "ENTRIES"}
}

// TableRow implements output.tableRow.
func (t TransactionOut) TableRow() []string {
	return []string{
		strconv.FormatInt(t.TransactionID, 10),
		t.Type,
		t.Description,
		strconv.Itoa(len(t.Entries)),
	}
}

// TransactionsListOut is the paged response.
type TransactionsListOut = ListOut[TransactionOut]

// TransactionsList returns transactions for the specified company.
func (c *Client) TransactionsList(ctx context.Context, in TransactionsListIn) Result[TransactionsListOut] {
	if in.Company == "" {
		return Err[TransactionsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTransactionsList,
		})
	}
	params := fiken.GetTransactionsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
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
	resp, err := c.gen.GetTransactions(ctx, params)
	if err != nil {
		return Err[TransactionsListOut](MapErr(OpTransactionsList, err))
	}
	return Ok[TransactionsListOut](translateTransactionsList(resp))
}

// translateTransactionsList converts the ogen response into the
// canonical ListOut[TransactionOut] envelope, including paging meta.
func translateTransactionsList(resp *fiken.GetTransactionsOKHeaders) TransactionsListOut {
	if resp == nil {
		return TransactionsListOut{Items: []TransactionOut{}, Meta: ListMeta{}}
	}
	items := make([]TransactionOut, 0, len(resp.Response))
	for _, tx := range resp.Response {
		items = append(items, transactionToOut(tx))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return TransactionsListOut{Items: items, Meta: meta}
}

// transactionToOut maps fiken.Transaction into the canonical
// TransactionOut. The nested Entries reuse the journalEntryToOut
// translator so line amounts and dates stay in lock-step with the
// journal_entries tag.
func transactionToOut(t fiken.Transaction) TransactionOut {
	out := TransactionOut{
		TransactionID: t.TransactionId.Or(0),
		Description:   t.Description.Or(""),
		Type:          t.Type.Or(""),
		Deleted:       t.Deleted.Or(false),
	}
	if d, ok := t.CreatedDate.Get(); ok {
		out.CreatedDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := t.LastModifiedDate.Get(); ok {
		out.LastModifiedDate = Date(d.Format("2006-01-02"))
	}
	if len(t.Entries) > 0 {
		out.Entries = make([]JournalEntryOut, 0, len(t.Entries))
		for _, je := range t.Entries {
			out.Entries = append(out.Entries, journalEntryToOut(je))
		}
	}
	return out
}

// TransactionsGetIn requires company + transaction id.
type TransactionsGetIn struct {
	Company       string `json:"company"`
	TransactionID int64  `json:"transaction_id"`
}

// TransactionsGet returns a single transaction by id.
func (c *Client) TransactionsGet(ctx context.Context, in TransactionsGetIn) Result[TransactionOut] {
	if in.Company == "" {
		return Err[TransactionOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTransactionsGet,
		})
	}
	if in.TransactionID == 0 {
		return Err[TransactionOut](&Error{
			Code: CodeValidation, Message: "transaction_id is required", Op: OpTransactionsGet,
		})
	}
	resp, err := c.gen.GetTransaction(ctx, fiken.GetTransactionParams{
		CompanySlug:   in.Company,
		TransactionId: in.TransactionID,
	})
	if err != nil {
		return Err[TransactionOut](MapErr(OpTransactionsGet, err))
	}
	if resp == nil {
		return Ok[TransactionOut](TransactionOut{})
	}
	return Ok[TransactionOut](transactionToOut(*resp))
}

// TransactionsDeleteIn identifies the transaction to mark deleted.
// Description is a required query parameter upstream (audit trail).
type TransactionsDeleteIn struct {
	Company       string `json:"company"`
	TransactionID int64  `json:"transaction_id"`
	Description   string `json:"description"`
}

// TransactionsDeleteOut surfaces the post-delete transaction snapshot.
type TransactionsDeleteOut = TransactionOut

// TransactionsDelete soft-archives a transaction. Despite the verb,
// the upstream route is a PATCH that takes a required description
// query param (audit trail) and returns the post-delete transaction
// snapshot.
func (c *Client) TransactionsDelete(ctx context.Context, in TransactionsDeleteIn) Result[TransactionsDeleteOut] {
	if in.Company == "" {
		return Err[TransactionsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTransactionsDelete,
		})
	}
	if in.TransactionID == 0 {
		return Err[TransactionsDeleteOut](&Error{
			Code: CodeValidation, Message: "transaction_id is required", Op: OpTransactionsDelete,
		})
	}
	if in.Description == "" {
		return Err[TransactionsDeleteOut](&Error{
			Code: CodeValidation, Message: "description is required", Op: OpTransactionsDelete,
		})
	}
	resp, err := c.gen.DeleteTransaction(ctx, fiken.DeleteTransactionParams{
		CompanySlug:   in.Company,
		TransactionId: in.TransactionID,
		Description:   in.Description,
	})
	if err != nil {
		return Err[TransactionsDeleteOut](MapErr(OpTransactionsDelete, err))
	}
	if resp == nil {
		return Ok[TransactionsDeleteOut](TransactionsDeleteOut{})
	}
	return Ok[TransactionsDeleteOut](transactionToOut(*resp))
}
