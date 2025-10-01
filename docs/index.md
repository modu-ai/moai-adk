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
    details: Claude Code 네이티브 통합. 8개 전문 에이전트, 5개 워크플로우 명령어, 개발 자동화 지원.
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

## 주요 특징

| 항목 | 상세 |
|------|------|
| **패키지 크기** | 195KB (경량화 완료) |
| **빌드 시간** | 182ms (tsup 기반) |
| **테스트 도구** | Vitest 3.2.4 (56개 테스트 중 52개 통과) |
| **코드 품질** | Biome 2.2.4 (린터+포매터 통합) |
| **TAG 시스템** | CODE-FIRST 방식 (ripgrep 직접 스캔) |
| **CLI 명령어** | 7개 명령어 (init, doctor, status, update, restore 등) |

## 핵심 원칙: TRUST 5원칙

- **T**est First: 테스트 없이는 코드 없음 (TDD 엄격 적용)
- **R**eadable: 요구사항 주도 가독성 (SPEC 기반 코드)
- **U**nified: SPEC 기반 아키텍처 (언어별 일관된 구조)
- **S**ecured: 설계 시점 보안 (입력 검증, 민감정보 마스킹)
- **T**rackable: CODE-FIRST TAG 추적성 (@SPEC → @TEST → @CODE → @DOC)

## 링크 및 리소스

- **공식 문서**: [https://adk.mo.ai.kr](https://adk.mo.ai.kr)
- **커뮤니티**: [https://mo.ai.kr](https://mo.ai.kr) *(오픈 예정)*
- **GitHub**: [github.com/modu-ai/moai-adk](https://github.com/modu-ai/moai-adk)
- **NPM Package**: [npmjs.com/package/moai-adk](https://www.npmjs.com/package/moai-adk)

## 왜 MoAI-ADK인가?

### TypeScript 기반, 8개 언어 지원
MoAI-ADK는 TypeScript로 구축된 CLI 도구입니다. 사용자 프로젝트는 TypeScript, Python, Java, Go, Rust, C++, C#, PHP 총 8개 언어를 지원합니다. 프로젝트 파일을 분석하여 언어를 감지하고, 해당 언어의 TDD 도구(pytest, Vitest, JUnit, go test 등)를 자동 추천합니다.

### CODE-FIRST TAG 시스템
중간 캐시를 사용하지 않습니다. TAG의 진실은 오직 코드 자체에만 존재하며, ripgrep으로 소스코드를 직접 스캔하여 실시간 추적성을 보장합니다.

### 시스템 진단 자동화
`moai doctor` 명령은 프로젝트 디렉토리를 분석하여 사용 중인 언어를 감지하고, 해당 언어에 필요한 개발 도구 설치 여부를 확인합니다. Runtime(Node.js, Git), Development(npm, TypeScript), Language-Specific(언어별 도구) 등의 카테고리로 체계적인 진단을 제공합니다.

### Claude Code 통합
8개 전문 에이전트(`spec-builder`, `code-builder`, `doc-syncer`, `git-manager`, `debug-helper`, `cc-manager`, `trust-checker`, `tag-agent`)가 SPEC-First TDD 워크플로우를 자동화합니다. 명세 작성부터 TDD 구현, 문서 동기화, Git 작업, 품질 검증까지 지원합니다.

### 개발 스택
- **TypeScript 5.9.2**: 엄격한 타입 검사
- **Bun 1.2.19**: 빠른 패키지 관리 (npm 대비 약 10-20배, Bun 공식 벤치마크 기준)
- **Vitest 3.2.4**: 테스트 자동화 (56개 테스트 중 52개 통과)
- **Biome 2.2.4**: 린터+포매터 통합 (ESLint+Prettier 대체)
- **tsup 8.5.0**: 빌드 182ms, ESM/CJS 듀얼 번들링

## 더 알아보기

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

## 기여하기

MoAI-ADK는 MIT 라이선스 오픈소스 프로젝트입니다. 기여를 환영합니다!

- [버그 리포트](https://github.com/modu-ai/moai-adk/issues) - 문제 발견 시 이슈 등록
- [기능 제안](https://github.com/modu-ai/moai-adk/discussions) - 새로운 아이디어 공유
- [문서 개선](https://github.com/modu-ai/moai-adk/pulls) - Pull Request 제출

---

**MoAI-ADK v0.0.1** - TypeScript 기반 SPEC-First TDD 개발 도구
*[MoAI Team](https://mo.ai.kr)*