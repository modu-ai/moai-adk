---
title: "필터링 & 검색"
category: "rest-api"
difficulty: "중급"
tags: [fastapi, filtering, search, sqlalchemy, query]
---

# 필터링 & 검색

## 개요

쿼리 파라미터를 사용한 동적 필터링과 전문 검색 기능을 구현합니다. 사용자가 원하는 조건으로 데이터를 효율적으로 찾을 수 있습니다.

## 사용 사례

- 상품 검색 (가격, 카테고리, 브랜드)
- 사용자 필터링 (상태, 등록일, 권한)
- 로그 조회 (날짜 범위, 레벨, 메시지)
- 게시글 검색 (제목, 내용, 작성자)

## 완전한 코드 예제

### 1. 필터 스키마 (`schemas.py`)

```python
# SPEC: API-020 - 필터링 스키마

from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime
from enum import Enum

class ProductStatus(str, Enum):
    """상품 상태"""
    ACTIVE = "active"
    INACTIVE = "inactive"
    OUT_OF_STOCK = "out_of_stock"

class ProductFilter(BaseModel):
    """상품 필터"""
    # 기본 필터
    name: Optional[str] = Field(None, description="상품명 (부분 일치)")
    category: Optional[str] = Field(None, description="카테고리")
    status: Optional[ProductStatus] = Field(None, description="상태")

    # 가격 범위
    min_price: Optional[float] = Field(None, ge=0, description="최소 가격")
    max_price: Optional[float] = Field(None, ge=0, description="최대 가격")

    # 날짜 범위
    created_after: Optional[datetime] = Field(None, description="생성일 시작")
    created_before: Optional[datetime] = Field(None, description="생성일 종료")

    # 검색
    search: Optional[str] = Field(None, min_length=2, description="전체 검색어")

class ProductFilterParams(BaseModel):
    """쿼리 파라미터 파싱"""
    filter: ProductFilter = Field(default_factory=ProductFilter)
    skip: int = Field(default=0, ge=0)
    limit: int = Field(default=20, ge=1, le=100)
    sort_by: str = Field(default="created_at")
    order: str = Field(default="desc")
```

### 2. 동적 쿼리 빌더 (`crud.py`)

```python
# SPEC: API-021 - 동적 쿼리 필터링

from sqlalchemy.orm import Session
from sqlalchemy import or_, and_, desc, asc
from typing import List, Tuple
import models
import schemas

def build_product_query(
    db: Session,
    filter_params: schemas.ProductFilter
):
    """
    동적 쿼리 빌더

    Args:
        db: 데이터베이스 세션
        filter_params: 필터 파라미터

    Returns:
        SQLAlchemy 쿼리 객체
    """
    query = db.query(models.Product)

    # 상품명 필터 (부분 일치)
    if filter_params.name:
        query = query.filter(
            models.Product.name.ilike(f"%{filter_params.name}%")
        )

    # 카테고리 필터 (정확히 일치)
    if filter_params.category:
        query = query.filter(
            models.Product.category == filter_params.category
        )

    # 상태 필터
    if filter_params.status:
        query = query.filter(
            models.Product.status == filter_params.status
        )

    # 가격 범위 필터
    if filter_params.min_price is not None:
        query = query.filter(
            models.Product.price >= filter_params.min_price
        )

    if filter_params.max_price is not None:
        query = query.filter(
            models.Product.price <= filter_params.max_price
        )

    # 날짜 범위 필터
    if filter_params.created_after:
        query = query.filter(
            models.Product.created_at >= filter_params.created_after
        )

    if filter_params.created_before:
        query = query.filter(
            models.Product.created_at <= filter_params.created_before
        )

    # 전문 검색 (제목 OR 설명)
    if filter_params.search:
        search_term = f"%{filter_params.search}%"
        query = query.filter(
            or_(
                models.Product.name.ilike(search_term),
                models.Product.description.ilike(search_term)
            )
        )

    return query

def get_filtered_products(
    db: Session,
    filter_params: schemas.ProductFilter,
    skip: int = 0,
    limit: int = 20,
    sort_by: str = "created_at",
    order: str = "desc"
) -> Tuple[List[models.Product], int]:
    """
    필터링된 상품 목록 조회

    Args:
        db: 데이터베이스 세션
        filter_params: 필터 조건
        skip: 건너뛸 레코드 수
        limit: 조회할 레코드 수
        sort_by: 정렬 기준
        order: 정렬 순서

    Returns:
        (상품 목록, 전체 개수)
    """
    # 기본 쿼리 생성
    query = build_product_query(db, filter_params)

    # 정렬
    sort_column = getattr(models.Product, sort_by, models.Product.created_at)
    if order == "desc":
        query = query.order_by(desc(sort_column))
    else:
        query = query.order_by(asc(sort_column))

    # 전체 개수
    total = query.count()

    # 페이지네이션
    products = query.offset(skip).limit(limit).all()

    return products, total
```

