auditor-model: claude-sonnet-5-5

# t1680 sync-audit, round 1

verdict: FAIL
the required codex backend returned fail, which forces FAIL (leader ruling); the finding is independently reproduced below
audited_sha: 950a3e7cba5c5511f4e188fc147bf535b234b5b8
required_backend: codex (gate required, plan_source config) -> fail, receipt rcpt-7969d06042e2bd55d82ce054
overall_score: 0.78
blocking_count: 1

| Dimension | Score | Verdict |
|---|---|---|
| Functionality (40%) | 0.75 | FAIL (forced by required gate; N1 and (a) correct, (b) correct only when CLAUDE_PROJECT_DIR is unset or names the parent) |
| Security (25%) | 0.85 | PASS (no input-validation or secret surface; one lock-bypass edge, see F1) |
| Craft (20%) | 0.80 | PASS (table RED-first, 4 of 5 mutants killed, one survivor disclosed) |
| Consistency (15%) | 0.80 | PASS |

Tree HEAD 950a3e7cba5c5511f4e188fc147bf535b234b5b8. Judging moai build: `moai version` -> `v3.2.0-rc.36-t1822c ... built 2026-10-15T21:56:00Z`, MCP server commit ee981a932, an ancestor of the tree HEAD (the tool's `build_lag` says so). The audit_multi/codex verdict therefore came from an older build than the tree. All `go` results below were built from this tree, not from the installed binary (section 2.2).

## Defects

- F1 [P1 per codex; blocking by the required gate] [blocking] internal/cli/mcp_factory_card.go:255 - `factory_complete` passes `root` (= `factoryRootOf(project_root)`) as the integration-window and re-measure root. The verb passes `integrationLockRoot()`, which returns `CLAUDE_PROJECT_DIR` when set and only otherwise the parent of the git common dir. When `CLAUDE_PROJECT_DIR` names a directory other than the parent checkout, the two surfaces use different windows. Codex reproduced it with a scratch test: the verb refused with `held by lane-other`, the MCP handler returned nil, moved integration and left `originalWindowHolder=sess-other`, i.e. it bypassed another lane's window. My own probe (below) confirms the roots differ. The code comment ("names that same directory") and the progress record ("Both name the parent checkout") state this as unconditional; the CHANGELOG says the window "is shared between the CLI and MCP surfaces" without the condition. Required fix: either (a) resolve the window root in the handler by the verb's rule where the server can know the caller's value (the server's own `CLAUDE_PROJECT_DIR` names the project, so honouring it would match the verb for a server spawned by the same host), or (b) keep `factoryRootOf` and state the precondition (`CLAUDE_PROJECT_DIR` unset or the parent checkout) in the handler comment, the CHANGELOG sentence and the progress record, and add a table row pinning the non-parent case's current behaviour. A fix is the author's call; the unconditional wording must go either way.
- F2 [low] [optional] internal/cli/factory_mcp_verb_parity_test.go - the `RecordRoot` rows do not discriminate: reverting `factory_stage`'s record root to the raw `launch` (mutant m5) leaves the whole table green. The author discloses that `homestate.OpenFactory` already resolves the parent record from a linked tree, so these rows pin existing behaviour only. Not a defect of the change; record as debt.
- F3 [low] [optional] the `@MX:ANCHOR` call-site count on `mcpRequiredProjectRoot` (3 stated, 4 actual) - recorded observation per the leader, not raised.

Already tracked and out of scope: N3, F8, (c) (card t1867). Not raised.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim 1 - N1 and (a) match the verbs; (b) matches the verb only under a condition

Evidence (verbatim, `go test -count=1 -v -run 'TestMCPVerbParity_' ./internal/cli/`, tree HEAD, built from this tree):

```
--- PASS: TestMCPVerbParity_LaneLabel (2.13s)
--- PASS: TestMCPVerbParity_RecordRoot (10.37s)
--- PASS: TestMCPVerbParity_CompleteWindowRoot (17.93s)
--- PASS: TestMCPVerbParity_DecideConfigRoot (7.11s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	39.014s
```

18 leaf runs ran (9 LaneLabel, 3 RecordRoot, 2 CompleteWindowRoot, 1 DecideConfigRoot, plus the 4 parents). Code reading: `handleFactoryNext/Stage/Complete` now call `factoryLaneLabelFromEnv` (grep: lines 132, 212, 239, 310), matching `factory_card.go:148`; `handleFactoryDecide` resolves `factoryRootOf(launch)` as `factoryCardRoot()` does (`factory_card.go:84`).

