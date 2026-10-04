// Package harness — Log retention and pruning implementation.
// REQ-HL-011: Archives old events and cleans up log files.
package harness

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// pruneSkipDuration is the duration within which pruning is skipped since last prune.
const pruneSkipDuration = time.Hour

// pruneStateSuffix is appended to the log path to name the prune state file.
// The state file is both the flock target and the carrier of the last prune
// ATTEMPT time (RFC3339Nano text from nowFn, not a file mtime, so tests that
// inject a mock clock stay deterministic).
const pruneStateSuffix = ".prune-state"

// tmpPattern names the temp file the log rewrite creates next to the log; the orphan
// sweep matches the same pattern so the two cannot drift apart.
const tmpPattern = "usage-log-*.tmp"

// orphanTmpMinAge is how old a tmpPattern file must be before the sweep deletes it. A live
// rewrite takes seconds and a hook is killed at 5 s, so a file this old has no writer.
const orphanTmpMinAge = 10 * time.Minute

// maxStampBytes bounds how much of the state file is read; a stamp is ~35 bytes.
//
// @MX:NOTE: [AUTO] Anything longer than this is not a stamp we wrote, so it parses as "no stamp".
const maxStampBytes = 128

// Retention archives and cleans up old entries in usage-log.jsonl.
// REQ-HL-011: Lazy pruning on every RecordEvent call, skip if within 1 hour of last prune.
//
// @MX:ANCHOR: [AUTO] PruneStaleEntries is called by observer and tests.
// @MX:REASON: [AUTO] fan_in >= 3: observer.go, observer_test.go, integration_test.go
type Retention struct {
	// logPath is the usage-log.jsonl file path.
	logPath string

	// archiveDir is the directory to store archive files.
	// Actual path: archiveDir/<YYYY-MM>.jsonl.gz
	archiveDir string

	// lastPruneAt is the last pruning execution time.
	lastPruneAt time.Time

	// nowFn is a function that returns current time (can inject mock-clock in tests).
	nowFn func() time.Time

	// ownerCheck reports whether the entry at a path is owned by the current user, reading the
	// entry's own record without following a symbolic link. NewRetention installs
	// entryOwnedByCurrentUser; tests replace it to reach the foreign-owned branch.
	ownerCheck func(path string) bool
}

// NewRetention creates a Retention instance.
// Uses time.Now if nowFn is nil.
func NewRetention(logPath, archiveDir string, nowFn func() time.Time) *Retention {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Retention{
		logPath:    logPath,
		archiveDir: archiveDir,
		nowFn:      nowFn,
		ownerCheck: entryOwnedByCurrentUser,
	}
}

// PruneStaleEntries removes events older than retentionDays from log
// and adds them to archive file (<YYYY-MM>.jsonl.gz).
// REQ-HL-011: Skips if within 1 hour of the last prune attempt, tracked in memory and on
// disk (<log>.prune-state) so that every new hook process shares one interval.
//
// Events appended while the prune runs are carried, not dropped: the prune classifies only the whole
// lines of the prefix it measured when it opened the log, and just before the rename it reads once
// whatever the log gained after that prefix and copies those bytes after the kept lines. A
// residual window remains between that final tail reading and the rename: an event appended in
// it, or written by a process that had already opened the old file, is lost. Appenders take no
// lock, and none is added here.
//
// Windows: the lock taken on the state file is an in-process mutex, so it gives
// no cross-process exclusion. A burst of hook processes that all find no fresh stamp may prune
// concurrently, once per interval. F5: not reproduced, not measured.
//
// @MX:WARN: [AUTO] The pruner replaces the log by rename; an event appended after its final tail reading is lost.
// @MX:REASON: [AUTO] The state-file flock admits a single pruner per interval (it was N concurrent rewriters),
// but appenders never take that lock, so the tail carry narrows the loss window to the residual window and
// does not close it. The attempt stamp is written before the work, so a killed pruner is not
// repeated until the interval ends. On Windows the lock is in-process only: a burst of hooks that all read
// "no stamp" before the first stamp lands can still prune concurrently, once per interval.
func (r *Retention) PruneStaleEntries(retentionDays int) error {
	now := r.nowFn()

	// Skip prune if within 1 hour (this process)
	if !r.lastPruneAt.IsZero() && now.Sub(r.lastPruneAt) < pruneSkipDuration {
		return nil
	}

	// Skip prune if another process pruned within 1 hour: one tiny read, no lock, no log read.
	statePath := r.logPath + pruneStateSuffix
	if stampIsFresh(readStampFile(statePath), now) {
		r.lastPruneAt = now
		return nil
	}

	// Skip if log file does not exist
	if _, err := os.Stat(r.logPath); os.IsNotExist(err) {
		r.lastPruneAt = now
		return nil
	}

	return r.pruneExclusive(statePath, retentionDays)
}

