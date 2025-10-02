# SPEC-003 구현 계획

## TAG BLOCK
```
PLAN:REFACTOR-003
CHAIN: SPEC:REFACTOR-003 -> TASK:ENFORCER-003 -> TEST:ENFORCER-003
STATUS: active
CREATED: 2025-10-01
```

## 우선순위별 마일스톤

### 1차 목표: 패턴 분리 (tag-patterns.ts)
**목표**: 정규식 패턴과 상수를 별도 모듈로 분리

**작업 항목**:
1. `tag-patterns.ts` 파일 생성
2. `CODE_FIRST_PATTERNS` 상수 이동
3. `VALID_CATEGORIES` 상수 이동
4. `generateTagSuggestions()` 함수 이동
5. `generateImmutabilityHelp()` 함수 이동
6. Export 인터페이스 정의
7. tag-enforcer.ts에서 import 및 참조 업데이트

**완료 조건**:
- tag-patterns.ts ≤ 100 LOC
- 기존 테스트 통과
- 타입 검사 통과

### 2차 목표: 검증 로직 분리 (tag-validator.ts)
**목표**: TAG 검증 로직을 별도 클래스로 분리

**작업 항목**:
1. `tag-validator.ts` 파일 생성
2. `TagValidator` 클래스 생성
3. `checkImmutability()` 메서드 이동
4. `validateCodeFirstTag()` 메서드 이동
5. `extractTagBlock()` 메서드 이동
6. `extractMainTag()` 메서드 이동
7. `normalizeTagBlock()` 메서드 이동
8. 인터페이스 정의 (`TagBlock`, `ImmutabilityCheck`, `ValidationResult`)
9. tag-enforcer.ts에서 TagValidator 인스턴스 사용

**완료 조건**:
- tag-validator.ts ≤ 250 LOC
- 단위 테스트 작성 및 통과
- @IMMUTABLE 보호 로직 100% 유지

### 3차 목표: Hook 슬림화 (tag-enforcer.ts)
**목표**: Hook 진입점만 유지하고 오케스트레이션에 집중

**작업 항목**:
1. CodeFirstTAGEnforcer 클래스 슬림화
2. TagValidator 인스턴스 생성 및 위임
3. 파일 경로/내용 추출 로직만 유지
4. 결과 포맷팅 로직 유지
5. CLI entry point 동작 확인
6. 불필요한 중복 코드 제거

**완료 조건**:
- tag-enforcer.ts ≤ 200 LOC
- Hook 인터페이스 호환성 100% 유지
- CLI 동작 검증

### 4차 목표: 테스트 작성 및 검증
**목표**: 리팩토링 후 전체 기능 검증

**작업 항목**:
1. tag-patterns.ts 단위 테스트
2. tag-validator.ts 단위 테스트
3. tag-enforcer.ts 통합 테스트
4. @IMMUTABLE 수정 시나리오 테스트
5. 유효/무효 TAG 블록 테스트
6. 성능 테스트 (< 100ms)
7. 커버리지 측정 (≥ 85%)

**완료 조건**:
- 모든 테스트 통과
- 커버리지 ≥ 85%
- 성능 요구사항 충족

### 최종 목표: 문서화 및 배포
**목표**: 리팩토링 완료 및 품질 게이트 통과

**작업 항목**:
1. JSDoc 주석 업데이트
2. TAG BLOCK 업데이트 (각 파일)
3. CHANGELOG 업데이트
4. ESLint/Biome 검사
5. TypeScript 타입 검사
6. 코드 복잡도 검증
7. PR 생성 및 리뷰 요청

**완료 조건**:
- 모든 품질 게이트 통과
- 문서화 완료
- PR 승인 및 머지

## 기술적 접근 방법

### 1. 모듈 분리 전략

#### 의존성 방향
```
tag-enforcer.ts (Hook)
  └─> tag-validator.ts (검증 로직)
      └─> tag-patterns.ts (패턴 및 상수)
```

**원칙**:
- 단방향 의존성 (순환 의존 금지)
- 계층적 구조 (Hook → Validator → Patterns)
- 각 모듈은 단일 책임만 가짐

### 2. 코드 이동 체크리스트

#### tag-patterns.ts로 이동
- [ ] `CODE_FIRST_PATTERNS` 상수
- [ ] `VALID_CATEGORIES` 상수
- [ ] `generateTagSuggestions()` 함수
- [ ] `generateImmutabilityHelp()` 함수
- [ ] `getFileType()` 함수 (옵션)

#### tag-validator.ts로 이동
- [ ] `checkImmutability()` 메서드
- [ ] `validateCodeFirstTag()` 메서드
- [ ] `extractTagBlock()` 메서드
- [ ] `extractMainTag()` 메서드
- [ ] `normalizeTagBlock()` 메서드
- [ ] `TagBlock`, `ImmutabilityCheck`, `ValidationResult` 인터페이스

#### tag-enforcer.ts에 유지
- [ ] `CodeFirstTAGEnforcer` 클래스
- [ ] `execute()` 메서드
- [ ] `isWriteOperation()` 메서드
- [ ] `extractFilePath()` 메서드
- [ ] `extractFileContent()` 메서드
- [ ] `shouldEnforceTags()` 메서드
- [ ] `getOriginalFileContent()` 메서드
- [ ] `main()` CLI entry point

