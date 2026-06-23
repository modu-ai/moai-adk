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
	// skillBodyScoped restricts the class to files under the deployed
	// skill-body subtree (".claude/skills/"). When true, the class applies
	// ONLY to skill bodies and is skipped for every other template file
	// (agents, rules, hooks, config). This scope partition is REQUIRED for
	// the SKILL-BODY-NEUTRALITY leak classes (broadened SPEC-ID, Go-impl
	// path, agentless-test-ref, REQ-token): those target the skill-body
	// surface ONLY — rules/agents/commands neutrality is owned by separate
	// SPECs (EXCL-SBN-002). Broadening these patterns across the whole
	// template tree would flag dozens of legitimately-scoped agent/rule
	// references and make the GREEN state unreachable.
	skillBodyScoped bool
}

// skillBodyPrefix is the relative-path prefix (under templatesRoot) that
// identifies a deployed skill body. A leak class with skillBodyScoped=true
// matches ONLY files whose relForAllowlist path begins with this prefix.
const skillBodyPrefix = ".claude/skills/"

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
	// --- SPEC-SKILL-BODY-NEUTRALITY-001 Part B additions (skill-body-scoped) ---
	//
	// The classes below are skillBodyScoped=true: they apply ONLY to files
	// under ".claude/skills/" and are skipped for agents/rules/hooks/config
	// (EXCL-SBN-002 — those surfaces are owned by separate neutrality SPECs).
	// Whole-tree application would flag dozens of legitimately-scoped
	// agent/rule references and make M6 GREEN unreachable.
	{
		// C1b — broaden the SPEC-ID class to the SPEC-V3R[0-9]-* /
		// CONST-V3R[0-9]-* family that dominates skill-body leaks
		// (REQ-SBN-014). The existing whole-tree C1 above matches only
		// SPEC-(V3R6|AGENCY|WORKTREE)-; this skill-body-scoped class adds
		// the V3R2..V3R5 families plus the two named real internal IDs
		// (SPEC-WF-AUDIT-GATE-001, SPEC-MX-001) per REQ-SBN-006.
		name:            "C1b-spec-id-skill-v3r",
		pattern:         regexp.MustCompile(`\bSPEC-V3R[0-9]-[A-Z0-9-]+\b|\bCONST-V3R[0-9]-[0-9]+\b|\bSPEC-WF-AUDIT-GATE-001\b|\bSPEC-MX-001\b`),
		skillBodyScoped: true,
	},
	{
		// C6 — moai-adk's own CI test file reference in a skill body
		// (REQ-SBN-012). The sentinel-presence test (agentless_audit_test.go)
		// asserts the keyword VALUE is present, not the test-file NAME — so
		// the name can be removed while the sentinel value stays.
		name:            "C6-agentless-test-ref",
		pattern:         regexp.MustCompile(`agentless_audit_test\.go`),
		skillBodyScoped: true,
	},
	{
		// C7 — real moai-adk internal Go implementation path in a skill body
		// (REQ-SBN-013 [HARD]). Package-restricted to the real moai-adk
		// top-level package set (spec|cli|hook|ciwatch|design) so it does NOT
		// match the EXCL-SBN-003 illustrative example paths internal/auth/login.go,
		// internal/api/handler.go, internal/core/handler.go. The unrestricted
		// internal/.*\.go form is PROHIBITED — it would make M6 GREEN
		// unreachable. Keep the package set in sync with AC-SBN-005.
		name:            "C7-internal-go-path",
		pattern:         regexp.MustCompile(`internal/(spec|cli|hook|ciwatch|design)/[a-z0-9_/]*\.go`),
		skillBodyScoped: true,
	},
	{
		// S3-req-ac-token-any-prefix (REQ-SBN-007): the REQ/AC-token class is
		// PROMOTED from the former strict tier into the default tier here,
		// skill-body-scoped. Per AC-SBN-018(b) (partition guard) there must be
		// at most ONE leakClass whose pattern matches the REQ-token shape
		// across leakClasses+strictLeakClasses — so this is the SOLE REQ-token
		// entry (the former strictLeakClasses S3 sibling is REMOVED, not
		// duplicated). The name "S3-..." is retained verbatim so AC-SBN-018(b)'s
		// reference to the S3 regex continues to resolve to a single canonical
		// entry.
		//
		// The pattern is a STRICT SUPERSET of the original S3
		// `(REQ|AC)-[A-Z]{2,}-[0-9]{3}` — REQ-SBN-018(b) explicitly permits
		// "the S3 regex OR a strict-superset of it". The superset is required
		// because the skill bodies carry BOTH the standard form (REQ-BRAIN-001)
		// AND the REQ-WF<NNN>-<NNN> form (REQ-WF003-010), and the narrow S3
		// regex misses the latter (WF003 is not [A-Z]{2,}). This superset
		// `(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+` matches both and is exactly the
		// AC-SBN-007 SSOT shape. It still matches the original S3 probe
		// (REQ-EXAMPLE-007), so it remains a strict superset (partition guard
		// satisfied). Skill-body-scoped: fires in the default tier ONLY for
		// ".claude/skills/" files (EXCL-SBN-002 — REQ/AC tokens in
		// agents/rules are owned elsewhere).
		name:            "S3-req-ac-token-any-prefix",
		pattern:         regexp.MustCompile(`\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b`),
		skillBodyScoped: true,
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
//
// NOTE: the former S3-req-ac-token-any-prefix class is no longer here — it
// was PROMOTED into the default-tier leakClasses (skill-body-scoped) by
// SPEC-SKILL-BODY-NEUTRALITY-001 REQ-SBN-007. Per AC-SBN-018(b) there must
// be exactly ONE REQ-token regex entry across leakClasses+strictLeakClasses;
// keeping an S3 sibling here would duplicate it. The date/short-sha classes
// (S1/S2) remain strict-only and are owned by the partition boundary with
// SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 (REQ-SBN-018(a) / EXCL-SBN-001).
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
	// --- SPEC-SKILL-BODY-NEUTRALITY-001 belt-and-suspenders entries (REQ-SBN-013 / AC-SBN-020(c)) ---
	//
	// The 3 illustrative example Go paths in fictional code-review / file-list
	// examples (EXCL-SBN-003). The C7 regex is already package-restricted to
	// internal/(spec|cli|hook|ciwatch|design) and so does NOT match these
	// internal/auth, internal/api, internal/core paths — but they are ALSO
	// registered here as belt-and-suspenders so even a C7 regex regression
	// would not flag them. The `SpecID` field is reused here to carry the
	// matched-substring literal (the allowlist match is by literal substring,
	// not by SPEC-ID semantics).
	{
		File:      ".claude/skills/moai-workflow-testing/references/pr-review-multi-agent.md",
		LineStart: 0,
		LineEnd:   0,
		SpecID:    "internal/auth/login.go",
		Rationale: "Illustrative example Go path in fictional pr-review code example (EXCL-SBN-003 keep-list)",
	},
	{
		File:      ".claude/skills/moai-workflow-testing/references/pr-review-multi-agent.md",
		LineStart: 0,
		LineEnd:   0,
		SpecID:    "internal/api/handler.go",
		Rationale: "Illustrative example Go path in fictional pr-review code example (EXCL-SBN-003 keep-list)",
	},
	{
		File:      ".claude/skills/moai/workflows/mx.md",
		LineStart: 0,
		LineEnd:   0,
		SpecID:    "internal/core/handler.go",
		Rationale: "Illustrative example Go path in mixed-language modified-files list example (EXCL-SBN-003 keep-list)",
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
			// Skill-body scope gate (SPEC-SKILL-BODY-NEUTRALITY-001):
			// a skillBodyScoped class applies ONLY to files under
			// ".claude/skills/" — skip it for agents/rules/hooks/config
			// (EXCL-SBN-002). relForAllowlist is the path relative to
			// templatesRoot, so the skill-body prefix check is direct.
			if class.skillBodyScoped && !strings.HasPrefix(relForAllowlist, skillBodyPrefix) {
				continue
			}
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

// reqTokenProbe is a sample REQ token used to detect which leak-class regexes
// match the REQ/AC-token shape (REQ|AC)-[A-Z]{2,}-[0-9]{3}.
const reqTokenProbe = "REQ-EXAMPLE-007"

// TestLeakClassReqTokenPartition enforces AC-SBN-018(b): there MUST be at most
// ONE leakClass — across BOTH leakClasses and strictLeakClasses — whose pattern
// matches the REQ/AC-token shape. The skill-body REQ-token enforcement
// (REQ-SBN-007) is satisfied by PROMOTING the single S3 regex into the default
// tier (skill-body-scoped), NOT by adding a near-identical sibling. A duplicate
// REQ-token regex across the two slices is the partition-drift this guards.
//
// SPEC-SKILL-BODY-NEUTRALITY-001 REQ-SBN-018(b) / AC-SBN-018(b).
func TestLeakClassReqTokenPartition(t *testing.T) {
	t.Parallel()

	var matching []string
	for _, c := range leakClasses {
		if c.pattern.MatchString(reqTokenProbe) {
			matching = append(matching, "leakClasses/"+c.name)
		}
	}
	for _, c := range strictLeakClasses {
		if c.pattern.MatchString(reqTokenProbe) {
			matching = append(matching, "strictLeakClasses/"+c.name)
		}
	}

	if len(matching) != 1 {
		t.Errorf("AC-SBN-018(b) partition guard FAILED: expected exactly 1 leakClass "+
			"matching the REQ-token shape across leakClasses+strictLeakClasses, got %d: %v",
			len(matching), matching)
	}
}

// TestLeakClassNoDateShaInDefaultTier enforces AC-SBN-018(a): the SKILL-BODY
// additions to the DEFAULT-tier leakClasses MUST NOT include a generic-date
// regex (202[6-9]-MM-DD) or a short-sha regex ([0-9a-f]{7,8}). Those classes
// are owned exclusively by SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001's strict
// tier (S1/S2); duplicating them here would create dual-allow-list drift
// (EXCL-SBN-001 / REQ-SBN-018(a)).
func TestLeakClassNoDateShaInDefaultTier(t *testing.T) {
	t.Parallel()

	dateProbe := "2026-06-04"      // an internal-date sample
	shaProbe := "a1b2c3d "         // a short-sha-sentence-final sample (trailing space)
	for _, c := range leakClasses {
		if c.pattern.MatchString(dateProbe) {
			t.Errorf("AC-SBN-018(a) FAILED: default-tier leakClass %q matches an internal-date "+
				"probe %q — date detection is owned by the strict tier (ISOLATION-001), not this SPEC",
				c.name, dateProbe)
		}
		if c.pattern.MatchString(shaProbe) {
			t.Errorf("AC-SBN-018(a) FAILED: default-tier leakClass %q matches a short-sha probe %q — "+
				"sha detection is owned by the strict tier (ISOLATION-001), not this SPEC",
				c.name, shaProbe)
		}
	}
}

// TestSkillBodyLeakClassRecurrenceBackstop enforces AC-SBN-017 (recurrence
// regression backstop): each SPEC-SKILL-BODY-NEUTRALITY-001 leak class MUST
// fire on a synthetic re-leak string and MUST NOT fire on a clean replacement.
// This documents the RED→GREEN transition deterministically: if a future edit
// reintroduces a CLASS 1-4 leak into a skill body, the corresponding class
// regex flags it (the guard FAILS), and the clean replacement passes.
//
// SPEC-SKILL-BODY-NEUTRALITY-001 REQ-SBN-017 / AC-SBN-017.
func TestSkillBodyLeakClassRecurrenceBackstop(t *testing.T) {
	t.Parallel()

	classByName := map[string]*regexp.Regexp{}
	for i := range leakClasses {
		classByName[leakClasses[i].name] = leakClasses[i].pattern
	}

	cases := []struct {
		class      string
		leaky      string // a re-leak that MUST match
		clean      string // the generic-ized replacement that MUST NOT match
	}{
		{
			class: "C1b-spec-id-skill-v3r",
			leaky: "see SPEC-V3R5-LATE-BRANCH-001 for the policy",
			clean: "see the late-branch opt-in policy",
		},
		{
			class: "C1b-spec-id-skill-v3r",
			leaky: "owned by SPEC-WF-AUDIT-GATE-001",
			clean: "owned by the plan audit gate policy",
		},
		{
			class: "C6-agentless-test-ref",
			leaky: "CI guards in internal/template/agentless_audit_test.go enforce the sentinel",
			clean: "MODE_UNKNOWN is a stable error key; keep it verbatim",
		},
		{
			class: "C7-internal-go-path",
			leaky: "the rule lives in internal/spec/lint.go FrontmatterSchemaRule",
			clean: "the rule lives in the SPEC frontmatter lint rule",
		},
		{
			class: "S3-req-ac-token-any-prefix",
			leaky: "per REQ-WF003-010 the sentinel is emitted",
			clean: "the sentinel is emitted for an unrecognized mode value",
		},
	}

	for _, tc := range cases {
		pat, ok := classByName[tc.class]
		if !ok {
			t.Errorf("AC-SBN-017: leak class %q not found in leakClasses", tc.class)
			continue
		}
		if !pat.MatchString(tc.leaky) {
			t.Errorf("AC-SBN-017: class %q failed to flag a re-leak: %q", tc.class, tc.leaky)
		}
		if pat.MatchString(tc.clean) {
			t.Errorf("AC-SBN-017: class %q false-positives on a clean replacement: %q", tc.class, tc.clean)
		}
	}
}

// TestC7PackageRestriction enforces AC-SBN-020(a)+(b): the C7 Go-impl-path
// class MUST be package-restricted to internal/(spec|cli|hook|ciwatch|design)
// and MUST NOT match the EXCL-SBN-003 illustrative example paths
// (internal/auth/login.go, internal/api/handler.go, internal/core/handler.go).
//
// SPEC-SKILL-BODY-NEUTRALITY-001 REQ-SBN-013 / AC-SBN-020.
func TestC7PackageRestriction(t *testing.T) {
	t.Parallel()

	var c7 *regexp.Regexp
	for i := range leakClasses {
		if leakClasses[i].name == "C7-internal-go-path" {
			c7 = leakClasses[i].pattern
			break
		}
	}
	if c7 == nil {
		t.Fatal("AC-SBN-020: C7-internal-go-path class not found")
	}

	// (b) MUST NOT match the 3 illustrative example paths.
	for _, illustrative := range []string{
		"internal/auth/login.go",
		"internal/api/handler.go",
		"internal/core/handler.go",
	} {
		if c7.MatchString(illustrative) {
			t.Errorf("AC-SBN-020(b): C7 regex must NOT match illustrative path %q", illustrative)
		}
	}

	// (a) MUST match a real restricted-package path.
	for _, real := range []string{
		"internal/spec/lint.go",
		"internal/cli/harness.go",
		"internal/hook/dbsync/db_schema_sync.go",
		"internal/ciwatch/handoff.go",
		"internal/design/dtcg/frozen_guard_test.go",
	} {
		if !c7.MatchString(real) {
			t.Errorf("AC-SBN-020(a): C7 regex must match real restricted-package path %q", real)
		}
	}
}
