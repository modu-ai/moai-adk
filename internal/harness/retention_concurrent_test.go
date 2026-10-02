//go:build !windows

package harness

// Multi-process load-regression guard for the retention prune (card t1425).
//
// The hook path builds a fresh Retention in every hook process, so a skip
// interval kept in process memory never suppresses anything: N concurrent hook
// processes each read the whole log, each append the same stale events to the
// monthly archive, and each rewrite the log. These tests re-execute the test
// binary N times (a true multi-process run, not goroutines) and assert that
// exactly ONE process prunes per interval.

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	envPruneHelper  = "MOAI_PRUNE_HELPER"
	envPruneLog     = "MOAI_PRUNE_LOG"
	envPruneArchive = "MOAI_PRUNE_ARCHIVE"
	envPruneNow     = "MOAI_PRUNE_NOW"
	envPruneGate    = "MOAI_PRUNE_GATE"
	envPruneReady   = "MOAI_PRUNE_READY"

	concurrentPruners = 8
	concurrentStale   = 10000 // stale events (split over two months)
	concurrentKept    = 10000 // events inside the retention window
	concurrentTestCap = 60 * time.Second
)

// TestPruneHelperProcess is the child body of the multi-process test. It is a
// no-op in a normal test run; the parent selects it with -test.run and sets
// MOAI_PRUNE_HELPER=1.
func TestPruneHelperProcess(t *testing.T) {
	if os.Getenv(envPruneHelper) != "1" {
		t.Skip("helper process body; only runs when re-executed by the multi-process test")
	}

	now, err := time.Parse(time.RFC3339Nano, os.Getenv(envPruneNow))
	if err != nil {
		t.Fatalf("helper: bad %s: %v", envPruneNow, err)
	}

	// Announce readiness, then spin on the start gate so all children overlap.
	if err := os.WriteFile(os.Getenv(envPruneReady), nil, 0o644); err != nil {
		t.Fatalf("helper: ready file: %v", err)
	}
	gate := os.Getenv(envPruneGate)
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(gate); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("helper: start gate never opened")
		}
		time.Sleep(time.Millisecond)
	}

	r := NewRetention(os.Getenv(envPruneLog), os.Getenv(envPruneArchive), func() time.Time { return now })
	if err := r.PruneStaleEntries(30); err != nil {
		t.Fatalf("helper: PruneStaleEntries: %v", err)
	}
}

// runPruneWave re-executes the test binary n times against the same log and
// archive with the given fixed clock, opens the start gate once every child
// reports ready, and waits for all of them. It fails the test on any child
// error.
func runPruneWave(ctx context.Context, t *testing.T, n int, logPath, archiveDir string, now time.Time, label string) {
	t.Helper()

	scratch := t.TempDir()
	gate := filepath.Join(scratch, "gate")
	cmds := make([]*exec.Cmd, 0, n)
	outs := make([]*bytes.Buffer, 0, n)

	t.Cleanup(func() {
		for _, c := range cmds {
			if c.Process != nil {
				_ = c.Process.Kill()
			}
		}
	})

	for i := 0; i < n; i++ {
		ready := filepath.Join(scratch, "ready-"+strconv.Itoa(i))
		c := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPruneHelperProcess$", "-test.count=1")
		c.Env = append(os.Environ(),
			envPruneHelper+"=1",
			envPruneLog+"="+logPath,
			envPruneArchive+"="+archiveDir,
			envPruneNow+"="+now.Format(time.RFC3339Nano),
			envPruneGate+"="+gate,
			envPruneReady+"="+ready,
		)
		buf := &bytes.Buffer{}
		c.Stdout, c.Stderr = buf, buf
		if err := c.Start(); err != nil {
			t.Fatalf("%s: start child %d: %v", label, i, err)
		}
		cmds = append(cmds, c)
		outs = append(outs, buf)
	}

	// Wait until every child is parked at the gate, then open it.
	for {
		ready := 0
		for i := 0; i < n; i++ {
			if _, err := os.Stat(filepath.Join(scratch, "ready-"+strconv.Itoa(i))); err == nil {
				ready++
			}
		}
		if ready == n {
			break
		}
		if ctx.Err() != nil {
			t.Fatalf("%s: children never became ready: %v", label, ctx.Err())
		}
		time.Sleep(time.Millisecond)
	}
	if err := os.WriteFile(gate, nil, 0o644); err != nil {
		t.Fatalf("%s: open gate: %v", label, err)
	}

	for i, c := range cmds {
		if err := c.Wait(); err != nil {
			t.Errorf("%s: child %d failed: %v\n%s", label, i, err, outs[i].String())
		}
	}
}

