# MoAI-ADK Technical Document

## 기술 스택 개요 @TECH:OVERVIEW-001

**MoAI-ADK**는 **Python 3.11+** 기반의 Claude Code 네이티브 패키지로, 현대적인 개발 도구 체인과 AI 기반 자동화를 결합한 기술 스택을 채택합니다.

## 언어 & 런타임

### 🐍 **Primary: Python 3.11+** @TECH:STACK-PYTHON-001

```python
# 선택 이유
- Claude Code 생태계와의 네이티브 호환성
- 풍부한 AI/ML 라이브러리 생태계
- 개발 도구 및 정적 분석 도구 성숙도
- 크로스 플랫폼 지원 (Windows/macOS/Linux)

# 의존성 관리
- Package Manager: pip + setuptools
- Virtual Environment: venv (표준 라이브러리)
- Lock File: requirements.txt + requirements-dev.txt
- Build System: setuptools + wheel
```

### 🌐 **Secondary: Multi-Language Support** @TECH:STACK-MULTI-001

```yaml
언어별 지원 전략:
  JavaScript/TypeScript:
    런타임: Node.js ≥18
    패키지매니저: npm, pnpm
    테스트: jest, vitest
    린터: eslint, prettier

  Go:
    버전: Go ≥1.21
    빌드: go build
    테스트: go test
    린터: golangci-lint

  Java/Kotlin:
    JVM: OpenJDK ≥17
    빌드: Gradle ≥8.0
    테스트: JUnit 5, TestNG
    린터: detekt (Kotlin), SpotBugs (Java)

  .NET:
    런타임: .NET ≥8.0
    빌드: dotnet CLI
    테스트: xUnit, NUnit
    린터: Roslyn Analyzers

  Rust:
    버전: Rust ≥1.70
    빌드: cargo
    테스트: cargo test
    린터: clippy, rustfmt
```

### 📦 **패키지 배포 전략**

```bash
# Primary Package (Python)
pip install moai-adk

# Language-specific Extensions (향후)
npm install @moai/adk-js
go install github.com/moai/adk-go
dotnet add package MoAI.ADK
cargo add moai-adk
```

## 핵심 프레임워크 & 라이브러리

### 🎯 **CLI 프레임워크**

```python
# Click - 명령어 인터페이스
import click

@click.group()
@click.version_option()
def moai():
    """MoAI-ADK CLI Interface"""
    pass

# Rich - 터미널 UI
from rich.console import Console
from rich.progress import Progress
from rich.table import Table

# Typer (선택적) - 타입 힌트 기반 CLI
from typer import Typer
```

### 🏗️ **아키텍처 프레임워크**

```python
# Pydantic - 데이터 검증 및 시리얼라이제이션
from pydantic import BaseModel, Field, validator

class SpecDocument(BaseModel):
    id: str = Field(..., regex=r"SPEC-\d{3}")
    title: str = Field(..., min_length=1, max_length=100)
    tags: List[str] = Field(default_factory=list)

    @validator('tags')
    def validate_tags(cls, v):
        return [tag for tag in v if tag.startswith('@')]

# Dependency Injection
from dependency_injector import containers, providers

class ApplicationContainer(containers.DeclarativeContainer):
    config = providers.Configuration()

    # Agent Services
    spec_builder = providers.Singleton(SpecBuilderAgent)
    code_builder = providers.Singleton(CodeBuilderAgent)
```

### 🔄 **비동기 처리**

```python
# asyncio - 비동기 에이전트 실행
import asyncio
from concurrent.futures import ThreadPoolExecutor

async def run_agents_parallel(*agents):
    """에이전트 병렬 실행"""
    tasks = [agent.execute() for agent in agents]
    results = await asyncio.gather(*tasks, return_exceptions=True)
    return results

# aiofiles - 비동기 파일 I/O
import aiofiles

async def read_spec_async(spec_path: str) -> str:
    async with aiofiles.open(spec_path, 'r') as f:
        return await f.read()
```

### 📁 **파일 시스템 관리**

