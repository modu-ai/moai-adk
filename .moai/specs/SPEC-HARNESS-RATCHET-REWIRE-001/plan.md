---
id: SPEC-HARNESS-RATCHET-REWIRE-001
title: "Wire Failure Signals into the Harness Learning Loop — Implementation Plan"
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "harness, learning-loop, failure-signals, classifier-eligibility, proposals, lessons-inbox, doctor, workflow-reflex, plan"
---

# SPEC-HARNESS-RATCHET-REWIRE-001 — Plan

> plan.md is the derived execution plan. WHAT/WHY SSOT is spec.md. This document carries the HOW skeleton (milestones / constraints / risks); function names and signatures are run-phase discretion.

## §A Context

### §A.1 Problem summary

The harness learning loop observes lifecycle events (usage-log.jsonl, 15,763 lines live) but: (1) failure signals never enter it — PostToolUseFailure handling and evidence-writer test-fail capture exist yet feed no classifier/proposal path; (2) the Stop-path auto-classifier promotes degenerate lifecycle-noise keys (`subagent_stop:unknown:` etc., empty context_hash, confidence 1); (3) the back half has zero lifetime throughput — `.moai/harness/proposals/` never existed, `learnings_count: 0`, propose is manual-trigger only. The only working ratchet is fully manual.

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash/Read, vci §2 attribution)

```
wc -l .moai/harness/usage-log.jsonl                              → 15,763 lines (~3.25 MB, mtime 2026-07-09 — live)
wc -l .moai/harness/learning-history/tier-promotions.jsonl       → 16 entries (all ts 2026-05-24)
ls .moai/harness/proposals                                       → No such file or directory (never existed)
ls .moai/evolution/learnings                                     → .gitkeep only (empty)
grep learnings_count .moai/evolution/manifest.yaml               → learnings_count: 0
grep -n auto_apply .moai/config/sections/harness.yaml            → 116: auto_apply: false
grep -n "PostToolUseFailure" internal/hook/coverage_table.go     → :46 {IsActive: true, HandlerFile: "post_tool_failure.go"}
ls internal/hook/evidence_writer.go                              → exists (15,431 bytes)
ls internal/cli/harness/{propose,doctor}.go                      → both exist (propose 5,724 B; doctor 9,169 B)
```

Content anchors (line numbers indicative; re-verify at run-phase per line-drift asymmetry lesson):
- `internal/cli/hook.go` — `classifyHarnessPatterns(root)` chained after `RecordExtendedEvent` on the Stop path, comment `SPEC-HARNESS-EVO-PIPE-REPAIR-001 REQ-HEP-003` (observed at lines 765-776). This is the pattern M2 mirrors for propose.
- `tier-promotions.jsonl` degenerate samples: `subagent_stop:unknown:` (count 41, confidence 1, to_tier auto_update), `session_stop::` (count 23), `agent_invocation:Bash:` (observation tier).

NOT yet located (pre-flight item): applier.go / rubric.go / regression_gate.go were NOT found in `internal/cli/harness/` — likely `internal/harness/`. Locate before M2 (they are PRESERVE surfaces, not targets).

### §A.3 Approach — three milestones, Go-primary (TDD), one doc deliverable

- **M1 — failure-event recording (Go, TDD)**: extend the PostToolUseFailure handler + evidence-writer test-fail path to emit `tool_failure:<tool>:<signature>` / `test_fail:<package>:` events through the existing observer (`RecordExtendedEvent` or equivalent). New event-key constants live beside the existing event-type taxonomy.
- **M2 — classifier eligibility + propose auto-run (Go, TDD)**: (a) add an eligibility predicate to the tier classifier excluding empty-context/unknown-subject keys, with a regression test proving `subagent_stop:unknown:` no longer promotes; (b) chain proposal generation after `classifyHarnessPatterns` on the Stop path when promotions > 0, fail-open, reusing the `moai harness propose` generation path.
- **M3 — lessons inbox + doctor + doctrine (Go + doc, TDD)**: (a) failure handler appends structured stubs to `.moai/lessons-inbox.jsonl`; (b) `moai harness doctor` gains the "promotions exist but proposals dir absent" dormancy check; (c) moai-constitution.md § Lessons Protocol gains the inbox-drain cross-reference (template-first).

