# Acceptance Criteria: Claude Code Features Integration

**SPEC ID**: CLAUDE-CODE-FEATURES-001
**Version**: 0.0.1
**Created**: 2025-11-02

---

## Test Scenarios (Given-When-Then Format)

### Scenario 1: Haiku Auto SonnetPlan - Cost Reduction Verification

**Feature**: Haiku Auto SonnetPlan Mode
**Goal**: 비용 70-90% 절감 검증

**Given**:
- MoAI-ADK v0.9.0 설치 완료
- `.moai/config.json`에서 `auto_sonnet_plan.enabled: true`
- 사용자가 `/alfred:1-plan "User authentication feature"` 실행

**When**:
- spec-builder 에이전트가 SPEC 분석 시작
- 계획 수립 단계 (STEP 1)에서 모델 선택

**Then**:
- ✅ spec-builder는 **Sonnet 4.5** 모델을 사용해야 함
- ✅ 로그에 "Using claude-sonnet-4-5 for planning mode" 메시지 출력
- ✅ 비용 측정: Plan 단계에서 Sonnet 비용 발생
- ✅ 비용 절감률: 전체 워크플로우에서 70-90% 절감 (Haiku 실행 모드 포함)

**Quality Gate**:
- Cost reduction ≥ 70%
- Plan quality score ≥ 0.9 (based on EARS compliance)

---

### Scenario 2: Haiku Auto SonnetPlan - Performance Improvement

**Feature**: Haiku Auto SonnetPlan Mode
**Goal**: 성능 4-6배 향상 검증

**Given**:
- spec-builder가 SPEC 생성 완료 (Sonnet으로 계획 수립)
- TodoWrite 작업 목록 초기화됨
- `/alfred:2-run SPEC-XXX-001` 실행 준비

**When**:
- tdd-implementer 에이전트가 RED-GREEN-REFACTOR 사이클 시작
- 코드 구현 및 테스트 실행

**Then**:
- ✅ tdd-implementer는 **Haiku 4.5** 모델을 사용해야 함
- ✅ 로그에 "Using claude-haiku-4-5 for execution mode" 메시지 출력
- ✅ 실행 속도: 기존 Sonnet 대비 4-6배 빠름
- ✅ 구현 품질: TRUST 5 원칙 준수 (Test Coverage ≥ 90%)

**Quality Gate**:
- Performance improvement ≥ 4x
- Test coverage ≥ 90%
- Linting: 0 errors

---

### Scenario 3: Haiku Auto SonnetPlan - Fallback Mechanism

**Feature**: Haiku Auto SonnetPlan Mode (Fallback)
**Goal**: Haiku가 복잡도를 처리 못할 때 Sonnet으로 자동 폴백

**Given**:
- tdd-implementer가 복잡한 알고리즘 구현 중
- Task complexity score: 0.85 (> 0.7 threshold)

**When**:
- Haiku 모델로 구현 시도
- Complexity analyzer가 0.85 점수 반환

**Then**:
- ✅ 로그에 "Complexity 0.85 > 0.7, falling back to Sonnet" 메시지 출력
- ✅ 자동으로 **Sonnet 4.5** 모델로 전환
- ✅ 사용자에게 알림: "Using Sonnet for complex task due to high complexity"
- ✅ 작업이 성공적으로 완료됨

**Quality Gate**:
- Fallback logic triggers correctly
- No task failures due to model limitations

---

### Scenario 4: PreCompact Hook - Automatic Context Saving

**Feature**: PreCompact Hook
**Goal**: 토큰 사용률 80% 도달 시 자동 저장

**Given**:
- 사용자가 긴 세션에서 여러 SPEC 작업 중
- 현재 토큰 사용량: 160,000 / 200,000 (80%)

**When**:
- 다음 요청으로 토큰 사용률이 80%를 초과
- PreCompact Hook이 트리거됨

