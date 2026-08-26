# cli-slice verification — tree c9eed8ac6

Slice: internal/cli + internal/mcp + internal/mx + internal/hook + config testdata (32 findings).
Dump source: `.moai/reports/t250/cr-round2-comments.md`. All evidence cited against the current worktree tree (c9eed8ac6, branch WT-t250-followup).

| # | comment-id | cited file:line | current file:line | verdict | one-line what-it-asks | evidence (≤3 lines) |
|---|---|---|---|---|---|---|
| 1 | 3855001906 | internal/cli/graph_check_test.go:62 | graph_check_test.go:59-61 (+ :168, :231) | still-valid | Serialize provenance fixtures with json.Marshal, not string interpolation | `pvJSON := "{\n  \"schema_version\": 1,\n  \"tree_root\": \"" + root +` (L59-60); same hand-interpolation at L168 and L231 — Windows backslash paths break the JSON |
| 2 | 3855001912 | internal/cli/graph_check_test.go:135 | internal/graph/check.go:422-447 | already-fixed | Stage the 41 generated files so the stale check counts them | `gitDiffNameCount` runs `git ls-files --others --exclude-standard` (L436) — "A file untracked at HEAD counts too... `git diff <commit>` alone would silently skip it" (L420-421); no staging needed, untracked files count |
| 3 | 3855001919 | internal/cli/graph_stamp.go:76 | internal/cli/graph_stamp_test.go:35-88 | already-fixed | Add command-level tests for `stamp codemaps` (valid output, replacement, temp cleanup) | `TestGraphStampCmd_ValidAndReplacement` executes newGraphStampCmd, asserts OK line + CommitSHA/GeneratedBy, replaces a broken stamp, asserts `target + ".tmp"` gone (L84-87); residue: failed-rename cleanup sub-case untested (see Notes) |
| 4 | 3855001928 | internal/cli/mcp_code_tools_test.go:67 | :113, :125, :153, :178, :198 | still-valid | Check result shape (nil/IsError/Content len) before the type assertion | No `toolText` helper exists; five one-value assertions remain, e.g. `body := res.Content[0].(mcp.TextContent).Text` (L113); `res.IsError` unasserted at these sites |
| 5 | 3855001933 | internal/cli/mcp_code_tools_test.go:105 | :179-190 | still-valid | Fix struct tags to wire names and assert Symbol/Via match content | Tags now match the wire (codequery.go:123/127 declares `json:"Symbol"`/`json:"Via"` — tag half moot), but the test still only checks `len(parsed.Matches) == 0` (L188) — Symbol/Via asserted nowhere |
| 6 | 3855001948 | internal/cli/mcp_code_tools_test.go:127 | whole file | still-valid | Table-driven tests for missing-required-parameter branches + `..` path case | Tests present: RejectsSymlinkEscape (:60), HonorProjectRoot (:90), FileAPI (:144), FindAndTrace (:169) — none passes empty Arguments to the 3 handlers; no literal `..` path case (symlink case covers a different escape vector) |
| 7 | 3855001953 | internal/cli/mcp_code_tools.go:50 | :28, :44, :64 | already-fixed | Resolve root via resolveToolProjectRoot(req) + projectRootOption() registrations | `root, err := resolveToolProjectRoot(req)` in all three handlers (:28/:44/:64); registrations use `projectRootOption()` (mcp_server.go:489/497/506); `TestGraphTools_HonorProjectRoot` (:90-140) covers distinct roots + invalid-root rejection |
| 8 | 3855001962 | internal/cli/mcp_code_tools.go:59 | :23-78 | still-valid | Propagate ctx into graph.FileAPI/FindCode/TraceCalls (or mark `_ ctx`) | Handlers accept `ctx context.Context` and never use it; calls are `graph.FileAPI(root, rel)` / `graph.FindCode(root, query)` / `graph.TraceCalls(root, symbol, depth)` — no ctx parameter, not `_`-named |
| 9 | 3855001968 | internal/cli/mcp_code_tools.go:59 | :63 | already-fixed | Use `req.GetInt("depth", 1)` with an explicit default | `depth := req.GetInt("depth", 1)` (L63) — exactly the proposed form; no discarded RequireInt error |
| 10 | 3855001978 | internal/cli/mcp_code_tools.go:78 | :26-36, :81-87 | still-valid | Replace jsonToolResult/NewToolResultError with the package's toolJSON/toolErr | `jsonToolResult` still defined and used (:36/:52/:72/:81-87); handlers return `mcp.NewToolResultError(err.Error())` (:26 etc.) while mcp_server.go's other tools use `toolErr(...)`/`toolJSON(...)` (mcp_server.go:585-784) |
| 11 | 3855001981 | internal/cli/mx_scan.go:94 | :108 | already-fixed | Pass projectRoot to ScanInventory so keys are provenance-root relative | `Provenance: mx.StampMXScan(projectRoot, s.ScanInventory(projectRoot))` (:108) with CR-comment (:95-100); out-of-project scan roots now rejected outright (:58-60) |
| 12 | 3855001991 | internal/config/testdata/shipped_key_inventory.yaml:394 | :380-394 | still-valid | Reclassify the 5 graph_freshness keys R→W with evidence: reader | All five entries still `class: R` / `evidence: none` (:380-394); production readers verified present: gate.go:167, pre_tool.go:840, graph_refresh_cli.go:53 |
| 13 | 3855002099 | internal/hook/quality/gate_graph_freshness_test.go:132 | :113-132 | still-valid | Stamp all three layers in the fixture and assert the "all layers fresh" notice | Fixture creates only the codemaps dir (:115-118); assertion is just `strings.Contains(out, "graph-freshness")` (:129) — comment (:126-128) reframes to notice-emission-only; `"graph-freshness: all layers fresh"` (gate.go:1216) is asserted by no test |
| 14 | 3855002107 | internal/mcp/catalog_test.go:16 | :13, :20-21 | already-fixed | Name the catalog-size invariant as a constant | `const wantCatalogSize = 28` (:13) with invariant doc comment, used in both the assertion and the failure message (:20-21) |
| 15 | 3855002115 | internal/mx/provenance.go:53 | internal/graph/check.go:153-172, :389-416 | already-fixed | Validate provenance scope (DescribedRoots containment; TreeRoot equality) before use | DescribedRoots validated per root via `validateDescribedRoot` (empty/absolute/lexical `..`/symlink escape all rejected, check.go:389-416) with comment citing "CR round-2 3855002115 / 3855149192"; TreeRoot equality deliberately NOT required for codemaps — tracked-artifact rationale documented at check.go:156-161 (wrong-tree covered for untracked layers at check.go:231-237) |
| 16 | 3855002126 | internal/mx/scanner.go:77 | :67-80 | already-fixed | Check extension/comment-prefix before os.ReadFile | Extension check first: `ext := ...; prefix := GetCommentPrefix(ext); if prefix == "" { return nil, nil }` (:70-75), `os.ReadFile` after (:77) — comment cites "CR round-2 3855002126" |
| 17 | 3855002133 | internal/mx/scanner.go:328 | :339 | already-fixed | Exclude only parent-directory escapes, keep `..generated.go` children | `if err != nil \|\| rel == ".." \|\| strings.HasPrefix(rel, ".."+string(filepath.Separator))` (:339) — the exact proposed form; ScanInventory doc comment (:332-334) records the `..generated.go` guarantee |
| 18 | 3855149230 | internal/cli/graph_check.go:115 | :71-79, :136-146 | already-fixed | Propagate gate.yaml load errors as exit 2, keep defaults only for absent config | `graphCheckThresholds` pre-validates a present gate.yaml and errors (:140-146); RunE wraps into `exitCodeError{code: exitSystemError}` (:77-78); regression test `TestGraphCheckCmd_MalformedGateYamlExitsTwo` (graph_check_test.go:192-219) |
| 19 | 3855149237 | internal/cli/graph_refresh_test.go:185 | :152-185 | still-valid | Make the budget-overrun test deterministic via an injected duration | No injection seam exists: `refreshEdgesArtifact` measures `time.Since(start)` (graph_refresh_cli.go:24,:40); the test relies on the real refresh exceeding the 1ms budget (comment :155 "any real refresh exceeds it") |
| 20 | 3855149248 | internal/cli/graph_stamp.go:68 | :46-68 | still-valid | Sanitize filesystem errors (paths in cmDir/tmp/target) at the CLI boundary | Errors still wrapped and returned verbatim: `fmt.Errorf("mkdir codemaps: %w", err)` (:47), write (:57), rename (:68) — the underlying os *PathError* carries the absolute path to the user |
| 21 | 3855149254 | internal/cli/graph.go:158 | :150 | still-valid | Evaluate freshness against the selected --edges artifact, not the default | `if refreshNeeded := graph.EdgesSourcesMoved(projectRoot) \|\| graph.MXIndexNeedsRefresh(projectRoot); refreshNeeded` (:150) — EdgesSourcesMoved still takes only projectRoot (probes the default artifact's meta), while the refresh itself targets `edgesFile` (:151) |
| 22 | 3855149264 | internal/cli/mcp_server.go:482 | :505 | already-fixed | Declare depth with mcp.WithInteger | `mcp.WithInteger("depth", mcp.Description("Traversal depth in hops (default 1, capped at 8)."))` (mcp_server.go:505) |
| 23 | 3855149345 | internal/hook/quality/gate.go:1186 | :1184-1190 | deferred-by-design | Emit an explicit unconfigured (nil GraphFreshness) skip notice | Nil branch still returns `"", false, ""` but with recorded rationale (:1185-1189 "absent from every production path... Preserve the pre-existing silent-pass contract"); SPEC record: progress.md §E.2 AC-GF-006 — "Deviation recorded: the UNCONFIGURED (nil) posture stays silent — the pre-existing unknown-project silent-pass contract (TestQualityGate_Run_UnknownProjectPasses) is preserved; production paths always populate the config" |
| 24 | 3855149353 | internal/mx/provenance.go:125 | :124-126 + refresh.go:197-199 | already-fixed | Add IsRegular guards to both walks (FIFO/symlink hang) | provenance.go: `if !info.Mode().IsRegular() { return nil }` (:124-126); refresh.go walkScanFiles: same guard (:197-199) — both cite "CR round-2 3855001937" |
| 25 | 3855149357 | internal/mx/provenance.go:156 | :157-158 | still-valid | Fix gitOut doc claim that empty output with nil error never happens | Comment unchanged: `// gitOut runs a git command in dir and returns trimmed stdout. Empty output` / `// with a nil error never happens; errors return "" (fail-open by callers).` (:157-158) — a clean `git status --porcelain` returns exactly that |
| 26 | 3855149371 | internal/mx/provenance.go:190 | :199 | still-valid | Anchor the stamp on the content fingerprint when git is unavailable (empty GitHead) | `pv.CommitSHA = GitHead(projectRoot)` (:199) with no empty-SHA fallback; checkCodemaps maps the result to `VerdictAbsent` / "clean stamp carries no commit sha — freshness-unjudgeable" (check.go:194-198); no deferral record in progress.md §E.2 Gaps |
| 27 | 3855149384 | internal/mx/provenance.go:197 | :205-207 | already-fixed | Add coverage for StampCodemaps (and consider dropping its error result) | Command-level coverage exists: graph_stamp_test.go:35-88 exercises StampCodemaps end-to-end (asserts GeneratedBy/CommitSHA, replacement, temp cleanup); note: the optional simplification was not taken — `StampCodemaps` still returns `(*Provenance, error)` with an always-nil error |
| 28 | 3855149390 | internal/mx/refresh.go:73 | :76-87 | already-fixed | Carry the previous digest forward for a transiently unreadable file | `if old, ok := oldInventory[rel]; ok { newInventory[rel] = old; s.RecordError(...) }` (:82-85) — comment cites "Transiently unreadable ≠ vanished (CR round-2 3855001935)"; vanished-classification loop (:110-116) can no longer see it |
| 29 | 3855149392 | internal/mx/refresh.go:84 | :84, :97 | already-fixed | Report scanner errors via a Scanner method, not the unexported field | `s.RecordError(fmt.Sprintf("refresh: unreadable...))` (:84) and `s.RecordError(fmt.Sprintf("refresh: scan %s: %v"...))` (:97); `func (s *Scanner) RecordError(msg string)` exists (scanner.go:325-327) with CR-citing doc |
| 30 | 3855149395 | internal/mx/refresh.go:123 | :129 | already-fixed | Remove the dead newSum variable | `if _, stillExists := newInventory[rel]; !stillExists {` (:129) — presence-flag-only binding, no `newSum`, no `_ =` discard |
| 31 | 3855149412 | internal/mx/refresh.go:158 | :170-176 | already-fixed | Skip unreadable entries instead of failing the whole refresh | Walk error path: `if info != nil && info.IsDir() { return filepath.SkipDir }; return nil` (:172-175) — comment cites "CR round-2 3855001941"; doc contract (:166-168) records the fail-open posture |
| 32 | 3855149419 | internal/mx/scanner_test.go:387 | :409-424 | already-fixed | Assert the ContentHash contract (pinned sha256) in the parseTag table | Empty-hash check for every row (:415-417) + pinned constant for the "NOTE tag" row: `const wantHash = "563b0d54..."` (:419-423), comment cites "CR round-2 3855149419"; scan-level line-drift preservation covered by refresh_test.go:213-250 (ContentHash identical while Line moves) |

## Notes

### MINOR-2 candidate resolution (current mcp_code_tools_test.go line numbers)

Current test-function map (current tree, 213 lines total):
- `TestGraphFileAPI_RejectsSymlinkEscape` — **L60-84**
- `TestGraphTools_HonorProjectRoot` — **L90-140**
- `TestHandleGraphFileAPI` — **L144-166**
- `TestHandleGraphFindAndTrace` — **L169-212**

The card's "around current L83" resolves to the tail of `TestGraphFileAPI_RejectsSymlinkEscape`:

```go
80:	body, _ := res.Content[0].(interface{ GetText() string })
81:	if body != nil && strings.Contains(body.GetText(), "Leaked") {
82:		t.Error("the external file's symbols leaked through the tool")
83:	}
```

**MINOR-2 candidate — confirmed.** L80 indexes `res.Content[0]` with no `len(res.Content)` guard: an error result carrying an empty `Content` slice panics the test instead of failing with a message. The type assertion itself is the safe two-value form (`body, _ :=`), so the panic risk is specifically the unchecked index — the same defect family as finding 4 (3855001928), at a site that did not exist in the round-2 tree. A `toolText`-style helper (or a length check) at L80 resolves it.

Finding-4 relevant current line numbers (the one-value assertion sites the finding asked to route through a helper): **L113, L125** (TestGraphTools_HonorProjectRoot), **L153** (TestHandleGraphFileAPI), **L178, L198** (TestHandleGraphFindAndTrace) — two more sites than the round-2 tree's three.

### Finding 5 — tag half is moot against the current tree

The finding claimed `graph.CodeMatch` declares `json:"symbol"`/`json:"via"`. In the current tree `internal/graph/codequery.go:123-127` declares the wire tags **capitalized**: `json:"Symbol"`, `json:"File"`, `json:"Line"`, `json:"Grade"`, `json:"Via"` — so the test's `json:"Symbol"`/`json:"Via"` (mcp_code_tools_test.go:181-182) already match the wire exactly. The capitalized wire tags in codequery.go are themselves the unusual surface (every other struct in that file uses lowercase tags: `json:"name"`, `json:"file"`, ...) — that file belongs to the verify-graph slice, flagged here only because it flips this finding's tag-mismatch premise. The assertion half (Symbol/Via declared but never asserted) remains valid and is the basis for the still-valid verdict.

### Finding 3 / 27 residue

`TestGraphStampCmd_ValidAndReplacement` covers valid output, replacement, and temp cleanup after a **successful** rename (graph_stamp_test.go:84-87). The failed-rename cleanup sub-case (`os.Rename` fails → deferred `os.Remove(tmp)` still reported via `cmd.PrintErrf`, graph_stamp.go:62-66) remains untested. Minor residue, not tracked by any other finding.

### deferred-by-design SPEC citations

- **#23 (3855149345)** — progress.md §E.2 AC-GF-006: "Deviation recorded: the UNCONFIGURED (nil) posture stays silent — the pre-existing unknown-project silent-pass contract (TestQualityGate_Run_UnknownProjectPasses) is preserved; production paths always populate the config." The in-code comment (gate.go:1185-1189) and the test-side comment (gate_graph_freshness_test.go:136-138) both restate the same decision.
- Checked §E.2 Gaps / §E.3 for every other non-fixed finding: **#1, #4, #5, #6, #8, #10, #12, #13, #19, #20, #21, #25, #26 have no deferral record** — they are simply not yet addressed (the Gaps section covers only CI-red observation, foreign figures, codemaps-md writer adoption, AC-GF-022 baseline, and astx captures).

### Ambiguities

- **#2**: the fix landed in production code (untracked files counted via `git ls-files --others`, check.go:436) rather than in the test (no `git add` was added) — classified already-fixed because the finding's stated failure mode (41 untracked files invisible to the stale count) cannot occur in the current tree.
- **#15**: the finding's TreeRoot-equality ask is consciously rejected in code (check.go:156-161, citing CR round-2 3855149192): codemaps is a tracked artifact that replicates to every checkout, so hard-matching TreeRoot would disable the check everywhere but the stamper's machine. The security-relevant half (DescribedRoots containment incl. symlink escape) is implemented and tested. Classified already-fixed on that basis.

### Count summary

**13 still-valid / 18 already-fixed / 0 invalid-premise / 1 deferred-by-design** (32 total)
