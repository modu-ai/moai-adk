"""Tests for core.credentials module."""

import os
import stat
import tempfile
from pathlib import Path
from unittest.mock import patch

from moai_adk.core.credentials import (
    get_anthropic_api_key,
    get_credentials_path,
    get_env_glm_path,
    get_glm_api_key,
    glm_env_exists,
    load_credentials,
    load_glm_key_from_env,
    remove_glm_key_from_shell_config,
    save_credentials,
    save_glm_key_to_env,
)


class TestPathFunctions:
    """Test path utility functions."""

    def test_get_credentials_path(self):
        """Test getting credentials file path."""
        path = get_credentials_path()
        assert path == Path.home() / ".moai" / "credentials.yaml"

    def test_get_env_glm_path(self):
        """Test getting GLM env file path."""
        path = get_env_glm_path()
        assert path == Path.home() / ".moai" / ".env.glm"


class TestGlmEnvFile:
    """Test GLM environment file operations."""

    def test_glm_env_exists_true(self):
        """Test glm_env_exists returns True when file exists."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"
            env_path.write_text("GLM_API_KEY=test")

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                assert glm_env_exists() is True

    def test_glm_env_exists_false(self):
        """Test glm_env_exists returns False when file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                assert glm_env_exists() is False

    def test_load_glm_key_from_env_success(self):
        """Test loading GLM key from env file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"
            env_path.write_text('GLM_API_KEY="test-key-123"')

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                key = load_glm_key_from_env()
                assert key == "test-key-123"

    def test_load_glm_key_with_single_quotes(self):
        """Test loading GLM key with single quotes."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"
            env_path.write_text("GLM_API_KEY='test-key-456'")

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                key = load_glm_key_from_env()
                assert key == "test-key-456"

    def test_load_glm_key_without_quotes(self):
        """Test loading GLM key without quotes."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"
            env_path.write_text("GLM_API_KEY=test-key-789")

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                key = load_glm_key_from_env()
                assert key == "test-key-789"

    def test_load_glm_key_with_comments(self):
        """Test loading GLM key ignores comments."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"
            env_path.write_text('# This is a comment\nGLM_API_KEY="test-key"\n# Another comment')

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                key = load_glm_key_from_env()
                assert key == "test-key"

    def test_load_glm_key_empty_value(self):
        """Test loading GLM key with empty value returns None."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"
            env_path.write_text('GLM_API_KEY=""')

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                key = load_glm_key_from_env()
                assert key is None

    def test_load_glm_key_file_not_found(self):
        """Test loading GLM key when file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                key = load_glm_key_from_env()
                assert key is None

    def test_load_glm_key_os_error(self):
        """Test loading GLM key with OSError."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"
            env_path.write_text("GLM_API_KEY=test-key")

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                # Mock read_text to raise OSError
                with patch.object(Path, "read_text", side_effect=OSError("Permission denied")):
                    key = load_glm_key_from_env()
                    assert key is None

    def test_save_glm_key_to_env(self):
        """Test saving GLM key to env file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            env_path = Path(tmpdir) / ".env.glm"

            with patch("moai_adk.core.credentials.get_env_glm_path", return_value=env_path):
                save_glm_key_to_env("test-api-key")

                # Check file was created
                assert env_path.exists()

                # Check content
                content = env_path.read_text()
                assert "GLM_API_KEY=" in content
                assert "test-api-key" in content

                # Check file permissions (owner read/write only)
                file_stat = os.stat(env_path)
                mode = file_stat.st_mode
                assert mode & stat.S_IRUSR  # Owner read
                assert mode & stat.S_IWUSR  # Owner write
                assert not (mode & stat.S_IRGRP)  # No group read
                assert not (mode & stat.S_IROTH)  # No other read


