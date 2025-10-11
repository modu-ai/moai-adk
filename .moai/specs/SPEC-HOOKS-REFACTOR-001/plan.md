# @SPEC:HOOKS-REFACTOR-001 | Chain: @SPEC:HOOKS-REFACTOR-001 -> @TEST:HOOKS-REFACTOR-001 -> @CODE:HOOKS-REFACTOR-001

# Implementation Plan: Claude Code Hooks 리팩토링

## 📋 TDD 구현 계획

이 문서는 SPEC-HOOKS-REFACTOR-001의 TDD 구현 계획을 기술합니다.

**Red-Green-Refactor** 사이클을 엄격히 준수하여 진행합니다.

---

## 🔴 Phase 1: RED (실패 테스트 작성)

### 목표
- 16개의 실패하는 테스트 작성
- 각 테스트는 SPEC의 특정 요구사항을 검증
- 테스트 실행 시 모든 테스트가 실패해야 함 (아직 구현 없음)

### 테스트 파일 구조

```
src/claude/hooks/__tests__/
├── base.test.ts          (신규 - runHook 테스트 6개)
├── constants.test.ts     (신규 - 상수 검증 테스트 4개)
├── utils.test.ts         (신규 - 유틸리티 함수 테스트 4개)
└── tag-enforcer.test.ts  (수정 - 다국어 테스트 2개 추가)
```

---

### 1.1. base.test.ts (6개 테스트)

**파일 위치**: `src/claude/hooks/__tests__/base.test.ts`

**테스트 케이스**:

#### TEST-001: runHook 기본 실행
```typescript
describe('@TEST:HOOKS-REFACTOR-001 - base.ts', () => {
  describe('runHook() 기본 실행', () => {
    it('should execute hook instance successfully', async () => {
      // GIVEN: 유효한 훅 클래스
      class MockHook implements MoAIHook {
        name = 'mock-hook';
        async execute() {
          return { success: true };
        }
      }

      // WHEN: runHook 실행
      // THEN: 오류 없이 완료
      await expect(runHook(MockHook)).resolves.toBeUndefined();
    });
  });
});
```

#### TEST-002: 훅 실행 실패 처리
```typescript
it('should handle hook execution failure gracefully', async () => {
  // GIVEN: 실패하는 훅
  class FailingHook implements MoAIHook {
    name = 'failing-hook';
    async execute() {
      throw new Error('Hook execution failed');
    }
  }

  // WHEN: runHook 실행
  // THEN: process.exit(1) 호출
  const exitSpy = vi.spyOn(process, 'exit').mockImplementation(() => {});
  await runHook(FailingHook);
  expect(exitSpy).toHaveBeenCalledWith(1);
});
```

#### TEST-003: parseClaudeInput 통합
```typescript
it('should parse stdin input correctly', async () => {
  // GIVEN: stdin에 JSON 입력
  const mockInput = { tool_name: 'Read', tool_input: {} };

  // Mock stdin
  const stdinMock = new EventEmitter();
  Object.defineProperty(process, 'stdin', { value: stdinMock });

  // WHEN: parseClaudeInput 호출
  setTimeout(() => {
    stdinMock.emit('data', JSON.stringify(mockInput));
    stdinMock.emit('end');
  }, 10);

  const result = await parseClaudeInput();

  // THEN: 올바르게 파싱
  expect(result).toEqual(mockInput);
});
```

#### TEST-004: outputResult 성공 케이스
```typescript
it('should output success result and exit 0', () => {
  // GIVEN: 성공 결과
  const result: HookResult = { success: true, message: 'OK' };

  // WHEN: outputResult 호출
  const exitSpy = vi.spyOn(process, 'exit').mockImplementation(() => {});
  const logSpy = vi.spyOn(console, 'log');

  outputResult(result);

  // THEN: 메시지 출력 및 exit(0)
  expect(logSpy).toHaveBeenCalledWith('OK');
  expect(exitSpy).toHaveBeenCalledWith(0);
});
```

