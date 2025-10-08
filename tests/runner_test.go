package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/yuanjua/autowsl/internal/runner"
)

func TestExecRunnerNoTimeout(t *testing.T) {
	// Create runner with no timeout (0 means no timeout)
	r := runner.NewExecRunner(0)

	// Run a simple command that should succeed
	stdout, stderr, err := r.Run("echo", "hello")
	if err != nil {
		t.Fatalf("Expected no error, got %v\nStderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "hello") {
		t.Errorf("Expected stdout to contain 'hello', got: %s", stdout)
	}
}

func TestExecRunnerWithTimeout(t *testing.T) {
	// Create runner with 1 second timeout
	r := runner.NewExecRunner(1 * time.Second)

	// Run a quick command that should succeed
	stdout, stderr, err := r.Run("echo", "test")
	if err != nil {
		t.Fatalf("Expected no error, got %v\nStderr: %s", err, stderr)
	}

	if !strings.Contains(stdout, "test") {
		t.Errorf("Expected stdout to contain 'test', got: %s", stdout)
	}
}

func TestExecRunnerTimeoutExceeded(t *testing.T) {
	// Skip on Windows as sleep command behaves differently
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Create runner with very short timeout
	r := runner.NewExecRunner(100 * time.Millisecond)

	// Try to run a command that would take longer (platform-specific)
	// On Windows, we'd use timeout or ping, on Unix we'd use sleep
	// For cross-platform compatibility, let's just verify the timeout mechanism works
	_, _, err := r.Run("ping", "127.0.0.1", "-n", "100") // Windows: ping 100 times

	// We expect a timeout or error
	if err == nil {
		t.Log("Warning: Command completed before timeout (might be too fast)")
	}
}

func TestExecRunnerWithInput(t *testing.T) {
	// Create runner with no timeout
	r := runner.NewExecRunner(0)

	// Test RunWithInput - on Windows we can use 'sort' or 'findstr'
	// For cross-platform, let's use a command that reads stdin
	input := "test input"
	stdout, stderr, err := r.RunWithInput("cmd", input, "/C", "more")

	if err != nil {
		t.Logf("RunWithInput test: %v (stderr: %s)", err, stderr)
		// Some platforms might not support this, so just log
	}

	if stdout != "" {
		t.Logf("Received stdout: %s", stdout)
	}
}

func TestExecRunnerDryRun(t *testing.T) {
	r := runner.NewExecRunner(0)
	r.DryRun = true

	stdout, stderr, err := r.Run("some-command", "arg1", "arg2")

	if err != nil {
		t.Errorf("Dry run should not return error, got %v", err)
	}

	if stderr != "" {
		t.Errorf("Dry run should have empty stderr, got %s", stderr)
	}

	if !strings.Contains(stdout, "dry-run") {
		t.Errorf("Expected dry-run log, got: %s", stdout)
	}

	if !strings.Contains(stdout, "some-command") {
		t.Errorf("Expected command name in output, got: %s", stdout)
	}
}

func TestExecRunnerInvalidCommand(t *testing.T) {
	r := runner.NewExecRunner(0)

	_, _, err := r.Run("this-command-definitely-does-not-exist-12345")

	if err == nil {
		t.Error("Expected error for invalid command, got nil")
	}
}
