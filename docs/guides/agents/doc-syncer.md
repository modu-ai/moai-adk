# doc-syncer: 테크니컬 라이터 에이전트

`doc-syncer`는 MoAI-ADK 생태계에서 Living Document 동기화와 TAG 체인 검증을 전담하는 전문 에이전트입니다.

## 개요

### 에이전트 정체성

- **아이콘**: 📖
- **직무**: 테크니컬 라이터 (Technical Writer)
- **전문 영역**: 문서-코드 동기화 및 API 문서화 전문가
- **페르소나**: Living Document 철학에 따라 코드와 문서의 완벽한 일치성을 보장하는 문서화 전문가
- **핵심 역할**: 실시간 문서-코드 동기화 및 @TAG 기반 완전한 추적성 문서 관리

### 전문가 특성

**사고 방식**:
- 코드 변경과 문서 갱신을 하나의 원자적 작업으로 처리
- CODE-FIRST 스캔 기반: 중간 캐시 없이 코드를 직접 스캔하여 TAG 추적성 보장
- 3단계 Phase 체계: 분석 → 실행 → 검증

**의사결정 기준**:
- 문서-코드 일치성 우선
- @TAG 무결성 보장
- 추적성 완전성 유지
- 프로젝트 유형별 조건부 문서화

**커뮤니케이션 스타일**:
- 동기화 범위와 영향도를 명확히 분석하여 보고
- 심각도별 아이콘 (❌ Critical / ⚠️ Warning / ℹ️ Info) 활용
- 구체적인 권장 조치 제시

---

## Living Document란?

### 개념

**Living Document**는 코드와 함께 살아 숨쉬는 문서입니다. 코드 변경 시 자동으로 동기화되며, 항상 최신 상태를 유지합니다.

### 핵심 가치

1. **실시간 동기화**: 코드 변경 즉시 문서 업데이트
2. **추적성 보장**: @TAG 시스템으로 코드-문서 간 완벽한 연결
3. **자동화**: 수동 문서 작성/관리 불필요
4. **신뢰성**: 항상 최신, 항상 정확

### 전통적 문서화 vs Living Document

| 전통적 문서화 | Living Document |
|-------------|----------------|
| 수동 작성 및 업데이트 | 자동 생성 및 동기화 |
| 코드와 문서 불일치 가능 | 코드-문서 일치 보장 |
| 문서 작성 부담 큰 | 문서 작성 부담 없음 |
| 오래된 문서 방치됨 | 항상 최신 상태 |
| 추적성 부족 | @TAG 완벽한 추적성 |

### Living Document의 이점

**개발자 관점**:
- 문서 작성 부담 제로
- 코드에 집중 가능
- TAG 주석만 추가하면 문서 자동 생성

**팀 관점**:
- 온보딩 시간 단축 (정확한 문서)
- 코드 리뷰 효율성 증가
- 지식 공유 용이

**프로젝트 관점**:
- 유지보수성 향상
- 기술 부채 감소
- 품질 일관성 유지

---

## 핵심 책임

### 1. Living Document 동기화

**코드 → 문서 동기화**:
- API 문서 자동 생성/갱신
- 함수/클래스 시그니처 추출
- README 및 아키텍처 문서 업데이트
- @CODE TAG 연결 확인

**문서 → 코드 동기화**:
- SPEC 변경 추적
- TAG 추적성 업데이트
- 끊어진 TAG 체인 복구

### 2. @TAG 체인 관리

**Primary Chain 관리**:
```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**TAG 체인 검증**:
- 완전한 체인 확인 (SPEC → TEST → CODE)
- 고아 TAG 탐지 (SPEC 없는 CODE/TEST)
- 끊어진 링크 감지 및 수정 제안
- 중복 TAG 병합/분리

### 3. 문서 품질 관리

**TRUST 원칙 검증**:
- **T** - Test First: 테스트 커버리지 ≥85%
- **R** - Readable: 린터 통과 (0 issues)
- **U** - Unified: 타입 체크 통과
- **S** - Secured: 보안 스캔 통과 (0 vulnerabilities)
- **T** - Trackable: TAG 체인 무결성 확인

**품질 메트릭 수집**:
- 코드 복잡도
- 파일/함수 크기
- 테스트 커버리지
- 보안 취약점

---

## 프로젝트 유형별 조건부 문서 생성

doc-syncer는 프로젝트 특성에 맞는 문서만 생성합니다.

### 매핑 규칙

| 프로젝트 유형 | 생성 문서 | 설명 |
|-------------|---------|------|
| **Web API** | `API.md`, `endpoints.md` | REST API, GraphQL 엔드포인트 문서화 |
| **CLI Tool** | `CLI_COMMANDS.md`, `usage.md` | 명령어 옵션, 사용법 문서화 |
| **Library** | `API_REFERENCE.md`, `modules.md` | 함수/클래스 API 문서화 |
| **Frontend** | `components.md`, `styling.md` | 컴포넌트, UI 패턴 문서화 |
| **Application** | `features.md`, `user-guide.md` | 기능 설명, 사용자 가이드 |

### 조건부 생성 원칙

**생성 조건**:
- 프로젝트에 해당 기능이 존재할 때만 생성
- 예: CLI 도구가 아니면 `CLI_COMMANDS.md` 생성 안 함

**자동 감지**:
doc-syncer가 자동으로 프로젝트 유형을 감지합니다:
```bash
# CLI 도구 감지
rg "#!/usr/bin/env" src/
rg "argparse|click|commander" src/

# Web API 감지
rg "@app.route|@api.route|app.get|router.get" src/

# Library 감지
package.json의 "main" 필드 확인
```

---

## 단일 책임 원칙

### doc-syncer 전담 영역

✅ **doc-syncer가 담당하는 작업**:
- Living Document 동기화 (코드 ↔ 문서)
- @TAG 시스템 검증 및 업데이트
- API 문서 자동 생성/갱신
- README 및 아키텍처 문서 동기화
- 문서-코드 일치성 검증
- Sync Report 생성

### git-manager에게 위임하는 작업

❌ **doc-syncer가 하지 않는 작업** (git-manager 전담):
- 모든 Git 커밋 작업 (add, commit, push)
- PR 상태 전환 (Draft → Ready)
- 리뷰어 자동 할당 및 라벨링
- GitHub CLI 연동 및 원격 동기화
- 브랜치 생성/삭제/머지

**중요**: doc-syncer는 git-manager를 직접 호출하지 않습니다. 모든 에이전트 간 조율은 Alfred가 담당합니다.

---

## 워크플로우: 2단계 Phase 시스템

### Phase 1: 분석 및 계획 수립 (2-3분)

Alfred가 현재 상태를 분석하고 동기화 계획을 수립합니다.

#### 1단계: Git 상태 확인

**실행 명령어**:
```bash
git status --short
git diff --stat
git branch
git rev-parse --abbrev-ref HEAD
```

**확인 항목**:
- **현재 브랜치**: `feature/SPEC-XXX` 형식인지
- **변경 파일**: staged/unstaged 파일 목록
- **PR 상태**: Draft/Ready 여부 (Team 모드)
- **변경 통계**: 추가/삭제 줄 수

**보고 예시**:
```markdown
📋 Git 상태 확인

