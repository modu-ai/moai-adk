# SPEC-MX-TAG-EDGES-001 — Implementation Plan

Tier M · development_mode: tdd (`.moai/config/sections/quality.yaml`) ·
RED-GREEN-REFACTOR per milestone, scoped to touched packages only (no local
full-suite runs — `origin/develop` CI owns the full verdict; repo-local
verification-load discipline).

## §A. Context

This SPEC implements graphify recommendations P1-2 (tag-kind edges) and P1-3
(graph-backed fan-in) on top of the already-merged confidence layer
(SPEC-GRAPH-EDGE-CONFIDENCE-001). The tree already carries: per-edge
`Resolution`/`Confidence` on `code-call` edges, the T1-T4 promotion tiers, the
committed doc golden, and the freshness/provenance sidecar. Two assets make
the tag layer cheap: `mxSpecEdges` (graph.go:252-277) already runs one full
`mx.Scanner.ScanDir` and discards every non-SpecRef tag, and `symbol.Extract`
already walks every function range and drops the ranges after the transient
`enclosingFunction` lookup.

## §B. Design Decisions

### B.1 (P1-2) Edge kind naming — one kind per tag kind [DECIDED]

**Decision: `mx-note` / `mx-warn` / `mx-anchor` / `mx-todo` / `mx-legacy` /
`mx-debt` — one kind per standalone tag kind, uniform over the closed
`TagKind` domain.**

| Option | Assessment |
|---|---|
| A — one kind per tag kind (chosen) | Kind is the artifact's question axis (`ImportFanIn` filters `KindImport`; `UnreferencedSpecs` filters `KindMXSpec`; `FindCallers`/`BlastRadius` traverse generically). A tag query is then a plain kind filter, sorting stays `EdgeLess`-natural, and the scanner's 6-kind domain is already closed. The delegation brief named four kinds (ANCHOR/DEBT/WARN/NOTE); emitting all six from one code path avoids a per-kind carve-out that the next tag kind would break. |
| B — one `mx-tag` kind + `tag:"DEBT"` attribute | Rejected: every consumer post-filters by attribute, the kind stops answering the question, and `Edge` gains a field that duplicates what the kind namespace already encodes. |

Cost note: this repo carries hundreds of `@MX:NOTE` tags — the artifact grows
by the full tag population. The report's risk note already accepts six-digit
edge counts; `mx-*` lines are short (4-5 keys). Accepted with observation, not
mitigation.

### B.2 (P1-2) Edge endpoints — file → enclosing symbol, self-edge fallback [DECIDED]

`source` = repo-relative file path (identical shape to `mx-spec`);
`target` = the innermost function/method whose declared range contains the tag
line, else the file path itself; `line` = the tag's 1-based line.

Rationale: the symbol target is the load-bearing join — it is exactly what
P1-3's fan-in and `--debt-fanin` ranking aggregate over. The self-edge
fallback (file-scope tags, or no range data under CGO-off) keeps every
occurrence represented: a dropped edge is indistinguishable from an unscanned
one, and "the graph silently lost tags" is the failure shape REQ-MTE-002
exists to prevent. The range join RETAINS what `Extract` already computes
(`astx.FuncRange` is consumed transiently today); `FileDecls` gains the
retained ranges (or a sibling type) — no second parse pass.

### B.3 (P1-2) Rot-risk / ceiling metadata — stays scanner-side [DECIDED]

Edges carry occurrence + endpoints only (REQ-MTE-007). The mx sidecar is the
tag-content SSOT (`moai mx query --kind DEBT --json` already flags
`rotRisk: "no-trigger"`); duplicating rot state in `edges.jsonl` would create
a second rot-truth that can drift from the sidecar on the refresh path.
`RotRisk` IS a pure function of tree content (deterministic), so this is a
single-source-of-truth ruling, not a determinism necessity — revisit only if
a graph-side consumer needs rot state without sidecar access (candidate
future work, t413-adjacent).

### B.4 (P1-2) JSONL back-compat — no new Edge fields, no migration [DECIDED]

