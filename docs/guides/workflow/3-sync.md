# Stage 3: Document Synchronization

`/alfred:3-sync` 커맨드를 사용하여 Living Document를 생성하고 TAG 체인을 검증합니다.

## Overview

Document Synchronization은 MoAI-ADK 3단계 워크플로우의 마지막 단계입니다. **"추적성 없이는 완성 없음"** 원칙을 따라 코드와 문서를 동기화하고, 전체 품질을 검증합니다.

### 담당 에이전트

- **doc-syncer** 📖: 테크니컬 라이터
- **역할**: TAG 스캔, Living Document 생성, PR 상태 업데이트
- **전문성**: 문서 동기화, TAG 체인 검증, TRUST 원칙 검증, Git 워크플로우

---

## When to Use

다음과 같은 경우 `/alfred:3-sync`를 사용합니다:

- ✅ `/alfred:2-build`로 TDD 구현이 완료되었을 때
- ✅ 모든 테스트가 통과했을 때
- ✅ PR을 머지하기 전 최종 검증이 필요할 때
- ✅ Living Document를 업데이트하고 싶을 때
- ✅ TAG 체인 무결성을 검증하고 싶을 때

---

## Command Syntax

### Basic Usage

```bash
# 기본 동기화
/alfred:3-sync

# 검증만 수행 (동기화 없이)
/alfred:3-sync --check

# Team 모드: 자동 머지 (CI/CD 통과 후)
/alfred:3-sync --auto-merge
```

### Advanced Usage

```bash
# 특정 SPEC만 동기화
/alfred:3-sync AUTH-001

# 여러 SPEC 동기화
/alfred:3-sync AUTH-001 UPLOAD-003

# 특정 경로만 스캔
/alfred:3-sync --path=src/auth
```

---

## Workflow (2단계)

### Phase 1: 분석 및 검증

Alfred가 다음 작업을 수행합니다:

#### 1. 프로젝트 상태 분석

```bash
# Git 상태 확인
git status
git branch

# 현재 브랜치 확인
git rev-parse --abbrev-ref HEAD
```

Alfred가 확인하는 항목:
- **현재 브랜치**: feature/SPEC-XXX 형식인지
- **변경사항**: staged/unstaged 파일 목록
- **PR 상태**: Draft/Ready 여부 (Team 모드)
- **CI/CD 상태**: 통과 여부 (Team 모드)

#### 2. TAG 체인 스캔

```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# SPEC별 TAG 추출
rg '@SPEC:([A-Z]+-[0-9]+)' -o -r '$1' .moai/specs/ | sort -u
```

**스캔 결과 예시**:

```
📋 TAG 체인 스캔 결과

검색된 TAG:
- @SPEC:AUTH-001 (1개)
- @TEST:AUTH-001 (1개)
- @CODE:AUTH-001 (1개)
- @DOC:AUTH-001 (0개)

TAG 체인 상태:
✅ SPEC-AUTH-001: SPEC → TEST → CODE (완전)
⚠️  SPEC-UPLOAD-003: SPEC → CODE (TEST 누락)
❌ ORPHAN: @CODE:PAYMENT-005 (SPEC 없음)

진행하시겠습니까? (진행/수정/중단)
```

#### 3. TAG 체인 검증

Alfred가 검증하는 항목:

| 검증 항목 | 설명 | 예시 |
|----------|------|------|
| **완전한 체인** | SPEC → TEST → CODE 모두 존재 | ✅ AUTH-001 |
| **불완전한 체인** | SPEC 존재, TEST 또는 CODE 누락 | ⚠️ UPLOAD-003 |
| **고아 TAG** | CODE/TEST 존재, SPEC 없음 | ❌ PAYMENT-005 |
| **중복 TAG** | 동일 TAG ID가 여러 곳에 존재 | ❌ AUTH-001 (2곳) |

#### 4. TRUST 원칙 검증

