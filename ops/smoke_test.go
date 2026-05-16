package ops

import (
	"testing"

	"github.com/kradalby/fiken-go/fiken"
)

// TestGeneratedClientImports asserts that the ogen-generated package
// can be imported and that at least one expected symbol exists. This
// is a smoke test, not a behavioural test — Plan B adds real ones.
func TestGeneratedClientImports(_ *testing.T) {
	// fiken.NewClient is the canonical ogen constructor; if ogen ever
	// renames it, this will fail and tell us at build time.
	_ = fiken.NewClient
}

// TestIsMutatingKnownOps spot-checks ops/mutating.gen.go for two
// op-ids: one mutating, one not. Belt for the codegen pipeline.
func TestIsMutatingKnownOps(t *testing.T) {
	if IsMutating("get_companies") {
		t.Errorf("get_companies should be non-mutating")
	}
	if !IsMutating("create_invoice") {
		t.Errorf("create_invoice should be mutating")
	}
}
