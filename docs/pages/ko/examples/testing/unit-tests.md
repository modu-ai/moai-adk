---
title: "Pytest 단위 테스트"
category: "testing"
difficulty: "초급"
tags: [pytest, unit-test, tdd, testing, python]
---

# Pytest 단위 테스트

## 개요

Pytest를 사용하여 개별 함수 및 클래스를 테스트하는 단위 테스트를 작성합니다. TDD (Test-Driven Development) 방법론의 핵심입니다.

## 사용 사례

- 함수 로직 검증
- 클래스 메서드 테스트
- 엣지 케이스 처리
- 리팩토링 안정성 확보

## 완전한 코드 예제

### 1. 프로젝트 구조

```
my-project/
├── app/
│   ├── __init__.py
│   ├── calculator.py     # 테스트 대상
│   └── user.py           # 테스트 대상
├── tests/
│   ├── __init__.py
│   ├── test_calculator.py
│   └── test_user.py
├── pytest.ini            # Pytest 설정
└── requirements.txt
```

### 2. 테스트 대상 코드 (`app/calculator.py`)

```python
# SPEC: CALC-001 - 계산기 함수

from typing import Union

def add(a: Union[int, float], b: Union[int, float]) -> Union[int, float]:
    """
    두 숫자를 더합니다

    Args:
        a: 첫 번째 숫자
        b: 두 번째 숫자

    Returns:
        합계
    """
    return a + b

def divide(a: Union[int, float], b: Union[int, float]) -> float:
    """
    두 숫자를 나눕니다

    Args:
        a: 분자
        b: 분모

    Returns:
        나눈 결과

    Raises:
        ValueError: 0으로 나누기 시도 시
    """
    if b == 0:
        raise ValueError("0으로 나눌 수 없습니다")
    return a / b

def calculate_discount(price: float, discount_percent: int) -> float:
    """
    할인 가격을 계산합니다

    Args:
        price: 원래 가격
        discount_percent: 할인율 (0-100)

    Returns:
        할인된 가격

    Raises:
        ValueError: 유효하지 않은 입력값
    """
    if price < 0:
        raise ValueError("가격은 0 이상이어야 합니다")

    if not 0 <= discount_percent <= 100:
        raise ValueError("할인율은 0-100 사이여야 합니다")

    discount_amount = price * (discount_percent / 100)
    return price - discount_amount
```

### 3. 단위 테스트 (`tests/test_calculator.py`)

```python
# SPEC: TEST-CALC-001 - 계산기 단위 테스트

import pytest
from app.calculator import add, divide, calculate_discount

class TestAdd:
    """덧셈 함수 테스트"""

    def test_add_positive_numbers(self):
        """양수 덧셈 테스트"""
        result = add(5, 3)
        assert result == 8

    def test_add_negative_numbers(self):
        """음수 덧셈 테스트"""
        result = add(-5, -3)
        assert result == -8

    def test_add_mixed_numbers(self):
        """양수와 음수 덧셈 테스트"""
        result = add(5, -3)
        assert result == 2

    def test_add_floats(self):
        """실수 덧셈 테스트"""
        result = add(5.5, 2.3)
        assert result == pytest.approx(7.8)  # 부동소수점 비교

    def test_add_zero(self):
        """0 덧셈 테스트"""
        result = add(5, 0)
        assert result == 5

class TestDivide:
    """나눗셈 함수 테스트"""

    def test_divide_positive_numbers(self):
        """양수 나눗셈 테스트"""
        result = divide(10, 2)
        assert result == 5.0

    def test_divide_with_remainder(self):
        """나머지가 있는 나눗셈 테스트"""
        result = divide(10, 3)
        assert result == pytest.approx(3.333333, rel=1e-5)

    def test_divide_by_zero(self):
        """0으로 나누기 테스트 (예외 발생)"""
        with pytest.raises(ValueError) as exc_info:
            divide(10, 0)

        assert "0으로 나눌 수 없습니다" in str(exc_info.value)

    def test_divide_negative_numbers(self):
        """음수 나눗셈 테스트"""
        result = divide(-10, 2)
        assert result == -5.0

class TestCalculateDiscount:
    """할인 계산 함수 테스트"""

    def test_calculate_50_percent_discount(self):
        """50% 할인 테스트"""
        result = calculate_discount(100, 50)
        assert result == 50.0

    def test_calculate_no_discount(self):
        """할인 없음 테스트"""
        result = calculate_discount(100, 0)
        assert result == 100.0

    def test_calculate_full_discount(self):
        """100% 할인 테스트"""
        result = calculate_discount(100, 100)
        assert result == 0.0

    def test_negative_price_raises_error(self):
        """음수 가격 예외 테스트"""
        with pytest.raises(ValueError) as exc_info:
            calculate_discount(-100, 10)

        assert "가격은 0 이상이어야 합니다" in str(exc_info.value)

    def test_invalid_discount_percent_raises_error(self):
        """잘못된 할인율 예외 테스트"""
        with pytest.raises(ValueError) as exc_info:
            calculate_discount(100, 150)  # 100 초과

        assert "할인율은 0-100 사이여야 합니다" in str(exc_info.value)

    @pytest.mark.parametrize("price,discount,expected", [
        (100, 10, 90),
        (200, 25, 150),
        (50, 20, 40),
        (1000, 5, 950),
    ])
    def test_various_discounts(self, price, discount, expected):
        """다양한 할인 시나리오 (파라미터화 테스트)"""
        result = calculate_discount(price, discount)
        assert result == expected
```

