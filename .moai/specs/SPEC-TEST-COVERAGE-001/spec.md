---
id: TEST-COVERAGE-001
version: 0.1.0
status: implementation-complete
created: 2025-10-15
updated: 2025-10-15
author: @Goos
priority: high
category: refactor
labels:
  - testing
  - coverage
  - quality
depends_on: []
blocks: []
related_specs: []
scope:
  packages:
    - src/moai_adk/cli/commands
    - src/moai_adk/core/git
    - tests/unit
    - tests/integration
---

# @SPEC:TEST-COVERAGE-001: CLI 및 Git 모듈 테스트 커버리지 85% 달성

## HISTORY

### v0.1.0 (2025-10-15)
- **COMPLETED**: TDD 구현 완료 (RED-GREEN-REFACTOR)
- **AUTHOR**: @Goos
- **ACHIEVEMENT**:
  - 272 tests 작성 (19 test files)
  - 85.61% coverage 달성 (726/848 statements)
  - 0 test failures
  - 0 linter warnings
- **TDD COMMITS**:
  - d74cd76: 🔴 RED - 테스트 인프라 구축
  - 9886550: 🟢 GREEN - 단위 테스트 (52% coverage)
  - 08aa938: 🟢 GREEN - 통합 테스트 (85.61% coverage)
  - 478729d: ♻️ REFACTOR - Ruff 린터 개선
- **TEST BREAKDOWN**:
  - Unit tests: 148 tests (17 files)
  - Integration tests: 124 tests (2 files)
  - 100% coverage modules: banner, git utils, template config, initializer
- **TOOLS**: pytest 8.4.2, pytest-cov 7.0.0, Click CliRunner, uv package manager

### v0.0.1 (2025-10-15)
- **INITIAL**: 테스트 커버리지 85% 달성 명세 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: CLI 명령어, Git 모듈, Template 엣지 케이스
- **CONTEXT**: TRUST 원칙 T (Test First) 위반 해소 - 현재 72.06% → 목표 85%

## Environment (환경)

- **Python 버전**: 3.13.1
- **테스트 프레임워크**: pytest 8.4.2
- **커버리지 도구**: pytest-cov 6.0.0
- **현재 커버리지**: 72.06% (619/859 statements)
- **목표 커버리지**: 85% (730/859 statements)
- **Gap**: 12.94% (111 statements 추가 커버 필요)

## Assumptions (가정)

- 기존 테스트 인프라 (pytest, pytest-cov)가 정상 작동한다
- 각 테스트는 독립적으로 실행 가능하다
- Git 모듈은 실제 Git 저장소에서 테스트된다 (임시 디렉토리 사용)
- CI/CD 파이프라인이 커버리지 85% 임계값을 검증한다

## Requirements (요구사항)

### Ubiquitous (필수 요구사항)

- 시스템은 85% 이상의 테스트 커버리지를 제공해야 한다
- 시스템은 모든 핵심 모듈(CLI, Git, Template)의 단위 테스트를 포함해야 한다
- 시스템은 실패 테스트 0개를 유지해야 한다
- 시스템은 커버리지 리포트를 자동 생성해야 한다

### Event-driven (이벤트 기반 요구사항)

- WHEN 새로운 CLI 명령어가 추가되면, 시스템은 통합 테스트를 포함해야 한다
- WHEN Git 모듈이 수정되면, 시스템은 브랜치/커밋 테스트를 실행해야 한다
- WHEN pytest 실행 시, 시스템은 커버리지 리포트를 HTML/터미널로 생성해야 한다
- WHEN 테스트가 실패하면, 시스템은 실패 원인과 함께 상세 로그를 출력해야 한다

### State-driven (상태 기반 요구사항)

- WHILE 테스트 커버리지가 85% 미만일 때, 시스템은 CI/CD 빌드를 차단해야 한다
- WHILE 실패 테스트가 존재할 때, 시스템은 경고를 표시하고 빌드를 실패시켜야 한다
- WHILE 테스트 실행 시간이 30초를 초과할 때, 시스템은 성능 경고를 표시해야 한다

### Optional (선택적 기능)

