# TAG 시스템 가이드

<!-- @CODE:DOCS-002 | SPEC: .moai/specs/SPEC-DOCS-002/spec.md -->

> "코드의 진실은 코드 자체에만 존재한다 - CODE-FIRST 원칙"

## 개요

**@TAG 시스템**은 MoAI-ADK의 코드 추적성을 보장하는 핵심 메커니즘입니다. SPEC 문서부터 테스트, 구현 코드, 문서까지 전체 라이프사이클을 연결하여 완벽한 추적성을 제공합니다.

### TAG 시스템의 철학

- **CODE-FIRST**: TAG의 진실은 코드 자체에만 존재 (중간 캐시 없음)
- **TDD 정렬**: RED (테스트) → GREEN (구현) → REFACTOR (문서)
- **단순성**: 4개 TAG로 전체 라이프사이클 관리
- **실시간 검증**: `rg` 명령어로 즉시 TAG 체인 검증

---

## TAG 라이프사이클

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

| TAG | 역할 | TDD 단계 | 위치 | 필수 |
|-----|------|---------|------|-----|
| `@SPEC:ID` | 요구사항 명세 (EARS) | 사전 준비 | `.moai/specs/` | ✅ |
| `@TEST:ID` | 테스트 케이스 | RED | `tests/` | ✅ |
| `@CODE:ID` | 구현 코드 | GREEN + REFACTOR | `src/` | ✅ |
| `@DOC:ID` | 문서화 | REFACTOR | `docs/` | ⚠️ |

---

## TAG ID 규칙

### 기본 형식

```
@TAG:<DOMAIN>-<3자리 숫자>
```

**예시**:

- `@SPEC:AUTH-001` - 인증 도메인, 첫 번째 SPEC
- `@TEST:UPLOAD-003` - 업로드 도메인, 세 번째 테스트
- `@CODE:PAYMENT-042` - 결제 도메인, 42번째 코드

### TAG ID 불변성

- **TAG ID는 영구 불변**: 한번 부여된 ID는 절대 변경 금지
- **TAG 내용은 자유롭게 수정 가능**: HISTORY 섹션에 변경 이력 기록

**예시**:

```markdown
# ❌ 잘못된 사용: TAG ID 변경
@SPEC:AUTH-001 → @SPEC:AUTH-002  # 금지!

# ✅ 올바른 사용: 내용 수정
@SPEC:AUTH-001 (v0.0.1 → v0.1.0)  # 내용 업데이트, ID 유지
```

### 디렉토리 명명 규칙

SPEC 디렉토리는 반드시 `SPEC-{ID}` 형식을 따라야 합니다:

```bash
# ✅ 올바른 예
.moai/specs/SPEC-AUTH-001/
.moai/specs/SPEC-UPLOAD-042/
.moai/specs/SPEC-REFACTOR-001/
.moai/specs/SPEC-UPDATE-REFACTOR-001/  # 복합 도메인

# ❌ 잘못된 예
.moai/specs/AUTH-001/              # SPEC- 접두사 누락
.moai/specs/SPEC-001-auth/         # 순서 잘못됨
.moai/specs/SPEC-AUTH-001-jwt/     # 추가 설명 불필요
```

---

## TAG 유형별 사용법

### 1. @SPEC:ID - SPEC 문서

**위치**: `.moai/specs/SPEC-{ID}/spec.md`

**템플릿**:

```markdown
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-11
updated: 2025-10-11
author: @Goos
priority: high
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v0.0.1 (2025-10-11)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
- **AUTHOR**: @Goos
```

### 2. @TEST:ID - 테스트 코드

**위치**: `tests/` 디렉토리

**Python 예시**:

```python
# @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

import pytest
from auth.service import AuthService

def test_should_authenticate_valid_user():
    """유효한 사용자 인증 테스트"""
    auth = AuthService()
    result = auth.authenticate("user@example.com", "password123")
    assert result.success is True
```

**TypeScript 예시**:

```typescript
// @TEST:UPLOAD-001 | SPEC: .moai/specs/SPEC-UPLOAD-001/spec.md

import { describe, it, expect } from 'vitest';
import { UploadService } from '../src/upload/service';

describe('UploadService', () => {
  it('should upload valid file', async () => {
    const service = new UploadService();
    const result = await service.upload(validFile);
    expect(result.success).toBe(true);
  });
});
```

### 3. @CODE:ID - 구현 코드

**위치**: `src/` 디렉토리

**Python 예시**:

```python
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.py

class AuthService:
    """사용자 인증 서비스"""

    def authenticate(self, username: str, password: str) -> AuthResult:
        # 구현 코드
        pass
```

**TypeScript 예시**:

```typescript
// @CODE:UPLOAD-001 | SPEC: .moai/specs/SPEC-UPLOAD-001/spec.md | TEST: tests/upload/service.test.ts

export class UploadService {
  async upload(file: File): Promise<UploadResult> {
    // 구현 코드
  }
}
```

### 4. @DOC:ID - 문서

**위치**: `docs/` 디렉토리

**예시**:

```markdown
<!-- @DOC:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md -->

# JWT 인증 시스템 가이드

이 문서는 @SPEC:AUTH-001에 정의된 JWT 인증 시스템의 사용 방법을 설명합니다.
```

---

## TAG 서브 카테고리

구현 세부사항은 `@CODE:ID` 내부에 주석으로 표기합니다.

### 주요 서브 카테고리

| 서브 카테고리 | 용도 | 예시 |
|--------------|------|------|
| `@CODE:ID:API` | REST API, GraphQL 엔드포인트 | `@CODE:AUTH-001:API` |
| `@CODE:ID:UI` | 컴포넌트, 뷰, 화면 | `@CODE:UPLOAD-001:UI` |
| `@CODE:ID:DATA` | 데이터 모델, 스키마, 타입 | `@CODE:PAYMENT-001:DATA` |
| `@CODE:ID:DOMAIN` | 비즈니스 로직, 도메인 규칙 | `@CODE:AUTH-001:DOMAIN` |
| `@CODE:ID:INFRA` | 인프라, 데이터베이스, 외부 연동 | `@CODE:UPLOAD-001:INFRA` |

### 실제 코드 예시

**Python (FastAPI)**:

```python
# @CODE:AUTH-001:API - REST API 엔드포인트
from fastapi import APIRouter

router = APIRouter()

@router.post("/auth/login")
def login(credentials: LoginRequest):
    # API 구현
    pass

# @CODE:AUTH-001:DOMAIN - 비즈니스 로직
class AuthService:
    def authenticate(self, username: str, password: str):
        # 도메인 로직
        pass

# @CODE:AUTH-001:DATA - 데이터 모델
from pydantic import BaseModel

class User(BaseModel):
    id: int
    username: str
    email: str
```

**TypeScript (React)**:

```typescript
// @CODE:UPLOAD-001:UI - 파일 업로드 컴포넌트
export function FileUploader() {
  return <input type="file" onChange={handleUpload} />;
}

// @CODE:UPLOAD-001:DOMAIN - 비즈니스 로직
export class UploadService {
  async upload(file: File): Promise<UploadResult> {
    // 업로드 로직
  }
}

// @CODE:UPLOAD-001:DATA - 데이터 모델
export interface UploadResult {
  success: boolean;
  url: string | null;
}
```

---

## TAG 검증 방법

### 1. 중복 확인

```bash
# 특정 도메인 TAG 조회
rg "@SPEC:AUTH" -n .moai/specs/

# 특정 ID 전체 검색
rg "AUTH-001" -n

# 결과 예시:
# .moai/specs/SPEC-AUTH-001/spec.md:26:# @SPEC:AUTH-001: JWT 인증 시스템
# tests/auth/service.test.py:1:# @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md
# src/auth/service.py:1:# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md
```

### 2. TAG 체인 검증

```bash
# 전체 TAG 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 특정 TAG 체인 추적
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC 확인
rg '@TEST:AUTH-001' -n tests/        # TEST 확인
rg '@CODE:AUTH-001' -n src/          # CODE 확인
rg '@DOC:AUTH-001' -n docs/          # DOC 확인
```

### 3. 고아 TAG 탐지

**고아 TAG**: SPEC 없이 존재하는 CODE/TEST

```bash
# CODE는 있는데 SPEC이 없는 경우
rg '@CODE:AUTH-002' -n src/          # CODE 발견
rg '@SPEC:AUTH-002' -n .moai/specs/  # SPEC 없음 → 고아!

# 자동 스크립트 (.moai/scripts/check-orphan-tags.sh)
#!/bin/bash
for tag in $(rg '@CODE:' -o -h | sort -u); do
  id=$(echo $tag | sed 's/@CODE://')
  if ! rg -q "@SPEC:$id" .moai/specs/; then
    echo "고아 TAG 발견: @CODE:$id"
  fi
done
```

---

## 실전 예시

### 예시 1: JWT 인증 시스템 (AUTH-001)

#### SPEC 문서

```markdown
# .moai/specs/SPEC-AUTH-001/spec.md
---
id: AUTH-001
version: 0.0.1
---

# @SPEC:AUTH-001: JWT 인증 시스템

## Requirements
- 시스템은 JWT 기반 인증 기능을 제공해야 한다
- WHEN 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
```

