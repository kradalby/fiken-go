package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kradalby/fiken-go/fiken"
	"github.com/kradalby/fiken-go/ops"
)

// TestReadBodyFileRoundTrip writes a fiken.Contact to a temp file and
// reads it back through ReadBodyFile. The round-trip is the canonical
// happy path; if it breaks, every --from-file subcommand breaks too.
func TestReadBodyFileRoundTrip(t *testing.T) {
	t.Parallel()

	want := fiken.Contact{
		Name:  "Acme",
		Email: fiken.OptString{Value: "acme@example.com", Set: true},
	}
	raw, err := json.Marshal(&want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	path := filepath.Join(t.TempDir(), "contact.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	got, oerr := ReadBodyFile[fiken.Contact](path)
	if oerr != nil {
		t.Fatalf("ReadBodyFile: %v", oerr)
	}
	if got.Name != "Acme" || got.Email.Value != "acme@example.com" {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

// TestReadBodyFileMissingPath asserts a missing file produces a
// validation error rather than a panic or generic internal error.
func TestReadBodyFileMissingPath(t *testing.T) {
	t.Parallel()

	_, oerr := ReadBodyFile[fiken.Contact](filepath.Join(t.TempDir(), "nope.json"))
	if oerr == nil {
		t.Fatal("want error, got nil")
	}
	if oerr.Code != ops.CodeValidation {
		t.Errorf("got Code=%q want %q", oerr.Code, ops.CodeValidation)
	}
}

// TestReadBodyFileInvalidJSON asserts malformed JSON also lands on
// CodeValidation — same surface as missing-path.
func TestReadBodyFileInvalidJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, oerr := ReadBodyFile[fiken.Contact](path)
	if oerr == nil {
		t.Fatal("want error, got nil")
	}
	if oerr.Code != ops.CodeValidation {
		t.Errorf("got Code=%q want %q", oerr.Code, ops.CodeValidation)
	}
}
