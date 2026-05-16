// Package output renders ops.Result[T] envelopes for the CLI.
// Two factories: JSON for `--json` mode (byte-equal to MCP tool
// result), Table for the default human-readable mode.
package output

import "io"

// Renderer writes one Result[T] envelope to its underlying writer.
type Renderer interface {
	Render(any) error
}

// ErrorTranslator maps (code, raw-message) -> localized human message.
// Provided by the CLI from i18n; nil to skip translation (raw used).
type ErrorTranslator func(code, message string) string

// JSON returns a Renderer that emits the envelope via
// encoding/json.NewEncoder.
func JSON(w io.Writer) Renderer { return &jsonRenderer{w: w} }

// Table returns a Renderer that prints success rows via tabwriter
// and localizes error envelopes via translator.
func Table(w io.Writer, translator ErrorTranslator) Renderer {
	return &tableRenderer{w: w, t: translator}
}
