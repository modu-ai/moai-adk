# Progress — SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 (card t1727)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_revision: 0.1.3
plan_complete_at: 2026-10-19T02:14:25Z
tier: M
requirement_count: 8 live (10 numbered; REQ-SWF-005 and REQ-SWF-006 superseded)
acceptance_criterion_count: 8 live (9 numbered; AC-SWF-006 superseded)
file_count: 14 paths (10 edited: four `_index.md`, four `_meta.yaml`, `main.yaml`, `.locale-parity-baseline`; 4 deleted: the `moai-design.md` pages)
tree_pin: 74ad197380a0711cdb9816d4e37c94d29af730ea
spec_lint_strict: `moai spec lint --strict .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001` printed one INFO finding (OwnershipTransitionUnmeasured on the 0.1.0 plan commit, which has no Authored-By-Agent trailer; Info never changes the exit status) and `0 error(s), 0 warning(s)`. Judging build: installed `moai_cp/20261007_195927-1888-g7705b365e2`; `git rev-list --count --left-right 7705b365e2...HEAD` printed `0	69`, so that commit is a strict ancestor of HEAD (69 commits behind) and the lint was not run by a build made from this tree (verification-claim-integrity §2.2: a gap, not a tree-built measurement)
plan_audit_verdict_files: .moai/reports/t1727/plan-audit.md
plan_artifact_hash: pending

Repair 0.1.2 (after plan-audit iteration 1, MP-9): D1 M2 check scoped to the M2 files, D2 guard pathspec adds `.moai/config`, D3 `new_items` exempt in AC-SWF-005, D4 C-12 after M3; no REQ/AC added. The lint line above was re-run at 0.1.2 (see the repair report).

Repair 0.1.3 (delta audit iteration 2, D7): the M2-closing C-5 targets `<M2-closing-sha>`, the `74ad197380` form is the pre-deletion control, and a closing-run convention covers sibling cells; AC-SWF-005 green path reads M2, M3.

The plan is ready for the independent plan audit; it is not approved. The artifact hash is recorded by the lane from the runtime value after the audit.

Decision rows, from decision-index.md. All decided: Q13 and Q17 (operator, board record `d-20261019T015908Z-220a`), Q14 (operator, board record `d-20261019T021508Z-258e`, verdict: remove the page in all four locales, with its `main.yaml` entry and baseline line; relayed by the dispatching coordinator, acceptance.md G-10). O-1 to O-4 closed by the leader ruling `d-20261019T020706Z-1c58` (leader under operator delegation, not an operator verdict). Open observations, not decisions and not in this change: O-5 (`/moai design` in `.moai/config/sections/design.yaml` comments), O-6 (`/moai gtd` and `/moai todo` have no row under "All subcommands").

## §E.2 Run-phase Evidence

### M1 — pre-flight and the Q13 determination

Cited decision records (board): `d-20261019T015908Z-220a` (operator: Q13 replace where an equivalent exists, delete where none exists; Q17 all four locales in the same change), `d-20261019T021508Z-258e` (operator: Q14 remove the `moai-design` page, its `main.yaml` entry, and baseline line 51; relayed by the dispatching coordinator, acceptance.md G-10), `d-20261019T020706Z-1c58` (leader ruling under operator delegation, not an operator verdict: O-1 to O-4). The records themselves were not opened in this pass; they are cited as given in the run-phase delegation.

Kickoff: autonomous (card in run state); plan-audit PASS 0.94 @62edea0974 (local file `.moai/reports/t1727/plan-audit.md`, not re-read here).

Baseline for every row below: tree `HEAD` = `62edea0974` (run on this tree, this run); `git merge-base --is-ancestor 74ad197380 62edea0974` printed nothing (exit 0 by git semantics: 74ad197380 is an ancestor); `git diff --stat 74ad197380 62edea0974 -- docs-site .claude README.md README.ko.md README.ja.md README.zh.md scripts .github` printed nothing, so the working-tree cells read content equal to `74ad197380`; `git status --short` printed nothing before the edits. Tool provenance: only system `git` and `awk` were used; no `moai` tool measurement is cited except the lint line at the end. Installed `moai` build `7705b365e2`: `git rev-list --count --left-right 7705b365e2...62edea0974` printed `0	72`, so that commit is a strict ancestor of HEAD (72 commits behind; 69 at plan time) and any `moai` tool output here was not produced by a build made from this tree.

Exit-code convention: the shell tool displays no status. A cell that printed output is recorded as exit 0 (the tool reported no failure). A cell with no output takes exit 1 from git grep semantics (`git grep` exits 1 when nothing matches), "by git grep semantics, not displayed"; this is a Gap (G-1 of acceptance.md).

