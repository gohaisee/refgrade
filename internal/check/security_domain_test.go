package check

import (
	"context"
	"testing"
)

func TestSecM01_flagsMongoPasswordURI(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/config/mongo.go", `package config

const URI = "mongodb://admin:secret@mongo.example.com:27017/db"
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/config/mongo.go"}}}
	findings, err := NewSecM01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-m01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecM02_flagsRemoteMongoNoTLS(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/config/mongo.go", `package config

const URI = "mongodb://mongo.example.com:27017/db"
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/config/mongo.go"}}}
	findings, err := NewSecM02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-m02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecRD01_flagsRedisPasswordURI(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/config/redis.go", `package config

const URI = "redis://:secret@redis.example.com:6379/0"
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/config/redis.go"}}}
	findings, err := NewSecRD01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-rd01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecMQ01_flagsBrokerCredentials(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/config/mq.go", `package config

const URL = "amqp://user:pass@broker.example.com:5672/"
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/config/mq.go"}}}
	findings, err := NewSecMQ01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-mq01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSec16_warnsWithoutWithSecurity(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mod := stubModule{root: dir}
	findings, err := NewSec16().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-16" || findings[0].Severity != SeverityWarn {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecMQ02_flagsPlaintextRemote(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/config/mq.go", `package config

const URL = "amqp://broker.example.com:5672/"
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/config/mq.go"}}}
	findings, err := NewSecMQ02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "sec-mq02" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestSecDB01_aliasSql02(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/repo/repo.go", `package repo

import (
	"database/sql"
	"fmt"
)

func ByID(db *sql.DB, id string) error {
	q := fmt.Sprintf("SELECT * FROM users WHERE id = %s", id)
	_, err := db.Query(q)
	return err
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/repo/repo.go"}}}
	findings, err := NewSecDB01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) < 1 || findings[0].ID != "sec-db01" {
		t.Fatalf("findings = %+v", findings)
	}
}
