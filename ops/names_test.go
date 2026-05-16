package ops

import "testing"

func TestRegistryHasNonEmpty(t *testing.T) {
	if len(Registry) == 0 {
		t.Fatal("Registry must have at least one entry")
	}
}

func TestRegistryMutatingMatchesGen(t *testing.T) {
	for name, entry := range Registry {
		gotMut := IsMutating(entry.OperationID)
		if entry.Mutating != gotMut {
			t.Errorf("Registry[%q].Mutating=%v but IsMutating(%q)=%v",
				name, entry.Mutating, entry.OperationID, gotMut)
		}
	}
}

func TestCompaniesListNonMutating(t *testing.T) {
	// companies_list is a GET — must be non-mutating so MCP
	// read-only mode exposes it.
	if Registry[OpCompaniesList].Mutating {
		t.Fatal("companies_list should NOT be mutating (it's GET /companies)")
	}
	if Registry[OpCompaniesGet].Mutating {
		t.Fatal("companies_get should NOT be mutating (it's GET /companies/{slug})")
	}
}

func TestOpConstNamesAreInRegistry(t *testing.T) {
	for _, name := range []string{OpCompaniesList, OpCompaniesGet} {
		if _, ok := Registry[name]; !ok {
			t.Errorf("op %q in const but not in Registry", name)
		}
	}
}
