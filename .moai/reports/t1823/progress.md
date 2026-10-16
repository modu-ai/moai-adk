# t1823 progress — main red x2 (managed TUI opt-in probe, binlag coordinates)

Card t1823 (Class B, run → sync). Lane lane-38. Tree `.moai/worktrees/t1823`, branch `WT-main-red-managed-tui-binlag`, base `ec5a4f392`.
Rules carried from the dispatch: TMPDIR / GOTMPDIR / HOME not set; contaminating tests skipped on `./internal/cli` runs; `~/.gitconfig`, `backlog.db`, `~/.moai/db` read before and after; the t1812 landing remeasure has slot priority.

## §1 Cause determination (recorded BEFORE any repair of item 1)

### Item 2 — `TestAuditLagUsesBinlagSeam` (repaired first; pure coordinate drift)

Reproduction on the base, this run: `go test ./internal/cli -run 'TestManagedTUINeverReachedWithoutOptIn|TestAuditLagUsesBinlagSeam' -count=1 -v` →
```
mcp_build_identity_test.go:809: sweep: baseline hit local_main_resync.go:414 MISSING ...
mcp_build_identity_test.go:809: sweep: baseline hit integration.go:287 MISSING ...
mcp_build_identity_test.go:809: sweep: baseline hit local_main_resync.go:417 MISSING ...
mcp_build_identity_test.go:814: sweep: NEW ancestry hit local_main_resync.go:438 ...
mcp_build_identity_test.go:814: sweep: NEW ancestry hit integration.go:288 ...
mcp_build_identity_test.go:814: sweep: NEW ancestry hit local_main_resync.go:435 ...
```
Cause (measured with `git grep -n -E 'merge-base|is-ancestor' <commit> -- <file>`, `git log`, `git diff`):
- `local_main_resync.go`: at `9d438dc97` (the last t1616 commit) the two hits sit at 414/417, matching the want map. `git diff 9d438dc97 HEAD` for that file shows +29/-8: the `localMainResyncBeforeFastForward` test seam (+6 lines) and the one-mutation-section rewrite of the revalidate-and-fast-forward step (+15 lines) — 21 lines above the helper, so 414→435 and 417→438. The text first appears in merge `cd19e5494` (card t1619, 2026-10-14) under `git log -m --full-history -S'localMainResyncBeforeFastForward'`. No added line is an ancestry comparison.
- `integration.go`: card t1589 (`a310ff444`, 2026-10-14) added one line (`cmd.Env = gitenv.Env()` in `canonicalTreeSpelling`) above the doc comment: 287→288. The hit is a doc mention of the primitive, as the existing map comment says.
Same hits, same count; only coordinates moved. Repair: the want map moves 414→435, 417→438, 287→288 (comment records both causes).

### Item 1 — `TestManagedTUINeverReachedWithoutOptIn/opted_in_probes_once` (probe=0, want 1)

Reproduction on the base: `probe=0 tui spawns=0 direct door=0, want 1/0/0` (`managed_codex_tui_test.go:2120`); the other three subtests pass.
The subtest ignores the launch error (`_ = launch(t)`). With a temporary log of that error (reverted; the file is unchanged in the diff) the run printed:
`launch err: managed codex launch does not support --add-dir`
Chain, read from source:
1. `runCodex` for the `cli` verb appends `codexMoaiHomeAddDirArgs()` to `childArgs` (`internal/cli/codex_launcher.go`, the `kind == codexVerbLaunchCli` block): `--add-dir <moai home>` whenever `paths.MoaiHome()` is an existing directory. The test sets `MOAI_HOME` to a fresh temp dir, so it exists.
2. The managed divert (same function, `factoryManagedRequested && factoryLaunchEnabled`) then passes that same `childArgs` to the managed owner.
3. The owner's `managedCodexOptions` (`managed_codex_factory.go`) refuses every flag except `-c/--config`, `-m/--model`: `default: ... does not support %s`. It returns before `planOperatorTUI`, so the probe never runs.
Origin: commit `063a7620f` (2026-10-14, merged by `ad002d3a0`) "hand the moai home to codex as a writable dir on CLI and lane launches" added step 1. It neutralized the seam in the argv-pinning tests (`codex_launcher_test.go`, `launch_session_defaults_test.go`) but not in this one, and the managed owner's refusal predates it (2026-10-02, `033221ced`).
So this is a PRODUCTION-path regression, not only a test artefact: any `moai codex` CLI launch that takes the managed divert with an existing moai home hands `--add-dir` to an owner that refuses it. The fixtures in `managed_launcher_wiring_test.go` stub `managedFactoryCodexLaunchFunc`, so no other test reaches the owner's parser.

Repair options considered: (a) neutralize `codexMoaiHomeAddDirArgs` in this test's setup only — turns the test green and leaves the production refusal in place; (b) keep the add-dir for the exec doors and build the managed divert's argv without it, restoring exactly the pre-`063a7620f` argv the managed owner received. Choice: (b). Reason: the owner is a headless `app-server` process that never accepted this flag, and (a) would hide a defect an operator would hit. Recorded as a judgment for the leader to veto before landing.