// pruneExclusive runs the prune while holding an exclusive lock on the state file.
//
// If the state file cannot be created, locked or stamped the prune is SKIPPED and the error
// returned (the observer ignores it): pruning without the lock or without a recorded attempt
// would let every hook process rewrite the log again, which is the storm this guard exists to stop.
//
// Waiting: lock waiters block with no timeout, and lockfile has no try-lock. The harness-observe hooks
// run with a 5 s hook timeout and async: true, and the observer's event is appended before the wait
// begins, so a waiter's delay holds up the hook's exit, not its event. How long a waiter actually
// waits, and whether it is killed at the timeout, was not observed. F6: not reproduced, not measured.
//
// @MX:NOTE: [AUTO] Double-checked locking: the stamp is read again after the lock is won, with a fresh
// clock reading, because the previous holder may have stamped while this process waited.
// @MX:NOTE: [AUTO] Stamp-before-work: a pruner killed after archiving and before the rename leaves the
// stamp, so the same events are archived again at most once per interval, not by every later hook.
// A kill mid-rewrite leaves an orphan usage-log-*.tmp; the lock holder sweeps the old ones
// on the next cycle (sweepOrphanTmp).
func (r *Retention) pruneExclusive(statePath string, retentionDays int) error {
	sf, err := r.openStateFile(statePath)
	if err != nil {
		return err
	}
	defer func() { _ = sf.Close() }()

	if err := lockfile.Lock(sf); err != nil {
		return fmt.Errorf("retention: prune state lock failed: %w", err)
	}
	defer func() { _ = lockfile.Unlock(sf) }()

	return r.pruneLocked(sf, retentionDays)
}

// pruneLocked is the locked phase of the prune: the stamp is re-checked with a fresh clock reading,
// the attempt stamp is written, the log is pruned and the orphan temp files are swept. The caller
// passes the already-opened, locked state file; a stamp write that fails skips the prune.
func (r *Retention) pruneLocked(sf *os.File, retentionDays int) error {
	// Fresh reading: a stale pre-lock "now" would make the holder's newer stamp look like the future.
	now := r.nowFn()
	if stampIsFresh(readStamp(sf), now) {
		r.lastPruneAt = now
		return nil
	}

	// Stamp the ATTEMPT before the work, so a pruner killed mid-way (hook timeout) or a prune that
	// fails leaves the stamp behind and later hooks skip until the interval ends. If the stamp cannot
	// be recorded the prune is skipped: without a recorded attempt every waiter would prune in turn.
	if err := writeStamp(sf, now); err != nil {
		return fmt.Errorf("retention: prune state write failed: %w", err)
	}

	err := r.prune(retentionDays, now)
	r.sweepOrphanTmp(now)
	return err
}

// maxStateInspections bounds how often openStateFile inspects the state path before it gives up.
const maxStateInspections = 3