- WHERE 커버리지가 90% 이상일 때, 시스템은 품질 배지를 표시할 수 있다
- WHERE 병렬 테스트 실행이 가능할 때, 시스템은 pytest-xdist를 사용할 수 있다

### Constraints (제약사항)

- IF 새로운 코드가 추가되면, 커버리지는 85% 이하로 떨어지지 않아야 한다
- 각 테스트는 독립적으로 실행 가능해야 한다
- 테스트 실행 시간은 30초를 초과하지 않아야 한다
- Git 테스트는 임시 디렉토리를 사용하고 테스트 종료 시 정리해야 한다

## Implementation Details (구현 세부사항)

### Phase 1: 실패 테스트 수정 (+8%)
**목표**: 72% → 80%

1. **CLI 통합 테스트 수정**:
   - `test_cli.py`: `test_no_arguments`, `test_init_command_*` 수정
   - `test_cli_commands.py`: 초기화 명령어 테스트 수정
   - `test_commands.py`: InitCommand, StatusCommand, RestoreCommand 수정

2. **moai_hooks.py 테스트 수정** (49개 실패):
   - Language Detection 테스트 18개 수정
   - JIT Context 테스트 6개 수정
   - Hook Handlers 테스트 11개 수정

3. **Git Utils 테스트 수정**:
   - `test_git_utils.py`: 3개 실패 테스트 수정
   - `test_git_manager.py`: 3개 실패 테스트 수정

### Phase 2: Git 모듈 테스트 추가 (+5%)
**목표**: 80% → 85%

1. **단위 테스트 작성**:
   - `tests/unit/test_git_manager.py`: GitManager 클래스 테스트
   - `tests/unit/test_git_branch.py`: 브랜치 생성/전환 테스트
   - `tests/unit/test_git_commit.py`: 커밋 메시지 생성 테스트

2. **통합 테스트 작성**:
   - `tests/integration/test_git_workflow.py`: 전체 Git 워크플로우 테스트

### Phase 3: 엣지 케이스 테스트 추가 (유지)
**목표**: 85% 유지

1. **Template Processor 엣지 케이스**: 9개 실패 테스트 수정
2. **Backup Utils 엣지 케이스**: 메타데이터 저장 테스트 수정
3. **Project Initializer 엣지 케이스**: 10개 실패 테스트 수정

## Acceptance Criteria (인수 기준)

### AC-1: 실패 테스트 수정
- **Given**: 100개의 실패 테스트가 존재할 때
- **When**: 각 테스트를 수정하고 `pytest -v` 실행 시
- **Then**:
  - 모든 테스트가 통과해야 한다 (0 failed)
  - 커버리지가 80% 이상이어야 한다
  - 테스트 실행 시간이 30초 이내여야 한다

### AC-2: Git 모듈 단위 테스트
- **Given**: Git 모듈이 0% 커버리지일 때
- **When**: 브랜치 생성/커밋 테스트를 작성하고 실행하면
- **Then**:
  - `core/git/manager.py` 커버리지가 60% 이상이어야 한다
  - `core/git/branch.py` 커버리지가 70% 이상이어야 한다
  - `core/git/commit.py` 커버리지가 70% 이상이어야 한다
  - 모든 Git 작업이 임시 디렉토리에서 실행되어야 한다

### AC-3: 전체 커버리지 목표 달성
- **Given**: 모든 테스트가 작성되어 있을 때
- **When**: `pytest --cov --cov-report=term-missing` 실행 시
- **Then**:
  - 전체 커버리지가 85% 이상이어야 한다
  - 실패 테스트가 0개여야 한다
  - 커버리지 리포트가 HTML로 생성되어야 한다
  - CI/CD 빌드가 통과해야 한다

## Traceability (@TAG)

- **SPEC**: `@SPEC:TEST-COVERAGE-001`
- **TEST**: `tests/unit/test_*`, `tests/integration/test_git_workflow.py`
- **CODE**: `src/moai_adk/cli/commands/*`, `src/moai_adk/core/git/*`

## References

- **TRUST 원칙**: `.moai/memory/development-guide.md#trust-5원칙`
- **테스트 전략**: `pyproject.toml` [tool.pytest.ini_options]
- **커버리지 설정**: `pyproject.toml` [tool.coverage.run]
