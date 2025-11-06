<!-- @DOC:AUTH-001 -->

# JWT 인증 시스템 문서

## 개요

MoAI-ADK 프로젝트의 JWT(JSON Web Token) 기반 인증 시스템에 대한 문서입니다.

## 기능

- JWT 토큰 생성 및 검증
- 사용자 로그인/로그아웃
- 역할 기반 접근 제어
- 토큰 갱신 메커니즘

## API

### 사용자 인증
```python
# 로그인
POST /api/auth/login
{
    "username": "user@example.com",
    "password": "password"
}

# 응답
{
    "access_token": "jwt_token_here",
    "refresh_token": "refresh_token_here",
    "expires_in": 3600
}
```

## 보안

- HMAC SHA256 알고리즘 사용
- 토큰 만료 시간: 1시간
- 리프레시 토큰: 7일 유효

## 사용 예시

```python
from src.auth.example import example_function

# 예제 함수 호출
result = example_function()
```

## Relates
- @SPEC:AUTH-001
- @TEST:AUTH-001
- @CODE:AUTH-001