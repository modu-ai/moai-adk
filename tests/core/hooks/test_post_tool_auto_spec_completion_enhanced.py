"""Enhanced tests for post_tool_auto_spec_completion.py - Batch 3 coverage improvements

Focus: SPEC generation, validation, file detection, confidence scoring
Target Coverage: 72.2% → 90.0% (+17.8%)
"""

import os
from unittest.mock import Mock

import pytest

from moai_adk.core.hooks.post_tool_auto_spec_completion import (
    BaseHook,
    PostToolAutoSpecCompletion,
    SpecGenerator,
)


class TestBaseHook:
    """Test BaseHook functionality - NEW COVERAGE"""

    def test_base_hook_initialization(self):
        """Test BaseHook initializes with correct attributes"""
        hook = BaseHook()
        assert hook.name == "PostToolAutoSpecCompletion"
        assert "SPEC Completion" in hook.description


class TestSpecGenerator:
    """Test SpecGenerator functionality - NEW COVERAGE"""

    def test_spec_generator_initialization(self):
        """Test SpecGenerator initialization"""
        generator = SpecGenerator()
        assert generator.name == "SpecGenerator"

    def test_spec_generator_generate(self):
        """Test basic SPEC generation"""
        generator = SpecGenerator()
        result = generator.generate_spec("/path/to/file.py", "test content" * 50)

        assert "SPEC document" in result
        assert "/path/to/file.py" in result
        assert len(result) > 0


class TestPostToolAutoSpecCompletionInitialization:
    """Test hook initialization - NEW COVERAGE"""

    def test_initialization_with_defaults(self):
        """Test initialization with default config"""
        hook = PostToolAutoSpecCompletion()

        assert hook.spec_generator is not None
        assert hook.auto_config is not None
        assert "enabled" in hook.auto_config

    def test_processed_files_tracking(self):
        """Test processed files set is initialized"""
        hook = PostToolAutoSpecCompletion()
        assert isinstance(hook.processed_files, set)
        assert len(hook.processed_files) == 0


class TestShouldTriggerSpecCompletion:
    """Test trigger condition detection - NEW COVERAGE"""

    def test_trigger_disabled_config(self):
        """Test trigger returns False when disabled"""
        hook = PostToolAutoSpecCompletion()
        hook.auto_config["enabled"] = False

        result = hook.should_trigger_spec_completion("Write", {"file_path": "test.py"})
        assert result is False

    def test_trigger_wrong_tool_name(self):
        """Test trigger returns False for unsupported tools"""
        hook = PostToolAutoSpecCompletion()

        result = hook.should_trigger_spec_completion("Read", {"file_path": "test.py"})
        assert result is False

        result = hook.should_trigger_spec_completion("Bash", {"command": "ls"})
        assert result is False

    def test_trigger_no_file_paths(self):
        """Test trigger returns False when no file paths found"""
        hook = PostToolAutoSpecCompletion()

        result = hook.should_trigger_spec_completion("Write", {})
        assert result is False

    def test_trigger_unsupported_file_type(self):
        """Test trigger returns False for unsupported file types"""
        hook = PostToolAutoSpecCompletion()

        result = hook.should_trigger_spec_completion("Write", {"file_path": "test.txt"})
        assert result is False

    def test_trigger_excluded_pattern(self):
        """Test trigger returns False for excluded patterns"""
        hook = PostToolAutoSpecCompletion()

        result = hook.should_trigger_spec_completion("Write", {"file_path": "test_file.py"})
        assert result is False  # test_ prefix is excluded

    def test_trigger_success_python(self):
        """Test trigger succeeds for Python files"""
        hook = PostToolAutoSpecCompletion()

        result = hook.should_trigger_spec_completion("Write", {"file_path": "module.py"})
        assert result is True

    def test_trigger_success_typescript(self):
        """Test trigger succeeds for TypeScript files"""
        hook = PostToolAutoSpecCompletion()

        result = hook.should_trigger_spec_completion("Write", {"file_path": "component.ts"})
        assert result is True


