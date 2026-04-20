---
title: Windows 사용 가이드
weight: 40
draft: false
---
# Windows 사용 가이드

## 지원 환경

| 환경 | 지원 여부 | 비고 |
|------|----------|------|
| **WSL (권장)** | ✅ 완전 지원 | 최적의 경험 |
| **PowerShell 7.x+** | ✅ 지원 | 대안 환경 |
| PowerShell 5.x (레거시) | ❌ 미지원 | Windows PowerShell |
| cmd.exe | ❌ 미지원 | 명령 프롬프트 |

**필수 요구사항:**
- [Git for Windows](https://gitforwindows.org/) 설치 필수
- WSL 또는 PowerShell 7.x 이상

## 설치 방법

### WSL (권장)

WSL은 Windows에서 Linux 환경을 제공하며, MoAI-ADK의 모든 기능을 완벽하게 지원합니다.

```bash
# WSL 설치 (관리자 PowerShell에서 실행)
wsl --install

# WSL 내에서 MoAI-ADK 설치
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh \
  | bash
```

### PowerShell 7.x+

> **참고**: 최적의 경험을 위해 WSL 사용을 권장합니다.

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.ps1 | iex
```

## 한글 사용자명 경로 에러

### 문제 현상

Windows 사용자명에 한글, 중국어 등 비-ASCII 문자가 포함된 경우, `EINVAL` 에러가 발생할 수 있습니다. 이는 Windows의 8.3 짧은 파일명 변환 과정에서 발생하는 문제입니다.

```
Error: EINVAL: invalid argument, open 'C:\Users\홍길동\AppData\Local\Temp\...'
```

### 해결 방법 1: 대체 임시 디렉터리 설정 (권장)

ASCII 문자만 포함된 경로에 임시 디렉터리를 생성합니다:

```bash
# Command Prompt
set MOAI_TEMP_DIR=C:\temp
mkdir C:\temp 2>/dev/null
```

```powershell
# PowerShell
$env:MOAI_TEMP_DIR="C:\temp"
New-Item -ItemType Directory -Path "C:\temp" -Force
```

환경 변수를 영구적으로 설정하려면 시스템 환경 변수에 `MOAI_TEMP_DIR`을 추가하세요.

### 해결 방법 2: 8.3 파일명 생성 비활성화

관리자 권한으로 실행:

```bash
fsutil 8dot3name set 1
```

> **주의**: 이 설정은 시스템 전체에 영향을 미칩니다. 일부 레거시 프로그램이 영향을 받을 수 있습니다.

### 해결 방법 3: ASCII 사용자 계정 생성

영어 이름으로 새 Windows 사용자 계정을 생성하면 경로 문제를 근본적으로 해결합니다.

## WSL 설정 가이드

### WSL 설치

```powershell
# 관리자 PowerShell에서 실행
wsl --install

# 기본 배포판: Ubuntu (권장)
# 재시작 후 사용자명 및 비밀번호 설정
```

### 프로젝트 파일 접근

WSL에서 Windows 파일에 접근:

```bash
# Windows 파일시스템 접근
cd /mnt/c/Users/사용자명/projects/

# WSL 네이티브 파일시스템 사용 (더 빠름)
cd ~/projects/
```

> **성능 팁**: WSL 네이티브 파일시스템(`~/` 하위)에서 작업하면 크로스 파일시스템 오버헤드 없이 최적의 성능을 얻을 수 있습니다.

### VS Code 연동

1. VS Code에 [WSL 확장](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl) 설치
2. WSL 터미널에서 `code .` 실행
3. VS Code가 자동으로 WSL 모드로 열림

## CG 모드에서의 tmux 사용

[CG 모드](/ko/multi-llm/cg-mode)를 사용하려면 tmux가 필요합니다. WSL에서 설치:

```bash
# Ubuntu/Debian
sudo apt install tmux

# tmux 세션 시작
tmux new -s moai

# CG 모드 실행
moai cg
```

## 문제 해결

| 문제 | 원인 | 해결 |
|------|------|------|
| `moai: command not found` | PATH에 Go bin 디렉터리 미포함 | `export PATH="$HOME/go/bin:$PATH"`를 `.bashrc`에 추가 |
| `EINVAL` 에러 | 한글 사용자명 | 위의 [한글 사용자명 경로 에러](#한글-사용자명-경로-에러) 참조 |
| 권한 거부 | 설치 스크립트 권한 | `chmod +x install.sh` 후 재실행 |
| Git 명령 실패 | Git for Windows 미설치 | [Git for Windows](https://gitforwindows.org/) 설치 |
| tmux 없음 | CG 모드 실행 불가 | `sudo apt install tmux` (WSL에서) |

## 다음 단계

- [설치](/ko/getting-started/installation) — 설치 상세 가이드
- [초기 설정](/ko/getting-started/init-wizard) — 프로젝트 초기화
- [CG 모드](/ko/multi-llm/cg-mode) — Claude + GLM 하이브리드 모드
