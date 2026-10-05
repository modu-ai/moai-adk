# t231 — `worktree clean` Distinguished Degraded-Run Signal (2026-09-02, lane-7)

## Claim

`moai worktree clean` now ends with the intentional exit code 2 when the authoritative
anchor (lock) source cannot be read — the t209 CR OD-2 request that was correctly ruled
unwireable at audit time, because every error return collapsed to exit 1 at `cmd/moai/main.go`.
The card's three scope items are all discharged: (1) the exit-2 signal is wired through all
three caller-facing clean limbs; (2) REQ-WR-016 is revised in SPEC-WORKTREE-REAPER-001 v0.4.1
to state the distinguished preservation signal; (3) the `cmd/moai` ExitCoder wiring was
RE-READ and needed no change — `main.go:18` resolves `ExitCoder` via `cli.ResolveExitCode`
(`internal/cli/exitcode.go`), proven in production by `moai worktree verify` (0/1/2/3); what
was missing was `clean` using the vehicle, not the vehicle itself.

## Design

- **Vehicle reuse, not a second one** (card's t263-style trap avoided by construction): the
  in-package `ExitCodeError` (`internal/cli/worktree/guard.go`) already exists for exactly
  this purpose; it gained an optional `Detail` field — zero `Detail` renders the old guard
  wording byte-identically, so `worktree verify`'s contract is untouched.
- **Limb semantics** (one code, two shapes):
  - `--merged-only`: fail-closed abort as before (no dirty guard of its own — without the
    anchor source nothing stands between the sweep and a live tree) → exit 2, "sweep
    aborted, nothing removed".
  - `--stale` and `--stale --json`: the sweep COMPLETES (keeps every tree, anchors
    undetermined; JSON printed in full first) → then exit 2. The signal never means work
    was destroyed; it only stops the exit status from telling a lie of silence.
- **`prMergeCleanup` untouched**: internal consumer, callers receive a function return —
  its non-blocking contract (AC-WR-024 / AC-WR-024b) is unchanged and still true.

## Evidence

Measured 2026-09-02 in worktree `.claude/worktrees/t231`, branch `WT-clean-exit-signal`,
based on local develop `2660bcd09`.

| Check | Command | Result |
|---|---|---|
| Revised degraded-run tests | `go test ./internal/cli/worktree -run 'Clean' -count=1` | ok |
| Full touched package | `go test ./internal/cli/worktree -count=1` | `ok ... 3.357s` |
| Lint | `golangci-lint run ./internal/cli/worktree/...` | `0 issues.` |
| Cross-platform build | `GOOS=windows` + `GOOS=linux go build ./internal/cli/worktree/` | both exit 0 |
| ExitCoder seam (pre-existing, re-read) | `internal/cli/exitcode.go` `ResolveExitCode` + `cmd/moai/main.go:18` | no change needed |

The three revised tests (`clean_lock_unreadable_test.go`) assert the typed signal via
`errors.As(err, &eec)` → `eec.ExitCode() == 2` → cause text present, PLUS the unchanged
preservation half: zero removals, notice naming `cause=lock-source-unreadable`, JSON records
carrying `anchored: "undetermined"` + `keep_reason: cause=lock-source-unreadable`. The
normal-path clean tests (`clean_anchor_test.go`, `clean_base_branch_test.go`,
`clean_ignored_content_test.go`) assert `err == nil` on non-degraded runs and pass unchanged —
`cleanDegradedExit(nil) == nil`.

## Files

| File | Change |
|---|---|
| `internal/cli/worktree/clean.go` | `cleanDegradedExit` helper; `--merged-only` returns the signal; `--stale` three return sites + `reportStaleWorktrees` complete-then-signal |
| `internal/cli/worktree/guard.go` | `ExitCodeError.Detail` optional field (backward-compatible rendering) |
| `internal/cli/worktree/clean_lock_unreadable_test.go` | three tests revised nil→typed signal; `assertDegradedSignal` helper |
| `.moai/specs/SPEC-WORKTREE-REAPER-001/spec.md` | REQ-WR-016 revised; HISTORY v0.4.1 row; frontmatter 0.4.1 / 2026-09-02 |
| `CHANGELOG.md` | [Unreleased] Changed entry |

## Baseline-attribution

All measurements in this run, this tree (`WT-clean-exit-signal` @ develop `2660bcd09` + these
commits). Package-scoped per dispatch rule 5; no full-suite run locally — CI on the develop
push is the full-suite verdict.

## Gaps

- The process-level exit code 2 was not exercised end-to-end through the real binary: the
  unreadable-lock state is only reachable via the `gitWorktreeCmd` test seam (`git worktree
  list --porcelain` failing while the provider's `List()` succeeds is not constructible from
  outside the process). The chain ExitCodeError → ResolveExitCode → os.Exit is the
  pre-existing, production-proven `worktree verify` path, re-read and unchanged.
- Windows/Linux: cross-build only; tests not run there (CI covers the PR head).

## Residual-risk

- `--json` callers that previously treated exit 0 as the only success signal now see exit 2
  on degraded runs. That is the intended contract change (card/CR OD-2); any caller that
  breaks was silently consuming a degraded report as clean — exactly the failure this
  signal exists to surface.
