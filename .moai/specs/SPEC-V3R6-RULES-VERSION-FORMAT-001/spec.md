---
id: SPEC-V3R6-RULES-VERSION-FORMAT-001
title: "Rules version-staleness corrections + format/consistency normalization"
version: "0.1.0"
status: completed
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tags: "rules, version-staleness, format-normalization, instruction-language, mirror-parity"
related_specs: [SPEC-V3R6-RULES-HOTFIX-001, SPEC-V3R6-RULES-CATALOG-SCRUB-001, SPEC-V3R6-RULES-SSOT-DEDUP-001]
---

# SPEC-V3R6-RULES-VERSION-FORMAT-001 — Rules version + format normalization

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-06-19 | 0.1.0 | Initial plan-phase draft. SPEC 4 (final) of the Sprint 16 rules-improvement cohort. Owns version-staleness + format/consistency normalization across `.claude/rules/moai/`. | manager-spec |

## §A. Context

This is the fourth and final SPEC of the Sprint 16 "rules-improvement" cohort. The four SPECs partition the `.claude/rules/moai/` improvement surface into non-overlapping lanes:

| SPEC | Lane (owns) |
|------|-------------|
| SPEC-V3R6-RULES-HOTFIX-001 | Token fixes |
| SPEC-V3R6-RULES-CATALOG-SCRUB-001 | Archived-agent references |
| SPEC-V3R6-RULES-SSOT-DEDUP-001 | De-duplication + structural consolidation |
| **SPEC-V3R6-RULES-VERSION-FORMAT-001 (this)** | **Version-staleness corrections + format/consistency normalization** |

This SPEC stays strictly in its lane: it corrects stale version identities (model + language toolchains), normalizes instruction-doc format/consistency (footers, instruction-language, emoji, time-estimation, precision framing). It does NOT scrub archived agents (CATALOG-SCRUB owns that), and does NOT de-duplicate or structurally consolidate SSOT content (SSOT-DEDUP owns that).

### §A.1 Verified repo facts (encoded into the build mechanism)

- **No `embedded.go`.** Templates embed via `//go:embed all:templates` (`internal/template/embed.go:28`) plus `//go:embed catalog.yaml`. Verified absent: `internal/template/embedded.go`.
- **`make build`** = templ-generate + gen-catalog-hashes + go build. There is no separate embed-regeneration step touching a generated `embedded.go`.
- **Mirror parity** is asserted by the Go-test trio, NEVER a naive diff:
  - `TestRuleTemplateMirrorDrift` (`internal/template/rule_template_mirror_test.go`) — byte-parity allowlist between `.claude/rules/**` and `internal/template/templates/.claude/rules/**`.
  - `TestTemplateNoInternalContentLeak` (`internal/template/internal_content_leak_test.go`) — §25-sanitized internal-content isolation.
  - `TestTemplateNeutralityAudit` (`internal/template/template_neutrality_audit_test.go`) — 16-language neutrality.
- Every edited `.claude/rules/moai/**` file has a mirror at `internal/template/templates/.claude/rules/moai/**` that needs the identical edit (verified present for all in-scope files).

### §A.2 Constitution-validator coupling (zone-registry clause-as-substring)

`internal/constitution/validator.go:261-271` performs a DRIFT check: each `zone-registry.md` entry's `clause:` field must appear as a (whitespace-normalized) substring of its source `file:`. `moai constitution validate` emits a `SentinelDrift` finding when the substring match fails. Three zone-registry clauses carry "Opus 4.7" text whose source lives in `moai-constitution.md` / `context-window-management.md`:

| Registry entry | Source file + anchor | Coupled source line | Edit decision |
|----------------|----------------------|---------------------|---------------|
| CONST-V3R2-028 (L291) | moai-constitution.md `#opus-47-prompt-philosophy` | Principle 4 line (constitution L56) | **EDIT** (atomic pair with source) |
| CONST-V3R2-029 (L299) | moai-constitution.md `#opus-47-prompt-philosophy` | Principle 5 line (constitution L57) | **EDIT** (atomic pair with source) |
| CONST-V3R5-022 (L845) | context-window-management.md `#context-window-targets` | model-threshold TABLE | **DO NOT EDIT** (D2 carve-out — already a non-substring; source is a table, not prose; AC-VFM-001e) |

