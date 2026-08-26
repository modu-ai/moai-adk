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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
