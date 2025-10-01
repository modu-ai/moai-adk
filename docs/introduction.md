---
title: MoAI-ADK 소개
description: TypeScript 기반 SPEC-First TDD 개발 도구
---

# MoAI-ADK 소개

> **명세 없이는 코드 없음. 테스트 없이는 구현 없음. 추적성 없이는 완성 없음.**

MoAI-ADK는 Claude Code 환경에서 **SPEC-First TDD 개발**을 자동화하는 Agentic Development Kit입니다. TypeScript 기반으로 구축되어 TypeScript, Python, Java, Go, Rust, C++, C#, PHP 총 8개 언어를 지원하며, 일관된 개발 경험을 제공합니다.

## MoAI-ADK란?

### 해결하는 문제

현대 소프트웨어 개발에는 다음과 같은 문제들이 존재합니다:

1. **AI 페어 프로그래밍의 체계 부재**: Claude Code 같은 AI 도구를 사용하지만 일관된 방법론이 없음
2. **언어별 도구 파편화**: 각 언어마다 다른 개발 도구와 워크플로우
3. **추적성 관리의 복잡성**: 요구사항부터 코드까지 연결 고리 관리 어려움
4. **문서-코드 불일치**: 개발 진행에 따라 문서가 오래됨

### MoAI-ADK의 해결책

MoAI-ADK는 **SPEC-First TDD 자동화 개발 도구**로 다음을 제공합니다:

- **3단계 워크플로우**: SPEC 작성 → TDD 구현 → 문서 동기화
- **8개 전문 에이전트**: 각 단계를 자동화하는 AI 에이전트
- **CODE-FIRST TAG 추적성**: 요구사항부터 구현까지 추적성 제공
- **8개 언어 지원**: TypeScript, Python, Java, Go, Rust, C++, C#, PHP

## 핵심 개념 3가지

### 1. SPEC-First: 명세 없이는 코드 없음

**EARS 방법론**을 활용한 체계적인 요구사항 작성:

```markdown
### Ubiquitous Requirements
- 시스템은 사용자 인증 기능을 제공해야 한다

### Event-driven Requirements
- WHEN 사용자가 로그인하면, 시스템은 JWT 토큰을 발급해야 한다

### State-driven Requirements
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다

### Constraints
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
```

**왜 SPEC-First인가?**

- **명확한 계약**: 구현 전에 무엇을 만들지 정의
- **커뮤니케이션 향상**: 팀 간 공통 언어
- **변경 추적**: 요구사항 변경 이력 관리
- **AI 친화적**: Claude Code가 SPEC을 기반으로 정확히 구현

### 2. TDD-First: 테스트 없이는 구현 없음

**Red-Green-Refactor 사이클**:

```typescript
// @TEST:AUTH-001: 유효한 사용자 인증 테스트
describe('AuthService', () => {
  test('should authenticate valid user', async () => {
    // RED: 실패하는 테스트 작성
    const service = new AuthService();
    const result = await service.authenticate('user', 'pass');
    expect(result.token).toBeDefined();
  });
});

// GREEN: 최소한의 코드로 통과
class AuthService {
  async authenticate(username: string, password: string) {
    // 구현...
    return { token: 'jwt-token' };
  }
}

// REFACTOR: 품질 개선
class AuthService {
  constructor(private tokenService: TokenService) {}

  async authenticate(username: string, password: string) {
    // 리팩토링된 구현...
  }
}
```

**언어별 TDD 지원**:

- **TypeScript**: Vitest + strict typing
- **Python**: pytest + mypy
- **Java**: JUnit + Maven/Gradle
- **Go**: go test + table-driven tests
- **Rust**: cargo test + doc tests

### 3. TAG-First: 추적성 없이는 완성 없음

**CODE-FIRST TAG 시스템**으로 요구사항부터 코드까지 추적성 제공:

```typescript
// @CODE:AUTH-001 | Chain: @SPEC:AUTH-001 →  → @CODE:AUTH-001 → @TEST:AUTH-001
// Related: @CODE:AUTH-001:API, @CODE:AUTH-001:UI, @CODE:AUTH-001:DATA

/**
 * @CODE:AUTH-001:API: 사용자 인증 API
 */
class AuthService {
  /**
   * @CODE:AUTH-001: 입력값 검증 및 인증 처리
   */
  async authenticate(username: string, password: string): Promise<AuthResult> {
    // 구현...
  }
}
```

