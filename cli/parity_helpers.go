package cli

import (
	"bytes"
	"encoding/json"
)

// jsonEqual compares two JSON byte slices semantically (whitespace
// and key-order tolerant). Returns false on parse error.
func jsonEqual(a, b []byte) bool {
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return false
	}
	aCanon, _ := json.Marshal(av)
	bCanon, _ := json.Marshal(bv)
	return bytes.Equal(aCanon, bCanon)
}
