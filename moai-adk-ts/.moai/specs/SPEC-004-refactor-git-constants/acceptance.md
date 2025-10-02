# SPEC-004: Git Constants 분리 - 수락 기준

## @TAG BLOCK

```text
# @TEST:REFACTOR-004 | Chain: @SPEC:CODE-QUALITY-004 -> @SPEC:REFACTOR-004 -> @CODE:SPLIT-CONSTANTS-004 -> @TEST:REFACTOR-004
# Related: @CODE:GIT-CONSTANTS-004:DATA
```

---

## 1. Given-When-Then 시나리오 (상세)

### 시나리오 1: 파일 분리 완료 검증

```gherkin
Feature: Git Constants 파일 분리
  As a 개발자
  I want constants.ts를 논리적 그룹별로 분리하고 싶다
  So that 파일당 300 LOC 이하를 준수하고 코드 가독성을 향상시킬 수 있다

Scenario: 4개 파일로 분리 완료
  Given constants.ts 파일이 454 LOC이고
    And 개발 가이드는 파일당 300 LOC 이하를 권장하고
  When 논리적 그룹별로 3개 파일 + 1개 barrel export로 분리하면
  Then src/core/git/constants/ 디렉토리가 생성되어야 한다
    And branch-constants.ts 파일이 존재해야 한다
    And commit-constants.ts 파일이 존재해야 한다
    And config-constants.ts 파일이 존재해야 한다
    And index.ts 파일이 존재해야 한다
    And 각 파일은 300 LOC 이하여야 한다

Scenario: 파일별 LOC 목표 달성
  Given 파일 분리가 완료되었고
  When wc -l 명령어로 파일 크기를 측정하면
  Then branch-constants.ts는 90-110 LOC 범위여야 한다
    And commit-constants.ts는 140-160 LOC 범위여야 한다
    And config-constants.ts는 180-220 LOC 범위여야 한다
    And index.ts는 15-25 LOC 범위여야 한다
    And 모든 파일의 합은 425-515 LOC 범위여야 한다 (주석 포함)
```

**검증 명령어**:
```bash
# 파일 존재 확인
ls -la src/core/git/constants/

# 파일 크기 확인
wc -l src/core/git/constants/*.ts

# 예상 출력:
#   ~100 src/core/git/constants/branch-constants.ts
#   ~150 src/core/git/constants/commit-constants.ts
#   ~200 src/core/git/constants/config-constants.ts
#    ~20 src/core/git/constants/index.ts
```

---

### 시나리오 2: Import 경로 호환성 검증

```gherkin
Feature: 기존 Import 경로 호환성 유지
  As a 기존 코드 작성자
  I want 파일 분리 후에도 기존 import 경로가 정상 동작하기를 원한다
  So that 모든 기존 코드를 수정할 필요가 없다

Scenario: 기존 import 경로 정상 작동
  Given 기존 코드가 'import { GitDefaults } from "@/core/git/constants"'를 사용하고
    And 파일 분리가 완료되었고
    And index.ts에 barrel export가 구현되었고
  When TypeScript 컴파일을 실행하면
  Then 타입 오류가 발생하지 않아야 한다
    And GitDefaults의 타입이 정확히 추론되어야 한다
    And 런타임에서 정상적으로 값을 가져올 수 있어야 한다

Scenario: 새로운 import 경로도 지원
  Given 파일 분리가 완료되었고
  When 개발자가 'import { GitDefaults } from "@/core/git/constants/config-constants"'를 사용하면
  Then 타입 오류가 발생하지 않아야 한다
    And 기존 import 방식과 동일한 타입이 추론되어야 한다
    And 런타임 동작이 기존 방식과 동일해야 한다

Scenario: 모든 상수가 barrel export에 포함
  Given index.ts가 모든 하위 파일을 re-export하고
  When 'import * as GIT from "@/core/git/constants"'로 전체 import하면
  Then GIT.GitDefaults가 존재해야 한다
    And GIT.GitNamingRules가 존재해야 한다
    And GIT.GitCommitTemplates가 존재해야 한다
    And GIT.GitHubDefaults가 존재해야 한다
    And GIT.GitTimeouts가 존재해야 한다
    And GIT.GitignoreTemplates가 존재해야 한다
```

