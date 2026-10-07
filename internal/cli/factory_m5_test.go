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
	"github.com/modu-ai/moai-adk/internal/factory"
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

// AC-SD-003 (card t1554) — `moai codex -l` starts ONE lane session: the
// boot auto-lease loop is removed, so the launcher leases nothing and the
// session runs in the parent checkout carrying the marker, the lane label,
// and the Codex backend value, with the `moai todo --auto` initiation prompt
// as its directive — the session consumes the queue itself.
func TestSD_AC003_CodexLaneSessionOneShot(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	sdRecordLeaderRun(t, root, fcRun, factory.BackendClaude)
	wantLabel := sdNextFreeLaneLabel(t, root)
	t.Chdir(root)
	sdScrubLauncherEnv(t)

	var got []sdCardLaunchCapture
	prevLook := codexLookPath
	codexLookPath = func(string) (string, error) { return "/sentinel/codex", nil }
	prevDirect := codexDirectLaunchFn
	codexDirectLaunchFn = func(c *exec.Cmd) error {
		got = append(got, sdCardLaunchCapture{
			argv: append([]string(nil), c.Args...),
			dir:  c.Dir,
			env:  sdEnvOf(t, c.Env),
		})
		return nil
	}
	t.Cleanup(func() { codexLookPath, codexDirectLaunchFn = prevLook, prevDirect })

	if _, _, err := runCodexCmd(t, "-l"); err != nil {
		t.Fatalf("codex lane: %v", err)
	}

	// The launcher's root resolves through the queue-root resolver, which
	// evaluates symlinks (/var → /private/var on darwin temp dirs) — compare
	// the evaluated spelling.
	primary, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("evaluate fixture root: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("the launcher started %d sessions, want 1 (the lane session)", len(got))
	}
	rec := got[0]
	// The session runs in the parent checkout — the lease machinery's own
	// precondition (REQ-SD-010) — and argv carries it through -C.
	if !sameDirPath(rec.dir, primary) {
		t.Errorf("session working directory = %q, want the parent checkout %q", rec.dir, primary)
	}
	if !containsPair(rec.argv, "-C", primary) {
		t.Errorf("argv %v does not carry -C with the parent checkout", rec.argv)
	}
	// The --auto initiation prompt is the session's directive.
	prompted := false
	for _, a := range rec.argv {
		if strings.Contains(a, "moai todo --auto") {
			prompted = true
		}
	}
	if !prompted {
		t.Errorf("argv %v carries no `moai todo --auto` initiation prompt", rec.argv)
	}
	if rec.env[config.EnvFactoryRole] != config.FactoryRoleLane {
		t.Errorf("session env %s = %q, want the value constant %q", config.EnvFactoryRole, rec.env[config.EnvFactoryRole], config.FactoryRoleLane)
	}
	if rec.env[config.EnvMoaiFactoryWorker] != wantLabel {
		t.Errorf("session env %s = %q, want the lane label %q", config.EnvMoaiFactoryWorker, rec.env[config.EnvMoaiFactoryWorker], wantLabel)
	}
	if rec.env[config.EnvFactoryBackend] != factory.BackendGPT {
		t.Errorf("session env %s = %q, want the Codex harness value %q", config.EnvFactoryBackend, rec.env[config.EnvFactoryBackend], factory.BackendGPT)
	}
	// The card-identifier variable stays unset: no card is leased at boot.
	if rec.env[config.EnvFactoryCard] != "" {
		t.Errorf("session env carries %s=%q; the launcher leases nothing", config.EnvFactoryCard, rec.env[config.EnvFactoryCard])
	}
	// The queue is untouched: no lease, no record rows.
	for _, cardID := range []string{"t1", "t2"} {
		if fcHasCard(t, root, cardID) {
			t.Errorf("%s gained a factory record row; the launcher must not lease", cardID)
		}
		if s := nmQueueState(t, store, cardID); s != factory.BacklogStatePicked {
			t.Errorf("%s queue state = %s, want picked (the seeded operator pick stands)", cardID, s)
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
		// {"--factory-run", "x"} left this table (card t1444 ②): the token is
		// no longer classified as an other-factory shape — it travels to
		// parseCodexFactoryEntry, whose "--factory-run requires -l/--lane"
		// refusal TestCodexFactoryRunWithoutLaneStillRefused pins.
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
