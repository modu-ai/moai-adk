---
id: SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001
title: "Steering-Align: CLAUDE.md always-loaded body diet (per-line-test deletion + rule-SSOT pointer-ization + changelog-footer removal + @import token-neutrality honesty)"
version: "0.1.1"
status: completed
created: 2026-06-22
updated: 2026-06-22
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "CLAUDE.md + internal/template/templates/CLAUDE.md"
lifecycle: spec-anchored
tags: "steering, claude-md, diet, always-loaded, context-budget, template-first, per-line-test"
tier: M
era: V3R6
---

## HISTORY

- 2026-06-22 — v0.1.0 — manager-spec — Plan-phase artifacts authored (Tier M, Section A-H, 4 artifacts: spec.md + plan.md + acceptance.md + progress.md). SPEC 2 of 5 in Epic Steering-Align. `status: draft`.
- 2026-06-22 — v0.1.1 — manager-spec — iter-2 audit revision (plan-auditor PASS-WITH-DEBT 0.83, Tier M thresh 0.80). Applied D1-D6 to convert to clean PASS; monotonic improvement (no regression of what passed). **D1 (SHOULD-FIX, top priority)**: independently re-verified that §16 Context-Search (`150,000` threshold → 0 SSOT hits), §15 CG-Mode (`moai cg` → 0 hits in spec-workflow.md), §11 Resumable-Agents (`agentId` → 0 hits) are GENUINELY UNIQUE; reclassified these + §11 Error-Recovery bullets + §14 File-Write-Conflict/Loop-Prevention/Background-Agent bullets from POINTER → KEEP (5 demotions — pointer-izing would DELETE behavioral content). Strengthened plan.md §C.2 to the prose-duplication bar (heading-presence ≠ duplication); encoded as run-phase precondition AC-CMD-009 (re-grep distinctive content, 0 hits → block POINTER + force KEEP). §14/§15 SPLIT (only the Worktree-Isolation-subsection / non-CG-team-prose are genuinely duplicated → POINTER). Derived target revised UP 350-430 → **400-470** (less content removed). **D2 (SHOULD-FIX)**: added AC-CMD-010 guarding the untagged-but-behavioral §4 anchors (8 retained-agent names, archived-agent list, Selection Decision Tree) — the 14-line [HARD] count does not cover them. **D3 (SHOULD-FIX)**: aligned AC-CMD-007 to the CI guard `TestTemplateNeutralityAudit` (`internal/template/template_neutrality_audit_test.go`) — the iter-1 broad grep false-failed on the pre-existing neutral-baseline L459 `feedback_worktree_autonomous` line; CI guard's allow-list is authoritative. **D4 (SHOULD-FIX)**: demoted REQ-CMD-008/AC-CMD-008 epistemic framing from VERIFIED-FACT → residual-risk; run-phase MUST WebFetch RE-confirm + reconcile with conflicting `agent-authoring.md:L100` (blocker if genuine conflict); AC-CMD-008 → SHOULD. **D5 (MINOR)**: corrected @import-SSOT attribution `## Paths Frontmatter` → `## File Size Limits` (L33, verified). **D6 (MINOR)**: tidied AC-CMD-001 `L==T` redundancy note (line-equality is the line-level cross-check; AC-CMD-002 `diff` is the authoritative byte-parity gate). All re-verified by command in §F.1 / acceptance.md. `status: draft`.

---

## A. Context / Background

### A.1 The doctrine being applied

Two Anthropic official documents anchor this SPEC:

1. **Blog — "Steering Claude Code: skills, hooks, rules, subagents and more"**. The blog distinguishes ALWAYS-LOADED steering surfaces (CLAUDE.md + unscoped rules + output-style) from CONDITIONALLY-loaded surfaces (path-scoped rules, on-demand skills). The entry SPEC of this Epic (SPEC-STEERING-ALIGN-RULE-SCOPING-001, COMPLETED) applied the blog's `paths:`-scoping guidance to the RULE layer. THIS SPEC applies the complementary best-practices per-line test to the CLAUDE.md always-loaded BODY.