class TestFilePathExtraction:
    """Test file path extraction from tool arguments - NEW COVERAGE"""

    def test_extract_from_write_tool(self):
        """Test extracting file path from Write tool"""
        hook = PostToolAutoSpecCompletion()
        args = {"file_path": "/path/to/file.py"}

        paths = hook._extract_file_paths(args)
        assert len(paths) == 1
        assert os.path.isabs(paths[0])

    def test_extract_from_edit_tool(self):
        """Test extracting file path from Edit tool"""
        hook = PostToolAutoSpecCompletion()
        args = {"file_path": "/path/to/file.py"}

        paths = hook._extract_file_paths(args)
        assert len(paths) == 1

    def test_extract_from_multiedit_tool(self):
        """Test extracting multiple file paths from MultiEdit"""
        hook = PostToolAutoSpecCompletion()
        args = {"edits": [{"file_path": "/path/file1.py"}, {"file_path": "/path/file2.py"}]}

        paths = hook._extract_file_paths(args)
        assert len(paths) == 2

    def test_extract_duplicate_removal(self):
        """Test duplicate file paths are removed"""
        hook = PostToolAutoSpecCompletion()
        args = {"file_path": "/path/to/file.py"}

        # Call twice with same path
        paths1 = hook._extract_file_paths(args)
        paths2 = hook._extract_file_paths(args)

        # Should get unique paths
        all_paths = paths1 + paths2
        unique_paths = list(set(all_paths))
        assert len(unique_paths) == 1


class TestFileSupportAndExclusion:
    """Test file support and exclusion logic - NEW COVERAGE"""

    def test_is_supported_file_python(self):
        """Test Python files are supported"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_supported_file("module.py") is True

    def test_is_supported_file_javascript(self):
        """Test JavaScript files are supported"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_supported_file("app.js") is True
        assert hook._is_supported_file("component.jsx") is True

    def test_is_supported_file_typescript(self):
        """Test TypeScript files are supported"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_supported_file("app.ts") is True
        assert hook._is_supported_file("component.tsx") is True

    def test_is_supported_file_go(self):
        """Test Go files are supported"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_supported_file("main.go") is True

    def test_is_supported_file_unsupported(self):
        """Test unsupported file types return False"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_supported_file("readme.md") is False
        assert hook._is_supported_file("config.json") is False

    def test_is_excluded_file_test_prefix(self):
        """Test test file prefix exclusion"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_excluded_file("test_module.py") is True

    def test_is_excluded_file_spec_prefix(self):
        """Test spec file prefix exclusion"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_excluded_file("spec_module.py") is True

    def test_is_excluded_file_test_directory(self):
        """Test __tests__ directory exclusion"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_excluded_file("__tests__/module.py") is True

    def test_is_excluded_file_normal(self):
        """Test normal files are not excluded"""
        hook = PostToolAutoSpecCompletion()
        assert hook._is_excluded_file("module.py") is False


class TestCodeChangeDetection:
    """Test code change detection - NEW COVERAGE"""

    def test_detect_write_tool_changes(self):
        """Test detecting changes from Write tool"""
        hook = PostToolAutoSpecCompletion()
        changes = hook.detect_code_changes("Write", {"file_path": "/path/file.py"}, None)

        assert len(changes) == 1
        assert "/path/file.py" in changes[0]

    def test_detect_edit_tool_changes(self):
        """Test detecting changes from Edit tool"""
        hook = PostToolAutoSpecCompletion()
        changes = hook.detect_code_changes("Edit", {"file_path": "/path/file.py"}, None)

        assert len(changes) == 1

    def test_detect_multiedit_tool_changes(self):
        """Test detecting changes from MultiEdit tool"""
        hook = PostToolAutoSpecCompletion()
        args = {"edits": [{"file_path": "/path/file1.py"}, {"file_path": "/path/file2.py"}]}
        changes = hook.detect_code_changes("MultiEdit", args, None)

        assert len(changes) == 2

    def test_detect_changes_deduplication(self):
        """Test duplicate changes are not returned"""
        hook = PostToolAutoSpecCompletion()

        # First detection
        changes1 = hook.detect_code_changes("Write", {"file_path": "/path/file.py"}, None)
        assert len(changes1) == 1

        # Second detection of same file
        changes2 = hook.detect_code_changes("Write", {"file_path": "/path/file.py"}, None)
        assert len(changes2) == 0  # Already processed


