---
id: CLAUDE-CODE-FEATURES-001
version: 0.0.1
status: draft
created: 2025-11-02
updated: 2025-11-02
priority: high
category: feature
labels:
  - claude-code
  - optimization
  - cost-reduction
  - performance
---

## HISTORY

### v0.0.1 (2025-11-02)
- **INITIAL**: MoAI-ADK v0.9.0을 위해 Claude Code v2.0.30+ 신규 기능 6개 통합
- **SCOPE**: Haiku Auto SonnetPlan, PreCompact Hook, Background Bash, Enhanced Grep, Plan Resume, TodoWrite Auto-Init
- **CONTEXT**: Claude Code 최신 기능을 MoAI-ADK에 통합하여 비용 70-90% 절감, 성능 4-6배 향상

---

# SPEC: Claude Code v2.0.30+ Features Integration


MoAI-ADK v0.9.0을 위한 Claude Code 신규 기능 6개 통합 명세

---

## 1. Environment (환경)

### Required Dependencies
- **Claude Code**: v2.0.30 이상
- **MoAI-ADK**: v0.12.0+ (Python 3.13+)
- **Configuration**: `.moai/config.json` 설정 완료
- **Workflow**: 기존 SPEC-first TDD 워크플로우 활성화

### System Requirements
- Python 3.13 or higher
- Git with branch management enabled
- GitHub CLI (`gh`) for PR operations
- Network access for Claude API calls

### Development Environment
- VSCode with Claude Code extension installed
- Terminal with bash/zsh support
- Sufficient disk space for context caching (min 100MB)

---

## 2. Assumptions (가정)

### Technical Assumptions
1. **Compatibility**: 모든 신규 기능은 기존 Alfred 워크플로우와 호환됨
2. **API Stability**: Claude Code의 최신 API가 안정적이고 계속 지원됨
3. **Model Availability**: Sonnet 4.5와 Haiku 4.5 모델이 지속적으로 제공됨
4. **Backward Compatibility**: 기존 v0.8.0 프로젝트가 v0.9.0으로 원활하게 마이그레이션됨

### Operational Assumptions
1. **User Access**: 사용자는 Claude Code 설정 파일을 수정할 권한이 있음
2. **Testing Environment**: 개발 환경에서 테스트 실행 및 검증 가능
3. **Documentation**: 각 기능에 대한 사용자 가이드가 제공됨

---

## 3. Requirements (핵심 요구사항)

### 3.1 Feature 1: Haiku 4.5 Auto SonnetPlan Mode

**Ubiquitous Requirement:**
- Plan 에이전트는 **Sonnet 4.5 모델**을 사용하여 고품질 계획을 수립해야 함
- 실행 에이전트는 **Haiku 4.5 모델**을 사용하여 비용을 최적화해야 함
- 모델 전환은 자동으로 이루어지며 사용자 개입 불필요

**Event-Driven Requirements:**
- **WHEN** spec-builder 에이전트가 STEP 1에서 호출되면
  - **THEN** Sonnet 4.5 모델로 분석하여 계획 보고서를 생성해야 함
  - **AND** 계획 품질 검증을 수행해야 함

- **WHEN** tdd-implementer가 RED-GREEN-REFACTOR 사이클을 시작하면
  - **THEN** Haiku 4.5 모델로 효율적인 구현을 진행해야 함
  - **AND** 실행 속도가 기존 대비 4-6배 향상되어야 함

**State-Driven Requirements:**
- **WHILE** 계획 모드(Plan Agent)가 활성일 때
  - **THEN** Sonnet을 사용하여 고품질 분석을 수행해야 함
  - **AND** 복잡한 의사결정 로직을 처리해야 함

- **WHILE** 구현 모드(Execution Agent)가 활성일 때
  - **THEN** Haiku로 빠른 실행을 제공해야 함
  - **AND** 비용이 기존 대비 70-90% 절감되어야 함

**Unwanted Behaviors:**
- **IF** Haiku가 복잡한 분석을 수행할 수 없으면
  - **THEN** 자동으로 Sonnet으로 폴백하고 사용자에게 알려야 함
  - **AND** 폴백 이유를 로깅해야 함

