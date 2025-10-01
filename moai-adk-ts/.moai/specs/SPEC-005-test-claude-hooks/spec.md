# SPEC-005: claude/hooks 모듈 테스트 작성

## @DOC:SPEC:TEST-HOOKS-001
## CHAIN: REQ:SECURITY-001 -> DESIGN:TEST-001 -> TASK:TEST-HOOKS-001 -> TEST:TEST-HOOKS-001
## DEPENDS: HOOK-001, HOOK-002, HOOK-003, HOOK-004
## STATUS: active
## CREATED: 2025-10-01
## @IMMUTABLE

---

## Environment (환경 및 가정사항)

### 현재 상황
- **모듈 위치**: `src/claude/hooks/`
- **파일 개수**: 4개 (policy-block.ts, pre-write-guard.ts, session-notice.ts, tag-enforcer.ts)
- **현재 테스트**: 0개 (hooks 디렉토리 내)
- **기존 테스트**: 3개 (security 디렉토리 내 - 레거시)
- **문제점**:
  - 테스트 커버리지 0%
  - 보안 중요 모듈임에도 불구하고 테스트 부재
  - 기존 테스트는 구 위치(security/)에 있어 실제 hooks와 분리됨

### 기술 스택
- **테스트 프레임워크**: Vitest 3.2.4
- **Mocking**: @vitest/spy 3.2.4
- **런타임**: Node.js ≥18.0.0, Bun ≥1.2.0
- **타입 안전성**: TypeScript strict mode

### 제약사항
- 테스트 파일은 소스 파일과 같은 위치에 배치 (*.test.ts)
- 각 테스트는 독립적으로 실행 가능해야 함
- 외부 의존성 최소화 (mocking 활용)
- 파일당 테스트 실행 시간 ≤ 5초

---

## Assumptions (전제 조건)

1. **타입 시스템 정의**
   - `HookInput`, `HookResult`, `MoAIHook` 인터페이스가 정의되어야 함
   - 현재는 각 파일에 분산되어 있으므로 통합 필요
   - 타입 파일: `src/claude/hooks/types.ts` (신규 생성 필요)

2. **테스트 환경**
   - Vitest 설정이 올바르게 구성되어 있음
   - TypeScript 경로 해석이 작동함 (vite-tsconfig-paths)
   - 모든 필수 devDependencies가 설치되어 있음

3. **Hook 동작 원리**
   - 각 Hook은 MoAIHook 인터페이스를 구현
   - execute() 메서드는 HookInput을 받아 HookResult를 반환
   - 차단 조건 충족 시 `{ success: false, blocked: true }` 반환
   - 정상 통과 시 `{ success: true }` 반환

---

## Requirements (기능 요구사항 - EARS 방식)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 각 Hook 파일에 대응하는 테스트 파일을 제공해야 한다
- 시스템은 최소 85% 테스트 커버리지를 달성해야 한다
- 시스템은 Vitest 프레임워크를 사용해야 한다
- 시스템은 모든 공개 메서드에 대한 테스트를 포함해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN Hook이 차단 조건을 만족하면, 시스템은 올바른 에러 메시지를 반환하는지 테스트해야 한다
- WHEN Hook이 허용 조건을 만족하면, 시스템은 정상 통과하는지 테스트해야 한다
- WHEN 잘못된 입력이 제공되면, 시스템은 graceful하게 처리하는지 테스트해야 한다

### State-driven Requirements (상태 기반)
- WHILE @IMMUTABLE TAG가 존재할 때, tag-enforcer가 수정을 차단하는지 테스트해야 한다
- WHILE 보안 패턴이 감지될 때, policy-block이 차단하는지 테스트해야 한다
- WHILE MoAI 프로젝트가 초기화되지 않았을 때, session-notice가 초기화 안내를 표시하는지 테스트해야 한다

### Optional Features (선택적 기능)
- WHERE 테스트 UI가 필요하면, `npm run test:ui`로 실행할 수 있다
- WHERE 커버리지 리포트가 필요하면, `npm run test:coverage`로 생성할 수 있다

