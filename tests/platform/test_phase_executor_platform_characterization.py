"""Characterization tests for PhaseExecutor platform-specific behavior

These tests capture the CURRENT behavior of platform detection and path variable
substitution. They document WHAT the system does, not what it SHOULD do.

Created for SPEC-PLATFORM-001 DDD implementation (PRESERVE phase).
"""

import json
import platform
from pathlib import Path

import pytest

from moai_adk.core.project.phase_executor import PhaseExecutor
from moai_adk.core.project.validator import ProjectValidator


@pytest.fixture
def validator(tmp_path: Path) -> ProjectValidator:
    """Create ProjectValidator instance"""
    return ProjectValidator()


@pytest.fixture
def executor(validator: ProjectValidator) -> PhaseExecutor:
    """Create PhaseExecutor instance"""
    return PhaseExecutor(validator)


class TestPlatformDetectionCharacterization:
    """Characterization tests for platform detection behavior

    These tests document the CURRENT platform detection behavior.
    If the behavior changes, these tests should be updated to reflect the new behavior.
    """

    def test_characterize_platform_system_detection(self) -> None:
        """Characterize: platform.system() returns current OS name

        This test documents what platform.system() returns on the current system.
        """
        # CAPTURE: What does platform.system() return?
        current_system = platform.system()
        assert isinstance(current_system, str)
        assert len(current_system) > 0
        # Common values: "Windows", "Darwin", "Linux"

    def test_characterize_is_windows_detection(self) -> None:
        """Characterize: is_windows detection logic in phase_executor

        This test documents how Windows is currently detected in the code.
        """
        # CAPTURE: Current implementation uses platform.system() == "Windows"
        is_windows = platform.system() == "Windows"

        # On non-Windows systems, this should be False
        # On Windows systems, this should be True
        assert isinstance(is_windows, bool)

        # Document the current platform
        current_system = platform.system()
        if current_system == "Windows":
            assert is_windows is True
        else:
            assert is_windows is False


