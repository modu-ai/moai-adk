# SPEC-CMD-001: Implementation Plan

## Metadata

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-CMD-001 |
| Title | /alfred Command - One-Shot Automation |
| Status | Planned |
| Priority | LOW |
| Depends On | SPEC-WEB-001, SPEC-MODEL-001 |

---

## 1. Implementation Overview (구현 개요)

### 1.1 목표

`/alfred` 명령어를 통해 Plan → Run → Sync 워크플로우를 단일 명령으로 자동화하여 개발자 생산성을 극대화한다.

### 1.2 핵심 가치

- **One-Shot Execution**: 단일 명령으로 전체 개발 워크플로우 실행
- **Flexible Git Workflow**: 다양한 Git 전략 지원 (worktree, branch, direct)
- **Parallel Development**: Worktree 기반 병렬 작업으로 처리량 증가
- **Minimal Checkpoints**: SPEC 승인만 필수, 나머지는 자동화

### 1.3 의존성 분석

```
SPEC-CMD-001 (this)
├── SPEC-WEB-001 (Web Dashboard)
│   ├── WebSocket 실시간 모니터링
│   └── 워크플로우 진행 상황 UI
└── SPEC-MODEL-001 (Multi-Model Support)
    ├── Opus 모델 (Planning)
    └── GLM 모델 (Implementation)
```

---

## 2. Milestones (마일스톤)

### Milestone 1: Core Command Infrastructure (Primary Goal)

**목표**: 기본 명령어 구조 및 워크플로우 엔진 구현

**구현 항목**:
- [ ] Slash command definition (`alfred.md`)
- [ ] CLI command implementation (`alfred.py`)
- [ ] Workflow orchestration service (`workflow_service.py`)
- [ ] Configuration loading and validation

**완료 기준**:
- `/alfred "feature"` 명령이 순차적 워크플로우 실행
- 단일 기능에 대해 Plan → Run → Sync 완료

**품질 게이트**:
- 단위 테스트 커버리지 85% 이상
- E2E 테스트 1개 이상 통과

---

### Milestone 2: Git Worktree Integration (Secondary Goal)

**목표**: Worktree 기반 격리 환경 지원

**구현 항목**:
- [ ] WorktreeManager 통합
- [ ] `--worktree` 옵션 구현
- [ ] Worktree 생성/동기화/정리 로직
- [ ] 병렬 worktree 관리

**완료 기준**:
- `--worktree` 옵션 시 격리된 환경에서 작업
- Worktree 동기화 후 메인 브랜치 병합 가능

**품질 게이트**:
- Worktree 생성/삭제 테스트 통과
- 동시성 테스트 통과

---

### Milestone 3: Parallel Execution (Secondary Goal)

**목표**: 병렬 워커를 통한 동시 구현

**구현 항목**:
- [ ] `--parallel N` 옵션 구현
- [ ] asyncio 기반 병렬 실행 엔진
- [ ] 워커별 진행 상황 추적
- [ ] 실패 워커 처리 및 복구

**완료 기준**:
- `--parallel 3` 시 3개 프로세스 동시 실행
- 하나의 워커 실패 시 나머지 계속 진행

**품질 게이트**:
- 병렬 실행 테스트 통과
- 리소스 경쟁 테스트 통과

---

### Milestone 4: WebSocket Monitoring (Final Goal)

**목표**: 실시간 진행 상황 모니터링 연동

**구현 항목**:
- [ ] WebSocket 이벤트 발행
- [ ] SPEC-WEB-001 Dashboard 연동
- [ ] 진행률 및 로그 스트리밍

**완료 기준**:
- 웹 대시보드에서 실시간 진행 상황 확인
- 각 Phase/SPEC별 상태 표시

**품질 게이트**:
- WebSocket 연결 안정성 테스트
- UI 통합 테스트

**의존성**: SPEC-WEB-001 완료 필수

---

### Milestone 5: PR & Auto-Merge (Optional Goal)

**목표**: GitHub PR 자동 생성 및 Team 모드 병합