### 4. 사용자 클래스 테스트 (`tests/test_user.py`)

```python
# SPEC: TEST-USER-001 - 사용자 클래스 테스트

import pytest
from app.user import User, UserManager

class TestUser:
    """User 클래스 테스트"""

    def test_user_creation(self):
        """사용자 생성 테스트"""
        user = User(username="john", email="john@example.com")
        assert user.username == "john"
        assert user.email == "john@example.com"
        assert user.is_active is True  # 기본값

    def test_user_full_name(self):
        """전체 이름 속성 테스트"""
        user = User(
            username="john",
            email="john@example.com",
            first_name="John",
            last_name="Doe"
        )
        assert user.full_name == "John Doe"

    def test_user_representation(self):
        """__repr__ 테스트"""
        user = User(username="john", email="john@example.com")
        assert "john" in repr(user)

class TestUserManager:
    """UserManager 클래스 테스트"""

    @pytest.fixture
    def manager(self):
        """UserManager 픽스처"""
        return UserManager()

    @pytest.fixture
    def sample_users(self, manager):
        """샘플 사용자들"""
        manager.add_user(User(username="alice", email="alice@example.com"))
        manager.add_user(User(username="bob", email="bob@example.com"))
        return manager

    def test_add_user(self, manager):
        """사용자 추가 테스트"""
        user = User(username="john", email="john@example.com")
        manager.add_user(user)
        assert manager.count() == 1

    def test_add_duplicate_user(self, manager):
        """중복 사용자 추가 예외 테스트"""
        user1 = User(username="john", email="john@example.com")
        user2 = User(username="john", email="different@example.com")

        manager.add_user(user1)

        with pytest.raises(ValueError) as exc_info:
            manager.add_user(user2)

        assert "이미 존재" in str(exc_info.value)

    def test_get_user_by_username(self, sample_users):
        """사용자명으로 조회 테스트"""
        user = sample_users.get_user("alice")
        assert user is not None
        assert user.username == "alice"

    def test_get_nonexistent_user(self, sample_users):
        """존재하지 않는 사용자 조회 테스트"""
        user = sample_users.get_user("nonexistent")
        assert user is None

    def test_remove_user(self, sample_users):
        """사용자 삭제 테스트"""
        assert sample_users.count() == 2
        sample_users.remove_user("alice")
        assert sample_users.count() == 1
        assert sample_users.get_user("alice") is None
```

### 5. Pytest 설정 (`pytest.ini`)

```ini
[pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
addopts =
    -v
    --strict-markers
    --tb=short
    --cov=app
    --cov-report=term-missing
    --cov-report=html
markers =
    slow: marks tests as slow
    integration: marks tests as integration tests
```

## 테스트 실행

### 기본 실행

```bash
# 모든 테스트 실행
pytest

# 특정 파일만 실행
pytest tests/test_calculator.py

# 특정 클래스만 실행
pytest tests/test_calculator.py::TestAdd

# 특정 테스트만 실행
pytest tests/test_calculator.py::TestAdd::test_add_positive_numbers
```

