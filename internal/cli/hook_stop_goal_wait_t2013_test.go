package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/goal"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// The stop-goal legitimate-wait exemption, driven through the real hook
// function with real stdin JSON, a real factory database in a sandboxed home,
// and a real goal state file. The hook resolves the card from the hook input's
// cwd and reads the factory record read-only.

const stopWaitSession = "t2013-wait"

type stopWaitEnv struct {
	root   string // project root (CLAUDE_PROJECT_DIR)
	wt     string // the recorded card worktree
	cwd    string // a directory inside the card worktree
	dbPath string // where the factory database lives (or would live)
}

func newStopWaitEnv(t *testing.T) stopWaitEnv {
	t.Helper()
	root := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv(paths.EnvHome, t.TempDir())
	wt := filepath.Join(t.TempDir(), "t2020")
	cwd := filepath.Join(wt, "internal", "x")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath, err := homestate.FactoryDBPath(factoryRootOf(root))
	if err != nil {
		t.Fatalf("factory db path: %v", err)
	}
	return stopWaitEnv{root: root, wt: wt, cwd: cwd, dbPath: dbPath}
}

// seedCard records the card t2020 bound to the worktree plus the given waits,
// then closes the database so the hook opens it fresh.
func (e stopWaitEnv) seedCard(t *testing.T, waits ...homestate.CardWait) {
	t.Helper()
	fcPlace(t, factoryRootOf(e.root), homestate.Card{CardID: "t2020", State: homestate.CardRun, WorktreePath: e.wt})
	db, err := homestate.OpenFactory(factoryRootOf(e.root))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	for _, w := range waits {
		if err := db.RecordCardWait(context.Background(), w, time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC)); err != nil {
			t.Fatalf("record wait: %v", err)
		}
	}
}

