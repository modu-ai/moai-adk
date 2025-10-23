# Hooks Phase 2: 검증 및 확장 설계

## Phase 1 완료 사항 요약

- ✅ 4개 hooks 활성화 (SessionStart, PreToolUse, UserPromptSubmit, SessionEnd)
- ✅ `$CLAUDE_PROJECT_DIR` 환경 변수 적용 (동적 경로 참조)
- ✅ 모듈화된 핸들러 구조 (`handlers/`, `core/`)
- ✅ 표준 JSON I/O 인터페이스 구현
- ✅ 분석 보고서 작성 (`.moai/reports/hooks-analysis-and-implementation.md`)

## Phase 2 목표

### 1단계: 새 프로젝트에서 Hooks 동작 검증

**목적**: 실제 프로젝트 초기화 시나리오에서 hooks가 정상 작동하는지 검증

**검증 대상**:
1. **SessionStart**: 프로젝트 상태 카드 표시
2. **PreToolUse**: 위험한 작업 감지 및 체크포인트 생성
3. **UserPromptSubmit**: JIT 컨텍스트 로드
4. **SessionEnd**: 세션 종료 정리

**검증 방법**:
- 임시 테스트 프로젝트 생성 (`/tmp/test-moai-hooks`)
- `/alfred:0-project` 명령 실행
- 각 hook 트리거 시나리오 실행
- 출력 로그 수집 및 분석

**예상 결과**:
```
SessionStart → 프로젝트 상태 요약 표시
  Language: Python
  Branch: main
  Changes: 0 files
  SPEC Progress: 0/0 (0%)

PreToolUse → 위험 작업 감지 시 체크포인트 생성
  (예: rm -rf, git merge, CLAUDE.md 편집)

UserPromptSubmit → 프롬프트 기반 컨텍스트 추가
  (예: "AUTH 수정" → AUTH 관련 파일 자동 로드)

SessionEnd → 정리 작업 완료
```

### 2단계: PostToolUse 훅 기능 확장 설계

**현재 상태**:
- `handlers/tool.py`의 `handle_post_tool_use()` → 빈 구현 (`return HookResult()`)
- PostToolUse 훅 → `settings.json`에서 비활성화 (`"PostToolUse": []`)

**확장 목표**:
코드 작성 후 자동으로 테스트를 실행하여 즉각적인 피드백 제공

**설계 원칙**:
1. **Non-blocking**: 테스트 실패 시에도 사용자 작업을 차단하지 않음
2. **Language-aware**: 파일 확장자 기반 언어 감지 및 적절한 테스트 명령 실행
3. **Performance**: <100ms 실행 시간 유지 (백그라운드 실행)
4. **Transparency**: 테스트 결과를 명확하게 표시

**트리거 조건**:
- Write/Edit 도구 사용 후
- 대상 파일이 코드 파일일 경우 (`.py`, `.ts`, `.go`, `.rs`, 등)
- 테스트 파일 자체 편집 시 제외 (`tests/`, `test_*.py`, `*.test.ts`, 등)

**언어별 테스트 명령 매핑**:

| 언어       | 파일 확장자    | 테스트 명령                | 조건                    |
| ---------- | -------------- | -------------------------- | ----------------------- |
| Python     | `.py`          | `pytest {file_path} -v`    | `pytest.ini` 존재       |
| TypeScript | `.ts`, `.tsx`  | `pnpm test {file_name}`    | `package.json` 존재     |
| JavaScript | `.js`, `.jsx`  | `npm test {file_name}`     | `package.json` 존재     |
| Go         | `.go`          | `go test ./{package}`      | `go.mod` 존재           |
| Rust       | `.rs`          | `cargo test`               | `Cargo.toml` 존재       |
| Java       | `.java`        | `./gradlew test`           | `build.gradle.kts` 존재 |

**구현 전략**:

```python
# handlers/tool.py

def handle_post_tool_use(payload: HookPayload) -> HookResult:
    """PostToolUse event handler (Auto Test Execution)

    Automatically runs tests after code file edits.

    Args:
        payload: Claude Code event payload
                 (includes tool, arguments, cwd keys)

    Returns:
        HookResult(
            message=test execution result summary;
            blocked=False (never blocks)
        )

    Trigger Conditions:
        - Tool: Write, Edit, MultiEdit
        - Target: Code files (not test files)
        - Detection: Language-aware test command selection

    Examples:
        Python file edit:
        → "✅ Tests passed: pytest tests/test_auth.py -v (2 passed)"

        Test failure:
        → "❌ Tests failed: 1 failed, 2 passed (see details above)"

    @TAG:POSTTOOL-AUTOTEST-001
    """
    tool_name = payload.get("tool", "")
    tool_args = payload.get("arguments", {})
    cwd = payload.get("cwd", ".")

    # Only trigger for Write/Edit tools
    if tool_name not in ["Write", "Edit", "MultiEdit"]:
        return HookResult()

    # Get edited file path
    file_path = _extract_file_path(tool_args)
    if not file_path:
        return HookResult()

    # Skip if editing test files
    if _is_test_file(file_path):
        return HookResult()

    # Detect language and get test command
    language = detect_language(cwd)
    test_cmd = _get_test_command(language, file_path, cwd)

    if not test_cmd:
        return HookResult()

    # Run tests (non-blocking, timeout 10s)
    result = _run_tests(test_cmd, cwd, timeout=10)

    return HookResult(
        message=result["message"],
        blocked=False
    )
```

**보조 함수**:

```python
def _extract_file_path(tool_args: dict) -> str | None:
    """Extract file path from tool arguments."""
    # Edit/Write → file_path key
    # MultiEdit → files list
    ...

def _is_test_file(file_path: str) -> bool:
    """Check if file is a test file."""
    # tests/, test_*.py, *.test.ts, *_spec.rb, etc.
    ...

def _get_test_command(language: str, file_path: str, cwd: str) -> str | None:
    """Get language-specific test command."""
    # Language detection → test command mapping
    ...

def _run_tests(cmd: str, cwd: str, timeout: int) -> dict:
    """Run test command and parse results."""
    # subprocess.run with timeout
    # Parse output → summary message
    ...
```

**출력 예시**:

```
Success:
✅ Tests passed (pytest)
   tests/test_auth.py::test_login PASSED
   tests/test_auth.py::test_logout PASSED
   2 passed in 0.5s

Failure:
❌ Tests failed (pytest)
   tests/test_auth.py::test_login FAILED
   1 failed, 1 passed in 0.7s

   Hint: Run 'pytest tests/test_auth.py -vv' for details
```

**성능 고려사항**:
- 테스트 실행 시간이 10초 초과 시 타임아웃
- 백그라운드 실행 (사용자 작업 차단 안 함)
- 빠른 테스트만 실행 (단위 테스트 위주)
- 통합 테스트는 별도 명령으로 실행

**에러 처리**:
- 테스트 명령 실패 → 에러 메시지만 표시, 작업 계속
- 타임아웃 → "테스트 실행 시간 초과" 메시지
- 테스트 도구 미설치 → 조용히 스킵

**보안 고려사항**:
- `subprocess.run(shell=False)` 사용 (명령 인젝션 방지)
- 허용된 테스트 명령만 실행 (화이트리스트)
- 테스트 출력 크기 제한 (1000자)

### 3단계: README.md에 Hooks 가이드 추가

**대상 파일**: `/Users/goos/MoAI/MoAI-ADK/README.md`

**추가 위치**:
- "Installation" 섹션과 "Usage" 섹션 사이
- 또는 별도 "Hooks System" 섹션 생성

**내용 구성**:

```markdown
## 🎣 Claude Code Hooks 가이드

### 개요

MoAI-ADK는 Claude Code와 통합되는 4가지 주요 Hook을 제공합니다.
자동 체크포인트, JIT 컨텍스트 로드, 세션 모니터링 등의 기능을 제공합니다.

### 설치된 Hooks

#### 1. SessionStart (세션 시작)

**언제**: Claude Code 세션 시작 시 자동 실행
**기능**: 프로젝트 상태 요약 표시

```
🚀 MoAI-ADK Session Started
   Language: Python
   Branch: develop
   Changes: 2 files
   SPEC Progress: 12/25 (48%)
```

**목적**: 세션 시작 시 프로젝트 현황을 한눈에 파악

#### 2. PreToolUse (도구 사용 전)

**언제**: 파일 편집, Bash 명령 실행 전
**기능**: 위험한 작업 감지 후 자동 체크포인트 생성

**감지 대상**:
- `rm -rf` (파일 삭제)
- `git merge`, `git reset --hard` (Git 위험 작업)
- `CLAUDE.md`, `config.json` 편집 (설정 파일 변경)
- 10개 이상 파일 동시 편집 (MultiEdit)

**예시**:
```
🛡️ Checkpoint created: before-delete-20251023-143000
   Operation: delete
