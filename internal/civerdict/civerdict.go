// Package civerdict records remote CI verdicts per head SHA as on-disk
// evidence (SPEC-CI-VERDICT-PRODUCER-001): one JSON record per judged head
// under .moai/state/ci-verdicts/ in the tree the producer runs in, written by
// the moai ci-verdict producer verb and consumed by the escalation detector's
// contradictory-evidence CI limb (REQ-CV-001). The package owns the five-field
// record schema (REQ-CV-004) and the atomic store; it writes evidence files
// and nothing else — it never invokes the escalation detector, any checkpoint,
// or any hook path (REQ-CV-005).
//
// @MX:ANCHOR: [AUTO] shared evidence schema for the ci-verdict producer verb and the escalation detector's CI limb
// @MX:REASON: two packages import this schema (internal/cli producer, internal/escalation detector); a field rename here silently desyncs the recorded evidence from its consumer
package civerdict

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// VerdictDir is the record directory under a project root — gitignored
// runtime state in the established evidence-persistence namespace
// (.moai/state/, class-5 precedent: .moai/state/audit-multi/).
const VerdictDir = ".moai/state/ci-verdicts"

// Record conclusions (REQ-CV-004). neutral is a completed observation: an
// observed verdict that does not contradict the local pass (skipped or
// cancelled CI), not an unobserved limb.
const (
	ConclusionSuccess = "success"
	ConclusionFailure = "failure"
	ConclusionNeutral = "neutral"
)

// ProducerID is the identity the producer verb writes into the producer
// field when the caller does not override it.
const ProducerID = "moai ci-verdict"

// Record is one recorded CI verdict for a head (REQ-CV-004): the head SHA is
// both the record's filename stem and its pinned head, ObservedAt is RFC 3339.
type Record struct {
	HeadSHA    string `json:"head_sha"`
	Conclusion string `json:"conclusion"`
	RunID      string `json:"run_id"` // empty when the backend reports none
	ObservedAt string `json:"observed_at"`
	Producer   string `json:"producer"`
}

// Validate reports whether the record satisfies REQ-CV-004: a head SHA, a
// known conclusion, and an RFC 3339 observed_at.
func (r Record) Validate() error {
	if r.HeadSHA == "" {
		return fmt.Errorf("civerdict: head_sha is empty")
	}
	switch r.Conclusion {
	case ConclusionSuccess, ConclusionFailure, ConclusionNeutral:
	default:
		return fmt.Errorf("civerdict: conclusion %q is not one of success|failure|neutral", r.Conclusion)
	}
	if _, err := time.Parse(time.RFC3339, r.ObservedAt); err != nil {
		return fmt.Errorf("civerdict: observed_at %q is not RFC 3339: %w", r.ObservedAt, err)
	}
	return nil
}

// ParseInput parses an offline input file naming head, conclusion, run id,
// observed_at, and producer (REQ-CV-002) into a validated record. The
// recorded bytes of a Save(ParseInput(...)) are indistinguishable in schema
// from a fetched verdict.
func ParseInput(data []byte) (Record, error) {
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return Record{}, fmt.Errorf("civerdict: parse input: %w", err)
	}
	if err := r.Validate(); err != nil {
		return Record{}, err
	}
	return r, nil
}

// Path returns the on-disk record path for a head: <root>/.moai/state/ci-verdicts/<head>.json.
func Path(projectRoot, headSHA string) string {
	return filepath.Join(projectRoot, VerdictDir, headSHA+".json")
}

// Save writes r atomically — a temp file in the target directory, then
// renamed into place, never a partial in-place write a concurrent reader
// could observe mid-write (REQ-CV-004; the verify store's Save discipline).
// Re-recording the same head with the same content is a byte-identical
// rewrite; differing content overwrites (last-writer-wins).
func Save(projectRoot string, r Record) error {
	if err := r.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("civerdict: marshal: %w", err)
	}
	dir := filepath.Join(projectRoot, VerdictDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("civerdict: mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".civerdict-*.tmp")
	if err != nil {
		return fmt.Errorf("civerdict: tmp create: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("civerdict: tmp write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("civerdict: tmp close: %w", err)
	}
	final := Path(projectRoot, r.HeadSHA)
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("civerdict: rename %s: %w", final, err)
	}
	cleanup = false
	return nil
}

// Load reads the record for head. A missing file returns (nil, nil) —
// absence is the plain no-record path, never an error. A stored-head
// mismatch (filename-collision defense, the verify store's Load discipline)
// is also treated as absent.
func Load(projectRoot, headSHA string) (*Record, error) {
	data, err := os.ReadFile(Path(projectRoot, headSHA))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("civerdict: load %s: %w", headSHA, err)
	}
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("civerdict: parse %s: %w", headSHA, err)
	}
	if r.HeadSHA != headSHA {
		return nil, nil
	}
	return &r, nil
}

// Loaded pairs one parsed record with its file identity, for consumers that
// need the source file name or raw bytes.
type Loaded struct {
	// Name is the record file's base name.
	Name string
	// Record is the parsed record.
	Record Record
	// Data is the raw file bytes (the detector's freshEvidence content-hash
	// gate consumes them).
	Data []byte
}

// LoadAll globs the record directory and parses every record file. An
// unreadable file is skipped and reported by base name, so the caller can
// list it not-observed (the convergenceFile parse precedent in
// internal/escalation). Absent directory yields empty results, not an error.
func LoadAll(projectRoot string) (loaded []Loaded, unreadable []string) {
	paths, _ := filepath.Glob(filepath.Join(projectRoot, VerdictDir, "*.json"))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			unreadable = append(unreadable, filepath.Base(p))
			continue
		}
		var r Record
		if err := json.Unmarshal(data, &r); err != nil {
			unreadable = append(unreadable, filepath.Base(p))
			continue
		}
		loaded = append(loaded, Loaded{Name: filepath.Base(p), Record: r, Data: data})
	}
	return loaded, unreadable
}
