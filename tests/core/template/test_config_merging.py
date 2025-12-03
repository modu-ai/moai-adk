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
from unittest.mock import patch

import pytest

from moai_adk.core.template.processor import TemplateProcessor


# Environment variables that need to be cleared for test isolation
MOAI_ENV_VARS = [
    "MOAI_USER_NAME",
    "MOAI_CONVERSATION_LANG",
    "MOAI_AGENT_PROMPT_LANG",
    "MOAI_CONVERSATION_LANG_NAME",
    "MOAI_GIT_COMMIT_MESSAGES_LANG",
    "MOAI_CODE_COMMENTS_LANG",
    "MOAI_DOCUMENTATION_LANG",
    "MOAI_ERROR_MESSAGES_LANG",
]


@pytest.fixture(autouse=True)
def clear_moai_env_vars():
    """Clear MOAI environment variables for test isolation."""
    saved_env = {}
    for key in MOAI_ENV_VARS:
        if key in os.environ:
            saved_env[key] = os.environ.pop(key)
    yield
    # Restore saved environment variables
    for key, value in saved_env.items():
        os.environ[key] = value


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
        processor = TemplateProcessor(tmp_project_with_config)
        processor.template_root = template_with_config  # Override for testing

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
        processor = TemplateProcessor(tmp_project_with_config)
        processor.template_root = template_with_config  # Override for testing

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

        processor = TemplateProcessor(tmp_path)
        processor.template_root = template_with_config  # Override for testing

        src = template_with_config / ".claude" / "config.json"
        processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # Should have template config
        assert merged_config["user"]["name"] == "TemplateUser"
        assert merged_config["project"]["mode"] == "team"

    def test_merge_config_json_with_backup_extraction(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test merging when existing config is removed.

        Note: The current implementation reads backup config for reference but doesn't
        use it in the merge. When there's no existing config, only template config is used.
        """
        processor = TemplateProcessor(tmp_project_with_config)
        processor.template_root = template_with_config  # Override for testing

        # Create backup with different config (for reference, not used in merge)
        with patch('moai_adk.core.template.backup.datetime') as mock_datetime:
            mock_datetime.now.return_value.strftime.return_value = "20241201_143022"
            processor.backup.create_backup()

        # Remove current config
        (tmp_project_with_config / ".moai" / "config" / "config.json").unlink()
        dst = tmp_project_with_config / ".moai" / "config" / "config.json"
        src = template_with_config / ".claude" / "config.json"

        processor._merge_config_json(src, dst)

        merged_config = json.loads(dst.read_text())

        # With no existing config, should use template config
        # Note: backup config is read but not merged in current implementation
        assert merged_config["user"]["name"] == "TemplateUser"
        assert merged_config["project"]["mode"] == "team"

    def test_merge_config_json_corrupted_template(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test handling of corrupted template config"""
        processor = TemplateProcessor(tmp_project_with_config)
        processor.template_root = template_with_config  # Override for testing

        # Create corrupted template config
        src = template_with_config / ".claude" / "corrupted_config.json"
        src.write_text("invalid json content")

        dst = tmp_project_with_config / ".moai" / "config" / "config.json"

        # Should not raise exception, just warn
        processor._merge_config_json(src, dst)

        # Original config should remain unchanged
        original_config = json.loads(dst.read_text())
        assert original_config["user"]["name"] == "ExistingUser"

    def test_merge_config_json_without_resolver(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test fallback merging when LanguageConfigResolver is unavailable"""
        # Simulate import failure by patching the import in the function
        import builtins
        real_import = builtins.__import__

        def mock_import(name, *args, **kwargs):
            if name == "moai_adk.core.language_config_resolver":
                raise ImportError("Simulated import failure")
            return real_import(name, *args, **kwargs)

        processor = TemplateProcessor(tmp_project_with_config)
        processor.template_root = template_with_config  # Override for testing

        src = template_with_config / ".claude" / "config.json"
        dst = tmp_project_with_config / ".moai" / "config" / "config.json"

        with patch.object(builtins, "__import__", mock_import):
            processor._merge_config_json(src, dst)

        # Should still merge with simple logic
        merged_config = json.loads(dst.read_text())
        assert "user" in merged_config
        assert "language" in merged_config
        assert "project" in merged_config

    def test_merge_config_json_deep_merge_objects(self, tmp_project_with_config: Path, template_with_config: Path) -> None:
        """Test deep merging of nested objects"""
        processor = TemplateProcessor(tmp_project_with_config)
        processor.template_root = template_with_config  # Override for testing

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
        """Test that metadata fields like config_source are excluded from merging"""
        processor = TemplateProcessor(tmp_project_with_config)
        processor.template_root = template_with_config  # Override for testing

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

        # config_source is explicitly excluded from merging (see _merge_config_json line 1066)
        # Only non-metadata fields should be preserved
        assert "config_source" not in merged_config  # Excluded from merge
        assert "last_updated" in merged_config  # Other metadata is preserved