```python
# pathlib - 현대적 경로 처리
from pathlib import Path

class MoaiPaths:
    def __init__(self, project_root: Path):
        self.root = project_root
        self.moai = self.root / ".moai"
        self.specs = self.moai / "specs"
        self.project = self.moai / "project"

    def ensure_structure(self):
        """필수 디렉토리 구조 생성"""
        for path in [self.moai, self.specs, self.project]:
            path.mkdir(parents=True, exist_ok=True)

# watchdog - 파일 변경 감지
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

class SpecFileHandler(FileSystemEventHandler):
    def on_modified(self, event):
        if event.src_path.endswith('.md'):
            # 태그 인덱스 업데이트 트리거
            self.update_tag_index(event.src_path)
```

## 테스트 전략 & 도구

### 🧪 **테스트 프레임워크**

```python
# pytest - 메인 테스트 프레임워크
import pytest
from pytest import fixture, mark, param

# pytest-asyncio - 비동기 테스트
@pytest.mark.asyncio
async def test_agent_parallel_execution():
    agents = [SpecBuilderAgent(), CodeBuilderAgent()]
    results = await run_agents_parallel(*agents)
    assert all(result.success for result in results)

# pytest-mock - 모킹
def test_git_integration(mocker):
    mock_git = mocker.patch('subprocess.run')
    git_manager = GitManager()
    git_manager.create_branch("feature/test")
    mock_git.assert_called_with(['git', 'checkout', '-b', 'feature/test'])

# pytest-cov - 커버리지
pytest --cov=moai_adk --cov-report=html tests/
```

### 📊 **테스트 전략 매트릭스**

```yaml
테스트 레벨:
  Unit Tests (80%):
    범위: 개별 함수, 클래스 메서드
    도구: pytest + unittest.mock
    실행: 로컬 개발 중 + CI/CD
    커버리지: ≥90%

  Integration Tests (15%):
    범위: 에이전트 간 협업, 외부 API 연동
    도구: pytest + docker-compose (필요시)
    실행: CI/CD + 릴리스 전
    커버리지: ≥80%

  E2E Tests (5%):
    범위: 전체 워크플로우 (/moai:0-project → /moai:3-sync)
    도구: pytest + 실제 Git 저장소
    실행: 릴리스 후보 검증
    커버리지: 주요 시나리오 100%

테스트 데이터:
  fixtures: tests/fixtures/ 디렉토리
  mock_data: JSON/YAML 형태 테스트 데이터
  temp_repos: 임시 Git 저장소 생성/정리
```

### 🎭 **모킹 전략**

```python
# Git 명령어 모킹
@pytest.fixture
def mock_git_commands(mocker):
    return {
        'status': mocker.patch('subprocess.run', return_value=CompletedProcess([], 0, "clean")),
        'commit': mocker.patch('subprocess.run', return_value=CompletedProcess([], 0, "")),
        'push': mocker.patch('subprocess.run', return_value=CompletedProcess([], 0, ""))
    }

# Claude Code API 모킹
@pytest.fixture
def mock_claude_code_api(mocker):
    mock_api = mocker.patch('moai_adk.integrations.claude_code.ClaudeCodeAPI')
    mock_api.get_project_context.return_value = {
        'language': 'python',
        'framework': 'click',
        'test_runner': 'pytest'
    }
    return mock_api

# 파일 시스템 모킹 (임시 디렉토리)
@pytest.fixture
def temp_moai_project(tmp_path):
    moai_paths = MoaiPaths(tmp_path)
    moai_paths.ensure_structure()
    return moai_paths
```

## 품질 관리 도구

### 🔍 **정적 분석 도구**

```python
# ruff - 통합 린터 (black + isort + flake8 + 기타)
[tool.ruff]
line-length = 88
target-version = "py311"
select = [
    "E",   # pycodestyle errors
    "W",   # pycodestyle warnings
    "F",   # pyflakes
    "I",   # isort
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
    "UP",  # pyupgrade
]
ignore = ["E501"]  # line too long (handled by formatter)

# mypy - 타입 체킹
[tool.mypy]
python_version = "3.11"
strict = true
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true

# bandit - 보안 취약점 스캔
bandit -r moai_adk/ -f json -o security-report.json
```

