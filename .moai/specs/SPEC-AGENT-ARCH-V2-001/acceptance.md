---
id: SPEC-AGENT-ARCH-V2-001
title: "MoAI Agent Architecture v2 — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents/moai + internal/config"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "agent-arch, super-advisor, manager-design, no-haiku, 3-tier, claude-design, token-policy, acceptance"
---

# SPEC-AGENT-ARCH-V2-001 — Acceptance Criteria

> Acceptance criteria (AC) enumerate testable, observable behaviors. Each AC cites a verification command whose verbatim output is the evidence (vci §3.2). Severity: MUST-PASS / SHOULD-PASS / NICE-TO-HAVE per §D.1.

## §D AC Matrix

### AC-AA2-001 — super-advisor agent file exists (MUST-PASS)

**Given** M1 has landed
**When** the verifier runs `ls .claude/agents/moai/super-advisor.md internal/template/templates/.claude/agents/moai/super-advisor.md`
**Then** both paths are files (exit 0) AND the frontmatter parses with `name: super-advisor`, `model: inherit`, `effort: xhigh`, `permissionMode: plan`, and `description:` contains the literal `NOT for:`.

Verification:
```bash
ls .claude/agents/moai/super-advisor.md internal/template/templates/.claude/agents/moai/super-advisor.md
grep -E '^(name|model|effort|permissionMode):' .claude/agents/moai/super-advisor.md
grep -c 'NOT for:' .claude/agents/moai/super-advisor.md   # ≥ 1
```

### AC-AA2-002 — CLAUDE.md §4 ceiling 8 → 10 (MUST-PASS)

**Given** M1 has landed
**When** the verifier runs `grep -c "10 retained agents" CLAUDE.md internal/template/templates/CLAUDE.md`
**Then** both return ≥ 1 AND the Retained Agents table contains rows for `super-advisor` AND `manager-design` AND the Selection Decision Tree contains entries 10 (super-advisor) and 11 (manager-design).

Verification:
```bash
grep -c "10 retained agents" CLAUDE.md internal/template/templates/CLAUDE.md
grep -E '^\| `super-advisor`' CLAUDE.md
grep -E '^\| `manager-design`' CLAUDE.md
grep -E '^1[01]\. ' CLAUDE.md   # entries 10 + 11 in the Selection Decision Tree
```

### AC-AA2-003 — E1-E4 escalation doctrine embedded (MUST-PASS)

**Given** M1 has landed
**When** the verifier runs `grep -c "E1\|E2\|E3\|E4" .claude/rules/moai/core/agent-common-protocol.md`
**Then** the count is ≥ 4 (one per entry condition) AND the doctrine file contains a cross-reference to `super-advisor.md`.

Verification:
```bash
grep -nE 'E[1-4]' .claude/rules/moai/core/agent-common-protocol.md | head -8
grep -c 'super-advisor' .claude/rules/moai/core/agent-common-protocol.md   # ≥ 1
```

### AC-AA2-004 — manager-design agent file exists with H1-H9 verbatim (MUST-PASS)

**Given** M2 has landed
**When** the verifier runs `grep -cE '^H[1-9]' .claude/agents/moai/manager-design.md`
**Then** the count is exactly 9 (H1 through H9) AND the frontmatter matches the §04 codeblock verbatim (`tools:` includes `DesignSync`, `model: inherit`, `effort: xhigh`, `permissionMode: acceptEdits`, `isolation: worktree`, `memory: project`, `skills: [moai-domain-frontend]`).

Verification:
```bash
ls .claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/manager-design.md
grep -cE '^H[1-9]' .claude/agents/moai/manager-design.md   # exactly 9
grep -E '^(tools|model|effort|permissionMode|isolation|memory):' .claude/agents/moai/manager-design.md
grep -c 'DesignSync' .claude/agents/moai/manager-design.md   # ≥ 1 (in tools:)
```

### AC-AA2-005 — D1-D5 design pipeline workflow skill exists (MUST-PASS)

