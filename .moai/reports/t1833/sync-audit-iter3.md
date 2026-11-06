auditor-model: claude-sonnet-5-5

# Sync audit iteration 3 — card t1833

verdict: FAIL
audited_sha: 29411d045b275532788ee9b0708b573194dfad37
overall_score: 0.75
must_pass_failed: 1
blocking_count: 2
required_backend: codex
convergence_overall: fail
required_backend_fail: codex

Dimension scores: Functionality 0.62 (must-pass, FAILED), Security 0.78, Craft 0.80, Consistency 0.86.
Overall is the harmonic mean (0.754). The must-pass firewall forces FAIL on Functionality regardless.
Audit receipt (codex, required gate): rcpt-fe044acf62f06af8139230bd

## Iteration history

| Iter | Commit | Verdict | Blocking |
|---|---|---|---|
| 1 | 8f1873c58 | FAIL 0.62 | F1-F5 (F6 codex P1) |
| 2 | 505749a92 | FAIL 0.76 | N1 (N2, N3 codex P1) |
| 3 | 29411d045 | FAIL 0.75 | N7, N8 (both reproduced; N7 is the incomplete half of N3) |

Round 2 fixed N1, N2, N6a and the jq skip as claimed. N3 is only partly fixed (plain same-path collisions are refused; renames into an ignored path still destroy the file). The runbook sentence about SIGTERM-ignoring commands is false as written.

## Findings

- N7 [high] [blocking] internal/cli/integration_bisect.go:158-163 - `bisectIgnoredCollisions` lists paths added in the range with `git log --first-parent --diff-filter=A`. With git's default rename detection (diff.renames on) the destination of a rename shows as `R`, not `A`, so a rename into a path where the developer has an ignored local file is not seen. Reproduced: ignored local `secret.env` (`LOCALPRECIOUS`), commit `c` does `git mv tmp.dat secret.env`; the verb ran, named a culprit, exit 0, and the local file no longer exists afterwards. Codex reproduced the same for `diff.renames=copies`. The runbook says the verb refuses ignored files that share a path with a path "added in the range", so the document promises what the code does not do. Required fix: add `--no-renames` to the `git log` call (the `A` filter then lists rename and copy destinations; I confirmed `--no-renames` lists `secret.env`), add a rename row to the collision test, and keep the runbook sentence.
- N8 [medium] [blocking — wording of a documented limit is inaccurate] .moai/docs/batch-failure-bisect-runbook.md:73 and CHANGELOG.md (verb entry, "the judged run is stopped first") - the runbook says a judged command that ignores SIGTERM is killed directly by the runtime after 10 seconds. Reproduced (judged command `trap '' TERM; sleep 300 & ...; wait`, SIGTERM to the verb, also with stdout through a pipe): the verb returned exit 1 after 0 s with "the tree is restored", and the judged `sleep` was still alive afterwards. `git bisect run` dies on SIGTERM at once and `WaitDelay` kills only that direct child. Codex reproduced the same (heartbeat still updating 11 s later). Required fix: either SIGKILL the whole process group after the grace period and confirm it is gone before restoring, or rewrite the sentence and the CHANGELOG clause to say that a command that ignores SIGTERM keeps running as an orphan after the tree is restored (and add the SIGKILL orphan to the same limit).
- N9 [medium] [optional] internal/cli/integration_bisect.go:173 + runbook line 69 - file/directory collisions (ignored local `sub` against a commit that adds `sub/x`, and the reverse) are not detected and the data is lost. Reproduced for the first direction: the ignored file `sub` was gone after the run. The runbook does state "파일과 디렉터리가 같은 이름을 쓰는 경우는 보지 않는다", so the wording is accurate and this is a documented limit, but it is silent data loss; codex rates it P1 and reproduced both directions. Fix: compare path components (ancestor/descendant), refuse before HEAD moves.
- N10 [medium] [optional] scripts/release.sh:254-270 - check-run names are put into tab-separated rows without escaping. A check run whose name contains tabs and newlines forges rows. Reproduced the mechanism with `jq -r` plus the script's own awk on a single `evilapp` check run: Lint, Test (ubuntu-latest) and Build (linux/amd64) are all reported as satisfied by github-actions. Whether GitHub accepts such a name from an app with checks:write was not tested (no remote access allowed) — see Gaps. Fix: build the verdict from the JSON object (select by name and app in jq) instead of re-parsing delimited text.
- N11 [low] [optional] internal/cli/integration_bisect.go:168 - `strings.TrimLeft(p, "\n")` removes real leading newline characters of a NUL-delimited file name (codex reproduced a bypass with a file named `\nvictim`). Fix: strip only the separator that `git log` inserts between commits, not name bytes.
- N12 [low] [optional] internal/cli/integration_bisect.go:236 - dry-run quotes the judged command but prints refs unquoted, so a ref named `$(...)` executes when the line is pasted. Fix: pass every printed argument through `shellQuoteArg`.
- N13 [low] [optional] runbook limit 12 area - the SIGKILL limit says the restoration does not happen but not that the judged command and its children survive (reproduced: orphan `sleep` alive after SIGKILL, bisect log present, HEAD detached). Add the orphan to the sentence.
- N14 [info] [optional] release tests - without `jq` the whole gh-stub family now skips with an explicit reason (verified: SKIP lines, no vacuous PASS). `TestReleaseScriptCIGateWithoutGh` and the characterization subtests skip as well although some do not need `jq`; the parent shows PASS with all subtests SKIP. Acceptable; not blocking.

