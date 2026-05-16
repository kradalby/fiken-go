package ops

import (
	"context"

	"github.com/kradalby/fiken-go/fiken"
)

// AccountsListIn carries paged-list input for the bookkeeping accounts
// of a company. Company is required because the endpoint is
// /companies/{slug}/accounts. FromAccount/ToAccount/Range mirror the
// upstream filter knobs; we expose them as-is so callers can scope
// large charts of accounts.
type AccountsListIn struct {
	Company     string `json:"company"`
	PageSize    int    `json:"page_size,omitempty"`
	Page        int    `json:"page,omitempty"`
	FromAccount int64  `json:"from_account,omitempty"`
	ToAccount   int64  `json:"to_account,omitempty"`
	Range       string `json:"range,omitempty"`
}

// AccountOut is the canonical single-account shape exposed to CLI/MCP.
// Account codes are opaque strings ("3020", "1500:10001"); we keep
// them as-is to preserve the reskontro variant.
type AccountOut struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

// TableHeader implements output.tableRow.
func (a AccountOut) TableHeader() []string {
	return []string{"CODE", "NAME"}
}

// TableRow implements output.tableRow.
func (a AccountOut) TableRow() []string {
	return []string{a.Code, a.Name}
}

// AccountsListOut is the paged response.
type AccountsListOut = ListOut[AccountOut]

// AccountsList returns bookkeeping accounts for the specified company.
func (c *Client) AccountsList(ctx context.Context, in AccountsListIn) Result[AccountsListOut] {
	if in.Company == "" {
		return Err[AccountsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpAccountsList,
		})
	}
	params := fiken.GetAccountsParams{CompanySlug: in.Company}
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
	if in.Range != "" {
		params.Range.SetTo(in.Range)
	}
	resp, err := c.gen.GetAccounts(ctx, params)
	if err != nil {
		return Err[AccountsListOut](MapErr(OpAccountsList, err))
	}
	return Ok[AccountsListOut](translateAccountsList(resp))
}

// translateAccountsList converts the ogen response into the canonical
// ListOut[AccountOut] envelope, including paging metadata.
func translateAccountsList(resp *fiken.GetAccountsOKHeaders) AccountsListOut {
	if resp == nil {
		return AccountsListOut{Items: []AccountOut{}, Meta: ListMeta{}}
	}
	items := make([]AccountOut, 0, len(resp.Response))
	for _, a := range resp.Response {
		items = append(items, accountToOut(a))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return AccountsListOut{Items: items, Meta: meta}
}

// accountToOut maps fiken.Account into the canonical AccountOut.
func accountToOut(a fiken.Account) AccountOut {
	return AccountOut{
		Code: a.Code.Or(""),
		Name: a.Name.Or(""),
	}
}

// AccountsGetIn requires company + account code.
type AccountsGetIn struct {
	Company     string `json:"company"`
	AccountCode string `json:"account_code"`
}

// AccountsGet returns a single bookkeeping account by code.
func (c *Client) AccountsGet(ctx context.Context, in AccountsGetIn) Result[AccountOut] {
	if in.Company == "" {
		return Err[AccountOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpAccountsGet,
		})
	}
	if in.AccountCode == "" {
		return Err[AccountOut](&Error{
			Code: CodeValidation, Message: "account_code is required", Op: OpAccountsGet,
		})
	}
	resp, err := c.gen.GetAccount(ctx, fiken.GetAccountParams{
		CompanySlug: in.Company,
		AccountCode: in.AccountCode,
	})
	if err != nil {
		return Err[AccountOut](MapErr(OpAccountsGet, err))
	}
	if resp == nil {
		return Ok[AccountOut](AccountOut{})
	}
	return Ok[AccountOut](accountToOut(*resp))
}
