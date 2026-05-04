package subprocess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// result of external command
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

var ErrNotFound = errors.New("executable not found")

// runs cmd with timeout; returns ErrNotFound when binary missing
func Run(ctx context.Context, name string, args []string, timeout time.Duration) (*Result, error) {
	return RunInDir(ctx, "", name, args, timeout)
}

// RunInDir runs cmd in dir with timeout
func RunInDir(ctx context.Context, dir, name string, args []string, timeout time.Duration) (*Result, error) {
	if _, err := exec.LookPath(name); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := &Result{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if err == nil {
		return res, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
		return res, nil
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return res, fmt.Errorf("timeout after %s: %w", timeout, err)
	}
	return res, err
}

// module dir for go commands
func ModuleDir(root string) string {
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return root
}