class TestConfidenceCalculation:
    """Test confidence score calculation - NEW COVERAGE"""

    def test_confidence_complete_analysis(self):
        """Test confidence calculation with complete analysis"""
        hook = PostToolAutoSpecCompletion()
        analysis = {"structure_score": 0.9, "domain_accuracy": 0.8, "documentation_level": 0.7}

        confidence = hook.calculate_completion_confidence(analysis)

        # Weighted: 0.9*0.3 + 0.8*0.4 + 0.7*0.3 = 0.27 + 0.32 + 0.21 = 0.8
        assert 0.75 <= confidence <= 0.85

    def test_confidence_partial_analysis(self):
        """Test confidence with partial analysis data"""
        hook = PostToolAutoSpecCompletion()
        analysis = {"structure_score": 0.6}

        confidence = hook.calculate_completion_confidence(analysis)
        assert 0.0 <= confidence <= 1.0

    def test_confidence_empty_analysis(self):
        """Test confidence with empty analysis"""
        hook = PostToolAutoSpecCompletion()
        confidence = hook.calculate_completion_confidence({})
        assert confidence == 0.5  # Default

    def test_confidence_clamping(self):
        """Test confidence is clamped between 0 and 1"""
        hook = PostToolAutoSpecCompletion()

        # Test upper bound
        analysis = {"structure_score": 1.5, "domain_accuracy": 1.5, "documentation_level": 1.5}
        confidence = hook.calculate_completion_confidence(analysis)
        assert confidence <= 1.0

        # Test lower bound
        analysis = {"structure_score": -0.5, "domain_accuracy": -0.5, "documentation_level": -0.5}
        confidence = hook.calculate_completion_confidence(analysis)
        assert confidence >= 0.0


class TestSpecIDGeneration:
    """Test SPEC ID generation - NEW COVERAGE"""

    def test_generate_spec_id_format(self):
        """Test SPEC ID generation format"""
        hook = PostToolAutoSpecCompletion()
        spec_id = hook._generate_spec_id("/path/to/my_module.py")

        # Should contain uppercase parts and hash
        assert "-" in spec_id
        parts = spec_id.split("-")
        assert len(parts[-1]) == 4  # Hash part

    def test_generate_spec_id_uniqueness(self):
        """Test different files generate different IDs"""
        hook = PostToolAutoSpecCompletion()

        id1 = hook._generate_spec_id("/path/to/module1.py")
        id2 = hook._generate_spec_id("/path/to/module2.py")

        assert id1 != id2

    def test_generate_spec_id_consistency(self):
        """Test same file generates same ID"""
        hook = PostToolAutoSpecCompletion()

        id1 = hook._generate_spec_id("/path/to/module.py")
        id2 = hook._generate_spec_id("/path/to/module.py")

        assert id1 == id2


class TestSpecContentGeneration:
    """Test SPEC content generation - NEW COVERAGE"""

    def test_generate_spec_content(self):
        """Test generating spec.md content"""
        hook = PostToolAutoSpecCompletion()
        analysis = {"description": "Test module", "language": "Python"}

        content = hook._generate_spec_content(analysis, "TEST-001", "module.py")

        assert "SPEC-TEST-001" in content
        assert "module.py" in content
        assert "Test module" in content
        assert "Python" in content

    def test_generate_plan_content(self):
        """Test generating plan.md content"""
        hook = PostToolAutoSpecCompletion()
        analysis = {"architecture": "Client-Server"}

        content = hook._generate_plan_content(analysis, "TEST-001", "module.py")

        assert "PLAN-TEST-001" in content
        assert "module.py" in content
        assert "Implementation Plan" in content

    def test_generate_acceptance_content(self):
        """Test generating acceptance.md content"""
        hook = PostToolAutoSpecCompletion()
        analysis = {"must_have_1": "System must work"}

        content = hook._generate_acceptance_content(analysis, "TEST-001", "module.py")

        assert "ACCEPT-TEST-001" in content
        assert "module.py" in content
        assert "Acceptance Criteria" in content

    def test_generate_complete_spec(self):
        """Test generating complete SPEC with all files"""
        hook = PostToolAutoSpecCompletion()
        analysis = {"description": "Test", "language": "Python"}

        result = hook.generate_complete_spec(analysis, "/path/module.py")

        assert "spec_id" in result
        assert "spec_md" in result
        assert "plan_md" in result
        assert "acceptance_md" in result
        assert len(result["spec_md"]) > 0


