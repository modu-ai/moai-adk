---
id: CMD-IMPROVE-001
version: 0.0.1
status: draft
created: 2025-11-12
updated: 2025-11-12
author: @goos
priority: high
category: refactoring
labels: [commands, context-passing, resume, architecture, integration]
depends_on: []
related_specs: []
scope: Commands 레이어 4개 파일 (0-project, 1-plan, 2-run, 3-sync) 리팩토링
---

# @SPEC:CMD-IMPROVE-001: Commands 레이어 컨텍스트 전달 및 Resume 기능 통합 개선

## HISTORY

### v0.0.1 - 2025-11-12
- **Author**: @goos
- **Created**: 초기 통합 SPEC 생성
- **Scope**: 명시적 컨텍스트 전달 시스템 + Resume 기능 통합
- **Impact**: 25건 이슈 해결 (CRITICAL 8건, HIGH 17건)

---

## SUMMARY (English)

**Objective**: Enhance MoAI-ADK Commands layer with explicit context passing and resume functionality

**Key Features**:
1. **Explicit Context Passing System**: Store phase results in JSON, pass to agents explicitly, replace template variables at runtime, validate absolute file paths
2. **Resume Functionality**: Utilize `.moai/memory/command-state/` directory, validate timestamp and state, resume interrupted commands, auto-invalidate after 30 days

**Expected Outcomes**:
- Eliminate 8 CRITICAL issues (remove session-dependent references)
- Resolve 17 HIGH issues (placeholder substitution, explicit paths)
- Improve user experience with resume capability

**Implementation Timeline**: 8 weeks
- Week 1-4: Context Passing System
- Week 5-8: Resume Functionality

---

## 요구사항 (EARS 구조)

### 1. Ubiquitous Requirements (필수 요구사항)

#### REQ-1.1: 명시적 컨텍스트 전달 시스템

**설명**: Commands 레이어의 각 Phase는 실행 결과를 구조화된 JSON 형식으로 저장하고, 다음 Phase의 Agent 호출 시 명시적으로 전달해야 한다.

**근거**:
- 현재 세션 의존 참조(`위에서 언급한`, `앞서 분석한`)로 인한 컨텍스트 손실 문제 해결
- Agent 간 데이터 전달의 신뢰성 및 추적 가능성 확보

**구현 세부사항**:
```python
# Phase 결과 저장 구조
{
  "phase": "0-project",
  "timestamp": "2025-11-12T10:30:00Z",
  "status": "completed",
  "outputs": {
    "project_name": "MyProject",
    "mode": "personal",
    "language": "ko",
    "tech_stack": ["python", "fastapi"]
  },
  "files_created": [
    "/abs/path/to/.moai/project/product.md",
    "/abs/path/to/.moai/project/structure.md"
  ],
  "next_phase": "1-plan"
}
```

**검증 기준**:
- Phase 완료 시 JSON 파일이 `.moai/memory/command-state/` 디렉토리에 생성됨
- Agent 호출 시 `Task()` prompt에 이전 Phase 결과가 포함됨
- 템플릿 변수 `{{PROJECT_NAME}}`, `{{MODE}}` 등이 런타임에 치환됨

---

#### REQ-1.2: 절대 경로 검증 시스템

**설명**: 모든 파일 경로는 절대 경로로 변환되고 존재 여부가 검증되어야 한다.

**근거**:
- 상대 경로로 인한 파일 접근 실패 방지
- Agent 간 경로 참조의 일관성 보장

**구현 세부사항**:
```python
import os

def validate_and_convert_path(relative_path: str, project_root: str) -> str:
    """상대 경로를 절대 경로로 변환하고 검증"""
    abs_path = os.path.abspath(os.path.join(project_root, relative_path))

    if not abs_path.startswith(project_root):
        raise ValueError(f"Path outside project root: {abs_path}")

    # 디렉토리는 존재 확인만, 파일은 생성 예정 경로도 허용
    if os.path.isdir(abs_path):
        return abs_path

    parent_dir = os.path.dirname(abs_path)
    if not os.path.exists(parent_dir):
        raise FileNotFoundError(f"Parent directory not found: {parent_dir}")

    return abs_path
```

**검증 기준**:
- 모든 경로가 절대 경로 형식으로 Agent에 전달됨
- 존재하지 않는 경로 참조 시 명확한 에러 메시지 제공
- Project root 외부 경로 접근 시도 차단

