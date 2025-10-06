# SPEC-INIT-002 구현 계획

## 목표
Session Notice Hook의 프로젝트 감지 로직을 Alfred 브랜딩에 정렬하고, 정확한 MoAI-ADK 프로젝트 판별 기능을 구현합니다.

## 우선순위별 마일스톤

### Phase 1: 테스트 케이스 작성 (RED)
**의존성**: 없음
**목표**: 요구사항 기반 실패 테스트 작성

#### 작업 항목
1. **테스트 파일 생성**
   - 경로: `moai-adk-ts/tests/claude/hooks/session-notice/utils.test.ts`
   - TAG: `@TEST:INIT-002`
   - 프레임워크: Vitest (기존 프로젝트 표준)

2. **테스트 시나리오**
   ```typescript
   describe('isMoAIProject', () => {
     it('.moai + .claude/commands/alfred 존재 시 true 반환', () => {});
     it('.moai만 존재 시 false 반환', () => {});
     it('.claude/commands/alfred만 존재 시 false 반환', () => {});
     it('둘 다 없으면 false 반환', () => {});
   });
   ```

3. **테스트 실행**
   ```bash
   cd moai-adk-ts
   npm test -- utils.test.ts
   # 예상 결과: 모든 테스트 실패 (RED)
   ```

### Phase 2: 구현 및 테스트 통과 (GREEN)
**의존성**: Phase 1 완료
**목표**: 최소한의 코드로 테스트 통과

#### 작업 항목
1. **TypeScript 원본 수정**
   - 파일: `moai-adk-ts/src/claude/hooks/session-notice/utils.ts`
   - TAG: `@CODE:INIT-002`
   - 변경: 라인 24 `moai` → `alfred`

2. **변경 내용**
   ```typescript
   // @CODE:INIT-002 | SPEC: SPEC-INIT-002.md | TEST: tests/claude/hooks/session-notice/utils.test.ts
   export function isMoAIProject(projectRoot: string): boolean {
     const moaiDir = path.join(projectRoot, '.moai');
     const alfredCommands = path.join(projectRoot, '.claude', 'commands', 'alfred');

     return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
   }
   ```

3. **테스트 재실행**
   ```bash
   npm test -- utils.test.ts
   # 예상 결과: 모든 테스트 통과 (GREEN)
   ```

### Phase 3: 빌드 및 배포 (REFACTOR)
**의존성**: Phase 2 완료
**목표**: 프로덕션 번들 생성 및 검증

#### 작업 항목
1. **빌드 실행**
   ```bash
   cd moai-adk-ts
   npm run build:hooks
   ```

2. **출력 검증**
   ```bash
   # 파일 생성 확인
   test -f templates/.claude/hooks/alfred/session-notice.cjs && echo "✅ Build OK"

   # 내용 검증
   grep "alfred" templates/.claude/hooks/alfred/session-notice.cjs
   ! grep "commands/moai" templates/.claude/hooks/alfred/session-notice.cjs
   ```

3. **통합 테스트**
   - 실제 MoAI-ADK 프로젝트에서 Session Notice Hook 동작 확인
   - Claude Code 세션 시작 시 Welcome 메시지 표시 검증

## 기술적 접근 방법

### 1. 파일 시스템 검증 전략
```typescript
// 동기 방식 사용 (Hook 특성상 비동기 불필요)
fs.existsSync(path.join(projectRoot, '.moai'))
fs.existsSync(path.join(projectRoot, '.claude', 'commands', 'alfred'))
```

**근거**:
- Session Notice Hook은 동기 실행 환경
- 파일 시스템 접근은 로컬 디스크 (네트워크 I/O 없음)
- 성능 영향 미미 (2개 디렉토리 체크)

### 2. 빌드 프로세스
```javascript
// tsup.hooks.config.ts 설정
export default defineConfig({
  entry: ['src/claude/hooks/session-notice/index.ts'],
  format: ['cjs'],
  outDir: 'templates/.claude/hooks/alfred',
  // ...
});
```

