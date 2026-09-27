package cli

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestHookPathWarnRecordReachesSink is the behavioral guard for AC-HDS-001
// (REQ-HDS-001, REQ-HDS-011): a warn record emitted on the `moai hook` path
// reaches the file sink under the resolved project root.
//
// It asserts the EFFECT, not the routing: the record must be READ BACK out of
// .moai/logs/hook-runtime.log. "the destination is not io.Discard" would pass
// against a writer that drops every byte, which is the exact defect this SPEC
// exists to close.
//
// Falsification (AC-HDS-002): restore `dest: io.Discard` in the hook branch of
// resolveLoggingDecision and this test fails while still being listed.
//
// Not parallel — t.Setenv.
func TestHookPathWarnRecordReachesSink(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	// Explicitly empty: the default minimum level (warn) is what admits the
	// record below, and an ambient MOAI_LOG_LEVEL must not decide that.
	t.Setenv(config.EnvLogLevel, "")

	const msg = "hook-sink-reachability-probe"

	d := resolveLoggingDecision([]string{"hook", "pre-tool"})
	slog.New(slog.NewTextHandler(d.dest, &slog.HandlerOptions{Level: d.level})).
		Warn(msg, "probe", "guard-1")

	sink := filepath.Join(root, hookRuntimeLogRelPath)
	raw, err := os.ReadFile(filepath.Clean(sink))
	if err != nil {
		t.Fatalf("read hook sink %s: %v\n"+
			"a warn record emitted on the hook path must reach the sink file (AC-HDS-001)", sink, err)
	}

	var hits int
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if strings.Contains(line, msg) {
			hits++
		}
	}
	if hits != 1 {
		t.Errorf("hook sink %s holds %d line(s) carrying %q, want exactly 1\nfile contents:\n%s",
			sink, hits, msg, raw)
	}
}
