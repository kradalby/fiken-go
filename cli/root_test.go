package cli

import (
	"bytes"
	"testing"
)

func TestRootBuilds(t *testing.T) {
	cmd, err := Root(new(bytes.Buffer), new(bytes.Buffer))
	if err != nil {
		t.Fatalf("Root: %v", err)
	}
	if cmd.Name != "fiken" {
		t.Errorf("name=%q want fiken", cmd.Name)
	}
}
