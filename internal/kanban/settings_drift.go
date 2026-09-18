// settings_drift.go — the pre-merge `.claude/settings.json` drift assertion
// (card t488, SPEC-PREMERGE-SETTINGS-DRIFT-001).
//
// Two card worktrees have been found with a modified working copy of the
// TRACKED `.claude/settings.json` and no author anyone could name. Neither was
// caught by the merge window: the first surfaced because a merge happened to
// be refused, the second in a 79-tree sweep nine days later. Authorship was
// closed as unattributable; this file addresses the other end — a detection
// point that runs whether or not anyone is looking.
//
// Three properties are load-bearing and each has a reason that is not obvious:
//
//  1. `--no-optional-locks` is mandatory. A plain `git status` takes an index
//     WRITE lock for tens of milliseconds (measured on card t485 against
//     `moai statusline`), so a check that runs immediately before a merge on a
//     machine with several lanes would manufacture the contention it exists to
//     protect. Removing the flag changes neither this code's output nor its
//     return value — only that side effect — so it is pinned by asserting the
//     argv the executor actually received, not by behaviour.
//
//  2. The verdict is the MATCH COUNT, never a process exit code. Card t474
//     watched a gate invert while its grep still exited 0.
//
//  3. Nothing here ever restores, reverts or deletes the file it finds. The
//     runtime writes that file and it can hold machine-specific values —
//     tokens, absolute paths, tmux pane ids — so an automatic restore is
//     itself data destruction. Disposal is a human decision.
package kanban

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SettingsDriftWatchedPath is the ONE tracked path this gate watches, relative
// to the target tree's top level. Widening the watched set is out of scope by
// operator decision: a wider gate was not chosen, and choosing it silently
// here would be the choice being made by whoever edits this line.
const SettingsDriftWatchedPath = ".claude/settings.json"

// SettingsDriftDirName and SettingsDriftLedgerName name the preservation
// directory and its ledger under the PRIMARY checkout's state directory. The
// state directory is gitignored, so preserved copies stay untracked — which is
// the intent: promoting one into history requires a human secret-scan first.
const (
	SettingsDriftDirName    = "settings-drift"
	SettingsDriftLedgerName = "ledger.jsonl"
)

// SettingsDriftStatus is the verdict surface. Three named states, not a
// boolean: a boolean cannot express "not measured", so a failed measurement
// would collapse into `false` or into an omitted field, and either reads as a
// pass to a consumer. That collapse is the nine-day blindness this card exists
// to close, so the third state is named rather than inferred.
type SettingsDriftStatus string

const (
	// SettingsDriftClean — the predicate ran and found no drift.
	SettingsDriftClean SettingsDriftStatus = "clean"
	// SettingsDriftDetected — the predicate ran and found drift.
	SettingsDriftDetected SettingsDriftStatus = "drift"
	// SettingsDriftUndetermined — the predicate could not run. Never folded
	// into clean, and the match count is not reported under it.
	SettingsDriftUndetermined SettingsDriftStatus = "undetermined"
)

// SettingsDriftMatchCountUnmeasured is the match count returned whenever no
// measurement was taken. It is negative on purpose: a caller that forgets to
// check the error gets an impossible value rather than a number that reads as
// "clean".
const SettingsDriftMatchCountUnmeasured = -1

// SettingsDriftRunner is the execution boundary. EVERY command this gate runs
// against the target tree or its repository goes through it, and it records
// the argv it was handed.
//
// The record is what the argv assertion reads. A separately exported argv
// builder would not do: if the execution site spelled out its own
// `exec.Command("git", ...)`, the builder and the executed command could drift
// apart and removing the flag from either would leave the test green — the
// vacuous-match shape card t477 measured.
//
// The recording is also what mechanizes "the gate never commits or pushes" and
// "the gate never modifies the target tree". Those absence assertions are only
// worth anything because the list is complete: a list holding some of the
// execution asserts nothing about the rest.
type SettingsDriftRunner interface {
	// Run executes argv with dir as the working directory and returns stdout.
	Run(dir string, argv []string) (string, error)
	// Commands returns the argv of every call, joined by single spaces, in
	// call order.
	Commands() []string
}

// ExecRunner is the production SettingsDriftRunner: os/exec, plus the record.
type ExecRunner struct {
	mu       sync.Mutex
	commands []string
}

// NewExecRunner returns a runner with an empty record.
func NewExecRunner() *ExecRunner { return &ExecRunner{} }

