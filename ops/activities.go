package ops

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// ActivitiesListIn carries paged-list input for activities. Company
// is required because the endpoint is /companies/{slug}/activities.
// Name + Archived mirror the upstream filter knobs; Archived is *bool
// to distinguish "unset" (return all) from "false" (return only
// non-archived).
type ActivitiesListIn struct {
	Company  string `json:"company"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Name     string `json:"name,omitempty"`
	Archived *bool  `json:"archived,omitempty"`
}

// ActivityOut is the canonical activity shape exposed to CLI/MCP.
// HourlyRate is int64 øre per the global money convention. Project
// and Product are flattened to their ids; callers needing the full
// shape can query the matching ops. Location surfaces the upstream
// Location header for Create/Update responses; list/get leave it zero.
type ActivityOut struct {
	ActivityID  int64  `json:"activity_id,omitempty"`
	Name        string `json:"name,omitempty"`
	HourlyRate  int64  `json:"hourly_rate,omitempty"`
	Description string `json:"description,omitempty"`
	Billable    bool   `json:"billable,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
	ProductID   int64  `json:"product_id,omitempty"`
	ProjectID   int64  `json:"project_id,omitempty"`
	Location    string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (ActivityOut) TableHeader() []string {
	return []string{"ID", "NAME", "HOURLY_RATE", "BILLABLE", "ARCHIVED"}
}

// TableRow implements output.tableRow.
func (a ActivityOut) TableRow() []string {
	return []string{
		strconv.FormatInt(a.ActivityID, 10),
		a.Name,
		strconv.FormatInt(a.HourlyRate, 10),
		strconv.FormatBool(a.Billable),
		strconv.FormatBool(a.Archived),
	}
}

// ActivitiesListOut is the paged response.
type ActivitiesListOut = ListOut[ActivityOut]

// ActivitiesList returns the activities for the specified company.
func (c *Client) ActivitiesList(ctx context.Context, in ActivitiesListIn) Result[ActivitiesListOut] {
	if in.Company == "" {
		return Err[ActivitiesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpActivitiesList,
		})
	}
	params := fiken.GetActivitiesParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.Name != "" {
		params.Name.SetTo(in.Name)
	}
	if in.Archived != nil {
		params.Archived.SetTo(*in.Archived)
	}
	resp, err := c.gen.GetActivities(ctx, params)
	if err != nil {
		return Err[ActivitiesListOut](MapErr(OpActivitiesList, err))
	}
	return Ok[ActivitiesListOut](translateActivitiesList(resp))
}

// translateActivitiesList converts the ogen response into the
// canonical ListOut[ActivityOut] envelope, including paging meta.
func translateActivitiesList(resp *fiken.GetActivitiesOKHeaders) ActivitiesListOut {
	if resp == nil {
		return ActivitiesListOut{Items: []ActivityOut{}, Meta: ListMeta{}}
	}
	items := make([]ActivityOut, 0, len(resp.Response))
	for _, a := range resp.Response {
		items = append(items, activityToOut(a))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return ActivitiesListOut{Items: items, Meta: meta}
}

// activityToOut maps fiken.ActivityResult into ActivityOut. Embedded
// Project / Product are flattened to their ids only — the full shapes
// are reachable via the dedicated ops, so we avoid bloating the canonical
// activity view.
func activityToOut(a fiken.ActivityResult) ActivityOut {
	out := ActivityOut{
		ActivityID:  a.ActivityId.Or(0),
		Name:        a.Name.Or(""),
		HourlyRate:  a.HourlyRate.Or(0),
		Description: a.Description.Or(""),
		Billable:    a.Billable.Or(false),
		Archived:    a.Archived.Or(false),
	}
	if p, ok := a.Product.Get(); ok {
		out.ProductID = p.ProductId.Or(0)
	}
	if p, ok := a.Project.Get(); ok {
		out.ProjectID = p.ProjectId.Or(0)
	}
	return out
}

// ActivitiesGetIn requires company + activity id.
type ActivitiesGetIn struct {
	Company    string `json:"company"`
	ActivityID int64  `json:"activity_id"`
}

// ActivitiesGet returns a single activity by id.
func (c *Client) ActivitiesGet(ctx context.Context, in ActivitiesGetIn) Result[ActivityOut] {
	if in.Company == "" {
		return Err[ActivityOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpActivitiesGet,
		})
	}
	if in.ActivityID == 0 {
		return Err[ActivityOut](&Error{
			Code: CodeValidation, Message: "activity_id is required", Op: OpActivitiesGet,
		})
	}
	resp, err := c.gen.GetActivity(ctx, fiken.GetActivityParams{
		CompanySlug: in.Company,
		ActivityId:  in.ActivityID,
	})
	if err != nil {
		return Err[ActivityOut](MapErr(OpActivitiesGet, err))
	}
	if resp == nil {
		return Ok[ActivityOut](ActivityOut{})
	}
	return Ok[ActivityOut](activityToOut(*resp))
}

