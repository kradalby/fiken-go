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
	var inflight int32
	var maxInflight int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
				_ = resp.Body.Close()
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

func TestBackoffHonorsRetryAfter(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	_ = resp.Body.Close()
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Fatalf("backoff too fast: %s (Retry-After=1 should sleep ~1s)", elapsed)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("attempts=%d want 2 (one retry)", got)
	}
}
