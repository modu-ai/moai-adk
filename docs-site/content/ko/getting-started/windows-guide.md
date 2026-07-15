---
title: Windows 사용 가이드
weight: 40
draft: false
---

Windows에서 MoAI-ADK를 사용할 때 알아야 할 환경 요구사항과 흔한 함정을 정리했습니다. 결론부터 말하면 **WSL이 가장 편합니다**. 네이티브 Windows 환경에서 겪는 경로·권한 문제 대부분이 WSL에서는 발생하지 않습니다.

## 지원 환경

| 환경 | 지원 여부 | 비고 |
|------|----------|------|
| **WSL (권장)** | {{< icon check ok >}} 완전 지원 | 최적의 경험 |
| **PowerShell 7.x+** | {{< icon check ok >}} 지원 | 대안 환경 |
| PowerShell 5.x (레거시) | {{< icon x danger >}} 미지원 | Windows PowerShell |
| cmd.exe | {{< icon x danger >}} 미지원 | 명령 프롬프트 |

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

Windows 사용자명에 한글, 중국어 등 비-ASCII 문자가 포함된 경우 일부 레거시 도구나 8.3 짧은 파일명 변환 과정에서 경로 처리 문제가 발생할 수 있습니다. 홈 디렉터리 경로에 비-ASCII 문자가 섞여 있으면 특정 명령이 실패할 수 있습니다.

```
C:\Users\홍길동\...
```

이 경우 아래 방법으로 ASCII 전용 경로 환경을 마련하는 것이 가장 확실합니다.

### 해결 방법 1: 8.3 파일명 생성 활성화

8.3 짧은 파일명(ASCII 대체 경로)이 생성되도록 관리자 권한으로 설정합니다.

```powershell
fsutil 8dot3name set 1
```

> **주의**: 이 설정은 시스템 전체에 영향을 미칩니다. 일부 레거시 프로그램이 영향을 받을 수 있습니다.

### 해결 방법 2: ASCII 사용자 계정 생성

영어 이름으로 새 Windows 사용자 계정을 생성하면 홈 디렉터리 경로 문제를 근본적으로 해결합니다.

### 해결 방법 3: WSL 사용

가장 권장하는 방법은 WSL(아래 [WSL 설정 가이드](#wsl-설정-가이드) 참조) 환경에서 작업하는 것입니다. WSL 네이티브 파일시스템은 비-ASCII 홈 경로 문제의 영향을 받지 않습니다.

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
| `moai: command not found` | PATH에 설치 디렉터리 미포함 | 설치 스크립트는 `~/.local/bin`에 설치 — `export PATH="$HOME/.local/bin:$PATH"`를 `.bashrc`에 추가 (`go install`로 설치한 경우 `$HOME/go/bin`) |
| 한글 경로 처리 실패 | 한글 사용자명 | 위의 [한글 사용자명 경로 에러](#한글-사용자명-경로-에러) 참조 |
| 권한 거부 | 설치 스크립트 권한 | `chmod +x install.sh` 후 재실행 |
| Git 명령 실패 | Git for Windows 미설치 | [Git for Windows](https://gitforwindows.org/) 설치 |
| tmux 없음 | CG 모드 실행 불가 | `sudo apt install tmux` (WSL에서) |

## 다음 단계

- [설치](/ko/getting-started/installation) — 설치 상세 가이드
- [초기 설정](/ko/getting-started/init-wizard) — 프로젝트 초기화
- [CG 모드](/ko/multi-llm/cg-mode) — Claude + GLM 하이브리드 모드
