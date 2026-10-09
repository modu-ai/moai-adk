package bugreport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/atomicfile"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// spoolFileName is the capture spool's name under <moai home>/state/bugreport.
const spoolFileName = "spool.jsonl"

// BugreportStoreDir is the pipeline's user-scoped store directory relative
// to the moai home (design.md section 6, D36): every bugreport store — the
// capture spool, the drain's queue, the dedupe ledger, and the outbox log —
// lives under <moai home>/state/bugreport/, never under the project tree.
// The spool carries the verdict the drain acts on, and in this SPEC's threat
// model a project file is hostile input: a repository that ships a
// well-formed verdict-moai spool at a project path must not be able to drive
// publication from a consenting user's account.
const BugreportStoreDir = "state/bugreport"

// ErrSpoolFull: the spool is bounded at DefaultBugreportSpoolMaxLines lines
// and DefaultBugreportSpoolMaxBytes bytes; capture drops a signal beyond
// either (fail-open).
var ErrSpoolFull = errors.New("bugreport: capture spool is full")

// SpoolPath returns the user-scoped spool's path.
func SpoolPath() (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, filepath.FromSlash(BugreportStoreDir), spoolFileName), nil
}

// SpoolEntry is one spool line (design.md section 6): the kind, the verdict
// fixed at capture, the reason token, the moai-internal frames, and the
// optional closed-set detail — plus the CAPTURE-TIME build identity
// (review-gate finding: an error captured by v3.2.0 and flushed by a
// v3.2.1 binary must be reported as v3.2.0; the payload and the
// fingerprint both derive from these fields at drain time, never from the
// flushing binary's). No error text, no paths, no timestamp — the payload
// the pipeline later builds derives only from these closed fields.
type SpoolEntry struct {
	Kind    Kind     `json:"kind"`
	Verdict Verdict  `json:"verdict"`
	Reason  string   `json:"reason,omitempty"`
	Frames  []string `json:"frames"`
	Detail  string   `json:"detail,omitempty"`
	Version string   `json:"version,omitempty"`
	Commit  string   `json:"commit,omitempty"`
}

// marshalSpoolEntry renders the JSONL line with its trailing newline.
func marshalSpoolEntry(entry SpoolEntry) ([]byte, error) {
	line, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	return append(line, '\n'), nil
}

// ReadSpool reads every entry from the user-scoped spool. A missing file is
// an empty spool; a malformed LINE is skipped (untrusted local input — the
// same distrust the read-back rule applies to the queue's detail), never a
// drain failure.
func ReadSpool() ([]SpoolEntry, error) {
	entries, _, err := ReadSpoolConsumable()
	return entries, err
}

// ReadSpoolConsumable is ReadSpool plus the exact bytes the entries were
// parsed from: the batch a consumer may later hand to ConsumeSpoolPrefix.
// Capture keeps appending beyond those bytes while the consumer works.
//
// The read is BOUNDED (review gate finding, P2): a non-regular spool file is
// refused WITHOUT opening it — a FIFO swapped in at the spool path used to
// park a plain os.ReadFile indefinitely, ignoring the drain's deadline — and
// the open+read runs under DefaultBugreportSpoolReadTimeBox with the
// DefaultBugreportSpoolMaxBytes size cap (the spool is capped at capture; a
// larger file is out of contract). A missing file is still an empty spool.
// On deadline the helper goroutine is left parked on the blocked handle; it
// exits when the blocking writer closes, and the caller never waits for it.
func ReadSpoolConsumable() ([]SpoolEntry, []byte, error) {
	path, err := SpoolPath()
	if err != nil {
		return nil, nil, err
	}
	if info, serr := os.Stat(path); serr == nil && !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("bugreport: spool is not a regular file: %s", path)
	}
	type readResult struct {
		raw []byte
		err error
	}
	done := make(chan readResult, 1)
	go func() {
		// Bound the read AT the cap (card-review finding, P2): os.ReadFile
		// loaded the whole file before the size judgment — an 8MiB spool
		// allocated ~8.4MB on its way to being refused, and larger files
		// scale to OOM. The reader reads at most cap+1 bytes, ever.
		f, err := os.Open(path)
		if err != nil {
			done <- readResult{err: err}
			return
		}
		defer func() { _ = f.Close() }()
		raw, err := io.ReadAll(io.LimitReader(f, config.DefaultBugreportSpoolMaxBytes+1))
		done <- readResult{raw: raw, err: err}
	}()
	var raw []byte
	select {
	case r := <-done:
		if r.err != nil {
			if os.IsNotExist(r.err) {
				return nil, nil, nil
			}
			return nil, nil, r.err
		}
		raw = r.raw
	case <-time.After(config.DefaultBugreportSpoolReadTimeBox):
		return nil, nil, fmt.Errorf("bugreport: spool read exceeded its %s time box", config.DefaultBugreportSpoolReadTimeBox)
	}
	if len(raw) > config.DefaultBugreportSpoolMaxBytes {
		return nil, nil, fmt.Errorf("bugreport: spool is %d bytes, over the %d cap — purge required", len(raw), config.DefaultBugreportSpoolMaxBytes)
	}
	return parseSpoolLines(raw), raw, nil
}

