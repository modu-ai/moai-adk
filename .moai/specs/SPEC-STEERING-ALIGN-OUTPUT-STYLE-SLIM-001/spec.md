---
id: SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001
title: "Steering-Align: output-style moai.md always-loaded body diet (§8 Session-Handoff pointer-ization + Localization-Contract duplicate-prose deletion, MODERATE bound, render-SSOT preservation)"
version: "0.1.0"
status: draft
created: 2026-06-23
updated: 2026-06-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/output-styles/moai/"
lifecycle: spec-anchored
tags: "steering, output-style, diet, always-loaded, context-budget, template-first, pointer-ization, render-ssot"
tier: M
era: V3R6
---

## HISTORY

- 2026-06-23 — v0.1.0 — manager-spec — Plan-phase artifacts authored (Tier M, Section A-H, 4 artifacts: spec.md + plan.md + acceptance.md + progress.md skeleton). SPEC 4 of 5 in Epic Steering-Align (P5). Sibling P2 = CLAUDEMD-DIET-001 (COMPLETED, M-DELETE + M-POINTER + AC-009 over-cut gate); P3 = GUARDRAIL-HOOK-001 (COMPLETED, origin c463257b8). This SPEC applies the same per-line-test diet techniques to the always-loaded **output-style** body (`moai.md`), bounded MODERATE (user-confirmed) and gated by the render-SSOT preservation invariant. `status: draft`.

---

## A. Context / Background

### A.1 The doctrine being applied

This SPEC continues Epic Steering-Align — the roadmap aligning moai-adk to Anthropic's official Claude Code "steering" guidance. Two canonical statements anchor it (the same two that anchored the COMPLETED sibling P2 CLAUDEMD-DIET-001):

1. **best-practices — "Write an effective always-loaded instruction surface"**: the per-line test — *"Would removing this cause Claude to make mistakes? If not, cut it."* — and the bloat warning — *"Bloated always-loaded instructions cause Claude to ignore your actual instructions."*
2. **blog — "Steering Claude Code: skills, hooks, rules, subagents and more"**: distinguishes ALWAYS-LOADED steering surfaces (CLAUDE.md + unscoped rules + **output-style**) from CONDITIONALLY-loaded surfaces.

The output-style `moai.md` is an ALWAYS-LOADED surface: it is loaded into the active context at every session launch alongside CLAUDE.md. It is therefore a per-line-test diet target, exactly like the CLAUDE.md body that P2 dieted. This SPEC applies the SAME two diet mechanisms P2 validated — **M-DELETE** (per-line-test deletion) and **M-POINTER** (rule-SSOT pointer-ization) — to `moai.md`, bounded MODERATE.

### A.2 Observed ground-truth (re-verified live; commands + output in §F.1)

| Metric | LIVE (`.claude/output-styles/moai/moai.md`) | TEMPLATE (`internal/template/templates/.claude/output-styles/moai/moai.md`) |
|--------|----------------------------------------------|------------------------------------------------------------------------------|
| Line count | **782** | **782** |
| Byte count | **55306** | **55306** |
| Section structure | §1..§12 + frontmatter | identical |
| §8 Response Templates span | **L211-731 (~520 lines = 66% of the file)** | identical |

`diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md` → **exit 0 (byte-identical)**. The two trees are an EXACT mirror (like the P2 CLAUDE.md case). This SPEC MUST keep them in byte-parity: both trees diet IDENTICALLY (REQ-OSS-005).

§8 (Response Templates) is the primary diet target — it is 66% of the file. Within §8, two sub-areas concentrate the diet-able content:

- **§8 Localization Contract [HARD] (L213-331, ~118L)** — the localization translation obligation + an anti-pattern catalogue (label-level, L253-274) + a "Banner body prose Anti-pattern catalogue" (ko-canonical mapping tables, L299-318) + a Pre-emit self-check (L320-330).
- **§8 Session Handoff [HARD] (L648-731, ~84L)** — already carries an explicit `<!-- render-only, not canonical — canonical lives in .claude/rules/moai/workflow/session-handoff.md (SSOT) -->` marker (L652) → the highest-priority pointer-ize candidate.