- **IF** 모델 전환 중 에러 발생 시
  - **THEN** 원래 모델로 복구하고 에러를 로깅해야 함
  - **AND** 작업을 계속 진행해야 함 (graceful degradation)

---

### 3.2 Feature 2: PreCompact Hook

**Ubiquitous Requirement:**
- 토큰 사용률이 **80% 이상**이면, 현재 컨텍스트를 자동으로 저장해야 함
- 상태 저장은 `.moai/memory/session-state.json`에 수행됨
- 복원 시 이전 작업 컨텍스트를 완벽하게 재현해야 함

**Event-Driven Requirements:**
- **WHEN** 토큰 사용률이 80%를 초과하면
  - **THEN** PreCompact Hook이 트리거되어 현재 상태를 저장해야 함
  - **AND** 사용자에게 상태 저장 알림을 표시해야 함

- **WHEN** 새 세션이 시작되면
  - **THEN** 저장된 상태를 자동으로 감지해야 함
  - **AND** 복원 여부를 사용자에게 물어봐야 함 (AskUserQuestion)

**State-Driven Requirements:**
- **WHILE** 토큰 사용률 모니터링 중
  - **THEN** 1000 토큰마다 사용률을 체크해야 함
  - **AND** 70% 도달 시 사용자에게 경고를 표시해야 함

**Unwanted Behaviors:**
- **IF** 토큰 저장에 실패하면
  - **THEN** 에러를 로깅하고 계속 진행해야 함 (graceful degradation)
  - **AND** 사용자에게 수동 저장을 권장해야 함

- **IF** 저장된 상태가 손상되면
  - **THEN** 복원 불가 메시지를 표시하고 새로 시작해야 함
  - **AND** 손상된 파일을 백업하고 삭제해야 함

---

### 3.3 Feature 3: Background Bash Commands

**Ubiquitous Requirement:**
- 테스트 및 빌드 명령어는 백그라운드에서 실행되어야 함
- 사용자는 실행 중에 새로운 작업을 계속 진행할 수 있어야 함
- 백그라운드 작업 상태를 실시간으로 모니터링할 수 있어야 함

**Event-Driven Requirements:**
- **WHEN** tdd-implementer가 pytest를 실행할 때
  - **THEN** `run_in_background=true` 옵션으로 백그라운드 실행해야 함
  - **AND** 작업 ID와 로그 파일 경로를 반환해야 함

- **WHEN** 백그라운드 작업이 완료되면
  - **THEN** 사용자에게 완료 알림을 제공해야 함
  - **AND** 결과 요약(성공/실패, 실행 시간)을 표시해야 함

- **WHEN** 사용자가 작업 취소를 요청하면
  - **THEN** 백그라운드 프로세스를 안전하게 종료해야 함
  - **AND** 부분 결과를 저장해야 함

**Optional Features:**
- **WHERE** 사용자가 명시적으로 요청하면
  - **THEN** 작업 로그를 실시간으로 스트림할 수 있음
  - **AND** 진행률 표시줄을 제공할 수 있음

**Unwanted Behaviors:**
- **IF** 백그라운드 작업이 타임아웃(10분)되면
  - **THEN** 자동으로 종료하고 사용자에게 알려야 함
  - **AND** 로그를 `.moai/logs/background-tasks/`에 저장해야 함

---

### 3.4 Feature 4: Enhanced Grep Tool

**Ubiquitous Requirement:**
- Grep 도구는 **multiline 패턴 매칭**과 **head_limit 파라미터**를 지원해야 함
- 기존 Grep 기능과 하위 호환성을 유지해야 함
- 성능이 기존 대비 개선되어야 함

**Event-Driven Requirements:**
  - **THEN** `head_limit`을 사용하여 처음 N개 결과만 반환할 수 있어야 함
  - **AND** 결과가 많을 경우 사용자에게 범위 좁히기를 제안해야 함

- **WHEN** 복잡한 정규식이 사용되면
  - **THEN** `multiline=true` 모드에서 여러 줄에 걸친 매칭을 수행해야 함
  - **AND** 매칭된 전체 블록을 반환해야 함