// parseSpoolLines parses a spool snapshot, skipping malformed lines
// (untrusted local input).
func parseSpoolLines(raw []byte) []SpoolEntry {
	var out []SpoolEntry
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry SpoolEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if !entry.Kind.Valid() || !Verdict(entry.Verdict).Valid() {
			continue
		}
		out = append(out, entry)
	}
	return out
}

// ConsumeSpoolPrefix removes exactly the consumed prefix from the spool —
// the batch a drain read and processed — under the spool's cross-process
// section, the same claim the capture append takes. Inside the section the
// file is re-read and replaced only when it STILL begins with the consumed
// bytes: entries capture appended after the drain's read survive for the
// next drain (the review-gate lost-entry finding — a whole-file clear
// deleted every capture that landed mid-drain). A file that no longer
// starts with the batch — withdrawn, purged, or already consumed — is left
// untouched: reprocessing is safe because the ledger dedupes, and failing
// open here only delays a cleanup that the next drain retries.
func ConsumeSpoolPrefix(consumed []byte) error {
	path, err := SpoolPath()
	if err != nil {
		return err
	}
	release, err := claimSpoolSection(path)
	if err != nil {
		// Fail-open: the batch stays; the next drain retries the consume.
		return nil
	}
	defer func() { _ = release() }()

	now, err := boundedSpoolReread(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // nothing left to consume
		}
		return nil // unreadable (or past its time box): leave it for the next drain
	}
	if len(consumed) == 0 || !bytes.HasPrefix(now, consumed) {
		return nil // not the batch this consumer read: touch nothing
	}
	rest := now[len(consumed):]
	if len(rest) == 0 {
		// Nothing unconsumed: removing the file keeps the store directory
		// tidy; the next capture recreates it.
		if rmErr := os.Remove(path); rmErr != nil && !os.IsNotExist(rmErr) {
			return rmErr
		}
		return nil
	}
	// Replace with the unconsumed remainder, atomically, inside the section.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".spool-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(rest); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// PrefixLenForEntries returns the byte length of the raw spool prefix that
// covers the first n valid entries — any malformed or blank lines among
// them included, the span always ending on a line boundary. It is the
// exact prefix a partial drain may hand ConsumeSpoolPrefix: only the
// entries it actually processed leave the spool (review-gate finding: a
// retryable queue-write failure used to consume the whole batch, losing
// the failed item and everything after it). An n greater than the valid
// count returns the whole read.
func PrefixLenForEntries(raw []byte, n int) int {
	if n <= 0 {
		return 0
	}
	valid := 0
	offset := 0
	for offset < len(raw) {
		lineEnd := len(raw) // EOF without a trailing newline
		if end := bytes.IndexByte(raw[offset:], '\n'); end >= 0 {
			lineEnd = offset + end + 1 // include the newline
		}
		line := strings.TrimSpace(string(raw[offset:lineEnd]))
		if line != "" {
			var entry SpoolEntry
			if json.Unmarshal([]byte(line), &entry) == nil && entry.Kind.Valid() && Verdict(entry.Verdict).Valid() {
				valid++
				if valid == n {
					return lineEnd
				}
			}
		}
		offset = lineEnd
	}
	return len(raw)
}

// ClearSpool empties the spool (the drain consumed every line). Removing
// the file rather than truncating keeps the store directory tidy; the next
// capture recreates it.
func ClearSpool() error {
	path, err := SpoolPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// spoolGenerationFileName marks the store's generation: the purge bumps
// it, and an in-flight drain re-reads it before each item, so a batch read
// BEFORE the purge stops instead of resurrecting withdrawn reports
// (review gate finding, P2).
const spoolGenerationFileName = "generation"

// SpoolGenerationPath is the generation marker's path beside the spool.
func SpoolGenerationPath() (string, error) {
	home, err := paths.MoaiHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, filepath.FromSlash(BugreportStoreDir), spoolGenerationFileName), nil
}