### A.3 The load-bearing design constraint — render-SSOT preservation

[HARD] Not all of §8 is duplicated-from-elsewhere prose. A large portion of §8 IS the **render single-source-of-truth**: the orchestrator reads these tables AT OUTPUT TIME, and they have NO external owner. Cutting them would DELETE behavioral render content (the exact over-cut failure mode P2 D1 caught). The render-SSOT content that MUST be preserved verbatim:

| Render-SSOT content | Why it has no external owner | Location |
|---------------------|------------------------------|----------|
| **14 banner skeletons** (Task Start, Delegation Dispatch, Checkpoint Gate, Insight, Verification Matrix, Plan Audit, Discovery Report, Race Absorbed, Epic Stats, Epic Status, Completion Report, Error Recovery, Progress Board, Session Handoff) | The orchestrator renders these at output time; there is no other file that carries the banner skeletons | L332-731 (one per `### ` sub-heading) |
| **8 per-banner en/ko/ja/zh header-translation tables** | The orchestrator reads the locale rendering at output time per `conversation_language`; these ARE the translation SSOT | L370, L396, L427, L460, L495, L529, L563, L686 |
| **§8 Localization Contract ko-canonical mapping tables** (label→ko-canonical L253-274; banner-body-prose→ko-natural L299-318) | The orchestrator reads the ko-canonical mapping at render time to avoid the raw-English-literal HARD violation | L253-274, L299-318 |
| **§8 Session Handoff Cut-line Marker + Header translation tables** | Bound by the drift-mitigation parity sentinel (L653) which requires the 4-locale column count to STAY at the render surface for drift-visibility | L679, L686 |
| **§9 Language Rules [HARD] / §10 Output Rules [HARD]** | Distinct binding directives (verbatim-preserve symbol list, free-form-prohibition, preview-field standard) | L732-753 |
| **Verbatim-preserve emoji/box-drawing/symbol lists + `ultrathink.` keyword token** | The render contract's preserve-verbatim list IS the SSOT | L234-238, L736 |

This SPEC's diet REMOVES only content that genuinely has an EXTERNAL SSOT (so the §8 render surface points AT it rather than re-stating it). It NEVER removes render-SSOT content. This is the load-bearing invariant — it generalizes the P2 D1 over-cut defense (AC-CMD-009) to the output-style file.

### A.4 The two diet mechanisms (functional, mirrors P2 §A.3)

REAL always-on token reduction comes from exactly two mechanisms in this SPEC's MODERATE scope (M-SCOPE is NOT used — no new paths-scoped rule is created from an output-style):

| Mechanism | What it does | Token effect |
|-----------|--------------|--------------|
| **M-DELETE** (per-line-test deletion) | Cut content that fails the per-line test — removal does NOT cause Claude to make mistakes (e.g. duplicated explanatory prose whose meaning is fully carried by the surviving render skeleton + the external SSOT pointer) | REAL reduction |
| **M-POINTER** (rule-SSOT pointer-ization) | Replace a duplicated EXPLANATION whose SSOT already lives in `.claude/rules/...` with a one-line cross-reference, PRESERVING the render skeleton the orchestrator needs at output time | REAL reduction (the prose shrinks; the SSOT is already loaded on its own budget) |

The §8 Session Handoff block (L648-731) is the canonical M-POINTER candidate: it ALREADY self-declares render-only (L652) and points at `session-handoff.md` SSOT. The duplicated field-by-field narration, the auto-memory persistence procedure detail, and the anti-pattern catalogue inside that block are owned by `session-handoff.md` and can be condensed to the render skeleton + a pointer — WHILE preserving the cut-line marker spec, the translation tables, the drift sentinel, and the render skeleton the parity contract requires.

### A.5 Diet bound — MODERATE (user-confirmed this session)

The diet is bounded MODERATE (user-confirmed). The estimated reduction is **~150-250 lines (782 → ~530-630L)** — a GUIDANCE band, NOT a hard requirement. The preservation invariants (§A.3) WIN over the line-count target. This inherits the P2 lesson explicitly: **behavioral-PASS over numeric-proxy — do NOT over-cut render-SSOT content to hit a line number** (P2 AC-009 / AC-CMD-009 over-cut gate). The rejected "적극" (aggressive) option — a banner-template restructure that would re-shape the 14 banner skeletons — is OUT OF SCOPE (§D).

