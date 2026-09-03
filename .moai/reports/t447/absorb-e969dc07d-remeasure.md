# t447 absorb re-measurement (develop e969dc07d, landed t191)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `ad04edbb2` (parents `9498142c8` pre-absorb tip, `e969dc07d` local develop tip).
  One conflict, CHANGELOG.md [Unreleased]: HEAD side held t295 + t191's PRE-landing entry;
  develop side held t191's LANDED entry + t280 + t295 + t239 — a superset. Resolved to the
  develop side (`git checkout --theirs`), which retires the stale pre-landing t191 entry copy.
- Measured tree: `e3105c66c309e1c19876ba5023b271d7d1583783` (evidence commit adds only this file)

## t191 ancestor premise (for window call)

Pre-absorb branch contained t191 old tip `55885aae3` (verified by merge-base, linear with
lane-4's current tip). This absorb brought the LANDED t191 via develop `e969dc07d`, so the
branch now carries the landed version; the window-time check remains
`git log --grep "card t191"` on local develop before merging t447.

## Scope (direct diff)

Card delta: `.claude/skills/moai/workflows/project/doc-generation.md` (local + template
mirror) + `internal/template/project_continuation_pipeline_signal_test.go` + evidence.
NOTE: absorbed develop ALSO edited doc-generation.md (t191 Phase-14 contract, 48 lines);
the auto-merge is exercised by the card's own guard test, which reads the merged doc.

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/template/... ./internal/spec/...` → only the canonical
  exclusion reds: TestCatalogHashCoversSkillSubfiles + TestManifestHashFormat
  (moai entry — computed whole-tree hash CHANGED to f01063f0… because this card's
  doc-generation.md edit is inside the .claude/skills/moai/ tree; t436 owns the refresh)
  + TestGoldenCommittedArtifactsMatchEmission + TestCatalogHashParity
  (sync-auditor, t443). project_continuation_pipeline_signal_test PASS — the
  auto-merged doc satisfies the pipeline-signal guard.

## Window-time follow-up owed by this landing

Merging t447 re-stales the catalog `moai` whole-tree entry AFTER t436's refresh lands
(lane-5 is ahead of lane-9 in window order). Scoped remedy inside my window: after the
last card merge, run `gen-catalog-hashes.go --entry moai` on the final develop state so
the tree I leave behind is green on that entry. No other entry is affected by this card.

## Excluded / not run

- cli suite — deferred to stack-tree solo run (card cli delta empty)
- go vet / lint — CI surface
