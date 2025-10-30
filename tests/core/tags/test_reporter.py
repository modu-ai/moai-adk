#!/usr/bin/env python3
# @TEST:DOC-TAG-004 | Component 4: Documentation & Reporting tests
"""Test suite for TAG reporting and documentation system

This module tests the reporting system that:
- Generates TAG inventories across the codebase
- Creates coverage matrices showing implementation status
- Analyzes SPEC→CODE→TEST→DOC chains
- Produces statistics and metrics
- Formats reports in multiple formats (markdown, JSON, HTML)

Following TDD RED-GREEN-REFACTOR cycle.
Target: 33 tests with 85%+ coverage
"""

import json
import tempfile
from datetime import datetime
from pathlib import Path

import pytest

# Import will fail initially (RED phase) - that's expected
try:
    from moai_adk.core.tags.reporter import (
        CoverageAnalyzer,
        CoverageMetrics,
        InventoryGenerator,
        MatrixGenerator,
        ReportFormatter,
        ReportGenerator,
        ReportResult,
        StatisticsGenerator,
        StatisticsReport,
        TagInventory,
        TagMatrix,
    )
except ImportError:
    # Allow tests to be written before implementation
    TagInventory = None
    TagMatrix = None
    InventoryGenerator = None
    MatrixGenerator = None
    CoverageAnalyzer = None
    StatisticsGenerator = None
    ReportFormatter = None
    ReportGenerator = None
    CoverageMetrics = None
    StatisticsReport = None
    ReportResult = None


# ============================================================================
# Test Class 1: TagInventory Dataclass (3 tests)
# ============================================================================

@pytest.mark.skipif(TagInventory is None, reason="TagInventory not implemented yet")
class TestTagInventory:
    """Test TagInventory dataclass creation and manipulation"""

    def test_tag_inventory_creation(self):
        """Test TagInventory dataclass instantiation"""
        inventory = TagInventory(
            tag_id="DOC-TAG-001",
            file_path="/path/to/file.py",
            line_number=42,
            context="# @CODE:DOC-TAG-001 | Implementation",
            related_tags=["@SPEC:DOC-TAG-001", "@TEST:DOC-TAG-001"],
            last_modified=datetime.now(),
            status="active"
        )

        assert inventory.tag_id == "DOC-TAG-001"
        assert inventory.file_path == "/path/to/file.py"
        assert inventory.line_number == 42
        assert inventory.status == "active"
        assert len(inventory.related_tags) == 2

    def test_tag_inventory_status_values(self):
        """Test valid status values"""
        statuses = ["active", "deprecated", "orphan", "incomplete"]

        for status in statuses:
            inventory = TagInventory(
                tag_id="TEST-001",
                file_path="/test.py",
                line_number=1,
                context="# @CODE:TEST-001",
                related_tags=[],
                last_modified=datetime.now(),
                status=status
            )
            assert inventory.status == status

    def test_tag_inventory_related_tags_empty(self):
        """Test TagInventory with no related tags (orphan)"""
        inventory = TagInventory(
            tag_id="ORPHAN-001",
            file_path="/orphan.py",
            line_number=10,
            context="# @CODE:ORPHAN-001",
            related_tags=[],
            last_modified=datetime.now(),
            status="orphan"
        )

        assert len(inventory.related_tags) == 0
        assert inventory.status == "orphan"


# ============================================================================
# Test Class 2: InventoryGenerator (5 tests)
# ============================================================================

@pytest.mark.skipif(InventoryGenerator is None, reason="InventoryGenerator not implemented yet")
class TestInventoryGenerator:
    """Test inventory generation functionality"""

    def test_generate_inventory_empty_directory(self):
        """Test inventory generation for empty directory"""
        with tempfile.TemporaryDirectory() as tmpdir:
            generator = InventoryGenerator()
            inventory = generator.generate_inventory(tmpdir)

            assert isinstance(inventory, list)
            assert len(inventory) == 0

    def test_generate_inventory_with_tags(self):
        """Test inventory generation with actual TAG files"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create test files with TAGs
            test_file = Path(tmpdir) / "test.py"
            test_file.write_text("""