**Then**:
- ✅ `.moai/memory/session-state.json`에 현재 상태 저장
- ✅ 로그에 "Token usage 80.0% >= 80%, triggering PreCompact" 메시지 출력
- ✅ 사용자에게 알림: "Context saved due to high token usage"
- ✅ 저장 내용: 작업 목록, 계획 상태, 진행 상황 포함

**Quality Gate**:
- State file created successfully
- All critical context preserved (TodoWrite, Plan, Progress)

---

### Scenario 5: PreCompact Hook - Session Recovery

**Feature**: PreCompact Hook (Session Recovery)
**Goal**: 새 세션에서 이전 상태 복원

**Given**:
- 이전 세션에서 `.moai/memory/session-state.json` 저장됨
- 사용자가 `/clear` 실행 후 새 세션 시작

**When**:
- 새 세션 시작 시 PreCompact Hook이 저장된 상태 감지
- 사용자에게 복원 여부 확인 (AskUserQuestion)

**Then**:
- ✅ AskUserQuestion으로 "이전 세션 상태를 복원하시겠습니까?" 질문
- ✅ 사용자가 "Yes" 선택 시 상태 복원
- ✅ TodoWrite 작업 목록, 계획 상태, 진행 상황 모두 복원됨
- ✅ 로그에 "Session state restored from .moai/memory/session-state.json" 메시지 출력

**Quality Gate**:
- State restoration 100% accurate
- No data loss during recovery

---

### Scenario 6: Background Bash - Long-Running Test Execution

**Feature**: Background Bash Commands
**Goal**: pytest를 백그라운드에서 실행하고 사용자는 다른 작업 계속 진행

**Given**:
- tdd-implementer가 테스트 실행 준비 (pytest 명령어)
- 예상 실행 시간: 5분

**When**:
- tdd-implementer가 `run_in_background=true` 옵션으로 pytest 실행
- 사용자는 백그라운드 실행 중 다른 작업 시작

**Then**:
- ✅ pytest가 백그라운드에서 실행됨
- ✅ 작업 ID 반환: `bg-a1b2c3d4`
- ✅ 로그 파일 경로 반환: `.moai/logs/background-tasks/bg-a1b2c3d4.log`
- ✅ 백그라운드 실행 중 사용자는 다른 명령어 실행 가능
- ✅ pytest 완료 후 알림: "Background task bg-a1b2c3d4 completed successfully"

**Quality Gate**:
- Background execution does not block user
- Log file contains complete test results

---

### Scenario 7: Background Bash - Timeout Handling

**Feature**: Background Bash Commands (Timeout)
**Goal**: 백그라운드 작업이 타임아웃되면 안전하게 종료

**Given**:
- 백그라운드 작업 실행 중 (타임아웃 설정: 10분)
- 작업이 11분째 실행 중

**When**:
- 타임아웃(10분) 초과 감지
- Background Bash Handler가 작업 종료

**Then**:
- ✅ 작업이 자동으로 종료됨
- ✅ 로그에 "Background task bg-a1b2c3d4 timed out after 10 minutes" 메시지 출력
- ✅ 사용자에게 알림: "Task timed out. Partial results saved to .moai/logs/..."
- ✅ 부분 결과가 로그 파일에 저장됨

**Quality Gate**:
- Timeout handling graceful (no crashes)
- Partial results preserved

---

### Scenario 8: Enhanced Grep - Multiline Pattern Matching

**Feature**: Enhanced Grep Tool
**Goal**: 여러 줄에 걸친 패턴 매칭 (multiline=true)

**Given**:
- tag-agent가 복잡한 정규식으로 코드 블록 검색
- 패턴: `@SPEC:[A-Z-]+\s*\n.*Requirements`

**When**:
- tag-agent가 `multiline=true` 옵션으로 Grep 실행
- 여러 줄에 걸친 SPEC 블록 검색

**Then**:
- ✅ Grep이 multiline 모드에서 패턴 매칭 수행
- ✅ 매칭된 전체 블록 반환 (SPEC 헤더 + Requirements 섹션)
- ✅ 결과가 정확히 패턴에 매칭된 블록만 포함
- ✅ 로그에 "Multiline mode enabled for pattern matching" 메시지 출력

