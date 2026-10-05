package hygiene

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Fixture time origin: every seed is dated relative to this instant and
// the GC's clock is pinned to it, making the deletion set deterministic.
var fixtureNow = time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)

// Fixed session keys (36-char hyphenated UUID shape).
const (
	keyDead    = "11111111-aaaa-bbbb-cccc-000000000001"
	keyGoal    = "11111111-aaaa-bbbb-cccc-000000000002"
	keyChain   = "11111111-aaaa-bbbb-cccc-000000000003"
	keyCap     = "11111111-aaaa-bbbb-cccc-000000000004"
	keyStops   = "11111111-aaaa-bbbb-cccc-000000000005"
	keyRouting = "11111111-aaaa-bbbb-cccc-000000000006"
	keyVerify  = "11111111-aaaa-bbbb-cccc-000000000007"
	keyTodo    = "11111111-aaaa-bbbb-cccc-000000000008"
	keyYoung   = "11111111-aaaa-bbbb-cccc-000000000009"
	keyMtime   = "11111111-aaaa-bbbb-cccc-00000000000a"
	keyLive    = "11111111-aaaa-bbbb-cccc-00000000000b"
)

// fixtureGC builds a GC over a temp .moai root with the clock pinned and
// a DEAD-by-default liveness seam (each test re-points the probes it
// needs).
func fixtureGC(t *testing.T, moaiRoot string) *GC {
	t.Helper()
	registerTestRoot(moaiRoot)
	g := &GC{
		MoaiRoot:         moaiRoot,
		MinAge:           testMinAge,
		TranscriptWindow: testActivityWindow,
		HeartbeatWindow:  testStaleHb,
		now:              func() time.Time { return fixtureNow },
	}
	g.Liveness = &Liveness{
		TranscriptWindow: testActivityWindow,
		HeartbeatWindow:  testStaleHb,
		now:              func() time.Time { return fixtureNow },
		registryEntry:    func(string) *RegistryEntry { return &RegistryEntry{PID: 4242} },
		pidProbe:         func(int) Signal { return SignalNegative },
		transcriptProbe:  func(string) Signal { return SignalNegative },
		heartbeatProbe:   func(string) Signal { return SignalNegative },
	}
	return g
}

// seedJSON writes a JSON object fixture under the .moai root (absolute
// paths only — REQ-HYG-015: no fixture may resolve against the process
// working directory).
func seedJSON(t *testing.T, moaiRoot, rel string, body map[string]any) {
	t.Helper()
	blob, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal %s: %v", rel, err)
	}
	seedFile(t, filepath.Join(moaiRoot, filepath.FromSlash(rel)), blob)
}

// relMoai returns the .moai-relative path joined onto a work root.
func relMoai(moaiRoot string, rel string) string {
	return filepath.Join(moaiRoot, filepath.FromSlash(rel))
}

// seedDeadTree plants one candidate per target class with old, datable
// content (the deletable set) plus the kept classes.
func seedDeadTree(t *testing.T, moaiRoot string) {
	old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	oldJSON := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC()
	young := fixtureNow.Add(-time.Hour).UTC().Format(time.RFC3339)

	seedJSON(t, moaiRoot, "state/context-usage/"+keyDead+".json", map[string]any{"captured_at": old})
	seedJSON(t, moaiRoot, "state/goal/"+keyGoal+".json", map[string]any{"created_at": old})
	seedFile(t, relMoai(moaiRoot, "state/goal/"+keyGoal+".html"), []byte("<html></html>"))
	seedJSON(t, moaiRoot, "state/goal/"+keyGoal+".verdict.json", map[string]any{"decision": "block"})
	seedJSON(t, moaiRoot, "state/codex-stop-chain/"+keyChain+".json", map[string]any{"recorded_at": oldJSON, "members": []any{}})
	// The cap counter carries no timestamp field: content-undatable.
	seedJSON(t, moaiRoot, "state/codex-stop-cap/"+keyCap+".json", map[string]any{"gates": map[string]any{}})
	seedJSON(t, moaiRoot, "state/agent-stops/"+keyStops+".json", map[string]any{
		"session_id": keyStops,
		"entries":    []any{map[string]any{"name": "t", "stopped_at": old}},
	})
	seedJSON(t, moaiRoot, "state/routing-pending-"+keyRouting+".json", map[string]any{"created_at": oldJSON})
	seedJSON(t, moaiRoot, "state/verify/"+keyVerify+"/old-check.json", map[string]any{"recorded_at": oldJSON})
	seedJSON(t, moaiRoot, "state/verify/"+keyVerify+"/young-check.json", map[string]any{"recorded_at": fixtureNow.Add(-time.Hour).UTC()})
	seedFile(t, relMoai(moaiRoot, "state/verify/"+keyVerify+"/notes.md"), []byte("undatable entry\n"))
	seedJSON(t, moaiRoot, "state/todo/"+keyTodo+".json", map[string]any{"entered_at": old, "role": "lane"})
	// Kept classes.
	seedJSON(t, moaiRoot, "state/context-usage/"+keyYoung+".json", map[string]any{"captured_at": young})
	// mtime-only: no content timestamp; the file mtime is left fresh —
	// either way it must never be consulted.
	seedFile(t, relMoai(moaiRoot, "state/context-usage/"+keyMtime+".json"), []byte(`{"raw_pct":0.1}`))
}

