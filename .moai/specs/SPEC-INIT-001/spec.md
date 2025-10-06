---
id: INIT-001
version: 0.1.0
status: draft
created: 2025-10-06
updated: 2025-10-06
author: @Goos
reference: .moai/reports/moai-adk-redesign-masterplan.md
related_issue: "https://github.com/modu-ai/moai-adk/issues/2"
---

# @SPEC:INIT-001: moai init 비대화형 환경 지원 및 의존성 자동 설치

## HISTORY

### v0.1.0 (2025-10-06)
- **INITIAL**: moai init 비대화형 환경 지원 및 의존성 자동 설치 명세 작성
- **AUTHOR**: @Goos
- **REVIEW**: @AI-Alfred
- **CONTEXT**: GitHub Issue #2 기반
- **SCOPE**: TTY 자동 감지, --yes 플래그, 의존성 자동 설치, 선택적 의존성 분리, Docker 멀티 플랫폼 테스트

---

## Environment (환경 및 전제조건)

### 실행 환경
- **Node.js**: ≥18.0.0 (LTS 권장)
- **OS**: macOS, Linux (Ubuntu 22.04+), Windows (Server Core)
- **TTY 환경**: 터미널(대화형), CI/CD(비대화형), Claude Code(비대화형), Docker(비대화형)
- **패키지 관리자**: npm, Homebrew(macOS), apt(Ubuntu), nvm, winget(Windows)

### 기술 스택
- **CLI Framework**: Commander.js
- **의존성 검증**: `moai doctor` 명령어 기반
- **테스트 환경**: Docker Compose (Linux, Windows 컨테이너)
- **파일 시스템**: `.moai/config.json` 기본 설정 파일

### 제약사항
- Node.js TTY 감지는 `process.stdin.isTTY`, `process.stdout.isTTY`를 사용
- nvm 설치 시 sudo 불필요 (사용자 홈 디렉토리 설치)
- Git LFS는 선택적 의존성으로 분류
- Docker 테스트 결과는 `.gitignore`로 제외

---

## Assumptions (가정사항)

1. **TTY 감지 가정**:
   - `process.stdin.isTTY === undefined` → 비대화형 환경
   - Claude Code, Docker, CI/CD는 모두 TTY 부재
   - SSH, tmux는 TTY 있지만 입력 제한 가능

2. **의존성 우선순위 가정**:
   - **필수(runtime)**: Git, Node.js
   - **개발(development)**: npm, Python (일부 스크립트)
   - **선택적(optional)**: Git LFS, Docker

3. **자동 설치 가정**:
   - macOS: Homebrew 존재 가정
   - Ubuntu: apt 사용 가정
   - Windows: winget 또는 수동 설치 안내
   - nvm 우선 사용 (sudo 회피)

4. **테스트 환경 가정**:
   - 로컬 Docker 설치 필수 (`npm run test:docker` 실행 시)
   - Linux, Windows 컨테이너 멀티 플랫폼 빌드 가능

---

## Requirements (EARS 요구사항)

### Ubiquitous Requirements (기본 기능)

**UR-001**: 시스템은 TTY 환경과 비대화형 환경 모두에서 `moai init` 실행을 지원해야 한다

**UR-002**: 시스템은 선택적 의존성(optional)과 필수 의존성(runtime, development)을 명확히 구분해야 한다

**UR-003**: 시스템은 초기화 실패 시 구체적인 원인을 사용자에게 표시해야 한다

**UR-004**: 시스템은 `.moai/config.json`에 초기화 설정을 저장해야 한다

**UR-005**: 시스템은 기본값으로 `{ mode: "personal", gitEnabled: true }`를 사용해야 한다

---

### Event-driven Requirements (이벤트 기반)

**ER-001**: WHEN TTY가 감지되지 않으면, 시스템은 비대화형 모드로 전환하여 기본값으로 초기화해야 한다
- **조건**: `process.stdin.isTTY === undefined`
- **동작**: 프롬프트 없이 기본값 사용
- **출력**: "Running in non-interactive mode with default settings" 메시지

**ER-002**: WHEN `--yes` 또는 `-y` 플래그가 제공되면, 시스템은 프롬프트 없이 기본값으로 초기화해야 한다
- **조건**: 명령줄 인자에 `--yes` 또는 `-y` 존재
- **동작**: TTY 존재 여부와 무관하게 프롬프트 스킵
- **출력**: "Initializing with default settings..." 메시지

