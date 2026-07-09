// Package template — rules-scope internal-date provenance guard.
//
// SPEC-TEMPLATE-RULES-CLEANUP-001 M1 deliverable (guard (c)).
//
// This test enforces the internal-date-provenance class against the template
// .claude/rules/ subtree. It is a SEPARATE test function with its OWN data
// structures — it MUST NOT be added to the default-tier leakClasses slice,
// because TestLeakClassNoDateShaInDefaultTier enforces a tier-ownership
// contract: "date detection is owned by the strict tier (ISOLATION-001), not
// this SPEC" (AC-SBN-018(a)). Adding the date pattern to leakClasses would
// immediately FAIL that contract test.
//
// Pattern: \b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b (ISO work-date strings).
// Sentinel: RULE_DATE_PROVENANCE_LEAK.
//
// False-positive mitigation (design.md §4):
//   1. Line-context exclusion: lines whose trimmed form starts with a metadata
//      prefix (Last Updated:, Version:, Status:, Relocated:, Origin:,
//      Classification:, created:, created_at:, updated:, updated_at:, date:)
//      are skipped — these are legitimate doc-metadata date placements.
//   2. File-level allowlist: NOTICE.md is fully excluded (third-party import
//      provenance dates — Out of Scope per spec.md §C NOTICE.md carve-out).
//
// Ratchet semantics: this guard is NOT an automatic date classifier. It forces
// human judgment (allowlist or remove) whenever a new date enters the rules
// tree. Internal work-dates (2026-05-17 policy annotations, HISTORY entries)
// are violations; legitimate metadata dates (frontmatter examples, footers)
// are excluded by context.
//
// RECURRENCE BACKSTOP: TestRuleDateProvenanceRecurrenceBackstop documents
// the RED→GREEN transition per REQ-TRC-066 / AC-TRC-F5.
package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ruleDatePattern matches ISO work-date strings in template rules files.
var ruleDatePattern = regexp.MustCompile(`\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b`)

// ruleDateExclusionPrefixes lists line prefixes whose dates are legitimate
// documentation metadata, not internal work-date provenance. A line whose
// trimmed form starts with any of these is skipped by the date guard.
//
// Rationale (design.md §4):
//   - Last Updated:/Version:/Status:/Relocated:/Origin:/Classification: —
//     footer metadata convention for rule files.
//   - created:/created_at:/updated:/updated_at:/date: — frontmatter example
//     dates in schema documentation (spec-frontmatter-schema.md anti-example,
//     skill-authoring.md format examples, archived-agent-rejection.md example
//     frontmatter). These absorb the 3 pedagogical date occurrences that the
//     line-context rule covers, per design.md §4 implementation-freedom note.
var ruleDateExclusionPrefixes = []string{
	"Last Updated:",
	"Version:",
	"Status:",
	"Relocated:",
	"Origin:",
	"Classification:",
	"created:",
	"created_at:",
	"updated:",
	"updated_at:",
	"date:",
}

// ruleDateFileExcludes is the file-level full-exclude list. Paths are matched
// by relForAllowlist suffix.
var ruleDateFileExcludes = []string{
	// NOTICE.md carries third-party import provenance dates (2026-04-26 at
	// lines 18, 91). Out of Scope per spec.md §C NOTICE.md carve-out — the
	// date guard explicitly excludes this file rather than relying on scope
	// (NOTICE.md is INSIDE the rules tree, so scope alone would flag it).
	".claude/rules/moai/NOTICE.md",
}