**State-Driven Requirements:**
- **WHILE** 대용량 파일(>10MB)을 검색할 때
  - **THEN** 스트리밍 모드로 처리하여 메모리 사용을 최소화해야 함

**Optional Features:**
- **WHERE** 사용자가 요청하면
  - **THEN** 컨텍스트 라인 수(`-A`, `-B`, `-C`)를 지정할 수 있음

---

### 3.5 Feature 5: Plan Resume

**Ubiquitous Requirement:**
- 사용자는 `--resume-plan` 옵션을 사용하여 이전 계획 상태를 복원할 수 있어야 함
- 복원된 계획은 수정 가능하며, 수정 내용이 자동으로 반영되어야 함
- 계획 이력을 추적하여 롤백 가능해야 함

**Event-Driven Requirements:**
- **WHEN** 사용자가 `--resume-plan`을 지정하면
  - **THEN** 마지막 저장된 계획 상태를 로드하고 수정할 수 있어야 함
  - **AND** 이전 계획과의 차이점을 표시해야 함

- **WHEN** 계획 수정이 완료되면
  - **THEN** 수정된 내용이 TodoWrite 작업 목록에 자동으로 반영되어야 함
  - **AND** 변경 이력을 `.moai/memory/plan-history.json`에 저장해야 함

**State-Driven Requirements:**
- **WHILE** 계획 복원 모드가 활성일 때
  - **THEN** 원본 계획을 백업하고 수정본을 별도 저장해야 함

**Unwanted Behaviors:**
- **IF** 저장된 계획이 없으면
  - **THEN** 사용자에게 새 계획 생성을 안내해야 함

---

### 3.6 Feature 6: TodoWrite Auto-Initialization

**Ubiquitous Requirement:**
- Plan 에이전트의 결과에서 TodoWrite 작업 목록이 자동으로 초기화되어야 함
- 작업 우선순위와 의존성이 자동으로 설정되어야 함
- 기존 수동 TodoWrite 호출과 호환되어야 함

**Event-Driven Requirements:**
- **WHEN** spec-builder가 계획 분석을 완료하면
  - **THEN** 결과에서 자동으로 TodoWrite의 작업 항목이 생성되어야 함
  - **AND** 각 작업은 `content`, `activeForm`, `status` 필드를 포함해야 함

- **WHEN** TodoWrite이 자동으로 초기화되면
  - **THEN** 모든 작업의 상태는 "pending"이어야 함
  - **AND** 첫 번째 작업만 "in_progress"로 설정되어야 함

**State-Driven Requirements:**
- **WHILE** 작업 실행 중
  - **THEN** TodoWrite 상태가 자동으로 업데이트되어야 함 (pending → in_progress → completed)

**Optional Features:**
- **WHERE** 사용자가 수동으로 TodoWrite을 수정하면
  - **THEN** 자동 초기화가 비활성화되고 수동 모드로 전환되어야 함

---

## 4. Specifications (구현 명세)

### 4.1 Architecture Design

```
Alfred (SuperAgent)
    ├─ Plan Agent (Sonnet 4.5) ────────┐
    │   ├─ spec-builder                │
    │   ├─ implementation-planner       │ Auto Model Selection
    │   └─ TodoWrite Auto-Init ◄────────┤
    │                                   │
    ├─ Execution Agents (Haiku 4.5) ───┘
    │   ├─ tdd-implementer
    │   ├─ doc-syncer
    │   └─ tag-agent
    │
    ├─ PreCompact Hook (Session Manager)
    │   ├─ Token Usage Monitor (80% threshold)
    │   └─ State Persistence (.moai/memory/)
    │
    ├─ Background Bash Handler
    │   ├─ run_in_background=true
    │   └─ BashOutput monitoring
    │
    └─ Enhanced Grep Engine
        ├─ multiline=true support
        └─ head_limit parameter
```

### 4.2 Data Flow

1. **Planning Phase** (Sonnet 4.5):
   - User request → Alfred → Plan Agent (Sonnet)
   - Plan Agent generates structured task breakdown
   - TodoWrite auto-initialized from Plan results