**검증 명령어**:
```bash
# 기존 import 위치 확인
rg "from '@/core/git/constants'" --files-with-matches

# 타입 검사
npm run type-check

# 테스트 실행 (런타임 검증)
npm test
```

---

### 시나리오 3: 타입 안전성 보존 검증

```gherkin
Feature: TypeScript 타입 안전성 유지
  As a TypeScript 개발자
  I want 파일 분리 후에도 타입 추론이 정확하기를 원한다
  So that 컴파일 타임에 오류를 잡을 수 있다

Scenario: as const 타입 추론 유지
  Given 모든 상수가 'as const' 어서션을 사용하고
    And 파일 분리가 완료되었고
  When TypeScript 컴파일러가 타입을 추론하면
  Then GitDefaults.mainBranch는 'string' 타입이 아닌 '"main"' 리터럴 타입이어야 한다
    And GitNamingRules.branchPrefixes.feature는 '"feature/"' 리터럴 타입이어야 한다
    And 모든 객체 속성이 readonly로 추론되어야 한다

Scenario: 타입 오류 검증
  Given 파일 분리가 완료되었고
  When tsc --noEmit 명령어를 실행하면
  Then 오류가 0개여야 한다
    And 경고가 0개여야 한다

Scenario: 타입 인텔리센스 작동
  Given 파일 분리가 완료되었고
    And VS Code에서 파일을 열었고
  When 'GitDefaults.' 를 입력하면
  Then 자동 완성 목록이 표시되어야 한다
    And mainBranch, remoteName 등의 속성이 표시되어야 한다
    And 각 속성의 타입이 정확히 표시되어야 한다
```

**검증 명령어**:
```bash
# 타입 검사
npm run type-check

# 타입 정보 확인 (선택적)
npx tsc --noEmit --extendedDiagnostics
```

---

### 시나리오 4: @TAG 체인 연속성 검증

```gherkin
Feature: @TAG 체인 연속성 보존
  As a 프로젝트 관리자
  I want 파일 분리 후에도 @TAG 추적성이 유지되기를 원한다
  So that 요구사항부터 구현까지 추적할 수 있다

Scenario: 각 파일에 @TAG 주석 존재
  Given 파일 분리가 완료되었고
  When 각 파일의 상단을 확인하면
  Then branch-constants.ts에 @CODE:GIT-NAMING-RULES-001:DATA TAG가 있어야 한다
    And commit-constants.ts에 @CODE:GIT-COMMIT-TEMPLATES-001:DATA TAG가 있어야 한다
    And config-constants.ts에 적절한 @CODE TAG들이 있어야 한다
    And index.ts에 @CODE:REFACTOR-004 TAG가 있어야 한다

Scenario: @TAG 체인 완전성
  Given 파일 분리가 완료되었고
  When rg 명령어로 @TAG를 검색하면
  Then @SPEC:CODE-QUALITY-004가 발견되어야 한다
    And @SPEC:REFACTOR-004가 발견되어야 한다
    And @CODE:SPLIT-CONSTANTS-004가 발견되어야 한다
    And @TEST:REFACTOR-004가 발견되어야 한다
    And @CODE:REFACTOR-004가 발견되어야 한다

Scenario: TAG 체인 연결성
  Given 모든 @TAG가 코드에 존재하고
  When Primary Chain을 따라가면
  Then @SPEC -> @SPEC -> @CODE -> @TEST 순서가 유지되어야 한다
    And 끊어진 링크가 없어야 한다
```

