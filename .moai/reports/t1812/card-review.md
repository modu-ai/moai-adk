# t1812 card review (sync gate)

Tree: `.moai/worktrees/t1812`, branch `WT-codex-gate-tmpdir-symlink`, HEAD `6738cf9fb`, review base `main` (merge base `3a8689d6b`).

## Claim

The three macOS-only failures (`TestCodexReviewGatePrimaryCheckoutSkip`, `TestCodexReviewGate_TreeScopeRequestShapeUnchanged`, `TestCodexReviewGate_StaleEnvKeepsTreeScope`) were a test-expectation defect (unresolved `t.TempDir()` symlink spelling), now repaired; the production path resolution is unchanged and intentional.

## Evidence

- Audit plan (`moai verify audit-plan --project-root <tree>`): codex = `required`, claude/glm = `off`; `enforced_required: ["codex"]`.
- Required backend run — `mcp__moai__codex_audit` (mode native, target baseBranch, project_root = this tree):
  `verdict: pass`, `findings: []`, `audit_receipt: rcpt-88a18cfb542ac23cb1d660c1`, `review_base: main (merge base 3a8689d6b…)`.
  The backend states it reviewed statically and did not execute tests, builds, or repro commands.
- Executed by this lane (this run): the 3 target tests `ok` after the fix and FAIL before it (progress.md §1, §3); `golangci-lint run --new-from-rev=HEAD ./internal/cli/...` → `0 issues.`; `go test ./internal/template -count=1` → `ok 243.716s`.
- Rule-C scope: `go test ./internal/cli -run 'Mirror|Parity|Doc'` → FAIL on 7 tests unrelated to the diff; the same 7 fail with the fix stashed (A/B, progress.md §4).

## Baseline-attribution

Commands above were run in this tree at `6738cf9fb` (tests at base+edits before the commit; the commit adds only the edits and the evidence file). Judging build for the codex result: the audit tool reported `build_commit 3a8689d6b`, a strict ancestor of HEAD; the diff since it is test-only, so the lag does not change the audited code path.

## Gaps

- Full `./internal/cli` run (with the SessionWorktree `-skip`) not executed; the go-test-heavy slot for lane-38 is last in the queue and the change is test-only.
- Linux/Windows not run (identity under `EvalSymlinks` there).
- The 7 unrelated sync-scope failures are not diagnosed — a separate card is needed (Codex mirror doctor ×6, `TestHookEntryParity_NoDivergenceEitherDirection`).
- codex did not execute anything; its pass is a static read.

## Residual-risk

- `~/.gitconfig` stood at 484 lines before and after every run here (baseline cited by the leader: 482). The +2 predates this card's runs; cause unknown to this lane.
- The 7 pre-existing failures would turn a naive "internal/cli green" claim false on this base; the leader should not read this card as restoring that.

## Card-scope self-review (`codex_review`, scope=card — the stage the doctrine names)

Call: `mcp__moai__codex_review` scope=card, project_root = this tree, run against HEAD `fba1b65ac`.
Result (verbatim fields): `verdict: pass`, `findings: []`, `advisory: true`, `backend: codex`, `base: 3a8689d6b48a4e7ad1f33f9081878473a153cbe8`, `truncated: false`, `excluded_untracked: []`.
Summary: static review found no defect; the change normalizes the tests' expected path to the real path and keeps the existing assertions; no tests, builds, or repro commands were executed by the reviewer.
Disposition of findings: none to disposition. This result is advisory and carries no receipt id; the binding evidence for the required codex gate is the `codex_audit` receipt `rcpt-88a18cfb542ac23cb1d660c1` above.