The kind namespace carries the tag kind, so `Edge` needs no new fields:
new kinds are new LINES. Consequences, all accepted: (a) `LoadJSONL` loads old
artifacts unchanged (absent kinds = absent edges, REQ-MTE-005); (b) old
binaries reading a NEW artifact will include `mx-*` edges in `FindCallers`/
`BlastRadius` (generic kind traversal) — acceptable because the artifact is
untracked/derived and the backfill is a rebuild (t411 precedent); (c) the
globally sorted output inserts `mx-*` lines between existing kinds, so
"byte-identity of existing kinds" is SET-identity under the golden's
per-kind-filtered comparison, never whole-file byte identity (REQ-MTE-004
phrased accordingly; the existing golden test already compares filtered).

### B.5 (P1-2) Golden scope — extends, same file [DECIDED]

The committed golden `internal/graph/testdata/edges-doc-golden.jsonl` gains
`mx-*` lines: the fixture tree gains tags of each of the six kinds (including
a file-scope tag and a tag inside a function, and a DEBT both with and without
an `@MX:UPGRADE` sub-line — rot state must NOT leak into the edge either way),
and the golden is regenerated in the same run-phase change set per t411's
documented procedure, commit message naming the base SHA. Rationale for
extension rather than a second golden: `mx-*` edges are doc-layer products
(scanner-derived, CGO-free, deterministic) — exactly the population the
golden was created to pin; two golden files for one mechanism is drift bait.
`code-call` lines remain unpinned (CGO-availability-dependent), unchanged from
t411's ruling. CGO-off output (all self-edges) is covered by table assertions,
not the golden (CI builds with CGO available).

### B.6 (P1-3) Wiring — two-tier: hook textual, batch surfaces graph-backed [DECIDED]

| Option | Assessment |
|---|---|
| (a) Hook consumes `edges.jsonl` (stale-but-labeled) inside PostToolUse | Rejected as the PRIMARY path: artifact load+parse is borderline against the 500ms budget at six-digit edge counts, the hook's existing textual index is already O(project bytes) and therefore not obviously worse, and freshness cannot be repaired inside the budget (no refresh in the hook). |
| (b) Two-tier: hook keeps the instant textual signal; the graph-backed source is the AUTHORITY on the batch surface (chosen) | PostToolUse keeps `fanInIndex` (500ms budget intact, REQ-MTE-010); the SessionEnd batch (4000ms) — the SOLE non-PostToolUse validator construction site, verified: `mx_scan.go` constructs scanner+manager only, no validator — constructs the validator with the edge-backed source, freshness-labeled via the existing meta sidecar probes (`EdgesSourcesMovedFor` / `MXIndexNeedsRefresh`) — stale artifact ⇒ answer from artifact with a labeled stderr note (stale-but-labeled beats no answer); absent ⇒ textual fallback, labeled (REQ-MTE-011). `moai mx scan`'s role is sidecar PRODUCER feeding those freshness probes — no validation behavior is added to it. The P1 verdict's authority therefore moves to where the budget allows evidence, without regressing hook latency. |
| (c) SessionStart deferred background refresh + hook reads the index | Rejected AND OUT OF SCOPE: deferred SessionStart background edges refresh is assigned to sibling card t413 / SPEC-GRAPH-REPORT-001 by both the analysis report and SPEC-GRAPH-EDGE-CONFIDENCE-001 §F. Building it here would collide with the sibling card's scope. |

Layering mechanics: `internal/hook/mx` defines `FanInEvidenceSource`
(`EvidenceBacked(ctx, funcName, currentFile) (evidence int, inferredOnly int,
label string)`); the textual implementation wraps today's `fanInIndex`; the
edge-backed implementation lives in `internal/graph` (structurally compatible
method) and is injected at the one construction site permitted and required to
import `internal/graph`: `internal/hook/session_end.go`. `internal/hook`
cannot import `internal/cli` (documented cycle) — this is why the adapter
lives with the data, not with the CLI.