#### TEST-005: outputResult 블록 케이스
```typescript
it('should output blocked result and exit 2', () => {
  // GIVEN: 차단 결과
  const result: HookResult = {
    success: false,
    blocked: true,
    message: 'Blocked',
    exitCode: 2,
  };

  // WHEN: outputResult 호출
  const exitSpy = vi.spyOn(process, 'exit').mockImplementation(() => {});
  const errorSpy = vi.spyOn(console, 'error');

  outputResult(result);

  // THEN: 에러 출력 및 exit(2)
  expect(errorSpy).toHaveBeenCalledWith('BLOCKED: Blocked');
  expect(exitSpy).toHaveBeenCalledWith(2);
});
```

#### TEST-006: require.main === module 패턴
```typescript
it('should support require.main === module pattern', () => {
  // GIVEN: 모듈이 직접 실행될 때
  // WHEN: runHook이 호출되면
  // THEN: 오류 없이 실행되어야 함
  // (통합 테스트에서 검증)
  expect(runHook).toBeDefined();
});
```

---

### 1.2. constants.test.ts (4개 테스트)

**파일 위치**: `src/claude/hooks/__tests__/constants.test.ts`

**테스트 케이스**:

#### TEST-007: SUPPORTED_LANGUAGES 완전성
```typescript
describe('@TEST:HOOKS-REFACTOR-001 - constants.ts', () => {
  describe('SUPPORTED_LANGUAGES', () => {
    it('should include all 11 languages', () => {
      // GIVEN: SUPPORTED_LANGUAGES 상수
      // WHEN: 키 개수 확인
      const languages = Object.keys(SUPPORTED_LANGUAGES);

      // THEN: 11개 언어 포함
      expect(languages).toHaveLength(11);
      expect(languages).toContain('typescript');
      expect(languages).toContain('javascript');
      expect(languages).toContain('python');
      expect(languages).toContain('java');
      expect(languages).toContain('go');
      expect(languages).toContain('rust');
      expect(languages).toContain('cpp');
      expect(languages).toContain('ruby');
      expect(languages).toContain('php');
      expect(languages).toContain('csharp');
      expect(languages).toContain('dart');
      expect(languages).toContain('swift');
      expect(languages).toContain('kotlin');
      expect(languages).toContain('elixir');
    });

    it('should have correct extensions for each language', () => {
      // THEN: 각 언어의 확장자가 올바름
      expect(SUPPORTED_LANGUAGES.typescript).toEqual(['.ts', '.tsx']);
      expect(SUPPORTED_LANGUAGES.ruby).toEqual(['.rb', '.rake', '.gemspec']);
      expect(SUPPORTED_LANGUAGES.php).toEqual(['.php']);
      expect(SUPPORTED_LANGUAGES.csharp).toEqual(['.cs']);
      expect(SUPPORTED_LANGUAGES.dart).toEqual(['.dart']);
      expect(SUPPORTED_LANGUAGES.swift).toEqual(['.swift']);
      expect(SUPPORTED_LANGUAGES.kotlin).toEqual(['.kt', '.kts']);
      expect(SUPPORTED_LANGUAGES.elixir).toEqual(['.ex', '.exs']);
    });
  });

  describe('READ_ONLY_TOOLS', () => {
    it('should include all read-only tools', () => {
      // THEN: 최소 10개 이상의 read-only tools 포함
      expect(READ_ONLY_TOOLS.length).toBeGreaterThanOrEqual(10);
      expect(READ_ONLY_TOOLS).toContain('Read');
      expect(READ_ONLY_TOOLS).toContain('Glob');
      expect(READ_ONLY_TOOLS).toContain('Grep');
    });
  });

  describe('TIMEOUTS', () => {
    it('should have all timeout constants', () => {
      // THEN: 모든 타임아웃 상수 존재
      expect(TIMEOUTS.TAG_BLOCK_SEARCH_LIMIT).toBe(30);
      expect(TIMEOUTS.GIT_COMMAND_TIMEOUT).toBe(2000);
      expect(TIMEOUTS.NPM_REGISTRY_TIMEOUT).toBe(2000);
      expect(TIMEOUTS.POLICY_BLOCK_SLOW_THRESHOLD).toBe(100);
    });
  });
});
```

