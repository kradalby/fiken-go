package ops

import (
	"context"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/kradalby/fiken-go/fiken"
)

// ContactsListIn carries paged-list input for contacts. Company is
// required because the endpoint is /companies/{slug}/contacts.
type ContactsListIn struct {
	Company  string `json:"company"`
	PageSize int    `json:"page_size,omitempty"`
	Page     int    `json:"page,omitempty"`
}

// ContactOut is the canonical single-contact shape exposed to CLI/MCP.
// We keep only the high-traffic identity fields; deeper details (notes,
// documents, contact persons) round-trip through MCP-only payloads or
// later targeted ops. Location surfaces the upstream Location header
// for Create/Update responses (empty for Get/List rows).
type ContactOut struct {
	ContactID          int64  `json:"contact_id,omitempty"`
	Name               string `json:"name"`
	Email              string `json:"email,omitempty"`
	OrganizationNumber string `json:"organization_number,omitempty"`
	PhoneNumber        string `json:"phone_number,omitempty"`
	Currency           string `json:"currency,omitempty"`
	Language           string `json:"language,omitempty"`
	Customer           bool   `json:"customer,omitempty"`
	Supplier           bool   `json:"supplier,omitempty"`
	Inactive           bool   `json:"inactive,omitempty"`
	Location           string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (c ContactOut) TableHeader() []string {
	return []string{"ID", "NAME", "EMAIL", "ORG.NR"}
}

// TableRow implements output.tableRow.
func (c ContactOut) TableRow() []string {
	return []string{strconv.FormatInt(c.ContactID, 10), c.Name, c.Email, c.OrganizationNumber}
}

// ContactsListOut is the paged response.
type ContactsListOut = ListOut[ContactOut]

// ContactsList returns contacts for the specified company.
func (c *Client) ContactsList(ctx context.Context, in ContactsListIn) Result[ContactsListOut] {
	if in.Company == "" {
		return Err[ContactsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsList,
		})
	}
	params := fiken.GetContactsParams{CompanySlug: in.Company}
	if in.Page > 0 {
		params.Page.SetTo(in.Page)
	}
	if in.PageSize > 0 {
		params.PageSize.SetTo(in.PageSize)
	}
	resp, err := c.gen.GetContacts(ctx, params)
	if err != nil {
		return Err[ContactsListOut](MapErr(OpContactsList, err))
	}
	return Ok[ContactsListOut](translateContactsList(resp))
}

// translateContactsList converts the ogen response into the canonical
// ListOut[ContactOut] envelope, including paging metadata.
func translateContactsList(resp *fiken.GetContactsOKHeaders) ContactsListOut {
	if resp == nil {
		return ContactsListOut{Items: []ContactOut{}, Meta: ListMeta{}}
	}
	items := make([]ContactOut, 0, len(resp.Response))
	for _, co := range resp.Response {
		items = append(items, contactToOut(co))
	}
	meta := ListMeta{Returned: len(items)}
	if page, ok := resp.FikenAPIPage.Get(); ok {
		if pageCount, ok2 := resp.FikenAPIPageCount.Get(); ok2 && page+1 < pageCount {
			meta.NextPage = page + 1
		}
	}
	return ContactsListOut{Items: items, Meta: meta}
}

// contactToOut maps fiken.Contact to the canonical ContactOut.
func contactToOut(co fiken.Contact) ContactOut {
	return ContactOut{
		ContactID:          co.ContactId.Or(0),
		Name:               co.Name,
		Email:              co.Email.Or(""),
		OrganizationNumber: co.OrganizationNumber.Or(""),
		PhoneNumber:        co.PhoneNumber.Or(""),
		Currency:           co.Currency.Or(""),
		Language:           co.Language.Or(""),
		Customer:           co.Customer.Or(false),
		Supplier:           co.Supplier.Or(false),
		Inactive:           co.Inactive.Or(false),
	}
}

// ContactsGetIn requires company + contact ID.
type ContactsGetIn struct {
	Company   string `json:"company"`
	ContactID int64  `json:"contact_id"`
}

// ContactsGet returns a single contact by id.
func (c *Client) ContactsGet(ctx context.Context, in ContactsGetIn) Result[ContactOut] {
	if in.Company == "" {
		return Err[ContactOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsGet,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsGet,
		})
	}
	resp, err := c.gen.GetContact(ctx, fiken.GetContactParams{
		CompanySlug: in.Company,
		ContactId:   in.ContactID,
	})
	if err != nil {
		return Err[ContactOut](MapErr(OpContactsGet, err))
	}
	if resp == nil {
		return Ok[ContactOut](ContactOut{})
	}
	return Ok[ContactOut](contactToOut(*resp))
}

