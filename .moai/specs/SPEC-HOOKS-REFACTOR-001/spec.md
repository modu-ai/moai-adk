---
id: HOOKS-REFACTOR-001
version: 0.0.1
status: draft
created: 2025-10-11
updated: 2025-10-11
author: @Goos
priority: high
category: refactoring
labels: ["hooks", "DRY", "multilang", "code-quality"]
depends_on: []
blocks: []
related_specs: []
scope: "Claude Code Hooks 리팩토링 - DRY 원칙 적용 및 다국어 지원"
---

# @SPEC:HOOKS-REFACTOR-001: Claude Code Hooks 리팩토링

## HISTORY

### v0.0.1 (2025-10-11)
- **INITIAL**: Claude Code Hooks 리팩토링 SPEC 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: CLI Entry Point 중복 제거, 하드코딩 상수 분리, 다국어 지원 확장
- **CONTEXT**: hooks 코드 분석 결과 100 LOC 중복 및 7/11 언어만 지원하는 문제 발견
- **TARGET FILES**:
  - 리팩토링: policy-block.ts, pre-write-guard.ts, tag-enforcer.ts, session-notice/index.ts
  - 신규 생성: base.ts, constants.ts, utils.ts
- **EXPECTED IMPACT**:
  - 코드 중복 100% 제거
  - 언어 지원 63.6% → 100% (7개 → 11개)
  - 유지보수성 3배 향상

---

## Environment (환경)

### 시스템 환경
- **프로젝트**: MoAI-ADK (TypeScript)
- **모드**: Team (GitFlow with develop 브랜치)
- **현재 브랜치**: feature/SPEC-DOCS-002
- **작업 브랜치**: feature/SPEC-HOOKS-REFACTOR-001 (신규 생성 예정)
- **타겟 디렉토리**: `moai-adk-ts/src/claude/hooks/`
- **빌드 도구**: tsup (TypeScript → CommonJS 컴파일)
- **배포 경로**: `templates/.claude/hooks/alfred/*.cjs`

### 기술 스택
- **언어**: TypeScript 5.9.2+
- **런타임**: Node.js 18+ / Bun 1.2+
- **테스트**: Vitest 3.2+
- **린터**: Biome 2.2+
- **타입 체커**: TypeScript strict mode

### 현재 코드 상태
- **총 LOC**: 3,506 LOC (소스 2,207 LOC + 테스트 1,299 LOC)
- **훅 파일**: 4개 (policy-block, pre-write-guard, tag-enforcer, session-notice)
- **테스트 커버리지**: 76개 테스트 (17 + 14 + 18 + 12 + 15)
- **코드 중복**: 약 100 LOC (CLI Entry Point)
- **하드코딩 상수**: 3개 영역 (파일 확장자, 도구 목록, 매직 넘버)

---

## Assumptions (가정)

### 기술적 가정
1. **후방 호환성 보장**: 기존 훅의 동작과 인터페이스는 절대 변경하지 않는다
2. **테스트 우선**: 모든 리팩토링은 TDD(RED-GREEN-REFACTOR) 방식으로 진행한다
3. **성능 유지**: Fast-Track 패턴 및 조기 리턴 최적화는 그대로 유지한다
4. **타입 안전성**: `Record<string, any>` 대신 구체적 타입을 사용한다

### 비즈니스 가정
1. **언어 지원 우선순위**: MoAI-ADK product.md에 명시된 "모든 주요 언어"를 지원한다
2. **코드 품질**: TRUST 5원칙(Test, Readable, Unified, Secured, Trackable)을 준수한다
3. **유지보수성**: 새로운 언어 추가 시 1개 파일(constants.ts)만 수정하면 된다

### 제약사항
1. **빌드 도구 의존성**: tsup 설정(tsup.hooks.config.ts)은 수정하지 않는다
2. **배포 형식**: CommonJS (.cjs) 형식은 유지한다 (Claude Code 호환성)
3. **디렉토리 구조**: `src/claude/hooks/` 구조는 변경하지 않는다

---

## Requirements (요구사항)

### Ubiquitous Requirements (필수 기능)

#### UR-001: CLI Entry Point 통합
- 시스템은 모든 훅에서 공통 CLI Entry Point 로직을 재사용해야 한다
- `runHook(HookClass)` 유틸리티 함수를 제공해야 한다
- 기존 4개 훅 파일(policy-block, pre-write-guard, tag-enforcer, session-notice)의 중복 코드를 제거해야 한다

