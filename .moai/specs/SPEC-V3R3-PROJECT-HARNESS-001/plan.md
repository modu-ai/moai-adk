# SPEC-V3R3-PROJECT-HARNESS-001 — Implementation Plan

## 1. Overview

본 plan은 `/moai project` Phase 5+ 확장 (16Q 인터뷰 + 5-Layer 통합 장치)을 5개 Phase로 분할 전달한다. Phase는 priority-ordered (시간 추정 금지, `agent-common-protocol.md` Time Estimation rule 준수). SPEC-V3R3-HARNESS-001 (meta-harness skill) 선행 완료가 본 SPEC의 모든 Phase의 entry condition이다.

## 2. Architectural Approach

### 2.1 Component Map

```
/moai project (existing Phase 1-4)
   └─> Phase 5: Socratic Interview (NEW)               (Phase 1)
        ├─ Round 1: Q1-Q4  (도메인/기술/규모/팀)
        ├─ Round 2: Q5-Q8  (방법론/디자인툴/UI/디자인시스템)
        ├─ Round 3: Q9-Q12 (보안/성능/배포/외부통합)
        └─ Round 4: Q13-Q16(customization/제약/우선순위/최종확인)
             └─> .moai/harness/interview-results.md (Phase 1)

   └─> Phase 6: meta-harness Invocation              (Phase 2)
        └─ Skill("moai-meta-harness") with interview answers
             └─> .claude/agents/my-harness/*.md  (DYNAMIC, user area)
             └─> .claude/skills/my-harness-*/SKILL.md (DYNAMIC, user area)

   └─> Phase 7: 5-Layer Activation                   (Phase 3)
        ├─ L1: skill frontmatter triggers inject
        ├─ L2: workflow.yaml.harness 섹션 갱신
        ├─ L3: CLAUDE.md @import marker inject
        ├─ L4: template static import line verify
        └─ L5: .moai/harness/ user directory scaffold

   └─> Phase 8: Integration & Verification           (Phase 4)
        ├─ moai update 안전성 회귀 테스트
        ├─ moai doctor 5-Layer 진단 추가
        └─ 새 세션 시뮬레이션 검증

   └─> Phase 9: Template-First Mirror & Release      (Phase 5)
        └─ internal/template/templates/.../workflows/{plan,run,sync,design}.md
           Layer 4 static import line 영구 mirror
```

### 2.2 Data Flow

1. 사용자가 `/moai project` 실행 → 기존 Phase 1-4 (project doc generation) 완료.
2. Phase 5 진입: orchestrator가 `AskUserQuestion`을 4번 호출 (Round 1-4 each with up to 4 questions).
3. 각 Round 종료 시 답변을 in-memory 버퍼에 누적; Round 4 Q16 (최종 확인) 응답이 "Confirm"이면 다음 Phase로 진행, "Restart"면 처음부터, "Abort"면 종료 + 산출물 생성 안 함 (REQ-PH-010).
4. Phase 6: orchestrator가 `Skill("moai-meta-harness")` 호출 시 16개 답변을 prompt context로 전달. meta-harness가 `.claude/agents/my-harness/*` + `.claude/skills/my-harness-*/` 생성.
5. Phase 7: orchestrator가 5-Layer 활성화를 순차 수행:
   - L1: 생성된 skill frontmatter `triggers` 섹션 추가 (paths/keywords/agents/phases).
   - L2: `.moai/config/sections/workflow.yaml`의 `harness:` 섹션 갱신 (custom_agents, custom_skills, chaining_rules).
   - L3: `CLAUDE.md` 끝에 `<!-- moai:harness-start id="..." -->` ~ `<!-- moai:harness-end -->` 블록 inject.
   - L4: template 정적 import line 존재 verify (Phase 9에서 mirror).
   - L5: `.moai/harness/` 디렉터리 scaffolding (main.md, plan-extension.md, run-extension.md, sync-extension.md, optional design-extension.md, chaining-rules.yaml, interview-results.md, README.md).
6. Phase 8: 새 세션 시뮬레이션 (CLAUDE.md re-load → @import follow → harness customization context 포함 확인).

### 2.3 Boundary Enforcement

- **FROZEN guard**: Phase 6 meta-harness 호출 시 `path-prefix matcher`가 first check; `.claude/agents/moai/` 또는 `.claude/skills/moai-*/`에 쓰려는 시도는 reject (REQ-PH-011).
- **Abort handling**: Phase 5 Round 1-4 어느 시점에서든 사용자가 "Abort" 선택 시, 누적 in-memory 버퍼만 폐기하고 디스크에 어떤 파일도 작성하지 않는다 (REQ-PH-010).
- **moai update safety**: Phase 8 회귀 테스트는 `moai update` 실행 전후 `.moai/harness/` + `.claude/agents/my-harness/` + `.claude/skills/my-harness-*/`를 `diff -rq`로 비교하여 0 changes를 확증한다 (REQ-PH-009).

