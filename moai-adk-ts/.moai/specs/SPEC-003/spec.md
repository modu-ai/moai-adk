# SPEC-003: tag-enforcer.ts 리팩토링

## TAG BLOCK
```
@SPEC:REFACTOR-003
CHAIN: REQ:TAG-001 -> DESIGN:CODE-FIRST-001 -> TASK:ENFORCER-003 -> TEST:ENFORCER-003
DEPENDS: @CODE:HOOK-004, @CODE:TAG-001
STATUS: active
CREATED: 2025-10-01
@IMMUTABLE
```

## Environment (환경 및 가정사항)

### 현재 환경
- **파일**: `src/claude/hooks/tag-enforcer.ts`
- **현재 LOC**: 554 라인
- **상태**: 개발 가이드 권장(300 LOC) 대비 185% 초과
- **역할**: Code-First TAG 불변성 검증 및 8-Core TAG 체계 검증
- **의존성**: Claude Code Hook 인터페이스, fs/promises, path

### 가정사항
- 기존 Hook 인터페이스(`HookInput`, `HookResult`, `MoAIHook`)는 변경하지 않음
- @IMMUTABLE TAG 보호 로직은 100% 기능 유지
- Hook 실행 시간 목표는 100ms 이하
- 테스트 커버리지 85% 이상 유지

## Assumptions (전제 조건)

### 기술적 전제
- TypeScript 5.x 환경
- Node.js 파일 시스템 API 사용
- 정규식 기반 TAG 패턴 매칭
- 동기/비동기 혼합 파일 처리

### 비즈니스 전제
- TAG 불변성은 MoAI-ADK의 핵심 원칙
- 300 LOC 제한은 코드 품질 유지를 위한 필수 규칙
- 리팩토링 후에도 기존 Hook 사용자에게 영향 없음

## Requirements (기능 요구사항)

### Ubiquitous Requirements (기본 요구사항)
- 시스템은 tag-enforcer.ts를 300 LOC 이하의 3개 모듈로 분리해야 한다
- 시스템은 @IMMUTABLE TAG 보호 로직을 100% 유지해야 한다
- 시스템은 Claude Code Hook 인터페이스 호환성을 보장해야 한다
- 시스템은 기존 TAG 검증 기능을 모두 유지해야 한다

### Event-driven Requirements (이벤트 기반 요구사항)
- WHEN @IMMUTABLE TAG가 수정되면, 시스템은 즉시 차단해야 한다
- WHEN TAG 형식 오류가 감지되면, 시스템은 명확한 개선 제안을 제공해야 한다
- WHEN 파일 쓰기 작업이 발생하면, 시스템은 TAG 유효성을 검증해야 한다
- WHEN 새 파일이 생성되면, 시스템은 TAG 블록 템플릿을 제안해야 한다

### State-driven Requirements (상태 기반 요구사항)
- WHILE 리팩토링 중일 때, 시스템은 기존 기능을 100% 보존해야 한다
- WHILE 개발 모드일 때, 시스템은 상세한 검증 경고를 출력해야 한다
- WHILE Hook이 실행 중일 때, 시스템은 100ms 이내에 응답해야 한다

### Optional Features (선택적 기능)
- WHERE 성능 개선이 가능하면, 시스템은 정규식 패턴을 최적화할 수 있다
- WHERE 확장성이 필요하면, 시스템은 플러그인 방식 검증기를 지원할 수 있다

### Constraints (제약사항)
- IF 모듈 파일이 300 LOC를 초과하면, 추가 분리가 필요하다
- Hook 실행 시간은 100ms를 초과하지 않아야 한다
- 정규식 패턴 복잡도는 최소화해야 한다 (복잡도 ≤ 10)
- 각 함수는 50 LOC 이하여야 한다
- 매개변수는 5개 이하여야 한다

## Specifications (상세 명세)

### 1. 모듈 분리 구조

#### 1.1 tag-enforcer.ts (메인 Hook, ~200 LOC)
**책임**: Hook 진입점 및 오케스트레이션
- `CodeFirstTAGEnforcer` 클래스 (MoAIHook 인터페이스 구현)
- `execute()` 메서드: 파일 쓰기 작업 인터셉트
- 파일 경로 추출 및 내용 추출 로직
- TAG 검증 대상 파일 필터링
- 결과 포맷팅 및 반환

