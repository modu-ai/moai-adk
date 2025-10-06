# SPEC-INIT-002 수락 기준

## 개요
Session Notice Hook의 프로젝트 감지 로직이 Alfred 브랜딩에 정렬되고, `.moai` + `.claude/commands/alfred` 동시 검증을 통해 정확하게 MoAI-ADK 프로젝트를 판별합니다.

## Given-When-Then 시나리오

### 시나리오 1: 정상 MoAI 프로젝트 감지
**Given**: 프로젝트 루트에 `.moai` 디렉토리와 `.claude/commands/alfred` 디렉토리가 존재
**When**: `isMoAIProject(projectRoot)` 함수 호출
**Then**: `true` 반환 및 Welcome 메시지 표시

**검증 방법**:
```typescript
// Test Code
const projectRoot = createTempDir({
  '.moai': {},
  '.claude/commands/alfred': {}
});
expect(isMoAIProject(projectRoot)).toBe(true);
```

**실제 동작**:
```bash
# Claude Code 세션 시작 시
$ claude-code .

🏗️ Welcome to Alfred - MoAI SuperAgent
📋 프로젝트 초기화 완료
🚀 /alfred:1-spec으로 SPEC 작성을 시작하세요
```

---

### 시나리오 2: .moai만 존재하는 경우
**Given**: 프로젝트 루트에 `.moai` 디렉토리만 존재, `.claude/commands/alfred` 없음
**When**: `isMoAIProject(projectRoot)` 함수 호출
**Then**: `false` 반환 및 Welcome 메시지 미표시

**검증 방법**:
```typescript
// Test Code
const projectRoot = createTempDir({
  '.moai': {}
  // .claude/commands/alfred 없음
});
expect(isMoAIProject(projectRoot)).toBe(false);
```

**실제 동작**:
```bash
# Claude Code 세션 시작 시
$ claude-code .

# (일반 프로젝트로 취급, MoAI Welcome 메시지 없음)
```

---

### 시나리오 3: .claude/commands/alfred만 존재하는 경우
**Given**: 프로젝트 루트에 `.claude/commands/alfred` 디렉토리만 존재, `.moai` 없음
**When**: `isMoAIProject(projectRoot)` 함수 호출
**Then**: `false` 반환 및 Welcome 메시지 미표시

**검증 방법**:
```typescript
// Test Code
const projectRoot = createTempDir({
  '.claude/commands/alfred': {}
  // .moai 없음
});
expect(isMoAIProject(projectRoot)).toBe(false);
```

**실제 동작**:
```bash
# Claude Code 세션 시작 시
$ claude-code .

# (일반 프로젝트로 취급, MoAI Welcome 메시지 없음)
```

---

### 시나리오 4: 둘 다 없는 경우
**Given**: 프로젝트 루트에 `.moai`와 `.claude/commands/alfred` 모두 없음
**When**: `isMoAIProject(projectRoot)` 함수 호출
**Then**: `false` 반환 및 Welcome 메시지 미표시

**검증 방법**:
```typescript
// Test Code
const projectRoot = createTempDir({});
expect(isMoAIProject(projectRoot)).toBe(false);
```

---

### 시나리오 5: 빌드 프로세스 검증
**Given**: `utils.ts` 수정 완료
**When**: `npm run build:hooks` 실행
**Then**: `templates/.claude/hooks/alfred/session-notice.cjs` 생성되고 `alfred` 경로 포함

**검증 방법**:
```bash
# 빌드 실행
cd moai-adk-ts
npm run build:hooks

# 파일 존재 확인
test -f templates/.claude/hooks/alfred/session-notice.cjs && echo "✅ Build OK"

# 내용 검증
grep -q "alfred" templates/.claude/hooks/alfred/session-notice.cjs && echo "✅ Alfred path OK"
! grep -q "commands/moai" templates/.claude/hooks/alfred/session-notice.cjs && echo "✅ Legacy path removed"
```

---

### 시나리오 6: 파일 시스템 오류 처리
**Given**: 파일 시스템 접근 권한 없음 (예외 상황)
**When**: `isMoAIProject(projectRoot)` 함수 호출
**Then**: 안전하게 `false` 반환 (예외 전파 없음)

**검증 방법**:
```typescript
// Test Code with Mock
vi.mock('fs', () => ({
  existsSync: vi.fn(() => { throw new Error('Permission denied'); })
}));

expect(() => isMoAIProject('/any/path')).not.toThrow();
expect(isMoAIProject('/any/path')).toBe(false);
```

---

## 기능 요구사항 (Functional Requirements)

### FR1: 경로 정확성
- [ ] `.moai` 디렉토리 경로: `${projectRoot}/.moai`
- [ ] `.claude/commands/alfred` 경로: `${projectRoot}/.claude/commands/alfred`
- [ ] 경로 결합에 `path.join()` 사용 (크로스 플랫폼 지원)

### FR2: 동시 검증 로직
- [ ] 두 디렉토리 **모두** 존재해야 `true`
- [ ] 하나라도 없으면 `false`
- [ ] AND 연산자 사용 (`&&`)

### FR3: 빌드 출력
- [ ] 출력 경로: `templates/.claude/hooks/alfred/session-notice.cjs`
- [ ] 포맷: CommonJS (`.cjs`)
- [ ] 번들 크기: ±5% 이내 (경로만 변경)

---

## 비기능 요구사항 (Non-Functional Requirements)

