"""Extended tests for moai_adk.core.version_sync module.

Comprehensive test coverage for VersionSynchronizer, VersionInfo, and VersionSource enums
with full method coverage and edge case handling.
"""

import json
from datetime import datetime
from pathlib import Path
from unittest.mock import Mock, patch

import pytest


class TestVersionSourceEnum:
    """Test VersionSource enum."""

    def test_version_source_enum_import(self):
        """Test that VersionSource enum can be imported."""
        from moai_adk.core.version_sync import VersionSource

        assert VersionSource is not None

    def test_version_source_enum_values(self):
        """Test VersionSource enum has expected values."""
        from moai_adk.core.version_sync import VersionSource

        assert hasattr(VersionSource, "PYPROJECT_TOML")
        assert hasattr(VersionSource, "CONFIG_JSON")
        assert hasattr(VersionSource, "PACKAGE_METADATA")
        assert hasattr(VersionSource, "FALLBACK")


class TestVersionInfoDataclass:
    """Test VersionInfo dataclass."""

    def test_version_info_import(self):
        """Test that VersionInfo can be imported."""
        from moai_adk.core.version_sync import VersionInfo, VersionSource

        assert VersionInfo is not None

    def test_version_info_creation_valid(self):
        """Test creating VersionInfo with valid version."""
        from moai_adk.core.version_sync import VersionInfo, VersionSource

        info = VersionInfo(
            version="1.0.0",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="version = '1.0.0'",
        )

        assert info.version == "1.0.0"
        assert info.source == VersionSource.PYPROJECT_TOML
        assert info.is_valid is True

    def test_version_info_creation_invalid_version(self):
        """Test creating VersionInfo with invalid version."""
        from moai_adk.core.version_sync import VersionInfo, VersionSource

        info = VersionInfo(
            version="invalid",
            source=VersionSource.CONFIG_JSON,
            file_path=Path("/test/config.json"),
            raw_content='{"version": "invalid"}',
        )

        assert info.version == "invalid"
        assert info.is_valid is False

    def test_version_info_timestamp(self):
        """Test VersionInfo timestamp is set automatically."""
        from moai_adk.core.version_sync import VersionInfo, VersionSource

        info = VersionInfo(
            version="1.0.0",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="",
        )

        assert info.timestamp is not None
        assert isinstance(info.timestamp, datetime)

    def test_version_info_semantic_versioning_valid(self):
        """Test VersionInfo validates semantic versioning."""
        from moai_adk.core.version_sync import VersionInfo, VersionSource

        valid_versions = ["1.0.0", "1.2.3", "0.0.1", "1.0.0-alpha"]

        for version in valid_versions:
            info = VersionInfo(
                version=version,
                source=VersionSource.PYPROJECT_TOML,
                file_path=Path("/test/pyproject.toml"),
                raw_content="",
            )
            assert info.is_valid is True, f"Version {version} should be valid"


class TestVersionSynchronizerInit:
    """Test VersionSynchronizer initialization."""

    def test_synchronizer_import(self):
        """Test that VersionSynchronizer can be imported."""
        from moai_adk.core.version_sync import VersionSynchronizer

        assert VersionSynchronizer is not None

    def test_synchronizer_init_default_working_dir(self):
        """Test VersionSynchronizer initialization with default working directory."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer()

        assert sync.working_dir is not None
        assert isinstance(sync.working_dir, Path)

    def test_synchronizer_init_custom_working_dir(self):
        """Test VersionSynchronizer initialization with custom working directory."""
        from moai_adk.core.version_sync import VersionSynchronizer

        custom_dir = Path("/custom/project")
        sync = VersionSynchronizer(custom_dir)

        assert sync.working_dir == custom_dir

    def test_synchronizer_init_version_files(self):
        """Test that version files are properly configured."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        assert sync.pyproject_path is not None
        assert sync.config_path is not None
        assert sync.init_path is not None
        assert len(sync.version_files) > 0

    def test_synchronizer_init_cache_directories(self):
        """Test that cache directories are properly configured."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        assert len(sync.cache_directories) > 0
        assert any("cache" in str(cache_dir) for cache_dir in sync.cache_directories)


class TestCheckConsistency:
    """Test check_consistency method."""

    def test_check_consistency_returns_tuple(self):
        """Test that check_consistency returns a tuple."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch.object(sync, "_extract_version", return_value=None):
            is_consistent, version_infos = sync.check_consistency()

        assert isinstance(is_consistent, bool)
        assert isinstance(version_infos, list)

    def test_check_consistency_consistent_versions(self):
        """Test check_consistency with consistent versions."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionInfo, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        mock_info = VersionInfo(
            version="1.0.0",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="",
        )

        with patch.object(sync, "_extract_version", return_value=mock_info):
            is_consistent, version_infos = sync.check_consistency()

        assert is_consistent is True

    def test_check_consistency_inconsistent_versions(self):
        """Test check_consistency with inconsistent versions."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionInfo, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        info1 = VersionInfo(
            version="1.0.0",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="",
        )
        info2 = VersionInfo(
            version="2.0.0",
            source=VersionSource.CONFIG_JSON,
            file_path=Path("/test/config.json"),
            raw_content="",
        )

        with patch.object(sync, "_extract_version", side_effect=[info1, info2]):
            is_consistent, version_infos = sync.check_consistency()

        assert is_consistent is False

    def test_check_consistency_no_versions_found(self):
        """Test check_consistency when no versions are found."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch.object(sync, "_extract_version", return_value=None):
            is_consistent, version_infos = sync.check_consistency()

        assert is_consistent is False
        assert len(version_infos) == 0


class TestGetMasterVersion:
    """Test get_master_version method."""

    def test_get_master_version_returns_version_info(self):
        """Test that get_master_version returns VersionInfo or None."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionInfo, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        mock_info = VersionInfo(
            version="1.0.0",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="",
        )

        with patch.object(sync, "_extract_version", return_value=mock_info):
            result = sync.get_master_version()

        assert result is not None
        assert result.version == "1.0.0"

    def test_get_master_version_invalid_version(self):
        """Test get_master_version with invalid version."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionInfo, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        mock_info = VersionInfo(
            version="invalid",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="",
        )

        with patch.object(sync, "_extract_version", return_value=mock_info):
            result = sync.get_master_version()

        assert result is None

    def test_get_master_version_not_found(self):
        """Test get_master_version when file not found."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch.object(sync, "_extract_version", return_value=None):
            result = sync.get_master_version()

        assert result is None


