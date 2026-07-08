---
id: SPEC-HANDOFF-GOALFIX-001
title: "Retire the inert in-block '# /goal' resume line — two-step post-paste /goal handoff"
version: "0.1.1"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "handoff, goal-directive, defect-fix"
tier: M
---

# SPEC-HANDOFF-GOALFIX-001 — Retire the inert in-block `# /goal` resume line; two-step post-paste `/goal` handoff

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-08 | GOOS행님 (via manager-spec) | Initial plan-phase draft — defect fix: the Block 1 `# /goal` line can never fire; replace with a two-step post-paste follow-up block + orchestrator reminder + paste-time activation matrix. |
| 0.1.1 | 2026-07-08 | GOOS행님 (via manager-spec) | Plan-audit iter1 (PASS-WITH-DEBT 0.87) fixes: D2 REQ-GF-003 trigger re-grounded in auto-memory detection; D4 `related_specs` frontmatter field dropped (lineage remains in §A.4/§E body prose). User-approved scope addition: REQ-GF-009 goal-first bootstrap variant (documented alternative, NOT the default). |

## §A. Context and Motivation (WHY)

### A.1 The defect (verified ground truth)

SPEC-V3R6-HANDOFF-GOAL-BINDING-001 bound a purpose-conditional `# /goal <completion-condition>` line into **Block 1 of the paste-ready resume message body** — inside the main cut-line-bounded fenced block that the user pastes verbatim. This placement is structurally inert; the line **can never fire**:

1. **Official — slash command parsing**: `https://code.claude.com/docs/en/interactive-mode` § Quick commands — slash commands are recognized only as "`/` at start" of the input. A `/goal` line mid-paste is plain text. The `#` prefix makes it doubly inert.
2. **Official — `/goal` semantics**: `https://code.claude.com/docs/en/goal` — "Run `/goal` followed by the condition" is a user-typed TUI command; "Setting a goal starts a turn immediately, with the condition itself as the directive"; `/clear` removes an active goal; restoration happens ONLY via `--resume`/`--continue`. The MODEL cannot invoke `/goal` (it is not a model-invocable tool/skill, unlike `/moai` which the orchestrator routes via the Skill tool).
3. **Runtime evidence**: a real project session (6.3 MB debug log) where the user pasted a resume containing `# /goal gradle jvmTest ...` shows ZERO goal-evaluator activity and zero `/goal` traces — the autonomous continuation loop never armed while the orchestrator merely mimicked the intent.
4. **Doctrine self-contradiction**: `session-handoff.md`'s own ultracode bullet already states "a `#`-commented slash line cannot execute at paste time" (for `/effort ultracode`) — the identical logic was never applied to the `# /goal` line the same doctrine prescribes.

Consequence: **every** handoff whose next SPEC is run-phase with a machine-verifiable end-state silently loses the autonomous-continuation loop, while the doctrine claims to preserve it. This is a P1 doctrine defect, not an enhancement.

### A.2 What DOES fire at paste time (verified)

- `ultrathink` — runtime keyword, position-independent in message text.
- bare `ultracode` — official: "include the keyword `ultracode` in your prompt".
- `fan out subagents ...` — official: "Asking in your own words also works — same opt-in".
- `mode:` — orchestrator-interpreted seed (by design not a runtime keyword).

`/goal`, `/effort`, `/clear` belong to a fourth class: **user-only TUI commands** that parse only at input start of a standalone user message and cannot be set by pasted body text nor by the model.

### A.3 Fix shape (WHAT)

Replace the single-message in-block line with a **two-step handoff**: the main resume block stays `/goal`-free; when the emission condition holds (unchanged from v1), a **second, separate cut-line-bounded copy block** is emitted OUTSIDE and AFTER the main block, containing exactly one line `/goal <completion-condition>` (no `#` prefix), plus a localized instruction that it MUST be sent as its own standalone message — RECOMMENDED after the resumed session's Implementation Kickoff Approval, since setting a goal starts a turn immediately. The resumed session's orchestrator gains a reminder obligation, because the model cannot set the goal on the user's behalf.

