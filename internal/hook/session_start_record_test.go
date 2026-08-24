package hook

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
)

// scrubKanbanEnv clears every launch-fact discriminator before a test sets the
// ones it means to exercise.
//
// It is load-bearing rather than tidy: these tests frequently RUN INSIDE a
// kanban or factory session, whose own launch variables are inherited by
// `go test`. Without the scrub a lane-8 session's MOAI_FACTORY_WORKER makes
// every case here read as lane 8 — a failure that reproduces only on the
// machine running the lane and never in CI.
func scrubKanbanEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		config.EnvMoaiFactoryWorker,
		config.EnvMoaiFactoryWorkers,
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanSpec,
		config.EnvMoaiKanbanBackend,
		config.EnvMoaiKanbanCard,
	} {
		t.Setenv(key, "")
	}
}

// stubToplevel returns a worktree-root resolver that always answers root.
func stubToplevel(root string) func(string) (string, error) {
	return func(string) (string, error) { return root, nil }
}

func failingToplevel() func(string) (string, error) {
	return func(string) (string, error) { return "", errors.New("not a repository") }
}

// AC-KRS-001 / AC-KRS-002: the record is keyed by the identifier the session's
// own runtime delivered, even when the project-wide single slot holds a
// different one. The pre-change writer produced the slot's identifier by
// construction, so this fixture is what separates the two.
func TestRecordIsKeyedByTheRuntimeSessionIDNotTheSidecar(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// T — the identifier the single slot holds, belonging to another session.
	const other = "T-11111111-2222-3333-4444-555555555555"
	if err := os.WriteFile(filepath.Join(root, session.CurrentSideChannelFile), []byte(other), 0o600); err != nil {
		t.Fatalf("seed sidecar: %v", err)
	}

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "run")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)
	t.Setenv(config.EnvMoaiKanbanSpec, "SPEC-EXAMPLE-001")

	const own = "S-99999999-8888-7777-6666-555555555555"
	writeKanbanSessionRecord(&HookInput{
		SessionID:  own,
		ProjectDir: root,
		CWD:        root,
		Source:     "startup",
	}, failingToplevel())

	rec, err := kanban.Read(root, own)
	if err != nil {
		t.Fatalf("no record under the session's own identifier: %v", err)
	}
	if rec.SessionID != own {
		t.Fatalf("session_id = %q, want %q", rec.SessionID, own)
	}
	if rec.Role != "run" {
		t.Fatalf("role = %q, want %q", rec.Role, "run")
	}
	if rec.Backend != kanban.BackendClaude {
		t.Fatalf("backend = %q, want %q", rec.Backend, kanban.BackendClaude)
	}
	if rec.SpecID != "SPEC-EXAMPLE-001" {
		t.Fatalf("spec_id = %q, want SPEC-EXAMPLE-001", rec.SpecID)
	}
	if _, statErr := os.Stat(kanban.RecordPath(root, other)); statErr == nil {
		t.Fatalf("a record was created under the sidecar's identifier %q", other)
	}
}

// AC-KRS-004: a factory lane records its number as data beside the role; a
// lead records 0.
func TestFactoryLaneRecordsItsNumberAndLeadRecordsZero(t *testing.T) {
	t.Run("lane-3", func(t *testing.T) {
		root := t.TempDir()
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiFactoryWorker, kanban.FactoryLaneLabel(3))
		t.Setenv(config.EnvMoaiFactoryWorkers, "4")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendGLM)

		writeKanbanSessionRecord(&HookInput{SessionID: "lane-sess", ProjectDir: root, CWD: root}, failingToplevel())

		rec, err := kanban.Read(root, "lane-sess")
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if rec.Lane != 3 {
			t.Fatalf("lane = %d, want 3", rec.Lane)
		}
		if rec.Role != kanban.RoleLane {
			t.Fatalf("role = %q, want %q", rec.Role, kanban.RoleLane)
		}
		if rec.Backend != kanban.BackendGLM {
			t.Fatalf("backend = %q, want %q", rec.Backend, kanban.BackendGLM)
		}
	})

	t.Run("kanban lead", func(t *testing.T) {
		root := t.TempDir()
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanban, "1")
		t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

		writeKanbanSessionRecord(&HookInput{SessionID: "lead-sess", ProjectDir: root, CWD: root}, failingToplevel())

		rec, err := kanban.Read(root, "lead-sess")
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if rec.Lane != 0 {
			t.Fatalf("lead lane = %d, want 0", rec.Lane)
		}
		if rec.Role != kanban.RoleLead {
			t.Fatalf("role = %q, want %q", rec.Role, kanban.RoleLead)
		}
	})
}

