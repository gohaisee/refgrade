package check

import (
	"context"
	"regexp"
	"strings"
)

var (
	mongoPasswordURIRe = regexp.MustCompile(`(?i)mongodb(\+srv)?://[^:]+:[^@]+@`)
	brokerCredURIRe    = regexp.MustCompile(`(?i)(amqp|amqps|nats)://[^:]+:[^@]+@`)
	redisPasswordURIRe = regexp.MustCompile(`(?i)redis(s)?://[^/]*:[^@]+@`)
)

func isLocalHostRef(s string) bool {
	lower := strings.ToLower(s)
	return strings.Contains(lower, "localhost") ||
		strings.Contains(lower, "127.0.0.1") ||
		strings.Contains(lower, "::1")
}

func scanSourceStringMatches(mod ModuleView, id, severity string, match func(string) bool) []Finding {
	var findings []Finding
	sev := effectiveSeverity(mod, id, severity)
	for _, gf := range mod.GoSourceFiles() {
		if mod.Excluded(gf.RelPath) {
			continue
		}
		lines, err := readFileLines(gf.Path)
		if err != nil {
			continue
		}
		for i, line := range lines {
			if lineMatchesURIHints(line, match) {
				findings = append(findings, finding(id, sev, gf.RelPath, i+1))
				break
			}
		}
	}
	return findings
}

func lineMatchesURIHints(line string, match func(string) bool) bool {
	if match(line) {
		return true
	}
	for _, part := range strings.Fields(line) {
		part = strings.Trim(part, "`\"',)")
		if match(part) {
			return true
		}
	}
	lower := strings.ToLower(line)
	for _, scheme := range []string{"mongodb://", "mongodb+srv://", "redis://", "rediss://", "amqp://", "amqps://", "nats://"} {
		if idx := strings.Index(lower, scheme); idx >= 0 {
			end := idx + len(scheme)
			for end < len(line) && line[end] != '"' && line[end] != '\'' && line[end] != ' ' && line[end] != '`' {
				end++
			}
			if match(line[idx:end]) {
				return true
			}
		}
	}
	return false
}

func isRemoteMongoWithoutTLS(s string) bool {
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "mongodb://") {
		return false
	}
	if isLocalHostRef(s) {
		return false
	}
	if strings.Contains(lower, "tls=false") || strings.Contains(lower, "ssl=false") {
		return true
	}
	if strings.Contains(lower, "tls=true") || strings.Contains(lower, "ssl=true") {
		return false
	}
	return true
}

func isRemoteRedisWithoutTLS(s string) bool {
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "rediss://") {
		return false
	}
	if !strings.Contains(lower, "redis://") {
		return false
	}
	return !isLocalHostRef(s)
}

func isPlaintextMQToRemote(s string) bool {
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "amqp://"):
		return !isLocalHostRef(s)
	case strings.HasPrefix(lower, "nats://"):
		if isLocalHostRef(s) {
			return false
		}
		return !strings.Contains(lower, "tls")
	default:
		return false
	}
}

// mongo uri password in committed source
type SecM01 struct{ Base }

func NewSecM01() *SecM01 {
	return &SecM01{Base: Base{meta: Meta{
		ID: "sec-m01", Gates: mongoGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecM01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	return scanSourceStringMatches(mod, c.ID(), SeverityFail, func(s string) bool {
		return mongoPasswordURIRe.MatchString(s)
	}), nil
}

// remote mongo without tls in uri
type SecM02 struct{ Base }

func NewSecM02() *SecM02 {
	return &SecM02{Base: Base{meta: Meta{
		ID: "sec-m02", Gates: mongoGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecM02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	return scanSourceStringMatches(mod, c.ID(), SeverityFail, isRemoteMongoWithoutTLS), nil
}

// redis password in committed source
type SecRD01 struct{ Base }

func NewSecRD01() *SecRD01 {
	return &SecRD01{Base: Base{meta: Meta{
		ID: "sec-rd01", Gates: redisGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecRD01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	return scanSourceStringMatches(mod, c.ID(), SeverityFail, func(s string) bool {
		if redisPasswordURIRe.MatchString(s) {
			return true
		}
		lower := strings.ToLower(s)
		return strings.Contains(lower, "password:") &&
			strings.Contains(s, `"`) &&
			!strings.Contains(lower, `password: ""`) &&
			!strings.Contains(lower, `password:""`)
	}), nil
}

// redis without tls to remote host
type SecRD02 struct{ Base }

func NewSecRD02() *SecRD02 {
	return &SecRD02{Base: Base{meta: Meta{
		ID: "sec-rd02", Gates: redisGates(), Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecRD02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	return scanSourceStringMatches(mod, c.ID(), SeverityFail, isRemoteRedisWithoutTLS), nil
}

// broker url credentials in repo
type SecMQ01 struct{ Base }

func NewSecMQ01() *SecMQ01 {
	return &SecMQ01{Base: Base{meta: Meta{
		ID: "sec-mq01", Gates: mqGates, Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecMQ01) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	return scanSourceStringMatches(mod, c.ID(), SeverityFail, func(s string) bool {
		return brokerCredURIRe.MatchString(s)
	}), nil
}

// plaintext amqp/nats to remote broker
type SecMQ02 struct{ Base }

func NewSecMQ02() *SecMQ02 {
	return &SecMQ02{Base: Base{meta: Meta{
		ID: "sec-mq02", Gates: mqGates, Domain: "security", DefaultSeverity: SeverityFail,
	}}}
}

func (c *SecMQ02) Run(ctx context.Context, mod ModuleView) ([]Finding, error) {
	_ = ctx
	if checkDisabled(mod, c.ID()) {
		return nil, nil
	}
	return scanSourceStringMatches(mod, c.ID(), SeverityFail, isPlaintextMQToRemote), nil
}
