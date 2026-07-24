package security

// Layer-1 buffer scanner for the in-session security guardian
// (SPEC-SEC-GUARDIAN-001). ScanBuffer is a deterministic, in-process regex scan
// over a single write buffer (REQ-SG-013): it imports NO os/exec and NO net/http,
// makes no language-model call, and never spawns a subprocess. It is advisory
// only — it produces findings; it never blocks (REQ-SG-014). The git-invoking and
// stdin-parsing plumbing lives in guardian.go, never here, so this file stays a
// pure pattern matcher (AC-SG-004: scan.go must contain no os/exec / net/http).

import (
	"strings"

	"github.com/modu-ai/moai-adk/internal/config"
)

// ScanBuffer scans a single text buffer for dangerous patterns using the
// single-source guardian table (REQ-SG-010). It returns one finding per pattern
// match, annotated with the 1-based line number. The scan is bounded to
// config.DefaultSecurityMaxScanBytes so a very large write is not scanned
// unboundedly (edge case: large payload). Binary content is skipped entirely so
// a Write to a non-source file does not spam findings (edge case: binary file).
//
// @MX:ANCHOR: [AUTO] ScanBuffer is the Layer-1 in-process regex scan shared by Layers 1/2/3 (Layer 2 scans diff hunks, Layer 3 scans each cross-file buffer).
// @MX:REASON: [AUTO] fan_in >= 3 — HandleSecurityScan (L1), ScanDiff (L2), CrossFileScan (L3) all call ScanBuffer; it is the single regex entry point.
func ScanBuffer(content string) []GuardianFinding {
	if content == "" {
		return nil
	}

	// Bound the scanned input (edge case: large payload). We truncate on a rune
	// boundary approximation by simply slicing bytes — the pattern engine reads
	// UTF-8 safely and a mid-rune cut at the tail cannot produce a false match.
	if len(content) > config.DefaultSecurityMaxScanBytes {
		content = content[:config.DefaultSecurityMaxScanBytes]
	}

	// Skip binary content (edge case: binary / non-source file). A NUL byte is a
	// reliable binary sentinel across the 16 languages' source files (all text).
	if looksBinary(content) {
		return nil
	}

	var findings []GuardianFinding
	lines := strings.Split(content, "\n")
	for _, class := range guardianClasses {
		for _, pat := range class.Patterns {
			for i, line := range lines {
				loc := pat.FindStringIndex(line)
				if loc == nil {
					continue
				}
				findings = append(findings, GuardianFinding{
					Class:    class.Name,
					Severity: class.Severity,
					Message:  class.Description,
					Line:     i + 1,
					Match:    strings.TrimSpace(line[loc[0]:loc[1]]),
				})
			}
		}
	}
	return findings
}

// looksBinary reports whether content appears to be a binary (non-source) blob.
// A NUL byte in the leading window is the sentinel; source files in all 16
// supported languages are NUL-free text.
func looksBinary(content string) bool {
	window := content
	const probe = 8000
	if len(window) > probe {
		window = window[:probe]
	}
	return strings.IndexByte(window, 0x00) >= 0
}
