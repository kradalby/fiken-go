package ops

import "fmt"

// Result is the universal return envelope for every ops operation.
// Both Ok and Error are non-omitempty so the JSON discriminator is
// always present: success has `"error": null`, failure has
// `"ok": null`.
type Result[T any] struct {
	Ok    *T     `json:"ok"`
	Error *Error `json:"error"`
}

// Error is the canonical error shape returned by every ops operation.
// Code is the stable machine-readable identifier; Message is stable
// EN prose (NOT translated — JSON bytes stay locale-stable). Table
// renderer translates by Code → `error.<code>` i18n key.
type Error struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	HTTPStatus int            `json:"http_status,omitempty"`
	Op         string         `json:"op,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return "<nil ops.Error>"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Ok is a convenience constructor for a successful Result[T].
func Ok[T any](v T) Result[T] { return Result[T]{Ok: &v} }

// Err is a convenience constructor for a failed Result[T].
func Err[T any](e *Error) Result[T] { return Result[T]{Error: e} }
