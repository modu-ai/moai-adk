"""Additional coverage tests for project initializer.

Tests for lines not covered by existing tests.
"""

import stat
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.project.initializer import ProjectInitializer


class TestCreateUserSettingsAnnouncementsExceptions:
    """Test exception handling in _create_user_settings announcements."""

    def test_create_user_settings_handles_translate_exception(self, tmp_path):
        """Should handle exception when translate_announcements fails."""
        initializer = ProjectInitializer(tmp_path)

        # Create utils directory to trigger announcement loading
        utils_dir = tmp_path / ".claude" / "hooks" / "moai" / "shared" / "utils"
        utils_dir.mkdir(parents=True)

        # Create announcement_translator.py that raises exception
        translator_file = utils_dir / "announcement_translator.py"
        translator_file.write_text("def get_language_from_config(p): raise RuntimeError('test')")

        result = initializer._create_user_settings()

        # Should complete despite exception
        assert result is not None

    def test_create_user_settings_handles_import_error(self, tmp_path):
        """Should handle exception when importing announcement module."""
        initializer = ProjectInitializer(tmp_path)

        # Create utils directory to trigger announcement loading
        utils_dir = tmp_path / ".claude" / "hooks" / "moai" / "shared" / "utils"
        utils_dir.mkdir(parents=True)

        # Create invalid announcement_translator.py that will cause import error
        translator_file = utils_dir / "announcement_translator.py"
        translator_file.write_text("import nonexistent_module")

        result = initializer._create_user_settings()

        # Should complete despite import error
        assert result is not None


class TestInitializeLanguageConfigFallback:
    """Test language config default fallback."""

    def test_initialize_with_unknown_locale(self, tmp_path):
        """Should use default language config when locale not in language_names."""
        initializer = ProjectInitializer(tmp_path)

        # Use a locale that doesn't exist in language_names
        result = initializer.initialize(locale="unknown_locale_XX")

        # Should complete with fallback language config
        assert result is not None
        assert result.success is True


class TestInitializeScriptPermissions:
    """Test shell script permission setting."""

    def test_initialize_sets_script_permissions(self, tmp_path):
        """Should set execute permissions on shell scripts."""
        initializer = ProjectInitializer(tmp_path)

        # Create scripts directory with a test shell script
        scripts_dir = tmp_path / ".moai" / "scripts"
        scripts_dir.mkdir(parents=True)
        test_script = scripts_dir / "test.sh"
        test_script.write_text("#!/bin/bash\necho test")

        # Set file to read-only for owner (no execute permission)
        test_script.chmod(0o644)

        result = initializer.initialize()

        # Check that execute permissions were added
        current_mode = test_script.stat().st_mode
        # At least owner execute should be set
        assert current_mode & stat.S_IXUSR != 0

    def test_initialize_handles_permission_error(self, tmp_path):
        """Should silently ignore permission errors when setting script permissions."""
        initializer = ProjectInitializer(tmp_path)

        # Create scripts directory with a test file
        scripts_dir = tmp_path / ".moai" / "scripts"
        scripts_dir.mkdir(parents=True)
        test_script = scripts_dir / "test.sh"
        test_script.write_text("#!/bin/bash\necho test")

        # Mock chmod to raise exception
        original_chmod = Path.chmod

        def mock_chmod_raises(self, mode):
            if "test.sh" in str(self):
                raise OSError("Permission denied")
            return original_chmod(self, mode)

        with patch.object(Path, "chmod", mock_chmod_raises):
            # Should not raise exception during initialization
            result = initializer.initialize()
            assert result is not None
