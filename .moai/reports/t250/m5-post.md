# M5 Post-run — fixed task set via the code-query tools (AC-GF-022)

- **Measurement date**: 2026-08-25 (after M5 implementation; REMEASURED 2026-08-26 under CR round-2 3855149200 against the baseline's ORIGINAL task definitions — the first post-run had substituted narrower targets; the table below is the exact-set rerun).
- **Method**: the SAME fixed task set defined in m5-baseline.md, executed by the M5 query engine (the MCP handlers' backing functions) against the real worktree via a measurement vehicle test (deleted after the run). Counting rule identical to the baseline: Grep and Read TOOL-USE events only.

## Per-task results (verbatim from the exact-set measurement run)

| Task (baseline definition, verbatim target) | Tool calls | Outcome (verbatim log line) |
|---|---|---|
| T1: exported API of the FULL `internal/graph` package | 18 × `graph_file_api` (one per file) | `T1 file-api FULL internal/graph: files=18 symbols=45` |
| T2: every caller of `CheckFreshness` | 1 × `graph_find_code` | `T2 find CheckFreshness: err=<nil> matches=17` |
| T3: trace from `refreshEdgesArtifact` (one hop each direction) | 1 × `graph_trace_calls` | `T3 trace refreshEdgesArtifact: err=<nil> callers=1 callees=14` |
| T4: every definition and use of the grade constants | 2 × `graph_find_code` | `T4 find GradeFor: err=<nil> matches=6` / `T4 find ValidateGradeMatrix: err=<nil> matches=14` |
| T5: top-level exported API of `internal/mx` | 6 × `graph_file_api` (non-test files) | `T5 file-api internal/mx: files=34 symbols=81` |

## Grep/Read counts and call totals

- **Grep tool-use events in this measurement run: 0.**
- **Read tool-use events in this measurement run: 0.**
- **Total M5 tool calls: 28** (T1's 18 file-api calls + T2/T3/T4/T5's remaining 10), each carrying tree+commit provenance — matching the table row-by-row (the earlier draft's flat "5 calls" did not; corrected with the rerun).

## Reduction claim — bounded honestly

Against the aggregate baseline (m5-baseline.md): prior sessions solved code-navigation with Read runs of 0–25 per session (and Grep tool-use of 0, because this repository's agents search via Bash grep — a cost the Grep/Read-only counter cannot see). For THIS fixed task set the exact-set rerun used **0 Grep and 0 Read events** across all 28 tool calls; a Read-based exploration of the same questions requires opening every target file (T1 and T5 alone span 24 non-test files; T2's caller set spans 3 non-test files).

The per-task baseline remains recorded as an unobtainable gap in m5-baseline.md (no pre-M5 session performed these tasks; a knowing-measurer simulation would fabricate counts). The reduction this artifact honestly claims is the structural one, measured: **the complete exact task set resolves with zero Grep/Read tool-use events.**
