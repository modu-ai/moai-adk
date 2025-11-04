# SPEC-001 수락 기준

## TAG BLOCK

```text
# @SPEC:REFACTOR-001: Git Manager 리팩토링 수락 기준
# Parent: SPEC-REFACTOR-001
# Phase 1: GitBranchManager (REFACTOR-001-P1)
# Phase 2: GitCommitManager (REFACTOR-001-P2)
# Phase 3: GitPRManager (REFACTOR-001-P3)
```

## 수락 기준 개요

이 문서는 SPEC-001 (Git Manager 리팩토링)의 완료를 판단하는 구체적이고 측정 가능한 기준을 정의합니다. 모든 기준을 충족해야 리팩토링이 완료된 것으로 간주됩니다.

---

## 1. 코드 품질 기준

### 1.1 Lines of Code (LOC)

**기준**:
- ✅ `git-manager.ts` ≤ 150 LOC
- ✅ `git-branch-manager.ts` ≤ 200 LOC
- ✅ `git-commit-manager.ts` ≤ 200 LOC
- ✅ `git-pr-manager.ts` ≤ 150 LOC

**검증 방법**:
```bash
wc -l src/core/git/git-manager.ts
wc -l src/core/git/git-branch-manager.ts
wc -l src/core/git/git-commit-manager.ts
wc -l src/core/git/git-pr-manager.ts
```

**Given**: 리팩토링 완료 후
**When**: LOC 측정 명령어 실행
**Then**: 모든 파일이 목표 LOC 이하여야 함

---

### 1.2 함수 복잡도

**기준**:
- ✅ 모든 함수의 순환 복잡도 ≤ 10
- ✅ 함수당 LOC ≤ 50
- ✅ 매개변수 개수 ≤ 5

**검증 방법**:
```bash
# 복잡도 분석 (eslint-plugin-complexity 또는 수동 검토)
npm run lint
```

**Given**: 각 매니저 파일
**When**: 복잡도 분석 실행
**Then**: 모든 함수가 복잡도 기준을 충족해야 함

---

### 1.3 코드 스타일

**기준**:
- ✅ Biome 린트 규칙 100% 통과
- ✅ Biome 포맷 규칙 준수
- ✅ TypeScript strict 모드 통과

**검증 방법**:
```bash
npm run lint
npm run format:check
npx tsc --noEmit
```

**Given**: 모든 변경된 파일
**When**: 린트/포맷/타입 체크 실행
**Then**: 0개의 에러, 0개의 경고

---

## 2. 테스트 품질 기준

### 2.1 테스트 커버리지

**기준**:
- ✅ 전체 테스트 커버리지 ≥ 85%
- ✅ `git-branch-manager.ts` 커버리지 ≥ 90%
- ✅ `git-commit-manager.ts` 커버리지 ≥ 90%
- ✅ `git-pr-manager.ts` 커버리지 ≥ 90%
- ✅ Critical path 커버리지 = 100%

**검증 방법**:
```bash
npm test -- --coverage
```

**Given**: 모든 테스트 작성 완료
**When**: 커버리지 리포트 생성
**Then**: 모든 커버리지 목표 달성

---

### 2.2 테스트 통과율

**기준**:
- ✅ 모든 기존 통합 테스트 100% 통과
- ✅ 모든 신규 단위 테스트 100% 통과
- ✅ 테스트 실행 시간 증가율 ≤ 10%

**검증 방법**:
```bash
npm test
```

**Given**: 리팩토링 완료 후
**When**: 전체 테스트 스위트 실행
**Then**:
- 0개의 실패 테스트
- 0개의 스킵된 테스트
- 기존 대비 ≤ 10% 시간 증가

---

### 2.3 테스트 시나리오 완성도

**기준**:
- ✅ 각 매니저별 최소 15개 이상의 테스트 케이스
- ✅ 정상 케이스, Edge 케이스, 에러 케이스 모두 포함
- ✅ Lock 통합 시나리오 테스트 포함

**검증 방법**:
```bash
npm test -- --reporter=verbose
```

**Given**: 테스트 스위트
**When**: 테스트 목록 확인
**Then**: 모든 카테고리의 테스트 케이스 존재

---

## 3. 기능 수락 기준

### 3.1 API 호환성

**Scenario 1: 브랜치 생성**

