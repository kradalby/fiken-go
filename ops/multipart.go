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
// On open/stat failure the returned Close is a no-op and err is the
// raw os error — callers wrap it with op context via MapErr or build
// their own ops.Error.
func OpenMultipartFile(path string, name string) (fiken.OptMultipartFile, func() error, error) {
	f, err := os.Open(path) //nolint:gosec // path is user-supplied by design (CLI attach)
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
