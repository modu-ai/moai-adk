package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/modu-ai/moai-adk/internal/session"
)

func TestCCCmd_Exists(t *testing.T) {
	if ccCmd == nil {
		t.Fatal("ccCmd should not be nil")
	}
}

func TestCCCmd_Use(t *testing.T) {
	if !strings.HasPrefix(ccCmd.Use, "cc") {
		t.Errorf("ccCmd.Use should start with 'cc', got %q", ccCmd.Use)
	}
}

func TestCCCmd_Short(t *testing.T) {
	if ccCmd.Short == "" {
		t.Error("ccCmd.Short should not be empty")
	}
}

func TestCCCmd_IsSubcommandOfRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "cc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("cc should be registered as a subcommand of root")
	}
}

func TestCCCmd_Execution_NoDeps(t *testing.T) {
	// Use a temporary project root to prevent any mutation of real project files.
	// The project root finder is overridden via findProjectRootFn.
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, ".moai"), 0o755); err != nil {
		t.Fatal(err)
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmpDir, nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	// Override launchClaude to skip actual exec
	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()

	var launchedProfile string
	launchClaudeFunc = func(profile string, args []string) error {
		launchedProfile = profile
		return nil
	}

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := ccCmd.RunE(ccCmd, []string{})
	if err != nil {
		t.Fatalf("cc command should not error with nil deps, got: %v", err)
	}

	if launchedProfile != "" {
		t.Errorf("default profile should be empty, got %q", launchedProfile)
	}
}

func TestCCCmd_WithProfile(t *testing.T) {
	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	origLaunch := launchClaudeFunc
	defer func() { launchClaudeFunc = origLaunch }()

	var launchedProfile string
	launchClaudeFunc = func(profile string, args []string) error {
		launchedProfile = profile
		return nil
	}

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := ccCmd.RunE(ccCmd, []string{"-p", "work"})
	if err != nil {
		t.Fatalf("cc -p work should not error, got: %v", err)
	}

	if launchedProfile != "work" {
		t.Errorf("profile should be 'work', got %q", launchedProfile)
	}
}

// ── Characterization tests: capture existing behavior of cc.go (M6-S1 DDD) ──

// TestCharacterize_CC_HelpFlag verifies that -h / --help flags are intercepted
// by runCC before profile parsing and trigger cobra's Help() output.
func TestCharacterize_CC_HelpFlag(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			buf := new(bytes.Buffer)
			ccCmd.SetOut(buf)
			ccCmd.SetErr(buf)

			// runCC returns nil after printing help; launchClaudeFunc must NOT be called.
			origLaunch := launchClaudeFunc
			defer func() { launchClaudeFunc = origLaunch }()
			called := false
			launchClaudeFunc = func(_ string, _ []string) error {
				called = true
				return nil
			}

			err := runCC(ccCmd, []string{flag})
			if err != nil {
				t.Errorf("runCC(%s) should not error, got: %v", flag, err)
			}
			if called {
				t.Errorf("launchClaudeFunc must not be called when %s is present", flag)
			}
		})
	}
}

// TestCharacterize_CC_BypassFlag verifies that -b / --bypass pass through to
// unifiedLaunch as extra args (not consumed by parseProfileFlag).
func TestCharacterize_CC_BypassFlag(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedArgs []string
	unifiedLaunchFunc = func(_ string, _ string, args []string) error {
		capturedArgs = args
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{"-b"})
	if err != nil {
		t.Fatalf("runCC(-b) should not error, got: %v", err)
	}
	// -b is not -p/--profile so parseProfileFlag passes it through unchanged.
	if len(capturedArgs) == 0 || capturedArgs[0] != "-b" {
		t.Errorf("expected -b in extra args, got: %v", capturedArgs)
	}
}

