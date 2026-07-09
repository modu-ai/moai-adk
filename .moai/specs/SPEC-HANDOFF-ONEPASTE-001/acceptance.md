---
id: SPEC-HANDOFF-ONEPASTE-001
title: "Session Handoff 1-Paste — acceptance criteria"
version: "0.1.1"
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
| AC-OP-007 | REQ-OP-003/012 | windowed grep: auto-flow section completeness (gate/boundary/precondition tokens) | mechanical (windowed) |
| AC-OP-008 | REQ-OP-005 | grep: two-step fallback retained | mechanical |
| AC-OP-009 | REQ-OP-006 | windowed grep: memory-pruning revision in § Auto-Memory Integration | mechanical (windowed) |
| AC-OP-010 | REQ-OP-008 | grep: goal-directive cross-ref (baseline 3) | mechanical (delta) |
| AC-OP-011 | REQ-OP-015 | git diff: zero `.go` changes | mechanical |
| AC-OP-012 | REQ-OP-002 | grep: fail-open clause proximity | mechanical |
| AC-OP-013 | REQ-OP-014 | diff-hunk: structural blocks untouched (authoring constraint keeps zero-match exact) | mechanical |
| AC-OP-014 | REQ-OP-001/003 | scenario S1 (doctrine walkthrough) | review |
| AC-OP-015 | REQ-OP-004 | windowed grep: ultrathink guidance + A2 caveat in auto-flow section | mechanical (windowed) |
| AC-OP-016 | REQ-OP-013 | scenario S3 (manual reversion) | review |
| AC-OP-017 | REQ-OP-011 | diff -q: session-handoff.md live ↔ template mirror byte-identical + mirror CI test PASS | mechanical |

## §D.1 Severity

- **MUST-PASS (blocker)**: AC-OP-001, 003, 004, 005, 006, 007, 008, 011, 017 — emission obligation, config split, neutrality, lint, gate invariant, fallback retention, zero-Go, byte-parity mirror.
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

Windowing uses flag-based awk (`awk '/^## <heading>/{f=1;next} f&&/^## /{f=0} f'`) rather than an awk range pattern — a range `/X/,/^## /` self-terminates on the section's own `## ` heading line (start and end pattern matching the same record yields a single-line range), which would silently window nothing.

```bash
mkdir -p /tmp/moai-verify

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

# AC-OP-007 — auto-flow section completeness (all greps windowed to the new section body — D6/D8/D9)
awk '/^## Auto-Injected Resume Flow/{f=1;next} f&&/^## /{f=0} f' .claude/rules/moai/workflow/session-handoff.md > /tmp/moai-verify/op-flow-section.txt
test -s /tmp/moai-verify/op-flow-section.txt; echo "exit=$?"                                  # section exists, non-empty
grep -c "Implementation Kickoff Approval" /tmp/moai-verify/op-flow-section.txt               # ≥ 1 (REQ-OP-012, windowed — file-wide baseline 9 makes unwindowed grep vacuous)
grep -cE "startup|resume|compact" /tmp/moai-verify/op-flow-section.txt                        # ≥ 1 (/clear-only boundary — D8)
grep -ci "notice-only" /tmp/moai-verify/op-flow-section.txt                                   # ≥ 1 (D8)
grep -ci "precondition" /tmp/moai-verify/op-flow-section.txt                                  # ≥ 1 (resumed-turn verification — D9)

# AC-OP-008 — two-step fallback retained
grep -c "## Post-Paste /goal Follow-up Block" .claude/rules/moai/workflow/session-handoff.md # == 1
grep -c "Paste-Time Activation Matrix" .claude/rules/moai/workflow/session-handoff.md        # ≥ 1

# AC-OP-009 — memory pruning revision, windowed to § Auto-Memory Integration (D3)
awk '/^## Auto-Memory Integration/{f=1;next} f&&/^## /{f=0} f' .claude/rules/moai/workflow/session-handoff.md > /tmp/moai-verify/op-automem-section.txt
grep -c "consumed/" /tmp/moai-verify/op-automem-section.txt                                   # ≥ 1 (audit-trail ownership; file-wide baseline 0)
grep -ciE "one-line summary|prun" /tmp/moai-verify/op-automem-section.txt                     # ≥ 1 (pruning-directive token)

# AC-OP-010 — goal-directive cross-ref (measured baseline 3, NOT 0 — D4)
grep -c "handoff" .claude/rules/moai/workflow/goal-directive.md                              # > 3 (i.e., ≥ 4)
grep -cE "mode=auto|mode: auto" .claude/rules/moai/workflow/goal-directive.md                # ≥ 1 (baseline 0)

# AC-OP-011 — zero Go changes across the SPEC line
git diff --name-only <base>..HEAD -- '*.go' | wc -l                                          # == 0

# AC-OP-012 — fail-open clause proximity
grep -A6 "moai handoff save --stdin" .claude/rules/moai/workflow/session-handoff.md | grep -ci "fail-open"   # ≥ 1

# AC-OP-013 — structural blocks untouched (fully mechanical per plan-audit Q1 resolution: new prose
# avoids the literal ✂ glyph and the "Cut-line top/bottom text" strings — plan.md §D authoring
# constraint — so zero-match is exact; AC-OP-017's diff -q additionally pins the whole file)
git diff -U0 <base>..HEAD -- .claude/rules/moai/workflow/session-handoff.md | grep -cE "^[-+][^-+].*(✂────|Cut-line top text|Cut-line bottom text)"   # == 0

# AC-OP-015 — ultrathink guidance + A2 caveat, windowed to the auto-flow section (D2 —
# file-wide baselines are 16 ("ultrathink") and 1 ("not documented to fire"), so file-wide greps are vacuous)
grep -c "ultrathink" /tmp/moai-verify/op-flow-section.txt                                     # ≥ 1
grep -ci "not documented to fire" /tmp/moai-verify/op-flow-section.txt                        # ≥ 1

# AC-OP-017 — byte-parity mirror (MUST-PASS — D1)
diff -q .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md; echo "exit=$?"   # identical (no output, exit 0)
go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak'                                                    # PASS
```