2. **best-practices — "Write an effective CLAUDE.md"**. Two canonical statements:
   - The per-line test: *"Would removing this cause Claude to make mistakes? If not, cut it."*
   - The bloat warning: *"Bloated [always-loaded instructions] cause Claude to ignore your actual instructions."*

The per-line test is the SINGLE decision rule of this SPEC: every line of CLAUDE.md's body is classified by whether its removal would cause Claude to make mistakes.

### A.2 Observed ground-truth (re-verified live; commands + output in §F.1)

| Metric | LIVE (`CLAUDE.md`) | TEMPLATE (`internal/template/templates/CLAUDE.md`) |
|--------|--------------------|----------------------------------------------------|
| Line count | **650** | **650** |
| Byte count | **35778** | **35778** |
| `[HARD]` directive lines | **14** | **14** (byte-identical) |
| `[ZONE:*]` tagged lines | **14** | **14** (byte-identical) |
| `@import` lines | **2** (`@.moai/config/sections/user.yaml`, `@.moai/config/sections/language.yaml`) | **2** (identical) |
| Section structure | §1..§17 + Version/Changes footer | identical |

`diff CLAUDE.md internal/template/templates/CLAUDE.md` → **exit 0 (byte-identical)**. The two trees are an EXACT mirror (unlike the RULE-SCOPING-001 case where the template was a deliberate subset). This SPEC MUST keep them in byte-parity: both trees diet IDENTICALLY.

The MoAI CI size heuristic is **40,000 chars** (`.claude/rules/moai/development/coding-standards.md` § File Size Limits); the official Claude Code target is **"under 200 lines per CLAUDE.md"**. The current 650 lines is **3.25× the official target**; 35778 chars is **~89% of the CI cap** (close to the ceiling).

### A.3 The B-CRITICAL load-bearing design constraint — `@import` token-neutrality

[HARD] Claude Code's CLAUDE.md `@path/to/file` import INLINE-EXPANDS the referenced file into the always-loaded context at load time (up to 5-hop), per `.claude/rules/moai/development/coding-standards.md` § File Size Limits (the statement lives at L33, under the `## File Size Limits` H2 — NOT under `## Paths Frontmatter`; D5 attribution fix): *"`@import` (`@path/to/file`) does NOT reduce context — imported files are expanded and loaded in full at launch. … Use it for organization, never for size reduction."*

Therefore splitting 650 lines into N `@import`-ed files that ALL remain always-loaded yields **IDENTICAL always-on token cost**. `@import` is a STRUCTURE / readability mechanism, NOT a token-reduction mechanism. The paste-ready hint "@import로 root 인덱스화" MUST NOT be counted toward always-on token reduction.

REAL always-on token reduction comes from exactly **three** mechanisms. Every diet action in this SPEC MUST be classified by which one it uses:

| Mechanism | What it does | Token effect |
|-----------|--------------|--------------|
| **M-DELETE** (per-line-test deletion) | Cut content that fails the per-line test — removal does NOT cause Claude to make mistakes (e.g. the changelog footer — git is its SSOT) | REAL reduction |
| **M-POINTER** (rule-SSOT pointer-ization) | Replace a duplicated EXPLANATION whose SSOT already lives in `.claude/rules/...` with a one-line cross-reference, PRESERVING any [HARD] directive unique to CLAUDE.md | REAL reduction (the prose shrinks; the SSOT is loaded conditionally or on its own always-load budget already accounted separately) |
| **M-SCOPE** (move-to-paths-scoped-rule) | Relocate a self-contained body block OUT to a new/existing paths-scoped rule (conditional load) | REAL reduction (block leaves the always-load set) |

`@import` is explicitly NOT a fourth mechanism. If `@import` is used at all, it MUST be justified as structure-only and MUST NOT be counted toward token savings. This honesty constraint is REQ-CMD-004 and AC-CMD-004.

### A.4 The §4 nesting-note content-correctness sub-item (separate from token diet)

