# SPEC-CMD-001: Acceptance Criteria

## Metadata

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-CMD-001 |
| Title | /alfred Command - One-Shot Automation |
| Status | Planned |
| Priority | LOW |
| Depends On | SPEC-WEB-001, SPEC-MODEL-001 |

---

## 1. Quality Gate Criteria (품질 게이트 기준)

### 1.1 코드 품질

| 기준 | 목표값 | 검증 방법 |
|------|--------|-----------|
| 테스트 커버리지 | >= 85% | pytest --cov |
| 린트 경고 | 0 | ruff check |
| 타입 힌트 커버리지 | 100% | mypy --strict |
| 문서 커버리지 | >= 90% | interrogate |

### 1.2 성능 기준

| 기준 | 목표값 | 검증 방법 |
|------|--------|-----------|
| 명령어 시작 시간 | < 2초 | 타이머 측정 |
| SPEC 생성 시간 | < 30초/SPEC | 타이머 측정 |
| Worktree 생성 시간 | < 5초/worktree | 타이머 측정 |
| WebSocket 메시지 지연 | < 100ms | 레이턴시 측정 |

### 1.3 보안 기준

| 기준 | 검증 방법 |
|------|-----------|
| 원격 푸시 전 사용자 확인 | 테스트 케이스 |
| 민감 정보 로깅 금지 | 로그 검사 |
| 권한 검증 | 접근 제어 테스트 |

---

## 2. Acceptance Test Scenarios (인수 테스트 시나리오)

### AC-001: 기본 워크플로우 실행

**Requirement Reference**: E1, U1, U2

```gherkin
Feature: Basic Workflow Execution
  As a developer
  I want to execute the full workflow with a single command
  So that I can automate the Plan-Run-Sync cycle

  Scenario: Execute single feature workflow
    Given MoAI-ADK가 초기화된 프로젝트
    And .moai/config/config.yaml가 유효한 설정을 포함
    When 사용자가 "/alfred 'user authentication'" 명령 실행
    Then 시스템이 SPEC-AUTH-XXX 생성
    And 사용자에게 SPEC 승인 요청 표시
    And 승인 후 TDD 구현 시작
    And 구현 완료 후 문서 동기화
    And 최종 보고서 출력

  Scenario: Execute multiple features sequentially
    Given MoAI-ADK가 초기화된 프로젝트
    When 사용자가 "/alfred 'auth, payment'" 명령 실행
    Then 시스템이 SPEC-AUTH-XXX, SPEC-PAYMENT-XXX 생성
    And 사용자에게 모든 SPEC 승인 요청 표시
    And 승인 후 순차적으로 구현 실행
    And 각 SPEC 완료 시 진행 상황 로깅
```

---

### AC-002: 진행 상황 로깅

**Requirement Reference**: U1, U2, U3

```gherkin
Feature: Progress Logging
  As a developer
  I want to see detailed progress logs
  So that I can monitor the workflow execution

  Scenario: Log all phases
    Given 워크플로우가 실행 중
    When 각 Phase가 시작/완료
    Then 시스템이 Phase 이름과 상태를 로깅
    And 로그에 타임스탬프 포함
    And 실패 시 상세 오류 메시지 포함

  Scenario: Generate cost report
    Given 워크플로우가 완료됨
    Then 시스템이 비용 분석 보고서 생성
    And 보고서에 Opus 토큰 사용량 포함
    And 보고서에 GLM 토큰 사용량 포함
    And 보고서에 총 비용(USD) 포함
```

---

### AC-003: Worktree 격리 환경

**Requirement Reference**: S3, E3

```gherkin
Feature: Worktree Isolation
  As a developer
  I want to use isolated worktree environments
  So that multiple features can be developed without conflicts

  Scenario: Create worktree for each SPEC
    Given MoAI-ADK가 초기화된 프로젝트
    And Git 저장소가 유효함
    When 사용자가 "/alfred 'auth, payment' --worktree" 명령 실행
    And 사용자가 SPEC 승인
    Then 시스템이 project-worktrees/SPEC-AUTH-XXX/ 디렉토리 생성
    And 시스템이 project-worktrees/SPEC-PAYMENT-XXX/ 디렉토리 생성
    And 각 worktree가 독립적인 Git 작업 디렉토리

  Scenario: Sync worktrees after completion
    Given Worktree에서 구현이 완료됨
    When 동기화 단계 실행
    Then 각 worktree의 변경사항이 메인 브랜치로 병합 가능
    And 병합 충돌 시 사용자에게 알림
```

---

### AC-004: Worktree 오류 처리

**Requirement Reference**: N3

