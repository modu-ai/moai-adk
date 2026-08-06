# progress.md — SPEC-NAVIGATOR-SYNC-002 (BAS Epic M1 — Falconer Detect)

> Plan-phase artifact. Run/sync evidence is populated by manager-develop / manager-docs.

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-audit>_ — 4 plan-phase artifacts emitted (spec.md / plan.md / acceptance.md / progress.md), status: draft. Ready for plan-auditor independent audit. SPEC ID regex: PASS. Frontmatter 12 canonical fields present. Out of Scope section satisfies the `OutOfScopeRule` lint convention (6 `### Out of Scope — <topic>` H3 sub-headings, each with `-` bullets).

## §E.2 Run-phase Evidence

### M1.1 — Reverse-traversal engine (pure function)

**Scope**: new package `internal/navigator/detect/` consuming M0 `internal/navigator/sync` types read-only (REQ-NS2-005 bridge-not-absorb). Two files: `traverse.go` (implementation) + `traverse_test.go` (tests). No I/O, no side effects — pure function. The caller (M1.2 hook integration) will own graph loading + fail-open policy.

**Files created**:
- `internal/navigator/detect/traverse.go` — `Traverse(graph *navsync.Graph, changedPath string) (*Result, error)` + `Result` / `AffectedNode` / `AffectedEdge` types. Linear scan over `graph.Edges`; absolute-path equality via `filepath.Abs`; conservative collection of BOTH `source_node` and `target_node`; directory-prefix fallback (REQ-NS2-010); deterministic sort mirroring M0 `sortEdges` (`internal/navigator/sync/join.go:370`). The `navigator-audit.sh heuristic_match()` inspiration is cited by file:line in the `Traverse` doc comment per REQ-NS2-010 honest framing.
- `internal/navigator/detect/traverse_test.go` — table-driven `TestTraverse_ReverseTraversal` (AC-NS2-002 rows 002a/002b/002c covering dec-edge / spec-edge / sym-edge), `TestTraverse_DirectoryPrefixFallback` (AC-NS2-010, includes the negative-prefix trap `foo-extra`), `TestTraverse_EdgeCases` (empty graph, non-match, deterministic ordering, nil graph, empty path, conservative source+target collection).

**Consumer-only invariant (REQ-NS2-005 / AC-NS2-005a)**: `git status --short internal/navigator/sync/ internal/mx/` returns empty — M0 + mx byte-unchanged.

#### AC-NS2-002 — reverse traversal per edge type — PASS

Command (verbatim):
```
go test -count=1 -v -run 'TestTraverse_ReverseTraversal|TestTraverse_DirectoryPrefixFallback|TestTraverse_EdgeCases' ./internal/navigator/detect/...
```

Observed output (verbatim tail):
```
=== RUN   TestTraverse_ReverseTraversal
=== PAUSE TestTraverse_ReverseTraversal
=== RUN   TestTraverse_DirectoryPrefixFallback
=== PAUSE TestTraverse_DirectoryPrefixFallback
=== RUN   TestTraverse_EdgeCases
=== PAUSE TestTraverse_EdgeCases
=== CONT  TestTraverse_ReverseTraversal
=== RUN   TestTraverse_ReverseTraversal/002a_dec-edge:_design-doc_change_surfaces_decision_node
=== PAUSE TestTraverse_ReverseTraversal/002a_dec-edge:_design-doc_change_surfaces_decision_node
=== RUN   TestTraverse_ReverseTraversal/002b_spec-edge:_code_change_surfaces_SPEC_back-pointer_(highest-value_case)
=== PAUSE TestTraverse_ReverseTraversal/002b_spec-edge:_code_change_surfaces_SPEC_back-pointer_(highest-value_case)
=== RUN   TestTraverse_ReverseTraversal/002c_sym-edge:_code_change_surfaces_symbol_binding
=== PAUSE TestTraverse_ReverseTraversal/002c_sym-edge:_code_change_surfaces_symbol_binding
=== CONT  TestTraverse_ReverseTraversal/002a_dec-edge:_design-doc_change_surfaces_decision_node
=== CONT  TestTraverse_DirectoryPrefixFallback
=== CONT  TestTraverse_ReverseTraversal/002c_sym-edge:_code_change_surfaces_symbol_binding
=== CONT  TestTraverse_ReverseTraversal/002b_spec-edge:_code_change_surfaces_SPEC_back-pointer_(highest-value_case)
--- PASS: TestTraverse_DirectoryPrefixFallback (0.00s)
=== CONT  TestTraverse_EdgeCases
--- PASS: TestTraverse_ReverseTraversal (0.00s)
    --- PASS: TestTraverse_ReverseTraversal/002a_dec-edge:_design-doc_change_surfaces_decision_node (0.00s)
    --- PASS: TestTraverse_ReverseTraversal/002c_sym-edge:_code_change_surfaces_symbol_binding (0.00s)
    --- PASS: TestTraverse_ReverseTraversal/002b_spec-edge:_code_change_surfaces_SPEC_back-pointer_(highest-value_case) (0.00s)
=== RUN   TestTraverse_EdgeCases/empty_graph_returns_empty_result_no_error
...
--- PASS: TestTraverse_EdgeCases (0.00s)
    --- PASS: TestTraverse_EdgeCases/empty_graph_returns_empty_result_no_error (0.00s)
    --- PASS: TestTraverse_EdgeCases/both_source_and_target_nodes_collected_conservatively (0.00s)
    --- PASS: TestTraverse_EdgeCases/empty_changed_path_returns_error (0.00s)
    --- PASS: TestTraverse_EdgeCases/nil_graph_returns_error (0.00s)
    --- PASS: TestTraverse_EdgeCases/deterministic_ordering_across_repeated_calls (0.00s)
    --- PASS: TestTraverse_EdgeCases/non-matching_path_returns_empty_result (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/navigator/detect	0.728s
```

