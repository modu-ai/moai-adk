---
title: "XSS 방어"
category: "security"
difficulty: "중급"
tags: [xss, security, sanitization, html]
---

# XSS 방어 (Cross-Site Scripting)

## 개요

XSS 공격을 방지하기 위한 출력 이스케이핑과 입력 검증을 구현합니다.

## HTML 이스케이핑

```python
import html

def sanitize_html(text: str) -> str:
    """HTML 특수 문자 이스케이핑"""
    return html.escape(text)

# 사용 예시
user_input = "<script>alert('XSS')</script>"
safe_output = sanitize_html(user_input)
# 결과: "&lt;script&gt;alert('XSS')&lt;/script&gt;"
```

## 안전한 HTML 렌더링

```python
from bleach import clean

ALLOWED_TAGS = ['p', 'br', 'strong', 'em', 'a']
ALLOWED_ATTRIBUTES = {'a': ['href']}

def clean_html(html_content: str) -> str:
    """허용된 HTML 태그만 유지"""
    return clean(
        html_content,
        tags=ALLOWED_TAGS,
        attributes=ALLOWED_ATTRIBUTES,
        strip=True
    )
```

## 관련 예제

- [입력 검증](/ko/examples/security/input-validation)
