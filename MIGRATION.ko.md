# MoAI-ADK v2.0 마이그레이션 가이드

MoAI-ADK v2.0은 Python 기반 v1.x(~73,000줄)을 Go로 완전히 재작성한 차세대 Agentic Development Kit입니다.

---

## 목차

- [v2.0 개요](#v20-개요)
- [1단계: Go 설치 확인](#1단계-go-설치-확인)
- [2단계: v1.x 제거하기](#2단계-v1x-제거하기)
- [3단계: MoAI-ADK v2.0 설치하기](#3단계-moai-adk-v20-설치하기)
- [1.x와 2.0의 주요 차이점](#1x와-20의-주요-차이점)
- [다가오는 기능: Agent Teams 모드](#다가오는-기능-agent-teams-모드)

---

## v2.0 개요

### 왜 Go인가?

| 항목 | Python 1.x 버전 | Go 2.0 버전 |
|------|-----------------|-------------|
| 배포 방식 | pip install + venv + 의존성 | 단일 바이너리, 의존성 없음 |
| 시작 속도 | ~800ms 인터프리터 부팅 | ~5ms 네이티브 실행 |
| 동시성 | asyncio / threading | 네이티브 goroutines |
| 타입 안전성 | 런타임 (mypy 선택) | 컴파일 타임 강제 |
| 크로스 플랫폼 | Python 런타임 필요 | 모든 플랫폼용 프리빌트 바이너리 |
| 훅 실행 | Shell 래퍼 + Python 인터프리터 | 컴파일된 바이너리, 직접 JSON 프로토콜 |

### 핵심 특징

- **48,688줄**의 Go 코드, 20개 패키지
- **85-100% 테스트 커버리지** (20개 테스트 패키지)
- **네이티브 동시성** via goroutines (병렬 LSP, 품질 검사, Git 작업)
- **내장된 템플릿** using `go:embed`
- **크로스 플랫폼** 빌드 (macOS arm64/amd64, Linux arm64/amd64, Windows)

---

## 1단계: Go 설치 확인

MoAI-ADK v2.0을 사용하려면 Go 1.22 이상이 필요합니다. 먼저 Go가 설치되어 있는지 확인하세요.

### Go 설치 확인

```bash
# Go 버전 확인
go version
```

**예상 출력:**
```
go version go1.23.0 darwin/arm64
```

### Go가 이미 설치된 경우

Go 1.22 이상이 설치되어 있다면 **[2단계: v1.x 제거하기](#2단계-v1x-제거하기)**로 건너뛰세요.

### Go가 설치되지 않은 경우

Go가 설치되지 않았거나 버전이 1.22 미만인 경우 아래 가이드를 따라 설치하세요:

#### macOS

```bash
# Homebrew로 설치 (추천)
brew install go

# 또는 공식 바이너리 다운로드
# https://go.dev/dl/
```

#### Linux

```bash
# 패키지 매니저로 설치
sudo apt install golang-go  # Ubuntu/Debian
sudo yum install golang      # CentOS/RHEL

# 또는 공식 바이너리 다운로드
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
```

#### Windows

```powershell
# winget으로 설치
winget install GoLang.Go

# 또는 공식 설치 프로그램
# https://go.dev/dl/
```

### 설치 후 PATH 설정

Go를 설치한 후 터미널을 다시 시작하거나 다음 명령어를 실행하세요:

```bash
# 터미널 재시작 또는 PATH 새로고침
source ~/.zshrc  # zsh 사용자
source ~/.bashrc # bash 사용자
```

---

## 2단계: v1.x 제거하기

Go 설치가 확인되었으면 기존 Python moai-adk를 제거하세요.

### 기존 Python moai-adk 제거

```bash
# uv로 설치한 경우
uv tool uninstall moai-adk

# pip로 설치한 경우
pip uninstall moai-adk

# 가상 환경 제거 (선택)
rm -rf ~/.local/share/uv/tools/moai-adk
```

---

## 3단계: MoAI-ADK v2.0 설치하기

### 설치 스크립트 (추천)

간단한 원라인 명령어로 설치하세요. OS와 아키텍처를 자동 감지하여 적합한 바이너리를 다운로드하고 설치합니다.

#### macOS, Linux, WSL

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash
```

또는:

```bash
wget -qO- https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash
```

#### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.ps1 | iex
```

또는:

```powershell
Invoke-Expression (Invoke-RestMethod -Uri "https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.ps1")
```

#### Windows CMD

```batch
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.bat -o install.bat
install.bat
```

curl이 없는 경우:
```batch
powershell -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.bat' -OutFile 'install.bat'"
install.bat
```

**특정 버전 설치:**
```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash -s -- --version 2.0.0
```

**사용자 정의 설치 경로:**
```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash -s -- --install-dir ~/bin
```

### 수동 설치 (GUI)

웹 기반 설치 페이지를 방문하여 설치할 수도 있습니다:

🔗 **[https://moai-adk.dev/install](https://moai-adk.dev/install)**

### 고급 사용자를 위한 수동 설치

```bash
# 플랫폼 감지 및 다운로드
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
  ARCH_NAME="darwin-arm64"
else
  ARCH_NAME="darwin-amd64"
fi

# 최신 릴리스 다운로드
echo "MoAI-ADK v2.0 설치 중 (macOS $ARCH)..."
curl -fsSL "https://github.com/modu-ai/moai-adk/releases/latest/download/moai-${ARCH_NAME}.tar.gz" -o moai.tar.gz

# 압축 해제 및 설치
tar -xzf moai.tar.gz
chmod +x moai
sudo mv moai /usr/local/bin/

# 정리
rm moai.tar.gz

# 설치 확인
echo ""
echo "✓ 설치 완료!"
moai version
```

#### Linux (ARM64 & AMD64)

```bash
# 플랫폼 감지 및 다운로드
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
  ARCH_NAME="linux-arm64"
else
  ARCH_NAME="linux-amd64"
fi

# 최신 릴리스 다운로드
echo "MoAI-ADK v2.0 설치 중 (Linux $ARCH)..."
curl -fsSL "https://github.com/modu-ai/moai-adk/releases/latest/download/moai-${ARCH_NAME}.tar.gz" -o moai.tar.gz

# 압축 해제 및 설치
tar -xzf moai.tar.gz
chmod +x moai
sudo mv moai /usr/local/bin/

# 정리
rm moai.tar.gz

# 설치 확인
echo ""
echo "✓ 설치 완료!"
moai version
```

#### Windows (PowerShell)

```powershell
# PowerShell 관리자 권한으로 실행
# 최신 릴리스 다운로드
Write-Host "MoAI-ADK v2.0 설치 중 (Windows)..."
Invoke-WebRequest -Uri "https://github.com/modu-ai/moai-adk/releases/latest/download/moai-windows-amd64.zip" -OutFile "moai.zip"

# 압축 해제
Expand-Archive -Path "moai.zip" -DestinationPath "." -Force

# 설치 디렉토리 생성 및 복사
$installDir = "$env:LOCALAPPDATA\MoAI-ADK"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Copy-Item -Path "moai.exe" -Destination "$installDir\" -Force

# PATH에 추가 (현재 세션만)
$env:PATH += ";$installDir"

# 정리
Remove-Item "moai.zip"
Remove-Item "moai.exe" -ErrorAction SilentlyContinue

# 설치 확인
Write-Host ""
Write-Host "✓ 설치 완료!"
& "$installDir\moai.exe" version
```

**Windows PATH 영구 추가 (선택):**

```powershell
# 시스템 PATH에 영구 추가 (관리자 권한 필요)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$installDir", "Machine")
```

### 방법 2: 소스에서 빌드 (개발용)

```bash
# 저장소 복론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 빌드
make build

# GOPATH에 설치
make install
```

### 설치 확인

```bash
moai version
moai doctor
```

---

## 1.x와 2.0의 주요 차이점

### 1. CLI 명령어 변경

| 1.x (Python) | 2.0 (Go) | 설명 |
|--------------|----------|------|
| `moai init` | `moai init` | 동일 (대화형 프로젝트 설정) |
| `moai doctor` | `moai doctor` | 동일 (시스템 진단) |
| `moai status` | `moai status` | 동일 (프로젝트 상태) |
| `moai version` | `moai version` | 동일 (버전 정보) |
| - | `moai worktree new <name>` | Git worktree 관리 추가 |
| - | `moai worktree list` | Worktree 목록 |
| - | `moai worktree switch <name>` | Worktree 전환 |
| - | `moai worktree sync` | Worktree 동기화 |
| - | `moai worktree remove <name>` | Worktree 제거 |
| - | `moai worktree clean` | 정체된 Worktree 정리 |
| - | `moai update` | 자체 업데이트 (롤백 지원) |
| - | `moai hook <event>` | Hook 디스패처 |

### 2. 성능 향상

| 항목 | 1.x (Python) | 2.0 (Go) | 개선폭 |
|------|--------------|----------|--------|
| 시작 시간 | ~800ms | ~5ms | **160배 더 빠름** |
| 메모리 사용 | ~100MB (기본) | ~15MB | **6배 더 적음** |
| 동시 작업 처리 | asyncio 오버헤드 | goroutines | **네이티브 병렬 처리** |

### 3. 훅 시스템 변경

**1.x (Python):**
- Shell 래퍼 스크립트 사용
- Python 인터프리터 필요
- 상대적으로 느린 실행

**2.x (Go):**
- 컴파일된 바이너리 훅
- 직접 JSON 프로토콜 통신
- 6가지 이벤트 타입 지원 (PreToolUse, PostToolUse, SessionStart, SessionEnd, PreCompact, Notification)
- 모든 훅 출력은 프로토콜 준수를 위해 `hookSpecificOutput`에 `hookEventName` 필드 포함

### 4. 새로운 기능

**Git Worktree 관리**
```bash
# 병렬 브랜치 개발을 위한 worktree 관리
moai worktree new feature/auth
moai worktree list
moai worktree switch feature/auth
moai worktree sync
moai worktree clean
```

**자체 업데이트**
```bash
# 최신 버전으로 업데이트 (롤백 지원)
moai update

# 업데이트 확인만
moai update --check

# 프로젝트 템플릿만 동기화
moai update --project
```

**LSP 통합**
- 16개 이상의 언어 지원
- 병렬 LSP 서버 관리
- TRUST 5 품질 게이트와 통합

### 5. 설정 파일 호환성

v2.0은 v1.x 설정 파일과 호환됩니다:
- `.claude/settings.json` - 그대로 사용 가능
- `.moai/config/` - 대부분 호환
- `CLAUDE.md` - 그대로 사용 가능

**새로운 설정 섹션:**
- `system.yaml` - Go 바이너리 관련 설정
- `workflow.yaml` - 워크플로우 설정

---

## 다가오는 기능: Agent Teams 모드

MoAI-ADK v2.0은 곧 **Agent Teams 모드**를 공개할 예정입니다.

### Agent Teams란?

Claude Code의 공식 [Agent Teams](https://code.claude.com/docs/en/agent-teams) 기능과 통합하여, 여러 Claude Code 세션이 팀으로 협력하여 복잡한 작업을 처리할 수 있습니다.

### 주요 사용 사례

- **연구 및 리뷰**: 여러 팀원이 문제의 다른 측면을 동시에 조사
- **새 모듈 또는 기능**: 각 팀원이 별도의 부분을 담당
- **경쟁 가설 테스트**: 여러 팀원이 병렬로 다른 이론을 테스트
- **계층간 조정**: 프론트엔드, 백엔드, 테스트를 각각 다른 팀원이 담당

### 하위 에이전트 vs Agent Teams

| 항목 | Subagents | Agent Teams |
|------|-----------|-------------|
| 컨텍스트 | 자체 컨텍스트, 결과만 반환 | 자체 컨텍스트, 완전히 독립적 |
| 통신 | 메인 에이전트에게만 보고 | 팀원 간 직접 메시지 |
| 조정 | 메인 에이전트가 모든 작업 관리 | 공유 작업 목록, 자체 조정 |
| 토큰 비용 | 낮음 (요약된 결과만) | 높음 (각 팀원이 별도 인스턴스) |
| 최적 용도 | 결과만 중요한 집중 작업 | 토론과 협력이 필요한 복잡한 작업 |

### Agent Teams 활성화

```bash
# 환경 변수 설정
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1

# 또는 settings.json에 추가
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

### MoAI Agent Teams 모드 출시 예정

MoAI-ADK v2.0의 Agent Teams 모드는 현재 개발 중이며, 다음 기능을 포함할 예정입니다:

- **자동 팀 구성**: 작업 유형에 따라 최적의 팀 구성 자동 제안
- **MoAI 워크플로우 통합**: `/moai plan`, `/moai run`, `/moai sync`와 팀 모드 통합
- **고급 조정 도구**: 팀 간 작업 분배, 진행 추적, 결과 종합
- **SPEC 기반 팀 작업**: SPEC 문서를 통한 팀 간 작업 조율

출시 일정은 곧 공개될 예정입니다.

---

## 빠른 시작 체크리스트

- [ ] Go 설치 확인 (`go version`)
- [ ] Go 1.22+ 설치 (필요시)
- [ ] 기존 Python moai-adk 제거 (`uv tool uninstall moai-adk`)
- [ ] MoAI-ADK v2.0 설치 (소스 빌드 또는 프리빌트 바이너리)
- [ ] `moai doctor`로 설치 확인
- [ ] 기존 프로젝트에서 `moai init` 실행 (설정 마이그레이션)
- [ ] `moai worktree` 기능으로 병렬 개발 시작

---

## 문제 해결

### Python 버전 잔존 파일

```bash
# 잔존 파일 제거
rm -rf ~/.local/share/uv/tools/moai-adk
rm -rf ~/.local/bin/moai  # Python 버전 심볼릭 링크
```

### Go 버전 확인

```bash
# Go 1.22+ 필요
go version
```

### PATH 설정

```bash
# Go bin이 PATH에 있는지 확인
which moai
# 출력: /usr/local/bin/moai 또는 $GOPATH/bin/moai

# PATH에 추가 (필요시)
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```

---

## 추가 정보

- [GitHub 저장소](https://github.com/modu-ai/moai-adk)
- [Claude Code 공식 문서](https://code.claude.com/docs/en/agent-teams)
- [Python 1.x 버전](https://github.com/modu-ai/moai-adk/tree/python) (레거시)

---

버전: 2.0.0
최종 업데이트: 2026-02-06
