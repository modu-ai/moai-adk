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

### M2, M3, M4 — closing evidence

Closing commit (content): `bcf1e3ac29` (parent `ac415b5817`, the M1 commit; plan base `74ad197380`). Baseline for every row below: `git rev-parse --short HEAD` printed `bcf1e3ac29`, `git branch --show-current` printed `WT-workflow-cmd-index`, `git status --short` printed nothing, all before and after the measurements (this run, this tree). M2 and M3 landed in the one commit `bcf1e3ac29` (plan.md M2/M3 allow one commit), so `<M2-closing-sha>` = `<M3-closing-sha>` = `<closing-sha>` = `bcf1e3ac29` and the M2-narrowed C-5 and the whole-set C-5 are both run on it. Per the closing-run convention each git cell is the acceptance.md command with `bcf1e3ac29` substituted for `74ad197380`; the cells that read the working tree (C-3, C-12, R-8, R-9) were run on this tree. Exit-code convention as in M1: output printed means exit 0; no output on `git grep` means exit 1 by git grep semantics, not displayed (G-1); `git diff` without `--exit-code` and `yq` with an empty selection print nothing and are not shown to fail.

#### Claim 4 — AC-PRC-001, AC-SWF-002, AC-SWF-005 hold at the closing commit

Evidence:

C-1 — `git grep -c -E "/moai[: ]design([^[:alnum:]-]|$)" bcf1e3ac29 -- ':(glob)docs-site/content/*/workflow-commands/_index.md' ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md'` — no output; exit 1 by git grep semantics, not displayed. Expected: nothing. Match. (Positive control: the same command on `74ad197380` printed the 12 lines recorded under Claim 1.)

C-2 — `git grep -c -P "/moai[: ](?!(clean|codemaps|e2e|feedback|fix|gate|goal|gtd|harness|loop|mx|plan|project|review|run|sync|todo)(?![a-z0-9-]))[a-z][a-z0-9-]*" bcf1e3ac29 -- <same three globs>` — no output; exit 1 by git grep semantics, not displayed. Expected: nothing. Match.

C-3 — `git ls-files -- <same three globs>` — exit 0, working tree; 8 paths (expected 8, the four `_index.md` and four `_meta.yaml`; no `moai-design.md`):

```
docs-site/content/en/workflow-commands/_index.md
docs-site/content/en/workflow-commands/_meta.yaml
docs-site/content/ja/workflow-commands/_index.md
docs-site/content/ja/workflow-commands/_meta.yaml
docs-site/content/ko/workflow-commands/_index.md
docs-site/content/ko/workflow-commands/_meta.yaml
docs-site/content/zh/workflow-commands/_index.md
docs-site/content/zh/workflow-commands/_meta.yaml
```

C-5 (whole set) — `git grep -n -E "moai-design" bcf1e3ac29 -- docs-site/content docs-site/data README.md README.ko.md README.ja.md README.zh.md` — no output; exit 1 by git grep semantics, not displayed. Expected: nothing. Match.

C-5 narrowed to the M2 files — `git grep -n -E "moai-design" bcf1e3ac29 -- ':(glob)docs-site/content/*/workflow-commands/_meta.yaml' ':(glob)docs-site/content/*/workflow-commands/moai-design.md' docs-site/data` — no output; exit 1 by git grep semantics, not displayed. Expected: nothing. Match.

R-5 — `yq 'keys'` on the four `_meta.yaml` files, each exit 0; every locale printed the same key list: `index`, `"moai-project"`, `"moai-plan"`, `"moai-run"`, `"moai-sync"`, `"moai-harness"`, `"moai-goal"`, the comment line(s), `new_items` (ko carries one extra comment line; no `moai-design`). `git ls-files docs-site/content/{ko,en,ja,zh}/workflow-commands` printed, per locale, `_index.md`, `_meta.yaml`, and `moai-goal`, `moai-harness`, `moai-plan`, `moai-project`, `moai-run`, `moai-sync` `.md` (8 files per locale, 32 lines). Every key other than `index` and `new_items` names an existing `<key>.md`, and every page has a key. Match.

Baseline-attribution: all commands run this run against `bcf1e3ac29`; C-3 and R-5 read the working tree, equal to that commit (`git status --short` empty).

#### Claim 5 — AC-SWF-003 and AC-SWF-007 hold at the closing commit

Evidence:

C-4 — `git grep -c -E "^\| \[.?/moai " bcf1e3ac29 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'` — exit 0:

```
bcf1e3ac29:docs-site/content/en/workflow-commands/_index.md:20
bcf1e3ac29:docs-site/content/ja/workflow-commands/_index.md:20
bcf1e3ac29:docs-site/content/ko/workflow-commands/_index.md:21
bcf1e3ac29:docs-site/content/zh/workflow-commands/_index.md:20
```