// openStateFile returns the opened, lockable state file at statePath.
//
// The path is inspected without following links. An absent path is created exclusively. A regular
// file is opened read-write without create and must still be the inspected file. A symbolic link is
// never opened: if the current user owns it, it is replaced by a regular file; otherwise it is left
// byte-identical and the prune is skipped with an error and a warning. A regular file that cannot be
// opened because of a permission error is handled the same way. Any other entry is opened as before,
// so a directory keeps failing the open.
//
// @MX:WARN: [AUTO] This path removes an entry from the log directory.
// @MX:REASON: [AUTO] Removal happens only for an entry the current user owns and only while it is still
// the inspected entry (removeStateEntryIfUnchanged); a concurrent healer's fresh state file must survive.
func (r *Retention) openStateFile(statePath string) (*os.File, error) {
	for range maxStateInspections {
		fi, err := os.Lstat(statePath)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			f, cerr := os.OpenFile(statePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
			if cerr == nil {
				return f, nil
			}
			if errors.Is(cerr, fs.ErrExist) {
				continue // another process created it first: inspect again
			}
			return nil, fmt.Errorf("retention: prune state open failed: %w", cerr)
		case err != nil:
			return nil, fmt.Errorf("retention: prune state open failed: %w", err)
		case fi.Mode()&os.ModeSymlink != 0:
			if err := r.healStateEntry(statePath, fi, "symbolic link"); err != nil {
				return nil, err
			}
		case fi.Mode().IsRegular():
			f, oerr := os.OpenFile(statePath, os.O_RDWR, 0o644)
			if oerr == nil {
				if of, serr := f.Stat(); serr == nil && os.SameFile(fi, of) {
					return f, nil
				}
				_ = f.Close()
				continue // swapped between inspection and open: inspect again
			}
			if errors.Is(oerr, fs.ErrNotExist) {
				continue
			}
			if !errors.Is(oerr, fs.ErrPermission) {
				return nil, fmt.Errorf("retention: prune state open failed: %w", oerr)
			}
			if err := r.healStateEntry(statePath, fi, "file"); err != nil {
				return nil, err
			}
		default:
			f, oerr := os.OpenFile(statePath, os.O_RDWR|os.O_CREATE, 0o644)
			if oerr != nil {
				return nil, fmt.Errorf("retention: prune state open failed: %w", oerr)
			}
			return f, nil
		}
	}
	return nil, fmt.Errorf("retention: prune state entry %s changed on every inspection; prune skipped", statePath)
}

// healStateEntry removes the inspected state-path entry so openStateFile can create a regular
// replacement, but only when the current user owns it. A foreign-owned entry, or one whose owner
// cannot be determined, is left untouched: the prune is skipped with an error and one warning line.
func (r *Retention) healStateEntry(statePath string, inspected os.FileInfo, kind string) error {
	if !r.ownerCheck(statePath) {
		fmt.Fprintf(os.Stderr, "[WARN] harness/retention: prune state %s %s is not owned by the current user or its owner cannot be determined; leaving it untouched and skipping the prune\n", kind, statePath)
		return fmt.Errorf("retention: prune state %s %s is not owned by the current user; prune skipped", kind, statePath)
	}
	if _, err := removeStateEntryIfUnchanged(statePath, inspected); err != nil {
		return fmt.Errorf("retention: prune state %s %s cannot be replaced: %w", kind, statePath, err)
	}
	return nil
}

// removeStateEntryIfUnchanged removes the entry at path only if it is still the inspected one: the
// same file identity, type, permission mode and modification time. It reports whether it removed
// anything; an entry that changed or vanished is left alone, so a state file created by a concurrent
// healer is never removed.
func removeStateEntryIfUnchanged(path string, inspected os.FileInfo) (bool, error) {
	cur, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if !os.SameFile(inspected, cur) || inspected.Mode() != cur.Mode() || !inspected.ModTime().Equal(cur.ModTime()) {
		return false, nil
	}
	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// sweepOrphanTmp deletes usage-log-*.tmp files next to the log that are older than
// orphanTmpMinAge: leftovers of a rewrite whose process was killed before the rename.
// The caller holds the state-file lock, so no other pruner on this machine is rewriting.
//
// Best effort: nothing here fails the prune. Only regular files qualify (a directory or
// symlink with a matching name is not ours). It runs after the prune, so a slow sweep (many
// large leftovers) cannot spend the hook's 5 s budget before the log itself is pruned; a sweep
// cut short by the kill resumes at the next cycle.
func (r *Retention) sweepOrphanTmp(now time.Time) {
	dir := filepath.Dir(r.logPath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if ok, _ := filepath.Match(tmpPattern, e.Name()); !ok || !e.Type().IsRegular() {
			continue
		}
		info, err := e.Info()
		if err != nil || now.Sub(info.ModTime()) < orphanTmpMinAge {
			continue
		}
		_ = os.Remove(filepath.Join(dir, e.Name()))
	}
}

// readStamp returns the (bounded) content of an open state file.
func readStamp(rd io.Reader) []byte {
	raw, err := io.ReadAll(io.LimitReader(rd, maxStampBytes))
	if err != nil {
		return nil
	}
	return raw
}

// readStampFile reads the state file without locking; a missing or unreadable file is "no stamp".
func readStampFile(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	return readStamp(f)
}

// stampIsFresh reports whether raw holds a prune time younger than pruneSkipDuration.
// An empty, partial, or unparsable stamp is "no stamp", and a stamp in the future
// (clock skew, restored file) counts as expired so it can never suppress pruning for long.
func stampIsFresh(raw []byte, now time.Time) bool {
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(raw)))
	if err != nil {
		return false
	}
	age := now.Sub(t)
	return age >= 0 && age < pruneSkipDuration
}

