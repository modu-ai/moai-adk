package escalation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// RecordSchemaVersion is the escalation record frontmatter schema version
// (spec.md §I.1).
const RecordSchemaVersion = 1

// Record kinds (spec.md §I.1). Only contract and operational records can make
// a card needs-decision; revoke records are reserved for A3 and never count.
const (
	KindContract    = "contract"
	KindOperational = "operational"
	KindRevoke      = "revoke"
)

// Record statuses.
const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
)

// Detection classes (spec.md §B.1). Classes 1-6 are the contract's
// escalate_on tokens; 7-10 are operational trips.
const (
	ClassAcceptanceChange      = "acceptance-change"
	ClassInvariantViolation    = "invariant-violation"
	ClassOwnershipMove         = "ownership-move"
	ClassNewArchitectureOrAPI  = "new-architecture-or-api"
	ClassContradictoryEvidence = "contradictory-evidence"
	ClassIrreversibleAction    = "irreversible-action"
	ClassBudgetExceeded        = "budget-exceeded"
	ClassSameDiagnosticRepeat  = "same-diagnostic-repeat"
	ClassAuditFailAtRetryCap   = "audit-fail-at-retry-cap"
	ClassDetectionDisarmed     = "detection-disarmed"
)

// Revoke classes reserved for A3 (spec.md §I.2). The detector never writes
// them.
const (
	ClassRevokeOperator   = "revoke-operator"
	ClassRevokeOnDecision = "revoke-on-decision"
)

// recordDirName is the per-card record directory under .moai/reports/<card>/.
const recordDirName = "escalation"

// Record is one escalation record: YAML frontmatter per spec.md §I.1 plus the
// three body sections (Observation, Options, Not observed).
type Record struct {
	SchemaVersion int      `yaml:"schema_version"`
	Card          string   `yaml:"card"`
	Spec          string   `yaml:"spec"`
	Kind          string   `yaml:"kind"`
	Class         string   `yaml:"class"`
	Fingerprint   string   `yaml:"fingerprint"`
	ContractRef   string   `yaml:"contract_ref"`
	EscalateOn    string   `yaml:"escalate_on"`
	Status        string   `yaml:"status"`
	Decider       string   `yaml:"decider"`
	Occurrences   int      `yaml:"occurrences"`
	HeadSHA       string   `yaml:"head_sha"`
	DetectedAt    string   `yaml:"detected_at"`
	UpdatedAt     string   `yaml:"updated_at"`
	NotObserved   []string `yaml:"not_observed"`

	// Body sections. They are rendered by Marshal and are not part of the
	// frontmatter; ParseRecord reads the frontmatter only.
	Observation string   `yaml:"-"`
	Options     []string `yaml:"-"`
}

// Marshal renders the record as Markdown with YAML frontmatter.
func (r Record) Marshal() ([]byte, error) {
	if r.NotObserved == nil {
		r.NotObserved = []string{}
	}
	fm, err := yaml.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("escalation: encode record frontmatter: %w", err)
	}
	var b bytes.Buffer
	b.WriteString("---\n")
	b.Write(fm)
	b.WriteString("---\n\n## Observation\n\n")
	b.WriteString(strings.TrimRight(r.Observation, "\n"))
	b.WriteString("\n\n## Options\n\n")
	for i, o := range r.Options {
		fmt.Fprintf(&b, "%d. %s\n", i+1, o)
	}
	b.WriteString("\n## Not observed\n\n")
	if len(r.NotObserved) == 0 {
		b.WriteString("- none\n")
	}
	for _, n := range r.NotObserved {
		b.WriteString("- " + n + "\n")
	}
	return b.Bytes(), nil
}

// errNoFrontmatter reports a record file without a leading frontmatter block.
var errNoFrontmatter = errors.New("escalation: record has no YAML frontmatter")

// ParseRecord decodes a record's frontmatter. Body fields stay empty.
func ParseRecord(data []byte) (Record, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return Record{}, errNoFrontmatter
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return Record{}, errNoFrontmatter
	}
	var r Record
	if err := yaml.Unmarshal([]byte(text[4:4+end+1]), &r); err != nil {
		return Record{}, fmt.Errorf("escalation: decode record frontmatter: %w", err)
	}
	return r, nil
}

// Fingerprint returns the first 16 lowercase hex characters of the SHA-256 of
// the class and its normalized observation parts (spec.md §I, design.md
// §C.10). Parts are separated by a NUL byte so no two part lists collide by
// concatenation.
func Fingerprint(class string, parts ...string) string {
	sum := sha256.Sum256([]byte(class + "\x00" + strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])[:16]
}

// RecordDir returns <worktree root>/.moai/reports/<card>/escalation.
func RecordDir(worktreeRoot, card string) string {
	return filepath.Join(worktreeRoot, ".moai", "reports", card, recordDirName)
}

// RecordPath returns the record file path for a class and fingerprint:
// <class>-<fingerprint>.md, or <class>-<fingerprint>-<n>.md for a re-trip
// after resolution with ordinal n >= 2 (REQ-AE-018, REQ-AE-020).
func RecordPath(worktreeRoot, card, class, fingerprint string, ordinal int) string {
	name := class + "-" + fingerprint
	if ordinal >= 2 {
		name += "-" + strconv.Itoa(ordinal)
	}
	return filepath.Join(RecordDir(worktreeRoot, card), name+".md")
}

// NeedsDecision reports whether the card has at least one contract or
// operational record with status open (REQ-AE-018). It reads only the card's
// record directory; the queue store is never read or written. A record that
// cannot be read or parsed is an error, never a silent "no decision needed".
func NeedsDecision(worktreeRoot, card string) (bool, error) {
	paths, err := filepath.Glob(filepath.Join(RecordDir(worktreeRoot, card), "*.md"))
	if err != nil {
		return false, fmt.Errorf("escalation: list records: %w", err)
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return false, fmt.Errorf("escalation: read record %s: %w", p, err)
		}
		r, err := ParseRecord(data)
		if err != nil {
			return false, fmt.Errorf("escalation: record %s: %w", p, err)
		}
		if r.Status == StatusOpen && (r.Kind == KindContract || r.Kind == KindOperational) {
			return true, nil
		}
	}
	return false, nil
}
