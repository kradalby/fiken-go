// Package mcp builds an MCP server over the ops.Client. Each ops.Op*
// becomes a tool with InputSchema derived from the In* struct.
// Read-only mode filters tools by consulting ops.Registry.
package mcp

import "github.com/kradalby/fiken-go/ops"

// Mode is the runtime policy switch.
type Mode int

const (
	// ModeReadOnly hides every Op flagged as Mutating in ops.Registry.
	ModeReadOnly Mode = iota
	// ModeReadWrite exposes every Op.
	ModeReadWrite
)

// AllowOp returns true if mode permits exposing op.
// The op name passed here is the Op* const value (e.g. "companies_list"),
// which is mapped to its OAS operationId via ops.Registry.
func AllowOp(mode Mode, opName string) bool {
	if mode == ModeReadWrite {
		return true
	}
	entry, ok := ops.Registry[opName]
	if !ok {
		return false // unknown — fail closed
	}
	return !entry.Mutating
}

// AllowOpForCap is the runtime sibling of AllowOp used by the tsnet
// transport. Reads are always permitted (tailnet membership implies
// read access); writes require a Capability with Write=true.
// Unknown op names fail closed.
func AllowOpForCap(opName string, c Capability) bool {
	entry, ok := ops.Registry[opName]
	if !ok {
		return false
	}
	if entry.Mutating {
		return c.Write
	}
	return true
}