[HARD] **The validator is baseline-RED today (D1).** `moai constitution validate` aborts at registry load — `entry 114 validation error: rule ID "CONST-V3R6-001" does not match pattern "^CONST-V3R[25]-\d{3,}$"` (`rule.go:10` admits only V3R2/V3R5; the deployed registry's `CONST-V3R6-001` entry is rejected). The substring drift loop never runs. Therefore the editable clauses (L291/L299) MUST be verified DIRECTLY via `grep -Fq` (each is a `strings.Contains` substring of its identically-edited source) — NOT via the dead validator. The full `moai constitution validate` green-gate is BLOCKED on the sibling `SPEC-V3R6-RULES-CONST-RULEID-001` (broadens `ruleIDPattern` to admit V3R6) landing first; that Go fix is OUT OF SCOPE here. See §C, §D, and AC-VFM-001a/b/e.

## §B. GEARS Requirements

### REQ-VFM-001 — Opus model identity correction (official-alignment)

The rules corpus **shall** state the current reasoning substrate as "Opus 4.7+ / 4.8" (or "1M-context models" where the threshold table reads generically) wherever a bare "Opus 4.7" reads as the current substrate, while preserving genuine 4.6/4.7 backward-compatibility caveats verbatim.

- The constitution **shall** retain its genuine back-compat caveats (e.g. "On Opus 4.6, the highest supported effort level is `high`") unchanged.
- **Where** a `zone-registry.md` `clause:` field carrying "Opus 4.7" is edited (CONST-V3R2-028 L291, CONST-V3R2-029 L299), the edited clause text **shall** remain a `grep -F` substring of its identically-edited source line. This invariant **shall** be verified DIRECTLY via `grep -Fq`, NOT via `moai constitution validate` — that gate is baseline-RED today (registry load aborts on the `CONST-V3R6-001` entry, which `ruleIDPattern = ^CONST-V3R[25]-\d{3,}$` rejects; `rule.go:10`), so the substring drift loop is dead code until a separate fix lands.
- The full `moai constitution validate` green-gate **shall** be treated as a SHOULD blocked on the sibling `SPEC-V3R6-RULES-CONST-RULEID-001` (1-line `ruleIDPattern` broadening to admit V3R6) landing FIRST. Sequencing dependency: CONST-RULEID-001 → this SPEC.
- The CONST-V3R5-022 (L845) clause **shall not** be edited: it is ALREADY a non-substring of its source (`grep -cF '1M context (Opus 4.7)' context-window-management.md` == 0 — the source is a TABLE, not the clause's prose). Editing it cannot produce a clean substring without a table-text rewrite (out of scope). L845 is SHOULD-scoped and left as-is.

### REQ-VFM-002 — Language toolchain version corrections (official-alignment)

The `languages/` rules **shall** carry current, non-stale toolchain version identities for the genuinely-stale pins, and **shall** leave verified-current pins unchanged.

- `csharp.md` **shall** read `.NET 10 (LTS) / C# 14` (was `.NET 8 / C# 12`; .NET 8 EOL Nov 2026).
- `kotlin.md` **shall** read `Kotlin 2.2+ (current 2.4) / Ktor 3.x (current 3.5)` (was `Kotlin 2.0 / Ktor 3.0`).
- `ruby.md` **shall** read `Ruby 4.0` for the version-identity pin (was `Ruby 3.3+`; Ruby 4.0 released Dec 2025) AND **shall** read `Rails 8.0` (was `Rails 7.2`; Rails 8.x is current). The Ruby and Rails pins **shall** be bumped together — a Ruby-4.0/Rails-7.2 pairing is internally inconsistent.
- `go.md` **shall** pin the Docker base image tag to `1.26-alpine` (was `golang:1.23-alpine`), while retaining the `Go 1.23+` floor as an acceptable minimum.
- The rules **shall not** alter verified-correct pins: `python.md` (Python 3.13+/Django 5.2 LTS), `swift.md` (Swift 6+), `php.md` (8.4), `elixir.md` (1.18), `rust.md` (1.92), and `typescript.md`'s React 19 / Next.js 16 references.
- **Where** an acceptable-floor pin is annotatable with the current release (SHOULD), the rules MAY note the current version: `java.md` (Java 21 LTS floor; current 25), `cpp.md` (current C++26 ratified Mar 2026), `typescript.md` (TS 5.9 floor; current 6.0).

### REQ-VFM-003 — Footer consistency policy (structure)

The rules corpus **shall** adopt and document a single footer-consistency policy resolving the ~36 rule files that lack a `Version`/`Status`/`Classification` footer.

- The policy **shall** be one of: (a) add a uniform footer to every rule file, OR (b) document that short path-scoped rules MAY omit footers (consistent-by-absence).
- The chosen policy **shall** be stated explicitly in plan.md and applied to the highest-value SSOT-owning files only (the policy statement itself, not bulk footer-adding, is the deliverable).

### REQ-VFM-004 — Instruction-language policy (best-practice)

**Where** a `.claude/rules/moai/**` instruction document contains Korean instruction prose, the rules **shall** render that prose in English, per `coding-standards.md` § Language Policy (instruction docs must be English).

- `core/settings-management.md` **shall** render its Korean `alwaysLoad` notes (L40-44) in English.
- `workflow/session-handoff.md` § Diet Constraints + § V0 Abort Gate Doctrine (the large Korean instruction blocks) **shall** be rendered in English.
- The rules **shall not** translate the canonical Korean paste-ready/localization-table example renderings in `session-handoff.md` (e.g. `전제 검증:` / `실행:` / `머지 후:`) — those are intentional locale-output examples, not instruction prose.

### REQ-VFM-005 — Emoji removal from instruction docs (best-practice)

The rules **shall not** use emoji characters (e.g. `✅`) in instruction-doc body text, per `coding-standards.md` § Content Restrictions.

- `development/skill-writing-craft.md` **shall** replace `✅` with text ("Required" / "Yes").
- `development/skill-ab-testing.md` **shall** replace `✅` with text ("PASS" / "Met").

### REQ-VFM-006 — Time-estimation removal (best-practice)

The rules **shall not** carry time predictions or duration estimates, per `agent-common-protocol.md` § Time Estimation; phase/priority ordering **shall** replace time language.

- `development/orchestrator-templates.md` **shall** replace wall-time language (e.g. "2 days") with phase/priority ordering.
- `development/manager-develop-prompt-template.md` § verification targets **shall** replace wall-time targets (e.g. "≤30분", "91분") with phase/priority ordering.
- `workflow/team-pattern-cookbook.md` ("Day 1/2/3" sequence) **shall** replace day-labeled steps with phase ordering — coordinated with SSOT-DEDUP-001 (see §C dependency note).

### REQ-VFM-007 — Precision framing fixes (official-alignment)

The rules **shall** state two verified-current technical facts precisely.

- `workflow/session-handoff.md` (Block 1 spec) **shall** clarify that `ultrathink` sets `effort: xhigh`, and that Adaptive Thinking is a DISTINCT axis (the thinking mode, explicitly enabled) — not phrased as if `ultrathink` toggles Adaptive Thinking.
- `core/hooks-system.md` **shall** disambiguate the "Setup REMOVED" framing: the upstream Claude Code `Setup` hook event is CURRENT (triggered via `--init-only`/`--init`/`--maintenance`); only the moai-adk Go `EventSetup` constant + `moai hook setup` subcommand were retired (internal). The rule **shall not** imply the upstream CC `Setup` event is gone.

### REQ-VFM-008 — Template-First mirror parity (HARD invariant, non-vacuous)

**When** any `.claude/rules/moai/**` file is edited, the corresponding `internal/template/templates/.claude/rules/moai/**` mirror **shall** receive the identical edit, and the Go-test trio (`TestRuleTemplateMirrorDrift`, `TestTemplateNoInternalContentLeak`, `TestTemplateNeutralityAudit`) **shall** pass.

- **Where** an edited file is NOT enrolled in the `TestRuleTemplateMirrorDrift` `workflowOptMirroredPaths` allowlist, mirror parity **shall** be asserted by an explicit per-file `diff -q <deployed> <mirror>` byte-identity check — a bare "`TestRuleTemplateMirrorDrift` PASS" clause is vacuously green for non-enrolled files and **shall not** be the sole assertion.
- **Where** an edited file is intentionally divergent deployed↔mirror (§25-sanitized — e.g. `zone-registry.md`'s deployed-only `CONST-V3R6-001` block), mirror parity **shall** be asserted by `TestTemplateNoInternalContentLeak` plus a scoped `diff` confirming only the edited lines changed identically; a whole-file `diff -q` **shall not** be used (it false-fails on the intentional divergence).
- The per-file mechanism assignment **shall** be derived from a verified `diff` of each edited file (acceptance.md §D.4), not assumed.

## §C. Constraints

- [HARD] **Template-First**: every deployed-file edit has a mirror edit; parity verified by the Go-test trio, NOT `embedded.go` (which does not exist) and NOT naive diff.
- [HARD] **§25 neutrality**: no internal SPEC IDs / REQ tokens / audit citations / internal dates / commit SHAs leak into `internal/template/templates/**`. The version/format edits are generic-prose and neutrality-safe; `TestTemplateNoInternalContentLeak` + `TestTemplateNeutralityAudit` must remain green.
- [HARD] **`moai constitution validate` clean**: after editing any zone-registry `clause:` field, the validator must emit no `SentinelDrift`. This requires the coupled source line be edited identically (§A.2 table). If a clause's source-sync is too entangled, scope that clause SHOULD with a note rather than risk a half-edit.
- [HARD] Every fix is a grep-verifiable AC in BOTH the deployed file AND its template mirror.
- [HARD] Stay in lane: no archived-agent scrubbing (CATALOG-SCRUB), no SSOT de-dup/structural consolidation (SSOT-DEDUP).

## §D. Out of Scope (exclusions)

The following are explicitly out of scope for this SPEC. Each excluded item names where it belongs instead.

### Out of Scope — Language-file MUST/MUST NOT skeleton unification
- The 12 (of 16) language files that lack `## MUST` / `## MUST NOT` sections (audit finding L3) are NOT re-authored here. Unifying all 16 language files on one structural skeleton is a substantial re-authoring effort larger than milestones A-G combined. DEFERRED to a dedicated follow-up SPEC `SPEC-V3R6-RULES-LANG-SKELETON-001` (forward-linked). This SPEC corrects version identity within language files but does NOT restructure them.

### Out of Scope — constitution `ruleIDPattern` Go fix (CONST-V3R6 admission)
- The 1-line Go change to broaden `internal/constitution/rule.go` `ruleIDPattern` from `^CONST-V3R[25]-\d{3,}$` to admit `V3R6` is owned by the sibling `SPEC-V3R6-RULES-CONST-RULEID-001`. This SPEC is doc-only and does NOT modify Go source. Sequencing dependency: CONST-RULEID-001 → this SPEC (the full `moai constitution validate` green-gate, AC-VFM-001b, only becomes reachable after CONST-RULEID-001 merges). This SPEC's own zone-registry edits are verified via `grep -Fq` substring proof, which does not depend on the validator running.

### Out of Scope — Archived-agent reference scrubbing
- Removal or migration of references to the 12 archived agents (`manager-strategy`, `expert-*`, etc.) is owned by SPEC-V3R6-RULES-CATALOG-SCRUB-001. This SPEC does NOT touch archived-agent references even when they appear in a file it edits for a version/format reason.

### Out of Scope — SSOT de-duplication and structural consolidation
- Removing duplicated SSOT content, consolidating cross-referenced blocks, or restructuring rule files for de-duplication is owned by SPEC-V3R6-RULES-SSOT-DEDUP-001. This SPEC does NOT de-duplicate; it normalizes version + format only. (Exception coordination: the `team-pattern-cookbook.md` time-estimation fix is owned HERE, sequenced against SSOT-DEDUP's structural edit to that same file — see plan.md §F.)

### Out of Scope — Token-level hotfixes
- Token-level corrections (broken links, typos, stale cross-reference tokens) are owned by SPEC-V3R6-RULES-HOTFIX-001. This SPEC does NOT perform token hotfixes.

### Out of Scope — CLAUDE.md and out-of-tree files
- `CLAUDE.md` §12 (referenced in target G) is out-of-tree relative to `.claude/rules/moai/` and is NOT edited here; only `.claude/rules/moai/**` files are in scope. Any note about CLAUDE.md §12 in plan.md is informational only.

### Out of Scope — Go source code changes
- This SPEC edits documentation rule files only. No `internal/` or `pkg/` Go source is modified (except that the Go-test trio is RUN, not changed). Zero production-code change is expected.

## §E. Self-Verification (plan-phase audit-ready signal)

- [ ] SPEC ID matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` — decomposition printed → PASS (see agent response).
- [ ] 12 canonical frontmatter fields present + valid (no snake_case aliases).
- [ ] Out of Scope section present with ≥1 `### Out of Scope —` H3 sub-heading + `-` bullets (6 sub-headings provided).
- [ ] No implementation HOW (function names, exact diffs) in spec.md — deferred to plan.md/run-phase.
- [ ] Zone-registry clause-sync coupling encoded as HARD constraint + AC.
- [ ] Mirror-parity (Go-test trio, no embedded.go, no naive diff) encoded as HARD constraint + AC.
- [ ] Language-skeleton deferral decision recorded (DEFER → SPEC-V3R6-RULES-LANG-SKELETON-001).
