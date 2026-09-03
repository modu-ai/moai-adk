# t300 absorb re-measurement (develop e969dc07d)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `764960f51` (parents `f7662da98` = pre-absorb branch tip, `e969dc07d` = local develop tip incl. t191 landing)
- Measured tree: `6d594232d1d46cb1320b989ff256e9cc804b7f92` (`git rev-parse HEAD^{tree}` after absorb; evidence commit below adds only this file under `.moai/reports/t300/`)
- Working tree at merge: clean (0 rows `git status --porcelain`)

## Scope (derived from direct diff, not lead summary)

Absorbed side: packages of `git diff --name-only b7462203a e969dc07d` (develop gain since branch fork point)
union card delta `git diff --name-only e969dc07d..HEAD` packages.
Non-cli batch run here once for all lane-9 cards (absorbed tip identical across cards);
cli deferred to the cumulative stack-tree run (all lane-9 card cli deltas empty).

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/template/... ./internal/config/... ./internal/core/... ./internal/hook/... ./internal/lockfile/... ./internal/settings/... ./internal/spec/... ./internal/statusline/... ./internal/web/...` → exit 1 with exactly the 4 known-owner reds below; all other packages ok
- t300's own guard (`internal/template/vci_ordering_clause_guard_test.go`) → PASS (not among failing tests)

## Observed reds — all match lead's canonical exclusion list, none repaired

| Test | Observation | Owner |
|---|---|---|
| TestCatalogHashCoversSkillSubfiles | CATALOG_HASH_SKINNY moai stored=3fff7dba… whole-tree=f005e873… | t436 (lane-5) |
| TestManifestHashFormat | CATALOG_HASH_UNSTABLE moai + sync-auditor.md | t443 (4244c4a06 family) |
| TestGoldenCommittedArtifactsMatchEmission | sync-auditor.toml sha256 mismatch | t443 (same root) |
| TestCatalogHashParity | DRIFT moai + sync-auditor entries | t443 / t436 (same two entries) |
| TestAlwaysLoadedTokenBudget | 77,267 / 76,400, overflow 867, 17 entries | t453 (lane-9, this batch) |

## t453 budget observation (for verdict)

Measured on THIS absorbed tree: always-loaded surface 77,267 — already above t453's proposed
77,200 (this card's own VCI rules growth is included here). Re-measure on the merged tree at
window call time; report to lead; no arbitrary raise (lead order 2026-09-03).

## Excluded / not run

- `./internal/cli/...` — deferred to stack-tree solo run (expected ~788s, lane-11 measurement)
- go vet / golangci-lint — CI surface; absorb authored no new code beyond git's automatic merge
