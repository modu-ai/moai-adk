package cli

// codex_factory_retire_test.go — SPEC-CODEX-FACTORY-RETIRE-001 M1.
//
// AC coverage in this file:
//   - AC-CFR-001 / AC-CFR-002 — `moai codex -k` / `-f` refusal lines
//   - AC-CFR-003 — the refusal has no state effect
//   - AC-CFR-004 — entry tokens after `--` pass through
//   - AC-CFR-005 / AC-CFR-006 — the eleven lane keys never reach a codex child
//   - AC-CFR-008 / AC-CFR-009 / AC-CFR-010 — cc/glm refuse a codex-led run
//   - AC-CFR-025 — codex-harness hooks register no factory peer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/kanban"
)

// laneKeyFixtureValues gives each of the eleven lane keys a distinct,
// non-empty value, so an assertion can tell which key leaked.
func laneKeyFixtureValues() map[string]string {
	values := map[string]string{}
	for i, key := range codexLaneLaunchEnvKeys {
		values[key] = "t1242-lane-value-" + strconv.Itoa(i)
	}
	return values
}

// TestCodexLaneLaunchEnvKeysAreTheElevenLaneKeys pins the scrub list to the
// eleven keys REQ-CFR-006 names, so a list that loses or gains a key fails here
// and not only in the per-key assertions below.
func TestCodexLaneLaunchEnvKeysAreTheElevenLaneKeys(t *testing.T) {
	want := []string{
		config.EnvMoaiKanban,
		config.EnvMoaiKanbanID,
		config.EnvMoaiKanbanSpec,
		config.EnvMoaiKanbanLabel,
		config.EnvMoaiKanbanLeadAddr,
		config.EnvMoaiKanbanLeadName,
		config.EnvMoaiKanbanBackend,
		config.EnvMoaiKanbanCard,
		config.EnvMoaiKanbanSettingsInjected,
		config.EnvMoaiFactoryWorker,
		config.EnvMoaiFactoryWorkers,
	}
	got := append([]string(nil), codexLaneLaunchEnvKeys...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("codexLaneLaunchEnvKeys = %v, want exactly %v", got, want)
	}
}

func codexRefusalCases(kind string) [][]string {
	if kind == "kanban" {
		return [][]string{
			{"-k"}, {"--kanban"}, {"-k", "SPEC-X-001"}, {"-k", "--name", "plan"},
			{"--kanban=SPEC-X-001"}, {"-k", "cli"}, {"-k", "status"}, {"-k", "-w", "t1"},
			{"-k", "--spawn"},
		}
	}
	return [][]string{
		{"-f"}, {"--factory"}, {"-f", "worker"}, {"-f", "worker-2"}, {"-f", "agent"},
		{"-f", "lane-3"}, {"--factory=worker"}, {"-f=worker-1"}, {"--factory-run", "r1"},
		{"-f", "app"}, {"-f", "status"},
	}
}

// pinCodexRefusalRoot points every project-root resolver at a fresh temp
// directory under an isolated MOAI_HOME, so an accepted entry cannot write
// state anywhere outside the test.
func pinCodexRefusalRoot(t *testing.T) string {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	withCodexProjectRoot(t, root)
	t.Setenv(config.EnvClaudeProjectDir, root)
	return root
}

func assertCodexRefused(t *testing.T, args []string, stdout, stderr string, err error, want ...string) {
	t.Helper()
	var exit *exitCodeError
	if !errors.As(err, &exit) || exit.code != 1 {
		t.Fatalf("codex %v: err = %v, want exit code 1", args, err)
	}
	if stdout != "" {
		t.Errorf("codex %v: stdout = %q, want empty", args, stdout)
	}
	lines := strings.Split(strings.TrimRight(stderr, "\n"), "\n")
	if len(lines) != 1 || lines[0] == "" {
		t.Fatalf("codex %v: stderr = %q, want exactly one line", args, stderr)
	}
	for _, w := range want {
		if !strings.Contains(lines[0], w) {
			t.Errorf("codex %v: stderr %q lacks %q", args, lines[0], w)
		}
	}
}

// AC-CFR-001 — every kanban entry shape is refused with the kanban sentinel.
func TestCodexKanbanEntryIsRefused(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	pinCodexRefusalRoot(t)
	for _, args := range codexRefusalCases("kanban") {
		stdout, stderr, err := runCodexCmd(t, args...)
		assertCodexRefused(t, args, stdout, stderr, err, kanbanUnsupportedBackendSentinel, "moai cc -k")
	}
	// Edge (acceptance §D.1): --name without -k stays the plain usage error.
	stdout, stderr, err := runCodexCmd(t, "--name", "plan")
	assertCodexRefused(t, []string{"--name", "plan"}, stdout, stderr, err, codexUsageDiag)
	if strings.Contains(stderr, kanbanUnsupportedBackendSentinel) {
		t.Errorf("--name plan carried the kanban sentinel: %q", stderr)
	}
	codexWantLaunches(t, cap, 0, 0, 0)
}

