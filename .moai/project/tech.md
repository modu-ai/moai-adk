---
id: TECH-001
version: 0.1.1
status: active
created: 2025-10-01
updated: 2025-10-17
author: @Goos
priority: high
---

# MoAI-ADK Technology Stack

## HISTORY

### v0.1.1 (2025-10-17)
- **UPDATED**: 템플릿 기본값을 실제 MoAI-ADK 기술 스택 내용으로 갱신
- **AUTHOR**: @Alfred
- **SECTIONS**: Stack, Framework, Quality, Security, Deploy 실제 내용 반영

### v0.1.0 (2025-10-01)
- **INITIAL**: 프로젝트 기술 스택 문서 작성
- **AUTHOR**: @tech-lead
- **SECTIONS**: Stack, Framework, Quality, Security, Deploy

---

## @DOC:STACK-001 언어 & 런타임

### 주 언어 선택

- **언어**: Python
- **버전**: ≥3.13
- **선택 이유**:
  - **성숙한 생태계**: PyPI 패키지 생태계 풍부, AI/ML 도구 최적화
  - **타입 힌트**: Python 3.13+ 타입 시스템으로 안정성 확보
  - **크로스 플랫폼**: Windows, macOS, Linux 완벽 지원
  - **CLI 친화적**: click, rich 등 CLI 라이브러리 강력
- **패키지 매니저**: uv (Astral의 초고속 패키지 매니저)

**uv 선택 이유**:
- **속도**: pip 대비 10~100배 빠른 의존성 설치
- **신뢰성**: Rust 기반 구현으로 안정성 보장
- **호환성**: pip, virtualenv 완전 호환
- **트레이드오프**: Python 3.13+ 필수, 일부 레거시 환경 비호환

### 멀티 플랫폼 지원

| 플랫폼 | 지원 상태 | 검증 도구 | 주요 제약 |
|--------|-----------|-----------|-----------|
| **Windows** | ✅ 완벽 지원 | PowerShell 스크립트 | 경로 구분자 `\`, Git Bash 권장 |
| **macOS** | ✅ 완벽 지원 | Bash 스크립트 | Python 3.13+ 수동 설치 필요 |
| **Linux** | ✅ 완벽 지원 | Bash 스크립트 | 배포판별 패키지 관리 차이 |

**플랫폼 감지 및 대응**:
```python
import platform
system = platform.system()  # 'Windows', 'Darwin', 'Linux'

if system == 'Windows':
    # PowerShell 명령어 사용
    separator = '\\'
elif system in ('Darwin', 'Linux'):
    # Bash 명령어 사용
    separator = '/'