**Given**: 기존 코드에서 `gitManager.createBranch()` 호출
**When**: 리팩토링 후 동일한 코드 실행
**Then**:
- 정상적으로 브랜치 생성됨
- 반환값 형식 동일
- 에러 처리 동일

```typescript
// 기존 코드가 수정 없이 작동해야 함
const gitManager = new GitManager(config);
await gitManager.createBranch('feature/test');
// ✅ 성공
```

---

**Scenario 2: 커밋 생성**

**Given**: 기존 코드에서 `gitManager.commitChanges()` 호출
**When**: 리팩토링 후 동일한 코드 실행
**Then**:
- 정상적으로 커밋 생성됨
- GitCommitResult 형식 동일
- 템플릿 적용 동일

```typescript
const result = await gitManager.commitChanges('test commit');
// ✅ result.hash 존재
// ✅ result.message 존재
// ✅ result.timestamp 존재
```

---

**Scenario 3: PR 생성**

**Given**: Team 모드에서 `gitManager.createPullRequest()` 호출
**When**: 리팩토링 후 동일한 코드 실행
**Then**:
- 정상적으로 PR 생성됨
- PR URL 반환
- Draft 옵션 작동

```typescript
const prUrl = await gitManager.createPullRequest({
  title: 'Test PR',
  body: 'Test body',
  baseBranch: 'main',
  draft: true
});
// ✅ prUrl 존재
```

---

### 3.2 Lock 통합 작동

**Scenario 4: 동시성 제어**

**Given**: 동시에 여러 Git 작업 요청
**When**: Lock 기반 메서드 사용
**Then**:
- 한 번에 하나의 작업만 실행
- 대기 중인 작업은 순차 실행
- Lock 해제 정상 작동

```typescript
// 동시 실행 시도
const [result1, result2] = await Promise.all([
  gitManager.commitWithLock('commit 1'),
  gitManager.commitWithLock('commit 2')
]);
// ✅ 순차적으로 실행됨
// ✅ 둘 다 성공
```

---

**Scenario 5: Lock 타임아웃**

**Given**: Lock을 오래 점유하는 작업 진행 중
**When**: 타임아웃 설정으로 새 작업 시도
**Then**:
- 타임아웃 초과 시 GitLockedException 발생
- Lock 상태 정보 포함

```typescript
try {
  await gitManager.commitWithLock('commit', undefined, true, 1); // 1초 타임아웃
} catch (error) {
  // ✅ GitLockedException 발생
  // ✅ error.timeout === 1
}
```

---

### 3.3 에러 처리

**Scenario 6: 잘못된 브랜치명**

**Given**: 유효하지 않은 브랜치명 입력
**When**: `createBranch()` 호출
**Then**:
- 검증 에러 발생
- 명확한 에러 메시지
- 브랜치 생성 안됨

```typescript
try {
  await gitManager.createBranch('invalid/branch/../name');
} catch (error) {
  // ✅ 에러 발생
  // ✅ error.message 포함: "Branch name validation failed"
}
```

---

**Scenario 7: GitHub CLI 미설치**

**Given**: GitHub CLI가 설치되지 않은 환경
**When**: `createPullRequest()` 호출 (Team 모드)
**Then**:
- 명확한 에러 메시지
- 설치 안내 포함

```typescript
try {
  await gitManager.createPullRequest({...});
} catch (error) {
  // ✅ error.message 포함: "GitHub CLI is not installed"
}
```

---

## 4. 성능 수락 기준

### 4.1 성능 유지

**기준**:
- ✅ 브랜치 생성 시간 증가율 ≤ 5%
- ✅ 커밋 생성 시간 증가율 ≤ 5%
- ✅ PR 생성 시간 증가율 ≤ 5%

**검증 방법**:
```bash
# 성능 벤치마크 실행
npm run benchmark
```

**Given**: 리팩토링 전후 벤치마크 데이터
**When**: 각 작업의 평균 실행 시간 비교
**Then**: 모든 작업의 시간 증가율 ≤ 5%

---

### 4.2 메모리 사용량

**기준**:
- ✅ 메모리 사용량 증가율 ≤ 10%
- ✅ 메모리 누수 없음

**검증 방법**:
```bash
# 메모리 프로파일링
node --expose-gc --inspect test-memory.js
```

**Given**: 반복적인 Git 작업 실행
**When**: 메모리 사용량 모니터링
**Then**:
- 메모리 증가율 ≤ 10%
- GC 후 메모리 정상 회수

