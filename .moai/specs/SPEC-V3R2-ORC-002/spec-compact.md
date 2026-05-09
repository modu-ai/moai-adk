---
spec_id: SPEC-V3R2-ORC-002
purpose: "Compact reference auto-extracted from spec.md + plan.md + acceptance.md"
created_at: 2026-05-10
---

# SPEC-V3R2-ORC-002 — Compact Reference

Quick-reference compact for SPEC-V3R2-ORC-002 (Agent Common Protocol CI
Lint — `moai agent lint`). For deep context, see `spec.md`, `plan.md`,
`acceptance.md`, `research.md`, and `tasks.md` in this directory.

---

## Goal

Ship `moai agent lint` as a build-time Go-binary subcommand + CI step that
scans every `.claude/agents/moai/*.md` (template + local trees) for 8 lint
rules (LR-01..LR-08) enforcing `agent-common-protocol.md` §User
Interaction Boundary [HARD] rules. Concurrently extract the 6-bullet
Skeptical-Evaluator Mandate block into `agent-common-protocol.md`
§"Skeptical Evaluation Stance" (EVOLVABLE) and remove duplicates from
manager-quality + evaluator-active.

---

## Scope

### In-Scope

- `internal/cli/agent_lint.go` — lint engine (frontmatter parser, body
  scanner with fence-state tracking, 8-rule dispatch).
- `internal/cli/agent_lint_test.go` — TDD tests + 9 testdata fixtures.
- `moai agent lint` cobra subcommand registration in `internal/cli/root.go`.
- 8 lint rules: LR-01 (literal AskUserQuestion), LR-02 (Agent in tools),
  LR-03 (missing effort, warning), LR-04 (dead hook), LR-05 (missing
  isolation:worktree, warning), LR-06 (--deepthink boilerplate, warning),
  LR-07 (duplicate Skeptical block), LR-08 (skill-preload drift, warning).
- `--strict` mode promoting LR-03/05/06/08 from warning to error.
- `--format=json` output with stable v1.0 schema.
- CI integration in `.github/workflows/ci.yaml`.
- Skeptical Stance section added to `agent-common-protocol.md` (EVOLVABLE
  per CON-001 zone model).
- expert-security tools list cleanup (drop dead `Agent` token).
- Pre-commit hook documentation snippet.

### Out-of-Scope (Exclusions)

1. ORC-002 SPEC self does NOT rewrite the 9 LR-01-violating agent
   bodies — ORC-001 / MIG-001 own those.
2. Effort-level matrix authoring — SPEC-V3R2-ORC-003.
3. `isolation: worktree` mandatory enforcement — SPEC-V3R2-ORC-004.
4. Hook handler implementation — SPEC-V3R2-RT-001 / RT-006.
5. Modifying FROZEN sections of agent-common-protocol.md (only the new
   EVOLVABLE §Skeptical Evaluation Stance section is added).
6. Adding new subagents or skills.
7. Runtime check during agent spawn (build-time only).
8. Lint rules for skill files / command files / hook wrappers.
9. AskUserQuestion occurrences inside fenced code blocks (REQ-015).
10. Performance optimisation beyond O(N) file pass.
11. Lint config file (`.moai-agent-lint.yaml`) — future SPEC.
12. Goroutine parallelism — only if CI exceeds 1s budget.
13. IDE plugins consuming `--format=json` — community contribution.

---

## Requirements (17 REQs)

