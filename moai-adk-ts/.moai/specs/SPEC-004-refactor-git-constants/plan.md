# SPEC-004: Git Constants 분리 - 구현 계획

## @TAG BLOCK

```text
# @CODE:SPLIT-CONSTANTS-004 | Chain: @SPEC:CODE-QUALITY-004 -> @SPEC:REFACTOR-004 -> @CODE:SPLIT-CONSTANTS-004
# Related: @CODE:GIT-CONSTANTS-004:DATA
```

---

## 1. 마일스톤 (우선순위 기반)

### 1차 목표: 파일 구조 생성 및 내용 분리

**우선순위**: High

**작업 내용**:
- [ ] `src/core/git/constants/` 디렉토리 생성
- [ ] `branch-constants.ts` 생성 및 GitNamingRules 이동
- [ ] `commit-constants.ts` 생성 및 GitCommitTemplates 이동
- [ ] `config-constants.ts` 생성 및 나머지 상수 이동
- [ ] `index.ts` barrel export 작성

**완료 기준**:
- 4개 파일이 생성되었다
- 각 파일에 적절한 @TAG 주석이 포함되었다
- 모든 상수가 새 위치로 이동되었다

### 2차 목표: 호환성 및 품질 검증

**우선순위**: High

**작업 내용**:
- [ ] TypeScript 타입 검사 실행 (`tsc --noEmit`)
- [ ] 기존 import 경로 정상 작동 확인
- [ ] 타입 추론 정확성 검증
- [ ] 순환 의존성 검사 실행
- [ ] 기존 테스트 전체 실행

**완료 기준**:
- `tsc --noEmit` 오류 0개
- 모든 기존 테스트 통과
- 순환 의존성 없음

### 3차 목표: 원본 파일 정리 및 문서화

**우선순위**: Medium

**작업 내용**:
- [ ] 기존 `constants.ts` 파일 삭제
- [ ] 파일 크기 검증 (각 파일 ≤ 300 LOC)
- [ ] 린트 검사 실행 및 통과
- [ ] `/alfred:3-sync` 실행하여 문서 동기화

**완료 기준**:
- 기존 constants.ts 삭제됨
- 모든 파일이 300 LOC 이하
- 린트 오류 0개
- 문서 동기화 완료

---

## 2. 기술적 접근 방법

### 2.1 파일 분리 전략

**원칙**: 단일 책임 원칙 (Single Responsibility Principle)

```typescript
// branch-constants.ts - 브랜치 명명 규칙만 담당
// @CODE:GIT-NAMING-RULES-001:DATA
export const GitNamingRules = {
  // 브랜치 prefix, 생성/검증 함수
} as const;

// commit-constants.ts - 커밋 메시지 템플릿만 담당
// @CODE:GIT-COMMIT-TEMPLATES-001:DATA
export const GitCommitTemplates = {
  // 커밋 타입, 이모지, 템플릿 함수
} as const;

// config-constants.ts - 설정값과 템플릿 담당
// @CODE:GIT-DEFAULTS-001:DATA, @CODE:GITHUB-DEFAULTS-001:DATA, @CODE:GIT-TIMEOUTS-001:DATA
export const GitDefaults = { /* ... */ } as const;
export const GitHubDefaults = { /* ... */ } as const;
export const GitTimeouts = { /* ... */ } as const;
export const GitignoreTemplates = { /* ... */ } as const;
```

### 2.2 barrel export 패턴

**목적**: 기존 import 경로 완벽 호환

```typescript
// index.ts
// @CODE:REFACTOR-004 | Chain: @SPEC:CODE-QUALITY-004 -> @SPEC:REFACTOR-004

// Re-export all constants to maintain backward compatibility
export * from './branch-constants';
export * from './commit-constants';
export * from './config-constants';

// This allows both:
// import { GitDefaults } from '@/core/git/constants';          // 기존 방식
// import { GitDefaults } from '@/core/git/constants/config-constants'; // 새 방식
```

### 2.3 타입 안전성 보존

**중요**: `as const` 어서션 유지

```typescript
// ✅ 올바른 방식 - 리터럴 타입 유지
export const GitNamingRules = {
  branchPrefixes: {
    feature: 'feature/',
    spec: 'spec/',
    // ...
  },
} as const;

// ❌ 잘못된 방식 - 타입 정보 손실
export const GitNamingRules: any = { /* ... */ };
```

### 2.4 순환 의존성 방지

**규칙**: 상수 파일끼리는 import 금지

```typescript
// ✅ 허용 - 상수만 export
// branch-constants.ts
export const GitNamingRules = { /* ... */ } as const;

// ❌ 금지 - 다른 상수 파일 import
// branch-constants.ts
import { GitDefaults } from './config-constants'; // 순환 의존 위험!
```

### 2.5 import 경로 호환성 전략

**전략**: 기존 코드 수정 없음

```typescript
// 기존 코드 (수정 불필요)
import { GitDefaults, GitNamingRules, GitCommitTemplates } from '@/core/git/constants';

// barrel export (index.ts)가 모든 상수를 re-export하므로
// 위 import 문은 수정 없이 그대로 동작
```