// growJSONL returns valid JSONL rows totalling at least n bytes — the
// over-threshold hygiene sink seed (the sink is parsed back as JSONL by
// the audit-read assertions).
func growJSONL(n int64) []byte {
	row := []byte(`{"ts":"2026-10-05T00:00:00Z","unit":"rotator","mode":"apply","outcome":"rotated","path":"seed/rule-load-audit.jsonl"}` + "\n")
	b := make([]byte, 0, n+int64(len(row)))
	for int64(len(b)) < n {
		b = append(b, row...)
	}
	return b
}

// findAll decisions by class+outcome.
func decisionsFor(r *GCReport, class TargetClass, outcome Outcome) []Decision {
	var out []Decision
	for _, d := range r.Decisions {
		if d.Class == class && d.Outcome == string(outcome) {
			out = append(out, d)
		}
	}
	return out
}

// TestApplyModeDeletionSet — AC-HYG-008 (L-008). Apply deletes exactly the
// DEAD + content-datable + aged set, keeps every other class, removes the
// verify directory only when empty, deletes the goal triple together with
// the dating member last, and reports already-vanished paths idempotently.
func TestApplyModeDeletionSet(t *testing.T) {
	t.Run("deletes exactly the eligible set", func(t *testing.T) {
		moaiRoot := filepath.Join(t.TempDir(), ".moai")
		seedDeadTree(t, moaiRoot)
		g := fixtureGC(t, moaiRoot)

		rep, err := g.Run(ModeApply)
		if err != nil {
			t.Fatalf("apply run: %v", err)
		}
		if len(decisionsFor(rep, ClassContextUsage, OutcomeDeleted)) != 1 {
			t.Fatalf("context-usage deletions = %d, want 1: %+v",
				len(decisionsFor(rep, ClassContextUsage, OutcomeDeleted)), rep.Decisions)
		}
		for _, rel := range []string{
			"state/context-usage/" + keyDead + ".json",
			"state/goal/" + keyGoal + ".json",
			"state/goal/" + keyGoal + ".html",
			"state/goal/" + keyGoal + ".verdict.json",
			"state/codex-stop-chain/" + keyChain + ".json",
			"state/agent-stops/" + keyStops + ".json",
			"state/routing-pending-" + keyRouting + ".json",
			"state/verify/" + keyVerify + "/old-check.json",
			"state/todo/" + keyTodo + ".json",
		} {
			if _, err := os.Stat(relMoai(moaiRoot, rel)); !os.IsNotExist(err) {
				t.Fatalf("expected deleted: %s (%v)", rel, err)
			}
		}
		// Kept: young, undatable cap counter, undatable verify entries,
		// mtime-only, todo's shared stores untouched by enumeration.
		for _, rel := range []string{
			"state/context-usage/" + keyYoung + ".json",
			"state/context-usage/" + keyMtime + ".json",
			"state/codex-stop-cap/" + keyCap + ".json",
			"state/verify/" + keyVerify + "/young-check.json",
			"state/verify/" + keyVerify + "/notes.md",
		} {
			if _, err := os.Stat(relMoai(moaiRoot, rel)); err != nil {
				t.Fatalf("kept candidate vanished: %s (%v)", rel, err)
			}
		}
		// The verify directory survived (young + undatable entries remain).
		if _, err := os.Stat(relMoai(moaiRoot, "state/verify/"+keyVerify)); err != nil {
			t.Fatalf("non-empty verify dir removed: %v", err)
		}
		// Keeping reasons present on kept decisions.
		for _, d := range rep.Decisions {
			if d.Outcome == string(OutcomeKept) && d.Reason == "" {
				t.Fatalf("kept decision without reason: %+v", d)
			}
		}
	})

	t.Run("verify dir removed only when empty", func(t *testing.T) {
		moaiRoot := filepath.Join(t.TempDir(), ".moai")
		oldJSON := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC()
		seedJSON(t, moaiRoot, "state/verify/"+keyVerify+"/old-check.json", map[string]any{"recorded_at": oldJSON})
		g := fixtureGC(t, moaiRoot)
		if _, err := g.Run(ModeApply); err != nil {
			t.Fatalf("apply: %v", err)
		}
		if _, err := os.Stat(relMoai(moaiRoot, "state/verify/"+keyVerify)); !os.IsNotExist(err) {
			t.Fatalf("emptied verify dir not removed: %v", err)
		}
	})

	t.Run("crash-mid-group: next pass completes the goal triple (D29)", func(t *testing.T) {
		moaiRoot := filepath.Join(t.TempDir(), ".moai")
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC()
		// A crashed pass removed the siblings; the dating member remains.
		seedJSON(t, moaiRoot, "state/goal/"+keyGoal+".json", map[string]any{"created_at": old.UTC().Format(time.RFC3339)})
		g := fixtureGC(t, moaiRoot)
		rep, err := g.Run(ModeApply)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if len(decisionsFor(rep, ClassGoalTriple, OutcomeDeleted)) != 1 {
			t.Fatalf("group not completed: %+v", rep.Decisions)
		}
		if _, err := os.Stat(relMoai(moaiRoot, "state/goal/"+keyGoal+".json")); !os.IsNotExist(err) {
			t.Fatalf("dating member still present: %v", err)
		}
	})

	t.Run("goal triple enumeration orders the dating member last (D29)", func(t *testing.T) {
		moaiRoot := filepath.Join(t.TempDir(), ".moai")
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC()
		seedJSON(t, moaiRoot, "state/goal/"+keyGoal+".json", map[string]any{"created_at": old.Format(time.RFC3339)})
		seedFile(t, relMoai(moaiRoot, "state/goal/"+keyGoal+".html"), []byte("x"))
		seedJSON(t, moaiRoot, "state/goal/"+keyGoal+".verdict.json", map[string]any{})
		cands, _, _, err := enumerateCandidates(moaiRoot)
		if err != nil {
			t.Fatalf("enumerate: %v", err)
		}
		for _, c := range cands {
			if c.Class == ClassGoalTriple {
				last := c.Paths[len(c.Paths)-1]
				if !strings.HasSuffix(last, keyGoal+".json") || strings.HasSuffix(last, ".verdict.json") {
					t.Fatalf("dating member not last: %v", c.Paths)
				}
				return
			}
		}
		t.Fatalf("goal candidate not enumerated")
	})

	t.Run("writer race: refreshed state survives via the re-judge (D28)", func(t *testing.T) {
		moaiRoot := filepath.Join(t.TempDir(), ".moai")
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		seedJSON(t, moaiRoot, "state/context-usage/"+keyDead+".json", map[string]any{"captured_at": old})
		g := fixtureGC(t, moaiRoot)
		g.preRejudge = func() {
			// The writer's temp-then-rename lands fresh state between
			// enumeration and the action-time re-judge.
			seedJSON(t, moaiRoot, "state/context-usage/"+keyDead+".json",
				map[string]any{"captured_at": fixtureNow.Add(-time.Second).UTC().Format(time.RFC3339)})
		}
		rep, err := g.Run(ModeApply)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if len(decisionsFor(rep, ClassContextUsage, OutcomeRejudgedKeep)) != 1 {
			t.Fatalf("expected rejudged-keep, got %+v", rep.Decisions)
		}
		if _, err := os.Stat(relMoai(moaiRoot, "state/context-usage/"+keyDead+".json")); err != nil {
			t.Fatalf("fresh state was deleted: %v", err)
		}
	})

	t.Run("already-vanished path reported without error", func(t *testing.T) {
		moaiRoot := filepath.Join(t.TempDir(), ".moai")
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC()
		// A crash left only the dating member: the missing sibling exercises
		// the already-gone arm inside a still-eligible group.
		seedJSON(t, moaiRoot, "state/goal/"+keyGoal+".json", map[string]any{"created_at": old.Format(time.RFC3339)})
		g := fixtureGC(t, moaiRoot)
		rep, err := g.Run(ModeApply)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if len(decisionsFor(rep, ClassGoalTriple, OutcomeAlreadyGone)) != 2 {
			t.Fatalf("already-gone rows = %d, want 2: %+v",
				len(decisionsFor(rep, ClassGoalTriple, OutcomeAlreadyGone)), rep.Decisions)
		}
	})
}