```bash
# T - Test First: 커버리지 확인
bun test --coverage
pytest --cov=src --cov-report=term-missing

# R - Readable: 린터 실행
biome check src/
ruff check src/

# U - Unified: 타입 체크
tsc --noEmit
mypy src/

# S - Secured: 보안 스캔
npm audit
bandit -r src/

# T - Trackable: TAG 무결성
rg '@(SPEC|TEST|CODE):' -n
```

**검증 보고서 예시**:

```markdown
✅ TRUST 검증 완료

### T - Test First
- ✅ 테스트 커버리지: 92% (목표 85% 초과)
- ✅ 모든 테스트 통과: 12/12

### R - Readable
- ✅ 린터 통과: 0 issues
- ✅ 파일 크기: 평균 156 LOC (≤300)
- ✅ 함수 복잡도: 최대 6 (≤10)

### U - Unified
- ✅ 타입 체크 통과
- ✅ 의존성 주입 패턴 사용

### S - Secured
- ✅ 보안 스캔: 0 vulnerabilities
- ✅ 입력 검증 구현

### T - Trackable
- ✅ TAG 체인 무결성 확인
- ✅ 고아 TAG 없음

**TRUST 점수**: 5/5 ✅
```

#### 5. 사용자 확인 대기

- **"진행"**: Phase 2로 이동
- **"수정 [내용]"**: 문제 해결 후 재실행
- **"중단"**: 작업 취소

---

### Phase 2: 문서 동기화 및 PR 처리

사용자가 "진행"하면 Alfred가 다음 작업을 수행합니다.

---

## Living Document 생성

### 1. Sync Report 생성

**`.moai/reports/sync-report-YYYY-MM-DD.md`**:

```markdown
# Sync Report - 2025-10-11T13:00:00Z

## Metadata
- **날짜**: 2025-10-11 13:00:00
- **브랜치**: feature/SPEC-AUTH-001
- **작성자**: @Goos
- **커밋**: a1b2c3d

---

## TAG Chain Summary

### Complete Chains (1)

#### SPEC-AUTH-001
- ✅ **SPEC**: .moai/specs/SPEC-AUTH-001/spec.md
- ✅ **TEST**: tests/auth/service.test.ts
- ✅ **CODE**: src/auth/service.ts
- ⚠️  **DOC**: not found (optional)

**Status**: Ready for review

---

### Incomplete Chains (0)

---

### Orphan TAGs (0)

---

## Test Coverage

| File | Coverage | Lines | Missing |
|------|----------|-------|---------|
| src/auth/service.ts | 92% | 156 | 45, 52 |
| **Total** | **92%** | **156** | **2** |

**Status**: ✅ Passed (≥85%)

---

## TRUST Compliance

| Principle | Status | Details |
|-----------|--------|---------|
| **T** - Test First | ✅ | 92% coverage, 4/4 tests passing |
| **R** - Readable | ✅ | 0 lint issues, complexity ≤10 |
| **U** - Unified | ✅ | TypeScript strict mode |
| **S** - Secured | ✅ | 0 vulnerabilities |
| **T** - Trackable | ✅ | TAG chain intact |

**TRUST Score**: 5/5 ✅

---

## Quality Metrics

- **코드 복잡도**: 평균 4.2 (최대 6)
- **파일 크기**: 평균 156 LOC
- **함수 크기**: 평균 18 LOC
- **매개변수**: 최대 3개

---

## Recommendations

✅ **Ready to Merge**
- All checks passed
- TAG chain complete
- TRUST compliance: 5/5

**Next Steps**:
1. Review PR changes
2. Merge to develop (Team mode)
3. Archive SPEC (optional)

---

**Generated by**: MoAI-ADK v0.2.17
**Command**: /alfred:3-sync
```

### 2. Feature Document 생성 (선택)

Alfred가 자동으로 생성하는 기능 문서 (선택적):

**`docs/features/auth/jwt-authentication.md`**:

