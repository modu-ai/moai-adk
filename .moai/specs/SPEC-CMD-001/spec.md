# SPEC-CMD-001: /alfred Command - One-Shot Automation

## Metadata

| 항목 | 값 |
|------|-----|
| SPEC ID | SPEC-CMD-001 |
| Title | /alfred Command - One-Shot Automation |
| Created | 2026-01-09 |
| Status | Planned |
| Priority | LOW |
| Lifecycle | spec-anchored |
| Depends On | SPEC-WEB-001, SPEC-MODEL-001 |
| Assigned | manager-spec |

---

## 1. Environment (환경)

### 1.1 시스템 컨텍스트

- **실행 환경**: MoAI-ADK CLI 및 Claude Code 통합 환경
- **대상 사용자**: MoAI-ADK를 사용하는 개발자
- **모드 지원**: Personal 모드 및 Team 모드
- **필수 의존성**:
  - SPEC-WEB-001: Web Dashboard (워크플로우 모니터링 UI)
  - SPEC-MODEL-001: Multi-Model 지원 (GLM 구현 모델)

### 1.2 기술 스택

```yaml
Integration:
  - MoAI-ADK CLI infrastructure (moai_adk/cli/)
  - claude-agent-sdk workflow API
  - WorktreeManager (moai_adk/worktree/)

Runtime:
  - Python 3.13+
  - asyncio for parallel execution
  - WebSocket for real-time monitoring

External Dependencies:
  - Git (worktree support)
  - GitHub CLI (gh) for PR creation
  - Claude Code CLI
```

### 1.3 파일 구조

```
.claude/commands/moai/
└── alfred.md           # Slash command definition

src/moai_adk/cli/commands/
└── alfred.py           # CLI implementation

src/moai_adk/web/services/
└── workflow_service.py      # Workflow orchestration service
```

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

| ID | 가정 | 신뢰도 | 근거 | 실패 시 위험 | 검증 방법 |
|----|------|--------|------|-------------|-----------|
| A1 | Git worktree 기능이 모든 대상 환경에서 사용 가능함 | High | Git 2.5+ 표준 기능 | worktree 기반 병렬 작업 불가 | git worktree list 실행 |
| A2 | Claude Code CLI가 설치되어 있고 인증됨 | High | MoAI-ADK 필수 의존성 | 구현 단계 실행 불가 | claude --version 확인 |
| A3 | 복수 터미널/프로세스 동시 실행 가능 | Medium | OS 표준 기능 | 병렬 실행 직렬화 필요 | subprocess 테스트 |
| A4 | WebSocket 연결이 안정적으로 유지됨 | Medium | SPEC-WEB-001 구현 | 모니터링 기능 제한 | 연결 테스트 |

### 2.2 비즈니스 가정

| ID | 가정 | 신뢰도 | 근거 | 실패 시 위험 |
|----|------|--------|------|-------------|
| B1 | 사용자는 여러 기능을 한 번에 구현하고자 함 | High | 워크플로우 효율성 요구 | 명령어 활용도 저하 |
| B2 | SPEC 승인은 반드시 사용자 확인 필요 | High | 품질 게이트 정책 | 의도치 않은 구현 진행 |
| B3 | PR 자동 생성/병합은 Team 모드에서만 의미 있음 | Medium | Personal 모드는 단일 개발자 | 모드별 기능 분리 필요 |

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous Requirements (시스템 전반)

| ID | 요구사항 | 검증 방법 |
|----|----------|-----------|
| U1 | 시스템은 **항상** 모든 작업 진행 상황을 로깅해야 한다 | 로그 파일 검사 |
| U2 | 시스템은 **항상** 실패한 단계의 상세 오류 정보를 제공해야 한다 | 오류 메시지 검증 |
| U3 | 시스템은 **항상** 작업 완료 시 비용 분석 보고서를 생성해야 한다 | 보고서 출력 확인 |

### 3.2 Event-Driven Requirements (이벤트 기반)

| ID | 이벤트 | 응답 | 검증 방법 |
|----|--------|------|-----------|
| E1 | **WHEN** `/alfred "desc1, desc2"` 실행 **THEN** 순차적 워크플로우 시작 | 워크플로우 시작 확인 |
| E2 | **WHEN** SPEC 생성 완료 **THEN** 사용자 승인 요청 (CHECKPOINT) | 승인 프롬프트 표시 |
| E3 | **WHEN** 사용자가 SPEC 승인 **THEN** worktree 생성 및 구현 시작 | worktree 생성 확인 |
| E4 | **WHEN** 모든 구현 완료 **THEN** 자동 동기화 및 PR 생성 | PR URL 출력 |
| E5 | **WHEN** `--parallel N` 옵션 지정 **THEN** N개의 병렬 워커 생성 | 프로세스 수 확인 |
| E6 | **WHEN** WebSocket 연결 **THEN** 실시간 진행 상황 전송 | 메시지 수신 확인 |

### 3.3 State-Driven Requirements (상태 기반)