CLAUDE.md §4 carries a "Watch (Claude Code 2.1.172)" note stating a subagent can "opt into spawning its own subagents up to 5 levels deep … disabled by default." A plan-phase WebFetch of the live official sub-agents doc (`code.claude.com/docs/en/sub-agents`, fetched 2026-06-22) returned text describing the mechanism as: a subagent spawns nested subagents when `Agent` is in its `tools` list; to PREVENT it, omit `Agent`; the depth limit appears **fixed at 5 and not configurable** ("A subagent at depth five does not receive the Agent tool and cannot spawn further").

**(D4 epistemic demotion)** This plan-phase WebFetch reading is NOT asserted as VERIFIED FACT — it is a single un-archived WebFetch with no reproducible artifact, and a CONFLICTING flat-hierarchy framing exists at `.claude/rules/moai/development/agent-authoring.md:L100`. The §4 note's "opt-in / disabled by default" framing MAY be imprecise, but the precise current CC behavior is a **residual risk requiring run-phase WebFetch RE-confirmation** (§F.3). The run-phase MUST: (1) re-fetch the official doc to re-confirm the `Agent`-in-`tools` / fixed-depth-5 mechanism, AND (2) reconcile the §4 note with `agent-authoring.md:L100` rather than asserting one source over the other (if the two genuinely conflict, return a blocker — do NOT silently pick a winner). This is a CONTENT-CORRECTNESS item, scoped as a DISCRETE REQ (REQ-CMD-008) so it is not lost inside the token-diet work. The run-phase decides the precise corrected wording AFTER the re-confirmation + reconciliation.

### A.5 Epic Steering-Align context

This is **SPEC 2 of 5** in Epic Steering-Align — a roadmap aligning moai-adk to Anthropic's official Claude Code "steering" guidance. The roadmap (named here for Epic context ONLY — this SPEC authors artifacts for the CLAUDE.md diet alone):

1. **SPEC-STEERING-ALIGN-RULE-SCOPING-001** (COMPLETED, origin ab81e7f42) — path-scope file-touch-triggered always-loaded rules.
2. **SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001** (THIS SPEC) — apply the per-line test to CLAUDE.md's always-loaded body.
3. (FUTURE) GUARDRAIL-HOOK (P3) — convert intent-triggered orchestration prose into deterministic guardrail hooks.
4. (FUTURE) OUTPUT-STYLE-SLIM (P5) — trim the always-loaded output-style body.
5. (FUTURE) LOCAL-DIET (P6) — apply the per-line test to the maintainer-local `CLAUDE.local.md` always-loaded body.

The paste-ready's "P1+P4+P7" shorthand does NOT map to concretely-defined RULE-SCOPING items; this SPEC therefore defines the DIET mechanisms FUNCTIONALLY (M-DELETE / M-POINTER / M-SCOPE per §A.3), not by P-number.

---

## B. Requirements (GEARS notation)

### B.1 Per-line-test body diet (M-DELETE)

- **REQ-CMD-001 (Ubiquitous)**: The CLAUDE.md body diet SHALL apply the best-practices per-line test — *"Would removing this cause Claude to make mistakes? If not, cut it."* — to every section of the always-loaded body, classifying each block as KEEP (removal causes mistakes → retain), CUT (removal causes no mistakes → M-DELETE), or POINTER (the explanation is duplicated from a `.claude/rules/...` SSOT → M-POINTER). The per-section classification table is the core plan.md deliverable; run-phase execution SHALL be mechanical from it.

- **REQ-CMD-002 (Ubiquitous)**: The Version / "Changes in vX.Y.Z" changelog footer (CLAUDE.md L634-650) SHALL be reduced: the `Version:` / `Language:` / `Core Rule:` identity lines MAY be retained (a system identifier is neutral per §D constraint), but the "Changes in v14.2.0 …" / "Changes in v14.1.0 …" per-line changelog prose SHALL be removed (M-DELETE). Rationale (per-line test): a per-version changelog narrative in an always-loaded file fails the test — git history (`git log`, CHANGELOG.md) is its SSOT; its removal does NOT cause Claude to make mistakes.

### B.2 Rule-SSOT pointer-ization (M-POINTER)

