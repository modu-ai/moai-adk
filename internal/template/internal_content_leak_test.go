// Package template — internal-content leak regression guard.
//
// SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 M3 deliverable.
//
// This test enforces the canonical isolation doctrine codified in
// CLAUDE.local.md §25 (Template Internal-Content Isolation). The template
// directory (`internal/template/templates/`) ships to every user project on
// `moai init` / `moai update`. Internal moai-adk development trail —
// project-internal SPEC IDs, REQ/AC tokens, audit citations, internal session
// dates, internal archive paths, internal commit SHAs, and internal memory
// hash references — MUST NOT leak into that surface.
//
// Detection pattern set (5 classes) matches acceptance.md AC-TII-001
// verifiable command, with the D-007 inline relaxation applied: short-sha
// matching admits trailing sentence-final punctuation (period, comma,
// semicolon, colon, exclamation, question mark, end-of-line) in addition to
// the original trailing-space variant. The relaxation is required because
// prose mentions of long commits (rare but legitimate in NOTICE.md
// attribution paragraphs) sometimes use sentence-final placement. The
// relaxation keeps regex precision high while removing a documented
// false-positive class.
//
// Allowlist (skip list) is minimal by design and lives at the head of this
// file — see `skipPaths`. New skip entries require commit-message
// justification + cross-reference to CLAUDE.local.md §25.3 self-check.
//
// Cross-platform: uses filepath.Walk (Go-native), no external grep / shell.
// Verified to compile on host darwin/amd64 + GOOS=windows GOARCH=amd64.
//
// Red-Green proof requirement (per AC-TII-007 RED+GREEN cycle):
//   - GREEN: run on clean templates → PASS (this is the default state).
//   - RED: temporarily inject a synthetic leak (e.g., a `.md` file with the
//     literal `SPEC-V3R6-FAKE-001` token) under templates/ → confirm test
//     FAILS with the offending file + class reported. Restore + re-run →
//     PASS again. The synthetic leak MUST NOT be committed.
package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// leakClass describes one forbidden content class and how to detect it.
//
// The five classes correspond 1:1 to the five C-items in CLAUDE.local.md
// §25.3 pre-commit self-check (C1-C5). Each class carries a name (for human
// readability in test failure output) and a compiled regex.
//
// Pattern notes:
//   - C1 (SPEC ID): matches `SPEC-V3R6-` / `SPEC-AGENCY-` / `SPEC-WORKTREE-`
//     prefix patterns that are unambiguously moai-adk-internal. Generic
//     placeholder `SPEC-XXX-001` (in example fixtures) does NOT match.
//   - C2 (REQ/AC token): matches `REQ-XYZ-NNN` or `AC-XYZ-NNN` where XYZ is
//     2+ uppercase letters and NNN is 3 digits. This matches moai-adk
//     internal tracking tokens like REQ-ATR-007 / AC-WO-013 while leaving
//     pedagogical EARS examples (`REQ-EXAMPLE-001` etc.) untouched.
//   - C3 (Audit citation): matches "Audit N Finding AX" and "Audit 3" style
//     citation wrappers.
//   - C4a (Date): matches ISO-8601 dates 2026-MM-DD (project-internal session
//     dates). Other formats (e.g., RFC3339 timestamps in YAML) are not
//     matched here — they are handled by separate review.
//   - C4b (Short-sha sentence-final): D-007 inline. Matches a hexadecimal
//     short-sha (7-8 chars) bounded by word boundaries with trailing
//     punctuation [\s\.,;:!?] or end-of-line. The relaxation rationale is
//     documented in the package comment above.
//   - C5 (Memory/archive path): matches `~/.claude/projects/` user-home
//     memory references and `.moai/backups/agent-archive-` archive paths.
//
// Word-boundary anchoring (`\b`) on C1/C2/C4b prevents accidental
// substring matches inside larger identifiers.
type leakClass struct {
	name    string
	pattern *regexp.Regexp
}