**ER-003**: WHEN 필수 의존성(Git, Node.js)이 누락되면, 시스템은 자동 설치를 제안하고 사용자 동의 시 설치해야 한다
- **조건**: `moai doctor`가 runtime 의존성 실패 감지
- **동작**: "Would you like to install [dependency] now? [Y/N/M]" 프롬프트 (대화형 모드)
- **자동 설치**: Y 선택 시 플랫폼별 설치 명령 실행
- **수동 설치**: M 선택 시 설치 명령어 출력 후 종료

**ER-004**: WHEN 선택적 의존성(Git LFS)이 누락되면, 시스템은 경고를 표시하되 초기화를 계속해야 한다
- **조건**: `moai doctor`가 optional 의존성 실패 감지
- **동작**: 경고 메시지 출력 ("Git LFS not found (optional, safe to ignore)")
- **결과**: `allPassed = true` 유지 (초기화 계속)

**ER-005**: WHEN `--config` 플래그로 JSON 파일 경로가 제공되면, 시스템은 해당 파일에서 설정을 읽어 초기화할 수 있다 (선택적 기능)
- **조건**: `--config config.json` 형태 인자
- **동작**: JSON 파일 읽기 → 유효성 검증 → `.moai/config.json`에 저장
- **실패 처리**: JSON 파일 오류 시 기본값으로 폴백

---

### State-driven Requirements (상태 기반)

**SR-001**: WHILE 비대화형 모드일 때, 시스템은 `.moai/config.json`에 기본값을 저장해야 한다
- **상태**: `isTTY === false` 또는 `--yes` 플래그
- **동작**: `{ mode: "personal", gitEnabled: true }` 저장
- **검증**: 파일 생성 후 JSON 유효성 확인

**SR-002**: WHILE 대화형 모드일 때, 시스템은 기존 프롬프트 방식을 유지해야 한다
- **상태**: `isTTY === true` 및 `--yes` 플래그 없음
- **동작**: `inquirer.prompt()` 실행
- **보장**: 기존 사용자 경험 변화 없음

**SR-003**: WHILE 의존성 설치 중일 때, 시스템은 진행 상태를 사용자에게 표시해야 한다
- **상태**: `brew install git` 또는 `nvm install --lts` 실행 중
- **동작**: 실시간 stdout/stderr 출력
- **타임아웃**: 5분 초과 시 설치 실패 처리

---

### Optional Features (선택적 기능)

**OF-001**: WHERE `--config` 플래그가 제공되면, 시스템은 JSON 파일에서 설정을 읽어 초기화할 수 있다
- **조건**: 사용자가 `moai init --config custom.json` 실행
- **구현**: Commander.js `--config <path>` 옵션 추가
- **우선순위**: Low (향후 확장)

**OF-002**: WHERE macOS 환경이면, 시스템은 Homebrew를 통한 자동 설치를 제공할 수 있다
- **조건**: `process.platform === 'darwin'`
- **구현**: `brew install git` 자동 실행
- **폴백**: Homebrew 없으면 수동 설치 가이드

**OF-003**: WHERE nvm이 설치되어 있으면, 시스템은 nvm을 통한 Node.js 설치를 우선할 수 있다
- **조건**: `which nvm` 성공
- **구현**: `nvm install --lts` 실행
- **장점**: sudo 불필요

---

### Constraints (제약사항)

**C-001**: IF TTY 자동 감지가 실패하면, 시스템은 비대화형 모드로 폴백해야 한다
- **조건**: `process.stdin.isTTY` 값이 `undefined` 또는 예외 발생
- **동작**: 안전하게 비대화형 모드로 전환
- **이유**: 조용히 종료하는 것보다 기본값 초기화가 나음

**C-002**: IF 필수 의존성 검증이 실패하면, 시스템은 `.moai` 디렉토리를 생성하지 않아야 한다
- **조건**: Git 또는 Node.js 설치 실패 또는 거부
- **동작**: 초기화 중단, 에러 메시지 출력
- **정리**: 부분 생성된 파일 삭제

**C-003**: 선택적 의존성 실패는 `allPassed` 판정에 영향을 주지 않아야 한다
- **조건**: Git LFS 또는 Docker 미설치
- **동작**: `allPassed = (runtime && development)` 로직
- **결과**: 초기화 계속 진행

**C-004**: IF 자동 설치 실패 시, 시스템은 수동 설치 명령어를 안내해야 한다
- **조건**: `brew install git` 실패
- **동작**: "Please install Git manually:" + 명령어 출력
- **종료**: exit code 1

**C-005**: Docker 테스트 결과 파일은 Git 추적에서 제외해야 한다
- **조건**: `.docker-test/*.json` 생성
- **구현**: `.gitignore`에 `.docker-test/` 추가
- **이유**: 로컬 테스트 결과는 임시 파일

---

## Specifications (상세 명세)

### 1. TTY 자동 감지 구현

