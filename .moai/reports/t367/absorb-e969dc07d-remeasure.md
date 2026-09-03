# t367 absorb re-measurement (develop e969dc07d)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `18612617a` (parents `18fc2c9ef` pre-absorb tip, `e969dc07d` local develop tip)
- Measured tree: re-measured AFTER the scoped repair below; final repair commit `2549f775f`
- Working tree at merge: clean (0 rows)

## Scope (direct diff)

Card delta: `.claude/agents/moai/plan-auditor.md` + template mirror + evidence.
Judgment surface: catalog parity + C3 golden (agent file changed).

## Findings — 2 card-owned defects found and repaired

The absorb brought t441's expanded catalog coverage; t367 predates it, so its close-time
runs never checked these:

1. `plan-auditor` catalog.yaml entry stale (stored 6a50b054 ≠ computed efb7167d)
2. `.codex/agents/moai/plan-auditor.toml` golden mismatch (md changed, C3 not regenerated)

Repair (scoped to t367 files): `gen-catalog-hashes.go --entry plan-auditor` +
`make agents-emit`, then `git restore` of `sync-auditor.toml` (t443/4244c4a06-owned,
deliberately not repaired). Commit `2549f775f`.

## Post-repair results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/template/... ./internal/spec/...` → only the 4 canonical
  exclusion reds remain: TestCatalogHashCoversSkillSubfiles (moai, t436),
  TestManifestHashFormat + TestGoldenCommittedArtifactsMatchEmission + TestCatalogHashParity
  (sync-auditor + moai entries, t443/t436). plan-auditor absent from all lists.

## Excluded / not run

- cli suite — deferred to stack-tree solo run (card cli delta empty)
- go vet / lint — CI surface
