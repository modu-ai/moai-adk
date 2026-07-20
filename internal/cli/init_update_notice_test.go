package cli

// SPEC-CLI-TUX-V3-002 M2b — deferred self-update order contract
// (REQ-TUX2-001..004, AC-TUX2-001..004).
//
// The binary self-update check must never execute before wizard completion
// (interactive) or before first output (non-interactive), must surface as a
// non-blocking stderr notice carrying the `moai update` hint, must never
// re-exec after wizard answers were collected, and must preserve the
// shouldSkipBinaryUpdate skip semantics (templates-only / env guard /
// dev-build).

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/printer"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/update"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// newInitTestCmd builds a cobra command with the init flag set registered,
// mirroring the production initCmd flag surface used by runInit.
func newInitTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "init"}
	cmd.Flags().String("root", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("language", "", "")
	cmd.Flags().String("framework", "", "")
	cmd.Flags().String("mode", "", "")
	cmd.Flags().String("git-mode", "", "")
	cmd.Flags().String("git-provider", "", "")
	cmd.Flags().String("github-username", "", "")
	cmd.Flags().String("gitlab-instance-url", "", "")
	cmd.Flags().Bool("non-interactive", false, "")
	cmd.Flags().Bool("force", false, "")
	cmd.Flags().Bool("no-hooks", true, "")
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().Bool("standard", false, "")
	cmd.Flags().Bool("advanced", false, "")
	cmd.Flags().String("project-mode", "", "")
	cmd.Flags().String("harness-profile", "", "")
	cmd.Flags().Bool("enable-lsp", false, "")
	cmd.Flags().Bool("enforce-quality", true, "")
	cmd.Flags().Bool("enable-design", true, "")
	cmd.Flags().String("model-policy", "", "")
	cmd.Flags().Bool("high", false, "")
	cmd.Flags().Bool("medium-alias", false, "")
	cmd.Flags().Bool("low", false, "")
	cmd.Flags().String("plan-type", "", "")
	return cmd
}

// eventRecorder collects ordered events across goroutines.
type eventRecorder struct {
	mu     sync.Mutex
	events []string
}

func (r *eventRecorder) record(ev string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, ev)
}

func (r *eventRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.events...)
}

func (r *eventRecorder) count(ev string) int {
	n := 0
	for _, e := range r.snapshot() {
		if e == ev {
			n++
		}
	}
	return n
}

// TestInitNoNetworkBeforeWizard asserts the network-order contract
// (AC-TUX2-001): the injected update-check seam fires only AFTER the wizard
// completion event; zero update-check calls happen before or during the
// wizard.
func TestInitNoNetworkBeforeWizard(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv(config.EnvSkipBinaryUpdate, "")

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	rec := &eventRecorder{}

	origInteractive := isInteractiveStdin
	isInteractiveStdin = func() bool { return true }
	t.Cleanup(func() { isInteractiveStdin = origInteractive })

	origWizard := runWizardFn
	runWizardFn = func(rootFlag, locale string, standardMode, advancedMode bool) (*wizard.WizardResult, error) {
		rec.record("wizard-start")
		if got := rec.count("update-check"); got != 0 {
			t.Errorf("update-check ran before/during wizard: %d calls", got)
		}
		rec.record("wizard-complete")
		return &wizard.WizardResult{
			ProjectName:     "netorder-proj",
			DevelopmentMode: "tdd",
			GitMode:         "manual",
		}, nil
	}
	t.Cleanup(func() { runWizardFn = origWizard })

	origEnabled := deferredUpdateEnabled
	deferredUpdateEnabled = func(*cobra.Command) bool { return true }
	t.Cleanup(func() { deferredUpdateEnabled = origEnabled })

	origCheck := deferredUpdateCheck
	deferredUpdateCheck = func(*cobra.Command) *deferredUpdateResult {
		rec.record("update-check")
		return &deferredUpdateResult{Available: true, LatestVersion: "v9.9.9", CurrentVersion: "v0.0.1"}
	}
	t.Cleanup(func() { deferredUpdateCheck = origCheck })

	projectDir := filepath.Join(t.TempDir(), "netorder-proj")
	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	events := rec.snapshot()
	if rec.count("update-check") != 1 {
		t.Fatalf("expected exactly 1 update-check call, got %d (events: %v)", rec.count("update-check"), events)
	}
	wizardDone, checkAt := -1, -1
	for i, ev := range events {
		switch ev {
		case "wizard-complete":
			wizardDone = i
		case "update-check":
			checkAt = i
		}
	}
	if wizardDone == -1 || checkAt == -1 || checkAt < wizardDone {
		t.Fatalf("order violation: update-check must come after wizard-complete (events: %v)", events)
	}
	if !strings.Contains(errBuf.String(), "moai update") {
		t.Errorf("stderr must carry the deferred update notice with 'moai update' hint; got:\n%s", errBuf.String())
	}
}