class TestSynchronizeAll:
    """Test synchronize_all method."""

    def test_synchronize_all_returns_bool(self):
        """Test that synchronize_all returns a boolean."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionInfo, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        mock_info = VersionInfo(
            version="1.0.0",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="",
        )

        with patch.object(sync, "get_master_version", return_value=mock_info):
            with patch.object(sync, "_synchronize_file", return_value=True):
                with patch.object(sync, "_clear_caches"):
                    result = sync.synchronize_all()

        assert isinstance(result, bool)

    def test_synchronize_all_with_target_version(self):
        """Test synchronize_all with explicit target version."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch.object(sync, "_synchronize_file", return_value=True):
            with patch.object(sync, "_clear_caches"):
                result = sync.synchronize_all(target_version="2.0.0")

        assert isinstance(result, bool)

    def test_synchronize_all_dry_run(self):
        """Test synchronize_all with dry_run enabled."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionInfo, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        mock_info = VersionInfo(
            version="1.0.0",
            source=VersionSource.PYPROJECT_TOML,
            file_path=Path("/test/pyproject.toml"),
            raw_content="",
        )

        with patch.object(sync, "get_master_version", return_value=mock_info):
            with patch.object(sync, "_synchronize_file", return_value=True):
                result = sync.synchronize_all(dry_run=True)

        assert isinstance(result, bool)

    def test_synchronize_all_no_master_version(self):
        """Test synchronize_all when no master version found."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch.object(sync, "get_master_version", return_value=None):
            result = sync.synchronize_all()

        assert result is False


class TestExtractVersion:
    """Test _extract_version method."""

    def test_extract_version_from_pyproject(self):
        """Test extracting version from pyproject.toml."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        # Create a mock file path that exists
        mock_path = Mock(spec=Path)
        mock_path.exists.return_value = True

        with patch("builtins.open", create=True) as mock_open:
            mock_file = Mock()
            mock_file.read.return_value = 'version = "1.0.0"'
            mock_open.return_value.__enter__.return_value = mock_file

            with patch.object(sync, "pyproject_path", mock_path):
                result = sync._extract_version(mock_path, VersionSource.PYPROJECT_TOML)

        assert result is not None
        assert result.version == "1.0.0"

    def test_extract_version_file_not_found(self):
        """Test _extract_version with nonexistent file."""
        from moai_adk.core.version_sync import VersionSynchronizer, VersionSource

        sync = VersionSynchronizer(Path("/test/project"))

        mock_path = Mock(spec=Path)
        mock_path.exists.return_value = False

        result = sync._extract_version(mock_path, VersionSource.PYPROJECT_TOML)

        assert result is None


class TestExtractFromPyproject:
    """Test _extract_from_pyproject method."""

    def test_extract_from_pyproject_valid(self):
        """Test extracting version from valid pyproject.toml content."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        content = 'version = "1.0.0"'
        result = sync._extract_from_pyproject(content)

        assert result == "1.0.0"

    def test_extract_from_pyproject_single_quotes(self):
        """Test extracting version with single quotes."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        content = "version = '1.2.3'"
        result = sync._extract_from_pyproject(content)

        assert result == "1.2.3"

    def test_extract_from_pyproject_not_found(self):
        """Test extracting version when not present."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        content = "name = 'test'"
        result = sync._extract_from_pyproject(content)

        assert result is None


