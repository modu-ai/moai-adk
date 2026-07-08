---
id: SPEC-HANDOFF-GOALFIX-001
title: "Acceptance criteria — retire inert '# /goal' line, two-step post-paste handoff"
version: "0.1.1"
status: completed
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

# SPEC-HANDOFF-GOALFIX-001 — Acceptance Criteria

All ACs are mechanically verifiable (grep/diff/make), baseline-delta style: negative-space checks cite the measured non-zero baseline (HEAD `3d35cc18d`, 2026-07-08); positive-space checks have a verified 0 baseline. Run-phase MUST re-measure baselines before editing; if a baseline shifted (parallel commit), record the new baseline in progress.md §E.2 and proceed with the delta logic unchanged.

Path shorthands:
- `SH` = `.claude/rules/moai/workflow/session-handoff.md`
- `OM` = `.claude/output-styles/moai/moai.md`
- `GD` = `.claude/rules/moai/workflow/goal-directive.md`
- `T/<x>` = `internal/template/templates/<x>`

## §D. AC Matrix

### AC-GF-001 — `# /goal` literal eliminated (negative space)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-001a | `grep -o '# /goal' SH \| wc -l` | 6 | 0 |
| AC-GF-001b | `grep -o '# /goal' T/SH \| wc -l` | 6 | 0 |
| AC-GF-001c | `grep -o '# /goal' OM \| wc -l` | 1 | 0 |
| AC-GF-001d | `grep -o '# /goal' T/OM \| wc -l` | 1 | 0 |
| AC-GF-001e | `grep -o '# /goal' GD T/GD \| wc -l` | 0 | 0 (guard: stays 0) |

### AC-GF-002 — "re-set" vocabulary retired with the mechanism (negative space)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-002a | `grep -o 're-set' SH \| wc -l` | 4 | 0 |
| AC-GF-002b | `grep -o 're-set' OM \| wc -l` | 1 | 0 |
| AC-GF-002c | `grep -o 're-set' GD \| wc -l` | 3 | 0 |
| AC-GF-002d | same three greps on T/SH, T/OM, T/GD | 4/1/3 | 0/0/0 |

Pre-verified: all 8 baseline occurrences are `/goal`-context (no collateral meaning loss).

### AC-GF-003 — Post-paste follow-up block section exists (positive space, REQ-GF-002)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-003a | `grep -c '^## Post-Paste /goal Follow-up Block' SH` (and T/SH) | 0 | 1 each |
| AC-GF-003b | `grep -c '^/goal <completion-condition>$' SH` (block skeleton, no `#` prefix) | 0 | ≥1 |
| AC-GF-003c | `grep -c 'standalone message' SH` | 0 | ≥2 (instruction spec + anti-pattern/matrix) |
| AC-GF-003d | Structural read: the skeleton's `/goal` line sits between its own `✂────` cut-line markers, OUTSIDE and AFTER the main block; instruction line sits outside the markers | n/a | PASS by inspection, cited in §E.2 evidence |
| AC-GF-003e | Emission condition text preserved: `grep -c 'machine-verifiable end-state' SH` | 3 (re-measured 2026-07-08) | ≥3 (not reduced; now additionally bound to the follow-up block) |

### AC-GF-004 — Resumed-session reminder obligation (positive space, REQ-GF-003)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-004a | `grep -c 'reminder obligation' SH` | 0 (verified 2026-07-08; NOTE — `grep -ci 'remind'` is UNUSABLE here: its SH baseline is 3, since `remind` is a substring of `reminder` in the pre-existing "ceremonial reminder" prose. Plan-audit D1 fix: distinguishing multi-word token, guaranteed non-vacuous by REQ-GF-003's literal-token mandate) | ≥1 |
| AC-GF-004b | `grep -ci 'remind' GD` | 0 (verified genuine — no substring collision on GD) | ≥1 |
| AC-GF-004c | Reminder prose specifies: natural-language status guidance, NOT AskUserQuestion; timing = post-Implementation-Kickoff-Approval; rationale = model cannot invoke `/goal` | n/a | PASS by inspection on both surfaces |

