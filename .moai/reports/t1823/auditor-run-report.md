auditor-model: claude-sonnet-5-5

# t1823 sync-phase audit run report

Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1823, HEAD 1a82b59917e2f17a0b58196455bee0982b31de14, base ec5a4f392.

## Claim

C1 the four target tests pass at HEAD. C2 the two new tests are meaningful. C3 production diff is minimal and correct. C4 want-map edit moves coordinates only. C5 nothing else changed in internal/.

## Evidence

1. `git diff --stat ec5a4f392..HEAD`
```
 .moai/reports/t1823/progress.md              | 94 ++++++++++++++++++++++++++++
 internal/cli/codex_launcher.go               |  6 +-
 internal/cli/codex_moai_home_add_dir_test.go | 52 +++++++++++++++++++++++
 internal/cli/mcp_build_identity_test.go      | 15 ++++-
 4 files changed, 163 insertions(+), 4 deletions(-)
```
`git diff --name-only` lists exactly those four paths. C5 PASS.

2. `go test ./internal/cli -run '^(TestManagedTUINeverReachedWithoutOptIn|TestAuditLagUsesBinlagSeam|TestCodexLaunch_ManagedDivertCarriesNoAddDir|TestCodexLaunch_DirectDoorKeepsAddDirWhenNotDiverted)$' -count=1 -v`
```
--- PASS: TestCodexLaunch_ManagedDivertCarriesNoAddDir (0.00s)
--- PASS: TestCodexLaunch_DirectDoorKeepsAddDirWhenNotDiverted (0.00s)
--- PASS: TestManagedTUINeverReachedWithoutOptIn (0.09s)
    --- PASS: .../no_switch  .../switch_only  .../stamps_only  .../opted_in_probes_once
--- PASS: TestAuditLagUsesBinlagSeam (1.17s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.660s
```

3. `go test ./internal/cli -run 'AddDir|ManagedLaunch|ManagedTUI|ManagedCodex|CodexLaunch|CodexDirect|RunCodex' -count=1`
```
ok  	github.com/modu-ai/moai-adk/internal/cli	43.903s
```

4. `go vet ./internal/cli` -> no output, exit 0. `gofmt -l internal/cli` -> no output.

5. `grep -n -E 'merge-base|is-ancestor' internal/cli/local_main_resync.go internal/cli/integration.go`
```
internal/cli/integration.go:288:// base died with `merge-base ..HEAD ...
internal/cli/local_main_resync.go:435:// included. `git merge-base --is-ancestor` exits 1 ...
internal/cli/local_main_resync.go:438:	_, err := git("merge-base", "--is-ancestor", a, b)
```
Exactly the three coordinates the want map now lists (288, 435, 438), same count as before (3). The production diff adds no ancestry comparison (its only added lines are in codex_launcher.go). C4 PASS.

6. Production read (codex_launcher.go ~1160-1232). `childArgs := localArgs`; the add-dir append happens only for `kind == codexVerbLaunchCli`. In the managed divert the new code builds `append(slices.Clone(localArgs), codexChildArgs(kind, tail)...)`. `localArgs` keeps its own length, so the clone holds only the local `-c developer_instructions=...` pair (codexLocalDeveloperInstructions returns `[]string{"-c", "developer_instructions=..."}`), which the owner's `managedCodexOptions` accepts. No aliasing hazard: the clone is taken from the unextended slice header, and the earlier `append(childArgs, ...)` cannot change `len(localArgs)`. Non-managed doors still use `req.Args` / `childArgs`, byte-identical to before; the factory override splice (`slices.Insert(req.Args, overrideAt, ...)`) sits after the divert and operates on `req.Args`, so it is untouched. C3 PASS.

7. Owner parser (managed_codex_factory.go:640-672): `managedCodexOptions` accepts only -c/--config and -m/--model and returns "managed codex launch does not support %s" for anything else, so a leading `--add-dir <path>` is refused. This confirms the premise of the fix.

8. C2 by reading. ManagedDivertCarriesNoAddDir: pins the add-dir seam to a non-empty pair (so without the fix argv would contain `--add-dir`), asserts `managed.calls == 1`, asserts `--add-dir` absent from `managed.args`, asserts `--model` present. It would fail against the old line, which passed `childArgs` (containing the pinned pair). DirectDoorKeepsAddDirWhenNotDiverted asserts `codexAddDirValue(argv) == dir` on the direct door; it is a regression guard that passes both before and after the fix by design. How to prove the fail-without-fix half without editing the tree: build the base commit in a separate temporary worktree (`moai worktree new`) and run the first test there; not done here. C2 PASS (by reading + execution of the post-fix state).

## Baseline-attribution

Tree HEAD 1a82b59917e2f17a0b58196455bee0982b31de14, clean at start. Tests were built by `go test` from this tree in this run; no installed `moai` binary judged anything (the tests compile the package directly), so no tool-lag coordinate applies.

## Gaps

- Not run: full `./internal/cli` package, `go test ./...`, `golangci-lint`, coverage measurement, race detector (heavy slot not mine; instruction forbade).
- Not observed: the pre-fix failure of the two reds and of the new test (no checkout/stash/edit allowed). The progress.md A/B section was read only as a pointer, not re-measured here.
- Not verified: the claim that card t1589/t1619 are the cause of the coordinate shifts (the commit attributions in the comments); only the resulting coordinates were checked, which is all the want map asserts.
- Not run: the managed owner end-to-end against a real codex binary.
- No cross-model audit_multi run (task scoped to narrow commands); the audit-plan verb was not queried.
- The binding-run-conditions heading in progress.md was not examined.

## Residual-risk

- The managed argv drops moai-home write access, so a managed headless session relies on whatever sandbox config the owner applies; that is the intended trade (flag unsupported), but whether the managed session can write ~/.moai databases is not tested here.
- The want-map is a line-number coordinate list; it will rot again with any edit above those lines (pre-existing design, optional hardening).
- The lane-authored direct edit of 54944481c was judged on its content (C4 passes); the process deviation itself is a leader matter, not a code defect.

## Findings

- F1 [low] [optional] internal/cli/codex_moai_home_add_dir_test.go:~150 - the managed test checks `--add-dir` absence and `--model` presence but not that the developer-instructions `-c` pair rides through when local instructions exist; the production code path for that is shared and unchanged semantically. Not blocking.
- F2 [low] [optional] internal/cli/mcp_build_identity_test.go:~676 and ~790 - coordinate-keyed want map is brittle by construction; this is the third re-measurement. Not blocking, out of scope for this card.

No blocking findings. What would change my mind: a pre-fix run showing the new managed test passing without the production change, or an added line in the diff that is itself an ancestry comparison, or the tests failing on a rerun.

## Per-claim

- C1 PASS
- C2 PASS (fail-without-fix reasoned, not executed)
- C3 PASS
- C4 PASS
- C5 PASS

auditor_verdict: PASS
