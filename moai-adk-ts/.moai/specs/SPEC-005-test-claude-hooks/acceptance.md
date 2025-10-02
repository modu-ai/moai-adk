# SPEC-005: claude/hooks 모듈 테스트 작성 - 수락 기준

## @DOC:ACCEPT:TEST-HOOKS-001
## CHAIN: REQ:SECURITY-001 -> DESIGN:TEST-001 -> TASK:TEST-HOOKS-001 -> TEST:TEST-HOOKS-001
## STATUS: active
## CREATED: 2025-10-01

---

## Given-When-Then 테스트 시나리오

### 1. 타입 시스템 정의 (@CODE:HOOK-TYPES-001:API)

#### Scenario 1.1: types.ts 파일 생성
```gherkin
Given 타입 파일이 존재하지 않음
When types.ts 파일을 생성함
Then HookInput, HookResult, MoAIHook 인터페이스가 정의되어야 함
And 모든 Hook 파일에서 import 가능해야 함
And TypeScript 컴파일 에러가 없어야 함
```

**검증 방법**:
```bash
# 파일 존재 확인
ls src/claude/hooks/types.ts

# 타입 체크
npm run type-check

# 모든 Hook에서 import 확인
grep -r "from './types'" src/claude/hooks/
```

#### Scenario 1.2: 기존 Hook 파일 타입 import
```gherkin
Given types.ts가 생성됨
When 각 Hook 파일에서 types.ts를 import함
Then 모든 Hook이 MoAIHook 인터페이스를 구현해야 함
And 기존 내부 타입 정의가 제거되어야 함
And 컴파일 에러가 없어야 함
```

**검증 방법**:
```typescript
// 각 Hook 파일 검증
import type { HookInput, HookResult, MoAIHook } from './types';

export class HookClass implements MoAIHook {
  name = 'hook-name';
  async execute(input: HookInput): Promise<HookResult> { /* ... */ }
}
```

---

### 2. PolicyBlock 테스트 (@TEST:POLICY-001)

#### Scenario 2.1: 비-Bash 도구 허용
```gherkin
Given PolicyBlock 인스턴스가 생성됨
When tool_name이 'Read'인 입력을 받음
Then { success: true } 결과를 반환해야 함
And blocked 플래그가 없어야 함
```

**테스트 코드**:
```typescript
test('should allow non-Bash tools', async () => {
  const input: HookInput = {
    tool_name: 'Read',
    tool_input: { file_path: '/some/file.txt' }
  };

  const result = await policyBlock.execute(input);

  expect(result.success).toBe(true);
  expect(result.blocked).toBeUndefined();
});
```

#### Scenario 2.2: 위험한 rm -rf 명령 차단
```gherkin
Given PolicyBlock 인스턴스가 생성됨
When 'rm -rf /' 명령을 받음
Then { success: false, blocked: true } 결과를 반환해야 함
And '위험 명령이 감지되었습니다' 메시지를 포함해야 함
And exitCode가 2여야 함
```

**테스트 코드**:
```typescript
test('should block dangerous rm -rf commands', async () => {
  const input: HookInput = {
    tool_name: 'Bash',
    tool_input: { command: 'rm -rf /' }
  };

  const result = await policyBlock.execute(input);

  expect(result.success).toBe(false);
  expect(result.blocked).toBe(true);
  expect(result.message).toContain('위험 명령이 감지되었습니다');
  expect(result.exitCode).toBe(2);
});
```

#### Scenario 2.3: 화이트리스트 명령 허용
```gherkin
Given PolicyBlock 인스턴스가 생성됨
When 'git status' 같은 허용된 명령을 받음
Then { success: true } 결과를 반환해야 함
And 경고 메시지가 없어야 함
```

**커버리지 목표**: 90% 이상

---

### 3. PreWriteGuard 테스트 (@TEST:PREWRITE-001)

#### Scenario 3.1: 안전한 파일 쓰기 허용
```gherkin
Given PreWriteGuard 인스턴스가 생성됨
When '/project/src/main.py' 파일 쓰기 요청을 받음
Then { success: true } 결과를 반환해야 함
And blocked 플래그가 없어야 함
```

**테스트 코드**:
```typescript
test('should allow safe file writes', async () => {
  const input: HookInput = {
    tool_name: 'Write',
    tool_input: {
      file_path: '/project/src/main.py',
      content: 'print("hello world")'
    }
  };

  const result = await preWriteGuard.execute(input);

  expect(result.success).toBe(true);
  expect(result.blocked).toBeUndefined();
});
```