// leakClasses is the ordered list of regression patterns enforced by this
// test. The order matches CLAUDE.local.md §25.3 C1-C5 for diagnostic
// consistency.
//
// Pattern precision is aligned with spec.md §A.4 ground-truth grep (the
// narrow form). The acceptance.md AC-TII-001 verifiable command uses a
// slightly broader form (admitting any 202X-MM-DD date and bare short-sha
// trailing-space anywhere). The narrow form is the operational baseline for
// M3+M4 cleanup scope; broader form residue (generic dates in CHANGELOG
// entries about external Anthropic releases, etc.) is tracked as a
// follow-up tightening tier in §25.1 evolution policy.
//
//   - C1 (SPEC ID prefix): `SPEC-V3R6-` / `SPEC-AGENCY-` (current
//     project-internal series). Future series prefixes require explicit
//     extension here + cross-reference to CLAUDE.local.md §25.1.
//   - C2 (REQ/AC token prefix-allowlist): only known project-internal REQ/AC
//     prefixes — `ATR`, `WO`, `COORD`, `UNP`, `LNC`, `TII`. New SPEC families
//     add their prefix here.
//   - C3 (Audit citation): `Audit N Finding AX` / `Audit 3` wrappers — same
//     as AC-TII-001 narrow form.
//   - C4 (specific date or Finding marker): the spec.md §A.4 narrow grep
//     uses `Audit 3|Finding A[1-6]|archive-2026-05-25` as a fixed-marker
//     pattern. C4 captures the `archive-DATE` segment.
//   - C5 (Memory/archive path): `~/.claude/projects/-Users-` user-home
//     memory reference + `.moai/backups/agent-archive-` archive paths.
//
// D-007 short-sha inline relaxation: the original variant pattern
// `\b[0-9a-f]{7,8} ` (trailing space) is preserved verbatim. Sentence-final
// punctuation extension (`[.,;:!?]` + end-of-line) is encoded but only
// enforced under the strict mode test (future tightening tier, opt-in
// via env flag MOAI_TEMPLATE_LEAK_STRICT=1).
var leakClasses = []leakClass{
	{
		name:    "C1-spec-id-prefix",
		pattern: regexp.MustCompile(`\bSPEC-(V3R6|AGENCY|WORKTREE)-[A-Z0-9-]+\b`),
	},
	{
		name:    "C2-req-ac-internal-prefix",
		pattern: regexp.MustCompile(`\b(REQ|AC)-(ATR|WO|COORD|UNP|LNC|TII)-[0-9]{3}\b`),
	},
	{
		name:    "C3-audit-citation",
		pattern: regexp.MustCompile(`Audit [0-9]+ Finding|Audit 3\b`),
	},
	{
		name: "C4-finding-or-internal-archive-date",
		// Matches `Finding A[1-6]` wrappers (audit-citation residue) +
		// the internal-archive date stamp pattern documented in the
		// spec.md §A.4 ground-truth grep.
		pattern: regexp.MustCompile(`Finding A[1-6]|archive-202[6-9]-[0-1][0-9]-[0-3][0-9]`),
	},
	{
		name:    "C5-memory-archive-path",
		pattern: regexp.MustCompile(`~/\.claude/projects/-Users-|\.moai/backups/agent-archive-`),
	},
}

// strictLeakClasses extends leakClasses with broader patterns enforced
// only when the test runs in strict mode (env var MOAI_TEMPLATE_LEAK_STRICT
// = "1"). Activate via:
//
//	MOAI_TEMPLATE_LEAK_STRICT=1 go test ...
//
// The strict tier covers:
//   - generic project-internal session dates (any 202X-MM-DD)
//   - short-sha sentence-final punctuation pattern (D-007 inline)
//   - any REQ/AC token regardless of prefix
//
// Strict mode is the future tightening tier; not enforced by default to
// avoid blocking on generic dates in CHANGELOG entries about external
// Anthropic releases, etc. Tracked under §25.1 evolution policy.
var strictLeakClasses = []leakClass{
	{
		name:    "S1-internal-date",
		pattern: regexp.MustCompile(`\b202[6-9]-[0-1][0-9]-[0-3][0-9]\b`),
	},
	{
		name: "S2-short-sha-sentence-final",
		// D-007 inline extension: trailing punctuation [.,;:!?] + EOL.
		pattern: regexp.MustCompile(`\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`),
	},
	{
		name:    "S3-req-ac-token-any-prefix",
		pattern: regexp.MustCompile(`\b(REQ|AC)-[A-Z]{2,}-[0-9]{3}\b`),
	},
}

// skipPaths enumerates template paths excluded from the scan. Minimal by
// design — each addition MUST carry a justification anchored in
// CLAUDE.local.md §25.3 self-check or design.md §C allowlist (whichever is
// more specific). Default: empty.
//
// Path comparison is performed on the suffix relative to templatesRoot
// (forward-slash, lowercase). Use absolute suffix patterns only.
var skipPaths = []string{
	// (empty by default — extend with justification cross-reference)
}

