# SPEC-MX-TAG-EDGES-001 — Acceptance Criteria

Verification layer. Every AC is binary-testable; the verification command is
the evidence contract. Given-When-Then scenarios live here; GEARS obligations
live in spec.md §C (REQ-MTE-001..015).

Fixture basis: all ACs run against a committed `t.TempDir()`-populated fixture
tree (Go module + tags of all six kinds — function-anchored, file-scope, DEBT
with/without `@MX:UPGRADE` — plus a test-fixture caller set and one
dynamic-language file) unless the AC names the real tree.

## §D. AC Matrix

| AC | REQ | Scenario (Given / When / Then) | Verification command |
|---|---|---|---|
| AC-MTE-001 | REQ-MTE-003 | Given a fixed project tree / When `moai graph build` runs twice into two fresh copies of it / Then both `edges.jsonl` artifacts are byte-identical INCLUDING all `mx-*` lines (no wall-clock, no scan-order dependence) | `go test ./internal/graph/ -run TestEdgesJSONLDeterministicWithTags -count=1` (double build + `bytes.Equal`); spot-check: `cmp <rootA>/.moai/project/graph/edges.jsonl <rootB>/.moai/project/graph/edges.jsonl` → exit 0 |
| AC-MTE-002 | REQ-MTE-001 | Given a built `edges.jsonl` / When every tag line is scanned / Then each standalone @MX tag occurrence appears as exactly one edge with `kind` ∈ {mx-note, mx-warn, mx-anchor, mx-todo, mx-legacy, mx-debt}, and no other new kind exists | `go test ./internal/graph/ -run TestTagEdgeKindDomain -count=1`; artifact spot-check: `jq -r '.kind' edges.jsonl \| grep '^mx-' \| sort -u` → exactly the six kinds present in the fixture |
| AC-MTE-003 | REQ-MTE-002 | Given a fixture with (a) a tag inside a function and (b) a file-scope tag / When the graph builds / Then (a)'s edge carries `source` = file, `target` = the enclosing symbol name, `line` = the tag line, and (b)'s edge carries `target` = the file path itself (self-edge) — no tag occurrence dropped | `go test ./internal/graph/ -run TestTagEdgeEndpoints -count=1` |
| AC-MTE-004 | REQ-MTE-004 | Given the committed golden `internal/graph/testdata/edges-doc-golden.jsonl` (regenerated in this change set, base SHA named in the commit message, now pinning doc kinds + `code-import` + all `mx-*` kinds) / When the fixture tree rebuilds / Then every golden-pinned kind's lines compare byte-identical under the per-kind-filtered comparison, and `mx-*` lines carry no `resolution`/`confidence`/`rot_risk` keys | `go test ./internal/graph/ -run TestDocEdgesByteIdentical -count=1`; artifact spot-check: `grep -E '"kind":"mx-' edges.jsonl \| grep -cE 'resolution\|confidence\|rot_risk'` → 0 |
| AC-MTE-005 | REQ-MTE-005 | Given an `edges.jsonl` written WITHOUT any `mx-*` lines (pre-upgrade artifact) / When consumers load and query it / Then load succeeds, queries answer, no error | `go test ./internal/graph/ -run TestLegacyArtifactLoad -count=1` |
| AC-MTE-006 | REQ-MTE-006 | Given a build over the fixture tree / When the doc layer runs / Then exactly ONE `mx.Scanner.ScanDir` pass executes per build (seam-counted), and its output feeds both `mx-spec` and `mx-*` edges | `go test ./internal/graph/ -run TestSingleScanPerBuild -count=1` (scan seam counter assertion) |
| AC-MTE-007 | REQ-MTE-003, 007 | Given the fixture's DEBT tags — one WITH and one WITHOUT an `@MX:UPGRADE` sub-line / When the graph builds / Then both emit edges with IDENTICAL key sets (occurrence only), and neither `RotRisk` state nor any scanner wall-clock value reaches any `mx-*` line | `go test ./internal/graph/ -run TestTagEdgesCarryNoMetadata -count=1` (includes the rot-state pair assertion) |
| AC-MTE-008 | REQ-MTE-008 | Given traversal queries over (a) the pre-change edge set with `mx-*` edges zeroed and (b) the full set / When `FindCallers` and `BlastRadius` run / Then (a) returns byte-identical results to the pre-change behavior, and (b) shows `mx-*` edges propagating reverse-only (a `mx-debt` target's caller set gains the tag's file; the tag's file's blast radius does NOT gain the target's callers) | `go test ./internal/graph/ -run TestTraversalAdditivityWithTags -count=1` |
| AC-MTE-009 | REQ-MTE-009, 013 | Given the parity fixture — symbol S declared in file D, called from 3 distinct evidence-backed EXTERNAL caller files (extracted/intra-package), plus 2 comment/string mentions, 2 same-file call sites inside D, 1 inferred-only caller file, and 3 test-file callers / When the P1 evidence source computes fan-in of S / Then the graph source returns blocking count = 3 (same-file calls and test callers excluded, inferred itemized in the violation `Reason` as "(+1 inferred-only)"), while the in-test textual-matched reference returns its larger occurrence count (same-file calls count there — `fanIn` subtracts only the declaration token); the delta direction (graph ≤ textual) is asserted as the DOCUMENTED outcome | `go test ./internal/graph/ -run TestGraphFanInParityFixture -count=1` (home per plan §C: the textual reference is an IN-TEST counter in the same binary — layering lock intact); artifact spot-check: `jq -r 'select(.kind=="code-call" and .target=="S") \| "\(.source) \(.resolution)"' edges.jsonl` |
| AC-MTE-010 | REQ-MTE-010 | Given the complete non-test validator construction set (post_tool.go textual-by-budget; session_end.go the sole batch site) / When each constructs its validator / Then the PostToolUse wiring constructs the TEXTUAL source only (default constructor unchanged — cost profile intact) and SessionEnd constructs the edge-backed source; and `moai mx scan` constructs NO validator before or after this SPEC (negative assertion: mx scan's code path contains no validator construction) | `go test ./internal/hook/ -run TestPostToolUseKeepsTextualFanIn -count=1` + `go test ./internal/hook/mx/ -run TestConstructorDefaultsTextualSource -count=1` + `go test ./internal/hook/ -run TestSessionEndSelectsEdgeSource -count=1` |
| AC-MTE-011 | REQ-MTE-011 | Given the edge-backed source selected but `edges.jsonl` ABSENT (fresh clone) or stale-by-probe / When P1 validation runs / Then the validator falls back to the textual source, the verdict carries the source label, and no error surfaces (stale artifacts answer from artifact with a labeled stderr note, never silently) | `go test ./internal/hook/mx/ -run TestFanInFallbackLabeled -count=1` |
| AC-MTE-012 | REQ-MTE-012 | Given a fixture where S has 3 test-file callers and 3 source callers / When the graph source computes the blocking count / Then it returns 3 (tests excluded); and Given S with ONLY test-file callers / Then the blocking count is 0 (REQ-SPC-004-040 hard-coded fallback patterns: `*_test.go`, `tests/`, `fixtures/`, `testdata/` — the package-level predicate at `internal/mx/fanin.go:47`; mx.yaml `test_paths` NOT honored, accepted divergence per spec.md §E) | `go test ./internal/graph/ -run TestHubExclusionTestCallers -count=1` |
| AC-MTE-013 | REQ-MTE-014 | Given a built artifact with DEBT tags — including one file-scope DEBT / When `moai graph query --debt-fanin` runs / Then it prints mx-debt targets ranked by evidence-backed fan-in (descending, ties by target), the file-scope DEBT appears at fan-in 0 with a `(self)` marker (listed, not omitted); `--fanin` still prints the import ranking unchanged; and the stand-in strings are retired (the `query.go` `@MX:NOTE` and the `--fanin` help carry neither the hyphenated nor the unhyphenated stand-in phrasing nor "no tag-kind edges yet") | `go test ./internal/cli/ -run TestGraphQueryDebtFanIn -count=1`; stand-in retirement: `grep -rnE "stands in for\|stand-in\|no tag-kind edges yet" internal/graph/query.go internal/cli/graph.go` → 0 matches |
| AC-MTE-014 | REQ-MTE-015 | Given `CGO_ENABLED=0` / When the binary builds and the graph + validator run / Then `mx-*` edges still emit (self-edge form — no range data), the graph-backed source degrades to the textual fallback with its label, zero `code-call` edges exist, and no new error surface appears | `CGO_ENABLED=0 go build ./...` + `CGO_ENABLED=0 go test ./internal/graph/ -run TestNoCGOTagEdgesSelfEdge -count=1` + `CGO_ENABLED=0 go test ./internal/hook/mx/ -run TestNoCGOFanInTextualFallback -count=1` |

## §D.1 Severity

- **Blocker**: AC-MTE-001, AC-MTE-004 (artifact contract — a violation
  invalidates the JSONL determinism guarantee), AC-MTE-009 (the P1-3 core
  semantics).
- **Major**: AC-MTE-002, AC-MTE-003, AC-MTE-006, AC-MTE-010, AC-MTE-011,
  AC-MTE-012, AC-MTE-013 (the feature surfaces), AC-MTE-014 (CGO gate).
- **Minor**: AC-MTE-005, AC-MTE-007, AC-MTE-008 (compat + regression locks —
  Minor in severity, still must-pass).

## §D.2 Edge cases covered

- File-scope tag (no enclosing function) → self-edge; never dropped
  (AC-MTE-003).
- Tag inside a nested closure → innermost-containing range wins (mirrors
  `enclosingFunction` semantics); unmatched line → self-edge.
- DEBT with vs without `@MX:UPGRADE` → identical edge bytes; rot state stays
  scanner-side (AC-MTE-007).
- Duplicate-looking tags (same file, same symbol, multiple tags) → one edge
  per tag occurrence, distinguished by `line` under `EdgeLess`.
- Symbol called only from its own file → graph blocking count 0. HONEST
  BASIS: the textual index does NOT exclude same-file callers (`fanIn`
  subtracts only the declaration token — same-file call occurrences still
  count), so this is a DELIBERATE SHARPENING by the graph source, not parity.
  It can FLIP hook-vs-batch verdicts: a symbol with 3+ same-file call sites is
  flagged P1 by the PostToolUse textual source and NOT flagged by the
  SessionEnd graph-backed authority — accepted by ruling (ANCHOR semantics is
  external dependents; the batch verdict is the enforced one).
- Test-fixture-only callers → blocking count 0 (AC-MTE-012).
- Fresh clone, no artifact → textual fallback, labeled, no error
  (AC-MTE-011).
- Direction guarantee: each graph-side filter (confidence, distinct caller
  files, declaring-file exclusion, test-caller exclusion) only REMOVES
  candidates the textual index counts ⇒ graph count ≤ textual count on every
  fixture — one-directional by construction (plan.md §B.7), asserted in
  AC-MTE-009.
- CGO-off build → self-edges + textual degrade (AC-MTE-014); golden compares
  only under the CGO-available CI build (documented in spec.md §E).
- Empty tree / zero tags → zero `mx-*` lines, no error.

## §D.3 Quality gates

- `gofmt` clean on touched files; `go vet` clean on touched packages;
  `golangci-lint run` on touched packages.
- Coverage on `internal/graph` and `internal/hook/mx` ≥ 85% for new code
  paths (repo package target).
- `go list -deps` guard: `internal/hook/mx` imports neither `internal/graph`
  nor `internal/cli` (layering lock, mirrors AC-GF-016's pattern).
- No `t.Setenv` with OTEL vars; all fixtures under `t.TempDir()`.

## §D.4 Definition of Done

1. All AC-MTE-001..014 verified with their commands; evidence persisted under
   the run-phase verify path.
2. No changes outside plan.md §C's file surface (esp. zero changes in
   `internal/navigator/astx`, `internal/mx/scanner.go`, and
   `internal/hook/post_tool.go`).
3. Determinism + byte-identity locks green; golden regenerated with base SHA
   attributed.
4. `--debt-fanin` answers from the artifact; PostToolUse latency profile
   unchanged; batch surfaces answer from evidence-backed fan-in with labels.