#### Scenario 3.2: .env 파일 쓰기 차단
```gherkin
Given PreWriteGuard 인스턴스가 생성됨
When '.env' 파일 쓰기 요청을 받음
Then { success: false, blocked: true } 결과를 반환해야 함
And '민감한 파일은 편집할 수 없습니다' 메시지를 포함해야 함
And exitCode가 2여야 함
```

**테스트 코드**:
```typescript
test('should block .env file writes', async () => {
  const input: HookInput = {
    tool_name: 'Write',
    tool_input: { file_path: '.env' }
  };

  const result = await preWriteGuard.execute(input);

  expect(result.success).toBe(false);
  expect(result.blocked).toBe(true);
  expect(result.message).toContain('민감한 파일은 편집할 수 없습니다');
  expect(result.exitCode).toBe(2);
});
```

#### Scenario 3.3: .moai/memory/ 디렉토리 보호
```gherkin
Given PreWriteGuard 인스턴스가 생성됨
When '.moai/memory/development-guide.md' 파일 수정 요청을 받음
Then { success: false, blocked: true } 결과를 반환해야 함
And 보호된 경로임을 알려야 함
```

**커버리지 목표**: 85% 이상

---

### 4. SessionNotifier 테스트 (@TEST:SESSION-001)

#### Scenario 4.1: MoAI 프로젝트 감지
```gherkin
Given .moai 디렉토리와 .claude/commands/alfred가 존재함
When isMoAIProject()를 호출함
Then true를 반환해야 함
```

**테스트 코드**:
```typescript
test('should detect MoAI project', async () => {
  // Mock file system
  vi.mocked(fs.existsSync).mockImplementation((path) => {
    return path.includes('.moai') || path.includes('.claude/commands/alfred');
  });

  const notifier = new SessionNotifier('/mock/project');
  const result = notifier.isMoAIProject();

  expect(result).toBe(true);
});
```

#### Scenario 4.2: 프로젝트 상태 정보 수집
```gherkin
Given MoAI 프로젝트가 초기화됨
When getProjectStatus()를 호출함
Then projectName, moaiVersion, initialized, constitutionStatus, pipelineStage, specProgress 정보를 반환해야 함
And 모든 필드가 유효한 값이어야 함
```

**테스트 코드**:
```typescript
test('should collect project status', async () => {
  // Mock config.json
  vi.mocked(fs.readFileSync).mockReturnValue(JSON.stringify({
    project: { version: '1.0.0' },
    pipeline: { current_stage: 'implementation' }
  }));

  const notifier = new SessionNotifier('/mock/project');
  const status = await notifier.getProjectStatus();

  expect(status).toHaveProperty('projectName');
  expect(status).toHaveProperty('moaiVersion');
  expect(status).toHaveProperty('initialized');
  expect(status).toHaveProperty('constitutionStatus');
  expect(status).toHaveProperty('pipelineStage');
  expect(status).toHaveProperty('specProgress');
});
```

#### Scenario 4.3: Git 정보 타임아웃 처리
```gherkin
Given Git 명령이 2초 이상 걸림
When runGitCommand()를 호출함
Then 2초 후 타임아웃되어야 함
And null을 반환해야 함
And 프로세스가 종료되어야 함
```

**테스트 코드**:
```typescript
test('should handle git command timeout', async () => {
  vi.useFakeTimers();

  const notifier = new SessionNotifier();
  const promise = notifier.runGitCommand(['status']);

  // Fast-forward time
  vi.advanceTimersByTime(2000);

  const result = await promise;
  expect(result).toBeNull();

  vi.useRealTimers();
});
```

**커버리지 목표**: 80% 이상 (외부 의존성 많음)

---

### 5. CodeFirstTAGEnforcer 테스트 (@TEST:ENFORCER-001)

#### Scenario 5.1: @IMMUTABLE TAG 블록 수정 차단
```gherkin
Given 기존 파일에 @IMMUTABLE TAG 블록이 있음
When TAG 블록 내용을 수정하려 함
Then { success: false, blocked: true } 결과를 반환해야 함
And '@IMMUTABLE TAG 수정 금지' 메시지를 포함해야 함
And 해결 방법 제안을 포함해야 함
```

**테스트 코드**:
```typescript
test('should block @IMMUTABLE TAG modification', async () => {
  const oldContent = `/**
   * @DOC:FEATURE:AUTH-001
   * @IMMUTABLE
   */
  class Auth {}`;

  const newContent = `/**
   * @DOC:FEATURE:AUTH-002  // 수정됨
   * @IMMUTABLE
   */
  class Auth {}`;

  vi.mocked(fs.readFile).mockResolvedValue(oldContent);

  const input: HookInput = {
    tool_name: 'Write',
    tool_input: {
      file_path: '/src/auth.ts',
      content: newContent
    }
  };

  const result = await enforcer.execute(input);

  expect(result.success).toBe(false);
  expect(result.blocked).toBe(true);
  expect(result.message).toContain('@IMMUTABLE TAG 수정 금지');
  expect(result.data?.suggestions).toBeDefined();
});
```