## 3. Phased Implementation

### Phase 1 — Socratic Interview (Priority: P0 Critical)

**Goal**: 16Q / 4 Round 인터뷰 구현 + 결과 영구 기록.

**Deliverables**:
- `.claude/skills/moai/workflows/project.md`에 Phase 5 신설 (인터뷰 프로토콜 명세).
- 4개 Round prompt template (각 Round별 4 질문 + 옵션 + 권장 마커 + 상세 설명).
- `.moai/harness/interview-results.md` 작성 로직 (Q1-Q16 + timestamp + project metadata).
- Round 4 Q16 분기 처리 (Confirm / Restart / Abort).

**REQ-IDs covered**: REQ-PH-001, REQ-PH-002, REQ-PH-003, REQ-PH-010.

**Exit gate**: 16Q 인터뷰 시나리오 (iOS 프로젝트 예시) end-to-end 통과; abort 분기 시 디스크에 0 파일 작성 검증.

### Phase 2 — meta-harness Invocation (Priority: P0 Critical)

**Goal**: 인터뷰 답변 기반 dynamic harness 산출물 생성.

**Deliverables**:
- `Skill("moai-meta-harness")` 호출 wrapper (orchestrator side, project.md Phase 6).
- 답변 → meta-harness prompt context 변환 schema (16개 답변 → structured context).
- 산출물 path-prefix 검증 (FROZEN guard).
- 생성 실패 시 cleanup 로직 (partial output 제거).

**REQ-IDs covered**: REQ-PH-004, REQ-PH-011.

**Dependency**: SPEC-V3R3-HARNESS-001 완료 필수.

**Exit gate**: iOS 시나리오에서 `.claude/agents/my-harness/{ios-architect,swiftui-engineer}.md` + `.claude/skills/my-harness-{ios-patterns,swiftui-best-practices}/SKILL.md` 생성 확인. moai-managed area 미수정 검증.

### Phase 3 — 5-Layer Activation (Priority: P0 Critical)

**Goal**: L1-L5 통합 장치 활성화 로직 구현.

**Deliverables**:
- **L1**: 생성된 `my-harness-*` skill의 frontmatter에 `triggers` 섹션 자동 inject (keywords, agents, phases, paths). 주 구현은 meta-harness 책임이지만 본 SPEC은 검증 책임.
- **L2**: `.moai/config/sections/workflow.yaml` 갱신 로직. `harness:` 섹션 (enabled, generated_at, domain, spec_id, custom_agents[], custom_skills[], chaining_rules[]) 추가. YAML merge 시 기존 설정 보존.
- **L3**: `CLAUDE.md` 끝에 marker 블록 inject. id (interview SPEC ID), generated (ISO8601), Domain, Harness level, @import 5줄 (workflow.yaml, custom rules, agents, skills). 동일 marker block 존재 시 idempotent 갱신.
- **L4**: Phase 9에서 template static import line mirror; 본 Phase는 local copy verify만.
- **L5**: `.moai/harness/` 디렉터리 scaffolding.
  - `main.md` (CLAUDE.md @import 진입점, project metadata + 도메인 요약)
  - `plan-extension.md` (manager-spec chain 명시)
  - `run-extension.md` (manager-tdd / manager-ddd chain rules)
  - `sync-extension.md` (manager-docs chain rules)
  - `design-extension.md` (Q13 답변이 "Advanced"일 때만, REQ-PH-012)
  - `chaining-rules.yaml` (machine-readable chain rules)
  - `interview-results.md` (Phase 1 산출, here re-referenced)
  - `README.md` (사용자 가시성 — 5-Layer 설명 + 편집 가능 영역 표시)

**REQ-IDs covered**: REQ-PH-005, REQ-PH-008, REQ-PH-012.

**Exit gate**: 5-Layer 각각 단위 테스트 통과 (AC-PH-02). idempotent re-run (재실행 시 충돌 없음) 검증.

### Phase 4 — Integration & Verification (Priority: P1 High)

**Goal**: 새 세션 자동 활성화 + moai update 안전성 + moai doctor 통합.