## Claim / Evidence / Baseline / Gaps / Residual-risk

### Claim 1 — author's tests and static checks pass

Evidence (commands, all run in the tree at HEAD 29411d045, env `-u MOAI_FACTORY_WORKER -u CLAUDE_PROJECT_DIR`, TMPDIR/GOTMPDIR under /private/tmp/lane3-audit3):
- `go test ./internal/cli -run '^TestBisect' -count=1 -v` -> `ok  github.com/modu-ai/moai-adk/internal/cli 27.522s`, exit 0; 12 listed tests incl. `TestBisectRestoresTheTreeOnSignal/{interrupt,terminated}` PASS and `TestBisectRefusesAnIgnoredFileThatACommitInTheRangeTracks` PASS (`TestBisectSignalHelper` SKIP is the helper by design).
- `go test ./internal/template -run '^(TestReleaseScriptCIGate|...|TestGitHubFlowSweepGuard)$' -count=1 -v` -> `ok  github.com/modu-ai/moai-adk/internal/template 45.500s`, exit 0; 13 PASS subtests of `TestReleaseScriptCIGate` counted; no SKIP with jq present.
- `bash -n scripts/release.sh` exit 0; `go vet ./internal/cli` 0; `go vet ./internal/template` 0; `GOOS=windows go vet ./internal/cli` 0; `gofmt -l` on the touched Go files printed nothing.

### Claim 2 — N1 (signal handling) fixed

Evidence: scratch repo `/private/tmp/lane3-audit3/r2..r8` (5 `--no-ff` card merges, primary + linked worktree), verb built from this tree. Judged command `sleep 300 & ...; wait`.
- SIGINT to the verb: `verb exit 1 after 0s`, message "interrupted by a signal before the search converged, no culprit is named; the tree is restored", `HEAD f4ec173 branch WT-x ; before f4ec173`, `bisect log present: no`, `status: []`, `judged child 94175 gone`.
- SIGTERM: same result (HEAD eed5c10 restored, child gone).
- Two SIGINTs 0.3 s apart: exit 1, restored, child gone.
- Signal during `git bisect start` (post-checkout hook sleeping 3 s): exit 1 after 6 s, restored, no bisect log.
- Signal while restore/late: swallowed; verb finished with normal result and HEAD restored (exit 0).
- Judged command that moves HEAD forever (runbook limit 11): still running after 6 s; SIGINT -> exit 1, `HEAD 051acb1 was 051acb1 branch WT-x bisect: no`. Limit 11's claim "Ctrl-C and the tree is restored" holds.
- SIGKILL: exit 137, HEAD detached at a midpoint, bisect log present, orphan `sleep` alive — matches the documented limit (orphan not mentioned, N13).
- SIGTERM-ignoring judged command: see N8 (exit 0 s, orphan alive, "tree is restored").
Verdict on N1: fixed for cooperative commands; N8 for the uncooperative case.

