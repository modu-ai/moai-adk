# Git Worktree CLI 실전 예제

## 예제 1: 병렬 SPEC 개발 (가장 일반적)

### 시나리오
팀이 동시에 3개의 기능을 개발해야 함:
- SPEC-002: 사용자 인증 (개발자 A)
- SPEC-003: 결제 시스템 (개발자 B)
- SPEC-004: 대시보드 (개발자 C)

### 실행 단계

**Step 1: 각 개발자가 자신의 워크트리 생성**

```bash
# 개발자 A
moai-wt new SPEC-002
wt-go SPEC-002

# 개발자 B
moai-wt new SPEC-003
wt-go SPEC-003

# 개발자 C
moai-wt new SPEC-004
wt-go SPEC-004
```

**Step 2: 각각 SPEC 계획 및 구현**

```bash
# 각 워크트리에서
/moai:1-plan "기능 설명" --worktree
# → SPEC 생성 및 워크트리 생성 (자동)

/moai:2-run SPEC-00X
# → TDD 사이클로 구현
```

**Step 3: 완료 및 병합**

```bash
# SPEC-002 완료 (개발자 A)
wt-go SPEC-002
git add .
git commit -m "feat(auth): implement user authentication"
moai-wt remove SPEC-002

# 문서 동기화
/moai:3-sync SPEC-002
```

**Step 4: 메인 브랜치 병합 (팀 리더)**

```bash
# 모든 SPEC 완료 후
git checkout main
git merge feature/SPEC-002
git merge feature/SPEC-003
git merge feature/SPEC-004
```

### 예상 결과
- 3개 기능 병렬 개발
- 개발 시간 50-60% 단축 (대역폭에 따라)
- 충돌 최소화 (독립적인 파일 변경)
- 팀 생산성 향상

---

## 예제 2: 긴급 패치 (Hotfix 워크플로우)

### 시나리오
본 개발 중에 프로덕션 버그 발견

### 실행 단계

```bash
# Step 1: 현재 작업 상태 확인
moai-wt status

# Step 2: Main 브랜치 기반 긴급 패치 워크트리 생성
moai-wt new hotfix/payment-crash --base main
wt-go hotfix/payment-crash

# Step 3: 버그 수정
# 코드 수정 작업...
git add src/payment/processor.py
git commit -m "fix: prevent payment crash on timeout"

# Step 4: 테스트 실행
npm test
npm run lint

# Step 5: 정리
moai-wt remove hotfix/payment-crash

# Step 6: 수동 병합 (즉시 반영)
git checkout main
git merge hotfix/payment-crash
git push origin main

# 또는 PR 생성
gh pr create --title "fix: payment crash" --base main --head hotfix/payment-crash
```

### 예상 결과
- 메인 개발 중단 없음
- 즉시 버그 패치 가능
- 원본 작업 상태 유지

---

## 예제 3: 대규모 리팩토링

### 시나리오
코드 품질 개선을 위한 리팩토링 필요

### 실행 단계

```bash
# Step 1: 리팩토링 워크트리 생성
moai-wt new refactor/auth-module

# Step 2: 격리된 환경에서 대규모 변경
wt-go refactor/auth-module
# 코드 리팩토링...
# 테스트 작성...
# 통합 테스트 실행...

# Step 3: 품질 검증
npm run test:coverage
npm run lint
npm run type-check

# Step 4: 리뷰 및 병합
git push origin refactor/auth-module
gh pr create --title "refactor: improve auth module structure"

# PR 리뷰 완료 후
moai-wt remove refactor/auth-module
```

### 예상 결과
- 본 개발과 격리된 안전한 리팩토링
- CI/CD 파이프라인 연동
- 팀 리뷰 프로세스 적용

---

## 예제 4: 기능 토글 개발

### 시나리오
미래 기능을 준비하되, 아직 활성화하지 않음

### 실행 단계

```bash
# Step 1: 기능 워크트리 생성
moai-wt new feature/new-payment-gateway

# Step 2: 기능 개발 (토글 기반)
wt-go feature/new-payment-gateway

// 코드 내에서 기능 토글 사용
if (featureFlags.newPaymentGateway) {
  // 새로운 결제 게이트웨이 로직
}

# Step 3: 테스트
npm test  # 기존 테스트는 토글로 격리

# Step 4: 병합 (토글 비활성화)
moai-wt remove feature/new-payment-gateway
git merge feature/new-payment-gateway  # 안전 - 토글로 비활성화됨

# 이후 필요할 때 토글 활성화
```

### 예상 결과
- 장기 개발 프로젝트 관리
- 메인 브랜치 안정성 유지
- 빠른 온보딩 가능

---

## 예제 5: 팀 협업 및 코드 리뷰

### 시나리오
4명 팀에서 다양한 기능 개발

### 실행 단계

```bash
# 팀 멤버들의 병렬 작업
moai-wt new SPEC-010  # 리더 - 리뷰 및 조정
moai-wt new SPEC-011  # 개발자 A - 백엔드
moai-wt new SPEC-012  # 개발자 B - 프론트엔드
moai-wt new SPEC-013  # 개발자 C - 인프라

# 각자 워크트리에서 개발
wt-go SPEC-011
/moai:2-run SPEC-011
# ... 개발 ...

# 주기적으로 진행상황 공유
moai-wt list
# 리더가 전체 상황 파악

# 완료 시 PR 생성 및 리뷰
gh pr create --title "feat(backend): SPEC-011 implementation"

# 리뷰 완료 후 병합
moai-wt remove SPEC-011
```

### 예상 결과
- 팀 전체의 병렬 진행
- 리뷰 단계 효율화
- 명확한 진행상황 추적

---

## 문제 상황 및 해결

### 상황 1: Merge 충돌 발생

```bash
# 동기화 시도
moai-wt sync SPEC-001
# → "Merge conflict detected"

# 해결
wt-go SPEC-001
# 충돌 파일 확인 및 수정
git add .
git commit -m "Resolve merge conflicts from main"
```

### 상황 2: 워크트리 실수로 삭제

```bash
# 복구 (Git 워크트리는 복구 불가, 다시 생성)
moai-wt new SPEC-001
# 이전 커밋 확인
git log --oneline feature/SPEC-001
# 해당 커밋 있으면 복구 가능
```

### 상황 3: 너무 많은 워크트리

```bash
# 정리
moai-wt clean  # 병합된 것만 정리
moai-wt list   # 활성 확인

# 수동 정리 필요시
moai-wt remove SPEC-OLD --force
```

---

## 성능 팁

### Worktree 개수 최적화
- 3-4개: 최적 (대부분의 프로젝트)
- 5개 이상: 시스템 리소스 주의
- 10개 이상: 권장하지 않음

### 디스크 공간 관리
```bash
# 전체 worktree 크기 확인
du -sh ~/worktrees/PROJECT

# 큰 파일 확인
find ~/worktrees/PROJECT -size +100M
```

---

**더 많은 정보**: [Git Worktree 사용 가이드](./WORKTREE_GUIDE.md)
