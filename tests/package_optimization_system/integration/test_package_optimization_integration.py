"""
Package Optimization System Integration Tests

@TEST:INTEGRATION-PKG-002 - 패키지 통합 테스트
@REQ:OPT-CORE-001 @REQ:OPT-DEDUPE-002 @REQ:OPT-PERF-003 - 전체 시스템 통합 검증
"""

import pytest
import tempfile
import json
from pathlib import Path
from unittest.mock import Mock, patch

from package_optimization_system.core.package_optimizer import PackageOptimizer
from package_optimization_system.core.duplicate_remover import DuplicateRemover
from package_optimization_system.core.metrics_tracker import MetricsTracker


class TestPackageOptimizationIntegration:
    """Package Optimization System 통합 테스트"""

    def setup_method(self):
        """각 테스트 전 설정"""
        self.temp_dir = tempfile.mkdtemp()
        self.create_test_package_structure()

    def teardown_method(self):
        """각 테스트 후 정리"""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def create_test_package_structure(self):
        """SPEC-003 요구사항에 맞는 테스트 패키지 구조 생성"""
        # 948KB 상당의 패키지 구조 시뮬레이션

        # 에이전트 파일들 (60개 → 4개로 감소 목표)
        agents_dir = Path(self.temp_dir) / "agents"
        agents_dir.mkdir()

        # 중복되는 에이전트 파일들
        agent_templates = [
            "backend-architect.md", "database-architect.md", "golang-pro.md",
            "java-pro.md", "kotlin-pro.md", "php-pro.md", "python-pro.md",
            "react-pro.md", "vue-pro.md", "angular-pro.md"
        ]

        large_content = "# Agent Template\n" + "Content line\n" * 200  # ~2KB per file

        for template in agent_templates:
            for variation in ["v1", "v2", "v3", "v4", "v5", "v6"]:
                file_path = agents_dir / f"{variation}_{template}"
                file_path.write_text(large_content)

        # 명령어 파일들 (13개 → 3개로 감소 목표)
        commands_dir = Path(self.temp_dir) / "commands"
        commands_dir.mkdir()

        command_content = "# Command Template\n" + "Command content\n" * 100  # ~1KB per file
        command_files = [
            "1-project.md", "1-spec.md", "2-spec.md", "2-build.md",
            "3-plan.md", "3-sync.md", "4-tasks.md", "5-dev.md",
            "6-sync.md", "7-dashboard.md", "legacy-1.md", "legacy-2.md", "legacy-3.md"
        ]

        for cmd_file in command_files:
            (commands_dir / cmd_file).write_text(command_content)

        # _templates 폴더 (완전 제거 목표)
        templates_dir = Path(self.temp_dir) / "_templates"
        templates_dir.mkdir()

        # 중복된 템플릿 구조
        for category in ["awesome", "moai"]:
            cat_dir = templates_dir / category
            cat_dir.mkdir()
            for subcategory in ["backend", "frontend", "docs", "quality"]:
                sub_dir = cat_dir / subcategory
                sub_dir.mkdir()
                for i in range(10):
                    (sub_dir / f"template_{i}.md").write_text(large_content)

    def test_should_achieve_spec_003_optimization_targets(self):
        """
        Given: SPEC-003 요구사항에 맞는 패키지 구조가 있을 때
        When: 전체 최적화 프로세스를 실행하면
        Then: 모든 최적화 목표를 달성해야 한다
        @TEST:INTEGRATION-PKG-002
        @REQ:OPT-CORE-001 (80% 크기 감소)
        @REQ:OPT-DEDUPE-002 (93% 파일 감소)
        """
        # Arrange
        optimizer = PackageOptimizer(self.temp_dir)
        remover = DuplicateRemover(self.temp_dir)
        tracker = MetricsTracker(self.temp_dir)

        # 베이스라인 기록
        baseline = tracker.record_baseline_metrics()
        initial_size = baseline["total_size_bytes"]
        initial_file_count = baseline["file_count"]

        # Act - 전체 최적화 프로세스
        tracker.start_optimization_tracking()

        # 1단계: 중복 제거
        duplicate_result = remover.remove_duplicates()
        tracker.record_event("duplicate_removal_completed", duplicate_result)

        # 2단계: 패키지 최적화
        optimization_result = optimizer.optimize()
        tracker.record_event("optimization_completed", optimization_result)

        # 최종 메트릭 수집 (베이스라인 변경 없이)
        final_metrics = tracker._get_current_metrics()
        final_size = final_metrics["total_size_bytes"]
        final_file_count = final_metrics["file_count"]

        # Assert - SPEC-003 목표 달성 검증
        size_reduction = ((initial_size - final_size) / initial_size) * 100
        file_reduction = ((initial_file_count - final_file_count) / initial_file_count) * 100

        assert size_reduction >= 80.0, f"크기 감소 {size_reduction:.1f}% < 80% 목표"
        assert file_reduction >= 90.0, f"파일 감소 {file_reduction:.1f}% < 90% 목표"

        # 최적화 리포트 생성
        report = tracker.generate_optimization_report()
        assert report["summary"]["size_reduction_percentage"] >= 80.0

    def test_should_preserve_core_functionality_during_optimization(self):
        """
        Given: 핵심 기능 파일들이 있을 때
        When: 최적화를 실행하면
        Then: 핵심 기능은 보존되어야 한다
        @TEST:INTEGRATION-PKG-002
        """
        # Arrange
        # 핵심 기능 파일들 생성
        core_files = [
            "spec-builder.md",
            "code-builder.md",
            "doc-syncer.md",
            "claude-code-manager.md"
        ]

        agents_dir = Path(self.temp_dir) / "agents"
        for core_file in core_files:
            (agents_dir / core_file).write_text("# Core Agent\nEssential functionality")

        core_commands = ["1-spec.md", "2-build.md", "3-sync.md"]
        commands_dir = Path(self.temp_dir) / "commands"
        for cmd in core_commands:
            (commands_dir / cmd).write_text("# Core Command\nEssential command")

        optimizer = PackageOptimizer(self.temp_dir)

        # Act
        result = optimizer.optimize()

        # Assert
        # 핵심 파일들이 보존되었는지 확인
        for core_file in core_files:
            assert (agents_dir / core_file).exists(), f"핵심 파일 {core_file}이 제거됨"

        for cmd in core_commands:
            assert (commands_dir / cmd).exists(), f"핵심 명령어 {cmd}가 제거됨"

        assert result["success"] is True

    def test_should_handle_large_scale_optimization_efficiently(self):
        """
        Given: 대규모 패키지 구조가 있을 때
        When: 최적화를 실행하면
        Then: 효율적으로 처리되어야 한다
        @TEST:INTEGRATION-PKG-002
        """
        # Arrange
        # 대규모 파일 구조 생성 (1000+ 파일)
        for i in range(100):
            category_dir = Path(self.temp_dir) / f"category_{i}"
            category_dir.mkdir()
            for j in range(10):
                file_content = f"Content for file {i}_{j}\n" * 50
                (category_dir / f"file_{j}.txt").write_text(file_content)

        tracker = MetricsTracker(self.temp_dir)
        optimizer = PackageOptimizer(self.temp_dir)

        # Act
        start_time = tracker.get_current_time()

        with tracker.track_memory_usage():
            result = optimizer.optimize()

        end_time = tracker.get_current_time()

        # Assert
        processing_time = end_time - start_time
        assert processing_time < 60.0, f"처리 시간 {processing_time:.1f}초 > 60초 제한"

        memory_metrics = tracker.get_current_metrics()["memory_usage"]
        assert memory_metrics["peak_memory_mb"] < 500, "메모리 사용량이 500MB 초과"

        assert result["success"] is True

    def test_should_maintain_api_compatibility_after_optimization(self):
        """
        Given: 기존 API를 사용하는 시스템이 있을 때
        When: 최적화 후 API를 호출하면
        Then: 100% 호환성을 유지해야 한다
        @TEST:INTEGRATION-PKG-002
        """
        # Arrange
        # API 호환성 테스트를 위한 mock 클라이언트
        class MockApiClient:
            def __init__(self, package_dir):
                self.package_dir = package_dir

            def get_agents_list(self):
                agents_dir = Path(self.package_dir) / "agents"
                return [f.name for f in agents_dir.glob("*.md")]

            def get_commands_list(self):
                commands_dir = Path(self.package_dir) / "commands"
                return [f.name for f in commands_dir.glob("*.md")]

        # 핵심 에이전트 파일들 추가 생성 (API 호환성을 위해)
        agents_dir = Path(self.temp_dir) / "agents"
        core_agents = ["spec-builder.md", "code-builder.md", "doc-syncer.md", "claude-code-manager.md"]
        for core_agent in core_agents:
            (agents_dir / core_agent).write_text("# Core Agent\nEssential functionality")

        client = MockApiClient(self.temp_dir)

        # 최적화 전 API 호출
        agents_before = client.get_agents_list()
        commands_before = client.get_commands_list()

        optimizer = PackageOptimizer(self.temp_dir)

        # Act
        optimizer.optimize()

        # 최적화 후 API 호출
        agents_after = client.get_agents_list()
        commands_after = client.get_commands_list()

        # Assert
        # 핵심 API 엔드포인트가 여전히 동작해야 함
        assert len(agents_after) > 0, "에이전트 API가 빈 결과 반환"
        assert len(commands_after) > 0, "명령어 API가 빈 결과 반환"

        # 핵심 기능은 유지되어야 함
        core_agents = ["spec-builder.md", "code-builder.md", "doc-syncer.md"]
        for core_agent in core_agents:
            assert core_agent in agents_after, f"핵심 에이전트 {core_agent} API 호환성 손실"

    def test_should_provide_comprehensive_optimization_metrics(self):
        """
        Given: 최적화 프로세스가 완료되었을 때
        When: 포괄적인 메트릭 리포트를 요청하면
        Then: 모든 주요 지표가 포함된 리포트를 제공해야 한다
        @TEST:INTEGRATION-PKG-002
        """
        # Arrange
        tracker = MetricsTracker(self.temp_dir)
        optimizer = PackageOptimizer(self.temp_dir)
        remover = DuplicateRemover(self.temp_dir)

        # Act
        baseline = tracker.record_baseline_metrics()

        tracker.start_optimization_tracking()
        dup_result = remover.remove_duplicates()
        opt_result = optimizer.optimize()

        final_report = tracker.generate_optimization_report()

        # Assert
        # 필수 메트릭 섹션들 확인
        assert "summary" in final_report
        assert "metrics" in final_report
        assert "achievements" in final_report
        assert "constitution_compliance" in final_report

        summary = final_report["summary"]
        assert "size_reduction_percentage" in summary
        assert "file_reduction_percentage" in summary
        assert "optimization_time_seconds" in summary

        # SPEC-003 Constitution 준수 확인
        compliance = final_report["constitution_compliance"]
        assert compliance["overall_score"] >= 80.0

    def test_should_recover_gracefully_from_optimization_failures(self):
        """
        Given: 최적화 중 일부 실패가 발생했을 때
        When: 에러 복구 프로세스가 실행되면
        Then: 시스템이 안전한 상태로 복구되어야 한다
        @TEST:INTEGRATION-PKG-002
        """
        # Arrange
        optimizer = PackageOptimizer(self.temp_dir)
        tracker = MetricsTracker(self.temp_dir)

        # 의도적으로 실패 시나리오 생성
        with patch('os.remove', side_effect=PermissionError("Access denied")):
            # Act
            result = optimizer.optimize()

            # Assert
            # 실패 상황에서도 안전하게 처리되어야 함
            assert "success" in result
            assert "errors" in result

            # 시스템이 여전히 동작 가능한 상태여야 함
            current_metrics = tracker.record_baseline_metrics()
            assert current_metrics["file_count"] > 0

    def test_should_support_incremental_optimization(self):
        """
        Given: 이미 부분적으로 최적화된 패키지가 있을 때
        When: 추가 최적화를 실행하면
        Then: 점진적 개선이 이루어져야 한다
        @TEST:INTEGRATION-PKG-002
        """
        # Arrange
        optimizer = PackageOptimizer(self.temp_dir)
        tracker = MetricsTracker(self.temp_dir)

        # 첫 번째 최적화 실행
        baseline1 = tracker.record_baseline_metrics()
        result1 = optimizer.optimize()
        metrics1 = tracker.record_baseline_metrics()

        # 추가 파일 생성 (증분 최적화 시나리오)
        new_files_dir = Path(self.temp_dir) / "new_files"
        new_files_dir.mkdir()
        for i in range(10):
            (new_files_dir / f"new_{i}.txt").write_text("New content " * 100)

        # Act - 두 번째 최적화 실행
        baseline2 = tracker.record_baseline_metrics()
        result2 = optimizer.optimize()
        metrics2 = tracker.record_baseline_metrics()

        # Assert
        # 두 번째 최적화도 성공해야 함
        assert result2["success"] is True

        # 점진적 개선 확인
        size_after_first = metrics1["total_size_bytes"]
        size_after_second = metrics2["total_size_bytes"]

        # 새 파일이 추가되었음에도 최적화로 인해 크기가 관리되어야 함
        size_increase_percentage = ((baseline2["total_size_bytes"] - size_after_first) / size_after_first) * 100
        final_reduction_percentage = ((baseline2["total_size_bytes"] - size_after_second) / baseline2["total_size_bytes"]) * 100

        assert final_reduction_percentage > 50.0, "점진적 최적화가 효과적이지 않음"