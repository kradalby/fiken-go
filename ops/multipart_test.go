package ops

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestOpenMultipartFile asserts the helper opens the file, populates
// Name and Size from the filesystem, and returns a Close that lets the
// caller release the handle. The contents are streamed via the File
// reader so we read them back to confirm the wiring.
func TestOpenMultipartFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "attachment.pdf")
	content := []byte("hello-multipart-file")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	opt, closer, err := OpenMultipartFile(path, "")
	if err != nil {
		t.Fatalf("OpenMultipartFile: %v", err)
	}
	defer func() {
		if cerr := closer(); cerr != nil {
			t.Errorf("close: %v", cerr)
		}
	}()

	if !opt.Set {
		t.Fatal("opt.Set is false, want true")
	}
	if opt.Value.Name != "attachment.pdf" {
		t.Errorf("Name=%q want %q", opt.Value.Name, "attachment.pdf")
	}
	if opt.Value.Size != int64(len(content)) {
		t.Errorf("Size=%d want %d", opt.Value.Size, len(content))
	}

	got, err := io.ReadAll(opt.Value.File)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q want %q", got, content)
	}
}

// TestOpenMultipartFileExplicitName verifies the override-name path
// wins over filepath.Base(path).
func TestOpenMultipartFileExplicitName(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "real.bin")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	opt, closer, err := OpenMultipartFile(path, "override.bin")
	if err != nil {
		t.Fatalf("OpenMultipartFile: %v", err)
	}
	defer func() { _ = closer() }()

	if opt.Value.Name != "override.bin" {
		t.Errorf("Name=%q want %q", opt.Value.Name, "override.bin")
	}
}

// TestOpenMultipartFileMissing asserts opening a missing path returns
// a non-nil err and the no-op closer (calling it is safe).
func TestOpenMultipartFileMissing(t *testing.T) {
	t.Parallel()

	_, closer, err := OpenMultipartFile(filepath.Join(t.TempDir(), "nope"), "")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if cerr := closer(); cerr != nil {
		t.Errorf("no-op closer returned %v", cerr)
	}
}
