---
title: "Tutorial 3: 데이터베이스 성능 최적화"
description: "N+1 문제, 인덱스, 캐싱 전략으로 데이터베이스 성능을 극대화합니다"
duration: "1시간"
difficulty: "고급"
tags: [tutorial, database, optimization, postgresql, redis, caching]
---

# Tutorial 3: 데이터베이스 성능 최적화

이 튜토리얼에서는 실무에서 마주치는 데이터베이스 성능 문제를 해결합니다. N+1 쿼리 문제, 인덱스 설계, 쿼리 최적화, 캐싱 전략을 배우고 API 응답 속도를 극적으로 개선합니다.

## 🎯 학습 목표

이 튜토리얼을 완료하면 다음을 할 수 있습니다:

- ✅ N+1 query problem 이해하고 해결하기
- ✅ 효과적인 인덱스 전략 수립하기
- ✅ SQLAlchemy Eager Loading으로 쿼리 최적화하기
- ✅ Connection Pooling 설정하기
- ✅ Redis로 캐싱 전략 구현하기
- ✅ 쿼리 성능 모니터링 및 분석하기
- ✅ Alfred Research Strategies로 최신 최적화 기법 학습하기

## 📋 사전 요구사항

### 필수 설치

- **Python 3.11+**
- **PostgreSQL 14+**
- **Redis 7+**
- **MoAI-ADK v0.23.0+**
- **Tutorial 1, 2 완료** (REST API, 인증)

### 선행 지식

- SQL 기본 (JOIN, WHERE, INDEX)
- SQLAlchemy ORM
- 데이터베이스 정규화
- HTTP 캐싱 헤더

### 환경 준비

```bash
# PostgreSQL 설치 (macOS)
brew install postgresql@14
brew services start postgresql@14

# Redis 설치
brew install redis
brew services start redis

# 프로젝트 디렉토리
mkdir db-optimization-tutorial
cd db-optimization-tutorial
moai-adk init
```

## 🎯 성능 문제 시나리오

블로그 시스템을 예제로 사용합니다:

- **Post**: 게시글 (제목, 내용, 작성자)
- **Comment**: 댓글 (내용, 작성자, 게시글 FK)
- **User**: 사용자 (이름, 이메일)

### 성능 목표

| 엔드포인트 | 현재 | 목표 | 개선율 |
|-----------|------|------|--------|
| GET /posts | 2,500ms | 50ms | **98%** |
| GET /posts/{id} | 800ms | 30ms | **96%** |
| GET /users/{id}/posts | 3,200ms | 80ms | **97%** |

## 🚀 프로젝트 구조

```
db-optimization-tutorial/
├── .moai/
│   └── specs/
│       └── SPEC-DB-OPT-001.md
├── src/
│   └── blog_api/
│       ├── __init__.py
│       ├── main.py
│       ├── config.py
│       ├── database.py          # DB 모델 및 세션
│       ├── models.py            # Pydantic 모델
│       ├── cache.py             # Redis 캐싱
│       ├── monitoring.py        # 성능 모니터링
│       └── routes/
│           ├── posts.py
│           └── users.py
├── tests/
│   ├── test_performance.py
│   └── test_optimization.py
├── benchmarks/
│   └── load_test.py             # 부하 테스트
├── requirements.txt
└── README.md
```

## 단계별 실습

### Step 1: SPEC 작성

```bash
/alfred:1-plan "블로그 API 성능 최적화"
```

**생성된 SPEC** (`.moai/specs/SPEC-DB-OPT-001.md`):