// AC-CFR-002 — every factory entry shape is refused with the factory sentinel.
func TestCodexFactoryEntryIsRefused(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	pinCodexRefusalRoot(t)
	for _, args := range codexRefusalCases("factory") {
		stdout, stderr, err := runCodexCmd(t, args...)
		assertCodexRefused(t, args, stdout, stderr, err, factoryUnsupportedBackendSentinel, "moai cc -f", "moai glm -f")
	}
	codexWantLaunches(t, cap, 0, 0, 0)
}

func fileDigest(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "absent"
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func laneEnvSnapshot() string {
	var keep []string
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "MOAI_KANBAN") || strings.HasPrefix(entry, "MOAI_FACTORY_") {
			keep = append(keep, entry)
		}
	}
	sort.Strings(keep)
	return strings.Join(keep, "\n")
}

// AC-CFR-003 — a refused entry writes no factory state, claims no slot, exports
// no lane variable and creates no worktree; the bare form still launches once.
func TestCodexEntryRefusalHasNoStateEffect(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	root := pinCodexRefusalRoot(t)
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(root, ".moai", "state", "factory", "workers.json")
	dbBefore, legacyBefore, envBefore := fileDigest(t, dbPath), fileDigest(t, legacy), laneEnvSnapshot()

	for _, kind := range []string{"kanban", "factory"} {
		for _, args := range codexRefusalCases(kind) {
			if _, _, err := runCodexCmd(t, args...); err == nil {
				t.Errorf("codex %v: accepted, want a refusal", args)
			}
		}
	}
	codexWantLaunches(t, cap, 0, 0, 0)
	if got := fileDigest(t, dbPath); got != dbBefore {
		t.Errorf("factory state DB changed: %s -> %s", dbBefore, got)
	}
	if got := fileDigest(t, legacy); got != legacyBefore {
		t.Errorf("worker registry file changed: %s -> %s", legacyBefore, got)
	}
	if got := laneEnvSnapshot(); got != envBefore {
		t.Errorf("lane environment changed:\nbefore %q\nafter  %q", envBefore, got)
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "worktrees")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a refusal created .claude/worktrees: %v", err)
	}

	// Positive control: the same harness records the bare launch.
	if _, stderr, err := runCodexCmd(t); err != nil {
		t.Fatalf("bare codex: %v (stderr %q)", err, stderr)
	}
	codexWantLaunches(t, cap, 1, 1, 0)
}

// AC-CFR-004 — entry tokens after `--` belong to codex and are not refused.
func TestCodexEntryTokensAfterDashDashPassThrough(t *testing.T) {
	cap := withCodexLaunchCapture(t)
	pinCodexRefusalRoot(t)
	if _, stderr, err := runCodexCmd(t, "--", "-f", "worker", "-k"); err != nil {
		t.Fatalf("codex -- -f worker -k: %v (stderr %q)", err, stderr)
	}
	codexWantLaunches(t, cap, 1, 1, 0)
	argv := cap.records[0].Argv
	if n := len(argv); n < 3 || strings.Join(argv[n-3:], " ") != "-f worker -k" {
		t.Fatalf("argv = %q, want a tail ending with -f worker -k", argv)
	}
}

func setLaneFixtureEnv(t *testing.T) map[string]string {
	t.Helper()
	values := laneKeyFixtureValues()
	for key, value := range values {
		t.Setenv(key, value)
	}
	t.Setenv("MOAI_HOME", t.TempDir())
	t.Setenv(config.EnvAutonomyTier, "t1242-tier")
	t.Setenv(config.EnvClaudeCodeMaxConcurrentSubagents, "7")
	t.Setenv("T1242_KEEP", "1")
	t.Setenv(config.EnvClaudeCodeSessionID, "foreign-claude-session")
	t.Setenv(config.EnvMoaiSessionPID, "12345")
	return values
}

