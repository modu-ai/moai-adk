# SPEC-V3R3-RETIRED-DDD-001 Implementation Plan (Phase 1B)

> Implementation plan for manager-ddd retire stub standardization (predecessor SPEC-V3R3-RETIRED-AGENT-001 pattern replication).
> Companion to `spec.md` v0.3.0 and `research.md` v0.3.0.
> Authored against branch `feature/SPEC-V3R3-RETIRED-DDD-001` (origin/main base after PR #776 merge `20d77d931`).

## HISTORY

| Version | Date       | Author                        | Description                                                                                                       |
|---------|------------|-------------------------------|-------------------------------------------------------------------------------------------------------------------|
| 0.3.0   | 2026-05-04 | MoAI Plan Workflow            | Force-accept after iter 3. §3 Cat A enumeration manual sync — 30 SUBSTITUTE files file-checklist verified via grep. C18 risk row added. |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow            | iter 2 regression fix. Cross-artifact partial sync resolved (D-NEW-1~5).                                          |
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B) | 최초 작성 — M1 RED (audit row) → M2-M4 GREEN (REWRITE + 30 SUBSTITUTE + 2 ANNOTATION) → M5 REFACTOR (CHANGELOG + lessons). |

---

## 1. Plan Overview

### 1.1 Goal restatement

본 plan은 spec.md REQ-RD-001..012를 실행 가능한 **5-milestone** 작업 분해로 변환한다. 핵심 deliverable:

- **Cat B B-01 REWRITE**: `internal/template/templates/.claude/agents/moai/manager-ddd.md` (현재 7674 bytes full agent definition) → ~1.4KB retired stub. predecessor `manager-tdd.md` (1392 bytes) 패턴 verbatim 차용.
- **Cat A SUBSTITUTE 30 files**: `manager-ddd` → `manager-cycle` (또는 `manager-cycle with cycle_type=ddd`) cross-references substitute. 모두 prose-level (action key 식별자는 보존).
- **Cat C UPDATE-WITH-ANNOTATION 2 files**: `agent-hooks.md` Agent Hook Actions 표 row 유지 + retired marker; `agent-authoring.md` Manager Agents (8) listing retired marker.
- **Audit test extension**: `internal/template/agent_frontmatter_audit_test.go`에 manager-ddd row 추가 (predecessor M1 산출 table-driven test).
- **Reference verification (no edit)**: `internal/hook/agents/factory.go` ddd_handler dispatch path 보존 검증.
- **Documentation**: `CHANGELOG.md` `[Unreleased]` 섹션 entry; lessons.md (auto-memory) 검토.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED (M1)**: `agent_frontmatter_audit_test.go` table에 manager-ddd row 추가. 실행 결과 RED 확인 (manager-ddd.md still full agent definition, frontmatter check fails).
- **GREEN part 1 (M2)**: Cat B B-01 — `manager-ddd.md` retire stub REWRITE. audit test → GREEN (5-field standard 충족).
- **GREEN part 2 (M3)**: Cat A 30 SUBSTITUTE files. file-by-file Read + Edit (replace_all 금지 — semantic context preservation). M3 결과로 `grep -rln "manager-ddd" internal/template/templates/`가 Cat B file (retired stub의 description / migration table 내부 self-reference)와 Cat C file (annotated row)로만 hit.
- **GREEN part 3 (M4)**: Cat C 2 UPDATE-WITH-ANNOTATION files. action-key dispatch 보존 + retired marker.
- **REFACTOR (M5)**: `make build` regenerate embedded.go + full `go test ./...` + `golangci-lint run`. CHANGELOG entry 추가. lessons.md 검토 (predecessor와 동일 패턴이므로 신규 lesson 없음).

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|---|---|---|
| Retired stub rewrite | `internal/template/templates/.claude/agents/moai/manager-ddd.md` (REWRITE Cat B B-01) | REQ-RD-001, REQ-RD-008 |
| 30 documentation substitutions | Cat A files §2.1 | REQ-RD-002, REQ-RD-009 |
| 2 annotation updates | Cat C files §2.1 | REQ-RD-003, REQ-RD-007 |
| Audit test extension | `internal/template/agent_frontmatter_audit_test.go` (MODIFY) | REQ-RD-004, REQ-RD-010, REQ-RD-012 |
| factory.go verification | `internal/hook/agents/factory.go` (READ-only verify) | REQ-RD-007 |
| Embedded FS regeneration | `internal/template/embedded.go` (auto-regenerated) | REQ-RD-006 |
| CHANGELOG entry | `CHANGELOG.md` `[Unreleased]` | (Trackable) |