**주요 함수**:
```typescript
class CodeFirstTAGEnforcer implements MoAIHook {
  async execute(input: HookInput): Promise<HookResult>
  private isWriteOperation(toolName?: string): boolean
  private extractFilePath(toolInput: Record<string, any>): string | null
  private extractFileContent(toolInput: Record<string, any>): string
  private shouldEnforceTags(filePath: string): boolean
  private async getOriginalFileContent(filePath: string): Promise<string>
}
```

#### 1.2 tag-validator.ts (검증 로직, ~250 LOC)
**책임**: TAG 불변성 검증 및 유효성 검사
- @IMMUTABLE TAG 블록 수정 감지
- TAG 블록 추출 및 파싱
- Code-First TAG 유효성 검증
- 8-Core TAG 카테고리 검증
- 체인 및 의존성 검증

**주요 함수**:
```typescript
export class TagValidator {
  checkImmutability(oldContent: string, newContent: string, filePath: string): ImmutabilityCheck
  validateCodeFirstTag(content: string): ValidationResult
  private extractTagBlock(content: string): TagBlock | null
  private extractMainTag(blockContent: string): string
  private normalizeTagBlock(blockContent: string): string
  private validateChain(chainStr: string): string[]
  private validateDependencies(dependsStr: string): string[]
}
```

#### 1.3 tag-patterns.ts (정규식 패턴, ~100 LOC)
**책임**: TAG 패턴 정의 및 매칭 유틸리티
- CODE_FIRST_PATTERNS 상수 정의
- VALID_CATEGORIES 상수 정의
- TAG 제안 생성 함수
- 도움말 메시지 생성 함수

**주요 내용**:
```typescript
export const CODE_FIRST_PATTERNS = {
  TAG_BLOCK: RegExp,
  MAIN_TAG: RegExp,
  CHAIN_LINE: RegExp,
  DEPENDS_LINE: RegExp,
  STATUS_LINE: RegExp,
  CREATED_LINE: RegExp,
  IMMUTABLE_MARKER: RegExp,
  TAG_REFERENCE: RegExp
};

export const VALID_CATEGORIES = {
  lifecycle: string[],
  implementation: string[]
};

export function generateTagSuggestions(filePath: string, content: string): string
export function generateImmutabilityHelp(immutabilityCheck: ImmutabilityCheck): string
```

### 2. 인터페이스 정의 (types.ts에 추가)

```typescript
/**
 * TAG 블록 추출 결과
 */
export interface TagBlock {
  content: string;
  lineNumber: number;
}

/**
 * 불변성 검사 결과
 */
export interface ImmutabilityCheck {
  violated: boolean;
  modifiedTag?: string;
  violationDetails?: string;
}

/**
 * TAG 유효성 검증 결과
 */
export interface ValidationResult {
  isValid: boolean;
  violations: string[];
  warnings: string[];
  hasTag: boolean;
}
```

### 3. 리팩토링 원칙

#### 3.1 단일 책임 원칙 (SRP)
- **tag-enforcer.ts**: Hook 오케스트레이션만 담당
- **tag-validator.ts**: 검증 로직만 담당
- **tag-patterns.ts**: 패턴 정의 및 메시지 생성만 담당

#### 3.2 개방-폐쇄 원칙 (OCP)
- 새로운 TAG 카테고리 추가 시 VALID_CATEGORIES만 수정
- 새로운 검증 규칙 추가 시 TagValidator에 메서드 추가
- Hook 인터페이스는 수정하지 않고 확장 가능

#### 3.3 의존성 역전 원칙 (DIP)
- CodeFirstTAGEnforcer는 TagValidator에 의존
- TagValidator는 tag-patterns에 의존
- 순환 의존성 없음

### 4. 성능 요구사항

- Hook 실행 시간: < 100ms (기존과 동일)
- 정규식 컴파일: 초기화 시 1회만 수행
- 파일 읽기: 비동기 처리 유지
- 메모리 사용: 증가하지 않음

### 5. 테스트 전략

