package ops

import (
	"encoding/json"
	"testing"
)

func TestListOutMarshal(t *testing.T) {
	items := []string{"a", "b"}
	out := ListOut[string]{Items: items, Meta: ListMeta{Returned: 2}}
	got, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"items":["a","b"],"meta":{"truncated":false,"returned":2}}`
	if string(got) != want {
		t.Fatalf("mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestListMetaTruncated(t *testing.T) {
	m := ListMeta{Truncated: true, NextPage: 4, Returned: 100, Cancelled: false}
	got, _ := json.Marshal(m)
	want := `{"truncated":true,"next_page":4,"returned":100}`
	if string(got) != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
