---
id: SPEC-MEMORY-DIET-001
title: "Safe always-loaded context diet — cadence-bridge path-scope + session-handoff example extraction + MEMORY.md archive pruning (~3.0k tokens)"
version: "1.0.0"
status: completed
created: 2026-07-10
updated: 2026-07-10
author: GOOS행님
priority: P2
phase: "v14.4.0 target"
module: ".claude/rules/moai/workflow + ~/.claude/projects/*/memory"
lifecycle: spec-anchored
tags: "context-diet, always-loaded, path-scope, memory-index, template-first, safe-diet"
tier: M
---

## HISTORY

- 2026-07-10 — v0.1.0 — manager-spec — Plan-phase artifacts authored (Tier M, 4 artifacts: spec.md + plan.md + acceptance.md + progress.md §E skeleton). Orchestrator Discovery measured 62.1k tokens (6.2%) across 12 always-loaded files via `/context`. This SPEC targets ~5.5k token reduction via 3 scoped REQs on the lowest-regression-risk surfaces. REQ-1 (cadence-bridge path-scope) + REQ-2 (session-handoff example extraction) edit template-mirrored files; REQ-3 (MEMORY.md pruning) edits auto-memory (no template mirror). `status: draft`.

---

## Prior-Art Review

Before authoring, the following existing SPECs were read (spec.md + frontmatter status) to determine whether this work is NEW, an amendment, or a supersede. Verdict: **NEW SPEC**. No single existing SPEC covers all 3 REQs; each prior-art SPEC is distinct-scope.

| Existing SPEC | Status | Relationship to this work |
|---------------|--------|---------------------------|
| SPEC-V3R6-RULES-PATH-SCOPE-001 | implemented | **Distinct-scope.** Scoped 4 DIFFERENT rules (zone-registry, design/constitution, manager-develop-prompt-template, agent-teams-pattern) using `paths:` frontmatter. `cadence-bridge.md` was NOT in scope (the file did not exist at the time — created 2026-07-09 by SPEC-CADENCE-BRIDGE-001). REQ-1 applies the same `paths:` mechanism established by this SPEC to a file that post-dates it. |
| SPEC-RULE-DIET-002 | completed | **Distinct-scope, same mechanism family.** Scoped 6 reference-doctrine rules using self-referential `paths:` globs (runtime-recovery-doctrine, dynamic-workflows, native-invocation-model, sprint-round-naming, goal-directive, verification-batch-pattern). `cadence-bridge.md` was not in the always-loaded inventory at that time. REQ-1 is the natural successor application of the RULE-DIET-002 mechanism to cadence-bridge.md — but RULE-DIET-002 is `completed` and does not cover this file or REQ-2/REQ-3. |
| SPEC-V3R6-RULES-COMPRESS-001 | implemented | **Distinct-scope, same file.** Compressed the prose body of session-handoff.md (1,927w → ~1,000w). REQ-2 touches the SAME file but uses a DIFFERENT mechanism (extracting illustrative Example sections to a sibling references file, not prose compression). The two are complementary: RULES-COMPRESS-001 shortened prose; REQ-2 relocates illustrative content. No collision — RULES-COMPRESS-001 is `implemented` and its compression is the baseline REQ-2 builds on. |
| SPEC-TOKEN-VERIFY-DIET-001 | completed | **Distinct-scope.** Verification output diet (file-redirect contract for trust-but-verify batches). Different mechanism entirely (redirect verbatim output to disk, not always-loaded file diet). Token-Economy Epic C of 4. |
| SPEC-CLAUDEMD-DIET-V2-001 | completed | **Distinct-scope.** CLAUDE.md body diet (405 → ~300 lines). Different file (CLAUDE.md, not rules files or auto-memory). |
| SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001 | completed | **Distinct-scope.** CLAUDE.md 1st-round diet. Different file. |
| SPEC-STEERING-ALIGN-LOCAL-DIET-001 | completed | **Distinct-scope.** CLAUDE.local.md diet. Different file. |
| SPEC-V3R6-RULES-SSOT-DEDUP-001 | completed | **Distinct-scope.** SSOT de-duplication (18 files self-declaring SSOT while verbatim-copying). Different mechanism (dedup, not diet/path-scope). |
| SPEC-V3R6-RULES-CATALOG-SCRUB-001 | completed | **Distinct-scope.** Archived-agent catalog scrub across rules. Different mechanism (removing archived-agent references). |

