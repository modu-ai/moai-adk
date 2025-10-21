---
id: TECH-001
version: 0.1.2
status: active
created: 2025-10-01
updated: 2025-10-21
author: @tech-lead
priority: medium
---

# MoAI-ADK Technology Stack

## HISTORY

### v0.1.2 (2025-10-21)
- **UPDATED**: Backup merge - 템플릿 플레이스홀더를 실제 MoAI-ADK 기술 스택으로 교체
- **AUTHOR**: @Alfred
- **SECTIONS**: Stack (Python 3.13+/uv), Framework (Click/Rich/GitPython), Quality (pytest/mypy/ruff), Security (pip-audit/bandit), Deploy (PyPI/GitHub Actions)
- **REASON**: Smart Update 전략에 따라 실제 프로젝트 기술 스택 반영

### v0.1.1 (2025-10-17)
- **UPDATED**: 템플릿 버전 동기화 (v0.3.8)
- **AUTHOR**: @Alfred
- **SECTIONS**: 메타데이터 표준화 (author 필드 단수형, priority 추가)

### v0.1.0 (2025-10-01)
- **INITIAL**: 프로젝트 기술 스택 문서 작성
- **AUTHOR**: @tech-lead
- **SECTIONS**: Stack, Framework, Quality, Security, Deploy

---

## @DOC:STACK-001 언어 & 런타임

### 주 언어 선택

- **언어**: Python
- **버전**: 3.13+ (최신 안정 버전)
- **선택 이유**:
  - **풍부한 생태계**: CLI 도구 (Click), TUI (Rich, Questionary), Git 통합 (GitPython)
  - **빠른 프로토타이핑**: 스크립팅 언어로서 빠른 반복 개발
  - **타입 안전성**: Type Hints + mypy로 정적 타입 검사
  - **PyPI 배포**: 표준 패키지 배포 채널
- **패키지 매니저**: uv (pip 대비 10~100배 빠른 설치/해결 속도)

**트레이드오프**:
- 장점: 빠른 개발, 풍부한 라이브러리, 크로스 플랫폼 지원
- 단점: 런타임 성능 (Python은 인터프리터 언어), 배포 시 런타임 의존성

### 멀티 플랫폼 지원