현재 브랜치: feature/SPEC-AUTH-001
PR 상태: Draft
변경 파일: 4개
- .moai/specs/SPEC-AUTH-001/spec.md (신규)
- tests/auth/service.test.ts (신규)
- src/auth/service.ts (신규)
- src/auth/types.ts (수정)

변경 통계:
- 추가: +256 줄
- 삭제: -12 줄
```

#### 2단계: 코드 스캔 (CODE-FIRST)

**TAG 시스템 검증**:
```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# TAG 총 개수 확인
rg '@SPEC:' -c .moai/specs/ | awk '{sum+=$1} END {print "Total: " sum}'
rg '@TEST:' -c tests/ | awk '{sum+=$1} END {print "Total: " sum}'
rg '@CODE:' -c src/ | awk '{sum+=$1} END {print "Total: " sum}'
rg '@DOC:' -c docs/ | awk '{sum+=$1} END {print "Total: " sum}'
```

**Primary Chain 검증**:
```bash
# SPEC별 TAG 추출 및 체인 검증
for spec_id in $(rg '@SPEC:([A-Z]+-[0-9]+)' -o -r '$1' .moai/specs/ | sort -u); do
  echo "Checking $spec_id..."

  spec_count=$(rg -c "@SPEC:$spec_id" .moai/specs/ | awk '{sum+=$1} END {print sum}')
  test_count=$(rg -c "@TEST:$spec_id" tests/ | awk '{sum+=$1} END {print sum}')
  code_count=$(rg -c "@CODE:$spec_id" src/ | awk '{sum+=$1} END {print sum}')
  doc_count=$(rg -c "@DOC:$spec_id" docs/ | awk '{sum+=$1} END {print sum}')

  echo "  SPEC: $spec_count | TEST: $test_count | CODE: $code_count | DOC: $doc_count"
done
```

**고아 TAG 탐지**:
```bash
# SPEC 없는 CODE TAG 탐지
for tag in $(rg '@CODE:([A-Z]+-[0-9]+)' -o -r '$1' -h src/ | sort -u); do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "❌ 고아 TAG: @CODE:$tag (SPEC 없음)"
  fi
done

# SPEC 없는 TEST TAG 탐지
for tag in $(rg '@TEST:([A-Z]+-[0-9]+)' -o -r '$1' -h tests/ | sort -u); do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "❌ 고아 TAG: @TEST:$tag (SPEC 없음)"
  fi
done
```

**끊어진 링크 감지**:
```bash
# @DOC 폐기 TAG 탐지 (더 이상 사용하지 않음)
rg '@DOC:' -n docs/

# TODO/FIXME 미완성 작업 탐지
rg 'TODO|FIXME' -n src/ tests/
```

**스캔 결과 보고 예시**:
```markdown
📋 TAG 체인 스캔 결과

검색된 TAG:
- @SPEC:AUTH-001 (1개)
- @TEST:AUTH-001 (1개)
- @CODE:AUTH-001 (1개)
- @DOC:AUTH-001 (0개)

Primary Chain 검증:
✅ SPEC-AUTH-001: SPEC → TEST → CODE (완전)
⚠️  SPEC-UPLOAD-003: SPEC → CODE (TEST 누락)
⚠️  SPEC-PAYMENT-002: SPEC → TEST (CODE 누락)

고아 TAG:
❌ @CODE:REFACTOR-010 (SPEC 없음)

미완성 작업 (TODO/FIXME):
⚠️  src/auth/service.ts:45: TODO: Rate limiting 구현
⚠️  tests/upload/service.test.ts:12: FIXME: Edge case 테스트 추가
```

#### 3단계: 문서 현황 파악

**기존 문서 목록 확인**:
```bash
# docs/ 디렉토리 구조 확인
find docs/ -name "*.md" -type f

# 주요 문서 존재 여부
ls -la README.md CHANGELOG.md 2>/dev/null
```

**문서 현황 보고 예시**:
```markdown
📚 문서 현황

기존 문서:
- README.md (존재)
- CHANGELOG.md (존재)
- docs/features/auth/jwt-authentication.md (없음, 생성 필요)
- .moai/reports/sync-report-*.md (최근: 2025-10-10)

생성 예정 문서:
1. Sync Report: .moai/reports/sync-report-2025-10-11.md
2. Feature Doc: docs/features/auth/jwt-authentication.md (선택)
3. README 업데이트: 새 기능 섹션 추가
```

#### 4단계: TRUST 원칙 검증

**검증 명령어 실행**:

**T - Test First**:
```bash
# TypeScript/JavaScript
bun test --coverage
# 또는
vitest --coverage

# Python
pytest --cov=src --cov-report=term-missing

# Java
mvn test jacoco:report

# Go
go test -cover ./...

# Rust
cargo test --all-features
```

**R - Readable**:
```bash
# TypeScript/JavaScript
biome check src/
# 또는
eslint src/

# Python
ruff check src/

# Go
golangci-lint run

# Rust
cargo clippy
```

**U - Unified**:
```bash
# TypeScript
tsc --noEmit

# Python
mypy src/

# Go
go vet ./...
```

**S - Secured**:
```bash
# TypeScript/JavaScript
npm audit

# Python
bandit -r src/

# Rust
cargo audit
```

**T - Trackable**:
```bash
# TAG 무결성 확인
rg '@(SPEC|TEST|CODE):' -n
```

**검증 보고 예시**:
```markdown
✅ TRUST 검증 완료

### T - Test First
- ✅ 테스트 커버리지: 92% (목표 85% 초과)
- ✅ 모든 테스트 통과: 12/12

### R - Readable
- ✅ 린터 통과: 0 issues
- ✅ 파일 크기: 평균 156 LOC (≤300)
- ✅ 함수 크기: 평균 18 LOC (≤50)
- ✅ 복잡도: 최대 6 (≤10)

### U - Unified
- ✅ 타입 체크 통과
- ✅ 의존성 주입 패턴 사용

### S - Secured
- ✅ 보안 스캔: 0 vulnerabilities
- ✅ 입력 검증 구현 (Zod 스키마)

### T - Trackable
- ✅ TAG 체인 무결성 확인
- ✅ 고아 TAG 없음

