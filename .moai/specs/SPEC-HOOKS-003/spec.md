---
id: HOOKS-003
version: 0.1.0
status: in-progress
created: 2025-10-16
updated: 2025-10-18
author: @Goos
priority: high
category: feature
labels:
  - hooks
  - trust
  - automation
  - quality-gate
  - post-tool-use
depends_on:
  - HOOKS-001
  - TRUST-001
related_specs:
  - HOOKS-001
  - TRUST-001
  - HOOKS-002
scope:
  packages:
    - .claude/hooks/alfred/handlers
  files:
    - tool.py
---

# @SPEC:HOOKS-003: TRUST 원칙 자동 검증 (PostToolUse 통합)

## HISTORY

### v0.1.0 (2025-10-18)
- **CHANGED**: TDD 구현 완료, status를 completed로 변경
- **AUTHOR**: @Goos
- **REVIEW**: 구현 검증 완료

### v0.0.1 (2025-10-16)
- **INITIAL**: `/alfred:2-run` 완료 후 TRUST 검증 자동 실행 명세 작성
- **AUTHOR**: @Goos
- **CONTEXT**: validation-logic-migration.md Phase 1 구현
- **DEPENDS_ON**: HOOKS-001 (Hooks 시스템 마스터플랜), TRUST-001 (검증 시스템)
- **REASON**: Alfred 3-stage 워크플로우에서 품질 게이트 자동화 필요

---

## Environment

**PostToolUse Hook 실행 환경**:
- **Trigger**: Claude Code의 PostToolUse 이벤트 (도구 실행 완료 후)
- **Constraint**: 100ms 제약 (동기 실행 금지)
- **Context**: Git 저장소 상태, 최근 커밋 로그, 파일 변경사항
- **Runtime**: Python 3.10+, subprocess를 통한 비동기 프로세스 실행

**Alfred 3-Stage 컨텍스트**:
- **Stage 1**: `/alfred:1-plan` → SPEC 문서 작성 (TRUST 검증 불필요)
- **Stage 2**: `/alfred:2-run` → TDD 구현 (RED-GREEN-REFACTOR)
- **Stage 3**: `/alfred:3-sync` → 문서 동기화 및 TAG 검증

**TRUST-001 검증 시스템**:
- **위치**: `scripts/validate_trust.py`
- **실행 방식**: CLI 도구 (`python scripts/validate_trust.py`)
- **출력**: JSON 형식 보고서 (stdout)
- **종료 코드**: 0 (성공), 1 (실패)

---

## Assumptions

1. **TDD 단계 감지 가능**:
   - Git 커밋 메시지에 `🟢 GREEN:` 또는 `♻️ REFACTOR:` 포함 시 TDD 구현 완료로 판단
   - 최근 5개 커밋 로그를 분석하여 RED → GREEN → REFACTOR 흐름 확인

2. **비동기 실행 가능**:
   - `subprocess.Popen()`으로 백그라운드 프로세스 실행
   - PostToolUse 핸들러는 100ms 이내에 반환 (blocked=false)
   - 검증 결과는 별도 notification 메시지로 전달

3. **Git 저장소 가용성**:
   - `.git/` 디렉토리 존재 확인
   - `git log` 명령 실행 가능

4. **TRUST-001 검증 도구 설치**:
   - `scripts/validate_trust.py` 파일 존재
   - 필수 의존성 (pytest, coverage, ruff 등) 설치 완료

5. **성능 제약**:
   - Git 로그 파싱: <10ms
   - 검증 프로세스 시작: <50ms
   - 전체 핸들러 실행: <100ms (동기 부분만)

---

## Requirements

### Ubiquitous Requirements (필수 기능)

**U1. TDD 완료 감지 시스템**:
- 시스템은 PostToolUse 이벤트 발생 시 Git 커밋 로그를 자동으로 분석해야 한다.
- 시스템은 최근 커밋이 `🟢 GREEN:` 또는 `♻️ REFACTOR:` 단계임을 감지해야 한다.

