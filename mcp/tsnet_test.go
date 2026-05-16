package mcp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/tailcfg"
)

// fakeWhoiser implements the whoiser interface for unit-testing
// capMiddleware without spinning up a tsnet device.
type fakeWhoiser struct {
	resp *apitype.WhoIsResponse
	err  error
}

func (f fakeWhoiser) WhoIs(_ context.Context, _ string) (*apitype.WhoIsResponse, error) {
	return f.resp, f.err
}

// capProbe is an http.Handler that records the Capability resolved by
// the middleware so tests can assert on it.
type capProbe struct {
	gotCap   Capability
	gotCapOK bool
}

func (p *capProbe) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	p.gotCap, p.gotCapOK = capFrom(r.Context())
}

func capMap(write bool) tailcfg.PeerCapMap {
	body := "{}"
	if write {
		body = `{"write": true}`
	}
	return tailcfg.PeerCapMap{
		tailcfg.PeerCapability(CapName): []tailcfg.RawMessage{tailcfg.RawMessage(body)},
	}
}

func TestCapMiddleware_WhoIsError(t *testing.T) {
	probe := &capProbe{}
	h := capMiddleware(fakeWhoiser{err: errors.New("nope")}, probe)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if probe.gotCapOK {
		t.Fatalf("downstream handler ran on WhoIs error")
	}
}

func TestCapMiddleware_NoCap_ReadOnly(t *testing.T) {
	probe := &capProbe{}
	h := capMiddleware(fakeWhoiser{resp: &apitype.WhoIsResponse{CapMap: tailcfg.PeerCapMap{}}}, probe)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !probe.gotCapOK {
		t.Fatalf("downstream did not run")
	}
	if probe.gotCap.Write {
		t.Fatalf("Write=true with no grant; want read-only")
	}
}

func TestCapMiddleware_WithWriteGrant(t *testing.T) {
	probe := &capProbe{}
	h := capMiddleware(fakeWhoiser{resp: &apitype.WhoIsResponse{CapMap: capMap(true)}}, probe)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !probe.gotCap.Write {
		t.Fatalf("Write=false despite explicit grant")
	}
}

func TestCapMiddleware_MultiGrantUnion(t *testing.T) {
	cm := tailcfg.PeerCapMap{
		tailcfg.PeerCapability(CapName): []tailcfg.RawMessage{
			tailcfg.RawMessage(`{}`),
			tailcfg.RawMessage(`{"write": true}`),
		},
	}
	probe := &capProbe{}
	h := capMiddleware(fakeWhoiser{resp: &apitype.WhoIsResponse{CapMap: cm}}, probe)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !probe.gotCap.Write {
		t.Fatalf("union of grants did not produce Write=true")
	}
}
