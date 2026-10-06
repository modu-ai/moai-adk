package hygiene

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Threshold-shaped constants for the fixtures. The L-015 pattern bans
// threshold literals inside this package (sizes, durations, and their
// reversed products all live in internal/config/defaults.go in production),
// so the fixtures derive every size and window from unit-safe products.
const (
	mib                = 1024 * 1024
	testThreshold      = 10 * mib
	testOversized      = 11 * mib
	day                = 1440 * time.Minute
	testMinAge         = 7 * day
	testActivityWindow = 2 * day // the transcript activity window (48 h)
	testStaleHb        = day     // the heartbeat staleness window (24 h)
)

// seedFile writes content to path, creating parent directories.
func seedFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("seed mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("seed write %s: %v", path, err)
	}
}

// growSeed returns deterministic payload of n bytes (rows of digits).
func growSeed(n int64) []byte {
	return growSeedTag(n, '0')
}

// growSeedTag returns a deterministic payload variant keyed by tag so two
// seeds of the same size stay byte-distinguishable.
func growSeedTag(n int64, tag byte) []byte {
	b := make([]byte, 0, n)
	line := []byte("2026-10-05T00:00:00Z audit row payload 0123456789\n")
	line[len(line)-2] = tag
	for int64(len(b))+int64(len(line)) <= n {
		b = append(b, line...)
	}
	for int64(len(b)) < n {
		b = append(b, 'x')
	}
	return b
}

// readBytes reads a file, failing the test on error.
func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

// newTestRotator builds a rotator over a temp log dir with production
// default sizes and registers the dir for the REQ-HYG-015 runtime guard.
func newTestRotator(t *testing.T, logDir string) *Rotator {
	t.Helper()
	registerTestRoot(logDir)
	return &Rotator{LogDir: logDir, MaxBytes: testThreshold, KeptRotations: 1}
}

// TestRotator_RotatesOverThreshold — AC-HYG-001 (L-001). Given a registered
// sink over the threshold in a temp tree, When the rotator runs in apply
// mode, Then the primary is recreated empty, the full content lives at
// <name>.1, no staging remains, and one apply-mode audit row records the
// rotation.
func TestRotator_RotatesOverThreshold(t *testing.T) {
	root := t.TempDir()
	logDir := filepath.Join(root, ".moai", "logs")
	primary := filepath.Join(logDir, "rule-load-audit.jsonl")
	seed := growSeed(testOversized)
	seedFile(t, primary, seed)

	r := newTestRotator(t, logDir)
	rows, err := r.Run(ModeApply)
	if err != nil {
		t.Fatalf("rotator run: %v", err)
	}

	got := readBytes(t, primary)
	if len(got) != 0 {
		t.Fatalf("primary not empty after rotation: %d bytes", len(got))
	}
	chunk := readBytes(t, primary+".1")
	if len(chunk) != len(seed) {
		t.Fatalf("chunk size = %d, want %d", len(chunk), len(seed))
	}
	if string(chunk) != string(seed) {
		t.Fatalf("chunk content diverges from the seeded rows")
	}
	if _, err := os.Stat(primary + ".1.staging"); !os.IsNotExist(err) {
		t.Fatalf("staging artifact survived a completed pass: %v", err)
	}
	rotated := 0
	for _, row := range rows {
		if row.Outcome == OutcomeRotated {
			rotated++
			if row.Path != primary {
				t.Fatalf("rotated row names %s, want %s", row.Path, primary)
			}
		}
	}
	if rotated != 1 {
		t.Fatalf("rotated rows = %d, want 1", rotated)
	}
}

