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

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-06
run_commit_sha: pending-backfill-M1.1
run_status: M1.1-GREEN
ac_pass_count: 2   # AC-NS2-002, AC-NS2-010
ac_fail_count: 0
preserve_list_post_run_count: 2   # internal/navigator/sync/, internal/mx/ byte-unchanged (REQ-NS2-005)
new_warnings_or_lints_introduced: 0
total_run_phase_files: 2   # internal/navigator/detect/{traverse,traverse_test}.go
m1_to_mN_commit_strategy: per-milestone (M1.1 is the first run-phase commit; orchestrator gates M1.1→M1.2 in semi-autonomous mode)
```

**Out-of-scope for M1.1 (deferred to later milestones, NOT failures)**:
- AC-NS2-001a/b (Write/Edit trigger surface) — M1.2
- AC-NS2-003 (systemMessage + JSONL output) — M1.3
- AC-NS2-004 (5-mode fail-open) — M1.4
- AC-NS2-005a/b (consumer-only grep + read-via-public-API) — M1.5 (the byte-unchanged half is already evidenced above)
- AC-NS2-006 (atomic-read concurrency) — M1.4
- AC-NS2-007 (≥80% coverage fixture corpus) — M1.5
- AC-NS2-008 (non-overlap grep test) — M1.5
- AC-NS2-009 (branch-not-fork integration) — M1.2
- AC-NS2-011 (template-first) — M1.5
- AC-NS2-012 (never-blocks) — M1.4

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
