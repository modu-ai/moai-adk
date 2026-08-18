// role_test.go — the role-declaration carrier (SPEC-KANBAN-BOARD-001
// REQ-KB-025, M1).
//
// The CONTRACT this carrier satisfies is REQ-KS-006's, consumed whole and by
// reference — these tests assert its observable properties (both cross-session
// directions, label-distinctness, uniqueness of the surface), never a summary
// of the contract. The carrier lives at the board root (single origin), so
// any session — lead or worker — resolves any other session's declaration;
// a session-private or lead-only carrier is the fork AP-30 exists to prevent.
package kanban

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRoleDeclaration_WorkersReadLead — AC-KB-017 carrier half, direction 1:
// a session that is NOT the lead resolves the lead's declared role to the
// value the subject session declared. This is the direction the worktree
// sibling's gates (REQ-KW-007 / REQ-KW-011) consume.
func TestRoleDeclaration_WorkersReadLead(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	if err := DeclareRole(root, "lead-sess-1", RoleLead, "plan-host"); err != nil {
		t.Fatalf("DeclareRole(lead): %v", err)
	}

	// Resolved by a session that is not the lead — the API takes no caller
	// role, and the observation below runs it in a worker's context.
	if err := DeclareRole(root, "worker-sess-1", "run", "run-alpha"); err != nil {
		t.Fatalf("DeclareRole(run worker): %v", err)
	}
	role, err := ResolveDeclaredRole(root, "lead-sess-1")
	if err != nil {
		t.Fatalf("ResolveDeclaredRole(lead from worker context): %v", err)
	}
	if role != RoleLead {
		t.Fatalf("resolved lead role = %q, want %q", role, RoleLead)
	}
}

// TestRoleDeclaration_LeadReadsWorkers — AC-KB-017 carrier half, direction 2
// (the one a lead-only carrier fails): the lead resolves a NON-lead session's
// declared role. This is the key on which the lead selects a dispatch target
// (REQ-KS-019) and over which quorum is accounted (REQ-KS-012).
func TestRoleDeclaration_LeadReadsWorkers(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	if err := DeclareRole(root, "lead-sess-1", RoleLead, "plan-host"); err != nil {
		t.Fatalf("DeclareRole(lead): %v", err)
	}
	if err := DeclareRole(root, "worker-sess-2", "run", "run-beta"); err != nil {
		t.Fatalf("DeclareRole(run worker): %v", err)
	}

	role, err := ResolveDeclaredRole(root, "worker-sess-2")
	if err != nil {
		t.Fatalf("ResolveDeclaredRole(worker from lead context): %v", err)
	}
	if role != "run" {
		t.Fatalf("resolved worker role = %q, want %q — a lead-only carrier cannot serve this direction", role, "run")
	}
}

// TestRoleDeclaration_CrossProcessResolution — the carrier is resolvable from
// a DIFFERENT OS process, which is what "resolvable by a session that is not
// the lead" means operationally: sessions are distinct processes.
func TestRoleDeclaration_CrossProcessResolution(t *testing.T) {
	root := t.TempDir()
	if err := DeclareRole(root, "lead-sess-1", RoleLead, "plan-host"); err != nil {
		t.Fatalf("DeclareRole: %v", err)
	}

	got := runHelperProcess(t, "resolve-role", map[string]string{
		"HELPER_ROOT":    root,
		"HELPER_SESSION": "lead-sess-1",
	})
	if strings.TrimSpace(got) != RoleLead {
		t.Fatalf("subprocess resolved role = %q, want %q", strings.TrimSpace(got), RoleLead)
	}
}

// TestRoleDeclaration_LabelDistinct — AC-KB-017 carrier half (AC-KS-030
// shape): one role under two different labels declares identically. The
// declaration carries the label as a SEPARATE field and the role is resolved
// from the role field alone — no code path computes a role from a label,
// because a role does not determine a label and a label does not determine a
// role.
func TestRoleDeclaration_LabelDistinct(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	if err := DeclareRole(root, "sess-label-a", "run", "run-a1b2c3"); err != nil {
		t.Fatalf("DeclareRole(label-a): %v", err)
	}
	roleA, err := ResolveDeclaredRole(root, "sess-label-a")
	if err != nil {
		t.Fatalf("ResolveDeclaredRole(label-a): %v", err)
	}

	if err := DeclareRole(root, "sess-label-b", "run", "run-totally-different"); err != nil {
		t.Fatalf("DeclareRole(label-b): %v", err)
	}
	roleB, err := ResolveDeclaredRole(root, "sess-label-b")
	if err != nil {
		t.Fatalf("ResolveDeclaredRole(label-b): %v", err)
	}

	if roleA != roleB {
		t.Fatalf("same role under two labels resolved differently: %q vs %q", roleA, roleB)
	}

	// The declaration artifact records both data as distinct fields.
	raw, err := os.ReadFile(roleDeclarationPath(root, "sess-label-a"))
	if err != nil {
		t.Fatalf("read declaration: %v", err)
	}
	body := string(raw)
	if !strings.Contains(body, `"role": "run"`) && !strings.Contains(body, `"role":"run"`) {
		t.Fatalf("declaration does not carry the role field: %s", body)
	}
	if !strings.Contains(body, "run-a1b2c3") {
		t.Fatalf("declaration does not record the label as its own datum: %s", body)
	}
}

// TestResolveDeclaredRole_AbsentFails — an undeclared session has no role to
// read. The failure is load-bearing: the board's write guard treats an
// unreadable role as a refusal (fail-closed), never as an admission.
func TestResolveDeclaredRole_AbsentFails(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if _, err := ResolveDeclaredRole(root, "never-declared"); err == nil {
		t.Fatal("ResolveDeclaredRole(undeclared) err = nil, want non-nil (fail-closed)")
	}
}

// TestResolveDeclaredRole_RejectsUnsafeSessionIDs — the declaration path is
// derived from the session id, so separator and parent-reference ids must be
// rejected rather than escaping the roles directory.
func TestResolveDeclaredRole_RejectsUnsafeSessionIDs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, bad := range []string{"", "..", "a/b", `a\b`} {
		if _, err := ResolveDeclaredRole(root, bad); err == nil {
			t.Errorf("ResolveDeclaredRole(%q) err = nil, want non-nil", bad)
		}
		if err := DeclareRole(root, bad, "run", "lbl"); err == nil {
			t.Errorf("DeclareRole(%q) err = nil, want non-nil", bad)
		}
	}
}

// roleDeclarationPath is the declaration artifact's path beneath the board
// root (exposed for tests that inspect the artifact directly).
func roleDeclarationPath(root, sessionID string) string {
	return filepath.Join(BoardDir(root), "roles", sessionID+".json")
}

// runHelperProcess re-executes the test binary as a separate OS process and
// returns its stdout. used by the cross-process criteria (AC-KB-017 carrier,
// AC-KB-019, AC-KB-023) — a goroutine measures nothing across sessions.
func runHelperProcess(t *testing.T, op string, env map[string]string) string {
	t.Helper()
	if isWindowsBuild() {
		t.Skip("helper re-exec uses sh-free env plumbing; windows covered by build only")
	}
	args := []string{"-test.run=TestKanbanHelperProcess", "--"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "MOAI_KANBAN_HELPER="+op)
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("helper %s failed: %v: %s", op, err, ee.Stderr)
		}
		t.Fatalf("helper %s: %v", op, err)
	}
	return string(out)
}

// isWindowsBuild reports the runtime OS (board_test.go owns the definition).
func isWindowsBuild() bool {
	return runtimeIsWindows()
}