---

#### REQ-1.3: Resume 기능 핵심 메커니즘

**설명**: 사용자는 중단된 Command를 저장된 상태로부터 재개할 수 있어야 한다.

**근거**:
- 긴 작업 중 세션 중단 시 처음부터 재실행 방지
- 사용자 경험 개선 및 작업 효율성 증대

**구현 세부사항**:
```python
# Resume 상태 저장 구조
{
  "command": "alfred:2-run",
  "spec_id": "SPEC-AUTH-001",
  "current_phase": "implementation",
  "completed_steps": [
    "spec_validation",
    "test_setup"
  ],
  "pending_steps": [
    "code_implementation",
    "integration_test"
  ],
  "timestamp": "2025-11-12T10:30:00Z",
  "expiry": "2025-12-12T10:30:00Z",  # 30일 후
  "context": {
    "branch": "feature/SPEC-AUTH-001",
    "last_commit": "abc123",
    "files_modified": ["/abs/path/to/auth.py"]
  }
}
```

**검증 기준**:
- Resume 명령 실행 시 저장된 단계부터 재개됨
- 30일 경과한 상태는 자동 무효화되고 경고 메시지 출력
- Resume 불가능한 상태(conflict, invalid)는 명확한 에러 제공

---

### 2. Event-driven Requirements (이벤트 기반 요구사항)

#### REQ-2.1: Phase 완료 이벤트

**WHEN** `/alfred:0-project` 명령이 완료되면,
**THEN** 시스템은 다음을 수행해야 한다:
1. `.moai/memory/command-state/0-project-{timestamp}.json` 파일 생성
2. 프로젝트 메타데이터, 생성된 파일 경로, 다음 Phase 정보 저장
3. 사용자에게 다음 단계 안내 출력

**검증**:
```bash
# Phase 완료 후 파일 존재 확인
ls .moai/memory/command-state/0-project-*.json

# JSON 내용 검증
jq '.status == "completed" and .next_phase == "1-plan"' .moai/memory/command-state/0-project-*.json
```

---

#### REQ-2.2: Agent 호출 이벤트

**WHEN** Command가 Sub-agent를 호출하면,
**THEN** 시스템은 다음을 수행해야 한다:
1. 이전 Phase의 JSON 결과를 로드
2. 템플릿 변수를 실제 값으로 치환
3. 절대 경로로 변환된 파일 목록을 포함하여 `Task()` 호출

**검증**:
```python
# Task() 호출 시 컨텍스트 전달 확인
assert "{{PROJECT_NAME}}" not in task_prompt
assert "/Users/goos/MoAI/MyProject/.moai/project/product.md" in task_prompt
```

---

#### REQ-2.3: Resume 요청 이벤트

**WHEN** 사용자가 `/alfred:resume 2-run SPEC-XXX` 명령을 실행하면,
**THEN** 시스템은 다음을 수행해야 한다:
1. `.moai/memory/command-state/2-run-SPEC-XXX-*.json` 파일 검색
2. Timestamp 유효성 검증 (30일 이내)
3. 저장된 `current_phase`부터 작업 재개
4. 이미 완료된 단계는 건너뛰고 대기 중인 단계만 실행

**검증**:
```bash
# Resume 상태 검증
jq '.expiry > now | tonumber' .moai/memory/command-state/2-run-SPEC-XXX-*.json

# 완료된 단계 스킵 확인
# (구현 시 로그 출력으로 검증)
```

---

### 3. State-driven Requirements (상태 기반 요구사항)

#### REQ-3.1: Command 실행 중 상태

**WHILE** Command가 실행 중일 때,
**THEN** 시스템은 다음을 유지해야 한다:
1. 현재 Phase 정보를 메모리에 보관
2. 각 단계 완료 시 `completed_steps` 배열 업데이트
3. 오류 발생 시 현재 상태를 JSON에 저장 후 중단

**상태 전이 다이어그램**:
```
[IDLE] → [RUNNING] → [COMPLETED]
   ↓          ↓
[RESUMABLE] ← [ERROR]
```

**검증**:
- 중단 후 저장된 JSON의 `status` 필드가 `error` 또는 `interrupted`
- Resume 시 `completed_steps` 길이가 증가

---

#### REQ-3.2: 데이터 저장 중 상태