### Constraints (제약사항)
- IF 타입이 정의되지 않았으면, 시스템은 컴파일을 거부해야 한다
- 각 테스트 파일은 해당 소스 파일 옆에 배치해야 한다 (.test.ts)
- 테스트는 격리되어 독립 실행 가능해야 한다
- 테스트 실행 시간은 파일당 5초 이하여야 한다
- 외부 파일 시스템 접근은 최소화하고 mocking을 활용해야 한다

---

## Specifications (상세 명세)

### 1. 타입 시스템 정의 (@CODE:HOOK-TYPES-001:API)

**파일**: `src/claude/hooks/types.ts`

```typescript
/**
 * @CODE:HOOK-TYPES-001 | Chain: @SPEC:SECURITY-001 -> @SPEC:TEST-001 -> @CODE:TEST-HOOKS-001 -> @TEST:TEST-HOOKS-001
 * Related: @CODE:HOOK-001:API, @CODE:HOOK-002:API, @CODE:HOOK-003:API, @CODE:HOOK-004:API
 *
 * Hook System Type Definitions
 */

export interface HookInput {
  tool_name?: string;
  tool_input?: Record<string, any>;
  [key: string]: unknown;
}

export interface HookResult {
  success: boolean;
  blocked?: boolean;
  message?: string;
  exitCode?: number;
  data?: unknown;
}

export interface MoAIHook {
  name: string;
  execute(input: HookInput): Promise<HookResult>;
}
```

### 2. policy-block.test.ts (@TEST:POLICY-001)

**파일**: `src/claude/hooks/policy-block.test.ts`

**테스트 시나리오**:
1. ✅ 비-Bash 도구 허용 테스트
2. ✅ 안전한 bash 명령어 허용 테스트
3. 🚫 위험한 rm -rf 명령 차단 테스트
4. 🚫 dd 명령 차단 테스트
5. 🚫 Fork bomb 차단 테스트
6. 🚫 mkfs 명령 차단 테스트
7. ✅ 배열 형식 명령어 처리 테스트
8. ✅ cmd 파라미터명 처리 테스트
9. ✅ 명령어 누락 시 graceful 처리 테스트
10. ⚠️ 미등록 명령어 경고 표시 테스트
11. ✅ 화이트리스트 명령어 프리픽스 허용 테스트
12. 🚫 대소문자 무시 위험 명령 차단 테스트

**커버리지 목표**: 90% 이상

### 3. pre-write-guard.test.ts (@TEST:PREWRITE-001)

**파일**: `src/claude/hooks/pre-write-guard.test.ts`

**테스트 시나리오**:
1. ✅ 비-쓰기 작업 허용 테스트
2. ✅ 안전한 파일 쓰기 허용 테스트
3. 🚫 .env 파일 쓰기 차단 테스트
4. 🚫 /secrets 디렉토리 쓰기 차단 테스트
5. 🚫 .git 디렉토리 쓰기 차단 테스트
6. 🚫 .ssh 디렉토리 쓰기 차단 테스트
7. 🚫 .moai/memory/ 디렉토리 쓰기 차단 테스트
8. ✅ 다양한 파일 경로 파라미터명 처리 테스트
9. ✅ 파일 경로 누락 시 graceful 처리 테스트
10. ✅ 빈 tool_input 처리 테스트
11. 🚫 대소문자 무시 경로 차단 테스트
12. ✅ .moai/project/ 디렉토리 쓰기 허용 테스트
13. ✅ Write/Edit/MultiEdit 작업 지원 테스트

**커버리지 목표**: 85% 이상

### 4. session-notice.test.ts (@TEST:SESSION-001)

**파일**: `src/claude/hooks/session-notice.test.ts`

**테스트 시나리오**:
1. ✅ MoAI 프로젝트 감지 테스트
2. ✅ 비-MoAI 프로젝트 초기화 안내 테스트
3. ✅ 프로젝트 상태 정보 수집 테스트
4. ✅ constitution 상태 검사 테스트
5. ✅ MoAI 버전 감지 테스트 (package.json / config.json)
6. ✅ 파이프라인 단계 감지 테스트
7. ✅ SPEC 진행률 계산 테스트
8. ✅ Git 정보 수집 테스트 (브랜치, 커밋, 변경사항)
9. ✅ 버전 업데이트 확인 테스트 (npm registry)
10. ✅ 세션 출력 메시지 생성 테스트
11. ✅ 에러 발생 시 silent failure 테스트
12. ✅ Git 명령 타임아웃 처리 테스트

