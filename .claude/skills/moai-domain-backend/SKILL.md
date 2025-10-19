---
name: moai-domain-backend
description: Server architecture, API design, caching strategies, and scalability patterns
allowed-tools:
  - Read
  - Bash
tier: 2
auto-load: "true"
---

# Backend Expert

## What it does

Provides expertise in backend server architecture, RESTful API design, caching strategies, database optimization, and horizontal/vertical scalability patterns.

## When to use

- "백엔드 아키텍처", "API 설계", "캐싱 전략", "확장성", "마이크로서비스", "로드 밸런싱", "데이터베이스 최적화", "메시지 큐", "비동기 처리"
- "Backend architecture", "API design", "Caching strategy", "Scalability", "Microservices", "Load balancing"
- Automatically invoked when working with backend projects
- Backend SPEC implementation (`/alfred:2-run`)

- "백엔드 아키텍처", "API 설계", "캐싱 전략", "확장성"
- Automatically invoked when working with backend projects
- Backend SPEC implementation (`/alfred:2-run`)

## How it works

**Server Architecture**:
- **Layered architecture**: Controller → Service → Repository
- **Microservices**: Service decomposition, inter-service communication
- **Monoliths**: When appropriate (team size, complexity)
- **Serverless**: Functions as a Service (AWS Lambda, Cloud Functions)

**API Design**:
- **RESTful principles**: Resource-based, stateless
- **GraphQL**: Schema-first design
- **gRPC**: Protocol buffers for high performance
- **WebSockets**: Real-time bidirectional communication

**Caching Strategies**:
- **Redis**: In-memory data store
- **Memcached**: Distributed caching
- **Cache invalidation**: TTL, cache-aside pattern
- **CDN caching**: Static asset delivery

**Database Optimization**:
- **Connection pooling**: Reuse connections
- **Query optimization**: EXPLAIN analysis
- **Read replicas**: Horizontal scaling
- **Sharding**: Data partitioning

**Scalability Patterns**:
- **Horizontal scaling**: Load balancing across instances
- **Vertical scaling**: Increasing instance resources
- **Async processing**: Message queues (RabbitMQ, Kafka)
- **Rate limiting**: API throttling

## TDD Workflow for Backend Development

### Phase 1: RED (Test)
```python
# @TEST:BACKEND-001 | SPEC: SPEC-BACKEND-001.md
def test_get_orders_with_caching():
    """캐시 없을 때 DB에서 조회"""
    result = get_orders(user_id=1)
    assert len(result) == 5
    assert result[0]['id'] == 'ORDER-001'

def test_get_orders_cache_hit():
    """캐시 있을 때 즉시 반환"""
    # 1회차: DB 조회
    result1 = get_orders(user_id=1)
    # 2회차: 캐시 조회 (1ms 이내)
    result2 = get_orders(user_id=1)
    assert result1 == result2
```

### Phase 2: GREEN (Implementation)
```python
# @CODE:BACKEND-001 | SPEC: SPEC-BACKEND-001.md | TEST: tests/test_orders.py
import redis
cache = redis.Redis(host='localhost', port=6379)

def get_orders(user_id: int):
    """주문 조회 (캐시 적용)"""
    cache_key = f"orders:{user_id}"

    # 1단계: 캐시 확인
    cached = cache.get(cache_key)
    if cached:
        return json.loads(cached)

    # 2단계: DB 조회
    orders = db.query("SELECT * FROM orders WHERE user_id = %s", user_id)

    # 3단계: 캐시 저장 (1시간)
    cache.setex(cache_key, 3600, json.dumps(orders))

    return orders
```

### Phase 3: REFACTOR (Optimization)
```python
# @CODE:BACKEND-001:REFACTOR | 성능 개선
def get_orders(user_id: int):
    """주문 조회 (Pipeline 최적화)"""
    cache_key = f"orders:{user_id}"

    # Pipeline 사용: 한 번의 왕복으로 캐시 확인 + TTL 갱신
    pipe = cache.pipeline()
    pipe.get(cache_key)
    pipe.expire(cache_key, 3600)  # TTL 갱신
    cached, _ = pipe.execute()

    if cached:
        return json.loads(cached)

    # DB 조회 + 캐시 저장
    orders = db.query("SELECT * FROM orders WHERE user_id = %s", user_id)
    cache.setex(cache_key, 3600, json.dumps(orders))

    return orders

# 개선 효과:
# Before: 2회 Redis 왕복 (GET + EXPIRE)
# After: 1회 Redis 왕복 (Pipeline)
# 성능: 50% 감소 ✅
```

## Examples

### Example 1: Layered Architecture with Dependency Injection

