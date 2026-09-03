# t353 absorb re-measurement (develop e969dc07d)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `53c1b80b3` (parents `105a41387` pre-absorb tip, `e969dc07d` local develop tip)
- Measured tree: `9f7bbbf91023234d850ee192fbd12dc7d76df9a6` (evidence commit below adds only this file)
- Working tree at merge: clean (0 rows)

## Scope (direct diff)

Card delta: `internal/spec/lint_movingref.go` + `lint_movingref_test.go` + testdata
`movingref/` (4 fixtures), `SPEC-MOVING-REF-GUARD-001` spec + acceptance, evidence.
All Go surface in `internal/spec`.

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/spec/...` → exit 1 with ONLY TestCatalogHashParity
  (the same two absorbed-side entries: moai = t436 owner, sync-auditor = t443 family).
  lint_movingref suite PASS; all other spec tests PASS.

## Excluded / not run

- cli suite — deferred to stack-tree solo run (card cli delta empty)
- go vet / lint — CI surface