// TestCharacterize_CC_ProfileFlag verifies that -p <name> sets profileName
// and is removed from extra args before being forwarded to unifiedLaunch.
func TestCharacterize_CC_ProfileFlag(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedProfile string
	var capturedArgs []string
	unifiedLaunchFunc = func(profile string, _ string, args []string) error {
		capturedProfile = profile
		capturedArgs = args
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{"-p", "myprofile", "--print"})
	if err != nil {
		t.Fatalf("runCC(-p myprofile) should not error, got: %v", err)
	}
	if capturedProfile != "myprofile" {
		t.Errorf("expected profile %q, got %q", "myprofile", capturedProfile)
	}
	// -p + value must be stripped; --print passes through.
	for _, a := range capturedArgs {
		if a == "-p" || a == "myprofile" {
			t.Errorf("profile flag/value must be stripped from extra args, got: %v", capturedArgs)
		}
	}
}

// TestCharacterize_CC_UnknownFlag verifies that unrecognised flags (e.g. --foo)
// are forwarded verbatim to unifiedLaunch as extra args.
func TestCharacterize_CC_UnknownFlag(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedArgs []string
	unifiedLaunchFunc = func(_ string, _ string, args []string) error {
		capturedArgs = args
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{"--unknown-flag", "value"})
	if err != nil {
		t.Fatalf("runCC(--unknown-flag) should not error, got: %v", err)
	}
	if !slices.Contains(capturedArgs, "--unknown-flag") {
		t.Errorf("unknown flag should be forwarded; got args: %v", capturedArgs)
	}
}

// TestCharacterize_CC_LaunchError verifies that errors from unifiedLaunch are
// propagated unchanged by runCC.
func TestCharacterize_CC_LaunchError(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	sentinel := errors.New("launch failed: exec not found")
	unifiedLaunchFunc = func(_ string, _ string, _ []string) error {
		return sentinel
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error to propagate, got: %v", err)
	}
}

// ── SPEC-FACTORY-MODE-001 M5: Kanban Mode entry surface ──

// TestParseKanbanFlag_LongFormWithoutSpec asserts the bare --kanban form
// enables Kanban Mode, reports no SPEC identifier (REQ-FM-005: absence means
// the chain heads at plan-phase), and strips the token from the forwarded args.
func TestParseKanbanFlag_LongFormWithoutSpec(t *testing.T) {
	p, err := parseKanbanFlag([]string{"--kanban", "-b"}); if err != nil { t.Fatal(err) }; spec, enabled, rest := p.Spec, p.KanbanEnabled, p.Rest
	if !enabled {
		t.Error("--kanban must enable Kanban Mode")
	}
	if spec != "" {
		t.Errorf("bare --kanban carries no SPEC identifier, got %q", spec)
	}
	if !slices.Equal(rest, []string{"-b"}) {
		t.Errorf("--kanban must be stripped from the forwarded args, got %v", rest)
	}
}

// TestParseKanbanFlag_ShortFormWithSpec is AC-FM-002: `moai cc -k
// SPEC-PLACEHOLDER` yields the identifier, enables the mode, and removes both
// tokens from the forwarded args.
func TestParseKanbanFlag_ShortFormWithSpec(t *testing.T) {
	p, err := parseKanbanFlag([]string{"-k", "SPEC-PLACEHOLDER", "--print"}); if err != nil { t.Fatal(err) }; spec, enabled, rest := p.Spec, p.KanbanEnabled, p.Rest
	if !enabled {
		t.Error("AC-FM-002: -k must enable Kanban Mode")
	}
	if spec != "SPEC-PLACEHOLDER" {
		t.Errorf("AC-FM-002: expected spec %q, got %q", "SPEC-PLACEHOLDER", spec)
	}
	if !slices.Equal(rest, []string{"--print"}) {
		t.Errorf("AC-FM-002: both tokens must be stripped, got %v", rest)
	}
}