// TestReportModeByteIdentical — AC-HYG-007 (L-007). Report mode leaves the
// tree byte-identical (hygiene-audit.jsonl excepted), creates no lockfile,
// appends exactly one summary row per unit, records the over-threshold
// hygiene sink as skipped-report-mode, and states a keeping reason for
// every kept candidate.
func TestReportModeByteIdentical(t *testing.T) {
	root := t.TempDir()
	moaiRoot := filepath.Join(root, ".moai")
	logDir := filepath.Join(moaiRoot, "logs")
	registerTestRoot(moaiRoot)
	registerTestRoot(logDir)
	seedDeadTree(t, moaiRoot)
	// The hygiene sink itself, seeded over threshold with valid JSONL.
	seedFile(t, filepath.Join(logDir, "hygiene-audit.jsonl"), growJSONL(testOversized))

	treeHash := func() string {
		h := sha256.New()
		filepath.WalkDir(moaiRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(moaiRoot, path)
			if rel == filepath.Join("logs", "hygiene-audit.jsonl") {
				return nil // the named exception
			}
			h.Write([]byte(rel))
			if d.IsDir() {
				h.Write([]byte("/"))
				return nil
			}
			blob, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			h.Write(blob)
			return nil
		})
		return hex.EncodeToString(h.Sum(nil))
	}

	before := treeHash()

	r := &Rotator{LogDir: logDir, MaxBytes: testThreshold, KeptRotations: 1}
	if _, err := r.Run(ModeReport); err != nil {
		t.Fatalf("rotator report: %v", err)
	}
	g := fixtureGC(t, moaiRoot)
	rep, err := g.Run(ModeReport)
	if err != nil {
		t.Fatalf("gc report: %v", err)
	}

	after := treeHash()
	if before != after {
		t.Fatalf("report mode mutated the tree (hash %s -> %s)", before, after)
	}
	if _, err := os.Stat(filepath.Join(logDir, PassLockName)); !os.IsNotExist(err) {
		t.Fatalf("report mode created a lockfile: %v", err)
	}
	rows, err := readAuditRows(filepath.Join(logDir, "hygiene-audit.jsonl"))
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	rotatorSummaries, gcSummaries := 0, 0
	sinkRows := 0
	selfSkip := false
	for _, row := range rows {
		sinkRows++
		if row.Unit == "rotator" && row.Mode == string(ModeReport) {
			rotatorSummaries++
			if row.Counts != nil && row.Counts[string(OutcomeSkippedReportMode)] >= 1 {
				selfSkip = true
			}
		}
		if row.Unit == "gc" && row.Mode == string(ModeReport) {
			gcSummaries++
		}
	}
	// The seeded hygiene sink content is one oversized line-payload; the
	// two appended summary rows plus the seeded body coexist — the sink is
	// the named exception and the counts are what the AC pins.
	if rotatorSummaries != 1 {
		t.Fatalf("rotator summary rows = %d, want 1", rotatorSummaries)
	}
	if gcSummaries != 1 {
		t.Fatalf("gc summary rows = %d, want 1", gcSummaries)
	}
	if !selfSkip {
		t.Fatalf("over-threshold hygiene sink not recorded skipped-report-mode")
	}
	for _, d := range rep.Decisions {
		if d.Outcome == string(OutcomeKept) && d.Reason == "" {
			t.Fatalf("kept candidate without keeping reason: %+v", d)
		}
	}
}

