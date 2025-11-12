# Implementation Plan: Claude Code Features Integration

**SPEC ID**: CLAUDE-CODE-FEATURES-001
**Version**: 0.0.1
**Created**: 2025-11-02

---

## 목표 (Objective)

Claude Code v2.0.30+ 신규 기능 6개를 MoAI-ADK v0.9.0에 통합하여:
- 비용 70-90% 절감 (Haiku Auto SonnetPlan Mode)
- 성능 4-6배 향상 (Haiku execution model)
- 개발자 경험 개선 (Background Bash, Plan Resume)
- 컨텍스트 관리 자동화 (PreCompact Hook)

---

## Implementation Timeline

### Week 1: Foundation Layer (Features 1, 4, 6)

#### Priority 1: Haiku Auto SonnetPlan Mode (Feature 1)
**Goal**: Plan 에이전트는 Sonnet, 실행 에이전트는 Haiku 사용

**Tasks**:
1. **Agent Model Configuration**
   - `.moai/config.json`에 모델 설정 추가
   - Plan Agent (spec-builder, implementation-planner)에 Sonnet 4.5 지정
   - Execution Agents (tdd-implementer, doc-syncer, tag-agent)에 Haiku 4.5 지정

2. **Model Selection Logic**
   - `src/moai_adk/core/agents/model_selector.py` 생성
   - Plan mode vs Execution mode 자동 감지
   - Fallback 로직 구현 (Haiku → Sonnet when complexity threshold exceeded)

3. **Integration with Existing Agents**
   - `spec-builder.py` 수정: Sonnet 모델 강제 사용
   - `tdd-implementer.py` 수정: Haiku 모델 강제 사용
   - 모델 전환 로깅 추가

**Acceptance**:
- ✅ spec-builder가 Sonnet 4.5를 사용하는지 로그로 확인
- ✅ tdd-implementer가 Haiku 4.5를 사용하는지 로그로 확인
- ✅ 비용 절감 70% 이상 달성 (기존 대비)

---

#### Priority 2: Enhanced Grep Tool (Feature 4)
**Goal**: multiline 패턴 매칭 및 head_limit 파라미터 지원

**Tasks**:
1. **Grep Engine Enhancement**
   - `src/moai_adk/core/grep_engine.py` 생성
   - `multiline=true` 파라미터 추가
   - `head_limit` 파라미터 추가 (기본값: 50)

2. **Integration with tag-agent**
   - `tag-agent.py` 수정: Enhanced Grep 호출 로직 추가

3. **Performance Optimization**
   - 대용량 파일(>10MB) 스트리밍 모드 구현
   - 메모리 사용량 최소화

**Acceptance**:
- ✅ multiline 패턴 매칭 테스트 통과
- ✅ head_limit이 정확히 N개 결과만 반환하는지 검증
- ✅ 대용량 파일(100MB) 검색 시 메모리 사용량 <500MB

---

#### Priority 3: TodoWrite Auto-Initialization (Feature 6)
**Goal**: Plan 에이전트 결과에서 TodoWrite 자동 초기화

**Tasks**:
1. **TodoWrite Initializer**
   - `src/moai_adk/core/todowrite_initializer.py` 생성
   - Plan Agent 결과 파싱 로직 구현
   - TodoWrite 작업 목록 자동 생성

2. **Integration with Plan Agent**
   - `spec-builder.py` 수정: 계획 완료 시 TodoWrite 자동 초기화 호출
   - 작업 우선순위 및 의존성 자동 설정

3. **Backward Compatibility**
   - 기존 수동 TodoWrite 호출과 충돌하지 않도록 조건부 활성화
   - `.moai/config.json`에서 `todowrite_auto_init.enabled` 확인

**Acceptance**:
- ✅ spec-builder 완료 후 TodoWrite이 자동으로 초기화됨
- ✅ 모든 작업이 "pending" 상태로 시작
- ✅ 첫 번째 작업만 "in_progress"로 설정됨

---

### Week 2-3: Core Features (Features 3, 5)

#### Priority 4: Background Bash Commands (Feature 3)
**Goal**: 테스트/빌드 명령어를 백그라운드에서 실행

**Tasks**:
1. **Background Bash Handler**
   - `src/moai_adk/core/bash_handler.py` 생성
   - `run_in_background=true` 파라미터 지원
   - 작업 ID 및 로그 파일 경로 반환

