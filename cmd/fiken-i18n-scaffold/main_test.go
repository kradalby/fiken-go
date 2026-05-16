package main

import (
	"strings"
	"testing"
)

// TestRenderEntry produces a TOML stanza for one op. Format:
//
//	[ops.<op>]
//	summary     = "<derived>"
//	when_to_use = ""
//	returns     = ""
//	example     = ""
//
// Keys are quoted; empty values are empty quoted strings.
func TestRenderEntry(t *testing.T) {
	got := renderEntry("invoices_create", "Creates an invoice. This corresponds to \"Ny faktura\".")
	want := `[ops.invoices_create]
summary     = "Creates an invoice. This corresponds to \"Ny faktura\"."
when_to_use = ""
returns     = ""
example     = ""

`
	if got != want {
		t.Fatalf("mismatch:\n got %q\nwant %q", got, want)
	}
}

// TestTruncate caps very long descriptions at the first sentence
// (".") or 200 chars, whichever comes first.
func TestTruncate(t *testing.T) {
	long := strings.Repeat("a ", 200) + "more"
	got := truncateDesc(long)
	if len(got) > 220 {
		t.Fatalf("not truncated: len=%d", len(got))
	}
}
