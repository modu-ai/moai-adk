package cli

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestHookSinkIsNotCreatedWithoutRecords is the lazy-open guard for AC-HDS-012
// (REQ-HDS-007): installing the hook path's logging decision and emitting no
// admitted record creates neither the sink file nor .moai/logs/.
//
// It carries more weight than its size suggests. Lazy open is the sole reason
// plan.md §B-1 chose this design over the alternatives — hooks fire dozens to
// hundreds of times a session, and a silent one must cost nothing — yet until
// this guard existed, AC-HDS-012 rode on guard 3
// (TestHookPathSinkFailureIsFailOpen), whose path is about what happens AFTER
// an open is attempted. That guard passes without ever observing lazy open.
//
// The assertion is wider than "the file does not exist": the DIRECTORY must not
// be created either. Opening the sink runs MkdirAll before OpenFile, so an
// implementation that created .moai/logs/ and then failed to open the file
// would satisfy a file-only check while having already paid the cost the
// laziness exists to avoid.
//
// Three phases, in order:
//
//  1. Resolve and install the decision, emit nothing. Nothing may exist.
//  2. Emit a record BELOW the level bar. slog rejects it before the writer sees
//     a byte, so nothing may exist still — laziness must survive a hook that
//     logs at debug, which is the common case once a record exists at all.
//  3. Positive control: emit a warn record. The file must now appear. Without
//     this phase the absence observed above would be indistinguishable from a
//     guard pointed at the wrong path, where nothing would ever appear and both
//     earlier phases would pass vacuously.
//
// Falsification: open the file in newHookSink instead of on first Write and
// phase 1 fails while the test is still listed.
//
// Not parallel — t.Setenv.
func TestHookSinkIsNotCreatedWithoutRecords(t *testing.T) {
	root := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	// Explicitly empty: phase 2 depends on the default minimum level (warn)
	// rejecting a debug record, and an ambient MOAI_LOG_LEVEL must not lower it.
	t.Setenv(config.EnvLogLevel, "")

	sink := filepath.Join(root, hookRuntimeLogRelPath)
	logDir := filepath.Dir(sink)

	assertAbsent := func(phase string) {
		t.Helper()
		if _, err := os.Stat(sink); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s: sink %s exists (stat err %v), want absent — the file must open on the first admitted record, not before (AC-HDS-012)",
				phase, sink, err)
		}
		if _, err := os.Stat(logDir); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s: directory %s exists (stat err %v), want absent — MkdirAll runs inside the lazy open, so the directory must not be created either (AC-HDS-012)",
				phase, logDir, err)
		}
	}

	// Phase 1 — installed, nothing emitted.
	d := resolveLoggingDecision([]string{"hook", "pre-tool"})
	log := slog.New(slog.NewTextHandler(d.dest, &slog.HandlerOptions{Level: d.level}))
	assertAbsent("no record emitted")

	// Phase 2 — a record the level bar rejects never reaches the writer.
	log.Debug("hook-sink-lazy-probe-below-bar")
	assertAbsent("record emitted below the level bar")

	// Phase 3 — positive control.
	log.Warn("hook-sink-lazy-probe-admitted")
	if _, err := os.Stat(sink); err != nil {
		t.Fatalf("positive control: stat %s: %v\n"+
			"an admitted warn record must open the sink — without this the absence asserted above proves nothing about laziness", sink, err)
	}
}