// writeStamp replaces the content of the locked state file with now.
func writeStamp(sf *os.File, now time.Time) error {
	if err := sf.Truncate(0); err != nil {
		return err
	}
	_, err := sf.WriteAt([]byte(now.UTC().Format(time.RFC3339Nano)), 0)
	return err
}

// prune partitions the log at the retention cutoff, archives the stale events, and
// rewrites the log with the kept ones. The caller holds the state-file lock.
func (r *Retention) prune(retentionDays int, now time.Time) error {
	cutoff := now.AddDate(0, 0, -retentionDays)

	// Read log file
	kept, stale, classifiedEnd, err := partitionEvents(r.logPath, cutoff)
	if err != nil {
		return fmt.Errorf("retention: 이벤트 분류 실패: %w", err)
	}

	if len(stale) == 0 {
		// No events to remove
		r.lastPruneAt = now
		return nil
	}

	// Save stale events to monthly archive
	if err := r.archiveEvents(stale); err != nil {
		return fmt.Errorf("retention: 아카이브 실패: %w", err)
	}

	// Overwrite log file with the kept events plus whatever the log gained past the classified prefix
	if err := overwriteWithEvents(r.logPath, kept, classifiedEnd); err != nil {
		return fmt.Errorf("retention: 로그 파일 갱신 실패: %w", err)
	}

	r.lastPruneAt = now
	return nil
}

// logLine is one line that survives a prune. A line that parsed carries its event, which
// is re-encoded on rewrite; a line that did not parse carries only its text (raw), which
// is written back as found (apart from its line terminator, normalized to "\n") so a
// damaged line is never lost to a prune.
type logLine struct {
	evt Event
	raw string
}

// scanTerminatedLines is bufio.ScanLines restricted to lines that end in a newline: a final line
// without its terminator is never returned, and *consumed counts the bytes of the lines returned.
func scanTerminatedLines(consumed *int64) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (int, []byte, error) {
		advance, token, err := bufio.ScanLines(data, atEOF)
		if advance > 0 && data[advance-1] != '\n' {
			return 0, nil, nil
		}
		*consumed += int64(advance)
		return advance, token, err
	}
}

// partitionEvents reads the log file and classifies kept/stale events based on cutoff.
// A line that fails JSON parsing is kept, in file order, as its original text.
//
// Only the prefix the log holds when it is opened is read, and only whole newline-terminated lines
// inside it are classified. classifiedEnd is the byte offset just past the last of them: the bytes
// from there on (a final line without a terminator, and anything appended later) are not classified.
//
// @MX:NOTE: [AUTO] classifiedEnd is the boundary overwriteWithEvents carries the log tail from; a
// line is never split between the classified part and the tail.
func partitionEvents(logPath string, cutoff time.Time) (kept []logLine, stale []Event, classifiedEnd int64, err error) {
	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, 0, nil
		}
		return nil, nil, 0, fmt.Errorf("파일 열기: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("파일 상태: %w", err)
	}

	scanner := bufio.NewScanner(io.NewSectionReader(f, 0, info.Size()))
	scanner.Split(scanTerminatedLines(&classifiedEnd))
	for scanner.Scan() {
		text := scanner.Text()
		line := strings.TrimSpace(text)
		if line == "" {
			continue
		}
		var evt Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			// Put parsing failure lines in kept to prevent data loss
			kept = append(kept, logLine{raw: text})
			continue
		}
		if evt.Timestamp.Before(cutoff) {
			stale = append(stale, evt)
		} else {
			kept = append(kept, logLine{evt: evt})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, 0, fmt.Errorf("파일 스캔: %w", err)
	}
	return kept, stale, classifiedEnd, nil
}

