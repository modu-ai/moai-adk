#!/usr/bin/env python3
"""
Integration tests for multilingual linting and formatting

Tests cover:
- Complete workflow with multilingual projects
- Hook orchestration
- File detection and routing
- Summary generation
"""

import pytest
import json
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock

from language_detector import LanguageDetector
from post_tool__multilingual_linting import MultilingualLintingHook
from post_tool__multilingual_formatting import MultilingualFormattingHook


class TestMultilingualLintingHook:
    """Test multilingual linting hook integration"""

    def test_hook_initialization(self):
        """Test hook initialization"""
        hook = MultilingualLintingHook()

        assert hook.detector is not None
        assert hook.linter_registry is not None
        assert len(hook.file_to_language_map) > 0

    def test_file_language_mapping(self):
        """Test file to language mapping"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create language markers
            (project_root / "pyproject.toml").touch()
            (project_root / "package.json").write_text(json.dumps({"name": "test"}))
            (project_root / "go.mod").touch()
            (project_root / "Cargo.toml").touch()
            (project_root / "tsconfig.json").touch()

            hook = MultilingualLintingHook(project_root)

            assert hook.get_language_for_file(Path("test.py")) == "python"
            assert hook.get_language_for_file(Path("test.js")) == "javascript"
            assert hook.get_language_for_file(Path("test.ts")) == "typescript"
            assert hook.get_language_for_file(Path("test.go")) == "go"
            assert hook.get_language_for_file(Path("test.rs")) == "rust"

    def test_unknown_file_extension(self):
        """Test handling of unknown file extensions"""
        hook = MultilingualLintingHook()

        result = hook.get_language_for_file(Path("test.xyz"))

        assert result is None

    def test_should_lint_file_validation(self):
        """Test file filtering for linting"""
        hook = MultilingualLintingHook()

        # Should lint regular source files
        assert hook.should_lint_file(Path("src/main.py")) is True

        # Should skip hidden files
        assert hook.should_lint_file(Path(".hidden.py")) is False

        # Should skip files in hidden directories
        assert hook.should_lint_file(Path(".git/config.py")) is False

        # Should skip node_modules
        assert hook.should_lint_file(Path("node_modules/package/index.js")) is False

        # Should skip __pycache__
        assert hook.should_lint_file(Path("__pycache__/module.cpython-39.pyc")) is False

    def test_lint_file_nonexistent(self):
        """Test linting nonexistent file"""
        hook = MultilingualLintingHook()

        result = hook.lint_file(Path("/nonexistent/test.py"))

        assert result is True  # Non-blocking

    def test_multilingual_project_summary(self):
        """Test summary generation for multilingual project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create multilingual project structure
            (project_root / "pyproject.toml").touch()
            (project_root / "src").mkdir()
            (project_root / "src" / "main.py").write_text("print('hello')")

            (project_root / "package.json").write_text(json.dumps({"name": "test"}))
            (project_root / "src" / "index.js").write_text("console.log('hello');")

            (project_root / "go.mod").touch()
            (project_root / "src" / "main.go").write_text("package main")

            hook = MultilingualLintingHook(project_root)

            # Create file paths to lint
            files_to_lint = [
                project_root / "src" / "main.py",
                project_root / "src" / "index.js",
                project_root / "src" / "main.go"
            ]

            summary = hook.lint_files(files_to_lint)

            assert summary["status"] == "completed"
            assert summary["total_files"] == 3
            assert "python" in summary["languages_detected"] or "javascript" in summary["languages_detected"]

    def test_summary_message_generation(self):
        """Test summary message generation"""
        hook = MultilingualLintingHook()

        summary = {
            "status": "completed",
            "total_files": 5,
            "files_checked": 4,
            "files_with_issues": 1,
            "files_by_language": {
                "python": {"count": 2, "passed": 1, "failed": 1},
                "javascript": {"count": 2, "passed": 2, "failed": 0}
            },
            "languages_detected": ["python", "javascript"]
        }

        message = hook.get_summary_message(summary)

        assert "Multilingual Linting Summary" in message
        assert "4/5" in message or "4" in message
        assert "python" in message.lower()
        assert "javascript" in message.lower()


