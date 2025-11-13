---
title: "데이터베이스 예제"
description: "SQLAlchemy와 Alembic을 사용한 데이터베이스 관리"
---

# 데이터베이스 예제

SQLAlchemy ORM과 Alembic 마이그레이션을 사용한 데이터베이스 관리 예제입니다.

## 📚 예제 목록

### [Alembic 마이그레이션](/ko/examples/database/migrations)
**난이도**: 초급 | **태그**: `alembic`, `migration`, `database`

데이터베이스 스키마 버전 관리를 위한 Alembic 사용법

### [SQLAlchemy 관계](/ko/examples/database/relationships)
**난이도**: 중급 | **태그**: `sqlalchemy`, `relationships`, `orm`

One-to-Many, Many-to-Many 등 테이블 관계 설정

### [트랜잭션 처리](/ko/examples/database/transactions)
**난이도**: 중급 | **태그**: `transaction`, `acid`, `rollback`

데이터 일관성을 보장하는 트랜잭션 관리

### [쿼리 최적화](/ko/examples/database/query-optimization)
**난이도**: 고급 | **태그**: `performance`, `optimization`, `indexing`

N+1 문제 해결, 인덱싱, 쿼리 성능 향상

---

## 🎯 핵심 개념

### SQLAlchemy ORM
```python
from sqlalchemy import Column, Integer, String, ForeignKey
from sqlalchemy.orm import relationship

class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True)
    name = Column(String, nullable=False)

    posts = relationship("Post", back_populates="author")
```

### Alembic 마이그레이션
```bash
# 마이그레이션 생성
alembic revision --autogenerate -m "Add users table"

# 마이그레이션 적용
alembic upgrade head

# 롤백
alembic downgrade -1
```

## 📖 관련 문서

- [Tutorial 01: FastAPI + SQLAlchemy](/ko/tutorials/tutorial-01-fastapi)
- [REST API 예제](/ko/examples/rest-api/)
- [성능 예제](/ko/examples/performance/)

---

**시작하기**: [Alembic 마이그레이션](/ko/examples/database/migrations) 예제부터 시작하세요!