**❌ Before (Tightly Coupled)**:
```python
# @CODE:ARCHITECTURE-001
class OrderService:
    def __init__(self):
        self.db = DatabaseConnection()  # 강한 결합
        self.cache = RedisConnection()

    def get_orders(self, user_id):
        orders = self.db.query("SELECT * FROM orders WHERE user_id = %s", user_id)
        return orders

# 문제: DB나 Cache 변경 시 코드 수정 필요, 테스트 어려움
```

**✅ After (Dependency Injection)**:
```python
# @CODE:ARCHITECTURE-001: DI 패턴
class OrderRepository:
    def __init__(self, db_connection):
        self.db = db_connection

    def find_by_user(self, user_id):
        return self.db.query("SELECT * FROM orders WHERE user_id = %s", user_id)

class CacheLayer:
    def __init__(self, cache, repository):
        self.cache = cache
        self.repo = repository

    def get_orders(self, user_id):
        cache_key = f"orders:{user_id}"
        if cached := self.cache.get(cache_key):
            return json.loads(cached)

        orders = self.repo.find_by_user(user_id)
        self.cache.setex(cache_key, 3600, json.dumps(orders))
        return orders

class OrderService:
    def __init__(self, cache_layer):
        self.cache_layer = cache_layer

    def get_orders(self, user_id):
        return self.cache_layer.get_orders(user_id)

# 테스트:
def test_order_service():
    mock_db = MockDatabase()
    mock_cache = MockCache()
    repo = OrderRepository(mock_db)
    cache_layer = CacheLayer(mock_cache, repo)
    service = OrderService(cache_layer)

    result = service.get_orders(user_id=1)
    assert result == []  # 쉬운 테스트!
```

### Example 2: Horizontal Scaling with Load Balancer

**Architecture**:
```
                    User Requests
                          |
                    [Load Balancer]
                    /      |      \
            Server1    Server2    Server3
                    \      |      /
                    [Shared Cache]
                          |
                    [Database]
```

**Implementation** (NGINX):
```bash
# /etc/nginx/nginx.conf
upstream backend {
    server localhost:8001;
    server localhost:8002;
    server localhost:8003;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;
    }
}
```

### Example 3: Message Queue for Async Processing

**❌ Synchronous (Slow)**:
```go
// @CODE:ASYNC-001: 동기 처리
func ProcessOrder(order Order) {
    // 1단계: 결제 (3초)
    chargeCard(order.CardNumber, order.Amount)

    // 2단계: 이메일 전송 (2초)
    sendEmail(order.Email, "Order confirmed")

    // 3단계: 재고 업데이트 (1초)
    updateInventory(order.Items)

    // 총: 6초 ⏱️
    return order
}
```

**✅ Async (Fast)**:
```go
// @CODE:ASYNC-001: 비동기 처리
func ProcessOrder(order Order) {
    // 1단계: 결제만 동기 처리 (3초)
    chargeCard(order.CardNumber, order.Amount)

    // 2-3단계: 메시지 큐에 발행 (비동기)
    queue.Publish("order.confirmed", order)  // 10ms

    return order  // 3.01초 ✅ (6초 → 3초, 50% 단축!)
}

// 별도 worker에서 처리
func OrderWorker() {
    for msg := range queue.Consume("order.confirmed") {
        order := msg.Order

        // 이메일 전송 (2초)
        sendEmail(order.Email, "Order confirmed")

        // 재고 업데이트 (1초)
        updateInventory(order.Items)
    }
}
```

### Example 4: Database Connection Pooling

**Configuration**:
```yaml
# @CODE:POOL-001: 연결 풀 설정
database:
  max_connections: 20        # 최대 20개 연결
  min_connections: 5         # 최소 5개 유지
  max_idle_time: 300         # 5분 유휴 연결 종료
  connection_timeout: 10000  # 10초 타임아웃

# 효과:
# - 연결 생성 비용 감소
# - 동시성 향상
# - 리소스 효율
```

## Keywords

"백엔드 아키텍처", "API 설계", "캐싱", "확장성", "마이크로서비스", "로드 밸런싱", "비동기 처리", "메시지 큐", "TDD", "dependency injection", "scalability patterns"

## Reference

- Backend architecture guide: `.moai/memory/development-guide.md#백엔드-아키텍처`
- API design patterns: CLAUDE.md#API-설계-패턴
- Caching strategies: `.moai/memory/development-guide.md#캐싱-전략`

## Works well with

- moai-foundation-trust (백엔드 테스트 검증)
- moai-domain-web-api (API 설계)
- moai-domain-database (DB 최적화)
- moai-domain-devops (배포 자동화)
