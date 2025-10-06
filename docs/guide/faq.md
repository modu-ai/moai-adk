# @CODE:DOCS-001:UI | SPEC: .moai/specs/SPEC-DOCS-001/spec.md

# FAQ (자주 묻는 질문)

MoAI-ADK 사용 중 자주 묻는 질문과 답변을 모았습니다.

## 🤔 기본 개념

### MoAI-ADK란 무엇인가요?

MoAI-ADK(Agentic Development Kit)는 **AI 시대의 코드 품질 문제를 해결하는 SPEC-First TDD 개발 프레임워크**입니다. Alfred SuperAgent와 9개 전문 에이전트로 구성된 10개 AI 에이전트 팀이 SPEC 작성부터 TDD 구현, 문서 동기화까지 전 과정을 자동화합니다.

### Alfred는 누구인가요?

**Alfred**는 MoAI-ADK의 SuperAgent이자 중앙 오케스트레이터 AI입니다. 배트맨의 집사 Alfred Pennyworth에서 영감을 받아 이름을 지었으며, 9개 전문 에이전트(spec-builder, code-builder, doc-syncer 등)를 조율하여 완벽한 품질의 코드를 생산합니다.

### 왜 SPEC-First인가요?

**"명세 없으면 코드 없다"**는 철학을 따릅니다. SPEC을 먼저 작성하면:

- ✅ 무엇을 만들지 명확히 정의
- ✅ 팀원 간 요구사항 공유
- ✅ 6개월 후에도 "왜"를 찾을 수 있음
- ✅ AI가 생성한 코드의 목적 추적 가능

코드를 먼저 작성하면 나중에 "왜 이렇게 만들었지?"라는 질문에 답할 수 없습니다.

### TDD는 왜 필요한가요?

**"테스트 없으면 구현 없다"**는 원칙입니다. TDD(Test-Driven Development)는:

- ✅ 코드가 요구사항을 만족하는지 자동 검증
- ✅ 리팩토링 시 기존 동작 보장
- ✅ 엣지 케이스 사전 발견
- ✅ 테스트 커버리지 ≥85% 자동 달성

AI가 생성한 코드는 아름답지만 작동하지 않을 수 있습니다. TDD는 이를 방지합니다.

## 🚀 시작하기

### 설치가 복잡한가요?

전혀 복잡하지 않습니다! 3단계면 충분합니다:

```bash
# 1. MoAI-ADK 설치
bun add -g moai-adk

# 2. 프로젝트 초기화
moai init my-project

# 3. Alfred 활성화 (Claude Code에서)
/alfred:8-project
```

### Claude Code가 필수인가요?

현재는 **Claude Code 전용**으로 설계되었습니다. 하지만 핵심 개념(SPEC-First TDD, @TAG 시스템)은 다른 AI 도구에서도 적용 가능합니다. 향후 다른 플랫폼 지원도 검토 중입니다.

### 기존 프로젝트에도 적용할 수 있나요?

네! 기존 프로젝트에 MoAI-ADK를 추가할 수 있습니다:

```bash
cd existing-project
moai init .
```

Alfred가 기존 코드 구조를 분석하고 최적의 설정을 제안합니다.

## 🏷️ @TAG 시스템

### @TAG는 무엇인가요?

**@TAG**는 코드 추적성을 보장하는 핵심 시스템입니다. 모든 코드 조각을 다음과 같이 연결합니다:

```
@SPEC:AUTH-001 → @TEST:AUTH-001 → @CODE:AUTH-001 → @DOC:AUTH-001
```

6개월 후에도 "왜 이 코드를 이렇게 만들었는지" 추적할 수 있습니다.

### CODE-FIRST 원칙이란?

**"TAG의 진실은 코드 자체에만 존재"**한다는 철학입니다.

- ❌ 별도 데이터베이스나 YAML 파일에 TAG 저장 안 함
- ✅ 코드를 직접 스캔하여 TAG 추출 (`rg '@TAG' -n`)
- ✅ 코드 변경 시 TAG도 함께 변경
- ✅ 코드와 문서가 따로 놀 수 없음

### TAG를 수동으로 관리해야 하나요?

아니요! Alfred가 **자동으로** TAG를 생성하고 관리합니다:

- `/alfred:1-spec` → `@SPEC:ID` 자동 생성
- `/alfred:2-build` → `@TEST:ID`, `@CODE:ID` 자동 적용
- `/alfred:3-sync` → TAG 체인 검증 및 `@DOC:ID` 생성

여러분은 TAG 문법만 이해하면 됩니다.

## 🛠️ 워크플로우

### 3단계 워크플로우가 너무 복잡하지 않나요?

오히려 **단순합니다**! 모든 개발 작업이 3단계로 표준화됩니다:

1. `/alfred:1-spec "기능 설명"` - SPEC 작성
2. `/alfred:2-build SPEC-ID` - TDD 구현
3. `/alfred:3-sync` - 문서 동기화

어떤 기능을 만들든 이 3단계만 반복하면 됩니다.

### Personal 모드와 Team 모드의 차이는?