**U2. 비동기 검증 실행**:
- 시스템은 TRUST 원칙 검증을 백그라운드 프로세스로 실행해야 한다.
- 시스템은 PostToolUse 핸들러를 100ms 이내에 반환해야 한다.

**U3. 검증 결과 보고**:
- 시스템은 검증 결과를 JSON 형식으로 파싱해야 한다.
- 시스템은 통과/실패 여부를 사용자에게 알림 메시지로 전달해야 한다.

### Event-driven Requirements (이벤트 기반)

**E1. TDD 단계 커밋 감지**:
- WHEN 최근 커밋 메시지에 `🟢 GREEN:` 또는 `♻️ REFACTOR:`가 포함되어 있으면, 시스템은 TRUST 검증을 트리거해야 한다.

**E2. Alfred 2-build 완료 감지**:
- WHEN PostToolUse 이벤트 payload에 `alfred:2-build` 키워드가 포함되어 있으면, 시스템은 TDD 구현 완료로 판단해야 한다.

**E3. 검증 실패 처리**:
- WHEN TRUST 검증이 실패하면, 시스템은 실패 원인을 포함한 상세 보고서를 생성해야 한다.
- WHEN 테스트 커버리지가 85% 미만이면, 시스템은 ⚠️ Warning 메시지를 출력해야 한다.

### State-driven Requirements (상태 기반)

**S1. Git 저장소 상태**:
- WHILE `.git/` 디렉토리가 존재하는 동안, 시스템은 Git 명령을 실행할 수 있어야 한다.
- WHILE Git 저장소가 "detached HEAD" 상태이면, 시스템은 검증을 건너뛰고 ℹ️ Info 메시지를 출력해야 한다.

**S2. 검증 프로세스 실행 중**:
- WHILE TRUST 검증 프로세스가 실행 중인 동안, 시스템은 중복 실행을 방지해야 한다.

### Constraints (제약사항)

**C1. 성능 제약**:
- IF PostToolUse 핸들러 실행 시간이 100ms를 초과하면, 시스템은 검증을 비동기로 전환해야 한다.
- IF Git 로그 파싱에 10ms 이상 소요되면, 시스템은 캐싱 메커니즘을 도입해야 한다.

**C2. 의존성 제약**:
- IF `scripts/validate_trust.py`가 존재하지 않으면, 시스템은 검증을 건너뛰고 ℹ️ Info 메시지를 출력해야 한다.
- IF 필수 의존성 (pytest, coverage, ruff)이 설치되지 않았으면, 시스템은 ❌ Critical 메시지를 출력하고 설치 가이드를 제공해야 한다.

---

## Specifications

### 1. TDD 완료 감지 로직 (handlers/tool.py)

**Git 로그 분석**:
```python
def detect_tdd_completion() -> bool:
    """
    최근 5개 커밋을 분석하여 TDD 구현 완료 여부 확인.

    Returns:
        True: GREEN 또는 REFACTOR 단계 감지
        False: TDD 구현 미완료
    """
    # Git 로그 가져오기
    result = subprocess.run(
        ["git", "log", "-5", "--pretty=format:%s"],
        capture_output=True,
        text=True,
        timeout=1.0
    )

    if result.returncode != 0:
        return False

    commit_messages = result.stdout.strip().split('\n')

    # TDD 단계 키워드 검색
    tdd_keywords = ["🟢 GREEN:", "♻️ REFACTOR:"]

    for msg in commit_messages:
        if any(keyword in msg for keyword in tdd_keywords):
            return True

    return False
```

