---
id: HOOKS-002
version: 0.1.0
status: completed
created: 2025-10-13
updated: 2025-10-14
author: @Goos
priority: critical
category: feature
labels:
  - hooks
  - python
  - self-contained
  - pep-723
  - zero-dependencies
depends_on:
  - PY314-001
scope:
  packages:
    - src/moai_adk/templates/.claude/hooks/
  files:
    - moai_hooks.py
---

# @SPEC:HOOKS-002: moai_hooks.py Self-contained Hook Script

## HISTORY

### v0.1.0 (2025-10-14)
- **TDD 완료**: moai_hooks.py 구현 완료 (373 LOC, 97% 커버리지)
- **테스트**: 49개 테스트 모두 통과
  - Phase 1: Utility Functions (36 tests) - Language Detection, Git Info, SPEC Count, JIT Context
  - Phase 2: Hook Handlers (9 tests) - 9개 Claude Code Hook 이벤트
  - Phase 3: Main Integration (4 tests) - main(), 라우팅, JSON I/O
- **품질**: mypy strict mode + ruff 린트 통과
- **성능**: SessionStart < 500ms, 기타 이벤트 < 100ms, 메모리 < 50MB
- **실행**: PEP 723 준수, 실행 가능 스크립트 (chmod +x)
- **AUTHOR**: @Claude
- **COMMITS**:
  - 0b8dc24: 🔴 RED: SPEC-HOOKS-002 테스트 작성 (moai_hooks.py)
  - 22756b2: 🟢 GREEN: SPEC-HOOKS-002 구현 완료 (moai_hooks.py)

### v0.0.1 (2025-10-13)
- **INITIAL**: moai_hooks.py 자립형 훅 스크립트 명세 작성
- **AUTHOR**: @Goos
- **SCOPE**: 600 LOC Python 단일 파일, Zero Dependencies, PEP 723
- **REASON**: TypeScript hooks 제거, Python 전환, Claude Code Hooks 지원
- **CONTEXT**: MoAI-ADK v0.3.0 전환 - 완전한 TypeScript 의존성 제거
- **BREAKING**: templates/.claude/hooks/index.ts 제거, moai_hooks.py로 대체

---

## 개요

`moai_hooks.py`는 Claude Code의 Hook 시스템을 구현하는 600 LOC 자립형 Python 스크립트입니다. **외부 의존성 없이** Python 표준 라이브러리만 사용하며, PEP 723 inline metadata를 포함합니다. `uv run --python 3.13 .claude/hooks/moai_hooks.py {event}` 형식으로 실행되며, stdin으로 JSON payload를 받고 stdout으로 JSON result를 출력합니다.

### 핵심 가치

1. **Zero Dependencies**: pip 설치 없이 실행 가능, Python 표준 라이브러리만 사용
2. **Self-contained**: 단일 파일로 모든 기능 제공 (600 LOC 이하)
3. **PEP 723 준수**: inline metadata로 requires-python 명시
4. **Fast Execution**: 대부분의 hook이 100ms 이하로 실행
5. **Multi-language Support**: 20개 언어 프로젝트 감지 지원
6. **JIT Context**: 필요한 순간에만 문서 로드 (메모리 효율)

### 전환 이유

**TypeScript hooks (v0.2.x) 문제점**:
- npm 의존성 관리 필요 (node_modules 크기)
- tsup 빌드 과정 필요
- 런타임 의존성 (Bun/Node.js)
- 배포 시 .js 파일 생성 필요

**Python hooks (v0.3.0) 장점**:
- 의존성 zero (표준 라이브러리)
- 빌드 과정 불필요 (Python 스크립트 직접 실행)
- uv로 Python 3.13 자동 설치
- 단일 파일 배포 (moai_hooks.py 하나만)

---

## Environment (환경 및 전제조건)

### Prerequisites

