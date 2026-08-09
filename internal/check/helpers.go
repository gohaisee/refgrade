package check

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gohaisee/refgrade/internal/astutil"
)

var (
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)password\s*=\s*['"]?[^'"\s@]+`),
		regexp.MustCompile(`sk-[a-zA-Z0-9]{10,}`),
		regexp.MustCompile(`BEGIN (RSA |EC )?PRIVATE KEY`),
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	}

	dbDriverImports = map[string]struct{}{
		"github.com/jackc/pgx/v5":                   {},
		"github.com/jackc/pgx/v5/pgxpool":           {},
		"gorm.io/gorm":                              {},
		"github.com/jmoiron/sqlx":                   {},
		"database/sql":                              {},
		"github.com/lib/pq":                           {},
		"go.mongodb.org/mongo-driver/mongo":         {},
		"github.com/go-sql-driver/mysql":            {},
	}

	networkSideEffectCalls = map[string]map[string]struct{}{
		"database/sql":                      {"Open": {}},
		"github.com/jackc/pgx/v5":           {"Connect": {}, "ParseConfig": {}},
		"github.com/jackc/pgx/v5/pgxpool":   {"New": {}, "NewWithConfig": {}},
		"go.mongodb.org/mongo-driver/mongo": {"Connect": {}, "NewClient": {}},
	}

	placeholderModulePaths = []*regexp.Regexp{
		regexp.MustCompile(`^example\.com/`),
		regexp.MustCompile(`^test$`),
		regexp.MustCompile(`^foo$`),
		regexp.MustCompile(`^tmp/`),
	}
)

func finding(id, severity, file string, line int) Finding {
	return Finding{ID: id, Severity: severity, File: file, Line: line}
}

func effectiveSeverity(mod ModuleView, id, defaultSev string) string {
	cfg := mod.Config()
	if cfg == nil {
		return defaultSev
	}
	return cfg.EffectiveSeverity(id, defaultSev)
}

func poolFor(mod ModuleView, filter astutil.Filter) (*astutil.Pool, error) {
	return mod.ASTPool(filter)
}

func fileImportsPath(f *astutil.File, forbidden ...string) bool {
	for _, imp := range f.Imports {
		for _, fb := range forbidden {
			if imp == fb || strings.HasPrefix(imp, fb+"/") {
				return true
			}
		}
	}
	return false
}

func pkgRelPath(mod ModuleView, importPath string) string {
	for _, p := range mod.Packages() {
		if p.ImportPath == importPath {
			return p.RelDir
		}
	}
	return ""
}

func pathUnder(rel, prefix string) bool {
	rel = filepath.ToSlash(rel)
	prefix = strings.TrimSuffix(filepath.ToSlash(prefix), "/")
	return rel == prefix || strings.HasPrefix(rel, prefix+"/")
}

func isHandlerPackage(rel string) bool {
	rel = filepath.ToSlash(rel)
	return strings.Contains(rel, "/handler") ||
		strings.Contains(rel, "/handlers") ||
		strings.Contains(rel, "/resolver") ||
		strings.Contains(rel, "/resolvers") ||
		strings.Contains(rel, "/grpc") ||
		strings.Contains(rel, "/transport")
}

func isServicePackage(rel string) bool {
	rel = filepath.ToSlash(rel)
	return strings.Contains(rel, "/service") || strings.Contains(rel, "/services")
}

func hasConfigPackage(mod ModuleView) bool {
	for _, p := range mod.Packages() {
		if strings.HasSuffix(p.ImportPath, "/internal/config") ||
			strings.HasSuffix(p.ImportPath, "/config") && strings.Contains(p.RelDir, "internal/") {
			return true
		}
	}
	return false
}

func layerForbidFindings(mod ModuleView, id string) []Finding {
	cfg := mod.Config()
	if cfg == nil || len(cfg.Layers) == 0 {
		return nil
	}
	var findings []Finding
	for _, rule := range cfg.Layers {
		for _, pkg := range mod.Packages() {
			if !pathUnder(pkg.RelDir, rule.Path) {
				continue
			}
			for _, gf := range mod.GoSourceFiles() {
				if !pathUnder(gf.RelPath, pkg.RelDir) {
					continue
				}
				for _, forbid := range rule.Forbid {
					pool, err := mod.ASTPool(astutil.DefaultFilter())
					if err != nil {
						continue
					}
					for _, f := range pool.Files {
						if f.RelPath != gf.RelPath {
							continue
						}
						if fileImportsPath(f, forbid) {
							findings = append(findings, finding(id, effectiveSeverity(mod, id, SeverityFail), gf.RelPath, 1))
						}
					}
				}
			}
		}
	}
	return findings
}

func applyLayerSeverity(mod ModuleView, id string, f Finding) Finding {
	f.Severity = effectiveSeverity(mod, id, f.Severity)
	return f
}

func checkDisabled(mod ModuleView, id string) bool {
	cfg := mod.Config()
	if cfg == nil {
		return false
	}
	return !cfg.CheckEnabled(id)
}