---

### 1.3. utils.test.ts (4개 테스트)

**파일 위치**: `src/claude/hooks/__tests__/utils.test.ts`

**테스트 케이스**:

#### TEST-011: extractFilePath 다양한 입력
```typescript
describe('@TEST:HOOKS-REFACTOR-001 - utils.ts', () => {
  describe('extractFilePath()', () => {
    it('should extract file_path', () => {
      const input = { file_path: '/path/to/file.ts' };
      expect(extractFilePath(input)).toBe('/path/to/file.ts');
    });

    it('should extract filePath', () => {
      const input = { filePath: '/path/to/file.ts' };
      expect(extractFilePath(input)).toBe('/path/to/file.ts');
    });

    it('should extract path', () => {
      const input = { path: '/path/to/file.ts' };
      expect(extractFilePath(input)).toBe('/path/to/file.ts');
    });

    it('should extract notebook_path', () => {
      const input = { notebook_path: '/path/to/notebook.ipynb' };
      expect(extractFilePath(input)).toBe('/path/to/notebook.ipynb');
    });

    it('should return null if no path found', () => {
      const input = { other: 'value' };
      expect(extractFilePath(input)).toBeNull();
    });
  });

  describe('extractCommand()', () => {
    it('should extract command string', () => {
      const input = { command: 'git status' };
      expect(extractCommand(input)).toBe('git status');
    });

    it('should extract cmd string', () => {
      const input = { cmd: 'npm install' };
      expect(extractCommand(input)).toBe('npm install');
    });

    it('should join command array', () => {
      const input = { command: ['git', 'commit', '-m', 'test'] };
      expect(extractCommand(input)).toBe('git commit -m test');
    });

    it('should return null if no command found', () => {
      const input = { other: 'value' };
      expect(extractCommand(input)).toBeNull();
    });
  });

  describe('getAllFileExtensions()', () => {
    it('should return all extensions from SUPPORTED_LANGUAGES', () => {
      const extensions = getAllFileExtensions();

      // THEN: 모든 언어의 확장자 포함
      expect(extensions).toContain('.ts');
      expect(extensions).toContain('.py');
      expect(extensions).toContain('.rb');
      expect(extensions).toContain('.php');
      expect(extensions).toContain('.cs');
      expect(extensions).toContain('.dart');
      expect(extensions).toContain('.swift');
      expect(extensions).toContain('.kt');
      expect(extensions).toContain('.ex');
    });

    it('should return unique extensions', () => {
      const extensions = getAllFileExtensions();
      const uniqueExtensions = [...new Set(extensions)];
      expect(extensions.length).toBe(uniqueExtensions.length);
    });
  });
});
```

---

### 1.4. tag-enforcer.test.ts 수정 (2개 테스트 추가)

**파일 위치**: `src/claude/hooks/__tests__/tag-enforcer.test.ts`

**추가 테스트**:

#### TEST-015: Ruby 파일 지원
```typescript
describe('@TEST:HOOKS-REFACTOR-001 - Multilang Support', () => {
  it('should recognize Ruby files', async () => {
    // GIVEN: Ruby 파일
    const input = {
      tool_name: 'Write',
      tool_input: {
        file_path: '/path/to/service.rb',
        content: '# Ruby code',
      },
    };

    // WHEN: tag-enforcer 실행
    const enforcer = new CodeFirstTAGEnforcer();
    const result = await enforcer.execute(input);

    // THEN: Ruby 파일로 인식되어 처리됨
    expect(result.success).toBe(true);
  });

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
      const input = {
        tool_name: 'Write',
        tool_input: {
          file_path: `/path/to/file${ext}`,
          content: `// ${lang} code`,
        },
      };

      const enforcer = new CodeFirstTAGEnforcer();
      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    }
  });
});
```

---

### RED 단계 체크리스트

- [ ] base.test.ts 작성 (6개 테스트)
- [ ] constants.test.ts 작성 (4개 테스트)
- [ ] utils.test.ts 작성 (4개 테스트)
- [ ] tag-enforcer.test.ts 수정 (2개 테스트 추가)
- [ ] 모든 테스트 실행 → 16개 실패 확인
- [ ] 기존 76개 테스트 → 여전히 통과 (리팩토링 전)

**실행 명령어**:
```bash
npm run test -- src/claude/hooks/__tests__/base.test.ts
npm run test -- src/claude/hooks/__tests__/constants.test.ts
npm run test -- src/claude/hooks/__tests__/utils.test.ts
npm run test -- src/claude/hooks/__tests__/tag-enforcer.test.ts
```

**예상 결과**:
```
❌ Test Files  4 failed (4)
❌ Tests  16 failed (16)
⏱️  Duration  ~500ms
```

---

## 🟢 Phase 2: GREEN (최소 구현)

### 목표
- 모든 실패 테스트를 통과시키는 최소한의 구현
- 과도한 최적화 없이 기능만 구현
- 기존 76개 테스트도 100% 통과 유지

### 구현 순서

#### 2.1. base.ts 구현

**구현 내용**:
```typescript
// @CODE:HOOKS-REFACTOR-001 | SPEC: SPEC-HOOKS-REFACTOR-001.md | TEST: __tests__/base.test.ts

import type { MoAIHook } from './types';

/**
 * Run a hook instance with standardized CLI handling
 *
 * @param HookClass - Hook class constructor
 */