// archiveReport is what a scan of the archive directory observed.
type archiveReport struct {
	streams    map[string]int // month file -> gzip member count
	bytes      int            // total archive size on disk (informational)
	headers    int            // gzip header magic occurrences (informational, upper-bound hint of appended streams)
	subjects   map[string]int // event subject -> occurrences across the archive
	corruption []string
}

// scanArchive counts gzip members per archive file (Multistream disabled so
// every appended stream is visible) and counts event occurrences by subject.
func scanArchive(t *testing.T, archiveDir string) archiveReport {
	t.Helper()
	rep := archiveReport{streams: map[string]int{}, subjects: map[string]int{}}

	files, err := filepath.Glob(filepath.Join(archiveDir, "*.jsonl.gz"))
	if err != nil {
		t.Fatalf("glob archive: %v", err)
	}
	for _, path := range files {
		name := filepath.Base(path)
		if raw, err := os.ReadFile(path); err == nil {
			rep.bytes += len(raw)
			rep.headers += bytes.Count(raw, []byte{0x1f, 0x8b, 0x08})
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open archive %s: %v", name, err)
		}
		br := bufio.NewReader(f)
		zr, err := gzip.NewReader(br)
		if err != nil {
			rep.corruption = append(rep.corruption, fmt.Sprintf("%s: first stream: %v", name, err))
			_ = f.Close()
			continue
		}
		for {
			zr.Multistream(false)
			body, err := io.ReadAll(zr)
			if err != nil {
				rep.corruption = append(rep.corruption, fmt.Sprintf("%s: stream %d: %v", name, rep.streams[name]+1, err))
				break
			}
			rep.streams[name]++
			sc := bufio.NewScanner(bytes.NewReader(body))
			sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
			for sc.Scan() {
				var evt Event
				if err := json.Unmarshal(sc.Bytes(), &evt); err != nil {
					rep.corruption = append(rep.corruption, fmt.Sprintf("%s: bad event line: %v", name, err))
					continue
				}
				rep.subjects[evt.Subject]++
			}
			if err := zr.Reset(br); err != nil {
				if !errors.Is(err, io.EOF) {
					rep.corruption = append(rep.corruption, fmt.Sprintf("%s: next stream: %v", name, err))
				}
				break
			}
		}
		_ = f.Close()
	}
	return rep
}

// logSubjects returns the subject -> occurrence map of the live log.
func logSubjects(t *testing.T, logPath string) map[string]int {
	t.Helper()
	f, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer func() { _ = f.Close() }()
	out := map[string]int{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			t.Fatalf("log line does not parse: %v", err)
		}
		out[evt.Subject]++
	}
	return out
}

func leftoverTmpFiles(t *testing.T, dir string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(dir, "usage-log-*.tmp"))
	if err != nil {
		t.Fatalf("glob tmp: %v", err)
	}
	return m
}

func archiveStreamTotal(rep archiveReport) int {
	total := 0
	for _, n := range rep.streams {
		total += n
	}
	return total
}

