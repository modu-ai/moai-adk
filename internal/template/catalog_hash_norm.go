package template

// @MX:NOTE: [AUTO] Shared hash normalization for catalog drift detection.
// Used by gen-catalog-hashes.go (T-007) and catalog_tier_audit_test.go (T-016).
// CRITICAL: This function MUST produce byte-identical output on all platforms
// (Windows CRLF, macOS LF, Linux LF) to ensure hash stability in CI.
// REQ: REQ-007, REQ-022, REQ-023, R7 (hash determinism Windows CI).

import (
	"bytes"
	"strings"
)

// NormalizeForHash normalizes raw file content for reproducible sha256 hashing.
//
// Normalization rules (HARD — must match between gen-catalog-hashes.go and audit suite):
//  1. Convert CRLF → LF (uniform line endings across Windows/macOS/Linux)
//  2. Strip trailing whitespace from each line (preserve internal whitespace)
//  3. Ensure exactly one trailing newline at end of content
//
// This function is the single source of truth for hash normalization.
// Both the hash generation helper (scripts/gen-catalog-hashes.go) and the
// audit test (TestManifestHashFormat in catalog_tier_audit_test.go) MUST
// import this function to guarantee identical byte sequences across platforms.
//
// Note on skill entries: For skill directories, only the root SKILL.md or skill.md
// is hashed. Sub-files (workflows/*.md, references/*) are NOT included in v1 hash.
func NormalizeForHash(raw []byte) []byte {
	// Step 1: Normalize line endings — CRLF → LF
	normalized := bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	// Also handle lone CR (old Mac style)
	normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))

	// Step 2: Strip trailing whitespace from each line
	lines := strings.Split(string(normalized), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	// Step 3: Join and ensure exactly one trailing newline
	content := strings.Join(lines, "\n")
	content = strings.TrimRight(content, "\n")
	content += "\n"

	return []byte(content)
}
