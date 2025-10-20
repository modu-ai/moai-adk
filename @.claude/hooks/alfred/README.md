# Alfred Hooks System

**Event-Driven Context Management for MoAI-ADK**

Alfred Hooks는 Claude Code의 이벤트 시스템과 통합되어 프로젝트 컨텍스트를 자동으로 관리하고, 위험한 작업 전에 checkpoint를 생성하며, JIT (Just-in-Time) 문서 로딩을 제공합니다.

---

## 📐 Architecture

### Modular Design (9 Files, ≤284 LOC each)

```
.claude/hooks/alfred/
├── alfred_hooks.py          # Main entry point (CLI router)
├── core/                    # Core business logic
│   ├── __init__.py         # Type definitions (HookPayload, HookResult)
│   ├── project.py          # Language detection, Git info, SPEC counting
│   ├── context.py          # JIT retrieval, workflow context
│   ├── checkpoint.py       # Event-driven checkpoint creation
│   └── tags.py             # TAG search, verification, caching
└── handlers/                # Event handlers
    ├── __init__.py         # Handler exports
    ├── session.py          # SessionStart, SessionEnd
    ├── user.py             # UserPromptSubmit
    ├── tool.py             # PreToolUse, PostToolUse
    └── notification.py     # Notification, Stop, SubagentStop
```

### Design Principles

- **Single Responsibility**: 각 모듈은 하나의 명확한 책임
- **Separation of Concerns**: core (비즈니스 로직) vs handlers (이벤트 처리)
- **CODE-FIRST**: 중간 캐시 없이 코드 직접 스캔 (mtime 기반 무효화)
- **Context Engineering**: JIT Retrieval로 초기 컨텍스트 부담 최소화

---

## 🎯 Core Modules

### `core/project.py` (284 LOC)

**프로젝트 메타데이터 및 언어 감지**

```python
# Public API
detect_language(cwd: str) -> str
get_project_language(cwd: str) -> str
get_git_info(cwd: str) -> dict[str, Any]
count_specs(cwd: str) -> dict[str, int]
```

**Features**:
- 20개 언어 자동 감지 (Python, TypeScript, Java, Go, Rust, etc.)
- `.moai/config.json` 우선, fallback to auto-detection
- Git 정보 조회 (branch, commit, changes)
- SPEC 진행도 계산 (total, completed, percentage)

### `core/context.py` (110 LOC)

**JIT Context Retrieval 및 워크플로우 관리**

```python
# Public API
get_jit_context(prompt: str, cwd: str) -> list[str]
save_phase_context(phase: str, data: Any, ttl: int = 600)
load_phase_context(phase: str, ttl: int = 600) -> Any | None
clear_workflow_context()
```

**Features**:
- 프롬프트 분석 기반 문서 자동 추천
  - `/alfred:1-spec` → `spec-metadata.md`
  - `/alfred:2-build` → `development-guide.md`
- 워크플로우 단계별 컨텍스트 캐싱 (TTL 10분)
- Anthropic Context Engineering 원칙 준수

### `core/checkpoint.py` (244 LOC)

**Event-Driven Checkpoint 자동화**

```python
# Public API
detect_risky_operation(tool: str, args: dict, cwd: str) -> tuple[bool, str]
create_checkpoint(cwd: str, operation: str) -> str
log_checkpoint(cwd: str, branch: str, description: str)
list_checkpoints(cwd: str, max_count: int = 10) -> list[dict]
```

**Features**:
- 위험한 작업 자동 감지:
  - Bash: `rm -rf`, `git merge`, `git reset --hard`
  - Edit/Write: `CLAUDE.md`, `config.json`
  - MultiEdit: ≥10 files
- Git checkpoint 자동 생성: `checkpoint/before-{operation}-{timestamp}`
- checkpoint 이력 관리 및 복구 가이드

### `core/tags.py` (244 LOC)

**CODE-FIRST TAG 시스템**

