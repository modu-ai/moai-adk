package security

// Layer-2 (turn-diff) and Layer-3 (commit cross-file) scanners for the
// in-session security guardian (SPEC-SEC-GUARDIAN-001). Both reuse the Layer-1
// pattern engine (ScanBuffer) over the single-source table — Layer 2 scans the
// added lines of a unified diff, Layer 3 scans each changed + related file and
// runs a cross-file data-flow heuristic (IDOR / broken-authorization / cross-file
// SSRF, REQ-SG-031).
//
// Like scan.go, this file is a PURE analyzer: it imports NO os/exec and NO
// net/http, makes no model/tool call, and never spawns a subprocess (AC-SG-004).
// Reading files from disk (os.ReadFile) IS permitted — it is not a subprocess
// and not a network call — so ReadRelatedFiles gathers the cross-file corpus
// here, while the git invocation that produces the diff / changed-file list
// lives in guardian.go.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ScanDiff scans the added lines of a unified diff for dangerous patterns
// (REQ-SG-020, Layer 2). Only added lines (`+` prefix, excluding the `+++`
// file header) are scanned, so a pattern that already existed and is merely
// shown as context is not re-flagged. It returns findings from the Layer-1
// engine run over the reconstructed added-line buffer.
//
// @MX:ANCHOR: [AUTO] ScanDiff is the Layer-2 turn-diff analyzer invoked at Stop by HandleSecurityTurn.
// @MX:REASON: [AUTO] fan_in >= 2 growing to 3 — HandleSecurityTurn (L2) + diff_test.go + the L3 changed-file path share the diff-scan contract; it is the sole added-line extractor.
func ScanDiff(diffText string) []GuardianFinding {
	if strings.TrimSpace(diffText) == "" {
		// Edge case: empty / whitespace diff at Stop — no-op cleanly.
		return nil
	}

	var added strings.Builder
	for _, line := range strings.Split(diffText, "\n") {
		// Added lines start with a single '+'. Exclude the '+++ b/file' header.
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added.WriteString(line[1:])
			added.WriteByte('\n')
		}
	}
	return ScanBuffer(added.String())
}

// ReadRelatedFiles gathers the changed files plus their sibling source files in
// the same directories, returning a path->content map for the cross-file scan
// (REQ-SG-030, Layer 3). It reads with os.ReadFile (no subprocess) and skips
// binary / oversized / unreadable files. relatedGlobDepth sibling files are the
// "related" corpus: single-file pattern matching misses multi-file flows, so the
// cross-file review reads the neighbours of each changed file.
func ReadRelatedFiles(projectRoot string, changedRelPaths []string) map[string]string {
	corpus := make(map[string]string)
	seenDirs := make(map[string]bool)

	readInto := func(abs, rel string) {
		// Normalize the corpus key to forward slashes so cross-file matching and
		// the map keys are platform-independent: git yields OS-native separators
		// on Windows (e.g. "handlers\route.js"), which would otherwise miss the
		// slash-keyed lookups. The filesystem read still uses the OS-native abs
		// path; only the logical corpus key is canonicalized.
		rel = filepath.ToSlash(rel)
		if _, ok := corpus[rel]; ok {
			return
		}
		info, err := os.Stat(abs)
		if err != nil || info.IsDir() || info.Size() > int64(maxCrossFileBytes) {
			return
		}
		data, err := os.ReadFile(abs)
		if err != nil || looksBinary(string(data)) {
			return
		}
		corpus[rel] = string(data)
	}

	for _, rel := range changedRelPaths {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		abs := filepath.Join(projectRoot, rel)
		readInto(abs, rel)

		// Add sibling files in the same directory (the "related" corpus).
		dir := filepath.Dir(abs)
		if seenDirs[dir] {
			continue
		}
		seenDirs[dir] = true
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			sibAbs := filepath.Join(dir, e.Name())
			sibRel, relErr := filepath.Rel(projectRoot, sibAbs)
			if relErr != nil {
				sibRel = sibAbs
			}
			readInto(sibAbs, sibRel)
		}
	}
	return corpus
}

// maxCrossFileBytes bounds a single related-file read so a large sibling does
// not blow the Layer-3 budget. It intentionally reuses the same magnitude as the
// Layer-1 scan cap.
const maxCrossFileBytes = 1048576

// CrossFileScan runs the cross-file data-flow review over a corpus of changed +
// related files (REQ-SG-031, Layer 3). It does two things:
//
//  1. Per-file pattern scan — runs the Layer-1 engine over every file so a
//     dangerous pattern in any changed OR related file is surfaced (multi-file
//     coverage the single-buffer Layer 1 misses).
//  2. IDOR / broken-authorization heuristic — when the corpus collectively
//     contains a user-supplied object-id SOURCE and an object-access SINK but NO
//     authorization check anywhere, it emits a cross-file candidate finding.
//     This targets the classic multi-file IDOR / auth-bypass shape.
func CrossFileScan(files map[string]string) []GuardianFinding {
	var findings []GuardianFinding

	// (1) Per-file pattern coverage.
	for path, content := range files {
		for _, f := range ScanBuffer(content) {
			f.Match = path + ": " + f.Match
			findings = append(findings, f)
		}
	}

	// (2) Cross-file IDOR / broken-authorization heuristic.
	hasSource := false
	hasSink := false
	hasAuthz := false
	for _, content := range files {
		if !hasSource {
			hasSource = anyMatch(crossFileIDSourceMarkers, content)
		}
		if !hasSink {
			hasSink = anyMatch(crossFileObjectSinkMarkers, content)
		}
		if !hasAuthz {
			hasAuthz = anyMatch(crossFileAuthzMarkers, content)
		}
	}
	if hasSource && hasSink && !hasAuthz {
		findings = append(findings, GuardianFinding{
			Class:    "cross-file-idor",
			Severity: SevHigh,
			Message:  "User-supplied id flows to an object-access sink with no authorization check across the changed files (possible IDOR / broken authorization)",
			Line:     0,
			Match:    "cross-file data-flow",
		})
	}
	return findings
}

// anyMatch reports whether any of the regexes matches content.
func anyMatch(pats []*regexp.Regexp, content string) bool {
	for _, p := range pats {
		if p.MatchString(content) {
			return true
		}
	}
	return false
}
