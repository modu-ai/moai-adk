# SPEC-006 구현 계획

## @TAG BLOCK
```
# @CODE:TEST-006 | Chain: @SPEC:QUAL-006 -> @SPEC:TEST-006 -> @CODE:TEST-006 -> @TEST:UTIL-006
# Related: @CODE:VALID-002:API, @CODE:INPUT-001, LOG-002
```

## 개요

본 문서는 SPEC-006 "Utils 모듈 테스트 작성"의 구현 계획을 정의합니다.
TDD 방식으로 3개의 핵심 유틸리티 파일에 대한 테스트를 작성하여 코드 품질과 신뢰성을 확보합니다.

## 우선순위별 마일스톤

### 1차 목표: errors.test.ts 작성
**우선순위**: Highest
**의존성**: 없음 (가장 단순한 구조)

**근거**:
- 가장 단순한 구조 (151 LOC)
- 다른 테스트의 기반이 됨 (에러 객체 사용)
- 100% 커버리지 달성 가능
- TDD 워밍업에 적합

**작업 항목**:
1. 5개 커스텀 에러 클래스 테스트
2. 5개 타입 가드 함수 테스트
3. 2개 헬퍼 함수 테스트
4. 에러 체인 및 상속 검증

**예상 테스트 수**: 15-20개
**커버리지 목표**: 100%

### 2차 목표: input-validator.test.ts 작성
**우선순위**: High
**의존성**: errors.test.ts 완료 (에러 처리 검증 필요)

**근거**:
- 보안에 중요 (@CODE:INPUT-VALIDATION-001)
- 가장 복잡한 로직 (458 LOC)
- 다양한 엣지 케이스 존재
- 실제 파일 시스템 테스트 필요

**작업 항목**:
1. validateProjectName() - 8개 시나리오
2. validatePath() - 10개 시나리오 (파일 시스템 mock)
3. validateTemplateType() - 5개 시나리오
4. validateBranchName() - 8개 시나리오
5. validateCommandOptions() - 6개 시나리오
6. 헬퍼 함수 3개 테스트

**예상 테스트 수**: 40-50개
**커버리지 목표**: 90% 이상

**기술적 도전**:
- 비동기 validatePath() 테스트
- 파일 시스템 mock 또는 임시 디렉토리 사용
- 다양한 위험 패턴 검증

### 3차 목표: winston-logger.test.ts 작성
**우선순위**: Medium
**의존성**: input-validator.test.ts 완료

**근거**:
- Winston 의존성 mock 필요
- 파일 시스템 cleanup 복잡도
- 민감 정보 마스킹 검증 중요
- 349 LOC, 중간 복잡도

**작업 항목**:
1. 로거 초기화 테스트 - 5개 시나리오
2. 로그 레벨 메서드 테스트 - 5개 시나리오
3. 민감 정보 마스킹 테스트 - 8개 시나리오
4. Verbose 모드 테스트 - 4개 시나리오
5. TAG 추적성 테스트 - 2개 시나리오
6. 파일 로깅 테스트 - 5개 시나리오

**예상 테스트 수**: 30-35개
**커버리지 목표**: 85% 이상

**기술적 도전**:
- Winston transport mock
- 파일 시스템 cleanup (beforeEach/afterEach)
- 비동기 로깅 처리
- 정규식 기반 민감 정보 마스킹 검증

### 최종 목표: 통합 검증 및 문서화
**우선순위**: Medium
**의존성**: 모든 테스트 작성 완료

**작업 항목**:
1. 전체 테스트 실행 및 커버리지 확인
2. 테스트 리포트 생성
3. @TAG 체인 검증
4. acceptance.md 완성도 검증
5. README 업데이트 (필요시)

## 기술적 접근 방법

### 1. 테스트 구조 설계

#### 디렉토리 구조
```
src/utils/
├── input-validator.ts
├── input-validator.test.ts     # 신규
├── errors.ts
├── errors.test.ts              # 신규
├── winston-logger.ts
└── winston-logger.test.ts      # 신규
```