**검증 명령어**:
```bash
# @TAG 검색
rg "@CODE:REFACTOR-004" src/core/git/constants/
rg "@CODE:GIT-:DATA" src/core/git/constants/
rg "@SPEC:CODE-QUALITY-004" .moai/specs/SPEC-004*/

# TAG 체인 검증 (선택적)
@agent-debug-helper "TAG 체인 검증을 수행해주세요"
```

---

### 시나리오 5: 순환 의존성 없음 검증

```gherkin
Feature: 순환 의존성 방지
  As a 아키텍트
  I want 분리된 파일들 간에 순환 의존성이 없기를 원한다
  So that 빌드 오류와 예측 불가능한 동작을 방지할 수 있다

Scenario: 상수 파일 간 import 없음
  Given 파일 분리가 완료되었고
  When 각 상수 파일의 import 문을 확인하면
  Then branch-constants.ts는 다른 상수 파일을 import하지 않아야 한다
    And commit-constants.ts는 다른 상수 파일을 import하지 않아야 한다
    And config-constants.ts는 다른 상수 파일을 import하지 않아야 한다
    And index.ts만 하위 파일을 import해야 한다

Scenario: madge 순환 의존성 검사 통과
  Given 파일 분리가 완료되었고
  When madge --circular 명령어를 실행하면
  Then "No circular dependencies found" 메시지가 출력되어야 한다
    And 오류 코드 0을 반환해야 한다
```

**검증 명령어**:
```bash
# 순환 의존성 검사
npx madge --circular src/core/git/constants/

# 예상 출력:
# ✔ No circular dependencies found!
```

---

### 시나리오 6: 기존 테스트 통과 검증

```gherkin
Feature: 기존 테스트 정상 통과
  As a QA 엔지니어
  I want 파일 분리 후에도 모든 기존 테스트가 통과하기를 원한다
  So that 기능 regression이 없음을 보장할 수 있다

Scenario: 전체 테스트 스위트 통과
  Given 파일 분리가 완료되었고
    And 기존 테스트가 constants.ts를 사용하고
  When npm test 명령어를 실행하면
  Then 모든 테스트가 통과해야 한다
    And 실패하는 테스트가 0개여야 한다
    And 스킵된 테스트가 없어야 한다

Scenario: Git 관련 테스트 통과
  Given 파일 분리가 완료되었고
  When npm test -- git 명령어를 실행하면
  Then Git 관련 모든 테스트가 통과해야 한다
    And constants를 import하는 테스트가 정상 작동해야 한다

Scenario: 테스트 커버리지 유지
  Given 파일 분리 전 테스트 커버리지가 X%이고
    And 파일 분리가 완료되었고
  When npm test -- --coverage 명령어를 실행하면
  Then 커버리지가 X% 이상이어야 한다
    And 커버리지가 감소하지 않아야 한다
```

**검증 명령어**:
```bash
# 전체 테스트 실행
npm test

# Git 관련 테스트만 실행
npm test -- git

# 커버리지 포함 테스트
npm test -- --coverage
```

---

## 2. 품질 게이트 기준

### 2.1 필수 품질 게이트 (MUST PASS)

| 게이트 | 기준 | 측정 도구 | 실패 시 조치 | 우선순위 |
|--------|------|-----------|--------------|----------|
| **파일 크기** | 각 파일 ≤ 300 LOC | `wc -l` | 추가 분리 또는 재그룹핑 | High |
| **타입 검사** | 오류 0개 | `tsc --noEmit` | 타입 오류 수정 | High |
| **린트 검사** | 오류 0개 | `npm run lint` | 린트 오류 수정 | High |
| **테스트 통과** | 통과율 100% | `npm test` | 실패 원인 분석 및 수정 | High |
| **순환 의존성** | 없음 | `madge --circular` | 의존성 제거 또는 재설계 | High |
| **Import 경로** | 기존 경로 유지 | 수동 검증 | barrel export 수정 | High |

