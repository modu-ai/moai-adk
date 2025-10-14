# Installation Guide

> **MoAI-ADK Python v0.3.0 설치 가이드**
>
> Python 3.13+ 기반 SPEC-First TDD 프레임워크

---

## 📋 Table of Contents

- [시스템 요구사항](#시스템-요구사항)
- [Python 환경 설정](#python-환경-설정)
- [MoAI-ADK 설치](#moai-adk-설치)
- [Claude Code 설정](#claude-code-설정)
- [설치 검증](#설치-검증)
- [문제 해결](#문제-해결)
- [다음 단계](#다음-단계)

---

## 시스템 요구사항

### 🔴 필수 요구사항

MoAI-ADK를 사용하려면 다음 소프트웨어가 **반드시** 설치되어 있어야 합니다:

| 소프트웨어     | 최소 버전 | 권장 버전 | 설명                           |
| -------------- | --------- | --------- | ------------------------------ |
| **Python**     | 3.13.0    | 3.13+     | MoAI-ADK 런타임                |
| **pip**        | 24.0      | 최신      | Python 패키지 관리자           |
| **Git**        | 2.30.0    | 2.40+     | 버전 관리 (선택사항)           |
| **Claude Code** | 1.2.0     | 최신      | AI 에이전트 통합 IDE (필수)    |

### 🟡 권장 요구사항

더 나은 개발 경험을 위해 다음을 권장합니다:

| 도구          | 버전  | 용도                        |
| ------------- | ----- | --------------------------- |
| **uv**        | 최신  | 빠른 Python 패키지 관리자   |
| **pyenv**     | 최신  | Python 버전 관리            |
| **gh CLI**    | 2.0+  | GitHub PR/Issue 자동화      |

### 🌍 지원 운영체제

| OS           | 버전            | 아키텍처         | 상태      |
| ------------ | --------------- | ---------------- | --------- |
| **macOS**    | 12 Monterey+    | Intel, Apple M1/M2/M3 | ✅ Stable |
| **Linux**    | Ubuntu 20.04+   | x86_64, ARM64    | ✅ Stable |
|              | CentOS 8+       | x86_64           | ✅ Stable |
|              | Debian 11+      | x86_64, ARM64    | ✅ Stable |
|              | Arch Linux      | x86_64           | ✅ Stable |
| **Windows**  | 10/11           | x86_64           | ✅ Stable |
|              | WSL2 (Ubuntu)   | x86_64           | ✅ Recommended |

### 💾 하드웨어 요구사항

| 항목       | 최소   | 권장   |
| ---------- | ------ | ------ |
| **RAM**    | 4 GB   | 8 GB+  |
| **디스크** | 500 MB | 2 GB+  |
| **CPU**    | 2 코어 | 4 코어 |

---

## Python 환경 설정

### Option A: uv 사용 (권장 ⚡ 빠름)

**uv**는 Rust로 작성된 초고속 Python 패키지 관리자입니다. pip보다 **10-100배 빠릅니다**.

#### 1. uv 설치

**macOS/Linux:**
```bash
# 공식 설치 스크립트 (권장)
curl -LsSf https://astral.sh/uv/install.sh | sh

# Homebrew
brew install uv
```

**Windows (PowerShell):**
```powershell
# 공식 설치 스크립트
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# Scoop
scoop install uv
```

#### 2. 설치 확인

```bash
# uv 버전 확인
uv --version
# 출력 예시: uv 0.5.1

# Python 버전 확인 (uv가 자동으로 Python 3.13 다운로드)
uv python list
```

#### 3. Python 3.13 설치 (uv가 자동 관리)

```bash
# Python 3.13 설치 (uv가 자동으로 다운로드)
uv python install 3.13

# 설치 확인
uv python list
# 출력 예시:
# cpython-3.13.0-macos-aarch64-none ✓
# cpython-3.13.1-linux-x86_64-gnu
```

**특징**:
- ✅ 수동 Python 설치 불필요
- ✅ 여러 Python 버전 동시 관리
- ✅ 가상환경 자동 생성/관리
- ✅ 패키지 설치 속도 10-100배 빠름

---

### Option B: 시스템 Python 사용

시스템에 Python 3.13이 이미 설치되어 있다면, 기존 설치를 그대로 사용할 수 있습니다.

#### 1. Python 설치

**macOS:**
```bash
# Homebrew 사용 (권장)
brew install python@3.13

# 설치 확인
python3.13 --version
# 출력 예시: Python 3.13.0
```

**Ubuntu/Debian:**
```bash
# APT 사용
sudo apt update
sudo apt install python3.13 python3.13-venv python3-pip

# 설치 확인
python3.13 --version
```

**Windows:**
1. [Python 공식 웹사이트](https://www.python.org/downloads/)에서 Python 3.13 다운로드
2. 설치 시 "Add Python to PATH" 체크 ✅
3. 터미널에서 확인:
   ```powershell
   python --version
   # 출력 예시: Python 3.13.0
   ```

#### 2. pip 업그레이드

```bash
# pip 최신 버전으로 업그레이드
python3.13 -m pip install --upgrade pip

# 버전 확인
pip --version
# 출력 예시: pip 24.0 from ... (python 3.13)
```

---

### Option C: pyenv로 Python 버전 관리

여러 Python 버전을 사용하는 개발자에게 권장합니다.

#### 1. pyenv 설치

**macOS/Linux:**
```bash
# pyenv 설치
curl https://pyenv.run | bash

# 셸 설정 추가 (~/.bashrc 또는 ~/.zshrc)
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
echo 'export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init --path)"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc

# 셸 재시작
source ~/.bashrc
```

**Windows:**
```powershell
# pyenv-win 설치
Invoke-WebRequest -UseBasicParsing -Uri "https://raw.githubusercontent.com/pyenv-win/pyenv-win/master/pyenv-win/install-pyenv-win.ps1" -OutFile "./install-pyenv-win.ps1"; &"./install-pyenv-win.ps1"
```

#### 2. Python 3.13 설치

```bash
# 사용 가능한 버전 목록 확인
pyenv install --list | grep 3.13

# Python 3.13 최신 버전 설치
pyenv install 3.13.0

# 글로벌 버전 설정
pyenv global 3.13.0

# 확인
python --version
# 출력 예시: Python 3.13.0
```

---

## MoAI-ADK 설치

### Option A: uv로 설치 (권장 ⚡)

uv를 사용하면 **가장 빠르고 안전하게** MoAI-ADK를 설치할 수 있습니다.

```bash
# 전역 설치 (권장)
uv tool install moai-adk

# 설치 확인
moai --version
# 출력 예시: moai-adk v0.3.0
```

**특징**:
- ✅ 의존성 충돌 자동 해결
- ✅ 격리된 환경 (다른 패키지와 충돌 없음)
- ✅ 빠른 설치 (pip 대비 10배 빠름)
- ✅ 자동 업데이트 지원

---

### Option B: pip로 설치 (표준)

시스템 Python을 사용하는 경우 pip로 설치할 수 있습니다.

```bash
# 전역 설치
pip install moai-adk

# 사용자 전용 설치 (권한 문제 시)
pip install --user moai-adk

# 설치 확인
moai --version
# 출력 예시: moai-adk v0.3.0
```

**주의**:
- ⚠️ 전역 설치 시 다른 Python 패키지와 충돌 가능
- ⚠️ 권한 에러 시 `sudo` 또는 `--user` 옵션 필요

---

### Option C: 개발자 모드 설치

MoAI-ADK 개발에 기여하거나, 소스 코드를 수정하려는 경우:

```bash
# 1. 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 2. 개발 의존성 설치 (uv 사용)
uv sync --all-extras

# 또는 pip 사용
pip install -e ".[dev]"

# 3. 설치 확인
moai --version
# 출력 예시: moai-adk v0.3.0 (dev)
```

**개발 환경 명령어**:
```bash
# 테스트 실행
pytest

# 커버리지 리포트
pytest --cov=moai_adk --cov-report=html

# 타입 체크
mypy src/moai_adk

# 코드 품질 검사
ruff check src/

# 코드 포맷팅
ruff format src/
```

---

## Claude Code 설정

### 1. Claude Code 설치

MoAI-ADK는 **Claude Code** 환경에서 **필수적으로** 사용해야 합니다. Alfred SuperAgent와 9개의 전문 AI 에이전트가 Claude Code에 통합되어 있습니다.

**Claude Code 설치**:
1. [Claude Code 다운로드 페이지](https://claude.ai/code) 방문
2. 운영체제에 맞는 설치 파일 다운로드
3. 설치 후 실행

**버전 확인**:
```bash
# Claude Code 버전 확인
claude --version
# 출력 예시: Claude Code v1.2.5
```

**최소 버전 요구사항**: Claude Code v1.2.0 이상

---

### 2. 프로젝트에 MoAI-ADK 설치

터미널에서 프로젝트 디렉토리로 이동 후:

```bash
# 새 프로젝트 생성
moai init my-project
cd my-project

# 기존 프로젝트에 설치
cd existing-project
moai init .
```

**`moai init .` 실행 결과**:
```
✅ MoAI-ADK 프로젝트 초기화 완료
📁 생성된 파일 및 디렉토리:
  - .moai/config.json         (프로젝트 설정)
  - .moai/memory/             (개발 가이드)
  - .moai/specs/              (SPEC 문서)
  - .moai/reports/            (동기화 리포트)
  - .claude/custom-commands/  (Alfred 커맨드)
  - .claude/agents/           (10개 AI 에이전트)

🎯 다음 단계:
  1. Claude Code 실행: claude
  2. 프로젝트 초기화: /alfred:0-project
```

---

### 3. Claude Code에서 Alfred 활성화

Claude Code를 실행하고, 다음 명령어로 Alfred를 활성화합니다:

```bash
# 터미널에서 Claude Code 실행
claude
```

**Claude Code 내에서**:
```text
/alfred:0-project
```

**Alfred가 자동으로 수행**:
1. 프로젝트 구조 분석 (파일, 디렉토리)
2. 언어 및 프레임워크 자동 감지
3. `.moai/project/` 디렉토리 생성
4. 3개 핵심 문서 작성:
   - `product.md` (제품 개요, 목표)
   - `structure.md` (디렉토리 구조, 모듈 설계)
   - `tech.md` (기술 스택, 도구 체인)

**실행 결과 예시**:
```
📖 Alfred가 프로젝트를 분석하고 있습니다...

✅ 프로젝트 분석 완료:
  - 언어: Python 3.13
  - 프레임워크: FastAPI 0.104.0
  - 테스트: pytest
  - 린터: ruff

📝 생성된 문서:
  - .moai/project/product.md (200 lines)
  - .moai/project/structure.md (150 lines)
  - .moai/project/tech.md (180 lines)

🎉 프로젝트 초기화 완료!
다음 단계: /alfred:1-spec "첫 기능 설명"
```

---

## 설치 검증

### 1. CLI 명령어 확인

```bash
# 버전 확인
moai --version
# 출력: moai-adk v0.3.0

# 도움말 확인
moai --help
# 출력: MoAI-ADK CLI 사용법

# 시스템 진단
moai doctor
```

---

### 2. `moai doctor` 상세 출력

`moai doctor` 명령어는 시스템 환경을 진단하고, 필수 요구사항이 충족되었는지 확인합니다.

**실행 예시**:
```bash
moai doctor
```

**정상 출력 예시**:
```
🔍 MoAI-ADK 시스템 진단 시작...

✅ Python 환경
  - Python: 3.13.0 ✓
  - pip: 24.0 ✓
  - 위치: /opt/homebrew/bin/python3.13

✅ Git 설정
  - Git: 2.42.0 ✓
  - 사용자: Goos (goos@example.com) ✓
  - 브랜치: main

✅ 필수 의존성
  - click: 8.1.7 ✓
  - rich: 13.7.0 ✓
  - gitpython: 3.1.40 ✓
  - jinja2: 3.1.2 ✓
  - pyyaml: 6.0.1 ✓

✅ Claude Code
  - 버전: v1.2.5 ✓
  - 경로: /usr/local/bin/claude

✅ 프로젝트 구조
  - .moai/config.json ✓
  - .moai/memory/ ✓
  - .moai/specs/ ✓
  - .claude/agents/ ✓

🎉 모든 시스템 요구사항이 충족되었습니다!
```

**경고/에러 출력 예시**:
```
⚠️ 경고 발견:

❌ Python 버전 부족
  - 현재: 3.12.5
  - 필요: 3.13.0+
  → 해결: Python 3.13 설치 필요

❌ Git 설치 안 됨
  - Git이 설치되지 않았습니다
  → 해결: brew install git (macOS)
         apt install git (Ubuntu)

⚠️ Claude Code 버전 낮음
  - 현재: v1.1.8
  - 권장: v1.2.0+
  → 해결: Claude Code 업데이트 권장

💡 문제 해결 가이드:
  - https://moai-adk.vercel.app/getting-started/installation#troubleshooting
```

---

### 3. Python 버전 확인

```bash
# Python 버전
python --version
# 출력: Python 3.13.0

# moai가 사용하는 Python 경로
which python
# 출력: /opt/homebrew/bin/python3.13
```

---

### 4. Claude Code 통합 확인

Claude Code를 실행하고, Alfred 커맨드가 인식되는지 확인:

```bash
# Claude Code 실행
claude
```

**Claude Code 내에서**:
```text
# 명령어 목록 확인 (/ 입력 시 자동완성)
/alfred:0-project   ✓
/alfred:1-spec      ✓
/alfred:2-build     ✓
/alfred:3-sync      ✓
```

---

## 문제 해결

### 문제 1: `moai: command not found`

**증상**:
```bash
moai --version
# zsh: command not found: moai
```

**원인**: PATH 환경 변수에 moai 실행 파일 경로가 없음

**해결 방법**:

#### A. uv로 설치한 경우

```bash
# uv 도구 경로 확인
uv tool list
# 출력: moai-adk v0.3.0 (/Users/goos/.local/bin/moai)

# PATH에 추가 (~/.bashrc 또는 ~/.zshrc)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 재확인
moai --version
```

#### B. pip로 설치한 경우

```bash
# Python 스크립트 경로 확인
python -m site --user-base
# 출력: /Users/goos/.local

# PATH에 추가
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 재확인
moai --version
```

---

### 문제 2: Python 버전 부족 (3.13 미만)

**증상**:
```bash
moai doctor
# ❌ Python 버전: 3.12.5 (3.13.0+ 필요)
```

**원인**: 시스템 Python이 3.13보다 낮음

**해결 방법**:

#### A. uv 사용 (권장)

```bash
# uv로 Python 3.13 설치
uv python install 3.13

# 설치 확인
uv python list
# cpython-3.13.0-macos-aarch64-none ✓
```

#### B. pyenv 사용

```bash
# Python 3.13 설치
pyenv install 3.13.0
pyenv global 3.13.0

# 확인
python --version
# Python 3.13.0
```

#### C. 시스템 Python 업그레이드

**macOS**:
```bash
brew upgrade python@3.13
```

**Ubuntu**:
```bash
sudo apt update
sudo apt install python3.13
```

---

### 문제 3: 권한 에러 (Permission Denied)

**증상**:
```bash
pip install moai-adk
# ERROR: Could not install packages due to an EnvironmentError: [Errno 13] Permission denied
```

**원인**: 시스템 디렉토리에 쓰기 권한 없음

**해결 방법**:

#### A. 사용자 디렉토리 설치 (권장)

```bash
# --user 옵션 사용
pip install --user moai-adk
```

#### B. uv 사용 (권한 문제 없음)

```bash
# uv는 사용자 디렉토리에 자동 설치
uv tool install moai-adk
```

#### C. 가상환경 사용

```bash
# venv 생성
python -m venv .venv

# 활성화
source .venv/bin/activate  # macOS/Linux
.venv\Scripts\activate     # Windows

# 설치
pip install moai-adk
```

---

### 문제 4: Claude Code가 에이전트를 인식하지 못함

**증상**:
```text
# Claude Code에서
/alfred:0-project
# Error: Unknown command
```

**원인**: `.claude/` 디렉토리가 생성되지 않았거나, 에이전트 파일이 누락됨

**해결 방법**:

```bash
# 1. 프로젝트 디렉토리에서 재초기화
moai init . --force

# 2. .claude/ 디렉토리 확인
ls -la .claude/agents/
# alfred.yaml
# spec-builder.yaml
# code-builder.yaml
# ...

# 3. Claude Code 재시작
# Claude Code를 완전히 종료하고 다시 실행

# 4. 재확인
claude
# Claude Code 내에서 /alfred 입력 시 자동완성 확인
```

---

### 문제 5: 의존성 충돌

**증상**:
```bash
pip install moai-adk
# ERROR: pip's dependency resolver does not currently take into account all the packages that are installed.
```

**원인**: 다른 패키지와 의존성 버전 충돌

**해결 방법**:

#### A. uv 사용 (자동 해결)

```bash
# uv는 의존성 충돌을 자동으로 해결
uv tool install moai-adk
```

#### B. 가상환경 사용

```bash
# 깨끗한 가상환경 생성
python -m venv moai-env
source moai-env/bin/activate

# 설치
pip install moai-adk
```

#### C. pip 업그레이드

```bash
# pip 최신 버전으로 업그레이드
python -m pip install --upgrade pip

# 재설치
pip install moai-adk
```

---

### 문제 6: Windows에서 설치 실패

**증상**:
```powershell
pip install moai-adk
# ERROR: Microsoft Visual C++ 14.0 is required
```

**원인**: C++ 빌드 도구 없음 (일부 Python 패키지는 C++ 컴파일 필요)

**해결 방법**:

#### A. Microsoft C++ Build Tools 설치

1. [Visual Studio Build Tools](https://visualstudio.microsoft.com/downloads/) 다운로드
2. "Desktop development with C++" 워크로드 선택
3. 설치 후 재시도

#### B. 사전 빌드된 휠 사용

```powershell
# 사전 빌드된 바이너리 설치
pip install moai-adk --only-binary :all:
```

#### C. WSL2 사용 (권장)

```powershell
# WSL2 설치 (Windows 11)
wsl --install

# Ubuntu 실행
wsl

# WSL 내에서 설치
pip install moai-adk
```

---

## 다음 단계

### 1. Quick Start 가이드

설치가 완료되었다면, 3분 만에 첫 프로젝트를 시작하세요:

➡️ **[Quick Start Guide](./quick-start.md)**

### 2. 첫 프로젝트 튜토리얼

실제 Todo 앱을 만들면서 MoAI-ADK의 3단계 워크플로우를 배우세요:

➡️ **[First Project Tutorial](./first-project.md)**

### 3. Alfred SuperAgent 가이드

10개 AI 에이전트 팀과 함께 개발하는 방법:

➡️ **[Alfred SuperAgent Guide](https://moai-adk.vercel.app/guides/alfred-superagent/)**

---

## 추가 리소스

### 공식 문서

- **[전체 문서 사이트](https://moai-adk.vercel.app)**
- **[SPEC-First TDD 가이드](https://moai-adk.vercel.app/guides/spec-first-tdd/)**
- **[TAG 시스템 가이드](https://moai-adk.vercel.app/guides/tag-system/)**
- **[TRUST 원칙](https://moai-adk.vercel.app/guides/trust-principles/)**

### 커뮤니티

- **[GitHub Repository](https://github.com/modu-ai/moai-adk)**
- **[GitHub Issues](https://github.com/modu-ai/moai-adk/issues)**
- **[GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)**

### 패키지

- **[PyPI Package](https://pypi.org/project/moai-adk/)**

---

**마지막 업데이트**: 2025-10-14
**버전**: v0.3.0
