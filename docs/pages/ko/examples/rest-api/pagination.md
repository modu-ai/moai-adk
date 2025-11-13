---
title: "페이지네이션 & 정렬"
category: "rest-api"
difficulty: "중급"
tags: [fastapi, pagination, sorting, performance, sqlalchemy]
---

# 페이지네이션 & 정렬

## 개요

대용량 데이터를 효율적으로 처리하기 위한 페이지네이션(Pagination)과 정렬(Sorting) 기능을 구현합니다. Offset/Limit 방식과 Cursor 기반 방식을 모두 다룹니다.

## 사용 사례

- 게시판 목록 조회 (페이지 번호)
- 무한 스크롤 (Cursor 기반)
- 검색 결과 정렬
- 대시보드 데이터 테이블

## 완전한 코드 예제

### 1. 스키마 정의 (`schemas.py`)

```python
# SPEC: API-010 - 페이지네이션 스키마

from pydantic import BaseModel, Field
from typing import List, Generic, TypeVar, Optional
from enum import Enum

T = TypeVar('T')

class SortOrder(str, Enum):
    """정렬 순서"""
    ASC = "asc"
    DESC = "desc"

class PaginationParams(BaseModel):
    """페이지네이션 파라미터"""
    skip: int = Field(default=0, ge=0, description="건너뛸 레코드 수")
    limit: int = Field(default=20, ge=1, le=100, description="조회할 레코드 수 (최대 100)")

class SortParams(BaseModel):
    """정렬 파라미터"""
    sort_by: str = Field(default="created_at", description="정렬 기준 필드")
    order: SortOrder = Field(default=SortOrder.DESC, description="정렬 순서")

class PaginatedResponse(BaseModel, Generic[T]):
    """페이지네이션 응답"""
    items: List[T]
    total: int = Field(..., description="전체 레코드 수")
    skip: int = Field(..., description="건너뛴 레코드 수")
    limit: int = Field(..., description="조회한 레코드 수")
    has_more: bool = Field(..., description="다음 페이지 존재 여부")

class CursorPaginationParams(BaseModel):
    """Cursor 기반 페이지네이션 파라미터"""
    cursor: Optional[str] = Field(default=None, description="다음 페이지 커서")
    limit: int = Field(default=20, ge=1, le=100, description="조회할 레코드 수")

class CursorPaginatedResponse(BaseModel, Generic[T]):
    """Cursor 기반 페이지네이션 응답"""
    items: List[T]
    next_cursor: Optional[str] = Field(None, description="다음 페이지 커서")
    has_more: bool = Field(..., description="다음 페이지 존재 여부")
```

### 2. 모델 정의 (`models.py`)

```python
# SPEC: API-011 - 게시글 모델

from sqlalchemy import Column, Integer, String, Text, DateTime, Index
from sqlalchemy.sql import func
from database import Base

class Post(Base):
    """
    게시글 모델

    Attributes:
        id: 기본 키
        title: 제목
        content: 내용
        author: 작성자
        view_count: 조회수
        created_at: 생성 시간
        updated_at: 수정 시간
    """
    __tablename__ = "posts"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(200), nullable=False, index=True)
    content = Column(Text, nullable=False)
    author = Column(String(100), nullable=False, index=True)
    view_count = Column(Integer, default=0, index=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now(), index=True)
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())

    # 복합 인덱스: 정렬 성능 향상
    __table_args__ = (
        Index('idx_created_at_desc', created_at.desc()),
        Index('idx_view_count_desc', view_count.desc()),
    )
```

### 3. CRUD 함수 (`crud.py`)

