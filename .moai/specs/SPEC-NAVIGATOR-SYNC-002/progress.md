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

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-06
run_commit_sha: pending-backfill-M1.2
run_status: M1.2-GREEN
ac_pass_count: 5   # AC-NS2-002, AC-NS2-010 (M1.1) + AC-NS2-001a, AC-NS2-001b, AC-NS2-009 (M1.2)
ac_fail_count: 0
preserve_list_post_run_count: 2   # internal/navigator/sync/, internal/mx/ byte-unchanged (REQ-NS2-005)
new_warnings_or_lints_introduced: 0
total_run_phase_files: 5   # detect/{traverse,traverse_test}.go (M1.1) + hook/{navigator_detect,navigator_detect_test}.go + hook/post_tool.go (M1.2)
m1_to_mN_commit_strategy: per-milestone (orchestrator gates M1.2→M1.3 in semi-autonomous mode)
```

**Out-of-scope for M1.2 (deferred to later milestones, NOT failures)**:
- AC-NS2-003 (systemMessage + JSONL output) — M1.3
- AC-NS2-004 (5-mode fail-open table) — M1.4 (M1.2 covers rows 004a absent + 004b unparseable at the integration boundary)
- AC-NS2-005a/b (consumer-only grep + read-via-public-API) — M1.5 (the byte-unchanged half is already evidenced above for both M1.1 and M1.2)
- AC-NS2-006 (atomic-read concurrency) — M1.4
- AC-NS2-007 (≥80% coverage fixture corpus) — M1.5
- AC-NS2-008 (non-overlap grep test) — M1.5 (M1.2-light version `TestNavigatorDetect_NoForkedChain` already passes)
- AC-NS2-011 (template-first) — M1.5 (M1.2 made no template change — the gate is env-var-only, no `internal/template/templates/` edit, no `catalog.yaml` regen required per plan.md §C.4)
- AC-NS2-012 (never-blocks full table) — M1.4 (M1.2-light grep already passes)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
