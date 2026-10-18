auditor-model: claude-sonnet-5-5

# Sync audit, card t1833, iteration 1

verdict: FAIL
audited_sha: 8f1873c58d32e8f09fd9655b0b835399b2ff8a48
overall_score: 0.62
must_pass_failed: 1
blocking_count: 4
required_backend: codex
convergence_overall: fail

Dimension scores (flat weighted mode, no `evaluator_mode: hierarchical`):

| Dimension | Score | Verdict |
|---|---|---|
| Functionality (40%) | 60/100 | FAIL (must-pass) |
| Security | 70/100 | PASS-below-threshold risk, see F2/F3 |
| Craft (20%) | 70/100 | PASS |
| Consistency (15%) | 80/100 | PASS |

Must-pass firewall: Functionality fails because the verb can name a false culprit (F1) and can leave a tree mid-bisect while reporting success (F3), and because the required codex gate returned `fail`. The Overall verdict is FAIL regardless of the other scores.

Audit plan read: `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833` returned `config_status: ok`, `model: codex`, `cross_model_required: true`, gates claude=off, codex=required, glm=off, `enforced_required: [codex]`.

## Defect list

- F1 [high] [blocking] internal/cli/integration_bisect.go:126-129 - the culprit is read with the first regex match over the mixed output of git and the judged test command (`io.MultiWriter(&log, out)`). A test command that prints `<40-hex> is the first bad commit` makes the verb report a card git did not converge on. Required fix: after a successful `git bisect run`, read the culprit from git's own state (`git rev-parse refs/bisect/bad` before reset, or parse only git's last such line) and never from test output. Status: code-confirmed; codex reproduced (`card: t1001` reported when git converged on t1003); my own repro attempt was malformed and is NOT counted as evidence (see Gaps).
- F2 [high] [blocking] internal/cli/integration_bisect.go:67-71 vs :77 - `bisectGit` runs git under the caller's environment, while the pre-checks use `runScrubbedGit`. With `GIT_DIR`/`GIT_WORK_TREE` set, the primary-checkout and dirty-tree checks inspect one tree and `bisect start/run/reset` act on another, which defeats the refusal of the primary checkout. Required fix: run start/run/reset with the repository-scoped GIT_* variables removed (the same scrub the pre-checks use) while leaving the rest of the environment to the judged command. Status: code-confirmed (comment at line 65 says the inheritance is deliberate); codex reproduced.
- F3 [high] [blocking] internal/cli/integration_bisect.go:120-124 - a failed `git bisect reset` prints a WARNING and the function returns nil, so the verb exits 0 with the tree still mid-bisect. This contradicts the CHANGELOG and runbook claim "always resets". Required fix: return an error that includes the reset failure (and verify HEAD and the end of the bisect) instead of reporting success. Status: code-confirmed; codex reproduced (`error=<nil>; bisect_still_active=true`).
- F4 [medium] [blocking] internal/cli/integration_bisect.go:116-119 - the `defer` that resets is registered only after `git bisect start` succeeds; a start that fails partway (for example a failing post-checkout hook) returns without any restore. Required fix: record the original branch/SHA before start and restore it on a failed start. Status: code-confirmed; codex reproduced (`HEAD_restored=false`).
- F5 [medium] [blocking] scripts/release.sh:424 - `git tag -a "$VERSION" -F "$TMP_NOTES"` names no commit, so the tag lands on whatever HEAD is at that instant. The re-check at line 421 narrows but does not close the window. Required fix: `git tag -a "$VERSION" -F "$TMP_NOTES" "$LOCAL_SHA"` so the tag target is the verified commit by construction. Status: code-confirmed (read lines 421-424); codex reproduced the tag landing on a changed HEAD. Practical exposure is small (no prompt between the two lines) but the fix is one token and the stated goal is "tag only a commit whose CI passed".
- F6 [medium] [optional] scripts/release.sh:241-251 - the gate passes when at least one check run exists and all are green; it never checks that the expected CI workflow ran. A single green check run from an unrelated app satisfies it. I reproduced the jq arithmetic for a one-element neutral page (`1 0 0` = pass). The runbook (limit 8) discloses strictness but not this hole. Required fix (or accept as debt): require a named workflow's check run to be present and completed. Codex rated this P1 and reproduced a real local tag; I rate it optional because the card text asks for "CI passed", the repo CI workflow is not identified in the SPEC-less card, and the residual is disclosed only partially.
- F7 [low] [optional] .moai/docs/batch-failure-bisect-runbook.md section 2.2 - the runbook states exit codes 126-127 are "bad". In a scratch repository, `git bisect run sh -c nonexistent-tool` aborted with `error: bogus exit code 127 for good revision`. A missing test tool therefore aborts the search (the verb then reports "did not converge"), it does not mark commits bad. Also the verb's error text says "exited >127" while this case is exactly 127. Required fix: correct the exit-code sentence and the error text.
- F8 [low] [optional] runbook wording - `동사` for the CLI verb and `끝점` are translation-style calques of "verb" and "endpoint" (Korean CLI documentation says `하위 명령` / `양 끝 ref`). The rest of the prose reads as written Korean. Required fix: replace `동사` with `하위 명령` (or `명령`).
- F9 [low] [optional] progress.md says "3 such chore commits" among one-parent entries; `chore(config)`-prefixed subjects in the 100-window count 2 (the third is `chore: commit the primary checkout's pending changes`). Wording only.