### 3. 아키텍처 설계 방향

#### 3.1 tag-patterns.ts (패턴 레이어)
**역할**: 정규식 패턴 및 상수 정의
```typescript
// 순수 상수 및 함수만 포함
export const CODE_FIRST_PATTERNS = { ... };
export const VALID_CATEGORIES = { ... };
export function generateTagSuggestions(...): string { ... }
export function generateImmutabilityHelp(...): string { ... }
```

**특징**:
- 상태 없음 (stateless)
- 순수 함수만 포함
- 외부 의존성 없음

#### 3.2 tag-validator.ts (검증 레이어)
**역할**: TAG 검증 로직 캡슐화
```typescript
// 검증 로직을 담당하는 클래스
export class TagValidator {
  constructor(private patterns = CODE_FIRST_PATTERNS) {}

  checkImmutability(...): ImmutabilityCheck { ... }
  validateCodeFirstTag(...): ValidationResult { ... }
  private extractTagBlock(...): TagBlock | null { ... }
}
```

**특징**:
- 상태 최소화 (패턴 주입 가능)
- 단위 테스트 용이
- tag-patterns 의존

#### 3.3 tag-enforcer.ts (Hook 레이어)
**역할**: Hook 인터페이스 구현 및 오케스트레이션
```typescript
// Hook 진입점만 담당
export class CodeFirstTAGEnforcer implements MoAIHook {
  private validator = new TagValidator();

  async execute(input: HookInput): Promise<HookResult> {
    // 파일 추출 -> 검증 위임 -> 결과 반환
  }
}
```

**특징**:
- Hook 인터페이스 준수
- 검증 로직 위임
- 결과 포맷팅만 담당

### 4. 성능 최적화 전략

#### 4.1 정규식 컴파일 최적화
- 패턴 상수를 모듈 레벨에서 한 번만 컴파일
- 반복 사용되는 패턴은 캐싱

#### 4.2 파일 읽기 최적화
- 비동기 처리 유지
- 불필요한 파일 읽기 제거
- 스트림 기반 처리 고려 (대용량 파일)

#### 4.3 조기 종료 (Early Return)
- TAG 검증 대상이 아니면 즉시 반환
- 불변성 위반 감지 시 즉시 차단
- 순차적 검증 (불필요한 검증 스킵)

### 5. 리스크 및 대응 방안

#### 리스크 1: 순환 의존성
**대응**:
- 의존성 방향을 명확히 정의 (Hook → Validator → Patterns)
- import 체크 (ESLint no-cycle 규칙)

#### 리스크 2: 타입 정의 중복
**대응**:
- 공통 타입은 types.ts에 정의
- 모듈별 필요 타입만 export

#### 리스크 3: 테스트 커버리지 저하
**대응**:
- 리팩토링 전 기존 테스트 확보
- 각 모듈별 단위 테스트 작성
- 통합 테스트 유지

### 6. 코드 품질 체크리스트

#### 개발 가이드 준수
- [ ] 파일당 ≤ 300 LOC
- [ ] 함수당 ≤ 50 LOC
- [ ] 매개변수 ≤ 5개
- [ ] 복잡도 ≤ 10
- [ ] 테스트 커버리지 ≥ 85%

#### TRUST 원칙
- [x] **T**est First: TDD로 리팩토링
- [x] **R**eadable: 의도 드러내는 이름
- [x] **U**nified: TypeScript 타입 안전성
- [x] **S**ecured: 불변성 보호 로직 유지
- [x] **T**rackable: @TAG 체인 유지

#### 코드 스타일
- [ ] ESLint/Biome 검사 통과
- [ ] TypeScript strict 모드
- [ ] JSDoc 주석 완비
- [ ] 일관된 네이밍 컨벤션

## 다음 단계 가이드

### 리팩토링 시작
```bash
# 1. SPEC 검토 및 승인 확인
# 2. 테스트 환경 확인
npm test

# 3. 1차 목표 실행: tag-patterns.ts 분리
/alfred:2-build SPEC-003 --step 1

# 4. 2차 목표 실행: tag-validator.ts 분리
/alfred:2-build SPEC-003 --step 2

# 5. 3차 목표 실행: tag-enforcer.ts 슬림화
/alfred:2-build SPEC-003 --step 3

# 6. 4차 목표 실행: 테스트 작성
/alfred:2-build SPEC-003 --step 4

# 7. 최종 목표: 문서화 및 배포
/alfred:3-sync
```

### 검증 방법
```bash
# LOC 측정
wc -l src/claude/hooks/tag-enforcer.ts
wc -l src/claude/hooks/tag-validator.ts
wc -l src/claude/hooks/tag-patterns.ts

# 테스트 실행
npm test

# 커버리지 측정
npm run test:coverage

# 타입 검사
npm run type-check

# 린터 검사
npm run lint

# 빌드 확인
npm run build
```

### 롤백 계획
```bash
# 문제 발생 시 원본 파일로 복구
git checkout HEAD -- src/claude/hooks/tag-enforcer.ts

# 또는 커밋 단위로 롤백
git revert <commit-hash>
```

---

**작성일**: 2025-10-01
**작성자**: @agent-spec-builder
**상태**: Draft (사용자 검토 대기)
