#!/usr/bin/env python3
"""
@TEST:VERSION-SYNC-MODULARIZATION-001
Test suite for version sync system modularization according to TRUST principles

RED Phase: Creating failing tests for modularized version sync components
"""

import pytest
from pathlib import Path
from unittest.mock import Mock, patch


class TestVersionPatterns:
    """@TEST:VERSION-PATTERNS-001 Test version patterns module"""

    def test_should_create_version_patterns_provider(self):
        """Should create a dedicated version patterns provider"""
        # This will fail initially - we need to create the module
        from moai_adk.core.version_sync.version_patterns import VersionPatternsProvider

        provider = VersionPatternsProvider("0.1.28")
        assert provider is not None
        assert provider.current_version == "0.1.28"

    def test_should_generate_file_specific_patterns(self):
        """Should generate patterns for different file types"""
        from moai_adk.core.version_sync.version_patterns import VersionPatternsProvider

        provider = VersionPatternsProvider("0.1.28")
        patterns = provider.get_patterns()

        # Should include main file types
        assert "pyproject.toml" in patterns
        assert "**/*.py" in patterns
        assert "**/*.md" in patterns
        assert "**/*.json" in patterns

    def test_should_provide_pattern_descriptions(self):
        """Should provide human-readable descriptions for each pattern"""
        from moai_adk.core.version_sync.version_patterns import VersionPatternsProvider

        provider = VersionPatternsProvider("0.1.28")
        patterns = provider.get_patterns()

        # Each pattern should have description
        for file_pattern, rules in patterns.items():
            for rule in rules:
                assert "description" in rule
                assert rule["description"]


class TestFileProcessor:
    """@TEST:FILE-PROCESSOR-001 Test file processing module"""

    def test_should_create_secure_file_processor(self):
        """Should create a file processor with security validation"""
        from moai_adk.core.version_sync.file_processor import SecureFileProcessor

        processor = SecureFileProcessor(Path("/tmp/test"))
        assert processor is not None
        assert processor.project_root == Path("/tmp/test")

    def test_should_skip_dangerous_files(self):
        """Should skip files in dangerous directories"""
        from moai_adk.core.version_sync.file_processor import SecureFileProcessor

        processor = SecureFileProcessor(Path("/tmp/test"))

        # Should skip dangerous paths
        assert processor.should_skip_file(Path("/tmp/test/.git/config"))
        assert processor.should_skip_file(Path("/tmp/test/__pycache__/file.pyc"))
        assert processor.should_skip_file(Path("/tmp/test/node_modules/package.json"))

        # Should allow safe files
        assert not processor.should_skip_file(Path("/tmp/test/src/main.py"))

    def test_should_process_file_with_replacements(self):
        """Should process file content with version replacements"""
        from moai_adk.core.version_sync.file_processor import SecureFileProcessor

        processor = SecureFileProcessor(Path("/tmp/test"))

        # This should work with mock file content
        content = 'version = "0.1.0"'
        rules = [{"pattern": r'version = "[^"]*"', "replacement": 'version = "0.1.28"'}]

        result = processor.apply_replacements(content, rules)
        assert result == 'version = "0.1.28"'


class TestSyncExecutor:
    """@TEST:SYNC-EXECUTOR-001 Test sync execution orchestrator"""

    def test_should_create_sync_executor(self):
        """Should create sync executor with dependencies"""
        from moai_adk.core.version_sync.sync_executor import SyncExecutor

        executor = SyncExecutor(Path("/tmp/test"), "0.1.28")
        assert executor is not None
        assert executor.current_version == "0.1.28"

    def test_should_orchestrate_full_sync_process(self):
        """Should orchestrate the complete sync process"""
        from moai_adk.core.version_sync.sync_executor import SyncExecutor

        executor = SyncExecutor(Path("/tmp/test"), "0.1.28")

        # Should return sync results
        results = executor.sync_all_versions(dry_run=True)
        assert isinstance(results, dict)

    def test_should_handle_dry_run_mode(self):
        """Should support dry-run mode without actual changes"""
        from moai_adk.core.version_sync.sync_executor import SyncExecutor

        executor = SyncExecutor(Path("/tmp/test"), "0.1.28")

        # Dry run should not make actual changes
        results = executor.sync_all_versions(dry_run=True)
        assert isinstance(results, dict)


