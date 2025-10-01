---
spec_id: SPEC-005
status: active
priority: medium
dependencies: []
tags:
  - claude-code
  - commands
  - user-experience
  - performance
  - error-handling
  - workflow
---

# SPEC-005: Claude Code 커맨드 최적화

## @SPEC:COMMAND-005 프로젝트 컨텍스트

### 배경

MoAI-ADK의 핵심 가치는 `/moai:8-project → /moai:3-sync` 4단계 파이프라인을 통한 직관적이고 효율적인 개발 워크플로우 제공입니다. 현재 사용자 경험 개선이 필요한 영역으로 명령어 실행 속도, 에러 처리, 진행률 표시 등이 확인되었습니다.

### 문제 정의

- **현재 상태**: 명령어 실행 시 피드백 부족으로 인한 사용자 불안감
- **핵심 문제**: 에러 발생 시 복구 메커니즘 부재, 진행률 표시 미흡
- **비즈니스 영향**: 1차 사용자(개인 개발자)의 채택률 저하 및 이탈 위험

### 목표

1. 4단계 파이프라인의 명확한 진행률 표시 및 상태 피드백 구현
2. 자동 복구 가능한 에러에 대한 복구 메커니즘 제공
3. 명령어 히스토리 및 안전한 롤백 기능 구현

## @SPEC:COMMAND-SYSTEM-005 환경 및 가정사항

### Environment (환경)

- **시스템**: Claude Code 환경 (.claude/commands/moai/ 기반)
- **명령어**: `/moai:8-project`, `/moai:1-spec`, `/moai:2-build`, `/moai:3-sync`
- **확장**: `/moai:git:*` 시리즈 (checkpoint, rollback 등)
- **UI/UX**: 터미널 기반 실시간 피드백

### Assumptions (가정사항)

- Claude Code의 명령어 시스템이 안정적으로 작동함
- 사용자는 터미널 기반 인터페이스에 익숙함
- Git 저장소가 초기화되어 있음 (git 관련 명령어 사용 시)
- 기본적인 Python 개발 환경이 구성되어 있음

## @CODE:IMPLEMENT-005 요구사항 명세

### R1. 실시간 진행률 표시 및 상태 피드백

**WHEN** 4단계 파이프라인 중 임의의 명령어가 실행될 때,
**THE SYSTEM SHALL** 현재 진행 상태와 예상 완료 시간을 실시간으로 표시해야 함

**상세 요구사항:**

- 각 단계별 진행률 바 또는 스피너 표시
- 현재 수행 중인 작업에 대한 명확한 설명
- 예상 완료 시간 또는 경과 시간 표시
- 단계 완료 시 성공/실패 상태 명확히 표시

### R2. 강화된 에러 처리 및 자동 복구

**WHEN** 명령어 실행 중 복구 가능한 에러가 발생할 때,
**THE SYSTEM SHALL** 자동 복구를 시도하고 성공/실패를 사용자에게 알려야 함

**상세 요구사항:**

- 네트워크 연결 실패 시 재시도 메커니즘
- 파일 권한 문제 시 권한 수정 시도
- 종속성 누락 시 자동 설치 시도
- 복구 불가능한 에러 시 명확한 수동 해결 방법 제시

### R3. 명령어 실행 성능 프로파일링 및 최적화

**WHEN** 명령어가 실행될 때,
**THE SYSTEM SHALL** 실행 시간을 측정하고 성능 병목 지점을 식별해야 함

**상세 요구사항:**

- 각 명령어별 실행 시간 측정 및 기록
- 느린 작업에 대한 최적화 제안
- 병렬 처리 가능한 작업 식별 및 적용
- 캐시 활용을 통한 반복 작업 최적화

### R4. 명령어 히스토리 및 롤백 기능

**WHEN** 사용자가 이전 상태로 되돌리기를 원할 때,
**THE SYSTEM SHALL** 안전한 롤백 기능을 제공해야 함

**상세 요구사항:**

- 명령어 실행 히스토리 자동 기록
- Git 기반 체크포인트 자동 생성
- 단계별 롤백 가능 (예: SPEC 단계로 돌아가기)
- 롤백 시 데이터 손실 방지 및 백업 보장

## @TEST:ACCEPTANCE-005 Acceptance Criteria

### AC1. 진행률 표시 및 피드백

**Given** `/moai:8-project` 명령을 실행할 때
**When** 프로젝트 초기화가 진행되면
**Then** 진행률 바가 표시되고, 현재 수행 중인 작업(예: "템플릿 복사 중", "권한 설정 중")이 명확히 표시되어야 함

**Given** 4단계 파이프라인을 순차 실행할 때
**When** 각 단계가 완료되면
**Then** 단계별 성공/실패 상태가 명확한 색상과 메시지로 표시되어야 함

### AC2. 자동 에러 복구

**Given** 네트워크 연결 문제로 패키지 다운로드가 실패할 때
**When** 자동 복구 메커니즘이 작동하면
**Then** 최대 3회까지 재시도하고, 성공 시 사용자에게 복구 완료를 알려야 함

**Given** 파일 권한 문제가 발생할 때
**When** 자동 복구를 시도하면
**Then** 적절한 권한 설정을 시도하고, 성공 여부를 명확히 보고해야 함

### AC3. 성능 최적화

**Given** 명령어 실행 시간이 측정될 때
**When** 성능 프로파일링을 확인하면
**Then** 각 작업별 소요 시간과 전체 대비 비율이 표시되어야 함

**Given** 반복적으로 동일한 명령어를 실행할 때
**When** 캐시 최적화가 적용되면
**Then** 두 번째 실행부터는 최소 30% 이상 성능 개선이 있어야 함