// TestAuditRowsComplete — AC-HYG-009 (L-009). Apply rows carry path, unit,
// mode, decision, reason, and signal evidence; report rows are summaries.
func TestAuditRowsComplete(t *testing.T) {
	root := t.TempDir()
	moaiRoot := filepath.Join(root, ".moai")
	logDir := filepath.Join(moaiRoot, "logs")
	registerTestRoot(moaiRoot)
	registerTestRoot(logDir)
	seedDeadTree(t, moaiRoot)
	seedFile(t, filepath.Join(logDir, "rule-load-audit.jsonl"), growSeed(testOversized))

	r := &Rotator{LogDir: logDir, MaxBytes: testThreshold, KeptRotations: 1}
	if _, err := r.Run(ModeApply); err != nil {
		t.Fatalf("rotator apply: %v", err)
	}
	g := fixtureGC(t, moaiRoot)
	if _, err := g.Run(ModeApply); err != nil {
		t.Fatalf("gc apply: %v", err)
	}

	rows, err := readAuditRows(filepath.Join(logDir, "hygiene-audit.jsonl"))
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	gcDeletes := 0
	rotates := 0
	for _, row := range rows {
		if row.Mode != string(ModeApply) {
			continue
		}
		switch row.Unit {
		case "gc":
			if row.Outcome == OutcomeDeleted {
				gcDeletes++
				if row.Path == "" || row.Reason == "" || row.Signal == nil ||
					row.Signal["pid"] == "" || row.Signal["heartbeat"] == "" ||
					row.Signal["transcript"] == "" || row.Signal["content_date_age"] == "" {
					t.Fatalf("incomplete gc apply row: %+v", row)
				}
			}
		case "rotator":
			if row.Outcome == OutcomeRotated {
				rotates++
				if row.Path == "" {
					t.Fatalf("rotation row without path: %+v", row)
				}
			}
		}
	}
	if gcDeletes == 0 || rotates == 0 {
		t.Fatalf("seed run produced no deletions/rotations (gc=%d rot=%d)", gcDeletes, rotates)
	}

	// A report run appends exactly the two summary rows.
	beforeN := len(rows)
	if _, err := r.Run(ModeReport); err != nil {
		t.Fatalf("rotator report: %v", err)
	}
	if _, err := g.Run(ModeReport); err != nil {
		t.Fatalf("gc report: %v", err)
	}
	rows2, err := readAuditRows(filepath.Join(logDir, "hygiene-audit.jsonl"))
	if err != nil {
		t.Fatalf("re-read audit: %v", err)
	}
	if len(rows2)-beforeN != 2 {
		t.Fatalf("report pass appended %d rows, want 2", len(rows2)-beforeN)
	}
	for _, row := range rows2[beforeN:] {
		if row.Outcome != OutcomeSummary {
			t.Fatalf("report-mode row is not a summary: %+v", row)
		}
	}
}

