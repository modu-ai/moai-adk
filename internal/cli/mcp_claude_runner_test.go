package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestClaudeRealRunner_CapturesBoundedStdout(t *testing.T) {
	if os.Getenv("MOAI_CLAUDE_RUNNER_HELPER") == "1" {
		_, _ = fmt.Fprint(os.Stdout, `{"ok":true}`)
		return
	}
	runner := realClaudeRunner{}
	env := append(os.Environ(), "MOAI_CLAUDE_RUNNER_HELPER=1")
	stdout, _, err := runner.RunAudit(
		context.Background(),
		os.Args[0],
		"",
		[]string{"-test.run=^TestClaudeRealRunner_CapturesBoundedStdout$"},
		env,
		nil,
	)
	if err != nil {
		t.Fatalf("real runner: %v", err)
	}
	if !strings.Contains(string(stdout), `{"ok":true}`) {
		t.Fatalf("stdout = %q, want helper JSON payload", string(stdout))
	}
}

func TestClaudeRealRunner_OutputLimitReturnsTruncationError_AC_CLA_014(t *testing.T) {
	if os.Getenv("MOAI_CLAUDE_TRUNCATION_HELPER") == "1" {
		_, _ = fmt.Fprint(os.Stdout, strings.Repeat("x", claudeAuditOutputLimit+1))
		return
	}
	runner := realClaudeRunner{}
	env := append(os.Environ(), "MOAI_CLAUDE_TRUNCATION_HELPER=1")
	stdout, _, err := runner.RunAudit(
		context.Background(),
		os.Args[0],
		"",
		[]string{"-test.run=^TestClaudeRealRunner_OutputLimitReturnsTruncationError_AC_CLA_014$"},
		env,
		nil,
	)
	if !errors.Is(err, errClaudeAuditOutputTruncated) {
		t.Fatalf("runner error = %v, want errClaudeAuditOutputTruncated", err)
	}
	if len(stdout) != claudeAuditOutputLimit {
		t.Fatalf("bounded stdout len = %d, want %d", len(stdout), claudeAuditOutputLimit)
	}
}
