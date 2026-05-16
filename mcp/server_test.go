package mcp

import (
	"context"
	"encoding/json"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/mockfiken"
	"github.com/kradalby/fiken-go/ops"
)

// newTestSession spins up an MCP server backed by ops + mockfiken and
// returns a connected client session ready for CallTool. The session
// + transports auto-close via t.Cleanup.
func newTestSession(t *testing.T, mode Mode) (*mcpsdk.ClientSession, *mockfiken.Server) {
	t.Helper()
	mock := mockfiken.New(t)
	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	bundle := i18n.MustLoad()
	srv, err := New(Options{Client: client, Mode: mode, Bundle: bundle, Lang: "en"})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	// Server.Connect spawns a session goroutine internally; we keep
	// the returned session alive for the test's lifetime via cleanup.
	srvSess, err := srv.Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatalf("server Connect: %v", err)
	}
	t.Cleanup(func() { _ = srvSess.Close() })

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "0.1"}, nil)
	cliSess, err := cs.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client Connect: %v", err)
	}
	t.Cleanup(func() { _ = cliSess.Close() })
	return cliSess, mock
}

// decodeResult re-marshals StructuredContent (any) into the
// strongly-typed envelope. The SDK populates StructuredContent on the
// server with the typed Out value; on the wire it becomes JSON, and
// the client side deserializes it as `any`. Round-tripping is the
// simplest way to recover the original shape from a test.
func decodeResult[T any](t *testing.T, resp *mcpsdk.CallToolResult) ops.Result[T] {
	t.Helper()
	if resp.StructuredContent == nil {
		t.Fatalf("nil StructuredContent")
	}
	b, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	var got ops.Result[T]
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal Result: %v (raw=%s)", err, b)
	}
	return got
}

func TestCompaniesListTool(t *testing.T) {
	cliSess, mock := newTestSession(t, ModeReadOnly)
	mock.Set(ops.OpCompaniesList, &fiken.GetCompaniesOKHeaders{
		Response: []fiken.Company{{
			Slug: fiken.OptString{Value: "acme", Set: true},
			Name: fiken.OptString{Value: "Acme Co", Set: true},
		}},
	})

	resp, err := cliSess.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      ops.OpCompaniesList,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("tool errored: %+v", resp)
	}

	got := decodeResult[ops.CompaniesListOut](t, resp)
	if got.Ok == nil {
		t.Fatalf("Ok nil; envelope shape wrong: %+v", got)
	}
	if len(got.Ok.Items) != 1 || got.Ok.Items[0].Slug != "acme" {
		t.Fatalf("unexpected items: %+v", got.Ok.Items)
	}
}

func TestCompaniesGetTool(t *testing.T) {
	cliSess, mock := newTestSession(t, ModeReadOnly)
	mock.Set(ops.OpCompaniesGet, &fiken.Company{
		Slug: fiken.OptString{Value: "acme", Set: true},
		Name: fiken.OptString{Value: "Acme Co", Set: true},
	})

	resp, err := cliSess.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      ops.OpCompaniesGet,
		Arguments: map[string]any{"company": "acme"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("tool errored: %+v", resp)
	}

	got := decodeResult[ops.CompanyOut](t, resp)
	if got.Ok == nil || got.Ok.Slug != "acme" {
		t.Fatalf("Ok mismatch: %+v", got)
	}
}

// TestReadOnlyMode covers AllowOp's gate via Registry without spinning
// a full server. The companies_* ops are non-mutating so they pass
// even in read-only mode.
func TestReadOnlyMode(t *testing.T) {
	cases := []struct {
		mode Mode
		op   string
		want bool
	}{
		{ModeReadOnly, ops.OpCompaniesList, true},
		{ModeReadOnly, ops.OpCompaniesGet, true},
		{ModeReadOnly, "unknown_op", false}, // fail-closed
		{ModeReadWrite, "unknown_op", true},
	}
	for _, tc := range cases {
		if got := AllowOp(tc.mode, tc.op); got != tc.want {
			t.Errorf("AllowOp(%v, %q) = %v, want %v", tc.mode, tc.op, got, tc.want)
		}
	}
}

// TestListTools verifies tool registration is gated by Mode at New
// time — read-only mode still exposes the two non-mutating companies
// ops.
func TestListTools(t *testing.T) {
	cliSess, _ := newTestSession(t, ModeReadOnly)
	got, err := cliSess.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range got.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{ops.OpCompaniesList, ops.OpCompaniesGet} {
		if !names[want] {
			t.Errorf("missing tool %q in %v", want, names)
		}
	}
}
