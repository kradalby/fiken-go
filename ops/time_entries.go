package ops

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/kradalby/fiken-go/fiken"
)

// TimeEntriesListIn carries paged-list input for time entries. The
// upstream endpoint exposes a broad filter surface (exact-date + date
// range, project, activity, time user, invoiced) — we forward each as
// an optional field. Invoiced is *bool so we distinguish "unset"
// (return all) from "false" (return only un-invoiced entries).
type TimeEntriesListIn struct {
	Company    string `json:"company"`
	Page       int    `json:"page,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
	Date       Date   `json:"date,omitempty"`
	DateGe     Date   `json:"date_ge,omitempty"`
	DateLe     Date   `json:"date_le,omitempty"`
	ProjectID  int64  `json:"project_id,omitempty"`
	ActivityID int64  `json:"activity_id,omitempty"`
	TimeUserID int64  `json:"time_user_id,omitempty"`
	Invoiced   *bool  `json:"invoiced,omitempty"`
}

// TimeEntryOut is the canonical time-entry shape exposed to CLI/MCP.
// Hours is float64 per the upstream spec. Activity / Project / TimeUser
// are flattened to their ids; callers can fetch the full shapes via the
// dedicated ops. Location surfaces the upstream Location header for
// Create/Update responses; list/get leave it zero.
type TimeEntryOut struct {
	TimeEntryID  int64   `json:"time_entry_id,omitempty"`
	Date         Date    `json:"date,omitempty"`
	Hours        float64 `json:"hours,omitempty"`
	StartTime    string  `json:"start_time,omitempty"`
	EndTime      string  `json:"end_time,omitempty"`
	Description  string  `json:"description,omitempty"`
	InternalNote string  `json:"internal_note,omitempty"`
	ActivityID   int64   `json:"activity_id,omitempty"`
	ProjectID    int64   `json:"project_id,omitempty"`
	TimeUserID   int64   `json:"time_user_id,omitempty"`
	Invoiced     bool    `json:"invoiced,omitempty"`
	InvoiceID    int64   `json:"invoice_id,omitempty"`
	Locked       bool    `json:"locked,omitempty"`
	Location     string  `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (TimeEntryOut) TableHeader() []string {
	return []string{"ID", "DATE", "HOURS", "ACTIVITY", "USER", "INVOICED"}
}

// TableRow implements output.tableRow.
func (e TimeEntryOut) TableRow() []string {
	return []string{
		strconv.FormatInt(e.TimeEntryID, 10),
		string(e.Date),
		strconv.FormatFloat(e.Hours, 'f', 2, 64),
		strconv.FormatInt(e.ActivityID, 10),
		strconv.FormatInt(e.TimeUserID, 10),
		strconv.FormatBool(e.Invoiced),
	}
}

// TimeEntriesListOut is the paged response.
type TimeEntriesListOut = ListOut[TimeEntryOut]

// TimeEntriesList returns the time entries for the specified company.
func (c *Client) TimeEntriesList(ctx context.Context, in TimeEntriesListIn) Result[TimeEntriesListOut] {
	if in.Company == "" {
		return Err[TimeEntriesListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeEntriesList,
		})
	}
	params := fiken.GetTimeEntriesParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	if in.Date != "" {
		d, err := parseDate(string(in.Date))
		if err != nil {
			return Err[TimeEntriesListOut](&Error{
				Code: CodeValidation, Message: "date must be YYYY-MM-DD", Op: OpTimeEntriesList,
			})
		}
		params.Date.SetTo(d)
	}
	if in.DateGe != "" {
		d, err := parseDate(string(in.DateGe))
		if err != nil {
			return Err[TimeEntriesListOut](&Error{
				Code: CodeValidation, Message: "date_ge must be YYYY-MM-DD", Op: OpTimeEntriesList,
			})
		}
		params.DateGe.SetTo(d)
	}
	if in.DateLe != "" {
		d, err := parseDate(string(in.DateLe))
		if err != nil {
			return Err[TimeEntriesListOut](&Error{
				Code: CodeValidation, Message: "date_le must be YYYY-MM-DD", Op: OpTimeEntriesList,
			})
		}
		params.DateLe.SetTo(d)
	}
	if in.ProjectID > 0 {
		params.ProjectId.SetTo(in.ProjectID)
	}
	if in.ActivityID > 0 {
		params.ActivityId.SetTo(in.ActivityID)
	}
	if in.TimeUserID > 0 {
		params.TimeUserId.SetTo(in.TimeUserID)
	}
	if in.Invoiced != nil {
		params.Invoiced.SetTo(*in.Invoiced)
	}
	resp, err := c.gen.GetTimeEntries(ctx, params)
	if err != nil {
		return Err[TimeEntriesListOut](MapErr(OpTimeEntriesList, err))
	}
	return Ok[TimeEntriesListOut](translateTimeEntriesList(resp))
}

