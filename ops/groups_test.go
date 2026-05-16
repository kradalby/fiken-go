package ops

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

func TestGroupsListMissingCompany(t *testing.T) {
	c, err := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.GroupsList(context.Background(), GroupsListIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error, got %+v", res)
	}
	if res.Error.Op != OpGroupsList {
		t.Fatalf("Op: got %q want %q", res.Error.Op, OpGroupsList)
	}
}

// TestGroupsListAgainstMock exercises the default empty-list path.
func TestGroupsListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.GroupsList(context.Background(), GroupsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 0 {
		t.Fatalf("default mock should return empty items, got %+v", res.Ok)
	}
}

// TestGroupsListAgainstMockOverride asserts the success-override path.
func TestGroupsListAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpGroupsList, &fiken.GetGroupsOKHeaders{
		Response: []string{"VIP", "Standard"},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.GroupsList(context.Background(), GroupsListIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || len(res.Ok.Items) != 2 {
		t.Fatalf("expected 2 items, got %+v", res.Ok)
	}
	if res.Ok.Items[0].Name != "VIP" || res.Ok.Items[1].Name != "Standard" {
		t.Fatalf("translation mismatch: %+v", res.Ok.Items)
	}
}