**TRUST 점수**: 5/5 ✅
```

**TRUST 검증 실패 예시**:
```markdown
❌ TRUST 검증 실패

### T - Test First
- ❌ 테스트 커버리지: 72% (목표 85% 미만)
  → 권장 조치: 누락된 테스트 케이스 추가

### R - Readable
- ❌ 린터 오류: 5개
  → 권장 조치: biome check src/ --apply 실행

### U - Unified
- ✅ 타입 체크 통과

### S - Secured
- ⚠️  보안 경고: 1개 (낮은 심각도)
  → 권장 조치: 의존성 업데이트

### T - Trackable
- ⚠️  불완전한 TAG 체인: SPEC-UPLOAD-003 (TEST 누락)
  → 권장 조치: 테스트 작성 후 재검증

**TRUST 점수**: 2/5 ❌

**권장사항**:
1. 테스트 커버리지 85% 이상 달성
2. 린터 오류 수정 (biome check src/ --apply)
3. TAG 체인 완성 (TEST 작성)
4. 수정 후 /alfred:3-sync 재실행
```

#### 5단계: 사용자 확인 대기

**확인 프롬프트**:
```markdown
📋 Phase 1 완료: 분석 및 계획 수립

요약:
- Git 상태: feature/SPEC-AUTH-001 브랜치, 4개 파일 변경
- TAG 체인: 1개 완전, 2개 불완전, 1개 고아 TAG
- TRUST 점수: 5/5 ✅
- 생성 예정 문서: 2개

다음 단계:
1. Sync Report 생성
2. Feature Document 생성 (선택)
3. PR 상태 Draft → Ready 전환 (Team 모드)

진행하시겠습니까? (진행/수정/중단)
```

**사용자 응답 패턴**:
- **"진행"** 또는 **"시작"**: Phase 2로 진행
- **"수정 [내용]"**: 문제 해결 후 Phase 1 재실행
- **"중단"**: 작업 취소

---

### Phase 2: 문서 동기화 실행 (5-10분)

사용자가 "진행"하면 Alfred가 Living Document를 생성합니다.

#### 1. Sync Report 생성

**`.moai/reports/sync-report-YYYY-MM-DD.md`**:

doc-syncer가 다음 정보를 포함한 상세 리포트를 생성합니다:

**메타데이터 섹션**:
```markdown
# Sync Report - 2025-10-11T13:00:00Z

## Metadata
- **날짜**: 2025-10-11 13:00:00
- **브랜치**: feature/SPEC-AUTH-001
- **작성자**: @Goos
- **커밋**: a1b2c3d4
- **모드**: Team
```

**TAG 체인 요약**:
```markdown
## TAG Chain Summary

### Complete Chains (1)

#### SPEC-AUTH-001: JWT 인증 시스템
- ✅ **SPEC**: .moai/specs/SPEC-AUTH-001/spec.md (v0.0.1)
- ✅ **TEST**: tests/auth/service.test.ts (4 tests)
- ✅ **CODE**: src/auth/service.ts (156 LOC)
- ⚠️  **DOC**: not found (optional)

**Status**: Ready for review
**TRUST Score**: 5/5 ✅

---

### Incomplete Chains (2)

#### SPEC-UPLOAD-003: 파일 업로드 기능
- ✅ **SPEC**: .moai/specs/SPEC-UPLOAD-003/spec.md
- ❌ **TEST**: not found
- ✅ **CODE**: src/upload/service.ts

**Issue**: TEST 누락
**Recommendation**: tests/upload/service.test.ts 작성 필요

#### SPEC-PAYMENT-002: 결제 연동
- ✅ **SPEC**: .moai/specs/SPEC-PAYMENT-002/spec.md
- ✅ **TEST**: tests/payment/gateway.test.ts
- ❌ **CODE**: not found

**Issue**: CODE 누락
**Recommendation**: src/payment/gateway.ts 구현 필요

---

### Orphan TAGs (1)

#### @CODE:REFACTOR-010
- **Location**: src/utils/formatter.ts:1
- **Issue**: SPEC 없음
- **Recommendation**: SPEC 생성 또는 TAG 제거
```

**테스트 커버리지**:
```markdown
## Test Coverage

| File | Coverage | Lines | Missing |
|------|----------|-------|---------|
| src/auth/service.ts | 92% | 156 | 45, 52 |
| src/auth/types.ts | 100% | 24 | - |
| src/utils/jwt.ts | 88% | 64 | 12-15 |
| **Total** | **92%** | **244** | **6** |

**Status**: ✅ Passed (≥85%)

**Missing Coverage**:
- src/auth/service.ts:45 - Error handling for expired token
- src/auth/service.ts:52 - Edge case: empty password
- src/utils/jwt.ts:12-15 - Invalid signature handling
```

**TRUST 준수**:
```markdown
## TRUST Compliance

| Principle | Status | Details |
|-----------|--------|---------|
| **T** - Test First | ✅ | 92% coverage, 12/12 tests passing |
| **R** - Readable | ✅ | 0 lint issues, avg complexity 4.2 |
| **U** - Unified | ✅ | TypeScript strict mode, 0 type errors |
| **S** - Secured | ✅ | 0 vulnerabilities, input validation |
| **T** - Trackable | ⚠️ | 1 complete chain, 2 incomplete, 1 orphan |

**TRUST Score**: 4.5/5 ⚠️

**Issues**:
- Incomplete TAG chains: SPEC-UPLOAD-003, SPEC-PAYMENT-002
- Orphan TAG: @CODE:REFACTOR-010
```

**품질 메트릭**:
```markdown
## Quality Metrics

### Code Complexity
- **평균 복잡도**: 4.2 (목표 ≤10) ✅
- **최대 복잡도**: 6 (service.ts:authenticate)
- **함수 수**: 8

### Code Size
- **파일 크기**: 평균 156 LOC (목표 ≤300) ✅
- **함수 크기**: 평균 18 LOC (목표 ≤50) ✅
- **매개변수**: 최대 3개 (목표 ≤5) ✅

### Test Metrics
- **테스트 파일**: 1개
- **테스트 케이스**: 12개
- **실행 시간**: 1.2초
- **통과율**: 100% (12/12)
```

**권장사항**:
```markdown
## Recommendations

### ⚠️ Action Required

**불완전한 TAG 체인**:
1. SPEC-UPLOAD-003: TEST 작성 필요
   - `tests/upload/service.test.ts` 생성
   - 최소 85% 커버리지 달성

2. SPEC-PAYMENT-002: CODE 구현 필요
   - `src/payment/gateway.ts` 작성
   - TDD 사이클 (RED-GREEN-REFACTOR) 준수