## 1. Claim: landings are per-card --no-ff merges carrying the card id

Evidence (commands run in this tree at HEAD 8f1873c58):

```
git log --first-parent --format='%P|%s' main -n 40   -> parents=2: 40; two-parent with card id: 40; without: 0
git log --first-parent --format='%P|%s' main -n 100  -> parents=2: 68, parents=1: 32; two-parent with card id: 64; without: 4
the 4 two-parent, no-card subjects:
  Merge WT-codex-moai-home-add-dir into main (codex --add-dir moai home, integration merge)
  Merge pull request #1816 from modu-ai/WT-always-loaded-char-budget
  Merge pull request #1812 from modu-ai/WT-t1538-factory-recovery
  Merge pull request #1772 from modu-ai/WT-user-asset-copy
oldest of the 100: 2026-10-07 feat(cli): unify todo --auto ... (#1787)
grep -n 'no-ff\|integration merge' internal/factory/integration_merge_step.go -> comment lines 2, 10 ("merges --no-ff of the PINNED SHA") and error strings
git cat-file -t 6dde1c1c1 -> commit; ancestor(M^1, M^2) exit 0; tree(M) == tree(M^2); git log --first-parent M^1..M^2 | wc -l -> 63
```

Result: the shape table in progress.md section 1 and runbook section 1 reproduces exactly (40/0/0 and 64/4/32; 63 units). The batch-shape claim reproduces.

Baseline-attribution: local `main` as read by this tree, run in this audit; judging build is the installed `moai` v3.2.0-rc.36-t1822c (commit ee981a932, an ancestor of tree HEAD 8f1873c58 per the audit_multi build_lag note). The measurements above used plain `git`, not that build.

Gaps: my `grep -n 'Merge %s into %s (card %s, integration merge)'` and `'git merge --no-ff'` patterns did not match as literal producer lines in my truncated `head` output (the output showed comment mentions at lines 2 and 10). I did not independently re-read lines 399 and 609 that progress.md cites. The mixed-claim "producer exists" is therefore supported by the comment lines only, plus the 40/40 observed subjects. I treat the producer claim as supported, not line-verified.

Residual-risk: direct and batch-assembly commits give `card: none`, as both documents say.

## 2. Claim: tests, syntax, vet, format are clean

Evidence (env -u MOAI_FACTORY_WORKER -u CLAUDE_PROJECT_DIR TMPDIR=GOTMPDIR=/private/tmp/lane3-audit; no -race):

```
go test ./internal/cli -run '^TestBisect' -count=1 -v        exit 0
  --- PASS: TestBisectNamesTheCulpritCard / RefusesThePrimaryCheckout / RefusesADirtyTree /
      RefusesWhenOneIsInProgress / DryRunPrintsTheGitCommandsAndMovesNothing /
      RequiresGoodAndAResolvableRef / ReportsAnAbortedSearchAndStillResets / CmdRejectsAnEmptyCommand
  ok  github.com/modu-ai/moai-adk/internal/cli  8.549s
go test ./internal/template -run '^(TestReleaseScriptCIGate|TestReleaseScriptCIGateWithoutGh|TestReleaseScriptCharacterization|TestReleaseScriptMatrixGate|TestReleaseScriptHeadAdmission|TestReleaseScriptRcChangelogSkip|TestGitHubFlowSweepGuard)$' -count=1 -v   exit 0
  all subtests PASS (CIGate 7 rows, CIGateWithoutGh 2 rows, Characterization 9, HeadAdmission 5, RcChangelogSkip 4, MatrixGate 14, SweepGuard 2)
  ok  github.com/modu-ai/moai-adk/internal/template  36.782s
bash -n scripts/release.sh          exit 0
gofmt -l <5 Go files>               exit 0, no output
go vet ./internal/cli ./internal/template   exit 0, no output
```

