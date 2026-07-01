---
id: SPEC-HAIKU-EFFORT-INERT-001
title: "Acceptance criteria — remove inert effort field from haiku-tier agents"
version: "0.1.1"
status: completed
created: 2026-07-01
updated: 2026-07-01
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tier: S
tags: "agent, effort, haiku, frontmatter, template"
---

# Acceptance — SPEC-HAIKU-EFFORT-INERT-001

## §D AC Matrix

| AC ID | REQ | Criterion | Verification (mechanical) |
|-------|-----|-----------|---------------------------|
| AC-HEI-001 | REQ-HEI-001 | Template `manager-docs.md` has NO `effort:` line and retains `model: haiku`. | `grep -c '^effort:' internal/template/templates/.claude/agents/moai/manager-docs.md` = 0 AND `grep -c '^model: haiku' ...` = 1 |
| AC-HEI-002 | REQ-HEI-001 | Template `manager-git.md` has NO `effort:` line and retains `model: haiku`. | `grep -c '^effort:' internal/template/templates/.claude/agents/moai/manager-git.md` = 0 AND `grep -c '^model: haiku' ...` = 1 |
| AC-HEI-003 | REQ-HEI-003 | Local mirror `manager-docs.md`/`manager-git.md` have no `effort:` line and keep `model: haiku`; the `^model:`/`^effort:` lines match the template source (whole-file byte-identity NOT asserted — §25 non-parity files). | `grep -c '^effort:' .claude/agents/moai/manager-docs.md` = 0 (same for `manager-git.md`); model/effort line parity: `diff <(grep -E '^model:\|^effort:' <template>) <(grep -E '^model:\|^effort:' <local>)` empty for both agents |
| AC-HEI-004 | REQ-HEI-002, REQ-HEI-006 | A guard test scans BOTH the template tree AND the local mirror; it fails when any `model: haiku` agent in either tree declares `effort:`, and passes after removal from both. | New/added test in `internal/template/`; run `go test ./internal/template/ -run <GuardTestName>` → PASS after fix |
| AC-HEI-005 | REQ-HEI-004 | `agentEffortMap` still contains exactly the 5 effort-supporting agents; no haiku agent added. | `grep -A8 'agentEffortMap = map' internal/template/model_policy.go` shows only `manager-spec`, `plan-auditor`, `sync-auditor`, `manager-develop`, `builder-harness`; `TestGetAgentEffort` PASS |
| AC-HEI-006 | REQ-HEI-007 | The 5 effort-supporting agents retain their `effort` fields unchanged. | `grep -h '^effort:' internal/template/templates/.claude/agents/moai/{manager-spec,plan-auditor,sync-auditor,manager-develop,builder-harness}.md` → 4×`xhigh` + 1×`high`, unchanged |
| AC-HEI-007 | REQ-HEI-005 | `make build` succeeds after the frontmatter edit (embedded FS recompiled). | `make build` exit 0 |
| AC-HEI-008 | REQ-HEI-004 | Generator does not re-add effort for haiku agents. | `TestApplyEffortPolicy/no_op_for_agent_not_in_effort_map` PASS |
| AC-HEI-009 | (quality gate) | Full template test suite passes; no collateral regression. | `go test ./internal/template/...` → PASS |

## §D.1 Given-When-Then Scenarios

### Scenario 1 — Frontmatter honesty (primary)
- **Given** `manager-docs.md` and `manager-git.md` declare `model: haiku` with a stray `effort: medium` line that Haiku silently ignores,
- **When** the `effort: medium` line is removed from both files in both the template source and the local mirror, and `make build` is run,
- **Then** neither file declares an `effort` field, both retain `model: haiku`, `make build` exits 0, and the embedded FS carries the corrected frontmatter.

### Scenario 2 — Regression guard
- **Given** a guard test that scans `internal/template/templates/.claude/agents/moai/*.md` frontmatter,
- **When** any file declaring `model: haiku` also declares an `effort:` field,
- **Then** the guard test fails; and after the two stray lines are removed, the guard test passes.

### Scenario 3 — No collateral damage to effort-supporting agents
- **Given** the 5 `model: inherit` agents (`manager-spec`, `plan-auditor`, `sync-auditor`, `manager-develop` at `xhigh`; `builder-harness` at `high`),
- **When** the haiku-agent fix is applied,
- **Then** all 5 retain their exact `effort` values, and `agentEffortMap` is unchanged (`TestGetAgentEffort` passes).

## §D.2 Edge Cases

- **Mirror drift:** editing only one tree (template OR local) leaves the two inconsistent. Both trees MUST be edited; AC-HEI-003 checks the local side explicitly.
- **`make build` not run:** the embedded binary would still carry the old frontmatter even after source edit; AC-HEI-007 gates on `make build` success.
- **Guard-test false negative:** a guard that only checks `manager-docs`/`manager-git` by name would miss a future haiku agent. The guard MUST key on the `model: haiku` field, not on the agent name (REQ-HEI-002 phrasing).

## §D.3 Quality Gate Criteria

- `go test ./internal/template/...` — all pass (includes the new guard + existing effort tests).
- `go vet ./internal/template/...` — clean.
- No change to any file outside: the 2 haiku agent files (× 2 trees) and 1 test file.

## §D.4 Definition of Done

- [ ] AC-HEI-001 .. AC-HEI-009 all satisfied with mechanical evidence.
- [ ] `^model:`/`^effort:` lines identical across both trees for the two haiku agents (whole-file byte-identity NOT required — §25 non-parity files).
- [ ] `make build` succeeds; embedded FS reflects the removal.
- [ ] Guard test added and passing; existing effort tests still passing.
- [ ] No implementation details leaked into spec.md; no out-of-scope files touched.
