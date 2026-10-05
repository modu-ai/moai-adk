---
id: SPEC-STATUSLINE-LANDED-LABEL-001
title: "Acceptance criteria — statusline landed annotation"
---

# Acceptance — SPEC-STATUSLINE-LANDED-LABEL-001 (card t1281)

All Go-test criteria live in `internal/statusline` and run under `go test ./internal/statusline/... ./internal/kanban/...`. Fake git output uses the scan shape `<sha>\x00<unix-ct>\x00<subject>` per line. Picked cards are seeded with an explicit `added_at`.

## §D AC Matrix

| AC | REQ | Given | When | Then |
|---|---|---|---|---|
| AC-SLL-001 | REQ-SLL-002 | picked card `t101`; the fake git returns one commit (committer time after `added_at`) whose subject NAMES `t101` in a non-attributing position: `chore: follow-up review notes for t101` | refresh runs | new criterion: cache `landed == 0`, `measured == true`. Discrimination control: the retired body-mention criterion (`\bt101\b` over the same text, i.e. old `countNamed`) yields `1` on this input — the test states both expectations. The body-only clause (an id visible only to a `%B` query) is pinned by AC-SLL-004 (argv equality) and AC-SLL-012 (`%B` grep), not by this row |
| AC-SLL-002 | REQ-SLL-001 | picked card `t101` added `2026-01-02T03:04:05Z`; subject `feat(statusline): x (t101)` with committer time after `added_at` | refresh runs | cache `landed == 1`, `measured == true` |
| AC-SLL-003 | REQ-SLL-003 | picked card `t101` added `2026-01-02T03:04:05Z`; the only attributing subject has committer time one day earlier | refresh runs | cache `landed == 0` |
| AC-SLL-003b | REQ-SLL-003 | picked card `t101` with empty `added_at`; attributing subject present | refresh runs | cache `landed == 0` (fails closed) |
| AC-SLL-004 | REQ-SLL-004 | 1 picked card, then 50 picked cards (separate runs) | refresh runs with a counting `landedGitRunner` | exactly 1 `git log` subject-stream query through `landedGitRunner` in each run (the count does not grow with the card set), and its argv equals `kanban.LandedScanArgs(kanban.LandedRefFor(root))`. Ref resolution (`kanban.LandedRefFor`, which may run one read-only `symbolic-ref` through kanban's own seam) is unchanged and not counted here |
| AC-SLL-004b | REQ-SLL-006 | picked card `t101` with an attributing subject | refresh runs successfully | the written cache JSON carries the current criterion identifier (a non-empty field equal to the package constant) |
| AC-SLL-005 | REQ-SLL-007 | a cache file written in the old schema — `{"landed":9,"ref":"origin/develop","measured":true,"fetched_at":<now>}` with no criterion field | the session line renders | `Known() == false`, the line contains neither `⚑` nor `✓`, and the TODO segment equals the unannotated `🔄 TODO: p/q` |
| AC-SLL-005b | REQ-SLL-007 | the same old-schema cache with a fresh `fetched_at` | `maybeRefreshLandedCounts` runs | `landedSpawnProbe` fires once (a criterion mismatch is stale regardless of TTL) |
| AC-SLL-006 | REQ-SLL-008 | the old-schema cache above; the git runner returns an error | refresh runs | the cache afterwards is still not `Known()` — the number 9 is never readable as a current-criterion measurement |
| AC-SLL-007 | REQ-SLL-005 | a current-criterion measured cache `landed == 3`; the git runner returns (a) an error, (b) a line with no `\x00` separator | refresh runs | cache `landed == 3`, `measured == true`, `fetched_at` advanced; never `landed == 0` |
| AC-SLL-007b | REQ-SLL-005 | no prior cache; the git runner errors | refresh runs | cache not `Known()` (existing never-measured test retained) |
| AC-SLL-007c | REQ-SLL-005 | no picked cards | refresh runs | 0 git invocations; cache `landed == 0`, `measured == true` |
| AC-SLL-008 | REQ-SLL-009 | `LandedCounts{Landed: 48, Measured: true, Available: true, <current criterion>}` and picked/queued 76/4 | the session line renders | line contains `🔄 TODO: 76/4 ⚑48` and contains no `✓` |
| AC-SLL-008b | REQ-SLL-009 | observed zero under the current criterion | the session line renders | line contains `🔄 TODO: 76/4 ⚑0` and no `✓` |
| AC-SLL-009 | REQ-SLL-011 | picked 76, queued 4, landed 48 (known) vs landed unknown | both lines render | both TODO segments start with the identical prefix `🔄 TODO: 76/4` — the picked number is not reduced |
| AC-SLL-010 | REQ-SLL-010 | the landed glyph constant | inspected in a test | it is exactly one rune `U+2691` and `runewidth.StringWidth(glyph) == 1` |
| AC-SLL-011 | REQ-SLL-012 | the tree after run phase | `grep -c '✓' docs-site/content/{ko,en,ja,zh}/advanced/statusline.md` and `grep -c '⚑N' …` | every `✓` count is 0 (catches both `✓N` and the `✓2` example) and every `⚑N` count is ≥ 1 (all 4 locales) |
| AC-SLL-012 | REQ-SLL-001 | the tree after run phase | `grep -c -- '--format=%B' internal/statusline/landed.go` | 0 — the body-stream query is gone |

## §D.1 Edge cases covered

- A subject that attributes two distinct cards attributes nothing (inherited from `subjectAttribution`; no statusline-specific test required beyond AC-SLL-001).
- Several commits attribute the same card: counted once (`LandedAttributions` keeps the newest).
- A picked id that does not match `landedCardToken` is excluded before the scan (existing behavior preserved).

## §D.2 Quality gates

- `go test ./internal/statusline/... ./internal/kanban/...` green.
- golangci-lint v2.1.6: 0 issues on changed packages.
- `go vet ./internal/statusline/...` clean.

## §D.3 Definition of Done

- All AC rows above PASS with command + verbatim output recorded in `progress.md` §E.2.
- Renderer comment and `landed.go` header describe the subject criterion and the "never subtracts" reason.
- docs-site 4 locales updated in the same change.