| 플랫폼      | 지원 상태 | 검증 도구             | 주요 제약                             |
| ----------- | --------- | --------------------- | ------------------------------------- |
| **Windows** | ✅ 지원    | GitHub Actions (Windows) | 경로 구분자 (`\`), ANSI 색상 제한     |
| **macOS**   | ✅ 지원    | GitHub Actions (macOS)   | M1/M2 Apple Silicon 호환 검증 필요    |
| **Linux**   | ✅ 지원    | GitHub Actions (Ubuntu)  | 없음 (주 개발 환경)                   |

**검증 전략**:
- CI/CD에서 3개 OS 모두 테스트 실행
- 플랫폼별 경로 처리: `pathlib.Path` 사용 (OS 독립적)

## @DOC:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### 1. 주요 의존성 (런타임)

```toml
[project.dependencies]
click = ">=8.1.0"           # CLI 프레임워크 (명령 파싱, 옵션 처리)
rich = ">=13.0.0"           # TUI 렌더링 (색상, 표, 진행 바)
pyfiglet = ">=1.0.2"        # ASCII 배너 생성
questionary = ">=2.0.0"     # 대화형 프롬프트 (선택, 입력)
gitpython = ">=3.1.45"      # Git 자동화 (브랜치, 커밋, 상태)
packaging = ">=21.0"        # 버전 비교 (Semantic Versioning)
```

**선택 근거**:
- **Click**: 업계 표준 CLI 프레임워크 (Flask, AWS CLI 사용)
- **Rich**: 최고의 TUI 라이브러리 (Markdown 렌더링, 진행 바 지원)
- **Questionary**: 직관적인 대화형 프롬프트 (TUI 통합)
- **GitPython**: Git 명령 추상화 (subprocess 대비 안전)

### 2. 개발 도구 (devDependencies)

```toml
[project.optional-dependencies]
dev = [
    "pytest>=8.4.2",           # 테스트 프레임워크
    "pytest-cov>=7.0.0",       # 커버리지 측정
    "pytest-xdist>=3.8.0",     # 병렬 테스트 실행
    "ruff>=0.1.0",             # 린터 + 포매터 (통합)
    "mypy>=1.7.0",             # 정적 타입 검사
    "types-PyYAML>=6.0.0"      # PyYAML 타입 스텁
]
security = [
    "pip-audit>=2.7.0",        # 의존성 보안 감사
    "bandit>=1.8.0"            # Python 코드 보안 분석
]
```

**도구 선택 이유**:
- **pytest**: Python 표준 테스트 프레임워크 (플러그인 생태계 풍부)
- **ruff**: Rust 기반 초고속 린터/포매터 (Flake8 + Black 통합, 10~100배 빠름)
- **mypy**: 공식 타입 체커 (strict 모드 지원)
- **pip-audit**: PyPI 공식 보안 감사 도구

### 3. 빌드 시스템

- **빌드 도구**: Hatchling (PEP 517 표준, setuptools 대체)
- **번들링**: Python wheel (`.whl`) 형식
- **타겟**: Python 3.13+ 환경
- **성능 목표**: 빌드 시간 <10s, 설치 시간 <30s (uv 사용 시)

**Hatchling 선택 이유**:
- PEP 517/518 표준 준수
- setuptools보다 빠르고 간결한 설정
- 템플릿 파일 포함 지원 (`.claude/`, `.moai/`)

## @DOC:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지

- **목표**: 85% 이상 (엄격한 품질 기준)
- **측정 도구**: pytest-cov (Coverage.py 기반)
- **실패 시 대응**: CI/CD에서 자동 실패, PR 머지 차단

**커버리지 설정**:
```toml
[tool.coverage.report]
precision = 2
show_missing = true
skip_covered = false
fail_under = 85  # 85% 미만 시 실패
```

### 정적 분석

| 도구       | 역할              | 설정 파일           | 실패 시 조치            |
| ---------- | ----------------- | ------------------- | ----------------------- |
| **ruff**   | 린터 + 포매터     | `pyproject.toml`    | CI/CD 실패, 자동 수정   |
| **mypy**   | 타입 체커         | `pyproject.toml`    | CI/CD 실패, 타입 수정   |
| **bandit** | 보안 분석         | CLI 실행            | 경고 리뷰, 중요도 높음  |

**ruff 설정**:
```toml
[tool.ruff]
line-length = 120
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "N"]  # 오류, 경고, import, 네이밍
ignore = []
```

**mypy 설정** (strict mode):
```bash
mypy src/ --strict --ignore-missing-imports
```

### 자동화 스크립트

```bash
# 품질 검사 파이프라인 (로컬 실행)
pytest --cov=src/moai_adk --cov-report=html  # 테스트 + 커버리지
ruff check src/ tests/                        # 린트 검사
ruff format src/ tests/                       # 코드 포매팅
mypy src/ --strict                            # 타입 검사
```

**CI/CD 자동 실행**:
- PR 생성 시: 모든 품질 검사 자동 실행
- 실패 시: PR 머지 차단
- 성공 시: 자동 승인 가능

## @DOC:SECURITY-001 보안 정책 & 운영

### 비밀 관리

- **정책**: 환경 변수 사용 (`.env` 파일, Git 제외)
- **도구**: 없음 (현재 API 키 불필요, 로컬 도구만 사용)
- **검증**: `.gitignore`에 `.env` 포함 여부 확인

**향후 계획**:
- Claude API 키 지원 시: `python-dotenv` 추가
- GitHub Actions Secrets 사용

### 의존성 보안

```toml
[project.optional-dependencies]
security = [
    "pip-audit>=2.7.0",  # 의존성 보안 감사 (CVE 데이터베이스 검사)
    "bandit>=1.8.0"      # Python 코드 보안 분석 (취약한 패턴 감지)
]
```

**보안 정책**:
- **audit_tool**: pip-audit (PyPI 공식 도구)
- **update_policy**: 보안 패치 발견 시 즉시 업데이트
- **vulnerability_threshold**: CRITICAL 취약점 0개 (PR 머지 차단)

**실행 예시**:
```bash
# 의존성 보안 감사
pip-audit