// TestDeferredUpdateNotice_StderrHintOnly asserts the notice surface contract
// (AC-TUX2-002): new-version detection produces a stderr notice with the
// `moai update` follow-up hint; stdout stays clean.
func TestDeferredUpdateNotice_StderrHintOnly(t *testing.T) {
	origEnabled := deferredUpdateEnabled
	deferredUpdateEnabled = func(*cobra.Command) bool { return true }
	t.Cleanup(func() { deferredUpdateEnabled = origEnabled })

	origCheck := deferredUpdateCheck
	deferredUpdateCheck = func(*cobra.Command) *deferredUpdateResult {
		return &deferredUpdateResult{Available: true, LatestVersion: "v9.9.9", CurrentVersion: "v1.0.0"}
	}
	t.Cleanup(func() { deferredUpdateCheck = origCheck })

	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))

	flush := startDeferredUpdateNotice(cmd)
	flush(p)

	errStr := errBuf.String()
	if !strings.Contains(errStr, "moai update") {
		t.Errorf("stderr notice must include 'moai update' hint, got: %q", errStr)
	}
	if !strings.Contains(errStr, "v9.9.9") {
		t.Errorf("stderr notice must name the new version, got: %q", errStr)
	}
	if out.Len() != 0 {
		t.Errorf("stdout must stay clean (data channel), got: %q", out.String())
	}
}

// TestDeferredUpdateNotice_NoReexecReference is the static re-exec guard
// (AC-TUX2-002): after the deferred-order refactor, neither init.go nor the
// deferred-notice implementation may reference reexecNewBinary — the process
// must never be replaced after wizard answers were collected (REQ-TUX2-002).
// Grep-based static guard follows the TestNew_NoAskUserQuestion idiom.
func TestDeferredUpdateNotice_NoReexecReference(t *testing.T) {
	for _, src := range []string{"init.go", "init_update_notice.go"} {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		if strings.Contains(string(data), "reexecNewBinary") {
			t.Errorf("%s must not reference reexecNewBinary (REQ-TUX2-002: no re-exec after wizard)", src)
		}
	}
}

// TestDeferredUpdateNotice_NotAvailableSilent asserts the no-update path
// stays silent and the check error path is non-fatal (AC scenario 1).
func TestDeferredUpdateNotice_NotAvailableSilent(t *testing.T) {
	origEnabled := deferredUpdateEnabled
	deferredUpdateEnabled = func(*cobra.Command) bool { return true }
	t.Cleanup(func() { deferredUpdateEnabled = origEnabled })

	for name, res := range map[string]*deferredUpdateResult{
		"not-available": {Available: false},
		"check-error":   {Err: os.ErrDeadlineExceeded},
		"nil-result":    nil,
	} {
		t.Run(name, func(t *testing.T) {
			origCheck := deferredUpdateCheck
			deferredUpdateCheck = func(*cobra.Command) *deferredUpdateResult { return res }
			t.Cleanup(func() { deferredUpdateCheck = origCheck })

			cmd := newInitTestCmd()
			var out, errBuf bytes.Buffer
			p := printer.New(printer.WithWriters(&out, &errBuf))
			flush := startDeferredUpdateNotice(cmd)
			flush(p)

			if errBuf.Len() != 0 || out.Len() != 0 {
				t.Errorf("no-notice path must be silent, stderr=%q stdout=%q", errBuf.String(), out.String())
			}
		})
	}
}