### 📏 **코드 품질 메트릭**

```yaml
품질 게이트:
  라인 수 제한:
    파일: ≤ 300 LOC
    함수: ≤ 50 LOC
    클래스: ≤ 500 LOC (데이터 클래스 예외)

  복잡도 제한:
    Cyclomatic Complexity: ≤ 10
    함수 매개변수: ≤ 5개
    중첩 레벨: ≤ 4단계

  중복 코드:
    허용 임계치: ≤ 5%
    검사 도구: ruff + sonarqube (향후)

  타입 힌트:
    커버리지: ≥ 95%
    엄격 모드: mypy strict=true
```

### 🔒 **보안 정책**

```python
# 시크릿 관리
import os
from pathlib import Path

class SecretManager:
    @staticmethod
    def get_github_token() -> str:
        """GitHub PAT 안전하게 로드"""
        token = os.getenv('GITHUB_TOKEN')
        if not token:
            # ~/.moai/secrets/github.token 파일에서 로드
            secret_file = Path.home() / '.moai' / 'secrets' / 'github.token'
            if secret_file.exists():
                token = secret_file.read_text().strip()

        if not token:
            raise ValueError("GitHub token not found")
        return token

# 로깅에서 민감정보 마스킹
import re
import logging

class SensitiveFormatter(logging.Formatter):
    def format(self, record):
        msg = super().format(record)
        # 토큰, 패스워드 등 마스킹
        msg = re.sub(r'token["\s]*[:=]["\s]*([a-zA-Z0-9]+)', r'token: ***redacted***', msg)
        msg = re.sub(r'password["\s]*[:=]["\s]*([^\s]+)', r'password: ***redacted***', msg)
        return msg
```

## 외부 시스템 연동

### 🐙 **Git 생태계**

```python
# GitPython - Git 저장소 조작
from git import Repo, GitCommandError

class GitManager:
    def __init__(self, repo_path: Path):
        self.repo = Repo(repo_path)

    def create_feature_branch(self, spec_id: str) -> str:
        """SPEC ID 기반 브랜치 생성"""
        branch_name = f"feature/{spec_id.lower()}"
        try:
            new_branch = self.repo.create_head(branch_name)
            new_branch.checkout()
            return branch_name
        except GitCommandError as e:
            raise GitOperationError(f"브랜치 생성 실패: {e}")

# GitHub API (PyGithub)
from github import Github, GithubException

class GitHubIntegration:
    def __init__(self, token: str):
        self.github = Github(token)

    def create_pull_request(self, repo_name: str, branch: str, spec: SpecDocument):
        """SPEC 기반 PR 생성"""
        repo = self.github.get_repo(repo_name)
        pr = repo.create_pull(
            title=f"[{spec.id}] {spec.title}",
            body=self._generate_pr_body(spec),
            head=branch,
            base="develop"
        )
        return pr.html_url
```

### 🤖 **AI CLI 도구 연동** (선택적)

```python
# Codex CLI Integration
import subprocess
import json
from typing import Optional

class CodexBridge:
    def __init__(self):
        self.available = self._check_codex_availability()

    def _check_codex_availability(self) -> bool:
        """Codex CLI 설치 및 인증 상태 확인"""
        try:
            result = subprocess.run(['codex', '--version'],
                                  capture_output=True, text=True)
            return result.returncode == 0
        except FileNotFoundError:
            return False

    async def brainstorm_solutions(self, problem: str) -> List[str]:
        """문제에 대한 해결책 브레인스토밍"""
        if not self.available:
            return []

        prompt = f"""
        다음 개발 문제에 대한 3가지 다른 접근 방법을 제시해주세요:
        {problem}

        각 접근법에 대해 장단점과 구현 복잡도를 포함해주세요.
        """

        cmd = ['codex', 'exec', '--model', 'gpt-4', '--prompt', prompt]
        result = subprocess.run(cmd, capture_output=True, text=True)

        if result.returncode == 0:
            return self._parse_brainstorm_response(result.stdout)
        return []

# Gemini CLI Integration
class GeminiBridge:
    def __init__(self):
        self.available = self._check_gemini_availability()

    async def analyze_code_quality(self, code_path: Path) -> dict:
        """코드 품질 분석 및 구조화된 리포트 생성"""
        if not self.available:
            return {}

        cmd = [
            'gemini', '-m', 'gemini-2.0-flash-exp',
            '-p', f'다음 코드의 품질을 TRUST 5원칙 관점에서 분석해주세요: {code_path}',
            '--output-format', 'json'
        ]

        result = subprocess.run(cmd, capture_output=True, text=True)
        if result.returncode == 0:
            return json.loads(result.stdout)
        return {}
```

