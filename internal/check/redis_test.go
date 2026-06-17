package check

import (
	"context"
	"testing"
)

func TestRedis01_flagsNewClientInHandler(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/handler/handler.go", `package handler

import (
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

func Cache(w http.ResponseWriter, r *http.Request) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	_ = client
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/handler/handler.go"}}}
	findings, err := NewRedis01().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != "redis-01" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis02_flagsSetWithoutTTL(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/cache/cache.go", `package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func Save(ctx context.Context, rdb *redis.Client, key, value string) error {
	return rdb.Set(ctx, key, value, 0).Err()
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/cache/cache.go"}}}
	findings, err := NewRedis02().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis03_flagsKeysInService(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/cache/cache.go", `package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func Purge(ctx context.Context, rdb *redis.Client) error {
	return rdb.Keys(ctx, "*").Err()
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/cache/cache.go"}}}
	findings, err := NewRedis03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis04_flagsCacheUpdateOnWrite(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/service/service.go", `package service

import (
	"context"
	"database/sql"

	redis "github.com/redis/go-redis/v9"
)

func UpdateUser(ctx context.Context, db *sql.DB, rdb *redis.Client, id int, name string) error {
	_, err := db.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", name, id)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "user", name, 0).Err()
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/service/service.go"}}}
	findings, err := NewRedis04().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis05_flagsAuthFromCacheOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/handler/auth.go", `package handler

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func SessionFromCache(ctx context.Context, rdb *redis.Client, token string) (bool, error) {
	_, err := rdb.Get(ctx, "session:"+token).Result()
	return err == nil, err
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/handler/auth.go"}}}
	findings, err := NewRedis05().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis06_flagsGetInLoop(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/cache/cache.go", `package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func LoadMany(ctx context.Context, rdb *redis.Client, keys []string) ([]string, error) {
	var out []string
	for _, key := range keys {
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, val)
	}
	return out, nil
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/cache/cache.go"}}}
	findings, err := NewRedis06().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis07_flagsFlatKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/cache/cache.go", `package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func Read(ctx context.Context, rdb *redis.Client) (string, error) {
	return rdb.Get(ctx, "users").Result()
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/cache/cache.go"}}}
	findings, err := NewRedis07().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis08_flagsStampede(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/service/service.go", `package service

import (
	"context"
	"database/sql"

	redis "github.com/redis/go-redis/v9"
)

func LoadUser(ctx context.Context, db *sql.DB, rdb *redis.Client, id int) (string, error) {
	val, err := rdb.Get(ctx, "user:1").Result()
	if err == nil {
		return val, nil
	}
	var name string
	err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", id).Scan(&name)
	return name, err
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/service/service.go"}}}
	findings, err := NewRedis08().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRedis09_flagsZeroPool(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeGoFile(t, dir, "internal/cache/cache.go", `package cache

import redis "github.com/redis/go-redis/v9"

func New() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "localhost:6379", PoolSize: 0})
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: "internal/cache/cache.go"}}}
	findings, err := NewRedis09().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}