Expected en 20, ja 20, ko 21, zh 20. Match.

C-12 — `awk '/^## /{s=($0 ~ /^## (All subcommands|全サブコマンド|전체 서브커맨드|全部子命令)/)} s && /^\| \[/{n[FILENAME]++} END{print n[ARGV[1]]+0, n[ARGV[2]]+0, n[ARGV[3]]+0, n[ARGV[4]]+0}' docs-site/content/en/workflow-commands/_index.md docs-site/content/ja/workflow-commands/_index.md docs-site/content/ko/workflow-commands/_index.md docs-site/content/zh/workflow-commands/_index.md` — exit 0 (working tree): `15 15 15 15`. Expected `15 15 15 15`. Match.

C-11 probe for 15 — `git grep -c -E "\(15\)|\(15 個\)|\(15개\)|\(15 个\)|routes 15 subcommands|15 個のサブコマンド|15개 서브커맨드|15 个子命令" bcf1e3ac29 -- ':(glob)docs-site/content/*/workflow-commands/_index.md'` — exit 0:

```
bcf1e3ac29:docs-site/content/en/workflow-commands/_index.md:2
bcf1e3ac29:docs-site/content/ja/workflow-commands/_index.md:2
bcf1e3ac29:docs-site/content/ko/workflow-commands/_index.md:2
bcf1e3ac29:docs-site/content/zh/workflow-commands/_index.md:2
```

Expected `:2` for each of the four files. Match. (The probe for 16 was not re-run at the closing commit; the heading and sentence edits are visible in the diff below.)

R-3 — `git show --stat --format= bcf1e3ac29` — exit 0: 14 files changed, 8 insertions(+), 468 deletions(-); lists all four `_index.md` paths in the one commit (`_index.md` en 5, ja 5, ko 6, zh 5 changed lines), the four `_meta.yaml` (2 deletions each), the four `moai-design.md` (121, 99, 121, 99 deletions), `docs-site/data/menu/main.yaml` (6 deletions), `docs-site/.locale-parity-baseline` (1 deletion). Match.

`_index.md` diff (`git diff 74ad197380 bcf1e3ac29 -- <four _index.md>`, exit 0, read in full): each locale deletes the one design row of the "All subcommands" table and changes `16` to `15` in the heading and in the sentence; ko additionally deletes the summary-table design row (`조건부`); the ko `/moai goal` summary row and every other row are unchanged; no row or sentence naming the command was added.

Baseline-attribution: this run, `bcf1e3ac29`; C-12 reads the working tree.

#### Claim 6 — AC-SWF-008 and AC-SWF-009 hold at the closing commit

Evidence:

C-13 — `git grep -n -E "/moai[: ]design([^[:alnum:]-]|$)" bcf1e3ac29 -- docs-site/data/menu/main.yaml` — no output; exit 1 by git grep semantics, not displayed. Expected nothing. Match.

C-14 — `git grep -n -E "moai-design" bcf1e3ac29 -- docs-site/.locale-parity-baseline` — no output; exit 1 by git grep semantics, not displayed. Expected nothing. Match.

R-8 — `yq '.main[] | select(.ref == "/workflow-commands") | .sub[] | select(.ref == "/workflow-commands/moai-design") | .name' docs-site/data/menu/main.yaml` — no output (working tree). Expected nothing. Match.

R-9 — `yq '.main[] | select(.ref == "/workflow-commands") | .sub | length' docs-site/data/menu/main.yaml` — exit 0 (working tree): `6`. Expected `6`. Match.

R-10 — `git diff --numstat 74ad197380 bcf1e3ac29 -- docs-site/.locale-parity-baseline` — exit 0: `0	1	docs-site/.locale-parity-baseline`. Expected that line. Match. The `main.yaml` hunk (`git diff 74ad197380 bcf1e3ac29 -- docs-site/data/menu/main.yaml`) removes exactly the six lines of the entry (`- name:`, four name-map lines, `ref`); the baseline hunk removes exactly line `workflow-commands/moai-design.md`.

Baseline-attribution: this run, `bcf1e3ac29`; R-8 and R-9 read the working tree.

#### Claim 7 — AC-SWF-004 (guard): the changed set is exactly the fourteen Scope paths and nothing else

Evidence:

R-4 — `git diff --name-status 74ad197380 bcf1e3ac29 -- docs-site README.md README.ko.md README.ja.md README.zh.md` — exit 0, 14 lines:

