# SPEC-005: claude/hooks 모듈 테스트 작성 - 구현 계획

## @DOC:PLAN:TEST-HOOKS-001
## CHAIN: REQ:SECURITY-001 -> DESIGN:TEST-001 -> TASK:TEST-HOOKS-001 -> TEST:TEST-HOOKS-001
## STATUS: active
## CREATED: 2025-10-01

---

## 우선순위별 마일스톤

### 1차 목표: 타입 시스템 및 기반 구축
- [ ] types.ts 파일 생성
  - HookInput, HookResult, MoAIHook 인터페이스 정의
  - 기존 Hook 파일에서 타입 추출 및 통합
  - 모든 Hook 파일에서 import 구문 추가
- [ ] 테스트 환경 검증
  - Vitest 설정 확인
  - Mock 라이브러리 준비 확인
  - 경로 해석 작동 확인

### 2차 목표: 보안 중요 Hook 테스트 (High Priority)
- [ ] policy-block.test.ts 작성
  - 12개 테스트 시나리오 구현
  - 커버리지 90% 이상 달성
  - 위험 명령 차단 검증
- [ ] tag-enforcer.test.ts 작성
  - 16개 테스트 시나리오 구현
  - @IMMUTABLE 불변성 검증
  - TAG 체인 유효성 검증

### 3차 목표: 데이터 보호 Hook 테스트 (Medium Priority)
- [ ] pre-write-guard.test.ts 작성
  - 13개 테스트 시나리오 구현
  - 민감 파일 보호 검증
  - 다양한 파라미터 처리 테스트

### 최종 목표: UX Hook 테스트 및 검증
- [ ] session-notice.test.ts 작성
  - 12개 테스트 시나리오 구현
  - 외부 의존성 mocking
  - Git/HTTP 요청 처리 테스트
- [ ] 전체 품질 게이트 검증
  - 전체 커버리지 85% 이상 확인
  - 모든 테스트 통과 확인
  - 타입 체크 및 린트 통과

---

## 기술적 접근 방법

### 1. TDD 사이클 적용

각 테스트 파일에 대해 Red-Green-Refactor 사이클 적용:

```
RED (실패하는 테스트 작성)
  ↓
GREEN (최소한의 구현으로 통과)
  ↓
REFACTOR (코드 품질 개선)
  ↓
REPEAT (다음 시나리오)
```

### 2. 테스트 작성 순서

#### 2.1 types.ts 먼저 생성
```typescript
// 1. 기존 Hook 파일에서 타입 추출
// 2. 공통 인터페이스 정의
// 3. 모든 Hook에서 import
```

#### 2.2 각 Hook별 테스트 작성 순서
1. **기본 시나리오** (Happy Path)
   - 정상 입력 → 정상 출력
   - 도구 무시 → 성공 반환

2. **차단 시나리오** (Blocking Cases)
   - 위험 패턴 감지 → 차단
   - 보호 대상 수정 → 차단

3. **엣지 케이스** (Edge Cases)
   - 빈 입력 → graceful 처리
   - 잘못된 형식 → 안전 처리
   - 타임아웃 → fallback

4. **에러 시나리오** (Error Handling)
   - 예외 발생 → non-blocking
   - 외부 실패 → silent failure

### 3. Mocking 전략

#### 3.1 파일 시스템 Mocking
```typescript
// fs/promises 모듈 전체 모킹
vi.mock('fs/promises', () => ({
  readFile: vi.fn((path) => {
    if (path.includes('existing-file')) {
      return Promise.resolve('file content');
    }
    return Promise.reject(new Error('File not found'));
  }),
  access: vi.fn(),
}));
```

#### 3.2 Child Process Mocking
```typescript
// Git 명령 모킹
vi.mock('node:child_process', () => ({
  spawn: vi.fn(() => ({
    stdout: {
      on: vi.fn((event, callback) => {
        if (event === 'data') {
          callback(Buffer.from('main\n'));
        }
      }),
    },
    on: vi.fn((event, callback) => {
      if (event === 'close') {
        callback(0);
      }
    }),
  })),
}));
```

#### 3.3 Fetch API Mocking
```typescript
// npm registry 요청 모킹
global.fetch = vi.fn((url) => {
  if (url.includes('moai-adk/latest')) {
    return Promise.resolve({
      ok: true,
      json: () => Promise.resolve({ version: '1.0.0' }),
    });
  }
  return Promise.reject(new Error('Network error'));
}) as any;
```

### 4. 테스트 격리 패턴

#### 4.1 beforeEach/afterEach 활용
```typescript
describe('HookName', () => {
  let hook: HookClass;
  let mockFs: any;

  beforeEach(() => {
    // 1. 인스턴스 생성
    hook = new HookClass();

    // 2. Mock 초기화
    vi.clearAllMocks();

    // 3. 환경 설정
    process.env.TEST_MODE = 'true';
  });

  afterEach(() => {
    // 1. Mock 복원
    vi.restoreAllMocks();

    // 2. 환경 정리
    delete process.env.TEST_MODE;
  });
});
```

#### 4.2 테스트 스위트 분리
```typescript
describe('PolicyBlock', () => {
  describe('execute', () => {
    describe('when Bash tool is used', () => {
      // Bash 관련 테스트
    });

    describe('when non-Bash tool is used', () => {
      // 기타 도구 테스트
    });
  });

  describe('extractCommand', () => {
    // private 메서드 간접 테스트
  });
});
```

---

## 아키텍처 설계 방향

### 1. 타입 시스템 통합

#### Before (분산된 타입)
```
policy-block.ts      → 내부 타입 정의
pre-write-guard.ts   → 내부 타입 정의
session-notice.ts    → 내부 타입 정의
tag-enforcer.ts      → 내부 타입 정의
```