### §A.4 Tier evidence (M)

- Files affected: ~8-12 (hook.go, post_tool_failure.go, evidence_writer.go, classifier file(s) in internal/harness, propose reuse seam, doctor.go, new inbox writer, constitution rule + template mirror, tests) — within Tier M's 5-15 band.
- LOC estimate: 400-800 (three small Go features + tests + one doc edit) — within Tier M's 300-1000 band.
- No constitutional surface, no >15-file cascade → not Tier L. Multiple packages + hook-budget risk → not Tier S.

### §A.5 PRESERVE / EXTEND map

| Surface | Disposition |
|---------|-------------|
| `internal/cli/hook.go` Stop path (classify chain) | EXTEND (append propose chain after classify) |
| `internal/hook/post_tool_failure.go` | EXTEND (emit failure event + inbox stub) |
| `internal/hook/evidence_writer.go` | EXTEND (emit test_fail event on fail detection) |
| tier classifier (internal/harness) | EXTEND (eligibility predicate) |
| `internal/cli/harness/propose.go` | REUSE (generation path invoked from Stop chain — refactor to callable seam if needed, semantics unchanged) |
| `internal/cli/harness/doctor.go` | EXTEND (one new check) |
| applier / rubric / regression gate | PRESERVE (untouched) |
| `harness.yaml` (`auto_apply: false`, FROZEN validation) | PRESERVE (asserted, not changed) |
| `.moai/harness/*` runtime artifacts | PRESERVE (tests in t.TempDir() only) |
| `moai-constitution.md` § Lessons Protocol | EXTEND (cross-ref paragraph; template-first) |

## §B Known Issues (filtered, Tier M)

- **B1 Cross-platform**: no syscall usage expected; still verify `GOOS=windows GOARCH=amd64 go build ./...`.
- **B2 Cross-SPEC conflicts**: this package carries heavy SPEC history (PIPE-REPAIR, PROPOSAL-GEN, LOOP-CLOSURE, LEARNER-FIX). Grep for retired/superseded markers in `internal/harness` before extending the classifier; REQ-HEP-003's classify chain is the explicit mirror precedent, not a conflict.
- **B3 C-HRA-008**: `internal/harness/`, `internal/hook/` are subagent-domain code — AskUserQuestion grep must stay 0; extend the existing `subagent_boundary_test.go` pattern if new files are added.
- **B7 capture path resolution**: observer path resolution uses project root; prefer `$CLAUDE_PROJECT_DIR`/root-derived paths, never `os.Getwd()` bare fallback (known `internal/hook/.moai/` leak anomaly).
- **B8 Working-tree hygiene**: never commit runtime-managed `.moai/harness/*` deltas; the live usage-log will grow during the session — exclude from `git add` (specific-path staging only).
- **B10 Scope discipline**: three parallel Workflow-Reflex SPECs exist; do not touch model-routing or loop.md surfaces from this SPEC.

## §C Pre-flight checklist

```bash
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5          # lint baseline (NEW vs pre-existing)
grep -rn "classifyHarnessPatterns" internal/cli/hook.go # M2 anchor re-verification
ls internal/harness/ | grep -iE 'applier|rubric|regression|classif|tier'   # locate PRESERVE + EXTEND surfaces
grep -rn "Retired\|superseded" internal/harness internal/cli/harness | head -5
go test ./internal/harness/... ./internal/hook/... ./internal/cli/... 2>&1 | tail -5  # test baseline
```

## §D Constraints + open design decisions (run-phase)