### 2.2 권장 품질 게이트 (SHOULD PASS)

| 게이트 | 기준 | 측정 도구 | 권장 조치 | 우선순위 |
|--------|------|-----------|-----------|----------|
| **코드 복잡도** | 함수당 ≤ 10 | ESLint | 함수 분리 | Medium |
| **주석 포함** | 주요 상수에 설명 주석 | 코드 리뷰 | 주석 추가 | Medium |
| **@TAG 완전성** | 모든 파일에 TAG | `rg '@TAG'` | TAG 추가 | Medium |
| **커버리지 유지** | 리팩토링 전후 동일 | Jest/Vitest | 테스트 보완 | Medium |

### 2.3 자동화된 품질 게이트 파이프라인

```bash
#!/bin/bash
# quality-gates.sh - 자동화된 품질 게이트 검증

set -e  # 오류 발생 시 즉시 중단

echo "🚦 SPEC-004 품질 게이트 검증 시작..."

# Gate 1: 파일 크기
echo "\n📏 Gate 1: 파일 크기 검증"
for file in src/core/git/constants/*.ts; do
  lines=$(wc -l < "$file")
  if [ "$lines" -gt 300 ]; then
    echo "❌ FAIL: $file has $lines lines (max: 300)"
    exit 1
  fi
  echo "✅ PASS: $file has $lines lines"
done

# Gate 2: 타입 검사
echo "\n🔬 Gate 2: TypeScript 타입 검사"
npm run type-check || { echo "❌ FAIL: Type check failed"; exit 1; }
echo "✅ PASS: Type check succeeded"

# Gate 3: 린트 검사
echo "\n✨ Gate 3: 린트 검사"
npm run lint || { echo "❌ FAIL: Lint check failed"; exit 1; }
echo "✅ PASS: Lint check succeeded"

# Gate 4: 순환 의존성
echo "\n🔄 Gate 4: 순환 의존성 검사"
npx madge --circular src/core/git/constants/ || { echo "❌ FAIL: Circular dependencies found"; exit 1; }
echo "✅ PASS: No circular dependencies"

# Gate 5: 테스트 통과
echo "\n🧪 Gate 5: 테스트 실행"
npm test || { echo "❌ FAIL: Tests failed"; exit 1; }
echo "✅ PASS: All tests passed"

# Gate 6: Import 경로 확인
echo "\n🔗 Gate 6: Import 경로 확인"
import_count=$(rg "from '@/core/git/constants'" --files-with-matches | wc -l)
echo "✅ PASS: Found $import_count files importing from constants"

echo "\n🎉 모든 품질 게이트 통과!"
```

---

## 3. 완료 조건 (Definition of Done)

### 3.1 코드 완료 기준

- [ ] `src/core/git/constants/` 디렉토리가 생성되었다
- [ ] `branch-constants.ts` 파일이 생성되었다 (~100 LOC)
- [ ] `commit-constants.ts` 파일이 생성되었다 (~150 LOC)
- [ ] `config-constants.ts` 파일이 생성되었다 (~200 LOC)
- [ ] `index.ts` barrel export가 생성되었다 (~20 LOC)
- [ ] 기존 `constants.ts` 파일이 삭제되었다
- [ ] 모든 상수가 새 위치로 이동되었다
- [ ] 각 파일에 적절한 @TAG 주석이 추가되었다

### 3.2 품질 검증 완료

- [ ] `tsc --noEmit` 통과 (타입 오류 0개)
- [ ] `npm run lint` 통과 (린트 오류 0개)
- [ ] `npm test` 통과 (테스트 통과율 100%)
- [ ] `madge --circular` 통과 (순환 의존성 없음)
- [ ] 각 파일이 300 LOC 이하
- [ ] 모든 `as const` 어서션 유지됨

### 3.3 호환성 검증 완료

