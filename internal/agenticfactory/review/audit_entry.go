package review

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"time"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// EntryStatus says whether an audit entry still describes the latest HEAD.
type EntryStatus string

const (
	EntryCurrent EntryStatus = "current"
	EntryStale   EntryStatus = "stale"
)

// AuditEntry is what review/ supplies for the audit section of an Issue's
// context ledger. Model and Effort are the values actually used, reported by
// the caller; they are never defaulted here.
type AuditEntry struct {
	Result       Verdict
	Model        string
	Effort       string
	Head         core.HeadSHA
	ArtifactHash string // sha256 hex of the verdict artifact
	At           time.Time
}

var (
	// ErrEntryInput: a required field was not observed, so no entry is made.
	ErrEntryInput = errors.New("review: audit entry input missing or malformed")
	hashPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// HashArtifact returns the sha256 hex digest of a verdict artifact's bytes.
func HashArtifact(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// NewAuditEntry builds the ledger entry from a completed record. Any verdict
// is recorded (FAIL and PASS-WITH-DEBT included); what it refuses is an entry
// whose model, effort, hash, time, or HEAD was not observed.
func NewAuditEntry(rec Record, model, effort, artifactHash string, at time.Time) (AuditEntry, error) {
	switch {
	case !rec.Reached(StateCompleted), !fullHead(rec.HeadSHA),
		model == "", effort == "", !hashPattern.MatchString(artifactHash), at.IsZero():
		return AuditEntry{}, ErrEntryInput
	}
	return AuditEntry{Result: rec.Verdict, Model: model, Effort: effort,
		Head: rec.HeadSHA, ArtifactHash: artifactHash, At: at}, nil
}

// Status is stale as soon as the latest HEAD differs from the audited one.
func (e AuditEntry) Status(latest core.HeadSHA) EntryStatus {
	if e.Head == latest {
		return EntryCurrent
	}
	return EntryStale
}
