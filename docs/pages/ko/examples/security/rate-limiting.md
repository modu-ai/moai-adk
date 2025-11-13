---
title: "속도 제한 (Rate Limiting)"
category: "security"
difficulty: "중급"
tags: [rate-limit, security, redis, fastapi]
---

# 속도 제한

## 개요

API 남용 및 DDoS 공격을 방지하기 위한 속도 제한을 구현합니다.

## Redis 기반 Rate Limiting

```python
from fastapi import FastAPI, Request, HTTPException
import redis
import time

app = FastAPI()
redis_client = redis.Redis(host='localhost', port=6379, decode_responses=True)

def rate_limit(max_requests: int = 10, window_seconds: int = 60):
    """속도 제한 데코레이터"""
    def decorator(func):
        async def wrapper(request: Request, *args, **kwargs):
            # 클라이언트 IP
            client_ip = request.client.host
            key = f"rate_limit:{client_ip}"

            # 현재 요청 수
            current = redis_client.get(key)

            if current and int(current) >= max_requests:
                raise HTTPException(
                    status_code=429,
                    detail="요청 한도를 초과했습니다. 잠시 후 다시 시도해주세요."
                )

            # 요청 수 증가
            pipe = redis_client.pipeline()
            pipe.incr(key)
            pipe.expire(key, window_seconds)
            pipe.execute()

            return await func(request, *args, **kwargs)
        return wrapper
    return decorator

@app.get("/api/data")
@rate_limit(max_requests=10, window_seconds=60)
async def get_data(request: Request):
    """분당 10회로 제한된 엔드포인트"""
    return {"data": "success"}
```

## Slowapi 라이브러리 사용

```python
from slowapi import Limiter, _rate_limit_exceeded_handler
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded

limiter = Limiter(key_func=get_remote_address)
app.state.limiter = limiter
app.add_exception_handler(RateLimitExceeded, _rate_limit_exceeded_handler)

@app.get("/limited")
@limiter.limit("5/minute")
async def limited_route(request: Request):
    """분당 5회로 제한"""
    return {"message": "success"}
```

## 관련 예제

- [Redis 캐싱](/ko/examples/performance/caching)
- [JWT 인증](/ko/examples/authentication/jwt-basic)