```markdown
# SPEC-DB-OPT-001: 블로그 API 성능 최적화

## 성능 요구사항

### 응답 시간 목표

- PR-001: 게시글 목록 조회 < 50ms (현재: 2,500ms)
- PR-002: 단일 게시글 조회 < 30ms (현재: 800ms)
- PR-003: 사용자 게시글 조회 < 80ms (현재: 3,200ms)

### 최적화 전략

- OPT-001: N+1 쿼리 문제 해결 (Eager Loading)
- OPT-002: 적절한 인덱스 생성
- OPT-003: Connection Pooling 설정
- OPT-004: Redis 캐싱 도입
- OPT-005: 쿼리 성능 모니터링

### 데이터 모델

User:
- id: int (PK)
- name: str (index)
- email: str (unique index)
- created_at: datetime

Post:
- id: int (PK)
- title: str (index)
- content: text
- author_id: int (FK, index)
- created_at: datetime (index)

Comment:
- id: int (PK)
- content: text
- author_id: int (FK, index)
- post_id: int (FK, index)
- created_at: datetime
```

### Step 2: 환경 설정

**requirements.txt**:
```txt
fastapi==0.104.1
uvicorn[standard]==0.24.0
sqlalchemy==2.0.23
psycopg2-binary==2.9.9
redis==5.0.1
hiredis==2.3.2
pydantic==2.5.0
pydantic-settings==2.1.0
locust==2.19.1  # 부하 테스트
pytest==7.4.3
pytest-benchmark==4.0.0
```

**.env**:
```env
DATABASE_URL=postgresql://user:password@localhost/blog_db
REDIS_URL=redis://localhost:6379/0

# Connection Pool 설정
DB_POOL_SIZE=20
DB_MAX_OVERFLOW=10

# 캐싱 설정
CACHE_TTL=300  # 5분
ENABLE_CACHE=true
```

### Step 3: 데이터베이스 모델 (최적화 전)

**src/blog_api/database.py**:

```python
"""
데이터베이스 모델
"""
from datetime import datetime
from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey, create_engine
from sqlalchemy.orm import declarative_base, relationship, Session, sessionmaker
from .config import settings

Base = declarative_base()


class User(Base):
    """사용자 모델"""
    __tablename__ = "users"

    id = Column(Integer, primary_key=True)
    name = Column(String(100), nullable=False, index=True)
    email = Column(String(255), unique=True, nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow, index=True)

    # 관계 (BEFORE: lazy loading - N+1 문제 발생)
    posts = relationship("Post", back_populates="author", lazy="select")
    comments = relationship("Comment", back_populates="author", lazy="select")


class Post(Base):
    """게시글 모델"""
    __tablename__ = "posts"

    id = Column(Integer, primary_key=True)
    title = Column(String(200), nullable=False, index=True)
    content = Column(Text, nullable=False)
    author_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow, index=True)

    # 관계
    author = relationship("User", back_populates="posts", lazy="select")
    comments = relationship("Comment", back_populates="post", lazy="select", cascade="all, delete-orphan")


class Comment(Base):
    """댓글 모델"""
    __tablename__ = "comments"

    id = Column(Integer, primary_key=True)
    content = Column(Text, nullable=False)
    author_id = Column(Integer, ForeignKey("users.id"), nullable=False, index=True)
    post_id = Column(Integer, ForeignKey("posts.id"), nullable=False, index=True)
    created_at = Column(DateTime, default=datetime.utcnow)

    # 관계
    author = relationship("User", back_populates="comments", lazy="select")
    post = relationship("Post", back_populates="comments", lazy="select")


# 데이터베이스 엔진 (BEFORE: Connection Pool 미설정)
engine = create_engine(
    settings.DATABASE_URL,
    echo=settings.DEBUG,  # 쿼리 로깅
)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def get_db():
    """데이터베이스 세션"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def init_db():
    """테이블 생성"""
    Base.metadata.create_all(bind=engine)
```

### Step 4: 문제 상황 재현 (N+1 Query)

**src/blog_api/routes/posts.py** (최적화 전):

