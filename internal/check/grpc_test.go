package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func grpcStacks() []string {
	return []string{"grpc", "connect", "grpc-gateway"}
}

func TestGrpc01_flagsInsecureDial(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	_, _ = grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "grpc-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGrpc01_singleFindingForDialWithInsecure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	_, _ = grpc.Dial("x", grpc.WithTransportCredentials(insecure.NewCredentials()))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %+v", findings)
	}
}

func TestGrpc01_skipsDevBuildTag(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `//go:build dev

package main

import "google.golang.org/grpc"

func main() {
	_, _ = grpc.Dial("localhost:50051")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGrpc02_flagsNewServerWithoutInterceptors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "google.golang.org/grpc"

func main() {
	s := grpc.NewServer()
	_ = s
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "grpc-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGrpc02_allowsInterceptors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "google.golang.org/grpc"

func main() {
	_ = grpc.NewServer(grpc.UnaryInterceptor(func(ctx interface{}, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGrpc03_flagsFatHandlerWithSQL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/grpc/server.go"
	body := "package grpc\n\nimport \"database/sql\"\n\nfunc Serve(db *sql.DB) {\n"
	for i := 0; i < 55; i++ {
		body += "\t_ = 1\n"
	}
	body += "\t_ = db.QueryRow(\"select 1\")\n}\n"
	path := writeGoFile(t, dir, rel, body)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}},
		pkgs:       []PackageInfo{{RelDir: "internal/grpc"}},
	}
	findings, err := NewGrpc03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "grpc-03" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGrpc04_flagsMetadataUserID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/grpc/auth.go"
	path := writeGoFile(t, dir, rel, `package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func User(ctx context.Context) string {
	md, _ := metadata.FromIncomingContext(ctx)
	return md.Get("x-user-id")[0]
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc04().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "grpc-04" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGrpc05_flagsReflection(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	s := grpc.NewServer()
	reflection.Register(s)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "grpc-05" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGrpc06_flagsMissingGracefulStop(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "google.golang.org/grpc"

func main() {
	s := grpc.NewServer()
	_ = s
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGrpc06().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "grpc-06" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestConn01_flagsClientWithoutTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/client/client.go"
	path := writeGoFile(t, dir, rel, `package client

import (
	"net/http"

	"connectrpc.com/connect"
)

func New() *connect.Client[struct{}, struct{}] {
	return connect.NewClient[struct{}, struct{}](http.DefaultClient, "http://localhost:8080")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewConn01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "conn-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestConn01_allowsWithTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/client/client.go"
	path := writeGoFile(t, dir, rel, `package client

import (
	"net/http"
	"time"

	"connectrpc.com/connect"
)

func New() *connect.Client[struct{}, struct{}] {
	return connect.NewClient[struct{}, struct{}](http.DefaultClient, "http://localhost:8080", connect.WithTimeout(time.Second))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewConn01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestConn02_flagsHandlerWithoutInterceptors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/handler/handler.go"
	path := writeGoFile(t, dir, rel, `package handler

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
)

func New() *connect.Handler[struct{}, struct{}] {
	return connect.NewUnaryHandler("/svc", func(context.Context, *connect.Request[struct{}]) (*connect.Response[struct{}], error) {
		return connect.NewResponse(&struct{}{}), nil
	}, connect.WithCodec(nil), connect.WithHandlerOptions())
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewConn02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "conn-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGw01_flagsGatewayWithoutAuth(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

func main() {
	_ = runtime.NewServeMux()
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGw01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "gw-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGw02_flagsRegisterBeforeAuth(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "cmd/app/main.go"
	path := writeGoFile(t, dir, rel, `package main

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func main() {
	mux := runtime.NewServeMux()
	_ = runtime.RegisterGatewayHandler(context.Background(), mux, nil, nil, nil, nil)
	http.Handle("/", AuthMiddleware(mux))
}

func AuthMiddleware(next http.Handler) http.Handler { return next }
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewGw02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "gw-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRegistry_hasAllGRPC(t *testing.T) {
	t.Parallel()

	want := []string{
		"grpc-01", "grpc-02", "grpc-03", "grpc-04", "grpc-05", "grpc-06",
		"conn-01", "conn-02", "gw-01", "gw-02",
	}
	seen := make(map[string]struct{})
	for _, id := range AllCheckIDs() {
		seen[id] = struct{}{}
	}
	for _, id := range want {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing %s in registry", id)
		}
		meta, ok := MetaForID(id)
		if !ok || meta.Domain != "grpc" {
			t.Fatalf("meta for %s = %+v", id, meta)
		}
	}
}

func TestCatalogForStack_includesGRPC(t *testing.T) {
	t.Parallel()

	cat := CatalogForStack(grpcStacks())
	var found int
	for _, ch := range cat {
		id := ch.ID()
		if strings.HasPrefix(id, "grpc-") || strings.HasPrefix(id, "conn-") || strings.HasPrefix(id, "gw-") {
			found++
		}
	}
	if found != 10 {
		t.Fatalf("expected 10 grpc/connect/gw checks, got %d", found)
	}
}

func TestFixturePathsExist_grpc(t *testing.T) {
	t.Parallel()

	names := []string{
		"bad-grpc-insecure",
		"bad-grpc-no-interceptor",
		"bad-grpc-metadata-auth",
		"bad-grpc-reflection",
		"bad-connect-timeout",
	}
	for _, name := range names {
		p := filepath.Join("..", "..", "testdata", "fixtures", name)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("fixture %s: %v", name, err)
		}
	}
}