### 📊 **모니터링 & 로깅**

```python
# 구조화된 로깅
import structlog
import json
from datetime import datetime

# structlog 설정
structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.JSONRenderer()
    ],
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    wrapper_class=structlog.stdlib.BoundLogger,
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger()

# 에이전트 실행 로깅
class AgentLogger:
    @staticmethod
    def log_agent_start(agent_name: str, command: str, spec_id: Optional[str] = None):
        logger.info("agent_started",
                   agent=agent_name,
                   command=command,
                   spec_id=spec_id,
                   timestamp=datetime.utcnow().isoformat())

    @staticmethod
    def log_agent_complete(agent_name: str, duration_ms: int, success: bool, **metrics):
        logger.info("agent_completed",
                   agent=agent_name,
                   duration_ms=duration_ms,
                   success=success,
                   metrics=metrics)

# 메트릭 수집
class MetricsCollector:
    def __init__(self):
        self.metrics = {}

    def record_command_execution(self, command: str, duration: float, success: bool):
        """명령어 실행 메트릭 기록"""
        key = f"command.{command}"
        if key not in self.metrics:
            self.metrics[key] = {"count": 0, "total_duration": 0, "success_count": 0}

        self.metrics[key]["count"] += 1
        self.metrics[key]["total_duration"] += duration
        if success:
            self.metrics[key]["success_count"] += 1

    def get_success_rate(self, command: str) -> float:
        """명령어 성공률 계산"""
        key = f"command.{command}"
        if key not in self.metrics:
            return 0.0

        data = self.metrics[key]
        return data["success_count"] / data["count"] if data["count"] > 0 else 0.0
```

## 빌드 & 배포 파이프라인

### 🏗️ **빌드 설정**

```toml
# pyproject.toml
[build-system]
requires = ["setuptools>=68.0", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "moai-adk"
version = "0.1.0"
description = "MoAI Agentic Development Kit for Claude Code"
authors = [{name = "MoAI Team", email = "team@moai.dev"}]
readme = "README.md"
license = {text = "MIT"}
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.11",
    "Programming Language :: Python :: 3.12",
]
dependencies = [
    "click>=8.1.0",
    "rich>=13.0.0",
    "pydantic>=2.0.0",
    "aiofiles>=23.0.0",
    "watchdog>=3.0.0",
    "gitpython>=3.1.0",
    "PyGithub>=1.59.0",
    "structlog>=23.0.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.4.0",
    "pytest-asyncio>=0.21.0",
    "pytest-mock>=3.11.0",
    "pytest-cov>=4.1.0",
    "ruff>=0.1.0",
    "mypy>=1.5.0",
    "bandit>=1.7.5",
]

[project.scripts]
moai = "moai_adk.cli:main"

[tool.setuptools.packages.find]
where = ["src"]
include = ["moai_adk*"]
```