### AC4. 히스토리 및 롤백

**Given** 여러 명령어를 순차적으로 실행한 후
**When** `/moai:git:rollback "SPEC 단계"` 명령을 실행하면
**Then** 안전하게 SPEC 작성 완료 시점으로 되돌아가고, 변경사항은 별도 브랜치에 백업되어야 함

**Given** 명령어 실행 히스토리를 조회할 때
**When** `/moai:git:history` 명령을 실행하면
**Then** 최근 10개 명령어와 각각의 체크포인트가 시간순으로 표시되어야 함

## 범위 및 모듈

### In Scope

- 4단계 핵심 파이프라인 최적화 (/moai:0~3)
- git 관련 확장 명령어 최적화 (/moai:git:\*)
- 실시간 UI/UX 피드백 시스템
- 자동 에러 복구 메커니즘
- 성능 프로파일링 및 캐싱 시스템
- 명령어 히스토리 및 롤백 기능

### Out of Scope

- Claude Code 내부 API 성능 개선 (외부 제어 불가)
- 명령어 문법 변경 (기존 호환성 유지)
- 다른 AI 플랫폼 지원 (Claude Code 전용)
- GUI 기반 인터페이스 (터미널 기반 유지)

## 기술 노트

### 구현 기술

- **진행률 표시**: Rich 라이브러리 활용한 progress bar
- **에러 복구**: tenacity 라이브러리 기반 재시도 로직
- **성능 측정**: cProfile + custom decorators
- **캐싱**: functools.lru_cache + 파일 기반 캐시

### 의존성

- **Rich**: 터미널 UI 향상 (진행률 바, 색상)
- **tenacity**: 재시도 로직 및 에러 복구
- **psutil**: 시스템 리소스 모니터링
- **기존**: click, toml, pathlib (현재 의존성 유지)

### 아키텍처 고려사항

- 명령어별 독립적인 성능 최적화
- 에러 복구 전략의 명령어별 커스터마이징
- 캐시 무효화 정책 (파일 변경 감지)
- 동시 실행 안전성 (file locking)

### 보안 고려사항

- 자동 복구 시 권한 에스컬레이션 방지
- 히스토리 데이터의 민감정보 마스킹
- 롤백 시 데이터 무결성 보장

### 성능 목표

- 명령어 실행 시간 20% 단축 (캐싱 적용 시)
- 에러 복구 성공률 80% 이상
- 사용자 피드백 응답 시간 < 100ms

## 추적성

### 연결된 요구사항

- @SPEC:USER-001: 1차 사용자(개인 개발자)의 즉시 결과 달성
- @DOC:STRATEGY-001: 4단계 워크플로우 완전 자동화
- @SPEC:SUCCESS-001: 주간 워크플로우 실행 빈도 향상

### 구현 우선순위

1. 진행률 표시 및 피드백 (High) - 즉시 사용자 경험 개선
2. 자동 에러 복구 (High) - 채택률 직결
3. 성능 최적화 (Medium) - 장기적 만족도
4. 히스토리 및 롤백 (Medium) - 고급 사용자 기능

### 테스트 전략

- 단위 테스트: 각 명령어별 성능 및 에러 처리 테스트
- 통합 테스트: 4단계 파이프라인 전체 시나리오 테스트
- 사용성 테스트: 실제 사용자 시나리오 기반 UX 검증

## 명령어 최적화 구현 예시

### 진행률 표시

```python
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.console import Console

def execute_with_progress(command_name: str, tasks: list):
    console = Console()

    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
    ) as progress:

        main_task = progress.add_task(f"Executing {command_name}...", total=len(tasks))

        for task in tasks:
            progress.update(main_task, description=f"Running: {task['name']}")
            result = task['function']()
            progress.advance(main_task)

            if not result['success']:
                console.print(f"❌ {task['name']} failed: {result['error']}", style="red")
                return False

        console.print(f"✅ {command_name} completed successfully!", style="green")
        return True
```

### 자동 에러 복구

```python
from tenacity import retry, stop_after_attempt, wait_exponential
import subprocess

@retry(stop=stop_after_attempt(3), wait=wait_exponential(multiplier=1, min=4, max=10))
def install_dependency(package_name: str) -> bool:
    """자동 종속성 설치 with 재시도"""
    try:
        result = subprocess.run([
            sys.executable, "-m", "pip", "install", package_name
        ], capture_output=True, text=True, check=True)
        return True
    except subprocess.CalledProcessError as e:
        print(f"패키지 설치 실패, 재시도 중... ({package_name})")
        raise

def auto_recover_file_permissions(file_path: Path) -> bool:
    """파일 권한 자동 복구"""
    try:
        if platform.system() != "Windows":
            os.chmod(file_path, 0o755)
        return True
    except PermissionError:
        print(f"권한 설정 실패: {file_path}")
        print("수동으로 다음 명령을 실행하세요:")
        print(f"chmod 755 {file_path}")
        return False
```

### 성능 프로파일링

```python
import time
import functools
from typing import Dict, Any

def profile_command(func):
    """명령어 실행 시간 프로파일링 데코레이터"""
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()

        execution_time = end_time - start_time

        # 성능 데이터 기록
        performance_log = {
            "command": func.__name__,
            "execution_time": execution_time,
            "timestamp": time.time(),
            "success": result.get("success", False)
        }

        save_performance_data(performance_log)

        if execution_time > 5.0:  # 5초 이상 소요 시 최적화 제안
            print(f"⚠️  {func.__name__} 실행 시간이 길었습니다 ({execution_time:.2f}s)")
            print("성능 최적화를 위해 캐시 또는 병렬 처리를 고려해보세요.")

        return result
    return wrapper
```
