package cli

// @MX:ANCHOR: [AUTO] update_noise.go — noise-suppression ledgers for `moai update` reserved + merge fallback warnings
// @MX:REASON: invariant contract — silent re-run for known reserved files (ack ledger) + silent fallback for transient 3-way merge regressions (3-strike threshold); user-visible signal preserved on first occurrence and on hash drift
// @MX:SPEC: SPEC-V3R6-UPDATE-NOISE-001

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// updateVerboseMode carries the `--verbose` flag through the update execution
// path without requiring signature changes to every helper. Set once by
// runUpdate at the top of the update command and consulted by
// checkReservedCollision + recordMergeFallback. `moai update` is single-process
// sequential so no synchronization is needed.
//
// Default value (false) preserves the production default behavior of
// noise-suppression. Tests can set this directly when exercising verbose
// branches.
var updateVerboseMode bool

// reservedAckLedgerRelPath is the path of the reserved-ack ledger relative to
// project root. The ledger is per-machine and gitignored.
const reservedAckLedgerRelPath = ".moai/state/reserved-acknowledged.json"

// mergeHistoryLedgerRelPath is the path of the 3-way merge fallback counter
// relative to project root. The ledger is per-machine and gitignored.
const mergeHistoryLedgerRelPath = ".moai/cache/merge-history.json"

// fallbackAdvisoryThreshold is the count at which the merge-history advisory
// is emitted exactly once per relPath. Counts below the threshold are silent
// (REQ-UN-007); the advisory fires on the count crossing the threshold
// (REQ-UN-008); counts above the threshold are silent until the counter is
// reset by a 3-way success (REQ-UN-009).
const fallbackAdvisoryThreshold = 3

// reservedAckEntry is the per-path record stored inside the reserved-ack ledger.
// The JSON field names are part of the ledger contract — do not rename without
// a migration plan (REQ-UN-001).
type reservedAckEntry struct {
	SHA256         string `json:"sha256"`
	AcknowledgedAt string `json:"acknowledged_at"`
}

// mergeHistoryEntry is the per-relPath record stored inside the merge-history
// ledger. The JSON field names are part of the ledger contract (REQ-UN-006).
type mergeHistoryEntry struct {
	FallbackCount int    `json:"fallback_count"`
	LastFailedAt  string `json:"last_failed_at"`
}

// loadReservedAckLedger reads and parses the reserved-ack ledger from disk.
// REQ-UN-011 fallback safety: if the file is absent or unparseable, an empty
// map is returned silently (no error surfaced).
func loadReservedAckLedger(projectRoot string) map[string]reservedAckEntry {
	path := filepath.Join(projectRoot, reservedAckLedgerRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]reservedAckEntry{}
	}
	ledger := map[string]reservedAckEntry{}
	if jsonErr := json.Unmarshal(data, &ledger); jsonErr != nil {
		// REQ-UN-011: corrupt JSON → re-initialize as empty map.
		return map[string]reservedAckEntry{}
	}
	return ledger
}

// saveReservedAckLedger persists the ledger to disk using an atomic temp+rename
// pattern (plan.md §3.2). The parent directory is created if absent.
func saveReservedAckLedger(projectRoot string, ledger map[string]reservedAckEntry) error {
	dir := filepath.Join(projectRoot, ".moai", "state")
	if err := os.MkdirAll(dir, defs.DirPerm); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}
	return atomicWriteJSON(filepath.Join(dir, "reserved-acknowledged.json"), ledger)
}

// loadMergeHistoryLedger reads and parses the merge-history ledger. REQ-UN-011
// applies symmetrically (absent or unparseable → empty map).
func loadMergeHistoryLedger(projectRoot string) map[string]mergeHistoryEntry {
	path := filepath.Join(projectRoot, mergeHistoryLedgerRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]mergeHistoryEntry{}
	}
	ledger := map[string]mergeHistoryEntry{}
	if jsonErr := json.Unmarshal(data, &ledger); jsonErr != nil {
		return map[string]mergeHistoryEntry{}
	}
	return ledger
}

// saveMergeHistoryLedger persists the merge-history ledger atomically.
func saveMergeHistoryLedger(projectRoot string, ledger map[string]mergeHistoryEntry) error {
	dir := filepath.Join(projectRoot, ".moai", "cache")
	if err := os.MkdirAll(dir, defs.DirPerm); err != nil {
		return fmt.Errorf("mkdir cache dir: %w", err)
	}
	return atomicWriteJSON(filepath.Join(dir, "merge-history.json"), ledger)
}

