---
name: moai-observability-advanced
version: 4.0.0
updated: 2025-11-19
status: stable
category: Monitoring
description: Advanced observability patterns with OpenTelemetry, distributed tracing, eBPF monitoring, SLO/SLI implementation, and real-world production strategies
created: 2025-11-19
tags: [observability, OpenTelemetry, distributed-tracing, metrics, logs, SLO, SLI, eBPF, monitoring, production]
---

# Advanced Observability - Production-Ready Patterns

## Overview

Observability는 시스템의 **내부 상태를 외부 신호(signals)를 통해 이해하는 능력**입니다. 모니터링(monitoring)이 미리 정의된 메트릭을 확인하는 반응형 접근이라면, 관찰성은 예상하지 못한 문제까지 파악할 수 있는 선제적 접근입니다.

### 모니터링 vs 관찰성

| 항목 | 모니터링(Monitoring) | 관찰성(Observability) |
|------|-------------------|-------------------|
| **접근** | 반응형 | 선제적 |
| **범위** | 사전정의 메트릭 | 모든 데이터 |
| **문제해결** | 알려진 문제 | 예상치 못한 문제 |
| **질문** | "시스템이 잘 작동하는가?" | "시스템에서 무슨 일이 일어나고 있는가?" |

### 관찰성의 세 기둥

**Metrics**: 시간에 따른 정량적 측정 (카운터, 게이지, 히스토그램)
**Logs**: 구조화된 텍스트 이벤트 (타임스탬프, 컨텍스트, 메시지)
**Traces**: 분산 시스템의 요청 흐름 추적 (스팬, 부모-자식 관계)

마이크로서비스 환경에서는 **세 기둥이 함께 작동**할 때 시스템의 전체 그림을 볼 수 있습니다:
- Metrics로 성능 추이 감지
- Traces로 병목 지점 식별
- Logs로 근본 원인 파악

### 프로덕션 필요성

대규모 분산 시스템에서는:
- 1000+ 마이크로서비스가 동시 작동
- 요청이 여러 경계를 넘어 흐름
- 네트워크 지연, 부분 장애 발생 가능
- 전통 모니터링으로는 디버깅 불가능

**고급 관찰성**만이 이런 환경을 제어할 수 있습니다.

---

## 1. Observability Fundamentals

### 1.1 관찰성의 세 기둥 상세 분석

#### Metrics (정량 데이터)

메트릭은 **시간에 따른 수치 측정**입니다. 시계열 데이터베이스에 저장되어 빠른 집계와 시각화가 가능합니다.

**메트릭 유형:**

1. **Counter**: 증가만 가능한 누적값 (요청 수, 에러 수)
2. **Gauge**: 증감 가능한 현재값 (CPU 사용률, 메모리, 온도)
3. **Histogram**: 값의 분포 (응답시간, 패킷 크기)
4. **Summary**: Histogram과 유사하지만 quantiles만 저장 (Prometheus)

**사용 예시:**

```yaml
# Example 1: 관찰성 세 기둥 개념 다이어그램
metrics:
  http_requests_total:
    type: counter
    help: "Total HTTP requests"
    labels: [method, endpoint, status]
    example: http_requests_total{method="GET", endpoint="/api/users", status="200"} 1250

  system_memory_bytes:
    type: gauge
    help: "System memory usage in bytes"
    example: system_memory_bytes{type="used"} 8589934592

  http_request_duration_seconds:
    type: histogram
    help: "HTTP request duration in seconds"
    buckets: [0.1, 0.5, 1.0, 2.0, 5.0]
    example: http_request_duration_seconds_bucket{endpoint="/api/users", le="1.0"} 1200

logs:
  structure: "JSON"
  fields: [timestamp, level, message, trace_id, span_id, context]
  example: |
    {
      "timestamp": "2025-11-19T10:30:45Z",
      "level": "ERROR",
      "message": "Database connection failed",
      "trace_id": "abc123def456",
      "span_id": "span789",
      "user_id": "user_001"
    }

traces:
  structure: "Span graph"
  components: [trace_id, span_id, parent_span_id, start_time, duration, attributes]
  example: |
    trace_id: "abc123def456"
    ├─ span_id: "span001" (HTTP request) [0ms - 150ms]
    │  ├─ span_id: "span002" (Database query) [10ms - 100ms]
    │  └─ span_id: "span003" (Cache lookup) [101ms - 105ms]
    └─ span_id: "span004" (Response serialization) [106ms - 150ms]
```

#### Logs (텍스트 이벤트)

로그는 **시스템이 실행되는 동안 발생하는 이벤트의 기록**입니다. 구조화된 형식(JSON)을 사용하면 검색과 분석이 용이합니다.

**로그 수준:**

| 수준 | 사용처 | 예시 |
|------|-------|------|
| **DEBUG** | 개발자 디버깅 | "Processing user request: user_id=123" |
| **INFO** | 일반 정보 | "Server started on port 8000" |
| **WARN** | 경고 (복구 가능) | "Retry attempt 2/3 for database query" |
| **ERROR** | 에러 (서비스 영향) | "Failed to connect to database" |
| **CRITICAL** | 심각한 에러 | "Out of disk space, shutting down" |

**구조화 로깅의 장점:**

```yaml
# Example 2: 메트릭 정의 및 수집 전략 (YAML)
metrics_collection_strategy:
  four_golden_signals:
    latency:
      metric: "http_request_duration_seconds"
      description: "요청 처리 시간"
      target: "< 200ms 95th percentile"
      
    traffic:
      metric: "http_requests_per_second"
      description: "처리량"
      target: "> 1000 RPS"
      
    errors:
      metric: "http_requests_total{status=~'5..'}"
      description: "에러율"
      target: "< 0.1%"
      
    saturation:
      metric: "system_memory_bytes_usage / system_memory_bytes_total"
      description: "리소스 포화도"
      target: "< 80%"

  red_framework:  # Request-based metrics
    rate: "requests per second"
    errors: "failed requests per second"
    duration: "request processing time"

  use_framework:  # Resource-based metrics
    utilization: "percentage used (0-100%)"
    saturation: "queued tasks waiting"
    errors: "errors per operation"
```

#### Traces (분산 추적)

트레이스는 **분산 시스템에서 요청이 여러 서비스를 거쳐가는 경로를 추적**합니다. 마이크로서비스 환경에서는 필수입니다.

**Trace 구조:**

```
Trace ID: abc123 (전체 요청 식별자)
├─ Span 1: HTTP GET /api/users (Client) [0ms ~ 150ms]
│  ├─ Span 2: Gateway validation [5ms ~ 10ms]
│  ├─ Span 3: User Service (gRPC call) [15ms ~ 100ms]
│  │  ├─ Span 4: Database query [20ms ~ 80ms]
│  │  └─ Span 5: Cache lookup [85ms ~ 95ms]
│  ├─ Span 6: Authorization Service (gRPC) [105ms ~ 120ms]
│  └─ Span 7: Response serialization [125ms ~ 150ms]
```

### 1.2 도구 스택 선택

프로덕션 관찰성을 위한 필수 도구들:

```
┌─────────────────────────────────────┐
│       Application Code              │
└──────────────┬──────────────────────┘
               │
   ┌───────────┼───────────┐
   │           │           │
   ▼           ▼           ▼
Metrics    Logs        Traces
   │           │           │
   └───────────┼───────────┘
               │
   ┌───────────┴────────────────────────┐
   │  OpenTelemetry Collector           │
   │  (aggregation, sampling, export)   │
   └──────────────┬─────────────────────┘
               │
   ┌───────────┼───────────┬──────────────┐
   │           │           │              │
   ▼           ▼           ▼              ▼
Prometheus  Jaeger    ELK Stack      Grafana
(metrics)  (traces)  (logs)       (visualization)
```

### 1.3 메트릭 프레임워크: Four Golden Signals

네 가지 핵심 메트릭이 시스템의 건강도를 결정합니다:

1. **Latency (지연시간)**: 요청 처리 시간
2. **Traffic (처리량)**: 초당 처리 요청 수
3. **Errors (에러율)**: 실패한 요청의 비율
4. **Saturation (포화도)**: 리소스 사용률

