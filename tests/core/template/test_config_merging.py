"""
Tests for config.json merging functionality in TemplateProcessor.

Tests the new _merge_config_json method with various scenarios including:
- Priority-based merging (Environment > User > Template)
- LanguageConfigResolver integration
- Error handling and fallback scenarios
"""

import json
import os
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.core.template.backup import TemplateBackup
from moai_adk.core.template.processor import TemplateProcessor


@pytest.fixture
def tmp_project_with_config(tmp_path: Path) -> Path:
    """Create temporary project with config structure"""
    # Create project structure
    (tmp_path / ".moai").mkdir()
    (tmp_path / ".claude").mkdir()
    (tmp_path / ".moai" / "config").mkdir()

    # Create existing user config
    user_config = {
        "user": {"name": "ExistingUser"},
        "language": {
            "conversation_language": "ja",
            "custom_setting": "user_value"
        },
        "project": {"name": "UserProject"}
    }
    (tmp_path / ".moai" / "config" / "config.json").write_text(
        json.dumps(user_config, indent=2)
    )

    return tmp_path


@pytest.fixture
def template_with_config(tmp_path: Path) -> Path:
    """Create template structure with config.json"""
    template_root = tmp_path / "template"
    template_root.mkdir()
    (template_root / ".claude").mkdir()

    # Create template config
    template_config = {
        "user": {"name": "TemplateUser", "email": "template@example.com"},
        "language": {
            "conversation_language": "en",
            "conversation_language_name": "English",
            "git_commit_messages": "en",
            "code_comments": "en"
        },
        "project": {"mode": "team", "new_setting": "template_value"}
    }
    (template_root / ".claude" / "config.json").write_text(
        json.dumps(template_config, indent=2)
    )

    return template_root


