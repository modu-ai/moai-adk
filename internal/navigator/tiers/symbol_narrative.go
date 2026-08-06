package tiers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"log/slog"
)

// narrativeMetadata is the metadata.json sidecar schema for a per-symbol
// narrative file (REQ-NS3-012). LastUpdatedCommit tracks the git baseline
// at which the narrative was last drafted; LastRecordHash is the hash of
// the deterministic record at draft time. The CodeWiki `--compare-to` gate
// re-drafts ONLY when the deterministic record changed since this commit.
type narrativeMetadata struct {
	LastUpdatedCommit string `json:"last_updated_commit"`
	LastRecordHash    string `json:"last_record_hash"`
}

// narrativeSlotPath returns the deterministic slot path for a symbol's
// narrative file (relative to project root).
func narrativeSlotPath(identifier string) string {
	safe := symbolFilename(identifier)
	return filepath.Join(".moai", "project", "navigator", "symbols", safe+".md")
}

// narrativeMetadataPath returns the metadata.json sidecar path.
func narrativeMetadataPath(identifier string) string {
	safe := symbolFilename(identifier)
	return filepath.Join(".moai", "project", "navigator", "symbols", safe+".metadata.json")
}

// symbolFilename rewrites an identifier into a filesystem-safe filename
// (replaces path separators + dots with underscores).
func symbolFilename(identifier string) string {
	s := strings.ReplaceAll(identifier, "/", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "(", "_")
	s = strings.ReplaceAll(s, ")", "_")
	s = strings.ReplaceAll(s, "*", "ptr")
	return s
}

// hashDeterministicRecord produces the hash used by the CodeWiki gate. The
// hash covers the deterministic fields (signature, declaration, references)
// but NOT the narrative — so a re-draft is triggered only when the
// deterministic record changes.
func hashDeterministicRecord(r SymbolEnrichment) string {
	// Sorted references for byte-stable hashing.
	refs := make([]SymbolRef, len(r.References))
	copy(refs, r.References)
	sortRefsStable(refs)
	h := sha256.New()
	h.Write([]byte(r.Identifier))
	h.Write([]byte{0})
	h.Write([]byte(r.Signature))
	h.Write([]byte{0})
	h.Write([]byte(r.DeclarationPath))
	h.Write([]byte{0})
	// Line as decimal string (deterministic).
	h.Write([]byte(itoaLine(r.DeclarationLine)))
	h.Write([]byte{0})
	for _, ref := range refs {
		h.Write([]byte(ref.Path))
		h.Write([]byte{0})
		h.Write([]byte(itoaLine(ref.Line)))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// sortRefsStable sorts refs by (Path, Line) for byte-stable hashing.
func sortRefsStable(refs []SymbolRef) {
	for i := 1; i < len(refs); i++ {
		for j := i; j > 0; j-- {
			a, b := refs[j-1], refs[j]
			if a.Path > b.Path || (a.Path == b.Path && a.Line > b.Line) {
				refs[j-1], refs[j] = refs[j], refs[j-1]
				continue
			}
			break
		}
	}
}

// itoaLine is a strconv.Itoa-free integer-to-string for the hash path.
func itoaLine(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}

// shouldRedraft reports whether the narrative for the symbol whose metadata
// lives at metadataPath should be re-drafted, given the current deterministic
// record hash (REQ-NS3-012 CodeWiki `--update`/`--compare-to` gate).
//
//   - metadata absent → redraft (first run).
//   - metadata.LastRecordHash == currentHash → DO NOT redraft (gate holds).
//   - metadata.LastRecordHash != currentHash → redraft.
//
// @MX:ANCHOR: [AUTO] CodeWiki narrative gate; load-bearing for the LLM-cost-bounding invariant
// @MX:REASON: without this gate, every run re-drafts every narrative — defeating the CodeWiki hierarchical-decomposition cost bound (design.md §5)
// @MX:SPEC:SPEC-NAVIGATOR-SYNC-003
func shouldRedraft(metadataPath, currentHash string) bool {
	body, err := os.ReadFile(metadataPath)
	if err != nil {
		// Absent metadata → first run → redraft.
		return true
	}
	var m narrativeMetadata
	if err := json.Unmarshal(body, &m); err != nil {
		slog.Debug("tiers: narrative metadata unparseable, redrafting", "path", metadataPath)
		return true
	}
	return m.LastRecordHash != currentHash
}

// writeNarrativeMetadata persists the metadata sidecar atomically (write +
// rename). Used by the narrative refresh step after a re-draft.
func writeNarrativeMetadata(metadataPath, commitSHA, recordHash string) error {
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0o755); err != nil {
		return err
	}
	m := narrativeMetadata{LastUpdatedCommit: commitSHA, LastRecordHash: recordHash}
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := metadataPath + ".tmp"
	if err := os.WriteFile(tmp, append(body, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, metadataPath)
}
