"""
@TEST:SPEC-009-TAG-PERFORMANCE-001 - TAG 시스템 성능 벤치마크 실패 테스트

RED 단계: 10x 성능 개선 및 50% 메모리 절약 목표 달성 실패 테스트
"""

import pytest
import time
import psutil
import os
import json
import sqlite3
import tempfile
from pathlib import Path
from typing import Dict, List, Any, Tuple
from concurrent.futures import ThreadPoolExecutor, ProcessPoolExecutor
from unittest.mock import Mock, patch

# 아직 구현되지 않은 모듈들 - 실패할 예정
from moai_adk.core.tag_system.benchmark import (
    TagPerformanceBenchmark,
    PerformanceMetrics,
    MemoryProfiler,
    BenchmarkResult,
    PerformanceComparison,
    LoadTestSuite,
)
from moai_adk.core.tag_system.database import TagDatabaseManager
from moai_adk.core.tag_system.adapter import TagIndexAdapter
from moai_adk.core.tag_system.index_manager import TagIndexManager


class TestTagPerformanceBenchmark:
    """TAG 시스템 성능 벤치마크 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.temp_db = self.temp_dir / "benchmark.db"
        self.temp_json = self.temp_dir / "benchmark.json"

        self.benchmark = TagPerformanceBenchmark(
            database_path=self.temp_db, json_path=self.temp_json
        )

        # 메모리 프로파일러 초기화
        self.memory_profiler = MemoryProfiler()

    def teardown_method(self):
        """각 테스트 후 정리"""
        import shutil

        if self.temp_dir.exists():
            shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_should_achieve_10x_search_performance_improvement(self):
        """
        Given: 동일한 대용량 TAG 데이터셋
        When: JSON 방식과 SQLite 방식의 검색 성능을 비교할 때
        Then: SQLite가 10x 이상 빠른 성능을 보여야 함
        """
        # GIVEN: 대용량 TAG 데이터셋 준비 (5000개 TAG)
        large_dataset = self._create_large_dataset(tag_count=5000)

        # JSON 백엔드 초기화
        json_manager = TagIndexManager(
            watch_directory=self.temp_dir, index_file=self.temp_json
        )
        json_manager.initialize_index()

        # SQLite 백엔드 초기화
        sqlite_adapter = TagIndexAdapter(
            database_path=self.temp_db, json_fallback_path=self.temp_json
        )
        sqlite_adapter.initialize()

        # 동일한 데이터로 두 백엔드 초기화
        self._populate_backends_with_dataset(
            large_dataset, json_manager, sqlite_adapter
        )

        # WHEN: 다양한 검색 시나리오 성능 측정
        search_scenarios = [
            ("category_search", "REQ"),
            ("identifier_pattern", "USER-*"),
            ("file_path_search", "/spec/*.md"),
            (
                "complex_query",
                {"category": "REQ", "file_pattern": "*.md", "line_range": (1, 100)},
            ),
        ]

        json_results = {}
        sqlite_results = {}

        for scenario_name, search_params in search_scenarios:
            # JSON 검색 성능
            json_time = self._measure_search_performance(
                json_manager, scenario_name, search_params
            )
            json_results[scenario_name] = json_time

            # SQLite 검색 성능
            sqlite_time = self._measure_search_performance(
                sqlite_adapter, scenario_name, search_params
            )
            sqlite_results[scenario_name] = sqlite_time

        # THEN: 10x 성능 개선 검증
        for scenario in search_scenarios:
            scenario_name = scenario[0]
            json_time = json_results[scenario_name]
            sqlite_time = sqlite_results[scenario_name]

            performance_ratio = json_time / sqlite_time
            assert performance_ratio >= 10.0, (
                f"{scenario_name}: SQLite가 {performance_ratio:.2f}x 빠름 (목표: 10x 이상)\n"
                f"JSON: {json_time:.3f}s, SQLite: {sqlite_time:.3f}s"
            )

        # 종합 성능 리포트 생성
        benchmark_result = self.benchmark.create_performance_report(
            json_results, sqlite_results
        )
        assert benchmark_result.overall_improvement_ratio >= 10.0
        assert benchmark_result.all_scenarios_meet_target is True

    def test_should_achieve_50_percent_memory_reduction(self):
        """
        Given: 동일한 대용량 TAG 데이터
        When: JSON과 SQLite 방식의 메모리 사용량을 측정할 때
        Then: SQLite가 50% 이상 메모리를 절약해야 함
        """
        # GIVEN: 메모리 사용량 측정 준비
        dataset_size = 3000  # 메모리 측정을 위한 적당한 크기
        large_dataset = self._create_large_dataset(tag_count=dataset_size)

        process = psutil.Process(os.getpid())

        # JSON 방식 메모리 측정
        initial_memory = process.memory_info().rss

        json_manager = TagIndexManager(
            watch_directory=self.temp_dir, index_file=self.temp_json
        )
        json_manager.initialize_index()
        self._populate_json_backend(large_dataset, json_manager)

        # JSON 전체 로딩
        json_index = json_manager.load_index()
        json_memory_after_load = process.memory_info().rss
        json_memory_usage = json_memory_after_load - initial_memory

        # 메모리 정리
        del json_index
        del json_manager
        import gc

        gc.collect()

        # SQLite 방식 메모리 측정
        reset_memory = process.memory_info().rss

        sqlite_adapter = TagIndexAdapter(
            database_path=self.temp_db, json_fallback_path=self.temp_json
        )
        sqlite_adapter.initialize()
        self._populate_sqlite_backend(large_dataset, sqlite_adapter)

        # SQLite 동일한 데이터 접근
        sqlite_index = sqlite_adapter.load_index()
        sqlite_memory_after_load = process.memory_info().rss
        sqlite_memory_usage = sqlite_memory_after_load - reset_memory

        # WHEN: 메모리 사용량 비교
        memory_reduction_ratio = (
            json_memory_usage - sqlite_memory_usage
        ) / json_memory_usage
        memory_reduction_percentage = memory_reduction_ratio * 100

        # THEN: 50% 이상 메모리 절약 검증
        assert memory_reduction_percentage >= 50.0, (
            f"메모리 절약률: {memory_reduction_percentage:.1f}% (목표: 50% 이상)\n"
            f"JSON 메모리 사용량: {json_memory_usage / 1024 / 1024:.2f}MB\n"
            f"SQLite 메모리 사용량: {sqlite_memory_usage / 1024 / 1024:.2f}MB"
        )

        # 메모리 프로파일 리포트 생성
        memory_report = self.memory_profiler.create_comparison_report(
            json_usage=json_memory_usage,
            sqlite_usage=sqlite_memory_usage,
            dataset_size=dataset_size,
        )

        assert memory_report.memory_efficiency_improvement >= 50.0
        assert memory_report.memory_per_tag_json > memory_report.memory_per_tag_sqlite

    def test_should_handle_concurrent_load_efficiently(self):
        """
        Given: 다중 스레드/프로세스 환경
        When: 동시에 많은 TAG 검색 요청을 처리할 때
        Then: SQLite가 JSON보다 높은 동시성 처리 능력을 보여야 함
        """
        # GIVEN: 동시성 테스트를 위한 데이터 준비
        concurrent_dataset = self._create_large_dataset(tag_count=2000)

        # 백엔드 초기화
        json_manager = TagIndexManager(
            watch_directory=self.temp_dir, index_file=self.temp_json
        )
        json_manager.initialize_index()
        self._populate_json_backend(concurrent_dataset, json_manager)

        sqlite_adapter = TagIndexAdapter(
            database_path=self.temp_db, json_fallback_path=self.temp_json
        )
        sqlite_adapter.initialize()
        self._populate_sqlite_backend(concurrent_dataset, sqlite_adapter)

        # WHEN: 동시 접근 성능 테스트
        concurrent_users = 20
        queries_per_user = 50

        def json_worker(user_id: int) -> Dict[str, float]:
            start_time = time.time()
            for i in range(queries_per_user):
                category = ["REQ", "DESIGN", "TASK", "TEST"][i % 4]
                # JSON 검색 시뮬레이션
                index_data = json_manager.load_index()  # 매번 파일 읽기
                results = [
                    tag_info
                    for category_group in index_data["categories"].values()
                    for cat_name, cat_tags in category_group.items()
                    if cat_name == category
                    for tag_id, tag_info in cat_tags.items()
                ]
            total_time = time.time() - start_time
            return {
                "user_id": user_id,
                "total_time": total_time,
                "queries": queries_per_user,
            }

        def sqlite_worker(user_id: int) -> Dict[str, float]:
            start_time = time.time()
            for i in range(queries_per_user):
                category = ["REQ", "DESIGN", "TASK", "TEST"][i % 4]
                # SQLite 검색 (연결 재사용)
                results = sqlite_adapter._database.search_tags_by_category(category)
            total_time = time.time() - start_time
            return {
                "user_id": user_id,
                "total_time": total_time,
                "queries": queries_per_user,
            }

        # 동시 실행 - JSON
        with ThreadPoolExecutor(max_workers=concurrent_users) as executor:
            json_start = time.time()
            json_futures = [
                executor.submit(json_worker, i) for i in range(concurrent_users)
            ]
            json_results = [f.result() for f in json_futures]
            json_total_time = time.time() - json_start

        # 동시 실행 - SQLite
        with ThreadPoolExecutor(max_workers=concurrent_users) as executor:
            sqlite_start = time.time()
            sqlite_futures = [
                executor.submit(sqlite_worker, i) for i in range(concurrent_users)
            ]
            sqlite_results = [f.result() for f in sqlite_futures]
            sqlite_total_time = time.time() - sqlite_start

        # THEN: 동시성 성능 검증
        total_queries = concurrent_users * queries_per_user

        json_throughput = total_queries / json_total_time
        sqlite_throughput = total_queries / sqlite_total_time

        throughput_improvement = sqlite_throughput / json_throughput

        assert throughput_improvement >= 5.0, (
            f"동시성 처리 개선: {throughput_improvement:.2f}x (목표: 5x 이상)\n"
            f"JSON 처리량: {json_throughput:.1f} queries/sec\n"
            f"SQLite 처리량: {sqlite_throughput:.1f} queries/sec"
        )

        # 사용자별 성능 분산 확인 (SQLite가 더 일관된 성능을 보여야 함)
        json_times = [r["total_time"] for r in json_results]
        sqlite_times = [r["total_time"] for r in sqlite_results]

        import statistics

        json_std_dev = statistics.stdev(json_times)
        sqlite_std_dev = statistics.stdev(sqlite_times)

        assert sqlite_std_dev < json_std_dev, "SQLite가 더 일관된 성능을 보여야 함"

    def test_should_scale_linearly_with_data_size(self):
        """
        Given: 다양한 크기의 TAG 데이터셋
        When: 데이터 크기를 증가시키며 성능을 측정할 때
        Then: SQLite가 더 나은 확장성을 보여야 함 (선형 증가)
        """
        # GIVEN: 다양한 크기의 데이터셋 준비
        data_sizes = [500, 1000, 2000, 4000, 8000]
        json_performance = []
        sqlite_performance = []

        for data_size in data_sizes:
            # 각 크기별 데이터셋 생성
            dataset = self._create_large_dataset(tag_count=data_size)

            # JSON 성능 측정
            json_temp = self.temp_dir / f"json_{data_size}.json"
            json_manager = TagIndexManager(
                watch_directory=self.temp_dir, index_file=json_temp
            )
            json_manager.initialize_index()
            self._populate_json_backend(dataset, json_manager)

            json_time = self._measure_full_search_performance(json_manager)
            json_performance.append((data_size, json_time))

            # SQLite 성능 측정
            sqlite_temp = self.temp_dir / f"sqlite_{data_size}.db"
            sqlite_adapter = TagIndexAdapter(
                database_path=sqlite_temp, json_fallback_path=json_temp
            )
            sqlite_adapter.initialize()
            self._populate_sqlite_backend(dataset, sqlite_adapter)

            sqlite_time = self._measure_full_search_performance(sqlite_adapter)
            sqlite_performance.append((data_size, sqlite_time))

        # WHEN: 확장성 분석
        json_scalability = self._analyze_scalability(json_performance)
        sqlite_scalability = self._analyze_scalability(sqlite_performance)

        # THEN: SQLite가 더 나은 확장성 보유
        # 시간 복잡도: JSON은 O(n), SQLite는 O(log n) 목표
        assert (
            sqlite_scalability.complexity_order < json_scalability.complexity_order
        ), (
            f"확장성 비교: JSON {json_scalability.complexity_order}, "
            f"SQLite {sqlite_scalability.complexity_order}"
        )

        # 가장 큰 데이터셋에서의 성능 차이
        largest_json_time = json_performance[-1][1]
        largest_sqlite_time = sqlite_performance[-1][1]
        final_performance_ratio = largest_json_time / largest_sqlite_time

        assert final_performance_ratio >= 15.0, (
            f"대용량 데이터 성능 비율: {final_performance_ratio:.2f}x (목표: 15x 이상)\n"
            f"8000개 TAG - JSON: {largest_json_time:.3f}s, SQLite: {largest_sqlite_time:.3f}s"
        )

    def test_should_optimize_database_with_proper_indexing(self):
        """
        Given: SQLite 데이터베이스와 인덱스 최적화
        When: 인덱스 유무에 따른 성능을 비교할 때
        Then: 인덱스가 쿼리 성능을 크게 향상시켜야 함
        """
        # GIVEN: 테스트 데이터 준비
        dataset = self._create_large_dataset(tag_count=5000)

        # 인덱스 없는 SQLite 데이터베이스
        no_index_db = self.temp_dir / "no_index.db"
        no_index_manager = TagDatabaseManager(no_index_db)
        no_index_manager.initialize(create_indexes=False)  # 인덱스 생성 안함

        # 인덱스 있는 SQLite 데이터베이스
        with_index_db = self.temp_dir / "with_index.db"
        with_index_manager = TagDatabaseManager(with_index_db)
        with_index_manager.initialize(create_indexes=True)  # 인덱스 생성

        # 동일한 데이터로 두 데이터베이스 채우기
        for tag_data in dataset:
            no_index_manager.insert_tag(**tag_data)
            with_index_manager.insert_tag(**tag_data)

        # WHEN: 다양한 쿼리 성능 비교
        query_scenarios = [
            ("category_search", lambda db: db.search_tags_by_category("REQ")),
            (
                "identifier_search",
                lambda db: db.search_tags_by_identifier("USER-LOGIN-001"),
            ),
            (
                "file_path_search",
                lambda db: db.search_tags_by_file("/spec/requirements.md"),
            ),
            ("line_range_search", lambda db: db.search_tags_by_line_range(1, 100)),
        ]

        performance_improvements = {}

        for scenario_name, query_func in query_scenarios:
            # 인덱스 없는 쿼리 성능
            no_index_time = self._measure_query_time(no_index_manager, query_func)

            # 인덱스 있는 쿼리 성능
            with_index_time = self._measure_query_time(with_index_manager, query_func)

            improvement_ratio = no_index_time / with_index_time
            performance_improvements[scenario_name] = improvement_ratio

        # THEN: 인덱스 효과 검증
        for scenario, improvement in performance_improvements.items():
            assert improvement >= 5.0, (
                f"{scenario}: 인덱스 성능 개선 {improvement:.2f}x (목표: 5x 이상)"
            )

        # 전체 평균 개선율
        avg_improvement = sum(performance_improvements.values()) / len(
            performance_improvements
        )
        assert avg_improvement >= 10.0, f"평균 인덱스 효과: {avg_improvement:.2f}x"

    def test_should_handle_memory_pressure_gracefully(self):
        """
        Given: 메모리 제약이 있는 환경
        When: 제한된 메모리에서 대용량 TAG 처리를 수행할 때
        Then: SQLite가 메모리 효율적으로 동작해야 함
        """
        # GIVEN: 메모리 제약 시뮬레이션
        memory_limit_mb = 100  # 100MB 제한
        initial_memory = psutil.Process(os.getpid()).memory_info().rss

        # 대용량 데이터 (메모리 압박 상황)
        large_dataset = self._create_large_dataset(tag_count=10000)

        # WHEN: 메모리 제약 하에서 처리
        sqlite_adapter = TagIndexAdapter(
            database_path=self.temp_db,
            json_fallback_path=self.temp_json,
            memory_limit_mb=memory_limit_mb,  # 메모리 제한 설정
        )
        sqlite_adapter.initialize()

        # 배치 처리로 메모리 효율성 확보
        batch_size = 1000
        for i in range(0, len(large_dataset), batch_size):
            batch = large_dataset[i : i + batch_size]
            sqlite_adapter.bulk_insert_tags(batch)

            # 메모리 사용량 모니터링
            current_memory = psutil.Process(os.getpid()).memory_info().rss
            memory_increase = (current_memory - initial_memory) / 1024 / 1024

            assert memory_increase < memory_limit_mb * 1.5, (
                f"메모리 사용량 {memory_increase:.1f}MB가 제한 {memory_limit_mb}MB를 초과"
            )

        # THEN: 메모리 효율성 검증
        final_memory = psutil.Process(os.getpid()).memory_info().rss
        total_memory_increase = (final_memory - initial_memory) / 1024 / 1024

        # 10,000개 TAG를 처리했지만 메모리 증가는 제한적이어야 함
        memory_per_tag = total_memory_increase * 1024 / len(large_dataset)  # KB per tag

        assert memory_per_tag < 5.0, (
            f"TAG당 메모리 사용량: {memory_per_tag:.2f}KB (목표: 5KB 미만)"
        )

        # 기능성 검증 (메모리 효율성이 기능을 해치지 않았는지)
        search_results = sqlite_adapter._database.search_tags_by_category("REQ")
        assert len(search_results) > 0, "메모리 최적화가 기능에 영향을 주면 안됨"

    def test_should_provide_real_time_performance_monitoring(self):
        """
        Given: 운영 환경에서의 성능 모니터링 필요
        When: 실시간으로 성능 메트릭을 수집할 때
        Then: 상세한 성능 통계와 알림을 제공해야 함
        """
        # GIVEN: 성능 모니터링 설정
        performance_monitor = self.benchmark.create_real_time_monitor(
            alert_thresholds={
                "query_time_ms": 100,  # 100ms 초과 시 알림
                "memory_usage_mb": 50,  # 50MB 초과 시 알림
                "error_rate_percent": 1,  # 1% 초과 시 알림
            }
        )

        sqlite_adapter = TagIndexAdapter(
            database_path=self.temp_db,
            json_fallback_path=self.temp_json,
            performance_monitor=performance_monitor,
        )
        sqlite_adapter.initialize()

        # 테스트 데이터 준비
        dataset = self._create_large_dataset(tag_count=1000)
        self._populate_sqlite_backend(dataset, sqlite_adapter)

        # WHEN: 다양한 작업 수행하며 모니터링
        with performance_monitor.monitoring_session("performance_test"):
            for i in range(100):  # 100회 반복
                start_time = time.time()
                category = ["REQ", "DESIGN", "TASK", "TEST"][i % 4]
                results = sqlite_adapter._database.search_tags_by_category(category)

                query_time_ms = (time.time() - start_time) * 1000

                # 일부 느린 쿼리 의도적 생성 (알림 테스트) - 더 적게
                if i % 50 == 0:  # 100개 중 2개만 느리게
                    time.sleep(0.15)  # 150ms 지연 (알림 임계값 초과)
                    query_time_ms = 150.0  # 알림 임계값 초과

                # 성능 모니터에 쿼리 기록
                performance_monitor.record_query(query_time_ms, success=True)

        # THEN: 성능 메트릭 수집 및 분석
        performance_report = performance_monitor.generate_report()

        # 기본 메트릭 확인
        assert performance_report.total_queries == 100
        assert performance_report.avg_query_time_ms < 50  # 평균은 양호해야 함
        assert performance_report.error_count == 0

        # 알림 발생 확인 (느린 쿼리 2회)
        alerts = performance_report.alerts
        slow_query_alerts = [a for a in alerts if a.type == "SLOW_QUERY"]
        assert len(slow_query_alerts) == 2, (
            f"예상 느린 쿼리 알림: 2개, 실제: {len(slow_query_alerts)}개"
        )

        # 성능 분포 확인
        percentiles = performance_report.query_time_percentiles
        assert percentiles["p50"] < 10  # 중앙값 10ms 미만
        assert percentiles["p95"] < 20  # 95% 20ms 미만
        assert percentiles["p99"] > 100  # 99% 지연 쿼리 포함

        # 리소스 사용량 모니터링
        resource_usage = performance_report.resource_usage
        assert resource_usage.peak_memory_mb < 100
        assert resource_usage.avg_cpu_percent < 50

    # 헬퍼 메서드들

    def _create_large_dataset(self, tag_count: int) -> List[Dict[str, Any]]:
        """대용량 TAG 데이터셋 생성"""
        categories = [
            "REQ",
            "DESIGN",
            "TASK",
            "TEST",
            "VISION",
            "STRUCT",
            "TECH",
            "FEATURE",
            "API",
            "PERF",
        ]
        dataset = []

        for i in range(tag_count):
            category = categories[i % len(categories)]
            dataset.append(
                {
                    "category": category,
                    "identifier": f"PERF-TEST-{i:06d}",
                    "description": f"성능 테스트용 TAG {i} - "
                    + "x" * (i % 100),  # 가변 길이 설명
                    "file_path": f"/perf/test/{category.lower()}/file_{i // 100:04d}.md",
                    "line_number": (i % 500) + 1,
                }
            )

        return dataset

    def _measure_search_performance(self, backend, scenario: str, params) -> float:
        """검색 성능 측정"""
        start_time = time.time()

        # 시나리오별 검색 실행
        if scenario == "category_search":
            for _ in range(10):  # 10회 반복
                results = backend.search_tags_by_category(params)
        elif scenario == "identifier_pattern":
            for _ in range(10):
                results = backend.search_tags_by_pattern(params)
        elif scenario == "file_path_search":
            for _ in range(10):
                results = backend.search_tags_by_file_pattern(params)
        elif scenario == "complex_query":
            for _ in range(10):
                results = backend.complex_search(**params)

        return time.time() - start_time

    def _populate_backends_with_dataset(self, dataset, json_manager, sqlite_adapter):
        """두 백엔드에 동일한 데이터 입력"""
        # JSON 백엔드 채우기
        self._populate_json_backend(dataset, json_manager)

        # SQLite 백엔드 채우기
        self._populate_sqlite_backend(dataset, sqlite_adapter)

    def _populate_json_backend(self, dataset, json_manager):
        """JSON 백엔드에 데이터 입력"""
        # JSON 인덱스에 직접 데이터 추가 (실제 파일 처리 시뮬레이션)
        for tag_data in dataset:
            # 임시 파일 생성 및 처리로 시뮬레이션
            temp_file = self.temp_dir / f"temp_{tag_data['identifier']}.md"
            content = f"@{tag_data['category']}:{tag_data['identifier']} {tag_data['description']}"
            temp_file.write_text(content)

            json_manager.process_file_change(temp_file, "created")
            temp_file.unlink()

    def _populate_sqlite_backend(self, dataset, sqlite_adapter):
        """SQLite 백엔드에 데이터 입력"""
        for tag_data in dataset:
            sqlite_adapter._database.insert_tag(**tag_data)

    def _measure_full_search_performance(self, backend) -> float:
        """전체 검색 성능 측정"""
        start_time = time.time()

        # 다양한 검색 수행
        categories = ["REQ", "DESIGN", "TASK", "TEST"]
        for category in categories:
            results = backend.search_tags_by_category(category)

        return time.time() - start_time

    def _analyze_scalability(self, performance_data) -> "ScalabilityAnalysis":
        """확장성 분석"""
        # 성능 데이터에서 시간 복잡도 추정
        sizes = [data[0] for data in performance_data]
        times = [data[1] for data in performance_data]

        # 선형 회귀로 복잡도 추정 (단순화)
        import math

        # O(n) vs O(log n) 구분을 위한 간단한 휴리스틱
        if len(sizes) >= 3:
            growth_rate = (times[-1] / times[0]) / (sizes[-1] / sizes[0])
            if growth_rate > 0.8:  # 거의 선형
                complexity_order = "O(n)"
            elif growth_rate < 0.3:  # 로그 또는 상수
                complexity_order = "O(log n)"
            else:
                complexity_order = "O(n*log n)"
        else:
            complexity_order = "Unknown"

        return type(
            "ScalabilityAnalysis",
            (),
            {
                "complexity_order": complexity_order,
                "growth_rate": growth_rate if "growth_rate" in locals() else 0,
                "performance_data": performance_data,
            },
        )()

    def _measure_query_time(self, db_manager, query_func) -> float:
        """쿼리 실행 시간 측정"""
        start_time = time.time()

        # 5회 반복 후 평균
        for _ in range(5):
            results = query_func(db_manager)

        return (time.time() - start_time) / 5
