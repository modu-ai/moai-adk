auditor-model: claude-sonnet-5-5

verdict: FAIL
note: R2-1 is fixed and proven; the required codex backend fails on a new P2 (GIT_LITERAL_PATHSPECS silently disables the new untracked-module-file query) that this audit reproduced. Narrow, one-line-class fix; not a regression against the pre-card full scope.
audited_sha: 788e1d894ac3932159b2d0ebaf59f747a616b70a
required_backend: codex fail (rcpt-004e21f1d6a55ad84226f3ff)
overall_score: 0.88
blocking_count: 1

## Evaluation Report
SPEC: SPEC-PRECOMMIT-AFFECTED-SCOPE-001 (card t1673), sync audit round 3, scope = `git diff d9f5b3e6e..788e1d894`
Judging builds (section 2.2): CLI probes used `moai` built from this tree's HEAD 788e1d894 (`bin/moai-r3`) and from the round-2 tree d9f5b3e6e (`bin/moai-r2tree`). MCP server / installed `moai` = build ee981a932 (strict ancestor of HEAD; tool reported build_lag) ran `audit_multi` and `moai verify audit-plan`.

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 82/100 | FAIL | R2-1 fixed (E1); one reproduced env-dependent bypass (R3-1); required backend fail. |
| Security (25%) | 90/100 | PASS | Delta adds one read-only fixed-argv git query; no config write. |
| Craft (20%) | 91/100 | PASS | vet/gofmt/builds clean, -race ok, new tests catch reverting mutants (E3). |
| Consistency (15%) | 91/100 | PASS | ko/en/ja/zh and CHANGELOG state the same extension; accurate except R3-1 corner. |

Overall (weighted harmonic): 0.88. Verdict FAIL forced by the required backend.

### Findings
- R3-1 [P2] [blocking by required-backend rule] internal/hook/quality/affected_scope_queries.go:136 - `git --literal-pathspecs commit` exports `GIT_LITERAL_PATHSPECS=1` to the hook; the new `ls-files` query's `:(top,glob)**/go.mod` patterns are then matched literally, the query succeeds with no output, `UnstagedGo` stays empty, and an untracked `b/go.mod` again under-selects. Reproduced here: with the HEAD build, `git --literal-pathspecs commit` with staged `a` and untracked `b/go.mod` ran `test ./a` (ordinary `git commit` ran `test ./...`). Same class as the pre-card behaviour of `go test ./...` in that tree (b excluded either way), and only reachable with that non-default git option or `GIT_LITERAL_PATHSPECS` set. Fix: drop `GIT_LITERAL_PATHSPECS` from the query environment (or list with `ls-files --others --exclude-standard -z` and filter by base name in Go, no pathspec), plus a regression test through `git --literal-pathspecs commit`. Or record as an accepted limit; the ruling is the leader's.
- No other demonstrable defect in the delta.

### Requested checks
- Negative controls and judgement on untracked plain .go / ignored files (real temp repos, HEAD build): untracked .txt, plain `.go` (valid package clause), and `go.sum` stay narrow (`test ./a ./b`); an untracked `.go` with a conflicting package clause makes `go list` report an errored package and the gate falls back to `./...` (so the "conflicting clause is already caught" comment holds, observed). A plain untracked `.go` can only add a package or edge, never remove one, so leaving it narrow is sound. Ignored nested go.mod: `--exclude-standard` skips it, the gate runs `test ./a`, and `go list ./...` in that tree also prints only `example.com/m/a`, so a plain full-module run has the same blind spot; accepted-limit class, not a delta defect.
- Mutants (scratch HEAD extract): dropping the untracked loop fails TestAffectedScope_UntrackedModuleFileFallsBackToFullModule, ...ThroughHook and ...WorkingTreeQueriesFromNestedModule; dropping `-c diff.relative=false` from the unstaged query fails WorkingTreeQueriesFromNestedModule (the round-2 survivor is now covered); dropping `top` magic fails two tests; dropping `-c` from the untracked query alone survives (`ok`), harmless because untracked names are only listed and base-name matched after a top-anchored pattern.

### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

#### Claim
C1 R2-1 fixed. C2 negative controls stay narrow. C3 new tests non-vacuous. C4 light checks, probe016 GREEN, probe009 RED witness. C5 codex fails on R3-1, real.

