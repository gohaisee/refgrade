package i18n

import (
	"fmt"

	"github.com/gohaisee/refgrade/internal/yamlflat"
	"github.com/gohaisee/refgrade/locales"
)

// localized strings keyed by dot path
type Bundle struct {
	lang     string
	flat     map[string]string
	fallback map[string]string
}

// reads lang yaml from embedded locales; en is always fallback
func Load(lang string) (*Bundle, error) {
	if lang != "en" && lang != "ru" {
		return nil, fmt.Errorf("unsupported language %q", lang)
	}

	enFlat, err := readLocale("en")
	if err != nil {
		return nil, err
	}

	b := &Bundle{
		lang:     lang,
		flat:     enFlat,
		fallback: enFlat,
	}
	if lang == "en" {
		return b, nil
	}

	ruFlat, err := readLocale("ru")
	if err != nil {
		return nil, err
	}
	b.flat = ruFlat
	return b, nil
}

func readLocale(lang string) (map[string]string, error) {
	data, err := locales.FS.ReadFile(lang + ".yaml")
	if err != nil {
		return nil, fmt.Errorf("read locale %s: %w", lang, err)
	}
	flat, err := yamlflat.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse locale %s: %w", lang, err)
	}
	return flat, nil
}

// localized string for key; falls back to en then key itself
func (b *Bundle) T(key string) string {
	if s, ok := b.flat[key]; ok && s != "" {
		return s
	}
	if s, ok := b.fallback[key]; ok && s != "" {
		return s
	}
	return key
}

// active language code
func (b *Bundle) Lang() string {
	return b.lang
}

// whether key exists in active locale or fallback
func (b *Bundle) Has(key string) bool {
	_, ok := b.flat[key]
	if ok {
		return true
	}
	_, ok = b.fallback[key]
	return ok
}
