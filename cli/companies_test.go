package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/kradalby/fiken-go/mockfiken"
)

func TestCompaniesListSmoke(t *testing.T) {
	mock := mockfiken.New(t)
	t.Setenv("FIKEN_TOKEN", "test")

	var stdout, stderr bytes.Buffer
	cmd, err := Root(&stdout, &stderr)
	if err != nil {
		t.Fatalf("Root: %v", err)
	}

	// Direct the ops.Client to mockfiken via FIKEN_API_URL? We don't
	// have that knob yet. For smoke, just confirm the subcommand
	// parses and Exec runs (it will fail at the HTTP layer pointing
	// at real api.fiken.no, but the parsing path is what we test).
	err = cmd.ParseAndRun(context.Background(), []string{
		"--json",
		"--config", "/dev/null",
		"companies", "list",
	})
	// Acceptable: any error from the network call (we have a fake token).
	// Failure mode we DON'T want: ff parse error.
	if err != nil {
		if strings.Contains(err.Error(), "flag") || strings.Contains(err.Error(), "unknown") {
			t.Fatalf("parse error: %v", err)
		}
		// Network error is fine for smoke.
		_ = mock // mock not actually consulted without BaseURL wiring
	}
}