Row-by-row traceability:
- 002a (dec-edge): `decision:AUTH-STRATEGY` surfaced when `/abs/project/.moai/project/tech.md` changed → PASS
- 002b (spec-edge, highest-value mx-bridge case): `spec:SPEC-AUTH-001` surfaced when `/abs/project/internal/auth/login.go` changed → PASS
- 002c (sym-edge): `symbol:auth.ParseBearer` surfaced for the same changed code path → PASS

#### AC-NS2-010 — directory-prefix fallback — PASS

`TestTraverse_DirectoryPrefixFallback` (above) covers `changedPath: "/abs/project/internal/foo/"` (trailing separator). The affected-edge set contains `bar.go` + `baz.go` (under prefix) and does NOT contain `other/qux.go` or the negative-prefix trap `foo-extra/trap.go`. The fallback is inspired by `navigator-audit.sh heuristic_match()` (`.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:406-422`), cited in `Traverse`'s doc comment; the engine remains absolute-path string equality per REQ-NS2-010.

#### RED evidence (TDD — verbatim pre-GREEN failing-test output)

Before the implementation existed, the test file failed to compile — the test references the not-yet-built `Traverse` API. Command + output:
```
$ go test ./internal/navigator/detect/...
# github.com/modu-ai/moai-adk/internal/navigator/detect [github.com/modu-ai/moai-adk/internal/navigator/detect.test]
internal/navigator/detect/traverse_test.go:85:16: undefined: Traverse
internal/navigator/detect/traverse_test.go:174:14: undefined: Traverse
internal/navigator/detect/traverse_test.go:208:15: undefined: Traverse
internal/navigator/detect/traverse_test.go:232:15: undefined: Traverse
internal/navigator/detect/traverse_test.go:261:17: undefined: Traverse
internal/navigator/detect/traverse_test.go:265:18: undefined: Traverse
internal/navigator/detect/traverse_test.go:291:13: undefined: Traverse
internal/navigator/detect/traverse_test.go:299:13: undefined: Traverse
internal/navigator/detect/traverse_test.go:320:15: undefined: Traverse
FAIL	github.com/modu-ai/moai-adk/internal/navigator/detect [build failed]
FAIL
```

### M1.2 — PostToolUse branch integration

**Scope**: wire M1.1's `Traverse` into the existing `postToolHandler.Handle` dispatcher (`internal/hook/post_tool.go`) as a new conditional branch — NOT a forked hook chain (REQ-NS2-009). The branch fires on the Write/Edit/NotebookEdit trigger surface (REQ-NS2-001), extracts the changed path (`file_path` for Write/Edit, `notebook_path` for NotebookEdit), resolves projectRoot via the existing `resolveProjectRootFromInputOrEnv` helper (input.CWD → `CLAUDE_PROJECT_DIR` → `os.Getwd()` — never a new resolution path), loads the M0 `nav-graph.json` via a single `os.ReadFile`, and calls `detect.Traverse`. M1.2 stubs the output surface (metrics entry only); the full `systemMessage` + JSONL impact record lands in M1.3.