// TestSkipBinaryUpdate_DeferredCheckSkipped characterizes the three
// shouldSkipBinaryUpdate skip semantics through the deferred path
// (AC-TUX2-003, REQ-TUX2-003): templates-only flag, EnvSkipBinaryUpdate env
// guard, and dev-build version detection each prevent the deferred check
// from ever running.
func TestSkipBinaryUpdate_DeferredCheckSkipped(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, cmd *cobra.Command)
	}{
		{
			name: "templates-only-flag",
			setup: func(t *testing.T, cmd *cobra.Command) {
				t.Setenv(config.EnvSkipBinaryUpdate, "")
				cmd.Flags().Bool("templates-only", false, "")
				if err := cmd.Flags().Set("templates-only", "true"); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "env-guard",
			setup: func(t *testing.T, _ *cobra.Command) {
				t.Setenv(config.EnvSkipBinaryUpdate, "1")
			},
		},
		{
			name: "dev-build-version",
			setup: func(t *testing.T, _ *cobra.Command) {
				t.Setenv(config.EnvSkipBinaryUpdate, "")
				orig := version.Version
				version.Version = "dev"
				t.Cleanup(func() { version.Version = orig })
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newInitTestCmd()
			tc.setup(t, cmd)

			var calls int32
			var mu sync.Mutex
			origCheck := deferredUpdateCheck
			deferredUpdateCheck = func(*cobra.Command) *deferredUpdateResult {
				mu.Lock()
				calls++
				mu.Unlock()
				return &deferredUpdateResult{Available: true, LatestVersion: "v9.9.9"}
			}
			t.Cleanup(func() { deferredUpdateCheck = origCheck })

			var out, errBuf bytes.Buffer
			p := printer.New(printer.WithWriters(&out, &errBuf))
			flush := startDeferredUpdateNotice(cmd) // default deferredUpdateEnabled
			flush(p)

			mu.Lock()
			got := calls
			mu.Unlock()
			if got != 0 {
				t.Errorf("deferred check must be skipped under %s, got %d calls", tc.name, got)
			}
			if errBuf.Len() != 0 {
				t.Errorf("skipped path must print nothing, got: %q", errBuf.String())
			}
		})
	}
}

// TestNonTTYUpdateCheckNonBlocking asserts the non-blocking property
// (AC-TUX2-004, REQ-TUX2-004): a slow in-flight check never blocks init exit
// beyond the bounded grace, and no notice is printed when the check misses
// the grace window.
func TestNonTTYUpdateCheckNonBlocking(t *testing.T) {
	origEnabled := deferredUpdateEnabled
	deferredUpdateEnabled = func(*cobra.Command) bool { return true }
	t.Cleanup(func() { deferredUpdateEnabled = origEnabled })

	origGrace := deferredNoticeGrace
	deferredNoticeGrace = 50 * time.Millisecond
	t.Cleanup(func() { deferredNoticeGrace = origGrace })

	release := make(chan struct{})
	origCheck := deferredUpdateCheck
	deferredUpdateCheck = func(*cobra.Command) *deferredUpdateResult {
		<-release // simulate a hung network check
		return &deferredUpdateResult{Available: true, LatestVersion: "v9.9.9"}
	}
	t.Cleanup(func() {
		close(release)
		deferredUpdateCheck = origCheck
	})

	cmd := newInitTestCmd()
	var out, errBuf bytes.Buffer
	p := printer.New(printer.WithWriters(&out, &errBuf))

	start := time.Now()
	flush := startDeferredUpdateNotice(cmd)
	flush(p)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("flush blocked for %v; must return within the bounded grace", elapsed)
	}
	if errBuf.Len() != 0 {
		t.Errorf("missed-grace path must print nothing, got: %q", errBuf.String())
	}
}

