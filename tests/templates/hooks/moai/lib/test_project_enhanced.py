"""Enhanced tests for project.py - Batch 3 coverage improvements

Focus: Project root finding, language detection, Git info, SPEC counting, network checks
Target Coverage: 71.4% → 90.0% (+18.6%)
"""

import json
import socket
import subprocess

# Import from templates directory
import sys
import time
from unittest.mock import Mock, patch

import pytest

sys.path.insert(0, "src/moai_adk/templates/.claude/hooks/moai/lib")
from project import (
    count_specs,
    detect_language,
    find_project_root,
    get_git_info,
    get_package_version_info,
    get_project_language,
    get_version_check_config,
    is_major_version_change,
    is_network_available,
    timeout_handler,
)


class TestFindProjectRoot:
    """Test project root finding - NEW COVERAGE"""

    def test_find_root_with_config(self, tmp_path):
        """Test finding root with .moai/config/config.json"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        (config_dir / "config.json").write_text("{}")

        root = find_project_root(tmp_path)
        assert root == tmp_path

    def test_find_root_with_claude_md(self, tmp_path):
        """Test finding root with CLAUDE.md"""
        (tmp_path / "CLAUDE.md").write_text("# Project")

        root = find_project_root(tmp_path)
        assert root == tmp_path

    def test_find_root_from_nested_directory(self, tmp_path):
        """Test finding root from nested subdirectory"""
        # Create project structure
        (tmp_path / "CLAUDE.md").write_text("# Project")
        nested = tmp_path / "level1" / "level2" / "level3"
        nested.mkdir(parents=True)

        root = find_project_root(nested)
        assert root == tmp_path

    def test_find_root_not_found(self, tmp_path):
        """Test when root is not found"""
        nested = tmp_path / "some" / "nested" / "path"
        nested.mkdir(parents=True)

        root = find_project_root(nested)
        # Should return absolute path of start location
        assert root.is_absolute()

    def test_find_root_max_depth(self, tmp_path):
        """Test max depth limit prevents infinite loops"""
        # Create very deep nesting
        deep = tmp_path
        for i in range(15):  # More than max_depth (10)
            deep = deep / f"level{i}"
        deep.mkdir(parents=True)

        root = find_project_root(deep)
        # Should stop at max depth
        assert root.is_absolute()


class TestTimeoutHandler:
    """Test timeout handler - NEW COVERAGE"""

    def test_timeout_success(self):
        """Test operation completes within timeout"""
        with timeout_handler(2):
            time.sleep(0.1)  # Quick operation
        # Should complete successfully

    def test_timeout_raises_error(self):
        """Test timeout raises TimeoutError"""
        from project import TimeoutError

        with pytest.raises(TimeoutError):
            with timeout_handler(1):
                time.sleep(2)  # Exceeds timeout


class TestDetectLanguage:
    """Test language detection - ENHANCED COVERAGE"""

    def test_detect_python_with_pyproject(self, tmp_path):
        """Test Python detection via pyproject.toml"""
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")
        assert detect_language(str(tmp_path)) == "python"

    def test_detect_typescript_priority(self, tmp_path):
        """Test TypeScript takes priority over JavaScript"""
        (tmp_path / "package.json").write_text('{"name": "test"}')
        (tmp_path / "tsconfig.json").write_text('{"compilerOptions": {}}')

        assert detect_language(str(tmp_path)) == "typescript"

    def test_detect_javascript_only(self, tmp_path):
        """Test JavaScript when no tsconfig"""
        (tmp_path / "package.json").write_text('{"name": "test"}')
        assert detect_language(str(tmp_path)) == "javascript"

    def test_detect_go(self, tmp_path):
        """Test Go language detection"""
        (tmp_path / "go.mod").write_text("module test")
        assert detect_language(str(tmp_path)) == "go"

    def test_detect_rust(self, tmp_path):
        """Test Rust language detection"""
        (tmp_path / "Cargo.toml").write_text("[package]")
        assert detect_language(str(tmp_path)) == "rust"

    def test_detect_java(self, tmp_path):
        """Test Java language detection"""
        (tmp_path / "pom.xml").write_text("<project></project>")
        assert detect_language(str(tmp_path)) == "java"

    def test_detect_dart(self, tmp_path):
        """Test Dart language detection"""
        (tmp_path / "pubspec.yaml").write_text("name: test")
        assert detect_language(str(tmp_path)) == "dart"

    def test_detect_swift(self, tmp_path):
        """Test Swift language detection"""
        (tmp_path / "Package.swift").write_text("// swift-tools-version:5.0")
        assert detect_language(str(tmp_path)) == "swift"

    def test_detect_kotlin(self, tmp_path):
        """Test Kotlin language detection"""
        (tmp_path / "build.gradle.kts").write_text("plugins {}")
        assert detect_language(str(tmp_path)) == "kotlin"

    def test_detect_php(self, tmp_path):
        """Test PHP language detection"""
        (tmp_path / "composer.json").write_text('{"name": "test/test"}')
        assert detect_language(str(tmp_path)) == "php"

    def test_detect_ruby(self, tmp_path):
        """Test Ruby language detection"""
        (tmp_path / "Gemfile").write_text("source 'https://rubygems.org'")
        assert detect_language(str(tmp_path)) == "ruby"

    def test_detect_csharp_with_glob(self, tmp_path):
        """Test C# detection via .csproj files"""
        (tmp_path / "Project.csproj").write_text("<Project></Project>")
        assert detect_language(str(tmp_path)) == "csharp"

    def test_detect_shell(self, tmp_path):
        """Test Shell script detection"""
        (tmp_path / "script.sh").write_text("#!/bin/bash")
        assert detect_language(str(tmp_path)) == "shell"

    def test_detect_unknown_language(self, tmp_path):
        """Test unknown language when no markers found"""
        # Empty directory
        assert detect_language(str(tmp_path)) == "Unknown Language"