**Deliverables**:
- 새 세션 시뮬레이션 테스트 (CLAUDE.md reload → @import follow 검증).
- `moai update` 회귀 테스트 (`internal/cli/update_test.go` 또는 신규 `internal/harness/update_safety_test.go`).
- `moai doctor` 5-Layer 진단 추가 (`internal/cli/doctor.go` 확장).
  - L1: my-harness-* skill에 triggers 섹션 존재 여부.
  - L2: workflow.yaml에 harness 섹션 존재 + 유효성.
  - L3: CLAUDE.md marker block 존재 + paired marker.
  - L4: template import line + local copy 일치 verify.
  - L5: `.moai/harness/` 필수 파일 존재.
- `my-harness-*` prefix 충돌 검사 (`moai-*`와 같은 이름 시 경고).

**REQ-IDs covered**: REQ-PH-006, REQ-PH-007, REQ-PH-009.

**Exit gate**: handoff §5.2 iOS 시나리오 verbatim end-to-end 검증 (manager-tdd가 chain 적용). `moai update` 후 사용자 영역 0 changes 검증.

### Phase 5 — Template Mirror & Release (Priority: P1 High)

**Goal**: Layer 4 template static import line 영구 mirror + 릴리스 준비.

**Deliverables**:
- `internal/template/templates/.claude/skills/moai/workflows/plan.md` 마지막에 정적 1줄 import line:
  ```
  ## Custom Harness Extension (Optional)

  @.moai/harness/plan-extension.md

  *(이 파일은 `/moai project --harness`로 생성됩니다. 파일이 없으면 자동으로 skip됩니다.)*
  ```
- 동일 패턴: `run.md`, `sync.md`, `design.md`. 4개 파일 각각 한 줄씩.
- `make build` → embedded.go regenerate.
- 로컬 `.claude/skills/moai/workflows/*.md`도 동기화.
- v2.17 release notes draft 갱신 (`.moai/release/v2.17.0-draft.md`에 본 SPEC 항목 추가).
- HISTORY 섹션 갱신.

**REQ-IDs covered**: REQ-PH-005 (L4 mirror), 모든 REQ에 대한 release readiness.

**Exit gate**: `moai update` 후 import line 보존 + 사용자 `.moai/harness/*` 보존 동시 검증. Template + local diff 0.

## 4. Data Schema

### 4.1 Interview Round Schema

각 Round는 `AskUserQuestion` 1회 호출, 최대 4 질문. 각 질문은 다음 구조:

```yaml
question_id: Q01..Q16
round: 1..4
text: <conversation_language의 질문 텍스트>
options:
  - label: <option label, recommended first with "(권장)" 마커>
    description: <상세 설명, 함의 / trade-off>
recommended_index: 0  # 항상 첫 번째 옵션
allow_other: true     # AskUserQuestion 자동 fallback
```

### 4.2 interview-results.md Schema

```yaml
---
spec_id: SPEC-PROJ-INIT-<NNN>
generated_at: <ISO8601>
project_root: <abs path>
conversation_language: <ko|en|ja|zh>
---

# Interview Results

## Round 1: Domain & Technology Foundation
- Q01: <question>
  - Answer: <user choice or "Other: ...">
- Q02: ... (similar)
- Q03: ...
- Q04: ...

## Round 2: Methodology & Design
- Q05-Q08

## Round 3: Security, Performance, Deployment
- Q09-Q12

## Round 4: Customization & Final Confirmation
- Q13-Q16 (Q16 답변이 Confirm/Restart/Abort 중 어느 것인지 명시)
```

### 4.3 workflow.yaml.harness 섹션 (L2)

```yaml
workflow:
  team:
    enabled: true
    role_profiles: [...]

  harness:
    enabled: true
    generated_at: <ISO8601>
    domain: <도메인 string, 예: "ios-mobile">
    spec_id: <SPEC-PROJ-INIT-NNN>
    custom_agents:
      - name: <agent name>
        path: ".claude/agents/my-harness/<name>.md"
        invoke_in: [plan|run|sync]
    custom_skills:
      - name: <skill name>
        path: ".claude/skills/my-harness-<name>/SKILL.md"
        triggers_in: [plan|run|sync|design]
    chaining_rules:
      - phase: <plan|run|sync>
        before_specialist: <agent name>
        after_specialist: <agent name>
```

### 4.4 CLAUDE.md marker block (L3)

```markdown
## Project-Specific Configuration (Harness-Generated)
<!-- moai:harness-start id="SPEC-PROJ-INIT-NNN" generated="<ISO>" -->
**Domain**: <도메인>
**Harness level**: <thorough|standard|minimal>
**Updated**: <ISO date>

See @.moai/config/sections/workflow.yaml for team roles + harness chaining
See @.moai/harness/main.md for project-specific customization entry
See @.claude/agents/my-harness/<name>.md
See @.claude/skills/my-harness-<name>/SKILL.md
<!-- moai:harness-end -->
```