# Python 코드 보안 분석
bandit -r src/ -ll  # Low severity 이상만 표시
```

### 로깅 정책

- **로그 수준**: INFO (운영), DEBUG (개발)
- **민감정보 마스킹**:
  - 파일 경로: 절대 경로 → 상대 경로 변환
  - 사용자 입력: 비밀번호/토큰 `***` 처리
- **보존 정책**: 로그 파일 없음 (stdout/stderr만 사용, 시스템 로그로 리다이렉션)

**로깅 구현**:
```python
import logging
logger = logging.getLogger("moai-adk")
logger.setLevel(logging.INFO)  # 운영 모드
```

## @DOC:DEPLOY-001 배포 채널 & 전략

### 1. 배포 채널

- **주 채널**: PyPI (Python Package Index)
- **릴리스 절차**:
  1. GitHub Release 태그 생성 (`v0.3.13`)
  2. GitHub Actions CI/CD 트리거
  3. 자동 빌드 (`python -m build`)
  4. PyPI 배포 (`twine upload`)
- **버전 정책**: Semantic Versioning (MAJOR.MINOR.PATCH)
  - MAJOR: 하위 호환성 깨지는 변경
  - MINOR: 새 기능 추가 (하위 호환)
  - PATCH: 버그 수정
- **rollback 전략**: PyPI는 삭제 불가 → 새 패치 버전 배포

**배포 명령**:
```bash
# 로컬 빌드 (테스트용)
python -m build