class TestCredentialsYaml:
    """Test credentials YAML file operations."""

    def test_load_credentials_file_not_found(self):
        """Test loading credentials when file doesn't exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                creds = load_credentials()
                assert creds == {"anthropic_api_key": None, "glm_api_key": None}

    def test_load_credentials_success(self):
        """Test loading credentials from YAML file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"
            creds_path.write_text("anthropic_api_key: sk-ant-123\nglm_api_key: glm-key-456\n")

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                creds = load_credentials()
                assert creds["anthropic_api_key"] == "sk-ant-123"
                assert creds["glm_api_key"] == "glm-key-456"

    def test_load_credentials_partial(self):
        """Test loading credentials with only one key."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"
            creds_path.write_text("anthropic_api_key: sk-ant-789\n")

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                creds = load_credentials()
                assert creds["anthropic_api_key"] == "sk-ant-789"
                assert creds["glm_api_key"] is None

    def test_load_credentials_invalid_yaml(self):
        """Test loading credentials with invalid YAML."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"
            creds_path.write_text("invalid: yaml: content: [")

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                creds = load_credentials()
                assert creds == {"anthropic_api_key": None, "glm_api_key": None}

    def test_load_credentials_empty_file(self):
        """Test loading credentials from empty file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"
            creds_path.write_text("")

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                creds = load_credentials()
                assert creds == {"anthropic_api_key": None, "glm_api_key": None}

    def test_save_credentials_new_file(self):
        """Test saving credentials to new file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                save_credentials(anthropic_api_key="sk-ant-new", glm_api_key="glm-new")

                # Check file was created
                assert creds_path.exists()

                # Check file permissions
                file_stat = os.stat(creds_path)
                mode = file_stat.st_mode
                assert mode & stat.S_IRUSR
                assert mode & stat.S_IWUSR

    def test_save_credentials_merge_mode(self):
        """Test saving credentials in merge mode."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"
            creds_path.write_text("anthropic_api_key: sk-ant-old\n")

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                # Save only GLM key, should keep Anthropic key
                save_credentials(glm_api_key="glm-new", merge=True)

                # Verify both keys exist
                creds = load_credentials()
                assert creds["anthropic_api_key"] == "sk-ant-old"
                assert creds["glm_api_key"] == "glm-new"

    def test_save_credentials_overwrite_mode(self):
        """Test saving credentials in overwrite mode."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"
            creds_path.write_text("anthropic_api_key: sk-ant-old\nglm_api_key: glm-old\n")

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                # Save only Anthropic key in overwrite mode
                save_credentials(anthropic_api_key="sk-ant-new", merge=False)

                # Verify only Anthropic key exists
                creds = load_credentials()
                assert creds["anthropic_api_key"] == "sk-ant-new"
                assert creds["glm_api_key"] is None

    def test_save_credentials_none_values(self):
        """Test saving credentials with None values doesn't remove existing keys."""
        with tempfile.TemporaryDirectory() as tmpdir:
            creds_path = Path(tmpdir) / "credentials.yaml"
            creds_path.write_text("anthropic_api_key: sk-ant-old\n")

            with patch("moai_adk.core.credentials.get_credentials_path", return_value=creds_path):
                # Save with None value should keep existing key
                save_credentials(anthropic_api_key=None, glm_api_key="glm-new")

                creds = load_credentials()
                assert creds["anthropic_api_key"] == "sk-ant-old"
                assert creds["glm_api_key"] == "glm-new"


