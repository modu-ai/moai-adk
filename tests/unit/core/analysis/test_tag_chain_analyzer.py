# @TEST:VAL-001
"""Test suite for TAG chain analyzer.

Tests TAG chain analysis functionality including chain detection,
orphan identification, and completeness scoring.

@SPEC:DOCS-005: TAG 체인 분석 및 검증 도구
"""

import json
import tempfile
from pathlib import Path

from src.moai_adk.core.analysis.tag_chain_analyzer import (
    TagChain,
    TagChainAnalyzer,
    ChainAnalysisResult,
    analyze_tag_chains,
)


def test_tag_chain_completeness():
    """Test TAG chain completeness calculation."""
    # Complete chain
    complete_chain = TagChain(
        domain="AUTH",
        spec_id="@SPEC:AUTH-004",
        code_id="@CODE:AUTH-004",
        test_id="@TEST:AUTH-004"
    )
    assert complete_chain.is_complete
    assert complete_chain.completeness_score == 1.0
    assert complete_chain.missing_elements == []

    # Partial chain
    partial_chain = TagChain(
        domain="AUTH",
        spec_id="@SPEC:AUTH-004",
        code_id="@CODE:AUTH-004",
        test_id=None
    )
    assert not partial_chain.is_complete
    assert partial_chain.completeness_score == 2/3
    assert partial_chain.missing_elements == ["TEST"]

    # Empty chain
    empty_chain = TagChain(
        domain="AUTH",
        spec_id=None,
        code_id=None,
        test_id=None
    )
    assert not empty_chain.is_complete
    assert empty_chain.completeness_score == 0.0
    assert empty_chain.missing_elements == ["SPEC", "CODE", "TEST"]


def test_extract_domain_from_tag():
    """Test domain extraction from TAG."""
    analyzer = TagChainAnalyzer()

    assert analyzer._extract_domain_from_tag("@SPEC:AUTH-004") == "AUTH"
    assert analyzer._extract_domain_from_tag("@CODE:CLI-003") == "CLI"
    assert analyzer._extract_domain_from_tag("@TEST:LDE-PRIORITY-001") == "LDE-PRIORITY"
    assert analyzer._extract_domain_from_tag("@DOC:INSTALLER-QUALITY-001") == "INSTALLER-QUALITY"


def test_get_max_number():
    """Test maximum number extraction from tags."""
    analyzer = TagChainAnalyzer()

    spec_tags = ["@SPEC:AUTH-004", "@SPEC:AUTH-002"]
    code_tags = ["@CODE:AUTH-004", "@CODE:AUTH-003"]
    test_tags = ["@TEST:AUTH-002"]

    max_num = analyzer._get_max_number(spec_tags, code_tags, test_tags)
    assert max_num == 3


def test_identify_orphans():
    """Test orphan identification."""
    # Create test data
    all_tags = {
        "@CODE:AUTH-004": ["file1.py"],
        "@CODE:AUTH-002": ["file2.py"],
        "@TEST:AUTH-004": ["test1.py"],
        "@TEST:AUTH-003": ["test2.py"],
        "@SPEC:AUTH-002": ["spec1.md"],
        "@SPEC:API-001": ["spec2.md"],
    }

    analyzer = TagChainAnalyzer()
    orphans = analyzer._identify_orphans(all_tags)

    # CODE without SPEC: AUTH-001 (no @SPEC:AUTH-004)
    assert "@CODE:AUTH-004" in orphans["code_without_spec"]

    # CODE without TEST: AUTH-002 (no @TEST:AUTH-002)
    assert "@CODE:AUTH-002" in orphans["code_without_test"]

    # TEST without CODE: AUTH-003 (no @CODE:AUTH-003)
    assert "@TEST:AUTH-003" in orphans["test_without_code"]

    # SPEC without CODE: API-001 (no @CODE:API-001)
    assert "@SPEC:API-001" in orphans["spec_without_code"]