// AC-KRS-005, all three halves: the derivation from a card worktree, the
// override, and the primary-checkout case that must NOT yield a card.
func TestCardIdentifierDerivation(t *testing.T) {
	cardWorktree := filepath.Join("/Users/goos/MoAI/moai-adk-go", ".claude", "worktrees", "t207")
	primaryCheckout := filepath.Join("/Users/goos/moai", "moai-adk-go")

	t.Run("(a) card worktree, no override", func(t *testing.T) {
		root := t.TempDir()
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanban, "1")
		t.Setenv(config.EnvMoaiKanbanCard, "")

		writeKanbanSessionRecord(&HookInput{SessionID: "a", ProjectDir: root, CWD: root}, stubToplevel(cardWorktree))

		rec, err := kanban.Read(root, "a")
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if rec.CardID != "t207" {
			t.Fatalf("card_id = %q, want t207", rec.CardID)
		}
	})

	t.Run("(b) override wins", func(t *testing.T) {
		root := t.TempDir()
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanban, "1")
		t.Setenv(config.EnvMoaiKanbanCard, "t999")

		writeKanbanSessionRecord(&HookInput{SessionID: "b", ProjectDir: root, CWD: root}, stubToplevel(cardWorktree))

		rec, err := kanban.Read(root, "b")
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if rec.CardID != "t999" {
			t.Fatalf("card_id = %q, want t999", rec.CardID)
		}
	})

	t.Run("(c) primary checkout yields no card", func(t *testing.T) {
		root := t.TempDir()
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanban, "1")
		t.Setenv(config.EnvMoaiKanbanCard, "")

		writeKanbanSessionRecord(&HookInput{SessionID: "c", ProjectDir: root, CWD: root}, stubToplevel(primaryCheckout))

		raw, err := os.ReadFile(kanban.RecordPath(root, "c"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if got := string(raw); strings.Contains(got, `"card_id"`) {
			t.Fatalf("card_id key present for a primary checkout:\n%s", got)
		}
		rec, err := kanban.Read(root, "c")
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if rec.CardID == "moai-adk-go" {
			t.Fatalf("card_id took the checkout basename — the containment test did not apply")
		}
	})
}

// The same derivation against a REAL git worktree root rather than a stub, so
// the production resolver (resolveWorktreeRepoRoot) is exercised end to end.
func TestCardIdentifierFromARealGitWorktreeRoot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")

	base := t.TempDir()
	card := filepath.Join(base, ".claude", "worktrees", "t207")
	if err := os.MkdirAll(card, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if out, err := exec.Command("git", "-C", card, "init", "-q").CombinedOutput(); err != nil {
		t.Skipf("git init unavailable: %v (%s)", err, out)
	}

	root := t.TempDir()
	writeKanbanSessionRecord(&HookInput{SessionID: "real-wt", ProjectDir: root, CWD: card}, resolveWorktreeRepoRoot)

	rec, err := kanban.Read(root, "real-wt")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if rec.CardID != "t207" {
		t.Fatalf("card_id = %q, want t207", rec.CardID)
	}
}

// The label-to-role parse that moved here from the launcher when the launcher
// stopped writing the record: a bare label and its bumped `<role>-<n>` form
// both resolve to the bare role, and a malformed label yields no record at all
// rather than a guessed role.
func TestCompanionLabelResolvesToItsBareRole(t *testing.T) {
	for label, want := range map[string]string{"plan": "plan", "plan-2": "plan"} {
		scrubKanbanEnv(t)
		t.Setenv(config.EnvMoaiKanbanLabel, label)
		role, lane, ok := kanbanRoleFromEnv()
		if !ok || role != want || lane != 0 {
			t.Fatalf("kanbanRoleFromEnv() for label %q = (%q, %d, %v), want (%q, 0, true)", label, role, lane, ok, want)
		}
	}

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanbanLabel, "not a role at all")
	if _, _, ok := kanbanRoleFromEnv(); ok {
		t.Fatalf("a malformed companion label resolved to a role")
	}
}