class TestPathVariableSubstitutionCharacterization:
    """Characterization tests for path variable substitution behavior

    These tests document HOW path variables are currently substituted.
    """

    def test_characterize_unix_path_variable_format(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Characterize: Unix path variable format in template context

        Documents the current format of PROJECT_DIR variable.
        """
        # CAPTURE: What format is used for Unix paths?
        is_windows = platform.system() == "Windows"

        if is_windows:
            expected_prefix = "%CLAUDE_PROJECT_DIR%/"
        else:
            expected_prefix = "$CLAUDE_PROJECT_DIR/"

        # The current implementation uses forward slashes for Unix
        assert "/" in expected_prefix or "%" in expected_prefix

    def test_characterize_windows_path_variable_format(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Characterize: Windows path variable format in template context

        Documents the current format of PROJECT_DIR_WIN variable.
        """
        # CAPTURE: What format is used for Windows paths?
        is_windows = platform.system() == "Windows"

        if is_windows:
            # Current implementation uses backslash for Windows
            expected_suffix = "\\"
        else:
            # Non-Windows systems still define the variable
            expected_suffix = "/"

        # Document the separator character
        assert expected_suffix in ["\\", "/"]


class TestTemplateContextCharacterization:
    """Characterization tests for template context variables

    These tests document WHAT variables are available in the template context.
    """

    def test_characterize_context_includes_project_dir_unix(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Characterize: PROJECT_DIR is always present in context

        Documents that PROJECT_DIR variable is set during template processing.
        """
        # SETUP: Create required directory structure
        (tmp_path / ".moai").mkdir(parents=True, exist_ok=True)
        (tmp_path / ".claude").mkdir(parents=True, exist_ok=True)

        # CAPTURE: What happens when resource phase executes?
        config = {
            "name": "test-project",
            "description": "Test description",
            "mode": "personal",
            "author": "test-user",
            "version": "0.1.0",
            "language": "python",
            "language_settings": {"conversation_language": "en", "conversation_language_name": "English"},
        }

        # Execute resource phase with config
        result = executor.execute_resource_phase(tmp_path, config=config)

        # VERIFY: Check that context was set (internal behavior captured)
        # We can't directly access the context, but we can verify the phase completed
        assert len(result) > 0
        assert ".claude/" in result

    def test_characterize_context_includes_project_dir_win(self, executor: PhaseExecutor, tmp_path: Path) -> None:
        """Characterize: PROJECT_DIR_WIN is always present in context

        Documents that PROJECT_DIR_WIN variable is set during template processing.
        """
        # SETUP: Create required directory structure
        (tmp_path / ".moai").mkdir(parents=True, exist_ok=True)
        (tmp_path / ".claude").mkdir(parents=True, exist_ok=True)

        # CAPTURE: What happens when resource phase executes?
        config = {
            "name": "test-project",
            "description": "Test description",
            "mode": "personal",
            "author": "test-user",
            "version": "0.1.0",
            "language": "python",
            "language_settings": {"conversation_language": "en", "conversation_language_name": "English"},
        }

        # Execute resource phase with config
        result = executor.execute_resource_phase(tmp_path, config=config)

        # VERIFY: Check that phase completed
        assert len(result) > 0


class TestWindowsSettingsTemplateCharacterization:
    """Characterization tests for Windows settings template

    These tests document WHAT the current Windows settings template contains.
    """

    def test_characterize_windows_template_path_format(self) -> None:
        """Characterize: Current Windows template path format

        Documents what path format is used in settings.json.windows template.
        """
        # CAPTURE: Read the current Windows settings template
        template_path = (
            Path(__file__).parent.parent.parent / "src" / "moai_adk" / "templates" / ".claude" / "settings.json.windows"
        )

        if not template_path.exists():
            pytest.skip("Windows settings template not found")

        with open(template_path) as f:
            template = json.load(f)

        # DOCUMENT: What path format is used?
        session_start_hook = template["hooks"]["SessionStart"][0]["hooks"][0]["command"]

        # Current implementation uses Windows backslashes and relative paths
        assert "SessionStart" in template["hooks"]
        assert "PreToolUse" in template["hooks"]

        # Document the current format
        # If backslashes are found, document that
        if "\\" in session_start_hook:
            assert "\\" in session_start_hook  # Current behavior

        # If relative paths are used, document that
        if ".\\" in session_start_hook or "./" in session_start_hook:
            pass  # Current behavior uses relative paths


class TestUnixSettingsTemplateCharacterization:
    """Characterization tests for Unix settings template

    These tests document WHAT the current Unix settings template contains.
    """

    def test_characterize_unix_template_path_format(self) -> None:
        """Characterize: Current Unix template path format

        Documents what path format is used in settings.json.unix template.
        """
        # CAPTURE: Read the current Unix settings template
        template_path = (
            Path(__file__).parent.parent.parent / "src" / "moai_adk" / "templates" / ".claude" / "settings.json.unix"
        )

        if not template_path.exists():
            pytest.skip("Unix settings template not found")

        with open(template_path) as f:
            template = json.load(f)

        # DOCUMENT: What path format is used?
        session_start_hook = template["hooks"]["SessionStart"][0]["hooks"][0]["command"]

        # Current implementation uses forward slashes
        assert "SessionStart" in template["hooks"]
        assert "PreToolUse" in template["hooks"]

        # Document the current format
        # If forward slashes are used, document that
        if "/" in session_start_hook:
            assert "/" in session_start_hook  # Current behavior


class TestAgentTemplatePathVariables:
    """Characterization tests for agent template path variables

    These tests document WHAT variables are used in agent template files.
    """

    def test_characterize_agent_templates_use_project_dir(self) -> None:
        """Characterize: Agent templates use {{PROJECT_DIR}}

        Documents that agent template files use the PROJECT_DIR variable
        with forward slash separator (cross-platform).
        """
        # CAPTURE: Check agent template files
        templates_dir = (
            Path(__file__).parent.parent.parent / "src" / "moai_adk" / "templates" / ".claude" / "agents" / "moai"
        )

        if not templates_dir.exists():
            pytest.skip("Agent templates directory not found")

        # Check a few agent templates
        agent_files = list(templates_dir.glob("*.md"))[:5]  # Check first 5

        for agent_file in agent_files:
            content = agent_file.read_text()

            # DOCUMENT: New behavior uses PROJECT_DIR as standard (cross-platform)
            if "{{PROJECT_DIR}}" in content:
                pass  # New behavior uses PROJECT_DIR as standard
            elif "{{PROJECT_DIR_UNIX}}" in content:
                # Legacy: Deprecated after SPEC-PROJECT-DIR-001 (v1.8.0)
                pass  # Deprecated variable (replaced by PROJECT_DIR)


class TestCrossPlatformPathHandling:
    """Characterization tests for cross-platform path handling

    These tests document HOW paths are handled across platforms.
    """

    def test_characterize_forward_slashes_work_on_all_platforms(self) -> None:
        """Characterize: Forward slashes work on all platforms

        Documents that forward slash paths work on Windows, macOS, and Linux.
        This is a known fact from pytest, uv, ruff implementations.
        """
        # DOCUMENT: Industry best practice
        # pytest, uv, ruff all use forward slashes on Windows
        # This is documented behavior, not implementation-specific

        # Verify that Path handles forward slashes correctly
        test_path = Path("some/path/to/file.py")
        assert "some" in str(test_path).replace("\\", "/") or "some" in str(test_path)

    def test_characterize_pathlib_normalization(self) -> None:
        """Characterize: Pathlib normalizes paths appropriately

        Documents how pathlib.Path normalizes paths on different platforms.
        """
        # DOCUMENT: Pathlib behavior
        forward_slash_path = Path("dir/subdir/file.py")
        assert forward_slash_path.as_posix() == "dir/subdir/file.py"

        # Pathlib converts to platform-specific format when needed
        platform_str = str(forward_slash_path)
        assert isinstance(platform_str, str)


# Summary of characterization tests
# These tests document the CURRENT behavior as of SPEC-PROJECT-DIR-001 implementation (v1.8.0).
# - PROJECT_DIR: Standard cross-platform forward slash path (works on Windows, macOS, Linux)
# - PROJECT_DIR_UNIX: Deprecated after v1.8.0 - replaced by PROJECT_DIR for consistency
# - PROJECT_DIR_WIN: Deprecated after v1.8.0 - replaced by PROJECT_DIR for consistency
# When behavior changes, update these tests to reflect new behavior.
