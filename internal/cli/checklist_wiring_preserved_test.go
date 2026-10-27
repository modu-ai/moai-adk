package cli

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/cli/wizard"
)

// AC-TCL-024 (SPEC-TUI-CHECKLIST-001): every preserved line in scope appears
// exactly once across both command writers, and the item that was open when it
// was written ends with the glyph its kind requires (a warning or error row
// ends the item with the warning glyph; a notice row leaves the glyph alone; a
// row written outside any item carries no glyph requirement).
//
// Every subtest drives the real production code path and makes the preserved
// line appear by a real failure condition on disk — never by writing the
// fragment itself — so a regression that drops or duplicates the line fails
// here.

// preservedCount returns how many times frag appears across both collapsed
// command writers of a rich run.
func preservedCount(stdout, stderr, frag string) int {
	return strings.Count(checklistCollapse(stdout)+checklistCollapse(stderr), frag)
}

// requirePreservedOnce fails unless frag appears exactly once across both writers.
func requirePreservedOnce(t *testing.T, stdout, stderr, frag string) {
	t.Helper()
	if n := preservedCount(stdout, stderr, frag); n != 1 {
		t.Errorf("%q appears %d times across both writers, want exactly 1:\nstdout:\n%s\nstderr:\n%s", frag, n, checklistCollapse(stdout), checklistCollapse(stderr))
	}
}

// requireItemGlyph asserts the item line for key ends with want and not other.
func requireItemGlyph(t *testing.T, stderr, key, want, other string) {
	t.Helper()
	label := checklistLabel(key)
	var hits []string
	for _, l := range checklistLines(checklistCollapse(stderr)) {
		// An item line leads with its glyph; a detail line that merely mentions
		// the label text (a preserved notice) does not.
		trimmed := strings.TrimSpace(l)
		for _, g := range []string{"✓", "!", "✗", "○"} {
			if strings.HasPrefix(trimmed, g+" "+label) {
				hits = append(hits, l)
			}
		}
	}
	if len(hits) != 1 {
		t.Errorf("item %q appears on %d glyph-led lines, want exactly 1:\n%s", label, len(hits), checklistCollapse(stderr))
		return
	}
	if !strings.Contains(hits[0], want) {
		t.Errorf("item line %q lacks the %s glyph", hits[0], want)
	}
	if strings.Contains(hits[0], other) {
		t.Errorf("item line %q must not carry the %s glyph", hits[0], other)
	}
}

// requireSummaryWarnings asserts the summary reports exactly n warnings: a
// preserved row is never added to the count (decision D-9).
func requireSummaryWarnings(t *testing.T, stderr string, n int) {
	t.Helper()
	want := "· " + strconv.Itoa(n) + " warnings ·"
	if n == 1 {
		want = "· 1 warning ·"
	}
	if !strings.Contains(checklistCollapse(stderr), want) {
		t.Errorf("the summary does not carry %q (a preserved row must not change the count):\n%s", want, checklistCollapse(stderr))
	}
}

// clMkdirAll creates dir or fails the test.
func clMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

