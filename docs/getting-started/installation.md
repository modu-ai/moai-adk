# MoAI-ADK 설치 가이드

이 가이드는 MoAI-ADK를 시스템에 설치하고 개발 환경을 구축하는 전체 과정을 안내합니다. UV 패키지 매니저를 중심으로 한 현대적인 Python 개발 환경을 설정하고, MCP (Model Context Protocol) 서버를 통해 Claude Code의 기능을 확장하는 방법을 상세히 설명합니다.

## 목차

1. [시스템 요구사항](#시스템-요구사항)
2. [UV 패키지 매니저 설치](#uv-패키지-매니저-설치)
3. [Python 3.13+ 설치](#python-313-설치)
4. [MoAI-ADK 패키지 설치](#moai-adk-패키지-설치)
5. [MCP 서버 설정](#mcp-서버-설정)
6. [Claude Code 설정 및 Alfred 활성화](#claude-code-설정-및-alfred-활성화)
7. [설치 후 검증](#설치-후-검증)
8. [운영체제별 설치 가이드](#운영체제별-설치-가이드)
9. [문제 해결](#문제-해결)
10. [추가 도구 및 설정](#추가-도구-및-설정)

---

## 시스템 요구사항

MoAI-ADK를 설치하기 전에 다음 시스템 요구사항을 확인하세요.

### 최소 요구사항

| 항목 | 최소 버전 | 권장 버전 | 비고 |
|------|-----------|------------|------|
| **운영체제** | - | macOS 14+, Ubuntu 20.04+, Windows 10+ | 현대적인 OS 권장 |
| **Python** | 3.13.0+ | 3.13.0+ | 필수 - UV가 자동 설치 가능 |
| **Git** | 2.30.0+ | 2.40.0+ | 버전 관리 필수 |
| **메모리** | 4GB RAM | 8GB+ RAM | AI 작업 시 더 많은 메모리 권장 |
| **저장공간** | 2GB 여유 공간 | 5GB+ 여유 공간 | 프로젝트 및 캐시 고려 |

### 선택적 요구사항

| 항목 | 목적 | 설치 방법 |
|------|------|-----------|
| **Node.js** | MCP 서버 실행 | 18.0.0+ 권장 |
| **GitHub CLI** | PR 자동화 | `gh` 명령어 |
| **Docker** | 배포 및 테스트 | 20.10.0+ 권장 |

### 시스템 호환성 확인

다음 명령어로 시스템 호환성을 확인하세요:

```bash
# 기본 정보 확인
echo "OS: $(uname -s)"
echo "Architecture: $(uname -m)"
echo "Shell: $SHELL"

# Python 버전 확인 (설치된 경우)
python3 --version 2>/dev/null || echo "Python 3 not found"

# Git 버전 확인
git --version 2>/dev/null || echo "Git not found"

# Node.js 버전 확인 (선택적)
node --version 2>/dev/null || echo "Node.js not found (optional)"
```

---

## UV 패키지 매니저 설치

UV는 현대적인 Python 패키지 매니저로, 빠른 속도와 효율적인 종속성 관리를 제공합니다. MoAI-ADK는 UV를 중심으로 설계되었습니다.

### UV란 무엇인가요?

UV는 Rust로 작성된 차세대 Python 패키지 매니저입니다.

- **속도**: pip보다 10-100배 빠른 종속성 해결 및 설치
- **효율성**: 최소한의 디스크 사용량과 빠른 캐싱
- **호환성**: 기존 pip와 완전 호환
- **통합**: 가상환경, 패키지 설치, 도구 관리를 하나로 통합

### UV 설치 방법

#### 방법 1: 공식 설치 스크립트 (권장)

**macOS 및 Linux:**

```bash
# 설치 스크립트 다운로드 및 실행
curl -LsSf https://astral.sh/uv/install.sh | sh

# 또는 wget 사용
wget -qO- https://astral.sh/uv/install.sh | sh
```

**Windows (PowerShell):**

```powershell
# PowerShell에서 설치 스크립트 실행
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### 방법 2: 패키지 매니저를 통한 설치

**macOS (Homebrew):**

```bash
# Homebrew로 설치
brew install uv

# 또는 cask 사용
brew install --cask uv
```

**Windows (Scoop):**

```powershell
# Scoop으로 설치
scoop install uv
```

**Windows (Chocolatey):**

```powershell
# Chocolatey로 설치
choco install uv
```

#### 방법 3: cargo를 통한 설치 (Rust 개발자)

```bash
# Rust가 설치된 경우
cargo install uv

# 또는 최신 버전 설치
cargo install uv --locked
```

### UV 설치 확인

```bash
# UV 버전 확인
uv --version

# 예상 출력:
# uv 0.5.1 (a1b2c3d4 2024-01-15)
```

### PATH 설정

설치 후 UV가 PATH에 추가되지 않은 경우:

**macOS/Linux:**

```bash
# bash/zsh 셸에 PATH 추가
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
# 또는 ~/.zshrc (zsh 사용자)

# 즉시 적용
source ~/.bashrc  # 또는 source ~/.zshrc
```

**Windows:**

```powershell
# 환경변수에 PATH 추가
[System.Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\.cargo\bin", "User")
```

### UV 기본 명령어

```bash
# UV 도움말
uv --help

# 자동 완성 설정
# bash/zsh:
echo 'eval "$(uv --completion-script bash)"' >> ~/.bashrc
echo 'eval "$(uv --completion-script zsh)"' >> ~/.zshrc

# 패키지 설치
uv add package-name

# 개발 의존성 설치
uv add --dev package-name

# 가상환경 생성 및 활성화
uv venv
source .venv/bin/activate  # macOS/Linux
.venv\Scripts\activate     # Windows

# 스크립트 실행
uv run python script.py
uv run pytest
```

---

## Python 3.13+ 설치

MoAI-ADK는 Python 3.13 이상을 필요로 합니다. UV를 사용하면 Python 버전 관리가 간편해집니다.

### 현재 Python 버전 확인

```bash
# 시스템 Python 버전 확인
python3 --version
python --version  # 일부 시스템에서

# UV가 관리하는 Python 목록
uv python list
```

### UV를 통한 Python 설치 (권장)

UV는 Python 버전 관리를 자동으로 처리합니다.

```bash
# 최신 Python 3.13 설치
uv python install 3.13

# 특정 버전 설치
uv python install 3.13.1

# 프로젝트에 Python 버전 고정
uv python pin 3.13

# 설치된 Python 확인
uv python list

# 예상 출력:
# python3.13.1           /home/user/.local/share/uv/python/cpython-3.13.1-linux-x86_64-gnu/bin/python3
# python3.12.0           /usr/bin/python3
```

### pyenv를 통한 Python 설치 (대체 방법)

pyenv는 다중 Python 버전 관리를 위한 도구입니다.

#### pyenv 설치

**macOS:**

```bash
# Homebrew로 설치
brew install pyenv

# 또는 직접 설치
curl https://pyenv.run | bash
```

**Linux:**

```bash
# 의존성 설치 (Ubuntu/Debian)
sudo apt update
sudo apt install -y make build-essential libssl-dev zlib1g-dev \
    libbz2-dev libreadline-dev libsqlite3-dev wget curl llvm \
    libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev libffi-dev liblzma-dev

# pyenv 설치
curl https://pyenv.run | bash
```

**pyenv 설정:**

```bash
# 셸 설정 파일에 추가
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.bashrc
echo 'command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(pyenv init -)"' >> ~/.bashrc

# zsh 사용자
echo 'export PYENV_ROOT="$HOME/.pyenv"' >> ~/.zshrc
echo 'command -v pyenv >/dev/null || export PATH="$PYENV_ROOT/bin:$PATH"' >> ~/.zshrc
echo 'eval "$(pyenv init -)"' >> ~/.zshrc

# 즉시 적용
exec "$SHELL"
```

#### pyenv로 Python 설치

```bash
# 설치 가능한 Python 버전 목록
pyenv install --list | grep " 3.13"

# Python 3.13 설치
pyenv install 3.13.1

# 전역 기본 버전 설정
pyenv global 3.13.1

# 프로젝트별 버전 설정
cd your-project
pyenv local 3.13.1

# 버전 확인
python --version
# Python 3.13.1
```

### 시스템 패키지 매니저를 통한 설치

**Ubuntu/Debian:**

```bash
# deadsnakes PPA 추가
sudo apt update
sudo apt install -y software-properties-common
sudo add-apt-repository ppa:deadsnakes/ppa
sudo apt update

# Python 3.13 설치
sudo apt install -y python3.13 python3.13-venv python3.13-dev python3.13-distutils

# 기본 python3 심볼릭 링크 업데이트
sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.13 1
```

**macOS (Homebrew):**

```bash
# Python 3.13 설치
brew install python@3.13

# PATH 설정
echo 'export PATH="$(brew --prefix python@3.13)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Windows (공식 설치 프로그램):**

1. [python.org](https://www.python.org/downloads/)에서 Python 3.13 다운로드
2. 설치 프로그램 실행 시 "Add Python to PATH" 체크
3. "Install for all users" 선택 (권장)

### Python 환경 확인

```bash
# Python 버전 확인
python --version
# Python 3.13.1

# pip 버전 확인
pip --version
# pip 24.0 from ...

# Python 경로 확인
which python
# /home/user/.local/share/uv/python/cpython-3.13.1-linux-x86_64-gnu/bin/python

# pip와 Python 호환성 확인
python -m pip --version
```

---

## MoAI-ADK 패키지 설치

이제 MoAI-ADK 패키지를 설치할 준비가 되었습니다.

### UV Tool을 통한 설치 (권장)

UV Tool은 시스템 전역에서 명령어를 관리하는 최상의 방법입니다.

```bash
# MoAI-ADK 설치
uv tool install moai-adk

# 설치 확인
moai-adk --version
# MoAI-ADK v0.17.0

# 설치된 도구 목록 확인
uv tool list
# moai-adk 0.17.0
```

### pip를 통한 설치 (대체 방법)

```bash
# pip로 설치
pip install moai-adk

# 사용자 설치 (권한 문제 방지)
pip install --user moai-adk

# 설치 확인
moai-adk --version
```

### 개발 버전 설치

최신 개발 버전을 설치하려면:

```bash
# GitHub에서 직접 설치
uv tool install git+https://github.com/modu-ai/moai-adk.git

# 또는 로컬 클론에서 설치
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
uv tool install -e .
```

### 설치된 도구 확인

```bash
# 도구 위치 확인
which moai-adk
# /home/user/.local/bin/moai-adk

# 도구 정보 확인
moai-adk --help
# Usage: moai-adk [OPTIONS] COMMAND [ARGS]...
#
# Options:
#   --version  Show the version and exit.
#   --help     Show this message and exit.
#
# Commands:
#   init     Initialize a new MoAI-ADK project
#   doctor   Check system requirements and configuration
#   update   Update MoAI-ADK and sync templates
```

---

## MCP 서버 설정

MCP (Model Context Protocol)는 Claude Code의 기능을 확장하는 표준 프로토콜입니다. MoAI-ADK는 4개의 핵심 MCP 서버를 자동으로 설정합니다.

### MCP 서버 종류 및 기능

| 서버 이름 | 기능 | 설치 방식 | 주요 사용 사례 |
|-----------|------|-----------|---------------|
| **Context7** | 최신 라이브러리 문서 검색 | NPX 자동 설치 | 실시간 API 문서 조회 |
| **Figma** | 디자인 시스템 연동 | Claude Code 원격 서버 | UI/UX 디자인 검토 |
| **Playwright** | E2E 테스트 자동화 | NPX 자동 설치 | 웹 애플리케이션 테스트 |
| **Sequential Thinking** | 복잡 추론 지원 | NPX 자동 설치 | 논리적 문제 해결 |

### 자동 MCP 설정

MoAI-ADK 프로젝트를 초기화하면 MCP 서버가 자동으로 설정됩니다:

```bash
# MCP 서버 포함 프로젝트 초기화
moai-adk init my-project --with-mcp

# 기존 프로젝트에 MCP 추가
cd existing-project
moai-adk init . --with-mcp
```

### 수동 MCP 설정

#### 1. MCP 설정 파일 생성

Claude Code 설정 디렉터리에 MCP 설정 파일을 생성합니다:

```bash
# Claude Code 설정 디렉터리 확인
ls -la ~/.claude/

# MCP 설정 파일 위치
# Linux/macOS: ~/.claude/mcp.json
# Windows: %APPDATA%\Claude\mcp.json
```

#### 2. MCP 설정 파일 작성

다음 내용으로 `~/.claude/mcp.json` 파일을 생성합니다:

```json
{
  "servers": {
    "context7": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@upstash/context7-mcp"
      ],
      "env": {}
    },
    "figma": {
      "type": "http",
      "url": "https://mcp.figma.com/mcp",
      "headers": {
        "Authorization": "Bearer ${FIGMA_ACCESS_TOKEN}"
      }
    },
    "playwright": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@playwright/mcp"
      ],
      "env": {}
    },
    "sequential-thinking": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-sequential-thinking"
      ],
      "env": {}
    }
  }
}
```

### Node.js 설치 (MCP 서버용)

MCP 서버는 Node.js 기반이므로 Node.js 18+가 필요합니다.

#### Node.js 설치 방법

**방법 1: 공식 설치 프로그램**

1. [nodejs.org](https://nodejs.org/)에서 LTS 버전 다운로드
2. 설치 프로그램 실행

**방법 2: UV를 통한 설치**

```bash
# Node.js 20 LTS 설치
uv tool install node@20

# 또는 nvm 사용 (권장)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 20
nvm use 20
```

**방법 3: 시스템 패키지 매니저**

```bash
# macOS (Homebrew)
brew install node

# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
```

#### Node.js 설치 확인

```bash
# Node.js 버전 확인
node --version
# v20.12.0

# npm 버전 확인
npm --version
# 10.5.0

# npx 확인
npx --version
# 10.5.0
```

### Figma Access Token 설정

Figma MCP 서버를 사용하려면 Access Token이 필요합니다.

#### Figma Access Token 생성

1. [Figma Developers](https://www.figma.com/developers/api#access-tokens) 방문
2. 로그인 후 "Generate new token" 클릭
3. 적절한 권한으로 토큰 생성
4. 토큰을 안전한 곳에 복사

#### 환경변수 설정

**macOS/Linux:**

```bash
# 셸 프로필에 추가
echo 'export FIGMA_ACCESS_TOKEN="your_figma_token_here"' >> ~/.zshrc
# 또는 ~/.bashrc

# 즉시 적용
source ~/.zshrc

# 확인
echo $FIGMA_ACCESS_TOKEN
```

**Windows (PowerShell):**

```powershell
# 사용자 환경변수 설정
[System.Environment]::SetEnvironmentVariable("FIGMA_ACCESS_TOKEN", "your_figma_token_here", "User")

# 또는 프로필에 추가
echo '$env:FIGMA_ACCESS_TOKEN = "your_figma_token_here"' >> $PROFILE
```

**Claude Code 설정 방법:**

```bash
# Claude Code 설정 실행
claude-code settings

# Environment variables 섹션에 추가
# FIGMA_ACCESS_TOKEN=your_figma_token_here
```

### MCP 서버 테스트

```bash
# Claude Code 재시작
claude

# MCP 서버가 로드되었는지 확인
# Claude Code가 시작될 때 MCP 서버 목록이 표시됩니다

# 각 서버 기능 테스트
# Context7: 라이브러리 문서 검색 요청
# Figma: 디자인 파일 접근 시도
# Playwright: 웹 테스트 자동화 명령
# Sequential Thinking: 복잡한 추론 요청
```

---

## Claude Code 설정 및 Alfred 활성화

이제 Claude Code를 설정하고 Alfred SuperAgent를 활성화할 차례입니다.

### Claude Code 설치

Claude Code가 아직 설치되지 않은 경우:

```bash
# macOS (Homebrew)
brew install claude-code

# 또는 공식 설치 프로그램 다운로드
# https://claude.ai/download
```

### Claude Code 로그인

```bash
# Claude Code 로그인
claude auth login

# 브라우저에서 인증 완료
# 로그인 확인
claude auth whoami
```

### Alfred 초기화

MoAI-ADK 프로젝트에서 Alfred를 초기화합니다:

```bash
# 프로젝트 생성 및 이동
moai-adk init hello-world
cd hello-world

# Claude Code 실행
claude

# Alfred 초기화 명령 실행
/alfred:0-project
```

#### Alfred 설정 과정

Alfred가 다음 정보를 수집합니다:

1. **프로젝트 이름**: 프로젝트의 고유 이름
2. **프로젝트 목표**: 프로젝트의 목적과 목표
3. **개발 언어**: 주요 개발 언어 (python, javascript, go 등)
4. **개발 모드**: personal 또는 team
5. **언어 설정**: conversation language (한국어, 영어 등)

#### Alfred 초기화 예시

```
Q1: 프로젝트 이름은 무엇인가요?
A: hello-world

Q2: 프로젝트의 목표는 무엇인가요?
A: MoAI-ADK 학습 및 테스트

Q3: 주요 개발 언어는 무엇인가요?
A: python

Q4: 개발 모드를 선택하세요
A: personal

Q5: 대화 언어를 선택하세요
A: 한국어

✅ 프로젝트 초기화 완료!
✅ .moai/config.json에 설정 저장
✅ .moai/project/에 문서 생성
✅ Alfred가 스킬 추천 완료

다음 단계: /alfred:1-plan "첫 기능 설명"
```

### Claude Code 권한 설정

Claude Code가 시스템 자원에 접근할 수 있도록 권한을 설정합니다:

```bash
# Claude Code 설정 파일 위치
ls -la ~/.claude/settings.json

# 권한 설정 (필요시)
chmod 644 ~/.claude/settings.json
```

### Claude Code Hooks 확인

MoAI-ADK는 5개의 자동화 Hook을 설치합니다:

```bash
# Hook 파일 확인
ls -la .claude/hooks/alfred/

# Hook 내용 확인
cat .claude/settings.json | grep -A 10 "hooks"
```

설치된 Hooks:
- **SessionStart**: 세션 시작 시 프로젝트 상태 요약
- **PreToolUse**: 위험 탐지 및 TAG Guard
- **UserPromptSubmit**: JIT 컨텍스트 로딩
- **PostToolUse**: 코드 변경 후 자동 테스트
- **SessionEnd**: 세션 정리 및 상태 보존

---

## 설치 후 검증

모든 구성요소가 올바르게 설치되었는지 확인합니다.

### 시스템 진단 실행

```bash
# MoAI-ADK 시스템 진단
moai-adk doctor

# 상세 진단
moai-adk doctor --verbose
```

**예상 출력:**

```
Running system diagnostics...

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┓
┃ Check                                    ┃ Status ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━┩
│ Python >= 3.13                           │   ✓    │
│ UV installed                             │   ✓    │
│ Git installed                            │   ✓    │
│ MoAI-ADK package                         │   ✓    │
│ .moai/ directory                         │   ✓    │
│ .claude/ directory                       │   ✓    │
│ MCP servers configured                   │   ✓    │
│ Claude Code installed                    │   ✓    │
└──────────────────────────────────────────┴────────┘

✓ All checks passed
```

### 프로젝트 구조 확인

```bash
# 프로젝트 구조 확인
tree -L 3 .

# 예상 출력:
# .
# ├── .claude/
# │   ├── agents/
# │   ├── commands/
# │   ├── skills/
# │   ├── hooks/
# │   ├── mcp.json
# │   └── settings.json
# ├── .moai/
# │   ├── config.json
# │   ├── project/
# │   ├── memory/
# │   ├── specs/
# │   └── reports/
# ├── CLAUDE.md
# └── README.md
```

### 기능 테스트

```bash
# Claude Code 실행
claude

# Alfred 기능 테스트
/alfred:0-project

# 간단한 SPEC 생성 테스트
/alfred:1-plan "테스트 API 엔드포인트"

# 도구 버전 확인
moai-adk --version
uv --version
python --version
claude --version
```

### MCP 서버 확인

```bash
# Claude Code에서 MCP 서버 목록 확인
# Claude Code 시작 시 로그에서 확인 가능

# Context7 테스트
# Claude Code에서: "FastAPI의 최신 문서를 검색해줘"

# Sequential Thinking 테스트
# Claude Code에서: "다음 문제를 단계별로 분석해줘: [복잡한 문제]"
```

---

## 운영체제별 설치 가이드

각 운영체제별 특이사항과 최적의 설치 방법을 안내합니다.

### macOS 설치 가이드

#### 시스템 요구사항

- macOS 14.0+ (Sonoma 이상)
- Xcode Command Line Tools
- Homebrew (권장)

#### 단계별 설치

**1단계: Xcode Command Line Tools 설치**

```bash
# Xcode Command Line Tools 설치
xcode-select --install

# 설치 확인
xcode-select -p
# /Applications/Xcode.app/Contents/Developer
```

**2단계: Homebrew 설치**

```bash
# Homebrew 설치
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# PATH 설정 (Apple Silicon Mac)
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zshrc
source ~/.zshrc

# 확인
brew --version
```

**3단계: UV 및 Python 설치**

```bash
# UV 설치
brew install uv

# Python 3.13 설치
uv python install 3.13

# 확인
uv --version
python --version
```

**4단계: MoAI-ADK 설치**

```bash
# MoAI-ADK 설치
uv tool install moai-adk

# Claude Code 설치
brew install claude-code

# 확인
moai-adk --version
claude --version
```

**5단계: Node.js (MCP용)**

```bash
# Node.js 설치
brew install node

# 확인
node --version
npm --version
```

#### macOS 특이사항

- **Apple Silicon**: Rosetta 없이 네이티브 실행 가능
- **보안**: 최초 실행 시 "허용" 필요할 수 있음
- **PATH**: `/opt/homebrew/bin`을 PATH에 추가 필요 (Apple Silicon)

### Linux 설치 가이드

#### Ubuntu/Debian

**1단계: 시스템 업데이트**

```bash
# 시스템 업데이트
sudo apt update && sudo apt upgrade -y

# 필수 패키지 설치
sudo apt install -y curl wget git build-essential \
    software-properties-common apt-transport-https \
    ca-certificates gnupg lsb-release
```

**2단계: UV 설치**

```bash
# UV 설치
curl -LsSf https://astral.sh/uv/install.sh | sh

# PATH 설정
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 확인
uv --version
```

**3단계: Python 설치**

```bash
# Python 3.13 설치 (UV 사용)
uv python install 3.13

# 또는 deadsnakes PPA 사용
sudo add-apt-repository ppa:deadsnakes/ppa
sudo apt update
sudo apt install -y python3.13 python3.13-venv python3.13-dev
```

**4단계: MoAI-ADK 설치**

```bash
# MoAI-ADK 설치
uv tool install moai-adk

# 확인
moai-adk --version
```

**5단계: Node.js (MCP용)**

```bash
# NodeSource 저장소 추가
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -

# Node.js 설치
sudo apt install -y nodejs

# 확인
node --version
npm --version
```

#### CentOS/RHEL/Fedora

**1단계: EPEL 저장소 활성화 (CentOS/RHEL)**

```bash
# EPEL 저장소 추가
sudo yum install -y epel-release
# 또는 (Fedora)
sudo dnf install -y fedora-repos-rawhide
```

**2단계: 개발 도구 설치**

```bash
# 개발 도구 그룹 설치
sudo yum groupinstall -y "Development Tools"
# 또는 (Fedora)
sudo dnf groupinstall -y "Development Tools"
```

**3단계: UV 설치**

```bash
# UV 설치
curl -LsSf https://astral.sh/uv/install.sh | sh

# PATH 설정
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**4단계: Python 설치**

```bash
# Python 3.13 설치 (UV 사용)
uv python install 3.13

# 또는 소스 컴파일
wget https://www.python.org/ftp/python/3.13.1/Python-3.13.1.tgz
tar xvf Python-3.13.1.tgz
cd Python-3.13.1
./configure --enable-optimizations
make -j$(nproc)
sudo make altinstall
```

### Windows 설치 가이드

#### 시스템 요구사항

- Windows 10 버전 1903+ (Build 18362+)
- Windows Terminal (권장)
- PowerShell 7+ (권장)

#### 단계별 설치

**1단계: Windows Terminal 설치**

```powershell
# Microsoft Store에서 Windows Terminal 설치
# 또는 winget 사용
winget install Microsoft.WindowsTerminal
```

**2단계: UV 설치**

```powershell
# UV 설치 스크립트 실행
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"

# PATH 재로드
$env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "User") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "Machine")

# 확인
uv --version
```

**3단계: Python 설치**

```powershell
# winget으로 Python 설치
winget install Python.Python.3.13

# 또는 공식 설치 프로그램 다운로드
# https://www.python.org/downloads/windows/

# 확인
python --version
```

**4단계: MoAI-ADK 설치**

```powershell
# MoAI-ADK 설치
uv tool install moai-adk

# 확인
moai-adk --version
```

**5단계: Claude Code 설치**

```powershell
# winget으로 Claude Code 설치
winget install Anthropic.ClaudeCode

# 또는 수동 다운로드
# https://claude.ai/download

# 확인
claude --version
```

**6단계: Node.js (MCP용)**

```powershell
# winget으로 Node.js 설치
winget install OpenJS.NodeJS

# 또는 Chocolatey 사용
choco install nodejs

# 확인
node --version
npm --version
```

#### Windows 특이사항

- **PowerShell 실행 정책**: 스크립트 실행을 위해 정책 조정 필요
- **PATH**: 사용자 PATH에 추가되도록 설치
- **긴 경로**: Windows 긴 경로 지원 활성화

```powershell
# PowerShell 실행 정책 확인
Get-ExecutionPolicy

# RemoteSigned로 변경 (필요시)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

---

## 문제 해결

설치 과정에서 발생할 수 있는 일반적인 문제와 해결 방법입니다.

### UV 관련 문제

#### 문제 1: UV 명령어를 찾을 수 없음

**증상:**
```bash
$ uv --version
bash: uv: command not found
```

**원인:** UV가 PATH에 없거나 설치가 실패함

**해결책:**

```bash
# UV 재설치
curl -LsSf https://astral.sh/uv/install.sh | sh

# PATH 수동 추가
echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 확인
which uv
# /home/user/.cargo/bin/uv
```

#### 문제 2: UV 권한 오류

**증상:**
```bash
$ uv tool install moai-adk
Permission denied: /home/user/.local/bin/moai-adk
```

**원인:** 디렉터리 권한 문제

**해결책:**

```bash
# 권한 수정
chmod 755 ~/.local/bin
chmod 755 ~/.local

# 또는 다른 위치에 설치
uv tool install moai-adk --install-dir ~/bin
export PATH="$HOME/bin:$PATH"
```

### Python 관련 문제

#### 문제 1: Python 버전 호환성

**증상:**
```bash
$ python --version
Python 3.11.0
```

**원인:** Python 3.13 미만 버전

**해결책:**

```bash
# UV로 Python 3.13 설치
uv python install 3.13

# 프로젝트에 고정
uv python pin 3.13

# 확인
python --version
```

#### 문제 2: 여러 Python 버전 충돌

**증상:**
```bash
$ which python
/usr/bin/python
$ python --version
Python 3.11.0
```

**해결책:**

```bash
# UV 관리 Python 확인
uv python list

# 우선순위 조정
uv python pin 3.13

# 가상환경 생성
uv venv --python 3.13
source .venv/bin/activate
```

### MoAI-ADK 관련 문제

#### 문제 1: moai-adk 명령어를 찾을 수 없음

**증상:**
```bash
$ moai-adk --version
bash: moai-adk: command not found
```

**해결책:**

```bash
# 재설치
uv tool install moai-adk --force-reinstall

# PATH 확인
echo $PATH | grep -o "[^:]*bin[^:]*"

# 수동 PATH 추가
export PATH="$HOME/.local/bin:$PATH"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

#### 문제 2: moai-adk doctor 실패

**증상:**
```bash
$ moai-adk doctor
❌ Python >= 3.13: Found 3.11.0
❌ Git not found
```

**해결책:**

```bash
# Python 버전 문제 해결
uv python install 3.13
uv python pin 3.13

# Git 설치
sudo apt install git  # Ubuntu/Debian
brew install git       # macOS
winget install Git.Git # Windows

# 재확인
moai-adk doctor
```

### Claude Code 관련 문제

#### 문제 1: Claude Code 로그인 실패

**증상:**
```bash
$ claude auth login
Error: Authentication failed
```

**해결책:**

```bash
# Claude Code 재시작
claude auth logout
claude auth login

# 브라우저 쿠키 초기화
# 브라우저에서 Claude 로그아웃 후 재시도

# 확인
claude auth whoami
```

#### 문제 2: /alfred 명령어를 찾을 수 없음

**증상:**
```bash
claude> /alfred:0-project
Unknown command: /alfred:0-project
```

**해결책:**

```bash
# .claude 디렉터리 확인
ls -la .claude/commands/

# 프로젝트 재초기화
moai-adk init .

# Claude Code 재시작
exit
claude

# 버전 확인
claude --version  # 1.5.0+ 필요
```

### MCP 서버 관련 문제

#### 문제 1: MCP 서버가 로드되지 않음

**증상:**
Claude Code 시작 시 MCP 서버 목록에 서버가 표시되지 않음

**해결책:**

```bash
# mcp.json 파일 확인
cat ~/.claude/mcp.json

# Node.js 확인
node --version  # 18+ 필요
npm --version

# npx 테스트
npx --version

# Claude Code 재시작
exit
claude
```

#### 문제 2: Figma Access Token 오류

**증상:**
```
Error: Invalid Figma access token
```

**해결책:**

```bash
# 토큰 확인
echo $FIGMA_ACCESS_TOKEN

# 토큰 재설정
export FIGMA_ACCESS_TOKEN="new_token_here"

# Claude Code 설정 파일 확인
cat ~/.claude/settings.json | grep FIGMA

# 재시작
claude
```

### 권한 관련 문제

#### 문제 1: 훅(Hook) 실행 권한 오류

**증상:**
```
Permission denied: .claude/hooks/alfred/session_start.py
```

**해결책:**

```bash
# 훅 권한 수정
chmod +x .claude/hooks/alfred/*.py
chmod +x .claude/hooks/alfred/core/*.py

# 소유자 확인
ls -la .claude/hooks/alfred/

# 재시작
claude
```

#### 문제 2: 프로젝트 디렉터리 권한

**증상:**
```
Error: Cannot write to .moai/ directory
```

**해결책:**

```bash
# 디렉터리 권한 수정
chmod 755 .moai/
chmod 644 .moai/config.json

# 소유자 변경 (필요시)
sudo chown -R $USER:$USER .moai/
sudo chown -R $USER:$USER .claude/
```

### 일반적인 디버깅 명령어

```bash
# 시스템 정보 수집
moai-adk doctor --verbose > system-info.txt

# 환경변수 확인
env | grep -E "(PATH|PYTHON|UV|CLAUDE|FIGMA)"

# 설치된 도구 목록
uv tool list
pip list | grep moai

# 프로젝트 구조 확인
tree -a .moai/ .claude/

# 로그 파일 확인
tail -f ~/.claude/logs/*.log  # 있는 경우

# 네트워크 연결 확인
curl -I https://pypi.org/project/moai-adk/
curl -I https://api.github.com
```

### 지원 요청

문제가 해결되지 않으면 다음 정보를 포함하여 지원을 요청하세요:

1. **시스템 정보:**
   ```bash
   uname -a
   python --version
   uv --version
   moai-adk --version
   claude --version
   ```

2. **오류 메시지:** 전체 오류 출력

3. **진단 결과:**
   ```bash
   moai-adk doctor --verbose
   ```

4. **재현 단계:** 문제가 발생하는 정확한 단계

5. **구성 파일:** `.moai/config.json` (민감정보 제외)

---

## 추가 도구 및 설정

개발 경험을 향상시키기 위한 추가 도구와 설정입니다.

### Git 설정

```bash
# Git 사용자 설정
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# 기본 브랜치 이름 설정
git config --global init.defaultBranch main

# Git aliases 설정
git config --global alias.st status
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit

# 확인
git config --list
```

### 개발 환경 설정

#### Python 개발 도구

```bash
# 개발 도구 설치
uv add --dev pytest black ruff mypy

# pre-commit 설정
uv add --dev pre-commit
pre-commit install

# .pre-commit-config.yaml 생성
cat > .pre-commit-config.yaml << EOF
repos:
  - repo: https://github.com/psf/black
    rev: 24.1.1
    hooks:
      - id: black
  - repo: https://github.com/charliermarsh/ruff-pre-commit
    rev: v0.1.9
    hooks:
      - id: ruff
        args: [--fix]
EOF
```

#### 코드 편집기 설정

**VS Code 설정 (`.vscode/settings.json`):**

```json
{
  "python.defaultInterpreterPath": "./.venv/bin/python",
  "python.linting.enabled": true,
  "python.linting.ruffEnabled": true,
  "python.formatting.provider": "black",
  "python.testing.pytestEnabled": true,
  "python.testing.pytestArgs": ["tests/"],
  "files.exclude": {
    "**/__pycache__": true,
    "**/*.pyc": true
  }
}
```

**VS Code 확장 프로그램 추천:**

- Python
- Pylance
- Black Formatter
- Ruff
- GitLens
- Claude Code

### 셸 설정

#### bash/zsh 설정 추가

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가

# UV와 Python
export PATH="$HOME/.cargo/bin:$HOME/.local/bin:$PATH"
eval "$(uv --completion-script bash)"  # zsh의 경우 --completion-script zsh

# Claude Code
export CLAUDE_API_KEY="your_api_key_here"  # 필요시

# Figma (MCP)
export FIGMA_ACCESS_TOKEN="your_figma_token_here"

# MoAI-ADK aliases
alias moai='moai-adk'
alias moai-doc='moai-adk doctor'
alias moai-update='moai-adk update'

# 프로젝트별 활성화 함수
moai-activate() {
    if [ -f ".venv/bin/activate" ]; then
        source .venv/bin/activate
        echo "✅ Virtual environment activated"
    fi

    if [ -d ".moai" ]; then
        echo "✅ MoAI-ADK project detected"
        moai-adk doctor
    fi
}
```

### 프로젝트 템플릿

#### pyproject.toml 기본 설정

```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "your-project-name"
version = "0.1.0"
description = "Project description"
authors = [{name = "Your Name", email = "your.email@example.com"}]
license = {text = "MIT"}
readme = "README.md"
requires-python = ">=3.13"
dependencies = []

[project.optional-dependencies]
dev = [
    "pytest>=7.0.0",
    "black>=23.0.0",
    "ruff>=0.1.0",
    "mypy>=1.0.0",
]

[tool.black]
line-length = 88
target-version = ['py313']

[tool.ruff]
line-length = 88
target-version = "py313"

[tool.mypy]
python_version = "3.13"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true

[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
```

### 성능 최적화

#### UV 캐시 설정

```bash
# UV 캐시 위치 확인
uv cache dir

# 캐시 크기 제한 (GB)
export UV_CACHE_SIZE=10

# 캬시 정리
uv cache clean

# 캐시 정보
uv cache info
```

#### Claude Code 성능

```bash
# Claude Code 캐시 정리
claude cache clear

# Claude Code 설정 최적화
cat > ~/.claude/settings.json << EOF
{
  "model": "claude-3-5-sonnet-20241022",
  "max_tokens": 8192,
  "temperature": 0.1,
  "timeout": 120,
  "auto_save": true,
  "cache_enabled": true
}
EOF
```

이제 MoAI-ADK가 완전히 설치되고 설정되었습니다! 다음 단계로 [기본 개념](concepts.md) 문서를 읽어보세요.