### Gaps (cause stage)

- Not run: the managed divert against a real codex binary (the sentinel binary does not exist in this test); whether the managed App Server needs the moai home writable by another route is not established by this lane — the pre-`063a7620f` behavior is what (b) restores.
- No A/B of `063a7620f` itself (checkout of its parent) was run; the origin is read from the commit's file list, the diff of the `kind == codexVerbLaunchCli` block and the refusal's own date.

## §2 Repairs and evidence

Attribution, stated plainly: the lane (lane-38) edited both files below DIRECTLY in its own session; no `manager-develop` spawn was involved. The leader flagged this. Handling: the lane reverted its own production edit to `codex_launcher.go` before any commit, and the production fix plus a unit test were handed to a `manager-develop` spawn (its result is recorded in §5 once it returns). The coordinate edit to `mcp_build_identity_test.go` was already committed (`54944481c`) as a lane edit and stays attributed that way.

- `internal/cli/mcp_build_identity_test.go` (lane edit, committed `54944481c`): want map 414→435, 417→438 (`local_main_resync.go`), 287→288 (`integration.go`), reasons recorded in the map comments.
- `internal/cli/codex_launcher.go` (lane's first version, REVERTED, never committed): the managed divert handed the owner `append(slices.Clone(localArgs), codexChildArgs(kind, tail)...)`. The evidence lines below were measured with that first version in the tree; they are NOT the final production fix's evidence.

Evidence of the lane's first version (tree `ec5a4f392` + the two edits):
- `go test ./internal/cli -run '^(TestManagedTUINeverReachedWithoutOptIn|TestAuditLagUsesBinlagSeam)$' -count=1 -v` → both PASS, all four TUI subtests PASS incl. `opted_in_probes_once`.
- `go test ./internal/cli -run 'AddDir|ManagedLaunch|ManagedTUI|ManagedCodex|CodexLaunch|CodexDirect|RunCodex' -count=1` → `ok ... 35.873s`.
- `go vet ./internal/cli` and `gofmt -l` on the two files → no output.
- `golangci-lint run --new-from-rev=HEAD --timeout=10m ./internal/cli/...` → `0 issues.`, exit 0. (The first run with `--timeout=2m` printed `0 issues.` and then `Timeout exceeded`, exit 4 — not counted; re-run above.)
- `~/.gitconfig`: 501 lines immediately before and after the family run above (unchanged by it). Its size at the t1812 measurement was 489; the growth predates this run and its source was not investigated.

## §3 Gaps / residual risk

- Not run: the full `./internal/cli` package (heavy; gated by the `go-test-heavy` slot, which the t1812 landing remeasure holds priority for), so unrelated reds elsewhere in the package are neither confirmed nor excluded by this card.
- Item 1's production claim rests on the source chain and the commit file list; no real codex binary and no A/B at `063a7620f`'s parent. The App Server's own need for a writable moai home under the managed path is not established here.
- Lint's first attempt timed out; only the 10-minute re-run is cited.
- Not measured this run: `backlog.db` hash and `~/.moai/db` count before and after (asked for the heavy remeasure; the family run above is not that run and does not touch the queue).

## §4 A/B of the introducing commit (item 1)

Method: `git archive --format=tar 063a7620f^` (= `2e387c789e082f5beeac62d741ce6d99082a6343`) extracted into a clean directory; `go test -C <dir> ./internal/cli -run '^TestManagedTUINeverReachedWithoutOptIn$' -count=1 -v`; TMPDIR/GOTMPDIR/HOME at their defaults.
- At `063a7620f^`: all four subtests PASS, including `opted_in_probes_once`; `ok github.com/modu-ai/moai-adk/internal/cli 1.783s`.
- At base `ec5a4f392` (this tree, §1): `opted_in_probes_once` FAILS with `probe=0 tui spawns=0 direct door=0, want 1/0/0`, launch error `managed codex launch does not support --add-dir`.
The first commit that appears between them on this path is `063a7620f` (it adds the `--add-dir` to the CLI verb's childArgs; its file list is in §1). Not run: `063a7620f` itself as a separate point, so "introduced by that commit" rests on the commit's content plus this pass/fail pair spanning it, not on a third measurement.

## §6 Public record: lane direct edit (leader-requested wording)

Lane direct edit, scope: test coordinate map only — `internal/cli/mcp_build_identity_test.go`, commit `54944481c`: the three want-map keys (`local_main_resync.go` 414→435 and 417→438, `integration.go` 287→288) and their comments, nothing else. No `manager-develop` spawn produced this edit. The leader ruled it need not be redone. The sync audit's scope for this card INCLUDES this commit (it is part of the `base..HEAD` diff the audit reads), and the audit prompt names it explicitly as a lane-authored change.

## §5 manager-develop result (production fix + unit test), verified by the lane in this tree

Commit `f37be81f7` (not pushed): `internal/cli/codex_launcher.go` (+5/-1: the managed divert builds `managedArgs := append(slices.Clone(localArgs), codexChildArgs(kind, tail)...)`; exec/direct/spawn doors untouched) and `internal/cli/codex_moai_home_add_dir_test.go` (+52: `pinCodexMoaiHomeAddDir`, `TestCodexLaunch_ManagedDivertCarriesNoAddDir`, `TestCodexLaunch_DirectDoorKeepsAddDirWhenNotDiverted`). Authored by a `manager-develop` spawn; the lane read the commit's stat and the production diff and re-ran the tests below.

RED (reported by manager-develop, observed on the unfixed code before the fix; not re-observed by the lane — the fix is already committed):
```
codex_moai_home_add_dir_test.go:151: managed argv = ["/sentinel/path/codex" "--add-dir" "<tmp dir>" "--model" "m1"], must not carry --add-dir
--- FAIL: TestCodexLaunch_ManagedDivertCarriesNoAddDir
--- PASS: TestCodexLaunch_DirectDoorKeepsAddDirWhenNotDiverted
```

GREEN (lane, this tree at `f37be81f7`): `go test ./internal/cli -run '^(TestManagedTUINeverReachedWithoutOptIn|TestAuditLagUsesBinlagSeam|TestCodexLaunch_ManagedDivertCarriesNoAddDir|TestCodexLaunch_DirectDoorKeepsAddDirWhenNotDiverted)$' -count=1 -v` → all PASS (the new managed test, the new direct-door test, the four `TestManagedTUINeverReachedWithoutOptIn` subtests, `TestAuditLagUsesBinlagSeam`), `ok ... 2.759s`.
Family run (manager-develop): `go test ./internal/cli -run 'AddDir|ManagedLaunch|ManagedTUI|ManagedCodex|CodexLaunch|CodexDirect|RunCodex' -count=1` → `ok ... 50.926s` (not re-run by the lane at this commit; the lane ran the same selection earlier on its own reverted first version, 35.873s).

Static checks at the final state (lane): `golangci-lint run --new-from-rev=ec5a4f392 --timeout=10m ./internal/cli/...` → `0 issues.`, exit 0; `go vet ./internal/cli` and `gofmt -l internal/cli` → no output.

Gaps: the full `./internal/cli` package and `go test ./...` not run (heavy slot, held for t1812's remeasure). The managed divert was not run against a real codex binary. Whether the managed App Server needs the moai home writable by another route is not established; the change restores the pre-`063a7620f` argv the owner received.
Residual-risk: the unit test pins the add-dir seam to a temp dir, not the real `MOAI_HOME` derivation (that derivation stays covered by the existing `TestCodexLaunch_CliAddsMoaiHomeWritableDir`).

The §2 evidence block above describes the lane's own reverted first version and is superseded by this section for the production fix.

## §7 Remeasure composition (set before the run; the tree takes no commit after this one)

Command shape (leader instruction, cross-session): `moai integration remeasure -- go test -json -count=1 -timeout=150m -skip '<pattern>' ./internal/cli/`, TMPDIR/GOTMPDIR/HOME not set, `~/.gitconfig` line count, `backlog.db` hash and `~/.moai/db` count recorded before (immediately before the run) and after.

`-skip` = ONE anchored regex of 64 exact names: 43 `*SessionWorktree*` names (from `go test ./internal/cli -list SessionWorktree`, incl. `TestEnterSessionWorktree_ProfileSubcommand`); the 7 approved base failures (Codex wiring mirror x6, `TestHookEntryParity_NoDivergenceEitherDirection`); `TestRunWeb_AutoEntryOn_SuppressesAdvisory` and `TestWorktreeNew_WiresTheSharedMaterializer` (write the real `~/.gitconfig`, per the leader); `TestCountCodexAgentTOMLs` and `TestCountCodexAgentTOMLs_BadPatternRoot` (read the real `~/.codex/agents`, leader-approved); and the 10 `TestAC_CLOSURE_*` tests that use the `closuretest` fixture and dirty `~/.moai/db` (`contract_closure_test.go`: 001, 019, 020, 024; `mcp_audit_multi_record_test.go`: 012, 012_UnknownCard, 012_AppendFailure, 012_NoCardID, 012_TargetRecorded, 012_PathTraversalCardID — the two files are the only ones in `internal/cli` that mention `closuretest`, checked with `grep -rl`).
NOT skipped, by the leader's rule: this card's own fixed tests, `TestManagedTUINeverReachedWithoutOptIn` and `TestAuditLagUsesBinlagSeam`.
`internal/hook` is not in scope of this remeasure, so `TestNoNewEnvLiteralsInDiff` is not in the pattern.
Order: after the t1812 remeasure on the same lane; the slot is held by `leader-t1822` at the time of writing.
