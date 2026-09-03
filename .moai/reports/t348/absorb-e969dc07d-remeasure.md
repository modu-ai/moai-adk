# t348 absorb re-measurement (develop e969dc07d)

Lane-9 window-readiness absorb, outside the integration window. 2026-09-03.

## Baseline attribution

- Absorb merge: `45c9f1065` (parents `3725072c5` pre-absorb tip, `e969dc07d` local develop tip)
- Measured tree: `63aed5e1a294971ec928426b0f43eb79b4718a4c` (evidence commit below adds only this file)
- Working tree at merge: clean (0 rows)

## Scope (direct diff)

Card delta: manager-docs agent triple (`.claude/agents/moai/manager-docs.md` + template
mirror + `.codex/agents/moai/manager-docs.toml` emission), `internal/spec/ac_count_clause_test.go`
+ testdata, `internal/template/catalog.yaml` entries, evidence.
NOTE: `internal/template/catalog.yaml` was AUTO-MERGED during absorb (both sides modified) —
semantic risk on hash content is exactly why the catalog tests were re-run on this tree.

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/template/... ./internal/spec/...` → exit 1 with exactly the
  4 known-owner reds (identical entries to the other trees): TestCatalogHashCoversSkillSubfiles
  (t436), TestManifestHashFormat + TestGoldenCommittedArtifactsMatchEmission +
  TestCatalogHashParity (t443 family, sync-auditor + moai entries only).
  Card-specific verdicts from the same runs:
  - t348's own catalog entries (manager-docs) — hashes match computed (NOT in drift list)
  - t348's manager-docs.toml golden pair — consistent (only sync-auditor.toml flagged)
  - ac_count_clause_test.go — PASS

## Excluded / not run

- cli suite — deferred to stack-tree solo run (card cli delta empty)
- go vet / lint — CI surface; absorb authored no new Go code beyond the card's own test