```python
"""
게시글 라우트 (최적화 전)
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from ..database import Post, get_db
from ..models import PostResponse

router = APIRouter(prefix="/posts", tags=["posts"])


@router.get("/", response_model=list[PostResponse])
def get_posts(db: Session = Depends(get_db)):
    """
    게시글 목록 조회 (최적화 전)

    ⚠️ 문제: N+1 쿼리 발생
    - 1개 쿼리: 게시글 목록 조회
    - N개 쿼리: 각 게시글의 작성자 조회 (lazy loading)
    - N개 쿼리: 각 게시글의 댓글 조회

    100개 게시글 = 1 + 100 + 100 = 201개 쿼리! 🔥
    """
    posts = db.query(Post).all()

    # 여기서 N+1 문제 발생
    # post.author 접근 시마다 추가 쿼리 실행
    result = []
    for post in posts:
        result.append({
            "id": post.id,
            "title": post.title,
            "author_name": post.author.name,  # +1 쿼리
            "comment_count": len(post.comments)  # +1 쿼리
        })

    return result
```

**성능 측정**:

```python
# 100개 게시글 로드 시간
# BEFORE: 2,500ms (201 queries)
# 목표: 50ms (1-3 queries)
```

### Step 5: N+1 문제 해결 (Eager Loading)

**src/blog_api/routes/posts.py** (최적화 후):

```python
"""
게시글 라우트 (최적화 후)
"""
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session, selectinload, joinedload
from sqlalchemy import func
from ..database import Post, Comment, get_db
from ..models import PostResponse, PostListResponse

router = APIRouter(prefix="/posts", tags=["posts"])


@router.get("/", response_model=PostListResponse)
def get_posts_optimized(
    limit: int = 20,
    offset: int = 0,
    db: Session = Depends(get_db)
):
    """
    게시글 목록 조회 (최적화 후)

    ✅ 해결: Eager Loading + Subquery
    - selectinload: 별도 쿼리로 관련 데이터 한 번에 로드
    - joinedload: JOIN으로 한 번에 로드
    - 총 2-3개 쿼리로 감소
    """
    # 방법 1: selectinload (1:N 관계에 적합)
    posts = (
        db.query(Post)
        .options(
            selectinload(Post.author),  # 작성자 정보 eager load
            selectinload(Post.comments)  # 댓글 eager load
        )
        .order_by(Post.created_at.desc())
        .limit(limit)
        .offset(offset)
        .all()
    )

    # 방법 2: 댓글 수를 subquery로 계산 (더 효율적)
    # comment_counts = (
    #     db.query(
    #         Comment.post_id,
    #         func.count(Comment.id).label("count")
    #     )
    #     .group_by(Comment.post_id)
    #     .subquery()
    # )
    #
    # posts = (
    #     db.query(Post, comment_counts.c.count)
    #     .outerjoin(comment_counts, Post.id == comment_counts.c.post_id)
    #     .options(joinedload(Post.author))
    #     .all()
    # )

    total = db.query(func.count(Post.id)).scalar()

    return PostListResponse(
        posts=[
            PostResponse(
                id=post.id,
                title=post.title,
                content=post.content[:200],  # 미리보기
                author_name=post.author.name,
                comment_count=len(post.comments),
                created_at=post.created_at
            )
            for post in posts
        ],
        total=total,
        limit=limit,
        offset=offset
    )


@router.get("/{post_id}", response_model=PostResponse)
def get_post_optimized(post_id: int, db: Session = Depends(get_db)):
    """
    단일 게시글 조회 (최적화)

    ✅ joinedload로 한 번에 로드
    """
    post = (
        db.query(Post)
        .options(
            joinedload(Post.author),
            selectinload(Post.comments).selectinload(Comment.author)
        )
        .filter(Post.id == post_id)
        .first()
    )

    if not post:
        raise HTTPException(status_code=404, detail="Post not found")

    return post
```

**성능 개선 결과**:

```python
# BEFORE: 201 queries, 2,500ms
# AFTER: 2 queries, 45ms
# 개선율: 98.2% ⚡
```

### Step 6: Connection Pooling 최적화

**src/blog_api/database.py** (개선):

