# t1862 — progress record (Class B defect, cause known)

Tree: `.moai/worktrees/t1862`, branch `WT-factory-cardroot-align` (renamed in place from the
lease-time `WT-10-18-t1827`), base `2178966bb` (== local main at lease).
Scope: internal/cli (`factory_handoff_recover.go`) + the t1827 SPEC-wording reversion after
t1827 lands. Implementation is HELD for t1827 (lane-29, `WT-db-paths-where`) landing per the
leader dispatch of 10-18; this record currently carries cause evidence + RED only.

## 1. Cause evidence (mechanism + code sites)

**Mechanism.** Both verbs open the factory at the raw working directory:

- `internal/cli/factory_handoff_recover.go:29-36` (`factory handoff recover-resume`):
  `root, err := os.Getwd()` → `homestate.OpenFactory(root)`, and the same raw `root` feeds
  `homestate.ReadRuntimeCensus(root)`.
- `internal/cli/factory_handoff_recover.go:86-95` (`factory runs`): the same `os.Getwd()` root
  feeds `homestate.OpenFactory(root)` and `factorymsg.ReconcileOptionsFor(root)`.

The two rules every other factory surface applies:

- moai-paths root: `resolveProjectDir()` (internal/cli/session.go:315-323) —
  `CLAUDE_PROJECT_DIR` first (`env:CLAUDE_PROJECT_DIR`), then the cwd (`server-cwd`).
- factory-root rule (SPEC-FACTORY-LEASE-RENEW-001, sync-audit F2, card t1663;
  internal/cli/factory_card.go:86-92): "every surface that opens the factory for a launch tree
  applies this rule" — `factoryRootOf(launch) = factory.ResolveTodoQueueRoot(launch)`, which
  maps a linked worktree to the primary checkout.
- `factoryCardRoot()` (internal/cli/factory_card.go:84) composes both; `factory next`, `stage`,
  `renew`, `complete`, `decide`, `bundle` all route through it.

Neither verb applies either rule, so a cwd naming a different repository than moai paths (a
different `CLAUDE_PROJECT_DIR`, or a linked worktree) reads that other repository's factory
DB. `OpenFactory` additionally "reads the lease keys from the root the factory is opened at"
(internal/homestate/factory.go:171-190), so lease state is judged on the same wrong root.

t1827's unmerged branch (`WT-db-paths-where`) documents the divergence in `moai paths` help
("moai factory runs and moai factory handoff recover-resume fix the root from the working
directory instead — SPEC-MOAI-DB-PATHS-001") and narrowed that SPEC's wording to declare the
divergence pre-existing; this card reverts that narrowing once t1827 lands (the divergence is
then gone). lane-29 reproduced the two-repository read during the t1827 sync audit.

## 2. RED — verbatim, before the fix

Reproduction: `internal/cli/factory_root_alignment_test.go` (new file, this commit). Boundary
as the card names it: cwd names project B while `CLAUDE_PROJECT_DIR` names project A; both
verbs must read A's factory, never B's.

Command: `go test -run 'TestFactoryRunsOpensMoaiPathsRootNotCwd|TestRecoverResumeOpensMoaiPathsRootNotCwd' -count=1 -timeout 8m ./internal/cli/`

```
--- FAIL: TestFactoryRunsOpensMoaiPathsRootNotCwd (0.58s)
    factory_root_alignment_test.go:56: factory runs read the cwd factory, not the moai-paths root (card t1862): output "RUN ID  STATUS  OWNER  PID  CREATED  UPDATED\n"
Error: legacy resume recovery CAS precondition: sql: no rows in result set
--- FAIL: TestRecoverResumeOpensMoaiPathsRootNotCwd (0.48s)
    factory_root_alignment_test.go:85: recover-resume read the cwd factory, not the moai-paths root (card t1862): legacy resume recovery CAS precondition: sql: no rows in result set
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.408s
```

Failure modes are exactly the defect: `runs` lists the cwd (B) factory — empty, header only —
and `recover-resume` finds no row because B's factory never held A's claim.

## 3. Plan (held)

1. After t1827 lands on local main: absorb main into this branch, keep the RED compiling
   against the t1827 tree (it changed `factory_handoff_recover.go` comments only), implement
   the alignment (both verbs open `factoryCardRoot()`), RED → GREEN.
2. Revert the t1827 SPEC wording narrowing (SPEC-MOAI-DB-PATHS-001) to the original meaning
   and drop the `moai paths` help sentence that documents the cwd divergence.
3. Scoped verification (`-run` narrowed, rule 35e9), card-review, sync audit, landing.

## Hold (explicit wait)

- Reason: implementation ordered after t1827 (lane-29) lands; the fix and the SPEC-wording
  reversion both build on t1827's tree.
- Whom: leader tn1p4e dispatch 10-18 ("구현은 t1827 착지 뒤에").
- Recheck point (on every wake): `git log main --oneline --grep='card t1827' -n 1` non-empty
  AND `moai factory status`-observable landing, then absorb and implement.
