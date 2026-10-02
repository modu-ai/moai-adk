// factory_m5_test.go — SPEC-FACTORY-SELF-DISPATCH-001 M5 AC tests (card
// t1240): the Codex per-card relaunch (`moai codex -l`, REQ-SD-003 /
// AC-SD-003) and the refusal of every other Codex factory entry shape
// (REQ-SD-004 / AC-SD-004). The AC-SD-007 enumeration extension lives with
// the walk it extends (factory_m4_test.go).
//
// Every fixture is built under t.TempDir() with an isolated git config and
// MOAI_HOME sandboxed away (§B of acceptance.md); ./internal/cli runs only
// through the anchored -run selectors naming one of these tests. No test
// starts a real codex binary: the binary lookup and the child launch are
// substituted, and the substitute simulates the session's own work.
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// sdCardLaunchCapture is one substituted Codex child: the argv the launcher
// assembled, the child's working directory, and its environment.
type sdCardLaunchCapture struct {
	argv []string
	dir  string
	env  map[string]string
}

// sdEnvOf flattens one exec.Cmd environment into a map.
func sdEnvOf(t *testing.T, env []string) map[string]string {
	t.Helper()
	m := make(map[string]string, len(env))
	for _, kv := range env {
		name, value, found := strings.Cut(kv, "=")
		if found {
			m[name] = value
		}
	}
	return m
}

// sdCodexSessionWork simulates the substituted Codex session: it moves its
// leased card to merge-ready through the same F1 stage chain the lane's own
// work uses (the AC-SD-006 shape), then the caller returns nil — the child's
// exit 0.
func sdCodexSessionWork(t *testing.T, root, cardID string) {
	t.Helper()
	card := fcCard(t, root, cardID)
	wt := card.WorktreePath
	if err := os.WriteFile(filepath.Join(wt, "done.txt"), []byte("work\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fcGit(t, wt, "add", "-A")
	fcGit(t, wt, "commit", "-q", "-m", "card work")
	sha := fcGit(t, wt, "rev-parse", "HEAD")
	if _, _, err := runFactory(t, "stage", cardID, "run", "--run", fcRun); err != nil {
		t.Fatalf("stage %s run: %v", cardID, err)
	}
	if _, _, err := runFactory(t, "stage", cardID, "sync", sha, "--run", fcRun); err != nil {
		t.Fatalf("stage %s sync: %v", cardID, err)
	}
	if _, _, err := runFactory(t, "stage", cardID, "sync-audit", sha+":done.txt", "--run", fcRun); err != nil {
		t.Fatalf("stage %s sync-audit: %v", cardID, err)
	}
	verdict := filepath.Join(wt, ".moai", "reports", cardID, "sync-audit.md")
	if err := os.MkdirAll(filepath.Dir(verdict), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(verdict, []byte("verdict: PASS\naudited_sha: "+sha+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runFactory(t, "stage", cardID, "merge-ready", "--run", fcRun); err != nil {
		t.Fatalf("stage %s merge-ready: %v", cardID, err)
	}
}

// AC-SD-003 — `moai codex -l` is a supervising loop: two operator-picked
// cards, a substituted Codex session that exits 0 after moving its card to
// merge-ready, then one more invocation per card with that card's worktree as
// the child's working directory and the marker, the lane label, the Codex
// backend value, and that card's id in the card-identifier variable — and the
// launcher exits 0 once `next` reports no card.
func TestSD_AC003_CodexRelaunchPerCard(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, kanban.BacklogStatePicked, kanban.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, kanban.BackendClaude)
	wantLabel := sdNextFreeLaneLabel(t, root)
	t.Chdir(root)
	sdScrubLauncherEnv(t)

	var got []sdCardLaunchCapture
	prevLook := codexLookPath
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	prevDirect := codexDirectLaunchFn
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		cap := sdCardLaunchCapture{
			argv: append([]string(nil), c.Args...),
			dir:  c.Dir,
			env:  sdEnvOf(t, c.Env),
		}
		got = append(got, cap)
		// The substituted session moves ITS OWN card (the id the launcher
		// handed it) to merge-ready, then exits 0.
		sdCodexSessionWork(t, root, cap.env[config.EnvMoaiKanbanCard])
		return nil
	}
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })

	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex lane: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("the substituted Codex session was invoked %d times, want 2 (once per operator-picked card)", len(got))
	}
	for i, cardID := range []string{"t1", "t2"} {
		card := fcCard(t, root, cardID)
		if card.State != homestate.CardMergeReady {
			t.Errorf("%s ended at %s, want merge-ready (the substitute moved it)", cardID, card.State)
		}
		if card.WorktreePath == "" {
			t.Fatalf("%s recorded no worktree", cardID)
		}
		rec := got[i]
		if rec.dir != card.WorktreePath {
			t.Errorf("invocation %d: child working directory = %q, want card %s's own worktree %q", i, rec.dir, cardID, card.WorktreePath)
		}
		if !containsPair(rec.argv, "-C", card.WorktreePath) {
			t.Errorf("invocation %d: argv %v does not carry -C with the card worktree", i, rec.argv)
		}
		if rec.env[config.EnvFactoryRole] != config.FactoryRoleLane {
			t.Errorf("invocation %d: child env %s = %q, want the value constant %q", i, config.EnvFactoryRole, rec.env[config.EnvFactoryRole], config.FactoryRoleLane)
		}
		if rec.env[config.EnvMoaiFactoryWorker] != wantLabel {
			t.Errorf("invocation %d: child env %s = %q, want the lane label %q", i, config.EnvMoaiFactoryWorker, rec.env[config.EnvMoaiFactoryWorker], wantLabel)
		}
		// The factory card verbs the owned-card session runs read the lane
		// label from MOAI_KANBAN_LABEL; a child without it could not stage
		// its own card.
		if rec.env[config.EnvMoaiKanbanLabel] != wantLabel {
			t.Errorf("invocation %d: child env %s = %q, want the lane label %q (the carrier the factory card verbs read)", i, config.EnvMoaiKanbanLabel, rec.env[config.EnvMoaiKanbanLabel], wantLabel)
		}
		if rec.env[config.EnvMoaiKanbanBackend] != kanban.BackendGPT {
			t.Errorf("invocation %d: child env %s = %q, want the Codex harness value %q", i, config.EnvMoaiKanbanBackend, rec.env[config.EnvMoaiKanbanBackend], kanban.BackendGPT)
		}
		if rec.env[config.EnvMoaiKanbanCard] != cardID {
			t.Errorf("invocation %d: child env %s = %q, want card %s's id", i, config.EnvMoaiKanbanCard, rec.env[config.EnvMoaiKanbanCard], cardID)
		}
		if rec.env[config.EnvMoaiKanbanID] != "" {
			t.Errorf("invocation %d: child env carries the run id %s=%q; the eleven-key scrub holds on the lane path too", i, config.EnvMoaiKanbanID, rec.env[config.EnvMoaiKanbanID])
		}
	}
}

