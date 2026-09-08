# SPEC-CODEX-DOCTOR-PATH-GUARD-001 — Progress

Card t570 · worktree `.claude/worktrees/t570` · branch `WT-codex-doctor-guard` · base `a4855f0b2`.

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier S; `acceptance.md` written
  at the lead's explicit request — Tier S normally inlines AC into `spec.md` §3).
- SPEC ID regex check executed as Bash at `a4855f0b2`:
  `ID="SPEC-CODEX-DOCTOR-PATH-GUARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`
- ID uniqueness: `ls .moai/specs | /usr/bin/grep -i "DOCTOR-PATH-GUARD"` → no match (rc=1).
- `development_mode: tdd` read this run from `.moai/config/sections/quality.yaml`.
- Requirements in GEARS notation; exclusions section carries four `### Out of Scope —` sub-headings.
- Status: `draft`. No production file touched by plan-phase.

### Plan-audit repair (2026-09-08, post-audit, still `draft`)

plan-audit returned PASS 0.87 with two blocking defects plus three coordinate corrections. All
repaired in place; every coordinate was independently re-derived in this tree before editing.

| Defect | Repair |
|---|---|
| D4 (major) — REQ-005/006/007 had no criterion carrier | Added AC-CDPG-005 (helper-name collision count) and AC-CDPG-007 (test-only scope, two-half git predicate); REQ-006 given an explicit no-AC-by-design section + named DoD carrier; added a REQ-to-carrier map. `/usr/bin/grep -nE '^### AC-' acceptance.md` now returns 7 rows covering REQ-001..007. |
| D5 (minor) — AC-001 RED-now cell claimed an executed revert | Rewritten in obligation tense; sibling cell in AC-002 swept likewise; AC-004 title changed to "to be executed"; a **Tense contract** paragraph added at the head of `acceptance.md`. |
| D1 — helper file coordinates wrong | `stubCodexHome` / `writeCodexHomeConfig` corrected to `doctor_codex_test.go:103/:122`; table now names four distinct files. Re-measured this run. |
| D2 — "three mutants" | Corrected to four logs (`bypass`, `blanket-wrap`, `reorder`, `seam`), with the seam mutant's uncommitted-working-tree status noted; t562's own §J sentence quoted verbatim in place of a count. |
| D3 — F6 credit misattributed | Corrected: t562's sync-audit F6 named the extended-length family and issued it forward; this card discharges it rather than discovering it. §F now cites F2 and F6 by `progress.md` line. |
| D6 (optional) — REQ-007 `Should` outside GEARS | Applied: rewritten as a `When … shall` unwanted-behaviour form with a mechanical predicate. |
| D7 (optional) — unpinned fixture | Applied: fixture pinned to `/Users/u/skills/probe/SKILL.md`, with a one-line non-vacuity note. |

None declined. Baselines measured at `a4855f0b2`: `type statRecorder` → exactly 1 declaration
(`doctor_codex_stale_skill_test.go:331`); working tree still carries no production change
(`git status --short` → only `.moai/` paths untracked).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
