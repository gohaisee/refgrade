package refgradeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gohaisee/refgrade/internal/yamlflat"
)

// optional .refgrade.yaml at module root
type Config struct {
	Lang string
}

// reads .refgrade.yaml when present; missing file → empty config
func Load(moduleRoot string) (*Config, error) {
	path := filepath.Join(moduleRoot, ".refgrade.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	flat, err := yamlflat.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &Config{Lang: strings.TrimSpace(flat["lang"])}, nil
}

// report language: flag > REFGRADE_LANG > yaml > en
func ResolveLang(flagLang, moduleRoot string) (string, error) {
	if flagLang != "" {
		return normalizeLang(flagLang)
	}
	if v := os.Getenv("REFGRADE_LANG"); v != "" {
		return normalizeLang(v)
	}
	cfg, err := Load(moduleRoot)
	if err != nil {
		return "", err
	}
	if cfg.Lang != "" {
		return normalizeLang(cfg.Lang)
	}
	return "en", nil
}

func normalizeLang(lang string) (string, error) {
	lang = strings.TrimSpace(lang)
	if lang != "en" && lang != "ru" {
		return "", fmt.Errorf("unsupported language %q", lang)
	}
	return lang, nil
}
