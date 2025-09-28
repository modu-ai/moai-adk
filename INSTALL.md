# MoAI-ADK 설치 가이드

🗿 **MoAI-ADK 설치 방법** - 수강생을 위한 간단한 설치 안내

## 🚀 빠른 설치 (권장)

### Windows 사용자

#### 방법 1: 단일 실행 파일 (Python 설치 불필요) ⭐
1. [GitHub Releases](https://github.com/modu-ai/moai-adk/releases/latest)에서 `moai-adk.exe` 다운로드
2. 다운로드한 파일을 더블클릭하여 실행
3. 명령 프롬프트에서 `moai-adk.exe init` 실행

#### 방법 2: 자동 설치 스크립트
```powershell
# PowerShell에서 실행 (관리자 권한 권장)
iwr https://raw.githubusercontent.com/modu-ai/moai-adk/main/scripts/install.ps1 -UseBasicParsing | iex
```

### macOS/Linux 사용자

#### 원클릭 설치 스크립트
```bash
# 터미널에서 실행
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/scripts/install.sh | bash
```

### 모든 플랫폼 (uv 사용) ⭐

```bash
# uv가 Python을 자동으로 다운로드하고 실행
uvx --from moai-adk moai-adk init
```

## 📋 설치 후 확인

설치가 완료되면 다음 명령어로 정상 설치를 확인하세요:

```bash
# 버전 확인
moai-adk --version

# 도움말 보기
moai-adk --help

# 시스템 진단
moai-adk doctor
```

## 🏃 첫 프로젝트 시작

```bash
# 새 프로젝트 생성
moai-adk init my-first-project

# 또는 현재 디렉토리에 초기화
moai-adk init
```

## 🔧 문제 해결

### 일반적인 문제

#### "command not found" 오류
```bash
# Windows: PATH에 추가되지 않은 경우
# 설치 스크립트를 다시 실행하거나 직접 실행 파일 경로 사용

# macOS/Linux: shell 재시작 필요
source ~/.bashrc
# 또는
source ~/.zshrc
```

#### Python 관련 오류
```bash
# uv 사용으로 해결 (Python 자동 설치)
uvx --from moai-adk moai-adk --help
```

#### 권한 오류 (macOS/Linux)
```bash
# 설치 스크립트에 실행 권한 부여
chmod +x install.sh
./install.sh
```

### 환경별 세부 설치

#### Windows 상세 설치

**요구사항**: Windows 10/11 (x64)

1. **단일 실행 파일 방식** (추천)
   - Python 설치 불필요
   - 모든 의존성 포함
   - 오프라인 실행 가능

2. **PowerShell 스크립트 방식**
   - uv 자동 설치
   - 최신 버전 자동 다운로드
   - 환경 변수 자동 설정

#### macOS 상세 설치

**요구사항**: macOS 10.15+ (Intel/Apple Silicon)

```bash
# Homebrew가 설치되어 있다면 uv 설치
brew install uv

# 그 다음 MoAI-ADK 실행
uvx --from moai-adk moai-adk init
```

#### Linux 상세 설치

**요구사항**: Ubuntu 20.04+, CentOS 8+, 또는 동등한 배포판

```bash
# 자동 설치 스크립트 사용 (권장)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/scripts/install.sh | bash

# 수동 설치
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.cargo/env
uvx --from moai-adk moai-adk init
```

## 📚 다음 단계

설치 완료 후:

1. **[MoAI-ADK 사용법](README.md#사용법)** 확인
2. **[개발 가이드](CLAUDE.md)** 읽기
3. **[예제 프로젝트](examples/)** 살펴보기

## 🆘 도움이 필요한 경우

- **문제 보고**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **설치 문의**: [Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **버그 신고**: `moai-adk doctor` 결과와 함께 이슈 생성

---

**설치 문제가 지속되면 `moai-adk doctor` 명령어 결과를 이슈에 포함해 주세요.**