**고아 TAG**:
3. @CODE:REFACTOR-010: SPEC 생성 또는 TAG 제거
   - Option 1: `/alfred:1-spec "REFACTOR-010: 포맷터 개선"`
   - Option 2: src/utils/formatter.ts:1의 TAG 주석 제거

### ✅ Ready to Merge (SPEC-AUTH-001)

**조건**:
- ✅ TAG 체인 완전
- ✅ TRUST 점수 5/5
- ✅ 모든 테스트 통과
- ✅ 커버리지 ≥85%

**Next Steps**:
1. Code review 요청
2. CI/CD 통과 확인 (Team 모드)
3. PR 머지 (squash)
4. develop 브랜치로 전환

---

**Generated by**: MoAI-ADK v0.2.17
**Command**: /alfred:3-sync
**Duration**: 8.3 seconds
```

#### 2. Feature Document 생성 (선택)

**프로젝트 유형별 문서 생성**:

doc-syncer는 프로젝트 유형을 자동 감지하여 적절한 문서만 생성합니다.

**Web API 프로젝트 예시**:

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

### Error Handling

```typescript
const result = await authService.authenticate(email, password)

if (!result.success) {
  switch (result.error) {
    case 'Invalid email format':
      // 이메일 형식 오류 처리
      break
    case 'Invalid credentials':
      // 인증 실패 처리
      break
    default:
      // 기타 오류 처리
  }
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

**Throws**:
- `Error('JWT secret is required')`: 비밀 키가 제공되지 않음

#### Methods

##### `authenticate(email, password)`

사용자 인증 및 JWT 토큰 발급

**Signature**:
```typescript
authenticate(email: string, password: string): Promise<AuthResult>
```

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

**Test Cases** (from @TEST:AUTH-001):
```typescript
// tests/auth/service.test.ts
✅ 유효한 자격증명으로 로그인 성공
✅ 잘못된 이메일 형식 거부
✅ 잘못된 비밀번호 거부
✅ 15분 만료 토큰 발급
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

**Fields**:
- `success`: 인증 성공 여부
- `token`: JWT 토큰 (성공 시)
- `tokenType`: 토큰 유형 (항상 'Bearer')
- `expiresIn`: 토큰 만료 시간 (초, 900 = 15분)
- `error`: 오류 메시지 (실패 시)

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
- **Salt**: 자동 생성 (bcrypt 내장)
- **해시 길이**: 60자

### JWT Token

- **Algorithm**: HS256 (HMAC-SHA256)
- **Expiry**: 15분 (900초)
- **Payload**: `{ userId: string, exp: number }`
- **Secret**: 환경변수 `JWT_SECRET`에서 로드

### Input Validation

```typescript
// Zod 스키마 사용
const AuthInputSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
})
```

---

## Testing

### Test Coverage

- **Overall**: 92%
- **Test Cases**: 12/12 passing
- **Execution Time**: 1.2 seconds

### Test Files

- `tests/auth/service.test.ts` - 메인 테스트 스위트

### Test Cases

| Test Case | Status | Coverage |
|-----------|--------|----------|
| 유효한 자격증명 로그인 | ✅ | L10-25 |
| 잘못된 이메일 형식 거부 | ✅ | L27-35 |
| 잘못된 비밀번호 거부 | ✅ | L37-45 |
| 15분 만료 토큰 발급 | ✅ | L47-55 |

### Running Tests

```bash
# 모든 테스트 실행
bun test tests/auth/

# 커버리지 포함
bun test --coverage tests/auth/

# 특정 테스트만 실행
bun test tests/auth/service.test.ts
```

---

## Performance

### Benchmarks

- **평균 인증 시간**: 120ms
- **bcrypt 해싱 시간**: 100ms (cost factor 12)
- **JWT 생성 시간**: 5ms

### Optimization Tips

- bcrypt cost factor 조정 (12 → 10, 50% 성능 향상)
- Redis 세션 캐싱 (선택)

---

## Related SPECs

- [SPEC-AUTH-002: 사용자 등록](/specs/SPEC-AUTH-002) - `depends_on`
- [SPEC-AUTH-003: 비밀번호 재설정](/specs/SPEC-AUTH-003) - `related_specs`
- [SPEC-USER-001: 사용자 프로필](/specs/SPEC-USER-001) - `related_specs`

---

## Changelog

### v0.0.1 (2025-10-11)
- **INITIAL**: JWT 기반 인증 시스템 구현
- **AUTHOR**: @Goos
- **COMMITS**:
  - `a1b2c3d`: 🔴 RED: 인증 테스트 작성
  - `e4f5g6h`: 🟢 GREEN: JWT 인증 구현
  - `i7j8k9l`: ♻️ REFACTOR: 입력 검증 추가

---

**Last Updated**: 2025-10-11
**TAG**: @DOC:AUTH-001
**Generated by**: MoAI-ADK v0.2.17
```

**CLI Tool 프로젝트 예시**:

**`docs/CLI_COMMANDS.md`**:

```markdown
<!-- @DOC:CLI-001 | SPEC: .moai/specs/SPEC-CLI-001/spec.md -->

# CLI Commands

> **SPEC**: CLI-001 | **Version**: 0.0.1

## Overview

MoAI-ADK CLI 명령어 레퍼런스

---

## Commands

### `moai init`

프로젝트 초기화

**Syntax**:
```bash
moai init [directory]
```

**Parameters**:
- `directory` (optional): 초기화할 디렉토리 (기본값: 현재 디렉토리)

**Examples**:
```bash
# 현재 디렉토리 초기화
moai init

# 특정 디렉토리 초기화
moai init ./my-project
```

**Exit Codes**:
- `0`: 성공
- `1`: 오류 (디렉토리 권한 없음 등)

---

### `moai doctor`

환경 진단

**Syntax**:
```bash
moai doctor [--fix]
```

**Options**:
- `--fix`: 발견된 문제 자동 수정

**Examples**:
```bash
# 진단만 수행
moai doctor

# 문제 자동 수정
moai doctor --fix
```

---

**Last Updated**: 2025-10-11
**TAG**: @DOC:CLI-001
```

#### 3. README 업데이트 (선택)

doc-syncer가 README.md의 특정 섹션을 자동 업데이트합니다.

**업데이트 대상 섹션**:
- `## Features` - 새 기능 추가
- `## Usage` - 사용법 예시 갱신
- `## Installation` - 설치 가이드 동기화

**예시**:

**Before**:
```markdown
## Features

- User authentication
- File upload
```

**After** (SPEC-AUTH-001 추가):
```markdown
## Features

- User authentication
  - ✨ **NEW**: JWT-based authentication (AUTH-001)
- File upload
```

#### 4. PR 상태 업데이트 (Team 모드)

**Draft → Ready 전환**:

```bash
# PR 상태 확인
gh pr view

# Draft → Ready 전환
gh pr ready

# 라벨 추가
gh pr edit --add-label "ready-for-review"
gh pr edit --add-label "tdd-complete"
gh pr edit --add-label "trust-score-5"
```

**PR 본문 업데이트**:

Alfred가 PR 본문에 TRUST 점수 및 동기화 결과를 추가합니다.

```markdown
# PR #42: SPEC-AUTH-001: JWT 인증 시스템

## Summary
JWT 기반 사용자 인증 시스템 구현

## Changes
- ✅ SPEC 작성 완료
- ✅ TDD 구현 완료 (RED-GREEN-REFACTOR)
- ✅ Living Document 자동 생성
- ✅ TAG 체인 검증 완료

## TRUST Score: 5/5 ✅

| Principle | Status | Details |
|-----------|--------|---------|
| **T** - Test First | ✅ | 92% coverage, 12/12 tests passing |
| **R** - Readable | ✅ | 0 lint issues, complexity ≤10 |
| **U** - Unified | ✅ | TypeScript strict mode |
| **S** - Secured | ✅ | 0 vulnerabilities |
| **T** - Trackable | ✅ | TAG chain intact |

## Test Results
```bash
✓ 12 tests passing
✓ Coverage: 92%
✓ Execution time: 1.2s
```

## Files Changed
- `.moai/specs/SPEC-AUTH-001/spec.md` (+156)
- `tests/auth/service.test.ts` (+89)
- `src/auth/service.ts` (+156)
- `src/auth/types.ts` (+24)
- `docs/features/auth/jwt-authentication.md` (+250)

## Living Documents
- 📄 [Sync Report](.moai/reports/sync-report-2025-10-11.md)
- 📖 [Feature Doc](docs/features/auth/jwt-authentication.md)

## Next Steps
- [ ] Code review
- [ ] Merge to develop

---

🤖 Generated with [MoAI-ADK](https://github.com/modu-ai/moai-adk)
```

#### 5. CI/CD 확인 및 자동 머지 (Team 모드 + --auto-merge)

**CI/CD 상태 확인**:

```bash
# CI/CD 상태 watch
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

**자동 머지 실행**:

```bash
# PR 머지 (squash)
gh pr merge --squash --delete-branch

# 머지 메시지
Merge pull request #42 from feature/SPEC-AUTH-001

SPEC-AUTH-001: JWT 인증 시스템

- TDD 구현 완료 (RED-GREEN-REFACTOR)
- TRUST 점수: 5/5
- 테스트 커버리지: 92%

Changes:
- JWT 기반 인증 서비스 구현
- bcrypt 비밀번호 해싱
- Zod 입력 검증

🤖 Generated with MoAI-ADK
```

**develop 브랜치로 전환**:

```bash
# develop 브랜치로 전환
git checkout develop

# 최신 변경사항 pull
git pull origin develop

# 다음 작업 준비 완료
echo "✅ Ready for next SPEC"
```

---

## TAG 체인 검증 규칙

### 완전한 TAG 체인 (Complete Chain) ✅

**구조**:
```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID (optional)
```

**검증 조건**:
- @SPEC:ID가 `.moai/specs/SPEC-{ID}/spec.md`에 존재
- @TEST:ID가 `tests/` 디렉토리에 존재
- @CODE:ID가 `src/` 디렉토리에 존재
- @DOC:ID는 선택 (없어도 완전한 체인으로 인정)

**예시**:

```bash
# TAG 검색
$ rg '@(SPEC|TEST|CODE):AUTH-001' -n

.moai/specs/SPEC-AUTH-001/spec.md:7:# @SPEC:AUTH-001
tests/auth/service.test.ts:1:// @TEST:AUTH-001
src/auth/service.ts:1:// @CODE:AUTH-001

✅ 완전한 TAG 체인
```

### 불완전한 TAG 체인 (Incomplete Chain) ⚠️

**Case 1: TEST 누락**:
```
@SPEC:UPLOAD-003 → @CODE:UPLOAD-003 (TEST 없음)
```

**권장 조치**:
```bash
# tests/upload/service.test.ts 생성
# @TEST:UPLOAD-003 | SPEC: .moai/specs/SPEC-UPLOAD-003/spec.md

describe('File Upload Service', () => {
  // 테스트 케이스 작성
})
```

**Case 2: CODE 누락**:
```
@SPEC:PAYMENT-002 → @TEST:PAYMENT-002 (CODE 없음)
```

**권장 조치**:
```bash
# src/payment/gateway.ts 생성
# @CODE:PAYMENT-002 | SPEC: .moai/specs/SPEC-PAYMENT-002/spec.md | TEST: tests/payment/gateway.test.ts

export class PaymentGateway {
  // 구현
}
```

**Case 3: TEST와 CODE 모두 누락**:
```
@SPEC:REFUND-005 (TEST/CODE 없음)
```

**권장 조치**:
```bash
# 1. TDD 구현 시작
/alfred:2-build REFUND-005
```

### 고아 TAG (Orphan TAG) ❌

**정의**: SPEC 없이 CODE 또는 TEST만 존재하는 TAG

**Case 1: SPEC 없는 CODE**:
```
@CODE:REFACTOR-010 (SPEC 없음)
```

**권장 조치**:

**Option 1: SPEC 생성**:
```bash
/alfred:1-spec "REFACTOR-010: 코드 리팩토링"
```

**Option 2: TAG 제거**:
```bash
# src/utils/formatter.ts에서 @CODE:REFACTOR-010 주석 제거
```

**Case 2: SPEC 없는 TEST**:
```
@TEST:BUGFIX-005 (SPEC 없음)
```

**권장 조치**:
```bash
# SPEC 생성 (버그 수정도 SPEC-First)
/alfred:1-spec "BUGFIX-005: 로그인 버그 수정"
```

### 중복 TAG ⚠️

**정의**: 동일한 TAG ID가 여러 곳에 존재

**검증 방법**:
```bash
# 중복 TAG 탐지
rg '@SPEC:AUTH-001' -c .moai/specs/ | awk '$1 > 1 {print "Duplicate: " FILENAME}'
```

**권장 조치**:
- 중복 TAG 중 하나 제거
- 다른 도메인/ID로 분리 (예: AUTH-001 → AUTH-001A, AUTH-001B)

### TAG 검증 명령어

**전체 TAG 스캔**:
```bash
# 모든 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

**특정 도메인 TAG 조회**:
```bash
# AUTH 도메인 TAG 조회
rg '@SPEC:AUTH' -n .moai/specs/
rg '@TEST:AUTH' -n tests/
rg '@CODE:AUTH' -n src/
```

