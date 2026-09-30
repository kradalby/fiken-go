package ops

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrencyTokenSerializes(t *testing.T) {
	var inflight atomic.Int32
	var maxInflight atomic.Int32
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := inflight.Add(1)
		if n > maxInflight.Load() {
			maxInflight.Store(n)
		}
		time.Sleep(20 * time.Millisecond)
		inflight.Add(-1)
		w.WriteHeader(200)
	}))

	c := &http.Client{Transport: newConcurrencyRT(srv.Client().Transport)}
	const N = 5
	done := make(chan struct{}, N)
	for range N {
		go func() {
			req, _ := http.NewRequest("GET", srv.URL, nil)
			resp, err := c.Do(req)
			if err != nil {
				t.Errorf("do: %v", err)
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
			done <- struct{}{}
		}()
	}
	for range N {
		<-done
	}
	if got := maxInflight.Load(); got != 1 {
		t.Fatalf("maxInflight=%d want 1 (concurrency token broken)", got)
	}
}

func TestBackoffHonorsRetryAfter(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))

	c := &http.Client{Transport: newBackoffRT(srv.Client().Transport, 3)}
	start := time.Now()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", srv.URL, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	_ = resp.Body.Close()
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Fatalf("backoff too fast: %s (Retry-After=1 should sleep ~1s)", elapsed)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts=%d want 2 (one retry)", got)
	}
}

// rtFunc adapts a func to http.RoundTripper.
type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func okBody(r *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok")), Request: r}, nil
}

// blockedFor reports whether a request through rt is still waiting for
// the concurrency slot after d.
func blockedFor(t *testing.T, rt http.RoundTripper, d time.Duration) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), d)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://fiken.test", nil)
	resp, err := rt.RoundTrip(req)
	if resp != nil {
		_ = resp.Body.Close()
	}
	return errors.Is(err, context.DeadlineExceeded)
}

// TestConcurrencyTokenHeldUntilBodyClosed: Fiken counts a request as in
// flight until its response body is done, so the slot must outlive the
// headers.
func TestConcurrencyTokenHeldUntilBodyClosed(t *testing.T) {
	rt := newConcurrencyRT(rtFunc(okBody))

	req, _ := http.NewRequestWithContext(t.Context(), "GET", "http://fiken.test", nil)
	first, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if !blockedFor(t, rt, 50*time.Millisecond) {
		t.Fatal("second request ran while first body still open")
	}

	// Callers (ogen, net/http) may close more than once; only the first
	// close may free the slot.
	closed := make(chan struct{})
	go func() {
		_ = first.Body.Close()
		_ = first.Body.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("double Close blocked")
	}

	second, err := rt.RoundTrip(req.Clone(t.Context()))
	if err != nil {
		t.Fatalf("second after close: %v", err)
	}
	if !blockedFor(t, rt, 50*time.Millisecond) {
		t.Fatal("double Close freed an extra slot")
	}
	_ = second.Body.Close()
}

// TestConcurrencyTokenReleasedOnError: a failed round trip has no body
// to close, so the slot must come back immediately.
func TestConcurrencyTokenReleasedOnError(t *testing.T) {
	var calls atomic.Int32
	rt := newConcurrencyRT(rtFunc(func(r *http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			return nil, errors.New("dial failed")
		}
		return okBody(r)
	}))
	req, _ := http.NewRequestWithContext(t.Context(), "GET", "http://fiken.test", nil)
	if _, err := rt.RoundTrip(req); err == nil {
		t.Fatal("want error from first round trip")
	}
	if blockedFor(t, rt, time.Second) {
		t.Fatal("slot leaked after failed round trip")
	}
}