// clWriteFile writes data to path (creating parents) or fails the test.
func clWriteFile(t *testing.T, path, data string, perm os.FileMode) {
	t.Helper()
	clMkdirAll(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(data), perm); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// userHookBody is a hook without the MoAI marker, so the installers preserve it.
const userHookBody = "#!/bin/sh\n# user hook, no managed-by marker\necho my hook\n"

// preservedSubtest is one platform-specific row, registered by a build-tagged file.
type preservedSubtest struct {
	name string
	fn   func(*testing.T)
}

// preservedPlatformSubtests holds the rows whose failure injection needs a
// platform feature (file permissions); the unix file fills it.
var preservedPlatformSubtests []preservedSubtest

func TestChecklistPreservedLinesExactlyOnce(t *testing.T) {
	for _, st := range preservedPlatformSubtests {
		t.Run(st.name, st.fn)
	}
	t.Run("inventory_complete", checklistInventoryComplete)

	t.Run("p01_init_snapshot_stage", func(t *testing.T) {
		root, projectDir := checklistFallbackEnv(t)
		_ = root
		forceChecklistOn(t)
		checklistRichClock(t)
		settingsPending := filepath.Join(backup.SnapshotDir(projectDir), "claude", "settings.json.pending")
		mcpPending := filepath.Join(backup.SnapshotDir(projectDir), "mcp.json.pending")
		injectIntoStep(t, "Initialization", func() {
			// A directory where the staging copy goes makes the WriteFile fail.
			clMkdirAll(t, settingsPending)
			clMkdirAll(t, mcpPending)
		})
		cmd, out, errBuf := newChecklistInitCmd(t, nil)
		if err := runInit(cmd, nil); err != nil {
			t.Fatalf("init: %v\n%s", err, errBuf.String())
		}
		requirePreservedOnce(t, out.String(), errBuf.String(), backup.SettingsSnapshotWriteFailedPrefix)
		requirePreservedOnce(t, out.String(), errBuf.String(), backup.MCPSnapshotWriteFailedPrefix)
		requireItemGlyph(t, errBuf.String(), "templates.deploy", "!", "✓")
		requireSummaryWarnings(t, errBuf.String(), 0)
	})

	t.Run("p02_init_snapshot_settle", func(t *testing.T) {
		_, projectDir := checklistFallbackEnv(t)
		forceChecklistOn(t)
		checklistRichClock(t)
		orig := applyAutonomyTierBundleFn
		applyAutonomyTierBundleFn = func(a, b, c, d string) error {
			err := orig(a, b, c, d)
			// A non-empty directory on the canonical path makes the promoting
			// rename fail; this runs while config.autonomy is open.
			clMkdirAll(t, filepath.Join(backup.SnapshotDir(projectDir), "claude", "settings.json", "keep"))
			clMkdirAll(t, filepath.Join(backup.SnapshotDir(projectDir), "mcp.json", "keep"))
			return err
		}
		t.Cleanup(func() { applyAutonomyTierBundleFn = orig })
		cmd, out, errBuf := newChecklistInitCmd(t, nil)
		if err := runInit(cmd, nil); err != nil {
			t.Fatalf("init: %v\n%s", err, errBuf.String())
		}
		requirePreservedOnce(t, out.String(), errBuf.String(), backup.SettingsSnapshotPromoteFailedPrefix)
		requirePreservedOnce(t, out.String(), errBuf.String(), backup.MCPSnapshotPromoteFailedPrefix)
		requireItemGlyph(t, errBuf.String(), "config.autonomy", "!", "✓")
		requireSummaryWarnings(t, errBuf.String(), 0)
	})

	t.Run("p03_init_resumed_notice", func(t *testing.T) {
		resumeHome(t)
		projectDir := filepath.Join(t.TempDir(), "resume-proj")
		resumeSeedRun(t, projectDir, "claude")
		forceChecklistOn(t)
		checklistRichClock(t)
		t.Setenv("NO_COLOR", "1")
		cmd := newInitTestCmd()
		cmd.Flags().Bool("verbose", false, "")
		if err := cmd.Flags().Set("llm", "claude"); err != nil {
			t.Fatal(err)
		}
		out, errBuf := &strings.Builder{}, &strings.Builder{}
		cmd.SetOut(out)
		cmd.SetErr(errBuf)
		if err := runInit(cmd, []string{projectDir}); err != nil {
			t.Fatalf("the resume must return nil: %v\n%s", err, errBuf.String())
		}
		requirePreservedOnce(t, out.String(), errBuf.String(), "Initialized MoAI project (resumed")
	})

	// P4 has two init call sites: the flush after the MCP entry (inside the
	// mcp.entry item) and the one after a resume (outside any item).
	t.Run("p04_init_update_notice", func(t *testing.T) {
		injectUpdateAvailable := func(t *testing.T) {
			t.Helper()
			origEnabled, origCheck := deferredUpdateEnabled, deferredUpdateCheck
			deferredUpdateEnabled = func(*cobra.Command) bool { return true }
			deferredUpdateCheck = func(*cobra.Command, *Dependencies) *deferredUpdateResult {
				return &deferredUpdateResult{Available: true, LatestVersion: "v9.9.9", CurrentVersion: "v0.0.1"}
			}
			t.Cleanup(func() { deferredUpdateEnabled, deferredUpdateCheck = origEnabled, origCheck })
		}

		t.Run("mcp_entry_site", func(t *testing.T) {
			checklistFallbackEnv(t)
			forceChecklistOn(t)
			checklistRichClock(t)
			injectUpdateAvailable(t)
			// The interactive seam provisions the MCP entry, so the flush runs
			// while the mcp.entry item is open.
			origInteractive := isInteractiveStdin
			isInteractiveStdin = func() bool { return true }
			t.Cleanup(func() { isInteractiveStdin = origInteractive })
			origDeps := deps
			deps = nil
			t.Cleanup(func() { deps = origDeps })
			origWizard := runWizardFn
			runWizardFn = func(_, _, _ string) (*wizard.WizardResult, error) {
				return &wizard.WizardResult{ConversationLang: "en", DevelopmentMode: "tdd"}, nil
			}
			t.Cleanup(func() { runWizardFn = origWizard })

			cmd, out, errBuf := newChecklistInitCmd(t, map[string]string{"non-interactive": "false"})
			if err := runInit(cmd, nil); err != nil {
				t.Fatalf("init: %v\n%s", err, errBuf.String())
			}
			requirePreservedOnce(t, out.String(), errBuf.String(), "A newer moai binary is available")
			// A notice row never changes the glyph of the item it was written in,
			// and the notice sits directly under the mcp.entry item.
			requireItemGlyph(t, errBuf.String(), "mcp.entry", "✓", "!")
			lines := checklistLines(checklistCollapse(errBuf.String()))
			under := false
			for i, l := range lines {
				if strings.HasPrefix(strings.TrimSpace(l), "✓ "+checklistLabel("mcp.entry")) {
					for _, next := range lines[i+1:] {
						if !strings.HasPrefix(next, "    ") {
							break
						}
						if strings.Contains(next, "A newer moai binary is available") {
							under = true
						}
					}
				}
			}
			if !under {
				t.Errorf("the notice is not expanded under the mcp.entry item:\n%s", checklistCollapse(errBuf.String()))
			}
			requireSummaryWarnings(t, errBuf.String(), 0)
		})

		t.Run("resume_site", func(t *testing.T) {
			resumeHome(t)
			projectDir := filepath.Join(t.TempDir(), "resume-proj")
			resumeSeedRun(t, projectDir, "claude")
			forceChecklistOn(t)
			checklistRichClock(t)
			t.Setenv("NO_COLOR", "1")
			injectUpdateAvailable(t)
			cmd := newInitTestCmd()
			cmd.Flags().Bool("verbose", false, "")
			if err := cmd.Flags().Set("llm", "claude"); err != nil {
				t.Fatal(err)
			}
			out, errBuf := &strings.Builder{}, &strings.Builder{}
			cmd.SetOut(out)
			cmd.SetErr(errBuf)
			if err := runInit(cmd, []string{projectDir}); err != nil {
				t.Fatalf("the resume must return nil: %v\n%s", err, errBuf.String())
			}
			requirePreservedOnce(t, out.String(), errBuf.String(), "A newer moai binary is available")
		})
	})

	t.Run("p05_init_worktree_advisory", func(t *testing.T) {
		checklistFallbackEnv(t)
		forceChecklistOn(t)
		checklistRichClock(t)
		cmd, out, errBuf := newChecklistInitCmd(t, nil)
		if err := runInit(cmd, nil); err != nil {
			t.Fatalf("init: %v\n%s", err, errBuf.String())
		}
		requirePreservedOnce(t, out.String(), errBuf.String(), "this checkout is shared across concurrent sessions;")
		requireItemGlyph(t, errBuf.String(), "hooks.precommit", "✓", "!")
		requireSummaryWarnings(t, errBuf.String(), 0)
	})

	t.Run("p20_prepush_preserved", func(t *testing.T) {
		_, projectDir := checklistFallbackEnv(t)
		forceChecklistOn(t)
		checklistRichClock(t)
		clWriteFile(t, filepath.Join(projectDir, ".git", "hooks", "pre-push"), userHookBody, 0o755)
		cmd, out, errBuf := newChecklistInitCmd(t, nil)
		if err := runInit(cmd, nil); err != nil {
			t.Fatalf("init: %v\n%s", err, errBuf.String())
		}
		requirePreservedOnce(t, out.String(), errBuf.String(), "existing pre-push hook preserved")
		requireItemGlyph(t, errBuf.String(), "hooks.prepush", "✓", "!")
		requireSummaryWarnings(t, errBuf.String(), 0)
	})

	t.Run("p21_precommit_preserved", func(t *testing.T) {
		_, projectDir := checklistFallbackEnv(t)
		forceChecklistOn(t)
		checklistRichClock(t)
		clWriteFile(t, filepath.Join(projectDir, ".git", "hooks", "pre-commit"), userHookBody, 0o755)
		cmd, out, errBuf := newChecklistInitCmd(t, nil)
		if err := runInit(cmd, nil); err != nil {
			t.Fatalf("init: %v\n%s", err, errBuf.String())
		}
		requirePreservedOnce(t, out.String(), errBuf.String(), "existing pre-commit hook preserved")
		requireItemGlyph(t, errBuf.String(), "hooks.precommit", "✓", "!")
		requireSummaryWarnings(t, errBuf.String(), 0)
	})
}
