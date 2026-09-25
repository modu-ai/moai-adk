package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// concurrentProbeMsg is the marker every record in the concurrency guard
// carries. It appears exactly once per intact line, so a line holding two of
// them is a torn record rather than a record.
const concurrentProbeMsg = "hook-sink-concurrent-probe"

// concurrentProbeLine matches one INTACT record: the whole line, anchored at
// both ends, with the writer and sequence it was emitted with.
//
// Full-line anchoring is the assertion. A substring search would pass against a
// line that merely contains a record — which is exactly what a torn write looks
// like, since the surviving fragment of one record still contains the fragment
// of another.
var concurrentProbeLine = regexp.MustCompile(
	`^time=\S+ level=WARN msg=` + concurrentProbeMsg + ` writer=(\d+) seq=(\d+) filler=\S+$`)

// TestHookSinkConcurrentAppendIsIntact is the concurrency guard for AC-HDS-003
// (REQ-HDS-006): N writers open on the SAME sink file, each emitting M
// identifiable records, produce N×M lines and no line carrying a fragment of
// two records.
//
// The arrangement models what actually happens in production: hook processes
// are one-shot and several can be in flight at once, so several independent
// writers — each with its own file handle, none aware of the others — append to
// one path. Separate hookSink values rather than one shared value is the point;
// a single value would exercise sync.Once, not O_APPEND.
//
// What this guard does NOT assert, deliberately: LINE ORDER. plan.md §B-3
// promises only that a concurrent writer's record is not corrupted, and records
// the reason — O_APPEND's atomicity is per write, so the promise holds by
// emitting each record in exactly one write, and says nothing about which
// writer reaches the file first. A guard requiring order would fail for a
// reason the design never claimed to prevent, and would be read as a defect in
// the sink rather than in the guard.
//
// Two residual risks stay out of scope for the same reason (plan.md §B-3, §E
// R3): a record long enough to exceed the platform's atomic-write bound, and
// Windows, where O_APPEND is emulated. The filler below keeps records
// comfortably inside the bound rather than probing it.
//
// Falsification: split the per-record write in hookSink.Write into two calls
// (write p[:len(p)/2] then the rest) and this test fails on torn lines while
// still being listed.
//
// Not parallel — t.Setenv.
func TestHookSinkConcurrentAppendIsIntact(t *testing.T) {
	const (
		writers        = 8
		perWriter      = 25
		expectedRecord = writers * perWriter
	)

	root := t.TempDir()
	t.Setenv(config.EnvClaudeProjectDir, root)
	// Explicitly empty: the default minimum level (warn) admits the records
	// below, and an ambient MOAI_LOG_LEVEL must not decide that.
	t.Setenv(config.EnvLogLevel, "")

	// Long enough that a torn write would be visible, short enough to stay well
	// inside any platform's atomic-append bound (plan.md §B-3 residual risk).
	filler := strings.Repeat("x", 256)

	var start sync.WaitGroup
	var done sync.WaitGroup
	start.Add(1)
	for w := range writers {
		done.Add(1)
		go func(w int) {
			defer done.Done()
			// Each goroutine resolves its own sink, as a separate hook process
			// would: its own handle on the same path.
			d := resolveLoggingDecision([]string{"hook", "pre-tool"})
			log := slog.New(slog.NewTextHandler(d.dest, &slog.HandlerOptions{Level: d.level}))
			start.Wait()
			for seq := range perWriter {
				log.Warn(concurrentProbeMsg, "writer", w, "seq", seq, "filler", filler)
			}
		}(w)
	}
	start.Done()
	done.Wait()

	sink := filepath.Join(root, hookRuntimeLogRelPath)
	raw, err := os.ReadFile(filepath.Clean(sink))
	if err != nil {
		t.Fatalf("read hook sink %s: %v\n"+
			"%d concurrent writers must have produced a readable sink (AC-HDS-003)", sink, err, writers)
	}

	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if got := len(lines); got != expectedRecord {
		t.Errorf("hook sink holds %d line(s), want %d (%d writers x %d records)",
			got, expectedRecord, writers, perWriter)
	}

	seen := make(map[string]int, expectedRecord)
	for i, line := range lines {
		if n := strings.Count(line, concurrentProbeMsg); n != 1 {
			// 0 is a torn record's tail, >1 is two records sharing a line —
			// the same defect seen from opposite ends.
			shape := "the line holds fragments of more than one record"
			if n == 0 {
				shape = "the line is the tail of a torn record"
			}
			t.Errorf("line %d carries the probe marker %d time(s), want exactly 1 — %s:\n%s", i, n, shape, line)
			continue
		}
		m := concurrentProbeLine.FindStringSubmatch(line)
		if m == nil {
			t.Errorf("line %d is not one intact record (want a full match of %s):\n%s",
				i, concurrentProbeLine, line)
			continue
		}
		seen[m[1]+"/"+m[2]]++
	}

	// Every (writer, seq) pair must appear exactly once. A count of 0 means a
	// record was lost; a count above 1 means one was duplicated. Both are
	// corruption of a kind the line-shape check above cannot see.
	var missing, duplicated []string
	for w := range writers {
		for seq := range perWriter {
			key := fmt.Sprintf("%d/%d", w, seq)
			switch n := seen[key]; {
			case n == 0:
				missing = append(missing, key)
			case n > 1:
				duplicated = append(duplicated, fmt.Sprintf("%s x%d", key, n))
			}
		}
	}
	if len(missing) > 0 {
		t.Errorf("%d writer/seq record(s) absent from the sink: %s", len(missing), strings.Join(missing, " "))
	}
	if len(duplicated) > 0 {
		t.Errorf("%d writer/seq record(s) appear more than once: %s", len(duplicated), strings.Join(duplicated, " "))
	}
}
