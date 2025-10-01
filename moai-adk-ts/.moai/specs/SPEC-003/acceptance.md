# SPEC-003 수락 기준

## TAG BLOCK
```
ACCEPTANCE:REFACTOR-003
CHAIN: SPEC:REFACTOR-003 -> TASK:ENFORCER-003 -> TEST:ENFORCER-003
STATUS: active
CREATED: 2025-10-01
```

## Given-When-Then 테스트 시나리오

### 시나리오 1: 패턴 분리 (tag-patterns.ts)

#### Given
- 기존 tag-enforcer.ts 파일에 정규식 패턴과 상수가 포함되어 있음
- CODE_FIRST_PATTERNS와 VALID_CATEGORIES 상수가 정의되어 있음

#### When
- tag-patterns.ts 파일을 생성하고 패턴 및 상수를 이동함

#### Then
- [ ] tag-patterns.ts 파일이 생성되어 있다
- [ ] tag-patterns.ts의 LOC가 100 이하다
- [ ] CODE_FIRST_PATTERNS가 모든 필수 패턴을 포함한다:
  - `TAG_BLOCK`, `MAIN_TAG`, `CHAIN_LINE`, `DEPENDS_LINE`
  - `STATUS_LINE`, `CREATED_LINE`, `IMMUTABLE_MARKER`, `TAG_REFERENCE`
- [ ] VALID_CATEGORIES가 올바른 카테고리를 정의한다:
  - lifecycle: `['SPEC', 'REQ', 'DESIGN', 'TASK', 'TEST']`
  - implementation: `['FEATURE', 'API', 'FIX']`
- [ ] 모든 패턴과 상수가 export되어 있다
- [ ] tag-enforcer.ts에서 import하여 사용할 수 있다
- [ ] 기존 테스트가 모두 통과한다

**검증 명령어**:
```bash
# LOC 확인
wc -l src/claude/hooks/tag-patterns.ts
# 예상: ≤ 100 LOC

# import 검증
grep "import.*tag-patterns" src/claude/hooks/tag-enforcer.ts
# 예상: import 문 존재

# 빌드 검증
npm run build
# 예상: 타입 에러 없음

# 테스트 실행
npm test -- tag-enforcer
# 예상: 모든 테스트 통과
```

---

### 시나리오 2: 검증 로직 분리 (tag-validator.ts)

#### Given
- tag-patterns.ts가 이미 분리되어 있음
- tag-enforcer.ts에 검증 로직이 포함되어 있음

#### When
- tag-validator.ts 파일을 생성하고 TagValidator 클래스를 구현함
- checkImmutability, validateCodeFirstTag 등의 메서드를 이동함

#### Then
- [ ] tag-validator.ts 파일이 생성되어 있다
- [ ] tag-validator.ts의 LOC가 250 이하다
- [ ] TagValidator 클래스가 다음 메서드를 포함한다:
  - `checkImmutability(oldContent, newContent, filePath)`
  - `validateCodeFirstTag(content)`
  - `extractTagBlock(content)` (private)
  - `extractMainTag(blockContent)` (private)
  - `normalizeTagBlock(blockContent)` (private)
- [ ] 인터페이스가 올바르게 정의되어 있다:
  - `TagBlock`, `ImmutabilityCheck`, `ValidationResult`
- [ ] TagValidator가 tag-patterns를 import한다
- [ ] 단위 테스트가 작성되어 있다 (커버리지 ≥ 85%)
- [ ] @IMMUTABLE 보호 로직이 100% 동작한다

**검증 명령어**:
```bash
# LOC 확인
wc -l src/claude/hooks/tag-validator.ts
# 예상: ≤ 250 LOC

# 클래스 구조 확인
grep "class TagValidator" src/claude/hooks/tag-validator.ts
# 예상: 클래스 정의 존재

# 메서드 확인
grep -E "(checkImmutability|validateCodeFirstTag)" src/claude/hooks/tag-validator.ts
# 예상: 두 메서드 모두 존재

# 단위 테스트 실행
npm test -- tag-validator.test
# 예상: 모든 테스트 통과

# 커버리지 확인
npm run test:coverage -- tag-validator.ts
# 예상: ≥ 85%
```

---

### 시나리오 3: Hook 슬림화 (tag-enforcer.ts)

#### Given
- tag-patterns.ts와 tag-validator.ts가 이미 분리되어 있음
- tag-enforcer.ts에 불필요한 코드가 남아 있음

#### When
- tag-enforcer.ts를 슬림화하여 Hook 진입점만 유지함
- TagValidator 인스턴스를 생성하고 검증 로직을 위임함

