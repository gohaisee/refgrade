package yamlflat

import (
	"fmt"
	"strings"
)

// flat key: value lines (locales and .refgrade.yaml)
func Parse(data []byte) (map[string]string, error) {
	out := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		colon := strings.Index(trim, ":")
		if colon < 0 {
			return nil, fmt.Errorf("line %d: expected key: value", i+1)
		}
		key := strings.TrimSpace(trim[:colon])
		val := strings.TrimSpace(trim[colon+1:])
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", i+1)
		}
		out[key] = unquote(val)
	}
	return out, nil
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