// AC-CFR-005 — the direct child carries no lane identity, and keeps the
// posture keys, the sentinel, MOAI_HOME and the resolved CODEX_HOME.
func TestCodexChildEnvScrubsLaneKeys(t *testing.T) {
	setLaneFixtureEnv(t)
	env := codexChildEnv()
	for _, key := range codexLaneLaunchEnvKeys {
		t.Run(key, func(t *testing.T) {
			for _, entry := range env {
				k, v, _ := strings.Cut(entry, "=")
				if k == key && v != "" {
					t.Fatalf("%s=%q reached the Codex child", key, v)
				}
			}
		})
	}
	for _, key := range []string{"MOAI_HOME", config.EnvAutonomyTier, config.EnvClaudeCodeMaxConcurrentSubagents} {
		if v, ok := codexEnvLast(env, key); !ok || v != os.Getenv(key) {
			t.Errorf("%s = %q (present %v), want the inherited %q", key, v, ok, os.Getenv(key))
		}
	}
	if v, ok := codexEnvLast(env, "T1242_KEEP"); !ok || v != "1" {
		t.Errorf("T1242_KEEP = %q (present %v), want 1", v, ok)
	}
	if home, _ := resolveCodexHomeDir(); home != "" {
		if v, ok := codexEnvLast(env, codexHomeEnvVar); !ok || v != home {
			t.Errorf("%s = %q, want the resolved %q", codexHomeEnvVar, v, home)
		}
	} else {
		t.Error("resolveCodexHomeDir returned no home; the CODEX_HOME axis is unmeasured")
	}
	for _, key := range []string{config.EnvClaudeCodeSessionID, config.EnvMoaiSessionPID} {
		if _, ok := codexEnvLast(env, key); ok {
			t.Errorf("%s leaked into the Codex child", key)
		}
	}
}

// AC-CFR-006 — the spawn command blanks every lane key and forwards none of
// their values; MOAI_HOME is still forwarded.
func TestCodexSpawnCommandBlanksLaneKeys(t *testing.T) {
	values := setLaneFixtureEnv(t)
	tokens := strings.Fields(buildCodexSpawnCommand("/x/codex", nil))
	has := func(tok string) bool {
		for _, got := range tokens {
			if got == tok {
				return true
			}
		}
		return false
	}
	for _, key := range codexLaneLaunchEnvKeys {
		t.Run(key, func(t *testing.T) {
			if !has(key + "=") {
				t.Errorf("spawn command lacks the blank assignment %s=", key)
			}
			if has(key + "=" + shellQuote(values[key])) {
				t.Errorf("spawn command forwards %s=%s", key, values[key])
			}
		})
	}
	if want := "MOAI_HOME=" + shellQuote(os.Getenv("MOAI_HOME")); !has(want) {
		t.Errorf("spawn command lacks %s", want)
	}
}