## §D.4 Edge Cases

- **Stale pending record**: covered by existing Go (auto-mode TTL cleanup precedence N1). Doctrine cites it; no AC beyond S1 citation accuracy.
- **Two sessions racing on `/clear`**: claim-then-inject rename means exactly one session injects; the losing session has already read the pending record but its claim-rename fails (any errno), so it skips injection fail-open (`handoff_inject.go` `claimAndInject`). Doctrine cites REQ-AUTORESUME-012/013.
- **Body near 10,000-char cap**: Diet Constraints keep the body far below; flow section notes the cap (A1). No mechanical AC (render layer out of scope).
- **`--lang` mismatch**: `--lang` snapshots `conversation_language` at save time; the injected body is verbatim, so no re-localization occurs. Doctrine states the body is injected verbatim.

## §D.5 Quality Gates

- `moai spec lint` clean on spec.md (AC-OP-006); no `LegacyEARSKeyword` findings (GEARS-only authoring).
- `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift|TestTemplateNoInternalContentLeak'` PASS (byte-parity + neutrality — D1, AC-OP-017 companion).
- `moai constitution validate` PASS (CONST-V3R2-152 generation-time verbatim persistence unchanged — D10).
- `make build` + `go build ./...` green after M1 template edits (embed integrity; no Go source diff).
- Each milestone commit `git show --stat` file list ⊆ plan.md §F change list (scope audit).
- CI `template-neutrality-check.yaml` green on template path changes.

## §D.6 Definition of Done

1. All MUST-PASS ACs green with verbatim command output recorded in progress.md §E.2 (verification-claim-integrity §3.2 — evidence, not summaries).
2. SHOULD-PASS ACs green or individually waived with recorded rationale.
3. Both drift sentinels (SSOT + moai.md §8) updated in the same M1 commit.
4. Local `handoff.yaml` = auto; template `handoff.yaml` = manual, byte-identical to base.
5. session-handoff.md live ↔ template mirror byte-identical (AC-OP-017), mirror staged in the SAME M1 commit; goal-directive.md + moai.md mirrors kept byte-identical.
6. Zero `.go` diffs across the SPEC line.
7. M1 + M2 commits pathspec-scoped on the shared checkout; no unrelated files swept in.

## §D.7 Forward-Looking Checks (sync-phase)

- CHANGELOG entry classifies this as doctrine+config (no runtime behavior change for distributed users — template default unchanged).
- Memory topic file for this SPEC applies the REQ-OP-006 pruning convention to its own resume block (dogfood the new rule at first close).