| REQ-ID | Type | One-liner |
|--------|------|-----------|
| REQ-ORC-002-001 | Ubiquitous | Binary exposes `moai agent lint` with --path/--format/--strict/--help |
| REQ-ORC-002-002 | Ubiquitous | Default scan = template tree + local tree |
| REQ-ORC-002-003 | Ubiquitous | Engine implements exactly 8 lint rules (LR-01..LR-08) |
| REQ-ORC-002-004 | Ubiquitous | Exit codes: 0 clean, 1 error, 2 malformed, 3 IO; --strict promotes warnings to 1 |
| REQ-ORC-002-005 | Ubiquitous | agent-common-protocol.md gains §Skeptical Evaluation Stance (EVOLVABLE) |
| REQ-ORC-002-006 | Event-Driven | LR-01: literal AskUserQuestion outside fence → error record with line + 2-line context |
| REQ-ORC-002-007 | Event-Driven | LR-02: `Agent` token in tools CSV → error record with exact match |
| REQ-ORC-002-008 | Event-Driven | LR-04: hook matcher refs absent tool → error record |
| REQ-ORC-002-009 | Event-Driven | LR-07: duplicate Skeptical block beyond canonical → error |
| REQ-ORC-002-010 | Event-Driven | --format=json emits version/summary/violations document |
| REQ-ORC-002-011 | State-Driven | CI required-status check; non-zero blocks PR merge |
| REQ-ORC-002-012 | State-Driven | --strict promotes LR-03/05/06/08 to errors |
| REQ-ORC-002-013 | Optional | Pre-commit hook configuration documented |
| REQ-ORC-002-014 | Optional | JSON `version` field stable through v3.0.0 minor versions |
| REQ-ORC-002-015 | Unwanted | LR-01 NOT emitted inside triple-backtick fences |
| REQ-ORC-002-016 | Unwanted | Malformed YAML → exit 2 PARSE_ERROR; other files continue |
| REQ-ORC-002-017 | Unwanted | Tree-set divergence → LINT_TREE_DRIFT warning |

---

## Acceptance Criteria (14 ACs)

| AC-ID | Verifies | Verification handle |
|-------|----------|---------------------|
| AC-V3R2-ORC-002-01 | --help surface | `moai agent lint --help`, exit 0, all 4 flags |
| AC-V3R2-ORC-002-02 | Baseline v2.13.2 violations | json: 9 LR-01, 4-5 LR-02, 1 LR-07; exit 1 |
| AC-V3R2-ORC-002-03 | Post-cleanup clean | exit 0, summary.errors == 0 |
| AC-V3R2-ORC-002-04 | JSON schema + jq parity | jq parses; summary matches violations |
| AC-V3R2-ORC-002-04.a | JSON version "1.0" | `jq -r '.version' == "1.0"` |
| AC-V3R2-ORC-002-04.b | Pre-commit doc | YAML snippet in `.moai/docs/agent-lint.md` |
| AC-V3R2-ORC-002-05 | New violation fails CI | inject AskUserQuestion → CI red |
| AC-V3R2-ORC-002-06 | LR-04 dead hook | fixture-lr04 → exit 1 with LR-04 |
| AC-V3R2-ORC-002-07 | LR-07 duplicate | fixture-lr07 → exit 1 with LR-07 |
| AC-V3R2-ORC-002-08 | LR-03 non-strict | warning only, exit 0 |
| AC-V3R2-ORC-002-09 | LR-03 strict | exit 1 |
| AC-V3R2-ORC-002-10 | Fenced-code exempt | fixture-fence-ok → exit 0 |
| AC-V3R2-ORC-002-11 | Malformed YAML | exit 2 with PARSE_ERROR |
| AC-V3R2-ORC-002-12 | Canonical Skeptical | rule file == 1 occurrence |
| AC-V3R2-ORC-002-13 | manager-brain carve-out | 0 LR-01 despite 11+ matches |
| AC-V3R2-ORC-002-14 | Tree drift | LINT_TREE_DRIFT warning emitted |

REQ ↔ AC traceability table in acceptance.md.

---

## Files to Modify

### NEW (5 files)

- `internal/cli/agent_lint.go` (lint engine, ~500-700 lines final)
- `internal/cli/agent_lint_test.go` (~400-600 lines, 14 test functions)
- `internal/cli/testdata/agent_lint/` (9 fixture files)
- `.moai/docs/agent-lint.md` (subcommand documentation)
- `internal/template/templates/.moai/docs/agent-lint.md` (mirror)

### MODIFIED — Source code (1 file)

- `internal/cli/root.go` (subcommand registration, +1 line)

### MODIFIED — Rule + zone-registry (2 files + mirrors)

- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`
  (gains §Skeptical Evaluation Stance section + 6 canonical bullets)
- `internal/template/templates/.claude/rules/moai/core/zone-registry.md`
  (new EVOLVABLE entry CONST-V3R2-049 for §Skeptical Evaluation Stance)
- Local mirrors via `make build`.

### MODIFIED — Agent files (3 files + mirrors)

- `internal/template/templates/.claude/agents/moai/manager-quality.md`
  (Skeptical Mandate block deleted, replaced with reference link)
- `internal/template/templates/.claude/agents/moai/evaluator-active.md`
  (same surgery)
- `internal/template/templates/.claude/agents/moai/expert-security.md`
  (drop dead `Agent` from tools list)
- Local mirrors via `make build`.

### MODIFIED — CI workflow (1 file)

- `.github/workflows/ci.yaml` (+3-line Lint job step)

### REGENERATED

- `internal/template/embedded.go` (via `make build` after rule + agent
  edits)

### NOT MODIFIED (intentionally)

- `spec.md` (preserved per task constraint; HISTORY amendment deferred to
  run phase M5.2)
- 9 LR-01-violating agent bodies (handled by ORC-001 / MIG-001)
- 4 remaining LR-02 agents (builder-{agent,skill,plugin}, expert-mobile —
  ORC-001 retires the 3 builders; expert-mobile is a follow-up)
- All Go code outside `internal/cli/agent_lint*` and `internal/cli/root.go`

---

## Lint Rule Severity Ladder

| Rule | Default | --strict | Notes |
|------|---------|----------|-------|
| LR-01 | error | error | No escape hatch (safety-critical) |
| LR-02 | error | error | No escape hatch |
| LR-03 | warning | error | Promoted by ORC-003 in future |
| LR-04 | error | error | Dead-hook config; companion warning LR-04-COMPLEX-MATCHER for regex |
| LR-05 | warning | error | Promoted by ORC-004 in future |
| LR-06 | warning | error | --deepthink boilerplate noise |
| LR-07 | error | error | Duplicate Skeptical block |
| LR-08 | warning | warning | Skill preload drift, never strict-promoted in this SPEC |
| Drift | warning | warning | LINT_TREE_DRIFT, LINT_TREE_FILE_MISMATCH |
| Parse | exit 2 | exit 2 | Per-file isolated error; other files continue |

---

## Dependencies

- **Blocked by**: SPEC-V3R2-CON-001 (zone registry — merged), SPEC-V3R2-ORC-001 (PR #811 merged 2026-05-09; provides v3r2 clean baseline target for AC-03).
- **Blocks**: SPEC-V3R2-ORC-003 (LR-03 promotion to error after effort
  matrix), SPEC-V3R2-ORC-004 (LR-05 promotion to error after worktree
  mandate), SPEC-V3R2-MIG-001 (legacy SPEC rewriter validates output via
  this lint).
- **Related**: SPEC-V3R2-CON-002 (constitutional amendment protocol —
  §Skeptical Evaluation Stance addition runs through 5-layer safety gate),
  SPEC-V3R2-CON-003 (constitution consolidation — same rule file but
  different scope).
- **Carry-over from**: SPEC-V3R2-CON-001 (FROZEN/EVOLVABLE codification),
  SPEC-V3R2-RT-004 (state subcommand registration pattern reused),
  SPEC-ASKUSER-ENFORCE-001 (canonical AskUserQuestion protocol — the rule
  this SPEC enforces).

---

## Wave Position

Wave 10 Tier 2 SPEC for V3R2 ORC series. Builds on ORC-001 (PR #811
merged) by enforcing the common-protocol rules ORC-001 cleaned up. Blocks
ORC-003 (effort matrix) and ORC-004 (worktree mandate) downstream upgrades
to error severity.

Parallel SPEC in Wave 10 Tier 2: SPEC-V3R2-SPC-003 (separate work).

---

## Methodology

- **Development mode**: TDD (per `.moai/config/sections/quality.yaml`)
- **TRUST 5**: Tested (≥85% coverage), Readable (Go idioms +
  golangci-lint clean), Unified (CSV CONVENTIONS + table-driven
  rule dispatch), Secured (no fs writes outside scan paths; no network
  calls), Trackable (8 @MX tags planned in plan.md §3)
- **Branch**: `feature/SPEC-V3R2-ORC-002-agent-lint` from
  `origin/main` HEAD `ab0fc4dda`
- **Worktree**: `/Users/goos/.moai/worktrees/moai-adk-go/orc-002-plan`
- **Template-First**: all agent + rule edits start under
  `internal/template/templates/`; `make build` mirrors

---

End of spec-compact.md.