class TestExtractFromConfig:
    """Test _extract_from_config method."""

    def test_extract_from_config_valid(self):
        """Test extracting version from valid config.json."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        content = json.dumps({"moai": {"version": "1.0.0"}})
        result = sync._extract_from_config(content)

        assert result == "1.0.0"

    def test_extract_from_config_invalid_json(self):
        """Test extracting version from invalid JSON."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        content = "invalid json {{{"
        result = sync._extract_from_config(content)

        assert result is None

    def test_extract_from_config_no_version(self):
        """Test extracting version when not present in config."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        content = json.dumps({"moai": {"name": "test"}})
        result = sync._extract_from_config(content)

        assert result is None


class TestValidateVersionFormat:
    """Test validate_version_format method."""

    def test_validate_version_format_valid(self):
        """Test validating valid version formats."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        valid_versions = ["1.0.0", "1.2.3", "0.0.1", "10.20.30", "1.0.0-alpha"]

        for version in valid_versions:
            is_valid, normalized = sync.validate_version_format(version)
            assert is_valid is True, f"Version {version} should be valid"

    def test_validate_version_format_with_v_prefix(self):
        """Test validating version with 'v' prefix."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        is_valid, normalized = sync.validate_version_format("v1.0.0")

        assert is_valid is True
        assert normalized == "1.0.0"

    def test_validate_version_format_invalid(self):
        """Test validating invalid version formats."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        invalid_versions = ["1.0", "1", "invalid", ""]

        for version in invalid_versions:
            is_valid, normalized = sync.validate_version_format(version)
            assert is_valid is False, f"Version {version} should be invalid"


class TestGetVersionReport:
    """Test get_version_report method."""

    def test_get_version_report_returns_dict(self):
        """Test that get_version_report returns a dictionary."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch.object(sync, "check_consistency", return_value=(True, [])):
            result = sync.get_version_report()

        assert isinstance(result, dict)
        assert "timestamp" in result
        assert "working_directory" in result
        assert "is_consistent" in result
        assert "version_count" in result
        assert "versions" in result
        assert "issues" in result
        assert "recommendations" in result

    def test_get_version_report_with_issues(self):
        """Test get_version_report identifies issues."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch.object(sync, "check_consistency", return_value=(False, [])):
            result = sync.get_version_report()

        assert result["is_consistent"] is False
        assert len(result["issues"]) > 0
        assert len(result["recommendations"]) > 0


class TestClearCaches:
    """Test _clear_caches method."""

    def test_clear_caches_success(self):
        """Test _clear_caches successfully clears caches."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        with patch("shutil.rmtree"):
            with patch.object(sync, "_clear_caches", wraps=sync._clear_caches):
                sync._clear_caches()

    def test_clear_caches_handles_missing_directories(self):
        """Test _clear_caches handles missing cache directories gracefully."""
        from moai_adk.core.version_sync import VersionSynchronizer

        sync = VersionSynchronizer(Path("/test/project"))

        # Mock cache directories to not exist
        for cache_dir in sync.cache_directories:
            cache_dir = Mock(spec=Path)
            cache_dir.exists.return_value = False

        # Should not raise
        sync._clear_caches()


class TestConvenienceFunctions:
    """Test module-level convenience functions."""

    def test_check_project_versions(self):
        """Test check_project_versions function."""
        from moai_adk.core.version_sync import check_project_versions

        with patch("moai_adk.core.version_sync.VersionSynchronizer") as mock_class:
            mock_instance = Mock()
            mock_instance.get_version_report.return_value = {
                "timestamp": datetime.now().isoformat(),
                "working_directory": "/test",
                "is_consistent": True,
                "version_count": 1,
                "versions": [],
                "issues": [],
                "recommendations": [],
            }
            mock_class.return_value = mock_instance

            result = check_project_versions()

            assert isinstance(result, dict)
            assert "is_consistent" in result

    def test_synchronize_project_versions(self):
        """Test synchronize_project_versions function."""
        from moai_adk.core.version_sync import synchronize_project_versions

        with patch("moai_adk.core.version_sync.VersionSynchronizer") as mock_class:
            mock_instance = Mock()
            mock_instance.synchronize_all.return_value = True
            mock_class.return_value = mock_instance

            result = synchronize_project_versions()

            assert isinstance(result, bool)
