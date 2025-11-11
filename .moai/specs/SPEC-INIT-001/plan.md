---
id: INIT-001-PLAN
version: 0.0.1
created: 2025-10-06
updated: 2025-10-06
---

# @CODE:INIT-001 구현 계획

## 목차
1. [구현 전략 개요](#구현-전략-개요)
2. [6단계 상세 계획](#6단계-상세-계획)
3. [기술 스택 및 도구](#기술-스택-및-도구)
4. [테스트 전략](#테스트-전략)
5. [리스크 및 완화 방안](#리스크-및-완화-방안)
6. [완료 조건](#완료-조건)

---

## 구현 전략 개요

### 핵심 목표
GitHub Issue #2의 문제를 근본적으로 해결하면서 기존 대화형 사용자의 경험을 유지합니다.

### 전략적 접근
1. **점진적 개선**: 기존 코드 수정 최소화, 새 기능 모듈화
2. **하위 호환성**: 대화형 모드 기존 동작 보존
3. **플랫폼 독립성**: macOS, Linux, Windows 모두 지원
4. **자동화 우선**: 수동 개입 최소화

### 우선순위 기준
- **High**: TTY 감지, --yes 플래그, 선택적 의존성 분리, Docker 테스트
- **Medium**: 에러 메시지 개선, 자동 설치 프롬프트
- **Low**: --config 플래그 (선택적 기능)

---

## 6단계 상세 계획

### 1단계: TTY 자동 감지 + --yes 플래그
**우선순위**: High

#### 목표
- 비대화형 환경(Claude Code, CI/CD, Docker)에서 조용히 종료되는 문제 해결
- 사용자가 명시적으로 프롬프트를 건너뛸 수 있는 옵션 제공

#### 작업 항목
- [ ] `isTTYAvailable()` 유틸 함수 작성
  - 위치: `src/utils/tty.ts`
  - 로직: `process.stdin.isTTY && process.stdout.isTTY`
  - 예외 처리: try-catch로 안전한 폴백
- [ ] Commander.js 옵션 추가
  - `-y, --yes`: "Skip prompts and use default settings"
  - `--config <path>`: "Initialize from JSON config file" (향후 확장)
- [ ] `initWithDefaults()` 함수 작성
  - 기본값: `{ mode: "personal", gitEnabled: true }`
  - `.moai/config.json` 생성 로직
- [ ] 모드 분기 로직 구현
  ```typescript
  if (options.yes || !isTTYAvailable()) {
    initWithDefaults();
  } else {
    initWithPrompts(); // 기존 로직
  }
  ```

#### 테스트 케이스
- [ ] TTY 없는 환경 (process.stdin.isTTY = undefined) → 비대화형 모드
- [ ] `moai init --yes` → 프롬프트 스킵
- [ ] 일반 터미널 → 기존 프롬프트 유지

#### 기대 결과
- Claude Code에서 `moai init` 실행 시 조용히 종료하지 않고 기본값으로 초기화
- "Running in non-interactive mode with default settings" 메시지 출력

---

### 2단계: 의존성 자동 설치 프롬프트
**우선순위**: Medium

#### 목표
- 필수 의존성(Git, Node.js) 누락 시 자동 설치 제안
- 플랫폼별 최적 설치 방법 제공

#### 작업 항목
- [ ] `DependencyInstaller` 클래스 생성
  - 위치: `src/utils/dependency-installer.ts`
  - 메서드:
    - `installGit(platform: string): Promise<boolean>`
    - `installNodeJS(platform: string): Promise<boolean>`
    - `execWithTimeout(command: string, timeout: number): Promise<void>`
- [ ] 플랫폼별 설치 전략 구현
  - macOS: `brew install git`, `nvm install --lts`
  - Ubuntu: `sudo apt install git -y`, `nvm install --lts`
  - Windows: `winget install Git.Git`, 수동 가이드
- [ ] nvm 우선 사용 로직
  - `which nvm` 체크
  - nvm 있으면 `nvm install --lts` (sudo 불필요)
  - 없으면 플랫폼별 패키지 관리자 사용
- [ ] 대화형 프롬프트 구현
  - "Would you like to install [dependency] now? [Y/N/M]"
  - Y: 자동 설치
  - N: 초기화 중단
  - M: 수동 설치 명령어 출력 후 종료

#### 테스트 케이스
- [ ] macOS + Git 없음 → Homebrew 자동 설치
- [ ] Ubuntu + Node.js 없음 + nvm 있음 → nvm으로 설치
- [ ] Windows + Git 없음 → winget 가이드
- [ ] 사용자 'N' 선택 → 초기화 중단
- [ ] 타임아웃 5분 초과 → 설치 실패 처리

#### 기대 결과
- 의존성 누락 시 사용자 친화적인 자동 설치 옵션 제공
- sudo 사용 최소화 (nvm 우선)

---

### 3단계: 선택적 의존성 분리
**우선순위**: High

#### 목표
- Git LFS, Docker를 선택적(optional) 의존성으로 분류
- 선택적 의존성 실패 시에도 초기화 계속 진행

#### 작업 항목
- [ ] `moai doctor` 의존성 분류 수정
  - 위치: `src/cli/doctor.ts`
  - 분류:
    ```typescript
    const dependencies = {
      runtime: ['git', 'node'],
      development: ['npm'],
      optional: ['git-lfs', 'docker']
    };
    ```
- [ ] `allPassed` 로직 수정
  ```typescript
  const allPassed = checkRuntime() && checkDevelopment();
  // optional은 allPassed 판정에 영향 없음
  ```
- [ ] 선택적 의존성 경고 메시지
  - "⚠ Git LFS (optional, safe to ignore)"
  - "⚠ Docker (optional, required for container workflows)"

#### 테스트 케이스
- [ ] Git LFS 없음 → 경고만 출력, 초기화 성공
- [ ] Docker 없음 → 경고만 출력, 초기화 성공
- [ ] Git 없음 → 초기화 실패 (runtime 필수)

#### 기대 결과
- Docker 환경에서 Git LFS 없어도 초기화 성공
- 선택적 의존성 상태가 명확히 구분됨

---

### 4단계: 에러 메시지 개선
**우선순위**: Medium

#### 목표
- 초기화 실패 시 "System verification failed" 대신 구체적인 원인 표시
- 각 의존성별 설치 명령어 안내

#### 작업 항목
- [ ] 상세 에러 메시지 템플릿 작성
  ```
  ✗ System verification failed

  Missing dependencies:
    - Git (runtime): brew install git
    - npm (development): nvm install --lts

  Run 'moai doctor' for detailed diagnostics
  ```
- [ ] 플랫폼별 설치 명령어 매핑
  - macOS: brew
  - Ubuntu: apt
  - Windows: winget
- [ ] `moai doctor` 힌트 추가
  - "Run 'moai doctor' for detailed diagnostics"

#### 테스트 케이스
- [ ] Git + npm 없음 → 두 항목 모두 에러 메시지에 표시
- [ ] 각 플랫폼별 올바른 설치 명령어 출력 확인

#### 기대 결과
- 사용자가 초기화 실패 원인을 즉시 파악 가능
- 다음 액션이 명확함 (설치 명령어 제공)

---

### 5단계: Docker 멀티 플랫폼 테스트
**우선순위**: High

#### 목표
- Linux(Ubuntu 22.04), Windows(Server Core) 환경에서 로컬 테스트
- CI/CD 배포 전 3개 플랫폼 모두 검증

#### 작업 항목
- [ ] `Dockerfile.linux` 작성
  - Base: `ubuntu:22.04`
  - Node.js 설치
  - moai CLI 복사 및 테스트 실행
- [ ] `Dockerfile.windows` 작성
  - Base: `mcr.microsoft.com/windows/servercore:ltsc2022`
  - Node.js 설치
  - moai CLI 복사 및 테스트 실행
- [ ] `docker-compose.test.yml` 작성
  ```yaml
  version: '3.8'
  services:
    linux-test:
      build:
        dockerfile: Dockerfile.linux
      volumes:
        - ./.docker-test:/app/results
    windows-test:
      build:
        dockerfile: Dockerfile.windows
      volumes:
        - ./.docker-test:/app/results
  ```
- [ ] `scripts/test-local.sh` 작성
  - Docker Compose 실행
  - 결과 파일 출력
  - exit code 반환
- [ ] `.gitignore`에 `.docker-test/` 추가

#### 테스트 시나리오 (Linux)
- [ ] TTY 없는 환경에서 moai init 성공
- [ ] --yes 플래그로 초기화 성공
- [ ] Git 누락 시 자동 설치 제안
- [ ] 선택적 의존성 실패 시 초기화 계속
- [ ] 에러 메시지 출력 확인

#### 테스트 시나리오 (Windows)
- [ ] 비대화형 모드 기본값 초기화
- [ ] winget 기반 의존성 설치 가이드

#### 기대 결과
- `npm run test:docker` 실행 시 Linux, Windows 모두 테스트 통과
- `.docker-test/linux-results.json`, `.docker-test/windows-results.json` 생성
- macOS는 로컬 npm test로 검증

---

### 6단계: --config 플래그 (선택적)
**우선순위**: Low

#### 목표
- JSON 파일에서 설정을 읽어 초기화하는 옵션 제공 (향후 확장)

#### 작업 항목
- [ ] Commander.js `--config <path>` 옵션 추가
- [ ] JSON 파일 읽기 및 유효성 검증
  - 스키마 검증: mode, gitEnabled, features 등
  - 실패 시 기본값으로 폴백
- [ ] `.moai/config.json` 저장

#### 테스트 케이스
- [ ] `moai init --config custom.json` → 커스텀 설정 적용
- [ ] JSON 파일 오류 → 기본값으로 폴백

#### 기대 결과
- 고급 사용자가 비대화형 환경에서 커스텀 설정 사용 가능

---

## 기술 스택 및 도구

### 핵심 라이브러리
- **Commander.js**: CLI 프레임워크
- **inquirer**: 대화형 프롬프트 (기존)
- **execa**: 외부 명령 실행 (설치 스크립트)
- **fs-extra**: 파일 시스템 작업

### 플랫폼 도구
- **macOS**: Homebrew, nvm
- **Linux**: apt, nvm
- **Windows**: winget, nvm-windows

### 테스트 도구
- **Vitest**: 단위 테스트
- **Docker Compose**: 멀티 플랫폼 통합 테스트
- **GitHub Actions**: CI/CD (향후)

### 개발 도구
- **TypeScript**: 타입 안전성
- **Biome**: 린터 및 포맷터
- **ripgrep (rg)**: TAG 체인 검증

---

## 테스트 전략

### 단위 테스트 (Vitest)
**위치**: `tests/cli/init.test.ts`

#### 테스트 케이스
- [ ] `isTTYAvailable()` 정확성
  - TTY 있음 → true
  - TTY 없음 → false
  - 예외 발생 → false
- [ ] `initWithDefaults()` 기본값 저장
  - .moai/config.json 생성
  - { mode: "personal", gitEnabled: true } 확인
- [ ] Commander.js 옵션 파싱
  - `--yes` 플래그 인식
  - `--config` 플래그 인식

#### 목표 커버리지
- **전체**: ≥85%
- **핵심 로직**: ≥95% (TTY 감지, 의존성 설치)

---

### 통합 테스트 (Docker)
**실행**: `npm run test:docker`

#### Linux 시나리오 (5개)
1. TTY 없는 환경에서 moai init 성공
2. --yes 플래그로 초기화 성공
3. Git 누락 시 자동 설치 제안 (모의)
4. 선택적 의존성 실패 시 초기화 계속
5. 에러 메시지 출력 확인

#### Windows 시나리오 (2개)
1. 비대화형 모드 기본값 초기화
2. winget 기반 의존성 설치 가이드

#### macOS 시나리오 (로컬)
1. Homebrew 기반 자동 설치
2. nvm 우선 사용
3. 대화형 프롬프트 정상 동작

---

### E2E 테스트 (수동)
#### Claude Code 환경
- [ ] Claude Code에서 moai init 실행
- [ ] 비대화형 모드 자동 전환 확인
- [ ] .moai/config.json 생성 확인

#### CI/CD 환경
- [ ] GitHub Actions에서 moai init --yes 실행
- [ ] 의존성 자동 설치 확인
- [ ] 초기화 성공 확인

---

## 리스크 및 완화 방안

### 리스크 1: TTY 감지 오류
**증상**: 대화형 환경에서 비대화형으로 오판

**원인**: `process.stdin.isTTY` 값이 일부 환경에서 부정확

**완화 방안**:
- `stdin.isTTY && stdout.isTTY` 모두 체크
- try-catch로 예외 안전하게 처리
- 테스트: SSH, tmux, screen 환경에서 검증

**우선순위**: High

---

### 리스크 2: 자동 설치 실패
**증상**: brew install git 실패 (Homebrew 없음)

**원인**: 플랫폼별 도구 부재

**완화 방안**:
- 자동 설치 전 도구 존재 여부 체크 (`which brew`)
- 실패 시 수동 설치 가이드 출력
- 타임아웃 5분 설정

**우선순위**: Medium

---

### 리스크 3: Windows 테스트 부재
**증상**: Windows 환경에서 초기화 실패

**원인**: 로컬 개발 환경이 macOS/Linux

**완화 방안**:
- Docker Windows 컨테이너로 로컬 테스트
- GitHub Actions Windows runner 활용
- winget 명령어 사전 검증

**우선순위**: High

---

### 리스크 4: 기존 사용자 경험 변화
**증상**: 대화형 모드 사용자가 혼란

**원인**: 새 옵션 추가로 인한 UX 변화

**완화 방안**:
- 기존 대화형 모드 로직 변경 없음
- --yes 플래그는 선택적 (기본 동작 유지)
- CHANGELOG에 명확히 기록

**우선순위**: High

---

### 리스크 5: Docker 테스트 시간 증가
**증상**: 로컬 테스트가 느려짐 (5분+)

**원인**: 멀티 플랫폼 컨테이너 빌드

**완화 방안**:
- Docker layer 캐싱 활용
- 필요 시에만 `npm run test:docker` 실행 (CI/CD 전)
- 기본 `npm test`는 빠른 단위 테스트만

**우선순위**: Low

---

## 완료 조건 (Definition of Done)

### 기능 완료
- [ ] 1-6단계 모든 작업 항목 체크
- [ ] 9개 Acceptance Criteria 모두 통과

### 테스트 완료
- [ ] 단위 테스트 커버리지 ≥85%
- [ ] Docker 멀티 플랫폼 테스트 통과 (Linux, Windows)
- [ ] macOS 로컬 테스트 통과
- [ ] Claude Code 환경 E2E 테스트 성공

### 문서 완료
- [ ] CHANGELOG.md 업데이트
- [ ] docs/cli/init.md 작성
- [ ] GitHub Issue #2에 결과 보고
- [ ] TAG 체인 검증 (`rg '@(SPEC|TEST|CODE|DOC):INIT-001' -n`)

### 배포 준비
- [ ] Git 브랜치: `feature/INIT-001` 생성
- [ ] Draft PR 생성
- [ ] CI/CD 통과
- [ ] 코드 리뷰 완료

### TRUST 5원칙 검증
- [ ] **Test First**: 테스트 커버리지 ≥85%
- [ ] **Readable**: Biome 린트 통과, 함수 ≤50 LOC
- [ ] **Unified**: TypeScript 타입 안전성 확보
- [ ] **Secured**: 입력 검증, 경로 traversal 방지
- [ ] **Trackable**: @TAG:INIT-001 체인 무결성 확인

---

## 다음 단계

### 즉시 실행
```bash
/alfred:2-run INIT-001  # TDD 구현 시작
```

### 구현 후
```bash
/alfred:3-sync            # TAG 체인 검증 및 문서 동기화
```

### GitHub 연동
- Issue #2에 구현 완료 보고
- PR 상태 Draft → Ready 전환

---

_이 계획은 SPEC-First TDD 방법론에 따라 작성되었습니다._
_모든 우선순위는 마일스톤 기준이며, 시간 예측은 제외되었습니다._
