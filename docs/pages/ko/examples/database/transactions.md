---
title: "트랜잭션 처리"
category: "database"
difficulty: "중급"
tags: [transaction, acid, rollback, sqlalchemy]
---

# 트랜잭션 처리

## 개요

데이터베이스 트랜잭션을 사용하여 데이터 일관성을 보장합니다.

## 기본 사용법

```python
from sqlalchemy.orm import Session

def transfer_money(db: Session, from_user_id: int, to_user_id: int, amount: float):
    """계좌 이체 (트랜잭션)"""
    try:
        # 출금
        from_user = db.query(User).filter(User.id == from_user_id).first()
        from_user.balance -= amount

        # 입금
        to_user = db.query(User).filter(User.id == to_user_id).first()
        to_user.balance += amount

        # 커밋
        db.commit()

    except Exception as e:
        # 롤백
        db.rollback()
        raise
```

## 관련 예제

- [기본 CRUD](/ko/examples/rest-api/basic-crud)