- **REQ-CMD-003 (Ubiquitous)**: For each CLAUDE.md section whose EXPLANATION duplicates content whose single-source-of-truth already lives in a verified `.claude/rules/...` file, the change SHALL replace the duplicated explanation with a one-line cross-reference pointer while PRESERVING every [HARD] directive that is unique to CLAUDE.md. The candidate pointer sections and their verified SSOTs are enumerated in plan.md §C (each SSOT tool-verified to exist + carry the content before pointer-ization is proposed, per verification-claim-integrity.md).

- **REQ-CMD-006 (State-driven)**: **While** a section is pointer-ized, any [HARD]/[ZONE:*]-tagged directive line that is UNIQUE to CLAUDE.md (i.e. not also present verbatim in the SSOT) SHALL be preserved in CLAUDE.md. Pointer-ization collapses surrounding explanatory prose, NEVER a binding directive. The set of load-bearing [HARD] lines that MUST survive is enumerated in plan.md §D.

### B.3 @import token-neutrality honesty (the load-bearing constraint)

- **REQ-CMD-004 (Unwanted behavior)**: The change SHALL NOT count any `@import` restructuring toward always-on token reduction. `@import` inline-expands its target into always-loaded context at load time, so `@import`-ing a file that remains always-loaded yields identical token cost. If `@import` is used at all in the diet, it SHALL be justified in plan.md §D as structure-only, and the token-reduction accounting (acceptance.md AC-CMD-006) SHALL exclude it. The SPEC MUST NOT pretend `@import` reduces always-on tokens.

### B.4 Template-mirror parity + neutrality preservation

- **REQ-CMD-005 (Ubiquitous)**: The diet SHALL be applied to BOTH trees identically (`CLAUDE.md` AND `internal/template/templates/CLAUDE.md`), keeping them byte-identical. Per the CLAUDE.local.md §2 Template-First rule, the template SSOT tree is edited FIRST, then re-embedded via `make build`, then live-tree parity is verified (`diff` → exit 0). Both trees are confirmed byte-identical at baseline (§F.1).

- **REQ-CMD-007 (Unwanted behavior)**: The diet SHALL NOT inject any moai-adk internal-development artifact into the CLAUDE.md body — no internal SPEC IDs, no internal work dates, no commit SHAs, no audit citations ("Audit N Finding AX"), no `feedback_*` / memory paths, no `/Users/` OS-bias paths (per CLAUDE.local.md §15 language-neutrality + §25 internal-content isolation; CI guard `internal/template/internal_content_leak_test.go`). CLAUDE.md is a 16-language-neutral distributed asset. (The existing `Version: 14.2.0` system identifier is neutral and may remain.)

### B.5 §4 nesting-note content correctness (discrete, separate from token diet)

- **REQ-CMD-008 (Ubiquitous)**: The CLAUDE.md §4 "Watch (Claude Code 2.1.172)" nesting note SHALL be reviewed for content accuracy and corrected if imprecise. **(D4 — the precise CC mechanism is a residual risk, NOT an asserted fact at plan-phase)**: the run-phase SHALL (1) **re-fetch** the live official sub-agents doc to re-confirm the candidate mechanism (a subagent spawns nested subagents when `Agent` is in its `tools` list; depth fixed at 5 and not configurable), AND (2) **reconcile** the §4 note with the conflicting flat-hierarchy framing at `agent-authoring.md:L100` — if the two genuinely conflict, the run-phase returns a blocker rather than silently choosing a winner. The note's "opt-in / disabled by default" framing SHALL be corrected ONLY after re-confirmation + reconciliation. This REQ is a content-correctness fix, NOT a token-diet action; it MAY be net-neutral or slightly additive in length, and that is acceptable.

### B.6 Derived target (range, not a hard number)

- **REQ-CMD-009 (Ubiquitous)**: The final CLAUDE.md line count SHALL be a RANGE derived from the KEEP/CUT/POINTER classification (plan.md §C), NOT an arbitrary fixed number. The paste-ready's "~300 lines" is GUIDANCE only. Behavioral KEEP content SHALL NOT be cut to hit a number; an honest diet that lands at 350-450 lines is acceptable if the classification justifies it. The derived target + its justification SHALL be stated in plan.md §C.

---

## C. Constraints

