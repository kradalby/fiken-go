package mcp

import (
	"context"
	"net"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/mockfiken"
	"github.com/kradalby/fiken-go/ops"
)

// TestMCPHTTPTransport spins up the MCP server with --transport=http
// on a random port, connects an MCP client over the streamable HTTP
// transport, and calls companies_list against mockfiken.
func TestMCPHTTPTransport(t *testing.T) {
	mock := mockfiken.New(t)
	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := New(Options{
		Client: client,
		Mode:   ModeReadOnly,
		Bundle: i18n.MustLoad(),
		Lang:   "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	// Pick a free port.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := lis.Addr().String()
	_ = lis.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	go func() { _ = RunHTTP(ctx, srv, addr) }()
	time.Sleep(200 * time.Millisecond) // let listener bind

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-http", Version: "0.1"}, nil)
	transport := &mcpsdk.StreamableClientTransport{Endpoint: "http://" + addr}
	session, err := cs.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	resp, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      ops.OpCompaniesList,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("tool returned error: %+v", resp)
	}
}
