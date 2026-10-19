# t1823 card review (sync gate)

Tree: `.moai/worktrees/t1823`, branch `WT-main-red-managed-tui-binlag`, audited HEAD `1a82b59917e2f17a0b58196455bee0982b31de14`, base `ec5a4f392`.

## Claim

Two red tests on main are repaired: (2) `TestAuditLagUsesBinlagSeam` was coordinate drift only (t1619 and t1589 moved the hits); (1) `TestManagedTUINeverReachedWithoutOptIn/opted_in_probes_once` was a production-path regression from `063a7620f` (the launcher injects `--add-dir <moai home>` and the managed divert handed it to an owner that refuses it), repaired in the managed divert only.

## Authorship (public record)

- `54944481c`: LANE direct edit, scope: test coordinate map only (`internal/cli/mcp_build_identity_test.go`). No `manager-develop` spawn. The leader ruled it need not be redone; it is inside the audit scope below.
- `f37be81f7`: `manager-develop` spawn — `internal/cli/codex_launcher.go` (+5/-1) and two new tests in `internal/cli/codex_moai_home_add_dir_test.go` (+52). RED on the unfixed code was reported by that spawn; the lane did not re-observe RED.
- The lane first edited `codex_launcher.go` itself, then reverted that edit before any commit (progress.md §2).

## Evidence

- Audit plan (`moai verify audit-plan --project-root <tree>`): codex = required, claude/glm = off, `enforced_required: ["codex"]`.
- Required backend: `mcp__moai__codex_audit` native / baseBranch, HEAD `1a82b5991` → `verdict: pass`, `findings: []`, `audit_receipt: rcpt-c285169dc42467aba2878c41`, review base `main (merge base ec5a4f392…)`. Static review; it executed no tests.
- Card scope: `mcp__moai__codex_review` scope=card → `verdict: pass`, `findings: []`, `advisory: true`, base `ec5a4f39202f3c8b404301864b90596e6af2e16f`.
- Test-executing sync auditor (`sync-auditor` spawn, report `.moai/reports/t1823/auditor-run-report.md`): per-claim C1–C5 PASS, `auditor_verdict: PASS`, no blocking findings. Its commands: the four target tests `ok 2.660s`; `go test ./internal/cli -run 'AddDir|ManagedLaunch|ManagedTUI|ManagedCodex|CodexLaunch|CodexDirect|RunCodex'` `ok 43.903s`; `go vet` and `gofmt -l` no output; `git diff --stat ec5a4f392..HEAD` = 4 paths (progress.md, codex_launcher.go, codex_moai_home_add_dir_test.go, mcp_build_identity_test.go).
- Lane (this tree, final state): `golangci-lint run --new-from-rev=ec5a4f392 --timeout=10m ./internal/cli/...` → `0 issues.` exit 0; the four target tests all PASS (`ok 2.759s`); `go vet` / `gofmt -l` no output.
- A/B: `063a7620f^` (`2e387c789`) → `TestManagedTUINeverReachedWithoutOptIn` PASS (4 subtests, `ok 1.783s`); base `ec5a4f392` → `opted_in_probes_once` FAIL (`probe=0 … want 1/0/0`, launch error `managed codex launch does not support --add-dir`). progress.md §1, §4.

## Finding dispositions

- F1 (auditor, low, optional): the managed test asserts `--add-dir` absence and `--model` presence but not that the developer-instructions `-c` pair rides through. Not adopted: the code path for that pair is shared and its semantics unchanged; a pin for it would be new scope.
- F2 (auditor, low, optional): the coordinate-keyed want map is brittle by construction (third re-measurement). Carried over, out of scope for this card.

## Baseline-attribution

All lane and auditor commands were run in `.moai/worktrees/t1823` at the HEADs named above, compiled by `go test` from that tree. The MCP audit tools judged with a build at `3a8689d6b`, a strict ancestor of HEAD (the tool reported the lag); the diff since is limited to the paths above, none of which the audit tools' own code depends on.

## Gaps

- Full `./internal/cli` package and `go test ./...` not run (go-test-heavy slot; the t1812 remeasure holds priority).
- Managed divert not run against a real codex binary; whether the managed App Server needs the moai home writable by another route is not established.
- The lane did not re-observe RED for the new test; and 063a7620f itself was not measured as a separate point.
- The sync-auditor's final message cited a second receipt (`rcpt-d4f3ea8c751186f194f91e5a`) while its report file says no `audit_multi` was run; that receipt was not verified by the lane and is not relied on.
- Backlog.db hash and `~/.moai/db` count were not measured around the narrow runs (they belong to the heavy remeasure).

## Residual-risk

- `~/.gitconfig` was 501 lines before and after this card's narrow runs; its growth from 489 (t1812 time) predates them, source not investigated.
- The managed headless session gets no `--add-dir`; whether it can write `~/.moai` databases is untested here (pre-`063a7620f` behavior restored).
- The want map will rot again on the next edit above its coordinates (F2).
