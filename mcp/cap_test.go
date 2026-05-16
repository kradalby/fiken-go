package mcp

import (
	"testing"

	"github.com/kradalby/fiken-go/ops"
)

func TestAllowOpForCap(t *testing.T) {
	// Pick two registry entries from each polarity rather than hardcoded
	// strings to stay robust against Op constant renames.
	var read, write string
	for name, entry := range ops.Registry {
		if entry.Mutating && write == "" {
			write = name
		}
		if !entry.Mutating && read == "" {
			read = name
		}
		if read != "" && write != "" {
			break
		}
	}
	if read == "" || write == "" {
		t.Fatalf("ops.Registry missing read or write entries (read=%q write=%q)", read, write)
	}

	tests := []struct {
		name string
		op   string
		cap  Capability
		want bool
	}{
		{"read op, no cap", read, Capability{}, true},
		{"read op, write cap", read, Capability{Write: true}, true},
		{"write op, no cap", write, Capability{}, false},
		{"write op, write cap", write, Capability{Write: true}, true},
		{"unknown op, no cap", "definitely_not_an_op_xyz", Capability{}, false},
		{"unknown op, write cap", "definitely_not_an_op_xyz", Capability{Write: true}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := AllowOpForCap(tc.op, tc.cap); got != tc.want {
				t.Fatalf("AllowOpForCap(%q, %+v) = %v, want %v", tc.op, tc.cap, got, tc.want)
			}
		})
	}
}
