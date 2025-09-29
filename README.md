# MoAI-ADK (Agentic Development Kit)

[![Version](https://img.shields.io/badge/version-v0.0.1-blue)](https://github.com/modu-ai/moai-adk)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.2+-blue)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/node-18.0+-green)](https://nodejs.org/)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-integrated-purple)](https://docs.anthropic.com/claude-code)

**SPEC-First TDD 개발을 위한 체계적 AI 보조 개발 프레임워크**

## 개요

MoAI-ADK는 Claude Code 환경에서 SPEC-First TDD 개발 방법론을 지원하는 TypeScript 기반 CLI 도구입니다. AI와 함께하는 체계적이고 자동화된 개발 워크플로우를 제공하여 일관된 품질과 완전한 추적성을 보장합니다.

### 주요 기능

- **4단계 개발 워크플로우**: 프로젝트 킥오프 → SPEC 작성 → TDD 구현 → 문서 동기화
- **8개 전문 에이전트**: 각 개발 단계별 전문화된 AI 보조 시스템
- **@AI-TAG 시스템**: 코드 기반 완전한 추적성 관리
- **범용 언어 지원**: Python, TypeScript, Java, Go, Rust 등 주요 언어 지원
- **시스템 자동 진단**: 개발 환경 요구사항 자동 검증 및 설정

### 지원 환경

- **운영체제**: Windows, macOS, Linux
- **Node.js**: 18.0 이상
- **TypeScript**: 5.9.2 이상
- **Claude Code**: 필수 (에이전트 시스템 연동)

## 설치 가이드

### 1. 시스템 요구사항 확인

다음 도구들이 설치되어 있어야 합니다:

```bash
node --version    # v18.0.0 이상
git --version     # 2.30.0 이상
npm --version     # 8.0.0 이상
```

### 2. MoAI-ADK 설치

```bash
# npm을 사용한 글로벌 설치
npm install -g moai-adk

# 또는 Bun 사용 (권장)
bun add -g moai-adk
```

### 3. 설치 검증

```bash
# 버전 확인
moai --version

# 시스템 진단 실행
moai doctor
```

시스템 진단에서 모든 요구사항이 충족되었다면 설치가 완료되었습니다.

## 기본 사용법

### 새 프로젝트 시작하기

```bash
# 1. 새 프로젝트 생성
moai init my-project

# 2. 프로젝트 디렉터리로 이동
cd my-project

# 3. 프로젝트 구조 확인
ls -la
```

생성된 프로젝트는 다음 구조를 가집니다:

```
my-project/
├── .moai/                 # MoAI-ADK 설정 및 문서
│   ├── project/          # 프로젝트 정의 문서
│   ├── memory/           # 개발 가이드
│   └── specs/            # SPEC 문서 저장소
├── .claude/              # Claude Code 통합 설정
│   ├── agents/           # 8개 전문 에이전트
│   ├── commands/         # 워크플로우 명령어
│   └── hooks/            # 자동화 훅
└── CLAUDE.md             # 프로젝트 개발 가이드
```

### 기본 CLI 명령어

```bash
# 프로젝트 상태 확인
moai status

# 시스템 진단
moai doctor

# 백업 목록 확인
moai doctor --list-backups

# 도움말
moai help
```

## 4단계 개발 워크플로우

MoAI-ADK는 Claude Code 환경에서 다음과 같은 4단계 워크플로우를 제공합니다.

### Stage 0: 프로젝트 킥오프

```bash
/moai:0-project PROJECT_NAME
```

**목적**: 프로젝트 기초 컨텍스트 설정

**수행 작업**:
- 제품 정의서 생성 (`.moai/project/product.md`)
- 시스템 구조 설계서 생성 (`.moai/project/structure.md`)
- 기술 스택 문서 생성 (`.moai/project/tech.md`)
- Claude Code 환경 최적화
- @AI-TAG 시스템 초기화

**담당 에이전트**: `project-manager`

### Stage 1: SPEC 작성

```bash
/moai:1-spec "기능명1" "기능명2" ...    # 새 SPEC 작성
/moai:1-spec SPEC-001 "수정내용"       # 기존 SPEC 수정
/moai:1-spec                           # 자동 제안 (권장)
```

**목적**: EARS 형식 명세 작성 및 개발 준비

**수행 작업**:
- 프로젝트 문서 분석 및 SPEC 후보 제안
- EARS (Easy Approach to Requirements Syntax) 명세 작성
- @AI-TAG 체인 생성 (@REQ → @DESIGN → @TASK → @TEST)
- Git 브랜치 자동 생성 (Personal/Team 모드별)
- GitHub Issue/PR 템플릿 생성 (Team 모드)

**담당 에이전트**: `spec-builder`

### Stage 2: TDD 구현

```bash
/moai:2-build SPEC-001    # 특정 SPEC 구현
/moai:2-build all         # 모든 SPEC 구현
```

**목적**: Red-Green-Refactor 사이클로 TDD 구현

**수행 작업**:
- 프로젝트 언어 자동 감지 및 최적 도구 선택
- 언어별 테스트 도구 자동 선택 (pytest, Vitest, JUnit, go test, cargo test 등)
- Red-Green-Refactor 사이클 실행
- @AI-TAG 시스템으로 완전한 추적성 보장
- 16-Core @TAG 체인 자동 적용
- TRUST 5원칙 자동 검증
- 체크포인트 자동 생성

**담당 에이전트**: `code-builder`

### Stage 3: 문서 동기화

```bash
/moai:3-sync [mode] [target-path]
```

**목적**: 코드와 문서 동기화 및 완료 처리

**수행 작업**:
- Living Document 자동 동기화
- @AI-TAG 인덱스 재구축 (`.moai/indexes/tags.json`)
- sync-report 생성 (`.moai/reports/sync-report.md`)
- API 문서 자동 생성/갱신
- PR Draft → Ready 상태 전환
- 리뷰어 자동 할당 (Team 모드)

**담당 에이전트**: `doc-syncer`

## 에이전트 시스템

MoAI-ADK는 8개의 전문 에이전트를 제공하여 각 개발 단계를 지원합니다.

### 핵심 에이전트 (8개)

| 에이전트 | 역할 | 자동화 |
|---------|------|--------|
| **spec-builder** | EARS 명세 작성 | 브랜치/PR 생성 |
| **code-builder** | 범용 언어 TDD 구현 | Red-Green-Refactor (Python, TypeScript, Java, Go, Rust 등) |
| **doc-syncer** | 문서 동기화 | PR 상태 전환/라벨링 |
| **cc-manager** | Claude Code 관리 | 설정 최적화/권한 |
| **debug-helper** | 오류 진단 | 개발 가이드 검사 |
| **git-manager** | Git 자동화 | 브랜치, 커밋, PR, 체크포인트 관리 |
| **trust-checker** | 품질 검증 | TRUST 5원칙 검사, 코드 품질 분석 |
| **project-manager** | 프로젝트 킥오프 | 기초 컨텍스트 설정 |

### 에이전트 사용 예제

```bash
# 오류 분석 및 디버깅
@agent-debug-helper "TypeError: Cannot read property 'name' of undefined"
@agent-debug-helper "빌드 실패 원인 분석해주세요"

# TDD 구현 요청
@agent-code-builder "SPEC-001 구현 계획 분석"
@agent-code-builder "테스트 케이스 작성 및 구현 시작"

# 문서 동기화
@agent-doc-syncer "코드 변경사항을 문서에 반영"
@agent-doc-syncer "TAG 인덱스 업데이트"

# Git 작업
@agent-git-manager "feature 브랜치 생성 및 체크포인트 설정"
@agent-git-manager "현재 작업을 커밋하고 PR 준비"

# 품질 검증
@agent-trust-checker "TRUST 5원칙 준수 여부 검사"
@agent-trust-checker "코드 복잡도 및 가독성 분석"

# TAG 관리
@agent-tag-agent "현재 프로젝트의 TAG 체인 분석"
@agent-tag-agent "누락된 TAG 검사 및 보완"
```

## CLI 명령어 레퍼런스

### moai init

프로젝트를 초기화합니다.

```bash
moai init <project-name> [options]
```

**옵션**:
- `--type <type>`: 프로젝트 타입 (web-api, library, cli, mobile)
- `--language <lang>`: 주 언어 (python, typescript, java, go, rust)
- `--template <template>`: 템플릿 (basic, advanced, enterprise)
- `--backup`: 초기화 전 백업 생성
- `--force`: 기존 디렉터리 덮어쓰기
- `--verbose`: 상세 로그 출력

**예제**:
```bash
moai init my-api --type web-api --language typescript
moai init my-lib --type library --template advanced --backup
```

### moai doctor

시스템 요구사항을 진단합니다.

```bash
moai doctor [options]
```

**옵션**:
- `--list-backups`: 사용 가능한 백업 목록 표시

**예제**:
```bash
moai doctor
moai doctor --list-backups
```

### moai status

프로젝트 상태를 확인합니다.

```bash
moai status [options]
```

**옵션**:
- `--detailed`: 상세 상태 정보 표시
- `--tags`: TAG 체인 분석 결과 포함
- `--specs`: SPEC 완성도 표시
- `--git`: Git 상태 정보 포함

### moai update

MoAI-ADK 템플릿을 업데이트합니다.

```bash
moai update [options]
```

**옵션**:
- `--check`: 업데이트 가능 여부만 확인
- `--backup`: 업데이트 전 백업 생성
- `--force`: 강제 업데이트

### moai restore

백업에서 프로젝트를 복원합니다.

```bash
moai restore [backup-path] [options]
```

**옵션**:
- `--list`: 사용 가능한 백업 목록 표시
- `--preview`: 복원될 내용 미리보기
- `--force`: 현재 내용 덮어쓰기

### moai help

도움말을 표시합니다.

```bash
moai help [command]
```

**예제**:
```bash
moai help           # 전체 도움말
moai help init      # init 명령어 도움말
moai help doctor    # doctor 명령어 도움말
```

## @AI-TAG 시스템

MoAI-ADK는 코드 기반 완전한 추적성을 위해 @AI-TAG 시스템을 사용합니다.

### @AI-TAG Lifecycle 2.0

**Primary Chain (필수 체인)**:
- `@REQ`: 요구사항 정의
- `@DESIGN`: 아키텍처 설계
- `@TASK`: 구현 작업
- `@TEST`: 테스트 검증

**Implementation (선택적)**:
- `@FEATURE`: 비지니스 기능
- `@API`: 인터페이스
- `@UI`: 사용자 인터페이스
- `@DATA`: 데이터 처리

**Quality (품질 속성)**:
- `@PERF`: 성능 최적화
- `@SEC`: 보안 강화
- `@DOCS`: 문서화
- `@DEBT`: 기술 부채

### TAG BLOCK 템플릿 (필수)

코드 파일 상단에 다음과 같은 TAG BLOCK을 작성합니다:

```text
# @FEATURE:<DOMAIN-ID> | Chain: @REQ:<ID> -> @DESIGN:<ID> -> @TASK:<ID> -> @TEST:<ID>
# Related: @SEC:<ID>, @DOCS:<ID>
```

**TAG BLOCK 가이드라인**:
- 새 코드/문서/테스트 파일 생성 시: TAG BLOCK을 파일 상단(주석)에 배치
- 수정 시: 기존 TAG BLOCK 검토 후 영향받는 TAG 업데이트
- 생성 전 중복 확인: `rg "@REQ:<키워드>" -n`로 기존 체인 검색
- TAG ID: `<도메인>-<3자리>` (예: `AUTH-003`) — 체인 내 모든 TAG는 동일 ID 사용

### 언어별 @AI-TAG 적용 예제

**Python**:
```python
# @FEATURE:LOGIN-001 | Chain: @REQ:AUTH-001 -> @DESIGN:AUTH-001 -> @TASK:AUTH-001 -> @TEST:AUTH-001
# Related: @SEC:LOGIN-001, @DOCS:LOGIN-001
class AuthenticationService:
    """@FEATURE:LOGIN-001: 사용자 인증 서비스 구현"""

    def authenticate(self, username: str, password: str) -> bool:
        """@API:LOGIN-001: 사용자 인증 API 엔드포인트"""
        # @SEC:LOGIN-001: 입력값 보안 검증
        if not self._validate_input(username, password):
            return False
        return self._verify_credentials(username, password)

# @TEST:LOGIN-001 연결: @TASK:LOGIN-001 -> @TEST:LOGIN-001
def test_should_authenticate_valid_user():
    """@TEST:LOGIN-001: 유효한 사용자 인증 테스트"""
    service = AuthenticationService()
    result = service.authenticate("user", "password")
    assert result is True
```

**TypeScript**:
```typescript
// @FEATURE:LOGIN-001 | Chain: @REQ:AUTH-001 -> @DESIGN:AUTH-001 -> @TASK:AUTH-001 -> @TEST:AUTH-001
// Related: @SEC:LOGIN-001, @DOCS:LOGIN-001
interface AuthService {
  // @API:LOGIN-001: 인증 API 인터페이스 정의
  authenticate(username: string, password: string): Promise<boolean>;
}

// @UI:LOGIN-001: 로그인 컴포넌트
const LoginForm: React.FC = () => {
  // @SEC:LOGIN-001: 클라이언트 사이드 입력 검증
  const handleSubmit = (username: string, password: string) => {
    // 구현...
  };

  return <form>...</form>;
};

// @TEST:LOGIN-001: Vitest/Jest 테스트
describe('AuthService', () => {
  test('@TEST:LOGIN-001: should authenticate valid user', () => {
    // 테스트 구현...
  });
});
```

## TRUST 5원칙 (범용 언어 지원)

MoAI-ADK는 다음 TRUST 5원칙을 모든 주요 프로그래밍 언어에서 준수합니다:

### T - Test First (테스트 우선)
- 모든 구현은 테스트부터 작성 (SPEC-First TDD)
- Red-Green-Refactor 사이클 준수
- 언어별 최적 도구 (Jest/Vitest, pytest, go test, cargo test, JUnit 등)

### R - Readable (가독성)
- 함수 크기 50줄 이하 유지
- 명확한 함수/변수 네이밍 (SPEC 용어 반영)
- 언어별 린터 (ESLint/Biome, ruff, golint, clippy 등)
- 의도를 드러내는 코드 구조

### U - Unified (단일 책임)
- 모듈당 300줄 이하 유지 (SPEC 기준)
- 각 모듈의 명확한 역할 분담
- 타입 안전성 (TypeScript, Go, Rust, Java) 또는 런타임 검증 (Python, JS)
- 낮은 결합도, 높은 응집도

### S - Secured (보안성)
- SPEC Security Requirements: 모든 SPEC에 보안 요구사항 정의
- 모든 외부 입력 검증 (SPEC 인터페이스 기반)
- 언어별 보안 도구 및 정적 분석
- 민감 정보 자동 마스킹 및 감사 로깅

### T - Trackable (추적성)
- @AI-TAG 시스템으로 완전한 추적성
- SPEC-to-Code Traceability: 모든 코드 변경이 SPEC ID 참조
- 3단계 워크플로우 추적: `/moai:1-spec` → `/moai:2-build` → `/moai:3-sync`
- JSON 기반 @AI-TAG 시스템 인덱스 관리

## 범용 언어 지원

MoAI-ADK는 다음 언어들을 지원하며, 각 언어별 최적 도구를 자동으로 선택합니다.

| 언어 | 테스트 도구 | 린터/포맷터 | 빌드 도구 |
|------|------------|-------------|----------|
| **TypeScript** | Vitest/Jest | Biome/ESLint | tsup/Vite |
| **Python** | pytest | ruff/black | uv/pip |
| **Java** | JUnit | checkstyle | Maven/Gradle |
| **Go** | go test | golint/gofmt | go mod |
| **Rust** | cargo test | clippy/rustfmt | cargo |
| **C++** | GoogleTest | clang-format | CMake |
| **C#** | xUnit | dotnet-format | dotnet |
| **PHP** | PHPUnit | PHP-CS-Fixer | Composer |

## 문제 해결

### 자주 발생하는 문제

**1. 설치 실패**

```bash
# 권한 문제 해결
sudo npm install -g moai-adk

# 캐시 클리어 후 재설치
npm cache clean --force
npm install -g moai-adk
```

**2. moai 명령어 인식 안 됨**

```bash
# PATH 확인
echo $PATH

# npm 전역 설치 경로 확인
npm list -g --depth=0

# 셸 재시작
source ~/.bashrc  # 또는 source ~/.zshrc
```

**3. 시스템 진단 실패**

```bash
# 시스템 진단 재실행
moai doctor

# 개별 도구 버전 확인
node --version
git --version
npm --version
```

**4. Claude Code 연동 문제**

- `.claude/settings.json` 파일 확인
- Claude Code 최신 버전 사용 여부 확인
- 에이전트 파일 권한 확인

### 로그 확인

MoAI-ADK 로그는 다음 위치에 저장됩니다:

```bash
# 일반 로그
~/.moai/logs/moai.log

# 에러 로그
~/.moai/logs/error.log

# 프로젝트별 로그
.moai/logs/
```

## 개발 참여

### 기여 방법

1. GitHub Repository fork
2. 기능 브랜치 생성 (`git checkout -b feature/새기능`)
3. 변경사항 커밋 (`git commit -am '새기능 추가'`)
4. 브랜치 푸시 (`git push origin feature/새기능`)
5. Pull Request 생성

### 개발 환경 설정

```bash
# 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# TypeScript 프로젝트로 이동
cd moai-adk-ts

# 의존성 설치
npm install

# 개발 모드 실행
npm run dev -- --help

# 빌드
npm run build

# 테스트
npm test
```

### 코딩 규칙

- TRUST 5원칙 준수
- @TAG 시스템 적용
- TypeScript strict 모드 사용
- 함수당 50줄 이하 유지
- 명확한 함수/변수 네이밍

## 라이선스 및 지원

### 라이선스

이 프로젝트는 [MIT License](LICENSE)를 따릅니다.

### 지원 및 문의

- **GitHub Issues**: [https://github.com/modu-ai/moai-adk/issues](https://github.com/modu-ai/moai-adk/issues)
- **GitHub Discussions**: [https://github.com/modu-ai/moai-adk/discussions](https://github.com/modu-ai/moai-adk/discussions)
- **Documentation**: [https://moai-adk.github.io](https://moai-adk.github.io)

---

**MoAI-ADK v0.0.1 - SPEC-First TDD 개발 프레임워크**

*"명세가 없으면 코드도 없다. 테스트가 없으면 구현도 없다."*