// TestInitNonInteractiveDeferredUpdateNotice asserts the non-interactive
// full path (AC-TUX2-004): phases complete without blocking on the check and
// the notice lands on stderr after the run.
func TestInitNonInteractiveDeferredUpdateNotice(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv(config.EnvSkipBinaryUpdate, "")

	origDeps := deps
	deps = nil
	t.Cleanup(func() { deps = origDeps })

	origEnabled := deferredUpdateEnabled
	deferredUpdateEnabled = func(*cobra.Command) bool { return true }
	t.Cleanup(func() { deferredUpdateEnabled = origEnabled })

	origCheck := deferredUpdateCheck
	deferredUpdateCheck = func(*cobra.Command) *deferredUpdateResult {
		return &deferredUpdateResult{Available: true, LatestVersion: "v9.9.9", CurrentVersion: "v0.0.1"}
	}
	t.Cleanup(func() { deferredUpdateCheck = origCheck })

	projectDir := filepath.Join(t.TempDir(), "noninteractive-proj")
	cmd := newInitTestCmd()
	if err := cmd.Flags().Set("non-interactive", "true"); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)

	if err := runInit(cmd, []string{projectDir}); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if !strings.Contains(errBuf.String(), "moai update") {
		t.Errorf("stderr must carry the deferred update notice, got:\n%s", errBuf.String())
	}
	if strings.Contains(out.String(), "moai update") && !strings.Contains(out.String(), "slim mode") {
		t.Errorf("update notice must not leak to stdout, got:\n%s", out.String())
	}
}

// fakeUpdateChecker implements update.Checker for the deferred-check test.
type fakeUpdateChecker struct {
	available bool
	info      *update.VersionInfo
	err       error
}

func (f *fakeUpdateChecker) CheckLatest(context.Context) (*update.VersionInfo, error) {
	return f.info, f.err
}

func (f *fakeUpdateChecker) IsUpdateAvailable(string) (bool, *update.VersionInfo, error) {
	return f.available, f.info, f.err
}

// TestDeferredUpdateNotice_DefaultCheckPaths covers defaultDeferredUpdateCheck
// against an injected deps.UpdateChecker: available / up-to-date / error —
// strictly CHECK-ONLY (no orchestrator, no install path touched).
func TestDeferredUpdateNotice_DefaultCheckPaths(t *testing.T) {
	origDeps := deps
	t.Cleanup(func() { deps = origDeps })

	cmd := newInitTestCmd()

	t.Run("nil-deps", func(t *testing.T) {
		deps = nil
		res := defaultDeferredUpdateCheck(cmd)
		if res == nil || res.Available || res.Err != nil {
			t.Errorf("nil deps must return an empty non-available result, got %+v", res)
		}
	})

	t.Run("available", func(t *testing.T) {
		deps = &Dependencies{UpdateChecker: &fakeUpdateChecker{
			available: true,
			info:      &update.VersionInfo{Version: "v9.9.9"},
		}}
		res := defaultDeferredUpdateCheck(cmd)
		if res.Err != nil || !res.Available || res.LatestVersion != "v9.9.9" {
			t.Errorf("expected available v9.9.9, got %+v", res)
		}
		if res.CurrentVersion == "" {
			t.Error("current version must be populated")
		}
	})

	t.Run("up-to-date", func(t *testing.T) {
		deps = &Dependencies{UpdateChecker: &fakeUpdateChecker{available: false}}
		res := defaultDeferredUpdateCheck(cmd)
		if res.Err != nil || res.Available {
			t.Errorf("expected non-available result, got %+v", res)
		}
	})

	t.Run("check-error", func(t *testing.T) {
		deps = &Dependencies{UpdateChecker: &fakeUpdateChecker{err: os.ErrDeadlineExceeded}}
		res := defaultDeferredUpdateCheck(cmd)
		if res.Err == nil {
			t.Errorf("expected error result, got %+v", res)
		}
	})
}
