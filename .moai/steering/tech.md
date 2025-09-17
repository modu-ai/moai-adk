# 기술 스택 @TECH:MOAI-ADK

> **@TECH:MOAI-ADK** "Python 3.11+ 코어 스택 기반 모던 개발 환경"

## 🐍 핵심 기술 스택

### 언어 및 런타임
- **Python**: 3.11+ (권장: 3.12+)
- **Node.js**: 18+ (Claude Code 지원용)
- **Shell**: Bash/Zsh (크로스 플랫폼 지원)

### 프로젝트 관리
- **의존성 관리**: poetry (추천) / pip + requirements.txt
- **환경 관리**: pyenv / conda
- **패키징**: setuptools + pyproject.toml

## 📦 의존성 구조

### Core Dependencies (필수)
```toml
[tool.poetry.dependencies]
python = "^3.11"
click = "^8.1.0"           # CLI 인터페이스
colorama = "^0.4.6"        # 터미널 색상
rich = "^13.0.0"           # 예쁜 출력
toml = "^0.10.0"           # 설정 파일 파싱
pydantic = "^2.0.0"        # 데이터 검증
jinja2 = "^3.1.0"          # 템플릿 엔진
gitpython = "^3.1.0"       # Git 조작
```

### Development Tools (개발용)
```toml
[tool.poetry.group.dev.dependencies]
# 테스팅
pytest = "^8.0.0"
pytest-cov = "^4.0.0"
pytest-asyncio = "^0.21.0"
pytest-mock = "^3.10.0"

# 코드 품질
ruff = "^0.1.0"            # 린팅 + 포맷팅 (black + flake8 대체)
mypy = "^1.7.0"            # 타입 체킹
bandit = "^1.7.0"          # 보안 검사

# 개발 도구
pre-commit = "^3.0.0"      # Git hooks
tox = "^4.0.0"             # 다중 환경 테스트
```

### Optional Dependencies (선택적)
```toml
[tool.poetry.group.docs.dependencies]
mkdocs = "^1.5.0"          # 문서 생성
mkdocs-material = "^9.0.0" # Material 테마

[tool.poetry.group.performance.dependencies]
uvloop = "^0.19.0"         # 빠른 이벤트 루프 (Unix only)
```

## 🔧 개발 도구 체인

### 1. 코드 품질 도구

#### Ruff (린팅 + 포맷팅)
```toml
[tool.ruff]
target-version = "py311"
line-length = 88
select = [
    "E",   # pycodestyle errors
    "W",   # pycodestyle warnings
    "F",   # pyflakes
    "I",   # isort
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
    "UP",  # pyupgrade
]

[tool.ruff.per-file-ignores]
"tests/*" = ["S101"]  # assert 사용 허용
```

#### MyPy (타입 체킹)
```toml
[tool.mypy]
python_version = "3.11"
strict = true
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

#### Bandit (보안 검사)
```toml
[tool.bandit]
exclude_dirs = ["tests", "docs"]
skips = ["B101"]  # assert_used
```

### 2. 테스트 프레임워크

#### Pytest 설정
```toml
[tool.pytest.ini_options]
minversion = "8.0"
addopts = [
    "--strict-markers",
    "--strict-config",
    "--cov=src",
    "--cov-report=term-missing:skip-covered",
    "--cov-report=html",
    "--cov-report=xml",
    "--cov-fail-under=85",
]
testpaths = ["tests"]
markers = [
    "unit: 단위 테스트",
    "integration: 통합 테스트",
    "e2e: E2E 테스트",
    "slow: 느린 테스트",
]
```

#### Coverage 목표
- **최소 커버리지**: 85%
- **핵심 모듈**: 95%+
- **에이전트**: 90%+
- **CLI**: 80%+

### 3. Pre-commit Hooks

#### .pre-commit-config.yaml
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-toml
      - id: check-json

  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix, --exit-non-zero-on-fix]
      - id: ruff-format

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.7.0
    hooks:
      - id: mypy
        additional_dependencies: [types-all]

  - repo: https://github.com/PyCQA/bandit
    rev: 1.7.5
    hooks:
      - id: bandit
        args: [-c, pyproject.toml]
```

## 🗄️ 데이터 저장 및 관리

### 1. 설정 파일 형식
- **TOML**: 주 설정 파일 (.moai/config.json → config.toml 마이그레이션 예정)
- **JSON**: TAG 인덱스, 상태 파일
- **YAML**: CI/CD 설정, pre-commit
- **Markdown**: 문서, SPEC

### 2. 데이터 구조
```python
# 설정 데이터 모델
from pydantic import BaseModel

class ProjectConfig(BaseModel):
    name: str
    version: str
    description: str
    language: str = "ko"

class ConstitutionPrinciple(BaseModel):
    enabled: bool
    description: str
    parameters: Dict[str, Any]

class TagSystem(BaseModel):
    version: str = "16-core"
    categories: Dict[str, List[str]]
    naming_convention: str
```

### 3. 파일 시스템 구조
```
.moai/
├── config.toml          # 메인 설정 (TOML)
├── indexes/
│   ├── tags.json        # TAG 인덱스 (JSON)
│   ├── state.json       # 프로젝트 상태 (JSON)
│   └── version.json     # 버전 정보 (JSON)
└── memory/
    ├── constitution.md  # 헌법 (Markdown)
    └── *.md            # 메모리 문서들 (Markdown)
```