**구현 항목**:
- [ ] `gh pr create` 통합
- [ ] `--no-pr` 옵션 구현
- [ ] `--auto-merge` 옵션 구현 (Team 모드)
- [ ] PR 템플릿 적용

**완료 기준**:
- PR 자동 생성 및 URL 출력
- Team 모드에서 승인 후 자동 병합

**품질 게이트**:
- PR 생성 테스트 (mock)
- 모드별 동작 테스트

---

## 3. Technical Approach (기술적 접근)

### 3.1 아키텍처 설계

```
┌─────────────────────────────────────────────────────────────┐
│                    /alfred Command                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   WorkflowOrchestrator                      │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │  Phase 0    │  Phase 1    │  Phase 2    │  Phase 3    │  │
│  │  Config     │  Planning   │  Implement  │  Sync       │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
└─────────────────────────────────────────────────────────────┘
         │                │              │              │
         ▼                ▼              ▼              ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ ConfigLoader│  │ SPECBuilder │  │  TDDRunner  │  │ DocsSyncer  │
└─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘
                        │              │
                        ▼              ▼
              ┌─────────────┐  ┌─────────────┐
              │ Checkpoint  │  │  Worktree   │
              │  Manager    │  │  Manager    │
              └─────────────┘  └─────────────┘
                        │              │
                        └──────┬───────┘
                               ▼
                    ┌─────────────────┐
                    │    WebSocket    │
                    │    Publisher    │
                    └─────────────────┘
```

### 3.2 핵심 컴포넌트

#### WorkflowOrchestrator

```python
class WorkflowOrchestrator:
    """워크플로우 전체 조율 담당"""

    def __init__(
        self,
        config_loader: ConfigLoader,
        spec_builder: SPECBuilder,
        tdd_runner: TDDRunner,
        docs_syncer: DocsSyncer,
        checkpoint_manager: CheckpointManager,
        worktree_manager: WorktreeManager,
        ws_publisher: WebSocketPublisher,
    ):
        pass

    async def execute(self, config: WorkflowConfig) -> WorkflowReport:
        """메인 실행 루프"""
        pass
```

#### CheckpointManager

```python
class CheckpointManager:
    """사용자 확인 체크포인트 관리"""

    async def require_approval(
        self,
        checkpoint_type: str,
        data: dict,
    ) -> bool:
        """AskUserQuestion을 통한 승인 요청"""
        pass

    async def save_state(self, workflow_id: str, state: dict) -> None:
        """상태 저장 (재개용)"""
        pass
```

#### ParallelExecutor

```python
class ParallelExecutor:
    """병렬 워커 실행 관리"""

    async def execute_parallel(
        self,
        specs: list[str],
        workers: int,
        model: str,
    ) -> list[SPECResult]:
        """asyncio.gather를 통한 병렬 실행"""
        pass
```

### 3.3 설정 스키마

```yaml
# .moai/config/sections/alfred.yaml
alfred:
  default_parallel: 1
  default_model: "glm"
  auto_cleanup_worktrees: true
  checkpoint_timeout_seconds: 300

  websocket:
    enabled: true
    port: 8765

  pr_template: |
    ## Summary
    Auto-generated by /alfred command

    ## SPECs Implemented
    {{specs}}

    ## Test Results
    {{test_results}}
```

### 3.4 오류 처리 전략

```python
class WorkflowErrorHandler:
    """계층적 오류 처리"""

    async def handle_phase_error(
        self,
        phase: WorkflowPhase,
        error: Exception,
    ) -> ErrorRecoveryAction:
        """
        Phase별 오류 처리 전략:
        - CONFIGURATION: 즉시 중단, 설정 수정 안내
        - PLANNING: 부분 SPEC 저장, 재개 가능
        - IMPLEMENTATION: 실패 SPEC 건너뛰기 옵션
        - SYNC: 수동 동기화 안내
        """
        pass
```

---

## 4. Risk Management (리스크 관리)

### 4.1 기술적 리스크

