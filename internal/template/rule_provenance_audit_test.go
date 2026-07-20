// Package template — rules-scope provenance leak guard.
//
// SPEC-TEMPLATE-RULES-CLEANUP-001 M1 deliverable (guards (a), (b), (d)).
//
// This test enforces three rules-scope neutrality classes against the template
// .claude/rules/ subtree that the existing leakClasses/strictLeakClasses
// slices do NOT cover:
//
//   - (a) REQ/AC internal-tracking tokens: \b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b
//     — scoped to .claude/rules/. Pedagogical -XXX- placeholders are excluded
//     post-match (REQ-XXX-* / AC-XXX-*). Sentinel: RULE_REQ_AC_TOKEN_LEAK.
//   - (b) lessons #N / incident-W# provenance: RE2-compatible narrow patterns
//     (4 W-idiom forms + lessons-ref). Sentinel: RULE_PROVENANCE_LEAK.
//   - (d) CONST-V3R* / SPEC-V3R* / MIG-NNN governance tokens: scoped to
//     .claude/rules/. Sentinel: RULE_GOVERNANCE_TOKEN_LEAK.
//
// WHY SEPARATE FROM leakClasses: TestLeakClassReqTokenPartition (in
// internal_content_leak_test.go) enforces exactly ONE REQ-token class across
// leakClasses+strictLeakClasses (the skill-body S3 class). Adding a rules-scope
// sibling to those slices would break that partition guard. The lessons/W# and
// governance-token shapes are new provenance classes not covered by any
// existing leakClass. Defining them here — in a fully independent data
// structure — preserves all three contract tests (Mirror, SanitizedPair,
// DateTier) untouched.
//
// RECURRENCE BACKSTOP: each class carries a self-test (TestRuleProvenance
// RecurrenceBackstop) documenting that a synthetic re-leak fires and a clean
// replacement passes — per REQ-TRC-066 / AC-TRC-F5.
package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ruleScopedPrefix identifies the rules subtree within the template tree.
// All three guards in this file apply ONLY to files under this prefix.
const ruleScopedPrefix = ".claude/rules/"

// ruleProvenanceClass describes one rules-scoped provenance detection pattern.
type ruleProvenanceClass struct {
	name     string
	pattern  *regexp.Regexp
	sentinel string // greppable failure marker
}

// ruleProvenanceClasses defines the three rules-scoped provenance guards.
//
// RE2-compatibility note: Go's regexp package does not support lookahead/
// lookbehind. The W# patterns use narrow idiom enumeration (4 forms) rather
// than a broad \bW[0-9]\b wildcard to avoid false-positives on identifiers
// like SPEC-W3-* or W3C. New W# expression shapes are handled by adding a
// new idiom alternation (documented in the test comment).
var ruleProvenanceClasses = []ruleProvenanceClass{
	{
		// (a) Generalized REQ/AC token. Matches REQ-DOMAIN-NNN / AC-DOMAIN-NNN
		// where DOMAIN is [A-Z][A-Z0-9]* and NNN is [0-9]+. Pedagogical
		// placeholders (REQ-XXX-* / AC-XXX-*) are excluded post-match by
		// checking strings.Contains(match, "-XXX-").
		name:     "REQ-AC-token",
		pattern:  regexp.MustCompile(`\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b`),
		sentinel: "RULE_REQ_AC_TOKEN_LEAK",
	},
	{
		// (b) lessons #N provenance + incident-W# idiom forms.
		//   - \blessons? #[0-9]+ — lessons references (EN/KR prose)
		//   - \bW[0-9] (meta-analysis|meta|fix)\b — W# + meta/fix qualifier
		//   - \bW[0-9]/W[0-9]\b — W# pair (e.g. W1/W2)
		//   - \bW[0-9] 케이스 — KR "W# case" idiom
		//   - \bW[0-9]에서 — KR "in W#" idiom
		// New W# expression shapes: add a new alternation branch here.
		name: "lessons-W-provenance",
		pattern: regexp.MustCompile(
			`\blessons? #[0-9]+` +
				`|\bW[0-9] (meta-analysis|meta|fix)\b` +
				`|\bW[0-9]/W[0-9]\b` +
				`|\bW[0-9] 케이스` +
				`|\bW[0-9]에서`,
		),
		sentinel: "RULE_PROVENANCE_LEAK",
	},
	{
		// (d) Governance tokens: SPEC-V3R*, CONST-V3R*, MIG-NNN.
		// The existing C1b leakClass (skill-body-scoped) uses the same
		// SPEC-V3R/CONST-V3R regex; this guard extends scope to .claude/rules/.
		// MIG-NNN is a standalone migration-ID heading (design.md §5 — included
		// per recommendation; no common English-word collision).
		name: "governance-token",
		pattern: regexp.MustCompile(
			`\bSPEC-V3R[0-9]-[A-Z0-9-]+\b` +
				`|\bCONST-V3R[0-9]-[0-9]+\b` +
				`|\bMIG-[0-9]{3}\b`,
		),
		sentinel: "RULE_GOVERNANCE_TOKEN_LEAK",
	},
}

