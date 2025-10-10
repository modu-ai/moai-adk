# SPEC-PLUGIN-003 수락 기준

## 📋 개요

플러그인 설치 스크립트의 성공 기준을 정의합니다. 모든 시나리오는 Given-When-Then 형식으로 작성하며, 검증 가능한 조건을 명시합니다.

---

## ✅ 인수 기준 (Acceptance Criteria)

### 1. 기본 설치 기능

#### AC-1.1: Git 클론 방식 설치 성공
**Given**: Git이 설치된 환경
**When**: `curl -sSL https://moai-adk.dev/install.sh | sh` 실행
**Then**:
- `~/.claude/plugins/moai-adk/` 디렉토리 생성됨
- `plugin.json` 파일이 존재함
- `commands/`, `agents/` 디렉토리가 존재함
- 설치 완료 메시지 출력: "✅ MoAI-ADK plugin installed successfully!"
- 다음 단계 안내 메시지 출력 (Claude Code 재시작)

**검증 방법**:
```bash
ls ~/.claude/plugins/moai-adk/plugin.json
ls -d ~/.claude/plugins/moai-adk/commands
ls -d ~/.claude/plugins/moai-adk/agents
```

#### AC-1.2: tar.gz 다운로드 방식 설치 성공
**Given**: Git이 설치되지 않은 환경
**When**: `curl -sSL https://moai-adk.dev/install.sh | sh` 실행
**Then**:
- GitHub Release API에서 최신 버전 조회됨
- tar.gz 파일 다운로드 및 압축 해제됨
- `~/.claude/plugins/moai-adk/plugin.json` 존재함
- 임시 파일 `/tmp/moai-adk.tar.gz` 삭제됨
- 설치 완료 메시지 출력

**검증 방법**:
```bash
# Git 비활성화 후 테스트
sudo mv /usr/bin/git /usr/bin/git.bak
bash install.sh
sudo mv /usr/bin/git.bak /usr/bin/git
```

---

### 2. 에러 처리

#### AC-2.1: 네트워크 오류 처리
**Given**: 네트워크 연결 불가 상태
**When**: 설치 스크립트 실행
**Then**:
- "❌ Error: Cannot reach GitHub API" 에러 메시지 출력
- 수동 설치 가이드 표시
- 종료 코드 1 반환

**검증 방법**:
```bash
# 네트워크 차단 (macOS 방화벽 또는 iptables)
./install.sh
echo $? # 1 출력 확인
```

#### AC-2.2: 권한 오류 처리
**Given**: `~/.claude/plugins/` 디렉토리 쓰기 권한 없음
**When**: 설치 스크립트 실행
**Then**:
- "❌ Error: No write permission to ~/.claude/plugins/" 에러 메시지 출력
- 권한 가이드 표시: `chmod 755 ~/.claude/plugins`
- 종료 코드 1 반환

**검증 방법**:
```bash
chmod 000 ~/.claude/plugins
./install.sh
echo $? # 1 출력 확인
chmod 755 ~/.claude/plugins # 복구
```

#### AC-2.3: 이미 설치된 경우 덮어쓰기 확인
**Given**: `~/.claude/plugins/moai-adk/` 디렉토리 이미 존재
**When**: 설치 스크립트 실행
**Then**:
- "⚠️ MoAI-ADK already installed. Overwrite? [y/N]:" 프롬프트 표시
- 사용자가 'y' 입력 시 → 기존 디렉토리 삭제 후 재설치
- 사용자가 'N' 입력 시 → "Installation cancelled." 메시지 출력 후 종료

**검증 방법**:
```bash
# 이미 설치된 상태 생성
mkdir -p ~/.claude/plugins/moai-adk
echo "test" > ~/.claude/plugins/moai-adk/test.txt

# 스크립트 실행 (대화형)
./install.sh
# 'N' 입력 → 종료 확인
# 'y' 입력 → test.txt 삭제 확인
```

#### AC-2.4: 플러그인 무결성 검증 실패
**Given**: 다운로드한 파일에 `plugin.json` 없음
**When**: 설치 검증 단계 실행
**Then**:
- "❌ Error: plugin.json not found. Installation may be corrupted." 에러 메시지 출력
- 재시도 안내 메시지 표시
- 종료 코드 1 반환

**검증 방법**:
```bash
# plugin.json 수동 삭제 후 검증 함수 실행
rm ~/.claude/plugins/moai-adk/plugin.json
verify_installation
echo $? # 1 출력 확인
```

---

### 3. 크로스 플랫폼 지원