```markdown
<!-- @DOC:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md -->

# JWT 인증 시스템

> **SPEC**: AUTH-001 | **Version**: 0.0.1 | **Status**: Active

## Overview

이 문서는 @SPEC:AUTH-001에 정의된 JWT 기반 인증 시스템을 설명합니다.

### 주요 기능

- ✅ 이메일/비밀번호 기반 인증
- ✅ JWT 토큰 발급 (15분 만료)
- ✅ bcrypt 비밀번호 해싱
- ✅ 입력 검증 (Zod 스키마)

---

## Usage

### Basic Authentication

```typescript
import { AuthService } from '@/auth/service'

const authService = new AuthService(userRepo, jwtSecret)

// 로그인
const result = await authService.authenticate('user@example.com', 'password123')

if (result.success) {
  console.log('Token:', result.token)
  console.log('Expires in:', result.expiresIn, 'seconds')
} else {
  console.error('Error:', result.error)
}
```

---

## API Reference

### `AuthService`

#### Constructor

```typescript
constructor(
  userRepo: UserRepository,
  jwtSecret: string
)
```

**Parameters**:
- `userRepo`: 사용자 저장소 인터페이스
- `jwtSecret`: JWT 서명에 사용할 비밀 키

#### Methods

##### `authenticate(email, password)`

사용자 인증 및 JWT 토큰 발급

**Parameters**:
- `email` (string): 사용자 이메일 (RFC 5322 형식)
- `password` (string): 비밀번호 (최소 8자)

**Returns**: `Promise<AuthResult>`

**Throws**:
- `Error('Invalid email format')`: 이메일 형식이 잘못됨
- `Error('Invalid credentials')`: 자격증명이 잘못됨

**Example**:
```typescript
const result = await authService.authenticate('user@example.com', 'password123')
```

---

## Data Models

### `AuthResult`

```typescript
interface AuthResult {
  success: boolean
  token?: string
  tokenType?: string  // 'Bearer'
  expiresIn?: number  // 900 (15 minutes)
  error?: string
}
```

### `User`

```typescript
interface User {
  id: string
  email: string
  passwordHash: string
}
```

---

## Security

### Password Hashing

- **Algorithm**: bcrypt
- **Cost Factor**: 12 (기본값)
- **Salt**: 자동 생성

### JWT Token

- **Algorithm**: HS256
- **Expiry**: 15분
- **Payload**: `{ userId: string, exp: number }`

---

## Testing

### Test Coverage

- **Overall**: 92%
- **Test Cases**: 4/4 passing

### Test Files

- `tests/auth/service.test.ts`

### Running Tests

```bash
# 모든 테스트 실행
bun test tests/auth/

# 커버리지 포함
bun test --coverage tests/auth/
```

---

## Related SPECs

- [SPEC-AUTH-002: 사용자 등록](/specs/SPEC-AUTH-002)
- [SPEC-AUTH-003: 비밀번호 재설정](/specs/SPEC-AUTH-003)

---

## Changelog

### v0.0.1 (2025-10-11)
- **INITIAL**: JWT 기반 인증 시스템 구현
- **AUTHOR**: @Goos

---

**Last Updated**: 2025-10-11
**TAG**: @DOC:AUTH-001
```

---

## PR 상태 업데이트 (Team Mode)

### 1. Draft → Ready 전환

```bash
# PR 상태 확인
gh pr view

# Draft → Ready 전환
gh pr ready

# 라벨 추가
gh pr edit --add-label "ready-for-review"
gh pr edit --add-label "tdd-complete"

# TRUST 점수 라벨
gh pr edit --add-label "trust-score-5"
```

**PR 업데이트 예시**:

```markdown
# PR #42: SPEC-AUTH-001: JWT 인증 시스템

## Summary
JWT 기반 사용자 인증 시스템 구현

## Changes
- ✅ SPEC 작성 완료
- ✅ TDD 구현 완료 (RED-GREEN-REFACTOR)
- ✅ Living Document 자동 생성
- ✅ TAG 체인 검증 완료

## TRUST Score: 5/5
- ✅ Test: 92% coverage
- ✅ Readable: 0 lint issues
- ✅ Unified: TypeScript strict
- ✅ Secured: 0 vulnerabilities
- ✅ Trackable: TAG chain intact

## Test Results
```bash
✓ 4 tests passing
✓ Coverage: 92%
```

## Files Changed
- `.moai/specs/SPEC-AUTH-001/spec.md`
- `tests/auth/service.test.ts`
- `src/auth/service.ts`
- `docs/features/auth/jwt-authentication.md`

## Next Steps
- [ ] Code review
- [ ] Merge to develop

---

🤖 Generated with [MoAI-ADK](https://github.com/modu-ai/moai-adk)
```

### 2. CI/CD 확인 (Team Mode + --auto-merge)

```bash
# CI/CD 상태 확인
gh pr checks --watch

# 출력 예시:
✓ Test (Node 18.x)
✓ Test (Node 20.x)
✓ Lint
✓ Type Check
✓ Security Scan
✓ Coverage

All checks have passed
```

### 3. 자동 머지 (Team Mode + --auto-merge)

```bash
# PR 머지 (squash)
gh pr merge --squash --delete-branch

# 머지 메시지
git commit -m "Merge pull request #42 from feature/SPEC-AUTH-001

SPEC-AUTH-001: JWT 인증 시스템

- TDD 구현 완료 (RED-GREEN-REFACTOR)
- TRUST 점수: 5/5
- 테스트 커버리지: 92%

🤖 Generated with MoAI-ADK"
```

### 4. develop 체크아웃

```bash
# develop 브랜치로 전환
git checkout develop

# 최신 변경사항 pull
git pull origin develop

# 다음 작업 준비 완료
echo "✅ Ready for next SPEC"
```

---

## TAG Chain Validation Rules

### Complete Chain ✅

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID (optional)
```

**예시**:
```bash
$ rg '@(SPEC|TEST|CODE):AUTH-001' -n

.moai/specs/SPEC-AUTH-001/spec.md:7:# @SPEC:AUTH-001
tests/auth/service.test.ts:1:// @TEST:AUTH-001
src/auth/service.ts:1:// @CODE:AUTH-001
```

### Broken Chain ❌

**Case 1: TEST 누락**
```
@SPEC:UPLOAD-003 → @CODE:UPLOAD-003 (TEST 없음)
```

**Case 2: CODE 누락**
```
@SPEC:PAYMENT-002 → @TEST:PAYMENT-002 (CODE 없음)
```

### Orphan TAG ⚠️

**Case 1: SPEC 없는 CODE**
```
@CODE:REFACTOR-010 (SPEC 없음)
```

**Case 2: SPEC 없는 TEST**
```
@TEST:BUGFIX-005 (SPEC 없음)
```

### TAG Validation Commands

```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 특정 도메인 TAG 조회
rg '@SPEC:AUTH' -n .moai/specs/

# 특정 TAG 체인 추적
rg '@SPEC:AUTH-001' -n .moai/specs/
rg '@TEST:AUTH-001' -n tests/
rg '@CODE:AUTH-001' -n src/
rg '@DOC:AUTH-001' -n docs/

# 고아 TAG 탐지 (스크립트)
for tag in $(rg '@CODE:([A-Z]+-[0-9]+)' -o -r '$1' -h src/ | sort -u); do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "❌ 고아 TAG: @CODE:$tag"
  fi
done
```

---

## Sync Report Structure

### Report Sections

| 섹션 | 내용 | 필수 |
|------|------|------|
| **Metadata** | 날짜, 브랜치, 작성자, 커밋 | ✅ |
| **TAG Chain Summary** | 완전/불완전/고아 TAG | ✅ |
| **Test Coverage** | 파일별 커버리지 | ✅ |
| **TRUST Compliance** | TRUST 5원칙 검증 결과 | ✅ |
| **Quality Metrics** | 코드 품질 지표 | ✅ |
| **Recommendations** | 다음 단계 제안 | ⚠️ |

### Report Format

```markdown
# Sync Report - [Date]

