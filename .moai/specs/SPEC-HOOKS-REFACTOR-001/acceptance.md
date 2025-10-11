# @SPEC:HOOKS-REFACTOR-001: Acceptance Criteria

# Acceptance Criteria: Claude Code Hooks 리팩토링

이 문서는 SPEC-HOOKS-REFACTOR-001의 인수 기준을 정의합니다.

**Given-When-Then** 형식의 시나리오로 작성되었으며, 각 시나리오는 독립적으로 검증 가능합니다.

---

## 📋 인수 기준 개요

### 검증 대상
1. CLI Entry Point 통합 (3개 시나리오)
2. 하드코딩 상수 중앙 관리 (3개 시나리오)
3. 다국어 확장자 지원 (2개 시나리오)
4. 공통 유틸리티 함수 (2개 시나리오)
5. 후방 호환성 보장 (2개 시나리오)
6. 코드 품질 (2개 시나리오)

### 성공 기준
- 모든 시나리오 통과
- 기존 76개 + 신규 16개 = 92개 테스트 100% 통과
- Biome 린터 0 warnings
- TypeScript strict mode 컴파일 성공

---

## 🎯 Scenario 1: CLI Entry Point 통합

### AC-001: runHook() 유틸리티 동작

**GIVEN** 4개의 훅 파일(policy-block, pre-write-guard, tag-enforcer, session-notice)이 각각 main() 함수를 가지고 있을 때

**WHEN** runHook(HookClass) 유틸리티 함수를 생성하고 적용하면

**THEN**
- [ ] 4개 파일 모두 `runHook(HookClass)`를 사용해야 한다
- [ ] `main()` 함수가 4개 파일에서 모두 제거되어야 한다
- [ ] 중복 코드가 80 LOC → 0 LOC로 감소해야 한다
- [ ] 모든 기존 테스트(76개)가 통과해야 한다
- [ ] 신규 테스트(base.test.ts 6개)가 통과해야 한다

**검증 방법**:
```bash
# 1. 중복 코드 확인 (0개여야 함)
rg "async function main\(\)" src/claude/hooks/*.ts | wc -l
# 예상 출력: 0

# 2. runHook 사용 확인 (4개여야 함)
rg "runHook\(" src/claude/hooks/*.ts | wc -l
# 예상 출력: 4

# 3. 테스트 실행
npm run test -- src/claude/hooks/__tests__/base.test.ts
# 예상 출력: ✅ Tests  6 passed (6)
```

---

### AC-002: runHook() 오류 처리

**GIVEN** 훅 실행 중 오류가 발생할 때

**WHEN** runHook()이 오류를 catch하면

**THEN**
- [ ] 일관된 형식의 오류 메시지를 출력해야 한다
- [ ] process.exit(1)을 호출해야 한다
- [ ] 오류 스택 트레이스가 표시되어야 한다

**검증 방법**:
```typescript
// base.test.ts
it('should handle hook execution failure gracefully', async () => {
  class FailingHook implements MoAIHook {
    name = 'failing-hook';
    async execute() {
      throw new Error('Test error');
    }
  }

  const exitSpy = vi.spyOn(process, 'exit').mockImplementation(() => {});
  const errorSpy = vi.spyOn(console, 'error');

  await runHook(FailingHook);

  expect(errorSpy).toHaveBeenCalledWith(
    expect.stringContaining('ERROR FailingHook: Test error')
  );
  expect(exitSpy).toHaveBeenCalledWith(1);
});
```

---

### AC-003: require.main === module 패턴 유지

**GIVEN** 훅 파일을 CLI에서 직접 실행할 때 (`node policy-block.cjs`)

**WHEN** `require.main === module` 조건이 만족되면

**THEN**
- [ ] runHook()이 자동으로 호출되어야 한다
- [ ] stdin에서 JSON 입력을 파싱해야 한다
- [ ] 훅을 실행하고 결과를 stdout에 출력해야 한다
- [ ] 오류 없이 완료되어야 한다

**검증 방법**:
```bash
# 1. 컴파일된 cjs 파일 실행
echo '{"tool_name":"Read","tool_input":{}}' | node templates/.claude/hooks/alfred/policy-block.cjs
# 예상 출력: (에러 없이 완료)

# 2. 각 훅 파일 직접 실행
for hook in policy-block pre-write-guard tag-enforcer session-notice; do
  echo '{"tool_name":"Read"}' | node templates/.claude/hooks/alfred/$hook.cjs
done
# 예상 출력: 모두 성공
```

---

## 🗂️ Scenario 2: 하드코딩 상수 중앙 관리