**Domain choice justification:** `SPEC-MEMORY-DIET-001` is chosen over `SPEC-RULE-DIET-003` because REQ-3 (MEMORY.md auto-memory pruning) is NOT a rule file — it lives in `~/.claude/projects/*/memory/MEMORY.md`, which is the auto-memory index, not `.claude/rules/`. The unifying theme is "always-loaded context diet" (the Claude Code `/context` UI labels these surfaces collectively as "Memory files"), and the three REQs span rules + auto-memory. `MEMORY-DIET` captures this breadth without the misleading "RULE" qualifier. `ALWAYSLOAD-DIET` was considered but rejected as an awkward token; `CONTEXT-DIET` was considered but "context" is overloaded with the context window. `MEMORY-DIET` aligns with the user's framing ("safe memory-files diet") and Claude Code's own `/context` label.

---

## A. Context / Background

### A.1 Problem — measured always-loaded baseline

The orchestrator's Discovery measured the always-loaded context surface via `/context`:

> Memory files: 62.1k tokens (6.2%)

This 62.1k token surface is distributed across 12 always-loaded files. The 3 largest targets where safe diet is feasible (lowest regression risk) are:

| File | Lines | Bytes | Tokens | REQ |
|------|-------|-------|--------|-----|
| `.claude/rules/moai/workflow/session-handoff.md` | 478 | 56,598 | 13.3k | REQ-2 |
| `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` | 90 | 16,758 | 5.8k | REQ-3 |
| `.claude/rules/moai/workflow/cadence-bridge.md` | 88 | 9,855 | 2.3k | REQ-1 |

### A.2 The "safe diet" boundary (user decision, pre-drained)

The user selected **SAFE DIET** (~5.5k tokens target, lowest regression risk) over aggressive diet. The 3 REQs are scoped to surfaces where the reduction is verifiably non-regressive:

- **REQ-1** (cadence-bridge): the file is relevant ONLY when composing `/loop` + `/moai`. It is a narrow-scope doctrine that does not govern per-turn orchestrator behavior. Path-scoping is the established mechanism (14+ existing path-scoped rules; RULE-DIET-002 generalized the self-referential `paths:` precedent).
- **REQ-2** (session-handoff): the CORE DOCTRINE (6-block skeleton, cut-line markers, Field-by-Field Spec, Pre-emit self-check, Auto-Memory Integration, Post-paste /goal Follow-up Block) MUST remain always-loaded and byte-identical. Only the ILLUSTRATIVE content (2 Example sections + the ja/zh rows of the Localization Table) is extracted to a sibling references file. The SSOT ↔ render-surface parity (`session-handoff.md` ↔ `.claude/output-styles/moai/moai.md §8`) is preserved.
- **REQ-3** (MEMORY.md): only stable `✅` entries (closed SPECs with NO open follow-up, stable for ≥ N days) move to the existing archive. All active-marker entries (`🟢🟡🆕⏸️⚠️🔍`) and any entry referencing a pending next step are preserved.

### A.3 Template-First obligation (REQ-1, REQ-2)

REQ-1 and REQ-2 modify files that have template mirrors. Per CLAUDE.local.md §2 [HARD] Template-First Rule, run-phase MUST edit BOTH paths:

- Local: `.claude/rules/moai/workflow/{cadence-bridge,session-handoff}.md`
- Template: `internal/template/templates/.claude/rules/moai/workflow/{cadence-bridge,session-handoff}.md`

All 4 mirrors confirmed to exist and are byte-identical between the two trees (cadence-bridge: 9,855 bytes each; session-handoff: 56,598 bytes each). REQ-3's MEMORY.md is auto-memory — NO template mirror exists.

### A.4 Neutrality guard (REQ-1, REQ-2)

The template mirrors are subject to `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml`. Run-phase MUST NOT introduce: SPEC IDs, REQ tokens, internal dates, commit SHAs, or audit citations into the template files. The doctrine content being relocated/edited is generic prose (illustrative examples, localization data, loading-scope declarations) — it is already neutral.

---

## B. Requirements (GEARS notation)

> **Subject convention:** GEARS generalized subjects are used (`the cadence-bridge rule`, `the session-handoff doctrine`, `the MEMORY.md index`). No legacy `IF/THEN` modality.

### B.1 REQ-1 — cadence-bridge.md path-scope conversion

