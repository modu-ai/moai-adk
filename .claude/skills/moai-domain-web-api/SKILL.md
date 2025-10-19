---
name: moai-domain-web-api
description: REST API and GraphQL design patterns with authentication, versioning, and OpenAPI documentation
allowed-tools:
  - Read
  - Bash
tier: 2
auto-load: "true"
---

# Web API Expert

## What it does

Provides expertise in designing and implementing RESTful APIs and GraphQL services, including authentication mechanisms (JWT, OAuth2), API versioning strategies, and OpenAPI documentation.

## When to use

- "API 설계", "REST API 패턴", "GraphQL 스키마", "JWT 인증", "OAuth2", "API 버전관리", "OpenAPI", "문서화", "HATEOAS", "N+1 문제"
- "API design", "REST API", "GraphQL", "JWT authentication", "OAuth2", "API versioning", "OpenAPI documentation"
- Automatically invoked when working with API projects
- Web API SPEC implementation (`/alfred:2-run`)

- "API 설계", "REST API 패턴", "GraphQL 스키마", "JWT 인증"
- Automatically invoked when working with API projects
- Web API SPEC implementation (`/alfred:2-run`)

## How it works

**REST API Design**:
- **RESTful principles**: Resource-based URLs, HTTP verbs (GET, POST, PUT, DELETE)
- **Status codes**: Proper use of 2xx, 4xx, 5xx codes
- **HATEOAS**: Hypermedia links in responses
- **Pagination**: Cursor-based or offset-based

**GraphQL Design**:
- **Schema definition**: Types, queries, mutations, subscriptions
- **Resolver implementation**: Data fetching logic
- **N+1 problem**: DataLoader for batching
- **Schema stitching**: Federated GraphQL

**Authentication & Authorization**:
- **JWT (JSON Web Token)**: Stateless authentication
- **OAuth2**: Authorization framework (flows: authorization code, client credentials)
- **API keys**: Simple authentication
- **RBAC/ABAC**: Role/Attribute-based access control

**API Versioning**:
- **URL versioning**: /v1/users, /v2/users
- **Header versioning**: Accept: application/vnd.api.v2+json
- **Deprecation strategy**: Sunset header

**Documentation**:
- **OpenAPI (Swagger)**: API specification
- **API documentation**: Auto-generated docs
- **Postman collections**: Request examples

## Examples

### Example 1: REST API with JWT Authentication

**RED (Test)**:
```python
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
def test_login_returns_jwt():
    """로그인 시 JWT 토큰 반환"""
    response = client.post('/auth/login', {
        'email': 'user@test.com',
        'password': 'password123'
    })
    assert response.status_code == 200
    assert 'token' in response.json()

def test_protected_endpoint_requires_jwt():
    """보호된 엔드포인트는 JWT 필수"""
    response = client.get('/api/profile')
    assert response.status_code == 401
    assert response.json()['error'] == 'Unauthorized'
```

**GREEN (Implementation)**:
```python
# @CODE:AUTH-001 | TEST: tests/test_auth.py
from datetime import datetime, timedelta
import jwt

@app.post('/auth/login')
def login(email: str, password: str):
    """로그인하여 JWT 토큰 발급"""
    user = db.query("SELECT * FROM users WHERE email = %s", email)

    if not user or not verify_password(password, user.password_hash):
        raise HTTPException(status_code=401, detail="Invalid credentials")

    token = jwt.encode({
        'user_id': user.id,
        'exp': datetime.utcnow() + timedelta(hours=24)
    }, secret_key='secret')

    return {'token': token}

@app.get('/api/profile')
def get_profile(authorization: str = Header(None)):
    """JWT 검증 후 프로필 반환"""
    if not authorization:
        raise HTTPException(status_code=401, detail="Unauthorized")

    try:
        token = authorization.replace('Bearer ', '')
        payload = jwt.decode(token, secret_key='secret')
        user_id = payload['user_id']
    except jwt.InvalidTokenError:
        raise HTTPException(status_code=401, detail="Invalid token")

    user = db.query("SELECT * FROM users WHERE id = %s", user_id)
    return user
```

**REFACTOR (Middleware)**:
```python
# @CODE:AUTH-001:REFACTOR | JWT 미들웨어
def verify_token(token: str = Depends(HTTPBearer())):
    """JWT 검증 미들웨어"""
    try:
        payload = jwt.decode(token.credentials, secret_key='secret')
        return payload['user_id']
    except jwt.InvalidTokenError:
        raise HTTPException(status_code=401, detail="Invalid token")

@app.get('/api/profile')
def get_profile(user_id: int = Depends(verify_token)):
    """미들웨어로 자동 검증"""
    user = db.query("SELECT * FROM users WHERE id = %s", user_id)
    return user

# 개선: 모든 보호된 엔드포인트에서 재사용 ✅
```

### Example 2: REST API vs GraphQL

**❌ REST API N+1 Problem**:
```bash
# 사용자 정보 + 주문 조회 (2개 요청)
GET /api/users/1
GET /api/users/1/orders

# 결과: 2회 왕복
```

**✅ GraphQL (Single Request)**:
```graphql
query GetUserWithOrders {
    user(id: 1) {
        id
        name
        email
        orders {
            id
            total
            items {
                name
                price
            }
        }
    }
}

# 결과: 1회 왕복 ✅
```

### Example 3: API Versioning Strategy

**URL Versioning**:
```
GET /v1/users      # 버전 1
GET /v2/users      # 버전 2 (다른 응답 형식)
```

**Header Versioning**:
```
GET /users
Headers: Accept: application/vnd.api+json;version=2
```

**Deprecation Header**:
```
HTTP/1.1 200 OK
Sunset: Wed, 21 Nov 2025 08:00:00 GMT
Deprecation: true
```

### Example 4: OpenAPI Documentation

**Auto-Generated (FastAPI)**:
```python
# @CODE:DOCS-001: Swagger 자동 생성
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class UserCreate(BaseModel):
    email: str
    name: str

@app.post("/users", response_model=UserCreate)
async def create_user(user: UserCreate):
    """사용자 생성

    - **email**: 사용자 이메일
    - **name**: 사용자 이름
    """
    return user

# 자동 생성 문서:
# - Swagger UI: http://localhost:8000/docs
# - ReDoc: http://localhost:8000/redoc
```

## Keywords

"API 설계", "REST", "GraphQL", "JWT 인증", "OAuth2", "API 버전관리", "OpenAPI", "문서화", "N+1 문제", "authentication", "authorization", "HATEOAS"

## Reference

- API design patterns: `.moai/memory/development-guide.md#API-설계-패턴`
- Authentication & Authorization: CLAUDE.md#인증-인가-패턴
- GraphQL best practices: `.moai/memory/development-guide.md#GraphQL-최적화`

## Works well with

- moai-domain-backend (서버 구현)
- moai-domain-security (보안 검증)
- moai-foundation-trust (API 테스트)
