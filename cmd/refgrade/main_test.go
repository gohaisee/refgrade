package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_scan_badIgnoredError_exitOK(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-ignored-error")
	code := run([]string{"scan", root})
	if code != exitOK {
		t.Fatalf("exit code = %d, warn should not fail scan", code)
	}
}

func TestRun_explain_err01_ru(t *testing.T) {
	var buf bytes.Buffer
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := run([]string{"explain", "err-01", "--lang", "ru"})
	w.Close()
	os.Stdout = old
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(buf.String(), "err-01") {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestRun_scan_badGetenv_exitFail(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-getenv")
	code := run([]string{"scan", root, "--format", "json"})
	if code != exitFail {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_scan_badDefaultClient_exitFail(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-default-client")
	code := run([]string{"scan", root})
	if code != exitFail {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_scan_goodMinimal_exitOK(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	code := run([]string{"scan", root})
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_detect_goodMinimal(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	code := run([]string{"detect", root})
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_scan_outputFile(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	out := filepath.Join(t.TempDir(), "report.txt")
	code := run([]string{"scan", root, "-o", out})
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "good-minimal") {
		t.Fatalf("report = %q", data)
	}
}

func TestRun_scan_ru(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	var buf bytes.Buffer
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := run([]string{"scan", root, "--lang", "ru"})
	w.Close()
	os.Stdout = old
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_scan_invalidPath(t *testing.T) {
	code := run([]string{"scan", filepath.Join(t.TempDir(), "missing")})
	if code != exitError {
		t.Fatalf("exit code = %d", code)
	}
}

func TestWriteOutput_stdout(t *testing.T) {
	if err := writeOutput("", []byte("ok")); err != nil {
		t.Fatal(err)
	}
}

func TestRun_scan_markdown(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "bad-getenv")
	code := run([]string{"scan", root, "--format", "markdown"})
	if code != exitFail {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_detect_invalidPath(t *testing.T) {
	code := run([]string{"detect", filepath.Join(t.TempDir(), "missing")})
	if code != exitError {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_scan_badFormat(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	code := run([]string{"scan", root, "--format", "xml"})
	if code != exitError {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_explain_cfg01_ru(t *testing.T) {
	var buf bytes.Buffer
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := run([]string{"explain", "cfg-01", "--lang", "ru"})
	w.Close()
	os.Stdout = old
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(buf.String(), "cfg-01") {
		t.Fatalf("output = %q", buf.String())
	}
}

func TestRun_explain_unknown(t *testing.T) {
	code := run([]string{"explain", "nope-99"})
	if code != exitError {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_explain_missingID(t *testing.T) {
	code := run([]string{"explain"})
	if code != exitError {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_help(t *testing.T) {
	code := run([]string{"help"})
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_langPrecedence_yaml(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "fixtures", "good-minimal")
	dir := t.TempDir()
	if err := copyDir(src, dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".refgrade.yaml"), []byte("lang: ru\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := run([]string{"scan", dir})
	w.Close()
	os.Stdout = old
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if code != exitOK {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(buf.String(), "отчёт refgrade") || !strings.Contains(buf.String(), "модуль:") {
		t.Fatalf("expected ru from yaml, got %q", buf.String())
	}

	t.Setenv("REFGRADE_LANG", "en")
	buf.Reset()
	r, w, err = os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code = run([]string{"scan", dir})
	w.Close()
	os.Stdout = old
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "refgrade scan report") || !strings.Contains(buf.String(), "module:") {
		t.Fatalf("expected en from env, got %q", buf.String())
	}

	t.Setenv("REFGRADE_LANG", "")
	buf.Reset()
	r, w, err = os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code = run([]string{"scan", dir, "--lang", "en"})
	w.Close()
	os.Stdout = old
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "refgrade scan report") || !strings.Contains(buf.String(), "module:") {
		t.Fatalf("expected en from flag, got %q", buf.String())
	}

}

func TestRun_unknownCommand(t *testing.T) {
	code := run([]string{"nope"})
	if code != exitError {
		t.Fatalf("exit code = %d", code)
	}
}

func TestRun_noArgs(t *testing.T) {
	code := run(nil)
	if code != exitError {
		t.Fatalf("exit code = %d", code)
	}
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

