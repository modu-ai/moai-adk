# SPEC-GRAPH-EDGE-CONFIDENCE-001 — Acceptance Criteria

Verification layer. Every AC is binary-testable; the verification command is the
evidence contract. Given-When-Then scenarios live here; GEARS obligations live in
spec.md §C (REQ-GEC-001..012).

Fixture basis: all ACs run against a committed `t.TempDir()`-populated fixture
tree (mini Go module — module-prefixed imports per AC-GEC-004's T2 domain — plus
one dynamic-language package) unless the AC names the real tree.

## §D. AC Matrix

| AC | REQ | Scenario (Given / When / Then) | Verification command |
|---|---|---|---|
| AC-GEC-001 | REQ-GEC-005 | Given a fixed project tree / When `moai graph build` runs twice into two fresh copies of it / Then both `edges.jsonl` artifacts are byte-identical | `go test ./internal/graph/ -run TestEdgesJSONLDeterministic -count=1` (test builds twice + `bytes.Equal`); spot-check: `cmp <rootA>/.moai/project/graph/edges.jsonl <rootB>/.moai/project/graph/edges.jsonl` → exit 0 |
| AC-GEC-002 | REQ-GEC-001, 004 | Given a built `edges.jsonl` / When every `code-call` line is scanned / Then `resolution` ∈ {extracted, intra-package, inferred}, `confidence` ∈ {1.0, 0.95, 0.85}, and confidence ≡ `ConfidenceFor(resolution)` (extracted↔1.0, intra-package↔0.95, inferred↔0.85); no other values present | `go test ./internal/graph/ -run TestCodeCallConfidenceDomain -count=1`; artifact spot-check: `jq -r 'select(.kind=="code-call") | "\(.resolution) \(.confidence)"' edges.jsonl \| sort -u` → exactly `extracted 1`, `intra-package 0.95`, `inferred 0.85` |
| AC-GEC-003 | REQ-GEC-002 | Given a fixture caller file invoking a function declared in the same file / When the graph builds / Then that edge carries `resolution: extracted`, `confidence: 1.0` | `go test ./internal/graph/ -run TestSameFilePromotion -count=1` |
| AC-GEC-004 | REQ-GEC-003 | Given a Go-module fixture where a caller file imports a module-prefixed local package that declares the callee (specifier strippable by `localizeModule` — the ONLY T2 domain) / When the graph builds / Then that cross-file edge carries `resolution: extracted`, `confidence: 1.0` | `go test ./internal/graph/ -run TestImportEvidencePromotion -count=1` |
| AC-GEC-005 | REQ-GEC-004 | Given a fixture caller calling a name declared only in a module it does NOT import and outside its own directory — including the NON-GO case (a python or typescript file importing a specifier whose target directory is not derivable, repair-round D1) / When the graph builds / Then that edge carries `resolution: inferred`, `confidence: 0.85` — T2 never fires for non-Go specifiers | `go test ./internal/graph/ -run TestNameOnlyFallback -count=1` (fixture includes one Go + one dynamic-language case) |
| AC-GEC-006 | REQ-GEC-006 | Given the committed golden file `internal/graph/testdata/edges-doc-golden.jsonl` (doc-kind + `code-import` lines generated on the base revision — plan.md §G) / When the fixture tree is rebuilt / Then those edge lines compare byte-identical to the golden and carry no `resolution`/`confidence` keys | `go test ./internal/graph/ -run TestDocEdgesByteIdentical -count=1`; artifact spot-check: `grep -c 'resolution' <(grep -v '"kind":"code-call"' edges.jsonl)` → 0 |
| AC-GEC-007 | REQ-GEC-008 | Given a built artifact / When `graph_find_code` and `graph_trace_calls` return matches over code-call edges / Then each match exposes the source edge's confidence | `go test ./internal/cli/ -run TestGraphMCPConfidenceExposed -count=1` (handlers via `go test` against the handler funcs) |
| AC-GEC-008 | REQ-GEC-009 | Given an `edges.jsonl` written WITHOUT the new fields / When consumers load and query it / Then load succeeds, confidence surfaces absent/unknown, no error | `go test ./internal/graph/ -run TestLegacyArtifactLoad -count=1` |
| AC-GEC-009 | REQ-GEC-010 | Given `CGO_ENABLED=0` / When the binary builds and `moai graph build` runs / Then behavior matches today: build + query succeed, zero code-derived edges, zero confidence-bearing lines | `CGO_ENABLED=0 go build ./... && CGO_ENABLED=0 go test ./internal/navigator/astx/ -run TestNoCGO -count=1` + `CGO_ENABLED=0 go test ./internal/graph/ -run TestNoCGOConfidenceInert -count=1` |
| AC-GEC-010 | REQ-GEC-007, 011 | Given the existing grade-matrix and traversal suites / When affected packages are tested / Then all pass unchanged — Grade values, GradeFor mapping, FindCallers/BlastRadius/TraceCalls/FindCode result sets untouched | `go test ./internal/graph/... ./internal/navigator/astx/ -count=1` → ok (exit 0) |
| AC-GEC-011 | REQ-GEC-005 | Given a build over the fixture tree / When `edges.meta.json` is inspected / Then provenance shape unchanged (commit SHA + source fingerprints; no wall-clock timestamp in `edges.jsonl`) | `go test ./internal/graph/ -run TestProvenanceShapeUnchanged -count=1` |
| AC-GEC-012 | REQ-GEC-012 | Given a fixture where a caller file calls a name declared in a SIBLING file of the same directory (Go: same package by construction; language-neutral) / When the graph builds / Then that cross-file edge carries `resolution: intra-package`, `confidence: 0.95` | `go test ./internal/graph/ -run TestSamePackagePromotion -count=1` |

