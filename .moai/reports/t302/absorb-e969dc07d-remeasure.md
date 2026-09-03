# t302 absorb re-measurement (develop e969dc07d)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `42083eece` (parents `77062dda8` pre-absorb tip, `e969dc07d` local develop tip)
- Measured tree: `2bfe2b0a74ee27f72ca6d9f9c9acaa82e2cfb77a` (evidence commit below adds only this file)
- Working tree at merge: clean (0 rows)

## Scope (direct diff)

Card delta `git diff --name-only e969dc07d..HEAD`: rules doc
`verification-completeness.md` + template mirror, `sync-audit-4dim.js` + template mirror,
evidence files. No Go files. Judgment surface = catalog/embed consumers
(internal/template, internal/spec). Build covers the rest.

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/template/... ./internal/spec/...` → exit 1 with exactly the
  4 known-owner reds (identical hashes to the t300-tree run → drift is absorbed-side):
  TestCatalogHashCoversSkillSubfiles (t436), TestManifestHashFormat +
  TestGoldenCommittedArtifactsMatchEmission + TestCatalogHashParity (t443 family).
  t302's own new mirror files introduce no new red.

## Excluded / not run

- cli suite — deferred to stack-tree solo run (card cli delta empty)
- go vet / lint — CI surface; no new Go code in this absorb
