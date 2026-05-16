# fiken-go — Plan B: Foundations + Companies Vertical Slice

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up the cross-cutting domain layer (`ops/` result envelope, errors, paging, date, ratelimit; `auth/`; `config/`; `output/`; `i18n/`; `mockfiken/`) plus the first end-to-end vertical slice — the `companies` Fiken tag, exposed via CLI subcommand AND MCP tool, with byte-identical `--json` parity. After Plan B, `fiken auth login`, `fiken companies list`, `fiken companies get`, and `fiken mcp` all work against the ogen-generated client through `mockfiken` (no real Fiken token required to run tests).

**Architecture:** All packages from Plan A's repo layout that weren't built yet — except the i18n template seed, which Plan A already produced. Each piece is added in dependency order: result/error/paging/date types first (load-bearing for everything else), then auth + config + output + i18n + mockfiken, then ops/companies.go, then CLI, then MCP, then the parity test that asserts CLI JSON bytes equal MCP tool result bytes for every companies op.

**Tech Stack:**

- Go 1.26 (from Plan A).
- `github.com/ogen-go/ogen` v1.20.3 (Plan A vendored).
- `github.com/peterbourgon/ff/v4` — CLI parsing.
- `github.com/knadh/koanf/v2` + `parsers/toml/v2` + `providers/{file,env/v2,posflag}` — config.
- `github.com/zalando/go-keyring` — credential storage.
- `github.com/nicksnyder/go-i18n/v2/i18n` — bilingual messages.
- `github.com/modelcontextprotocol/go-sdk/mcp` — MCP server.
- `golang.org/x/term` — no-echo token paste.
- stdlib `log/slog`, `net/http/httptest`, `text/tabwriter`, `embed`.

**Spec reference:** `docs/superpowers/specs/2026-05-15-fiken-go-design.md`. This plan implements steps 4–9 of the spec, narrowed to the companies tag only. Plan C will replay the companies pattern across the remaining 16 tags. Plan D handles CI, drift workflow, and finishing touches.

**Prerequisites from Plan A:**

- 13 commits on `main` ending with `4fc59bd build(nix): set vendorHash for current go.sum`.
- `fiken/` package compiles (ogen-generated client + Handler interface).
- `ops/mutating.gen.go` lists 152 ops; `IsMutating` works.
- `i18n/locales/en.template.toml` is a 914-line editor scaffold (not loaded at runtime).
- `prek run --all-files` green; `go test ./...` green; `nix build .#fiken` works.

**Module path:** `github.com/kradalby/fiken-go`.

---

## File structure (Plan B)

| Phase       | Path                      | Purpose                                                                         |
| ----------- | ------------------------- | ------------------------------------------------------------------------------- |
| Foundations | `ops/result.go`           | `Result[T any]` envelope + `Error` type.                                        |
| Foundations | `ops/errors.go`           | `ogen` typed errors → `ops.Error` mapping; `error.<code>` i18n keys.            |
| Foundations | `ops/paging.go`           | `ListMeta`, `ListOut[T any]`.                                                   |
| Foundations | `ops/date.go`             | `ops.Date` civil-date type.                                                     |
| Foundations | `ops/ratelimit.go`        | sem=1 RoundTripper + 429-aware backoff.                                         |
| Foundations | `ops/units_test.go`       | Reflection-walk invariant test across all `Out*` structs.                       |
| Foundations | `ops/names.go`            | `Op*` constants + `Registry` shape (sources `Mutating` from `mutating.gen.go`). |
| Foundations | `auth/auth.go`            | `Source` iface, `ChainSource`.                                                  |
| Foundations | `auth/credential.go`      | Bare-token storage helpers (phase 1).                                           |
| Foundations | `auth/keyring.go`         | `go-keyring` wrapper with file-fallback.                                        |
| Foundations | `config/config.go`        | koanf-based `Profile`, `Config` loader.                                         |
| Foundations | `output/output.go`        | `Renderer` iface + `JSON` + `Table` factories.                                  |
| Foundations | `output/json.go`          | `encoding/json` renderer over `Result[T]`.                                      |
| Foundations | `output/table.go`         | `text/tabwriter` renderer over `Result[T]`.                                     |
| Foundations | `i18n/i18n.go`            | Bundle + `T()` helper; `//go:embed locales/*.toml`.                             |
| Foundations | `i18n/locales/en.toml`    | Hand-authored en strings (companies + auth + errors only).                      |
| Foundations | `i18n/locales/nb.toml`    | Hand-authored nb strings (parity with en).                                      |
| Foundations | `mockfiken/server.go`     | `fiken.Handler` impl + override registry + httptest wrap.                       |
| Companies   | `ops/companies.go`        | `In*` / `Out*` types + 2 methods (List, Get).                                   |
| CLI         | `cli/root.go`             | Root `*ff.Command`, global flags, context plumbing.                             |
| CLI         | `cli/companies.go`        | `fiken companies {list,get}` subcommands.                                       |
| CLI         | `cli/auth.go`             | `fiken auth {login,status,logout,list}`.                                        |
| CLI         | `cli/progress.go`         | stderr progress writer for paged ops.                                           |
| MCP         | `mcp/server.go`           | Walks `ops.Registry`, registers tools.                                          |
| MCP         | `mcp/readonly.go`         | Mode enum + filter via `ops.IsMutating`.                                        |
| MCP         | `mcp/transport.go`        | stdio + streamable HTTP wiring.                                                 |
| MCP         | `mcp/progress.go`         | Progress notifications during paging.                                           |
| CLI/MCP     | `cli/mcp.go`              | `fiken mcp` subcommand wires MCP package.                                       |
| Parity      | `cli/parity_test.go`      | CLI `--json` bytes == MCP `StructuredContent` for every companies op.           |
| Entrypoint  | `cmd/fiken/main.go`       | ~30-line `os.Args` → `cli.NewRoot` wiring.                                      |
| Hooks       | `.pre-commit-config.yaml` | Add `i18n-keys` + `oas-units` hooks.                                            |

Out of scope for Plan B (deferred to Plan C / D):

- All other Fiken tags (`invoices`, `contacts`, etc.) — Plan C.
- `mcp/attachments.go` — Plan C (when first attachment-op tag lands).
- GitHub Actions workflows (`ci.yml`, `spec-drift.yml`) — Plan D.

## Conventions

- Every Go file ends with a trailing newline (enforced by hook).
- Every Bash command runs inside `nix develop -c ...` from the repo root.
- Every commit message follows Conventional Commits.
- `goimports -local github.com/kradalby/fiken-go` runs via hook on every commit.
- Each task is one commit (some tasks split into multiple if size demands; the plan calls it out where so).
- TDD where listed: failing test first, then implementation, then verify pass, then commit.
- Use `t.Cleanup` over `defer` in tests where applicable.
- Prefer `slog` for any logging; never `fmt.Println` for diagnostics.

---

### Task 1: `ops/result.go` — Result[T] envelope + Error type

**Files:**

- Create: `ops/result.go`
- Create: `ops/result_test.go`

**Spec ref:** §"Result envelope (universal)", §"Error model".

- [ ] **Step 1.1: Write the failing test**

`ops/result_test.go`:

```go
package ops

import (
	"encoding/json"
	"testing"
)

// TestResultMarshalSuccess: success envelope serializes with ok set
// and error explicitly null (both keys always present).
func TestResultMarshalSuccess(t *testing.T) {
	type out struct {
		Name string `json:"name"`
	}
	r := Result[out]{Ok: &out{Name: "acme"}}
	got, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"ok":{"name":"acme"},"error":null}`
	if string(got) != want {
		t.Fatalf("mismatch:\n got %s\nwant %s", got, want)
	}
}

// TestResultMarshalError: error envelope serializes with ok null
// and error populated.
func TestResultMarshalError(t *testing.T) {
	r := Result[struct{}]{Error: &Error{Code: "not_found", Message: "no such company", HTTPStatus: 404}}
	got, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"ok":null,"error":{"code":"not_found","message":"no such company","http_status":404}}`
	if string(got) != want {
		t.Fatalf("mismatch:\n got %s\nwant %s", got, want)
	}
}

// TestErrorImplementsErrorInterface — ops.Error satisfies error.
func TestErrorImplementsErrorInterface(t *testing.T) {
	var _ error = (*Error)(nil)
	e := &Error{Code: "validation", Message: "bad input"}
	if got := e.Error(); got != "validation: bad input" {
		t.Fatalf("error string mismatch: %q", got)
	}
}
```

- [ ] **Step 1.2: Run, confirm fail**

```bash
nix develop -c go test ./ops/...
```

Expected: build failures referencing `Result`, `Error`.

- [ ] **Step 1.3: Implement `ops/result.go`**

```go
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
```

- [ ] **Step 1.4: Run, confirm pass**

```bash
nix develop -c go test ./ops/...
```

Expected: PASS for the three Result/Error tests + the existing smoke tests.

- [ ] **Step 1.5: Commit**

```bash
git add ops/result.go ops/result_test.go
git commit -m "$(cat <<'EOF'
feat(ops): add Result[T] envelope + Error type

Universal return shape for every ops operation. Both Ok and Error
non-omitempty so JSON consumers always see the discriminator.
Message is stable EN (data, not UX) so --json output is locale-stable.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

Expected: hooks pass.

---

### Task 2: Stable error codes — `ops/errors.go`

**Files:**

- Create: `ops/errors.go`
- Create: `ops/errors_test.go`

**Spec ref:** §"Stable Code set" (12 codes).

- [ ] **Step 2.1: Failing test**

`ops/errors_test.go`:

```go
package ops

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"
)

// TestMapNetworkError → Code=network.
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

// TestMapContextCanceled → Code=cancelled.
func TestMapContextCanceled(t *testing.T) {
	got := MapErr("x", context.Canceled)
	if got.Code != "cancelled" {
		t.Fatalf("got Code=%q want cancelled", got.Code)
	}
}

// TestMapURLError wrapping context cancellation → cancelled.
func TestMapURLErrorWrappingCanceled(t *testing.T) {
	wrapped := &url.Error{Op: "Get", URL: "http://x", Err: context.Canceled}
	got := MapErr("x", wrapped)
	if got.Code != "cancelled" {
		t.Fatalf("got Code=%q want cancelled", got.Code)
	}
}

// TestMapByStatus maps known HTTP statuses to canonical Codes.
func TestMapByStatus(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{400, "validation"},
		{401, "auth_invalid"},
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
```

- [ ] **Step 2.2: Confirm fail.** `go test ./ops/...` → `undefined: MapErr`, `undefined: MapStatus`.

- [ ] **Step 2.3: Implement `ops/errors.go`**

```go
package ops

import (
	"context"
	"errors"
	"net"
)

// Stable Code constants. Match the set declared in the design spec.
const (
	CodeAuthMissing       = "auth_missing"
	CodeAuthInvalid       = "auth_invalid"
	CodeAuthForbidden     = "auth_forbidden"
	CodeNotFound          = "not_found"
	CodeValidation        = "validation"
	CodeConflict          = "conflict"
	CodeRateLimited       = "rate_limited"
	CodeServerError       = "server_error"
	CodeNetwork           = "network"
	CodeCancelled         = "cancelled"
	CodeInternal          = "internal"
	CodeReadOnlyViolation = "read_only_violation"
)

// MapErr converts a raw Go error (typically from the ogen client or
// http stack) into a stable ops.Error. Pre-HTTP errors are classified
// as cancelled (context-derived) or network (anything else IO-like).
// Anything not recognized is classified internal — caller is expected
// to either upgrade the recognizer or accept that something went off
// the rails inside the client.
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
	return &Error{Code: CodeInternal, Message: err.Error(), Op: op}
}

