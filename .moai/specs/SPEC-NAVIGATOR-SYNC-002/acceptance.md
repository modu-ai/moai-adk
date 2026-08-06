# acceptance.md — SPEC-NAVIGATOR-SYNC-002 (BAS Epic M1 — Falconer Detect)

> Acceptance criteria for the M1 Detect layer. Each AC is a binary-testable Given-When-Then,
> traceable to a spec.md §C REQ. The coverage AC (AC-NS2-007) names the exact command + ratio formula
> (verification-claim-integrity §1.1 surface 3 + §2).

## §A. Severity Legend

- **MUST** — failure blocks merge (M1 cannot ship without this passing).
- **SHOULD** — failure is a debt item, mergeable with a recorded `[NEEDS CLARIFICATION]` or follow-up.
- **MAY** — quality bar, not a gate.

## §B. Traceability Matrix

| REQ | AC(s) | Severity | Milestone |
|---|---|---|---|
| REQ-NS2-001 (trigger surface) | AC-NS2-001a, AC-NS2-001b | MUST | M1.2 |
| REQ-NS2-002 (reverse traversal) | AC-NS2-002 (table-driven: dec/spec/sym) | MUST | M1.1 |
| REQ-NS2-003 (advisory output) | AC-NS2-003 (sub-assertions a/b/c) | MUST | M1.3 |
| REQ-NS2-004 (fail-open) | AC-NS2-004 (table-driven: 5 modes) | MUST | M1.4 |
| REQ-NS2-005 (consumer-only) | AC-NS2-005a, AC-NS2-005b | MUST | M1.5 |
| REQ-NS2-006 (concurrency atomic read) | AC-NS2-006 | MUST | M1.4 |
| REQ-NS2-007 (≥80% coverage, measured) | AC-NS2-007 | MUST | M1.5 |
| REQ-NS2-008 (non-overlap) | AC-NS2-008 | MUST | M1.5 |
| REQ-NS2-009 (branch, not fork) | AC-NS2-009 (sub-assertions a/b) | MUST | M1.2 |
| REQ-NS2-010 (asset reuse) | AC-NS2-010 | MUST | M1.1 |
| REQ-NS2-011 (template-first) | AC-NS2-011 | MUST | M1.5 |
| REQ-NS2-012 (never blocks) | AC-NS2-012 | MUST | M1.4 |

## §C. Definition of Done

- Every MUST AC in §D shows PASS with a cited command + observed output (no narrative "passes").
- `go test ./internal/hook/ -run TestNavigatorDetect -count=1` exits 0.
- `go test ./internal/hook/ -run TestNavigatorDetectCoverage -v` prints an observed coverage percentage ≥ 80.0.
- `go test ./internal/hook/ -run TestNavigatorDetectNonOverlap -count=1` exits 0.
- `moai spec lint --strict .moai/specs/SPEC-NAVIGATOR-SYNC-002/` reports 0 errors (the repo-wide baseline of pre-existing findings is NOT attributed to this SPEC).
- A concurrency test using `NAVIGATOR_PRE_RENAME_BARRIER` exits 0 (the Detect reader observes the prior graph during a held rename).
- No new `moai hook navigator-detect` subcommand exists; the existing `handle-post-tool.sh` wrapper is the only PostToolUse entry point (AC-NS2-009 sub-assertion a).
- `internal/navigator/sync/` and `internal/mx/` are byte-unchanged by the M1 run-phase diff (AC-NS2-005a).

## §D. Acceptance Criteria (Given-When-Then)

### AC-NS2-001a — Write/Edit trigger (REQ-NS2-001, MUST, M1.2)

**Given** a `nav-graph.json` fixture containing an edge with `source_path` matching `/abs/project/internal/foo/bar.go` and a PostToolUse `HookInput` with `ToolName: "Write"` and `ToolInput: {"file_path": "/abs/project/internal/foo/bar.go"}`,
**When** the post-tool handler dispatches the Detect branch,
**Then** the Detect branch produces an affected-row set containing that edge (and its target node) and the process exits 0 with `Decision: "allow"`.

### AC-NS2-001b — Bash NOT triggered (REQ-NS2-001, MUST, M1.2)

**Given** a PostToolUse `HookInput` with `ToolName: "Bash"` and `ToolInput: {"command": "sed -i 's/x/y/' internal/foo/bar.go"}`,
**When** the post-tool handler runs,
**Then** the Detect branch is NOT invoked (no `navigator-detect` log line, no JSONL append), and the handler proceeds normally.