### NFR1: 성능
- [ ] 파일 시스템 체크 시간 < 10ms
- [ ] 빌드 시간 < 5초 (tsup 실행)
- [ ] Session Notice Hook 실행 시간 < 100ms

### NFR2: 보안
- [ ] 경로 탐색 공격 방지 (`path.join()` 사용)
- [ ] 파일 시스템 오류 안전 처리 (예외 전파 없음)
- [ ] 민감 정보 로그 출력 없음

### NFR3: 유지보수성
- [ ] TAG 주석 명확성 (`@CODE:INIT-002`, `@TEST:INIT-002`)
- [ ] 함수명 의도 드러냄 (`isMoAIProject`)
- [ ] 테스트 커버리지 100%

### NFR4: 호환성
- [ ] Node.js >= 18.0.0
- [ ] macOS, Linux, Windows 크로스 플랫폼
- [ ] TypeScript >= 5.0.0

---

## 테스트 요구사항

### 단위 테스트 (Unit Tests)
- [ ] **파일**: `moai-adk-ts/tests/claude/hooks/session-notice/utils.test.ts`
- [ ] **프레임워크**: Vitest
- [ ] **커버리지**: 100% (라인, 브랜치, 함수)
- [ ] **테스트 케이스**: 최소 6개 (시나리오 1-6)

### 통합 테스트 (Integration Tests)
- [ ] 실제 MoAI 프로젝트에서 Session Notice 동작 확인
- [ ] Claude Code 세션 시작 시 Welcome 메시지 표시
- [ ] 비-MoAI 프로젝트에서 메시지 미표시

### 빌드 테스트 (Build Tests)
- [ ] `npm run build:hooks` 성공
- [ ] `.cjs` 파일 생성 확인
- [ ] 번들 내용 검증 (grep 테스트)

---

## 품질 게이트 (Quality Gates)

### TRUST 원칙 준수

#### T - Test First
- [ ] RED 단계: 실패 테스트 작성 완료
- [ ] GREEN 단계: 모든 테스트 통과
- [ ] REFACTOR 단계: 코드 품질 개선 및 빌드 검증

#### R - Readable
- [ ] ESLint 규칙 통과 (no-unused-vars, explicit-function-return-type)
- [ ] 함수명, 변수명 의도 명확
- [ ] 주석 최소화 (코드 자체로 설명)

#### U - Unified
- [ ] TypeScript 타입 안전성 (strict mode)
- [ ] 명시적 반환 타입 (`boolean`)
- [ ] 타입 추론 의존 최소화

#### S - Secured
- [ ] 경로 탐색 공격 방지
- [ ] 파일 시스템 오류 안전 처리
- [ ] 의존성 보안 스캔 (npm audit)

#### T - Trackable
- [ ] TAG 체인 완전성: `@SPEC:INIT-002` → `@TEST:INIT-002` → `@CODE:INIT-002`
- [ ] Git 커밋 메시지에 SPEC ID 포함
- [ ] HISTORY 섹션 업데이트

---

## 완료 조건 (Definition of Done)

### 개발 완료
1. ✅ `utils.ts` 라인 24 수정 (`alfred` 경로)
2. ✅ `utils.test.ts` 테스트 6개 작성 및 통과
3. ✅ 테스트 커버리지 100% 달성
4. ✅ ESLint, TypeScript 컴파일 오류 없음

### 빌드 및 배포
5. ✅ `npm run build:hooks` 성공
6. ✅ `session-notice.cjs` 파일 생성 및 검증
7. ✅ 번들 크기 ±5% 이내

### 문서화
8. ✅ TAG 체인 완전성 확인 (`rg '@(SPEC|TEST|CODE):INIT-002' -n`)
9. ✅ SPEC 문서 HISTORY 업데이트
10. ✅ `acceptance.md` 시나리오 검증 완료

### 통합 검증
11. ✅ 실제 MoAI 프로젝트에서 Session Notice 동작 확인
12. ✅ 비-MoAI 프로젝트에서 메시지 미표시 확인
13. ✅ `/alfred:3-sync` 실행 가능 상태

---

## 검증 도구 및 명령어

### 테스트 실행
```bash
# 단위 테스트
cd moai-adk-ts
npm test -- utils.test.ts

# 커버리지
npm test -- utils.test.ts --coverage

# 전체 테스트
npm test
```

### 빌드 검증
```bash
# 빌드 실행
npm run build:hooks

# 출력 확인
ls -la templates/.claude/hooks/alfred/session-notice.cjs

# 내용 검증
grep "alfred" templates/.claude/hooks/alfred/session-notice.cjs
! grep "commands/moai" templates/.claude/hooks/alfred/session-notice.cjs
```

### TAG 체인 검증
```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE):INIT-002' -n

# 고아 TAG 감지
rg '@CODE:INIT-002' -n moai-adk-ts/src/
rg '@SPEC:INIT-002' -n .moai/specs/

# 중복 확인
rg '@SPEC:INIT-002' -n | wc -l  # 예상: 1
```

### 통합 테스트
```bash
# MoAI 프로젝트에서 실행
cd /path/to/moai-project
claude-code .
# 예상: Alfred Welcome 메시지 표시

# 비-MoAI 프로젝트에서 실행
cd /path/to/normal-project
claude-code .
# 예상: Welcome 메시지 없음
```

---

**승인 기준**: 위 모든 체크리스트 항목이 ✅ 완료되고, Definition of Done 13개 조건 모두 충족 시 SPEC-INIT-002 완료로 간주합니다.