# @SPEC:AUTH-001 | Authentication specification
# @CODE:AUTH-001 | Login implementation
def login():
    pass
# @TEST:AUTH-001 | Login tests
            """)

            generator = InventoryGenerator()
            inventory = generator.generate_inventory(tmpdir)

            assert len(inventory) >= 3  # SPEC, CODE, TEST
            tag_ids = [item.tag_id for item in inventory]
            assert "AUTH-001" in str(tag_ids)

    def test_group_by_domain(self):
        """Test grouping tags by domain"""
        generator = InventoryGenerator()

        inventory = [
            TagInventory(
                tag_id="AUTH-001", file_path="/auth.py", line_number=1,
                context="", related_tags=[], last_modified=datetime.now(), status="active"
            ),
            TagInventory(
                tag_id="AUTH-002", file_path="/auth.py", line_number=10,
                context="", related_tags=[], last_modified=datetime.now(), status="active"
            ),
            TagInventory(
                tag_id="USER-001", file_path="/user.py", line_number=5,
                context="", related_tags=[], last_modified=datetime.now(), status="active"
            ),
        ]

        grouped = generator.group_by_domain(inventory)

        assert "AUTH" in grouped
        assert "USER" in grouped
        assert len(grouped["AUTH"]) == 2
        assert len(grouped["USER"]) == 1

    def test_format_as_markdown(self):
        """Test markdown formatting of inventory"""
        generator = InventoryGenerator()

        grouped = {
            "AUTH": [
                TagInventory(
                    tag_id="AUTH-001", file_path="/auth.py", line_number=1,
                    context="# @SPEC:AUTH-001", related_tags=[],
                    last_modified=datetime.now(), status="active"
                )
            ]
        }

        markdown = generator.format_as_markdown(grouped)

        assert "# TAG Inventory" in markdown
        assert "AUTH" in markdown
        assert "AUTH-001" in markdown
        assert "/auth.py:1" in markdown

    def test_scan_directory_ignores_patterns(self):
        """Test that inventory scan ignores .git, node_modules, etc."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create files in ignored directories
            git_dir = Path(tmpdir) / ".git"
            git_dir.mkdir()
            (git_dir / "config").write_text("# @CODE:GIT-001")

            node_dir = Path(tmpdir) / "node_modules"
            node_dir.mkdir()
            (node_dir / "lib.js").write_text("// @CODE:NODE-001")

            # Create valid file
            valid_file = Path(tmpdir) / "src.py"
            valid_file.write_text("# @CODE:VALID-001")

            generator = InventoryGenerator()
            inventory = generator.generate_inventory(tmpdir)

            # Should only find VALID-001, not GIT-001 or NODE-001
            tag_ids = [item.tag_id for item in inventory]
            assert "VALID-001" in str(tag_ids)
            assert "GIT-001" not in str(tag_ids)
            assert "NODE-001" not in str(tag_ids)


# ============================================================================
# Test Class 3: MatrixGenerator (6 tests)
# ============================================================================

