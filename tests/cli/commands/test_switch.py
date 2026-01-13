"""Tests for cli.commands.switch module."""

import json
import os
import tempfile
from pathlib import Path
from unittest.mock import patch

import click
import pytest

from moai_adk.cli.commands.switch import (
    GLM_ENV_KEYS,
    _get_credential_value,
    _get_paths,
    _has_glm_env,
    _substitute_env_vars,
    _switch_to_claude,
    _switch_to_glm,
    switch_to_claude,
    switch_to_glm,
    update_glm_key,
)


class TestGetCredentialValue:
    """Test _get_credential_value function."""

    def test_get_glm_key_from_env_file(self):
        """Test getting GLM API key from .env.glm file."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value="glm-from-env"):
            result = _get_credential_value("GLM_API_KEY")
            assert result == "glm-from-env"

    def test_get_glm_token_from_env_file(self):
        """Test getting GLM_API_TOKEN (alias) from .env.glm file."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value="glm-token"):
            result = _get_credential_value("GLM_API_TOKEN")
            assert result == "glm-token"

    def test_get_glm_key_from_credentials_fallback(self):
        """Test getting GLM key from credentials.yaml as fallback."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": "glm-from-creds"},
            ):
                result = _get_credential_value("GLM_API_KEY")
                assert result == "glm-from-creds"

    def test_get_glm_key_from_environment_fallback(self):
        """Test getting GLM key from environment variable as final fallback."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": None},
            ):
                with patch.dict(os.environ, {"GLM_API_KEY": "glm-from-env-var"}):
                    result = _get_credential_value("GLM_API_KEY")
                    assert result == "glm-from-env-var"

    def test_get_anthropic_key_from_credentials(self):
        """Test getting Anthropic key from credentials.yaml."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"anthropic_api_key": "sk-ant-creds"},
            ):
                result = _get_credential_value("ANTHROPIC_API_KEY")
                assert result == "sk-ant-creds"

    def test_get_anthropic_key_from_environment(self):
        """Test getting Anthropic key from environment variable."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"anthropic_api_key": None},
            ):
                with patch.dict(os.environ, {"ANTHROPIC_API_KEY": "sk-ant-env"}):
                    result = _get_credential_value("ANTHROPIC_API_KEY")
                    assert result == "sk-ant-env"

    def test_get_credential_not_found(self):
        """Test getting credential when not found anywhere."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": None, "anthropic_api_key": None},
            ):
                with patch.dict(os.environ, {}, clear=True):
                    result = _get_credential_value("UNKNOWN_KEY")
                    assert result is None


class TestSubstituteEnvVars:
    """Test _substitute_env_vars function."""

    def test_substitute_single_var(self):
        """Test substituting a single environment variable."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="value123"):
            result, missing = _substitute_env_vars("https://api.example.com/${VAR}")
            assert result == "https://api.example.com/value123"
            assert missing == []

    def test_substitute_multiple_vars(self):
        """Test substituting multiple environment variables."""
        cred_values = {"VAR1": "val1", "VAR2": "val2"}

        def mock_get_cred(var_name):
            return cred_values.get(var_name)

        with patch("moai_adk.cli.commands.switch._get_credential_value", side_effect=mock_get_cred):
            result, missing = _substitute_env_vars("${VAR1}/path/${VAR2}")
            assert result == "val1/path/val2"
            assert missing == []

    def test_substitute_with_missing_vars(self):
        """Test substituting with missing variables."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=None):
            result, missing = _substitute_env_vars("${MISSING_VAR}/path")
            assert result == "${MISSING_VAR}/path"
            assert missing == ["MISSING_VAR"]

    def test_substitute_mixed_available_missing(self):
        """Test substituting with mix of available and missing variables."""

        def mock_get_cred(var_name):
            return "value" if var_name == "AVAILABLE" else None

        with patch("moai_adk.cli.commands.switch._get_credential_value", side_effect=mock_get_cred):
            result, missing = _substitute_env_vars("${AVAILABLE}/${MISSING}")
            assert result == "value/${MISSING}"
            assert missing == ["MISSING"]

    def test_substitute_no_vars(self):
        """Test string with no variables to substitute."""
        with patch("moai_adk.cli.commands.switch._get_credential_value") as mock_get:
            result, missing = _substitute_env_vars("https://static.url.com/path")
            assert result == "https://static.url.com/path"
            assert missing == []
            mock_get.assert_not_called()

    def test_substitute_duplicate_missing_vars(self):
        """Test that duplicate missing vars are only reported once."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=None):
            result, missing = _substitute_env_vars("${VAR1}/${VAR2}/${VAR1}")
            assert result == "${VAR1}/${VAR2}/${VAR1}"
            # Missing vars should be collected, but deduplication happens later
            assert len(missing) == 3


