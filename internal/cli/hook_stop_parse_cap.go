package cli

// hook_stop_parse_cap.go — SPEC-HOOK-STOP-PARSE-CAP-001. moai's own cap on
// consecutive stdin-parse-failure Stops under the Claude harness. A Stop whose
// stdin cannot be parsed is denied fail-closed, and the host's Stop block cap
// does not reliably end that loop: a session launched with the raised cap of
// 200 keeps it going, and the default cap appears to count only blocks with no
// tool use in between. So moai counts the consecutive parse-failure Stops per
// counting key and, above N, answers with no opinion instead of a deny.
//
// Counting key (REQ-SPC-014): CLAUDE_CODE_SESSION_ID when set and non-blank,
// else the session owner process the canonical resolver finds. The count lives
// in one small record per key under the project's state area; a parsed Stop
// deletes it (REQ-SPC-005) and a record older than the expiry counts as absent
// (REQ-SPC-007). Every failure to establish a key or to use the record keeps
// the deny (REQ-SPC-008) — the cap only ever releases on a count it could
// actually read and write.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/codexadapter"
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/hook"
	"github.com/modu-ai/moai-adk/internal/session"
)

// stopParseCapKeyKind names which rule produced a counting key. Two keys of
// different kinds never share a record, even when their values are equal.
type stopParseCapKeyKind string

const (
	stopParseCapKeySession stopParseCapKeyKind = "session"
	stopParseCapKeyProcess stopParseCapKeyKind = "process"

	// stopParseCapStateRel is the state area, relative to the project root.
	stopParseCapStateRel = ".moai/state/stop-parse-cap"

	// stopParseCapReleasedDiscardKey marks a released-cap record in the
	// adapter's sink, distinct from the three stdin/fault keys.
	stopParseCapReleasedDiscardKey = "stdin-parse-stop-cap-released"
)

// stopParseCapRecordName is the only name shape a record may have; the sweep
// touches nothing else (REQ-SPC-007, REQ-SPC-015).
var stopParseCapRecordName = regexp.MustCompile(`^[0-9a-f]{64}\.json$`)

// Seams. Tests swap the process view to walk a synthetic ancestry with the
// real resolver, and the remover to observe a failed delete.
var (
	stopParseCapProcessView = session.LiveProcessView
	stopParseCapRemove      = os.Remove
	stopParseCapNow         = time.Now
)

type stopParseCapKey struct {
	kind  stopParseCapKeyKind
	value string
}

// stopParseCapRecord is one key's count. It carries the key kind but never the
// key value, so a session id is not stored anywhere (REQ-SPC-015).
type stopParseCapRecord struct {
	Count     int                 `json:"count"`
	UpdatedAt time.Time           `json:"updated_at"`
	KeyKind   stopParseCapKeyKind `json:"key_kind"`
}

// stopParseCapFileName derives a record's file name from its key: the
// lowercase hex sha256 of "<kind>:<value>" plus ".json". Whatever the value
// holds — separators, "..", NUL, control characters — the name has a fixed
// length and alphabet, so it cannot leave the state area (REQ-SPC-015).
func stopParseCapFileName(kind stopParseCapKeyKind, value string) string {
	sum := sha256.Sum256([]byte(string(kind) + ":" + value))
	return hex.EncodeToString(sum[:]) + ".json"
}

// resolveStopParseCapKey picks the counting key (REQ-SPC-014): the session id
// when set and non-blank, else the session owner process. It never falls back
// to the raw parent pid — a wrapper shell between the host and moai would make
// that differ on every call.
func resolveStopParseCapKey() (stopParseCapKey, bool) {
	if id := strings.TrimSpace(os.Getenv(config.EnvClaudeCodeSessionID)); id != "" {
		return stopParseCapKey{kind: stopParseCapKeySession, value: id}, true
	}
	if pid, ok := stopParseCapProcessView().ResolveOwnerPID(os.Getenv(config.EnvMoaiSessionPID)); ok {
		return stopParseCapKey{kind: stopParseCapKeyProcess, value: strconv.Itoa(pid)}, true
	}
	return stopParseCapKey{}, false
}

// stopParseCapExpired reports whether a record counts as absent. A record
// dated implausibly far in the future is treated the same way, so a forged
// timestamp cannot keep a count alive forever.
func stopParseCapExpired(r stopParseCapRecord, now time.Time) bool {
	age := now.Sub(r.UpdatedAt)
	return age > config.DefaultStopParseCapExpiry || age < -config.DefaultStopParseCapExpiry
}

// stopParseCapStateDir checks the state area without following links. It
// reports exists=false when there is nothing there and create is false.
func stopParseCapStateDir(root string, create bool) (dir string, exists bool, err error) {
	if root == "" {
		return "", false, errors.New("project root could not be resolved")
	}
	dir = filepath.Join(root, stopParseCapStateRel)
	fi, err := os.Lstat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		if !create {
			return dir, false, nil
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return dir, false, fmt.Errorf("create state area: %w", err)
		}
		fi, err = os.Lstat(dir)
	}
	if err != nil {
		return dir, false, fmt.Errorf("inspect state area: %w", err)
	}
	if fi.Mode()&fs.ModeSymlink != 0 {
		return dir, false, errors.New("state area is a symbolic link")
	}
	if !fi.IsDir() {
		return dir, false, errors.New("state area is not a directory")
	}
	return dir, true, nil
}

// readStopParseCapRecord reads a regular-file record. ok is false when the
// content cannot be parsed.
func readStopParseCapRecord(path string) (stopParseCapRecord, bool) {
	var r stopParseCapRecord
	b, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(b, &r) != nil || r.Count < 0 {
		return stopParseCapRecord{}, false
	}
	return r, true
}