// ContactsCreateIn carries the create payload alongside the company.
// Body is the upstream fiken.Contact shape so the field surface stays
// in lock-step with the spec without us hand-curating every column.
type ContactsCreateIn struct {
	Company string         `json:"company"`
	Body    *fiken.Contact `json:"body"`
}

// ContactsCreate posts a new contact. Fiken returns 201 with the new
// resource URL in the Location header; we expose that via ContactOut.
func (c *Client) ContactsCreate(ctx context.Context, in ContactsCreateIn) Result[ContactOut] {
	if in.Company == "" {
		return Err[ContactOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsCreate,
		})
	}
	if in.Body == nil {
		return Err[ContactOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpContactsCreate,
		})
	}
	resp, err := c.gen.CreateContact(ctx, in.Body, fiken.CreateContactParams{
		CompanySlug: in.Company,
	})
	if err != nil {
		return Err[ContactOut](MapErr(OpContactsCreate, err))
	}
	out := ContactOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ContactOut](out)
}

// ContactsUpdateIn carries the update payload + identifiers.
type ContactsUpdateIn struct {
	Company   string         `json:"company"`
	ContactID int64          `json:"contact_id"`
	Body      *fiken.Contact `json:"body"`
}

// ContactsUpdate patches an existing contact. Fiken returns 200 with a
// Location header pointing back at the contact; ContactID echoes the
// input for table-renderer ergonomics.
func (c *Client) ContactsUpdate(ctx context.Context, in ContactsUpdateIn) Result[ContactOut] {
	if in.Company == "" {
		return Err[ContactOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsUpdate,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsUpdate,
		})
	}
	if in.Body == nil {
		return Err[ContactOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpContactsUpdate,
		})
	}
	resp, err := c.gen.UpdateContact(ctx, in.Body, fiken.UpdateContactParams{
		CompanySlug: in.Company,
		ContactId:   in.ContactID,
	})
	if err != nil {
		return Err[ContactOut](MapErr(OpContactsUpdate, err))
	}
	out := ContactOut{ContactID: in.ContactID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ContactOut](out)
}

// ContactsDeleteIn identifies the contact to remove.
type ContactsDeleteIn struct {
	Company   string `json:"company"`
	ContactID int64  `json:"contact_id"`
}

// ContactsDeleteOut is intentionally empty — DELETE returns NoContent
// or the (now inactive) Contact, both surfaced as success.
type ContactsDeleteOut struct{}

// TableHeader implements output.tableRow.
func (ContactsDeleteOut) TableHeader() []string { return []string{"STATUS"} }

// TableRow implements output.tableRow.
func (ContactsDeleteOut) TableRow() []string { return []string{"deleted"} }

// ContactsDelete removes (or deactivates) a contact. Upstream returns
// NoContent or the now-inactive Contact; both flow through as success.
func (c *Client) ContactsDelete(ctx context.Context, in ContactsDeleteIn) Result[ContactsDeleteOut] {
	if in.Company == "" {
		return Err[ContactsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsDelete,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactsDeleteOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsDelete,
		})
	}
	if _, err := c.gen.DeleteContact(ctx, fiken.DeleteContactParams{
		CompanySlug: in.Company,
		ContactId:   in.ContactID,
	}); err != nil {
		return Err[ContactsDeleteOut](MapErr(OpContactsDelete, err))
	}
	return Ok[ContactsDeleteOut](ContactsDeleteOut{})
}

// ContactPersonOut is the canonical contact-person shape. Location
// surfaces the Created/OK header for Create / Update responses.
type ContactPersonOut struct {
	ContactPersonID int64  `json:"contact_person_id,omitempty"`
	Name            string `json:"name"`
	Email           string `json:"email,omitempty"`
	PhoneNumber     string `json:"phone_number,omitempty"`
	Location        string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (p ContactPersonOut) TableHeader() []string {
	return []string{"ID", "NAME", "EMAIL", "PHONE"}
}

// TableRow implements output.tableRow.
func (p ContactPersonOut) TableRow() []string {
	return []string{strconv.FormatInt(p.ContactPersonID, 10), p.Name, p.Email, p.PhoneNumber}
}

// ContactPersonsListOut is the paged response for contact persons.
// Contact-persons endpoint returns a bare array (no pagination
// headers), so Meta.Returned is the only field that gets set.
type ContactPersonsListOut = ListOut[ContactPersonOut]

// ContactsPersonsListIn requires company + contactID.
type ContactsPersonsListIn struct {
	Company   string `json:"company"`
	ContactID int64  `json:"contact_id"`
}

// ContactsPersonsList returns all contact persons attached to a
// contact.
func (c *Client) ContactsPersonsList(ctx context.Context, in ContactsPersonsListIn) Result[ContactPersonsListOut] {
	if in.Company == "" {
		return Err[ContactPersonsListOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsPersonsList,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactPersonsListOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsPersonsList,
		})
	}
	resp, err := c.gen.GetContactContactPerson(ctx, fiken.GetContactContactPersonParams{
		CompanySlug: in.Company,
		ContactId:   in.ContactID,
	})
	if err != nil {
		return Err[ContactPersonsListOut](MapErr(OpContactsPersonsList, err))
	}
	items := make([]ContactPersonOut, 0, len(resp))
	for _, p := range resp {
		items = append(items, contactPersonToOut(p))
	}
	return Ok[ContactPersonsListOut](ContactPersonsListOut{
		Items: items, Meta: ListMeta{Returned: len(items)},
	})
}