func archiveDuplicates(rep archiveReport) int {
	dups := 0
	for _, n := range rep.subjects {
		if n > 1 {
			dups += n - 1
		}
	}
	return dups
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestPruneConcurrentProcessesSingleRewrite is the load-regression guard: N
// concurrent hook-like processes must produce ONE archive stream per month,
// no duplicate archived event, one clean log, and a second wave inside the
// skip interval must change nothing.
func TestPruneConcurrentProcessesSingleRewrite(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), concurrentTestCap)
	defer cancel()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage-log.jsonl")
	archiveDir := filepath.Join(dir, "archive")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)

	// Fixture: half the events are older than the 30-day cutoff (2026-09-02),
	// spread over two months so two archive files are expected; the other half
	// is inside the window.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	wantStale := map[string]bool{}
	wantKept := map[string]bool{}
	for i := 0; i < concurrentStale; i++ {
		month := time.July
		if i%2 == 1 {
			month = time.August
		}
		subject := fmt.Sprintf("stale-%05d", i)
		wantStale[subject] = true
		if err := enc.Encode(Event{
			Timestamp:     time.Date(2026, month, 1+i%27, 3, i%60, 0, 0, time.UTC),
			EventType:     EventTypeMoaiSubcommand,
			Subject:       subject,
			ContextHash:   "ctx-" + strconv.Itoa(i),
			SchemaVersion: LogSchemaVersion,
		}); err != nil {
			t.Fatalf("encode stale: %v", err)
		}
	}
	for i := 0; i < concurrentKept; i++ {
		subject := fmt.Sprintf("kept-%05d", i)
		wantKept[subject] = true
		if err := enc.Encode(Event{
			Timestamp:     time.Date(2026, time.September, 10+i%20, 5, i%60, 0, 0, time.UTC),
			EventType:     EventTypeAgentInvocation,
			Subject:       subject,
			ContextHash:   "ctx-" + strconv.Itoa(i),
			SchemaVersion: LogSchemaVersion,
		}); err != nil {
			t.Fatalf("encode kept: %v", err)
		}
	}
	if err := os.WriteFile(logPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// Wave 1: N processes race to prune.
	runPruneWave(ctx, t, concurrentPruners, logPath, archiveDir, now, "wave1")

	rep := scanArchive(t, archiveDir)
	streams := archiveStreamTotal(rep)
	dups := archiveDuplicates(rep)
	t.Logf("RESULT wave1 streams=%d dups=%d archive_files=%d corruption=%d archive_bytes=%d gzip_headers=%d", streams, dups, len(rep.streams), len(rep.corruption), rep.bytes, rep.headers)

	if len(rep.corruption) > 0 {
		t.Errorf("archive corruption: %v", rep.corruption)
	}
	if len(rep.streams) != 2 {
		t.Errorf("archive files: want 2 (one per month), got %v", sortedKeys(rep.streams))
	}
	for name, n := range rep.streams {
		if n != 1 {
			t.Errorf("archive %s: want exactly 1 gzip stream, got %d", name, n)
		}
	}
	if dups != 0 {
		t.Errorf("archive holds %d duplicate events, want 0", dups)
	}
	for subject := range wantStale {
		if rep.subjects[subject] != 1 {
			t.Errorf("stale event %s archived %d times, want 1", subject, rep.subjects[subject])
			break
		}
	}
	for subject := range rep.subjects {
		if !wantStale[subject] {
			t.Errorf("non-stale event %s reached the archive", subject)
			break
		}
	}

	live := logSubjects(t, logPath)
	if len(live) != concurrentKept {
		t.Errorf("log holds %d distinct events, want exactly the %d kept ones", len(live), concurrentKept)
	}
	for subject, n := range live {
		if !wantKept[subject] || n != 1 {
			t.Errorf("log event %s occurs %d times (kept=%v), want a kept event exactly once", subject, n, wantKept[subject])
			break
		}
	}
	if tmp := leftoverTmpFiles(t, dir); len(tmp) != 0 {
		t.Errorf("leftover temp files: %v", tmp)
	}

	// Wave 2: inside the skip interval a new stale event must survive, and the
	// archive must not grow, because the stamp written by wave 1 is respected
	// by brand-new processes.
	extra, err := json.Marshal(Event{
		Timestamp:     time.Date(2026, time.June, 5, 1, 0, 0, 0, time.UTC),
		EventType:     EventTypeFeedback,
		Subject:       "stale-wave2-extra",
		ContextHash:   "ctx-extra",
		SchemaVersion: LogSchemaVersion,
	})
	if err != nil {
		t.Fatalf("marshal extra: %v", err)
	}
	lf, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open log for append: %v", err)
	}
	if _, err := lf.Write(append(extra, '\n')); err != nil {
		t.Fatalf("append extra: %v", err)
	}
	_ = lf.Close()

	beforeLog, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	runPruneWave(ctx, t, concurrentPruners, logPath, archiveDir, now.Add(10*time.Minute), "wave2")

	rep2 := scanArchive(t, archiveDir)
	afterLog, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	t.Logf("RESULT wave2 streams=%d dups=%d log_unchanged=%v", archiveStreamTotal(rep2), archiveDuplicates(rep2), bytes.Equal(beforeLog, afterLog))
	if got := archiveStreamTotal(rep2); got != streams {
		t.Errorf("second wave changed the archive stream count: %d -> %d", streams, got)
	}
	if !bytes.Equal(beforeLog, afterLog) {
		t.Errorf("second wave rewrote the log inside the skip interval")
	}
	if tmp := leftoverTmpFiles(t, dir); len(tmp) != 0 {
		t.Errorf("leftover temp files after wave 2: %v", tmp)
	}
}
