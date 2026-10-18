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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
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
	// The cards carry non-default version and evidence_path values, so a
	// mutant that writes cards diverges on the discriminating fields the
	// cards snapshot compares (sync-audit F2).
	for _, card := range []struct{ id, state, evidence string }{
		{"c-1", "run", ".moai/reports/" + runID + "/card-c-1.md"},
		{"c-2", "blocked", ""},
	} {
		if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,evidence_path,updated_at) VALUES(?,?,?,?,?,?)`,
			runID, card.id, card.state, 3, card.evidence, "2026-10-14T00:00:00Z"); err != nil {
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
// byte-identical before/after pair proves the adopt touched no card. Every
// selected column is compared — version and evidence_path included (the
// sync-audit F2 finding: a snapshot that drops selected columns lets a
// card-writing mutant pass the byte-identical assertion).
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
		b.WriteString(cardID + "|" + owner + "|" + state + "|" + strconv.Itoa(version) + "|" + evidence + "|" + stage + "|" + updated + "\n")
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

// AC-FLAC-004 (sync-audit F1) — the codex-led and live-legacy-peer refusals
// are re-applied INSIDE the classify-to-write window of the deliberate
// adopt. enterSelectedFactoryRun applies both before the adopt (the T1
// checks), but the window is wide: a backend flip to codex, or a legacy-peer
// registration, can land between the fresh classification and the adoption
// write — and the stamp the compare-and-set checks did not change, so
// without a re-check the takeover would adopt a run its guards refuse.
func TestExplicitFactoryRunAdoptReappliesWindowRefusals(t *testing.T) {
	t.Run("backend-flip-in-window", func(t *testing.T) {
		root := seedAdoptEntryRun(t, "adopt-flip", "claude", false, 1)
		cardsBefore := adoptCardsSnapshot(t, root, "adopt-flip")
		prod := classifyNamedRun
		classifyNamedRun = func(ctx context.Context, root, runID string) (homestate.RunOwner, error) {
			owner, err := prod(ctx, root, runID)
			if err != nil {
				return owner, err
			}
			// The competing writer flips the run's backend to codex between
			// classification and write; the stamp is untouched, so a bare
			// compare-and-set would still match.
			db, openErr := homestate.OpenFactory(root)
			if openErr != nil {
				return owner, openErr
			}
			defer func() { _ = db.Close() }()
			if _, err := db.DB.Exec(`UPDATE runs SET lead_backend='codex' WHERE run_id=?`, runID); err != nil {
				return owner, err
			}
			return owner, nil
		}
		t.Cleanup(func() { classifyNamedRun = prod })

		c := installAdoptEntryEnv(t, root)
		err := runAdoptEntry(t, "-f", "--factory-run", "adopt-flip")
		if err == nil {
			t.Fatal("the adopt proceeded over a run whose backend flipped to codex inside the window; want the refusal")
		}
		if !strings.Contains(err.Error(), "codex") {
			t.Errorf("refusal %q does not name the codex-led refusal", err.Error())
		}
		if c.args != nil {
			t.Errorf("the launch seam was called on a refused adopt")
		}
		if status, backend, pid, start, _ := adoptRunRow(t, root, "adopt-flip"); status != "active" || backend != "codex" || pid != 99999999 || start != "gone-before-explicit" {
			t.Errorf("row after refused adopt = (%s, %s, %d, %s), want the competing writer's (active, codex, 99999999, gone-before-explicit)", status, backend, pid, start)
		}
		if n := adoptEventCount(t, root, "adopt-flip", "run.resumed"); n != 0 {
			t.Errorf("run.resumed events = %d, want 0", n)
		}
		if after := adoptCardsSnapshot(t, root, "adopt-flip"); after != cardsBefore {
			t.Errorf("cards changed across the refused adopt")
		}
		if n := adoptRunsCount(t, root); n != 1 {
			t.Errorf("runs rows = %d, want 1", n)
		}
	})

	t.Run("legacy-peer-in-window", func(t *testing.T) {
		root := seedAdoptEntryRun(t, "adopt-peerw", "claude", false, 1)
		cardsBefore := adoptCardsSnapshot(t, root, "adopt-peerw")
		prod := classifyNamedRun
		classifyNamedRun = func(ctx context.Context, root, runID string) (homestate.RunOwner, error) {
			owner, err := prod(ctx, root, runID)
			if err != nil {
				return owner, err
			}
			// A legacy peer registers inside the window: the broker gains a
			// LIVE record the messaging layer refuses to address alongside
			// this leader.
			s, openErr := factorymsg.Open(root, runID)
			if openErr != nil {
				return owner, openErr
			}
			p := factorymsg.Peer{
				ProjectKey: homestate.ProjectKey(root), RunID: runID, Backend: "claude",
				Role: "worker", Slot: "worker-2", Generation: 1,
				PID: os.Getpid(), ProcessStart: homestate.CurrentProcessFingerprint(),
			}
			if _, err := s.RegisterLaunchPending(context.Background(), p); err != nil {
				_ = s.Close()
				return owner, err
			}
			_ = s.Close()
			return owner, nil
		}
		t.Cleanup(func() { classifyNamedRun = prod })

		c := installAdoptEntryEnv(t, root)
		err := runAdoptEntry(t, "-f", "--factory-run", "adopt-peerw")
		if err == nil {
			t.Fatal("the adopt proceeded over a run that gained a live legacy peer inside the window; want the refusal")
		}
		for _, want := range []string{"worker-2", "adopt-peerw"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal %q lacks %q", err.Error(), want)
			}
		}
		if c.args != nil {
			t.Errorf("the launch seam was called on a refused adopt")
		}
		if status, backend, pid, start, _ := adoptRunRow(t, root, "adopt-peerw"); status != "active" || backend != "claude" || pid != 99999999 || start != "gone-before-explicit" {
			t.Errorf("row after refused adopt = (%s, %s, %d, %s), want unchanged (active, claude, 99999999, gone-before-explicit)", status, backend, pid, start)
		}
		if n := adoptEventCount(t, root, "adopt-peerw", "run.resumed"); n != 0 {
			t.Errorf("run.resumed events = %d, want 0", n)
		}
		if after := adoptCardsSnapshot(t, root, "adopt-peerw"); after != cardsBefore {
			t.Errorf("cards changed across the refused adopt")
		}
		if n := adoptRunsCount(t, root); n != 1 {
			t.Errorf("runs rows = %d, want 1", n)
		}
	})
}

// ─── M3 — the orphan refusal gate (AC-FLAC-002, 005..009) ──────────────────

// countingReader records every Read invocation; the AC-FLAC-002 no-stdin
// guarantee counts the calls instead of consuming bytes (an empty reader's
// EOF proves nothing).
type countingReader struct{ calls int }

func (r *countingReader) Read([]byte) (int, error) { r.calls++; return 0, io.EOF }

// runEntryWithStderr drives one cc or glm entry through the real command and
// returns everything the entry wrote to its error stream.
func runEntryWithStderr(t *testing.T, launcher string, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	var err error
	if launcher == "glm" {
		glmCmd.SetOut(new(bytes.Buffer))
		glmCmd.SetErr(buf)
		err = runGLM(glmCmd, args)
	} else {
		ccCmd.SetOut(new(bytes.Buffer))
		ccCmd.SetErr(buf)
		err = runCC(ccCmd, args)
	}
	// Read the buffer only after the entry ran — evaluating buf.String()
	// inside the return statement would snapshot it before the call.
	return buf.String(), err
}

// gateRunSeed describes one run row the gate fixtures seed.
type gateRunSeed struct {
	id        string
	backend   string
	live      bool
	cards     []string // card states, in insert order
	workers   int
	status    string // "" = active
	createdAt string // pinned run-row created_at ("" = record time)
	laneCap   int
}

// seedGateRuns builds a sandbox git root with the requested run rows and
// returns the root.
func seedGateRuns(t *testing.T, seeds []gateRunSeed) string {
	t.Helper()
	root := t.TempDir()
	initGitRepo(t, root)
	if len(seeds) == 0 {
		return root
	}
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	for _, s := range seeds {
		run := homestate.FactoryRun{RunID: s.id, Backend: s.backend, ManifestJSON: "{}", LaneCapacity: s.laneCap}
		if s.live {
			run.LeadPID = os.Getpid()
			run.LeadProcessStart = homestate.CurrentProcessFingerprint()
		} else {
			run.LeadPID = 99999999
			run.LeadProcessStart = "gone-before-gate"
		}
		if err := db.RecordRun(context.Background(), run); err != nil {
			t.Fatalf("record run %s: %v", s.id, err)
		}
		for i, state := range s.cards {
			if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES(?,?,?,?)`,
				s.id, fmt.Sprintf("card-%s-%d", s.id, i), state, "2026-10-14T00:00:00Z"); err != nil {
				t.Fatalf("insert card for %s: %v", s.id, err)
			}
		}
		for i := 0; i < s.workers; i++ {
			if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
				fmt.Sprintf("lane-%d", i+1), 1, s.id, "2026-10-14T00:00:00Z", "2026-10-14T00:00:00Z"); err != nil {
				t.Fatalf("insert worker for %s: %v", s.id, err)
			}
		}
		if s.status != "" {
			if _, err := db.DB.Exec(`UPDATE runs SET status=? WHERE run_id=?`, s.status, s.id); err != nil {
				t.Fatalf("pin status for %s: %v", s.id, err)
			}
		}
		if s.createdAt != "" {
			if _, err := db.DB.Exec(`UPDATE runs SET created_at=? WHERE run_id=?`, s.createdAt, s.id); err != nil {
				t.Fatalf("pin created_at for %s: %v", s.id, err)
			}
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close seeding handle: %v", err)
	}
	return root
}

