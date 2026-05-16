package main

import (
	"strings"
	"testing"
)

// TestCheckPropertyOK passes when the field has the right format.
func TestCheckPropertyOK(t *testing.T) {
	violations := checkProperty("Invoice", "issueDate", property{Type: "string", Format: "date"}, nil)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %v", violations)
	}
}

// TestCheckPropertyMissingFormat fails on *Date with no format.
func TestCheckPropertyMissingFormat(t *testing.T) {
	violations := checkProperty("Invoice", "issueDate", property{Type: "string"}, nil)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %v", violations)
	}
	if !strings.Contains(violations[0], "issueDate") {
		t.Fatalf("violation msg missing field: %q", violations[0])
	}
}

// TestCheckPropertyDateTime accepts either `date` or `date-time` on
// fields ending in *At/*Time/*DateTime.
func TestCheckPropertyDateTime(t *testing.T) {
	cases := []struct {
		field  string
		format string
		ok     bool
	}{
		{"createdAt", "date-time", true},
		{"createdAt", "date", false},
		{"updatedAt", "", false},
		{"dueDate", "date", true},
		{"dueDate", "date-time", true},
		{"name", "", true},
	}
	for _, c := range cases {
		t.Run(c.field+"/"+c.format, func(t *testing.T) {
			v := checkProperty("X", c.field, property{Type: "string", Format: c.format}, nil)
			if c.ok && len(v) != 0 {
				t.Fatalf("expected ok, got %v", v)
			}
			if !c.ok && len(v) == 0 {
				t.Fatalf("expected violation, got none")
			}
		})
	}
}

// TestCheckPropertyIgnored verifies that fields whose names appear in
// the ignored set bypass all unit-format checks. This is how the
// pre-commit hook tolerates the wall-clock startTime/endTime fields
// in timeEntryRequest/timeEntryResult/updateTimeEntryRequest.
func TestCheckPropertyIgnored(t *testing.T) {
	ignored := map[string]bool{"startTime": true, "endTime": true}
	// Without ignore: would violate (Time suffix, no date-time format).
	v := checkProperty("timeEntryRequest", "startTime", property{Type: "string"}, nil)
	if len(v) == 0 {
		t.Fatalf("baseline: expected violation for startTime, got none")
	}
	// With ignore: must be silent.
	v = checkProperty("timeEntryRequest", "startTime", property{Type: "string"}, ignored)
	if len(v) != 0 {
		t.Fatalf("ignored: expected no violations, got %v", v)
	}
	v = checkProperty("timeEntryRequest", "endTime", property{Type: "string"}, ignored)
	if len(v) != 0 {
		t.Fatalf("ignored: expected no violations, got %v", v)
	}
}
