auditor-model: claude-opus-5-5

# t2089 sync-audit — `moai factory wait` takes `--run`

verdict: PASS
audited_sha: e351fc5049560919f2b8a10973c9bc0c1960005d

Card: t2089 (Class A, no SPEC). Range audited: base 9b632c1eb3 .. head e351fc5049 (one commit, 2 files, +58/-4).
Overall verdict: **PASS**, score **0.93** (weighted: Functionality 96, Security 95, Craft 88, Consistency 92).
Iteration: 1 (first audit of this card).

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 96/100 | PASS | Focused suite `ok github.com/modu-ai/moai-adk/internal/cli 17.599s` (28 top-level tests selected, the two new ones among them); pre-fix RED `unknown flag: --run` reproduced; mutant M1 (flag declared, ignored) killed; auditor probes confirm the named run's `--after` hint is read (`condition=predecessor-landed t0b`) and an inactive run is refused stale |
| Security (25%) | 95/100 | PASS | The new input reaches only `resolveFactoryCardRun` → `factoryRunRowNotActive` (`SELECT status FROM runs WHERE run_id=?`, bound parameter) and `db.LoadCard(ctx, runID, cardID)`. No new string-built SQL, no shell, no path build from the value |
| Craft (20%) | 88/100 | PASS | `go vet` exit 0, `gofmt -l` empty; function coverage under the focused wait tests: `factoryRecordWait 88.2%`, `newFactoryWaitCommand 94.1%`. Debt: the committed tests pin only that the run resolves, not the end-to-end hint read nor the stale refusal (F1) |
| Consistency (15%) | 92/100 | PASS | Flag help ends with the sibling clause `(default: the single active run; a run whose row is not active refuses stale)`, matching factory_card.go:2550/2874/3079; same resolver (`resolveFactoryCardRun`) as stage/renew siblings. Doc synopsis not updated (F3, outside) |

Must-pass firewall (Functionality, Security): both pass. No binding run conditions apply (no SPEC, no `progress.md` binding-conditions heading).

## Findings

- F1 [Low] [optional] [in-diff] internal/cli/factory_wait_run_t2089_test.go:22-37 — the subject assertion checks that the refusal moves from "run is unresolved" to "card record is unreadable"; a nonexistent run id also produces that message (a missing run row is not refused by `factoryRunRowNotActive`), so the test does not prove the named run's card is read, and the stale refusal on `wait` is not pinned. Auditor probes in a scratch copy confirmed both behaviours are correct today. Required fix (optional): add a case that seeds the card with `HintAfter` in two runs and asserts the hint of the `--run` one, and one retired-run case.
- F2 [Low] [optional] [in-diff] internal/cli/factory_wait.go:137-152,276-279 — `--run` is accepted and ignored with `--resolve`, `--ack`, `--operator`, and `--on` (those paths read waits by card id, not by run). Consistent with the test comment's intent ("a lane can pass --run on every call") and stated in the author's Gaps; no wrong result was observed (probe: `--resolve --run runx` → `card t2089 has no operator wait to resolve`). Required fix (optional): none, or mention it in the help text.
- F3 [Low] [optional] [outside] .claude/rules/moai/workflow/factory-dispatch.md:134 and its template mirror — the `moai factory wait` synopsis does not list `[--run <id>]`. Not part of this card's diff; no verdict effect. Required fix (optional): a later doc card adds the flag to both copies.
- F4 [Info] [optional] [in-diff] internal/cli/factory_wait.go:173 — the help text is longer than the siblings (it explains when the run is read). Wording is accurate. No fix needed.

No blocking findings. No repair-born findings (first audit).

## Claim

1. The change wires a `--run` flag from the cobra command through `factoryRecordWait` into `resolveFactoryCardRun(ctx, root, run)`, replacing the hard-coded `""`; `factoryRecordWait` has exactly one caller.
2. The new tests fail on the base code for the stated reason (`unknown flag: --run`) and the control assertion passes on the base code.
3. The control assertion depends on two live runs.
4. A flag that is declared but ignored would fail the new test.
5. With `--run` naming an active run, the hint of that run's card is used; with `--run` naming a non-active run, the verb refuses with the resolver's stale message.
6. The focused suite, `go vet`, and `gofmt` are clean on the audited tree.

## Evidence

All commands run in this audit, on 2026-11-06, toolchain `go version go1.26.8 darwin/arm64`.

**E1 — diff and callers** (claim 1)

