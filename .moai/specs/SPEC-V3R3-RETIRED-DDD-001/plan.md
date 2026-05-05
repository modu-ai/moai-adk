# SPEC-V3R3-RETIRED-DDD-001 Implementation Plan (Phase 1B)

> Implementation plan for manager-ddd retired stub standardization (follow-up to SPEC-V3R3-RETIRED-AGENT-001).
> Companion to `spec.md` v0.3.0, `research.md` v0.3.0, `acceptance.md` v0.3.0.
> Authored against branch `feature/SPEC-V3R3-RETIRED-DDD-001` at `/Users/goos/MoAI/moai-adk-go` (solo mode, no worktree).

## HISTORY

| Version | Date       | Author                        | Description                                                                                                                                                                                                                                                                |
|---------|------------|-------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | MoAI Plan Workflow (Phase 1B) | 최초 작성 — 5-milestone plan: M1 RED test 확장 → M2-M3 GREEN frontmatter+audit → M4 substitution → M5 REFACTOR docs+CHANGELOG. REQ↔AC matrix.                                                                                                                                  |
| 0.2.0   | 2026-05-04 | MoAI Plan Workflow (iter 2)   | Audit iter 1 defects D1/D2/D5 fix: REQ 순서 sequential REQ-RD-001..012 (gaps 005/006/010/015 제거); deliverables 행 라벨↔내용 정합 (D5: 11 agents not 10); single source of truth (research.md §6: Cat A 30 / Cat B 3 / Cat C 2). 각 deliverable의 REQ Coverage 갱신.                |
| 0.3.0   | 2026-05-04 | manager-spec, iter 3 atomic sync | Audit iter 2 defect D-NEW-3 fix: §3 plan.md numeric ambiguity resolved (canonical 33 within taxonomy + 3 ancillary). 5 artifacts version bumped to current.                                                                                                                                                  |

---

## 1. Plan Overview

### 1.1 Goal Restatement

본 plan은 `spec.md` REQ-RD-001..REQ-RD-012 (sequential renumbered iter 2)을 실행 가능한 5-milestone 작업 분해로 변환한다. 핵심 deliverable:

- **manager-ddd.md retired stub rewrite**: 7628 bytes 활성 정의 → ~1500 bytes retired stub (mirror manager-tdd 패턴, research.md §3.3).
- **agent_frontmatter_audit_test.go 확장**: 339 lines → ~457 lines (2 subtests + 1 new top-level function + 1 helper, research.md §5).
- **factory.go @MX:NOTE expansion**: switch-level comment에 SPEC-V3R3-RETIRED-DDD-001 + `case "ddd":` backward compat rationale 추가 (research.md §4.3).
- **Cat A 30 file documentation substitution**: HARD substitution `manager-ddd` → `manager-cycle` (research.md §6.2). Cat B 3 files KEEP-AS-IS + Cat C 2 files UPDATE-WITH-ANNOTATION는 별도 disposition.
- **handle-agent-hook.sh.tmpl @MX:NOTE 추가**: Cat C2 — legacy `ddd-*` references에 backward compat 명시.
- **agent-hooks.md @MX:NOTE 추가**: Cat C1 — manager-ddd legacy table row backward compat 명시.
- **CHANGELOG.md entry**: Unreleased 섹션 dual-language (한국어 + 영문).

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd` (assumed; verify at run time). Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED (M1)**: agent_frontmatter_audit_test.go에 manager-ddd 단언 (subtests + new top-level function) 추가. 모두 실패 상태 확인 (`go test` FAIL with `RETIREMENT_INCOMPLETE_manager-ddd` and `ORPHANED_MANAGER_DDD_REFERENCE` sentinels).
- **GREEN part 1 (M2)**: `manager-ddd.md` retired stub rewrite (mirror manager-tdd 패턴). RED 테스트 중 `RETIREMENT_INCOMPLETE_manager-ddd` 단언 → GREEN.
- **GREEN part 2 (M3)**: Cat A 30 file documentation substitution 카테고리별 분할 commit (rules, agent definitions, output-style, skill files). RED 테스트 중 `ORPHANED_MANAGER_DDD_REFERENCE` 단언 → GREEN.
- **GREEN part 3 (M4)**: factory.go @MX:NOTE expansion + Cat C 2 files (agent-hooks.md + handle-agent-hook.sh.tmpl) @MX:NOTE 추가. backward compat rationale documentation.
- **REFACTOR (M5)**: CHANGELOG entry + final `make build` + full `go test ./...` + `golangci-lint run` + manual smoke test (mo.ai.kr `moai update` simulation in `/tmp/test-project`).

### 1.3 Deliverables

(File counts traceable to research.md §6 ground truth — single source of truth.)

| Deliverable | Path | REQ Coverage | Action |
|---|---|---|---|
| Retired stub rewrite | `internal/template/templates/.claude/agents/moai/manager-ddd.md` | REQ-RD-001, REQ-RD-008 | MODIFY (rewrite, -123 lines / -5800 bytes) |
| Audit test extension | `internal/template/agent_frontmatter_audit_test.go` | REQ-RD-002, REQ-RD-007, REQ-RD-012 | MODIFY (+118 lines: 2 subtests + 1 top-level test + 1 helper; checkFiles slice = **30 Cat A files**) |
| factory.go @MX:NOTE expansion | `internal/hook/agents/factory.go` | REQ-RD-006 | MODIFY (~+5 lines, comment only) |
| Cat C2 hook wrapper @MX:NOTE | `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` | REQ-RD-010 | MODIFY (~+3 lines, bash comment only) |
| Cat A1 — Rule substitutions (3 files) | `rules/moai/development/agent-authoring.md` (REMOVE list entry), `rules/moai/workflow/spec-workflow.md` (substitute), `rules/moai/workflow/worktree-integration.md` (substitute) | REQ-RD-010 | MODIFY (1 REMOVE + 2 substitute) |
| Cat C1 — Rule UPDATE-WITH-ANNOTATION (1 file) | `rules/moai/core/agent-hooks.md` | REQ-RD-010 | MODIFY (preserve manager-ddd substring; ADD @MX:NOTE) |
| Cat A2 — Agent definition substitutions (**11 files**) | `agents/moai/{manager-strategy, manager-quality, manager-spec, expert-backend, expert-frontend, expert-testing, expert-debug, expert-devops, expert-mobile, expert-refactoring, evaluator-active}.md` | REQ-RD-010 | MODIFY (HARD substitute, ~22 occurrences) |
| Cat A3 — Output-style substitution (1 file) | `output-styles/moai/moai.md` (line 127) | REQ-RD-010 | MODIFY (HARD substitute, 1 occurrence; iter 2 newly discovered) |
| Cat A4 — Skill substitutions (15 files) | per research.md §6.2 sub-section A4 | REQ-RD-010 | MODIFY (HARD substitute, ~32 occurrences) |
| CHANGELOG entry | `CHANGELOG.md` | Trackable | APPEND (Unreleased section dual-language) |

**Disposition totals (research.md §6.5)**: Cat A 30 (= 3 + 11 + 1 + 15) + Cat B 3 (managed via M2 manager-ddd.md rewrite preserving frontmatter `name:`) + Cat C 2 (= 1 in agent-hooks.md UPDATE-WITH-ANNOTATION + 1 in handle-agent-hook.sh.tmpl @MX:NOTE addition).

[HARD] Embedded-template parity: `.claude/...` 변경은 모두 `internal/template/templates/.claude/...` mirror + `make build` 필수 (CLAUDE.local.md §2 Template-First HARD).

### 1.4 Traceability Matrix (REQ → AC mapping)

Plan-auditor PASS criterion #2: every REQ maps to at least one AC. REQ numbering sequential REQ-RD-001..REQ-RD-012 with no gaps (iter 2 fix).

| REQ ID | Category | Mapped AC(s) |
|---|---|---|
| REQ-RD-001 | Ubiquitous | AC-RD-01 |
| REQ-RD-002 | Ubiquitous | AC-RD-01, AC-RD-04 |
| REQ-RD-003 | Ubiquitous | AC-RD-01 |
| REQ-RD-004 | Ubiquitous | AC-RD-02 |
| REQ-RD-005 | Event-Driven | AC-RD-02 |
| REQ-RD-006 | Event-Driven | AC-RD-03 |
| REQ-RD-007 | Event-Driven | AC-RD-04 |
| REQ-RD-008 | State-Driven | AC-RD-01 |
| REQ-RD-009 | State-Driven | AC-RD-02 |
| REQ-RD-010 | State-Driven | AC-RD-04 |
| REQ-RD-011 | Optional | (deferred; no blocking AC) |
| REQ-RD-012 | Unwanted (Composite) | AC-RD-01, AC-RD-04 |

Coverage: **11/12 REQs mapped to 4 ACs** (REQ-RD-011 explicitly deferred per spec.md §5.4). 100% mapping for non-deferred REQs.

---

## 2. Milestone Breakdown (M1-M5)

각 milestone은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test Scaffolding (RED phase) — Priority P0

Reference: `internal/template/agent_frontmatter_audit_test.go` lines 139–161 (predecessor `manager-tdd must be retired` subtest pattern), lines 235–298 (`TestNoOrphanedManagerTDDReference` pattern).

Owner role: `expert-backend` (Go test) or direct `manager-cycle` execution (with `cycle_type=tdd`).

Scope:

1. **Extend `TestAgentFrontmatterAudit`** (REQ-RD-002):
   - Insert `"manager-ddd must be retired"` subtest immediately after existing `"manager-tdd must be retired"` (line 161).
   - Subtest skeleton: read `.claude/agents/moai/manager-ddd.md`, parse frontmatter, assert `rf.retired == true`, emit `RETIREMENT_INCOMPLETE_manager-ddd` sentinel on failure.
   - **Expected RED state**: manager-ddd.md still has full active frontmatter (no `retired: true` field), so subtest FAILS with `RETIREMENT_INCOMPLETE_manager-ddd: ... 'retired: true' 없음. SPEC-V3R3-RETIRED-DDD-001 M2에서 retired stub으로 교체 필요`.

2. **Extend `TestRetirementCompletenessAssertion`** (REQ-RD-012):
   - Insert `"manager-ddd replacement manager-cycle must exist"` subtest immediately after existing manager-tdd subtest (line 188).
   - Subtest skeleton: assert `fs.Stat(fsys, ".claude/agents/moai/manager-cycle.md")` succeeds.
   - **Expected RED state**: manager-cycle.md already exists (predecessor M2), so subtest PASSES at insertion. This is defensive symmetry, not a behavioral RED trigger. (M1 acceptance: subtest exists and passes; verify it does not regress.)

3. **Add new top-level `TestNoOrphanedManagerDDDReference`** (REQ-RD-007, REQ-RD-010):
   - Sibling of `TestNoOrphanedManagerTDDReference`. Same structure but:
     - Helper renamed: `findManagerDDDReferences(content string) []string`
     - checkFiles 슬라이스: **30 Cat A files** per research.md §6.2 (3 rules + 11 agents + 1 output-style + 15 skills). Cat B (3) and Cat C (2) files are EXCLUDED from this slice.
     - Sentinel: `ORPHANED_MANAGER_DDD_REFERENCE`
   - Allow-list exclusions in helper (defensive; not expected to trigger for Cat A files):
     - `# deprecated` 마크다운 헤더
     - `<!--` HTML 코멘트
   - **Expected RED state**: 30 Cat A 파일 모두에 `manager-ddd` 참조가 존재하므로 모든 subtest FAIL.