2. **Integration with tdd-implementer**
   - `tdd-implementer.py` 수정: pytest 실행 시 `run_in_background=true` 사용
   - 백그라운드 작업 완료 알림 구현

3. **Task Monitoring**
   - 백그라운드 작업 상태 모니터링 (실시간 로그 스트림)
   - 타임아웃 처리 (기본값: 10분)
   - 작업 취소 기능 구현

**Acceptance**:
- ✅ pytest가 백그라운드에서 실행되는지 확인
- ✅ 작업 완료 시 사용자에게 알림이 표시됨
- ✅ 타임아웃(10분) 후 자동 종료됨

---

#### Priority 5: Plan Resume (Feature 5)
**Goal**: 이전 계획 상태를 복원하고 수정 가능

**Tasks**:
1. **Plan Manager**
   - `src/moai_adk/core/plan_manager.py` 생성
   - 계획 상태 저장 로직 (`.moai/memory/plan-history.json`)
   - `--resume-plan` 옵션 처리

2. **Plan History Tracking**
   - 계획 수정 이력 저장 (최대 10개)
   - 이전 계획과의 차이점 비교 (diff)
   - 롤백 기능 구현

3. **Integration with spec-builder**
   - `spec-builder.py` 수정: `--resume-plan` 옵션 감지
   - 복원된 계획 수정 후 TodoWrite 자동 업데이트

**Acceptance**:
- ✅ `--resume-plan` 실행 시 마지막 계획이 로드됨
- ✅ 계획 수정 후 TodoWrite이 자동으로 업데이트됨
- ✅ 계획 이력이 `.moai/memory/plan-history.json`에 저장됨

---

### Week 4: Advanced Feature + Integration (Feature 2)

#### Priority 6: PreCompact Hook (Feature 2)
**Goal**: 토큰 사용률 80% 이상 시 컨텍스트 자동 저장

**Tasks**:
1. **PreCompact Hook Implementation**
   - `.claude/hooks/PreCompact.py` 생성
   - 토큰 사용률 모니터링 (1000 토큰마다 체크)
   - 80% 도달 시 상태 자동 저장

2. **State Persistence**
   - 현재 컨텍스트를 `.moai/memory/session-state.json`에 저장
   - 작업 목록, 계획 상태, 진행 상황 포함
   - 복원 시 완전한 컨텍스트 재현

3. **Session Recovery**
   - 새 세션 시작 시 저장된 상태 자동 감지
   - 사용자에게 복원 여부 확인 (AskUserQuestion)
   - 손상된 상태 파일 처리 (백업 후 삭제)

**Acceptance**:
- ✅ 토큰 사용률 80% 도달 시 상태가 자동 저장됨
- ✅ 새 세션에서 저장된 상태를 복원할 수 있음
- ✅ 손상된 상태 파일이 안전하게 처리됨

---

#### Priority 7: Integration Testing
**Goal**: 모든 6개 기능이 함께 작동하는지 검증

**Tasks**:
1. **End-to-End Test Scenarios**
   - Scenario 1: Haiku Auto SonnetPlan + TodoWrite Auto-Init
   - Scenario 2: Background Bash + Enhanced Grep
   - Scenario 3: PreCompact Hook + Plan Resume
   - Scenario 4: 전체 워크플로우 통합 테스트

2. **Performance Benchmarking**
   - 비용 절감률 측정 (목표: 70-90%)
   - 성능 향상률 측정 (목표: 4-6배)
   - 토큰 사용량 최적화 검증

3. **Backward Compatibility Testing**
   - v0.8.0 프로젝트를 v0.9.0으로 마이그레이션
   - 기존 SPEC 문서 호환성 검증
   - 기존 명령어 동작 검증

**Acceptance**:
- ✅ Integration tests 통과율 95% 이상
- ✅ 비용 절감 70-90% 달성
- ✅ 성능 4-6배 향상 달성
- ✅ v0.8.0 → v0.9.0 마이그레이션 성공

---

## Technical Approach

### 1. Model Selection Strategy