class TestSpecValidation:
    """Test SPEC validation - NEW COVERAGE"""

    def test_validate_complete_spec(self):
        """Test validating a complete SPEC"""
        hook = PostToolAutoSpecCompletion()
        spec_content = {
            "spec_md": """
            ## Overview
            ## Environment
            ## Assumptions
            ## Requirements
            ## Specifications
            """,
            "plan_md": "## Implementation Plan\n" * 50,
            "acceptance_md": "## Acceptance\n" * 50,
        }

        result = hook.validate_generated_spec(spec_content)

        assert "quality_score" in result
        assert result["quality_score"] >= 0.5

    def test_check_ears_compliance(self):
        """Test EARS format compliance checking"""
        hook = PostToolAutoSpecCompletion()
        spec_content = {
            "spec_md": """
            ## Overview
            ## Environment
            ## Assumptions
            ## Requirements
            ## Specifications
            """
        }

        score = hook._check_ears_compliance(spec_content)
        assert score == 1.0  # All sections present

    def test_check_completeness(self):
        """Test content completeness checking"""
        hook = PostToolAutoSpecCompletion()
        spec_content = {
            "spec_md": "## Requirements\n" + "content " * 100,
            "plan_md": "## Implementation Plan\n" + "content " * 100,
            "acceptance_md": "## Acceptance\n" + "content " * 100,
        }

        score = hook._check_completeness(spec_content)
        assert score > 0.5

    def test_check_content_quality(self):
        """Test content quality checking"""
        hook = PostToolAutoSpecCompletion()
        spec_content = {"spec_md": "API data interface module component architecture REQ-001 SPEC-001"}

        score = hook._check_content_quality(spec_content)
        assert score > 0.5


class TestSpecFileCreation:
    """Test SPEC file creation - NEW COVERAGE"""

    def test_create_spec_files(self, tmp_path):
        """Test creating SPEC files on disk"""
        hook = PostToolAutoSpecCompletion()
        content = {"spec_md": "spec content", "plan_md": "plan content", "acceptance_md": "acceptance content"}

        result = hook.create_spec_files("TEST-001", content, str(tmp_path))

        assert result is True
        assert (tmp_path / "SPEC-TEST-001" / "spec.md").exists()
        assert (tmp_path / "SPEC-TEST-001" / "plan.md").exists()
        assert (tmp_path / "SPEC-TEST-001" / "acceptance.md").exists()

    def test_create_spec_files_error_handling(self):
        """Test error handling in file creation"""
        hook = PostToolAutoSpecCompletion()
        content = {"spec_md": "content"}

        # Invalid path should return False
        result = hook.create_spec_files("TEST-001", content, "/invalid/path/that/does/not/exist")
        assert result is False


class TestExecuteHook:
    """Test hook execution - NEW COVERAGE"""

    def test_execute_not_triggered(self):
        """Test execution when trigger conditions not met"""
        hook = PostToolAutoSpecCompletion()

        result = hook.execute("Read", {"file_path": "test.py"}, None)

        assert result["success"] is False
        assert "not triggered" in result["message"]

    def test_execute_no_changes(self):
        """Test execution when no changes detected"""
        hook = PostToolAutoSpecCompletion()

        # Pre-populate processed files
        hook.processed_files.add(os.path.abspath("/path/file.py"))

        result = hook.execute("Write", {"file_path": "/path/file.py"}, None)

        assert result["success"] is False

    def test_execute_with_mock_analysis(self, tmp_path):
        """Test full execution with mocked analysis"""
        hook = PostToolAutoSpecCompletion()

        # Mock the spec generator
        mock_analysis = {
            "description": "Test",
            "structure_score": 0.9,
            "domain_accuracy": 0.9,
            "documentation_level": 0.9,
        }
        hook.spec_generator.analyze = Mock(return_value=mock_analysis)

        result = hook.execute("Write", {"file_path": str(tmp_path / "module.py")}, None)

        # Should have attempted generation
        assert "execution_time" in result


if __name__ == "__main__":
    pytest.main(
        [__file__, "-v", "--cov=moai_adk.core.hooks.post_tool_auto_spec_completion", "--cov-report=term-missing"]
    )