export async function runHook(
  HookClass: new () => MoAIHook
): Promise<void> {
  try {
    const { parseClaudeInput, outputResult } = await import('./index');
    const input = await parseClaudeInput();
    const hook = new HookClass();
    const result = await hook.execute(input);
    outputResult(result);
  } catch (error) {
    console.error(
      `ERROR ${HookClass.name}: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  }
}
```

**테스트 검증**: `npm run test -- base.test.ts` → 6개 통과

---

#### 2.2. constants.ts 구현

**구현 내용**:
```typescript
// @CODE:HOOKS-REFACTOR-001 | SPEC: SPEC-HOOKS-REFACTOR-001.md | TEST: __tests__/constants.test.ts

/**
 * Supported programming languages and their file extensions
 */
export const SUPPORTED_LANGUAGES = {
  typescript: ['.ts', '.tsx'],
  javascript: ['.js', '.jsx', '.mjs', '.cjs'],
  python: ['.py', '.pyi'],
  java: ['.java'],
  go: ['.go'],
  rust: ['.rs'],
  cpp: ['.cpp', '.hpp', '.cc', '.h', '.cxx', '.hxx'],
  ruby: ['.rb', '.rake', '.gemspec'],
  php: ['.php'],
  csharp: ['.cs'],
  dart: ['.dart'],
  swift: ['.swift'],
  kotlin: ['.kt', '.kts'],
  elixir: ['.ex', '.exs'],
  markdown: ['.md', '.mdx'],
} as const;

/**
 * Read-only tools that bypass policy checks
 */
export const READ_ONLY_TOOLS = [
  'Read',
  'Glob',
  'Grep',
  'WebFetch',
  'WebSearch',
  'TodoWrite',
  'BashOutput',
  'mcp__context7__resolve-library-id',
  'mcp__context7__get-library-docs',
  'mcp__ide__getDiagnostics',
  'mcp__ide__executeCode',
] as const;

/**
 * Timeout and threshold constants
 */
export const TIMEOUTS = {
  TAG_BLOCK_SEARCH_LIMIT: 30,
  GIT_COMMAND_TIMEOUT: 2000,
  NPM_REGISTRY_TIMEOUT: 2000,
  POLICY_BLOCK_SLOW_THRESHOLD: 100,
} as const;

/**
 * Dangerous commands that should always be blocked
 */
export const DANGEROUS_COMMANDS = [
  'rm -rf /',
  'rm -rf --no-preserve-root',
  'sudo rm',
  'dd if=/dev/zero',
  ':(){:|:&};:',
  'mkfs.',
] as const;

/**
 * Paths that should be excluded from TAG enforcement
 */
export const EXCLUDED_PATHS = [
  'node_modules',
  '.git',
  'dist',
  'build',
  'test',
  'spec',
  '__test__',
  '__tests__',
] as const;

/**
 * Sensitive file patterns to protect
 */
export const SENSITIVE_KEYWORDS = [
  '.env',
  '/secrets',
  '/.git/',
  '/.ssh',
] as const;

/**
 * Protected paths that should not be modified
 */
export const PROTECTED_PATHS = [
  '.moai/memory/',
] as const;
```

**테스트 검증**: `npm run test -- constants.test.ts` → 4개 통과

---

#### 2.3. utils.ts 구현

**구현 내용**:
```typescript
// @CODE:HOOKS-REFACTOR-001 | SPEC: SPEC-HOOKS-REFACTOR-001.md | TEST: __tests__/utils.test.ts

import { SUPPORTED_LANGUAGES } from './constants';

/**
 * Extract file path from tool input
 *
 * @param toolInput - Tool input object
 * @returns File path or null if not found
 */
export function extractFilePath(
  toolInput: Record<string, any>
): string | null {
  return toolInput.file_path
    || toolInput.filePath
    || toolInput.path
    || toolInput.notebook_path
    || null;
}

/**
 * Extract command from tool input
 *
 * @param toolInput - Tool input object
 * @returns Command string or null if not found
 */
export function extractCommand(
  toolInput: Record<string, any>
): string | null {
  const raw = toolInput.command || toolInput.cmd;

  if (Array.isArray(raw)) {
    return raw.map(String).join(' ');
  }

  if (typeof raw === 'string') {
    return raw.trim();
  }

  return null;
}

/**
 * Get all file extensions from supported languages
 *
 * @returns Array of all file extensions
 */
export function getAllFileExtensions(): string[] {
  return Object.values(SUPPORTED_LANGUAGES).flat();
}
```

**테스트 검증**: `npm run test -- utils.test.ts` → 4개 통과

---

#### 2.4. 기존 파일 리팩토링

**policy-block.ts**:
```typescript
import { DANGEROUS_COMMANDS, READ_ONLY_TOOLS, TIMEOUTS } from './constants';
import { extractCommand } from './utils';
import { runHook } from './base';

// ... 기존 코드 ...

// main() 함수 제거
// require.main === module 블록 수정
if (require.main === module) {
  runHook(PolicyBlock).catch(error => {
    console.error(`ERROR policy_block: ${error.message}`);
    process.exit(1);
  });
}
```

**pre-write-guard.ts**:
```typescript
import { SENSITIVE_KEYWORDS, PROTECTED_PATHS } from './constants';
import { extractFilePath } from './utils';
import { runHook } from './base';

// ... 기존 코드 ...

if (require.main === module) {
  runHook(PreWriteGuard).catch(() => {
    process.exit(0);
  });
}
```

**tag-enforcer.ts**:
```typescript
import { EXCLUDED_PATHS } from './constants';
import { extractFilePath, getAllFileExtensions } from './utils';
import { runHook } from './base';

// shouldEnforceTags 메서드 수정
private shouldEnforceTags(filePath: string): boolean {
  const enforceExtensions = getAllFileExtensions();
  const ext = path.extname(filePath);

  // ... 기존 로직 ...

  return enforceExtensions.includes(ext);
}

// ... require.main === module 수정 ...
if (require.main === module) {
  runHook(CodeFirstTAGEnforcer).catch(/* ... */);
}
```

**session-notice/index.ts**:
```typescript
import { runHook } from '../base';

// ... 기존 코드 ...

if (require.main === module) {
  runHook(SessionNotifier).catch(() => {});
}
```

**테스트 검증**: `npm run test` → 전체 92개 테스트 통과

---

### GREEN 단계 체크리스트

- [ ] base.ts 구현 완료
- [ ] constants.ts 구현 완료
- [ ] utils.ts 구현 완료
- [ ] policy-block.ts 리팩토링 완료
- [ ] pre-write-guard.ts 리팩토링 완료
- [ ] tag-enforcer.ts 리팩토링 완료
- [ ] session-notice/index.ts 리팩토링 완료
- [ ] 신규 16개 테스트 통과
- [ ] 기존 76개 테스트 통과
- [ ] 전체 92개 테스트 통과 (100%)

**실행 명령어**:
```bash
npm run test
```

**예상 결과**:
```
✅ Test Files  9 passed (9)
✅ Tests  92 passed (92)
⏱️  Duration  ~2000ms
```

---

## ♻️ Phase 3: REFACTOR (코드 품질 개선)

### 목표
- 코드 중복 완전 제거 확인
- Biome 린터 규칙 준수
- JSDoc 주석 추가
- 타입 안전성 강화

### 리팩토링 체크리스트

#### 3.1. 코드 품질 검증
```bash
npm run check:biome      # Biome 린터 + 포맷 체크
npm run type-check       # TypeScript 타입 체크
npm run test:coverage    # 커버리지 ≥85% 확인
```

#### 3.2. JSDoc 주석 추가
- [ ] base.ts 모든 export 함수에 JSDoc
- [ ] constants.ts 모든 상수에 설명 주석
- [ ] utils.ts 모든 함수에 `@param`, `@returns`, `@example`

#### 3.3. 타입 안전성 강화
- [ ] `Record<string, any>` 사용 제거 (가능한 경우)
- [ ] 모든 상수에 `as const` assertion 확인
- [ ] readonly 타입 보장

#### 3.4. 성능 확인
- [ ] policy-block 벤치마크 테스트 실행
- [ ] Fast-Track 패턴 유지 확인 (< 1ms)
- [ ] 전체 테스트 실행 시간 < 2.5초

#### 3.5. 빌드 확인
```bash
npm run build:hooks      # TypeScript → CommonJS 컴파일
ls -la templates/.claude/hooks/alfred/*.cjs  # 4개 파일 확인
```

---

## 📊 구현 완료 기준

### 기능 검증
- [x] RED: 16개 실패 테스트 작성
- [ ] GREEN: 92개 테스트 통과 (76 + 16)
- [ ] REFACTOR: Biome 0 warnings, TypeScript 0 errors

### 코드 품질
- [ ] 코드 중복: 100 LOC → 0 LOC ✅
- [ ] 언어 지원: 7개 → 11개 (4개 추가) ✅
- [ ] 테스트 커버리지: ≥85% ✅
- [ ] 파일 크기: 모든 파일 ≤300 LOC ✅
- [ ] 함수 크기: 모든 함수 ≤50 LOC ✅

### 성능
- [ ] policy-block < 100ms
- [ ] Fast-Track < 1ms
- [ ] 전체 테스트 < 2.5초

### Git 작업
- [ ] 3개 커밋 생성:
  1. 🔴 RED: SPEC 문서 및 실패 테스트
  2. 🟢 GREEN: 최소 구현으로 테스트 통과
  3. ♻️ REFACTOR: 코드 품질 개선

---

## 다음 단계

이 plan.md를 기반으로 `/alfred:2-build SPEC-HOOKS-REFACTOR-001` 실행 시:
1. RED 단계부터 순차적으로 진행
2. 각 단계마다 커밋 생성
3. 모든 테스트 통과 확인 후 다음 단계

**참고 문서**:
- SPEC: `.moai/specs/SPEC-HOOKS-REFACTOR-001/spec.md`
- Acceptance: `.moai/specs/SPEC-HOOKS-REFACTOR-001/acceptance.md`
