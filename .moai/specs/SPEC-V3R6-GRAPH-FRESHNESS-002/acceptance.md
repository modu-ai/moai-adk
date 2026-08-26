# Acceptance Criteria — SPEC-V3R6-GRAPH-FRESHNESS-002

Given-When-Then scenarios, binary-testable. REQ↔AC traceability at §D.2 is 100% in both directions. Every AC carries its RED-now baseline pinned at tree `c9eed8ac6` (the verify reports are the standing evidence; the two-cell adoption discipline — RED-now cell + green path — is satisfied per AC by the Baseline line). The `AC-GFR-*` token set has zero pre-implementation hits in the tree by construction.

## §D. Acceptance Criteria

### AC-GFR-001 — CGO skip guards in graph query tests (REQ-GFR-001)

- **Given** the worktree toolchain with `CGO_ENABLED=0`
- **When** `CGO_ENABLED=0 go test ./internal/graph/ -run 'TestFileAPI_SignaturesOnly|TestFindCodeAndTraceCalls' -count=1` runs, then `CGO_ENABLED=0 go test ./internal/graph/ -count=1`
- **Then** the two named tests report SKIP with a reason naming extraction-unsupported (not FAIL), and the full package run exits ok with zero failures attributable to missing CGO
- **Baseline** (c9eed8ac6): `codequery_test.go:42-45,146,175` — `if err != nil { t.Fatal(err) }`, no skip guard (verify-graph #2)

### AC-GFR-002 — Citation hash-inconsistency branch tested (REQ-GFR-002)

- **Given** a citation whose `RegionHash` is populated but disagrees with the sha256 of its own excerpt, with the cited file content otherwise intact
- **When** the citation resolves
- **Then** `Resolution.Reason` equals "citation region hash does not cover its excerpt" (the `citation.go:148-152` branch), exercised by a named test that asserts the reason string
- **Baseline**: `grep -n 'does not cover' internal/graph/*_test.go` → 0 hits; the existing edited-file test exercises the content-search path, not this branch (verify-graph #1, Notes a)

### AC-GFR-003 — MCP handler coverage additions (REQ-GFR-002)

- **Given** the three MCP code-tool handlers with (a) empty/missing required `Arguments` and (b) a literal `..` traversal path
- **When** the new table-driven test runs
- **Then** each rejection case asserts an error result naming the rejected input, and the FindAndTrace happy path asserts the parsed matches' `Symbol` and `Via` content (not merely match-count)
- **Baseline**: no test passes empty Arguments to any handler; no literal `..` case (symlink test covers a different vector); Symbol/Via asserted nowhere (verify-cli #5, #6)

### AC-GFR-004 — All-layers-fresh notice asserted (REQ-GFR-002)

- **Given** a fixture project whose codemaps, mx-index, and edges layers are all stamped and in sync
- **When** the graph-freshness gate step runs
- **Then** the test asserts the literal all-fresh notice observed at `gate.go:1216` ("graph-freshness: all layers fresh"), not merely the generic "graph-freshness" substring
- **Baseline**: fixture stamps only the codemaps dir; assertion is `strings.Contains(out, "graph-freshness")` only (verify-cli #13)

### AC-GFR-005 — Vacuous-guard + fixture serialization hygiene (REQ-GFR-003)

- **Given** the tree-A/B fixture in `codequery_test.go` and the provenance fixtures in `graph_check_test.go`
- **When** the updated tests run and the files are inspected
- **Then** (a) tree A's answer set carries a positive content assertion (a known tree-A edge present), so an empty answer cannot pass vacuously; (b) `grep -n 'func itoa' internal/graph/check_test.go` → 0 hits — fixture names built via `strconv`/`fmt`; (c) all three provenance fixture sites (former :59-61, :168, :231) build JSON via `json.Marshal` — `grep -F '"schema_version": 1,\\n' internal/cli/graph_check_test.go` hand-interpolation form → 0 hits — and a Windows-style backslash path fixture round-trips
- **Baseline**: absence-only ":B" check; itoa at :225 + padding at :199; hand-interpolated `pvJSON` at three sites — Windows-path JSON breakage is a real defect (verify-graph #3, #13; verify-cli #1)

### AC-GFR-006 — Result-shape helper replaces unguarded indexing (REQ-GFR-003)

- **Given** `mcp_code_tools_test.go`'s six unguarded `Content[0]` indexing sites (:80, :113, :125, :153, :178, :198 — receiver variables `res`/`resA`/`resB`/`findRes`/`traceRes`, so a literal `res.Content[0]` grep would match only a subset)
- **When** the helper is introduced and the tests run
- **Then** every site routes through a shape-check helper asserting `len(res.Content) > 0` and the expected `IsError` state before the type assertion; an error-result input with empty Content fails the test with a message, not a panic; `grep -nE '[A-Za-z]+\.Content\[0\]' internal/cli/mcp_code_tools_test.go` (regex over ANY receiver variable — closes the rename mutant that a literal-`res` grep admits) → hits only inside the helper
- **Baseline**: five one-value assertions + the :80 site — the :80 panic risk confirmed (verify-cli #4 + Minor-2a resolution)

### AC-GFR-007 — Budget-overrun determinism (REQ-GFR-004)

- **Given** a new injection seam in `refreshEdgesArtifact` (graph_refresh_cli.go) and the budget-overrun test
- **When** the test runs with an injected duration exceeding the configured budget
- **Then** the overrun warning fires deterministically; the test no longer relies on "any real refresh exceeds it"; the seam's production default is wall-clock measurement — verified by a named assertion in the new test's file that the UN-injected seam constructs the wall-clock measurer (default-construction check), so CLI behavior is unchanged
- **Baseline**: no injection seam; test comment states the wall-clock assumption (verify-cli #19)

### AC-GFR-008 — internal/graph contract honesty (REQ-GFR-005)

- **Given** `check.go`/`meta.go` after remediation
- **When** inspected and the package tests run
- **Then** (a) the checkCodemaps error path and the doc comment agree — layer reports returned on error, or the "complete report (every layer present, even on failure paths)" claim removed; (b) `grep -n 'var sidecarAbsentReason' internal/graph/check.go` → 0 hits (immutable const) and the comment no longer says "distinguishes the three" inaccurately; (c) `MXIndexNeedsRefresh` no longer calls `DefaultThresholds()` internally — the threshold arrives from the caller (signature or variant), with existing callers updated; (d) one shared fingerprint-comparison helper is called by both `EdgesSourcesMoved` (meta.go) and `checkEdges` (check.go) — the two hand-rolled loops are gone
- **Baseline**: verify-graph #14 (return-before-append + stale comment), #15 (mutable var), #16 (hardcoded default), #18 (two copies)

### AC-GFR-009 — Error-surfacing consistency (REQ-GFR-006)

- **Given** the remediated production files
- **When** inspected and tested
- **Then** (a) `symbol.go`'s two bare error returns (former :31-34, :94-97) wrap with operation context and `%w`; (b) `grep -n 'jsonToolResult\|NewToolResultError' internal/cli/mcp_code_tools.go` → 0 hits — handlers return via the package's `toolJSON`/`toolErr`; (c) a `graph stamp` failure path surfaces an operation-naming error without the absolute temp/target path — a test injects a failure under a `t.TempDir()` root and asserts `!strings.Contains(errOutput, tmpDir)`; (d) the `gitOut` comment no longer claims empty-output-with-nil-error never happens; (e) the swapped CR-ID comments at codequery.go :153/:244 sit on the correct fixes, the scan-window literal `8` at :323 is a named constant, and the codequery.go:17-19 "shared by the MCP tool description" claim matches reality alongside the mcp_server.go:506 literal
- **Baseline**: verify-graph Notes (a) + residuals; verify-cli #8, #10, #20

### AC-GFR-010 — Selected `--edges` artifact freshness (REQ-GFR-007)

- **Given** a fixture where the default edges artifact is fresh and a custom `--edges` artifact is stale (source fingerprints moved), and the symmetric arrangement
- **When** the refresh/build path evaluates with `--edges <custom>`
- **Then** the refresh-needed decision follows the selected artifact (stale custom → refresh; fresh custom + stale default → no refresh of the selected path); the decision reads the selected artifact's meta/provenance
- **Baseline**: `graph.go:150` — `EdgesSourcesMoved(projectRoot)` probes the default artifact while the refresh targets `edgesFile` (verify-cli #21)

### AC-GFR-011 — Shipped key inventory reclassification (REQ-GFR-008)

- **Given** `internal/config/testdata/shipped_key_inventory.yaml`
- **When** the five `graph_freshness` entries (former :380-394) are read after remediation and the internal/config inventory test runs
- **Then** each entry carries `class: W` with evidence naming the three production readers (`gate.go:167`, `pre_tool.go:840`, `graph_refresh_cli.go:53`), and the inventory-consistency test passes
- **Baseline**: all five `class: R` / `evidence: none` (verify-cli #12)

### AC-GFR-012 — astx raw-string capture + CGO gating (REQ-GFR-009, REQ-GFR-010)

- **Given** a Go fixture importing a package via a raw string literal path, and the astx test files
- **When** extraction runs and the tagged test build executes
- **Then** (a) the import edge is captured — `go.scm`'s import_spec path pattern includes a `raw_string_literal` alternative, proven by a named test; (b) `grep -l 'go:build cgo' internal/navigator/astx/*_test.go` returns the CGO-positive files and a `!cgo` fallback test exists; (c) `CGO_ENABLED=0 go test ./internal/navigator/astx/ -count=1` exits ok (fallback exercised, no FAIL)
- **Baseline**: 5 untagged test files, `calls_test.go:51-52` fatals on Supported=false; go.scm:18 carries only `interpreted_string_literal` (verify-docs #9, #10)

### AC-GFR-013 — Documentation accuracy and redaction (REQ-GFR-011)

- **Given** the four documentation targets
- **When** each is read after remediation and the docs-site build runs
- **Then** the `dependencies.md` hook summary bullet lists `graph` (matching the `:77` Mermaid `hook --> graph` edge); `grep -n '~/.claude/projects' .moai/reports/t250/m5-baseline.md` → 0 hits (repository-relative label in place); `CHANGELOG.md`'s graph entry reads "25 to 28"; `docs-site/content/ko/cli-reference/graph.md:53` reads "오래되었으면"; and `hugo -s docs-site --minify --gc` exits 0 (content-only change — build gate per repo docs convention)
- **Baseline**: verify-docs #2, #3, #8, #17 (graph absent from bullet; local username path committed verbatim; "21 to 24"; invalid form present)

### AC-GFR-014 — Predecessor SPEC body corrections (REQ-GFR-012)

- **Given** SPEC-V3R6-GRAPH-FRESHNESS-001's artifacts after the M4 manager-spec re-delegation
- **When** read and linted
- **Then** `spec.md` REQ-GF-004 carries a third When-clause defining exit 2 for not-comparable system errors (names the failing operation; reports the affected layer with no verdict — neither fresh, stale, nor absent); `acceptance.md` §D.1's MUST row lists AC-GF-008 and the SHOULD row no longer does; AC-GF-020's Then-clause carries the non-Go declaration-set qualifier (non-Go responses return the extracted declaration set without non-exported filtering); `spec.md` frontmatter reads `version: "1.2.0"` with `updated` advanced; `moai spec lint .moai/specs/SPEC-V3R6-GRAPH-FRESHNESS-001/spec.md` reports no findings
- **Baseline**: verify-docs #4, #5, #7 (no language qualifier; SHOULD/closure-gate contradiction; only exit 0/1 defined)

### AC-GFR-015 — Predecessor close with observed evidence (REQ-GFR-013)

- **Given** the M1-M3 remediations landed, the M4 body corrections committed, and predecessor `§E.4 sync_commit_sha` manually backfilled to `2fc4b40a6` BEFORE the close (D3 exemption surface; no `§E.5` section authored — the modern 4-section schema is preserved)
- **When** `moai spec close SPEC-V3R6-GRAPH-FRESHNESS-001` runs (attempt 1, default mode) and — **where the observed output shows a precondition failure (exit 1)** — `moai spec close SPEC-V3R6-GRAPH-FRESHNESS-001 --backfill-only` runs
- **Then** whichever path succeeded, its verbatim CLI output (command, exit code, stdout/stderr, and the actually-generated close-commit subject) is recorded in THIS SPEC's `progress.md` §E.2; predecessor `spec.md` frontmatter `status` reads `completed`; predecessor `progress.md` §E.4 `sync_commit_sha` still reads `2fc4b40a6` (the manual backfill preceded the close because `needsSHABackfill` does not recognize the prose placeholder — without it the close would succeed and freeze `pending-backfill — <prose>` as the permanent value); predecessor `progress.md` contains no `§E.5` heading. The path choice is evidenced by the recorded output, not asserted in advance
- **Baseline**: close preconditions vs predecessor state, code-verified at `internal/spec/closer.go` (research.md §4): precondition 2 satisfied via §E.4 + populated SHA; `hasGenuinePassWithDebtVerdict` scans acceptance.md only — grep of the predecessor acceptance.md shows 0 genuine PASS-WITH-DEBT/FAIL markers, so `§E.3 ac_pass_with_debt_count: 2` does not block

### AC-GFR-016 — Final-stamp main-reachability (REQ-GFR-014)

- **Given** the delivering PR's final tree (all M1-M4 + sync work landed; M1-M3 described-source churn ~15-20 files, within threshold 40)
- **When** the codemaps provenance stamp is read and tested — `git merge-base --is-ancestor <stamp-sha> origin/main` — and `moai graph check` runs on the PR head
- **Then** the stamp names a main-reachable commit (`c9eed8ac6` unless a recorded, main-reachable refresh was required); the ancestor test exits 0; `moai graph check` exits 0 with all three layers fresh; and no commit in the PR's history restamped against a branch-local HEAD (`git log -p -- .moai/project/codemaps/provenance.json` shows every `commit_sha` value main-reachable)
- **Baseline**: M0 delivered exactly this state — commit `52f7ba135`, measured chain mx scan → build → stamp → rebuild → check with 3/3 layers fresh exit 0 (triage-table.md §F5); the `0d15864ae90b` orphaning this guards against is the observed failure (lane-4 #1662)

## §D.1 Severity Classification

| Severity | ACs | Meaning |
|---|---|---|
| MUST | AC-GFR-001..016 | Every adopted finding is binary-verified scope; comment/wording rows are file-state checks with grep/read evidence. Failure blocks close |

All rows MUST: the remediation's unit of delivery IS the adopted-finding set; a missing cosmetic fix is a missing adopted item, not acceptable debt.

## §D.2 Traceability Matrix

| REQ | ACs |
|---|---|
| REQ-GFR-001 | AC-GFR-001 |
| REQ-GFR-002 | AC-GFR-002, AC-GFR-003, AC-GFR-004 |
| REQ-GFR-003 | AC-GFR-005, AC-GFR-006 |
| REQ-GFR-004 | AC-GFR-007 |
| REQ-GFR-005 | AC-GFR-008 |
| REQ-GFR-006 | AC-GFR-009 |
| REQ-GFR-007 | AC-GFR-010 |
| REQ-GFR-008 | AC-GFR-011 |
| REQ-GFR-009 | AC-GFR-012 |
| REQ-GFR-010 | AC-GFR-012 |
| REQ-GFR-011 | AC-GFR-013 |
| REQ-GFR-012 | AC-GFR-014 |
| REQ-GFR-013 | AC-GFR-015 |
| REQ-GFR-014 | AC-GFR-016 |

Every AC-GFR-001..016 appears exactly once on the right-hand side — 100% AC→REQ coverage; every REQ-GFR-001..014 appears at least once — 100% REQ→AC coverage.

## §D.3 Indirect Verification (where direct runtime assertion is insufficient)

- **Comment/wording corrections** (AC-GFR-008(b), AC-GFR-009(d)(e), parts of AC-GFR-013): file-state checks — grep for the old form returning 0 + read of the new form — are the accepted evidence; a comment has no runtime observable.
- **No absolute-path leak** (AC-GFR-009(c)): negative `strings.Contains` against the fixture's `t.TempDir()` root on an injected failure — the negation is the observable.
- **CGO legs** (AC-GFR-001, AC-GFR-012(c)): `CGO_ENABLED=0 go test <pkg>` exit + per-test SKIP verdicts are the direct evidence; no cross-compilation of test binaries is claimed.

## §D.4 Closure Gates (Definition of Done)

1. All 16 MUST ACs PASS with observed evidence (command + verbatim output recorded in progress.md §E.2).
2. Targeted package tests green — `go test ./internal/graph/ ./internal/graph/symbol/ ./internal/cli/ ./internal/hook/quality/ ./internal/navigator/astx/ ./internal/config/ ./internal/mx/ -count=1` — plus the two `CGO_ENABLED=0` legs; **never** `go test ./...` locally (CI judges the full suite on the delivering PR).
3. `go vet` clean on every touched package.
4. `hugo -s docs-site --minify --gc` exit 0 (M3 docs-site leg).
5. `moai spec lint` clean on BOTH this SPEC's spec.md and the amended predecessor spec.md.
6. No new entries in go.mod (constraint §E verified by diff).
7. Delivering-PR final stamp reachability: `git merge-base --is-ancestor <stamp-sha> origin/main` exits 0 on the PR head, and `moai graph check` exits 0 there (AC-GFR-016).

## §D.5 Edge Cases

- Close-tool lock contention (`.moai/state/spec-close-<SPEC-ID>.lock` held by a concurrent attempt): retry once after the lock releases; a second contention is a blocker report, not `--force`.
- The close tool's own commit subject may differ from the plan's suggested subject — record what the tool actually produced; AC-GFR-015 binds on the recorded output and end state, not the subject.
- The manual pre-close backfill is the only sanctioned route for `sync_commit_sha` (`2fc4b40a6`): without it the close would succeed and freeze the prose placeholder as the permanent §E.4 value — `needsSHABackfill` (closer.go:397-405) matches only empty/`(this commit)`/`(pending)`/`<pending>`, not the predecessor's `pending-backfill — …` form (research.md §4).
- A test file receiving both M1 changes and M3 build tags must keep the tags file-scoped (`//go:build cgo` lines at top), not function-scoped noise.
