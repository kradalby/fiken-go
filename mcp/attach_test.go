package mcp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"tailscale.com/client/tailscale/apitype"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/i18n"
	"github.com/kradalby/fiken-go/ops"
)

// newTsnetSession serves an MCP server the way RunTsnet does (WhoIs cap
// middleware + cap gate) to a peer holding a write grant. Fiken is a
// stand-in that records every uploaded body.
func newTsnetSession(t *testing.T, opts Options) (*mcpsdk.ClientSession, func() [][]byte) {
	t.Helper()
	var mu sync.Mutex
	var uploads [][]byte
	fiken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		uploads = append(uploads, b)
		mu.Unlock()
		w.Header().Set("Location", "https://fiken.test/attachments/1")
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(fiken.Close)

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: fiken.URL,
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	opts.Client = client
	opts.Bundle = i18n.MustLoad()
	opts.Lang = "en"
	srv, err := New(opts)
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	peer := fakeWhoiser{resp: &apitype.WhoIsResponse{CapMap: capMap(true)}}
	base := mcpsdk.NewStreamableHTTPHandler(func(*http.Request) *mcpsdk.Server { return srv }, nil)
	hs := httptest.NewServer(capMiddleware(peer, base))
	t.Cleanup(hs.Close)

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "peer", Version: "0.1"}, nil)
	sess, err := cs.Connect(t.Context(), &mcpsdk.StreamableClientTransport{Endpoint: hs.URL}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess, func() [][]byte {
		mu.Lock()
		defer mu.Unlock()
		return uploads
	}
}

func callAttach(t *testing.T, sess *mcpsdk.ClientSession, path string) *mcpsdk.CallToolResult {
	t.Helper()
	res, err := sess.CallTool(t.Context(), &mcpsdk.CallToolParams{
		Name: ops.OpContactsAttachmentsAttach,
		Arguments: map[string]any{
			"company": "acme", "contact_id": 1, "file_path": path,
			"filename": "", "comment": "",
		},
	})
	if err != nil {
		t.Fatalf("CallTool(%q): %v", path, err)
	}
	return res
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestTsnetAttachConfinedToRoot: file_path comes from the remote peer,
// so it must never reach server-local files outside the attachments
// directory.
func TestTsnetAttachConfinedToRoot(t *testing.T) {
	dir := t.TempDir()
	secret := filepath.Join(dir, "credentials.json")
	writeFile(t, secret, "SECRET-TOKEN")
	attachDir := filepath.Join(dir, "attachments")
	if err := os.Mkdir(attachDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(attachDir, "receipt.pdf"), "%PDF-receipt")
	if err := os.Symlink(secret, filepath.Join(attachDir, "link.json")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	root, err := os.OpenRoot(attachDir)
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}
	t.Cleanup(func() { _ = root.Close() })

	sess, uploads := newTsnetSession(t, Options{
		Mode:              ModeReadWrite,
		EnableAttachments: true,
		AttachmentRoot:    root,
		CapGated:          true,
	})

	for _, p := range []string{secret, "../credentials.json", "link.json"} {
		if res := callAttach(t, sess, p); !res.IsError {
			t.Errorf("file_path %q accepted, want rejected", p)
		}
	}
	if res := callAttach(t, sess, "receipt.pdf"); res.IsError {
		t.Fatalf("file inside root rejected: %+v", res.Content)
	}

	got := uploads()
	if len(got) != 1 || !bytes.Contains(got[0], []byte("%PDF-receipt")) {
		t.Fatalf("uploads=%q, want only receipt.pdf", got)
	}
}

// TestCapGatedAttachmentsRequireRoot: a remote-facing server must not
// start with unconfined attachment reads.
func TestCapGatedAttachmentsRequireRoot(t *testing.T) {
	_, err := New(Options{
		Mode:              ModeReadWrite,
		Bundle:            i18n.MustLoad(),
		Lang:              "en",
		EnableAttachments: true,
		CapGated:          true,
	})
	if err == nil {
		t.Fatal("New accepted cap-gated attachments without AttachmentRoot")
	}
}
