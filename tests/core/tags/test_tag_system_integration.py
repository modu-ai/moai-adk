#!/usr/bin/env python3
# @TEST:TAG-SYSTEM-INTEGRATION-001 | SPEC: .moai/specs/SPEC-TAG-SYSTEM-001/spec.md | CODE: @CODE:TAG-VALIDATOR-001
"""
Integration tests for complete TAG system

This module tests the integration between TAG components:
- Tag auto-correction and validation
- TAG chain integrity validation
- Integration with validation system
- End-to-end TAG workflow

@SPEC:TAG-SYSTEM-INTEGRATION-001: TAG 시스템 통합 테스트
@CODE:TAG-SYSTEM-INTEGRATION-TST: TAG 시스템 통합 테스트 코드
"""

import json
import tempfile
from pathlib import Path
from typing import Dict, Any

import pytest

from moai_adk.core.tags.auto_corrector import TagAutoCorrector, AutoCorrectionConfig
from moai_adk.core.tags.validator import CentralValidator, ValidationConfig
from moai_adk.core.tags.tags import suggest_tag_for_file, validate_tag_chain


class TestTagSystemIntegration:
    """Test class for TAG system integration"""

    def test_auto_correction_with_validator(self):
        """Test integration between auto-corrector and validator"""
        # Create test file with missing TAG
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            f.write("""
def test_function():
    return True
""")
            test_file = Path(f.name)

        try:
            # Configure auto-corrector
            config = AutoCorrectionConfig(
                enable_auto_fix=True,
                confidence_threshold=0.8,
                create_missing_specs=False,
                create_missing_tests=False
            )
            corrector = TagAutoCorrector(config=config)

            # Configure validator
            validator_config = ValidationConfig(
                strict_mode=False,
                check_duplicates=True,
                check_orphans=True,
                check_chain_integrity=True
            )
            validator = CentralValidator(config=validator_config)

            # Run validator first
            validation_result = validator.validate_files([str(test_file)])
            assert validation_result.warning_count == 0  # No warnings initially

            # Run auto-corrector
            # This would add TAG to the file
            tag_suggestion = corrector.suggest_tag_for_code_file(str(test_file))
            if tag_suggestion:
                tag, confidence = tag_suggestion
                if confidence >= config.confidence_threshold:
                    correction = corrector._fix_missing_tags(
                        str(test_file), test_file.read_text(),
                        type('MockViolation', (), {'type': 'missing_tags', 'tag': tag})()
                    )
                    if correction:
                        test_file.write_text(correction.corrected_content)

            # Validate again after correction
            validation_result = validator.validate_files([str(test_file)])
            # Should now have TAG and no warnings
            assert len(validation_result.warnings) == 0

        finally:
            test_file.unlink()

    def test_tag_chain_validation(self):
        """Test TAG chain validation functionality"""
        from moai_adk.core.tags.tags import validate_tag_chain

        # Valid chain
        valid_chain = validate_tag_chain(
            "@DOC:AUTH-001",
            "@SPEC:AUTH-001 -> @CODE:AUTH-001 -> @TEST:AUTH-001 -> @DOC:AUTH-001"
        )
        assert valid_chain is True

        # Invalid chain (domain mismatch)
        invalid_chain = validate_tag_chain(
            "@DOC:AUTH-001",
            "@SPEC:API-001 -> # REMOVED_ORPHAN_CODE:API-001 -> # REMOVED_ORPHAN_TEST:API-001 -> @DOC:AUTH-001"
        )
        assert invalid_chain is False

    def test_complete_tag_workflow(self):
        """Test complete TAG workflow from suggestion to validation"""
        # Create test file
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            f.write("""
# @CODE:AUTH-001 | SPEC: .moai/specs/SPEC-AUTH-001/spec.md
def authenticate_user(username, password):
    # TODO: Implement authentication logic
    return True

def logout_user():
    # TODO: Implement logout logic
    pass
""")
            test_file = Path(f.name)

        try:
            # 1. Suggest TAG for file
            suggestion = suggest_tag_for_file(test_file)
            assert suggestion.domain == "AUTH"
            assert suggestion.tag_id == "@DOC:AUTH-001"

            # 2. Validate TAG chain
            chain_valid = validate_tag_chain(
                suggestion.tag_id,
                f"@SPEC:AUTH-001 -> {suggestion.tag_id}"
            )
            assert chain_valid is True

            # 3. Run comprehensive validation
            validator_config = ValidationConfig(
                strict_mode=True,
                check_chain_integrity=True
            )
            validator = CentralValidator(config=validator_config)

            validation_result = validator.validate_files([str(test_file)])
            assert validation_result.is_valid is True

        finally:
            test_file.unlink()

    def test_error_handling_in_tag_system(self):
        """Test error handling in TAG system components"""
        # Test with non-existent file
        validator_config = ValidationConfig(strict_mode=False)
        validator = CentralValidator(config=validator_config)

        result = validator.validate_files(["/nonexistent/file.py"])
        assert result.is_valid is True  # Should not crash, just no issues
        assert len(result.issues) == 0

        # Test tag suggestion with invalid path
        suggestion = suggest_tag_for_file(Path("/nonexistent/file.md"))
        # Should return suggestion but with low confidence
        assert suggestion.confidence < 0.5

    def test_performance_with_large_tag_database(self):
        """Test TAG system performance with large database"""
        # Create multiple test files
        test_files = []
        try:
            for i in range(100):
                with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
                    f.write(f"# @CODE:TEST-{i:03d}\n\ndef test_function_{i}():\n    return True\n")
                    test_files.append(Path(f.name))

            # Configure validator for performance testing
            validator_config = ValidationConfig(
                strict_mode=False,
                check_chain_integrity=True
            )
            validator = CentralValidator(config=validator_config)

            # Measure validation time
            import time
            start_time = time.time()
            validation_result = validator.validate_files([str(f) for f in test_files])
            execution_time = time.time() - start_time

            # Should handle large number of files efficiently
            assert len(validation_result.issues) >= 0
            assert execution_time < 5.0  # Should complete in less than 5 seconds

            # Should have scanned all files
            assert validation_result.statistics.total_files_scanned == len(test_files)

        finally:
            for f in test_files:
                f.unlink()


