auditor-model: claude-opus-5-5

# Sync re-audit — card t2105 (Class B, one SPEC frontmatter change)

verdict: PASS-WITH-DEBT
audited_sha: acfd92d3d78d1e63dc71ed1f5572810175c8c424
audit_receipt: cited in the final AUDIT-VERDICT line

- Score: 95/100 (weighted: Functionality 95 x0.40, Security 100 x0.25, Craft 90 x0.20, Consistency 95 x0.15)
- Base: 9cf8a1fc14 (ancestor of HEAD: `git merge-base --is-ancestor` exit 0). Scope: `git diff 9cf8a1fc14..HEAD`.
- Iteration: 2 (re-audit of the current HEAD after the D1 repair commit acfd92d3d7).
- Audit plan (`moai verify audit-plan --project-root <tree>`): `config_status: ok`, `model: codex`, gates claude off / codex required / glm off, `enforced_required: ["codex"]`. Cross-model leg: one `codex_audit` (native, baseBranch, delta_base 9cf8a1fc14), called after this file is committed; its receipt is cited in the final message, not here. The verdict above assumes a clean Codex pass meets the gate.
- debt: D2 dispose_in=t2107 Remove `lint.skip: [OwnershipTransitionInvalid]` from SPEC-FACTORY-COMPLETION-RECOVERY-001 once the rule fix lands, or at that SPEC's next status transition, whichever comes first. Until then the skip silences the whole rule for this SPEC. Follow-up card t2107 (picked) carries both the rule fix and this removal.

## Iteration history

| # | Audited SHA | Verdict | Score | Notes |
|---|---|---|---|---|
| 1 | 00cc223852 | PASS-WITH-DEBT | 90 | D1 (comment blamed the history; the real cause is a rule false positive on a squash merge), D2 (skip removal condition). Claude-only, installed build. |
| 2 | acfd92d3d7 | PASS-WITH-DEBT | 95 | D1 disposed (comment repaired; follow-up card t2107 exists). D2 carried to t2107. Tree-built binary, control reproduced, whole corpus run. |

## Findings (structured defect-list)

- F1 [Info] [optional] [in-diff] D1 from iteration 1 is disposed. The comment at `spec.md:15` now matches the commit facts (E3). Not blocking.
- F2 [Low] [optional, carried as D2] [in-diff] `spec.md:16-18`: the skip turns off the whole rule for this SPEC. The early return at `internal/spec/lint_ownership.go:377-388` sits before the git lookup. Exit conditions are written in the comment and in card t2107; nothing enforces them. Required fix: D2.
- F3 [Low] [optional] [in-diff] The D1 wording asked for the cause to be recorded in progress.md as well. This diff does not touch progress.md, and the record t2107 cites ("t2105 progress.md") is outside this tree's diff and was not read. Not blocking: the frontmatter comment and card t2107 both state the cause.
- F4 [Info] [optional] [outside] The first-trailer-wins parse (`parseAuthoredByAgent`, `FindStringSubmatch`) also affects other squash merges. Card t2107 owns this (it names 55b20438fd and 5ef49c4625 for RED).

No blocking finding. Both must-pass dimensions (Functionality, Security) pass.

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | Target lint `baseline: OK`, exit 0 (E4). Control without the skip: `EXCEEDED +1`, exit 1 (E5). Whole corpus: `baseline: OK`, exit 0 (E6). |
| Security (25%) | 100 | PASS | Frontmatter-only YAML change. No code, input surface or secret. A grep for token/secret/password over the SPEC frontmatter matched nothing. |
| Craft (20%) | 90 | PASS | `go test ./internal/spec -run 'Ownership\|LintSkip\|Skip' -count=1` → `ok ... 16.993s`. Exit condition stated. The skip is whole-rule (F2). |
| Consistency (15%) | 95 | PASS | Nested `lint:`/`skip:` form, which the decoder binds (`lint.go:637-638`). Corpus run reports no `LintSkipFlatForm` and no `ParseFailure` for this SPEC. |

## Claim

1. The card diff is the SPEC frontmatter (comment line, nested `lint.skip` block, `updated:`) plus this report. Nothing else changed. The baseline file is unchanged.
2. The skip has the shape the linter reads and is scoped to one rule and one SPEC.
3. The comment states the real cause accurately.
4. With the skip, the target SPEC and the whole corpus pass the baseline gate. Without it, the gate fails by exactly +1 `OwnershipTransitionInvalid` on this SPEC.
5. No other SPEC's lint result depends on this change.

## Evidence