class TestMultilingualFormattingHook:
    """Test multilingual formatting hook integration"""

    def test_hook_initialization(self):
        """Test hook initialization"""
        hook = MultilingualFormattingHook()

        assert hook.detector is not None
        assert hook.formatter_registry is not None
        assert len(hook.file_to_language_map) > 0

    def test_file_language_mapping(self):
        """Test file to language mapping"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create language markers
            (project_root / "pyproject.toml").touch()
            (project_root / "package.json").write_text(json.dumps({"name": "test"}))
            (project_root / "go.mod").touch()
            (project_root / "Cargo.toml").touch()
            (project_root / "tsconfig.json").touch()

            hook = MultilingualFormattingHook(project_root)

            assert hook.get_language_for_file(Path("test.py")) == "python"
            assert hook.get_language_for_file(Path("test.js")) == "javascript"
            assert hook.get_language_for_file(Path("test.ts")) == "typescript"
            assert hook.get_language_for_file(Path("test.go")) == "go"
            assert hook.get_language_for_file(Path("test.rs")) == "rust"

    def test_unknown_file_extension(self):
        """Test handling of unknown file extensions"""
        hook = MultilingualFormattingHook()

        result = hook.get_language_for_file(Path("test.xyz"))

        assert result is None

    def test_should_format_file_validation(self):
        """Test file filtering for formatting"""
        hook = MultilingualFormattingHook()

        # Should format regular source files
        assert hook.should_format_file(Path("src/main.py")) is True

        # Should skip hidden files
        assert hook.should_format_file(Path(".hidden.py")) is False

        # Should skip files in hidden directories
        assert hook.should_format_file(Path(".git/config.py")) is False

        # Should skip node_modules
        assert hook.should_format_file(Path("node_modules/package/index.js")) is False

        # Should skip minified files
        assert hook.should_format_file(Path("dist/bundle.min.js")) is False

        # Should skip bundled files
        assert hook.should_format_file(Path("dist/app.bundle.js")) is False

    def test_format_file_nonexistent(self):
        """Test formatting nonexistent file"""
        hook = MultilingualFormattingHook()

        result = hook.format_file(Path("/nonexistent/test.py"))

        assert result is True  # Non-blocking

    def test_multilingual_project_summary(self):
        """Test summary generation for multilingual project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create multilingual project structure
            (project_root / "pyproject.toml").touch()
            (project_root / "src").mkdir()
            (project_root / "src" / "main.py").write_text("print(  'hello'  )")

            (project_root / "package.json").write_text(json.dumps({"name": "test"}))
            (project_root / "src" / "index.js").write_text("console.log(  'hello'  );")

            (project_root / "go.mod").touch()
            (project_root / "src" / "main.go").write_text("package main")

            hook = MultilingualFormattingHook(project_root)

            # Create file paths to format
            files_to_format = [
                project_root / "src" / "main.py",
                project_root / "src" / "index.js",
                project_root / "src" / "main.go"
            ]

            summary = hook.format_files(files_to_format)

            assert summary["status"] == "completed"
            assert summary["total_files"] == 3

    def test_summary_message_generation(self):
        """Test summary message generation"""
        hook = MultilingualFormattingHook()

        summary = {
            "status": "completed",
            "total_files": 5,
            "files_formatted": 5,
            "files_by_language": {
                "python": {"count": 2, "formatted": 2},
                "javascript": {"count": 3, "formatted": 3}
            },
            "languages_detected": ["python", "javascript"]
        }

        message = hook.get_summary_message(summary)

        assert "Code Formatting Summary" in message
        assert "5/5" in message or "5" in message
        assert "python" in message.lower()
        assert "javascript" in message.lower()