**근거**: 100 LOC 중복 코드 제거, 유지보수성 향상

#### UR-002: 하드코딩 상수 중앙 관리
- 시스템은 모든 하드코딩된 상수를 `constants.ts` 파일에서 중앙 집중식으로 관리해야 한다
- 파일 확장자, 도구 목록, 타임아웃 값을 상수로 export해야 한다
- 타입 안전성을 보장하기 위해 `as const` assertion을 사용해야 한다

**근거**: 새 언어 추가 시 1개 파일만 수정, 타입 안전성 보장

#### UR-003: 다국어 파일 확장자 지원
- 시스템은 MoAI-ADK가 지원하는 모든 언어의 파일 확장자를 인식해야 한다
- Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin, Ruby, PHP, C#, Elixir를 지원해야 한다
- `SUPPORTED_LANGUAGES` 상수에 모든 언어와 확장자를 정의해야 한다

**근거**: product.md의 "모든 주요 언어" 지원 약속 이행

#### UR-004: 공통 유틸리티 함수 제공
- 시스템은 `extractFilePath()` 공통 함수를 제공해야 한다
- 시스템은 `extractCommand()` 공통 함수를 제공해야 한다
- tag-enforcer, pre-write-guard에서 중복된 로직을 제거해야 한다

**근거**: DRY 원칙, 일관성 있는 입력 처리

---

### Event-driven Requirements (이벤트 기반)

#### ER-001: 언어 추가 시 자동 반영
- WHEN 새로운 언어 지원이 추가되면(constants.ts 수정), 시스템은 모든 훅에 자동으로 반영되어야 한다
- WHEN constants.ts의 `SUPPORTED_LANGUAGES`에 새 언어를 추가하면, tag-enforcer가 즉시 해당 확장자를 인식해야 한다
- WHEN 새 도구가 READ_ONLY_TOOLS에 추가되면, policy-block이 자동으로 Fast-Track 처리해야 한다

**근거**: 중앙 집중식 관리의 핵심 가치

#### ER-002: 훅 CLI 실행 시 표준화 처리
- WHEN 훅을 CLI에서 직접 실행하면(`node policy-block.cjs`), 시스템은 `runHook()` 유틸리티를 통해 표준화된 처리를 수행해야 한다
- WHEN `require.main === module` 조건이 만족되면, 시스템은 `runHook(HookClass)`를 자동 호출해야 한다
- WHEN 훅 실행 중 오류가 발생하면, 시스템은 일관된 형식의 오류 메시지를 출력해야 한다

**근거**: 일관성 있는 CLI 인터페이스

#### ER-003: 타입 오류 발생 시 컴파일 차단
- WHEN constants.ts의 타입 정의가 변경되면, 시스템은 TypeScript 컴파일러를 통해 타입 안전성을 검증해야 한다
- WHEN 잘못된 타입으로 constants를 사용하면, 시스템은 컴파일 타임 오류를 발생시켜야 한다
- WHEN Biome 린터가 경고를 발견하면, 시스템은 빌드를 차단해야 한다

**근거**: 타입 안전성 보장, 런타임 오류 사전 차단

---

### State-driven Requirements (상태 기반)

#### SR-001: 리팩토링 중 테스트 통과 유지
- WHILE 코드 리팩토링 중일 때, 시스템은 기존 76개 테스트를 100% 통과해야 한다
- WHILE 새 파일(base.ts, constants.ts, utils.ts)을 작성할 때, 시스템은 기존 훅의 동작을 절대 변경하지 않아야 한다
- WHILE TDD RED 단계일 때, 시스템은 실패하는 테스트를 먼저 작성해야 한다

**근거**: TDD 원칙, 후방 호환성 보장

#### SR-002: constants.ts 수정 중 타입 안전성 보장
- WHILE constants.ts를 수정 중일 때, 시스템은 `as const` assertion을 통해 리터럴 타입을 보장해야 한다
- WHILE 새 상수를 추가할 때, 시스템은 export된 타입이 readonly임을 보장해야 한다
- WHILE 기존 상수를 수정할 때, 시스템은 모든 사용처에서 타입 오류가 발생하지 않음을 확인해야 한다

**근거**: 타입 안전성, 실수 방지