4. **Run RED verification**:
   - `go test ./internal/template/ -run "TestAgentFrontmatterAudit|TestRetirementCompletenessAssertion|TestNoOrphanedManagerDDDReference" -v`
   - Expected output:
     - `TestAgentFrontmatterAudit/manager-ddd must be retired` FAIL with sentinel
     - `TestRetirementCompletenessAssertion/manager-ddd replacement manager-cycle must exist` PASS
     - `TestNoOrphanedManagerDDDReference/<each-of-30-Cat-A-paths>` FAIL with sentinel
   - All other existing tests (manager-tdd assertions etc.) continue PASS — no regression.

Exit criteria for M1:
- ≥31 new test failures with `RETIREMENT_INCOMPLETE_manager-ddd` (1) or `ORPHANED_MANAGER_DDD_REFERENCE` (30 Cat A subtests) sentinels
- 0 regressions in pre-existing tests
- `go build ./internal/template/...` succeeds
- `golangci-lint run ./internal/template/...` clean
- helper `findManagerDDDReferences` lint-clean

### M2: manager-ddd Retired Stub Rewrite (GREEN part 1) — Priority P0

Reference: `internal/template/templates/.claude/agents/moai/manager-tdd.md` (post-PR #776, 1392 bytes, 39 lines) — exact structural mirror.

Owner role: `expert-backend` (template authoring) or direct `manager-cycle` execution.

Scope:

1. **Rewrite `internal/template/templates/.claude/agents/moai/manager-ddd.md`** (REQ-RD-001, REQ-RD-008):
   - Replace existing 7628 bytes (163 lines) with retired stub mirroring manager-tdd.md structure.
   - Frontmatter (10–12 lines):
     - `name: manager-ddd`
     - `description:` (multi-line, 3 lines): "Retired (SPEC-V3R3-RETIRED-DDD-001) — use manager-cycle with cycle_type=ddd. ... See manager-cycle.md for the active replacement."
     - `retired: true`
     - `retired_replacement: manager-cycle`
     - `retired_param_hint: "cycle_type=ddd"`
     - `tools: []`
     - `skills: []`
   - Body (5 H2 sections):
     - `# manager-ddd — Retired Agent` (H1)
     - `## Replacement` (cite manager-cycle with cycle_type=ddd)
     - `## Migration Guide` (table with 2 rows: legacy DDD invocation → manager-cycle invocation)
     - `## Why This Change` (cite both SPEC-V3R3-RETIRED-AGENT-001 and SPEC-V3R3-RETIRED-DDD-001; consolidation rationale)
     - `## Active Agent` (pointer `.claude/agents/moai/manager-cycle.md`)
   - Target file size: ~1500 bytes (vs predecessor manager-tdd 1392 bytes; slight delta acceptable).

2. **Run M1 RED test for manager-ddd**:
   - `go test ./internal/template/ -run "TestAgentFrontmatterAudit/manager-ddd_must_be_retired" -v`
   - Expected: PASS (retired:true now present, all 5 fields validated)
   - Other M1 tests (TestNoOrphanedManagerDDDReference) still FAIL — those are addressed in M3.

3. **Verify embedded FS regeneration**:
   - `make build` from `/Users/goos/MoAI/moai-adk-go`
   - `git diff internal/template/embedded.go` shows non-empty changes
   - `go test ./internal/template/ -run "TestRetirementCompletenessAssertion" -v` continues to PASS

Exit criteria for M2:
- `TestAgentFrontmatterAudit/manager-ddd_must_be_retired` PASSES
- Generic `TestAgentFrontmatterAudit` walk loop PASSES for manager-ddd entry (5 fields validated)
- Generic `TestRetirementCompletenessAssertion` loop PASSES for manager-ddd → manager-cycle pair
- `make build` succeeds; embedded FS regenerated
- `go vet ./internal/template/...` clean
- `golangci-lint run ./internal/template/...` clean
- manual diff: manager-ddd.md structure mirrors manager-tdd.md exactly (same 5 H2 sections, same frontmatter shape modulo `cycle_type=ddd` substring)

### M3: 30-file Cat A Documentation Substitution (GREEN part 2) — Priority P0

Reference: research.md §6.2 exhaustive Cat A grep results (30 files, ~58 occurrences). Cat B (3) and Cat C (2) are handled separately (Cat B by M2 manager-ddd.md rewrite; Cat C by M4 @MX:NOTE addition).

Owner role: `manager-docs` for documentation substitutions; verification via grep.

Scope (categorized commits to reduce review burden):

**Subset A1 — Cat A1 Rule files (3 files; substitution scope)**:
1. `rules/moai/development/agent-authoring.md` line 104: REMOVE list entry `- manager-ddd: DDD implementation cycle` (manager-cycle already listed; manager-ddd retired).
2. `rules/moai/workflow/spec-workflow.md` line 219: HARD substitute `manager-ddd/tdd (sequential)` → `manager-cycle (sequential, cycle_type per quality.yaml development_mode)`.
3. `rules/moai/workflow/worktree-integration.md` line 135: HARD substitute `expert-backend, expert-frontend, manager-ddd, manager-tdd` → `expert-backend, expert-frontend, manager-cycle`.

**(`rules/moai/core/agent-hooks.md` is Cat C1 — UPDATE-WITH-ANNOTATION; handled in M4 not M3.)**

**Subset A2 — Cat A2 Agent definition files (11 files, ~22 occurrences)** (D5 fix: row label and content count match):
HARD substitution `manager-ddd` → `manager-cycle` (or `manager-cycle with cycle_type=ddd` where appropriate context):
1. `manager-strategy.md` line 136
2. `manager-quality.md` lines 64, 107, 113
3. `manager-spec.md` line 58
4. `expert-backend.md` lines 62, 119
5. `expert-frontend.md` line 119
6. `expert-testing.md` lines 55, 59, 98
7. `expert-debug.md` lines 59, 90
8. `expert-devops.md` lines 59, 114
9. `expert-mobile.md` line 105
10. `expert-refactoring.md` line 54
11. `evaluator-active.md` line 94 (`manager-ddd/tdd` → `manager-cycle`)

Cat B KEEP-AS-IS (do not modify; research.md §6.3):
- B1: `manager-ddd.md` line 2 frontmatter `name:` (self-reference; M2 rewrites body but preserves `name:` field)
- B2: `manager-cycle.md` lines 61, 65, 70 (intended migration table entries naming retired agents)
- B3: `manager-tdd.md` body line 31 (consolidation cross-reference)

**Subset A3 — Cat A3 Output-style file (1 file, 1 occurrence; iter 2 newly discovered)**:
- `output-styles/moai/moai.md` line 127 table cell: HARD substitute `manager-ddd / manager-tdd` → `manager-cycle`

**Subset A4 — Cat A4 Skill files (15 files, ~32 occurrences)**:
HARD substitution `manager-ddd` → `manager-cycle` per research.md §6.2 sub-section A4. Notable structural substitutions:
- `skills/moai-workflow-ddd/SKILL.md` line 20: `agent: "manager-ddd"` → `agent: "manager-cycle"` (skill metadata field)
- `skills/moai-workflow-testing/SKILL.md` line 20: same pattern
- `skills/moai/workflows/run.md` (7 occurrences in workflow rules): each requires careful substitution to preserve workflow semantics

**Verification after each subset**:
- `grep -c "manager-ddd" <files>` to confirm substitution count matches expected
- Re-run `TestNoOrphanedManagerDDDReference` to track progressive PASS

4. **Run M1 RED tests**:
   - After Subset A1 complete (3 rule files): `go test ./internal/template/ -run "TestNoOrphanedManagerDDDReference" -v` shows 3 PASS / 27 FAIL.
   - After Subset A2 complete (11 agents): 14 PASS / 16 FAIL.
   - After Subset A3 complete (1 output-style): 15 PASS / 15 FAIL.
   - After Subset A4 complete (15 skills): all 30 Cat A checkFiles paths PASS.

Exit criteria for M3:
- All 30 Cat A files updated per research.md §6.2 (3 rules + 11 agents + 1 output-style + 15 skills)
- Cat B (3 files) preserved unchanged (verified via grep): manager-ddd.md L2 frontmatter `name:`, manager-cycle.md L61/65/70, manager-tdd.md L31
- Cat C (2 files) NOT yet modified — handled in M4
- `TestNoOrphanedManagerDDDReference` ALL 30 Cat A subtests PASS
- Final grep `grep -rln "manager-ddd" internal/template/templates/` returns 4 files (3 Cat B + 1 Cat C1 agent-hooks.md)
- `make build` succeeds
- `go test ./internal/template/...` full template package PASSES

### M4: factory.go @MX:NOTE Expansion + Cat C UPDATE-WITH-ANNOTATION (GREEN part 3) — Priority P1

Reference: research.md §4.3 + §6.4 (Cat C taxonomy) + factory.go switch-level comment lines 19–28.

Owner role: `expert-backend` (Go) or direct `manager-cycle` execution.

Scope:

1. **Expand `internal/hook/agents/factory.go` switch-level @MX:NOTE** (REQ-RD-006):
   - Current comment block (lines 19–28) cites SPEC-V3R3-RETIRED-AGENT-001 + manager-tdd preservation rationale.
   - Add 1-2 lines extending the citation to SPEC-V3R3-RETIRED-DDD-001 + `case "ddd":` preservation rationale.
   - Skeleton (preserved + appended):
     ```
     // SPEC-V3R3-RETIRED-AGENT-001 + SPEC-V3R3-RETIRED-DDD-001: cycle handler dispatches
     // manager-cycle's unified DDD/TDD workflow hooks. manager-tdd + manager-ddd retired
     // stubs use no hooks (frontmatter cleared) but `case "tdd":` and `case "ddd":` are
     // preserved for backward compatibility with legacy user projects that have not run
     // `moai update`.
     ```
   - **No code change** (no switch case added/removed/modified).

2. **Cat C1 — `rules/moai/core/agent-hooks.md` UPDATE-WITH-ANNOTATION** (REQ-RD-010):
   - Current file lines 48 (table row `manager-ddd | ddd-pre-transformation | ddd-post-transformation | ddd-completion`) and 79 (stdin JSON example) PRESERVE the `manager-ddd` substring (legacy backward compat doc).
   - Add markdown @MX:NOTE comment after the Agent Hook Actions table:
     ```
     <!-- @MX:NOTE: manager-ddd retired (SPEC-V3R3-RETIRED-DDD-001), action set
          preserved here for backward compat with pre-update user projects.
          Active manager-cycle uses cycle-* actions. See manager-cycle.md. -->
     ```
   - Verify post-edit: `grep -c "manager-ddd" rules/moai/core/agent-hooks.md` returns 2 (unchanged) + the file contains the new @MX:NOTE comment.

3. **Cat C2 — `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` @MX:NOTE** (REQ-RD-010):
   - Current file (51 lines) lines 5, 9 reference `ddd-pre-transformation` example. File contains 0 occurrences of `manager-ddd` substring (no substring substitution needed).
   - Add bash comment after lines 5/9:
     ```bash
     # @MX:NOTE: manager-ddd is retired per SPEC-V3R3-RETIRED-DDD-001.
     # ddd-* action references are preserved for backward compat with pre-update user projects.
     # Active manager-cycle uses cycle-* actions (cycle-pre-implementation, etc.).
     ```
   - Insert after line 9 to preserve original example structure.

4. **Verify factory_test.go regression**:
   - `go test ./internal/hook/agents/ -run "TestFactory" -v` — existing manager-ddd factory dispatch tests (lines 200, 229, 231) MUST continue to PASS (case "ddd" preserved).
   - `go vet ./internal/hook/...` clean.
   - `golangci-lint run ./internal/hook/...` clean.

5. **Run full template tests**:
   - `make build` regenerates embedded FS.
   - `go test ./internal/template/ ./internal/hook/ ./internal/cli/...` — ALL pre-existing + new tests PASS.
   - `TestNoOrphanedManagerDDDReference` continues PASS (Cat C files NOT in checkFiles slice; substring preservation in agent-hooks.md does NOT trigger orphan detection).

Exit criteria for M4:
- `factory.go` @MX:NOTE expanded (lines 19–28 area)
- Cat C1 `agent-hooks.md` @MX:NOTE inserted; manager-ddd substring preserved (count: 2)
- Cat C2 `handle-agent-hook.sh.tmpl` @MX:NOTE inserted; ddd-* action references preserved
- `factory_test.go TestFactory*` continues PASSES (no DDD handler regression)
- All M1 tests PASS (audit + completeness + orphan reference 30 Cat A subtests)
- `go vet ./internal/hook/...` clean
- `make build` succeeds
- shell syntax of handle-agent-hook.sh.tmpl validated (`bash -n` after template render)

### M5: CHANGELOG + Final Validation (REFACTOR phase) — Priority P2

Reference: CLAUDE.local.md §18.9 release-drafter integration; predecessor PR #776 CHANGELOG entry pattern.

Owner role: `manager-docs` for CHANGELOG entry; `expert-backend` for final test validation.

Scope:

1. **CHANGELOG entry** (Trackable):
   - APPEND to `CHANGELOG.md` Unreleased section.
   - Dual-language structure (per CLAUDE.local.md §18.0.1 release담당자 절차):
     - `### Bug Fixes / Improvements (Follow-up)` (English) section per spec.md §10 BC Migration.
     - Korean section if predecessor pattern uses one.
   - Cite SPEC-V3R3-RETIRED-DDD-001 + reference predecessor SPEC-V3R3-RETIRED-AGENT-001.
   - Include explicit "User action: run `moai update` to sync" directive.

2. **Final validation**:
   - `make build` — embedded FS regenerated.
   - `go test ./...` — full repo test suite PASS (predecessor 17 file changes + this SPEC's changes).
   - `go vet ./...` — clean.
   - `golangci-lint run ./...` — clean.
   - `make install` — local binary updated.
   - **Manual smoke test**:
     - Spawn a fresh test project: `mkdir -p /tmp/test-ddd-retire && cd /tmp/test-ddd-retire && /path/to/moai init --quick` (or similar setup).
     - `cat .claude/agents/moai/manager-ddd.md` — verify retired stub frontmatter.
     - Simulate Claude Code invoking manager-ddd → SubagentStart hook fires → agentStartHandler returns block decision (verifiable via hook log).
     - Cleanup: `rm -rf /tmp/test-ddd-retire`.

3. **Pre-merge checklist**:
   - [ ] All M1 tests PASS (TestAgentFrontmatterAudit + TestRetirementCompletenessAssertion + TestNoOrphanedManagerDDDReference)
   - [ ] No regressions in predecessor manager-tdd tests
   - [ ] `make build` clean
   - [ ] `go test -race ./...` clean (concurrency safety)
   - [ ] CHANGELOG.md updated with dual-language entry
   - [ ] manual smoke test successful
   - [ ] Branch ready for PR creation

Exit criteria for M5:
- CHANGELOG.md Unreleased section has SPEC-V3R3-RETIRED-DDD-001 entry with `moai update` user action directive
- Full test suite (`go test ./...`) PASS
- `make build && make install` clean
- No regressions detected
- Pre-merge checklist all checked

---

## 3. File-Level Changes Summary

(Authoritative file count from research.md §6.5: Cat A 30 + Cat B 1 modified [B1 manager-ddd.md rewrite] + Cat C 2 = **33 files within taxonomy**, plus 3 ancillary (audit test + factory.go + CHANGELOG) = **36 grand total**. Canonical primary count: **33 (taxonomy-bounded)**. iter 3 manual sync per audit D-NEW-3.)

| File / File Group | Action | LOC delta (estimated) | Phase |
|---|---|---|---|
| `internal/template/templates/.claude/agents/moai/manager-ddd.md` (Cat B1 rewrite) | REWRITE (7628→1500 bytes) | -123 lines | M2 |
| `internal/template/agent_frontmatter_audit_test.go` | MODIFY (extend) | +118 lines | M1 |
| `internal/hook/agents/factory.go` | MODIFY (@MX:NOTE only) | +5 lines (comment) | M4 |
| `internal/template/templates/.claude/rules/moai/core/agent-hooks.md` (Cat C1) | MODIFY (UPDATE-WITH-ANNOTATION; preserve manager-ddd substring) | +3 lines (comment) | M4 |
| `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` (Cat C2) | MODIFY (@MX:NOTE only; no manager-ddd substring) | +3 lines (comment) | M4 |
| Cat A1 — 3 rule files (agent-authoring REMOVE list / spec-workflow / worktree-integration) | MODIFY (1 REMOVE + 2 substitute) | -1 / ±2 lines | M3-A1 |
| Cat A2 — **11 agent definition files** | MODIFY (HARD substitute, ~22 lines edited) | ±22 lines | M3-A2 |
| Cat A3 — 1 output-style file (output-styles/moai/moai.md) | MODIFY (HARD substitute, line 127) | ±1 line | M3-A3 |
| Cat A4 — 15 skill files | MODIFY (HARD substitute, ~32 lines edited) | ±32 lines | M3-A4 |
| `CHANGELOG.md` | APPEND (Unreleased entry) | +20 lines | M5 |

**Total estimated**: -123 + 118 + 5 + 3 + 3 + 1 + 22 + 1 + 32 + 20 = **+82 lines net** (LOC dominated by audit test extension; manager-ddd.md rewrite produces the largest single file shrinkage).

**Files modified by category** (research.md §6.5 ground truth — single canonical count):
- Cat A SUBSTITUTE-TO-CYCLE (research.md §6.2): 30 files (= 3 rules + 11 agents + 1 output-style + 15 skills)
- Cat B B1 manager-ddd.md rewrite (body changed in M2; frontmatter `name:` preserved): 1 file
- Cat C UPDATE-WITH-ANNOTATION (research.md §6.4): 2 files (agent-hooks.md + handle-agent-hook.sh.tmpl)
- Test infrastructure (NOT in any Cat): 1 file (agent_frontmatter_audit_test.go)
- Hook factory (NOT in any Cat): 1 file (factory.go)
- CHANGELOG (NOT in any Cat): 1 file (CHANGELOG.md)

> **Authoritative numeric (canonical)**: **33 files within taxonomy + 3 ancillary = 36 grand total**. Decomposition: 30 Cat A + 1 Cat B B1 + 2 Cat C = 33 (taxonomy); 1 audit test + 1 factory + 1 CHANGELOG = 3 (ancillary). Per research.md §6.5 disposition totals (Cat A 30 / Cat B 3 dispositions / Cat C 2 dispositions = 35 dispositions); only B1 of Cat B is modified by this SPEC, so total modifications = 30+1+2 = 33 within the categorized taxonomy. Use **33 (taxonomy-bounded)** as canonical primary count; cite **36 (grand total)** only when total scope including ancillary is required.

> **Disambiguation rule for cross-artifact references**: When citing "files modified by this SPEC," use **33** (= 30 Cat A + 1 Cat B B1 + 2 Cat C, all dispositions within the taxonomy). When citing "total files touched including test/factory/changelog ancillary," use **36**. Never cite "34" — that number was a transient iter 2 typo and is explicitly retired. The canonical primary count for plan-auditor is **33 (taxonomy-bounded)**.

**Files created**: 0 (all changes are MODIFY).

**Files deleted**: 0 (factory.go `case "ddd"` and `ddd_handler.go` preserved per research.md §4.2 D1).

---

## 4. mx_plan (5 MX tags / 4 files)

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and predecessor SPEC-V3R3-RETIRED-AGENT-001 mx_plan pattern.

| MX type | File | Anchor function/section | Reason |
|---|---|---|---|
| @MX:ANCHOR | `internal/template/agent_frontmatter_audit_test.go` | `TestNoOrphanedManagerDDDReference` (REQ-RD-007) | High fan_in CI gate (every `go test ./internal/template/...` invocation); single point of orphan-reference enforcement |
| @MX:NOTE | `internal/hook/agents/factory.go` | switch-level comment lines 19–28 (REQ-RD-006) | Documents `case "ddd"` + `case "tdd"` backward compat preservation rationale (mirrors predecessor pattern) |
| @MX:NOTE | `internal/template/templates/.claude/agents/moai/manager-ddd.md` | Frontmatter section (REQ-RD-001, REQ-RD-008) | Documents retirement decision + replacement rationale (cite both SPECs in `## Why This Change`) |
| @MX:NOTE | `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` | After lines 5, 9 (legacy `ddd-*` examples; Cat C2) | Documents `ddd-*` action retention for backward compat (REQ-RD-010) |
| @MX:LEGACY | `internal/hook/agents/ddd_handler.go` | Top of file (existing handler, no edit) | (Optional) Add `@MX:LEGACY: SPEC-V3R3-RETIRED-DDD-001 — handler kept for case "ddd" backward compat`. Decision deferred to M4 implementation; if M4 LOC budget exceeds +10, omit this tag and rely on factory.go @MX:NOTE alone. |

Total: 4 required MX tags + 1 optional (5 MX tags / 4 files; 1 ANCHOR + 3 NOTE + 1 optional LEGACY).

---

## 5. Risk Table (file-anchored mitigations)

| 리스크 | 영향 | 확률 | File anchor | 완화 |
|---|---|---|---|---|
| 30 Cat A file substitution accidentally modifies KEEP-allowed reference (false positive substitution) | M | M | Cat B B1 manager-ddd.md L2, B2 manager-cycle.md L61/65/70, B3 manager-tdd.md L31, Cat C C1 agent-hooks.md L48/79 | M3 substitution per Subset (A1/A2/A3/A4); verify each via `grep -c manager-ddd <file>` after edit; manual diff review pre-commit |
| `TestNoOrphanedManagerDDDReference` allow-list misses legitimate exception | M | L | `agent_frontmatter_audit_test.go findManagerDDDReferences` helper | research.md §6.6 allow-list explicit (Cat B/C excluded from checkFiles entirely); M1 verify allow-list with each subtest path; if false positive, add allow-list entry |
| manager-ddd.md retired stub body deviates from manager-tdd pattern (5 H2 sections) | L | M | `agents/moai/manager-ddd.md` body | M2 mirror manager-tdd structure exactly; manual diff review; spec.md §3.3 specifies 5 H2 sections |
| factory.go `case "ddd"` removal pressure (reviewer may suggest cleanup) | L | L | `internal/hook/agents/factory.go` switch | research.md §4.2 Option A 결정 명시; @MX:NOTE rationale documented; defer cleanup to telemetry-driven follow-up SPEC |
| `skills/moai-workflow-ddd/SKILL.md` line 20 `agent:` substitution breaks skill loading | L | M | skill metadata field | M3-C verify skill file YAML parses; skill loader error tracking via `go test ./internal/template/...` after sub |
| `make build` regeneration timing causes stale embedded FS in CI | L | M | `internal/template/embedded.go` | M2 + M3 + M4 each run `make build` checkpoint; CI pipeline runs `make build` before `go test` |
| 30-file Cat A substitution generates large diff (~82 LOC delta + change context) creating PR review burden | M | M | M3 commit set (Subsets A1/A2/A3/A4) | Categorized commits per Subset (A1=rules, A2=agents, A3=output-style, A4=skills); reviewer can review each commit independently with grep verification |
| `agent-hooks.md` UPDATE-WITH-ANNOTATION (Cat C1) strategy creates maintenance debt (legacy `manager-ddd` row remains) | M | M | `rules/moai/core/agent-hooks.md` lines 48, 79 | research.md §6.4 explicit Cat C1 decision; @MX:NOTE makes intent explicit; future cleanup SPEC may remove after telemetry confirmation |
| `handle-agent-hook.sh.tmpl` shell syntax break from added @MX:NOTE comments | L | L | `hooks/moai/handle-agent-hook.sh.tmpl` | M4: `bash -n <rendered>` validation; comments are bash-comment syntax (`#`) so syntactically safe |
| Predecessor `agentStartHandler` does not actually block manager-ddd spawn (despite generic implementation) | L | L | `internal/hook/subagent_start.go` agentStartHandler | research.md §10.3 verified generic implementation (no agent-name special-casing); M5 manual smoke test in `/tmp/test-ddd-retire` confirms |
| mo.ai.kr maintainers miss `moai update` directive after merge | M | M | CHANGELOG.md Unreleased entry | M5 explicit "User action: run `moai update`" directive in dual-language CHANGELOG; predecessor sample (PR #776) provides effective template |
| audit test extension's allow-list rule for Cat B/C edge cases creates loop conflict | L | L | `findManagerDDDReferences` helper allow-list | Cat B (3 files) + Cat C (2 files) are EXCLUDED from checkFiles slice entirely (research.md §6.6); helper allow-list need only cover defensive `# deprecated` and `<!--` patterns |
| Conflict during M3 substitution if predecessor M5 sub left partial state | L | L | All 33 in-taxonomy files (research.md §6 raw grep matched 34 hits with 1 Cat C overlap; post-PR #776 starting state) | research.md §6 grep evidence captured at session start (commit `90b849669`); M3 will verify pre-substitution state matches research.md before substitution |

---

## 6. Solo Mode Path Discipline (4 HARD rules)

Per CLAUDE.local.md §15 + worktree-integration.md HARD rules:

1. [HARD] All write-target paths use project-root-relative form (`internal/template/...`, `internal/hook/...`).
2. [HARD] No `cd /absolute/path` in Bash commands within plan-driven agent invocations.
3. [HARD] Reference files (skills via `${CLAUDE_SKILL_DIR}`) may use absolute paths; write targets cannot.
4. [HARD] Solo mode (no worktree per user directive): all changes happen in `feature/SPEC-V3R3-RETIRED-DDD-001` branch in working tree at `/Users/goos/MoAI/moai-adk-go`.

---

## 7. No Implementation Code in Plan Documents

Per CLAUDE.local.md §16 + spec.md §1.2 (Non-Goals):

- This plan describes WHAT each milestone delivers, WHERE the changes go, and WHY they matter.
- Concrete Go function bodies, test fixture data, full retired stub body text, full @MX:NOTE comment block, etc. are NOT included verbatim — only structural skeletons (e.g., "5 H2 sections per research.md §3.3", "subtest skeleton: read file, parse, assert").
- Implementation details are deferred to `/moai run SPEC-V3R3-RETIRED-DDD-001` execution phase.

Plan-auditor verification: search this plan.md for code blocks containing full Go function bodies — only stubs/skeletons should appear (e.g., `t.Run("manager-ddd must be retired", func(t *testing.T) { ... })` outline, not full implementation).

---

## 8. Plan-Audit-Ready Checklist

All 21 criteria PASS per acceptance.md + spec.md cross-reference (iter 2 expanded with C19-C21 catching audit defects D1/D2/D3/D4):

- C1: Frontmatter v0.2.0 (9 required fields per Step 4 schema) ✅
- C2: HISTORY v0.2.0 entry (iter 2 added) ✅
- C3: 12 EARS REQs across 5 categories (Ubiquitous 4, Event-Driven 3, State-Driven 3, Optional 1, Unwanted 1) ✅
- C4: 4 ACs with 100% non-deferred REQ mapping (11/11 mapped; REQ-RD-011 explicitly deferred per spec.md §5.4) ✅
- C5: BC scope clarity (`breaking: false`, `bc_id: []`) — backward-compatible follow-up ✅
- C6: File:line anchors ≥10 (research.md: 30+, plan.md: 25+) ✅
- C7: Exclusions section present (spec.md §1.2 Non-Goals + §2.2 Out of Scope, ≥4 entries) ✅
- C8: TDD methodology declared ✅
- C9: mx_plan section (5 tags / 4 files; 1 ANCHOR + 3 NOTE + 1 optional LEGACY) ✅
- C10: Risk table with file-anchored mitigations (spec.md §8: 12 risks; plan.md §5: 13 risks) ✅
- C11: Solo mode path discipline (4 HARD rules) ✅
- C12: No implementation code in plan documents ✅
- C13: Acceptance.md G/W/T format with edge cases (4 ACs covered, ≥2 minimum) ✅
- C14: Predecessor pattern reuse documented (research.md §2 + spec.md §1.1) ✅
- C15: Cross-SPEC consistency (depends_on SPEC-V3R3-RETIRED-AGENT-001 with verified merge state, related SPEC-V3R2-ORC-001 + SPEC-V3R3-HYBRID-001) ✅
- C16: BC migration completeness (spec.md §10: no BC, backward-compatible follow-up; CHANGELOG entry skeleton provided) ✅
- C17: Predecessor decision chain documented (research.md §1 + spec.md §1.1) ✅
- C18: External evidence verified (predecessor PR #776 merge confirmed via `git log`, manager-cycle.md presence verified via `ls -la`, 33 file substitution scope (within taxonomy) verified via `grep -rln`; raw grep returned 34 hits with 1 Cat C overlap already accounted) ✅

**Iter 2 additions (audit defect prevention)**:

- C19: REQ numbering sequential REQ-RD-001..REQ-RD-012 with no gaps (audit D1 fix) ✅
- C20: File-count claims consistent across all artifacts (audit D2 fix; single source of truth research.md §6.5) ✅
- C21: AC sentinel messages cite this SPEC's `REQ-RD-NNN` (NOT predecessor's `REQ-RA-NNN`); 3-category disposition taxonomy (Cat A/B/C) used consistently (audit D3/D8 fix) ✅

---

End of plan.md.