// ContactsPersonsGetIn identifies a single contact person.
type ContactsPersonsGetIn struct {
	Company         string `json:"company"`
	ContactID       int64  `json:"contact_id"`
	ContactPersonID int64  `json:"contact_person_id"`
}

// ContactsPersonsGet returns a single contact person.
func (c *Client) ContactsPersonsGet(ctx context.Context, in ContactsPersonsGetIn) Result[ContactPersonOut] {
	if in.Company == "" {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsPersonsGet,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsPersonsGet,
		})
	}
	if in.ContactPersonID == 0 {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "contact_person_id is required", Op: OpContactsPersonsGet,
		})
	}
	resp, err := c.gen.GetContactPerson(ctx, fiken.GetContactPersonParams{
		CompanySlug:     in.Company,
		ContactId:       in.ContactID,
		ContactPersonId: in.ContactPersonID,
	})
	if err != nil {
		return Err[ContactPersonOut](MapErr(OpContactsPersonsGet, err))
	}
	if resp == nil {
		return Ok[ContactPersonOut](ContactPersonOut{})
	}
	return Ok[ContactPersonOut](contactPersonToOut(*resp))
}

func contactPersonToOut(p fiken.ContactPerson) ContactPersonOut {
	return ContactPersonOut{
		ContactPersonID: p.ContactPersonId.Or(0),
		Name:            p.Name,
		Email:           p.Email,
		PhoneNumber:     p.PhoneNumber.Or(""),
	}
}

// ContactsPersonsCreateIn carries the create payload + parent ids.
type ContactsPersonsCreateIn struct {
	Company   string               `json:"company"`
	ContactID int64                `json:"contact_id"`
	Body      *fiken.ContactPerson `json:"body"`
}

// ContactsPersonsCreate attaches a new contact-person to a contact.
// Fiken returns 200 with a Location header.
func (c *Client) ContactsPersonsCreate(ctx context.Context, in ContactsPersonsCreateIn) Result[ContactPersonOut] {
	if in.Company == "" {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsPersonsCreate,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsPersonsCreate,
		})
	}
	if in.Body == nil {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpContactsPersonsCreate,
		})
	}
	resp, err := c.gen.AddContactPersonToContact(ctx, in.Body, fiken.AddContactPersonToContactParams{
		CompanySlug: in.Company,
		ContactId:   in.ContactID,
	})
	if err != nil {
		return Err[ContactPersonOut](MapErr(OpContactsPersonsCreate, err))
	}
	out := ContactPersonOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ContactPersonOut](out)
}

// ContactsPersonsUpdateIn carries the update payload + identifiers.
type ContactsPersonsUpdateIn struct {
	Company         string               `json:"company"`
	ContactID       int64                `json:"contact_id"`
	ContactPersonID int64                `json:"contact_person_id"`
	Body            *fiken.ContactPerson `json:"body"`
}

// ContactsPersonsUpdate patches a contact-person in place. Upstream
// returns 200 with a Location header echoing the resource URL.
func (c *Client) ContactsPersonsUpdate(ctx context.Context, in ContactsPersonsUpdateIn) Result[ContactPersonOut] {
	if in.Company == "" {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsPersonsUpdate,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsPersonsUpdate,
		})
	}
	if in.ContactPersonID == 0 {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "contact_person_id is required", Op: OpContactsPersonsUpdate,
		})
	}
	if in.Body == nil {
		return Err[ContactPersonOut](&Error{
			Code: CodeValidation, Message: "body is required", Op: OpContactsPersonsUpdate,
		})
	}
	resp, err := c.gen.UpdateContactContactPerson(ctx, in.Body, fiken.UpdateContactContactPersonParams{
		CompanySlug:     in.Company,
		ContactId:       in.ContactID,
		ContactPersonId: in.ContactPersonID,
	})
	if err != nil {
		return Err[ContactPersonOut](MapErr(OpContactsPersonsUpdate, err))
	}
	out := ContactPersonOut{ContactPersonID: in.ContactPersonID}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ContactPersonOut](out)
}