### 3. API 엔드포인트 (`main.py`)

```python
# SPEC: API-022 - 필터링 API

from fastapi import FastAPI, Depends, Query
from sqlalchemy.orm import Session
from typing import Optional, Annotated
from datetime import datetime

import models
import schemas
import crud
from database import engine, get_db

models.Base.metadata.create_all(bind=engine)

app = FastAPI(title="Filtering API")

@app.get("/products/", response_model=schemas.PaginatedResponse[schemas.ProductResponse])
def search_products(
    # 기본 필터
    name: Optional[str] = Query(None, description="상품명 검색"),
    category: Optional[str] = Query(None, description="카테고리"),
    status: Optional[schemas.ProductStatus] = Query(None, description="상태"),

    # 가격 범위
    min_price: Optional[float] = Query(None, ge=0, description="최소 가격"),
    max_price: Optional[float] = Query(None, ge=0, description="최대 가격"),

    # 날짜 범위
    created_after: Optional[datetime] = Query(None, description="생성일 시작"),
    created_before: Optional[datetime] = Query(None, description="생성일 종료"),

    # 전문 검색
    search: Optional[str] = Query(None, min_length=2, description="통합 검색어"),

    # 페이지네이션 & 정렬
    skip: Annotated[int, Query(ge=0)] = 0,
    limit: Annotated[int, Query(ge=1, le=100)] = 20,
    sort_by: str = Query("created_at", description="정렬 기준"),
    order: str = Query("desc", pattern="^(asc|desc)$", description="정렬 순서"),

    db: Session = Depends(get_db)
):
    """
    상품 검색 API

    다양한 필터 조건을 조합하여 상품을 검색합니다.

    **필터 조합 예시**:
    - 카테고리가 '전자제품'이고 가격이 10만원~50만원인 상품
    - 이름에 '노트북'이 포함되고 재고 있는 상품
    - 지난 7일간 등록된 모든 상품

    **검색 우선순위**:
    1. search 파라미터가 있으면 전문 검색 수행
    2. 개별 필터 조건들을 AND로 조합
    3. 정렬 후 페이지네이션 적용
    """
    # 필터 객체 생성
    filter_params = schemas.ProductFilter(
        name=name,
        category=category,
        status=status,
        min_price=min_price,
        max_price=max_price,
        created_after=created_after,
        created_before=created_before,
        search=search
    )

    # 필터링 & 조회
    products, total = crud.get_filtered_products(
        db=db,
        filter_params=filter_params,
        skip=skip,
        limit=limit,
        sort_by=sort_by,
        order=order
    )

    return schemas.PaginatedResponse(
        items=products,
        total=total,
        skip=skip,
        limit=limit,
        has_more=(skip + limit) < total
    )

@app.get("/products/advanced/")
def advanced_search(
    # 복잡한 필터 (JSON으로 전달)
    filters: str = Query(..., description="JSON 형식 필터"),
    db: Session = Depends(get_db)
):
    """
    고급 검색 (JSON 필터)

    복잡한 조건을 JSON으로 전달합니다.

    **예시**:
    ```json
    {
        "AND": [
            {"field": "category", "op": "eq", "value": "전자제품"},
            {
                "OR": [
                    {"field": "price", "op": "lt", "value": 100000},
                    {"field": "status", "op": "eq", "value": "sale"}
                ]
            }
        ]
    }
    ```
    """
    import json
    from sqlalchemy import and_, or_

    try:
        filter_json = json.loads(filters)
    except json.JSONDecodeError:
        raise HTTPException(status_code=400, detail="잘못된 JSON 형식입니다")

    # JSON 필터를 SQLAlchemy 조건으로 변환
    # (구현 생략 - 재귀적으로 파싱)

    return {"message": "고급 검색 기능 (구현 예정)"}
```

## 단계별 설명

### 1. 쿼리 파라미터 사용

```bash
# 단일 필터
curl "http://localhost:8000/products/?category=전자제품"

# 복합 필터
curl "http://localhost:8000/products/?category=전자제품&min_price=100000&max_price=500000"

# 검색 + 정렬
curl "http://localhost:8000/products/?search=노트북&sort_by=price&order=asc"

# 날짜 범위
curl "http://localhost:8000/products/?created_after=2024-01-01T00:00:00&created_before=2024-12-31T23:59:59"
```

### 2. 동적 쿼리 생성

```python
# 조건부로 필터 추가
query = db.query(Product)

if name:
    query = query.filter(Product.name.ilike(f"%{name}%"))

if min_price:
    query = query.filter(Product.price >= min_price)

# 최종 실행
results = query.all()
```

### 3. OR 조건 사용

```python
from sqlalchemy import or_

# 이름 또는 설명에서 검색
query = query.filter(
    or_(
        Product.name.ilike(f"%{search}%"),
        Product.description.ilike(f"%{search}%")
    )
)
```