### 4.5 chaining-rules.yaml (L5)

```yaml
# .moai/harness/chaining-rules.yaml
version: 1
chains:
  - phase: plan
    when:
      agent: manager-spec
    insert_before: []
    insert_after: [my-harness/<arch-agent>]
  - phase: run
    when:
      agent: manager-tdd
    insert_before: [my-harness/<arch-agent>]
    insert_after: [my-harness/<engineer-agent>]
```

## 5. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| 16Q 인터뷰 token 폭발 (각 Round AskUserQuestion 응답 + meta-harness prompt 1회) | runtime.yaml.budget로 인터뷰 phase token 제한; Round 종료마다 in-memory 버퍼만 누적 (디스크 X). |
| moai update가 user area 침범 | path-prefix exclusion list 강화 (`internal/template/sync.go`); Phase 4 회귀 테스트 추가. |
| CLAUDE.md marker idempotency | id-based replace 로직 (동일 spec_id marker 존재 시 block 단위 갱신). |
| Layer 4 import line 누락 (사용자가 수동 삭제) | moai doctor가 누락 감지 → restore 안내. |
| Q13 "Advanced" 분기로 design-extension.md 생성 후 SPEC-V3R3-DESIGN-PIPELINE-001 미완성 | design-extension.md는 placeholder + warning ("DESIGN-PIPELINE-001 머지 후 실제 분기 활성"). |
| Template static import line이 사용자가 수정한 workflows/*.md와 충돌 | moai update가 한 줄 단위로만 갱신; 사용자 customization은 `.moai/harness/`에만 작성하도록 README.md 강제 안내. |
| **/clear mid-interview**: 사용자가 Round 1~3 도중 `/clear`를 실행하면 in-memory 버퍼 손실 + 부분 답변만 누적된 상태에서 세션 종료 (plan-auditor D8 보강) | Buffer는 매 Round 종료 시점에도 디스크에 commit하지 않으므로 `/clear` 시 깨끗한 abort 효과 (REQ-PH-010 자동 충족). 단, 사용자에게 "인터뷰 진행 중에는 `/clear` 시 답변 손실"을 Round 1 안내문에 명시 (project.md Phase 5 헤더). 재시작 시 처음 Round 1부터 다시 진행. |

## 6. Test Strategy

### 6.1 Unit Tests

- `internal/harness/interview_test.go`: 16Q schema validation, abort handling, conversation_language fallback.
- `internal/harness/layer1_test.go` ~ `layer5_test.go`: 각 Layer 단위 활성화 + idempotency.
- `internal/cli/doctor_harness_test.go`: 5-Layer 진단 + prefix 충돌 감지.

### 6.2 Integration Tests

- `internal/harness/integration_test.go`: end-to-end iOS 시나리오 (handoff §5.2 verbatim).
- `internal/cli/update_safety_test.go`: moai update 전후 user area 0 changes (`diff -rq`).

### 6.3 Manual Verification

- 새 빈 프로젝트 (`moai init my-ios-app`) → `/moai project` 인터뷰 → 새 세션 시작 → `/moai plan "FaceID auth"` → manager-spec이 ios-architect chain 사용 확인.

## 7. Open Questions

- **OQ1**: Round 4 Q15 (우선순위)는 Q16 (최종 확인) 직전인가, Q14 (특수 제약) 직후인가? — **결정**: Q15 (priority: thorough/standard/minimal) → Q16 (final confirm) 순서 고정.
- **OQ2**: design-extension.md를 Q13 "Advanced" 외에도 Q5 "Figma+Pencil 둘 다" 같은 답변에서도 생성해야 하는가? — **결정**: Q13 단독 분기 (REQ-PH-012). Q5는 design-pipeline SPEC 책임.
- **OQ3**: 16Q 답변이 16개 미만 (사용자가 일부 옵션 skip)이면 어떻게 처리? — **결정**: AskUserQuestion은 모든 옵션에 default + recommended 보장하므로 빈 답변 불가능; 그래도 누락 발생 시 Round 재실행.

## 8. References

- handoff: `.moai/release/v3r3-extreme-aggressive-handoff.md` §0.2 (5-Layer), §0.3 (활성화 표), §4.2 (3 SPECs 동시 작성), §5.2 (iOS 검증), §5.3 (moai update 안전성), §6 (Risks).
- 디자인 헌법 §2 (FROZEN), `.claude/rules/moai/design/constitution.md`.
- 운영 정책: CLAUDE.md §8 (User Interaction), `.claude/rules/moai/core/agent-common-protocol.md` (AskUserQuestion boundary).
- Reference SPEC: `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/plan.md` (구조 동형).