// gateRunRow is the comparable tuple of one run row.
type gateRunRow struct {
	status, backend, start string
	pid                    int
}

func gateRunRows(t *testing.T, root string) map[string]gateRunRow {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for row map: %v", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.DB.Query(`SELECT run_id, status, lead_backend, lead_pid, lead_process_start FROM runs ORDER BY run_id`)
	if err != nil {
		t.Fatalf("read run rows: %v", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]gateRunRow{}
	for rows.Next() {
		var id string
		var row gateRunRow
		if err := rows.Scan(&id, &row.status, &row.backend, &row.pid, &row.start); err != nil {
			t.Fatalf("scan run rows: %v", err)
		}
		out[id] = row
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("run rows: %v", err)
	}
	return out
}

// adoptContentSnapshot renders every row of the four tables a launch must not
// touch, ordered — the AC-FLAC-007 byte-identical instrument.
func adoptContentSnapshot(t *testing.T, root string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for content snapshot: %v", err)
	}
	defer func() { _ = db.Close() }()
	var b strings.Builder
	for _, table := range []struct {
		name  string
		cols  int
		query string
	}{
		{"runs", 10, `SELECT run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,lane_capacity,created_at,updated_at FROM runs ORDER BY run_id`},
		{"cards", 8, `SELECT run_id,card_id,owner_label,state,version,evidence_path,updated_at,stage FROM cards ORDER BY run_id,card_id`},
		{"events", 5, `SELECT seq,run_id,kind,payload_json,created_at FROM events ORDER BY seq`},
		{"workers", 7, `SELECT label,pid,backend,session_id,run_id,registered_at,heartbeat_at FROM workers ORDER BY label`},
	} {
		b.WriteString("== " + table.name + "\n")
		rows, err := db.DB.Query(table.query)
		if err != nil {
			t.Fatalf("snapshot %s: %v", table.name, err)
		}
		for rows.Next() {
			vals := make([]any, table.cols)
			ptrs := make([]any, table.cols)
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatalf("snapshot scan %s: %v", table.name, err)
			}
			for _, v := range vals {
				if raw, ok := v.([]byte); ok {
					b.WriteString(strings.ReplaceAll(string(raw), "\n", " "))
				} else {
					b.WriteString(strings.ReplaceAll(fmt.Sprintf("%v", v), "\n", " "))
				}
				b.WriteString("|")
			}
			b.WriteString("\n")
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("snapshot %s rows: %v", table.name, err)
		}
		_ = rows.Close()
	}
	return b.String()
}