#### Scenario 5.2: 새 TAG 블록 유효성 검증
```gherkin
Given 새 파일에 TAG 블록을 추가함
When TAG 형식이 올바르지 않음
Then { success: false, blocked: true } 결과를 반환해야 함
And 구체적인 검증 실패 이유를 알려야 함
And TAG 제안을 제공해야 함
```

**테스트 코드**:
```typescript
test('should validate new TAG block', async () => {
  const invalidTag = `/**
   * @DOC:INVALID_CATEGORY:TEST-001
   */
  class Test {}`;

  const input: HookInput = {
    tool_name: 'Write',
    tool_input: {
      file_path: '/src/test.ts',
      content: invalidTag
    }
  };

  const result = await enforcer.execute(input);

  expect(result.success).toBe(false);
  expect(result.blocked).toBe(true);
  expect(result.message).toContain('유효하지 않은 TAG 카테고리');
});
```

#### Scenario 5.3: TAG 블록 없는 파일 권장 경고
```gherkin
Given 새 파일에 TAG 블록이 없음
When 파일을 생성하려 함
Then { success: true } 결과를 반환해야 함 (차단 안 함)
And 'TAG 블록이 없습니다 (권장사항)' 경고를 출력해야 함
```

**테스트 코드**:
```typescript
test('should warn about missing TAG block', async () => {
  const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});

  const input: HookInput = {
    tool_name: 'Write',
    tool_input: {
      file_path: '/src/new-file.ts',
      content: 'class NewClass {}'
    }
  };

  const result = await enforcer.execute(input);

  expect(result.success).toBe(true);
  expect(consoleError).toHaveBeenCalledWith(
    expect.stringContaining('TAG 블록이 없습니다')
  );

  consoleError.mockRestore();
});
```

**커버리지 목표**: 90% 이상

---

## 품질 게이트 기준

### 1. 테스트 통과율
```gherkin
Given 모든 테스트가 작성됨
When npm run test를 실행함
Then 모든 테스트가 통과해야 함
And 실패한 테스트가 0개여야 함
```

**검증 명령**:
```bash
npm run test
# Expected: All tests passed
```

### 2. 테스트 커버리지
```gherkin
Given 모든 테스트가 통과함
When npm run test:coverage를 실행함
Then 전체 커버리지가 85% 이상이어야 함
And 각 Hook별 목표 커버리지를 달성해야 함
```

**검증 명령**:
```bash
npm run test:coverage

# Expected output:
# --------------------|---------|----------|---------|---------|
# File                | % Stmts | % Branch | % Funcs | % Lines |
# --------------------|---------|----------|---------|---------|
# All files           |   85+   |   80+    |   85+   |   85+   |
#  policy-block.ts    |   90+   |   85+    |   90+   |   90+   |
#  pre-write-guard.ts |   85+   |   80+    |   85+   |   85+   |
#  session-notice.ts  |   80+   |   75+    |   80+   |   80+   |
#  tag-enforcer.ts    |   90+   |   85+    |   90+   |   90+   |
# --------------------|---------|----------|---------|---------|
```

### 3. 타입 안전성
```gherkin
Given 모든 코드가 작성됨
When npm run type-check를 실행함
Then TypeScript 에러가 0개여야 함
And 모든 타입이 명시적으로 정의되어야 함
```

**검증 명령**:
```bash
npm run type-check
# Expected: No type errors
```

### 4. 코드 품질
```gherkin
Given 모든 코드가 작성됨
When npm run check:biome를 실행함
Then 린트 에러가 0개여야 함
And 코드 포맷팅이 일관되어야 함
```

**검증 명령**:
```bash
npm run check:biome
# Expected: No linting errors
```

### 5. 테스트 격리
```gherkin
Given 모든 테스트가 작성됨
When 테스트를 랜덤 순서로 실행함
Then 모든 테스트가 여전히 통과해야 함
And 테스트 간 간섭이 없어야 함
```

**검증 명령**:
```bash
npm run test -- --sequence.shuffle
# Expected: All tests pass in random order
```

### 6. 테스트 실행 시간
```gherkin
Given 모든 테스트가 작성됨
When 테스트 suite를 실행함
Then 전체 실행 시간이 20초 이하여야 함
And 각 파일 실행 시간이 5초 이하여야 함
```

**검증 방법**:
```bash
npm run test -- --reporter=verbose
# Check individual file execution times
```

---

## 검증 방법 및 도구