class TestGetPaths:
    """Test _get_paths function."""

    def test_get_paths_success(self):
        """Test getting paths when project is initialized."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            claude_dir = project_path / ".claude"
            claude_dir.mkdir()

            with patch("moai_adk.cli.commands.switch.Path.cwd", return_value=project_path):
                claude, settings_local, glm_config = _get_paths()

                assert claude == claude_dir
                assert settings_local == claude_dir / "settings.local.json"
                assert glm_config == project_path / ".moai" / "llm-configs" / "glm.json"

    def test_get_paths_not_initialized(self):
        """Test getting paths when project is not initialized."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            # Don't create .claude directory

            with patch("moai_adk.cli.commands.switch.Path.cwd", return_value=project_path):
                with pytest.raises(click.Abort):
                    _get_paths()


class TestHasGlmEnv:
    """Test _has_glm_env function."""

    def test_has_glm_env_true(self):
        """Test _has_glm_env returns True when GLM env is configured."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_file = Path(tmpdir) / "settings.local.json"
            settings_data = {
                "env": {
                    "ANTHROPIC_BASE_URL": "https://api.glm.com",
                    "OTHER_VAR": "value",
                }
            }
            settings_file.write_text(json.dumps(settings_data))

            result = _has_glm_env(settings_file)
            assert result is True

    def test_has_glm_env_false(self):
        """Test _has_glm_env returns False when GLM env is not configured."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_file = Path(tmpdir) / "settings.local.json"
            settings_data = {
                "env": {
                    "OTHER_VAR": "value",
                }
            }
            settings_file.write_text(json.dumps(settings_data))

            result = _has_glm_env(settings_file)
            assert result is False

    def test_has_glm_env_no_file(self):
        """Test _has_glm_env returns False when file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_file = Path(tmpdir) / "settings.local.json"

            result = _has_glm_env(settings_file)
            assert result is False

    def test_has_glm_env_invalid_json(self):
        """Test _has_glm_env returns False for invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_file = Path(tmpdir) / "settings.local.json"
            settings_file.write_text("invalid json {")

            result = _has_glm_env(settings_file)
            assert result is False

    def test_has_glm_env_no_env_section(self):
        """Test _has_glm_env returns False when no env section."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_file = Path(tmpdir) / "settings.local.json"
            settings_data = {"otherKey": "value"}
            settings_file.write_text(json.dumps(settings_data))

            result = _has_glm_env(settings_file)
            assert result is False


class TestUpdateGlmKey:
    """Test update_glm_key function."""

    def test_update_glm_key_success(self):
        """Test updating GLM API key successfully."""
        with patch("moai_adk.core.credentials.save_glm_key_to_env"):
            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=Path("/.moai/.env.glm")):
                with patch.dict(os.environ, {}, clear=True):
                    update_glm_key("test-api-key")

    def test_update_glm_key_with_env_var_cleanup(self):
        """Test updating GLM key removes environment variable from shell config."""
        with patch("moai_adk.core.credentials.save_glm_key_to_env"):
            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=Path("/.moai/.env.glm")):
                with patch(
                    "moai_adk.core.credentials.remove_glm_key_from_shell_config",
                    return_value={"zshrc": True, "bashrc": False},
                ):
                    with patch.dict(os.environ, {"GLM_API_KEY": "old-key"}):
                        update_glm_key("new-api-key")

    def test_update_glm_key_env_var_not_in_config(self):
        """Test updating GLM key when env var not in shell config."""
        with patch("moai_adk.core.credentials.save_glm_key_to_env"):
            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=Path("/.moai/.env.glm")):
                with patch(
                    "moai_adk.core.credentials.remove_glm_key_from_shell_config",
                    return_value={"zshrc": False, "bashrc": False},
                ):
                    with patch.dict(os.environ, {"GLM_API_KEY": "old-key"}):
                        update_glm_key("new-api-key")


class TestSwitchToGlm:
    """Test switch_to_glm function."""

    def test_switch_to_glm_no_key(self):
        """Test switching to GLM when API key is not configured."""
        with patch("moai_adk.core.credentials.glm_env_exists", return_value=False):
            with pytest.raises(click.Abort):
                switch_to_glm()

    def test_switch_to_glm_success(self):
        """Test successful switch to GLM backend."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            claude_dir = project_path / ".claude"
            claude_dir.mkdir()
            settings_local = claude_dir / "settings.local.json"

            glm_config = project_path / ".moai" / "llm-configs" / "glm.json"
            glm_config.parent.mkdir(parents=True)
            glm_config.write_text('{"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}')

            with patch("moai_adk.core.credentials.glm_env_exists", return_value=True):
                with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value="test-key-12345"):
                    with patch("moai_adk.cli.commands.switch.Path.cwd", return_value=project_path):
                        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="cred-value"):
                            switch_to_glm()

                            # Verify settings.local.json was created/updated
                            assert settings_local.exists()
                            data = json.loads(settings_local.read_text())
                            assert "env" in data
                            assert "ANTHROPIC_BASE_URL" in data["env"]