// incrementStopParseCap adds one to key's count and returns the new count
// (REQ-SPC-001). A missing, unparseable, or expired record starts from zero;
// a record slot that is not a regular file is an error (REQ-SPC-008).
func incrementStopParseCap(root string, key stopParseCapKey, now time.Time) (int, error) {
	dir, _, err := stopParseCapStateDir(root, true)
	if err != nil {
		return 0, err
	}
	name := stopParseCapFileName(key.kind, key.value)
	path := filepath.Join(dir, name)

	prev := 0
	fi, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return 0, fmt.Errorf("inspect count record: %w", err)
	case !fi.Mode().IsRegular():
		return 0, errors.New("count record is not a regular file")
	default:
		if r, ok := readStopParseCapRecord(path); ok && !stopParseCapExpired(r, now) {
			prev = r.Count
		}
	}

	if err := sweepStopParseCap(dir, name, now); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "moai hook Stop: stop-parse cap sweep: %v\n", err)
	}

	count := prev + 1
	if err := writeStopParseCapRecord(dir, path, stopParseCapRecord{Count: count, UpdatedAt: now, KeyKind: key.kind}); err != nil {
		return 0, err
	}
	return count, nil
}

// writeStopParseCapRecord writes through a temporary file and a rename, so a
// reader never sees a half-written record (REQ-SPC-008).
func writeStopParseCapRecord(dir, path string, r stopParseCapRecord) error {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("encode count record: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("write count record: %w", err)
	}
	_, werr := tmp.Write(b)
	cerr := tmp.Close()
	if werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Rename(tmp.Name(), path)
	}
	if werr != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("write count record: %w", werr)
	}
	return nil
}

// sweepStopParseCap removes other keys' expired records. Only regular files
// whose name matches the record shape and whose content parses as an expired
// record are removed; links are never followed and anything else is left
// alone (REQ-SPC-007). It returns the first failure.
func sweepStopParseCap(dir, own string, now time.Time) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var first error
	for _, e := range entries {
		name := e.Name()
		if name == own || !stopParseCapRecordName.MatchString(name) {
			continue
		}
		p := filepath.Join(dir, name)
		fi, err := os.Lstat(p)
		if err != nil || !fi.Mode().IsRegular() {
			continue
		}
		r, ok := readStopParseCapRecord(p)
		if !ok || !stopParseCapExpired(r, now) {
			continue
		}
		if err := stopParseCapRemove(p); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// @MX:WARN: [AUTO] releasing a Stop here lets a parse-failure Stop through with no opinion — every Stop guard handler is skipped until a parsed Stop or the expiry resets the count
// @MX:REASON: [AUTO] REQ-SPC-003 — a parse failure hides stop_hook_active, so the turn boundary is unreadable and the release persists; the release fires only on a count that was read and written successfully (REQ-SPC-008)
// applyStopParseCap counts one Claude parse-failure Stop and reports whether
// the cap is exceeded, in which case it has already written the stderr line
// and the record, and the caller answers with no opinion. Any failure to
// count keeps the deny and says so on stderr. The cap covers Stop alone — the
// other decision events block a tool call or a prompt, not the end of a turn,
// so their deny cannot loop (spec.md §D) — and every other event returns
// false without touching the count (REQ-SPC-006).
func applyStopParseCap(label string, event hook.EventType, stdinBytes int, parseErr error) bool {
	if event != hook.EventStop {
		return false
	}
	key, ok := resolveStopParseCapKey()
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "moai hook %s: stop-parse cap not applied: no counting key (no session id and no session owner process); keeping the fail-closed deny\n", label)
		return false
	}
	count, err := incrementStopParseCap(resolveHookProjectRoot(), key, stopParseCapNow())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "moai hook %s: stop-parse cap not applied: %v; keeping the fail-closed deny\n", label, err)
		return false
	}
	if count <= config.DefaultStopParseCapLimit {
		return false
	}

	_, _ = fmt.Fprintf(os.Stderr, "moai hook %s: invalid stdin JSON (%v) on %s, harness %s; stop-parse cap released: consecutive parse failures %d > N=%d (key: %s); answered with no opinion\n",
		label, parseErr, event, codexadapter.HarnessClaude, count, config.DefaultStopParseCapLimit, key.kind)
	recordStdinParseFailure(codexadapter.Discard{
		Event:         event,
		Key:           stopParseCapReleasedDiscardKey,
		ContentLength: stdinBytes,
		Reason:        "stdin parse failure on Stop answered with no opinion: moai's consecutive parse-failure limit was exceeded",
	})
	return true
}

// resetStopParseCap deletes the counting record of this process's key after a
// Claude Stop parsed (REQ-SPC-005). It creates nothing, and a failed delete is
// one stderr line that does not change the hook's answer.
func resetStopParseCap() {
	key, ok := resolveStopParseCapKey()
	if !ok {
		return
	}
	if err := deleteStopParseCap(resolveHookProjectRoot(), key); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "moai hook Stop: stop-parse cap reset: %v\n", err)
	}
}

func deleteStopParseCap(root string, key stopParseCapKey) error {
	dir, exists, err := stopParseCapStateDir(root, false)
	if err != nil || !exists {
		// A missing, linked, or non-directory state area holds no record
		// this process may delete.
		return nil
	}
	path := filepath.Join(dir, stopParseCapFileName(key.kind, key.value))
	fi, err := os.Lstat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("inspect count record: %w", err)
	case !fi.Mode().IsRegular():
		return errors.New("count record is not a regular file; left in place")
	}
	if err := stopParseCapRemove(path); err != nil {
		return fmt.Errorf("delete count record: %w", err)
	}
	return nil
}