// isRuleDateLineExcluded reports whether a line's date match should be skipped
// because the line carries legitimate metadata (footer or frontmatter example).
// Leading list markers (-, *, •) are stripped before the prefix check so that
// documentation lines like "- updated: ISO date (e.g., 2026-01-28)" are also
// excluded.
func isRuleDateLineExcluded(line string) bool {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimLeft(trimmed, "-*• ")
	for _, prefix := range ruleDateExclusionPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// isRuleDateFileExcluded reports whether an entire file is excluded from the
// date guard (full-file allowlist — NOTICE.md).
func isRuleDateFileExcluded(relForAllowlist string) bool {
	for _, suffix := range ruleDateFileExcludes {
		if relForAllowlist == suffix {
			return true
		}
	}
	return false
}

// TestRuleDateProvenance scans template .claude/rules/ files for internal
// work-date provenance strings and reports non-excluded matches.
//
// Failure output format (per REQ-TRC-064 / AC-TRC-F4):
//
//	<path> | sentinel=RULE_DATE_PROVENANCE_LEAK | line=<N> | match=<date>
//
// RED→GREEN contract (REQ-TRC-065): at M1 (pre-cleanup) this test FAILS
// naming known work-date violations; at M7 (post-cleanup) it PASSES. Push
// is gated on GREEN.
func TestRuleDateProvenance(t *testing.T) {
	root := templatesRoot
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("template root %q not found: %v", root, err)
	}

	var violations []string

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !shouldScanForLeak(path) {
			return nil
		}

		rel := filepath.ToSlash(path)
		relForAllowlist := strings.TrimPrefix(rel, root+"/")

		// Scope gate: only .claude/rules/ files.
		if !strings.HasPrefix(relForAllowlist, ruleScopedPrefix) {
			return nil
		}
		// File-level exclusion (NOTICE.md).
		if isRuleDateFileExcluded(relForAllowlist) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		lines := strings.Split(string(content), "\n")
		for lineIdx, line := range lines {
			lineNum := lineIdx + 1
			if !ruleDatePattern.MatchString(line) {
				continue
			}
			// Line-context exclusion (footer/frontmatter metadata).
			if isRuleDateLineExcluded(line) {
				continue
			}
			matches := ruleDatePattern.FindAllString(line, -1)
			seen := map[string]struct{}{}
			for _, m := range matches {
				if _, ok := seen[m]; ok {
					continue
				}
				seen[m] = struct{}{}
				violations = append(violations, relForAllowlist+
					" | sentinel=RULE_DATE_PROVENANCE_LEAK"+
					" | line="+itoa(lineNum)+
					" | match="+m)
			}
		}
		return nil
	})

	if walkErr != nil {
		t.Fatalf("filepath.WalkDir error: %v", walkErr)
	}

	if len(violations) > 0 {
		t.Errorf("rules date-provenance leak detected (%d occurrences):", len(violations))
		limit := 40
		if len(violations) < limit {
			limit = len(violations)
		}
		for i := 0; i < limit; i++ {
			t.Errorf("  [%d] %s", i+1, violations[i])
		}
		if len(violations) > limit {
			t.Errorf("  ... %d more (capped)", len(violations)-limit)
		}
		t.Errorf("Remediation: remove internal work-date provenance strings from " +
			"template .claude/rules/ files. Legitimate metadata dates (footers, " +
			"frontmatter examples) are excluded by line-context. See " +
			"SPEC-TEMPLATE-RULES-CLEANUP-001 design.md §4.")
	}
}

// TestRuleDateProvenanceRecurrenceBackstop enforces REQ-TRC-066 / AC-TRC-F5:
// the date guard MUST fire on a synthetic internal-work-date re-leak and MUST
// NOT fire on a clean replacement or a metadata-excluded line.
func TestRuleDateProvenanceRecurrenceBackstop(t *testing.T) {
	t.Parallel()

	// (a) Re-leak: inline work-date in prose — MUST be detected.
	leakyLine := "Per user policy 2026-05-17, L2 worktree is opt-in."
	if !ruleDatePattern.MatchString(leakyLine) {
		t.Fatalf("AC-TRC-F5: date pattern failed to match re-leak line: %q", leakyLine)
	}
	if isRuleDateLineExcluded(leakyLine) {
		t.Errorf("AC-TRC-F5: re-leak line falsely excluded by metadata prefix: %q", leakyLine)
	}

	// (b) Clean replacement: no date — MUST NOT match.
	cleanLine := "Per user policy, L2 worktree is opt-in by default."
	if ruleDatePattern.MatchString(cleanLine) {
		t.Errorf("AC-TRC-F5: date pattern false-positives on clean line: %q", cleanLine)
	}

	// (c) Metadata-excluded line: date in footer context — MUST be excluded.
	metadataLines := []string{
		"Last Updated: 2026-07-15",
		"Version: 2.0.0",
		"Relocated: 2026-06-01",
		"created_at: 2026-05-16   # WRONG — use created:",
		`updated: "2026-01-28"`,
		`- updated: ISO date as string (e.g., "2026-07-15")`,
	}
	for _, ml := range metadataLines {
		if !isRuleDateLineExcluded(ml) {
			t.Errorf("AC-TRC-F5: metadata line NOT excluded by prefix check: %q", ml)
		}
	}

	// (d) NOTICE.md file-level exclusion.
	if !isRuleDateFileExcluded(".claude/rules/moai/NOTICE.md") {
		t.Errorf("AC-TRC-F5: NOTICE.md NOT in file-level exclude list")
	}
	if isRuleDateFileExcluded(".claude/rules/moai/core/askuser-protocol.md") {
		t.Errorf("AC-TRC-F5: non-NOTICE file falsely excluded by file-level list")
	}
}

// TestRuleDateProvenanceNotInDefaultTier is a belt-and-suspenders guard
// verifying the date pattern is NOT present in the default-tier leakClasses
// slice — protecting the tier-ownership contract (AC-TRC-F3). This complements
// the existing TestLeakClassNoDateShaInDefaultTier.
func TestRuleDateProvenanceNotInDefaultTier(t *testing.T) {
	t.Parallel()

	dateProbe := "2026-07-15"
	for _, c := range leakClasses {
		if c.pattern.MatchString(dateProbe) {
			t.Errorf("date pattern erroneously present in default-tier leakClass %q "+
				"— date detection must stay in this separate test file, not leakClasses",
				c.name)
		}
	}
}