#### Then
- [ ] tag-enforcer.ts의 LOC가 200 이하다
- [ ] CodeFirstTAGEnforcer가 MoAIHook 인터페이스를 구현한다
- [ ] execute() 메서드가 다음 작업만 수행한다:
  - 파일 쓰기 작업 확인
  - 파일 경로 및 내용 추출
  - TagValidator에 검증 위임
  - 결과 포맷팅 및 반환
- [ ] TagValidator 인스턴스가 생성되어 있다
- [ ] CLI entry point (main 함수)가 정상 동작한다
- [ ] Hook 실행 시간이 100ms 이하다
- [ ] 기존 Hook 사용자에게 영향이 없다

**검증 명령어**:
```bash
# LOC 확인
wc -l src/claude/hooks/tag-enforcer.ts
# 예상: ≤ 200 LOC

# MoAIHook 인터페이스 구현 확인
grep "implements MoAIHook" src/claude/hooks/tag-enforcer.ts
# 예상: 구현 선언 존재

# TagValidator 사용 확인
grep "new TagValidator()" src/claude/hooks/tag-enforcer.ts
# 예상: 인스턴스 생성 코드 존재

# CLI 동작 확인
node dist/claude/hooks/tag-enforcer.js --help
# 예상: 정상 출력

# 성능 측정
npm run test:perf -- tag-enforcer
# 예상: < 100ms
```

---

### 시나리오 4: @IMMUTABLE TAG 보호 (핵심 기능)

#### Given
- 파일에 @IMMUTABLE 마커가 있는 TAG 블록이 존재함

#### When
- 해당 TAG 블록을 수정하려고 시도함

#### Then
- [ ] Hook이 즉시 차단한다 (blocked: true)
- [ ] 명확한 에러 메시지를 출력한다:
  - "🚫 @IMMUTABLE TAG 수정 금지"
  - 위반 세부사항 표시
- [ ] 개선 제안을 제공한다:
  - 새로운 TAG ID로 기능 구현 권장
  - @DOC 마커 추가 방법 안내
  - REPLACES 참조 방법 안내
- [ ] 수정된 TAG ID를 명시한다
- [ ] Exit code가 2다 (차단 상태)

**검증 명령어**:
```bash
# 테스트 시나리오 실행
npm test -- "should block @IMMUTABLE TAG modification"
# 예상: 테스트 통과

# 실제 파일로 검증
echo "modified content" > /tmp/test-immutable.ts
# 예상: Hook이 차단하고 에러 메시지 출력
```

**테스트 케이스**:
```typescript
describe('IMMUTABLE TAG Protection', () => {
  test('should block TAG deletion', async () => {
    const oldContent = `
      /**
       * @DOC:FEATURE:AUTH-001
       * @IMMUTABLE
       */
    `;
    const newContent = `// TAG 블록 삭제됨`;

    const result = await enforcer.execute(createHookInput(oldContent, newContent));

    expect(result.blocked).toBe(true);
    expect(result.message).toContain('@IMMUTABLE TAG 블록이 삭제');
    expect(result.exitCode).toBe(2);
  });

  test('should block TAG content modification', async () => {
    const oldContent = `
      /**
       * @DOC:FEATURE:AUTH-001
       * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001
       * @IMMUTABLE
       */
    `;
    const newContent = `
      /**
       * @DOC:FEATURE:AUTH-002  // ID 변경
       * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001
       * @IMMUTABLE
       */
    `;

    const result = await enforcer.execute(createHookInput(oldContent, newContent));

    expect(result.blocked).toBe(true);
    expect(result.message).toContain('@IMMUTABLE TAG 블록의 내용이 변경');
  });

  test('should allow new file with TAG', async () => {
    const oldContent = ''; // 새 파일
    const newContent = `
      /**
       * @DOC:FEATURE:NEW-001
       * @IMMUTABLE
       */
    `;

    const result = await enforcer.execute(createHookInput(oldContent, newContent));

    expect(result.success).toBe(true);
    expect(result.blocked).toBe(false);
  });
});
```

---

### 시나리오 5: TAG 유효성 검증

#### Given
- 파일에 잘못된 형식의 TAG 블록이 포함되어 있음

#### When
- TAG 블록을 검증함

#### Then
- [ ] 필수 요소 누락 시 violations를 반환한다
- [ ] 형식 오류 시 warnings를 반환한다
- [ ] 유효한 TAG 카테고리만 허용한다 (8-Core 체계)
- [ ] 도메인 ID 형식을 검증한다 (`DOMAIN-001`)
- [ ] 체인 TAG 형식을 검증한다 (`@SPEC:ID -> @SPEC:ID`)
- [ ] 의존성 TAG 형식을 검증한다 (`@CODE:ID:API, @CODE:ID:DATA`)
- [ ] STATUS 값을 검증한다 (`active`, `deprecated`, `completed`)
- [ ] 생성 날짜 형식을 검증한다 (`YYYY-MM-DD`)

**테스트 케이스**:
```typescript
describe('TAG Validation', () => {
  test('should validate TAG category', () => {
    const content = `
      /**
       * @DOC:INVALID:TEST-001
       */
    `;

    const result = validator.validateCodeFirstTag(content);

    expect(result.isValid).toBe(false);
    expect(result.violations).toContain('유효하지 않은 TAG 카테고리: INVALID');
  });

  test('should validate domain ID format', () => {
    const content = `
      /**
       * @DOC:FEATURE:invalid-id
       */
    `;

    const result = validator.validateCodeFirstTag(content);

    expect(result.warnings).toContainEqual(
      expect.stringContaining('도메인 ID 형식 권장')
    );
  });

  test('should validate chain format', () => {
    const content = `
      /**
       * @DOC:FEATURE:TEST-001
       * CHAIN: INVALID_FORMAT
       */
    `;

    const result = validator.validateCodeFirstTag(content);

    expect(result.warnings).toContainEqual(
      expect.stringContaining('체인의 TAG 형식을 확인')
    );
  });

  test('should allow valid TAG', () => {
    const content = `
      /**
       * @DOC:FEATURE:AUTH-001
       * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001
       * DEPENDS: NONE
       * STATUS: active
       * CREATED: 2025-10-01
       * @IMMUTABLE
       */
    `;

    const result = validator.validateCodeFirstTag(content);

    expect(result.isValid).toBe(true);
    expect(result.violations).toHaveLength(0);
  });
});
```

---

### 시나리오 6: 성능 요구사항

#### Given
- 다양한 크기의 파일에 대해 Hook을 실행함

#### When
- Hook 실행 시간을 측정함

#### Then
- [ ] 소형 파일 (< 100 LOC): < 10ms
- [ ] 중형 파일 (100-500 LOC): < 50ms
- [ ] 대형 파일 (> 500 LOC): < 100ms
- [ ] 평균 실행 시간: < 50ms
- [ ] 메모리 사용량이 증가하지 않음

**검증 명령어**:
```bash
# 성능 테스트 실행
npm run test:perf

