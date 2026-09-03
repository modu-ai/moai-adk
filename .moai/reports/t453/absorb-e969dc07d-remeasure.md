# t453 absorb re-measurement (develop e969dc07d)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `0606440f2` (parents `4abe30a8d` pre-absorb tip, `e969dc07d` local develop tip).
  Both sides edited `internal/config/defaults.go` (card: budget raise; develop/t191: new
  inbox constants) in disjoint hunks — auto-merge clean, config suite verdict below.
- Measured tree: `1c49bbcae5f4d4529aa480950788b716799adf08` (evidence commit adds only this file)

## Card delta (direct diff)

`internal/config/token_budget_guard.go` (AlwaysLoadedTokenBudget 76,400 → 77,200 with
per-clause attribution) + evidence files.

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/config/...` → ALL GREEN, including TestAlwaysLoadedTokenBudget:
  on this tree the always-loaded surface (~76,939, develop e969 level) is under the raised
  budget 77,200. The fix is live and effective here.

## Landing-time projection (for lead decision — NOT raised unilaterally)

Measured anchors (all attributable to their trees):
- lead's measurement on develop `3bdd5a803` (t458 included): surface 77,104 → headroom 96
- lane-5 t436 contribution when it lands before lane-9: +81 → headroom 15
- lane-9 cards' own always-loaded growth (measured on t300's absorbed tree:
  77,267 at budget 76,400, i.e. ~+328 over develop e969's 76,939 — t300's VCI edits +
  t302/t345's NEW `verification-completeness.md` rules file): the stack-tree run will
  measure the exact combined number.
Projected landing-time surface ≈ 77,104 + 81 + (stack-measured card growth) — likely
ABOVE 77,200. Final numbers reported to lead after the stack-tree run; the raise stays
77,200 until the lead decides on the measured projection.

## Excluded / not run

- cli suite — deferred to stack-tree solo run
- go vet / lint — CI surface
