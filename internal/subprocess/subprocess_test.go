package subprocess

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRun_echo(t *testing.T) {
	t.Parallel()
	res, err := Run(context.Background(), "echo", []string{"ok"}, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytesContains(res.Stdout, "ok") {
		t.Fatalf("stdout = %q", res.Stdout)
	}
}

func TestRun_notFound(t *testing.T) {
	t.Parallel()
	_, err := Run(context.Background(), "refgrade-nonexistent-binary-xyz", nil, time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func bytesContains(b []byte, s string) bool {
	return len(b) >= len(s) && string(b) == s+"\n" || string(b) == s
}