```python
# SPEC: API-012 - 페이지네이션 CRUD

from sqlalchemy.orm import Session
from sqlalchemy import desc, asc
from typing import List, Tuple, Optional
import base64
import json

import models
import schemas

def get_posts_paginated(
    db: Session,
    pagination: schemas.PaginationParams,
    sort: schemas.SortParams
) -> Tuple[List[models.Post], int]:
    """
    Offset/Limit 기반 페이지네이션

    Args:
        db: 데이터베이스 세션
        pagination: 페이지네이션 파라미터
        sort: 정렬 파라미터

    Returns:
        (게시글 목록, 전체 레코드 수)
    """
    # 정렬 필드 결정
    sort_column = getattr(models.Post, sort.sort_by, models.Post.created_at)

    # 정렬 순서 결정
    if sort.order == schemas.SortOrder.DESC:
        sort_column = desc(sort_column)
    else:
        sort_column = asc(sort_column)

    # 쿼리 생성
    query = db.query(models.Post).order_by(sort_column)

    # 전체 개수 조회
    total = query.count()

    # 페이지네이션 적용
    posts = query.offset(pagination.skip).limit(pagination.limit).all()

    return posts, total

def get_posts_cursor_paginated(
    db: Session,
    cursor_params: schemas.CursorPaginationParams,
    sort_by: str = "created_at"
) -> Tuple[List[models.Post], Optional[str]]:
    """
    Cursor 기반 페이지네이션 (무한 스크롤)

    Args:
        db: 데이터베이스 세션
        cursor_params: Cursor 파라미터
        sort_by: 정렬 기준 필드

    Returns:
        (게시글 목록, 다음 커서)
    """
    query = db.query(models.Post)

    # Cursor 디코딩
    if cursor_params.cursor:
        try:
            cursor_data = json.loads(base64.b64decode(cursor_params.cursor))
            last_id = cursor_data.get("id")
            last_value = cursor_data.get("value")

            # Cursor 이후 데이터 조회
            sort_column = getattr(models.Post, sort_by)
            query = query.filter(
                (sort_column < last_value) |
                ((sort_column == last_value) & (models.Post.id < last_id))
            )
        except Exception:
            # 잘못된 커서는 무시
            pass

    # 정렬 및 조회 (limit + 1로 다음 페이지 존재 여부 확인)
    sort_column = getattr(models.Post, sort_by, models.Post.created_at)
    posts = query.order_by(desc(sort_column), desc(models.Post.id)).limit(
        cursor_params.limit + 1
    ).all()

    # 다음 커서 생성
    next_cursor = None
    has_more = len(posts) > cursor_params.limit

    if has_more:
        posts = posts[:-1]  # 마지막 항목 제거
        last_post = posts[-1]
        cursor_data = {
            "id": last_post.id,
            "value": getattr(last_post, sort_by).isoformat() if hasattr(getattr(last_post, sort_by), 'isoformat') else str(getattr(last_post, sort_by))
        }
        next_cursor = base64.b64encode(json.dumps(cursor_data).encode()).decode()

    return posts, next_cursor
```

### 4. API 엔드포인트 (`main.py`)

