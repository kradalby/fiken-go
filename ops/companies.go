package ops

import (
	"context"

	"github.com/kradalby/fiken-go/fiken"
)

// CompaniesListIn carries paged-list input.
type CompaniesListIn struct {
	PageSize int `json:"page_size,omitempty"`
	Page     int `json:"page,omitempty"`
}

// CompanyOut is the canonical single-company shape.
type CompanyOut struct {
	Slug               string `json:"slug"`
	Name               string `json:"name"`
	OrganizationNumber string `json:"organization_number,omitempty"`
}

// TableHeader implements output.tableRow.
func (c CompanyOut) TableHeader() []string {
	return []string{"SLUG", "NAME", "ORG.NR"}
}

// TableRow implements output.tableRow.
func (c CompanyOut) TableRow() []string {
	return []string{c.Slug, c.Name, c.OrganizationNumber}
}

// CompaniesListOut is the paged response.
type CompaniesListOut = ListOut[CompanyOut]

// CompaniesList returns all companies the authenticated user has
// access to.
func (c *Client) CompaniesList(ctx context.Context, in CompaniesListIn) Result[CompaniesListOut] {
	params := fiken.GetCompaniesParams{}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetCompanies(ctx, params)
	if err != nil {
		return Err[CompaniesListOut](MapErr(OpCompaniesList, err))
	}
	return Ok[CompaniesListOut](translateCompaniesList(resp))
}

// translateCompaniesList converts the ogen response (with header
// metadata) into our canonical ListOut[CompanyOut].
func translateCompaniesList(resp *fiken.GetCompaniesOKHeaders) CompaniesListOut {
	if resp == nil {
		return CompaniesListOut{Items: []CompanyOut{}, Meta: ListMeta{}}
	}
	items := make([]CompanyOut, 0, len(resp.Response))
	for _, co := range resp.Response {
		items = append(items, companyToOut(co))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return CompaniesListOut{Items: items, Meta: meta}
}

// companyToOut maps fiken.Company into CompanyOut, unwrapping the
// optional string wrappers ogen emits.
func companyToOut(co fiken.Company) CompanyOut {
	return CompanyOut{
		Slug:               co.Slug.Or(""),
		Name:               co.Name.Or(""),
		OrganizationNumber: co.OrganizationNumber.Or(""),
	}
}

// CompaniesGetIn requires a company slug.
type CompaniesGetIn struct {
	Company string `json:"company"`
}

// CompaniesGet returns a single company by slug.
func (c *Client) CompaniesGet(ctx context.Context, in CompaniesGetIn) Result[CompanyOut] {
	if in.Company == "" {
		return Err[CompanyOut](&Error{
			Code:    CodeValidation,
			Message: "company slug is required",
			Op:      OpCompaniesGet,
		})
	}
	resp, err := c.gen.GetCompany(ctx, fiken.GetCompanyParams{CompanySlug: in.Company})
	if err != nil {
		return Err[CompanyOut](MapErr(OpCompaniesGet, err))
	}
	return Ok[CompanyOut](translateCompanyGet(resp))
}

func translateCompanyGet(resp *fiken.Company) CompanyOut {
	if resp == nil {
		return CompanyOut{}
	}
	return companyToOut(*resp)
}
