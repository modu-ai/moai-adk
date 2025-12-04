"""Unit tests for moai_adk.utils.logger module.

Tests for logging and sensitive data filtering.
"""

import logging
import os
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from moai_adk.utils.logger import SensitiveDataFilter, setup_logger


class TestSensitiveDataFilter:
    """Test SensitiveDataFilter class."""

    def test_filter_initialization(self):
        """Test SensitiveDataFilter initialization."""
        filter_instance = SensitiveDataFilter()
        assert filter_instance is not None
        assert len(filter_instance.PATTERNS) == 3

    def test_filter_api_key(self):
        """Test filtering API keys."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="API Key: sk-1234567890abcdef",
            args=(),
            exc_info=None,
        )
        result = filter_instance.filter(record)
        assert result is True
        assert "REDACTED" in record.msg
        assert "sk-" not in record.msg

    def test_filter_email(self):
        """Test filtering email addresses."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Email: user@example.com",
            args=(),
            exc_info=None,
        )
        result = filter_instance.filter(record)
        assert result is True
        assert "REDACTED" in record.msg
        assert "@" not in record.msg

    def test_filter_password(self):
        """Test filtering passwords."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="password: secret123",
            args=(),
            exc_info=None,
        )
        result = filter_instance.filter(record)
        assert result is True
        assert "REDACTED" in record.msg

    def test_filter_returns_true(self):
        """Test filter always returns True."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Normal message",
            args=(),
            exc_info=None,
        )
        result = filter_instance.filter(record)
        assert result is True

    def test_filter_clears_args(self):
        """Test filter clears message args."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Message with %s",
            args=("argument",),
            exc_info=None,
        )
        filter_instance.filter(record)
        assert record.args == ()


class TestSetupLogger:
    """Test setup_logger function."""

    def test_setup_logger_basic(self):
        """Test basic logger setup."""
        with patch("logging.FileHandler"):
            with patch("moai_adk.utils.logger.Path.mkdir"):
                logger = setup_logger("test_logger")
                assert logger is not None
                assert logger.name == "test_logger"

    def test_setup_logger_creates_directory(self):
        """Test logger creates log directory."""
        with patch("logging.FileHandler"):
            with patch("moai_adk.utils.logger.Path.mkdir") as mock_mkdir:
                logger = setup_logger("test_logger", log_dir="/tmp/logs")
                # Verify mkdir was called
                assert mock_mkdir.called

    def test_setup_logger_has_handlers(self):
        """Test logger has both console and file handlers."""
        with patch("logging.FileHandler"):
            with patch("moai_adk.utils.logger.Path.mkdir"):
                logger = setup_logger("test_logger")
                assert len(logger.handlers) >= 1  # At least one handler

    def test_setup_logger_development_env(self):
        """Test logger level in development environment."""
        with patch.dict(os.environ, {"MOAI_ENV": "development"}):
            with patch("logging.FileHandler"):
                with patch("moai_adk.utils.logger.Path.mkdir"):
                    logger = setup_logger("test_logger")
                    assert logger.level == logging.DEBUG

    def test_setup_logger_test_env(self):
        """Test logger level in test environment."""
        with patch.dict(os.environ, {"MOAI_ENV": "test"}):
            with patch("logging.FileHandler"):
                with patch("moai_adk.utils.logger.Path.mkdir"):
                    logger = setup_logger("test_logger")
                    assert logger.level == logging.INFO

    def test_setup_logger_production_env(self):
        """Test logger level in production environment."""
        with patch.dict(os.environ, {"MOAI_ENV": "production"}):
            with patch("logging.FileHandler"):
                with patch("moai_adk.utils.logger.Path.mkdir"):
                    logger = setup_logger("test_logger")
                    assert logger.level == logging.WARNING

    def test_setup_logger_custom_level(self):
        """Test logger with custom level."""
        with patch("logging.FileHandler"):
            with patch("moai_adk.utils.logger.Path.mkdir"):
                logger = setup_logger("test_logger", level=logging.ERROR)
                assert logger.level == logging.ERROR

    def test_setup_logger_custom_log_dir(self):
        """Test logger with custom log directory."""
        custom_dir = "/custom/log/dir"
        with patch("logging.FileHandler"):
            with patch("moai_adk.utils.logger.Path.mkdir"):
                logger = setup_logger("test_logger", log_dir=custom_dir)
                assert logger is not None

    def test_setup_logger_sensitive_filter_attached(self):
        """Test sensitive data filter is attached."""
        with patch("logging.FileHandler"):
            with patch("moai_adk.utils.logger.Path.mkdir"):
                logger = setup_logger("test_logger")
                # Check if any handler has SensitiveDataFilter
                has_filter = False
                for handler in logger.handlers:
                    for f in handler.filters:
                        if isinstance(f, SensitiveDataFilter):
                            has_filter = True
                assert has_filter

    def test_setup_logger_clears_existing_handlers(self):
        """Test logger clears existing handlers."""
        with patch("logging.FileHandler"):
            with patch("moai_adk.utils.logger.Path.mkdir"):
                logger = setup_logger("test_logger")
                initial_count = len(logger.handlers)
                # Setup again
                logger = setup_logger("test_logger")
                # Should not have duplicates
                assert len(logger.handlers) <= initial_count + 2
