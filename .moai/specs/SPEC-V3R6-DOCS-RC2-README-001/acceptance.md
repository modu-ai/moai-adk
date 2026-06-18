---
id: SPEC-V3R6-DOCS-RC2-README-001
title: "Acceptance Criteria — v3.0.0-rc2 README + CHANGELOG factual-alignment"
version: "0.2.0"
status: draft
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "repo-root docs"
lifecycle: spec-anchored
tags: "docs, readme, changelog, acceptance, v3r6, rc2"
era: V3R6
tier: M
---

# Acceptance Criteria — SPEC-V3R6-DOCS-RC2-README-001

## §A. Verification Philosophy

Every AC in this document is **independently verifiable** via a mechanical grep against the repo-root docs. The verification model is:

- **Stale-token AC**: `grep -c <stale-token> <file>` MUST return `0` (the stale value is absent).
- **New-value AC**: `grep -c <new-value> <file>` MUST return `≥1` (the authoritative value is present) — **AND when the new-value token already exists elsewhere in the file, the AC MUST be line-anchored** (`sed -n '<line>p' | grep` or `sed -n '<start>,<end>p' | grep`) so it cannot vacuously pass on a pre-existing correct token (see edge case E.6, iter-1 D6 audit).

This model is borrowed from the verification-claim-integrity doctrine (`.claude/rules/moai/core/verification-claim-integrity.md`): a doc-level factual claim is only verified when the mechanical grep reproduces the expected count. Visual inspection is NOT verification.

### A.1 Evidence Format (per AC)

Each AC MUST be reported in the 5-Section Evidence-Bearing Report Format at run-phase M6:

- **Claim**: the AC statement.
- **Evidence**: the verbatim `grep` command + output block.
- **Baseline-attribution**: the `file:line` from `spec.md` §C + the authoritative source from §B.
- **Gaps**: what was NOT verified (e.g., "Mermaid render correctness not verified by grep — visual check at M2").
- **Residual-risk**: e.g., "future LOC growth will re-fire the Go-scale AC if the range phrasing is not honored".

---

## §B. Severity Model

| Severity | Meaning | Failure handling |
|----------|---------|------------------|
| **MUST-PASS** | AC gates SPEC closure. Failure blocks `completed` status transition. | Block; fix before M6 commit. |
| **SHOULD-PASS** | AC gates doc quality but not mechanical correctness. Failure requires justification in M6 evidence. | Justify or fix. |
| **INFO** | Observability AC; failure is a note, not a block. | Record and proceed. |

---

## §C. Traceability Matrix (REQ → AC)

| REQ | AC(s) | Severity | Milestone |
|-----|-------|----------|-----------|
| REQ-EN-001 | AC-EN-001a, AC-EN-001b | MUST-PASS | M1 |
| REQ-EN-002 | AC-EN-002a, AC-EN-002b, AC-EN-002c | MUST-PASS | M1 |
| REQ-EN-003 | AC-EN-003a, AC-EN-003b | MUST-PASS | M1 |
| REQ-EN-004 | AC-EN-004 | MUST-PASS | M1 |
| REQ-EN-005 | AC-EN-005 | MUST-PASS | M1 |
| REQ-EN-006 | AC-EN-006a, AC-EN-006b | MUST-PASS | M2 |
| REQ-EN-007 | AC-EN-007a, AC-EN-007b | MUST-PASS | M2 |
| REQ-EN-008 | AC-EN-008 | MUST-PASS | M2 |
| REQ-EN-009 | AC-EN-009 | SHOULD-PASS | M2 |
| REQ-EN-010 | AC-EN-010 | SHOULD-PASS | M1 |
| REQ-KO-001 | AC-KO-001a, AC-KO-001b | MUST-PASS | M3 |
| REQ-KO-002 | AC-KO-002a, AC-KO-002b | MUST-PASS | M3 |
| REQ-KO-003 | AC-KO-003 | MUST-PASS | M3 |
| REQ-CL-001 | AC-CL-001 | MUST-PASS | M4 |
| REQ-CL-002 | AC-CL-002 | MUST-PASS | M4 |
| REQ-CL-003 | AC-CL-003 | MUST-PASS | M4 |
| REQ-CLAUDE-001 | AC-CLAUDE-001a, AC-CLAUDE-001b | MUST-PASS | M5 |
| REQ-X-001 | AC-X-001 | MUST-PASS | M6 |
| REQ-X-002 | AC-X-002 | MUST-PASS | M6 |
| REQ-X-003 | AC-X-003 | SHOULD-PASS | M6 |

