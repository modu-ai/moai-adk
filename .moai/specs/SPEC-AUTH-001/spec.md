# SPEC-AUTH-001: JWT 토큰 기반 사용자 인증 시스템

**EARS 형식 명세 (시스템 요구사항)**

---

## 1. 개요

JWT(JSON Web Token) 기반의 사용자 인증 시스템을 구현합니다. 사용자는 이메일과 비밀번호로 로그인하여 JWT 토큰을 발급받고, 토큰으로 사용자 정보를 조회하며, 로그아웃 시 토큰을 무효화할 수 있습니다.

---

## 2. 요구사항

### 2.1 사용자 로그인 (Essentials)

**Given** 등록된 사용자가 존재함
**When** 올바른 이메일과 비밀번호로 로그인을 시도함
**Then** JWT 토큰이 발급됨
**And** 토큰에는 사용자 ID, 이메일, 발급 시간이 포함됨
**And** 토큰의 유효시간은 1시간임

### 2.2 잘못된 자격증명 처리 (Essentials)

**Given** 사용자 로그인 요청이 있음
**When** 이메일이 존재하지 않음
**Then** "User not found" 에러를 반환함
**And** HTTP 상태 코드는 401 Unauthorized임

**When** 비밀번호가 올바르지 않음
**Then** "Invalid password" 에러를 반환함
**And** HTTP 상태 코드는 401 Unauthorized임

### 2.3 토큰 검증 및 사용자 정보 조회 (Essentials)

**Given** 유효한 JWT 토큰이 있음
**When** Authorization 헤더에 "Bearer {token}" 형식으로 토큰을 전달함
**Then** 사용자의 ID, 이메일, 생성일시가 반환됨
**And** HTTP 상태 코드는 200 OK임

### 2.4 만료된 토큰 처리 (Essentials)

**Given** 만료된 JWT 토큰이 있음
**When** 이 토큰으로 사용자 정보를 조회함
**Then** "Token expired" 에러를 반환함
**And** HTTP 상태 코드는 401 Unauthorized임

### 2.5 로그아웃 (Essentials)

**Given** 사용자가 로그인되어 있음(유효한 토큰 보유)
**When** 로그아웃 엔드포인트를 호출함
**Then** 해당 토큰이 블랙리스트에 추가됨
**And** HTTP 상태 코드는 200 OK임

### 2.6 블랙리스트 토큰 처리 (Essentials)

**Given** 로그아웃한 토큰이 있음(블랙리스트에 포함)
**When** 이 토큰으로 요청을 시도함
**Then** "Token has been revoked" 에러를 반환함
**And** HTTP 상태 코드는 401 Unauthorized임

### 2.7 비밀번호 암호화 (Essential-Quality)

**Given** 새로운 사용자가 생성됨
**When** 비밀번호가 저장됨
**Then** bcrypt를 사용하여 해싱되어 저장됨
**And** 평문 비밀번호는 저장되지 않음

---

## 3. 기술 요구사항

### 3.1 스택
- **프레임워크**: FastAPI
- **데이터베이스**: SQLite (테스트용)
- **인증**: JWT (PyJWT)
- **비밀번호**: bcrypt
- **테스트**: pytest

### 3.2 모듈 구조
```
src/moai_adk/auth/
├── __init__.py
├── models.py           # 사용자 모델
├── services.py         # 인증 비즈니스 로직
├── router.py          # FastAPI 라우트
└── security.py        # 보안 유틸리티 (해싱, JWT)

tests/
├── test_auth_login.py          # 로그인 테스트
├── test_auth_validate.py        # 토큰 검증 테스트
├── test_auth_logout.py         # 로그아웃 테스트
└── test_auth_password.py       # 비밀번호 처리 테스트
```

### 3.3 API 엔드포인트

| 메서드 | 엔드포인트        | 설명                    | 요청 본문                          |
|--------|-------------------|------------------------|-----------------------------------|
| POST   | `/auth/login`     | 사용자 로그인           | `{"email": str, "password": str}` |
| POST   | `/auth/logout`    | 토큰 무효화             | Authorization 헤더 필수           |
| GET    | `/auth/me`        | 현재 사용자 정보 조회   | Authorization 헤더 필수           |

---

## 4. 데이터 모델

### 4.1 User 모델
```python
class User:
    id: str                    # UUID
    email: str                 # 유일한 이메일
    hashed_password: str       # bcrypt로 해싱된 비밀번호
    created_at: datetime       # 사용자 생성 시간
```

### 4.2 LoginRequest
```python
class LoginRequest:
    email: str
    password: str
```

### 4.3 LoginResponse
```python
class LoginResponse:
    access_token: str
    token_type: str = "bearer"
    expires_in: int = 3600
```

### 4.4 UserResponse
```python
class UserResponse:
    id: str
    email: str
    created_at: datetime
```

---

## 5. 품질 기준 (TRUST-5)

### Test-first
- 모든 기능은 테스트부터 작성
- 테스트 커버리지 최소 85% 이상

### Readable
- 함수/변수명은 명확한 의도 표현
- 복잡한 로직은 함수로 분리

### Unified
- 일관된 에러 응답 형식
- 표준 HTTP 상태 코드 사용

### Secured
- bcrypt를 사용한 비밀번호 해싱
- JWT 서명을 통한 토큰 무결성 보장
- 로그아웃 시 토큰 블랙리스트 관리

### Trackable
- 사용자 생성 시간 기록
- 모든 로그인/로그아웃 시도 로깅

---

## 6. 구현 체크리스트

- [ ] User 모델 구현 (ORM)
- [ ] 비밀번호 해싱 함수 (bcrypt)
- [ ] JWT 생성 및 검증 함수
- [ ] 로그인 비즈니스 로직
- [ ] 사용자 정보 조회 로직
- [ ] 로그아웃 및 블랙리스트 관리
- [ ] FastAPI 라우터 구현
- [ ] 모든 엔드포인트에 대한 테스트 (RED 단계)
- [ ] 최소 구현으로 테스트 통과 (GREEN 단계)
- [ ] 코드 리팩토링 (REFACTOR 단계)
- [ ] 테스트 커버리지 85% 이상 확인

---

## 7. 성공 기준

- 모든 요구사항에 대한 테스트 작성 완료
- 모든 테스트 통과 (100% 성공률)
- 코드 복잡도 감소
- 테스트 커버리지 85% 이상
- 코드 가독성 개선