**WHILE** Phase 결과를 JSON에 저장하는 동안,
**THEN** 시스템은 다음을 보장해야 한다:
1. 원자적 쓰기 (임시 파일 사용 후 atomic rename)
2. 저장 실패 시 이전 상태 파일 유지
3. 파일 권한 검증 (읽기/쓰기 가능 여부)

**구현 패턴**:
```python
import os
import json
import tempfile

def atomic_write_json(data: dict, target_path: str):
    """원자적 JSON 파일 쓰기"""
    temp_fd, temp_path = tempfile.mkstemp(dir=os.path.dirname(target_path))

    try:
        with os.fdopen(temp_fd, 'w') as f:
            json.dump(data, f, indent=2)

        os.replace(temp_path, target_path)  # Atomic rename
    except Exception as e:
        os.unlink(temp_path)
        raise IOError(f"Failed to write {target_path}: {e}")
```

**검증**:
- 저장 중 프로세스 강제 종료 시 손상된 JSON 파일 미발생
- 권한 오류 시 명확한 에러 메시지 출력

---

### 4. Optional Requirements (선택적 요구사항)

#### REQ-4.1: Resume 재시작 옵션

**설명**: 사용자는 Resume 시 특정 단계부터 강제 재시작할 수 있다.

**사용 예시**:
```bash
# 저장된 위치부터 재개 (기본값)
/alfred:resume 2-run SPEC-AUTH-001

# 특정 단계부터 재시작
/alfred:resume 2-run SPEC-AUTH-001 --from=integration_test

# 처음부터 재실행 (Resume 상태 무시)
/alfred:resume 2-run SPEC-AUTH-001 --restart
```

**구현 우선순위**: LOW (Phase 2 완료 후 추가 고려)

---

#### REQ-4.2: 디버깅 로그 출력

**설명**: 개발자는 `--debug` 플래그를 통해 컨텍스트 전달 과정을 상세히 확인할 수 있다.

**출력 예시**:
```
[DEBUG] Loading context from .moai/memory/command-state/0-project-20251112.json
[DEBUG] Replacing template variable: {{PROJECT_NAME}} → MyProject
[DEBUG] Validating path: .moai/project/product.md → /Users/goos/MoAI/MyProject/.moai/project/product.md
[DEBUG] Calling Task(subagent_type="plan-agent", prompt="...")
```

**구현 우선순위**: MEDIUM (Phase 1 후반부에 추가)

---

### 5. Unwanted Behaviors (비정상 동작 방지)

#### UB-5.1: JSON 저장 실패 시 세션 손실

**문제**: Phase 완료 후 JSON 저장 실패 시 모든 결과 손실

**방지 방법**:
- 저장 실패 시 임시 메모리 백업 유지
- 사용자에게 재시도 옵션 제공
- 에러 로그에 저장 시도 내용 기록

**검증**:
```python
# 저장 실패 시나리오 테스트
with mock.patch('os.replace', side_effect=OSError):
    result = save_phase_result(data)
    assert result.status == "save_failed"
    assert result.backup_available is True
```

---

#### UB-5.2: 오래된 Resume 상태 사용

**문제**: 30일 이상 경과한 상태로 Resume 시도 시 코드베이스 불일치

**방지 방법**:
- Resume 시 Timestamp 검증
- 경과 시간에 따른 경고 메시지 출력
- 30일 초과 시 자동 무효화 및 재시작 권장

**검증**:
```bash
# 만료된 상태 검증
expiry_date=$(jq -r '.expiry' state.json)
current_date=$(date -u +%s)

if [ $current_date -gt $expiry_date ]; then
  echo "ERROR: Resume state expired"
  exit 1
fi
```

---

#### UB-5.3: 잘못된 템플릿 변수 치환

**문제**: 치환되지 않은 `{{VARIABLE}}` 형식의 문자열이 Agent에 전달됨

**방지 방법**:
- Agent 호출 전 미치환 변수 검사
- 정규식 패턴으로 `{{.*}}` 형식 검출
- 검출 시 명확한 에러 메시지 출력

**검증**:
```python
import re

def validate_no_template_vars(text: str):
    """미치환 템플릿 변수 검사"""
    pattern = r'\{\{[A-Z_]+\}\}'
    matches = re.findall(pattern, text)

    if matches:
        raise ValueError(f"Unsubstituted template variables: {matches}")
```

