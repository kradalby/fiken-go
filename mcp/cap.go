package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// CapName is the Tailscale capability-grant key under which fiken-mcp
// looks up write permissions. Read access is implicit for any tailnet
// peer; only writes need an explicit grant.
const CapName = "kradalby.no/cap/fiken-mcp"

// Capability is the JSON shape granted via Tailscale ACL grants.
// A peer with no matching grant is treated as Capability{} (read-only).
type Capability struct {
	Write bool `json:"write"`
}

type capCtxKey struct{}

func withCap(ctx context.Context, c Capability) context.Context {
	return context.WithValue(ctx, capCtxKey{}, c)
}

func capFrom(ctx context.Context) (Capability, bool) {
	v, ok := ctx.Value(capCtxKey{}).(Capability)
	return v, ok
}

// capGateMiddleware returns a receiving-middleware that gates tools/call
// requests by the Capability stored in ctx. On all other methods it
// passes through. Denied calls return an in-band CallToolResult with
// IsError=true so the MCP client surfaces the failure without tearing
// down the JSON-RPC session.
func capGateMiddleware(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
	return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
		if method != "tools/call" {
			return next(ctx, method, req)
		}
		params, ok := req.GetParams().(*mcpsdk.CallToolParamsRaw)
		if !ok || params == nil {
			return next(ctx, method, req)
		}
		c, _ := capFrom(ctx) // absent → zero value (read-only)
		if !AllowOpForCap(params.Name, c) {
			return &mcpsdk.CallToolResult{
				IsError: true,
				Content: []mcpsdk.Content{&mcpsdk.TextContent{
					Text: fmt.Sprintf("tool %q denied: write requires capability grant under %q", params.Name, CapName),
				}},
			}, nil
		}
		return next(ctx, method, req)
	}
}
