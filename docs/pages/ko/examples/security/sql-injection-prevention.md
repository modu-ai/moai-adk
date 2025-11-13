---
title: "SQL 인젝션 방지"
category: "security"
difficulty: "초급"
tags: [sql, security, sqlalchemy, injection]
---

# SQL 인젝션 방지

## 개요

SQL 인젝션 공격을 방지하는 안전한 쿼리 작성 방법을 다룹니다.

## ❌ 나쁜 예 (취약)

```python
# 절대 이렇게 하지 마세요!
user_id = request.query_params.get("id")
query = f"SELECT * FROM users WHERE id = {user_id}"
db.execute(query)  # SQL Injection 취약!
```

## ✅ 좋은 예 (안전)

```python
# SQLAlchemy ORM 사용
user = db.query(User).filter(User.id == user_id).first()

# 또는 파라미터 바인딩
from sqlalchemy import text
query = text("SELECT * FROM users WHERE id = :id")
result = db.execute(query, {"id": user_id})
```

## 관련 예제

- [입력 검증](/ko/examples/security/input-validation)
- [기본 CRUD](/ko/examples/rest-api/basic-crud)
