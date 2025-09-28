# MoAI-ADK Technology Stack

## @TECH:STACK-001 언어 & 런타임

### Python 3.10+ (주 언어 - 기존)

- **선택 이유**: Claude Code 생태계 통합, 풍부한 라이브러리, 빠른 프로토타이핑
- **지원 버전**: Python 3.10, 3.11, 3.12
- **배포 타겟**: PyPI, conda-forge, Docker 컨테이너
- **설치 요구사항**: pip, pipx, Git, SQLite3

### TypeScript 5.0+ (신규 추가 - SPEC-012)

- **선택 이유**: 고성능 CLI, 타입 안전성, Node.js 생태계 활용
- **지원 버전**: TypeScript 5.0+, Node.js 18.0+
- **배포 타겟**: npm, 바이너리 배포
- **설치 요구사항**: Node.js, npm/yarn

### 멀티 플랫폼 지원 현황

| 플랫폼      | Python 지원 | TypeScript 지원 | 검증 도구          | 주요 제약              |
| ----------- | ----------- | --------------- | ------------------ | ---------------------- |
| **Windows** | ✅ 완성     | ✅ 완성         | GitHub Actions CI  | None                   |
| **macOS**   | ✅ 완성     | ✅ 완성         | 로컬 테스트        | None                   |
| **Linux**   | ✅ 완성     | ✅ 완성         | GitHub Actions CI  | None                   |

## @TECH:FRAMEWORK-001 핵심 프레임워크 & 라이브러리

### 1. Python 런타임 의존성

```toml
# pyproject.toml
dependencies = [
    "click>=8.0.0",              # CLI 프레임워크
    "rich>=13.0.0",             # 터미널 UI/UX
    "pydantic>=2.0.0",          # 데이터 검증
    "jinja2>=3.0.0",            # 템플릿 엔진
    "gitpython>=3.1.0",         # Git 작업
    "pyyaml>=6.0.0",            # YAML 파싱
    "requests>=2.28.0",         # HTTP 클라이언트
]
```

### 2. TypeScript 런타임 의존성

```json
// package.json
{
  "dependencies": {
    "commander": "^11.1.0",      // CLI 프레임워크
    "chalk": "^4.1.2",          // 터미널 색상
    "inquirer": "^8.2.6",       // 대화형 프롬프트
    "semver": "^7.5.4",         // 버전 비교
    "execa": "^5.1.1"           // 프로세스 실행
  }
}
```

### 3. 개발 도구체인

```toml
# Python 개발 도구
dev = [
    "pytest>=7.0.0",            # 테스트 프레임워크
    "mypy>=1.0.0",              # 타입 검사
    "black>=23.0.0",            # 코드 포맷터
    "isort>=5.12.0",            # import 정렬
    "flake8>=6.0.0",            # 린터
    "pytest-cov>=4.0.0",       # 커버리지
]

# TypeScript 개발 도구
"devDependencies": {
    "typescript": "^5.0.0",     // 타입스크립트
    "tsup": "^8.0.0",           // 빌드 도구
    "jest": "^29.0.0",          // 테스트 프레임워크
    "eslint": "^8.0.0",         // 린터
    "prettier": "^3.0.0"        // 포맷터
}
```

### 4. 빌드 시스템

- **Python 빌드**: setuptools + build 기반 wheel 생성
- **TypeScript 빌드**: tsup 기반 ESM/CJS 듀얼 번들링 (686ms 성능)
- **버전 관리**: 통합 semantic versioning (MAJOR.MINOR.PATCH)
- **배포 자동화**: GitHub Actions → PyPI + npm 동시 배포

## @TECH:QUALITY-001 품질 게이트 & 정책

### 테스트 커버리지 목표

- **현재 상태**: Python 85%+, TypeScript 100% (SPEC-012 완료)
- **목표**: 전체 코드베이스 85% 이상 유지
- **측정 도구**: pytest-cov (Python), Jest coverage (TypeScript)
- **실패 시 대응**: PR 블록, 추가 테스트 작성 요구

### 정적 분석 도구

| 도구        | 역할          | 설정 파일         | 실패 시 조치         |
| ----------- | ------------- | ------------------ | -------------------- |
| **mypy**    | 타입 검사     | `mypy.ini`         | 타입 힌트 추가 요구  |
| **ESLint**  | TS 코드 품질  | `.eslintrc.json`   | 린트 오류 수정 요구  |
| **black**   | 코드 포맷     | `pyproject.toml`   | 자동 포맷팅 적용     |
| **Prettier** | TS 포맷      | `.prettierrc`      | 자동 포맷팅 적용     |

### 자동화 스크립트

```bash
# Python 품질 검사 파이프라인
pytest --cov=src tests/                    # 테스트 + 커버리지
mypy src/                                   # 타입 검사
black --check src/ tests/                  # 포맷 검사
isort --check-only src/ tests/             # import 정렬 검사
flake8 src/ tests/                         # 린트 검사

# TypeScript 품질 검사 파이프라인
npm run build                              # 빌드 검증
npm run test:coverage                      # 테스트 + 커버리지
npm run lint                               # ESLint 검사
npm run type-check                         # 타입 검사
```

## @TECH:SECURITY-001 보안 정책 & 운영

### 비밀 관리

- **정책**: GitHub Secrets, .env 파일 gitignore 필수
- **검증 도구**: pre-commit hooks, 정적 비밀 스캔
- **통합**: CI/CD에서 자동 비밀 검사 및 마스킹

### 접근 제어