// TestParseKanbanFlag_PassThroughBoundary is AC-FM-003: a --kanban token at
// or after the -- pass-through marker belongs to the child process, not to
// this parser. Mirrors stripSpawnFlag's boundary discipline exactly.
func TestParseKanbanFlag_PassThroughBoundary(t *testing.T) {
	p, err := parseKanbanFlag([]string{"--", "--kanban"}); if err != nil { t.Fatal(err) }; spec, enabled, rest := p.Spec, p.KanbanEnabled, p.Rest
	if enabled {
		t.Error("AC-FM-003: --kanban after -- must NOT enable Kanban Mode")
	}
	if spec != "" {
		t.Errorf("AC-FM-003: no SPEC identifier is consumed past --, got %q", spec)
	}
	if !slices.Equal(rest, []string{"--", "--kanban"}) {
		t.Errorf("AC-FM-003: tokens past -- are forwarded verbatim, got %v", rest)
	}
}

// TestParseKanbanFlag_SpecIsNotStolenFromAFlag guards the optional-positional
// parse: a following flag token is a separate argument, never the SPEC value.
func TestParseKanbanFlag_SpecIsNotStolenFromAFlag(t *testing.T) {
	p, err := parseKanbanFlag([]string{"--kanban", "--print"}); if err != nil { t.Fatal(err) }; spec, enabled, rest := p.Spec, p.KanbanEnabled, p.Rest
	if !enabled {
		t.Error("--kanban must enable Kanban Mode")
	}
	if spec != "" {
		t.Errorf("a flag token must not be consumed as the SPEC identifier, got %q", spec)
	}
	if !slices.Equal(rest, []string{"--print"}) {
		t.Errorf("the following flag must survive, got %v", rest)
	}
}

