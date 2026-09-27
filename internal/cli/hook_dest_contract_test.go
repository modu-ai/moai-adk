package cli

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestHookPathHookDestIsNeitherStdStream is the contract guard for AC-HDS-004
// and AC-HDS-009 (REQ-HDS-002, REQ-HDS-005, REQ-HDS-013): the `moai hook`
// path's logging destination is neither standard stream, and MOAI_LOG_LEVEL
// cannot make it one.
//
// Three axes, and the third is the one most easily lost:
//
//	(1) DECISION VALUE — the resolved dest is neither os.Stdout nor os.Stderr by
//	    pointer identity, across both hook arg shapes and every MOAI_LOG_LEVEL
//	    value. Deliberately NOT a runtime byte comparison of the two streams:
//	    under `go test` this repo's test binary receives ONE *os.File for both,
//	    measured as dest == os.Stdout == os.Stderr, so a stream-identity check
//	    reports a false failure while the production selection is correct. The
//	    comment on TestLoggingNeverTargetsStdout records that measurement.
//	(2) SOURCE TOKENS, per file — the rule differs by file, and applying one
//	    rule to both is the wrong answer: logging.go's non-hook branch writes to
//	    os.Stderr legitimately, so only os.Stdout is barred there; the sink
//	    writer's file is hook-path-only, so both are barred.
//	(3) FILE-SET MEMBERSHIP — the set is DERIVED at test time, never listed.
//	    See discoverHookDestFiles: membership of the sink writer's file holds by
//	    construction rather than by an assertion that a hardcoded list contains
//	    what it lists, and an unreadable candidate fails instead of being skipped.
//
// The pre-existing TestLoggingNeverTargetsStdout keeps its narrow scope; this
// guard carries the widened set. The overlap on logging.go / os.Stdout is a
// second look, not a cost.
//
// Not parallel — subtests call t.Setenv.
func TestHookPathHookDestIsNeitherStdStream(t *testing.T) {
	argShapes := map[string][]string{
		"bare":         {"hook", "pre-tool"},
		"behind_aflag": {"--verbose", "hook", "stop"},
	}
	// "" is the unset case; the four named values are AC-HDS-009's table.
	levels := []string{"", "DEBUG", "INFO", "WARN", "ERROR"}

	for shape, args := range argShapes {
		for _, level := range levels {
			name := shape + "_level_" + level
			if level == "" {
				name = shape + "_level_unset"
			}
			t.Run(name, func(t *testing.T) {
				t.Setenv(config.EnvLogLevel, level)
				t.Setenv(config.EnvClaudeProjectDir, t.TempDir())

				got := resolveLoggingDecision(args).dest
				if got == nil {
					t.Fatalf("resolveLoggingDecision(%q).dest is nil; the hook path must resolve a writer", args)
				}
				if got == os.Stdout {
					t.Errorf("resolveLoggingDecision(%q).dest with %s=%q is os.Stdout; "+
						"stdout carries the hook's structured JSON contract (REQ-HDS-002)",
						args, config.EnvLogLevel, level)
				}
				if got == os.Stderr {
					t.Errorf("resolveLoggingDecision(%q).dest with %s=%q is os.Stderr; "+
						"stderr is read by the Claude Code runtime (REQ-HDS-002)",
						args, config.EnvLogLevel, level)
				}
			})
		}
	}

	t.Run("source_tokens_absent_per_file", func(t *testing.T) {
		for path, forbidden := range discoverHookDestFiles(t) {
			src, err := os.ReadFile(filepath.Clean(path))
			if err != nil {
				// A listed file that cannot be read FAILS. Skipping it would
				// empty the swept set, and a check that swept nothing reports
				// success identically to one that passed everything.
				t.Fatalf("read %s: %v (a file in the hook-destination set must fail, not be skipped)", path, err)
			}
			for _, token := range forbidden {
				if bytes.Contains(src, []byte(token)) {
					t.Errorf("%s references %s; it is part of the hook path's logging destination, "+
						"which must reach neither standard stream (REQ-HDS-002, AC-HDS-004)", path, token)
				}
			}
		}
	})
}