class TestGetApiKeyFunctions:
    """Test API key retrieval functions."""

    def test_get_glm_api_key_from_env_file(self):
        """Test getting GLM API key from .env.glm file."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value="glm-from-env"):
            key = get_glm_api_key()
            assert key == "glm-from-env"

    def test_get_glm_api_key_from_credentials(self):
        """Test getting GLM API key from credentials.yaml."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": "glm-from-creds"},
            ):
                key = get_glm_api_key()
                assert key == "glm-from-creds"

    def test_get_glm_api_key_from_environment(self):
        """Test getting GLM API key from environment variable."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": None},
            ):
                with patch.dict(os.environ, {"GLM_API_KEY": "glm-from-env-var"}):
                    key = get_glm_api_key()
                    assert key == "glm-from-env-var"

    def test_get_glm_api_key_not_found(self):
        """Test getting GLM API key when not found anywhere."""
        with patch("moai_adk.core.credentials.load_glm_key_from_env", return_value=None):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"glm_api_key": None},
            ):
                with patch.dict(os.environ, {}, clear=True):
                    key = get_glm_api_key()
                    assert key is None

    def test_get_anthropic_api_key_from_environment(self):
        """Test getting Anthropic API key from environment variable."""
        with patch.dict(os.environ, {"ANTHROPIC_API_KEY": "sk-ant-env"}):
            key = get_anthropic_api_key()
            assert key == "sk-ant-env"

    def test_get_anthropic_api_key_from_credentials(self):
        """Test getting Anthropic API key from credentials file."""
        with patch.dict(os.environ, {}, clear=True):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"anthropic_api_key": "sk-ant-creds"},
            ):
                key = get_anthropic_api_key()
                assert key == "sk-ant-creds"

    def test_get_anthropic_api_key_not_found(self):
        """Test getting Anthropic API key when not found."""
        with patch.dict(os.environ, {}, clear=True):
            with patch(
                "moai_adk.core.credentials.load_credentials",
                return_value={"anthropic_api_key": None},
            ):
                key = get_anthropic_api_key()
                assert key is None


class TestRemoveGlmKeyFromShellConfig:
    """Test removing GLM API key from shell config files."""

    def test_remove_glm_key_no_files_exist(self):
        """Test removing GLM key when no shell configs exist."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Don't create any files - test the case where they don't exist
            with patch("moai_adk.core.credentials.Path.home", return_value=Path(tmpdir)):
                result = remove_glm_key_from_shell_config()

                assert result["zshrc"] is False
                assert result["bashrc"] is False

    def test_remove_glm_key_from_zshrc(self):
        """Test removing GLM key from .zshrc."""
        with tempfile.TemporaryDirectory() as tmpdir:
            zshrc_path = Path(tmpdir) / ".zshrc"
            zshrc_path.write_text('# Some config\nexport GLM_API_KEY="test-key"\n# More config\n')

            with patch("moai_adk.core.credentials.Path.home", return_value=Path(tmpdir)):
                result = remove_glm_key_from_shell_config()

                assert result["zshrc"] is True

                # Verify export line was removed
                content = zshrc_path.read_text()
                assert "export GLM_API_KEY" not in content

                # Verify backup was created
                backup_path = Path(tmpdir) / ".zshrc.moai-backup"
                assert backup_path.exists()

    def test_remove_glm_key_not_present(self):
        """Test removing GLM key when it's not in the file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            zshrc_path = Path(tmpdir) / ".zshrc"
            zshrc_path.write_text("# Some config\n# More config\n")

            with patch("moai_adk.core.credentials.Path.home", return_value=Path(tmpdir)):
                result = remove_glm_key_from_shell_config()

                assert result["zshrc"] is False

    def test_remove_glm_key_with_whitespace(self):
        """Test removing GLM key with various whitespace patterns."""
        with tempfile.TemporaryDirectory() as tmpdir:
            zshrc_path = Path(tmpdir) / ".zshrc"
            zshrc_path.write_text('  export   GLM_API_KEY="test-key"\n')

            with patch("moai_adk.core.credentials.Path.home", return_value=Path(tmpdir)):
                result = remove_glm_key_from_shell_config()

                assert result["zshrc"] is True
                content = zshrc_path.read_text()
                assert "export GLM_API_KEY" not in content

    def test_remove_glm_key_multiple_consecutive_newlines(self):
        """Test removing GLM key cleans up multiple consecutive newlines."""
        with tempfile.TemporaryDirectory() as tmpdir:
            zshrc_path = Path(tmpdir) / ".zshrc"
            zshrc_path.write_text('# Config\nexport GLM_API_KEY="test"\n\n\n# More config\n')

            with patch("moai_adk.core.credentials.Path.home", return_value=Path(tmpdir)):
                result = remove_glm_key_from_shell_config()

                assert result["zshrc"] is True
                content = zshrc_path.read_text()
                # Should not have more than 2 consecutive newlines
                assert "\n\n\n" not in content