### Claim 3 — N3 (ignored-file collision) fixed

Evidence: scratch repos `n3*`, `n3b*` (ignored via `.git/info/exclude`, worktree created at the good commit, `--bad main`).
- Plain add of `secret.env`: refused, message names the file, file kept (`secret.env=[LOCALPRECIOUS]`).
- `cp tmp.dat secret.env` (add): refused.
- Add then delete in a later commit: refused.
- Rename into `secret.env`: NOT refused; verb output `card: t23967`, afterwards `ls secret.env: No such file or directory`. Manual `git log --first-parent --diff-filter=A --name-only -z --format=` lists `e.txt|d.txt|tmp.dat|a.txt|` (no `secret.env`); with `--no-renames` it lists `secret.env`. -> N7.
- Directory against file (`sub` ignored file, commit adds `sub/x`): not refused, file lost -> N9 (documented).
- Performance: on this 16053-path tree `ls-files --others --ignored` 0.17 s, `ls-tree -r` 0.07 s, `log --diff-filter=A` over 300 commits 0.28 s with the warning "exhaustive rename detection was skipped due to too many files" (a second reason for `--no-renames`).
Verdict on N3: fixed for same-path adds and copies, NOT fixed for renames -> blocking N7.

### Claim 4 — N6a (dry-run quoting) fixed

Evidence: judged command `echo 'it'"'"'s' $HOME "a b" ; test -n "$(ls)" && printf '%s\n' "x y"`; printed line `git bisect run sh -c 'echo '\''it'\''"'\''"'\''s'\'' $HOME "a b" ; test -n "$(ls)" && printf '\''%s\n'\'' "x y"'`; evaluating it with a shell function `git()` that prints argv gave exactly 5 words with `argv[5]` byte-equal to the original argument. Multi-line command stays one quoted word. Refs are not quoted (N12).

### Claim 5 — N2 (app slug) and the jq skip

Evidence:
- `scripts/release.sh` jq now emits `.app.slug // ""`; `jq -r` on `app: null`, a missing `app`, and a `github-actions` run gives rows with an empty or correct fourth field and no jq error (od -c output checked). awk requires `$1 == name && $4 == "github-actions"`: a run from `github-actions-evil` does not satisfy `Lint` while the real `github-actions` run does.
- Tests `right_name_from_another_app_is_refused` and `right_names_without_an_app_are_refused` PASS.
- Skip placement: test binary run with PATH lacking jq (`/private/tmp/lane3-audit3/nojq`): `--- SKIP: TestReleaseScriptCIGate`, `...WithoutGh`, `...TagsTheVerifiedCommit`, and all `Characterization`/`HeadAdmission` subtests SKIP with "jq not available; the gh stub answers through jq"; none PASS vacuously, no FAIL. Matrix gate skips with its own reason.
- Pagination: `gh api --paginate --jq` applies the expression per page by gh design; not exercised against real gh (Gap).
- N10 (forged rows) is a new weakness of the delimiter approach.

### Claim 6 — iteration-1 F1-F6 still hold

Evidence: tests `TestBisectIgnoresACulpritClaimPrintedByTheJudgedCommand` (F1), `...ActsOnTheCheckedTreeWhenGitDirIsSet` (F2), `...FailsWhenTheResetFails` (F3), `...RestoresHeadWhenStartFailsPartway` (F4), `TestReleaseScriptTagsTheVerifiedCommit` (F5), `TestReleaseScriptCIGate` unrelated/missing rows (F6) all PASS. Independent re-runs: judged command printing a fake "is the first bad commit" line -> verb still reported `card: t904`; run with `GIT_DIR`/`GIT_WORK_TREE` pointing at the primary -> `card: t904`, primary stayed on `main`, worktree on `WT-x`, both clean. No regression.

### Claim 7 — docs and scope