**특정 TAG 체인 추적**:
```bash
# AUTH-001 TAG 체인 추적
rg '@SPEC:AUTH-001' -n .moai/specs/
rg '@TEST:AUTH-001' -n tests/
rg '@CODE:AUTH-001' -n src/
rg '@DOC:AUTH-001' -n docs/
```

**고아 TAG 자동 탐지 스크립트**:
```bash
#!/bin/bash
# detect-orphan-tags.sh

echo "🔍 고아 TAG 탐지 중..."

# CODE 고아 TAG
for tag in $(rg '@CODE:([A-Z]+-[0-9]+)' -o -r '$1' -h src/ | sort -u); do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "❌ 고아 TAG: @CODE:$tag (SPEC 없음)"
  fi
done

# TEST 고아 TAG
for tag in $(rg '@TEST:([A-Z]+-[0-9]+)' -o -r '$1' -h tests/ | sort -u); do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "❌ 고아 TAG: @TEST:$tag (SPEC 없음)"
  fi
done

echo "✅ 고아 TAG 탐지 완료"
```

**사용법**:
```bash
chmod +x detect-orphan-tags.sh
./detect-orphan-tags.sh
```

---

## 모드별 차이점 (Personal vs Team)

### Personal Mode

**특징**: 로컬 Git 워크플로우, PR 없음

**실행**:
```bash
/alfred:3-sync
```

**수행 작업**:
1. TAG 체인 검증
2. TRUST 검증
3. Living Document 생성
4. 로컬 머지 (develop/main으로)
5. 브랜치 정리 (선택)

**장점**:
- 빠른 개발 사이클
- PR 오버헤드 없음
- 소규모 프로젝트/개인 프로젝트에 적합

**단점**:
- 코드 리뷰 없음
- CI/CD 검증 없음

### Team Mode

**특징**: GitHub PR 자동화, CI/CD 통합

**실행**:
```bash
# 기본 동기화 (Draft → Ready)
/alfred:3-sync

# 자동 머지 (CI/CD 통과 후)
/alfred:3-sync --auto-merge
```

**수행 작업**:
1. TAG 체인 검증
2. TRUST 검증
3. Living Document 생성
4. **PR 상태 Draft → Ready** ✨
5. **CI/CD 확인** ✨ (--auto-merge 시)
6. **PR 자동 머지 (squash)** ✨ (--auto-merge 시)
7. **develop 체크아웃** ✨ (--auto-merge 시)
8. 다음 작업 준비 완료 ✅

**장점**:
- 코드 리뷰 프로세스
- CI/CD 자동 검증
- 팀 협업에 최적
- 완전 자동화된 GitFlow

**단점**:
- PR 생성/관리 오버헤드
- CI/CD 설정 필요

### 모드 설정

`.moai/config.json`:
```json
{
  "project": {
    "mode": "team"  // "personal" 또는 "team"
  }
}
```

---

## 실전 시나리오

### 시나리오 1: TODO App - 우선순위 기능 동기화

**상황**: TODO 항목에 우선순위 필드를 추가하는 기능을 TDD로 구현 완료

#### Step 1: TDD 구현 완료 확인

```bash
# 현재 브랜치 확인
$ git branch
* feature/SPEC-TODO-PRIORITY-001

# 테스트 통과 확인
$ bun test
✓ src/todo/model.test.ts (3 tests)

# 커버리지 확인
$ bun test --coverage
Coverage: 95%
```

#### Step 2: 동기화 실행

```bash
$ /alfred:3-sync

📋 Phase 1: 분석 및 계획 수립

Git 상태:
- 브랜치: feature/SPEC-TODO-PRIORITY-001
- 변경 파일: 3개
  - .moai/specs/SPEC-TODO-PRIORITY-001/spec.md (신규)
  - tests/todo/model.test.ts (신규)
  - src/todo/model.ts (수정)

TAG 체인 스캔:
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

생성 예정 문서:
1. Sync Report: .moai/reports/sync-report-2025-10-11.md
2. Feature Doc: docs/features/todo/priority.md (선택)

진행하시겠습니까? (진행/수정/중단)
```

**답변**: `진행`

#### Step 3: Living Document 생성

Alfred가 자동 생성:

1. **Sync Report**: `.moai/reports/sync-report-2025-10-11.md`
   - TAG 체인 요약
   - 테스트 커버리지
   - TRUST 준수 현황
   - 품질 메트릭

2. **Feature Doc**: `docs/features/todo/priority.md`
   - 우선순위 필드 사용법
   - API 레퍼런스
   - 데이터 모델
   - 테스트 결과

#### Step 4: PR 상태 업데이트 (Team 모드)

```bash
# PR 확인
$ gh pr view 43

#43 [Draft] SPEC-TODO-PRIORITY-001: TODO 우선순위 필드 추가
  ⏳ Checks in progress

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

#### Step 5: 자동 머지 (--auto-merge 옵션)

```bash
$ /alfred:3-sync --auto-merge

📋 CI/CD 확인 중...

✓ Test (Node 18.x)
✓ Test (Node 20.x)
✓ Lint
✓ Type Check
✓ Security Scan
✓ Coverage

All checks passed ✅

📋 PR 자동 머지 중...

$ gh pr merge --squash --delete-branch

✅ Merged pull request #43
✅ Deleted branch feature/SPEC-TODO-PRIORITY-001

📋 develop 브랜치로 전환 중...

$ git checkout develop
$ git pull origin develop

✅ 동기화 & 머지 완료!
다음 작업 준비 완료!
```

#### Step 6: 최종 확인

```bash
# Sync Report 확인
$ cat .moai/reports/sync-report-2025-10-11.md

# TAG 체인 확인
$ rg '@(SPEC|TEST|CODE):TODO-PRIORITY-001' -n

.moai/specs/SPEC-TODO-PRIORITY-001/spec.md:7:# @SPEC:TODO-PRIORITY-001
tests/todo/model.test.ts:2:// @TEST:TODO-PRIORITY-001
src/todo/model.ts:1:// @CODE:TODO-PRIORITY-001

✅ TAG 체인 무결성 확인됨
```

---

### 시나리오 2: E-commerce - 결제 기능 TAG 체인 복구

**상황**: SPEC-PAYMENT-001은 있는데 TEST가 누락됨 (불완전한 TAG 체인)

#### Step 1: 문제 발견

```bash
$ /alfred:3-sync

⚠️  불완전한 TAG 체인 발견

SPEC-PAYMENT-001: 결제 게이트웨이 연동
- ✅ SPEC: .moai/specs/SPEC-PAYMENT-001/spec.md
- ❌ TEST: not found
- ✅ CODE: src/payment/gateway.ts

권장 조치: tests/payment/gateway.test.ts 작성 필요

