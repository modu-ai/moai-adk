---
id: INIT-001-ACCEPTANCE
version: 0.0.1
created: 2025-10-06
updated: 2025-10-06
---


## 목차
1. [AC1: TTY 자동 감지](#ac1-tty-자동-감지)
2. [AC2: --yes 플래그 지원](#ac2---yes-플래그-지원)
3. [AC3: Git 자동 설치 제안 (대화형)](#ac3-git-자동-설치-제안-대화형)
4. [AC4: Node.js 자동 설치 제안 (nvm 우선)](#ac4-nodejs-자동-설치-제안-nvm-우선)
5. [AC5: 상세 에러 메시지](#ac5-상세-에러-메시지)
6. [AC6: 기존 대화형 사용자 경험 유지](#ac6-기존-대화형-사용자-경험-유지)
7. [AC7: Docker 멀티 플랫폼 테스트](#ac7-docker-멀티-플랫폼-테스트)
8. [AC8: 로컬 검수 워크플로우](#ac8-로컬-검수-워크플로우)
9. [AC9: 선택적 의존성 분리](#ac9-선택적-의존성-분리)
10. [검증 체크리스트](#검증-체크리스트)

---

## AC1: TTY 자동 감지

### Given-When-Then 시나리오

```gherkin
GIVEN Claude Code 환경
  AND process.stdin.isTTY === undefined
  AND process.stdout.isTTY === undefined
WHEN 사용자가 moai init 실행
THEN 시스템은 isTTYAvailable() 함수를 호출
  AND 함수는 false를 반환
  AND 시스템은 비대화형 모드로 자동 전환
  AND 프롬프트 없이 기본값으로 초기화
  AND 콘솔에 "Running in non-interactive mode with default settings" 출력
  AND .moai/config.json 파일 생성
  AND 파일 내용:
    {
      "mode": "personal",
      "gitEnabled": true
    }
```

### 상세 테스트 시나리오

#### 시나리오 1.1: Claude Code 환경
**환경**:
- Claude Code 터미널
- TTY 없음

**실행**:
```bash
moai init
```

**기대 출력**:
```
Running in non-interactive mode with default settings
Creating .moai directory...
✓ Project initialized successfully
```

**검증**:
- [ ] `.moai/config.json` 파일 존재
- [ ] JSON 파싱 성공
- [ ] `mode === "personal"`
- [ ] `gitEnabled === true`

---

#### 시나리오 1.2: Docker 컨테이너
**환경**:
- Docker 내부
- TTY 없음

**실행**:
```bash
docker run --rm -it moai-test moai init
```

**기대 출력**:
```
Running in non-interactive mode with default settings
✓ Project initialized successfully
```

**검증**:
- [ ] exit code 0
- [ ] `.moai/config.json` 생성
- [ ] 에러 로그 없음

---

#### 시나리오 1.3: CI/CD 환경
**환경**:
- GitHub Actions
- TTY 없음

**실행**:
```yaml
- run: npx moai init
```

**기대 동작**:
- 조용히 종료하지 않음
- 기본값으로 초기화
- 워크플로우 계속 진행

**검증**:
- [ ] GitHub Actions 성공
- [ ] 다음 스텝 실행됨

---

### 검증 방법

#### 단위 테스트
```typescript
// tests/cli/init.test.ts
describe('isTTYAvailable', () => {
  it('should return false when stdin.isTTY is undefined', () => {
    // Mock process.stdin.isTTY
    Object.defineProperty(process.stdin, 'isTTY', {
      value: undefined,
      configurable: true
    });

    expect(isTTYAvailable()).toBe(false);
  });

  it('should return true when both stdin and stdout are TTY', () => {
    Object.defineProperty(process.stdin, 'isTTY', { value: true });
    Object.defineProperty(process.stdout, 'isTTY', { value: true });

    expect(isTTYAvailable()).toBe(true);
  });
});
```

#### 통합 테스트 (Docker)
```bash
# Dockerfile.linux
FROM ubuntu:22.04
RUN npm install -g moai-adk
CMD ["moai", "init"]

# 실행
docker build -f Dockerfile.linux -t moai-test .
docker run --rm moai-test
```

---

## AC2: --yes 플래그 지원

### Given-When-Then 시나리오

```gherkin
GIVEN 터미널 환경 (TTY 있음 또는 없음)
WHEN 사용자가 moai init --yes 실행
THEN 시스템은 Commander.js가 --yes 옵션을 파싱
  AND 프롬프트를 건너뛰고 기본값으로 초기화
  AND 콘솔에 "Initializing with default settings..." 출력
  AND .moai/config.json에 기본값 저장:
    {
      "mode": "personal",
      "gitEnabled": true
    }
  AND "✓ Project initialized successfully" 출력
```

### 상세 테스트 시나리오

#### 시나리오 2.1: TTY 환경에서 --yes
**환경**:
- 일반 터미널
- TTY 있음

**실행**:
```bash
moai init --yes
```

**기대 출력**:
```
Initializing with default settings...
Creating .moai directory...
✓ Project initialized successfully
```

**검증**:
- [ ] 프롬프트 표시 안 됨
- [ ] `.moai/config.json` 생성
- [ ] 기본값 저장

---

#### 시나리오 2.2: -y 단축 플래그
**환경**:
- 일반 터미널

**실행**:
```bash
moai init -y
```

**기대 동작**:
- `--yes`와 동일

**검증**:
- [ ] 단축 플래그 인식
- [ ] 기본값 초기화

---

#### 시나리오 2.3: --yes + 비대화형 환경
**환경**:
- Claude Code (TTY 없음)

**실행**:
```bash
moai init --yes
```

**기대 동작**:
- 두 조건 모두 비대화형 모드 트리거
- 한 번만 초기화

**검증**:
- [ ] 중복 초기화 없음
- [ ] 메시지 명확

---

### 검증 방법

#### 단위 테스트
```typescript
describe('moai init with --yes flag', () => {
  it('should initialize with defaults when --yes is provided', async () => {
    const result = await runCLI(['init', '--yes']);

    expect(result.exitCode).toBe(0);
    expect(fs.existsSync('.moai/config.json')).toBe(true);

    const config = JSON.parse(fs.readFileSync('.moai/config.json', 'utf-8'));
    expect(config.mode).toBe('personal');
    expect(config.gitEnabled).toBe(true);
  });

  it('should accept -y as shorthand', async () => {
    const result = await runCLI(['init', '-y']);
    expect(result.exitCode).toBe(0);
  });
});
```

---

## AC3: Git 자동 설치 제안 (대화형)

### Given-When-Then 시나리오

```gherkin
GIVEN macOS 환경
  AND Git이 설치되지 않음 (which git 실패)
  AND TTY 환경 (대화형 가능)
WHEN 사용자가 moai init 실행
THEN 시스템은 moai doctor를 실행하여 Git 누락 감지
  AND 콘솔에 "Git is required but not installed" 출력
  AND 프롬프트 표시: "Would you like to install Git now? [Y/N/M]"

WHEN 사용자가 'Y' 입력
THEN 시스템은 which brew로 Homebrew 존재 확인
  AND brew install git 실행
  AND 설치 진행 상태 실시간 출력
  AND 설치 완료 후 "✓ Git installed successfully" 출력
  AND 초기화 계속 진행

WHEN 사용자가 'N' 입력
THEN 시스템은 초기화 중단
  AND "Initialization aborted" 출력
  AND exit code 1

WHEN 사용자가 'M' 입력
THEN 시스템은 "Manual installation: brew install git" 출력
  AND 초기화 중단
  AND exit code 1
```

### 상세 테스트 시나리오

#### 시나리오 3.1: macOS + Git 없음 + 자동 설치 'Y'
**환경**:
- macOS
- Git 없음
- Homebrew 있음

**실행**:
```bash
moai init
```

**기대 출력**:
```
Checking system dependencies...
✗ Git (runtime): not found

Git is required but not installed.
Would you like to install Git now? [Y/N/M]: Y

Installing Git via Homebrew...
==> Downloading https://...
==> Installing git...
✓ Git installed successfully

✓ Project initialized successfully
```

**검증**:
- [ ] `which git` 성공
- [ ] `.moai/config.json` 생성
- [ ] exit code 0

---

#### 시나리오 3.2: Ubuntu + Git 없음 + 수동 설치 'M'
**환경**:
- Ubuntu 22.04
- Git 없음

**실행**:
```bash
moai init
```

**사용자 입력**: `M`

**기대 출력**:
```
Git is required but not installed.
Would you like to install Git now? [Y/N/M]: M

Manual installation:
  sudo apt update
  sudo apt install git -y

Initialization aborted.
```

**검증**:
- [ ] 초기화 중단
- [ ] exit code 1
- [ ] 명령어 안내 표시

---

#### 시나리오 3.3: Git 설치 실패 (타임아웃)
**환경**:
- macOS
- Homebrew 느림 (5분 초과)

**실행**:
```bash
moai init
```

**사용자 입력**: `Y`

**기대 출력**:
```
Installing Git via Homebrew...
✗ Installation timed out after 5 minutes

Please install Git manually:
  brew install git

Initialization aborted.
```

**검증**:
- [ ] 타임아웃 감지
- [ ] 수동 가이드 제공
- [ ] exit code 1

---

### 검증 방법

#### 단위 테스트 (Mock)
```typescript
describe('DependencyInstaller.installGit', () => {
  it('should install Git via Homebrew on macOS', async () => {
    // Mock execa
    jest.mock('execa');
    execa.mockResolvedValueOnce({ stdout: '/usr/local/bin/brew' }); // which brew
    execa.mockResolvedValueOnce({ stdout: 'Installing git...' }); // brew install

    const installer = new DependencyInstaller();
    const result = await installer.installGit('darwin');

    expect(result).toBe(true);
    expect(execa).toHaveBeenCalledWith('brew', ['install', 'git']);
  });

  it('should timeout after 5 minutes', async () => {
    jest.setTimeout(10000);
    execa.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 6 * 60 * 1000)));

    const installer = new DependencyInstaller();
    await expect(installer.installGit('darwin')).rejects.toThrow('timeout');
  });
});
```

---

## AC4: Node.js 자동 설치 제안 (nvm 우선)

### Given-When-Then 시나리오

```gherkin
GIVEN 환경에서 Node.js가 설치되지 않음
  AND nvm이 설치되어 있음 (which nvm 성공)
WHEN 사용자가 moai init 실행
  AND 자동 설치에 동의
THEN 시스템은 nvm install --lts 실행
  AND Node.js 20 LTS 설치
  AND node --version으로 검증 성공
  AND 초기화 계속 진행
```

### 상세 테스트 시나리오

#### 시나리오 4.1: nvm 우선 사용
**환경**:
- macOS
- Node.js 없음
- nvm 있음

**실행**:
```bash
moai init
```

**사용자 입력**: `Y`

**기대 출력**:
```
Node.js is required but not installed.
Would you like to install Node.js now? [Y/N/M]: Y

Detected nvm, installing via nvm...
Downloading Node.js 20.x.x...
✓ Node.js 20.10.0 installed successfully

✓ Project initialized successfully
```

**검증**:
- [ ] `node --version` 성공
- [ ] v20.x.x 버전
- [ ] sudo 사용 안 함

---

#### 시나리오 4.2: nvm 없음 → Homebrew 폴백
**환경**:
- macOS
- Node.js 없음
- nvm 없음
- Homebrew 있음

**실행**:
```bash
moai init
```

**사용자 입력**: `Y`

**기대 출력**:
```
Node.js is required but not installed.
Would you like to install Node.js now? [Y/N/M]: Y

Installing Node.js via Homebrew...
==> Downloading https://...
✓ Node.js installed successfully

✓ Project initialized successfully
```

**검증**:
- [ ] `node --version` 성공
- [ ] brew install node 실행

---

#### 시나리오 4.3: 모든 도구 없음
**환경**:
- Ubuntu
- Node.js, nvm, apt 없음 (권한 부족)

**실행**:
```bash
moai init
```

**사용자 입력**: `Y`

**기대 출력**:
```
Node.js is required but not installed.
Would you like to install Node.js now? [Y/N/M]: Y

✗ Unable to install Node.js automatically

Please install Node.js manually:
  curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
  nvm install --lts

Initialization aborted.
```

**검증**:
- [ ] 수동 가이드 제공
- [ ] exit code 1

---

### 검증 방법

#### 단위 테스트
```typescript
describe('DependencyInstaller.installNodeJS', () => {
  it('should prefer nvm over system package manager', async () => {
    execa.mockResolvedValueOnce({ stdout: '/usr/local/bin/nvm' }); // which nvm
    execa.mockResolvedValueOnce({ stdout: 'v20.10.0' }); // nvm install --lts

    const installer = new DependencyInstaller();
    const result = await installer.installNodeJS('darwin');

    expect(result).toBe(true);
    expect(execa).toHaveBeenCalledWith('nvm', ['install', '--lts']);
  });

  it('should fallback to brew if nvm not found', async () => {
    execa.mockRejectedValueOnce(new Error('nvm not found')); // which nvm
    execa.mockResolvedValueOnce({ stdout: '/usr/local/bin/brew' }); // which brew
    execa.mockResolvedValueOnce({ stdout: 'Installing node...' }); // brew install

    const installer = new DependencyInstaller();
    const result = await installer.installNodeJS('darwin');

    expect(result).toBe(true);
    expect(execa).toHaveBeenCalledWith('brew', ['install', 'node']);
  });
});
```

---

## AC5: 상세 에러 메시지

### Given-When-Then 시나리오

```gherkin
GIVEN 환경에서 Git과 npm이 설치되지 않음
WHEN 사용자가 moai init 실행
THEN 시스템은 moai doctor를 실행하여 누락된 의존성 감지
  AND 콘솔에 "✗ System verification failed" 출력
  AND 누락된 의존성 목록 출력:
    Missing dependencies:
      - Git (runtime): brew install git
      - npm (development): nvm install --lts
  AND 각 의존성별 플랫폼 맞춤 설치 명령어 표시
  AND "Run 'moai doctor' for detailed diagnostics" 힌트 제공
  AND exit code 1
```

### 상세 테스트 시나리오

#### 시나리오 5.1: 여러 의존성 누락
**환경**:
- macOS
- Git, npm 없음

**실행**:
```bash
moai init
```

**기대 출력**:
```
Checking system dependencies...
✗ Git (runtime): not found
✗ npm (development): not found

✗ System verification failed

Missing dependencies:
  - Git (runtime): brew install git
  - npm (development): nvm install --lts

Run 'moai doctor' for detailed diagnostics
```

**검증**:
- [ ] 두 항목 모두 표시
- [ ] 플랫폼별 명령어
- [ ] exit code 1

---

#### 시나리오 5.2: Ubuntu 플랫폼별 명령어
**환경**:
- Ubuntu 22.04
- Git 없음

**실행**:
```bash
moai init
```

**기대 출력**:
```
Missing dependencies:
  - Git (runtime): sudo apt install git -y

Run 'moai doctor' for detailed diagnostics
```

**검증**:
- [ ] apt 명령어 표시 (brew 아님)

---

#### 시나리오 5.3: Windows 플랫폼별 명령어
**환경**:
- Windows
- Git 없음

**실행**:
```bash
moai init
```

**기대 출력**:
```
Missing dependencies:
  - Git (runtime): winget install Git.Git

Run 'moai doctor' for detailed diagnostics
```

**검증**:
- [ ] winget 명령어 표시

---

### 검증 방법

#### 단위 테스트
```typescript
describe('Error messages', () => {
  it('should show platform-specific install commands', () => {
    const errors = generateErrorMessage(['git', 'npm'], 'darwin');

    expect(errors).toContain('brew install git');
    expect(errors).toContain('nvm install --lts');
  });

  it('should suggest moai doctor', () => {
    const errors = generateErrorMessage(['git'], 'linux');
    expect(errors).toContain("Run 'moai doctor'");
  });
});
```

---

## AC6: 기존 대화형 사용자 경험 유지

### Given-When-Then 시나리오

```gherkin
GIVEN 기존 사용자가 터미널에서 작업
  AND TTY 환경이 감지됨 (process.stdin.isTTY === true)
  AND --yes 플래그 없음
WHEN 사용자가 moai init 실행
THEN 시스템은 기존 대화형 프롬프트 방식 유지
  AND inquirer.prompt() 정상 실행
  AND 사용자 경험 변화 없음
  AND 선택한 설정대로 .moai/config.json 생성
```

### 상세 테스트 시나리오

#### 시나리오 6.1: 일반 터미널 (변화 없음)
**환경**:
- macOS Terminal.app
- TTY 있음

**실행**:
```bash
moai init
```

**기대 출력**:
```
? Select project mode: (Use arrow keys)
❯ personal
  team
? Enable Git integration? (Y/n)
```

**검증**:
- [ ] 프롬프트 정상 표시
- [ ] 키보드 입력 가능
- [ ] 선택 결과 반영

---

#### 시나리오 6.2: 기존 사용자 워크플로우
**환경**:
- iTerm2
- 기존 사용자

**실행**:
```bash
moai init
# 화살표 ↓ → Team 선택
# Y → Git 활성화
```

**기대 동작**:
- 기존과 동일한 경험
- 설정 정상 저장

**검증**:
- [ ] `.moai/config.json`:
  ```json
  {
    "mode": "team",
    "gitEnabled": true
  }
  ```

---

### 검증 방법

#### E2E 테스트 (수동)
```bash
# 터미널에서 직접 실행
moai init

# 체크리스트:
# [ ] 프롬프트 표시
# [ ] 키보드 입력 가능
# [ ] 선택 결과 반영
# [ ] 이전 버전과 동일한 경험
```

---

## AC7: Docker 멀티 플랫폼 테스트

### Given-When-Then 시나리오

```gherkin
GIVEN 로컬 개발 환경에서 Docker가 설치됨
WHEN npm run test:docker 실행
THEN Docker Compose가 다음 작업 수행:
  - Linux (Ubuntu 22.04) 컨테이너 빌드
  - Windows (Server Core) 컨테이너 빌드
  AND 각 컨테이너에서 moai init 테스트 실행
  AND .docker-test/linux-results.json 생성
  AND .docker-test/windows-results.json 생성
  AND 모든 테스트 통과 시 exit 0 반환
```

### 상세 테스트 시나리오

#### 시나리오 7.1: Linux 컨테이너 (5개 시나리오)
**테스트 파일**: `.docker-test/linux-test.sh`

**시나리오**:
1. TTY 없는 환경에서 moai init 성공
2. --yes 플래그로 초기화 성공
3. Git 누락 시 자동 설치 제안 (모의)
4. 선택적 의존성 실패 시 초기화 계속
5. 에러 메시지 출력 확인

**실행**:
```bash
docker-compose -f docker-compose.test.yml up linux-test
```

**기대 출력**:
```json
{
  "platform": "linux",
  "tests": [
    {"name": "non-interactive", "status": "passed"},
    {"name": "yes-flag", "status": "passed"},
    {"name": "git-auto-install", "status": "passed"},
    {"name": "optional-deps", "status": "passed"},
    {"name": "error-messages", "status": "passed"}
  ],
  "summary": "5/5 passed"
}
```

**검증**:
- [ ] 5개 테스트 모두 passed
- [ ] JSON 파일 생성

---

#### 시나리오 7.2: Windows 컨테이너 (2개 시나리오)
**테스트 파일**: `.docker-test/windows-test.ps1`

**시나리오**:
1. 비대화형 모드 기본값 초기화
2. winget 기반 의존성 설치 가이드

**실행**:
```bash
docker-compose -f docker-compose.test.yml up windows-test
```

**기대 출력**:
```json
{
  "platform": "windows",
  "tests": [
    {"name": "non-interactive", "status": "passed"},
    {"name": "winget-guide", "status": "passed"}
  ],
  "summary": "2/2 passed"
}
```

**검증**:
- [ ] 2개 테스트 모두 passed
- [ ] JSON 파일 생성

---

### 검증 방법

#### 로컬 실행
```bash
npm run test:docker

# 결과 확인
cat .docker-test/linux-results.json
cat .docker-test/windows-results.json
```

#### CI/CD 통합
```yaml
# .github/workflows/test.yml
- name: Run Docker tests
  run: npm run test:docker

- name: Upload test results
  uses: actions/upload-artifact@v3
  with:
    name: docker-test-results
    path: .docker-test/*.json
```

---

## AC8: 로컬 검수 워크플로우

### Given-When-Then 시나리오

```gherkin
GIVEN 코드 수정 완료
WHEN npm run test:docker 실행
THEN 다음 플랫폼별 테스트 수행:
  - Linux: 5개 테스트 시나리오 통과
  - Windows: 2개 테스트 시나리오 통과
  - macOS: 로컬 npm test 수동 실행으로 확인
  AND 모든 플랫폼 테스트 통과 시 배포 준비 완료
```

### 상세 테스트 시나리오

#### 시나리오 8.1: 전체 워크플로우
**단계**:
1. 코드 수정 완료
2. 로컬 단위 테스트: `npm test`
3. Docker 멀티 플랫폼: `npm run test:docker`
4. macOS 통합 테스트: 수동 `moai init` 실행
5. TAG 체인 검증: `rg '@(SPEC|TEST|CODE):INIT-001' -n`

**체크리스트**:
- [ ] 단위 테스트 ≥85% 커버리지
- [ ] Linux 5/5 통과
- [ ] Windows 2/2 통과
- [ ] macOS 수동 테스트 성공
- [ ] TAG 체인 무결성 확인
- [ ] 배포 준비 완료

---

#### 시나리오 8.2: 실패 시 디버깅
**조건**: Linux 테스트 1개 실패

**실행**:
```bash
npm run test:docker
cat .docker-test/linux-results.json
```

**결과**:
```json
{
  "tests": [
    {"name": "non-interactive", "status": "passed"},
    {"name": "yes-flag", "status": "failed", "error": "..."}
  ]
}
```

**액션**:
1. 에러 로그 확인
2. 코드 수정
3. 재테스트

---

### 검증 방법

#### 전체 워크플로우 스크립트
```bash
#!/bin/bash
# scripts/test-all.sh

set -e

echo "1. Running unit tests..."
npm test

echo "2. Running Docker multi-platform tests..."
npm run test:docker

echo "3. Checking SPEC chain integrity..."
rg '@(SPEC|TEST|CODE):INIT-001' -n

echo "All tests passed! Ready for deployment."
```

---

## AC9: 선택적 의존성 분리

### Given-When-Then 시나리오

```gherkin
GIVEN Git LFS가 설치되지 않은 환경
WHEN 사용자가 moai init 실행
THEN moai doctor가 다음 분류로 의존성 체크:
  - runtime: git, node
  - development: npm
  - optional: git-lfs, docker
  AND Git LFS를 optional로 분류
  AND allPassed = (runtime && development) 계산
  AND allPassed = true (runtime + development만 검증)
  AND 콘솔에 "⚠ Git LFS not found (optional, safe to ignore)" 경고 출력
  AND 초기화 성공
  AND .moai/config.json 생성 확인
```

### 상세 테스트 시나리오

#### 시나리오 9.1: Git LFS 없음 → 성공
**환경**:
- macOS
- Git, Node.js, npm 있음
- Git LFS 없음

**실행**:
```bash
moai init --yes
```

**기대 출력**:
```
Checking system dependencies...
✓ Git (runtime)
✓ Node.js (runtime)
✓ npm (development)
⚠ Git LFS (optional, safe to ignore)

Running in non-interactive mode with default settings
✓ Project initialized successfully
```

**검증**:
- [ ] 초기화 성공 (중단 안 됨)
- [ ] `.moai/config.json` 생성
- [ ] exit code 0
- [ ] 경고 메시지 표시

---

#### 시나리오 9.2: Docker 없음 → 성공
**환경**:
- Ubuntu
- Git, Node.js, npm 있음
- Docker 없음

**실행**:
```bash
moai init --yes
```

**기대 출력**:
```
✓ Git (runtime)
✓ Node.js (runtime)
✓ npm (development)
⚠ Docker (optional, required for container workflows)

✓ Project initialized successfully
```

**검증**:
- [ ] 초기화 성공
- [ ] Docker 경고만 표시

---

#### 시나리오 9.3: 필수 의존성 없음 → 실패
**환경**:
- macOS
- Git 없음
- Git LFS 없음

**실행**:
```bash
moai init --yes
```

**기대 출력**:
```
✗ Git (runtime): not found
⚠ Git LFS (optional, safe to ignore)

✗ System verification failed

Missing dependencies:
  - Git (runtime): brew install git
```

**검증**:
- [ ] 초기화 실패 (Git은 필수)
- [ ] Git LFS는 실패 원인 아님
- [ ] exit code 1

---

### 검증 방법

#### 단위 테스트
```typescript
describe('Optional dependencies', () => {
  it('should pass when optional deps are missing', () => {
    const result = checkDependencies({
      git: true,
      node: true,
      npm: true,
      gitLfs: false,
      docker: false
    });

    expect(result.allPassed).toBe(true);
    expect(result.warnings).toContain('Git LFS (optional)');
  });

  it('should fail when runtime deps are missing', () => {
    const result = checkDependencies({
      git: false,
      node: true,
      npm: true,
      gitLfs: false
    });

    expect(result.allPassed).toBe(false);
    expect(result.errors).toContain('Git (runtime)');
  });
});
```

---

## 검증 체크리스트

### 기능 검증
- [ ] AC1: TTY 자동 감지 (3개 시나리오)
- [ ] AC2: --yes 플래그 지원 (3개 시나리오)
- [ ] AC3: Git 자동 설치 제안 (3개 시나리오)
- [ ] AC4: Node.js 자동 설치 제안 (3개 시나리오)
- [ ] AC5: 상세 에러 메시지 (3개 시나리오)
- [ ] AC6: 기존 대화형 사용자 경험 유지 (2개 시나리오)
- [ ] AC7: Docker 멀티 플랫폼 테스트 (2개 시나리오)
- [ ] AC8: 로컬 검수 워크플로우 (2개 시나리오)
- [ ] AC9: 선택적 의존성 분리 (3개 시나리오)

### 테스트 커버리지
- [ ] 단위 테스트 ≥85%
- [ ] 핵심 로직 ≥95% (TTY 감지, 의존성 설치)
- [ ] 통합 테스트 (Docker) 통과
- [ ] E2E 테스트 (Claude Code) 통과

### TRUST 5원칙
- [ ] **Test First**: 모든 AC에 대한 테스트 작성
- [ ] **Readable**: 함수 ≤50 LOC, 복잡도 ≤10
- [ ] **Unified**: TypeScript 타입 안전성 확보
- [ ] **Secured**: 입력 검증, 경로 traversal 방지

### 문서 동기화
- [ ] CHANGELOG.md 업데이트
- [ ] docs/cli/init.md 작성
- [ ] GitHub Issue #2에 결과 보고
- [ ] SPEC 체인 검증 완료

### 배포 준비
- [ ] 모든 플랫폼 테스트 통과 (macOS, Linux, Windows)
- [ ] CI/CD 파이프라인 통과
- [ ] Draft PR → Ready 전환 준비

---

_이 문서는 SPEC-First TDD 방법론에 따라 작성되었습니다._
_모든 시나리오는 Given-When-Then 형식으로 작성되었습니다._
