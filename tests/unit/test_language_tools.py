"""Unit tests for language tool chain mapping

Tests for LANGUAGE_TOOLS dictionary and language-specific tool detection.
"""

from moai_adk.core.project.checker import SystemChecker


class TestLanguageToolsMapping:
    """Test LANGUAGE_TOOLS constant"""

    def test_has_language_tools_constant(self):
        """Should define LANGUAGE_TOOLS dictionary"""
        assert hasattr(SystemChecker, "LANGUAGE_TOOLS")
        assert isinstance(SystemChecker.LANGUAGE_TOOLS, dict)

    def test_language_tools_has_python(self):
        """Should define Python tools"""
        tools = SystemChecker.LANGUAGE_TOOLS
        assert "python" in tools
        assert "required" in tools["python"]
        assert "recommended" in tools["python"]
        assert "optional" in tools["python"]

    def test_language_tools_has_typescript(self):
        """Should define TypeScript tools"""
        tools = SystemChecker.LANGUAGE_TOOLS
        assert "typescript" in tools
        assert "node" in tools["typescript"]["required"]
        assert "npm" in tools["typescript"]["required"]

    def test_language_tools_has_20_languages(self):
        """Should support 20 programming languages"""
        tools = SystemChecker.LANGUAGE_TOOLS
        # 20개 언어 지원 확인
        expected_languages = [
            "python",
            "typescript",
            "javascript",
            "java",
            "go",
            "rust",
            "dart",
            "swift",
            "kotlin",
            "csharp",
            "php",
            "ruby",
            "elixir",
            "scala",
            "clojure",
            "haskell",
            "c",
            "cpp",
            "lua",
            "ocaml",
        ]
        assert len(tools) >= 20, f"Expected at least 20 languages, got {len(tools)}"
        for lang in expected_languages:
            assert lang in tools, f"Missing language: {lang}"

    def test_python_required_tools(self):
        """Python should have python3 and pip as required"""
        tools = SystemChecker.LANGUAGE_TOOLS["python"]["required"]
        assert "python3" in tools
        assert "pip" in tools

    def test_python_recommended_tools(self):
        """Python should have pytest, mypy, ruff as recommended"""
        tools = SystemChecker.LANGUAGE_TOOLS["python"]["recommended"]
        assert "pytest" in tools
        assert "mypy" in tools
        assert "ruff" in tools

    def test_typescript_required_tools(self):
        """TypeScript should have node and npm as required"""
        tools = SystemChecker.LANGUAGE_TOOLS["typescript"]["required"]
        assert "node" in tools
        assert "npm" in tools

    def test_typescript_recommended_tools(self):
        """TypeScript should have vitest and biome as recommended"""
        tools = SystemChecker.LANGUAGE_TOOLS["typescript"]["recommended"]
        assert "vitest" in tools
        assert "biome" in tools

    def test_ruby_required_tools(self):
        """Ruby should have ruby and gem as required"""
        tools = SystemChecker.LANGUAGE_TOOLS["ruby"]["required"]
        assert "ruby" in tools
        assert "gem" in tools

    def test_ruby_recommended_tools(self):
        """Ruby should have bundler and rspec as recommended"""
        tools = SystemChecker.LANGUAGE_TOOLS["ruby"]["recommended"]
        assert "bundler" in tools
        assert "rspec" in tools

    def test_ruby_optional_tools(self):
        """Ruby should have rubocop as optional"""
        tools = SystemChecker.LANGUAGE_TOOLS["ruby"]["optional"]
        assert "rubocop" in tools

    def test_all_languages_have_required_field(self):
        """All languages should have 'required' field"""
        tools = SystemChecker.LANGUAGE_TOOLS
        for lang, config in tools.items():
            assert "required" in config, f"{lang} missing 'required' field"
            assert isinstance(config["required"], list), f"{lang} 'required' must be list"

    def test_all_languages_have_recommended_field(self):
        """All languages should have 'recommended' field"""
        tools = SystemChecker.LANGUAGE_TOOLS
        for lang, config in tools.items():
            assert "recommended" in config, f"{lang} missing 'recommended' field"
            assert isinstance(config["recommended"], list), f"{lang} 'recommended' must be list"


class TestCheckLanguageTools:
    """Test check_language_tools method"""

    def test_check_language_tools_method_exists(self):
        """Should have check_language_tools method"""
        checker = SystemChecker()
        assert hasattr(checker, "check_language_tools")
        assert callable(checker.check_language_tools)

    def test_check_language_tools_returns_dict(self):
        """Should return dictionary of tool statuses"""
        checker = SystemChecker()
        result = checker.check_language_tools("python")
        assert isinstance(result, dict)

    def test_check_language_tools_python(self):
        """Should check Python tools correctly"""
        checker = SystemChecker()
        result = checker.check_language_tools("python")

        # Required tools should be checked
        assert "python3" in result
        assert "pip" in result

        # Recommended tools should be checked
        assert "pytest" in result or "mypy" in result or "ruff" in result

    def test_check_language_tools_unknown_language(self):
        """Should return empty dict for unknown language"""
        checker = SystemChecker()
        result = checker.check_language_tools("unknown_lang_xyz")
        assert isinstance(result, dict)
        # Should return empty or raise informative error
        assert len(result) == 0 or "error" in result

    def test_check_language_tools_none_language(self):
        """Should handle None language gracefully"""
        checker = SystemChecker()
        result = checker.check_language_tools(None)
        assert isinstance(result, dict)
        assert len(result) == 0


class TestGetToolVersion:
    """Test get_tool_version method"""

    def test_get_tool_version_method_exists(self):
        """Should have get_tool_version method"""
        checker = SystemChecker()
        assert hasattr(checker, "get_tool_version")
        assert callable(checker.get_tool_version)

    def test_get_tool_version_python(self):
        """Should get Python version"""
        checker = SystemChecker()
        version = checker.get_tool_version("python3")
        # Python should be available (we're running in it)
        assert version is not None
        assert isinstance(version, str)
        assert len(version) > 0

    def test_get_tool_version_unavailable_tool(self):
        """Should return None for unavailable tool"""
        checker = SystemChecker()
        version = checker.get_tool_version("nonexistent_tool_xyz")
        assert version is None

    def test_get_tool_version_empty_string(self):
        """Should handle empty string"""
        checker = SystemChecker()
        version = checker.get_tool_version("")
        assert version is None