**빌드 체인**:
1. TypeScript → JavaScript 트랜스파일
2. 의존성 번들링 (path, fs 모듈)
3. CommonJS 포맷 출력
4. 템플릿 디렉토리 배포

### 3. 에러 처리
```typescript
try {
  return fs.existsSync(moaiDir) && fs.existsSync(alfredCommands);
} catch (error) {
  // 파일 시스템 접근 오류 시 false 반환 (안전 실패)
  return false;
}
```

## 아키텍처 설계 방향

### 디렉토리 구조
```
moai-adk-ts/
├── src/
│   └── claude/
│       └── hooks/
│           └── session-notice/
│               ├── index.ts          # Hook 진입점
│               └── utils.ts          # @CODE:INIT-002 (수정 대상)
├── tests/
│   └── claude/
│       └── hooks/
│           └── session-notice/
│               └── utils.test.ts     # @TEST:INIT-002 (신규 생성)
└── templates/
    └── .claude/
        └── hooks/
            └── alfred/
                └── session-notice.cjs # 빌드 출력
```

### 의존성 다이어그램
```
Session Start (Claude Code)
    ↓
session-notice.cjs (Hook)
    ↓
isMoAIProject(projectRoot)
    ├─→ .moai 존재?
    └─→ .claude/commands/alfred 존재?
        ↓
    true → Welcome 메시지
    false → 무시
```

## 리스크 및 대응 방안

### R1: 빌드 실패 위험
**증상**: `npm run build:hooks` 실행 시 오류
**원인**: tsup 설정 또는 TypeScript 컴파일 오류
**대응**:
1. `npm run build` 전체 빌드로 타입 체크
2. `tsup --config tsup.hooks.config.ts --verbose` 상세 로그 확인
3. `node_modules` 재설치 (`npm ci`)

### R2: 기존 사용자 영향
**증상**: 기존 프로젝트에서 Welcome 메시지 미표시
**원인**: `.claude/commands/alfred` 미생성
**대응**:
1. `/alfred:8-project` 재실행 안내
2. 마이그레이션 가이드 문서 작성 (선택)
3. Session Notice에 업그레이드 안내 메시지 추가 (선택)

### R3: 테스트 환경 격리 실패
**증상**: 테스트 간 파일 시스템 상태 공유
**원인**: Mock 디렉토리 정리 누락
**대응**:
1. `afterEach()` 훅에서 임시 디렉토리 삭제
2. `vi.mock('fs')` 모킹 사용 (의존성 격리)
3. 테스트 순서 독립성 보장

## 검증 체크리스트

### 기능 검증
- [ ] 유닛 테스트 100% 통과
- [ ] `.moai` + `.claude/commands/alfred` 동시 검증 로직 정상
- [ ] 부재 시 `false` 반환 확인

### 빌드 검증
- [ ] `.cjs` 파일 생성 성공
- [ ] 번들 크기 변화 없음 (±5% 이내)
- [ ] `alfred` 경로 포함, `moai` 경로 미포함

### 통합 검증
- [ ] 실제 프로젝트에서 Session Notice 동작
- [ ] Claude Code 세션 시작 시 Welcome 메시지 표시
- [ ] 비-MoAI 프로젝트에서 메시지 미표시

### 문서 검증
- [ ] TAG 체인 완전성 (`@SPEC → @TEST → @CODE`)
- [ ] HISTORY 섹션 작성
- [ ] 다음 단계 안내 (`/alfred:2-build`, `/alfred:3-sync`)

## 완료 조건 (Definition of Done)

1. ✅ `utils.test.ts` 테스트 케이스 4개 작성 및 통과
2. ✅ `utils.ts` 라인 24 `alfred` 경로로 수정
3. ✅ `session-notice.cjs` 빌드 성공 및 검증
4. ✅ TAG 체인 완전성 확인 (`rg '@(SPEC|TEST|CODE):INIT-002' -n`)
5. ✅ SPEC 문서 HISTORY 업데이트 (구현 완료 기록)
6. ✅ `/alfred:3-sync` 실행 준비 완료

---

**다음 단계**: `/alfred:2-build SPEC-INIT-002`로 TDD 구현 시작
