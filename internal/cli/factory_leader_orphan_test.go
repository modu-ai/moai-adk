package cli

// factory_leader_orphan_test.go — SPEC-FACTORY-LEADER-ADOPT-CORE-001: the
// deliberate --factory-run adopt through the adoption write (AC-FLAC-003,
// AC-FLAC-004) and the leader-start orphan refusal gate (AC-FLAC-002,
// AC-FLAC-005..009). The entry is driven through the real cc command with
// the launch seam capturing the factory environment at launch time; the
// factory database is a per-test sandbox, never the developer's own.

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// seedAdoptEntryRun records one active run named runID with the requested
// backend and owner stamp in a sandbox git root, seeds two cards and the
// lane capacity, closes the handle, and returns the root. live=false seeds a
// stamp that can never classify live (pid beyond pid_max, a fingerprint that
// never ran); live=true seeds this test process's own identity.
func seedAdoptEntryRun(t *testing.T, runID, backend string, live bool, laneCapacity int) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	run := homestate.FactoryRun{RunID: runID, Backend: backend, ManifestJSON: "{}", LaneCapacity: laneCapacity}
	if live {
		run.LeadPID = os.Getpid()
		run.LeadProcessStart = homestate.CurrentProcessFingerprint()
	} else {
		run.LeadPID = 99999999
		run.LeadProcessStart = "gone-before-explicit"
	}
	if err := db.RecordRun(context.Background(), run); err != nil {
		t.Fatalf("record run %s: %v", runID, err)
	}
	for _, card := range []struct{ id, state string }{{"c-1", "run"}, {"c-2", "blocked"}} {
		if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES(?,?,?,?)`,
			runID, card.id, card.state, "2026-10-14T00:00:00Z"); err != nil {
			t.Fatalf("insert card %s: %v", card.id, err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close seeding handle: %v", err)
	}
	return root
}

// installAdoptEntryEnv isolates the launch environment and points it at the
// sandbox root, then installs the launch seam. Returns the capture. The root
// keeps its project-local DB: no MOAI_HOME override here, because an absolute
// MOAI_HOME would move the factory database out of the seeded root
// (homestate's temp-base exception) and the entry would read an empty one.
func installAdoptEntryEnv(t *testing.T, root string) *factoryLaunchCapture {
	t.Helper()
	clearFactoryTestEnv(t)
	for _, key := range codexLaneLaunchEnvKeys {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)
	return installFactoryLaunchSeam(t)
}

// adoptCardsSnapshot renders the run's card rows as one string so a
// byte-identical before/after pair proves the adopt touched no card.
func adoptCardsSnapshot(t *testing.T, root, runID string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for cards snapshot: %v", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT card_id,owner_label,state,version,evidence_path,updated_at,stage FROM cards WHERE run_id=? ORDER BY card_id`, runID)
	if err != nil {
		t.Fatalf("cards snapshot: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var b strings.Builder
	for rows.Next() {
		var cardID, owner, state, updated, stage string
		var version int
		var evidence string
		if err := rows.Scan(&cardID, &owner, &state, &version, &evidence, &updated, &stage); err != nil {
			t.Fatalf("cards snapshot scan: %v", err)
		}
		b.WriteString(cardID + "|" + owner + "|" + state + "|" + stage + "|" + updated + "\n")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("cards snapshot rows: %v", err)
	}
	return b.String()
}

// adoptEventCount counts one event kind for a run.
func adoptEventCount(t *testing.T, root, runID, kind string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for event count: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM events WHERE run_id=? AND kind=?`, runID, kind).Scan(&n); err != nil {
		t.Fatalf("count events %s: %v", kind, err)
	}
	return n
}

// adoptRunsCount counts every run row in the sandbox.
func adoptRunsCount(t *testing.T, root string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for runs count: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM runs`).Scan(&n); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	return n
}

