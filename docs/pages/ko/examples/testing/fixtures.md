---
title: "테스트 픽스처"
category: "testing"
difficulty: "초급"
tags: [pytest, fixtures, setup, testing]
---

# 테스트 픽스처

## 개요

재사용 가능한 테스트 데이터 및 환경 설정을 픽스처로 관리합니다.

## 기본 사용법

```python
import pytest

@pytest.fixture
def sample_user():
    """샘플 사용자 픽스처"""
    return User(username="test", email="test@example.com")

@pytest.fixture
def db_session():
    """테스트 DB 세션"""
    # 설정
    session = TestingSessionLocal()
    yield session
    # 정리
    session.close()

def test_user_creation(sample_user, db_session):
    """픽스처 사용 예시"""
    db_session.add(sample_user)
    db_session.commit()
    assert sample_user.id is not None
```

## 관련 예제

- [단위 테스트](/ko/examples/testing/unit-tests)
