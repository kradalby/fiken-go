package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestProjectsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsList(context.Background(), ProjectsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpProjectsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpProjectsList)
	}
}

func TestProjectsGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ProjectsGet(context.Background(), ProjectsGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ProjectsGet(context.Background(), ProjectsGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing project_id: want validation error, got %+v", got)
	}
}

// TestProjectsListAgainstMock exercises the default empty-list path.
func TestProjectsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsList(context.Background(), ProjectsListIn{Company: "acme"})
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

// TestProjectsListAgainstMockOverride asserts the success-override
// flows back through the translation layer including the embedded
// Contact + date fields.
func TestProjectsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	mock.Set(OpProjectsList, &fiken.GetProjectsOKHeaders{
		Response: []fiken.ProjectResult{{
			ProjectId:   fiken.OptInt64{Value: 42, Set: true},
			Number:      fiken.OptString{Value: "P-001", Set: true},
			Name:        fiken.OptString{Value: "Roadrunner Hunt", Set: true},
			Description: fiken.OptString{Value: "ACME contract", Set: true},
			StartDate:   fiken.OptDate{Value: start, Set: true},
			EndDate:     fiken.OptDate{Value: end, Set: true},
			Contact: fiken.OptContact{Value: fiken.Contact{
				ContactId: fiken.OptInt64{Value: 7, Set: true},
				Name:      "Wile E. Coyote",
			}, Set: true},
			Completed: fiken.OptBool{Value: false, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsList(context.Background(), ProjectsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.ProjectID != 42 || got.Number != "P-001" || got.Name != "Roadrunner Hunt" ||
		got.Description != "ACME contract" || got.StartDate != "2024-01-15" ||
		got.EndDate != "2024-12-31" || got.ContactID != 7 ||
		got.ContactName != "Wile E. Coyote" || got.Completed {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestProjectsGetAgainstMock asserts the single-resource happy path.
func TestProjectsGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpProjectsGet, &fiken.ProjectResult{
		ProjectId: fiken.OptInt64{Value: 99, Set: true},
		Name:      fiken.OptString{Value: "Phase 2", Set: true},
		Completed: fiken.OptBool{Value: true, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsGet(context.Background(), ProjectsGetIn{Company: "acme", ProjectID: 99})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ProjectID != 99 || res.Ok.Name != "Phase 2" || !res.Ok.Completed {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestProjectsCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ProjectsCreate(context.Background(), ProjectsCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ProjectsCreate(context.Background(), ProjectsCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestProjectsCreateAgainstMock asserts the create path renders the
// Location header through ProjectOut.Location.
func TestProjectsCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/projects/42")
	mock.Set(OpProjectsCreate, &fiken.CreateProjectCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsCreate(context.Background(), ProjectsCreateIn{
		Company: "acme",
		Body:    &fiken.ProjectRequest{Name: "Roadrunner Hunt"},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestProjectsCreatePaymentRequired confirms a 402 from the paid-tier
// endpoint maps through MapErr → CodePaymentRequired.
func TestProjectsCreatePaymentRequired(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpProjectsCreate, 402, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsCreate(context.Background(), ProjectsCreateIn{
		Company: "acme",
		Body:    &fiken.ProjectRequest{Name: "Roadrunner Hunt"},
	})
	if res.Error == nil || res.Error.Code != CodePaymentRequired {
		t.Fatalf("expected CodePaymentRequired, got %+v", res.Error)
	}
}

func TestProjectsUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/projects/7")
	mock.Set(OpProjectsUpdate, &fiken.UpdateProjectCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsUpdate(context.Background(), ProjectsUpdateIn{
		Company: "acme", ProjectID: 7,
		Body: &fiken.UpdateProjectRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ProjectID != 7 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestProjectsUpdateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ProjectsUpdate(context.Background(), ProjectsUpdateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ProjectsUpdate(context.Background(), ProjectsUpdateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing project_id: want validation error, got %+v", got)
	}
	if got := c.ProjectsUpdate(context.Background(), ProjectsUpdateIn{Company: "acme", ProjectID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestProjectsDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	_ = mock

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ProjectsDelete(context.Background(), ProjectsDeleteIn{Company: "acme", ProjectID: 7})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("unexpected nil ok")
	}
}

func TestProjectsDeleteValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ProjectsDelete(context.Background(), ProjectsDeleteIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ProjectsDelete(context.Background(), ProjectsDeleteIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing project_id: want validation error, got %+v", got)
	}
}