- **C-1** [HARD] Template-First (CLAUDE.local.md §2): edit `internal/template/templates/CLAUDE.md` FIRST, then `make build` to re-embed, then verify live-tree byte-parity (`diff` exit 0). BOTH trees diet IDENTICALLY (REQ-CMD-005).
- **C-2** [HARD] BODY editing (NOT frontmatter-only like RULE-SCOPING-001) → strictly higher risk. The plan.md KEEP/CUT/POINTER map MUST be precise enough that run-phase is purely mechanical — no run-phase judgment call about whether a line is behavioral.
- **C-3** [HARD] [HARD]/[ZONE:*]-tagged directives carry behavioral value (per-line test = "YES, removing causes mistakes") → KEEP. Pointer-ization may collapse the surrounding EXPLANATION but MUST preserve the binding directive (REQ-CMD-006). plan.md §D enumerates the load-bearing [HARD] lines.
- **C-4** [HARD] Neutrality / isolation (CLAUDE.local.md §15 + §25): no internal SPEC IDs / dates / SHAs / audit citations / memory paths injected (REQ-CMD-007).
- **C-5** [HARD] `@import` token-neutrality honesty (REQ-CMD-004): never count `@import` as a token reduction.
- **C-6** No Go code change, no new lint rule. This SPEC is markdown body editing + `make build` re-embed only.
- **C-7** A candidate POINTER section is proposed ONLY after its SSOT rule file is tool-verified to EXIST and CARRY the content (grep, per verification-claim-integrity.md §1.1 — a "this is duplicated" claim is a defect claim that MUST be tool-verified, never assumed). plan.md §C records the verification command + observed output per candidate.

---

## D. Out of Scope / Exclusions

The SPEC scope is exactly: per-line-test body diet (M-DELETE) + rule-SSOT pointer-ization (M-POINTER) + changelog-footer removal + @import-honesty + §4 nesting-note correctness, applied to BOTH CLAUDE.md trees in parity. The following are explicitly excluded.

### Out of Scope — @import as a token-reduction mechanism

- Restructuring CLAUDE.md into `@import`-ed sub-files and CLAIMING a token saving. `@import` inline-expands at load (§A.3); it is structure-only. Any `@import` use in this SPEC is justified as readability, never counted as token reduction. This is the load-bearing exclusion — the SPEC MUST NOT pretend `@import` reduces always-on tokens.

### Out of Scope — the other 4 Epic Steering-Align SPECs

- GUARDRAIL-HOOK (P3), OUTPUT-STYLE-SLIM (P5), LOCAL-DIET (P6), and the RULE-SCOPING entry SPEC (already COMPLETED) — named in §A.5 for Epic context only. This SPEC authors artifacts ONLY for the CLAUDE.md diet. In particular, `CLAUDE.local.md` (the maintainer-local always-loaded body) is OUT OF SCOPE here — it is the FUTURE LOCAL-DIET SPEC.

### Out of Scope — moai-adk internal-content injection

- Injecting internal SPEC IDs, internal work dates, commit SHAs, audit citations, or `feedback_*`/memory paths into the CLAUDE.md body (forbidden by REQ-CMD-007 / CLAUDE.local.md §25). The diet REMOVES bloat; it MUST NOT ADD internal-development noise.

### Out of Scope — behavioral content removal to hit a line-number target

- Cutting any KEEP-classified ([HARD]/[ZONE:*] or otherwise behavioral) content purely to reach "~300 lines". The "~300 lines" figure is guidance; the real target is a range derived from the classification (REQ-CMD-009). Over-cutting behavioral content to hit a number is forbidden.

### Out of Scope — Go code / lint / SSOT-rule edits

- Go code changes, new lint rules, and edits to the `.claude/rules/...` SSOT files themselves. Pointer-ization (M-POINTER) only edits CLAUDE.md to point AT an existing SSOT; it does NOT modify the SSOT. (If a pointer-ization reveals a genuine SSOT gap, run-phase returns a blocker — it does not silently edit the SSOT under this SPEC.)

---

## E. Acceptance Criteria Reference

Concrete GEARS-format acceptance criteria with re-runnable verification commands live in `acceptance.md`. Summary of AC themes:

- **AC-CMD-001** — per-tree line-count drop measurable by `wc -l`, LIVE and TEMPLATE each dropping into the derived range (plan.md §C); both trees drop by the SAME amount.
- **AC-CMD-002** — template ↔ live byte-parity preserved (`diff CLAUDE.md internal/template/templates/CLAUDE.md` → exit 0).
- **AC-CMD-003** — load-bearing [HARD]-directive count NOT reduced below the enumerated load-bearing set (`grep -c '\[HARD\]'` before vs after — the protected count MUST survive; only redundant duplicate-of-SSOT [HARD] restatements, if any, may drop, and each such drop is enumerated + justified in plan.md §D).
- **AC-CMD-004** — `@import` not double-counted: the 2 `@import` lines (§9) are accounted as structure-only; the token-reduction figure (AC-CMD-006) excludes any `@import` restructuring.
- **AC-CMD-005** — changelog footer removed: `grep -c '^Changes in v' CLAUDE.md` → 0 in both trees; identity lines (`Version:` etc.) optionally retained.
- **AC-CMD-006** — always-on char/byte-sum reduced per tree (`wc -c`), LIVE and TEMPLATE separately, the reduction attributable solely to M-DELETE + M-POINTER + M-SCOPE (NOT @import).
- **AC-CMD-007** — neutrality preserved, CI-guard-aligned (D3): the authoritative `TestTemplateNeutralityAudit` in `internal/template/template_neutrality_audit_test.go` STILL passes post-diet (no NEW internal-artifact leak); pre-existing neutral-baseline content (e.g. the L459 `feedback_worktree_autonomous` reference) is NOT a regression.
- **AC-CMD-008** — §4 nesting note reconciled (SHOULD, D4): run-phase re-fetches the official doc + reconciles with `agent-authoring.md:L100` BEFORE asserting the `Agent`-in-`tools` / fixed-depth-5 mechanism.
- **AC-CMD-009** (D1, MUST) — POINTER edits gated on prose-duplication re-verification: each surviving POINTER section must show ≥1 distinctive-content SSOT hit before its edit; a 0-hit section is reclassified KEEP (the core defense against the iter-1 over-cut).
- **AC-CMD-010** (D2, MUST) — behavioral-but-untagged §4 anchors survive: the 8 retained-agent names, the archived-agent list, and the Selection Decision Tree must not be dropped (the [HARD]-count guard does not cover these).

---

## F. Evidence (verification-claim-integrity)

### F.1 Re-verified ground-truth (command → observed output, this tree, 2026-06-22)