진행하시겠습니까? (진행/수정/중단)
```

**답변**: `중단`

#### Step 2: 누락된 TEST 작성

```typescript
// tests/payment/gateway.test.ts
// @TEST:PAYMENT-001 | SPEC: .moai/specs/SPEC-PAYMENT-001/spec.md

import { describe, it, expect } from 'vitest'
import { PaymentGateway } from '@/payment/gateway'

describe('PaymentGateway', () => {
  it('should process payment successfully', async () => {
    // 테스트 구현
  })

  it('should handle payment failure', async () => {
    // 테스트 구현
  })
})
```

#### Step 3: TAG 체인 재검증

```bash
$ /alfred:3-sync --check

✅ TAG 체인 검증 완료

SPEC-PAYMENT-001: 결제 게이트웨이 연동
- ✅ SPEC: .moai/specs/SPEC-PAYMENT-001/spec.md
- ✅ TEST: tests/payment/gateway.test.ts
- ✅ CODE: src/payment/gateway.ts

TAG 체인 무결성 확인됨 ✅

진행하시겠습니까? (진행/수정/중단)
```

**답변**: `진행`

#### Step 4: 동기화 완료

Alfred가 Sync Report 및 Living Document 생성

---

### 시나리오 3: Mobile App - 고아 TAG 정리

**상황**: SPEC 없는 CODE TAG 발견 (@CODE:REFACTOR-010)

#### Step 1: 고아 TAG 발견

```bash
$ /alfred:3-sync

❌ 고아 TAG 발견

@CODE:REFACTOR-010
- Location: src/utils/formatter.ts:1
- Issue: SPEC 없음

권장 조치:
1. SPEC 생성: /alfred:1-spec "REFACTOR-010: 포맷터 개선"
2. TAG 제거: src/utils/formatter.ts:1의 TAG 주석 제거

진행하시겠습니까? (진행/수정/중단)
```

**답변**: `중단`

#### Step 2: SPEC 생성 결정

**Option 1: SPEC 생성 (리팩토링이 중요한 경우)**:
```bash
$ /alfred:1-spec "REFACTOR-010: 날짜 포맷터 성능 개선"

# SPEC 작성 후 TAG 체인 완성
```

**Option 2: TAG 제거 (사소한 리팩토링인 경우)**:
```typescript
// src/utils/formatter.ts
// Before:
// @CODE:REFACTOR-010 | SPEC: ???

// After:
// (TAG 주석 제거)

export function formatDate(date: Date): string {
  // 구현
}
```

#### Step 3: TAG 체인 재검증

```bash
$ /alfred:3-sync --check

✅ TAG 체인 검증 완료

고아 TAG: 없음

진행하시겠습니까? (진행/수정/중단)
```

**답변**: `진행`

---

## Best Practices (모범 사례)

### 1. Sync Early, Sync Often

✅ **권장사항**:
```bash
# 매 SPEC 구현 후 즉시 동기화
/alfred:2-build AUTH-001
/alfred:3-sync  # 바로 실행

# 작은 단위로 자주 동기화
```

❌ **피해야 할 것**:
```bash
# 여러 SPEC을 누적하지 않기
/alfred:2-build AUTH-001 AUTH-002 AUTH-003
/alfred:3-sync  # 한 번에 동기화 (비권장)
```

**이유**:
- 문제 발견 시 즉시 수정 가능
- TAG 체인 끊김 방지
- Sync Report 가독성 향상

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
- "나중에 수정하겠다"는 TODO 주석만 남기기

**이유**:
- 추적성 무결성 유지
- 기술 부채 방지
- 팀원 혼란 방지

### 3. Review Sync Reports

✅ **권장사항**:
```bash
# Sync Report 확인
cat .moai/reports/sync-report-2025-10-11.md

# 문제가 있으면 수정 후 재동기화
/alfred:3-sync
```

❌ **피해야 할 것**:
- Sync Report를 읽지 않고 바로 머지
- 경고 메시지 무시
- TRUST 점수만 확인하고 세부사항 무시

**이유**:
- 잠재적 문제 조기 발견
- 품질 메트릭 추적
- 개선 기회 파악

### 4. Use Auto-merge Carefully

✅ **권장사항** (Team 모드):
```bash
# CI/CD 설정이 완벽한 경우에만 사용
/alfred:3-sync --auto-merge

# 중요한 변경사항은 수동 리뷰 후 머지
/alfred:3-sync  # auto-merge 없이
# 코드 리뷰 후
gh pr merge --squash  # 수동 머지
```

❌ **피해야 할 것**:
- CI/CD 설정 없이 --auto-merge 사용
- 모든 PR을 자동 머지
- 중요한 보안 패치를 리뷰 없이 자동 머지

**이유**:
- 코드 품질 유지
- 팀 리뷰 문화 유지
- 중요한 변경사항은 인간 검증 필요

### 5. Keep Living Documents Updated

✅ **권장사항**:
```bash
# TAG 주석 일관성 유지
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.ts

# SPEC 버전 업데이트 시 Living Document 재생성
/alfred:3-sync
```

❌ **피해야 할 것**:
- TAG 주석 형식 변경
- 수동으로 Feature Doc 편집 (자동 생성된 부분)
- Sync Report 수동 수정

**이유**:
- TAG 파싱 일관성 유지
- Living Document 신뢰성 유지
- 자동화 이점 활용

---

## Common Pitfalls (흔한 실수)

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
# 문제 수정
/alfred:3-sync  # 동기화
# 이후 머지
```

**결과**:
- ❌ 잘못된 예: 추적성 무결성 깨짐, 나중에 고아 TAG 발생
- ✅ 올바른 예: 완벽한 TAG 체인 유지, 추적성 보장

---

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
# (누락된 테스트 케이스 작성)

# 커버리지 재확인
$ pytest --cov
Coverage: 88%  # ✅

# 이후 동기화
/alfred:3-sync
```

**결과**:
- ❌ 잘못된 예: 품질 기준 미달, 기술 부채 누적
- ✅ 올바른 예: 품질 기준 충족, 유지보수성 향상

---

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

**결과**:
- ❌ 잘못된 예: 잠재적 문제 놓침, 품질 메트릭 무시
- ✅ 올바른 예: 문제 조기 발견, 품질 개선 기회 확보

---

### ❌ Pitfall 4: 여러 SPEC을 한 번에 동기화

**잘못된 예**:
```bash
# 3개 SPEC을 구현 후 한 번에 동기화
/alfred:2-build AUTH-001 AUTH-002 AUTH-003
/alfred:3-sync
```

**올바른 예**:
```bash
# 1개 SPEC씩 동기화
/alfred:2-build AUTH-001
/alfred:3-sync

/alfred:2-build AUTH-002
/alfred:3-sync

