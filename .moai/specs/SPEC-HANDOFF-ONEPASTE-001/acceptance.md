---
id: SPEC-HANDOFF-ONEPASTE-001
title: "Session Handoff 1-Paste — acceptance criteria"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow + .moai/config/sections"
lifecycle: spec-anchored
tags: "handoff, session, auto-resume, doctrine, config, goal-first"
---

# SPEC-HANDOFF-ONEPASTE-001 — Acceptance Criteria

## §D AC Matrix

| AC | REQ | Verification | Class |
|----|-----|--------------|-------|
| AC-OP-001 | REQ-OP-001 | grep: emission clause in SSOT | mechanical |
| AC-OP-002 | REQ-OP-007 | grep: render-surface parity | mechanical |
| AC-OP-003 | REQ-OP-009 | grep: local handoff.yaml `mode: auto` | mechanical |
| AC-OP-004 | REQ-OP-010 | grep: template handoff.yaml `mode: manual` | mechanical |
| AC-OP-005 | REQ-OP-011 | grep: template neutrality (0 matches) | mechanical |
| AC-OP-006 | (all) | `moai spec lint` clean | mechanical |
| AC-OP-007 | REQ-OP-003/012 | grep: auto-flow section + Kickoff gate clause | mechanical (delta) |
| AC-OP-008 | REQ-OP-005 | grep: two-step fallback retained | mechanical |
| AC-OP-009 | REQ-OP-006 | grep: memory-pruning revision | mechanical |
| AC-OP-010 | REQ-OP-008 | grep: goal-directive cross-ref | mechanical (delta) |
| AC-OP-011 | REQ-OP-015 | git diff: zero `.go` changes | mechanical |
| AC-OP-012 | REQ-OP-002 | grep: fail-open clause proximity | mechanical |
| AC-OP-013 | REQ-OP-014 | diff-hunk: structural blocks untouched | semi-mechanical |
| AC-OP-014 | REQ-OP-001/003 | scenario S1 (doctrine walkthrough) | review |
| AC-OP-015 | REQ-OP-004 | grep: ultrathink guidance + A2 caveat | mechanical |
| AC-OP-016 | REQ-OP-013 | scenario S3 (manual reversion) | review |

## §D.1 Severity

- **MUST-PASS (blocker)**: AC-OP-001, 003, 004, 005, 006, 007, 008, 011 — emission obligation, config split, neutrality, lint, gate invariant, fallback retention, zero-Go.
- **SHOULD-PASS**: AC-OP-002, 009, 010, 012, 013, 015 — parity, pruning, cross-ref, fail-open text, structural-untouched, effort guidance.
- **REVIEW**: AC-OP-014, 016 — Given-When-Then walkthroughs (doctrine coherence; no runtime execution required since zero Go changes).

## §D.2 Given-When-Then Scenarios

### S1 — Auto-mode happy path, goal-first (AC-OP-014)

- **Given** `handoff.mode: auto` locally, and the orchestrator has just emitted a paste-ready resume for a run-phase next SPEC with a machine-verifiable end-state,
- **When** the doctrine-following orchestrator completes the emission turn,
- **Then** the doctrine requires it to have run `moai handoff save --stdin --spec <ID> --phase run --goal "<condition>" [--ultrathink] [--lang ko] [--session <uuid>]` with the cut-line-bounded main block on stdin (REQ-OP-001),
- **And when** the user runs `/clear`,
- **Then** the SessionStart handler claim-renames `pending.json` to `consumed/` and injects the body + directive-restoration guidance as `additionalContext` (existing verified Go — cited, not re-implemented),
- **And when** the user sends the single message `/goal <condition>`,
- **Then** the resumed session has the resume state (injected) + the goal loop armed in ONE user message, and the doctrine still requires Implementation Kickoff Approval before run-phase entry (REQ-OP-012).

### S2 — Fail-open: CLI absent or save error (AC-OP-012)

- **Given** the orchestrator is emitting a paste-ready resume and `moai` is not on PATH (or `handoff save` exits non-zero),
- **When** the save step fails,
- **Then** the doctrine requires the paste-ready surface to be emitted unchanged (cut-line block + memory persistence + optional two-step `/goal` follow-up), with no blocking, no retry loop, and no altered output — the manual path is fully functional (REQ-OP-002).

### S3 — Manual reversion byte-identical (AC-OP-016)

- **Given** a user (or the maintainer) restores `mode: manual` in `.moai/config/sections/handoff.yaml`,
- **When** the next session starts (any source, including `/clear`),
- **Then** the injector is a pure no-op (verified REQ-AUTORESUME-009 behavior; `pending.json` untouched even when stale), and the doctrine's manual path (6-block paste + two-step `/goal` follow-up per § Post-Paste /goal Follow-up Block) is complete without reference to the auto-mode section (REQ-OP-013, REQ-OP-005).

### S4 — Non-goal next SPEC (approval-message variant)

- **Given** `handoff.mode: auto` and the next SPEC is plan-phase (or lacks a machine-verifiable end-state),
- **When** the orchestrator emits + saves the handoff (no `--goal` flag per the emission condition) and the user runs `/clear`,
- **Then** the injected guidance leads the user to send ONE short approval message including the `ultrathink` keyword (REQ-OP-003/004), and no `/goal` line is involved anywhere.

## §D.3 Machine-Verification Commands (run-phase closure batch)

Baseline-delta discipline: baselines recorded at pre-flight (plan.md §C step 3); "delta" ACs assert post-edit count > baseline, preventing vacuous pre-existing-token passes.