### AC-004: constants.ts 파일 생성

**GIVEN** 파일 확장자, 도구 목록, 타임아웃 값이 여러 파일에 하드코딩되어 있을 때

**WHEN** constants.ts 파일을 생성하고 상수를 이전하면

**THEN**
- [ ] 모든 하드코딩 값이 constants.ts에서 export되어야 한다
- [ ] `SUPPORTED_LANGUAGES`, `READ_ONLY_TOOLS`, `TIMEOUTS`, `DANGEROUS_COMMANDS`, `EXCLUDED_PATHS` 상수가 존재해야 한다
- [ ] 모든 상수에 `as const` assertion이 적용되어야 한다
- [ ] TypeScript가 리터럴 타입을 추론해야 한다

**검증 방법**:
```bash
# 1. constants.ts 파일 존재 확인
ls src/claude/hooks/constants.ts
# 예상 출력: src/claude/hooks/constants.ts

# 2. 상수 export 확인
rg "export const (SUPPORTED_LANGUAGES|READ_ONLY_TOOLS|TIMEOUTS)" src/claude/hooks/constants.ts
# 예상 출력: 3개 매치

# 3. as const 사용 확인
rg "as const" src/claude/hooks/constants.ts | wc -l
# 예상 출력: ≥5

# 4. 테스트 실행
npm run test -- src/claude/hooks/__tests__/constants.test.ts
# 예상 출력: ✅ Tests  4 passed (4)
```

---

### AC-005: constants.ts import 적용

**GIVEN** constants.ts에 상수가 정의되어 있을 때

**WHEN** policy-block, tag-enforcer 등이 constants를 import하면

**THEN**
- [ ] 하드코딩된 배열/객체가 4개 파일에서 제거되어야 한다
- [ ] `import { XXX } from './constants'` 구문이 추가되어야 한다
- [ ] 모든 기존 테스트가 여전히 통과해야 한다

**검증 방법**:
```bash
# 1. constants import 확인
rg "from '\./constants'" src/claude/hooks/*.ts
# 예상 출력: 4개 파일 (policy-block, pre-write-guard, tag-enforcer, session-notice/index)

# 2. 하드코딩 제거 확인 (0개여야 함)
rg "const DANGEROUS_COMMANDS = \[" src/claude/hooks/*.ts | wc -l
# 예상 출력: 0

rg "const READ_ONLY_TOOLS = \[" src/claude/hooks/*.ts | wc -l
# 예상 출력: 0

# 3. 전체 테스트 실행
npm run test
# 예상 출력: ✅ Tests  92 passed (92)
```

---

### AC-006: 타입 안전성 보장

**GIVEN** constants.ts의 상수에 `as const`가 적용되어 있을 때

**WHEN** 잘못된 타입으로 상수를 사용하면

**THEN**
- [ ] TypeScript 컴파일러가 타입 오류를 발생시켜야 한다
- [ ] readonly 타입이 추론되어야 한다
- [ ] 상수 값을 변경하려고 하면 컴파일 오류가 발생해야 한다

**검증 방법**:
```typescript
// 타입 오류 테스트
import { SUPPORTED_LANGUAGES } from './constants';

// ❌ 컴파일 오류 발생해야 함
// SUPPORTED_LANGUAGES.typescript = ['.ts']; // Cannot assign to 'typescript' because it is a read-only property

// ✅ 타입 추론 확인
type TSExtensions = typeof SUPPORTED_LANGUAGES.typescript;
// 타입: readonly [".ts", ".tsx"]
```

```bash
# TypeScript 컴파일 확인
npm run type-check
# 예상 출력: (0 errors)
```

---

## 🌍 Scenario 3: 다국어 확장자 지원

### AC-007: 11개 언어 확장자 포함

**GIVEN** MoAI-ADK가 11개 주요 언어를 지원한다고 명시했을 때

**WHEN** `SUPPORTED_LANGUAGES` 상수를 확인하면

**THEN**
- [ ] 11개 언어가 모두 포함되어야 한다
  - TypeScript, JavaScript, Python, Java, Go, Rust, C++
  - Ruby, PHP, C#, Dart, Swift, Kotlin, Elixir
- [ ] 각 언어의 주요 확장자가 모두 포함되어야 한다
- [ ] markdown (.md, .mdx)도 포함되어야 한다

**검증 방법**:
```bash
# constants.test.ts
npm run test -- src/claude/hooks/__tests__/constants.test.ts -t "should include all 11 languages"
# 예상 출력: ✅ 1 passed
```