// TestUnitIndependence — AC-HYG-010 (L-010). An injected rotator failure
// never blocks the GC and an injected GC failure never blocks the rotator.
func TestUnitIndependence(t *testing.T) {
	t.Run("rotator failure; gc completes", func(t *testing.T) {
		root := t.TempDir()
		moaiRoot := filepath.Join(root, ".moai")
		logDir := filepath.Join(moaiRoot, "logs")
		seedDeadTree(t, moaiRoot)
		e := &Engine{
			LogDir: logDir, MoaiRoot: moaiRoot,
			MaxBytes: testThreshold, KeptRotations: 2, // config-invalid by design
			MinAge: testMinAge, TranscriptWindow: testActivityWindow, HeartbeatWindow: testStaleHb,
		}
		registerTestRoot(moaiRoot)
		e.Run(ModeApply)
		// The engine call itself must not panic; assert through a direct
		// composition that the GC still completes after the rotator errors.
		r := &Rotator{LogDir: logDir, MaxBytes: testThreshold, KeptRotations: 2}
		_, rotErr := r.Run(ModeApply)
		if rotErr == nil {
			t.Fatalf("expected the injected rotator failure")
		}
		g := fixtureGC(t, moaiRoot)
		rep, err := g.Run(ModeApply)
		if err != nil || rep == nil || len(rep.Decisions) == 0 {
			t.Fatalf("gc did not complete after rotator failure: %v %+v", err, rep)
		}
	})

	t.Run("gc failure; rotator completes", func(t *testing.T) {
		root := t.TempDir()
		moaiRoot := filepath.Join(root, ".moai")
		logDir := filepath.Join(moaiRoot, "logs")
		registerTestRoot(logDir)
		seedFile(t, filepath.Join(logDir, "rule-load-audit.jsonl"), growSeed(testOversized))
		// A broken GC: its root path names a file, so enumeration fails.
		blocked := filepath.Join(root, ".moai-blocked")
		seedFile(t, blocked, []byte("not a directory"))
		registerTestRoot(blocked)
		g := fixtureGC(t, blocked)
		_, gcErr := g.Run(ModeApply)
		if gcErr == nil {
			t.Fatalf("expected the injected gc failure")
		}
		r := &Rotator{LogDir: logDir, MaxBytes: testThreshold, KeptRotations: 1}
		if _, err := r.Run(ModeApply); err != nil {
			t.Fatalf("rotator blocked by the gc failure: %v", err)
		}
		if _, err := os.Stat(filepath.Join(logDir, "rule-load-audit.jsonl.1")); err != nil {
			t.Fatalf("rotation did not happen: %v", err)
		}
	})
}

