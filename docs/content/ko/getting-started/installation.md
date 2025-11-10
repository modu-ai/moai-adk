# 설치 가이드

몇 분 만에 시스템에 MoAI-ADK를 설치하고 실행하세요. 이 가이드는 시스템 요구사항, 설치 방법, 확인 단계를 다룹니다.

## 시스템 요구사항

### 최소 요구사항

- **Python**: 3.13 이상
- **운영체제**:
  - macOS (10.15+)
  - Linux (Ubuntu 20.04+, CentOS 8+, Debian 11+)
  - Windows 10+ (PowerShell 권장)
- **Git**: 2.25 이상
- **메모리**: 4GB RAM 최소, 8GB 권장
- **저장공간**: 500MB 여유 공간

### 권장 요구사항

- **Python**: 3.13+ (최신 안정 버전)
- **패키지 관리자**: UV 0.5.0+ (권장) 또는 pip 24.0+
- **IDE**: Claude Code 확장 프로그램이 설치된 VS Code 또는 선호하는 편집기
- **터미널**: UTF-8을 지원하는 현대적 터미널

## 설치 방법

### 방법 1: UV 패키지 관리자 (권장)

UV는 MoAI-ADK를 설치하는 가장 빠르고 신뢰할 수 있는 방법입니다. 자동 의존성 관리와 가상 환경 처리를 제공합니다.

#### 1단계: UV 설치

**macOS/Linux:**

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

**Windows (PowerShell):**

```powershell
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"
```

#### 2단계: UV 설치 확인

```bash
uv --version
# 예상 출력: uv 0.5.1 이상
```

#### 3단계: MoAI-ADK 설치

```bash
uv tool install moai-adk
```

#### 4단계: 설치 확인

```bash
moai-adk --version
# 예상 출력: MoAI-ADK v1.0.0 이상
```

### 방법 2: PyPI 설치 (대안)

pip를 사용하거나 UV를 사용할 수 없는 경우입니다.

#### 1단계: pip 업그레이드 (필요한 경우)

```bash
python -m pip install --upgrade pip
```

#### 2단계: MoAI-ADK 설치

```bash
pip install moai-adk
```

#### 3단계: 설치 확인

```bash
moai-adk --version
```

### 방법 3: 개발용 설치

MoAI-ADK에 기여하고 싶은 개발자용입니다.

#### 1단계: 저장소 클론

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

#### 2단계: 개발 모드로 설치

```bash
# UV 사용 (권장)
uv pip install -e .

# 또는 pip 사용
pip install -e .
```

#### 3단계: 설치 확인

```bash
moai-adk --version
```

## 설치 후 설정

### 환경 변수

선택적이지만 권장되는 환경 변수:

```bash
# 셸 프로필에 추가 (~/.bashrc, ~/.zshrc 등)
export MOAI_LOG_LEVEL=INFO
export MOAI_CACHE_DIR="$HOME/.moai/cache"
export CLAUDE_PROJECT_DIR=$(pwd)
```

### Claude Code 통합

MoAI-ADK는 전체 경험을 위해 Claude Code가 필요합니다.

#### Claude Code 설치

```bash
# macOS
brew install claude-ai/claude/claude

# Linux
curl -fsSL https://claude.ai/install.sh | sh

# Windows
winget install Anthropic.Claude
```

#### Claude Code 확인

```bash
claude --version
# 예상: Claude Code v1.5.0 이상
```

### 선택적 MCP 서버

MoAI-ADK는 향상된 기능을 위해 Model Context Protocol (MCP) 서버를 지원합니다.

#### 권장 MCP 서버 설치

```bash
# Context7 - 최신 라이브러리 문서
npx -y @upstash/context7-mcp

# Playwright - 웹 E2E 테스트
npx -y @playwright/mcp

# Sequential Thinking - 복잡한 추론
npx -y @modelcontextprotocol/server-sequential-thinking
```

## 확인

### 시스템 상태 확인

내장된 doctor 명령어를 실행하여 설치를 확인하세요:

```bash
moai-adk doctor
```

**예상 출력:**

```
시스템 진단 실행 중...

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┓
┃ 확인                                    ┃ 상태 ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━┩
│ Python >= 3.13                           │   ✓    │
│ uv 설치됨                               │   ✓    │
│ Git 설치됨                              │   ✓    │
│ Claude Code 사용 가능                    │   ✓    │
│ 패키지 레지스트리 접근 가능              │   ✓    │
└──────────────────────────────────────────┴────────┘

✅ 모든 검사 통과!
```

### 테스트 프로젝트 생성

모든 것이 작동하는지 확인하기 위해 간단한 테스트 프로젝트를 생성하세요:

```bash
# 테스트 프로젝트 생성
moai-adk init test-project
cd test-project

# Claude Code 시작
claude

# Claude Code에서 다음 실행:
/alfred:0-project
```

## 문제 해결

### 일반적인 문제

#### 문제: "uv: command not found"

**해결책:**

1. UV가 올바르게 설치되었는지 확인
2. UV를 PATH에 추가:
   ```bash
   export PATH="$HOME/.cargo/bin:$PATH"
   ```
3. 터미널 재시작

#### 문제: "Python 3.8 found, but 3.13+ required"

**해결책:**

```bash
# pyenv 사용
curl https://pyenv.run | bash
pyenv install 3.13
pyenv global 3.13

# 또는 UV 사용
uv python install 3.13
uv python pin 3.13
```

#### 문제: 설치 중 "Permission denied"

**해결책:**

```bash
# 사용자 설치 사용
pip install --user moai-adk

# 또는 sudo 사용 (Linux/macOS)
sudo pip install moai-adk
```

#### 문제: Claude Code를 인식하지 못함

**해결책:**

1. Claude Code 설치 확인: `claude --version`
2. PATH에 있는지 확인
3. 필요한 경우 재설치

#### 문제: 의존성에 대한 ModuleNotFoundError

**해결책:**

```bash
# 프로젝트 디렉터리에서
uv sync

# 또는 특정 의존성 설치
uv add fastapi pytest
```

### 도움 얻기

여기에서 다루지 않는 문제가 발생한 경우:

1. **GitHub Issues 확인**: https://github.com/modu-ai/moai-adk/issues에서 기존 이슈 검색
2. **상세 진단 실행**: `moai-adk doctor --verbose`
3. **이슈 생성**: Claude Code에서 `/alfred:9-feedback`를 사용하여 자동으로 GitHub 이슈 생성

## 다음 단계

성공적인 설치 후:

1. **[빠른 시작 가이드](quick-start.md)** - 10분 안에 첫 프로젝트 실행
2. **[핵심 개념](concepts.md)** - SPEC-First, TDD, @TAG, TRUST 5 원칙 이해
3. **[프로젝트 초기화](guides/project/init.md)** - 프로젝트 설정 및 구성 학습

## 설치 요약

```bash
# 원라인 설치 (권장)
curl -LsSf https://astral.sh/uv/install.sh | sh && uv tool install moai-adk

# 설치 확인
moai-adk doctor

# 첫 프로젝트 생성
moai-adk init my-project && cd my-project && claude
```

이제 Alfred SuperAgent와 함께 SPEC-First TDD 개발의 강력한 기능을 경험할 준비가 되었습니다!
