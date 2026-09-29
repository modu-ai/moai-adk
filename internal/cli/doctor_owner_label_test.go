// doctor_owner_label_test.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// AC-TSP-051: the doctor owner-label drift check's verdicts.
package cli

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"

	_ "modernc.org/sqlite"
)

func TestDoctorOwnerLabelDrift(t *testing.T) {
	t.Run("no legacy rows is OK", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvHome, t.TempDir())
		// An absent queue database reads as no assignments at all.
		check := checkOwnerLabelDrift(root, false)
		if check.Status != uikit.CheckOK {
			t.Errorf("status = %v, want OK (message %q)", check.Status, check.Message)
		}
		if check.Name != ownerLabelDriftCheckName {
			t.Errorf("name = %q, want %q", check.Name, ownerLabelDriftCheckName)
		}
	})

	t.Run("legacy rows present is WARN with counts", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv(config.EnvHome, t.TempDir())
		// The runtime tables come into being through the store's own write
		// path (whose normalization is why the legacy rows below are then
		// hand-written raw: after REQ-TSP-052 no code path writes them).
		store := kanban.NewBacklogStore(kanban.BacklogPathForRoot(root))
		if _, _, err := store.Add("drift fixture card"); err != nil {
			t.Fatalf("seed card: %v", err)
		}
		if err := kanban.RecordFactoryCardState(root, "run-canonical", "t1", "lane-1", "", "assigned", "card.assigned"); err != nil {
			t.Fatalf("seed canonical assignment: %v", err)
		}
		db, err := sql.Open("sqlite", store.EnginePath())
		if err != nil {
			t.Fatal(err)
		}
		stmt := `INSERT INTO todo_runtime_assignments(run_id,card_id,owner_label,reported_state,event_kind,provenance_json)
		         VALUES(?,?,?,?,?,?)`
		for i, label := range []string{"lead", "lead", "lead", "worker-67"} {
			if _, err := db.Exec(stmt, "legacy-"+label+"-"+string(rune('a'+i)), "t1", label, "assigned", "card.event", "{}"); err != nil {
				t.Fatalf("seed legacy row %d: %v", i, err)
			}
		}
		_ = db.Close()

		check := checkOwnerLabelDrift(root, false)

		if check.Status != uikit.CheckWarn {
			t.Errorf("status = %v, want WARN (message %q)", check.Status, check.Message)
		}
		joined := check.Message + "\n" + check.Detail
		if !strings.Contains(joined, "4 owner_label row(s)") {
			t.Errorf("message does not carry the drift total:\n%s\n%s", check.Message, check.Detail)
		}
		if !strings.Contains(joined, "lead") || !strings.Contains(joined, "worker-67") {
			t.Errorf("message does not name the legacy labels:\n%s\n%s", check.Message, check.Detail)
		}
	})
}
