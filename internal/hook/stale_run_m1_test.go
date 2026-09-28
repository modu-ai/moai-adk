package hook

// stale_run_m1_test.go — SPEC-ROLE-NAMING-CODE-001 M1: SessionStart refuses
// to adopt a legacy-vocabulary session. A launch label or an existing session
// record carrying a legacy role value produces the stale-run notice instead
// of a leader or lane notice, and the session-record writer leaves the record
// absent (or byte-identical) rather than re-deriving a new-vocabulary role
// (REQ-RNC-025, AC-RNC-022 hook clause, AC-RNC-025).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

func newStaleRunRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestStaleRunNoticeLegacyLeaderSpelling(t *testing.T) { // AC-RNC-025 (a) + debt P4
	root := newStaleRunRoot(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanLeadName, "lead")

	input := &HookInput{SessionID: "stale-lead-session", Source: "startup", ProjectDir: root, CWD: root}
	writeKanbanSessionRecord(input)

	if _, err := os.Stat(kanban.RecordPath(root, "stale-lead-session")); !os.IsNotExist(err) {
		t.Errorf("session record exists after legacy-label SessionStart, want no file created")
	}
	notice := staleRunNoticeFor(root, "stale-lead-session", "en")
	if notice == "" {
		t.Fatalf("staleRunNoticeFor = \"\", want a notice")
	}
	if !strings.Contains(notice, "lead") {
		t.Errorf("notice %q does not name the legacy label lead", notice)
	}
	if strings.Contains(notice, "leader-") {
		t.Errorf("notice %q must not advertise a leader label", notice)
	}
}

func TestStaleRunNoticeLegacySessionRecord(t *testing.T) { // AC-RNC-025 (b)
	root := newStaleRunRoot(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanLeadName, "leader")

	// A pre-rename binary wrote this record with role "lead" — written as raw
	// JSON because the current-vocabulary WithRole setter correctly refuses
	// the legacy value.
	path := kanban.RecordPath(root, "legacy-record-session")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	legacyRecord := []byte(`{"session_id":"legacy-record-session","spec_id":"","role":"lead","backend":"claude","entered_at":"2026-09-01T00:00:00Z","deepscan_dir":"","verify_reentries":0}` + "\n")
	if err := os.WriteFile(path, legacyRecord, 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	input := &HookInput{SessionID: "legacy-record-session", Source: "startup", ProjectDir: root, CWD: root}
	writeKanbanSessionRecord(input)

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("session record was rewritten by SessionStart (writer re-derived the role)")
	}
	notice := staleRunNoticeFor(root, "legacy-record-session", "en")
	if notice == "" {
		t.Fatalf("staleRunNoticeFor = \"\", want a notice naming the legacy record role")
	}
	if !strings.Contains(notice, "lead") {
		t.Errorf("notice %q does not name the legacy role lead", notice)
	}
}

func TestStaleRunNoticeFactoryLegacyLabel(t *testing.T) { // AC-RNC-022 hook clause
	root := newStaleRunRoot(t)
	t.Setenv(config.EnvMoaiKanbanID, "runR")
	t.Setenv(config.EnvMoaiFactoryWorkers, "4")
	t.Setenv(config.EnvMoaiFactoryWorker, "worker-2")

	input := &HookInput{SessionID: "stale-lane-session", Source: "startup", ProjectDir: root, CWD: root}
	writeKanbanSessionRecord(input)
	if _, err := os.Stat(kanban.RecordPath(root, "stale-lane-session")); !os.IsNotExist(err) {
		t.Errorf("session record exists after legacy lane label, want none")
	}

	notice := staleRunNoticeFor(root, "stale-lane-session", "en")
	for _, want := range []string{"worker-2", "runR", "moai factory runs --retire runR"} {
		if !strings.Contains(notice, want) {
			t.Errorf("stale-run notice %q missing %q", notice, want)
		}
	}

	// The bootstrap notice surface carries the stale-run message instead of a
	// leader or lane notice.
	factory := factoryBootstrapNotice(root, "", "en")
	if !strings.Contains(factory, "worker-2") || !strings.Contains(factory, "runR") {
		t.Errorf("factoryBootstrapNotice = %q, want the stale-run message naming worker-2 and runR", factory)
	}
}

func TestKanbanRoleFromEnvLegacyLabelsNotRecognized(t *testing.T) { // REQ-RNC-009
	t.Setenv(config.EnvMoaiFactoryWorker, "worker-2")
	t.Setenv(config.EnvMoaiFactoryWorkers, "4")
	role, lane, ok := kanbanRoleFromEnv()
	if ok {
		t.Errorf("kanbanRoleFromEnv(worker-2) = (%q, %d, true), want ok=false — legacy is detection only", role, lane)
	}
}

func TestKanbanRoleFromEnvNewVocabulary(t *testing.T) {
	t.Setenv(config.EnvMoaiFactoryWorker, "lane-2")
	t.Setenv(config.EnvMoaiFactoryWorkers, "4")
	role, lane, ok := kanbanRoleFromEnv()
	if !ok || role != kanban.RoleLane || lane != 2 {
		t.Errorf("kanbanRoleFromEnv(lane-2) = (%q, %d, %v), want (lane, 2, true)", role, lane, ok)
	}
}

func TestFactoryHookPeerRefusesLegacyLabel(t *testing.T) {
	// The factory message hook must not register a peer under a legacy label;
	// it reports the stale-run condition instead.
	if got := legacyFactoryHookNotice("worker-2", "runR", "en"); got == "" || !strings.Contains(got, "worker-2") {
		t.Errorf("legacyFactoryHookNotice(worker-2) = %q, want a stale-run notice naming worker-2", got)
	}
	if got := legacyFactoryHookNotice("lane-2", "runR", "en"); got != "" {
		t.Errorf("legacyFactoryHookNotice(lane-2) = %q, want \"\" (current vocabulary)", got)
	}
}