#### Claim 1 — acceptance.md cells C-1, C-2, C-4, C-8, C-11, C-12, C-13, C-14 reproduce unchanged on HEAD `62edea0974`; no drift

Evidence (command, verbatim stdout, exit code):

C-1 — `git grep -c -E "/moai[: ]design([^[:alnum:]-]|$)" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'` — exit 0:

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:1
74ad197380:docs-site/content/en/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/en/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/ja/workflow-commands/_index.md:1
74ad197380:docs-site/content/ja/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ja/workflow-commands/moai-design.md:4
74ad197380:docs-site/content/ko/workflow-commands/_index.md:2
74ad197380:docs-site/content/ko/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/ko/workflow-commands/moai-design.md:5
74ad197380:docs-site/content/zh/workflow-commands/_index.md:1
74ad197380:docs-site/content/zh/workflow-commands/_meta.yaml:1
74ad197380:docs-site/content/zh/workflow-commands/moai-design.md:4
```

C-2 — `git grep -c -P "/moai[: ](?!(clean|codemaps|e2e|feedback|fix|gate|goal|gtd|harness|loop|mx|plan|project|review|run|sync|todo)(?![a-z0-9-]))[a-z][a-z0-9-]*" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'` — exit 0; stdout identical, line for line, to C-1 (12 lines, same paths and counts: 1, 1, 5, 1, 1, 4, 2, 1, 5, 1, 1, 4).

C-3 (sweep control) — `git ls-files -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'` — exit 0; 12 paths printed (`_index.md`, `_meta.yaml`, `moai-design.md` for each of en, ja, ko, zh), equal to the recorded 12 lines.

C-4 — `git grep -c -E "^\| \[.?/moai " 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'` — exit 0:

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:21
74ad197380:docs-site/content/ja/workflow-commands/_index.md:21
74ad197380:docs-site/content/ko/workflow-commands/_index.md:23
74ad197380:docs-site/content/zh/workflow-commands/_index.md:21
```

C-8 — `git ls-tree -r --name-only 74ad197380 -- .claude/commands/moai` — exit 0; 17 lines: `clean`, `codemaps`, `e2e`, `feedback`, `fix`, `gate`, `goal`, `gtd`, `harness`, `loop`, `mx`, `plan`, `project`, `review`, `run`, `sync`, `todo` (each `.claude/commands/moai/<name>.md`); no `design`.

C-11 probe for 15 — `git grep -c -E "\(15\)|\(15 個\)|\(15개\)|\(15 个\)|routes 15 subcommands|15 個のサブコマンド|15개 서브커맨드|15 个子命令" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'` — no output; exit 1 by git grep semantics, not displayed.

C-11 control for 16 — `git grep -c -E "\(16\)|\(16 個\)|\(16개\)|\(16 个\)|routes 16 subcommands|16 個のサブコマンド|16개 서브커맨드|16 个子命令" 74ad197380 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'` — exit 0:

```
74ad197380:docs-site/content/en/workflow-commands/_index.md:2
74ad197380:docs-site/content/ja/workflow-commands/_index.md:2
74ad197380:docs-site/content/ko/workflow-commands/_index.md:2
74ad197380:docs-site/content/zh/workflow-commands/_index.md:2
```

C-12 — `awk '/^## /{s=($0 ~ /^## (All subcommands|全サブコマンド|전체 서브커맨드|全部子命令)/)} s && /^\| \[/{n[FILENAME]++} END{print n[ARGV[1]]+0, n[ARGV[2]]+0, n[ARGV[3]]+0, n[ARGV[4]]+0}' docs-site/content/en/workflow-commands/_index.md docs-site/content/ja/workflow-commands/_index.md docs-site/content/ko/workflow-commands/_index.md docs-site/content/zh/workflow-commands/_index.md` — exit 0 (working tree; docs-site equal to `74ad197380` per the baseline above): `16 16 16 16`

C-13 — `git grep -n -E "/moai[: ]design([^[:alnum:]-]|$)" 74ad197380 -- docs-site/data/menu/main.yaml` — exit 0:

```
74ad197380:docs-site/data/menu/main.yaml:204:          ko: /moai design
74ad197380:docs-site/data/menu/main.yaml:205:          en: /moai design
74ad197380:docs-site/data/menu/main.yaml:206:          ja: /moai design
74ad197380:docs-site/data/menu/main.yaml:207:          zh: /moai design
```

C-14 — `git grep -n -E "moai-design" 74ad197380 -- docs-site/.locale-parity-baseline` — exit 0: `74ad197380:docs-site/.locale-parity-baseline:51:workflow-commands/moai-design.md`