// discoverHookDestFiles returns the non-test Go files that constitute the hook
// path's logging destination, mapped to the source tokens each one forbids.
//
// The set is DISCOVERED, never listed. Both members are found by content:
//
//   - the decision site, by the declaration of resolveLoggingDecision;
//   - the sink writer's file, by the sink path constant's declaration, which
//     AC-HDS-011 fixes at exactly one non-test occurrence.
//
// That is what makes membership non-circular. A hardcoded list asserting it
// contains what it lists proves nothing, and — the defect this SPEC was written
// from — silently omits a file added later. Anchoring on the constant's
// declaration also means the set follows a rename of the file.
//
// Every failure mode here is fatal rather than skipped: an unreadable
// candidate, a swept population of zero, the constant occurring zero or more
// than once. A sweep that selected nothing exits green, which is
// indistinguishable from one where everything passed.
func discoverHookDestFiles(t *testing.T) map[string][]string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}

	var swept int
	var decisionSites, sinkSites []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, readErr := os.ReadFile(filepath.Clean(name))
		if readErr != nil {
			t.Fatalf("read candidate %s: %v (an unreadable candidate must fail, not be skipped — "+
				"skipping it would shrink the swept set silently)", name, readErr)
		}
		swept++
		if bytes.Contains(src, []byte("func resolveLoggingDecision(")) {
			decisionSites = append(decisionSites, name)
		}
		if bytes.Contains(src, []byte(strconv.Quote(hookRuntimeLogRelPath))) {
			sinkSites = append(sinkSites, name)
		}
	}

	if swept == 0 {
		t.Fatal("swept 0 non-test Go files in the package directory; " +
			"a check whose swept set is empty asserts nothing")
	}
	if len(decisionSites) != 1 {
		t.Fatalf("found %d non-test file(s) declaring resolveLoggingDecision (%v), want exactly 1; "+
			"the hook-destination set cannot be derived", len(decisionSites), decisionSites)
	}
	if len(sinkSites) != 1 {
		t.Fatalf("found %d non-test file(s) carrying the sink path literal %q (%v), want exactly 1; "+
			"AC-HDS-011 fixes it at one declaration and this set is anchored on it",
			len(sinkSites), hookRuntimeLogRelPath, sinkSites)
	}
	if decisionSites[0] == sinkSites[0] {
		t.Fatalf("the decision site and the sink writer both resolve to %s; the per-file token rule is "+
			"asymmetric (the decision site writes os.Stderr on its non-hook branch legitimately, the sink "+
			"writer may reference neither stream) and cannot be applied to a single file", decisionSites[0])
	}

	return map[string][]string{
		decisionSites[0]: {"os.Stdout"},
		sinkSites[0]:     {"os.Stdout", "os.Stderr"},
	}
}