---

## §D. Acceptance Criteria

### §D.1 README.md (EN) — Quantitative (M1)

#### AC-EN-001a — Hero agent count (stale absent)
**Given** the `README.md` hero line at the DRIFT-EN-01 location,
**When** `grep -c "24 specialized" README.md` is run,
**Then** the command MUST return `0`.

#### AC-EN-001b — Hero agent count (new present, line-anchored on L40)
**Given** the `README.md` hero line at the DRIFT-EN-01 location (`README.md:40`),
**When** `sed -n '40p' README.md | grep -cE "8 retained|8 (AI )?agents"` is run,
**Then** the command MUST return `1`.
> **Why line-anchored (iter-1 D6 fix)**: `grep -cE "8 (retained )?agents" README.md` returns 3 today (L297, L322, L334 — verified LIVE 2026-06-19), so a whole-file grep would pass WITHOUT touching the L40 hero "24 specialized AI agents". Anchoring on `sed -n '40p'` forces L40 to change. Run-phase: re-resolve the hero line via `grep -n "specialized AI agents" README.md` if L40 has shifted.

#### AC-EN-002a — Key Numbers stale agents absent
**Given** the `README.md` Key Numbers block,
**When** `grep -c "26 agents\|26\b.*specialized" README.md` is run,
**Then** the command MUST return `0`.

#### AC-EN-002b — Key Numbers stale skills absent
**Given** the Key Numbers block,
**When** `grep -c "47 skills" README.md` is run,
**Then** the command MUST return `0`.

#### AC-EN-002c — Key Numbers new values present (line-anchored on Key Numbers block)
**Given** the Key Numbers block (`README.md:62-65` region — the `- **N** ...` bulleted list),
**When** `sed -n '60,70p' README.md | grep -c "32"` is run (the skills count "32" appears in the Key Numbers block),
**Then** the command MUST return `≥1`;
**And when** `sed -n '60,70p' README.md | grep -c "moai-\*\|32 .*skills"` is run, the command MUST return `≥1` (the count is qualified as moai-* template-managed per REQ-EN-001; the harness-moaiadk-* exclusion is stated in adjacent prose or the §B reproduction note).
> **Why line-anchored (iter-1 D6 fix)**: although "32 skills" does not pre-exist anywhere today, the agent count token "8" DOES pre-exist in the same block region; scoping to `sed -n '60,70p'` ensures the new value lands in the Key Numbers block specifically and the moai-* qualification is co-located.

#### AC-EN-003a — Delegation prose stale absent
**Given** the `README.md` delegation prose at DRIFT-EN-04,
**When** `grep -c "delegates to 24 specialized agents" README.md` is run,
**Then** the command MUST return `0`.

#### AC-EN-003b — Delegation prose new present (line-anchored on L262)
**Given** the delegation prose at DRIFT-EN-04 (`README.md:262`),
**When** `sed -n '262p' README.md | grep -cE "8 retained|8 (AI )?agents"` is run,
**Then** the command MUST return `1`.
> **Why line-anchored (iter-1 D2 fix)**: `README.md:297` already contains `**Total: 8 retained agents**` (verified LIVE 2026-06-19). A whole-file `grep -cE "8 (retained )?agents" README.md` returns 3 today and would pass WITHOUT touching L262. Anchoring on `sed -n '262p'` forces the stale "delegates to 24 specialized agents" line to actually change. The run-phase line number MAY drift slightly if earlier edits shift content; the run-phase agent MUST re-resolve the line via `grep -n "delegates to" README.md` and anchor on whichever line that returns.

