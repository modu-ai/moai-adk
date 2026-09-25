package cli

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// levelProbes names the three records every arm of the level boundary emits.
// The messages are distinct so a single scan of the sink separates them; the
// levels are the three that straddle defaultLogLevel.
var levelProbes = []struct {
	msg   string
	level slog.Level
}{
	{"hook-sink-level-probe-debug", slog.LevelDebug},
	{"hook-sink-level-probe-info", slog.LevelInfo},
	{"hook-sink-level-probe-warn", slog.LevelWarn},
}

// TestHookSinkLevelBoundary is the level guard for AC-HDS-007 and AC-HDS-008
// (REQ-HDS-004, REQ-HDS-005): the sink sits BEHIND the level gate, at the same
// default minimum as every other path, and MOAI_LOG_LEVEL moves that minimum.
//
// This boundary is what bounds the SPEC's own scope. Of the slog call sites
// reachable on the hook path, warn-and-above are the records that would have
// been read had the hook path not discarded them; info and debug sit below
// defaultLogLevel and stay silent on every other path too. Routing them to the
// sink by default would grow .moai/logs/ on every one of the dozens-to-hundreds
// of hook invocations a session makes, recovering nothing anyone would have
// seen. That is a new cost, not a defect repaired.
//
// The third subtest is the one that makes AC-HDS-008 non-vacuous. A test that
// never actually varied the environment would still satisfy "with the level
// lowered, an info record reaches the sink" — by reading a sink that was never
// at the default in the first place. The differential arm asserts the SAME
// record across the TWO settings and requires them to disagree, so an omitted
// or ineffective t.Setenv fails rather than passes.
//
// AC-HDS-009 (MOAI_LOG_LEVEL never moves the DESTINATION) is deliberately NOT
// re-asserted here: guard 2, TestHookPathHookDestIsNeitherStdStream, already
// sweeps the four named level values across both hook arg shapes and asserts
// the resolved dest is neither standard stream. Repeating it would be a second
// place to update when that table changes, for no additional coverage — the two
// guards divide the variable's two axes, destination and level, one each.
//
// Not parallel — subtests call t.Setenv.
func TestHookSinkLevelBoundary(t *testing.T) {
	t.Run("default_config_admits_warn_only", func(t *testing.T) {
		// AC-HDS-007: the unset case, which is what a hook invocation actually
		// runs under in normal operation.
		got := emitLevelProbes(t, "", defaultLogLevel)

		assertProbeCount(t, got, "hook-sink-level-probe-warn", 1,
			"warn is at defaultLogLevel and must reach the sink (AC-HDS-007)")
		assertProbeCount(t, got, "hook-sink-level-probe-info", 0,
			"info sits below defaultLogLevel and must stay out of the sink (AC-HDS-007, REQ-HDS-004)")
		assertProbeCount(t, got, "hook-sink-level-probe-debug", 0,
			"debug sits below defaultLogLevel and must stay out of the sink (AC-HDS-007, REQ-HDS-004)")
	})

	t.Run("lowered_level_admits_info", func(t *testing.T) {
		// AC-HDS-008: the same sink, one variable lowered.
		got := emitLevelProbes(t, "INFO", slog.LevelInfo)

		assertProbeCount(t, got, "hook-sink-level-probe-info", 1,
			"MOAI_LOG_LEVEL=INFO lowers the minimum, so info must reach the sink (AC-HDS-008, REQ-HDS-005)")
		assertProbeCount(t, got, "hook-sink-level-probe-warn", 1,
			"warn is above the lowered minimum and must still reach the sink (AC-HDS-008)")
		assertProbeCount(t, got, "hook-sink-level-probe-debug", 0,
			"debug is still below INFO and must stay out of the sink (AC-HDS-008)")
	})

	t.Run("the_same_record_differs_across_the_two_settings", func(t *testing.T) {
		// The anti-vacuity arm: one record, two environments, one required
		// disagreement. Both sinks are read back from disk, so this fails if
		// the variable never reached resolveLoggingDecision.
		atDefault := emitLevelProbes(t, "", defaultLogLevel)
		lowered := emitLevelProbes(t, "INFO", slog.LevelInfo)

		const probe = "hook-sink-level-probe-info"
		if atDefault[probe] == lowered[probe] {
			t.Errorf("the info probe appears %d time(s) at the default AND %d time(s) with %s=INFO; "+
				"the two settings must disagree, or the level boundary is not being exercised at all "+
				"(AC-HDS-007 / AC-HDS-008)",
				atDefault[probe], lowered[probe], config.EnvLogLevel)
		}
	})
}

// emitLevelProbes installs the hook path's logging decision under a fresh
// project root with MOAI_LOG_LEVEL set to level, emits one record per entry in
// levelProbes, and reports how many sink lines carry each probe message.
//
// It also asserts the resolved minimum level, which is what pins REQ-HDS-005 to
// the decision itself: counting sink lines alone cannot distinguish "the
// variable was honored" from "the handler happened to admit the record".
//
// A missing sink file reports zero for every probe rather than failing — that
// is the correct reading when no record was admitted (REQ-HDS-007's lazy open
// means nothing is created), and it is exactly the state AC-HDS-007's
// debug/info arms expect.
func emitLevelProbes(t *testing.T, level string, wantLevel slog.Level) map[string]int {
	t.Helper()

	root := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	t.Setenv(config.EnvLogLevel, level)

	d := resolveLoggingDecision([]string{"hook", "pre-tool"})
	if d.level != wantLevel {
		t.Fatalf("resolveLoggingDecision with %s=%q resolved level %v, want %v; "+
			"the variable governs the hook sink's minimum level (REQ-HDS-005)",
			config.EnvLogLevel, level, d.level, wantLevel)
	}

	logger := slog.New(slog.NewTextHandler(d.dest, &slog.HandlerOptions{Level: d.level}))
	for _, p := range levelProbes {
		logger.Log(t.Context(), p.level, p.msg, "probe", "guard-6")
	}

	counts := make(map[string]int, len(levelProbes))
	for _, p := range levelProbes {
		counts[p.msg] = 0
	}

	sink := filepath.Join(root, hookRuntimeLogRelPath)
	raw, err := os.ReadFile(filepath.Clean(sink))
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("read hook sink %s: %v", sink, err)
		}
		return counts
	}

	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		for _, p := range levelProbes {
			if strings.Contains(line, p.msg) {
				counts[p.msg]++
			}
		}
	}
	return counts
}

// assertProbeCount reports a mismatch with the reason the count matters, so a
// failure names the boundary rather than only the number.
func assertProbeCount(t *testing.T, counts map[string]int, probe string, want int, why string) {
	t.Helper()

	if got := counts[probe]; got != want {
		t.Errorf("sink holds %d line(s) carrying %q, want %d: %s", got, probe, want, why)
	}
}
