---
id: SPEC-V3R3-RETIRED-DDD-001
version: "0.3.0"
status: completed
created_at: 2026-05-04
updated_at: 2026-05-04
author: Goos Kim
priority: Medium
labels: [retire, agent-runtime, ddd, standardization, follow-up]
issue_number: 778
depends_on: [SPEC-V3R3-RETIRED-AGENT-001]
related_specs: [SPEC-V3R2-ORC-001, SPEC-V3R3-RETIRED-AGENT-001]
breaking: false
bc_id: []
lifecycle: spec-anchored
---

# SPEC-V3R3-RETIRED-DDD-001: Manager-DDD Retired Stub Standardization (Follow-up to RETIRED-AGENT-001)

## HISTORY

| Version | Date       | Author     | Description                                                                                                                                                                                                                                              |
|---------|------------|------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | Goos Kim   | 최초 작성. Predecessor SPEC-V3R3-RETIRED-AGENT-001 §2.2 Out-of-Scope에서 명시된 follow-up. manager-tdd retirement 패턴을 manager-ddd에 동일 적용.                                                                                                                  |
| 0.2.0   | 2026-05-04 | Goos Kim   | Audit iter 1 defects D1-D8 fixed: REQ 순서 sequential REQ-RD-001..012 (gaps 005/006/010/015 제거); file count taxonomy 단일 진실 (research.md §6: Cat A 30 / Cat B 3 / Cat C 2); output-styles/moai/moai.md 추가 발견; AC sentinel REQ-RD-NNN 정정; Self-Audit 3 항목 추가.        |
| 0.3.0   | 2026-05-04 | manager-spec, iter 3 atomic sync | Audit iter 2 defects D-NEW-1 ~ D-NEW-5 fixed via textual substitution. Orphan out-of-range REQ-RD reference in research.md test skeleton replaced with sequential REQ-RD-012. spec.md §3/§4/§7/§8/§10 stale legacy file-count text replaced with canonical Cat A 30 files. plan.md §3 numeric ambiguity resolved (canonical 33 taxonomy + 3 ancillary). spec.md §12.1 self-audit version literal updated to current. Risk row checkFiles count aligned to canonical 30 Cat A files. All 5 artifacts pass grep verification. |

---

## 1. Goal (목적)