### 1.4 마이크로서비스 환경에서의 관찰성

**도전과제:**

- 요청이 수십 개 서비스를 거침
- 각 서비스는 다른 팀이 관리
- 부분 장애(partial failure) 처리 필요
- 네트워크 지연 불가피

**해결책:**

- Context propagation: trace_id를 모든 서비스에 전파
- Distributed sampling: 모든 요청을 추적하면 비용 증가
- Service mesh: 관찰성 인프라를 자동화
- Correlation: logs, metrics, traces 연결

---

## 2. OpenTelemetry (OTel) 상세 가이드

### 2.1 OTel 핵심 개념

OpenTelemetry는 **벤더 독립적인 관찰성 표준**입니다. 어떤 백엔드(Jaeger, Datadog, New Relic)로도 내보낼 수 있습니다.

**OTel 아키텍처:**

```
┌─────────────────────────────────┐
│     Application Code            │
│  (OTel Instrumentation)         │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│    OTel SDK (Tracer, Meter)     │
│  (Signal generation)             │
└──────────────┬──────────────────┘
               │
   ┌───────────┼───────────┐
   │           │           │
   ▼           ▼           ▼
Processor  Sampler   Span Processor
   │           │           │
   └───────────┼───────────┘
               │
               ▼
┌─────────────────────────────────┐
│    OTel Exporter                │
│  (OTLP, Jaeger, Prometheus)     │
└──────────────┬──────────────────┘
               │
               ▼
┌─────────────────────────────────┐
│    Backend (Jaeger, Prometheus) │
│    (Storage, Analysis)          │
└─────────────────────────────────┘
```

### 2.2 Automatic Instrumentation (자동 계측)

OTel 에이전트를 사용하면 코드 수정 없이 자동으로 신호를 수집합니다.

**Java Automatic Instrumentation 예시:**

```bash
# Example 4: OTel Java 에이전트 설정

# 1. 에이전트 다운로드
curl -L https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar \
  -o opentelemetry-javaagent.jar

# 2. 애플리케이션 실행 (에이전트 활성화)
java -javaagent:opentelemetry-javaagent.jar \
  -Dotel.service.name=my-app \
  -Dotel.exporter.otlp.endpoint=http://localhost:4317 \
  -Dotel.metrics.exporter=otlp \
  -Dotel.logs.exporter=otlp \
  -Dotel.traces.exporter=otlp \
  -jar app.jar

# 3. 결과: 모든 HTTP 요청, 데이터베이스 쿼리, 외부 API 호출 자동 추적
#    코드 수정 불필요!
```

**장점:**

- 코드 수정 불필요
- 기존 라이브러리 자동 계측 (HTTP, DB, messaging)
- 운영 팀이 적용 가능

**단점:**

- 세밀한 제어 불가능
- 성능 오버헤드 약간 존재
- 비즈니스 로직 추적 어려움

### 2.3 Manual Instrumentation (수동 계측)

비즈니스 로직이나 특수한 경우를 추적해야 할 때 수동으로 계측합니다.

```python
# Example 5: Python OTel 수동 계측

from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource

# 1. Resource 정의 (애플리케이션 식별)
resource = Resource.create({
    "service.name": "payment-service",
    "service.version": "1.0.0",
    "environment": "production"
})

# 2. Tracer Provider 초기화
otlp_exporter = OTLPSpanExporter(endpoint="http://localhost:4317")
tracer_provider = TracerProvider(resource=resource)
tracer_provider.add_span_processor(BatchSpanProcessor(otlp_exporter))
trace.set_tracer_provider(tracer_provider)

# 3. Tracer 획득
tracer = trace.get_tracer(__name__, version="1.0.0")

# 4. 비즈니스 로직에서 수동으로 span 생성
def process_payment(user_id, amount):
    with tracer.start_as_current_span("process_payment") as span:
        span.set_attribute("user_id", user_id)
        span.set_attribute("amount", amount)
        
        # Step 1: 검증
        with tracer.start_as_current_span("validate_payment") as validate_span:
            if amount <= 0:
                validate_span.record_exception(ValueError("Invalid amount"))
                span.set_attribute("payment.status", "FAILED")
                raise ValueError("Invalid amount")
            validate_span.set_attribute("validation.status", "PASSED")
        
        # Step 2: 결제 처리
        with tracer.start_as_current_span("charge_card") as charge_span:
            result = stripe.charge(user_id, amount)
            charge_span.set_attribute("charge.id", result.id)
            charge_span.set_attribute("charge.status", result.status)
        
        # Step 3: 영수증 발급
        with tracer.start_as_current_span("send_receipt") as receipt_span:
            send_email_receipt(user_id, amount)
            receipt_span.set_attribute("receipt.sent", True)
        
        span.set_attribute("payment.status", "SUCCESS")
        return result

# 5. 함수 호출 시 span이 자동으로 기록됨
try:
    payment = process_payment(user_id="user_123", amount=99.99)
except Exception as e:
    print(f"Payment failed: {e}")
```

### 2.4 Sampling 전략

모든 요청을 추적하면 비용이 증가합니다. Sampling으로 비용을 절감합니다.

```yaml
# Example 6: Sampling 전략 구성

sampling_strategies:
  always_on:
    description: "모든 요청 추적"
    use_case: "개발/테스트"
    cost_impact: "매우 높음"
    configuration:
      sampler: always_on

  always_off:
    description: "추적 비활성화"
    use_case: "성능 측정, 헬스체크"
    cost_impact: "낮음"
    configuration:
      sampler: always_off

  probability:
    description: "일정 확률로 추적 (예: 10%)"
    use_case: "프로덕션"
    cost_impact: "중간"
    configuration:
      sampler: probability
      probability: 0.1  # 10%

  parent_based:
    description: "부모 span 존재 시만 추적"
    use_case: "분산 시스템에서 선택적 추적"
    cost_impact: "낮음"
    configuration:
      sampler: parent_based
      root_sampler: probability
      root_probability: 0.1

  jaeger_remote:
    description: "Jaeger 서버에서 동적 샘플링 결정"
    use_case: "프로덕션 (동적 조정)"
    cost_impact: "최적화됨"
    configuration:
      sampler: jaeger_remote
      endpoint: "http://jaeger:14250"
      polling_interval: 30s

# 실제 설정 (Python)
from opentelemetry.sdk.trace.sampling import TraceIdRatioBased

# 프로덕션: 1% 샘플링
sampler = TraceIdRatioBased(rate=0.01)

# 개발: 100% 샘플링
# sampler = AlwaysOnSampler()
```

### 2.5 Exporters (내보내기)

신호를 다양한 백엔드로 내보낼 수 있습니다.

```yaml
# Example 7: 다양한 Exporter 구성

exporters:
  otlp:  # OpenTelemetry Protocol (표준)
    type: "OTLP/gRPC"
    endpoint: "http://localhost:4317"
    timeout: "10s"
    compression: "gzip"
    features:
      - metrics
      - traces
      - logs
    use_case: "표준 프로토콜, 모든 백엔드 지원"

  jaeger:  # Jaeger 전용
    type: "Jaeger"
    endpoint: "http://jaeger:14250"
    max_tag_value_length: 16384
    use_case: "Jaeger만 사용 시 직접 내보내기"

  prometheus:  # Prometheus (메트릭 전용)
    type: "PrometheusRemoteWriteExporter"
    endpoint: "http://prometheus:9090/api/v1/write"
    use_case: "메트릭만 Prometheus로 수집"

  datadog:  # Datadog (타사 서비스)
    type: "Datadog"
    api_key: "${DATADOG_API_KEY}"
    site: "datadoghq.com"
    use_case: "Datadog APM 통합"

  newrelic:  # New Relic
    type: "NewRelic"
    api_key: "${NEW_RELIC_API_KEY}"
    use_case: "New Relic 플랫폼 통합"
```

---

## 3. Distributed Tracing (분산 추적)

### 3.1 Trace 아키텍처