// archiveEvents adds stale events to monthly gzip archives.
// Filename: archiveDir/<YYYY-MM>.jsonl.gz
func (r *Retention) archiveEvents(events []Event) error {
	if len(events) == 0 {
		return nil
	}

	if err := os.MkdirAll(r.archiveDir, 0o755); err != nil {
		return fmt.Errorf("아카이브 디렉토리 생성: %w", err)
	}

	// Group events by month
	byMonth := make(map[string][]Event)
	for _, evt := range events {
		key := evt.Timestamp.UTC().Format("2006-01")
		byMonth[key] = append(byMonth[key], evt)
	}

	for month, evts := range byMonth {
		archivePath := filepath.Join(r.archiveDir, month+".jsonl.gz")
		if err := appendToGzip(archivePath, evts); err != nil {
			return fmt.Errorf("월별 아카이브 %s 기록 실패: %w", month, err)
		}
	}
	return nil
}

// appendToGzip appends events to gzip-compressed JSONL file.
// Creates new file if it does not exist.
//
// @MX:WARN: [AUTO] Gzip files are not append-safe, so use read-and-rewrite method.
// @MX:REASON: [AUTO] Gzip format supports concatenated streams so append is actually possible,
// but use read-modify-write pattern for compatibility with standard readers.
func appendToGzip(archivePath string, events []Event) error {
	// gzip concatenation: append by adding new gzip stream to existing file.
	// Standard gzip reader can read concatenated streams sequentially.
	f, err := os.OpenFile(archivePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("아카이브 파일 열기: %w", err)
	}
	defer func() { _ = f.Close() }()

	gw := gzip.NewWriter(f)
	enc := json.NewEncoder(gw)
	for _, evt := range events {
		if err := enc.Encode(evt); err != nil {
			_ = gw.Close()
			return fmt.Errorf("gzip 인코딩: %w", err)
		}
	}
	if err := gw.Close(); err != nil {
		return fmt.Errorf("gzip 닫기: %w", err)
	}
	return nil
}

// appendLogTail copies, verbatim, every byte the log holds from offset from to its current end into
// dst, and adds one newline when the copied bytes are non-empty and do not end with one, so the
// replacement log stays empty or newline-terminated.
func appendLogTail(dst io.Writer, logPath string, from int64) error {
	src, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()
	if _, err := src.Seek(from, io.SeekStart); err != nil {
		return err
	}
	tail, err := io.ReadAll(src)
	if err != nil {
		return err
	}
	if len(tail) == 0 {
		_, err = dst.Write([]byte("\n"))
		return err
	}
	if tail[len(tail)-1] != '\n' {
		tail = append(tail, '\n')
	}
	_, err = dst.Write(tail)
	return err
}

// overwriteWithEvents overwrites the log file with the kept lines followed by the log tail from
// tailFrom on (see partitionEvents). The tail is read once, immediately before the rename.
func overwriteWithEvents(logPath string, lines []logLine, tailFrom int64) error {
	// Write to temporary file first, then atomic replacement
	dir := filepath.Dir(logPath)
	tmp, err := os.CreateTemp(dir, tmpPattern)
	if err != nil {
		return fmt.Errorf("임시 파일 생성: %w", err)
	}
	tmpPath := tmp.Name()

	enc := json.NewEncoder(tmp)
	for _, l := range lines {
		if l.raw != "" {
			if _, err := tmp.WriteString(l.raw + "\n"); err != nil {
				_ = tmp.Close()
				_ = os.Remove(tmpPath)
				return fmt.Errorf("임시 파일 쓰기: %w", err)
			}
			continue
		}
		if err := enc.Encode(l.evt); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("임시 파일 인코딩: %w", err)
		}
	}

	if err := appendLogTail(tmp, logPath, tailFrom); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("log tail carry: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("임시 파일 닫기: %w", err)
	}

	// Atomic replacement
	if err := os.Rename(tmpPath, logPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("파일 교체: %w", err)
	}
	return nil
}
