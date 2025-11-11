---
title: "통합 테스트"
category: "testing"
difficulty: "중급"
tags: [pytest, integration-test, api, fastapi]
---

# 통합 테스트

## 개요

API 엔드포인트, 데이터베이스 등 여러 컴포넌트를 함께 테스트합니다.

## 기본 구조

```python
from fastapi.testclient import TestClient

def test_user_creation_flow():
    """사용자 생성 전체 플로우 테스트"""
    client = TestClient(app)

    # 1. 회원가입
    response = client.post("/auth/register", json={
        "username": "testuser",
        "email": "test@example.com",
        "password": "Password123"
    })
    assert response.status_code == 201

    # 2. 로그인
    response = client.post("/auth/login", data={
        "username": "testuser",
        "password": "Password123"
    })
    token = response.json()["access_token"]

    # 3. 프로필 조회
    response = client.get("/users/me", headers={
        "Authorization": f"Bearer {token}"
    })
    assert response.status_code == 200
```

## 관련 예제

- [단위 테스트](/ko/examples/testing/unit-tests)
- [테스트 픽스처](/ko/examples/testing/fixtures)