```
M	docs-site/.locale-parity-baseline
M	docs-site/content/en/workflow-commands/_index.md
M	docs-site/content/en/workflow-commands/_meta.yaml
D	docs-site/content/en/workflow-commands/moai-design.md
M	docs-site/content/ja/workflow-commands/_index.md
M	docs-site/content/ja/workflow-commands/_meta.yaml
D	docs-site/content/ja/workflow-commands/moai-design.md
M	docs-site/content/ko/workflow-commands/_index.md
M	docs-site/content/ko/workflow-commands/_meta.yaml
D	docs-site/content/ko/workflow-commands/moai-design.md
M	docs-site/content/zh/workflow-commands/_index.md
M	docs-site/content/zh/workflow-commands/_meta.yaml
D	docs-site/content/zh/workflow-commands/moai-design.md
M	docs-site/data/menu/main.yaml
```

Ten modified, four deleted; none of the six other workflow-commands pages per locale, `docs-site/vercel.json`, `docs-site/layouts/`, `docs-site/static/`, or the four READMEs appears. Match.

R-4 companion — `git diff --name-status 74ad197380 bcf1e3ac29 -- .claude .moai/docs .moai/config scripts internal cmd` — no output (exit 0, `git diff` without `--exit-code`). Expected nothing. Match. Control that the form can print: the acceptance.md control (`git diff --name-only 74ad197380 HEAD -- .moai/config .moai/specs` prints the SPEC files) was not re-run in this pass (see Gaps).

AC-SWF-009 numstat — `git diff --numstat 74ad197380 bcf1e3ac29 -- docs-site README.md README.ko.md README.ja.md README.zh.md` — exit 0:

```
0	1	docs-site/.locale-parity-baseline
2	3	docs-site/content/en/workflow-commands/_index.md
0	2	docs-site/content/en/workflow-commands/_meta.yaml
0	121	docs-site/content/en/workflow-commands/moai-design.md
2	3	docs-site/content/ja/workflow-commands/_index.md
0	2	docs-site/content/ja/workflow-commands/_meta.yaml
0	99	docs-site/content/ja/workflow-commands/moai-design.md
2	4	docs-site/content/ko/workflow-commands/_index.md
0	2	docs-site/content/ko/workflow-commands/_meta.yaml
0	121	docs-site/content/ko/workflow-commands/moai-design.md
2	3	docs-site/content/zh/workflow-commands/_index.md
0	2	docs-site/content/zh/workflow-commands/_meta.yaml
0	99	docs-site/content/zh/workflow-commands/moai-design.md
0	6	docs-site/data/menu/main.yaml
```

The baseline file is `0	1` (one deleted line, none added). Match.

Baseline-attribution: this run, diffs between `74ad197380` and `bcf1e3ac29`.

#### Claim 8 — bounded sweep: no tracked reference to the deleted page remains outside the Scope, except the recorded open item

Evidence:

`git grep -n -E "workflow-commands/moai-design|/moai design|moai:design" bcf1e3ac29 -- docs-site README.md README.ko.md README.ja.md README.zh.md` — no output; exit 1 by git grep semantics, not displayed.

`git grep -n -E "workflow-commands/moai-design|/moai design|moai:design" bcf1e3ac29 -- .moai/config` — exit 0:

```
bcf1e3ac29:.moai/config/sections/design.yaml:2:# Controls /moai design workflow, GAN loop, Claude Design integration, and brand context.
bcf1e3ac29:.moai/config/sections/design.yaml:36:  # Claude Design integration (path A of /moai design)
```

