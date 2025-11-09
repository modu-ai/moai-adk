Translate the following Korean markdown document to Japanese.

**CRITICAL RULES:**
1. Preserve ALL markdown structure (headers, code blocks, links, tables, diagrams)
2. Keep ALL code blocks and technical terms UNCHANGED
3. Maintain the EXACT same file structure and formatting
4. Translate ONLY Korean text content
5. Keep ALL @TAG references unchanged (e.g., @SPEC:AUTH-001)
6. Preserve ALL file paths and URLs
7. Keep ALL emoji and icons as-is
8. Maintain ALL frontmatter (YAML) structure

**Source File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ko/contributing/style.md
**Target Language:** Japanese
**Target File:** /Users/goos/MoAI/MoAI-ADK/docs/src/ja/contributing/style.md

**Content to Translate:**

---
title: 코드 스타일 가이드
description: MoAI-ADK Python, Markdown, YAML 코드 스타일 표준
status: stable
---

# 코드 스타일 가이드

MoAI-ADK의 코드 스타일 표준을 설명합니다. 모든 기여자는 이 가이드를 따라야 합니다.

## 🐍 Python 코드 스타일

### 표준 준수

- **표준**: PEP 8 + Black 포매팅
- **린터**: Ruff + mypy (타입 검사)
- **포매터**: Black (자동 포매팅)

### 파일 구조

```python
"""
모듈 설명.

이 모듈은... 상세 설명
"""

# 표준 라이브러리
import os
import sys
from pathlib import Path
from typing import Optional

# 서드파티 라이브러리
import pytest
from pydantic import BaseModel

# 로컬 라이브러리
from moai_adk.core import Agent
from moai_adk.utils import logger


class MyClass:
    """클래스 설명."""

    def method(self) -> None:
        """메서드 설명."""
        pass
```

### 네이밍 규칙

| 항목 | 규칙 | 예 |
|------|------|------|
| **클래스** | PascalCase | `class MyAgent:` |
| **함수/메서드** | snake_case | `def get_config():` |
| **상수** | UPPER_SNAKE_CASE | `DEFAULT_TIMEOUT = 30` |
| **비공개** | _leading_underscore | `def _internal_method():` |
| **모듈** | snake_case | `my_module.py` |

### 타입 힌트

```python
from typing import Optional, List, Dict, Union

def process_data(
    items: List[str],
    config: Optional[Dict[str, int]] = None,
) -> bool:
    """
    데이터 처리 함수.

    Args:
        items: 처리할 항목 목록
        config: 선택적 설정 딕셔너리

    Returns:
        처리 성공 여부

    Raises:
        ValueError: 잘못된 입력
    """
    if not items:
        raise ValueError("items cannot be empty")
    return True
```

### 주석 및 독스트링

```python
def calculate_score(value: int) -> float:
    """
    점수 계산.

    이 함수는 입력 값을 기반으로 정규화된 점수를 계산합니다.
    범위는 0.0에서 1.0 사이입니다.

    Args:
        value: 계산할 입력 값 (0-100)

    Returns:
        정규화된 점수 (0.0-1.0)

    Examples:
        >>> calculate_score(50)
        0.5
    """
    # 범위 검증
    if not 0 <= value <= 100:
        raise ValueError(f"Value must be 0-100, got {value}")

    # 점수 계산
    return value / 100.0
```

### 라인 길이 및 포매팅

```python
# Black 기본값: 88자
# 길이가 길면 자동으로 줄바꿈됨

def long_function_name(
    param1: str,
    param2: int,
    param3: Optional[Dict[str, Any]] = None,
) -> Tuple[str, int]:
    """긴 함수 정의 예."""
    pass
```

## 📝 마크다운 스타일

### 파일 구조

```markdown
---
title: 페이지 제목
description: 페이지 설명
status: stable
---

# H1 제목

모든 마크다운 파일은 이 구조를 따릅니다.

## H2 섹션

### H3 소섹션

더 깊은 제목은 피합니다 (H4+ 사용 안 함).

### 리스트 형식

**숨은 리스트 (bullet points)**:
- 첫 번째 항목
- 두 번째 항목
- 세 번째 항목

**순서 리스트 (numbered)**:
1. 첫 번째 단계
2. 두 번째 단계
3. 세 번째 단계

### 강조

- **굵은글** (중요 강조)
- *기울임* (용어 강조)
- ` ` (코드 인라인)
```

### 코드 블록

````markdown
```python
# Python 코드
def hello():
    print("Hello, World!")
```

```bash
# Bash 커맨드
uv run pytest
```

```yaml
# YAML 설정
key: value
nested:
  item: value
```
````

### 테이블

```markdown
| 헤더 1 | 헤더 2 | 헤더 3 |
|--------|--------|--------|
| 내용 A | 내용 B | 내용 C |
| 내용 D | 내용 E | 내용 F |
```

## 🔧 YAML 스타일

### 설정 파일

```yaml
# 주석은 # 다음에 한 칸 띄움
key: value

# 중첩 구조는 2칸 들여쓰기
parent:
  child: value
  list_item:
    - item1
    - item2

# 복잡한 값은 여러 줄로 표현
description: |
  여러 줄
  텍스트는
  파이프로 표현합니다.
```

## :bullseye: 자동화된 스타일 검사

### Ruff (린팅)

```bash
# 스타일 검사
uv run ruff check src/

# 자동 수정
uv run ruff check --fix src/
```

### Black (포매팅)

```bash
# 포매팅 확인
uv run black --check src/

# 자동 포매팅
uv run black src/
```

### mypy (타입 검사)

```bash
# 타입 검사
uv run mypy src/moai_adk
```

### 통합 검사

```bash
# 모든 검사 실행
uv run ruff check src/
uv run black --check src/
uv run mypy src/moai_adk
```

## 📋 Pre-commit 설정

`.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix]

  - repo: https://github.com/psf/black
    rev: 23.10.0
    hooks:
      - id: black

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.6.0
    hooks:
      - id: mypy
```

## ✅ 체크리스트

PR 제출 전 확인:

- [ ] Python 코드가 Black으로 포매팅됨
- [ ] Ruff 린팅 통과 (오류 없음)
- [ ] mypy 타입 검사 통과
- [ ] 마크다운 파일이 올바른 구조임
- [ ] 코드에 주석 및 독스트링 추가됨
- [ ] 테스트 코드 포함 (테스트 커버리지 87%+)

## <span class="material-icons">library_books</span> 참고 자료

- [PEP 8](https://www.python.org/dev/peps/pep-0008/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [Black Code Style](https://black.readthedocs.io/)
- [CommonMark Spec](https://spec.commonmark.org/)

---

**Questions?** GitHub Discussions에서 스타일 관련 질문을 하세요!


**Instructions:**
- Translate the content above to Japanese
- Output ONLY the translated markdown content
- Do NOT include any explanations or comments
- Maintain EXACT markdown formatting
- Preserve ALL code blocks exactly as-is
