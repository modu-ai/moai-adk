---
layout: home

hero:
  name: "MoAI-ADK"
  text: "Agentic Development Kit for Agents"
  tagline: Claude Code 기반 SPEC-First TDD 범용 개발 툴킷
  actions:
    - theme: brand
      text: 시작하기
      link: /getting-started/installation
    - theme: alt
      text: GitHub 보기
      link: https://github.com/modu-ai/moai-adk
  image:
    light: /moai-logo-light.png
    dark: /moai-logo-dark.png
    alt: MoAI-ADK

features:
  - icon:
      src: /icons/spec.svg
      alt: SPEC
    title: SPEC 우선 개발
    details: 코드보다 명세를 먼저 작성합니다. 모든 구현은 EARS 방법론을 사용한 명확한 요구사항으로 시작합니다.
  - icon:
      src: /icons/test.svg
      alt: Test
    title: 테스트 주도 개발
    details: Red-Green-Refactor 사이클 엄격 적용. 테스트 없이는 구현 없음. Vitest 3.2.4 기반 자동화 테스트 지원.
  - icon:
      src: /icons/tag.svg
      alt: TAG
    title: CODE-FIRST TAG 시스템
    details: 요구사항부터 구현까지 추적성 제공. 중간 캐시 없이 소스코드 직접 스캔. ripgrep 기반 실시간 검증.
  - icon:
      src: /icons/language.svg
      alt: Languages
    title: 다중 언어 지원
    details: TypeScript, Python, Java, Go, Rust, C++, C#, PHP 총 8개 언어 지원. 프로젝트 파일 분석 기반 언어 감지 및 도구 매핑.
  - icon:
      src: /icons/performance.svg
      alt: Performance
    title: 고성능 빌드
    details: Bun 패키지 관리 (npm 대비 약 10-20배 빠른 설치). tsup 빌드 182ms. Biome 린터 (ESLint+Prettier 통합).
  - icon:
      src: /icons/claude.svg
      alt: Claude
    title: Claude Code 통합
    details: Claude Code 네이티브 통합. 9개 전문 에이전트, 5개 워크플로우 명령어, 개발 자동화 지원.
---

## 🎩 SuperAgent Alfred

**모두의 AI 집사 Alfred**는 MoAI-ADK의 핵심 오케스트레이터입니다. 정확하고 예의 바르며, 모든 요청을 체계적으로 처리하는 전문 지휘자 역할을 수행합니다.

### 핵심 역할

- **사용자 요청 분석 및 라우팅**: 요청의 본질을 파악하고 적절한 Sub-Agent 식별
- **Sub-Agent 위임 전략**: 직접 처리, 단일 에이전트, 순차 실행, 병렬 실행 등 최적의 실행 전략 수립
- **품질 게이트 검증**: 각 단계 완료 시 TRUST 원칙 준수 확인, @TAG 추적성 체인 무결성 검증

Alfred는 복잡한 개발 작업을 체계적으로 분해하고, 9개의 전문 에이전트를 효율적으로 조율하여 SPEC-First TDD 방법론을 통한 완벽한 코드 품질을 보장합니다.

---

## 📋 9개 전문 에이전트 생태계

MoAI-ADK는 전문 개발사의 직무 체계를 모델링한 9개의 전문 에이전트를 제공합니다:

| 에이전트 | 직무 페르소나 | 전문 영역 | 핵심 책임 |
|---------|--------------|----------|----------|
| 🏗️ **spec-builder** | 시스템 아키텍트 | 요구사항 설계 | EARS 명세, 아키텍처 설계 |
| 💎 **code-builder** | 수석 개발자 | TDD 구현 | Red-Green-Refactor, 코드 품질 |
| 📖 **doc-syncer** | 테크니컬 라이터 | 문서 관리 | Living Document, API 문서 동기화 |
| 🏷️ **tag-agent** | 지식 관리자 | 추적성 관리 | TAG 시스템, 코드 스캔, 체인 검증 |
| 🚀 **git-manager** | 릴리스 엔지니어 | 버전 관리 | Git 워크플로우, 브랜치 전략, 배포 |
| 🔬 **debug-helper** | 트러블슈팅 전문가 | 문제 해결 | 오류 진단, 근본 원인 분석, 해결 방안 |
| ✅ **trust-checker** | 품질 보증 리드 | 품질 검증 | TRUST 5원칙, 성능/보안 검사 |
| 🛠️ **cc-manager** | 데브옵스 엔지니어 | 개발 환경 | Claude Code 설정, 권한, 표준화 |
| 📋 **project-manager** | 프로젝트 매니저 | 프로젝트 관리 | 초기화, 문서 구축, 전략 수립 |