## 테스트 코드

```python
# SPEC: TEST-API-020 - 필터링 테스트

import pytest
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

@pytest.fixture
def sample_products():
    """테스트 상품 데이터"""
    products = [
        {"name": "노트북 A", "category": "전자제품", "price": 1200000, "status": "active"},
        {"name": "노트북 B", "category": "전자제품", "price": 800000, "status": "active"},
        {"name": "의자", "category": "가구", "price": 150000, "status": "active"},
        {"name": "책상", "category": "가구", "price": 300000, "status": "out_of_stock"},
    ]

    for product in products:
        client.post("/products/", json=product)

    yield
    # 정리 로직

def test_filter_by_category(sample_products):
    """카테고리 필터 테스트"""
    response = client.get("/products/?category=전자제품")
    assert response.status_code == 200
    data = response.json()

    assert data["total"] == 2
    assert all(item["category"] == "전자제품" for item in data["items"])

def test_filter_by_price_range(sample_products):
    """가격 범위 필터 테스트"""
    response = client.get("/products/?min_price=500000&max_price=1000000")
    assert response.status_code == 200
    data = response.json()

    for item in data["items"]:
        assert 500000 <= item["price"] <= 1000000

def test_search_by_name(sample_products):
    """이름 검색 테스트"""
    response = client.get("/products/?name=노트북")
    assert response.status_code == 200
    data = response.json()

    assert data["total"] == 2
    assert all("노트북" in item["name"] for item in data["items"])

def test_multiple_filters(sample_products):
    """복합 필터 테스트"""
    response = client.get(
        "/products/?category=전자제품&min_price=1000000&status=active"
    )
    assert response.status_code == 200
    data = response.json()

    assert data["total"] == 1
    assert data["items"][0]["name"] == "노트북 A"

def test_search_no_results(sample_products):
    """결과 없는 검색 테스트"""
    response = client.get("/products/?search=존재하지않는상품")
    assert response.status_code == 200
    data = response.json()

    assert data["total"] == 0
    assert len(data["items"]) == 0
```

## Best Practices

### 쿼리 성능
- ✅ **인덱스 활용**: 필터 컬럼에 인덱스 생성
- ✅ **LIKE 최적화**: 앞쪽 와일드카드(`%term%`) 지양, 뒤쪽만 사용(`term%`)
- ✅ **불필요한 COUNT 제거**: 전체 개수가 필요 없으면 생략
- ✅ **복합 인덱스**: 자주 조합되는 필터는 복합 인덱스

### API 설계
- ✅ **명확한 파라미터명**: `min_price`, `max_price` (명확함)
- ✅ **기본값 제공**: 필터 없으면 전체 조회
- ✅ **검증**: Pydantic으로 타입/범위 검증
- ✅ **문서화**: Query 파라미터에 description 추가

### 보안
- ✅ **SQL 인젝션 방지**: ORM 사용, Raw query 지양
- ✅ **입력 검증**: 최소/최대값, 패턴 검증
- ✅ **허용 필드 제한**: 정렬/필터 가능한 필드 화이트리스트

## 주의사항

### 성능 문제
- ⚠️ **와일드카드 남용**: `%term%`는 인덱스 미사용 (느림)
- ⚠️ **복잡한 OR 조건**: 여러 OR 조건은 성능 저하
- ⚠️ **대량 데이터**: 필터 없이 전체 조회 금지

### 보안 취약점
- ❌ **Raw SQL 사용 금지**: f-string으로 쿼리 생성하지 마세요
- ❌ **검증 없는 정렬**: `sort_by`를 사용자 입력 그대로 사용 금지
- ❌ **무제한 검색**: 최소 검색어 길이 설정 (2-3자 이상)

### 사용자 경험
- ⚠️ **복잡한 필터**: 너무 많은 필터는 혼란
- ⚠️ **결과 없음**: 적절한 안내 메시지 제공
- ⚠️ **필터 초기화**: 필터 해제 기능 제공

## 관련 예제

- [기본 CRUD 작업](/ko/examples/rest-api/basic-crud) - CRUD 기초
- [페이지네이션 & 정렬](/ko/examples/rest-api/pagination) - 대량 데이터 처리
- [쿼리 최적화](/ko/examples/database/query-optimization) - 성능 향상
- [입력 검증](/ko/examples/security/input-validation) - 보안 강화

## 확장 아이디어

### Full-Text Search
```python
# PostgreSQL Full-Text Search
from sqlalchemy import func

query = query.filter(
    func.to_tsvector('korean', Product.name).match(search_term)
)
```

### Elasticsearch 통합
```python
# Elasticsearch로 고급 검색
from elasticsearch import Elasticsearch

es = Elasticsearch()
results = es.search(index="products", body={
    "query": {
        "multi_match": {
            "query": search_term,
            "fields": ["name^2", "description"]
        }
    }
})
```
