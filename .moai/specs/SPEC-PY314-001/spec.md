---
id: PY314-001
version: 0.1.0
status: completed
created: 2025-10-13
updated: 2025-10-14
author: @Goos
priority: critical
category: feature
labels:
  - python-3.13
  - build-system
  - foundation
  - migration
blocks:
  - CLI-001
  - CORE-GIT-001
  - CORE-TEMPLATE-001
  - CORE-PROJECT-001
  - HOOKS-001
  - HOOKS-002
scope:
  packages:
    - src/moai_adk/
  files:
    - pyproject.toml
    - src/moai_adk/cli/main.py
    - src/moai_adk/core/template/config.py
    - src/moai_adk/core/template/processor.py
---

# @SPEC:PY314-001: Python 3.13 Foundation & Build System

## HISTORY

### v0.1.0 (2025-10-14)
- **GREEN**: TDD 구현 완료 (cli/main.py, template/config.py, template/processor.py)
- **REFACTOR**: ruff 린터, mypy strict mode 통과
- **TEST**: ConfigManager 8/8 통과, TemplateProcessor 11/11 통과
- **CHANGED**: Python 3.14 → 3.13 (3.14는 Alpha 단계, 3.13 Stable 사용)
- **COVERAGE**: 30% (Foundation only, CLI-001 구현 후 85% 목표)
- **AUTHOR**: @Goos

### v0.0.1 (2025-10-13)
- **INITIAL**: Python 3.14 기반 프로젝트 구조 및 uv 빌드 시스템 명세 작성
- **AUTHOR**: @Goos
- **SCOPE**: TypeScript → Python 전환의 기초 인프라 구축
- **CONTEXT**: MoAI-ADK v0.3.0 전환 프로젝트의 첫 번째 SPEC
- **REASON**: 단일 언어 기반 유지보수 간소화, Python 생태계 활용, uv의 빠른 성능 활용

---

## Overview

MoAI-ADK를 TypeScript에서 Python 3.14로 전환합니다. uv 빌드 시스템을 사용하여 현대적인 Python 프로젝트 구조를 구축하고, 의존성을 12개에서 6개로 단순화합니다. 이는 유지보수성을 향상시키고, Python 에코시스템의 강력한 도구들을 활용할 수 있게 합니다.

### 핵심 가치

1. **Zero npm Dependencies**: npm, Node.js 의존성 완전 제거
2. **자동 Python 설치**: uv가 Python 3.14를 자동으로 다운로드하고 설치
3. **빠른 실행 속도**: uv의 Rust 기반 빌드 시스템으로 5-10배 빠른 패키지 설치
4. **단순한 의존성**: 핵심 6개 패키지만 사용 (click, rich, questionary, GitPython, jinja2, pyyaml)

### 전환 근거

**TypeScript (v0.2.x) 문제점**:
- npm 의존성 관리 복잡도 (node_modules 12MB+)
- Bun/Node.js 런타임 의존성
- tsup, Biome 등 빌드 도구 체인 복잡도
- 타입 선언 파일 관리 오버헤드

**Python (v0.3.0) 장점**:
- 표준 라이브러리가 강력함 (subprocess, pathlib, asyncio)
- uv로 빌드 시간 5초 이내 (TypeScript tsc 대비 50% 단축)
- PyPI 배포 단순화 (uv publish 한 줄)
- 언어 학습자가 더 많음 (Python > TypeScript)

---

## Environment

### Prerequisites

**필수 소프트웨어**:
- **Python**: 3.14+ (uv가 자동 설치하므로 사전 설치 불필요)
- **uv**: 0.2.0+ (패키지 매니저 및 빌드 도구)
  ```bash
  # macOS/Linux
  curl -LsSf https://astral.sh/uv/install.sh | sh

  # Windows
  powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
  ```

**선택 도구**:
- Git 2.40+ (버전 관리)
- VS Code + Python extension (개발 환경)

### System Requirements

| 항목 | 최소 요구사항 | 권장 요구사항 |
|-----|-------------|-------------|
| OS | Windows 10, macOS 12, Linux (Ubuntu 20.04+) | macOS 14+, Ubuntu 22.04+ |
| RAM | 2GB | 8GB+ |
| Disk | 500MB (프로젝트 + 의존성) | 2GB+ |
| Python | 3.14+ (uv 자동 설치) | 3.14+ |
| Network | 인터넷 연결 (패키지 다운로드) | - |