### AC-NS2-002 — reverse traversal per edge type (REQ-NS2-002, MUST, M1.1)

**Given** a `nav-graph.json` fixture with edges of each type whose `source_path` is an absolute path under `/abs/project/`,
**When** the traversal runs with the matching `changedPath` for each row in the table below,
**Then** for every row the affected-edge set contains that edge and the affected-node set contains the named node.

| case | edge_type | changed-source-path (fixture) | expected affected-node |
|---|---|---|---|
| 002a dec-edge | dec-edge | `/abs/project/.moai/project/tech.md` (line 42, `source_node: "decision:AUTH-STRATEGY"`) | `decision:AUTH-STRATEGY` |
| 002b spec-edge (`@MX:SPEC` bridge — the highest-value case) | spec-edge | `/abs/project/internal/auth/login.go` (line 17, `target_node: "spec:SPEC-AUTH-001"`, produced by the M0 mx-bridge) | `spec:SPEC-AUTH-001` |
| 002c sym-edge | sym-edge | `/abs/project/internal/auth/login.go` (line 30, `source_node: "symbol:auth.ParseBearer"`) | `symbol:auth.ParseBearer` |

### AC-NS2-003 — advisory output: systemMessage + JSONL + no-promotion (REQ-NS2-003, MUST, M1.3)

**Given** a non-empty affected-row set from AC-NS2-002 row 002b and `sessionID: "test-session-001"`,
**When** the Detect branch completes,
**Then** ALL THREE sub-assertions hold:

- **(a) systemMessage emitted, advisory**: the `HookOutput.SystemMessage` field carries a multi-line string beginning with `Navigator Detect:` and naming the changed path + ≤10 affected rows, AND the `HookOutput.HookSpecificOutput.Decision` (if present) is `"allow"` (never `"block"`).
- **(b) JSONL impact record appended**: the file `.moai/state/navigator-detect/test-session-001.jsonl` exists and its last line is valid JSON with keys `changed_path`, `changed_at`, `affected_nodes` (array), `affected_edges` (array), and each `affected_edges` element carries `edge_type`, `source_node`, `target_node`, `source_path`, `line_number`.
- **(c) no work-item promotion**: no GitHub issue is created, no `.moai/specs/` file is created or modified, no TODO file is created, and no source file is mutated. The only writes are the JSONL append (`.moai/state/navigator-detect/`) and the advisory log line (`.moai/logs/navigator-sync.log`).

### AC-NS2-004 — fail-open across 5 error modes (REQ-NS2-004, MUST, M1.4)

**Given** any of the 5 failure modes in the table below,
**When** a Write PostToolUse fires,
**Then** for every row the Detect branch produces the expected `exit-0` behavior and the expected `advisory-or-silent` output. In ALL 5 cases: `Decision: "allow"`, no user-facing error, no cascade into sibling PostToolUse branches.

| case | trigger | expected exit-0 | expected advisory-or-silent |
|---|---|---|---|
| 004a graph absent | `nav-graph.json` does NOT exist at `<projectRoot>/.moai/project/navigator/nav-graph.json` | empty affected-row set, process exits 0, no user-facing error | exactly one diagnostic line appended to `.moai/logs/navigator-sync.log` containing substring `navigator-detect` |
| 004b unparseable JSON | `nav-graph.json` exists but contains invalid JSON (e.g. `{not json`) | empty affected-row set, exits 0, no user-facing error | one log line |
| 004c schema-invalid | `nav-graph.json` is valid JSON but missing the `edges` array | empty affected-row set, exits 0 | one log line |
| 004d traversal error | a graph edge has a `source_node` key that does not resolve to any node (malformed graph) | malformed edge is skipped (not fatal); well-formed edges still returned; if ALL edges malformed → empty set returned without aborting the handler | one log line per malformed edge; if ALL malformed, one summary line |
| 004e timeout | Detect branch configured with a 200ms context timeout and a graph large enough to exceed it (or a test that cancels the context mid-traversal) | partial set collected so far (possibly empty), exits 0 | NO log line (context cancellation is not an error to advertise) |

### AC-NS2-005a — consumer-only: M0 + mx byte-unchanged (REQ-NS2-005, MUST, M1.5)