#### Evidence
E1 (C1, C2) `sh /private/tmp/t1673-audit/repro-r3.sh <binary>` (real temp repos, hook running `MOAI_PRECOMMIT=1 moai gate`, recording `go` wrapper). HEAD build:
```
r21_untracked_b_gomod: commit exit=0, go test calls=[test ./...;] (want: test ./...), notice lines naming untracked=1
neg_untracked_txt: ... go test calls=[test ./a ./b;]
neg_untracked_plain_go: ... go test calls=[test ./a ./b;]
neg_untracked_conflicting_go: ... go test calls=[test ./...;]
neg_untracked_gosum: ... go test calls=[test ./a ./b;]
untracked_gowork: ... [test ./...;] notice=1
untracked_gowork_sum: ... [test ./...;] notice=1
nested_untracked_b_gomod: ... [test ./...;] notice=1
nested_root_gowork: ... [test ./...;] notice=1
ignored_b_gomod: ... [test ./a;]
```
Round-2 build, same script: `r21_untracked_b_gomod ... [test ./a;]`, `untracked_gowork ... [test ./a ./b;]`, `untracked_gowork_sum ... [test ./a ./b;]`, `nested_untracked_b_gomod ... [test ./a;]`, `nested_root_gowork ... [test ./a ./b;]`.
E3 (C3) pristine HEAD extract: all subtests of TestAffectedScope_UntrackedModuleFileFallsBackToFullModule (3 widen, 3 narrow), ...ThroughHook (plain under commit -a temp index; linked worktree), TestAffectedScope_WorkingTreeQueriesFromNestedModule PASS. Mutant results as listed above.
E4 (C4) `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`, `go vet ./internal/hook/quality/`, `gofmt -l` on the four delta Go files: no output. `go test -race -count=1 ./internal/hook/quality/`: `ok ... 71.135s`. probe016.sh with HEAD build: four GREEN lines, `exit=0`. probe009.sh: rows (ii)(iii)(iv) `[DIFFERENT]`, `exit=1`.
E5 (C5) audit_multi: `overall_verdict: fail`; codex required, fail, one P2 (`affected_scope_queries.go:136`); `plan_source: config`; `audit_receipt: rcpt-004e21f1d6a55ad84226f3ff`. Repro: in `/private/tmp/t1673-audit/r3/moai-r3/r21_untracked_b_gomod/repo`, staged `a/a.go`, untracked `b/go.mod`, `git --literal-pathspecs commit` -> `exit=0`, go log `test ./a`.
E6 convergence: digest `{"overall_verdict":"fail","gate_unmet":"","plan_source":"config","per_backend_verdicts":[{"backend":"codex","gate":"required","verdict":"fail"}]}` written to `.moai/state/audit-plan-result.json` (untracked); `moai verify audit-plan --project-root <tree> --result-file <path>` printed `"enforced_required": ["codex"]` and `"convergence_check": {"ok": true, "unmet": []}` (the required backend answered; verdict remains fail).

#### Baseline-attribution
All observations this run against HEAD 788e1d894ac3932159b2d0ebaf59f747a616b70a (clean) or scratch extracts under /private/tmp/t1673-audit.

#### Gaps
- Slot not acquired (leader-ordered turn); not run: full internal/template, full internal/cli, -race beyond internal/hook/quality, round-1 internal/cli/config selections (delta touches quality package + docs only).
- Docs pages compared by diff, not hugo-built; I did not diff the four page diffs line by line beyond confirming they were part of the delta.
- R3-1 fix not attempted.

#### Residual-risk
R3-1 (env-dependent git option); ignored nested go.mod blind spot shared with plain `go test ./...`; F2 over-widening conservative; Q1=B unchanged.

### Iteration history
R1 FAIL 0.74 (rcpt-8b5dc83de1625ec3b29b2e61); R2 FAIL 0.84 (rcpt-61e4f2bdaaa1e03d1d4a8342); R3 (this) FAIL 0.88 (rcpt-004e21f1d6a55ad84226f3ff).

AUDIT-VERDICT: FAIL spec=SPEC-PRECOMMIT-AFFECTED-SCOPE-001 receipts=rcpt-004e21f1d6a55ad84226f3ff