// The containment test itself, exercised directly over both live shapes.
func TestCardIDFromWorktreeRootRequiresAWorktreesParent(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t207": "t207",
		"/Users/goos/moai/moai-adk-go":                        "",
		"/Users/goos/.moai/worktrees/card-wtiso":              "card-wtiso",
		"/":                                                   "",
		"":                                                    "",
	}
	for root, want := range cases {
		if got := cardIDFromWorktreeRoot(root); got != want {
			t.Fatalf("cardIDFromWorktreeRoot(%q) = %q, want %q", root, got, want)
		}
	}
}

// §E edge case: a session that is neither a kanban nor a factory session gets
// no record. The absence is the correct answer, not a degraded one.
func TestNonKanbanSessionWritesNoRecord(t *testing.T) {
	root := t.TempDir()
	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "")
	t.Setenv(config.EnvMoaiKanbanLabel, "")
	t.Setenv(config.EnvMoaiFactoryWorker, "")
	t.Setenv(config.EnvMoaiFactoryWorkers, "")

	writeKanbanSessionRecord(&HookInput{SessionID: "plain", ProjectDir: root, CWD: root}, failingToplevel())

	if _, err := os.Stat(kanban.RecordPath(root, "plain")); err == nil {
		t.Fatalf("a record was written for a non-kanban session")
	}
}

// A re-entry (resume / clear / compact / fork) must not rewrite the record:
// the orchestrator fills DeepScanDir, VerifyRung, and VerifyReentries in
// later, and a rewrite would discard them.
func TestReEntrySourcesDoNotWriteOrOverwrite(t *testing.T) {
	for _, source := range []string{"resume", "clear", "compact", "fork"} {
		t.Run(source, func(t *testing.T) {
			root := t.TempDir()
			scrubKanbanEnv(t)
			t.Setenv(config.EnvMoaiKanban, "1")
			t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

			writeKanbanSessionRecord(&HookInput{SessionID: "s", ProjectDir: root, CWD: root, Source: source}, failingToplevel())

			if _, err := os.Stat(kanban.RecordPath(root, "s")); err == nil {
				t.Fatalf("source %q wrote a record", source)
			}
		})
	}
}

func TestExistingRecordIsNotClobbered(t *testing.T) {
	root := t.TempDir()
	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	seeded := kanban.NewRecord("s", "SPEC-SEED", kanban.BackendGLM).WithRole(kanban.RoleLead)
	seeded.DeepScanDir = "/tmp/scan"
	seeded.VerifyReentries = 2
	if err := kanban.Write(root, seeded); err != nil {
		t.Fatalf("seed: %v", err)
	}

	writeKanbanSessionRecord(&HookInput{SessionID: "s", ProjectDir: root, CWD: root, Source: "startup"}, failingToplevel())

	rec, err := kanban.Read(root, "s")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if rec.DeepScanDir != "/tmp/scan" || rec.VerifyReentries != 2 {
		t.Fatalf("orchestrator-written fields were clobbered: %+v", rec)
	}
}

// AC-KRS-008: the write path IS reached, its attempt fails, and the session
// start is unaffected — no panic, no error, no record.
func TestRecordWriteFailsOpenOnUnwritableStateDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX directory permissions do not apply on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}

	root := t.TempDir()
	stateDir := filepath.Join(root, ".moai", "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.Chmod(stateDir, 0o500); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(stateDir, 0o755) })

	scrubKanbanEnv(t)
	t.Setenv(config.EnvMoaiKanban, "1")
	t.Setenv(config.EnvMoaiKanbanBackend, kanban.BackendClaude)

	// The path is reached (the environment marks a kanban session) and the
	// write attempt fails inside WriteBestEffort, which discards the error.
	writeKanbanSessionRecord(&HookInput{SessionID: "unwritable", ProjectDir: root, CWD: root, Source: "startup"}, failingToplevel())

	if _, err := os.Stat(kanban.RecordPath(root, "unwritable")); err == nil {
		t.Fatalf("a record exists though the state directory was unwritable")
	}
}