#### SR-003: 빌드 중 배포 파일 생성
- WHILE `npm run build:hooks` 실행 중일 때, 시스템은 TypeScript 소스를 CommonJS로 컴파일해야 한다
- WHILE tsup 빌드 중일 때, 시스템은 `templates/.claude/hooks/alfred/*.cjs` 경로에 실행 파일을 생성해야 한다
- WHILE 빌드 완료 후, 시스템은 각 `.cjs` 파일이 독립 실행 가능함을 보장해야 한다

**근거**: 배포 프로세스 유지, Claude Code 호환성

---

### Optional Features (선택적 기능)

#### OF-001: 정규식 컴파일 최적화
- WHERE 성능 최적화가 필요하면, 시스템은 tag-validator.ts에 `parseTagBlock()` 메서드를 추가할 수 있다
- WHERE TAG 검증 시간이 100ms를 초과하면, 시스템은 정규식을 한 번만 실행하도록 최적화할 수 있다
- WHERE 벤치마크 테스트가 필요하면, 시스템은 `policy-block-benchmark.test.ts` 패턴을 따를 수 있다

**근거**: 성능 향상 여지 남김, 필수는 아님

#### OF-002: JSDoc 자동 생성
- WHERE 코드 문서화가 필요하면, 시스템은 모든 public 함수에 JSDoc 주석을 추가할 수 있다
- WHERE 예시 코드가 필요하면, 시스템은 `@example` 태그를 포함할 수 있다
- WHERE 타입 설명이 필요하면, 시스템은 `@param`, `@returns` 태그를 추가할 수 있다

**근거**: 개발자 경험 향상, 선택 사항

---

### Constraints (제약사항)

#### C-001: 기존 훅 동작 불변
- IF 리팩토링으로 기존 훅의 동작이 변경되면, 시스템은 리팩토링을 중단하고 경고해야 한다
- IF 기존 테스트가 실패하면, 시스템은 커밋을 차단해야 한다
- IF Fast-Track 패턴이 제거되면, 시스템은 성능 저하 경고를 출력해야 한다

**근거**: 후방 호환성 최우선

#### C-002: 테스트 커버리지 85% 유지
- IF 신규 파일(base.ts, constants.ts, utils.ts)의 테스트 커버리지가 85% 미만이면, 시스템은 PR 머지를 차단해야 한다
- IF 기존 훅 파일의 커버리지가 감소하면, 시스템은 경고를 출력해야 한다
- IF 테스트가 없는 함수가 발견되면, 시스템은 린터 오류를 발생시켜야 한다

**근거**: TRUST 5원칙 중 Test First

#### C-003: 타입 안전성 필수
- IF `Record<string, any>`가 새로 추가되면, 시스템은 컴파일을 차단해야 한다
- IF `as any` 타입 단언이 사용되면, 시스템은 린터 오류를 발생시켜야 한다
- IF 암시적 `any` 타입이 발견되면, 시스템은 `tsconfig.json`의 `strict: true` 설정으로 차단해야 한다

**근거**: TRUST 5원칙 중 Unified (타입 안전성)

#### C-004: 파일 크기 제약
- IF 단일 파일이 300 LOC를 초과하면, 시스템은 경고를 출력해야 한다
- IF 단일 함수가 50 LOC를 초과하면, 시스템은 린터 경고를 발생시켜야 한다
- IF 순환 복잡도가 10을 초과하면, 시스템은 리팩토링을 권장해야 한다

**근거**: development-guide.md 코드 제약 준수

---

## Traceability (@TAG)

### TAG 체계
```
@SPEC:HOOKS-REFACTOR-001 (이 문서)
  └─> @TEST:HOOKS-REFACTOR-001 (src/claude/hooks/__tests__/base.test.ts, constants.test.ts, utils.test.ts)
      └─> @CODE:HOOKS-REFACTOR-001 (src/claude/hooks/base.ts, constants.ts, utils.ts)
          └─> @DOC:HOOKS-REFACTOR-001 (sync 후 자동 생성)
```

### 관련 파일
- **SPEC**: `.moai/specs/SPEC-HOOKS-REFACTOR-001/spec.md` (이 문서)
- **TEST**:
  - `src/claude/hooks/__tests__/base.test.ts` (신규)
  - `src/claude/hooks/__tests__/constants.test.ts` (신규)
  - `src/claude/hooks/__tests__/utils.test.ts` (신규)
  - `src/claude/hooks/__tests__/tag-enforcer.test.ts` (수정)