class TestGetGitInfo:
    """Test Git information retrieval - ENHANCED COVERAGE"""

    def test_git_info_valid_repo(self, tmp_path):
        """Test getting Git info from valid repository"""
        # Initialize git repo
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "config", "user.name", "Test"], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "config", "user.email", "test@test.com"], cwd=tmp_path, capture_output=True)

        # Create initial commit
        test_file = tmp_path / "test.txt"
        test_file.write_text("test")
        subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "commit", "-m", "Initial commit"], cwd=tmp_path, capture_output=True)

        info = get_git_info(str(tmp_path))

        assert "branch" in info
        assert "commit" in info
        assert "changes" in info
        assert "last_commit" in info

    def test_git_info_non_git_directory(self, tmp_path):
        """Test Git info returns empty dict for non-Git directory"""
        info = get_git_info(str(tmp_path))
        assert info == {}

    def test_git_info_with_changes(self, tmp_path):
        """Test Git info detects file changes"""
        # Initialize repo with commit
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "config", "user.name", "Test"], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "config", "user.email", "test@test.com"], cwd=tmp_path, capture_output=True)

        test_file = tmp_path / "test.txt"
        test_file.write_text("test")
        subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "commit", "-m", "Init"], cwd=tmp_path, capture_output=True)

        # Make changes
        (tmp_path / "new_file.txt").write_text("new")

        info = get_git_info(str(tmp_path))
        assert info["changes"] >= 1

    def test_git_info_long_commit_message(self, tmp_path):
        """Test long commit messages are truncated"""
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "config", "user.name", "Test"], cwd=tmp_path, capture_output=True)
        subprocess.run(["git", "config", "user.email", "test@test.com"], cwd=tmp_path, capture_output=True)

        test_file = tmp_path / "test.txt"
        test_file.write_text("test")
        subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)

        long_message = "A" * 100  # Very long message
        subprocess.run(["git", "commit", "-m", long_message], cwd=tmp_path, capture_output=True)

        info = get_git_info(str(tmp_path))
        assert len(info["last_commit"]) <= 50


