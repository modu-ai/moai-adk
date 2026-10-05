# SPEC-GRAPH-EDGE-CONFIDENCE-001 — Implementation Plan

Tier M · development_mode: tdd (`.moai/config/sections/quality.yaml`) · RED-GREEN-REFACTOR
per milestone, scoped to touched packages only (no local full-suite runs — CI owns
the full verdict).

## §A. Context

The graphify gap analysis (`.moai/reports/graphify-codegraph-analysis-20260901.md`,
P1-1) recommends per-edge resolution confidence on `code-call` edges as the
pragmatic bypass of the never-assigned `GradeFull` gap. All extraction inputs
already exist in `internal/graph/symbol.Extract`'s single walk: call sites
(`CallEdge`), import sites (`ImportEdge`, with the `Local` flag), and function
declaration ranges (`astx.FuncRange` — currently consumed only for
`enclosingFunction`). The promotion join is a pure post-pass over that walk's
output.

## §B. Key Design Decision — edge model

**Decision: add two new optional `Edge` fields (`Resolution`, `Confidence`);
do NOT extend `Edge.Grade` semantics.**

| Option | Assessment |
|---|---|
| A — extend `Grade` | Rejected. `Grade` is a per-language CAPABILITY claim with a closed value set (`full`/`name-based`/`none`) enforced by `ValidateGradeMatrix` and published by `GradeMatrix`/`symbol.GradeMatrix`. Per-edge resolution EVIDENCE is an orthogonal axis. Overloading would (a) break the closed-set validator contract, (b) force a relabel of every existing `name-based` edge — changing bytes the doc-vs-code tests and grade-matrix tests pin, (c) conflate "what the language CAN resolve" with "how strongly THIS edge is believed". |
| B — additive fields (chosen) | `Resolution string \`json:"resolution,omitempty"\`` ∈ {`extracted`,`intra-package`,`inferred`} + `Confidence float64 \`json:"confidence,omitempty"\`` ∈ {1.0, 0.95, 0.85}. Appended AFTER `DisagreesWith` in the struct so existing field order — and therefore the key order of every existing serialized edge — is unchanged. `omitempty` keeps doc-derived and `code-import` edges byte-identical (REQ-GEC-006). Old consumers ignore unknown JSON keys; old artifacts load fine without the keys (REQ-GEC-009). Confidence is defined as a pure function of Resolution at exactly one point (`ConfidenceFor` in `internal/graph`), so the fields cannot drift. All values non-zero, so `omitempty` never drops a populated confidence. |

Both values are non-zero, so `omitempty` never drops a populated confidence.

## §C. Resolution derivation (WHAT, fixed here; exact code shape deferred to run)

Inputs per caller file F (all from the existing `symbol.Extract` walk):

1. `declaredNames(F)` — function/method names from `set.Functions` (`astx.FuncRange`).
   The seam must RETAIN these (today it uses them only transiently for
   `enclosingFunction`); no new parsing.
2. `dir(F)` — the file's directory (derivable from the already-retained rel path).
3. `imports(F)` — the file's `ImportEdge` set, with the seam's `Local` flag (true
   exactly when the raw specifier carried the go.mod module prefix and was
   stripped by `localizeModule`).
4. Global name→declaring-file index built from all files' declared names, keyed
   deterministically (sorted iteration; multiple declaring files all match the
   tier tests identically).

**DECISION (repair-round D3): add an explicit same-package tier.** Encoding:
`resolution: "intra-package"`, `confidence: 0.95`. Rationale: in Go, same
directory equals same package by construction (Go spec — no import statement is
needed or possible within a package), so a name declared in a sibling file is
stronger evidence than a bare name match, but weaker than an explicitly declared
cross-package dependency. A distinct label (not just a nudged `extracted`) keeps
the evidence KINDS honest — 1.0 stays reserved for declared-dependency or
same-file evidence — while the closed three-value domain keeps
`ConfidenceFor` a total single-definition mapping. Directory equality is
language-neutral (no import parsing), so this tier also gives the non-Go
call-grade languages their only promotion path in this delivery.

Tier evaluation per `code-call` edge (first match wins):