| Claim | Command | Observed |
|-------|---------|----------|
| LIVE 650 lines | `wc -l CLAUDE.md` | `650` |
| TEMPLATE 650 lines | `wc -l internal/template/templates/CLAUDE.md` | `650` |
| Byte-identical trees | `diff CLAUDE.md internal/template/templates/CLAUDE.md; echo $?` | `0` (no diff) |
| LIVE 35778 bytes | `wc -c CLAUDE.md` | `35778` |
| `[HARD]` lines | `grep -c '\[HARD\]' CLAUDE.md` | `14` |
| `[ZONE:*]` lines | `grep -c '\[ZONE:' CLAUDE.md` | `14` |
| `@import` lines | `grep -nE '^@[.]' CLAUDE.md` | 2 lines: L331 `@.moai/config/sections/user.yaml`, L332 `@.moai/config/sections/language.yaml` |
| Section structure | `grep -nE '^## ' CLAUDE.md` | §1..§17 at L3,31,74,96,147,188,217,307,327,364,392,412,426,446,470,544,593 + footer L634 |
| Changelog footer | `grep -nE '^Version:\|^Changes in v' CLAUDE.md` | `634:Version: 14.2.0 ...`, `638:Changes in v14.2.0 ...`, `644:Changes in v14.1.0 ...` |
| §7 POINTER prose-dup (D1 bar) | `grep -c 'Ambiguous pronouns\|Multi-interpretable\|Unclear boundaries\|conflict with existing' .../askuser-protocol.md` | `3`+ (all 4 triggers present) → **POINTER confirmed** |
| §8 POINTER prose-dup (D1 bar) | `grep -c 'select:AskUserQuestion\|max 4 questions\|First option MUST' .../askuser-protocol.md` | `9` → **POINTER confirmed** |
| §10 POINTER prose-dup (D1 bar) | `grep -c 'WebSearch\|WebFetch\|verify each URL\|Sources:' .../moai-constitution.md` | `3` → **POINTER confirmed** |
| §11 prose-dup (D1 bar) | `grep -c 'Token limit\|Permission error' .../agent-common-protocol.md`; `grep -c 'agentId\|Resumable' .../agent-common-protocol.md` | **`0` / `0`** → **DEMOTE to KEEP** (heading present but DISTINCTIVE prose ABSENT) |
| §13 POINTER prose-dup (D1 bar) | `grep -c 'Level 1\|Level 2\|67%\|skillListingBudget' .../skill-authoring.md` | `5` → **POINTER confirmed** |
| §14 SPLIT prose-dup (D1 bar) | `grep -c 'File Write Conflict\|Loop Prevention\|Background.*Agent' .../moai-constitution.md`; `grep -c 'Worktree Isolation\|isolation.*worktree' .../worktree-integration.md` | **`0`** / `32` → **operational bullets KEEP; Worktree-subsection POINTER** |
| §15 SPLIT prose-dup (D1 bar) | `grep -c 'CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS\|SendMessage\|TeammateIdle' .../spec-workflow.md`; `grep -c 'moai cg' .../spec-workflow.md` | `5` / **`0`** → **non-CG team prose POINTER; CG-Mode KEEP** |
| §16 prose-dup (D1 bar) | `grep -c '150,000\|150000' .../context-window-management.md`; `grep -c 'Search Process\|previous session\|injection' .../context-window-management.md` | **`0`** / `1` weak → **DEMOTE to KEEP** (threshold + process both UNIQUE; SSOT covers `/clear` thresholds NOT search protocol) |
| §4 nesting (D4 — un-archived single fetch, NOT asserted fact) | `code.claude.com/docs/en/sub-agents` WebFetch 2026-06-22 (no reproducible artifact) | reading: "a subagent can spawn its own subagents" (v2.1.172); "A subagent at depth five does not receive the Agent tool and cannot spawn further. The limit is fixed and not configurable"; "To prevent … omit `Agent` from its `tools` list". **CONFLICTS with `agent-authoring.md:L100` flat-hierarchy framing → run-phase RE-confirm + reconcile (REQ-CMD-008, §F.3 residual risk)** |
| `@import` non-reduction (D5 attribution fix) | `grep -n 'import.*does NOT reduce\|Use it for organization' .claude/rules/moai/development/coding-standards.md` | `33:` — under the **`## File Size Limits`** H2 (L23), **NOT** under `## Paths Frontmatter` (L93). "`@import` … does NOT reduce context — imported files are expanded and loaded in full at launch. Use it for organization, never for size reduction" |
| CI size heuristic | `grep -n '40,000 characters\|200 lines' .claude/rules/moai/development/coding-standards.md` | `25:CLAUDE.md should stay under 40,000 characters … "under 200 lines per CLAUDE.md"` |

### F.2 Gaps (explicitly NOT observed at plan-phase)

- The exact final line-count is NOT fixed at plan-phase — it is a RANGE derived from the KEEP/CUT/POINTER classification (REQ-CMD-009); the precise number is a run-phase outcome.
- The token reduction is APPROXIMATED via line/byte-sum, not a real tokenizer count (acceptable per the per-line-test framing; exact token count deferred to run-phase if needed).
- The PRECISE corrected wording of the §4 nesting note (REQ-CMD-008) is NOT authored at plan-phase — only the candidate mechanism is identified; the run-phase RE-confirms via WebFetch + reconciles with `agent-authoring.md:L100` before authoring corrected text (D4).
- Whether any [HARD] line is a pure duplicate of an SSOT [HARD] line (thus droppable under M-POINTER without losing a directive) is assessed per-line in plan.md §D; the plan-phase default is to PRESERVE every [HARD] line unless the duplicate is explicitly proven.
- **(D1)** The prose-duplication re-audit (plan.md §C.2) used `grep` of the DISTINCTIVE content per section. This is stronger than iter-1's heading-presence check but is still a plan-phase grep snapshot — the run-phase MUST re-run the prose-duplication grep (AC-CMD-009) before each POINTER edit, since the SSOT files could change between plan and run.