- [ ] 기존 import 경로가 수정 없이 동작한다
- [ ] 타입 추론이 정확히 작동한다
- [ ] 런타임 동작이 리팩토링 전과 동일하다
- [ ] 모든 기존 코드가 정상 작동한다

### 3.4 문서화 완료

- [ ] 각 파일 상단에 파일 책임 주석이 추가되었다
- [ ] 복잡한 상수에 설명 주석이 추가되었다
- [ ] `/alfred:3-sync` 실행으로 문서 동기화 완료
- [ ] SPEC 문서가 최신 상태로 업데이트되었다

### 3.5 코드 리뷰 완료

- [ ] PR이 생성되었다
- [ ] 코드 리뷰가 완료되었다
- [ ] 리뷰 코멘트가 모두 반영되었다
- [ ] 최소 1명 이상의 Approve를 받았다

### 3.6 배포 준비 완료

- [ ] develop 브랜치로 머지 준비 완료
- [ ] CI/CD 파이프라인 통과
- [ ] 프로덕션 배포 체크리스트 확인

---

## 4. 자동 검증 스크립트

### 4.1 완료 조건 자동 체크 스크립트

```bash
#!/bin/bash
# check-dod.sh - Definition of Done 자동 체크

echo "✅ SPEC-004 완료 조건 체크리스트"

# 코드 완료 기준
echo "\n📁 1. 코드 완료 기준"

check_file() {
  if [ -f "$1" ]; then
    echo "  ✅ $1 exists"
    return 0
  else
    echo "  ❌ $1 NOT found"
    return 1
  fi
}

check_file "src/core/git/constants/branch-constants.ts"
check_file "src/core/git/constants/commit-constants.ts"
check_file "src/core/git/constants/config-constants.ts"
check_file "src/core/git/constants/index.ts"

if [ -f "src/core/git/constants.ts" ]; then
  echo "  ❌ Old constants.ts still exists (should be deleted)"
else
  echo "  ✅ Old constants.ts deleted"
fi

# 품질 검증 완료
echo "\n🔍 2. 품질 검증 완료"

echo "  🔬 Running type check..."
npm run type-check > /dev/null 2>&1 && echo "  ✅ Type check passed" || echo "  ❌ Type check failed"

echo "  ✨ Running lint..."
npm run lint > /dev/null 2>&1 && echo "  ✅ Lint passed" || echo "  ❌ Lint failed"

echo "  🧪 Running tests..."
npm test > /dev/null 2>&1 && echo "  ✅ Tests passed" || echo "  ❌ Tests failed"

echo "  🔄 Checking circular dependencies..."
npx madge --circular src/core/git/constants/ > /dev/null 2>&1 && echo "  ✅ No circular deps" || echo "  ❌ Circular deps found"

# 파일 크기 체크
echo "\n📏 3. 파일 크기 체크"
for file in src/core/git/constants/*.ts; do
  lines=$(wc -l < "$file")
  if [ "$lines" -le 300 ]; then
    echo "  ✅ $(basename $file): $lines LOC (≤ 300)"
  else
    echo "  ❌ $(basename $file): $lines LOC (> 300)"
  fi
done

# @TAG 체크
echo "\n🏷️  4. @TAG 완전성 체크"
tag_count=$(rg "@CODE:REFACTOR-004" src/core/git/constants/ | wc -l)
if [ "$tag_count" -gt 0 ]; then
  echo "  ✅ @CODE:REFACTOR-004 found ($tag_count occurrences)"
else
  echo "  ❌ @CODE:REFACTOR-004 NOT found"
fi

echo "\n✅ 체크 완료!"
```

### 4.2 수락 기준 자동 테스트