// TestSymlinkRefusalParentSwap — AC-HYG-011 (L-011). Symlinked components
// strictly below the resolved root are refused; symlinked ancestors ABOVE
// the root never count; a parent swapped to a symlink after the check is
// caught by the anchored action instead of escaping.
func TestSymlinkRefusalParentSwap(t *testing.T) {
	t.Run("symlinked component below the root refused", func(t *testing.T) {
		root := t.TempDir()
		moaiRoot := filepath.Join(root, ".moai")
		// The verify class dir is a symlink to a sibling real dir carrying
		// a datable, aged entry.
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC()
		seedJSON(t, root, "verify-real/"+keyVerify+"/old-check.json",
			map[string]any{"recorded_at": old})
		if err := os.MkdirAll(filepath.Join(moaiRoot, "state"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.Symlink(filepath.Join(root, "verify-real"),
			filepath.Join(moaiRoot, "state", "verify")); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		g := fixtureGC(t, moaiRoot)
		rep, err := g.Run(ModeApply)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if len(decisionsFor(rep, ClassVerifyScratch, OutcomeSymlinkRefused)) == 0 {
			t.Fatalf("expected symlink-refused decisions: %+v", rep.Decisions)
		}
		// The path is untouched.
		if _, err := os.Lstat(filepath.Join(root, "verify-real", keyVerify, "old-check.json")); err != nil {
			t.Fatalf("target touched through the symlink: %v", err)
		}
	})

	t.Run("symlinked ancestors above the root never count", func(t *testing.T) {
		real := t.TempDir()
		linkParent := t.TempDir()
		link := filepath.Join(linkParent, "link-to-real")
		if err := os.Symlink(real, link); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		moaiRoot := filepath.Join(link, ".moai")
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC()
		seedJSON(t, moaiRoot, "state/context-usage/"+keyDead+".json", map[string]any{"captured_at": old.Format(time.RFC3339)})
		g := fixtureGC(t, moaiRoot)
		rep, err := g.Run(ModeApply)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if len(decisionsFor(rep, ClassContextUsage, OutcomeDeleted)) != 1 {
			t.Fatalf("candidate under symlinked ancestors not processed: %+v", rep.Decisions)
		}
	})

	t.Run("parent swapped to a symlink after the check cannot escape (D27)", func(t *testing.T) {
		root := t.TempDir()
		moaiRoot := filepath.Join(root, ".moai")
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		seedJSON(t, moaiRoot, "state/context-usage/"+keyDead+".json", map[string]any{"captured_at": old})

		external := filepath.Join(t.TempDir(), "outside")
		if err := os.MkdirAll(external, 0o755); err != nil {
			t.Fatalf("mkdir external: %v", err)
		}
		sentinel := filepath.Join(external, keyDead+".json")
		if err := os.WriteFile(sentinel, []byte("external sentinel\n"), 0o644); err != nil {
			t.Fatalf("seed sentinel: %v", err)
		}

		g := fixtureGC(t, moaiRoot)
		g.preAction = func() {
			// Swap the class directory to a symlink pointing OUTSIDE the
			// resolved root after the component check has passed: the real
			// dir is renamed aside and a symlink takes its place.
			classDir := filepath.Join(moaiRoot, "state", "context-usage")
			if err := os.Rename(classDir, classDir+".swapped"); err != nil {
				t.Errorf("swap rename: %v", err)
				return
			}
			if err := os.Symlink(external, classDir); err != nil {
				t.Errorf("swap symlink: %v", err)
			}
		}
		rep, err := g.Run(ModeApply)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		// The anchored action refused; nothing was deleted through the
		// swapped path; the external sentinel survived.
		if len(decisionsFor(rep, ClassContextUsage, OutcomeDeleted)) != 0 {
			t.Fatalf("deletion escaped through a swapped parent: %+v", rep.Decisions)
		}
		if _, err := os.Stat(sentinel); err != nil {
			t.Fatalf("external sentinel harmed: %v", err)
		}
	})
}

// TestLockClassExcluded — AC-HYG-012 (L-012). Lock-named scan hits are
// reported excluded and left byte-identical in both modes, and the target
// registry is structurally free of a lock class.
func TestLockClassExcluded(t *testing.T) {
	for _, mode := range []Mode{ModeReport, ModeApply} {
		root := t.TempDir()
		moaiRoot := filepath.Join(root, ".moai")
		lockRel := "state/spec-close-SPEC-FOO-123.lock"
		lockBody := []byte("lock byte body — never touched\n")
		seedFile(t, relMoai(moaiRoot, lockRel), lockBody)
		old := fixtureNow.Add(-testMinAge).Add(-24 * time.Hour).UTC().Format(time.RFC3339)
		seedJSON(t, moaiRoot, "state/context-usage/"+keyDead+".json", map[string]any{"captured_at": old})

		g := fixtureGC(t, moaiRoot)
		rep, err := g.Run(mode)
		if err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		excluded := 0
		for _, d := range rep.Decisions {
			if d.Outcome == string(OutcomeLockClassExcluded) &&
				strings.HasSuffix(d.Path, "spec-close-SPEC-FOO-123.lock") {
				excluded++
			}
		}
		if excluded != 1 {
			t.Fatalf("%s: lock-class-excluded rows = %d, want 1", mode, excluded)
		}
		if got := readBytes(t, relMoai(moaiRoot, lockRel)); string(got) != string(lockBody) {
			t.Fatalf("%s: lock file modified", mode)
		}
	}
	for _, class := range TargetClasses() {
		if strings.Contains(strings.ToLower(string(class)), "lock") {
			t.Fatalf("target registry carries a lock class: %s", class)
		}
	}
}

// TestTargetRegistryShapes — structural pins for the D31 table: the
// unresolvable-key class is reported spared, and the todo enumeration
// skips the shared stores.
func TestTargetRegistryShapes(t *testing.T) {
	root := t.TempDir()
	moaiRoot := filepath.Join(root, ".moai")
	seedJSON(t, moaiRoot, "state/context-usage/not-a-session.json", map[string]any{"captured_at": "x"})
	seedJSON(t, moaiRoot, "state/todo/"+keyTodo+".json", map[string]any{"entered_at": "2026-08-01T00:00:00Z"})
	seedFile(t, relMoai(moaiRoot, "state/todo/backlog.json"), []byte("shared store\n"))

	g := fixtureGC(t, moaiRoot)
	rep, err := g.Run(ModeReport)
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	spared := false
	for _, d := range rep.Decisions {
		if strings.HasSuffix(d.Path, "not-a-session.json") && d.Outcome == string(OutcomeKept) {
			spared = true
		}
		if strings.Contains(d.Path, "backlog") {
			t.Fatalf("shared store evaluated: %+v", d)
		}
	}
	if !spared {
		t.Fatalf("unresolvable-key candidate not reported spared: %+v", rep.Decisions)
	}
	if len(decisionsFor(rep, ClassTodoRecord, OutcomeKept)) != 1 {
		t.Fatalf("todo session record not evaluated exactly once: %+v", rep.Decisions)
	}
}