// Run records argv, then executes it. The record is taken BEFORE the call, so
// a command that fails is still recorded — an assertion about what the gate
// attempted must not depend on whether the attempt succeeded.
func (r *ExecRunner) Run(dir string, argv []string) (string, error) {
	if len(argv) == 0 {
		return "", errors.New("settings-drift: empty command")
	}
	r.mu.Lock()
	r.commands = append(r.commands, strings.Join(argv, " "))
	r.mu.Unlock()

	cmd := exec.Command(argv[0], argv[1:]...) // #nosec G204 -- argv is built by this package from fixed literals
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return string(out), fmt.Errorf("settings-drift: %s: %w: %s",
			strings.Join(argv, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

// Commands returns a copy of the recorded argv strings.
func (r *ExecRunner) Commands() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.commands))
	copy(out, r.commands)
	return out
}

// settingsDriftPredicateArgv is the ONE place the predicate's argv is built.
// It is unexported deliberately: the assertion that pins these tokens reads
// the executor's record, not this function's return value.
func settingsDriftPredicateArgv() []string {
	return []string{"git", "--no-optional-locks", "status", "--porcelain", "--", SettingsDriftWatchedPath}
}

// settingsDriftTopLevelArgv resolves the target tree's top level. It goes
// through the same runner as the predicate, so the recorded list holds the
// whole of the gate's execution rather than a chosen part of it.
func settingsDriftTopLevelArgv() []string {
	return []string{"git", "rev-parse", "--show-toplevel"}
}

// DetectSettingsDrift runs the predicate against worktree and reports how many
// lines it matched.
//
// The count is the verdict: 0 is a pass, 1 or more is drift. On failure the
// count is SettingsDriftMatchCountUnmeasured and the error is returned — the
// absence of a signal is not evidence of cleanliness, so there is no path here
// that answers 0 without having measured 0.
//
// An untracked watched file counts: `git status --porcelain` prints a `?? `
// line for it, and a card worktree where this file is untracked is an
// anomaly worth reporting rather than passing quietly.
func DetectSettingsDrift(worktree string, runner SettingsDriftRunner) (int, string, error) {
	out, err := runner.Run(worktree, settingsDriftPredicateArgv())
	if err != nil {
		return SettingsDriftMatchCountUnmeasured, "", err
	}
	raw := strings.TrimRight(out, "\n")
	if strings.TrimSpace(raw) == "" {
		return 0, "", nil
	}
	count := 0
	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count, raw, nil
}

// SettingsDriftParams carries what the gate needs. Dir is the caller's working
// directory (the tree being measured); Root is the PRIMARY checkout, whose
// state directory receives the preserved copy — a preserved copy inside the
// card's own worktree is invisible to the lead, who is not in that tree.
type SettingsDriftParams struct {
	Dir      string
	Root     string
	Card     string
	Branch   string
	Bypassed bool
	Runner   SettingsDriftRunner
}

// SettingsDriftResult is the gate's report. Err and PreserveErr are separate
// fields because they are separate failures: a preservation failure must not
// be able to change the verdict, which is decided by the match count alone.
type SettingsDriftResult struct {
	Status        SettingsDriftStatus
	MatchCount    int
	Worktree      string
	Path          string
	SHA256        string
	SizeBytes     int64
	PreservedPath string
	Bypassed      bool
	Err           error
	PreserveErr   error
}

// SettingsDriftDir returns the preservation directory under root.
func SettingsDriftDir(root string) string {
	return filepath.Join(root, ".moai", "state", SettingsDriftDirName)
}

// SettingsDriftRowStatus names what a ledger row RECORDS. Card t765: the
// ledger held detections and nothing else, so every hit read as still open
// however long ago it was dealt with — the reader could not tell a live
// anomaly from one a lane cleaned up minutes later.
type SettingsDriftRowStatus string

const (
	// SettingsDriftRowDetected — drift was found and preserved.
	SettingsDriftRowDetected SettingsDriftRowStatus = "detected"
	// SettingsDriftRowResolved — a later measurement of the SAME worktree
	// found it clean, closing the detection the row names.
	SettingsDriftRowResolved SettingsDriftRowStatus = "resolved"
)