// containsPair reports whether argv carries flag immediately followed by value.
func containsPair(argv []string, flag, value string) bool {
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == flag && argv[i+1] == value {
			return true
		}
	}
	return false
}

// AC-SD-004 — every other Codex factory entry shape is refused with ONE line,
// defined once: it carries FACTORY_MODE_UNSUPPORTED_BACKEND, names
// `moai codex -l` as the Codex lane entry and `moai cc -f` / `moai glm -f` for
// the leader, exits 1, and starts no child. M3 (SPEC-LAUNCHER-ENTRY-FLAGS-001)
// re-pinned the line: `-f` is no Codex entry in any shape, `-f lane` included.
func TestSD_AC004_CodexOtherFactoryShapesRefused(t *testing.T) {
	for _, shape := range [][]string{
		{"-f"},
		{"--factory"},
		{"--factory-run", "x"},
		{"-f", "lane"},
		{"-f", "lane-2"},
		{"-f", "3"},
		{"--factory", "lane"},
		{"-f="},
		{"--factory="},
	} {
		prevDirect, prevSpawn := codexDirectLaunchFn, codexSpawnLaunchFn
		codexDirectLaunchFn = func(*exec.Cmd) error {
			t.Errorf("shape %v: the direct launch ran; the refusal must start no child", shape)
			return nil
		}
		codexSpawnLaunchFn = func(string, string, []string, []string) error {
			t.Errorf("shape %v: the spawn launch ran; the refusal must start no child", shape)
			return nil
		}
		t.Cleanup(func() { codexDirectLaunchFn, codexSpawnLaunchFn = prevDirect, prevSpawn })

		_, errB, err := runCodexCmd(t, shape...)

		if errB != codexFactoryRefusalDiag+"\n" {
			t.Errorf("shape %v: stderr = %q, want exactly the one REQ-SD-004 refusal line %q", shape, errB, codexFactoryRefusalDiag)
		}
		for _, want := range []string{
			factoryUnsupportedBackendSentinel,
			"moai codex -l",
			"moai cc -f",
			"moai glm -f",
		} {
			if !strings.Contains(codexFactoryRefusalDiag, want) {
				t.Errorf("the refusal line does not name %q: %q", want, codexFactoryRefusalDiag)
			}
		}
		if loc := removedFormPattern.FindString(codexFactoryRefusalDiag); loc != "" {
			t.Errorf("the refusal line names the removed form %q: %q", loc, codexFactoryRefusalDiag)
		}
		code, ok := ResolveExitCode(err)
		if !ok || code != 1 {
			t.Errorf("shape %v: exit code = (%d, %v), want (1, true); err=%v", shape, code, ok, err)
		}
	}
}