**Given** M2 has landed
**When** the verifier runs `ls .claude/skills/moai/workflows/design.md`
**Then** the file exists AND contains the 5 step headings (D1 연결 준비, D2 디자인 시스템 생성·동기화, D3 화면 결과물 생성, D4 핸드오프 수신·붙여넣기, D5 구현 연결) AND cites the DesignSync tool methods (`list_projects`, `finalize_plan`, `write_files`, `get_file`, `list_files`, `report_validate`, `create_project`, `get_project`).

Verification:
```bash
ls .claude/skills/moai/workflows/design.md
grep -cE '^### D[1-5]' .claude/skills/moai/workflows/design.md   # exactly 5
grep -c 'finalize_plan\|write_files\|report_validate' .claude/skills/moai/workflows/design.md   # ≥ 3
```

### AC-AA2-006 — spec-workflow plan→design→run conditional route (MUST-PASS)

**Given** M2 has landed
**When** the verifier runs `grep -c 'design → run\|plan → design' .claude/rules/moai/workflow/spec-workflow.md`
**Then** the count is ≥ 1 AND the conditional route is documented as additive (applies ONLY to UI-surfaced SPECs).

Verification:
```bash
grep -nE 'plan.{0,3}design.{0,3}run' .claude/rules/moai/workflow/spec-workflow.md
grep -c 'UI-surfaced\|UI surface' .claude/rules/moai/workflow/spec-workflow.md   # ≥ 1
```

### AC-AA2-007 — designer role_profile + pencil MCP absorbed-by annotation (SHOULD-PASS)

**Given** M2 has landed
**When** the verifier runs `grep -c 'Absorbed by manager-design' .moai/config/sections/workflow.yaml .claude/rules/moai/workflow/team-protocol.md .claude/rules/moai/core/settings-management.md`
**Then** the count is ≥ 1 across the three surfaces.

Verification:
```bash
grep -n 'Absorbed by manager-design' .moai/config/sections/workflow.yaml .claude/rules/moai/workflow/team-protocol.md .claude/rules/moai/core/settings-management.md
```

### AC-AA2-008 — RouteModelFor 3-arg signature (MUST-PASS)

**Given** M3 has landed
**When** the verifier runs `grep -n 'func (c \*Config) RouteModelFor' internal/config/model_routing.go`
**Then** the signature is `RouteModelFor(specTier, phase, perfTier string)` (3-arg) AND the file defines `validRoutingPerfTiers = map[string]bool{"max": true, "medium": true, "low": true}` AND zero external call sites remain unchanged in behavior (fallback semantics preserved).

Verification:
```bash
grep -n 'func (c \*Config) RouteModelFor' internal/config/model_routing.go
grep -c 'validRoutingPerfTiers' internal/config/model_routing.go   # ≥ 1
go test ./internal/config/... -run TestRouteModelFor   # 3-arg test PASSes
```

### AC-AA2-009 — model_routing_profiles 3 matrices (MUST-PASS)

**Given** M3 has landed
**When** the verifier runs `grep -cE '^(    )?(max|medium|low):$' .moai/config/sections/workflow.yaml`
**Then** the workflow config carries `model_routing_profiles:` with 3 sub-tiers (max/medium/low), each with 12 entries (S/M/L × plan/run/sync/mx), AND a golden test `TestRouteModelFor_3x12Matrix` exercises all 36 entries.

Verification:
```bash
grep -nE 'model_routing_profiles:' .moai/config/sections/workflow.yaml
grep -cE '^(max|medium|low):' <(awk '/^model_routing_profiles:/,/^[a-z]/' .moai/config/sections/workflow.yaml)   # 3
go test ./internal/config/... -run TestRouteModelFor_3x12Matrix -v
```

### AC-AA2-010 — moai init --model-policy flag (MUST-PASS)

**Given** M3 has landed
**When** the verifier runs `moai init --model-policy max` (and medium, low) AND `moai init --model-policy invalid`
**Then** valid values populate `llm.yaml performance_tier` AND the invalid value exits non-zero with a stderr usage error naming the 3-enum.