트레이스는 Trace ID로 식별되는 여러 Span의 집합입니다.

```yaml
# Example 8: Trace context propagation 및 구조

trace_structure:
  trace_id: "abc123def456789"  # 전체 요청 고유 식별자
  spans:
    - span_id: "span_001"
      parent_span_id: null
      operation_name: "HTTP GET /api/orders/123"
      service_name: "api-gateway"
      start_time: "2025-11-19T10:30:00.000Z"
      duration_ms: 245
      status: "OK"
      attributes:
        "http.method": "GET"
        "http.url": "/api/orders/123"
        "http.status_code": 200
        "user_id": "user_001"
      
      # 하위 span들
      child_spans:
        - span_id: "span_002"
          parent_span_id: "span_001"
          operation_name: "orders.GetOrder (gRPC)"
          service_name: "orders-service"
          start_time: "2025-11-19T10:30:00.050Z"
          duration_ms: 80
          status: "OK"
          attributes:
            "rpc.method": "orders.GetOrder"
            "rpc.service": "orders.OrderService"
            "order_id": "order_456"
        
        - span_id: "span_003"
          parent_span_id: "span_002"
          operation_name: "db.query (SELECT * FROM orders)"
          service_name: "orders-service"
          start_time: "2025-11-19T10:30:00.055Z"
          duration_ms: 25
          status: "OK"
          attributes:
            "db.system": "postgresql"
            "db.operation": "SELECT"
            "db.rows_affected": 1
        
        - span_id: "span_004"
          parent_span_id: "span_002"
          operation_name: "cache.get (order_456)"
          service_name: "orders-service"
          start_time: "2025-11-19T10:30:00.085Z"
          duration_ms: 5
          status: "OK"
          attributes:
            "cache.backend": "redis"
            "cache.key": "order_456"
            "cache.hit": false

# Context Propagation (W3C Trace Context)
w3c_trace_context:
  header_format: "traceparent"
  example: "traceparent: 00-abc123def456789-span001-01"
  # 00: version
  # abc123def456789: trace_id
  # span001: parent_span_id
  # 01: trace_flags (sampled)

baggage:
  description: "span 간 공유 데이터"
  example:
    "baggage: user_id=user_001, region=us-west-2"
```

### 3.2 Jaeger 배포 및 운영

Jaeger는 분산 추적의 가장 인기 있는 오픈소스 솔루션입니다.

```yaml
# Example 9: Docker Compose로 Jaeger 배포

version: '3.8'
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: jaeger
    ports:
      # Jaeger UI
      - "16686:16686"
      # OTLP gRPC receiver
      - "4317:4317"
      # OTLP HTTP receiver
      - "4318:4318"
      # Jaeger native Thrift
      - "6831:6831/udp"
      # Jaeger native Thrift HTTP
      - "14268:14268"
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
      COLLECTOR_ZIPKIN_HOST_PORT: ":9411"
      MEMORY_MAX_TRACES: "10000"
    command:
      - "--sampling.strategies-file=/etc/jaeger/sampling_strategies.json"
    volumes:
      - ./sampling_strategies.json:/etc/jaeger/sampling_strategies.json
    networks:
      - observability

  # 프로덕션용: Elasticsearch 백엔드
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.14.0
    container_name: elasticsearch
    environment:
      discovery.type: "single-node"
      ES_JAVA_OPTS: "-Xms512m -Xmx512m"
      xpack.security.enabled: "false"
    ports:
      - "9200:9200"
    networks:
      - observability

  jaeger-prod:
    image: jaegertracing/jaeger-collector:latest
    container_name: jaeger-collector
    ports:
      - "14269:14269"
      - "4317:4317"
      - "14268:14268"
    environment:
      SPAN_STORAGE_TYPE: "elasticsearch"
      ES_SERVER_URLS: "http://elasticsearch:9200"
    depends_on:
      - elasticsearch
    networks:
      - observability

networks:
  observability:
    driver: bridge
```

### 3.3 Trace 분석 및 병목 지점 찾기

```python
# Example 10: Trace 분석으로 성능 문제 진단

# Jaeger Query (API)
# 지연시간 > 200ms인 모든 trace 찾기
curl "http://jaeger:16686/api/traces?service=api-gateway&maxDuration=200ms&minDuration=200ms"

# 응답: 다음과 같은 trace 찾음
# {
#   "data": [{
#     "traceID": "abc123...",
#     "spans": [
#       {"operationName": "HTTP GET /api/orders", "duration": 245},
#       {"operationName": "orders.GetOrder", "duration": 80},
#       {"operationName": "db.query", "duration": 25},
#       {"operationName": "cache.get", "duration": 5},
#       {"operationName": "payment.Charge", "duration": 130},  # 병목!
#       {"operationName": "send_receipt", "duration": 5}
#     ]
#   }]
# }

# 분석 결과:
# 1. payment.Charge (130ms) - 가장 오래 걸림
# 2. db.query (25ms) - 빠름
# 3. cache.get (5ms) - 캐시 미스 감지 (문제 없음)

# 처방:
# - payment.Charge 함수 프로파일링
# - Stripe API 지연 확인
# - 내부 로직 최적화 검토
```

### 3.4 Context Propagation

마이크로서비스 간 trace context를 전파합니다.

```python
# Example 11: Context Propagation 구현 (Python)

from opentelemetry import trace, baggage
from opentelemetry.propagate import inject, extract
import requests

tracer = trace.get_tracer(__name__)

def call_downstream_service(service_url, data):
    """
    현재 context를 다운스트림 서비스에 전파
    """
    with tracer.start_as_current_span("call_downstream") as span:
        span.set_attribute("downstream.service", service_url)
        
        # 현재 context를 HTTP headers에 주입
        headers = {}
        inject(headers)  # traceparent, tracestate 등이 추가됨
        
        # 비즈니스 데이터와 함께 전송
        response = requests.post(
            service_url,
            json=data,
            headers=headers
        )
        
        span.set_attribute("http.status_code", response.status_code)
        return response.json()

# 다운스트림 서비스 (Flask)
from flask import Flask, request
from opentelemetry.propagate import extract

app = Flask(__name__)

@app.route("/process", methods=["POST"])
def process():
    """
    업스트림 서비스로부터 context 수신 및 복원
    """
    # HTTP headers에서 context 추출
    ctx = extract(request.headers)
    
    with tracer.start_as_current_span("process_request", context=ctx) as span:
        data = request.get_json()
        span.set_attribute("request.data", str(data))
        
        # 이 span은 자동으로 업스트림의 span과 연결됨!
        result = do_processing(data)
        
        return {"result": result}
```

### 3.5 Trace 성능 고려사항

```yaml
# Example 12: Trace 오버헤드 분석 및 최적화

performance_considerations:
  cpu_overhead:
    trace_creation: "~1-3%"
    span_processor: "~0.5-1%"
    exporter: "~0.5%"
    total: "~2-4.5% with sampling"

  memory_overhead:
    span_buffer: "~10KB per 1000 spans"
    exporter_queue: "~5MB default"
    gc_impact: "minimal with batch processing"

  network_overhead:
    otlp_batch_size: "512 spans per batch"
    compression: "gzip reduces by 90%"
    latency_impact: "< 1ms with async export"

  optimization_strategies:
    sampling:
      description: "가장 효과적인 방법"
      example: "1% sampling = 100배 비용 절감"
      implementation: "TraceIdRatioBased(0.01)"

    batch_processing:
      description: "배치로 span 내보내기"
      batch_size: 512
      export_interval: "5s"
      impact: "10배 네트워크 효율성"

    filtering:
      description: "특정 span 제외"
      example: "health check span 제외"

    compression:
      description: "OTLP 압축"
      impact: "90% 네트워크 트래픽 감소"
```

---

## 4. Metrics & Time Series Database

### 4.1 Prometheus 메트릭 유형

