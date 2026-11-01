# t2006 verdict (lane-7, Class B, wording-only) — lane's own verdict, evidence below

Policy rules applied: `.claude/rules/moai/core/verification-claim-integrity.md` §2 (attribution), §3.1 (refusal is a Gap); `.claude/rules/moai/development/rule-authoring.md` (statement duty, +357 B < 1000 B); `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § The lane's standard landing.

Tree measured: branch WT-cache-aware-execution, commit 2e7fd7b9d7455510eb52353cbdae39aa7206cb3f, base origin/main 5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4. Judging build: `go test` ran against this tree's source (no installed moai build judged anything); `make build` was run by the delegate in this tree.

Verdict: PASS for the wording change; delivery stops at a draft PR (operator ruling 11-01).

## Claims

| # | Claim | Evidence (command -> observed output) | Baseline |
|---|---|---|---|
| C1 | Tree contains origin/main | `git merge-base --is-ancestor origin/main HEAD` -> exit 0 (run before the edit, HEAD 5c3f0ebf27 after `git merge --ff-only origin/main`) | this run |
| C2 | Live and template mirror identical | `cmp .claude/rules/moai/workflow/cache-aware-execution.md internal/template/templates/.claude/rules/moai/workflow/cache-aware-execution.md` -> `exit=0`; same for `cache-aware-execution-reference.md` -> `exit=0` | this run, post-edit |
| C3 | Always-loaded growth under duty threshold | `wc -c` core file -> `5396` (was 5039 at HEAD per intake `wc -c`) = +357 B | this run |
| C4 | RED -> GREEN on the wording | `git show HEAD:.claude/rules/moai/workflow/cache-aware-execution.md \| grep -c idleCompaction` (run before the commit, when HEAD was 5c3f0ebf27) -> `0`; post-edit `grep -c idleCompaction` core -> `1`, reference -> `2` | this run |
| C5 | No date literals in rule files | `grep -rnE '20[0-9]{2}-[0-9]{2}-[0-9]{2}' <reference file>` -> no output | this run |
| C6 | Template mirror and rule tests pass | `go test ./internal/template/ -run 'Mirror\|Rule\|Embed\|Manifest\|DateProvenance' -count=1 -v` -> exit 0, 59 top-level `--- PASS`, 0 `--- FAIL`, `ok github.com/modu-ai/moai-adk/internal/template 13.365s`, 0 lines `no tests to run`; `--- PASS: TestRuleTemplateMirrorDrift/cache-aware-execution.md` and `.../cache-aware-execution-reference.md`; `--- PASS: TestRuleDateProvenance`. Full log: `.moai/reports/t2006/go-test-template.log` (machine-local, gitignored) | this run |
| C7 | No workflow files changed | `git diff --name-only 5c3f0ebf27..HEAD -- .github/workflows` -> no output | this run |
| C8 | Changed-file set is exactly the 4 rule files | `git status --short` before commit -> 4 `M` lines (2 live, 2 mirror); commit stat `4 files changed, 38 insertions(+), 12 deletions(-)` | this run |
| C9 | Upstream facts re-measured | fetched this run from code.claude.com: changelog 2.1.299 / 2.1.302 / 2.1.314, settings-reference#idlecompaction, prompt-caching#which-ttl-each-request-gets, env-vars row; quoted lines recorded in `.moai/reports/t2006/progress.md` | this run |

## Card review

`.moai/reports/t2006/card-review.md`: codex_review scope=card, `verdict: pass`, 0 findings, base 5c3f0ebf27.

## Gaps (not observed)

- GLM / gateway sessions on the one-hour TTL or idle compaction: NOT measured; the rule text asserts nothing about it and lists it under "Not measured".
- The `DISABLE_AUTO_COMPACT` env-vars row was not read; only the settings-reference "Per-session overrides" statement was.
- Windows / Codex runtime: not applicable, not measured.
- CI and native Codex PR review: not yet run (PR not opened at the time of writing); nothing here claims them.
- `go test` was scoped to `./internal/template/` with a name selector (lane-local rule); the full suite is CI's job.
- Refusal: the worktree guard refused two compound shell batches (a 7-command batch and one with a compound `git` call) during verification; they did not run. Each check they carried was re-issued as plain separate commands (C2-C5, C8) and observed above. The delegate's first test run also failed once for `TestRuleDateProvenance` (literal `2026-11-01` in the reference file); the dates were removed, so the "state the fetch date" instruction in the brief is met only in progress.md, not in the rule files.

## Residual risk

- The wording relies on quoted upstream values that the provider may change; the reference says to re-quote before relying on one.
- A reader on a session where idle compaction is on gets a summary instead of the full conversation; this change documents that, it does not add any mitigation or setting.
