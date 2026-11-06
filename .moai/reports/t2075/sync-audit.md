auditor-model: claude-opus-5-5

# t2075 sync-audit — Windows worktree migration no longer skips every tree for want of lsof

verdict: PASS-WITH-DEBT
audited_sha: 2a04549ec77bbb0f395c6487e52d28a61c09ba7f

- Card: t2075 (Class B, Go, internal/cli/worktree); no SPEC, so there is no `Binding run conditions` heading to dispose.
- Scope audited: the lane delta `6b53e7c1e7..2a04549ec7` only (6b53e7c1e7 = merge of t2070 2ac39de91c; t2070's own work is not under audit, only checked for breakage).
- Auditor: sync-auditor, Claude only, independent of the lane. Iteration: 1 (first audit of this card).
- Score: 0.86 (harmonic mean of the four dimensions below: 85 / 85 / 85 / 90).

## Dimension scores

| Dimension | Score | Verdict | Evidence (verbatim, commands below) |
|---|---|---|---|
| Functionality (40%) | 85 | PASS | `TestPlanMigration_OSGuardedProbeDoesNotSkipTrees` PASS, `TestPlanMigration_OtherProbeErrorStillSkipsTrees` PASS; `moai update` routes through `worktree.RunUpdateConsolidation` -> `planMigrationWith` (the fixed planner). Windows premise not executed here (debt D1). |
| Security (25%) | 85 | PASS | No new input surface; only new exec is `exec.Command("ping", "-n", "30", "127.0.0.1")` in a Windows test. The process-level move guard is weaker on Windows than lsof on POSIX (F2), with no data-loss path found (rename is atomic, failure handling classifies location). |
| Craft (20%) | 85 | PASS | gofmt -l rc=0 with no files listed; `go vet` host rc=0; `GOOS=windows go vet` rc=0; Windows test binary compiles and carries both Windows tests. Coverage not measured (Gap). |
| Consistency (15%) | 90 | PASS | Build-tag split mirrors the existing `sweep_cwd_posix.go` / `sweep_cwd_windows.go` and `migrate_lock_other.go` / `migrate_lock_windows.go` pairs; sentinel and comments match the file's style. |

Must-pass firewall (Functionality, Security): both PASS on their own.

## Findings (structured defect list)

- F1 [Medium] [optional, debt D1] [in-diff] internal/cli/worktree/migrate_cwd_windows.go:14 — the safety premise of option D (Windows refuses to rename a directory a process holds as its working directory) was not executed anywhere yet: the lane host and this auditor host are macOS. Confidence that the premise holds for native Win32 processes: high (it is documented Windows behaviour), but unobserved here. Required fix: observe `TestWindowsRefusesRenamingAProcessWorkingDirectory` and `TestWindowsProbeReportsOSGuard` green on a Windows runner (pr-multi-os-gate windows leg runs `go test ./internal/cli/worktree/...` for PRs to main touching `internal/cli/worktree/**`) before the batch reaches origin/main; on red, make the Windows probe return an ordinary error again (restores skip-all).
- F2 [Medium] [optional, debt D2] [in-diff] internal/cli/worktree/migrate_cwd_windows.go:3-9 — the OS guard covers "a process whose process-level cwd is the tree", which is narrower on Windows than what lsof-cwd catches on POSIX. Hypotheses NOT verified on a Windows host (confidence medium): (a) PowerShell `Set-Location` changes the shell location without changing the process working directory, so an interactive pwsh that `cd`'d into a tree does not hold it; (b) Cygwin/MSYS2 (Git Bash) may open the cwd handle with delete sharing to emulate POSIX, so a Git Bash shell inside a tree may not block the rename. Consequence if either holds: the tree moves under that shell (shell's later commands see a missing path); no content loss found — `git worktree move` is a single rename, and the manifest comparison plus reverse order still run. Moai/Claude sessions stay covered by the live-anchored-session guard. Required fix: a follow-up card that either measures (a)/(b) on Windows or implements the deferred option B (PEB read) as an additional skip cause; record the outcome.
- F3 [Low] [optional] [touched-file] internal/cli/worktree/migrate.go:1361-1377 (existing code, reached newly on Windows) — a busy tree on Windows now appears in the dry run as `planned` and at apply time as `Worktree migration failed <tree>: git worktree move: exit status ...` followed by `undo with: moai worktree migrate --rollback <journal>`, although nothing moved (atSource branch). The reason string carries only `exit status N` because `boundedOutput` returns `cmd.Output()`'s error without stderr. Not a correctness defect; the operator gets a misleading rollback hint. Required fix (optional): none for this card; consider a busy-specific message in a follow-up.
- F4 [Info] [optional] [outside] internal/cli/update_worktree_processes.go:13 — the package-cli lsof probe (used by `todo_auto.go:77`) still has no Windows variant; it degrades by design ("lsof unavailable — degraded measurement" note at todo_auto.go:204). Unaffected by this delta; listed only.

