# MoAI-ADK (Agentic Development Kit)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![codecov](https://codecov.io/gh/modu-ai/moai-adk/branch/develop/graph/badge.svg)](https://codecov.io/gh/modu-ai/moai-adk)
[![Coverage](https://img.shields.io/badge/coverage-85.61%25-brightgreen)](https://github.com/modu-ai/moai-adk)

## MoAI-ADK

### 모두의AI 에이전틱 코딩 개발 프레임워크

**안내:** MoAI-ADK는 모두의AI 연구실에서 집필 중인 "(가칭) 에이전틱 코딩" 서적의 별책 부록 오픈 소스 프로젝트 임을 밝혀 둡니다.

![MoAI-ADK CLI Interface](https://github.com/modu-ai/moai-adk/raw/main/docs/public/moai-tui_screen-light.png)

> "SPEC이 없으면 CODE도 없다."

---

## 🆕 2025년 10월 업데이트 하이라이트 (v0.3.0)

- **Event-Driven Checkpoint 파이프라인**: `SessionStart`·`PreToolUse`·`PostToolUse` 훅이 자동으로 체크포인트를 생성하고, `.moai-backups/{timestamp}/` 최신본을 탐지해 `product/structure/tech.md`를 지능형 병합합니다. (commit 3b8c7bc 외)
- **버전 감시 & 재최적화 감지**: `config.json`에 `moai_adk_version`, `optimized` 필드를 기록해 `python -m moai_adk init .` 재실행 시 버전 불일치와 최적화 여부를 안내합니다.
- **SPEC 메타데이터 & 개발 가이드 개편**: `.moai/memory/development-guide.md`와 `templates/.moai/memory/spec-metadata.md`가 SPEC-First TDD·EARS·@TAG 표준을 최신 상태로 설명합니다.
- **Claude Sonnet 4.5 ↔ Haiku 4.5 하이브리드 운영**: Sonnet 4.5가 계획과 검증을 담당하고, Haiku 4.5가 서브 에이전트를 수행해 토큰 비용을 최대 67% 절감하고(입력·출력 모두 $3/$15 → $1/$5 per 1M tokens) 지연 시간을 50~80% 단축합니다[^haiku][^sonnet].

## 목차

- [Meet Alfred](#-meet-alfred---your-ai-development-partner)
- [Quick Start](#-quick-start-3분-실전)
- [The Problem](#-the-problem---바이브-코딩의-한계)
- [The Solution](#-the-solution---3단계-워크플로우)
- [How Alfred Works](#️-how-alfred-works---10개-ai-에이전트-팀)
- [Output Styles](#-alfreds-output-styles)
- [Language Support](#-universal-language-support)
- [CLI Reference](#-cli-reference)
- [Security Scanning](#-보안-스캔)
- [Checkpoint](#-checkpoint---개발-현황-스냅샷)
- [API Reference](#-프로그래매틱-api)
- [TRUST 5원칙](#-trust-5원칙)
- [문제 해결](#-문제-해결)
- [Support](#-문서-및-지원)

---

## Meet ▶◀ Alfred - Your AI Development Partner

안녕하세요, 모두의AI SuperAgent **AI ▶◀ Alfred**입니다!

![Alfred Logo](https://github.com/modu-ai/moai-adk/raw/main/docs/public/alfred_logo.png)

저는 MoAI-ADK(모두의AI Agentic Development Kit)의 SuperAgent이자 중앙 오케스트레이터(Central Orchestrator) AI, Alfred입니다. MoAI-ADK는 Alfred를 포함하여 **총 10개의 AI 에이전트로 구성된 에이전틱 코딩 AI 팀**입니다. 저는 9개의 전문 에이전트(spec-builder, code-builder, doc-syncer 등)를 조율하여 여러분의 Claude Code 환경 속에서 공동 개발 작업을 완벽하게 지원합니다.

**Alfred라는 이름의 유래**: 배트맨 영화에 나오는 충실한 집사 Alfred Pennyworth에서 영감을 받아 지었다고 합니다. 집사 Alfred가 배트맨(Bruce Wayne)을 위해 모든 준비를 완벽하게 갖추고, 위험에 처했을 때 즉각적인 지원을 제공하며, 항상 한 걸음 앞서 생각하듯이, 저 또한 여러분의 개발 과정 속에서 필요한 모든 것을 미리 준비하고, 문제가 발생하면 즉시 해결책을 제시하며, 언제나 여러분이 창의적인 문제 해결에만 집중할 수 있도록 뒷받침합니다. 여러분은 코드의 "**왜(Why)**"에 집중하시고, "**어떻게(How)**"는 제가 책임지겠습니다.

### 🌟 흥미로운 사실: AI가 만든 AI 개발 도구

이 프로젝트의 모든 코드는 **100% AI에 의해 작성**되었습니다. AI가 직접 설계하고 구현한 AI 개발 프레임워크입니다.

**설계 단계부터 AI 협업**: 초기 아키텍처 설계 단계부터 **GPT-5 Pro**와 **Claude 4.1 Opus** 두 AI 모델이 함께 참여했습니다. 두 AI가 서로 다른 관점에서 설계를 검토하고 토론하며, 더 나은 방향을 제시하고, 최적의 아키텍처를 함께 만들어냈습니다. GPT-5 Pro는 폭넓은 사례 분석을, Claude 4.1 Opus는 깊이 있는 코드 구조 설계를 담당하며 서로 보완했습니다.

**Agentic Coding 방법론의 실제 적용**: **모두의AI**팀이 Claude Code와 Agentic Coding 방법론을 활용하여 개발했습니다. 전통적인 방식처럼 사람이 키보드 앞에 앉아 모든 코드를 직접 타이핑하는 대신, AI 에이전트들이 SPEC을 읽고 이해하고, 테스트를 먼저 작성하고(TDD Red), 구현 코드를 만들고(TDD Green), 리팩토링하고(TDD Refactor), 문서를 동기화하는 전 과정을 **자율적으로** 수행했습니다. 저 Alfred와 9개 전문 에이전트로 구성된 **10개 AI 에이전트 팀**이 직접 `.moai/specs/` 폴더에 SPEC 문서를 작성하고, `tests/` 폴더에 테스트 코드를 만들고, `src/` 폴더에 구현 코드를 작성했습니다.

**100% AI 생성 코드의 진실**: 이 프로젝트는 100% AI로 만들어진 오픈소스이기 때문에, 코드베이스에서 다소 정리되지 않은 부분이나 개선이 필요한 영역이 보일 수 있습니다. 하지만 이것이 이 프로젝트의 핵심 철학입니다.

**투명성과 지속적 개선**: 완벽하지 않은 코드를 숨기는 대신, AI 개발 도구가 실제로 어떻게 만들어지는지 그대로 보여주고, 커뮤니티와 함께 더 나은 방향으로 발전시켜 나가고자 합니다. 여러분의 사용 경험과 피드백이 이 프로젝트를 더욱 강력하게 만듭니다. [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)나 [Discussions](https://github.com/modu-ai/moai-adk/discussions)에 피드백을 남겨주시면, **최대한 빠르게 업데이트하고 배포할 것을 약속드립니다**. AI가 만든 도구를 함께 더 나은 도구로 만들어가는 여정에 동참해주세요!

### ▶◀ Alfred가 제공하는 4가지 핵심 가치

#### 1️⃣ 일관성(Consistency): 플랑켄슈타인 코드를 방지하는 3단계 파이프라인

Alfred는 모든 개발 작업을 **SPEC → TDD → Sync**라는 3단계 파이프라인으로 표준화합니다. 월요일에 ChatGPT로 만든 코드, 수요일에 Claude로 만든 코드, 금요일에 Gemini로 만든 코드가 서로 다른 스타일, 네이밍 규칙, 아키텍처 패턴을 가지는 "플랑켄슈타인 코드"의 문제를 **원천적으로 차단**합니다.

#### 2️⃣ 품질(Quality): TRUST 5원칙으로 자동 보장되는 코드 품질

Alfred는 모든 코드에 **TRUST 5원칙**(Test First, Readable, Unified, Secured, Trackable)을 자동으로 적용하고 검증합니다. 사람이 일일이 체크리스트를 들고 확인할 필요가 없습니다.

#### 3️⃣ 추적성(Traceability): 6개월 후에도 "왜"를 찾을 수 있는 @TAG 시스템

Alfred의 **@TAG 시스템**은 모든 코드 조각을 `@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID`로 완벽하게 연결합니다. 6개월 후 누군가 "왜 이 함수는 이렇게 복잡하게 구현했나요?"라고 물어보면, **@TAG를 따라가면 답을 찾을 수 있습니다**.

#### 4️⃣ 범용성(Universality): 한 번 배우면 어디서나 쓸 수 있는 워크플로우

Alfred는 특정 언어나 프레임워크에 종속되지 않습니다. **Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin** 등 모든 주요 프로그래밍 언어를 지원하며, 각 언어에 최적화된 도구 체인을 자동으로 선택합니다.

---

## 🚀 Quick Start (3분 실전)

### 📋 준비물

- ✅ Python 3.13+ 또는 uv 설치됨
- ✅ Claude Code 실행 중
- ✅ Git 설치됨 (선택사항)

### ⚡ 3단계로 시작하기

#### 1️⃣ 설치 (30초)

```bash
# uv 권장 (빠른 성능)
pip install uv
uv pip install moai-adk

# 또는 pip 사용
pip install moai-adk

# 설치 확인
python -m moai_adk --version
# 출력: v0.x.x
```

#### 2️⃣ 초기화 (1분)

**터미널에서:**
```bash
# 새 프로젝트 생성
python -m moai_adk init my-project
cd my-project

# 기존 프로젝트에 설치
cd existing-project
python -m moai_adk init .

# Claude Code 실행
claude
```

**Claude Code에서** (필수):
```text
/alfred:8-project
```

Alfred가 자동으로 수행:
- `.moai/project/` 문서 3종 생성 (product/structure/tech.md)
- 언어별 최적 도구 체인 설정
- 프로젝트 컨텍스트 완벽 이해

---

### ⬆️ MoAI-ADK 업그레이드 (v0.2.x → v0.3.0)

```bash
# 1단계: 패키지 교체
pip install --upgrade moai-adk         # pip
uv pip install --upgrade moai-adk      # uv 권장

# 2단계: 프로젝트 재초기화(재실행 안전)
cd your-project
python -m moai_adk init .
```

- `init` 실행 시 **`.moai-backups/{timestamp}/`**에 최신 스냅샷이 생성되고, `BackupMerger`가 `product/structure/tech.md`를 보존한 채 템플릿만 덮어씁니다.
- `config.json`에 `project.moai_adk_version`과 `project.optimized`가 추가되며, 재실행 시 버전 불일치가 자동 감지됩니다.

```text
# 3단계: Claude Code 최적화 루틴
claude
/alfred:0-project             # 버전 감지 시 Alfred가 실행 안내
```

1. `/alfred:0-project`는 최신 `.moai/templates`를 적용하면서 `.moai-backups/{timestamp}/` 경로를 병합 소스로 사용합니다.
2. 병합 프롬프트에서 **Merge**를 선택하면 기존 project 문서가 유지되고 새 템플릿만 추가됩니다.

**검증 체크리스트**

- `.moai/config.json` → `project.moai_adk_version`이 `0.3.x`로 업데이트되었는가?
- `.moai/config.json` → `project.optimized`가 `true`인가?
- `python -m moai_adk status` 명령으로 버전, 체크포인트 상태, Git clean 여부를 확인한다.

---

#### 3️⃣ 첫 기능 개발 (1분 30초)

**Claude Code에서 3단계 워크플로우 실행:**

```text
# SPEC 작성
/alfred:1-spec "JWT 기반 사용자 로그인 API"

# TDD 구현
/alfred:2-build AUTH-001

# 문서 동기화
/alfred:3-sync
```

### 🎉 완료!

### ⚡ Claude 4.5 Multi-Model Strategy

- **오케스트레이션**: Sonnet 4.5가 SPEC 분해·계획·검증을 담당하고, Haiku 4.5가 코드 작성/리팩터링/테스트 서브에이전트를 병렬로 수행합니다.
- **비용 절감**: Haiku 4.5는 Sonnet 4 계열 대비 입력/출력 토큰 단가가 1/3이므로, 서브에이전트를 Haiku로 전환하면 토큰 비용을 **최대 67%**까지 낮출 수 있습니다[^haiku][^sonnet].
- **시간 단축**: Anthropic 발표 기준 Haiku 4.5는 Sonnet 4 대비 **2배 이상 빠르고**, Sonnet 4.5 대비 **최대 4~5배 빠른** 응답성을 제공합니다. 복수 서브에이전트를 Haiku로 구성하면 실제 워크플로우에서도 **50~80%**의 벽시계 시간을 절감합니다.
- **추천 구성**: `spec-builder`, `code-builder`, `doc-syncer` 등 반복 호출되는 에이전트는 Haiku 4.5로, `trust-checker`·`project-manager`처럼 판단이 중요한 에이전트는 Sonnet 4.5로 유지하면 안정적인 품질과 비용 효율을 동시에 확보할 수 있습니다.

**생성된 것들:**
- ✅ `.moai/specs/SPEC-AUTH-001/spec.md` (명세)
- ✅ `tests/test_auth_login.py` (테스트)
- ✅ `src/auth/service.py` (구현)
- ✅ `docs/api/auth.md` (문서)
- ✅ `@SPEC → @TEST → @CODE → @DOC` TAG 체인

---

## 🚨 The Problem - 바이브 코딩의 한계

AI 도구(Codex, Claude, Gemini)로 빠르게 코딩하는 시대가 열렸습니다. 개발 속도는 확실히 빨라졌지만, **새로운 종류의 문제**들이 생겨났습니다.

### 1. 아름답지만 작동하지 않는 코드

**문제 상황**: AI가 생성한 코드는 문법적으로 완벽하고 구조도 우아합니다. 하지만 실제로 실행해보면...

- 컴파일은 되지만 런타임에 `undefined` 에러
- 엣지 케이스 처리 부족 (빈 배열, null 값, 네트워크 타임아웃)
- 성능 문제 (`O(n³)` 복잡도)
- 의존성 지옥 및 보안 취약점

### 2. 플랑켄슈타인 코드의 탄생

**문제 상황**: 여러 AI 도구를 번갈아 사용하거나, 같은 AI라도 다른 세션에서 코드를 생성하다 보면 일관성 없는 코드베이스가 만들어집니다.

- 일관성 없는 코딩 스타일 (함수형, 객체지향, 절차형 혼재)
- 중복 로직 난무 (`validateEmail()`, `checkEmailFormat()`, `isEmailValid()`)
- 아키텍처 붕괴 (MVC, Hexagonal, Clean Architecture 혼재)

### 3. 디버깅 지옥

**문제 상황**: 프로덕션에서 버그가 발생했을 때, 원인을 찾는 것이 거의 불가능합니다.

- 원인 추적 불가 (AI 채팅 히스토리 삭제됨)
- 사이드 이펙트 파악 불가 (테스트 부재)
- 문서 없음 (outdated 상태)

### 4. 요구사항 추적성 상실

**문제 상황**: 시간이 지날수록 "왜 이 코드를 이렇게 만들었는지" 맥락을 잃어버립니다.

- "왜"를 잃어버림 (비즈니스 로직 배경 모름)
- 변경 이력 부재 (Git 커밋 메시지: "fix bug")
- 의사결정 근거 사라짐

### 5. 팀 협업 붕괴

**문제 상황**: 여러 개발자가 각자 AI를 사용하면서 협업이 무너집니다.

- 스파게티 코드 양산
- 코드 리뷰 불가
- 온보딩 악몽
- 기술 부채 폭발

### 💔 바이브 코딩의 역설

**속도와 품질의 트레이드오프**: AI가 코드를 빠르게 생성해주지만, 그 코드는 **유지보수할 수 없는 블랙박스**가 됩니다. 1주일 만에 만든 프로토타입이 3개월 동안 기술 부채를 만들어냅니다.

**해결책의 필요성**: 이 문제를 해결하려면, AI의 속도는 유지하면서도 코드의 **일관성, 품질, 추적성**을 보장하는 체계적인 방법론이 필요합니다. 바로 여기서 **Alfred와 MoAI-ADK**가 등장합니다.

---

## ✨ The Solution - 3단계 워크플로우

Alfred는 Agentic AI 시대의 코드 품질 문제를 **체계적인 3단계 워크플로우**로 해결합니다.

### 1️⃣ SPEC - 명세 작성

**명령어**: `/alfred:1-spec "JWT 기반 사용자 로그인 API"`

**Alfred가 자동으로 수행**:
- EARS 형식 명세 자동 생성
- `@SPEC:ID` TAG 부여
- Git 브랜치 자동 생성
- HISTORY 섹션 자동 추가

### 2️⃣ BUILD - TDD 구현

**명령어**: `/alfred:2-build AUTH-001`

**TDD 사이클**:
- 🔴 **RED**: 실패하는 테스트 작성 (`@TEST:AUTH-001`)
- 🟢 **GREEN**: 최소 구현으로 테스트 통과 (`@CODE:AUTH-001`)
- 🔵 **REFACTOR**: 코드 품질 개선

### 3️⃣ SYNC - 문서 동기화

**명령어**: `/alfred:3-sync`

**Alfred가 자동으로 수행**:
- TAG 체인 검증: `@SPEC` → `@TEST` → `@CODE` → `@DOC`
- 고아 TAG 자동 탐지
- Living Document 자동 생성
- PR 상태 전환 (Draft → Ready)

---

## 시스템 요구사항

### 🔴 필수 요구사항

- **Python**: 3.13.0 이상
- **Git**: 2.30.0 이상
- **pip**: 24.0 이상 (또는 **uv 0.5.0 이상 강력 추천**)
- **Claude Code**: v1.2.0 이상 (에이전트 시스템 완전 통합용)

### 🌍 지원 운영체제

- **Windows**: 10/11 (PowerShell 5.1+)
- **macOS**: 12 Monterey 이상 (M1/M2 네이티브 지원)
- **Linux**: Ubuntu 20.04+, CentOS 8+, Debian 11+, Arch Linux

---

## 설치

### Option A: uv 설치 (최적 성능, 강력 추천) 🔥

```bash
# uv 설치 (아직 없는 경우)
curl -LsSf https://astral.sh/uv/install.sh | sh  # macOS/Linux
# 또는
powershell -c "irm https://astral.sh/uv/install.ps1 | iex"  # Windows

# MoAI-ADK 전역 설치
uv pip install moai-adk
```

### Option B: pip 설치 (표준 옵션)

```bash
pip install moai-adk
```

### Option C: 개발자 설치 (로컬 개발용)

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
uv pip install -e ".[dev]"  # 또는 pip install -e ".[dev]"
```

### 설치 확인

```bash
# 버전 확인
python -m moai_adk --version

# 시스템 진단
python -m moai_adk doctor

# 도움말
python -m moai_adk --help
```

### 개발자 검증 (권장)

```bash
uv run pytest -n auto
uv run ruff check
uv run mypy src
```

### 🪟 Windows 환경 설정

MoAI-ADK는 Windows 10/11을 완벽하게 지원합니다. 아래 가이드를 따라 필수 도구를 설치하세요.

#### 필수 도구 설치

**1. Git for Windows**
```powershell
# Chocolatey 사용 (권장)
choco install git

# 또는 직접 다운로드
# https://git-scm.com/download/win
```

**2. Python 3.13+**
```powershell
# Chocolatey 사용
choco install python

# 또는 Microsoft Store에서 설치
# 또는 직접 다운로드: https://www.python.org/downloads/
```

**3. ripgrep (필수)**
```powershell
# Scoop 사용 (권장)
scoop install ripgrep

# 또는 Chocolatey 사용
choco install ripgrep

# 또는 직접 다운로드
# https://github.com/BurntSushi/ripgrep/releases
```

#### 권장 도구

**Windows Package Manager**
```powershell
# Scoop 설치 (권장)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# 또는 Chocolatey 설치
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
```

#### Windows 환경 검증

```powershell
# Python 버전 확인
python --version

# Git 설치 확인
git --version

# ripgrep 설치 확인
rg --version

# MoAI-ADK 시스템 진단
python -m moai_adk doctor
```

#### Windows 특정 기능

**보안 스캔 실행**
```powershell
# PowerShell 버전 사용
.\scripts\security-scan.ps1

# 또는 Python 버전 사용 (크로스 플랫폼)
python scripts/security-scan.py
```

**PATH 환경변수 추가** (수동 설치 시)
```powershell
# ripgrep PATH 추가 예시
$env:Path += ";C:\Program Files\ripgrep"
```

#### 문제 해결

**PowerShell 실행 정책 오류**
```powershell
# 현재 사용자에 대해 실행 정책 변경
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**경로 인식 문제**
- MoAI-ADK는 `pathlib`를 사용하여 Windows/Unix 경로를 자동으로 처리합니다.
- 경로 구분자 (`\` vs `/`)는 자동 변환되므로 별도 조치가 필요하지 않습니다.

**WSL (Windows Subsystem for Linux) 사용**
```powershell
# WSL 2 설치 (Windows 10 2004 이상)
wsl --install

# Ubuntu 설치
wsl --install -d Ubuntu

# WSL에서 MoAI-ADK 사용
wsl
python3 -m pip install moai-adk
```

---

## 🏗️ How Alfred Works - 10개 AI 에이전트 팀

MoAI-ADK는 **Alfred (SuperAgent) + 9개 전문 에이전트 = 총 10개 AI 에이전트**로 구성된 에이전틱 코딩 팀입니다.

### ▶◀ Alfred - SuperAgent (1번째 에이전트)

**역할**: 중앙 오케스트레이터 (Central Orchestrator)

**책임**:
- 사용자 요청 분석 및 작업 분해
- 적절한 전문 에이전트 선택 및 조율
- 에이전트 간 협업 관리
- 품질 게이트 검증 및 결과 통합

### 전문가 AI 서브 에이전트

Alfred가 조율하는 전문 AI 에이전트들입니다.

#### 핵심 3단계 에이전트 (자동 호출)

| 에이전트 | 페르소나 | 전문 영역 | 호출 시점 |
|---------|---------|----------|----------|
| **spec-builder** 🏗️ | 시스템 아키텍트 | EARS 명세 작성 | `/alfred:1-spec` |
| **code-builder** 💎 | 수석 개발자 | TDD 구현 | `/alfred:2-build` |
| **doc-syncer** 📖 | 테크니컬 라이터 | 문서 동기화 | `/alfred:3-sync` |

#### 품질 보증 에이전트 (온디맨드)

| 에이전트 | 페르소나 | 전문 영역 | 호출 방법 |
|---------|---------|----------|----------|
| **tag-agent** 🏷️ | 지식 관리자 | TAG 체인 검증 | `@agent-tag-agent` |
| **debug-helper** 🔬 | SRE 전문가 | 오류 진단 | `@agent-debug-helper` |
| **trust-checker** ✅ | QA 리드 | TRUST 검증 | `@agent-trust-checker` |
| **git-manager** 🚀 | 릴리스 엔지니어 | Git 워크플로우 | `@agent-git-manager` |

#### 시스템 관리 에이전트

| 에이전트 | 페르소나 | 전문 영역 | 호출 방법 |
|---------|---------|----------|----------|
| **cc-manager** 🛠️ | 데브옵스 엔지니어 | Claude Code 설정 | `@agent-cc-manager` |
| **project-manager** 📋 | 프로젝트 매니저 | 프로젝트 초기화 | `/alfred:8-project` |

### 협업 원칙

**단일 책임 (Single Responsibility)**:
- 각 에이전트는 자신의 전문 영역만 담당
- 다른 에이전트의 영역을 침범하지 않음

**중앙 조율 (Central Orchestration)**:
- Alfred만이 에이전트 간 작업을 조율
- 에이전트끼리 직접 호출 금지

**품질 게이트 (Quality Gates)**:
- 각 단계 완료 시 TRUST 원칙 자동 검증
- TAG 무결성 자동 확인

---

## 🎨 Alfred's Output Styles

Alfred는 개발 상황에 따라 **4가지 대화 스타일**을 제공합니다. Claude Code에서 `/output-style` 명령어로 언제든 전환할 수 있습니다.

### 📋 제공되는 Output Styles

| 스타일 이름 | 설명 |
|-----------|------|
| **MoAI Professional** | SPEC-First TDD 전문가를 위한 간결하고 기술적인 개발 스타일 |
| **MoAI Beginner Learning** | 개발 초보자를 위한 상세하고 친절한 단계별 학습 가이드 (학습 전용) |
| **MoAI Pair Collaboration** | AI와 함께 브레인스토밍, 계획 수립, 실시간 코드 리뷰를 진행하는 협업 모드 |
| **MoAI Study Deep** | 새로운 개념, 도구, 언어, 프레임워크를 체계적으로 학습하는 심화 교육 모드 |

### 🔄 스타일 전환 방법

Claude Code에서 `/output-style` 명령어로 전환:

```bash
/output-style alfred-pro           # MoAI Professional (기본값)
/output-style beginner-learning    # MoAI Beginner Learning
/output-style pair-collab          # MoAI Pair Collaboration
/output-style study-deep           # MoAI Study Deep
```

### 🎯 스타일 선택 가이드

| 상황 | 추천 스타일 | 대상 | 특징 |
|------|-----------|------|------|
| 실무 프로젝트 빠른 개발 | `alfred-pro` | 실무 개발자, 프로젝트 리더 | 간결, 기술적, 결과 중심 |
| 프로그래밍 처음 배우기 | `beginner-learning` | 개발 입문자 | 친절, 상세 설명, 단계별 안내 |
| 팀 기술 선택 & 설계 논의 | `pair-collab` | 협업 개발자, 아키텍트 | 질문 기반, 브레인스토밍 |
| 새로운 기술 학습 | `study-deep` | 신기술 학습자 | 개념 → 실습 → 전문가 팁 |

### 💡 모든 스타일에서 동일하게 작동

- ✅ 10개 AI 에이전트 팀 조율
- ✅ SPEC-First TDD 워크플로우
- ✅ TRUST 5원칙 자동 검증
- ✅ @TAG 추적성 보장

**차이점은 오직 설명 방식**:
- 📝 간결 vs 상세
- 🎓 빠른 구현 vs 개념 학습
- 💬 기술적 vs 친절 vs 협업적 vs 교육적

---

## SPEC 메타데이터 구조

모든 SPEC 문서는 표준화된 메타데이터 구조를 따릅니다.

### 필수 필드 (7개)

```yaml
id: AUTH-001                    # SPEC 고유 ID
version: 0.1.0                  # Semantic Version (v0.1.0 = INITIAL)
status: draft                   # draft|active|completed|deprecated
created: 2025-09-15            # 생성일 (YYYY-MM-DD)
updated: 2025-10-01            # 최종 수정일
author: @Goos                   # 작성자 (GitHub ID)
priority: high                  # low|medium|high|critical
```

### 선택 필드 (의존성 그래프 & 범위)

```yaml
# 분류
category: security              # feature|bugfix|refactor|security|docs|perf
labels: [authentication, jwt]   # 검색 태그

# 관계 (의존성 그래프)
depends_on: [USER-001]          # 의존하는 SPEC
blocks: [AUTH-002]              # 차단하는 SPEC
related_specs: [TOKEN-002]      # 관련 SPEC
related_issue: "github.com/..."  # GitHub Issue

# 범위 (영향 분석)
scope:
  packages: [src/core/auth]     # 영향받는 패키지
  files: [src/core/auth/service.py]  # 핵심 파일
```

**상세 가이드**: [SPEC 메타데이터 가이드](../.moai/memory/spec-metadata.md)

---

## @TAG 시스템

### TAG 체계 철학

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**핵심 원칙**:
1. **단순성**: 4개의 핵심 TAG만 사용
2. **TDD 완벽 정렬**: RED (TEST) → GREEN (CODE) → REFACTOR (DOC)
3. **CODE-FIRST**: TAG는 코드 자체에만 존재 (정규식 패턴으로 직접 스캔)
4. **무결성**: 고아 TAG 자동 탐지, 끊어진 참조 검증

### TAG 사용 규칙

**TAG ID 형식**: `<도메인>-<3자리>` (예: AUTH-003)

**중복 방지**:
```bash
# 새 TAG 생성 전 기존 TAG 검색
rg "@SPEC:AUTH" -n          # SPEC 문서에서 AUTH 도메인 검색
rg "@CODE:AUTH-001" -n      # 특정 ID 검색
```

**TAG 체인 검증**:
```bash
# /alfred:3-sync 실행 시 자동 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

### 사용 예시 (Python)

```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001/spec.md | TEST: tests/test_auth.py
"""
@CODE:AUTH-001: JWT 인증 서비스

TDD 이력:
- RED: pytest 테스트 작성
- GREEN: bcrypt + PyJWT 구현
- REFACTOR: 타입 힌트 추가
"""

class AuthService:
    # @CODE:AUTH-001:API: 인증 API 엔드포인트
    async def authenticate(
        self,
        username: str,
        password: str
    ) -> AuthResult:
        # @CODE:AUTH-001:DOMAIN: 입력 검증
        self._validate_input(username, password)

        # @CODE:AUTH-001:DATA: 사용자 조회
        user = await self.user_repo.find_by_username(username)

        return self._verify_credentials(user, password)
```

### 언어별 TAG 사용 예시

#### Flutter/Dart

```dart
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001/spec.md | TEST: test/auth_test.dart

/// @CODE:AUTH-001: JWT 인증 서비스
///
/// TDD 이력:
/// - RED: widget test 작성
/// - GREEN: dio + flutter_secure_storage 구현
/// - REFACTOR: Riverpod 상태 관리 통합
class AuthService {
  // @CODE:AUTH-001:API: 인증 API 엔드포인트
  Future<AuthResult> authenticate({
    required String username,
    required String password,
  }) async {
    // @CODE:AUTH-001:DOMAIN: 입력 검증
    _validateInput(username, password);

    // @CODE:AUTH-001:DATA: 사용자 조회
    final user = await userRepository.findByUsername(username);

    return _verifyCredentials(user, password);
  }
}
```

---

## 🌍 Universal Language Support

MoAI-ADK는 모든 주요 언어를 지원하며, 언어별 최적 도구 체인을 자동으로 선택합니다.

### 웹/백엔드

| 언어 | 테스트 | 린터 | 타입 | 상태 |
|------|--------|------|------|------|
| **TypeScript** | Vitest/Jest | Biome/ESLint | ✅ | Full |
| **Python** | pytest | ruff/black | mypy | Full |
| **Java** | JUnit 5 | checkstyle | ✅ | Full |
| **Go** | go test | golint | ✅ | Full |
| **Rust** | cargo test | clippy | ✅ | Full |

### 모바일

| 언어/프레임워크 | 테스트 | 린터 | 상태 |
|----------------|--------|------|------|
| **Flutter/Dart** | flutter test | dart analyze | Full |
| **Swift/iOS** | XCTest | SwiftLint | Full |
| **Kotlin/Android** | JUnit + Espresso | detekt | Full |
| **React Native** | Jest + RNTL | ESLint | Full |

### 자동 언어 감지

시스템이 프로젝트를 스캔하여 자동으로 언어를 감지하고 최적의 도구 체인을 선택합니다:

#### 웹/백엔드

| 감지 파일 | 언어/프레임워크 | 상태 |
|-----------|----------------|------|
| `package.json` | TypeScript/JavaScript | ✅ Full |
| `pyproject.toml`, `requirements.txt` | Python | ✅ Full |
| `pom.xml`, `build.gradle` | Java | ✅ Full |
| `go.mod` | Go | ✅ Full |
| `Cargo.toml` | Rust | ✅ Full |
| `*.csproj`, `*.sln` | C# | ✅ Full |
| `CMakeLists.txt`, `Makefile` | C/C++ | ✅ Full |
| `composer.json` | PHP | ✅ Full |
| `Gemfile`, `*.gemspec` | Ruby | ✅ Full |
| `build.sbt` | Scala | ✅ Full |
| `mix.exs` | Elixir | ✅ Full |

#### 모바일

| 감지 파일 | 언어/프레임워크 | 상태 |
|-----------|----------------|------|
| `pubspec.yaml` | Flutter/Dart | ✅ Full |
| `*.xcodeproj`, `Package.swift` | Swift/iOS | ✅ Full |
| `build.gradle` (Android) | Kotlin/Android | ✅ Full |
| `package.json` + `react-native` | React Native | ✅ Full |
| `*.xcodeproj` (Objective-C) | Objective-C/iOS | ✅ Full |

---

## 💻 CLI Reference

MoAI-ADK provides a Click 기반 CLI. 현재 패키지에는 `console_scripts` 엔트리포인트가 없으므로 아래처럼 `python -m moai_adk` 형태로 실행하는 것이 가장 확실합니다.

```bash
python -m moai_adk --help
python -m moai_adk init --help
```

### 사용 가능한 명령어

| Command | 설명 | 주요 옵션 |
| --- | --- | --- |
| `init [PATH]` | 5단계 Phase 파이프라인으로 프로젝트를 초기화/재초기화 | `--non-interactive/-y`, `--mode {personal|team}`, `--locale {ko|en|ja|zh}`, `--language`, `--force` |
| `doctor` | Python, Git, 프로젝트 구조를 점검하는 환경 진단 | `--verbose`, `--fix`, `--export`, `--check` |
| `status` | `.moai/config.json`과 SPEC 개수, Git 상태를 요약 | _없음_ |
| `backup` | `.moai`/`.claude`/`CLAUDE.md`를 선택적으로 백업 | `--path` |
| `restore` | `.moai-backups/{timestamp}/`에 저장된 스냅샷 복원 (현재는 미완성) | `--timestamp` |
| `update` | 번들된 템플릿으로 프로젝트 리소스를 업데이트 | `--path`, `--force`, `--check` |

### 명령어 상세

#### `init [PATH]`

```bash
python -m moai_adk init .
python -m moai_adk init my-project --non-interactive --mode team --locale en
python -m moai_adk init . --force
```

- 기본 동작은 대화형 모드이며, `--non-interactive/-y` 옵션을 사용하면 질문 없이 진행됩니다.
- `PhaseExecutor`가 5단계(Preparation → Directory → Resource → Configuration → Validation)를 순차 실행하며, 재초기화 시 `.moai-backups/{timestamp}/` 백업을 생성합니다 (최신 1개만 유지).
- `LanguageDetector`가 프로젝트 루트를 스캔하여 언어를 자동 감지합니다 (`--language`로 오버라이드 가능).

#### `doctor`

```bash
python -m moai_adk doctor
python -m moai_adk doctor --verbose          # 언어별 도구 + 버전 표시
python -m moai_adk doctor --fix              # 누락 도구 설치 제안
python -m moai_adk doctor --export report.json  # JSON 파일 저장
python -m moai_adk doctor --check pytest     # 특정 도구만 검증
```

**기능**:
- 시스템 진단: Python ≥ 3.13, Git 설치 여부, `.moai/` 구조 유무 확인
- **20개 언어별 도구 체인 검증**: Python, TypeScript, JavaScript, Java, Go, Rust, Dart, Swift, Kotlin, C#, PHP, Ruby, Elixir, Scala, Clojure, Haskell, C, C++, Lua, OCaml
- 언어 감지 자동화: 프로젝트 구조 기반 언어 자동 감지 및 필수 도구 검증

**옵션**:
- `--verbose`: 모든 선택 도구 및 버전 정보 표시
- `--fix`: 누락된 도구에 대한 설치 명령어 제안 및 사용자 승인 후 실행
- `--export <file>`: 진단 결과를 JSON 파일로 저장
- `--check <tool>`: 특정 도구만 검증 (예: `pytest`, `vitest`)

**언어별 도구 체인 예시**:
- **Python**: pytest, mypy, ruff, black
- **TypeScript**: vitest, biome, typescript, eslint
- **Java**: maven, gradle, junit
- **Go**: golangci-lint, gofmt, go test
- **Rust**: rustfmt, clippy, cargo test

**@CODE TAG**: `@CODE:CLI-001`

#### `status`

```bash
python -m moai_adk status
```

현재 디렉토리의 `.moai/config.json`을 읽고, SPEC 문서 개수를 집계하며, GitPython을 통해 브랜치/dirty 여부를 보고합니다. 설정 파일이 없으면 실행을 중단하고 `init`을 안내합니다.

#### `backup`

```bash
python -m moai_adk backup
python -m moai_adk backup --path /path/to/project
```

`TemplateProcessor.create_backup()`을 사용해 `.moai-backup/<timestamp>/`에 백업을 생성합니다. SPEC/보고서는 보호 경로로 간주되어 제외됩니다.

#### `restore`

```bash
python -m moai_adk restore
python -m moai_adk restore --timestamp 20250301-130500
```

가장 최근 백업 혹은 지정한 타임스탬프를 찾아 복원 절차를 안내합니다. 현재 구현은 경로만 안내하며 실질적인 파일 복원은 TODO 상태입니다.

#### `update`

```bash
python -m moai_adk update
python -m moai_adk update --check
python -m moai_adk update --force
```

템플릿을 재적용하여 `.claude/`, `.moai/`, `CLAUDE.md`, `.gitignore`를 동기화합니다. `--check`는 버전 비교만 수행하며, `--force`는 백업 생성을 건너뜁니다.

### 종료 코드

- 정상 종료: `0`
- 사용자가 `Ctrl+C` 등으로 중단: `130`
- `click.ClickException`/예상 가능한 오류: 해당 exit code (기본 1)
- 알 수 없는 예외: `1`

---

## 🔒 보안 스캔

MoAI-ADK는 코드 보안을 위해 두 가지 도구를 제공합니다:

### 보안 도구

- **pip-audit**: 의존성 패키지의 알려진 취약점 검사
- **bandit**: Python 소스 코드의 보안 이슈 검사

### 로컬 보안 스캔 실행

**자동화 스크립트 사용 (권장)**:
```bash
./scripts/security-scan.sh
```

**개별 도구 실행**:
```bash
# 보안 도구 설치
pip install pip-audit bandit

# 의존성 취약점 스캔
pip-audit

# 코드 보안 스캔 (Low severity 제외)
bandit -r src/ -ll
```

### CI/CD 통합

GitHub Actions 워크플로우가 자동으로 보안 스캔을 실행합니다:
- `main`, `develop`, `feature/**` 브랜치 push 시
- Pull Request 생성/업데이트 시

워크플로우 파일: `.github/workflows/security.yml`

### 취약점 해결

**pip-audit 취약점 발견 시**:
1. 출력된 `Fix Versions` 확인
2. `pyproject.toml`에서 해당 패키지 버전 업데이트
3. `pip install -e ".[security]"` 재설치
4. 다시 `pip-audit` 실행하여 확인

**bandit 보안 이슈 발견 시**:
1. 파일 위치로 이동
2. bandit 권장사항 확인
3. 코드 수정 또는 정당한 사유가 있으면 `# nosec` 주석 추가

**@CODE TAG**: `@CODE:SECURITY-001`

---

## 🏁 Checkpoint - 개발 현황 스냅샷

> **마지막 업데이트**: 2025-10-15
> **현재 버전**: v0.3.1
> **프로젝트 상태**: INIT-003 v0.3.1 완료 (Event-Driven Checkpoint 자동화)

### 📊 주요 지표

| 항목 | 현재 상태 | 목표 | 상태 |
|------|----------|------|------|
| **테스트 커버리지** | 85.61% | 85%+ | ✅ |
| **SPEC 문서** | 18개 | - | ✅ |
| **지원 언어** | 20개 | 10+ | ✅ |
| **CLI 명령어** | 8개 | - | ✅ |
| **AI 에이전트** | 10개 | - | ✅ |

### 🎯 최근 완료된 작업 (SPEC-INIT-003 v0.3.1)

**SPEC-INIT-003: Event-Driven Checkpoint 시스템**
- **버전**: v0.3.1 (active)
- **완료일**: 2025-10-15
- **주요 성과**:
  - ✅ Claude Code Hooks 통합 (SessionStart, PreToolUse, PostToolUse)
  - ✅ BackupMerger 클래스 구현 (백업 병합 기능)
  - ✅ 버전 추적 시스템 (config.json에 moai_adk_version, optimized 필드)
  - ✅ 자동 최적화 감지 (Claude 접속 시 버전 불일치 알림)
  - ✅ 변경량: +1,180줄 추가, -2,076줄 삭제

**TDD 이력**:
- 🟢 GREEN: 커밋 3b8c7bc (Claude Code Hooks 기반 Checkpoint 자동화 구현 완료)
- 📝 DOCS: 커밋 c3c48ac (CHECKPOINT-EVENT-001 문서 동기화 완료)
- 📝 DOCS: 커밋 1714724 (SPEC-INIT-003 v0.3.1 작성 완료)

**이전 완료 작업 (SPEC-CLI-001)**:
- ✅ doctor 명령어 고도화: 20개 언어 도구 체인 검증 구현
- ✅ 옵션 추가: `--verbose`, `--fix`, `--export`, `--check`
- ✅ 테스트 100% 통과 (50개 테스트)
- ✅ 커버리지 91.58% (doctor.py)

### 📦 현재 기능 목록

#### 핵심 3단계 워크플로우
- ✅ `/alfred:1-spec` - EARS 명세 작성
- ✅ `/alfred:2-build` - TDD 구현 (RED-GREEN-REFACTOR)
- ✅ `/alfred:3-sync` - 문서 동기화 및 TAG 검증

#### CLI 명령어
- ✅ `python -m moai_adk init` - 프로젝트 초기화 (5 Phase 파이프라인)
- ✅ `python -m moai_adk doctor` - 환경 진단 (20개 언어 도구 체인 검증) 🆕
- ✅ `python -m moai_adk status` - 프로젝트 현황 요약
- ✅ `python -m moai_adk backup` - 프로젝트 백업
- ✅ `python -m moai_adk restore` - 백업 복원
- ✅ `python -m moai_adk update` - 템플릿 업데이트

#### AI 에이전트 생태계
- ✅ Alfred (SuperAgent) - 중앙 오케스트레이터
- ✅ spec-builder - SPEC 작성 전문가
- ✅ code-builder - TDD 구현 전문가
- ✅ doc-syncer - 문서 동기화 전문가
- ✅ tag-agent - TAG 시스템 관리자
- ✅ debug-helper - 디버깅 전문가
- ✅ trust-checker - 품질 검증 전문가
- ✅ git-manager - Git 워크플로우 관리자
- ✅ cc-manager - Claude Code 설정 관리자
- ✅ project-manager - 프로젝트 초기화 관리자

### 🚀 다음 단계

**우선순위 높음**:
1. status 명령어 고도화 (TAG 체인 무결성, 커버리지 표시)
2. restore 명령어 선택적 복원 기능
3. 프로젝트 템플릿 최적화

**우선순위 중간**:
4. 추가 언어 지원 (Zig, Nim, Crystal)
5. CI/CD GitHub Actions 워크플로우 개선
6. 성능 최적화 (doctor 실행 시간 단축)

**우선순위 낮음**:
7. GUI 인터페이스 프로토타입
8. 플러그인 시스템 설계
9. 원격 진단 기능

### 📈 성장 지표

**프로젝트 성숙도**:
- 코드 품질: 85.61% 테스트 커버리지, TRUST 5원칙 준수
- 문서화: 18개 SPEC, README.md, development-guide.md 완비
- 자동화: 10개 AI 에이전트, 8개 CLI 명령어
- 언어 지원: 20개 주요 프로그래밍 언어

**커뮤니티**:
- GitHub Stars: 진행 중
- Contributors: 2명 ([@Goos](https://github.com/Goos), [@Workuul](https://github.com/Workuul))
- Issues: 활발한 피드백 수집 중

### 🔗 참고 링크

- **SPEC 문서**: [.moai/specs/SPEC-CLI-001/spec.md](/.moai/specs/SPEC-CLI-001/spec.md)
- **개발 가이드**: [.moai/memory/development-guide.md](/.moai/memory/development-guide.md)
- **SPEC 메타데이터 표준**: [.moai/memory/spec-metadata.md](/.moai/memory/spec-metadata.md)
- **GitHub Repository**: [modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)

---

## 프로그래매틱 API

CLI 없이도 `moai_adk.core` 모듈을 직접 사용할 수 있습니다.

### ProjectInitializer

```python
from pathlib import Path
from moai_adk.core.project.initializer import ProjectInitializer

initializer = ProjectInitializer(Path("./demo"))
result = initializer.initialize(mode="team", locale="en", backup_enabled=True)

if result.success:
    print("created", result.created_files)
else:
    print("errors", result.errors)
```

### TemplateProcessor

```python
from pathlib import Path
from moai_adk.core.template.processor import TemplateProcessor

processor = TemplateProcessor(Path("./demo"))
processor.copy_templates(backup=True, silent=False)
config = processor.merge_config(detected_language="python")
```

### Environment Checks

```python
from moai_adk.core.project.checker import SystemChecker, check_environment

checker = SystemChecker()
tools = checker.check_all()
diagnostics = check_environment()
```

### ConfigManager

```python
from pathlib import Path
from moai_adk.core.template.config import ConfigManager

config_path = Path("./demo/.moai/config.json")
manager = ConfigManager(config_path)
config = manager.load()
config["mode"] = "team"
manager.save(config)
```

이러한 빌딩 블록을 조합하면 CI 파이프라인이나 맞춤형 워크플로우에서 MoAI-ADK의 핵심 기능을 재사용할 수 있습니다.

---

## TRUST 5원칙

모든 개발 과정에서 TRUST 원칙을 준수합니다:

### T - Test First (테스트 우선)

**SPEC → Test → Code 사이클**:
- **@SPEC**: EARS 형식 명세서 우선 작성
- **RED**: `@TEST` TAG - 실패하는 테스트 작성
- **GREEN**: `@CODE` TAG - 최소 구현으로 테스트 통과
- **REFACTOR**: `@CODE` TAG - 코드 품질 개선

### R - Readable (가독성)

**코드 제약**:
- 파일당 ≤300 LOC
- 함수당 ≤50 LOC
- 매개변수 ≤5개
- 복잡도 ≤10

### U - Unified (통합성)

**SPEC 기반 아키텍처**:
- 모듈 간 명확한 책임 분리
- 타입 안전성 보장
- 언어별 경계를 SPEC이 정의

### S - Secured (보안성)

**입력 검증**:
- 모든 사용자 입력 검증 (정규식, 화이트리스트)
- 파일 업로드 제한 (확장자, 크기, MIME 타입)

**주요 취약점 방어**:
- **SQL Injection**: Prepared Statement, ORM 사용
- **XSS**: HTML 이스케이핑, CSP 헤더
- **CSRF**: CSRF 토큰, SameSite 쿠키
- **비밀번호**: bcrypt/argon2 해싱 (최소 10 라운드)

**보안 스캐닝**:
- 정적 분석 도구 (Snyk, OWASP Dependency-Check)
- 환경 변수 보안 (`.env` Git 제외)

### T - Trackable (추적성)

**@TAG 시스템으로 완전한 추적성**:
- `@SPEC` → `@TEST` → `@CODE` → `@DOC` 체인
- 코드 직접 스캔으로 무결성 검증
- 고아 TAG 자동 탐지

---

## 문제 해결

### 자주 발생하는 문제

#### 1. `/alfred:2-build` 실행 시 "SPEC not found" 에러

**증상**: TDD 구현 중 SPEC 파일을 찾을 수 없다는 에러 발생

**원인**: `/alfred:1-spec` 단계를 건너뛰었거나, SPEC 파일 경로가 잘못됨

**해결 방법**:

```bash
# 1. SPEC 파일 존재 여부 확인
ls .moai/specs/SPEC-*.md

# 2. SPEC이 없다면 먼저 작성
/alfred:1-spec "기능 설명"

# 3. SPEC ID 확인 후 재실행
/alfred:2-build SPEC-ID
```

#### 2. 테스트 실패 시 복구

**증상**: `/alfred:2-build` 실행 후 테스트가 계속 실패

**원인**: 엣지 케이스 누락, 의존성 문제, 환경 변수 미설정

**해결 방법**:

```bash
# 1. 테스트 수동 실행으로 정확한 에러 확인
npm test  # 또는 bun test, pytest 등

# 2. debug-helper 에이전트 호출
@agent-debug-helper "테스트 실패 에러 메시지"

# 3. 환경 변수 확인
cat .env.example  # 필요한 환경 변수 확인
cp .env.example .env  # 환경 변수 파일 생성

# 4. 의존성 재설치
rm -rf node_modules && npm install
```

#### 3. TAG 체인 끊어짐 경고

**증상**: `/alfred:3-sync` 실행 시 "고아 TAG 발견" 경고

**원인**: SPEC 없이 CODE만 작성했거나, TAG ID 불일치

**해결 방법**:

```bash
# 1. 고아 TAG 찾기
rg '@CODE:' -n src/  # CODE TAG 목록
rg '@SPEC:' -n .moai/specs/  # SPEC TAG 목록

# 2. 누락된 SPEC 작성
/alfred:1-spec "해당 기능 설명"

# 3. TAG ID 일치시키기
# CODE와 SPEC의 ID가 동일한지 확인 (예: AUTH-001)

# 4. 재검증
/alfred:3-sync
```

#### 4. Git 브랜치 충돌

**증상**: SPEC 생성 시 브랜치 생성 실패

**원인**: 동일한 이름의 브랜치가 이미 존재

**해결 방법**:

```bash
# 1. 기존 브랜치 확인
git branch -a

# 2. 기존 브랜치로 전환 (계속 작업하려면)
git checkout feature/SPEC-XXX-YYY

# 3. 또는 새 브랜치 강제 생성 (처음부터 다시 시작)
git branch -D feature/SPEC-XXX-YYY
/alfred:1-spec "기능 설명"
```

#### 5. 권한 에러 (Permission Denied)

**증상**: `moai-adk init` 실행 시 권한 에러

**원인**: 파일 실행 권한 부족

**해결 방법**:

```bash
# 1. .claude/commands/ 디렉토리 권한 확인
ls -la .claude/commands/

# 2. 실행 권한 추가
chmod +x .claude/commands/*.md

# 3. 또는 재초기화
moai-adk init . --force
```

#### 6. 테스트 커버리지 85% 미만

**증상**: TRUST 검증 실패 - 테스트 커버리지 부족

**원인**: 엣지 케이스 테스트 누락

**해결 방법**:

```bash
# 1. 커버리지 리포트 확인
npm test -- --coverage  # 또는 bun test --coverage

# 2. 누락된 브랜치 확인
# 커버리지 리포트에서 빨간색(미테스트) 라인 확인

# 3. 엣지 케이스 테스트 추가
# - null/undefined 입력
# - 빈 배열/객체
# - 경계값 (0, -1, 최대값)
# - 에러 케이스

# 4. 재실행
/alfred:2-build SPEC-ID
```

#### 7. 설치 실패

**권한 문제:**
```bash
sudo npm install -g moai-adk
```

**캐시 문제:**
```bash
npm cache clean --force
npm install -g moai-adk
```

#### 8. 명령어 인식 안 됨

**PATH 확인:**
```bash
echo $PATH
npm list -g --depth=0
```

**셸 재시작:**
```bash
source ~/.bashrc  # bash
source ~/.zshrc   # zsh
```

#### 9. Claude Code 연동 문제

- `.claude/settings.json` 파일 확인
- Claude Code 최신 버전 사용 확인
- 에이전트 파일 권한 확인

### 로그 확인

문제 원인 파악을 위한 로그 위치:

```bash
# MoAI-ADK 시스템 로그
~/.moai/logs/moai.log

# 에러 로그
~/.moai/logs/error.log

# 프로젝트별 로그
.moai/logs/

# Claude Code 로그
~/.claude/logs/
```

### 긴급 복구

심각한 문제 발생 시 백업에서 복원:

```bash
# 1. 백업 목록 확인
python -m moai_adk status

# 2. 최신 백업으로 복원
python -m moai_adk restore

# 3. 특정 백업으로 복원
python -m moai_adk restore --timestamp YYYY-MM-DD-HHMMSS
```

---

## 개발 참여

### 기여 방법

1. Repository Fork
2. 기능 브랜치 생성 (`git checkout -b feature/new-feature`)
3. 변경사항 커밋 (`git commit -am 'Add new feature'`)
4. 브랜치 푸시 (`git push origin feature/new-feature`)
5. Pull Request 생성

### 개발 환경 설정

```bash
# 저장소 클론
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk

# 의존성 설치 (uv 권장)
uv pip install -e ".[dev]"

# 개발 모드 실행
python -m moai_adk

# 테스트
pytest

# 코드 품질 검사
ruff check .
mypy src/
```

### 코딩 규칙

- TRUST 5원칙 준수
- @TAG 시스템 적용
- Python 타입 힌트 필수 (mypy strict)
- ≤50 LOC per function
- Test coverage ≥85%

---

## 문서 및 지원

- **🐛 Issues**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **📦 npm Package**: [moai-adk](https://www.npmjs.com/package/moai-adk)

## 🙏 Contributors

MoAI-ADK 프로젝트에 기여해주신 분들께 감사드립니다:

- **[@Workuul](https://github.com/Workuul)** - 심볼릭 링크 실행 문제 수정 ([PR #1](https://github.com/modu-ai/moai-adk/pull/1))
  - `realpathSync()` 적용으로 글로벌 설치 이슈 해결
  - REPL/eval 환경 방어 로직 추가
  - JSDoc 문서화 개선

[^haiku]: [Anthropic — Introducing Claude Haiku 4.5 (2025-10-15)](https://www.anthropic.com/news/claude-haiku-4-5)
[^sonnet]: [Anthropic — Introducing Claude Sonnet 4.5 (2025-09-29)](https://www.anthropic.com/news/claude-sonnet-4-5)

---

Made with MoAI's 🪿

---

## 라이선스

이 프로젝트는 [MIT License](LICENSE)를 따릅니다.