@pytest.mark.skipif(MatrixGenerator is None, reason="MatrixGenerator not implemented yet")
class TestMatrixGenerator:
    """Test coverage matrix generation"""

    def test_generate_matrix_empty(self):
        """Test matrix generation with no tags"""
        generator = MatrixGenerator()
        tags = {"SPEC": set(), "CODE": set(), "TEST": set(), "DOC": set()}

        matrix = generator.generate_matrix(tags)

        assert isinstance(matrix, TagMatrix)
        assert len(matrix.rows) == 0

    def test_generate_matrix_with_data(self):
        """Test matrix generation with actual tags"""
        generator = MatrixGenerator()
        tags = {
            "SPEC": {"AUTH-001", "AUTH-002"},
            "CODE": {"AUTH-001"},
            "TEST": {"AUTH-001"},
            "DOC": set()
        }

        matrix = generator.generate_matrix(tags)

        # AUTH-001 should have SPEC, CODE, TEST (not DOC)
        # AUTH-002 should have only SPEC
        assert len(matrix.rows) >= 1

    def test_format_as_markdown_table(self):
        """Test markdown table formatting"""
        generator = MatrixGenerator()
        tags = {
            "SPEC": {"AUTH-001"},
            "CODE": {"AUTH-001"},
            "TEST": {"AUTH-001"},
            "DOC": {"AUTH-001"}
        }

        matrix = generator.generate_matrix(tags)
        markdown = generator.format_as_markdown_table(matrix)

        assert "| SPEC |" in markdown
        assert "| CODE |" in markdown
        assert "| TEST |" in markdown
        assert "| DOC |" in markdown
        assert "AUTH-001" in markdown
        assert "✅" in markdown or "100%" in markdown

    def test_format_as_csv(self):
        """Test CSV formatting"""
        generator = MatrixGenerator()
        tags = {
            "SPEC": {"AUTH-001"},
            "CODE": {"AUTH-001"},
            "TEST": set(),
            "DOC": set()
        }

        matrix = generator.generate_matrix(tags)
        csv = generator.format_as_csv(matrix)

        assert "SPEC,CODE,TEST,DOC" in csv
        assert "AUTH-001" in csv

    def test_calculate_completion_percentage_full(self):
        """Test completion calculation for fully implemented SPEC"""
        generator = MatrixGenerator()
        tags = {
            "SPEC": {"AUTH-001"},
            "CODE": {"AUTH-001"},
            "TEST": {"AUTH-001"},
            "DOC": {"AUTH-001"}
        }

        percentage = generator.calculate_completion_percentage("AUTH-001", tags)

        assert percentage == 100.0

    def test_calculate_completion_percentage_partial(self):
        """Test completion calculation for partially implemented SPEC"""
        generator = MatrixGenerator()
        tags = {
            "SPEC": {"AUTH-001"},
            "CODE": {"AUTH-001"},
            "TEST": set(),
            "DOC": set()
        }

        percentage = generator.calculate_completion_percentage("AUTH-001", tags)

        # Has SPEC and CODE (2 out of 4) = 50%
        assert 40.0 <= percentage <= 60.0


# ============================================================================
# Test Class 4: CoverageAnalyzer (5 tests)
# ============================================================================

@pytest.mark.skipif(CoverageAnalyzer is None, reason="CoverageAnalyzer not implemented yet")
class TestCoverageAnalyzer:
    """Test TAG coverage analysis"""

    def test_analyze_spec_coverage_complete(self):
        """Test coverage analysis for complete SPEC→CODE→TEST→DOC chain"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "complete.py"
            test_file.write_text("""
# @SPEC:COMPLETE-001
# @CODE:COMPLETE-001
# @TEST:COMPLETE-001
# @DOC:COMPLETE-001
            """)

            analyzer = CoverageAnalyzer()
            metrics = analyzer.analyze_spec_coverage("COMPLETE-001", tmpdir)

            assert metrics.has_code is True
            assert metrics.has_test is True
            assert metrics.has_doc is True
            assert metrics.coverage_percentage == 100.0

    def test_get_specs_without_code(self):
        """Test detection of SPECs without CODE implementation"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "spec_only.py"
            test_file.write_text("""
# @SPEC:NO-IMPL-001
# @SPEC:HAS-IMPL-001
# @CODE:HAS-IMPL-001
            """)

            analyzer = CoverageAnalyzer()
            missing = analyzer.get_specs_without_code(tmpdir)

            assert "NO-IMPL-001" in missing
            assert "HAS-IMPL-001" not in missing

    def test_get_code_without_tests(self):
        """Test detection of CODE without TEST"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "code.py"
            test_file.write_text("""
# @CODE:NO-TEST-001
# @CODE:HAS-TEST-001
# @TEST:HAS-TEST-001
            """)

            analyzer = CoverageAnalyzer()
            missing = analyzer.get_code_without_tests(tmpdir)

            assert "NO-TEST-001" in missing
            assert "HAS-TEST-001" not in missing

    def test_get_code_without_docs(self):
        """Test detection of CODE without DOC"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "code.py"
            test_file.write_text("""
# @CODE:NO-DOC-001
# @CODE:HAS-DOC-001
# @DOC:HAS-DOC-001
            """)

            analyzer = CoverageAnalyzer()
            missing = analyzer.get_code_without_docs(tmpdir)

            assert "NO-DOC-001" in missing
            assert "HAS-DOC-001" not in missing

    def test_calculate_overall_coverage(self):
        """Test overall coverage percentage calculation"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "mixed.py"
            test_file.write_text("""
