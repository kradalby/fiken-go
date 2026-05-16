package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestComputeSHA256 verifies the helper produces the canonical lowercase hex
// digest used in api/SOURCE.txt so drift on the format is caught early.
func TestComputeSHA256(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "x.yaml")
	if err := os.WriteFile(path, []byte("hello\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, err := computeSHA256(path)
	if err != nil {
		t.Fatalf("computeSHA256: %v", err)
	}
	want := "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03"
	if got != want {
		t.Fatalf("sha mismatch:\n got %q\nwant %q", got, want)
	}
}

// TestRenderSource verifies api/SOURCE.txt rendering is byte-stable.
func TestRenderSource(t *testing.T) {
	var buf bytes.Buffer
	if err := renderSource(&buf, "https://example.com/x.yaml", "2026-05-15T10:00:00Z", "abc123"); err != nil {
		t.Fatalf("renderSource: %v", err)
	}
	got := buf.String()
	want := "source-url: https://example.com/x.yaml\n" +
		"fetched-at: 2026-05-15T10:00:00Z\n" +
		"sha256:     abc123\n"
	if got != want {
		t.Fatalf("render mismatch:\n got %q\nwant %q", got, want)
	}
}
