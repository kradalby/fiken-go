package ops

import "strings"

// canonicalVatType returns vat_type lowercased so consumers see a
// stable form regardless of upstream casing. Fiken returns "high"
// from products but "HIGH" from sale lines (and is inconsistent
// elsewhere); we normalize at translate time.
func canonicalVatType(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