**함수 시그니처**:
```typescript
function isTTYAvailable(): boolean {
  return process.stdin.isTTY === true && process.stdout.isTTY === true;
}
```

**검증 로직**:
- `stdin.isTTY` AND `stdout.isTTY` 모두 true → 대화형 모드
- 하나라도 false/undefined → 비대화형 모드

**폴백 전략**:
- TTY 감지 예외 발생 시 `false` 반환 (안전한 폴백)

---

### 2. --yes 플래그 구현

**Commander.js 등록**:
```typescript
program
  .command('init')
  .option('-y, --yes', 'Skip prompts and use default settings')
  .option('--config <path>', 'Initialize from JSON config file')
  .action((options) => {
    if (options.yes || !isTTYAvailable()) {
      // 비대화형 모드
      initWithDefaults();
    } else {
      // 대화형 모드
      initWithPrompts();
    }
  });
```

**기본값**:
```json
{
  "mode": "personal",
  "gitEnabled": true
}
```

---

### 3. 의존성 자동 설치 구현

**DependencyInstaller 클래스 구조**:
```typescript
class DependencyInstaller {
  async installGit(platform: string): Promise<boolean>;
  async installNodeJS(platform: string): Promise<boolean>;
  private execWithTimeout(command: string, timeout: number): Promise<void>;
}
```

**플랫폼별 설치 전략**:
| 플랫폼 | Git 설치 | Node.js 설치 |
|--------|----------|-------------|
| macOS | `brew install git` | `nvm install --lts` → `brew install node` |
| Ubuntu | `sudo apt install git -y` | `nvm install --lts` → `sudo apt install nodejs -y` |
| Windows | `winget install Git.Git` | `nvm install lts` → 수동 가이드 |

**타임아웃**: 5분 (300,000ms)

---

### 4. 선택적 의존성 분리

**moai doctor 수정**:
```typescript
const dependencies = {
  runtime: ['git', 'node'],
  development: ['npm'],
  optional: ['git-lfs', 'docker']
};

const allPassed = checkRuntime() && checkDevelopment();
// optional은 allPassed 판정에 영향 없음
```

**에러 메시지**:
```
✓ Git (runtime)
✓ Node.js (runtime)
✓ npm (development)
⚠ Git LFS (optional, safe to ignore)
```

---

### 5. Docker 멀티 플랫폼 테스트

**docker-compose.test.yml**:
```yaml
version: '3.8'
services:
  linux-test:
    build:
      context: .
      dockerfile: Dockerfile.linux
    volumes:
      - ./.docker-test:/app/results
    command: npm run test:init

  windows-test:
    build:
      context: .
      dockerfile: Dockerfile.windows
    volumes:
      - ./.docker-test:/app/results
    command: npm run test:init
```

**테스트 스크립트**:
```bash
# scripts/test-local.sh
docker-compose -f docker-compose.test.yml up --build
cat .docker-test/linux-results.json
cat .docker-test/windows-results.json
```

---

### 6. 에러 메시지 개선

**실패 시 출력 예시**:
```
✗ System verification failed

Missing dependencies:
  - Git (runtime): brew install git
  - npm (development): nvm install --lts

Run 'moai doctor' for detailed diagnostics
```

---

## Acceptance Criteria (수락 기준)

### AC1: TTY 자동 감지
```gherkin
GIVEN Claude Code 환경 (process.stdin.isTTY === undefined)
WHEN moai init 실행
THEN 시스템은 비대화형 모드로 자동 전환
AND 프롬프트 없이 기본값으로 초기화
AND "Running in non-interactive mode with default settings" 메시지 출력
AND .moai/config.json 파일 생성
AND { mode: "personal", gitEnabled: true } 저장 확인
```

### AC2: --yes 플래그 지원
```gherkin
GIVEN TTY 환경 또는 비대화형 환경
WHEN moai init --yes 실행
THEN 시스템은 프롬프트를 건너뛰고 기본값으로 초기화
AND .moai/config.json에 { mode: "personal", gitEnabled: true } 저장
AND "Initializing with default settings..." 메시지 출력
AND 초기화 완료 메시지 출력
```

### AC3: Git 자동 설치 제안 (대화형)
```gherkin
GIVEN macOS 환경에서 Git이 설치되지 않음
AND TTY 환경 (대화형 가능)
WHEN moai init 실행
THEN "Git is required but not installed" 메시지 표시
AND "Would you like to install Git now? [Y/N/M]" 프롬프트 표시
WHEN 사용자가 'Y' 입력
THEN brew install git 실행
AND 설치 완료 후 초기화 계속 진행
WHEN 사용자가 'N' 입력
THEN 초기화 중단
AND 수동 설치 가이드 출력
WHEN 사용자가 'M' 입력
THEN "brew install git" 명령어 출력
AND 초기화 중단
```