// noRetiredEventAnywhere asserts REQ-FLAC-007's negative: no run.retired
// event exists in the sandbox at all.
func noRetiredEventAnywhere(t *testing.T, root string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory for retired scan: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM events WHERE kind='run.retired'`).Scan(&n); err != nil {
		t.Fatalf("count run.retired events: %v", err)
	}
	if n != 0 {
		t.Fatalf("run.retired events = %d, want 0 (no path this SPEC adds may retire a run)", n)
	}
}

// AC-FLAC-002 — the refusal precedes the run write: non-zero exit, nothing
// in the database changed, no session started, and the command read its
// input zero times.
func TestLeaderStartOrphanRefusalPrecedesRunWrite(t *testing.T) {
	root := seedGateRuns(t, []gateRunSeed{{id: "gate-orphan", backend: "claude", cards: []string{"run"}}})
	before := adoptContentSnapshot(t, root)
	c := installAdoptEntryEnv(t, root)
	reader := &countingReader{}
	ccCmd.SetIn(reader)
	t.Cleanup(func() { ccCmd.SetIn(nil) })
	stderr, err := runEntryWithStderr(t, "cc", "-f")
	if err == nil {
		t.Fatal("the launch over an orphan succeeded; want the orphan refusal")
	}
	if c.args != nil {
		t.Fatalf("the launch seam was called; the refusal must start no session")
	}
	if after := adoptContentSnapshot(t, root); after != before {
		t.Fatalf("the refusal wrote to the database:\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if reader.calls != 0 {
		t.Fatalf("the command read its input %d time(s); the refusal must read no stdin", reader.calls)
	}
	if !strings.Contains(stderr, leaderStartOrphanHeader) {
		t.Fatalf("stderr %q does not carry the guidance block header", stderr)
	}
}

// AC-FLAC-005 — one guidance block naming EACH orphan run: one fact line
// (run id, dead, lead pid, KST start time, non-terminal card count) and one
// adopt command per orphan, the launcher's own entry command, exactly one
// header, and nothing written anywhere.
func TestLeaderStartOrphanGuidanceRendersFacts(t *testing.T) {
	orphanSeeds := []gateRunSeed{
		{id: "orphan-a", backend: "claude", cards: []string{"run", "blocked", "done"}, createdAt: "2026-10-17T13:39:10.123456789Z", laneCap: 1},
		{id: "orphan-b", backend: "claude", cards: []string{"needs-decision"}, createdAt: "2026-10-16T08:00:00Z", laneCap: 1},
	}
	for _, launcher := range []string{"cc", "glm"} {
		t.Run(launcher, func(t *testing.T) {
			root := seedGateRuns(t, orphanSeeds)
			before := adoptContentSnapshot(t, root)
			rowsBefore := gateRunRows(t, root)
			c := installAdoptEntryEnv(t, root)
			entryCmd := "moai cc -f"
			if launcher == "glm" {
				entryCmd = "moai glm -f"
			}
			stderr, err := runEntryWithStderr(t, launcher, "-f")
			if err == nil {
				t.Fatal("the launch over two orphans succeeded; want the orphan refusal")
			}
			if c.args != nil {
				t.Fatalf("the launch seam was called")
			}
			if n := strings.Count(stderr, leaderStartOrphanHeader); n != 1 {
				t.Fatalf("guidance header occurred %d time(s), want exactly 1:\n%s", n, stderr)
			}
			for _, orphan := range []struct {
				id, pid, cards, started string
			}{
				{"orphan-a", "99999999", "2 in-progress card(s)", "2026-10-17 22:39:10 KST"},
				{"orphan-b", "99999999", "1 in-progress card(s)", "2026-10-16 17:00:00 KST"},
			} {
				if !strings.Contains(stderr, "run "+orphan.id+": owner dead") {
					t.Errorf("stderr does not name orphan %s with its classification:\n%s", orphan.id, stderr)
				}
				if !strings.Contains(stderr, "lead pid "+orphan.pid) {
					t.Errorf("stderr does not name orphan %s's lead pid:\n%s", orphan.id, stderr)
				}
				if !strings.Contains(stderr, orphan.cards) {
					t.Errorf("stderr does not carry %s's non-terminal card count (%s):\n%s", orphan.id, orphan.cards, stderr)
				}
				if !strings.Contains(stderr, "started "+orphan.started) {
					t.Errorf("stderr does not carry %s's start time rendered in KST (%s):\n%s", orphan.id, orphan.started, stderr)
				}
				if !strings.Contains(stderr, entryCmd+" --factory-run "+orphan.id) {
					t.Errorf("stderr does not carry %s's deliberate adopt command naming the launcher's own entry:\n%s", orphan.id, stderr)
				}
			}
			if !strings.Contains(stderr, "no run was created and no session was started") {
				t.Errorf("stderr does not carry the closing line:\n%s", stderr)
			}
			if after := adoptContentSnapshot(t, root); after != before {
				t.Errorf("the refusal wrote to the database")
			}
			rowsAfter := gateRunRows(t, root)
			for id, row := range rowsBefore {
				if rowsAfter[id] != row {
					t.Errorf("orphan row %s changed across the refusal", id)
				}
			}
		})
	}
}

// AC-FLAC-006 — the exemption table: every exempt launch behaves exactly as
// on the base tree (no guidance, no orphan refusal), while a `-c`/`-cr`
// start over an orphan is refused like a bare start.
func TestLeaderStartOrphanGating(t *testing.T) {
	orphanOnly := []gateRunSeed{{id: "gate-orphan", backend: "claude", cards: []string{"run"}}}
	twoOrphans := []gateRunSeed{
		{id: "orphan-a", backend: "claude", cards: []string{"run"}},
		{id: "orphan-b", backend: "claude", cards: []string{"blocked"}},
	}
	rows := []struct {
		name   string
		seed   []gateRunSeed
		env    func(t *testing.T)
		args   []string
		launch bool // the launch seam must be called
		err    bool // the entry must return a non-nil error
	}{
		{name: "resume-space-form", seed: orphanOnly, args: []string{"-f", "--resume", "sess-1"}, launch: true},
		{name: "resume-attached-form", seed: orphanOnly, args: []string{"-f", "-rsess-1"}, launch: true},
		{name: "factory-run-adopts", seed: orphanOnly, args: []string{"-f", "--factory-run", "gate-orphan"}, launch: true},
		{name: "non-empty-env-run-id", seed: orphanOnly, args: []string{"-f"}, launch: true, env: func(t *testing.T) { t.Setenv(config.EnvFactoryRunID, "deliberate04") }},
		{name: "pasted-leader-name", seed: orphanOnly, args: []string{"-f", "--name", "leader-deliberate05"}, launch: true},
		{name: "terminal-cards-only", seed: []gateRunSeed{{id: "gate-fin", backend: "claude", cards: []string{"done", "failed", "abandoned"}}}, args: []string{"-f"}, launch: true},
		{name: "workers-only", seed: []gateRunSeed{{id: "gate-w", backend: "claude", workers: 1}}, args: []string{"-f"}, launch: true},
		{name: "live-owner-with-card", seed: []gateRunSeed{{id: "gate-live", backend: "claude", live: true, cards: []string{"run"}}}, args: []string{"-f"}, launch: true},
		{name: "retired-run-with-card", seed: []gateRunSeed{{id: "gate-ret", backend: "claude", cards: []string{"run"}, status: "retired"}}, args: []string{"-f"}, launch: true},
		{name: "no-runs-at-all", seed: nil, args: []string{"-f"}, launch: true},
		{name: "continue-flag", seed: orphanOnly, args: []string{"-f", "-c", "sess-1"}, err: true},
		{name: "short-cluster-cr", seed: orphanOnly, args: []string{"-f", "-cr", "sess-1"}, err: true},
		{name: "resume-over-two-orphans", seed: twoOrphans, args: []string{"-f", "--resume", "sess-1"}, launch: true},
		{name: "factory-run-over-two-orphans", seed: twoOrphans, args: []string{"-f", "--factory-run", "orphan-a"}, launch: true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			root := seedGateRuns(t, row.seed)
			c := installAdoptEntryEnv(t, root)
			// Row env is set AFTER the installer: clearFactoryTestEnv inside
			// installAdoptEntryEnv would otherwise unset it again.
			if row.env != nil {
				row.env(t)
			}
			rowsBefore := gateRunRows(t, root)
			stderr, err := runEntryWithStderr(t, "cc", row.args...)
			if row.err && err == nil {
				t.Fatalf("launch = nil error, want the orphan refusal")
			}
			if !row.err && err != nil {
				t.Fatalf("launch refused: %v", err)
			}
			if strings.Contains(stderr, leaderStartOrphanHeader) != row.err {
				t.Fatalf("guidance header present = %v, want %v (the block rides only the refusal):\n%s", strings.Contains(stderr, leaderStartOrphanHeader), row.err, stderr)
			}
			if (c.args != nil) != row.launch {
				t.Fatalf("launch seam called = %v, want %v", c.args != nil, row.launch)
			}
			rowsAfter := gateRunRows(t, root)
			switch row.name {
			case "resume-space-form", "resume-attached-form", "factory-run-adopts", "factory-run-over-two-orphans":
				// The orphan the launch named was adopted: its row now names
				// this process; every OTHER row is byte-unchanged.
				adoptedID := "gate-orphan"
				if row.name == "factory-run-over-two-orphans" {
					adoptedID = "orphan-a"
				}
				if got := rowsAfter[adoptedID]; got.pid != os.Getpid() {
					t.Fatalf("adopted run %s stamp pid = %d, want this process", adoptedID, got.pid)
				}
				for id, before := range rowsBefore {
					if id == adoptedID {
						continue
					}
					if rowsAfter[id] != before {
						t.Errorf("row %s changed across the launch", id)
					}
				}
			case "resume-over-two-orphans":
				// The warn-and-mint outcome: BOTH orphans survive unchanged,
				// and the resume path warned on stderr.
				for id, before := range rowsBefore {
					if rowsAfter[id] != before {
						t.Errorf("orphan row %s changed across the resume launch", id)
					}
				}
				for _, id := range []string{"orphan-a", "orphan-b"} {
					if !strings.Contains(stderr, id) {
						t.Errorf("resume warn %q does not name candidate %s", stderr, id)
					}
				}
			case "non-empty-env-run-id":
				// Today's behavior: the launch records under the deliberate
				// environment id; the orphan row is untouched.
				if _, ok := rowsAfter["deliberate04"]; !ok {
					t.Fatalf("the deliberate environment id was not recorded")
				}
				if rowsAfter["gate-orphan"] != rowsBefore["gate-orphan"] {
					t.Errorf("the orphan row changed under a deliberate-id launch")
				}
			case "pasted-leader-name":
				if c.runID != "deliberate05" {
					t.Fatalf("published id = %q, want the pasted name's run id deliberate05", c.runID)
				}
			default:
				// Mint rows: no pre-existing row changed.
				for id, before := range rowsBefore {
					if rowsAfter[id] != before {
						t.Errorf("row %s changed across the mint launch", id)
					}
				}
			}
		})
	}
}

// AC-FLAC-007 — no path this SPEC adds touches any run other than the one
// adopted (or minted): the refusal changes nothing, the adopt changes only
// its own run, the mint adds only its own row, and the finder itself writes
// nothing. No run.retired event exists anywhere.
func TestLeaderStartOrphanTouchesNoOtherRun(t *testing.T) {
	fixture := []gateRunSeed{
		{id: "t-orphan", backend: "claude", cards: []string{"run"}},
		{id: "t-fin", backend: "claude", cards: []string{"done"}},
		{id: "t-live", backend: "claude", live: true, cards: []string{"run"}},
		{id: "t-ret", backend: "claude", cards: []string{"run"}, status: "retired"},
	}

	// The bare -f refusal over the orphan: every row byte-identical.
	root := seedGateRuns(t, fixture)
	before := adoptContentSnapshot(t, root)
	c := installAdoptEntryEnv(t, root)
	if _, err := runEntryWithStderr(t, "cc", "-f"); err == nil {
		t.Fatal("refusal launch = nil error")
	}
	if c.args != nil {
		t.Fatal("refusal launch started a session")
	}
	if after := adoptContentSnapshot(t, root); after != before {
		t.Fatalf("the refusal launch wrote to the database")
	}
	noRetiredEventAnywhere(t, root)

	// The deliberate adopt: only the adopted run's own row and event move.
	root = seedGateRuns(t, fixture)
	othersBefore := gateRunRows(t, root)
	installAdoptEntryEnv(t, root)
	if _, err := runEntryWithStderr(t, "cc", "-f", "--factory-run", "t-orphan"); err != nil {
		t.Fatalf("adopt launch refused: %v", err)
	}
	othersAfter := gateRunRows(t, root)
	for id, beforeRow := range othersBefore {
		if id == "t-orphan" {
			continue
		}
		if othersAfter[id] != beforeRow {
			t.Errorf("row %s changed across the adopt launch", id)
		}
	}
	if got := othersAfter["t-orphan"]; got.pid != os.Getpid() {
		t.Errorf("adopted run stamp pid = %d, want this process", got.pid)
	}
	noRetiredEventAnywhere(t, root)

	// The no-orphan mint: the minted row is the only new row.
	root = seedGateRuns(t, nil)
	c = installAdoptEntryEnv(t, root)
	if _, err := runEntryWithStderr(t, "cc", "-f"); err != nil {
		t.Fatalf("mint launch refused: %v", err)
	}
	if c.args == nil {
		t.Fatal("mint launch never started the session")
	}
	if n := adoptRunsCount(t, root); n != 1 {
		t.Errorf("runs rows after mint = %d, want 1", n)
	}
	noRetiredEventAnywhere(t, root)

	// The finder itself is read-only: a bare FindOrphanRuns call leaves a
	// full content snapshot byte-identical.
	root = seedGateRuns(t, fixture)
	before = adoptContentSnapshot(t, root)
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	if _, err := db.FindOrphanRuns(context.Background(), factorymsg.ReconcileOptionsFor(root)); err != nil {
		t.Fatalf("find orphan runs: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if after := adoptContentSnapshot(t, root); after != before {
		t.Fatalf("FindOrphanRuns wrote to the database")
	}
}

// AC-FLAC-008 — glm mirrors cc: one gate, one adoption path, differing only
// in the backend recorded and the entry command the text carries.
func TestLeaderStartOrphanGLMMirrorsCC(t *testing.T) {
	t.Run("takeover-backend", func(t *testing.T) {
		// cc adopting a glm-led run flips the row's backend to claude; glm
		// adopting a claude-led run flips it to glm — each leg's seed is the
		// OTHER launcher's backend, so neither leg's equality holds without
		// the adoption write.
		root := seedGateRuns(t, []gateRunSeed{{id: "mx-glm", backend: "glm", cards: []string{"run"}}})
		installAdoptEntryEnv(t, root)
		if _, err := runEntryWithStderr(t, "cc", "-f", "--factory-run", "mx-glm"); err != nil {
			t.Fatalf("cc adopt refused: %v", err)
		}
		if _, backend, _, _, _ := adoptRunRow(t, root, "mx-glm"); backend != "claude" {
			t.Errorf("cc takeover backend = %q, want claude", backend)
		}
		if n := adoptEventCount(t, root, "mx-glm", "run.resumed"); n != 1 {
			t.Errorf("cc leg run.resumed events = %d, want 1", n)
		}

		root = seedGateRuns(t, []gateRunSeed{{id: "mx-cc", backend: "claude", cards: []string{"run"}}})
		installAdoptEntryEnv(t, root)
		if _, err := runEntryWithStderr(t, "glm", "-f", "--factory-run", "mx-cc"); err != nil {
			t.Fatalf("glm adopt refused: %v", err)
		}
		if _, backend, _, _, _ := adoptRunRow(t, root, "mx-cc"); backend != "glm" {
			t.Errorf("glm takeover backend = %q, want glm", backend)
		}
		if n := adoptEventCount(t, root, "mx-cc", "run.resumed"); n != 1 {
			t.Errorf("glm leg run.resumed events = %d, want 1", n)
		}
	})

	t.Run("codex-refusal-text-equal", func(t *testing.T) {
		var texts []string
		for _, launcher := range []string{"cc", "glm"} {
			root := seedGateRuns(t, []gateRunSeed{{id: "mx-cx", backend: "codex", cards: []string{"run"}}})
			installAdoptEntryEnv(t, root)
			_, err := runEntryWithStderr(t, launcher, "-f", "--factory-run", "mx-cx")
			if err == nil {
				t.Fatalf("%s adopt of a codex-led run = nil error", launcher)
			}
			texts = append(texts, err.Error())
		}
		if texts[0] != texts[1] {
			t.Errorf("codex refusal texts differ:\ncc: %s\nglm: %s", texts[0], texts[1])
		}
	})

	t.Run("guidance-text-equal", func(t *testing.T) {
		// created_at is pinned: the block renders the row's start time, and
		// record-time values would differ between the two fixtures.
		guidanceSeeds := []gateRunSeed{
			{id: "orphan-a", backend: "claude", cards: []string{"run"}, createdAt: "2026-10-17T13:39:10.123456789Z"},
			{id: "orphan-b", backend: "claude", cards: []string{"blocked"}, createdAt: "2026-10-16T08:00:00Z"},
		}
		root := seedGateRuns(t, guidanceSeeds)
		installAdoptEntryEnv(t, root)
		stderrCC, errCC := runEntryWithStderr(t, "cc", "-f")
		if errCC == nil {
			t.Fatal("cc refusal = nil error")
		}
		if strings.Contains(stderrCC, "moai glm -f") {
			t.Errorf("cc guidance names the other launcher's command:\n%s", stderrCC)
		}
		if !strings.Contains(stderrCC, "moai cc -f --factory-run orphan-a") {
			t.Errorf("cc guidance does not name its own adopt command:\n%s", stderrCC)
		}
		root = seedGateRuns(t, guidanceSeeds)
		installAdoptEntryEnv(t, root)
		stderrGLM, errGLM := runEntryWithStderr(t, "glm", "-f")
		if errGLM == nil {
			t.Fatal("glm refusal = nil error")
		}
		if strings.Contains(stderrGLM, "moai cc -f") {
			t.Errorf("glm guidance names the other launcher's command:\n%s", stderrGLM)
		}
		if !strings.Contains(stderrGLM, "moai glm -f --factory-run orphan-a") {
			t.Errorf("glm guidance does not name its own adopt command:\n%s", stderrGLM)
		}
		if mirrored := strings.ReplaceAll(stderrGLM, "moai glm -f", "moai cc -f"); mirrored != stderrCC {
			t.Errorf("glm guidance does not mirror cc after the entry-command swap:\ncc:  %s\nglm: %s", stderrCC, mirrored)
		}
	})
}

// AC-FLAC-009 — every run start time the guidance block renders is KST
// (UTC+9), regardless of the host zone; an unparseable value renders
// unchanged.
func TestLeaderStartOrphanTimesRenderInKST(t *testing.T) {
	// The test pins time.Local to a zone that is neither UTC nor KST for its
	// duration: a renderer that uses time.Local or UTC fails.
	prev := time.Local
	time.Local = time.FixedZone("UTCTest", -5*3600)
	t.Cleanup(func() { time.Local = prev })

	if got := renderRunStartKST("2026-10-17T13:39:10.123456789Z"); got != "2026-10-17 22:39:10 KST" {
		t.Fatalf("renderRunStartKST = %q, want 2026-10-17 22:39:10 KST", got)
	}
	if got := renderRunStartKST("not-a-time"); got != "not-a-time" {
		t.Fatalf("renderRunStartKST = %q, want the unparseable value unchanged", got)
	}
	// The same rendering through the guidance block itself.
	var block bytes.Buffer
	renderLeaderStartOrphanGuidance(&block, "moai cc -f", []homestate.OrphanRun{
		{RunOwner: homestate.RunOwner{RunID: "orphan-a", Classification: homestate.OwnerDead, LeadPID: 99999999, CreatedAt: "2026-10-17T13:39:10.123456789Z"}, NonTerminalCards: 2},
	})
	if !strings.Contains(block.String(), "started 2026-10-17 22:39:10 KST") {
		t.Fatalf("guidance block does not render the KST time:\n%s", block.String())
	}
}

// AC-FLAC-010 — the distributed guidance carries the two anchor spans: each
// of the ten guidance files contains each anchor exactly once as a backticked
// span, no file carries a different value for either key, the four banned
// README legacy substrings are gone, the four docs-site pages agree on their
// level-2 heading count with both anchors inside one section each, and the
// skill template twin is byte-identical to the local skill. The spans are
// built from the same constants the guidance renderer ships with, so the
// docs and the behavior cannot drift apart. The skill document's own
// behavior-replay instrument (TestFactorySkillAssertionsMatchBehavior, E3)
// stays green alongside this test.
func TestLeaderAdoptGuidanceDocsMatchAnchors(t *testing.T) {
	root := repoRootForTest(t)
	read := func(rel string) string {
		raw, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		return string(raw)
	}
	anchors := []struct{ key, span string }{
		{"leader-start/orphan", "`" + LeaderStartOrphanAnchor + "`"},
		{"lane-join/dead-owner-with-cards-or-workers", "`" + LaneJoinDeadOwnerAnchor + "`"},
	}
	files := []string{
		"docs-site/content/en/advanced/factory-mode.md",
		"docs-site/content/ko/advanced/factory-mode.md",
		"docs-site/content/ja/advanced/factory-mode.md",
		"docs-site/content/zh/advanced/factory-mode.md",
		"README.md", "README.ko.md", "README.ja.md", "README.zh.md",
		factorySkillLocalRel, factorySkillTemplateRel,
	}
	for _, rel := range files {
		text := read(rel)
		for _, anchor := range anchors {
			if n := strings.Count(text, anchor.span); n != 1 {
				t.Errorf("%s: anchor %s occurs %d time(s), want exactly 1", rel, anchor.span, n)
			}
			// No span carries a different value for the same key.
			for _, span := range backtickSpan.FindAllStringSubmatch(text, -1) {
				if strings.HasPrefix(span[1], anchor.key+":") && "`"+span[1]+"`" != anchor.span {
					t.Errorf("%s: span %q states a different value for key %s", rel, span[1], anchor.key)
				}
			}
		}
	}

	// The banned legacy README sentences are gone (plan D16: REPLACED, never
	// qualified in place).
	banned := map[string]string{
		"README.md":    "is retired automatically when the next lane joins",
		"README.ko.md": "자동으로 회수되므로",
		"README.ja.md": "次にレーンが合流した時点で自動的に retire されるので",
		"README.zh.md": "会在下一条泳道加入时自动退役",
	}
	for rel, substr := range banned {
		if strings.Contains(read(rel), substr) {
			t.Errorf("%s still carries the banned legacy sentence fragment %q", rel, substr)
		}
	}

	// The four docs-site pages agree on their level-2 heading count, and both
	// anchors sit inside one level-2 section of each.
	sectionSplit := regexp.MustCompile("(?m)^## ")
	headings := map[string]int{}
	for _, rel := range []string{
		"docs-site/content/en/advanced/factory-mode.md",
		"docs-site/content/ko/advanced/factory-mode.md",
		"docs-site/content/ja/advanced/factory-mode.md",
		"docs-site/content/zh/advanced/factory-mode.md",
	} {
		text := read(rel)
		headings[rel] = len(sectionSplit.FindAllString(text, -1))
		sections := sectionSplit.Split(text, -1)
		for _, anchor := range anchors {
			home := -1
			for i, section := range sections {
				if strings.Contains(section, anchor.span) {
					home = i
					break
				}
			}
			if home < 0 {
				t.Errorf("%s: anchor %s sits in no level-2 section", rel, anchor.span)
				continue
			}
			other := anchors[0]
			if anchor.key == anchors[0].key {
				other = anchors[1]
			}
			shared := false
			for _, section := range sections {
				if strings.Contains(section, anchor.span) && strings.Contains(section, other.span) {
					shared = true
					break
				}
			}
			if !shared {
				t.Errorf("%s: anchors %s and %s do not share one level-2 section", rel, anchor.key, other.key)
			}
		}
	}
	counts := map[int]bool{}
	for _, n := range headings {
		counts[n] = true
	}
	if len(counts) != 1 {
		t.Errorf("docs-site level-2 heading counts diverge across locales: %v", headings)
	}

	// The skill template twin is byte-identical to the local skill.
	if read(factorySkillLocalRel) != read(factorySkillTemplateRel) {
		t.Errorf("the template skill twin differs from the local skill; make the twins byte-identical")
	}
}