**Alfred 2-build 감지**:
```python
def is_alfred_build_command(payload: dict) -> bool:
    """
    PostToolUse payload에서 alfred:2-build 실행 여부 확인.

    Args:
        payload: PostToolUse 이벤트 데이터

    Returns:
        True: alfred:2-build 실행됨
        False: 다른 명령 실행됨
    """
    tool_name = payload.get("tool", "")
    tool_input = payload.get("input", {})

    # Bash 명령어 또는 Agent 호출 확인
    command = tool_input.get("command", "")
    description = tool_input.get("description", "")

    return "alfred:2-build" in command or "alfred:2-build" in description
```

### 2. 비동기 TRUST 검증 실행

**백그라운드 프로세스 실행**:
```python
def trigger_trust_validation() -> subprocess.Popen:
    """
    TRUST 검증을 백그라운드 프로세스로 실행.

    Returns:
        subprocess.Popen 객체 (비동기 실행)
    """
    process = subprocess.Popen(
        ["python", "scripts/validate_trust.py", "--json"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        cwd=project_root()
    )

    return process
```

**검증 결과 수집** (별도 스레드 또는 다음 Hook 이벤트):
```python
def collect_validation_result(process: subprocess.Popen) -> dict:
    """
    TRUST 검증 결과를 수집하고 파싱.

    Args:
        process: 실행 중인 검증 프로세스

    Returns:
        JSON 형식 검증 보고서
    """
    stdout, stderr = process.communicate(timeout=30.0)

    if process.returncode != 0:
        return {
            "status": "failed",
            "error": stderr,
            "exit_code": process.returncode
        }

    return json.loads(stdout)
```

### 3. PostToolUse 핸들러 통합

**handlers/tool.py 확장**:
```python
def handle_post_tool_use(payload: dict) -> HookResult:
    """
    PostToolUse 이벤트 처리: TDD 완료 감지 및 TRUST 검증.

    Args:
        payload: PostToolUse 이벤트 데이터

    Returns:
        HookResult (blocked=False, 알림 메시지 포함)
    """
    # 1. TDD 완료 감지
    if not (detect_tdd_completion() or is_alfred_build_command(payload)):
        return HookResult(blocked=False)  # 검증 불필요

    # 2. 검증 도구 존재 확인
    validate_script = project_root() / "scripts" / "validate_trust.py"
    if not validate_script.exists():
        return HookResult(
            blocked=False,
            message="ℹ️ TRUST 검증 도구가 없습니다. scripts/validate_trust.py 설치 필요"
        )

    # 3. 비동기 검증 실행
    try:
        process = trigger_trust_validation()

        # 프로세스 ID를 임시 파일에 저장 (다음 Hook에서 수집)
        save_validation_pid(process.pid)

        return HookResult(
            blocked=False,
            message="🔍 TRUST 원칙 검증 중... (백그라운드 실행)"
        )

    except Exception as e:
        return HookResult(
            blocked=False,
            message=f"⚠️ TRUST 검증 시작 실패: {str(e)}"
        )
```

### 4. 검증 결과 알림 (handlers/notification.py 확장)

**SessionStart 또는 UserMessage에서 결과 수집**:
```python
def collect_pending_validation_results() -> list[str]:
    """
    이전 PostToolUse에서 시작된 TRUST 검증 결과를 수집.

    Returns:
        알림 메시지 목록
    """
    messages = []

    # 저장된 프로세스 ID 목록 읽기
    pids = load_validation_pids()

    for pid in pids:
        try:
            process = psutil.Process(pid)

            # 프로세스가 아직 실행 중이면 건너뜀
            if process.is_running():
                continue

            # 프로세스 종료 → 결과 파일 읽기
            result_file = get_validation_result_path(pid)
            if result_file.exists():
                result = json.loads(result_file.read_text())
                messages.append(format_validation_result(result))
                result_file.unlink()  # 읽은 후 삭제

        except (psutil.NoSuchProcess, FileNotFoundError):
            continue

    # 처리 완료된 PID 제거
    clear_validation_pids(pids)

    return messages


def format_validation_result(result: dict) -> str:
    """
    TRUST 검증 결과를 Markdown 형식으로 변환.

    Args:
        result: JSON 형식 검증 보고서

    Returns:
        Markdown 형식 알림 메시지
    """
    if result["status"] == "passed":
        return f"""
✅ **TRUST 원칙 검증 통과**
- 테스트 커버리지: {result["test_coverage"]}%
- 코드 제약 준수: {result["code_constraints_passed"]}/{result["code_constraints_total"]}
- TAG 체인 무결성: OK
"""

    else:
        return f"""
❌ **TRUST 원칙 검증 실패**
- 실패 원인: {result["error"]}
- 테스트 커버리지: {result.get("test_coverage", "N/A")}% (목표 85%)
- 권장 조치: {result.get("recommendation", "scripts/validate_trust.py 실행하여 상세 확인")}
"""
```