## §D.1 Severity

- **Blocker**: AC-GEC-001, AC-GEC-002, AC-GEC-006 (artifact contract + value
  domain — a violation invalidates the JSONL determinism guarantee).
- **Major**: AC-GEC-003, AC-GEC-004, AC-GEC-005, AC-GEC-012 (the feature
  itself), AC-GEC-007 (consumer surface), AC-GEC-009 (CGO gate).
- **Minor**: AC-GEC-008, AC-GEC-010, AC-GEC-011 (compat + regression locks —
  Minor in severity, still must-pass).

## §D.2 Edge cases covered

- Name declared in both an imported module and a same-directory sibling →
  `extracted` (T2 before T3; declared dependency outranks co-location).
- Name declared in both an imported and a non-imported, non-sibling module →
  `extracted` (T2; documented evidence semantics).
- Callee name matching a method vs plain function (selector `x.Do` → `Do`) —
  tier logic operates on the extractor's normalized name.
- Caller at top level (no enclosing function; `Caller == ""`) — tier logic is
  file-scoped, not function-scoped; unaffected.
- File with zero imports calling cross-file → `intra-package` when the declaring
  file shares the caller's directory (T3), else `inferred` (T4). (Repair-round
  D3 correction: the same-directory path is promotion, not fallback.)
- Non-Go language calling a name whose declaring directory cannot be derived
  from the raw import specifier → never T2; `intra-package` only when
  same-directory, else `inferred` (repair-round D1).
- Value-domain reduction (repair-round D5): the evidence source's AMBIGUOUS
  value is deliberately absent — annotation-only scope, no LLM tier; the
  closed domain is exactly {extracted, intra-package, inferred}.
- Empty tree / no sources → zero edges, no confidence lines, no error.

## §D.3 Quality gates

- `gofmt` clean on touched files; `go vet` clean on touched packages;
  `golangci-lint run` on touched packages.
- Coverage on `internal/graph` and `internal/graph/symbol` ≥ 85% for new code
  paths (repo package target).
- No `t.Setenv` with OTEL vars; all fixtures under `t.TempDir()`.

## §D.4 Definition of Done

1. All AC-GEC-001..012 verified with their commands, evidence persisted under the
   run-phase verify path.
2. No changes outside the §D file surface of plan.md (esp. zero changes in
   `internal/navigator/astx`).
3. `edges.jsonl` determinism and doc-edge byte-identity locks green.
4. MCP responses expose confidence; legacy artifacts still load.
