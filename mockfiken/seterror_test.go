package mockfiken

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/ops"
)

// TestSetErrorStatusPropagates asserts that the status passed to
// SetError surfaces as the HTTP status code in ops.Error, going
// through ops + ogen + the mock's custom ErrorHandler.
//
// Before the A4 fix this test would observe HTTPStatus=500 regardless
// of the requested code, because handlers wrapped the override in
// fmt.Errorf and ogen's default handler mapped unknown errors to 500.
func TestSetErrorStatusPropagates(t *testing.T) {
	t.Parallel()

	mock := New(t)
	mock.SetError(ops.OpContactsList, 404, []byte(`{"validationErrors":[]}`))

	c, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}

	res := c.ContactsList(context.Background(), ops.ContactsListIn{Company: "acme"})
	if res.Error == nil {
		t.Fatalf("want error, got ok=%+v", res.Ok)
	}
	if res.Error.HTTPStatus != 404 {
		t.Errorf("HTTPStatus=%d want 404", res.Error.HTTPStatus)
	}
	if res.Error.Code != ops.CodeNotFound {
		t.Errorf("Code=%q want %q", res.Error.Code, ops.CodeNotFound)
	}
}

// TestSetErrorRateLimited covers a second status to ensure the
// mapping table is exercised end-to-end, not just the 404 case.
func TestSetErrorRateLimited(t *testing.T) {
	t.Parallel()

	mock := New(t)
	mock.SetError(ops.OpContactsList, 429, nil)

	c, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}

	res := c.ContactsList(context.Background(), ops.ContactsListIn{Company: "acme"})
	if res.Error == nil {
		t.Fatal("want error, got ok")
	}
	if res.Error.HTTPStatus != 429 {
		t.Errorf("HTTPStatus=%d want 429", res.Error.HTTPStatus)
	}
	if res.Error.Code != ops.CodeRateLimited {
		t.Errorf("Code=%q want %q", res.Error.Code, ops.CodeRateLimited)
	}
}