#### 테스트 코드

```python
# tests/auth/service.test.py
# @TEST:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md

def test_should_issue_jwt_on_valid_login():
    auth = AuthService()
    result = auth.authenticate("user@example.com", "password123")
    assert result.token is not None
    assert result.token_type == "Bearer"
```

#### 구현 코드

```python
# src/auth/service.py
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md | TEST: tests/auth/service.test.py

import jwt
from datetime import datetime, timedelta

class AuthService:
    """@CODE:AUTH-001:DOMAIN - 인증 비즈니스 로직"""

    def authenticate(self, username: str, password: str) -> AuthResult:
        # 구현
        token = self._generate_jwt(user_id)
        return AuthResult(token=token, token_type="Bearer")

    def _generate_jwt(self, user_id: int) -> str:
        """@CODE:AUTH-001:DOMAIN - JWT 토큰 생성"""
        payload = {
            "sub": user_id,
            "exp": datetime.utcnow() + timedelta(minutes=15)
        }
        return jwt.encode(payload, "secret", algorithm="HS256")
```

### 예시 2: 파일 업로드 기능 (UPLOAD-001)

#### TAG 체인 전체

```
@SPEC:UPLOAD-001
  ↓
@TEST:UPLOAD-001
  ↓
@CODE:UPLOAD-001:UI      (React 컴포넌트)
@CODE:UPLOAD-001:DOMAIN  (업로드 서비스)
@CODE:UPLOAD-001:INFRA   (스토리지 연동)
  ↓
@DOC:UPLOAD-001
```

#### 구현 예시

```typescript
// @CODE:UPLOAD-001:UI
export function FileUploader() {
  const upload = useUpload();
  return <input type="file" onChange={(e) => upload(e.target.files[0])} />;
}

// @CODE:UPLOAD-001:DOMAIN
export class UploadService {
  async upload(file: File): Promise<UploadResult> {
    const validation = this.validate(file);
    if (!validation.success) return validation;

    return await this.storage.save(file);
  }
}

// @CODE:UPLOAD-001:INFRA
export class S3Storage implements StorageProvider {
  async save(file: File): Promise<string> {
    const url = await this.s3Client.upload(file);
    return url;
  }
}
```

---

## TAG 문제 해결 시나리오

### 시나리오 1: TAG 중복 발생

**문제**:

```bash
$ rg "@SPEC:AUTH-001" -n
.moai/specs/SPEC-AUTH-001/spec.md:26:# @SPEC:AUTH-001: JWT 인증
.moai/specs/SPEC-AUTH-002/spec.md:26:# @SPEC:AUTH-001: OAuth 인증  # 중복!
```

**해결**:

```markdown
# 1. 잘못된 TAG 수정
# SPEC-AUTH-002/spec.md
- # @SPEC:AUTH-001: OAuth 인증
+ # @SPEC:AUTH-002: OAuth 인증

# 2. 중복 검증
$ rg "@SPEC:AUTH-001" -n
.moai/specs/SPEC-AUTH-001/spec.md:26:# @SPEC:AUTH-001: JWT 인증  # 유일!
```

### 시나리오 2: 고아 TAG 발생

**문제**:

```bash
$ rg '@CODE:PAYMENT-005' -n src/
src/payment/service.py:1:# @CODE:PAYMENT-005  # CODE 존재

$ rg '@SPEC:PAYMENT-005' -n .moai/specs/
# 결과 없음 → SPEC 없는 고아 TAG!
```

**해결**:

```markdown
# 1. SPEC 생성
$ /alfred:1-spec "PAYMENT-005: 환불 처리 기능"

# 2. TAG 체인 연결
# .moai/specs/SPEC-PAYMENT-005/spec.md
---
id: PAYMENT-005
---
# @SPEC:PAYMENT-005: 환불 처리 기능

# 3. 검증
$ rg '@(SPEC|CODE):PAYMENT-005' -n
.moai/specs/SPEC-PAYMENT-005/spec.md:7:# @SPEC:PAYMENT-005
src/payment/service.py:1:# @CODE:PAYMENT-005
```

### 시나리오 3: TAG 체인 끊김

**문제**:

```bash
# SPEC과 CODE는 있는데 TEST가 없음
$ rg '@SPEC:UPLOAD-003' -n  # ✅
$ rg '@CODE:UPLOAD-003' -n  # ✅
$ rg '@TEST:UPLOAD-003' -n  # ❌ 없음!
```

**해결**:

```python
# 1. 테스트 작성
# tests/upload/service.test.py
# @TEST:UPLOAD-003 | SPEC: .moai/specs/SPEC-UPLOAD-003/spec.md

def test_should_validate_file_size():
    service = UploadService()
    large_file = create_file(size=100 * 1024 * 1024)  # 100MB
    result = service.upload(large_file)
    assert result.success is False

# 2. 체인 검증
$ rg '@(SPEC|TEST|CODE):UPLOAD-003' -n
.moai/specs/SPEC-UPLOAD-003/spec.md:7:# @SPEC:UPLOAD-003
tests/upload/service.test.py:1:# @TEST:UPLOAD-003
src/upload/service.py:1:# @CODE:UPLOAD-003
```

### 시나리오 4: 복합 도메인 TAG 관리

**문제**: 리팩토링이 여러 도메인에 걸쳐 있을 때

**해결**:

```markdown
# 1. 복합 도메인 TAG 생성
.moai/specs/SPEC-REFACTOR-AUTH-UPLOAD-001/spec.md
---
id: REFACTOR-AUTH-UPLOAD-001
related_specs:
  - AUTH-001
  - UPLOAD-001
---

# @SPEC:REFACTOR-AUTH-UPLOAD-001: 인증 및 업로드 통합 리팩토링

# 2. 관련 TAG 명시
# src/shared/middleware.py
# @CODE:REFACTOR-AUTH-UPLOAD-001
# Related: @CODE:AUTH-001, @CODE:UPLOAD-001
```

### 시나리오 5: 버전 업그레이드 시 TAG 관리

**문제**: SPEC 버전 업데이트 시 TAG는?

**해결**:

```markdown
# TAG ID는 불변, 내용만 업데이트
# SPEC-AUTH-001/spec.md

# ❌ 잘못된 방법: TAG ID 변경
- # @SPEC:AUTH-001: JWT 인증 (v1.0.0)
+ # @SPEC:AUTH-002: JWT 인증 (v2.0.0)  # 금지!

# ✅ 올바른 방법: HISTORY 업데이트
---
id: AUTH-001
version: 2.0.0  # 버전만 증가
---

# @SPEC:AUTH-001: JWT 인증 시스템

## HISTORY
### v2.0.0 (2025-10-15)
- **BREAKING**: 토큰 만료시간 15분 → 5분으로 변경
- **AUTHOR**: @Goos
- **REASON**: 보안 강화

### v1.0.0 (2025-10-11)
- **INITIAL**: JWT 기반 인증 시스템 명세 작성
```

---

## rg 명령어 치트 시트

### 기본 검색

```bash
# TAG 패턴 검색
rg '@(SPEC|TEST|CODE|DOC):' -n

# 특정 TAG 검색
rg '@SPEC:AUTH-001' -n

# 대소문자 무시
rg -i '@spec:auth' -n

# 파일명만 출력
rg '@SPEC:' -l
```

### 고급 검색

```bash
# 특정 디렉토리만 검색
rg '@CODE:' -n src/

# 특정 파일 타입만 검색
rg '@TEST:' -t py        # Python 파일만
rg '@CODE:' -t ts        # TypeScript 파일만

# 컨텍스트 포함 (앞뒤 3줄)
rg '@SPEC:AUTH-001' -C 3

# JSON 형식 출력
rg '@TAG:' --json
```

### 검증 스크립트

```bash
# .moai/scripts/validate-tags.sh
#!/bin/bash

echo "=== TAG 체인 검증 ==="

# 1. 모든 SPEC TAG 추출
SPEC_TAGS=$(rg '@SPEC:([A-Z]+-[0-9]+)' -o -r '$1' -h .moai/specs/ | sort -u)

for tag in $SPEC_TAGS; do
  echo "검증 중: $tag"

  # TEST 존재 확인
  if ! rg -q "@TEST:$tag" tests/; then
    echo "  ⚠️  @TEST:$tag 없음"
  fi

  # CODE 존재 확인
  if ! rg -q "@CODE:$tag" src/; then
    echo "  ⚠️  @CODE:$tag 없음"
  fi
done

# 2. 고아 CODE TAG 탐지
CODE_TAGS=$(rg '@CODE:([A-Z]+-[0-9]+)' -o -r '$1' -h src/ | sort -u)

for tag in $CODE_TAGS; do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "❌ 고아 TAG: @CODE:$tag (SPEC 없음)"
  fi
done

echo "=== 검증 완료 ==="
```

---

## 관련 문서

- [EARS 요구사항 작성 가이드](./ears-guide.md)
- [TRUST 5원칙 가이드](./trust-principles.md)
- [SPEC-First TDD 워크플로우](./spec-first-tdd.md)
- 개발 가이드

---

**작성일**: 2025-10-11
**버전**: v1.0.0
**TAG**: @CODE:DOCS-002