```json
// .moai/security.json
{
  "security": {
    "allowedCommands": ["git", "npm", "python", "node"],
    "blockedPatterns": ["rm -rf", "sudo", "chmod 777"],
    "requireApproval": ["--force", "--hard"]
  }
}
```

### 로깅 정책

- **구조화 로깅**: JSON Lines 포맷 (Python: structlog, TypeScript: winston)
- **로그 수준**: DEBUG, INFO, WARNING, ERROR, CRITICAL
- **민감정보 마스킹**: API 키, 비밀번호, 토큰을 `***redacted***`로 마스킹
- **감사 로그**: 모든 CLI 명령어, Git 작업, 파일 변경 기록

## @TECH:DEPLOY-001 배포 채널 & 전략

### 1. Python 배포 (주 채널)

- **패키지명**: `moai-adk`
- **릴리스 절차**: GitHub Release → PyPI 자동 배포
- **버전 정책**: Semantic Versioning (현재 v0.1.28)
- **rollback 전략**: PyPI yanking + 이전 버전 재배포

### 1.1 TypeScript 배포 (신규 채널 - SPEC-012)

- **패키지명**: `moai-adk` (npm)
- **릴리스 절차**: GitHub Release → npm 자동 배포
- **버전 정책**: Python과 동기화된 버전 (현재 v0.0.1)
- **rollback 전략**: npm unpublish + 이전 버전 재배포

### 2. 개발 설치

```bash
# Python 개발자 모드
pip install -e .[dev]                      # editable 설치
pipx install --editable .

# TypeScript 개발자 모드
cd moai-adk-ts && npm install              # 의존성 설치
npm run build                              # 빌드
npm link                                   # 글로벌 링크

# 전체 개발 환경 (듀얼 스택)
git clone https://github.com/your-org/moai-adk.git
cd moai-adk
pip install -e .[dev]                     # Python 환경
cd moai-adk-ts && npm install && npm run build  # TypeScript 환경
```

### 3. 배포 채널 계획

| 채널            | 현재 상태 | 목표               | 배포 주기          |
| --------------- | --------- | ------------------ | ------------------ |
| **PyPI**        | ✅ 활성   | 안정화 배포        | 매 SPEC 완료 시    |
| **npm**         | 🚧 준비중 | TypeScript 배포    | Python과 동기화    |
| **conda-forge** | 📋 계획   | conda 생태계 지원  | 월 1회 안정 버전   |
| **Docker Hub**  | 📋 계획   | 컨테이너 배포      | 분기별 LTS 버전    |

## 현재 프로젝트 현황

### 기술 스택 현황 요약

- **Python v0.1.28**: 완전 기능, 70+ 모듈, 85%+ 커버리지, PyPI 배포 중
- **TypeScript v0.0.1**: SPEC-012 완료, 시스템 검증 모듈, CLI 기반 구축
- **Claude Code 통합**: 7개 에이전트, 5개 명령어, 8개 훅 완성
- **문서화**: MkDocs 자동 생성, API 문서 85개 모듈, 0.54초 빌드

### @DEBT:TECH-DEBT-001 기술 부채 개선 계획

1. **Python-TypeScript 브릿지**: 두 런타임 간 상태 동기화 매커니즘 구축
2. **단일 CLI 통합**: Python/TypeScript CLI를 하나의 명령어로 통합
3. **성능 최적화**: Python 4.6초 → 2초, TypeScript 686ms 유지

### @TASK:TECH-UPGRADE-001 기술 스택 업그레이드 후보

1. **Python 3.13 지원**: 2024년 10월 출시 예정, 성능 개선 15%+
2. **TypeScript 5.3+**: 최신 언어 기능, 더 나은 타입 추론
3. **Rust 백엔드**: 극고성능이 필요한 부분을 Rust로 포팅 (Week 5 계획)

## 환경별 설정

### 개발 환경 (`dev`)

```bash
# Python 개발 환경
export MOAI_MODE=development
export MOAI_LOG_LEVEL=DEBUG
pytest --cov=src tests/

# TypeScript 개발 환경
export NODE_ENV=development
npm run dev                               # tsx 기반 개발 서버
npm run test:watch                       # 감시 모드 테스트
```

### CI/CD 환경 (`ci`)

```bash
# GitHub Actions 환경
export MOAI_MODE=ci
export MOAI_LOG_LEVEL=INFO
pytest --cov=src --cov-report=xml tests/
npm ci && npm run build && npm test
```

### 프로덕션 환경 (`production`)

```bash
# 사용자 설치 환경
export MOAI_MODE=production
export MOAI_LOG_LEVEL=WARNING
moai init my-project                     # Python CLI
moai doctor                              # TypeScript CLI (Week 2+)
```

---

## @SUCCESS:TYPESCRIPT-INTEGRATION-012 TypeScript 스택 통합 완료 ✅

**SPEC-012 기술 스택 달성 지표:**
- ✅ TypeScript 5.0 strict 모드 완전 지원
- ✅ tsup 빌드 686ms 고성능 달성 (30초 목표 대비 99% 개선)
- ✅ Jest 테스트 프레임워크 100% 커버리지
- ✅ ESLint + Prettier 코드 품질 도구 통합
- ✅ 크로스 플랫폼 지원 (Windows/macOS/Linux)

**다음 단계 (Week 2-5):**
- Python-TypeScript 런타임 브릿지 구축
- 통합 CLI 명령어 체계 완성
- 고성능 하이브리드 아키텍처 달성

_이 기술 스택은 `/moai:2-build` 실행 시 TDD 도구 선택과 품질 게이트 적용의 기준이 됩니다._