#### AC-3.1: macOS 설치 성공
**Given**: macOS 환경 (Big Sur 이상)
**When**: `curl -sSL https://moai-adk.dev/install.sh | sh` 실행
**Then**:
- 설치 성공
- 진행률 메시지 정상 출력
- `~/.claude/plugins/moai-adk/` 생성됨

**검증 방법**:
```bash
uname -s # Darwin 확인
./install.sh
ls ~/.claude/plugins/moai-adk/
```

#### AC-3.2: Linux 설치 성공
**Given**: Ubuntu 22.04 LTS 환경
**When**: `curl -sSL https://moai-adk.dev/install.sh | sh` 실행
**Then**:
- 설치 성공
- 진행률 메시지 정상 출력
- `~/.claude/plugins/moai-adk/` 생성됨

**검증 방법**:
```bash
uname -s # Linux 확인
./install.sh
ls ~/.claude/plugins/moai-adk/
```

#### AC-3.3: Windows PowerShell 설치 성공 (선택)
**Given**: Windows 10/11 + PowerShell 7
**When**: `irm https://moai-adk.dev/install.ps1 | iex` 실행
**Then**:
- 설치 성공
- 진행률 메시지 정상 출력 (PowerShell 진행률 바)
- `%USERPROFILE%\.claude\plugins\moai-adk\` 생성됨

**검증 방법**:
```powershell
$PSVersionTable.PSVersion # 7.x 확인
.\install.ps1
Test-Path "$env:USERPROFILE\.claude\plugins\moai-adk\plugin.json"
```

---

### 4. 설치 검증

#### AC-4.1: 설치 완료 후 플러그인 구조 확인
**Given**: 설치 완료 상태
**When**: 설치 검증 함수 실행
**Then**:
- `plugin.json` 존재 확인 통과
- `commands/`, `agents/` 디렉토리 존재 확인 통과
- "✅ MoAI-ADK plugin installed successfully!" 메시지 출력

**검증 방법**:
```bash
verify_installation
echo $? # 0 출력 확인
```

#### AC-4.2: Claude Code 재시작 후 플러그인 인식
**Given**: 설치 완료 후 Claude Code 재시작
**When**: `/alfred:8-project` 커맨드 입력
**Then**:
- 커맨드가 인식됨 (오류 없음)
- Alfred SuperAgent 활성화 확인

**검증 방법**:
```bash
# Claude Code 재시작 후
/alfred:8-project
# "Alfred SuperAgent initialized" 메시지 확인
```

---

### 5. 사용자 안내 메시지

#### AC-5.1: 설치 진행 메시지 출력
**Given**: 설치 스크립트 실행 중
**When**: Git 클론 방식 선택됨
**Then**:
- "Git detected. Using git clone method..." 메시지 출력
- "Cloning MoAI-ADK plugin..." 메시지 출력

**검증 방법**:
```bash
./install.sh 2>&1 | grep "Git detected"
./install.sh 2>&1 | grep "Cloning"
```

#### AC-5.2: 설치 완료 후 다음 단계 안내
**Given**: 설치 성공 상태
**When**: 설치 완료 메시지 출력
**Then**:
- "Next steps:" 섹션 표시
- "1. Restart Claude Code" 안내
- "2. Verify plugin: ls ~/.claude/plugins/moai-adk" 안내
- "3. Quick start: /alfred:8-project" 안내

**검증 방법**:
```bash
./install.sh 2>&1 | grep "Next steps"
./install.sh 2>&1 | grep "Restart Claude Code"
./install.sh 2>&1 | grep "/alfred:8-project"
```

---

## 🧪 테스트 시나리오 (Given-When-Then)

### 시나리오 1: 정상 설치 (Git 환경)
**Given**:
- Git 설치됨 (`which git` 성공)
- `~/.claude/plugins/` 디렉토리 존재
- 네트워크 연결 가능

**When**:
```bash
curl -sSL https://moai-adk.dev/install.sh | sh
```

**Then**:
1. "Git detected. Using git clone method..." 출력
2. `git clone https://github.com/modu-ai/moai-adk ~/.claude/plugins/moai-adk` 실행
3. `plugin.json` 존재 확인 통과
4. "✅ MoAI-ADK plugin installed successfully!" 출력
5. 종료 코드 0

---

### 시나리오 2: 정상 설치 (Git 미설치)
**Given**:
- Git 미설치 (`which git` 실패)
- curl/wget 설치됨
- `~/.claude/plugins/` 디렉토리 존재
- 네트워크 연결 가능

**When**:
```bash
curl -sSL https://moai-adk.dev/install.sh | sh
```