| 구분 | Personal 모드 | Team 모드 |
|------|--------------|-----------|
| **대상** | 혼자 개발 | 팀 협업 |
| **Git 전략** | 로컬 브랜치 | GitFlow (feature/develop/main) |
| **PR 관리** | 없음 | Draft PR → Ready → Auto Merge |
| **커밋** | 자동 커밋 | 단계별 커밋 |

`.moai/config.json`에서 `mode: "personal"` 또는 `"team"` 설정 가능합니다.

### GitFlow를 모르는데 사용할 수 있나요?

네! Alfred가 **자동으로** GitFlow를 적용합니다:

- `feature/SPEC-ID` 브랜치 자동 생성
- Draft PR 자동 생성 (Team 모드)
- 코드 리뷰 후 자동 머지

여러분은 `/alfred:1-spec`, `/alfred:2-build`, `/alfred:3-sync`만 실행하면 됩니다.

## 🌍 언어 지원

### 어떤 프로그래밍 언어를 지원하나요?

**모든 주요 프로그래밍 언어**를 지원합니다:

- **백엔드**: Python, TypeScript, Java, Go, Rust
- **프론트엔드**: TypeScript, JavaScript (React, Vue, Angular)
- **모바일**: Dart (Flutter), Swift (iOS), Kotlin (Android)

Alfred가 프로젝트 언어를 자동 감지하고 최적의 도구 체인을 선택합니다.

### Python 프로젝트에도 사용할 수 있나요?

네! Python 프로젝트에서는 자동으로:

- **테스트**: pytest
- **린터**: ruff
- **타입 체커**: mypy
- **포매터**: black

이 활성화됩니다. 설정 파일도 Alfred가 자동 생성합니다.

### TypeScript 전용인가요?

아니요! MoAI-ADK는 **TypeScript로 구축**되었지만, **모든 언어를 지원**합니다. TypeScript, Python, Go, Rust 등 각 언어에 최적화된 TDD 워크플로우를 제공합니다.

## 🔧 문제 해결

### 테스트 커버리지가 85%가 안 되면 어떻게 되나요?

Alfred가 **자동으로 경고**하고 추가 테스트를 제안합니다:

```
⚠️ 테스트 커버리지 부족: 72% (목표: 85%)
추가 테스트 필요:
- 엣지 케이스: 빈 배열 처리
- 에러 케이스: 네트워크 타임아웃
```

Alfred가 제안하는 테스트를 추가하면 쉽게 85%를 달성할 수 있습니다.

### TRUST 원칙을 위반하면 어떻게 되나요?

Alfred가 **즉시 중단**하고 구체적인 개선 제안을 제공합니다:

```
❌ TRUST 원칙 위반:
- Simplicity: 함수 길이 72줄 (제한: 50줄)
- Secured: SQL Injection 취약점 발견

✅ 개선 제안:
- processUserData() 함수를 3개로 분리하세요
- Prepared Statement를 사용하세요
```

### 에러가 발생하면 어떻게 하나요?

**debug-helper 에이전트**를 호출하세요:

```
@agent-debug-helper "TypeError: 'NoneType' object has no attribute 'name'"
```

debug-helper가 에러를 분석하고 해결 방법을 제시합니다.

## 💡 고급 기능

### 10개 AI 에이전트 팀은 어떻게 구성되나요?

Alfred(SuperAgent) + 9개 전문 에이전트:

| 에이전트 | 역할 | 호출 시점 |
|---------|------|----------|
| **spec-builder** | SPEC 작성 | `/alfred:1-spec` |
| **code-builder** | TDD 구현 | `/alfred:2-build` |
| **doc-syncer** | 문서 동기화 | `/alfred:3-sync` |
| **tag-agent** | TAG 관리 | 자동 |
| **git-manager** | Git 워크플로우 | 자동 |
| **debug-helper** | 디버깅 | `@agent-debug-helper` |
| **trust-checker** | 품질 검증 | 자동 |
| **cc-manager** | 설정 관리 | 자동 |
| **project-manager** | 프로젝트 초기화 | `/alfred:8-project` |

### EARS 문법이 어려운데 배워야 하나요?

Alfred가 **자동으로 변환**해주므로 배울 필요 없습니다:

```
여러분: "JWT 로그인 기능"

Alfred가 자동 변환:
- Ubiquitous: 시스템은 JWT 인증을 제공해야 한다
- Event-driven: WHEN 로그인하면 토큰 발급해야 한다
- Constraints: IF 토큰 만료되면 401 반환해야 한다
```

하지만 EARS 문법을 알면 더 정확한 SPEC을 작성할 수 있습니다.

## 🆘 지원

### 버그를 발견했어요!

[GitHub Issues](https://github.com/modu-ai/moai-adk/issues)에 버그 리포트를 작성해주세요. 100% AI로 만든 프로젝트이므로 피드백이 매우 중요합니다.

### 기능 요청을 하고 싶어요!

[GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)에서 아이디어를 공유해주세요. 커뮤니티와 함께 더 나은 도구로 만들어갑니다.

### 더 많은 도움이 필요해요!

- **문서**: [MoAI-ADK 공식 문서](https://moai-adk.vercel.app)
- **가이드**: [Quick Start](/guide/getting-started)
- **개념**: [SPEC 우선 TDD](/concepts/spec-first-tdd)

---

질문이 해결되지 않았나요? [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)에서 질문하세요!