### A.6 Epic Steering-Align context

This is **SPEC 4 of 5** in Epic Steering-Align (P5). The roadmap (named here for Epic context ONLY — this SPEC authors artifacts for the output-style diet alone):

1. **SPEC-STEERING-ALIGN-RULE-SCOPING-001** (COMPLETED, origin ab81e7f42) — path-scope file-touch-triggered always-loaded rules.
2. **SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001** (COMPLETED, P2) — per-line-test diet of CLAUDE.md's always-loaded body (M-DELETE + M-POINTER + AC-009 over-cut gate).
3. **SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001** (COMPLETED, P3, origin c463257b8) — deterministic SessionStart guardrail hook + always-load removal.
4. **SPEC-STEERING-ALIGN-OUTPUT-STYLE-SLIM-001** (THIS SPEC, P5) — trim the always-loaded output-style body.
5. (FUTURE) **LOCAL-DIET** (P6) — apply the per-line test to the maintainer-local `CLAUDE.local.md` always-loaded body.

---

## B. Requirements (GEARS notation)

> **REQ numbering note**: REQ-OSS-004 was withdrawn during authoring (an earlier "@import token-neutrality honesty" REQ inherited from the P2 sibling template, which does NOT apply here — the output-style file has no `@import` directives, verified §F.1). The survivor REQ numbers (005-009) are RETAINED rather than renumbered to preserve AC-binding stability (the acceptance.md D.2 traceability table and the AC bodies bind to REQ-OSS-005/006/007/008/009 by number; renumbering would cascade through 16 references). The REQ sequence is therefore intentionally `001, 002, 003, [004 withdrawn], 005, 006, 007, 008, 009` — the gap is documented, not an omission.

### B.1 §8 Session Handoff pointer-ization (M-POINTER)

- **REQ-OSS-001 (Ubiquitous)**: The `moai.md` §8 Session Handoff block (L648-731) SHALL be condensed to a compact render skeleton + a one-line pointer to its SSOT (`.claude/rules/moai/workflow/session-handoff.md`). The block ALREADY self-declares render-only (L652); the duplicated field-by-field spec / auto-memory persistence procedure / anti-pattern catalogue prose that is fully owned by `session-handoff.md` SHALL be replaced with the pointer. The cut-line marker spec, the Cut-line Marker translation table (L679), the Header translation table (L686), the drift-mitigation parity sentinel (L653), and the render skeleton the orchestrator needs at output time SHALL be preserved (render-SSOT + parity-contract content, §A.3).

- **REQ-OSS-006 (State-driven)**: **While** the §8 Session Handoff block is being pointer-ized, the drift-mitigation parity sentinel's 4-locale column count (en / ko / ja / zh) SHALL be preserved at the render surface. The sentinel (L653) requires the render surface to mirror the SSOT's locale column count for drift-visibility; pointer-ization collapses the duplicated explanatory prose, NEVER the parity-bearing tables or the sentinel itself.

### B.2 §8 Localization Contract + per-banner duplicate-prose deletion (M-DELETE / M-POINTER)

- **REQ-OSS-002 (Ubiquitous)**: For each §8 explanatory passage whose canonical SSOT already lives in a verified `.claude/rules/...` file, the change SHALL replace the duplicated explanation with a one-line cross-reference pointer, PRESERVING all render-SSOT content (the 14 banner skeletons, the 8 per-banner translation tables, the ko-canonical mapping tables, the verbatim-preserve symbol lists). External SSOTs to cross-check: `.claude/rules/moai/core/askuser-protocol.md` (AskUserQuestion mechanics, preview field, free-form prohibition), `.claude/rules/moai/workflow/session-handoff.md` (handoff format), `.claude/rules/moai/development/sprint-round-naming.md` (Epic/Milestone taxonomy). Only prose that is genuinely duplicated AND owned elsewhere is deletable.

