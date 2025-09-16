# 코딩 및 아키텍처 기준

> MoAI-ADK 프로젝트의 코딩 표준 및 아키텍처 설계 원칙

## 💻 언어별 코딩 표준

### Python
```python
# 파일 헤더 (필수)
"""
Module description goes here.

This module implements [specific functionality].
"""

# Import 순서
import os
import sys
from pathlib import Path

import requests
import click

from .local_module import LocalClass

# 클래스 정의
class ExampleClass:
    """Class docstring with clear description."""
    
    def __init__(self, param: str) -> None:
        """Initialize with parameter validation."""
        self.param = param
    
    def public_method(self, arg: int) -> str:
        """Public method with type hints and docstring."""
        return self._private_method(arg)
    
    def _private_method(self, arg: int) -> str:
        """Private method prefix with underscore."""
        return f"{self.param}: {arg}"

    def process_data(self, data: any) -> str:
        """Python 3.11+ match-case 문법 활용"""
        match data:
            case str() if len(data) > 10:
                return f"Long string: {data[:10]}..."
            case str():
                return f"Short string: {data}"
            case int() | float() as number:
                return f"Number: {number}"
            case [first, *rest]:
                return f"List starting with: {first}"
            case {"name": str(name), "age": int(age)}:
                return f"Person: {name}, {age}"
            case _:
                return "Unknown data type"

    def handle_errors(self):
        """Python 3.11+ Exception Groups 활용"""
        errors = []
        try:
            # 여러 작업 수행
            pass
        except* ValueError as eg:
            errors.extend(eg.exceptions)
        except* TypeError as eg:
            errors.extend(eg.exceptions)

        if errors:
            raise ExceptionGroup("Multiple errors occurred", errors)
```

### TypeScript/JavaScript
```typescript
// 파일 헤더 (필수)
/**
 * @fileoverview Module description
 * @version 1.0.0
 */

// Import 순서
import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs';

import { LocalService } from './local.service';

// 인터페이스 정의
export interface ApiResponse<T = unknown> {
  data: T;
  status: number;
  message?: string;
}

// 클래스 정의
export class ExampleComponent implements OnInit {
  private readonly apiService: ApiService;
  
  constructor(apiService: ApiService) {
    this.apiService = apiService;
  }
  
  public ngOnInit(): void {
    this.loadData();
  }
  
  private async loadData(): Promise<void> {
    try {
      const response = await this.apiService.getData();
      this.processData(response);
    } catch (error) {
      this.handleError(error);
    }
  }
}
```

## 🏗️ 아키텍처 패턴

### 계층형 아키텍처
```
├── presentation/     # UI Layer
│   ├── controllers/
│   ├── middleware/
│   └── validators/
├── application/      # Business Logic Layer  
│   ├── services/
│   ├── use-cases/
│   └── dto/
├── domain/          # Domain Layer
│   ├── entities/
│   ├── repositories/
│   └── value-objects/
└── infrastructure/   # Data Layer
    ├── database/
    ├── external-apis/
    └── config/
```

### 마이크로서비스 패턴
- 서비스별 독립적 데이터베이스
- API Gateway 통한 라우팅
- 서비스 디스커버리 구현
- Circuit Breaker 패턴 적용

## 📏 네이밍 컨벤션

### 파일 및 디렉토리
```
snake_case.py         # Python 파일
kebab-case.ts         # TypeScript 파일
PascalCase.tsx        # React 컴포넌트
camelCase.service.ts  # 서비스 파일
```

### 변수 및 함수
```python
# Python
variable_name = "snake_case"
CONSTANT_VALUE = "UPPER_SNAKE_CASE"

def function_name(param_name: str) -> str:
    return param_name

class ClassName:
    pass
```

```typescript
// TypeScript
const variableName = "camelCase";
const CONSTANT_VALUE = "UPPER_SNAKE_CASE";

function functionName(paramName: string): string {
  return paramName;
}

class ClassName {
}

interface InterfaceName {
}
```

## 🔧 도구 및 설정

### Python 도구
```toml
# pyproject.toml
[tool.black]
line-length = 88
target-version = ['py39']

[tool.ruff]
select = ["E", "F", "I", "N", "W"]
line-length = 88

[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py", "*_test.py"]
```

### TypeScript/JavaScript 도구
```json
// .eslintrc.json
{
  "extends": [
    "@typescript-eslint/recommended",
    "prettier"
  ],
  "rules": {
    "@typescript-eslint/no-unused-vars": "error",
    "@typescript-eslint/explicit-function-return-type": "warn"
  }
}

// prettier.config.js
module.exports = {
  semi: true,
  trailingComma: 'es5',
  singleQuote: true,
  printWidth: 80,
  tabWidth: 2
};
```

## 🧪 테스트 표준

### 테스트 구조
```python
# Python pytest
def test_should_return_expected_result_when_valid_input():
    # Given
    input_data = "valid_input"
    expected = "expected_result"
    
    # When
    result = function_under_test(input_data)
    
    # Then
    assert result == expected

class TestExampleClass:
    def test_init_should_set_param(self):
        # Given/When
        instance = ExampleClass("test")
        
        # Then
        assert instance.param == "test"
```

```typescript
// TypeScript/Jest
describe('ExampleService', () => {
  let service: ExampleService;
  
  beforeEach(() => {
    service = new ExampleService();
  });
  
  it('should return expected result when valid input', () => {
    // Given
    const input = 'valid_input';
    const expected = 'expected_result';
    
    // When
    const result = service.process(input);
    
    // Then
    expect(result).toBe(expected);
  });
});
```

## 📋 코드 리뷰 체크리스트

### 기능성
- [ ] 요구사항 충족 여부
- [ ] 엣지 케이스 처리
- [ ] 에러 핸들링

### 코드 품질
- [ ] 네이밍 컨벤션 준수
- [ ] 중복 코드 제거
- [ ] 복잡도 관리 (McCabe < 10)

### 테스트
- [ ] 단위 테스트 작성
- [ ] 통합 테스트 커버리지
- [ ] 테스트 시나리오 충분성

### 보안
- [ ] 입력값 검증
- [ ] SQL 인젝션 방지
- [ ] XSS 방지

### 성능
- [ ] 알고리즘 최적화
- [ ] 메모리 사용량
- [ ] 데이터베이스 쿼리 최적화

## 📦 의존성 관리

### Python
```toml
# pyproject.toml
[tool.poetry.dependencies]
python = "^3.11"
requests = "^2.28.0"
click = "^8.1.0"

[tool.poetry.group.dev.dependencies]
pytest = "^7.1.0"
black = "^22.0.0"
ruff = "^0.1.0"
```

### JavaScript/TypeScript
```json
{
  "dependencies": {
    "express": "^4.18.0",
    "typescript": "^4.9.0"
  },
  "devDependencies": {
    "@types/node": "^18.0.0",
    "jest": "^29.0.0",
    "eslint": "^8.0.0"
  }
}
```

## 🔄 CI/CD 파이프라인

### GitHub Actions 예제
```yaml
name: CI/CD Pipeline

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      - name: Install dependencies
        run: pip install -e .[dev]
      - name: Run tests
        run: pytest
      - name: Run linting
        run: ruff check .
      - name: Check formatting
        run: black --check .
```

---

**마지막 업데이트**: 2025-09-15  
**버전**: v0.1.12