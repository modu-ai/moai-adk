# Verdict — card t1589

- Card: t1589 (Class B — defect, cause unknown; run → sync)
- Run: tmvgvn · Lane: lane-27 · Stage: run→sync (lane-local, no lease)
- Branch: WT-factory-card-pr-git-dir · Worktree: .moai/worktrees/t1589
- Commits: 6b12d5b830 (RED test) → 4faf5ff818 (fix) · Base: ec193e4777 (local main)
- Tree measured: 4faf5ff818 · Judging build: moai v3.2.0-rc.31 (ec193e4777) for `moai factory status` only; tests built from the tree under test (`go test ./internal/cli/`).
- Remote push: none (merged-local + SHA only, per leader dispatch).

## Claim

The delivery edge's git reads (`factoryGitRead`, `factoryBranchOfWorktree`,
`factoryWorktreeForBranchIn`) act on the repository named by the worktree
argument even when an inherited `GIT_DIR` / `GIT_WORK_TREE` names another repository.

## Evidence

RED — command: `go test ./internal/cli/ -run 'TestFactoryDeliveryGitReadsIgnoreInheritedGitEnv' -count=1 -v` (test file committed alone at 6b12d5b830, before the fix):

```
--- FAIL: TestFactoryDeliveryGitReadsIgnoreInheritedGitEnv (4.78s)
    factory_card_pr_gitenv_test.go:25: factoryGitRead(log) = "unrelated repository tip", want the card worktree's tip "card worktree tip" (the inherited GIT_DIR answered for "…/002")
    factory_card_pr_gitenv_test.go:29: factoryBranchOfWorktree() = "other-branch", want the card worktree's branch "WT-card-branch" (the inherited GIT_DIR answered for "…/002")
FAIL
```

RED (third assertion, added before its fix) — command: same `go test … -count=1 -v`:

```
    factory_card_pr_gitenv_test.go:33: factoryWorktreeForBranchIn() = "", want the card worktree "…/001" (the inherited GIT_DIR listed "…/002")
--- FAIL: TestFactoryDeliveryGitReadsIgnoreInheritedGitEnv (3.47s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	5.054s
```

GREEN — command: `go test ./internal/cli/ -run 'TestFactoryDeliveryGitReads|TestFactoryComplete|TestMergeReadiness|TestFactoryGitHubFlow|TestIntegrationMergeWorktree|TestIntegrationMergeAccepts|TestLocalMainResync' -count=1 -v` (at 4faf5ff818):

```
--- PASS: … (40 tests, names in progress.md)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	225.964s
exit=0
```

Format — command: `gofmt -l internal/cli/factory_card.go internal/cli/factory_card_pr.go internal/cli/factory_card_pr_gitenv_test.go` → empty output (clean).

## Baseline-attribution

- RED baseline: the same test at 6b12d5b830 (no fix), observed in this run: all three assertions fail with the other repository's answer.
- GREEN baseline: the 40-test set at 4faf5ff818, observed in this run, `exit=0`.
- The fix is the same mechanism the repository already uses on sibling sites: `gitenv.Env()` (`internal/gitenv/gitenv.go`, the `init_gitdetect.go` / `integration_candidate.go` / `integration.go` call sites).

## Gaps

- **Lease refused (§3.1 refusal → Gap):** `moai factory next` was run 3 times, each returning `no card is available`; `moai factory status --run tmvgvn` read `t1589 … owner=lane-27 lease=none`. The work proceeded without a lease, as the leader instructed ("임대 없음(고부하, 리더 지시)"). This is a refusal, not a silent substitution.
- **Second defect in the card body not addressed:** the card body also names a strategy-branch mismatch (`git-strategy.yaml` `workflow=git-flow` vs the operated github-flow; `t1583` complete taking the local-develop merge path). Not repaired here — the leader's dispatch scoped this lane to the GIT_DIR cause. Left for leader decision.
- **Other raw git sites on the same class, not on the delivery path, not changed:** `internal/cli/factory_card.go:633` (`git -C <wt> branch -m` during card-worktree creation) and `internal/cli/factory_card.go:2272` (`cliGitIn`, adoption probes). Named as residual risk.
- **Not run:** the full `internal/cli` suite (`go test ./...` forbidden by dispatch), `go vet` and `golangci-lint` on the package, and any CI. The lint diagnostics the editor reports on the two files are pre-existing lines not touched by this change.
- **No remote evidence:** no push, no PR, no CodeRabbit (remote push forbidden by dispatch).

## Residual-risk

- Two `git` calls outside the delivery edge still take an inherited `GIT_DIR` (see Gaps). They fire only when a `GIT_DIR` is set in the environment, typically inside a git hook.
- The new test sets `GIT_DIR` / `GIT_WORK_TREE` with `t.Setenv`; `TestMain` in `internal/cli` scrubs the process environment at startup (`gitenv.ScrubProcess`), so the test's own fixtures are not affected by a caller's `GIT_DIR`.