### AC-GF-005 — Paste-time activation matrix (positive space, REQ-GF-004)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-005a | `grep -c '^## Paste-Time Activation Matrix' SH` (and T/SH) | 0 | 1 each |
| AC-GF-005b | `grep -c 'code.claude.com/docs/en/interactive-mode' SH` | 0 | ≥1 |
| AC-GF-005c | `grep -c 'code.claude.com/docs/en/goal' SH` | 0 | ≥1 |
| AC-GF-005d | Matrix carries 4 classes; class (d) groups `/goal`, `/effort`, `/clear` as user-only TUI commands requiring a standalone user message; class (c) covers `mode:` + Block 5 `/moai` (orchestrator-routed via Skill tool, not auto-executed) | n/a | PASS by inspection |

### AC-GF-006 — goal-directive.md correction (REQ-GF-005)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-006a | ``grep -c 'Block 1 `/goal` line' GD`` (plain backticks inside single quotes — GNU grep treats `\`` as a start-of-buffer anchor, so backslash-escaped backticks false-pass on Linux; line-count form suffices since the target is 0) | 1 line (carrying 2 in-line occurrences; darwin re-verified 2026-07-08) | 0 |
| AC-GF-006b | `grep -c 'follow-up block' GD` | 0 | ≥1 |
| AC-GF-006c | Kickoff-Approval invariant retained: `grep -c 'Implementation Kickoff Approval' GD` | ≥1 (pre-existing) | ≥ baseline (not reduced) |

### AC-GF-007 — FANOUT-001 debt cleared (REQ-GF-006)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-007a | ``grep -c 'bare `ultracode` keyword or fan-out steering phrase' SH`` (and T/SH; plain backticks inside single quotes per plan-audit D3) | 0 (darwin re-verified 2026-07-08 with the corrected pattern) | 1 each (in the rewritten line-order invariant) |

### AC-GF-008 — Localization row (REQ-GF-008)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-008a | `grep -c 'Post-paste /goal instruction line' SH` (Localization Table Element column; and T/SH) | 0 | 1 each |
| AC-GF-008b | The row carries 4 locale columns (en/ko/ja/zh); the `/goal` token appears verbatim (untranslated) in all 4 renderings | n/a | PASS by inspection |
| AC-GF-008c | No new cut-line MARKER rows added: cut-line top/bottom rows count unchanged (2 rows) | 2 | 2 |

### AC-GF-009 — Render surface parity (REQ-GF-007 item 2)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-009a | OM §8 skeleton contains no `# /goal` line (covered by AC-GF-001c) and gains a follow-up block render note: `grep -c 'follow-up block' OM` | 0 | ≥1 |
| AC-GF-009b | Pre-emit counts frozen: `grep -c 'Pre-emit self-check (paste-ready budget) — 10 items' SH` = 1 AND `grep -c 'Pre-emit self-check (12 items)' OM` = 1 | 1 / 1 | 1 / 1 |
| AC-GF-009c | Sentinel concern-name qualifiers unchanged: `grep -c 'paste-ready budget' SH OM` each ≥1 | ≥1 | ≥1 (no reduction) |

### AC-GF-010 — Mirror byte-identity + rebuild (REQ-GF-007 items 4-7)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-010a | `diff -q SH T/SH` | identical | identical (exit 0) |
| AC-GF-010b | `diff -q OM T/OM` | identical | identical (exit 0) |
| AC-GF-010c | `diff -q GD T/GD` | identical | identical (exit 0) |
| AC-GF-010d | `make build` | exit 0 | exit 0 |

### AC-GF-011 — Template neutrality (REQ-GF-007 closing clause)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-011a | `grep -rn 'SPEC-HANDOFF-GOALFIX' internal/template/templates/ \| wc -l` | 0 | 0 |
| AC-GF-011b | `grep -rn 'SPEC-V3R6-HANDOFF-GOAL-BINDING' internal/template/templates/ \| wc -l` | 0 | 0 |

