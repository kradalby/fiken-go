package ops

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// ProjectsListIn carries paged-list input for projects. The upstream
// endpoint exposes completed / name / number filters in addition to
// the standard paging — surface them so callers can narrow without
// post-filtering. Completed is *bool to distinguish "unset" (return
// everything) from "false" (return only open projects).
type ProjectsListIn struct {
	Company   string `json:"company"`
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
	Completed *bool  `json:"completed,omitempty"`
	Name      string `json:"name,omitempty"`
	Number    string `json:"number,omitempty"`
}

// ProjectOut is the canonical project shape exposed to CLI/MCP. We
// flatten Contact.contactId into a scalar ContactID since the upstream
// payload is the entire Contact object — callers that need more than
// the id can hit ContactsGet. Dates round-trip as YYYY-MM-DD strings.
// Location surfaces the Location header for Create/Update responses;
// list/get leave it zero.
type ProjectOut struct {
	ProjectID   int64  `json:"project_id,omitempty"`
	Number      string `json:"number,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	StartDate   Date   `json:"start_date,omitempty"`
	EndDate     Date   `json:"end_date,omitempty"`
	ContactID   int64  `json:"contact_id,omitempty"`
	ContactName string `json:"contact_name,omitempty"`
	Completed   bool   `json:"completed"`
	Location    string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (ProjectOut) TableHeader() []string {
	return []string{"ID", "NUMBER", "NAME", "COMPLETED"}
}

// TableRow implements output.tableRow.
func (p ProjectOut) TableRow() []string {
	return []string{
		strconv.FormatInt(p.ProjectID, 10),
		p.Number,
		p.Name,
		strconv.FormatBool(p.Completed),
	}
}

// ProjectsListOut is the paged response.
type ProjectsListOut = ListOut[ProjectOut]

// ProjectsList returns projects for the specified company.
func (c *Client) ProjectsList(ctx context.Context, in ProjectsListIn) Result[ProjectsListOut] {
	if in.Company == "" {
		return Err[ProjectsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProjectsList,
		})
	}
	params := fiken.GetProjectsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.Completed != nil {
		params.Completed.SetTo(*in.Completed)
	}
	if in.Name != "" {
		params.Name.SetTo(in.Name)
	}
	if in.Number != "" {
		params.Number.SetTo(in.Number)
	}
	resp, err := c.gen.GetProjects(ctx, params)
	if err != nil {
		return Err[ProjectsListOut](MapErr(OpProjectsList, err))
	}
	return Ok[ProjectsListOut](translateProjectsList(resp))
}

// translateProjectsList converts the ogen response into the canonical
// ListOut[ProjectOut] envelope, including paging meta.
func translateProjectsList(resp *fiken.GetProjectsOKHeaders) ProjectsListOut {
	if resp == nil {
		return ProjectsListOut{Items: []ProjectOut{}, Meta: ListMeta{}}
	}
	items := make([]ProjectOut, 0, len(resp.Response))
	for _, p := range resp.Response {
		items = append(items, projectToOut(p))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return ProjectsListOut{Items: items, Meta: meta}
}

// projectToOut maps fiken.ProjectResult into the canonical ProjectOut.
// Contact is flattened to id + name since the full Contact carries a
// lot of fields a project view rarely needs. Dates are rendered as
// YYYY-MM-DD strings via the shared Date type.
func projectToOut(p fiken.ProjectResult) ProjectOut {
	out := ProjectOut{
		ProjectID:   p.ProjectId.Or(0),
		Number:      p.Number.Or(""),
		Name:        p.Name.Or(""),
		Description: p.Description.Or(""),
		Completed:   p.Completed.Or(false),
	}
	if d, ok := p.StartDate.Get(); ok {
		out.StartDate = Date(d.Format("2006-01-02"))
	}
	if d, ok := p.EndDate.Get(); ok {
		out.EndDate = Date(d.Format("2006-01-02"))
	}
	if co, ok := p.Contact.Get(); ok {
		out.ContactID = co.ContactId.Or(0)
		out.ContactName = co.Name
	}
	return out
}

// ProjectsGetIn requires company + project id.
type ProjectsGetIn struct {
	Company   string `json:"company"`
	ProjectID int64  `json:"project_id"`
}

// ProjectsGet returns a single project by id.
func (c *Client) ProjectsGet(ctx context.Context, in ProjectsGetIn) Result[ProjectOut] {
	if in.Company == "" {
		return Err[ProjectOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProjectsGet,
		})
	}
	if in.ProjectID == 0 {
		return Err[ProjectOut](&Error{
			Code: CodeValidation, Message: "project_id is required", Op: OpProjectsGet,
		})
	}
	resp, err := c.gen.GetProject(ctx, fiken.GetProjectParams{
		CompanySlug: in.Company,
		ProjectId:   in.ProjectID,
	})
	if err != nil {
		return Err[ProjectOut](MapErr(OpProjectsGet, err))
	}
	if resp == nil {
		return Ok[ProjectOut](ProjectOut{})
	}
	return Ok[ProjectOut](projectToOut(*resp))
}

// ProjectsCreateIn carries the create payload alongside the company.
// Body is the upstream fiken.ProjectRequest shape so the field surface
// stays in lock-step with the spec without us hand-curating every
// column.
type ProjectsCreateIn struct {
	Company string                `json:"company"`
	Body    *fiken.ProjectRequest `json:"body"`
}

// ProjectsCreate posts a new project. Upstream returns 201 with a
// Location header pointing at the new resource; surfaced via
// ProjectOut.Location. The endpoint is paid-tier on the real API so
// callers without billing will get 402 → CodePaymentRequired.
func (c *Client) ProjectsCreate(ctx context.Context, in ProjectsCreateIn) Result[ProjectOut] {
	if in.Company == "" {
		return Err[ProjectOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProjectsCreate,
		})
	}
	if in.Body == nil {
		return Err[ProjectOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpProjectsCreate,
		})
	}
	resp, err := c.gen.CreateProject(ctx, in.Body, fiken.CreateProjectParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[ProjectOut](MapErr(OpProjectsCreate, err))
	}
	out := ProjectOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ProjectOut](out)
}

// ProjectsUpdateIn carries the update payload + identifiers. Body uses
// the upstream UpdateProjectRequest (PATCH-shaped, optional fields)
// rather than the create request because the endpoint is PATCH.
type ProjectsUpdateIn struct {
	Company   string                      `json:"company"`
	ProjectID int64                       `json:"project_id"`
	Body      *fiken.UpdateProjectRequest `json:"body"`
}

// ProjectsUpdate patches an existing project. Upstream returns 201
// with a Location header pointing back at the resource — we surface
// the project id alongside Location.
func (c *Client) ProjectsUpdate(ctx context.Context, in ProjectsUpdateIn) Result[ProjectOut] {
	if in.Company == "" {
		return Err[ProjectOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProjectsUpdate,
		})
	}
	if in.ProjectID == 0 {
		return Err[ProjectOut](&Error{
			Code: CodeValidation, Message: "project_id is required", Op: OpProjectsUpdate,
		})
	}
	if in.Body == nil {
		return Err[ProjectOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpProjectsUpdate,
		})
	}
	resp, err := c.gen.UpdateProject(ctx, in.Body, fiken.UpdateProjectParams{
		CompanySlug: in.Company,
		ProjectId:   in.ProjectID,
	})
	if err != nil {
		return Err[ProjectOut](MapErr(OpProjectsUpdate, err))
	}
	out := ProjectOut{ProjectID: in.ProjectID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ProjectOut](out)
}

// ProjectsDeleteIn identifies the project to remove.
type ProjectsDeleteIn struct {
	Company   string `json:"company"`
	ProjectID int64  `json:"project_id"`
}

// ProjectsDeleteOut is intentionally empty — DELETE returns NoContent.
type ProjectsDeleteOut struct{}

// TableHeader implements output.tableRow.
func (ProjectsDeleteOut) TableHeader() []string { return []string{"STATUS"} }

// TableRow implements output.tableRow.
func (ProjectsDeleteOut) TableRow() []string { return []string{"deleted"} }

// ProjectsDelete removes a project. Upstream returns 204 NoContent; we
// surface the success as a zero-value ProjectsDeleteOut.
func (c *Client) ProjectsDelete(ctx context.Context, in ProjectsDeleteIn) Result[ProjectsDeleteOut] {
	if in.Company == "" {
		return Err[ProjectsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpProjectsDelete,
		})
	}
	if in.ProjectID == 0 {
		return Err[ProjectsDeleteOut](&Error{
			Code: CodeValidation, Message: "project_id is required", Op: OpProjectsDelete,
		})
	}
	if err := c.gen.DeleteProject(ctx, fiken.DeleteProjectParams{
		CompanySlug: in.Company,
		ProjectId:   in.ProjectID,
	}); err != nil {
		return Err[ProjectsDeleteOut](MapErr(OpProjectsDelete, err))
	}
	return Ok[ProjectsDeleteOut](ProjectsDeleteOut{})
}