```python
"""
Connection Pool 최적화
"""
from sqlalchemy import create_engine
from sqlalchemy.pool import QueuePool

# BEFORE: 기본 설정
# engine = create_engine(settings.DATABASE_URL)

# AFTER: Connection Pool 최적화
engine = create_engine(
    settings.DATABASE_URL,
    poolclass=QueuePool,
    pool_size=20,          # 기본 연결 수
    max_overflow=10,       # 추가 생성 가능 연결 수
    pool_timeout=30,       # 연결 대기 시간 (초)
    pool_recycle=3600,     # 연결 재생성 주기 (1시간)
    pool_pre_ping=True,    # 연결 상태 확인
    echo=False,            # 프로덕션에서는 False
)

# 연결 풀 상태 모니터링
def get_pool_status():
    """Connection Pool 상태 확인"""
    pool = engine.pool
    return {
        "size": pool.size(),
        "checked_in": pool.checkedin(),
        "checked_out": pool.checkedout(),
        "overflow": pool.overflow(),
        "total": pool.size() + pool.overflow()
    }
```

### Step 7: Redis 캐싱 구현

**src/blog_api/cache.py**:

```python
"""
Redis 캐싱
"""
import json
from typing import Optional, Any
from functools import wraps
import redis
from .config import settings

# Redis 클라이언트
redis_client = redis.from_url(
    settings.REDIS_URL,
    encoding="utf-8",
    decode_responses=True
)


def cache_key(*args, **kwargs) -> str:
    """캐시 키 생성"""
    key_parts = [str(arg) for arg in args]
    key_parts.extend(f"{k}:{v}" for k, v in sorted(kwargs.items()))
    return ":".join(key_parts)


def cached(prefix: str, ttl: int = 300):
    """
    캐싱 데코레이터

    Args:
        prefix: 캐시 키 접두사
        ttl: Time To Live (초)

    Usage:
        @cached("posts", ttl=300)
        def get_posts():
            return expensive_query()
    """
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            if not settings.ENABLE_CACHE:
                return func(*args, **kwargs)

            # 캐시 키 생성
            key = f"{prefix}:{cache_key(*args, **kwargs)}"

            # 캐시 확인
            cached_value = redis_client.get(key)
            if cached_value:
                return json.loads(cached_value)

            # 캐시 미스: 함수 실행
            result = func(*args, **kwargs)

            # 캐시 저장
            redis_client.setex(
                key,
                ttl,
                json.dumps(result, default=str)
            )

            return result

        # 캐시 무효화 함수 추가
        def invalidate(*args, **kwargs):
            key = f"{prefix}:{cache_key(*args, **kwargs)}"
            redis_client.delete(key)

        wrapper.invalidate = invalidate
        return wrapper

    return decorator


class CacheManager:
    """캐시 관리 클래스"""

    @staticmethod
    def invalidate_pattern(pattern: str):
        """패턴에 맞는 모든 캐시 무효화"""
        keys = redis_client.keys(pattern)
        if keys:
            redis_client.delete(*keys)

    @staticmethod
    def clear_all():
        """모든 캐시 삭제"""
        redis_client.flushdb()

    @staticmethod
    def get_stats() -> dict:
        """캐시 통계"""
        info = redis_client.info("stats")
        return {
            "hits": info.get("keyspace_hits", 0),
            "misses": info.get("keyspace_misses", 0),
            "hit_rate": info.get("keyspace_hits", 0) / max(
                info.get("keyspace_hits", 0) + info.get("keyspace_misses", 0), 1
            ) * 100
        }
```

**캐싱 적용**:

```python
"""
캐싱이 적용된 라우트
"""
from .cache import cached, CacheManager


@router.get("/posts/", response_model=PostListResponse)
@cached("posts:list", ttl=300)  # 5분 캐싱
def get_posts_cached(
    limit: int = 20,
    offset: int = 0,
    db: Session = Depends(get_db)
):
    """
    게시글 목록 조회 (캐싱 적용)

    ✅ Redis 캐싱으로 DB 부하 감소
    - 첫 요청: DB 쿼리 실행, Redis에 저장
    - 이후 5분: Redis에서 직접 반환
    - 응답 시간: 45ms → 2ms (95% 개선)
    """
    posts = (
        db.query(Post)
        .options(selectinload(Post.author))
        .order_by(Post.created_at.desc())
        .limit(limit)
        .offset(offset)
        .all()
    )

    total = db.query(func.count(Post.id)).scalar()

    return {
        "posts": [serialize_post(p) for p in posts],
        "total": total
    }


@router.post("/posts/", response_model=PostResponse)
def create_post(post_data: PostCreate, db: Session = Depends(get_db)):
    """
    게시글 생성

    ✅ 생성 후 캐시 무효화
    """
    post = Post(**post_data.dict())
    db.add(post)
    db.commit()

    # 목록 캐시 무효화
    CacheManager.invalidate_pattern("posts:list:*")

    return post
```

### Step 8: 인덱스 최적화

**migration.sql**:

```sql
-- 인덱스 전략

-- 1. 단일 컬럼 인덱스 (이미 생성됨)
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_created_at ON posts(created_at);

-- 2. 복합 인덱스 (여러 컬럼 조합)
-- 게시글 목록 조회 최적화 (author_id + created_at)
CREATE INDEX idx_posts_author_created ON posts(author_id, created_at DESC);

-- 댓글 조회 최적화 (post_id + created_at)
CREATE INDEX idx_comments_post_created ON comments(post_id, created_at);

-- 3. 부분 인덱스 (조건부 인덱스)
-- 활성 사용자만 인덱싱
CREATE INDEX idx_users_active_email ON users(email) WHERE is_active = true;

-- 4. 전문 검색 인덱스 (PostgreSQL Full-Text Search)
-- 게시글 제목 + 내용 검색
CREATE INDEX idx_posts_search ON posts USING GIN(
    to_tsvector('english', title || ' ' || content)
);

-- 인덱스 사용 확인 쿼리
-- EXPLAIN ANALYZE SELECT * FROM posts WHERE author_id = 1 ORDER BY created_at DESC;
```

**인덱스 성능 분석**:

```python
"""
인덱스 성능 분석
"""
from sqlalchemy import text


def analyze_query_performance(db: Session, query: str):
    """
    쿼리 실행 계획 분석

    EXPLAIN ANALYZE 결과를 반환
    """
    explain_query = f"EXPLAIN ANALYZE {query}"
    result = db.execute(text(explain_query))

    return [row[0] for row in result]


# 사용 예제
analysis = analyze_query_performance(
    db,
    "SELECT * FROM posts WHERE author_id = 1 ORDER BY created_at DESC LIMIT 20"
)

# 결과:
# BEFORE (인덱스 없음):
# Seq Scan on posts  (cost=0.00..1234.56 rows=100 width=123) (actual time=125.123..250.456 rows=100 loops=1)
# Planning Time: 0.123 ms
# Execution Time: 250.789 ms

# AFTER (복합 인덱스):
# Index Scan using idx_posts_author_created on posts  (cost=0.29..8.31 rows=1 width=123) (actual time=0.015..0.018 rows=1 loops=1)
# Planning Time: 0.054 ms
# Execution Time: 0.032 ms
```

### Step 9: 성능 모니터링

**src/blog_api/monitoring.py**:

```python
"""
성능 모니터링
"""
import time
from functools import wraps
from typing import Callable
from fastapi import Request
import logging

logger = logging.getLogger(__name__)


class PerformanceMonitor:
    """성능 메트릭 수집"""

    def __init__(self):
        self.metrics = {
            "requests": 0,
            "total_time": 0,
            "slow_queries": []
        }

    def record_request(self, path: str, duration: float):
        """요청 기록"""
        self.metrics["requests"] += 1
        self.metrics["total_time"] += duration

        # 느린 쿼리 기록 (100ms 이상)
        if duration > 0.1:
            self.metrics["slow_queries"].append({
                "path": path,
                "duration": duration,
                "timestamp": time.time()
            })

    def get_stats(self) -> dict:
        """통계 반환"""
        return {
            "total_requests": self.metrics["requests"],
            "avg_response_time": (
                self.metrics["total_time"] / self.metrics["requests"]
                if self.metrics["requests"] > 0 else 0
            ),
            "slow_queries_count": len(self.metrics["slow_queries"]),
            "recent_slow_queries": self.metrics["slow_queries"][-10:]
        }


monitor = PerformanceMonitor()


async def performance_middleware(request: Request, call_next):
    """성능 측정 미들웨어"""
    start_time = time.time()

    response = await call_next(request)

    duration = time.time() - start_time
    monitor.record_request(request.url.path, duration)

    # 응답 헤더에 소요 시간 추가
    response.headers["X-Process-Time"] = str(duration)

    # 느린 요청 로깅
    if duration > 0.1:
        logger.warning(
            f"Slow request: {request.method} {request.url.path} "
            f"took {duration:.3f}s"
        )

    return response
```

**main.py에 미들웨어 추가**:

```python
from .monitoring import performance_middleware

app.middleware("http")(performance_middleware)


@app.get("/metrics")
def get_metrics():
    """성능 메트릭 조회"""
    from .monitoring import monitor
    from .cache import CacheManager
    from .database import get_pool_status

    return {
        "performance": monitor.get_stats(),
        "cache": CacheManager.get_stats(),
        "database_pool": get_pool_status()
    }
```

### Step 10: 부하 테스트

**benchmarks/load_test.py** (Locust):

```python
"""
부하 테스트
"""
from locust import HttpUser, task, between


class BlogUser(HttpUser):
    """블로그 사용자 시뮬레이션"""
    wait_time = between(1, 3)

    @task(3)
    def get_posts(self):
        """게시글 목록 조회 (가장 빈번)"""
        self.client.get("/posts/?limit=20&offset=0")

    @task(2)
    def get_post_detail(self):
        """게시글 상세 조회"""
        post_id = 1  # 실제로는 랜덤 ID
        self.client.get(f"/posts/{post_id}")

    @task(1)
    def get_user_posts(self):
        """사용자 게시글 조회"""
        user_id = 1
        self.client.get(f"/users/{user_id}/posts")


# 실행:
# locust -f benchmarks/load_test.py --host=http://localhost:8000
```

**부하 테스트 결과**:

```
# BEFORE 최적화:
# RPS: 45 requests/sec
# 평균 응답: 2,200ms
# P95: 4,500ms
# 에러율: 12%

# AFTER 최적화:
# RPS: 1,250 requests/sec (27배 증가)
# 평균 응답: 35ms (98% 개선)
# P95: 120ms (97% 개선)
# 에러율: 0.1%
```

## ✅ 최적화 체크리스트

### 쿼리 최적화

- ✅ N+1 쿼리 제거 (selectinload, joinedload)
- ✅ 필요한 컬럼만 조회 (defer, load_only)
- ✅ 페이지네이션 적용
- ✅ 서브쿼리로 집계 최적화

### 인덱스 전략

- ✅ 자주 조회되는 컬럼에 인덱스
- ✅ 복합 인덱스 (WHERE + ORDER BY)
- ✅ 부분 인덱스 (조건부)
- ✅ EXPLAIN ANALYZE로 검증

### 캐싱

- ✅ Redis 캐싱 도입
- ✅ 적절한 TTL 설정
- ✅ 캐시 무효화 전략
- ✅ 캐시 히트율 모니터링

### Connection Pool

- ✅ Pool size 최적화
- ✅ Connection 재사용
- ✅ Timeout 설정
- ✅ Pool 상태 모니터링

## 🔧 문제 해결

### 문제 1: Too many connections

**증상**:
```
sqlalchemy.exc.OperationalError: (psycopg2.OperationalError) FATAL: sorry, too many clients already
```