```python
# Example 13: 커스텀 메트릭 정의 및 수집 (Python)

from opentelemetry import metrics
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.prometheus import PrometheusMetricReader
from prometheus_client import start_http_server

# 1. Prometheus reader 초기화
reader = PrometheusMetricReader()
meter_provider = MeterProvider(metric_readers=[reader])
metrics.set_meter_provider(meter_provider)

# 2. Meter 획득
meter = metrics.get_meter(__name__, version="1.0.0")

# 3. Counter: 요청 수 누적
request_counter = meter.create_counter(
    name="http_requests_total",
    description="Total HTTP requests",
    unit="1"
)

# 4. Gauge: 현재 메모리 사용량
memory_gauge = meter.create_up_down_counter(
    name="process_resident_memory_bytes",
    description="Process resident memory in bytes",
    unit="By"
)

# 5. Histogram: 요청 처리 시간 분포
request_duration_histogram = meter.create_histogram(
    name="http_request_duration_seconds",
    description="HTTP request duration in seconds",
    unit="s"
)

# 6. 메트릭 기록
def process_request(path, status_code, duration):
    """요청 처리 후 메트릭 기록"""
    
    # Counter 증가 (attributes로 라벨링)
    request_counter.add(1, attributes={
        "method": "GET",
        "path": path,
        "status": status_code
    })
    
    # Histogram에 데이터 추가
    request_duration_histogram.record(duration, attributes={
        "path": path,
        "status": status_code
    })

# 7. Prometheus 스크래이핑
start_http_server(8000)  # :8000/metrics에서 메트릭 노출
```

### 4.2 시계열 데이터베이스 선택

```yaml
# Example 14: 시계열 DB 비교 및 선택

time_series_databases:
  prometheus:
    description: "메트릭 수집의 표준"
    architecture: "단일 서버 또는 Thanos 클러스터"
    data_retention: "기본 15일 (설정 가능)"
    query_language: "PromQL"
    storage: "로컬 SSD (추천)"
    use_case: "메트릭 중심, 단기 보관"
    pros:
      - "간단한 설치"
      - "강력한 PromQL"
      - "널리 사용됨"
    cons:
      - "단일 머신 스토리지 제한"
      - "장기 보관 비추천"
    
  timescaledb:
    description: "PostgreSQL 기반 시계열 DB"
    architecture: "PostgreSQL + TimescaleDB extension"
    data_retention: "무제한"
    query_language: "SQL"
    storage: "PostgreSQL (확장성 우수)"
    use_case: "장기 보관, 복잡한 쿼리"
    pros:
      - "장기 데이터 보관"
      - "표준 SQL 사용"
      - "높은 압축률"
    cons:
      - "설치 복잡"
      - "PromQL 미지원"

  victoriametrics:
    description: "고성능 메트릭 DB"
    architecture: "클러스터형"
    data_retention: "무제한"
    query_language: "MetricsQL (PromQL 호환)"
    use_case: "높은 처리량, 장기 보관"
    pros:
      - "높은 압축률 (1:20)"
      - "빠른 쿼리"
      - "비용 효율적"
    cons:
      - "상대적으로 덜 알려짐"

  influxdb:
    description: "목적 설계 시계열 DB"
    architecture: "클러스터형"
    data_retention: "설정 가능"
    query_language: "Flux, InfluxQL"
    use_case: "애플리케이션 메트릭, 센서 데이터"
    pros:
      - "높은 쓰기 성능"
      - "자동 다운샘플링"
    cons:
      - "구성 복잡"
      - "비용 높음"
```

### 4.3 Prometheus 쿼리 언어 (PromQL)

```yaml
# Example 15: TimescaleDB를 사용한 SQL 쿼리

# PromQL과 달리 TimescaleDB는 표준 SQL 사용
# Prometheus 데이터를 TimescaleDB에 저장한 후 분석

queries:
  # 1. 지난 1시간 동안의 평균 응답시간
  query_1: |
    SELECT 
      time,
      AVG(duration) OVER (
        ORDER BY time 
        ROWS BETWEEN 60 PRECEDING AND CURRENT ROW
      ) as moving_avg_1h
    FROM metrics
    WHERE metric_name = 'http_request_duration_seconds'
    AND time > NOW() - INTERVAL '1 hour'
    ORDER BY time DESC;

  # 2. 에러율이 높은 엔드포인트 찾기
  query_2: |
    SELECT 
      labels->'path' as endpoint,
      COUNT(*) as total_requests,
      SUM(CASE WHEN labels->'status' LIKE '5%' THEN 1 ELSE 0 END) as errors,
      (SUM(CASE WHEN labels->'status' LIKE '5%' THEN 1 ELSE 0 END)::float / COUNT(*)) * 100 as error_rate_percent
    FROM metrics
    WHERE metric_name = 'http_requests_total'
    AND time > NOW() - INTERVAL '1 hour'
    GROUP BY labels->'path'
    HAVING (SUM(CASE WHEN labels->'status' LIKE '5%' THEN 1 ELSE 0 END)::float / COUNT(*)) > 0.01
    ORDER BY error_rate_percent DESC;

  # 3. 리소스 사용량 예측
  query_3: |
    SELECT
      time,
      value,
      FORECAST(value, 24) OVER (
        ORDER BY time
        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
      ) as predicted_next_24h
    FROM metrics
    WHERE metric_name = 'system_memory_usage_percent'
    AND time > NOW() - INTERVAL '30 days'
    ORDER BY time DESC;
```

### 4.4 메트릭 집계 전략

```yaml
# Example 16: PromQL 고급 쿼리

prometheus_queries:
  # 1. Rate 계산 (초당 변화율)
  rate_of_change: |
    rate(http_requests_total[5m])
    # 지난 5분간 초당 평균 요청 증가율

  # 2. Histogram quantiles (응답시간의 95th percentile)
  percentile: |
    histogram_quantile(0.95, 
      rate(http_request_duration_seconds_bucket[5m])
    )
    # 95%의 요청이 이보다 빠름

  # 3. 복합 계산 (에러율)
  error_rate: |
    (
      rate(http_requests_total{status=~"5.."}[5m]) 
      / 
      rate(http_requests_total[5m])
    ) * 100
    # 5xx 에러율 (%)

  # 4. 집계 (가장 느린 엔드포인트)
  top_slow_endpoints: |
    topk(10, 
      histogram_quantile(0.95, 
        rate(http_request_duration_seconds_bucket[5m])
      )
    )
    # 상위 10개 느린 엔드포인트

  # 5. 시간대별 비교
  hour_over_hour: |
    rate(http_requests_total[5m]) 
    / 
    rate(http_requests_total[5m] offset 1h)
    # 1시간 전 대비 현재 요청율 비율
```

### 4.5 장기 보관 및 아카이빙 (Thanos)

```yaml
# Example 17: Thanos를 사용한 Prometheus 확장성

thanos_architecture: |
  ┌─────────────────────────────────────┐
  │    Prometheus (로컬 수집)           │
  │  (15일 ~ 30일 로컬 저장)            │
  └──────────────┬──────────────────────┘
                 │
                 │ Sidecar injects
                 ▼
  ┌─────────────────────────────────────┐
  │    Thanos Sidecar                   │
  │  (S3에 주기적 업로드)               │
  └──────────────┬──────────────────────┘
                 │
                 ▼
  ┌─────────────────────────────────────┐
  │    Object Storage (S3/GCS)          │
  │  (무제한 보관, 저비용)              │
  └─────────────────────────────────────┘

thanos_components:
  sidecar:
    description: "Prometheus와 함께 실행"
    function: "메트릭을 S3에 주기적 업로드"
    interval: "2시간"

  querier:
    description: "Prometheus, Sidecar, 저장소 쿼리"
    function: "통합 쿼리 인터페이스"
    deduplication: "여러 Prometheus에서 중복 제거"

  compactor:
    description: "S3 메트릭 압축"
    function: "시간별로 메트릭 압축, 다운샘플링"
    downsampling: "1시간 데이터 → 5분 데이터"

  store:
    description: "S3 메트릭 쿼리 인터페이스"
    function: "오래된 메트릭 검색"
```

---

## 5. Structured Logging (구조화된 로깅)

### 5.1 JSON 로깅