```python
# SPEC: API-013 - 페이지네이션 API

from fastapi import FastAPI, Depends, Query
from sqlalchemy.orm import Session
from typing import Annotated

import models
import schemas
import crud
from database import engine, get_db

models.Base.metadata.create_all(bind=engine)

app = FastAPI(title="Pagination API")

@app.get("/posts/", response_model=schemas.PaginatedResponse[schemas.PostResponse])
def list_posts(
    skip: Annotated[int, Query(ge=0)] = 0,
    limit: Annotated[int, Query(ge=1, le=100)] = 20,
    sort_by: str = "created_at",
    order: schemas.SortOrder = schemas.SortOrder.DESC,
    db: Session = Depends(get_db)
):
    """
    게시글 목록 조회 (Offset/Limit 페이지네이션)

    Args:
        skip: 건너뛸 레코드 수
        limit: 조회할 레코드 수 (최대 100)
        sort_by: 정렬 기준 (created_at, view_count, title)
        order: 정렬 순서 (asc, desc)
        db: 데이터베이스 세션

    Returns:
        페이지네이션된 게시글 목록
    """
    # 파라미터 객체 생성
    pagination = schemas.PaginationParams(skip=skip, limit=limit)
    sort = schemas.SortParams(sort_by=sort_by, order=order)

    # 데이터 조회
    posts, total = crud.get_posts_paginated(db, pagination, sort)

    # 응답 생성
    return schemas.PaginatedResponse(
        items=posts,
        total=total,
        skip=skip,
        limit=limit,
        has_more=(skip + limit) < total
    )

@app.get("/posts/cursor/", response_model=schemas.CursorPaginatedResponse[schemas.PostResponse])
def list_posts_cursor(
    cursor: Optional[str] = None,
    limit: Annotated[int, Query(ge=1, le=100)] = 20,
    sort_by: str = "created_at",
    db: Session = Depends(get_db)
):
    """
    게시글 목록 조회 (Cursor 기반 페이지네이션)

    무한 스크롤에 적합한 방식입니다.

    Args:
        cursor: 다음 페이지 커서 (Base64 인코딩)
        limit: 조회할 레코드 수
        sort_by: 정렬 기준 필드
        db: 데이터베이스 세션

    Returns:
        Cursor 기반 페이지네이션된 게시글 목록
    """
    # 파라미터 객체 생성
    cursor_params = schemas.CursorPaginationParams(cursor=cursor, limit=limit)

    # 데이터 조회
    posts, next_cursor = crud.get_posts_cursor_paginated(db, cursor_params, sort_by)

    # 응답 생성
    return schemas.CursorPaginatedResponse(
        items=posts,
        next_cursor=next_cursor,
        has_more=next_cursor is not None
    )

# 테스트 데이터 생성 엔드포인트
@app.post("/posts/seed/")
def seed_posts(count: int = 100, db: Session = Depends(get_db)):
    """테스트 데이터 생성"""
    import random
    from datetime import datetime, timedelta

    for i in range(count):
        post = models.Post(
            title=f"게시글 제목 {i+1}",
            content=f"게시글 내용 {i+1}",
            author=f"작성자{random.randint(1, 10)}",
            view_count=random.randint(0, 1000),
            created_at=datetime.utcnow() - timedelta(days=random.randint(0, 365))
        )
        db.add(post)

    db.commit()
    return {"message": f"{count}개의 게시글이 생성되었습니다"}
```

## 단계별 설명

### 1. Offset/Limit 방식

**장점**:
- 구현이 간단
- 페이지 번호로 이동 가능
- 전체 개수 제공 가능

**단점**:
- 대량 데이터에서 성능 저하
- 데이터 변경 시 중복/누락 가능

**사용 예시**:
```bash
# 1페이지 (0-19)
curl "http://localhost:8000/posts/?skip=0&limit=20"

# 2페이지 (20-39)
curl "http://localhost:8000/posts/?skip=20&limit=20"

# 조회수 내림차순 정렬
curl "http://localhost:8000/posts/?sort_by=view_count&order=desc"
```

### 2. Cursor 기반 방식

**장점**:
- 대량 데이터에서 빠른 성능
- 데이터 변경에도 안정적
- 무한 스크롤에 적합

**단점**:
- 페이지 번호로 이동 불가
- 전체 개수 제공 어려움
- 구현이 복잡

**사용 예시**:
```bash
# 첫 페이지
curl "http://localhost:8000/posts/cursor/?limit=20"

# 다음 페이지 (응답의 next_cursor 사용)
curl "http://localhost:8000/posts/cursor/?cursor=eyJpZCI6MjAsInZhbHVlIjoiMjAyNC0wMS0wMVQwMDowMDowMCJ9&limit=20"
```

## 테스트 코드

