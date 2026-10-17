# t1812 sync-audit verdict

Written by lane-38 (the lane's own completion verdict on mechanical evidence, per `factory-dispatch-detail.md` § The verdict's home). This is NOT a `sync-auditor` agent run; the independent evidence beneath it is the required-backend codex receipt below.

verdict: PASS
audited_sha: 12ac4803d8166c24ab4db5fd7a1e62892c2f0a59
convergence_overall: pass
required_backend: codex pass

## Claim

Card t1812 repaired three macOS-only test failures with a test-side change only; no production file changed, and nothing in the repaired tests regressed.

## Evidence

- Required backend (`moai verify audit-plan --project-root <tree>`: codex = required, claude/glm = off): `mcp__moai__codex_audit` native / baseBranch → `verdict: pass`, `findings: []`, receipt `rcpt-88a18cfb542ac23cb1d660c1`, review base `main` (merge base `3a8689d6b`).
- Card-scope review (`mcp__moai__codex_review`, scope=card) → pass, findings none, advisory. Detail: `.moai/reports/t1812/card-review.md`.
- Cause determination (before repair) and the before/after runs: `.moai/reports/t1812/progress.md` §1, §3.
- Diff since base is three files: two `internal/cli/*_test.go` edits plus the evidence files under `.moai/reports/t1812/`.

## Baseline-attribution

Audited commit `12ac4803d` is the HEAD of branch `WT-codex-gate-tmpdir-symlink` in `.moai/worktrees/t1812`; tests and lint were run in that tree (see progress.md and card-review.md for the commands and outputs). The codex audit tool reported its judging build as `3a8689d6b`, a strict ancestor of HEAD; the commits after it are test-only and evidence files, so the lag does not change the audited code path.

## Gaps

- The full `./internal/cli` suite (SessionWorktree `-skip`) was not run: the go-test-heavy slot turn for lane-38 comes last and the change is test-only.
- codex judged statically; it executed no tests.
- `go test ./internal/cli -run 'Mirror|Parity|Doc'` fails on 7 tests; the same 7 fail without this card's edits (A/B in progress.md §4), so they are carried, not caused, by this card. Their cause is not diagnosed here.
- Not run on Linux/Windows.

## Residual-risk

- `./internal/cli` is not green on this base because of the 7 carried failures; anyone reading this PASS as "internal/cli is green" would be wrong.
- `~/.gitconfig` stands at 484 lines against the leader's 482 baseline, unchanged by this lane's runs; cause not investigated.
