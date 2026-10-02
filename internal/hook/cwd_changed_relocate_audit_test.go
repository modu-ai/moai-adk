package hook

// SPEC-SESSION-ANCHOR-ATTR-001 W2 — hook-side relocation audit + ownership
// plausibility (REQ-SAA-003..005). The relocation path resolves the target
// tree's git worktree lock context and hands it to the session registry, so
// every cwd rewrite carries a trigger-attributed audit row and the opt-in
// anchor relocation guard can refuse an implausible target.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/session"
)

// guardTestConfigProvider is the minimal ConfigProvider stub for the guard
// flag tests.
type guardTestConfigProvider struct {
	cfg *config.Config
}

func (p guardTestConfigProvider) Get() *config.Config { return p.cfg }

// stubRelocationGitContext swaps the git-context seam for the duration of a
// test and restores it afterwards.
func stubRelocationGitContext(t *testing.T, treeRoot, porcelain string, ok bool) {
	t.Helper()
	orig := relocationGitContext
	relocationGitContext = func(dir string) (string, string, bool) {
		return treeRoot, porcelain, ok
	}
	t.Cleanup(func() { relocationGitContext = orig })
}

// auditTestInput builds a HookInput whose candidate walk finds the registry
// written under root.
func auditTestInput(root, sessionID string) *HookInput {
	return &HookInput{
		SessionID: sessionID,
		OldCwd:    root,
		NewCwd:    root,
		CWD:       root,
	}
}

// writeAuditRegistry writes a one-entry registry under root and returns the
// entry's registry path.
func writeAuditRegistry(t *testing.T, root, sessionID string, pid int) {
	t.Helper()
	dir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir registry dir: %v", err)
	}
	host, _ := os.Hostname()
	entries := []session.Entry{{
		SessionID:     sessionID,
		SpecID:        "SPEC-X",
		Phase:         "run",
		StartedAt:     time.Now().UTC(),
		LastHeartbeat: time.Now().UTC(),
		PID:           pid,
		Host:          host,
		CWD:           "/tree/old",
	}}
	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatalf("marshal entries: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "active-sessions.json"), data, 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}
}

// readAuditRegistry reads back the registry entries under root.
func readAuditRegistry(t *testing.T, root string) []session.Entry {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".moai", "state", "active-sessions.json"))
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	var entries []session.Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("unmarshal registry: %v", err)
	}
	return entries
}

// readRelocationRows reads the relocation audit log under root.
func readRelocationRows(t *testing.T, root string) []session.RelocationAudit {
	t.Helper()
	data, err := os.ReadFile(session.RelocationAuditPath(root))
	if err != nil {
		t.Fatalf("read relocation audit log: %v", err)
	}
	var rows []session.RelocationAudit
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var row session.RelocationAudit
		if err := json.Unmarshal([]byte(line), &row); err == nil {
			rows = append(rows, row)
		}
	}
	return rows
}

// otherLiveLockPorcelain renders a porcelain text whose target-tree lock is
// held by the test process — alive by construction, distinct from the entry
// pid arm.
func otherLiveLockPorcelain(treeRoot string) string {
	return "worktree " + treeRoot + "\nlocked claude session t-other (pid " + strconv.Itoa(os.Getpid()) + " start Wed Sep 30)\n"
}