#### AC-EN-004 — Go scale graceful-aging
**Given** the Go scale citation at DRIFT-EN-02 (`README.md:62`),
**When** `grep -c "38,700" README.md` is run, the command MUST return `0`;
**And when** `grep -cE "100K\+ lines" README.md` (or `grep -c "100+ packages" README.md`) is run, the command MUST return `≥1`.
> Non-vacuous today (`100K+` returns 0); no line-anchor needed. The stale `38,700` figure MUST be absent.

#### AC-EN-005 — Language count 16 (line-anchored on Key Numbers block)
**Given** the language count citation at DRIFT-EN-03 (`README.md:64` Key Numbers block),
**When** `grep -c "18 languages\|18 programming" README.md` is run, the command MUST return `0`;
**And when** `sed -n '60,70p' README.md | grep -c "16"` is run (the corrected "16 languages" lands in the Key Numbers block), the command MUST return `≥1`.
> **Why line-anchored (iter-1 D6 fix)**: `grep -c "16 languages" README.md` returns 3 today (L80, L514, L1168 — the language-agnostic table rows), so a whole-file grep would pass WITHOUT touching the L64 Key Numbers `18 languages`. Anchoring on `sed -n '60,70p'` forces the Key Numbers block to change. (The other 3 pre-existing `16 languages` mentions are correct and stay.)

#### AC-EN-010 — Statusline FAQ refresh (concrete grep predicate)
**Given** the statusline FAQ at DRIFT-EN-09,
**When** EITHER `grep -cE "explicitly illustrative|example only" README.md` returns `≥1` (option b — marked illustrative) OR `grep -cE "Opus 4\.[78]|CC 2\.1\.17" README.md` returns `≥1` (option a — refreshed to current),
**Then** the AC PASSES (at least one of the two grep predicates MUST return ≥1).
> **Why concrete predicate (iter-1 D7 fix)**: the prior "M6 MUST capture grep OR prose note" wording was untestable. The two predicates above are the mechanical oracles. Severity remains SHOULD-PASS; failure requires both predicates to return 0, in which case M6 MUST record a justification.

---

### §D.2 README.md (EN) — Mermaid + /moai db + expert-frontend + lifecycle (M2)

