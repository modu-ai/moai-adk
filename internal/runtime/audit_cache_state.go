// audit_cache_state.go — SPEC-MCP-NEUTRALITY-001 M6 (card t1632, REQ-MN-020):
// the disk-persisted audit_cache store. Entries live under the project's
// state root instead of process memory, each entry carries the SHA-256 of an
// EXISTING verdict file as its evidence, and a lookup revalidates that file's
// existence and digest before reusing the entry — a verdict file that changed
// (or vanished) since the store is never reused as a pass.
package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// stateCacheDirRel is the store's location under the project root.
const stateCacheDirRel = ".moai/state/audit-cache"

// stateEntry is the on-disk form of one cached PASS verdict.
type stateEntry struct {
	SpecID           string    `json:"spec_id"`
	PlanArtifactHash string    `json:"plan_artifact_hash"`
	VerdictFileSHA   string    `json:"verdict_file_sha"`
	ReportPath       string    `json:"report_path"`
	AuditorVersion   string    `json:"auditor_version"`
	AuditAt          time.Time `json:"audit_at"`
}

// StateAuditCache is the AuditCache implementation the MCP audit_cache tool
// serves: in-process reads over an on-disk store, so an entry written by one
// process (or one server restart ago) is still a hit for the next.
type StateAuditCache struct {
	root string
}

// NewStateAuditCache builds the store over projectRoot's state root.
func NewStateAuditCache(projectRoot string) *StateAuditCache {
	return &StateAuditCache{root: projectRoot}
}

func (c *StateAuditCache) dir() string { return filepath.Join(c.root, stateCacheDirRel) }

func (c *StateAuditCache) entryPath(specID string) string {
	return filepath.Join(c.dir(), sanitizeStateKey(specID)+".json")
}

// sanitizeStateKey keeps a caller-supplied SPEC id inside the store directory —
// the same discipline the receipt store's sanitizeKey applies to its keys.
func sanitizeStateKey(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := b.String()
	if out == "" || out == "." || out == ".." {
		return "_"
	}
	return out
}

// ComputeHash delegates to the shared artifact hashing (which refuses a
// non-SPEC directory with the empty-input error).
func (c *StateAuditCache) ComputeHash(specDir string) (string, error) {
	return (&InMemoryCache{}).ComputeHash(specDir)
}

// resolveReportPath anchors a verdict report path to the project root when it
// is RELATIVE (card t1632 card-review P2-7): a path a caller wrote relative to
// its own working directory must not be re-interpreted against whichever cwd
// the STORE or the LOOKUP happens to run from — both resolve it against the
// project root the cache is bound to.
func (c *StateAuditCache) resolveReportPath(report string) string {
	if filepath.IsAbs(report) {
		return report
	}
	return filepath.Join(c.root, report)
}

// StoreVerdict records one PASS verdict whose evidence is an EXISTING verdict
// file: the file's SHA-256 is computed and stored WITH the entry, and a store
// whose verdict file does not exist (or cannot be read) writes nothing and
// returns the error — an entry without file evidence is exactly the
// self-attesting cache REQ-MN-020 removes.
func (c *StateAuditCache) StoreVerdict(specID, hash string, result *AuditResult) error {
	if result == nil || result.Verdict != VerdictPass {
		return fmt.Errorf("audit cache stores PASS verdicts only")
	}
	report := c.resolveReportPath(result.ReportPath)
	if strings.TrimSpace(report) == "" {
		return fmt.Errorf("audit cache store requires a verdict report path")
	}
	sum, err := fileSHA256(report)
	if err != nil {
		return fmt.Errorf("verdict file evidence: %w", err)
	}
	e := stateEntry{
		SpecID:           specID,
		PlanArtifactHash: hash,
		VerdictFileSHA:   sum,
		ReportPath:       report,
		AuditorVersion:   result.AuditorVersion,
		AuditAt:          result.AuditAt,
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.dir(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.entryPath(specID), append(data, '\n'), 0o644)
}

// Lookup returns the cached entry when one exists for (specID, hash) AND its
// verdict file still exists with the SAME SHA-256 the store recorded — the
// reuse validation REQ-MN-020 requires before any reuse. The stored SpecID is
// compared too (sync-audit r3 F4): sanitizeStateKey folds SPEC IDs that differ
// only by a separator onto one file name, so with identical artifact hashes a
// lookup keyed by the colliding spelling must never be served another SPEC's
// PASS. The `now` parameter is vestigial and ignored.
func (c *StateAuditCache) Lookup(specID, hash string, _ time.Time) (*CachedEntry, bool) {
	data, err := os.ReadFile(c.entryPath(specID))
	if err != nil {
		return nil, false
	}
	var e stateEntry
	if err := json.Unmarshal(data, &e); err != nil || e.SpecID != specID || e.PlanArtifactHash != hash {
		return nil, false
	}
	sum, err := fileSHA256(c.resolveReportPath(e.ReportPath))
	if err != nil || sum != e.VerdictFileSHA {
		// The verdict file changed (or vanished) since the store: the entry is
		// evidence for a state that no longer exists — a miss, never a hit.
		return nil, false
	}
	return &CachedEntry{
		AuditAt:          e.AuditAt,
		AuditorVersion:   e.AuditorVersion,
		ReportPath:       e.ReportPath,
		PlanArtifactHash: e.PlanArtifactHash,
		VerdictFileSHA:   e.VerdictFileSHA,
	}, true
}

// Store implements the AuditCache interface: PASS results with an existing
// verdict file are stored; anything else is refused silently at this
// signature (the error-carrying surface is StoreVerdict).
func (c *StateAuditCache) Store(specID, hash string, result *AuditResult) {
	_ = c.StoreVerdict(specID, hash, result)
}

// maxVerdictFileBytes bounds how much of a verdict file is read as evidence;
// a verdict report is prose measured in kilobytes, so a few MiB is ample.
const maxVerdictFileBytes = 4 << 20

// fileSHA256 computes the hex SHA-256 of a verdict file's bytes. Only a
// REGULAR file is read (card t1632 card-review P2-B): a FIFO or device blocks
// os.ReadFile, and a symlink is refused unopened — the policy of
// internal/auditreceipt, where a link is bound as a link and never followed.
// The read is capped at maxVerdictFileBytes; a larger file is refused.
func fileSHA256(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file (mode %s)", path, info.Mode().Type())
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxVerdictFileBytes+1))
	if err != nil {
		return "", err
	}
	if n > maxVerdictFileBytes {
		return "", fmt.Errorf("%s exceeds the %d-byte verdict file limit", path, maxVerdictFileBytes)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