// settingsDriftLedgerRow is one ledger line. Contents of the drifted file are
// never recorded — only its path, hash and size — because it can hold secrets.
//
// `sha256` and `size_bytes` are omitempty because a resolution row measures
// nothing about the file: it neither reads nor hashes it, and a zero-valued
// digest printed there would read as a digest of the restored content. A
// detection row always carries both (it cannot reach the append without having
// hashed the bytes), so the tag changes no detection output.
//
// Rows written before t765 carry no `status` at all. That absence is the
// legacy encoding of `detected`, and it is read as such — see
// lastSettingsDriftRowForWorktree.
type settingsDriftLedgerRow struct {
	MeasuredAt    string                 `json:"measured_at"`
	Card          string                 `json:"card"`
	Branch        string                 `json:"branch"`
	Worktree      string                 `json:"worktree"`
	SourcePath    string                 `json:"source_path"`
	PreservedPath string                 `json:"preserved_path,omitempty"`
	SHA256        string                 `json:"sha256,omitempty"`
	SizeBytes     int64                  `json:"size_bytes,omitempty"`
	MatchCount    int                    `json:"match_count"`
	Bypassed      bool                   `json:"bypassed"`
	PreserveError string                 `json:"preserve_error,omitempty"`
	Status        SettingsDriftRowStatus `json:"status,omitempty"`

	// ResolvesSHA256 and ResolvesMeasuredAt are the join. They are present on
	// a resolution row only, and they name the detection it closes by the two
	// values that identify one in this file. Without them a reader with two
	// detections and one resolution cannot say which was closed.
	ResolvesSHA256     string `json:"resolves_sha256,omitempty"`
	ResolvesMeasuredAt string `json:"resolves_measured_at,omitempty"`
}

// AssessSettingsDrift resolves the target tree, runs the predicate, and on a
// hit preserves the working copy and appends a ledger row.
//
// It writes nothing to the target tree, ever — the only writes leave for the
// primary checkout's state directory.
//
// The verdict below is `matchCount > 0`. That comparison is the gate's
// decision point: inverting it makes a drifted tree read clean, which no
// predicate-level fixture can see, since the predicate's own count is
// unchanged either way. What sees it is the refusal path (AC-PSD-009).
func AssessSettingsDrift(p SettingsDriftParams) SettingsDriftResult {
	runner := p.Runner
	if runner == nil {
		runner = NewExecRunner()
	}
	result := SettingsDriftResult{
		Status:     SettingsDriftUndetermined,
		MatchCount: SettingsDriftMatchCountUnmeasured,
		Worktree:   p.Dir,
	}

	top, err := runner.Run(p.Dir, settingsDriftTopLevelArgv())
	if err != nil {
		result.Err = err
		return result
	}
	worktree := strings.TrimSpace(top)
	if worktree == "" {
		result.Err = fmt.Errorf("settings-drift: could not resolve the top level of %s", p.Dir)
		return result
	}
	result.Worktree = worktree
	result.Path = filepath.Join(worktree, filepath.FromSlash(SettingsDriftWatchedPath))

	count, _, err := DetectSettingsDrift(worktree, runner)
	if err != nil {
		result.Err = err
		return result
	}
	result.MatchCount = count
	if count > 0 {
		result.Status = SettingsDriftDetected
	} else {
		result.Status = SettingsDriftClean
		// A clean measurement is the only evidence a detection is over, and
		// until t765 it was thrown away. Recording it is an APPEND like every
		// other row: the detection it closes is left exactly as written, and
		// nothing about the file on disk is touched, read or restored.
		if err := recordSettingsDriftResolution(p, worktree); err != nil {
			result.PreserveErr = err
		}
		return result
	}

	result.Bypassed = p.Bypassed
	// ONE read. The hash, the size and the preserved bytes all come from this
	// single buffer, and nothing re-reads the file afterwards.
	//
	// This is not defensive tidiness. The premise of this whole gate is that
	// something writes that file unpredictably — it is why REQ-PSD-006 forbids
	// an automatic restore. Hashing one read and copying a second would let
	// the ledger's digest describe bytes the preserved copy does not contain,
	// and the preserved copy's fingerprint is exactly the evidence a later
	// reader compares a third instance against. A silently wrong fingerprint
	// is worse than no fingerprint, because it is trusted.
	data, sum, size, readErr := readAndHashSettingsDriftSource(result.Path)
	if readErr != nil {
		result.PreserveErr = readErr
		return result
	}
	result.SHA256 = sum
	result.SizeBytes = size

	// The interleaving point, between the measurement and the write. Nil in
	// production; see the hook's declaration.
	if settingsDriftPreserveTestHook != nil {
		settingsDriftPreserveTestHook()
	}

	preserved, preserveErr := preserveSettingsDriftCopy(p.Root, p.Card, p.Branch, data, sum)
	if preserveErr != nil {
		result.PreserveErr = preserveErr
	} else {
		result.PreservedPath = preserved
	}

	if ledgerErr := appendSettingsDriftLedger(p.Root, settingsDriftLedgerRow{
		MeasuredAt:    time.Now().UTC().Format(time.RFC3339Nano),
		Card:          p.Card,
		Branch:        p.Branch,
		Worktree:      worktree,
		SourcePath:    result.Path,
		PreservedPath: result.PreservedPath,
		SHA256:        sum,
		SizeBytes:     size,
		MatchCount:    count,
		Bypassed:      p.Bypassed,
		PreserveError: errText(preserveErr),
		Status:        SettingsDriftRowDetected,
	}); ledgerErr != nil && result.PreserveErr == nil {
		result.PreserveErr = ledgerErr
	}

	return result
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// readAndHashSettingsDriftSource reads the drifted file ONCE and returns the
// bytes alongside their digest and length. Every downstream value — the ledger
// digest, the reported size, the preserved copy — is derived from the returned
// buffer, so no second read can make them disagree.
func readAndHashSettingsDriftSource(path string) ([]byte, string, int64, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is <worktree>/.claude/settings.json
	if err != nil {
		return nil, "", 0, fmt.Errorf("settings-drift: read %s: %w", path, err)
	}
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), int64(len(data)), nil
}