// armModelGoal writes an armed model-only goal, the long-wait goal shape.
func (e stopWaitEnv) armModelGoal(t *testing.T) {
	t.Helper()
	dir := filepath.Join(e.root, goalStateDirForTest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	state := `{"session_id":"` + stopWaitSession + `","goal":"wait","conditions":[{"type":"model","claim":"the predecessor landed"}],` +
		`"ceiling":{"max_turns":30},"turns_used":0,"progress":[],"progression_mode":"autonomous","created_at":"2026-09-13T00:00:00Z","status":"armed"}`
	if err := os.WriteFile(filepath.Join(dir, stopWaitSession+".json"), []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (e stopWaitEnv) hookInput(withCwd bool) string {
	if !withCwd {
		return `{"session_id":"` + stopWaitSession + `"}`
	}
	return `{"session_id":"` + stopWaitSession + `","cwd":"` + e.cwd + `"}`
}

func (e stopWaitEnv) goalState(t *testing.T) *goal.Goal {
	t.Helper()
	g, err := goal.LoadGoal(e.root, stopWaitSession)
	if err != nil || g == nil {
		t.Fatalf("load goal: %v %v", g, err)
	}
	return g
}

func operatorWait(reason string) homestate.CardWait {
	return homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondOperatorRuling, Waker: homestate.WaitWakerOperator, Reason: reason, RecordedBy: "lane-3"}
}

func assertBlocked(t *testing.T, stdout string) {
	t.Helper()
	assertSingleJSONObject(t, stdout)
	if !strings.Contains(stdout, `"decision":"block"`) {
		t.Fatalf("want a block decision, got stdout=%q", stdout)
	}
}

// (7) An open wait that names a next event, with the cwd inside the card's
// recorded worktree: no block, the goal stays armed, the turn still counts.
func TestStopGoalWait_OpenWaitWithNextEventDoesNotBlock(t *testing.T) {
	e := newStopWaitEnv(t)
	next := "operator ruling arrives"
	reason, err := factoryWaitReason("needs the ruling", nil, &next)
	if err != nil {
		t.Fatal(err)
	}
	e.seedCard(t, operatorWait(reason))
	e.armModelGoal(t)

	stdout, stderr := driveStopGoalStreams(t, e.hookInput(true))
	assertSingleJSONObject(t, stdout)
	if strings.Contains(stdout, "block") {
		t.Fatalf("a valid wait must not block, stdout=%q stderr=%q", stdout, stderr)
	}
	g := e.goalState(t)
	if g.Status != goal.StatusArmed || g.TurnsUsed != 1 || len(g.Progress) != 1 {
		t.Errorf("status=%q turns=%d progress=%+v, want armed / 1 / one entry", g.Status, g.TurnsUsed, g.Progress)
	}
}

// (7c) The read-only lookup works while another process holds the database
// open for writing, the situation a running factory is in.
func TestStopGoalWait_ReadsWhileAWriterHoldsTheDatabase(t *testing.T) {
	e := newStopWaitEnv(t)
	next := "operator ruling arrives"
	reason, _ := factoryWaitReason("needs the ruling", nil, &next)
	e.seedCard(t, operatorWait(reason))
	e.armModelGoal(t)
	writer, err := homestate.OpenFactory(factoryRootOf(e.root))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = writer.Close() }()
	if _, err := writer.DB.Exec(`BEGIN IMMEDIATE`); err != nil {
		t.Fatalf("hold a write lock: %v", err)
	}
	defer func() { _, _ = writer.DB.Exec(`ROLLBACK`) }()

	stdout, stderr := driveStopGoalStreams(t, e.hookInput(true))
	if strings.Contains(stdout, "block") {
		t.Fatalf("a live writer must not defeat the read-only lookup, stdout=%q stderr=%q", stdout, stderr)
	}
}

// (7b) A scope tag in front of the next tag is read the same way.
func TestStopGoalWait_ScopeThenNextIsValid(t *testing.T) {
	e := newStopWaitEnv(t)
	scope, next := "Write internal/x.go denied", "permission granted"
	reason, err := factoryWaitReason("blocked", &scope, &next)
	if err != nil {
		t.Fatal(err)
	}
	e.seedCard(t, operatorWait(reason))
	e.armModelGoal(t)
	stdout, _ := driveStopGoalStreams(t, e.hookInput(true))
	if strings.Contains(stdout, "block") {
		t.Fatalf("scope then next must still be a valid wait, stdout=%q", stdout)
	}
}

// (8) A wait without a next event never exempts.
func TestStopGoalWait_WaitWithoutNextEventStillBlocks(t *testing.T) {
	scope := "Write internal/x.go denied"
	scopeOnly, err := factoryWaitReason("blocked", &scope, nil)
	if err != nil {
		t.Fatal(err)
	}
	unknown := "unknown"
	nextUnknown, err := factoryWaitReason("blocked", nil, &unknown)
	if err != nil {
		t.Fatal(err)
	}
	for name, reason := range map[string]string{
		"plain reason":     "needs the ruling",
		"empty reason":     "",
		"scope only":       scopeOnly,
		"next unknown":     nextUnknown,
		"next in the text": "needs the ruling, next=soon", // not a leading tag
	} {
		t.Run(name, func(t *testing.T) {
			e := newStopWaitEnv(t)
			e.seedCard(t, operatorWait(reason))
			e.armModelGoal(t)
			stdout, _ := driveStopGoalStreams(t, e.hookInput(true))
			assertBlocked(t, stdout)
		})
	}
}

// (9) A woken wait is history, and no wait at all is the plain case: both block.
func TestStopGoalWait_WokenOrAbsentWaitBlocks(t *testing.T) {
	t.Run("woken", func(t *testing.T) {
		e := newStopWaitEnv(t)
		next := "t2010 lands"
		reason, err := factoryWaitReason("needs t2010", nil, &next)
		if err != nil {
			t.Fatal(err)
		}
		w := homestate.CardWait{CardID: "t2020", CondKind: homestate.WaitCondPredecessorLanded, CondRef: "t2010", Waker: "lane-3", Reason: reason, RecordedBy: "lane-3"}
		e.seedCard(t, w)
		db, err := homestate.OpenFactory(factoryRootOf(e.root))
		if err != nil {
			t.Fatal(err)
		}
		settled, err := db.SettleLandedWaits(context.Background(), "t2010", time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC))
		_ = db.Close()
		if err != nil || len(settled) != 1 {
			t.Fatalf("settle: %v %+v", err, settled)
		}
		e.armModelGoal(t)
		stdout, _ := driveStopGoalStreams(t, e.hookInput(true))
		assertBlocked(t, stdout)
	})
	t.Run("no wait", func(t *testing.T) {
		e := newStopWaitEnv(t)
		e.seedCard(t)
		e.armModelGoal(t)
		stdout, _ := driveStopGoalStreams(t, e.hookInput(true))
		assertBlocked(t, stdout)
	})
	t.Run("cwd outside any recorded card worktree", func(t *testing.T) {
		e := newStopWaitEnv(t)
		next := "operator ruling arrives"
		reason, _ := factoryWaitReason("needs the ruling", nil, &next)
		e.seedCard(t, operatorWait(reason))
		e.armModelGoal(t)
		e.cwd = t.TempDir()
		stdout, _ := driveStopGoalStreams(t, e.hookInput(true))
		assertBlocked(t, stdout)
	})
}

