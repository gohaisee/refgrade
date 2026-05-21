package check

import (
	"context"
	"testing"
)

func TestSql01_flagsExecInLoop(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Activate(ctx context.Context, pool *pgxpool.Pool, ids []int) error {
	for _, id := range ids {
		_, err := pool.Exec(ctx, "UPDATE users SET active = true WHERE id = $1", id)
		if err != nil {
			return err
		}
	}
	return nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewSql01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sql-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSql02_flagsSprintfSQL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import (
	"database/sql"
	"fmt"
)

func ByEmail(db *sql.DB, email string) error {
	q := fmt.Sprintf("SELECT id FROM users WHERE email = '%s'", email)
	_, err := db.Query(q)
	return err
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewSql02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sql-02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSql03_flagsListWithoutLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import "database/sql"

func ListUsers(db *sql.DB) error {
	_, err := db.Query("SELECT id, name FROM users")
	return err
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewSql03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSql05_flagsPoolInHandler(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/handler/handler.go", `package handler

import (
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Serve(w http.ResponseWriter, r *http.Request) {
	pool, _ := pgxpool.New(r.Context(), "postgres://localhost/db")
	_ = pool
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/handler/handler.go"}}}
	findings, err := NewSql05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestPgx01_flagsConnect(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/db/db.go", `package db

import (
	"context"
	"github.com/jackc/pgx/v5"
)

func Open(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, "postgres://localhost/db")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/db/db.go"}}}
	findings, err := NewPgx01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestPgx03_flagsRemoteSSLDisable(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/config/config.go", `package config

const DSN = "postgres://user:pass@db.example.com:5432/app?sslmode=disable"
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/config/config.go"}}}
	findings, err := NewPgx03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGorm01_flagsRawConcat(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import "gorm.io/gorm"

func Find(db *gorm.DB, name string) {
	db.Raw("SELECT * FROM users WHERE name = " + name).Scan(nil)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewGorm01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGorm03_flagsAutoMigrateInMain(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "cmd/app/main.go", `package main

import "gorm.io/gorm"

type User struct{}

func main() {
	var db *gorm.DB
	_ = db.AutoMigrate(&User{})
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "cmd/app/main.go"}}}
	findings, err := NewGorm03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestGorm05_flagsGlobalDB(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/store/store.go", `package store

import "gorm.io/gorm"

var DB *gorm.DB
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/store/store.go"}}}
	findings, err := NewGorm05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSqlx01_flagsGetWithoutContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import "github.com/jmoiron/sqlx"

func Load(db *sqlx.DB, id int) error {
	var name string
	return db.Get(&name, "SELECT name FROM users WHERE id = $1", id)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewSqlx01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSqlc01_flagsHandEditedSQLGo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/db/users.sql.go", `package db

func GetUser() {}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/db/users.sql.go"}}}
	findings, err := NewSqlc01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestEnt01_flagsEdgesInLoop(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/service/service.go", `package service

import "entgo.io/ent"

type User struct {
	Edges struct{ Posts []Post }
}
type Post struct{}

func Walk(users []User) {
	for _, u := range users {
		_ = u.Edges
	}
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/service/service.go"}}}
	findings, err := NewEnt01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}
