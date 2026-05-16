// Package mockfiken is a spec-driven HTTP mock for the Fiken API.
// It wraps the ogen-generated fiken.Handler interface in an
// httptest.Server, returning zero-value responses for the handful of
// operations we implement and ht.ErrNotImplemented (via the embedded
// fiken.UnimplementedHandler) for everything else.
//
// Per-op overrides are registered via Set / SetError and consulted
// by the implemented handlers (currently GetCompanies, GetCompany).
// Operations without a real implementation are accessible only by
// adding a method on handlerImpl that consults the registry.
package mockfiken

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/kradalby/fiken-go/fiken"
)

// defaultErrorHandler is the ogen default — pulled out as a var so the
// override path can fall through cleanly without re-implementing the
// content-type/jx-encoder dance.
var defaultErrorHandler = ogenerrors.DefaultErrorHandler

// Server is a running httptest.Server fronting a fiken.Handler.
//
// Lifecycle is bound to the *testing.T via t.Cleanup; tests should
// not call Close directly unless they want early shutdown.
type Server struct {
	t        testing.TB
	srv      *httptest.Server
	mu       sync.Mutex
	override map[string]any
	errOver  map[string]*errOverride
}

// errOverride is a status+body pair the mock returns instead of the
// default success path. It implements error so handler methods can
// return it directly; mockErrorHandler intercepts it before ogen's
// default error path and writes the configured status + body verbatim.
//
// Body may be []byte (written as-is), a string (written as-is), nil
// (no body), or any other value (json-encoded best-effort, falling
// back to fmt.Sprint).
type errOverride struct {
	status int
	body   any
}

// Error implements error. The string includes the status so logs are
// useful even when something accidentally surfaces this outside the
// mockErrorHandler path.
func (e *errOverride) Error() string {
	return fmt.Sprintf("mockfiken: forced error (status=%d)", e.status)
}

// New starts a mock server bound to t and registers Close on t's
// cleanup hook. The returned Server's URL is what http clients
// should target.
func New(t testing.TB) *Server {
	t.Helper()
	s := &Server{
		t:        t,
		override: map[string]any{},
		errOver:  map[string]*errOverride{},
	}
	handler, err := fiken.NewServer(
		&handlerImpl{server: s},
		&securityHandler{},
		fiken.WithErrorHandler(mockErrorHandler),
	)
	if err != nil {
		t.Fatalf("mockfiken.New: %v", err)
	}
	s.srv = httptest.NewServer(handler)
	t.Cleanup(s.srv.Close)
	return s
}

// URL returns the base URL clients should target. Append the same
// paths the real Fiken API uses (e.g. "/companies").
func (s *Server) URL() string { return s.srv.URL }

// Close shuts the server down. Tests don't normally need this — the
// t.Cleanup hook in New takes care of it.
func (s *Server) Close() { s.srv.Close() }

// Set registers a success-override for the given op-name. The value
// must be the response type the corresponding handler returns
// (e.g. *fiken.GetCompaniesOKHeaders for ops.OpCompaniesList).
func (s *Server) Set(op string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.override[op] = value
}

// SetError registers an error-override for the given op-name.
// Currently consulted by implemented handlers only; status is
// surfaced as the HTTP response code via ogen's error path.
func (s *Server) SetError(op string, status int, body any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errOver[op] = &errOverride{status: status, body: body}
}

// lookup returns any override registered for op. Callers receive
// the success value, the error override, and a hit-flag in that
// order. Callers that get hit=true and err=nil should return the
// success value; hit=true and err non-nil means return an error.
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

// mockErrorHandler intercepts errors returned by handlerImpl methods.
// If the error is an *errOverride, the status + body configured via
// SetError are written verbatim — this is how tests assert that 4xx /
// 5xx codes propagate through the client mapping. All other errors
// fall through to ogen's default handler (which surfaces
// ht.ErrNotImplemented as 501, etc).
func mockErrorHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	var over *errOverride
	if errors.As(err, &over) {
		if over.status == 0 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(over.status)
		}
		switch b := over.body.(type) {
		case nil:
			// no body
		case []byte:
			_, _ = w.Write(b)
		case string:
			_, _ = w.Write([]byte(b))
		default:
			_, _ = fmt.Fprint(w, b)
		}
		return
	}
	// Fallback: ogen's default behaviour.
	defaultErrorHandler(ctx, w, r, err)
}

// securityHandler accepts any Bearer token. ogen rejects requests
// without a Bearer header before security runs, so we don't have to
// validate the token here.
type securityHandler struct{}

// HandleFikenAPIOAuth implements fiken.SecurityHandler.
func (securityHandler) HandleFikenAPIOAuth(ctx context.Context, _ fiken.OperationName, _ fiken.FikenAPIOAuth) (context.Context, error) {
	return ctx, nil
}

// Compile-time assertion that handlerImpl satisfies fiken.Handler.
// We rely on embedded fiken.UnimplementedHandler to supply the ~150
// methods we don't override.
var _ fiken.Handler = (*handlerImpl)(nil)

// handlerImpl embeds fiken.UnimplementedHandler so we only need to
// implement the methods that participate in tests. Everything else
// returns ht.ErrNotImplemented via the embedded type, which ogen's
// default error handler maps to HTTP 501.
type handlerImpl struct {
	fiken.UnimplementedHandler
	server *Server
}