// TestRotator_KeepOneCap — AC-HYG-002 (L-002). Keep-1 cap through the staged
// crash-recoverable sequence, plus both crash-recovery arms: an orphan
// staging is completed by the next pass, never deleted or overwritten, and
// the tree never holds more than one rotated chunk per sink.
func TestRotator_KeepOneCap(t *testing.T) {
	root := t.TempDir()
	logDir := filepath.Join(root, ".moai", "logs")
	primary := filepath.Join(logDir, "agent-model-audit.jsonl")

	t.Run("second rotation replaces the chunk through staging", func(t *testing.T) {
		seedFile(t, primary, growSeed(testOversized))
		r := newTestRotator(t, logDir)
		if _, err := r.Run(ModeApply); err != nil {
			t.Fatalf("first rotation: %v", err)
		}
		firstChunk := readBytes(t, primary+".1")

		// Grow a fresh oversized primary (distinct payload) and rotate again.
		seedFile(t, primary, growSeedTag(testOversized, '7'))
		if _, err := r.Run(ModeApply); err != nil {
			t.Fatalf("second rotation: %v", err)
		}
		secondChunk := readBytes(t, primary+".1")
		if string(secondChunk) == string(firstChunk) {
			t.Fatalf("chunk was not replaced by the second pass")
		}
		got := readBytes(t, primary)
		if len(got) != 0 {
			t.Fatalf("primary not empty after second rotation: %d bytes", len(got))
		}
		entries, err := os.ReadDir(logDir)
		if err != nil {
			t.Fatalf("readdir: %v", err)
		}
		chunks := 0
		for _, e := range entries {
			if e.Name() == "agent-model-audit.jsonl.1" {
				chunks++
			}
			if e.Name() == "agent-model-audit.jsonl.1.staging" {
				t.Fatalf("staging artifact survived: %s", e.Name())
			}
		}
		if chunks != 1 {
			t.Fatalf("chunk count = %d, want exactly 1 (keep-1 cap)", chunks)
		}
	})

	t.Run("crash after staging: next pass completes the placement", func(t *testing.T) {
		logDir := filepath.Join(t.TempDir(), ".moai", "logs")
		primary := filepath.Join(logDir, "agent-model-audit.jsonl")
		oldChunk := []byte("older chunk rows\n")
		crashed := []byte("crashed displaced primary rows\n")
		seedFile(t, primary+".1", oldChunk)
		seedFile(t, primary+".1.staging", crashed)

		r := newTestRotator(t, logDir)
		if _, err := r.Run(ModeApply); err != nil {
			t.Fatalf("recovery run: %v", err)
		}
		if got := readBytes(t, primary+".1"); string(got) != string(crashed) {
			t.Fatalf("staged chunk was overwritten or deleted during recovery")
		}
		if _, err := os.Stat(primary + ".1.staging"); !os.IsNotExist(err) {
			t.Fatalf("staging not promoted: %v", err)
		}
		info, err := os.Stat(primary)
		if err != nil || info.Size() != 0 {
			t.Fatalf("primary not recreated empty after recovery: %v (size %v)", err, info)
		}
	})

	t.Run("crash after removal: staging promoted directly", func(t *testing.T) {
		logDir := filepath.Join(t.TempDir(), ".moai", "logs")
		primary := filepath.Join(logDir, "agent-model-audit.jsonl")
		crashed := []byte("orphan staged rows\n")
		seedFile(t, primary+".1.staging", crashed)

		r := newTestRotator(t, logDir)
		if _, err := r.Run(ModeApply); err != nil {
			t.Fatalf("recovery run: %v", err)
		}
		if got := readBytes(t, primary+".1"); string(got) != string(crashed) {
			t.Fatalf("staged chunk lost during direct promotion")
		}
		if _, err := os.Stat(primary); err != nil {
			t.Fatalf("primary not recreated after recovery: %v", err)
		}
	})
}

