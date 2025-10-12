# SPEC-HOOKS-001 구현 계획서

> Pure JavaScript 훅 시스템 재설계

**SPEC ID**: HOOKS-001
**버전**: 0.0.1
**작성일**: 2025-10-12
**작성자**: @Goos

---

## 📋 실행 요약

**목표**: TypeScript 훅을 Pure JavaScript로 변환하여 빌드 의존성 제거 및 크로스 플랫폼 호환성 강화

**예상 소요 시간**: 5~8시간 (1~2일)

**핵심 변경사항**:
1. 4개 훅 파일 TypeScript → Pure JavaScript 변환
2. tsup 빌드 프로세스 제거
3. `templates/.claude/hooks/alfred/*.js` 직접 작성
4. TypeScript 소스 병행 유지 (개발용)

---

## 🎯 Phase 1: Pure JavaScript 변환 (3~4시간)

### Step 1.1: policy-block.ts → policy-block.js

**현재 구조**:
```typescript
// src/claude/hooks/policy-block.ts (111 LOC)
import type { HookInput, HookResult, MoAIHook } from '../types';
import { DANGEROUS_COMMANDS, ALLOWED_PREFIXES } from './constants';

export class PolicyBlock implements MoAIHook {
  name = 'policy-block';
  async execute(input: HookInput): Promise<HookResult> {
    // ...
  }
}
```

**변환 작업**:
1. ✅ 타입 어노테이션 제거
2. ✅ `import` → `require` 변환
3. ✅ JSDoc 타입 주석 추가
4. ✅ ES6 클래스 유지

**예상 결과**:
```javascript
// templates/.claude/hooks/alfred/policy-block.js
'use strict';

/**
 * @typedef {Object} HookInput
 * @property {string} tool_name
 * @property {Object} tool_input
 * @property {Object} context
 */

const DANGEROUS_COMMANDS = [
  'rm -rf /',
  'rm -rf --no-preserve-root',
  // ...
];

class PolicyBlock {
  constructor() {
    this.name = 'policy-block';
  }

  /**
   * @param {HookInput} input
   * @returns {Promise<HookResult>}
   */
  async execute(input) {
    // ... (로직 동일)
  }
}

module.exports = { PolicyBlock };
```

### Step 1.2: pre-write-guard.ts → pre-write-guard.js

**현재 구조**: 83 LOC, Write/Edit 전 보안 검증

**변환 작업**:
1. ✅ 타입 제거 + JSDoc 추가
2. ✅ 파일 경로 검증 로직 유지
3. ✅ 민감 정보 감지 로직 유지

**예상 시간**: 30분

### Step 1.3: tag-enforcer.ts → tag-enforcer.js

**현재 구조**: 291 LOC, TAG 규칙 강제

**변환 작업**:
1. ✅ 타입 제거 + JSDoc 추가
2. ✅ TAG 정규식 패턴 유지
3. ✅ SPEC ID 검증 로직 유지

**예상 시간**: 1시간 (가장 복잡한 파일)

### Step 1.4: session-notice/index.ts → session-notice.js

**현재 구조**: 81 LOC, 세션 시작 알림

**변환 작업**:
1. ✅ 타입 제거 + JSDoc 추가
2. ✅ Git 상태 확인 로직 유지
3. ✅ SPEC 진행률 표시 로직 유지

**예상 시간**: 30분

### Step 1.5: 공통 유틸리티 변환

**파일**:
- `constants.ts` → 인라인 상수로 변환
- `utils.ts` → 각 파일에 필요한 함수만 복사
- `base.ts` → 공통 실행 로직 인라인화

**변환 전략**:
- 작은 유틸리티는 각 파일에 복사 (중복 허용)
- 큰 공통 로직은 별도 `utils.js` 생성

**예상 시간**: 1시간

---

## 🔧 Phase 2: 빌드 프로세스 제거 (30분)

### Step 2.1: tsup.hooks.config.ts 삭제

```bash
rm moai-adk-ts/tsup.hooks.config.ts
```

### Step 2.2: package.json 업데이트

**Before**:
```json
{
  "scripts": {
    "build": "tsup && bun run build:hooks",
    "build:hooks": "tsup --config tsup.hooks.config.ts"
  }
}
```

**After**:
```json
{
  "scripts": {
    "build": "tsup"
  }
}
```

### Step 2.3: .gitignore 확인

**추가 필요 (없으면)**:
```
templates/.claude/hooks/alfred/*.cjs
```

**유지**:
```
templates/.claude/hooks/alfred/*.js   # Pure JS는 Git 추적
```

---

## ✅ Phase 3: 테스트 및 검증 (2~3시간)

### Step 3.1: 맥에서 실행 테스트

```bash
# 각 훅 단독 실행
node templates/.claude/hooks/alfred/policy-block.js
node templates/.claude/hooks/alfred/pre-write-guard.js
node templates/.claude/hooks/alfred/tag-enforcer.js
node templates/.claude/hooks/alfred/session-notice.js

# 예상 결과: 빈 입력이므로 기본 동작
```

### Step 3.2: stdin 입력 테스트

```bash
echo '{"tool_name":"Bash","tool_input":{"command":"rm -rf /"}}' | \
  node templates/.claude/hooks/alfred/policy-block.js

# 예상 결과: BLOCKED 메시지 + exit code 2
```