---

## 3. 아키텍처 설계 방향

### 3.1 디렉토리 구조

```
src/core/git/
├── constants/                     # 상수 디렉토리 (신규)
│   ├── index.ts                   # barrel export (~20 LOC)
│   ├── branch-constants.ts        # 브랜치 관련 (~100 LOC)
│   ├── commit-constants.ts        # 커밋 관련 (~150 LOC)
│   └── config-constants.ts        # 설정/템플릿 (~200 LOC)
├── constants.ts                   # 기존 파일 (삭제 예정)
└── ... (기타 Git 관련 파일)
```

### 3.2 의존성 그래프

```
외부 모듈
    ↓
constants/index.ts (barrel export)
    ↓
├── branch-constants.ts   (독립)
├── commit-constants.ts   (독립)
└── config-constants.ts   (독립)

주의: 상수 파일끼리는 의존성 없음 (순환 의존 방지)
```

### 3.3 코드 이동 계획

| 기존 위치 | 새 위치 | 예상 LOC | 설명 |
|----------|---------|----------|------|
| `GitNamingRules` | `branch-constants.ts` | ~100 | 브랜치 prefix, 생성/검증 함수 |
| `GitCommitTemplates` | `commit-constants.ts` | ~150 | 커밋 타입, 이모지, 템플릿 함수 |
| `GitignoreTemplates` | `config-constants.ts` | ~50 | .gitignore 템플릿 |
| `GitDefaults` | `config-constants.ts` | ~50 | Git 기본 설정 |
| `GitHubDefaults` | `config-constants.ts` | ~70 | GitHub API, 라벨, 템플릿 |
| `GitTimeouts` | `config-constants.ts` | ~30 | 작업별 타임아웃 |

---

## 4. 리스크 및 대응 방안

### 4.1 리스크 매트릭스

| 리스크 | 발생 가능성 | 영향도 | 완화 방안 | 우선순위 |
|--------|-------------|--------|-----------|----------|
| **순환 의존성 발생** | 낮음 | 높음 | 상수 파일은 다른 파일을 import하지 않도록 설계 | High |
| **타입 추론 실패** | 낮음 | 중간 | `as const` 유지, `tsc --noEmit`으로 검증 | High |
| **기존 코드 호환성 깨짐** | 중간 | 높음 | barrel export로 기존 경로 유지, 테스트로 검증 | High |
| **파일 크기 불균형** | 중간 | 낮음 | LOC 기준 재검토, 필요 시 추가 분리 | Medium |
| **테스트 실패** | 낮음 | 높음 | 리팩토링 전 모든 테스트 실행, 단계별 검증 | High |
| **import 경로 누락** | 중간 | 중간 | `rg` 명령어로 모든 import 위치 확인 | Medium |

### 4.2 상세 대응 방안

#### 리스크 1: 순환 의존성

**대응**:
- 상수 파일끼리는 절대 import하지 않음
- 필요 시 공통 타입은 별도 `types.ts`에 분리
- `madge --circular` 명령어로 검증

**검증 스크립트**:
```bash
# 순환 의존성 검사
npx madge --circular src/core/git/constants/
# 예상 결과: No circular dependencies found
```

#### 리스크 2: 타입 추론 실패

**대응**:
- 모든 상수에 `as const` 어서션 유지
- TypeScript strict mode 활성화 상태에서 검증
- 타입 테스트 케이스 추가 (필요 시)

**검증 스크립트**:
```bash
# 타입 검사
npm run type-check
# 예상 결과: Found 0 errors
```

#### 리스크 3: 기존 코드 호환성

**대응**:
- barrel export (`index.ts`)로 기존 import 경로 유지
- 모든 named export 유지 (default export 사용 안 함)
- 리팩토링 전후 테스트 비교

**검증 스크립트**:
```bash
# 기존 import 위치 확인
rg "from '@/core/git/constants'" --files-with-matches

# 모든 테스트 실행
npm test
```

#### 리스크 4: 파일 크기 불균형

**대응**:
- 초기 계획: branch(~100), commit(~150), config(~200)
- 만약 config-constants.ts가 300 LOC 초과 시:
  - `gitignore-templates.ts`를 별도 파일로 추가 분리
  - `github-constants.ts`를 별도 파일로 분리

**검증 스크립트**:
```bash
# 파일 크기 확인
wc -l src/core/git/constants/*.ts
```

---

## 5. 코드 품질 체크리스트

### 5.1 구현 전 체크리스트

- [ ] 기존 constants.ts 파일 내용 분석 완료
- [ ] 논리적 그룹핑 계획 수립
- [ ] barrel export 패턴 이해 완료
- [ ] @TAG 체인 연속성 계획 수립

### 5.2 구현 중 체크리스트

- [ ] 각 상수에 `as const` 어서션 유지
- [ ] 모든 export는 named export만 사용
- [ ] 상수 파일끼리 import 없음
- [ ] 각 파일 상단에 @TAG 주석 추가
- [ ] 파일별 LOC 목표 준수 (~100, ~150, ~200)

### 5.3 구현 후 체크리스트

