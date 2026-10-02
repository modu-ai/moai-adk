package cli

// codex_debug_uniformity_test.go — SPEC-CODEX-DEBUG-MODE-001 M3: the
// three-runner uniformity contract (AC-005, AC-006, AC-007). cc and glm are
// OBSERVE-ONLY: the debug token stays in the child arguments untouched (the
// child's native -d handling is preserved — REQ-004) while the launcher
// emits its own trace under the SAME prefix constant and destination as the
// codex launcher (REQ-013). The matrix iterates all six launcher x spelling
// combinations and asserts the per-backend linkage.

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/kanban"
	"github.com/spf13/cobra"
)

// freshLauncherEntry builds a launcher command with buffer-backed streams
// and hands the stderr buffer back for prefix assertions.
func freshLauncherEntry(t *testing.T, use string) (*cobra.Command, *strings.Builder) {
	t.Helper()
	c := &cobra.Command{Use: use}
	c.SetContext(context.Background())
	errB := new(strings.Builder)
	c.SetOut(new(strings.Builder))
	c.SetErr(errB)
	return c, errB
}

// launcherTraceLineCount counts the debug-prefix lines in a stderr buffer.
func launcherTraceLineCount(stderr string) int {
	n := 0
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, testDebugTracePrefix) {
			n++
		}
	}
	return n
}

// AC-005: the cc launch is observe-only — the child keeps -d and the
// launcher traces its pre-exec path to stderr.
func TestCCDebugTokenObserveOnly(t *testing.T) {
	c, errB := freshLauncherEntry(t, "cc")
	launches := 0
	var gotArgs []string
	err := runClaudeEntry(c, []string{"-d", "--", "extra"}, "cc", "claude", kanban.BackendClaude, func(_ string, _ string, args []string) error {
		launches++
		gotArgs = args
		return nil
	})
	if err != nil {
		t.Fatalf("runClaudeEntry(-d -- extra): %v", err)
	}
	if launches != 1 {
		t.Fatalf("launches = %d, want 1", launches)
	}
	if !slices.Contains(gotArgs, "-d") {
		t.Errorf("child args dropped -d — REQ-004 observe-only violation: %v", gotArgs)
	}
	if !slices.Contains(gotArgs, "extra") {
		t.Errorf("child args dropped the post--- tail: %v", gotArgs)
	}
	if launcherTraceLineCount(errB.String()) == 0 {
		t.Errorf("cc launcher emitted no debug trace lines under the %q prefix:\n%s", testDebugTracePrefix, errB.String())
	}
}

// AC-006: the glm launch is observe-only under the same contract.
func TestGLMDebugTokenObserveOnly(t *testing.T) {
	c, errB := freshLauncherEntry(t, "glm")
	prevLaunch := unifiedLaunchFunc
	launches := 0
	var gotArgs []string
	unifiedLaunchFunc = func(_ string, _ string, args []string) error {
		launches++
		gotArgs = args
		return nil
	}
	t.Cleanup(func() { unifiedLaunchFunc = prevLaunch })
	if err := runGLM(c, []string{"-d"}); err != nil {
		t.Fatalf("runGLM(-d): %v", err)
	}
	if launches != 1 {
		t.Fatalf("launches = %d, want 1", launches)
	}
	if !slices.Contains(gotArgs, "-d") {
		t.Errorf("child args dropped -d — REQ-004 observe-only violation: %v", gotArgs)
	}
	if launcherTraceLineCount(errB.String()) == 0 {
		t.Errorf("glm launcher emitted no debug trace lines under the %q prefix:\n%s", testDebugTracePrefix, errB.String())
	}
}