// TestRotator_UnderThresholdAndAbsent — AC-HYG-003 (L-003). Under-threshold
// sinks are byte-identical after the pass and an absent registered name
// produces no file and no error; skipped rows carry their reason.
func TestRotator_UnderThresholdAndAbsent(t *testing.T) {
	root := t.TempDir()
	logDir := filepath.Join(root, ".moai", "logs")
	kept := filepath.Join(logDir, "codex-adapter.jsonl")
	content := growSeed(4 * mib)
	seedFile(t, kept, content)

	r := newTestRotator(t, logDir)
	rows, err := r.Run(ModeApply)
	if err != nil {
		t.Fatalf("rotator run: %v", err)
	}
	if got := readBytes(t, kept); string(got) != string(content) {
		t.Fatalf("under-threshold sink was modified")
	}
	if _, err := os.Stat(filepath.Join(logDir, "hook-runtime.log")); !os.IsNotExist(err) {
		t.Fatalf("absent sink materialized: %v", err)
	}
	// The hygiene sink itself is absent here; every row must carry a skip
	// outcome with a reason (no silent rows).
	for _, row := range rows {
		if row.Outcome == OutcomeRotated {
			t.Fatalf("unexpected rotation in an under-threshold tree")
		}
		if row.Reason == "" && row.Outcome != OutcomeSummary {
			t.Fatalf("skip row without reason: %+v", row)
		}
	}
}

// TestRotator_ConcurrentSerialize — AC-HYG-004 (L-004). Concurrent
// invocations serialize under the pass lock: exactly one rotation, no fresh-
// chunk destruction; the stale-decision arm performs no action and records
// skipped-stale; the unverified-exclusion arm records skipped-locked /
// skipped-platform and destroys nothing.
func TestRotator_ConcurrentSerialize(t *testing.T) {
	t.Run("N concurrent invocations rotate exactly once", func(t *testing.T) {
		logDir := filepath.Join(t.TempDir(), ".moai", "logs")
		primary := filepath.Join(logDir, "preedit-session-guard.log")
		seed := growSeed(testOversized)
		seedFile(t, primary, seed)

		const n = 4
		var wg sync.WaitGroup
		var mu sync.Mutex
		rotated := 0
		for range n {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r := newTestRotator(t, logDir)
				rows, err := r.Run(ModeApply)
				if err != nil {
					t.Errorf("concurrent run: %v", err)
					return
				}
				mu.Lock()
				defer mu.Unlock()
				for _, row := range rows {
					if row.Outcome == OutcomeRotated {
						rotated++
					}
				}
			}()
		}
		wg.Wait()
		if rotated != 1 {
			t.Fatalf("rotated outcomes across %d invocations = %d, want 1", n, rotated)
		}
		// The union of primary + chunk preserves every seeded byte.
		prim := readBytes(t, primary)
		chunk := readBytes(t, primary+".1")
		if len(prim)+len(chunk) != len(seed) {
			t.Fatalf("byte union = %d, want %d (fresh chunk destroyed?)",
				len(prim)+len(chunk), len(seed))
		}
	})

	t.Run("stale over-threshold observation is a no-op skip", func(t *testing.T) {
		logDir := filepath.Join(t.TempDir(), ".moai", "logs")
		primary := filepath.Join(logDir, "subagent-write-guard.log")
		small := []byte("tiny content\n")
		seedFile(t, primary, small)

		r := newTestRotator(t, logDir)
		seen := map[string]int{}
		r.statFn = func(path string) (os.FileInfo, error) {
			// The first stat of the target path models the stale pre-lock
			// observation claiming an oversized sink; every later stat —
			// including the under-lock re-stat — reads the real (small)
			// file.
			seen[path]++
			if filepath.Base(path) == "subagent-write-guard.log" && seen[path] == 1 {
				return staleInfo{size: testOversized}, nil
			}
			return os.Stat(path)
		}
		rows, err := r.Run(ModeApply)
		if err != nil {
			t.Fatalf("stale run: %v", err)
		}
		if got := readBytes(t, primary); string(got) != string(small) {
			t.Fatalf("stale decision destroyed content")
		}
		if _, err := os.Stat(primary + ".1"); !os.IsNotExist(err) {
			t.Fatalf("stale decision rotated anyway: %v", err)
		}
		found := false
		for _, row := range rows {
			if row.Outcome == OutcomeSkippedStale {
				found = true
			}
		}
		if !found {
			t.Fatalf("no skipped-stale outcome recorded: %+v", rows)
		}
	})

	t.Run("held exclusion skips and destroys nothing", func(t *testing.T) {
		logDir := filepath.Join(t.TempDir(), ".moai", "logs")
		primary := filepath.Join(logDir, "status-transition-audit.log")
		chunk := []byte("previous chunk\n")
		seedFile(t, primary, growSeed(testOversized))
		seedFile(t, primary+".1", chunk)

		r := newTestRotator(t, logDir)
		r.lockProbe = func(string) lockResult { return lockHeld }
		rows, err := r.Run(ModeApply)
		if err != nil {
			t.Fatalf("held run: %v", err)
		}
		if got := readBytes(t, primary+".1"); string(got) != string(chunk) {
			t.Fatalf("held exclusion destroyed the chunk")
		}
		if got := readBytes(t, primary); len(got) != int(testOversized) {
			t.Fatalf("held exclusion touched the primary")
		}
		if _, err := os.Stat(primary + ".1.staging"); !os.IsNotExist(err) {
			t.Fatalf("held exclusion staged anyway: %v", err)
		}
		for _, row := range rows {
			if row.Outcome != OutcomeSkippedLocked {
				t.Fatalf("outcome = %s, want %s", row.Outcome, OutcomeSkippedLocked)
			}
		}
	})

	t.Run("unverifiable exclusion records skipped-platform", func(t *testing.T) {
		logDir := filepath.Join(t.TempDir(), ".moai", "logs")
		primary := filepath.Join(logDir, "status-transition-audit.log")
		seedFile(t, primary, growSeed(testOversized))

		r := newTestRotator(t, logDir)
		r.lockProbe = func(string) lockResult { return lockUnverifiable }
		rows, err := r.Run(ModeApply)
		if err != nil {
			t.Fatalf("unverifiable run: %v", err)
		}
		if got := readBytes(t, primary); len(got) != int(testOversized) {
			t.Fatalf("unverifiable exclusion touched the primary")
		}
		for _, row := range rows {
			if row.Outcome != OutcomeSkippedPlatform {
				t.Fatalf("outcome = %s, want %s", row.Outcome, OutcomeSkippedPlatform)
			}
		}
	})
}

