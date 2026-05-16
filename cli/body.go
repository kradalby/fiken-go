// Package cli — body.go centralizes the --from-file flow used by every
// mutating subcommand. Subcommands that take a JSON request body call
// ReadBodyFile with the target type and the user-supplied path; the
// helper handles read + parse and turns failures into a typed
// ops.Error (CodeValidation) so the CLI's error rendering stays
// uniform with the rest of the ops envelope.
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kradalby/fiken-go/ops"
)

// ReadBodyFile reads JSON from path and unmarshals it into a fresh *T.
// Read or parse failures surface as ops.Error with Code=validation; Op
// is left empty for the caller to populate after the lookup since this
// helper has no per-op context.
func ReadBodyFile[T any](path string) (*T, *ops.Error) {
	raw, err := os.ReadFile(path) //nolint:gosec // path is user-supplied by design (CLI --from-file)
	if err != nil {
		return nil, &ops.Error{
			Code:    ops.CodeValidation,
			Message: fmt.Sprintf("read %s: %v", path, err),
		}
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, &ops.Error{
			Code:    ops.CodeValidation,
			Message: fmt.Sprintf("parse %s: %v", path, err),
		}
	}
	return &v, nil
}
