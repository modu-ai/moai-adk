//go:build !windows

// Package cli — the codex-launcher arm of the audit-gates fail-closed ledger
// (SPEC-AUDIT-CEILING-002 M3, AC-ACR-011's R4), split out of
// audit_gates_failclosed_test.go so the Windows build keeps the
// platform-neutral arms: the fake codex is a POSIX shell script (see
// codex_audit_launch_test.go), so this arm cannot build or run on windows.
package cli

import (
	"strings"
	"testing"
)

// TestGateErrorPropagatesToCallerSitesCodexPrepare — the plan file 18 ledger's
// prepareCodexAudit site (the sibling test carries the other five): a launch
// whose audit configuration cannot be read must refuse with the pins error
// surfaced.
func TestGateErrorPropagatesToCallerSitesCodexPrepare(t *testing.T) {
	repo := newAuditRepo(t)
	installFakeCodex(t)
	breakTreeWorkflow(t, repo.a1)

	r := runAudit(t, codexAuditRequest{Role: "plan-auditor", ProjectRoot: repo.a, Root: repo.a1})
	if r.res.ExitCode == 0 {
		t.Fatal("a launch whose audit configuration cannot be read must refuse")
	}
	if !strings.Contains(r.stderr, "workflow.audit pins unreadable") || !strings.Contains(r.stderr, "workflow.yaml") {
		t.Errorf("stderr = %q, want the pins error surfaced", r.stderr)
	}
}
