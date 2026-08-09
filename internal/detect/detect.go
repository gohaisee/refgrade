package detect

import (
	"context"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gohaisee/refgrade/internal/project"
)

// one detected library
type Stack struct {
	Name       string
	ImportPath string
}

var importToStack = map[string]string{
	"github.com/gin-gonic/gin":                  "gin",
	"github.com/labstack/echo/v4":               "echo",
	"github.com/go-chi/chi/v5":                  "chi",
	"github.com/99designs/gqlgen":               "gqlgen",
	"github.com/graph-gophers/graphql-go":       "graphql-go",
	"google.golang.org/grpc":                    "grpc",
	"github.com/bufbuild/connect-go":            "connect",
	"github.com/grpc-ecosystem/grpc-gateway/v2": "grpc-gateway",
	"github.com/jackc/pgx/v5":                   "pgx",
	"gorm.io/gorm":                              "gorm",
	"github.com/jmoiron/sqlx":                   "sqlx",
	"github.com/sqlc-dev/sqlc":                  "sqlc",
	"entgo.io/ent":                              "ent",
	"go.mongodb.org/mongo-driver/mongo":         "mongo-driver",
	"github.com/redis/go-redis/v9":              "go-redis",
	"github.com/segmentio/kafka-go":             "kafka-go",
	"github.com/rabbitmq/amqp091-go":            "rabbitmq",
	"github.com/nats-io/nats.go":                "nats",
}

// known stack libraries from module imports
func Detect(_ context.Context, mod *project.Module) ([]Stack, error) {
	seen := make(map[string]Stack)
	fset := token.NewFileSet()

	for _, pkg := range mod.Packages {
		imports, err := packageImports(fset, pkg.Dir, pkg.GoFiles)
		if err != nil {
			return nil, err
		}
		for _, imp := range imports {
			if name, ok := importToStack[imp]; ok {
				seen[name] = Stack{Name: name, ImportPath: imp}
			}
		}
	}

	out := make([]Stack, 0, len(seen))
	for _, s := range seen {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func packageImports(fset *token.FileSet, dir string, goFiles []string) ([]string, error) {
	seen := make(map[string]struct{})
	for _, name := range goFiles {
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}
		for _, spec := range file.Imports {
			imp := strings.Trim(spec.Path.Value, `"`)
			seen[imp] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for imp := range seen {
		out = append(out, imp)
	}
	sort.Strings(out)
	return out, nil
}

// stack list for detect command
func FormatText(stacks []Stack, title string, none string) string {
	if len(stacks) == 0 {
		return title + "\n" + none + "\n"
	}
	var b strings.Builder
	b.WriteString(title)
	b.WriteByte('\n')
	for _, s := range stacks {
		b.WriteString("- ")
		b.WriteString(s.Name)
		b.WriteString(" (")
		b.WriteString(s.ImportPath)
		b.WriteString(")\n")
	}
	return b.String()
}

// import path to stack name (tests)
func MatchImport(importPath string) (string, bool) {
	name, ok := importToStack[importPath]
	return name, ok
}
