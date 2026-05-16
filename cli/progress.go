package cli

import (
	"fmt"
	"io"
)

// Progress writes a single-line progress update to stderr when CLI
// verbosity is >= INFO. MCP server uses a separate path
// (mcp/progress.go) — same shape, different transport.
type Progress struct {
	w       io.Writer
	enabled bool
}

// NewProgress returns a Progress reporter that emits to w when verbosity >= 1.
func NewProgress(w io.Writer, verbosity int) *Progress {
	return &Progress{w: w, enabled: verbosity >= 1}
}

// Page reports a single page fetch. total may be 0 when unknown.
func (p *Progress) Page(page, total, items int) {
	if !p.enabled {
		return
	}
	if total > 0 {
		_, _ = fmt.Fprintf(p.w, "page %d/%d, %d items so far\n", page, total, items)
	} else {
		_, _ = fmt.Fprintf(p.w, "page %d, %d items so far\n", page, items)
	}
}