Compared with the recorded stdout and exit codes of acceptance.md section C: all eight cells (plus the C-3 and C-11-control companions) match byte for byte. No drift.

Baseline-attribution: each command above was run in this run, on HEAD `62edea0974`; `git` cells name `74ad197380` in the command and read git objects, so they reproduce as recorded; C-3 and C-12 read the working tree, equal to `74ad197380` for `docs-site` by the baseline above.

#### Claim 2 — Q13 determination holds: no registered command or CLI verb is a current equivalent of the design family

Evidence:

C-7 — `git grep -n -i -E "workflows/design|design phase|manager-design|plan.{1,3}design" 74ad197380 -- .claude/skills/moai/workflows/run.md .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/sync.md .claude/skills/moai/SKILL.md` — no output; exit 1 by git grep semantics, not displayed.

C-9 (positive control) — `git grep -n -E "user-invocable" 74ad197380 -- .claude/skills/moai/workflows/design.md` — exit 0: `74ad197380:.claude/skills/moai/workflows/design.md:11:user-invocable: false`

CLI verb check (added in this run) — `git grep -n -E "Use:[[:space:]]+\"design" 62edea0974 -- internal/cli cmd` — no output; exit 1 by git grep semantics, not displayed. The C-8 list above is the slash-command side.

Reading: the registered commands are the 17 of C-8, none is `design`; the design workflow skill is `user-invocable: false`; none of the four routing files names the design workflow; no cobra command declares a `design` verb. The Q13 branch is therefore "delete where none exists", as applied in plan.md M3. Nothing has appeared; no stop.

#### Gaps

- G-1: no-output cells (C-7, C-11 probe for 15, the CLI verb check, the ancestor and diff checks) take their exit code from git semantics; the tool does not display it and no wrapper was used (the worktree guard refuses compound forms).
- The CLI verb check is a pattern over `internal/cli` and `cmd` for a command declared as `Use: "design...`; a verb registered under another spelling or dynamically would not match. The installed `moai --help` output was not read.
- The three board records were not opened; they are cited as given in the delegation.
- The plan-audit verdict file was not re-read in this pass.
- Hugo build, the verify recipe, and the i18n parity script were not run (no edits exist yet).
- M2, M3, and M4 evidence (page, `_meta.yaml`, `main.yaml`, baseline, `_index.md` edits; R-4, R-5, R-8 to R-10; closing C-cells) will be appended by later steps; this section covers M1 only.

#### Residual-risk

- A cell pinned to `74ad197380` stays correct only while the git object is unchanged; the working-tree cells (C-3, C-12) were measured on a tree equal to `74ad197380` for `docs-site` and are re-run at the closing commit.
- The installed `moai` build is 72 commits behind HEAD, so the lint result below is a gap under tool provenance, not a tree-built measurement.
- A design-family spelling the grep patterns do not match (acceptance.md G-9) is not excluded by these re-runs.

#### Claim 3 — status transition `draft → in-progress` applied and the SPEC directory lints clean

Evidence: `spec.md` frontmatter `status: draft` changed to `status: in-progress` (`updated: 2026-10-19` already held that value, so it is unchanged); `plan.md` and `acceptance.md` carry no `status:` or `updated:` field (`git grep -n -E "^(updated|status):" 62edea0974 -- <plan.md> <acceptance.md> <decision-index.md>` printed nothing), so they are untouched. `moai spec lint --strict .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001` — exit 0 (printed `0 error(s), 0 warning(s)`), stdout:

```
INFO      OwnershipTransitionUnmeasured  .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001/spec.md  1     SPEC SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001 transition "(none)" → "draft" expected owner "manager-spec" but commit 5d16c25af8e250491104f25b6e3f73e6268d4001 (feat(SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001): plan-phase artifacts (Tier M, 5 artifacts) (card t1727)) has no Authored-By-Agent trailer — ownership transition unmeasured

0 error(s), 0 warning(s)
```

Baseline-attribution: run on the working tree with the two edits above, before the M1 commit; judging build is the installed `moai_cp/20261007_195927` (commit `7705b365e2`, a strict ancestor of HEAD, 72 commits behind), so this lint line is a gap under verification-claim-integrity §2.2, not a tree-built measurement. The one INFO finding concerns the first plan commit's missing trailer and is pre-existing; Info does not change the exit status. Gap: the status transition owner check will first measure the M1 commit (trailer `Authored-By-Agent: manager-develop`) after it lands.

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase (manager-develop)_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase (manager-docs)_