class TestSwitchToClaude:
    """Test switch_to_claude function."""

    def test_switch_to_claude_success(self):
        """Test successful switch to Claude backend."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            claude_dir = project_path / ".claude"
            claude_dir.mkdir()
            settings_local = claude_dir / "settings.local.json"

            # Create settings with GLM env
            settings_data = {"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}
            settings_local.write_text(json.dumps(settings_data))

            with patch("moai_adk.cli.commands.switch.Path.cwd", return_value=project_path):
                switch_to_claude()

            # Verify GLM env was removed
            data = json.loads(settings_local.read_text())
            assert "ANTHROPIC_BASE_URL" not in data.get("env", {})

    def test_switch_to_claude_already_claude(self):
        """Test switching to Claude when already using Claude."""
        with tempfile.TemporaryDirectory() as tmpdir:
            project_path = Path(tmpdir)
            claude_dir = project_path / ".claude"
            claude_dir.mkdir()
            settings_local = claude_dir / "settings.local.json"

            # Create settings without GLM env
            settings_data = {"env": {"OTHER_VAR": "value"}}
            settings_local.write_text(json.dumps(settings_data))

            with patch("moai_adk.cli.commands.switch.Path.cwd", return_value=project_path):
                switch_to_claude()

            # File should still exist
            assert settings_local.exists()


class TestInternalSwitchToGlm:
    """Test _switch_to_glm internal function."""

    def test_switch_to_glm_already_glm(self):
        """Test _switch_to_glm when already using GLM."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            settings_data = {"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}
            settings_local.write_text(json.dumps(settings_data))

            glm_config = Path(tmpdir) / "glm.json"
            glm_config.write_text('{"env": {}}')

            _switch_to_glm(settings_local, glm_config)

            # Settings should not be modified
            data = json.loads(settings_local.read_text())
            assert data["env"]["ANTHROPIC_BASE_URL"] == "https://api.glm.com"

    def test_switch_to_glm_no_config_file(self):
        """Test _switch_to_glm when GLM config file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            glm_config = Path(tmpdir) / "nonexistent.json"

            with pytest.raises(click.Abort):
                _switch_to_glm(settings_local, glm_config)

    def test_switch_to_glm_missing_credentials(self):
        """Test _switch_to_glm when required credentials are missing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            glm_config = Path(tmpdir) / "glm.json"
            glm_config.write_text('{"env": {"ANTHROPIC_BASE_URL": "${MISSING_KEY}"}}')

            with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=None):
                with pytest.raises(click.Abort):
                    _switch_to_glm(settings_local, glm_config)

    def test_switch_to_glm_creates_new_settings(self):
        """Test _switch_to_glm creates new settings.local.json."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            glm_config = Path(tmpdir) / "glm.json"
            glm_config.write_text('{"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}')

            with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="cred-value"):
                _switch_to_glm(settings_local, glm_config)

            # Verify settings file was created
            assert settings_local.exists()
            data = json.loads(settings_local.read_text())
            assert "env" in data

    def test_switch_to_glm_merges_with_existing_settings(self):
        """Test _switch_to_glm merges with existing settings."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            existing_data = {"env": {"EXISTING_VAR": "keep-me"}}
            settings_local.write_text(json.dumps(existing_data))

            glm_config = Path(tmpdir) / "glm.json"
            glm_config.write_text('{"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}')

            with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="cred-value"):
                _switch_to_glm(settings_local, glm_config)

            # Verify existing settings were preserved
            data = json.loads(settings_local.read_text())
            assert data["env"]["EXISTING_VAR"] == "keep-me"
            assert "ANTHROPIC_BASE_URL" in data["env"]

    def test_switch_to_glm_substitutes_env_vars(self):
        """Test _switch_to_glm substitutes environment variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            glm_config = Path(tmpdir) / "glm.json"
            glm_config.write_text('{"env": {"API_KEY": "${CREDENTIAL_VAR}"}}')

            with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="substituted-value"):
                _switch_to_glm(settings_local, glm_config)

            # Verify substitution occurred
            data = json.loads(settings_local.read_text())
            assert data["env"]["API_KEY"] == "substituted-value"


class TestInternalSwitchToClaude:
    """Test _switch_to_claude internal function."""

    def test_switch_to_claude_already_claude(self):
        """Test _switch_to_claude when already using Claude."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            settings_data = {"env": {"OTHER_VAR": "value"}}
            settings_local.write_text(json.dumps(settings_data))

            _switch_to_claude(settings_local)

            # Settings should remain unchanged
            data = json.loads(settings_local.read_text())
            assert data["env"]["OTHER_VAR"] == "value"

    def test_switch_to_claude_removes_glm_env(self):
        """Test _switch_to_claude removes GLM environment variables."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            settings_data = {
                "env": {
                    "ANTHROPIC_BASE_URL": "https://api.glm.com",
                    "ANTHROPIC_AUTH_TOKEN": "token",
                    "OTHER_VAR": "keep",
                }
            }
            settings_local.write_text(json.dumps(settings_data))

            _switch_to_claude(settings_local)

            # Verify GLM env keys were removed
            data = json.loads(settings_local.read_text())
            assert "ANTHROPIC_BASE_URL" not in data.get("env", {})
            assert "ANTHROPIC_AUTH_TOKEN" not in data.get("env", {})
            assert data["env"]["OTHER_VAR"] == "keep"

    def test_switch_to_claude_removes_empty_env_section(self):
        """Test _switch_to_claude removes env section when empty."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            settings_data = {"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}
            settings_local.write_text(json.dumps(settings_data))

            _switch_to_claude(settings_local)

            # Verify env section was removed
            data = json.loads(settings_local.read_text())
            assert "env" not in data

    def test_switch_to_claude_invalid_json(self):
        """Test _switch_to_claude with invalid JSON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            settings_local.write_text("invalid json")

            # Should handle gracefully and not raise
            _switch_to_claude(settings_local)

    def test_glm_env_keys_constant(self):
        """Test GLM_ENV_KEYS contains expected keys."""
        assert "ANTHROPIC_AUTH_TOKEN" in GLM_ENV_KEYS
        assert "ANTHROPIC_BASE_URL" in GLM_ENV_KEYS
        assert "ANTHROPIC_DEFAULT_HAIKU_MODEL" in GLM_ENV_KEYS
        assert "ANTHROPIC_DEFAULT_SONNET_MODEL" in GLM_ENV_KEYS
        assert "ANTHROPIC_DEFAULT_OPUS_MODEL" in GLM_ENV_KEYS