### F.3 Residual risk

- BODY editing is higher-risk than the frontmatter-only RULE-SCOPING-001. Mitigated by the precise plan.md §C KEEP/CUT/POINTER map (run-phase is mechanical) + the AC-CMD-003 [HARD]-count guard (load-bearing directives cannot silently drop).
- Over-aggressive pointer-ization could remove a directive that LOOKS duplicated but carries a CLAUDE.md-unique nuance. **This exact risk materialized in iter-1 (D1): 5 sections (§11 recovery, §11 resumable, §14 operational bullets, §15 CG-Mode, §16 search protocol) were classified POINTER on heading-presence alone, but their distinctive content is UNIQUE (0 SSOT prose hits) — pointer-izing would have DELETED behavioral content.** Mitigated by the STRONGER prose-duplication bar (plan.md §C.2) + the run-phase precondition AC-CMD-009 (re-grep distinctive content; 0 hits → block POINTER, force KEEP) + REQ-CMD-006 (preserve unique directives).
- **(D4) §4 nesting-note CC behavior is a residual risk, not a verified fact.** The plan-phase WebFetch reading (`Agent`-in-`tools` mechanism, fixed depth-5) is a single un-archived fetch and CONFLICTS with `agent-authoring.md:L100`'s flat-hierarchy framing. The run-phase MUST re-fetch + reconcile (REQ-CMD-008); if the conflict is genuine, return a blocker rather than asserting one source. AC-CMD-008 is therefore SHOULD (non-blocking) — it confirms the mechanism is NAMED, not that a specific assertion is true.
- `make build` re-embed could surface unrelated template drift. Mitigated by scoping the run-phase diff to CLAUDE.md only and asserting byte-parity (AC-CMD-002).
- A pointer-ization that points at an SSOT which is itself path-scoped (conditionally loaded) means the explanation is no longer always-available — but that is the INTENT (the SSOT loads when relevant). Risk only if a CLAUDE.md-unique always-needed directive is removed; guarded by REQ-CMD-006.

---

## G. SPEC ID Pre-Write Self-Check (recorded per protocol)

decomposition: SPEC ✓ | STEERING ✓ | ALIGN ✓ | CLAUDEMD ✓ | DIET ✓ | 001 ✓ → PASS

Canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`: first segment literal `SPEC`; middle segments STEERING / ALIGN / CLAUDEMD / DIET each match `[A-Z][A-Z0-9]*` (CLAUDEMD = uppercase alphanumeric, length ≥ 1, valid); last segment `001` matches `\d{3}` (digit-only, no alpha suffix). PASS.

---

## H. Cross-References

- Anthropic best-practices "Write an effective CLAUDE.md" (per-line test, bloat warning).
- Anthropic blog "Steering Claude Code: skills, hooks, rules, subagents and more" (always-loaded vs conditional steering surfaces).
- Anthropic sub-agents doc `code.claude.com/docs/en/sub-agents` (§4 nesting-note correctness source — `Agent`-in-`tools` mechanism, fixed depth-5 limit).
- `.claude/rules/moai/development/coding-standards.md` § File Size Limits (40K-char CI heuristic / 200-line official target) + § Paths Frontmatter (`@import` non-reduction note).
- CLAUDE.local.md §2 (Template-First), §15 (language-neutrality), §25 (internal-content isolation).
- SSOT rule files for the M-POINTER candidates (each verified in §F.1): `askuser-protocol.md`, `agent-common-protocol.md`, `moai-constitution.md`, `glm-web-tooling.md`, `skill-authoring.md`, `spec-workflow.md`, `context-window-management.md`.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec.
- SPEC-STEERING-ALIGN-RULE-SCOPING-001 (Epic entry SPEC, COMPLETED) — established the MIRRORED-vs-LIVE-ONLY split + per-line evidence discipline this SPEC mirrors.