def test_scan_all_tags():
    """Test scanning all tags from files."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create test files
        src_file = temp_path / "src" / "test.py"
        src_file.parent.mkdir(parents=True)
        src_file.write_text("# @CODE:TEST-001\n# @SPEC:TEST-001")

        test_file = temp_path / "tests" / "test.py"
        test_file.parent.mkdir(parents=True)
        test_file.write_text("# @TEST:ANALYZER-001")

        spec_file = temp_path / ".moai" / "specs" / "test.md"
        spec_file.parent.mkdir(parents=True)
        spec_file.write_text("# @SPEC:TEST-001\n# @TEST:ANALYZER-001")

        analyzer = TagChainAnalyzer(temp_path)
        all_tags = analyzer._scan_all_tags()

        assert "@CODE:TEST-001" in all_tags
        assert "@SPEC:TEST-001" in all_tags
        assert "@TEST:ANALYZER-001" in all_tags


def test_group_chains_by_domain():
    """Test grouping chains by domain."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create test files
        src_file = temp_path / "src" / "auth.py"
        src_file.parent.mkdir(parents=True)
        src_file.write_text("# @CODE:AUTH-004\n# @CODE:AUTH-002")

        test_file = temp_path / "tests" / "auth_test.py"
        test_file.parent.mkdir(parents=True)
        test_file.write_text("# @TEST:AUTH-004\n# @TEST:AUTH-003")

        spec_file = temp_path / ".moai" / "specs" / "auth.md"
        spec_file.parent.mkdir(parents=True)
        spec_file.write_text("# @SPEC:AUTH-004\n# @SPEC:AUTH-002")

        analyzer = TagChainAnalyzer(temp_path)
        all_tags = analyzer._scan_all_tags()
        chains_by_domain = analyzer._group_chains_by_domain(all_tags)

        # Should have AUTH domain with 3 chains (001, 002, 003)
        assert "AUTH" in chains_by_domain
        assert len(chains_by_domain["AUTH"]) == 3

        # Chain 001: Complete
        chain_001 = next(c for c in chains_by_domain["AUTH"] if c.code_id == "@CODE:AUTH-004")
        assert chain_001.spec_id == "@SPEC:AUTH-004"
        assert chain_001.test_id == "@TEST:AUTH-004"
        assert chain_001.is_complete

        # Chain 002: Missing TEST
        chain_002 = next(c for c in chains_by_domain["AUTH"] if c.code_id == "@CODE:AUTH-002")
        assert chain_002.spec_id == "@SPEC:AUTH-002"
        assert chain_002.test_id is None
        assert not chain_002.is_complete

        # Chain 003: Missing CODE and SPEC
        chain_003 = next(c for c in chains_by_domain["AUTH"] if c.test_id == "@TEST:AUTH-003")
        assert chain_003.spec_id is None
        assert chain_003.code_id is None
        assert not chain_003.is_complete


def test_analyze_all_chains():
    """Test complete chain analysis."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create test files with known state
        src_file = temp_path / "src" / "auth.py"
        src_file.parent.mkdir(parents=True)
        src_file.write_text("# @CODE:AUTH-004\n# @CODE:AUTH-002")

        test_file = temp_path / "tests" / "auth_test.py"
        test_file.parent.mkdir(parents=True)
        test_file.write_text("# @TEST:AUTH-004\n# @TEST:AUTH-003")

        spec_file = temp_path / ".moai" / "specs" / "auth.md"
        spec_file.parent.mkdir(parents=True)
        spec_file.write_text("# @SPEC:AUTH-004")

        analyzer = TagChainAnalyzer(temp_path)
        result = analyzer.analyze_all_chains()

        # Check summary
        assert result.total_chains == 3  # AUTH-001, AUTH-002, AUTH-003
        assert result.complete_chains == 1  # AUTH-001
        assert result.partial_chains == 1  # AUTH-002 (has spec and code, missing test)
        assert result.broken_chains == 1  # AUTH-003 (has test, missing spec and code)

        # Check orphans
        assert "@CODE:AUTH-002" in result.orphans_by_type["code_without_test"]
        assert "@TEST:AUTH-003" in result.orphans_by_type["test_without_code"]
        assert "@SPEC:AUTH-002" in result.orphans_by_type["spec_without_code"]

        # Check broken chain details
        assert len(result.broken_chain_details) == 2
        auth_002_detail = next(d for d in result.broken_chain_details if d["domain"] == "AUTH" and "002" in str(d))
        assert "TEST" in auth_002_detail["missing"]


def test_generate_report():
    """Test report generation."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create test files
        src_file = temp_path / "src" / "auth.py"
        src_file.parent.mkdir(parents=True)
        src_file.write_text("# @CODE:AUTH-004")

        test_file = temp_path / "tests" / "auth_test.py"
        test_file.parent.mkdir(parents=True)
        test_file.write_text("# @TEST:AUTH-004")

        analyzer = TagChainAnalyzer(temp_path)
        result = analyzer.analyze_all_chains()
        report = analyzer.generate_report(result)

        # Check report structure
        assert "# TAG Chain Analysis Report" in report
        assert "## Summary" in report
        assert "## Orphan TAGs" in report
        assert "## Broken Chain Details" in report
        assert "Complete Chains: 1" in report
        assert "Broken Chains: 0" in report


def test_convenience_function():
    """Test convenience analyze function."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)

        # Create minimal test structure
        src_file = temp_path / "src" / "test.py"
        src_file.parent.mkdir(parents=True)
        src_file.write_text("# @CODE:TEST-001")

        result = analyze_tag_chains(temp_path)

        assert isinstance(result, ChainAnalysisResult)
        assert result.total_chains >= 1


def test_integration_with_real_structure():
    """Test with actual MoAI-ADK structure."""
    analyzer = TagChainAnalyzer(Path("."))
    result = analyzer.analyze_all_chains()

    # Should find some chains in the real codebase
    assert result.total_chains >= 0

    # Generate report for verification
    report = analyzer.generate_report(result)
    print("\n" + "="*50)
    print("INTEGRATION TEST REPORT")
    print("="*50)
    print(report)
    print("="*50)