[HARD] Embedded-template parity: 모든 `.claude/...` 변경은 `internal/template/templates/.claude/...` mirror + `make build` 필수 (CLAUDE.local.md §2 Template-First HARD).

### 1.4 Traceability Matrix (REQ → AC mapping)

Plan-auditor must-pass criterion: every REQ maps to at least one AC.

| REQ ID | Category | Mapped AC(s) |
|---|---|---|
| REQ-RD-001 | Ubiquitous | AC-RD-01 |
| REQ-RD-002 | Ubiquitous | AC-RD-01 |
| REQ-RD-003 | Ubiquitous | AC-RD-02 |
| REQ-RD-004 | Ubiquitous | AC-RD-03 |
| REQ-RD-005 | Event-Driven | AC-RD-04 |
| REQ-RD-006 | Event-Driven | AC-RD-01 |
| REQ-RD-007 | Event-Driven | AC-RD-04 |
| REQ-RD-008 | State-Driven | AC-RD-02 |
| REQ-RD-009 | State-Driven | AC-RD-02 |
| REQ-RD-010 | State-Driven | AC-RD-03 |
| REQ-RD-011 | Optional | (deferred per predecessor REQ-RA-014) |
| REQ-RD-012 | Unwanted (Composite) | AC-RD-03, AC-RD-04 |

Coverage: **11/12 REQs mapped to 4/4 ACs** (REQ-RD-011 deferred).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone은 priority-ordered (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD).

### M1: Audit row scaffolding (RED phase) — Priority P0

Reference: predecessor SPEC-V3R3-RETIRED-AGENT-001 M1 산출 `internal/template/agent_frontmatter_audit_test.go` (table-driven schema).

**Tasks**:
- M1.1: Read existing `agent_frontmatter_audit_test.go` (predecessor M1 산출). 확인할 schema: `{name, expected_replacement, expected_param_hint, expected_tools_empty, expected_skills_empty}` 또는 유사한 table struct.
- M1.2: Add manager-ddd row to test table:
  ```go
  {
    name:                   "manager-ddd",
    expected_replacement:   "manager-cycle",
    expected_param_hint:    "cycle_type=ddd",
    expected_tools_empty:   true,
    expected_skills_empty:  true,
  }
  ```
- M1.3: Run `go test ./internal/template/ -run TestAgentFrontmatterAudit -v`. Expected: **RED** (manager-ddd.md still full agent definition, frontmatter check fails — `retired: true` field absent).

**Acceptance**: M1.3 RED confirmed; commit message includes `RED: M1 manager-ddd audit row added`.

### M2: Cat B B-01 REWRITE (GREEN phase part 1) — Priority P0

**Target**: `internal/template/templates/.claude/agents/moai/manager-ddd.md` (7674 bytes → ~1.4KB).

**Pattern source**: `internal/template/templates/.claude/agents/moai/manager-tdd.md` (1392 bytes; predecessor M2 산출).

**Tasks**:
- M2.1: Read predecessor `manager-tdd.md` for template structure (frontmatter format + body sections: Migration Guide table + Why This Change + Active Agent reference).
- M2.2: Read current `manager-ddd.md` to identify content to discard (DDD agent body 7000+ bytes will be replaced entirely; retain only `name: manager-ddd` field as identity).
- M2.3: Write new `manager-ddd.md` with:
  ```yaml
  ---
  name: manager-ddd
  description: |
    Retired (SPEC-V3R3-RETIRED-DDD-001) — use manager-cycle with cycle_type=ddd.
    This agent has been consolidated into the unified manager-cycle agent.
    See manager-cycle.md for the active replacement.
  retired: true
  retired_replacement: manager-cycle
  retired_param_hint: "cycle_type=ddd"
  tools: []
  skills: []
  ---
  ```
  Body: # manager-ddd — Retired Agent + Replacement section + Migration Guide table (DDD-specific old → new patterns) + Why This Change + Active Agent reference.