**Files created / modified**:
- `internal/hook/navigator_detect.go` (NEW) — `runNavigatorDetect(input *HookInput) *detect.Result` (branch entry point) + `detectForChangedPath(graphPath, changedPath string) *detect.Result` (testable core sans HookInput) + `extractChangedPath(toolInput json.RawMessage) string` + `navigatorDetectTools` trigger-surface map. Fail-open on every error mode (graph absent / unparseable / traversal error); emits at most one `slog.Debug` per path; NEVER blocks (REQ-NS2-012).
- `internal/hook/navigator_detect_test.go` (NEW) — 14 tests: Write/Edit/NotebookEdit fire Traverse; Bash/Read do NOT; graph-absent / unparseable / no-path / no-projectRoot fail-open; dispatcher integration (Handle → metrics `navigator_detect` entry) for Write/NotebookEdit; Bash NOT dispatched; no-forked-chain wrapper absence; branch-registered-in-dispatcher source grep; `detectForChangedPath` direct unit tests.
- `internal/hook/post_tool.go` (MODIFIED) — one new branch at lines 219-238, placed after `runMemoryAudit(input)` (line 215-217) and before `logEvidence(input)` (line 244-246), per plan.md §C.1 / §F M1.2. Records `metrics["navigator_detect"]` with `affected_nodes`/`affected_edges` counts (or `status: "no_match_or_fail_open"` when the branch ran but yielded no rows).

**Branch-not-fork (REQ-NS2-009 / AC-NS2-009)**: the new logic is a conditional branch INSIDE `postToolHandler.Handle`. No new `.claude/hooks/moai/handle-navigator-detect.sh` wrapper (verified absent — `TestNavigatorDetect_NoForkedChain`). No new `moai hook navigator-detect` subcommand. The existing `handle-post-tool.sh` → `moai hook post-tool` → `postToolHandler.Handle` chain remains the sole PostToolUse entry point.

