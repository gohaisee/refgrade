package detect_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gohaisee/refgrade/internal/detect"
	"github.com/gohaisee/refgrade/internal/project"
)

func TestDetect_goodMinimal_empty(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	mod, err := project.Load(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	stacks, err := detect.Detect(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(stacks) != 0 {
		t.Fatalf("stacks = %+v", stacks)
	}
}

func TestDetect_findsGin(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/ginapp\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	svcDir := filepath.Join(dir, "internal", "api")
	if err := os.MkdirAll(svcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(svcDir, "api.go"), []byte(`package api

import "github.com/gin-gonic/gin"

func Router() *gin.Engine {
	return gin.New()
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	mod, err := project.Load(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	stacks, err := detect.Detect(context.Background(), mod)
	if err != nil {
		t.Fatal(err)
	}
	if len(stacks) != 1 || stacks[0].Name != "gin" {
		t.Fatalf("stacks = %+v", stacks)
	}
}

func TestMatchImport(t *testing.T) {
	t.Parallel()

	name, ok := detect.MatchImport("github.com/gin-gonic/gin")
	if !ok || name != "gin" {
		t.Fatalf("MatchImport = %q %v", name, ok)
	}
	_, ok = detect.MatchImport("example.com/unknown")
	if ok {
		t.Fatal("expected false")
	}
}

func TestFormatText_none(t *testing.T) {
	t.Parallel()

	out := detect.FormatText(nil, "title", "none")
	if out != "title\nnone\n" {
		t.Fatalf("out = %q", out)
	}
}

func TestFormatText_withStacks(t *testing.T) {
	t.Parallel()

	out := detect.FormatText([]detect.Stack{{Name: "gin", ImportPath: "github.com/gin-gonic/gin"}}, "title", "none")
	if out == "title\nnone\n" {
		t.Fatalf("out = %q", out)
	}
}