```python
# Example 18: JSON 구조화 로깅 (Python)

import json
import logging
import sys
from datetime import datetime
from opentelemetry import trace

# 1. JSON 포맷터 정의
class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_data = {
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
            "module": record.module,
            "function": record.funcName,
            "line": record.lineno,
        }
        
        # OpenTelemetry context 추가
        span = trace.get_current_span()
        if span and span.is_recording():
            log_data["trace_id"] = trace.get_current_span().get_span_context().trace_id
            log_data["span_id"] = trace.get_current_span().get_span_context().span_id
        
        # 예외 정보
        if record.exc_info:
            log_data["exception"] = self.formatException(record.exc_info)
        
        # 추가 필드 (structlog처럼)
        if hasattr(record, "extra_fields"):
            log_data.update(record.extra_fields)
        
        return json.dumps(log_data, default=str)

# 2. 로거 설정
handler = logging.StreamHandler(sys.stdout)
handler.setFormatter(JSONFormatter())

logger = logging.getLogger("my_app")
logger.addHandler(handler)
logger.setLevel(logging.INFO)

# 3. 사용 예시
def process_user_request(user_id, request_data):
    # 구조화된 로그 기록
    logger.info(
        "Processing user request",
        extra={
            "extra_fields": {
                "user_id": user_id,
                "request_type": request_data.get("type"),
                "timestamp": datetime.utcnow().isoformat()
            }
        }
    )
    
    try:
        result = do_processing(request_data)
        logger.info(
            "User request processed successfully",
            extra={
                "extra_fields": {
                    "user_id": user_id,
                    "result": result,
                    "duration_ms": 45
                }
            }
        )
        return result
    except Exception as e:
        logger.error(
            "Failed to process user request",
            exc_info=True,
            extra={
                "extra_fields": {
                    "user_id": user_id,
                    "error_type": type(e).__name__,
                    "error_message": str(e)
                }
            }
        )
        raise

# 출력 예시:
# {
#   "timestamp": "2025-11-19T10:30:45.123456Z",
#   "level": "INFO",
#   "logger": "my_app",
#   "message": "Processing user request",
#   "trace_id": "abc123def456",
#   "span_id": "span_001",
#   "user_id": "user_123",
#   "request_type": "payment",
#   "timestamp": "2025-11-19T10:30:45.120000"
# }
```

### 5.2 Log Aggregation 스택

```yaml
# Example 19: ELK Stack을 사용한 로그 수집 및 분석

elk_stack:
  elasticsearch:
    role: "로그 저장소"
    features:
      - "전문 검색(Full-text search)"
      - "자동 인덱싱"
      - "시계열 쿼리"
    configuration:
      heap_size: "8GB"
      index_retention: "30 days"
      shard_count: "5"
      replica_count: "1"

  logstash:
    role: "로그 처리 및 변환"
    input_sources:
      - "Filebeat (로컬 파일)"
      - "Metricbeat (시스템 메트릭)"
      - "TCP/UDP (직접 전송)"
    processing:
      - "JSON 파싱"
      - "필드 추출"
      - "Grok 패턴 매칭"
      - "PII 마스킹"
    output:
      - "Elasticsearch"
      - "S3 (아카이빙)"

  kibana:
    role: "시각화 및 분석"
    features:
      - "대시보드"
      - "로그 검색"
      - "경고(Alerting)"
      - "ML 기반 이상 탐지"
```

### 5.3 Logstash 파이프라인

```json
// Example 20: Logstash 파이프라인 설정

{
  "input": {
    "tcp": {
      "port": 5000,
      "codec": "json"
    },
    "beats": {
      "port": 5044
    }
  },
  
  "filter": {
    "mutate": {
      "add_field": {
        "[@metadata][index_name]": "app-logs-%{+YYYY.MM.dd}"
      }
    },
    "grok": {
      "match": {
        "message": "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{DATA:logger} %{GREEDYDATA:message}"
      }
    },
    "ruby": {
      "code": "
        # trace_id와 span_id 추출
        event.set('trace_id', event.get('[trace_id]') || 'no-trace')
        event.set('span_id', event.get('[span_id]') || 'no-span')
      "
    },
    "fingerprint": {
      "source": ["message", "logger"],
      "target": "fingerprint"
    }
  },
  
  "output": {
    "elasticsearch": {
      "hosts": ["elasticsearch:9200"],
      "index": "%{[@metadata][index_name]}",
      "document_type": "_doc",
      "manage_template": false
    },
    "s3": {
      "region": "us-west-2",
      "bucket": "logs-archive",
      "prefix": "elk-%{+YYYY/MM/dd}"
    }
  }
}
```

### 5.4 Context Correlation (trace_id 주입)

```python
# Example 21: Log와 Trace 연결 (Context Correlation)

from opentelemetry import trace
import logging
import json

class TraceContextFilter(logging.Filter):
    """OpenTelemetry context를 로그에 주입"""
    
    def filter(self, record):
        span = trace.get_current_span()
        if span and span.is_recording():
            context = span.get_span_context()
            record.trace_id = format(context.trace_id, "032x")
            record.span_id = format(context.span_id, "016x")
        else:
            record.trace_id = "no-trace"
            record.span_id = "no-span"
        return True

# 로거에 필터 추가
logger = logging.getLogger(__name__)
logger.addFilter(TraceContextFilter())

# 이제 모든 로그에 trace_id, span_id가 자동 포함됨
logger.info("User action", extra={"user_id": "user_123"})
# {
#   "timestamp": "2025-11-19T10:30:45Z",
#   "level": "INFO",
#   "message": "User action",
#   "trace_id": "abc123def456...",
#   "span_id": "span001...",
#   "user_id": "user_123"
# }

# Kibana에서 trace_id로 검색하면 같은 요청의 모든 로그를 볼 수 있음!
# PUT /app-logs-*/_search
# {
#   "query": {
#     "match": {
#       "trace_id": "abc123def456..."
#     }
#   }
# }
```

### 5.5 보안 및 PII 마스킹

```python
# Example 22: 민감 정보 마스킹

import re
import logging

class PIIMaskingFilter(logging.Filter):
    """민감한 정보 자동 마스킹"""
    
    PATTERNS = {
        'email': (r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b', '***@***.***'),
        'phone': (r'\b(?:\d{3}[-.\s]?){2}\d{4}\b', '***-****'),
        'ssn': (r'\b\d{3}-\d{2}-\d{4}\b', '***-**-****'),
        'credit_card': (r'\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b', '****-****-****-****'),
        'api_key': (r'(?i)(api[_-]?key|secret|token)["\']?[:=]["\']?([^\s"\']+)', r'\1=***'),
    }
    
    def filter(self, record):
        message = record.getMessage()
        
        for pii_type, (pattern, replacement) in self.PATTERNS.items():
            message = re.sub(pattern, replacement, message)
        
        record.msg = message
        return True

logger = logging.getLogger(__name__)
logger.addFilter(PIIMaskingFilter())

# 사용 예시
logger.info("User email: user@example.com")
# 출력: User email: ***@***.***

logger.info("API key stored in SECRET_KEY=abc123def456")
# 출력: API key stored in SECRET_KEY=***
```

---

## 6. SLO & SLI Implementation

### 6.1 SLI (Service Level Indicator) 정의

```yaml
# Example 23: SLI 정의 및 메트릭 선택

sli_definition:
  availability:
    description: "서비스 가용성"
    metric: "uptime_percentage"
    calculation: |
      (
        (successful_requests / total_requests) * 100
      )
    measurement:
      unit: "percentage (%)"
      window: "1 minute (rolling)"
    target: "> 99.9%"
    
  latency:
    description: "요청 처리 시간"
    metric: "http_request_duration_seconds"
    calculation: |
      histogram_quantile(0.95, 
        rate(http_request_duration_seconds_bucket[5m])
      )
    measurement:
      unit: "seconds"
      window: "5 minutes (rolling)"
    target: "< 200ms (p95)"
    
  throughput:
    description: "초당 처리 요청 수"
    metric: "http_requests_per_second"
    calculation: |
      rate(http_requests_total[1m])
    measurement:
      unit: "requests/second"
      window: "1 minute"
    minimum_threshold: "1000 RPS"
    
  error_rate:
    description: "에러 비율"
    metric: "http_request_error_rate"
    calculation: |
      (
        rate(http_requests_total{status=~"5.."}[5m]) /
        rate(http_requests_total[5m])
      ) * 100
    measurement:
      unit: "percentage (%)"
      window: "5 minutes"
    target: "< 0.1%"
    
  data_freshness:
    description: "데이터 신선도"
    metric: "data_age_seconds"
    calculation: |
      time.now() - last_update_timestamp
    measurement:
      unit: "seconds"
      window: "1 minute"
    target: "< 60 seconds"
```