// ContactsPersonsDeleteIn identifies the contact person to remove.
type ContactsPersonsDeleteIn struct {
	Company         string `json:"company"`
	ContactID       int64  `json:"contact_id"`
	ContactPersonID int64  `json:"contact_person_id"`
}

// ContactsPersonsDeleteOut mirrors ContactsDeleteOut.
type ContactsPersonsDeleteOut struct{}

// TableHeader implements output.tableRow.
func (ContactsPersonsDeleteOut) TableHeader() []string { return []string{"STATUS"} }

// TableRow implements output.tableRow.
func (ContactsPersonsDeleteOut) TableRow() []string { return []string{"deleted"} }

// ContactsPersonsDelete removes a contact-person. Upstream returns
// NoContent; the empty Out signals success.
func (c *Client) ContactsPersonsDelete(ctx context.Context, in ContactsPersonsDeleteIn) Result[ContactsPersonsDeleteOut] {
	if in.Company == "" {
		return Err[ContactsPersonsDeleteOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsPersonsDelete,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactsPersonsDeleteOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsPersonsDelete,
		})
	}
	if in.ContactPersonID == 0 {
		return Err[ContactsPersonsDeleteOut](&Error{
			Code: CodeValidation, Message: "contact_person_id is required", Op: OpContactsPersonsDelete,
		})
	}
	if err := c.gen.DeleteContactContactPerson(ctx, fiken.DeleteContactContactPersonParams{
		CompanySlug:     in.Company,
		ContactId:       in.ContactID,
		ContactPersonId: in.ContactPersonID,
	}); err != nil {
		return Err[ContactsPersonsDeleteOut](MapErr(OpContactsPersonsDelete, err))
	}
	return Ok[ContactsPersonsDeleteOut](ContactsPersonsDeleteOut{})
}

// ContactsAttachmentsAttachIn carries the multipart upload payload
// for contact attachments. Filename overrides the form-field name;
// FilePath is the local path read by ops.OpenMultipartFile. Comment
// is an optional free-text annotation.
type ContactsAttachmentsAttachIn struct {
	Company   string `json:"company"`
	ContactID int64  `json:"contact_id"`
	Filename  string `json:"filename"`
	FilePath  string `json:"file_path"`
	Comment   string `json:"comment"`
}

// ContactsAttachmentsAttachOut mirrors the upstream Created response.
type ContactsAttachmentsAttachOut struct {
	Location string `json:"location,omitempty"`
}

// TableHeader implements output.tableRow.
func (ContactsAttachmentsAttachOut) TableHeader() []string { return []string{"LOCATION"} }

// TableRow implements output.tableRow.
func (a ContactsAttachmentsAttachOut) TableRow() []string { return []string{a.Location} }

// ContactsAttachmentsAttach uploads a file to a contact as multipart
// form data. The form-field filename defaults to the basename of
// FilePath; pass Filename to override.
func (c *Client) ContactsAttachmentsAttach(ctx context.Context, in ContactsAttachmentsAttachIn) Result[ContactsAttachmentsAttachOut] {
	if in.Company == "" {
		return Err[ContactsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "company is required", Op: OpContactsAttachmentsAttach,
		})
	}
	if in.ContactID == 0 {
		return Err[ContactsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "contact_id is required", Op: OpContactsAttachmentsAttach,
		})
	}
	if in.FilePath == "" {
		return Err[ContactsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: "file is required", Op: OpContactsAttachmentsAttach,
		})
	}
	file, closeFn, err := OpenMultipartFile(in.FilePath, in.Filename)
	if err != nil {
		return Err[ContactsAttachmentsAttachOut](&Error{
			Code: CodeValidation, Message: err.Error(), Op: OpContactsAttachmentsAttach,
		})
	}
	defer func() { _ = closeFn() }()
	name := in.Filename
	if name == "" {
		name = filepath.Base(in.FilePath)
	}
	body := fiken.AddAttachmentToContactReq{
		Filename: fiken.NewOptString(name),
		File:     file,
	}
	if in.Comment != "" {
		body.Comment = fiken.NewOptString(in.Comment)
	}
	req := fiken.NewOptAddAttachmentToContactReq(body)
	resp, err := c.gen.AddAttachmentToContact(ctx, req, fiken.AddAttachmentToContactParams{
		CompanySlug: in.Company,
		ContactId:   in.ContactID,
	})
	if err != nil {
		return Err[ContactsAttachmentsAttachOut](MapErr(OpContactsAttachmentsAttach, err))
	}
	out := ContactsAttachmentsAttachOut{}
	if resp != nil {
		u := resp.Location.Or(url.URL{})
		out.Location = u.String()
	}
	return Ok[ContactsAttachmentsAttachOut](out)
}