### 기존 시스템 (TypeScript v0.2.31)

**디렉토리 구조**:
```
moai-adk/
├── moai-adk-ts/
│   ├── src/
│   │   ├── cli/
│   │   ├── core/
│   │   └── index.ts
│   ├── package.json
│   ├── tsconfig.json
│   └── tsup.config.ts
├── templates/
│   ├── .claude/
│   └── .moai/
```

**의존성 (12개)**:
- commander, chalk, ora, fs-extra, simple-git (5개 런타임)
- @biomejs/biome, vitest, tsx, typescript (4개 개발)
- esbuild, tsup (2개 빌드)

---

## Assumptions

1. **Python 3.14 자동 설치**: uv는 Python 3.14가 없으면 자동으로 다운로드하고 설치한다
2. **uv 전역 설치**: 사용자가 uv를 시스템에 전역으로 설치했다
3. **템플릿 호환성**: 기존 `.claude/`, `.moai/` 템플릿 구조는 Python에서도 그대로 사용 가능하다
4. **PyPI 배포 가능**: `moai-adk` 패키지명이 PyPI에서 사용 가능하다
5. **기존 사용자 마이그레이션**: TypeScript v0.2.x 사용자는 `npm uninstall -g moai-adk && uvx moai-adk` 명령으로 마이그레이션한다

---

## Requirements (EARS 구문)

### Ubiquitous Requirements (필수 기능)

- 시스템은 Python 3.14 기반 프로젝트 구조를 제공해야 한다
- 시스템은 uv를 사용한 빌드 및 의존성 관리를 지원해야 한다
- 시스템은 pyproject.toml에 모든 메타데이터를 정의해야 한다 (PEP 621 준수)
- 시스템은 src/ 레이아웃(src/moai_adk/)을 사용해야 한다
- 시스템은 6개 핵심 의존성만 사용해야 한다 (click, rich, questionary, GitPython, jinja2, pyyaml)
- 시스템은 `moai` 명령어를 제공해야 한다 (진입점: moai_adk.__main__:main)

### Event-driven Requirements (이벤트 기반)

- WHEN `uvx moai-adk init .` 명령이 실행되면, 시스템은 Python 3.14를 자동 설치하고 프로젝트를 초기화해야 한다
- WHEN `uv build` 명령이 실행되면, 시스템은 5초 이내에 wheel 패키지를 생성해야 한다
- WHEN `uv pip install -e .` 명령이 실행되면, 시스템은 editable 모드로 설치해야 한다
- WHEN 의존성 충돌이 발생하면, 시스템은 uv.lock 파일로 버전을 고정해야 한다
- WHEN Python 3.14가 없을 때, 시스템은 자동으로 다운로드하고 설치해야 한다

### State-driven Requirements (상태 기반)

- WHILE 개발 모드일 때, 시스템은 dev 의존성(pytest, pytest-asyncio, pytest-cov, mypy, ruff)을 설치해야 한다
- WHILE 프로덕션 빌드일 때, 시스템은 core 의존성 6개만 포함해야 한다
- WHILE CI/CD 실행 중일 때, 시스템은 uv cache를 활용하여 빌드 시간을 단축해야 한다

### Optional Features (선택적 기능)

- WHERE 개발자 환경이면, 시스템은 pre-commit hooks를 설정할 수 있다
- WHERE GitHub Actions가 구성되면, 시스템은 자동 PyPI 배포를 수행할 수 있다
- WHERE Docker 환경이면, 시스템은 multi-stage 빌드를 지원할 수 있다

### Constraints (제약사항)

- IF Python 버전이 3.14 미만이면, 시스템은 설치를 중단하고 에러 메시지를 표시해야 한다
- 패키지명은 `moai-adk`여야 한다 (npm: `moai-adk`, PyPI: `moai-adk`)
- 진입점 명령어는 `moai`여야 한다 (이전과 동일)
- src/ 디렉토리 구조를 준수해야 한다 (PEP 420 namespace package 방지)
- 빌드 시간은 5초를 초과하지 않아야 한다
- 패키지 크기는 5MB를 초과하지 않아야 한다
- 의존성은 6개를 초과하지 않아야 한다 (개발 의존성 제외)