| ID | 조건 | 동작 | 검증 방법 |
|----|------|------|-----------|
| S1 | **IF** 이전 단계 실패 **THEN** 다음 단계 진행 안 함 | 단계 중단 확인 |
| S2 | **IF** `--no-branch` 옵션 **THEN** 현재 브랜치에서 작업 | 브랜치 미생성 확인 |
| S3 | **IF** `--worktree` 옵션 **THEN** 격리된 worktree 생성 | worktree 존재 확인 |
| S4 | **IF** `--no-pr` 옵션 **THEN** PR 생성 생략 | PR 미생성 확인 |
| S5 | **IF** `--auto-merge` 옵션 (Team 모드) **THEN** 승인 후 자동 병합 | 병합 상태 확인 |
| S6 | **IF** `--model` 옵션 **THEN** 지정된 모델로 구현 | 모델 사용 로그 확인 |

### 3.4 Unwanted Requirements (금지 사항)

| ID | 금지 동작 | 이유 | 검증 방법 |
|----|-----------|------|-----------|
| N1 | 시스템은 사용자 확인 없이 원격 저장소에 **푸시하지 않아야 한다** | 의도치 않은 코드 배포 방지 | 푸시 로그 검사 |
| N2 | 시스템은 SPEC 승인 없이 구현을 **시작하지 않아야 한다** | 품질 게이트 준수 | 워크플로우 로그 검사 |
| N3 | 시스템은 실패한 worktree를 자동으로 **삭제하지 않아야 한다** | 디버깅 정보 보존 | worktree 상태 확인 |
| N4 | 시스템은 병렬 작업 중 공유 리소스를 **직접 수정하지 않아야 한다** | 경쟁 조건 방지 | 동시성 테스트 |

### 3.5 Optional Requirements (선택적)

| ID | 조건 | 기능 | 우선순위 |
|----|------|------|----------|
| O1 | **가능하면** 이전 세션 상태 복원 기능 제공 | 중단된 워크플로우 재개 | Low |
| O2 | **가능하면** Slack/Discord 알림 연동 제공 | 완료 알림 | Low |
| O3 | **가능하면** 비용 예측 기능 제공 | 실행 전 비용 추정 | Medium |

---

## 4. Specifications (상세 명세)

### 4.1 Command Interface

```markdown
# /alfred Command

## Usage
/alfred "<feature1>, <feature2>, ..." [options]

## Options
--worktree          격리된 worktree 환경에서 작업
--parallel N        N개의 병렬 워커 실행 (기본값: 1)
--no-branch         현재 브랜치에서 작업 (브랜치 생성 안 함)
--no-pr             PR 생성 생략
--auto-merge        Team 모드에서 자동 병합 (사용자 확인 후)
--model MODEL       구현에 사용할 모델 지정 (기본값: GLM)

## Examples
/alfred "user authentication, payment processing"
/alfred "auth, payment, dashboard" --worktree --parallel 3
/alfred "bug fix #123" --no-branch --no-pr
```

### 4.2 Execution Flow

```
/alfred "auth, payment, dashboard" --worktree --parallel 3

PHASE 0: Configuration
├── Load .moai/config/sections/git-strategy.yaml
├── Parse command flags
├── Validate dependencies (SPEC-WEB-001, SPEC-MODEL-001)
└── Determine execution mode (Personal/Team)

PHASE 1: Planning (Opus Model)
├── Analyze requirements ("auth", "payment", "dashboard")
├── Create SPEC documents
│   ├── SPEC-AUTH-XXX
│   ├── SPEC-PAYMENT-XXX
│   └── SPEC-DASHBOARD-XXX
├── [CHECKPOINT] User reviews and approves SPECs
│   └── AskUserQuestion: "다음 SPEC을 승인하시겠습니까?"
└── Create worktrees with GLM configuration
    ├── project-worktrees/SPEC-AUTH-XXX/
    ├── project-worktrees/SPEC-PAYMENT-XXX/
    └── project-worktrees/SPEC-DASHBOARD-XXX/

PHASE 2: Parallel Implementation (GLM Model)
├── Open parallel terminals/processes
├── Each worker executes:
│   └── claude /moai:2-run SPEC-{ID} --model glm
├── Monitor progress via WebSocket (SPEC-WEB-001)
│   └── Real-time status updates to Dashboard
└── Wait for all completions (asyncio.gather)

PHASE 3: Sync & PR
├── moai-adk worktree sync --all
├── For each SPEC:
│   └── /moai:3-sync SPEC-{ID}
├── Create PRs (if enabled)
│   └── gh pr create --title "SPEC-{ID}: {title}"
└── [CHECKPOINT] Merge decision (Team mode only)
    └── AskUserQuestion: "PR을 병합하시겠습니까?"

PHASE 4: Completion
├── Generate summary report
│   ├── Completed SPECs
│   ├── Implementation time per SPEC
│   └── Total token usage
├── Cost analysis
│   ├── Opus tokens (Planning)
│   └── GLM tokens (Implementation)
└── Cleanup merged worktrees (optional)
```