class TestPythonProject:
    """Integration tests for Python projects"""

    def test_python_only_project(self):
        """Test linting Python-only project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Python project
            (project_root / "pyproject.toml").write_text(
                "[project]\nname = 'test'\nversion = '0.1.0'"
            )
            (project_root / "src").mkdir()
            (project_root / "src" / "main.py").write_text("print('hello')")

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "python" in languages
            assert detector.detect_primary_language() == "python"


class TestJavaScriptProject:
    """Integration tests for JavaScript projects"""

    def test_javascript_only_project(self):
        """Test linting JavaScript-only project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create JavaScript project
            package_json = {
                "name": "test-project",
                "version": "1.0.0",
                "main": "index.js"
            }
            (project_root / "package.json").write_text(json.dumps(package_json))
            (project_root / "index.js").write_text("console.log('hello');")

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "javascript" in languages or "typescript" in languages


class TestTypeScriptProject:
    """Integration tests for TypeScript projects"""

    def test_typescript_project(self):
        """Test linting TypeScript project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create TypeScript project
            (project_root / "tsconfig.json").write_text('{"compilerOptions": {}}')
            package_json = {
                "name": "test-project",
                "version": "1.0.0",
                "devDependencies": {"typescript": "^5.0.0"}
            }
            (project_root / "package.json").write_text(json.dumps(package_json))
            (project_root / "main.ts").write_text("const x: string = 'hello';")

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "typescript" in languages
            assert detector.detect_primary_language() == "typescript"


class TestGoProject:
    """Integration tests for Go projects"""

    def test_go_project(self):
        """Test linting Go project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Go project
            (project_root / "go.mod").write_text("module example.com/hello")
            (project_root / "main.go").write_text("package main\n\nfunc main() {}")

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "go" in languages


class TestRustProject:
    """Integration tests for Rust projects"""

    def test_rust_project(self):
        """Test linting Rust project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create Rust project
            (project_root / "Cargo.toml").write_text(
                "[package]\nname = 'hello'\nversion = '0.1.0'"
            )
            (project_root / "src").mkdir()
            (project_root / "src" / "main.rs").write_text("fn main() {}")

            detector = LanguageDetector(project_root)
            languages = detector.detect_languages()

            assert "rust" in languages


class TestEdgeCases:
    """Test edge cases and error handling"""

    def test_empty_project(self):
        """Test handling of empty project"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            hook = MultilingualLintingHook(project_root)
            summary = hook.lint_files([])

            assert summary["status"] == "skipped"

    def test_no_matching_files(self):
        """Test when no files match language"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)
            (project_root / "test.txt").write_text("not code")

            hook = MultilingualLintingHook(project_root)
            summary = hook.lint_files([project_root / "test.txt"])

            assert summary["status"] == "completed"
            assert summary["files_checked"] == 0

    def test_mixed_file_extensions(self):
        """Test project with many different file types"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create language markers for all languages
            (project_root / "pyproject.toml").touch()
            (project_root / "package.json").write_text(json.dumps({"name": "test"}))
            (project_root / "tsconfig.json").touch()
            (project_root / "go.mod").touch()
            (project_root / "Cargo.toml").touch()
            (project_root / "pom.xml").touch()
            (project_root / "Gemfile").touch()
            (project_root / "composer.json").touch()

            # Create files with many extensions
            files = [
                (project_root / "test.py", "python"),
                (project_root / "test.js", "javascript"),
                (project_root / "test.ts", "typescript"),
                (project_root / "test.go", "go"),
                (project_root / "test.rs", "rust"),
                (project_root / "test.java", "java"),
                (project_root / "test.rb", "ruby"),
                (project_root / "test.php", "php"),
                (project_root / "test.txt", None),
                (project_root / "test.md", None),
            ]

            for file_path, _ in files:
                file_path.write_text("test content")

            hook = MultilingualLintingHook(project_root)

            for file_path, expected_lang in files:
                detected_lang = hook.get_language_for_file(file_path)
                assert detected_lang == expected_lang


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
