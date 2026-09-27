package cli

// role_naming_m3_doctor_test.go — SPEC-ROLE-NAMING-CODE-001 M3, AC-RNC-013
// doctor clause: the doctor factory section recognizes a persisted leader
// role of `leader` (present, label `leader`) and renders the literal
// "legacy run: relaunch required" for a persisted `lead`.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

func seedKanbanRecord(t *testing.T, root, sessionID, role string) {
	t.Helper()
	path := kanban.RecordPath(root, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"session_id":"` + sessionID + `","spec_id":"","role":"` + role + `","backend":"claude","entered_at":"2026-09-01T00:00:00Z","deepscan_dir":"","verify_reentries":0}` + "\n"
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestDoctorFactoryRunCheckRecognizesLeader(t *testing.T) {
	root := t.TempDir()
	seedKanbanRecord(t, root, "lead-session", "leader")

	check := checkFactoryRun(root, false)
	if check.Status != uikit.CheckOK {
		t.Errorf("status = %q, want ok; check = %+v", check.Status, check)
	}
	if !strings.Contains(check.Message+check.Detail, "leader") {
		t.Errorf("check %+v does not display the leader label", check)
	}
}

func TestDoctorFactoryRunCheckRendersLegacyRelaunchLiteral(t *testing.T) {
	root := t.TempDir()
	seedKanbanRecord(t, root, "legacy-session", "lead")

	check := checkFactoryRun(root, false)
	if check.Status != uikit.CheckFail {
		t.Errorf("status = %q, want fail for a legacy run; check = %+v", check.Status, check)
	}
	if !strings.Contains(check.Message, "legacy run: relaunch required") {
		t.Errorf("message %q does not contain the literal %q", check.Message, "legacy run: relaunch required")
	}
}

func TestDoctorFactoryRunCheckWithoutRecordsIsInformational(t *testing.T) {
	root := t.TempDir()

	check := checkFactoryRun(root, false)
	if check.Status != uikit.CheckInfo {
		t.Errorf("status = %q, want info with no records; check = %+v", check.Status, check)
	}
}