### 커버리지 리포트

```bash
# 커버리지 포함 실행
pytest --cov=app tests/

# HTML 리포트 생성
pytest --cov=app --cov-report=html tests/

# 누락된 라인 표시
pytest --cov=app --cov-report=term-missing tests/
```

### 출력 제어

```bash
# 상세 출력
pytest -v

# 매우 상세한 출력
pytest -vv

# 실패한 테스트만 재실행
pytest --lf

# 실패한 테스트 먼저 실행
pytest --ff
```

## TDD 워크플로우

### RED-GREEN-REFACTOR 사이클

```python
# 1️⃣ RED: 실패하는 테스트 먼저 작성
def test_multiply():
    result = multiply(3, 4)
    assert result == 12

# 실행: pytest -k test_multiply
# 결과: FAILED (multiply 함수가 없음)

# 2️⃣ GREEN: 최소한의 코드로 테스트 통과
def multiply(a, b):
    return a * b

# 실행: pytest -k test_multiply
# 결과: PASSED

# 3️⃣ REFACTOR: 코드 개선 (테스트는 그대로)
def multiply(a: Union[int, float], b: Union[int, float]) -> Union[int, float]:
    """두 숫자를 곱합니다"""
    if not isinstance(a, (int, float)) or not isinstance(b, (int, float)):
        raise TypeError("숫자만 입력 가능합니다")
    return a * b

# 실행: pytest -k test_multiply
# 결과: PASSED (여전히 통과)
```

## Best Practices

### 테스트 작성
- ✅ **하나의 개념만 테스트**: 각 테스트는 하나의 동작만 검증
- ✅ **명확한 테스트명**: `test_divide_by_zero` (무엇을 테스트하는지 명확)
- ✅ **AAA 패턴**: Arrange (준비) → Act (실행) → Assert (검증)
- ✅ **엣지 케이스 테스트**: 0, 음수, 빈 값, None 등
- ✅ **예외 테스트**: `pytest.raises()` 사용

### 테스트 구조
- ✅ **클래스로 그룹화**: 관련된 테스트를 `TestXxx` 클래스로 묶기
- ✅ **픽스처 활용**: 반복되는 설정 코드 제거
- ✅ **파라미터화**: `@pytest.mark.parametrize`로 여러 케이스 테스트
- ✅ **격리**: 각 테스트는 독립적으로 실행 가능

### 코드 품질
- ✅ **커버리지 80% 이상**: 중요한 로직은 반드시 테스트
- ✅ **빠른 실행**: 단위 테스트는 빠르게 (<1초)
- ✅ **결정론적**: 같은 입력에 항상 같은 결과
- ✅ **읽기 쉬운 코드**: 주석보다 명확한 테스트명

## 주의사항

### 피해야 할 것
- ❌ **여러 개념 한 번에 테스트**: 테스트가 실패하면 원인 파악 어려움
- ❌ **외부 의존성**: 단위 테스트는 DB, API 등 격리 (모킹 사용)
- ❌ **랜덤 값 사용**: 테스트 결과가 불안정해짐
- ❌ **테스트 순서 의존**: 테스트 간 순서 보장 안 됨

### 성능 고려
- ⚠️ **느린 테스트**: 1초 이상 걸리면 `@pytest.mark.slow` 표시
- ⚠️ **불필요한 설정**: 픽스처로 공통 코드 재사용
- ⚠️ **과도한 테스트**: Getter/Setter 같은 간단한 코드는 생략 가능

## 관련 예제

- [통합 테스트](/ko/examples/testing/integration-tests) - API, DB 전체 테스트
- [테스트 픽스처](/ko/examples/testing/fixtures) - 재사용 가능한 설정
- [모킹](/ko/examples/testing/mocking) - 외부 의존성 격리
- [TDD 개발 가이드](/ko/guides/tdd-development) - TDD 방법론 상세

## 참고 자료

- [Pytest 공식 문서](https://docs.pytest.org/)
- [Effective Python Testing With Pytest](https://realpython.com/pytest-python-testing/)
- [Test-Driven Development with Python](https://www.obeythetestinggoat.com/)
