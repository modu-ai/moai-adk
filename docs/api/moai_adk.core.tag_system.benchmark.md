# moai_adk.core.tag_system.benchmark

@FEATURE:SPEC-009-TAG-PERFORMANCE-001 - TAG 성능 벤치마크 도구

GREEN 단계: 10x 성능 개선 및 50% 메모리 절약 검증

## Functions

### create_comparison_report

메모리 사용량 비교 리포트 생성

```python
create_comparison_report(self, json_usage, sqlite_usage, dataset_size)
```

### __init__

벤치마크 도구 초기화

```python
__init__(self, database_path, json_path)
```

### monitoring_session

모니터링 세션 컨텍스트 매니저

```python
monitoring_session(self, session_name)
```

### record_query

쿼리 기록

```python
record_query(self, query_time_ms, success)
```

### generate_report

모니터링 리포트 생성

```python
generate_report(self)
```

### _percentile

퍼센타일 계산

```python
_percentile(self, sorted_list, percentile)
```

### __enter__

```python
__enter__(self)
```

### __exit__

```python
__exit__(self, exc_type, exc_val, exc_tb)
```

### create_performance_report

성능 비교 리포트 생성

```python
create_performance_report(self, json_results, sqlite_results)
```

### create_real_time_monitor

실시간 성능 모니터 생성

```python
create_real_time_monitor(self, alert_thresholds)
```

## Classes

### PerformanceMetrics

성능 메트릭

### BenchmarkResult

벤치마크 결과

### MemoryProfileResult

메모리 프로파일 결과

### ScalabilityResult

확장성 분석 결과

### Alert

성능 알림

### MonitoringReport

모니터링 리포트

### ResourceUsage

리소스 사용량

### PerformanceComparison

성능 비교 결과

### LoadTestSuite

부하 테스트 도구

### MemoryProfiler

메모리 프로파일러

### PerformanceMonitor

실시간 성능 모니터

### MonitoringSession

모니터링 세션

### TagPerformanceBenchmark

TAG 성능 벤치마크 도구

TRUST 원칙 적용:
- Test First: 성능 테스트 요구사항 충족
- Readable: 명확한 벤치마크 로직
- Unified: 성능 측정 책임만 담당