본 SPEC은 SPEC-V3R3-RETIRED-AGENT-001 (PR #776, commit `20d77d931` merged)의 명시적 후속 작업을 수행한다. Predecessor §2.2 Out-of-Scope:

> "manager-ddd retired stub의 동등 standardization (동일 패턴 적용 가능하지만 별도 SPEC; 본 SPEC은 manager-tdd case 집중)"

`manager-tdd` retirement 패턴을 `manager-ddd`에 동일 적용하여 manager-cycle을 진정한 통합 DDD/TDD 에이전트로 만든다. predecessor가 확립한 5-field standardized retirement frontmatter (`retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=ddd"`, `tools: []`, `skills: []`)와 sentinel CI assertion 패턴 (`RETIREMENT_INCOMPLETE_<agent>` / `ORPHANED_MANAGER_<AGENT>_REFERENCE`)을 그대로 차용한다.

**핵심 시정 조치**:

1. `internal/template/templates/.claude/agents/moai/manager-ddd.md` (현재 7628 bytes 활성 정의)을 retired stub (~1500 bytes)으로 교체.
2. `internal/template/agent_frontmatter_audit_test.go`에 manager-ddd 단언 추가 (subtest 2개 + 신규 top-level test 함수 1개).
3. 30개 Cat A template 파일의 `manager-ddd` 참조를 `manager-cycle`로 일괄 substitution (research.md §6.2; Cat B 3 + Cat C 2 = 5 파일은 별도 disposition; 총 grep-detected 34 파일).
4. `internal/hook/agents/factory.go` `case "ddd":` backward compat 보존 결정 + @MX:NOTE 확장.
5. mo.ai.kr 사이드 프로젝트는 본 SPEC 머지 후 `moai update` 실행으로 자동 sync (CHANGELOG에 명시).

본 SPEC은 **brownfield delta scope** — 기존 codebase 위에 retirement 패턴 1건 추가. 모든 변경 영역은 Delta marker로 명시된다 (§2 Scope).

### 1.1 배경 — Predecessor Decision Chain

SPEC-V3R2-ORC-001 (Agent Roster Consolidation, completed)이 `manager-ddd` + `manager-tdd` 의 retirement decision을 내렸다. predecessor SPEC-V3R3-RETIRED-AGENT-001은 이 retirement decision의 실행 incompleteness를 manager-tdd 단일 케이스로 시정했다 (mo.ai.kr 5-layer defect chain 차단). 본 SPEC은 동일 시정을 manager-ddd에 적용하여 retirement decision의 **실행 완결성**을 달성한다.

Predecessor 패턴 효과 측정:
- agentStartHandler.Handle 응답 시간: 0.056ms (REQ-RA-012 ≤500ms 목표 대비 8927× 여유)
- plan-auditor 점수: 0.88 (PASS criterion 0.80 초과)
- evaluator iter 2 PASS, 17 files changed, no regressions

본 SPEC은 동일 핸들러 인프라를 재사용하므로 추가 perf 부담 없음 (§5 REQ-RD-012 references predecessor budget).

### 1.2 비목표 (Non-Goals)

- mo.ai.kr 사이드 프로젝트 직접 수정 (`moai update`로 자동 sync; CHANGELOG에 사용자 action 명시).
- agentStartHandler 또는 subagent_start.go 수정 (predecessor 구현이 모든 retired:true 에이전트를 generic하게 커버; research.md §10.3).
- manager-cycle.md 본문 변경 (predecessor에서 이미 manager-ddd + manager-tdd를 deprecated로 명시; research.md §6.1 KEEP 결정).
- 신규 retirement workflow / governance 정의 (별도 SPEC; 본 SPEC은 단일 retirement 적용).
- manager-ddd 부활 또는 cycle 설계 변경 (SPEC-V3R2-ORC-001 retirement decision 유지).
- `text/template` 추가 마이그레이션 (predecessor REQ-RA-006에서 처리; 본 SPEC scope 외).
- worktreePath validation guard 추가/변경 (predecessor `validateWorktreeReturn` + `WORKTREE_PATH_INVALID` sentinel 그대로 사용).
- factory.go `case "ddd":` 제거 (research.md §4.2 D1 결정: KEEP for backward compat).
- `ddd_handler.go` Go 소스 파일 삭제 (case "ddd" 보존과 symmetric).
- `docs-site/` 4-locale 문서 sync (CLAUDE.local.md §17 scope; v3.0 release window 후속).
- `agent-hooks.md` 테이블 row 제거 (legacy backward compat doc로 KEEP; research.md §6.3).
- agency/* (copywriter, designer 등) retired 에이전트 처리 (SPEC-AGENCY-ABSORB-001 영역).
- Claude Code Agent runtime 자체 변경 요청 (외부 Anthropic 의존).

---

## 2. Scope (범위)

### 2.1 In Scope

**File disposition (single source of truth — research.md §6 ground truth)**:

| Category | Files | Action |
|----------|-------|--------|
| **Cat A — SUBSTITUTE-TO-CYCLE** | **30** | `manager-ddd` substring replaced with `manager-cycle`; file modified |
| **Cat B — KEEP-AS-IS** | **3** | Intentional retirement-related references; no edit (with M2 exception for manager-ddd.md whose body is rewritten while frontmatter `name:` is preserved) |
| **Cat C — UPDATE-WITH-ANNOTATION** | **2** | File modified (add @MX:NOTE) but `manager-ddd` substring (where present) preserved as legacy backward-compat doc |
| **Total grep-detected files** | **34** (= 30 Cat A + 3 Cat B + 1 Cat C-with-substring `agent-hooks.md`; Cat C `handle-agent-hook.sh.tmpl` has 0 occurrences but is in modify scope) | — |
| **Total files modified by this SPEC** | **33** (= 30 Cat A + 1 Cat B-rewrite manager-ddd.md + 2 Cat C) | — |

**Owns (MODIFIED files, brownfield delta)**:

- [MODIFY] `internal/template/templates/.claude/agents/moai/manager-ddd.md` — 7628 bytes 활성 정의 → ~1500 bytes retired stub (mirror manager-tdd 패턴; research.md §3.3 5 H2 sections). Cat B special case (frontmatter `name:` preserved).
- [MODIFY] `internal/template/agent_frontmatter_audit_test.go` — 339 lines → ~457 lines (research.md §5.3). 추가 항목:
  - `TestAgentFrontmatterAudit/manager-ddd must be retired` subtest (mirror TDD subtest pattern, lines 139–161).
  - `TestRetirementCompletenessAssertion/manager-ddd replacement manager-cycle must exist` subtest.
  - `TestNoOrphanedManagerDDDReference` 신규 top-level test 함수 (mirror `TestNoOrphanedManagerTDDReference`, **30 file checkFiles** — Cat A files only per research.md §6.2).
  - Helper `findManagerDDDReferences(content string) []string` (mirror existing TDD helper).
- [MODIFY] `internal/hook/agents/factory.go` — `case "ddd":` 보존 + switch-level @MX:NOTE 확장 (manager-ddd retirement backward compat 명시; research.md §4.3).

- [MODIFY Cat A — 30 files SUBSTITUTE-TO-CYCLE] (research.md §6.2 sub-sections A1–A4):
  - **A1 — 3 rule files** (research.md §6.2 A1):
    - `rules/moai/development/agent-authoring.md` (line 104: REMOVE manager-ddd list entry)
    - `rules/moai/workflow/spec-workflow.md` (line 219: substitute)
    - `rules/moai/workflow/worktree-integration.md` (line 135: substitute)
  - **A2 — 11 agent definition files** (research.md §6.2 A2): manager-strategy, manager-quality, manager-spec, expert-backend, expert-frontend, expert-testing, expert-debug, expert-devops, expert-mobile, expert-refactoring, evaluator-active (HARD substitute manager-ddd → manager-cycle)
  - **A3 — 1 output-style file** (research.md §6.2 A3, iter 2 newly discovered): `output-styles/moai/moai.md` (line 127 table cell)
  - **A4 — 15 skill files** (research.md §6.2 A4):
    - `skills/moai/SKILL.md`, `skills/moai/references/mx-tag.md`, `skills/moai/references/reference.md`, `skills/moai/workflows/moai.md`, `skills/moai/workflows/run.md`
    - `skills/moai-foundation-cc/SKILL.md`, `skills/moai-foundation-core/SKILL.md`, `skills/moai-foundation-quality/SKILL.md`, `skills/moai-meta-harness/SKILL.md`
    - `skills/moai-workflow-ddd/SKILL.md` (line 20 `agent: "manager-ddd"` → `agent: "manager-cycle"`)
    - `skills/moai-workflow-loop/SKILL.md`, `skills/moai-workflow-spec/SKILL.md`, `skills/moai-workflow-spec/references/reference.md`, `skills/moai-workflow-spec/references/examples.md`
    - `skills/moai-workflow-testing/SKILL.md` (line 20 `agent: "manager-ddd"` → `agent: "manager-cycle"`)

- [MODIFY Cat C — 2 files UPDATE-WITH-ANNOTATION] (research.md §6.4):
  - C1: `rules/moai/core/agent-hooks.md` (lines 48, 79 manager-ddd substring preserved; ADD @MX:NOTE after Agent Hook Actions table noting backward compat)
  - C2: `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` (no manager-ddd substring; ADD @MX:NOTE bash comment after lines 5, 9 noting `ddd-*` action backward compat)

- [APPEND] `CHANGELOG.md` — Unreleased 섹션에 SPEC-V3R3-RETIRED-DDD-001 항목 추가 (한국어 + 영문 dual section per CLAUDE.local.md §18 enhanced flow).

**Owns (Cat B KEEP-AS-IS — 3 files, no modification per research.md §6.3 decisions)**:

- [EXISTING] B1: `internal/template/templates/.claude/agents/moai/manager-ddd.md` line 2 frontmatter `name: manager-ddd` — M2 rewrites body; frontmatter `name:` field preserved (only this single line constitutes Cat B since rest of file is in Cat A modification scope via body rewrite).
- [EXISTING] B2: `internal/template/templates/.claude/agents/moai/manager-cycle.md` — migration table (lines 61, 65, 70)이 manager-ddd 명시함은 의도된 retirement notice.
- [EXISTING] B3: `internal/template/templates/.claude/agents/moai/manager-tdd.md` — body line 31 (`manager-tdd and manager-ddd ... consolidated`)은 retired stub의 의도된 cross-reference.

**Owns (Predecessor-managed — no modification needed)**:

- [EXISTING] `internal/hook/subagent_start.go` agentStartHandler — predecessor 구현이 generic하게 retired:true 에이전트 모두를 커버 (research.md §10.3).
- [EXISTING] `internal/hook/agents/ddd_handler.go` — factory.go case "ddd" KEEP과 symmetric (research.md §4.2 D1 Option A).
- [EXISTING] `internal/hook/agents/factory_test.go` lines 200, 229, 231 (`NewDDDHandler` 단언) — case 보존과 symmetric.
- [EXISTING] `internal/cli/worktree_validation.go` `validateWorktreeReturn` + `WORKTREE_PATH_INVALID` sentinel — predecessor 구현 그대로 사용.

**Test surface**:

- `internal/template/agent_frontmatter_audit_test.go` (REQ-RD-002, REQ-RD-010, REQ-RD-012): 확장된 retirement audit
- 기존 `TestAgentFrontmatterAudit` 일반 walk loop는 manager-ddd retirement 자동 검증 (research.md §5.1)
- 기존 `TestRetirementCompletenessAssertion` 일반 loop는 manager-cycle 교체 자동 검증
- 신규 `TestNoOrphanedManagerDDDReference`는 **30 Cat A 파일** manager-ddd substring 부재 검증 (research.md §6.2)

**CHANGELOG**: 본 SPEC은 BREAKING CHANGE 없음. `breaking: false`, `bc_id: []`. retired stub은 backward-compatible (legacy 호출자는 retirement message + migration hint를 동일하게 수신).

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 구현 코드 (Go function body, retired stub body 본문 텍스트, sed 스크립트 등) — 본 SPEC은 plan-phase 산출물; 실제 구현은 `/moai run SPEC-V3R3-RETIRED-DDD-001` 단계.
- `internal/hook/subagent_start.go` agentStartHandler 수정 — predecessor 구현이 모든 retired:true 에이전트를 generic하게 커버 (research.md §10.3에서 verified).
- `internal/template/templates/.claude/agents/moai/manager-cycle.md` 본문 변경 — manager-cycle은 이미 통합 에이전트로 존재; deprecated 표기 (lines 61, 65, 70)도 의도된 retirement notice.
- mo.ai.kr 사이드 프로젝트 직접 수정 — `moai update`로 자동 sync; CHANGELOG에 사용자 action 명시 (READ-ONLY 직접 모니터링 / alert도 본 SPEC scope 외).
- `factory.go case "ddd":` 제거 — research.md §4.2 D1 결정: KEEP backward compat (mid-migration 사용자 보호).
- `internal/hook/agents/ddd_handler.go` 또는 `internal/hook/agents/factory_test.go` 수정 — case "ddd" 보존과 symmetric, 변경 불필요.
- 신규 retirement governance / workflow 정의 — 별도 SPEC; 본 SPEC은 단일 케이스 fix.
- agency/* (copywriter, designer 등) retired 에이전트 일괄 처리 — SPEC-AGENCY-ABSORB-001 영역.
- `docs-site/` 4-locale (ko/en/zh/ja) 문서 sync — CLAUDE.local.md §17 영역; v3.0 release window 후속 SPEC.
- Claude Code Agent runtime 또는 frontmatter 검증 강화 요청 — 외부 Anthropic 의존.
- 5-layer defect chain의 layer 5 (`stream_idle_partial`) 직접 fix — predecessor scope, `feedback_large_spec_wave_split.md` lesson #9 영역.
- `text/template` migration scope 확장 — predecessor REQ-RA-006에서 처리됨; 본 SPEC scope 외.
- moai-adk-go의 다른 활성 에이전트 (manager-spec, manager-quality, expert-* 등) retirement — 본 SPEC은 manager-ddd 단일 케이스.
- `agent-hooks.md` 테이블 row 또는 `handle-agent-hook.sh.tmpl` `ddd-*` references 제거 — backward compat doc로 KEEP (research.md §6.3, §4.2 symmetric tdd-* preservation).

---

## 3. Environment (환경)

- 런타임: Go 1.23+ (per `go.mod`), Cobra CLI, Claude Code Agent() runtime v2.1.97+ (worktree CWD isolation), bash hook wrappers.

- 영향 디렉터리:
  - 수정: `internal/template/templates/.claude/agents/moai/manager-ddd.md` (rewrite to retired stub)
  - 수정: `internal/template/agent_frontmatter_audit_test.go` (extend with manager-ddd subtests + new top-level test)
  - 수정: `internal/hook/agents/factory.go` (@MX:NOTE expansion only, no code change)
  - 수정: 30 Cat A files + 2 Cat C files (substitution per research.md §6.5; total 33 files within taxonomy)
  - 수정: `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` (@MX:NOTE only)
  - 수정: `CHANGELOG.md` (Unreleased 섹션 한 entry)
  - Reference (no edit): `internal/hook/subagent_start.go` agentStartHandler (predecessor 구현 그대로 사용), `internal/hook/agents/factory_test.go`, `internal/hook/agents/ddd_handler.go`, `internal/cli/worktree_validation.go`, `internal/template/embedded.go` (auto-regenerated by `make build`).

- 외부 endpoints / 외부 시스템:
  - Claude Code Agent() runtime: SubagentStart hook fire 시 agentStartHandler가 retired:true frontmatter detect → block decision JSON + exit code 2 (predecessor 구현). 본 SPEC은 manager-ddd frontmatter만 추가; runtime 동작 변경 없음.
  - 외부 reference 없음 (인터넷 endpoint 호출 없음).
  - mo.ai.kr 사이드 프로젝트: `moai update` 실행으로 sync; 본 SPEC merge 후 CHANGELOG에 user action 명시.

- 외부 레퍼런스 (codebase grounded scan):
  - `internal/template/templates/.claude/agents/moai/manager-ddd.md:1-39` (current full definition, 7628 bytes — 교체 대상)
  - `internal/template/templates/.claude/agents/moai/manager-tdd.md:1-39` (predecessor retired stub pattern, 1392 bytes — reference)
  - `internal/template/templates/.claude/agents/moai/manager-cycle.md:1-237` (active replacement, 11385 bytes; migration table lines 61–70 already names manager-ddd)
  - `internal/template/agent_frontmatter_audit_test.go:1-339` (predecessor audit test surface — 확장 대상)
  - `internal/hook/agents/factory.go:19-43` (factory dispatch + switch-level comment)
  - `internal/hook/subagent_start.go:152-312` (agentStartHandler generic implementation, no edit)
  - `internal/hook/agents/ddd_handler.go` (existing handler, KEEP)
  - 34 files containing `manager-ddd` (per research.md §6.5 exhaustive grep ground truth: Cat A 30 + Cat B 3 + Cat C-with-substring 1)

---

## 4. Assumptions (가정)

- Predecessor SPEC-V3R3-RETIRED-AGENT-001이 (1) standardized 5-field retirement frontmatter schema, (2) `RETIREMENT_INCOMPLETE_<agent>` / `ORPHANED_*_REFERENCE` sentinel 패턴, (3) `agentStartHandler` generic 구현, (4) `validateWorktreeReturn` worktree validation guard를 이미 머지했음 (PR #776, commit `20d77d931`). 본 SPEC은 이 인프라를 재사용 — 추가 인프라 구축 없음.
- agentStartHandler `Handle()` 메서드는 agent-name special-casing 없이 generic하게 작동한다 (research.md §10.3 verified). manager-ddd frontmatter에 `retired: true` 추가만으로 runtime guard가 작동한다.
- Claude Code Agent runtime은 retirement frontmatter를 graceful 처리한다 (predecessor 사례에서 manager-tdd 검증 완료).
- `internal/template/agent_frontmatter_audit_test.go`의 generic 단언 (`TestAgentFrontmatterAudit` 일반 walk + `TestRetirementCompletenessAssertion` 일반 loop)은 manager-ddd retirement을 자동으로 커버한다. 명시적 subtest 추가는 **defensive symmetry** 목적 (predecessor manager-tdd와 동일 패턴).
- factory.go `case "ddd":` 보존은 mid-migration 사용자에 대한 backward compat 보장 (mo.ai.kr이 정확한 use case). predecessor가 동일 결정으로 `case "tdd":`를 보존했으므로 symmetric pattern 유지.
- `moai-workflow-ddd/SKILL.md` line 20 `agent: "manager-ddd"` 필드 substitution은 skill metadata 변경; 실제 skill body 또는 invocation logic은 변하지 않음.
- 30 Cat A file substitution은 단순 `manager-ddd` → `manager-cycle` 문자열 교체 (HARD substitution case)이며, Cat B (3 files) + Cat C (2 files) 예외 (research.md §6.3/§6.4)는 disposition taxonomy로 명시 관리한다. False positive 가능성은 manual review로 차단.
- TDD methodology 적용 (`.moai/config/sections/quality.yaml`): RED (M1 test 추가) → GREEN (M2-M4 implementation) → REFACTOR (M5 documentation + CHANGELOG).
- mo.ai.kr 사이드 프로젝트는 본 SPEC merge 후 사용자가 명시적으로 `moai update` 실행. CHANGELOG에 이 user action을 명시한다.
- Solo mode, no worktree (사용자 directive). 모든 변경은 `feature/SPEC-V3R3-RETIRED-DDD-001` 단일 브랜치에서 작성.

---

## 5. Requirements (EARS 요구사항)

총 **12개 REQs** — Ubiquitous 4, Event-Driven 3, State-Driven 3, Optional 1, Unwanted 1.

REQ numbering: **sequential `REQ-RD-001` through `REQ-RD-012` with no gaps** (iter 2 fix per audit D1; iter 1 had non-sequential numbering 001-002-003-004-007-008-009-011-012-013-014-016 which violated MP-1 "even one gap or duplicate = FAIL"). Namespace prefix `REQ-RD-NNN` (RD = Retired DDD) distinguishes from predecessor's `REQ-RA-NNN`.

### 5.1 Ubiquitous Requirements (always active)

**REQ-RD-001**
The MoAI template tree **shall** standardize `internal/template/templates/.claude/agents/moai/manager-ddd.md` as a retired-agent stub with the canonical 5-field retirement frontmatter: `retired: true` (boolean), `retired_replacement: manager-cycle` (string), `retired_param_hint: "cycle_type=ddd"` (string), `tools: []` (explicit empty YAML array), `skills: []` (explicit empty YAML array). The stub body **shall** describe (a) the retirement reason citing SPEC-V3R3-RETIRED-AGENT-001 + SPEC-V3R3-RETIRED-DDD-001, (b) the replacement agent (`manager-cycle` with `cycle_type=ddd`), (c) the migration command pattern (Old → New invocation table).

**REQ-RD-002**
The MoAI agent frontmatter audit test (`internal/template/agent_frontmatter_audit_test.go`) **shall** include explicit assertions for manager-ddd retirement. Specifically: (a) `TestAgentFrontmatterAudit` **shall** include a `"manager-ddd must be retired"` subtest emitting `RETIREMENT_INCOMPLETE_manager-ddd` when retirement frontmatter is missing; (b) `TestRetirementCompletenessAssertion` **shall** include a `"manager-ddd replacement manager-cycle must exist"` subtest validating manager-cycle.md presence in embedded FS; (c) a new top-level test function `TestNoOrphanedManagerDDDReference` **shall** validate absence of `manager-ddd` substring across the 30 Cat A substitution-target files (per research.md §6.2) emitting `ORPHANED_MANAGER_DDD_REFERENCE` sentinel.

**REQ-RD-003**
The MoAI embedded FS regeneration (`make build`) **shall** include the standardized manager-ddd retired stub. The existing `TestAgentFrontmatterAudit` walk loop and `TestRetirementCompletenessAssertion` generic loop **shall** validate manager-ddd retirement automatically without requiring code modification beyond the explicit subtests in REQ-RD-002.

**REQ-RD-004**
The MoAI agent runtime SubagentStart guard (`agentStartHandler` in `internal/hook/subagent_start.go`, predecessor SPEC-V3R3-RETIRED-AGENT-001) **shall** block manager-ddd spawn at the runtime layer once frontmatter has `retired: true`, returning `{"decision": "block", "reason": "agent manager-ddd retired (SPEC-V3R3-RETIRED-AGENT-001), use manager-cycle with cycle_type=ddd"}` with exit code 2. **No agentStartHandler modification is required by this SPEC**; the predecessor's generic implementation (research.md §10.3) covers manager-ddd automatically.

### 5.2 Event-Driven Requirements

**REQ-RD-005**
**When** the SubagentStart hook receives an event for `agentName == "manager-ddd"` and the agent file frontmatter has `retired: true`, the hook **shall** emit `{"decision":"block","reason":"agent manager-ddd retired (SPEC-V3R3-RETIRED-AGENT-001), use manager-cycle with cycle_type=ddd"}` to stdout and exit with code 2 (within ≤500ms response time, predecessor REQ-RA-012 budget).

**REQ-RD-006**
**When** `internal/hook/agents/factory.go` `CreateHandler` is invoked with `action == "ddd-<phase>"`, the factory **shall** continue to route to `NewDDDHandler(act)` (preserved for backward compatibility per research.md §4.2 D1). The switch-level `@MX:NOTE` comment **shall** be expanded to cite SPEC-V3R3-RETIRED-DDD-001 alongside SPEC-V3R3-RETIRED-AGENT-001 with explicit rationale that `case "ddd":` and `case "tdd":` are preserved for legacy user projects pre-`moai update`.

**REQ-RD-007**
**When** a developer runs `go test ./internal/template/...`, the agent frontmatter audit test suite **shall** validate manager-ddd retirement compliance and report failures using `RETIREMENT_INCOMPLETE_manager-ddd` (frontmatter incompleteness) or `ORPHANED_MANAGER_DDD_REFERENCE` (residual `manager-ddd` substring in Cat A substitution-target files) sentinel patterns.

### 5.3 State-Driven Requirements

**REQ-RD-008**
**While** `manager-ddd.md` carries `retired: true` in frontmatter, the agent body content **shall** describe (a) retirement reason citing both SPEC-V3R3-RETIRED-AGENT-001 (predecessor) and SPEC-V3R3-RETIRED-DDD-001 (current), (b) replacement agent `manager-cycle` with `cycle_type=ddd`, (c) migration command pattern in a Markdown table, (d) pointer to `.claude/agents/moai/manager-cycle.md`. The body structure **shall** mirror manager-tdd retired stub (5 H2 sections per research.md §3.3).

**REQ-RD-009**
**While** the SubagentStart retired-rejection guard is active (predecessor implementation), manager-ddd spawn block response time **shall** remain ≤500ms (predecessor REQ-RA-012 budget; observed 0.056ms). No new performance burden is introduced by this SPEC.

**REQ-RD-010**
**While** `manager-ddd.md` is in retired-stub state, the documentation references **shall** be disposed across **3 disposition categories** (research.md §6.1):

- **Cat A — SUBSTITUTE-TO-CYCLE (30 files)**: `manager-ddd` substring is replaced with `manager-cycle` (or `manager-cycle with cycle_type=ddd`). Composition (research.md §6.2):
  - 3 rule files (agent-authoring.md REMOVE list entry, spec-workflow.md, worktree-integration.md)
  - 11 agent definition files (manager-strategy, manager-quality, manager-spec, expert-backend, expert-frontend, expert-testing, expert-debug, expert-devops, expert-mobile, expert-refactoring, evaluator-active)
  - 1 output-style file (output-styles/moai/moai.md)
  - 15 skill files (per research.md §6.2 sub-section A4)
- **Cat B — KEEP-AS-IS (3 files)** (research.md §6.3): `manager-ddd.md` line 2 frontmatter `name:` (self-reference; body rewritten in M2 but `name:` field preserved); `manager-cycle.md` lines 61/65/70 migration table; `manager-tdd.md` body line 31 consolidation cross-reference.
- **Cat C — UPDATE-WITH-ANNOTATION (2 files)** (research.md §6.4): `agent-hooks.md` (modified to add @MX:NOTE; `manager-ddd` substring preserved at lines 48, 79); `handle-agent-hook.sh.tmpl` (modified to add bash @MX:NOTE comment; no `manager-ddd` substring but contains `ddd-*` action references).

`TestNoOrphanedManagerDDDReference.checkFiles` slice **shall** contain exactly the 30 Cat A files (Cat B and Cat C are excluded from this slice).

### 5.4 Optional Requirements

**REQ-RD-011**
**Where** the user wants to introspect retired agents, the existing `moai agents list --retired` subcommand (predecessor REQ-RA-014, deferred to follow-up) **may** be extended to surface manager-ddd in the list. **This is deferred** to a separate command-surface SPEC; M5 acceptance does not block on this requirement.

### 5.5 Unwanted Behavior Requirements

**REQ-RD-012 (Unwanted Behavior — Composite, mirrors predecessor REQ-RA-016)**
**If** any one of the following retirement-completion criteria is missing for manager-ddd, **then** CI **shall** fail with `RETIREMENT_INCOMPLETE_manager-ddd` or `ORPHANED_MANAGER_DDD_REFERENCE` sentinel:
- (a) `manager-ddd.md` frontmatter has all 5 standardized fields (`retired: true`, `retired_replacement`, `retired_param_hint`, `tools: []`, `skills: []`).
- (b) Documentation references in the 30 Cat A substitution-target files (REQ-RD-010) name `manager-cycle` instead of `manager-ddd`. Cat B (3) and Cat C (2) files are excluded from this assertion per research.md §6.6 allow-list.
- (c) `agent_frontmatter_audit_test.go` has manager-ddd assertions (REQ-RD-002).
- (d) Active replacement agent `manager-cycle.md` exists in template (predecessor SPEC-V3R3-RETIRED-AGENT-001 verified, no new check needed for this SPEC).

**And** the SubagentStart guard (REQ-RD-005) **shall not** produce silent acceptance (exit 0) for manager-ddd once frontmatter has `retired: true`.

---

## 6. Acceptance Criteria (수용 기준 요약)

상세 Given/When/Then은 `acceptance.md` 참조. 본 SPEC은 4 G/W/T 시나리오로 12 REQs 100% 커버 (REQ-RD-011 제외 deferred). Summary mapping (sequential REQ-RD-001..REQ-RD-012):

- **AC-RD-01** (REQ-RD-001, REQ-RD-002, REQ-RD-003, REQ-RD-008, REQ-RD-012): manager-ddd retirement frontmatter conformance — 5 fields validated by `TestAgentFrontmatterAudit`.
- **AC-RD-02** (REQ-RD-004, REQ-RD-005, REQ-RD-009): SubagentStart runtime guard regression check — manager-ddd spawn blocked at runtime via predecessor's agentStartHandler (≤500ms).
- **AC-RD-03** (REQ-RD-006): Backward compatibility preservation — `factory.go.CreateHandler("ddd-pre-transformation")` routes to DDDHandler (mid-migration users protected); switch-level @MX:NOTE expanded.
- **AC-RD-04** (REQ-RD-002, REQ-RD-007, REQ-RD-010, REQ-RD-012): CI assertion failure semantics — incomplete retirement triggers `RETIREMENT_INCOMPLETE_manager-ddd` and orphaned manager-ddd substring (in Cat A files only) triggers `ORPHANED_MANAGER_DDD_REFERENCE`.

---

## 7. Constraints (제약)

- **TDD methodology** (per `.moai/config/sections/quality.yaml development_mode: tdd`): RED (M1 test 추가) → GREEN (M2-M4 implementation) → REFACTOR (M5 documentation + CHANGELOG).
- **16-language neutrality** (CLAUDE.local.md §15): manager-ddd retired stub body는 language-agnostic 유지 (mirror manager-tdd 패턴).
- **Template-First HARD rule** (CLAUDE.local.md §2): 모든 변경은 `internal/template/templates/` 미러 + `make build` 필수.
- **No drive-by refactor** (CLAUDE.md §7 Rule 2 / Agent Core Behavior #5): 30 Cat A file substitution은 만 `manager-ddd` → `manager-cycle` HARD substitution + 명시된 Cat B/C 예외 (research.md §6.3/§6.4)에 한정. 다른 개선은 별도 SPEC.
- **No flat file** (manager-spec skill HARD rule): `.moai/specs/SPEC-V3R3-RETIRED-DDD-001/` 디렉토리 + 4-5 files (spec.md, plan.md, acceptance.md, research.md, optionally spec-compact.md).
- **No new infra**: predecessor agentStartHandler / validateWorktreeReturn / sentinel pattern 모두 그대로 사용; 본 SPEC은 데이터 추가 (manager-ddd frontmatter)와 substitution만 수행.
- **No mo.ai.kr direct modification**: `moai update` 사용; CHANGELOG에 user action 명시.
- **No docs-site sync in this SPEC**: CLAUDE.local.md §17 4-locale 영역, v3.0 release window 후속.
- **Solo mode, no worktree** (사용자 directive): `feature/SPEC-V3R3-RETIRED-DDD-001` 단일 브랜치에서 plan + run 작성.
- **Bash hook timeout**: agent-hooks.md timeout 5s 유지 (변경 없음).

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 영향 | 확률 | 완화 |
|---|---|---|---|
| 30 Cat A file substitution scope에서 false positive 발생 (manager-ddd가 의도된 reference인 경우) | M | M | research.md §6.3 Cat B + §6.4 Cat C exception list 명시; M3에서 substitution 후 grep 재검증; manual review at commit |
| factory.go `case "ddd":` 제거에 대한 user pressure (cleanup 요구) | L | L | research.md §4.2 Option A 결정 명시; backward compat 우선; 텔레메트리 6개월 관찰 후 별도 cleanup SPEC |
| mo.ai.kr 사이드 프로젝트 사용자가 `moai update` 미실행 시 SubagentStart guard 미적용 | M | M | CHANGELOG entry 명시 ("Run `moai update` to sync"); predecessor 사례에서 동일 가이드라인 효과 입증 |
| manager-ddd retired stub body가 manager-tdd 패턴과 정확히 일치하지 않을 위험 | L | L | research.md §3.3에서 5 H2 sections 명시; M2 implementation 시 mirror commit 후 manual diff 검증 |
| `agent_frontmatter_audit_test.go` 확장 시 기존 manager-tdd 단언 회귀 | L | L | M3 implementation: NEW subtests + NEW top-level function (modify-not-add) — 기존 코드 비편집 유지 |
| `TestNoOrphanedManagerDDDReference` 30 Cat A file checkFiles에서 Cat B/C 예외 미반영 → false positive failure | M | L | research.md §6.6 명시 allow-list 그대로 활용 (Cat B 3 files + Cat C 2 files는 checkFiles slice에서 EXCLUDED); helper `findManagerDDDReferences()` allow-list는 defensive `# deprecated` + `<!--` 패턴만 |
| `skills/moai-workflow-ddd/SKILL.md` line 20 substitution이 skill 작동 영향 | L | M | skill metadata field만 변경 (`agent: "manager-cycle"`); skill body / invocation logic 불변; M5 manual smoke test |
| 30 Cat A file 동시 변경으로 PR review burden 증가 | M | M | M3 단계에서 substitution을 카테고리별 (Cat A1 rules / A2 agents / A3 output-style / A4 skills) 분할 commit; reviewer가 grep -c로 검증 가능 |
| Predecessor 인프라 의존성 검증 부족 → manager-ddd retirement이 무동작 | M | L | research.md §10에서 verifications 명시: predecessor PR #776 머지 확인, agentStartHandler generic 구현 확인, sentinel pattern 재사용 가능 확인 |
| Embedded template `make build` regeneration이 캐시 충돌 | L | M | CLAUDE.local.md "Hard Constraints"에 따라 `make build && make install` 후 재시작 권장; M5 final validation 단계에 명시 |
| `agent-hooks.md` table row + `handle-agent-hook.sh.tmpl` `ddd-*` references KEEP 결정에 대한 retroactive challenge | L | L | research.md §4.2 + §6.3 explicit 결정 + factory.go @MX:NOTE에 backward compat rationale 명시 |
| `internal/hook/agents/ddd_handler.go` Go 소스 코드가 retired stub과 동시에 존재하는 inconsistency | L | L | factory.go `case "ddd":` 보존과 symmetric; mid-migration 사용자 보호; cleanup은 별도 SPEC (텔레메트리 후) |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R3-RETIRED-AGENT-001** (status: completed, PR #776 commit `20d77d931` merged): Standardized retirement frontmatter schema, sentinel CI assertion pattern, agentStartHandler generic implementation, validateWorktreeReturn worktree validation, manager-cycle.md active replacement file. 본 SPEC은 모든 인프라를 재사용; 추가 인프라 구축 없음.

### 9.2 Blocks

- 향후 retirement workflow / governance 표준화 SPEC: 본 SPEC + predecessor가 두 케이스에서 retirement 패턴을 입증하면 일반화된 정책 SPEC을 도출할 수 있다.
- 향후 factory.go `case "ddd"` + `case "tdd"` 제거 cleanup SPEC: 텔레메트리 6개월 관찰 후 zero invocation 확인 시.
- 향후 docs-site 4-locale retired agent 표기 sync SPEC (v3.0 release window).

### 9.3 Related

- **SPEC-V3R2-ORC-001** (Agent Roster Consolidation, completed): manager-ddd + manager-tdd retirement decision의 source. 본 SPEC은 이 decision의 manager-ddd 케이스 실행 완결.
- **SPEC-V3R3-HYBRID-001** (PR #770 merged): provider abstraction + closed-set + audit test 패턴 — 본 SPEC의 retirement 패턴과 유사한 구조 (audit test 검증).
- mo.ai.kr 사이드 프로젝트 (2026-05-04 21:14:54 incident logs): predecessor evidence base. 본 SPEC은 mo.ai.kr가 이미 1000-byte retired stub로 deploy된 상태와 moai-adk-go template의 7628-byte 활성 정의 사이의 inconsistency를 시정.

---

## 10. BC Migration

본 SPEC은 BREAKING CHANGE 없음 (`breaking: false`, `bc_id: []`).

이유: P0 fix는 backward-compatible.

- `manager-ddd.md` frontmatter standardization: 기존 retired stub(없음) → 표준화 retired stub. 기존 활성 manager-ddd 호출자는 retirement message + migration hint를 동일하게 받음. agentStartHandler가 spawn 차단하므로 기존 활성 동작과 동일하게 종료 (단, 차단 시점이 11.4s에서 ≤500ms로 단축됨; predecessor REQ-RA-012 budget 적용).
- factory.go `case "ddd":` 보존: 기존 `ddd-pre-transformation` 등 hook 호출자(레거시 사용자 프로젝트)는 변함없이 DDDHandler routing.
- 30 Cat A file documentation substitution: 사용자에게 visible한 변경은 문서에 `manager-cycle` 명시. 실행 동작 변경 없음. Cat B (3) + Cat C (2)는 별도 disposition.
- mo.ai.kr 사이드 프로젝트: `moai update` 실행만 하면 자동 sync. 별도 user action (frontmatter 직접 편집 등) 불필요.

마이그레이션 절차: 사용자는 `moai update` 실행. CHANGELOG 항목 (proposed):

```markdown
## [Unreleased]

### Bug Fixes / Improvements (Follow-up)

- **SPEC-V3R3-RETIRED-DDD-001**: Standardized manager-ddd as a retired-agent stub (follow-up to SPEC-V3R3-RETIRED-AGENT-001 manager-tdd retirement). manager-cycle is now the canonical unified DDD/TDD implementation agent.
  - Replaced `internal/template/templates/.claude/agents/moai/manager-ddd.md` (7628 bytes active definition → ~1500 bytes retired stub) with the standardized 5-field retirement frontmatter (`retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=ddd"`, `tools: []`, `skills: []`).
  - Extended `internal/template/agent_frontmatter_audit_test.go` with manager-ddd retirement assertions and a new `TestNoOrphanedManagerDDDReference` test enforcing absence of `manager-ddd` substring in 30 Cat A substitution-target template files (per research.md §6.2).
  - Substituted `manager-ddd` references with `manager-cycle` across 30 Cat A template files (3 rules + 11 agent definitions + 1 output-style + 15 skills) per research.md §6.2. Cat B (3 files KEEP-AS-IS) and Cat C (2 files UPDATE-WITH-ANNOTATION) dispositions documented in spec.md §2.1.
  - Preserved `internal/hook/agents/factory.go` `case "ddd":` for backward compatibility with legacy user projects pre-`moai update` (mirrors predecessor `case "tdd":` decision).
  - User action: run `moai update` to sync the standardized retired stub. Existing manager-ddd invocations will be blocked at SubagentStart hook level via predecessor SPEC-V3R3-RETIRED-AGENT-001's agentStartHandler (≤500ms response).
```

---

## 11. Traceability (추적성)

- REQ 총 12개 (sequential REQ-RD-001..REQ-RD-012, no gaps): Ubiquitous 4 (REQ-RD-001..004), Event-Driven 3 (REQ-RD-005..007), State-Driven 3 (REQ-RD-008..010), Optional 1 (REQ-RD-011), Unwanted 1 (REQ-RD-012).
- AC 총 4개 (G/W/T scenarios), 모든 non-deferred REQ에 최소 1개 AC 매핑 (REQ-RD-011 explicit deferred) — 매트릭스는 plan.md §1.4 참조.
- Predecessor mapping (REQ-RA-NNN → REQ-RD-NNN, sequential renumbered iter 2):
  - REQ-RA-002 (frontmatter standardization) → REQ-RD-001 (manager-ddd 케이스)
  - REQ-RA-007 (SubagentStart guard) → REQ-RD-005 (manager-ddd 케이스, 무수정)
  - REQ-RA-013 (documentation references) → REQ-RD-010 (30 Cat A files, manager-ddd 케이스, with Cat B/C taxonomy)
  - REQ-RA-016 (CI assertion) → REQ-RD-012 (manager-ddd sentinel)
- BC 영향: 0건 (`breaking: false`, `bc_id: []`).
- 의존성: SPEC-V3R3-RETIRED-AGENT-001 (completed) 1건; SPEC-V3R2-ORC-001 (completed) related; SPEC-V3R3-HYBRID-001 (PR #770 merged) related.
- 구현 경로 예상 (file count consistent with research.md §6 ground truth):
  - 1 retired stub rewrite (`manager-ddd.md`)
  - 1 audit test extension (3 추가 항목: 2 subtests + 1 new top-level function + 1 helper)
  - 1 factory.go @MX:NOTE expansion (no code change)
  - 30 Cat A documentation substitutions (3 rules + 11 agents + 1 output-style + 15 skills; HARD substitute, with 1 list entry REMOVE in agent-authoring.md)
  - 2 Cat C UPDATE-WITH-ANNOTATION (agent-hooks.md @MX:NOTE; handle-agent-hook.sh.tmpl @MX:NOTE)
  - 1 CHANGELOG entry
  - **Total files modified**: 33 (= 1 retired stub + 1 audit test + 1 factory.go + 30 Cat A + 2 Cat C - manager-ddd.md double-counted = 33; CHANGELOG.md is +1 making it 34)

---

## 12. Self-Audit (Pre-Write Validation Checklist)

### 12.1 Frontmatter Validation

[VERIFIED before Write]:
- ✅ All 9 required fields present in frontmatter: `id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number`
- ✅ `id` matches regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` (SPEC-V3R3-RETIRED-DDD-001)
- ✅ `status` is one of 5 enum values (draft)
- ✅ `priority` is Title-case (Medium)
- ✅ `created_at` / `updated_at` are ISO YYYY-MM-DD (NOT `created` / `updated`)
- ✅ `labels` is YAML array (5 entries: retire, agent-runtime, ddd, standardization, follow-up)
- ✅ `version` is quoted string ("0.3.0") not unquoted float
- ✅ NO legacy aliases present: `created`, `updated`, `spec_id`, `title`-in-frontmatter
- ✅ Optional fields used: `depends_on`, `related_specs`, `breaking`, `bc_id`, `lifecycle`

### 12.2 REQ Numbering Validation (added iter 2 per audit D4)

- ✅ REQ numbers are sequential `REQ-RD-001..REQ-RD-012` with NO gaps (per agent identity rule M5: "even one gap or duplicate = FAIL")
- ✅ All REQ references in this spec.md (including §6 AC summary, §10 BC migration, §11 Traceability) use the new sequential numbers
- ✅ All REQ-RD-NNN cross-artifact references (plan.md §1.4 traceability matrix, acceptance.md §3 coverage table, spec-compact.md REQ list) use the new sequential numbers
- ✅ Predecessor `REQ-RA-NNN` is referenced ONLY in mapping context (§11 Predecessor mapping) — never used for this SPEC's own REQs

### 12.3 Cross-Artifact Consistency Validation (added iter 2 per audit D4)

- ✅ File-count claims match across all 5 artifacts (single source of truth: research.md §6)
  - research.md §6.5 totals table: Cat A 30 / Cat B 3 / Cat C 2 / total grep-detected 34 / total modified 33
  - spec.md §2.1 file disposition table: identical numbers
  - spec.md REQ-RD-010 (state-driven): cites Cat A 30 explicitly
  - plan.md §1.3 deliverables table: 30 Cat A + Cat B/C breakdown
  - acceptance.md AC-RD-04 + Quality Gate: 30 Cat A files
  - spec-compact.md Files-to-Modify table: 30 Cat A + 3 Cat B + 2 Cat C breakdown
- ✅ AC sentinel messages cite this SPEC's `REQ-RD-NNN`, not predecessor's `REQ-RA-NNN` (audit D3 fix)
- ✅ Disposition category taxonomy (Cat A / Cat B / Cat C) used consistently in all artifacts (audit D8 fix)
- ✅ No "approximately X bytes" non-binary-testable acceptance criteria; size verification uses tolerance bands or structural-only checks (audit D7 fix)

---

End of SPEC.