```python
# Public API
search_tags(pattern: str, scope: list[str], cache_ttl: int = 60) -> list[dict]
verify_tag_chain(tag_id: str) -> dict[str, Any]
find_all_tags_by_type(tag_type: str) -> dict[str, list[str]]
suggest_tag_reuse(keyword: str) -> list[str]
get_library_version(library: str, cache_ttl: int = 86400) -> str | None
set_library_version(library: str, version: str)
```

**Features**:
- ripgrep 기반 TAG 검색 (JSON 출력 파싱)
- mtime 기반 캐시 무효화 (CODE-FIRST 보장)
- TAG 체인 검증 (@SPEC → @TEST → @CODE 완전성 확인)
- 라이브러리 버전 캐싱 (TTL 24시간)

---

## 🎬 Event Handlers

### `handlers/session.py`

**SessionStart, SessionEnd 핸들러**

- **SessionStart**: 프로젝트 정보 표시
  - 언어, Git 상태, SPEC 진행도, 최근 checkpoint
  - `systemMessage` 필드로 사용자에게 직접 표시
- **SessionEnd**: 정리 작업 (stub)

### `handlers/user.py`

**UserPromptSubmit 핸들러**

- JIT Context 추천 문서 리스트 반환
- 사용자 프롬프트 패턴 분석 및 관련 문서 로드

### `handlers/tool.py`

**PreToolUse, PostToolUse 핸들러**

- **PreToolUse**: 위험한 작업 감지 시 자동 checkpoint 생성
- **PostToolUse**: 후처리 작업 (stub)

### `handlers/notification.py`

**Notification, Stop, SubagentStop 핸들러**

- 기본 구현 (stub, 향후 확장 가능)

---

## 🧪 Testing

### Test Suite

```bash
# Run all tests
uv run pytest tests/unit/test_alfred_hooks_*.py -v --no-cov

# Run specific module tests
uv run pytest tests/unit/test_alfred_hooks_core_tags.py -v
uv run pytest tests/unit/test_alfred_hooks_core_context.py -v
uv run pytest tests/unit/test_alfred_hooks_core_project.py -v
```

### Test Coverage (18 tests)

- ✅ **tags.py**: 7 tests (캐시, TAG 검증, 버전 관리)
- ✅ **context.py**: 5 tests (JIT, 워크플로우 컨텍스트)
- ✅ **project.py**: 6 tests (언어 감지, Git, SPEC 카운트)

### Test Structure

```python
# Dynamic module loading for isolated testing
def _load_{module}_module(module_name: str):
    repo_root = Path(__file__).resolve().parents[2]
    hooks_dir = repo_root / ".claude" / "hooks" / "alfred"
    sys.path.insert(0, str(hooks_dir))
    
    module_path = hooks_dir / "core" / "{module}.py"
    spec = importlib.util.spec_from_file_location(module_name, module_path)
    # ...
```

---

## 🔄 Migration from moai_hooks.py

### Before (Monolithic)

- **1 file**: 1233 LOC
- **Issues**: 
  - 모든 기능이 하나의 파일에 집중
  - 테스트 어려움, 유지보수 복잡
  - 책임 분리 불명확

### After (Modular)

- **9 files**: ≤284 LOC each
- **Benefits**:
  - 명확한 책임 분리 (SRP)
  - 독립적인 모듈 테스트 가능
  - 확장 용이, 유지보수 간편
  - Context Engineering 원칙 준수

### Breaking Changes

**없음** - 외부 API는 동일하게 유지됩니다.

---

## 📚 References

### Internal Documents

- **CLAUDE.md**: MoAI-ADK 사용자 가이드
- **.moai/memory/development-guide.md**: SPEC-First TDD 워크플로우
- **.moai/memory/spec-metadata.md**: SPEC 메타데이터 표준

### External Resources

- [Claude Code Hooks Documentation](https://docs.claude.com/en/docs/claude-code)
- [Anthropic Context Engineering](https://docs.anthropic.com/claude/docs/context-engineering)

---

**Last Updated**: 2025-10-16  
**Author**: @Alfred (MoAI-ADK SuperAgent)
