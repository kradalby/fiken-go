package ops

import (
	"context"

	"github.com/kradalby/fiken-go/fiken"
)

// GroupsListIn carries paged-list input for customer groups. Company
// is required. Groups are returned as a flat list of names — no other
// fields per the upstream spec.
type GroupsListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// GroupOut wraps a customer-group name. The upstream payload is just
// a string; we wrap it so the canonical CLI/MCP output stays a
// keyed object instead of a bare scalar.
type GroupOut struct {
	Name string `json:"name,omitempty"`
}

// TableHeader implements output.tableRow.
func (GroupOut) TableHeader() []string { return []string{"NAME"} }

// TableRow implements output.tableRow.
func (g GroupOut) TableRow() []string { return []string{g.Name} }

// GroupsListOut is the paged response.
type GroupsListOut = ListOut[GroupOut]

// GroupsList returns the customer groups for the specified company.
func (c *Client) GroupsList(ctx context.Context, in GroupsListIn) Result[GroupsListOut] {
	if in.Company == "" {
		return Err[GroupsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpGroupsList,
		})
	}
	params := fiken.GetGroupsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetGroups(ctx, params)
	if err != nil {
		return Err[GroupsListOut](MapErr(OpGroupsList, err))
	}
	return Ok[GroupsListOut](translateGroupsList(resp))
}

// translateGroupsList converts the ogen response into the canonical
// ListOut[GroupOut] envelope, including paging meta.
func translateGroupsList(resp *fiken.GetGroupsOKHeaders) GroupsListOut {
	if resp == nil {
		return GroupsListOut{Items: []GroupOut{}, Meta: ListMeta{}}
	}
	items := make([]GroupOut, 0, len(resp.Response))
	for _, name := range resp.Response {
		items = append(items, GroupOut{Name: name})
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return GroupsListOut{Items: items, Meta: meta}
}
