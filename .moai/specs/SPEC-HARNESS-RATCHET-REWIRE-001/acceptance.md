---
id: SPEC-HARNESS-RATCHET-REWIRE-001
title: "Wire Failure Signals into the Harness Learning Loop — Acceptance Criteria"
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
tags: "harness, learning-loop, failure-signals, classifier-eligibility, proposals, lessons-inbox, doctor, workflow-reflex, acceptance"
---

# SPEC-HARNESS-RATCHET-REWIRE-001 — Acceptance Criteria

> Observable, testable assertions derived from spec.md §Requirements (GEARS). Each AC traces to one or more REQs and one audit finding (H1/H2/L2).

## §D AC Matrix

| AC ID | REQ trace | Finding | Severity | Description |
|-------|-----------|---------|----------|-------------|
| AC-HRR-001 | REQ-HRR-001 | H1 | MUST-PASS | PostToolUseFailure emits `tool_failure:<tool>:<signature>` event to usage-log |
| AC-HRR-002 | REQ-HRR-002 | H1 | MUST-PASS | Evidence-writer test-fail detection emits `test_fail:<package>:` event |
| AC-HRR-003 | REQ-HRR-003 | H2 | MUST-PASS | Regression test: `subagent_stop:unknown:` (empty context_hash, unknown subject) is NOT promoted |
| AC-HRR-004 | REQ-HRR-003 | H2 | MUST-PASS | Eligibility predicate excludes empty-context/empty-subject keys generally (`session_stop::`, `user_prompt::`) |
| AC-HRR-005 | REQ-HRR-004 | L2 | MUST-PASS | Stop path chains propose after classify when promotions > 0; proposal files land in `.moai/harness/proposals/` |
| AC-HRR-006 | REQ-HRR-004 | L2 | MUST-PASS | Propose chain is fail-open: injected propose error → stderr log, session end NOT blocked (exit 0) |
| AC-HRR-007 | REQ-HRR-005 | L2 | MUST-PASS | Human apply gate preserved: `auto_apply: false` unchanged; no apply invocation on the Stop chain (grep 0) |
| AC-HRR-008 | REQ-HRR-006 | H1/L2 | MUST-PASS | Failure event appends structured stub to `.moai/lessons-inbox.jsonl` with schema minimum fields |
| AC-HRR-009 | REQ-HRR-007 | L2 | MUST-PASS | moai-constitution.md § Lessons Protocol cross-references the inbox drain (live + template mirror) |
| AC-HRR-010 | REQ-HRR-008 | L2 | MUST-PASS | Doctor emits dormancy warning when promotions ≥1 AND proposals dir absent |
| AC-HRR-011 | REQ-HRR-009 | — | MUST-PASS | Template-first applied to doctrine edit; Go code not templated |
| AC-HRR-012 | (constraint) | — | MUST-PASS | C-HRA-008 subagent-boundary grep 0 matches on touched packages |
| AC-HRR-013 | (gate) | — | MUST-PASS | Full suite + cross-platform build green; touched-package coverage ≥85% |

## §D.1 Severity Classification

All 13 ACs are MUST-PASS. The scope is a safety-adjacent pipeline (learning loop feeding an apply gate); no SHOULD-PASS/NICE-TO-HAVE tier is defined — every criterion is load-bearing for either the ratchet function (001-010) or the repo's standing constraints (011-013).

## §D.2 Given-When-Then Scenarios

### AC-HRR-001 — tool-failure event recorded

**Given** a `t.TempDir()` project root with an initialized harness observer log path
**When** the PostToolUseFailure handler processes a failure input (tool name `Bash`, a representative error)
**Then** `.moai/harness/usage-log.jsonl` (temp) gains exactly one event whose key matches `tool_failure:Bash:<signature>` with non-empty signature, and the handler exits 0.

### AC-HRR-003 — degenerate key no longer promotes (regression)

**Given** a temp usage-log seeded with ≥ the promotion-threshold count of `subagent_stop:unknown:` events (empty context hash, subject `unknown`) — the exact shape observed in the live 2026-05-24 promotions
**When** the tier classifier runs (same path as the Stop-hook `classifyHarnessPatterns`)
**Then** the resulting tier-promotions log contains ZERO entries for `subagent_stop:unknown:`, `session_stop::`, and `user_prompt::`, while a control non-degenerate key (e.g. `agent_invocation:Bash:<hash>` with non-empty context) still promotes.

