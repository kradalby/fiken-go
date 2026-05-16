package ops

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

// TestUserGetAgainstMock exercises the default empty-Userinfo path —
// the mock returns the registered override or a zero Userinfo when no
// override is set.
func TestUserGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.UserGet(context.Background(), UserGetIn{})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("nil Ok")
	}
	if res.Ok.Name != "" || res.Ok.Email != "" {
		t.Fatalf("default should be zero, got %+v", res.Ok)
	}
}

// TestUserGetAgainstMockOverride asserts the success-override flows
// back through the translation layer (Opt unwrap → bare strings).
func TestUserGetAgainstMockOverride(t *testing.T) {
	mock := startMockForTest(t)
	mock.Set(OpUserGet, &fiken.Userinfo{
		Name:  fiken.OptString{Value: "Test Testesen", Set: true},
		Email: fiken.OptString{Value: "test@fiken.no", Set: true},
	})

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.UserGet(context.Background(), UserGetIn{})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Name != "Test Testesen" || res.Ok.Email != "test@fiken.no" {
		t.Fatalf("translation mismatch: %+v", res.Ok)
	}
}
