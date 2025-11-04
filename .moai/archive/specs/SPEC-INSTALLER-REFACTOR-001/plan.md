# SPEC-INSTALLER-REFACTOR-001 구현 계획서

## 📋 개요

- **SPEC ID**: INSTALLER-REFACTOR-001
- **제목**: LOC 제한 준수 리팩토링
- **예상 기간**: 3-4시간
- **우선순위**: High

---

## 🎯 목표

phase-executor.ts (358 LOC) 및 template-processor.ts (371 LOC)를 300 LOC 이하로 분리하여 TRUST 원칙 준수

---

## 🔧 구현 단계

### Phase 1: phase-executor.ts 분리 (1.5h)

**신규 파일 생성**:
1. `backup-manager.ts` - 백업 로직 (<100 LOC)
2. `directory-builder.ts` - 디렉토리 생성 (<100 LOC)
3. `git-initializer.ts` - Git 초기화 (<100 LOC)

**리팩토링**:
- phase-executor.ts: 358 → ~180 LOC
- 의존성 주입 패턴 적용

---

### Phase 2: template-processor.ts 분리 (1.5h)

**신규 파일 생성**:
1. `template-path-resolver.ts` - 경로 해석 (<150 LOC)
2. `template-renderer.ts` - 렌더링 (<100 LOC)

**리팩토링**:
- template-processor.ts: 371 → ~150 LOC
- 기존 API 유지, 내부 구현만 변경

---

### Phase 3: 테스트 작성 (1h)

**신규 테스트**:
- `backup-manager.test.ts`
- `directory-builder.test.ts`
- `git-initializer.test.ts`
- `template-path-resolver.test.ts`
- `template-renderer.test.ts`

**기존 테스트 수정**:
- `phase-executor.test.ts` (DI mock 적용)
- `template-processor.test.ts` (DI mock 적용)

---

## ✅ 체크리스트

### 파일 분리
- [ ] backup-manager.ts 생성 및 로직 이동
- [ ] directory-builder.ts 생성 및 로직 이동
- [ ] git-initializer.ts 생성 및 로직 이동
- [ ] template-path-resolver.ts 생성 및 로직 이동
- [ ] template-renderer.ts 생성 및 로직 이동

### LOC 검증
- [ ] phase-executor.ts ≤ 300 LOC
- [ ] template-processor.ts ≤ 300 LOC
- [ ] 모든 신규 파일 ≤ 300 LOC
- [ ] 모든 함수 ≤ 50 LOC

### 테스트
- [ ] 신규 클래스별 테스트 작성
- [ ] 기존 테스트 수정 (DI 적용)
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 테스트 통과

### 품질
- [ ] ESLint 검사 통과
- [ ] TypeScript 컴파일 성공
- [ ] 성능 저하 < 5% 확인

---

## 📅 일정

| Phase | 작업 | 예상 시간 |
|-------|------|----------|
| 1 | phase-executor.ts 분리 | 1.5h |
| 2 | template-processor.ts 분리 | 1.5h |
| 3 | 테스트 작성 | 1h |
| **Total** | | **4h** |

---

## 🚨 주의사항

- **public API 변경 금지**: 외부에서 사용 중인 메서드 시그니처 유지
- **후방 호환성**: 모든 기존 테스트가 통과해야 함
- **작은 단계**: 한 번에 하나의 클래스만 분리
- **테스트 먼저**: 로직 이동 전 테스트 작성