// governanceTokenFileAllowlist exempts specific rule files from the
// governance-token class ONLY. zone-registry.md IS the constitution zone
// registry — its 115 CONST-V3R* entries (119 token occurrences) are its
// legitimate content, not a leak. File-level (not per-token) because
// enumerating 119 CONST tokens would be brittle; parallel to the C2
// allowListSet in template_neutrality_audit_test.go. Single-path: the guard
// still fires on governance tokens in EVERY OTHER rules file.
var governanceTokenFileAllowlist = map[string]bool{
	".claude/rules/moai/core/zone-registry.md": true,
}

// TestRuleProvenanceAudit scans template .claude/rules/ files for the three
// provenance classes and reports any non-allowlisted match as a violation.
//
// Failure output format (per REQ-TRC-064 / AC-TRC-F4):
//
//	<path> | sentinel=<SENTINEL> | class=<name> | line=<N> | match=<token>
//
// The output includes file path, line number, matched token, and a greppable
// sentinel — making each finding actionable.
//
// RED→GREEN contract (REQ-TRC-065): at M1 (pre-cleanup) this test FAILS
// naming known violations; at M7 (post-cleanup) it PASSES. Push is gated
// on GREEN.
func TestRuleProvenanceAudit(t *testing.T) {
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

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		lines := strings.Split(string(content), "\n")
		for lineIdx, line := range lines {
			lineNum := lineIdx + 1
			for _, class := range ruleProvenanceClasses {
				matches := class.pattern.FindAllString(line, -1)
				if len(matches) == 0 {
					continue
				}
				seen := map[string]struct{}{}
				for _, m := range matches {
					trimmed := strings.TrimSpace(m)
					// (a) Exclude pedagogical -XXX- placeholders.
					if class.name == "REQ-AC-token" && strings.Contains(trimmed, "-XXX-") {
						continue
					}
					// (d) Governance-token gates: file-level exemption first
					// (single-path, zone-registry.md only — its CONST-V3R*
					// entries are legitimate registry content), then the
					// per-(path,token) pedagogical allowlist.
					if class.name == "governance-token" {
						if governanceTokenFileAllowlist[relForAllowlist] {
							continue
						}
						if isPedagogicallyAllowed(relForAllowlist, trimmed) {
							continue
						}
					}
					if _, ok := seen[trimmed]; ok {
						continue
					}
					seen[trimmed] = struct{}{}
					violations = append(violations, relForAllowlist+
						" | sentinel="+class.sentinel+
						" | class="+class.name+
						" | line="+itoa(lineNum)+
						" | match="+trimmed)
				}
			}
		}
		return nil
	})

	if walkErr != nil {
		t.Fatalf("filepath.WalkDir error: %v", walkErr)
	}

	if len(violations) > 0 {
		t.Errorf("rules provenance leak detected (%d occurrences):", len(violations))
		limit := 60
		if len(violations) < limit {
			limit = len(violations)
		}
		for i := 0; i < limit; i++ {
			t.Errorf("  [%d] %s", i+1, violations[i])
		}
		if len(violations) > limit {
			t.Errorf("  ... %d more (capped)", len(violations)-limit)
		}
		t.Errorf("Remediation: remove internal REQ/AC tokens, lessons/W# provenance, " +
			"and CONST/SPEC-V3R/MIG governance tokens from template .claude/rules/ " +
			"files. See SPEC-TEMPLATE-RULES-CLEANUP-001 design.md.")
	}
}