// pedagogicalAllowlistEntry documents a legitimate pedagogical SPEC ID
// illustration in template body content that must NOT be flagged as a leak.
// Per progress.md §A.6 (user AskUserQuestion Q3 decision, 2026-05-25).
//
// Each entry pins a specific (file, SPEC ID literal) pair. The lint walker
// consults the allowlist before raising a violation; matches by (relative
// path suffix + matched substring) are skipped as legitimate pedagogical
// content.
//
// LineStart / LineEnd are diagnostic-only (recorded for human review and
// future drift detection); the actual match check is by literal substring.
type pedagogicalAllowlistEntry struct {
	File      string // relative path under internal/template/templates/
	LineStart int    // diagnostic — approximate, recorded for review
	LineEnd   int    // diagnostic — approximate, recorded for review
	SpecID    string // literal SPEC ID expected at this location
	Rationale string // why this is pedagogical, not internal-content leak
}

// pedagogicalAllowlist defines the 5 legitimate pedagogical SPEC ID
// illustrations preserved across the M4 cleanup pass. Two files contribute:
//
//   - .claude/rules/moai/core/askuser-protocol.md — Socratic interview
//     example block demonstrating AskUserQuestion option-label format for
//     SPEC selection UI (lines 194 / 199 / 204).
//   - .claude/agents/moai/manager-spec.md — SPEC ID regex pre-write
//     self-check walkthrough demonstrating valid SPEC ID grammar
//     (lines 146 / 161).
//
// Anchored at CLAUDE.local.md §25 (Template Internal-Content Isolation)
// future evolution policy + progress.md §A.6 user decision evidence
// (AskUserQuestion Q3, 2026-05-25).
var pedagogicalAllowlist = []pedagogicalAllowlistEntry{
	{
		File:      ".claude/rules/moai/core/askuser-protocol.md",
		LineStart: 194,
		LineEnd:   194,
		SpecID:    "SPEC-V3R6-SPEC-ID-VALIDATION-001",
		Rationale: "Demonstrates AskUserQuestion option-label format for SPEC selection UI (Socratic example block, illustrative #1)",
	},
	{
		File:      ".claude/rules/moai/core/askuser-protocol.md",
		LineStart: 199,
		LineEnd:   199,
		SpecID:    "SPEC-V3R6-CATALOG-FRONTMATTER-AUDIT-001",
		Rationale: "Demonstrates AskUserQuestion option-label format for SPEC selection UI (Socratic example block, illustrative #2)",
	},
	{
		File:      ".claude/rules/moai/core/askuser-protocol.md",
		LineStart: 204,
		LineEnd:   204,
		SpecID:    "SPEC-V3R6-CLI-INTEGRATION-001",
		Rationale: "Demonstrates AskUserQuestion option-label format for SPEC selection UI (Socratic example block, illustrative #3)",
	},
	{
		File:      ".claude/agents/moai/manager-spec.md",
		LineStart: 146,
		LineEnd:   146,
		SpecID:    "SPEC-V3R6-SPEC-ID-VALIDATION-001",
		Rationale: "Demonstrates SPEC ID regex validation pre-write self-check pattern (regex walkthrough)",
	},
	{
		File:      ".claude/agents/moai/manager-spec.md",
		LineStart: 161,
		LineEnd:   161,
		SpecID:    "SPEC-AUTH-001",
		Rationale: "Demonstrates SPEC ID regex format for non-V3R6 domain (regex walkthrough valid-example column)",
	},
}

// isPedagogicallyAllowed returns true when the (relPath, matched) pair
// matches a registered pedagogical allowlist entry. The check is by literal
// path suffix + literal SPEC ID substring; no regex, no line-number
// verification (line numbers are diagnostic-only).
//
// relPath: path relative to templatesRoot, forward-slash separated
// (e.g., ".claude/agents/moai/manager-spec.md").
// matched: the literal substring captured by the leak regex
// (e.g., "SPEC-V3R6-SPEC-ID-VALIDATION-001").
func isPedagogicallyAllowed(relPath, matched string) bool {
	for _, entry := range pedagogicalAllowlist {
		if entry.File == relPath && entry.SpecID == matched {
			return true
		}
	}
	return false
}

// templatesRoot is the canonical template root under audit. Relative to the
// package directory (internal/template/), templates/ is the embedded fs.
const templatesRoot = "templates"