### 4.3 Data Models

```python
from dataclasses import dataclass
from enum import Enum
from typing import Optional
from datetime import datetime

class WorkflowPhase(Enum):
    CONFIGURATION = "configuration"
    PLANNING = "planning"
    IMPLEMENTATION = "implementation"
    SYNC = "sync"
    COMPLETION = "completion"

class WorkflowStatus(Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    CHECKPOINT = "checkpoint"
    COMPLETED = "completed"
    FAILED = "failed"

@dataclass
class WorkflowConfig:
    features: list[str]
    use_worktree: bool = False
    parallel_workers: int = 1
    create_branch: bool = True
    create_pr: bool = True
    auto_merge: bool = False
    model: str = "glm"
    mode: str = "personal"  # personal | team

@dataclass
class SPECResult:
    spec_id: str
    title: str
    status: WorkflowStatus
    worktree_path: Optional[str]
    pr_url: Optional[str]
    tokens_used: int
    duration_seconds: float

@dataclass
class WorkflowReport:
    workflow_id: str
    started_at: datetime
    completed_at: Optional[datetime]
    config: WorkflowConfig
    specs: list[SPECResult]
    total_tokens: int
    total_cost_usd: float
    phase: WorkflowPhase
    status: WorkflowStatus
```

### 4.4 API Contracts

```python
# CLI Entry Point
async def alfred_command(
    features: str,
    worktree: bool = False,
    parallel: int = 1,
    no_branch: bool = False,
    no_pr: bool = False,
    auto_merge: bool = False,
    model: str = "glm",
) -> WorkflowReport:
    """
    Execute one-shot automation workflow.

    Args:
        features: Comma-separated feature descriptions
        worktree: Use isolated worktree environment
        parallel: Number of parallel workers
        no_branch: Work on current branch
        no_pr: Skip PR creation
        auto_merge: Auto-merge after approval (Team mode)
        model: Model for implementation

    Returns:
        WorkflowReport with execution results

    Raises:
        DependencyError: When SPEC-WEB-001 or SPEC-MODEL-001 unavailable
        CheckpointAbortError: When user rejects at checkpoint
        WorktreeError: When worktree operations fail
    """
    pass

# WebSocket Events
class WorkflowEvents:
    PHASE_START = "workflow:phase:start"
    PHASE_COMPLETE = "workflow:phase:complete"
    SPEC_START = "workflow:spec:start"
    SPEC_PROGRESS = "workflow:spec:progress"
    SPEC_COMPLETE = "workflow:spec:complete"
    CHECKPOINT = "workflow:checkpoint"
    ERROR = "workflow:error"
```

### 4.5 Error Handling

| 오류 유형 | 원인 | 복구 전략 |
|----------|------|-----------|
| `DependencyError` | SPEC-WEB-001 또는 SPEC-MODEL-001 미완료 | 의존성 SPEC 완료 안내 |
| `CheckpointAbortError` | 사용자가 체크포인트에서 중단 | 현재 상태 저장 및 재개 가능 안내 |
| `WorktreeError` | Git worktree 생성/관리 실패 | 수동 worktree 정리 안내 |
| `ParallelExecutionError` | 병렬 워커 중 하나 이상 실패 | 실패 워커 로그 제공, 재시도 옵션 |
| `PRCreationError` | GitHub PR 생성 실패 | 수동 PR 생성 안내 |

---

## 5. Traceability (추적성)

### 5.1 요구사항-테스트 매핑

| Requirement ID | Test Scenario ID | acceptance.md Reference |
|----------------|------------------|-------------------------|
| E1 | TC-001 | Given-When-Then #1 |
| E2 | TC-002 | Given-When-Then #2 |
| E3 | TC-003 | Given-When-Then #3 |
| S1 | TC-004 | Given-When-Then #4 |
| S3 | TC-005 | Given-When-Then #5 |
| N1 | TC-006 | Given-When-Then #6 |

### 5.2 관련 문서

- `plan.md`: 구현 계획 및 마일스톤
- `acceptance.md`: 인수 테스트 시나리오
- SPEC-WEB-001: Web Dashboard (모니터링 UI)
- SPEC-MODEL-001: Multi-Model 지원 (GLM)

---

## 6. Constitution Alignment (프로젝트 헌법 정렬)

### 6.1 기술 스택 정렬

| Constitution 요구사항 | SPEC 정렬 상태 |
|----------------------|----------------|
| Python 3.13+ | Compliant |
| asyncio for concurrency | Compliant |
| Click for CLI | Compliant |
| WebSocket via FastAPI | Compliant (SPEC-WEB-001) |

### 6.2 아키텍처 패턴 정렬

| Constitution 요구사항 | SPEC 정렬 상태 |
|----------------------|----------------|
| Single Responsibility | Compliant (Phase 분리) |
| Dependency Injection | Compliant (Service layer) |
| Event-Driven | Compliant (WebSocket events) |

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-09 | manager-spec | Initial SPEC creation |