- **CODE**:
  - `src/claude/hooks/base.ts` (신규)
  - `src/claude/hooks/constants.ts` (신규)
  - `src/claude/hooks/utils.ts` (신규)
  - `src/claude/hooks/policy-block.ts` (수정)
  - `src/claude/hooks/pre-write-guard.ts` (수정)
  - `src/claude/hooks/tag-enforcer.ts` (수정)
  - `src/claude/hooks/session-notice/index.ts` (수정)
- **DOC**: `.moai/reports/sync-report-HOOKS-REFACTOR-001.md` (sync 후 자동 생성)

---

## Implementation Details

### 1. 신규 파일: base.ts

**목적**: CLI Entry Point 중복 제거

**주요 기능**:
- `runHook(HookClass)`: 훅 인스턴스 생성 및 실행
- `parseClaudeInput()` import 및 재사용
- `outputResult()` import 및 재사용

**타입 시그니처**:
```typescript
export async function runHook(
  HookClass: new () => MoAIHook
): Promise<void>
```

**사용 예시**:
```typescript
// policy-block.ts
if (require.main === module) {
  runHook(PolicyBlock).catch(/* ... */);
}
```

---

### 2. 신규 파일: constants.ts

**목적**: 하드코딩 상수 중앙 관리

**주요 상수**:

#### SUPPORTED_LANGUAGES
```typescript
export const SUPPORTED_LANGUAGES = {
  typescript: ['.ts', '.tsx'],
  javascript: ['.js', '.jsx', '.mjs', '.cjs'],
  python: ['.py', '.pyi'],
  java: ['.java'],
  go: ['.go'],
  rust: ['.rs'],
  cpp: ['.cpp', '.hpp', '.cc', '.h', '.cxx', '.hxx'],
  ruby: ['.rb', '.rake', '.gemspec'],
  php: ['.php'],
  csharp: ['.cs'],
  dart: ['.dart'],
  swift: ['.swift'],
  kotlin: ['.kt', '.kts'],
  elixir: ['.ex', '.exs'],
  markdown: ['.md', '.mdx'],
} as const;
```

#### READ_ONLY_TOOLS
```typescript
export const READ_ONLY_TOOLS = [
  'Read', 'Glob', 'Grep', 'WebFetch', 'WebSearch', 'TodoWrite', 'BashOutput',
  'mcp__context7__resolve-library-id', 'mcp__context7__get-library-docs',
  'mcp__ide__getDiagnostics', 'mcp__ide__executeCode',
] as const;
```

#### TIMEOUTS
```typescript
export const TIMEOUTS = {
  TAG_BLOCK_SEARCH_LIMIT: 30,
  GIT_COMMAND_TIMEOUT: 2000,
  NPM_REGISTRY_TIMEOUT: 2000,
  POLICY_BLOCK_SLOW_THRESHOLD: 100,
} as const;
```

#### DANGEROUS_COMMANDS
```typescript
export const DANGEROUS_COMMANDS = [
  'rm -rf /',
  'rm -rf --no-preserve-root',
  'sudo rm',
  'dd if=/dev/zero',
  ':(){:|:&};:',
  'mkfs.',
] as const;
```

#### EXCLUDED_PATHS
```typescript
export const EXCLUDED_PATHS = [
  'node_modules',
  '.git',
  'dist',
  'build',
  'test',
  'spec',
  '__test__',
  '__tests__',
] as const;
```

---

### 3. 신규 파일: utils.ts

**목적**: 공통 유틸리티 함수 제공

**주요 함수**:

#### extractFilePath
```typescript
export function extractFilePath(
  toolInput: Record<string, any>
): string | null {
  return toolInput.file_path
    || toolInput.filePath
    || toolInput.path
    || toolInput.notebook_path
    || null;
}
```

#### extractCommand
```typescript
export function extractCommand(
  toolInput: Record<string, any>
): string | null {
  const raw = toolInput.command || toolInput.cmd;

  if (Array.isArray(raw)) {
    return raw.map(String).join(' ');
  }

  if (typeof raw === 'string') {
    return raw.trim();
  }

  return null;
}
```

