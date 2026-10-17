package cli

// `factory assign --reassign` (card t1696, defect D2): the leader path that
// moves an assigned card whose owner lane is gone to another registered lane.
// A lane session is refused; the reassignment changes the owner only.

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// reassignRun runs `moai factory assign` in-process, mirroring qasRunNext.
func reassignRun(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := newFactoryCommand()
	var silence func(c *cobra.Command)
	silence = func(c *cobra.Command) {
		c.SilenceUsage, c.SilenceErrors = true, true
		for _, sub := range c.Commands() {
			silence(sub)
		}
	}
	silence(cmd)
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(append([]string{"assign"}, args...))
	err := cmd.Execute()
	return out.String(), errBuf.String(), err
}

func TestFactoryAssignReassignHappyPath(t *testing.T) {
	root, _ := nmBase(t, factory.BacklogStatePicked)
	if _, _, err := reassignRun(t, "t1", "--run", fcRun, "--to", "lane-9"); err != nil {
		t.Fatalf("seed assign: %v", err)
	}
	// Register the takeover lane (a joined lane session registers itself).
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open the record: %v", err)
	}
	stamp := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at) VALUES('lane-15',0,?,?)`, stamp, stamp); err != nil {
		t.Fatalf("register lane-15: %v", err)
	}
	_ = db.Close()

	out, _, err := reassignRun(t, "t1", "--run", fcRun, "--reassign", "--to", "lane-15")
	if err != nil {
		t.Fatalf("reassign: %v", err)
	}
	if !strings.Contains(out, "reassigned from lane-9") {
		t.Fatalf("output %q does not name the previous owner", out)
	}
	c := fcCard(t, root, "t1")
	if c.State != homestate.CardAssigned || c.OwnerLabel != "lane-15" {
		t.Fatalf("card = %s/%s, want assigned/lane-15", c.State, c.OwnerLabel)
	}
}

func TestFactoryAssignReassignRefusals(t *testing.T) {
	t.Run("a lane session is refused", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStatePicked)
		_, _, _ = reassignRun(t, "t1", "--run", fcRun, "--to", "lane-9")
		nmLaneEnv(t, "lane-1", "")
		_, stderr, err := reassignRun(t, "t1", "--run", fcRun, "--reassign", "--to", "lane-15")
		if err == nil || !strings.Contains(stderr+err.Error(), "lane boundary") {
			t.Fatalf("err=%v stderr=%q, want the lane-boundary refusal", err, stderr)
		}
		_ = root
	})
	t.Run("a missing --to is refused", func(t *testing.T) {
		_, _ = nmBase(t, factory.BacklogStatePicked)
		_, stderr, err := reassignRun(t, "t1", "--run", fcRun, "--reassign")
		if err == nil || !strings.Contains(stderr+err.Error(), "--reassign needs --to") {
			t.Fatalf("err=%v stderr=%q, want the --to refusal", err, stderr)
		}
	})
	t.Run("other fields are refused", func(t *testing.T) {
		_, _ = nmBase(t, factory.BacklogStatePicked)
		_, stderr, err := reassignRun(t, "t1", "--run", fcRun, "--reassign", "--to", "lane-15", "--spec", "SPEC-X")
		if err == nil || !strings.Contains(stderr+err.Error(), "changes no other field") {
			t.Fatalf("err=%v stderr=%q, want the field refusal", err, stderr)
		}
	})
	t.Run("an unregistered target is refused", func(t *testing.T) {
		root, _ := nmBase(t, factory.BacklogStatePicked)
		_, _, _ = reassignRun(t, "t1", "--run", fcRun, "--to", "lane-9")
		_, stderr, err := reassignRun(t, "t1", "--run", fcRun, "--reassign", "--to", "lane-99")
		if err == nil || !strings.Contains(stderr+err.Error(), "not a registered lane") {
			t.Fatalf("err=%v stderr=%q, want the registration refusal", err, stderr)
		}
		_ = root
	})
	t.Run("a stale run refuses", func(t *testing.T) {
		// Sync-audit F6 (card t1696): the card's dispatch mirror names its
		// current run; a reassign resolved against another run must refuse
		// instead of dragging the mirror back and admitting a second lease.
		_, store := nmBase(t, factory.BacklogStatePicked)
		_, _, _ = reassignRun(t, "t1", "--run", fcRun, "--to", "lane-9")
		mdb, err := sql.Open("sqlite", "file:"+filepath.ToSlash(store.EnginePath()))
		if err != nil {
			t.Fatal(err)
		}
		// The fixture queue may predate the runtime migration: the full
		// runtime schema the migration writes (todoRuntimeDDL's shape) is
		// created here when absent, so LoadPure reads the mirror instead of
		// reading a half-stamped schema as corrupt.
		for _, stmt := range []string{
			`CREATE TABLE IF NOT EXISTS todo_runtime_runs (
 run_id TEXT PRIMARY KEY, backend TEXT NOT NULL, manifest_json TEXT NOT NULL
)`,
			`CREATE TABLE IF NOT EXISTS todo_runtime_assignments (
 run_id TEXT NOT NULL, card_id TEXT NOT NULL, owner_label TEXT NOT NULL,
 reported_state TEXT NOT NULL, event_kind TEXT NOT NULL, provenance_json TEXT NOT NULL,
 PRIMARY KEY(run_id, card_id)
)`,
			`CREATE TABLE IF NOT EXISTS todo_dispatch_current (
 card_id TEXT PRIMARY KEY, run_id TEXT NOT NULL, owner_label TEXT NOT NULL
)`,
			`INSERT INTO meta(key,value) VALUES('runtime_schema_version','1') ON CONFLICT(key) DO NOTHING`,
		} {
			if _, err := mdb.Exec(stmt); err != nil {
				t.Fatalf("ensure the runtime schema: %v", err)
			}
		}
		if _, err := mdb.Exec(`INSERT INTO todo_dispatch_current(card_id,run_id,owner_label) VALUES('t1','run-stale','lane-9')`); err != nil {
			t.Fatalf("seed the stale mirror: %v", err)
		}
		_ = mdb.Close()
		_, stderr, err := reassignRun(t, "t1", "--run", fcRun, "--reassign", "--to", "lane-15")
		if err == nil || !strings.Contains(stderr+err.Error(), "current dispatch is run run-stale") {
			t.Fatalf("err=%v stderr=%q, want the stale-run refusal", err, stderr)
		}
	})
}
