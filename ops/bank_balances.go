package ops

import (
	"context"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// BankBalancesListIn carries paged-list input for bank balances as of
// an optional Date. The upstream endpoint exposes the Date as a query
// parameter; we forward it only when set so omitting --date returns
// the latest balance per bank account.
type BankBalancesListIn struct {
	Company  string `json:"company"`
	Date     Date   `json:"date,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// BankBalanceOut is the canonical single bank-balance shape exposed
// to CLI/MCP. Amount is int64 øre per the global money convention.
type BankBalanceOut struct {
	Source          string `json:"source,omitempty"`
	BankAccountID   int64  `json:"bank_account_id,omitempty"`
	BankAccountCode string `json:"bank_account_code,omitempty"`
	Date            Date   `json:"date,omitempty"`
	Amount          int64  `json:"amount,omitempty"`
}

// TableHeader implements output.tableRow.
func (BankBalanceOut) TableHeader() []string {
	return []string{"ID", "CODE", "DATE", "AMOUNT", "SOURCE"}
}

// TableRow implements output.tableRow.
func (b BankBalanceOut) TableRow() []string {
	return []string{
		strconv.FormatInt(b.BankAccountID, 10),
		b.BankAccountCode,
		string(b.Date),
		strconv.FormatInt(b.Amount, 10),
		b.Source,
	}
}

// BankBalancesListOut is the paged response.
type BankBalancesListOut = ListOut[BankBalanceOut]

// BankBalancesList returns the bank balances for the specified company.
func (c *Client) BankBalancesList(ctx context.Context, in BankBalancesListIn) Result[BankBalancesListOut] {
	if in.Company == "" {
		return Err[BankBalancesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpBankBalancesList,
		})
	}
	params := fiken.GetBankBalancesParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.Date != "" {
		d, err := parseDate(string(in.Date))
		if err != nil {
			return Err[BankBalancesListOut](&Error{
				Code: CodeValidation, Message: "date must be YYYY-MM-DD", Op: OpBankBalancesList,
			})
		}
		params.Date.SetTo(d)
	}
	resp, err := c.gen.GetBankBalances(ctx, params)
	if err != nil {
		return Err[BankBalancesListOut](MapErr(OpBankBalancesList, err))
	}
	return Ok[BankBalancesListOut](translateBankBalancesList(resp))
}

// translateBankBalancesList converts the ogen response into the
// canonical ListOut[BankBalanceOut] envelope including paging meta.
func translateBankBalancesList(resp *fiken.GetBankBalancesOKHeaders) BankBalancesListOut {
	if resp == nil {
		return BankBalancesListOut{Items: []BankBalanceOut{}, Meta: ListMeta{}}
	}
	items := make([]BankBalanceOut, 0, len(resp.Response))
	for _, b := range resp.Response {
		items = append(items, bankBalanceToOut(b))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return BankBalancesListOut{Items: items, Meta: meta}
}

// bankBalanceToOut maps fiken.BankBalanceResult into BankBalanceOut.
func bankBalanceToOut(b fiken.BankBalanceResult) BankBalanceOut {
	out := BankBalanceOut{
		Source:          b.Source.Or(""),
		BankAccountID:   b.BankAccountId.Or(0),
		BankAccountCode: b.BankAccountCode.Or(""),
		Amount:          b.Amount.Or(0),
	}
	if d, ok := b.Date.Get(); ok {
		out.Date = Date(d.Format("2006-01-02"))
	}
	return out
}