Constraints: see spec.md §Constraints (fail-open + 5s hook budget; auto_apply false; C-HRA-008; runtime-file hygiene; frozen surfaces untouched).

Open decisions (run-phase discretion, record in progress.md when resolved):

1. **D1 — failure signature derivation** (REQ-HRR-001): error-class token vs truncated error-text hash for `<signature>`. Recommendation: low-cardinality error-class token (hash of full error text explodes key cardinality and defeats pattern aggregation).
2. **D2 — propose invocation seam** (REQ-HRR-004): extract propose.go's generation core into a callable function reused by both the CLI command and the Stop chain (recommended), vs shelling out to `moai harness propose` (rejected: subprocess cost inside 5s hook budget).
3. **D3 — inbox stub schema** (REQ-HRR-006): minimum fields ts / event_key / summary / source; decide drained-marking (in-place `drained: true` rewrite vs companion offset file). Recommendation: keep JSONL append-only; drain marking via companion state file to preserve append-only discipline.
4. **D4 — eligibility predicate placement** (REQ-HRR-003): filter at classification time (key never becomes a promotion candidate — recommended) vs at promotion-write time.

## §E Self-Verification (run-phase deliverables)

Per manager-develop-prompt-template.md §E (E1-E7), each reported per vci 5-section format:
- E1: AC matrix (acceptance.md §D) binary PASS/FAIL with verification commands + verbatim output.
- E2: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- E3: `go test -cover` for `internal/harness`, `internal/hook`, `internal/cli/harness` (≥85% package target).
- E4: subagent-boundary grep 0 matches (C-HRA-008).
- E5: golangci-lint — NEW findings vs baseline distinguished.
- E6: commit SHAs + push state (Route A main-direct, Conventional Commits `feat(SPEC-HARNESS-RATCHET-REWIRE-001): M{N} ...`).
- E7: blocker report if any decision exceeds delegated authority.

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|----------------|
| M1 — failure-event recording | post_tool_failure.go + evidence_writer.go emit `tool_failure:*` / `test_fail:*` observer events (TDD) | REQ-HRR-001, REQ-HRR-002 | AC-HRR-001, AC-HRR-002 PASS |
| M2 — eligibility + propose auto-run | classifier eligibility predicate + regression test; Stop-path propose chain (fail-open) landing files in `.moai/harness/proposals/` | REQ-HRR-003, REQ-HRR-004, REQ-HRR-005 | AC-HRR-003..007 PASS |
| M3 — inbox + doctor + doctrine | `.moai/lessons-inbox.jsonl` writer; doctor dormancy check; constitution § Lessons Protocol cross-ref (template-first) | REQ-HRR-006, REQ-HRR-007, REQ-HRR-008, REQ-HRR-009 | AC-HRR-008..011 PASS |

Ordering rationale: M1 produces the failure events M2's classifier consumes; M3 is additive surface work independent of M2 internals but sequenced last to reuse M1's failure-handler seam.

## §G Anti-Patterns (do NOT)

- Auto-applying proposals or flipping `auto_apply` — the apply gate is the load-bearing safety boundary.
- Blocking session end on propose errors (exit ≠ 0 from the Stop hook path) — mirrors the death-spiral hazard; fail-open is mandatory.
- Testing against the live `.moai/harness/usage-log.jsonl` — t.TempDir() only.
- Hashing full error text into the event key (cardinality explosion → every failure becomes a unique unlearnable pattern).
- Deleting/rewriting historical tier-promotions.jsonl entries (append-only runtime artifact).
- Touching applier/rubric/regression-gate semantics "while here".

## §H Cross-References

- spec.md (SSOT), acceptance.md (AC matrix), progress.md (§E lifecycle skeleton).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` (§B/§C/§E injected above).
- SPEC-HARNESS-EVO-PIPE-REPAIR-001 (classify-at-Stop precedent), SPEC-V3R6-HARNESS-PROPOSAL-GEN-001 (propose origin).
- `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol (M3 doc target; template mirror verified present).