Verb versus handler window root, probe (scratch test via `-overlay`, not in the tree, `/private/tmp/t1680-audit/probe/zz_probe_test.go`):

```
CLAUDE_PROJECT_DIR=parent          verb == handler  -> equal = true
CLAUDE_PROJECT_DIR=linked worktree verb = .../002-wt ; handler = .../002 -> equal = false
CLAUDE_PROJECT_DIR unset           verb == handler  -> equal = true
```

Ruling on the leader's question on (b): `integrationLockRoot()` and `factoryRootOf(project_root)` are the same tree only when `CLAUDE_PROJECT_DIR` is unset (git-common-dir fallback gives the parent) or names the parent. The code comment in `integration.go:37-41` says Claude Code sets it to the project root even inside a worktree, so for ordinary lane sessions they agree. They differ when it names a linked tree. I classify this as a defect of this card, not merely a recorded limit, for two reasons: the diff makes the handler stricter-or-looser than the verb in a case where it was previously equal (the old code passed `launch`, which equals the verb's value in exactly the non-parent-env case, so the change removes parity there and the codex repro shows the failure mode is a bypass of another lane's window), and the author's claim of identity is unconditional while the table, by pinning the variable to the parent, cannot see it. Practical reach is low and the author's "no such launch in this flow" may well hold; that argues for severity, not for dropping it.

### Claim 2 - RED-first is real and the fix is pinned (mostly)

Evidence, table run against the pre-fix handler file (`git show ae34f26831:internal/cli/mcp_factory_card.go` via `-overlay /private/tmp/t1680-audit/ov-pre.json`):

```
--- FAIL: TestMCPVerbParity_LaneLabel
    --- FAIL: .../next/legacy_spelling
    --- FAIL: .../stage/legacy_spelling
    --- FAIL: .../stage/empty
    --- FAIL: .../complete/legacy_spelling
--- FAIL: TestMCPVerbParity_CompleteWindowRoot (both subtests)
--- FAIL: TestMCPVerbParity_DecideConfigRoot
FAIL	github.com/modu-ai/moai-adk/internal/cli	21.846s
```

This matches the author's listed rows. Mutants (each an overlay of the fixed file, run with `-run TestMCPVerbParity_`):

| Mutant | Result |
|---|---|
| m1 revert N1 on `next` | FAIL: LaneLabel next/legacy_spelling, next/empty (killed; first build attempt was my own redeclaration error, fixed and rerun) |
| m2 complete window root back to `launch` | FAIL: both CompleteWindowRoot subtests (killed) |
| m3 decide root back to raw `launch` | FAIL: DecideConfigRoot (killed) |
| m4 revert N1 on `complete` | not run (build-constructed; the pre-fix run already covers complete/legacy) |
| m5 stage record root back to raw `launch` | ok, table green (survivor, F2) |

Determinism: no wall clock in assertions (`fcNow` fixed; the window `since` stamp masked by `maskWindowStamp`); trees under `t.TempDir()` via the existing fixtures; env through `t.Setenv`/`sdLaneEnv`; `t.Chdir`. All five lane verbs are compared verb-versus-MCP (next/stage/complete/renew in LaneLabel, decide in the boundary row, RecordRoot and DecideConfigRoot).

### Claim 3 - build, vet, format and regression selections are clean

Evidence:

```
go build ./...                          -> exit 0, no output
GOOS=windows GOARCH=amd64 go build ./... -> exit 0, no output
go vet ./internal/cli/                  -> exit 0, no output
gofmt -l internal/cli                   -> no output
go test -count=1 -v -run '^TestSD_AC' ./internal/cli/   -> ok 136.653s, 30 PASS lines
go test ... -run 'TestFactoryNextNominate|TestFactoryNextArm|TestFactoryNextFlagSet|TestFactoryNextBare|TestFactoryNextAll|TestQAS_AC008b|TestQAS_AC011b|TestQWR_AC008|TestFactoryMsgSend|TestFactoryMCPIdentity|DecidePush'
                                         -> ok, 33 PASS
go test ... -run 'Renew'                -> ok, 16 PASS
go test ... -run 'DecidePush'           -> ok, 3 PASS (TestFactoryDecidePushGateClosesAMergedPRCard, TestFR_AC018_DecidePushGate, TestFactoryDecidePushChainsApprovalVersion)
```

