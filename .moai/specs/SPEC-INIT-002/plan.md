---
id: INIT-002
version: 1.0.0
status: active
created: 2025-10-06
---

# INIT-002 구현 계획

## @SPEC:INIT-002 | Implementation Plan

---

## 목표

Session Notice Hook의 프로젝트 인식 로직을 Alfred 브랜딩에 정렬하여 정확성과 일관성을 향상시킵니다.

**핵심 변경사항**:
- 경로 체크: `.claude/commands/moai` → `.claude/commands/alfred`
- 코드 개선: 배열 기반 체크 → 명시적 변수 사용

---

## 마일스톤

### Phase 1: TypeScript 원본 수정 (우선순위: HIGH)

**대상 파일**: `moai-adk-ts/src/claude/hooks/session-notice/utils.ts`

**작업 내용**:

#### 1.1. `isMoAIProject()` 함수 리팩토링

**변경 전** (Line 21-28):
```typescript
export function isMoAIProject(projectRoot: string): boolean {
  const requiredPaths = [
    path.join(projectRoot, '.moai'),
    path.join(projectRoot, '.claude', 'commands', 'moai'),
  ];

  return requiredPaths.every(p => fs.existsSync(p));
}
```

**변경 후**:
```typescript
/**
 * Check if this is a MoAI project
 * @CODE:INIT-002 | SPEC: .moai/specs/SPEC-INIT-002/spec.md
 */
export function isMoAIProject(projectRoot: string): boolean {
  const moaiDir = path.join(projectRoot, '.moai');
  const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');

  return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
}
```

**변경 이유**:
1. **브랜딩 정렬**: `moai` → `alfred` 경로 변경
2. **가독성**: 명시적 변수명으로 의도 명확화
3. **TAG 추적**: `@CODE:INIT-002` 주석 추가

#### 1.2. 기타 수정 사항

- **파일 헤더 TAG 확인**: `@CODE:SESSION-NOTICE-001:UTILS` 유지
- **타입 안전성**: TypeScript 타입 체크 통과 확인
- **Lint**: ESLint/Biome 검증 통과

---

### Phase 2: 빌드 및 배포 (우선순위: HIGH)

**빌드 명령어**:
```bash
cd moai-adk-ts
npm run build:hooks
```

**빌드 과정**:
1. **입력**: `src/claude/hooks/session-notice/index.ts`
2. **설정**: `tsup.hooks.config.ts`
3. **출력**: `templates/.claude/hooks/alfred/session-notice.cjs`
4. **포맷**: CommonJS (`.cjs`)

**배포 확인**:
```bash
# 빌드 결과물 확인
cat templates/.claude/hooks/alfred/session-notice.cjs | grep -i "alfred"
cat templates/.claude/hooks/alfred/session-notice.cjs | grep -i "moai" | grep -v "\.moai"

# 최종 배포 경로 확인
ls -la .claude/hooks/alfred/session-notice.cjs
```

**예상 결과**:
- `alfred` 경로가 코드에 포함됨
- 레거시 `moai` 경로는 `.moai` 디렉토리 참조만 남음

---

### Phase 3: 검증 및 테스트 (우선순위: MEDIUM)

#### 3.1. 수동 검증 시나리오

**Scenario 1: 정상 프로젝트 인식**
```bash
# 전제조건
ls .moai                              # 존재
ls .claude/commands/alfred            # 존재

# 실행
claude-code .                         # 새 세션 시작

# 기대 결과
✅ MoAI 프로젝트로 인식
✅ SPEC 진행 상황 표시
✅ 초기화 안내 메시지 미표시
```

**Scenario 2: 초기화 필요 프로젝트**
```bash
# 전제조건
ls .moai                              # 없음 또는
ls .claude/commands/alfred            # 없음

# 실행
claude-code .                         # 새 세션 시작

# 기대 결과
✅ MoAI 프로젝트로 미인식
✅ `/alfred:0-project` 안내 메시지 표시
✅ SPEC 상태 미표시
```

**Scenario 3: 레거시 경로 확인**
```bash
# 전제조건
ls .moai                              # 존재
ls .claude/commands/moai              # 존재 (레거시)
ls .claude/commands/alfred            # 없음

# 실행
claude-code .                         # 새 세션 시작

# 기대 결과
❌ MoAI 프로젝트로 미인식 (alfred 경로 없음)
✅ 초기화 안내 메시지 표시
```

#### 3.2. 자동 검증 (선택적)

