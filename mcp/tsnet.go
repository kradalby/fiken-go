package mcp

import (
	"context"
	"fmt"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"tailscale.com/client/local"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/tailcfg"
	"tailscale.com/tsnet"
)

// TsnetOptions configures a tsnet-backed MCP HTTP listener.
type TsnetOptions struct {
	Hostname string // tailnet device name; required.
	AuthKey  string // pre-auth key; empty reuses StateDir.
	StateDir string // tsnet state directory; required.
}

// RunTsnet starts a tsnet device, listens on port 80 of the tailnet
// interface, and serves the MCP streamable-HTTP transport behind a
// capability gate (see capMiddleware + capGateMiddleware).
//
// Tailnet membership grants implicit read access. Mutating tools
// require an explicit grant under CapName with {"write": true}.
func RunTsnet(ctx context.Context, srv *mcpsdk.Server, opts TsnetOptions) error {
	if opts.Hostname == "" {
		return fmt.Errorf("tsnet: Hostname is required")
	}
	if opts.StateDir == "" {
		return fmt.Errorf("tsnet: StateDir is required")
	}
	ts := &tsnet.Server{
		Hostname: opts.Hostname,
		Dir:      opts.StateDir,
		AuthKey:  opts.AuthKey,
	}
	defer func() { _ = ts.Close() }()

	if _, err := ts.Up(ctx); err != nil {
		return fmt.Errorf("tsnet up: %w", err)
	}
	lc, err := ts.LocalClient()
	if err != nil {
		return fmt.Errorf("tsnet local client: %w", err)
	}

	base := mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return srv }, nil,
	)
	handler := capMiddleware(localWhoiser{lc}, base)

	ln, err := ts.Listen("tcp", ":80")
	if err != nil {
		return fmt.Errorf("tsnet listen :80: %w", err)
	}
	httpSrv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	// #nosec G118 — separate ctx for shutdown so we don't reuse the
	// already-cancelled parent.
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
		defer cancel()
		_ = httpSrv.Shutdown(sctx)
	}()
	return httpSrv.Serve(ln)
}

// whoiser is the slice of *local.Client the cap middleware needs.
// Defining it as an interface lets tests inject a fake without
// spinning up a real tailnet device.
type whoiser interface {
	WhoIs(ctx context.Context, remoteAddr string) (*apitype.WhoIsResponse, error)
}

type localWhoiser struct{ *local.Client }

// capMiddleware resolves the calling tailnet peer via WhoIs, merges any
// capability grants under CapName, and attaches the resulting
// Capability to the request context. WhoIs failures → 401. Read access
// is implicit for any tailnet peer.
func capMiddleware(w whoiser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		who, err := w.WhoIs(r.Context(), r.RemoteAddr)
		if err != nil || who == nil {
			http.Error(rw, "unauthorized", http.StatusUnauthorized)
			return
		}
		merged := Capability{}
		caps, _ := tailcfg.UnmarshalCapJSON[Capability](who.CapMap, tailcfg.PeerCapability(CapName))
		for _, c := range caps {
			if c.Write {
				merged.Write = true
			}
		}
		next.ServeHTTP(rw, r.WithContext(withCap(r.Context(), merged)))
	})
}