- M2.4: Run `go test ./internal/template/ -run TestAgentFrontmatterAudit -v`. Expected: **GREEN** (manager-ddd row passes).
- M2.5: `make build` regenerate `internal/template/embedded.go`; verify `git diff internal/template/embedded.go` non-empty.

**Acceptance**: M2.4 GREEN; M2.5 embedded regenerated; commit message includes `GREEN: M2 manager-ddd Cat B B-01 retire stub REWRITE`.

### M3: Cat A 30 SUBSTITUTE (GREEN phase part 2) — Priority P0

**Targets**: 30 files enumerated in spec.md §2.1 Cat A.

**Tasks**:
- M3.1: Run `grep -rln "manager-ddd" internal/template/templates/ | sort` and validate against §2.1 Cat A list (expected 30 files + Cat B target + Cat C 2 files + ANC-01 file count). Adjust list if grep evidence changes.
- M3.2: For each Cat A file, file-by-file:
  - Read file, identify each `manager-ddd` occurrence with surrounding prose context.
  - Decide substitute form per occurrence:
    - General mention: `manager-cycle`
    - Parameter-relevant context: `manager-cycle with cycle_type=ddd`
    - Migration guidance: `manager-cycle (cycle_type=ddd)` form
  - Edit file with targeted replacements (NO `replace_all`).
- M3.3: After all 30 files edited, run `grep -rln "manager-ddd" internal/template/templates/`. Expected hits: ONLY in Cat B file (retired stub self-reference) + Cat C 2 files (annotated rows preserved) + ANC-01 audit test (table row). All Cat A files should have ZERO `manager-ddd` references.
- M3.4: `make build` regenerate embedded.go; verify diff covers 30 files.
- M3.5: Run `go test ./...` to ensure no regression (no test should depend on `manager-ddd` literal in template files).

**Acceptance**: M3.3 grep evidence shows zero hits in Cat A files; M3.5 full test suite GREEN; commit message: `GREEN: M3 manager-ddd Cat A 30 SUBSTITUTE files`.

[HARD] Anti-pattern reminder: NO drive-by edits in Cat A files. Touch only `manager-ddd` references and immediate prose context. CLAUDE.md §7 Rule 2 + Agent Core Behavior #5 (Maintain Scope Discipline).

### M4: Cat C 2 UPDATE-WITH-ANNOTATION (GREEN phase part 3) — Priority P0

**Targets**: `agent-hooks.md` (C-01) + `agent-authoring.md` (C-02) per spec.md §2.1.

**Tasks**:

**C-01: `internal/template/templates/.claude/rules/moai/core/agent-hooks.md`**
- M4.1: Read `agent-hooks.md` § Agent Hook Actions table.
- M4.2: Locate manager-ddd row (PreToolUse/PostToolUse/SubagentStop columns). DO NOT delete the row — `factory.go` `ddd_handler.go` action keys (`ddd-pre-transformation`, `ddd-post-transformation`, `ddd-completion`) are still dispatched for backward compat (REQ-RD-007 invariant).
- M4.3: Add retired marker to row:
  ```markdown
  | manager-ddd (retired — see manager-cycle) | ddd-pre-transformation | ddd-post-transformation | ddd-completion |
  ```
- M4.4: Add explanatory note below table: "Note: manager-ddd action keys are dispatched only for backward compatibility with hook events that bypass the SubagentStart guard. New `manager-ddd` Agent() spawns are blocked by the SubagentStart retired-rejection guard (SPEC-V3R3-RETIRED-AGENT-001 REQ-RA-007)."