**필수 요구사항**:
- **Python**: 3.13+ (uv가 자동 설치)
- **uv**: 0.2.0+ (Python 패키지 관리자)
- **Claude Code**: Latest version (Hook 시스템 지원)

**실행 예시**:
```bash
# uv run으로 직접 실행
uv run --python 3.13 .claude/hooks/moai_hooks.py SessionStart

# stdin으로 JSON payload 전달
echo '{"event":"SessionStart","cwd":"/path/to/project"}' | \
  uv run --python 3.13 .claude/hooks/moai_hooks.py SessionStart
```

### System Requirements

| 항목 | 최소 요구사항 | 권장 요구사항 |
|-----|-------------|-------------|
| Python | 3.13+ (uv 자동 설치) | 3.13+ |
| 실행 시간 | < 500ms (SessionStart) | < 100ms (대부분) |
| 메모리 | < 50MB | < 30MB |
| 파일 크기 | ≤ 600 LOC | ≤ 500 LOC |

### Claude Code Hook Events (9개)

Claude Code는 다음 9가지 이벤트를 지원합니다:

1. **SessionStart**: Claude Code 세션 시작 시
2. **SessionEnd**: Claude Code 세션 종료 시
3. **PreToolUse**: 도구 사용 전
4. **PostToolUse**: 도구 사용 후
5. **UserPromptSubmit**: 사용자 프롬프트 제출 시
6. **Notification**: 알림 발생 시
7. **Stop**: 세션 중단 시
8. **SubagentStop**: 서브에이전트 중단 시
9. **PreCompact**: 컨텍스트 압축 전

---

## Assumptions (가정)

1. **Python 3.13 자동 설치**: uv가 Python 3.13을 자동으로 다운로드하여 사용할 수 있다고 가정합니다
2. **표준 라이브러리 호환성**: Python 표준 라이브러리만 사용하므로 모든 환경에서 호환됩니다
3. **Git 리포지토리**: 대부분의 프로젝트는 Git 리포지토리를 사용한다고 가정 (Git 없으면 gracefully skip)
4. **파일 시스템 접근**: `.moai/specs/` 디렉토리 읽기 권한이 있다고 가정합니다
5. **stdin/stdout 통신**: Claude Code는 JSON 형식 stdin/stdout 통신을 지원한다고 가정합니다
6. **타임아웃**: 각 hook은 2초 이내에 완료되어야 한다고 가정 (Claude Code 제약)

---

## Requirements (EARS 방식)

### Ubiquitous Requirements (기본 요구사항)

- 시스템은 PEP 723 형식 inline metadata를 포함해야 한다
- 시스템은 외부 의존성 없이 Python 표준 라이브러리만 사용해야 한다
- 시스템은 600 LOC 이하 단일 파일로 구성되어야 한다
- 시스템은 9개 hook 이벤트를 지원해야 한다 (SessionStart, SessionEnd, PreToolUse, PostToolUse, UserPromptSubmit, Notification, Stop, SubagentStop, PreCompact)
- 시스템은 stdin으로 JSON payload를 받아야 한다
- 시스템은 stdout으로 JSON result를 출력해야 한다
- 시스템은 shebang `#!/usr/bin/env python3`를 포함해야 한다
- 시스템은 실행 가능해야 한다 (chmod +x)
- 시스템은 20개 언어를 감지할 수 있어야 한다
- 시스템은 SPEC 카운트를 계산해야 한다 (.moai/specs/ 스캔)
- 시스템은 Git 정보를 수집해야 한다 (branch, commit, changes)
- 시스템은 JIT Context Retrieval을 지원해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN SessionStart 이벤트가 발생하면, 시스템은 Git 정보 + SPEC 카운트 정보를 출력해야 한다
- WHEN UserPromptSubmit 이벤트가 발생하면, 시스템은 JIT Context Retrieval을 수행해야 한다
- WHEN PreCompact 이벤트가 발생하면, 시스템은 세션 요약을 생성해야 한다
- WHEN stdin으로 JSON payload를 받으면, 시스템은 해당 hook handler를 호출해야 한다
- WHEN Git 리포지토리가 없으면, 시스템은 Git 정보를 생략해야 한다
- WHEN SPEC 디렉토리가 없으면, 시스템은 "0/0 (0%)" 카운트를 표시해야 한다
- WHEN 언어 감지가 실패하면, 시스템은 "Unknown Language"를 반환해야 한다
- WHEN JSON parsing 오류 발생 시, 시스템은 오류 메시지를 stderr로 출력하고 exit code 1로 종료해야 한다
- WHEN 타임아웃(2초) 초과 시, 시스템은 강제 종료해야 한다
- WHEN /alfred:1-spec 명령어를 감지하면, 시스템은 spec-metadata.md를 참조해야 한다
- WHEN /alfred:2-build 명령어를 감지하면, 시스템은 development-guide.md를 참조해야 한다
- WHEN 테스트 관련 도구를 감지하면, 시스템은 tests/ 디렉토리를 참조해야 한다

