"""Tests for rank.config module."""

import os
import stat
import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest

from moai_adk.rank.config import RankConfig, RankCredentials


class TestRankCredentials:
    """Test RankCredentials dataclass."""

    def test_credentials_creation(self):
        """Test creating credentials with all fields."""
        creds = RankCredentials(
            api_key="test_key_123",
            username="testuser",
            user_id="user_456",
            created_at="2024-01-01T00:00:00Z",
        )
        assert creds.api_key == "test_key_123"
        assert creds.username == "testuser"
        assert creds.user_id == "user_456"
        assert creds.created_at == "2024-01-01T00:00:00Z"


class TestRankConfig:
    """Test RankConfig class."""

    def test_default_initialization(self):
        """Test default initialization."""
        config = RankConfig()
        assert config.base_url == "https://rank.mo.ai.kr"
        assert config.api_base_url == "https://rank.mo.ai.kr/api/v1"

    def test_custom_base_url(self):
        """Test initialization with custom base URL."""
        config = RankConfig(base_url="https://custom.test")
        assert config.base_url == "https://custom.test"
        assert config.api_base_url == "https://custom.test/api/v1"

    def test_environment_variable_override(self):
        """Test MOAI_RANK_URL environment variable override."""
        with patch.dict(os.environ, {"MOAI_RANK_URL": "https://env.test"}):
            config = RankConfig()
            assert config.base_url == "https://env.test"

    def test_api_base_url_property(self):
        """Test api_base_url property includes version."""
        config = RankConfig(base_url="https://test.api")
        assert config.api_base_url == "https://test.api/api/v1"

    @pytest.fixture
    def temp_config_dir(self):
        """Create a temporary config directory for testing."""
        with tempfile.TemporaryDirectory() as tmpdir:
            temp_path = Path(tmpdir)
            yield temp_path

    def test_ensure_config_dir_creates_directory(self, temp_config_dir):
        """Test ensure_config_dir creates directory with correct permissions."""
        with patch.object(RankConfig, "CONFIG_DIR", temp_config_dir / "rank"):
            result = RankConfig.ensure_config_dir()
            assert result.exists()
            assert result.is_dir()

    def test_save_credentials_creates_secure_file(self, temp_config_dir):
        """Test save_credentials creates file with mode 600."""
        creds = RankCredentials(
            api_key="test_key",
            username="user",
            user_id="123",
            created_at="2024-01-01T00:00:00Z",
        )

        with patch.object(RankConfig, "CONFIG_DIR", temp_config_dir):
            with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "credentials.json"):
                RankConfig.save_credentials(creds)

                # Check file exists
                assert RankConfig.CREDENTIALS_FILE.exists()

                # Check file permissions (owner read/write only)
                file_stat = os.stat(RankConfig.CREDENTIALS_FILE)
                mode = file_stat.st_mode
                assert mode & stat.S_IRUSR  # Owner read
                assert mode & stat.S_IWUSR  # Owner write
                assert not (mode & stat.S_IRGRP)  # No group read
                assert not (mode & stat.S_IROTH)  # No other read

    def test_save_and_load_credentials(self, temp_config_dir):
        """Test saving and loading credentials."""
        creds = RankCredentials(
            api_key="saved_key",
            username="saved_user",
            user_id="saved_id",
            created_at="2024-01-01T12:00:00Z",
        )

        with patch.object(RankConfig, "CONFIG_DIR", temp_config_dir):
            with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "credentials.json"):
                RankConfig.save_credentials(creds)
                loaded = RankConfig.load_credentials()

                assert loaded is not None
                assert loaded.api_key == "saved_key"
                assert loaded.username == "saved_user"
                assert loaded.user_id == "saved_id"
                assert loaded.created_at == "2024-01-01T12:00:00Z"

    def test_load_credentials_returns_none_when_missing(self, temp_config_dir):
        """Test load_credentials returns None when file doesn't exist."""
        with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "nonexistent.json"):
            assert RankConfig.load_credentials() is None

    def test_load_credentials_handles_invalid_json(self, temp_config_dir):
        """Test load_credentials handles invalid JSON gracefully."""
        with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "invalid.json"):
            RankConfig.CREDENTIALS_FILE.write_text("invalid json {{{")
            assert RankConfig.load_credentials() is None

    def test_load_credentials_handles_missing_fields(self, temp_config_dir):
        """Test load_credentials handles missing required fields."""
        with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "incomplete.json"):
            RankConfig.CREDENTIALS_FILE.write_text('{"api_key": "test"}')
            assert RankConfig.load_credentials() is None

    def test_delete_credentials_removes_file(self, temp_config_dir):
        """Test delete_credentials removes the credentials file."""
        creds = RankCredentials(
            api_key="to_delete",
            username="user",
            user_id="123",
            created_at="2024-01-01T00:00:00Z",
        )

        with patch.object(RankConfig, "CONFIG_DIR", temp_config_dir):
            with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "credentials.json"):
                RankConfig.save_credentials(creds)
                assert RankConfig.has_credentials()

                result = RankConfig.delete_credentials()
                assert result is True
                assert not RankConfig.has_credentials()

    def test_delete_credentials_returns_false_when_missing(self, temp_config_dir):
        """Test delete_credentials returns False when file doesn't exist."""
        with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "nonexistent.json"):
            assert RankConfig.delete_credentials() is False

    def test_has_credentials(self, temp_config_dir):
        """Test has_credentials method."""
        with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "creds.json"):
            assert not RankConfig.has_credentials()

            RankConfig.CREDENTIALS_FILE.write_text("{}")
            assert RankConfig.has_credentials()

    def test_get_api_key(self, temp_config_dir):
        """Test get_api_key returns key or None."""
        creds = RankCredentials(
            api_key="secret_key",
            username="user",
            user_id="123",
            created_at="2024-01-01T00:00:00Z",
        )

        with patch.object(RankConfig, "CONFIG_DIR", temp_config_dir):
            with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "credentials.json"):
                assert RankConfig.get_api_key() is None

                RankConfig.save_credentials(creds)
                assert RankConfig.get_api_key() == "secret_key"

    def test_atomic_write_rollback_on_error(self, temp_config_dir):
        """Test that temp file is cleaned up on write error."""
        with patch.object(RankConfig, "CONFIG_DIR", temp_config_dir):
            with patch.object(RankConfig, "CREDENTIALS_FILE", temp_config_dir / "credentials.json"):
                # Mock open to raise an exception
                with patch("builtins.open", side_effect=IOError("Write error")):
                    creds = RankCredentials(
                        api_key="test",
                        username="user",
                        user_id="123",
                        created_at="2024-01-01T00:00:00Z",
                    )

                    with pytest.raises(IOError):
                        RankConfig.save_credentials(creds)

                    # Verify temp file was cleaned up
                    temp_file = RankConfig.CREDENTIALS_FILE.with_suffix(".tmp")
                    assert not temp_file.exists()