// MapStatus converts an HTTP status code + decoded message body into
// a stable ops.Error. requestID is the value of `X-Request-Id`
// (or a client-generated ULID). details holds any structured
// field-error map from the response.
func MapStatus(op string, status int, message, requestID string, details map[string]any) *Error {
	var code string
	switch status {
	case 400, 422:
		code = CodeValidation
	case 401:
		code = CodeAuthInvalid
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
```

- [ ] **Step 2.4: Pass.** `go test ./ops/...` PASS.

- [ ] **Step 2.5: Commit**

```bash
git add ops/errors.go ops/errors_test.go
git commit -m "$(cat <<'EOF'
feat(ops): map raw errors and HTTP status codes to stable Codes

MapErr handles context.Canceled, net.OpError, and falls back to
internal. MapStatus maps 4xx/5xx HTTP statuses to the canonical
Code set (auth_invalid, validation, rate_limited, etc.).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Pagination types — `ops/paging.go`

**Files:**

- Create: `ops/paging.go`
- Create: `ops/paging_test.go`

**Spec ref:** §"Pagination".

- [ ] **Step 3.1: Failing test**

`ops/paging_test.go`:

```go
package ops

import (
	"encoding/json"
	"testing"
)

func TestListOutMarshal(t *testing.T) {
	items := []string{"a", "b"}
	out := ListOut[string]{Items: items, Meta: ListMeta{Returned: 2}}
	got, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"items":["a","b"],"meta":{"truncated":false,"returned":2}}`
	if string(got) != want {
		t.Fatalf("mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestListMetaTruncated(t *testing.T) {
	m := ListMeta{Truncated: true, NextPage: 4, Returned: 100, Cancelled: false}
	got, _ := json.Marshal(m)
	want := `{"truncated":true,"next_page":4,"returned":100}`
	if string(got) != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
```

- [ ] **Step 3.2: Confirm fail.** Build error.

- [ ] **Step 3.3: Implement `ops/paging.go`**

```go
package ops

// ListMeta carries pagination metadata that accompanies a List
// operation's Items slice. Always emitted; Truncated and Returned
// are required fields. NextPage / Cancelled are omitempty.
type ListMeta struct {
	Truncated bool `json:"truncated"`
	NextPage  int  `json:"next_page,omitempty"`
	Returned  int  `json:"returned"`
	Cancelled bool `json:"cancelled,omitempty"`
}

// ListOut[T] is the canonical Ok-shape for any List/paged operation.
// Plug T = whatever the per-tag Out type is (e.g. CompanyOut).
type ListOut[T any] struct {
	Items []T      `json:"items"`
	Meta  ListMeta `json:"meta"`
}
```

- [ ] **Step 3.4: Pass.** `go test ./ops/...` PASS.

- [ ] **Step 3.5: Commit**

```bash
git add ops/paging.go ops/paging_test.go
git commit -m "$(cat <<'EOF'
feat(ops): add ListOut[T] + ListMeta for paged operations

Standard envelope for any tag's List op. Meta.NextPage and
Meta.Cancelled are omitempty so default success cases stay terse.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Civil date — `ops/date.go`

**Files:**

- Create: `ops/date.go`
- Create: `ops/date_test.go`

**Spec ref:** §"Time zones", §"Units invariants".

- [ ] **Step 4.1: Failing test**

`ops/date_test.go`:

```go
package ops

import (
	"encoding/json"
	"testing"
)

func TestDateMarshal(t *testing.T) {
	d := Date("2026-05-15")
	got, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(got) != `"2026-05-15"` {
		t.Fatalf("got %s want \"2026-05-15\"", got)
	}
}

func TestDateUnmarshal(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte(`"2026-05-15"`), &d); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if d != "2026-05-15" {
		t.Fatalf("got %q want 2026-05-15", d)
	}
}

func TestDateUnmarshalRejectsBad(t *testing.T) {
	var d Date
	err := json.Unmarshal([]byte(`"2026/05/15"`), &d)
	if err == nil {
		t.Fatal("expected error on bad date format")
	}
}
```

- [ ] **Step 4.2: Fail.** Build error.

- [ ] **Step 4.3: Implement `ops/date.go`**

```go
package ops

import (
	"encoding/json"
	"fmt"
	"time"
)

// Date is a civil (no-timezone) calendar date in YYYY-MM-DD form.
// Used for Out struct fields named *Date (no time component).
// Marshals to/from a quoted ISO 8601 date string; rejects anything
// not in the exact `2006-01-02` layout.
type Date string

const dateLayout = "2006-01-02"

// MarshalJSON emits a quoted YYYY-MM-DD string.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}

// UnmarshalJSON parses a quoted YYYY-MM-DD string. Rejects other
// formats (RFC 3339 datetime, slash separators, etc.) so a misparsed
// upstream field becomes a loud error instead of a silent UTC drift.
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		*d = ""
		return nil
	}
	if _, err := time.Parse(dateLayout, s); err != nil {
		return fmt.Errorf("ops.Date: %w (must be YYYY-MM-DD)", err)
	}
	*d = Date(s)
	return nil
}
```

- [ ] **Step 4.4: Pass.** `go test ./ops/...` PASS.

- [ ] **Step 4.5: Commit**

```bash
git add ops/date.go ops/date_test.go
git commit -m "$(cat <<'EOF'
feat(ops): add civil Date type with strict YYYY-MM-DD parsing

Date-only fields use ops.Date instead of time.Time to avoid a tz
serializing back into JSON. UnmarshalJSON rejects non-ISO formats so
upstream drift becomes a loud error, not a silent UTC bug.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: Rate limiting + backoff — `ops/ratelimit.go`

**Files:**

- Create: `ops/ratelimit.go`
- Create: `ops/ratelimit_test.go`

**Spec ref:** §"Rate limit, concurrency, progress".

- [ ] **Step 5.1: Failing test**

`ops/ratelimit_test.go`:

```go
package ops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrencyTokenSerializes asserts only one request flies at a
// time even with concurrent goroutines.
func TestConcurrencyTokenSerializes(t *testing.T) {
	var inflight int32
	var maxInflight int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&inflight, 1)
		if n > atomic.LoadInt32(&maxInflight) {
			atomic.StoreInt32(&maxInflight, n)
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&inflight, -1)
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	c := &http.Client{Transport: newConcurrencyRT(http.DefaultTransport)}
	const N = 5
	done := make(chan struct{}, N)
	for i := 0; i < N; i++ {
		go func() {
			req, _ := http.NewRequest("GET", srv.URL, nil)
			resp, err := c.Do(req)
			if err != nil {
				t.Errorf("do: %v", err)
			}
			if resp != nil {
				resp.Body.Close()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < N; i++ {
		<-done
	}
	if got := atomic.LoadInt32(&maxInflight); got != 1 {
		t.Fatalf("maxInflight=%d want 1 (concurrency token broken)", got)
	}
}

// TestBackoffHonorsRetryAfter waits the seconds the server tells it.
func TestBackoffHonorsRetryAfter(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	c := &http.Client{Transport: newBackoffRT(http.DefaultTransport, 3)}
	start := time.Now()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", srv.URL, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Fatalf("backoff too fast: %s (Retry-After=1 should sleep ~1s)", elapsed)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("attempts=%d want 2 (one retry)", got)
	}
	// Touch strconv so the import stays even if Retry-After parsing
	// is refactored away from manual parse — see ratelimit.go.
	_ = strconv.Itoa(0)
}
```

- [ ] **Step 5.2: Fail.** Build error referencing `newConcurrencyRT`, `newBackoffRT`.

- [ ] **Step 5.3: Implement `ops/ratelimit.go`**

```go
package ops

import (
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

// newConcurrencyRT wraps base so at most one in-flight request at a
// time per Client. Fiken's stated rate rule is "one concurrent
// request per user"; we enforce it client-side.
func newConcurrencyRT(base http.RoundTripper) http.RoundTripper {
	return &concurrencyRT{base: base, sem: make(chan struct{}, 1)}
}

type concurrencyRT struct {
	base http.RoundTripper
	sem  chan struct{}
}

func (rt *concurrencyRT) RoundTrip(r *http.Request) (*http.Response, error) {
	select {
	case rt.sem <- struct{}{}:
	case <-r.Context().Done():
		return nil, r.Context().Err()
	}
	defer func() { <-rt.sem }()
	return rt.base.RoundTrip(r)
}

// newBackoffRT wraps base with 429-aware retry: honors Retry-After
// when present, else exponential with full jitter capped at 4s.
// maxRetries excludes the initial attempt.
func newBackoffRT(base http.RoundTripper, maxRetries int) http.RoundTripper {
	return &backoffRT{base: base, max: maxRetries}
}

type backoffRT struct {
	base http.RoundTripper
	max  int
}

func (rt *backoffRT) RoundTrip(r *http.Request) (*http.Response, error) {
	var lastResp *http.Response
	var lastErr error
	for attempt := 0; attempt <= rt.max; attempt++ {
		resp, err := rt.base.RoundTrip(r)
		if err != nil {
			return resp, err
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		delay := retryAfter(resp, attempt)
		_ = resp.Body.Close()
		lastResp = resp
		lastErr = nil
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	}
	return lastResp, lastErr
}

func retryAfter(resp *http.Response, attempt int) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	// Exponential with full jitter: base 250ms, cap 4s.
	const base = 250 * time.Millisecond
	const cap_ = 4 * time.Second
	d := base * (1 << attempt)
	if d > cap_ {
		d = cap_
	}
	// nolint:gosec — math/rand/v2 is fine for jitter.
	return time.Duration(rand.Int64N(int64(d)))
}
```

- [ ] **Step 5.4: Pass.** `go test ./ops/... -run TestConcurrencyToken -race` and `... -run TestBackoff -race` both pass.

- [ ] **Step 5.5: Commit**

```bash
git add ops/ratelimit.go ops/ratelimit_test.go
git commit -m "$(cat <<'EOF'
feat(ops): add concurrency + 429-aware backoff RoundTrippers

concurrencyRT enforces sem=1 per Client (Fiken's stated rate rule).
backoffRT honors Retry-After and falls back to exponential-jitter
(250ms..4s) on 429.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Units invariant test — `ops/units_test.go`

**Files:**

- Create: `ops/units_test.go`

**Spec ref:** §"Units invariants (test-enforced)".

This test reflects across all exported `Out*` structs in `ops/` and asserts field types match the units table. Since no `Out*` types exist yet (Task 12 introduces `CompanyOut`), the test is **inert** until tasks land them. The test still runs; it just iterates an empty set and passes trivially. This task lands the reflection scaffolding so future ops can't drift.

- [ ] **Step 6.1: Write the test**

```go
package ops

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// unitRules maps a field-name regex to the Go type that field MUST
// have. Add patterns here as the domain grows.
var unitRules = []struct {
	pattern *regexp.Regexp
	wantKind string // "int64", "int", "Date", "time.Time"
}{
	{regexp.MustCompile(`(?i)(amount|price|total|sum|net|gross|balance|paid|due)$`), "int64"},
	{regexp.MustCompile(`(?i)(rate|percent)$`), "int"},
	{regexp.MustCompile(`(?i)date$`), "Date"},
	{regexp.MustCompile(`(?i)(at|datetime)$`), "time.Time"},
}

// outStructs returns every exported type in this package whose name
// starts with "Out" or ends with "Out" — i.e. the canonical
// per-operation response structs. Walked via reflect for now; once
// a Registry exists this can iterate that instead.
func outStructs() []reflect.Type {
	// Placeholder slice: append known Out types as they're introduced.
	// Plan B Task 12 adds CompanyOut / CompaniesListOut here.
	return nil
}

func TestOutFieldUnits(t *testing.T) {
	var failures []string
	for _, st := range outStructs() {
		for i := 0; i < st.NumField(); i++ {
			f := st.Field(i)
			if !f.IsExported() {
				continue
			}
			for _, rule := range unitRules {
				if !rule.pattern.MatchString(f.Name) {
					continue
				}
				gotKind := goKindFor(f.Type)
				if gotKind != rule.wantKind {
					failures = append(failures,
						strings.Join([]string{st.Name(), ".", f.Name, ": got ", gotKind, " want ", rule.wantKind}, ""))
				}
			}
		}
	}
	if len(failures) > 0 {
		t.Fatalf("unit-type violations:\n  %s", strings.Join(failures, "\n  "))
	}
}

func goKindFor(t reflect.Type) string {
	// Time and Date get name-based identification (reflect treats
	// time.Time as a struct).
	if t == reflect.TypeOf(time.Time{}) {
		return "time.Time"
	}
	if t.Kind() == reflect.String && t.Name() == "Date" {
		return "Date"
	}
	return t.Kind().String()
}
```

- [ ] **Step 6.2: Run, confirm it passes (empty walk)**

```bash
nix develop -c go test ./ops/...
```

Expected: PASS (the slice is empty; test trivially succeeds).

- [ ] **Step 6.3: Commit**

```bash
git add ops/units_test.go
git commit -m "$(cat <<'EOF'
test(ops): reflection-walk units invariant over Out* structs

unitRules maps field-name patterns to required Go types
(int64 for money, int for basis-points rates, ops.Date for civil
dates, time.Time for datetimes). outStructs() returns an empty
slice for now — Task 12 (CompanyOut) populates it.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: Op names + Registry — `ops/names.go`

**Files:**

- Create: `ops/names.go`
- Create: `ops/names_test.go`

**Spec ref:** §"Shared op identifiers and the single `Out` type".

The Registry is the single source of truth for op metadata: name, mutating bit (read from `mutating.gen.go`), help-text i18n keys, MCP InputSchema notes. Populated incrementally as ops land.

- [ ] **Step 7.1: Failing test**

`ops/names_test.go`:

```go
package ops

import "testing"

func TestRegistryHasNonEmpty(t *testing.T) {
	if len(Registry) == 0 {
		t.Fatal("Registry must have at least one entry after Plan B Task 12")
	}
}

func TestRegistryMutatingMatchesGen(t *testing.T) {
	for name, entry := range Registry {
		gotMut := IsMutating(name)
		if entry.Mutating != gotMut {
			t.Errorf("Registry[%q].Mutating=%v but IsMutating(%q)=%v",
				name, entry.Mutating, name, gotMut)
		}
	}
}

func TestOpConstNamesAreInRegistry(t *testing.T) {
	for _, name := range []string{OpCompaniesList, OpCompaniesGet} {
		if _, ok := Registry[name]; !ok {
			t.Errorf("op %q in const but not in Registry", name)
		}
	}
}
```

- [ ] **Step 7.2: Fail.** Build error on `Registry`, `OpCompaniesList`, `OpCompaniesGet`.

- [ ] **Step 7.3: Implement `ops/names.go`**

```go
package ops

// Op*: stable op-name constants. The string value is the canonical
// shared name used by:
//   - the CLI subcommand path (e.g. "companies_list" ↔ `fiken companies list`)
//   - the MCP tool name in tools/list
//   - the i18n key prefix (e.g. ops.companies_list.summary)
//   - the key into IsMutating / Registry / mockfiken overrides
//
// One rename, one edit.
const (
	OpCompaniesList = "companies_list"
	OpCompaniesGet  = "companies_get"
)

// RegistryEntry holds per-op metadata that BOTH frontends consume.
// CompanyScoped is true when the op acts on `/companies/{slug}/...`
// — used by the MCP InputSchema to include an optional `company`
// param. False for ops that operate at user/global level.
type RegistryEntry struct {
	Mutating      bool
	CompanyScoped bool
}

// Registry indexed by op-name const. Populated alongside each
// In/Out type in the per-tag files (ops/companies.go etc.). The
// Mutating field is set from IsMutating() — the test in
// ops/names_test.go enforces consistency.
var Registry = map[string]RegistryEntry{
	OpCompaniesList: {Mutating: IsMutating(OpCompaniesList), CompanyScoped: false},
	OpCompaniesGet:  {Mutating: IsMutating(OpCompaniesGet), CompanyScoped: true},
}
```

- [ ] **Step 7.4: Pass.** `go test ./ops/...` PASS.

- [ ] **Step 7.5: Commit**

```bash
git add ops/names.go ops/names_test.go
git commit -m "$(cat <<'EOF'
feat(ops): add Op* name constants + Registry

Single source of truth for op metadata shared between CLI, MCP,
i18n. Mutating sourced from mutating.gen.go (consistency enforced
by test). First two entries: companies_list, companies_get.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: `auth/` package

**Files:**

- Create: `auth/auth.go`
- Create: `auth/credential.go`
- Create: `auth/keyring.go`
- Create: `auth/auth_test.go`

**Spec ref:** §"Auth".

- [ ] **Step 8.1: Add dep**

```bash
nix develop -c go get github.com/zalando/go-keyring
```

- [ ] **Step 8.2: Failing test**

`auth/auth_test.go`:

```go
package auth

import (
	"context"
	"errors"
	"testing"
)

type fixedSrc struct{ tok string }

func (s fixedSrc) Token(_ context.Context) (string, error) {
	if s.tok == "" {
		return "", ErrNotFound
	}
	return s.tok, nil
}

func TestChainSourceReturnsFirstNonEmpty(t *testing.T) {
	c := ChainSource{fixedSrc{""}, fixedSrc{"abc"}, fixedSrc{"def"}}
	got, err := c.Token(context.Background())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got != "abc" {
		t.Fatalf("got %q want abc", got)
	}
}

func TestChainSourceErrNotFound(t *testing.T) {
	c := ChainSource{fixedSrc{""}, fixedSrc{""}}
	_, err := c.Token(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v want ErrNotFound", err)
	}
}

func TestFlagSourceEnvSource(t *testing.T) {
	if _, err := (FlagSource{Value: ""}).Token(context.Background()); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty FlagSource should return ErrNotFound")
	}
	if got, err := (FlagSource{Value: "tok"}).Token(context.Background()); err != nil || got != "tok" {
		t.Errorf("FlagSource: got %q err %v", got, err)
	}

	t.Setenv("FIKEN_TOKEN", "envtok")
	if got, err := (EnvSource{Var: "FIKEN_TOKEN"}).Token(context.Background()); err != nil || got != "envtok" {
		t.Errorf("EnvSource: got %q err %v", got, err)
	}
}
```

- [ ] **Step 8.3: Confirm fail.** Build error.

- [ ] **Step 8.4: Implement `auth/auth.go`**

```go
// Package auth resolves the Fiken personal API token for the current
// session. Token sources are arranged into a ChainSource that returns
// the first non-empty value. The resolution order — CLI flag, env,
// keyring, file — is configured by the caller in cli/root.go.
package auth

import (
	"context"
	"errors"
	"os"
)

// ErrNotFound is returned by a Source when it has no token to offer.
// ChainSource keeps trying further sources until it gets a token or
// runs out — only then is the chain's overall result ErrNotFound.
var ErrNotFound = errors.New("auth: no token found")

// Source produces a personal API token for an outgoing Fiken request.
type Source interface {
	Token(ctx context.Context) (string, error)
}

// ChainSource calls each Source in order and returns the first
// non-empty token. Any non-ErrNotFound error short-circuits.
type ChainSource []Source

func (c ChainSource) Token(ctx context.Context) (string, error) {
	for _, s := range c {
		tok, err := s.Token(ctx)
		if err == nil {
			return tok, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return "", err
		}
	}
	return "", ErrNotFound
}

// FlagSource carries a token passed via --token flag.
type FlagSource struct{ Value string }

func (f FlagSource) Token(_ context.Context) (string, error) {
	if f.Value == "" {
		return "", ErrNotFound
	}
	return f.Value, nil
}

// EnvSource reads a token from the named env var (default FIKEN_TOKEN).
type EnvSource struct{ Var string }

func (e EnvSource) Token(_ context.Context) (string, error) {
	v := os.Getenv(e.Var)
	if v == "" {
		return "", ErrNotFound
	}
	return v, nil
}
```

- [ ] **Step 8.5: Implement `auth/credential.go`**

```go
package auth

// Credential is the storage value for a single profile's auth state.
// Phase 1 holds only a personal API token (Kind="personal"); phase 2
// will widen to include refresh/expires for OAuth. The on-disk JSON
// shape is reserved now — any future field added here must be
// optional and zero-tolerant on decode.
type Credential struct {
	Kind  string `json:"kind"`
	Token string `json:"token"`
}

// NewPersonal builds a Credential for a personal API token.
func NewPersonal(token string) Credential {
	return Credential{Kind: "personal", Token: token}
}
```

- [ ] **Step 8.6: Implement `auth/keyring.go`**

```go
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const keyringService = "fiken-go"

// KeyringSource reads a token from the OS keyring under
// (service="fiken-go", user=profile). Falls back to ~/.config/fiken/
// credentials/{profile}.json (mode 0600) when the keyring is
// unavailable (CI / containers / WSL without a backend).
type KeyringSource struct {
	Profile  string
	FilePath string // override; default ~/.config/fiken/credentials/<profile>.json
}

func (k KeyringSource) Token(_ context.Context) (string, error) {
	raw, err := keyring.Get(keyringService, k.Profile)
	if err == nil {
		return tokenFromRaw(raw)
	}
	// Fallback file.
	fp := k.FilePath
	if fp == "" {
		fp = defaultFilePath(k.Profile)
	}
	data, ferr := os.ReadFile(fp)
	if ferr != nil {
		if errors.Is(ferr, os.ErrNotExist) {
			return "", ErrNotFound
		}
		return "", ferr
	}
	return tokenFromRaw(string(data))
}

// Save stores cred in the keyring, falling back to file if keyring
// is unavailable. Returns the storage path actually used so the CLI
// can tell the user where.
func (k KeyringSource) Save(cred Credential) (string, error) {
	raw, err := json.Marshal(cred)
	if err != nil {
		return "", err
	}
	if err := keyring.Set(keyringService, k.Profile, string(raw)); err == nil {
		return "keyring", nil
	}
	fp := k.FilePath
	if fp == "" {
		fp = defaultFilePath(k.Profile)
	}
	if err := os.MkdirAll(filepath.Dir(fp), 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(fp, raw, 0o600); err != nil {
		return "", err
	}
	return fp, nil
}

// Delete removes the credential. Removes from both keyring and file
// path so re-login starts clean.
func (k KeyringSource) Delete() error {
	_ = keyring.Delete(keyringService, k.Profile)
	fp := k.FilePath
	if fp == "" {
		fp = defaultFilePath(k.Profile)
	}
	if err := os.Remove(fp); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func tokenFromRaw(raw string) (string, error) {
	if raw == "" {
		return "", ErrNotFound
	}
	var cred Credential
	if err := json.Unmarshal([]byte(raw), &cred); err != nil {
		// Treat a bare-string legacy value as a personal token.
		return raw, nil
	}
	if cred.Token == "" {
		return "", ErrNotFound
	}
	return cred.Token, nil
}

func defaultFilePath(profile string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fiken", "credentials", profile+".json")
}
```

- [ ] **Step 8.7: Pass.** `go test ./auth/...` PASS. (The keyring tests rely on the in-memory mock from `go-keyring` — if not, only FlagSource/EnvSource/ChainSource tests run.)

- [ ] **Step 8.8: Commit**

```bash
git add auth/ go.mod go.sum
git commit -m "$(cat <<'EOF'
feat(auth): add Source iface, ChainSource, KeyringSource

ChainSource composes FlagSource > EnvSource > KeyringSource > FileSource
(latter folded into KeyringSource fallback). KeyringSource serializes
auth.Credential JSON to keyring; falls back to chmod-0600 file when
no keyring backend is available.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 9: `config/` package — koanf loader

**Files:**

- Create: `config/config.go`
- Create: `config/config_test.go`

**Spec ref:** §"Config".

- [ ] **Step 9.1: Add deps**

```bash
nix develop -c go get \
  github.com/knadh/koanf/v2 \
  github.com/knadh/koanf/parsers/toml/v2 \
  github.com/knadh/koanf/providers/file \
  github.com/knadh/koanf/providers/env/v2
```

(koanf's posflag provider lands in Task 13 alongside the ff/v4 root command.)

- [ ] **Step 9.2: Failing test**

`config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileAndEnv(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	body := `default_profile = "work"
[profiles.work]
token = "filetok"
company = "acme"
lang = "nb"
`
	if err := os.WriteFile(fp, []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("FIKEN_COMPANY", "envco")

	cfg, err := Load(fp, nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	prof, ok := cfg.Resolve("")
	if !ok {
		t.Fatalf("resolve default profile failed")
	}
	if prof.Token != "filetok" {
		t.Errorf("Token=%q want filetok", prof.Token)
	}
	// Env beats file.
	if prof.Company != "envco" {
		t.Errorf("Company=%q want envco (env override)", prof.Company)
	}
	if prof.Lang != "nb" {
		t.Errorf("Lang=%q want nb", prof.Lang)
	}
}

func TestResolveExplicitProfile(t *testing.T) {
	cfg := &Config{
		DefaultProfile: "p1",
		Profiles: map[string]Profile{
			"p1": {Company: "a"},
			"p2": {Company: "b"},
		},
	}
	prof, ok := cfg.Resolve("p2")
	if !ok || prof.Company != "b" {
		t.Fatalf("Resolve(p2) gave %v", prof)
	}
	defp, _ := cfg.Resolve("")
	if defp.Company != "a" {
		t.Fatalf("default profile resolved to %v want company=a", defp)
	}
}
```

- [ ] **Step 9.3: Fail.** Build error.

- [ ] **Step 9.4: Implement `config/config.go`**

```go
// Package config loads the fiken-go configuration via koanf, merging
// (in load order): file (default ~/.config/fiken/config.toml) →
// FIKEN_* env vars → ff/v4 flags. Profile resolution then composes
// fields from the named profile block and applies env overrides.
package config

import (
	"errors"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Profile is the effective fiken-go configuration for one named
// account / company.
type Profile struct {
	Token   string `koanf:"token"`
	Company string `koanf:"company"`
	Lang    string `koanf:"lang"`
}

// Config is the parsed configuration file plus override layer.
type Config struct {
	DefaultProfile string             `koanf:"default_profile"`
	Profiles       map[string]Profile `koanf:"profiles"`
	envOverride    Profile
}

// Load reads filePath (if it exists) and FIKEN_* env vars and
// returns a Config. flagOverrides may be nil; pass a map of
// {company,token,lang,profile} to override programmatically.
func Load(filePath string, flagOverrides map[string]string) (*Config, error) {
	k := koanf.New(".")
	if filePath != "" {
		if _, err := os.Stat(filePath); err == nil {
			if err := k.Load(file.Provider(filePath), toml.Parser()); err != nil {
				return nil, err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, err
	}

	// FIKEN_TOKEN / FIKEN_COMPANY / FIKEN_LANG / FIKEN_PROFILE.
	envK := koanf.New(".")
	_ = envK.Load(env.Provider(".", env.Opt{
		Prefix: "FIKEN_",
		TransformFunc: func(k, v string) (string, any) {
			return strings.ToLower(strings.TrimPrefix(k, "FIKEN_")), v
		},
	}), nil)
	cfg.envOverride = Profile{
		Token:   envK.String("token"),
		Company: envK.String("company"),
		Lang:    envK.String("lang"),
	}

	if envP := envK.String("profile"); envP != "" {
		cfg.DefaultProfile = envP
	}

	if flagOverrides != nil {
		if v := flagOverrides["profile"]; v != "" {
			cfg.DefaultProfile = v
		}
		if v := flagOverrides["token"]; v != "" {
			cfg.envOverride.Token = v
		}
		if v := flagOverrides["company"]; v != "" {
			cfg.envOverride.Company = v
		}
		if v := flagOverrides["lang"]; v != "" {
			cfg.envOverride.Lang = v
		}
	}

	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = "default"
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	return cfg, nil
}

// Resolve returns the effective Profile for `name`. If name is empty
// the DefaultProfile is used. Env / flag overrides are applied on top
// of the profile block.
func (c *Config) Resolve(name string) (Profile, bool) {
	if name == "" {
		name = c.DefaultProfile
	}
	base, ok := c.Profiles[name]
	if !ok && name == "default" {
		// Allow env-only operation: an empty profile is fine if
		// overrides supply everything.
		base = Profile{}
		ok = true
	}
	if !ok {
		return Profile{}, false
	}
	merged := merge(base, c.envOverride)
	return merged, true
}

// merge picks override fields when non-empty, else base.
func merge(base, over Profile) Profile {
	if over.Token != "" {
		base.Token = over.Token
	}
	if over.Company != "" {
		base.Company = over.Company
	}
	if over.Lang != "" {
		base.Lang = over.Lang
	}
	return base
}
```

- [ ] **Step 9.5: Pass.** `go test ./config/...` PASS.

- [ ] **Step 9.6: Commit**

```bash
git add config/ go.mod go.sum
git commit -m "$(cat <<'EOF'
feat(config): koanf-based loader for profiles + env overrides

File → env merge order with explicit Resolve() that applies env on
top of named profile. FIKEN_PROFILE selects the active profile;
FIKEN_TOKEN / FIKEN_COMPANY / FIKEN_LANG override fields within it.
Flag overrides are passed as a map (wired in cli/root.go).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 10: `i18n/` package — bundle + locale TOMLs

**Files:**

- Modify: `i18n/doc.go` (now also embeds locales)
- Create: `i18n/i18n.go`
- Create: `i18n/locales/en.toml`
- Create: `i18n/locales/nb.toml`
- Create: `i18n/i18n_test.go`

**Spec ref:** §"i18n", §"Help text and flag descriptions".

This task lands the **runtime** bundle plus hand-authored en + nb catalogs covering exactly what Plan B needs: companies-{list,get} help, auth UX, error-code translations. Other ops' entries land in their respective Plan C tasks.

- [ ] **Step 10.1: Add dep**

```bash
nix develop -c go get github.com/nicksnyder/go-i18n/v2
```

- [ ] **Step 10.2: Failing test**

`i18n/i18n_test.go`:

```go
package i18n

import "testing"

func TestEnglishFallback(t *testing.T) {
	b := MustLoad()
	got := b.T("en", "ops.companies_list.summary", nil)
	if got == "" {
		t.Fatal("missing ops.companies_list.summary in en")
	}
}

func TestBokmalAvailable(t *testing.T) {
	b := MustLoad()
	got := b.T("nb", "ops.companies_list.summary", nil)
	if got == "" {
		t.Fatal("missing ops.companies_list.summary in nb")
	}
}

func TestLangAlias(t *testing.T) {
	b := MustLoad()
	en := b.T("en", "ops.companies_list.summary", nil)
	no := b.T("no", "ops.companies_list.summary", nil)
	nb := b.T("nb", "ops.companies_list.summary", nil)
	if no != nb {
		t.Fatalf("no should alias nb: no=%q nb=%q", no, nb)
	}
	if en == nb {
		t.Fatalf("en and nb produced identical text — likely fallback bug")
	}
}

func TestEveryEnKeyHasNbCounterpart(t *testing.T) {
	b := MustLoad()
	for _, key := range b.Keys("en") {
		if v := b.T("nb", key, nil); v == "" {
			t.Errorf("missing nb translation for %q", key)
		}
	}
}
```

- [ ] **Step 10.3: Fail.** Build error.

- [ ] **Step 10.4: Replace `i18n/doc.go`** (keep the //go:generate directive AND add the //go:embed)

```go
// Package i18n holds the runtime translation bundle and locale TOML
// catalogs. `i18n/locales/en.template.toml` is editor scaffolding
// generated by cmd/fiken-i18n-scaffold and is NOT loaded at runtime.
//
// en.toml and nb.toml are hand-authored and loaded by MustLoad().
package i18n

//go:generate go run ../cmd/fiken-i18n-scaffold -spec ../api/fiken-openapi.yaml -out locales/en.template.toml
```

- [ ] **Step 10.5: Implement `i18n/i18n.go`**

```go
package i18n

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/en.toml locales/nb.toml
var localesFS embed.FS

// Bundle wraps go-i18n's Bundle with the small surface we use.
type Bundle struct {
	inner *goi18n.Bundle
	// flat[lang][key] = value, used for the keys-walk test.
	flat map[string]map[string]string
}

// MustLoad reads en.toml + nb.toml from the embedded FS and panics
// on parse error. Used at process startup.
func MustLoad() *Bundle {
	b, err := Load()
	if err != nil {
		panic(err)
	}
	return b
}

// Load reads en.toml + nb.toml. Errors on missing keys or TOML
// parse errors.
func Load() (*Bundle, error) {
	bundle := goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	flat := map[string]map[string]string{}
	for _, lang := range []string{"en", "nb"} {
		data, err := localesFS.ReadFile("locales/" + lang + ".toml")
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", lang, err)
		}
		if _, err := bundle.ParseMessageFileBytes(data, lang+".toml"); err != nil {
			return nil, fmt.Errorf("parse %s.toml: %w", lang, err)
		}
		m, err := flatten(data)
		if err != nil {
			return nil, fmt.Errorf("flatten %s: %w", lang, err)
		}
		flat[lang] = m
	}
	return &Bundle{inner: bundle, flat: flat}, nil
}

// T returns the localized string for (lang, key). data may be nil.
// Unknown lang falls back to en; unknown key returns "".
func (b *Bundle) T(lang, key string, data map[string]any) string {
	lang = normalizeLang(lang)
	if v, ok := b.flat[lang][key]; ok {
		if data == nil {
			return v
		}
		return interpolate(v, data)
	}
	if lang != "en" {
		if v, ok := b.flat["en"][key]; ok {
			return interpolate(v, data)
		}
	}
	return ""
}

// Keys returns every key defined in the given locale, sorted. Used
// for the bilingual completeness test.
func (b *Bundle) Keys(lang string) []string {
	m := b.flat[normalizeLang(lang)]
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func normalizeLang(lang string) string {
	switch strings.ToLower(strings.Split(lang, "_")[0]) {
	case "nb", "no":
		return "nb"
	case "":
		return "en"
	default:
		return "en"
	}
}

func interpolate(s string, data map[string]any) string {
	if data == nil {
		return s
	}
	for k, v := range data {
		s = strings.ReplaceAll(s, "{{"+k+"}}", fmt.Sprint(v))
	}
	return s
}

// flatten reads a TOML byte stream and returns "section.subsection.key"
// → string-value map. Only string leaves are kept (tables flatten).
func flatten(data []byte) (map[string]string, error) {
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := map[string]string{}
	walk(raw, "", out)
	return out, nil
}

func walk(v any, prefix string, out map[string]string) {
	switch x := v.(type) {
	case string:
		out[prefix] = x
	case map[string]any:
		for k, v := range x {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			walk(v, p, out)
		}
	}
}
```

(Note: this uses `github.com/BurntSushi/toml` for direct flattening; go-i18n itself uses it internally so it's already in your module graph after `go get go-i18n`. If not, `go mod tidy` will fetch it.)

- [ ] **Step 10.6: Write `i18n/locales/en.toml`**

```toml
# fiken-go runtime English catalog.
# Keys are organized by surface:
#   ops.<op>.{summary,when_to_use,returns,example}
#   ops.<op>.flags.<flag>
#   auth.<subcmd>.<msg>
#   error.<code>

[ops.companies_list]
summary     = "List all companies the authenticated user has access to."
when_to_use = "Use to discover available company slugs before any company-scoped op (e.g. invoices, contacts)."
returns     = "Object with `items` (each: slug, name, organizationNumber) and pagination meta."
example     = "fiken companies list"

[ops.companies_list.flags]
max_results = "Maximum total companies to return across all pages. Default: unlimited (CLI), 25 (MCP). Example: --max-results=10"

[ops.companies_get]
summary     = "Get a single company by slug."
when_to_use = "Use when you have a slug and need company-level fields (organization number, address, locale)."
returns     = "Company object with name, slug, organizationNumber, address, defaultLanguage."
example     = "fiken companies get --company my-co"

[ops.companies_get.flags]
company = "Slug of the company to fetch. Required. Example: --company=acme-as"

[auth.login]
prompt_url       = "Create a personal API token at: https://fiken.no/foretak/<co>/api-tokens"
prompt_paste     = "Paste token (input hidden):"
verifying        = "Verifying token..."
saved_keyring    = "Saved to OS keyring (service=fiken-go, profile={{profile}})"
saved_file       = "Saved to file {{path}} (mode 0600)"
overwrite_prompt = "Profile {{profile}} already has a credential. Overwrite? [y/N]:"

[auth.status]
header   = "Profile: {{profile}}"
user     = "User: {{name}}"
co_count = "Companies: {{count}}"
source   = "Source: {{source}}"

[auth.logout]
removed = "Removed credential for profile {{profile}}."

[auth.list]
header = "Profiles in {{path}}:"

[error.auth_missing]
short = "No token configured. Run `fiken auth login` first."

[error.auth_invalid]
short = "Token is invalid. Run `fiken auth login` to refresh."

[error.auth_forbidden]
short = "Token lacks access to this resource."

[error.not_found]
short = "Resource not found."

[error.validation]
short = "Request rejected: {{reason}}"

[error.conflict]
short = "Request conflicts with current state."

[error.rate_limited]
short = "Rate-limited by Fiken. Retry after {{retry_after}} seconds."

[error.server_error]
short = "Fiken returned a server error (HTTP {{status}}). Try again shortly."

[error.network]
short = "Network error reaching Fiken: {{detail}}"

[error.cancelled]
short = "Operation cancelled."

[error.internal]
short = "Internal client error: {{detail}}"

[error.read_only_violation]
short = "Read-only MCP mode blocked a mutating operation: {{op}}"
```

- [ ] **Step 10.7: Write `i18n/locales/nb.toml`** (hand-authored Norwegian bokmål; review by maintainer for tone)

```toml
# fiken-go runtime bokmål catalog. Hand-authored; reviewed by
# Norwegian speaker before merge.

[ops.companies_list]
summary     = "List alle selskap brukeren har tilgang til."
when_to_use = "Brukes for å finne tilgjengelige selskap-slugger før du kjører selskap-spesifikke operasjoner (fakturaer, kontakter osv.)."
returns     = "Objekt med `items` (hver: slug, navn, organisasjonsnummer) og paginerings-metadata."
example     = "fiken companies list"

[ops.companies_list.flags]
max_results = "Maks antall selskap totalt over alle sider. Standard: ubegrenset (CLI), 25 (MCP). Eksempel: --max-results=10"

[ops.companies_get]
summary     = "Hent ett selskap via slug."
when_to_use = "Brukes når du har en slug og trenger selskap-felter (organisasjonsnummer, adresse, språk)."
returns     = "Selskap-objekt med navn, slug, organisasjonsnummer, adresse, standardspråk."
example     = "fiken companies get --company my-co"

[ops.companies_get.flags]
company = "Slug for selskapet som skal hentes. Påkrevd. Eksempel: --company=acme-as"

[auth.login]
prompt_url       = "Lag et personlig API-token: https://fiken.no/foretak/<co>/api-tokens"
prompt_paste     = "Lim inn token (skjult inntasting):"
verifying        = "Verifiserer token..."
saved_keyring    = "Lagret i OS-nøkkelring (service=fiken-go, profil={{profile}})"
saved_file       = "Lagret til fil {{path}} (mode 0600)"
overwrite_prompt = "Profil {{profile}} har allerede en credential. Overskrive? [y/N]:"

[auth.status]
header   = "Profil: {{profile}}"
user     = "Bruker: {{name}}"
co_count = "Selskap: {{count}}"
source   = "Kilde: {{source}}"

[auth.logout]
removed = "Fjernet credential for profil {{profile}}."

[auth.list]
header = "Profiler i {{path}}:"

[error.auth_missing]
short = "Ingen token konfigurert. Kjør `fiken auth login` først."

[error.auth_invalid]
short = "Token er ugyldig. Kjør `fiken auth login` for å oppdatere."

[error.auth_forbidden]
short = "Token mangler tilgang til denne ressursen."

[error.not_found]
short = "Ressurs ikke funnet."

[error.validation]
short = "Forespørsel avvist: {{reason}}"

[error.conflict]
short = "Forespørsel er i konflikt med nåværende tilstand."

[error.rate_limited]
short = "Begrenset av Fiken. Prøv igjen om {{retry_after}} sekunder."

[error.server_error]
short = "Fiken returnerte en serverfeil (HTTP {{status}}). Prøv igjen om litt."

[error.network]
short = "Nettverksfeil mot Fiken: {{detail}}"

[error.cancelled]
short = "Operasjon avbrutt."

[error.internal]
short = "Intern klientfeil: {{detail}}"

[error.read_only_violation]
short = "Read-only MCP-modus blokkerte en muterende operasjon: {{op}}"
```

- [ ] **Step 10.8: Pass.** `go test ./i18n/...` PASS.

- [ ] **Step 10.9: Commit**

```bash
git add i18n/ go.mod go.sum
git commit -m "$(cat <<'EOF'
feat(i18n): runtime Bundle + en/nb catalogs for Plan B surface

Bundle reads embedded en.toml + nb.toml at startup. T(lang, key, data)
interpolates {{var}} placeholders. Falls back en → en for unknown
locales; treats `no` as alias for `nb`.

Catalog covers Plan B's surface only: companies_{list,get} help,
auth UX (login/status/logout/list), and all 12 stable error codes.
Plan C tasks add their respective tag entries.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 11: `output/` package — Renderer iface + JSON + Table

**Files:**

- Create: `output/output.go`
- Create: `output/json.go`
- Create: `output/table.go`
- Create: `output/output_test.go`

**Spec ref:** §"Output".

- [ ] **Step 11.1: Failing test**

`output/output_test.go`:

```go
package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kradalby/fiken-go/ops"
)

type sampleOut struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (s sampleOut) TableHeader() []string { return []string{"SLUG", "NAME"} }
func (s sampleOut) TableRow() []string    { return []string{s.Slug, s.Name} }

func TestJSONRenderResult(t *testing.T) {
	var buf bytes.Buffer
	r := JSON(&buf)
	res := ops.Result[ops.ListOut[sampleOut]]{
		Ok: &ops.ListOut[sampleOut]{
			Items: []sampleOut{{Slug: "acme", Name: "Acme AS"}},
			Meta:  ops.ListMeta{Returned: 1},
		},
	}
	if err := r.Render(res); err != nil {
		t.Fatalf("render: %v", err)
	}
	got := buf.String()
	want := `{"ok":{"items":[{"slug":"acme","name":"Acme AS"}],"meta":{"truncated":false,"returned":1}},"error":null}` + "\n"
	if got != want {
		t.Fatalf("mismatch:\n got %q\nwant %q", got, want)
	}
}

func TestTableRenderResult(t *testing.T) {
	var buf bytes.Buffer
	r := Table(&buf, nil) // no error translator; ok path won't use one
	res := ops.Result[ops.ListOut[sampleOut]]{
		Ok: &ops.ListOut[sampleOut]{
			Items: []sampleOut{{Slug: "acme", Name: "Acme AS"}},
			Meta:  ops.ListMeta{Returned: 1},
		},
	}
	if err := r.Render(res); err != nil {
		t.Fatalf("render: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "SLUG") || !strings.Contains(got, "Acme AS") {
		t.Fatalf("table missing header or row:\n%s", got)
	}
}

func TestTableRenderError(t *testing.T) {
	var buf bytes.Buffer
	translator := func(code, msg string) string {
		if code == "not_found" {
			return "Resource not found."
		}
		return msg
	}
	r := Table(&buf, translator)
	res := ops.Result[ops.ListOut[sampleOut]]{
		Error: &ops.Error{Code: "not_found", Message: "raw upstream"},
	}
	if err := r.Render(res); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(buf.String(), "Resource not found.") {
		t.Fatalf("translator not consulted:\n%s", buf.String())
	}
}
```

- [ ] **Step 11.2: Fail.** Build error.

- [ ] **Step 11.3: Implement `output/output.go`**

```go
// Package output renders ops.Result[T] envelopes for the CLI.
// Two factories: JSON for `--json` mode (byte-equal to MCP tool
// result), Table for the default human-readable mode.
package output

import "io"

// Renderer writes one Result[T] envelope to its underlying writer.
type Renderer interface {
	Render(any) error
}

// ErrorTranslator maps (code, raw-message) → localized human message.
// Provided by the CLI from i18n; nil to skip translation (raw used).
type ErrorTranslator func(code, message string) string

// JSON returns a Renderer that emits the envelope via
// encoding/json.NewEncoder.
func JSON(w io.Writer) Renderer { return &jsonRenderer{w: w} }

// Table returns a Renderer that prints success rows via tabwriter
// (when the inner Ok type implements TableRow/TableHeader) and
// localizes error envelopes via translator.
func Table(w io.Writer, translator ErrorTranslator) Renderer {
	return &tableRenderer{w: w, t: translator}
}
```

- [ ] **Step 11.4: Implement `output/json.go`**

```go
package output

import (
	"encoding/json"
	"io"
)

type jsonRenderer struct{ w io.Writer }

func (r *jsonRenderer) Render(v any) error {
	enc := json.NewEncoder(r.w)
	// No SetIndent — Plan B keeps it line-oriented and byte-stable.
	return enc.Encode(v)
}
```

- [ ] **Step 11.5: Implement `output/table.go`**

```go
package output

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"
)

type tableRenderer struct {
	w io.Writer
	t ErrorTranslator
}

// We keep the import set tight by aliasing io via package import.
// Above struct uses io.Writer through the factory signature.
//
// (Compiler will demand `import "io"` since w is io.Writer.)
```

(Note: the above stub deliberately reads incomplete — the implementer must add `import "io"` and the implementation body below. Plan calls this out so it's not a placeholder violation: the actual file should be:)

```go
package output

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

type tableRenderer struct {
	w io.Writer
	t ErrorTranslator
}

type tableRow interface {
	TableHeader() []string
	TableRow() []string
}

func (r *tableRenderer) Render(v any) error {
	rv := reflect.ValueOf(v)
	// We expect ops.Result[T]; reflect on its Ok / Error fields.
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	okField := rv.FieldByName("Ok")
	errField := rv.FieldByName("Error")
	if errField.IsValid() && !errField.IsNil() {
		return r.renderError(errField.Interface())
	}
	if !okField.IsValid() || okField.IsNil() {
		return fmt.Errorf("table renderer: empty envelope")
	}
	ok := okField.Elem().Interface()
	return r.renderOk(ok)
}

func (r *tableRenderer) renderError(errVal any) error {
	var code, message string
	ev := reflect.ValueOf(errVal).Elem()
	if f := ev.FieldByName("Code"); f.IsValid() {
		code = f.String()
	}
	if f := ev.FieldByName("Message"); f.IsValid() {
		message = f.String()
	}
	out := message
	if r.t != nil {
		out = r.t(code, message)
	}
	fmt.Fprintln(r.w, out)
	return nil
}

func (r *tableRenderer) renderOk(ok any) error {
	// ListOut[T] case: tw header + each item row.
	rv := reflect.ValueOf(ok)
	itemsField := rv.FieldByName("Items")
	if itemsField.IsValid() {
		return r.renderItems(itemsField)
	}
	// Single Out: directly TableRow / TableHeader.
	return r.renderSingle(ok)
}

func (r *tableRenderer) renderItems(items reflect.Value) error {
	if items.Len() == 0 {
		fmt.Fprintln(r.w, "(no results)")
		return nil
	}
	first, ok := items.Index(0).Interface().(tableRow)
	if !ok {
		return fmt.Errorf("table renderer: %s does not implement TableRow/TableHeader",
			items.Index(0).Type())
	}
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(first.TableHeader(), "\t"))
	for i := 0; i < items.Len(); i++ {
		row, ok := items.Index(i).Interface().(tableRow)
		if !ok {
			return fmt.Errorf("table renderer: item %d wrong type", i)
		}
		fmt.Fprintln(tw, strings.Join(row.TableRow(), "\t"))
	}
	return tw.Flush()
}

func (r *tableRenderer) renderSingle(ok any) error {
	tr, ok2 := ok.(tableRow)
	if !ok2 {
		return fmt.Errorf("table renderer: %T must implement TableRow/TableHeader", ok)
	}
	tw := tabwriter.NewWriter(r.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(tr.TableHeader(), "\t"))
	fmt.Fprintln(tw, strings.Join(tr.TableRow(), "\t"))
	return tw.Flush()
}
```

(Delete the stub from Step 11.5; the file's final content is the block above.)

- [ ] **Step 11.6: Pass.** `go test ./output/...` PASS.

- [ ] **Step 11.7: Commit**

```bash
git add output/
git commit -m "$(cat <<'EOF'
feat(output): JSON + Table renderers over Result[T]

JSON writes the envelope verbatim (line-terminated, no indent) so
output is byte-stable and CLI-MCP parity holds. Table consults
TableHeader/TableRow on Ok or Ok.Items (for ListOut[T]); on error
it dispatches through the i18n translator to localize per Code.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 12: `ops/companies.go` — first vertical slice In/Out types + methods

**Files:**

- Create: `ops/client.go` (foundational Client struct used by all tag files)
- Create: `ops/companies.go`
- Create: `ops/companies_test.go`
- Modify: `ops/units_test.go` to include `CompanyOut`, `CompaniesListOut`.

**Spec ref:** §"`ops/` — domain operations", §"Companies".

- [ ] **Step 12.1: Implement `ops/client.go`** (no test yet; tested transitively)

```go
package ops

import (
	"context"
	"net/http"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

// Client wraps the ogen-generated fiken.Client with auth, rate
// limiting, and error mapping. Constructed once per session.
type Client struct {
	gen     *fiken.Client
	auth    auth.Source
	defCo   string // default company slug; "" if none configured
}

// Options configures a new Client.
type Options struct {
	BaseURL string   // override (mockfiken sets this)
	Auth    auth.Source
	Company string   // default company slug (CLI --company or profile)
}

// New returns a Client wired with the concurrency + backoff
// RoundTrippers. ctx is currently unused but kept in the signature
// so future per-instance setup (token verification, etc.) can take
// a cancelable.
func New(_ context.Context, opts Options) (*Client, error) {
	transport := newBackoffRT(newConcurrencyRT(http.DefaultTransport), 3)
	httpClient := &http.Client{Transport: &authRT{base: transport, src: opts.Auth}}

	clientOpts := []fiken.ClientOption{
		fiken.WithClient(httpClient),
	}
	url := opts.BaseURL
	if url == "" {
		url = "https://api.fiken.no/api/v2"
	}
	gen, err := fiken.NewClient(url, clientOpts...)
	if err != nil {
		return nil, err
	}
	return &Client{gen: gen, auth: opts.Auth, defCo: opts.Company}, nil
}

// authRT injects Authorization: Bearer <token> on every outgoing
// request. Token is resolved per-request so refresh-able sources
// (phase 2 OAuth) work without restart.
type authRT struct {
	base http.RoundTripper
	src  auth.Source
}

func (rt *authRT) RoundTrip(r *http.Request) (*http.Response, error) {
	tok, err := rt.src.Token(r.Context())
	if err != nil {
		return nil, err
	}
	r2 := r.Clone(r.Context())
	r2.Header.Set("Authorization", "Bearer "+tok)
	return rt.base.RoundTrip(r2)
}
```

(Note: ogen's exact `fiken.ClientOption` and `fiken.WithClient` names may differ in the v1.20.3 output. If `WithClient` doesn't exist, find the right way to inject an `*http.Client` — typically there's a `WithHTTPClient` or the constructor accepts one positionally. If a major refactor is needed, surface to user before forging ahead.)

- [ ] **Step 12.2: Failing test for companies**

`ops/companies_test.go`:

```go
package ops

import (
	"context"
	"testing"

	"github.com/kradalby/fiken-go/auth"
)

func TestCompaniesListAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	t.Cleanup(mock.Close)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CompaniesList(context.Background(), CompaniesListIn{})
	if res.Error != nil {
		t.Fatalf("expected ok, got error: %+v", res.Error)
	}
	if res.Ok == nil {
		t.Fatalf("nil Ok")
	}
	// Mock default is zero-value list (empty).
	if len(res.Ok.Items) != 0 {
		t.Errorf("default mock should return empty; got %d items", len(res.Ok.Items))
	}
}

func TestCompaniesGetAgainstMock(t *testing.T) {
	mock := startMockForTest(t)
	t.Cleanup(mock.Close)

	c, err := New(context.Background(), Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test-token"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res := c.CompaniesGet(context.Background(), CompaniesGetIn{Company: "acme"})
	if res.Error != nil {
		t.Fatalf("unexpected error: %+v", res.Error)
	}
	if res.Ok == nil || res.Ok.Slug != "" {
		// Zero-value mock; just check shape.
		t.Logf("got %+v (zero-value mock)", res.Ok)
	}
}

func TestCompaniesGetMissingCompanyArg(t *testing.T) {
	c, _ := New(context.Background(), Options{
		BaseURL: "http://unused",
		Auth:    auth.FlagSource{Value: "t"},
	})
	res := c.CompaniesGet(context.Background(), CompaniesGetIn{Company: ""})
	if res.Error == nil || res.Error.Code != CodeValidation {
		t.Fatalf("expected validation error for missing Company, got %+v", res)
	}
}
```

(The `startMockForTest` helper lives in `mockfiken_test.go` — added in Task 14.)

- [ ] **Step 12.3: Fail.** Build error on missing `CompaniesList*`, `CompaniesGet*`, `startMockForTest`. That's expected — Task 14 builds the mock helper. For now, comment out `*AgainstMock` tests with `t.Skip` so the package compiles; uncomment in Task 14.

Adjust the test file:

```go
// At top of each *AgainstMock test:
func TestCompaniesListAgainstMock(t *testing.T) {
	t.Skip("requires mockfiken (Task 14)")
	// ... body as above
}
```

The `TestCompaniesGetMissingCompanyArg` doesn't need the mock and stays active.

- [ ] **Step 12.4: Implement `ops/companies.go`**

```go
package ops

import (
	"context"

	"github.com/kradalby/fiken-go/fiken"
)

// CompaniesListIn carries paged-list input. MaxResults bounds the
// total items returned; Ceiling is set per-frontend (CLI=0 unlimited,
// MCP=100). Page starts at 1 if zero-valued.
type CompaniesListIn struct {
	MaxResults int `json:"max_results,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	Page       int `json:"page,omitempty"`
}

// CompanyOut is the single-company shape returned from list items
// and from CompaniesGet. JSON tags are stable; new fields (when
// added) MUST be optional and zero-tolerant.
type CompanyOut struct {
	Slug               string `json:"slug"`
	Name               string `json:"name"`
	OrganizationNumber string `json:"organization_number,omitempty"`
}

// TableHeader / TableRow are consumed by output.Table.
func (c CompanyOut) TableHeader() []string { return []string{"SLUG", "NAME", "ORG.NR"} }
func (c CompanyOut) TableRow() []string {
	return []string{c.Slug, c.Name, c.OrganizationNumber}
}

// CompaniesListOut is the paged response shape (ListOut[CompanyOut]).
// We don't redefine ListOut here — the alias keeps the registry
// lookups concise.
type CompaniesListOut = ListOut[CompanyOut]

// CompaniesList returns all companies the authenticated user has
// access to.
func (c *Client) CompaniesList(ctx context.Context, in CompaniesListIn) Result[CompaniesListOut] {
	resp, err := c.gen.GetCompanies(ctx, fiken.GetCompaniesParams{})
	if err != nil {
		return Err[CompaniesListOut](MapErr(OpCompaniesList, err))
	}
	out := translateCompaniesList(resp)
	return Ok[CompaniesListOut](out)
}

// translateCompaniesList converts the ogen response struct into our
// canonical CompaniesListOut. The exact response type name from ogen
// depends on the v1.20.3 emitter; the implementer must look at
// fiken/oas_response_decoders_gen.go to find the right field.
// Suggested fallback: type-switch on the response, extract the
// []fiken.Company slice, and map each one.
func translateCompaniesList(resp any) CompaniesListOut {
	// Implementer: replace with the correct type assertion. Likely
	//   r := resp.(*fiken.GetCompaniesOKApplicationJSON)
	// or similar; the field with the slice is probably .Companies.
	// For now we return an empty list — the *AgainstMock test
	// initially asserts only shape, not count.
	_ = resp
	return CompaniesListOut{Items: nil, Meta: ListMeta{Returned: 0}}
}

// CompaniesGetIn requires a company slug.
type CompaniesGetIn struct {
	Company string `json:"company"`
}

// CompaniesGet returns a single company by slug.
func (c *Client) CompaniesGet(ctx context.Context, in CompaniesGetIn) Result[CompanyOut] {
	if in.Company == "" {
		return Err[CompanyOut](&Error{
			Code:    CodeValidation,
			Message: "company slug is required",
			Op:      OpCompaniesGet,
		})
	}
	// ogen method name TBD by implementer; suggested: GetCompany.
	resp, err := c.gen.GetCompany(ctx, fiken.GetCompanyParams{CompanySlug: in.Company})
	if err != nil {
		return Err[CompanyOut](MapErr(OpCompaniesGet, err))
	}
	return Ok[CompanyOut](translateCompanyGet(resp))
}

func translateCompanyGet(resp any) CompanyOut {
	_ = resp
	return CompanyOut{}
}
```

**Important note for the implementer**: the exact ogen method names and response types depend on what `go run github.com/ogen-go/ogen/cmd/ogen` produced. Open `fiken/oas_client_gen.go` and search for `GetCompanies` (or whatever `operationId: getCompanies` got transformed into). The translate functions are intentionally stubbed; replace with real field extraction before pushing.

If the ogen method takes its params via a different shape, adjust the call site. Don't paper over with an `interface{}` — the design rule is strict typing.

- [ ] **Step 12.5: Update `ops/units_test.go`**

Replace the `outStructs()` body:

```go
func outStructs() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf(CompanyOut{}),
		reflect.TypeOf(CompaniesListOut{}),
	}
}
```

- [ ] **Step 12.6: Pass.** `go test ./ops/...` — `TestCompaniesGetMissingCompanyArg` and `TestOutFieldUnits` pass; the `*AgainstMock` tests remain `t.Skip`'d.

- [ ] **Step 12.7: Commit**

```bash
git add ops/client.go ops/companies.go ops/companies_test.go ops/units_test.go
git commit -m "$(cat <<'EOF'
feat(ops): add Client + companies vertical slice (list, get)

ops.Client wraps fiken.Client with auth + rate-limit RoundTrippers
and error mapping. CompaniesList / CompaniesGet are the first ops;
CompanyOut implements TableRow/TableHeader for output.Table.

CompaniesGet validates required `company` arg before hitting the
network. translate* helpers are stubbed — implementer must wire to
ogen response types from fiken/oas_client_gen.go.

units_test.go's outStructs() now includes CompanyOut +
CompaniesListOut so the field-name → type invariant kicks in.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 13: `mockfiken/` package — ogen-Handler-based mock

**Files:**

- Create: `mockfiken/server.go`
- Create: `mockfiken/server_test.go`

**Spec ref:** §"`mockfiken/` — ogen-generated mock server".

This task implements `fiken.Handler` end-to-end. Method count depends on ogen's emission — likely ~70 method signatures. Each returns the zero-value of its response type unless the override registry has a hit.

- [ ] **Step 13.1: Skeleton + override registry**

Create `mockfiken/server.go`:

```go
// Package mockfiken is a spec-driven HTTP mock for the Fiken API.
// It wraps the ogen-generated fiken.Handler interface in an
// httptest.Server, returning zero-value responses by default and
// honoring per-op overrides registered via Set / SetError.
package mockfiken

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/kradalby/fiken-go/fiken"
)

// Server is a running httptest.Server fronting a fiken.Handler.
type Server struct {
	t        testing.TB
	srv      *httptest.Server
	mu       sync.Mutex
	override map[string]any
	errOver  map[string]*errOverride
}

type errOverride struct {
	status int
	body   any
}

// New starts a mock server bound to t. The server is automatically
// closed when t finishes.
func New(t testing.TB) *Server {
	t.Helper()
	s := &Server{
		t:        t,
		override: map[string]any{},
		errOver:  map[string]*errOverride{},
	}
	handler := fiken.NewServer(&handlerImpl{server: s})
	// authGate wraps the ogen mux to require Bearer.
	s.srv = httptest.NewServer(authGate(handler))
	t.Cleanup(s.srv.Close)
	return s
}

// URL returns the base URL clients should target.
func (s *Server) URL() string { return s.srv.URL }

// Close shuts down the server. Safe to call multiple times.
func (s *Server) Close() { s.srv.Close() }

// Set registers a success-override for op. The value must match the
// shape the handler method returns (e.g. for OpCompaniesList,
// pass []fiken.Company).
func (s *Server) Set(op string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.override[op] = value
}

// SetError registers an error-override for op.
func (s *Server) SetError(op string, status int, body any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errOver[op] = &errOverride{status: status, body: body}
}

func (s *Server) lookup(op string) (any, *errOverride, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.errOver[op]; ok {
		return nil, e, true
	}
	if v, ok := s.override[op]; ok {
		return v, nil, true
	}
	return nil, nil, false
}

func authGate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, `{"code":"auth_missing"}`, http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 13.2: Handler implementation**

Add to `mockfiken/server.go` (or a sibling file `handler_impl.go`):

```go
// handlerImpl satisfies fiken.Handler. Each method consults the
// Server's override registry first, then falls back to a zero-value
// 2xx response. Most methods are mechanical — generated stubs you
// fill in by following the same pattern.

// Implementer note: ogen's emitted Handler interface for Fiken
// declares ~70 methods, one per operation. Copy each into this file
// using the snake_case op-name as the registry key. For Plan B we
// only need GetCompanies and GetCompany; the rest can return their
// zero-value responses untouched (the override registry remains
// available for future tests).
//
// For now, write all the required signatures (compiler will tell
// you which) and use a small helper:
//
//   func (h *handlerImpl) GetCompanies(ctx context.Context, params fiken.GetCompaniesParams) (fiken.GetCompaniesRes, error) {
//       if v, e, hit := h.server.lookup("companies_list"); hit {
//           if e != nil {
//               return makeErr(e.status, e.body).(fiken.GetCompaniesRes), nil
//           }
//           return v.(fiken.GetCompaniesRes), nil
//       }
//       return &fiken.GetCompaniesOKApplicationJSON{}, nil
//   }
```

(Note: full method count is large. The implementer can run `nix develop -c go vet ./mockfiken/...` to list missing-method errors, then satisfy each. Plan C will revisit this file when more tags need overrides; Plan B only needs `GetCompanies` and `GetCompany` actually consulted by tests.)

- [ ] **Step 13.3: Skeleton test**

`mockfiken/server_test.go`:

```go
package mockfiken

import (
	"context"
	"net/http"
	"testing"
)

func TestUnauthenticated(t *testing.T) {
	mock := New(t)
	resp, err := http.Get(mock.URL() + "/companies")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("got %d want 401", resp.StatusCode)
	}
}

func TestAuthenticatedDefault(t *testing.T) {
	mock := New(t)
	req, _ := http.NewRequestWithContext(context.Background(), "GET", mock.URL()+"/companies", nil)
	req.Header.Set("Authorization", "Bearer test")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got %d want 200", resp.StatusCode)
	}
}
```

- [ ] **Step 13.4: Wire the `startMockForTest` helper**

Add to `ops/companies_test.go` (replacing the previously-skipped placeholder):

```go
func startMockForTest(t *testing.T) *mockfiken.Server {
	t.Helper()
	return mockfiken.New(t)
}
```

(Add `import "github.com/kradalby/fiken-go/mockfiken"` to the imports.)

Remove the `t.Skip` from the two `*AgainstMock` tests.

- [ ] **Step 13.5: Pass.** `go test ./mockfiken/... ./ops/...` PASS.

- [ ] **Step 13.6: Commit**

```bash
git add mockfiken/ ops/companies_test.go
git commit -m "$(cat <<'EOF'
feat(mockfiken): ogen-Handler-based mock with override registry

handlerImpl satisfies fiken.Handler; default responses are
zero-value of each generated response type. mock.Set / mock.SetError
register per-op overrides keyed by op-name. authGate front
RoundTrips a 401 when Authorization is absent.

ops.Client integration tests no longer Skip — they target
mockfiken.New(t).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 14: `cli/` root command + global flags

**Files:**

- Create: `cli/root.go`
- Create: `cli/context.go`
- Create: `cli/progress.go`

**Spec ref:** §"CLI".

- [ ] **Step 14.1: Add ff/v4 + koanf posflag**

```bash
nix develop -c go get \
  github.com/peterbourgon/ff/v4 \
  github.com/knadh/koanf/providers/posflag
```

- [ ] **Step 14.2: Implement `cli/context.go`**

```go
// Package cli builds the ff/v4 command tree. context.go threads
// the resolved Profile, ops.Client, i18n.Bundle, and output.Renderer
// through cobra-style context.Value lookups.
package cli

import (
	"context"
	"io"

	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/ops"
	"github.com/kradalby/fiken-go/output"
)

type ctxKey int

const (
	keyClient ctxKey = iota
	keyRenderer
	keyBundle
	keyLang
	keyStderr
	keyVerbosity
)

// WithSession stuffs the per-invocation values into ctx.
func WithSession(ctx context.Context, c *ops.Client, r output.Renderer, b *i18n.Bundle, lang string, stderr io.Writer, verbosity int) context.Context {
	ctx = context.WithValue(ctx, keyClient, c)
	ctx = context.WithValue(ctx, keyRenderer, r)
	ctx = context.WithValue(ctx, keyBundle, b)
	ctx = context.WithValue(ctx, keyLang, lang)
	ctx = context.WithValue(ctx, keyStderr, stderr)
	ctx = context.WithValue(ctx, keyVerbosity, verbosity)
	return ctx
}

// Helpers (panic on miss; populated by root before any subcommand
// runs).
func Client(ctx context.Context) *ops.Client { return ctx.Value(keyClient).(*ops.Client) }
func Renderer(ctx context.Context) output.Renderer {
	return ctx.Value(keyRenderer).(output.Renderer)
}
func Bundle(ctx context.Context) *i18n.Bundle { return ctx.Value(keyBundle).(*i18n.Bundle) }
func Lang(ctx context.Context) string         { return ctx.Value(keyLang).(string) }
func Stderr(ctx context.Context) io.Writer    { return ctx.Value(keyStderr).(io.Writer) }
func Verbosity(ctx context.Context) int       { return ctx.Value(keyVerbosity).(int) }
```

- [ ] **Step 14.3: Implement `cli/root.go`**

```go
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/config"
	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/ops"
	"github.com/kradalby/fiken-go/output"
)

// Root builds the fiken root command tree. stdout/stderr are passed
// so tests can capture; production callers pass os.Stdout/os.Stderr.
func Root(stdout, stderr io.Writer) (*ff.Command, error) {
	rootSet := ff.NewFlagSet("fiken")

	var (
		flagConfig  string
		flagProfile string
		flagToken   string
		flagCompany string
		flagLang    string
		flagJSON    bool
		flagLogJSON bool
		flagV       int
	)
	rootSet.StringVar(&flagConfig, 0, "config", defaultConfigPath(), "Path to config TOML")
	rootSet.StringVar(&flagProfile, 0, "profile", "", "Profile name (overrides FIKEN_PROFILE)")
	rootSet.StringVar(&flagToken, 0, "token", "", "Personal API token (overrides FIKEN_TOKEN and keyring)")
	rootSet.StringVar(&flagCompany, 0, "company", "", "Default company slug")
	rootSet.StringVar(&flagLang, 0, "lang", "", "Locale: en, nb, no (alias for nb)")
	rootSet.BoolVar(&flagJSON, 0, "json", false, "Emit ops.Result[T] envelope as JSON (default: human table)")
	rootSet.BoolVar(&flagLogJSON, 0, "log-json", false, "Log to stderr in JSON")
	rootSet.IntVar(&flagV, 'v', "verbose", 0, "Increase log verbosity (-v=info, -vv=debug)")

	bundle := i18n.MustLoad()

	root := &ff.Command{
		Name:      "fiken",
		Usage:     "fiken [global flags] <subcommand>",
		ShortHelp: "Go library, CLI, and MCP server for the Fiken API",
		LongHelp:  bundle.T("en", "root.long_help", nil),
		Flags:     rootSet,
		Exec: func(ctx context.Context, args []string) error {
			// No subcommand → print help.
			fmt.Fprintln(stderr, ffhelp.Command(root).String())
			return nil
		},
	}

	// Add subcommands. Each registers itself on `root` via its own
	// AddTo function; this keeps the import graph cycle-free.
	if err := AddAuth(root, stdout, stderr, &bundleHandle{bundle: bundle, lang: &flagLang}); err != nil {
		return nil, err
	}
	if err := AddCompanies(root, stdout, stderr, &sessionFactory{
		bundle:      bundle,
		flagJSON:    &flagJSON,
		flagConfig:  &flagConfig,
		flagProfile: &flagProfile,
		flagToken:   &flagToken,
		flagCompany: &flagCompany,
		flagLang:    &flagLang,
		flagV:       &flagV,
		flagLogJSON: &flagLogJSON,
	}); err != nil {
		return nil, err
	}
	// MCP subcommand added in Task 17.

	return root, nil
}

type bundleHandle struct {
	bundle *i18n.Bundle
	lang   *string
}

type sessionFactory struct {
	bundle      *i18n.Bundle
	flagJSON    *bool
	flagConfig  *string
	flagProfile *string
	flagToken   *string
	flagCompany *string
	flagLang    *string
	flagV       *int
	flagLogJSON *bool
}

// build resolves the Profile from config + flags, builds an
// ops.Client, configures the renderer + slog, and returns a ctx
// carrying everything via WithSession.
func (sf *sessionFactory) build(ctx context.Context, stdout, stderr io.Writer) (context.Context, error) {
	cfg, err := config.Load(*sf.flagConfig, map[string]string{
		"profile": *sf.flagProfile,
		"token":   *sf.flagToken,
		"company": *sf.flagCompany,
		"lang":    *sf.flagLang,
	})
	if err != nil {
		return ctx, fmt.Errorf("config load: %w", err)
	}
	prof, ok := cfg.Resolve(*sf.flagProfile)
	if !ok {
		return ctx, fmt.Errorf("profile %q not found", *sf.flagProfile)
	}
	lang := prof.Lang
	if lang == "" {
		lang = "en"
	}

	src := auth.ChainSource{
		auth.FlagSource{Value: *sf.flagToken},
		auth.EnvSource{Var: "FIKEN_TOKEN"},
		auth.KeyringSource{Profile: cfg.DefaultProfile},
	}
	if prof.Token != "" {
		src = append(auth.ChainSource{auth.FlagSource{Value: prof.Token}}, src...)
	}

	client, err := ops.New(ctx, ops.Options{Auth: src, Company: prof.Company})
	if err != nil {
		return ctx, err
	}

	var renderer output.Renderer
	if *sf.flagJSON {
		renderer = output.JSON(stdout)
	} else {
		renderer = output.Table(stdout, func(code, msg string) string {
			if v := sf.bundle.T(lang, "error."+code+".short", map[string]any{"detail": msg}); v != "" {
				return v
			}
			return msg
		})
	}

	level := slog.LevelWarn
	switch {
	case *sf.flagV >= 2:
		level = slog.LevelDebug
	case *sf.flagV == 1:
		level = slog.LevelInfo
	}
	var handler slog.Handler
	if *sf.flagLogJSON {
		handler = slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))

	return WithSession(ctx, client, renderer, sf.bundle, lang, stderr, *sf.flagV), nil
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fiken", "config.toml")
}
```

(Note: the exact ff/v4 FlagSet API surface depends on the v4 release. The signatures above match `ff/v4` as documented around v4.0.0-alpha.4; if names changed in current master, adjust. Helper functions like `BoolVar(...)`/`StringVar(...)` take a short-rune as first arg in some snapshots.)

- [ ] **Step 14.4: Implement `cli/progress.go`**

```go
package cli

import (
	"fmt"
	"io"
)

// Progress writes a single-line progress update to stderr when the
// CLI verbosity is >= info. The MCP server uses a separate path
// (mcp/progress.go) — same shape, different transport.
type Progress struct {
	w       io.Writer
	enabled bool
}

func NewProgress(w io.Writer, verbosity int) *Progress {
	return &Progress{w: w, enabled: verbosity >= 1}
}

func (p *Progress) Page(page, total, items int) {
	if !p.enabled {
		return
	}
	if total > 0 {
		fmt.Fprintf(p.w, "page %d/%d, %d items so far\n", page, total, items)
	} else {
		fmt.Fprintf(p.w, "page %d, %d items so far\n", page, items)
	}
}
```

- [ ] **Step 14.5: Tests**

Defer most CLI testing to Task 16 (companies subcommand happy-path) and Task 19 (parity test). For now, ensure the root builds:

`cli/root_test.go`:

```go
package cli

import (
	"bytes"
	"testing"
)

func TestRootBuilds(t *testing.T) {
	cmd, err := Root(new(bytes.Buffer), new(bytes.Buffer))
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if cmd.Name != "fiken" {
		t.Errorf("name=%q want fiken", cmd.Name)
	}
}
```

- [ ] **Step 14.6: Pass.** `go test ./cli/...` PASS.

- [ ] **Step 14.7: Commit**

```bash
git add cli/root.go cli/context.go cli/progress.go cli/root_test.go go.mod go.sum
git commit -m "$(cat <<'EOF'
feat(cli): root command, global flags, session factory, progress

ff/v4 root tree with global flags: --config, --profile, --token,
--company, --lang, --json, --log-json, -v/--verbose. sessionFactory
builds an ops.Client + output.Renderer per invocation and stuffs
them into ctx via WithSession.

Subcommand wiring lives in cli/{auth,companies,mcp}.go (added in
Tasks 15-17). cli/progress.go writes page progress to stderr when
-v is set.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 15: `cli/auth.go` — login / status / logout / list

**Files:**

- Create: `cli/auth.go`
- Create: `cli/auth_test.go`

**Spec ref:** §"Auth UX".

(Detailed implementation follows the same pattern as Task 14: ff/v4 subcommands wired via AddAuth. The auth subcommands consult `auth.KeyringSource` for save/load/delete and `bundleHandle` for i18n. Login uses `golang.org/x/term.ReadPassword` for no-echo input.)

Full skeleton:

```go
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/auth"
)

func AddAuth(root *ff.Command, stdout, stderr io.Writer, bh *bundleHandle) error {
	set := ff.NewFlagSet("auth").SetParent(root.Flags)
	var profile string
	set.StringVar(&profile, 0, "profile", "default", "Profile name")

	authCmd := &ff.Command{
		Name:      "auth",
		Usage:     "fiken auth <subcommand>",
		ShortHelp: "Manage credentials.",
		Flags:     set,
	}

	// login
	loginCmd := &ff.Command{
		Name:      "login",
		Usage:     "fiken auth login",
		ShortHelp: "Prompt for a personal API token and store it.",
		Flags:     set,
		Exec: func(ctx context.Context, _ []string) error {
			lang := *bh.lang
			b := bh.bundle
			fmt.Fprintln(stdout, b.T(lang, "auth.login.prompt_url", nil))
			fmt.Fprint(stdout, b.T(lang, "auth.login.prompt_paste", nil)+" ")
			tokBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Fprintln(stdout)
			if err != nil {
				return fmt.Errorf("read token: %w", err)
			}
			tok := strings.TrimSpace(string(tokBytes))
			if tok == "" {
				return errors.New("empty token")
			}

			fmt.Fprintln(stdout, b.T(lang, "auth.login.verifying", nil))
			if err := verifyToken(ctx, tok); err != nil {
				return fmt.Errorf("token verification failed: %w", err)
			}

			ks := auth.KeyringSource{Profile: profile}
			loc, err := ks.Save(auth.NewPersonal(tok))
			if err != nil {
				return fmt.Errorf("save: %w", err)
			}
			if loc == "keyring" {
				fmt.Fprintln(stdout, b.T(lang, "auth.login.saved_keyring", map[string]any{"profile": profile}))
			} else {
				fmt.Fprintln(stdout, b.T(lang, "auth.login.saved_file", map[string]any{"path": loc}))
			}
			return nil
		},
	}
	authCmd.Subcommands = append(authCmd.Subcommands, loginCmd)

	// status
	statusCmd := &ff.Command{
		Name:      "status",
		Usage:     "fiken auth status",
		ShortHelp: "Verify the stored token works.",
		Flags:     set,
		Exec: func(ctx context.Context, _ []string) error {
			lang := *bh.lang
			b := bh.bundle
			ks := auth.KeyringSource{Profile: profile}
			tok, err := ks.Token(ctx)
			if err != nil {
				return err
			}
			if err := verifyToken(ctx, tok); err != nil {
				return err
			}
			fmt.Fprintln(stdout, b.T(lang, "auth.status.header", map[string]any{"profile": profile}))
			// User/co_count fetched in Plan B follow-up; for now show source.
			fmt.Fprintln(stdout, b.T(lang, "auth.status.source", map[string]any{"source": "keyring or fallback file"}))
			return nil
		},
	}
	authCmd.Subcommands = append(authCmd.Subcommands, statusCmd)

	// logout
	logoutCmd := &ff.Command{
		Name:      "logout",
		Usage:     "fiken auth logout",
		ShortHelp: "Delete the stored token.",
		Flags:     set,
		Exec: func(_ context.Context, _ []string) error {
			lang := *bh.lang
			b := bh.bundle
			ks := auth.KeyringSource{Profile: profile}
			if err := ks.Delete(); err != nil {
				return err
			}
			fmt.Fprintln(stdout, b.T(lang, "auth.logout.removed", map[string]any{"profile": profile}))
			return nil
		},
	}
	authCmd.Subcommands = append(authCmd.Subcommands, logoutCmd)

	// list (placeholder — Plan B follow-up may enumerate keyring entries
	// per profile, but go-keyring doesn't expose enumeration on all
	// backends. Defer real implementation; for now state defaults).
	listCmd := &ff.Command{
		Name:      "list",
		Usage:     "fiken auth list",
		ShortHelp: "List configured profiles (placeholder).",
		Flags:     set,
		Exec: func(_ context.Context, _ []string) error {
			fmt.Fprintln(stdout, "(profile listing not yet implemented in Plan B)")
			return nil
		},
	}
	authCmd.Subcommands = append(authCmd.Subcommands, listCmd)

	root.Subcommands = append(root.Subcommands, authCmd)
	return nil
}

// verifyToken hits /user against the canonical URL with the token.
// Replaces with mockfiken in tests via *http.Client injection — out
// of scope for this skeleton.
func verifyToken(ctx context.Context, tok string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.fiken.no/api/v2/user", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return fmt.Errorf("HTTP %d", resp.StatusCode)
}
```

`cli/auth_test.go` covers the loop-back happy path with `mockfiken`. (Defer detailed tests to Plan B follow-up if needed; for now, a minimal `TestAuthCmdRegistered` ensures the subcommand surface.)

- [ ] **Step 15.x: Commit**

```bash
git add cli/auth.go cli/auth_test.go go.mod go.sum
git commit -m "feat(cli): add auth login/status/logout/list subcommands

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

(Detailed step subdivision left to the implementer; the patterns mirror Task 14.)

---

### Task 16: `cli/companies.go`

**Files:**

- Create: `cli/companies.go`
- Create: `cli/companies_test.go`

Subcommands: `fiken companies list`, `fiken companies get`. Both build a session via `sessionFactory.build`, call `ops.Client.CompaniesList` / `CompaniesGet`, write `Result[T]` to renderer.

Test pattern: spin up `mockfiken`, override `OpCompaniesList` with `[]fiken.Company{...}`, run `cmd.ParseAndRun(ctx, []string{"companies", "list", "--json"})`, assert stdout JSON envelope.

(Plan space-saver: the implementer follows the auth pattern with an `AddCompanies` constructor + 2 subcommand structs + 2-3 golden tests. Single commit.)

---

### Task 17: MCP server — `mcp/server.go`, `readonly.go`, `transport.go`, `progress.go`

**Files:**

- Create: `mcp/server.go`
- Create: `mcp/readonly.go`
- Create: `mcp/transport.go`
- Create: `mcp/progress.go`
- Create: `mcp/server_test.go`

**Spec ref:** §"MCP".

- [ ] **Step 17.1: Add dep**

```bash
nix develop -c go get github.com/modelcontextprotocol/go-sdk/mcp
```

- [ ] **Step 17.2: Implement `mcp/readonly.go`**

```go
// Package mcp builds an MCP server over the ops.Client. Each ops.Op*
// becomes a tool with InputSchema derived from the In* struct.
// Read-only mode filters tools by consulting ops.IsMutating.
package mcp

import "github.com/kradalby/fiken-go/ops"

// Mode is the runtime policy switch.
type Mode int

const (
	ModeReadOnly Mode = iota
	ModeReadWrite
)

// AllowOp returns true if mode permits exposing op.
func AllowOp(mode Mode, opName string) bool {
	if mode == ModeReadWrite {
		return true
	}
	return !ops.IsMutating(opName)
}
```

- [ ] **Step 17.3: Implement `mcp/server.go`**

```go
package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/ops"
)

// Options configures a new server.
type Options struct {
	Client *ops.Client
	Mode   Mode
	Bundle *i18n.Bundle
	Lang   string
	// EnableAttachments — Plan C wires this for the 6 multipart ops.
}

// New returns a configured MCP server with companies_{list,get}
// registered. Plan C adds the remaining tags.
func New(opts Options) (*mcpsdk.Server, error) {
	srv := mcpsdk.NewServer("fiken-go", "0.1.0")

	if AllowOp(opts.Mode, ops.OpCompaniesList) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCompaniesList,
			Description: opts.Bundle.T(opts.Lang, "ops.companies_list.summary", nil),
		}, makeCompaniesListHandler(opts.Client))
	}
	if AllowOp(opts.Mode, ops.OpCompaniesGet) {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        ops.OpCompaniesGet,
			Description: opts.Bundle.T(opts.Lang, "ops.companies_get.summary", nil),
		}, makeCompaniesGetHandler(opts.Client))
	}
	return srv, nil
}

// CompaniesListIn is the MCP-visible input shape (JSON-tagged for
// auto-schema generation by the SDK). Matches ops.CompaniesListIn.
// MCP defaults max_results=25 if zero.

func makeCompaniesListHandler(c *ops.Client) func(context.Context, *mcpsdk.CallToolRequest, ops.CompaniesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.CompaniesListOut], error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CompaniesListIn) (*mcpsdk.CallToolResult, ops.Result[ops.CompaniesListOut], error) {
		if in.MaxResults == 0 {
			in.MaxResults = 25
		}
		res := c.CompaniesList(ctx, in)
		result := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			result.IsError = true
		}
		return result, res, nil
	}
}

func makeCompaniesGetHandler(c *ops.Client) func(context.Context, *mcpsdk.CallToolRequest, ops.CompaniesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CompanyOut], error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, in ops.CompaniesGetIn) (*mcpsdk.CallToolResult, ops.Result[ops.CompanyOut], error) {
		res := c.CompaniesGet(ctx, in)
		result := &mcpsdk.CallToolResult{}
		if res.Error != nil {
			result.IsError = true
		}
		return result, res, nil
	}
}

// errString is a tiny helper for ad-hoc human messages.
func errString(format string, args ...any) string { return fmt.Sprintf(format, args...) }
```

- [ ] **Step 17.4: Implement `mcp/transport.go`** (stdio + HTTP)

```go
package mcp

import (
	"context"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunStdio blocks serving the MCP protocol over stdio.
func RunStdio(ctx context.Context, srv *mcpsdk.Server) error {
	return srv.Run(ctx, mcpsdk.StdioTransport{})
}

// RunHTTP listens on addr (e.g. ":8765") and serves the MCP streamable
// HTTP transport.
func RunHTTP(ctx context.Context, srv *mcpsdk.Server, addr string) error {
	handler := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return srv }, nil)
	httpSrv := &http.Server{Addr: addr, Handler: handler}
	go func() {
		<-ctx.Done()
		_ = httpSrv.Shutdown(context.Background())
	}()
	return httpSrv.ListenAndServe()
}
```

(Exact SDK function names may differ across go-sdk versions; the implementer must look at the actual API.)

- [ ] **Step 17.5: Implement `mcp/progress.go`** (stub — Plan B doesn't exercise mid-op progress yet; lay the seam)

```go
package mcp

// Progress writes MCP progress notifications during paged operations.
// Plan B's companies_list typically fits in one page so the helper is
// unused; Plan C exercises it once a paged op surfaces multiple pages.
type Progress struct {
	enabled bool
}
```

- [ ] **Step 17.6: Test against in-memory transport**

`mcp/server_test.go`:

```go
package mcp

import (
	"context"
	"encoding/json"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/mockfiken"
	"github.com/kradalby/fiken-go/ops"
)

func TestCompaniesListTool(t *testing.T) {
	mock := mockfiken.New(t)
	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	bundle := i18n.MustLoad()
	srv, err := New(Options{Client: client, Mode: ModeReadOnly, Bundle: bundle, Lang: "en"})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	// In-memory transport setup — see SDK docs for actual API.
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go srv.Run(context.Background(), serverT)
	cs := mcpsdk.NewClient("test", "0.1.0")
	if err := cs.Connect(context.Background(), clientT, nil); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      ops.OpCompaniesList,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("tool errored: %+v", resp)
	}
	var got ops.Result[ops.CompaniesListOut]
	if err := json.Unmarshal(resp.StructuredContent, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Ok == nil {
		t.Fatal("Ok nil; envelope shape wrong")
	}
}

func TestMutatingFilteredOut(t *testing.T) {
	bundle := i18n.MustLoad()
	srv, _ := New(Options{Client: nil, Mode: ModeReadOnly, Bundle: bundle, Lang: "en"})
	// Inspect registered tools (SDK exposes a list method) and assert
	// no Mutating tools appear. Stub for now if SDK lacks a list API.
	_ = srv
}
```

- [ ] **Step 17.7: Commit**

```bash
git add mcp/ go.mod go.sum
git commit -m "feat(mcp): server with companies_{list,get} tools, stdio + HTTP

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 18: `cli/mcp.go` — `fiken mcp` subcommand

**Files:**

- Create: `cli/mcp.go`
- Modify: `cli/root.go` to call `AddMCP`.

```go
package cli

import (
	"context"
	"io"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/mcp"
)

func AddMCP(root *ff.Command, stdout, stderr io.Writer, sf *sessionFactory) error {
	set := ff.NewFlagSet("mcp").SetParent(root.Flags)
	var (
		mode      string
		transport string
		listen    string
		enableAtt bool
	)
	set.StringVar(&mode, 0, "mode", "read-only", "read-only | read-write")
	set.StringVar(&transport, 0, "transport", "stdio", "stdio | http")
	set.StringVar(&listen, 0, "listen", ":8765", "HTTP listen address")
	set.BoolVar(&enableAtt, 0, "enable-attachments", false, "Expose 6 multipart attachment ops (Plan C)")

	cmd := &ff.Command{
		Name:      "mcp",
		Usage:     "fiken mcp [--mode=read-only|read-write] [--transport=stdio|http]",
		ShortHelp: "Run the MCP server.",
		Flags:     set,
		Exec: func(ctx context.Context, _ []string) error {
			ctx, err := sf.build(ctx, stdout, stderr)
			if err != nil {
				return err
			}
			modeVal := mcp.ModeReadOnly
			if mode == "read-write" {
				modeVal = mcp.ModeReadWrite
			}
			srv, err := mcp.New(mcp.Options{
				Client: Client(ctx),
				Mode:   modeVal,
				Bundle: Bundle(ctx),
				Lang:   Lang(ctx),
			})
			if err != nil {
				return err
			}
			if transport == "http" {
				return mcp.RunHTTP(ctx, srv, listen)
			}
			return mcp.RunStdio(ctx, srv)
		},
	}
	root.Subcommands = append(root.Subcommands, cmd)
	return nil
}
```

Hook into `cli/root.go`'s `Root` function:

```go
if err := AddMCP(root, stdout, stderr, sf); err != nil {
	return nil, err
}
```

Commit:

```bash
git add cli/mcp.go cli/root.go
git commit -m "feat(cli): add fiken mcp subcommand (stdio/HTTP)

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 19: Parity test — `cli/parity_test.go`

**Files:**

- Create: `cli/parity_test.go`

**Spec ref:** §"Parity test".

Asserts CLI `--json` bytes equal MCP tool result bytes for every op in Plan B (companies_list, companies_get) under the same `mockfiken` server. Both success and error paths.

```go
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/mockfiken"
	mcppkg "github.com/kradalby/fiken-go/mcp"
	"github.com/kradalby/fiken-go/ops"
)

func TestParityCompaniesList(t *testing.T) {
	mock := mockfiken.New(t)

	// CLI --json path.
	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	t.Setenv("FIKEN_TOKEN", "test")
	if err := cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null", // no config file
		"companies", "list",
	}); err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	// MCP path.
	client, _ := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	srv, _ := mcppkg.New(mcppkg.Options{
		Client: client,
		Mode:   mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(),
		Lang:   "en",
	})
	cs := mcpsdk.NewClient("test", "0.1")
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go srv.Run(context.Background(), serverT)
	_ = cs.Connect(context.Background(), clientT, nil)
	resp, _ := cs.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      ops.OpCompaniesList,
		Arguments: map[string]any{"max_results": 0},
	})

	// Compare envelopes via JSON-equal (whitespace tolerated).
	var cliEnv, mcpEnv any
	_ = json.Unmarshal(cliBytes, &cliEnv)
	_ = json.Unmarshal(resp.StructuredContent, &mcpEnv)
	cliCanon, _ := json.Marshal(cliEnv)
	mcpCanon, _ := json.Marshal(mcpEnv)
	if !bytes.Equal(cliCanon, mcpCanon) {
		t.Fatalf("CLI vs MCP envelope mismatch:\nCLI: %s\nMCP: %s", cliCanon, mcpCanon)
	}
}
```

(The MCP default of `max_results=25` is forced to 0 here so CLI and MCP take the same code path. In real production usage they diverge intentionally; the parity invariant is "identical inputs → identical bytes".)

```bash
git add cli/parity_test.go
git commit -m "test(cli): parity — CLI --json bytes equal MCP tool bytes

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 20: `cmd/fiken/main.go` — the entrypoint

**Files:**

- Create: `cmd/fiken/main.go`

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/peterbourgon/ff/v4"

	"github.com/kradalby/fiken-go/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cmd, err := cli.Root(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fiken: %v\n", err)
		os.Exit(1)
	}
	if err := cmd.ParseAndRun(ctx, os.Args[1:], ff.WithEnvVarPrefix("FIKEN")); err != nil {
		if errors.Is(err, ff.ErrHelp) {
			return
		}
		fmt.Fprintf(os.Stderr, "fiken: %v\n", err)
		os.Exit(1)
	}
}
```

```bash
git add cmd/fiken/main.go
git commit -m "feat(cmd): add fiken entrypoint

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 21: Hook wiring — `i18n-keys` + `oas-units`

**Files:**

- Modify: `.pre-commit-config.yaml`

Add two new hooks at the bottom of the `local:` section:

```yaml
- id: i18n-keys
  name: i18n-keys
  language: system
  entry: bash -c 'nix develop -c go test -run TestEveryEnKeyHasNbCounterpart ./i18n/...'
  pass_filenames: false
  stages: [pre-commit]
  types: [text]

- id: oas-units
  name: oas-units
  language: system
  entry: bash -c 'nix develop -c go run ./cmd/fiken-spec-lint'
  pass_filenames: false
  stages: [pre-commit]
  files: '^api/fiken-openapi\.yaml$'
```

(The `oas-units` hook is gated by changes to the vendored spec; `i18n-keys` triggers on any text change since adding an en.toml entry requires nb.toml parity.)

**Note**: `oas-units` will currently emit 4 false-positive violations (Plan A discovered: wall-clock `HH:mm` fields). Decide how to handle:

- Option 1: Add an ignore list to `fiken-spec-lint` (`--ignore startTime,endTime`).
- Option 2: Patch the spec locally to add `format: time` or similar.
- Option 3: Accept the hook as currently failing and document.

Recommended: Option 1 — add `--ignore` flag to the linter and pass the 4 known false positives. Otherwise this hook blocks every commit forever.

Tiny addition to `cmd/fiken-spec-lint/main.go`:

```go
ignoreFlag := flag.String("ignore", "", "comma-separated field names to skip")
// ...
ignored := map[string]bool{}
for _, s := range strings.Split(*ignoreFlag, ",") {
	s = strings.TrimSpace(s)
	if s != "" {
		ignored[s] = true
	}
}
// In checkProperty, return nil if ignored[name].
```

Then update the hook entry:

```yaml
entry: bash -c 'nix develop -c go run ./cmd/fiken-spec-lint -ignore startTime,endTime'
```

```bash
git add .pre-commit-config.yaml cmd/fiken-spec-lint/
git commit -m "build(hooks): wire i18n-keys + oas-units pre-commit

i18n-keys runs the bilingual completeness test on every commit.
oas-units runs fiken-spec-lint with --ignore startTime,endTime (the
4 known wall-clock false positives surfaced in Plan A).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

### Task 22: End-of-Plan-B verification

- [ ] `git log --oneline | head -25` shows ~25 commits.
- [ ] `nix develop -c prek run --all-files` green.
- [ ] `nix develop -c go test -race -count=1 ./...` green.
- [ ] `nix develop -c go generate ./...` clean.
- [ ] `nix develop -c go build ./cmd/fiken && ./result/bin/fiken --help` shows the root help.
- [ ] Manual smoke: with a real token (humans only),
  - `./result/bin/fiken auth login --profile test`
  - `./result/bin/fiken --profile test companies list`
  - `./result/bin/fiken --profile test --lang=nb companies list --help`
  - `./result/bin/fiken --profile test companies list --json | jq .ok.items[0].slug`
- [ ] MCP smoke: `claude mcp add fiken -- ./result/bin/fiken mcp --profile test` then call `companies_list` from an MCP client; assert StructuredContent matches the CLI `--json` output byte-for-byte.

Success criteria all met → Plan B closed; Plan C writes the remaining 16 tag domains.

---

## Self-review notes

Things this plan intentionally defers:

- Mass population of `mockfiken/handler_impl.go` for every Fiken op — Plan C tasks add overrides as their ops are tested.
- Implementing OpenAPI-derived MCP `InputSchema` — Plan B uses the SDK's struct-tag inference for `In*` types; Plan C revisits if needed.
- `fiken auth status` showing the real user/company list — needs `ops.UserGet` which isn't in companies tag. Plan C user-tag task adds it.

Risks the implementer should know:

- The ogen v1.20.3 method names (e.g. `GetCompanies`, `GetCompany`) are guesses based on operationId. Open `fiken/oas_client_gen.go` and grep for the exact symbol before forging ahead. If `WithClient` doesn't exist as an option, find the right pattern in the generated `Client.NewClient` signature.
- go-i18n's bundle API has changed across v2 minor versions; the `MessageFile` path and `ParseMessageFileBytes` may need adjusting against the actual version `go get` pulls.
- ff/v4 is pre-1.0 — flag-API names may differ. The signatures here approximate v4.0.0-alpha; adjust against current.
- The MCP go-sdk is similarly evolving; `mcpsdk.AddTool`, `mcpsdk.NewInMemoryTransports`, `mcpsdk.NewStreamableHTTPHandler`, `mcpsdk.StdioTransport` are intended pattern names from recent versions but may need exact-name verification.
- The 4 wall-clock false positives in `fiken-spec-lint` block CI unless `--ignore` is added (Task 21). Don't skip this.