### B.7 (P1-3) Confidence + declaring-file ruling — evidence-backed external callers only [DECIDED]

**The ≥3 threshold counts DISTINCT caller FILES of `code-call` edges with
confidence ∈ {extracted, intra-package}, EXCLUDING the file that declares the
symbol; `inferred` callers are counted separately and reported in the
violation `Reason` — never added.** Rationale: (a) a name-coincidence match is
not a caller (the same honest-evidence philosophy as the confidence SPEC);
(b) ANCHOR's semantics is external dependents — a same-file call is not
external blast radius; a weighted sum (1.0/0.95/0.85) would put a fractional
threshold inside a violation reason string — unexplainable in review.

**The declaring-file exclusion is a DELIBERATE SHARPENING, not parity**: the
textual index does NOT exclude same-file callers — `fanIn`
(validator.go:132-138) subtracts only the declaration token (`count - 1` for
`currentFile`), so same-file call occurrences still count toward the textual
fan-in. Consequence to state plainly: on a symbol called only from its own
file (3+ same-file call sites), the PostToolUse textual source CAN flag P1
while the SessionEnd graph-backed authority does NOT — the two tiers can flip
verdicts on same-file-only callers. This is accepted: the batch authority's
verdict is the enforced one (sync-phase strict), and the sharpened semantics
is the intended ruling; the hook's textual signal remains a superset
advisory. Documented in acceptance.md §D.2 as the flip case.

**Direction argument under both filters (survives verification)**: each of
the three graph-side filters — confidence ∈ {extracted, intra-package},
distinct caller FILES, declaring-file exclusion — only removes candidates the
textual index would have counted (the textual index counts occurrences across
all .go files incl. tests, same-file included, comments and strings
included). Therefore graph count ≤ textual count holds on EVERY fixture; the
delta is one-directional by construction, not by fixture luck. This is why
AC-MTE-009 asserts the direction plus the itemized tails, not numeric parity.

Distinctness note: the textual index counts OCCURRENCES (multiple call sites
in one file each count); the graph counts DISTINCT caller files. The spec's
threshold wording ("3 callers") is satisfied more faithfully by distinctness;
recorded as a deliberate semantic tightening of the evidence, not a threshold
change (3 stays 3).

### B.8 (P1-3) Hub exclusion — IN, test-pattern set only [DECIDED]

Test-fixture callers are EXCLUDED from the blocking count, using the
REQ-SPC-004-040 hard-coded fallback pattern set already implemented as the
package-level unexported predicate `isTestFileWithPatterns`
(`internal/mx/fanin.go:47` — not a method; `*_test.go` suffix; `tests/`,
`fixtures/`, `testdata/` path components). Rationale: a test caller
exercises the contract, it does not create the maintenance obligation the
ANCHOR tag records; the same exclusion already governs `excludeTests` on the
query-side counter, so one discipline now governs both. Generated/vendor files
need no new exclusion (extractor walk skips `testdata/` + dot-dirs; validator
skips `vendor`/`*_generated.go`/`mock_*.go`). Run-phase verification item:
`symbol.Extract` walks the described roots without an explicit `*_test.go`
skip observed in this tree — the exclusion is applied at the fan-in
AGGREGATION layer (REQ-MTE-012) precisely so it does not depend on extractor
scope.

**mx.yaml `test_paths` divergence — ACCEPTED, not wired (minimal change)**:
the graph-backed source uses only the hard-coded fallback patterns; a user's
mx.yaml `test_paths` globs (accepted by `TextualFanInCounter.TestPaths`) are
NOT honored by the authority in this delivery — the two sources can diverge
for a project configuring custom globs. Recorded as accepted with candidate
future work in spec.md §E; wiring mx.yaml through `internal/graph` would add a
config dependency to the data layer for a rare configuration, which the
minimal-change ladder rejects here.

### B.9 Query surface [DECIDED]

