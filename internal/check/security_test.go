package check

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSec01_flagsWeakPasswordHash(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/auth/auth.go", `package auth

import (
	"crypto/md5"
)

func HashPassword(password string) []byte {
	return md5.Sum([]byte(password))
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/auth/auth.go"}}}
	findings, err := NewSec01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSec02_flagsInsecureTLS(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "cmd/app/main.go", `package main

import "crypto/tls"

func main() {
	_ = tls.Config{InsecureSkipVerify: true}
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "cmd/app/main.go"}}}
	findings, err := NewSec02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSec03_flagsMathRandToken(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/auth/token.go", `package auth

import "math/rand"

func SessionToken() int {
	return rand.Int()
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/auth/token.go"}}}
	findings, err := NewSec03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-03" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSec05_flagsJWTNone(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/auth/jwt.go", `package auth

import "github.com/golang-jwt/jwt/v5"

func Bad() {
	_ = jwt.SigningMethodNone
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/auth/jwt.go"}}}
	findings, err := NewSec05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-05" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSec08_flagsTemplateHTML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/view/view.go", `package view

import "html/template"

func Render(userInput string) template.HTML {
	return template.HTML(userInput)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/view/view.go"}}}
	findings, err := NewSec08().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-08" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSec10_flagsTrackedDotenv(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("KEY=val\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mod := stubModule{root: dir}
	findings, err := NewSec10().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-10" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecR03_flagsMassAssign(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	handler := writeGoFile(t, dir, "internal/handler/handler.go", `package handler

import (
	"encoding/json"
	"net/http"

	"example.com/app/internal/model"
)

func Update(w http.ResponseWriter, r *http.Request) {
	_ = json.Unmarshal([]byte("{}"), &model.User{})
}
`)
	model := writeGoFile(t, dir, "internal/model/user.go", `package model

type User struct {
	ID string
}
`)
	mod := stubModuleWithPkgs{
		stubModule: stubModule{root: dir, files: []GoFile{
			{Path: handler, RelPath: "internal/handler/handler.go"},
			{Path: model, RelPath: "internal/model/user.go"},
		}},
		pkgs: []PackageInfo{{RelDir: "internal/handler"}},
	}
	findings, err := NewSecR03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-r03" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecDB05_flagsDynamicTable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import (
	"database/sql"
	"fmt"
)

func ByTable(db *sql.DB, tableName string) error {
	q := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", tableName)
	_, err := db.Query(q)
	return err
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewSecDB05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) < 1 || findings[0].ID != "sec-db05" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRegistry_hasSecurityChecks(t *testing.T) {
	t.Parallel()
	want := []string{
		"sec-01", "sec-02", "sec-03", "sec-04", "sec-05", "sec-07", "sec-08", "sec-09", "sec-10",
		"sec-13", "sec-14", "sec-15", "sec-16",
		"sec-r01", "sec-r03", "sec-r04", "sec-r05", "sec-r06", "sec-r09", "sec-r10",
		"sec-g01", "sec-g02", "sec-g03", "sec-g04", "sec-g05", "sec-g06",
		"sec-g07", "sec-g08", "sec-g09", "sec-g10",
		"sec-db01", "sec-db02", "sec-db03", "sec-db05",
		"sec-m01", "sec-m02", "sec-rd01", "sec-rd02", "sec-mq01", "sec-mq02",
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
		if !ok || meta.Domain != "security" {
			t.Fatalf("meta for %s = %+v", id, meta)
		}
	}
}