```

**목적**: 실수로 인한 데이터 손실 방지

#### 3. UserPromptSubmit (프롬프트 입력)

**언제**: 사용자가 프롬프트 입력 시
**기능**: JIT 원칙에 따라 관련 문서 자동 로드

**예시**:
- "AUTH 수정" → AUTH 관련 SPEC, 테스트, 구현 파일 자동 추가
- "로그인 기능" → 로그인 관련 코드 컨텍스트 자동 로드

**목적**: 필요한 정보만 정확하게 로드하여 컨텍스트 효율성 향상

#### 4. SessionEnd (세션 종료)

**언제**: Claude Code 세션 종료 시
**기능**: 정리 작업 및 상태 저장

**목적**: 세션 종료 시 필요한 정리 작업 자동 수행

### 기술 상세

- **위치**: `.claude/hooks/alfred/`
- **환경 변수**: `$CLAUDE_PROJECT_DIR` (프로젝트 루트 동적 참조)
- **성능**: 각 훅 <100ms 실행 시간
- **로깅**: stderr로 에러 메시지 출력 (stdout은 JSON 전용)

### 비활성화 방법

Hooks를 비활성화하려면 `.claude/settings.json`에서 해당 훅을 제거하거나 주석 처리하세요:

```json
{
  "hooks": {
    "SessionStart": [],  // 비활성화
    "PreToolUse": [...]  // 활성화 유지
  }
}
```

### 트러블슈팅

**문제: Hook이 실행되지 않음**
- `.claude/settings.json` 파일 확인
- `uv` 명령이 설치되어 있는지 확인 (`which uv`)
- Hook 스크립트 권한 확인 (`chmod +x .claude/hooks/alfred/alfred_hooks.py`)

**문제: 성능 저하**
- Hook 실행 시간이 100ms를 초과하는지 확인
- 필요 없는 Hook 비활성화
- 로그에서 에러 메시지 확인 (`stderr` 출력)

**문제: 체크포인트가 너무 많이 생성됨**
- PreToolUse 트리거 조건 확인
- 필요 시 `core/checkpoint.py`에서 감지 조건 조정

### 향후 확장 계획

- **PostToolUse**: 코드 작성 후 자동 테스트 실행 (Phase 2 진행 중)
- **Notification**: 중요 이벤트 알림
- **Stop/SubagentStop**: Agent 종료 시 정리 작업

### 참고 문서

- 상세 분석: `.moai/reports/hooks-analysis-and-implementation.md`
- Phase 2 설계: `.moai/reports/hooks-phase2-design.md`
```

## 검증 계획

### 1단계 검증 체크리스트

- [ ] 테스트 프로젝트 생성 (`/tmp/test-moai-hooks`)
- [ ] Git 저장소 초기화
- [ ] MoAI-ADK hooks 복사 및 설정
- [ ] SessionStart 출력 확인
- [ ] PreToolUse 트리거 (rm 명령 실행)
- [ ] UserPromptSubmit 시뮬레이션
- [ ] SessionEnd 정리 확인

### 2단계 설계 검증

- [ ] `handle_post_tool_use()` 함수 시그니처 정의
- [ ] 언어 감지 로직 재사용 (`core/project.py`)
- [ ] 테스트 명령 매핑 테이블 작성
- [ ] 보조 함수 인터페이스 설계
- [ ] 에러 처리 시나리오 정의
- [ ] 성능 요구사항 명시 (<100ms 또는 백그라운드)

### 3단계 문서화 검증

- [ ] README.md 구조 분석
- [ ] Hooks 섹션 삽입 위치 결정
- [ ] 각 훅별 설명 작성 (When, What, Why)
- [ ] 예시 코드 및 출력 추가
- [ ] 트러블슈팅 가이드 작성
- [ ] 향후 확장 계획 명시

## 타임라인

- **1단계**: 검증 및 로그 수집 (1시간)
- **2단계**: PostToolUse 설계 및 코드 스켈레톤 작성 (1시간)
- **3단계**: README.md 문서화 (30분)
- **최종 검토**: 통합 검증 및 보고서 작성 (30분)

## 성공 기준

- ✅ 4개 hooks가 실제 프로젝트에서 정상 작동
- ✅ PostToolUse 확장 설계 완료 (구현 전 설계)
- ✅ README.md에 Hooks 가이드 추가
- ✅ Phase 2 완료 보고서 작성

---

**작성일**: 2025-10-23
**작성자**: cc-manager Agent (Phase 2)
**상태**: 진행 중
