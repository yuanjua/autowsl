package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Runner executes external commands
type Runner interface {
	Run(name string, args ...string) (stdout string, stderr string, err error)
	RunWithInput(name string, stdin string, args ...string) (stdout string, stderr string, err error)
}

// ExecRunner executes real system commands with timeout support
type ExecRunner struct {
	Timeout time.Duration
	DryRun  bool
}

// NewExecRunner creates a new runner with the given timeout
func NewExecRunner(timeout time.Duration) *ExecRunner {
	return &ExecRunner{
		Timeout: timeout,
		DryRun:  false,
	}
}

// Run executes a command and returns stdout, stderr, and error
func (r *ExecRunner) Run(name string, args ...string) (string, string, error) {
	if r.DryRun {
		return r.dryRunLog(name, args...), "", nil
	}

	// Create context with or without timeout
	var ctx context.Context
	var cancel context.CancelFunc
	if r.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), r.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB

	err := cmd.Run()
	return outB.String(), errB.String(), err
}

// RunWithInput executes a command with stdin and returns stdout, stderr, and error
func (r *ExecRunner) RunWithInput(name string, stdin string, args ...string) (string, string, error) {
	if r.DryRun {
		return r.dryRunLog(name, args...), "", nil
	}

	// Create context with or without timeout
	var ctx context.Context
	var cancel context.CancelFunc
	if r.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), r.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB
	cmd.Stdin = bytes.NewBufferString(stdin)

	err := cmd.Run()
	return outB.String(), errB.String(), err
}

func (r *ExecRunner) dryRunLog(name string, args ...string) string {
	return fmt.Sprintf("[dry-run] %s %v", name, args)
}