**Quality Gate**:
- Multiline pattern matching 100% accurate
- No false positives/negatives

---

### Scenario 9: Enhanced Grep - Head Limit Parameter

**Feature**: Enhanced Grep Tool
**Goal**: head_limit 파라미터로 결과 개수 제한

**Given**:
- tag-agent가 `@SPEC:` 패턴으로 전체 프로젝트 검색
- 예상 매칭 결과: 50개 SPEC 문서

**When**:
- tag-agent가 `head_limit=10` 옵션으로 Grep 실행
- 처음 10개 결과만 필요

**Then**:
- ✅ Grep이 정확히 10개 결과만 반환
- ✅ 결과가 수정 시간 기준 최신순 정렬
- ✅ 로그에 "Returning first 10 results (head_limit=10)" 메시지 출력
- ✅ 성능: 대용량 프로젝트에서도 빠른 응답 (<1초)

**Quality Gate**:
- Result count exactly matches head_limit
- Performance: <1 second for large codebases

---

### Scenario 10: Plan Resume - Restore Previous Plan

**Feature**: Plan Resume
**Goal**: 이전 계획 상태를 복원하고 수정

**Given**:
- 사용자가 이전 세션에서 `/alfred:1-plan` 실행 완료
- 계획 상태가 `.moai/memory/plan-history.json`에 저장됨

**When**:
- 사용자가 `--resume-plan` 옵션으로 `/alfred:1-plan` 실행
- spec-builder가 마지막 계획 상태 로드

**Then**:
- ✅ 마지막 저장된 계획이 로드됨
- ✅ 사용자에게 이전 계획 내용 표시
- ✅ AskUserQuestion으로 수정 여부 확인
- ✅ 수정 완료 후 TodoWrite이 자동으로 업데이트됨
- ✅ 변경 이력이 `.moai/memory/plan-history.json`에 추가됨

**Quality Gate**:
- Plan restoration 100% accurate
- TodoWrite updates reflect plan modifications

---

### Scenario 11: TodoWrite Auto-Initialization - From Plan Results

**Feature**: TodoWrite Auto-Initialization
**Goal**: Plan Agent 결과에서 TodoWrite 자동 초기화

**Given**:
- spec-builder가 SPEC 분석 완료 (Sonnet으로 계획 수립)
- 계획 결과에 5개 작업 항목 포함

**When**:
- spec-builder가 계획 완료 후 TodoWrite Auto-Init 호출
- TodoWrite Initializer가 계획 결과 파싱

**Then**:
- ✅ TodoWrite이 자동으로 5개 작업 항목 생성
- ✅ 모든 작업이 "pending" 상태로 시작
- ✅ 첫 번째 작업만 "in_progress"로 설정
- ✅ 각 작업은 `content`, `activeForm`, `status` 필드 포함
- ✅ 로그에 "TodoWrite auto-initialized with 5 tasks from Plan results" 메시지 출력

**Quality Gate**:
- All tasks correctly initialized
- Task priorities and dependencies set accurately

---

### Scenario 12: Integration Test - Full Workflow with All Features

**Feature**: All 6 Features Integration
**Goal**: 전체 워크플로우에서 모든 기능이 함께 작동하는지 검증

**Given**:
- MoAI-ADK v0.9.0 설치 완료
- 모든 6개 기능이 `.moai/config.json`에서 활성화됨
- 사용자가 새 SPEC 작업 시작

**When**:
1. `/alfred:1-plan "User profile management"` 실행
   - spec-builder (Sonnet) → 계획 수립
   - TodoWrite Auto-Init → 작업 목록 생성
2. `/alfred:2-run SPEC-PROFILE-001` 실행
   - tdd-implementer (Haiku) → 코드 구현
   - Background Bash → pytest 백그라운드 실행
   - Enhanced Grep → @TAG 검색
3. 토큰 사용률 80% 도달
   - PreCompact Hook → 상태 자동 저장
