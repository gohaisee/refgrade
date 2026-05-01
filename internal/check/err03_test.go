package check

import (
	"context"
	"testing"
)

func TestErr03_flagsHTTPGet(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/client/api.go"
	path := writeGoFile(t, dir, rel, `package client

import "net/http"

func Fetch(url string) (*http.Response, error) {
	return http.Get(url)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].ID != err03ID {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr03_flagsHTTPPost(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/client/api.go"
	path := writeGoFile(t, dir, rel, `package client

import "net/http"

func Post(url, body string) (*http.Response, error) {
	return http.Post(url, "text/plain", nil)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr03_flagsDefaultClient(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/client/api.go"
	path := writeGoFile(t, dir, rel, `package client

import "net/http"

func Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr03_allowsCustomClient(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/client/api.go"
	path := writeGoFile(t, dir, rel, `package client

import (
	"net/http"
	"time"
)

func Fetch(url string) (*http.Response, error) {
	c := &http.Client{Timeout: 5 * time.Second}
	return c.Get(url)
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr03_skipsTestFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rel := "internal/client/api_test.go"
	path := writeGoFile(t, dir, rel, `package client

import "net/http"

func TestX(t *testing.T) {
	_, _ = http.Get("http://example.com")
}
`)
	mod := stubModule{root: dir, files: []GoFile{{Path: path, RelPath: rel}}}
	findings, err := NewErr03().Run(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestErr03_ID(t *testing.T) {
	t.Parallel()

	if NewErr03().ID() != err03ID {
		t.Fatal("unexpected id")
	}
}

func TestIsHTTPPackageCall_negative(t *testing.T) {
	t.Parallel()

	if isHTTPPackageCall(nil, "Get") {
		t.Fatal("nil should be false")
	}
}

func TestExplain_known(t *testing.T) {
	t.Parallel()

	when, why, fix, ok := Explain("cfg-01", func(k string) string { return "v:" + k })
	if !ok || when == "" || why == "" || fix == "" {
		t.Fatalf("Explain = %q %q %q %v", when, why, fix, ok)
	}
}

func TestExplain_unknown(t *testing.T) {
	t.Parallel()

	_, _, _, ok := Explain("missing-99", func(k string) string { return k })
	if ok {
		t.Fatal("expected false")
	}
}

func TestKnown(t *testing.T) {
	t.Parallel()

	if !Known("cfg-01") || !Known("err-01") || !Known("err-03") {
		t.Fatal("expected known checks")
	}
	if Known("nope") {
		t.Fatal("unexpected known")
	}
}