#### After (통합된 타입)
```
types.ts             → 공통 타입 정의
  ↓ import
policy-block.ts
pre-write-guard.ts
session-notice.ts
tag-enforcer.ts
```

### 2. 테스트 파일 배치

#### 선택한 구조: Colocation
```
src/claude/hooks/
├── types.ts
├── policy-block.ts
├── policy-block.test.ts  ← 소스 옆에 배치
├── pre-write-guard.ts
├── pre-write-guard.test.ts
├── session-notice.ts
├── session-notice.test.ts
├── tag-enforcer.ts
└── tag-enforcer.test.ts
```

**장점**:
- 소스와 테스트 간 이동 용이
- 관련 코드 한 곳에 모임
- 리팩토링 시 일관성 유지

**단점 (해결책)**:
- 빌드 산출물에 테스트 포함 → .npmignore로 제외
- 디렉토리 복잡 → glob 패턴으로 필터링

### 3. 의존성 관리

#### 외부 의존성 최소화
```typescript
// ❌ 실제 파일 시스템 접근
const content = await fs.readFile(path, 'utf-8');

// ✅ Mocking으로 대체
const mockReadFile = vi.fn(() => Promise.resolve('content'));
```

#### 테스트 유틸리티 분리
```typescript
// tests/utils/mock-helpers.ts (필요시)
export const createMockHookInput = (overrides = {}): HookInput => ({
  tool_name: 'Bash',
  tool_input: {},
  ...overrides,
});
```

---

## 리스크 및 대응 방안

### 1. 타입 불일치 리스크
**문제**: 각 Hook마다 타입 정의가 다를 수 있음

**대응**:
- Step 1: 모든 Hook 파일 검토 및 타입 추출
- Step 2: 공통 인터페이스 정의 (types.ts)
- Step 3: 기존 Hook 파일에서 타입 import
- Step 4: TypeScript 컴파일 에러 해결

### 2. 외부 의존성 리스크
**문제**: 파일 시스템, Git, HTTP 요청에 의존

**대응**:
- 모든 외부 호출 mocking
- 타임아웃 시나리오 테스트
- Fallback 동작 검증
- 네트워크 격리 환경 테스트

### 3. 테스트 격리 실패 리스크
**문제**: 테스트 간 상태 공유로 간섭

**대응**:
- beforeEach에서 완전 초기화
- afterEach에서 mock 복원
- 전역 변수 사용 금지
- 독립 실행 검증 (--run)

### 4. 커버리지 부족 리스크
**문제**: 복잡한 로직 미커버

**대응**:
- 엣지 케이스 명시적 테스트
- 에러 경로 전부 커버
- private 메서드 간접 테스트
- 커버리지 리포트 분석

### 5. 성능 저하 리스크
**문제**: 테스트 실행 시간 초과

**대응**:
- 병렬 실행 활용 (Vitest 기본)
- 무거운 mock 최소화
- 타임아웃 짧게 설정
- 불필요한 setup 제거

---

## 검증 체크리스트

### 코드 품질
- [ ] TypeScript strict mode 준수
- [ ] Biome 린트 규칙 통과
- [ ] 모든 타입 명시적 정의
- [ ] any 타입 최소화

### 테스트 품질
- [ ] 각 테스트 독립 실행 가능
- [ ] 명확한 테스트 이름 (should...)
- [ ] Arrange-Act-Assert 패턴
- [ ] 한 테스트당 한 가지만 검증

### 커버리지
- [ ] 전체 85% 이상
- [ ] 핵심 로직 100% 커버
- [ ] 엣지 케이스 포함
- [ ] 에러 경로 포함

### 문서화
- [ ] 각 테스트 파일에 @TAG 주석
- [ ] 복잡한 테스트에 설명 주석
- [ ] README 업데이트 (필요시)

---

## 예상 작업량

### 타입 시스템 (1차 목표)
- types.ts 생성: 1 unit
- Hook 파일 수정: 4 units (각 파일 0.25 unit)
- **소계**: 2 units

### 테스트 파일 작성 (2-4차 목표)
- policy-block.test.ts: 3 units (12 scenarios)
- tag-enforcer.test.ts: 4 units (16 scenarios, 복잡)
- pre-write-guard.test.ts: 3 units (13 scenarios)
- session-notice.test.ts: 4 units (12 scenarios, mocking 많음)
- **소계**: 14 units

### 검증 및 수정
- 커버리지 분석: 1 unit
- 누락 시나리오 추가: 2 units
- 타입/린트 오류 수정: 1 unit
- **소계**: 4 units

### 총 예상 작업량: 20 units

---

## 도구 및 명령어

### 개발 중
```bash
# 테스트 watch mode
npm run test:watch

# 특정 파일만 테스트
npm run test -- policy-block.test.ts

# UI 모드로 테스트
npm run test:ui
```

### 검증
```bash
# 전체 테스트 실행
npm run test

# 커버리지 확인
npm run test:coverage

# 타입 체크
npm run type-check

# 린트 검사
npm run check:biome
```

### CI 파이프라인
```bash
# CI 통합 검증
npm run test:ci
```

---

## 다음 단계 연계

### SPEC-005 완료 후
1. `/alfred:2-build SPEC-005` 실행
   - TDD로 테스트 구현
   - Red-Green-Refactor 사이클

2. `/alfred:3-sync` 실행
   - 문서 동기화
   - TAG 체인 검증
   - PR 상태 전환

3. 기존 테스트 정리
   - `src/__tests__/claude/hooks/security/` 마이그레이션
   - 중복 제거
   - 통합 검증

---

**계획 버전**: 1.0.0
**최종 수정일**: 2025-10-01