// translateTimeEntriesList converts the ogen response into the canonical
// ListOut[TimeEntryOut] envelope, including paging meta.
func translateTimeEntriesList(resp *fiken.GetTimeEntriesOKHeaders) TimeEntriesListOut {
	if resp == nil {
		return TimeEntriesListOut{Items: []TimeEntryOut{}, Meta: ListMeta{}}
	}
	items := make([]TimeEntryOut, 0, len(resp.Response))
	for _, e := range resp.Response {
		items = append(items, timeEntryToOut(e))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return TimeEntriesListOut{Items: items, Meta: meta}
}

// timeEntryToOut maps fiken.TimeEntryResult into TimeEntryOut. Embedded
// Activity / Project / TimeUser are flattened to their ids.
func timeEntryToOut(e fiken.TimeEntryResult) TimeEntryOut {
	out := TimeEntryOut{
		TimeEntryID:  e.TimeEntryId.Or(0),
		Hours:        e.Hours.Or(0),
		StartTime:    e.StartTime.Or(""),
		EndTime:      e.EndTime.Or(""),
		Description:  e.Description.Or(""),
		InternalNote: e.InternalNote.Or(""),
		Invoiced:     e.Invoiced.Or(false),
		InvoiceID:    e.InvoiceId.Or(0),
		Locked:       e.Locked.Or(false),
	}
	if d, ok := e.Date.Get(); ok {
		out.Date = Date(d.Format("2006-01-02"))
	}
	if a, ok := e.Activity.Get(); ok {
		out.ActivityID = a.ActivityId.Or(0)
	}
	if p, ok := e.Project.Get(); ok {
		out.ProjectID = p.ProjectId.Or(0)
	}
	if u, ok := e.TimeUser.Get(); ok {
		out.TimeUserID = u.TimeUserId.Or(0)
	}
	return out
}

// TimeEntriesGetIn requires company + time entry id.
type TimeEntriesGetIn struct {
	Company     string `json:"company"`
	TimeEntryID int64  `json:"time_entry_id"`
}

// TimeEntriesGet returns a single time entry by id.
func (c *Client) TimeEntriesGet(ctx context.Context, in TimeEntriesGetIn) Result[TimeEntryOut] {
	if in.Company == "" {
		return Err[TimeEntryOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeEntriesGet,
		})
	}
	if in.TimeEntryID == 0 {
		return Err[TimeEntryOut](&Error{
			Code: CodeValidation, Message: "time_entry_id is required", Op: OpTimeEntriesGet,
		})
	}
	resp, err := c.gen.GetTimeEntry(ctx, fiken.GetTimeEntryParams{
		CompanySlug: in.Company,
		TimeEntryId: in.TimeEntryID,
	})
	if err != nil {
		return Err[TimeEntryOut](MapErr(OpTimeEntriesGet, err))
	}
	if resp == nil {
		return Ok[TimeEntryOut](TimeEntryOut{})
	}
	return Ok[TimeEntryOut](timeEntryToOut(*resp))
}

