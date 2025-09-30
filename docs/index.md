---
layout: home

hero:
  name: "MoAI-ADK"
  text: "모두의 ADK / Agentic Development Kit"
  tagline: Claude Code 기반 범용 언어 지원 개발 툴킷
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
    details: Red-Green-Refactor 사이클 강제 적용. 테스트 없이는 구현 없음. 92.9% 테스트 성공률 (Vitest).
  - icon:
      src: /icons/tag.svg
      alt: TAG
    title: CODE-FIRST TAG 시스템
    details: 요구사항부터 구현까지 완전한 추적성 제공. 중간 캐시 없이 소스코드 직접 스캔으로 94% 최적화 달성.
  - icon:
      src: /icons/language.svg
      alt: Languages
    title: 범용 언어 지원
    details: TypeScript, Python, Java, Go, Rust 등 지원. 지능형 언어 감지 및 자동 도구 매핑.
  - icon:
      src: /icons/performance.svg
      alt: Performance
    title: 초고속 성능
    details: Bun으로 98% 빠른 패키지 관리. 빌드 182ms, TAG 로딩 < 50ms. Biome 94.8% 성능 향상.
  - icon:
      src: /icons/claude.svg
      alt: Claude
    title: Claude Code 완전 통합
    details: Claude Code 네이티브 통합. 7개 전문 에이전트, 5개 워크플로우 명령어, 8개 이벤트 훅.
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

## 📊 현대화 성과 (v0.0.1)

| 지표 | 성과 |
|------|------|
| **패키지 크기** | 99% 절감 (15MB → 195KB) |
| **빌드 시간** | 96% 단축 (4.6초 → 182ms) |
| **테스트 성공률** | 92.9% (Vitest 3.2.4) |
| **코드 품질** | 94.8% 성능 향상 (Biome) |
| **TAG 시스템** | CODE-FIRST 방식으로 94% 최적화 |
| **CLI 완성도** | 100% (7개 명령어 완성) |

## 🎨 핵심 원칙: TRUST 5원칙

- **T**est First: 테스트 없이는 코드 없음 (TDD 강제)
- **R**eadable: 요구사항 주도 가독성 (SPEC 기반 코드)
- **U**nified: 통합 SPEC 아키텍처 (언어 무관 일관성)
- **S**ecured: SPEC 준수 보안 (설계 시점 보안)
- **T**rackable: SPEC 추적성 (CODE-FIRST TAG 시스템)

## 🌐 링크 및 리소스

- **📚 공식 문서**: [https://adk.mo.ai.kr](https://adk.mo.ai.kr)
- **💬 커뮤니티**: [https://mo.ai.kr](https://mo.ai.kr) *(오픈 예정)*
- **🐙 GitHub**: [github.com/modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)
- **📦 NPM Package**: [npmjs.com/package/moai-adk](https://www.npmjs.com/package/moai-adk)

## 💡 왜 MoAI-ADK인가?

### TypeScript 단일 스택, 모든 언어 지원
MoAI-ADK 자체는 TypeScript로 구축된 고성능 CLI 도구입니다. 하지만 사용자 프로젝트는 Python, JavaScript, Java, Go, Rust, C++, C# 등 모든 주요 언어를 지원합니다. 프로젝트 언어를 자동으로 감지하고, 해당 언어에 최적화된 TDD 도구(pytest, Vitest, JUnit, go test 등)를 동적으로 추천합니다.

### CODE-FIRST TAG 시스템 (94% 최적화)
SQLite3와 모든 중간 캐시를 완전히 제거했습니다. TAG의 진실은 오직 코드 자체에만 존재하며, ripgrep으로 코드를 직접 스캔하여 실시간 추적성을 보장합니다. 로딩 성능 50ms 이하, 저장공간 94% 절감을 달성했습니다.

### 지능형 시스템 진단
`moai doctor` 명령은 단순히 도구 설치 여부만 확인하지 않습니다. 프로젝트 디렉토리를 분석하여 사용 중인 언어를 자동으로 감지하고, 해당 언어에 필요한 개발 도구를 동적으로 추가합니다. Runtime(Node.js, Git), Development(npm, TypeScript), Optional(Docker, GitHub CLI), Language-Specific(언어별 도구), Performance(디스크 I/O) 등 5개 카테고리로 체계적인 진단을 제공합니다.

### Claude Code 완전 통합
7개 전문 에이전트(`spec-builder`, `code-builder`, `doc-syncer`, `git-manager`, `debug-helper`, `cc-manager`, `trust-checker`)가 SPEC-First TDD 워크플로우의 각 단계를 자동화합니다. 명세 작성부터 TDD 구현, 문서 동기화, Git 작업, 품질 검증까지 모든 과정이 에이전트의 지원을 받습니다.

### 현대적 개발 스택
- **TypeScript 5.9.2**: 엄격한 타입 검사와 최신 언어 기능
- **Bun 1.2.19**: npm 대비 98% 빠른 패키지 관리
- **Vitest 3.2.4**: 92.9% 테스트 성공률을 달성한 고속 테스트 러너
- **Biome 2.2.4**: ESLint + Prettier를 통합하여 94.8% 성능 향상
- **tsup 8.5.0**: 182ms 초고속 빌드, ESM/CJS 듀얼 번들링

## 📚 더 알아보기

### 시작하기
- [설치 가이드](/getting-started/installation) - 시스템 요구사항 및 설치 방법
- [빠른 시작](/getting-started/quick-start) - 5분 안에 첫 프로젝트 시작

### 핵심 가이드
- [3단계 워크플로우](/guide/workflow) - SPEC → Build → Sync
- [SPEC-First TDD](/guide/spec-first-tdd) - EARS 방식 명세 작성법
- [TAG 시스템](/guide/tag-system) - CODE-FIRST 추적성 관리

### CLI 명령어
- `moai init` - 프로젝트 초기화
- `moai doctor` - 지능형 시스템 진단
- `moai status` - 프로젝트 상태 확인
- `moai update` - 업데이트 관리
- `moai restore` - 백업 복원

## 🤝 기여하기

MoAI-ADK는 MIT 라이선스 오픈소스 프로젝트입니다. 기여를 환영합니다!

- 🐛 [버그 리포트](https://github.com/modu-ai/moai-adk/issues) - 문제 발견 시 이슈 등록
- 💡 [기능 제안](https://github.com/modu-ai/moai-adk/discussions) - 새로운 아이디어 공유
- 📝 [문서 개선](https://github.com/modu-ai/moai-adk/pulls) - Pull Request 제출

---

**MoAI-ADK v0.0.1** - TypeScript 기반 고성능 SPEC-First TDD 프레임워크
*Made with ❤️ by [MoAI Team](https://mo.ai.kr)*