2. **Execution Phase** (Haiku 4.5):
   - TodoWrite tasks → Execution Agents (Haiku)
   - Background Bash for long-running commands
   - Enhanced Grep for efficient code search

3. **Context Management**:
   - PreCompact Hook monitors token usage
   - Auto-save at 80% threshold
   - Resume capability via `--resume-plan`

### 4.3 Configuration Schema

**`.moai/config.json` 확장:**

```json
{
  "claude_code": {
    "version": "2.0.30+",
    "features": {
      "auto_sonnet_plan": {
        "enabled": true,
        "plan_model": "claude-sonnet-4-5",
        "exec_model": "claude-haiku-4-5",
        "fallback_threshold": 0.7
      },
      "precompact_hook": {
        "enabled": true,
        "token_threshold": 0.8,
        "auto_save": true,
        "state_file": ".moai/memory/session-state.json"
      },
      "background_bash": {
        "enabled": true,
        "timeout_ms": 600000,
        "log_dir": ".moai/logs/background-tasks/"
      },
      "enhanced_grep": {
        "enabled": true,
        "multiline_support": true,
        "default_head_limit": 50
      },
      "plan_resume": {
        "enabled": true,
        "history_file": ".moai/memory/plan-history.json",
        "max_history": 10
      },
      "todowrite_auto_init": {
        "enabled": true,
        "auto_start_first_task": true
      }
    }
  }
}
```

### 4.4 Implementation Priority

**Week 1: Foundation (Features 1, 4, 6)**
- Haiku Auto SonnetPlan Mode (Feature 1)
- Enhanced Grep Tool (Feature 4)
- TodoWrite Auto-Init (Feature 6)

**Week 2-3: Core (Features 3, 5)**
- Background Bash Commands (Feature 3)
- Plan Resume (Feature 5)

**Week 4: Advanced (Feature 2) + Integration**
- PreCompact Hook (Feature 2)
- Integration testing across all features
- Performance benchmarking

---

## 5. Constraints (제약사항)

### Technical Constraints
1. **Backward Compatibility**: v0.8.0 프로젝트는 v0.9.0으로 자동 마이그레이션되어야 함
2. **Performance**: 각 기능은 기존 성능을 저하시키지 않아야 함
3. **Token Limits**: Haiku 모드에서도 200K 토큰 버짓 준수

### Operational Constraints
1. **User Experience**: 모든 신규 기능은 선택적 활성화 가능해야 함 (config.json)
2. **Error Handling**: 각 기능 실패 시 graceful degradation 제공
3. **Documentation**: 각 기능에 대한 사용자 가이드 필수 제공

---


### TAG Hierarchy
- SPEC:CLAUDE-CODE-FEATURES-001 (이 문서)

### Related SPECs

---

## 7. Acceptance Criteria

### Definition of Done
- ✅ 모든 6개 기능이 `.moai/config.json`에서 활성화/비활성화 가능
- ✅ Integration tests 통과율 95% 이상
- ✅ 비용 절감 목표 70-90% 달성 (Haiku mode 기준)
- ✅ 성능 향상 목표 4-6배 달성 (Haiku mode 기준)
- ✅ v0.8.0 → v0.9.0 자동 마이그레이션 성공
- ✅ 사용자 가이드 문서 작성 완료

### Quality Gates
1. **TRUST 5 Compliance**: 모든 코드가 TRUST 5 원칙 준수
2. **Test Coverage**: 신규 코드 90% 이상 커버리지
3. **Linting**: Ruff 100% 통과
4. **Type Checking**: Pyright strict mode 통과

---

## 8. References

### Internal Documentation
- Skill("moai-alfred-agent-guide") - Agent selection criteria
- Skill("moai-alfred-rules") - Skill invocation rules
- Skill("moai-alfred-config-schema") - Configuration reference
- Skill("moai-alfred-practices") - Best practices

### External Resources
- Claude Code Release Notes v2.0.30+
- Anthropic Model Documentation (Sonnet 4.5, Haiku 4.5)
- MoAI-ADK Architecture Guide

---

**Document Version**: 0.0.1
**Last Updated**: 2025-11-02
**Next Review**: Upon implementation completion