## Metadata
- Date: YYYY-MM-DDTHH:mm:ssZ
- Branch: feature/SPEC-XXX-YYY
- Author: @username
- Commit: abc1234

## TAG Chain Summary
### Complete Chains (N)
[List of complete chains]

### Incomplete Chains (N)
[List of incomplete chains]

### Orphan TAGs (N)
[List of orphan TAGs]

## Test Coverage
[Coverage table]

## TRUST Compliance
[TRUST verification results]

## Quality Metrics
[Code quality metrics]

## Recommendations
[Next steps]
```

---

## Auto-merge Strategy (Team Mode)

### Pre-merge Checklist

Alfred가 자동 머지 전 확인하는 항목:

- [ ] **TAG 체인 완전**: 모든 SPEC이 완전한 TAG 체인을 가짐
- [ ] **테스트 통과**: 모든 테스트가 통과
- [ ] **커버리지 충족**: ≥85% 커버리지
- [ ] **린터 통과**: 0 lint issues
- [ ] **타입 체크 통과**: 0 type errors
- [ ] **보안 스캔 통과**: 0 vulnerabilities
- [ ] **CI/CD 통과**: 모든 checks 성공
- [ ] **PR Ready**: Draft → Ready 전환됨

### Merge Methods

| Method | 설명 | 사용 시점 |
|--------|------|----------|
| **Squash** | 모든 커밋을 하나로 합침 | 기본값 (권장) |
| **Merge** | 머지 커밋 생성 | 히스토리 보존 필요 시 |
| **Rebase** | 커밋 재정렬 | 선형 히스토리 유지 시 |

**기본 설정**: Squash merge (TDD 히스토리 보존)

### Merge Message Template

```bash
Merge pull request #[PR_NUMBER] from [BRANCH]

[SPEC-ID]: [Title]

- TDD 구현 완료 (RED-GREEN-REFACTOR)
- TRUST 점수: [SCORE]/5
- 테스트 커버리지: [COVERAGE]%

Changes:
- [List of changes]

🤖 Generated with MoAI-ADK
```

---

## Best Practices

### 1. Sync Early, Sync Often

✅ **권장사항**:
```bash
# 매 SPEC 구현 후 즉시 동기화
/alfred:2-build AUTH-001
/alfred:3-sync  # 바로 실행

# 여러 SPEC을 누적하지 않기
❌ /alfred:2-build AUTH-001 AUTH-002 AUTH-003
   /alfred:3-sync  # 한 번에 동기화 (비권장)
```

### 2. Fix Broken Chains Immediately

✅ **권장사항**:
```bash
# TAG 체인이 끊어지면 즉시 수정
/alfred:3-sync --check  # 문제 확인
# 문제 해결 (TEST 또는 CODE 추가)
/alfred:3-sync  # 재검증
```

❌ **피해야 할 것**:
- 끊어진 TAG 체인을 그대로 두고 PR 머지
- 고아 TAG를 방치

### 3. Review Sync Reports

✅ **권장사항**:
```bash
# Sync Report 확인
cat .moai/reports/sync-report-2025-10-11.md

# 문제가 있으면 수정 후 재동기화
/alfred:3-sync
```

### 4. Use Auto-merge Carefully

✅ **권장사항** (Team 모드):
```bash
# CI/CD 설정이 완벽한 경우에만 사용
/alfred:3-sync --auto-merge

# 중요한 변경사항은 수동 리뷰 후 머지
/alfred:3-sync  # auto-merge 없이
gh pr merge --squash  # 수동 머지
```

---

## Common Pitfalls

### ❌ Pitfall 1: TAG 체인 검증 없이 머지

**잘못된 예**:
```bash
# TAG 체인 확인 없이 바로 머지
git add .
git commit -m "feature complete"
git push
gh pr merge
```

**올바른 예**:
```bash
# TAG 체인 검증 후 머지
/alfred:3-sync --check  # 먼저 검증
/alfred:3-sync  # 동기화
# 이후 머지
```

### ❌ Pitfall 2: 불완전한 TRUST 검증

**잘못된 예**:
```bash
# 테스트 커버리지 60%로 머지
$ pytest --cov
Coverage: 60%
$ gh pr merge  # 그냥 머지 (비권장)
```

**올바른 예**:
```bash
# 커버리지 충족 확인
$ pytest --cov
Coverage: 60%  # 85% 미만