| 리스크 | 확률 | 영향 | 완화 전략 |
|--------|------|------|-----------|
| Worktree 동시성 충돌 | Medium | High | 락 메커니즘 구현 |
| Claude Code CLI 호환성 변경 | Low | High | 버전 고정 및 호환성 테스트 |
| WebSocket 연결 불안정 | Medium | Medium | 재연결 로직 및 폴백 |
| 병렬 실행 리소스 고갈 | Low | Medium | 워커 수 제한 (max 5) |

### 4.2 의존성 리스크

| 의존성 | 리스크 | 완화 전략 |
|--------|--------|-----------|
| SPEC-WEB-001 지연 | WebSocket 모니터링 불가 | CLI 기반 진행률 표시로 폴백 |
| SPEC-MODEL-001 지연 | GLM 모델 미지원 | 기본 모델(Sonnet)로 폴백 |

---

## 5. Implementation Order (구현 순서)

```
Phase 1: Foundation (Milestone 1)
├── 1.1 Slash command definition
├── 1.2 CLI entry point
├── 1.3 WorkflowOrchestrator skeleton
└── 1.4 Basic sequential execution

Phase 2: Git Integration (Milestone 2)
├── 2.1 WorktreeManager integration
├── 2.2 Branch creation logic
└── 2.3 Worktree sync logic

Phase 3: Parallel (Milestone 3)
├── 3.1 ParallelExecutor implementation
├── 3.2 Progress tracking
└── 3.3 Error aggregation

Phase 4: Monitoring (Milestone 4) [Depends on SPEC-WEB-001]
├── 4.1 WebSocket event publisher
├── 4.2 Dashboard integration
└── 4.3 Real-time logging

Phase 5: PR Automation (Milestone 5)
├── 5.1 GitHub CLI integration
├── 5.2 PR template system
└── 5.3 Auto-merge flow
```

---

## 6. Testing Strategy (테스트 전략)

### 6.1 단위 테스트

| 컴포넌트 | 테스트 범위 | 우선순위 |
|----------|-------------|----------|
| WorkflowOrchestrator | Phase 전환 로직 | High |
| CheckpointManager | 승인 요청/상태 저장 | High |
| ParallelExecutor | 병렬 실행/오류 처리 | High |
| ConfigLoader | 설정 로드/검증 | Medium |

### 6.2 통합 테스트

| 시나리오 | 테스트 범위 | 우선순위 |
|----------|-------------|----------|
| 단일 기능 워크플로우 | Plan → Run → Sync | High |
| Worktree 기능 워크플로우 | 격리 환경 생성/동기화 | High |
| 병렬 실행 워크플로우 | 동시 3개 SPEC 처리 | Medium |
| 오류 복구 시나리오 | Phase별 실패 처리 | Medium |

### 6.3 E2E 테스트

| 시나리오 | 테스트 범위 | 의존성 |
|----------|-------------|--------|
| 전체 워크플로우 | CLI → SPEC → Code → PR | 모든 의존성 |
| 웹 모니터링 | CLI → WebSocket → UI | SPEC-WEB-001 |

---

## 7. Documentation Plan (문서화 계획)

### 7.1 사용자 문서

- [ ] `/alfred` 명령어 가이드
- [ ] 옵션별 사용 예시
- [ ] 트러블슈팅 가이드

### 7.2 개발자 문서

- [ ] 아키텍처 다이어그램
- [ ] API 레퍼런스
- [ ] 확장 가이드

---

## 8. Traceability (추적성)

| Plan Item | SPEC Requirement | Acceptance Criteria |
|-----------|------------------|---------------------|
| M1: Core Command | E1, U1, U2 | AC-001, AC-002 |
| M2: Worktree | S3, E3 | AC-003, AC-004 |
| M3: Parallel | E5, N4 | AC-005, AC-006 |
| M4: WebSocket | E6 | AC-007 |
| M5: PR | E4, S4, S5, N1 | AC-008, AC-009, AC-010 |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-09 | manager-spec | Initial plan creation |
