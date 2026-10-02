---
id: SPEC-CTX-TABLE1M-001
title: "Acceptance — context-window 1M-default correction"
version: "0.1.1"
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
---

# Acceptance — SPEC-CTX-TABLE1M-001

> Discipline: `.claude/rules/moai/development/verification-completeness.md`
> §2 (two-cell adoption) and §2.1 (RED-now four elements). Every
> release-blocking criterion carries a RED-now cell observed on the
> pre-implementation tree — worktree t1415, HEAD `c50da9c2f`, working tree
> clean, 2026-10-02 — plus a green-path cell naming the milestone that flips
> it. Criteria whose starting state cannot be red (invariants and procedure
> guards) are classified **guard**, carry their starting observation honestly,
> and are never recorded as release-blocking passes.

## Evidence Ledger (RED-now, measured this run, tree `c50da9c2f`)

Every command is single-invocation (no pipes, `&&`, `;`, or subshells). Exit
codes are recorded as their own field; zero-hit greps were exit-probed (`; echo
exit=$?` on a re-run — the probe does not alter grep's own stdout or exit
status). Each entry's swept set is exactly the one named file, which is
non-empty — sizes measured this run: cwm live and mirror 6,999 B each;
model-policy live and mirror 27,051 B each (`wc -c`).

| ID | Command | verbatim stdout | exit | swept set |
|----|---------|-----------------|------|-----------|
| E1 | `grep -c "Opus 4.8+ running with a 200K window" .claude/rules/moai/workflow/context-window-management.md` | `1` | 0 | cwm live (6,999 B) |
| E2 | `grep -c "Opus 4.8+ running with a 200K window" internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | `1` | 0 | cwm mirror (6,999 B) |
| E3 | `grep -c "unless \`sonnet\[1m\]\` is selected" .claude/rules/moai/workflow/context-window-management.md` | `1` | 0 | cwm live |
| E4 | `grep -c "unless \`sonnet\[1m\]\` is selected" internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | `1` | 0 | cwm mirror |
| E5 | `grep -c "CC 2.1.285" .claude/rules/moai/workflow/context-window-management.md` | `0` | 1 | cwm live |
| E6 | `grep -c "CC 2.1.285" internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | `0` | 1 | cwm mirror |
| E7 | `grep -c "or with \`CLAUDE_CODE_DISABLE_1M_CONTEXT=1\`, \`sonnet\` budgets 200K" .claude/rules/moai/development/model-policy.md` | `2` | 0 | model-policy live (27,051 B) |
| E8 | `grep -c "or with \`CLAUDE_CODE_DISABLE_1M_CONTEXT=1\`, \`sonnet\` budgets 200K" internal/template/templates/.claude/rules/moai/development/model-policy.md` | `2` | 0 | model-policy mirror |
| E9 | `grep -c "run with a 200K window on some providers" .claude/rules/moai/development/model-policy.md` | `1` | 0 | model-policy live |
| E10 | `grep -c "run with a 200K window on some providers" internal/template/templates/.claude/rules/moai/development/model-policy.md` | `1` | 0 | model-policy mirror |
| E11 | `grep -c "Opus on a 200K provider such as Amazon Bedrock" .claude/rules/moai/development/model-policy.md` | `1` | 0 | model-policy live |
| E12 | `grep -c "Opus on a 200K provider such as Amazon Bedrock" internal/template/templates/.claude/rules/moai/development/model-policy.md` | `1` | 0 | model-policy mirror |
| E13 | `grep -c "CC 2.1.285" .claude/rules/moai/development/model-policy.md` | `0` | 1 | model-policy live |
| E14 | `grep -c "CC 2.1.285" internal/template/templates/.claude/rules/moai/development/model-policy.md` | `0` | 1 | model-policy mirror |
| E15 | `grep -c "CLAUDE_CODE_DISABLE_1M_CONTEXT=1" .claude/rules/moai/workflow/context-window-management.md` | `1` | 0 | cwm live (anchor) |
| E16 | `grep -c "CLAUDE_CODE_DISABLE_1M_CONTEXT" .claude/rules/moai/development/model-policy.md` | `3` | 0 | model-policy live (anchor) |
| E17 | `cmp .claude/rules/moai/workflow/context-window-management.md internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | (no output) | 0 | both cwm copies |
| E18 | `cmp .claude/rules/moai/development/model-policy.md internal/template/templates/.claude/rules/moai/development/model-policy.md` | (no output) | 0 | both model-policy copies |
| E19 | `grep -c "takes the 200K row" .claude/rules/moai/workflow/context-window-management.md` | `1` | 0 | cwm live (anchor) |
| E20 | `wc -c .claude/rules/moai/workflow/context-window-management.md` | `6999 .claude/rules/moai/workflow/context-window-management.md` | 0 | cwm live (size baseline) |
| E21 | `grep -c "t1415" .claude/rules/moai/workflow/context-window-management.md` | `0` | 1 | cwm live (neutrality) |
| E22 | `grep -c "t1415" internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | `0` | 1 | cwm mirror (neutrality) |
| E23 | `grep -c "t1415" .claude/rules/moai/development/model-policy.md` | `0` | 1 | model-policy live (neutrality) |
| E24 | `grep -c "t1415" internal/template/templates/.claude/rules/moai/development/model-policy.md` | `0` | 1 | model-policy mirror (neutrality) |
| E25 | `grep -c "SPEC-CTX-TABLE" .claude/rules/moai/workflow/context-window-management.md` | `0` | 1 | cwm live (neutrality) |
| E26 | `grep -c "SPEC-CTX-TABLE" internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | `0` | 1 | cwm mirror (neutrality) |
| E27 | `grep -c "SPEC-CTX-TABLE" .claude/rules/moai/development/model-policy.md` | `0` | 1 | model-policy live (neutrality) |
| E28 | `grep -c "SPEC-CTX-TABLE" internal/template/templates/.claude/rules/moai/development/model-policy.md` | `0` | 1 | model-policy mirror (neutrality) |

