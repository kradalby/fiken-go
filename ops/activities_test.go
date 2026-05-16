package ops

import (
	"context"
	"net/url"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestActivitiesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesList(context.Background(), ActivitiesListIn{})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
}

func TestActivitiesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ActivitiesGet(context.Background(), ActivitiesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ActivitiesGet(context.Background(), ActivitiesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing activity_id: want validation error, got %+v", got)
	}
}

// TestActivitiesListAgainstMock exercises the default empty-list path.
func TestActivitiesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesList(context.Background(), ActivitiesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %+v", res.Ok)
	}
}

// TestActivitiesListAgainstMockOverride asserts the success-override
// flows back through translation.
func TestActivitiesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpActivitiesList, &fiken.GetActivitiesOKHeaders{
		Response: []fiken.ActivityResult{{
			ActivityId:  fiken.OptInt64{Value: 9, Set: true},
			Name:        fiken.OptString{Value: "Programmering", Set: true},
			HourlyRate:  fiken.OptInt64{Value: 125000, Set: true},
			Description: fiken.OptString{Value: "Generelt utviklingsarbeid", Set: true},
			Billable:    fiken.OptBool{Value: true, Set: true},
			Archived:    fiken.OptBool{Value: false, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesList(context.Background(), ActivitiesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.ActivityID != 9 || got.Name != "Programmering" || got.HourlyRate != 125000 ||
		got.Description != "Generelt utviklingsarbeid" || !got.Billable || got.Archived {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestActivitiesGetAgainstMock asserts the single-resource happy path.
func TestActivitiesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpActivitiesGet, &fiken.ActivityResult{
		ActivityId: fiken.OptInt64{Value: 11, Set: true},
		Name:       fiken.OptString{Value: "Design", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesGet(context.Background(), ActivitiesGetIn{Company: "acme", ActivityID: 11})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ActivityID != 11 || res.Ok.Name != "Design" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestActivitiesCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ActivitiesCreate(context.Background(), ActivitiesCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ActivitiesCreate(context.Background(), ActivitiesCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestActivitiesCreateAgainstMock asserts the create path renders the
// Location header through ActivityOut.Location.
func TestActivitiesCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/activities/55")
	mock.Set(OpActivitiesCreate, &fiken.CreateActivityCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesCreate(context.Background(), ActivitiesCreateIn{
		Company: "acme",
		Body:    &fiken.ActivityRequest{Name: "Coding"},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestActivitiesCreatePaymentRequired confirms a 402 from the paid-tier
// time-tracking endpoint maps through MapErr → CodePaymentRequired.
func TestActivitiesCreatePaymentRequired(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpActivitiesCreate, 402, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesCreate(context.Background(), ActivitiesCreateIn{
		Company: "acme",
		Body:    &fiken.ActivityRequest{Name: "Coding"},
	})
	if res.Error == nil || res.Error.Code != CodePaymentRequired {
		t.Fatalf("expected CodePaymentRequired, got %+v", res.Error)
	}
}

func TestActivitiesUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/activities/7")
	mock.Set(OpActivitiesUpdate, &fiken.UpdateActivityOK{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesUpdate(context.Background(), ActivitiesUpdateIn{
		Company: "acme", ActivityID: 7,
		Body: &fiken.UpdateActivityRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.ActivityID != 7 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestActivitiesUpdateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ActivitiesUpdate(context.Background(), ActivitiesUpdateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ActivitiesUpdate(context.Background(), ActivitiesUpdateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing activity_id: want validation error, got %+v", got)
	}
	if got := c.ActivitiesUpdate(context.Background(), ActivitiesUpdateIn{Company: "acme", ActivityID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestActivitiesDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	_ = mock

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.ActivitiesDelete(context.Background(), ActivitiesDeleteIn{Company: "acme", ActivityID: 7})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("unexpected nil ok")
	}
}

func TestActivitiesDeleteValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.ActivitiesDelete(context.Background(), ActivitiesDeleteIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.ActivitiesDelete(context.Background(), ActivitiesDeleteIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing activity_id: want validation error, got %+v", got)
	}
}