**특수 고려사항**:
- 파일 시스템 접근 mocking (fs-extra 활용)
- Git 명령 mocking (child_process.spawn)
- HTTP 요청 mocking (fetch API)
- 타임아웃 시나리오 테스트

**커버리지 목표**: 80% 이상 (외부 의존성 많음)

### 5. tag-enforcer.test.ts (@TEST:ENFORCER-001)

**파일**: `src/claude/hooks/tag-enforcer.test.ts`

**테스트 시나리오**:
1. ✅ 비-쓰기 작업 허용 테스트
2. ✅ TAG 미적용 파일 타입 허용 테스트 (테스트 파일, node_modules)
3. ✅ TAG 블록 추출 테스트
4. 🚫 @IMMUTABLE TAG 블록 수정 차단 테스트
5. 🚫 @IMMUTABLE TAG 블록 삭제 차단 테스트
6. ✅ 새 파일에 TAG 블록 추가 허용 테스트
7. ✅ TAG 블록 유효성 검증 테스트 (카테고리, 도메인 ID)
8. ✅ 체인 검증 테스트 (CHAIN 라인)
9. ✅ 의존성 검증 테스트 (DEPENDS 라인)
10. ✅ 상태 검증 테스트 (STATUS)
11. ✅ 생성 날짜 검증 테스트 (CREATED)
12. ⚠️ @IMMUTABLE 마커 권장 경고 테스트
13. ✅ TAG 제안 생성 테스트
14. ✅ 파일 타입별 처리 테스트 (.ts, .py, .md 등)
15. ✅ Write/Edit/MultiEdit/NotebookEdit 지원 테스트
16. ✅ 에러 발생 시 non-blocking 처리 테스트

**특수 고려사항**:
- 파일 시스템 mocking (fs/promises)
- 정규식 패턴 테스트
- TAG 블록 정규화 로직 테스트
- 복잡한 TAG 체인 시나리오

**커버리지 목표**: 90% 이상

---

## Test Structure (테스트 구조)

### 디렉토리 구조
```
src/claude/hooks/
├── types.ts (신규)
├── policy-block.ts
├── policy-block.test.ts (신규)
├── pre-write-guard.ts
├── pre-write-guard.test.ts (신규)
├── session-notice.ts
├── session-notice.test.ts (신규)
├── tag-enforcer.ts
└── tag-enforcer.test.ts (신규)
```

### 테스트 파일 템플릿
```typescript
// @TEST:<HOOK-ID>-001 | Chain: @SPEC:SECURITY-001 -> @SPEC:TEST-001 -> @CODE:TEST-HOOKS-001 -> @TEST:<HOOK-ID>-001
// Related: @CODE:<HOOK-ID>

/**
 * @file <hook-name>.test.ts
 * @description Tests for <HookClass> hook
 */

import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { <HookClass> } from './<hook-file>';
import type { HookInput } from './types';

describe('<HookClass>', () => {
  let hook: <HookClass>;

  beforeEach(() => {
    hook = new <HookClass>();
    // Setup mocks
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('execute', () => {
    test('should handle expected scenario', async () => {
      const input: HookInput = {
        tool_name: 'ToolName',
        tool_input: { /* ... */ }
      };

      const result = await hook.execute(input);

      expect(result.success).toBe(true);
    });

    // More tests...
  });
});
```

---

## Quality Gates (품질 게이트)

### 1. 테스트 커버리지
- **전체**: ≥ 85%
- **policy-block**: ≥ 90%
- **pre-write-guard**: ≥ 85%
- **session-notice**: ≥ 80%
- **tag-enforcer**: ≥ 90%

### 2. 테스트 실행 시간
- **파일당**: ≤ 5초
- **전체 suite**: ≤ 20초

### 3. 타입 안전성
- TypeScript strict mode 준수
- 모든 타입 명시적 정의
- any 타입 사용 금지 (Record<string, any> 예외)