**TAG 체계**:

- **TAG 흐름**: @SPEC → @TEST → @CODE → @DOC (필수)
- **@CODE 서브카테고리**: @CODE 서브카테고리 (API, UI, DATA 등) (필수)

**CODE-FIRST 철학**:
- TAG의 진실은 오직 코드 자체에만 존재
- 중간 캐시 없이 ripgrep으로 직접 스캔
- 실시간 검증

## 아키텍처 다이어그램

```
MoAI-ADK Architecture
┌─────────────────────────────────────────────────┐
│         TypeScript CLI & Core Engine            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │   init   │ │  doctor  │ │  status  │  ...  │
│  └──────────┘ └──────────┘ └──────────┘       │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│      Universal Language Support                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │ TypeScript│ │  Python  │ │   Java   │  ...  │
│  └──────────┘ └──────────┘ └──────────┘       │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│        Claude Code Integration                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐       │
│  │  Agents  │ │ Commands │ │   Hooks  │       │
│  └──────────┘ └──────────┘ └──────────┘       │
└─────────────────────────────────────────────────┘
```

## TypeScript 전환 (SPEC-013)

### 전환 전후 비교

#### Before (Python 기반)

```
복잡한 아키텍처:
MoAI-ADK (Python) ↔ TypeScript 브릿지 ↔ 사용자 프로젝트

- 15MB 패키지 크기
- 4.6초 빌드 시간
- Python + TypeScript 이중 의존성
```

#### After (TypeScript 기반)

```
단순한 아키텍처:
MoAI-ADK (TypeScript) → 언어별 TDD 도구 → 사용자 프로젝트

- 195KB 패키지 크기
- 182ms 빌드 시간
- Node.js 단일 런타임
```

### 주요 개선 사항

| 항목 | Before | After |
|------|--------|-------|
| 패키지 크기 | 15MB | 195KB |
| 빌드 시간 | 4.6초 | 182ms |
| 테스트 도구 | 제한적 | Vitest (56개 중 52개 통과) |
| 런타임 | Python+Node.js | Node.js |
| 언어 지원 | 제한적 | 8개 언어 |

## 주요 특징

### 성능

- **빌드 시간**: 182ms (tsup 기반)
- **패키지 크기**: 195KB
- **테스트**: Vitest (56개 중 52개 통과)
- **린터**: Biome (ESLint+Prettier 통합)

### TRUST 5원칙

- **T** (Test First): TDD 엄격 적용
- **R** (Readable): SPEC 기반 가독성
- **U** (Unified): 언어별 일관된 구조
- **S** (Secured): 설계 시점 보안
- **T** (Trackable): CODE-FIRST TAG 추적성

### 언어 지원

- **8개 언어**: TypeScript, Python, Java, Go, Rust, C++, C#, PHP
- 프로젝트 파일 분석 기반 언어 감지
- 언어별 최적 테스트 프레임워크 자동 추천

## 다음 단계

### 시작하기

1. **[설치](/getting-started/installation)**: 시스템 요구사항 확인 및 설치
2. **[빠른 시작](/getting-started/quick-start)**: 5분 안에 첫 기능 구현
3. **[프로젝트 초기화](/getting-started/project-setup)**: 프로젝트 구조 이해

### 핵심 개념 학습

1. **[SPEC-First TDD](/concepts/spec-first-tdd)**: 방법론 완전 가이드
2. **[TAG 시스템](/concepts/tag-system)**: CODE-FIRST 추적성 관리 방법
3. **[3단계 워크플로우](/concepts/workflow)**: 개발 사이클 이해
4. **[TRUST 원칙](/concepts/trust-principles)**: 품질 기준 학습

### Claude Code 통합

1. **[에이전트 가이드](/claude/agents)**: 7개 전문 에이전트 활용
2. **[워크플로우 명령어](/claude/commands)**: 5개 핵심 명령어 사용
3. **[이벤트 훅](/claude/hooks)**: 8개 자동화 시스템 이해

## 참고 자료

- **GitHub**: [MoAI-ADK Repository](https://github.com/modu-ai/moai-adk)
- **NPM**: [moai-adk](https://www.npmjs.com/package/moai-adk)
- **커뮤니티**: [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)