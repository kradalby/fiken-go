package ops

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kradalby/fiken-go/auth"
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

// TestBackoffReplaysBody: a retry must resend the full body, not the
// drained reader the first attempt left behind.
func TestBackoffReplaysBody(t *testing.T) {
	const payload = `{"name":"acme"}`
	var mu sync.Mutex
	var got []string
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		got = append(got, string(b))
		n := len(got)
		mu.Unlock()
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))

	c := &http.Client{Transport: newBackoffRT(srv.Client().Transport, 3)}
	req, _ := http.NewRequestWithContext(t.Context(), "POST", srv.URL, strings.NewReader(payload))
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status=%d want 201", resp.StatusCode)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 2 || got[0] != payload || got[1] != payload {
		t.Fatalf("server saw bodies %q, want two copies of %q", got, payload)
	}
}

// TestBackoffOneShotBodyNotRetried: bodies without GetBody (ogen's
// multipart pipes) can't be resent, so the 429 goes back to the caller
// with Fiken's payload intact.
func TestBackoffOneShotBodyNotRetried(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		attempts.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("slow down"))
	}))

	c := &http.Client{Transport: newBackoffRT(srv.Client().Transport, 3)}
	req, _ := http.NewRequestWithContext(t.Context(), "POST", srv.URL, io.MultiReader(strings.NewReader("part")))
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status=%d want 429", resp.StatusCode)
	}
	if b, err := io.ReadAll(resp.Body); err != nil || string(b) != "slow down" {
		t.Fatalf("body=%q err=%v, want Fiken's payload", b, err)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempts=%d want 1", got)
	}
}

// TestBackoffExhaustedKeepsBody: once retries run out the caller still
// needs the last 429's body to report Fiken's error.
func TestBackoffExhaustedKeepsBody(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("slow down"))
	}))

	c := &http.Client{Transport: newBackoffRT(srv.Client().Transport, 1)}
	req, _ := http.NewRequestWithContext(t.Context(), "GET", srv.URL, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status=%d want 429", resp.StatusCode)
	}
	if b, err := io.ReadAll(resp.Body); err != nil || string(b) != "slow down" {
		t.Fatalf("body=%q err=%v, want Fiken's payload", b, err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts=%d want 2", got)
	}
}

// TestAttachRateLimited drives a real multipart upload into a 429: the
// caller gets rate_limited from one intact attempt, not a pipe error
// from a replay.
func TestAttachRateLimited(t *testing.T) {
	var mu sync.Mutex
	var uploadErrs []error
	// Real listener: ops.New dials through http.DefaultTransport.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := readUpload(r)
		mu.Lock()
		uploadErrs = append(uploadErrs, err)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"slow down"}`))
	}))
	t.Cleanup(srv.Close)
	path := filepath.Join(t.TempDir(), "receipt.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	c, err := New(t.Context(), Options{BaseURL: srv.URL, Auth: auth.FlagSource{Value: "test"}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res := c.ContactsAttachmentsAttach(t.Context(), ContactsAttachmentsAttachIn{
		Company: "acme", ContactID: 1, FilePath: path,
	})
	if res.Error == nil || res.Error.Code != CodeRateLimited {
		t.Fatalf("err=%+v want code %q", res.Error, CodeRateLimited)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(uploadErrs) != 1 || uploadErrs[0] != nil {
		t.Fatalf("server attempts=%d upload errors=%v, want one intact upload", len(uploadErrs), uploadErrs)
	}
}

// readUpload drains every part of a multipart request; a truncated
// stream surfaces as an error.
func readUpload(r *http.Request) error {
	mr, err := r.MultipartReader()
	if err != nil {
		return err
	}
	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := io.Copy(io.Discard, p); err != nil {
			return err
		}
	}
}