**Then**:
1. "Git not found. Using tar.gz download method..." 출력
2. GitHub Release API 호출 (`/repos/modu-ai/moai-adk/releases/latest`)
3. tar.gz 다운로드 및 압축 해제
4. 임시 파일 삭제 (`/tmp/moai-adk.tar.gz`)
5. `plugin.json` 존재 확인 통과
6. "✅ MoAI-ADK plugin installed successfully!" 출력
7. 종료 코드 0

---

### 시나리오 3: 네트워크 오류
**Given**:
- 네트워크 연결 불가 (GitHub API 접근 실패)

**When**:
```bash
./install.sh
```

**Then**:
1. "❌ Error: Cannot reach GitHub API" 출력
2. "→ Check your internet connection" 출력
3. "→ Manual installation: git clone ..." 수동 가이드 출력
4. 종료 코드 1

---

### 시나리오 4: 이미 설치된 경우 (덮어쓰기 거부)
**Given**:
- `~/.claude/plugins/moai-adk/` 이미 존재

**When**:
```bash
./install.sh
# 프롬프트에서 'N' 입력
```

**Then**:
1. "⚠️ MoAI-ADK already installed. Overwrite? [y/N]:" 프롬프트 표시
2. 사용자 'N' 입력
3. "Installation cancelled." 출력
4. 기존 디렉토리 유지 (삭제 안 됨)
5. 종료 코드 0

---

## 🎯 품질 게이트 기준 (Definition of Done)

### 필수 조건

- [ ] `scripts/install.sh` 파일 생성 및 실행 권한 부여 (`chmod +x`)
- [ ] Git 클론 방식 정상 작동 (Git 설치 환경)
- [ ] tar.gz 다운로드 방식 정상 작동 (Git 미설치 환경)
- [ ] 모든 에러 시나리오 처리 (네트워크, 권한, 무결성)
- [ ] 설치 검증 통과 (`plugin.json` 존재 확인)
- [ ] 설치 완료 메시지 및 다음 단계 안내 출력
- [ ] macOS, Linux 환경에서 테스트 성공

### 선택 조건

- [ ] Windows PowerShell 스크립트 작성 (`scripts/install.ps1`)
- [ ] PowerShell 환경에서 테스트 성공
- [ ] 자동화 테스트 스크립트 작성 (`tests/scripts/install.test.sh`)
- [ ] 설치 가이드 문서 작성 (`docs/installation.md`)

### TRUST 5원칙 준수

- **T**est First:
  - [ ] `tests/scripts/install.test.sh` 작성
  - [ ] 모든 에러 시나리오 테스트 케이스 포함

- **R**eadable:
  - [ ] 스크립트 주석 작성 (각 함수 설명)
  - [ ] 함수 이름 명확성 (예: `verify_installation()`, `detect_git()`)

- **U**nified:
  - [ ] 일관된 에러 메시지 형식 (❌, ⚠️, ✅)
  - [ ] 일관된 함수 네이밍 컨벤션

- **S**ecured:
  - [ ] 권한 확인 로직 포함
  - [ ] 임시 파일 안전하게 삭제
  - [ ] GitHub Release API 응답 검증

- **T**rackable:
  - [ ] `@CODE:PLUGIN-003` TAG 주석 추가
  - [ ] `@SPEC:PLUGIN-003` 참조 명시

---

## 📚 검증 방법 및 도구

### 수동 테스트
```bash
# 1. Git 환경 테스트
./install.sh

# 2. Git 미설치 환경 테스트 (Git 임시 비활성화)
sudo mv /usr/bin/git /usr/bin/git.bak
./install.sh
sudo mv /usr/bin/git.bak /usr/bin/git

# 3. 덮어쓰기 테스트
./install.sh # 'N' 입력
./install.sh # 'y' 입력

# 4. 권한 오류 테스트
chmod 000 ~/.claude/plugins
./install.sh
chmod 755 ~/.claude/plugins
```

### 자동화 테스트 (선택)
```bash
# tests/scripts/install.test.sh
bash tests/scripts/install.test.sh
```

### 설치 검증
```bash
# 플러그인 구조 확인
ls -la ~/.claude/plugins/moai-adk/
cat ~/.claude/plugins/moai-adk/plugin.json

# Claude Code에서 플러그인 인식 확인
# Claude Code 재시작 후
/alfred:8-project
```

---

## 다음 단계

1. **TDD 구현**: `/alfred:2-build SPEC-PLUGIN-003`
2. **테스트 실행**: `bash tests/scripts/install.test.sh`
3. **문서 동기화**: `/alfred:3-sync`
4. **배포 URL 설정**: `https://moai-adk.dev/install.sh` 호스팅

**완료 조건 달성 시 SPEC-PLUGIN-003 종료** ✅
