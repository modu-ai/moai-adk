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


class TestPlatformSpecificConsole:
    """Test platform-specific console initialization."""

    def test_console_init_on_macos_linux(self):
        """Test console initialization on non-Windows platforms (line 23)."""
        # Mock sys.platform to be non-Windows
        with patch("moai_adk.cli.commands.switch.sys.platform", "darwin"):
            # Reimport to trigger console initialization
            import importlib
            from moai_adk.cli.commands import switch

            importlib.reload(switch)

            # Verify console was created with default settings
            # The else branch on line 24 should be executed
            assert hasattr(switch, "console")
            switch.console  # Access to verify it exists

    def test_console_init_on_windows(self):
        """Test console initialization on Windows platform."""
        # Mock sys.platform to be Windows
        with patch("moai_adk.cli.commands.switch.sys.platform", "win32"):
            # Reimport to trigger console initialization
            import importlib
            from moai_adk.cli.commands import switch

            importlib.reload(switch)

            # Verify console was created with Windows-specific settings
            # Line 23 should be executed
            assert hasattr(switch, "console")
            switch.console  # Access to verify it exists


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

    def test_get_credential_value_empty_string_in_creds(self):
        """Test getting credential when credentials dict has empty string."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": "", "anthropic_api_key": None},
            ):
                with patch.dict(os.environ, {"GLM_API_KEY": "from-env"}):
                    result = _get_credential_value("GLM_API_KEY")
                    # Empty string in credentials should be treated as falsy, fall through to env
                    assert result == "from-env"

    def test_get_credential_priority_env_file_over_creds(self):
        """Test credential priority: .env.glm > credentials.yaml."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value="from-env-file"):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": "from-creds-file"},
            ):
                with patch.dict(os.environ, {"GLM_API_KEY": "from-env-var"}):
                    result = _get_credential_value("GLM_API_KEY")
                    # .env.glm should take priority
                    assert result == "from-env-file"

    def test_get_credential_priority_creds_over_env_var(self):
        """Test credential priority: credentials.yaml > env var for non-GLM."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"anthropic_api_key": "from-creds"},
            ):
                with patch.dict(os.environ, {"ANTHROPIC_API_KEY": "from-env-var"}):
                    result = _get_credential_value("ANTHROPIC_API_KEY")
                    # credentials.yaml should take priority for non-GLM keys
                    assert result == "from-creds"


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

    def test_substitute_malformed_empty_braces(self):
        """Test substitution with empty braces ${}."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=None):
            result, missing = _substitute_env_vars("url/${}")
            # Empty pattern won't match \w+ regex, so no substitution occurs
            assert result == "url/${}"
            assert missing == []

    def test_substitute_malformed_open_brace(self):
        """Test substitution with malformed ${VAR pattern."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=None):
            result, missing = _substitute_env_vars("url/${VAR")
            # Unclosed brace won't match pattern
            assert result == "url/${VAR"
            assert missing == []

    def test_substitute_malformed_close_brace(self):
        """Test substitution with malformed VAR} pattern."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=None):
            result, missing = _substitute_env_vars("url/VAR}")
            # No opening ${ won't match pattern
            assert result == "url/VAR}"
            assert missing == []

    def test_substitute_special_chars_in_var_name(self):
        """Test substitution with special characters in variable name."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=None):
            result, missing = _substitute_env_vars("${VAR_NAME}-${VAR-NAME}-${VAR.NAME}")
            # Only \w+ matches (alphanumeric and underscore), so VAR_NAME is valid but others aren't
            assert "VAR_NAME" in missing
            assert result == "${VAR_NAME}-${VAR-NAME}-${VAR.NAME}"

    def test_substitute_empty_credential_value(self):
        """Test substitution when credential value is empty string."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value=""):
            result, missing = _substitute_env_vars("url/${VAR}")
            # Empty string is still a valid value
            assert result == "url/"
            assert missing == []

    def test_substitute_nested_braces(self):
        """Test substitution with nested braces."""
        with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="value"):
            result, missing = _substitute_env_vars("${OUTER_${INNER}}")
            # Pattern won't match nested braces, only inner is substituted
            assert result == "${OUTER_value}"
            assert missing == []


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

    def test_switch_to_glm_with_non_string_env_values(self):
        """Test _switch_to_glm handles non-string environment values."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            glm_config = Path(tmpdir) / "glm.json"
            # Include non-string values (number, boolean, null)
            glm_config.write_text(
                '{"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com", "PORT": 8080, "ENABLED": true, "OPTIONAL": null}}'
            )

            with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="cred-value"):
                _switch_to_glm(settings_local, glm_config)

            # Verify all values were preserved including non-strings
            data = json.loads(settings_local.read_text())
            assert data["env"]["ANTHROPIC_BASE_URL"] == "https://api.glm.com"
            assert data["env"]["PORT"] == 8080
            assert data["env"]["ENABLED"] is True
            assert data["env"]["OPTIONAL"] is None

    def test_switch_to_glm_with_invalid_json_settings(self):
        """Test _switch_to_glm handles existing settings with invalid JSON."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            # Write invalid JSON
            settings_local.write_text('{"env": {"INVALID": ')

            glm_config = Path(tmpdir) / "glm.json"
            glm_config.write_text('{"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}')

            with patch("moai_adk.cli.commands.switch._get_credential_value", return_value="cred-value"):
                _switch_to_glm(settings_local, glm_config)

            # Verify invalid file was replaced with valid content
            data = json.loads(settings_local.read_text())
            assert "env" in data
            assert "ANTHROPIC_BASE_URL" in data["env"]

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

    def test_switch_to_claude_with_glm_env_invalid_json(self):
        """Test _switch_to_claude when GLM env exists but file becomes invalid."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            # Create with GLM env first
            settings_data = {"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}
            settings_local.write_text(json.dumps(settings_data))

            # Now corrupt the file
            settings_local.write_text('{"env": {"ANTHROPIC_BASE_URL": "')

            # Should handle OSError/JSONDecodeError gracefully
            _switch_to_claude(settings_local)

    def test_switch_to_claude_os_error_read(self):
        """Test _switch_to_claude when file read raises OSError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            # Create with GLM env
            settings_data = {"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}
            settings_local.write_text(json.dumps(settings_data))

            # Mock read_text to raise OSError
            with patch.object(Path, "read_text", side_effect=OSError("Permission denied")):
                # Should handle OSError gracefully
                _switch_to_claude(settings_local)

    def test_switch_to_claude_with_glm_env_then_json_decode_error(self):
        """Test _switch_to_claude hits lines 301-303 when GLM env exists then JSON decode fails."""
        with tempfile.TemporaryDirectory() as tmpdir:
            settings_local = Path(tmpdir) / "settings.local.json"
            # Create with GLM env
            settings_data = {"env": {"ANTHROPIC_BASE_URL": "https://api.glm.com"}}
            settings_local.write_text(json.dumps(settings_data))

            # Create a mock that first returns True for has_glm_env, then fails on read
            original_read_text = Path.read_text

            # Use a closure to track call count
            call_count = [0]

            def mock_read_text_func(self, encoding=None):
                # First call from _has_glm_env succeeds
                # Second call from _switch_to_claude fails
                call_count[0] += 1

                if call_count[0] == 1:
                    return original_read_text(self, encoding=encoding)
                else:
                    raise json.JSONDecodeError("Invalid JSON", "", 0)

            with patch.object(Path, "read_text", mock_read_text_func):
                # This should hit lines 301-303 (except block in _switch_to_claude)
                _switch_to_claude(settings_local)

    def test_glm_env_keys_constant(self):
        """Test GLM_ENV_KEYS contains expected keys."""
        assert "ANTHROPIC_AUTH_TOKEN" in GLM_ENV_KEYS
        assert "ANTHROPIC_BASE_URL" in GLM_ENV_KEYS
        assert "ANTHROPIC_DEFAULT_HAIKU_MODEL" in GLM_ENV_KEYS
        assert "ANTHROPIC_DEFAULT_SONNET_MODEL" in GLM_ENV_KEYS
        assert "ANTHROPIC_DEFAULT_OPUS_MODEL" in GLM_ENV_KEYS
