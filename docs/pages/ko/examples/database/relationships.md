---
title: "SQLAlchemy 관계"
category: "database"
difficulty: "중급"
tags: [sqlalchemy, relationships, orm, foreign-key]
---

# SQLAlchemy 관계

## 개요

SQLAlchemy에서 One-to-Many, Many-to-Many 등 테이블 관계를 설정합니다.

## One-to-Many

```python
class User(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True)
    posts = relationship("Post", back_populates="author")

class Post(Base):
    __tablename__ = "posts"
    id = Column(Integer, primary_key=True)
    author_id = Column(Integer, ForeignKey("users.id"))
    author = relationship("User", back_populates="posts")
```

## Many-to-Many

```python
association_table = Table('user_roles', Base.metadata,
    Column('user_id', ForeignKey('users.id')),
    Column('role_id', ForeignKey('roles.id'))
)

class User(Base):
    __tablename__ = "users"
    roles = relationship("Role", secondary=association_table)
```

## 관련 예제

- [쿼리 최적화](/ko/examples/database/query-optimization)