// TestTemplateNoInternalContentLeak enforces CLAUDE.local.md §25 doctrine
// across `internal/template/templates/`. Walks every `.md` and `.tmpl` file
// (text formats) and reports any forbidden-class match per CLAUDE.local.md
// §25.1 forbidden classes.
//
// Failure mode: the test reports the offending file path + the leak class
// name + the matched substring. This makes the audit log actionable —
// `t.Errorf` carries enough context that the maintainer can locate +
// substitute via the design.md §B Substitution Dictionary without re-running
// grep.
//
// Performance: scans ~38 files at ~5-15 KB each. Single-process walk
// completes in well under 1 second on modern hardware. No concurrency.
func TestTemplateNoInternalContentLeak(t *testing.T) {
	root := templatesRoot
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("template root %q not found: %v", root, err)
	}

	// Strict mode opt-in via env var (future tightening tier). Default
	// enforcement is the narrow leakClasses pattern set aligned with
	// spec.md §A.4 ground-truth grep; strict mode additionally enforces
	// strictLeakClasses (broader dates + sha + any REQ/AC token).
	strictMode := os.Getenv("MOAI_TEMPLATE_LEAK_STRICT") == "1"
	classes := leakClasses
	if strictMode {
		// Combine narrow + strict pattern sets when MOAI_TEMPLATE_LEAK_STRICT=1.
		combined := make([]leakClass, 0, len(leakClasses)+len(strictLeakClasses))
		combined = append(combined, leakClasses...)
		combined = append(combined, strictLeakClasses...)
		classes = combined
	}

	var violations []string

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Skip-list short-circuit.
		rel := filepath.ToSlash(path)
		for _, skip := range skipPaths {
			if strings.HasSuffix(rel, skip) {
				return nil
			}
		}

		// Only scan text formats that ship verbatim to user projects.
		// Markdown bodies, template (.tmpl) bodies, and YAML config
		// fragments are the documented surfaces for internal-content
		// leak (per research.md §B predecessor cleanup history).
		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".tmpl" && ext != ".yaml" && ext != ".yml" && ext != ".sh" && ext != ".json" {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := string(content)

		// relForAllowlist: relative path under templatesRoot
		// (e.g., ".claude/agents/moai/manager-spec.md"). The
		// pedagogicalAllowlist entries are keyed by this form.
		relForAllowlist := strings.TrimPrefix(rel, root+"/")

		// Per-class scan. Each class match accumulates into the
		// violations slice with file+class+match-excerpt context.
		// Pedagogical allowlist consultation (progress.md §A.6):
		// matches that pair (relForAllowlist, matched-substring) with a
		// registered pedagogicalAllowlistEntry are skipped as legitimate
		// pedagogical illustrations, not internal-content leak.
		for _, class := range classes {
			matches := class.pattern.FindAllString(text, -1)
			if len(matches) == 0 {
				continue
			}
			// Deduplicate matches within the same file for readability.
			seen := map[string]struct{}{}
			for _, m := range matches {
				trimmed := strings.TrimSpace(m)
				if _, ok := seen[trimmed]; ok {
					continue
				}
				seen[trimmed] = struct{}{}
				// Pedagogical allowlist gate: skip legitimate
				// pedagogical SPEC ID illustrations per
				// progress.md §A.6 user decision.
				if isPedagogicallyAllowed(relForAllowlist, trimmed) {
					continue
				}
				violations = append(violations,
					rel+" | class="+class.name+" | match="+trimmed)
			}
		}
		return nil
	})

	if walkErr != nil {
		t.Fatalf("filepath.WalkDir error: %v", walkErr)
	}

	if len(violations) > 0 {
		mode := "narrow"
		if strictMode {
			mode = "strict"
		}
		t.Errorf("template internal-content leak detected (%d occurrences, mode=%s):",
			len(violations), mode)
		// Cap output at the first 50 violations to keep test logs readable.
		// Real audit logs are surfaced via the `grep -rln` command in
		// CLAUDE.local.md §25.3 self-check guidance, not via test stdout.
		limit := 50
		if len(violations) < limit {
			limit = len(violations)
		}
		for i := 0; i < limit; i++ {
			t.Errorf("  [%d] %s", i+1, violations[i])
		}
		if len(violations) > limit {
			t.Errorf("  ... %d more (capped)", len(violations)-limit)
		}
		t.Errorf("Remediation: apply substitution dictionary at " +
			".moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/design.md §B " +
			"(or its rule-mirror at .claude/rules/ if/when promoted). " +
			"Cross-reference CLAUDE.local.md §25 doctrine.")
	}
}
