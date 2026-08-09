package refgradeconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gohaisee/refgrade/internal/yamlflat"
)

const (
	ProfileMinimal  = "minimal"
	ProfileStandard = "standard"
	ProfileStrict   = "strict"

	SeverityOff  = "off"
	SeverityInfo = "info"
	SeverityWarn = "warn"
	SeverityFail = "fail"
)

// optional .refgrade.yaml at module root
type Config struct {
	Lang    string
	Profile string
	Kind    string
	Exclude []string
	Layers  []LayerRule
	Checks  map[string]CheckSetting
}

// custom import forbid rule for a package path prefix
type LayerRule struct {
	Path   string
	Forbid []string
}

// per-check override from yaml
type CheckSetting struct {
	Enabled  bool
	Severity string
}

// reads .refgrade.yaml when present; missing file → empty config
func Load(moduleRoot string) (*Config, error) {
	path := filepath.Join(moduleRoot, ".refgrade.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Checks: make(map[string]CheckSetting)}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	flat, err := yamlflat.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return parseFlat(flat), nil
}

func parseFlat(flat map[string]string) *Config {
	cfg := &Config{
		Lang:    strings.TrimSpace(flat["lang"]),
		Profile: strings.TrimSpace(flat["profile"]),
		Kind:    strings.TrimSpace(flat["kind"]),
		Checks:  make(map[string]CheckSetting),
	}
	cfg.Exclude = indexedStrings(flat, "exclude")
	cfg.Layers = parseLayers(flat)
	parseChecks(flat, cfg)
	if cfg.Profile == "" {
		cfg.Profile = ProfileStandard
	}
	return cfg
}

func indexedStrings(flat map[string]string, prefix string) []string {
	var keys []string
	prefixDot := prefix + "."
	for k := range flat {
		if strings.HasPrefix(k, prefixDot) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var out []string
	for _, k := range keys {
		v := strings.TrimSpace(flat[k])
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseLayers(flat map[string]string) []LayerRule {
	indexes := map[int]struct{}{}
	for k := range flat {
		if !strings.HasPrefix(k, "layers.") {
			continue
		}
		parts := strings.Split(k, ".")
		if len(parts) < 3 {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(parts[1], "%d", &idx); err != nil {
			continue
		}
		indexes[idx] = struct{}{}
	}
	var idxs []int
	for i := range indexes {
		idxs = append(idxs, i)
	}
	sort.Ints(idxs)

	var rules []LayerRule
	for _, i := range idxs {
		pathKey := fmt.Sprintf("layers.%d.path", i)
		path := strings.TrimSpace(flat[pathKey])
		if path == "" {
			continue
		}
		forbid := indexedLayerForbid(flat, i)
		rules = append(rules, LayerRule{Path: path, Forbid: forbid})
	}
	return rules
}

func indexedLayerForbid(flat map[string]string, layerIdx int) []string {
	prefix := fmt.Sprintf("layers.%d.forbid.", layerIdx)
	var keys []string
	for k := range flat {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var out []string
	for _, k := range keys {
		v := strings.TrimSpace(flat[k])
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func parseChecks(flat map[string]string, cfg *Config) {
	for k, v := range flat {
		if !strings.HasPrefix(k, "checks.") {
			continue
		}
		id := strings.TrimPrefix(k, "checks.")
		if strings.Contains(id, ".") {
			// checks.cfg-01.severity style
			parts := strings.SplitN(id, ".", 2)
			if len(parts) != 2 || parts[1] != "severity" {
				continue
			}
			id = parts[0]
			set := cfg.Checks[id]
			set.Severity = strings.TrimSpace(v)
			set.Enabled = set.Severity != SeverityOff
			cfg.Checks[id] = set
			continue
		}
		set := cfg.Checks[id]
		v = strings.TrimSpace(v)
		switch v {
		case SeverityOff, "false", "disable", "disabled":
			set.Enabled = false
		case SeverityInfo, SeverityWarn, SeverityFail, "on", "true", "enable", "enabled":
			set.Enabled = true
			if v != "on" && v != "true" && v != "enable" && v != "enabled" {
				set.Severity = v
			}
		default:
			set.Enabled = true
			set.Severity = v
		}
		cfg.Checks[id] = set
	}
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

// whether check id is enabled (default on)
func (c *Config) CheckEnabled(id string) bool {
	if c == nil || c.Checks == nil {
		return true
	}
	set, ok := c.Checks[id]
	if !ok {
		return true
	}
	return set.Enabled
}

// effective severity: yaml override > profile > default
func (c *Config) EffectiveSeverity(id, defaultSev string) string {
	if c != nil && c.Checks != nil {
		if set, ok := c.Checks[id]; ok && set.Severity != "" {
			return set.Severity
		}
	}
	return ApplyProfile(c.Profile, id, defaultSev)
}

// profile adjusts default severities
func ApplyProfile(profile, id, defaultSev string) string {
	switch profile {
	case ProfileMinimal:
		if defaultSev == SeverityWarn {
			return SeverityInfo
		}
	case ProfileStrict:
		if defaultSev == SeverityWarn {
			return SeverityFail
		}
		if defaultSev == SeverityInfo {
			return SeverityWarn
		}
	}
	return defaultSev
}

// glob exclude matcher
type ExcludeMatcher struct {
	patterns []string
}

func NewExcludeMatcher(patterns []string) *ExcludeMatcher {
	return &ExcludeMatcher{patterns: append([]string(nil), patterns...)}
}

func (m *ExcludeMatcher) Excluded(relPath string) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}
	rel := filepath.ToSlash(relPath)
	for _, pat := range m.patterns {
		if ok, _ := filepath.Match(pat, rel); ok {
			return true
		}
		if ok, _ := filepath.Match(pat, filepath.Base(rel)); ok {
			return true
		}
	}
	return false
}

// init template content
func InitTemplate() string {
	return `# refgrade config — see https://github.com/gohaisee/refgrade/docs/en/config.md
lang: en
profile: standard
kind: auto

# exclude:
#   - "**/generated/**"
#   - "**/*_gen.go"

# layers:
#   - path: internal/service
#     forbid:
#       - github.com/jackc/pgx/v5

# checks:
#   cfg-01: fail
#   err-01: warn
`
}