### A.4 Lineage

- **Amends**: SPEC-V3R6-HANDOFF-GOAL-BINDING-001 (introduced the Block 1 `/goal` re-set line; its emission CONDITION — run-phase next SPEC AND machine-verifiable end-state — is preserved verbatim; only the PLACEMENT and delivery mechanism change).
- **Absorbs debt from**: SPEC-HANDOFF-FANOUT-001 (recorded Low debt: the Block 1 line-order invariant opener parenthetical mentions only the bare `ultracode` keyword, not the fan-out steering phrase — the sentence is rewritten by REQ-GF-001 anyway).

## §B. Requirements (GEARS)

Notation: GEARS. `<subject>` generalized. `shall`/`shall not` normative. Baselines measured at HEAD `3d35cc18d` (2026-07-08); run-phase MUST re-measure before editing (parallel-session commits are frequent on these surfaces).

### REQ-GF-001 — Retire the in-block `# /goal` line (Ubiquitous + Unwanted)

The paste-ready resume-message doctrine shall define the main cut-line-bounded resume block WITHOUT any `# /goal` line. The `session-handoff.md` canonical skeleton, the Block 1 line-order invariant, the directive-binding prose, the `mode:` bullet cross-reference, and the worked Example shall not contain the `# /goal` line or references to a "Block 1 `/goal` re-set line". The rewritten Block 1 line order is: `ultrathink.` opener (with an optional appended bare `ultracode` keyword or fan-out steering phrase) → `mode:` line (when present) → `applied lessons:` → `source_session_id:`. The doctrine shall record the retirement rationale: a mid-paste `#`-prefixed slash line is inert per official slash-command parsing rules (input-start-only recognition).

- Baseline: `# /goal` literal — 6 occurrences in `session-handoff.md` (live and template each), 1 in `moai.md` (live and template each). Target: 0 on all surfaces.
- Baseline: `re-set` literal — 4 occurrences in `session-handoff.md`, 1 in `moai.md`, 3 in `goal-directive.md` (live and template each; all in `/goal` context). Target: 0 (the "re-set line" vocabulary is retired with the mechanism; rewritten prose uses "post-paste `/goal` follow-up block").

### REQ-GF-002 — Two-step post-paste `/goal` follow-up block (Capability gate + Event-driven)

**Where** the `/goal` emission condition holds (next SPEC is run-phase AND declares a machine-verifiable end-state — condition UNCHANGED from v1), **When** the orchestrator emits a paste-ready resume message, the orchestrator shall emit a second, separate copy block OUTSIDE and AFTER the main cut-line block, consisting of:

1. A localized **instruction line** (prose, outside the cut-line markers) stating the block MUST be sent as its own standalone message (slash commands parse only at input start), RECOMMENDED after the resumed session's Implementation Kickoff Approval (setting a goal starts a turn immediately).
2. A cut-line-bounded fenced block (reusing the existing Cut-line top/bottom marker text rows) containing exactly one line: `/goal <completion-condition>` — no `#` prefix, no additional lines.

**Where** the emission condition does NOT hold, the orchestrator shall emit no follow-up block (output byte-identical to the pre-existing no-`/goal` form). The `session-handoff.md` § Output Surface ordering shall be updated to: (1) main fenced block → (2) conditional instruction line + `/goal` follow-up block → (3) memory file path → (4) one-sentence next-session summary. The § Auto-Memory Integration obligation shall include the follow-up block verbatim in the persisted memory entry when emitted. The doctrine section carrying this mechanism shall use the H2 heading `## Post-Paste /goal Follow-up Block` (grep-stable token).

### REQ-GF-003 — Resumed-session orchestrator `/goal` reminder obligation (Event-driven)

