# MoAI-ADK (Agentic Development Kit)

[![npm version](https://img.shields.io/npm/v/moai-adk)](https://www.npmjs.com/package/moai-adk)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.2+-blue)](https://www.typescriptlang.org/)
[![Node.js](https://img.shields.io/badge/node-18.0+-green)](https://nodejs.org/)
[![Bun](https://img.shields.io/badge/Bun-1.2.19+-black)](https://bun.sh/)

## MoAI-ADK

### 모두의AI 에이전틱 코딩 개발 프레임워크

**안내:** MoAI-ADK는 모두의AI 연구실에서 집필 중인 "(가칭) 에이전틱 코딩" 서적의 별책 부록 오픈 소스 프로젝트 임을 밝혀 둡니다.

![MoAI-ADK CLI Interface](https://github.com/modu-ai/moai-adk/raw/main/docs/public/moai-tui_screen-light.png)

> "SPEC이 없으면 CODE도 없다."

---

## 목차

- [Meet Alfred](#-meet-alfred---your-ai-development-partner)
- [Quick Start](#-quick-start-3분-실전)
- [The Problem](#-the-problem---바이브-코딩의-한계)
- [The Solution](#-the-solution---3단계-워크플로우)
- [How Alfred Works](#️-how-alfred-works---10개-ai-에이전트-팀)
- [Output Styles](#-alfreds-output-styles)
- [Language Support](#-universal-language-support)
- [Future Roadmap](#-future-roadmap)
- [CLI Reference](#-cli-reference)
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

- ✅ Bun 또는 npm 설치됨
- ✅ Claude Code 실행 중
- ✅ Git 설치됨 (선택사항)

### ⚡ 3단계로 시작하기

#### 1️⃣ 설치 (30초)

```bash
# Bun 권장 (5배 빠른 성능)
curl -fsSL https://bun.sh/install | bash
bun add -g moai-adk

# 또는 npm 사용
npm install -g moai-adk

# 설치 확인
moai --version
# 출력: v0.x.x
```

#### 2️⃣ 초기화 (1분)

**터미널에서:**
```bash
# 새 프로젝트 생성
moai init my-project
cd my-project

# 기존 프로젝트에 설치
cd existing-project
moai init .

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

**생성된 것들:**
- ✅ `.moai/specs/SPEC-AUTH-001.md` (명세)
- ✅ `tests/auth/login.test.ts` (테스트)
- ✅ `src/services/auth.ts` (구현)
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

- **Node.js**: 18.0 이상
- **Git**: 2.30.0 이상
- **npm**: 8.0.0 이상 (또는 **Bun 1.2.0 이상 강력 추천**)
- **Claude Code**: v1.2.0 이상 (에이전트 시스템 완전 통합용)

### 🌍 지원 운영체제

- **Windows**: 10/11 (PowerShell 5.1+)
- **macOS**: 12 Monterey 이상 (M1/M2 네이티브 지원)
- **Linux**: Ubuntu 20.04+, CentOS 8+, Debian 11+, Arch Linux

---

## 설치

### Option A: Bun 설치 (최적 성능, 강력 추천) 🔥

```bash
# Bun 설치 (아직 없는 경우)
curl -fsSL https://bun.sh/install | bash  # macOS/Linux
# 또는
powershell -c "iwr bun.sh/install.ps1|iex"  # Windows

# MoAI-ADK 전역 설치
bun add -g moai-adk
```

### Option B: npm 설치 (표준 옵션)

```bash
npm install -g moai-adk
```

### Option C: 개발자 설치 (로컬 개발용)

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk/moai-adk-ts
bun install  # 또는 npm install
bun run build
npm link
```

### 설치 확인

```bash
# 버전 확인
moai --version

# 시스템 진단
moai doctor

# 도움말
moai help
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

### 사용 예시

```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

/**
 * @CODE:AUTH-001: JWT 인증 서비스
 *
 * TDD 이력:
 * - RED: tests/auth/service.test.ts 작성
 * - GREEN: 최소 구현 (bcrypt, JWT)
 * - REFACTOR: 타입 안전성 추가
 */
export class AuthService {
  // @CODE:AUTH-001:API: 인증 API 엔드포인트
  async authenticate(username: string, password: string): Promise<AuthResult> {
    // @CODE:AUTH-001:DOMAIN: 입력 검증
    this.validateInput(username, password);

    // @CODE:AUTH-001:DATA: 사용자 조회
    const user = await this.userRepository.findByUsername(username);

    return this.verifyCredentials(user, password);
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

시스템이 프로젝트를 스캔하여 자동으로 감지:

- `package.json` → TypeScript/JavaScript
- `requirements.txt` → Python
- `go.mod` → Go
- `Cargo.toml` → Rust
- `pubspec.yaml` → Flutter/Dart

---

## 🔮 Future Roadmap

### Alfred - 첫 번째 공개 SuperAgent

**Alfred**는 MoAI-ADK의 **첫 번째 공개 SuperAgent**입니다. 개발 워크플로우(SPEC-First TDD)에 특화되어 있으며, 10개 AI 에이전트 팀을 조율합니다.

### 향후 추가 예정 SuperAgent

MoAI-ADK는 **모듈형 에이전트 아키텍처**로 설계되어, 다양한 도메인별 SuperAgent를 추가할 수 있습니다:

| SuperAgent | 전문 영역 | 주요 기능 | 예상 출시 |
|-----------|---------|----------|----------|
| 🖋️ **Writing Agent** | 기술 문서 작성 | 블로그, 튜토리얼, API 문서 자동 생성 | TBD |
| 🚀 **Startup MVP Agent** | 제품 개발 | 기획→개발→배포 End-to-End 지원 | TBD |
| 📊 **Analytics Agent** | 데이터 분석 | EDA, 시각화, ML 모델 추천 | TBD |
| 🎨 **Design System Agent** | UI/UX | 컴포넌트 라이브러리, 디자인 토큰 관리 | TBD |

### 에이전트 생태계 비전

```text
MoAI-ADK Ecosystem
├─ ▶◀ Alfred (개발 워크플로우) ← 현재 공개
│   └─ 9개 서브에이전트
├─ 🖋️ Writing Agent (콘텐츠 제작) ← 향후
├─ 🚀 Startup MVP Agent (제품 개발) ← 향후
├─ 📊 Analytics Agent (데이터 분석) ← 향후
└─ 🎨 Design System Agent (디자인 시스템) ← 향후
```

**핵심 철학**: "한 가지를 완벽하게, 그 다음 확장하라"

---

## 💻 CLI Reference

### 핵심 명령어

```bash
# 프로젝트 초기화
moai init [project] [options]

# 시스템 진단
moai doctor [options]

# 프로젝트 상태 확인
moai status [options]

# 백업 복원
moai restore <backup-path> [options]
```

### Claude Code 전용 명령어

```text
# 템플릿 업데이트 (권장 ⭐)
/alfred:9-update

# 프로젝트 초기화
/alfred:8-project
```

### moai init [project]

새 MoAI-ADK 프로젝트를 초기화하거나 기존 프로젝트에 MoAI-ADK를 설치합니다.

**옵션**:

- `--personal`: Personal 모드로 초기화 (기본값)
- `--team`: Team 모드로 초기화 (GitHub 통합)
- `-b, --backup`: 설치 전 백업 생성
- `-f, --force`: 기존 파일 강제 덮어쓰기

**사용 예시**:

```bash
# 새 프로젝트 생성 (Personal 모드)
moai init my-project

# 현재 디렉토리에 설치
moai init .

# Team 모드로 초기화
moai init my-project --team

# 백업 생성 후 설치
moai init . -b

# 기존 파일 강제 덮어쓰기
moai init . -f
```

### moai doctor

시스템 진단을 실행하여 MoAI-ADK가 올바르게 설치되었는지 확인합니다.

**옵션**:

- `-l, --list-backups`: 사용 가능한 백업 목록 표시

**사용 예시**:

```bash
# 시스템 진단 실행
moai doctor

# 백업 목록 확인
moai doctor -l
```

### moai status

MoAI-ADK 프로젝트 상태를 표시합니다.

**옵션**:

- `-v, --verbose`: 상세 상태 정보 표시
- `-p, --project-path <path>`: 프로젝트 디렉토리 경로 지정 (경로 필수)

**사용 예시**:

```bash
# 현재 디렉토리 상태 확인
moai status

# 상세 정보 포함
moai status -v

# 특정 경로 프로젝트 상태 확인
moai status -p /path/to/project

# 상세 정보 + 특정 경로
moai status -v -p /path/to/project
```

### moai restore <backup-path>

백업 디렉토리에서 MoAI-ADK를 복원합니다.

**인자**:

- `<backup-path>`: 복원할 백업 디렉토리 경로 (필수)

**옵션**:

- `--dry-run`: 변경 없이 복원할 내용 미리보기
- `--force`: 기존 파일 강제 덮어쓰기

**사용 예시**:

```bash
# 백업에서 복원 (미리보기)
moai restore .moai-backup-2025-10-02 --dry-run

# 실제 복원 실행
moai restore .moai-backup-2025-10-02

# 강제 복원 (기존 파일 덮어쓰기)
moai restore .moai-backup-2025-10-02 --force
```

**참고**: MoAI-ADK 업데이트는 Claude Code에서 `/alfred:9-update` 명령어를 사용하세요.

---

## 프로그래매틱 API

### 기본 사용

```typescript
import { CLIApp, SystemChecker, TemplateManager } from 'moai-adk';

// CLI 앱 초기화
const app = new CLIApp();
await app.run();

// 시스템 체크
const checker = new SystemChecker();
const result = await checker.checkSystem();

// 템플릿 관리
const templateManager = new TemplateManager();
await templateManager.copyTemplates(projectPath);
```

### 설정 파일 (.moai/config.json)

```json
{
  "project": {
    "name": "my-project",
    "mode": "personal",
    "language": "typescript"
  },
  "workflow": {
    "enableAutoSync": true,
    "gitIntegration": true
  }
}
```

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

**보안 by 설계**:
- 입력 검증
- 정적 분석
- 보안 스캐닝
- 접근 제어

### T - Trackable (추적성)

**@TAG 시스템으로 완전한 추적성**:
- `@SPEC` → `@TEST` → `@CODE` → `@DOC` 체인
- 코드 직접 스캔으로 무결성 검증
- 고아 TAG 자동 탐지

---

## 문제 해결

### 자주 발생하는 문제

#### 1. 설치 실패

**권한 문제:**
```bash
sudo npm install -g moai-adk
```

**캐시 문제:**
```bash
npm cache clean --force
npm install -g moai-adk
```

#### 2. 명령어 인식 안 됨

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

#### 3. Claude Code 연동 문제

- `.claude/settings.json` 파일 확인
- Claude Code 최신 버전 사용 확인
- 에이전트 파일 권한 확인

### 로그 확인

```bash
# 일반 로그
~/.moai/logs/moai.log

# 에러 로그
~/.moai/logs/error.log

# 프로젝트별 로그
.moai/logs/
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
cd moai-adk/moai-adk-ts

# 의존성 설치 (Bun 권장)
bun install

# 개발 모드 실행
bun run dev

# 빌드
bun run build

# 테스트
bun test

# 코드 품질 검사
bun run check
```

### 코딩 규칙

- TRUST 5원칙 준수
- @TAG 시스템 적용
- TypeScript strict 모드 사용
- ≤50 LOC per function
- Test coverage ≥85%

---

## 라이선스

이 프로젝트는 [MIT License](LICENSE)를 따릅니다.

---

## 문서 및 지원

- **🐛 Issues**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)
- **📦 npm Package**: [moai-adk](https://www.npmjs.com/package/moai-adk)

---

Made with MoAI's 🪿

---

## 라이선스

이 프로젝트는 [MIT License](LICENSE)를 따릅니다.