```
$ git show e351fc5049560919f2b8a10973c9bc0c1960005d --stat
 internal/cli/factory_wait.go                |  9 ++---
 internal/cli/factory_wait_run_t2089_test.go | 53 +++++++++++++++++++++++++++++
 2 files changed, 58 insertions(+), 4 deletions(-)

$ grep -rn "factoryRecordWait(" --include='*.go' .
internal/cli/factory_wait.go:161:			return factoryRecordWait(ctx, cmd.OutOrStdout(), factoryCardRoot(), args[0], on, operator, waker, reason, actor, run, scopeArg, nextArg)
internal/cli/factory_wait.go:264:func factoryRecordWait(ctx context.Context, out io.Writer, root, cardID, on string, operator bool, waker, reason, actor, run string, scope, next *string) error {

$ git show 9b632c1eb3:internal/cli/factory_wait.go | grep -n '"run"\|resolveFactoryCardRun'
282:			runID, err := resolveFactoryCardRun(ctx, root, "")
```

Other references to the wait command: `internal/cli/factory_handoff_recover.go:95: factory.AddCommand(newFactoryWaitCommand())` (registration only). Resolver (factory_card.go:3949-3961): an explicit run is refused when its row exists and is not active (`run %q is not the currently active factory run; pass the active run or omit --run`); otherwise used as-is.

**E2 — focused suite on the audited tree** (claim 6)

```
$ timeout 560 go test ./internal/cli/ -run 'FactoryWait|FactoryWake|FactoryAck|FactoryResolve|FactoryCard' -count=1
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	17.599s

$ go test ./internal/cli/ -list 'FactoryWait|FactoryWake|FactoryAck|FactoryResolve|FactoryCard'
exit=0
28            (count of lines starting with "Test")
21:TestFactoryWaitTakesRunWhenSeveralRunsAreActive
22:TestFactoryWaitAcceptsRunAlongsideOn

$ go vet ./internal/cli/
vet exit=0

$ gofmt -l internal/cli/
gofmt exit=0          (no files listed)
```

**E3 — pre-fix RED** (claim 2). Scratch copy from `git archive` of e351fc5049, with `internal/cli/factory_wait.go` replaced by `git show 9b632c1eb3:internal/cli/factory_wait.go`:

```
$ go test ./internal/cli/ -run 'TestFactoryWaitTakesRun|TestFactoryWaitAcceptsRun' -count=1 -v
    factory_wait_run_t2089_test.go:30: wait does not take --run: unknown flag: --runError: unknown flag: --run
--- FAIL: TestFactoryWaitTakesRunWhenSeveralRunsAreActive (0.65s)
    factory_wait_run_t2089_test.go:48: wait with --on and --run: unknown flag: --run
--- FAIL: TestFactoryWaitAcceptsRunAlongsideOn (0.34s)
FAIL
```

The control (line 18-19) did not fire on the base code; the first failure is line 30, the "unknown flag" assertion.

**E4 — mutant M1: flag declared but ignored** (claim 4). Scratch copy, fixed code with `resolveFactoryCardRun(ctx, root, run)` → `resolveFactoryCardRun(ctx, root, "")`:

```
--- FAIL: TestFactoryWaitTakesRunWhenSeveralRunsAreActive (0.33s)
    factory_wait_run_t2089_test.go:33: --run runa01 was not used to resolve the run: factory wait: no --on given and the run is unresolved: no --run given and no single active factory run (AMBIGUOUS_FACTORY: runa01 (owner live, ...), runb02 (owner live, ...))
    factory_wait_run_t2089_test.go:36: want the card-record refusal once the run resolves, got: factory wait: no --on given and the run is unresolved: ...
FAIL
```

Killed. (`TestFactoryWaitAcceptsRunAlongsideOn` cannot kill M1 by design: with `--on` the run is never read.)

**E5 — control depends on two live runs** (claim 3). Scratch copy, fixed code, second `seedLiveRun` removed:

```
    factory_wait_run_t2089_test.go:19: control: want the unresolved-run refusal, got err=factory wait: no --on given and the card record is unreadable: factory card not found ...
--- FAIL: TestFactoryWaitTakesRunWhenSeveralRunsAreActive (0.25s)
```

With one live run the resolver succeeds, so the control assertion is genuinely tied to the two-run ambiguity.

**E6 — auditor behaviour probes** (claim 5). Scratch-only test file `zz_audit_probe_t2089_test.go` in the copy (never in the audited tree):