// SpoolGeneration reads the store's generation counter. An absent marker
// is generation 0. The read is bounded the way every file in this store
// is: a non-regular marker is refused without opening, and the read costs
// one small capped allocation.
func SpoolGeneration() (uint64, error) {
	path, err := SpoolGenerationPath()
	if err != nil {
		return 0, err
	}
	if info, serr := os.Stat(path); serr == nil && !info.Mode().IsRegular() {
		return 0, fmt.Errorf("bugreport: generation marker is not a regular file: %s", path)
	}
	type readResult struct {
		raw []byte
		err error
	}
	done := make(chan readResult, 1)
	go func() {
		f, oerr := os.Open(path)
		if oerr != nil {
			done <- readResult{err: oerr}
			return
		}
		defer func() { _ = f.Close() }()
		raw, rerr := io.ReadAll(io.LimitReader(f, 64))
		done <- readResult{raw: raw, err: rerr}
	}()
	var raw []byte
	select {
	case r := <-done:
		if r.err != nil {
			if os.IsNotExist(r.err) {
				return 0, nil
			}
			return 0, r.err
		}
		raw = r.raw
	case <-time.After(config.DefaultBugreportSpoolReadTimeBox):
		return 0, fmt.Errorf("bugreport: generation read exceeded its %s time box", config.DefaultBugreportSpoolReadTimeBox)
	}
	gen, perr := strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
	if perr != nil {
		return 0, fmt.Errorf("bugreport: generation marker is unreadable: %w", perr)
	}
	return gen, nil
}

// BumpSpoolGeneration invalidates every batch an in-flight drain read
// before this bump. The purge calls it before removing the stores; a lost
// race between two concurrent bumps only skips a number.
func BumpSpoolGeneration() error {
	return BumpSpoolGenerationContext(context.Background())
}

// BumpSpoolGenerationContext is BumpSpoolGeneration with the caller's
// cancellation: the claim retry loop honors ctx (a withdrawal with a
// deadline must not wait out the full bump claim deadline behind a held
// lock — gate finding, the DrainContext withdrawal path).
func BumpSpoolGenerationContext(ctx context.Context) error {
	path, err := SpoolGenerationPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	// The read-increment-write is SERIALIZED under the marker's own section
	// lock (review gate finding, P1): two concurrent purges read the SAME
	// counter and wrote the SAME next number, so the second purge's
	// withdrawal was invisible to a sender that had already read the first
	// purge's generation — and it published. A lock-acquire failure
	// propagates: an unbumped generation is a withdrawal that did not
	// happen, and every in-flight reader treating the store as live is the
	// failure this bump exists to prevent.
	//
	// Contention beyond one ClaimSection retry budget (CI, Race Test 2:
	// 32 concurrent purges under -race exhausted the 8×5ms budget — "lock
	// held" skips a bump) retries on the lock-held shape until the bump
	// deadline — or until the CALLER's ctx cancels, whichever comes first
	// (gate finding: the retry loop originally used context.Background(),
	// so a DrainContext deadline behind a held lock waited the full bump
	// deadline out).
	deadline := time.Now().Add(spoolBumpClaimDeadline)
	var release func() error
	for {
		release, err = atomicfile.ClaimSection(ctx, path+".lock", 0o600, spoolSectionRetries, spoolSectionDelay)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("bugreport: bump claim canceled: %w", ctx.Err())
		case <-time.After(spoolSectionDelay):
		}
		if !time.Now().Before(deadline) {
			return err
		}
	}
	defer func() { _ = release() }()
	gen, err := SpoolGeneration()
	if err != nil {
		gen = 0 // an unreadable marker still bumps past itself
	}
	// Atomic replacement (tmp + rename): a concurrent reader never sees a
	// half-written number — a partial read already degrades safely (the
	// drain passes through, the sender stops), but the store never writes
	// one in the first place.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".generation-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, werr := tmp.Write([]byte(strconv.FormatUint(gen+1, 10))); werr != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return werr
	}
	if cerr := tmp.Close(); cerr != nil {
		_ = os.Remove(tmpName)
		return cerr
	}
	if cerr := os.Chmod(tmpName, 0o600); cerr != nil {
		_ = os.Remove(tmpName)
		return cerr
	}
	return os.Rename(tmpName, path)
}