**C-02: `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`**
- M4.5: Read `agent-authoring.md` § Manager Agents (8) listing.
- M4.6: Substitute `manager-ddd: DDD implementation cycle` line. Two acceptable forms (pick consistent with predecessor C-02 of `manager-tdd`):
  - Option A (predecessor pattern): keep listing count "8" but replace `manager-ddd: DDD implementation cycle` with `manager-cycle: Unified DDD/TDD implementation cycle (replaces manager-ddd, manager-tdd)`.
  - Option B: keep both manager-tdd and manager-ddd lines but add ` (retired — see manager-cycle)` suffix.
- M4.7: Verify against predecessor M5 산출 of agent-authoring.md to determine which option was chosen for manager-tdd (template version already shows manager-cycle inserted; predecessor used Option A). Use Option A for consistency.

**Common**:
- M4.8: Run `make build` regenerate embedded.go.
- M4.9: Run `go test ./...` full suite GREEN.

**Acceptance**: C-01 row preserved with annotation; C-02 listing updated; M4.9 full test suite GREEN; commit message: `GREEN: M4 manager-ddd Cat C 2 UPDATE-WITH-ANNOTATION`.

### M5: Documentation + final validation (REFACTOR phase) — Priority P1

**Tasks**:
- M5.1: Append CHANGELOG.md `[Unreleased]` entry per spec.md §10.
- M5.2: Run final `make build && go test ./... && golangci-lint run`. All GREEN required.
- M5.3: Verify `git status` — no orphan changes outside §2.1 file taxonomy.
- M5.4: Verify `grep -rln "manager-ddd" internal/template/templates/` — only Cat B + Cat C + ANC-01 hits.
- M5.5: Run `internal/hook/agents/factory.go` reference verification (no edits): confirm `dddHandler` dispatch is intact, `ddd_handler.go` exists, action keys `ddd-pre-transformation`, `ddd-post-transformation`, `ddd-completion` still routed.
- M5.6: lessons.md review — predecessor RETIRED-AGENT lessons #11 already documents 5-layer defect chain; this SPEC follows the same pattern, no new lesson needed unless audit catches new defect class.
- M5.7: PR creation. Title: `release: SPEC-V3R3-RETIRED-DDD-001 manager-ddd retired stub standardization`. Description references issue #778, predecessor PR #776, plan-audit reports iter 1/2/3.

**Acceptance**: M5.2 final CI green; M5.4 grep evidence verified; PR opened with all 36 files (33 within taxonomy + 3 ancillary) listed in description.

---

## 3. File Manifest (Cat A enumeration — manual sync verified, iter 3)

iter 3 plan-audit force-accept manual sync 시점에서 grep -rln evidence로 검증된 30 Cat A files (spec.md §2.1과 정합):

**Agent files (13)**:
1. `internal/template/templates/.claude/agents/moai/manager-cycle.md`
2. `internal/template/templates/.claude/agents/moai/manager-spec.md`
3. `internal/template/templates/.claude/agents/moai/manager-strategy.md`
4. `internal/template/templates/.claude/agents/moai/manager-quality.md`
5. `internal/template/templates/.claude/agents/moai/manager-tdd.md`
6. `internal/template/templates/.claude/agents/moai/expert-backend.md`
7. `internal/template/templates/.claude/agents/moai/expert-frontend.md`
8. `internal/template/templates/.claude/agents/moai/expert-mobile.md`
9. `internal/template/templates/.claude/agents/moai/expert-debug.md`
10. `internal/template/templates/.claude/agents/moai/expert-devops.md`
11. `internal/template/templates/.claude/agents/moai/expert-testing.md`
12. `internal/template/templates/.claude/agents/moai/expert-refactoring.md`
13. `internal/template/templates/.claude/agents/moai/evaluator-active.md`

**Rule files (3)**:
14. `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`
15. `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md`
16. `internal/template/templates/.claude/output-styles/moai/moai.md`