- [ ] `tsc --noEmit` 통과
- [ ] `npm run lint` 통과
- [ ] `npm test` 통과
- [ ] `wc -l` 확인 (각 파일 ≤ 300 LOC)
- [ ] `madge --circular` 통과
- [ ] 기존 import 경로 정상 작동 확인
- [ ] @TAG 체인 무결성 확인

### 5.4 문서화 체크리스트

- [ ] 각 파일 상단에 파일 책임 주석 추가
- [ ] 복잡한 상수에 설명 주석 추가
- [ ] `/alfred:3-sync` 실행하여 문서 동기화
- [ ] SPEC 문서 최신화 (필요 시)

---

## 6. 자동화 스크립트

### 6.1 전체 검증 파이프라인

```bash
#!/bin/bash
# validate-refactor.sh - SPEC-004 검증 스크립트

echo "🔍 SPEC-004 Git Constants 리팩토링 검증 시작..."

# 1. 파일 크기 검증
echo "\n📏 1. 파일 크기 검증 (각 파일 ≤ 300 LOC)"
wc -l src/core/git/constants/*.ts

# 2. 타입 검사
echo "\n🔬 2. TypeScript 타입 검사"
npm run type-check

# 3. 린트 검사
echo "\n✨ 3. 린트 검사"
npm run lint

# 4. 순환 의존성 검사
echo "\n🔄 4. 순환 의존성 검사"
npx madge --circular src/core/git/constants/

# 5. 테스트 실행
echo "\n🧪 5. 테스트 실행"
npm test

# 6. import 경로 확인
echo "\n🔗 6. 기존 import 경로 확인"
rg "from '@/core/git/constants'" --files-with-matches | wc -l

echo "\n✅ 모든 검증 완료!"
```

### 6.2 단계별 검증 명령어

```bash
# Step 1: 파일 크기 확인
wc -l src/core/git/constants/*.ts

# Step 2: 타입 검사
npm run type-check

# Step 3: 린트 검사
npm run lint

# Step 4: 순환 의존성 검사
npx madge --circular src/core/git/constants/

# Step 5: 테스트 실행
npm test

# Step 6: import 경로 확인
rg "from '@/core/git/constants'" --files-with-matches

# Step 7: @TAG 체인 검증
rg "@CODE:REFACTOR-004" src/core/git/constants/
```

---

## 7. 롤백 계획

### 7.1 롤백 시나리오

**트리거**:
- 타입 검사 실패
- 기존 테스트 실패
- 순환 의존성 발견
- 프로덕션 오류 발생

### 7.2 롤백 절차

```bash
# 1. Git 브랜치 확인
git status

# 2. 변경사항 되돌리기
git checkout src/core/git/constants.ts  # 원본 복원
rm -rf src/core/git/constants/          # 신규 디렉토리 삭제

# 3. 테스트 재실행
npm test

# 4. 롤백 원인 분석
# - 로그 확인
# - 오류 메시지 분석
# - 리스크 매트릭스 재검토
```

### 7.3 롤백 후 조치

- 실패 원인 문서화
- SPEC 문서 업데이트 (리스크 추가)
- 대응 방안 재수립 후 재시도

---

## 8. 후속 작업 계획

### 8.1 즉시 후속 작업

**우선순위**: Medium

- [ ] constants 디렉토리 README.md 작성
- [ ] 각 상수 파일에 대한 단위 테스트 추가 (선택적)
- [ ] 다른 대형 상수 파일에 동일 패턴 적용 검토

### 8.2 장기 개선 계획

**우선순위**: Low

- [ ] 상수의 타입 정의를 별도 types.ts로 분리 검토
- [ ] 상수 검증 로직을 별도 validators.ts로 분리 검토
- [ ] constants 디렉토리에 대한 통합 테스트 추가

---

## 9. 참고 자료

### 9.1 관련 문서

- **개발 가이드**: `/.moai/memory/development-guide.md`
  - TRUST 5원칙: 파일당 300 LOC 이하
  - 단일 책임 원칙 (Single Responsibility Principle)

- **SPEC 문서**: `.moai/specs/SPEC-004-refactor-git-constants/spec.md`
  - 상세 요구사항 및 설계

### 9.2 TypeScript Best Practices

- **barrel exports**: 모듈 re-export로 import 경로 단순화
- **as const**: 리터럴 타입 유지로 타입 안전성 극대화
- **named exports**: default export 대신 named export로 트리 셰이킹 향상

### 9.3 도구 문서

- [TypeScript Handbook - Modules](https://www.typescriptlang.org/docs/handbook/modules.html)
- [madge - Circular Dependency Detection](https://github.com/pahen/madge)
- [ripgrep (rg) - Fast Search Tool](https://github.com/BurntSushi/ripgrep)

---

**작성일**: 2025-10-01
**작성자**: @agent-spec-builder
**상태**: Ready for Implementation
**예상 작업 범위**: High
**우선순위**: High

---

_이 계획은 `/alfred:2-build SPEC-004` 실행 시 TDD 구현의 가이드라인이 됩니다._
_시간 예측 없이 우선순위 기반으로 작업을 진행합니다._
