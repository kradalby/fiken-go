package ops

import (
	"encoding/json"
	"fmt"
	"time"
)

// Date is a civil (no-timezone) calendar date in YYYY-MM-DD form.
// Used for Out struct fields named *Date (no time component).
// Marshals to/from a quoted ISO 8601 date string; rejects anything
// not in the exact `2006-01-02` layout.
type Date string

const dateLayout = "2006-01-02"

// MarshalJSON emits a quoted YYYY-MM-DD string.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}

// parseDate parses a YYYY-MM-DD string into a UTC time.Time. Returns
// an error for any other format so callers don't silently produce
// midnight-in-local-tz drift.
func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}

// UnmarshalJSON parses a quoted YYYY-MM-DD string. Rejects other
// formats so misparsed upstream fields become loud errors instead
// of silent UTC drift.
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		*d = ""
		return nil
	}
	if _, err := time.Parse(dateLayout, s); err != nil {
		return fmt.Errorf("ops.Date: %w (must be YYYY-MM-DD)", err)
	}
	*d = Date(s)
	return nil
}