```python
# src/moai_adk/core/agents/model_selector.py

class ModelSelector:
    def select_model(self, agent_type: str, task_complexity: float) -> str:
        """
        Plan Agent → Sonnet 4.5
        Execution Agent → Haiku 4.5
        Fallback: complexity > 0.7 → Sonnet
        """
        if agent_type in ["spec-builder", "implementation-planner"]:
            return "claude-sonnet-4-5"

        if task_complexity > 0.7:
            logger.warning(f"Complexity {task_complexity} > 0.7, falling back to Sonnet")
            return "claude-sonnet-4-5"

        return "claude-haiku-4-5"
```

### 2. Background Bash Handler

```python
# src/moai_adk/core/bash_handler.py

class BackgroundBashHandler:
    def execute(self, command: str, timeout_ms: int = 600000) -> dict:
        """
        Execute command in background
        Returns: {task_id, log_file, status}
        """
        task_id = f"bg-{uuid.uuid4().hex[:8]}"
        log_file = f".moai/logs/background-tasks/{task_id}.log"

        # Start background process
        process = subprocess.Popen(
            command,
            stdout=open(log_file, 'w'),
            stderr=subprocess.STDOUT,
            shell=True
        )

        return {
            "task_id": task_id,
            "log_file": log_file,
            "status": "running"
        }
```

### 3. PreCompact Hook

```python
# .claude/hooks/PreCompact.py

class PreCompactHook:
    def check_token_usage(self, current_tokens: int, max_tokens: int) -> None:
        """
        Check token usage and auto-save at 80% threshold
        """
        usage_ratio = current_tokens / max_tokens

        if usage_ratio >= 0.8:
            logger.info(f"Token usage {usage_ratio:.1%} >= 80%, triggering PreCompact")
            self.save_session_state()
            self.notify_user("Context saved due to high token usage")

        elif usage_ratio >= 0.7:
            logger.warning(f"Token usage {usage_ratio:.1%} approaching limit")
```

### 4. Enhanced Grep Engine

```python
# src/moai_adk/core/grep_engine.py

class EnhancedGrep:
    def search(
        self,
        pattern: str,
        path: str,
        multiline: bool = False,
        head_limit: int = 50
    ) -> list[str]:
        """
        Enhanced grep with multiline and head_limit support
        """
        args = ["rg", pattern, path]

        if multiline:
            args.extend(["-U", "--multiline-dotall"])

        results = subprocess.check_output(args).decode().splitlines()

        return results[:head_limit]
```

---

## Architecture Design

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Alfred SuperAgent                         │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  Model Selector (Auto Sonnet/Haiku)                       │  │
│  │  - Plan Mode: Sonnet 4.5                                  │  │
│  │  - Execution Mode: Haiku 4.5                              │  │
│  │  - Fallback Logic: complexity > 0.7 → Sonnet             │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
         │                           │                          │
         ▼                           ▼                          ▼
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│  Plan Agents    │      │ Execution Agents│      │  Hook Layer     │
│  (Sonnet 4.5)   │      │  (Haiku 4.5)    │      │                 │
│                 │      │                 │      │                 │
│ • spec-builder  │      │ • tdd-impl      │      │ • PreCompact    │
│ • impl-planner  │      │ • doc-syncer    │      │   (Token 80%)   │
│                 │      │ • tag-agent     │      │                 │
└─────────────────┘      └─────────────────┘      └─────────────────┘
         │                           │                          │
         └───────────────┬───────────┘                          │
                         ▼                                      │
              ┌─────────────────────┐                           │
              │ TodoWrite Auto-Init │                           │
              │ (from Plan results) │                           │
              └─────────────────────┘                           │
                         │                                      │
                         ▼                                      │
              ┌─────────────────────┐                           │
              │ Background Bash     │                           │
              │ (run_in_background) │                           │
              └─────────────────────┘                           │
                         │                                      │
                         ▼                                      │
              ┌─────────────────────┐                           │
              │ Enhanced Grep       │                           │
              │ (multiline, limit)  │                           │
              └─────────────────────┘                           │
                                                                │
                         ┌──────────────────────────────────────┘
                         ▼
              ┌─────────────────────┐
              │ Session Manager     │
              │ • State Save/Restore│
              │ • Plan Resume       │
              └─────────────────────┘