`moai graph query --debt-fanin [--limit N]`: ranks `mx-debt` edge targets by
REQ-MTE-009 fan-in, descending, ties by target, printing
`count\ttarget\tfile` rows. Self-edge targets (file-scope DEBT tags, no
enclosing symbol) rank at fan-in 0 and are LISTED explicitly — omitting them
would recreate the silent-drop shape REQ-MTE-002 exists to prevent; a
`(self)` marker distinguishes them. `--fanin` (import ranking) keeps its
behavior —
the IMPORT ranking remains a legitimate question — but its stand-in framing
retires: the `@MX:NOTE` at `query.go:15` and the `--fanin` help text at
`graph.go:96-97` are rewritten to point at `--debt-fanin`. `FindCallers` /
`BlastRadius` need NO code change (generic kind traversal); `mx-*` edges join
reverse-only by the traversal rule (REQ-MTE-008) — not bidirectional like
`mx-spec`, because a tag lives in the same file as its symbol, so the reverse
direction already carries the dependency semantics and bidirectionality would
inflate every tag-carrying file's blast radius.

## §C. File-level changes (expected touch surface)

| File | Change |
|---|---|
| `internal/graph/graph.go` | Add 6 `mx-*` kind constants; `tagEdges(tags []mx.Tag, ranges) []Edge` derived from the SAME scan `mxSpecEdges` uses (refactor to one `ScanDir` + two mappings); `Build()` includes the tag layer. No `Edge` field changes. |
| `internal/graph/symbol/symbol.go` | Retain per-file `astx.FuncRange`s in the seam output (extend `FileDecls` or add `FileRanges`); ranges are already computed — retention only. |
| `internal/graph/query.go` | Add `SymbolFanIn(edges, name) (evidence int, inferredOnly int)` (or struct return) — the REQ-MTE-009 pure query; rewrite the stand-in `@MX:NOTE` on `ImportFanIn`. |
| `internal/graph/fanin_edge.go` (new) | Edge-backed `FanInEvidenceSource` implementation: confidence tiers + test-pattern hub exclusion (REQ-MTE-012) + distinct-caller counting. |
| `internal/hook/mx/validator.go` | `FanInEvidenceSource` interface; `fanInIndex` wrapped as the textual implementation; P1 rule consumes the injected source; violation `Reason` carries the inferred-only tail (REQ-MTE-013). Default constructor keeps textual (REQ-MTE-010). |
| `internal/hook/mx/types.go` | Interface doc if needed; no Validator interface change beyond an optional source-injection constructor. |
| `internal/hook/session_end.go` | Construct validator with the edge-backed source (+ freshness label, REQ-MTE-011). The SOLE batch injection site — `moai mx scan` constructs no validator and is NOT touched (its role stays sidecar producer feeding the freshness probes). |
| `internal/cli/graph.go` | `--debt-fanin` mode + selector-set update; `--fanin` help text rewrite. |
| `internal/graph/testdata/edges-doc-golden.jsonl` | Regenerated: fixture gains per-kind tags; base SHA named in commit message. |
| `internal/graph/fanin_edge_test.go` (new) | Parity fixture (AC-MTE-009): the edge-backed source is compared against an IN-TEST textual-matched REFERENCE counter (a test-local occurrence counter replicating `fanInIndex` semantics over the fixture bytes) — this is the only package where both the graph source and the textual semantics are observable in one test binary without violating the layering lock (`internal/hook/mx` imports neither `graph` nor `cli`; `internal/hook` importing `hook/mx` blocks the reverse). The reference is test-only and never a production seam. |
| `internal/hook/mx/*_test.go`, `internal/hook/*_test.go` | Source-selection tests: constructor default = textual (in `internal/hook/mx`); SessionEnd wiring observes the edge-backed injection (in `internal/hook`); fallback labels; PostToolUse-untouched assertion. |
| `*_test.go` across other touched packages | RED-first table tests: kind domain, endpoint join, self-edge fallback, determinism, golden, legacy load, single-scan, traversal additivity, hub exclusion, `--debt-fanin`, CGO-off. |