class TestSyncValidator:
    """@TEST:SYNC-VALIDATOR-001 Test sync validation module"""

    def test_should_create_sync_validator(self):
        """Should create validator for version consistency checks"""
        from moai_adk.core.version_sync.sync_validator import SyncValidator

        validator = SyncValidator(Path("/tmp/test"), "0.1.28")
        assert validator is not None
        assert validator.current_version == "0.1.28"

    def test_should_detect_version_mismatches(self):
        """Should detect files with incorrect version information"""
        from moai_adk.core.version_sync.sync_validator import SyncValidator

        validator = SyncValidator(Path("/tmp/test"), "0.1.28")

        # Should return mismatch information
        mismatches = validator.verify_sync()
        assert isinstance(mismatches, dict)

    def test_should_validate_specific_patterns(self):
        """Should validate specific version patterns"""
        from moai_adk.core.version_sync.sync_validator import SyncValidator

        validator = SyncValidator(Path("/tmp/test"), "0.1.28")

        # Should find mismatches for specific pattern
        pattern = r"v[0-9]+\.[0-9]+\.[0-9]+"
        mismatches = validator.find_pattern_mismatches(pattern, "v0.1.28")
        assert isinstance(mismatches, list)


class TestScriptGenerator:
    """@TEST:SCRIPT-GENERATOR-001 Test automated script generation"""

    def test_should_create_script_generator(self):
        """Should create script generator for automation"""
        from moai_adk.core.version_sync.script_generator import ScriptGenerator

        generator = ScriptGenerator(Path("/tmp/test"))
        assert generator is not None
        assert generator.project_root == Path("/tmp/test")

    def test_should_generate_update_script(self):
        """Should generate version update automation script"""
        from moai_adk.core.version_sync.script_generator import ScriptGenerator

        generator = ScriptGenerator(Path("/tmp/test"))

        # Should return script path
        script_path = generator.create_version_update_script()
        assert isinstance(script_path, str)
        assert "update_version.py" in script_path

    def test_should_create_executable_script(self):
        """Should create executable Python script"""
        from moai_adk.core.version_sync.script_generator import ScriptGenerator

        generator = ScriptGenerator(Path("/tmp/test"))

        # Generated script should be valid Python
        script_path = generator.create_version_update_script()
        assert Path(script_path).exists()


class TestVersionSyncIntegration:
    """@TEST:INTEGRATION-001 Test integration between all modules"""

    def test_should_preserve_existing_api(self):
        """Should preserve existing VersionSyncManager API for backward compatibility"""
        from moai_adk.core.version_sync import VersionSyncManager

        # Original API should still work
        manager = VersionSyncManager()
        assert hasattr(manager, 'sync_all_versions')
        assert hasattr(manager, 'verify_sync')
        assert hasattr(manager, 'create_version_update_script')

    def test_should_maintain_functionality(self):
        """Should maintain all original functionality"""
        from moai_adk.core.version_sync import VersionSyncManager

        manager = VersionSyncManager()

        # Should work with dry run
        results = manager.sync_all_versions(dry_run=True)
        assert isinstance(results, dict)

        # Should verify sync
        verification = manager.verify_sync()
        assert isinstance(verification, dict)

    @pytest.mark.parametrize("pattern,expected_files", [
        ("**/*.py", "python files"),
        ("**/*.md", "markdown files"),
        ("**/*.json", "json files"),
    ])
    def test_should_handle_different_file_patterns(self, pattern, expected_files):
        """Should handle different file patterns correctly"""
        from moai_adk.core.version_sync import VersionSyncManager

        manager = VersionSyncManager()

        # Should not crash on different patterns
        results = manager.sync_all_versions(dry_run=True)
        assert isinstance(results, dict)


# TRUST Principles Compliance Tests
class TestTRUSTPrinciples:
    """@TEST:TRUST-COMPLIANCE-001 Verify TRUST principles compliance"""

    def test_modules_should_be_under_size_limit(self):
        """Each module should be under 300 LOC (TRUST-U: Unified)"""
        import inspect
        from pathlib import Path

        # Check each module file size when created
        version_sync_dir = Path(__file__).parent.parent / "src" / "moai_adk" / "core" / "version_sync"

        if version_sync_dir.exists():
            for py_file in version_sync_dir.glob("*.py"):
                if py_file.name != "__init__.py":
                    lines = py_file.read_text().count('\n')
                    assert lines <= 300, f"{py_file.name} has {lines} lines, should be ≤300"

    def test_modules_should_have_single_responsibility(self):
        """Each module should have single, clear responsibility (TRUST-R: Readable)"""
        # Pattern provider: only patterns
        from moai_adk.core.version_sync.version_patterns import VersionPatternsProvider
        provider = VersionPatternsProvider("0.1.28")

        # Should only provide patterns, not process files
        assert hasattr(provider, 'get_patterns')
        assert not hasattr(provider, 'process_file')
        assert not hasattr(provider, 'verify_sync')

    def test_security_validation_should_be_isolated(self):
        """Security validation should be isolated (TRUST-S: Secured)"""
        from moai_adk.core.version_sync.file_processor import SecureFileProcessor

        processor = SecureFileProcessor(Path("/tmp/test"))

        # Should have dedicated security methods
        assert hasattr(processor, 'should_skip_file')
        # Should validate dangerous paths
        assert processor.should_skip_file(Path("/tmp/test/.git/secret"))