// TimeEntriesCreateIn carries the create payload alongside the company.
// Body is the upstream fiken.TimeEntryRequest so the field surface stays
// in lock-step with the spec.
type TimeEntriesCreateIn struct {
	Company string                  `json:"company"`
	Body    *fiken.TimeEntryRequest `json:"body"`
}

// TimeEntriesCreate posts a new time entry. Upstream returns 201 with
// a Location header pointing at the new resource; surfaced via
// TimeEntryOut.Location. Time tracking is paid-tier on the real API so
// callers without billing will get 402 → CodePaymentRequired.
func (c *Client) TimeEntriesCreate(ctx context.Context, in TimeEntriesCreateIn) Result[TimeEntryOut] {
	if in.Company == "" {
		return Err[TimeEntryOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeEntriesCreate,
		})
	}
	if in.Body == nil {
		return Err[TimeEntryOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpTimeEntriesCreate,
		})
	}
	resp, err := c.gen.CreateTimeEntry(ctx, in.Body, fiken.CreateTimeEntryParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[TimeEntryOut](MapErr(OpTimeEntriesCreate, err))
	}
	out := TimeEntryOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[TimeEntryOut](out)
}

// TimeEntriesUpdateIn carries the update payload + identifiers. Body
// uses fiken.UpdateTimeEntryRequest (PATCH-shaped, optional fields)
// rather than TimeEntryRequest because the endpoint is PATCH.
type TimeEntriesUpdateIn struct {
	Company     string                        `json:"company"`
	TimeEntryID int64                         `json:"time_entry_id"`
	Body        *fiken.UpdateTimeEntryRequest `json:"body"`
}

// TimeEntriesUpdate patches an existing time entry. Upstream returns
// 200 with a Location header pointing back at the resource — we
// surface the time entry id alongside Location.
func (c *Client) TimeEntriesUpdate(ctx context.Context, in TimeEntriesUpdateIn) Result[TimeEntryOut] {
	if in.Company == "" {
		return Err[TimeEntryOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeEntriesUpdate,
		})
	}
	if in.TimeEntryID == 0 {
		return Err[TimeEntryOut](&Error{
			Code: CodeValidation, Message: "time_entry_id is required", Op: OpTimeEntriesUpdate,
		})
	}
	if in.Body == nil {
		return Err[TimeEntryOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpTimeEntriesUpdate,
		})
	}
	resp, err := c.gen.UpdateTimeEntry(ctx, in.Body, fiken.UpdateTimeEntryParams{
		CompanySlug: in.Company,
		TimeEntryId: in.TimeEntryID,
	})
	if err != nil {
		return Err[TimeEntryOut](MapErr(OpTimeEntriesUpdate, err))
	}
	out := TimeEntryOut{TimeEntryID: in.TimeEntryID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[TimeEntryOut](out)
}

// TimeEntriesDeleteIn identifies the time entry to remove.
type TimeEntriesDeleteIn struct {
	Company     string `json:"company"`
	TimeEntryID int64  `json:"time_entry_id"`
}

// TimeEntriesDeleteOut is intentionally empty — DELETE returns NoContent.
type TimeEntriesDeleteOut struct{}

// TableHeader implements output.tableRow.
func (TimeEntriesDeleteOut) TableHeader() []string { return []string{"STATUS"} }

// TableRow implements output.tableRow.
func (TimeEntriesDeleteOut) TableRow() []string { return []string{"deleted"} }

