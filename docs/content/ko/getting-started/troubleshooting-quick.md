# 빠른 문제 해결

**@DOC:TROUBLESHOOTING-QUICK-001** | **최종 업데이트**: 2025-11-10 | **대상**: 초보자

MoAI-ADK 설치 및 사용 중 자주 발생하는 5가지 문제와 해결책입니다. 각 문제에 대해 단계적인 해결 방법을 제시하므로 차례대로 따라하세요.

______________________________________________________________________

## 1. Alfred 명령어를 인식하지 못함

<span class="material-icons">error</span> **명령어 인식 문제**

### 증상

다음과 같은 오류 메시지가 표시됩니다:

```
Command not found: /alfred:1-plan
zsh: command not found: /alfred:1-plan
```

### 원인

Claude Code가 `.claude/commands/` 디렉토리를 아직 인식하지 못했거나, 프로젝트 루트 디렉토리가 아닌 다른 위치에서 명령을 실행했습니다.

### 해결책

#### 방법 1: Claude Code 재시작 (권장)

1. 현재 Claude Code 세션을 종료합니다
2. 프로젝트 루트 디렉토리에서 Claude Code를 다시 실행합니다
3. 다음 명령으로 프로젝트가 인식되었는지 확인합니다:

```bash
ls -la .claude/commands/
```

**예상 결과**: 명령 파일 목록이 표시되어야 합니다
```
0-project.md
1-plan.md
2-run.md
3-sync.md
...
```

4. Alfred 명령 실행 테스트:

```bash
/alfred:0-project setting
```

**성공 시**: Alfred가 프로젝트 설정 정보를 표시합니다

#### 방법 2: 프로젝트 디렉토리 확인

1. 현재 작업 디렉토리 확인:

```bash
pwd
```

2. 프로젝트 루트로 이동:

```bash
cd /path/to/your/project
```

3. CLAUDE.md 파일이 있는지 확인:

```bash
ls CLAUDE.md
```

**없으면**: 잘못된 디렉토리에 있는 것입니다. 프로젝트 루트로 이동하세요.

#### 방법 3: 프로젝트 재초기화

1. `.moai/` 디렉토리 백업:

```bash
mv .moai .moai.backup
```

2. 프로젝트 재초기화:

```bash
moai-adk init --reinit
```

3. 설정 복원 (필요시):

```bash
cp .moai.backup/config.json .moai/config.json
```

### 확인 방법

다음 명령이 작동하면 문제가 해결된 것입니다:

```bash
/alfred:1-plan test
```

**성공 메시지 예시**:
```
Alfred: SPEC 생성을 시작합니다...
Feature: test
SPEC-001 생성 완료
```

______________________________________________________________________

## 2. UV 패키지 매니저 설치 오류

<span class="material-icons">package</span> **패키지 매니저 설정**

### 증상

UV 명령이 인식되지 않습니다:

```bash
uv --version
# 출력: uv: command not found
```

### 원인

UV가 설치되지 않았거나, PATH 환경변수에 등록되지 않았습니다.

### 해결책

#### 방법 1: 공식 스크립트로 설치 (권장)

**macOS / Linux**:

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Windows (PowerShell)**:

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

설치 후 터미널을 재시작합니다.

#### 방법 2: pip로 설치

```bash
pip install uv
```

#### 방법 3: PATH 환경변수 추가

UV가 설치되어 있지만 인식되지 않는 경우:

**macOS / Linux**:

1. UV 설치 경로 확인:

```bash
find ~ -name uv -type f 2>/dev/null
```

2. `~/.zshrc` 또는 `~/.bashrc`에 추가:

```bash
export PATH="$HOME/.cargo/bin:$PATH"
```

3. 설정 적용:

```bash
source ~/.zshrc
# 또는
source ~/.bashrc
```

**Windows**:

1. 시스템 환경변수 설정 열기
2. Path 변수에 UV 설치 경로 추가: `C:\Users\YourName\.cargo\bin`
3. 터미널 재시작

### 확인 방법

```bash
uv --version
```

**예상 결과**:
```
uv 0.4.30 (또는 최신 버전)
```

### 추가 문제 해결

UV가 설치되었지만 프로젝트 의존성 설치가 실패하는 경우:

```bash
# 캐시 정리
uv cache clean

# 의존성 재설치
uv sync --force

# 특정 Python 버전 지정
uv sync --python 3.13
```

______________________________________________________________________