```typescript
// acceptance.test.ts - SPEC-004 수락 기준 자동 테스트
import { describe, test, expect } from 'vitest';
import { existsSync, statSync } from 'fs';
import * as path from 'path';

describe('@TEST:REFACTOR-004 - SPEC-004 수락 기준', () => {
  const constantsDir = path.join(__dirname, '../src/core/git/constants');

  describe('파일 분리 완료 검증', () => {
    test('constants 디렉토리가 존재해야 한다', () => {
      expect(existsSync(constantsDir)).toBe(true);
    });

    test('branch-constants.ts 파일이 존재해야 한다', () => {
      expect(existsSync(path.join(constantsDir, 'branch-constants.ts'))).toBe(true);
    });

    test('commit-constants.ts 파일이 존재해야 한다', () => {
      expect(existsSync(path.join(constantsDir, 'commit-constants.ts'))).toBe(true);
    });

    test('config-constants.ts 파일이 존재해야 한다', () => {
      expect(existsSync(path.join(constantsDir, 'config-constants.ts'))).toBe(true);
    });

    test('index.ts 파일이 존재해야 한다', () => {
      expect(existsSync(path.join(constantsDir, 'index.ts'))).toBe(true);
    });

    test('각 파일은 300 LOC 이하여야 한다', () => {
      const files = ['branch-constants.ts', 'commit-constants.ts', 'config-constants.ts', 'index.ts'];

      files.forEach((file) => {
        const filePath = path.join(constantsDir, file);
        const content = require('fs').readFileSync(filePath, 'utf8');
        const lineCount = content.split('\n').length;

        expect(lineCount).toBeLessThanOrEqual(300);
      });
    });
  });

  describe('Import 경로 호환성 검증', () => {
    test('기존 import 경로로 GitDefaults를 가져올 수 있어야 한다', async () => {
      const { GitDefaults } = await import('@/core/git/constants');

      expect(GitDefaults).toBeDefined();
      expect(GitDefaults.mainBranch).toBeDefined();
    });

    test('기존 import 경로로 GitNamingRules를 가져올 수 있어야 한다', async () => {
      const { GitNamingRules } = await import('@/core/git/constants');

      expect(GitNamingRules).toBeDefined();
      expect(GitNamingRules.branchPrefixes).toBeDefined();
    });

    test('기존 import 경로로 GitCommitTemplates를 가져올 수 있어야 한다', async () => {
      const { GitCommitTemplates } = await import('@/core/git/constants');

      expect(GitCommitTemplates).toBeDefined();
      expect(GitCommitTemplates.types).toBeDefined();
    });

    test('새로운 import 경로로도 상수를 가져올 수 있어야 한다', async () => {
      const { GitDefaults } = await import('@/core/git/constants/config-constants');

      expect(GitDefaults).toBeDefined();
      expect(GitDefaults.mainBranch).toBeDefined();
    });
  });

  describe('타입 안전성 검증', () => {
    test('GitDefaults의 타입이 정확히 추론되어야 한다', async () => {
      const { GitDefaults } = await import('@/core/git/constants');

      // as const로 인해 리터럴 타입이 유지되어야 함
      const mainBranch: 'main' = GitDefaults.mainBranch;
      expect(mainBranch).toBe('main');
    });

    test('GitNamingRules의 타입이 정확히 추론되어야 한다', async () => {
      const { GitNamingRules } = await import('@/core/git/constants');

      // as const로 인해 리터럴 타입이 유지되어야 함
      const featurePrefix: 'feature/' = GitNamingRules.branchPrefixes.feature;
      expect(featurePrefix).toBe('feature/');
    });
  });

  describe('@TAG 체인 연속성 검증', () => {
    test('각 파일에 적절한 @TAG 주석이 있어야 한다', () => {
      const files = [
        { name: 'branch-constants.ts', tag: '@CODE:GIT-NAMING-RULES-001:DATA' },
        { name: 'commit-constants.ts', tag: '@CODE:GIT-COMMIT-TEMPLATES-001:DATA' },
        { name: 'index.ts', tag: '@CODE:REFACTOR-004' },
      ];

      files.forEach(({ name, tag }) => {
        const filePath = path.join(constantsDir, name);
        const content = require('fs').readFileSync(filePath, 'utf8');

        expect(content).toContain(tag);
      });
    });
  });
});
```