#### getAllFileExtensions
```typescript
export function getAllFileExtensions(): string[] {
  return Object.values(SUPPORTED_LANGUAGES).flat();
}
```

---

### 4. 기존 파일 수정: policy-block.ts

**변경 사항**:
1. `DANGEROUS_COMMANDS` → `import { DANGEROUS_COMMANDS } from './constants'`
2. `READ_ONLY_TOOLS` → `import { READ_ONLY_TOOLS } from './constants'`
3. `extractCommand()` → `import { extractCommand } from './utils'`
4. `main()` 함수 제거 → `import { runHook } from './base'`
5. `require.main === module` → `runHook(PolicyBlock)`

**예상 LOC 변경**: -30 LOC (중복 제거)

---

### 5. 기존 파일 수정: pre-write-guard.ts

**변경 사항**:
1. `SENSITIVE_KEYWORDS` → `import { SENSITIVE_KEYWORDS } from './constants'` (신규 상수)
2. `PROTECTED_PATHS` → `import { PROTECTED_PATHS } from './constants'`
3. `extractFilePath()` → `import { extractFilePath } from './utils'`
4. `main()` 함수 제거 → `import { runHook } from './base'`
5. `require.main === module` → `runHook(PreWriteGuard)`

**예상 LOC 변경**: -25 LOC

---

### 6. 기존 파일 수정: tag-enforcer.ts

**변경 사항**:
1. `enforceExtensions` 배열 → `import { getAllFileExtensions } from './utils'`
2. `shouldEnforceTags()` 로직 → `getAllFileExtensions().includes(ext)` 사용
3. `EXCLUDED_PATHS` → `import { EXCLUDED_PATHS } from './constants'`
4. `extractFilePath()` → `import { extractFilePath } from './utils'`
5. `main()` 함수 제거 → `import { runHook } from './base'`
6. `require.main === module` → `runHook(CodeFirstTAGEnforcer)`

**예상 LOC 변경**: -35 LOC

---

### 7. 기존 파일 수정: session-notice/index.ts

**변경 사항**:
1. `main()` 함수 제거 → `import { runHook } from '../base'`
2. `require.main === module` → `runHook(SessionNotifier)`
3. `TIMEOUTS` → `import { TIMEOUTS } from '../constants'` (utils.ts의 타임아웃 사용)

**예상 LOC 변경**: -15 LOC

---

## 리팩토링 체크리스트

### Phase 1: SPEC 작성 (현재)
- [x] SPEC ID 중복 확인 완료
- [x] spec.md 작성 완료 (EARS 5가지 구문 포함)
- [ ] plan.md 작성 (TDD 계획)
- [ ] acceptance.md 작성 (Given-When-Then)

### Phase 2: TDD 구현 (/alfred:2-build)
- [ ] RED: 실패 테스트 16개 작성
- [ ] GREEN: 최소 구현으로 테스트 통과
- [ ] REFACTOR: 코드 품질 개선, Biome 린트 통과

### Phase 3: 문서 동기화 (/alfred:3-sync)
- [ ] TAG 체인 무결성 검증
- [ ] Living Document 자동 생성
- [ ] PR Ready 전환

---

## 성공 기준

### 기능 검증
- [ ] 기존 76개 테스트 100% 통과
- [ ] 신규 16개 테스트 100% 통과
- [ ] 모든 언어 확장자(11개 언어) 인식 성공

### 코드 품질
- [ ] Biome 린터 0 경고
- [ ] TypeScript strict mode 컴파일 성공
- [ ] 테스트 커버리지 ≥85%

### 리팩토링 효과
- [ ] 코드 중복 100 LOC → 0 LOC
- [ ] 언어 지원 7개 → 11개 (57% 증가)
- [ ] 하드코딩 상수 3개 영역 → 1개 파일

### 성능
- [ ] Fast-Track 패턴 유지 (Read-only tools < 1ms)
- [ ] policy-block 실행 시간 < 100ms
- [ ] 빌드 시간 증가 < 5%

---

## 참고 문서

- **개발 가이드**: `.moai/memory/development-guide.md`
- **SPEC 메타데이터 표준**: `.moai/memory/spec-metadata.md`
- **Product 정의**: `.moai/project/product.md`
- **기술 스택**: `.moai/project/tech.md`
- **기존 리팩토링 SPEC**: `.moai/specs/SPEC-REFACTOR-001/spec.md`
