package ops

import (
	"context"
	"net/http"
	"net/http/httptest"
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