Anchor baselines for the green cells: E15 (flag anchor, cwm — must stay 1),
E16 (flag anchor, model-policy — must stay 3: corrected lines 22, 66, 90 each
keep their flag mention), E19 (precedence sentence — must stay 1). The
tree-wide scope sweeps that bounded the defect set (recorded in plan §B / spec
§1.2): `grep -rn "budgets 200K" .claude internal --include='*.md'` → 4 hits
(model-policy lines 22/90 in both trees); `grep -rn "running with a 200K
window" ...` → 2 hits (cwm line 16 in both trees); lines 28/66 located by
full-file read of model-policy (the provider-window paragraph and its
referrer). No stale instance exists outside the four target files.

## Acceptance criteria

| AC | Class | Requirement | RED-now (ledger) | Green path |
|----|-------|-------------|------------------|------------|
| AC-CTM-001 | release-blocking | cwm pair corrected per REQ-CTM-001/002 | E1-E4 = stale clauses present (1/1/1/1, exit 0); E5-E6 = 2.1.285 fact absent (0/0, exit 1) | **M1 flips it**: E1-E4 → `0` (exit 1) on BOTH copies; E5-E6 → `≥1` (expected `1`, exit 0); anchors E15 → `1` (exit 0) and E19 → `1` (exit 0) on the live copy |
| AC-CTM-002 | release-blocking | model-policy pair corrected per REQ-CTM-003/004 | E7-E12 = stale claims present (2/2/1/1/1/1, exit 0); E13-E14 = 2.1.285 fact absent (0/0, exit 1) | **M2 flips it**: E7-E12 → `0` (exit 1) on BOTH copies; E13-E14 → `≥1` (expected `3` — corrected lines 22, 66, and 90 each carry a `CC 2.1.285` citation; line 66's reads "CC 2.1.285 / 2.1.287" and still matches the pattern), exit 0; anchor E16 → `3` (exit 0) |
| AC-CTM-003 | guard | Live/mirror byte parity survives the edit (REQ-CTM-005) | E17-E18: parity already holds pre-edit (no output, exit 0) — green at arrival by design; recorded so nobody mistakes this for a flipped gate | **M1/M2 verify it**: both `cmp` invocations exit 0 with no output after each pair's edit; a non-zero cmp mid-milestone is the expected transient (plan §F R1) |
| AC-CTM-004 | guard | Always-loaded byte budget (REQ-CTM-006) | Starting observation E20: `wc -c` on cwm live = `6999` (exit 0), this run, tree `c50da9c2f` — the budget bound is satisfied now and must remain satisfied after growth | **M1 produces the check**: post-edit `wc -c` ≤ `7099` (expected `7098`); projected model-policy ≈ 27688 B recorded, no duty (`paths:`-scoped) |
| AC-CTM-005 | guard | Precedence sentence untouched and coherent (REQ-CTM-002) | E19 = `1` (exit 0) — the sentence exists pre-edit | **M1 verifies it**: post-M1 grep returns `1` (exit 0) on the live copy — invariance, not change |
| AC-CTM-006 | guard | Embed refresh ran (REQ-CTM-007) | Not adoptable as RED (a passing build on the stale tree proves nothing about the new embeds) — the check is meaningful only post-edit | **M3 produces it**: `make build` exit 0 on the post-M2 tree; output bounded to exit code + tail |
| AC-CTM-007 | guard | Template neutrality of the corrected texts (REQ-CTM-008) | E21-E28: card id and SPEC id absent from all 4 files today (`0`, exit 1, ×8) — must remain 0 (the correction must not leak card provenance into mirrored rules) | **M2 re-checks it**: all 8 greps (2 patterns × 4 files) → `0` (exit 1); CC version citations (`CC 2.1.285`/`2.1.287`) are permitted upstream facts (plan §E.6) |
| AC-CTM-008 | guard | Docs-only scope held (REQ-CTM-008) | Working tree clean at `c50da9c2f` (measured, `git status --short` → empty) | **M2 closes it**: no tracked modification outside the 4 target files; zero `.go` paths (audit D1 — wording independent of plan-artifact commit timing) |

## Swept-set discipline (zero-hit cannot masquerade as a pass)

Per verification-completeness.md §1.1, a green with an empty swept set is
uninterpreted output. Therefore, for every absence-form green check above:

1. The command names exactly one file whose byte size was measured non-empty
   (6,999 B / 27,051 B baselines in the ledger header); a green `0` on a
   missing or gutted file fails the paired anchor.
2. Every absence check is paired with a presence anchor that must read `≥1`
   (exit 0) in the same verdict: E15 (cwm flag anchor), E16 (model-policy flag
   anchor, expected to stay exactly 3), E19 (precedence sentence), and the
   E5/E6/E13/E14 presence checks themselves (new fact must appear).
3. The pair (absence of stale + presence of new + anchor intact) is a
   rule-pair per §2's mutant probe: deleting a whole line to zero a stale
   grep breaks the anchors; adding the new citation without removing stale
   text fails the absence checks. Both one-sided mutants are closed.

## Final Batch (M3, one turn, parallel read-only calls)

1. AC-CTM-001 green greps (6 commands: E1-E6 forms).
2. AC-CTM-002 green greps (8 commands: E7-E14 forms).
3. AC-CTM-003: `cmp` ×2.
4. AC-CTM-004: `wc -c` ×2 (live files).
5. AC-CTM-005: E19 form.
6. AC-CTM-007: neutrality greps (8 commands).
7. AC-CTM-008: `git status --short`.
8. `make build` runs ALONE before this batch (VK-BUILD is mutating; never
   batched — verification-batch-pattern.md).

## Residual risks (named at adoption)

- The CC 2.1.285/2.1.287 release-note facts are cited from the leader's sweep
  U2, not re-fetched — a mis-citation there would propagate (mitigation:
  upstream facts are version-scoped and checkable against the cited CHANGELOG).
- Whether an explicit `[1m]` suffix overrides `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`
  where both are present is not established by the cited changelog bullets;
  P1's NEW text preserves (not strengthens) the pre-edit line's claim on that
  point (audit D5; tracked as decision-index.md Q1).
- The green-path expectation `E13/E14 → 3` depends on all three corrected
  lines (22, 66, 90) keeping a `CC 2.1.285` citation; plan §B's new text is
  deterministic, but a run-phase wording deviation could change the count —
  the criterion binds `≥1`, so a count of 1 or 2 is still a pass and a count
  of 0 is a hard fail.
- `make build` embeds the whole template tree; an unrelated foreign defect in
  another template would surface here as a build failure owned by no card
  (plan §F R5 — stop and report, do not fix from this card).

## Gaps (explicitly unobserved at plan close)

- No run-phase command has executed: every green cell above is a designed
  expectation, not an observation. M1/M2/M3 own the observations.
- The upstream release-note contents (2.1.285/2.1.287) were not fetched in
  plan phase; the card's factual source stands as cited.
- docs-site / README / CHANGELOG state was not swept; sync phase owns any
  surface beyond the four files.