---

## Traceability (@TAG)

### TAG 체인
- **SPEC**: `@SPEC:HOOKS-003` (본 문서)
- **TEST**: `@TEST:HOOKS-003` (tests/unit/test_hooks_trust_validation.py)
- **CODE**: `@CODE:HOOKS-003` (.claude/hooks/alfred/handlers/tool.py)

### 의존성 TAG
- **@SPEC:HOOKS-001**: Hooks 시스템 아키�ecture (PostToolUse 핸들러 기반)
- **@SPEC:TRUST-001**: TRUST 원칙 검증 시스템 (scripts/validate_trust.py)
- **@SPEC:HOOKS-002**: SPEC 메타데이터 검증 (유사한 자동화 패턴)

### 코드 위치
```
.claude/hooks/alfred/
├── handlers/
│   ├── tool.py              # @CODE:HOOKS-003 (PostToolUse 핸들러)
│   └── notification.py      # @CODE:HOOKS-003 (결과 알림)
├── core/
│   └── validation.py        # @CODE:HOOKS-003 (검증 유틸리티)
└── tests/
    └── unit/
        └── test_hooks_trust_validation.py  # @TEST:HOOKS-003
```

---

## Testing Strategy

### 단위 테스트 (≥85% 커버리지)

1. **TDD 감지 로직**:
   - Git 로그 파싱 정확성 (GREEN, REFACTOR 키워드)
   - Alfred 2-build 명령 감지 (payload 분석)

2. **비동기 실행**:
   - subprocess.Popen() 호출 성공
   - 프로세스 ID 저장/로드

3. **결과 수집**:
   - JSON 파싱 정확성
   - 통과/실패 메시지 포맷

4. **에러 처리**:
   - Git 저장소 없음
   - 검증 도구 없음
   - 의존성 미설치

### 통합 테스트

1. **End-to-End 시나리오**:
   - `/alfred:2-run SPEC-XXX` 실행
   - REFACTOR 커밋 생성
   - PostToolUse 트리거
   - 검증 결과 알림 확인

2. **성능 테스트**:
   - PostToolUse 핸들러 실행 시간 <100ms
   - Git 로그 파싱 <10ms

---

## Implementation Notes

### 기술적 도전과제

1. **비동기 실행 타이밍**:
   - PostToolUse는 도구 실행 직후 호출됨
   - 검증 결과는 다음 Hook 이벤트에서 수집해야 함

2. **프로세스 관리**:
   - 좀비 프로세스 방지 (psutil 사용)
   - 프로세스 ID 영속화 (임시 파일)

3. **Git 저장소 상태**:
   - Detached HEAD 상태 처리
   - Bare repository 제외

### 마이그레이션 전략 (Phase 1)

**참조 문서**: `validation-logic-migration.md`

**단계**:
1. **SPEC 작성** (현재 문서)
2. **TDD 구현** (`/alfred:2-run HOOKS-003`)
3. **Alfred 통합** (handlers/tool.py 확장)

---

**Last Updated**: 2025-10-18
**Author**: @Goos
**Status**: Completed (v0.1.0)