Verification:
```bash
moai init --model-policy max 2>&1 | tail -5     # success
moai init --model-policy invalid 2>&1; echo "exit=$?"   # non-zero + usage error
grep -E '^(    )?performance_tier:' /tmp/test-init/.moai/config/sections/llm.yaml   # populated
```

### AC-AA2-011 — claude_models low: haiku → sonnet (MUST-PASS)

**Given** M3 has landed
**When** the verifier runs `grep -E '^\s+low:' .moai/config/sections/llm.yaml`
**Then** the value is `sonnet` (NOT `haiku`) AND the haiku key is absent from `claude_models`.

Verification:
```bash
grep -E '^\s+(low|medium|high):' .moai/config/sections/llm.yaml
grep -c 'haiku' .moai/config/sections/llm.yaml   # 0
```

### AC-AA2-012 — HaikuResidualRule lint gate (MUST-PASS, HARD)

**Given** M3 has landed
**When** the verifier runs `grep -rn 'haiku' .claude/agents/moai/ .moai/config/sections/{llm,workflow}.yaml .claude/rules/moai/ internal/config/model_routing.go internal/spec/` (excluding `_test.go` fixtures and `model-policy.md` historical references clearly marked as legacy)
**Then** the count is 0 (HARD success metric per §08 row 1) AND the `HaikuResidualRule` is registered in `internal/spec/lint.go` AND is NOT skip-able via `lint.skip`.

Verification:
```bash
grep -rn 'haiku' .claude/agents/moai/ .moai/config/sections/llm.yaml .moai/config/sections/workflow.yaml internal/config/model_routing.go 2>&1 | grep -v _test | grep -v '#.*legacy'
# Expected: 0 lines (HARD)
grep -c 'HaikuResidualRule' internal/spec/lint.go   # ≥ 1 (registered)
grep -A 5 'HaikuResidualRule' internal/spec/lint.go | grep -i 'skip\|lint.skip'   # NOT skip-able
```

### AC-AA2-013 — workflow_agents + role_profiles haiku→sonnet (MUST-PASS)

**Given** M3 has landed
**When** the verifier runs `grep -E 'model:\s*haiku' .moai/config/sections/workflow.yaml`
**Then** the count is 0 AND `workflow_agents.read-only-extract` is `sonnet/low` AND `role_profiles.researcher` is `sonnet`.

Verification:
```bash
grep -nE 'model:\s*haiku' .moai/config/sections/workflow.yaml   # 0 lines
grep -A 1 'read-only-extract' .moai/config/sections/workflow.yaml
grep -A 3 'researcher' .moai/config/sections/workflow.yaml | head -5
```

### AC-AA2-014 — doctrine refresh (SHOULD-PASS)

**Given** M4 has landed
**When** the verifier runs `grep -c "2-B\|2-C\|3-tier\|max/medium/low" .claude/rules/moai/development/model-policy.md`
**Then** the count is ≥ 3 (doctrine references the v2 tier matrix) AND the § Inherit-by-Default haiku-exception prose is absent AND `agent-authoring.md` references the 10-agent catalog.

Verification:
```bash
grep -cE '2-B|2-C|3-tier|max/medium/low' .claude/rules/moai/development/model-policy.md   # ≥ 3
grep -c 'except manager-docs and manager-git which use model: haiku' .claude/rules/moai/development/model-policy.md   # 0 (haiku-exception removed)
grep -c '10-agent\|10 retained\|super-advisor\|manager-design' .claude/rules/moai/development/agent-authoring.md   # ≥ 1
```

### AC-AA2-015 — template-first boundary (MUST-PASS)

**Given** M1-M4 have landed
**When** the verifier runs `diff` between live and template mirrors for every edited surface
**Then** live and template are byte-identical (or template-first applied with `make build` regenerating the embedded FS).

