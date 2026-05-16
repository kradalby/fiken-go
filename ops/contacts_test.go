package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestContactsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsList(context.Background(), ContactsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpContactsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpContactsList)
	}
}

func TestContactsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ContactsGet(context.Background(), ContactsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ContactsGet(context.Background(), ContactsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing contact_id: want validation error, got %+v", got)
	}
}

// TestContactsListAgainstMock exercises the default empty-list path
// through ops + ogen + httptest.
func TestContactsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsList(context.Background(), ContactsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("nil Ok")
	}
	if len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %d", len(res.Ok.Items))
	}
}

// TestContactsListAgainstMockOverride asserts the success-override
// flows back through the translation layer.
func TestContactsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpContactsList, &fiken.GetContactsOKHeaders{
		Response: []fiken.Contact{{
			ContactId: fiken.OptInt64{Value: 42, Set: true},
			Name:      "Wile E. Coyote",
			Email:     fiken.OptString{Value: "wile@acme.test", Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsList(context.Background(), ContactsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.ContactID != 42 || got.Name != "Wile E. Coyote" || got.Email != "wile@acme.test" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestContactsGetAgainstMock asserts the single-resource happy path.
func TestContactsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpContactsGet, &fiken.Contact{
		ContactId: fiken.OptInt64{Value: 99, Set: true},
		Name:      "Acme Co",
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsGet(context.Background(), ContactsGetIn{Company: "acme", ContactID: 99})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ContactID != 99 || res.Ok.Name != "Acme Co" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestContactsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ContactsCreate(context.Background(), ContactsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ContactsCreate(context.Background(), ContactsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestContactsCreateAgainstMock asserts the create path renders the
// Location header through ContactOut.Location.
func TestContactsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/contacts/42")
	mock.Set(OpContactsCreate, &fiken.CreateContactCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsCreate(context.Background(), ContactsCreateIn{
		Company: "acme",
		Body:    &fiken.Contact{Name: "Acme"},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestContactsCreateError verifies upstream non-2xx is mapped through
// MapErr — exercising mockfiken.SetError's HTTP status propagation.
func TestContactsCreateError(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpContactsCreate, 404, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsCreate(context.Background(), ContactsCreateIn{
		Company: "acme",
		Body:    &fiken.Contact{Name: "Acme"},
	})
	if res.Error == nil || res.Error.Code != CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %+v", res.Error)
	}
}

// TestContactsUpdateAgainstMock asserts the update path returns the
// echoed ContactID + Location header.
func TestContactsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/contacts/42")
	mock.Set(OpContactsUpdate, &fiken.UpdateContactOK{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsUpdate(context.Background(), ContactsUpdateIn{
		Company:   "acme",
		ContactID: 42,
		Body:      &fiken.Contact{Name: "Acme"},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ContactID != 42 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestContactsDeleteAgainstMock asserts the delete path returns the
// empty success struct. The mock's default NoContent response satisfies
// the upstream DeleteContactRes interface via DeleteContactNoContent.
func TestContactsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpContactsDelete, &fiken.DeleteContactNoContent{})
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsDelete(context.Background(), ContactsDeleteIn{
		Company:   "acme",
		ContactID: 42,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
}

// TestContactsPersonsCreateAgainstMock covers the contact-person CUD
// happy paths in one go.
func TestContactsPersonsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/contacts/42/contactPerson/7")
	mock.Set(OpContactsPersonsCreate, &fiken.AddContactPersonToContactOK{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsPersonsCreate(context.Background(), ContactsPersonsCreateIn{
		Company:   "acme",
		ContactID: 42,
		Body:      &fiken.ContactPerson{Name: "Wile"},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestContactsPersonsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/contacts/42/contactPerson/7")
	mock.Set(OpContactsPersonsUpdate, &fiken.UpdateContactContactPersonOK{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsPersonsUpdate(context.Background(), ContactsPersonsUpdateIn{
		Company:         "acme",
		ContactID:       42,
		ContactPersonID: 7,
		Body:            &fiken.ContactPerson{Name: "Wile"},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ContactPersonID != 7 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestContactsPersonsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsPersonsDelete(context.Background(), ContactsPersonsDeleteIn{
		Company:         "acme",
		ContactID:       42,
		ContactPersonID: 7,
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
}

func TestContactsAttachmentsAttachValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ContactsAttachmentsAttach(context.Background(), ContactsAttachmentsAttachIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ContactsAttachmentsAttach(context.Background(), ContactsAttachmentsAttachIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing contact_id: want validation error, got %+v", got)
	}
	if got := c.ContactsAttachmentsAttach(context.Background(), ContactsAttachmentsAttachIn{Company: "acme", ContactID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing file: want validation error, got %+v", got)
	}
}

func TestContactsAttachmentsAttachAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/contacts/1/attachments/9")
	mock.Set(OpContactsAttachmentsAttach, &fiken.AddAttachmentToContactCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	path := writeTempAttachment(t)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ContactsAttachmentsAttach(context.Background(), ContactsAttachmentsAttachIn{
		Company:   "acme",
		ContactID: 1,
		FilePath:  path,
		Comment:   "annual report",
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