---

## Technical Design

### Architecture

**Python 프로젝트 구조 (PEP 420, PEP 621 준수)**:

```
moai-adk/
├── pyproject.toml              # PEP 621 메타데이터 + 빌드 설정
├── uv.lock                     # 의존성 잠금 파일 (uv 자동 생성)
├── README.md                   # Python 프로젝트 README
├── LICENSE                     # MIT 라이선스
├── .gitignore                  # Python .venv/, __pycache__/ 제외
│
├── moai-adk-py/
│   └── src/
│       └── moai_adk/
│           ├── __init__.py         # 패키지 초기화
│           ├── __main__.py         # CLI 진입점
│           │
│           ├── cli/                # Click 기반 CLI
│           │   ├── __init__.py
│           │   ├── main.py         # @click.group()
│           │   └── commands/
│           │       ├── init.py
│           │       ├── doctor.py
│           │       ├── status.py
│           │       └── restore.py
│           │
│           ├── core/               # 핵심 모듈
│           │   ├── __init__.py
│           │   ├── git/            # GitPython 래퍼
│           │   │   ├── __init__.py
│           │   │   ├── manager.py
│           │   │   └── branch.py
│           │   ├── template/       # Jinja2 템플릿
│           │   │   ├── __init__.py
│           │   │   ├── processor.py
│           │   │   └── config.py
│           │   └── project/        # 프로젝트 초기화
│           │       ├── __init__.py
│           │       ├── init.py
│           │       ├── detector.py  # 20개 언어 감지
│           │       └── doctor.py
│           │
│           └── hooks/              # Hooks 시스템
│               ├── __init__.py
│               ├── runtime.py      # HooksRuntime (asyncio)
│               └── events.py       # HookEvent enum
│
├── tests/
│   ├── __init__.py
│   ├── unit/
│   │   ├── test_cli.py
│   │   ├── test_git.py
│   │   └── test_template.py
│   ├── integration/
│   │   ├── test_init_workflow.py
│   │   └── test_doctor_workflow.py
│   └── fixtures/
│       ├── sample_project/
│       └── test_templates/
│
└── templates/
    ├── .claude/
    │   ├── CLAUDE.md
    │   ├── settings.json           # Claude Code 설정
    │   └── commands/
    │       ├── 1-spec.md
    │       ├── 2-build.md
    │       └── 3-sync.md
    └── .moai/
        ├── config.json
        ├── project/
        │   ├── product.md
        │   ├── structure.md
        │   └── tech.md
        ├── specs/
        └── memory/
            ├── development-guide.md
            └── spec-metadata.md
```

### Data Models

**pyproject.toml (완전한 버전)**:

```toml
[project]
name = "moai-adk"
version = "0.3.0"
description = "MoAI Agentic Development Kit - SPEC-First TDD Framework with Alfred SuperAgent"
authors = [
    { name = "MoAI Team", email = "dev@modu.ai" }
]
requires-python = ">=3.14"
license = { text = "MIT" }
readme = "README.md"
keywords = [
    "ai",
    "tdd",
    "spec",
    "development",
    "automation",
    "alfred",
    "claude-code",
    "spec-first"
]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Programming Language :: Python :: 3.14",
    "Programming Language :: Python :: 3 :: Only",
    "Topic :: Software Development :: Code Generators",
    "Topic :: Software Development :: Testing",
    "Operating System :: OS Independent",
]

# 핵심 의존성 (6개)
dependencies = [
    "click>=8.1.0",         # CLI 프레임워크
    "rich>=13.0.0",         # 터미널 출력
    "questionary>=2.0.0",   # 대화형 프롬프트
    "gitpython>=3.1.0",     # Git 조작
    "jinja2>=3.1.0",        # 템플릿 엔진
    "pyyaml>=6.0",          # YAML 파싱
]

[project.optional-dependencies]
# 개발 의존성 (5개)
dev = [
    "pytest>=8.0.0",
    "pytest-asyncio>=0.23.0",
    "pytest-cov>=4.1.0",
    "mypy>=1.8.0",
    "ruff>=0.1.0",
]

# 테스트 전용
test = [
    "pytest>=8.0.0",
    "pytest-asyncio>=0.23.0",
    "pytest-cov>=4.1.0",
]

# 진입점
[project.scripts]
moai = "moai_adk.__main__:main"

[project.urls]
Homepage = "https://github.com/modu-ai/moai-adk"
Documentation = "https://moai-adk.dev"
Repository = "https://github.com/modu-ai/moai-adk"
Issues = "https://github.com/modu-ai/moai-adk/issues"
Changelog = "https://github.com/modu-ai/moai-adk/blob/main/CHANGELOG.md"

# 빌드 시스템 (uv + hatchling)
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = ["moai-adk-py/src/moai_adk"]

# Ruff 설정
[tool.ruff]
line-length = 100
target-version = "py314"
select = ["E", "F", "I", "N", "W", "UP", "B", "A", "C4", "SIM"]
ignore = ["E501"]  # 라인 길이는 100으로 설정했으므로 E501 무시

[tool.ruff.format]
quote-style = "double"
indent-style = "space"

# MyPy 설정
[tool.mypy]
python_version = "3.14"
strict = true
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true

# Pytest 설정
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
addopts = [
    "--cov=moai_adk",
    "--cov-report=term-missing",
    "--cov-report=html",
    "--cov-fail-under=85",
]
```

