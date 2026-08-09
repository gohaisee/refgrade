package check

import (
	"context"
	"testing"
)

func TestMongo01_flagsConnectInHandler(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/handler/handler.go", `package handler

import (
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Users(w http.ResponseWriter, r *http.Request) {
	client, _ := mongo.Connect(r.Context(), options.Client().ApplyURI("mongodb://localhost"))
	_ = client
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/handler/handler.go"}}}
	findings, err := NewMongo01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "mongo-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMongo02_flagsBackgroundCtx(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/handler/handler.go", `package handler

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Users(w http.ResponseWriter, r *http.Request, client *mongo.Client) {
	coll := client.Database("app").Collection("users")
	_, _ = coll.Find(context.Background(), bson.M{})
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/handler/handler.go"}}}
	findings, err := NewMongo02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "mongo-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMongo03_flagsListWithoutLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ListUsers(ctx context.Context, client *mongo.Client) error {
	coll := client.Database("app").Collection("users")
	_, err := coll.Find(ctx, bson.M{"active": true})
	return err
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewMongo03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMongo04_flagsNoPingInMain(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "cmd/app/main.go", `package main

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	_, _ = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost"))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "cmd/app/main.go"}}}
	findings, err := NewMongo04().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMongo05_flagsDefaultPool(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/db/db.go", `package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Open(ctx context.Context) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost"))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/db/db.go"}}}
	findings, err := NewMongo05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMongo06_flagsWhereInjection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import "go.mongodb.org/mongo-driver/bson"

func Filter(input string) bson.M {
	return bson.M{"$where": input}
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewMongo06().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMongo07_flagsRegexWithoutEscape(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import "go.mongodb.org/mongo-driver/bson"

func Search(term string) bson.M {
	return bson.M{"name": bson.M{"$regex": term}}
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewMongo07().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestMongo08_flagsBsonInHandler(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/handler/handler.go", `package handler

import "go.mongodb.org/mongo-driver/bson"

type userDoc struct {
	ID string ` + "`bson:\"_id\"`" + `
}
`)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/handler/handler.go"}}},
		pkgs:       []PackageInfo{{ImportPath: "example.com/app/internal/handler", RelDir: "internal/handler"}},
	}
	findings, err := NewMongo08().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}