```typescript
// 수동 검증
import { SUPPORTED_LANGUAGES } from './constants';

const languages = Object.keys(SUPPORTED_LANGUAGES);
console.log(languages.length); // 14 (11 + cpp + markdown + 2 script languages)

// 새로 추가된 언어 확인
expect(SUPPORTED_LANGUAGES.ruby).toEqual(['.rb', '.rake', '.gemspec']);
expect(SUPPORTED_LANGUAGES.php).toEqual(['.php']);
expect(SUPPORTED_LANGUAGES.csharp).toEqual(['.cs']);
expect(SUPPORTED_LANGUAGES.dart).toEqual(['.dart']);
expect(SUPPORTED_LANGUAGES.swift).toEqual(['.swift']);
expect(SUPPORTED_LANGUAGES.kotlin).toEqual(['.kt', '.kts']);
expect(SUPPORTED_LANGUAGES.elixir).toEqual(['.ex', '.exs']);
```

---

### AC-008: tag-enforcer에서 새 언어 인식

**GIVEN** tag-enforcer가 파일 확장자를 검사할 때

**WHEN** Ruby, PHP, C#, Dart, Swift, Kotlin, Elixir 파일을 처리하면

**THEN**
- [ ] 모든 새 언어 파일이 TAG 검증 대상으로 인식되어야 한다
- [ ] `shouldEnforceTags()` 메서드가 true를 반환해야 한다
- [ ] tag-enforcer 테스트가 모든 언어에 대해 통과해야 한다

**검증 방법**:
```bash
# tag-enforcer 다국어 테스트 실행
npm run test -- src/claude/hooks/__tests__/tag-enforcer.test.ts -t "Multilang Support"
# 예상 출력: ✅ Tests  2 passed (2)
```

```typescript
// tag-enforcer.test.ts
it('should recognize all new language extensions', async () => {
  const newLanguages = [
    { ext: '.rb', lang: 'Ruby' },
    { ext: '.php', lang: 'PHP' },
    { ext: '.cs', lang: 'C#' },
    { ext: '.dart', lang: 'Dart' },
    { ext: '.swift', lang: 'Swift' },
    { ext: '.kt', lang: 'Kotlin' },
    { ext: '.ex', lang: 'Elixir' },
  ];

  for (const { ext, lang } of newLanguages) {
    const enforcer = new CodeFirstTAGEnforcer();
    const filePath = `/path/to/file${ext}`;

    // shouldEnforceTags() private 메서드를 간접적으로 테스트
    const input = {
      tool_name: 'Write',
      tool_input: { file_path: filePath, content: `// ${lang}` },
    };

    const result = await enforcer.execute(input);
    expect(result.success).toBe(true);
  }
});
```

---

## 🛠️ Scenario 4: 공통 유틸리티 함수

### AC-009: extractFilePath() 함수 동작

**GIVEN** tag-enforcer와 pre-write-guard에서 파일 경로를 추출할 때

**WHEN** `extractFilePath(toolInput)` 함수를 사용하면

**THEN**
- [ ] `file_path`, `filePath`, `path`, `notebook_path` 모두 지원해야 한다
- [ ] 우선순위대로 경로를 반환해야 한다 (file_path > filePath > path > notebook_path)
- [ ] 경로가 없으면 null을 반환해야 한다
- [ ] 2개 파일의 중복 로직이 제거되어야 한다

**검증 방법**:
```bash
# utils.test.ts 실행
npm run test -- src/claude/hooks/__tests__/utils.test.ts -t "extractFilePath"
# 예상 출력: ✅ Tests  5 passed (5)
```

```typescript
// utils.test.ts
describe('extractFilePath()', () => {
  it('should extract file_path', () => {
    expect(extractFilePath({ file_path: '/path/file.ts' })).toBe('/path/file.ts');
  });

  it('should extract filePath', () => {
    expect(extractFilePath({ filePath: '/path/file.ts' })).toBe('/path/file.ts');
  });

  it('should prioritize file_path over filePath', () => {
    expect(extractFilePath({
      file_path: '/priority.ts',
      filePath: '/secondary.ts',
    })).toBe('/priority.ts');
  });

  it('should return null if no path found', () => {
    expect(extractFilePath({ other: 'value' })).toBeNull();
  });
});
```

---

### AC-010: extractCommand() 함수 동작

**GIVEN** policy-block에서 Bash 명령어를 추출할 때

**WHEN** `extractCommand(toolInput)` 함수를 사용하면

**THEN**
- [ ] 문자열 명령어를 그대로 반환해야 한다
- [ ] 배열 명령어를 공백으로 join해야 한다
- [ ] `command`와 `cmd` 모두 지원해야 한다
- [ ] 명령어가 없으면 null을 반환해야 한다

**검증 방법**:
```bash
# utils.test.ts 실행
npm run test -- src/claude/hooks/__tests__/utils.test.ts -t "extractCommand"
# 예상 출력: ✅ Tests  4 passed (4)
```

```typescript
// utils.test.ts
describe('extractCommand()', () => {
  it('should extract command string', () => {
    expect(extractCommand({ command: 'git status' })).toBe('git status');
  });

  it('should join command array', () => {
    expect(extractCommand({ command: ['git', 'commit', '-m', 'test'] }))
      .toBe('git commit -m test');
  });

  it('should extract cmd', () => {
    expect(extractCommand({ cmd: 'npm install' })).toBe('npm install');
  });

  it('should return null if no command found', () => {
    expect(extractCommand({ other: 'value' })).toBeNull();
  });
});
```

---

## ⚙️ Scenario 5: 후방 호환성 보장

### AC-011: 기존 훅 동작 불변

**GIVEN** 리팩토링 전 4개 훅의 동작이 정의되어 있을 때

**WHEN** 리팩토링 후 훅을 실행하면

**THEN**
- [ ] 모든 기존 76개 테스트가 100% 통과해야 한다
- [ ] Fast-Track 패턴이 유지되어야 한다 (Read-only tools < 1ms)
- [ ] 훅 실행 결과(HookResult)가 동일해야 한다
- [ ] 성능 저하가 없어야 한다 (< 5% 증가)

**검증 방법**:
```bash
# 1. 전체 기존 테스트 실행
npm run test -- src/claude/hooks/__tests__/policy-block.test.ts
npm run test -- src/claude/hooks/__tests__/pre-write-guard.test.ts
npm run test -- src/claude/hooks/__tests__/tag-enforcer.test.ts
npm run test -- src/claude/hooks/__tests__/session-notice.test.ts
# 예상 출력: ✅ Tests  76 passed (76)

