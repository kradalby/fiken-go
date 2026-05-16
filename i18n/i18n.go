package i18n

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/en.toml locales/nb.toml
var localesFS embed.FS

// Bundle wraps go-i18n's Bundle plus a flat key map for parity tests.
type Bundle struct {
	inner *goi18n.Bundle
	flat  map[string]map[string]string
}

// MustLoad reads en.toml + nb.toml from the embedded FS and panics
// on parse error.
func MustLoad() *Bundle {
	b, err := Load()
	if err != nil {
		panic(err)
	}
	return b
}

// Load reads en.toml + nb.toml.
func Load() (*Bundle, error) {
	bundle := goi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	flat := map[string]map[string]string{}
	for _, lang := range []string{"en", "nb"} {
		data, err := localesFS.ReadFile("locales/" + lang + ".toml")
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", lang, err)
		}
		if _, err := bundle.ParseMessageFileBytes(data, lang+".toml"); err != nil {
			return nil, fmt.Errorf("parse %s.toml: %w", lang, err)
		}
		m, err := flatten(data)
		if err != nil {
			return nil, fmt.Errorf("flatten %s: %w", lang, err)
		}
		flat[lang] = m
	}
	return &Bundle{inner: bundle, flat: flat}, nil
}

// T returns the localized string for (lang, key). data may be nil.
// Unknown lang falls back to en; unknown key returns "".
func (b *Bundle) T(lang, key string, data map[string]any) string {
	lang = normalizeLang(lang)
	if v, ok := b.flat[lang][key]; ok {
		return interpolate(v, data)
	}
	if lang != "en" {
		if v, ok := b.flat["en"][key]; ok {
			return interpolate(v, data)
		}
	}
	return ""
}

// Keys returns every key defined in the given locale, sorted.
func (b *Bundle) Keys(lang string) []string {
	m := b.flat[normalizeLang(lang)]
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func normalizeLang(lang string) string {
	switch strings.ToLower(strings.Split(lang, "_")[0]) {
	case "nb", "no":
		return "nb"
	case "":
		return "en"
	default:
		return "en"
	}
}

func interpolate(s string, data map[string]any) string {
	if data == nil {
		return s
	}
	for k, v := range data {
		s = strings.ReplaceAll(s, "{{"+k+"}}", fmt.Sprint(v))
	}
	return s
}

func flatten(data []byte) (map[string]string, error) {
	var raw map[string]any
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := map[string]string{}
	walk(raw, "", out)
	return out, nil
}

func walk(v any, prefix string, out map[string]string) {
	switch x := v.(type) {
	case string:
		out[prefix] = x
	case map[string]any:
		for k, v := range x {
			p := k
			if prefix != "" {
				p = prefix + "." + k
			}
			walk(v, p, out)
		}
	}
}