**When** the resumed session's orchestrator detects a pending `/goal` condition for the resumed SPEC — detection mechanism: the **handoff memory entry** (the resume message AND its post-paste follow-up block are persisted verbatim to auto-memory per `session-handoff.md` § Auto-Memory Integration, so the resumed session reads the pending condition from memory, NOT from the pasted body), or, failing a memory hit, by **re-deriving the emission condition** (resumed SPEC is run-phase AND declares a machine-verifiable end-state) — the orchestrator shall remind the user — via natural-language status guidance, NOT `AskUserQuestion` (announcement, not a decision) — to send the `/goal` line as a standalone message at the recommended moment (after Implementation Kickoff Approval), because the model cannot set a goal on the user's behalf. Note: post-REQ-GF-001 the pasted main block itself carries NO `/goal` reference; the memory entry / condition re-derivation is the only detection path. This **reminder obligation** shall be recorded — carrying the literal grep-stable token `reminder obligation` — in both `session-handoff.md` (within the Post-Paste /goal Follow-up Block section) and `goal-directive.md` § MoAI Integration Notes.

### REQ-GF-004 — Paste-time activation matrix (Ubiquitous)

The `session-handoff.md` doctrine shall contain a compact normative table under the H2 heading `## Paste-Time Activation Matrix` classifying every handoff directive by activation mechanism:

| Class | Directives | Mechanism | Fires from pasted body? |
|-------|-----------|-----------|------------------------|
| (a) Paste-time keyword | `ultrathink`, bare `ultracode` | Runtime keyword, position-independent in message text | YES |
| (b) Paste-time natural-language phrase | `fan out subagents (<scope>)` | Explicit multi-agent opt-in phrase — same opt-in class as (a) | YES |
| (c) Orchestrator-interpreted text | `mode:` seed, Block 5 `실행: /moai <subcommand>` | The orchestrator reads the text and routes (`/moai` via the Skill tool); NOT auto-executed as a slash command | YES (via orchestrator interpretation) |
| (d) User-only TUI command | `/goal`, `/effort`, `/clear` | Slash command parsed ONLY at input start; not model-invocable; cannot be set by pasted body text NOR by the model | NO — requires a standalone user message |

The matrix shall cite the official doc URLs: `https://code.claude.com/docs/en/interactive-mode` and `https://code.claude.com/docs/en/goal`. (Exact row wording is run-phase latitude; class structure, the four directive groupings, and both citations are normative.)

### REQ-GF-005 — `goal-directive.md` § MoAI Integration Notes correction (Ubiquitous)

The `goal-directive.md` § MoAI Integration Notes shall not describe a "Block 1 `/goal` re-set line" (baseline: 2 occurrences of the phrase `Block 1 ` + backtick-`/goal`-backtick + ` line` on one bullet). The "resume pairing" bullet shall be rewritten to describe the two-step mechanism: the paste-ready resume body carries NO `/goal` line (mid-paste slash lines are inert); the orchestrator emits a separate post-paste `/goal` follow-up block; the resumed session's orchestrator reminds the user to send it post-Kickoff-Approval. The Implementation Kickoff Approval invariant (a `/goal` never authorizes autonomous run-phase entry) is preserved verbatim in meaning.

### REQ-GF-006 — FANOUT-001 residual debt absorption (Ubiquitous)

The rewritten Block 1 line-order invariant opener parenthetical (REQ-GF-001) shall mention both optional opener attachments using the literal phrase "bare `ultracode` keyword or fan-out steering phrase" — clearing the recorded Low debt from SPEC-HANDOFF-FANOUT-001. Baseline: 0 occurrences of that phrase on any surface.

### REQ-GF-007 — Surface parity (Ubiquitous)

The changes of REQ-GF-001..006 shall be applied consistently to all six surfaces plus the embed rebuild:

1. `.claude/rules/moai/workflow/session-handoff.md` (SSOT) — skeleton, Field-by-Field Block 1, line-order invariant, directive binding, Example, § Output Surface, § Auto-Memory Integration, Anti-Patterns (the existing "Omitting the `/goal` re-set line..." anti-pattern REWRITTEN to the two-step form + a NEW anti-pattern for embedding any slash command inside the main resume body), Pre-emit self-check (item reworded in place; item count unchanged at 10).
2. `.claude/output-styles/moai/moai.md` §8 (render surface) — skeleton `# /goal` annotation removed; compact post-paste follow-up block render note added; the pre-emit 12-item list's `/goal` item reworded to the two-step form; concern-name qualifier parity with the SSOT sentinel preserved.
3. `.claude/rules/moai/workflow/goal-directive.md` — per REQ-GF-005 + REQ-GF-003.
4. `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` — byte-identical mirror of (1).
5. `internal/template/templates/.claude/output-styles/moai/moai.md` — byte-identical mirror of (2).
6. `internal/template/templates/.claude/rules/moai/workflow/goal-directive.md` — byte-identical mirror of (3).
7. `make build` shall exit 0 after the template edits (re-embeds templates into the binary).

The template mirrors shall not contain any internal SPEC ID (template neutrality — the rewritten doctrine prose cites official URLs only, never `SPEC-HANDOFF-GOALFIX-001` or any other internal SPEC ID; since live and template are byte-identical, the live prose is equally SPEC-ID-free in the edited regions).

### REQ-GF-008 — Localization of the follow-up block (Ubiquitous)

The follow-up block's cut-line marker text shall reuse the existing Cut-line top/bottom rows of the Localization Table (no new marker rows; `✂`/`─` symbols verbatim). The **instruction line** shall translate per `conversation_language`: the Localization Table gains exactly one new row with Element label `Post-paste /goal instruction line`, carrying en/ko/ja/zh renderings (en and ko canonical below; ja/zh per the naturalization principle):

- en: instructs sending the `/goal` line below as its own standalone message AFTER Implementation Kickoff Approval, noting slash commands parse only at input start and setting a goal starts a turn immediately.
- ko (canonical): 아래 `/goal` 라인을 구현 착수 승인 후 **별도 메시지로 단독 전송** — 슬래시 커맨드는 입력 시작에서만 인식되며, goal 설정 즉시 턴이 시작됨.

The `/goal` command token itself and the `<completion-condition>` content are **locale-verbatim** (never translated), consistent with the existing `mode:` / `fan out subagents` locale-verbatim protocol-token policy.

### REQ-GF-009 — Goal-first bootstrap variant (documented alternative, NOT the default) (Ubiquitous)

The `## Post-Paste /goal Follow-up Block` doctrine section (via a sibling subsection within it) shall document an explicit **goal-first bootstrap** alternative single-paste form: a standalone one-line `/goal <resume pointer + compact condition>` message (illustrative: `/goal SPEC-X run 재개: memory의 <handoff-file>와 progress.md를 읽고 이어서 진행. 완료 조건: <machine-verifiable end-state>, 또는 N턴 후 중단.`). Normative content:

- **(a) Selection criterion**: the user wants one-paste + autonomous continuation; the two-step handoff (REQ-GF-002) remains the DEFAULT.
- **(b) Caveats (stated verbatim-class)**: effort keywords (`ultrathink` / `ultracode`) inside a slash-command argument are NOT documented to fire — the session may run at default effort; precondition verification shifts from paste-time structure (Block 4 verifiable commands) to **model discretion** via the directive text.
- **(c) Invariants preserved**: the condition must stay compact (official guidance: one measurable end state); Implementation Kickoff Approval unaffected; the `/goal` token locale-verbatim.

Grounding: official goal doc — "Setting a goal starts a turn immediately, with the condition itself as the directive" (`https://code.claude.com/docs/en/goal`). Surfaces unchanged (same 6 + `make build`) — this is additional prose within the same new section, mirrored byte-identically per REQ-GF-007. The subsection shall carry the literal grep-stable tokens `goal-first bootstrap` and `model discretion` (both baseline 0, verified 2026-07-08 including case-insensitive `goal-first` / `bootstrap` sweeps).

## §C. Constraints