---

## 5. 구조적 수락 기준

### 5.1 모듈 독립성

**Scenario 8: 매니저 개별 인스턴스화**

**Given**: GitBranchManager를 단독으로 사용
**When**: GitManager 없이 인스턴스 생성
**Then**:
- 정상적으로 생성됨
- 모든 기능 작동
- 순환 의존성 없음

```typescript
const git = simpleGit();
const lockManager = new GitLockManager();
const branchManager = new GitBranchManager(git, lockManager);

await branchManager.createBranch('test');
// ✅ 성공
```

---

### 5.2 의존성 방향

**기준**:
- ✅ GitManager → *Manager (단방향)
- ✅ *Manager → GitLockManager (단방향)
- ✅ 순환 의존성 0건

**검증 방법**:
```bash
# 의존성 그래프 분석
npm run analyze-deps
# 또는 madge 도구 사용
npx madge --circular src/core/git
```

**Given**: 모든 Git 관련 모듈
**When**: 의존성 그래프 생성
**Then**: 순환 의존성 검출 안됨

---

## 6. 문서화 수락 기준

### 6.1 코드 문서화

**기준**:
- ✅ 모든 public 메서드에 JSDoc 주석
- ✅ 모든 파일에 TAG BLOCK
- ✅ 복잡한 로직에 설명 주석

**검증 방법**:
```bash
# TypeDoc 문서 생성
npm run docs
```

**Given**: 모든 변경된 파일
**When**: TypeDoc 실행
**Then**:
- 0개의 문서화 경고
- 모든 public API 문서화됨

---

### 6.2 TAG 추적성

**Scenario 9: TAG 체인 검증**

**Given**: 모든 파일에 TAG BLOCK 추가 완료
**When**: TAG 검증 명령어 실행
**Then**:
- 모든 TAG 체인 연결됨
- 고아 TAG 없음
- TAG 형식 준수

```bash
rg '@CODE:REFACTOR-001' -n src/core/git/
rg '@SPEC:REFACTOR-001' -n src/core/git/
rg '@SPEC:REFACTOR-001' -n src/core/git/
rg '@CODE:REFACTOR-001' -n src/core/git/
rg '@TEST:REFACTOR-001' -n src/core/git/
# ✅ 모든 TAG 발견됨
```

---

## 7. 통합 시나리오

### 7.1 전체 워크플로우

**Scenario 10: 브랜치 → 커밋 → 푸시 → PR 생성**

**Given**: 새로운 기능 개발 시작
**When**: 전체 Git 워크플로우 실행
**Then**: 모든 단계 성공

```typescript
const gitManager = new GitManager({
  mode: 'team',
  autoCommit: false,
  branchPrefix: 'feature/',
  commitMessageTemplate: 'conventional',
});

// 1. 브랜치 생성
await gitManager.createBranch('feature/new-feature');
// ✅ 성공

// 2. 파일 변경 및 커밋
await gitManager.commitChanges('feat: add new feature', ['src/index.ts']);
// ✅ 성공

// 3. 원격으로 푸시
await gitManager.pushChanges();
// ✅ 성공

// 4. PR 생성
const prUrl = await gitManager.createPullRequest({
  title: 'feat: Add new feature',
  body: 'This PR adds a new feature',
  baseBranch: 'main',
  draft: true
});
// ✅ PR URL 반환
```

---

### 7.2 Personal 모드 워크플로우

**Scenario 11: Personal 모드 전체 흐름**

**Given**: Personal 모드 설정
**When**: 로컬 개발 워크플로우 실행
**Then**: GitHub 통합 없이 정상 작동

```typescript
const gitManager = new GitManager({
  mode: 'personal',
  autoCommit: true,
  branchPrefix: 'dev/',
  commitMessageTemplate: 'simple',
});

// GitHub 관련 메서드는 에러 발생
try {
  await gitManager.createPullRequest({...});
} catch (error) {
  // ✅ "only available in team mode" 에러
}

// 로컬 Git 작업은 정상 작동
await gitManager.createBranch('dev/test');
await gitManager.commitChanges('test commit');
// ✅ 모두 성공
```

---

## 8. 회귀 테스트 (Regression Tests)

### 8.1 기존 기능 보존

**Scenario 12: 기존 통합 테스트 100% 통과**