Both hits are the already-recorded open item O-5 (comment lines of `.moai/config/sections/design.yaml`), outside the fourteen Scope paths and unchanged by this SPEC (REQ-SWF-007; the R-4 companion above printed nothing for `.moai/config`). `docs-site/vercel.json` carries no match of this pattern (its `/design` redirects are REDIR's, Q15, a different pattern). No unaccounted hit.

Baseline-attribution: this run, `bcf1e3ac29`; pattern sweep, not a parse (acceptance.md G-9).

#### Claim 9 — the SPEC directory lints clean

Evidence: `moai spec lint --strict .moai/specs/SPEC-V3R6-DOCS-PREREWRITE-SWEEPF-001` — exit 0, stdout (the tool reported no failure):

```
✓ No findings — all SPEC documents are valid
```

Tool provenance: the installed `moai` build is `moai_cp/20261007_195927`, commit `7705b365e2`; `git rev-list --count --left-right 7705b365e2...bcf1e3ac29` printed `0	74`, so that commit is a strict ancestor of HEAD (74 commits behind) and the lint was not run by a build made from this tree (verification-claim-integrity §2.2: a gap under tool provenance, not a tree-built measurement). The M1-time lint printed one INFO finding (OwnershipTransitionUnmeasured on the first plan commit); this run printed none; no cause for the difference was investigated.

#### Claim 10 — Hugo build of the edited site is warning-free (reproduced by this run)

Evidence: `hugo --minify --gc --source docs-site --destination /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/94352e8a-6862-42bf-bba5-b9cfa0b1a4b4/scratchpad/hugo-out-t1727` — exit 0 (the tool reported no failure), stdout (summary):

```
hugo v0.160.1+extended+withdeploy darwin/arm64 BuildDate=2026-04-08T14:02:42Z VendorInfo=Homebrew
 Pages        │ 186 │ 184 │ 184 │ 184   (KO / EN / JA / ZH)
 Total in 7299 ms
```

No `WARN` or `ERROR` line was printed. `git status --short` printed nothing after the build (the output went outside the repository). This matches the curator report (Hugo v0.160.1, pages KO 186, EN 184, JA 184, ZH 184). Baseline-attribution: this run, working tree equal to `bcf1e3ac29`.

#### Gaps

- G-1 (carried): no-output `git grep` cells take exit 1 from git semantics, not displayed; `yq` and `git diff` no-output cells show no failure but their exit code is not displayed either.
- The `hns-oss-docs-verify` recipe (URL blacklist, Mermaid direction, 4-locale parity script, README parity, body-emoji scan) was not run.
- Rendered output was not inspected: the sidebar and the previous/next chain were verified by the Hugo build passing and the `main.yaml` hunk, not by reading the generated HTML (acceptance.md G-4).
- The R-4 control (`.moai/config .moai/specs` pathspec printing the SPEC files) and the probe for 16 at the closing commit were not re-run.
- The three board records were cited as given, not opened; the specialists' own run reports (curator, content author, translators) were not re-read beyond the Hugo figures they reported.
- The lint line is not a tree-built measurement (installed build `7705b365e2` is 74 commits behind HEAD).
- Pattern sweeps (C-1, C-2, C-5, the Claim 8 sweep) do not cover a spelling they do not match (acceptance.md G-9).

#### Residual-risk

- Inbound links to `/<locale>/workflow-commands/moai-design` return 404 until REDIR (Q15) or another card adds a redirect.
- "All subcommands (15)" is a true row count but an inexact "All": `/moai gtd` and `/moai todo` have pages and no row (O-6).
- The `.moai/config/sections/design.yaml` comments still name `/moai design` (O-5), outside this change set by REQ-SWF-007.
- The CI-side docs checks (the "Validate 4-locale parity" step, the full Hugo build in CI) have not run on this commit; this lane's Hugo build and grep cells are local observations.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-19
run_commit_sha: bcf1e3ac29
run_milestones: M1 `ac415b5817` (pre-flight, draft to in-progress); M2 and M3 `bcf1e3ac29` (page, navigation, baseline deletions and `_index.md` rows and counts, one commit); M4 closing evidence in the commit that carries this section
ac_pass_count: 7 (AC-PRC-001, AC-SWF-002, AC-SWF-003, AC-SWF-005, AC-SWF-007, AC-SWF-008, AC-SWF-009; each Then-clause matched in Claims 4 to 6)
ac_guard_count: 1 (AC-SWF-004 guard held, Claim 7)
ac_fail_count: 0
ac_superseded: AC-SWF-006 (not applicable, Q14 verdict)
total_run_phase_files: 14 content paths (10 modified, 4 deleted) plus this progress.md and spec.md status
preserve_list_post_run_count: 0 unintended changes (R-4 and its companion, Claim 7)
new_warnings_or_lints_introduced: none observed (`moai spec lint --strict` printed no findings; Hugo build printed no WARN or ERROR)
cross_platform_build: not applicable (documentation-only change, no Go source touched)
l44_pre_commit_fetch: not run (no fetch was issued by this milestone; the lane's push and sync state stay with the operator)
l44_post_push_fetch: not applicable (no push by this lane)
m1_to_mN_commit_strategy: M1 own commit; M2 and M3 one commit; M4 evidence one commit; no push, no PR, no amend, no force
repository_wide_test_verdict: PENDING on the CI run on the project's integration branch (not owned by this run)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-19
sync_commit_sha: pending-backfill
changelog_entry_position: CHANGELOG.md `[Unreleased]` / `### Changed`, first entry (B12 pre-emission grep for the SPEC id printed 0 before emission)
b12_ac_count: acceptance.md counter printed 9 (live=9 excluded=0 ambiguous=0, no reserved tokens in the file); the entry cites 9 numbered with AC-SWF-006 superseded (8 live per acceptance.md prose)
frontmatter_status_transitions: spec.md `in-progress` to `completed` on this single sync commit (`updated: 2026-10-19` already held that value)
readme_check: `git grep -n -E "moai[ :]design" HEAD -- README.md README.ko.md README.ja.md README.zh.md` printed nothing; READMEs untouched
