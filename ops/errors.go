package ops

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/ogen-go/ogen/validate"
)

// Stable Code constants. Match the set declared in the design spec.
const (
	CodeAuthMissing       = "auth_missing"
	CodeAuthInvalid       = "auth_invalid"
	CodeAuthForbidden     = "auth_forbidden"
	CodeNotFound          = "not_found"
	CodeValidation        = "validation"
	CodeConflict          = "conflict"
	CodePaymentRequired   = "payment_required"
	CodeRateLimited       = "rate_limited"
	CodeServerError       = "server_error"
	CodeNetwork           = "network"
	CodeCancelled         = "cancelled"
	CodeInternal          = "internal"
	CodeReadOnlyViolation = "read_only_violation"
)

// MapErr converts a raw Go error into a stable ops.Error.
// context-derived errors → cancelled; network errors → network;
// ogen unexpected-status errors → routed via MapStatus;
// everything else falls through to internal.
func MapErr(op string, err error) *Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return &Error{Code: CodeCancelled, Message: err.Error(), Op: op}
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return &Error{Code: CodeNetwork, Message: err.Error(), Op: op}
	}
	if status := extractOgenStatus(err); status > 0 {
		return MapStatus(op, status, err.Error(), "", nil)
	}
	return &Error{Code: CodeInternal, Message: err.Error(), Op: op}
}

// extractOgenStatus tries to recover the HTTP status code from an
// ogen-generated unexpected-status-code error. Returns 0 if the
// error isn't recognized.
func extractOgenStatus(err error) int {
	var uErr *validate.UnexpectedStatusCodeError
	if errors.As(err, &uErr) && uErr != nil {
		return uErr.StatusCode
	}
	// Fallback: parse "unexpected status code: NNN" from the message,
	// in case the typed error has been flattened by a wrapper.
	msg := err.Error()
	const prefix = "unexpected status code: "
	idx := strings.Index(msg, prefix)
	if idx < 0 {
		return 0
	}
	var status int
	if _, perr := fmt.Sscanf(msg[idx+len(prefix):], "%d", &status); perr != nil {
		return 0
	}
	return status
}

// MapStatus converts an HTTP status code into a stable ops.Error.
func MapStatus(op string, status int, message, requestID string, details map[string]any) *Error {
	var code string
	switch status {
	case 400, 422:
		code = CodeValidation
	case 401:
		code = CodeAuthInvalid
	case 402:
		code = CodePaymentRequired
	case 403:
		code = CodeAuthForbidden
	case 404:
		code = CodeNotFound
	case 409:
		code = CodeConflict
	case 429:
		code = CodeRateLimited
	default:
		if status >= 500 && status <= 599 {
			code = CodeServerError
		} else {
			code = CodeInternal
		}
	}
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Op:         op,
		RequestID:  requestID,
		Details:    details,
	}
}