No blocking finding.

## Debts

- debt: D1 dispose_in=sync Windows execution of TestWindowsRefusesRenamingAProcessWorkingDirectory and TestWindowsProbeReportsOSGuard must be observed green on a Windows runner before the batch carrying this card lands on origin/main; red reverts the Windows probe to an ordinary error (skip-all).
- debt: D2 dispose_in=sync file a follow-up card for the Windows guard gap (PowerShell location vs process cwd, MSYS/Cygwin cwd handle sharing — unverified hypotheses) or option B (PEB read) as an extra skip cause.
- debt: D3 dispose_in=sync the tree's audit plan names codex as a required gate; this audit ran Claude-only at the caller's explicit instruction and called no codex tool, so the required codex leg is unmet by this audit (card-review.md records a static, advisory codex review pass on a wider base). The leader decides whether that review discharges the gate.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim

1. The delta touches only the six named files.
2. With the change, an OS-guarded probe leaves the tree unskipped; any other probe error still skips with the old reason.
3. Without the one condition change, the OSGuarded test fails (the RED is not vacuous).
4. Per-OS file selection is right: windows gets `migrate_cwd_windows.go` + `migrate_oscwd_windows_test.go`, darwin gets `migrate_cwd_posix.go`; both build.
5. No other caller of `defaultActiveProcessCWDs` is broken; `update_worktree_processes.go` / `todo_auto.go` are separate and unchanged; `moai update` reaches the fixed planner.
6. A Windows busy tree, if the OS refuses the rename, ends as a `failed` row with the tree left at its source, never `moved`.
7. POSIX behaviour is unchanged.

### Evidence

Claim 1 — `git -C <tree> diff 6b53e7c1e7..HEAD --stat`:
```
 internal/cli/worktree/migrate.go                   | 28 ++++--------
 internal/cli/worktree/migrate_bounded_test.go      |  7 +++
 internal/cli/worktree/migrate_cwd_posix.go         | 30 ++++++++++++
 internal/cli/worktree/migrate_cwd_windows.go       | 16 +++++++
 internal/cli/worktree/migrate_oscwd_t2075_test.go  | 52 +++++++++++++++++++++
 .../cli/worktree/migrate_oscwd_windows_test.go     | 53 ++++++++++++++++++++++
 6 files changed, 168 insertions(+), 18 deletions(-)
```
The full diff was read: the posix file is the removed function body moved verbatim (same lsof argv, same parsing); migrate.go loses the `bytes` import, gains the sentinel and changes one condition; migrate_bounded_test.go adds `runtime` and two `t.Skip` on windows. No stray change.

