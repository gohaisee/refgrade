package check

import (
	"context"
	"go/ast"
	"testing"

	"github.com/gohaisee/refgrade/internal/astutil"
)

type stubModuleWithGoMod struct {
	stubModule
	goMod []byte
}

func (s stubModuleWithGoMod) GoModContent() []byte { return s.goMod }

func TestGql01_flagsResolverDbImport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "internal/resolver/schema.resolvers.go"
	path := writeGoFile(t, dir, rel, `package resolver

import "database/sql"

type Resolver struct { db *sql.DB }
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGql01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "gql-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGql02_flagsQueryInLoop(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "internal/resolver/user.resolvers.go"
	path := writeGoFile(t, dir, rel, `package resolver

type Resolver struct { db DB }
type DB struct{}
func (d DB) Query(string, ...interface{}) (interface{}, error) { return nil, nil }

func (r *Resolver) Load(ids []int) error {
	for _, id := range ids {
		_, _ = r.db.Query("SELECT name FROM users WHERE id = $1", id)
	}
	return nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGql02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "gql-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGql03_flagsHandEditedGenerated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "internal/graph/generated.go"
	path := writeGoFile(t, dir, rel, `package graph

type Query struct{}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGql03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGql05_flagsIntrospection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "internal/server/graphql.go"
	path := writeGoFile(t, dir, rel, `package server

import "github.com/99designs/gqlgen/graphql/handler/extension"

func setup() { _ = extension.Introspection{} }
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGql05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGql06_flagsPlayground(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "github.com/99designs/gqlgen/graphql/playground"

func main() { _ = playground.Handler("GQL", "/query") }
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGql06().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGqlgen01_missingConfig(t *testing.T) {
	t.Parallel()
	mod := stubModule{root: t.TempDir()}
	findings, err := NewGqlgen01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGgl02_legacyGraphqlOnly(t *testing.T) {
	t.Parallel()
	mod := stubModuleWithGoMod{
		stubModule: stubModule{root: t.TempDir()},
		goMod:      []byte("module example.com/app\n\ngo 1.22\n\nrequire github.com/graphql-go/graphql v0.8.1\n"),
	}
	findings, err := NewGgl02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestIsMapStringInterface(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := "internal/resolver/x.go"
	path := writeGoFile(t, dir, rel, `package resolver

var m map[string]interface{}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	pool, err := mod.ASTPool(astutil.DefaultFilter())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	pool.Inspect(func(_ *astutil.File, n ast.Node) bool {
		if isMapStringInterface(n) {
			found = true
		}
		return true
	})
	if !found {
		t.Fatal("expected map type")
	}
}
