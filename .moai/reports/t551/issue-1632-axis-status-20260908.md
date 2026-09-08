# GH #1632 — per-axis status in this tree (card t551)

**Tree**: `.claude/worktrees/t551` · **HEAD**: `3ac58b5a1` · **Measured**: 2026-09-08

The lead's dispatch required the external report's premise be re-measured here before any
repair. It was. The issue carries five axes; four already have landed work, and one is still
live. That is the finding — not a preamble to it.

## Claim

Four of #1632's five axes are already repaired at `3ac58b5a1`. One — a blank codex review body
read as `pass` — remains live and reachable through the production path. Card t551 scopes to
that one.

## Evidence

Per-axis, each citation re-read in this tree at this HEAD:

| Axis | Issue text | Status here | Citation |
|---|---|---|---|
| 1 | `findings[]` always empty; findings arrive only as `summary` prose | **closed** (card t234) | `internal/cli/mcp_codex.go:1473` `codexFindingsOf` parses `- [P1] …` bullets into `Finding{Severity,Title,Body,File,Line}` |
| 2 | envelope `verdict` unreliable in both directions | **closed** | `mcp_codex.go:1435` `synthesizeReviewOutput` adopts the most conservative of all signals; divergence recorded in `SynthesisNote` (field `:276`, assigned `:1446`) |
| 3 | `gates.codex: required` not enforced | **partly closed** — the unmet state is now a structured field, not enforcement | `mcp_codex.go:291` (`GateUnmet` field), `:1608` (`applyGateUnmet`) |
| 4 | `codex_setup` reports `auth_provider: "unknown"` while codex works | **closed** | `mcp_codex.go:1848` — comment names "#1632 axis 4" and mirrors codex's own inference |
| 5 | `codex_task` writes a spurious timeout record instantly | **closed** (card t514 / GH #1687) | `internal/cli/codex_task.go:71` `codexTaskCallerEndedMessage` separates a caller-ended turn from the tool's own bound |

Live hole, measured end-to-end through `runCodexReviewRPC` with a stubbed codex conn
(`.moai/reports/t551/probe-reachability-20260908.txt`):

```
exactly-empty      verdict="inconclusive" summary="codex unavailable: codex review produced no verdict text"
whitespace-only    verdict="pass"         summary=""      findings=0
newline-only       verdict="pass"         summary=""      findings=0
real-clean-review  verdict="pass"         summary="The change introduces no blocking issues."
```

Three emptiness checks use exact equality, so a whitespace-only body counts as content:
`mcp_codex.go:1134` (collection), `:1253` (selection), `:817` (the guard). The blank body then
reaches the synthesizer, whose native-mode default for an unrecognized body is `pass`
(`:1409`).

Mechanical consumer: `internal/cli/mcp_convergence.go:190` — `allPass(required)` yields
`overall_verdict: pass`, so one blank codex output passes a `required` gate.

## Baseline-attribution

Every row above was read from the working tree at `git rev-parse HEAD` = `3ac58b5a1` in
`.claude/worktrees/t551`, in this run. The probe outputs are the verbatim stdout of
`go test ./internal/cli/ -run TestT551Probe… -v -count=1` executed in this tree; both files are
under `.moai/reports/t551/`. No figure here is carried over from another tree, package, or run.

## Gaps

- Axis 3 is recorded as **partly** closed on the strength of the `GateUnmet` annotation alone.
  Whether any consumer BLOCKS on that field was not measured — only that the field is produced.
  The issue's own complaint ("`required` reads as a guarantee but behaves as a suggestion") is
  therefore not established as fully answered, and no live queue card owns the remainder.
- Axes 1, 2, 4 and 5 were judged closed by reading the code and its provenance comments. They
  were NOT re-exercised against a live codex binary; no live-backend run was performed in this
  tree.
- The GLM backend's own blank-output handling (`internal/cli/mcp_glm.go:342` `parseGLMReview`)
  was read and appears to fail closed to `inconclusive`, but this was not probed. It is out of
  this card's scope either way.
- The queue was searched for sibling cards naming #1632 (`moai todo list --json`, live items
  and archived): none found besides t551.

## Residual-risk

- Judging axes closed from provenance comments risks trusting an annotation over behaviour: a
  comment naming an axis proves someone intended to close it, not that it is closed. Axes 1 and
  2 have dedicated tests in `codex_findings_parse_test.go` and `audit_blind_verdict_test.go`;
  axes 4 and 5 were accepted on the code alone.
- The whitespace-only body is demonstrated reachable through a stubbed conn. Whether codex-cli
  actually emits one in the field is not established — the repair is justified by the guard
  being wrong, not by a field sighting.