# 테스트 추가
# 커버리지 재확인
$ pytest --cov
Coverage: 88%  # ✅

# 이후 동기화
/alfred:3-sync
```

### ❌ Pitfall 3: Sync Report 무시

**잘못된 예**:
```bash
/alfred:3-sync
# Sync Report 안 읽고 바로 머지
```

**올바른 예**:
```bash
/alfred:3-sync
# Sync Report 확인
cat .moai/reports/sync-report-*.md
# 문제 확인 후 머지 결정
```

---

## Troubleshooting

### Issue 1: TAG 체인 끊김

**증상**:
```bash
$ /alfred:3-sync

⚠️ 불완전한 TAG 체인 발견
- SPEC-UPLOAD-003: SPEC → CODE (TEST 누락)
```

**해결**:
```bash
# 1. 누락된 TEST 작성
# tests/upload/service.test.ts
// @TEST:UPLOAD-003 | SPEC: .moai/specs/SPEC-UPLOAD-003/spec.md

# 2. TAG 체인 재검증
/alfred:3-sync --check

# 3. 동기화 재실행
/alfred:3-sync
```

### Issue 2: 고아 TAG 발견

**증상**:
```bash
$ /alfred:3-sync

❌ 고아 TAG 발견
- @CODE:REFACTOR-010 (SPEC 없음)
```

**해결**:
```bash
# Option 1: SPEC 생성
/alfred:1-spec "REFACTOR-010: 기존 코드 리팩토링"

# Option 2: TAG 제거 (리팩토링이 불필요한 경우)
# src/some-file.ts에서 @CODE:REFACTOR-010 주석 제거
```

### Issue 3: TRUST 검증 실패

**증상**:
```bash
$ /alfred:3-sync

❌ TRUST 검증 실패
- Test: 커버리지 72% (목표 85%)
- Readable: 린터 오류 5개
```

**해결**:
```bash
# 1. 테스트 추가 (커버리지 향상)
# tests/에 누락된 테스트 케이스 추가

# 2. 린터 오류 수정
biome check src/ --apply

# 3. 재검증
/alfred:3-sync --check

# 4. 동기화 재실행
/alfred:3-sync
```

### Issue 4: CI/CD 실패 (Team 모드)

**증상**:
```bash
$ /alfred:3-sync --auto-merge

❌ CI/CD 검증 실패
- ✗ Test (Node 18.x): Failed
- ✓ Lint: Passed
```

**해결**:
```bash
# 1. 로컬에서 테스트 재실행
bun test

# 2. 문제 수정

# 3. 커밋 및 푸시
git add .
git commit -m "fix: test failure"
git push

# 4. CI/CD 재확인
gh pr checks --watch

# 5. 동기화 재실행
/alfred:3-sync --auto-merge
```

---

## Real-world Example: TODO App

### 시나리오: TODO 우선순위 기능 동기화

#### Step 1: TDD 구현 완료 확인

```bash
# 현재 브랜치 확인
$ git branch
* feature/SPEC-TODO-PRIORITY-001

# 테스트 통과 확인
$ bun test
✓ 3 tests passing

# 커버리지 확인
$ bun test --coverage
Coverage: 95%
```

#### Step 2: 동기화 실행

```bash
$ /alfred:3-sync

📋 문서 동기화 분석

검색된 TAG:
- @SPEC:TODO-PRIORITY-001 (1개)
- @TEST:TODO-PRIORITY-001 (1개)
- @CODE:TODO-PRIORITY-001 (1개)

TAG 체인 검증: ✅ 무결성 확인됨

TRUST 검증:
- ✅ Test: 95% coverage
- ✅ Readable: 0 lint issues
- ✅ Unified: TypeScript strict
- ✅ Secured: 0 vulnerabilities
- ✅ Trackable: TAG chain intact

