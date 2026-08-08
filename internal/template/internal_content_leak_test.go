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
	"fmt"
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
//   - C1 (SPEC ID): matches `SPEC-V3R2-`…`SPEC-V3R6-` / `SPEC-AGENCY-` /
//     `SPEC-WORKTREE-` prefix patterns that are unambiguously moai-adk-internal.
//     Generic placeholder `SPEC-XXX-001` (in example fixtures) does NOT match.
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
	// skillMoaiScoped is a narrower sibling of skillBodyScoped: it restricts
	// the class to files under skillMoaiPrefix (".claude/skills/moai/") only,
	// excluding every sibling skill package. See skillMoaiPrefix doc comment
	// for the false-positive rationale (SPEC-MOAI-SKILL-DOCTRINE-FIX-001
	// REQ-SKF-053). A class MUST NOT set both skillBodyScoped and
	// skillMoaiScoped — skillMoaiScoped alone is the strictly narrower gate.
	skillMoaiScoped bool
	// requireHexLetter, when true, restricts a regex match to strings that
	// contain at least one [a-f] hex letter. The S2 short-sha class sets it
	// so a purely-decimal run (e.g. the 10485760 byte constant in the hook
	// log-rotation) is NOT flagged: a genuine git short-sha almost always
	// carries a hex letter, while a decimal size constant never does.
	requireHexLetter bool
	// dateCarveOut, when true, subjects the class to the date carve-out —
	// the DC-1/DC-4 structural gate (isStructurallyCarvedDate) plus the
	// DC-3/DC-2b/DC-5-PRESERVE content-anchored allowlist (isDateAllowlisted).
	// Only the S1 internal-date class sets it: the carve-out categories are
	// defined over date literals, so applying it to a SPEC-ID or REQ-token
	// class would silently widen the exemption beyond its adjudicated scope.
	dateCarveOut bool
}

// skillBodyPrefix is the relative-path prefix (under templatesRoot) that
// identifies a deployed skill body. A leak class with skillBodyScoped=true
// matches ONLY files whose relForAllowlist path begins with this prefix.
const skillBodyPrefix = ".claude/skills/"