### TypeScript → Python 의존성 매핑

| TypeScript 패키지 | Python 패키지 | 용도 | 변경 사항 |
|------------------|---------------|------|----------|
| commander | click | CLI 프레임워크 | @click.command() 데코레이터 방식 |
| simple-git | gitpython | Git 조작 | Repo() 객체 지향 API |
| ejs | jinja2 | 템플릿 엔진 | {% %} 구문, 더 강력한 필터 |
| chalk | rich | 터미널 출력 | Rich console, 표, 프로그레스 바 |
| ora | rich.progress | 로딩 스피너 | Progress context manager |
| inquirer | questionary | 대화형 프롬프트 | async 지원, 더 풍부한 UI |
| fs-extra | pathlib (표준 라이브러리) | 파일 시스템 | Path 객체, 내장 메서드 |
| execa | subprocess (표준 라이브러리) | 외부 프로세스 | run(), Popen() |
| zod | pydantic (선택적) | 데이터 검증 | 타입 힌트 + mypy로 대체 가능 |
| js-yaml | pyyaml | YAML 파싱 | safe_load(), safe_dump() |

### 제거된 의존성 (npm 생태계)

**더 이상 필요하지 않음**:
- @biomejs/biome → ruff (Python 린터, 100배 빠름)
- vitest → pytest (Python 표준 테스트 프레임워크)
- tsx → Python 인터프리터 (빌드 불필요)
- typescript → Python (정적 타이핑 내장)
- tsup → uv build (빌드 도구 통합)
- esbuild → Python 인터프리터 (번들링 불필요)

---

## Implementation Strategy

### Phase 1: 프로젝트 구조 생성 (1일)

**목표**: Python 프로젝트 골격 구축

**작업**:
1. `moai-adk-py/src/moai_adk/` 디렉토리 생성
2. `pyproject.toml` 작성 (위 Data Models 참조)
3. `__init__.py`, `__main__.py` 파일 생성
4. `.gitignore` 업데이트 (`__pycache__/`, `.venv/`, `*.pyc`, `uv.lock` 추가)
5. `README.md` Python 버전으로 업데이트

**검증 기준**:
- [ ] `moai-adk-py/src/moai_adk/__init__.py` 파일 존재
- [ ] `pyproject.toml` 파일 유효성 검증: `uv run python -m build --check`
- [ ] 디렉토리 구조가 위 Architecture와 일치

### Phase 2: 의존성 설치 및 환경 설정 (1일)

**목표**: uv 기반 의존성 관리 및 개발 환경 구축

**작업**:
1. `uv venv` 실행하여 가상환경 생성
2. `uv pip install -e ".[dev]"` 실행하여 editable 모드로 설치
3. `uv.lock` 파일 생성 확인
4. pre-commit hooks 설정 (ruff, mypy)
5. VS Code `.vscode/settings.json` 설정

**검증 기준**:
- [ ] `.venv/` 디렉토리 생성됨
- [ ] `uv.lock` 파일 존재
- [ ] `moai --version` 명령어 실행 가능 (0.3.0 출력)
- [ ] `ruff check moai-adk-py/src/` 통과
- [ ] `mypy moai-adk-py/src/` 통과