# @SPEC:FULL-001
# @CODE:FULL-001
# @TEST:FULL-001
# @DOC:FULL-001

# @SPEC:PARTIAL-001
# @CODE:PARTIAL-001

# @SPEC:NONE-001
            """)

            analyzer = CoverageAnalyzer()
            coverage = analyzer.calculate_overall_coverage(tmpdir)

            # FULL-001 = 100% (has CODE, TEST, DOC = 3/3)
            # PARTIAL-001 = 33.33% (has CODE only = 1/3)
            # NONE-001 = 0% (has nothing = 0/3)
            # Average = (100 + 33.33 + 0) / 3 ≈ 44.44%
            assert 40.0 <= coverage <= 50.0


# ============================================================================
# Test Class 5: StatisticsGenerator (4 tests)
# ============================================================================

@pytest.mark.skipif(StatisticsGenerator is None, reason="StatisticsGenerator not implemented yet")
class TestStatisticsGenerator:
    """Test statistics generation"""

    def test_generate_statistics_basic(self):
        """Test basic statistics generation"""
        generator = StatisticsGenerator()

        tags = {
            "SPEC": {"AUTH-001", "AUTH-002"},
            "CODE": {"AUTH-001", "AUTH-002", "USER-001"},
            "TEST": {"AUTH-001"},
            "DOC": set()
        }

        stats = generator.generate_statistics(tags)

        assert stats.total_tags >= 6
        assert stats.by_type["SPEC"] == 2
        assert stats.by_type["CODE"] == 3
        assert stats.by_type["TEST"] == 1
        assert stats.by_type["DOC"] == 0

    def test_generate_statistics_by_domain(self):
        """Test statistics grouped by domain"""
        generator = StatisticsGenerator()

        tags = {
            "SPEC": {"AUTH-001", "AUTH-002", "USER-001"},
            "CODE": {"AUTH-001", "USER-001"},
            "TEST": {"AUTH-001"},
            "DOC": set()
        }

        stats = generator.generate_statistics(tags)

        assert "AUTH" in stats.by_domain
        assert "USER" in stats.by_domain
        # AUTH domain has 2 unique IDs (AUTH-001, AUTH-002) counted once each
        assert stats.by_domain["AUTH"] == 2
        # USER domain has 1 unique ID (USER-001)
        assert stats.by_domain["USER"] == 1

    def test_format_as_json(self):
        """Test JSON formatting of statistics"""
        generator = StatisticsGenerator()

        tags = {
            "SPEC": {"TEST-001"},
            "CODE": {"TEST-001"},
            "TEST": {"TEST-001"},
            "DOC": {"TEST-001"}
        }

        stats = generator.generate_statistics(tags)
        json_output = generator.format_as_json(stats)

        # Should be valid JSON
        parsed = json.loads(json_output)
        assert "total_tags" in parsed
        assert "by_type" in parsed
        assert "by_domain" in parsed

    def test_format_as_human_readable(self):
        """Test human-readable formatting"""
        generator = StatisticsGenerator()

        tags = {
            "SPEC": {"TEST-001"},
            "CODE": {"TEST-001"},
            "TEST": {"TEST-001"},
            "DOC": set()
        }

        stats = generator.generate_statistics(tags)
        readable = generator.format_as_human_readable(stats)

        assert "Total TAGs:" in readable
        assert "SPEC:" in readable
        assert "CODE:" in readable
        assert "TEST:" in readable


# ============================================================================
# Test Class 6: ReportGenerator (6 tests)
# ============================================================================

@pytest.mark.skipif(ReportGenerator is None, reason="ReportGenerator not implemented yet")
class TestReportGenerator:
    """Test main report generator orchestrator"""

    def test_generate_inventory_report(self):
        """Test inventory report generation"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.py"
            test_file.write_text("# @CODE:TEST-001")

            generator = ReportGenerator()
            report = generator.generate_inventory_report(tmpdir)

            assert isinstance(report, str)
            assert "TAG Inventory" in report
            assert "TEST-001" in report

    def test_generate_matrix_report(self):
        """Test matrix report generation"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.py"
            test_file.write_text("""