// TestHookPathSinkFailureIsFailOpen is the degradation guard for AC-HDS-005 and
// AC-HDS-006 (REQ-HDS-003, REQ-HDS-014): when the sink cannot be resolved or
// opened, the hook path degrades to discarding and keeps running.
//
// The precedent is internal/config/log.go logTierReadFailure, which downgrades
// a failed log open to a warning and never panics. The contract asserted here
// is the writer's: Write reports the full length and a nil error on every
// failure path, so slog's caller never learns the record was dropped and never
// retries.
//
// Not parallel — subtests call t.Setenv.
func TestHookPathSinkFailureIsFailOpen(t *testing.T) {
	t.Run("unresolvable_root_degrades_to_discard", func(t *testing.T) {
		// A regular file stands where a directory would have to be, so the
		// sink's MkdirAll cannot succeed.
		blocker := filepath.Join(t.TempDir(), "not-a-directory")
		if err := os.WriteFile(blocker, []byte("occupied"), 0o600); err != nil {
			t.Fatalf("write blocker file: %v", err)
		}
		root := filepath.Join(blocker, "nested-root")
		t.Setenv(config.EnvClaudeProjectDir, root)
		t.Setenv(config.EnvLogLevel, "")

		dest := emitWarnWithoutPanic(t, []string{"hook", "pre-tool"})
		assertFailOpenWrite(t, dest)

		// Nothing may have been created under the blocked root. The stat error
		// is deliberately not pinned to a specific errno — a path under a
		// regular file reports ENOTDIR on this platform and ENOENT elsewhere;
		// what the AC asserts is that no tree was fabricated.
		if _, err := os.Stat(root); err == nil {
			t.Errorf("os.Stat(%s) succeeded; an unresolvable root must degrade to discarding "+
				"rather than fabricate a tree (AC-HDS-005)", root)
		}
		occupant, err := os.ReadFile(filepath.Clean(blocker))
		if err != nil {
			t.Fatalf("stat the blocking file %s: %v", blocker, err)
		}
		if string(occupant) != "occupied" {
			t.Errorf("%s now holds %q, want %q; a failed resolution must not write through "+
				"whatever stands in the way (AC-HDS-005)", blocker, occupant, "occupied")
		}
	})

	t.Run("empty_root_degrades_to_discard", func(t *testing.T) {
		// resolveHookProjectRoot returns "" when neither CLAUDE_PROJECT_DIR nor
		// os.Getwd() resolves. The sink accepts it and degrades on first write
		// rather than failing at construction.
		assertFailOpenWrite(t, newHookSink(""))
	})

	t.Run("sink_path_preempted_by_directory", func(t *testing.T) {
		root := t.TempDir()
		sink := filepath.Join(root, hookRuntimeLogRelPath)
		if err := os.MkdirAll(sink, 0o750); err != nil {
			t.Fatalf("pre-empt sink path with a directory: %v", err)
		}
		t.Setenv(config.EnvClaudeProjectDir, root)
		t.Setenv(config.EnvLogLevel, "")

		dest := emitWarnWithoutPanic(t, []string{"hook", "stop"})
		assertFailOpenWrite(t, dest)

		info, err := os.Stat(sink)
		if err != nil {
			t.Fatalf("stat pre-empted sink %s: %v", sink, err)
		}
		if !info.IsDir() {
			t.Errorf("%s is no longer a directory; a failed open must degrade, never clobber "+
				"what occupies the path (AC-HDS-006)", sink)
		}
	})
}

// emitWarnWithoutPanic resolves the hook path's logging decision, emits one warn
// record through it, and reports the destination. A panic is converted into a
// failure naming the contract it broke: a hook must not die because its logging
// could not open (REQ-HDS-003).
func emitWarnWithoutPanic(t *testing.T, args []string) io.Writer {
	t.Helper()

	d := resolveLoggingDecision(args)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("emitting a warn record on the hook path panicked: %v\n"+
					"a sink that cannot be opened must degrade to discarding, never panic (AC-HDS-006)", r)
			}
		}()
		slog.New(slog.NewTextHandler(d.dest, &slog.HandlerOptions{Level: d.level})).
			Warn("hook-sink-fail-open-probe", "probe", "guard-3")
	}()
	return d.dest
}

// assertFailOpenWrite checks the writer's fail-open contract directly: the full
// length and a nil error, even though nothing was written anywhere. A short
// count would make slog's caller believe the record mattered enough to retry.
func assertFailOpenWrite(t *testing.T, dest io.Writer) {
	t.Helper()

	probe := []byte("fail-open contract probe\n")
	n, err := dest.Write(probe)
	if err != nil {
		t.Errorf("Write on a degraded sink returned err = %v, want nil; the sink must never report "+
			"a logging failure upward (REQ-HDS-003)", err)
	}
	if n != len(probe) {
		t.Errorf("Write on a degraded sink returned n = %d, want %d (the full length)", n, len(probe))
	}
}