// itoa is a minimal int→string converter to avoid importing strconv for a
// single call site.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// TestRuleProvenanceRecurrenceBackstop enforces REQ-TRC-066 / AC-TRC-F5:
// each provenance class MUST fire on a synthetic re-leak and MUST NOT fire
// on a clean replacement. This documents the RED→GREEN transition
// deterministically: if a future edit reintroduces a cleaned leak shape,
// the corresponding class flags it.
func TestRuleProvenanceRecurrenceBackstop(t *testing.T) {
	t.Parallel()

	classByName := map[string]*regexp.Regexp{}
	for i := range ruleProvenanceClasses {
		classByName[ruleProvenanceClasses[i].name] = ruleProvenanceClasses[i].pattern
	}

	cases := []struct {
		class string
		leaky string // a re-leak that MUST match
		clean string // the generic replacement that MUST NOT match
	}{
		{
			class: "REQ-AC-token",
			leaky: "per REQ-NEW-001 the sentinel is emitted",
			clean: "per the requirement the sentinel is emitted",
		},
		{
			class: "REQ-AC-token",
			leaky: "see AC-XYZ-042 for the criterion",
			clean: "see the acceptance criterion",
		},
		{
			class: "lessons-W-provenance",
			leaky: "per lessons #42 the fix applies",
			clean: "per the established pattern the fix applies",
		},
		{
			class: "lessons-W-provenance",
			leaky: "(W3 meta-analysis: 10 min serial overhead)",
			clean: "(serial verification adds significant overhead)",
		},
		{
			class: "lessons-W-provenance",
			leaky: "W0 fix pattern for build tags",
			clean: "build-tag fix pattern",
		},
		{
			class: "lessons-W-provenance",
			leaky: "W1/W2 chicken-and-egg vs NEW defect",
			clean: "pre-existing baseline vs NEW defect",
		},
		{
			class: "lessons-W-provenance",
			leaky: "cross-platform collision (W3 케이스)",
			clean: "cross-platform collision case",
		},
		{
			class: "governance-token",
			leaky: "per CONST-V3R6-099 the clause is frozen",
			clean: "per the frozen-zone constitution the clause is frozen",
		},
		{
			class: "governance-token",
			leaky: "owned by SPEC-V3R6-EXAMPLE-001",
			clean: "owned by the example policy",
		},
		{
			class: "governance-token",
			leaky: "MIG-003 new loaders were added",
			clean: "the new loaders were added",
		},
	}

	for _, tc := range cases {
		pat, ok := classByName[tc.class]
		if !ok {
			t.Errorf("AC-TRC-F5: provenance class %q not found", tc.class)
			continue
		}
		if !pat.MatchString(tc.leaky) {
			t.Errorf("AC-TRC-F5: class %q failed to flag re-leak: %q", tc.class, tc.leaky)
		}
		if pat.MatchString(tc.clean) {
			t.Errorf("AC-TRC-F5: class %q false-positives on clean replacement: %q", tc.class, tc.clean)
		}
	}

	// Verify -XXX- placeholder exclusion for REQ-AC-token class: the regex
	// structurally matches AC-XXX-001, and the post-match -XXX- check filters
	// it out (not the regex itself).
	xxxProbe := "AC-XXX-001 is a pedagogical placeholder"
	reqAcPat := classByName["REQ-AC-token"]
	matches := reqAcPat.FindAllString(xxxProbe, -1)
	foundXXX := false
	for _, m := range matches {
		if strings.Contains(m, "-XXX-") {
			foundXXX = true
		}
	}
	if !foundXXX {
		t.Errorf("AC-TRC-F5: REQ-AC-token did not capture AC-XXX-001 for post-match exclusion")
	}
}