**Given**: `__tests__/core/git/git-manager.test.ts` 파일
**When**: 테스트 실행
**Then**:
- 모든 테스트 통과
- 코드 수정 없이 통과
- 테스트 실행 시간 증가 ≤ 10%

```bash
npm test __tests__/core/git/git-manager.test.ts
# ✅ 100% 통과
```

---

### 8.2 Edge Case 처리

**Scenario 13: 빈 저장소에서 첫 커밋**

**Given**: 커밋이 없는 새 저장소
**When**: `commitChanges()` 호출
**Then**:
- 자동으로 README.md 생성
- 초기 커밋 생성
- 정상적으로 작동

```typescript
// 새 저장소
await gitManager.initializeRepository(projectPath);
await gitManager.commitChanges('Initial commit');
// ✅ 성공
```

---

**Scenario 14: 이미 존재하는 브랜치**

**Given**: 이미 존재하는 브랜치명
**When**: `createBranch()` 호출
**Then**:
- 명확한 에러 메시지
- 기존 브랜치 보호

```typescript
await gitManager.createBranch('feature/test');
try {
  await gitManager.createBranch('feature/test');
} catch (error) {
  // ✅ 에러 발생
}
```

---

## 9. 품질 게이트 체크리스트

### Phase 1 완료 조건 (GitBranchManager)

- [ ] `git-branch-manager.ts` ≤ 200 LOC
- [ ] 단위 테스트 ≥ 15개
- [ ] 테스트 커버리지 ≥ 90%
- [ ] 모든 테스트 통과
- [ ] Biome 린트 통과
- [ ] TAG BLOCK 추가
- [ ] JSDoc 주석 완료
- [ ] GitManager 위임 구현
- [ ] 기존 통합 테스트 통과

---

### Phase 2 완료 조건 (GitCommitManager)

- [ ] `git-commit-manager.ts` ≤ 200 LOC
- [ ] 단위 테스트 ≥ 15개
- [ ] 테스트 커버리지 ≥ 90%
- [ ] 모든 테스트 통과
- [ ] Biome 린트 통과
- [ ] TAG BLOCK 추가
- [ ] JSDoc 주석 완료
- [ ] GitManager 위임 구현
- [ ] 기존 통합 테스트 통과

---

### Phase 3 완료 조건 (GitPRManager)

- [ ] `git-pr-manager.ts` ≤ 150 LOC
- [ ] 단위 테스트 ≥ 15개
- [ ] 테스트 커버리지 ≥ 90%
- [ ] 모든 테스트 통과
- [ ] Biome 린트 통과
- [ ] TAG BLOCK 추가
- [ ] JSDoc 주석 완료
- [ ] GitManager 위임 구현
- [ ] 기존 통합 테스트 통과

---

### Phase 4 완료 조건 (GitManager 최종)

- [ ] `git-manager.ts` ≤ 150 LOC
- [ ] 모든 위임 메서드 구현
- [ ] 기존 통합 테스트 100% 통과
- [ ] 전체 테스트 커버리지 ≥ 85%
- [ ] 성능 저하 ≤ 5%
- [ ] 메모리 증가 ≤ 10%
- [ ] 순환 의존성 0건
- [ ] TAG 체인 검증 완료
- [ ] API 문서 생성 완료

---

## 10. Definition of Done (완료 정의)

### 최종 수락 조건

리팩토링이 완료되었다고 판단하려면 다음 모든 조건을 만족해야 합니다:

#### 코드 품질
- [x] 모든 파일 LOC 목표 달성
- [x] 함수 복잡도 기준 충족
- [x] Biome 린트/포맷 100% 통과
- [x] TypeScript strict 모드 통과

#### 테스트 품질
- [x] 전체 커버리지 ≥ 85%
- [x] 각 매니저 커버리지 ≥ 90%
- [x] 모든 테스트 100% 통과
- [x] 테스트 시간 증가 ≤ 10%

#### 기능 완성도
- [x] 모든 기존 API 호환성 유지
- [x] Lock 통합 정상 작동
- [x] 에러 처리 완벽
- [x] Personal/Team 모드 모두 지원

#### 성능
- [x] 성능 저하 ≤ 5%
- [x] 메모리 증가 ≤ 10%
- [x] 메모리 누수 없음

#### 구조
- [x] 순환 의존성 0건
- [x] 단방향 의존성 준수
- [x] 모듈 독립성 확보