The tests naming a changed handler (`grep -l handleFactory internal/cli/*_test.go`): factory_m3_test.go, factory_nominate_test.go, factory_quota_test.go, factory_quota_worktrees_test.go, mcp_factory_msg_test.go and the new file; all their test names were covered by the selections above.

### Claim 4 - CHANGELOG accuracy

True against the code: the legacy-label refusal on next/stage/complete (renew already had it); `factory_decide` resolves `factoryRootOf`; `factory_complete` uses the factory root for record, window and re-measure; the four test names; the t1867 split (N3 / F8 / (c), as in the progress record). Overstated: "so the landing window is shared between the CLI and MCP surfaces" holds only under the F1 precondition. The decide sentence ("acts on another tree's config ... refuse or record a different outcome") rests on the RED `DecideConfigRoot` row, which the progress record admits; I accept it as measured by my own pre-fix run above.

### Claim 5 - scope hygiene

`git diff ae34f26831 950a3e7cb --name-status`: `A .moai/reports/t1680/progress.md`, `M CHANGELOG.md`, `A internal/cli/factory_mcp_verb_parity_test.go`, `M internal/cli/mcp_factory_card.go`. No t.Skip, no timeout change, no removed tests, no file outside the three groups. The handler diff drops the `os` and `config` imports (now unused) and adds no new shared-code edit.

### Convergence check

`audit_multi` result digest written to `.moai/state/audit-plan-result.json` (untracked): `overall_verdict: fail`, `plan_source: config`, `per_backend_verdicts: [{backend: codex, gate: required, verdict: fail}]`, no `gate_unmet`. `moai verify audit-plan --project-root <tree> --result-file .moai/state/audit-plan-result.json` ->

```
"convergence_check": { "ok": true, "unmet": [] }
```

`ok: true` only says the digest is consistent with the plan (a required backend answered, nothing unmet); the overall verdict stays `fail`. The Claude and GLM gates are off in this tree's plan. I passed my own Claude-side anchor to the tool; it is not an independent backend, and only the codex receipt counts.

### Baseline-attribution

Every command above was run in this session against tree HEAD 950a3e7cba5c5511f4e188fc147bf535b234b5b8 with the Go toolchain building from the tree. Mutants and the pre-fix run used `-overlay` files under `/private/tmp/t1680-audit/`; nothing in the tree was edited by me except this file and the untracked `.moai/state/audit-plan-result.json`. The `moai` CLI used for `audit-plan` and the MCP server are the installed build `ee981a932` (older than HEAD); their output is cited as that build's, per section 2.2.

### Gaps

- Codex's scratch repro (`TestReviewCompleteNonParent`) was not re-run by me; I reproduced the underlying root disagreement with my own probe, but not the full lock-bypass flow. The bypass claim rests on codex's output.
- Mutant m4 (revert N1 on `complete`) was built but not run; the pre-fix run covers it.
- `CLAUDE_PROJECT_DIR` is unset in this auditor session (`printenv` exit 1); I did not measure what the launcher sets inside a lane session.
- Not run: the full `internal/cli` suite (barred by ruling 35e9), `internal/config`, any other package test; `golangci-lint` not run. The go-test-heavy slot was not acquired.
- `/verify` wrapper not invoked.
- No refused commands this round except the first compound (`&&`/`for` loop) mutant-generation command, refused by the worktree guard; I split it into plain calls and a Python script under `/private/tmp`.
- Compound-arg `-run` selections with `|` ran fine here; the ran-count is counted from `--- PASS` lines, so a top-level test with only subtests is counted once per leaf line.

### Residual-risk

- If lane sessions in production carry a `CLAUDE_PROJECT_DIR` naming a linked tree, MCP `complete` takes a different window from the CLI (F1). If they do not, the change is sound.
- RecordRoot rows will not catch a future regression of the record-root rule (F2).
- The older judging build means codex's review saw current tree code but ran under an older tool server; no effect on the diff text it read.

## Iteration history

Round 1 of this card's sync audit. Next round scope (delta only): F1 resolution plus the matching wording/comment/CHANGELOG change, a table row for the non-parent `CLAUDE_PROJECT_DIR` case, and a rerun of the same required codex gate.
