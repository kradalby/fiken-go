package ops

import (
	"context"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// AccountBalancesListIn carries paged-list input for account balances
// as of a given Date. Company + Date are required; both are required
// by the upstream /accountBalances endpoint (date is a query string
// in YYYY-MM-DD form). FromAccount / ToAccount mirror the upstream
// numeric range filter.
type AccountBalancesListIn struct {
	Company     string `json:"company"`
	Date        Date   `json:"date"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
	FromAccount int64  `json:"from_account,omitempty"`
	ToAccount   int64  `json:"to_account,omitempty"`
}

// AccountBalanceOut is the canonical single-balance shape exposed to
// CLI/MCP. Balance is int64 øre per the global money convention. Code
// stays a string because the upstream value can be "3020" or
// "1500:10001" (the reskontro variant).
type AccountBalanceOut struct {
	Code    string `json:"code,omitempty"`
	Name    string `json:"name,omitempty"`
	Balance int64  `json:"balance,omitempty"`
}

// TableHeader implements output.tableRow.
func (AccountBalanceOut) TableHeader() []string {
	return []string{"CODE", "NAME", "BALANCE"}
}

// TableRow implements output.tableRow.
func (a AccountBalanceOut) TableRow() []string {
	return []string{a.Code, a.Name, strconv.FormatInt(a.Balance, 10)}
}

// AccountBalancesListOut is the paged response.
type AccountBalancesListOut = ListOut[AccountBalanceOut]

// AccountBalancesList returns the closing balances for the bookkeeping
// accounts of the specified company as of `date`.
func (c *Client) AccountBalancesList(ctx context.Context, in AccountBalancesListIn) Result[AccountBalancesListOut] {
	if in.Company == "" {
		return Err[AccountBalancesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpAccountBalancesList,
		})
	}
	if in.Date == "" {
		return Err[AccountBalancesListOut](&Error{
			Code: CodeValidation, Message: "date is required (YYYY-MM-DD)", Op: OpAccountBalancesList,
		})
	}
	d, err := parseDate(string(in.Date))
	if err != nil {
		return Err[AccountBalancesListOut](&Error{
			Code: CodeValidation, Message: "date must be YYYY-MM-DD", Op: OpAccountBalancesList,
		})
	}
	params := fiken.GetAccountBalancesParams{
		CompanySlug: in.Company,
		Date:        d,
	}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.FromAccount > 0 {
		params.FromAccount.SetTo(in.FromAccount)
	}
	if in.ToAccount > 0 {
		params.ToAccount.SetTo(in.ToAccount)
	}
	resp, err := c.gen.GetAccountBalances(ctx, params)
	if err != nil {
		return Err[AccountBalancesListOut](MapErr(OpAccountBalancesList, err))
	}
	return Ok[AccountBalancesListOut](translateAccountBalancesList(resp))
}

// translateAccountBalancesList converts the ogen response into the
// canonical ListOut[AccountBalanceOut] envelope including paging meta.
func translateAccountBalancesList(resp *fiken.GetAccountBalancesOKHeaders) AccountBalancesListOut {
	if resp == nil {
		return AccountBalancesListOut{Items: []AccountBalanceOut{}, Meta: ListMeta{}}
	}
	items := make([]AccountBalanceOut, 0, len(resp.Response))
	for _, a := range resp.Response {
		items = append(items, accountBalanceToOut(a))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return AccountBalancesListOut{Items: items, Meta: meta}
}

// accountBalanceToOut maps fiken.AccountBalance into the canonical
// AccountBalanceOut.
func accountBalanceToOut(a fiken.AccountBalance) AccountBalanceOut {
	return AccountBalanceOut{
		Code:    a.Code.Or(""),
		Name:    a.Name.Or(""),
		Balance: a.Balance.Or(0),
	}
}

// AccountBalancesGetIn requires company + account code + date.
type AccountBalancesGetIn struct {
	Company     string `json:"company"`
	AccountCode string `json:"account_code"`
	Date        Date   `json:"date"`
}

// AccountBalancesGet returns a single account balance as of `date`.
func (c *Client) AccountBalancesGet(ctx context.Context, in AccountBalancesGetIn) Result[AccountBalanceOut] {
	if in.Company == "" {
		return Err[AccountBalanceOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpAccountBalancesGet,
		})
	}
	if in.AccountCode == "" {
		return Err[AccountBalanceOut](&Error{
			Code: CodeValidation, Message: "account_code is required", Op: OpAccountBalancesGet,
		})
	}
	if in.Date == "" {
		return Err[AccountBalanceOut](&Error{
			Code: CodeValidation, Message: "date is required (YYYY-MM-DD)", Op: OpAccountBalancesGet,
		})
	}
	d, err := parseDate(string(in.Date))
	if err != nil {
		return Err[AccountBalanceOut](&Error{
			Code: CodeValidation, Message: "date must be YYYY-MM-DD", Op: OpAccountBalancesGet,
		})
	}
	resp, err := c.gen.GetAccountBalance(ctx, fiken.GetAccountBalanceParams{
		CompanySlug: in.Company,
		AccountCode: in.AccountCode,
		Date:        d,
	})
	if err != nil {
		return Err[AccountBalanceOut](MapErr(OpAccountBalancesGet, err))
	}
	if resp == nil {
		return Ok[AccountBalanceOut](AccountBalanceOut{})
	}
	return Ok[AccountBalanceOut](accountBalanceToOut(*resp))
}
