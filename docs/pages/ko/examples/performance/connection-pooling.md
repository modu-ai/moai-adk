---
title: "커넥션 풀링"
category: "performance"
difficulty: "초급"
tags: [database, pooling, optimization, sqlalchemy]
---

# 커넥션 풀링

## 개요

데이터베이스 커넥션을 재사용하여 리소스 효율성을 높입니다.

## SQLAlchemy 설정

```python
from sqlalchemy import create_engine

engine = create_engine(
    DATABASE_URL,
    pool_size=10,          # 기본 커넥션 수
    max_overflow=20,       # 최대 추가 커넥션
    pool_timeout=30,       # 타임아웃 (초)
    pool_recycle=3600,     # 커넥션 재생성 주기
    pool_pre_ping=True     # 커넥션 유효성 체크
)
```

## 관련 예제

- [Redis 캐싱](/ko/examples/performance/caching)
