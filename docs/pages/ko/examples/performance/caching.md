---
title: "Redis 캐싱"
category: "performance"
difficulty: "중급"
tags: [redis, caching, performance, optimization]
---

# Redis 캐싱

## 개요

Redis를 사용하여 자주 조회되는 데이터를 캐싱해 응답 속도를 향상시킵니다.

## 기본 구현

```python
import redis
import json

redis_client = redis.Redis(host='localhost', port=6379, decode_responses=True)

def get_user_cached(user_id: int):
    """캐시를 사용한 사용자 조회"""
    cache_key = f"user:{user_id}"

    # 1. 캐시 확인
    cached = redis_client.get(cache_key)
    if cached:
        return json.loads(cached)

    # 2. DB 조회
    user = db.query(User).filter(User.id == user_id).first()

    # 3. 캐시 저장 (TTL: 5분)
    redis_client.setex(
        cache_key,
        300,  # 5분
        json.dumps(user.dict())
    )

    return user
```

## 캐시 무효화

```python
def update_user(user_id: int, data: dict):
    """사용자 수정 + 캐시 무효화"""
    # DB 업데이트
    user = db.query(User).filter(User.id == user_id).first()
    for key, value in data.items():
        setattr(user, key, value)
    db.commit()

    # 캐시 삭제
    redis_client.delete(f"user:{user_id}")
```

## 관련 예제

- [쿼리 최적화](/ko/examples/database/query-optimization)
