package kanban

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Artifact file names inside a deep-scan results directory. Both are produced
// by the review workflow agents; nothing in this package writes either one.
const (
	revisionFileName = "revision.json"
	findingsFileName = "findings.jsonl"
)

// scopeRepo is the only scope value a whole-tree scan can carry. A narrower
// scope examined less than sync is about to document, so it never matches.
const scopeRepo = "repo"

// Revision mirrors the deep-scan results directory's revision.json.
//
// The schema carries no completion or status field, so this record alone cannot
// distinguish an aborted scan that stamped it from a completed one. That gap is
// why RevisionMatch checks findings.jsonl before it ever reads this file.
type Revision struct {
	// ScannedCommit is the head SHA the scan ran against.
	ScannedCommit string `json:"scanned_commit"`

	// EffortTier is the scan's effort level. It is deliberately NOT part of any
	// predicate here: an effort level is not a rigor rung, and a max-effort
	// single-pass scan would clear any effort floor while still being
	// rigor-reduced. The rung travels on the kanban state record instead.
	EffortTier string `json:"effort_tier"`

	// WorkingTreeIncluded reports whether uncommitted edits were scanned.
	WorkingTreeIncluded bool `json:"working_tree_included"`

	// Scope is the scan's breadth; only scopeRepo can match.
	Scope string `json:"scope"`

	// GeneratedAt is the instant the scan stamped this file.
	GeneratedAt string `json:"generated_at"`
}

// LoadRevision reads and decodes the revision.json at path.
//
// An absent, unreadable, or malformed file is an error rather than a zero-valued
// Revision, so a caller can never mistake "nothing was recorded" for "a scan
// recorded nothing".
func LoadRevision(path string) (*Revision, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load revision: %w", err)
	}

	var rev Revision
	if err := json.Unmarshal(raw, &rev); err != nil {
		return nil, fmt.Errorf("load revision: decoding %s: %w", path, err)
	}
	return &rev, nil
}

// @MX:ANCHOR: [AUTO] revision-match predicate — the dedup gate's fail-safe default is FALSE
// @MX:REASON: this predicate decides whether an independent security analysis is suppressed, so every failure mode must resolve toward running the check; an implementation that returned true on absence, on a decode failure, or on an unrecognized scope would silently skip the analysis it exists to guard
//
// Matches is the pure revision-match predicate over an already-loaded Revision.
//
// It returns true only when every conjunct holds: rev is non-nil, its
// scanned_commit equals headSHA, its scope is repo, and — where the tree is
// dirty — the scan included the working tree. A clean tree makes
// WorkingTreeIncluded irrelevant, because there were no uncommitted edits for
// the scan to have missed.
//
// Every other input yields false, and false means the analysis runs.
func Matches(rev *Revision, headSHA string, treeDirty bool) bool {
	if rev == nil {
		return false
	}
	if headSHA == "" || rev.ScannedCommit != headSHA {
		return false
	}
	if rev.Scope != scopeRepo {
		return false
	}
	if treeDirty && !rev.WorkingTreeIncluded {
		return false
	}
	return true
}

// RevisionMatch is the directory-level revision-match predicate: the composition
// the dedup gate consults.
//
// The order is deliberate. findings.jsonl is checked before revision.json
// because revision.json cannot answer the question it appears to answer — it
// carries no completion field, so an aborted run that stamped it reads exactly
// like a finished one. findings.jsonl is the completion signal: a clean scan
// writes it with zero lines, which this predicate accepts, while an aborted scan
// characteristically never writes it at all, which this predicate rejects.
func RevisionMatch(deepScanDir, headSHA string, treeDirty bool) bool {
	if deepScanDir == "" {
		return false
	}
	info, err := os.Stat(deepScanDir)
	if err != nil || !info.IsDir() {
		return false
	}
	if !findingsComplete(filepath.Join(deepScanDir, findingsFileName)) {
		return false
	}

	rev, err := LoadRevision(filepath.Join(deepScanDir, revisionFileName))
	if err != nil {
		return false
	}
	return Matches(rev, headSHA, treeDirty)
}

// findingsComplete reports whether the findings file exists and every one of its
// lines parses as JSON. A zero-line file is complete — that is a scan that ran
// and found nothing, which is the outcome the whole gate is hoping for.
//
// Blank lines are tolerated as formatting, not treated as records.
func findingsComplete(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(trimSpaceBytes(line)) == 0 {
			continue
		}
		if !json.Valid(line) {
			return false
		}
	}
	return scanner.Err() == nil
}

// trimSpaceBytes strips ASCII whitespace without allocating a string.
func trimSpaceBytes(b []byte) []byte {
	start := 0
	for start < len(b) && isASCIISpace(b[start]) {
		start++
	}
	end := len(b)
	for end > start && isASCIISpace(b[end-1]) {
		end--
	}
	return b[start:end]
}

func isASCIISpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\f'
}

// @MX:ANCHOR: [AUTO] composed suppression decision — an allow-list, never a deny-list
// @MX:REASON: the kanban state record is best-effort and its rung field lands independently of the results directory, so a record carrying a results directory but no rung is reachable; a "not DEGRADED" deny-list would treat that absent rung as permission to suppress, which is the fail-open shape this allow-list exists to prevent
//
// SuppressStep0551 reports whether the sync-phase security analysis may be
// suppressed and its findings inherited.
//
// Suppression requires BOTH a passing revision-match predicate AND a rung that
// was actually recorded as PRIMARY or FALLBACK. A nil rung (never written), a
// recorded-empty rung, a DEGRADED rung, and any unrecognized value all fall
// outside the allow-list and yield no suppression.
func SuppressStep0551(rung *Rung, revisionMatched bool) bool {
	if !revisionMatched || rung == nil {
		return false
	}
	switch *rung {
	case RungPrimary, RungFallback:
		return true
	default:
		return false
	}
}
