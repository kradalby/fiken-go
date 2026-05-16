package ops

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// BankAccountsListIn carries paged-list input for a company's bank
// accounts. Company is required because the endpoint is
// /companies/{slug}/bankAccounts. Inactive mirrors the upstream
// filter — false = active accounts, true = inactive accounts.
type BankAccountsListIn struct {
	Company  string `json:"company"`
	PageSize int    `json:"page_size,omitempty"`
	Page     int    `json:"page,omitempty"`
	Inactive bool   `json:"inactive,omitempty"`
}

// BankAccountOut is the canonical single bank-account shape exposed
// to CLI/MCP. ReconciledBalance is int64 øre per the global money
// convention; BankAccountNumber stays a string because Fiken accepts
// non-numeric prefixes for foreign accounts. Location surfaces the
// upstream Location header for Create responses.
type BankAccountOut struct {
	BankAccountID     int64  `json:"bank_account_id,omitempty"`
	Name              string `json:"name,omitempty"`
	AccountCode       string `json:"account_code,omitempty"`
	BankAccountNumber string `json:"bank_account_number,omitempty"`
	Iban              string `json:"iban,omitempty"`
	Bic               string `json:"bic,omitempty"`
	ForeignService    string `json:"foreign_service,omitempty"`
	Type              string `json:"type,omitempty"`
	ReconciledBalance int64  `json:"reconciled_balance,omitempty"`
	ReconciledDate    Date   `json:"reconciled_date,omitempty"`
	Inactive          bool   `json:"inactive,omitempty"`
	Location          string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (b BankAccountOut) TableHeader() []string {
	return []string{"ID", "NAME", "ACCOUNT_CODE", "NUMBER", "TYPE"}
}

// TableRow implements output.tableRow.
func (b BankAccountOut) TableRow() []string {
	return []string{
		strconv.FormatInt(b.BankAccountID, 10),
		b.Name,
		b.AccountCode,
		b.BankAccountNumber,
		b.Type,
	}
}

// BankAccountsListOut is the paged response.
type BankAccountsListOut = ListOut[BankAccountOut]

// BankAccountsList returns bank accounts for the specified company.
func (c *Client) BankAccountsList(ctx context.Context, in BankAccountsListIn) Result[BankAccountsListOut] {
	if in.Company == "" {
		return Err[BankAccountsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpBankAccountsList,
		})
	}
	params := fiken.GetBankAccountsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.Inactive {
		params.Inactive.SetTo(true)
	}
	resp, err := c.gen.GetBankAccounts(ctx, params)
	if err != nil {
		return Err[BankAccountsListOut](MapErr(OpBankAccountsList, err))
	}
	return Ok[BankAccountsListOut](translateBankAccountsList(resp))
}

// translateBankAccountsList converts the ogen response into the
// canonical ListOut[BankAccountOut] envelope including paging meta.
func translateBankAccountsList(resp *fiken.GetBankAccountsOKHeaders) BankAccountsListOut {
	if resp == nil {
		return BankAccountsListOut{Items: []BankAccountOut{}, Meta: ListMeta{}}
	}
	items := make([]BankAccountOut, 0, len(resp.Response))
	for _, b := range resp.Response {
		items = append(items, bankAccountToOut(b))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return BankAccountsListOut{Items: items, Meta: meta}
}

// bankAccountToOut maps fiken.BankAccountResult to the canonical
// BankAccountOut. ReconciledDate is rendered as YYYY-MM-DD; an unset
// upstream date stays an empty Date, omitted from JSON.
func bankAccountToOut(b fiken.BankAccountResult) BankAccountOut {
	out := BankAccountOut{
		BankAccountID:     b.BankAccountId.Or(0),
		Name:              b.Name.Or(""),
		AccountCode:       b.AccountCode.Or(""),
		BankAccountNumber: b.BankAccountNumber.Or(""),
		Iban:              b.Iban.Or(""),
		Bic:               b.Bic.Or(""),
		ForeignService:    b.ForeignService.Or(""),
		ReconciledBalance: b.ReconciledBalance.Or(0),
		Inactive:          b.Inactive.Or(false),
	}
	if t, ok := b.Type.Get(); ok {
		out.Type = string(t)
	}
	if d, ok := b.ReconciledDate.Get(); ok {
		out.ReconciledDate = Date(d.Format("2006-01-02"))
	}
	return out
}

// BankAccountsGetIn requires company + bank account id.
type BankAccountsGetIn struct {
	Company       string `json:"company"`
	BankAccountID int64  `json:"bank_account_id"`
}

// BankAccountsGet returns a single bank account by id.
func (c *Client) BankAccountsGet(ctx context.Context, in BankAccountsGetIn) Result[BankAccountOut] {
	if in.Company == "" {
		return Err[BankAccountOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpBankAccountsGet,
		})
	}
	if in.BankAccountID == 0 {
		return Err[BankAccountOut](&Error{
			Code: CodeValidation, Message: "bank_account_id is required", Op: OpBankAccountsGet,
		})
	}
	resp, err := c.gen.GetBankAccount(ctx, fiken.GetBankAccountParams{
		CompanySlug:   in.Company,
		BankAccountId: in.BankAccountID,
	})
	if err != nil {
		return Err[BankAccountOut](MapErr(OpBankAccountsGet, err))
	}
	if resp == nil {
		return Ok[BankAccountOut](BankAccountOut{})
	}
	return Ok[BankAccountOut](bankAccountToOut(*resp))
}

// BankAccountsCreateIn carries the create payload alongside the
// company. Body is the upstream fiken.BankAccountRequest shape so the
// field surface stays in lock-step with the spec.
type BankAccountsCreateIn struct {
	Company string                    `json:"company"`
	Body    *fiken.BankAccountRequest `json:"body"`
}

// BankAccountsCreate posts a new bank account. Fiken returns 201 with a
// Location header pointing at the new resource; we expose it through
// BankAccountOut.Location.
func (c *Client) BankAccountsCreate(ctx context.Context, in BankAccountsCreateIn) Result[BankAccountOut] {
	if in.Company == "" {
		return Err[BankAccountOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpBankAccountsCreate,
		})
	}
	if in.Body == nil {
		return Err[BankAccountOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpBankAccountsCreate,
		})
	}
	resp, err := c.gen.CreateBankAccount(ctx, in.Body, fiken.CreateBankAccountParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[BankAccountOut](MapErr(OpBankAccountsCreate, err))
	}
	out := BankAccountOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[BankAccountOut](out)
}