```

## @DOC:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### 1. 주요 의존성

```toml
[project.dependencies]
click = ">=8.1.0"              # CLI 프레임워크
rich = ">=13.0.0"              # 터미널 출력 포맷팅
pyfiglet = ">=1.0.2"           # ASCII 아트 배너
questionary = ">=2.0.0"        # 인터랙티브 프롬프트
gitpython = ">=3.1.45"         # Git 조작
```

**라이브러리 선택 이유**:
- **click**: 타입 힌트 지원, 자동 도움말 생성, 플러그인 시스템
- **rich**: 터미널 색상, 테이블, 진행 표시줄, 마크다운 렌더링
- **pyfiglet**: ASCII 아트로 브랜드 정체성 강화
- **questionary**: 사용자 친화적 프롬프트 (체크박스, 선택 메뉴)
- **gitpython**: 순수 Python Git 라이브러리, 크로스 플랫폼

### 2. 개발 도구

```toml
[project.optional-dependencies]
dev = [
    "pytest>=8.4.2",           # 테스트 프레임워크
    "pytest-cov>=7.0.0",       # 커버리지 측정
    "pytest-xdist>=3.8.0",     # 병렬 테스트 실행
    "ruff>=0.1.0",             # 초고속 린터+포매터
    "mypy>=1.7.0"              # 정적 타입 검사
]
security = [
    "pip-audit>=2.7.0",        # 의존성 취약점 스캔
    "bandit>=1.8.0"            # 보안 코드 스캔
]
```

**도구 선택 이유**:
- **pytest**: Python 표준 테스트 프레임워크, 플러그인 생태계 풍부
- **pytest-cov**: 커버리지 측정 및 HTML 리포트 생성
- **pytest-xdist**: 멀티코어 병렬 실행으로 테스트 속도 향상
- **ruff**: Rust 기반 초고속 린터 (flake8, black, isort 통합)
- **mypy**: 타입 힌트 검증으로 런타임 에러 사전 방지

### 3. 빌드 시스템

- **빌드 도구**: hatchling
- **번들링**: wheel + sdist (PyPI 표준 형식)
- **타겟**: Python 3.13+ 환경
- **성능 목표**: 빌드 시간 <10초

**hatchling 선택 이유**:
- **현대적**: PEP 517/518 완벽 준수
- **유연성**: 플러그인 시스템으로 확장 가능
- **속도**: 빠른 빌드 속도
- **템플릿 지원**: 비코드 파일 (`.moai/`, `.claude/`) 포함 용이

## @DOC:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지

- **목표**: 85% 이상
- **현재**: 87.66%
- **측정 도구**: pytest-cov + codecov
- **실패 시 대응**: 커버리지 85% 미만 시 CI 실패

**커버리지 측정**:
```bash
pytest --cov=src/moai_adk --cov-report=html --cov-report=term-missing
```

**커버리지 정책**:
```toml
[tool.coverage.report]
precision = 2
show_missing = true
skip_covered = false
fail_under = 85
```

### 정적 분석

| 도구 | 역할 | 설정 파일 | 실패 시 조치 |
|------|------|-----------|--------------|
| **ruff** | 린터+포매터 | `pyproject.toml` | CI 실패, 자동 수정 제안 |
| **mypy** | 타입 검사 | `pyproject.toml` | CI 경고, 타입 힌트 추가 권장 |
| **bandit** | 보안 스캔 | `pyproject.toml` | 보안 이슈 즉시 수정 |
| **pip-audit** | 의존성 취약점 | CLI | 취약 패키지 업데이트 |

**ruff 설정**:
```toml
[tool.ruff]
line-length = 120
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "N"]
ignore = []
```

**mypy 설정** (미래 계획):
```toml
[tool.mypy]
python_version = "3.13"
strict = true
warn_return_any = true
warn_unused_configs = true
```

### 자동화 스크립트

```bash
# 품질 검사 파이프라인 (로컬)
pytest --cov=src/moai_adk --cov-report=html         # 테스트 실행
ruff check src/moai_adk                              # 린트 검사
ruff format src/moai_adk                             # 코드 포맷팅
mypy src/moai_adk                                    # 타입 검사 (선택)
bandit -r src/moai_adk                               # 보안 스캔
pip-audit                                            # 의존성 취약점 스캔

# CI/CD 파이프라인 (GitHub Actions)
# .github/workflows/moai-gitflow.yml 참조
```

## @DOC:SECURITY-001 보안 정책 & 운영

### 비밀 관리

- **정책**: 환경 변수 기반 비밀 관리, 코드베이스에 비밀 저장 금지
- **도구**: GitHub Secrets (CI/CD), .env 파일 (로컬 개발)
- **검증**: `.gitignore`에 `.env` 포함, pre-commit hook으로 비밀 스캔

**비밀 환경 변수**:
- `GITHUB_TOKEN`: GitHub API 인증 (PR 생성, 자동 머지)
- `PYPI_TOKEN`: PyPI 배포 인증

**비밀 검증**:
```bash
# .gitignore 확인
grep -q ".env" .gitignore || echo ".env 누락!"

# 커밋 전 비밀 스캔 (선택)
rg "(GITHUB_TOKEN|PYPI_TOKEN|api_key|password)" src/ --ignore-case
```

### 의존성 보안

```toml
[project.optional-dependencies]
security = [
    "pip-audit>=2.7.0",        # 의존성 취약점 스캔
    "bandit>=1.8.0"            # 보안 코드 스캔
]
```

**보안 스캔 주기**:
- **일간**: GitHub Actions 자동 실행 (Dependabot)
- **주간**: 수동 pip-audit 실행
- **출시 전**: 전체 보안 스캔 필수

**보안 정책**:
```json
{
  "security": {
    "audit_tool": "pip-audit",
    "update_policy": "patch-within-24h",
    "vulnerability_threshold": "medium"
  }
}
```

### 로깅 정책

- **로그 수준**:
  - `DEBUG`: 개발 환경 전용 (상세 디버그 정보)
  - `INFO`: 일반 작업 정보 (기본값)
  - `WARNING`: 주의 필요한 상황
  - `ERROR`: 에러 발생 시
  - `CRITICAL`: 치명적 오류
- **민감정보 마스킹**: API 키, 토큰 자동 마스킹 (예: `GITHUB_TOKEN=***`)
- **보존 정책**: 로컬 로그 파일 7일 보존, 이후 자동 삭제

**로깅 구현** (rich 기반):
```python
from rich.console import Console
from rich.logging import RichHandler

