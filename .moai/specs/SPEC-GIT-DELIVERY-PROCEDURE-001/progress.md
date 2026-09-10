# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file). Revision 0.1.1 after plan-audit iteration 1 (`.moai/reports/t622/plan-audit.md`, FAIL 0.67).
- **Tier**: M. **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree measured**: `b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` for scope-file citations; revision authored on HEAD `647460835` (scope files byte-identical to the base per the audit's `git diff --stat b412f8a33 HEAD`).
- **Pre-write self-checks executed** (0.1.0):
  - SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` on `SPEC-GIT-DELIVERY-PROCEDURE-001` → `PASS`.
  - SPEC ID dedup: `ls -d .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001` → `No such file or directory`, exit 1 (before creation).
  - Frontmatter: 12 canonical fields present. Extra fields: `tier` (listed in the schema's Optional Fields table); `era` and `related_specs` are not listed in that table but are carried by existing SPECs (for example SPEC-WORKTREE-BRANCH-GUARD-001), and the lifecycle audit accepted them. `phase` is a release target, not a stage token.
- **Counts after revision**: 15 REQ definitions, 16 AC definitions (Tier M ceilings 16 each, counted independently).
- **Open decisions (not decided here)**: OD-1 (Late-branch handling), OD-2 (auto-merge default source) — spec.md §C.
- **Deferred requirements**: REQ-GDP-006 (OD-2); REQ-GDP-009, 010, 011, 012 (OD-1).
- **Positive controls measured at plan time** (scratch files in the session scratchpad, not exported — run-phase re-measures into `.moai/reports/t622/run/`):
  - AC-GDP-001: widened absent detector `fetch.*(independent|para[l]lel|batch)|…` on the extracted base Synchronization section → line 3 (manager-git.md:156), exit 0. Order detector on the same section → exit 1 (red on base).
  - AC-GDP-002: `awk` extraction of the Pre-Spawn bash block → 9 lines. (a) fetch-without-rev-list lines → 1; (b) standalone rev-list → block line 5, exit 0; (c) joined line count → 0, exit 1. All three red on base. Session-list line present. Interpretation-matrix rows `^[|] ` in the section → 8 (0.1.0 recorded 9; the separator rows start `|-` and are not counted).
  - AC-GDP-004: `grep -c 'pr merge --squash --delete-branch'` on `delivery.md` → 2. `grep -n 'merge_method'` on `delivery.md` → exit 1 (source detector red on base).
  - AC-GDP-005: `gh pr merge[^|]*--squash` across the five local scope files and the `.toml` → manager-git.md 32·114, delivery.md 343·355, `.toml` 26·108; spec-workflow, spec-assembly, agent-common-protocol none.
  - AC-GDP-007: widened detector (checkout · switch · reset --hard · pull · merge · rebase · stash · mutating branch) with `git grep -n -E` at `b412f8a33` on the four template scope files → 20 lines: manager-git 10 (42, 96, 110, 119, 121, 122, 125, 127, 137, 160), spec-workflow 5 (50, 56, 58, 59, 62), spec-assembly 1 (336), delivery 4 (276, 323, 324, 328). The 0.1.0 detector counted 14.
  - AC-GDP-008: strict per-file reference counts spec-assembly 1 · spec-workflow 1 · delivery 0; loose counts 1 · 1 · 0 (equal). Heading list `### Late-Branch Invocation Pattern` → lookup on a copy of the current manager-git.md → 1, exit 0; on the fixture with that heading removed → 0, exit 1.
  - AC-GDP-010: `worktree add` detector → `main-checkout-branch-guard.md:38`. AC-GDP-012: `main_late_branch` → `manager-git.md:137`; `default skips GitHub Issue creation` in `SKILL.md` → 2.
  - AC-GDP-013: `diff` of the two `delivery.md` copies → exit 1; header-stripped body holds only the 275 / 278 / footer differences.
  - AC-GDP-015 controls on the revised spec.md, counted separately with `/usr/bin/grep -c -E`: SPEC-ID `SPEC-([A-Z][A-Z0-9]*-)+[0-9]{3}` → 18; REQ token `REQ-([A-Z][A-Z0-9]*-)+[0-9]{3}` → 27; date → 5; SHA-shaped word → 7. (0.1.0 recorded "4" for a combined SPEC-ID/date detector; that 4 was the date count alone and the single-segment SPEC-ID detector counted 0.)
- **Iteration-1 mutants re-run against the revised judges** (session scratchpad):
  - D2 mutant (`# step 1` trailing comment on the fetch line, joined form added to prose): block line 2 `git fetch origin main 2>&1  # step 1`, block line 5 standalone rev-list → a-lines 1, b-exit 0, c 0 → FAIL on all three checks. The 0.1.0 file-wide count on the same mutant was 2 (passed). A correctly fixed fixture (one joined line inside the block) → a-lines 0, b-exit 1, c 1 → PASS.
  - D4 mutant ("see the manager-git agent § Late-Branch Invocation Pattern") → strict 0, loose 1 → count mismatch → FAIL. A first attempt wrote `§` as a Latin-1 byte and the grep wrapper printed no count at all; rebuilt as UTF-8 (`file` → "Unicode text, UTF-8 text") before counting.
  - D3 mutant (`pull --rebase`, `branch -f`, `merge --ff-only` lines) → widened detector catches all 3 (exit 0); the 0.1.0 detector caught 0.
  - D5 mutant ("`git fetch` then … all in parallel" in one batch) → widened absent detector hits (exit 0) → FAIL. A correctly fixed sentence fixture → absent exit 1, order detector 1 line.
- **Lifecycle audit** (`mcp__moai__spec_audit`, project_root = this worktree, filter this SPEC; server build v3.2.0-rc.5 commit 84fa4ece4) after revision → total 1, modern_era_clean 1, drift_findings [].
- **plan_status**: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