각 에이전트는 단일 책임 원칙을 준수하며, Alfred의 조율 하에 협업하여 최고의 코드 품질을 보장합니다.

---

## 🚀 3단계 워크플로우

MoAI-ADK는 SPEC-First TDD 방법론을 3단계 워크플로우로 자동화합니다:

### 1️⃣ /moai:1-spec - SPEC 작성

**담당**: spec-builder (🏗️ 설계자) + git-manager (🚀 정원사)

- EARS 방법론을 사용한 요구사항 명세 작성
- 아키텍처 설계 및 기술적 의사결정
- 브랜치 생성 및 GitHub Issue/PR 자동 생성 (Team 모드)
- `.moai/specs/SPEC-XXX/` 디렉토리 생성 (Personal 모드)

**핵심 원칙**: 명세 없이는 코드 없음

### 2️⃣ /moai:2-build - TDD 구현

**담당**: code-builder (💎 장인) + git-manager (🚀 정원사)

- **RED**: 실패하는 테스트 작성 및 실패 확인
- **GREEN**: 테스트를 통과하는 최소한의 코드 구현
- **REFACTOR**: 코드 품질 개선 및 @TAG 자동 적용
- 언어별 최적 도구 자동 선택 (pytest, Vitest, JUnit, go test 등)
- TDD 단계별 자동 커밋

**핵심 원칙**: 테스트 없이는 구현 없음

### 3️⃣ /moai:3-sync - 문서 동기화

**담당**: doc-syncer (📖 편집자) + git-manager (🚀 정원사)

- Living Document 자동 갱신
- @TAG 체인 무결성 검증 (ripgrep 코드 스캔)
- API 문서 자동 생성
- PR 상태 Draft → Ready 전환 (Team 모드)
- GitHub 라벨 자동 적용

**핵심 원칙**: 추적성 없이는 완성 없음

---

## 빠른 시작

:::code-group

```bash [Bun (권장)]
# MoAI-ADK 설치
bun add -g moai-adk

# 새 프로젝트 초기화
moai init my-project

# 시스템 진단 (언어 자동 감지)
moai doctor
```

```bash [npm]
# MoAI-ADK 설치
npm install -g moai-adk

# 새 프로젝트 초기화
moai init my-project

# 시스템 진단
moai doctor
```

:::

### 3단계 워크플로우 시작

```bash
# 1. SPEC 작성 (Claude Code에서)
/moai:1-spec "사용자 인증 기능 구현"

# 2. TDD 구현 (언어별 자동 도구 선택)
/moai:2-build SPEC-001

# 3. 문서 동기화 및 TAG 검증
/moai:3-sync
```

---

## 더 알아보기

### 시작하기
- [설치 가이드](/getting-started/installation) - 시스템 요구사항 및 설치 방법
- [빠른 시작](/getting-started/quick-start) - 5분 안에 첫 프로젝트 시작

### 핵심 가이드
- [3단계 워크플로우](/guide/workflow) - SPEC → Build → Sync
- [SPEC-First TDD](/guide/spec-first-tdd) - EARS 방식 명세 작성법
- [TAG 시스템](/guide/tag-system) - CODE-FIRST 추적성 관리

### Claude Code 에이전트
- [에이전트 개요](/claude/agents) - 9개 전문 에이전트 소개
- [워크플로우 명령어](/claude/commands) - /moai: 명령어 상세
- [이벤트 훅](/claude/hooks) - 자동화 시스템

### CLI 명령어
- `moai init` - 프로젝트 초기화
- `moai doctor` - 지능형 시스템 진단
- `moai status` - 프로젝트 상태 확인
- `moai update` - 업데이트 관리
- `moai restore` - 백업 복원

---

**MoAI-ADK v0.0.1** - TypeScript 기반 SPEC-First TDD 개발 도구
*[MoAI Team](https://mo.ai.kr)*