// TestRelocateSessionCwd_AuditsTriggerAndOwnershipFlag covers REQ-SAA-003/004
// end to end from the hook path: the audit row carries the trigger hook event
// and the ownership classification from the target-tree lock (here: another
// live card — advisory flag, relocation proceeds).
func TestRelocateSessionCwd_AuditsTriggerAndOwnershipFlag(t *testing.T) {
	root := t.TempDir()
	writeAuditRegistry(t, root, "sess-1", 4243)
	treeRoot := "/tree/other-target"
	stubRelocationGitContext(t, treeRoot, otherLiveLockPorcelain(treeRoot), true)

	relocateSessionCwd(auditTestInput(root, "sess-1"), treeRoot, "CwdChanged", false)

	entries := readAuditRegistry(t, root)
	if entries[0].CWD != treeRoot {
		t.Errorf("advisory flagged relocation must proceed, cwd = %q", entries[0].CWD)
	}
	rows := readRelocationRows(t, root)
	if len(rows) != 1 {
		t.Fatalf("want exactly 1 audit row, got %d", len(rows))
	}
	if rows[0].Trigger != "CwdChanged" {
		t.Errorf("trigger = %q, want CwdChanged", rows[0].Trigger)
	}
	if rows[0].Owner != session.OwnerOtherLive || !rows[0].Flagged {
		t.Errorf("owner = %q flagged=%v, want other-live flagged", rows[0].Owner, rows[0].Flagged)
	}
	if rows[0].HolderCardID != "t-other" {
		t.Errorf("holder_card_id = %q, want t-other", rows[0].HolderCardID)
	}
}

// TestRelocateSessionCwd_GuardRefusesFlaggedRelocation covers REQ-SAA-005's
// opt-in refusal arm: with the guard on, a flagged relocation is refused and
// the refusal recorded.
func TestRelocateSessionCwd_GuardRefusesFlaggedRelocation(t *testing.T) {
	root := t.TempDir()
	writeAuditRegistry(t, root, "sess-1", 4243)
	treeRoot := "/tree/other-target"
	stubRelocationGitContext(t, treeRoot, otherLiveLockPorcelain(treeRoot), true)

	relocateSessionCwd(auditTestInput(root, "sess-1"), treeRoot, "CwdChanged", true)

	entries := readAuditRegistry(t, root)
	if entries[0].CWD != "/tree/old" {
		t.Errorf("refused relocation must not rewrite cwd, got %q", entries[0].CWD)
	}
	rows := readRelocationRows(t, root)
	if len(rows) != 1 || !rows[0].Refused {
		t.Fatalf("refusal must be recorded in the audit log, got %+v", rows)
	}
}

// TestRelocateSessionCwd_GitContextFailureFailsOpen covers the fail-open
// doctrine: when the git context cannot be resolved, the relocation proceeds
// and the audit row still lands (owner "none" — no lock observable).
func TestRelocateSessionCwd_GitContextFailureFailsOpen(t *testing.T) {
	root := t.TempDir()
	writeAuditRegistry(t, root, "sess-1", 4243)
	stubRelocationGitContext(t, "", "", false)

	target := "/tree/unresolvable"
	relocateSessionCwd(auditTestInput(root, "sess-1"), target, "CwdChanged", false)

	entries := readAuditRegistry(t, root)
	if entries[0].CWD != target {
		t.Errorf("git-context failure must fail open, cwd = %q", entries[0].CWD)
	}
	rows := readRelocationRows(t, root)
	if len(rows) != 1 {
		t.Fatalf("audit row must still land, got %d", len(rows))
	}
	if rows[0].Owner != session.OwnerNone {
		t.Errorf("owner = %q, want none (no lock observable)", rows[0].Owner)
	}
}

// TestAnchorRelocationGuardEnabled covers the nil-safe config read for the
// opt-in guard flag.
func TestAnchorRelocationGuardEnabled(t *testing.T) {
	if anchorRelocationGuardEnabled(nil) {
		t.Error("nil config provider must read as disabled (default-OFF)")
	}
	on := guardTestConfigProvider{cfg: &config.Config{Workflow: config.WorkflowConfig{
		AnchorRelocationGuard: config.AnchorRelocationGuardConfig{Enabled: true},
	}}}
	if !anchorRelocationGuardEnabled(on) {
		t.Error("enabled config must read as enabled")
	}
	off := guardTestConfigProvider{cfg: &config.Config{}}
	if anchorRelocationGuardEnabled(off) {
		t.Error("unset config must read as disabled (default-OFF)")
	}
}