Claim 2 — `go test ./internal/cli/worktree -run 'OSGuarded|OtherProbeError' -count=1 -v`:
```
=== RUN   TestPlanMigration_OSGuardedProbeDoesNotSkipTrees
--- PASS: TestPlanMigration_OSGuardedProbeDoesNotSkipTrees (3.03s)
=== RUN   TestPlanMigration_OtherProbeErrorStillSkipsTrees
--- PASS: TestPlanMigration_OtherProbeErrorStillSkipsTrees (4.61s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	8.504s
```

Claim 3 — reasoned from the diff, no tracked file modified. The pre-change condition at migrate.go:589 was `if processCWDErr != nil {`. The test's probe returns `errProcessCWDOSGuarded` (non-nil), so the old branch sets `p.skip = "cannot inspect active process working directories: process working directories are guarded by the operating system"`, and the test asserts `skip == ""` -> it fails. That matches the RED text the lane recorded in progress.md. The test is also not passing for an unrelated reason: the control test, same fixture, reaches the same branch and reports the probe skip (the tree is listed and reaches the probe case), and the OSGuarded run shows every later guard (registry, anchored session, source/destination) leaves the tree unskipped. `m6NewWorld` restores the seam via `t.Cleanup` (migrate_command_test.go:86-88), so `planSkipFor` overwriting it leaks nothing.

Claim 4 — `GOOS=windows go vet ./internal/cli/worktree` -> no output (rc 0). `GOOS=windows go list -f '{{.GoFiles}} {{.TestGoFiles}}' ./internal/cli/worktree` -> GoFiles include `migrate_cwd_windows.go migrate_lock_windows.go ... sweep_cwd_windows.go`, no `migrate_cwd_posix.go`; TestGoFiles include `migrate_oscwd_t2075_test.go migrate_oscwd_windows_test.go`. Host `go list` -> GoFiles include `migrate_cwd_posix.go migrate_lock_other.go ... sweep_cwd_posix.go`; TestGoFiles include `migrate_oscwd_t2075_test.go` and not `migrate_oscwd_windows_test.go`. `GOOS=windows go test -c -o <scratch>/worktree_windows.test.exe ./internal/cli/worktree` -> rc=0; `go tool nm` on it lists `worktree.TestWindowsProbeReportsOSGuard`, `worktree.TestWindowsRefusesRenamingAProcessWorkingDirectory`, `worktree.TestPlanMigration_OSGuardedProbeDoesNotSkipTrees`, `worktree.TestPlanMigration_OtherProbeErrorStillSkipsTrees`, `worktree.defaultActiveProcessCWDs`. Host `go vet ./internal/cli/worktree` -> rc=0. `gofmt -l` on the six delta files -> no files, rc=0.

Claim 5 — `grep -rn "defaultActiveProcessCWDs|errProcessCWDOSGuarded|activeProcessCWDProbe|\"lsof\"" internal cmd`: `defaultActiveProcessCWDs` is referenced only by migrate.go:288 (seam default), migrate_bounded_test.go:186/206 (now skipped on windows) and migrate_oscwd_windows_test.go:21. `grep -rnw activeProcessCWDs internal/cli` -> `update_worktree_processes.go:13` (definition, own lsof) and `todo_auto.go:77` (its only caller) — not touched by the delta. `update.go:366,410` -> `runUpdateWorktreeMigration` -> `update_worktree_migration.go:32 return worktree.RunUpdateConsolidation(root, dryRun, out)` -> `runMigrationEngineWith` -> `planMigrationWith`: the card's symptom path is the fixed one.

Claim 6 — code read (migrate.go:1112-1160, 1199-1223, 918-1025, 1341-1377): before the move only the manifest is read, the instance token is written into the git admin dir (`movelist.go:157-178`, not the working tree), the journal pending entry and the move-list intent line are appended. `forwardGitMove` runs `git worktree move` (a single rename). On error, `forwardFailureLocation`: destination absent -> source carries the journal ident -> `atSource` -> a `reversed` release line is appended and the outcome is `failed` ("git worktree move: <err>"). `runMigrationEngineWith` counts it under `report.failed`; `printUpdateReport` prints `Worktree migration failed <src>: <reason>` and the summary counts it as failed. A failed tree does not make `runUpdateConsolidation` return an error (only move-list write errors and journal-settle errors do), so `moai update` is not failed by a busy tree. Not executed on Windows (Gap).