Verification:
```bash
diff .claude/agents/moai/super-advisor.md internal/template/templates/.claude/agents/moai/super-advisor.md
diff .claude/agents/moai/manager-design.md internal/template/templates/.claude/agents/moai/manager-design.md
diff CLAUDE.md internal/template/templates/CLAUDE.md | head -20
diff .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml | head -20
make build && go build ./...   # exit 0
```

### AC-AA2-016 — haiku-residual-0 HARD success metric closure (MUST-PASS, HARD)

**Given** M3 has landed AND the SPEC is approaching sync-phase close
**When** the verifier runs the AC-AA2-012 grep (the comprehensive haiku-residual scan)
**Then** the count is 0 (co-equal closure gate with AC-AA2-012) AND the sync-auditor verdict records `haiku_residual: 0` as a MUST-PASS dimension.

Verification:
```bash
# Re-run AC-AA2-012 verification at sync-phase close
grep -rn 'haiku' .claude/agents/moai/ .moai/config/sections/ internal/config/model_routing.go 2>&1 | grep -v _test | grep -v '#.*legacy' | wc -l
# Expected: 0
```

### AC-AA2-017 — supersede flips applied (MUST-PASS, plan-phase)

**Given** this SPEC's plan-phase authoring is complete
**When** the verifier runs `grep -E '^status:' .moai/specs/SPEC-ADVISOR-RUNG-001/spec.md .moai/specs/SPEC-MODEL-ROUTING-WIRE-001/spec.md`
**Then** both return `status: superseded` AND both carry `superseded_by: SPEC-AGENT-ARCH-V2-001` AND both have `updated: 2026-07-09` AND both have an inline supersede note in their Epic Context section.

Verification:
```bash
grep -E '^status:' .moai/specs/SPEC-ADVISOR-RUNG-001/spec.md .moai/specs/SPEC-MODEL-ROUTING-WIRE-001/spec.md
grep -E '^superseded_by:' .moai/specs/SPEC-ADVISOR-RUNG-001/spec.md .moai/specs/SPEC-MODEL-ROUTING-WIRE-001/spec.md
grep -c 'SUPERSEDED by SPEC-AGENT-ARCH-V2-001' .moai/specs/SPEC-ADVISOR-RUNG-001/spec.md .moai/specs/SPEC-MODEL-ROUTING-WIRE-001/spec.md   # ≥ 1 each
```

---

## §D.1 Severity classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-AA2-001, 002, 003, 004, 005, 006 | MUST-PASS | M1 + M2 foundational — the new agent files and workflow integration are the v2 architecture's load-bearing additions |
| AC-AA2-007 | SHOULD-PASS | Absorption annotation is doc-layer; pencil MCP remains functional without it |
| AC-AA2-008, 009, 010, 011, 013 | MUST-PASS | M3 Go + config changes — the 3-tier routing must work mechanically |
| AC-AA2-012, AC-AA2-016 | MUST-PASS (HARD) | Haiku-residual-0 is §08 row 1 — a HARD success metric; closure blocker |
| AC-AA2-014 | SHOULD-PASS | M4 doctrine refresh is important but not architecturally load-bearing for M1-M3 |
| AC-AA2-015 | MUST-PASS | Template-First is a CLAUDE.local.md §2 HARD rule |
| AC-AA2-017 | MUST-PASS | Supersede flips are this SPEC's own plan-phase scope |

## §D.2 Traceability (REQ ↔ AC)

| REQ | AC | Milestone |
|-----|-----|-----------|
| REQ-AA2-001 | AC-AA2-001 | M1 |
| REQ-AA2-002 | AC-AA2-002 | M1 |
| REQ-AA2-003 | AC-AA2-003 | M1 |
| REQ-AA2-004 | AC-AA2-004 | M2 |
| REQ-AA2-005 | AC-AA2-005 | M2 |
| REQ-AA2-006 | AC-AA2-006 | M2 |
| REQ-AA2-007 | AC-AA2-007 | M2 |
| REQ-AA2-008 | AC-AA2-008 | M3 |
| REQ-AA2-009 | AC-AA2-009 | M3 |
| REQ-AA2-010 | AC-AA2-010 | M3 |
| REQ-AA2-011 | AC-AA2-011 | M3 |
| REQ-AA2-012 | AC-AA2-012 | M3 |
| REQ-AA2-013 | AC-AA2-013 | M3 |
| REQ-AA2-014 | AC-AA2-014 | M4 |
| REQ-AA2-015 | AC-AA2-015 | Cross-cutting |
| REQ-AA2-016 | AC-AA2-016 | M3 (closure) |
| REQ-AA2-017 | AC-AA2-017 | Plan-phase (this SPEC) |