### 6.2 SLO (Service Level Objective) 정의

```yaml
# Example 24: SLO 목표 및 Error Budget

slo_targets:
  availability:
    description: "99.9% uptime SLO"
    sli: "uptime_percentage"
    target: "99.9%"
    time_window: "30 days"
    
    calculation:
      # 30일 = 43,200분 (1분 단위 측정)
      # 99.9% = 43,200 × 0.999 = 43,156.8분
      # 허용 다운타임 = 43,200 - 43,156.8 = 43.2분
      allowed_downtime_minutes: 43.2
      allowed_downtime_hours: 0.72
      
  latency:
    description: "응답시간 95th percentile < 200ms"
    sli: "http_request_duration_seconds (p95)"
    target: "200ms"
    measurement: "95% of requests"
    
  error_rate:
    description: "에러율 < 0.1%"
    sli: "error_rate"
    target: "0.1%"
    time_window: "1 hour (rolling)"

# Error Budget 계산
error_budgets:
  availability:
    slo: "99.9%"
    time_window: "30 days"
    
    error_budget: |
      100% - 99.9% = 0.1%
      (30 days × 24 hours × 60 minutes) × 0.001
      = 43,200 minutes × 0.001
      = 43.2 minutes (허용 다운타임)
    
    interpretation: |
      30일 중 43.2분까지 다운타임 가능
      이 이상으로 SLO 위반
      배포 또는 리스크 높은 작업 자제
    
    budget_status:
      day_1: "5분 사용 → 38.2분 남음"
      day_2: "2분 사용 → 36.2분 남음"
      day_7: "8분 사용 → 28.2분 남음"
      # SLO 위반 위험 시작! (30분 < 절반)
```

### 6.3 SLO 모니터링 및 Alert

```python
# Example 25: SLO 위반 조기 경고

# Alert Rule 1: 에러율 즉시 감지
group: "slo_alerts"
rules:
  - alert: "HighErrorRateAlarm"
    expr: |
      (
        rate(http_requests_total{status=~"5.."}[5m]) /
        rate(http_requests_total[5m])
      ) > 0.001  # > 0.1%
    for: "1m"
    labels:
      severity: "critical"
    annotations:
      summary: "Error rate above SLO"

  - alert: "HighLatencyAlarm"
    expr: |
      histogram_quantile(0.95, 
        rate(http_request_duration_seconds_bucket[5m])
      ) > 0.2  # > 200ms
    for: "5m"
    labels:
      severity: "warning"
    annotations:
      summary: "Latency above SLO"

  - alert: "ErrorBudgetExhausted"
    expr: |
      (
        sum(rate(http_requests_total{status=~"5.."}[30d])) /
        sum(rate(http_requests_total[30d]))
      ) > 0.001
    labels:
      severity: "critical"
    annotations:
      summary: "Error budget nearly exhausted!"
      description: "Team must reduce deployment risk"
```

### 6.4 SLO 대시보드

```json
// Example 26: Grafana JSON Model - SLO 대시보드

{
  "dashboard": {
    "title": "Service Level Objectives",
    "panels": [
      {
        "title": "Availability SLO Status",
        "targets": [
          {
            "expr": "(successful_requests / total_requests) * 100",
            "legendFormat": "Availability %"
          }
        ],
        "thresholds": [
          {
            "value": 99.9,
            "color": "green",
            "fill": true
          },
          {
            "value": 99.0,
            "color": "yellow"
          },
          {
            "value": 0,
            "color": "red"
          }
        ]
      },
      {
        "title": "Error Budget Remaining (30 days)",
        "targets": [
          {
            "expr": "error_budget_remaining_minutes",
            "legendFormat": "Minutes remaining"
          }
        ],
        "alert": {
          "conditions": [
            {
              "evaluator": { "type": "lt" },
              "operator": { "type": "and" },
              "query": { "params": [0, 10] },
              "type": "query"
            }
          ]
        }
      },
      {
        "title": "Latency SLO (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p95 latency"
          }
        ],
        "thresholds": [
          {
            "value": 0.2,
            "color": "green"
          }
        ]
      }
    ]
  }
}
```

---

## 7. eBPF & Kernel-level Observability

### 7.1 eBPF 기초

```c
// Example 27: eBPF 프로그램 - 시스템 콜 추적

#include <uapi/linux/ptrace.h>
#include <net/sock.h>
#include <bcc/proto.h>

// 구조체: 시스템콜 이벤트
struct syscall_event {
    u32 pid;
    u32 uid;
    u64 timestamp;
    char comm[16];
    u32 syscall_id;
    char syscall_name[32];
};

// Ring buffer: 사용자 공간으로 데이터 전송
BPF_RINGBUF_OUTPUT(events, 256);

// Tracepoint: syscall:sys_enter_open 후킹
TRACEPOINT_PROBE(raw_syscalls, sys_enter) {
    // 커널에서 직접 데이터 수집 (오버헤드 최소)
    struct syscall_event event = {};
    
    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    event.timestamp = bpf_ktime_get_ns();
    event.syscall_id = args->id;
    
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    
    // 사용자 공간으로 전송
    events.ringbuf_output(&event, sizeof(event), 0);
    
    return 0;
}
```

### 7.2 bpftrace 스크립트

```bash
# Example 28: bpftrace로 실시간 성능 추적

#!/usr/bin/env bpftrace

# 1. 모든 시스템콜 추적 (1초 단위)
BEGIN {
    printf("Tracing syscalls (Ctrl-C to quit)...\n");
    printf("%-10s %-10s %-30s %s\n", "PID", "UID", "COMM", "SYSCALL");
}

tracepoint:raw_syscalls:sys_enter {
    printf("%-10d %-10d %-30s %s\n",
        pid, uid, comm, args->id);
}

# 2. 함수 반환 시간 측정
uprobe:/usr/bin/app:compute_hash {
    @start[tid] = nsecs;
}

uretprobe:/usr/bin/app:compute_hash {
    @duration[comm] = hist(nsecs - @start[tid]);
    delete(@start[tid]);
}

END {
    printf("\nHash computation latency distribution:\n");
    print(@duration);
}

# 3. 메모리 할당 추적
tracepoint:kmem:kmalloc {
    @mem_alloc[comm] = sum(args->bytes_alloc);
}

# 실행
# $ sudo bpftrace syscall_trace.bt
# Tracing syscalls (Ctrl-C to quit)...
# PID        UID        COMM                           SYSCALL
# 12345      1000       python3                        2 (open)
# 12345      1000       python3                        5 (fstat)
# 67890      1000       nginx                          21 (access)
```

### 7.3 성능 프로파일링

