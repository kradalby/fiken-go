package ops

import (
	"os"
	"path/filepath"

	ht "github.com/ogen-go/ogen/http"

	"github.com/kradalby/fiken-go/fiken"
)

// OpenMultipartFile opens path for streaming as an ogen multipart form
// part. The returned Close func must be invoked when the request
// completes (typically via defer in the caller) so the underlying
// *os.File is released. name overrides the form-field filename; pass
// "" to use filepath.Base(path).
//
// A non-nil root confines path beneath it: absolute paths, ".." and
// symlink escapes are rejected. nil opens path as given, for callers
// where the local user supplies it.
//
// On open/stat failure the returned Close is a no-op and err is the
// raw os error — callers wrap it with op context via MapErr or build
// their own ops.Error.
func OpenMultipartFile(root *os.Root, path string, name string) (fiken.OptMultipartFile, func() error, error) {
	f, err := openFile(root, path)
	if err != nil {
		return fiken.OptMultipartFile{}, func() error { return nil }, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fiken.OptMultipartFile{}, func() error { return nil }, err
	}
	if name == "" {
		name = filepath.Base(path)
	}
	return fiken.OptMultipartFile{
		Value: ht.MultipartFile{
			Name: name,
			File: f,
			Size: info.Size(),
		},
		Set: true,
	}, f.Close, nil
}

func openFile(root *os.Root, path string) (*os.File, error) {
	if root != nil {
		return root.Open(path)
	}
	return os.Open(path) //nolint:gosec // local caller's own path; remote callers pass a root
}