### Phase 3: CLI 진입점 구현 (1일)

**목표**: `moai` 명령어 기본 구조 구현

**작업**:
1. `cli/main.py` 작성 (@click.group() 정의)
2. ASCII 로고 함수 구현 (TypeScript 버전 100% 재현)
3. `--version`, `--help` 옵션 구현
4. Rich console 설정
5. 기본 에러 핸들링 구현

**검증 기준**:
- [ ] `moai` 실행 시 로고 출력
- [ ] `moai --version` → "MoAI-ADK v0.3.0" 출력
- [ ] `moai --help` → 명령어 목록 출력
- [ ] Rich 색상 출력 동작

### Phase 4: 템플릿 호환성 검증 (1일)

**목표**: 기존 `.claude/`, `.moai/` 템플릿이 Python에서 정상 동작하는지 확인

**작업**:
1. `templates/` 디렉토리 복사
2. Jinja2로 템플릿 렌더링 테스트
3. CLAUDE.md, settings.json 파싱 테스트
4. config.json YAML 파싱 테스트
5. 언어 감지 로직 20개 언어 테스트

**검증 기준**:
- [ ] 20개 언어 모두 감지됨 (Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin, etc.)
- [ ] 템플릿 렌더링 정상 동작
- [ ] config.json 읽기/쓰기 정상 동작

---

## Testing Strategy

### Unit Tests

**테스트 대상**:
- CLI 명령어 파싱 (`test_cli.py`)
- Git 래퍼 함수 (`test_git.py`)
- 템플릿 렌더링 (`test_template.py`)
- 언어 감지 (`test_detector.py`)
- Config 읽기/쓰기 (`test_config.py`)

**테스트 도구**:
- pytest + pytest-cov
- pytest-asyncio (비동기 테스트)
- unittest.mock (모킹)

**커버리지 목표**: ≥85%

### Integration Tests

**테스트 시나리오**:
1. `moai init .` → 프로젝트 초기화 → `.moai/` 생성 확인
2. `moai doctor` → 환경 검증 → 모든 체크 통과 확인
3. `moai status` → 프로젝트 상태 → JSON 형식 출력 확인
4. `moai restore` → 백업 복원 → 파일 복구 확인

**테스트 환경**:
- GitHub Actions (Ubuntu, macOS, Windows)
- Python 3.14 (latest)

### End-to-End Tests

**E2E 시나리오**:
1. **신규 프로젝트 생성**:
   ```bash
   uvx moai-adk init my-project
   cd my-project
   moai doctor
   # 예상: 모든 체크 통과
   ```

2. **기존 프로젝트 복원**:
   ```bash
   cd existing-project
   moai restore --timestamp 2025-10-13T10:00:00
   # 예상: 백업 복원 완료
   ```

3. **크로스 플랫폼 테스트**:
   - Windows 10: PowerShell에서 `moai init .` 실행
   - macOS 14: Terminal에서 `moai init .` 실행
   - Ubuntu 22.04: Bash에서 `moai init .` 실행

---

## Dependencies

### Internal Dependencies (MoAI-ADK 내부)

- **SPEC-CLI-001**: CLI 명령어 구현 (Phase 3에서 필요)
- **SPEC-CORE-PROJECT-001**: 프로젝트 초기화 로직 (Phase 4에서 필요)
- **SPEC-CORE-TEMPLATE-001**: Jinja2 템플릿 엔진 (Phase 4에서 필요)
- **SPEC-CORE-GIT-001**: GitPython 래퍼 (나중에 필요)

### External Dependencies (Python 패키지)

**런타임 (6개)**:
1. **click**: CLI 프레임워크, 명령어 파싱
2. **rich**: 터미널 출력, 색상, 표, 프로그레스 바
3. **questionary**: 대화형 프롬프트 (moai init 시 사용)
4. **gitpython**: Git 조작, 브랜치 관리
5. **jinja2**: 템플릿 엔진, 변수 치환
6. **pyyaml**: YAML 파싱, config.json 읽기/쓰기

**개발 도구 (5개)**:
1. **pytest**: 테스트 프레임워크
2. **pytest-asyncio**: 비동기 테스트
3. **pytest-cov**: 커버리지 측정
4. **mypy**: 타입 체크
5. **ruff**: 린터 + 포매터

