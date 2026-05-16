package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/kradalby/fiken-go/auth"
	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/i18n"
	mcppkg "github.com/kradalby/fiken-go/mcp"
	"github.com/kradalby/fiken-go/mockfiken"
	"github.com/kradalby/fiken-go/ops"
)

// TestParityCompaniesList asserts CLI --json bytes ≈ MCP
// StructuredContent for the same input + mock state.
func TestParityCompaniesList(t *testing.T) {
	mock := mockfiken.New(t)
	// Deterministic response: no overrides → both paths see same
	// empty CompaniesListOut.

	// CLI path
	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"companies", "list",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	// MCP path
	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      ops.OpCompaniesList,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}

	// resp.StructuredContent is `any` (decoded JSON). Marshal back
	// to canonical bytes for comparison.
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}

	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityContactsList exercises CLI vs. MCP for contacts_list on
// the same mock state (empty result).
func TestParityContactsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"contacts", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpContactsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}

	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityContactsGet exercises CLI vs. MCP for the single-contact
// path against a Contact override.
func TestParityContactsGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpContactsGet, &fiken.Contact{
		ContactId: fiken.OptInt64{Value: 42, Set: true},
		Name:      "Acme Co",
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"contacts", "get",
		"--company", "acme",
		"--contact-id", "42",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpContactsGet,
		Arguments: map[string]any{
			"company":    "acme",
			"contact_id": 42,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityAccountsList exercises CLI vs. MCP for accounts_list on
// the same mock state (empty result).
func TestParityAccountsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"accounts", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpAccountsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}

	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityAccountsGet exercises CLI vs. MCP for the single-account
// path against an Account override.
func TestParityAccountsGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpAccountsGet, &fiken.Account{
		Code: fiken.OptString{Value: "3020", Set: true},
		Name: fiken.OptString{Value: "Salgsinntekt", Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"accounts", "get",
		"--company", "acme",
		"--account-code", "3020",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpAccountsGet,
		Arguments: map[string]any{
			"company":      "acme",
			"account_code": "3020",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityBankAccountsList exercises CLI vs. MCP for
// bank_accounts_list on the same mock state (empty result).
func TestParityBankAccountsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"bank-accounts", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpBankAccountsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}

	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityBankAccountsGet exercises CLI vs. MCP for the
// single-resource path against a BankAccountResult override.
func TestParityBankAccountsGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpBankAccountsGet, &fiken.BankAccountResult{
		BankAccountId: fiken.OptInt64{Value: 2747365, Set: true},
		Name:          fiken.OptString{Value: "Driftskonto", Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"bank-accounts", "get",
		"--company", "acme",
		"--bank-account-id", "2747365",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpBankAccountsGet,
		Arguments: map[string]any{
			"company":         "acme",
			"bank_account_id": 2747365,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityJournalEntriesList exercises CLI vs. MCP for
// journal_entries_list on the same mock state (empty result).
func TestParityJournalEntriesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"journal-entries", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpJournalEntriesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}

	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityJournalEntriesGet exercises CLI vs. MCP for the
// single-resource path against a JournalEntry override.
func TestParityJournalEntriesGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpJournalEntriesGet, &fiken.JournalEntry{
		JournalEntryId: fiken.OptInt64{Value: 1234, Set: true},
		Description:    "Parity fetch",
		Date:           time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"journal-entries", "get",
		"--company", "acme",
		"--journal-entry-id", "1234",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpJournalEntriesGet,
		Arguments: map[string]any{
			"company":          "acme",
			"journal_entry_id": 1234,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityTransactionsList exercises CLI vs. MCP for transactions_list
// on the same mock state (empty result).
func TestParityTransactionsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"transactions", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpTransactionsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}

	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityTransactionsGet exercises CLI vs. MCP for the
// single-resource path against a Transaction override.
func TestParityTransactionsGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpTransactionsGet, &fiken.Transaction{
		TransactionId: fiken.OptInt64{Value: 9988, Set: true},
		Description:   fiken.OptString{Value: "Parity fetch", Set: true},
		Type:          fiken.OptString{Value: "General Journal Entry", Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"transactions", "get",
		"--company", "acme",
		"--transaction-id", "9988",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpTransactionsGet,
		Arguments: map[string]any{
			"company":        "acme",
			"transaction_id": 9988,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityCompaniesGet for the single-resource op.
func TestParityCompaniesGet(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"companies", "get",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      ops.OpCompaniesGet,
		Arguments: map[string]any{"company": "acme"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}

	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityInvoicesList exercises CLI vs. MCP for invoices_list on
// the same mock state (empty result).
func TestParityInvoicesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"invoices", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}

	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()

	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpInvoicesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityInvoicesGet exercises CLI vs. MCP for the single-resource
// invoices_get path against an InvoiceResult override.
func TestParityInvoicesGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpInvoicesGet, &fiken.InvoiceResult{
		InvoiceId:     fiken.OptInt64{Value: 1234, Set: true},
		InvoiceNumber: fiken.OptInt64{Value: 42, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"invoices", "get",
		"--company", "acme",
		"--invoice-id", "1234",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpInvoicesGet,
		Arguments: map[string]any{
			"company":    "acme",
			"invoice_id": 1234,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityInvoiceDraftsList exercises CLI vs. MCP for
// invoices_drafts_list on the same mock state (empty result).
func TestParityInvoiceDraftsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"invoices", "drafts", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpInvoicesDraftsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityCreditNotesList exercises CLI vs. MCP for credit_notes_list
// on the same mock state (empty result).
func TestParityCreditNotesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"credit-notes", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpCreditNotesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityCreditNotesGet exercises CLI vs. MCP for the
// single-resource credit_notes_get path against a CreditNoteResult
// override.
func TestParityCreditNotesGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpCreditNotesGet, &fiken.CreditNoteResult{
		CreditNoteId:     7,
		CreditNoteNumber: 10001,
		Customer:         fiken.Contact{Name: "Acme Buyer"},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"credit-notes", "get",
		"--company", "acme",
		"--credit-note-id", "7",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpCreditNotesGet,
		Arguments: map[string]any{
			"company":        "acme",
			"credit_note_id": "7",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityCreditNoteDraftsList exercises CLI vs. MCP for
// credit_notes_drafts_list on the same mock state (empty result).
func TestParityCreditNoteDraftsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"credit-notes", "drafts", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpCreditNotesDraftsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityInvoiceDraftsGet exercises CLI vs. MCP for the
// single-resource invoices_drafts_get path against an InvoiceishDraftResult
// override.
func TestParityInvoiceDraftsGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpInvoicesDraftsGet, &fiken.InvoiceishDraftResult{
		DraftId: fiken.OptInt64{Value: 99, Set: true},
		UUID:    fiken.OptString{Value: "deadbeef-1111-2222-3333-444444444444", Set: true},
		Type:    fiken.OptInvoiceishDraftResultType{Value: fiken.InvoiceishDraftResultTypeInvoice, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"invoices", "drafts", "get",
		"--company", "acme",
		"--draft-id", "99",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpInvoicesDraftsGet,
		Arguments: map[string]any{
			"company":  "acme",
			"draft_id": 99,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityOffersList exercises CLI vs. MCP for offers_list on the
// same mock state (empty result).
func TestParityOffersList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"offers", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpOffersList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityOffersGet exercises CLI vs. MCP for the single-resource
// offers_get path against an Offer override.
func TestParityOffersGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpOffersGet, &fiken.Offer{
		OfferId:     fiken.OptInt64{Value: 11, Set: true},
		OfferNumber: fiken.OptInt32{Value: 10001, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"offers", "get",
		"--company", "acme",
		"--offer-id", "11",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpOffersGet,
		Arguments: map[string]any{
			"company":  "acme",
			"offer_id": "11",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityOfferDraftsList exercises CLI vs. MCP for offers_drafts_list
// on the same mock state (empty result).
func TestParityOfferDraftsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"offers", "drafts", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpOffersDraftsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityOrderConfirmationsList exercises CLI vs. MCP for
// order_confirmations_list on the same mock state (empty result).
func TestParityOrderConfirmationsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"order-confirmations", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpOrderConfirmationsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityOrderConfirmationsGet exercises CLI vs. MCP for the
// single-resource order_confirmations_get path against an
// OrderConfirmation override.
func TestParityOrderConfirmationsGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpOrderConfirmationsGet, &fiken.OrderConfirmation{
		ConfirmationId:     fiken.OptInt64{Value: 11, Set: true},
		ConfirmationNumber: fiken.OptInt32{Value: 10001, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"order-confirmations", "get",
		"--company", "acme",
		"--confirmation-id", "11",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpOrderConfirmationsGet,
		Arguments: map[string]any{
			"company":         "acme",
			"confirmation_id": "11",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityOrderConfirmationDraftsList exercises CLI vs. MCP for
// order_confirmations_drafts_list on the same mock state (empty
// result).
func TestParityOrderConfirmationDraftsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"order-confirmations", "drafts", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpOrderConfirmationsDraftsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityProductsList exercises CLI vs. MCP for products_list on
// the same mock state (empty result).
func TestParityProductsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"products", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpProductsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityProductsGet exercises CLI vs. MCP for products_get
// against a Product override.
func TestParityProductsGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpProductsGet, &fiken.Product{
		ProductId:     fiken.OptInt64{Value: 7, Set: true},
		Name:          "Spade",
		IncomeAccount: "3000",
		VatType:       "HIGH",
		Active:        true,
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"products", "get",
		"--company", "acme",
		"--product-id", "7",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpProductsGet,
		Arguments: map[string]any{
			"company":    "acme",
			"product_id": 7,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParitySalesList exercises CLI vs. MCP for sales_list on the
// same mock state (empty result).
func TestParitySalesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"sales", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpSalesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParitySalesGet exercises CLI vs. MCP for sales_get against a
// SaleResult override.
func TestParitySalesGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpSalesGet, &fiken.SaleResult{
		SaleId:     fiken.OptInt64{Value: 42, Set: true},
		SaleNumber: fiken.OptString{Value: "XK455L", Set: true},
		NetAmount:  fiken.OptInt64{Value: 1000000, Set: true},
		VatAmount:  fiken.OptInt64{Value: 250000, Set: true},
		Customer: fiken.OptContact{Value: fiken.Contact{
			ContactId: fiken.OptInt64{Value: 88, Set: true},
			Name:      "Acme Buyer",
		}, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"sales", "get",
		"--company", "acme",
		"--sale-id", "42",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpSalesGet,
		Arguments: map[string]any{
			"company": "acme",
			"sale_id": 42,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityPurchasesList exercises CLI vs. MCP for purchases_list on
// the same mock state (empty result).
func TestParityPurchasesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"purchases", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpPurchasesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityPurchasesGet exercises CLI vs. MCP for purchases_get
// against a PurchaseResult override.
func TestParityPurchasesGet(t *testing.T) {
	mock := mockfiken.New(t)
	purchaseDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	mock.Set(ops.OpPurchasesGet, &fiken.PurchaseResult{
		PurchaseId: fiken.OptInt64{Value: 99, Set: true},
		Identifier: fiken.OptString{Value: "INV-99", Set: true},
		Date:       purchaseDate,
		Kind:       fiken.PurchaseResultKindSupplier,
		Currency:   "NOK",
		Supplier: fiken.OptContact{Value: fiken.Contact{
			ContactId: fiken.OptInt64{Value: 77, Set: true},
			Name:      "Acme Supplier",
		}, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"purchases", "get",
		"--company", "acme",
		"--purchase-id", "99",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpPurchasesGet,
		Arguments: map[string]any{
			"company":     "acme",
			"purchase_id": 99,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityInboxList exercises CLI vs. MCP for inbox_list on the
// same mock state (empty result).
func TestParityInboxList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"inbox", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpInboxList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityProjectsList exercises CLI vs. MCP for projects_list on
// the same mock state (empty result).
func TestParityProjectsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"projects", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpProjectsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityProjectsGet exercises CLI vs. MCP for projects_get against
// a ProjectResult override that exercises the embedded Contact and
// the Date round-trip.
func TestParityProjectsGet(t *testing.T) {
	mock := mockfiken.New(t)
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	mock.Set(ops.OpProjectsGet, &fiken.ProjectResult{
		ProjectId: fiken.OptInt64{Value: 42, Set: true},
		Number:    fiken.OptString{Value: "P-001", Set: true},
		Name:      fiken.OptString{Value: "Roadrunner Hunt", Set: true},
		StartDate: fiken.OptDate{Value: start, Set: true},
		Contact: fiken.OptContact{Value: fiken.Contact{
			ContactId: fiken.OptInt64{Value: 7, Set: true},
			Name:      "Wile E. Coyote",
		}, Set: true},
		Completed: fiken.OptBool{Value: false, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"projects", "get",
		"--company", "acme",
		"--project-id", "42",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpProjectsGet,
		Arguments: map[string]any{
			"company":    "acme",
			"project_id": 42,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityInboxGet exercises CLI vs. MCP for inbox_get against an
// InboxResult override.
func TestParityInboxGet(t *testing.T) {
	mock := mockfiken.New(t)
	createdAt := time.Date(2024, 6, 1, 12, 30, 0, 0, time.UTC)
	mock.Set(ops.OpInboxGet, &fiken.InboxResult{
		DocumentId: fiken.OptInt64{Value: 42, Set: true},
		Name:       fiken.OptString{Value: "Invoice for August", Set: true},
		Filename:   fiken.OptString{Value: "invoice.pdf", Set: true},
		Status:     fiken.OptBool{Value: false, Set: true},
		CreatedAt:  fiken.OptDateTime{Value: createdAt, Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"inbox", "get",
		"--company", "acme",
		"--document-id", "42",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpInboxGet,
		Arguments: map[string]any{
			"company":     "acme",
			"document_id": 42,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityUserGet exercises CLI vs. MCP for user_get against a
// Userinfo override. user_get is the only registered op that takes no
// company slug, so the MCP InputSchema and the CLI flag-set are both
// minimal — this test guards that asymmetry.
func TestParityUserGet(t *testing.T) {
	mock := mockfiken.New(t)
	mock.Set(ops.OpUserGet, &fiken.Userinfo{
		Name:  fiken.OptString{Value: "Test Testesen", Set: true},
		Email: fiken.OptString{Value: "test@fiken.no", Set: true},
	})

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"user", "me",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name:      ops.OpUserGet,
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityAccountBalancesList exercises CLI vs. MCP for
// account_balances_list against an empty mock state. The op requires
// a date, which both surfaces forward verbatim.
func TestParityAccountBalancesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"account-balances", "list",
		"--company", "acme",
		"--date", "2024-12-31",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpAccountBalancesList,
		Arguments: map[string]any{
			"company": "acme",
			"date":    "2024-12-31",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityBankBalancesList exercises CLI vs. MCP for bank_balances_list
// against an empty mock state.
func TestParityBankBalancesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"bank-balances", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpBankBalancesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityGroupsList exercises CLI vs. MCP for groups_list against
// an empty mock state.
func TestParityGroupsList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"groups", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpGroupsList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityActivitiesList exercises CLI vs. MCP for activities_list
// against an empty mock state.
func TestParityActivitiesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"activities", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpActivitiesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityTimeEntriesList exercises CLI vs. MCP for time_entries_list
// against an empty mock state.
func TestParityTimeEntriesList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"time-entries", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpTimeEntriesList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// TestParityTimeUsersList exercises CLI vs. MCP for time_users_list
// against an empty mock state.
func TestParityTimeUsersList(t *testing.T) {
	mock := mockfiken.New(t)

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"time-users", "list",
		"--company", "acme",
	})
	if err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: ops.OpTimeUsersList,
		Arguments: map[string]any{
			"company": "acme",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

// === Plan D / 21-op tail parity coverage ===

// runReadParityTest is a thin helper for the new tail tests: it
// boots the same CLI + MCP plumbing as the longhand tests above,
// then asserts CLI --json bytes ≈ MCP StructuredContent.
func runReadParityTest(t *testing.T, opName string, cliArgs []string, mcpArgs map[string]any, register func(mock *mockfiken.Server)) {
	t.Helper()
	mock := mockfiken.New(t)
	if register != nil {
		register(mock)
	}

	t.Setenv("FIKEN_TOKEN", "test")
	t.Setenv("FIKEN_API_URL", mock.URL())

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	args := append([]string{"--json", "--config", "/dev/null"}, cliArgs...)
	if err := cmd.ParseAndRun(context.Background(), args); err != nil {
		t.Fatalf("CLI ParseAndRun: %v", err)
	}
	cliBytes := stdout.Bytes()

	client, err := ops.New(context.Background(), ops.Options{
		BaseURL: mock.URL(),
		Auth:    auth.FlagSource{Value: "test"},
	})
	if err != nil {
		t.Fatalf("ops.New: %v", err)
	}
	srv, err := mcppkg.New(mcppkg.Options{
		Client: client, Mode: mcppkg.ModeReadOnly,
		Bundle: i18n.MustLoad(), Lang: "en",
	})
	if err != nil {
		t.Fatalf("mcp.New: %v", err)
	}
	clientT, serverT := mcpsdk.NewInMemoryTransports()
	go func() { _ = srv.Run(context.Background(), serverT) }()
	cs := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "parity-test", Version: "0.1"}, nil)
	cliSession, err := cs.Connect(context.Background(), clientT, nil)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	resp, err := cliSession.CallTool(context.Background(), &mcpsdk.CallToolParams{
		Name: opName, Arguments: mcpArgs,
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if resp.IsError {
		t.Fatalf("MCP tool returned error: %+v", resp)
	}
	mcpBytes, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	if !jsonEqual(cliBytes, mcpBytes) {
		t.Fatalf("envelope mismatch:\nCLI:  %s\nMCP:  %s", cliBytes, mcpBytes)
	}
}

func TestParityInvoicesCounterGet(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpInvoicesCounterGet,
		[]string{"invoices", "counter", "get", "--company", "acme"},
		map[string]any{"company": "acme"},
		func(mock *mockfiken.Server) {
			mock.Set(ops.OpInvoicesCounterGet, &fiken.Counter{
				Value: fiken.OptInt32{Value: 4242, Set: true},
			})
		},
	)
}

func TestParityCreditNotesCounterGet(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpCreditNotesCounterGet,
		[]string{"credit-notes", "counter", "get", "--company", "acme"},
		map[string]any{"company": "acme"},
		func(mock *mockfiken.Server) {
			mock.Set(ops.OpCreditNotesCounterGet, &fiken.Counter{
				Value: fiken.OptInt32{Value: 17, Set: true},
			})
		},
	)
}

func TestParityOffersCounterGet(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpOffersCounterGet,
		[]string{"offers", "counter", "get", "--company", "acme"},
		map[string]any{"company": "acme"},
		func(mock *mockfiken.Server) {
			mock.Set(ops.OpOffersCounterGet, &fiken.Counter{
				Value: fiken.OptInt32{Value: 31, Set: true},
			})
		},
	)
}

func TestParityOrderConfirmationsCounterGet(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpOrderConfirmationsCounterGet,
		[]string{"order-confirmations", "counter", "get", "--company", "acme"},
		map[string]any{"company": "acme"},
		func(mock *mockfiken.Server) {
			mock.Set(ops.OpOrderConfirmationsCounterGet, &fiken.Counter{
				Value: fiken.OptInt32{Value: 5, Set: true},
			})
		},
	)
}

func TestParitySalesDraftsList(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpSalesDraftsList,
		[]string{"sales", "drafts", "list", "--company", "acme"},
		map[string]any{"company": "acme"},
		nil,
	)
}

func TestParitySalesDraftsGet(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpSalesDraftsGet,
		[]string{"sales", "drafts", "get", "--company", "acme", "--draft-id", "77"},
		map[string]any{"company": "acme", "draft_id": 77},
		func(mock *mockfiken.Server) {
			mock.Set(ops.OpSalesDraftsGet, &fiken.DraftResult{
				DraftId: fiken.OptInt64{Value: 77, Set: true},
				UUID:    fiken.OptString{Value: "s-77", Set: true},
			})
		},
	)
}

func TestParityPurchasesDraftsList(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpPurchasesDraftsList,
		[]string{"purchases", "drafts", "list", "--company", "acme"},
		map[string]any{"company": "acme"},
		nil,
	)
}

func TestParityPurchasesDraftsGet(t *testing.T) {
	runReadParityTest(
		t,
		ops.OpPurchasesDraftsGet,
		[]string{"purchases", "drafts", "get", "--company", "acme", "--draft-id", "55"},
		map[string]any{"company": "acme", "draft_id": 55},
		func(mock *mockfiken.Server) {
			mock.Set(ops.OpPurchasesDraftsGet, &fiken.DraftResult{
				DraftId: fiken.OptInt64{Value: 55, Set: true},
				UUID:    fiken.OptString{Value: "p-55", Set: true},
			})
		},
	)
}