## 3. Python 버전 호환성 문제

<span class="material-icons">code</span> **Python 버전 관리**

### 증상

프로젝트 실행 시 Python 버전 오류가 발생합니다:

```
Error: Python 3.13 or higher is required
Current version: Python 3.11.5
```

### 원인

MoAI-ADK는 Python 3.13 이상을 요구하지만, 시스템에 이전 버전이 설치되어 있습니다.

### 해결책

#### 방법 1: pyenv로 Python 버전 관리 (권장)

**macOS / Linux**:

1. pyenv 설치:

```bash
# macOS
brew install pyenv

# Linux
curl https://pyenv.run | bash
```

2. Python 3.13 설치:

```bash
pyenv install 3.13.0
```

3. 프로젝트 디렉토리에서 버전 지정:

```bash
cd /path/to/your/project
pyenv local 3.13.0
```

4. 확인:

```bash
python --version
# 출력: Python 3.13.0
```

#### 방법 2: 공식 Python 설치

**macOS**:

```bash
brew install python@3.13
```

**Windows**:

1. [Python 공식 사이트](https://www.python.org/downloads/)에서 Python 3.13 다운로드
2. 설치 시 "Add Python to PATH" 체크
3. 설치 완료 후 터미널 재시작

**Linux (Ubuntu/Debian)**:

```bash
sudo add-apt-repository ppa:deadsnakes/ppa
sudo apt update
sudo apt install python3.13 python3.13-venv python3.13-dev
```

### 확인 방법

```bash
python --version
```

**예상 결과**:
```
Python 3.13.0 (또는 3.13 이상)
```

### 가상환경 재생성

Python 버전을 변경한 후 가상환경을 재생성해야 합니다:

```bash
# 기존 가상환경 제거
rm -rf .venv

# UV로 새로운 가상환경 생성
uv venv --python 3.13

# 의존성 재설치
uv sync
```

______________________________________________________________________

## 4. .moai 디렉토리 권한 오류

<span class="material-icons">lock</span> **파일 권한 설정**

### 증상

파일 생성 또는 수정 시 권한 오류가 발생합니다:

```
PermissionError: [Errno 13] Permission denied: '.moai/config.json'
```

### 원인

`.moai/` 디렉토리 또는 파일의 권한이 올바르지 않습니다. 일반적으로 다른 사용자가 생성한 파일이거나, sudo로 실행했을 때 발생합니다.

### 해결책

#### 방법 1: 디렉토리 소유권 변경 (권장)

**macOS / Linux**:

```bash
# 현재 사용자로 소유권 변경
sudo chown -R $USER:$USER .moai

# 권한 확인
ls -la .moai/
```

**예상 결과**:
```
drwxr-xr-x  5 yourname  staff  160 Nov 10 10:00 .moai
-rw-r--r--  1 yourname  staff 1234 Nov 10 10:00 config.json
```

#### 방법 2: 권한 수정

```bash
# 디렉토리 권한 수정
chmod -R 755 .moai

# 파일 권한 수정
chmod 644 .moai/config.json
```

#### 방법 3: 디렉토리 재생성

심각한 권한 문제인 경우:

```bash
# 기존 디렉토리 백업
mv .moai .moai.old

# 새 디렉토리 생성
mkdir -p .moai/reports

# 설정 파일 복사
cp .moai.old/config.json .moai/

# 권한 확인
ls -la .moai/
```

### 확인 방법

```bash
# 파일 생성 테스트
touch .moai/test.txt

# 성공 시 파일 제거
rm .moai/test.txt
```

### 예방 조치

1. **sudo 사용 금지**: MoAI-ADK 명령은 일반 사용자 권한으로 실행하세요
2. **가상환경 사용**: 항상 가상환경 내에서 작업하세요
3. **정기 권한 점검**: 주기적으로 `.moai/` 권한을 확인하세요

```bash
# 권한 점검 스크립트
ls -la .moai/ | grep -v "^d.*$USER"
```

______________________________________________________________________

## 5. SPEC 문서 생성 실패

<span class="material-icons">description</span> **문서 생성 문제**

### 증상

SPEC 생성 명령이 실패하거나 불완전한 문서가 생성됩니다:

```bash
/alfred:1-plan "사용자 인증"
# 출력: Error: Failed to generate SPEC document
```

### 원인

1. 프로젝트 초기화가 완료되지 않았습니다
2. `.claude/agents/` 디렉토리가 손상되었습니다
3. 네트워크 연결 문제로 AI 응답을 받지 못했습니다
4. CLAUDE.md 설정이 잘못되었습니다

### 해결책

#### 방법 1: 프로젝트 상태 확인

```bash
# 프로젝트 초기화 확인
/alfred:0-project info
```

**예상 결과**: 프로젝트 정보가 표시되어야 합니다

프로젝트가 초기화되지 않았다면:

```bash
/alfred:0-project setting
```

#### 방법 2: 에이전트 파일 확인

```bash
# spec-builder 에이전트 존재 확인
ls -la .claude/agents/alfred/spec-builder.md
```

파일이 없거나 손상된 경우:

```bash
# MoAI-ADK 재설치
pip install --upgrade moai-adk

# 템플릿 동기화
moai-adk sync-templates
```

#### 방법 3: 네트워크 및 API 연결 확인

```bash
# 네트워크 연결 테스트
ping -c 3 api.anthropic.com
```

네트워크가 정상이면:

```bash
# Claude Code 재시작
# 1. 현재 세션 종료
# 2. 네트워크 연결 확인
# 3. Claude Code 재실행
```

#### 방법 4: 간단한 SPEC 생성 테스트

복잡한 요청 대신 간단한 테스트:

```bash
/alfred:1-plan "hello world"
```

성공하면 원래 요청으로 재시도:

```bash
/alfred:1-plan "사용자 이메일 및 비밀번호 인증 기능"
```

#### 방법 5: 수동 SPEC 템플릿 사용

자동 생성이 계속 실패하는 경우:

1. 템플릿 복사:

```bash
cp .claude/skills/moai-foundation-specs/templates/spec-template.md \
   .moai/specs/SPEC-001-user-auth.md
```

2. 수동으로 내용 작성:

```markdown
# SPEC-001: 사용자 인증

@SPEC:AUTH-001

## 개요
사용자 이메일과 비밀번호를 이용한 인증 기능

## 요구사항

WHEN 사용자가 로그인 폼을 제출하면
  AND 올바른 이메일과 비밀번호를 입력했을 때
THEN 시스템은 사용자를 인증하고
  AND 대시보드로 리디렉션한다
```

### 확인 방법

```bash
# 생성된 SPEC 확인
ls -la .moai/specs/

# 내용 확인
cat .moai/specs/SPEC-001-*.md
```

### 디버깅 팁

문제가 지속되면 로그를 확인하세요:

```bash
# Alfred 로그 확인
cat .moai/logs/alfred.log | tail -n 50

# 오류 패턴 검색
grep -i "error\|fail" .moai/logs/alfred.log
```

______________________________________________________________________

## 추가 도움말

<span class="material-icons">help</span> **추가 지원 정보**

### 로그 확인 방법

모든 작업은 `.moai/logs/` 디렉토리에 기록됩니다:

```bash
# 최근 로그 확인
tail -f .moai/logs/alfred.log

# 오류만 필터링
grep -i error .moai/logs/alfred.log

# 특정 날짜 로그
ls -lt .moai/logs/ | head
```

### 전체 시스템 재설정

모든 방법이 실패한 경우 마지막 수단:

```bash
# 1. 백업
cp -r .moai .moai.backup
cp CLAUDE.md CLAUDE.md.backup

# 2. MoAI-ADK 재설치
pip uninstall -y moai-adk
pip install moai-adk

# 3. 프로젝트 재초기화
moai-adk init --reinit

# 4. 설정 복원
cp .moai.backup/config.json .moai/config.json
```

### 커뮤니티 지원

문제가 해결되지 않으면:

1. **GitHub Issues**: [기술 문제 제기](https://github.com/moai-adk/MoAI-ADK/issues)
2. **GitHub Discussions**: [질의응답](https://github.com/moai-adk/MoAI-ADK/discussions)
3. **이메일 지원**: support@mo.ai.kr

______________________________________________________________________

## 관련 문서

<span class="material-icons">link</span> **참조 문서**

- **설치 가이드**: [installation.md](installation.md)
- **빠른 시작**: [quick-start-ko.md](quick-start-ko.md)
- **용어집**: [glossary.md](glossary.md)
- **전체 문제해결**: [../troubleshooting/index.md](troubleshooting/index.md)

______________________________________________________________________

*최종 업데이트: 2025-11-10 | 버전: v0.21.2 | 상태: Phase 1 완료*