---

## 5. 검증 체크리스트 (수동)

### 5.1 코드 리뷰 체크리스트

**구조 검토**:
- [ ] 파일이 논리적으로 그룹핑되었는가?
- [ ] 각 파일의 책임이 명확한가?
- [ ] 파일 크기가 적절한가?

**타입 안전성**:
- [ ] 모든 상수에 `as const` 어서션이 있는가?
- [ ] 타입 추론이 정확한가?
- [ ] any 타입 사용이 없는가?

**호환성**:
- [ ] 기존 import 경로가 정상 작동하는가?
- [ ] barrel export가 올바르게 구현되었는가?
- [ ] 순환 의존성이 없는가?

**코드 품질**:
- [ ] 주요 상수에 설명 주석이 있는가?
- [ ] @TAG 주석이 모든 파일에 있는가?
- [ ] 네이밍이 일관적인가?

**테스트**:
- [ ] 모든 기존 테스트가 통과하는가?
- [ ] 테스트 커버리지가 유지되는가?
- [ ] Edge case가 고려되었는가?

### 5.2 QA 체크리스트

**기능 검증**:
- [ ] 상수값이 올바르게 이동되었는가?
- [ ] 런타임 동작이 리팩토링 전과 동일한가?
- [ ] 오류 처리가 정상 작동하는가?

**통합 검증**:
- [ ] 다른 모듈과의 통합이 정상인가?
- [ ] import하는 모든 파일이 정상 작동하는가?
- [ ] 빌드가 성공하는가?

**성능 검증**:
- [ ] 빌드 시간이 증가하지 않았는가?
- [ ] 번들 크기가 증가하지 않았는가?
- [ ] 런타임 성능이 유지되는가?

---

## 6. 롤백 기준

### 6.1 롤백 트리거

다음 조건 중 하나라도 해당하면 즉시 롤백:

| 조건 | 심각도 | 조치 |
|------|--------|------|
| 타입 검사 실패 | Critical | 즉시 롤백 |
| 기존 테스트 실패 | Critical | 즉시 롤백 |
| 순환 의존성 발견 | High | 즉시 롤백 |
| 프로덕션 오류 발생 | Critical | 즉시 롤백 |
| 빌드 실패 | High | 즉시 롤백 |
| 성능 저하 (>20%) | Medium | 검토 후 결정 |

### 6.2 롤백 후 조치

1. **원인 분석**: 롤백 원인 상세 문서화
2. **리스크 재평가**: 리스크 매트릭스 업데이트
3. **대응 방안 수립**: 새로운 접근 방식 검토
4. **재시도 계획**: 수정된 계획으로 재시도

---

## 7. 프로덕션 배포 체크리스트

### 7.1 배포 전 최종 확인

- [ ] 모든 DoD 항목 완료
- [ ] 모든 품질 게이트 통과
- [ ] 코드 리뷰 승인 완료
- [ ] QA 검증 완료
- [ ] 문서 업데이트 완료
- [ ] 배포 계획 수립 완료
- [ ] 롤백 계획 수립 완료

### 7.2 배포 후 모니터링

- [ ] 오류 로그 모니터링 (24시간)
- [ ] 성능 메트릭 확인
- [ ] 사용자 피드백 수집
- [ ] 롤백 준비 상태 유지 (48시간)

---

**작성일**: 2025-10-01
**작성자**: @agent-spec-builder
**상태**: Ready for Testing
**우선순위**: High
**예상 테스트 범위**: High

---

_이 수락 기준은 `/alfred:2-build SPEC-004` 실행 후 검증에 사용됩니다._
_모든 품질 게이트를 통과해야 완료로 간주됩니다._
