---
id: HOOKS-001
version: 0.0.1
status: draft
created: 2025-10-12
updated: 2025-10-12
author: @Goos
priority: high
category: refactor
labels:
  - hooks
  - javascript
  - cross-platform
  - build-optimization
scope:
  packages:
    - moai-adk-ts/src/claude/hooks
    - moai-adk-ts/templates/.claude/hooks/alfred
  files:
    - policy-block.ts → policy-block.js
    - pre-write-guard.ts → pre-write-guard.js
    - tag-enforcer.ts → tag-enforcer.js
    - session-notice/index.ts → session-notice.js
---

# @SPEC:HOOKS-001: Pure JavaScript 훅 시스템 재설계

## HISTORY

### v0.0.1 (2025-10-12)
- **INITIAL**: Pure JavaScript 훅 시스템 재설계 명세 작성
- **AUTHOR**: @Goos
- **SCOPE**: TypeScript 훅 → Pure JS 변환, 빌드 프로세스 제거
- **CONTEXT**: 크로스 플랫폼 호환성 강화 및 빌드 의존성 제거
- **RELATED**: SPEC-HOOKS-REFACTOR-001 (훅 리팩토링 기반)

---

## Environment (환경)

### 현재 시스템 구성

**TypeScript 훅 시스템** (v0.2.19):
- **개발**: `moai-adk-ts/src/claude/hooks/*.ts` (786 LOC)
- **빌드**: tsup → CommonJS 컴파일
- **배포**: `templates/.claude/hooks/alfred/*.cjs` (42KB)
- **실행**: Node.js 18+ 필수

### 전제 조건

1. **Node.js 런타임 보장**
   - Claude Code는 Node.js 기반 (항상 설치됨)
   - 모든 사용자가 Node.js 18+ 보유

2. **크로스 플랫폼 요구사항**
   - 맥: ✅ 기본 지원
   - 윈도우: ✅ 기본 지원
   - 리눅스: ✅ 기본 지원

3. **기존 훅 로직 호환성**
   - 4개 훅 파일: policy-block, pre-write-guard, tag-enforcer, session-notice
   - 현재 동작 100% 유지

---

## Assumptions (가정)

1. **빌드 없는 배포 가능성**
   - Pure JavaScript는 Node.js에서 직접 실행 가능
   - 소스 코드 = 배포 코드 (디버깅 용이)

2. **성능 유지**
   - Pure JS도 V8 엔진 최적화 수혜
   - TypeScript 컴파일 오버헤드 제거 (더 빠를 가능성)

3. **타입 안전성 유지**
   - TypeScript 소스 병행 유지 (개발용)
   - JSDoc 주석으로 타입 정보 보완

4. **외부 의존성 불필요**
   - Node.js 내장 모듈만 사용 (fs, path, process)
   - JSON 파싱 내장 기능 활용

---

## Requirements (요구사항)

### Ubiquitous Requirements (필수 기능)

- 시스템은 빌드 없이 직접 실행 가능한 Pure JavaScript 훅을 제공해야 한다
- 시스템은 맥/윈도우/리눅스에서 동일하게 작동해야 한다
- 시스템은 Node.js 18+ 환경에서 실행 가능해야 한다
- 시스템은 외부 npm 의존성 없이 작동해야 한다
- 시스템은 기존 TypeScript 훅의 모든 기능을 지원해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN `moai init .` 실행하면, 시스템은 Pure JS 훅을 `templates/` 에서 프로젝트로 복사해야 한다
- WHEN 훅이 실행되면, 시스템은 stdin에서 JSON 입력을 파싱해야 한다
- WHEN 위험 명령이 감지되면, 시스템은 exit code 2로 차단해야 한다
- WHEN 훅 실행이 완료되면, 시스템은 결과를 stdout으로 출력해야 한다
- WHEN 훅 실행 시간이 100ms를 초과하면, 시스템은 경고 메시지를 stderr로 출력해야 한다

### State-driven Requirements (상태 기반)

- WHILE 개발 모드일 때, 시스템은 상세한 디버그 정보를 출력해야 한다
- WHILE 프로덕션 모드일 때, 시스템은 경고/에러만 출력해야 한다
- WHILE 훅이 실행 중일 때, 시스템은 타임아웃(10초)을 강제해야 한다

### Optional Features (선택적 기능)

- WHERE 개발자가 타입 안전성을 원하면, 시스템은 TypeScript 소스를 제공할 수 있다
- WHERE IDE 자동완성이 필요하면, 시스템은 JSDoc 타입 주석을 제공할 수 있다
- WHERE 성능 모니터링이 필요하면, 시스템은 실행 시간을 측정할 수 있다

### Constraints (제약사항)

- IF 외부 npm 패키지를 사용하면, 시스템은 설치 오류를 발생시켜야 한다
- 훅 실행 시간은 100ms를 초과하지 않아야 한다 (slow threshold)
- 훅 파일 크기는 각각 20KB를 초과하지 않아야 한다
- JSON 파싱 실패 시 시스템은 명확한 에러 메시지를 출력해야 한다
- TypeScript 소스와 Pure JS는 동기화되어야 한다 (로직 일치)

