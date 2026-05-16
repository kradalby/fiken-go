package ops

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestTimeUsersListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeUsersList(context.Background(), TimeUsersListIn{})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
}

func TestTimeUsersGetMissingArgs(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.TimeUsersGet(context.Background(), TimeUsersGetIn{}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing company: want validation error, got %+v", got)
	}
	if got := c.TimeUsersGet(context.Background(), TimeUsersGetIn{Company: "x"}); got.Error == nil ||
		got.Error.Code != CodeValidation {
		t.Fatalf("missing time_user_id: want validation error, got %+v", got)
	}
}

// TestTimeUsersListAgainstMock exercises the default empty-list path.
func TestTimeUsersListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeUsersList(context.Background(), TimeUsersListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %+v", res.Ok)
	}
}

// TestTimeUsersListAgainstMockOverride asserts translation.
func TestTimeUsersListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpTimeUsersList, &fiken.GetTimeUsersOKHeaders{
		Response: []fiken.TimeUserResult{{
			TimeUserId: fiken.OptInt64{Value: 3, Set: true},
			Name:       fiken.OptString{Value: "Olav Olsen", Set: true},
			Email:      fiken.OptString{Value: "olav@example.com", Set: true},
		}},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeUsersList(context.Background(), TimeUsersListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", res.Ok)
	}
	got := res.Ok.Items[0]
	if got.TimeUserID != 3 || got.Name != "Olav Olsen" || got.Email != "olav@example.com" {
		t.Fatalf("translation mismatch: %+v", got)
	}
}

// TestTimeUsersGetAgainstMock asserts the single-resource happy path.
func TestTimeUsersGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpTimeUsersGet, &fiken.TimeUserResult{
		TimeUserId: fiken.OptInt64{Value: 5, Set: true},
		Name:       fiken.OptString{Value: "Kari Karlsen", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.TimeUsersGet(context.Background(), TimeUsersGetIn{Company: "acme", TimeUserID: 5})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.TimeUserID != 5 || res.Ok.Name != "Kari Karlsen" {
		t.Fatalf("unexpected: %+v", res.Ok)
	}
}
