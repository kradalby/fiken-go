package ops

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestTimeEntriesListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesList(context.Background(), TimeEntriesListIn{})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
}

func TestTimeEntriesGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TimeEntriesGet(context.Background(), TimeEntriesGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TimeEntriesGet(context.Background(), TimeEntriesGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing time_entry_id: want validation error, got %+v", got)
	}
}

// TestTimeEntriesListAgainstMock exercises the default empty-list path.
func TestTimeEntriesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesList(context.Background(), TimeEntriesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %+v", res.Ok)
	}
}

// TestTimeEntriesListAgainstMockOverride asserts translation including
// the flattened Activity / Project / TimeUser ids and date encoding.
func TestTimeEntriesListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	d := time.Date(2024, 3, 14, 0, 0, 0, 0, time.UTC)
	mock.Set(OpTimeEntriesList, &fiken.GetTimeEntriesOKHeaders{
		Response: []fiken.TimeEntryResult{{
			TimeEntryId:  fiken.OptInt64{Value: 100, Set: true},
			Date:         fiken.OptDate{Value: d, Set: true},
			Hours:        fiken.OptFloat64{Value: 4.5, Set: true},
			StartTime:    fiken.OptString{Value: "09:00", Set: true},
			EndTime:      fiken.OptString{Value: "13:30", Set: true},
			Description:  fiken.OptString{Value: "Sprint planlegging", Set: true},
			InternalNote: fiken.OptString{Value: "internal", Set: true},
			Activity: fiken.OptActivityResult{Value: fiken.ActivityResult{
				ActivityId: fiken.OptInt64{Value: 11, Set: true},
			}, Set: true},
			Project: fiken.OptProjectResult{Value: fiken.ProjectResult{
				ProjectId: fiken.OptInt64{Value: 42, Set: true},
			}, Set: true},
			TimeUser: fiken.OptTimeUserResult{Value: fiken.TimeUserResult{
				TimeUserId: fiken.OptInt64{Value: 3, Set: true},
			}, Set: true},
			Invoiced:  fiken.OptBool{Value: false, Set: true},
			InvoiceId: fiken.OptInt64{Value: 0, Set: true},
			Locked:    fiken.OptBool{Value: false, Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesList(context.Background(), TimeEntriesListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.TimeEntryID != 100 || got.Date != "2024-03-14" || got.Hours != 4.5 ||
		got.StartTime != "09:00" || got.EndTime != "13:30" ||
		got.Description != "Sprint planlegging" || got.InternalNote != "internal" ||
		got.ActivityID != 11 || got.ProjectID != 42 || got.TimeUserID != 3 || got.Invoiced {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestTimeEntriesGetAgainstMock asserts the single-resource happy path.
func TestTimeEntriesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpTimeEntriesGet, &fiken.TimeEntryResult{
		TimeEntryId: fiken.OptInt64{Value: 7, Set: true},
		Hours:       fiken.OptFloat64{Value: 1.0, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesGet(context.Background(), TimeEntriesGetIn{Company: "acme", TimeEntryID: 7})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.TimeEntryID != 7 || res.Ok.Hours != 1.0 {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestTimeEntriesCreateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TimeEntriesCreate(context.Background(), TimeEntriesCreateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TimeEntriesCreate(context.Background(), TimeEntriesCreateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

// TestTimeEntriesCreateAgainstMock asserts the create path renders the
// Location header through TimeEntryOut.Location.
func TestTimeEntriesCreateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/timeEntries/123")
	mock.Set(OpTimeEntriesCreate, &fiken.CreateTimeEntryCreated{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesCreate(context.Background(), TimeEntriesCreateIn{
		Company: "acme",
		Body:    &fiken.TimeEntryRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestTimeEntriesCreatePaymentRequired confirms a 402 from the paid-
// tier time-tracking endpoint maps through MapErr →
// CodePaymentRequired.
func TestTimeEntriesCreatePaymentRequired(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpTimeEntriesCreate, 402, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesCreate(context.Background(), TimeEntriesCreateIn{
		Company: "acme",
		Body:    &fiken.TimeEntryRequest{},
	})
	if res.Error == nil || res.Error.Code != CodePaymentRequired {
		t.Fatalf("expected CodePaymentRequired, got %+v", res.Error)
	}
}

func TestTimeEntriesUpdateAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/timeEntries/7")
	mock.Set(OpTimeEntriesUpdate, &fiken.UpdateTimeEntryOK{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesUpdate(context.Background(), TimeEntriesUpdateIn{
		Company: "acme", TimeEntryID: 7,
		Body: &fiken.UpdateTimeEntryRequest{},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.TimeEntryID != 7 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

func TestTimeEntriesUpdateValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TimeEntriesUpdate(context.Background(), TimeEntriesUpdateIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TimeEntriesUpdate(context.Background(), TimeEntriesUpdateIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing time_entry_id: want validation error, got %+v", got)
	}
	if got := c.TimeEntriesUpdate(context.Background(), TimeEntriesUpdateIn{Company: "acme", TimeEntryID: 1}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
}

func TestTimeEntriesDeleteAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	// The DeleteTimeEntry server hook needs an explicit NoContent
	// response — the default nil DeleteTimeEntryRes can't be encoded.
	mock.Set(OpTimeEntriesDelete, &fiken.DeleteTimeEntryNoContent{})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesDelete(context.Background(), TimeEntriesDeleteIn{Company: "acme", TimeEntryID: 7})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("unexpected nil ok")
	}
}

func TestTimeEntriesDeleteValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TimeEntriesDelete(context.Background(), TimeEntriesDeleteIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TimeEntriesDelete(context.Background(), TimeEntriesDeleteIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing time_entry_id: want validation error, got %+v", got)
	}
}

func TestTimeEntriesInvoiceDraftValidation(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TimeEntriesInvoiceDraftFromTimes(context.Background(), TimeEntriesInvoiceDraftFromTimesIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TimeEntriesInvoiceDraftFromTimes(context.Background(), TimeEntriesInvoiceDraftFromTimesIn{Company: "acme"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing body: want validation error, got %+v", got)
	}
	if got := c.TimeEntriesInvoiceDraftFromTimes(context.Background(), TimeEntriesInvoiceDraftFromTimesIn{
		Company: "acme",
		Body:    &fiken.TimeEntryInvoiceDraftRequest{},
	}); got.Error == nil || got.Error.Code != CodeValidation {
		t.Fatalf("empty timeEntryIds: want validation error, got %+v", got)
	}
}

// TestTimeEntriesInvoiceDraftAgainstMock asserts the draft id flows
// from both the response body and the Location header.
func TestTimeEntriesInvoiceDraftAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/drafts/777")
	mock.Set(OpTimeEntriesInvoiceDraftFromTimes, &fiken.InvoiceishDraftResultHeaders{
		Location: fiken.OptURI{Value: *loc, Set: true},
		Response: fiken.InvoiceishDraftResult{
			DraftId: fiken.OptInt64{Value: 777, Set: true},
		},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesInvoiceDraftFromTimes(context.Background(), TimeEntriesInvoiceDraftFromTimesIn{
		Company: "acme",
		Body: &fiken.TimeEntryInvoiceDraftRequest{
			TimeEntryIds:     []int64{1, 2, 3},
			CustomerId:       42,
			DaysUntilDueDate: 14,
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 777 || res.Ok.Location != loc.String() {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}

// TestTimeEntriesInvoiceDraftFallbackFromLocation confirms the draft
// id falls back to the trailing path segment of Location when the
// response body does not echo it.
func TestTimeEntriesInvoiceDraftFallbackFromLocation(t *testing.T) {
	mock := startMockForTest(t)
	loc, _ := url.Parse("https://api.fiken.no/api/v2/companies/acme/invoices/drafts/9001")
	mock.Set(OpTimeEntriesInvoiceDraftFromTimes, &fiken.InvoiceishDraftResultHeaders{
		Location: fiken.OptURI{Value: *loc, Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesInvoiceDraftFromTimes(context.Background(), TimeEntriesInvoiceDraftFromTimesIn{
		Company: "acme",
		Body: &fiken.TimeEntryInvoiceDraftRequest{
			TimeEntryIds: []int64{1},
			CustomerId:   42,
		},
	})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.DraftID != 9001 {
		t.Fatalf("expected DraftID 9001 from Location fallback, got %+v", res.Ok)
	}
}

// TestTimeEntriesInvoiceDraftPaymentRequired confirms a 402 from the
// paid-tier endpoint maps through MapErr → CodePaymentRequired.
func TestTimeEntriesInvoiceDraftPaymentRequired(t *testing.T) {
	mock := startMockForTest(t)
	mock.SetError(OpTimeEntriesInvoiceDraftFromTimes, 402, []byte(`{"validationErrors":[]}`))

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeEntriesInvoiceDraftFromTimes(context.Background(), TimeEntriesInvoiceDraftFromTimesIn{
		Company: "acme",
		Body: &fiken.TimeEntryInvoiceDraftRequest{
			TimeEntryIds: []int64{1},
			CustomerId:   42,
		},
	})
	if res.Error == nil || res.Error.Code != CodePaymentRequired {
		t.Fatalf("expected CodePaymentRequired, got %+v", res.Error)
	}
}
