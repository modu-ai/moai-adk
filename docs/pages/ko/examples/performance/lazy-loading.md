---
title: "지연 로딩"
category: "performance"
difficulty: "중급"
tags: [sqlalchemy, lazy-loading, performance, n+1]
---

# 지연 로딩

## 개요

필요한 시점에 데이터를 로드하여 메모리를 최적화합니다.

## Lazy vs Eager Loading

```python
# Lazy Loading (기본)
user = db.query(User).first()
posts = user.posts  # 이 시점에 쿼리 실행

# Eager Loading
from sqlalchemy.orm import joinedload

user = db.query(User).options(joinedload(User.posts)).first()
posts = user.posts  # 이미 로드됨
```

## 관련 예제

- [쿼리 최적화](/ko/examples/database/query-optimization)
