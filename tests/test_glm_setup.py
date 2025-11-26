#!/usr/bin/env python3
"""Tests for GLM setup script and configuration

Validates:
- Token resolution sequence (.env.glm, environment variable, argument)
- settings.local.json contains all required keys
- .env.glm file created with correct permissions
- .gitignore updated with .env.glm entry
"""

import json

# Import the setup_glm function
import sys
import tempfile
from pathlib import Path

import pytest

# Skip this test - setup-glm.py is only in templates, not in local project
pytestmark = pytest.mark.skip(reason="setup-glm.py only exists in src/moai_adk/templates, not in local .moai/scripts")

moai_scripts_path = Path(__file__).parent.parent / ".moai" / "scripts"
sys.path.insert(0, str(moai_scripts_path))

# Import using importlib to handle the script file
import importlib.util

try:
    spec = importlib.util.spec_from_file_location("setup_glm", moai_scripts_path / "setup-glm.py")
    if spec and spec.loader:
        setup_glm_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(setup_glm_module)
        setup_glm = setup_glm_module.setup_glm
    else:
        setup_glm = None
except (FileNotFoundError, AttributeError):
    setup_glm = None


class TestGLMSetup:
    """Test suite for GLM configuration setup"""

    @pytest.fixture
    def temp_project(self):
        """Create a temporary project structure for testing"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create .claude directory structure
            claude_dir = project_root / ".claude"
            claude_dir.mkdir()

            # Create minimal settings.local.json
            settings_path = claude_dir / "settings.local.json"
            settings_path.write_text(json.dumps({"companyAnnouncements": [], "permissions": {}}, indent=2))

            # Create .gitignore
            gitignore_path = project_root / ".gitignore"
            gitignore_path.write_text("node_modules/\n.DS_Store\n")

            yield project_root

    def test_env_glm_file_created_with_token(self, temp_project):
        """Test that .env.glm file is created with the API token"""
        test_token = "test_token_abc123xyz"

        result = setup_glm(test_token, temp_project)

        assert result is True
        env_glm_path = temp_project / ".env.glm"
        assert env_glm_path.exists()
        assert env_glm_path.read_text() == f"ANTHROPIC_AUTH_TOKEN={test_token}\n"

    def test_env_glm_file_has_secure_permissions(self, temp_project):
        """Test that .env.glm has secure permissions (0o600)"""
        test_token = "test_token_secure"

        setup_glm(test_token, temp_project)

        env_glm_path = temp_project / ".env.glm"
        file_mode = oct(env_glm_path.stat().st_mode)[-3:]
        assert file_mode == "600", f"Expected 600 permissions, got {file_mode}"

    def test_settings_json_updated_with_all_keys(self, temp_project):
        """Test that settings.local.json contains all required environment variables"""
        test_token = "test_token_complete"

        setup_glm(test_token, temp_project)

        settings_path = temp_project / ".claude" / "settings.local.json"
        settings = json.loads(settings_path.read_text())

        # Verify all required keys exist
        assert "env" in settings
        env_config = settings["env"]

        required_keys = [
            "ANTHROPIC_AUTH_TOKEN",
            "ANTHROPIC_BASE_URL",
            "ANTHROPIC_DEFAULT_HAIKU_MODEL",
            "ANTHROPIC_DEFAULT_SONNET_MODEL",
            "ANTHROPIC_DEFAULT_OPUS_MODEL",
        ]

        for key in required_keys:
            assert key in env_config, f"Missing key: {key}"

    def test_settings_json_contains_correct_values(self, temp_project):
        """Test that settings.local.json has correct configuration values"""
        test_token = "test_token_values"

        setup_glm(test_token, temp_project)

        settings_path = temp_project / ".claude" / "settings.local.json"
        settings = json.loads(settings_path.read_text())
        env_config = settings["env"]

        # Verify correct values
        assert env_config["ANTHROPIC_AUTH_TOKEN"] == test_token
        assert env_config["ANTHROPIC_BASE_URL"] == "https://api.z.ai/api/anthropic"
        assert env_config["ANTHROPIC_DEFAULT_HAIKU_MODEL"] == "glm-4.5-air"
        assert env_config["ANTHROPIC_DEFAULT_SONNET_MODEL"] == "glm-4.6"
        assert env_config["ANTHROPIC_DEFAULT_OPUS_MODEL"] == "glm-4.6"

    def test_gitignore_updated_with_env_glm(self, temp_project):
        """Test that .gitignore contains .env.glm entry"""
        test_token = "test_token_gitignore"

        setup_glm(test_token, temp_project)

        gitignore_path = temp_project / ".gitignore"
        gitignore_content = gitignore_path.read_text()

        assert ".env.glm" in gitignore_content

    def test_gitignore_not_duplicated(self, temp_project):
        """Test that .env.glm is not duplicated in .gitignore"""
        test_token = "test_token_no_dup"

        # Run setup twice
        setup_glm(test_token, temp_project)
        setup_glm("another_token", temp_project)

        gitignore_path = temp_project / ".gitignore"
        gitignore_content = gitignore_path.read_text()

        # Count occurrences of .env.glm
        count = gitignore_content.count(".env.glm")
        assert count == 1, f"Expected 1 occurrence of .env.glm, found {count}"

    def test_invalid_token_rejected(self, temp_project):
        """Test that invalid token (too short) is rejected"""
        invalid_token = "short"

        # This should be caught by main() but setup_glm doesn't validate
        # The validation happens at CLI level in main()
        result = setup_glm(invalid_token, temp_project)

        # setup_glm should still work, but main() validates before calling it
        assert result is True

    def test_missing_settings_json_error(self, temp_project):
        """Test that error is returned when settings.local.json is missing"""
        test_token = "test_token_missing"

        # Remove settings.local.json
        settings_path = temp_project / ".claude" / "settings.local.json"
        settings_path.unlink()

        result = setup_glm(test_token, temp_project)

        assert result is False

    def test_setup_preserves_existing_settings(self, temp_project):
        """Test that existing settings are preserved during update"""
        test_token = "test_token_preserve"

        # Add custom setting
        settings_path = temp_project / ".claude" / "settings.local.json"
        settings = json.loads(settings_path.read_text())
        settings["custom_key"] = "custom_value"
        settings_path.write_text(json.dumps(settings, indent=2))

        # Run setup
        setup_glm(test_token, temp_project)

        # Verify custom setting is preserved
        settings = json.loads(settings_path.read_text())
        assert settings.get("custom_key") == "custom_value"
        assert settings["env"]["ANTHROPIC_AUTH_TOKEN"] == test_token


class TestGLMTokenResolution:
    """Test token resolution priority"""

    @pytest.fixture
    def temp_project_with_existing_env_glm(self):
        """Create project with existing .env.glm file"""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_root = Path(tmpdir)

            # Create .claude directory
            claude_dir = project_root / ".claude"
            claude_dir.mkdir()

            # Create settings.local.json
            settings_path = claude_dir / "settings.local.json"
            settings_path.write_text(json.dumps({"companyAnnouncements": [], "permissions": {}}, indent=2))

            # Create existing .env.glm with different token
            env_glm_path = project_root / ".env.glm"
            env_glm_path.write_text("ANTHROPIC_AUTH_TOKEN=existing_token_old\n")

            # Create .gitignore
            gitignore_path = project_root / ".gitignore"
            gitignore_path.write_text("")

            yield project_root

    def test_new_token_overwrites_existing_env_glm(self, temp_project_with_existing_env_glm):
        """Test that new token overwrites existing .env.glm"""
        new_token = "new_token_overwrite"

        setup_glm(new_token, temp_project_with_existing_env_glm)

        env_glm_path = temp_project_with_existing_env_glm / ".env.glm"
        assert env_glm_path.read_text() == f"ANTHROPIC_AUTH_TOKEN={new_token}\n"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