The first run of both test commands returned exit 127 because the compound command lost the Go toolchain from PATH in that call; I re-ran each separately and the exit-0 results above are from the re-runs. Passing tests do not exercise F1-F5: no test sets GIT_DIR, hooks, a spoofing test command, a failing reset, or a moved HEAD between the check and the tag.

## 3. Claim: the verb and the release gate refuse what they say they refuse

Reproduced by running (isolated stubs/repos, never the real script on the real repo):

- Release gate rows via the repo's stubbed-gh fixture (TestReleaseScriptCIGate): green passes; pending, failed, failed-beats-pending, absent, unreadable all exit 1 with the named condition and stop before the dry-run preview; `--skip-ci-check` on a failed run exits 0 and prints `CI gate OVERRIDDEN (--skip-ci-check)` and `CI gate:     OVERRIDDEN (--skip-ci-check)`. No `gh` on PATH: exit 1 with `CI gate: the gh CLI is not available`; with the override, exit 0.
- jq arithmetic checked against hand-built pages: `neutral` only -> `1 0 0` (pass); `cancelled`, `timed_out`, `action_required`, `stale`, completed-with-null-conclusion -> `1 0 1` (fail); `queued`, `waiting` -> `1 1 0` (pending); empty -> `0 0 0` (absent). Two pages summed by the script's awk: `2 1 0`. Garbage or empty gh output sums to `0 0 0`, which the script treats as "no check runs" and refuses (fail-closed).
- Bisect behaviours in a scratch repo with real git: missing test tool aborts (F7); an always-failing command names the first first-parent merge (card t101); after `git bisect reset` the detached HEAD equals the original; `--bad` equal to `--good` makes `git bisect start` exit 1 with "We are not bisecting" (verb would return the start-failed error); `--bad` an ancestor of `--good` prints a "mistook good and bad revs" warning and git bisect starts no search.
- By reading: the verb requires `--good`, defaults `--bad` to HEAD, refuses the primary checkout through `identifyPrimaryCheckout`+`factorySameTree`, refuses a dirty tree and an existing `BISECT_LOG`; a detached HEAD (the test shape) is accepted and restored.

Baseline-attribution: tree 8f1873c58; no claim of a live `gh` shape.

Gaps: the stub does not exercise the real `gh api --paginate --jq` JSON/page shape (the stub pipes one canned file through jq). Per the audit instruction this is a Gap, not a finding. I did not run the verb against the real repository history. The HEAD-moved window (F5) and the F1-F4 conditions were not independently executed by me; they are code-confirmed and, for F1-F5, reproduced by the codex backend.

## 4. Claim: the runbook agrees with code and is local-only