### AC-GF-012 — Anti-pattern rewrite (REQ-GF-007 item 1)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-012a | ``grep -c 'Omitting the `/goal` re-set line' SH`` (plain backticks inside single quotes per plan-audit D3) | 1 (darwin re-verified 2026-07-08 with the corrected pattern) | 0 |
| AC-GF-012b | ``grep -c 'Omitting the post-paste `/goal` follow-up block' SH`` (plain backticks inside single quotes per plan-audit D3) | 0 (darwin re-verified 2026-07-08 with the corrected pattern) | 1 |
| AC-GF-012c | New in-body-slash anti-pattern present: `grep -c 'input start' SH` | 0 | ≥1 (in the new anti-pattern and/or matrix) |

### AC-GF-013 — Goal-first bootstrap variant (REQ-GF-009)

| Sub | Command | Baseline | Target |
|-----|---------|----------|--------|
| AC-GF-013a | `grep -c 'goal-first bootstrap' SH` (and T/SH) | 0 (verified 2026-07-08; case-insensitive sweeps `goal-first` = 0 and `bootstrap` = 0 confirm no substring collision) | ≥1 each |
| AC-GF-013b | `grep -c 'model discretion' SH` (and T/SH) | 0 (verified 2026-07-08) | ≥1 each |
| AC-GF-013c | Variant prose states all five invariant/caveat clauses: two-step remains DEFAULT; effort keywords (`ultrathink`/`ultracode`) inside a slash-command argument NOT documented to fire (session may run at default effort); precondition verification shifts to model discretion; condition compact (one measurable end state); Implementation Kickoff Approval unaffected + `/goal` token locale-verbatim | n/a | PASS by inspection |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — run-phase next SPEC with verifiable end-state (happy path)

- **Given** the orchestrator prepares a handoff whose next SPEC is run-phase and declares "test suite passes AND lint clean, or stop after 20 turns",
- **When** it emits the paste-ready resume per the rewritten doctrine,
- **Then** the main cut-line block contains NO `/goal` or `# /goal` line, AND a localized instruction line + a second cut-line-bounded block containing exactly `/goal the SPEC's test suite passes AND lint is clean, or stop after 20 turns` appears AFTER the main block, AND the memory entry persists both blocks verbatim.

### Scenario 2 — plan-phase next SPEC (condition not met)

- **Given** the next SPEC is plan-phase (or lacks a machine-verifiable end-state),
- **When** the handoff is emitted,
- **Then** no follow-up block and no instruction line are emitted; the output is byte-identical to the pre-existing no-`/goal` form.

### Scenario 3 — resumed session reminder

- **Given** a resumed session whose handoff memory entry records a post-paste `/goal` follow-up block (persisted verbatim per § Auto-Memory Integration — the pasted main block itself carries no `/goal` reference; alternatively the orchestrator re-derives the emission condition from the resumed SPEC's run-phase + machine-verifiable end-state),
- **When** the orchestrator processes the paste and later passes Implementation Kickoff Approval,
- **Then** the orchestrator issues a natural-language reminder (NOT AskUserQuestion) that the user must send the `/goal` line as a standalone message now, because the model cannot set the goal itself.

### Scenario 4 — edge: user pastes `/goal` inside the main body anyway

- **Given** a user mistakenly leaves a `/goal` line inside the main pasted body,
- **When** the doctrine is consulted,
- **Then** the Paste-Time Activation Matrix and the new anti-pattern document that the line is inert plain text (input-start-only parsing), so the orchestrator falls back to the reminder obligation rather than assuming an armed goal.

## §D.2 Quality Gates / Definition of Done

- All AC-GF-001..013 sub-items PASS with verbatim command outputs recorded (verification-claim-integrity 5-section format; no summarized evidence).
- Single-turn parallel verification batch (greps + diffs + build) per agent-common-protocol § Parallel Execution.
- Commits pathspec-limited; working-tree unrelated files untouched.
- LSP/lint: no Go changes expected → `golangci-lint` delta 0; markdown surfaces not lint-gated.
- Gaps/Residual-risk explicitly enumerated (e.g., ja/zh naturalization quality is human-review residual; runtime behavior of a REAL paste is not machine-testable here — doctrine-layer fix only).