```
    zz_audit_probe_t2089_test.go:22: runb02 out: wait recorded: card=t2089 condition=predecessor-landed t0b waker=lane-3
--- PASS: TestAuditProbeT2089HappyPathReadsNamedRun (0.30s)
    zz_audit_probe_t2089_test.go:42: stale err: factory wait: no --on given and the run is unresolved: run "runb02" is not the currently active factory run; pass the active run or omit --run
--- PASS: TestAuditProbeT2089StaleRunRefused (0.33s)
    zz_audit_probe_t2089_test.go:52: resolve+run err: factory wait: card t2089 has no operator wait to resolve
    zz_audit_probe_t2089_test.go:54: ack+run err: factory wait: card t2089 has no recorded wait to acknowledge
--- PASS: TestAuditProbeT2089RunWithResolveAck (0.21s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.633s
```

Setup for the happy path: two live runs, the card seeded in runa01 with `HintAfter: t0a` and in runb02 with `HintAfter: t0b`; `--run runb02` recorded `t0b`. Stale path: runb02 set to `status='retired'`.

**E7 — coverage of the changed functions** (Craft)

```
$ go test ./internal/cli/ -run 'FactoryWait|FactoryWake|FactoryAck|FactoryResolve' -count=1 -coverprofile=<scratch>/t2089.cov
exit=0
factory_wait.go:125:	newFactoryWaitCommand		94.1%
factory_wait.go:180:	factoryAckWake			85.7%
factory_wait.go:230:	factoryResolveOperatorWait	86.4%
factory_wait.go:264:	factoryRecordWait		88.2%
factory_card.go:3949:	resolveFactoryCardRun		75.0%
```

**E8 — audited tree left untouched**

```
$ git -C /Users/goos/.moai/worktrees/projects/moai-adk-go/t2089 status --short
(no output)
$ git -C /Users/goos/.moai/worktrees/projects/moai-adk-go/t2089 rev-parse HEAD
e351fc5049560919f2b8a10973c9bc0c1960005d
```

## Baseline-attribution

- Tree: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2089`, HEAD `e351fc5049560919f2b8a10973c9bc0c1960005d` (E8), clean before and after.
- Pre-fix baseline: `9b632c1eb3544cef5f275e502922a375e48d20aa`, read with `git show 9b632c1eb3:internal/cli/factory_wait.go` (E1, E3).
- Mutant and probe runs (E3-E6) ran on a scratch extraction of `git archive e351fc5049…` under `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/10944804-d6fb-4180-9533-737a355e2e6e/scratchpad/mut`, one file changed per probe as stated.
- Judging build: the Go toolchain `go1.26.8 darwin/arm64` compiled the tree under test directly; no installed `moai` binary judged anything (§2.2 does not bind a measurement taken by the compiler on the tree itself).
- Author's claims checked against this run: RED `unknown flag: --run` (confirmed, E3), control passing on base (confirmed, E3), focused tests green (confirmed, E2; the author's run omitted `FactoryCard`, this audit included it), vet/gofmt clean (confirmed, E2).

## Gaps

- **Cross-model required leg not run, on operator instruction (2026-11-06).** The configured cross-model audit (Codex / `audit_multi`) was deliberately not invoked; no `mcp__moai__*` audit tool and no backend-invoking `moai verify` command was called. `moai verify audit-plan` was also not run, so the plan's `cross_model_active` value for this tree was not read. No audit receipt exists for this verdict; if a receipt guard refuses this PASS for that reason, that refusal is expected and was not worked around.
- Not run: the full `internal/cli` suite, other packages, `-race`, `golangci-lint`, other OSes. The focused set covered every wait/wake/ack/resolve test and the `FactoryCard` tests.
- The stale-run and happy-path behaviour (E6) was verified only by auditor probes in a scratch copy; they are not committed tests (F1).
- Coverage figures (E7) are function-level under the focused wait tests only, not a package total.

## Residual-risk

- The committed test would still pass for an implementation that forwards `--run` but then reads the card from a different run; only the scratch probe (E6) shows the named run's hint is read. Low likelihood, since the code passes `runID` straight into `db.LoadCard`.
- `--run` naming a run id with no row at all is not refused (the shared resolver only refuses an existing non-active row); the verb then fails on the missing card record. This is the shared resolver's documented behaviour, same as the sibling verbs, not new in this card.
- Lanes reading the doctrine synopsis (F3) will not learn of the flag from the rule text.

## Operational Notes (unverified)

- A follow-up doc touch could add `[--run <id>]` to the `moai factory wait` synopsis in `factory-dispatch.md` (both the live copy and `internal/template/templates/`), with the template-first rebuild.
- If the leader wants F1 closed before landing, the auditor probe file content in the scratch copy is a ready starting point for two extra test cases.