Read against the code: flags (`--good`, `--bad` default HEAD, `--dry-run`), the three git commands, the refusals, `card: none` text (code prints `none in the subject (a direct commit, not a card landing)`; the runbook's `card: none in the subject` matches), and the `did not converge` error all agree. Exceptions: F7 (126-127 sentence) and F8 (calques). Section 4 limit 9 honestly discloses the unexercised real GitHub shape.

Local-only claim: `.moai/docs/batch-failure-bisect-runbook.md` is a tracked file (it appears in `git diff --stat`), absent from the template tree: `internal/template/templates/.moai/docs/` holds 6 files (agent-lint, audit-artifact-convention, generic-patterns-guide, hook-stdin-fail-closed, session-handoff-appendix, todo-queue-storage). The only template test that pins a `.moai/docs` file by path in my grep is spec_ceiling_mirror_test.go (audit-artifact-convention.md). "Not mirrored" is accurate; "local" means local to the repo, not untracked. I did not run the whole `internal/template` package, so a test that enumerates `.moai/docs` and rejects an unmirrored file is not excluded (Gap).

## 5. CHANGELOG

Two entries inserted at the top of `### Added` and `### Changed`, in the form `**title (card tNNNN, class B, no SPEC).** body... 🗿 MoAI`, matching neighbouring entries. Cited paths exist (internal/cli/integration_bisect.go, internal/cli/integration.go, the runbook, progress.md). The claim "always resets on the way out" in the Added entry is contradicted by F3/F4 and must change with the fix.

## 6. Scope hygiene

`git diff --stat fb749d372..HEAD`: 9 files, all in the audit list (runbook, progress.md, CHANGELOG, integration.go, integration_bisect.go, integration_bisect_test.go, release_ci_gate_test.go, release_fixture_test.go, scripts/release.sh). `git status --short | wc -l` -> 0 (clean) at the end of the audit; HEAD stayed 8f1873c58d32e8f09fd9655b0b835399b2ff8a48 throughout. My scratch files are in /private/tmp/lane3-audit. No foreign commit observed.

## 7. Audit procedure (required codex gate)

`mcp__moai__audit_multi` was called with `project_root` set to this tree, `target: baseBranch`, and a focus on tag-placement and bisect-state failures, no `gates` argument (the plan applies).

audit_multi result members:
- overall_verdict: fail
- gate_unmet: not present in the returned object (`residual_risk_note`: "required-backend FAIL: codex")
- plan_source: config
- per_backend_verdicts: backend codex, gate required, verdict fail (6 findings: 3 P1, 3 P2; summarised as F1-F6 above)
- participant_count: 1; disagreement_flag: null
- build_commit ee981a932; build_lag: the installed binary is an ancestor of tree HEAD 8f1873c58 (fixes after ee981a932 are not in it)
- audit_receipt: rcpt-12da61c41fa93963600ea2a4

This verdict is not final until the orchestrator's `moai verify audit-plan --result-file` check passes; the orchestrator writes the digest. The codex findings were not filtered: F1-F5 are code-confirmed by me, F6 is judged optional.

Baseline-attribution (judging build): tool measurements of `moai verify audit-plan` and `audit_multi` were produced by the installed moai build ee981a932, which the server reports as a strict ancestor of the tree HEAD 8f1873c58 (a lag of this tree's commits; those commits only add the new verb, so audit_plan and audit_multi do not depend on them). The Go tests were built from the tree itself by `go test`.

## Gaps (consolidated)

- Real `gh` JSON shape and `--paginate` page boundaries were not observed (stub only); live GitHub forbidden.
- My attempt to reproduce F1 by git scratch repo was malformed (the fake SHA used was a genuine culprit-adjacent merge, so it proved nothing); F1 rests on code reading plus codex's reproduction.
- F5 and F2/F4 were not independently executed; they rest on reading plus the codex reproductions.
- Producer lines 399/609 of integration_merge_step.go not re-read by me; supported via comments and 40/40 subjects.
- Whole-package `go test` for `internal/cli` and `internal/template` not run (load rule); `golangci-lint` not run.
- Two initial test commands returned exit 127 (toolchain missing from that call's PATH); I re-ran them and report only the re-run results.
- No tool was refused by the guard in this audit.

## Residual-risk

Even with F1-F5 fixed: the strict all-check-runs gate still passes on a single unrelated green check run (F6) and a hand-made tag skips release.sh; the verb still trusts the caller to verify both endpoints (runbook limit 1); the judged command runs with the caller's full environment by design.

## Iteration history

- Iteration 1: FAIL on the required codex gate and four code-confirmed blocking defects (F1-F4) plus the tag-target pin (F5). A confirming re-audit is scoped to F1-F5 (and the CHANGELOG sentence), not a from-scratch audit.

## Operational Notes (unverified)

- assumption: F6 may be accepted as a PASS-WITH-DEBT item if the operator confirms the release CI workflow name; measure with `gh api repos/modu-ai/moai-adk/commits/<release-sha>/check-runs --jq '.check_runs[].name'` on a real release commit.
- inferred: F1 and F2 share a root cause (reading and acting through unscrubbed git); a fix that routes start/run/reset through one scrubbed helper and reads `refs/bisect/bad` before reset closes both.