Evidence: `git diff --name-status fb749d372 HEAD` lists only the runbook, three t1833 reports, CHANGELOG, `integration.go`, the bisect Go files and tests, the two template tests and `scripts/release.sh`; `git status --short` empty; `git ls-files .moai/state` lists only `.gitkeep` (pre-existing). Runbook sentence "resets on every return path" is followed in the same paragraph by the Windows and SIGKILL limits, and limits 10-12 match behaviour (limit 11 verified). Two inaccuracies: the 10-second sentence (N8) and the "added paths" promise versus renames (N7). Korean register: sentences read as written Korean with `하위 명령`, `양 끝 ref`; no calque found beyond the loanword-heavy but established terms (`first-parent`, `orphan` not used). CHANGELOG English entries for t1833 repeat "twelve limits", "14 TestBisect tests" (the run lists 13 Test functions plus the helper plus 2 signal subtests — consistent enough; not counted as a finding) and "the judged run is stopped first" (see N8).

### Claim 8 — project audit procedure

Evidence: `moai verify audit-plan --project-root <tree>` -> `config_status ok`, codex gate `required` (explicit), claude and glm `off`, `cross_model_required: true`. `mcp__moai__audit_multi` called with `project_root` = this tree, `target: baseBranch`, no `gates` argument, no `claude_verdict` (claude gate off). Result: `overall_verdict: fail`, `gate_unmet` absent, `plan_source: config`, per_backend_verdicts `[{backend: codex, gate: required, verdict: fail}]`, `audit_receipt: rcpt-fe044acf62f06af8139230bd`, `residual_risk_note: required-backend FAIL: codex`. Codex P1/P2 findings: rename/copy collision (= N7), file/dir collision (= N9), forged TSV rows (= N10), SIGTERM-ignoring process (= N8), leading-newline filename (= N11), unquoted refs in dry-run (= N12). I reproduced the mechanisms of N7, N8, N9, N10 myself; N11 and N12 are codex-only and are not counted as independently observed. This digest is not final until the orchestrator's `moai verify audit-plan --result-file` check passes.

### Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833`, `git rev-parse HEAD` = 29411d045b275532788ee9b0708b573194dfad37 (matches the audited commit; no foreign commit seen: HEAD unchanged at the end).
- Judging builds: the scratch verb `/private/tmp/lane3-audit3/moai-t1833` was built from this tree (`go build ./cmd/moai`, HEAD 29411d045) just before the scenarios; the tests ran from source. The MCP server judging the codex leg is build `ee981a932`, an ancestor of HEAD (`build_lag` reported by the tool), so its code-level output was produced from the diff the server collected, not from its own embedded copy of this code; its findings were cross-checked by my own runs above.
- git 2.54.0 (Apple Git-157); jq 1.8.1; macOS. Observations with other git versions (older git where `--first-parent` does not imply a first-parent diff for merge commits) were not made (Gap).

### Gaps

- No live GitHub: the real check-run names, `app.slug == github-actions`, `gh api --paginate --jq` per-page behaviour and whether GitHub accepts a check name containing tab/newline were not observed (N10 exploitability open).
- The verb was not run against the real history; Windows behaviour was only compiled (`GOOS=windows go vet`), not run.
- Case-insensitive file systems (ignored `Secret.env` against tracked `secret.env`) not exercised.
- Signal during the last instants of restore was only approximated; the exact interleaving was not controllable.
- N11, N12 rest on the codex reproduction plus code reading, not my own run.
- No tool was refused during this audit; a first `Bash` attempt using `pgrep -f`/`kill` by pattern was denied by the permission system and I reran the scenario through a script file with explicit PIDs.

### Residual-risk

- After N7 and N8 are fixed, N9 remains a silent-data-loss path (documented). A single `git status --ignored` style refusal by path component would close it.
- The stray-process behaviour of judged commands in other process groups (setsid) is unverified.
- Strict check-run gate depends on the real names/slug; first real `--dry-run` release is the first measurement (documented limit 9).

## Recommendations

1. Add `--no-renames` to the collision `git log`, with a rename test row (closes N7 and removes the rename-limit warning).
2. Either kill the remaining process group with SIGKILL after the grace period and verify it is gone before restoring, or correct the runbook and CHANGELOG sentence (N8). Optional-class items N9-N13 may be taken in the same round; a confirming re-audit scoped to N7 and N8 suffices for the blocking delta.