---

## Migration Path

### From TypeScript v0.2.31 to Python v0.3.0

**기존 사용자 마이그레이션 가이드**:

#### Step 1: TypeScript 버전 제거
```bash
# npm 전역 패키지 제거
npm uninstall -g moai-adk

# 또는 Bun 사용 시
bun remove -g moai-adk
```

#### Step 2: Python 버전 설치
```bash
# uv 설치 (없으면)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Python 버전 설치 (한 줄로 실행)
uvx moai-adk init .

# uv가 자동으로:
# 1. Python 3.14 다운로드 및 설치
# 2. moai-adk 패키지 설치
# 3. 프로젝트 초기화
```

#### Step 3: 기존 프로젝트 검증
```bash
# 프로젝트 상태 확인
moai status

# 환경 검증
moai doctor

# 버전 확인
moai --version
# 출력: MoAI-ADK v0.3.0
```

### Breaking Changes (하위 호환성 깨짐)

**명령어 변경 없음**: 모든 명령어는 동일하게 유지됩니다.
```bash
# TypeScript v0.2.x
moai init .
moai doctor

# Python v0.3.0
moai init .
moai doctor
# 명령어 동일!
```

**설정 파일 호환**: `.moai/config.json`은 그대로 사용 가능합니다.

**템플릿 호환**: `.claude/`, `.moai/` 디렉토리 구조 동일합니다.

### Backward Compatibility (하위 호환성)

**유지되는 것**:
- ✅ 명령어 (init, doctor, status, restore)
- ✅ 설정 파일 (.moai/config.json)
- ✅ 템플릿 구조 (.claude/, .moai/)
- ✅ @TAG 시스템 (@SPEC:ID, @TEST:ID, @CODE:ID)
- ✅ SPEC 문서 형식 (YAML Front Matter + EARS)

**변경되는 것**:
- ❌ npm 설치 방식 → uvx 설치 방식
- ❌ Node.js/Bun 런타임 → Python 인터프리터
- ❌ package.json → pyproject.toml
- ❌ TypeScript 소스 → Python 소스

---

## Risks and Mitigation

### Risk 1: uv 설치 실패 (확률: 낮음)

**Impact**: 사용자가 uv를 설치하지 못하면 moai-adk를 사용할 수 없음

**Probability**: 낮음 (5%) - uv 설치 스크립트는 매우 안정적

**Mitigation**:
1. README.md에 uv 설치 가이드 명시
2. 설치 실패 시 대체 방법 안내 (pip install)
3. `moai doctor` 명령어로 uv 설치 확인

### Risk 2: Python 3.14 호환성 문제 (확률: 중간)

**Impact**: Python 3.14가 일부 시스템에서 설치되지 않을 수 있음

**Probability**: 중간 (20%) - Python 3.14는 비교적 최신 버전

**Mitigation**:
1. Python 3.12+ 지원으로 하향 조정 (필요 시)
2. uv가 자동으로 Python 3.14를 다운로드하고 설치
3. 설치 실패 시 명확한 에러 메시지 및 해결 방법 안내

### Risk 3: 기존 사용자 혼란 (확률: 높음)

**Impact**: TypeScript → Python 전환으로 기존 사용자가 혼란스러울 수 있음

**Probability**: 높음 (40%) - 주요 버전 변경 (v0.2 → v0.3)

**Mitigation**:
1. CHANGELOG.md에 마이그레이션 가이드 명시
2. 블로그 포스트 및 마이그레이션 비디오 제작
3. GitHub Discussions에서 Q&A 지원
4. v0.2.x는 6개월간 유지보수 계속

### Risk 4: 의존성 충돌 (확률: 낮음)

**Impact**: click, rich 등이 다른 패키지와 버전 충돌 가능

**Probability**: 낮음 (10%) - 6개 의존성만 사용, 모두 안정적

**Mitigation**:
1. uv.lock 파일로 버전 고정
2. 의존성 버전 범위 최소화 (click>=8.1.0 대신 click==8.1.7)
3. Dependabot으로 자동 업데이트

### Risk 5: 성능 저하 (확률: 매우 낮음)

**Impact**: Python이 TypeScript보다 느릴 수 있음

