package ops

import (
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// newConcurrencyRT enforces at most one in-flight request at a time.
// Fiken's stated rate rule: one concurrent request per user.
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
	release := sync.OnceFunc(func() { <-rt.sem })
	resp, err := rt.base.RoundTrip(r)
	if err != nil {
		release()
		return resp, err
	}
	// Fiken counts the request as in flight until the body is streamed,
	// so the slot travels with the body.
	resp.Body = &releaseOnClose{ReadCloser: resp.Body, release: release}
	return resp, nil
}

type releaseOnClose struct {
	io.ReadCloser
	release func()
}

func (b *releaseOnClose) Close() error {
	defer b.release()
	return b.ReadCloser.Close()
}

// newBackoffRT wraps base with 429-aware retry: honors Retry-After
// when present, else exponential with full jitter capped at 4s.
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
	const baseDelay = 250 * time.Millisecond
	const maxDelay = 4 * time.Second
	d := min(baseDelay*(1<<attempt), maxDelay)
	//nolint:gosec // jitter only, non-cryptographic
	return rand.N(d)
}
