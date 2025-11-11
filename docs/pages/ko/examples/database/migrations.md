---
title: "Alembic 마이그레이션"
category: "database"
difficulty: "초급"
tags: [alembic, migration, database, sqlalchemy]
---

# Alembic 마이그레이션

## 개요

Alembic을 사용하여 데이터베이스 스키마 변경사항을 버전 관리합니다.

## 주요 명령어

```bash
# 초기 설정
alembic init alembic

# 마이그레이션 생성
alembic revision --autogenerate -m "Add users table"

# 마이그레이션 적용
alembic upgrade head

# 롤백
alembic downgrade -1

# 현재 버전 확인
alembic current

# 마이그레이션 히스토리
alembic history
```

## 관련 예제

- [SQLAlchemy 관계](/ko/examples/database/relationships)
- [기본 CRUD](/ko/examples/rest-api/basic-crud)