// atomicWriteJSON writes the JSON-encoded value to targetPath atomically via
// a temp file + rename pattern. The temp file lives in the same directory as
// the target so the rename stays on the same filesystem.
func atomicWriteJSON(targetPath string, value any) error {
	dir := filepath.Dir(targetPath)
	tmp, err := os.CreateTemp(dir, ".tmp-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if encErr := enc.Encode(value); encErr != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("encode JSON: %w", encErr)
	}
	if closeErr := tmp.Close(); closeErr != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", closeErr)
	}
	if renameErr := os.Rename(tmpName, targetPath); renameErr != nil {
		cleanup()
		return fmt.Errorf("rename to target: %w", renameErr)
	}
	return nil
}

// sha256FileHex computes the SHA-256 hash of the file at path and returns it
// as a 64-character hex string. Returns empty string + error on I/O failure.
func sha256FileHex(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, copyErr := io.Copy(h, f); copyErr != nil {
		return "", copyErr
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// shouldEmitReservedWarning encapsulates the ack-ledger consultation policy
// (REQ-UN-002 + REQ-UN-005). Returns true iff the warning should be emitted
// for the given reserved-name occurrence.
//
// The decision matrix is:
//
//	verbose=true                        → always emit
//	verbose=false + ack absent          → emit (first occurrence)
//	verbose=false + ack present + drift → emit (hash drift re-emit per REQ-UN-004)
//	verbose=false + ack present + same  → silent
func shouldEmitReservedWarning(ledger map[string]reservedAckEntry, name, currentHash string, verbose bool) bool {
	if verbose {
		return true
	}
	entry, acknowledged := ledger[name]
	if !acknowledged {
		return true
	}
	return entry.SHA256 != currentHash
}

// recordReservedAck stamps the ledger entry for the given name with the
// current hash + timestamp (REQ-UN-003).
func recordReservedAck(ledger map[string]reservedAckEntry, name, currentHash string) {
	ledger[name] = reservedAckEntry{
		SHA256:         currentHash,
		AcknowledgedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// recordMergeFallback updates the merge-history ledger for relPath and, when
// appropriate, emits either the legacy warning (verbose) or the threshold
// advisory (REQ-UN-007 + REQ-UN-008 + REQ-UN-009 + REQ-UN-010).
//
// success=true means mergeYAML3Way succeeded — counter is reset to 0
// (REQ-UN-009) regardless of verbose flag, and no output is emitted.
//
// success=false means mergeYAML3Way failed and the caller is about to fall
// back. The counter is incremented; output depends on verbose:
//   - verbose=true  → emit legacy "Warning: 3-way merge failed for X, falling
//     back to 2-way" message every time (REQ-UN-010).
//   - verbose=false → silent unless the post-increment count exactly equals
//     fallbackAdvisoryThreshold (3), in which case the advisory is emitted
//     exactly once (REQ-UN-008). Subsequent silent counts ≥4 stay silent
//     until a 3-way success resets the counter.
//
// errOut MUST NOT be nil — pass os.Stderr in production.
func recordMergeFallback(projectRoot, relPath string, success, verbose bool, errOut io.Writer) {
	ledger := loadMergeHistoryLedger(projectRoot)
	entry := ledger[relPath]
	if success {
		// REQ-UN-009: reset counter on 3-way success. No output regardless of verbose.
		if entry.FallbackCount != 0 {
			entry.FallbackCount = 0
			ledger[relPath] = entry
			_ = saveMergeHistoryLedger(projectRoot, ledger)
		}
		return
	}
	// Failure path: increment counter and decide whether to emit.
	entry.FallbackCount++
	entry.LastFailedAt = time.Now().UTC().Format(time.RFC3339)
	ledger[relPath] = entry
	_ = saveMergeHistoryLedger(projectRoot, ledger)

	if verbose {
		// REQ-UN-010: every fallback emits the legacy warning under --verbose.
		_, _ = fmt.Fprintf(errOut, "Warning: 3-way merge failed for %s, falling back to 2-way\n", relPath)
		return
	}
	// REQ-UN-008: emit advisory exactly once when post-increment count equals
	// the threshold. Counts 1, 2 are silent; counts >=4 stay silent.
	if entry.FallbackCount == fallbackAdvisoryThreshold {
		_, _ = fmt.Fprintf(errOut, "hint: 'moai update -c' to resync templates for %s\n", relPath)
	}
}
