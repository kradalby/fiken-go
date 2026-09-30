package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

// TestMCPNetworkAttachmentsNeedDir: network peers choose file_path, so
// attachments must not start unconfined on a network transport.
func TestMCPNetworkAttachmentsNeedDir(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"tsnet", []string{"--tsnet", "--tsnet-state-dir", t.TempDir()}},
		{"http", []string{"--transport=http", "--listen=127.0.0.1:0"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := Root(new(bytes.Buffer), new(bytes.Buffer))
			if err != nil {
				t.Fatalf("Root: %v", err)
			}
			// Bounded: a regression would otherwise start serving.
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			args := append([]string{"--config", "/dev/null", "mcp", "--enable-attachments"}, tc.args...)
			err = cmd.ParseAndRun(ctx, args)
			if err == nil || !strings.Contains(err.Error(), "--attachments-dir") {
				t.Fatalf("err=%v, want --attachments-dir required", err)
			}
		})
	}
}