PRESERVE (no changes): `internal/mx/scanner.go` (the scanner already emits
`Kind`/`File`/`Line` per tag — zero scanner changes), `internal/mx/tag.go`,
the mx sidecar/refresh path, `internal/cli/mx_scan.go` (constructs no
validator — sidecar producer only, not a wiring surface),
`internal/navigator/astx` (ranges consumed, not
changed), `internal/graph/reader.go` (generic already), `internal/graph/
check.go`/`meta.go` (freshness fingerprint covers SOURCE files — new kinds do
not affect the gate), `internal/cli/mcp_code_tools.go` (MCP confidence
exposure unchanged), `internal/hook/post_tool.go` (PostToolUse wiring
untouched — textual by construction).

## §D. Test strategy

- Table-driven unit tests (repo convention), `t.TempDir()` fixtures; no
  `t.Setenv` with OTEL vars; no new goroutines.
- **Fixture tree** (one committed `t.TempDir()`-populated builder shared by
  graph tests): Go module + tags of all six kinds (function-anchored,
  file-scope, DEBT with/without `@MX:UPGRADE`) + a test-fixture caller set for
  hub exclusion + a python/typescript file for language coverage.
- **Determinism**: double-build `bytes.Equal` (AC-MTE-001); wall-clock mutant
  probe — stamp `LastSeenAt` into an edge → golden/determinism test must RED.
- **Golden**: regenerated per §B.5; the golden test's mutant probe (t411
  pattern) — emit a rotated `mx-*` kind or a `RotRisk` key → RED observed
  before accepted.
- **Parity fixture (AC-MTE-009, home: `internal/graph/fanin_edge_test.go`
  per §C)**: fixture with (a) a symbol called from 3 distinct evidence-backed
  EXTERNAL caller files, (b) 2 comment/string mentions, (c) 2 same-file call
  sites (excluded by the declaring-file rule — the flip case), (d) 1
  inferred-only caller file, (e) 3 test-file callers. Assert: the in-test
  textual-matched reference returns its larger occurrence count; the graph
  source returns blocking count = 3, inferred-only = 1, reason string carries
  "(+1 inferred-only)". The delta is the DOCUMENTED output, not a failure.
- **RED discipline**: every AC's RED observed before GREEN per
  verification-completeness §2 (two-cell adoption); for CGO-off paths where
  RED is structurally unreachable, the t411 precedent applies — observe the
  test RUNNING and sweeping (non-empty) via `-v` and record the gap.
- Verification load: `go test ./internal/graph/... ./internal/hook/mx/...
  ./internal/cli/` scoped runs only; `CGO_ENABLED=0` variant for AC-MTE-014.

## §E. Milestones (priority-ordered; no time estimates)

- **M1 (High) — Edge model + tag layer**: kind constants, single-scan
  refactor, range retention in the seam, endpoint join + self-edge fallback,
  sort inclusion. RED first: kind-domain and endpoint-join tests (AC-MTE-002/
  003) + single-scan seam test (AC-MTE-006). Then GREEN.
- **M2 (High) — Determinism + byte-identity locks**: golden extension +
  regeneration (AC-MTE-004), double-build (AC-MTE-001), legacy load
  (AC-MTE-005), metadata-absence (AC-MTE-007), traversal additivity +
  reverse-only rule (AC-MTE-008). Regression locks in the same milestone so
  the tag layer cannot silently break the artifact contract.
- **M3 (High) — Graph-backed fan-in source**: `SymbolFanIn` query,
  edge-backed source with tiers + hub exclusion (AC-MTE-009, AC-MTE-012),
  `FanInEvidenceSource` interface + textual wrapper in `internal/hook/mx`,
  inferred-tail reason text (AC-MTE-009).
- **M4 (Medium) — Batch-surface wiring + fallback**: SessionEnd injection —
  the sole batch site (AC-MTE-010), freshness label + absent-artifact
  textual fallback (AC-MTE-011).
- **M5 (Medium) — Query surface + CGO gate**: `--debt-fanin` mode,
  stand-in retirement (AC-MTE-013), `CGO_ENABLED=0` build + self-edge +
  textual-degrade assertions (AC-MTE-014).