- **T1 same-file**: callee ∈ declaredNames(caller file) → `extracted`/1.0 (REQ-GEC-002).
- **T2 import evidence (Go-module domain ONLY)**: callee is declared in at least
  one file whose directory equals a module-prefixed import of the caller file
  (the seam's `Local` import set) → `extracted`/1.0 (REQ-GEC-003). HONEST DOMAIN,
  repair-round D1: `ImportSite.Module` is the specifier as written and
  `localizeModule` resolves declaring directories only for go.mod-prefixed
  paths — so T2 is well-defined ONLY for Go-module imports. The non-Go
  call-grade languages (python, javascript, typescript, java, rust) NEVER match
  T2 in this delivery; they fall to T3 or T4. Widening T2 to non-Go specifiers
  is out of scope (spec.md §F, candidate future work).
- **T3 same-package**: callee is declared in a file whose directory equals
  dir(caller file), cross-file → `intra-package`/0.95 (REQ-GEC-012).
- **T4 fallback**: otherwise → `inferred`/0.85 (REQ-GEC-004).

Value-domain reduction note (repair-round D5): the evidence source's third value
AMBIGUOUS is deliberately dropped — this delivery is annotation-only (no
consumer filters, ranks, or thresholds on the label), so an "unsure" tier would
add taxonomy without behavior. Two promotion classes plus the same-package
middle tier cover every edge; revisit when a confidence-consuming consumer lands.

Determinism guards: the pass consumes only sorted/derivable-from-content inputs;
no map-iteration-order dependence in any emitted value; the join is a pure
function of the tree.

Documented semantics note: T2 asserts "a module the file explicitly imports
declares this name" — evidence, not scope-aware proof (that remains
`GradeFull`'s job, out of scope). The resolution label must carry this honest
meaning in godoc.

## §D. File-level changes (expected touch surface)

| File | Change |
|---|---|
| `internal/graph/graph.go` | Add `Resolution` + `Confidence` fields to `Edge` (after `DisagreesWith`); add `ConfidenceFor(resolution) float64` single-definition helper. No changes to `EdgeLess`, `Build`, `WriteJSONL`. |
| `internal/graph/symbol.go` | Compute per-edge resolution in `mapSymbolEdges` (or a sibling pass) using the seam's enriched outputs; stamp `Resolution`/`Confidence` on `code-call` edges only. |
| `internal/graph/symbol/symbol.go` | Retain per-file declared function names in the extraction result (extend `CallEdge`/return shape or add a declarations index); enrich each `CallEdge` with its tier (or leave tier computation to `internal/graph` — seam stays domain-typed; prefer returning declared-names index + imports and letting the mapper decide, preserving the existing layering). |
| `internal/graph/codequery.go` | Add `Confidence` to `CodeMatch` and `CallTraceEdge`; copy from the source edge in `FindCode`/`TraceCalls`. No filtering/ranking changes (REQ-GEC-011). |
| `internal/cli/mcp_code_tools.go` | Pass-through only — handlers marshal the structs; verify descriptions mention confidence (description text update optional, non-breaking). |
| `internal/graph/*_test.go` | New table-driven tests: tier matrix, `ConfidenceFor` mapping, doc-edge byte-identity, double-build determinism, legacy-artifact load. |
| `internal/navigator/astx/*` | NO changes (REQ-GEC-007). Tests there must keep passing untouched. |

Freshness/provenance (`internal/graph/meta.go`, `check.go`) untouched: the edges
fingerprint covers SOURCE files, not artifact bytes, so adding fields does not
affect the gate. Existing artifacts simply rebuild.

## §E. JSONL schema compatibility

- **New artifacts**: `code-call` lines gain trailing `"resolution"`/`"confidence"`
  keys; all other lines byte-identical.
- **Existing artifacts**: no migration. `edges.jsonl` is untracked/derived — the
  backfill IS `moai graph build`. Consumers must accept absent fields (REQ-GEC-009).
- **Doc edges**: `omitempty` + field-order preservation ⇒ byte-identical
  (AC-GEC-006 pins this with a before/after comparison on a fixture tree).

## §F. Milestones (priority-ordered; no time estimates)

- **M1 (High) — Edge model + confidence derivation**: `Edge` fields,
  `ConfidenceFor`, seam declared-names retention, tier pass in the mapper,
  `mapSymbolEdges` stamping. RED first: failing tests for the AC-GEC-002/003/004/
  005/012 tier matrix on a `t.TempDir()` fixture tree (Go module + one
  dynamic-language fixture for grammar coverage). Then GREEN.
- **M2 (High) — Determinism + doc-edge additivity locks**: double-build
  byte-identity test (AC-GEC-001), doc-edge/code-import byte-identity before/after
  test (AC-GEC-006), provenance shape test (AC-GEC-011). These are regression
  locks — write them in the same milestone so the derivation cannot silently
  break the artifact contract.
- **M3 (Medium) — Consumer exposure**: `CodeMatch`/`CallTraceEdge` confidence
  fields + MCP handler pass-through (AC-GEC-007); legacy-artifact load test
  (AC-GEC-008).
- **M4 (Medium) — CGO-disabled + cross-layer regression**: `CGO_ENABLED=0` build +
  graph build/query behavior test (AC-GEC-009); run affected-package suites
  (`go test ./internal/graph/... ./internal/graph/symbol/... ./internal/cli/...`
  graph-scoped) confirming grade-matrix and traversal suites pass unchanged
  (AC-GEC-010).

Order rationale: M1/M2 are the highest-change-likelihood decisions (data model +
derivation semantics) and land first for review; M3/M4 are mechanical wiring and
regression confirmation.

## §G. Test strategy

- Table-driven unit tests (repo convention) for the tier matrix: same-file hit,
  import-evidence hit, name-only miss, plus ambiguity cases (name declared in BOTH
  an imported and a non-imported module → T2 wins; name undeclared anywhere → T3).
- Golden determinism: build a `t.TempDir()` fixture tree twice, `bytes.Equal` the
  two `edges.jsonl` writes.
- **Byte-identity guard (committed mechanism, repair-round D7)**: AC-GEC-006
  compares doc-kind + `code-import` lines against a COMMITTED golden file,
  `internal/graph/testdata/edges-doc-golden.jsonl` — the only CI-executable
  option, since a pre-change build cannot exist inside a post-change CI run.
  Pinning procedure: (1) on the base revision (before any confidence code
  lands), run the doc-layer builder over the committed fixture tree;
  (2) write the doc-kind + `code-import` JSONL lines to the golden file and
  commit it in the SAME run-phase change set, clearly attributed in the commit
  message as "generated on base <sha>"; (3) the test rebuilds the identical
  fixture tree and compares only `kind` ∈ {import, mx-spec, spec-depends,
  report-milestone, milestone-card, code-import} lines, byte-for-byte.
  Regeneration: re-run the generation step on the new base when the FIXTURE
  itself changes (fixture edits require the golden to be refreshed in the same
  PR, with the base sha named); the golden is never hand-edited.
- No goroutines introduced — the confidence pass is a synchronous post-walk; no
  `-race` surface change.
- CGO-disabled path covered by the existing `nocgo_test.go` pattern extended to
  assert zero confidence-bearing edges.

## §H. Risks

| Risk | Mitigation |
|---|---|
| Map-iteration nondeterminism leaking into tier results | Tier pass consumes only sorted inputs; M2's double-build lock fails the build if any ordering leaks. |
| Dynamic-language import specifiers cannot join T2 (repair-round D1 correction: `ImportSite.Module` is the specifier as written; `localizeModule` strips only a go.mod prefix) | Accepted by design: T2's effective domain is Go-module imports ONLY; non-Go calls resolve via T3 same-directory or T4 fallback. Widening is out of scope (spec.md §F). Tests must assert the non-Go fallback explicitly (AC-GEC-005), not assume promotion. |
| Name-index memory growth on large trees | Index built once per build over already-walked data; six-digit edge counts are within the artifact's observed scale (report risk note). No action beyond observation. |
| Temptation to filter/rank by confidence "while here" | REQ-GEC-011 forbids it; scope discipline — annotation only. |

## §I. Cross-References

- spec.md §C (REQ-GEC-*) · acceptance.md (AC-GEC-*)
- `.moai/reports/graphify-codegraph-analysis-20260901.md` § P1-1
- Sibling cards: SPEC-MX-TAG-EDGES-001 (t412), SPEC-GRAPH-REPORT-001 (t413)