console = Console()
logger = logging.getLogger("moai-adk")
logger.addHandler(RichHandler(console=console, rich_tracebacks=True))
```

## @DOC:DEPLOY-001 배포 채널 & 전략

### 1. 배포 채널

- **주 채널**: PyPI (https://pypi.org/project/moai-adk/)
- **릴리스 절차**:
  1. 버전 태그 생성 (`git tag v0.3.1`)
  2. GitHub Actions 자동 트리거
  3. hatch build 실행 (wheel + sdist)
  4. twine upload 실행 (PyPI)
  5. GitHub Release 생성
- **버전 정책**: Semantic Versioning (0.x.y)
  - Patch (0.3.1 → 0.3.2): 버그 수정
  - Minor (0.3.x → 0.4.0): 기능 추가
  - Major (0.x.y → 1.0.0): 하위 호환 깨지는 변경 (명시적 승인 필요)
- **rollback 전략**: PyPI는 rollback 불가, 새 버전 배포로 대응

### 2. 개발 설치

```bash
# 1. uv 설치 (필수)
curl -LsSf https://astral.sh/uv/install.sh | sh

# 2. 프로젝트 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 3. 개발 환경 설정
uv venv                                    # 가상 환경 생성
source .venv/bin/activate                  # 활성화 (macOS/Linux)
# .venv\Scripts\activate                   # 활성화 (Windows)

uv pip install -e ".[dev,security]"        # 개발 의존성 포함 설치

# 4. 테스트 실행
pytest

# 5. 로컬 빌드
hatch build
```

### 3. CI/CD 파이프라인

| 단계 | 목적 | 사용 도구 | 성공 조건 |
|------|------|-----------|-----------|
| **Checkout** | 코드 체크아웃 | actions/checkout@v4 | 항상 성공 |
| **Setup** | Python 3.13 설치 | actions/setup-python@v5 | Python 3.13+ |
| **Install** | 의존성 설치 | uv pip install | 모든 패키지 설치 |
| **Lint** | 코드 품질 검사 | ruff check | 위반 없음 |
| **Test** | 테스트 실행 | pytest | 모든 테스트 통과 |
| **Coverage** | 커버리지 측정 | pytest-cov | ≥85% |
| **Security** | 보안 스캔 | bandit, pip-audit | 치명적 취약점 없음 |
| **Build** | 패키지 빌드 | hatch build | 빌드 성공 |
| **Deploy** | PyPI 배포 | twine upload | tag push 시에만 |

**GitHub Actions 워크플로우**:
- `.github/workflows/moai-gitflow.yml`: 메인 CI/CD
- `.github/workflows/publish-pypi.yml`: PyPI 배포 전용

**트리거 조건**:
- `push` (main, develop 브랜치): 전체 파이프라인 실행
- `pull_request` (모든 PR): 테스트 및 품질 검사만 실행
- `tag push` (v*): 배포 파이프라인 실행

## 환경별 설정

### 개발 환경 (`dev`)

```bash
export PROJECT_MODE=development
export LOG_LEVEL=debug
export MOAI_ADK_DEV=1

# uv 기반 개발
uv pip install -e ".[dev,security]"
pytest --cov=src/moai_adk
```

**특징**:
- 상세한 디버그 로그
- Hot-reload (선택)
- 로컬 파일 시스템만 사용

### 테스트 환경 (`test`)

```bash
export PROJECT_MODE=test
export LOG_LEVEL=info

# CI 환경에서 실행
pytest --cov=src/moai_adk --cov-report=xml --cov-report=term-missing
```

**특징**:
- 커버리지 측정 필수
- 모든 테스트 통과 필수
- 보안 스캔 자동 실행

### 프로덕션 환경 (`production`)

```bash
export PROJECT_MODE=production
export LOG_LEVEL=warning

# PyPI 설치
pip install moai-adk
moai-adk --version
```

**특징**:
- 최소 로깅 (경고 이상만)
- 에러 트래킹 (선택)
- 성능 최적화

## @CODE:TECH-DEBT-001 기술 부채 관리

### 현재 기술 부채

#### 높은 우선순위

1. **타입 힌트 완전성** (진행률: 70%)
   - **문제**: 일부 모듈에 타입 힌트 누락
   - **영향**: mypy strict 모드 적용 불가
   - **완화 방안**: 점진적 타입 힌트 추가, `# type: ignore` 최소화
   - **해결 계획**: v0.4.0에서 mypy strict 모드 완전 적용

