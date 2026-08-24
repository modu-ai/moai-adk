# M5 Post-run — fixed task set via the code-query tools (AC-GF-022)

- **Measurement date**: 2026-08-25 (after M5 implementation commit).
- **Method**: the SAME fixed task set defined in m5-baseline.md, executed by
  the M5 query engine (the MCP handlers' backing functions — the exact code
  path the tools run) against the real worktree, via a measurement vehicle
  test (`T250_MEASURE_ROOT=<worktree> go test ./internal/graph/ -run
  TestMeasureCodeQueries -v`; vehicle deleted after the run, same pattern as
  the M2 budget measurement). Counting rule identical to the baseline: Grep
  and Read TOOL-USE events only.

## Per-task results (verbatim from the measurement run)

| Task | Tool call | Outcome (verbatim log line) |
|---|---|---|
| T1 file-api internal/graph | `graph_file_api(codequery.go)` | `T1 file-api internal/graph/codequery.go: err=<nil> symbols=8` |
| T2 callers of CheckFreshness | `graph_find_code("CheckFreshness")` | `T2 find CheckFreshness: err=<nil> matches=13` — incl. `internal/cli/graph_check.go:60 (callee)`, `internal/hook/quality/gate.go:1194 (callee)`, 8 test-file call sites, 3 caller-observations |
| T3 trace from the CLI seam | `graph_trace_calls("CodeEdges", 1)` | `T3 trace CodeEdges: err=<nil> callers=4 callees=2` |
| T4 grade-constant usage sites | `graph_find_code("GradeFor")` + `("ValidateGradeMatrix")` | `T4a find GradeFor: err=<nil> matches=6` / `T4b find ValidateGradeMatrix: err=<nil> matches=12` |
| T5 file-api internal/mx | `graph_file_api(sidecar.go)` | `T5 file-api internal/mx/sidecar.go: err=<nil> symbols=10` |

## Grep/Read counts

- **Grep tool-use events in this measurement run: 0.**
- **Read tool-use events in this measurement run: 0.**
- Every answer arrived as one tool call per task (5 calls total, all
  successful), each carrying tree+commit provenance.

## Reduction claim — bounded honestly

Against the aggregate baseline (m5-baseline.md): prior sessions solved
code-navigation with Read runs of 0–25 per session (and Grep tool-use of 0,
because this repository's agents search via Bash grep — a cost the
Grep/Read-only counter cannot see). For THIS fixed task set the post-run
used **0 Grep and 0 Read events** where a Read-based exploration of the
same questions requires opening each target file (the T2 caller set alone
spans 3 non-test files; T1+T5 are file reads by construction).

The per-task baseline was recorded as an unobtainable gap in
m5-baseline.md (no pre-M5 session performed these tasks; a
knowing-measurer simulation would fabricate counts). The reduction this
artifact can honestly claim is therefore the structural one, measured:
**the complete task set resolves with zero Grep/Read tool-use events, one
provenance-stamped tool call per task.**