### AC4: Node.js 자동 설치 제안 (nvm 우선)
```gherkin
GIVEN 환경에서 Node.js가 설치되지 않음
AND nvm이 설치되어 있음
WHEN moai init 실행
AND 사용자가 자동 설치 동의
THEN nvm install --lts 실행
AND Node.js 20 LTS 설치
AND node --version 검증 성공
AND 초기화 계속 진행
```

### AC5: 상세 에러 메시지
```gherkin
GIVEN Git과 npm이 설치되지 않은 환경
WHEN moai init 실행
THEN 실패한 의존성 목록 출력:
  - Git (runtime)
  - npm (development)
AND 각 의존성별 설치 명령어 표시
AND "Run 'moai doctor' for detailed diagnostics" 힌트 제공
AND exit code 1
```

### AC6: 기존 대화형 사용자 경험 유지
```gherkin
GIVEN 기존 사용자가 터미널에서 moai init 실행
WHEN TTY 환경이 감지되고 --yes 플래그 없음
THEN 기존 대화형 프롬프트 방식 유지
AND 사용자 경험 변화 없음
AND inquirer.prompt() 정상 실행
```

### AC7: Docker 멀티 플랫폼 테스트
```gherkin
GIVEN 로컬 개발 환경
WHEN npm run test:docker 실행
THEN Docker Compose가 Linux와 Windows 컨테이너를 빌드
AND 각 컨테이너에서 moai init 테스트 실행
AND .docker-test/linux-results.json 생성
AND .docker-test/windows-results.json 생성
AND 모든 테스트 통과 시 exit 0 반환
```

### AC8: 로컬 검수 워크플로우
```gherkin
GIVEN 코드 수정 완료
WHEN npm run test:docker 실행
THEN Linux: 5개 테스트 시나리오 통과
  - TTY 없는 환경에서 moai init 성공
  - --yes 플래그로 초기화 성공
  - Git 누락 시 자동 설치 제안
  - 선택적 의존성 실패 시 초기화 계속
  - 에러 메시지 출력 확인
AND Windows: 2개 테스트 시나리오 통과
  - 비대화형 모드 기본값 초기화
  - winget 기반 의존성 설치 가이드
AND macOS: 로컬 npm test 수동 실행으로 확인
  - Homebrew 기반 자동 설치
  - nvm 우선 사용
AND 모든 플랫폼 테스트 통과 시 배포 준비 완료
```

### AC9: 선택적 의존성 분리
```gherkin
GIVEN Git LFS가 설치되지 않은 환경
WHEN moai init 실행
THEN moai doctor가 Git LFS를 optional로 분류
AND allPassed = true (runtime + development만 검증)
AND "Git LFS not found (optional, safe to ignore)" 경고 출력
AND 초기화 성공
AND .moai/config.json 생성 확인
```

---

## Traceability (@TAG 체인)

### TAG 체인 구조
```
@SPEC:INIT-001 (본 문서)
  ↓
@TEST:INIT-001 (tests/cli/init.test.ts)
  ↓
@CODE:INIT-001 (src/cli/init.ts)
  ├─ @CODE:INIT-001:CLI (Commander.js 등록)
  ├─ @CODE:INIT-001:TTY (TTY 감지 로직)
  ├─ @CODE:INIT-001:INSTALLER (의존성 자동 설치)
  └─ @CODE:INIT-001:DOCTOR (선택적 의존성 분리)
  ↓
@DOC:INIT-001 (docs/cli/init.md, CHANGELOG.md)
```

### 검증 명령어
```bash
# SPEC 문서 확인
rg '@SPEC:INIT-001' -n .moai/specs/

# 테스트 파일 확인
rg '@TEST:INIT-001' -n tests/

# 구현 코드 확인
rg '@CODE:INIT-001' -n src/

# 문서 확인
rg '@DOC:INIT-001' -n docs/

# 전체 TAG 체인 무결성 검증
rg '@(SPEC|TEST|CODE|DOC):INIT-001' -n
```

---

## 다음 단계

### 구현 단계
1. `/alfred:2-build INIT-001` 실행 (TDD 구현)
2. RED: 테스트 작성 및 실패 확인
3. GREEN: 구현 및 테스트 통과
4. REFACTOR: 코드 품질 개선

### 동기화 단계
1. `/alfred:3-sync` 실행 (문서 동기화)
2. TAG 체인 검증
3. Living Document 생성
4. GitHub Issue #2 연동

---

_이 문서는 SPEC-First TDD 방법론에 따라 작성되었습니다._
_관련 이슈: https://github.com/modu-ai/moai-adk/issues/2_