// adoptEventPayload returns the payload of the run's single event of one kind.
func adoptEventPayload(t *testing.T, root, runID, kind string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for event payload: %v", err)
	}
	defer func() { _ = db.Close() }()
	var payload string
	if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE run_id=? AND kind=?`, runID, kind).Scan(&payload); err != nil {
		t.Fatalf("read event %s payload: %v", kind, err)
	}
	return payload
}

// adoptRunRow reads one run row's observable state.
func adoptRunRow(t *testing.T, root, runID string) (status, backend string, pid int, start string, laneCapacity int) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for row read: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.DB.QueryRow(`SELECT status, lead_backend, lead_pid, lead_process_start, lane_capacity FROM runs WHERE run_id=?`, runID).
		Scan(&status, &backend, &pid, &start, &laneCapacity); err != nil {
		t.Fatalf("read run row %s: %v", runID, err)
	}
	return status, backend, pid, start, laneCapacity
}

// runAdoptEntry drives one `cc -f --factory-run <id>` entry through the real
// command with fresh output buffers, returning the error.
func runAdoptEntry(t *testing.T, args ...string) error {
	t.Helper()
	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)
	return runCC(ccCmd, args)
}

// assertAdoptedRunInvariants is the shared post-adopt assertion set of
// AC-FLAC-003: the address republished, the launcher's stamp and backend on
// the row, one run.resumed with the leader-start basis, no second run.started,
// no second run row, the lane capacity preserved, and the cards untouched.
func assertAdoptedRunInvariants(t *testing.T, c *factoryLaunchCapture, root, runID, cardsBefore string, wantBackend string) {
	t.Helper()
	if c.runID != runID {
		t.Fatalf("launch env %s = %q, want the adopted run id %q", config.EnvFactoryRunID, c.runID, runID)
	}
	if want := factory.FactoryLeaderSocketPath(runID); c.addr != want {
		t.Fatalf("launch env %s = %q, want the adopted run's address %q", config.EnvFactoryLeadAddr, c.addr, want)
	}
	status, backend, pid, start, laneCapacity := adoptRunRow(t, root, runID)
	if status != "active" {
		t.Fatalf("run %s status = %q, want active", runID, status)
	}
	if backend != wantBackend {
		t.Fatalf("run %s lead_backend = %q, want %q (the adopting launcher's backend, written by the adoption write)", runID, backend, wantBackend)
	}
	if pid != os.Getpid() || start != homestate.CurrentProcessFingerprint() {
		t.Fatalf("run %s stamp = (%d, %q), want this process (%d, %q)", runID, pid, start, os.Getpid(), homestate.CurrentProcessFingerprint())
	}
	if laneCapacity != 7 {
		t.Fatalf("run %s lane_capacity = %d, want 7 (the adoption write never touches it)", runID, laneCapacity)
	}
	if n := adoptEventCount(t, root, runID, "run.resumed"); n != 1 {
		t.Fatalf("run.resumed events = %d, want exactly 1", n)
	}
	payload := adoptEventPayload(t, root, runID, "run.resumed")
	if !strings.Contains(payload, "leader-start-adoption") {
		t.Fatalf("run.resumed payload %q does not name the leader-start basis", payload)
	}
	if n := adoptEventCount(t, root, runID, "run.started"); n != 1 {
		t.Fatalf("run.started events = %d, want exactly 1 (the seeding one only — the adopt must not append one)", n)
	}
	if n := adoptRunsCount(t, root); n != 1 {
		t.Fatalf("runs rows = %d, want 1 (the adopt must not mint a second run)", n)
	}
	if after := adoptCardsSnapshot(t, root, runID); after != cardsBefore {
		t.Fatalf("cards changed across the adopt:\nbefore:\n%s\nafter:\n%s", cardsBefore, after)
	}
}

// AC-FLAC-003 — the deliberate `--factory-run` takeover goes through the same
// adoption write the resume adoption uses: compare-and-set on the classified
// owner stamp, one run.resumed event with the leader-start basis, the adopted
// id's address republished, and the adopting launcher's backend written by
// the adoption write's own transaction (the seeded backend is the OTHER
// launcher's, so the backend equality cannot hold without the write).
func TestExplicitFactoryRunAdoptTakesOverAtomically(t *testing.T) {
	t.Run("dead-stamp cross-backend", func(t *testing.T) {
		root := seedAdoptEntryRun(t, "adopt-run", "glm", false, 7)
		cardsBefore := adoptCardsSnapshot(t, root, "adopt-run")
		c := installAdoptEntryEnv(t, root)
		if err := runAdoptEntry(t, "-f", "--factory-run", "adopt-run"); err != nil {
			t.Fatalf("explicit adopt refused: %v", err)
		}
		if c.args == nil {
			t.Fatalf("the launch seam was never called — the adopt must proceed to the session")
		}
		assertAdoptedRunInvariants(t, c, root, "adopt-run", cardsBefore, "claude")
	})

	// The proceed direction of the classifyNamedRun input seam: the same
	// launch over a run whose owner classifies LIVE takes it over through the
	// same write with the same invariants.
	t.Run("live-owner", func(t *testing.T) {
		root := seedAdoptEntryRun(t, "adopt-live", "claude", true, 7)
		cardsBefore := adoptCardsSnapshot(t, root, "adopt-live")
		c := installAdoptEntryEnv(t, root)
		if err := runAdoptEntry(t, "-f", "--factory-run", "adopt-live"); err != nil {
			t.Fatalf("explicit adopt of a live-owner run refused: %v", err)
		}
		if c.args == nil {
			t.Fatalf("the launch seam was never called")
		}
		assertAdoptedRunInvariants(t, c, root, "adopt-live", cardsBefore, "claude")
	})
}

// AC-FLAC-004 — every refused adopt aborts the launch with a non-zero exit
// and the refusal text, minting nothing and modifying nothing: (a) a codex-led
// run, (b) a run holding a live legacy peer record, (c) the owner changed
// between classification and write so the compare-and-set misses.
func TestExplicitFactoryRunAdoptRefusals(t *testing.T) {
	t.Run("codex-led", func(t *testing.T) {
		root := seedAdoptEntryRun(t, "adopt-cx", "codex", false, 1)
		statusBefore, backendBefore, pidBefore, startBefore, capBefore := adoptRunRow(t, root, "adopt-cx")
		eventsBefore := adoptEventCount(t, root, "adopt-cx", "run.started") + adoptEventCount(t, root, "adopt-cx", "run.resumed")
		c := installAdoptEntryEnv(t, root)
		err := runAdoptEntry(t, "-f", "--factory-run", "adopt-cx")
		if err == nil {
			t.Fatal("the codex-led run was adopted; want a refusal")
		}
		for _, want := range []string{"adopt-cx", "codex", "moai factory runs --retire"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal %q lacks %q", err.Error(), want)
			}
		}
		if c.args != nil {
			t.Errorf("the launch seam was called on a refused adopt")
		}
		if status, backend, pid, start, cap := adoptRunRow(t, root, "adopt-cx"); status != statusBefore || backend != backendBefore || pid != pidBefore || start != startBefore || cap != capBefore {
			t.Errorf("run row changed across the refused adopt: (%s,%s,%d,%s,%d) -> (%s,%s,%d,%s,%d)", statusBefore, backendBefore, pidBefore, startBefore, capBefore, status, backend, pid, start, cap)
		}
		if got := adoptEventCount(t, root, "adopt-cx", "run.started") + adoptEventCount(t, root, "adopt-cx", "run.resumed"); got != eventsBefore {
			t.Errorf("events changed across the refused adopt: %d -> %d", eventsBefore, got)
		}
		if n := adoptRunsCount(t, root); n != 1 {
			t.Errorf("runs rows = %d, want 1 (a refusal must not mint)", n)
		}
	})

	t.Run("live-legacy-peer", func(t *testing.T) {
		root := seedLegacyPeerRun(t, "adopt-peer", "worker", "worker-2", true)
		statusBefore, backendBefore, pidBefore, startBefore, capBefore := adoptRunRow(t, root, "adopt-peer")
		c := installAdoptEntryEnv(t, root)
		err := runAdoptEntry(t, "-f", "--factory-run", "adopt-peer")
		if err == nil {
			t.Fatal("the run holding a live legacy peer was adopted; want a refusal")
		}
		for _, want := range []string{"worker-2", "adopt-peer", "moai factory runs --retire"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal %q lacks %q", err.Error(), want)
			}
		}
		if c.args != nil {
			t.Errorf("the launch seam was called on a refused adopt")
		}
		if status, backend, pid, start, cap := adoptRunRow(t, root, "adopt-peer"); status != statusBefore || backend != backendBefore || pid != pidBefore || start != startBefore || cap != capBefore {
			t.Errorf("run row changed across the refused adopt")
		}
	})

	t.Run("owner-change-race", func(t *testing.T) {
		root := seedAdoptEntryRun(t, "adopt-race", "claude", false, 1)
		// The input seam's abort direction: the override classifies the run
		// through the production closure, then a competing writer restamps the
		// row before the returned classification is written — the
		// classify-to-write window the compare-and-set must close.
		prod := classifyNamedRun
		classifyNamedRun = func(ctx context.Context, root, runID string) (homestate.RunOwner, error) {
			owner, err := prod(ctx, root, runID)
			if err != nil {
				return owner, err
			}
			db, openErr := homestate.OpenFactory(root)
			if openErr != nil {
				return owner, openErr
			}
			defer func() { _ = db.Close() }()
			if _, err := db.DB.Exec(`UPDATE runs SET lead_pid=4242, lead_process_start='competing-writer' WHERE run_id=?`, runID); err != nil {
				return owner, err
			}
			return owner, nil
		}
		t.Cleanup(func() { classifyNamedRun = prod })

		eventsBefore := adoptEventCount(t, root, "adopt-race", "run.started") + adoptEventCount(t, root, "adopt-race", "run.resumed")
		c := installAdoptEntryEnv(t, root)
		err := runAdoptEntry(t, "-f", "--factory-run", "adopt-race")
		if err == nil {
			t.Fatal("the owner-changed adopt succeeded; want the compare-and-set refusal")
		}
		if !strings.Contains(err.Error(), "no longer carries the classified owner stamp") {
			t.Errorf("refusal %q does not name the compare-and-set miss", err.Error())
		}
		if c.args != nil {
			t.Errorf("the launch seam was called on a refused adopt")
		}
		// The competing writer's stamp survives: the write refused, so the row
		// keeps the value the competitor stamped between classification and
		// write, and the seeded backend is unchanged.
		if status, backend, pid, start, cap := adoptRunRow(t, root, "adopt-race"); status != "active" || backend != "claude" || pid != 4242 || start != "competing-writer" || cap != 1 {
			t.Errorf("row after refused adopt = (%s, %s, %d, %s, %d), want (active, claude, 4242, competing-writer, 1)", status, backend, pid, start, cap)
		}
		if got := adoptEventCount(t, root, "adopt-race", "run.started") + adoptEventCount(t, root, "adopt-race", "run.resumed"); got != eventsBefore {
			t.Errorf("events changed across the refused adopt: %d -> %d", eventsBefore, got)
		}
		if n := adoptRunsCount(t, root); n != 1 {
			t.Errorf("runs rows = %d, want 1", n)
		}
	})
}