**Given** the M1 run-phase diff is applied,
**When** `git diff --name-only origin/main...HEAD` is run,
**Then** the output does NOT contain any path under `internal/navigator/sync/` or `internal/mx/` (M1 consumes these, does not modify them). Command: `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/sync|mx)/' ; [ $? -eq 1 ]` (grep exit 1 = no matches = PASS).

### AC-NS2-005b — consumer-only: read is via public API (REQ-NS2-005, MUST, M1.5)

**Given** the M1 source files,
**When** `grep -rn 'os.WriteFile\|os.Rename' internal/hook/navigator_detect*.go` is run,
**Then** no write/rename call targets `internal/navigator/sync/` or `internal/mx/` paths. (The only writes are the JSONL append under `.moai/state/` and the log under `.moai/logs/`.)

### AC-NS2-006 — concurrency: atomic read during regen (REQ-NS2-006, MUST, M1.4)

**Given** a writer that has created `<graph>.tmp` and is blocked on the `NAVIGATOR_PRE_RENAME_BARRIER` (so `<graph>` still holds the PRIOR committed version),
**When** the Detect reader runs against `<graph>` for a changed path present in the prior version,
**Then** the reader returns the prior version's affected rows without error (no partial-file read, no retry, no block). The test sets the barrier, spawns reader + writer in goroutines, and asserts the reader's result matches the prior-graph expectation. This reuses the M0 barrier (`internal/navigator/sync/write.go:41-49`) — no new test hook added.

### AC-NS2-007 — ≥80% coverage, mechanically measured (REQ-NS2-007, MUST, M1.5)

**Given** the fixture corpus at `internal/hook/testdata/navigator-detect-corpus/` (N≥20 files: design-doc fragments with `@NAV:DEC`/`@NAV:SYM`, Go source with `@NAV:SYM`/`@MX:SPEC`, and 4-5 deliberately out-of-scope files) and the pre-built `nav-graph.json` fixture covering the in-scope corpus,
**When** the command `go test ./internal/hook/ -run TestNavigatorDetectCoverage -v` is run,
**Then** the test emits an observed coverage percentage (the ratio `(in-scope-mapped) / (in-scope-mapped + in-scope-unmapped)`, where out-of-scope files are excluded from BOTH numerator and denominator) and the observed percentage is `>= 80.0`. The test FAILS (non-zero exit) if the observed percentage is `< 80.0` and prints the observed value on failure. The percentage printed on the PASS line is the Evidence; the ≥80% assertion is the Claim.

**Attribution (verification-claim-integrity §2)**: the coverage number is attributable to the exact `go test … -v` invocation + its stdout; it is NOT a carried-over estimate, NOT a sampled figure, NOT a narrative "high coverage". A subsequent run on the same fixture corpus produces the same percentage (deterministic).

### AC-NS2-008 — non-overlap with predecessor chains (REQ-NS2-008, MUST, M1.5)

**Given** the M1 production source files (`internal/hook/navigator_detect*.go` excluding `*_test.go`),
**When** `grep -rn 'capability-map\.md\|audit-report\.\(md\|json\)\|capability-symbols\.\(md\|json\)\|nav-graph\.json' internal/hook/navigator_detect.go` is run,
**Then** the ONLY match for `nav-graph.json` is in a READ context (`os.ReadFile`, `os.Open`), and there are ZERO matches for `capability-map.md`, `audit-report.{md,json}`, `capability-symbols.{md,json}` as write targets. A new `internal/hook/navigator_detect_nonoverlap_test.go` encodes this assertion and carries forward the pattern from `internal/navigator/sync/nonoverlap_test.go`.

### AC-NS2-009 — integration as a new branch, NOT a forked chain (REQ-NS2-009, MUST, M1.2)

**Given** the M1 run-phase diff,
**When** the integration checks run,
**Then** BOTH sub-assertions hold:

- **(a) no forked hook chain**: `ls .claude/hooks/moai/handle-navigator-detect.sh 2>/dev/null` shows the file does NOT exist; `grep -rn 'navigator-detect\|navigator_detect' internal/cli/` shows no new `moai hook navigator-detect` subcommand registration. The only PostToolUse entry point remains `handle-post-tool.sh` → `moai hook post-tool` → `postToolHandler.Handle`.
- **(b) branch registered inside the dispatcher**: `grep -n 'runNavigatorDetect\|navigator_detect' internal/hook/post_tool.go` shows exactly ONE call site inside `postToolHandler.Handle`, gated on `ToolName == "Write" || "Edit" || "NotebookEdit"`, placed alongside (not replacing) `runAstScan` / `runMxValidation` / `runMemoryAudit` / `logEvidence`.