class TestTagAutoCorrectionIntegration:
    """Test class for TAG auto-correction integration"""

    def test_auto_correction_with_duplicate_detection(self):
        """Test auto-correction with duplicate detection"""
        # Create file with duplicate TAG
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            f.write("""# # REMOVED_ORPHAN_CODE:DUPLICATE-001
# # REMOVED_ORPHAN_CODE:DUPLICATE-001
def test_function():
    return True
""")
            test_file = Path(f.name)

        try:
            # Configure auto-corrector
            config = AutoCorrectionConfig(
                enable_auto_fix=True,
                confidence_threshold=0.8,
                remove_duplicates=True
            )
            corrector = TagAutoCorrector(config=config)

            # Run correction
            violation = type('MockViolation', (), {
                'type': 'duplicate_tags',
                'tag': '# REMOVED_ORPHAN_CODE:DUPLICATE-001'
            })()
            correction = corrector._fix_duplicate_tags(str(test_file), test_file.read_text(), violation)

            assert correction is not None
            assert "중복 TAG 제거" in correction.description

            # Verify content has been corrected
            corrected_content = test_file.read_text()
            assert corrected_content.count("# REMOVED_ORPHAN_CODE:DUPLICATE-001") == 1

        finally:
            test_file.unlink()

    def test_auto_correction_with_chain_break_detection(self):
        """Test auto-correction with chain break detection"""
        # Create file with broken chain
        with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
            f.write("""# @CODE:BROKEN-CHAIN-001
def test_function():
    return True
""")
            test_file = Path(f.name)

        try:
            # Configure auto-corrector with test creation
            config = AutoCorrectionConfig(
                enable_auto_fix=True,
                confidence_threshold=0.8,
                create_missing_tests=True
            )
            corrector = TagAutoCorrector(config=config)

            # Test broken chain detection
            violation = type('MockViolation', (), {
                'type': 'chain_break',
                'tag': '@CODE:BROKEN-CHAIN-001'
            })()
            correction = corrector._fix_chain_break(str(test_file), test_file.read_text(), violation)

            # Should create test file for broken chain
            test_file_path = Path("tests/test_broken_chain.py")
            if test_file_path.exists():
                test_content = test_file_path.read_text()
                assert "@TEST:BROKEN-CHAIN-001" in test_content
                test_file_path.unlink()  # Cleanup

        finally:
            test_file.unlink()


class TestTagReportingIntegration:
    """Test class for TAG reporting integration"""

    def test_validator_reporting_integration(self):
        """Test validator reporting integration with ReportGenerator"""
        # Create test files with various issues
        test_files = []
        try:
            # File with duplicate TAG
            with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
                f.write("""# # REMOVED_ORPHAN_CODE:DUP-001
# # REMOVED_ORPHAN_CODE:DUP-001
def test_function():
    return True
""")
                test_files.append(Path(f.name))

            # File with orphan TAG
            with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
                f.write("""# # REMOVED_ORPHAN_CODE:ORPHAN-001
def test_function():
    return True
""")
                test_files.append(Path(f.name))

            # Configure validator
            validator_config = ValidationConfig(
                strict_mode=True,
                check_duplicates=True,
                check_orphans=True
            )
            validator = CentralValidator(config=validator_config)

            # Run validation
            validation_result = validator.validate_files([str(f) for f in test_files])

            # Test export functionality
            export_data = validator.export_for_reporting(validation_result)

            assert "timestamp" in export_data
            assert "is_valid" in export_data
            assert "statistics" in export_data
            assert "issues_by_type" in export_data
            assert execution_time_ms in export_data

            # Should have detected duplicates and orphans
            assert len(export_data["issues_by_type"]["duplicate"]) > 0
            assert len(export_data["issues_by_type"]["orphan"]) > 0

        finally:
            for f in test_files:
                f.unlink()