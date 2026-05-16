package ops

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
)

// Client wraps the ogen-generated fiken.Client with auth, rate
// limiting, and error mapping.
type Client struct {
	gen   *fiken.Client
	auth  auth.Source
	defCo string
}

// Options configures a new Client.
type Options struct {
	BaseURL string // override; production = "https://api.fiken.no/api/v2"
	Auth    auth.Source
	Company string // default company slug
}

// New returns a Client wired with the concurrency + backoff
// RoundTrippers. Token injection rides on the ogen-generated
// SecuritySource (OAuth bearer); the HTTP transport stack handles
// concurrency cap + 429 backoff. A logging RoundTripper sits at the
// outermost layer so it observes the final response (and only logs
// when slog's default handler is configured at INFO or below).
func New(_ context.Context, opts Options) (*Client, error) {
	transport := newLoggingRT(newBackoffRT(newConcurrencyRT(http.DefaultTransport), 3))
	httpClient := &http.Client{Transport: transport}

	url := opts.BaseURL
	if url == "" {
		url = "https://api.fiken.no/api/v2"
	}

	gen, err := fiken.NewClient(url, &securitySource{src: opts.Auth}, fiken.WithClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &Client{gen: gen, auth: opts.Auth, defCo: opts.Company}, nil
}

// securitySource adapts auth.Source to fiken.SecuritySource.
type securitySource struct {
	src auth.Source
}

// FikenAPIOAuth resolves the bearer token from the configured
// auth.Source. The ogen client injects it as `Authorization: Bearer
// <token>` per operation.
func (s *securitySource) FikenAPIOAuth(ctx context.Context, _ fiken.OperationName) (fiken.FikenAPIOAuth, error) {
	if s.src == nil {
		return fiken.FikenAPIOAuth{}, auth.ErrNotFound
	}
	tok, err := s.src.Token(ctx)
	if err != nil {
		return fiken.FikenAPIOAuth{}, err
	}
	return fiken.FikenAPIOAuth{Token: tok}, nil
}

// loggingRT emits an slog INFO event per outgoing request. Sensitive
// request headers (notably Authorization) are never read or logged;
// only method, path, status, duration, and Fiken's X-Request-Id (when
// present) reach the log line.
type loggingRT struct {
	base http.RoundTripper
}

func newLoggingRT(base http.RoundTripper) http.RoundTripper {
	return &loggingRT{base: base}
}

func (rt *loggingRT) RoundTrip(r *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := rt.base.RoundTrip(r)
	dur := time.Since(start)

	attrs := []any{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Duration("dur", dur),
	}
	if resp != nil {
		attrs = append(attrs, slog.Int("status", resp.StatusCode))
		if rid := resp.Header.Get("X-Request-Id"); rid != "" {
			attrs = append(attrs, slog.String("request_id", rid))
		}
	}
	if err != nil {
		attrs = append(attrs, slog.String("err", err.Error()))
	}
	slog.Info("fiken request", attrs...)
	return resp, err
}