// boundedSpoolReread re-reads the spool under the same bounds as the first
// read (review gate finding, P2): a non-regular file is refused without
// opening — a FIFO swapped in between the two reads parked the consume
// WITH the spool section lock held — and the read is capped and time-boxed.
// On deadline the helper goroutine is left parked on the blocked handle;
// it exits when the blocking writer closes, and the caller never waits.
func boundedSpoolReread(path string) ([]byte, error) {
	if info, serr := os.Stat(path); serr == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("bugreport: spool is not a regular file: %s", path)
	}
	type readResult struct {
		raw []byte
		err error
	}
	done := make(chan readResult, 1)
	go func() {
		f, oerr := os.Open(path)
		if oerr != nil {
			done <- readResult{err: oerr}
			return
		}
		defer func() { _ = f.Close() }()
		raw, rerr := io.ReadAll(io.LimitReader(f, config.DefaultBugreportSpoolMaxBytes+1))
		done <- readResult{raw: raw, err: rerr}
	}()
	select {
	case r := <-done:
		return r.raw, r.err
	case <-time.After(config.DefaultBugreportSpoolReadTimeBox):
		return nil, fmt.Errorf("bugreport: spool reread exceeded its %s time box", config.DefaultBugreportSpoolReadTimeBox)
	}
}

// appendSpoolLine appends one JSONL line to the user-scoped spool, bounded
// by the line and byte ceilings. The directory is created 0700 and the file
// 0600, so the store never depends on init having run and is never
// world-readable.
//
// The ceiling check and the append share ONE cross-process critical section
// — atomicfile.Claim on a sibling lock, the D37 discipline — because as two
// unsynchronized steps, 32 concurrent writers each saw room and the spool
// finished over its 200-line ceiling (review-gate finding #4). A writer
// that cannot take the section within its short budget drops the signal
// (fail-open): the ceiling is the point.
func appendSpoolLine(entry SpoolEntry) error {
	path, err := SpoolPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	line, err := marshalSpoolEntry(entry)
	if err != nil {
		return err
	}

	release, err := claimSpoolSection(path)
	if err != nil {
		return err
	}
	defer func() { _ = release() }()

	// Inside the section: bounds are checked against the file exactly as it
	// will be appended to. A 64 KiB read at capture is well inside the
	// 50 ms box.
	if info, statErr := os.Stat(path); statErr == nil {
		if info.Size()+int64(len(line)) > int64(config.DefaultBugreportSpoolMaxBytes) {
			return ErrSpoolFull
		}
		raw, readErr := os.ReadFile(path)
		if readErr == nil && bytes.Count(raw, []byte{'\n'})+1 > config.DefaultBugreportSpoolMaxLines {
			return ErrSpoolFull
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(line); err != nil {
		return err
	}
	return nil
}

// spoolSectionClaim is the O_EXCL lock artifact's name next to the spool.
const spoolSectionClaim = "spool.lock"

// spoolSectionRetries / spoolSectionDelay bound a writer's wait for the
// section: short, because dropping the signal is always the acceptable
// outcome on the capture path.
const (
	spoolSectionRetries = 8
	spoolSectionDelay   = 5 * time.Millisecond
)

// spoolBumpClaimDeadline bounds how long a generation bump keeps retrying
// the section claim under contention before propagating the lock-held
// error. Sized for the observed worst contender fan-out (the CI race test's
// 32 concurrent purges under -race scheduling) with an order of magnitude
// of headroom — the bump's P1 contract is that every purge's withdrawal
// lands, serialized, so contention waits instead of failing.
const spoolBumpClaimDeadline = 10 * time.Second

// claimSpoolSection takes the spool's critical section, returning its
// release func. The lock is the SAME owner-verified claim the queue uses
// (atomicfile.ClaimSection, the D37 machinery): labelled with the owner's
// pid and boot identity on acquire, and its contention path breaks the lock
// ONLY on a verified-dead owner. Before this wiring the lock was a bare
// exclusive create with no owner record and no reclaim — a process dying
// while it held the section left the lock forever, every later capture
// write failed its claim budget, and purge did not remove the artifact, so
// the spool went silently dead for the life of the boot.
func claimSpoolSection(spoolPath string) (func() error, error) {
	lockPath := filepath.Join(filepath.Dir(spoolPath), spoolSectionClaim)
	return atomicfile.ClaimSection(context.Background(), lockPath, 0o600, spoolSectionRetries, spoolSectionDelay)
}