# PyPI 배포 (CI/CD에서 자동)
twine upload dist/*
```

### 2. 개발 설치

```bash
# 개발자 모드 설정 (editable install)
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# uv 사용 (권장)
uv venv .venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate
uv pip install -e ".[dev,security]"

# pip 사용 (대안)
pip install -e ".[dev,security]"

# 테스트 실행
pytest --cov=src/moai_adk
```

### 3. CI/CD 파이프라인

| 단계             | 목적               | 사용 도구                   | 성공 조건                               |
| ---------------- | ------------------ | --------------------------- | --------------------------------------- |
| **Checkout**     | 코드 체크아웃      | actions/checkout@v4         | N/A                                     |
| **Setup**        | Python 환경 설정   | actions/setup-python@v5     | Python 3.13 설치                        |
| **Install**      | 의존성 설치        | uv pip install              | 모든 의존성 설치 성공                   |
| **Lint**         | 코드 품질 검사     | ruff check                  | 린트 오류 0개                           |
| **Type Check**   | 타입 검사          | mypy --strict               | 타입 오류 0개                           |
| **Test**         | 단위 테스트        | pytest --cov                | 테스트 통과 + 커버리지 ≥85%             |
| **Security**     | 보안 감사          | pip-audit, bandit           | CRITICAL 취약점 0개                     |
| **Build**        | 빌드 검증          | python -m build             | wheel 빌드 성공                         |
| **Deploy**       | PyPI 배포          | twine upload (release만)    | 배포 성공 (태그 생성 시에만 실행)       |

**GitHub Actions 워크플로우**:
```yaml
name: MoAI-ADK CI/CD

on:
  push:
    branches: [develop, main]
  pull_request:
    branches: [develop, main]
  release:
    types: [published]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        python-version: ["3.13"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: ${{ matrix.python-version }}
      - run: pip install uv
      - run: uv pip install -e ".[dev,security]"
      - run: ruff check src/ tests/
      - run: mypy src/ --strict
      - run: pytest --cov=src/moai_adk --cov-report=xml
      - run: pip-audit
      - run: bandit -r src/ -ll

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'release'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
      - run: pip install build twine
      - run: python -m build
      - run: twine upload dist/*
        env:
          TWINE_USERNAME: __token__
          TWINE_PASSWORD: ${{ secrets.PYPI_API_TOKEN }}
```

## 환경별 설정

### 개발 환경 (`dev`)

```bash
export MOAI_ENV=development
export MOAI_LOG_LEVEL=DEBUG
export MOAI_DEBUG=true

# 개발 서버 실행 (자동 리로드)
moai-adk --debug
```

### 테스트 환경 (`test`)

```bash
export MOAI_ENV=test
export MOAI_LOG_LEVEL=INFO
export MOAI_COVERAGE=true

# 테스트 실행 (커버리지 포함)
pytest --cov=src/moai_adk --cov-report=html
```

### 프로덕션 환경 (`production`)

```bash
export MOAI_ENV=production
export MOAI_LOG_LEVEL=WARNING

# PyPI에서 설치 (일반 사용자)
uv tool install moai-adk
moai-adk init my-project
```

## @CODE:TECH-DEBT-001 기술 부채 관리

### 현재 기술 부채

1. **MCP (Model Context Protocol) 미지원** - 우선순위: 중간
   - 현재: `.claude/` 디렉토리 기반 통합만 지원
   - 목표: Claude Desktop과 직접 통합 (MCP 서버 구현)
   - 타임라인: v0.4.0 (3개월)

2. **플러그인 시스템 부재** - 우선순위: 낮음
   - 현재: 모든 에이전트/Skills가 하드코딩
   - 목표: 사용자 정의 에이전트/Skills 동적 로드
   - 타임라인: v0.5.0 (6개월)

3. **성능 최적화 미흡** - 우선순위: 낮음
   - 현재: 템플릿 처리 시 파일 I/O 비효율
   - 목표: 캐싱, 병렬 처리 도입
   - 타임라인: v0.4.x (패치)

### 개선 계획

- **단기 (1개월)**: pip-audit + bandit 자동화 (CI/CD 통합)
- **중기 (3개월)**: MCP 서버 구현, Claude Desktop 통합
- **장기 (6개월+)**: 플러그인 아키텍처, 성능 최적화

## EARS 기술 요구사항 작성법

### 기술 스택에서의 EARS 활용

기술적 의사결정과 품질 게이트 설정 시 EARS 구문을 활용하여 명확한 기술 요구사항을 정의하세요:

#### 기술 스택 EARS 예시
```markdown
### Ubiquitous Requirements (기본 기술 요구사항)
- 시스템은 Python 3.13+ 타입 안전성을 보장해야 한다
- 시스템은 크로스 플랫폼 호환성을 제공해야 한다

### Event-driven Requirements (이벤트 기반 기술)
- WHEN 코드가 커밋되면, 시스템은 자동으로 테스트를 실행해야 한다
- WHEN 빌드가 실패하면, 시스템은 개발자에게 즉시 알림을 보내야 한다

### State-driven Requirements (상태 기반 기술)
- WHILE 개발 모드일 때, 시스템은 DEBUG 로그를 제공해야 한다
- WHILE 프로덕션 모드일 때, 시스템은 최적화된 빌드를 생성해야 한다

### Optional Features (선택적 기술)
- WHERE uv가 설치되면, 시스템은 pip 대신 uv를 사용할 수 있다
- WHERE CI/CD가 구성되면, 시스템은 자동 배포를 수행할 수 있다

### Constraints (기술적 제약사항)
- IF 의존성에 CRITICAL 취약점이 발견되면, 시스템은 빌드를 중단해야 한다
- 테스트 커버리지는 85% 이상을 유지해야 한다
- 빌드 시간은 10초를 초과하지 않아야 한다
```

---

_이 기술 스택은 `/alfred:2-run` 실행 시 TDD 도구 선택과 품질 게이트 적용의 기준이 됩니다._
