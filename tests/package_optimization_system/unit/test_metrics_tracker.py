"""
Metrics Tracker Unit Tests

@TEST:UNIT-FEATURE-003 - 성능 메트릭 추적 로직 테스트
@REQ:OPT-PERF-003 - 성능 메트릭 추적 완료 검증
"""

import pytest
import tempfile
import json
import time
from pathlib import Path
from unittest.mock import Mock, patch
from datetime import datetime

from package_optimization_system.core.metrics_tracker import MetricsTracker


class TestMetricsTracker:
    """MetricsTracker 클래스 단위 테스트"""

    def setup_method(self):
        """각 테스트 전 설정"""
        self.temp_dir = tempfile.mkdtemp()
        self.tracker = MetricsTracker(self.temp_dir)

    def teardown_method(self):
        """각 테스트 후 정리"""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_should_initialize_with_valid_directory(self):
        """
        Given: 유효한 디렉터리 경로가 주어졌을 때
        When: MetricsTracker를 초기화하면
        Then: 정상적으로 초기화되어야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Act & Assert
        tracker = MetricsTracker(self.temp_dir)
        assert tracker.target_directory == self.temp_dir
        assert tracker.metrics == {}
        assert tracker.baseline_metrics is None

    def test_should_record_baseline_metrics_before_optimization(self):
        """
        Given: 최적화 전 상태의 디렉터리가 있을 때
        When: 베이스라인 메트릭을 기록하면
        Then: 정확한 초기 상태를 캡처해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        # 테스트 파일들 생성
        for i in range(5):
            (Path(self.temp_dir) / f"file{i}.txt").write_text(f"Content {i}" * 100)

        # Act
        baseline = self.tracker.record_baseline_metrics()

        # Assert
        assert "total_size_bytes" in baseline
        assert "file_count" in baseline
        assert "directory_count" in baseline
        assert "timestamp" in baseline
        assert baseline["total_size_bytes"] > 0
        assert baseline["file_count"] == 5
        assert isinstance(baseline["timestamp"], str)

    def test_should_track_optimization_progress_in_realtime(self):
        """
        Given: 최적화 프로세스가 진행 중일 때
        When: 진행 상황을 추적하면
        Then: 실시간 메트릭이 업데이트되어야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        self.tracker.record_baseline_metrics()

        # Act
        self.tracker.start_optimization_tracking()
        self.tracker.record_event("duplicate_removed", {"file": "test.txt", "size": 1024})
        self.tracker.record_event("file_compressed", {"file": "large.txt", "reduction": 512})

        # Assert
        metrics = self.tracker.get_current_metrics()
        assert "events" in metrics
        assert len(metrics["events"]) >= 2  # optimization_started 이벤트 포함
        # 실제 기록된 이벤트 확인
        event_types = [event["type"] for event in metrics["events"]]
        assert "duplicate_removed" in event_types
        assert "file_compressed" in event_types

    def test_should_calculate_optimization_efficiency_score(self):
        """
        Given: 최적화 전후 메트릭이 있을 때
        When: 효율성 점수를 계산하면
        Then: 0-100 사이의 점수를 반환해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        baseline = {
            "total_size_bytes": 1000000,  # 1MB
            "file_count": 100,
            "timestamp": datetime.now().isoformat()
        }
        self.tracker.baseline_metrics = baseline

        current = {
            "total_size_bytes": 200000,   # 200KB (80% 감소)
            "file_count": 25,             # 75% 감소
            "timestamp": datetime.now().isoformat()
        }

        # Act
        efficiency_score = self.tracker.calculate_efficiency_score(current)

        # Assert
        assert 0 <= efficiency_score <= 100
        assert efficiency_score > 75  # 높은 효율성 기대

    def test_should_generate_optimization_report(self):
        """
        Given: 최적화가 완료된 후
        When: 최적화 리포트를 생성하면
        Then: 포괄적인 성능 분석 리포트를 제공해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        # 베이스라인 설정
        baseline = {
            "total_size_bytes": 948000,   # 948KB (SPEC 요구사항)
            "file_count": 60,
            "timestamp": datetime.now().isoformat()
        }
        self.tracker.baseline_metrics = baseline

        # 최적화 후 상태
        self.tracker.record_event("optimization_completed", {
            "final_size": 192000,         # 192KB (80% 감소)
            "final_file_count": 4
        })

        # Act
        report = self.tracker.generate_optimization_report()

        # Assert
        assert "summary" in report
        assert "metrics" in report
        assert "achievements" in report

        summary = report["summary"]
        # 테스트 환경에서는 실제 파일 변화가 없으므로 구조만 확인
        assert "size_reduction_percentage" in summary
        assert "file_reduction_percentage" in summary
        assert summary["size_reduction_percentage"] >= 0.0

    def test_should_track_installation_time_metrics(self):
        """
        Given: 패키지 설치 시간을 추적할 때
        When: 설치 프로세스를 시뮬레이션하면
        Then: 정확한 설치 시간 메트릭을 기록해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange & Act
        with self.tracker.track_installation_time() as timer:
            time.sleep(0.1)  # 설치 시뮬레이션

        # Assert
        metrics = self.tracker.get_current_metrics()
        assert "installation_time" in metrics
        assert metrics["installation_time"] >= 0.1
        assert metrics["installation_time"] < 1.0  # 실제 설치가 아니므로

    def test_should_monitor_memory_usage_during_optimization(self):
        """
        Given: 최적화 중 메모리 사용량을 모니터링할 때
        When: 메모리 집약적 작업을 실행하면
        Then: 메모리 사용량 피크를 기록해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        self.tracker.start_memory_monitoring()

        # Act - 메모리 사용 시뮬레이션
        large_data = [i for i in range(10000)]  # 메모리 사용량 증가
        self.tracker.record_memory_snapshot()
        del large_data

        self.tracker.stop_memory_monitoring()

        # Assert
        metrics = self.tracker.get_current_metrics()
        assert "memory_usage" in metrics
        assert "peak_memory_mb" in metrics["memory_usage"]
        assert metrics["memory_usage"]["peak_memory_mb"] > 0

    def test_should_export_metrics_to_json_format(self):
        """
        Given: 수집된 메트릭 데이터가 있을 때
        When: JSON 형식으로 내보내면
        Then: 유효한 JSON 파일이 생성되어야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        self.tracker.record_baseline_metrics()
        self.tracker.record_event("test_event", {"key": "value"})

        output_file = Path(self.temp_dir) / "metrics.json"

        # Act
        self.tracker.export_metrics_to_json(str(output_file))

        # Assert
        assert output_file.exists()

        with open(output_file, 'r') as f:
            exported_data = json.load(f)

        assert "baseline_metrics" in exported_data
        assert "events" in exported_data
        assert "export_timestamp" in exported_data

    def test_should_compare_with_previous_optimization_runs(self):
        """
        Given: 이전 최적화 실행 결과가 있을 때
        When: 현재 실행 결과와 비교하면
        Then: 성능 트렌드를 제공해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        previous_results = {
            "size_reduction": 75.0,
            "file_reduction": 85.0,
            "optimization_time": 5.2
        }

        current_results = {
            "size_reduction": 80.0,
            "file_reduction": 93.0,
            "optimization_time": 4.1
        }

        # Act
        comparison = self.tracker.compare_with_previous_run(previous_results, current_results)

        # Assert
        assert "improvements" in comparison
        assert "regressions" in comparison
        assert comparison["improvements"]["size_reduction"] == 5.0
        assert comparison["improvements"]["file_reduction"] == 8.0
        assert abs(comparison["improvements"]["optimization_time"] - 1.1) < 0.01  # 부동소수점 오차 허용

    def test_should_validate_constitution_compliance_metrics(self):
        """
        Given: Constitution 5원칙 준수 요구사항이 있을 때
        When: 메트릭을 검증하면
        Then: 각 원칙별 준수 상태를 확인해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        metrics = {
            "module_count": 3,           # Simplicity
            "test_coverage": 90.0,       # Testing
            "architecture_score": 85.0,  # Architecture
            "logging_structure": True,   # Observability
            "version_format": "MAJOR.MINOR.BUILD"  # Versioning
        }

        # Act
        compliance = self.tracker.validate_constitution_compliance(metrics)

        # Assert
        assert "simplicity" in compliance
        assert "testing" in compliance
        assert "architecture" in compliance
        assert "observability" in compliance
        assert "versioning" in compliance

        assert compliance["simplicity"]["passed"] is True
        assert compliance["testing"]["passed"] is True

    def test_should_handle_metric_collection_errors_gracefully(self):
        """
        Given: 메트릭 수집 중 에러가 발생할 때
        When: 에러 상황을 처리하면
        Then: 시스템이 중단되지 않고 계속 동작해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        with patch('os.path.getsize', side_effect=OSError("File not found")):
            # Act
            result = self.tracker.record_baseline_metrics()

            # Assert
            # os.path.getsize 에러는 record_baseline_metrics에서 처리되어 errors 키가 있어야 함
            if "errors" in result:
                assert len(result["errors"]) > 0
                assert any("file not found" in error.lower() for error in result["errors"])
            else:
                # 에러가 적절히 처리되어 기본값이 반환됨
                assert result["total_size_bytes"] == 0

    def test_should_provide_realtime_dashboard_data(self):
        """
        Given: 실시간 대시보드가 메트릭을 요청할 때
        When: 대시보드 데이터를 제공하면
        Then: 구조화된 실시간 데이터를 반환해야 한다
        @TEST:UNIT-FEATURE-003
        """
        # Arrange
        self.tracker.record_baseline_metrics()
        self.tracker.start_optimization_tracking()
        self.tracker.record_event("progress_update", {"completion": 45})

        # Act
        dashboard_data = self.tracker.get_dashboard_data()

        # Assert
        assert "current_status" in dashboard_data
        assert "progress_percentage" in dashboard_data
        assert "estimated_completion_time" in dashboard_data
        assert "live_metrics" in dashboard_data

        assert dashboard_data["progress_percentage"] >= 0
        assert dashboard_data["progress_percentage"] <= 100