/alfred:2-build AUTH-003
/alfred:3-sync
```

**결과**:
- ❌ 잘못된 예: Sync Report 복잡, 문제 발견 어려움
- ✅ 올바른 예: 명확한 Sync Report, 문제 즉시 수정

---

### ❌ Pitfall 5: Personal 모드에서 --auto-merge 사용

**잘못된 예**:
```bash
# Personal 모드에서 --auto-merge
/alfred:3-sync --auto-merge
```

**올바른 예**:
```bash
# Personal 모드에서는 --auto-merge 불필요
/alfred:3-sync
```

**결과**:
- ❌ 잘못된 예: 옵션 무시됨 (PR이 없으므로)
- ✅ 올바른 예: 로컬 머지만 수행

---

## Troubleshooting (문제 해결)

### Issue 1: TAG 체인 끊김

**증상**:
```bash
$ /alfred:3-sync

⚠️ 불완전한 TAG 체인 발견
- SPEC-UPLOAD-003: SPEC → CODE (TEST 누락)
```

**원인**:
- TDD 사이클을 건너뛰고 CODE만 작성
- TEST 파일을 삭제했지만 TAG는 남아있음

**해결**:
```bash
# 1. 누락된 TEST 작성
# tests/upload/service.test.ts
// @TEST:UPLOAD-003 | SPEC: .moai/specs/SPEC-UPLOAD-003/spec.md

describe('File Upload Service', () => {
  // 테스트 케이스 작성
})

# 2. TAG 체인 재검증
/alfred:3-sync --check

# 3. 동기화 재실행
/alfred:3-sync
```

---

### Issue 2: 고아 TAG 발견

**증상**:
```bash
$ /alfred:3-sync

❌ 고아 TAG 발견
- @CODE:REFACTOR-010 (SPEC 없음)
```

**원인**:
- SPEC 없이 CODE를 작성
- SPEC 파일을 삭제했지만 CODE의 TAG는 남아있음

**해결**:

**Option 1: SPEC 생성**:
```bash
/alfred:1-spec "REFACTOR-010: 기존 코드 리팩토링"
```

**Option 2: TAG 제거**:
```bash
# src/some-file.ts에서 @CODE:REFACTOR-010 주석 제거
```

**선택 기준**:
- 중요한 변경사항 → SPEC 생성
- 사소한 리팩토링 → TAG 제거

---

### Issue 3: TRUST 검증 실패

**증상**:
```bash
$ /alfred:3-sync

❌ TRUST 검증 실패
- Test: 커버리지 72% (목표 85%)
- Readable: 린터 오류 5개
```

**원인**:
- 테스트 케이스 누락
- 린터 규칙 위반

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

---

### Issue 4: CI/CD 실패 (Team 모드)

**증상**:
```bash
$ /alfred:3-sync --auto-merge

❌ CI/CD 검증 실패
- ✗ Test (Node 18.x): Failed
- ✓ Lint: Passed
```

**원인**:
- 특정 Node 버전에서 테스트 실패
- 환경 변수 누락
- 의존성 버전 불일치

**해결**:
```bash
# 1. 로컬에서 해당 Node 버전으로 테스트
nvm use 18
bun test

# 2. 문제 수정

# 3. 커밋 및 푸시
git add .
git commit -m "fix: Node 18 호환성 문제 수정"
git push

# 4. CI/CD 재확인
gh pr checks --watch

# 5. 동기화 재실행
/alfred:3-sync --auto-merge
```

---

### Issue 5: Feature Document 생성 실패

**증상**:
```bash
$ /alfred:3-sync

⚠️ Feature Document 생성 실패
- docs/features/auth/jwt-authentication.md: 권한 거부
```

**원인**:
- docs/ 디렉토리 권한 문제
- 파일 시스템 쓰기 권한 없음

**해결**:
```bash
# 1. 디렉토리 권한 확인
ls -la docs/features/

# 2. 권한 수정
chmod -R 755 docs/

# 3. 동기화 재실행
/alfred:3-sync
```

---

### Issue 6: PR 상태 전환 실패 (Team 모드)

**증상**:
```bash
$ /alfred:3-sync

❌ PR 상태 전환 실패
- gh pr ready: PR not found
```

**원인**:
- PR이 존재하지 않음
- GitHub CLI 인증 문제

**해결**:
```bash
# 1. PR 존재 여부 확인
gh pr list

# 2. GitHub CLI 인증 확인
gh auth status

# 3. 인증 재설정 (필요 시)
gh auth login

# 4. 동기화 재실행
/alfred:3-sync
```

---

## 관련 문서

### 워크플로우 가이드

- **[Stage 1: SPEC Writing](/guides/workflow/1-spec)** - SPEC 작성 가이드
- **[Stage 2: TDD Implementation](/guides/workflow/2-build)** - TDD 구현 가이드

### 개념 가이드

- **[TAG System](/guides/concepts/tag-system)** - TAG 시스템 상세
- **[TRUST Principles](/guides/concepts/trust-principles)** - 품질 원칙 상세
- **[SPEC-First TDD](/guides/concepts/spec-first-tdd)** - 개발 방법론

### 에이전트 가이드

- **[Alfred Agent Ecosystem](/guides/agents/overview)** - 에이전트 생태계 개요
- **[spec-builder](/guides/agents/spec-builder)** - SPEC 작성 에이전트
- **[code-builder](/guides/agents/code-builder)** - TDD 구현 에이전트

---

## 마무리

### doc-syncer의 핵심 가치

1. **Living Document 자동화**
   - 코드 변경 즉시 문서 동기화
   - 수동 문서 작성 부담 제로
   - 항상 최신, 항상 정확

2. **완벽한 추적성**
   - TAG 체인 무결성 보장
   - 고아 TAG 자동 탐지
   - SPEC → TEST → CODE → DOC 완전한 연결

3. **품질 자동 검증**
   - TRUST 5원칙 자동 검증
   - 코드 품질 메트릭 수집
   - 문제 발견 즉시 알림

4. **GitFlow 완전 자동화**
   - Draft → Ready 자동 전환
   - CI/CD 확인 및 자동 머지
   - 다음 작업 즉시 준비

### 다음 단계

문서 동기화가 완료되면:

1. **다음 SPEC 작성**: `/alfred:1-spec "새 기능"`
2. **SPEC-First TDD 반복**: 1-spec → 2-build → 3-sync
3. **프로젝트 릴리스**: 여러 SPEC 완료 후 버전 태그

---

<div style="text-align: center; margin-top: 40px;">
  <p><strong>"추적성 없이는 완성 없음"</strong> 📖</p>
  <p>Living Document로 완벽한 추적성을 유지하세요!</p>
  <p><em>doc-syncer가 당신의 문서를 살아있게 합니다.</em></p>
</div>