### State-driven Requirements (상태 기반)

- WHILE 20개 언어 중 하나라면, 시스템은 해당 언어 감지 정보를 반환해야 한다
- WHILE SPEC 카운트 계산 중이면, 시스템은 `.moai/specs/` 디렉토리를 스캔해야 한다
- WHILE Git 명령 실행 중이면, 시스템은 타임아웃(2초)을 적용해야 한다
- WHILE JIT context retrieval 중이면, 시스템은 명령어 패턴에 따라 문서를 로드해야 한다 (/alfred:1-spec이면 spec-metadata.md)
- WHILE debug 모드가 활성화되면, 시스템은 상세 로그를 stderr로 출력해야 한다
- WHILE SessionStart 정보 생성 중이면, 시스템은 Git/SPEC/언어 정보를 수집해야 한다

### Optional Features (선택적 기능)

- WHERE DEBUG 환경 변수가 설정되면, 시스템은 상세 로그를 출력할 수 있다
- WHERE Git 리포지토리가 없으면, 시스템은 Git 정보를 생략할 수 있다
- WHERE .moai/specs/ 디렉토리가 없으면, 시스템은 SPEC 카운트를 "0/0"로 표시할 수 있다
- WHERE timeout 설정이 있으면, 시스템은 사용자 정의 timeout을 적용할 수 있다
- WHERE 언어 감지가 실패하면, 시스템은 "Unknown Language"를 반환할 수 있다
- WHERE 컨텍스트 압축이 필요하면, 시스템은 요약 힌트를 제공할 수 있다

### Constraints (제약사항)

- IF JSON parsing 오류 발생 시, 시스템은 exit code 1로 종료해야 한다 (오류 메시지는 stderr)
- IF timeout 초과 시, 시스템은 강제 종료해야 한다 (2초)
- IF Git 명령 실패 시, 시스템은 Git 정보를 생략해야 한다 (gracefully degrade)
- 파일 크기는 600 LOC를 초과하지 않아야 한다
- 외부 pip 패키지를 사용하지 않아야 한다
- 실행 시간은 SessionStart를 제외하고 100ms를 초과하지 않아야 한다
- SessionStart 실행 시간은 500ms를 초과하지 않아야 한다
- 메모리 사용량은 50MB를 초과하지 않아야 한다
- Python 3.13 이상 버전이 필요하다
- 모든 hook handler는 HookResult 타입을 반환해야 한다
- JSON 출력은 indent=2로 포맷팅해야 한다

---

## Specifications (상세 명세)

### 1. PEP 723 Header

모든 moai_hooks.py 파일은 다음 헤더로 시작해야 합니다:

```python
#!/usr/bin/env python3
# /// script
# requires-python = ">=3.13"
# dependencies = []
# ///
"""
MoAI-ADK Claude Code Hooks (Self-contained)

600 LOC self-contained Python script for Claude Code hook system.
Zero external dependencies - uses only Python standard library.

Usage:
    uv run --python 3.13 .claude/hooks/moai_hooks.py {event}

Events:
    SessionStart, SessionEnd, PreToolUse, PostToolUse,
    UserPromptSubmit, Notification, Stop, SubagentStop, PreCompact

Input: JSON payload via stdin
Output: JSON result via stdout
"""
```

### 2. Core Data Structures

```python
from typing import TypedDict, Literal, Optional
from dataclasses import dataclass
import json
import sys
import os
import subprocess
from pathlib import Path

# Hook Event Types
HookEvent = Literal[
    "SessionStart",
    "SessionEnd",
    "PreToolUse",
    "PostToolUse",
    "UserPromptSubmit",
    "Notification",
    "Stop",
    "SubagentStop",
    "PreCompact"
]

# Hook Payload
class HookPayload(TypedDict, total=False):
    event: str
    cwd: str
    toolName: Optional[str]
    toolArgs: Optional[dict]
    userPrompt: Optional[str]
    notificationMessage: Optional[str]

# Hook Result
@dataclass
class HookResult:
    message: Optional[str] = None
    blocked: bool = False
    suggestions: list[str] = None
    contextFiles: list[str] = None

    def to_json(self) -> str:
        return json.dumps({
            "message": self.message,
            "blocked": self.blocked,
            "suggestions": self.suggestions or [],
            "contextFiles": self.contextFiles or []
        }, indent=2)
```

### 3. Language Detection (20개 언어)

```python
LANGUAGE_PATTERNS = {
    "python": ["pyproject.toml", "setup.py", "requirements.txt", "*.py"],
    "typescript": ["tsconfig.json", "*.ts", "*.tsx"],
    "javascript": ["package.json", "*.js", "*.jsx"],
    "java": ["pom.xml", "build.gradle", "*.java"],
    "go": ["go.mod", "go.sum", "*.go"],
    "rust": ["Cargo.toml", "Cargo.lock", "*.rs"],
    "dart": ["pubspec.yaml", "*.dart"],
    "swift": ["Package.swift", "*.swift"],
    "kotlin": ["build.gradle.kts", "*.kt", "*.kts"],
    "csharp": ["*.csproj", "*.sln", "*.cs"],
    "php": ["composer.json", "*.php"],
    "ruby": ["Gemfile", "Gemfile.lock", "*.rb"],
    "elixir": ["mix.exs", "*.ex", "*.exs"],
    "scala": ["build.sbt", "*.scala"],
    "clojure": ["project.clj", "deps.edn", "*.clj"],
    "haskell": ["stack.yaml", "*.cabal", "*.hs"],
    "cpp": ["CMakeLists.txt", "*.cpp", "*.hpp"],
    "c": ["Makefile", "*.c", "*.h"],
    "shell": ["*.sh", "*.bash"],
    "lua": ["*.lua"],
}

def detect_language(cwd: str) -> str:
    """Detect project language from file patterns"""
    project_path = Path(cwd)

    for language, patterns in LANGUAGE_PATTERNS.items():
        for pattern in patterns:
            if "*" in pattern:
                if list(project_path.rglob(pattern)):
                    return language
            else:
                if (project_path / pattern).exists():
                    return language

    return "Unknown Language"
```

### 4. Git Information Collection

```python
def get_git_info(cwd: str) -> dict:
    """Collect Git repository information with 2s timeout"""
    try:
        # Check if Git repository
        result = subprocess.run(
            ["git", "rev-parse", "--git-dir"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2
        )

        if result.returncode != 0:
            return {}

        # Get current branch
        branch = subprocess.run(
            ["git", "branch", "--show-current"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2
        ).stdout.strip()

        # Get latest commit
        commit = subprocess.run(
            ["git", "rev-parse", "--short", "HEAD"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2
        ).stdout.strip()

        # Get change status
        status = subprocess.run(
            ["git", "status", "--short"],
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=2
        ).stdout.strip()

        changes = len(status.split("\n")) if status else 0

        return {
            "branch": branch,
            "commit": commit,
            "changes": changes
        }

    except (subprocess.TimeoutExpired, FileNotFoundError):
        return {}
```