### Step 3.3: 윈도우 테스트 (GitHub Actions)

**CI/CD 추가**:
```yaml
# .github/workflows/test-hooks.yml
test-hooks-windows:
  runs-on: windows-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
    - run: node templates/.claude/hooks/alfred/policy-block.js
```

### Step 3.4: 리눅스 테스트 (GitHub Actions)

```yaml
test-hooks-linux:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
    - run: node templates/.claude/hooks/alfred/policy-block.js
```

### Step 3.5: 기존 Vitest 테스트 실행

```bash
cd moai-adk-ts
bun test src/__tests__/claude/hooks/

# 예상: TypeScript 소스 테스트는 유지
# Pure JS는 간접 테스트 (설치 후 실행 검증)
```

### Step 3.6: 성능 벤치마크

```javascript
// benchmark.js
const { PolicyBlock } = require('./templates/.claude/hooks/alfred/policy-block.js');

const start = Date.now();
const hook = new PolicyBlock();
await hook.execute({
  tool_name: 'Bash',
  tool_input: { command: 'echo hello' },
  context: {}
});
const duration = Date.now() - start;

console.log(`Execution time: ${duration}ms`);
// 목표: < 100ms
```

---

## 📚 Phase 4: 문서화 (1시간)

### Step 4.1: README.md 업데이트

**추가 섹션**:
```markdown
## Hooks System

MoAI-ADK uses Pure JavaScript hooks for cross-platform compatibility.

### Development

- **TypeScript source**: `moai-adk-ts/src/claude/hooks/*.ts`
- **Pure JS distribution**: `templates/.claude/hooks/alfred/*.js`

### Synchronization

When modifying TypeScript hooks, manually update Pure JS files:

1. Edit TypeScript source
2. Convert to Pure JavaScript (remove types, add JSDoc)
3. Test on all platforms (mac, windows, linux)
4. Update CHANGELOG.md
```

### Step 4.2: CHANGELOG.md 기록

**v0.2.20 예상**:
```markdown
## [v0.2.20] - 2025-10-12

### Changed

#### 🔄 Pure JavaScript 훅 시스템으로 전환

**변경 사항**:
- TypeScript 훅을 Pure JavaScript로 재작성
- tsup 빌드 프로세스 제거
- 직접 실행 가능한 `.js` 파일 배포

**혜택**:
- ✅ 빌드 시간 0초 (5초 → 0초)
- ✅ 디버깅 용이성 향상 (소스 = 배포)
- ✅ 크로스 플랫폼 호환성 강화

**Technical Details**:
- 변환된 파일: 4개 (policy-block, pre-write-guard, tag-enforcer, session-notice)
- 파일 크기: ~45KB (이전 42KB)
- 성능: < 100ms (유지)
```

### Step 4.3: 개발자 가이드 작성

**파일**: `moai-adk-ts/docs/hooks-development.md`

**내용**:
- TypeScript → Pure JS 변환 가이드
- JSDoc 타입 주석 규칙
- 동기화 체크리스트
- 테스트 절차

---

## 🔒 Phase 5: 품질 검증 (30분)

### Checklist

- [ ] **크로스 플랫폼 실행**: 맥/윈도우/리눅스 모두 성공
- [ ] **성능**: 각 훅 < 100ms
- [ ] **파일 크기**: 각 훅 < 20KB
- [ ] **외부 의존성**: 0개 (Node.js 내장만)
- [ ] **기존 테스트 통과**: Vitest 100%
- [ ] **문서화**: README + CHANGELOG 업데이트
- [ ] **Git 커밋**: 명확한 커밋 메시지

---

## 📊 예상 타임라인

| Phase | 작업 | 예상 시간 | 완료 기준 |
|-------|------|-----------|-----------|
| **1** | Pure JS 변환 | 3~4시간 | 4개 파일 변환 완료 |
| **2** | 빌드 제거 | 30분 | tsup 설정 삭제 |
| **3** | 테스트 | 2~3시간 | 크로스 플랫폼 검증 |
| **4** | 문서화 | 1시간 | README + CHANGELOG |
| **5** | 품질 검증 | 30분 | 체크리스트 완료 |
| **총계** | **5~8시간** | **1~2일** | 배포 준비 완료 |

---

## 🚨 위험 요소 및 대응

### Risk 1: 변환 시 로직 오류

**확률**: 낮음
**영향**: 높음
**대응**:
- TypeScript 테스트 100% 통과 확인
- Pure JS 실행 후 동일 출력 검증
- 코드 리뷰 2회 실시

### Risk 2: 성능 저하

**확률**: 매우 낮음
**영향**: 중간
**대응**:
- 벤치마크 스크립트로 100ms 목표 검증
- 필요 시 캐싱 로직 추가

### Risk 3: TypeScript 소스 불일치

**확률**: 중간
**영향**: 중간
**대응**:
- CI/CD에서 동기화 검증 스크립트 실행
- TypeScript 소스를 진실의 원천으로 유지

---

## 🎯 다음 단계

구현 완료 후:
1. `/alfred:3-sync` 실행 → TAG 체인 검증
2. `package.json` 버전 범프 (0.2.19 → 0.2.20)
3. Git 커밋 + 태그 생성
4. npm publish 준비

---

**작성자**: @Goos
**검토자**: (TBD)
**승인일**: (TBD)
