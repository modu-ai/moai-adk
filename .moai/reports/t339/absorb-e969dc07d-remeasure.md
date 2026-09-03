# t339 absorb re-measurement (develop e969dc07d)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `b3aa32f34` (parents `fc0d04ef9` pre-absorb tip, `e969dc07d` local develop tip)
- Measured tree: `5a15c89ba81f51c0a4f2d0b2fce8f25d4a550647` (evidence commit below adds only this file)
- Working tree at merge: clean (0 rows)

## Scope (direct diff)

Card delta: SPEC documents only (`SPEC-AGENT-EMIT-LINEAGE-001` plan.md + spec.md) + evidence.
No Go / no templates / no agents. Judgment surface: the absorbed spec linters over repo SPECs.

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/spec/...` → exit 1 with ONLY TestCatalogHashParity
  (the two canonical absorbed-side entries: moai = t436, sync-auditor = t443).
  SPEC lint suites PASS — t339's document edits satisfy the absorbed linters.

## Excluded / not run

- cli suite — deferred to stack-tree solo run (card cli delta empty)
- go vet / lint — CI surface