**Probability**: 매우 낮음 (5%) - CLI 도구는 I/O 바운드

**Mitigation**:
1. uv는 Rust 기반으로 빠름
2. asyncio로 비동기 처리
3. 벤치마크 테스트로 성능 검증 (목표: 명령어 실행 < 1초)

---

## Traceability (@TAG)

- **SPEC**: @SPEC:PY314-001
- **TEST**: tests/unit/test_foundation.py (@TEST:PY314-001)
- **CODE**: moai-adk-py/src/moai_adk/ (@CODE:PY314-001)
- **DOC**: README.md, CHANGELOG.md (@DOC:PY314-001)

**TAG 체인**:
```
@SPEC:PY314-001
  → @TEST:PY314-001 (프로젝트 구조 테스트)
  → @CODE:PY314-001 (pyproject.toml, __init__.py)
  → @DOC:PY314-001 (README.md)
```

**Blocks (차단하는 SPEC)**:
- CLI-001: CLI 명령어는 PY314-001 이후 구현 가능
- CORE-GIT-001: Git 래퍼는 PY314-001 이후 구현 가능
- CORE-TEMPLATE-001: 템플릿 엔진은 PY314-001 이후 구현 가능
- CORE-PROJECT-001: 프로젝트 초기화는 PY314-001 이후 구현 가능
- HOOKS-001: Hooks 시스템은 PY314-001 이후 구현 가능
- HOOKS-002: moai_hooks.py는 PY314-001 이후 구현 가능

---

## References

### Python 공식 문서
- [PEP 621 - Python Package Metadata](https://peps.python.org/pep-0621/)
- [PEP 420 - Implicit Namespace Packages](https://peps.python.org/pep-0420/)
- [Python 3.14 Release Notes](https://www.python.org/downloads/release/python-3140/)

### 빌드 도구
- [uv Documentation](https://github.com/astral-sh/uv)
- [Hatchling Build Backend](https://hatch.pypa.io/)

### 의존성 문서
- [Click Documentation](https://click.palletsprojects.com/)
- [Rich Documentation](https://rich.readthedocs.io/)
- [GitPython Documentation](https://gitpython.readthedocs.io/)
- [Jinja2 Documentation](https://jinja.palletsprojects.com/)

### MoAI-ADK 내부 문서
- `.moai/memory/development-guide.md` - TRUST 5원칙, TDD 워크플로우
- `.moai/memory/spec-metadata.md` - SPEC 메타데이터 표준
- `CLAUDE.md` - Alfred SuperAgent 오케스트레이션

---

## Appendix

### Glossary

- **uv**: Rust 기반 Python 패키지 매니저, pip/poetry 대체품
- **pyproject.toml**: PEP 621 기반 Python 프로젝트 메타데이터 파일
- **uvx**: uv의 "npx" 같은 도구, 패키지 설치 없이 실행
- **wheel**: Python 패키지 배포 형식 (.whl 파일)
- **editable install**: 개발 중 소스 코드 직접 수정하며 사용하는 설치 방식
- **PEP 621**: Python 패키지 메타데이터 표준화 제안

### Code Examples

#### pyproject.toml 기본 구조
```toml
[project]
name = "moai-adk"
version = "0.3.0"
requires-python = ">=3.14"
dependencies = ["click>=8.1.0"]

[project.scripts]
moai = "moai_adk.__main__:main"

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

#### CLI 진입점 (__main__.py)
```python
import click
from rich.console import Console

console = Console()

@click.group()
@click.version_option(version="0.3.0")
def cli():
    """MoAI-ADK v0.3.0"""
    console.print("[cyan]▶◀ MoAI-ADK v0.3.0[/cyan]")

def main():
    cli()

if __name__ == "__main__":
    main()
```

### Configuration Examples

#### uv 설치 스크립트
```bash
# macOS/Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows PowerShell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### 개발 환경 설정
```bash
# 가상환경 생성
uv venv

# 활성화 (macOS/Linux)
source .venv/bin/activate

# 활성화 (Windows)
.\.venv\Scripts\activate

# editable 설치
uv pip install -e ".[dev]"

# 테스트 실행
pytest tests/ -v --cov=moai_adk
```

---

**작성자**: @Goos
**작성일**: 2025-10-13
**버전**: v0.0.1
**상태**: draft
