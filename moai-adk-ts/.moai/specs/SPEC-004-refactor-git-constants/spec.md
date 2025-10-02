# SPEC-004: Git Constants 파일 분리 리팩토링

## @TAG BLOCK

```text
# @CODE:REFACTOR-004 | Chain: @SPEC:CODE-QUALITY-004 -> @SPEC:REFACTOR-004 -> @CODE:SPLIT-CONSTANTS-004 -> @TEST:REFACTOR-004
# Related: @CODE:GIT-CONSTANTS-004:DATA, @CODE:GIT-CONFIG-004:API
```

---

## 1. Environment (환경 및 가정사항)

### 현재 환경

- **대상 파일**: `src/core/git/constants.ts` (454 LOC)
- **프로젝트**: MoAI-ADK TypeScript 구현체
- **개발 가이드**: 파일당 300 LOC 이하 권장
- **초과율**: 151% (454/300)

### 가정사항

- 기존 import 경로를 사용하는 코드가 다수 존재
- barrel export를 통한 호환성 유지 필요
- TypeScript `as const` 타입 안전성 유지
- @TAG 체인 연속성 보존

---

## 2. Requirements (기능 요구사항)

### @SPEC:CODE-QUALITY-004 코드 품질 개선

#### Ubiquitous Requirements (기본 요구사항)

- 시스템은 constants.ts를 논리적 그룹별로 3개 파일로 분리해야 한다
- 시스템은 각 파일이 300 LOC 이하를 유지해야 한다
- 시스템은 타입 안전성을 유지해야 한다 (`as const` 보존)
- 시스템은 @TAG 체인 연속성을 보존해야 한다

#### Event-driven Requirements (이벤트 기반)

- WHEN 기존 코드가 `import { GitDefaults } from '@/core/git/constants'`를 사용하면, 시스템은 정상 동작을 보장해야 한다
- WHEN 새 코드가 개별 파일에서 import하면, 시스템은 해당 상수만 제공해야 한다

#### State-driven Requirements (상태 기반)

- WHILE 리팩토링 진행 중일 때, 시스템은 기존 테스트가 통과해야 한다
- WHILE barrel export 사용 중일 때, 시스템은 순환 의존성이 없어야 한다

#### Constraints (제약사항)

- IF 파일이 300 LOC를 초과하면, 시스템은 추가 분리를 요구해야 한다
- 각 상수 파일은 단일 책임 원칙을 준수해야 한다
- 모든 export는 named export만 허용 (default export 금지)

---

## 3. Specifications (상세 명세)

### @SPEC:REFACTOR-004 분리 설계

#### 3.1 파일 구조

```
src/core/git/constants/
├── index.ts                  # barrel export (~20 LOC)
├── branch-constants.ts       # GitNamingRules (~100 LOC)
├── commit-constants.ts       # GitCommitTemplates (~150 LOC)
└── config-constants.ts       # GitignoreTemplates, GitDefaults, GitHubDefaults, GitTimeouts (~200 LOC)
```

#### 3.2 branch-constants.ts

**책임**: Git 브랜치 명명 규칙 및 검증 로직

**포함 내용**:
- `GitNamingRules` 객체 전체
- 브랜치 prefix 상수
- 브랜치명 생성 함수 (feature, spec, bugfix, hotfix)
- 브랜치명 검증 함수

**예상 LOC**: ~100 라인

**TAG**: `@CODE:GIT-NAMING-RULES-001:DATA`

#### 3.3 commit-constants.ts

**책임**: Git 커밋 메시지 템플릿 및 포맷팅

**포함 내용**:
- `GitCommitTemplates` 객체 전체
- 커밋 타입별 템플릿
- 이모지 매핑
- 메시지 생성 함수 (apply, createAutoCommit, createCheckpoint)
- 타입별 이모지 반환 함수

**예상 LOC**: ~150 라인

**TAG**: `@CODE:GIT-COMMIT-TEMPLATES-001:DATA`

