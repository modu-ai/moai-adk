---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Acceptance criteria — card sessions start inside their worktree"
version: "0.1.0"
---

# Acceptance — SPEC-SESSION-DOUBLELOAD-001

Every criterion is binary and is checked by the command shown, run from the worktree root. `E` = `.moai/reports/t1219/m1-measure.md`. `T` = `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`. `L` = `.claude/rules/moai/workflow/kanban-dispatch.md`. `G` = `.claude/rules/local/gitflow-lane-protocol.md`.

## §D AC Matrix

### Measurement (M1)

- **AC-SDL-001** (REQ-SDL-001) — Given the branch history, When `git log --format=%H --diff-filter=A -- E` and `git log --format=%H -1 -- T` are compared, Then the commit adding E is an ancestor of the first commit touching T (`git merge-base --is-ancestor <E-commit> <T-commit>` exits 0).
- **AC-SDL-002** (REQ-SDL-002) — Given E, When `grep -cE '^(skills_dirs|instruction_paths|prompt_tokens)_(A|B|C): ([0-9]+|gap)$' E` runs, Then it prints `9`.
- **AC-SDL-003** (REQ-SDL-003) — Given E, When `grep -nE '^caps: probes<=8 turns<=4 timeout=300s wall<=45m$' E` and `grep -nE '^probes_run: [1-8]$' E` run, Then each prints exactly one line, and the caps line number is smaller than the line number of the first `skills_dirs_` line.
- **AC-SDL-004** (REQ-SDL-004) — Given E, When any field from AC-SDL-002 or AC-SDL-005 has value `gap`, Then `grep -c '^## Gaps' E` prints `1` and the Gaps section names that field (`grep -A50 '^## Gaps' E | grep -c '<field-name>'` ≥ 1).
- **AC-SDL-005** (REQ-SDL-005) — Given E, When `grep -E '^clear_restores_single_load: (yes|no|gap)$' E` runs, Then it prints exactly one line.
- **AC-SDL-006** (REQ-SDL-006) — Given the probes have finished, When `git -C <primary> branch --show-current` is read before and after M1 and `pgrep -f 'claude -p'` is run from the probe shell, Then the branch is unchanged and no probe process remains; E records both reads verbatim.
- **AC-SDL-007** (REQ-SDL-007) — Given E records a zero count for any `skills_dirs_*` or `instruction_paths_*` duplicate, When `grep -E '^positive_control: .*count=[1-9][0-9]*$' E` runs, Then it prints at least one line naming the probe file under `.moai/reports/t1219/probes/` that produced the non-zero count.

### Decision (M2)

- **AC-SDL-008** (REQ-SDL-008, REQ-SDL-009, REQ-SDL-010) — Given E, When `grep -E '^decision_primary: (launcher|none)$' E` and `grep -E '^decision_fallback: (clear|relaunch|none)$' E` run, Then each prints exactly one line, and the values are consistent with the recorded fields: `decision_fallback: clear` appears only if `clear_restores_single_load: yes`.

### Doctrine (M3, M4)

- **AC-SDL-009** (REQ-SDL-011) — Given `decision_primary: launcher`, When `grep 'wt. names the new card' T | grep -c 'moai cc -w'` runs, Then it prints `1` (baseline before this SPEC: `0`, the negative control).
- **AC-SDL-010** (REQ-SDL-011) — Given `decision_fallback` is `clear` or `relaunch`, When `grep -c '/clear' T` (for `clear`) or `grep -ci 'relaunch' T` (for `relaunch`) runs, Then the count is higher than on develop (`git show develop:<T-path> | grep -c …`).
- **AC-SDL-011** (REQ-SDL-012) — Given both copies, When `cmp T L` runs, Then it exits 0.
- **AC-SDL-012** (REQ-SDL-013) — Given T, When `grep -nE '\bt[0-9]{3,4}\b|SPEC-[A-Z][A-Z0-9-]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}' T` runs, Then it prints nothing; positive control: the same grep on `.moai/reports/t1219/verdict.md` prints at least one line. Additionally `go test ./internal/template/ -run 'Neutral|Leak'` exits 0.
- **AC-SDL-013** (REQ-SDL-014) — Given G, When `grep -n 'moai cc -w <card-id>' G` runs, Then the matched clause also names the fallback form (`grep -A3 'moai cc -w <card-id>' G | grep -cE '/clear|relaunch'` ≥ 1); and `ls internal/template/templates/.claude/rules/local/gitflow-lane-protocol.md` exits non-zero.
- **AC-SDL-014** (REQ-SDL-015) — Given T, When `grep -c '^\[HARD\]' T` runs, Then the count is ≥ `hard_clauses_before` recorded in E; and each of `git branch -m WT-<slug>`, `exit any previous one first`, `entered through the launcher` returns ≥ 1 from `grep -c` on T.

### Scope guard

- **AC-SDL-015** — Given the branch, When `git diff --stat develop...HEAD -- '*.go'` runs, Then it prints nothing; positive control: `git diff --stat develop...HEAD -- .moai/specs/SPEC-SESSION-DOUBLELOAD-001` prints a non-empty stat.
- **AC-SDL-016** — Given the SPEC directory, When `moai spec lint .moai/specs/SPEC-SESSION-DOUBLELOAD-001` and `make build` run, Then both exit 0.

## §D.1 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-SDL-001 | REQ-SDL-001 | M1 |
| AC-SDL-002 | REQ-SDL-002 | M1 |
| AC-SDL-003 | REQ-SDL-003 | M1 |
| AC-SDL-004 | REQ-SDL-004 | M1 |
| AC-SDL-005 | REQ-SDL-005 | M1 |
| AC-SDL-006 | REQ-SDL-006 | M1 |
| AC-SDL-007 | REQ-SDL-007 | M1 |
| AC-SDL-008 | REQ-SDL-008, REQ-SDL-009, REQ-SDL-010 | M2 |
| AC-SDL-009 | REQ-SDL-011 | M3 |
| AC-SDL-010 | REQ-SDL-011 | M3 |
| AC-SDL-011 | REQ-SDL-012 | M3 |
| AC-SDL-012 | REQ-SDL-013 | M3 |
| AC-SDL-013 | REQ-SDL-014 | M4 |
| AC-SDL-014 | REQ-SDL-015 | M3 |
| AC-SDL-015 | REQ-SDL-015 | M5 |
| AC-SDL-016 | REQ-SDL-012 | M5 |

## §D.2 Edge Cases

- Path C shows no second skills directory but does show a second instruction set: doctrine changes for instruction files only; E states the skill finding with its positive control (AC-SDL-007, REQ-SDL-010).
- Launcher path cannot be driven within the caps: `skills_dirs_A: gap`; Path B serves as the nearest proxy and E says so; AC-SDL-004 applies.
- `/clear` measurement impossible headless: `clear_restores_single_load: gap` → `decision_fallback: relaunch`.
- Premise contradicted wholesale: `decision_primary: none`, `decision_fallback: none`; M3/M4 do not run; AC-SDL-009..014 are not applicable and the orchestrator receives a blocker.

## §D.3 Quality Gate

- Only changed packages are tested (`./internal/template/...`); the full suite is left to CI.
- Evidence persists under `.moai/reports/t1219/`, not `/tmp`.

## §D.4 Definition of Done

- AC-SDL-001..008 and AC-SDL-015..016 pass; AC-SDL-009..014 pass or are marked not applicable by the AC-SDL-008 decision.
- E carries the five-section report with a non-empty Gaps section if any field is `gap`.
- All commits name t1219; nothing pushed.
