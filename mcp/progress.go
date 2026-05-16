package mcp

// Progress writes MCP progress notifications during paged operations.
// Plan B's companies_list typically fits in one page; helper is
// unused. Plan C exercises it once a paged op surfaces multi-page
// responses.
type Progress struct {
	enabled bool
}

// NewProgress creates a Progress; verbosity gates whether updates emit.
func NewProgress(verbosity int) *Progress {
	return &Progress{enabled: verbosity >= 1}
}

// Enabled reports whether the Progress would emit notifications.
// Kept for symmetry with the Plan C wiring; callers will gate the
// per-page Notify call (added later) on this.
func (p *Progress) Enabled() bool {
	if p == nil {
		return false
	}
	return p.enabled
}