```gherkin
Feature: Worktree Error Handling
  As a developer
  I want failed worktrees to be preserved
  So that I can debug issues manually

  Scenario: Preserve failed worktree
    Given Worktree에서 구현 실행 중
    When 구현이 실패함
    Then 시스템이 해당 worktree를 삭제하지 않음
    And 시스템이 실패 로그와 함께 worktree 경로 출력
    And 사용자에게 수동 디버깅 안내

  Scenario: Cleanup successful worktrees only
    Given 여러 worktree 중 일부만 성공
    When 정리 단계 실행
    Then 시스템이 성공한 worktree만 정리 옵션 제공
    And 실패한 worktree는 보존
```

---

### AC-005: 병렬 실행

**Requirement Reference**: E5

```gherkin
Feature: Parallel Execution
  As a developer
  I want to run multiple implementations in parallel
  So that I can reduce total execution time

  Scenario: Execute with parallel workers
    Given MoAI-ADK가 초기화된 프로젝트
    And 3개의 SPEC이 승인됨
    When 사용자가 "--parallel 3" 옵션 지정
    Then 시스템이 3개의 병렬 프로세스 생성
    And 각 프로세스가 독립적으로 SPEC 구현
    And 진행 상황이 개별적으로 추적됨

  Scenario: Limit parallel workers
    Given 5개의 SPEC이 승인됨
    When 사용자가 "--parallel 3" 옵션 지정
    Then 시스템이 최대 3개의 동시 워커 실행
    And 완료된 워커가 대기 중인 SPEC 처리
```

---

### AC-006: 병렬 실행 오류 처리

**Requirement Reference**: N4

```gherkin
Feature: Parallel Execution Error Handling
  As a developer
  I want parallel failures to be isolated
  So that one failure doesn't affect other implementations

  Scenario: Continue on worker failure
    Given 3개의 병렬 워커 실행 중
    When 워커 1이 실패
    Then 워커 2, 3은 계속 실행
    And 시스템이 실패한 워커의 오류 로깅
    And 최종 보고서에 부분 성공 표시

  Scenario: Prevent shared resource conflicts
    Given 병렬 워커들이 실행 중
    When 공유 리소스 접근 시도
    Then 시스템이 락 메커니즘 적용
    And 경쟁 조건 방지
```

---

### AC-007: WebSocket 실시간 모니터링

**Requirement Reference**: E6

**의존성**: SPEC-WEB-001 완료 필수

```gherkin
Feature: WebSocket Real-time Monitoring
  As a developer
  I want to see real-time progress in the web dashboard
  So that I can monitor workflow status visually

  Scenario: Stream progress to dashboard
    Given Web Dashboard가 실행 중 (SPEC-WEB-001)
    And WebSocket 연결이 수립됨
    When 워크플로우가 실행 중
    Then 시스템이 Phase 시작/완료 이벤트 발행
    And 시스템이 SPEC별 진행률 이벤트 발행
    And Dashboard에 실시간 상태 표시

  Scenario: Handle WebSocket disconnection
    Given WebSocket 연결 중
    When 연결이 끊김
    Then 시스템이 재연결 시도 (최대 3회)
    And 재연결 실패 시 CLI 기반 진행률로 폴백
    And 워크플로우 실행은 계속됨
```

---

### AC-008: PR 자동 생성

**Requirement Reference**: E4, S4

```gherkin
Feature: Automatic PR Creation
  As a developer
  I want PRs to be created automatically
  So that I can review changes easily

  Scenario: Create PR for completed SPEC
    Given SPEC 구현이 완료됨
    And GitHub CLI가 인증됨
    When 동기화 단계 실행
    Then 시스템이 SPEC별 PR 생성
    And PR 제목에 SPEC ID 포함
    And PR 본문에 구현 요약 포함
    And PR URL 출력

  Scenario: Skip PR creation with --no-pr
    Given SPEC 구현이 완료됨
    When "--no-pr" 옵션 지정
    Then 시스템이 PR 생성 생략
    And 변경사항은 로컬에만 유지
```

---

### AC-009: 원격 푸시 확인

**Requirement Reference**: N1

```gherkin
Feature: Remote Push Confirmation
  As a developer
  I want to confirm before pushing to remote
  So that I don't accidentally deploy unreviewed code

  Scenario: Require confirmation before push
    Given PR 생성을 위해 푸시 필요
    When 시스템이 원격 저장소 푸시 시도
    Then 시스템이 사용자에게 확인 요청
    And 사용자 승인 없이 푸시 실행 안 함

  Scenario: Allow push after explicit confirmation
    Given 시스템이 푸시 확인 요청
    When 사용자가 승인
    Then 시스템이 원격 저장소에 푸시
    And 푸시 결과 로깅
```

---

### AC-010: Team 모드 자동 병합

**Requirement Reference**: S5

```gherkin
Feature: Team Mode Auto-Merge
  As a team lead
  I want approved PRs to be merged automatically
  So that I can streamline the integration process

  Scenario: Auto-merge in Team mode
    Given Team 모드가 활성화됨
    And "--auto-merge" 옵션 지정
    And PR이 생성됨
    When 사용자가 병합 승인
    Then 시스템이 PR 자동 병합
    And 병합 결과 로깅
    And 병합된 브랜치 정리

  Scenario: Skip auto-merge in Personal mode
    Given Personal 모드가 활성화됨
    When "--auto-merge" 옵션 지정
    Then 시스템이 옵션 무시
    And 경고 메시지 출력 "Auto-merge is only available in Team mode"
```