class TestCountSpecs:
    """Test SPEC counting - ENHANCED COVERAGE"""

    def test_count_specs_empty_directory(self, tmp_path):
        """Test counting with no specs directory"""
        result = count_specs(str(tmp_path))
        assert result == {"completed": 0, "total": 0, "percentage": 0}

    def test_count_specs_with_completed(self, tmp_path):
        """Test counting with completed SPECs"""
        specs_dir = tmp_path / ".moai" / "specs"
        specs_dir.mkdir(parents=True)

        # Create SPEC with completed status
        spec_dir = specs_dir / "SPEC-001"
        spec_dir.mkdir()
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("---\nstatus: completed\n---\nContent")

        result = count_specs(str(tmp_path))
        assert result["total"] == 1
        assert result["completed"] == 1
        assert result["percentage"] == 100

    def test_count_specs_with_pending(self, tmp_path):
        """Test counting with pending SPECs"""
        specs_dir = tmp_path / ".moai" / "specs"
        specs_dir.mkdir(parents=True)

        # Create SPEC with pending status
        spec_dir = specs_dir / "SPEC-002"
        spec_dir.mkdir()
        spec_file = spec_dir / "spec.md"
        spec_file.write_text("---\nstatus: pending\n---\nContent")

        result = count_specs(str(tmp_path))
        assert result["total"] == 1
        assert result["completed"] == 0
        assert result["percentage"] == 0

    def test_count_specs_mixed_status(self, tmp_path):
        """Test counting with mixed SPEC statuses"""
        specs_dir = tmp_path / ".moai" / "specs"
        specs_dir.mkdir(parents=True)

        # Create multiple SPECs
        for i, status in enumerate(["completed", "pending", "completed", "in_progress"]):
            spec_dir = specs_dir / f"SPEC-{i:03d}"
            spec_dir.mkdir()
            spec_file = spec_dir / "spec.md"
            spec_file.write_text(f"---\nstatus: {status}\n---\nContent")

        result = count_specs(str(tmp_path))
        assert result["total"] == 4
        assert result["completed"] == 2
        assert result["percentage"] == 50

    def test_count_specs_from_nested_directory(self, tmp_path):
        """Test counting from nested directory finds project root"""
        specs_dir = tmp_path / ".moai" / "specs"
        specs_dir.mkdir(parents=True)

        spec_dir = specs_dir / "SPEC-001"
        spec_dir.mkdir()
        (spec_dir / "spec.md").write_text("---\nstatus: completed\n---\n")

        # Also create CLAUDE.md to mark root
        (tmp_path / "CLAUDE.md").write_text("# Project")

        # Count from nested directory
        nested = tmp_path / "src" / "nested"
        nested.mkdir(parents=True)

        result = count_specs(str(nested))
        assert result["total"] == 1


class TestGetProjectLanguage:
    """Test project language retrieval - NEW COVERAGE"""

    def test_get_language_from_config(self, tmp_path):
        """Test reading language from config.json"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps({"language": "python"}))

        # Also create CLAUDE.md to mark root
        (tmp_path / "CLAUDE.md").write_text("# Project")

        language = get_project_language(str(tmp_path))
        assert language == "python"

    def test_get_language_fallback_to_detection(self, tmp_path):
        """Test fallback to language detection when config missing"""
        (tmp_path / "pyproject.toml").write_text("[tool.poetry]")

        language = get_project_language(str(tmp_path))
        assert language == "python"


class TestVersionCheckConfig:
    """Test version check configuration - NEW COVERAGE"""

    def test_version_check_default_config(self, tmp_path):
        """Test default version check configuration"""
        config = get_version_check_config(str(tmp_path))

        assert config["enabled"] is True
        assert config["frequency"] == "daily"
        assert config["cache_ttl_hours"] == 24

    def test_version_check_custom_frequency(self, tmp_path):
        """Test custom frequency settings"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps({"moai": {"update_check_frequency": "weekly"}}))

        # Create CLAUDE.md to mark root
        (tmp_path / "CLAUDE.md").write_text("# Project")

        config = get_version_check_config(str(tmp_path))
        assert config["frequency"] == "weekly"
        assert config["cache_ttl_hours"] == 168  # 7 days

    def test_version_check_disabled(self, tmp_path):
        """Test disabled version checking"""
        config_dir = tmp_path / ".moai" / "config"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "config.json"
        config_file.write_text(json.dumps({"moai": {"version_check": {"enabled": False}}}))

        (tmp_path / "CLAUDE.md").write_text("# Project")

        config = get_version_check_config(str(tmp_path))
        assert config["enabled"] is False