**선택 이유**:
- 기존 path-validator.test.ts는 `src/__tests__/utils/`에 위치
- 일관성을 위해 동일한 위치에 배치 가능
- 또는 소스 파일과 동일 디렉토리 배치 (근접성)

**최종 결정**: `src/__tests__/utils/`에 배치 (기존 패턴 준수)

#### 테스트 파일 템플릿
```typescript
/**
 * @file [파일명] test suite
 * @author MoAI Team
 * @tags @TEST:[TAG-ID] @SPEC:QUAL-006
 */

import { describe, expect, test, beforeEach, afterEach, vi } from 'vitest';
import { ... } from '@/utils/[target-file]';

describe('[ModuleName]', () => {
  // Setup/Teardown
  beforeEach(() => {
    // Initialize
  });

  afterEach(() => {
    // Cleanup
  });

  describe('[MethodName]', () => {
    test('should [expected behavior] when [condition]', () => {
      // Given: Arrange

      // When: Act

      // Then: Assert
      expect(result).toBe(expected);
    });
  });
});
```

### 2. Mock 전략

#### Winston Logger Mock
```typescript
// Mock Winston transports
const mockTransport = {
  log: vi.fn(),
  // ... other methods
};

// Inject via LoggerOptions
const logger = new MoaiLogger({
  transports: [mockTransport],
});
```

#### 파일 시스템 Mock
```typescript
// Option 1: 임시 디렉토리 사용 (실제 파일 시스템)
import * as fs from 'node:fs';
import * as os from 'node:os';

let tempDir: string;
beforeEach(() => {
  tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moai-test-'));
});

afterEach(() => {
  fs.rmSync(tempDir, { recursive: true, force: true });
});

// Option 2: vi.mock() 사용
vi.mock('node:fs/promises', () => ({
  stat: vi.fn(),
  readFile: vi.fn(),
}));
```

**권장**:
- input-validator.test.ts: 임시 디렉토리 (실제 파일 시스템 검증)
- winston-logger.test.ts: Mock transport (파일 생성 방지)

### 3. 비동기 테스트 패턴

```typescript
test('should validate path asynchronously', async () => {
  // Given
  const testPath = '/valid/path';

  // When
  const result = await validatePath(testPath);

  // Then
  expect(result.isValid).toBe(true);
});
```

### 4. 에러 케이스 테스트

```typescript
test('should throw ValidationError when invalid input', () => {
  // Given
  const invalidName = '../dangerous';

  // When/Then
  expect(() => {
    validateProjectName(invalidName);
  }).toThrow(ValidationError);

  // Verify error properties
  try {
    validateProjectName(invalidName);
  } catch (error) {
    expect(isValidationError(error)).toBe(true);
    expect(error.pattern).toBeDefined();
  }
});
```

### 5. 커버리지 측정

```bash
# Vitest coverage 실행
npm run test:coverage

# 특정 파일만 테스트
npm test input-validator.test.ts

# Watch 모드
npm test -- --watch
```

**커버리지 설정** (vitest.config.ts):
```typescript
export default defineConfig({
  test: {
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['src/utils/**/*.ts'],
      exclude: ['**/*.test.ts', '**/*.spec.ts'],
      threshold: {
        lines: 85,
        functions: 85,
        branches: 85,
        statements: 85,
      },
    },
  },
});
```

## 아키텍처 설계 방향

### 테스트 격리 원칙
- 각 테스트는 독립적으로 실행 가능
- 테스트 간 상태 공유 금지
- beforeEach/afterEach로 cleanup 보장

### Given-When-Then 패턴
```typescript
test('should mask sensitive data in logs', () => {
  // Given: 민감 정보가 포함된 로그 데이터
  const sensitiveData = {
    username: 'test',
    password: 'secret123',
    apiKey: 'key-abc-123',
  };

  // When: 로거에 기록
  logger.info('User login', sensitiveData);

  // Then: 민감 정보가 마스킹되어 기록됨
  expect(mockTransport.log).toHaveBeenCalledWith(
    expect.objectContaining({
      password: '***REDACTED***',
      apiKey: '***REDACTED***',
    })
  );
});
```