## §D.3 Indirect verification (cross-cutting invariants)

- **Subagent boundary (C-HRA-008)**: `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/agents/moai/{super-advisor,manager-design}.md | grep -v '^[^:]*:[0-9]*:[ \t]*#'` returns 0 matches.
- **Conventional Commits**: every M1-M4 commit subject follows `feat(SPEC-AGENT-ARCH-V2-001): M{N} ...` or `docs(SPEC-AGENT-ARCH-V2-001): ...` per the Status Transition Ownership Matrix.
- **Cross-platform build**: `GOOS=windows GOARCH=amd64 go build ./...` exits 0 (no syscall layer introduced).

## §D.4 Closure gates (Definition of Done)

The SPEC is `completed` when ALL of the following hold:

1. All MUST-PASS ACs are green (AC-AA2-001..006, 008..013, 015, 016, 017).
2. SHOULD-PASS ACs (AC-AA2-007, 014) are green OR carry an explicit debt note in progress.md §E.4.
3. HaikuResidualRule reports 0 findings (HARD closure gate — AC-AA2-012 + AC-AA2-016).
4. The 3×12 routing matrix golden test PASSes (AC-AA2-009).
5. Template mirrors are byte-identical to live (AC-AA2-015).
6. sync-auditor verdict is PASS (≥0.85) OR PASS-WITH-DEBT with explicit debt notes.
7. The single sync commit carries the `implemented → completed` transition per the Status Transition Ownership Matrix; `sync_commit_sha` is populated in progress.md §E.4.

## §D.5 Forward-looking checks (post-close)

- **super-advisor spawn in the wild**: post-close, at least one real-world consultation should fire (§08 row 3 success metric — "super-advisor spawn (opus 주입 + xhigh 실효) 실측"). Captured opportunistically; not a close gate.
- **manager-design E2E**: post-close, at least one UI-surfaced SPEC exercises D1→D4 end-to-end (§08 row 4). Captured opportunistically; the DesignSync MCP availability gate (Constraint #4) may delay this.
- **git/bash effort-low regression**: post-close, monitor for commit/push/PR defects after M3's effort-low substitution (§08 row 5 — "회귀 0건"). Captured opportunistically.

## §D.6 Edge cases

- **DesignSync MCP unavailable at M2 run-phase**: per Constraint #4 + B11, M2 authors the agent file + workflow skill against the §04 documented contract; D2-D5 live execution is gated on tool registration. The agent file + skill are valid deliverables even without the tool registered.
- **moai init alias collision**: if a user has an existing `--high` script, the alias path emits a stderr warning but does not break (per D4).
- **HaikuResidualRule false-positive on `_test.go` fixtures**: the rule MUST scope its grep to exclude `_test.go` (test fixtures may reference haiku for regression-test purposes) AND exclude clearly-marked historical/legacy prose references in `model-policy.md`.

## §D.7 Quality gate criteria

- **plan-auditor threshold (Tier L)**: ≥ 0.85 aggregate score.
- **sync-auditor dimensions**: Functionality / Security / Craft / Consistency — each ≥ 0.80; harmonic mean ≥ 0.85.
- **coverage**: ≥ 85% per package on `internal/config/...` (3-arg accessor + profiles), `internal/cli/...` (init flag), `internal/spec/...` (HaikuResidualRule).
- **lint**: `moai spec lint` clean; `golangci-lint run --timeout=2m` clean (NEW vs baseline).