#### 문서화
- [x] 모든 public API JSDoc 완료
- [x] TAG BLOCK 모든 파일 추가
- [x] TAG 체인 검증 완료
- [x] TypeDoc 문서 생성 성공

#### 회귀 방지
- [x] 기존 통합 테스트 100% 통과
- [x] Edge case 모두 처리
- [x] 전체 워크플로우 정상 작동

---

## 검증 스크립트

### 자동 검증 스크립트

```bash
#!/bin/bash

echo "🔍 SPEC-001 수락 기준 검증 시작..."

# 1. LOC 검증
echo "\n📏 1. LOC 검증..."
LOC_GM=$(wc -l < src/core/git/git-manager.ts)
LOC_BM=$(wc -l < src/core/git/git-branch-manager.ts)
LOC_CM=$(wc -l < src/core/git/git-commit-manager.ts)
LOC_PM=$(wc -l < src/core/git/git-pr-manager.ts)

[ $LOC_GM -le 150 ] && echo "✅ git-manager.ts: $LOC_GM LOC" || echo "❌ git-manager.ts: $LOC_GM LOC (목표: 150)"
[ $LOC_BM -le 200 ] && echo "✅ git-branch-manager.ts: $LOC_BM LOC" || echo "❌ git-branch-manager.ts: $LOC_BM LOC (목표: 200)"
[ $LOC_CM -le 200 ] && echo "✅ git-commit-manager.ts: $LOC_CM LOC" || echo "❌ git-commit-manager.ts: $LOC_CM LOC (목표: 200)"
[ $LOC_PM -le 150 ] && echo "✅ git-pr-manager.ts: $LOC_PM LOC" || echo "❌ git-pr-manager.ts: $LOC_PM LOC (목표: 150)"

# 2. 린트 검증
echo "\n🔧 2. 린트 검증..."
npm run lint || exit 1
echo "✅ Biome 린트 통과"

# 3. 타입 체크
echo "\n📘 3. 타입 체크..."
npx tsc --noEmit || exit 1
echo "✅ TypeScript 타입 체크 통과"

# 4. 테스트 실행
echo "\n🧪 4. 테스트 실행..."
npm test || exit 1
echo "✅ 모든 테스트 통과"

# 5. 커버리지 검증
echo "\n📊 5. 커버리지 검증..."
npm test -- --coverage || exit 1
echo "✅ 커버리지 기준 충족"

# 6. TAG 검증
echo "\n🏷️  6. TAG 체인 검증..."
rg '@CODE:REFACTOR-001' src/core/git/ -q && echo "✅ @CODE TAG 존재" || echo "❌ @CODE TAG 누락"
rg '@SPEC:REFACTOR-001' src/core/git/ -q && echo "✅ @SPEC TAG 존재" || echo "❌ @SPEC TAG 누락"
rg '@SPEC:REFACTOR-001' src/core/git/ -q && echo "✅ @SPEC TAG 존재" || echo "❌ @SPEC TAG 누락"
rg '@CODE:REFACTOR-001' src/core/git/ -q && echo "✅ @CODE TAG 존재" || echo "❌ @CODE TAG 누락"
rg '@TEST:REFACTOR-001' src/core/git/ -q && echo "✅ @TEST TAG 존재" || echo "❌ @TEST TAG 누락"

# 7. 순환 의존성 검증
echo "\n🔄 7. 순환 의존성 검증..."
npx madge --circular src/core/git && echo "❌ 순환 의존성 발견" || echo "✅ 순환 의존성 없음"

echo "\n✨ 모든 검증 완료!"
```

---

## 수동 검증 체크리스트

### 리뷰어 체크리스트

- [ ] 코드 리뷰 완료
- [ ] 아키텍처 설계 검토 완료
- [ ] 의존성 방향 확인 완료
- [ ] 테스트 커버리지 확인 완료
- [ ] 성능 벤치마크 확인 완료
- [ ] 문서화 품질 확인 완료
- [ ] TAG 체인 추적성 확인 완료

### 개발자 셀프 체크리스트

- [ ] 모든 Phase 완료
- [ ] 자동 검증 스크립트 통과
- [ ] Edge case 모두 테스트
- [ ] 에러 메시지 명확성 확인
- [ ] 성능 프로파일링 완료
- [ ] 메모리 누수 검사 완료
- [ ] 문서 동기화 완료 (`/alfred:3-sync`)

---

**작성일**: 2025-10-01
**버전**: 1.0
**상태**: Draft