### AC-NS2-010 — directory-prefix fallback resolves to an affected edge (REQ-NS2-010, MUST, M1.1)

**Given** a `nav-graph.json` fixture containing edges whose `source_path` values include `/abs/project/internal/foo/bar.go` and `/abs/project/internal/foo/baz.go`,
**When** the traversal runs with `changedPath: "/abs/project/internal/foo/"` (a directory prefix, NOT a file path),
**Then** the affected-edge set contains at least one edge whose `source_path` is under that prefix (e.g. `/abs/project/internal/foo/bar.go`) — proving the directory-prefix fallback (inspired by `navigator-audit.sh heuristic_match()` last-segment resolution at `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:406-422`) actually resolves a prefix to an edge, NOT merely that a comment cites the script.

### AC-NS2-011 — template-first (REQ-NS2-011, MUST, M1.5)

**Given** the M1 run-phase diff touches any template-managed file under `.claude/`,
**When** `git diff --name-only origin/main...HEAD | grep '^internal/template/templates/'` is run,
**Then** every template-managed path modified locally ALSO appears in `internal/template/templates/` (the template source is never bypassed). AND if `internal/template/templates/.claude/settings.json` changed, `make build` regenerated `catalog.yaml` and the regenerated catalog is committed in the same PR. (If the gate is env-var-only and no template file changed, this AC reduces to: no template path in the diff, no catalog regen required — document this in the PR body.)

### AC-NS2-012 — never blocks (REQ-NS2-012, MUST, M1.4)

**Given** ANY PostToolUse invocation that triggers the Detect branch (success OR any fail-open case from AC-NS2-004 rows 004a..004e),
**When** the handler returns,
**Then** the process exits 0 AND the `HookOutput.HookSpecificOutput.Decision` (where present) is `"allow"` — NEVER `"block"`. A grep test asserts no `Decision: "block"` and no `os.Exit(2)` appears in `internal/hook/navigator_detect*.go`: `grep -n 'block\|Exit(2)' internal/hook/navigator_detect.go` returns zero matches.

## §E. Indirect Verification (cross-cutting)

- **LSP / lint gate**: `go vet ./internal/hook/...` and `golangci-lint run ./internal/hook/...` exit 0.
- **Race detector**: `go test -race ./internal/hook/ -run TestNavigatorDetect` exits 0 (the traversal + JSONL append are concurrency-safe; the atomic-read test exercises the race window).
- **Subagent boundary** (`internal/hook/CLAUDE.md` C-HRA-008): `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/navigator_detect*.go` returns zero matches — the hook runs in subagent context and never prompts.

## §F. Edge Cases (quality bar, not merge gates)

- **Empty graph**: a graph with `nodes: []`, `edges: []` → traversal returns empty, no error (distinct from absent-graph fail-open).
- **Changed path outside scan roots**: e.g. `.moai/specs/SPEC-FOO-001/spec.md` → not in any edge's `source_path`, traversal returns empty, no error (out-of-scope per §E measurement partition).
- **NotebookEdit with `notebook_path` instead of `file_path`**: the branch extracts whichever field is present; both normalized to absolute.
- **SystemMessage overflow**: > 10 affected rows → systemMessage truncates with `…and N more`, JSONL records the full set (the JSONL is the SSOT for M2).
- **Session-id rotation**: each session gets its own `<session-id>.jsonl`; no cross-session writes.
- **Same path edited repeatedly in one session**: each edit appends one JSONL line (append-only, no dedup at write time; M2 may dedup at read time).

## §G. Forward-Looking Checks (advisory, not gates)

- **M2 readiness**: the JSONL schema (AC-NS2-003 sub-assertion b) is the contract M2 Route consumes. Any M1→M2 schema drift should be caught by an M2 plan-phase audit, not an M1 AC.
- **Performance headroom**: if a future graph exceeds 10K edges and the p99 traversal exceeds 200ms, the Simplicity-ladder exception (plan.md §G AP-NS2-007) permits revisiting the linear-scan decision. M1 does NOT pre-build an index.
- **Bash trigger future**: if Claude Code later surfaces structured file paths for Bash file mutations, REQ-NS2-001 may be extended in a follow-up SPEC without reworking M1.