### 5. SPEC Count Calculation

```python
def count_specs(cwd: str) -> dict:
    """Count SPEC files in .moai/specs/ directory"""
    specs_dir = Path(cwd) / ".moai" / "specs"

    if not specs_dir.exists():
        return {"completed": 0, "total": 0, "percentage": 0}

    total = 0
    completed = 0

    for spec_dir in specs_dir.iterdir():
        if spec_dir.is_dir() and spec_dir.name.startswith("SPEC-"):
            total += 1
            spec_file = spec_dir / "spec.md"
            if spec_file.exists():
                content = spec_file.read_text(encoding="utf-8")
                if "status: completed" in content:
                    completed += 1

    percentage = int(completed / total * 100) if total > 0 else 0

    return {
        "completed": completed,
        "total": total,
        "percentage": percentage
    }
```

### 6. JIT Context Retrieval

```python
def get_jit_context(user_prompt: str, cwd: str) -> list[str]:
    """Just-in-Time context retrieval based on command patterns"""
    context_files = []

    # Pattern matching for commands
    if "/alfred:1-spec" in user_prompt:
        context_files.append(".moai/memory/spec-metadata.md")

    if "/alfred:2-build" in user_prompt:
        context_files.append(".moai/memory/development-guide.md")

    if any(word in user_prompt.lower() for word in ["test", "pytest", "jest"]):
        tests_dir = Path(cwd) / "tests"
        if tests_dir.exists():
            context_files.append("tests/")

    # Filter existing files
    existing_files = []
    for file_path in context_files:
        full_path = Path(cwd) / file_path
        if full_path.exists():
            existing_files.append(file_path)

    return existing_files
```

### 7. Hook Handlers

```python
def handle_session_start(payload: HookPayload) -> HookResult:
    """Handle SessionStart event"""
    cwd = payload.get("cwd", ".")

    # Collect information
    language = detect_language(cwd)
    git_info = get_git_info(cwd)
    spec_count = count_specs(cwd)

    # Build message
    parts = []
    parts.append(f"🚀 MoAI-ADK Session Started")
    parts.append(f"Language: {language}")

    if git_info:
        parts.append(f"Git: {git_info['branch']} @ {git_info['commit']}")
        if git_info['changes'] > 0:
            parts.append(f"Changes: {git_info['changes']} files")

    if spec_count['total'] > 0:
        parts.append(
            f"SPECs: {spec_count['completed']}/{spec_count['total']} "
            f"({spec_count['percentage']}%)"
        )

    return HookResult(message="\n".join(parts))

def handle_user_prompt_submit(payload: HookPayload) -> HookResult:
    """Handle UserPromptSubmit event with JIT context"""
    cwd = payload.get("cwd", ".")
    user_prompt = payload.get("userPrompt", "")

    context_files = get_jit_context(user_prompt, cwd)

    if context_files:
        return HookResult(
            contextFiles=context_files,
            message=f"📚 Loaded {len(context_files)} context file(s)"
        )

    return HookResult()

def handle_pre_compact(payload: HookPayload) -> HookResult:
    """Handle PreCompact event"""
    return HookResult(
        message="💡 Tip: Use `/clear` or `/new` to start fresh session",
        suggestions=[
            "Summarize current session decisions",
            "Save important context to .moai/memory/",
            "Continue with clean context"
        ]
    )
```

### 8. Main Entry Point