Order rationale: M1-M3 are the highest-change-likelihood decisions (data
model, derivation semantics, evidence-source semantics) and land first for
review; M4-M5 are mechanical wiring and regression confirmation.

## §F. Risks

| Risk | Mitigation |
|---|---|
| `Tag.LastSeenAt` (wall-clock) leaking into the artifact | Edges are mapped from (File, Kind, Line) only; M2's determinism + golden tests RED on any leakage; mutant probe pins it. |
| Artifact growth from six kinds (this repo's tag population) | Accepted with observation (§B.1); the report already accepts six-digit edge counts; lines are short. |
| Graph-backed count diverges from grep expectations in existing tests | Parity fixture (AC-MTE-009) documents the delta direction; existing validator tests keep running against the textual source (default constructor unchanged). |
| Range join mismatches scanner line numbers (tag inside a nested closure, generated code) | Innermost-range semantics mirror `enclosingFunction`; unmatched lines fall to the self-edge — never dropped, never misattributed. |
| Hook/mx accidentally imports `internal/graph` | Interface defined at the consumer; adapter lives in `internal/graph`; `go list -deps` guard test mirrors AC-GF-016's pattern. |
| `--debt-fanin` on a fresh clone (no edges.jsonl) | Existing `--edges` not-found error message already directs to `moai graph build` — reused verbatim; no special casing. |

## §G. Preserved contracts

- `edges.jsonl` determinism + atomic write + `EdgeLess` ordering.
- Doc/code confidence semantics (REQ-GEC-*) untouched — `mx-*` edges carry NO
  `resolution`/`confidence` fields.
- Freshness gate + provenance sidecar (source fingerprints, no mtime).
- PostToolUse 500ms / SessionEnd 4000ms budget seams and their tests.
- `moai mx query` filters, sidecar schema, refresh path.
- The validator's non-P1 rules (P2/P3/P4) — untouched by the source seam.

## §H. Cross-References

- spec.md §C (REQ-MTE-001..015) · acceptance.md (AC-MTE-001..014)
- `.moai/reports/graphify-codegraph-analysis-20260901.md` § P1-2, P1-3
- SPEC-GRAPH-EDGE-CONFIDENCE-001 (predecessor layer; golden mechanism §B.5)
- Sibling card: SPEC-GRAPH-REPORT-001 (t413 — deferred refresh owner)

## §J. Drift vs older citations (this tree @ 57d2f3ae3)

- Report cited `validator.go:126` for the grep fan-in — confirmed here:
  `fanInIndex` comment at validator.go:124-125, struct at 126; P1 check at
  validator.go:270-283; threshold 3 hardcoded at validator.go:77.
- Report cited `query.go:15` (stand-in note) — confirmed verbatim.
- Report cited `graph.go:204-233` for mx-spec edges — MOVED in this tree:
  `mxSpecEdges` is now graph.go:252-277 (confidence work grew the file).
- Report cited `gate.go:1176-1216` for the quality-gate mx step — DOES NOT
  resolve here: `internal/cli/gate.go` is 259 lines in this tree (the
  graph-freshness step wiring sits at lines ~113-205). The report measured
  main @ `48239c7dc`. No mx-validation call site exists in `internal/cli`
  consuming `ValidateFiles` — `moai mx scan` constructs scanner+manager only
  (mx_scan.go:64-114, verified) — so the batch surface this SPEC wires is
  `internal/hook/session_end.go:296` (ValidateFiles, 4000ms budget), the SOLE
  non-PostToolUse validator construction site; the sync-phase strict
  enforcement rides the `/moai mx` pipeline and the sync gate's mx step.
- Report cited "internal/graph 2,355 LOC" — grown since (confidence + report
  work); not re-measured here, no decision rests on the figure.
- `symbol.Extract` walks skip `testdata/` and dot-dirs but carry no explicit
  `*_test.go` skip in this tree (symbol.go:118-130) — hub exclusion is
  therefore applied at the aggregation layer (§B.8), not assumed from
  extractor scope.