// codexLedRun records one active factory run with the given lead backend and
// owner identity, and returns the project root it lives under.
func codexLedRun(t *testing.T, runID, backend string) string {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	// SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-005: a lane join needs a git
	// working tree, so the entry-test fixture is one.
	initGitRepo(t, root)
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{
		RunID: runID, Backend: backend, ManifestJSON: "{}", LeadPID: 424242, LeadProcessStart: "t1242-start",
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	// The first registry read of a project writes a one-time
	// legacy_workers_imported marker (homestate ImportLegacyWorkers). Any
	// project that already ran a factory lead carries it; take it here so a
	// digest taken later measures claims, not that one-time marker.
	_ = loadFactoryRegistry(factoryRegistryPath(root))
	clearFactoryTestEnv(t)
	for _, key := range codexLaneLaunchEnvKeys {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
	t.Setenv(config.EnvClaudeProjectDir, root)
	return root
}

func runRow(t *testing.T, root, runID string) (string, string, int) {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var status, backend string
	var pid int
	if err := db.DB.QueryRow(`SELECT status, lead_backend, lead_pid FROM runs WHERE run_id=?`, runID).Scan(&status, &backend, &pid); err != nil {
		t.Fatal(err)
	}
	return status, backend, pid
}

func workerRows(t *testing.T, root string) int {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM workers`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// driveFactoryEntry runs one cc or glm entry through its launch seam and
// reports the error and the number of launches the seam saw.
func driveFactoryEntry(t *testing.T, launcher string, args ...string) (int, error) {
	t.Helper()
	launches := 0
	c := &cobra.Command{Use: launcher}
	c.SetContext(context.Background())
	c.SetOut(new(strings.Builder))
	c.SetErr(new(strings.Builder))
	if launcher == "glm" {
		prev := unifiedLaunchFunc
		unifiedLaunchFunc = func(string, string, []string) error { launches++; return nil }
		t.Cleanup(func() { unifiedLaunchFunc = prev })
		err := runGLM(c, args)
		return launches, err
	}
	err := runClaudeEntry(c, args, "cc", "claude", kanban.BackendClaude, func(string, string, []string) error {
		launches++
		return nil
	})
	return launches, err
}

// AC-CFR-008 — a lane join into a codex-led run is refused before any claim.
func TestFactoryJoinRefusesCodexLedRun(t *testing.T) {
	for _, tc := range []struct {
		launcher string
		args     []string
	}{
		{"cc", []string{"-f", "lane"}},
		{"cc", []string{"-f", "lane-2"}},
		{"glm", []string{"-f", "lane"}},
	} {
		t.Run(tc.launcher+" "+strings.Join(tc.args, " "), func(t *testing.T) {
			root := codexLedRun(t, "rc", "codex")
			dbPath, err := homestate.FactoryDBPath(root)
			if err != nil {
				t.Fatal(err)
			}
			// Count first: opening the DB may checkpoint it, so the digest is
			// taken after the last test-side open and before the entry runs.
			rowsBefore := workerRows(t, root)
			before := fileDigest(t, dbPath)
			launches, err := driveFactoryEntry(t, tc.launcher, tc.args...)
			if err == nil {
				t.Fatal("join accepted, want a refusal")
			}
			for _, w := range []string{"rc", "codex", "moai factory runs --retire"} {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("refusal %q lacks %q", err.Error(), w)
				}
			}
			if launches != 0 {
				t.Errorf("launches = %d, want 0", launches)
			}
			if got := fileDigest(t, dbPath); got != before {
				t.Errorf("worker registry file changed: %s -> %s", before, got)
			}
			if got := workerRows(t, root); got != rowsBefore {
				t.Errorf("worker rows %d -> %d, want no claim", rowsBefore, got)
			}
		})
	}
}

// AC-CFR-009 — a claude-led run is joined and led exactly as before.
func TestFactoryJoinAndLeadAcceptNonCodexRun(t *testing.T) {
	t.Run("lane", func(t *testing.T) {
		root := codexLedRun(t, "rc", "claude")
		rowsBefore := workerRows(t, root)
		launches, err := driveFactoryEntry(t, "cc", "-f", "lane")
		if err != nil {
			t.Fatalf("join refused: %v", err)
		}
		if launches != 1 {
			t.Errorf("launches = %d, want 1", launches)
		}
		if got := workerRows(t, root); got != rowsBefore+1 {
			t.Errorf("worker rows %d -> %d, want one claimed slot", rowsBefore, got)
		}
	})
	t.Run("lead", func(t *testing.T) {
		root := codexLedRun(t, "rc", "claude")
		launches, err := driveFactoryEntry(t, "cc", "-f", "--factory-run", "rc")
		if err != nil {
			t.Fatalf("lead refused: %v", err)
		}
		if launches != 1 {
			t.Errorf("launches = %d, want 1", launches)
		}
		if status, backend, _ := runRow(t, root, "rc"); status != "active" || backend != "claude" {
			t.Errorf("run rc = (%s, %s), want (active, claude)", status, backend)
		}
	})
}

// AC-CFR-010 — a claude lead cannot adopt a codex-led run.
func TestFactoryLeadRefusesCodexLedRun(t *testing.T) {
	root := codexLedRun(t, "rc", "codex")
	_, backendBefore, pidBefore := runRow(t, root, "rc")
	launches, err := driveFactoryEntry(t, "cc", "-f", "--factory-run", "rc")
	if err == nil {
		t.Fatal("lead adopted a codex run, want a refusal")
	}
	for _, w := range []string{"rc", "codex", "moai factory runs --retire"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("refusal %q lacks %q", err.Error(), w)
		}
	}
	if launches != 0 {
		t.Errorf("launches = %d, want 0", launches)
	}
	if _, backend, pid := runRow(t, root, "rc"); backend != backendBefore || pid != pidBefore {
		t.Errorf("run rc = (%s, %d), want unchanged (%s, %d)", backend, pid, backendBefore, pidBefore)
	}
}

// refuseCodexLeaderRun fails closed when the selected run's lead backend cannot
// be read, rather than letting an unread run be joined as if it were claude-led.
func TestRefuseCodexLedRunFailsClosedOnUnreadableRun(t *testing.T) {
	root := codexLedRun(t, "rc", "claude")
	if err := refuseCodexLeaderRun(root, "rc"); err != nil {
		t.Fatalf("claude-led run refused: %v", err)
	}
	err := refuseCodexLeaderRun(root, "no-such-run")
	if err == nil || !strings.Contains(err.Error(), "no-such-run") {
		t.Fatalf("missing run = %v, want a read error naming the run", err)
	}
}

// ─── AC-CFR-025 — codex-harness hooks register no factory peer ─────────────

// realHandlerRegistry dispatches the two events that bind a factory peer to
// the production handlers, so the peer rows this test counts are the rows a
// real hook run writes.
type realHandlerRegistry struct{ cfg hook.ConfigProvider }

func (r *realHandlerRegistry) Register(hook.Handler)                  {}
func (r *realHandlerRegistry) Handlers(hook.EventType) []hook.Handler { return nil }

func (r *realHandlerRegistry) Dispatch(ctx context.Context, event hook.EventType, input *hook.HookInput) (*hook.HookOutput, error) {
	switch event {
	case hook.EventSessionStart:
		// WithSynchronousDeferredScans: this test owns the project dir via
		// t.TempDir; deferred scans must not outlive the test body.
		return hook.NewSessionStartHandler(r.cfg, hook.WithSynchronousDeferredScans()).Handle(ctx, input)
	case hook.EventUserPromptSubmit:
		return hook.NewUserPromptSubmitHandler(r.cfg).Handle(ctx, input)
	}
	return &hook.HookOutput{}, nil
}

func runFactoryHook(t *testing.T, subcommand, harness string, input *hook.HookInput) {
	t.Helper()
	origDeps := deps
	deps = &Dependencies{
		HookRegistry: &realHandlerRegistry{cfg: config.NewConfigManager()},
		HookProtocol: &codexHarnessProtocol{input: input},
	}
	t.Cleanup(func() { deps = origDeps })
	var sub *cobra.Command
	for _, c := range hookCmd.Commands() {
		if c.Name() == subcommand {
			sub = c
		}
	}
	if sub == nil {
		t.Fatalf("hook subcommand %q not found", subcommand)
	}
	flags := []string{}
	if harness != "" {
		flags = []string{"--harness", harness}
	}
	if err := sub.ParseFlags(flags); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sub.Flags().Set("harness", "") })
	defer func() { _ = sub.Flags().Set("harness", "") }()
	sub.SetContext(context.Background())
	var runErr error
	_ = captureStdoutDuring(t, func() { runErr = sub.RunE(sub, []string{}) })
	if runErr != nil {
		t.Fatalf("hook %s --harness %q: %v", subcommand, harness, runErr)
	}
}

func peerCount(t *testing.T, root, runID, slot string) int {
	t.Helper()
	s, err := factorymsg.Open(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, lane := range status.Lanes {
		if lane.Slot == slot {
			n++
		}
	}
	return n
}

func TestCodexHarnessHooksRegisterNoFactoryPeer(t *testing.T) {
	for _, shape := range []struct {
		name string
		slot string
		env  map[string]string
	}{
		{"lane", "lane-1", map[string]string{config.EnvMoaiFactoryWorker: "lane-1"}},
		{"leader", "leader", map[string]string{config.EnvMoaiFactoryWorkers: "2"}},
	} {
		t.Run(shape.name, func(t *testing.T) {
			root := codexLedRun(t, "r1", "claude")
			t.Setenv(config.EnvMoaiKanbanID, "r1")
			t.Setenv(config.EnvMoaiKanbanBackend, "claude")
			for key, value := range shape.env {
				t.Setenv(key, value)
			}
			t.Setenv(config.EnvMoaiSessionPID, strconv.Itoa(os.Getpid()))
			in := func(event, session string) *hook.HookInput {
				return &hook.HookInput{HookEventName: event, SessionID: session, ProjectDir: root, CWD: root, Prompt: "t1242 probe"}
			}

			runFactoryHook(t, "session-start", "codex", in("SessionStart", "codex-"+shape.name))
			runFactoryHook(t, "user-prompt-submit", "codex", in("UserPromptSubmit", "codex-"+shape.name))
			if got := peerCount(t, root, "r1", shape.slot); got != 0 {
				t.Fatalf("after --harness codex: peers for (r1, %s) = %d, want 0", shape.slot, got)
			}
			if os.Getenv(config.EnvMoaiKanbanID) != "r1" {
				t.Fatalf("the codex-harness hook left %s changed in this process", config.EnvMoaiKanbanID)
			}

			runFactoryHook(t, "session-start", "", in("SessionStart", "claude-"+shape.name))
			runFactoryHook(t, "user-prompt-submit", "", in("UserPromptSubmit", "claude-"+shape.name))
			if got := peerCount(t, root, "r1", shape.slot); got != 1 {
				t.Fatalf("after the Claude hooks (positive control): peers for (r1, %s) = %d, want 1", shape.slot, got)
			}
		})
	}
}
