package cli

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
)

type claudeCommandRunner interface {
	RunAuthStatus(ctx context.Context, binary string, env []string) ([]byte, []byte, error)
	RunAudit(ctx context.Context, binary, dir string, args, env []string, stdin []byte) ([]byte, []byte, error)
}

type realClaudeRunner struct{}

var (
	claudeRunner   claudeCommandRunner = realClaudeRunner{}
	claudeLookPath                     = exec.LookPath
)

type boundedWriter struct {
	buf       bytes.Buffer
	n         int
	truncated bool
}

var errClaudeAuditOutputTruncated = errors.New("claude audit process output exceeded the bounded capture limit")

func (w *boundedWriter) Write(p []byte) (int, error) {
	original := len(p)
	remaining := claudeAuditOutputLimit - w.n
	if remaining > 0 {
		if len(p) > remaining {
			w.truncated = true
			p = p[:remaining]
		}
		written, _ := w.buf.Write(p)
		w.n += written
	} else if len(p) > 0 {
		w.truncated = true
	}
	return original, nil
}

func (w *boundedWriter) Bytes() []byte { return w.buf.Bytes() }

func runClaudeCommand(ctx context.Context, binary, dir string, args, env []string, stdin []byte) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdin = bytes.NewReader(stdin)
	stdout := &boundedWriter{}
	stderr := &boundedWriter{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	configureClaudeAuditProcess(cmd)
	cmd.WaitDelay = 2 * time.Second
	err := runClaudeAuditProcess(cmd)
	if stdout.truncated || stderr.truncated {
		if err == nil {
			err = errClaudeAuditOutputTruncated
		} else {
			err = errors.Join(err, errClaudeAuditOutputTruncated)
		}
	}
	return append([]byte(nil), stdout.Bytes()...), append([]byte(nil), stderr.Bytes()...), err
}

func (realClaudeRunner) RunAuthStatus(ctx context.Context, binary string, env []string) ([]byte, []byte, error) {
	return runClaudeCommand(ctx, binary, "", []string{"auth", "status", "--json"}, env, nil)
}

func (realClaudeRunner) RunAudit(ctx context.Context, binary, dir string, args, env []string, stdin []byte) ([]byte, []byte, error) {
	return runClaudeCommand(ctx, binary, dir, args, env, stdin)
}
