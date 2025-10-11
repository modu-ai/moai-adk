# Getting Started

MoAI-ADK에 오신 것을 환영합니다! 이 가이드는 5분 안에 MoAI-ADK를 시작할 수 있도록 도와드립니다.

## What is MoAI-ADK?

MoAI-ADK (Modu-AI Agentic Development Kit)는 SPEC-First TDD 방법론을 기반으로 한 범용 개발 도구입니다.

### 핵심 특징

- **SPEC-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **Alfred SuperAgent**: 9명의 전문 에이전트를 조율하는 중앙 오케스트레이터
- **Universal Language Support**: Python, TypeScript, Java, Go, Rust, Ruby, Dart, Swift, Kotlin 등 모든 주요 언어 지원
- **TAG Traceability**: `@SPEC → @TEST → @CODE → @DOC` 완벽한 추적성

---

## Prerequisites

MoAI-ADK를 사용하기 전에 다음 도구들이 설치되어 있어야 합니다:

### 필수 요구사항

| 도구 | 최소 버전 | 권장 버전 | 확인 명령어 |
|------|----------|----------|------------|
| **Node.js** | 18.0.0+ | 20.0.0+ | `node --version` |
| **npm/pnpm/bun** | - | bun 1.2.0+ | `bun --version` |
| **Git** | 2.0+ | 최신 | `git --version` |
| **Claude Code** | - | 최신 | VSCode 확장 |

### 선택 요구사항

- **GitHub CLI**: PR 자동 생성 및 관리 (`gh` 명령어)

---

## Quick Installation

### Global Installation (권장)

::: code-group

```bash [bun]
bun add -g moai-adk
```

```bash [npm]
npm install -g moai-adk
```

```bash [pnpm]
pnpm add -g moai-adk
```

```bash [yarn]
yarn global add moai-adk
```

:::

### Local Installation (프로젝트별)

::: code-group

```bash [bun]
bun add -D moai-adk
```

```bash [npm]
npm install --save-dev moai-adk
```

```bash [pnpm]
pnpm add -D moai-adk
```

```bash [yarn]
yarn add -D moai-adk
```

:::

---

## Verify Installation

설치가 완료되면 다음 명령어로 확인하세요:

```bash
moai --version
# Expected output: v0.2.x
```

도움말 보기:

```bash
moai help
```

---

## Your First MoAI Project

### 1. 프로젝트 초기화

MoAI-ADK는 두 가지 방법으로 초기화할 수 있습니다:

::: code-group

```bash [새 프로젝트]
# 새 프로젝트 생성 (디렉토리 자동 생성)
moai init my-moai-project

# 생성된 디렉토리로 이동
cd my-moai-project
```

```bash [기존 프로젝트]
# 기존 디렉토리에서 초기화
cd existing-project
moai init .
```

:::

대화형 프롬프트가 나타나면 다음 정보를 입력합니다:

```
? Project name: my-moai-project
? Project description: My first MoAI-ADK project
? Development mode: personal (또는 team)
? Primary language: TypeScript (선택)
? Initialize Git? Yes
```

### 2. 프로젝트 구조 확인

초기화가 완료되면 다음과 같은 구조가 생성됩니다:

```
my-moai-project/
├── .moai/
│   ├── config.json          # 프로젝트 설정
│   ├── specs/               # SPEC 문서 저장소
│   ├── reports/             # 동기화 리포트
│   ├── memory/              # 개발 가이드
│   │   ├── development-guide.md
│   │   └── spec-metadata.md
│   └── project/             # 프로젝트 정보
│       ├── product.md
│       ├── structure.md
│       └── tech.md
├── .claude/
│   ├── commands/            # Alfred 커맨드
│   ├── agents/              # 전문 에이전트
│   ├── hooks/               # Git 훅
│   └── output-styles/       # 출력 스타일
└── CLAUDE.md                # 프로젝트 지침
```

### 3. 시스템 진단

프로젝트가 올바르게 설정되었는지 확인합니다:

```bash
moai doctor
```

출력 예시:

```
✅ Git: Installed (v2.39.0)
✅ Node.js: v20.10.0
✅ Package Manager: bun v1.2.19
✅ Claude Code: Available
✅ Project Configuration: Valid

All systems ready! 🚀
```

---

## 3-Stage Development Workflow

MoAI-ADK의 핵심 개발 사이클을 체험해보세요:

### Stage 1: SPEC 작성

```bash
# Claude Code에서 실행
/alfred:1-spec "사용자 인증 기능"
```

**결과**:
- `.moai/specs/SPEC-AUTH-001/spec.md` 생성
- `feature/SPEC-AUTH-001` 브랜치 생성
- Draft PR 생성

### Stage 2: TDD 구현

```bash
/alfred:2-build SPEC-AUTH-001
```

**결과**:
- `@TEST:AUTH-001` 테스트 작성 (RED)
- `@CODE:AUTH-001` 구현 (GREEN)
- 리팩토링 (REFACTOR)

### Stage 3: 문서 동기화

```bash
/alfred:3-sync
```

**결과**:
- Living Document 자동 생성
- TAG 체인 검증 (`@SPEC → @TEST → @CODE → @DOC`)
- PR Ready 전환

---

## What's Next?

축하합니다! 이제 MoAI-ADK를 사용할 준비가 되었습니다. 🎉

### 추천 학습 경로

1. **[Installation Guide](/guides/installation)** - 상세 설치 가이드
2. **[Quick Start Tutorial](/guides/quick-start)** - 실습 튜토리얼
3. **[SPEC-First TDD](/guides/concepts/spec-first-tdd)** - 핵심 개념 이해
4. **[EARS Requirements](/guides/concepts/ears-guide)** - 요구사항 작성법
5. **[TAG System](/guides/concepts/tag-system)** - 추적성 시스템
6. **[TRUST Principles](/guides/concepts/trust-principles)** - 품질 원칙

### 유용한 링크

- [API Reference](/api/index.html)
- [GitHub Repository](https://github.com/modu-ai/moai-adk)
- [Issue Tracker](https://github.com/modu-ai/moai-adk/issues)
- [Changelog](https://github.com/modu-ai/moai-adk/releases)

---

## Need Help?

문제가 발생하면 다음 리소스를 활용하세요:

### Troubleshooting

```bash
# 시스템 진단
moai doctor

# 프로젝트 상태 확인
moai status

# 상세 로그 보기
moai doctor --verbose
```

### Community

- **GitHub Issues**: 버그 리포트 및 기능 요청
- **Discussions**: 질문 및 아이디어 공유

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>Happy Coding with MoAI-ADK!</strong> 🗿</p>
</div>
