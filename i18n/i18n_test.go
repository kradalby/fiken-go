package i18n

import "testing"

func TestEnglishFallback(t *testing.T) {
	b := MustLoad()
	got := b.T("en", "ops.companies_list.summary", nil)
	if got == "" {
		t.Fatal("missing ops.companies_list.summary in en")
	}
}

func TestBokmalAvailable(t *testing.T) {
	b := MustLoad()
	got := b.T("nb", "ops.companies_list.summary", nil)
	if got == "" {
		t.Fatal("missing ops.companies_list.summary in nb")
	}
}

func TestLangAlias(t *testing.T) {
	b := MustLoad()
	en := b.T("en", "ops.companies_list.summary", nil)
	no := b.T("no", "ops.companies_list.summary", nil)
	nb := b.T("nb", "ops.companies_list.summary", nil)
	if no != nb {
		t.Fatalf("no should alias nb: no=%q nb=%q", no, nb)
	}
	if en == nb {
		t.Fatalf("en and nb produced identical text — likely fallback bug")
	}
}

func TestEveryEnKeyHasNbCounterpart(t *testing.T) {
	b := MustLoad()
	for _, key := range b.Keys("en") {
		if v := b.T("nb", key, nil); v == "" {
			t.Errorf("missing nb translation for %q", key)
		}
	}
}