- **REQ-OSS-009 (Event-driven)**: **When** the run-phase is about to apply a POINTER edit to a `moai.md` passage, the run-phase SHALL FIRST re-grep the candidate SSOT rule file for the passage's DISTINCTIVE content; **When** the grep returns 0 hits of the distinctive content, the run-phase SHALL block the POINTER edit and reclassify the passage as KEEP. This is the P2 D1 over-cut defense (AC-CMD-009) applied verbatim to the output-style file — a "this is duplicated elsewhere" claim is a defect claim that MUST be tool-verified before deletion (verification-claim-integrity.md §1.1).

### B.3 Render-SSOT preservation (the load-bearing invariant)

- **REQ-OSS-003 (Unwanted behavior)**: The diet SHALL NOT remove any render-SSOT content (§A.3): the 14 banner skeletons SHALL all remain present and renderable; the 8 per-banner en/ko/ja/zh header-translation tables SHALL be preserved verbatim; the §8 Localization Contract ko-canonical mapping tables (label→ko-canonical + banner-body-prose→ko-natural) SHALL be preserved; the §9 Language Rules + §10 Output Rules SHALL be preserved; the verbatim-preserve emoji/box-drawing/symbol lists and the `ultrathink.` keyword token SHALL be preserved. Post-diet, `moai.md` SHALL remain a complete, self-sufficient render spec for everything that does NOT have an external SSOT.

### B.4 Template-mirror parity + neutrality preservation

- **REQ-OSS-005 (Ubiquitous)**: The diet SHALL be applied to BOTH trees identically (LIVE `.claude/output-styles/moai/moai.md` AND `internal/template/templates/.claude/output-styles/moai/moai.md`), keeping them byte-identical. Per the CLAUDE.local.md §2 Template-First rule, the template SSOT tree is edited FIRST, then re-embedded via `make build`, then live-tree parity is verified (`diff` → exit 0). Both trees are confirmed byte-identical at baseline (§F.1). The output-styles count CI guards (`TestOutputStylesExactlyTwo`, `TestOutputStylesTemplateLiveParity`) SHALL still pass post-diet.