### 🔄 **CI/CD 파이프라인**

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [develop, main]
  pull_request:
    branches: [develop]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ["3.11", "3.12"]

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v4
        with:
          python-version: ${{ matrix.python-version }}

      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -e .[dev]

      - name: Run linting
        run: |
          ruff check src/ tests/
          ruff format --check src/ tests/
          mypy src/

      - name: Run security scan
        run: bandit -r src/ -f json -o security-report.json

      - name: Run tests
        run: |
          pytest --cov=moai_adk --cov-report=xml --cov-report=html tests/

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.xml

  build:
    needs: test
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v4
        with:
          python-version: "3.11"

      - name: Build package
        run: |
          python -m pip install build
          python -m build

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: dist
          path: dist/

  release:
    if: github.ref == 'refs/heads/main'
    needs: [test, build]
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
      - name: Download artifacts
        uses: actions/download-artifact@v3
        with:
          name: dist
          path: dist/

      - name: Publish to PyPI
        uses: pypa/gh-action-pypi-publish@v1.8.10
        with:
          password: ${{ secrets.PYPI_API_TOKEN }}
```

### 🚀 **배포 전략**

```yaml
배포 채널:
  Development:
    트리거: develop 브랜치 푸시
    대상: TestPyPI (test.pypi.org)
    버전: 0.1.0.dev{timestamp}

  Staging:
    트리거: release/* 브랜치 생성
    대상: PyPI (staging tag)
    버전: 0.1.0-rc.{build_number}

  Production:
    트리거: main 브랜치 푸시 + Git tag
    대상: PyPI (latest tag)
    버전: 0.1.0 (Semantic Versioning)

환경별 설정:
  개발환경:
    claude_code_integration: mock
    git_operations: local_only
    external_ai_tools: disabled

  스테이징:
    claude_code_integration: sandbox
    git_operations: test_repo
    external_ai_tools: limited

  프로덕션:
    claude_code_integration: production
    git_operations: full
    external_ai_tools: opt_in
```

## 성능 & 확장성

### ⚡ **성능 최적화**

```python
# 비동기 에이전트 실행
async def execute_workflow_parallel(workflow_steps: List[WorkflowStep]):
    """워크플로우 단계별 병렬 실행"""
    results = []

    for step_group in workflow_steps:
        # 의존성 없는 단계들은 병렬 실행
        if step_group.can_run_parallel:
            tasks = [step.execute() for step in step_group.steps]
            step_results = await asyncio.gather(*tasks)
            results.extend(step_results)
        else:
            # 순차 실행이 필요한 단계들
            for step in step_group.steps:
                result = await step.execute()
                results.append(result)

    return results

# 캐싱 전략
from functools import lru_cache
import hashlib

class SpecCache:
    def __init__(self, cache_dir: Path):
        self.cache_dir = cache_dir
        self.cache_dir.mkdir(exist_ok=True)

    def _get_cache_key(self, spec_content: str) -> str:
        """SPEC 내용 기반 캐시 키 생성"""
        return hashlib.sha256(spec_content.encode()).hexdigest()[:16]

    @lru_cache(maxsize=128)
    def get_parsed_spec(self, spec_content: str) -> ParsedSpec:
        """파싱된 SPEC 캐시에서 조회/생성"""
        cache_key = self._get_cache_key(spec_content)
        cache_file = self.cache_dir / f"{cache_key}.json"

        if cache_file.exists():
            return ParsedSpec.parse_file(cache_file)

        # 캐시 미스 - 새로 파싱
        parsed = SpecParser.parse(spec_content)
        cache_file.write_text(parsed.json())
        return parsed
```

### 📈 **확장성 설계**

```python
# 플러그인 시스템
from abc import ABC, abstractmethod
from typing import Protocol

class AgentPlugin(Protocol):
    """에이전트 플러그인 인터페이스"""

    @property
    def name(self) -> str: ...

    @property
    def version(self) -> str: ...

    async def execute(self, context: ExecutionContext) -> AgentResult: ...

    def can_handle(self, command: str) -> bool: ...

class PluginRegistry:
    def __init__(self):
        self._plugins: Dict[str, AgentPlugin] = {}

    def register(self, plugin: AgentPlugin):
        """플러그인 등록"""
        self._plugins[plugin.name] = plugin

    def get_handler(self, command: str) -> Optional[AgentPlugin]:
        """명령어를 처리할 수 있는 플러그인 반환"""
        for plugin in self._plugins.values():
            if plugin.can_handle(command):
                return plugin
        return None

# 커스텀 에이전트 예시
class SecurityScannerAgent(AgentPlugin):
    @property
    def name(self) -> str:
        return "security-scanner"

    @property
    def version(self) -> str:
        return "1.0.0"

    def can_handle(self, command: str) -> bool:
        return command.startswith("/moai:security")

    async def execute(self, context: ExecutionContext) -> AgentResult:
        # 보안 스캔 로직 구현
        scan_results = await self._run_security_scan(context.project_path)
        return AgentResult(
            success=True,
            data=scan_results,
            message="보안 스캔 완료"
        )
```

## Legacy Context @DEBT:TECH-ANALYSIS-001

### 현재 기술 스택 상태 분석

**기존 기술 현황**:

- 주 언어: Python 3.11+ (Claude Code 네이티브)
- 프레임워크: Click (CLI), Pydantic (데이터 검증), structlog (로깅)
- 테스트: pytest 기반 (현재 커버리지 < 85%)
- 패키지 관리: pip + setuptools (모던 pyproject.toml 부분 적용)

### 기술 부채 및 개선 계획

**식별된 기술 부채**:

- @DEBT:DEPENDENCY-001: 의존성 관리 현대화 필요 (poetry/pipenv 고려)
- @DEBT:ASYNC-001: 비동기 처리 최적화 부족 (asyncio 활용 미흡)
- @DEBT:TYPING-001: 타입 힌트 커버리지 < 95%
- @DEBT:PERFORMANCE-001: 에이전트 간 통신 최적화 필요
- @DEBT:SECURITY-001: 시크릿 관리 체계 표준화 필요

**우선순위 기술 개선 계획**:

- @TODO:MODERN-PYTHON-001: pyproject.toml 완전 전환 및 ruff 통합
- @TODO:ASYNC-OPTIMIZATION-001: 에이전트 병렬 실행 최적화 (10개 → 20개)
- @TODO:TYPE-SAFETY-001: mypy strict 모드 적용 및 95% 타입 커버리지
- @TODO:MONITORING-001: 구조화된 로깅 및 메트릭 수집 강화
- @TODO:SECURITY-HARDENING-001: Secrets Manager 통합 및 보안 스캔 자동화

### Initial Migration Plan

**Phase 1 (즉시 필요)**:

- @TASK:PYPROJECT-001: pyproject.toml 완전 전환
- @TASK:RUFF-INTEGRATION-001: ruff로 통합 린팅 환경 구축
- @TASK:TEST-COVERAGE-001: pytest 커버리지 85% 달성

**Phase 2 (1개월 내)**:

- @TASK:ASYNC-REFACTOR-001: 비동기 처리 최적화
- @TASK:TYPE-ANNOTATIONS-001: 전체 코드베이스 타입 힌트 추가
- @TASK:CI-CD-001: GitHub Actions 파이프라인 고도화

**Phase 3 (3개월 내)**:

- @TASK:PERFORMANCE-001: 성능 벤치마킹 및 최적화
- @TASK:SECURITY-001: 보안 강화 및 감사 체계 구축
- @TASK:PACKAGING-001: PyPI 배포 자동화 완성

---

## Next Steps

이 tech.md 문서로 MoAI-ADK의 **기술적 설계가 완료**되었습니다.

**완성된 문서 세트**:
✅ `.moai/project/product.md` - 비즈니스 요구사항
✅ `.moai/project/structure.md` - 아키텍처 설계 + 레거시 분석
✅ `.moai/project/tech.md` - 기술 스택 상세 + 현대화 로드맵

**다음 단계**: `/moai:1-spec` 실행으로 구체적 기능 명세 작성

---

_@TAG: @TECH:MOAI-ADK-001 @STACK:PYTHON-001 @DEBT:TECH-ANALYSIS-001_
_문서 생성일: 2024-09-24_
_최종 수정: 브레인스토밍 통합 및 기술 부채 분석 추가_