---

## Traceability (@TAG)

- **SPEC**: @SPEC:HOOKS-001
- **TEST**: `moai-adk-ts/src/__tests__/claude/hooks/*.test.ts` (기존 유지)
- **CODE**:
  - TypeScript: `moai-adk-ts/src/claude/hooks/*.ts` (개발용)
  - Pure JS: `moai-adk-ts/templates/.claude/hooks/alfred/*.js` (배포용)
- **DOC**: 이 문서 (spec.md)

---

## Technical Details

### 변환 전략

**TypeScript → Pure JavaScript**:
```typescript
// Before (TypeScript)
import type { HookInput, HookResult } from '../types';

export class PolicyBlock implements MoAIHook {
  name = 'policy-block';

  async execute(input: HookInput): Promise<HookResult> {
    // ...
  }
}
```

```javascript
// After (Pure JavaScript + JSDoc)
/**
 * @typedef {Object} HookInput
 * @property {string} tool_name
 * @property {Object} tool_input
 * @property {Object} context
 */

/**
 * @typedef {Object} HookResult
 * @property {boolean} success
 * @property {boolean} [blocked]
 * @property {string} [message]
 * @property {number} [exitCode]
 */

/**
 * Policy Block Hook
 */
class PolicyBlock {
  constructor() {
    this.name = 'policy-block';
  }

  /**
   * @param {HookInput} input
   * @returns {Promise<HookResult>}
   */
  async execute(input) {
    // ...
  }
}

module.exports = { PolicyBlock };
```

### 파일 구조

**개발 환경**:
```
moai-adk-ts/src/claude/hooks/
├── *.ts                    # TypeScript 소스 (개발용)
├── __tests__/*.test.ts     # Vitest 테스트
└── types.ts                # 타입 정의
```

**배포 환경**:
```
templates/.claude/hooks/alfred/
├── policy-block.js         # Pure JS (배포용)
├── pre-write-guard.js
├── tag-enforcer.js
└── session-notice.js
```

### 성능 목표

| 지표 | 현재 (TypeScript) | 목표 (Pure JS) |
|------|-------------------|----------------|
| **실행 시간** | < 100ms | < 100ms (유지) |
| **파일 크기** | 42KB (빌드 후) | < 50KB |
| **빌드 시간** | ~5초 (tsup) | 0초 (빌드 없음) |
| **메모리 사용** | ~10MB | < 15MB |

---

## Migration Strategy

### Phase 1: Pure JS 변환
1. TypeScript 소스 분석
2. 타입 제거 + JSDoc 추가
3. `import` → `require` 변환
4. ES6 클래스 유지

### Phase 2: 빌드 제거
1. `tsup.hooks.config.ts` 삭제
2. `package.json` scripts 업데이트
3. CI/CD 파이프라인 수정

### Phase 3: 테스트 및 검증
1. 맥/윈도우/리눅스 실행 테스트
2. 기존 Vitest 테스트 통과 확인
3. 성능 벤치마크 (< 100ms)

### Phase 4: 문서화
1. README.md 업데이트
2. 개발자 가이드 작성
3. CHANGELOG.md 기록

---

## Dependencies

**없음** - Node.js 내장 모듈만 사용:
- `fs`: 파일 시스템 작업
- `path`: 경로 처리
- `process`: 환경 변수, stdin/stdout

---

## Risks & Mitigation

### Risk 1: TypeScript 소스와 Pure JS 불일치

**위험도**: 중간
**완화 방안**:
- TypeScript 소스를 진실의 원천으로 유지
- Pure JS 변경 시 TypeScript 소스도 수정
- CI/CD에서 동기화 검증 스크립트 실행

### Risk 2: JSDoc 타입 정보 부족

**위험도**: 낮음
**완화 방안**:
- TypeScript 소스 병행 제공
- IDE에서 TypeScript 정의 파일 참조
- 핵심 인터페이스만 JSDoc 명시

### Risk 3: 성능 저하

**위험도**: 매우 낮음
**완화 방안**:
- Pure JS도 V8 최적화 수혜
- 빌드 오버헤드 제거로 오히려 빠를 가능성
- 성능 테스트로 100ms 목표 검증

---

## Success Criteria

1. **크로스 플랫폼 실행**: 맥/윈도우/리눅스 모두 성공
2. **빌드 없는 배포**: `npm publish` 시 빌드 스텝 불필요
3. **성능 유지**: 훅 실행 < 100ms
4. **기존 테스트 통과**: Vitest 테스트 100% 통과
5. **파일 크기**: 각 훅 < 20KB
6. **외부 의존성**: 0개 (Node.js 내장 모듈만)

---

## Related SPECs

- **SPEC-HOOKS-REFACTOR-001**: 훅 리팩토링 (TypeScript 구조 개선)
- **SPEC-INSTALLER-001**: moai init 로직 (템플릿 복사)