- **REQ-OSS-007 (Unwanted behavior)**: The diet SHALL NOT inject any moai-adk internal-development artifact into the `moai.md` body — no internal SPEC IDs, no internal work dates, no commit SHAs, no audit citations ("Audit N Finding AX"), no `feedback_*` / memory paths, no `/Users/` OS-bias paths, no `CLAUDE.local.md` references (per CLAUDE.local.md §15 language-neutrality + §25 internal-content isolation; CI guard `TestTemplateNeutralityAudit`). `moai.md` is a 16-language-neutral distributed asset. (Any pre-existing SPEC-ID *examples* inside §8 worked-examples MAY be neutralized as a side benefit, but neutralization is NOT the SPEC's goal — scope discipline: this SPEC is a diet, not a neutrality sweep.)

### B.5 Derived target (range, not a hard number)

- **REQ-OSS-008 (Ubiquitous)**: The final `moai.md` line count SHALL be a RANGE derived from the KEEP/CUT/POINTER classification (plan.md §C), NOT an arbitrary fixed number. The "~150-250 lines reduction (782 → ~530-630L)" figure is GUIDANCE only. Render-SSOT KEEP content SHALL NOT be cut to hit a number; an honest diet that lands above ~630L is acceptable if the preservation invariants force fewer cuts. The derived target + its justification SHALL be stated in plan.md §C.

---

## C. Constraints

- **C-1** [HARD] Template-First (CLAUDE.local.md §2): edit `internal/template/templates/.claude/output-styles/moai/moai.md` FIRST, then `make build` to re-embed, then verify live-tree byte-parity (`diff` exit 0). BOTH trees diet IDENTICALLY (REQ-OSS-005).
- **C-2** [HARD] BODY editing (NOT frontmatter-only) → higher risk. The plan.md KEEP/CUT/POINTER map MUST be precise enough that run-phase is purely mechanical — no run-phase judgment call about whether a passage is render-SSOT or duplicated-prose.
- **C-3** [HARD] Render-SSOT content carries behavioral value (per-line test = "YES, removing causes the orchestrator to render wrong") → KEEP. Pointer-ization may collapse the surrounding EXPLANATION but MUST preserve the render skeleton + the translation tables + the verbatim-preserve symbol lists (REQ-OSS-003). plan.md §D enumerates the render-SSOT content that MUST survive.
- **C-4** [HARD] Neutrality / isolation (CLAUDE.local.md §15 + §25): no internal SPEC IDs / dates / SHAs / audit citations / memory paths / CLAUDE.local refs injected (REQ-OSS-007).
- **C-5** [HARD] MODERATE bound (user-confirmed): two techniques only (§8 Session Handoff pointer-ize + duplicate-prose deletion). NO banner-template restructure (the rejected "적극" option). Preservation invariants WIN over the ~150-250L target (behavioral-PASS over numeric-proxy, P2 lesson).
- **C-6** No Go code change, no new lint rule. This SPEC is markdown body editing + `make build` re-embed only.
- **C-7** A candidate POINTER passage is proposed ONLY after its SSOT rule file is tool-verified to EXIST and CARRY the content (grep, per verification-claim-integrity.md §1.1 — a "this is duplicated" claim is a defect claim that MUST be tool-verified, never assumed). plan.md §C records the verification command + observed output per candidate.
- **C-8** [HARD] Parity-sentinel honesty (REQ-OSS-006): the §8 Session Handoff drift-mitigation parity sentinel (L653) requires the render surface to keep the 4-locale column count. The diet MUST NOT collapse the parity-bearing tables; pointer-ization shrinks only the duplicated explanatory prose around them.

---

## D. Out of Scope / Exclusions

The SPEC scope is exactly: §8 Session Handoff pointer-ization (M-POINTER) + §8 Localization-Contract / per-banner duplicate-prose deletion (M-DELETE / M-POINTER), applied to BOTH `moai.md` trees in byte-parity, bounded MODERATE and gated by the render-SSOT preservation invariant. The following are explicitly excluded.

### Out of Scope — banner-template restructure (the rejected "적극" option)

- Re-shaping, merging, or re-designing the 14 banner skeletons (Task Start, Delegation Dispatch, Checkpoint Gate, Insight, Verification Matrix, Plan Audit, Discovery Report, Race Absorbed, Epic Stats, Epic Status, Completion Report, Error Recovery, Progress Board, Session Handoff). The MODERATE bound (user-confirmed) is a duplicate-prose diet, NOT a banner restructure. The aggressive restructure option was explicitly REJECTED this session. The banner skeletons + their translation tables are render-SSOT (§A.3) and survive verbatim.

### Out of Scope — render-SSOT content removal to hit a line-number target

- Cutting any render-SSOT content (the 14 banner skeletons, the 8 per-banner translation tables, the ko-canonical mapping tables, the §9/§10 directives, the verbatim-preserve symbol lists, the `ultrathink.` token) purely to reach "~530-630 lines". The "~150-250 lines reduction" figure is guidance; the real target is a range derived from the classification (REQ-OSS-008). Over-cutting render-SSOT content to hit a number is the P2 D1 over-cut failure mode and is forbidden (REQ-OSS-009 gate).

### Out of Scope — moai-adk internal-content injection / neutrality sweep

- Injecting internal SPEC IDs, internal work dates, commit SHAs, audit citations, or `feedback_*`/memory paths into the `moai.md` body (forbidden by REQ-OSS-007 / CLAUDE.local.md §25). The diet REMOVES bloat; it MUST NOT ADD internal-development noise. Neutralizing pre-existing SPEC-ID *examples* in §8 worked-examples is a permitted side benefit but is NOT the SPEC's goal — this SPEC is a diet, not a neutrality sweep (scope discipline).

### Out of Scope — the other Epic Steering-Align SPECs and parallel-session SPECs

- RULE-SCOPING (COMPLETED entry SPEC), CLAUDEMD-DIET (COMPLETED P2), GUARDRAIL-HOOK (COMPLETED P3), and LOCAL-DIET (FUTURE P6) — named in §A.6 for Epic context only. This SPEC authors artifacts ONLY for the output-style diet. In particular, `CLAUDE.local.md` (the maintainer-local always-loaded body) is OUT OF SCOPE here — it is the FUTURE LOCAL-DIET SPEC. **SPEC-DIVECC-INVENTORY-VIEW-001 (a parallel session's SPEC) MUST NOT be touched** (scope discipline, B10 untouched-paths PRESERVE).

### Out of Scope — Go code / lint / SSOT-rule edits

- Go code changes, new lint rules, and edits to the `.claude/rules/...` SSOT files themselves. Pointer-ization (M-POINTER) only edits `moai.md` to point AT an existing SSOT; it does NOT modify the SSOT. (If a pointer-ization reveals a genuine SSOT gap, run-phase returns a blocker — it does not silently edit the SSOT under this SPEC.)

---

## E. Acceptance Criteria Reference

Concrete GEARS-format acceptance criteria with re-runnable verification commands live in `acceptance.md`. Summary of AC themes:

- **AC-OSS-001** — per-tree line-count drop measurable by `wc -l`, LIVE and TEMPLATE each dropping into the derived range (plan.md §C); both trees drop by the SAME amount. SOFT band (~530-630L) — behavioral-PASS allowed if preservation forces fewer cuts.
- **AC-OSS-002** — template ↔ live byte-parity preserved (`diff` → exit 0).
- **AC-OSS-003** — all 14 banner skeletons present post-diet (grep-verifiable per banner heading).
- **AC-OSS-004** — all 8 per-banner en/ko/ja/zh header-translation tables intact (grep-verifiable); the ko-canonical mapping tables intact.
- **AC-OSS-005** — always-on byte-sum reduced per tree (`wc -c`), the reduction attributable solely to M-DELETE + M-POINTER (NOT a banner restructure).
- **AC-OSS-006** — POINTER edits gated on prose-duplication re-verification: each surviving POINTER passage must show ≥1 distinctive-content SSOT hit before its edit; a 0-hit passage is reclassified KEEP (the core over-cut defense, P2 D1 pattern).
- **AC-OSS-007** — neutrality preserved (`TestTemplateNeutralityAudit` STILL passes) + output-styles count guards (`TestOutputStylesExactlyTwo`, `TestOutputStylesTemplateLiveParity`) STILL pass post-diet.
- **AC-OSS-008** — §9 Language Rules + §10 Output Rules + verbatim-preserve symbol lists + `ultrathink.` token survive (grep-verifiable).
- **AC-OSS-009** — every deleted-prose line has a verified external-SSOT home (run-phase enumerates the deleted passage → its SSOT owner + grep evidence).

Each AC is mechanically verifiable. Severity / blocking classification and REQ→AC traceability are in acceptance.md §D.1 / §D.2.

---

## F. Evidence (verification-claim-integrity)

### F.1 Re-verified ground-truth (command → observed output, this tree, 2026-06-23)

| Claim | Command | Observed |
|-------|---------|----------|
| LIVE 782 lines | `wc -l < .claude/output-styles/moai/moai.md` | `782` |
| LIVE 55306 bytes | `wc -c < .claude/output-styles/moai/moai.md` | `55306` |
| Byte-identical trees | `diff .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md; echo $?` | `0` (no diff) → IDENTICAL |
| §8 spans L211-731 | `grep -nE '^## ' .claude/output-styles/moai/moai.md` | §8 at L211, §9 at L732 → §8 = L211-731 (~520 lines = 66% of 782) |
| 14 banner sub-sections | `grep -nE '^### ' moai.md` (within §8) | Task Start L332, Delegation L341, Gate L351, Insight L360, Verification Matrix L376, Plan Audit L409, Discovery L441, Race Absorbed L476, Epic Stats L512, Epic Status L544, Completion L582, Error Recovery L593, Progress Board L606, Session Handoff L648 = 14 banners |
| 8 per-banner header tables (render-SSOT) | `grep -n 'Header translation table\|Cut-line Marker translation' moai.md` | L370, L396, L427, L460, L495, L529, L563, L679 (cut-line), L686 (session-handoff header) — render-SSOT, NO external owner → PRESERVE |
| §8 Session Handoff render-only marker | `grep -n 'render-only, not canonical\|canonical lives in' moai.md` | `652:` `<!-- render-only, not canonical — canonical lives in .claude/rules/moai/workflow/session-handoff.md (SSOT) ... -->` → **POINTER candidate confirmed** |
| §8 Session Handoff parity sentinel | `grep -n 'Drift-mitigation self-check sentinel' moai.md` | `653:` parity sentinel requires "same locale column count (en / ko / ja / zh — 4 columns)" at render surface → cut-line + header tables MUST stay (REQ-OSS-006) |
| session-handoff.md owns 6-block + cut-line + field-by-field + pre-emit | `grep -c '6-block\|Block 1\|Block 6\|Field-by-Field\|Cut-line Marker Specification\|Pre-emit self-check' session-handoff.md` | 23 (6-block) / 12 (cut-line) / 10 (pre-emit) → **§8 Session Handoff duplicated PROSE is owned by session-handoff.md** |
| session-handoff.md owns source_session_id + env fallback | `grep -c 'source_session_id\|environment-fallback\|moai session current' session-handoff.md` | 6 → the §8 source_session_id field detail is owned externally → POINTER |
| session-handoff.md owns effort-ultracode re-set | `grep -c 'effort ultracode\|workflow fan-out' session-handoff.md` | 3 → the §8 ultracode re-set detail is owned externally → POINTER |
| askuser-protocol.md owns AskUserQuestion mechanics + preview + free-form | `grep -c 'select:AskUserQuestion\|max 4 questions\|Channel Monopoly\|Free-form\|Preview Field' askuser-protocol.md` | 18 → §10 Output Rules AskUserQuestion + preview-field prose is owned externally → POINTER candidate |
| sprint-round-naming.md owns Epic/Milestone taxonomy | `grep -c 'Epic\|Milestone\|multi-SPEC\|within-SPEC' sprint-round-naming.md` | 50 → Epic Stats/Epic Status banner taxonomy reference is owned externally (the banner skeleton stays render-SSOT; only the taxonomy EXPLANATION is pointer-able) |
| output-styles count CI guard | `grep -n 'func TestOutputStylesExactlyTwo\|func TestOutputStylesTemplateLiveParity' internal/template/output_styles_audit_test.go` | `295:`, `370:` → both guards exist; must pass post-diet (REQ-OSS-005) |
| neutrality CI guard | `grep -n 'func TestTemplateNeutralityAudit' internal/template/template_neutrality_audit_test.go` | `321:` → authoritative neutrality guard; must pass post-diet (REQ-OSS-007) |

### F.2 Gaps (explicitly NOT observed at plan-phase)

- The exact final line-count is NOT fixed at plan-phase — it is a RANGE derived from the KEEP/CUT/POINTER classification (REQ-OSS-008); the precise number is a run-phase outcome.
- The token reduction is APPROXIMATED via line/byte-sum, not a real tokenizer count (acceptable per the per-line-test framing; exact token count deferred to run-phase if needed).
- The PRECISE condensed wording of the §8 Session Handoff render skeleton (REQ-OSS-001) is NOT authored at plan-phase — plan.md §C identifies WHICH prose is duplicated-and-owned-elsewhere vs render-SSOT; the run-phase authors the condensed render skeleton while preserving the parity-bearing tables.
- Whether any specific §8 passage is render-SSOT vs duplicated-prose is assessed per-passage in plan.md §C; the plan-phase default is to KEEP (render-SSOT) unless the passage is explicitly proven duplicated-and-owned-elsewhere via grep (C-7).
- **(D1-equivalent)** The prose-duplication re-audit (plan.md §C) used `grep` of the DISTINCTIVE content per passage. This is a plan-phase grep snapshot — the run-phase MUST re-run the prose-duplication grep (AC-OSS-006) before each POINTER edit, since the SSOT files could change between plan and run.

### F.3 Residual risk

- BODY editing of a 66%-banner-heavy file is higher-risk than the frontmatter-only RULE-SCOPING-001. Mitigated by the precise plan.md §C KEEP/CUT/POINTER map (run-phase is mechanical) + the AC-OSS-003/004/008 render-SSOT-survival guards (banner skeletons, translation tables, directives cannot silently drop).
- Over-aggressive pointer-ization could remove a render passage that LOOKS duplicated but IS render-SSOT (the orchestrator reads it at output time). **This is the P2 D1 over-cut risk applied to the output-style file.** Mitigated by the render-SSOT preservation invariant (REQ-OSS-003, the §A.3 table) + the run-phase precondition AC-OSS-006 (re-grep distinctive content; 0 hits → block POINTER, force KEEP) + the parity-sentinel honesty constraint (C-8 / REQ-OSS-006).
- The §8 Session Handoff parity sentinel (L653) is itself a drift-mitigation contract — pointer-izing the block while preserving the sentinel + the 4-locale tables is a precise edit; a careless collapse would BREAK the parity contract the sentinel enforces. Mitigated by C-8 + REQ-OSS-006 (the sentinel + tables are explicitly carved out of the pointer-ization).
- `make build` re-embed could surface unrelated template drift. Mitigated by scoping the run-phase diff to `moai.md` only and asserting byte-parity (AC-OSS-002) + the output-styles parity guard (AC-OSS-007).
- A pointer-ization that points at an SSOT which is itself path-scoped (conditionally loaded) means the duplicated explanation is no longer always-available — but that is the INTENT (session-handoff.md is always-loaded itself per its no-`paths:` declaration; askuser-protocol.md / sprint-round-naming.md load on their own budgets). Risk only if a render-SSOT always-needed passage is removed; guarded by REQ-OSS-003.

---

## G. SPEC ID Pre-Write Self-Check (recorded per protocol)

decomposition: SPEC ✓ | STEERING ✓ | ALIGN ✓ | OUTPUT ✓ | STYLE ✓ | SLIM ✓ | 001 ✓ → PASS

Canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`: first segment literal `SPEC`; middle segments STEERING / ALIGN / OUTPUT / STYLE / SLIM each match `[A-Z][A-Z0-9]*` (first char uppercase letter, rest uppercase alphanumerics, length ≥ 1, valid); last segment `001` matches `\d{3}` (digit-only, no alpha suffix). All segments PASS.

---

## H. Cross-References

- Anthropic best-practices "Write an effective always-loaded instruction surface" (per-line test, bloat warning) — the doctrine applied.
- Anthropic blog "Steering Claude Code: skills, hooks, rules, subagents and more" (always-loaded vs conditional steering surfaces; output-style is always-loaded).
- `.claude/output-styles/moai/moai.md` — the diet target (§8 Response Templates L211-731; §8 Localization Contract L213-331; §8 Session Handoff L648-731).
- `.claude/rules/moai/workflow/session-handoff.md` — the SSOT for the §8 Session Handoff render-only block (the M-POINTER target's canonical owner; the §8 block's L652 marker already names it).
- `.claude/rules/moai/core/askuser-protocol.md` — SSOT for AskUserQuestion mechanics / preview field / free-form prohibition (§10 Output Rules pointer candidate).
- `.claude/rules/moai/development/sprint-round-naming.md` — SSOT for Epic/Milestone taxonomy (Epic Stats / Epic Status banner taxonomy reference; banner skeleton stays render-SSOT).
- CLAUDE.local.md §2 (Template-First), §15 (language-neutrality), §25 (internal-content isolation).
- CI guards: `internal/template/output_styles_audit_test.go` (`TestOutputStylesExactlyTwo` L295, `TestOutputStylesTemplateLiveParity` L370); `internal/template/template_neutrality_audit_test.go` (`TestTemplateNeutralityAudit` L321).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — `(none) → draft` owned by manager-spec.
- SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001 (sibling P2, COMPLETED) — the M-DELETE + M-POINTER diet technique + the AC-009 over-cut-gate pattern this SPEC mirrors (the render-SSOT preservation invariant generalizes P2's D1 demotion defense to the output-style file).
- SPEC-STEERING-ALIGN-GUARDRAIL-HOOK-001 (sibling P3, COMPLETED, origin c463257b8) — Epic Steering-Align lineage.
