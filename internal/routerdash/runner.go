package routerdash

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type TimeoutRunner interface {
	RunWithTimeout(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error)
}

func NewRunner(fake bool) Runner {
	if fake {
		return FakeRunner{}
	}
	return ExecRunner{Timeout: 8 * time.Second}
}

type ExecRunner struct {
	Timeout time.Duration
}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	return r.RunWithTimeout(ctx, r.Timeout, name, args...)
}

func (r ExecRunner) RunWithTimeout(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	commandTimeout := timeout
	if commandTimeout == 0 {
		commandTimeout = 8 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return string(out), ctx.Err()
	}
	if err != nil {
		return string(out), fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return string(out), nil
}

func errUnavailable(name string, err error) Availability {
	if err == nil {
		return Availability{Available: true}
	}
	message := err.Error()
	if errors.Is(err, exec.ErrNotFound) {
		message = name + " is not installed"
	}
	return Availability{Available: false, Error: message}
}