**Skill files (13)**:
17. `internal/template/templates/.claude/skills/moai/SKILL.md`
18. `internal/template/templates/.claude/skills/moai/workflows/run.md`
19. `internal/template/templates/.claude/skills/moai/workflows/moai.md`
20. `internal/template/templates/.claude/skills/moai/references/mx-tag.md`
21. `internal/template/templates/.claude/skills/moai/references/reference.md`
22. `internal/template/templates/.claude/skills/moai-foundation-cc/SKILL.md`
23. `internal/template/templates/.claude/skills/moai-foundation-core/SKILL.md`
24. `internal/template/templates/.claude/skills/moai-foundation-quality/SKILL.md`
25. `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md`
26. `internal/template/templates/.claude/skills/moai-workflow-loop/SKILL.md`
27. `internal/template/templates/.claude/skills/moai-workflow-spec/SKILL.md`
28. `internal/template/templates/.claude/skills/moai-workflow-spec/references/reference.md`
29. `internal/template/templates/.claude/skills/moai-workflow-spec/references/examples.md`

**CLAUDE.md (1)**:
30. `internal/template/templates/CLAUDE.md`

**Cat B (1)**: `internal/template/templates/.claude/agents/moai/manager-ddd.md` (REWRITE)

**Cat C (2)**:
- `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` (annotation)
- `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (annotation)

**Ancillary (3)**:
- `internal/template/agent_frontmatter_audit_test.go` (MODIFY — manager-ddd row)
- `internal/hook/agents/factory.go` (READ-only verify)
- `CHANGELOG.md` (APPEND)

**Total**: 30 + 1 + 2 + 3 = **36 grand total** (consistent with spec.md §2.1).

> **Note (iter 3 manual sync)**: `internal/template/templates/.claude/skills/moai-workflow-ddd/SKILL.md` and `internal/template/templates/.claude/skills/moai-workflow-testing/SKILL.md` from initial grep evidence may also reference `manager-ddd`. M3.1 must re-verify against the live template tree at the implementation time and adjust the Cat A enumeration if hits diverge from this manifest. The 30-count is the iter 3 baseline; tolerance ±2 acceptable as long as audit test (REQ-RD-004) and grep verification (M3.3) pass.

---

## 4. Risks Specific to Plan Execution

(spec.md §8 covers high-level risks; this section covers plan-execution-specific risks.)

| 리스크 | 완화 |
|---|---|
| C18 — Cat A enumeration baseline drift between iter 3 freeze and implementation start | M3.1 re-runs `grep -rln "manager-ddd" internal/template/templates/` and compares against §3 manifest. Discrepancies are reconciled before M3.2 begins. |
| Cat A file에 `manager-ddd-pre-transformation` 등 action key 식별자가 prose 내 언급 시 substitute 잘못 적용 | Edit 도구의 `replace_all` 금지 정책. action key 식별자는 hyphen-suffix 패턴 (`-pre-transformation`, `-completion`)을 함께 노출하므로 grep 단계에서 시각적 식별 가능. |
| `make build` 실패 (embedded.go regeneration) | M2.5 / M3.4 / M4.8에서 즉시 stop + investigate. embed/ 디렉토리 손상이 원인이면 manual `cd internal/template && go generate` fallback. |
| `go test ./...`이 본 SPEC 외 기존 flaky test (e.g., predecessor TestSupervisor_NonZeroExit ETXTBSY race per CLAUDE.local.md §18.11)로 fail | M5.2에서 재실행 + ETXTBSY race는 retry 로직 적용 (CLAUDE.local.md §18.11 mitigation 참조). 본 SPEC scope 외이므로 별도 fix 시도 금지. |

---

## 5. Cross-references

- Predecessor: `.moai/specs/SPEC-V3R3-RETIRED-AGENT-001/plan.md` (M1-M5 패턴 source)
- spec.md (companion)
- acceptance.md (companion)
- research.md (companion, codebase grounded scan)
- CLAUDE.local.md §2 Template-First HARD rule
- CLAUDE.local.md §18 Enhanced GitHub Flow (PR creation, branch protection)

---

End of plan.md