---

## Traceability (@TAG)

### SPEC
- **Primary**: @SPEC:CMD-IMPROVE-001
- **Related**: @SPEC:ALF-WORKFLOW-001 (Alfred workflow architecture)

### TEST
- **Unit Tests**: `tests/commands/test_context_passing.py`
- **Integration Tests**: `tests/commands/test_resume_functionality.py`
- **E2E Tests**: `tests/e2e/test_full_workflow_with_resume.py`

### CODE
- **Main Implementation**:
  - `src/moai_adk/templates/.claude/commands/alfred/0-project.md`
  - `src/moai_adk/templates/.claude/commands/alfred/1-plan.md`
  - `src/moai_adk/templates/.claude/commands/alfred/2-run.md`
  - `src/moai_adk/templates/.claude/commands/alfred/3-sync.md`
- **Context Manager**: `src/moai_adk/core/context_manager.py` (new)
- **Resume Handler**: `src/moai_adk/core/resume_handler.py` (new)

### DOC
- **User Guide**: `docs/commands/context-passing-guide.md`
- **Developer Guide**: `docs/architecture/command-state-management.md`
- **API Reference**: `docs/api/context-manager.md`

---

## Constraints and Assumptions (제약사항 및 가정)

### Constraints
1. **하위 호환성**: 기존 `.moai/config.json` 구조 변경 불가
2. **파일 시스템 의존성**: JSON 저장/로드를 위한 파일 시스템 접근 필수
3. **메모리 제약**: 대규모 프로젝트에서 전체 컨텍스트를 메모리에 로드할 수 없음

### Assumptions
1. 사용자는 `.moai/memory/` 디렉토리에 읽기/쓰기 권한을 가짐
2. Python 3.9+ 환경 (pathlib, json 표준 라이브러리 사용)
3. 세션 간 파일 시스템 상태는 유지됨 (컨테이너/VM 재시작 고려 필요)

---

## Risks and Mitigations (위험 요소 및 대응 방안)

### Risk 1: JSON 파싱 오류
- **발생 가능성**: MEDIUM
- **영향도**: HIGH
- **대응 방안**:
  - JSON 스키마 검증 추가
  - 손상된 파일 감지 시 백업 파일 로드
  - 사용자에게 명확한 복구 가이드 제공

### Risk 2: 디스크 공간 부족
- **발생 가능성**: LOW
- **영향도**: MEDIUM
- **대응 방안**:
  - 30일 이상 경과한 상태 파일 자동 삭제
  - 최대 50개 상태 파일 제한
  - 디스크 공간 부족 시 경고 메시지 출력

### Risk 3: 병렬 실행 충돌
- **발생 가능성**: LOW (Personal 모드에서는 거의 없음)
- **영향도**: HIGH
- **대응 방안**:
  - 파일 잠금(file locking) 메커니즘 추가
  - 중복 실행 감지 및 경고
  - Team 모드에서만 활성화

---

## Definition of Done (완료 기준)

### Phase 1: 컨텍스트 전달 시스템
- [ ] 4개 Command 파일에서 명시적 컨텍스트 전달 구현 완료
- [ ] `.moai/memory/command-state/` 디렉토리 생성 및 JSON 저장 기능 구현
- [ ] 템플릿 변수 치환 엔진 구현 및 검증
- [ ] 절대 경로 변환 및 검증 함수 구현
- [ ] Unit 테스트 커버리지 90% 이상
- [ ] 통합 테스트 시나리오 5개 이상 통과

### Phase 2: Resume 기능
- [ ] Resume 명령 파싱 및 상태 로드 기능 구현
- [ ] Timestamp 검증 및 자동 만료 로직 구현
- [ ] 단계별 재개 메커니즘 구현
- [ ] 사용자 가이드 문서 작성 완료
- [ ] E2E 테스트 시나리오 3개 이상 통과
- [ ] 베타 사용자 피드백 수집 및 반영

---

## References (참고 자료)

- **Plan Document**: `.moai/specs/SPEC-CMD-IMPROVE-001/plan.md`
- **Acceptance Criteria**: `.moai/specs/SPEC-CMD-IMPROVE-001/acceptance.md`
- **Related Issue**: GitHub Issue #XXX (생성 예정)
- **Design Document**: `.moai/memory/design/command-context-architecture.md` (생성 예정)
