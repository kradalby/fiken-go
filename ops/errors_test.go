package ops

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"

	"github.com/ogen-go/ogen/validate"
)

func TestMapNetworkError(t *testing.T) {
	netErr := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	got := MapErr("companies_list", netErr)
	if got == nil {
		t.Fatal("MapErr returned nil")
	}
	if got.Code != "network" {
		t.Fatalf("got Code=%q want network", got.Code)
	}
	if got.Op != "companies_list" {
		t.Fatalf("got Op=%q want companies_list", got.Op)
	}
}

func TestMapContextCanceled(t *testing.T) {
	got := MapErr("x", context.Canceled)
	if got.Code != "cancelled" {
		t.Fatalf("got Code=%q want cancelled", got.Code)
	}
}

func TestMapURLErrorWrappingCanceled(t *testing.T) {
	wrapped := &url.Error{Op: "Get", URL: "http://x", Err: context.Canceled}
	got := MapErr("x", wrapped)
	if got.Code != "cancelled" {
		t.Fatalf("got Code=%q want cancelled", got.Code)
	}
}

func TestMapByStatus(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{400, "validation"},
		{401, "auth_invalid"},
		{402, "payment_required"},
		{403, "auth_forbidden"},
		{404, "not_found"},
		{409, "conflict"},
		{422, "validation"},
		{429, "rate_limited"},
		{500, "server_error"},
		{503, "server_error"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			got := MapStatus("x", c.status, "msg", "req-1", nil)
			if got.Code != c.want {
				t.Fatalf("status %d → Code=%q want %q", c.status, got.Code, c.want)
			}
			if got.HTTPStatus != c.status {
				t.Fatalf("HTTPStatus=%d want %d", got.HTTPStatus, c.status)
			}
			if got.RequestID != "req-1" {
				t.Fatalf("RequestID=%q want req-1", got.RequestID)
			}
		})
	}
}

func TestMapErrUnexpectedStatus(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		want   string
		status int
	}{
		{
			name:   "typed_404",
			err:    &validate.UnexpectedStatusCodeError{StatusCode: 404},
			want:   CodeNotFound,
			status: 404,
		},
		{
			name:   "typed_402",
			err:    &validate.UnexpectedStatusCodeError{StatusCode: 402},
			want:   CodePaymentRequired,
			status: 402,
		},
		{
			name:   "wrapped_403",
			err:    fmt.Errorf("decode response: %w", &validate.UnexpectedStatusCodeError{StatusCode: 403}),
			want:   CodeAuthForbidden,
			status: 403,
		},
		{
			name:   "string_fallback_500",
			err:    errors.New("decode response: unexpected status code: 500"),
			want:   CodeServerError,
			status: 500,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := MapErr("x", c.err)
			if got == nil {
				t.Fatal("MapErr returned nil")
			}
			if got.Code != c.want {
				t.Fatalf("Code=%q want %q", got.Code, c.want)
			}
			if got.HTTPStatus != c.status {
				t.Fatalf("HTTPStatus=%d want %d", got.HTTPStatus, c.status)
			}
		})
	}
}