// TestCC_KanbanFlagStrippedBeforeLaunch is AC-FM-001: the seam sees neither
// --kanban nor -k, and the process environment carries the kanban signal at
// the moment the launch happens (the signal REQ-FM-023 transports).
func TestCC_KanbanFlagStrippedBeforeLaunch(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedArgs []string
	var kanbanAtLaunch, specAtLaunch string
	unifiedLaunchFunc = func(_ string, _ string, args []string) error {
		capturedArgs = args
		kanbanAtLaunch = os.Getenv(config.EnvMoaiKanban)
		specAtLaunch = os.Getenv(config.EnvMoaiKanbanSpec)
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	if err := runCC(ccCmd, []string{"--kanban", "SPEC-PLACEHOLDER"}); err != nil {
		t.Fatalf("AC-FM-001: runCC(--kanban) should not error, got: %v", err)
	}
	for _, a := range capturedArgs {
		if a == "--kanban" || a == "-k" {
			t.Errorf("AC-FM-001: kanban token must not reach the launcher, got %v", capturedArgs)
		}
	}
	if kanbanAtLaunch != "1" {
		t.Errorf("AC-FM-001: %s must be set at launch, got %q", config.EnvMoaiKanban, kanbanAtLaunch)
	}
	if specAtLaunch != "SPEC-PLACEHOLDER" {
		t.Errorf("AC-FM-001: %s must carry the identifier at launch, got %q", config.EnvMoaiKanbanSpec, specAtLaunch)
	}
}

// TestCC_KanbanWritesStateRecord is AC-FM-023a: entering Kanban Mode leaves a
// session-keyed record under .moai/state/kanban/, and a state directory that
// cannot be written never blocks the launch (fail-open).
func TestCC_KanbanWritesStateRecord(t *testing.T) {
	tmp := t.TempDir()
	sessionID := "kanban-record-session"
	sidecar := filepath.Join(tmp, session.CurrentSideChannelFile)
	if err := os.MkdirAll(filepath.Dir(sidecar), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sidecar, []byte(sessionID), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_PROJECT_DIR", tmp)

	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()
	unifiedLaunchFunc = func(_ string, _ string, _ []string) error { return nil }

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return tmp, nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	if err := runCC(ccCmd, []string{"--kanban", "SPEC-PLACEHOLDER"}); err != nil {
		t.Fatalf("AC-FM-023a: runCC should not error, got: %v", err)
	}

	rec, err := kanban.Read(tmp, sessionID)
	if err != nil {
		t.Fatalf("AC-FM-023a: kanban record not readable: %v", err)
	}
	if rec.SessionID != sessionID {
		t.Errorf("AC-FM-023a: session_id = %q, want %q", rec.SessionID, sessionID)
	}
	if rec.SpecID != "SPEC-PLACEHOLDER" {
		t.Errorf("AC-FM-023a: spec_id = %q, want %q", rec.SpecID, "SPEC-PLACEHOLDER")
	}
	if rec.Backend != kanban.BackendClaude {
		t.Errorf("AC-FM-023a: backend = %q, want %q", rec.Backend, kanban.BackendClaude)
	}
	if rec.EnteredAt == "" {
		t.Error("AC-FM-023a: entered_at must be stamped at launch")
	}
	// verify_rung is an orchestrator-written field: at launch it is deliberately
	// unrecorded (nil), which is how a later reader tells "never written" from
	// "written blank" (M3 record contract).
	if rec.VerifyRung != nil {
		t.Errorf("AC-FM-023a: verify_rung must be unrecorded at launch, got %q", *rec.VerifyRung)
	}

	// Fail-open: a state directory that cannot be written into must not block
	// the launch. The sidecar is written first, then the directory is sealed,
	// so session resolution still succeeds and only the record write fails.
	// Scoped to a subtest so the record-contract assertions above still run on
	// Windows, where the seal cannot be applied.
	t.Run("fail_open_unwritable_state_dir", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("a 0500 directory does not deny writes on Windows; the record would be written and the fail-open control asserts it is genuinely absent")
		}
		blocked := t.TempDir()
		stateDir := filepath.Join(blocked, ".moai", "state")
		if err := os.MkdirAll(stateDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(blocked, session.CurrentSideChannelFile), []byte("blocked-session"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(stateDir, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(stateDir, 0o755) })

		t.Setenv("CLAUDE_PROJECT_DIR", blocked)
		if err := runCC(ccCmd, []string{"--kanban"}); err != nil {
			t.Errorf("AC-FM-023a: an unwritable state directory must not block the launch, got: %v", err)
		}
		if _, err := kanban.Read(blocked, "blocked-session"); err == nil {
			t.Error("AC-FM-023a fail-open control: the record must genuinely be absent, otherwise the case is vacuous")
		}
	})
}

// TestCC_KanbanEnvMutationIsRestored is AC-FM-023d: the process-environment
// mutation is scoped to the call. Both the success path and the error path are
// asserted — a defer that only runs on success is the same leak with a
// narrower trigger — and an initially-absent variable is restored to ABSENT,
// not to the empty string.
func TestCC_KanbanEnvMutationIsRestored(t *testing.T) {
	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	// Baseline: neither variable is set in this process.
	_ = os.Unsetenv(config.EnvMoaiKanban)
	_ = os.Unsetenv(config.EnvMoaiKanbanSpec)

	cases := []struct {
		name    string
		launch  func(string, string, []string) error
		wantErr bool
	}{
		{"success path", func(string, string, []string) error { return nil }, false},
		{"error path", func(string, string, []string) error { return errors.New("launch failed") }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			beforeKanban, beforeKanbanSet := os.LookupEnv(config.EnvMoaiKanban)
			beforeSpec, beforeSpecSet := os.LookupEnv(config.EnvMoaiKanbanSpec)

			unifiedLaunchFunc = tc.launch
			err := runCC(ccCmd, []string{"--kanban", "SPEC-PLACEHOLDER"})
			if tc.wantErr && err == nil {
				t.Fatal("AC-FM-023d: expected the seam error to propagate")
			}

			for _, pair := range []struct {
				key    string
				val    string
				wasSet bool
			}{
				{config.EnvMoaiKanban, beforeKanban, beforeKanbanSet},
				{config.EnvMoaiKanbanSpec, beforeSpec, beforeSpecSet},
			} {
				got, gotSet := os.LookupEnv(pair.key)
				if gotSet != pair.wasSet {
					t.Errorf("AC-FM-023d: %s presence = %v after runCC, want %v (an absent variable must be unset, not set to \"\")",
						pair.key, gotSet, pair.wasSet)
				}
				if got != pair.val {
					t.Errorf("AC-FM-023d: %s = %q after runCC, want %q", pair.key, got, pair.val)
				}
			}
		})
	}
}