2. **Windows 경로 처리** (진행률: 80%)
   - **문제**: 일부 경로 하드코딩 (`/` 구분자)
   - **영향**: Windows에서 간헐적 오류 발생
   - **완화 방안**: `pathlib.Path` 사용 강제
   - **해결 계획**: v0.3.2에서 전체 경로 처리 통일

#### 중간 우선순위

3. **테스트 격리** (진행률: 85%)
   - **문제**: 일부 통합 테스트가 파일 시스템 상태에 의존
   - **영향**: 테스트 순서에 따라 실패 가능
   - **완화 방안**: 임시 디렉토리 사용 (`pytest.tmpdir`)
   - **해결 계획**: v0.4.0에서 완전 격리

4. **Hooks 오류 처리** (진행률: 90%)
   - **문제**: Hooks 실패 시 fallback 로직 부족
   - **영향**: Claude Code 환경에서 일부 기능 동작 불가
   - **완화 방안**: 에이전트 직접 호출로 대체
   - **해결 계획**: v0.3.2에서 fallback 로직 강화

### 개선 계획

#### 단기 (1개월 - v0.3.x)

- **타입 힌트 100%**: 모든 공개 함수에 타입 힌트 추가
- **Windows 경로 처리 통일**: `pathlib.Path` 전환 완료
- **Hooks fallback 로직 강화**: 에이전트 자동 전환

#### 중기 (3개월 - v0.4.x)

- **테스트 완전 격리**: 모든 테스트가 독립 실행 가능
- **mypy strict 모드**: 타입 안전성 100% 보장
- **성능 최적화**: TAG 스캔 캐싱 (CODE-FIRST 원칙 유지)

#### 장기 (6개월+ - v0.5.x)

- **Plugin 시스템**: 사용자 정의 에이전트/커맨드 지원
- **Web UI**: 프로젝트 대시보드 및 TAG 그래프 시각화
- **멀티 프로젝트**: 여러 프로젝트 동시 관리

## EARS 기술 요구사항 작성법

### 기술 스택에서의 EARS 활용

기술적 의사결정과 품질 게이트 설정 시 EARS 구문을 활용하여 명확한 기술 요구사항을 정의하세요:

#### 기술 스택 EARS 예시

```markdown
### Ubiquitous Requirements (기본 기술 요구사항)
- 시스템은 Python 3.13+ 타입 안전성을 보장해야 한다
- 시스템은 크로스 플랫폼 호환성을 제공해야 한다 (Windows, macOS, Linux)
- 시스템은 uv 패키지 매니저를 사용해야 한다

### Event-driven Requirements (이벤트 기반 기술)
- WHEN 코드가 커밋되면, 시스템은 자동으로 테스트를 실행해야 한다
- WHEN 빌드가 실패하면, 시스템은 개발자에게 즉시 알림을 보내야 한다
- WHEN 의존성에 취약점이 발견되면, 시스템은 자동으로 이슈를 생성해야 한다

### State-driven Requirements (상태 기반 기술)
- WHILE 개발 모드일 때, 시스템은 상세한 디버그 정보를 제공해야 한다
- WHILE 프로덕션 모드일 때, 시스템은 최적화된 로깅을 사용해야 한다
- WHILE CI 환경일 때, 시스템은 커버리지 리포트를 생성해야 한다

### Optional Features (선택적 기술)
- WHERE Docker 환경이면, 시스템은 컨테이너 기반 배포를 지원할 수 있다
- WHERE GitHub API 접근 가능하면, 시스템은 자동 PR 생성을 수행할 수 있다
- WHERE mypy strict 모드이면, 시스템은 타입 안전성을 100% 보장할 수 있다

### Constraints (기술적 제약사항)
- IF 의존성에 보안 취약점이 발견되면, 시스템은 빌드를 중단해야 한다
- 테스트 커버리지는 85% 이상을 유지해야 한다
- 빌드 시간은 10초를 초과하지 않아야 한다
- Python 버전은 3.13 미만을 지원하지 않아야 한다
```

---

_이 기술 스택은 `/alfred:2-build` 실행 시 TDD 도구 선택과 품질 게이트 적용의 기준이 됩니다._
