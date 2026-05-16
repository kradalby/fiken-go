package output

import (
	"encoding/json"
	"io"
)

type jsonRenderer struct{ w io.Writer }

func (r *jsonRenderer) Render(v any) error {
	enc := json.NewEncoder(r.w)
	return enc.Encode(v)
}