E1: diff scope.
`git diff --stat 9cf8a1fc14..HEAD`:
```
 .moai/reports/t2105/sync-audit.md                  | 102 +++++++++++++++++++++
 .../SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md   |   6 +-
 2 files changed, 107 insertions(+), 1 deletion(-)
```
spec.md hunk: `-updated: 2026-11-02` / `+updated: 2026-11-06`, then after `tier: L`:
```
+# t2105 documented debt: lint false positive, not an ownership violation. Commit 55b20438fd is a squash merge (#1930) whose body carries many Authored-By-Agent trailers; the rule reads only the first (manager-spec), while the sub-commit that moved the status draft -> in-progress carries manager-develop, the expected owner. History is not rewritten. Remove this skip at this SPEC's next status transition (its sync commit); the rule's squash handling is to be fixed separately.
+lint:
+  skip:
+    - OwnershipTransitionInvalid
```
`git show --stat acfd92d3d7` touches only the spec.md line (1 insertion, 1 deletion). `git diff --stat 9cf8a1fc14..HEAD -- .moai/spec-lint-baseline.json` printed nothing. `git status --short` was empty at the start of the audit.

E2: skip parsing (this tree's source).
- `internal/spec/lint.go:637-638`: `LintConfig struct { Skip []string \`yaml:"skip"\` } \`yaml:"lint"\``. The nested form binds; the flat `lint.skip:` key would be dropped and flagged by `LintSkipFlatFormRule`.
- `lint.go:313`: `applylintSkip(ruleFindings, doc.LintSkip)`, applied per document. It drops only findings whose `Code` is in that document's list (`lint.go:432-447`).
- `lint_ownership.go:346-352`: `hasOwnershipSkipOptOut` matches exactly `"OwnershipTransitionInvalid"`. `Check` (377-388) then returns a single Info `OwnershipTransitionSkipped`.

E3: accuracy of the comment.
- `git log -1 --format='%H%n%s%nparents:%P' 55b20438fd` shows a single parent `2a1898ea22…` and a subject ending in `(#1930)`, so it is a squash merge.
- `git log -1 --format=%b 55b20438fd | grep -c 'Authored-By-Agent'` → `50`. The first match is body line 8, `Authored-By-Agent: manager-spec`.
- Body line 46 is the sub-commit `* feat(...): M2-B run consolidation operation and worker-liveness predicate`. Its line 53 reads `spec.md status draft -> in-progress (first run-phase M2-B commit).` and its trailer at line 59 reads `Authored-By-Agent: manager-develop`.
- `git show 55b20438fd -- <spec.md>` contains `-status: draft` / `+status: in-progress`.
- Parser: `parseAuthoredByAgent` uses `authoredByAgentLine.FindStringSubmatch(body)`, which returns the first match only (`lint_ownership.go:256-266`).
- The matrix (`spec-frontmatter-schema.md`) assigns `draft → in-progress` to manager-develop. Each fact in the comment checks out.

E4: target lint with a tree-built binary.
`<scratch>/moai-t2105 spec lint SPEC-FACTORY-COMPLETION-RECOVERY-001 --baseline .moai/spec-lint-baseline.json` (exit 0):
```
INFO      OwnershipTransitionSkipped  .../SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md  1     SPEC SPEC-FACTORY-COMPLETION-RECOVERY-001 ownership transition check skipped via lint.skip: [OwnershipTransitionInvalid] opt-out
0 error(s), 0 warning(s)
baseline: OK — .moai/spec-lint-baseline.json
  inventory: 0 warning(s) total (advisory included), 0 non-advisory tracked across 1 recorded rule(s)
```

E5: control without the skip. The control ran in a clone of this tree in the scratchpad, made with `git clone --no-hardlinks <tree> <scratch>/ctl-clone`. The clone's HEAD was `acfd92d3d78d1e63dc71ed1f5572810175c8c424`. Only the 3-line `lint:` block was removed from its spec.md, so the git history the rule reads is intact. The tree itself was never edited. Same binary, run from the clone root (exit 1):
```
WARNING   OwnershipTransitionInvalid  .../ctl-clone/.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md  1     SPEC SPEC-FACTORY-COMPLETION-RECOVERY-001 transition "draft" → "in-progress" expected owner "manager-develop" but commit 55b20438fd9d3cc13e671cd8cdbb43b2ebc2d611 (Authored-By-Agent: manag…
0 error(s), 1 warning(s)
baseline: EXCEEDED — .moai/spec-lint-baseline.json
  OwnershipTransitionInvalid: recorded 0 -> current 1 (+1)
  inventory: 1 warning(s) total (advisory included), 1 non-advisory tracked across 1 recorded rule(s)
  recorded at: d5638ba81 (2026-09-24)
```

E6: whole corpus at HEAD (1137 SPEC dirs).
`<scratch>/moai-t2105 spec lint --baseline .moai/spec-lint-baseline.json` (exit 0, about 33 minutes):
```
0 error(s), 8598 warning(s)
baseline: OK — .moai/spec-lint-baseline.json
  inventory: 8598 warning(s) total (advisory included), 4 non-advisory tracked across 1 recorded rule(s)
  recorded at: d5638ba81 (2026-09-24)
```
In that output, the target SPEC has exactly one line: `INFO OwnershipTransitionSkipped`. `OwnershipTransitionInvalid` appears 4 times, all on other SPECs. None of them is non-advisory, because the only tracked rule is the recorded one. No `LintSkipFlatForm` or `ParseFailure` finding appears; the single grep hit was a `VacuousTestAssertion` whose message contains the string "ParseFailure".

E7: the three SPECs from iteration 1, same binary.
`spec lint SPEC-PROGRESS-RECORD-IO-001 SPEC-FACTORY-COMPLETION-RECOVERY-001 SPEC-LSPMCP-001 --baseline …` (exit 0):
```
WARNING   OwnershipTransitionInvalid  .moai/specs/SPEC-PROGRESS-RECORD-IO-001/spec.md  1  … commit 5ef49c4625… (Autho…
INFO      OwnershipTransitionSkipped  .moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md  1  …
WARNING   OwnershipTransitionInvalid  .moai/specs/SPEC-LSPMCP-001/spec.md  1  … commit 94d9fccfc7… (Authored-By-Agent: manager-devel…
0 error(s), 2 warning(s)
baseline: OK — .moai/spec-lint-baseline.json
  decreased: TierArtifactMissing recorded 4 -> current 0 (-4)
  inventory: 2 warning(s) total (advisory included), 0 non-advisory tracked across 1 recorded rule(s)
```
This is identical to iteration 1's E2. The other two SPECs' results are unchanged.

E8: the follow-up card exists. `moai todo list | grep -i squash` matched:
`t2107 picked [P1·Class B·t2105 후속] spec lint OwnershipTransitionInvalid 규칙 오탐: 다중 트레일러 squash 병합 커밋(예 55b20438fd #1930, …) … 착지 뒤 SPEC-FACTORY-COMPLETION-RECOVERY-001의 lint.skip(t2105 임시 조치) 제거. t2087(lint_ownership.go 수정) 착지 뒤 시작.`

E9: binding run conditions. `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md:17` carries a `Binding run conditions` ledger. Line 205 records that four items were moved to `dispose_in` the final SPEC-close sync, with the leader's approval recorded at line 173. This card is not that SPEC's run or sync; it does not touch progress.md, and it neither disposes of those items nor changes them. They stay owed at the final SPEC close, outside this card's verdict.

## Baseline-attribution

- Tree: worktree `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2105`, branch WT-spec-lint-baseline-3, HEAD `acfd92d3d78d1e63dc71ed1f5572810175c8c424`, clean.
- Judging build (§2.2): `go build -o <scratch>/moai-t2105 ./cmd/moai` from this clean tree at that HEAD, invoked by path. `version` prints `v3.2.0 none built unknown`, because a plain build carries no ldflags. The build commit is therefore the tree HEAD `acfd92d3d7` by construction, not by a stamped field. The installed `moai` binary was used only for `verify audit-plan` and `todo list`, not for any lint measurement.

## Gaps

- The whole-corpus control (clone without the skip) was stopped by its PID after about 4 minutes to free CPU for the HEAD corpus run, so there is no corpus-level before/after diff. Claim 5 rests on the per-document skip path in source (E2: `applylintSkip` and the opt-out read only `doc`'s own frontmatter), the target-level control (E5), and the unchanged three-SPEC result (E7).
- The progress.md record of the cause that D1 asked for, which t2107 cites, is not in this diff and was not read (F3).
- `rm -rf` of the scratch clone directory `<scratch>/ctl-clone` was denied by the permission layer, so the clone remains in the session scratchpad. The scratch binary and the lint output files were deleted (`rm -f`, rc 0).
- Coverage percentage not measured: no Go code changed. Only the rule's existing tests were run.
- The Codex gate result is not in this file. Its receipt is cited in the auditor's final message, after this file was committed.

## Residual-risk

- While the skip stands, a genuinely mis-owned future transition on this SPEC would not be reported (D2; removal tracked by t2107).
- Scratch outputs are deleted. The deciding lines are copied above, and nothing in scratch is cited.
- The whole-corpus run is slow (about 33 minutes on this tree, before t2087). Anyone re-running it should expect that, or use the t2087 build.
