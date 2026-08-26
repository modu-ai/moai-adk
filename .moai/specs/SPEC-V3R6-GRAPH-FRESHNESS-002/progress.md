# Progress — SPEC-V3R6-GRAPH-FRESHNESS-002

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set: Tier M 3-file set (spec.md, plan.md, acceptance.md) + research.md (explicit brief addition — provenance/triage/close-path analysis; deviation recorded in research.md §6.1) + progress.md.
- Requirements: 14 REQ (GEARS, pattern-annotated — incl. cross-cutting REQ-GFR-014 final-stamp main-reachability); 16 AC (Given-When-Then) each carrying a RED-now baseline pinned at `c9eed8ac6`; REQ↔AC traceability 100% both directions (acceptance.md §D.2). AC count (16) sits at the Tier M ceiling exactly; REQ count (14) is below its 16 ceiling.
- Plan-audit iter-1: PASS 0.856 (Tier M threshold 0.80), 0 BLOCKING — report at `.moai/reports/plan-audit/SPEC-V3R6-GRAPH-FRESHNESS-002-review-1.md`. D1-D5 textual defects corrected post-audit (D1 premise code-verified: `needsSHABackfill` closer.go:397-405 matches only empty/`(this commit)`/`(pending)`/`<pending>` — the predecessor's `pending-backfill — …` form is unrecognized, so the manual pre-close backfill prevents placeholder-freezing); D6/D7 recorded-not-fixed per auditor discretion.
- Frontmatter: 12 canonical fields + `era: V3R6` (deliberate H-override — plan-phase skeleton lacks sync_commit_sha, which would heuristically classify V3R5) + `tier: M` + `related_specs: [SPEC-V3R6-GRAPH-FRESHNESS-001]`; SPEC ID regex-validated (Bash verbatim `PASS`); ID uniqueness verified against both this worktree's and the primary checkout's `.moai/specs/` catalogs.
- Scope: 29 adopted CR round-2 findings (triage A-1..A-4) + predecessor correction-and-close; 5 follow-ups (F1-F5) / 2 rejections / 1 deferral / 33 already-fixed all fenced out in spec.md §F with citations.
- M0: codemaps provenance restamp `52f7ba135` against main-reachable `c9eed8ac6` executed pre-issuance (squash-merge orphan repair; triage-table.md §F5) — recorded as an already-executed precondition (plan.md §F M0, spec.md §B.1/§C); forward obligation encoded as REQ-GFR-014/AC-GFR-016 (no branch-HEAD restamp; final stamp main-reachable).
- Split decision: new SPEC, not an increment — predecessor is `implemented` with a frozen audit trail; M4 owns its amendment (1.1.0 → 1.2.0) and close as THIS SPEC's scope. No `depends_on` (predecessor `implemented` ≠ `completed`; circular — research.md §6.3).
- Evidence base: `.moai/reports/t279/triage-table.md` + `verify-{graph,cli,docs}.md` (read, not re-derived); `moai spec close --help` + the close implementation code-verified at `internal/spec/closer.go` this run (research.md §4 — §E.5 authoring dropped, manual SHA backfill ordered before the close, pass-with-debt non-blocking).

## §F Phase 4 Mode Selection

Logged by the orchestrator before the first run-phase spawn (2026-08-26, after Implementation Kickoff Approval: approve + autonomous progression, goal armed via `moai goal arm` under session 13b75f36 — mechanical condition: predecessor completed + sync_commit_sha 2fc4b40a6 + this SPEC implemented + targeted package tests green).

Input parameters: tier M · scope ~22 files (11 test/code + astx + docs + SPEC artifacts) · domains: Go tests, Go production policy, astx query + build tags, docs/CHANGELOG/docs-site, SPEC lifecycle (5) · language mix Go+markdown · concurrency benefit LOW (coding-heavy).

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file, semantic changes across 5 domains |
| serial | **yes** | coding-heavy Tier M; M1→M2→M3 dependency chain (M2 touches files M1 tests; M4 depends on all); Anthropic coding-task parallelism caveat |
| fanout | no | not research-heavy; write-capable parallelism forbidden |
| sweep | no | < 30 files; semantic, multi-rule; not mechanical-uniform |
| agent-team | no | not operator-requested |

Decision: serial — one manager-develop delegation per milestone (M1, M2, M3), M4 split into a manager-spec re-delegation (predecessor body corrections + manual backfill) followed by the close CLI. Boundary case: none (5 domains would tempt fanout, but every domain is write-capable coding work → serial per the tie-breaker).

## §E.2 Run-phase Evidence

### M1 — test-policy remediation (rides the M1 commit; row ledger below)

Baseline attribution (pre-change): branch `WT-t250-followup` @ `fe575d8f4` + the inherited uncommitted `codequery_test.go` edit from the 429-killed run-m1 delegation (verified: compiles, `go vet ./internal/graph/` rc=0). Pre-flight verbatim: `go build ./...` exit 0; `CGO_ENABLED=0 go build ./...` exit 0; `go test ./internal/graph/ ./internal/cli/ ./internal/hook/quality/ -count=1` → `ok internal/graph 12.234s` / `ok internal/cli 303.815s` / `ok internal/hook/quality 13.044s` rc=0; `golangci-lint run --timeout=2m` → `0 issues.` (clean baseline ⇒ any post-change finding would be NEW).

Row ledger (plan.md §F M1 #1–#10; #2/#3 inherited from the killed delegation, verified and kept):

| # | comment-id | file(s) | AC | Evidence (a command / b verbatim / c this run, this tree) |
|---|---|---|---|---|
| 1 | 3855001995 | internal/graph/citation_test.go | AC-GFR-002 | `TestCitationRegionHashMismatchReported` asserts the literal reason + a correct-hash control. RED via mutation (reason string mutated): `--- FAIL: TestCitationRegionHashMismatchReported ... reason = "MUTATION-PROBE reason", want the hash-does-not-cover-its-excerpt branch reason`; restored (`git diff internal/graph/citation.go` empty) → `ok` |
| 2 | 3855002004 | internal/graph/codequery_test.go + symbol_test.go | AC-GFR-001 | Inherited `requireCodeExtraction` skip helper (wired into SignaturesOnly / FindAndTraceCalls / PerTreeAnswers). M1 completion — 4 further unguarded CGO tests in symbol_test.go guarded (UndocumentedCallAppears, CodeLayersAreAdditive, DisagreementAllRevives, ImportEdges): required by the AC Then-clause "full package run exits ok". RED (pre-guard): `symbol_test.go:62: code-call edge A→B missing; edges=[]` (+3 analogous). GREEN: `CGO_ENABLED=0 go test ./internal/graph/ -count=1` → `ok 7.667s`; named-two → `--- SKIP: TestFileAPI_SignaturesOnly` with reason `extraction unsupported (tree-sitter unavailable or CGO disabled)` |
| 3 | 3855002013 | internal/graph/codequery_test.go | AC-GFR-005(a) | Positive `inA` ":B" assertion in TestCodeQueries_PerTreeAnswers (inherited diff: `inA := false` loop + `if !inA { t.Errorf(...vacuously...) }`); standing RED-now baseline = verify-graph #3 grep (absence-only check); GREEN in package run |
| 4 | 3855149281 | internal/graph/check_test.go | AC-GFR-005(b) | `fmt.Sprintf("gen%03d.go", i)` replaces hand-rolled pad+itoa; `grep -n 'func itoa' internal/graph/check_test.go` → 0 hits (baseline :225); `TestCheckFreshness_AgedLayerFails` → `ok` |
| 5 | 3855001906 | internal/cli/graph_check_test.go | AC-GFR-005(c) | 3 hand-interpolated pvJSON sites (:59-61, :168, :231) → `marshalCodemapsProvenance` (json.Marshal of `mx.Provenance`); `grep -nF 'pvJSON := "' internal/cli/graph_check_test.go` → 0 hits. `TestProvenanceFixtureJSONRoundTripsWindowsPaths`: `C:\Users\dev\proj` marshals+round-trips; negative control — the retired hand-interp form on the same root fails `json.Valid` (the `\U` escape does not exist) |
| 6 | 3855001928 | internal/cli/mcp_code_tools_test.go | AC-GFR-006 | `toolText` (t.Fatal wrapper) + `toolTextShape` (error-returning core) — shape before type assertion. RED via transient probe (DELETED after capture): `TestProbeUnguardedContentIndexPanics` → `panic: runtime error: index out of range [0] with length 0`. All 6 unguarded sites rerouted incl. the :80 panic-risk site; `grep -nE '[A-Za-z]+\.Content\[0\]' internal/cli/mcp_code_tools_test.go` → only :162/:164 (inside the helper); `TestToolTextShapeContract` pins empty-content → error, never panic |
| 7 | 3855001933 | internal/cli/mcp_code_tools_test.go | AC-GFR-003 | Symbol/Via asserted in TestHandleGraphFindAndTrace: a match with `Symbol=Finish` AND `Via="callee (called at)"` must exist. Observed fixture nuance recorded in-test: Finish is ALSO observed as `Via="caller (calls Println)"` (Finish calls fmt.Println) — both observations correct; the assertion pins the Run→Finish callee observation the fixture guarantees |
| 8 | 3855001948 | internal/cli/mcp_code_tools_test.go | AC-GFR-003 | `TestGraphTools_RequiredParamsAndDotDotPathRejected` table: 3 truly-empty-Arguments rows (one per handler; rejection names `file`/`query`/`symbol` — mcp-go RequireString `required argument %q not found`) + literal `..` row (`../secret.go` → rejection containing `escapes the project root`) |
| 9 | 3855002099 | internal/hook/quality/gate_graph_freshness_test.go | AC-GFR-004 | `gfStampAllLayers` stamps all three layers (codemaps provenance @HEAD → mx sidecar → edges.jsonl+meta with CURRENT fingerprints, order-sensitive); `TestGateGraphFreshness_AllLayersFreshNotice` asserts literal `graph-freshness: all layers fresh` (gate.go:1216). RED via mutation (notice mutated): `--- FAIL: ... got: "graph-freshness: MUTATION-PROBE fresh"`; restored (`git diff internal/hook/quality/gate.go` empty) → `ok` |
| 10 | 3855149237 | internal/cli/graph_refresh_test.go + graph_refresh_cli.go | AC-GFR-007 | RED (test-first, compile): `undefined: edgesRefreshClock` / `undefined: newEdgesRefreshClock` ×4 → GREEN after the seam: `var edgesRefreshClock = newEdgesRefreshClock` (production default = wall-clock constructor; the ONLY M1 production change, @MX:NOTE-tagged). `TestGraphQuery_BudgetOverrunWarns` now injects a fixed 50ms duration (1ms budget) — no real-timing reliance; `TestEdgesRefreshClockDefaultIsWallClock` pins default construction (`%p` equality with `newEdgesRefreshClock`) + monotonic wall-clock advance |

RED-now standing baselines (grep, pre-change tree, this run): `'does not cover' internal/graph/*_test.go` → 0; `'func itoa'` → :225; `'pvJSON := "'` → :59/:168/:231; `[A-Za-z]+\.Content\[0\]` → 6 sites; Symbol/Via assertions → none; empty-Arguments cases → none; `'all layers fresh'` in quality tests → 0; injection seam → none (`time.Now()`/`time.Since` inline).

Post-change verification (attribution: this run, this tree, M1 complete pre-commit):

- **E2 builds**: `go build ./...` exit 0 · `CGO_ENABLED=0 go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E2 tests**: `go test ./internal/graph/ ./internal/cli/ ./internal/hook/quality/ -count=1` → `ok internal/graph 10.983s` / `ok internal/cli 298.126s` / `ok internal/hook/quality 10.011s` rc=0. graph + hook/quality re-measured on the final tree (post symbol_test.go guards): `ok 10.706s` / `ok 14.075s`; internal/cli measured after its last edit (the #10 seam) with no subsequent cli change.
- **E2 CGO legs**: `CGO_ENABLED=0 go test ./internal/graph/ -count=1` → `ok` (AC-GFR-001 satisfied). `CGO_ENABLED=0 go test ./internal/navigator/astx/ -count=1` → `FAIL: TestPolyglot_AllFourteenGrammarsExtract/swift — Extract(swift) Supported=false, want true` — **pre-existing, M3-scoped** (row #11 / AC-GFR-012(b,c); zero astx files modified in M1 per `git status`).
- **E3 coverage**: `go test -cover ./internal/graph/` → `86.3%` · `./internal/hook/quality/` → `89.6%` · `./internal/cli/` → `79.1%` (package-level; M1's cli diff is test-additive plus a 3-line behavior-identical seam — it cannot reduce coverage, so the shortfall below the 85% bar is the package's standing structure, not M1-introduced).
- **E5 lint/vet**: `golangci-lint run --timeout=2m` → `0 issues.` (baseline-identical — 0 NEW); `go vet` on the 3 touched packages → rc=0.
- **E8**: per-row above — mutation-RED ×2 (#1, #9), transient probe-panic-RED ×1 (#6, probe deleted), compile-RED ×1 (#10), standing grep-zero baselines for the characterization rows (#3, #4, #5, #7, #8).
- **Scope discipline**: production diff confined to the finding-#10 seam in `internal/cli/graph_refresh_cli.go`; all other diffs are test files + this SPEC's artifacts. `.moai/state/` and runtime-untracked `.moai/project/graph/` untouched.

### M2 — code-policy remediation (rides the M2 commit; row ledger below)

Baseline attribution (pre-change): branch `WT-t250-followup` @ `371b8799c` (M1 landed). Pre-flight verbatim: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `go test ./internal/graph/ ./internal/cli/ ./internal/mx/ ./internal/config/ -count=1` → `ok internal/graph 5.420s` / `ok internal/cli 208.075s` / `ok internal/mx 1.882s` / `ok internal/config 1.638s` rc=0 · `golangci-lint run --timeout=2m` → `0 issues.` (clean baseline ⇒ any post-change finding would be NEW).

Row ledger (plan.md §F M2 #12–#21 + #22b):

| # | comment-id | file(s) | AC | Evidence (a command / b verbatim / c this run, this tree) |
|---|---|---|---|---|
| 12 | 3855149289 | internal/graph/check.go | AC-GFR-008(a) | BOTH remedies taken (1-line each): error path appends the computed codemaps row (`res.Layers = append(res.Layers, codemapsRep)`), doc comment rewritten to the actual partial-report contract. True-RED pre-change: `--- FAIL: TestCheckFreshness_NotComparableIsSystemError ... error-path report must carry the failing codemaps row, got 0 layers: []` → GREEN in package run; existing caller surfaces (graph_check.go exit-2 path prints the error only) unchanged |
| 13 | 3855149309 | internal/graph/check.go:254-258 | AC-GFR-008(b) | `grep -c 'var sidecarAbsentReason' internal/graph/check.go` → `0` (now `const`); comment rewritten to name the three collapsed load failures honestly |
| 14 | 3855149315 | internal/graph/check.go:309+ / graph.go / mx_query.go | AC-GFR-008(c) | Signature → `MXIndexNeedsRefresh(projectRoot string, mxIndexChangedFiles int)`. Call graph (cited): production = internal/cli/graph.go (via new `edgesRefreshNeeded`) + internal/cli/mx_query.go:114 — both inject `graph.DefaultThresholds().MXIndexChangedFiles` (behavior-identical; policy now explicit at the caller); tests = check_test.go ProbeStates (threshold 1) — no external callers exist (repo-internal package). Mutation-RED: `--- FAIL: TestMXIndexNeedsRefresh_ThresholdFromCaller ... one drifted file must NOT cross a caller-injected red line of 2`; restored → `ok` |
| 15 | 3855149325 | internal/graph/meta.go + check.go | AC-GFR-008(d) | One shared `compareSourceFingerprints(stamped, current) []string` in meta.go; `EdgesSourcesMovedFor` and `checkEdges` both call it (the two hand-rolled loops are gone). Mutation-RED: `--- FAIL: TestCompareSourceFingerprints ... moved = [mx-index specs], want [mx-index reports specs]`; restored → `ok` |
| 16 | 3855149332 | internal/graph/symbol.go:33,98 | AC-GFR-009(a) | Both bare returns wrapped: `graph: extract code edges: %w` / `graph: build doc edges: %w` (package's "graph: <op>: %w" convention); `%w` preserves errors.Is — package tests green (no test asserted the bare form) |
| 17 | 3855001978 | internal/cli/mcp_code_tools.go + _test.go | AC-GFR-009(b) | `grep -c 'jsonToolResult\|NewToolResultError' internal/cli/mcp_code_tools.go` → `0` (first pass left a doc-comment mention counting 1 — reworded); handlers return `toolErr("<tool>", err)` / `toolJSON("<tool>", data)`. B2 wire note: success data now rides StructuredContent (text = `"<tool>: ok"` fallback), error text gains the `"<tool>: "` prefix — M1 tests updated minimally via a `graphToolJSON` helper (reads StructuredContent, sessionMsgStructuredMap convention; shape check still first); error-path assertions unchanged (`toolText(t, res, true)`) |
| 18 | 3855149248 | internal/cli/graph_stamp.go + _test.go | AC-GFR-009(c) | New `TestGraphStampCmd_FSErrorsCarryNoAbsolutePath` (3 injected failures: mkdir/write/rename under `t.TempDir()`) asserts `!strings.Contains(surface, root)` + operation naming. Dev-time catch: the rename leg initially LEAKED — `os.Rename` returns `*os.LinkError`, whose own message names BOTH absolute paths; `stampFSDetail` now unwraps LinkError + PathError to the errno. Mutation-RED: `--- FAIL: .../mkdir_blocked... create directory error leaks the absolute local path` (all 3 legs); restored → `ok`. Deferred temp-cleanup report sanitized with the same helper |
| 19 | 3855149254 | internal/graph/meta.go + internal/cli/graph.go/graph_refresh_cli.go | AC-GFR-010 | New `EdgesSourcesMovedFor(projectRoot, edgesFile)` reads the SELECTED artifact's meta (`filepath.Dir(edgesFile)`); `EdgesSourcesMoved(projectRoot)` = default-path wrapper (callers unchanged). graph.go:150 now calls `edgesRefreshNeeded(projectRoot, edgesFile, th)` (extracted into graph_refresh_cli.go; mx-index probe unchanged/tree-anchored). Mutation-RED at BOTH levels: `--- FAIL: TestEdgesSourcesMovedFor_SelectedArtifact ... fresh SELECTED artifact must probe not-moved even with a stale default` and `--- FAIL: TestEdgesRefreshNeeded_FollowsSelectedEdgesArtifact ... must not trigger a refresh of the selected path`; restored → `ok` |
| 20 | 3855149357 | internal/mx/provenance.go:157-160 | AC-GFR-009(d) | Comment now states empty-output-with-nil-error is a real expected state (clean `git status --porcelain`) that treeDirty depends on; the never-happens claim is gone |
| 21 | 3855001991 | internal/config/testdata/shipped_key_inventory.yaml:380-394 | AC-GFR-011 | 5 `gate.graph_freshness.*` keys R→W with per-key reader evidence (gate.go runGraphFreshnessStep + pre_tool.go:840 for 4 keys; graph_refresh_cli.go:71 graphRefreshBudgetMS for update_budget_ms). `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders -count=1` → `ok` (anti-rot guard accepts the reclassification) |
| 22b | (residual) | internal/graph/codequery.go:17-20,153,246,323 | AC-GFR-009(e) | Swapped CR-IDs corrected (:153 dedupe comment → 3855002040, :246 absent-layer comment → 3855002033, per verify-graph Notes a); scan-window literal → `const signatureScanWindow = 8` (grep `startLine-1+8` → 0 hits); "shared by the MCP tool description" overclaim rewritten to state the const is the enforced bound and the description literal restates it (a static string cannot reference the const) |

Post-change verification (attribution: this run, this tree, M2 complete pre-commit):

- **E2 builds**: `go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `CGO_ENABLED=0 go build ./...` exit 0 · `go vet` on the 4 touched packages rc=0.
- **E2 tests**: `go test ./internal/graph/ ./internal/cli/ ./internal/mx/ ./internal/config/ ./internal/hook/quality/ -count=1` → `ok internal/graph 6.961s` / `ok internal/cli 202.625s` / `ok internal/mx 1.725s` / `ok internal/config 3.407s` / `ok internal/hook/quality 8.005s` rc=0 (hook/quality added: it consumes CheckFreshness).
- **E2 CGO leg**: `CGO_ENABLED=0 go test ./internal/graph/ -count=1` → `ok 5.847s`.
- **E3 coverage**: `go test -cover` → graph `86.4%` · mx `84.8%` · config `80.6%` · cli `79.2%`. graph ≥85; mx/config diffs are comment/testdata-only (no production logic), cli +0.1 vs M1's 79.1% — shortfalls are the packages' standing structure, not M2-introduced.
- **E5 lint**: `golangci-lint run --timeout=2m` → `0 issues.` (baseline-identical — 0 NEW).
- **E8**: true-RED ×1 (#12, run-failure on the pre-change tree); mutation-RED ×4 (#14, #15, #19 at graph+cli level, #18 — each mutant observed FAIL then restored, `grep -rn MUTATION-PROBE internal/ --include='*.go'` → 0 residue); #17's standing RED-now baseline is verify-cli #10's grep (jsonToolResult/NewToolResultError present at :26-36,:81-87 pre-change).
- **Scope discipline**: 15 files in the plan's M2 envelope; `internal/spec/` untouched (M4 territory); no template mirrors required (Go internal/ + testdata only); `.moai/state/` and untracked `.moai/project/graph/` untouched.

### M3 — astx + docs remediation (rides the M3 commit; row ledger below)

Baseline attribution (pre-change): branch `WT-t250-followup` @ `58bb7d8ba` (M2 landed). Pre-flight verbatim: `go build ./...` exit 0 · `go test ./internal/navigator/astx/ -count=1` → `ok ... 1.440s` rc=0 (cgo-on) · `CGO_ENABLED=0 go test ./internal/navigator/astx/ -count=1` → rc=1 with **8 failing test functions** (the #11 RED baseline): `--- FAIL: TestExtract_GoFixture` (`astx_test.go:95: Extract(go) Supported = false, want true (CGO build required for this test)`), `TestExtractCalls_Go` (`calls_test.go:52`), `TestExtractCalls_Python` (`:102`), `TestExtractCalls_NameBasedLanguagesCompile` (`:193` + subtests), `TestEnrichRows_HeaderDrivenJoin` (`:41`), `TestEnrichRows_MissingPathOnDiskVerifiedFalse` (`:72`), `TestEnrichRows_FileCountCeilingTruncation` (`:92`), `TestPolyglot_AllFourteenGrammarsExtract` (`:44` + subtests). `grep -c 'SPEC-V3R6-GRAPH-FRESHNESS-002' CHANGELOG.md` → 0 (B12 pre-append duplicate guard).

Row ledger (plan.md §F M3 #11, #30, #26–#29):

| # | comment-id | file(s) | AC | Evidence (a command / b verbatim / c this run, this tree) |
|---|---|---|---|---|
| 11 | 3855002141 | internal/navigator/astx/{astx,calls,enrich,polyglot}_test.go + nocgo_test.go (new) | AC-GFR-012(b,c) | `//go:build cgo` at :1 of the 4 CGO-positive files — the RED baseline measured exactly these 4 files as the failing set; `specids_test.go`'s 2 tests PASSED under `CGO_ENABLED=0` in the pre-flight (pure capability-map parsing) so it stays untagged and keeps real coverage on the !cgo leg. New `nocgo_test.go` (`//go:build !cgo`): `TestNoCGO_FallbackStubsAreUnsupported` walks all 16 registered languages + `klingon`, asserting Extract/ExtractCalls → Supported=false with nil error and no panic (REQ-NT-015 stub contract). GREEN: `CGO_ENABLED=0 go test ./internal/navigator/astx/ -count=1 -v` → rc=0 `ok ... 0.527s` with `--- PASS: TestNoCGO_FallbackStubsAreUnsupported` + `--- PASS: TestSpecIDsFromCapabilityMap_ReadsSpecIDColumn` + `--- PASS: TestSpecIDsFromCapabilityMap_AbsentFileReturnsEmpty` — 3 tests RUN (non-vacuous; the CGO-positive tests are excluded from the !cgo build, not skipped-and-failed) |
| 30 | 3855002146 | internal/navigator/astx/queries/go.scm:19 + measure_cgo.go `stripQuotes` + calls_test.go | AC-GFR-012(a) | RED (test-first, cgo-on, pre-fix): `--- FAIL: TestExtractCalls_GoRawStringImport (0.00s)` / `calls_test.go:116: raw-string import strings not captured: [{fmt .../raw.go 4}]` — the interpreted `fmt` control WAS captured, the raw-string `strings` was missing (the right stated reason). GREEN: go.scm gains `(import_spec path: (raw_string_literal) @code.import)` at :19 alongside the interpreted form at :18; `stripQuotes` extended to strip backtick pairs (raw-string imports arrive backtick-quoted — without this the capture would carry the backticks into Module). Re-run: `ok ... 0.572s` rc=0; `grep -n 'raw_string_literal' internal/navigator/astx/queries/go.scm` → `19:(import_spec path: (raw_string_literal) @code.import)` |
| 26 | 3855001858 | .moai/project/codemaps/dependencies.md:116 | AC-GFR-013 | Surgical edit, NOT a regeneration (AC-GF-012 debt fenced out): `grep -n 'hook` →' dependencies.md` → `116:- \`hook\` → config, lsp, session, mx, graph` — matches the `:77 hook --> graph` Mermaid edge |
| 27 | 3855001863 | .moai/reports/t250/m5-baseline.md:18-20 | AC-GFR-013 | Developer-local transcript path → repository-relative label ("the developer-local Claude Code session store for this repository (machine-specific path outside the repository; redacted from the committed report)"); the 8-most-recent-sessions reference preserved. `grep -n '~/.claude/projects' .moai/reports/t250/m5-baseline.md` → 0 hits |
| 28 | 3855001901 | CHANGELOG.md:46 | AC-GFR-013 | t250 entry's cumulative count corrected: `grep -n 'grows from 25 to 28' CHANGELOG.md` → `46:` (prior entry :97 records 25 after session messaging; `internal/mcp/catalog.go` mechanically = 28). Only the count corrected — this SPEC's own entry is the sync phase's to add (pre-append grep guard → 0) |
| 29 | 3855149226 | docs-site/content/ko/cli-reference/graph.md:53 | AC-GFR-013 | `grep -n '오래되었으면' docs-site/content/ko/cli-reference/graph.md` → `53:`; `grep -rn '오래했으면' docs-site/content/` → 0 hits; en/ja/zh untouched |

Post-change verification (attribution: this run, this tree, M3 complete pre-commit):

- **E2 builds**: `go build ./...` exit 0 · `CGO_ENABLED=0 go build ./...` exit 0 · `GOOS=windows GOARCH=amd64 go build ./...` exit 0 · `go vet ./internal/navigator/astx/` rc=0 (clean output).
- **E2 tests (astx both legs)**: `go test ./internal/navigator/astx/ -count=1` → `ok ... 1.818s` rc=0 (cgo-on; includes the raw-string test) · `CGO_ENABLED=0 go test ./internal/navigator/astx/ -count=1` → `ok ... 0.527s` rc=0 (AC-GFR-012(c)).
- **E5 lint**: `golangci-lint run --timeout=2m` → `0 issues.` (baseline-identical — 0 NEW). gofmt note: `gofmt -l` flags 3 edited test files, but the flagged regions are pre-existing (HEAD versions verified gofmt-dirty at the same hunks — map/struct alignment in untouched bodies); the M3 additions themselves are format-clean and no drive-by reformat was performed.
- **docs-site gate**: `hugo -s docs-site --minify --gc` → rc=0, `Total in 2879 ms`, 0 WARN/ERROR lines in the build output (warning-free).
- **E8**: true-RED ×1 (#30, run-failure observed pre-fix); #11's RED is the pre-flight CGO0 leg (8 FAILs verbatim above); #26–#29's standing RED-now baselines are verify-docs #2/#3/#8/#17.
- **Graph-freshness hygiene**: the codemaps/reports edits do NOT trip the codemaps gate (`.moai/` is not described-source); NO local `moai graph check` read was taken in M3 (a post-mutation read would need a `moai graph build` re-run first — §B stamp/build ordering); NOTHING restamped (REQ-GFR-014: the M0 stamp at `c9eed8ac6` carries to merge; AC-GFR-016 is the sync-phase obligation).
- **Scope discipline**: astx diff = 4 build-tag lines + nocgo_test.go + 1 test func + 1 go.scm line + 1 stripQuotes condition; docs = 4 single-line surgical edits; no docs-site layouts/config/shared navigation touched; no codemaps regeneration.

### M4 — predecessor SPEC correction + close (amendment by manager-spec; close by orchestrator-direct)

Row ledger (plan.md §F M4 steps 1-3):

| Step | comment-id | file(s) | AC | Evidence (a command / b verbatim / c this run, this tree) |
|---|---|---|---|---|
| 1a | 3855001890 | …-001/spec.md:87 | AC-GFR-014 | Third When-clause added: exit 2 for not-comparable system errors — names the failing operation, affected layer unmeasured, NO verdict fabricated (never `stale` for a measurement that never ran). Matches the F4 fix, graph-freshness.yml's documented 0/1/2, and docs-site ko graph.md:66. Frontmatter version 1.1.0 → 1.2.0, updated 2026-08-27 |
| 1b | 3855001874 | …-001/acceptance.md §D.1 | AC-GFR-014 | AC-GF-008 moved SHOULD → MUST. Deviation (recorded): the MUST row folds into range notation `AC-GF-001..010` (008's removal from SHOULD makes the range contiguous) — semantically identical, no mechanical consumer parses §D.1 rows (closer scans FAIL/PASS-WITH-DEBT markers only). §D.4 gate 2's existing MUST listing unchanged |
| 1c | 3855001867 | …-001/acceptance.md AC-GF-020 | AC-GFR-014 | Then-clause carries the non-Go qualifier (declaration set without non-exported filtering; Exported filtering is Go-only — `isExported` returns true unconditionally for non-Go). Signatures-only + provenance-naming assertions preserved |
| 2 | (card task 4) | …-001/progress.md §E.4 :261 | AC-GFR-015 | `sync_commit_sha: "2fc4b40a6"` — manual backfill BEFORE close (D3 exemption surface). Adjacent rationale line records the placeholder-freezing rationale (needsSHABackfill closer.go:397-405 recognizes none of the 4 placeholder forms in `pending-backfill — <prose>`; without the manual step the close would freeze the placeholder permanently). open_followups backfill row marked RESOLVED (t279 M4); the other 3 follow-ups stand |
| 3 | (close) | …-001 spec.md + progress.md | AC-GFR-015 | Orchestrator-direct invocation, verbatim: `$ moai spec close SPEC-V3R6-GRAPH-FRESHNESS-001` → `[full-close] SPEC SPEC-V3R6-GRAPH-FRESHNESS-001 — close transitions computed. / Computed transitions: / spec.md:frontmatter.status → completed / progress.md:§E.3.status → completed / progress.md:§E.5.mx_commit_sha → <derived-from-recent-mx-commit> / Commit: f32e9a3460d6a0a41e1fe79dbf320f22fb05525d` — **exit 0 on the FULL-CLOSE path (no fallback needed, no --force)**. sync_commit_sha NOT in the transition list — needsSHABackfill was false, the manual value held. Close commit subject (machine-generated, the branch's one commit without the t279 card id, as plan.md predicted): `chore(SPEC-V3R6-GRAPH-FRESHNESS-001): Mx-phase audit-ready signal + 3-phase close` |

Post-close end state (observed this run, this tree):

- `grep -m1 '^status:' …-001/spec.md` → `status: completed`
- `grep -n 'sync_commit_sha' …-001/progress.md` → `:261 - sync_commit_sha: "2fc4b40a6"` (preserved through the close) + `:298 RESOLVED (t279 M4)` follow-up row
- `grep -c '§E.5' …-001/progress.md` → 0 (no §E.5 authored — modern 4-section schema preserved; close precondition 2 satisfied via the §E.4+SHA 3-phase predicate)
- `git show --stat f32e9a346` → 2 files (spec.md 2±, progress.md +1 — the appended `mx_commit_sha: (this commit)` L60 placeholder at :306, the sanctioned chicken-and-egg form; the 5 already-discharged dogfood SPECs left the same placeholder per closer.go:193-194)
- Commit chain: `fe575d8f4` (mode log) → `371b8799c` (M1) → `58bb7d8ba` (M2) → `e70d6acde` (M3) → `bc9192411` (M4 amendment) → `f32e9a346` (close)

Amendment-commit verification (manager-spec, pre-close, verbatim): `moai spec lint …-001/spec.md` → `✓ No findings` rc=0 · exit-2 clause grep → :87 hit · AC-GF-020 qualifier cites `isExported` (function-name citation; :298-304 Go-only unicode.IsUpper re-verified this run) · `git show --stat bc9192411` → 3 files, +15/−9.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: "M0 52f7ba135 / M1 371b8799c / M2 58bb7d8ba / M3 e70d6acde / M4-amendment bc9192411 / predecessor-close f32e9a346"
run_status: "M1-M4 implemented; predecessor SPEC-V3R6-GRAPH-FRESHNESS-001 corrected (v1.2.0) and closed (completed, sync_commit_sha 2fc4b40a6) via full-close path exit 0 — no fallback, no --force"
ac_pass_count: 15             # AC-GFR-001..015, each with attribution-triple evidence in §E.2
ac_fail_count: 0
ac_pending_sync_count: 1       # AC-GFR-016 (final stamp main-reachable + no branch-HEAD restamp) — observable only on the PR head; M0 stamp c9eed8ac6 carried unmodified through M1-M4
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues at every milestone, baseline-identical
cross_platform_build:
  darwin: "pass"
  windows: "pass (GOOS=windows GOARCH=amd64 go build)"
  cgo_disabled: "pass (CGO_ENABLED=0 build + graph/astx test legs — skips, not failures)"
coverage_touched_packages:
  graph: "86.4%"
  hook_quality: "89.6%"
  cli: "79.2% (standing package structure; M1/M2 diffs test-additive + seams — cannot reduce coverage)"
  mx: "84.8% (comment-only M2 diff)"
  config: "80.6% (testdata-only M2 diff)"
total_run_phase_files: 12+16+12+3   # per-milestone: M1 / M2 / M3 / M4(+close)
delegations: "run-m1 (429-killed after finding #2 — ledger closed, state inherited) / run-m1b / run-m2 / run-m3 / spec-m4 + orchestrator-direct close"
known_deviations:
  - "M2 #12 took BOTH remedies (error-path report + honest doc) — stronger than the allowed cheaper-one"
  - "M2 #17 wire change: graph tool success JSON rides StructuredContent, text fallback '<tool>: ok' (package convention the finding asked to join)"
  - "M3 #11: specids_test.go left untagged (CGO-independent — keeps !cgo coverage non-vacuous)"
  - "M4 #1b: §D.1 MUST row folded to range notation AC-GF-001..010"
  - "close commit f32e9a346 subject is machine-generated (no t279 card id — predicted by plan.md; traceability rides the dispatch)"
```

Verification commands (final tree, this run): `go test ./internal/graph/ ./internal/cli/ ./internal/hook/quality/ ./internal/mx/ ./internal/config/ ./internal/navigator/astx/ -count=1` → all `ok` rc=0 (cli 218s the slowest); `CGO_ENABLED=0 go test ./internal/graph/ ./internal/navigator/astx/ -count=1` → `ok` ×2; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `golangci-lint run --timeout=2m` → `0 issues.`; `moai spec lint` on both SPECs → `✓ No findings`.

Gaps (explicitly NOT observed): full-suite run on a clean matrix (CI's job on the PR head); cli/mx/config coverage shortfalls are standing package structure (per-milement attribution in §E.2); AC-GFR-016 unobserved until the PR exists.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-27
sync_commit_sha: "ffd98485c"
sync_status: "complete — the single sync commit carries the in-progress → implemented → completed close + the CHANGELOG [Unreleased] ### Fixed entry + this §E.4; markdown-only (no code, no README, no docs-site)"
changelog_entry_position: "CHANGELOG.md [Unreleased] ### Fixed — first entry in the section"
docs_site_sync: "none — M3 already landed all four doc corrections (dependencies.md hook bullet, m5-baseline.md redaction, prior-entry MCP count, docs-site ko graph.md phrase); the 002 SPEC adds no user-facing surface beyond the M1-M3 deliverables (minimal-addition rule)"
frontmatter_status_transitions:
  spec_md_status: "in-progress → completed (merged into the single sync commit per the 3-phase close; no separate Mx chore commit)"
  spec_md_updated: "2026-08-26 → 2026-08-27"
  other_artifacts: "plan.md / acceptance.md untouched (no frontmatter blocks of their own to transition)"
b12_self_test_a_duplicate_grep: "pre-append `grep -c 'SPEC-V3R6-GRAPH-FRESHNESS-002' CHANGELOG.md` → 0 (rc=1, measured this run before the edit); post-append → 1 (rc=0) — exactly this entry, no duplicates"
b12_self_test_b_ac_count: "distinct AC-GFR tokens in acceptance.md → 16 (AC-GFR-001..016, non-zero and plausible — anchored on the AC-ID token); §E.3 records 15 PASS + 1 pending-sync (AC-GFR-016)"
b12_self_test_c_path_verification: "every path cited in the CHANGELOG entry verified by ls this run, rc=0: internal/graph/{check,meta,symbol}.go + citation_test.go, internal/cli/{mcp_code_tools,graph_stamp,graph_refresh_cli}.go, internal/hook/quality/gate.go, internal/navigator/astx/nocgo_test.go + queries/go.scm, internal/config/testdata/shipped_key_inventory.yaml, both SPEC dirs, .moai/project/codemaps/dependencies.md, docs-site/content/ko/cli-reference/graph.md"
sync_verification:
  spec_lint: "`moai spec lint .moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-002/spec.md` → '✓ No findings — all SPEC documents are valid' rc=0 — re-run AFTER the frontmatter close (status: completed), this run, this tree"
  graph_check_final: "go run ./cmd/moai graph check → 3×fresh exit 0 (codemaps described-source-diff 27 threshold 40 fresh / mx-index inventory-content-diff 0 threshold 1 fresh / edges source-fingerprint-mismatch 0 threshold 0 fresh) — this run, this tree, after the edges rebuild narrated below"
open_followups:
  - "AC-GFR-016 (final codemaps stamp main-reachable + no branch-HEAD restamp) — observable only on the delivering PR head (manager-git's PR); local halves verified this sync: .moai/project/codemaps/provenance.json commit_sha c9eed8ac69a48ac42f74740a0806843001757284 (the M0 stamp, main-reachable) unchanged, and no stamp was refreshed against a branch-local HEAD"
  - "sync_commit_sha backfill — the orchestrator replaces the placeholder above with the real sync-commit SHA in the immediately following commit (D3)"
```

Sync-phase graph-check narrative (first measurement, divergence, and recovery — all this run, this tree):

1. First `go run ./cmd/moai graph check` → **exit 1**: codemaps fresh (27 < 40) · mx-index fresh (0 < 1) · **edges stale** (`value=1 threshold=0 — source set(s) moved: reports`) — diverging from the dispatch's expected 3×fresh.
2. Cause attribution: the edges stamp (`edges.meta.json` commit_sha `c1ed9baae…`, generated_at 2026-08-26T16:08:08Z, dirty=false — the orchestrator's final-tree verification build) predates the CR-thread-resolution work: 5 files under `.moai/reports/t279/` (thread-replies.tsv, triage-table.md, resolve-only.sh, thread-disposition.tsv, resolve-threads.sh) were written after the stamp. `.moai/reports/` is an edges described source, so the fingerprint moved. All 5 are gitignored/untracked session artifacts — none rides the PR, and the codemaps gate is unaffected (`.moai/` is not codemaps described-source; M3's hygiene note held).
3. Recovery: `go run ./cmd/moai graph build` → rc=0 (edges layer regenerated, edge_count 179,256). REQ-GFR-014 compliance re-verified post-build: codemaps provenance `c9eed8ac6…` unchanged (the build does not touch the codemaps stamp); edges meta carries HEAD `c1ed9baae…` (untracked artifact, regenerated by CI's own bootstrap on the PR head).
4. Re-measurement: `go run ./cmd/moai graph check` → **3×fresh exit 0** (values in `sync_verification.graph_check_final` above).

Sync-phase evidence-attribution: all commands above executed this run against worktree `WT-t250-followup` @ `c1ed9baae` (pre-sync-commit HEAD). Gaps: no CI verdict exists yet (branch unpushed; the PR is manager-git's); AC-GFR-016's PR-head half remains unobserved until that PR exists.