#### 5.1 단위 테스트
- `TagValidator.checkImmutability()` 테스트
- `TagValidator.validateCodeFirstTag()` 테스트
- `extractTagBlock()` 엣지 케이스 테스트
- 정규식 패턴 매칭 테스트

#### 5.2 통합 테스트
- `CodeFirstTAGEnforcer.execute()` 통합 테스트
- @IMMUTABLE TAG 수정 시나리오 테스트
- 유효/무효한 TAG 블록 시나리오 테스트

#### 5.3 성능 테스트
- Hook 실행 시간 측정 (< 100ms)
- 대용량 파일 처리 테스트

### 6. 마이그레이션 가이드

#### 6.1 단계별 리팩토링
1. **1단계**: tag-patterns.ts 분리 (패턴 및 상수)
2. **2단계**: tag-validator.ts 분리 (검증 로직)
3. **3단계**: tag-enforcer.ts 슬림화 (Hook 진입점만 유지)
4. **4단계**: 테스트 작성 및 검증
5. **5단계**: 성능 측정 및 최적화

#### 6.2 호환성 보장
- 기존 Hook 인터페이스 유지
- CLI entry point (`main()`) 동작 유지
- 에러 메시지 및 제안 포맷 유지

## Traceability (추적성)

### TAG 체인
```
@SPEC:REFACTOR-003
  └─> @SPEC:TAG-001 (CODE-FIRST TAG 체계 요구사항)
      └─> @SPEC:CODE-FIRST-001 (불변성 보장 설계)
          └─> @CODE:ENFORCER-003 (리팩토링 작업)
              └─> @TEST:ENFORCER-003 (검증 테스트)
```

### 관련 SPEC
- **SPEC-001**: Claude Code Hook 시스템 통합
- **SPEC-002**: 8-Core TAG 체계 정의

### 관련 파일
- `src/claude/hooks/tag-enforcer.ts` (현재 554 LOC)
- `src/claude/types.ts` (Hook 인터페이스)
- `src/claude/index.ts` (Hook 등록)

## Definition of Done (완료 조건)

### 필수 조건
- [ ] tag-enforcer.ts ≤ 200 LOC
- [ ] tag-validator.ts ≤ 250 LOC
- [ ] tag-patterns.ts ≤ 100 LOC
- [ ] 모든 기존 기능 100% 동작
- [ ] @IMMUTABLE TAG 보호 로직 100% 유지
- [ ] Hook 실행 시간 < 100ms
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 테스트 통과 (단위 + 통합)

### 품질 게이트
- [ ] ESLint 또는 Biome 검사 통과
- [ ] TypeScript 타입 검사 통과
- [ ] 함수당 ≤ 50 LOC
- [ ] 매개변수 ≤ 5개
- [ ] 복잡도 ≤ 10

### 문서화
- [ ] JSDoc 주석 업데이트
- [ ] TAG BLOCK 업데이트 (각 파일)
- [ ] CHANGELOG 업데이트

## Risk & Mitigation (리스크 및 대응 방안)

### 리스크 1: 기능 손실
**완화 방안**:
- TDD로 기존 기능 테스트 먼저 작성
- 리팩토링 중 테스트 지속 실행
- 단계별 검증 (1단계 완료 후 테스트)

### 리스크 2: 성능 저하
**완화 방안**:
- 각 단계마다 성능 측정
- 정규식 컴파일 최적화
- 불필요한 파일 읽기 제거

### 리스크 3: 호환성 문제
**완화 방안**:
- Hook 인터페이스 변경 금지
- CLI entry point 동작 유지
- 기존 사용자 시나리오 테스트

## Success Metrics (성공 지표)

### 정량적 지표
- **LOC 감소**: 554 LOC → 550 LOC 이하 (3개 파일 합계)
- **모듈당 LOC**: 각 파일 ≤ 300 LOC
- **실행 시간**: < 100ms
- **테스트 커버리지**: ≥ 85%

### 정성적 지표
- 코드 가독성 향상
- 모듈 간 책임 명확화
- 유지보수성 개선
- 확장 용이성 확보

---

**작성일**: 2025-10-01
**작성자**: @agent-spec-builder
**상태**: Draft (사용자 검토 대기)