// TestEnterKanbanMode_RestoresPriorValue covers the other half of the restore
// contract: when a variable was ALREADY set before entry, the restore returns
// its prior value rather than unsetting it. Absence-restoration is asserted by
// TestCC_KanbanEnvMutationIsRestored; both branches must hold for the restore
// to be a genuine round-trip.
func TestEnterKanbanMode_RestoresPriorValue(t *testing.T) {
	t.Setenv(config.EnvMoaiKanban, "prior-kanban")
	t.Setenv(config.EnvMoaiKanbanSpec, "SPEC-PRIOR")

	restore := enterKanbanMode("SPEC-PLACEHOLDER", "")
	if got := os.Getenv(config.EnvMoaiKanban); got != "1" {
		t.Errorf("inside Kanban Mode %s = %q, want %q", config.EnvMoaiKanban, got, "1")
	}
	if got := os.Getenv(config.EnvMoaiKanbanSpec); got != "SPEC-PLACEHOLDER" {
		t.Errorf("inside Kanban Mode %s = %q, want %q", config.EnvMoaiKanbanSpec, got, "SPEC-PLACEHOLDER")
	}

	restore()
	if got, ok := os.LookupEnv(config.EnvMoaiKanban); !ok || got != "prior-kanban" {
		t.Errorf("%s = (%q, %v) after restore, want (%q, true)", config.EnvMoaiKanban, got, ok, "prior-kanban")
	}
	if got, ok := os.LookupEnv(config.EnvMoaiKanbanSpec); !ok || got != "SPEC-PRIOR" {
		t.Errorf("%s = (%q, %v) after restore, want (%q, true)", config.EnvMoaiKanbanSpec, got, ok, "SPEC-PRIOR")
	}
}

// TestEnterKanbanMode_WithoutSpecLeavesSpecVarUntouched asserts the optional
// identifier is genuinely optional: a bare --kanban publishes the mode signal
// without inventing a SPEC identifier the operator never supplied.
func TestEnterKanbanMode_WithoutSpecLeavesSpecVarUntouched(t *testing.T) {
	_ = os.Unsetenv(config.EnvMoaiKanbanSpec)
	restore := enterKanbanMode("", "")
	defer restore()

	if got := os.Getenv(config.EnvMoaiKanban); got != "1" {
		t.Errorf("%s = %q, want %q", config.EnvMoaiKanban, got, "1")
	}
	if got, ok := os.LookupEnv(config.EnvMoaiKanbanSpec); ok {
		t.Errorf("%s must stay unset when no identifier was supplied, got %q", config.EnvMoaiKanbanSpec, got)
	}
}

// TestCharacterize_CC_ModeIsAlwaysClaude verifies that runCC always passes
// modeOverride="claude" to unifiedLaunch regardless of extra flags.
func TestCharacterize_CC_ModeIsAlwaysClaude(t *testing.T) {
	origLaunch := unifiedLaunchFunc
	defer func() { unifiedLaunchFunc = origLaunch }()

	var capturedMode string
	unifiedLaunchFunc = func(_ string, mode string, _ []string) error {
		capturedMode = mode
		return nil
	}

	origFn := findProjectRootFn
	findProjectRootFn = func() (string, error) { return t.TempDir(), nil }
	defer func() { findProjectRootFn = origFn }()

	origDeps := deps
	defer func() { deps = origDeps }()
	deps = nil

	buf := new(bytes.Buffer)
	ccCmd.SetOut(buf)
	ccCmd.SetErr(buf)

	err := runCC(ccCmd, []string{})
	if err != nil {
		t.Fatalf("runCC should not error, got: %v", err)
	}
	if capturedMode != "claude" {
		t.Errorf("modeOverride must always be %q, got %q", "claude", capturedMode)
	}
}
