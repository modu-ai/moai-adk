// factory_lane_backend_test.go — SPEC-QUOTA-AWARE-SCHEDULING-001 M4 AC test
// (card t1347): the lane's backend recorded at claim. AC-QAS-023 (cli half):
// each of the four launcher claim sites passes its launcher-local backend to the
// claim, so the registry row carries claude, glm, gpt, and gpt (the Codex
// factory-entry token normalized), and the Codex pid re-stamp leaves the
// backend alone. The claim engine's own half is
// internal/kanban TestQAS_AC023_ClaimRecordsBackend.
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (the factory test family's convention); the engine
// launches are the launchers' existing seams (unifiedLaunchFunc, codexLookPath,
// codexDirectLaunchFn) and the liveness probe is the real factoryProcessAlive.
package cli

import (
	"os"
	"os/exec"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// qasLaneRowFields is one workers row, read column by column.
type qasLaneRowFields struct {
	pid          int
	backend      string
	sessionID    string
	runID        string
	registeredAt string
}

// qasLaneRow reads the workers row for label under root.
func qasLaneRow(t *testing.T, root, label string) qasLaneRowFields {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var r qasLaneRowFields
	if err := db.DB.QueryRow(`SELECT pid, backend, session_id, run_id, registered_at FROM workers WHERE label=?`, label).
		Scan(&r.pid, &r.backend, &r.sessionID, &r.runID, &r.registeredAt); err != nil {
		t.Fatalf("read the %s row: %v", label, err)
	}
	return r
}

// qasOnlyLaneLabel returns the label of the single registered lane under root.
func qasOnlyLaneLabel(t *testing.T, root string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT label FROM workers`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var labels []string
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			t.Fatal(err)
		}
		labels = append(labels, label)
	}
	if len(labels) != 1 {
		t.Fatalf("registered lanes = %v, want exactly one", labels)
	}
	return labels[0]
}

// AC-QAS-023 (cli half) — the four launcher claim sites pass their backend.
func TestQAS_AC023b_LaunchersPassTheirBackendToTheClaim(t *testing.T) {
	for _, tc := range []struct {
		name    string
		entry   func([]string) error
		backend string
	}{
		{"cc_claude", sdCCEntry, kanban.BackendClaude},
		{"glm_glm", sdGLMEntry, kanban.BackendGLM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, _ := fcFixture(t)
			sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
			t.Chdir(root)
			t.Setenv(config.EnvClaudeProjectDir, root)
			label := sdNextFreeLaneLabel(t, root)

			sdDriveLaneLaunch(t, root, tc.entry)

			if got := qasLaneRow(t, root, label).backend; got != tc.backend {
				t.Errorf("%s lane %s recorded backend %q, want %q", tc.name, label, got, tc.backend)
			}
		})
	}

	t.Run("codex_loop_gpt", func(t *testing.T) {
		root, _ := fcFixture(t)
		sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
		t.Chdir(root)
		sdScrubLauncherEnv(t)
		label := sdNextFreeLaneLabel(t, root)
		prevLook, prevDirect := codexLookPath, codexDirectLaunchFn
		codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
		codexDirectLaunchFn = func(*exec.Cmd) error { return nil }
		t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })

		if _, _, err := runCodexCmd(t, "-f", "lane"); err != nil {
			t.Fatalf("codex lane loop: %v", err)
		}

		if got := qasLaneRow(t, root, label).backend; got != kanban.BackendGPT {
			t.Errorf("codex loop lane %s recorded backend %q, want %q", label, got, kanban.BackendGPT)
		}
	})

	t.Run("codex_factory_entry_gpt", func(t *testing.T) {
		root := codexLedRun(t, "entry-run", BackendCodex)
		entry := factoryFlagParse{Enabled: true, LaneRole: true, RunID: "entry-run"}
		restore, err := enterCodexFactory(root, entry, nil)
		if err != nil {
			t.Fatalf("enterCodexFactory lane: %v", err)
		}
		restore()

		label := qasOnlyLaneLabel(t, root)
		if got := qasLaneRow(t, root, label).backend; got != kanban.BackendGPT {
			t.Errorf("codex factory-entry lane %s recorded backend %q, want %q (the %q token normalized)", label, got, kanban.BackendGPT, BackendCodex)
		}
	})

	t.Run("codex_pid_update_preserves_backend", func(t *testing.T) {
		root := codexLedRun(t, "pid-run", BackendCodex)
		t.Setenv(config.EnvMoaiKanbanID, "pid-run")
		label, err := resolveFactoryLaneName(root, "lane-1", kanban.BackendGPT, false, nil)
		if err != nil {
			t.Fatal(err)
		}
		before := qasLaneRow(t, root, label)
		if before.backend != kanban.BackendGPT {
			t.Fatalf("control: the claimed row recorded backend %q, want %q", before.backend, kanban.BackendGPT)
		}
		env := []string{
			config.EnvMoaiKanbanID + "=pid-run",
			config.EnvMoaiFactoryWorker + "=" + label,
		}
		childPID := os.Getpid() + 1000
		if err := stampCodexLaneClaim(root, env, childPID); err != nil {
			t.Fatal(err)
		}

		after := qasLaneRow(t, root, label)
		if after.pid != childPID {
			t.Fatalf("control: the pid re-stamp did not land (pid = %d, want %d)", after.pid, childPID)
		}
		before.pid = after.pid
		if after != before {
			t.Errorf("the pid re-stamp changed another column: before %+v, after %+v", before, after)
		}
	})
}