# 2. 성능 벤치마크 테스트
npm run test -- src/claude/hooks/__tests__/policy-block-benchmark.test.ts
# 예상 출력: Fast-Track < 1ms, 전체 < 100ms
```

---

### AC-012: 빌드 및 배포 프로세스 유지

**GIVEN** 기존 빌드 프로세스(tsup)가 정의되어 있을 때

**WHEN** 리팩토링 후 빌드를 실행하면

**THEN**
- [ ] TypeScript → CommonJS 컴파일이 성공해야 한다
- [ ] `templates/.claude/hooks/alfred/*.cjs` 파일이 생성되어야 한다
- [ ] 각 .cjs 파일이 독립 실행 가능해야 한다
- [ ] 빌드 시간이 크게 증가하지 않아야 한다 (< 10% 증가)

**검증 방법**:
```bash
# 1. 빌드 실행
npm run build:hooks
# 예상 출력: ✅ Build successful

# 2. 컴파일된 파일 확인
ls templates/.claude/hooks/alfred/*.cjs
# 예상 출력: 4개 파일 (policy-block.cjs, pre-write-guard.cjs, tag-enforcer.cjs, session-notice.cjs)

# 3. 각 cjs 파일 직접 실행
echo '{"tool_name":"Read"}' | node templates/.claude/hooks/alfred/policy-block.cjs
# 예상 출력: (에러 없이 완료)

# 4. 빌드 시간 측정
time npm run build:hooks
# 예상 출력: < 5초
```

---

## 📏 Scenario 6: 코드 품질

### AC-013: TRUST 5원칙 준수

**GIVEN** MoAI-ADK의 TRUST 5원칙이 정의되어 있을 때

**WHEN** 리팩토링된 코드를 검증하면

**THEN**
- [ ] **Test First**: 테스트 커버리지 ≥85%
- [ ] **Readable**: Biome 린터 0 warnings, 함수 ≤50 LOC, 파일 ≤300 LOC
- [ ] **Unified**: TypeScript strict mode 컴파일 성공
- [ ] **Secured**: 위험 명령 차단 로직 유지
- [ ] **Trackable**: @TAG 시스템 적용 (@CODE:HOOKS-REFACTOR-001)

**검증 방법**:
```bash
# 1. Test First
npm run test:coverage
# 예상 출력: Coverage ≥85%

# 2. Readable
npm run check:biome
# 예상 출력: ✅ No errors found

# 3. Unified
npm run type-check
# 예상 출력: ✅ 0 errors

# 4. Secured
npm run test -- policy-block.test.ts -t "위험 명령 차단"
# 예상 출력: ✅ 4 passed

# 5. Trackable
rg "@CODE:HOOKS-REFACTOR-001" src/claude/hooks/
# 예상 출력: 3개 파일 (base.ts, constants.ts, utils.ts)
```

---

### AC-014: 리팩토링 효과 측정

**GIVEN** 리팩토링 목표가 설정되어 있을 때

**WHEN** 리팩토링 완료 후 코드를 측정하면

**THEN**
- [ ] 코드 중복: 100 LOC → 0 LOC (100% 감소)
- [ ] 언어 지원: 7개 → 11개 (57% 증가)
- [ ] 유지보수성: 새 언어 추가 시 1개 파일만 수정
- [ ] 테스트 증가: 76개 → 92개 (21% 증가)

**검증 방법**:
```bash
# 1. 코드 중복 측정
rg "async function main" src/claude/hooks/*.ts | wc -l
# 예상 출력: 0 (리팩토링 전: 4)

# 2. 언어 지원 측정
rg "typescript:|javascript:|python:|ruby:|php:" src/claude/hooks/constants.ts | wc -l
# 예상 출력: 11

# 3. 테스트 개수 측정
npm run test | grep "Tests"
# 예상 출력: Tests  92 passed (92)

# 4. 파일 개수 확인
ls src/claude/hooks/*.ts | wc -l
# 예상 출력: 7 (기존 4 + 신규 3)
```

---

## ✅ 최종 인수 체크리스트

### 기능 검증
- [ ] AC-001: runHook() 유틸리티 동작
- [ ] AC-002: runHook() 오류 처리
- [ ] AC-003: require.main === module 패턴 유지
- [ ] AC-004: constants.ts 파일 생성
- [ ] AC-005: constants.ts import 적용
- [ ] AC-006: 타입 안전성 보장
- [ ] AC-007: 11개 언어 확장자 포함
- [ ] AC-008: tag-enforcer에서 새 언어 인식
- [ ] AC-009: extractFilePath() 함수 동작
- [ ] AC-010: extractCommand() 함수 동작
- [ ] AC-011: 기존 훅 동작 불변
- [ ] AC-012: 빌드 및 배포 프로세스 유지
- [ ] AC-013: TRUST 5원칙 준수
- [ ] AC-014: 리팩토링 효과 측정

### 코드 품질
- [ ] Biome 린터: 0 warnings
- [ ] TypeScript: 0 errors (strict mode)
- [ ] 테스트 커버리지: ≥85%
- [ ] 파일 크기: 모든 파일 ≤300 LOC
- [ ] 함수 크기: 모든 함수 ≤50 LOC
- [ ] 순환 복잡도: ≤10

### 성능
- [ ] policy-block 실행 시간: < 100ms
- [ ] Fast-Track 패턴: < 1ms
- [ ] 전체 테스트 실행: < 2.5초
- [ ] 빌드 시간: < 5초

### 문서화
- [ ] SPEC 문서 완성 (spec.md, plan.md, acceptance.md)
- [ ] @TAG 시스템 적용
- [ ] JSDoc 주석 추가 (모든 export 함수)
- [ ] HISTORY 섹션 업데이트

---

## 🚀 인수 절차

### 1단계: 자동 검증
```bash
# 전체 자동 검증 실행
npm run ci

# 개별 검증
npm run test              # 92개 테스트
npm run check:biome       # 린터
npm run type-check        # 타입 체크
npm run test:coverage     # 커버리지
npm run build:hooks       # 빌드
```

### 2단계: 수동 검증
- [ ] README 업데이트 확인
- [ ] 컴파일된 .cjs 파일 실행 테스트
- [ ] 새 언어 파일로 실제 테스트 (Ruby, PHP 등)
- [ ] 성능 벤치마크 결과 확인

### 3단계: 문서 동기화
```bash
# /alfred:3-sync 실행
- TAG 체인 무결성 검증
- Living Document 생성
- PR Ready 전환
```

---

## 📝 참고 문서

- **SPEC**: `.moai/specs/SPEC-HOOKS-REFACTOR-001/spec.md`
- **Plan**: `.moai/specs/SPEC-HOOKS-REFACTOR-001/plan.md`
- **Development Guide**: `.moai/memory/development-guide.md`
- **SPEC Metadata**: `.moai/memory/spec-metadata.md`

---

**인수 기준 작성 완료일**: 2025-10-11
**다음 단계**: `/alfred:2-build SPEC-HOOKS-REFACTOR-001`로 TDD 구현 시작