#### 3.4 config-constants.ts

**책임**: Git 및 GitHub 설정, .gitignore 템플릿, 타임아웃

**포함 내용**:
- `GitignoreTemplates` (MOAI, NODE, PYTHON)
- `GitDefaults` (기본 브랜치, 리모트, 명령어 목록)
- `GitHubDefaults` (API URL, PR/Issue 템플릿, 라벨)
- `GitTimeouts` (작업별 타임아웃)

**예상 LOC**: ~200 라인

**TAG**: `@CODE:GIT-DEFAULTS-001:DATA`, `@CODE:GITHUB-DEFAULTS-001:DATA`, `@CODE:GIT-TIMEOUTS-001:DATA`, `@CODE:GITIGNORE-TEMPLATES-001:DATA`

#### 3.5 index.ts (barrel export)

**책임**: 모든 상수를 re-export하여 기존 import 경로 유지

```typescript
// @CODE:REFACTOR-004 | Chain: @SPEC:CODE-QUALITY-004 -> @SPEC:REFACTOR-004
export * from './branch-constants';
export * from './commit-constants';
export * from './config-constants';
```

**예상 LOC**: ~20 라인

---

## 4. Traceability (추적성)

### @CODE:SPLIT-CONSTANTS-004 구현 작업

#### 4.1 마일스톤 (우선순위 기반)

**1차 목표**: 파일 분리 및 barrel export 구성
- [ ] `constants/` 디렉토리 생성
- [ ] `branch-constants.ts` 파일 생성 및 GitNamingRules 이동
- [ ] `commit-constants.ts` 파일 생성 및 GitCommitTemplates 이동
- [ ] `config-constants.ts` 파일 생성 및 나머지 상수 이동
- [ ] `index.ts` barrel export 작성

**2차 목표**: 기존 코드 호환성 검증
- [ ] 기존 import 경로가 정상 작동하는지 확인
- [ ] 타입 추론이 올바르게 작동하는지 확인
- [ ] 순환 의존성이 없는지 확인

**3차 목표**: 원본 파일 정리
- [ ] 기존 `constants.ts` 파일 삭제
- [ ] import 경로 업데이트 (필요 시)

#### 4.2 기술적 접근 방법

**타입 안전성 보존**:
```typescript
// as const 어서션 유지
export const GitNamingRules = {
  // ...
} as const;
```

**barrel export 패턴**:
```typescript
// index.ts
export * from './branch-constants';
export * from './commit-constants';
export * from './config-constants';
```

**import 경로 호환성**:
```typescript
// 기존 방식 (유지)
import { GitDefaults } from '@/core/git/constants';

// 새로운 방식 (선택적)
import { GitDefaults } from '@/core/git/constants/config-constants';
```

#### 4.3 리스크 및 대응 방안

| 리스크 | 발생 가능성 | 영향도 | 완화 방안 |
|--------|-------------|--------|-----------|
| 순환 의존성 발생 | 낮음 | 높음 | 상수는 다른 파일을 import하지 않도록 설계 |
| 타입 추론 실패 | 낮음 | 중간 | as const 유지, 타입 테스트 추가 |
| 기존 코드 호환성 깨짐 | 중간 | 높음 | barrel export로 기존 경로 유지 |
| 파일 크기 불균형 | 중간 | 낮음 | LOC 기준 재검토 및 조정 |

---

## 5. Acceptance Criteria (수락 기준)

### @TEST:REFACTOR-004 테스트 및 검증

#### 5.1 Given-When-Then 시나리오

**시나리오 1: 파일 분리 완료**

```gherkin
Given constants.ts 파일이 454 LOC이고
When 3개 파일로 분리하면
Then 각 파일은 300 LOC 이하여야 한다
And branch-constants.ts는 ~100 LOC여야 한다
And commit-constants.ts는 ~150 LOC여야 한다
And config-constants.ts는 ~200 LOC여야 한다
And index.ts는 ~20 LOC여야 한다
```

