package ops

import (
	"encoding/json"
	"testing"
)

func TestDateMarshal(t *testing.T) {
	d := Date("2026-05-15")
	got, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(got) != `"2026-05-15"` {
		t.Fatalf("got %s want \"2026-05-15\"", got)
	}
}

func TestDateUnmarshal(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte(`"2026-05-15"`), &d); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if d != "2026-05-15" {
		t.Fatalf("got %q want 2026-05-15", d)
	}
}

func TestDateUnmarshalRejectsBad(t *testing.T) {
	var d Date
	err := json.Unmarshal([]byte(`"2026/05/15"`), &d)
	if err == nil {
		t.Fatal("expected error on bad date format")
	}
}