// skillMoaiPrefix is a NARROWER relative-path prefix than skillBodyPrefix:
// it identifies files under the single `.claude/skills/moai/` skill package
// (distinct from sibling packages like `.claude/skills/moai-foundation-core/`
// or `.claude/skills/moai-harness-learner/` — the trailing slash after "moai"
// makes the prefix match unambiguous, since "moai-foundation-core" does not
// begin with "moai/"). A leak class with skillMoaiScoped=true matches ONLY
// files within this package.
//
// SPEC-MOAI-SKILL-DOCTRINE-FIX-001 REQ-SKF-053: this narrower scope is
// REQUIRED (not merely preferred) for the new REQ/AC-short-code and
// C-PH-citation classes below, because the broader skillBodyPrefix scope
// would false-positive on legitimate pedagogical EARS-format examples in
// `.claude/skills/moai-foundation-core/` (e.g. `SPEC-001-REQ-01` teaching
// content) and on the policy-preserved `REQ-HRN-FND-0NN` tokens in
// `.claude/skills/moai-harness-learner/` and `.claude/skills/moai-meta-harness/`
// — none of which are in scope for this SPEC's 42-file `.claude/skills/moai/`
// doctrine-fix target.
const skillMoaiPrefix = ".claude/skills/moai/"

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
//   - C1 (SPEC ID prefix): `SPEC-V3R2-`…`SPEC-V3R6-` / `SPEC-AGENCY-` /
//     `SPEC-WORKTREE-` (current project-internal series, whole-tree). Future
//     series prefixes require explicit extension here + cross-reference to
//     CLAUDE.local.md §25.1.
//   - C2 (REQ/AC token prefix-allowlist): only known project-internal REQ/AC
//     prefixes — `ATR`, `WO`, `COORD`, `UNP`, `LNC`, `TII`, `HRN`, `ORC`. New
//     SPEC families add their prefix here.
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
		pattern: regexp.MustCompile(`\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b`),
	},
	{
		name:    "C2-req-ac-internal-prefix",
		pattern: regexp.MustCompile(`\b(REQ|AC)-(ATR|WO|COORD|UNP|LNC|TII|HRN|ORC)-[0-9]{3}\b`),
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
	// --- SPEC-MOAI-SKILL-DOCTRINE-FIX-001 REQ-SKF-053 additions ---
	//
	// The four classes below extend regex-family coverage to shapes the
	// existing classes structurally cannot match (2-segment REQ/AC short
	// codes, the C-PH-NNN citation shape, non-V3R SPEC-ID prefixes, and
	// 4-segment REQ tokens). The first three of these are skillMoaiScoped
	// (narrower than skillBodyScoped): they fire ONLY under
	// ".claude/skills/moai/", NOT the whole ".claude/skills/" tree, because
	// the broader scope would false-positive on legitimate pedagogical
	// EARS-format examples in moai-foundation-core and on the
	// policy-preserved REQ-HRN-FND tokens in moai-harness-learner /
	// moai-meta-harness (see skillMoaiPrefix doc comment).
	{
		// C1c — non-V3R/AGENCY/WORKTREE SPEC-ID prefixes (REQ-SKF-053c). The
		// whole-tree C1 class above matches only SPEC-(V3R[2-6]|AGENCY|WORKTREE)-;
		// this sibling enumerates known single-domain-family prefixes that
		// escape it (e.g. SPEC-DB-SYNC-RELOC-001, SPEC-PROJECT-DB-HINT-001).
		// Deliberately NARROW (enumerated families, not a generic
		// `SPEC-[A-Z-]+-[0-9]+` wildcard): a generic form would flag dozens
		// of legitimate pedagogical placeholder SPEC IDs used throughout
		// skill bodies (SPEC-BUG-042, SPEC-X-001, SPEC-PAY-001, etc.). New
		// families require an explicit extension here, matching the C1
		// enumeration precedent. Whole-tree (no skill scoping), matching C1.
		name:    "C1c-spec-id-non-v3r-known-families",
		pattern: regexp.MustCompile(`\bSPEC-(DB-SYNC-RELOC|PROJECT-DB-HINT)-[0-9]{3}\b`),
	},
	{
		// C2b — 2-segment REQ/AC short-code tokens with no domain segment
		// (REQ-SKF-053a), e.g. REQ-006 / AC-6. Structurally distinct from
		// the existing C2/S3 classes, which both require an alpha domain
		// segment between the prefix and the trailing number
		// ((REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+) — REQ-006 has no such segment, so
		// digits follow the hyphen directly and neither existing class can
		// match it. skillMoaiScoped: true (see rationale above).
		name:            "C2b-req-ac-2segment",
		pattern:         regexp.MustCompile(`\b(REQ|AC)-[0-9]+\b`),
		skillMoaiScoped: true,
	},
	{
		// C2c — 4-segment REQ-HRN-FND-NNN shape (REQ-SKF-053d). The existing
		// single-segment (REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+ classes structurally
		// cannot match a token with two alpha segments before the trailing
		// number. Enumerated to the HRN-FND family named by the audit
		// finding (matching the C1/C1c enumeration precedent) rather than a
		// generic 2-alpha-segment wildcard.
		name:            "C2c-req-4segment-hrn-fnd",
		pattern:         regexp.MustCompile(`\bREQ-HRN-FND-[0-9]{3}\b`),
		skillMoaiScoped: true,
	},
	{
		// C8 — C-PH-NNN constraint-citation shape (REQ-SKF-053b). No prior
		// class covers this shape.
		name:            "C8-constraint-token-c-ph",
		pattern:         regexp.MustCompile(`\bC-PH-[0-9]{3}\b`),
		skillMoaiScoped: true,
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
		pattern: regexp.MustCompile(`\b202[5-9]-[0-1][0-9]-[0-3][0-9]\b`),
		// The date carve-out (structural gate + content-anchored allowlist)
		// applies to this class only — see the leakClass.dateCarveOut comment.
		dateCarveOut: true,
	},
	{
		name: "S2-short-sha-sentence-final",
		// D-007 inline extension: trailing punctuation [.,;:!?] + EOL.
		pattern: regexp.MustCompile(`\b[0-9a-f]{7,8}([\s\.,;:!?]|$)`),
		// A short-sha run that is ALL decimal digits is a byte/size constant
		// (e.g. 10485760 = 10 MiB in hook log-rotation), not a git sha.
		// Require >=1 [a-f] hex letter so decimal-constant false positives
		// are excluded while genuine shas still match.
		requireHexLetter: true,
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
//   - .claude/rules/moai/core/askuser-protocol-reference.md — Socratic
//     interview example block demonstrating AskUserQuestion option-label
//     format for SPEC selection UI (lines 137 / 142 / 147). The block moved
//     here from askuser-protocol.md when the preview/recommendation detail
//     was split into this path-scoped sidecar to shrink the always-loaded
//     rule footprint; the illustrations themselves are unchanged.
//   - .claude/agents/moai/manager-spec.md — SPEC ID regex pre-write
//     self-check walkthrough demonstrating valid SPEC ID grammar
//     (lines 146 / 161).
//
// Anchored at CLAUDE.local.md §25 (Template Internal-Content Isolation)
// future evolution policy + progress.md §A.6 user decision evidence
// (AskUserQuestion Q3, 2026-05-25).
var pedagogicalAllowlist = []pedagogicalAllowlistEntry{
	{
		File:      ".claude/rules/moai/core/askuser-protocol-reference.md",
		LineStart: 137,
		LineEnd:   137,
		SpecID:    "SPEC-V3R6-SPEC-ID-VALIDATION-001",
		Rationale: "Demonstrates AskUserQuestion option-label format for SPEC selection UI (Socratic example block, illustrative #1)",
	},
	{
		File:      ".claude/rules/moai/core/askuser-protocol-reference.md",
		LineStart: 142,
		LineEnd:   142,
		SpecID:    "SPEC-V3R6-CATALOG-FRONTMATTER-AUDIT-001",
		Rationale: "Demonstrates AskUserQuestion option-label format for SPEC selection UI (Socratic example block, illustrative #2)",
	},
	{
		File:      ".claude/rules/moai/core/askuser-protocol-reference.md",
		LineStart: 147,
		LineEnd:   147,
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
	// --- C1 whole-tree V3R2-5 expansion: mirror-parity-enforced retained tokens ---
	//
	// When C1 broadened to SPEC-V3R[2-6]- (whole-tree), one surface retains a
	// legitimate SPEC-V3R5 token that must NOT be flagged:
	//   spec-workflow.md: byte-parity-enforced with its .claude/ source
	//   (rule_template_mirror_test allowlist). Its internal provenance is
	//   retained on BOTH trees by design; stripping the template mirror
	//   alone would break mirror parity, so the tokens are allowlisted here.
	//   (The former plan-auditor.md SPEC-V3R5-WO-001 regex-example illustration
	//   was genericized to a neutral placeholder; its allowlist entry was
	//   dropped at that time.)
	{
		File:      ".claude/rules/moai/workflow/spec-workflow.md",
		LineStart: 0,
		LineEnd:   0,
		SpecID:    "SPEC-V3R5-LATE-BRANCH-001",
		Rationale: "Mirror-parity-enforced provenance (spec-workflow.md byte-parity with .claude/ source); internal SPEC provenance retained on both trees",
	},
	{
		File:      ".claude/rules/moai/workflow/spec-workflow.md",
		LineStart: 0,
		LineEnd:   0,
		SpecID:    "SPEC-V3R5-WORKFLOW-LEAN-001",
		Rationale: "Mirror-parity-enforced provenance (spec-workflow.md byte-parity with .claude/ source); internal SPEC provenance retained on both trees",
	},
	{
		File:      ".claude/rules/moai/workflow/spec-workflow.md",
		LineStart: 0,
		LineEnd:   0,
		SpecID:    "SPEC-V3R5-WORKFLOW-OPT-001",
		Rationale: "Mirror-parity-enforced provenance (spec-workflow.md byte-parity with .claude/ source); internal SPEC provenance retained on both trees",
	},
	// --- SPEC-MOAI-SKILL-DOCTRINE-FIX-001 REQ-SKF-053(a) new-class allowlist ---
	//
	// The new C2b-req-ac-2segment class (2-segment REQ/AC short codes) flags
	// a legitimate illustrative task-decomposition table in
	// workflows/run/phase-execution.md — a generic placeholder example
	// (task IDs T-001/T-002, file names file1.go/file2.go, requirement
	// codes REQ-001/REQ-002) demonstrating table format, not an internal
	// tracking-token leak.
	{
		File:      ".claude/skills/moai/workflows/run/phase-execution.md",
		LineStart: 324,
		LineEnd:   324,
		SpecID:    "REQ-001",
		Rationale: "Generic placeholder requirement code in an illustrative task-decomposition table example (format demonstration, not a tracked internal REQ)",
	},
	{
		File:      ".claude/skills/moai/workflows/run/phase-execution.md",
		LineStart: 325,
		LineEnd:   325,
		SpecID:    "REQ-002",
		Rationale: "Generic placeholder requirement code in an illustrative task-decomposition table example (format demonstration, not a tracked internal REQ)",
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

// --- Date carve-out: structural gate (DC-1 / DC-4) -------------------------
//
// The date carve-out is HYBRID: a structural gate here for the two
// mechanically-decidable recurring shapes, plus the content-anchored
// dateAllowlist below for the judgement-call categories. A shape earns a
// structural gate only when it is decidable from the line's own syntax (or
// the file's own path) AND is expected to recur in ordinary authoring — so
// that an ordinary edit does not become a Go-allowlist edit.
//
// Both halves are CONTENT-anchored, never line-number-anchored: the gate
// matches a line's own text and a path suffix; the allowlist matches a path
// suffix and a literal date. No line number participates in any decision.
//
// DC-1 — skill/rule frontmatter `updated:` field. Every frontmatter bump
// touches such a line, so an allowlist would levy a Go edit per skill edit.
// The regex mirrors the committed triage classifier's LS-FM rule verbatim
// (`^[[:space:]]+updated:[[:space:]]*"?20`) so the guard and the triage
// record stay mutually auditable.
var dc1FrontmatterUpdatedRe = regexp.MustCompile(`^[ \t]+updated:[ \t]*"?20`)

// dc1FenceToggleRe matches a fenced-code-block delimiter line, mirroring the
// triage classifier's fence tracking. A frontmatter-shaped line INSIDE a fence
// is a pedagogical example, not real frontmatter: it carries no schema-break
// rationale, and the classifier assigns it LS-FM-FENCED -> DC-5 rather than
// DC-1. Tracking fences here keeps the structural gate's reach identical to
// the DC-1 category it implements; the one fenced instance in the tree is an
// explicit dateAllowlist entry instead.
var dc1FenceToggleRe = regexp.MustCompile("^[ \t]*```")

// DC-4 — third-party import-provenance attribution records. In an attribution
// file the import dates ARE the content, and adding an entry is ordinary
// authoring, so the gate must survive a future attribution line.
//
// The gate is deliberately (file AND line shape), not file alone. A whole-file
// exemption would carve out EVERY date in the file, including one that is
// genuinely an internal-development stamp merely sitting in the same document
// — and it would make AC-TDN-015 probe A (a future attribution line must not
// be flagged) and AC-TDN-010's injection probe (an arbitrary dated line in the
// same file MUST be flagged) mutually unsatisfiable. Requiring the attribution
// line shape satisfies both.
var dc4AttributionFiles = []string{
	".claude/rules/moai/NOTICE.md",
}

// dc4AttributionLineRe matches the two attribution-record line shapes used in
// the attribution files: the inline `(imported <date>)` parenthetical and the
// `**Import Date (<source>)**: <date>` summary row.
var dc4AttributionLineRe = regexp.MustCompile(`\(imported 20[0-9]{2}-|^\*\*Import Date\b`)

// isStructurallyCarvedDate reports whether a date match on this line is
// covered by the DC-1 / DC-4 structural gate.
//
// relPath: path relative to templatesRoot, forward-slash separated.
// line: the full text of the line carrying the match (fence state resolved by
// the caller, which tracks it while walking the file).
// inFence: whether the line sits inside a fenced code block.
func isStructurallyCarvedDate(relPath, line string, inFence bool) bool {
	// DC-4 — an attribution-record line inside an attribution file.
	if dc4AttributionLineRe.MatchString(line) {
		for _, f := range dc4AttributionFiles {
			if relPath == f {
				return true
			}
		}
	}
	// DC-1 — real (unfenced) frontmatter `updated:` field.
	return !inFence && dc1FrontmatterUpdatedRe.MatchString(line)
}

// --- Date carve-out: content-anchored allowlist (DC-3 / DC-2b / DC-5) ------

// dateAllowlistEntry pins one (file, date literal) pair that the internal-date
// class must not flag. Every entry is traceable to a PRESERVE row of the
// committed triage record (triage.tsv), and Category records which triage
// category the row was adjudicated into.
//
// Matching is by relative path + literal date substring only. There is
// deliberately NO line field, not even a diagnostic one: the pedagogical
// allowlist's unused LineStart/LineEnd fields have proven to be dead weight
// that invites line-number coupling, and the carve-out is required to stay
// content-anchored.
type dateAllowlistEntry struct {
	File      string // relative path under internal/template/templates/
	Date      string // literal date expected in this file
	Category  string // triage category this row was adjudicated into
	Rationale string // why the date is preserved rather than removed
}

// dateAllowlist holds the judgement-call halves of the carve-out:
//
//   - DC-3  — functional deadline literals (a real future deadline, not a
//     record of when work happened).
//   - DC-2b — mirror-capture stamps on third-party documentation mirrors,
//     where the date is the only staleness signal a reader has.
//   - DC-5  — per-row adjudicated preservations (pedagogical examples,
//     upstream-repository facts, verification-capture stamps).
//
// Each of these is a judgement about the specific content, not a mechanically
// decidable shape, so each stays an explicit entry a reviewer must read.
var dateAllowlist = []dateAllowlistEntry{
	{
		File:      ".claude/skills/moai-foundation-cc/reference/advanced-agent-patterns.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-cli-reference-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-devcontainers-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-discover-plugins-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-headless-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-plugin-marketplaces-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-plugins-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-sandboxing-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-skills-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-statusline-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md",
		Date:      "2026-01-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, the sole staleness signal for a third-party document mirror",
	},
	{
		File:      ".claude/agents/moai/plan-auditor.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai-foundation-core/modules/commands-reference.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai-foundation-core/modules/spec-first-ddd.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai-foundation-core/references/examples.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai-foundation-core/references/reference.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai-foundation-core/SKILL.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai/workflows/plan.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai/workflows/plan/clarity-interview.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/skills/moai/workflows/plan/spec-assembly.md",
		Date:      "2026-11-22",
		Category:  "DC-3",
		Rationale: "functional deadline literal 2026-11-22",
	},
	{
		File:      ".claude/output-styles/moai/moai-learn.md",
		Date:      "2026-04-11",
		Category:  "DC-5",
		Rationale: "illustrative filename in a naming-convention example, carries no internal development state",
	},
	{
		File:      ".claude/rules/moai/development/skill-authoring.md",
		Date:      "2026-01-28",
		Category:  "DC-5",
		Rationale: "frontmatter schema example inside a fenced block; deleting the value would require the placeholder substitution REQ-TDN-007 forbids",
	},
	{
		File:      ".claude/rules/moai/development/spec-frontmatter-schema.md",
		Date:      "2026-05-16",
		Category:  "DC-5",
		Rationale: "deliberate anti-pattern example (snake_case alias); same REQ-TDN-007 rationale",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-plugins-official.md",
		Date:      "2026-07-03",
		Category:  "DC-5",
		Rationale: "mirror-capture annotation on a third-party doc mirror; same staleness-signal rationale as the REQ-TDN-011 DC-2b set",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/claude-code-sub-agents-official.md",
		Date:      "2026-07-03",
		Category:  "DC-5",
		Rationale: "mirror-capture annotation on a third-party doc mirror; same staleness-signal rationale as the REQ-TDN-011 DC-2b set",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/SKILL.md",
		Date:      "2026-07-03",
		Category:  "DC-5",
		Rationale: "official-doc verification-capture stamp; the date is the reader staleness signal for the doc-lag caveat",
	},
	{
		File:      ".claude/skills/moai-meta-harness/SKILL.md",
		Date:      "2026-03-26",
		Category:  "DC-5",
		Rationale: "public upstream repository creation date cited alongside star and fork counts",
	},
	// --- SPEC-TEMPLATE-DATE-NEUTRALITY-002: 2025 date allowlist (24 entries) ---
	// Each entry pins one (file, 2025-date) PRESERVE finding from the committed
	// triage.tsv. Inert until M5 widens the S1-internal-date year class to include 2025.
	{
		File:      ".claude/skills/moai-foundation-cc/reference/best-practices-checklist.md",
		Date:      "2025-11-25",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, sole staleness signal for a third-party document mirror (REQ-TDN2-010)",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/examples.md",
		Date:      "2025-11-26",
		Category:  "DC-5",
		Rationale: "EX-FM frontmatter block shown as a syntax example (REQ-TDN2-011 absolute pin)",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/reference.md",
		Date:      "2025-12-06",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, sole staleness signal for a third-party document mirror (REQ-TDN2-010)",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/skill-examples.md",
		Date:      "2025-11-25",
		Category:  "DC-5",
		Rationale: "EX-FM frontmatter block shown as a syntax example (REQ-TDN2-011 absolute pin)",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md",
		Date:      "2025-11-25",
		Category:  "DC-5",
		Rationale: "EX-FM frontmatter block shown as a syntax example (REQ-TDN2-011 absolute pin)",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md",
		Date:      "2025-12-25",
		Category:  "DC-5",
		Rationale: "DEADLINE forward-looking review or expiry date",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md",
		Date:      "2025-11-25",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, sole staleness signal for a third-party document mirror (REQ-TDN2-010)",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md",
		Date:      "2025-11-25",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, sole staleness signal for a third-party document mirror (REQ-TDN2-010)",
	},
	{
		File:      ".claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-integration-patterns.md",
		Date:      "2025-11-25",
		Category:  "DC-2b",
		Rationale: "mirror-capture stamp, sole staleness signal for a third-party document mirror (REQ-TDN2-010)",
	},
	{
		File:      ".claude/skills/moai-foundation-core/references/reference.md",
		Date:      "2025-11-15",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-foundation-core/references/reference.md",
		Date:      "2025-11-20",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-foundation-core/references/reference.md",
		Date:      "2025-11-25",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-foundation-core/references/reference.md",
		Date:      "2025-11-28",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-foundation-core/references/reference.md",
		Date:      "2025-12-03",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-workflow-project/references/examples.md",
		Date:      "2025-12-06",
		Category:  "DC-5",
		Rationale: "EX-DATA structured-data or code-sample value (REQ-TDN2-011 absolute pin)",
	},
	{
		File:      ".claude/skills/moai-workflow-project/references/reference.md",
		Date:      "2025-11-15",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-workflow-project/references/reference.md",
		Date:      "2025-11-20",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-workflow-project/references/reference.md",
		Date:      "2025-11-27",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-workflow-project/schemas/tab_schema.json",
		Date:      "2025-12-22",
		Category:  "DC-5",
		Rationale: "EX-DATA structured-data or code-sample value (REQ-TDN2-011 absolute pin)",
	},
	{
		File:      ".claude/skills/moai-workflow-spec/references/examples.md",
		Date:      "2025-12-07",
		Category:  "DC-5",
		Rationale: "CREATED documentation-example Created: stamp",
	},
	{
		File:      ".claude/skills/moai-workflow-testing/references/reference.md",
		Date:      "2025-11-15",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-workflow-testing/references/reference.md",
		Date:      "2025-11-20",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-workflow-testing/references/reference.md",
		Date:      "2025-11-25",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
	{
		File:      ".claude/skills/moai-workflow-testing/references/reference.md",
		Date:      "2025-11-30",
		Category:  "DC-5",
		Rationale: "HIST version-history table row; PRESERVE as legitimate release-note documentation",
	},
}

// isDateAllowlisted reports whether the (relPath, matched) pair is a
// registered date-allowlist entry. The check is by literal path suffix +
// literal date substring; no regex, no line-number verification.
//
// relPath: path relative to templatesRoot, forward-slash separated.
// matched: the literal date captured by the internal-date regex.
func isDateAllowlisted(relPath, matched string) bool {
	for _, entry := range dateAllowlist {
		if entry.File == relPath && entry.Date == matched {
			return true
		}
	}
	return false
}

// --- Failure reporting --------------------------------------------------

// leakReportConsoleCap bounds how many violations are printed to the test log.
// A 135-row failure is unreadable in a CI log, so the console output stays
// capped — but the cap never hides findings: the remainder is written in full
// to the file named in the truncation message (REQ-TDN-016).
const leakReportConsoleCap = 50

// leakReportFilePattern is the os.CreateTemp pattern for the full-listing
// file. The file is created under os.TempDir() (so it is cross-platform and
// stays out of the working tree) and is deliberately NOT registered for
// cleanup: the truncation message names its path, so it has to outlive the
// test run for anyone to read it. It is written only when a run actually
// truncates, so a green run leaves nothing behind.
const leakReportFilePattern = "moai-template-leak-*.log"

// writeLeakFullListing writes every violation to a file under os.TempDir()
// and returns its path. Errors are returned rather than fatal: a listing that
// cannot be written must not mask the violations themselves.
func writeLeakFullListing(violations []string, mode string) (string, error) {
	f, err := os.CreateTemp("", leakReportFilePattern)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	var b strings.Builder
	fmt.Fprintf(&b, "template internal-content leak — %d occurrences, mode=%s\n\n",
		len(violations), mode)
	for i, v := range violations {
		fmt.Fprintf(&b, "[%d] %s\n", i+1, v)
	}
	if _, err := f.WriteString(b.String()); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// templatesRoot is the canonical template root under audit. Relative to the
// package directory (internal/template/), templates/ is the embedded fs.
const templatesRoot = "templates"

// leakTextExtensions is the set of file extensions scanned verbatim for
// internal-content leak — the text formats that ship to user projects.
// `.js` is scanned because dynamic-workflow fan-out scripts ship verbatim under
// `.claude/workflows/`. Their header comments routinely cite the SPEC that
// authored them, so without this entry the walker would skip the files entirely
// and every neutrality judgment on them would be vacuous — a green produced by
// not reading the file, not by the file being clean.
var leakTextExtensions = map[string]bool{
	".md":   true,
	".tmpl": true,
	".yaml": true,
	".yml":  true,
	".sh":   true,
	".json": true,
	".js":   true,
}

// leakScannedDotfiles is the basename allowlist of EXTENSIONLESS dotfiles that
// ship verbatim to user projects but carry no scannable extension
// (filepath.Ext(".gitignore") == ".gitignore", which is not in
// leakTextExtensions). Before this allowlist these files escaped the walker
// entirely, so a SPEC-ID leak in .gitignore slipped past the audit.
// Scanning them by basename closes that leak class going forward.
//
// `.gitkeep` is deliberately EXCLUDED: some `.gitkeep` files legitimately carry
// a provenance marker (e.g. the plan-audit reports dir's `.gitkeep` references
// SPEC-WF-AUDIT-GATE-001, and that reference is REQUIRED by
// skills_audit_test.go TestReportsDirGitkeepExists). The leak walker must not
// police `.gitkeep` or it would forbid a guard-required provenance token.
var leakScannedDotfiles = map[string]bool{
	".gitignore":     true,
	".gitattributes": true,
}

// shouldScanForLeak reports whether the file at path is an internal-content
// leak-scan target — either a text format (by extension) OR an extensionless
// dotfile that ships verbatim to user projects (by basename).
func shouldScanForLeak(path string) bool {
	if leakTextExtensions[filepath.Ext(path)] {
		return true
	}
	return leakScannedDotfiles[filepath.Base(path)]
}

// scanLine is one line of a scanned file together with its fenced-code-block
// state. The scan is line-aware (rather than whole-file) because the DC-1
// structural gate is a property of the line a match sits on, which a
// whole-text FindAllString cannot recover.
type scanLine struct {
	text    string
	inFence bool
}

// splitScanLines splits file text into lines and resolves each line's
// fenced-code-block state, mirroring the triage classifier's fence tracking
// (a delimiter line toggles the state and is itself reported as outside).
func splitScanLines(text string) []scanLine {
	raw := strings.Split(text, "\n")
	lines := make([]scanLine, 0, len(raw))
	inFence := false
	for _, l := range raw {
		if dc1FenceToggleRe.MatchString(l) {
			inFence = !inFence
			lines = append(lines, scanLine{text: l, inFence: false})
			continue
		}
		lines = append(lines, scanLine{text: l, inFence: inFence})
	}
	return lines
}

// collectLeakViolations scans text for every applicable leak class and returns
// human-readable violation strings (one per distinct match). The per-class
// gates are applied here: skill-body scope (skillBodyScoped classes apply only
// under ".claude/skills/"), requireHexLetter (S2 decimal-constant exclusion),
// the pedagogical allowlist, and — for the date class only — the DC-1/DC-4
// structural gate plus the DC-3/DC-2b/DC-5 date allowlist.
//
// Deduplication is per file, per class, on the trimmed match text: the same
// literal appearing on several lines yields one violation. This is unchanged
// from the whole-file scan the line-aware walk replaced, and other classes
// depend on it.
//
// relForAllowlist is the templatesRoot-relative, forward-slash path used for the
// skill-body scope check and the allowlist lookups; displayPath is the
// label rendered in each violation string.
func collectLeakViolations(displayPath, relForAllowlist, text string, classes []leakClass) []string {
	var out []string
	lines := splitScanLines(text)
	for _, class := range classes {
		// Skill-body scope gate (SPEC-SKILL-BODY-NEUTRALITY-001):
		// a skillBodyScoped class applies ONLY to files under ".claude/skills/"
		// — skip it for agents/rules/hooks/config (EXCL-SBN-002).
		if class.skillBodyScoped && !strings.HasPrefix(relForAllowlist, skillBodyPrefix) {
			continue
		}
		// Narrower skill-body scope gate (SPEC-MOAI-SKILL-DOCTRINE-FIX-001
		// REQ-SKF-053): a skillMoaiScoped class applies ONLY to files under
		// ".claude/skills/moai/" — skip it for every other skill package.
		if class.skillMoaiScoped && !strings.HasPrefix(relForAllowlist, skillMoaiPrefix) {
			continue
		}
		// Deduplicate matches within the same file for readability.
		seen := map[string]struct{}{}
		for _, ln := range lines {
			for _, m := range class.pattern.FindAllString(ln.text, -1) {
				trimmed := strings.TrimSpace(m)
				// requireHexLetter gate (S2): a match with no [a-f] hex letter is a
				// decimal byte/size constant, not a short-sha.
				if class.requireHexLetter && !strings.ContainsAny(trimmed, "abcdef") {
					continue
				}
				// DC-1/DC-4 structural gate. Evaluated BEFORE the dedup bookkeeping
				// because it is a property of THIS line: recording a structurally
				// carved match as seen would suppress a genuine occurrence of the
				// same literal on a later, uncarved line.
				if class.dateCarveOut && isStructurallyCarvedDate(relForAllowlist, ln.text, ln.inFence) {
					continue
				}
				if _, ok := seen[trimmed]; ok {
					continue
				}
				seen[trimmed] = struct{}{}
				// Pedagogical allowlist gate: skip legitimate pedagogical SPEC ID
				// illustrations per progress.md §A.6 user decision.
				if isPedagogicallyAllowed(relForAllowlist, trimmed) {
					continue
				}
				// DC-3/DC-2b/DC-5-PRESERVE content-anchored date allowlist.
				if class.dateCarveOut && isDateAllowlisted(relForAllowlist, trimmed) {
					continue
				}
				out = append(out, displayPath+" | class="+class.name+" | match="+trimmed)
			}
		}
	}
	return out
}

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

		// Scan text formats + extensionless dotfiles that ship verbatim to
		// user projects. Markdown/tmpl/YAML/sh/json bodies are the documented
		// leak surfaces; .gitignore / .gitattributes carry no scannable
		// extension (filepath.Ext(".gitignore") == ".gitignore") so they are
		// matched by basename — previously they escaped the walker entirely,
		// letting a SPEC-ID leak in .gitignore slip past. (.gitkeep is
		// deliberately NOT scanned — it may carry a guard-required provenance
		// marker; see leakScannedDotfiles.)
		if !shouldScanForLeak(path) {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		// relForAllowlist: relative path under templatesRoot
		// (e.g., ".claude/agents/moai/manager-spec.md"). The
		// pedagogicalAllowlist entries are keyed by this form.
		relForAllowlist := strings.TrimPrefix(rel, root+"/")

		// Per-class scan (skill-body scope + requireHexLetter + pedagogical
		// allowlist gates applied inside collectLeakViolations).
		violations = append(violations,
			collectLeakViolations(rel, relForAllowlist, string(content), classes)...)
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
		// Cap console output to keep test logs readable, but never let the cap
		// hide findings: whatever is truncated is written in full to a file and
		// that path is named in the truncation message (REQ-TDN-016). The
		// former behaviour reported only a residual count and pointed at an
		// indirect `grep -rln` recipe, which is what kept the true finding
		// count invisible.
		shown := min(len(violations), leakReportConsoleCap)
		for i := range shown {
			t.Errorf("  [%d] %s", i+1, violations[i])
		}
		if len(violations) > shown {
			t.Errorf("  ... %d more (capped)", len(violations)-shown)
		}
		// The path accompanies EVERY failure, not only a truncated one. That is
		// a superset of REQ-TDN-016 ("shall not truncate without one") and it
		// is what makes the requirement testable: AC-TDN-010's injection recipe
		// exercises the reporting path with a single synthetic finding, which
		// by construction does not reach the cap.
		if path, err := writeLeakFullListing(violations, mode); err == nil {
			t.Errorf("  full listing: %s", path)
		} else {
			// Never mask the write failure and never panic: the findings above
			// are still the actionable output.
			t.Errorf("  full listing unavailable: %v", err)
		}
		t.Errorf("Remediation: apply substitution dictionary at " +
			".moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/design.md §B " +
			"(or its rule-mirror at .claude/rules/ if/when promoted). " +
			"Cross-reference CLAUDE.local.md §25 doctrine.")
	}
}

// TestTemplateLeakWalkerScansExtensionlessDotfiles enforces that the walker now
// scans extensionless dotfiles that ship verbatim to user projects (.gitignore,
// .gitattributes). Before this, filepath.Ext(".gitignore") returned
// ".gitignore" (not in the text-extension set), so these files were skipped and
// a SPEC-ID leak in them escaped the audit. This test documents the RED→GREEN
// transition: it plants a synthetic internal SPEC-ID in a temp .gitignore-named
// file and asserts the walker's real leak classes flag it; a clean dotfile is
// NOT flagged.
//
// `.gitkeep` is deliberately OUT of scope: a `.gitkeep` may carry a
// guard-required provenance marker (e.g. the plan-audit reports dir's `.gitkeep`
// references SPEC-WF-AUDIT-GATE-001, required by TestReportsDirGitkeepExists),
// so the leak walker must NOT police it.
func TestTemplateLeakWalkerScansExtensionlessDotfiles(t *testing.T) {
	t.Parallel()

	// (a) .gitignore / .gitattributes are in scope; .gitkeep is NOT (it may
	// carry a guard-required provenance marker); ordinary source files are not
	// (they are covered by their own leak-scan surfaces, not this walker).
	for _, base := range []string{".gitignore", ".gitattributes"} {
		if !shouldScanForLeak(filepath.Join("some", "dir", base)) {
			t.Errorf("shouldScanForLeak(%q) = false, want true — extensionless dotfile must be scanned", base)
		}
	}
	if shouldScanForLeak(filepath.Join("some", "dir", ".gitkeep")) {
		t.Errorf("shouldScanForLeak(.gitkeep) = true, want false — .gitkeep may carry a guard-required provenance marker and must not be policed")
	}
	if shouldScanForLeak(filepath.Join("some", "dir", "main.go")) {
		t.Errorf("shouldScanForLeak(main.go) = true, want false — .go is not a leak-scan surface")
	}

	// (b) End-to-end: plant a synthetic internal SPEC-ID (V3R6 series → matched
	// by the whole-tree C1 class) in a temp .gitignore-named file and confirm the
	// walker flags it. This proves extensionless files are now walked + scanned.
	planted := "SPEC-V3R6-DEMO-001"
	leakTree := t.TempDir()
	giPath := filepath.Join(leakTree, ".gitignore")
	if err := os.WriteFile(giPath, []byte("bin/\n# leaked internal token: "+planted+"\n*.log\n"), 0o644); err != nil {
		t.Fatalf("write planted .gitignore: %v", err)
	}

	var flagged []string
	walkErr := filepath.WalkDir(leakTree, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !shouldScanForLeak(path) {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel := filepath.ToSlash(strings.TrimPrefix(path, leakTree+string(os.PathSeparator)))
		flagged = append(flagged, collectLeakViolations(rel, rel, string(content), leakClasses)...)
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk planted tree: %v", walkErr)
	}
	if len(flagged) == 0 {
		t.Errorf("planted internal SPEC-ID %q in a .gitignore was NOT flagged — extensionless-dotfile scanning is not working", planted)
	}

	// (c) Regression backstop: a clean dotfile (no internal token) is NOT flagged.
	cleanTree := t.TempDir()
	cleanPath := filepath.Join(cleanTree, ".gitignore")
	if err := os.WriteFile(cleanPath, []byte("bin/\n*.log\nnode_modules/\n"), 0o644); err != nil {
		t.Fatalf("write clean .gitignore: %v", err)
	}
	cleanContent, _ := os.ReadFile(cleanPath)
	if v := collectLeakViolations(".gitignore", ".gitignore", string(cleanContent), leakClasses); len(v) != 0 {
		t.Errorf("clean .gitignore falsely flagged as leak: %v", v)
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
// regex (202[5-9]-MM-DD) or a short-sha regex ([0-9a-f]{7,8}). Those classes
// are owned exclusively by SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001's strict
// tier (S1/S2); duplicating them here would create dual-allow-list drift
// (EXCL-SBN-001 / REQ-SBN-018(a)).
func TestLeakClassNoDateShaInDefaultTier(t *testing.T) {
	t.Parallel()

	dateProbe := "2026-06-04" // an internal-date sample
	shaProbe := "a1b2c3d "    // a short-sha-sentence-final sample (trailing space)
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
		class string
		leaky string // a re-leak that MUST match
		clean string // the generic-ized replacement that MUST NOT match
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
		"internal/hook/security/scan.go",
		"internal/ciwatch/handoff.go",
	} {
		if !c7.MatchString(real) {
			t.Errorf("AC-SBN-020(a): C7 regex must match real restricted-package path %q", real)
		}
	}
}

// TestReqSkf053NewLeakClassesDetectShapes enforces REQ-SKF-053 end-to-end via
// collectLeakViolations (not just regexp.MatchString): each of the 4 shapes
// the audit found the pre-existing regex families structurally missed MUST
// be flagged when planted in a file under the appropriate scope, and MUST
// NOT be flagged when the same literal appears outside that scope (proving
// the skillMoaiScoped gate, not just the raw pattern, is exercised).
//
// SPEC-MOAI-SKILL-DOCTRINE-FIX-001 REQ-SKF-053 (a)-(d).
func TestReqSkf053NewLeakClassesDetectShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string // sub-test label
		class        string // expected leakClass name that must fire
		text         string // synthetic content containing the shape
		inScopePath  string // relForAllowlist path where the class IS scoped to fire
		outScopePath string // relForAllowlist path where a skillMoaiScoped class must NOT fire (empty = whole-tree class, skip this check)
	}{
		{
			name:         "2-segment REQ short code",
			class:        "C2b-req-ac-2segment",
			text:         "the fix addresses REQ-006 in this milestone",
			inScopePath:  ".claude/skills/moai/workflows/example.md",
			outScopePath: ".claude/skills/moai-foundation-core/modules/example.md",
		},
		{
			name:         "2-segment AC short code",
			class:        "C2b-req-ac-2segment",
			text:         "see AC-6 for the acceptance criterion",
			inScopePath:  ".claude/skills/moai/workflows/example.md",
			outScopePath: ".claude/skills/moai-foundation-core/modules/example.md",
		},
		{
			name:         "C-PH-NNN constraint citation",
			class:        "C8-constraint-token-c-ph",
			text:         "bounded by constraint C-PH-003 per the plan",
			inScopePath:  ".claude/skills/moai/workflows/example.md",
			outScopePath: ".claude/skills/moai-foundation-core/modules/example.md",
		},
		{
			name:         "4-segment REQ-HRN-FND token",
			class:        "C2c-req-4segment-hrn-fnd",
			text:         "preserved verbatim under REQ-HRN-FND-004",
			inScopePath:  ".claude/skills/moai/workflows/example.md",
			outScopePath: ".claude/skills/moai-harness-learner/SKILL.md",
		},
		{
			name:        "non-V3R SPEC-ID known family (whole-tree class)",
			class:       "C1c-spec-id-non-v3r-known-families",
			text:        "migrated per SPEC-DB-SYNC-RELOC-001 and SPEC-PROJECT-DB-HINT-001",
			inScopePath: ".claude/agents/moai/example.md",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inViolations := collectLeakViolations(tc.inScopePath, tc.inScopePath, tc.text, leakClasses)
			if !anyViolationHasClass(inViolations, tc.class) {
				t.Errorf("expected class %q to fire for text %q at path %q, got violations: %v",
					tc.class, tc.text, tc.inScopePath, inViolations)
			}

			if tc.outScopePath == "" {
				return
			}
			outViolations := collectLeakViolations(tc.outScopePath, tc.outScopePath, tc.text, leakClasses)
			if anyViolationHasClass(outViolations, tc.class) {
				t.Errorf("expected class %q NOT to fire outside its scope for text %q at path %q, got violations: %v",
					tc.class, tc.text, tc.outScopePath, outViolations)
			}
		})
	}
}

// anyViolationHasClass reports whether any violation string produced by
// collectLeakViolations carries the given class name (violation strings are
// formatted as "path | class=<name> | match=<substring>").
func anyViolationHasClass(violations []string, class string) bool {
	needle := "class=" + class + " "
	for _, v := range violations {
		if strings.Contains(v, needle) {
			return true
		}
	}
	return false
}

// TestTemplateLearnedWorkflowBlockNeutral pins the MOAI:LEARNED-WORKFLOW managed
// block shipped in templates/CLAUDE.md (the Template-First empty marker). The
// block ships to every user project on `moai init` / `moai update`, so it MUST:
//
//	(a) be PRESENT — the heading + start/end markers exist in the template;
//	(b) ship EMPTY — zero bullets between the markers (a populated block in the
//	    template would leak internal learning data — the AP-HEV2-006 anti-pattern);
//	(c) be NEUTRAL — no forbidden-class content (internal SPEC IDs / REQ-AC
//	    tokens / internal dates / short-shas) in the block region.
//
// This EXTENDS the leak-scan coverage for the new block: the whole-tree
// TestTemplateNoInternalContentLeak already walks templates/CLAUDE.md, but this
// test pins the new managed block explicitly with a targeted default+strict
// forbidden-class scan over the block region, plus presence + emptiness guards.
// A future edit that plants an internal token (or a bullet) in the template
// block fails here.
func TestTemplateLearnedWorkflowBlockNeutral(t *testing.T) {
	t.Parallel()

	const (
		heading     = "## MOAI:LEARNED-WORKFLOW"
		startMarker = "<!-- moai:learned-start -->"
		endMarker   = "<!-- moai:learned-end -->"
	)

	rel := filepath.Join(templatesRoot, "CLAUDE.md")
	data, err := os.ReadFile(rel)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	content := string(data)

	// (a) Presence: heading + start + end markers all ship in the template.
	for _, marker := range []string{heading, startMarker, endMarker} {
		if !strings.Contains(content, marker) {
			t.Errorf("templates/CLAUDE.md missing MOAI:LEARNED-WORKFLOW marker %q", marker)
		}
	}

	headingIdx := strings.Index(content, heading)
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if headingIdx < 0 || start < 0 || end < 0 || end < start {
		t.Fatalf("MOAI:LEARNED-WORKFLOW markers malformed (heading=%d start=%d end=%d)",
			headingIdx, start, end)
	}

	// (b) Emptiness: the block ships with ZERO bullets — the body strictly
	// between the start and end markers is whitespace-only.
	body := content[start+len(startMarker) : end]
	if strings.TrimSpace(body) != "" {
		t.Errorf("MOAI:LEARNED-WORKFLOW template block must ship EMPTY (zero bullets); body=%q", body)
	}

	// (c) Neutrality: the block region (heading through end marker) carries no
	// forbidden-class content across the default + strict tiers. An empty block
	// trivially passes; a planted internal SPEC ID / REQ token / date / SHA fails.
	region := content[headingIdx : end+len(endMarker)]
	classes := make([]leakClass, 0, len(leakClasses)+len(strictLeakClasses))
	classes = append(classes, leakClasses...)
	classes = append(classes, strictLeakClasses...)
	if v := collectLeakViolations(rel, "CLAUDE.md", region, classes); len(v) != 0 {
		t.Errorf("MOAI:LEARNED-WORKFLOW template block leaked forbidden content: %v", v)
	}
}