```

### Data Flow

1. **User Request** → Alfred SuperAgent
2. **Model Selection** → Plan Agent (Sonnet) OR Execution Agent (Haiku)
3. **Plan Agent** (Sonnet) → Structured task breakdown
4. **TodoWrite Auto-Init** → Parse Plan results, create task list
5. **Execution Agents** (Haiku) → Process tasks with Background Bash + Enhanced Grep
6. **PreCompact Hook** → Monitor token usage, auto-save at 80%
7. **Session Manager** → Save/restore state, enable Plan Resume

---

## Risk Management

### Risk 1: Plan Agent Complexity Overload
**Risk**: Plan Agent가 너무 복잡한 작업을 받으면 Sonnet도 실패할 수 있음

**Mitigation**:
- Complexity scoring 로직 구현 (0.0-1.0 scale)
- 0.9 이상 시 사용자에게 작업 분할 제안
- Fallback chain: Haiku → Sonnet → Opus (manual escalation)

---

### Risk 2: TodoWrite Logic Conflicts
**Risk**: 기존 수동 TodoWrite 호출과 Auto-Init 충돌 가능성

**Mitigation**:
- `.moai/config.json`에서 `todowrite_auto_init.enabled` 확인
- 수동 TodoWrite 감지 시 Auto-Init 비활성화
- 명확한 로깅: "TodoWrite manual mode detected, disabling auto-init"

---

### Risk 3: Template Synchronization
**Risk**: 패키지 템플릿(`src/moai_adk/templates/.claude/`)과 로컬 프로젝트(`/`) 간 동기화 누락

**Mitigation**:
- **자동 검증**: 모든 `.claude/` 파일 변경 시 템플릿 동기화 검증 스크립트 실행
- **Pre-commit Hook**: 템플릿 파일 수정 시 로컬과 패키지 템플릿 동시 업데이트 강제
- **CI/CD Check**: GitHub Actions에서 동기화 검증 (diff 0개 확인)

---

### Risk 4: Background Bash Timeout Handling
**Risk**: 백그라운드 작업이 타임아웃되면 부분 결과 손실 가능

**Mitigation**:
- 부분 결과를 `.moai/logs/background-tasks/{task_id}.log`에 저장
- 타임아웃 시 로그 파일 경로를 사용자에게 제공
- 재실행 옵션 제공 (--retry-task {task_id})

---

### Risk 5: PreCompact Hook Performance Overhead
**Risk**: 토큰 사용률 체크가 성능 저하를 유발할 수 있음

**Mitigation**:
- 1000 토큰마다만 체크 (매 요청마다 체크 X)
- 비동기 처리로 메인 스레드 블로킹 방지
- 캐싱: 마지막 체크 이후 토큰 변화량만 계산

---

## Dependencies

### Internal Dependencies
- `src/moai_adk/core/agents/` - 기존 에이전트 코드 수정
- `.claude/hooks/` - PreCompact Hook 추가
- `.moai/config.json` - 신규 설정 필드 추가

### External Dependencies
- Claude Code v2.0.30+ (필수)
- Claude API (Sonnet 4.5, Haiku 4.5 모델 지원)
- Python 3.13+ (async/await 기능 사용)

### Template Synchronization
- **Critical**: 모든 변경사항은 다음 위치에 동시 적용:
  - Local: `.claude/`, `.moai/`
  - Package: `src/moai_adk/templates/.claude/`, `src/moai_adk/templates/.moai/`

---

## Success Criteria

### Quantitative Metrics
- ✅ **Cost Reduction**: 70-90% (Haiku mode 기준)
- ✅ **Performance Improvement**: 4-6x faster (Haiku mode 기준)
- ✅ **Test Coverage**: 90% 이상
- ✅ **Integration Tests**: 95% 통과율

### Qualitative Metrics
- ✅ **User Experience**: 백그라운드 작업으로 인터럽션 최소화
- ✅ **Developer Experience**: Plan Resume로 컨텍스트 유지 용이
- ✅ **Reliability**: PreCompact Hook으로 토큰 초과 방지

---

## Next Steps

1. **Week 1 시작**: Haiku Auto SonnetPlan Mode 구현
2. **CI/CD 설정**: 템플릿 동기화 자동 검증
3. **Documentation**: 각 기능별 사용자 가이드 작성
4. **Migration Guide**: v0.8.0 → v0.9.0 마이그레이션 가이드 작성

---

**Plan Version**: 0.0.1
**Last Updated**: 2025-11-02
**Estimated Completion**: 4 weeks (우선순위 기반, 시간 예측 아님)