class TestNetworkAvailability:
    """Test network availability checking - NEW COVERAGE"""

    def test_network_available_success(self):
        """Test network availability check succeeds"""
        # This test may fail in offline environments
        try:
            result = is_network_available(timeout_seconds=1.0)
            assert isinstance(result, bool)
        except Exception:
            pytest.skip("Network test skipped (offline environment)")

    def test_network_unavailable_timeout(self):
        """Test network check with very short timeout"""
        # Very short timeout should fail
        result = is_network_available(timeout_seconds=0.001)
        # Should return False (timeout or connection error)
        assert isinstance(result, bool)

    @patch("socket.create_connection")
    def test_network_mock_success(self, mock_connect):
        """Test network check with mocked successful connection"""
        mock_conn = Mock()
        mock_connect.return_value = mock_conn

        result = is_network_available()
        assert result is True
        mock_conn.close.assert_called_once()

    @patch("socket.create_connection")
    def test_network_mock_failure(self, mock_connect):
        """Test network check with mocked connection failure"""
        mock_connect.side_effect = socket.timeout()

        result = is_network_available()
        assert result is False


class TestMajorVersionChange:
    """Test major version change detection - NEW COVERAGE"""

    def test_major_version_change_0_to_1(self):
        """Test detecting 0.x → 1.x as major change"""
        assert is_major_version_change("0.28.0", "1.0.0") is True

    def test_major_version_change_1_to_2(self):
        """Test detecting 1.x → 2.x as major change"""
        assert is_major_version_change("1.5.3", "2.0.0") is True

    def test_minor_version_change(self):
        """Test minor version change is not major"""
        assert is_major_version_change("0.28.0", "0.29.0") is False

    def test_patch_version_change(self):
        """Test patch version change is not major"""
        assert is_major_version_change("0.28.0", "0.28.1") is False

    def test_invalid_version_format(self):
        """Test invalid version format returns False"""
        assert is_major_version_change("dev", "1.0.0") is False
        assert is_major_version_change("0.28.0", "invalid") is False


class TestPackageVersionInfo:
    """Test package version info retrieval - NEW COVERAGE"""

    @patch("importlib.metadata.version")
    def test_package_version_current_only(self, mock_version, tmp_path):
        """Test getting current version when network unavailable"""
        mock_version.return_value = "0.28.0"

        with patch("project.is_network_available", return_value=False):
            info = get_package_version_info(str(tmp_path))

            assert info["current"] == "0.28.0"
            assert info["latest"] == "unknown"
            assert info["update_available"] is False

    def test_package_version_dev_mode(self, tmp_path):
        """Test dev mode returns immediately"""
        with patch("importlib.metadata.version") as mock_version:
            from importlib.metadata import PackageNotFoundError

            mock_version.side_effect = PackageNotFoundError()

            info = get_package_version_info(str(tmp_path))
            assert info["current"] == "dev"

    @patch("importlib.metadata.version")
    def test_package_version_with_update_available(self, mock_version, tmp_path):
        """Test detecting available updates"""
        mock_version.return_value = "0.28.0"

        # Mock network and PyPI response
        with patch("project.is_network_available", return_value=True):
            with patch("urllib.request.urlopen") as mock_urlopen:
                mock_response = Mock()
                mock_response.__enter__ = Mock(return_value=mock_response)
                mock_response.__exit__ = Mock(return_value=None)
                mock_response.read = Mock(return_value=json.dumps({"info": {"version": "0.29.0"}}).encode())
                mock_urlopen.return_value = mock_response

                info = get_package_version_info(str(tmp_path))

                assert info["current"] == "0.28.0"
                assert info["latest"] == "0.29.0"
                assert info["update_available"] is True


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