// settingsDriftPreserveTestHook is a nil-by-default, TEST-ONLY interleaving
// point invoked once between reading the drifted file and writing the
// preserved copy. It exists so the single-read invariant can be CONSTRUCTED
// rather than waited for: the window between a hash and a copy is
// microseconds, and a criterion that waits for a concurrent writer to land in
// it has no stop rule.
//
// This is the same seam, for the same reason, as
// integrationLockMutationTestHook next door. It is unexported and
// package-level, so only `package kanban` can assign it, and no non-test file
// does. Every production path leaves it nil, and the call site is nil-guarded
// — invoking a nil func() panics in Go, so the guard is what makes "with the
// hook nil, behaviour is byte-for-byte unchanged" true rather than intended.
var settingsDriftPreserveTestHook func()

// settingsDriftLabel names the preserved copy's owner: the card id when the
// caller supplied one, else the branch, else `unknown`. `unknown` is not a
// dead end — the ledger row carries the absolute worktree path, so the trail
// survives an unlabelled preserve.
func settingsDriftLabel(card, branch string) string {
	if c := sanitizeSettingsDriftLabel(card); c != "" {
		return c
	}
	if b := sanitizeSettingsDriftLabel(branch); b != "" {
		return b
	}
	return "unknown"
}

// sanitizeSettingsDriftLabel keeps the label to characters that are safe in a
// file name on every supported platform. A branch name carries slashes
// (`release/v3.2.0`), which would otherwise turn the preserve into a write to
// a directory that does not exist.
func sanitizeSettingsDriftLabel(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// preserveSettingsDriftCopy copies the drifted working copy into the primary
// checkout's state directory and returns the path it landed on.
//
// The name is `settings.json.<label>.<UTC timestamp to the millisecond>.<first
// 8 of the sha256>`, with a `.2`, `.3` … sequence suffix on collision. The
// suffix is not belt-and-braces: two hits with the same card and the same
// content inside one millisecond produce an identical name from the first
// three components, and only the suffix makes "an earlier copy is never
// overwritten" decidable rather than probable. O_EXCL is what makes the loop
// safe against a concurrent second writer rather than merely sequential.
// It takes the BYTES rather than the source path deliberately: a path would
// license a second read, and the digest recorded beside the copy would then be
// a digest of different bytes than the copy holds.
func preserveSettingsDriftCopy(root, card, branch string, data []byte, sum string) (string, error) {
	dir := SettingsDriftDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("settings-drift: create %s: %w", dir, err)
	}

	short := sum
	if len(short) > 8 {
		short = short[:8]
	}
	base := fmt.Sprintf("settings.json.%s.%s.%s",
		settingsDriftLabel(card, branch),
		time.Now().UTC().Format("20060102T150405.000Z"),
		short)

	for n := 1; n <= 1000; n++ {
		name := base
		if n > 1 {
			name = fmt.Sprintf("%s.%d", base, n)
		}
		path := filepath.Join(dir, name)
		f, openErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if openErr != nil {
			if errors.Is(openErr, os.ErrExist) {
				continue
			}
			return "", fmt.Errorf("settings-drift: create %s: %w", path, openErr)
		}
		if _, writeErr := f.Write(data); writeErr != nil {
			_ = f.Close()
			return "", fmt.Errorf("settings-drift: write %s: %w", path, writeErr)
		}
		if closeErr := f.Close(); closeErr != nil {
			return "", fmt.Errorf("settings-drift: write %s: %w", path, closeErr)
		}
		return path, nil
	}
	return "", fmt.Errorf("settings-drift: could not find a free name for %s in %s", base, dir)
}

