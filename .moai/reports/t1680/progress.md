# t1680 progress record

Card t1680 (Class B, run tn1p4e, lane-27, mode parallelizable). Tree `.moai/worktrees/t1680`, branch `WT-mcp-verb-parity` (renamed in place from the leased `WT-t1663-sync-t1663`), base local main `ae34f26831` (HEAD equals main at lease time, tree clean).

## Cause evidence (read from code in this run, not yet reproduced; the RED tests below reproduce each one)

Baseline: `internal/cli/mcp_factory_card.go` and `internal/cli/factory_card.go` at `ae34f26831`.

| Item | Verb (reference rule) | MCP handler (divergent) |
|---|---|---|
| N1 lane label | `factoryLaneLabelFromEnv(verb)` at `factory_card.go:1697` (next), `:1832` (stage), `:2038` (complete); it refuses legacy labels (`worker-1`) before any work (`:148-164`) | `handleFactoryNext` (`mcp_factory_card.go:132`) and `handleFactoryComplete` (`:235`) read `MOAI_FACTORY_WORKER` directly; `handleFactoryStage` passes `os.Getenv(...)` inline at `:219`; only `handleFactoryRenew` (`:299`) calls the shared check |
| (a) decide root | `factoryDecideCards(ctx, factoryCardRoot(), ...)` at `factory_card.go:3289`, `factoryCardRoot() = factoryRootOf(resolveProjectDir())` (`:84`) | `handleFactoryDecide` uses `resolveToolProjectRoot(req)` (raw `project_root`) at `mcp_factory_card.go:257` and passes it at `:278` |
| (b) complete window root | `factoryCompleteCard(ctx, out, factoryCardRoot(), integrationLockRoot(), ...)` at `factory_card.go:2050` | `handleFactoryComplete` passes `launch` (the caller's raw tree) as the window root at `mcp_factory_card.go:247` |
| N3 / F8 / (c) | shared run resolution `resolveFactoryCardRun` (`factory_card.go:2646`) and the renew reconcile path | not yet examined in this block |

Already aligned by t1663 (do not redo): `handleFactoryRenew` (label check and `factoryRootOf(launch)`), `handleFactoryNext` / `handleFactoryStage` / `handleFactoryComplete` use `factoryRootOf(launch)` for the record root.

Debt definitions (from `.moai/specs/SPEC-FACTORY-LEASE-RENEW-001/progress.md` lines 2811-2813 and 3742-3744): N1 as above; N3 = the other lane verbs that resolve a run through `resolveFactoryCardRun` still refuse when several runs are active and no `--run` is given, only the renewal path was narrowed; F8 = a refused renewal still retires a dead run through reconcile and writes an event.

Scope guard from the leader: touch the MCP handler side only; t1688 (lane-26) changes the `complete` Args in `factory_card.go` and t1839 (lane-17) changes the `next` selection, so any change to shared code in `factory_card.go` needs an explicit reason and an absorb-conflict check at landing. No test skips, no timeout enlargement, `internal/cli` only by `-run` selection (ruling 35e9).