#### AC-EN-006a — Mermaid stale bare labels absent (line-anchored on L264-286 block)
**Given** the rewritten AI Agent Orchestration Mermaid at DRIFT-EN-05 (fence `README.md:264`, close `README.md:286`),
**When** `sed -n '264,286p' README.md | grep -cE 'Manager \(8\)|Expert \(8\)|Builder \(3\)|Evaluator \(2\)|Design System \(4\+1\)'` is run, the command MUST return `0`;
**And when** `sed -n '264,286p' README.md | grep -cE 'backend · frontend · security · devops'` is run (the archived-agent node body), the command MUST return `0`.
> **Why the bare-label form (iter-1 D3 fix)**: the current Mermaid block uses BARE labels `Manager (8)` / `Expert (8)` / `Builder (3)` / `Evaluator (2)` / `Design System (4+1)` + a node body `backend · frontend · security · devops<br/>performance · debug · testing · refactoring`. The hyphenated `expert-(frontend|backend|...)` form returns 0 today because those forms appear ONLY in the Design System table (L921+), NOT in the Mermaid block — so a whole-file `grep "expert-(frontend|...)"` would let run-phase leave the Mermaid untouched. Anchoring on `sed -n '264,286p'` + the bare-label form forces the actual stale content to change. (The run-phase line range MAY shift slightly; re-resolve via `grep -n '```mermaid' README.md` to find the orchestration-diagram fence and use the matching close-fence line.)
**Companion whole-file checks (unchanged, still required)**: `grep -cE "manager-(strategy|quality|brain|project)" README.md` MUST return `0`; `grep -c "claude-code-guide" README.md` MUST return `0` for MoAI-custom references (a reference to the Anthropic built-in helper is NOT a violation; M6 MUST distinguish).

#### AC-EN-006b — Mermaid retained agents present (line-anchored on L264-286 block)
**Given** the rewritten Mermaid,
**When** `sed -n '264,286p' README.md | grep -cE "manager-spec|manager-develop|manager-docs|manager-git|plan-auditor|sync-auditor|builder-harness|Explore"` is run,
**Then** the command MUST return `≥1` (at least one retained-agent node name appears INSIDE the rewritten Mermaid block, not merely elsewhere in the file).

#### AC-EN-007a — /moai db section removed
**Given** the `/moai db` section at DRIFT-EN-08,
**When** `grep -cE "^#+ .*/moai db" README.md` is run (matching a heading introducing the section), the command MUST return `0`;
**And when** `grep -cE "/moai db" README.md` is run, the count MUST be `≤1` (at most the one-line pointer).

#### AC-EN-007b — db-schema-sync pointer present
**Given** the one-line pointer substitution,
**When** `grep -c "moai hook db-schema-sync" README.md` is run, the command MUST return `≥1`.

#### AC-EN-008 — expert-frontend replaced (vicinity-anchored on Design System table)
**Given** the Design System implementer table at DRIFT-EN-07 (`README.md:921,939,969`),
**When** `grep -n "expert-frontend" README.md` is run, the command MUST return `0` lines (no matches anywhere);
**And when** `sed -n '915,975p' README.md | grep -cE "manager-develop|harness|cycle_type"` is run (the Design System implementer table region), the command MUST return `≥1`.
> **Why vicinity-anchored (iter-1 D6 fix)**: `manager-develop` already appears 2× elsewhere in README today, so a whole-file grep would pass without touching the Design System table. Anchoring on `sed -n '915,975p'` forces the replacement to land in the implementer table. (Line range MAY shift; re-resolve via `grep -n "implementer\|Design System" README.md`.)

#### AC-EN-009 — Lifecycle note present (vicinity-anchored on Plan→Run→Sync pipeline section)
**Given** the Plan→Run→Sync pipeline documentation at DRIFT-EN-06 (`README.md:451-478`),
**When** `sed -n '445,485p' README.md | grep -cE "3-phase|plan.{0,3}run.{0,3}sync|Mx.*retir"` is run (the pipeline section region), the command MUST return `≥1`;
**And when** `grep -cE "SPEC-V3R6-LIFECYCLE-REDESIGN-001" README.md` is run, the command MUST return `≥1` (OR the lifecycle note is phrased without a SPEC citation, in which case this sub-check is INFO).
> **Why vicinity-anchored (iter-1 D6 fix)**: `grep -cE "3-phase|plan.{0,3}run.{0,3}sync" README.md` already returns 1 today (elsewhere in the file), so a whole-file grep would pass without adding the lifecycle note to the L451-478 pipeline section. Anchoring on `sed -n '445,485p'` forces the note to land in the pipeline section.

---

### §D.3 README.ko.md (KO) — Mirror + What's New (M3)

#### AC-KO-001a — KO stale tokens absent
**Given** the `README.ko.md` file,
**When** the following commands are each run, each MUST return `0`:
- `grep -c "24 specialized" README.ko.md` (or the KO equivalent phrasing)
- `grep -c "26 agents" README.ko.md` (or KO equivalent)
- `grep -c "18 languages" README.ko.md` (or KO equivalent)
- `grep -c "expert-frontend" README.ko.md`
- `grep -cE "^#+ .*/moai db" README.ko.md`

#### AC-KO-001b — KO new tokens present (line-anchored; D6 fix)
**Given** the `README.ko.md` file,
**When** the following commands are each run, each MUST return `≥1`:
- `grep -cE "16개.*(언어|프로그래밍)|16 (programming )?languages" README.ko.md` — the corrected language count (LIVE 2026-06-19: KO file currently says `18개 프로그래밍 언어` at L111, so "16" does NOT pre-exist; this check is non-vacuous)
- `grep -c "moai hook db-schema-sync" README.ko.md` — the db pointer (does NOT pre-exist)
- `grep -cE "32.*(moai-\*|skills|스킬)" README.ko.md` — the skills count with moai-* qualification (does NOT pre-exist in KO today)
> **Why not `8 (retained )?agents` (iter-1 D6 fix)**: the KO file ALREADY contains `8개 retained 에이전트` at L40, L110, L308, L343, L368, L380 (6 places, verified LIVE 2026-06-19). The agent-count surface in KO is already correct — the actual KO staleness is the language count (18→16), the missing skills count, and the missing db-schema-sync pointer. Those three predicates are the non-vacuous checks. The KO-native phrasing (`8개`, `16개`, `언어`, `스킬`) is accepted per edge case E.3.

#### AC-KO-002a — KO What's New v2.17.0 stale heading absent
**Given** the `README.ko.md` What's New section at DRIFT-KO-02,
**When** `grep -cE "^#+ .*v2\.17\.0" README.ko.md` is run, the command MUST return `0` for the active What's New heading (a historical reference inside a changelog-style block is NOT a violation; M6 MUST distinguish).

#### AC-KO-002b — KO What's New v3/V3R6 content present
**Given** the rewritten What's New section,
**When** the following commands are each run, each MUST return `≥1`:
- `grep -cE "v3\.0\.0-rc2|V3R6" README.ko.md`
- `grep -c "glm-5.2" README.ko.md`
- `grep -cE "8 (retained )?agents" README.ko.md`
- `grep -cE "3-phase|plan.{0,3}run.{0,3}sync" README.ko.md`

#### AC-KO-003 — KO harness namespace fix
**Given** the harness namespace line at DRIFT-KO-03,
**When** `grep -c "my-harness-" README.ko.md` is run, the command MUST return `0`;
**And when** `grep -cE "moai-\* .*template-managed.*harness-\* .*user-owned|harness-\* .*user-owned.*moai-\* .*template-managed" README.ko.md` is run (or the KO equivalent phrasing), the command MUST return `≥1`.

---

### §D.4 CHANGELOG.md (M4)

#### AC-CL-001 — rc2 section promoted
**Given** the `CHANGELOG.md` top section,
**When** `grep -cE "^## \[v3\.0\.0-rc2\]" CHANGELOG.md` is run, the command MUST return `≥1`;
**And when** `grep -cE "^## \[Unreleased\]" CHANGELOG.md` is run, the command MUST return `0` for the rc2-cohort content (a fresh empty `## [Unreleased]` placeholder for future work is permissible; M6 MUST distinguish).

#### AC-CL-002 — Mis-labeled HEADING lines cleaned (heading-anchored)
**Given** the mis-labeled subsection headers at DRIFT-CL-02 (LIVE 2026-06-19: 4 heading lines at L378/L487/L507/L527),
**When** `grep -cE '^##.*v2\.20\.0-rc1' CHANGELOG.md` is run (heading-line form — `## ` followed by the mis-labeled version), the command MUST return `0`.
> **Why heading-anchored (iter-1 D4 fix)**: `grep -c "v2.20.0-rc1" CHANGELOG.md` returns 28 today (4 headings + 24 body mentions) and would require purging every body mention — over-scope. The REQ-CL-002 scope decision is **heading-line removal**: consolidate the 4 `## [Unreleased] — v2.20.0-rc1 ...` headings into the correct `## [v3.0.0-rc2]` section or a fresh `## [Unreleased]`. Body prose that legitimately references the historical milestone inside consolidated content is acceptable; the AC greps the anchored `^##.*v2\.20\.0-rc1` form only.

#### AC-CL-003 — No stable-version claim
**Given** the `CHANGELOG.md`,
**When** `grep -cE "^## \[v3\.0\.0\]" CHANGELOG.md` is run, the command MUST return `0` (no `v3.0.0` stable section).

---

### §D.5 CLAUDE.md (M5)

#### AC-CLAUDE-001a — Stale builder-harness path absent
**Given** the `CLAUDE.md` §4 Agent Catalog,
**When** `grep -c "agents/builder/builder-harness" CLAUDE.md` is run, the command MUST return `0`.

#### AC-CLAUDE-001b — Correct builder-harness path present
**Given** the `CLAUDE.md` §4 Agent Catalog,
**When** `grep -c "agents/moai/builder-harness" CLAUDE.md` is run, the command MUST return `≥1`.

---

### §D.6 Cross-Cutting (M6)

#### AC-X-001 — Version-wording invariant
**Given** all 4 touched files (`README.md`, `README.ko.md`, `CHANGELOG.md`, `CLAUDE.md`),
**When** `grep -rn "v3\.0\.0 stable" README.md README.ko.md CHANGELOG.md CLAUDE.md` is run, the command MUST return `0` lines;
**And when** `grep -rn "v3\.0\.0-rc2" README.md README.ko.md CHANGELOG.md CLAUDE.md` is run, the command MUST return `≥1` line per file where a version is cited.

#### AC-X-002 — Superseded SPEC citation invariant
**Given** all 4 touched files,
**When** `grep -rn "SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001" README.md README.ko.md CHANGELOG.md CLAUDE.md` is run, the command MUST return `0` lines.

#### AC-X-003 — Draft-SPEC phrasing invariant
**Given** all 4 touched files,
**When** `grep -nE "SPEC-V3R6-RULES-(SSOT-DEDUP|VERSION-FORMAT)-001" README.md README.ko.md CHANGELOG.md CLAUDE.md` is run,
**Then** any match MUST be accompanied (within the same paragraph or section) by in-flight phrasing ("draft", "in progress", "planned", "not yet finalized"); a bare finalized assertion is a failure. M6 MUST capture the context for each match.

---

## §E. Edge Cases

### E.1 `claude-code-guide` ambiguity (AC-EN-006a)

The archived MoAI-custom `claude-code-guide` agent file is distinct from the official Claude Code built-in helper agent of the same name. The `README.md` may legitimately reference the built-in. M6 MUST distinguish: a reference that describes the built-in is NOT a violation; a reference that describes the archived MoAI-custom agent IS a violation.

**Mitigation**: M6 MUST capture the context line for every `claude-code-guide` match and classify each as built-in vs archived.

### E.2 `## [Unreleased]` placeholder (AC-CL-001)

CHANGELOG convention often keeps a fresh `## [Unreleased]` section at the top for future entries. AC-CL-001 permits a fresh empty `## [Unreleased]` placeholder; the violation is only if the rc2 cohort content remains under `[Unreleased]` instead of being promoted to `## [v3.0.0-rc2]`.

**Mitigation**: M6 MUST distinguish "fresh empty placeholder" from "rc2 content still under Unreleased".

### E.3 KO numeric phrasing (AC-KO-001b)

Korean prose may phrase "8 agents" as "8개 에이전트", "8개의 에이전트", or "여덟 에이전트". The grep AC MUST accept any of these variants. The numeric "8" + the Korean word for "agent" is the load-bearing signal.

**Mitigation**: The grep pattern in AC-KO-001b uses `8 (retained )?agents` as the canonical form but M6 MAY substitute the KO-native phrasing provided the numeric "8" is present.

### E.4 Graceful-aging range drift (AC-EN-004)

The "100K+ lines" range is intentionally loose. If the codebase grows past 200K lines, the range remains technically correct but becomes less informative. This is NOT an AC failure — it is the intended aging behavior. The next drift sweep re-evaluates.

### E.5 Mermaid render correctness (AC-EN-006a/b)

Grep cannot verify that the rewritten Mermaid diagram renders correctly in GitHub. The AC verifies token presence/absence (declared nodes, no archived nodes); visual render correctness is a residual risk documented in the M6 evidence Gaps section.

### E.6 Line-anchor discipline for "new value present" ACs (iter-1 D6 audit)

A factual-alignment SPEC is uniquely vulnerable to **vacuously-satisfied ACs**: the same authoritative value (e.g., "8 retained agents", "16 languages") often already appears CORRECTLY in one part of the file while remaining STALE in the drift line the SPEC targets. A whole-file `grep -c <new-value>` then passes without touching the drift line.

**Discipline (applied across this suite iter-1)**: every "new value present" AC where the token pre-exists elsewhere MUST be line-anchored via `sed -n '<line>p' | grep` OR `sed -n '<start>,<end>p' | grep` (vicinity). Known pre-existing tokens verified LIVE 2026-06-19:

| Token | Pre-existing correct location(s) | Drift line the AC must force |
|-------|----------------------------------|------------------------------|
| `8 retained agents` / `8 (retained )?agents` | README.md L297, L322, L334; README.ko.md L40/L110/L308/L343/L368/L380 (`8개 retained 에이전트`) | README.md L40 (hero), L262 (delegation prose), L64 (Key Numbers) |
| `16 languages` | README.md L80, L514, L1168 (language-agnostic table rows) | README.md L64 (Key Numbers `18 languages`) |
| `manager-develop` | README.md (2 pre-existing mentions) | README.md L921/939/969 (Design System implementer table `expert-frontend`) |
| `3-phase` / `plan→run→sync` | README.md (1 pre-existing mention) | README.md L451-478 (Plan→Run→Sync pipeline section, needs lifecycle note) |

**Run-phase line-shift handling**: the drift line numbers in spec.md §C were measured at plan-phase (2026-06-19). If an earlier milestone edit shifts content, the run-phase agent MUST re-resolve the line via a `grep -n <distinctive-token>` (e.g., `grep -n "delegates to" README.md` for AC-EN-003b, `grep -n "```mermaid" README.md` for AC-EN-006a/b) and anchor on the re-resolved line. The line-anchor contract is "the specific drift line", not "the literal plan-phase number".

---

## §F. Definition of Done

This SPEC is `completed` only when ALL of the following hold:

- **F.1** Every MUST-PASS AC in §D returns the expected grep count (stale=0, new=≥1), with verbatim command + output captured in M6 evidence.
- **F.2** Every SHOULD-PASS AC is either PASS or carries a recorded justification in M6 evidence.
- **F.3** The 4 touched files (`README.md`, `README.ko.md`, `CHANGELOG.md`, `CLAUDE.md`) carry `v3.0.0-rc2` wording and contain zero `v3.0.0 stable` claims (AC-X-001).
- **F.4** The 4 touched files contain zero citations of the superseded `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` (AC-X-002).
- **F.5** `moai spec lint` on this SPEC returns 0 findings.
- **F.6** The CLAUDE.md template counterpart decision (plan.md §G.1) has been surfaced and resolved (fixed in-scope OR deferred to a follow-up SPEC with a recorded decision).
- **F.7** The §E.2 run-phase evidence section of `progress.md` is populated with the M6 grep output (per the era classification H-4 predicate — `§E.2` marker + `sync_commit_sha` present).

---

## §G. Quality Gate

| Gate | Criterion | Enforcement |
|------|-----------|-------------|
| Stale-token sweep | All stale tokens in spec.md §C grep-absent | M6 AC suite (MUST-PASS) |
| New-value sweep | All authoritative values in spec.md §B grep-present | M6 AC suite (MUST-PASS) |
| Version-wording | `v3.0.0-rc2` only; no `v3.0.0 stable` | AC-X-001 (MUST-PASS) |
| Superseded-SPEC ban | No `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` citation | AC-X-002 (MUST-PASS) |
| spec-lint clean | `moai spec lint` 0 findings | F.5 (MUST-PASS) |
| Era classification | SPEC classifies as V3R6 (not unclassified) | Frontmatter `era: V3R6` + H-4 predicate via progress.md §E.2 |

---

## §H. Forward-Looking Checks

- **H.1** After the v3.0.0 stable release (future SPEC), the `v3.0.0-rc2` wording in these docs will need re-evaluation. This SPEC's ACs pin the rc2 wording; a future stable-release SPEC MUST supersede the relevant ACs.
- **H.2** The CLAUDE.md template counterpart (plan.md §G.1) may need a follow-up SPEC if the template source also carries the stale builder-harness path.
- **H.3** The local `llm.yaml` drift (glm-5.1 vs TEMPLATE SSOT glm-5.2[1m]) is out of scope here but is a candidate for a config-alignment follow-up SPEC.
- **H.4** The KO What's New section will need re-rewriting at the next major version boundary; the rewrite in M3 is pinned to the rc2 generation.