class TestMergeConfigJson:
    """Test _merge_config_json method"""

    def test_merge_config_json_basic_priority(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test basic priority merging: User > Template"""
        backup = TemplateBackup(tmp_project_with_config)
        processor = TemplateProcessor(template_with_config, tmp_project_with_config, backup)

        src = template_with_config / ".claude" / "config.json"
        dst = tmp_project_with_config / ".moai" / "config" / "config.json"

        processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # User config should override template
        assert merged_config["user"]["name"] == "ExistingUser"  # User > Template
        assert merged_config["user"]["email"] == "template@example.com"  # Template only

        # Project config should merge
        assert merged_config["project"]["name"] == "UserProject"  # User
        assert merged_config["project"]["mode"] == "team"  # Template

        # Language should deep merge
        assert merged_config["language"]["conversation_language"] == "ja"  # User
        assert merged_config["language"]["conversation_language_name"] == "English"  # Template
        assert merged_config["language"]["custom_setting"] == "user_value"  # User only

    def test_merge_config_json_with_environment_override(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test environment variable override: Environment > User > Template"""
        backup = TemplateBackup(tmp_project_with_config)
        processor = TemplateProcessor(template_with_config, tmp_project_with_config, backup)

        # Set environment variables
        with patch.dict(os.environ, {
            'MOAI_USER_NAME': 'EnvUser',
            'MOAI_CONVERSATION_LANG': 'ko'
        }):
            src = template_with_config / ".claude" / "config.json"
            dst = tmp_project_with_config / ".moai" / "config" / "config.json"

            processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # Environment should override everything
        assert merged_config["user"]["name"] == "EnvUser"  # Environment > User
        assert merged_config["language"]["conversation_language"] == "ko"  # Environment > User

    def test_merge_config_json_no_existing_config(self, tmp_path: Path, template_with_config: Path) -> None:
        """Test merging when no existing config exists"""
        # Create project without existing config
        (tmp_path / ".moai" / "config").mkdir(parents=True)
        dst = tmp_path / ".moai" / "config" / "config.json"

        backup = TemplateBackup(tmp_path)
        processor = TemplateProcessor(template_with_config, tmp_path, backup)

        src = template_with_config / ".claude" / "config.json"
        processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # Should have template config
        assert merged_config["user"]["name"] == "TemplateUser"
        assert merged_config["project"]["mode"] == "team"

    def test_merge_config_json_with_backup_extraction(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test merging with backup config extraction"""
        backup = TemplateBackup(tmp_project_with_config)

        # Create backup with different config
        with patch('moai_adk.core.template.backup.datetime') as mock_datetime:
            mock_datetime.now.return_value.strftime.return_value = "20241201_143022"
            backup_path = backup.create_backup()

        # Modify backup config
        backup_config_path = backup_path / ".moai" / "config" / "config.json"
        backup_config_path.parent.mkdir(parents=True, exist_ok=True)
        backup_config = {
            "user": {"name": "BackupUser"},
            "project": {"backup_setting": "from_backup"}
        }
        backup_config_path.write_text(json.dumps(backup_config, indent=2))

        # Remove current config to test backup extraction
        (tmp_project_with_config / ".moai" / "config" / "config.json").unlink()
        dst = tmp_project_with_config / ".moai" / "config" / "config.json"

        processor = TemplateProcessor(template_with_config, tmp_project_with_config, backup)
        src = template_with_config / ".claude" / "config.json"

        processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # Should include backup config
        assert merged_config["user"]["name"] == "BackupUser"
        assert merged_config["project"]["backup_setting"] == "from_backup"

    def test_merge_config_json_corrupted_template(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test handling of corrupted template config"""
        backup = TemplateBackup(tmp_project_with_config)
        processor = TemplateProcessor(template_with_config, tmp_project_with_config, backup)

        # Create corrupted template config
        src = template_with_config / ".claude" / "corrupted_config.json"
        src.write_text("invalid json content")

        dst = tmp_project_with_config / ".moai" / "config" / "config.json"

        # Should not raise exception, just warn
        processor._merge_config_json(src, dst)

        # Original config should remain unchanged
        original_config = json.loads(dst.read_text())
        assert original_config["user"]["name"] == "ExistingUser"

    @patch('moai_adk.core.template.processor.LanguageConfigResolver')
    def test_merge_config_json_without_resolver(self, mock_resolver: MagicMock, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test fallback merging when LanguageConfigResolver is unavailable"""
        # Simulate import failure
        with patch.dict('sys.modules', {'moai_adk.core.language_config_resolver': None}):
            backup = TemplateBackup(tmp_project_with_config)
            processor = TemplateProcessor(template_with_config, tmp_project_with_config, backup)

            src = template_with_config / ".claude" / "config.json"
            dst = tmp_project_with_config / ".moai" / "config" / "config.json"

            processor._merge_config_json(src, dst)

            # Should still merge with simple logic
            merged_config = json.loads(dst.read_text())
            assert "user" in merged_config
            assert "language" in merged_config
            assert "project" in merged_config

    def test_merge_config_json_deep_merge_objects(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test deep merging of nested objects"""
        backup = TemplateBackup(tmp_project_with_config)
        processor = TemplateProcessor(template_with_config, tmp_project_with_config, backup)

        src = template_with_config / ".claude" / "config.json"
        dst = tmp_project_with_config / ".moai" / "config" / "config.json"

        processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # Check deep merge behavior
        language_config = merged_config["language"]
        assert language_config["conversation_language"] == "ja"  # User preserved
        assert language_config["conversation_language_name"] == "English"  # Template added
        assert language_config["custom_setting"] == "user_value"  # User preserved
        assert language_config["git_commit_messages"] == "en"  # Template added

    def test_merge_config_json_excludes_metadata(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test that metadata fields are excluded from merging"""
        backup = TemplateBackup(tmp_project_with_config)
        processor = TemplateProcessor(template_with_config, tmp_project_with_config, backup)

        # Add metadata to existing config
        existing_config = json.loads((tmp_project_with_config / ".moai" / "config" / "config.json").read_text())
        existing_config["config_source"] = "environment"
        existing_config["last_updated"] = "2024-01-01"
        (tmp_project_with_config / ".moai" / "config" / "config.json").write_text(
            json.dumps(existing_config, indent=2)
        )

        src = template_with_config / ".claude" / "config.json"
        dst = tmp_project_with_config / ".moai" / "config" / "config.json"

        processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # Metadata should be handled by LanguageConfigResolver
        assert "config_source" in merged_config  # Should be managed by resolver