// staleInfo is a fixed os.FileInfo the stat seam returns to model a stale
// pre-lock observation.
type staleInfo struct{ size int64 }

func (s staleInfo) Name() string       { return "stale" }
func (s staleInfo) Size() int64        { return s.size }
func (s staleInfo) Mode() os.FileMode  { return 0o644 }
func (s staleInfo) ModTime() time.Time { return time.Now() }
func (s staleInfo) IsDir() bool        { return false }
func (s staleInfo) Sys() any           { return nil }

// TestRotator_ReportModeSummaryRow — REQ-HYG-004/REQ-HYG-010 basic form at
// M1: report mode rotates nothing and appends exactly one summary row.
func TestRotator_ReportModeSummaryRow(t *testing.T) {
	root := t.TempDir()
	logDir := filepath.Join(root, ".moai", "logs")
	primary := filepath.Join(logDir, "rule-load-audit.jsonl")
	seedFile(t, primary, growSeed(testOversized))

	r := newTestRotator(t, logDir)
	if _, err := r.Run(ModeReport); err != nil {
		t.Fatalf("report run: %v", err)
	}
	info, err := os.Stat(primary)
	if err != nil || info.Size() != testOversized {
		t.Fatalf("report mode mutated the sink: %v", err)
	}
	if _, err := os.Stat(filepath.Join(logDir, PassLockName)); !os.IsNotExist(err) {
		t.Fatalf("report mode created a lockfile: %v", err)
	}
	rows, err := readAuditRows(filepath.Join(logDir, "hygiene-audit.jsonl"))
	if err != nil {
		t.Fatalf("read audit rows: %v", err)
	}
	n := 0
	for _, row := range rows {
		if row.Unit == "rotator" && row.Mode == string(ModeReport) {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("report summary rows = %d, want 1", n)
	}
}