### AC-HRR-005 — propose auto-run lands proposals

**Given** a temp project where the classify pass will produce ≥1 eligible promotion and `.moai/harness/proposals/` does not exist
**When** the Stop-path hook handler runs to completion
**Then** `.moai/harness/proposals/` exists and contains ≥1 generated proposal file, and the handler completes within the hook budget.

### AC-HRR-006 — fail-open on propose error

**Given** the same setup with a fault injected into the propose path (e.g. proposals dir made unwritable)
**When** the Stop-path handler runs
**Then** a propose error is logged to stderr, the handler exits 0, and the observer event + classify results are still recorded (session end never blocked).

### AC-HRR-008 — lessons-inbox stub

**Given** a temp project root
**When** a `tool_failure` or `test_fail` event is recorded
**Then** `.moai/lessons-inbox.jsonl` (temp) gains one JSON line carrying at minimum: timestamp, event key, failure summary, and source identifier; the line parses as valid JSON.

### AC-HRR-010 — doctor dormancy check

**Given** a temp project with a tier-promotions.jsonl containing ≥1 entry and NO proposals directory
**When** `moai harness doctor` runs against that root
**Then** the doctor report contains a dormancy warning naming both facts ("promotions exist", "proposals dir absent"); **and given** a proposals dir IS present, the warning is NOT emitted.

## §D.3 Edge Cases

- **EC-1**: PostToolUseFailure with empty tool name → event key degrades gracefully (no `tool_failure::` degenerate key emitted — either skip or substitute `unknown-tool` with non-empty signature; MUST NOT create a new degenerate class that AC-HRR-004 would exclude).
- **EC-2**: propose auto-run when zero NEW promotions this pass → propose is skipped (no empty-proposal churn).
- **EC-3**: lessons-inbox file absent at first write → created with 0o600-class permissions consistent with sibling state files.
- **EC-4**: concurrent Stop hooks (multi-session) appending to inbox/usage-log → append-only JSONL semantics tolerate interleaving; no read-modify-write of whole files.
- **EC-5**: harness learning disabled (`isHarnessLearningEnabled` false) → neither classify nor propose chain runs (existing precondition preserved).

## §D.4 Verification Commands (per AC, indicative)

```bash
go test -run TestPostToolFailureEvent ./internal/hook/...        # AC-HRR-001/002/008, EC-1/3
go test -run TestClassifierEligibility ./internal/harness/... ./internal/cli/...   # AC-HRR-003/004
go test -run TestStopProposeChain ./internal/cli/...              # AC-HRR-005/006, EC-2/5
grep -n "auto_apply" .moai/config/sections/harness.yaml           # AC-HRR-007 (expect: false, unchanged)
grep -rn "harness apply\|ApplyProposal" internal/cli/hook.go      # AC-HRR-007 (expect: 0 matches on Stop chain)
grep -n "lessons-inbox" .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md  # AC-HRR-009/011
go test -run TestDoctorDormancy ./internal/cli/harness/...        # AC-HRR-010
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ internal/cli/harness/ | grep -v _test | grep -v '^[^:]*:[0-9]*:\s*//'  # AC-HRR-012 (expect 0)
go test ./... && GOOS=windows GOARCH=amd64 go build ./...         # AC-HRR-013
```

## §D.5 Quality Gate Criteria

- TRUST 5: Tested (TDD, RED commits precede GREEN), Readable (event-key constants named, no magic strings), Unified (gofmt/golangci-lint clean vs baseline), Secured (no secrets; inbox file perms consistent), Trackable (Conventional Commits per milestone).
- Coverage: `internal/hook`, `internal/harness`, `internal/cli/harness` touched files ≥85% package coverage.
- Lint: zero NEW golangci-lint findings vs pre-flight baseline.

## §D.6 Definition of Done

1. All 13 ACs PASS with verbatim command output recorded in the run-phase §E self-verification (vci 5-section format).
2. `.moai/harness/proposals/` is created by the auto-chain in the test environment (live-project artifact creation happens naturally at next real session Stop — NOT manufactured during the run-phase).
3. Doctrine cross-ref merged template-first; `make build` green.
4. progress.md §E.2/§E.3 populated by manager-develop; sync-phase close per 3-phase lifecycle.