---

### AC-011: SPEC 승인 체크포인트

**Requirement Reference**: E2, N2

```gherkin
Feature: SPEC Approval Checkpoint
  As a developer
  I want to review and approve SPECs before implementation
  So that I can ensure requirements are correct

  Scenario: Require SPEC approval
    Given SPEC 문서가 생성됨
    When Planning Phase 완료
    Then 시스템이 SPEC 승인 요청 표시
    And 시스템이 SPEC 요약 표시
    And 사용자 응답 대기

  Scenario: Abort on SPEC rejection
    Given SPEC 승인 요청 표시됨
    When 사용자가 거부
    Then 시스템이 워크플로우 중단
    And 현재 상태 저장 (재개 가능)
    And 중단 사유 로깅

  Scenario: Block implementation without approval
    Given SPEC이 생성됨
    When 승인 없이 구현 시작 시도
    Then 시스템이 구현 시작 차단
    And 오류 메시지 "SPEC approval required before implementation"
```

---

### AC-012: 옵션 조합 테스트

**Requirement Reference**: S1, S2, S6

```gherkin
Feature: Option Combinations
  As a developer
  I want various options to work together correctly
  So that I can customize the workflow

  Scenario: --no-branch with --no-pr
    Given 현재 브랜치에서 작업 필요
    When "--no-branch --no-pr" 옵션 지정
    Then 시스템이 새 브랜치 생성 안 함
    And 시스템이 PR 생성 안 함
    And 현재 브랜치에 직접 커밋

  Scenario: --worktree with --parallel
    Given 격리 환경에서 병렬 작업 필요
    When "--worktree --parallel 3" 옵션 지정
    Then 시스템이 3개의 worktree 생성
    And 각 worktree에서 병렬 구현

  Scenario: --model specification
    Given 특정 모델로 구현 필요
    When "--model claude-3-opus" 옵션 지정
    Then 시스템이 지정된 모델로 구현 실행
    And 모델 사용 로깅
```

---

## 3. Definition of Done (완료 정의)

### 3.1 기능 완료 기준

- [ ] 모든 인수 테스트 시나리오 통과
- [ ] 단위 테스트 커버리지 85% 이상
- [ ] 통합 테스트 통과
- [ ] E2E 테스트 통과

### 3.2 문서화 완료 기준

- [ ] 사용자 가이드 작성
- [ ] API 레퍼런스 작성
- [ ] 트러블슈팅 가이드 작성

### 3.3 품질 완료 기준

- [ ] 린트 경고 0개
- [ ] 타입 힌트 100%
- [ ] 보안 검토 통과
- [ ] 코드 리뷰 완료

### 3.4 배포 완료 기준

- [ ] Staging 환경 테스트 통과
- [ ] 성능 기준 충족
- [ ] 롤백 계획 수립

---

## 4. Verification Methods (검증 방법)

### 4.1 자동화 테스트

```bash
# 단위 테스트
pytest tests/unit/cli/test_alfred.py -v --cov=moai_adk.cli.commands.alfred

# 통합 테스트
pytest tests/integration/test_workflow_integration.py -v

# E2E 테스트
pytest tests/e2e/test_alfred_e2e.py -v
```

### 4.2 수동 검증

| 검증 항목 | 방법 | 담당 |
|----------|------|------|
| UI 통합 | Web Dashboard에서 실시간 모니터링 확인 | QA |
| 성능 | 대용량 프로젝트에서 실행 시간 측정 | Performance |
| 보안 | 민감 정보 노출 여부 확인 | Security |

---

## 5. Traceability Matrix (추적 매트릭스)

| AC ID | Requirement IDs | Test File | Status |
|-------|-----------------|-----------|--------|
| AC-001 | E1, U1, U2 | test_basic_workflow.py | Pending |
| AC-002 | U1, U2, U3 | test_logging.py | Pending |
| AC-003 | S3, E3 | test_worktree.py | Pending |
| AC-004 | N3 | test_worktree_error.py | Pending |
| AC-005 | E5 | test_parallel.py | Pending |
| AC-006 | N4 | test_parallel_error.py | Pending |
| AC-007 | E6 | test_websocket.py | Pending |
| AC-008 | E4, S4 | test_pr_creation.py | Pending |
| AC-009 | N1 | test_push_confirmation.py | Pending |
| AC-010 | S5 | test_auto_merge.py | Pending |
| AC-011 | E2, N2 | test_checkpoint.py | Pending |
| AC-012 | S1, S2, S6 | test_options.py | Pending |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-09 | manager-spec | Initial acceptance criteria |