### 자동화된 검증
```bash
# 1. 모든 테스트 실행
npm run test

# 2. 커버리지 확인
npm run test:coverage

# 3. 타입 체크
npm run type-check

# 4. 린트 검사
npm run check:biome

# 5. 통합 검증
npm run ci
```

### 수동 검증 체크리스트

#### 파일 구조 검증
- [ ] `src/claude/hooks/types.ts` 파일 존재
- [ ] `src/claude/hooks/policy-block.test.ts` 파일 존재
- [ ] `src/claude/hooks/pre-write-guard.test.ts` 파일 존재
- [ ] `src/claude/hooks/session-notice.test.ts` 파일 존재
- [ ] `src/claude/hooks/tag-enforcer.test.ts` 파일 존재

#### 타입 시스템 검증
- [ ] 모든 Hook이 `MoAIHook` 인터페이스 구현
- [ ] `HookInput` 타입 일관되게 사용
- [ ] `HookResult` 타입 일관되게 사용
- [ ] 내부 타입 정의 제거됨

#### 테스트 품질 검증
- [ ] 모든 공개 메서드 테스트 커버
- [ ] 엣지 케이스 포함
- [ ] 에러 시나리오 포함
- [ ] Mock 적절히 활용
- [ ] beforeEach/afterEach로 격리

#### TAG 체인 검증
- [ ] 모든 테스트 파일에 @TAG 주석
- [ ] CHAIN이 올바르게 연결됨
- [ ] DEPENDS가 명시됨
- [ ] STATUS가 active

---

## 완료 조건 (Definition of Done)

### 필수 조건 (Must Have)
- [x] types.ts 파일 생성 완료
- [x] 4개 테스트 파일 생성 완료
- [ ] 모든 테스트 통과 (npm run test)
- [ ] 테스트 커버리지 85% 이상
- [ ] TypeScript 컴파일 에러 0개
- [ ] Biome 린트 에러 0개

### 권장 조건 (Should Have)
- [ ] 테스트 UI로 확인 (npm run test:ui)
- [ ] 커버리지 리포트 생성
- [ ] 각 Hook의 핵심 시나리오 100% 커버
- [ ] Mock 전략 문서화

### 선택 조건 (Could Have)
- [ ] 기존 테스트 마이그레이션
- [ ] 테스트 유틸리티 함수 작성
- [ ] 성능 벤치마크

---

## 실패 시나리오 및 롤백

### Scenario: 테스트 실패
```gherkin
Given 테스트가 실패함
When 원인을 분석함
Then 테스트 로직 수정 또는 소스 코드 수정
And 다시 테스트 실행
And 통과할 때까지 반복
```

### Scenario: 커버리지 미달
```gherkin
Given 커버리지가 85% 미만임
When 커버리지 리포트를 분석함
Then 미커버 영역 식별
And 누락된 테스트 케이스 추가
And 커버리지 재측정
```

### Scenario: 타입 에러
```gherkin
Given TypeScript 컴파일 에러 발생
When 에러 메시지를 확인함
Then 타입 정의 수정
And types.ts 또는 Hook 파일 수정
And 컴파일 에러 해결
```

### Scenario: 롤백 필요
```gherkin
Given 치명적 문제 발생
When 이전 커밋으로 롤백 결정
Then git revert 또는 git reset 수행
And 문제 원인 재분석
And 새로운 접근 방식 시도
```

---

## 최종 검증 스크립트

```bash
#!/bin/bash
# validate-spec-005.sh

echo "=== SPEC-005 검증 시작 ==="

# 1. 파일 존재 확인
echo "1. 파일 구조 검증..."
test -f src/claude/hooks/types.ts || exit 1
test -f src/claude/hooks/policy-block.test.ts || exit 1
test -f src/claude/hooks/pre-write-guard.test.ts || exit 1
test -f src/claude/hooks/session-notice.test.ts || exit 1
test -f src/claude/hooks/tag-enforcer.test.ts || exit 1
echo "✅ 파일 구조 통과"

# 2. 타입 체크
echo "2. 타입 체크..."
npm run type-check || exit 1
echo "✅ 타입 체크 통과"

# 3. 린트 검사
echo "3. 린트 검사..."
npm run check:biome || exit 1
echo "✅ 린트 검사 통과"

# 4. 테스트 실행
echo "4. 테스트 실행..."
npm run test || exit 1
echo "✅ 테스트 통과"

# 5. 커버리지 확인
echo "5. 커버리지 확인..."
npm run test:coverage || exit 1
echo "✅ 커버리지 통과"

echo "=== ✅ SPEC-005 검증 완료 ==="
```

**사용 방법**:
```bash
chmod +x validate-spec-005.sh
./validate-spec-005.sh
```

---

**수락 기준 버전**: 1.0.0
**최종 수정일**: 2025-10-01
