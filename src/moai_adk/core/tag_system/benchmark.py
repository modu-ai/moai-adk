"""
@FEATURE:SPEC-009-TAG-PERFORMANCE-001 - TAG 성능 벤치마크 도구

GREEN 단계: 10x 성능 개선 및 50% 메모리 절약 검증
"""

import time
from dataclasses import dataclass
from pathlib import Path


@dataclass
class PerformanceMetrics:
    """성능 메트릭"""

    query_time_ms: float
    memory_usage_mb: float
    throughput_qps: float
    error_rate_percent: float


@dataclass
class BenchmarkResult:
    """벤치마크 결과"""

    overall_improvement_ratio: float
    all_scenarios_meet_target: bool
    scenario_results: dict[str, dict[str, float]]


@dataclass
class MemoryProfileResult:
    """메모리 프로파일 결과"""

    memory_efficiency_improvement: float
    memory_per_tag_json: float
    memory_per_tag_sqlite: float


@dataclass
class ScalabilityResult:
    """확장성 분석 결과"""

    complexity_order: str
    growth_rate: float
    performance_data: list[tuple]


@dataclass
class Alert:
    """성능 알림"""

    type: str
    message: str
    threshold: float
    actual_value: float
    timestamp: float


@dataclass
class MonitoringReport:
    """모니터링 리포트"""

    total_queries: int
    avg_query_time_ms: float
    error_count: int
    alerts: list[Alert]
    query_time_percentiles: dict[str, float]
    resource_usage: "ResourceUsage"


@dataclass
class ResourceUsage:
    """리소스 사용량"""

    peak_memory_mb: float
    avg_cpu_percent: float


@dataclass
class PerformanceComparison:
    """성능 비교 결과"""

    json_performance: dict[str, float]
    sqlite_performance: dict[str, float]
    improvement_ratios: dict[str, float]


@dataclass
class LoadTestSuite:
    """부하 테스트 도구"""

    concurrent_users: int
    queries_per_user: int


class MemoryProfiler:
    """메모리 프로파일러"""

    def create_comparison_report(
        self, json_usage: int, sqlite_usage: int, dataset_size: int
    ) -> MemoryProfileResult:
        """메모리 사용량 비교 리포트 생성"""
        json_usage_mb = json_usage / 1024 / 1024
        sqlite_usage_mb = sqlite_usage / 1024 / 1024

        efficiency_improvement = (
            (json_usage_mb - sqlite_usage_mb) / json_usage_mb
        ) * 100

        return MemoryProfileResult(
            memory_efficiency_improvement=efficiency_improvement,
            memory_per_tag_json=json_usage / dataset_size,
            memory_per_tag_sqlite=sqlite_usage / dataset_size,
        )


class PerformanceMonitor:
    """실시간 성능 모니터"""

    def __init__(self, alert_thresholds: dict[str, float]):
        self.alert_thresholds = alert_thresholds
        self._queries = []
        self._alerts = []
        self._start_time = None

    def monitoring_session(self, session_name: str):
        """모니터링 세션 컨텍스트 매니저"""
        return MonitoringSession(self, session_name)

    def record_query(self, query_time_ms: float, success: bool = True):
        """쿼리 기록"""
        query_record = {
            "timestamp": time.time(),
            "query_time_ms": query_time_ms,
            "success": success,
        }
        self._queries.append(query_record)

        # 임계값 체크
        if query_time_ms > self.alert_thresholds.get("query_time_ms", 100):
            alert = Alert(
                type="SLOW_QUERY",
                message=f"Query took {query_time_ms:.1f}ms (threshold: {self.alert_thresholds['query_time_ms']}ms)",
                threshold=self.alert_thresholds["query_time_ms"],
                actual_value=query_time_ms,
                timestamp=time.time(),
            )
            self._alerts.append(alert)

    def generate_report(self) -> MonitoringReport:
        """모니터링 리포트 생성"""
        query_times = [q["query_time_ms"] for q in self._queries]
        successful_queries = [q for q in self._queries if q["success"]]

        # 퍼센타일 계산
        if query_times:
            sorted_times = sorted(query_times)
            percentiles = {
                "p50": self._percentile(sorted_times, 50),
                "p95": self._percentile(sorted_times, 95),
                "p99": self._percentile(sorted_times, 99),
            }
        else:
            percentiles = {"p50": 0, "p95": 0, "p99": 0}

        return MonitoringReport(
            total_queries=len(self._queries),
            avg_query_time_ms=sum(query_times) / len(query_times) if query_times else 0,
            error_count=len(self._queries) - len(successful_queries),
            alerts=self._alerts,
            query_time_percentiles=percentiles,
            resource_usage=ResourceUsage(
                peak_memory_mb=50, avg_cpu_percent=25
            ),  # 임시값
        )

    def _percentile(self, sorted_list: list[float], percentile: int) -> float:
        """퍼센타일 계산"""
        if not sorted_list:
            return 0.0
        index = int(len(sorted_list) * percentile / 100)
        return sorted_list[min(index, len(sorted_list) - 1)]


class MonitoringSession:
    """모니터링 세션"""

    def __init__(self, monitor: PerformanceMonitor, session_name: str):
        self.monitor = monitor
        self.session_name = session_name

    def __enter__(self):
        self.monitor._start_time = time.time()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        pass


class TagPerformanceBenchmark:
    """
    TAG 성능 벤치마크 도구

    TRUST 원칙 적용:
    - Test First: 성능 테스트 요구사항 충족
    - Readable: 명확한 벤치마크 로직
    - Unified: 성능 측정 책임만 담당
    """

    def __init__(self, database_path: Path, json_path: Path):
        """벤치마크 도구 초기화"""
        self.database_path = Path(database_path)
        self.json_path = Path(json_path)

    def create_performance_report(
        self, json_results: dict[str, float], sqlite_results: dict[str, float]
    ) -> BenchmarkResult:
        """성능 비교 리포트 생성"""
        scenario_results = {}
        improvement_ratios = []

        for scenario in json_results:
            json_time = json_results[scenario]
            sqlite_time = sqlite_results[scenario]
            ratio = json_time / sqlite_time if sqlite_time > 0 else 0

            scenario_results[scenario] = {
                "json_time": json_time,
                "sqlite_time": sqlite_time,
                "improvement_ratio": ratio,
            }
            improvement_ratios.append(ratio)

        overall_improvement = (
            sum(improvement_ratios) / len(improvement_ratios)
            if improvement_ratios
            else 0
        )
        all_meet_target = all(ratio >= 10.0 for ratio in improvement_ratios)

        return BenchmarkResult(
            overall_improvement_ratio=overall_improvement,
            all_scenarios_meet_target=all_meet_target,
            scenario_results=scenario_results,
        )

    def create_real_time_monitor(
        self, alert_thresholds: dict[str, float]
    ) -> PerformanceMonitor:
        """실시간 성능 모니터 생성"""
        return PerformanceMonitor(alert_thresholds)