```python
def main():
    """Main entry point"""
    try:
        # Parse command line arguments
        if len(sys.argv) < 2:
            print("Usage: moai_hooks.py {event}", file=sys.stderr)
            sys.exit(1)

        event = sys.argv[1]

        # Read JSON payload from stdin
        payload_json = sys.stdin.read()
        payload: HookPayload = json.loads(payload_json) if payload_json else {}

        # Route to appropriate handler
        handlers = {
            "SessionStart": handle_session_start,
            "UserPromptSubmit": handle_user_prompt_submit,
            "PreCompact": handle_pre_compact,
            # Add other handlers as needed
        }

        handler = handlers.get(event)
        if not handler:
            # Default: no-op for unimplemented events
            result = HookResult()
        else:
            result = handler(payload)

        # Output JSON result
        print(result.to_json())
        sys.exit(0)

    except json.JSONDecodeError as e:
        print(f"JSON parse error: {e}", file=sys.stderr)
        sys.exit(1)

    except Exception as e:
        print(f"Hook execution error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
```

---

## Traceability (추적성)

- **SPEC ID**: @SPEC:HOOKS-002
- **Depends on**: PY314-001 ✅
- **TAG 체인**: @SPEC:HOOKS-002 → @TEST:HOOKS-002 → @CODE:HOOKS-002
- **구현 위치**: `src/moai_adk/templates/.claude/hooks/moai_hooks.py`
- **테스트 위치**: `tests/unit/test_moai_hooks.py`

---

## Test Strategy (테스트 전략)

### Unit Tests

1. **Language Detection Tests** (20개)
   - 각 언어별 패턴 매칭 검증
   - Unknown language 처리

2. **Git Info Tests** (5개)
   - Git repository 정보 수집
   - Non-git directory 처리
   - Timeout 처리

3. **SPEC Count Tests** (4개)
   - SPEC 파일 카운팅
   - Empty directory 처리
   - Completed status 필터링

4. **JIT Context Tests** (6개)
   - Command pattern matching
   - File existence 검증
   - Context file 리스트 반환

5. **Hook Handler Tests** (9개)
   - 각 이벤트별 handler 동작 검증
   - JSON 입출력 검증

### Integration Tests

1. **End-to-End Hook Execution**
   - stdin → stdout 전체 플로우
   - Error handling

2. **Performance Tests**
   - 실행 시간 측정 (< 100ms 목표)
   - 메모리 사용량 측정 (< 50MB)

### Test Coverage Target

- **목표**: 85% 이상
- **핵심 로직**: 95% 이상 (language detection, git info, spec count)
- **Error handling**: 80% 이상

---

## Performance Requirements (성능 요구사항)

| Hook Event | 최대 실행시간 | 메모리 |
|-----------|-------------|--------|
| SessionStart | 500ms | 50MB |
| UserPromptSubmit | 100ms | 30MB |
| PreCompact | 100ms | 30MB |
| 기타 이벤트 | 50ms | 20MB |

---

## Security Considerations (보안 고려사항)

1. **Command Injection 방지**: subprocess 호출 시 shell=False 사용
2. **Path Traversal 방지**: pathlib.Path 사용, resolve() 검증
3. **Timeout 적용**: 모든 subprocess 호출에 2초 timeout
4. **Error Message 노출 최소화**: 민감한 경로 정보 제외
5. **JSON Parsing**: 안전한 json.loads 사용

---

## Migration Notes (마이그레이션 노트)

### TypeScript → Python 전환 체크리스트

- [x] npm dependencies 제거
- [x] tsup 빌드 프로세스 제거
- [x] Pure Python 구현 (stdlib only)
- [x] PEP 723 metadata 추가
- [x] uv 실행 방식 적용
- [x] 9개 hook events 구현
- [x] 20개 언어 감지 유지
- [x] JIT Context Retrieval 유지
- [x] Git 정보 수집 유지
- [x] SPEC 카운트 계산 유지

---

## 참고 문서

- **PEP 723**: Inline script metadata
- **Claude Code Hooks**: Hook system documentation
- **Python 3.13**: New features and improvements
- **development-guide.md**: MoAI-ADK 개발 가이드
- **spec-metadata.md**: SPEC 메타데이터 표준
