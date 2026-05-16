package ops

import (
	"context"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// TimeUsersListIn carries paged-list input for the persons who can
// register time entries on the company. Name + Email mirror the
// upstream filter knobs.
type TimeUsersListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

// TimeUserOut is the canonical time-user shape exposed to CLI/MCP.
type TimeUserOut struct {
	TimeUserID int64  `json:"time_user_id,omitempty"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
}

// TableHeader implements output.tableRow.
func (TimeUserOut) TableHeader() []string {
	return []string{"ID", "NAME", "EMAIL"}
}

// TableRow implements output.tableRow.
func (t TimeUserOut) TableRow() []string {
	return []string{strconv.FormatInt(t.TimeUserID, 10), t.Name, t.Email}
}

// TimeUsersListOut is the paged response.
type TimeUsersListOut = ListOut[TimeUserOut]

// TimeUsersList returns the time users for the specified company.
func (c *Client) TimeUsersList(ctx context.Context, in TimeUsersListIn) Result[TimeUsersListOut] {
	if in.Company == "" {
		return Err[TimeUsersListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeUsersList,
		})
	}
	params := fiken.GetTimeUsersParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.Name != "" {
		params.Name.SetTo(in.Name)
	}
	if in.Email != "" {
		params.Email.SetTo(in.Email)
	}
	resp, err := c.gen.GetTimeUsers(ctx, params)
	if err != nil {
		return Err[TimeUsersListOut](MapErr(OpTimeUsersList, err))
	}
	return Ok[TimeUsersListOut](translateTimeUsersList(resp))
}

// translateTimeUsersList converts the ogen response into the canonical
// ListOut[TimeUserOut] envelope, including paging meta.
func translateTimeUsersList(resp *fiken.GetTimeUsersOKHeaders) TimeUsersListOut {
	if resp == nil {
		return TimeUsersListOut{Items: []TimeUserOut{}, Meta: ListMeta{}}
	}
	items := make([]TimeUserOut, 0, len(resp.Response))
	for _, t := range resp.Response {
		items = append(items, timeUserToOut(t))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return TimeUsersListOut{Items: items, Meta: meta}
}

// timeUserToOut maps fiken.TimeUserResult into TimeUserOut.
func timeUserToOut(t fiken.TimeUserResult) TimeUserOut {
	return TimeUserOut{
		TimeUserID: t.TimeUserId.Or(0),
		Name:       t.Name.Or(""),
		Email:      t.Email.Or(""),
	}
}

// TimeUsersGetIn requires company + time user id.
type TimeUsersGetIn struct {
	Company    string `json:"company"`
	TimeUserID int64  `json:"time_user_id"`
}

// TimeUsersGet returns a single time user by id.
func (c *Client) TimeUsersGet(ctx context.Context, in TimeUsersGetIn) Result[TimeUserOut] {
	if in.Company == "" {
		return Err[TimeUserOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeUsersGet,
		})
	}
	if in.TimeUserID == 0 {
		return Err[TimeUserOut](&Error{
			Code: CodeValidation, Message: "time_user_id is required", Op: OpTimeUsersGet,
		})
	}
	resp, err := c.gen.GetTimeUser(ctx, fiken.GetTimeUserParams{
		CompanySlug: in.Company,
		TimeUserId:  in.TimeUserID,
	})
	if err != nil {
		return Err[TimeUserOut](MapErr(OpTimeUsersGet, err))
	}
	if resp == nil {
		return Ok[TimeUserOut](TimeUserOut{})
	}
	return Ok[TimeUserOut](timeUserToOut(*resp))
}