4. 새 세션에서 `--resume-plan` 실행
   - Plan Resume → 이전 계획 복원

**Then**:
- ✅ **Feature 1**: spec-builder는 Sonnet, tdd-implementer는 Haiku 사용
- ✅ **Feature 2**: 토큰 80% 도달 시 상태 자동 저장됨
- ✅ **Feature 3**: pytest가 백그라운드에서 실행되고 완료 알림 수신
- ✅ **Feature 4**: Enhanced Grep이 @TAG 패턴 정확히 검색
- ✅ **Feature 5**: 새 세션에서 이전 계획 성공적으로 복원
- ✅ **Feature 6**: TodoWrite이 Plan 결과에서 자동 초기화됨

**Quality Gate**:
- ✅ Integration tests 통과율 ≥ 95%
- ✅ 비용 절감 70-90% 달성
- ✅ 성능 4-6배 향상 달성
- ✅ 모든 기능이 충돌 없이 작동

---

## Definition of Done

### Functional Requirements
- ✅ 모든 12개 테스트 시나리오 통과
- ✅ 각 기능이 `.moai/config.json`에서 활성화/비활성화 가능
- ✅ Backward compatibility: v0.8.0 프로젝트가 v0.9.0으로 마이그레이션 성공

### Non-Functional Requirements
- ✅ **Performance**: 성능 4-6배 향상 (Haiku mode)
- ✅ **Cost**: 비용 70-90% 절감 (Haiku mode)
- ✅ **Reliability**: 에러 발생 시 graceful degradation
- ✅ **Usability**: 모든 기능에 대한 사용자 가이드 제공

### Quality Gates
1. **TRUST 5 Compliance**
   - **Test First**: 모든 코드가 테스트 커버리지 90% 이상
   - **Readable**: Ruff 100% 통과
   - **Unified**: 일관된 코딩 스타일
   - **Secured**: 보안 취약점 0개
   - **Trackable**: @TAG 체인 100% 연결

2. **Test Coverage**
   - Unit tests: 90% 이상
   - Integration tests: 95% 통과율
   - E2E tests: 모든 시나리오 통과

3. **Code Quality**
   - Linting (Ruff): 0 errors
   - Type Checking (Pyright): strict mode 통과
   - Security (Bandit): 0 high/medium vulnerabilities

---

## Verification Methods

### Automated Testing
- **Unit Tests**: `tests/unit/test_*` (pytest)
- **Integration Tests**: `tests/integration/test_claude_code_features.py`
- **E2E Tests**: `tests/e2e/test_full_workflow.py`

### Manual Verification
- **Cost Measurement**: Claude API dashboard에서 비용 비교
- **Performance Benchmarking**: 실행 시간 측정 (Sonnet vs Haiku)
- **User Experience**: 실제 프로젝트에서 워크플로우 테스트

### Acceptance Sign-off
- ✅ **Developer**: 모든 테스트 통과 확인
- ✅ **Product Owner**: 비용 절감 및 성능 향상 목표 달성 확인
- ✅ **QA**: 품질 게이트 모두 통과 확인

---

## Test Data & Fixtures

### Sample SPEC for Testing
```yaml
# .moai/specs/SPEC-TEST-001/spec.md
---
id: TEST-001
version: 0.0.1
status: draft
---

## @SPEC:TEST-001
Sample SPEC for testing Claude Code features
```

### Mock Configuration
```json
{
  "claude_code": {
    "features": {
      "auto_sonnet_plan": {"enabled": true},
      "precompact_hook": {"enabled": true, "token_threshold": 0.8},
      "background_bash": {"enabled": true, "timeout_ms": 600000},
      "enhanced_grep": {"enabled": true},
      "plan_resume": {"enabled": true},
      "todowrite_auto_init": {"enabled": true}
    }
  }
}
```

---

**Acceptance Criteria Version**: 0.0.1
**Last Updated**: 2025-11-02
**Next Review**: Upon implementation completion
