package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stubModuleWithPkgs struct {
	stubModule
	pkgs []PackageInfo
	mod  []byte
}

func (s stubModuleWithPkgs) Packages() []PackageInfo { return s.pkgs }

func (s stubModuleWithPkgs) GoModContent() []byte { return s.mod }

func restStacks() []string {
	return []string{"gin", "echo", "chi", "net/http"}
}

func TestRest01_flagsFatHandler(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/handler/handler.go"
	body := "package handler\n\nfunc Handle() {\n"
	for i := 0; i < 85; i++ {
		body += "\t_ = 1\n"
	}
	body += "}\n"
	path := writeGoFile(t, dir, rel, body)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}},
		pkgs:       []PackageInfo{{RelDir: "internal/handler"}},
	}
	findings, err := NewRest01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest01_allowsServiceCall(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/handler/handler.go"
	body := "package handler\n\nimport \"example.com/app/internal/service\"\n\n"
	for i := 0; i < 85; i++ {
		body += "// line\n"
	}
	body += `func Handle(s *service.Svc) {
	s.Do()
}
`
	path := writeGoFile(t, dir, rel, body)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}},
		pkgs:       []PackageInfo{{RelDir: "internal/handler"}},
	}
	findings, err := NewRest01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest02_flagsSQLImport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/handler/handler.go"
	path := writeGoFile(t, dir, rel, `package handler

import "database/sql"

func Handle(db *sql.DB) {}
`)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}},
		pkgs:       []PackageInfo{{RelDir: "internal/handler"}},
	}
	findings, err := NewRest02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest03_flagsMissingTimeouts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "net/http"

func main() {
	srv := &http.Server{Addr: ":8080"}
	_ = srv
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-03" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest03_allowsTimeouts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import (
	"net/http"
	"time"
)

func main() {
	srv := &http.Server{Addr: ":8080", ReadHeaderTimeout: time.Second}
	_ = srv
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest04_flagsPostWithoutLimit(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.New()
	r.POST("/upload", func(c *gin.Context) {})
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest04().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatalf("expected rest-04 findings, got %+v", findings)
	}
}

func TestRest05_flagsCORSWildcard(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/server/cors.go"
	path := writeGoFile(t, dir, rel, `package server

type Config struct {
	AllowOrigins     []string
	AllowCredentials bool
}

func Default() Config {
	return Config{AllowOrigins: []string{"*"}, AllowCredentials: true}
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-05" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest06_flagsTrustHeader(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/handler/auth.go"
	path := writeGoFile(t, dir, rel, `package handler

import "net/http"

func Auth(w http.ResponseWriter, r *http.Request) {
	_ = r.Header.Get("X-User-Id")
}
`)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}},
		pkgs:       []PackageInfo{{RelDir: "internal/handler"}},
	}
	findings, err := NewRest06().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-06" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest07_flagsErrErrorInJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/handler/handler.go"
	path := writeGoFile(t, dir, rel, `package handler

import "github.com/gin-gonic/gin"

func Handle(c *gin.Context, err error) {
	c.JSON(500, gin.H{"error": err.Error()})
}
`)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}},
		pkgs:       []PackageInfo{{RelDir: "internal/handler"}},
	}
	findings, err := NewRest07().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-07" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest08_flagsGinDebugMode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "github.com/gin-gonic/gin"

func main() {
	gin.SetMode(gin.DebugMode)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest08().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-08" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest08_skipsDevBuildTag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `//go:build dev

package main

import "github.com/gin-gonic/gin"

func main() {
	gin.SetMode(gin.DebugMode)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest08().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest09_flagsGorillaMux(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir},
		mod:        []byte("module example.com/app\n\ngo 1.22\n\nrequire github.com/gorilla/mux v1.8.1\n"),
	}
	findings, err := NewRest09().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-09" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest10_flagsNoRecover(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.New()
	r.Run(":8080")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest10().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "rest-10" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRest10_allowsRecoverMiddleware(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Run(":8080")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewRest10().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRegistry_hasAllRestChecks(t *testing.T) {
	t.Parallel()

	want := []string{"rest-01", "rest-02", "rest-03", "rest-04", "rest-05", "rest-06", "rest-07", "rest-08", "rest-09", "rest-10"}
	seen := make(map[string]struct{})
	for _, id := range AllCheckIDs() {
		seen[id] = struct{}{}
	}
	for _, id := range want {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing %s in registry", id)
		}
		meta, ok := MetaForID(id)
		if !ok || meta.Domain != "rest" {
			t.Fatalf("meta for %s = %+v", id, meta)
		}
	}
}

func TestCatalogForStack_includesRest(t *testing.T) {
	t.Parallel()

	cat := CatalogForStack(restStacks())
	var found int
	for _, ch := range cat {
		if strings.HasPrefix(ch.ID(), "rest-") {
			found++
		}
	}
	if found != 10 {
		t.Fatalf("expected 10 rest checks, got %d", found)
	}
}

func TestFixturePathsExist(t *testing.T) {
	t.Parallel()

	names := []string{
		"bad-fat-handler",
		"bad-handler-sql",
		"bad-server-timeouts",
		"bad-no-body-limit",
		"bad-cors-wildcard",
		"bad-header-auth",
		"bad-error-leak",
		"bad-gin-debug",
		"bad-no-recover",
	}
	for _, name := range names {
		p := filepath.Join("..", "..", "testdata", "fixtures", name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("fixture %s: %v", name, err)
		}
	}
}

func TestBuildTagHasDev_negation(t *testing.T) {
	t.Parallel()
	if buildTagHasDev("!dev") {
		t.Fatal("!dev must not count as dev build")
	}
	if !buildTagHasDev("dev") {
		t.Fatal("dev must count as dev build")
	}
}
