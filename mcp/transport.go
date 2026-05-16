package mcp

import (
	"context"
	"net/http"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// readHeaderTimeout caps how long the HTTP server waits for request
// headers — guards against Slowloris. 10s is generous for a localhost
// LLM-driver but trims pathological peers.
const readHeaderTimeout = 10 * time.Second

// shutdownGrace bounds the graceful shutdown wait when ctx is
// cancelled — long enough for in-flight tool calls to drain, short
// enough to keep CI hangs at bay.
const shutdownGrace = 10 * time.Second

// RunStdio blocks serving the MCP protocol over stdio.
// Stdin/stdout uses newline-delimited JSON per the SDK contract.
func RunStdio(ctx context.Context, srv *mcpsdk.Server) error {
	return srv.Run(ctx, &mcpsdk.StdioTransport{})
}

// RunHTTP listens on addr and serves streamable HTTP. ctx cancellation
// triggers a clean shutdown; ListenAndServe returns the resulting
// http.ErrServerClosed on success.
func RunHTTP(ctx context.Context, srv *mcpsdk.Server, addr string) error {
	handler := mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return srv }, nil,
	)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	// #nosec G118 — the goroutine's job is to wait for ctx cancel and
	// then start a *fresh* deadline for graceful shutdown. Reusing the
	// already-cancelled parent ctx would short-circuit Shutdown.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()
	return httpSrv.ListenAndServe()
}
