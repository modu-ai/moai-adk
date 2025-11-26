"""Unit tests for checker.py module

Tests for SystemChecker class and check_environment function.
"""

from moai_adk.core.project.checker import SystemChecker, check_environment


class TestSystemCheckerConstants:
    """Test SystemChecker constants"""

    def test_has_required_tools(self):
        """Should define required tools"""
        assert hasattr(SystemChecker, "REQUIRED_TOOLS")
        assert isinstance(SystemChecker.REQUIRED_TOOLS, dict)
        assert "git" in SystemChecker.REQUIRED_TOOLS
        assert "python" in SystemChecker.REQUIRED_TOOLS

    def test_has_optional_tools(self):
        """Should define optional tools"""
        assert hasattr(SystemChecker, "OPTIONAL_TOOLS")
        assert isinstance(SystemChecker.OPTIONAL_TOOLS, dict)


class TestSystemCheckerCheckTool:
    """Test _check_tool method"""

    def test_check_tool_returns_true_for_available_tool(self):
        """Should return True when tool is available"""
        checker = SystemChecker()

        # Python should always be available (we're running in it!)
        result = checker._check_tool("python3 --version")
        assert result is True

    def test_check_tool_returns_false_for_unavailable_tool(self):
        """Should return False when tool is not available"""
        checker = SystemChecker()

        result = checker._check_tool("nonexistent_tool_xyz --version")
        assert result is False

    def test_check_tool_returns_false_for_empty_command(self):
        """Should return False for empty command"""
        checker = SystemChecker()

        result = checker._check_tool("")
        assert result is False

    def test_check_tool_handles_exceptions(self):
        """Should handle exceptions gracefully"""
        checker = SystemChecker()

        # Should not crash even with malformed command
        result = checker._check_tool(None)
        assert result is False


class TestSystemCheckerCheckAll:
    """Test check_all method"""

    def test_check_all_returns_dict(self):
        """Should return dictionary of tool statuses"""
        checker = SystemChecker()

        result = checker.check_all()

        assert isinstance(result, dict)
        assert len(result) > 0

    def test_check_all_includes_required_tools(self):
        """Should include all required tools in result"""
        checker = SystemChecker()

        result = checker.check_all()

        for tool in SystemChecker.REQUIRED_TOOLS:
            assert tool in result

    def test_check_all_includes_optional_tools(self):
        """Should include all optional tools in result"""
        checker = SystemChecker()

        result = checker.check_all()

        for tool in SystemChecker.OPTIONAL_TOOLS:
            assert tool in result

    def test_check_all_returns_boolean_values(self):
        """All values should be boolean"""
        checker = SystemChecker()

        result = checker.check_all()

        for value in result.values():
            assert isinstance(value, bool)


class TestCheckEnvironment:
    """Test check_environment function"""

    def test_check_environment_returns_dict(self):
        """Should return dictionary"""
        result = check_environment()

        assert isinstance(result, dict)
        assert len(result) > 0

    def test_check_environment_checks_python_version(self):
        """Should check Python version"""
        result = check_environment()

        assert "Python >= 3.13" in result
        # Should be True (we're running Python 3.13.1)
        assert result["Python >= 3.13"] is True

    def test_check_environment_checks_git(self):
        """Should check Git installation"""
        result = check_environment()

        assert "Git installed" in result
        # Value depends on system, just verify it's boolean
        assert isinstance(result["Git installed"], bool)

    def test_check_environment_checks_project_structure(self):
        """Should check for .moai directory"""
        result = check_environment()

        assert "Project structure (.moai/)" in result
        # Value depends on where test is run
        assert isinstance(result["Project structure (.moai/)"], bool)

    def test_check_environment_checks_config_file(self):
        """Should check for config.json"""
        result = check_environment()

        assert "Config file (.moai/config.json)" in result
        assert isinstance(result["Config file (.moai/config.json)"], bool)

    def test_check_environment_all_values_are_boolean(self):
        """All check results should be boolean"""
        result = check_environment()

        for key, value in result.items():
            assert isinstance(value, bool), f"{key} should have boolean value"