TRUST 점수: 5/5 ✅

진행하시겠습니까? (진행/수정/중단)
```

**답변**: `진행`

#### Step 3: Living Document 생성

Alfred가 자동 생성:

1. **Sync Report**: `.moai/reports/sync-report-2025-10-11.md`
2. **Feature Doc**: `docs/features/todo/priority.md` (선택)

#### Step 4: PR 상태 업데이트 (Team 모드)

```bash
# PR 확인
$ gh pr view 43

#43 [Draft] SPEC-TODO-PRIORITY-001: TODO 우선순위 필드 추가

# Draft → Ready 전환
$ gh pr ready

# 라벨 추가
$ gh pr edit --add-label "ready-for-review"
$ gh pr edit --add-label "trust-score-5"

# PR 상태 재확인
$ gh pr view 43

#43 SPEC-TODO-PRIORITY-001: TODO 우선순위 필드 추가
  ✅ Ready for review
  ✅ All checks passed
```

#### Step 5: 자동 머지 (--auto-merge 옵션 사용)

```bash
$ /alfred:3-sync --auto-merge

# CI/CD 확인 중...
✓ Test (Node 18.x)
✓ Test (Node 20.x)
✓ Lint
✓ Type Check
✓ Security Scan
✓ Coverage

All checks passed ✅

# PR 자동 머지
$ gh pr merge --squash --delete-branch

Merged pull request #43
Deleted branch feature/SPEC-TODO-PRIORITY-001

# develop 체크아웃
$ git checkout develop
$ git pull origin develop

✅ 동기화 & 머지 완료
다음 작업 준비 완료!
```

#### Step 6: 최종 확인

```bash
# Sync Report 확인
$ cat .moai/reports/sync-report-2025-10-11.md

# TAG 체인 확인
$ rg '@(SPEC|TEST|CODE):TODO-PRIORITY-001' -n

.moai/specs/SPEC-TODO-PRIORITY-001/spec.md:7:# @SPEC:TODO-PRIORITY-001
tests/todo/model.test.py:2:# @TEST:TODO-PRIORITY-001
src/todo/model.py:1:# @CODE:TODO-PRIORITY-001

✅ TAG 체인 무결성 확인됨
```

---

## Mode Differences (Personal vs Team)

### Personal Mode

**특징**: 로컬 Git 워크플로우

```bash
/alfred:3-sync

# 수행 작업:
1. TAG 체인 검증
2. TRUST 검증
3. Living Document 생성
4. 로컬 머지 (develop/main으로)
5. 브랜치 정리 (선택)

# PR 생성 없음
```

### Team Mode

**특징**: GitHub PR 자동화

```bash
/alfred:3-sync --auto-merge

# 수행 작업:
1. TAG 체인 검증
2. TRUST 검증
3. Living Document 생성
4. PR 상태 Draft → Ready ✨
5. CI/CD 확인 ✨
6. PR 자동 머지 (squash) ✨
7. develop 체크아웃 ✨
8. 다음 작업 준비 완료 ✅
```

---

## Next Steps

문서 동기화가 완료되면:

1. **다음 SPEC 작성**: `/alfred:1-spec "새 기능"`
2. **SPEC-First TDD 반복**: 1-spec → 2-build → 3-sync
3. **프로젝트 릴리스**: 여러 SPEC 완료 후 버전 태그

### Related Guides

- **[Stage 1: SPEC Writing](/guides/workflow/1-spec)** - SPEC 작성 가이드
- **[Stage 2: TDD Implementation](/guides/workflow/2-build)** - TDD 구현 가이드
- **[TAG System](/guides/concepts/tag-system)** - TAG 시스템 상세
- **[TRUST Principles](/guides/concepts/trust-principles)** - 품질 원칙 상세

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>추적성 없이는 완성 없음</strong> 📖</p>
  <p>Living Document로 완벽한 추적성을 유지하세요!</p>
</div>
