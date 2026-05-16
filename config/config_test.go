package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileAndEnv(t *testing.T) {
	// Neutralize any FIKEN_* env vars inherited from the shell (e.g.
	// .envrc.local). Load() unconditionally consumes them via koanf's
	// env provider, so a token in the live shell would shadow the
	// fixture file's value and break the test.
	t.Setenv("FIKEN_TOKEN", "")
	t.Setenv("FIKEN_LANG", "")
	t.Setenv("FIKEN_PROFILE", "")

	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	body := `default_profile = "work"
[profiles.work]
token = "filetok"
company = "acme"
lang = "nb"
`
	if err := os.WriteFile(fp, []byte(body), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("FIKEN_COMPANY", "envco")

	cfg, err := Load(fp, nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	prof, ok := cfg.Resolve("")
	if !ok {
		t.Fatalf("resolve default profile failed")
	}
	if prof.Token != "filetok" {
		t.Errorf("Token=%q want filetok", prof.Token)
	}
	if prof.Company != "envco" {
		t.Errorf("Company=%q want envco (env override)", prof.Company)
	}
	if prof.Lang != "nb" {
		t.Errorf("Lang=%q want nb", prof.Lang)
	}
}

func TestResolveExplicitProfile(t *testing.T) {
	cfg := &Config{
		DefaultProfile: "p1",
		Profiles: map[string]Profile{
			"p1": {Company: "a"},
			"p2": {Company: "b"},
		},
	}
	prof, ok := cfg.Resolve("p2")
	if !ok || prof.Company != "b" {
		t.Fatalf("Resolve(p2) gave %v", prof)
	}
	defp, _ := cfg.Resolve("")
	if defp.Company != "a" {
		t.Fatalf("default profile resolved to %v want company=a", defp)
	}
}