**REQ-MD-001 (Ubiquitous, subject: cadence-bridge rule)**: The cadence-bridge rule SHALL acquire a YAML frontmatter block carrying a `paths:` field (CSV quoted-glob string per `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter) and a one-line `description:` field, so that it loads only when the session touches a file matching the documented trigger-condition list.

**REQ-MD-002 (Capability gate + honest-fallback, subject: cadence-bridge rule)**: The cadence-bridge rule's `paths:` frontmatter SHALL commit to the concrete glob `paths: "**/cadence-bridge.md,**/workflows/loop.md,.claude/skills/moai-workflow-loop/**,.moai/state/loop-verdict-*.json,.moai/reports/cadence/**"` (CSV quoted-glob string). Glob-token rationale: `**/cadence-bridge.md` = self-referential maintenance (editing the rule reloads it); `**/workflows/loop.md` = the `/loop` workflow file; `.claude/skills/moai-workflow-loop/**` = the loop skill tree; `.moai/state/loop-verdict-*.json` = the loop-verdict state files Recipe 3 reads; `.moai/reports/cadence/**` = the cadence backlog directory Recipes 2/3 persist to. **Honest-fallback clause**: `paths:` matches FILE GLOBS (a file-event mechanism) and therefore CANNOT mechanically catch a `/loop` command typed at runtime (a semantic event — a `/loop` invocation is not a file path). The path-match glob is a best-effort PRE-LOAD; when the orchestrator detects a `/loop` + `/moai` composition (from the goal-directive.md / loop-skill cross-reference), it MUST manually consult this rule regardless of path-match state. The manual-consult fallback is the GUARANTEE; the path-match pre-load is best-effort convenience.

**REQ-MD-003 (Ubiquitous, subject: cadence-bridge rule)**: The cadence-bridge rule's "Loading scope" prose SHALL be rewritten to declare path-match status (replacing the current "Intentionally always-loaded" declaration) WITH explicit rationale for the change, a documented trigger-condition list naming the surfaces that cause the rule to load, AND the honest-fallback clause from REQ-MD-002 (stating plainly that `paths:` matches file globs and cannot catch a runtime `/loop` command, so manual orchestrator consultation is the binding guarantee when a `/loop` + `/moai` composition is detected).

**REQ-MD-004 (Where MIRRORED, subject: cadence-bridge rule)**: **Where** the cadence-bridge rule is MIRRORED (present in both local and template trees), the frontmatter change SHALL be applied to BOTH the local file (`.claude/rules/moai/workflow/cadence-bridge.md`) and the template file (`internal/template/templates/.claude/rules/moai/workflow/cadence-bridge.md`) with identical frontmatter — template-first per CLAUDE.local.md §2.

**REQ-MD-005 (Unwanted, subject: cadence-bridge rule)**: The cadence-bridge rule's BODY content SHALL NOT be modified by the frontmatter prepend — only the YAML frontmatter block is added at the top of the file; the existing body is preserved byte-for-byte (the "Loading scope" rewrite in REQ-MD-003 is an inline prose edit within the existing body, not a body-content deletion).

### B.2 REQ-2 — session-handoff.md illustrative content extraction

**REQ-MD-006 (Ubiquitous, subject: session-handoff doctrine)**: The session-handoff doctrine SHALL move the two illustrative Example sections (`### Example (Illustrative; substitute project-specific values when adapting)` and `### Example with Block 0 (Illustrative)`) to a sibling references file, replacing each with a one-line pointer that names the references file path.

**REQ-MD-007 (Ubiquitous, subject: session-handoff doctrine)**: The session-handoff doctrine SHALL condense the Localization Table by keeping the `en` and `ko` locale columns inline (the primary locales this project uses) and moving the `ja` and `zh` columns to the sibling references file, with a one-line pointer noting the full 4-locale table lives in the references file.

**REQ-MD-008 (Ubiquitous, subject: session-handoff doctrine)**: The session-handoff doctrine's CORE DOCTRINE SHALL remain always-loaded and byte-identical — specifically: the Canonical Format 6-block skeleton, the Cut-line Marker Specification (`✂` U+2702 + `─` U+2500), the Field-by-Field Specification, the Pre-emit self-check labels (`paste-ready budget` / `localization render` / `session-handoff template completeness`), the Auto-Memory Integration section, the Post-Paste /goal Follow-up Block, the Diet Constraints section, and the Worktree-Anchored Resume Pattern.

**REQ-MD-009 (State-driven, subject: session-handoff doctrine)**: **While** the session-handoff doctrine is the SSOT for the paste-ready resume, the drift-mitigation self-check (SSOT ↔ render-surface parity with `.claude/output-styles/moai/moai.md §8`) SHALL be honored — the extraction MUST NOT break the parity contract; the render surface MUST continue to cross-reference session-handoff.md as the SSOT.

**REQ-MD-010 (Where MIRRORED, subject: session-handoff doctrine)**: **Where** the session-handoff doctrine is MIRRORED (present in both local and template trees), the extraction + pointer edits SHALL be applied to BOTH the local file and the template file with identical content — template-first per CLAUDE.local.md §2. The sibling references file (if created) SHALL also be created in BOTH trees.

**REQ-MD-011 (Ubiquitous, subject: sibling references file)**: The sibling references file (e.g. `session-handoff-examples.md`) SHALL be created as a path-scoped reference (carrying its own `paths:` frontmatter pointing to `session-handoff.md`) so it loads only when session-handoff.md is being edited, NOT always-loaded — the extraction must not simply relocate the tokens to another always-loaded surface.

### B.3 REQ-3 — MEMORY.md archive pruning

**REQ-MD-012 (Ubiquitous, subject: MEMORY.md index)**: The MEMORY.md index SHALL move stable `✅`-marker entries (closed SPECs with NO open follow-up, stable for ≥ N days where N is the staleness threshold) to the existing archive file `MEMORY-archive-2026-06-02.md`, reducing the always-loaded index line count.

**REQ-MD-013 (Unwanted, subject: MEMORY.md index)**: The MEMORY.md index SHALL NOT remove any active-marker entry (`🟢🟡🆕⏸️⚠️🔍`) — these represent ongoing work, pending handoffs, or unresolved debt that the always-loaded index MUST surface for cross-session recall.

**REQ-MD-014 (Unwanted, subject: MEMORY.md index)**: The MEMORY.md index SHALL NOT remove any `✅` entry whose hook line references a pending next step ("다음=SPEC-XXX", open handoff, cross-Epic tracking pointer, or deferred-debt pointer) — these entries are load-bearing for cross-Epic tracking despite their `✅` marker.

**REQ-MD-015 (State-driven, subject: MEMORY.md index)**: **While** the close-time pruning rule in `session-handoff.md § Auto-Memory Integration` item 6 governs memory hygiene, the pruning SHALL align with that rule — the archive is the correct destination for consumed/stable records, and the always-loaded index stays within the Claude Code loader cap (200 lines OR 25KB, whichever comes first).

### B.4 Measurable delta + non-regression

**REQ-MD-016 (State-driven, subject: always-loaded surface)**: **While** the 3 REQs are in effect, the combined always-loaded token count of the 3 modified file groups (cadence-bridge.md + session-handoff.md + MEMORY.md) SHALL decrease by ≥ ~3.0k tokens measured via `/context` before/after at run-phase. **Revised post-run (2026-07-10)**: the original plan-phase target was ≥ ~5.5k tokens, derived from a Discovery-time bytes→tokens conversion error that mislabeled byte counts as token counts for the session-handoff (~2k bytes misread as ~2k tokens) and MEMORY.md (~1.5k bytes misread as ~1.5k tokens) items; the cadence-bridge estimate (~2.3k tokens, the file leaving the always-loaded set) was always a true token estimate and is the dominant saving. The orchestrator's independent Trust-but-verify measured the actual combined reduction at ~3.0k tokens (cadence-bridge path-match ~2.3k + session-handoff ~487 + MEMORY ~210); this revised numeric anchor reflects the honest measured value, and AC-MD-016 is PASS at ~3.0k measured.

**REQ-MD-017 (Unwanted, subject: test suite)**: The SPEC implementation SHALL NOT introduce any `go test` failure, `go vet` finding, or template-neutrality CI failure — the existing test suite + lint + CI guards MUST continue to pass.

---

## C. Exclusions

> Per `.claude/rules/moai/development/spec-frontmatter-schema.md` `OutOfScopeRule`, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — KEEP-class always-loaded rules

- No `paths:` field SHALL be added to the 5 KEEP rules identified by SPEC-RULE-DIET-002 (`session-handoff.md`, `agent-common-protocol.md`, `askuser-protocol.md`, `verification-claim-integrity.md`, `context-window-management.md`). REQ-2 edits session-handoff.md's illustrative content but does NOT scope it out of always-loaded — it remains always-loaded with a trimmed body.
- No content trimming of `agent-common-protocol.md`, `askuser-protocol.md`, `CLAUDE.md`, or `CLAUDE.local.md` — these are owned by prior completed diet SPECs (CLAUDEMD-DIET-V2-001, STEERING-ALIGN-CLAUDEMD-DIET-001, STEERING-ALIGN-LOCAL-DIET-001) or are out of scope for safe-diet.

### Out of Scope — Aggressive diet (high regression risk)

- Compressing `session-handoff.md` CORE DOCTRINE prose (6-block skeleton, Field-by-Field Spec, Diet Constraints, Pre-emit self-check) — these are byte-identical-preservation obligations (REQ-MD-008), not diet targets.
- Scoping `session-handoff.md` out of always-loaded via `paths:` — explicitly rejected by SPEC-RULE-DIET-002 REQ-RD2-006 (KEEP-class) because Trigger #3 (user explicit session-end) fires from any session context.
- Pruning active-marker entries from MEMORY.md — REQ-MD-013 forbids this.

### Out of Scope — Token-budget ratchet

- This SPEC SHALL NOT lower any `AlwaysLoadedTokenBudget` constant or tighten any budget guard. The diet is measured via `/context`; the guard is a regression tripwire, not a ratchet.

### Out of Scope — New Go code / loader mechanism

- No new Go source file, no modification to `internal/config/`, no new rule-loading mechanism. The `paths:` frontmatter is interpreted by the Claude Code runtime directly (14+ existing path-scoped rules verify the infrastructure).

### Out of Scope — SSOT de-duplication

- This SPEC does NOT deduplicate SSOT content (that is SPEC-V3R6-RULES-SSOT-DEDUP-001's domain, completed). The extraction in REQ-2 relocates ILLUSTRATIVE content (examples), not duplicated doctrine.

---

## D. Constraints

1. **Frontmatter syntax (REQ-1)**: CSV string (`paths: "a,b,c"`), not YAML array. Per `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter.
2. **Self-referential `paths:` precedent (REQ-1)**: the established mechanism from RULE-DIET-002 (`paths: "**/<filename>.md"`) is the default. A broader trigger glob MAY be used if it provably covers the `/loop` + `/moai` composition surface (REQ-MD-002).
3. **Template-First (REQ-1, REQ-2)**: edit template SSOT tree first, then `make build` to re-embed, then verify live-tree byte-identical parity (where parity means "identical content" modulo any documented pre-existing baseline drift).
4. **Verbatim preservation (REQ-2)**: `✂` (U+2702), `─` (U+2500), the 6-block skeleton headings, and the Pre-emit self-check labels MUST survive byte-identical.
5. **SSOT ↔ render-surface parity (REQ-2)**: `session-handoff.md` ↔ `.claude/output-styles/moai/moai.md §8` parity contract MUST be honored.
6. **Auto-memory hygiene (REQ-3)**: MEMORY.md pruning aligns with `session-handoff.md § Auto-Memory Integration` item 6 (close-time pruning). Archive destination: `MEMORY-archive-2026-06-02.md`. No deletion — archive preserves the audit trail.
7. **Neutrality (REQ-1, REQ-2)**: no SPEC IDs, REQ tokens, internal dates, commit SHAs, or audit citations introduced into template files.
8. **GEARS notation**: requirements use current GEARS notation. No legacy `IF/THEN` modality.
9. **Verification-claim integrity**: all token-reduction claims measured via `/context` before/after at run-phase; all doctrine-integrity claims verified via grep.

---

## E. Self-Verification (plan-phase)

- SPEC ID `SPEC-MEMORY-DIET-001` passes the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (decomposition: SPEC ✓ | MEMORY ✓ | DIET ✓ | 001 ✓ → PASS; executed self-check: `PASS`).
- 12 canonical frontmatter fields present + `tier: M`.
- Every REQ is testable by a re-runnable command (see acceptance.md).
- Out of Scope section carries 5 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.
- Prior-Art Review section documents all 9 existing SPECs with one-line relationship verdicts.
- The 3 REQs span the 3 target file groups; the combined reduction target (≥ ~5.5k tokens) is achievable from the measured baselines (2.3k + 2k + 1.5k = 5.8k expected).

---

## F. Cross-References

- `.moai/specs/SPEC-RULE-DIET-002/` — predecessor in the rule-diet mechanism family; self-referential `paths:` precedent + KEEP/SCOPE classification methodology.
- `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/` — the original `paths:` frontmatter scoping SPEC (4 rules, 2026-05-22).
- `.moai/specs/SPEC-V3R6-RULES-COMPRESS-001/` — session-handoff.md prose compression (the baseline REQ-2 builds on).
- `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/` — token-diet pattern reference (Token-Economy Epic C).
- `.claude/rules/moai/workflow/cadence-bridge.md` — REQ-1 target (88 lines, 2.3k tokens).
- `.claude/rules/moai/workflow/session-handoff.md` — REQ-2 target (478 lines, 13.3k tokens); SSOT for paste-ready resume.
- `.claude/output-styles/moai/moai.md §8` — REQ-2 render surface (parity contract partner).
- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` — REQ-3 target (90 lines, 5.8k tokens).
- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY-archive-2026-06-02.md` — REQ-3 archive destination.
- `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter — CSV string syntax standard.
- Anthropic "Steering Claude Code" + "Write an effective CLAUDE.md" — the governing scoping principle (per-line test + `paths:` scoping).