# 벤치마크 결과 확인
# 예상:
# Small file (50 LOC): ~5ms
# Medium file (300 LOC): ~30ms
# Large file (1000 LOC): ~80ms
```

**성능 테스트 코드**:
```typescript
describe('Performance', () => {
  test('should execute within 100ms for large file', async () => {
    const largeContent = generateLargeFile(1000); // 1000 LOC

    const startTime = performance.now();
    await enforcer.execute(createHookInput('', largeContent));
    const endTime = performance.now();

    const duration = endTime - startTime;
    expect(duration).toBeLessThan(100);
  });

  test('should not increase memory usage', async () => {
    const initialMemory = process.memoryUsage().heapUsed;

    // 100번 실행
    for (let i = 0; i < 100; i++) {
      await enforcer.execute(createHookInput('', sampleContent));
    }

    const finalMemory = process.memoryUsage().heapUsed;
    const memoryIncrease = finalMemory - initialMemory;

    // 메모리 증가량이 5MB 이하
    expect(memoryIncrease).toBeLessThan(5 * 1024 * 1024);
  });
});
```

---

## 품질 게이트 기준

### 1. 코드 품질

#### LOC 제한
```bash
# 각 파일의 LOC 확인
wc -l src/claude/hooks/tag-enforcer.ts   # ≤ 200 LOC
wc -l src/claude/hooks/tag-validator.ts  # ≤ 250 LOC
wc -l src/claude/hooks/tag-patterns.ts   # ≤ 100 LOC

# 총합 확인
wc -l src/claude/hooks/tag-*.ts          # ≤ 550 LOC
```

**통과 조건**: 모든 파일이 제한 이하

#### 함수 복잡도
```bash
# ESLint 복잡도 검사
npm run lint -- --rule "complexity: [error, 10]"
```

**통과 조건**: 모든 함수 복잡도 ≤ 10

#### 매개변수 개수
```bash
# ESLint 매개변수 검사
npm run lint -- --rule "max-params: [error, 5]"
```

**통과 조건**: 모든 함수 매개변수 ≤ 5개

### 2. 테스트 커버리지

```bash
# 커버리지 측정
npm run test:coverage