Claim 7 — `migrate_cwd_posix.go` body is byte-for-byte the removed function (diff read); the condition change can only differ for an error that `errors.Is` the new unexported sentinel, which only the windows file returns.

Security probe — grep for `exec.|os/exec|token|secret|password` over the four new files -> only `migrate_oscwd_windows_test.go:14 "os/exec"` and `:35 exec.Command("ping", "-n", "30", "127.0.0.1")` (fixed argv, test-only).

Audit plan — `moai verify audit-plan --project-root <tree>` -> `"config_status": "ok"`, `"model": "codex"`, codex gate `required`, claude and glm `off`, `"cross_model_active": true`, `"enforced_required": ["codex"]`.

### Baseline-attribution

- Tree: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2075, HEAD `git rev-parse HEAD` = 2a04549ec77bbb0f395c6487e52d28a61c09ba7f, status clean at session start; diff base 6b53e7c1e7.
- Toolchain: host `go` (darwin) and `GOOS=windows` cross-builds of the same tree; tests and vet build from source in this tree, so no installed-binary lag applies to them.
- `moai verify audit-plan` ran on the installed moai binary (the MCP server banner reports v3.2.0-rc.41, commit 9b632c1eb3, which is the tree's merge base, an ancestor of HEAD). The delta touches no audit-plan code, so the lag does not bear on that reading.
- CI: `.github/workflows/ci.yml` `test` job runs on `ubuntu-latest` only; `pr-multi-os-gate.yml` runs `go test -count=1 -json -timeout 25m ./internal/hook/... ./internal/cli/worktree/...` on `[macos-latest, windows-latest]` for PRs to main whose paths include `internal/cli/worktree/**`; `release-pr-multi-os.yml` has a windows leg. Read from the workflow files, not from a run.

### Gaps

- G1: the Windows-only tests were compiled (binary carries them) but not executed: no Windows host available. The OS premise of option D is unobserved in this audit (debt D1).
- G2: no real `git worktree move` of a busy tree on Windows was run; the `failed` row text and the Git for Windows rename retry behaviour (it retries on access-denied and only asks interactively when stdin and stderr are terminals; `exec` gives the child a null stdin) are reasoned from knowledge of Git for Windows, not observed. `gitMoveTimeout` bounds the call regardless.
- G3: F2's hypotheses (PowerShell location vs process cwd; MSYS/Cygwin cwd sharing) are not verified; they are reported as risks, not defects.
- G4: only the two focused tests ran (caller constraint). The rest of the package, including t2070's tests, was not executed; breakage of them is excluded only by compilation of the whole test package on darwin and windows and by reading that the delta moves code without changing it on POSIX.
- G5: coverage and golangci-lint were not measured by this audit (the lane reports lint 0 issues; not re-measured).
- G6: the plan's required codex leg was not run: the caller instructed a Claude-only audit with no codex/glm/audit_multi/claude_audit call. No MCP audit tool was called, so no `audit_multi` digest exists and no receipt was issued. If the stop-time receipt guard asks for a codex receipt, this audit does not comply, per the caller's instruction (debt D3).
- G7: no tool call was refused during this audit; `git` was run with `-C <tree>` and single commands only.

### Residual-risk

- If the Windows premise is false for native processes (D1 red), a tree with a process inside it can be moved on Windows; detection waits on the Windows CI leg.
- Even with the premise true, shells whose location is not their process cwd may not hold the tree (F2), so a Windows tree can move under an idle interactive shell; no content loss path was found.
- The dry run on Windows no longer shows any busy tree as skipped; operators see busy trees only as `failed` rows at apply time, with a rollback hint that does not apply (F3).