### 4. 테스트 격리
- 각 테스트는 독립 실행 가능
- 외부 의존성 mocking
- beforeEach/afterEach로 상태 초기화

---

## Traceability (추적성)

### TAG Chain Mapping
```
@SPEC:SECURITY-001 (보안 요구사항)
  └─> @SPEC:TEST-001 (테스트 설계)
      └─> @CODE:TEST-HOOKS-001 (테스트 작업)
          └─> @TEST:POLICY-001 (PolicyBlock 테스트)
          └─> @TEST:PREWRITE-001 (PreWriteGuard 테스트)
          └─> @TEST:SESSION-001 (SessionNotifier 테스트)
          └─> @TEST:ENFORCER-001 (TagEnforcer 테스트)
```

### Related TAGs
- @CODE:HOOK-001 (PolicyBlock)
- @CODE:HOOK-002 (PreWriteGuard)
- @CODE:HOOK-003 (SessionNotifier)
- @CODE:HOOK-004 (CodeFirstTAGEnforcer)
- @CODE:HOOK-TYPES-001:API (타입 시스템)

---

## Success Criteria (성공 기준)

### Definition of Done
- [ ] types.ts 파일 생성 완료
- [ ] 4개 테스트 파일 생성 완료
- [ ] 모든 테스트 통과 (npm run test)
- [ ] 테스트 커버리지 85% 이상 달성
- [ ] TypeScript 컴파일 에러 없음
- [ ] Biome 린트 검사 통과
- [ ] 각 Hook의 핵심 시나리오 커버
- [ ] @TAG 체인 유지 및 검증

### Acceptance Tests
1. `npm run test` 실행 시 모든 테스트 통과
2. `npm run test:coverage` 실행 시 85% 이상 커버리지 확인
3. `npm run type-check` 실행 시 타입 에러 없음
4. `npm run check:biome` 실행 시 린트 에러 없음

---

## Implementation Notes (구현 참고사항)

### Mocking 가이드

#### 파일 시스템 Mocking
```typescript
import { vi } from 'vitest';
import * as fs from 'fs/promises';

vi.mock('fs/promises', () => ({
  readFile: vi.fn(),
  writeFile: vi.fn(),
  access: vi.fn(),
}));
```

#### Child Process Mocking
```typescript
import { vi } from 'vitest';
import { spawn } from 'node:child_process';

vi.mock('node:child_process', () => ({
  spawn: vi.fn(),
}));
```

#### Fetch API Mocking
```typescript
global.fetch = vi.fn(() =>
  Promise.resolve({
    ok: true,
    json: () => Promise.resolve({ version: '1.0.0' }),
  })
) as any;
```

### 테스트 우선순위
1. **High Priority**: policy-block, tag-enforcer (보안 중요)
2. **Medium Priority**: pre-write-guard (데이터 보호)
3. **Low Priority**: session-notice (UX 개선)

### 기존 테스트 마이그레이션
- `src/__tests__/claude/hooks/security/` 내 기존 테스트 3개 참고
- 새 위치(`src/claude/hooks/`)로 로직 통합
- 중복 제거 및 개선된 테스트 케이스 작성

---

## Risk Mitigation (위험 완화)

### 식별된 위험
1. **타입 불일치**: 각 Hook마다 타입이 다를 수 있음
   - **완화**: types.ts 통합 정의 및 모든 Hook에서 import

2. **외부 의존성**: 파일 시스템, Git, HTTP 요청
   - **완화**: Mocking 철저히 활용, 실제 환경 테스트 최소화

3. **테스트 격리**: 테스트 간 상태 공유
   - **완화**: beforeEach/afterEach 철저히 구현

4. **커버리지 부족**: 복잡한 로직 미커버
   - **완화**: 엣지 케이스 포함, 에러 시나리오 테스트

---

## References (참고 자료)

- Vitest 공식 문서: https://vitest.dev/
- 기존 테스트: `src/__tests__/claude/hooks/security/policy-block.test.ts`
- TRUST 원칙: `.moai/memory/development-guide.md`
- TypeScript 핸드북: https://www.typescriptlang.org/docs/

---

**SPEC 버전**: 1.0.0
**최종 수정일**: 2025-10-01
**승인 필요**: No (Draft)