// recordSettingsDriftResolution appends ONE resolution row when — and only
// when — the most recent row naming this worktree is an unresolved detection.
//
// Both halves of that condition are load-bearing:
//
//   - Only when there is something to resolve. Every `acquire` runs this gate,
//     and almost every tree is clean, so a row written on each clean pass
//     would bury the detections it sits among under its own volume.
//
//   - Only the most recent row, so a resolution is not written again on every
//     later pass. The resolution itself is the stop condition.
//
// The ledger is READ here to make that decision, and only appended to — no row
// is edited, none is removed, and a detection's preserved copy is untouched.
// A missing ledger means there has never been a detection: nothing to resolve,
// and nothing is created, so a project that has never drifted keeps an empty
// state directory.
//
// A read failure is returned rather than swallowed, but it reaches the caller
// as PreserveErr, which by construction cannot change the verdict — the
// verdict is the match count, and it is already `clean`.
func recordSettingsDriftResolution(p SettingsDriftParams, worktree string) error {
	open, found, err := lastSettingsDriftRowForWorktree(p.Root, worktree)
	if err != nil || !found {
		return err
	}
	if open.Status == SettingsDriftRowResolved {
		return nil
	}
	return appendSettingsDriftLedger(p.Root, settingsDriftLedgerRow{
		MeasuredAt: time.Now().UTC().Format(time.RFC3339Nano),
		Card:       p.Card,
		Branch:     p.Branch,
		Worktree:   worktree,
		SourcePath: filepath.Join(worktree, filepath.FromSlash(SettingsDriftWatchedPath)),
		// The measured count, which is what made this a resolution. Not a
		// default: a `clean` verdict IS a count of zero.
		MatchCount:         0,
		Status:             SettingsDriftRowResolved,
		ResolvesSHA256:     open.SHA256,
		ResolvesMeasuredAt: open.MeasuredAt,
	})
}

// lastSettingsDriftRowForWorktree returns the final row naming worktree.
//
// An absent `status` is read as `detected`. Every row written before t765 has
// no such field — all five in this repository's own ledger — and reading the
// absence as anything else would leave them permanently unresolvable.
//
// A line that does not parse is skipped rather than fatal: this file is
// appended to by concurrent lanes, and one torn or hand-edited line must not
// stop the rest of the record from being read.
func lastSettingsDriftRowForWorktree(root, worktree string) (settingsDriftLedgerRow, bool, error) {
	path := filepath.Join(SettingsDriftDir(root), SettingsDriftLedgerName)
	data, err := os.ReadFile(path) // #nosec G304 -- path is <root>/.moai/state/settings-drift/ledger.jsonl
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return settingsDriftLedgerRow{}, false, nil
		}
		return settingsDriftLedgerRow{}, false, fmt.Errorf("settings-drift: read %s: %w", path, err)
	}
	var last settingsDriftLedgerRow
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var row settingsDriftLedgerRow
		if json.Unmarshal([]byte(line), &row) != nil {
			continue
		}
		if row.Worktree != worktree {
			continue
		}
		if row.Status == "" {
			row.Status = SettingsDriftRowDetected
		}
		last, found = row, true
	}
	return last, found, nil
}

// appendSettingsDriftLedger appends one JSON line. Append mode, never a
// read-modify-write: two lanes hitting at once must both leave a row.
func appendSettingsDriftLedger(root string, row settingsDriftLedgerRow) error {
	dir := SettingsDriftDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("settings-drift: create %s: %w", dir, err)
	}
	line, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("settings-drift: encode ledger row: %w", err)
	}
	path := filepath.Join(dir, SettingsDriftLedgerName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("settings-drift: open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("settings-drift: append %s: %w", path, err)
	}
	return nil
}