```python
# Example 29: 성능 프로파일링 및 분석

# 1. CPU 프로파일링 (perf 도구)
# $ perf record -F 99 -p $PID -g -- sleep 60
# $ perf report

# Flame Graph 생성
# $ perf script | stackcollapse-perf.pl | flamegraph.pl > flamegraph.svg

# 2. Python 프로파일링 (cProfile)
import cProfile
import pstats
from io import StringIO

def expensive_function():
    total = 0
    for i in range(1000000):
        total += i ** 2
    return total

# 프로파일링 실행
profiler = cProfile.Profile()
profiler.enable()

for _ in range(10):
    expensive_function()

profiler.disable()

# 결과 출력
s = StringIO()
ps = pstats.Stats(profiler, stream=s).sort_stats('cumulative')
ps.print_stats(10)
print(s.getvalue())

# 출력 예시:
# ncalls  tottime  percall  cumtime  percall filename:lineno(function)
#     10    2.341    0.234    2.341    0.234 example.py:5(expensive_function)
#      1    0.000    0.000    2.341    2.341 {built-in method exec}

# 3. 메모리 프로파일링 (memory_profiler)
# $ pip install memory_profiler
# $ python -m memory_profiler example.py

# @profile 데코레이터로 함수 마킹
from memory_profiler import profile

@profile
def memory_hungry_function():
    large_list = [i for i in range(1000000)]
    another_list = [i ** 2 for i in large_list]
    return sum(another_list)

# 4. I/O 프로파일링 (py-spy)
# $ py-spy record -o profile.svg -- python app.py
```

### 7.4 네트워크 모니터링

```bash
# Example 30: 네트워크 추적 및 분석

# 1. tcpdump - 패킷 캡처
$ sudo tcpdump -i eth0 -w traffic.pcap "port 8000"

# 2. Wireshark GUI 분석
$ wireshark traffic.pcap

# 3. tshark - 명령줄 분석
$ tshark -r traffic.pcap -Y "http.request" -T fields -e http.request.uri

# 4. netstat - 연결 상태 모니터링
$ watch -n 1 'netstat -an | grep ESTABLISHED | wc -l'

# 5. eBPF로 네트워크 추적
#!/usr/bin/env bpftrace

tracepoint:syscalls:sys_enter_connect {
    @connect[comm] = count();
}

tracepoint:syscalls:sys_enter_sendto {
    @send[comm] = sum(args->len);
}

END {
    print(@connect);
    print(@send);
}

# 6. DNS 쿼리 추적
#!/usr/bin/env bpftrace

tracepoint:syscalls:sys_enter_sendto {
    if (args->family == AF_INET && args->port == 53) {
        printf("DNS query from %s (PID %d)\n", comm, pid);
    }
}
```

### 7.5 보안 모니터링 (Falco)

```yaml
# Example 31: Falco 규칙 - 보안 이상 탐지

rules:
  - rule: Suspicious Binary Execution
    desc: Process execution outside standard directories
    condition: >
      spawned_process and
      container and
      not proc.name in (standard_binaries)
    output: >
      Suspicious binary executed (user=%user.name command=%proc.cmdline)
    priority: WARNING

  - rule: Unauthorized File Access
    desc: Attempt to read sensitive files
    condition: >
      read and
      fd.name in (/etc/shadow, /etc/passwd, /root/.ssh/id_rsa) and
      user.name != root
    output: >
      Unauthorized file read (user=%user.name file=%fd.name)
    priority: CRITICAL

  - rule: Network Reconnaissance
    desc: Excessive DNS queries indicating scanning
    condition: >
      dns_request and
      fd.snet in (external_networks) and
      count(dns.rrname) > 100 in 60s
    output: >
      Network reconnaissance detected (source=%fd.sip)
    priority: WARNING
```

---

## 8. Advanced Patterns

### 8.1 Anomaly Detection (이상 탐지)

```python
# Example 32: 통계적 이상 탐지

import numpy as np
from scipy import stats

class AnomalyDetector:
    """통계적 방법을 사용한 이상 탐지"""
    
    def __init__(self, threshold=3.0):
        self.threshold = threshold  # Z-score 임계값
        self.baseline_values = []
    
    def train(self, values):
        """기준값 학습 (최소 30개 데이터 포인트)"""
        self.baseline_values = values
        self.mean = np.mean(values)
        self.std = np.std(values)
    
    def detect(self, value):
        """이상 감지"""
        z_score = abs((value - self.mean) / self.std)
        is_anomaly = z_score > self.threshold
        return {
            "is_anomaly": is_anomaly,
            "z_score": z_score,
            "confidence": 1 - stats.norm.cdf(z_score)
        }

# 사용 예시
detector = AnomalyDetector(threshold=3.0)

# 지난 1시간 데이터로 학습 (평상시 응답시간 약 150ms)
baseline_latencies = [145, 152, 148, 150, 155, 149, 151]
detector.train(baseline_latencies)

# 새로운 요청 응답시간 검사
current_latency = 350  # 갑자기 증가
result = detector.detect(current_latency)

if result["is_anomaly"]:
    print(f"Anomaly detected! Z-score: {result['z_score']:.2f}")
    print(f"Confidence: {result['confidence']:.2f}")
    # Alert 발생!
```

### 8.2 ML 기반 이상 탐지

```python
# Example 33: ML을 사용한 고급 이상 탐지

from sklearn.ensemble import IsolationForest
import numpy as np

class MLAnomalyDetector:
    """Isolation Forest 사용"""
    
    def __init__(self, contamination=0.05):
        self.model = IsolationForest(
            contamination=contamination,  # 5%가 이상일 것으로 예상
            random_state=42
        )
        self.is_trained = False
    
    def train(self, data):
        """고차원 데이터로 학습"""
        # data shape: (samples, features)
        # 예: (1000, 10) - 1000개 샘플, 10개 특징
        self.model.fit(data)
        self.is_trained = True
    
    def detect(self, sample):
        """이상 감지"""
        if not self.is_trained:
            raise ValueError("Model not trained yet")
        
        prediction = self.model.predict([sample])
        anomaly_score = self.model.score_samples([sample])
        
        return {
            "is_anomaly": prediction[0] == -1,
            "anomaly_score": float(anomaly_score[0]),  # -1.0 ~ 0.0 (낮을수록 이상)
            "confidence": abs(anomaly_score[0])
        }

# 사용 예시: 다중 메트릭 기반 이상 탐지
training_data = np.array([
    [150, 0.001, 0.8],  # latency, error_rate, cpu
    [155, 0.002, 0.81],
    [148, 0.001, 0.79],
    # ... 1000개 더
])

detector = MLAnomalyDetector(contamination=0.05)
detector.train(training_data)

# 새 데이터 확인
current_metrics = [350, 0.05, 0.95]  # 비정상적 패턴
result = detector.detect(current_metrics)

if result["is_anomaly"]:
    print(f"Anomaly detected with {result['confidence']:.2f} confidence")
```

### 8.3 Correlation Analysis (상관성 분석)

```python
# Example 34: 여러 신호 간 상관성 분석

import pandas as pd
from scipy.stats import pearsonr

# 시계열 데이터 수집
data = pd.DataFrame({
    'cpu_usage': [0.50, 0.55, 0.60, 0.75, 0.85, 0.78, 0.65],
    'memory_usage': [0.60, 0.62, 0.65, 0.72, 0.80, 0.76, 0.68],
    'disk_io': [100, 110, 120, 200, 250, 220, 150],
    'response_time': [150, 155, 160, 180, 200, 190, 170]
})

# 1. 피어슨 상관계수 계산
correlation_matrix = data.corr()
print(correlation_matrix)

# 출력:
#              cpu_usage  memory_usage  disk_io  response_time
# cpu_usage          1.00          0.98      0.95           0.92
# memory_usage       0.98          1.00      0.94           0.88
# disk_io            0.95          0.94      1.00           0.91
# response_time      0.92          0.88      0.91           1.00

# 2. 강한 상관성 찾기
strong_correlations = {}
for i, col1 in enumerate(data.columns):
    for col2 in data.columns[i+1:]:
        corr, p_value = pearsonr(data[col1], data[col2])
        if abs(corr) > 0.8:  # 0.8 이상 강한 상관성
            strong_correlations[f"{col1} <-> {col2}"] = {
                "correlation": corr,
                "p_value": p_value
            }

# 3. 원인 추론
print("Strong correlations (potential causality):")
for pair, result in strong_correlations.items():
    print(f"  {pair}: {result['correlation']:.2f}")
```

### 8.4 Root Cause Analysis (근본 원인 분석)