// (10) A hook input that carries no cwd cannot resolve a card: the goal blocks
// as before and the skipped lookup is named on stderr. The hook protocol fills
// an absent cwd from CLAUDE_PROJECT_DIR, so the truly cwd-less input is one with
// neither: the project root then falls back to the process directory.
func TestStopGoalWait_NoCwdBlocksAndSaysSo(t *testing.T) {
	e := newStopWaitEnv(t)
	next := "operator ruling arrives"
	reason, _ := factoryWaitReason("needs the ruling", nil, &next)
	e.seedCard(t, operatorWait(reason))
	e.armModelGoal(t)
	t.Setenv(config.EnvClaudeProjectDir, "")
	t.Chdir(e.root)
	stdout, stderr := driveStopGoalStreams(t, e.hookInput(false))
	assertBlocked(t, stdout)
	if !strings.Contains(stderr, "stop-goal: wait lookup skipped (no cwd in hook input)") {
		t.Errorf("stderr lacks the skipped-lookup line: %q", stderr)
	}
	if strings.Contains(stdout, "skipped") {
		t.Errorf("the diagnostic must stay off stdout: %q", stdout)
	}
}

// (10b) With CLAUDE_PROJECT_DIR set, the protocol fills the absent cwd from it;
// a session standing in the project root is not inside a card worktree, so the
// goal blocks and the lookup says why.
func TestStopGoalWait_ProjectDirFallbackIsNotACardWorktree(t *testing.T) {
	e := newStopWaitEnv(t)
	next := "operator ruling arrives"
	reason, _ := factoryWaitReason("needs the ruling", nil, &next)
	e.seedCard(t, operatorWait(reason))
	e.armModelGoal(t)
	stdout, stderr := driveStopGoalStreams(t, e.hookInput(false))
	assertBlocked(t, stdout)
	if !strings.Contains(stderr, "stop-goal: wait lookup skipped (cwd is not inside a recorded card worktree)") {
		t.Errorf("stderr lacks the not-a-card-worktree line: %q", stderr)
	}
}

// (11) A missing factory database: the goal blocks as before and nothing is
// created on disk, neither the database nor its directory.
func TestStopGoalWait_MissingDatabaseBlocksAndCreatesNothing(t *testing.T) {
	e := newStopWaitEnv(t)
	e.armModelGoal(t)
	stdout, _ := driveStopGoalStreams(t, e.hookInput(true))
	assertBlocked(t, stdout)
	for _, p := range []string{e.dbPath, e.dbPath + "-wal", e.dbPath + "-shm", filepath.Dir(e.dbPath)} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s exists after a lookup that must write nothing (stat err=%v)", p, err)
		}
	}
}

// (12) The bounds that the exemption must not touch stay what they were.
func TestStopGoalWait_BoundsUnchanged(t *testing.T) {
	if DefaultRaisedStopHookBlockCap != 200 {
		t.Errorf("DefaultRaisedStopHookBlockCap = %d, want 200", DefaultRaisedStopHookBlockCap)
	}
	if goal.DefaultStagnationThreshold != 3 {
		t.Errorf("DefaultStagnationThreshold = %d, want 3", goal.DefaultStagnationThreshold)
	}
	if goal.DefaultMaxTurns != 30 || goal.NewGoal("s", "g", nil).Ceiling.MaxTurns != 30 {
		t.Errorf("default MaxTurns changed: const=%d", goal.DefaultMaxTurns)
	}
	if stopGoalHookTimeout != 90*time.Second {
		t.Errorf("stopGoalHookTimeout = %s, want 90s", stopGoalHookTimeout)
	}
	tmpl, err := os.ReadFile(filepath.Join("..", "template", "templates", ".claude", "settings.json.tmpl"))
	if err != nil {
		t.Fatalf("read settings template: %v", err)
	}
	lines := strings.Split(string(tmpl), "\n")
	found := false
	for i, l := range lines {
		if strings.Contains(l, "handle-stop-goal.sh") {
			found = true
			if i+1 >= len(lines) || !strings.Contains(lines[i+1], `"timeout": 120`) {
				t.Errorf("the stop-goal hook wrapper timeout is no longer 120: %q", lines[i+1])
			}
		}
	}
	if !found {
		t.Error("handle-stop-goal.sh not found in the settings template")
	}
}