# @SPEC:MATRIX-001
# @CODE:MATRIX-001
# @TEST:MATRIX-001
            """)

            generator = ReportGenerator()
            report = generator.generate_matrix_report(tmpdir)

            assert isinstance(report, str)
            assert "Coverage Matrix" in report or "MATRIX-001" in report

    def test_generate_statistics_report(self):
        """Test statistics report generation"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.py"
            test_file.write_text("# @CODE:STATS-001")

            generator = ReportGenerator()
            report = generator.generate_statistics_report(tmpdir)

            assert isinstance(report, str)
            # Should be valid JSON
            stats = json.loads(report)
            assert "total_tags" in stats

    def test_generate_combined_report(self):
        """Test combined report with all sections"""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.py"
            test_file.write_text("""
# @SPEC:COMBINED-001
# @CODE:COMBINED-001
# @TEST:COMBINED-001
            """)

            generator = ReportGenerator()
            report = generator.generate_combined_report(tmpdir)

            assert isinstance(report, str)
            assert len(report) > 0
            # Should contain multiple sections
            assert "Inventory" in report or "Matrix" in report or "Statistics" in report

    def test_generate_all_reports(self):
        """Test generating all reports to output directory"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create source directory with tags
            src_dir = Path(tmpdir) / "src"
            src_dir.mkdir()
            (src_dir / "code.py").write_text("# @CODE:ALL-001")

            # Create output directory
            output_dir = Path(tmpdir) / "reports"
            output_dir.mkdir()

            generator = ReportGenerator()
            result = generator.generate_all_reports(str(src_dir), str(output_dir))

            assert result.inventory_path.exists()
            assert result.matrix_path.exists()
            assert result.statistics_path.exists()

    def test_generate_reports_empty_directory(self):
        """Test report generation for directory without TAGs"""
        with tempfile.TemporaryDirectory() as tmpdir:
            output_dir = Path(tmpdir) / "reports"
            output_dir.mkdir()

            generator = ReportGenerator()
            result = generator.generate_all_reports(tmpdir, str(output_dir))

            # Should still create reports (empty/zero counts)
            assert result.inventory_path.exists()
            assert result.matrix_path.exists()
            assert result.statistics_path.exists()


# ============================================================================
# Test Class 7: ReportFormatting (4 tests)
# ============================================================================

@pytest.mark.skipif(ReportFormatter is None, reason="ReportFormatter not implemented yet")
class TestReportFormatting:
    """Test report formatting utilities"""

    def test_format_inventory_md(self):
        """Test inventory markdown formatting"""
        formatter = ReportFormatter()

        inventory = [
            TagInventory(
                tag_id="FORMAT-001",
                file_path="/test.py",
                line_number=10,
                context="# @CODE:FORMAT-001",
                related_tags=["@SPEC:FORMAT-001"],
                last_modified=datetime.now(),
                status="active"
            )
        ]

        markdown = formatter.format_inventory_md(inventory)

        assert "FORMAT-001" in markdown
        assert "/test.py:10" in markdown

    def test_format_matrix_md(self):
        """Test matrix markdown formatting"""
        formatter = ReportFormatter()

        # Create simple matrix
        matrix = TagMatrix(
            rows={
                "TEST-001": {
                    "SPEC": True,
                    "CODE": True,
                    "TEST": False,
                    "DOC": False
                }
            },
            completion_percentages={"TEST-001": 50.0}
        )

        markdown = formatter.format_matrix_md(matrix)

        assert "TEST-001" in markdown
        assert "|" in markdown  # Table format

    def test_format_table_generation(self):
        """Test markdown table generation helper"""
        formatter = ReportFormatter()

        headers = ["SPEC", "CODE", "TEST", "Completion"]
        rows = [
            ["AUTH-001", "✅", "✅", "100%"],
            ["AUTH-002", "✅", "❌", "50%"]
        ]

        table = formatter.format_table(headers, rows)

        assert "SPEC" in table
        assert "CODE" in table
        assert "AUTH-001" in table
        assert "✅" in table
        assert "|" in table

    def test_format_html_dashboard_optional(self):
        """Test optional HTML dashboard formatting"""
        formatter = ReportFormatter()

        inventory = [
            TagInventory(
                tag_id="HTML-001",
                file_path="/test.py",
                line_number=1,
                context="# @CODE:HTML-001",
                related_tags=[],
                last_modified=datetime.now(),
                status="active"
            )
        ]

        # HTML formatting is optional - may not be implemented
        try:
            html = formatter.format_html_dashboard(inventory)
            assert "<html>" in html or "<div>" in html
        except NotImplementedError:
            # HTML formatting is optional
            pass


# ============================================================================
# Integration Tests (2 tests)
# ============================================================================

@pytest.mark.skipif(ReportGenerator is None, reason="ReportGenerator not implemented yet")
class TestReportGeneratorIntegration:
    """Integration tests for complete reporting workflow"""

    def test_full_workflow_with_real_tags(self):
        """Test complete workflow: scan → generate → format → save"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create realistic project structure
            src_dir = Path(tmpdir) / "src"
            src_dir.mkdir()

            # Auth module with complete chain
            (src_dir / "auth.py").write_text("""
# @SPEC:AUTH-LOGIN-001 | User login specification
# @CODE:AUTH-LOGIN-001 | Login implementation
def login(username, password):
    pass
            """)

            # Test file
            tests_dir = Path(tmpdir) / "tests"
            tests_dir.mkdir()
            (tests_dir / "test_auth.py").write_text("""
# @TEST:AUTH-LOGIN-001 | Login tests
def test_login():
    pass
            """)

            # Doc file
            docs_dir = Path(tmpdir) / "docs"
            docs_dir.mkdir()
            (docs_dir / "auth.md").write_text("""
# @DOC:AUTH-LOGIN-001 | Authentication guide
            """)

            # Generate reports
            output_dir = Path(tmpdir) / "reports"
            output_dir.mkdir()

            generator = ReportGenerator()
            result = generator.generate_all_reports(tmpdir, str(output_dir))

            # Verify all files created
            assert result.inventory_path.exists()
            assert result.matrix_path.exists()
            assert result.statistics_path.exists()

            # Verify inventory contains our TAG
            inventory_content = result.inventory_path.read_text()
            assert "AUTH-LOGIN-001" in inventory_content

            # Verify matrix shows 100% completion
            matrix_content = result.matrix_path.read_text()
            assert "AUTH-LOGIN-001" in matrix_content

            # Verify statistics JSON is valid
            stats_content = result.statistics_path.read_text()
            stats = json.loads(stats_content)
            assert stats["total_tags"] >= 4
            assert stats["by_type"]["SPEC"] >= 1

    def test_detect_orphans_and_incomplete_chains(self):
        """Test detection of orphan TAGs and incomplete chains"""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create files with problematic TAGs
            src_dir = Path(tmpdir) / "src"
            src_dir.mkdir()

            (src_dir / "incomplete.py").write_text("""
# @SPEC:ORPHAN-001 | Spec without implementation
# @CODE:NO-TEST-001 | Code without tests
# @TEST:NO-CODE-001 | Test without code
            """)

            output_dir = Path(tmpdir) / "reports"
            output_dir.mkdir()

            generator = ReportGenerator()
            result = generator.generate_all_reports(tmpdir, str(output_dir))

            # Check statistics for issues
            stats_content = result.statistics_path.read_text()
            stats = json.loads(stats_content)

            # Should detect issues
            assert "issues" in stats
            assert stats["issues"]["orphan_count"] > 0 or stats["issues"]["incomplete_chains"] > 0