// TimeEntriesDelete removes a time entry. Upstream returns 204
// NoContent; we surface the success as a zero-value
// TimeEntriesDeleteOut. The ogen signature here returns a sum type
// (DeleteTimeEntryRes) because the spec also documents non-204
// responses — we only care that no error was returned by the client.
func (c *Client) TimeEntriesDelete(ctx context.Context, in TimeEntriesDeleteIn) Result[TimeEntriesDeleteOut] {
	if in.Company == "" {
		return Err[TimeEntriesDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeEntriesDelete,
		})
	}
	if in.TimeEntryID == 0 {
		return Err[TimeEntriesDeleteOut](&Error{
			Code: CodeValidation, Message: "time_entry_id is required", Op: OpTimeEntriesDelete,
		})
	}
	if _, err := c.gen.DeleteTimeEntry(ctx, fiken.DeleteTimeEntryParams{
		CompanySlug: in.Company,
		TimeEntryId: in.TimeEntryID,
	}); err != nil {
		return Err[TimeEntriesDeleteOut](MapErr(OpTimeEntriesDelete, err))
	}
	return Ok[TimeEntriesDeleteOut](TimeEntriesDeleteOut{})
}

// TimeEntriesInvoiceDraftFromTimesIn carries the request payload for
// the "create invoice draft from time entries" op. Body is the upstream
// fiken.TimeEntryInvoiceDraftRequest so the grouping + selection knobs
// stay in lock-step with the spec.
type TimeEntriesInvoiceDraftFromTimesIn struct {
	Company string                              `json:"company"`
	Body    *fiken.TimeEntryInvoiceDraftRequest `json:"body"`
}

// TimeEntriesInvoiceDraftFromTimesOut carries the result of bundling
// the supplied time entries into a new invoice draft. DraftID echoes
// the body of the response; Location surfaces the URL of the new
// draft from the response header. Both default to zero on absence.
type TimeEntriesInvoiceDraftFromTimesOut struct {
	DraftID  int64  `json:"draft_id,omitempty"`
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (TimeEntriesInvoiceDraftFromTimesOut) TableHeader() []string {
	return []string{"DRAFT_ID", "LOCATION"}
}

// TableRow implements output.tableRow.
func (o TimeEntriesInvoiceDraftFromTimesOut) TableRow() []string {
	return []string{strconv.FormatInt(o.DraftID, 10), o.Location}
}

// TimeEntriesInvoiceDraftFromTimes bundles a set of time entries into
// a new invoice draft. Upstream returns 201 with the draft id in the
// body and a Location header pointing at the new draft.
func (c *Client) TimeEntriesInvoiceDraftFromTimes(ctx context.Context, in TimeEntriesInvoiceDraftFromTimesIn) Result[TimeEntriesInvoiceDraftFromTimesOut] {
	if in.Company == "" {
		return Err[TimeEntriesInvoiceDraftFromTimesOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpTimeEntriesInvoiceDraftFromTimes,
		})
	}
	if in.Body == nil {
		return Err[TimeEntriesInvoiceDraftFromTimesOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpTimeEntriesInvoiceDraftFromTimes,
		})
	}
	if len(in.Body.TimeEntryIds) == 0 {
		return Err[TimeEntriesInvoiceDraftFromTimesOut](&Error{
			Code: CodeValidation, Message: "body.timeEntryIds is required", Op: OpTimeEntriesInvoiceDraftFromTimes,
		})
	}
	resp, err := c.gen.CreateInvoiceDraftFromTimeEntries(ctx, in.Body, fiken.CreateInvoiceDraftFromTimeEntriesParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[TimeEntriesInvoiceDraftFromTimesOut](MapErr(OpTimeEntriesInvoiceDraftFromTimes, err))
	}
	out := TimeEntriesInvoiceDraftFromTimesOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
		out.DraftID = resp.Response.DraftId.Or(0)
		// Fall back to the trailing path segment of the Location header
		// if the body did not echo the id — guards against shapes where
		// only the header is populated.
		if out.DraftID == 0 && out.Location != "" {
			if idx := strings.LastIndex(out.Location, "/"); idx >= 0 {
				if n, perr := strconv.ParseInt(out.Location[idx+1:], 10, 64); perr == nil {
					out.DraftID = n
				}
			}
		}
	}
	return Ok[TimeEntriesInvoiceDraftFromTimesOut](out)
}