```bash
# AC-OP-001 — emission clause in SSOT (baseline 0 → ≥1); [HARD] within the clause's section
grep -c "moai handoff save --stdin" .claude/rules/moai/workflow/session-handoff.md          # ≥ 1
grep -B3 -A3 "moai handoff save --stdin" .claude/rules/moai/workflow/session-handoff.md | grep -c "\[HARD\]"   # ≥ 1

# AC-OP-002 — render-surface parity (baseline 0 → ≥1)
grep -c "moai handoff save" .claude/output-styles/moai/moai.md                               # ≥ 1

# AC-OP-003 — local config auto
grep -cE "^\s*mode: auto" .moai/config/sections/handoff.yaml                                 # == 1

# AC-OP-004 — template config manual (unchanged)
grep -cE "^\s*mode: manual" internal/template/templates/.moai/config/sections/handoff.yaml   # == 1
git diff --name-only <base>..HEAD -- internal/template/templates/.moai/config/sections/handoff.yaml | wc -l   # == 0

# AC-OP-005 — template neutrality (SPEC token + measurement anecdote absent from mirrors)
grep -rn "SPEC-HANDOFF-ONEPASTE" internal/template/templates/ | wc -l                        # == 0
grep -rn "82%" internal/template/templates/.claude/rules/moai/workflow/session-handoff.md | wc -l   # == 0

# AC-OP-006 — spec lint clean
moai spec lint .moai/specs/SPEC-HANDOFF-ONEPASTE-001/spec.md                                 # "No findings"

# AC-OP-007 — auto-flow section + Kickoff gate (delta vs pre-flight baseline N₁)
grep -c "Auto-Injected Resume Flow" .claude/rules/moai/workflow/session-handoff.md           # ≥ 1
grep -c "Implementation Kickoff Approval" .claude/rules/moai/workflow/session-handoff.md     # > N₁ (baseline)

# AC-OP-008 — two-step fallback retained
grep -c "## Post-Paste /goal Follow-up Block" .claude/rules/moai/workflow/session-handoff.md # == 1
grep -c "Paste-Time Activation Matrix" .claude/rules/moai/workflow/session-handoff.md        # ≥ 1

# AC-OP-009 — memory pruning revision (consumed/ ownership named in § Auto-Memory Integration)
grep -c "consumed/" .claude/rules/moai/workflow/session-handoff.md                           # ≥ 1 (delta vs baseline)

# AC-OP-010 — goal-directive cross-ref (delta: baseline 0 → ≥1)
grep -c "handoff" .claude/rules/moai/workflow/goal-directive.md                              # > baseline
grep -c "mode=auto\|mode: auto" .claude/rules/moai/workflow/goal-directive.md                # ≥ 1

# AC-OP-011 — zero Go changes across the SPEC line
git diff --name-only <base>..HEAD -- '*.go' | wc -l                                          # == 0

# AC-OP-012 — fail-open clause proximity
grep -A6 "moai handoff save --stdin" .claude/rules/moai/workflow/session-handoff.md | grep -ci "fail-open"   # ≥ 1

# AC-OP-013 — structural blocks untouched (no diff line touches cut-line/skeleton/localization literals)
git diff -U0 <base>..HEAD -- .claude/rules/moai/workflow/session-handoff.md | grep -cE "^[-+][^-+].*(✂────|Cut-line top text|Cut-line bottom text)"   # == 0
# (New additive sections MAY cite markers in prose; this check targets the specification tables/skeleton lines. Reviewer confirms hunk placement.)

# AC-OP-015 — ultrathink guidance + A2 caveat in the auto-flow section
grep -c "ultrathink" .claude/rules/moai/workflow/session-handoff.md                          # > baseline
grep -ci "not documented to fire" .claude/rules/moai/workflow/session-handoff.md             # ≥ 1 (pre-existing caveat cited or restated in auto-flow)
```

## §D.4 Edge Cases

- **Stale pending record**: covered by existing Go (auto-mode TTL cleanup precedence N1). Doctrine cites it; no AC beyond S1 citation accuracy.
- **Two sessions racing on `/clear`**: claim-then-inject rename means exactly one session injects; the loser sees no pending. Doctrine cites REQ-AUTORESUME-012/013.
- **Body near 10,000-char cap**: Diet Constraints keep the body far below; flow section notes the cap (A1). No mechanical AC (render layer out of scope).
- **`--lang` mismatch**: `--lang` snapshots `conversation_language` at save time; the injected body is verbatim, so no re-localization occurs. Doctrine states the body is injected verbatim.

## §D.5 Quality Gates

- `moai spec lint` clean on spec.md (AC-OP-006); no `LegacyEARSKeyword` findings (GEARS-only authoring).
- `make build` + `go build ./...` green after M2 template edits (embed integrity; no Go source diff).
- Each milestone commit `git show --stat` file list ⊆ plan.md §F change list (scope audit).
- CI `template-neutrality-check.yaml` green on template path changes.

## §D.6 Definition of Done

1. All MUST-PASS ACs green with verbatim command output recorded in progress.md §E.2 (verification-claim-integrity §3.2 — evidence, not summaries).
2. SHOULD-PASS ACs green or individually waived with recorded rationale.
3. Both drift sentinels (SSOT + moai.md §8) updated in the same M1 commit.
4. Local `handoff.yaml` = auto; template `handoff.yaml` = manual, byte-identical to base.
5. Zero `.go` diffs across the SPEC line.
6. M1 + M2 commits pathspec-scoped on the shared checkout; no unrelated files swept in.

## §D.7 Forward-Looking Checks (sync-phase)

- CHANGELOG entry classifies this as doctrine+config (no runtime behavior change for distributed users — template default unchanged).
- Memory topic file for this SPEC applies the REQ-OP-006 pruning convention to its own resume block (dogfood the new rule at first close).