# 기준:
# Statements: ≥ 85%
# Branches: ≥ 80%
# Functions: ≥ 85%
# Lines: ≥ 85%
```

**통과 조건**:
- [ ] Statements ≥ 85%
- [ ] Branches ≥ 80%
- [ ] Functions ≥ 85%
- [ ] Lines ≥ 85%

### 3. 타입 안전성

```bash
# TypeScript strict 모드 검사
npm run type-check
```

**통과 조건**:
- [ ] 타입 에러 0개
- [ ] any 타입 사용 0개 (불가피한 경우 제외)
- [ ] 모든 함수에 반환 타입 명시

### 4. 코드 스타일

```bash
# Biome 또는 ESLint 검사
npm run lint
```

**통과 조건**:
- [ ] 린터 에러 0개
- [ ] 경고 최소화 (≤ 5개)

### 5. 빌드 검증

```bash
# 빌드 실행
npm run build

# 빌드 산출물 확인
ls -la dist/claude/hooks/
```

**통과 조건**:
- [ ] 빌드 성공
- [ ] 모든 파일 정상 생성
- [ ] 빌드 경고 없음

---

## 완료 조건 (Definition of Done)

### 필수 조건 체크리스트

#### 코드 분리
- [ ] tag-enforcer.ts ≤ 200 LOC
- [ ] tag-validator.ts ≤ 250 LOC
- [ ] tag-patterns.ts ≤ 100 LOC
- [ ] 총 3개 파일로 분리 완료

#### 기능 유지
- [ ] @IMMUTABLE TAG 보호 100% 동작
- [ ] TAG 유효성 검증 100% 동작
- [ ] Hook 인터페이스 호환성 100%
- [ ] CLI entry point 정상 동작

#### 성능
- [ ] Hook 실행 시간 < 100ms
- [ ] 메모리 사용량 증가 없음
- [ ] 정규식 최적화 적용

#### 테스트
- [ ] 단위 테스트 작성 완료
- [ ] 통합 테스트 작성 완료
- [ ] 성능 테스트 작성 완료
- [ ] 모든 테스트 통과
- [ ] 커버리지 ≥ 85%

#### 품질 게이트
- [ ] 타입 검사 통과
- [ ] 린터 검사 통과
- [ ] 빌드 검증 통과
- [ ] 함수당 ≤ 50 LOC
- [ ] 매개변수 ≤ 5개
- [ ] 복잡도 ≤ 10

#### 문서화
- [ ] JSDoc 주석 업데이트
- [ ] TAG BLOCK 추가 (각 파일)
- [ ] CHANGELOG 업데이트
- [ ] 이 SPEC 문서 검증 완료

---

## 수락 검증 절차

### 1단계: 자동 검증
```bash
#!/bin/bash
# acceptance-test.sh

echo "=== SPEC-003 수락 검증 시작 ==="

# LOC 검증
echo "1. LOC 검증..."
LOC_ENFORCER=$(wc -l < src/claude/hooks/tag-enforcer.ts)
LOC_VALIDATOR=$(wc -l < src/claude/hooks/tag-validator.ts)
LOC_PATTERNS=$(wc -l < src/claude/hooks/tag-patterns.ts)

if [ $LOC_ENFORCER -le 200 ] && [ $LOC_VALIDATOR -le 250 ] && [ $LOC_PATTERNS -le 100 ]; then
  echo "✅ LOC 검증 통과"
else
  echo "❌ LOC 검증 실패: Enforcer=$LOC_ENFORCER, Validator=$LOC_VALIDATOR, Patterns=$LOC_PATTERNS"
  exit 1
fi

# 테스트 검증
echo "2. 테스트 검증..."
npm test || exit 1
echo "✅ 테스트 통과"

# 커버리지 검증
echo "3. 커버리지 검증..."
npm run test:coverage -- --coverage-threshold=85 || exit 1
echo "✅ 커버리지 통과"

# 타입 검증
echo "4. 타입 검증..."
npm run type-check || exit 1
echo "✅ 타입 검사 통과"

# 린터 검증
echo "5. 린터 검증..."
npm run lint || exit 1
echo "✅ 린터 검사 통과"

# 빌드 검증
echo "6. 빌드 검증..."
npm run build || exit 1
echo "✅ 빌드 통과"

echo "=== 모든 검증 통과 🎉 ==="
```

### 2단계: 수동 검증
1. **코드 리뷰**: 모듈 분리가 적절한지 확인
2. **기능 테스트**: @IMMUTABLE TAG 보호 실제 동작 확인
3. **성능 테스트**: 실제 파일로 성능 측정
4. **문서 검토**: JSDoc 및 TAG BLOCK 완성도 확인

### 3단계: 승인
- [ ] 모든 자동 검증 통과
- [ ] 코드 리뷰 승인
- [ ] 기능 테스트 승인
- [ ] 문서 검토 완료

---

**작성일**: 2025-10-01
**작성자**: @agent-spec-builder
**상태**: Draft (사용자 검토 대기)