- **[HARD] Re-measure before edit**: all line numbers in this SPEC are content-token anchors measured at HEAD `3d35cc18d`; parallel-session commits land frequently on these surfaces (FANOUT-001 closed today). Run-phase MUST re-grep every anchor before editing and MUST use content-token `Edit` old_strings, never line-number-derived paraphrases.
- **[HARD] SSOT→render direction**: `session-handoff.md` is the SSOT; `moai.md` §8 is the render surface. Edit the SSOT first, then align the render surface per the existing drift-mitigation sentinel (concern-name qualifiers preserved).
- **[HARD] Template-First + neutrality**: edit live and template mirrors to byte-identical content; no internal SPEC IDs, REQ tokens, dates, or commit SHAs in template text (CI guard: `template-neutrality-check.yaml`, `internal_content_leak_test.go`).
- **[HARD] Pathspec-limited commits**: unrelated modified/untracked files exist in the working tree; every commit stages explicit paths only (never `git add -A`).
- **No Go source changes**: `make build` is required only to re-embed templates; zero `internal/**/*.go` edits.
- **Emission condition frozen**: the run-phase + machine-verifiable-end-state condition from GOAL-BINDING-001 is reused verbatim; this SPEC changes placement/delivery only.
- **Pre-emit self-check counts frozen**: SSOT stays 10 items, render stays 12 items (items reworded in place, none added/removed) — avoids cascading count-parity churn.

## §D. Acceptance Criteria

Canonical AC enumeration lives in `acceptance.md` (AC-GF-001..013, baseline-delta style: every negative-space check cites the measured non-zero baseline; every positive-space check has a verified 0 baseline under a distinguishing multi-word token — single-word substring greps are prohibited after the plan-audit D1 `remind`⊂`reminder` false-baseline finding). Summary of gate classes:

- Negative space: `# /goal` → 0 (baseline 6/6/1/1), `re-set` → 0 (baseline 4/4/1/1/3/3), goal-directive "Block 1 `/goal` line" phrase → 0 (baseline 1 line / 2 in-line occurrences).
- Positive space: `## Post-Paste /goal Follow-up Block` = 1, `## Paste-Time Activation Matrix` = 1, both official URLs present, FANOUT debt phrase = 1, Localization row = 1, `reminder obligation` token on 2 surfaces, `goal-first bootstrap` + `model discretion` tokens (REQ-GF-009).
- Structural: 3× live↔template byte-identical diff, `make build` exit 0, template SPEC-ID neutrality grep = 0.

## §E. Cross-References

- `.moai/specs/SPEC-V3R6-HANDOFF-GOAL-BINDING-001/` — amended predecessor (status: completed; its REQ-HGB emission condition survives, its placement is superseded by this SPEC).
- `.moai/specs/SPEC-HANDOFF-FANOUT-001/` — sibling Mode 4 doctrine (closed `cdaea2fbb`); Low debt absorbed by REQ-GF-006.
- `https://code.claude.com/docs/en/interactive-mode` — slash-command input-start parsing (ground truth for the defect).
- `https://code.claude.com/docs/en/goal` — `/goal` semantics, `/clear` removal, `--resume` restoration.
- `.claude/rules/moai/workflow/session-handoff.md` § Cut-line Marker Specification + § Localization Table — reused marker rows (REQ-GF-008).

## §F. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — Runtime/Go enforcement

- No Go code parses, validates, or emits the follow-up block; this is doctrine-only. `make build` re-embeds templates but no `internal/**/*.go` source changes.
- No lint rule addition for detecting in-body slash commands (a future candidate, not this SPEC).

### Out of Scope — `/effort ultracode` mechanics

- The bare-`ultracode` paste-time keyword and the `/effort ultracode` session-persistence variant are retained as-is; they are classified by the activation matrix (REQ-GF-004) but their emission rules do not change.

### Out of Scope — `mode:` seed and fan-out phrase semantics

- Mode-seed enum, thresholds, and fan-out steering phrase semantics are owned by SPEC-HANDOFF-MSGMODE-001 / SPEC-HANDOFF-FANOUT-001; this SPEC touches only the line-order invariant opener parenthetical (REQ-GF-006).

### Out of Scope — Native `--resume`/`--continue` goal restoration

- Claude Code's native goal restoration on `--resume`/`--continue` is documented, not modified; no attempt to make `/clear`-survival automatic.
