package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// runOwnerSandbox returns a project root OUTSIDE this repository's worktree set
// with HOME / MOAI_HOME / MOAI_CLAUDE_BIN scrubbed to it (REQ-012).
func runOwnerSandbox(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MOAI_HOME", filepath.Join(home, ".moai"))
	t.Setenv("MOAI_CLAUDE_BIN", filepath.Join(home, "bin", "claude"))
	root := filepath.Join(t.TempDir(), "sandbox-project")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatalf("create sandbox project root: %v", err)
	}
	return root
}

func seedRun(t *testing.T, root, runID string, pid int, start string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: "claude", ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: start,
	}); err != nil {
		t.Fatalf("record run: %v", err)
	}
}

func seedLeadPeer(t *testing.T, root, runID string, pid int, start string) {
	t.Helper()
	s, err := factorymsg.Open(root, runID)
	if err != nil {
		t.Fatalf("open broker: %v", err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.RegisterLaunchPending(context.Background(), factorymsg.Peer{
		ProjectKey: homestate.ProjectKey(root), RunID: runID, Backend: "claude",
		Role: "lead", Slot: "lead", Generation: 1, PID: pid, ProcessStart: start,
	}); err != nil {
		t.Fatalf("register lead peer: %v", err)
	}
}

func runOwnerStamp(t *testing.T, root, runID string) (int, string) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start FROM runs WHERE run_id=?`, runID).Scan(&pid, &start); err != nil {
		t.Fatalf("read owner stamp: %v", err)
	}
	return pid, start
}

// AC-016 leg 1 — for every launch shape, the run row's owner identity and the
// run's role='lead' peer identity name ONE process. The three shapes are
// driven through the build-tag-free seam with fixture identities, so this runs
// on any host: no Windows host and no live tmux server required.
func TestRunOwnerAndLeadPeerNameOneProcessOnEveryShape(t *testing.T) {
	const launcherPID, sessionPID = 90001, 90002
	const launcherStart, sessionStart = "launcher-start", "session-start"

	t.Run("replace (no restamp needed)", func(t *testing.T) {
		root := runOwnerSandbox(t)
		// syscall.Exec preserves pid and start time, so the record-time stamp
		// already IS the session identity.
		seedRun(t, root, "run-replace", launcherPID, launcherStart)
		seedLeadPeer(t, root, "run-replace", launcherPID, launcherStart)
		assertStampMatchesPeer(t, root, "run-replace")
	})

	for _, shape := range []string{"spawn", "pane"} {
		t.Run(shape+" (restamped to the session)", func(t *testing.T) {
			root := runOwnerSandbox(t)
			runID := "run-" + shape
			// The record-time stamp names the launcher — correct only for the
			// instant it describes.
			seedRun(t, root, runID, launcherPID, launcherStart)
			seedLeadPeer(t, root, runID, sessionPID, sessionStart)

			if pid, _ := runOwnerStamp(t, root, runID); pid != launcherPID {
				t.Fatalf("pre-restamp owner pid = %d, want the launcher %d", pid, launcherPID)
			}
			if err := stampFactoryRunOwner(root, runID, sessionPID, sessionStart); err != nil {
				t.Fatalf("restamp: %v", err)
			}
			assertStampMatchesPeer(t, root, runID)
		})
	}
}

func assertStampMatchesPeer(t *testing.T, root, runID string) {
	t.Helper()
	pid, start := runOwnerStamp(t, root, runID)
	peerPID, peerStart, ok := factorymsg.LeadPeerIdentity(root, runID)
	if !ok {
		t.Fatalf("%s: no lead peer identity", runID)
	}
	if pid != peerPID || start != peerStart {
		t.Fatalf("%s: run row names (%d, %q) but the lead peer names (%d, %q) — the two sources must name one process",
			runID, pid, start, peerPID, peerStart)
	}
}

// AC-016 leg 2 — the seam is actually CALLED at every non-replace call site.
// Driving the seam with fixture identities proves the seam behaves; it says
// nothing about whether each door reaches it. A door that never calls the seam
// leaves its run stamped with the launching process and passes leg 1 unchanged.
//
// The assertion is source-level because two of the three sites sit behind
// //go:build windows: a darwin host can neither execute them nor compile a call
// to them. That host constraint is what kept the gap invisible, so this leg is
// written to hold without it.
//
// Mutation direction, ONE SITE AT A TIME: delete the seam call at any single
// site and this test turns red on that site's own subtest. An assertion that
// only went red when all three were missing would stay green while two sites
// are correct and one is silently not.
func TestRestampSeamIsCalledAtEveryNonReplaceCallSite(t *testing.T) {
	const seamCall = "stampFactoryRunOwner("

	required := []string{
		"launch_exec_windows.go",  // spawn
		"codex_direct_windows.go", // spawn
		"codex_launcher.go",       // pane
	}
	for _, name := range required {
		t.Run("required/"+name, func(t *testing.T) {
			if !strings.Contains(readCallSite(t, name), seamCall) {
				t.Fatalf("%s does not call %s — this door leaves its run stamped with the launching process", name, seamCall)
			}
		})
	}

	// A replace-shaped door needs no restamp, and demanding one there would
	// make this leg false about the design it is checking.
	for _, name := range []string{"launch_exec_posix.go", "codex_direct_posix.go"} {
		t.Run("absent/"+name, func(t *testing.T) {
			if strings.Contains(readCallSite(t, name), seamCall) {
				t.Fatalf("%s calls %s, but a replace-shaped door has nothing to correct", name, seamCall)
			}
		})
	}
}

func readCallSite(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(raw)
}

// AC-012 — when a door that does not replace its launching process cannot
// obtain a session identity, the launch is refused AND no runs row is left
// carrying the launching process's pid. A refusal that cleans up the pane but
// leaves the run stamped fails this criterion.
func TestPaneDoorRefusalLeavesNoLauncherStampedRun(t *testing.T) {
	root := runOwnerSandbox(t)
	const runID = "run-pane-refusal"
	launcherPID := os.Getpid()
	seedRun(t, root, runID, launcherPID, "launcher-start")

	t.Setenv(config.EnvMoaiKanbanID, runID)
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvMoaiKanbanBackend, "codex")

	restoreSpawn := tmuxSpawnFn
	restoreIdentity := codexSpawnPaneIdentityFn
	restoreCleanup := codexSpawnCleanupPaneFn
	t.Cleanup(func() {
		tmuxSpawnFn = restoreSpawn
		codexSpawnPaneIdentityFn = restoreIdentity
		codexSpawnCleanupPaneFn = restoreCleanup
	})

	tmuxSpawnFn = func(string, string) (string, error) { return "%42", nil }
	codexSpawnPaneIdentityFn = func(string) (int, string, error) {
		return 0, "", errors.New("spawned Codex pane process identity unavailable")
	}
	cleaned := false
	codexSpawnCleanupPaneFn = func(string) error { cleaned = true; return nil }

	err := defaultCodexSpawnLaunch(root, "codex", []string{"--version"})
	if err == nil {
		t.Fatal("launch succeeded, want a refusal")
	}
	if !cleaned {
		t.Fatal("the pane was not cleaned up on refusal")
	}
	pid, start := runOwnerStamp(t, root, runID)
	if pid == launcherPID {
		t.Fatalf("run row still carries the launching process's pid %d (start %q); that identity is known in advance to die", pid, start)
	}
}

// REQ-002d on the anchor-refusal path — in factory mode, when the spawned
// pane's identity resolves but the worktree anchor lock is refused (a second
// writer, or a tree already anchored), the launch is refused with the anchor
// error AND the run row no longer carries the launching process's identity.
// The refusal returns before the pane restamp, so without the owner clear on
// this path the launcher's soon-dead pid would survive on the run.
func TestPaneDoorAnchorRefusalClearsFactoryRunOwner(t *testing.T) {
	root := runOwnerSandbox(t)
	const runID = "run-pane-anchor-refusal"
	launcherPID := os.Getpid()
	seedRun(t, root, runID, launcherPID, "launcher-start")
	if pid, start := runOwnerStamp(t, root, runID); pid != launcherPID || start != "launcher-start" {
		t.Fatalf("precondition: seeded owner = (%d, %q), want (%d, %q)", pid, start, launcherPID, "launcher-start")
	}

	t.Setenv(config.EnvMoaiKanbanID, runID)
	t.Setenv(config.EnvMoaiFactoryWorkers, "2")
	t.Setenv(config.EnvMoaiKanbanBackend, "codex")

	restoreSpawn := tmuxSpawnFn
	restoreIdentity := codexSpawnPaneIdentityFn
	restoreCleanup := codexSpawnCleanupPaneFn
	restoreAnchor := codexSpawnAnchorFn
	t.Cleanup(func() {
		tmuxSpawnFn = restoreSpawn
		codexSpawnPaneIdentityFn = restoreIdentity
		codexSpawnCleanupPaneFn = restoreCleanup
		codexSpawnAnchorFn = restoreAnchor
	})

	tmuxSpawnFn = func(string, string) (string, error) { return "%43", nil }
	codexSpawnPaneIdentityFn = func(string) (int, string, error) { return 90002, "pane-start", nil }
	cleaned := false
	codexSpawnCleanupPaneFn = func(string) error { cleaned = true; return nil }
	errAnchor := errors.New("codex worktree already anchored by a live writer")
	anchorCalls := 0
	codexSpawnAnchorFn = func(int, string) error { anchorCalls++; return errAnchor }

	err := defaultCodexSpawnLaunch(root, "codex", []string{"--version"})
	if !errors.Is(err, errAnchor) {
		t.Fatalf("launch error = %v, want the anchor refusal", err)
	}
	if anchorCalls != 1 || !cleaned {
		t.Fatalf("anchor calls = %d, pane cleaned = %v; want the anchor-refusal path taken once with pane cleanup", anchorCalls, cleaned)
	}
	if pid, start := runOwnerStamp(t, root, runID); pid != 0 || start != "" {
		t.Fatalf("run owner after anchor refusal = (%d, %q), want cleared (0, \"\"); the launcher's identity is known in advance to die", pid, start)
	}
}