// ActivitiesCreateIn carries the create payload alongside the company.
// Body is the upstream fiken.ActivityRequest so the field surface
// stays in lock-step with the spec.
type ActivitiesCreateIn struct {
	Company string                 `json:"company"`
	Body    *fiken.ActivityRequest `json:"body"`
}

// ActivitiesCreate posts a new activity. Upstream returns 201 with a
// Location header pointing at the new resource; surfaced via
// ActivityOut.Location. Time tracking is paid-tier on the real API so
// callers without billing will get 402 → CodePaymentRequired.
func (c *Client) ActivitiesCreate(ctx context.Context, in ActivitiesCreateIn) Result[ActivityOut] {
	if in.Company == "" {
		return Err[ActivityOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpActivitiesCreate,
		})
	}
	if in.Body == nil {
		return Err[ActivityOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpActivitiesCreate,
		})
	}
	resp, err := c.gen.CreateActivity(ctx, in.Body, fiken.CreateActivityParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[ActivityOut](MapErr(OpActivitiesCreate, err))
	}
	out := ActivityOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ActivityOut](out)
}

// ActivitiesUpdateIn carries the update payload + identifiers. Body
// uses fiken.UpdateActivityRequest (PATCH-shaped, optional fields)
// rather than ActivityRequest because the endpoint is PATCH.
type ActivitiesUpdateIn struct {
	Company    string                       `json:"company"`
	ActivityID int64                        `json:"activity_id"`
	Body       *fiken.UpdateActivityRequest `json:"body"`
}

// ActivitiesUpdate patches an existing activity. Upstream returns 200
// with a Location header pointing back at the resource — we surface
// the activity id alongside Location.
func (c *Client) ActivitiesUpdate(ctx context.Context, in ActivitiesUpdateIn) Result[ActivityOut] {
	if in.Company == "" {
		return Err[ActivityOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpActivitiesUpdate,
		})
	}
	if in.ActivityID == 0 {
		return Err[ActivityOut](&Error{
			Code: CodeValidation, Message: "activity_id is required", Op: OpActivitiesUpdate,
		})
	}
	if in.Body == nil {
		return Err[ActivityOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpActivitiesUpdate,
		})
	}
	resp, err := c.gen.UpdateActivity(ctx, in.Body, fiken.UpdateActivityParams{
		CompanySlug: in.Company,
		ActivityId:  in.ActivityID,
	})
	if err != nil {
		return Err[ActivityOut](MapErr(OpActivitiesUpdate, err))
	}
	out := ActivityOut{ActivityID: in.ActivityID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ActivityOut](out)
}

// ActivitiesDeleteIn identifies the activity to remove. Per upstream
// semantics, archived activities are kept (just marked archived); a
// 204 No Content covers both delete + archive.
type ActivitiesDeleteIn struct {
	Company    string `json:"company"`
	ActivityID int64  `json:"activity_id"`
}

// ActivitiesDeleteOut is intentionally empty — DELETE returns NoContent.
type ActivitiesDeleteOut struct{}

// TableHeader implements output.tableRow.
func (ActivitiesDeleteOut) TableHeader() []string { return []string{"STATUS"} }

// TableRow implements output.tableRow.
func (ActivitiesDeleteOut) TableRow() []string { return []string{"deleted"} }

// ActivitiesDelete removes (or archives) an activity. Upstream returns
// 204 NoContent — surface success as a zero-value ActivitiesDeleteOut.
func (c *Client) ActivitiesDelete(ctx context.Context, in ActivitiesDeleteIn) Result[ActivitiesDeleteOut] {
	if in.Company == "" {
		return Err[ActivitiesDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpActivitiesDelete,
		})
	}
	if in.ActivityID == 0 {
		return Err[ActivitiesDeleteOut](&Error{
			Code: CodeValidation, Message: "activity_id is required", Op: OpActivitiesDelete,
		})
	}
	if err := c.gen.DeleteActivity(ctx, fiken.DeleteActivityParams{
		CompanySlug: in.Company,
		ActivityId:  in.ActivityID,
	}); err != nil {
		return Err[ActivitiesDeleteOut](MapErr(OpActivitiesDelete, err))
	}
	return Ok[ActivitiesDeleteOut](ActivitiesDeleteOut{})
}
