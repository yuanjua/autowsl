package wsl

import "github.com/yuanjua/autowsl/internal/runner"

// Client is a wrapper for executing WSL commands.
// It uses dependency injection to allow for easy testing.
type Client struct {
	runner runner.Runner
}

// NewClient creates a new WSL client with the provided runner.
func NewClient(r runner.Runner) *Client {
	return &Client{runner: r}
}

// DefaultClient returns a client configured with default settings.
func DefaultClient() *Client {
	return NewClient(runner.NewExecRunner(0)) // 0 = no timeout
}
