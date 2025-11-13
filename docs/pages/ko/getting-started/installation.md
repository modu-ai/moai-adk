---
title: "설치 가이드"
description: "MoAI-ADK 설치 및 초기 설정 가이드 - 시스템 요구사항, 설치 방법, 환경 설정, 문제 해결"
---

# 설치 가이드

MoAI-ADK를 설치하고 개발 환경을 구성하는 방법을 안내합니다.

## 시스템 요구사항

### 필수 요구사항

- **Python**: 3.8+ (권장: 3.11+)
- **Node.js**: 18+ (Claude Code 및 관련 도구용)
- **Git**: 2.30+
- **운영체제**: macOS, Linux, Windows (WSL2)

### 권장 사양

- **메모리**: 8GB+ (16GB 권장)
- **저장 공간**: 10GB+ 여유 공간
- **CPU**: 4코어 이상

### 사전 설치 도구

```bash
# macOS (Homebrew)
brew install python node git

# Ubuntu/Debian
sudo apt update
sudo apt install python3 python3-pip nodejs npm git

# Windows (WSL2)
# WSL2에 Ubuntu를 설치하고 위 Ubuntu 명령어 실행
```

## 설치 방법

### 방법 1: uv tool 사용 (권장)

uv는 빠르고 효율적인 Python 패키지 관리자입니다.

```bash
# 1. uv 설치 (이미 설치되지 않은 경우)
curl -LsSf https://astral.sh/uv/install.sh | sh

# 2. MoAI-ADK 전역 설치
uv tool install moai-adk

# 3. 설치 확인
moai-adk --version
```

### 방법 2: pip 설치

```bash
# 1. pip를 최신 버전으로 업그레이드
python -m pip install --upgrade pip

# 2. MoAI-ADK 설치
pip install moai-adk

# 3. 설치 확인
python -m moai_adk --version
```

### 방법 3: 개발 버전 설치

최신 기능을 시험하려면 개발 버전을 설치하세요.

```bash
# 1. 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 2. 개발 모드로 설치
uv tool install -e .  # 또는 pip install -e .

# 3. 설치 확인
moai-adk --version
```

## 프로젝트 생성 및 초기화

### 새 프로젝트 생성

```bash
# 1. 새 프로젝트 생성
moai-adk init my-awesome-project

# 2. 프로젝트 디렉토리로 이동
cd my-awesome-project

# 3. Claude Code 실행 및 프로젝트 설정
claude-code .
# Claude Code에서 다음 명령 실행: /alfred:0-project
```

### 기존 프로젝트에 MoAI-ADK 추가

```bash
# 1. 기존 프로젝트 디렉토리로 이동
cd existing-project

# 2. MoAI-ADK 초기화
moai-adk init .

# 3. Claude Code 실행 및 프로젝트 설정 최적화
claude-code .
# Claude Code에서 다음 명령 실행: /alfred:0-project
```

## Claude Code 설정

### 1. Claude Code 설치

```bash
# macOS
brew install claude-code

# 다른 플랫폼
# https://claude.ai/download 에서 설치
```

### 2. 프로젝트 열기

```bash
# Claude Code로 프로젝트 열기
claude-code my-awesome-project
```

### 3. 프로젝트 설정

Claude Code에서 `/alfred:0-project` 명령을 실행하여 프로젝트를 설정합니다.

```bash
/alfred:0-project
```

이 명령은 다음을 자동으로 구성합니다:
- 프로젝트 메타데이터
- 개발 언어 감지
- Git 전략 설정
- Alfred 슈퍼에이전트 초기화
- 다국어 시스템 설정

## 환경 설정 확인

### 설치 검증

```bash
# 1. MoAI-ADK 버전 확인
moai-adk --version

# 2. 프로젝트 상태 확인
moai-adk status

# 3. 설정 파일 확인
ls -la .moai/config.json
```

### Claude Code 상태줄

Claude Code 터미널 하단에 상태줄이 표시되어야 합니다:

```
🤖 Haiku 4.5 | 🗿 Ver 0.23.0 | 📊 Git: main | Changes: +0 M0 ?0
```

## 문제 해결

### 일반적인 문제들

#### 1. "moai-adk: command not found" 오류

