package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kradalby/fiken-go/ops"
)

type sampleOut struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func (s sampleOut) TableHeader() []string { return []string{"SLUG", "NAME"} }
func (s sampleOut) TableRow() []string    { return []string{s.Slug, s.Name} }

func TestJSONRenderResult(t *testing.T) {
	var buf bytes.Buffer
	r := JSON(&buf)
	res := ops.Result[ops.ListOut[sampleOut]]{
		Ok: &ops.ListOut[sampleOut]{
			Items: []sampleOut{{Slug: "acme", Name: "Acme AS"}},
			Meta:  ops.ListMeta{Returned: 1},
		},
	}
	if err := r.Render(res); err != nil {
		t.Fatalf("render: %v", err)
	}
	got := buf.String()
	want := `{"ok":{"items":[{"slug":"acme","name":"Acme AS"}],"meta":{"truncated":false,"returned":1}},"error":null}` + "\n"
	if got != want {
		t.Fatalf("mismatch:\n got %q\nwant %q", got, want)
	}
}

func TestTableRenderResult(t *testing.T) {
	var buf bytes.Buffer
	r := Table(&buf, nil)
	res := ops.Result[ops.ListOut[sampleOut]]{
		Ok: &ops.ListOut[sampleOut]{
			Items: []sampleOut{{Slug: "acme", Name: "Acme AS"}},
			Meta:  ops.ListMeta{Returned: 1},
		},
	}
	if err := r.Render(res); err != nil {
		t.Fatalf("render: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "SLUG") || !strings.Contains(got, "Acme AS") {
		t.Fatalf("table missing header or row:\n%s", got)
	}
}

func TestTableRenderError(t *testing.T) {
	var buf bytes.Buffer
	translator := func(code, msg string) string {
		if code == "not_found" {
			return "Resource not found."
		}
		return msg
	}
	r := Table(&buf, translator)
	res := ops.Result[ops.ListOut[sampleOut]]{
		Error: &ops.Error{Code: "not_found", Message: "raw upstream"},
	}
	if err := r.Render(res); err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(buf.String(), "Resource not found.") {
		t.Fatalf("translator not consulted:\n%s", buf.String())
	}
}
