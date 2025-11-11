---
title: "외부 API 모킹"
category: "testing"
difficulty: "중급"
tags: [pytest, mock, unittest, testing]
---

# 외부 API 모킹

## 개요

외부 의존성(API, DB 등)을 격리하여 테스트합니다.

## 기본 사용법

```python
from unittest.mock import Mock, patch

@patch('requests.get')
def test_fetch_user_data(mock_get):
    """외부 API 호출 모킹"""
    # Mock 응답 설정
    mock_get.return_value.json.return_value = {
        "id": 1,
        "name": "John Doe"
    }

    # 테스트 실행
    result = fetch_user_data(user_id=1)

    # 검증
    assert result["name"] == "John Doe"
    mock_get.assert_called_once_with("https://api.example.com/users/1")
```

## 관련 예제

- [단위 테스트](/ko/examples/testing/unit-tests)