**해결**:
```python
# Connection Pool 크기 줄이기
pool_size=10
max_overflow=5

# PostgreSQL max_connections 증가
# postgresql.conf:
# max_connections = 200
```

### 문제 2: Redis 연결 실패

**증상**:
```
redis.exceptions.ConnectionError: Error connecting to Redis
```

**해결**:
```bash
# Redis 서비스 확인
brew services list
brew services start redis

# 또는 Docker
docker run -d -p 6379:6379 redis:7-alpine
```

### 문제 3: 캐시 일관성 문제

**증상**: 데이터 수정 후 이전 데이터가 표시됨

**해결**:
```python
# 데이터 수정 시 캐시 무효화
@router.put("/posts/{post_id}")
def update_post(post_id: int, data: PostUpdate, db: Session = Depends(get_db)):
    post = db.query(Post).filter(Post.id == post_id).first()
    # ... 업데이트 로직

    # 관련 캐시 모두 무효화
    CacheManager.invalidate_pattern(f"posts:*:{post_id}")
    CacheManager.invalidate_pattern("posts:list:*")

    return post
```

## 💡 Best Practices

### 1. 항상 EXPLAIN ANALYZE 사용

```python
# 느린 쿼리 찾기
EXPLAIN ANALYZE SELECT * FROM posts WHERE author_id = 1;

# Index Scan vs Seq Scan 확인
# Index Scan = 좋음 ✅
# Seq Scan = 나쁨 (인덱스 필요) ❌
```

### 2. Eager Loading 전략

```python
# 1:1, N:1 관계 → joinedload (JOIN 사용)
query.options(joinedload(Post.author))

# 1:N 관계 → selectinload (별도 쿼리)
query.options(selectinload(Post.comments))

# 깊은 관계 → 체이닝
query.options(
    selectinload(Post.comments).selectinload(Comment.author)
)
```

### 3. 캐싱 전략

```python
# 읽기 많음 → 긴 TTL (1시간)
@cached("posts:popular", ttl=3600)

# 자주 변경 → 짧은 TTL (5분)
@cached("posts:recent", ttl=300)

# 실시간 중요 → 캐싱 안 함
# (주문, 결제 등)
```

### 4. Connection Pool 크기

```python
# 공식: pool_size = (CPU 코어 수 * 2) + 1
# 예: 4코어 → pool_size = 9

# 웹 서버:
pool_size = 20
max_overflow = 10

# 백그라운드 작업:
pool_size = 5
max_overflow = 0
```

## 📊 성능 개선 결과 요약

| 항목 | BEFORE | AFTER | 개선율 |
|------|--------|-------|--------|
| 게시글 목록 | 2,500ms | 45ms | **98.2%** |
| 단일 게시글 | 800ms | 30ms | **96.3%** |
| 사용자 게시글 | 3,200ms | 75ms | **97.7%** |
| 쿼리 수 (목록) | 201 | 2 | **99%** |
| RPS | 45 | 1,250 | **27배** |
| 에러율 | 12% | 0.1% | **99.2%** |

## 🚀 다음 단계

축하합니다! 데이터베이스 성능을 극대화했습니다.

### 추가 최적화

1. **Database Replication**: 읽기/쓰기 분리
2. **Partitioning**: 대용량 테이블 분할
3. **Materialized Views**: 복잡한 집계 미리 계산
4. **CDN**: 정적 콘텐츠 캐싱

### 다음 튜토리얼

- **[Tutorial 4: Supabase BaaS](/ko/tutorials/tutorial-04-baas-supabase)**
  - 백엔드 개발 속도 극대화

## 📚 참고 자료

- [PostgreSQL Performance Tips](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [SQLAlchemy ORM Tutorial](https://docs.sqlalchemy.org/en/20/orm/tutorial.html)
- [Redis Caching Best Practices](https://redis.io/docs/manual/patterns/)

---

**Happy Optimizing! ⚡**