## 🔄 CI/CD 파이프라인

### 1. GitHub Actions 워크플로우

#### 기본 테스트 (.github/workflows/test.yml)
```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ["3.11", "3.12"]

    steps:
      - uses: actions/checkout@v4
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: ${{ matrix.python-version }}

      - name: Install poetry
        run: pip install poetry

      - name: Install dependencies
        run: poetry install

      - name: Run tests
        run: poetry run pytest

      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

#### 코드 품질 (.github/workflows/quality.yml)
```yaml
name: Code Quality
on: [push, pull_request]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: "3.11"

      - name: Run ruff
        run: |
          pip install ruff
          ruff check .
          ruff format --check .

      - name: Run mypy
        run: |
          pip install mypy
          mypy src/

      - name: Run bandit
        run: |
          pip install bandit
          bandit -r src/
```

### 2. 배포 전략
- **개발**: develop 브랜치 → 자동 테스트
- **스테이징**: release/* 브랜치 → 통합 테스트
- **프로덕션**: main 브랜치 → 수동 승인 후 배포

## 🔒 보안 및 성능

### 1. 보안 도구
```toml
[tool.poetry.group.security.dependencies]
bandit = "^1.7.0"          # 보안 취약점 스캔
safety = "^2.3.0"          # 의존성 취약점 체크
semgrep = "^1.45.0"        # 정적 보안 분석
```

### 2. 성능 모니터링
```toml
[tool.poetry.group.performance.dependencies]
memory-profiler = "^0.61.0"  # 메모리 사용량 분석
line-profiler = "^4.1.0"     # 라인별 성능 분석
py-spy = "^0.3.0"            # 프로파일링 도구
```

### 3. 로깅 및 모니터링
```python
import structlog

# 구조화된 로깅 설정
structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.StackInfoRenderer(),
        structlog.dev.ConsoleRenderer()
    ],
    wrapper_class=structlog.make_filtering_bound_logger(logging.INFO),
    logger_factory=structlog.WriteLoggerFactory(),
    cache_logger_on_first_use=False,
)
```

## 📈 성능 목표 및 벤치마크

### 1. 성능 목표
- **CLI 명령어 응답**: < 2초
- **파일 파싱**: < 100ms (1000 LOC 기준)
- **TAG 인덱싱**: < 500ms (10000 태그 기준)
- **메모리 사용량**: < 500MB

### 2. 벤치마크 스크립트
```python
# scripts/benchmark.py
import time
import psutil
from typing import Callable

def benchmark(func: Callable, iterations: int = 100) -> dict:
    """성능 벤치마크 실행"""
    times = []
    process = psutil.Process()

    for _ in range(iterations):
        start_time = time.perf_counter()
        memory_before = process.memory_info().rss

        func()

        end_time = time.perf_counter()
        memory_after = process.memory_info().rss

        times.append(end_time - start_time)

    return {
        "avg_time": sum(times) / len(times),
        "min_time": min(times),
        "max_time": max(times),
        "memory_delta": memory_after - memory_before
    }
```

## 🔄 업그레이드 전략

### 1. 단계적 도구 업그레이드

#### 현재 → 목표
```toml
# 현재 스택
pytest = "^7.0"     # → 8.0
black = "^22.0"     # → ruff (통합)
flake8 = "^5.0"     # → ruff (통합)

# 목표 스택
pytest = "^8.0"
ruff = "^0.1.0"     # black + flake8 대체
pre-commit = "^3.0"
tox = "^4.0"
```

### 2. Python 버전 지원
- **최소 지원**: Python 3.11
- **권장**: Python 3.12+
- **향후 계획**: Python 3.13 (2024년 10월 출시 예정)

### 3. 의존성 관리 전략
```bash
# 정기적 업데이트 (월 1회)
poetry update

# 보안 패치 (수시)
safety check
bandit -r src/

# 성능 벤치마크 (릴리스 전)
python scripts/benchmark.py
```

## 🔧 개발 환경 설정

### 1. 로컬 개발 환경
```bash
# 1. Poetry 설치
curl -sSL https://install.python-poetry.org | python3 -

# 2. 프로젝트 의존성 설치
poetry install

# 3. Pre-commit hooks 설정
poetry run pre-commit install

# 4. 개발 서버 실행
poetry run python -m src.cli
```

### 2. IDE 설정 (VS Code)
```json
// .vscode/settings.json
{
    "python.defaultInterpreterPath": ".venv/bin/python",
    "python.linting.enabled": true,
    "python.linting.ruffEnabled": true,
    "python.formatting.provider": "none",
    "python.testing.pytestEnabled": true,
    "python.testing.pytestArgs": ["tests/"],
    "files.associations": {
        "*.toml": "toml"
    }
}
```

### 3. 추천 VS Code 확장
```json
// .vscode/extensions.json
{
    "recommendations": [
        "ms-python.python",
        "charliermarsh.ruff",
        "ms-python.mypy-type-checker",
        "tamasfe.even-better-toml",
        "yzhang.markdown-all-in-one"
    ]
}
```

---

> **@TECH:MOAI-ADK** 태그를 통해 이 기술 결정사항들이 프로젝트 전체에 일관되게 적용됩니다.
>
> **모든 도구는 Constitution 5원칙을 지원하며, 개발자 경험과 코드 품질을 동시에 보장합니다.**