---
title: "쿼리 최적화"
category: "database"
difficulty: "고급"
tags: [performance, optimization, sqlalchemy, indexing]
---

# 쿼리 최적화

## 개요

N+1 문제 해결, 인덱싱, 쿼리 성능 향상 기법을 다룹니다.

## N+1 문제 해결

```python
# ❌ N+1 문제 (나쁜 예)
users = db.query(User).all()
for user in users:
    print(user.posts)  # 각 사용자마다 쿼리 1번 실행

# ✅ Eager Loading (좋은 예)
from sqlalchemy.orm import joinedload

users = db.query(User).options(joinedload(User.posts)).all()
for user in users:
    print(user.posts)  # 쿼리 1번만 실행
```

## 인덱스 추가

```python
class User(Base):
    __tablename__ = "users"
    email = Column(String, index=True)  # 인덱스 추가
    created_at = Column(DateTime, index=True)

    __table_args__ = (
        Index('idx_email_created', 'email', 'created_at'),  # 복합 인덱스
    )
```

## 관련 예제

- [페이지네이션](/ko/examples/rest-api/pagination)
- [Redis 캐싱](/ko/examples/performance/caching)