// The pre--- scoping is shared: a post--- debug token activates nothing on
// cc either (the token rides the child verbatim, REQ-002's cc form).
func TestCCDebugTokenAfterDashDashInactive(t *testing.T) {
	c, errB := freshLauncherEntry(t, "cc")
	launches := 0
	err := runClaudeEntry(c, []string{"--", "-d"}, "cc", "claude", kanban.BackendClaude, func(_ string, _ string, _ []string) error {
		launches++
		return nil
	})
	if err != nil {
		t.Fatalf("runClaudeEntry(-- -d): %v", err)
	}
	if launches != 1 {
		t.Fatalf("launches = %d, want 1", launches)
	}
	if launcherTraceLineCount(errB.String()) != 0 {
		t.Errorf("post--- debug token activated the cc trace — REQ-002 scoping violation:\n%s", errB.String())
	}
}

// AC-007: the uniformity matrix — every launcher x spelling combination
// accepts the token pre---, traces under the SAME prefix constant to
// stderr, and keeps the per-backend linkage (cc/glm forward the token;
// codex strips it and injects RUST_LOG=debug).
func TestThreeRunnerDebugUniformityMatrix(t *testing.T) {
	type outcome struct {
		childHasToken bool // the token reached the child argv
		rustLogDebug  bool // the child env carries RUST_LOG=debug
		traceLines    int
	}
	for _, tc := range []struct {
		launcher, token string
	}{
		{"cc", "-d"}, {"cc", "--debug"},
		{"glm", "-d"}, {"glm", "--debug"},
		{"codex", "-d"}, {"codex", "--debug"},
	} {
		t.Run(tc.launcher+" "+tc.token, func(t *testing.T) {
			var got outcome
			switch tc.launcher {
			case "codex":
				unsetRustLog(t)
				cap := withCodexLaunchCapture(t)
				withCodexProjectRoot(t, t.TempDir())
				_, stderr, err := runCodexCmd(t, tc.token)
				if err != nil {
					t.Fatalf("runCodex(%s): %v", tc.token, err)
				}
				if cap.count() != 1 {
					t.Fatalf("launches = %d, want 1", cap.count())
				}
				got.childHasToken = slices.Contains(cap.records[0].Argv, tc.token)
				got.rustLogDebug = func() bool {
					v := envValues(cap.records[0].Env, config.EnvRustLog)
					return len(v) > 0 && v[len(v)-1] == "debug"
				}()
				got.traceLines = launcherTraceLineCount(stderr)
			case "cc":
				c, errB := freshLauncherEntry(t, "cc")
				var args []string
				err := runClaudeEntry(c, []string{tc.token}, "cc", "claude", kanban.BackendClaude, func(_ string, _ string, a []string) error {
					args = a
					return nil
				})
				if err != nil {
					t.Fatalf("runClaudeEntry(%s): %v", tc.token, err)
				}
				got.childHasToken = slices.Contains(args, tc.token)
				got.traceLines = launcherTraceLineCount(errB.String())
			case "glm":
				c, errB := freshLauncherEntry(t, "glm")
				prevLaunch := unifiedLaunchFunc
				var args []string
				unifiedLaunchFunc = func(_ string, _ string, a []string) error {
					args = a
					return nil
				}
				t.Cleanup(func() { unifiedLaunchFunc = prevLaunch })
				if err := runGLM(c, []string{tc.token}); err != nil {
					t.Fatalf("runGLM(%s): %v", tc.token, err)
				}
				got.childHasToken = slices.Contains(args, tc.token)
				got.traceLines = launcherTraceLineCount(errB.String())
			}
			if !got.childHasToken && tc.launcher != "codex" {
				t.Errorf("%s dropped the %s token from the child argv — observe-only violation", tc.launcher, tc.token)
			}
			if got.childHasToken && tc.launcher == "codex" {
				t.Errorf("codex forwarded the %s token to the child — REQ-001 violation", tc.token)
			}
			if tc.launcher == "codex" && !got.rustLogDebug {
				t.Errorf("codex did not inject RUST_LOG=debug — REQ-010 linkage missing")
			}
			if got.traceLines == 0 {
				t.Errorf("%s emitted no trace lines under the %q prefix — REQ-013 uniformity violation", tc.launcher, testDebugTracePrefix)
			}
		})
	}
}