```python
# SPEC: TEST-API-010 - 페이지네이션 테스트

import pytest
from fastapi.testclient import TestClient
from main import app
import base64
import json

client = TestClient(app)

@pytest.fixture
def seed_data():
    """테스트 데이터 생성"""
    response = client.post("/posts/seed/?count=50")
    assert response.status_code == 200
    yield
    # 정리 로직 (생략)

def test_offset_pagination(seed_data):
    """Offset/Limit 페이지네이션 테스트"""
    # 첫 페이지
    response = client.get("/posts/?skip=0&limit=10")
    assert response.status_code == 200
    data = response.json()

    assert len(data["items"]) == 10
    assert data["total"] == 50
    assert data["skip"] == 0
    assert data["limit"] == 10
    assert data["has_more"] is True

def test_last_page(seed_data):
    """마지막 페이지 테스트"""
    response = client.get("/posts/?skip=40&limit=10")
    assert response.status_code == 200
    data = response.json()

    assert len(data["items"]) == 10
    assert data["has_more"] is False

def test_sorting_desc(seed_data):
    """내림차순 정렬 테스트"""
    response = client.get("/posts/?sort_by=created_at&order=desc")
    assert response.status_code == 200
    data = response.json()

    items = data["items"]
    # 날짜가 내림차순인지 확인
    for i in range(len(items) - 1):
        assert items[i]["created_at"] >= items[i + 1]["created_at"]

def test_cursor_pagination(seed_data):
    """Cursor 기반 페이지네이션 테스트"""
    # 첫 페이지
    response = client.get("/posts/cursor/?limit=10")
    assert response.status_code == 200
    data = response.json()

    assert len(data["items"]) == 10
    assert data["has_more"] is True
    assert data["next_cursor"] is not None

    # 다음 페이지
    next_cursor = data["next_cursor"]
    response2 = client.get(f"/posts/cursor/?cursor={next_cursor}&limit=10")
    assert response2.status_code == 200
    data2 = response2.json()

    assert len(data2["items"]) == 10
    # 중복 없음
    first_ids = [item["id"] for item in data["items"]]
    second_ids = [item["id"] for item in data2["items"]]
    assert len(set(first_ids) & set(second_ids)) == 0

def test_limit_validation():
    """Limit 검증 테스트"""
    # 최대값 초과
    response = client.get("/posts/?limit=101")
    assert response.status_code == 422

    # 음수
    response = client.get("/posts/?limit=-1")
    assert response.status_code == 422
```

## Best Practices

### 성능 최적화
- ✅ **인덱스 생성**: 정렬/필터링 컬럼에 인덱스 추가
- ✅ **Limit 제한**: 최대 100개로 제한
- ✅ **COUNT 쿼리 최적화**: 필요한 경우만 전체 개수 조회
- ✅ **복합 인덱스**: 정렬 방향 고려 (`idx_created_at_desc`)

### API 설계
- ✅ **일관된 응답 형식**: `PaginatedResponse` 스키마 사용
- ✅ **기본값 제공**: skip=0, limit=20
- ✅ **메타데이터 포함**: total, has_more
- ✅ **정렬 옵션**: sort_by, order 파라미터

### 사용자 경험
- ✅ **페이지 번호 계산**: `page = skip // limit + 1`
- ✅ **다음 페이지 여부**: `has_more` 플래그
- ✅ **유효성 검증**: Pydantic으로 파라미터 검증

## 주의사항

### Offset/Limit 방식
- ⚠️ **대량 데이터**: 100만 건 이상에서 성능 저하
- ⚠️ **데이터 변경**: 조회 중 데이터 추가/삭제 시 중복/누락
- ⚠️ **깊은 페이지**: skip 값이 클수록 느림

### Cursor 방식
- ⚠️ **복잡한 정렬**: 다중 컬럼 정렬 시 커서 복잡도 증가
- ⚠️ **페이지 이동**: 특정 페이지로 바로 이동 불가
- ⚠️ **커서 보안**: Base64는 암호화 아님 (민감 정보 제외)

### 공통 주의사항
- ❌ **무제한 조회 금지**: 반드시 limit 설정
- ❌ **전체 COUNT 남용**: 대량 데이터에서 느림
- ❌ **정렬 필드 미검증**: SQL 인젝션 위험

## 관련 예제

- [기본 CRUD 작업](/ko/examples/rest-api/basic-crud) - CRUD 기초
- [필터링 & 검색](/ko/examples/rest-api/filtering) - 동적 쿼리
- [쿼리 최적화](/ko/examples/database/query-optimization) - 성능 향상
- [Redis 캐싱](/ko/examples/performance/caching) - 응답 속도 개선

## 참고 자료

- [FastAPI 공식 문서 - Query Parameters](https://fastapi.tiangolo.com/tutorial/query-params/)
- [SQLAlchemy - Ordering](https://docs.sqlalchemy.org/en/20/tutorial/data_select.html#order-by)
- [Cursor Pagination 설명](https://jsonapi.org/profiles/ethanresnick/cursor-pagination/)