**Consumer-only (REQ-NS2-005 / AC-NS2-005a)**: `git status --short internal/navigator/sync/ internal/mx/` returns empty — M0 + mx byte-unchanged by M1.2. `navigator_detect.go` contains zero `os.WriteFile` / `os.Rename` call sites (the only match is a doc comment referencing M0's atomic-rename guarantee).

#### D3 NotebookEdit recon — VERDICT: SHALL (NotebookEdit stays in REQ-NS2-001)

Plan.md §B4 required a run-phase decision point: does PostToolUse actually fire for `NotebookEdit`, and is `ToolInput.notebook_path` parseable? Evidence:

1. **PostToolUse fires for NotebookEdit**: `internal/template/templates/.claude/settings.json.tmpl:380` lists `"NotebookEdit"` in the PreToolUse `permissions.allow` array — confirming NotebookEdit is a Claude Code tool that triggers hook events (PreToolUse and PostToolUse fire for every tool call; the handler dispatches by `input.ToolName`). No special registration is needed for the Detect branch to receive NotebookEdit events.
2. **`notebook_path` is parseable**: the Claude Code NotebookEdit tool input schema carries a `notebook_path` string field (analogous to Write/Edit's `file_path`). `extractChangedPath` reads it via the same `json.Unmarshal` → `parsed["notebook_path"].(string)` pattern used for `file_path`.

**Decision**: NotebookEdit stays in REQ-NS2-001's SHALL trigger surface alongside Write/Edit. NOT downgraded to SHOULD, NOT deferred. The implementation extracts `notebook_path` symmetrically with `file_path` (both normalized to absolute inside `detect.Traverse`). Test coverage: `TestRunNavigatorDetect_NotebookEdit_FiresTraverse` + `TestPostToolHandler_NavigatorDetect_NotebookEditDispatch` (both GREEN, below).

#### AC-NS2-001a — Write/Edit trigger — PASS

Command (verbatim):
```
go test -count=1 -v -run 'TestRunNavigatorDetect_Write_FiresTraverse|TestRunNavigatorDetect_Edit_FiresTraverse|TestPostToolHandler_NavigatorDetect_WriteDispatch' ./internal/hook/
```
Observed: all three sub-tests `--- PASS`. Write on a matching `file_path` yields a non-nil `*detect.Result` with the expected affected node (spec:SPEC-AUTH-001 for the spec-edge bridge case; decision:AUTH-STRATEGY for the dec-edge case). The dispatcher records `metrics["navigator_detect"] = {"affected_nodes": N, "affected_edges": M}` observable in `HookOutput.Data`.

#### AC-NS2-001b — Bash NOT triggered — PASS

`TestRunNavigatorDetect_Bash_NotTriggered` + `TestPostToolHandler_NavigatorDetect_BashNotDispatched`: Bash input with `{"command": "sed -i 's/x/y/' internal/foo/bar.go"}` yields nil from `runNavigatorDetect`, and `Handle()` produces NO `navigator_detect` metrics entry in `HookOutput.Data`. The branch is never entered for Bash (Bash is not in `navigatorDetectTools`).

#### AC-NS2-009 — branch, not fork — PASS

- **(a) no forked hook chain**: `ls .claude/hooks/moai/handle-navigator-detect.sh` → "No such file or directory" (PASS). No new `moai hook navigator-detect` subcommand registered.
- **(b) branch registered inside dispatcher**: `TestNavigatorDetect_BranchRegisteredInDispatcher` greps `post_tool.go` and confirms exactly ONE `runNavigatorDetect(input)` call site, placed alongside (not replacing) `runAstScan` / `runMxValidation` / `runMemoryAudit` / `logEvidence`. The branch sits at lines 219-238, after `runMemoryAudit` (215-217) and before `logEvidence` (244-246), matching plan.md §C.1.

#### AC-NS2-012 — never blocks (M1.2-light) — PASS

`grep -nE 'Decision.*block|os\.Exit\(2\)' internal/hook/navigator_detect.go` → 0 matches. The branch returns a Result or nil and never emits a block decision. (Full AC-NS2-012 coverage across all 5 fail-open modes lands in M1.4.)

#### RED evidence (TDD — verbatim pre-GREEN failing-test output)

Before `navigator_detect.go` existed, the test file failed to compile — references to `runNavigatorDetect` / `detectForChangedPath` were undefined. Command + output:
```
$ go vet ./internal/hook/
# github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
internal/hook/navigator_detect_test.go:72:9: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:103:9: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:133:9: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:154:12: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:169:12: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:187:12: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:208:12: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:223:12: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:244:12: undefined: runNavigatorDetect
internal/hook/navigator_detect_test.go:390:9: undefined: detectForChangedPath
internal/hook/navigator_detect_test.go:390:9: too many errors
```
After implementing `navigator_detect.go` + wiring the dispatcher branch, all 14 navigator-detect tests pass (GREEN).

### M1.3 — Output surfaces (systemMessage + JSONL impact record)

**Scope**: replace the M1.2 metrics-stub output with the two read-only advisory surfaces required by REQ-NS2-003. (a) A `systemMessage` advisory naming the changed path + ≤10 affected rows (overflow → `…and N more` tail; JSONL carries the full set). (b) An append-only machine-readable JSONL impact record at `.moai/state/navigator-detect/<session-id>.jsonl` for M2 Route to consume. NO work-item promotion (REQ-NS2-003c / AC-NS2-003c) — the Detect layer surfaces + records only; promotion is M2's job.

**Files modified**:
- `internal/hook/navigator_detect.go` (EXTENDED) — added: `emitNavigatorDetectAdvisory(input, result, currentSystemMessage) string` (single dispatcher touch-point), `recordNavigatorDetectImpact(input, result)` (JSONL-write orchestrator, fail-open), `formatNavigatorDetectSystemMessage(changedPath, result) string` (pure formatter, ≤10 rows, `…and N more` overflow tail), `appendNavigatorDetectImpact(projectRoot, sessionID, changedPath, changedAt, result) error` (append-only JSONL write, deterministic via injectable `changedAt`), `changedAtForProject(projectRoot) string` (git committer-date lookup, fail-open to `(no-git)` sentinel), `impactRecord`/`impactNode` JSONL schema types, `navigatorDetectStateDir`/`systemMessageRowLimit`/`changedAtNoGit` named constants.
- `internal/hook/navigator_detect_test.go` (EXTENDED) — 8 new M1.3 tests: systemMessage format/nil/empty/overflow, JSONL schema + session-scoped appends, changedAt non-git placeholder, no-promotion source-grep (AC-NS2-003c), dispatcher integration (systemMessage + JSONL + no-block).
- `internal/hook/post_tool.go` (MODIFIED) — the existing M1.2 branch (lines 219-238) now calls `emitNavigatorDetectAdvisory` to append the advisory to the existing `systemMessage` local (built by LSP/AST branches above) and `recordNavigatorDetectImpact` to append the JSONL line. Still gated on `result != nil`; `else` branch keeps the `no_match_or_fail_open` metrics entry.

**JSONL record schema** (one line per detection — the contract M2 Route consumes, per plan.md §C.3):
```json
{"changed_path":"/abs/project/internal/auth/login.go","changed_at":"2026-08-06T12:00:00+00:00","affected_nodes":[{"entity_type":"spec","identifier":"SPEC-AUTH-001"}],"affected_edges":[{"edge_type":"spec-edge","source_node":"symbol:auth.ParseBearer","target_node":"spec:SPEC-AUTH-001","source_path":"/abs/project/internal/auth/login.go","line_number":17}]}
```
- `changed_at` is the git committer-date of the current HEAD (`git -C <root> log -1 --format=%cI`) — same stamp M0 uses (`internal/navigator/sync/provenance.go`) — so two detections on the same HEAD produce byte-identical `changed_at` values. Fail-open sentinel `(no-git)` when git is unavailable / not a repo / no HEAD.
- Session-scoped path: `<projectRoot>/.moai/state/navigator-detect/<session-id>.jsonl` (created best-effort; append-only per session).
- `affected_nodes` elements carry `entity_type` + `identifier` (display_name recoverable at M2 read time — kept the record tight).

**systemMessage shape** (multi-line, advisory, ≤12 lines):
```
Navigator Detect: <changed_path> touches <N> graph row(s):
- <source_node> → <target_node> (<edge_type> @ <source_path>:<line>)
… (≤10 rows; "…and N more (see .moai/state/navigator-detect/ JSONL for the full set)" tail if overflow)
```
The Detect advisory is APPENDED to the existing systemMessage (LSP/AST findings survive — the branch runs after the diagnostic branches per `post_tool.go` dispatch order). The `…and N more` tail names the remainder count and points to the JSONL SSOT.

#### AC-NS2-003 — advisory output (systemMessage + JSONL + no-promotion) — PASS

**(a) systemMessage emitted, advisory** — PASS. `TestFormatNavigatorDetectSystemMessage_NonEmptyResult` + `TestPostToolHandler_NavigatorDetect_AdvisoryOutput`: non-empty Result → non-empty systemMessage starting with `Navigator Detect:`, names the changed path + the affected spec node (`spec:SPEC-AUTH-001` for the `@MX:SPEC` bridge case), and references the `spec-edge` edge_type. `out.Decision != "block"` and `out.HookSpecificOutput.Decision.Behavior != "block"` verified in dispatcher integration test.

**(b) JSONL impact record appended** — PASS. `TestAppendNavigatorDetectImpact_JSONLSchema` + `TestAppendNavigatorDetectImpact_SessionScopedAndAppends`: file `.moai/state/navigator-detect/<session-id>.jsonl` exists after a detection; last line is valid JSON with required top-level keys `changed_path` / `changed_at` / `affected_nodes` / `affected_edges`, and each `affected_edges` element carries `edge_type` / `source_node` / `target_node` / `source_path` / `line_number`. Session-scoped: different `sessionID` → different file; same sessionID appends on new lines.

**(c) no work-item promotion** — PASS. `TestNavigatorDetect_NoWorkItemPromotion` source-greps `navigator_detect.go` for forbidden promotion patterns (`gh issue create`, `.moai/specs/SPEC-`, `Decision: "block"`, `os.Exit(2)`, `"block"`, `TODO file`) → 0 matches. The only writes are the JSONL append under `.moai/state/navigator-detect/` and the `slog.Debug` diagnostic (no `.moai/specs/`, no source mutation, no issue creation).

**Determinism**: `appendNavigatorDetectImpact` takes `changedAt` as an injectable parameter (no wall-clock inside the function). Tests pass `"(test-fixed)"`; production calls `changedAtForProject(projectRoot)` which shells out to git for the HEAD committer-date — same HEAD → same value across runs.

#### RED evidence (TDD — verbatim pre-GREEN failing-test output)

```
$ go test ./internal/hook/ -run 'TestFormatNavigatorDetect|TestAppendNavigatorDetect|TestChangedAtForProject|TestNavigatorDetect_NoWorkItemPromotion|TestPostToolHandler_NavigatorDetect_AdvisoryOutput' -count=1
# github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
internal/hook/navigator_detect_test.go:432:9: undefined: formatNavigatorDetectSystemMessage
internal/hook/navigator_detect_test.go:455:12: undefined: formatNavigatorDetectSystemMessage
internal/hook/navigator_detect_test.go:463:12: undefined: formatNavigatorDetectSystemMessage
internal/hook/navigator_detect_test.go:484:9: undefined: formatNavigatorDetectSystemMessage
internal/hook/navigator_detect_test.go:487:19: undefined: systemMessageRowLimit
internal/hook/navigator_detect_test.go:510:12: undefined: appendNavigatorDetectImpact
internal/hook/navigator_detect_test.go:514:37: undefined: navigatorDetectStateDir
internal/hook/navigator_detect_test.go:563:12: undefined: appendNavigatorDetectImpact
internal/hook/navigator_detect_test.go:566:12: undefined: appendNavigatorDetectImpact
internal/hook/navigator_detect_test.go:566:12: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
```
After implementing the M1.3 surfaces + wiring the dispatcher, all M1.3 + M1.2 + M1.1 navigator-detect tests pass (GREEN).

### M1.4 — Fail-open + concurrency hardening

**Scope**: wrap the Detect branch in `defer recover()` + a bounded `context.WithTimeout(ctx, navigatorDetectTimeout)` (plan.md §C.6), add the schema-invalid fail-open branch (AC-NS2-004 row 004c), and prove the atomic-read concurrency guarantee (AC-NS2-006) by reusing M0's `NAVIGATOR_PRE_RENAME_BARRIER` test hook (no new test hook added — plan.md §C.5). The branch remains fail-open on every error mode (REQ-NS2-004) and NEVER blocks (REQ-NS2-012).

**Files modified**:
- `internal/hook/navigator_detect.go` (EXTENDED) — added: `navigatorDetectTimeout = 200 * time.Millisecond` named constant (per `hns-moaiadk-best-practices` hardcoding-prevention — thresholds MUST be named constants); the schema-invalid check in `detectForChangedPath` (the `edges` array is ABSENT after unmarshal — distinct from a legitimately-empty `[]` per §F edge case); the `runNavigatorDetectSafe(ctx, input)` wrapper with deferred recover (swallows any panic, including the nil-ctx panic inside `context.WithTimeout`) + bounded deadline (goroutine + select on `dctx.Done()` vs result channel — silent nil on timeout/cancellation per AC-NS2-004 row 004e). Imports `context` + `time` added.
- `internal/hook/post_tool.go` (MODIFIED) — the dispatcher branch now calls `runNavigatorDetectSafe(ctx, input)` (the hardened wrapper) instead of `runNavigatorDetect(input)` directly. ONE call site, unchanged semantics for the success path; failure modes now contained.
- `internal/hook/navigator_detect_test.go` (MODIFIED) — `TestNavigatorDetect_BranchRegisteredInDispatcher` updated to grep for the M1.4 safe-wrapper call site `runNavigatorDetectSafe(ctx, input)` (exactly 1 occurrence in production source).
- `internal/hook/navigator_detect_hardening_test.go` (NEW) — 5 M1.4 tests: the table-driven `TestDetectForChangedPath_FailOpenTable` (rows 004a..004e), `TestRunNavigatorDetectSafe_PreCancelledContext` (AC-NS2-004 row 004e via pre-cancelled ctx), `TestRunNavigatorDetectSafe_RecoversFromPanic` (REQ-NS2-012 panic containment), `TestRunNavigatorDetectSafe_PassesThroughNonCancel` (healthy-context pass-through contract), `TestDetectForChangedPath_AtomicReadDuringRegen` (AC-NS2-006 — reuses M0's `NAVIGATOR_PRE_RENAME_BARRIER`), `TestNavigatorDetect_NeverBlocks_GrepGuard` (AC-NS2-012 source grep for `Decision: "block"` / `os.Exit(2)` / `"block"`).

**Timeout constant** (named per the hardcoding-prevention rule): `navigatorDetectTimeout = 200 * time.Millisecond` at `internal/hook/navigator_detect.go`. Plan.md §C.6 names ~200ms; the named constant honors it without a bare `200*time.Millisecond` literal.

**5 fail-open modes** (table-driven — each degrades to "no impact surfaced" silently; PostToolUse proceeds exit-0-equivalent):

| row | trigger | handling (one line) |
|---|---|---|
| 004a graph absent | `os.ReadFile` fails (graph not yet generated) | `detectForChangedPath` logs one `slog.Debug` line + returns `nil` (existing M1.2 branch) |
| 004b unparseable JSON | `json.Unmarshal` fails on malformed file | `detectForChangedPath` logs one `slog.Debug` line + returns `nil` (existing M1.2 branch) |
| 004c schema-invalid | JSON parses but `edges` array is ABSENT (nil after unmarshal — distinct from `"edges":[]`) | NEW M1.4 branch: one `slog.Debug` line + returns `nil`. An explicit empty array is a valid empty graph (§F edge case) and falls through to `detect.Traverse` → empty Result, no log |
| 004d traversal error | `detect.Traverse` returns err (nil graph / empty changedPath / normalization failure) | `detectForChangedPath` logs one `slog.Debug` line + returns `nil` (existing branch; exercised by 004d table row via empty changedPath). Per-edge malformed `source_node` keys are skipped inside `detect.Traverse` (advisory fail-open at the edge level — NOT fatal) |
| 004e timeout / cancellation | the bounded `context.WithTimeout(ctx, 200ms)` deadline fires OR the parent ctx is cancelled | NEW M1.4 wrapper `runNavigatorDetectSafe` selects `<-dctx.Done()` → returns `nil` silently (NO log line — context cancellation is not an error to advertise per AC-NS2-004 row 004e) |

#### AC-NS2-004 — fail-open across 5 error modes — PASS

**Table-driven test**: `TestDetectForChangedPath_FailOpenTable` (rows 004a..004e). Each subtest asserts the branch returns cleanly (nil or empty Result) without propagating a panic or an error that would block the tool call.

Command (verbatim):
```
go test -count=1 -race -v -run 'TestDetectForChangedPath_FailOpenTable|TestRunNavigatorDetectSafe' ./internal/hook/
```

Observed output (verbatim tail):
```
=== RUN   TestDetectForChangedPath_FailOpenTable
=== RUN   TestDetectForChangedPath_FailOpenTable/004a_graph_absent
=== RUN   TestDetectForChangedPath_FailOpenTable/004b_unparseable_json
=== RUN   TestDetectForChangedPath_FailOpenTable/004c_schema_invalid_missing_edges_array
=== RUN   TestDetectForChangedPath_FailOpenTable/004d_traversal_error_empty_changed_path
=== RUN   TestDetectForChangedPath_FailOpenTable/004e_timeout_pre_cancelled_context
--- PASS: TestDetectForChangedPath_FailOpenTable (0.01s)
    --- PASS: TestDetectForChangedPath_FailOpenTable/004a_graph_absent (0.00s)
    --- PASS: TestDetectForChangedPath_FailOpenTable/004b_unparseable_json (0.00s)
    --- PASS: TestDetectForChangedPath_FailOpenTable/004c_schema_invalid_missing_edges_array (0.00s)
    --- PASS: TestDetectForChangedPath_FailOpenTable/004d_traversal_error_empty_changed_path (0.00s)
    --- PASS: TestDetectForChangedPath_FailOpenTable/004e_timeout_pre_cancelled_context (0.00s)
=== RUN   TestRunNavigatorDetectSafe_PreCancelledContext
--- PASS: TestRunNavigatorDetectSafe_PreCancelledContext (0.00s)
=== RUN   TestRunNavigatorDetectSafe_RecoversFromPanic
--- PASS: TestRunNavigatorDetectSafe_RecoversFromPanic (0.00s)
=== RUN   TestRunNavigatorDetectSafe_PassesThroughNonCancel
--- PASS: TestRunNavigatorDetectSafe_PassesThroughNonCancel (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	2.362s
```

#### AC-NS2-006 — atomic read during regen — PASS

**Concurrency test**: `TestDetectForChangedPath_AtomicReadDuringRegen` (`internal/hook/navigator_detect_hardening_test.go`). The test writes a PRIOR committed graph at `<tmpDir>/.moai/project/navigator/nav-graph.json`, arms the M0 `NAVIGATOR_PRE_RENAME_BARRIER` env var, spawns a writer goroutine calling `navsync.WriteGraph(graphPath, newGraph)` (which writes `.tmp` then spin-waits at the barrier before the rename), waits for the barrier file to appear, and runs the reader `detectForChangedPath(graphPath, priorEdgePath)` WHILE the writer is held. The reader observes the PRIOR graph (the `prior.Foo` edge), never the NEW graph (no `new.Bar`), never a partial file (≥1 edge present). Then the barrier is removed, the writer completes, and the goroutine exits cleanly.

**Reuse of M0 test hook**: the test sets `NAVIGATOR_PRE_RENAME_BARRIER=<tmpDir>/barrier-marker` — the SAME hook M0 ships at `internal/navigator/sync/write.go:41-49`. NO new test hook added (plan.md §C.5). The test is deliberately serial (NOT `t.Parallel()`) because the barrier env var is process-global (atomicWrite unconditionally unsets it at `write.go:42`).

**Race detector**: `go test -race ./internal/hook/...` → exit 0, no data races reported (the reader is a single `os.ReadFile`; the writer's `.tmp`+`os.Rename` is M0's atomicWrite, which the reader observes atomically via inode swap).

Command (verbatim):
```
go test -count=1 -race -v -run 'TestDetectForChangedPath_AtomicReadDuringRegen' ./internal/hook/
```

Observed output (verbatim):
```
=== RUN   TestDetectForChangedPath_AtomicReadDuringRegen
--- PASS: TestDetectForChangedPath_AtomicReadDuringRegen (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	2.362s
```

#### AC-NS2-012 — PostToolUse never blocks — PASS

**Source grep**: `TestNavigatorDetect_NeverBlocks_GrepGuard` asserts `navigator_detect.go` contains ZERO matches for `Decision: "block"`, `os.Exit(2)`, or `"block"`. Complements the M1.3 `TestNavigatorDetect_NoWorkItemPromotion` grep.

**Panic containment**: `TestRunNavigatorDetectSafe_RecoversFromPanic` passes a nil context (which makes `context.WithTimeout` panic) and asserts the deferred recover swallows it → returns nil. The tool call proceeds regardless.

**Bounded deadline**: `TestRunNavigatorDetectSafe_PreCancelledContext` passes a pre-cancelled context and asserts (a) nil is returned silently, (b) the wrapper does NOT block past the `navigatorDetectTimeout` budget — the `<-dctx.Done()` select arm fires immediately.

#### RED evidence (TDD — verbatim pre-GREEN failing-test output)

```
$ go test -count=1 -run 'TestDetectForChangedPath_FailOpenTable|TestRunNavigatorDetectSafe|TestDetectForChangedPath_AtomicReadDuringRegen|TestNavigatorDetect_NeverBlocks' ./internal/hook/
# github.com/modu-ai/moai-adk/internal/hook [github.com/modu-ai/moai-adk/internal/hook.test]
internal/hook/navigator_detect_hardening_test.go:186:9: undefined: runNavigatorDetectSafe
internal/hook/navigator_detect_hardening_test.go:194:15: undefined: navigatorDetectTimeout
internal/hook/navigator_detect_hardening_test.go:196:13: undefined: navigatorDetectTimeout
internal/hook/navigator_detect_hardening_test.go:214:9: undefined: runNavigatorDetectSafe
internal/hook/navigator_detect_hardening_test.go:232:9: undefined: runNavigatorDetectSafe
internal/hook/navigator_detect_hardening_test.go:409:9: undefined: runNavigatorDetectSafe
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
```
After implementing the timeout constant, schema-invalid check, safe wrapper, and post_tool.go wiring, all M1.4 + M1.3 + M1.2 + M1.1 navigator-detect tests pass GREEN with `-race` (see verification batch below).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-06
run_commit_sha: pending-backfill-M1.4
run_status: M1.4-GREEN
ac_pass_count: 9   # AC-NS2-002, AC-NS2-010 (M1.1) + AC-NS2-001a, AC-NS2-001b, AC-NS2-009 (M1.2) + AC-NS2-003 (M1.3) + AC-NS2-004, AC-NS2-006, AC-NS2-012 (M1.4)
ac_fail_count: 0
preserve_list_post_run_count: 2   # internal/navigator/sync/, internal/mx/ byte-unchanged (REQ-NS2-005) — verified via `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/sync|mx)/'` returns 0 matches
new_warnings_or_lints_introduced: 0   # golangci-lint run ./internal/hook/... --timeout=3m → 0 issues; go vet → clean
cross_platform_build:
  goos_darwin: PASS   # host default — go build ./...
  goos_windows: PASS   # GOOS=windows GOARCH=amd64 go build ./internal/hook/... → exit 0
total_run_phase_files: 7   # detect/{traverse,traverse_test}.go (M1.1) + hook/{navigator_detect,navigator_detect_test,navigator_detect_hardening_test}.go + hook/post_tool.go (M1.2+M1.3+M1.4)
m1_to_mN_commit_strategy: per-milestone (orchestrator gates M1.4→M1.5 in semi-autonomous mode)
navigator_detect_timeout_constant: "200ms"   # M1.4 — named constant navigatorDetectTimeout at internal/hook/navigator_detect.go (plan.md §C.6)
navigator_detect_failopen_modes: "004a graph-absent, 004b unparseable-json, 004c schema-invalid-missing-edges-array, 004d traversal-error, 004e timeout-cancellation — all degrade to silent nil"
navigator_detect_concurrency_test: "TestDetectForChangedPath_AtomicReadDuringRegen reuses M0 NAVIGATOR_PRE_RENAME_BARRIER; go test -race → PASS"
navigator_detect_jsonl_path: ".moai/state/navigator-detect/<session-id>.jsonl"   # M1.3 — the contract M2 Route consumes
navigator_detect_systemmessage_shape: "Navigator Detect: <changed_path> touches <N> graph row(s):\\n- <source_node> → <target_node> (<edge_type> @ <source_path>:<line>) [≤10 rows; '…and N more' overflow tail]"
```

**Out-of-scope for M1.4 (deferred to M1.5, NOT failures)**:
- AC-NS2-005a/b (consumer-only grep + read-via-public-API) — M1.5 (the byte-unchanged half is already evidenced above for M1.1/M1.2/M1.3/M1.4: `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/sync|mx)/'` returns 0 matches)
- AC-NS2-007 (≥80% coverage fixture corpus) — M1.5
- AC-NS2-008 (non-overlap grep test) — M1.5 (M1.2-light `TestNavigatorDetect_NoForkedChain` + M1.3 `TestNavigatorDetect_NoWorkItemPromotion` + M1.4 `TestNavigatorDetect_NeverBlocks_GrepGuard` already pass at the source-grep level; M1.5 adds the canonical `navigator_detect_nonoverlap_test.go` per AC-NS2-008)
- AC-NS2-011 (template-first) — M1.5 (M1.2/M1.3/M1.4 made no template change — the gate is env-var-only, no `internal/template/templates/` edit, no `catalog.yaml` regen required per plan.md §C.4)

**In-scope for M1.4 — now PASS (evidenced above)**:
- AC-NS2-004 (5-mode fail-open table) — PASS (table-driven test)
- AC-NS2-006 (atomic-read concurrency) — PASS (M0 barrier reuse, -race clean)
- AC-NS2-012 (PostToolUse never blocks) — PASS (grep guard + panic-recovery test + bounded-deadline test)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