**시나리오 2: Import 경로 호환성**

```gherkin
Given 기존 코드가 'import { GitDefaults } from "@/core/git/constants"'를 사용하고
When 파일 분리가 완료되면
Then 기존 import 문이 수정 없이 동작해야 한다
And 타입 추론이 올바르게 작동해야 한다
```

**시나리오 3: 타입 안전성**

```gherkin
Given 모든 상수가 'as const' 어서션을 사용하고
When 파일 분리가 완료되면
Then 각 상수의 타입이 정확히 추론되어야 한다
And 리터럴 타입이 유지되어야 한다
```

**시나리오 4: @TAG 체인 연속성**

```gherkin
Given 기존 파일에 @TAG 주석이 있고
When 파일 분리가 완료되면
Then 각 파일에 적절한 @TAG 주석이 유지되어야 한다
And @TAG 체인이 끊어지지 않아야 한다
```

#### 5.2 품질 게이트

| 기준 | 목표 | 측정 방법 |
|------|------|-----------|
| 파일 크기 | 각 파일 ≤ 300 LOC | `wc -l` 명령어 |
| 타입 검사 | 오류 0개 | `tsc --noEmit` |
| 린트 검사 | 오류 0개 | `npm run lint` |
| 테스트 통과 | 100% | `npm test` |
| 순환 의존성 | 없음 | `madge --circular` |
| Import 경로 | 기존 경로 유지 | 수동 검증 |

#### 5.3 검증 방법

**1. 파일 크기 검증**

```bash
wc -l src/core/git/constants/*.ts
# 예상 결과:
#  ~100 branch-constants.ts
#  ~150 commit-constants.ts
#  ~200 config-constants.ts
#   ~20 index.ts
```

**2. 타입 검사**

```bash
npm run type-check
# 예상: 0 errors
```

**3. 기존 테스트 실행**

```bash
npm test -- git/constants
# 예상: All tests pass
```

**4. Import 경로 확인**

```bash
rg "from '@/core/git/constants'" --files-with-matches
# 모든 파일이 정상 동작해야 함
```

#### 5.4 완료 조건 (Definition of Done)

- [ ] 3개의 상수 파일이 생성되었다
- [ ] index.ts barrel export가 작성되었다
- [ ] 각 파일이 300 LOC 이하다
- [ ] 기존 constants.ts가 삭제되었다
- [ ] 모든 TypeScript 타입 검사가 통과한다
- [ ] 모든 기존 테스트가 통과한다
- [ ] 순환 의존성이 없다
- [ ] 기존 import 경로가 수정 없이 동작한다
- [ ] @TAG 체인이 모든 파일에 유지되었다
- [ ] 코드 리뷰가 완료되었다
- [ ] 문서가 업데이트되었다 (필요 시)

---

## 6. 부록

### 6.1 영향받는 파일 목록

기존 constants.ts를 import하는 파일들:

```bash
# 검색 명령어
rg "from '@/core/git/constants'" -l
rg "from '@core/git/constants'" -l
rg 'from "./constants"' -l src/core/git/
```

### 6.2 참고 자료

- **개발 가이드**: `/.moai/memory/development-guide.md`
- **TRUST 5원칙**: 파일 크기 제약 (≤ 300 LOC)
- **TypeScript Best Practices**: barrel exports, as const

### 6.3 후속 작업 (선택사항)

- 개별 상수 파일에 대한 단위 테스트 추가
- constants 디렉토리 README 작성
- 다른 대형 상수 파일에 동일 패턴 적용

---

**작성일**: 2025-10-01
**상태**: Draft
**우선순위**: High
**카테고리**: Refactoring / Code Quality
**예상 영향 범위**: Git 모듈 전체, import하는 모든 파일

---

_이 SPEC은 MoAI-ADK 개발 가이드 (파일당 300 LOC 이하)를 준수하기 위한 리팩토링입니다._
_브랜치는 `/alfred:2-build SPEC-004` 실행 시 자동 생성됩니다._