**단위 테스트 추가** (선택 사항):
```typescript
// tests/hooks/session-notice/utils.test.ts
import { describe, it, expect } from 'vitest';
import { isMoAIProject } from '../../../src/claude/hooks/session-notice/utils';

describe('@TEST:INIT-002 | isMoAIProject', () => {
  it('should return true when both .moai and .claude/commands/alfred exist', () => {
    // Given: MoAI 프로젝트 경로
    const projectRoot = '/path/to/moai-project';

    // When: isMoAIProject 호출
    const result = isMoAIProject(projectRoot);

    // Then: true 반환
    expect(result).toBe(true);
  });

  it('should return false when .claude/commands/alfred is missing', () => {
    // Given: alfred 명령어 없음
    const projectRoot = '/path/to/non-moai-project';

    // When: isMoAIProject 호출
    const result = isMoAIProject(projectRoot);

    // Then: false 반환
    expect(result).toBe(false);
  });
});
```

**실행**:
```bash
npm run test -- utils.test.ts
```

---

## 기술적 접근 방법

### 1. 경로 체크 전략

**현재 방식** (배열 기반):
```typescript
const requiredPaths = [path1, path2];
return requiredPaths.every(p => fs.existsSync(p));
```

**개선 방식** (명시적):
```typescript
const moaiDir = path.join(projectRoot, '.moai');
const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');
return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
```

**장점**:
- 변수명으로 의도 명확화
- 디버깅 용이 (어떤 경로가 없는지 쉽게 파악)
- 성능 동일 (O(1) 유지)

### 2. 빌드 시스템 활용

**tsup 설정** (`tsup.hooks.config.ts`):
```typescript
export default defineConfig({
  entry: ['src/claude/hooks/session-notice/index.ts'],
  outDir: 'templates/.claude/hooks/alfred',
  format: ['cjs'],
  clean: false,
  dts: false,
});
```

**특징**:
- **Tree Shaking**: 미사용 코드 제거
- **Minification**: 선택적 (개발 중 비활성화 가능)
- **Source Map**: 디버깅용 생성

### 3. 레거시 호환성 고려

**옵션 1: Hard Cut (권장)**
- `alfred` 경로만 체크
- 레거시 프로젝트는 `/alfred:0-project` 재실행 필요

**옵션 2: Soft Migration**
```typescript
export function isMoAIProject(projectRoot: string): boolean {
  const moaiDir = path.join(projectRoot, '.moai');
  const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');
  const legacyMoaiCommands = path.join(projectRoot, '.claude', 'commands', 'moai');

  return fs.existsSync(moaiDir) &&
         (fs.existsSync(alfredCommands) || fs.existsSync(legacyMoaiCommands));
}
```

**현재 선택**: **옵션 1 (Hard Cut)** - 깔끔한 전환, 레거시 부채 제거

---

## 리스크 및 대응 방안

### 리스크 1: 기존 사용자 프로젝트 인식 실패

**가능성**: MEDIUM
**영향도**: HIGH

**대응**:
1. **문서화**: CHANGELOG에 Breaking Change 명시
2. **마이그레이션 가이드**: `/alfred:0-project` 재실행 안내
3. **경고 메시지**: Session Notice에서 명확한 안내

### 리스크 2: 빌드 실패

**가능성**: LOW
**영향도**: HIGH

**대응**:
1. **빌드 전 테스트**: `npm run build:hooks` 수동 실행
2. **CI/CD 검증**: GitHub Actions에서 빌드 체크
3. **롤백 계획**: Git 이력으로 즉시 복원 가능

### 리스크 3: 타입 에러

**가능성**: LOW
**영향도**: MEDIUM

**대응**:
1. **TypeScript 검증**: `tsc --noEmit` 실행
2. **ESLint/Biome**: 정적 분석 통과 확인

---

## 다음 단계

### 즉시 실행
1. ✅ **spec.md** 작성 완료
2. ✅ **plan.md** 작성 완료
3. ⏳ **acceptance.md** 작성 대기

### 구현 단계 (`/alfred:2-build INIT-002`)
1. `utils.ts` 수정 및 TAG 추가
2. `npm run build:hooks` 실행
3. 배포 파일 검증

### 동기화 단계 (`/alfred:3-sync`)
1. TAG 체인 검증
2. Living Document 생성
3. PR 상태 전환 (Draft → Ready)

---

_INIT-002 Implementation Plan | Alfred 브랜딩 정렬 로드맵_