**원인**: PATH에 moai-adk가 추가되지 않음

**해결책**:
```bash
# uv tool 설치 경로 확인
uv tool list

# PATH에 추가 (예: zsh)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

#### 2. "Python 3.8+ required" 오류

**원인**: Python 버전이 너무 낮음

**해결책**:
```bash
# Python 버전 확인
python --version

# 최신 Python 설치 (macOS)
brew install python@3.11

# 우선순위 설정
echo 'export PATH="$(brew --prefix python@3.11)/bin:$PATH"' >> ~/.zshrc
```

#### 3. Claude Code에서 Alfred 명령어 작동하지 않음

**원인**: 프로젝트가 초기화되지 않음

**해결책**:
```bash
# 1. 프로젝트 초기화 확인
ls -la .moai/

# 2. 없다면 초기화 실행
/alfred:0-project

# 3. 설정 파일 확인
cat .moai/config.json
```

#### 4. Git 관련 오류

**원인**: Git이 설치되지 않거나 설정되지 않음

**해결책**:
```bash
# Git 설치 확인
git --version

# Git 설정
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# SSH 키 설정 (선택사항)
ssh-keygen -t ed25519 -C "your.email@example.com"
```

### 디버깅 모드

문제가 계속되면 디버깅 모드로 실행하세요:

```bash
# 상세 정보 출력
moai-adk --verbose status

# 디버그 모드
MOAI_DEBUG=1 moai-adk status

# 설정 파일 진단
moai-adk doctor
```

### 로그 확인

```bash
# MoAI-ADK 로그 위치
ls -la ~/.moai/logs/

# Claude Code 로그
ls -la ~/.claude/logs/
```

## v0.23.1 최신 기능

### Skills Ecosystem v4.0
- **292 Skills 지원** (기존 55개에서 5배 확장)
- **12 BaaS 플랫폼 통합** (Supabase, Firebase, Vercel, Cloudflare, Auth0, Convex, Railway, Neon, Clerk 등)
- **95%+ 검증 성공률** 달성
- [Skills 전체 목록 보기](/ko/skills/ecosystem-upgrade-v4)

### Expert Delegation System v2.0
- **4단계 자동 전문가 할당** 시스템
- **60% 사용자 상호작용 감소** 달성
- **95%+ 정확도** 유지
- [Expert Delegation System 자세히 보기](/ko/alfred/expert-delegation-system)

### Senior Engineer Thinking
- **8가지 연구 전략** 통합 (v0.22.0+)
- **병렬 연구 작업** 시스템
- **학습 및 복리 효과**
- [Senior Engineer Thinking 자세히 보기](/ko/features/senior-engineer-thinking)

### 설치 후 확인

```bash
# MoAI-ADK 버전 확인 (v0.23.1+)
moai-adk --version

# Skills 목록 확인 (292 Skills)
moai-adk skills list

# 프로젝트 초기화
moai-adk init my-project
cd my-project

# ⚠️ 필수: 프로젝트 설정
/alfred:0-project
```

## 다음 단계

설치가 완료되었습니다! 다음 단계를 진행하세요:

1. **[5분 빠른 시작](./quick-start)**: 첫 프로젝트 즉시 실행
2. **[실전 튜토리얼](/ko/tutorials)**: 단계별 학습 (REST API, JWT 인증, DB 최적화)
3. **[코드 예제 라이브러리](/ko/examples)**: 즉시 사용 가능한 예제
4. **[BaaS 생태계 가이드](/ko/skills/baas-ecosystem)**: 12개 플랫폼 완전 가이드

## 추가 리소스

- **[GitHub 저장소](https://github.com/modu-ai/moai-adk)**: 소스 코드 및 이슈
- **[문제 해결](../troubleshooting)**: 더 많은 해결책
- **[커뮤니티](https://github.com/modu-ai/moai-adk/discussions)**: 도움 및 토론

---

**도움이 필요하신가요?**
- 📧 이메일: <support@mo.ai.kr>
- 💬 GitHub Discussions: [질문하기](https://github.com/modu-ai/moai-adk/discussions)
- 🐛 버그 보고: [Issues](https://github.com/modu-ai/moai-adk/issues)