### 테스트 더블 전략
- **Stub**: 고정된 값 반환 (파일 시스템 stat)
- **Mock**: 호출 검증 (Winston transport)
- **Fake**: 실제 구현 대체 (임시 디렉토리)
- **Spy**: 실제 구현 + 호출 추적

## 리스크 및 대응 방안

### 리스크 1: 파일 시스템 테스트 불안정성
**확률**: Medium
**영향**: High

**대응 방안**:
1. 임시 디렉토리 사용으로 격리
2. afterEach에서 강제 cleanup
3. 실패 시 재시도 로직 (flaky test 방지)
4. CI 환경에서 권한 문제 확인

### 리스크 2: Winston Mock 복잡도
**확률**: Medium
**영향**: Medium

**대응 방안**:
1. Transport 의존성 주입 활용
2. 최소한의 mock 인터페이스 구현
3. 실제 Winston 동작과 일치하는지 검증
4. 필요시 integration test 추가

### 리스크 3: 비동기 테스트 타이밍 이슈
**확률**: Low
**영향**: Medium

**대응 방안**:
1. async/await 일관성 있게 사용
2. waitFor() 유틸리티 활용
3. timeout 설정 (필요시)
4. Promise rejection 처리

### 리스크 4: 커버리지 목표 미달
**확률**: Low
**영향**: High

**대응 방안**:
1. 우선순위별 단계적 접근
2. 핵심 로직부터 테스트
3. 에지 케이스는 2차로 추가
4. Waiver 작성 (정당한 이유 시)

## 검증 체크리스트

### 코드 품질
- [ ] 모든 테스트 파일 < 300 LOC
- [ ] 모든 테스트 함수 < 50 LOC
- [ ] ESLint/Biome 검사 통과
- [ ] TypeScript strict mode 컴파일 성공

### 테스트 커버리지
- [ ] errors.ts: 100% 커버리지
- [ ] input-validator.ts: 90% 이상
- [ ] winston-logger.ts: 85% 이상
- [ ] 전체 utils 모듈: 50% 이상

### TDD 원칙
- [ ] Red-Green-Refactor 사이클 준수
- [ ] 테스트 우선 작성
- [ ] 독립적이고 격리된 테스트
- [ ] 결정적 테스트 (매번 동일 결과)

### 문서화
- [ ] @TAG 체인 명시
- [ ] Given-When-Then 주석
- [ ] describe/test 이름 명확
- [ ] acceptance.md 검증

### 실행 성능
- [ ] 전체 테스트 < 10초
- [ ] 개별 테스트 < 100ms
- [ ] 파일 시스템 cleanup 완료

## 완료 조건 (Definition of Done)

1. **3개 테스트 파일 작성 완료**
   - `src/__tests__/utils/errors.test.ts`
   - `src/__tests__/utils/input-validator.test.ts`
   - `src/__tests__/utils/winston-logger.test.ts`

2. **모든 테스트 통과**
   - `npm test` 성공
   - 실패 테스트 0개

3. **커버리지 목표 달성**
   - errors.ts: 100%
   - input-validator.ts: 90%+
   - winston-logger.ts: 85%+
   - 전체 utils: 50%+

4. **품질 게이트 통과**
   - ESLint/Biome 검사 통과
   - TypeScript 컴파일 성공
   - 빌드 성공

5. **문서화 완료**
   - @TAG 체인 검증
   - acceptance.md 업데이트
   - PR 설명 작성 (필요시)

6. **코드 리뷰 준비**
   - 변경 사항 요약
   - 테스트 전략 설명
   - 커버리지 리포트 첨부

---

**작성일**: 2025-10-01
**버전**: 1.0.0
**상태**: Draft