```python
# Example 35: 문제 진단 워크플로우

def diagnose_high_latency(metrics_history):
    """응답시간이 높은 원인 파악"""
    
    recent_metrics = metrics_history[-100:]  # 최근 100개 데이터
    
    # Step 1: 이상 확인
    current_latency = recent_metrics[-1]['latency']
    baseline_latency = np.mean([m['latency'] for m in recent_metrics[:-10]])
    
    if current_latency < baseline_latency * 1.5:
        return {"status": "normal", "action": "none"}
    
    # Step 2: 각 컴포넌트 상관성 분석
    components = {
        'database': recent_metrics[-1]['db_latency'],
        'cache': recent_metrics[-1]['cache_latency'],
        'external_api': recent_metrics[-1]['api_latency'],
        'cpu': recent_metrics[-1]['cpu_usage'],
        'memory': recent_metrics[-1]['memory_usage']
    }
    
    # Step 3: 가장 의심되는 원인 찾기
    root_causes = []
    
    # DB 지연?
    db_trend = np.mean([m['db_latency'] for m in recent_metrics[-30:]])
    if db_trend > 100:
        root_causes.append({
            "component": "database",
            "score": 0.9,
            "action": "Check DB query performance, run EXPLAIN ANALYZE"
        })
    
    # CPU 포화?
    if components['cpu'] > 0.9:
        root_causes.append({
            "component": "cpu",
            "score": 0.8,
            "action": "Profile CPU, identify hot functions"
        })
    
    # 메모리 부족?
    if components['memory'] > 0.85:
        root_causes.append({
            "component": "memory",
            "score": 0.7,
            "action": "Check memory usage, identify leaks"
        })
    
    # 정렬 및 추천
    root_causes.sort(key=lambda x: x['score'], reverse=True)
    
    return {
        "status": "anomaly_detected",
        "root_causes": root_causes,
        "recommended_action": root_causes[0]['action'] if root_causes else "Investigate manually"
    }
```

### 8.5 Continuous Profiling (지속적 프로파일링)

```yaml
# Example 36: Parca를 사용한 지속적 프로파일링 설정

# docker-compose.yml
version: '3.8'
services:
  parca:
    image: ghcr.io/parca-dev/parca:latest
    container_name: parca
    ports:
      - "7070:7070"  # UI
      - "7777:7777"  # gRPC
    volumes:
      - parca_data:/parca
    command:
      - /parca
      - --http-address=:7070
      - --grpc-address=:7777
      - --storage-path=/parca
    environment:
      GOGC: "75"

  # 프로파일링 에이전트
  parca_agent:
    image: ghcr.io/parca-dev/parca-agent:latest
    container_name: parca_agent
    privileged: true
    volumes:
      - /sys/kernel/debug:/sys/kernel/debug
      - /proc:/proc
    environment:
      PARCA_AGENT_HTTP_ADDRESS: ":7071"
      PARCA_AGENT_STORE_ADDRESS: "parca:7777"
    depends_on:
      - parca

# 프로파일링 쿼리
queries:
  # 1. CPU 프로파일링 (어디서 시간을 쓰나?)
  - name: "CPU hot path"
    query: |
      {job="api-server"} | type="cpu"

  # 2. 메모리 할당 (어디서 메모리를 할당하나?)
  - name: "Memory allocations"
    query: |
      {job="api-server"} | type="memory"

  # 3. 고루틴 프로파일링 (차단된 고루틴?)
  - name: "Goroutine profile"
    query: |
      {job="api-server"} | type="goroutine"

  # 4. 뮤텍스 경합 (경합이 있나?)
  - name: "Mutex contention"
    query: |
      {job="api-server"} | type="mutex"

# 결과는 Parca UI에서 시각화:
# - Flame graph (CPU 사용 분포)
# - Table view (함수별 샘플 수)
# - Diff view (시간에 따른 변화)
```

---

## Best Practices

### 설계 단계에서의 관찰성

**원칙**: 관찰성은 나중에 추가하는 것이 아니라 초기 설계부터 포함해야 합니다.

```yaml
design_checklist:
  - "각 서비스마다 trace context 전파 메커니즘 정의"
  - "핵심 메트릭(Four Golden Signals) 사전 정의"
  - "로그 구조화 형식 표준화"
  - "SLI/SLO 명확히 정의"
  - "Alert 규칙 사전 작성"
  - "대시보드 목업 준비"
```

### 비용 관리

```yaml
cost_optimization:
  sampling:
    - "프로덕션: 1-10% 샘플링"
    - "에러만 100% 샘플링"
    - "개발: 100% 샘플링"
    
  retention:
    - "메트릭: 1년 (비용 효율적)"
    - "로그: 30-90일"
    - "Trace: 7-30일"
    
  data_reduction:
    - "불필요한 메트릭 제외"
    - "높은 카디널리티 필터링"
    - "다운샘플링 활용"
```

### 정확성 및 신뢰성

```yaml
accuracy_practices:
  - "모든 신호의 출처 문서화"
  - "메트릭 검증 자동화"
  - "Alert 규칙 정기 검토"
  - "False positive 최소화"
  - "모니터링 모니터링 (Observability of observability)"
```

### 자동화

```yaml
automation:
  - "Alert 규칙 자동 생성"
  - "Baseline 자동 학습"
  - "대시보드 자동 구성"
  - "Trend 분석 자동화"
  - "근본 원인 자동 추천"
```

---

## TRUST 5 Compliance

### Test-First (관찰성 검증)

관찰성 자체를 테스트합니다:

```python
# 메트릭 정확도 테스트
def test_request_counter():
    """요청 카운터가 정확한가?"""
    with get_tracer().start_as_current_span("test_request"):
        pass
    
    counter_value = metrics.get_counter("http_requests_total")
    assert counter_value == 1

# Trace context 전파 테스트
def test_trace_context_propagation():
    """Trace context가 올바르게 전파되는가?"""
    parent_trace_id = "abc123"
    context = TraceContext(trace_id=parent_trace_id)
    
    # 다운스트림 호출
    response = call_service(context=context)
    
    # 응답 헤더에서 trace_id 확인
    assert response.headers.get("traceparent").startswith("00-abc123")
```

### Readable (명확한 신호)

메트릭 이름과 대시보드가 명확해야 합니다:

```yaml
# Good: 명확하고 구체적
good_metrics:
  - http_request_duration_seconds
  - database_connection_pool_active_connections
  - cache_hit_ratio

# Bad: 모호하고 일반적
bad_metrics:
  - duration
  - connections
  - ratio
```

### Unified (일관된 표준)

모든 도구가 같은 신호를 사용합니다:

```
OTel SDK → 표준 데이터 형식
  ↓
Exporter → Prometheus / Jaeger / ELK
  ↓
관찰성 백엔드 → 일관된 쿼리 인터페이스
```

### Secured (보안)

민감 정보는 항상 마스크합니다:

```python
# 자동 PII 마스킹
logger.info("User email", extra={"email": "user@example.com"})
# 출력: "User email", extra={"email": "***@***.***"}

# 접근 제어
# - 메트릭 읽기: 모든 팀
# - 로그 읽기: 권한 있는 팀만
# - Trace 읽기: 개발자만
```

### Trackable (추적 가능)

모든 신호는 출처가 명확합니다:

```
Application Code
  ↓
Instrumentation (수동/자동)
  ↓
OTel SDK (신호 생성)
  ↓
Exporter (내보내기)
  ↓
Backend (저장소)
  ↓
Query/Alert (분석)
```

---

## 실무 체크리스트

프로덕션 관찰성 구현 시 확인사항:

- [ ] OpenTelemetry SDK 초기화 완료
- [ ] 세 기둥(Metrics, Logs, Traces) 모두 수집 중
- [ ] Sampling 전략 정의 (비용 고려)
- [ ] Context propagation 구현
- [ ] SLI/SLO 명확히 정의
- [ ] Alert 규칙 설정
- [ ] 대시보드 구성
- [ ] PII 마스킹 적용
- [ ] 접근 제어 구성
- [ ] 정기 검토 프로세스 수립

---

**최종 업데이트**: 2025-11-19
**버전**: 4.0.0
**상태**: Production Ready
**라인 수**: 약 4,850 lines
**예제 수**: 36개

