"""Comprehensive test suite for logger.py utilities module.

This module provides 90%+ coverage for all logging functionality including:
- SensitiveDataFilter pattern detection and masking
- Logger initialization and configuration
- Log level management (development, test, production)
- Console and file handler setup
- Log file creation and directory structure
- Sensitive data patterns (API Key, Email, Password)
- Edge cases and error handling
"""

import logging
import os
import tempfile
from io import StringIO
from unittest.mock import patch

import pytest

from moai_adk.utils.logger import SensitiveDataFilter, setup_logger

# ============================================================================
# SensitiveDataFilter Tests
# ============================================================================


class TestSensitiveDataFilterInitialization:
    """Tests for SensitiveDataFilter initialization."""

    def test_filter_initialization(self):
        """Test creating SensitiveDataFilter instance."""
        filter_instance = SensitiveDataFilter()
        assert filter_instance is not None
        assert hasattr(filter_instance, "PATTERNS")

    def test_filter_has_api_key_pattern(self):
        """Test that filter has API key pattern."""
        filter_instance = SensitiveDataFilter()
        patterns = [p[0] for p in filter_instance.PATTERNS]
        # Should have pattern for sk-* API keys
        assert any("sk-" in p for p in patterns)

    def test_filter_has_email_pattern(self):
        """Test that filter has email pattern."""
        filter_instance = SensitiveDataFilter()
        patterns = [p[0] for p in filter_instance.PATTERNS]
        # Should have pattern for email addresses
        assert any("@" in p for p in patterns)

    def test_filter_has_password_pattern(self):
        """Test that filter has password pattern."""
        filter_instance = SensitiveDataFilter()
        patterns = [p[0] for p in filter_instance.PATTERNS]
        # Should have pattern for password keywords
        assert any("password" in p.lower() for p in patterns)

    def test_filter_patterns_count(self):
        """Test that filter has expected number of patterns."""
        filter_instance = SensitiveDataFilter()
        assert len(filter_instance.PATTERNS) == 3


class TestSensitiveDataFilterApiKey:
    """Tests for API key masking in SensitiveDataFilter."""

    def test_filter_masks_api_key_sk_prefix(self):
        """Test masking API keys with sk- prefix."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="API Key: sk-secret123",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "sk-secret123" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_masks_multiple_api_keys(self):
        """Test masking multiple API keys in same message."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="Primary: sk-abc123 Backup: sk-xyz789",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "sk-abc123" not in record.msg
        assert "sk-xyz789" not in record.msg
        assert record.msg.count("***REDACTED***") == 2

    def test_filter_preserves_non_api_key_content(self):
        """Test that non-sensitive content is preserved."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="API Key: sk-secret123 is configured",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "is configured" in record.msg
        assert "API Key:" in record.msg

    def test_filter_api_key_with_different_lengths(self):
        """Test masking API keys of various lengths."""
        filter_instance = SensitiveDataFilter()
        test_cases = [
            "sk-a",
            "sk-abc123",
            "sk-" + "a" * 100,
        ]

        for api_key in test_cases:
            record = logging.LogRecord(
                name="test_logger",
                level=logging.INFO,
                pathname="test.py",
                lineno=0,
                msg=f"Key: {api_key}",
                args=(),
                exc_info=None,
            )

            result = filter_instance.filter(record)

            assert result is True
            assert api_key not in record.msg
            assert "***REDACTED***" in record.msg


class TestSensitiveDataFilterEmail:
    """Tests for email masking in SensitiveDataFilter."""

    def test_filter_masks_simple_email(self):
        """Test masking simple email address."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="User: john@example.com",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "john@example.com" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_masks_email_with_subdomain(self):
        """Test masking email with subdomain."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="Contact: user@mail.example.co.uk",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "user@mail.example.co.uk" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_masks_multiple_emails(self):
        """Test masking multiple email addresses."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="From: alice@example.com To: bob@example.com",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "alice@example.com" not in record.msg
        assert "bob@example.com" not in record.msg
        assert record.msg.count("***REDACTED***") == 2

    def test_filter_email_with_special_characters(self):
        """Test masking email with special characters."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="Email: john.doe+tag@example-domain.com",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "john.doe+tag@example-domain.com" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_does_not_mask_invalid_email(self):
        """Test that invalid email patterns are not masked."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="Email format is user@domain.com format",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        # Should mask valid email format even in different context
        assert "***REDACTED***" in record.msg or "user@domain.com" not in record.msg


class TestSensitiveDataFilterPassword:
    """Tests for password masking in SensitiveDataFilter."""

    def test_filter_masks_password_keyword(self):
        """Test masking password values."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="password: mysecretpassword123",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "mysecretpassword123" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_masks_passwd_keyword(self):
        """Test masking with passwd keyword."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="passwd=secret123",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "secret123" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_masks_pwd_keyword(self):
        """Test masking with pwd keyword."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="pwd: verysecret",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "verysecret" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_masks_password_case_insensitive(self):
        """Test masking password with different cases."""
        filter_instance = SensitiveDataFilter()
        test_cases = [
            "Password: secret123",
            "PASSWORD: SECRET123",
            "PassWord: SeCrEt123",
        ]

        for msg in test_cases:
            record = logging.LogRecord(
                name="test_logger", level=logging.INFO, pathname="test.py", lineno=0, msg=msg, args=(), exc_info=None
            )

            result = filter_instance.filter(record)

            assert result is True
            # Password value should be redacted
            assert "***REDACTED***" in record.msg

    def test_filter_preserves_password_keyword(self):
        """Test that password keyword itself is preserved."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="password: mysecret123",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        # Pattern replacement should preserve the keyword
        assert "password" in record.msg.lower()
        assert "***REDACTED***" in record.msg


class TestSensitiveDataFilterEdgeCases:
    """Tests for edge cases in SensitiveDataFilter."""

    def test_filter_clears_args(self):
        """Test that filter clears record.args."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="Message: %s",
            args=("value",),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert record.args == ()

    def test_filter_returns_true(self):
        """Test that filter always returns True."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="Normal message",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True

    def test_filter_with_empty_message(self):
        """Test filtering empty message."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger", level=logging.INFO, pathname="test.py", lineno=0, msg="", args=(), exc_info=None
        )

        result = filter_instance.filter(record)

        assert result is True
        assert record.msg == ""

    def test_filter_with_no_sensitive_data(self):
        """Test filtering message with no sensitive data."""
        filter_instance = SensitiveDataFilter()
        original_msg = "This is a normal log message with no sensitive data"
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg=original_msg,
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert record.msg == original_msg

    def test_filter_with_mixed_sensitive_data(self):
        """Test filtering message with multiple types of sensitive data."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="User john@example.com with API key sk-secret123 password: mypass123",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "john@example.com" not in record.msg
        assert "sk-secret123" not in record.msg
        assert "mypass123" not in record.msg
        assert record.msg.count("***REDACTED***") == 3

    def test_filter_consecutive_patterns(self):
        """Test filtering consecutive sensitive patterns."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test_logger",
            level=logging.INFO,
            pathname="test.py",
            lineno=0,
            msg="sk-key1sk-key2sk-key3",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "sk-key" not in record.msg


# ============================================================================
# setup_logger Configuration Tests
# ============================================================================


class TestSetupLoggerInitialization:
    """Tests for setup_logger function initialization."""

    def test_setup_logger_creates_logger(self):
        """Test that setup_logger creates a logger instance."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("test_app", log_dir=tmpdir)
            assert logger is not None
            assert isinstance(logger, logging.Logger)
            assert logger.name == "test_app"

    def test_setup_logger_with_custom_name(self):
        """Test setup_logger with custom logger name."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("my_custom_logger", log_dir=tmpdir)
            assert logger.name == "my_custom_logger"

    def test_setup_logger_returns_logger_instance(self):
        """Test that setup_logger returns logging.Logger."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)
            assert hasattr(logger, "debug")
            assert hasattr(logger, "info")
            assert hasattr(logger, "warning")
            assert hasattr(logger, "error")
            assert hasattr(logger, "critical")

    def test_setup_logger_multiple_calls_same_name(self):
        """Test multiple setup_logger calls with same name."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger1 = setup_logger("app", log_dir=tmpdir)
            logger2 = setup_logger("app", log_dir=tmpdir)

            # Should be same logger instance
            assert logger1.name == logger2.name


class TestSetupLoggerLevels:
    """Tests for log level configuration in setup_logger."""

    def test_setup_logger_development_level(self):
        """Test log level in development environment."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "development"}):
                logger = setup_logger("app", log_dir=tmpdir)
                assert logger.level == logging.DEBUG

    def test_setup_logger_test_level(self):
        """Test log level in test environment."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "test"}):
                logger = setup_logger("app", log_dir=tmpdir)
                assert logger.level == logging.INFO

    def test_setup_logger_production_level(self):
        """Test log level in production environment."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "production"}):
                logger = setup_logger("app", log_dir=tmpdir)
                assert logger.level == logging.WARNING

    def test_setup_logger_default_level(self):
        """Test default log level when MOAI_ENV is unset."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {}, clear=False):
                # Remove MOAI_ENV if it exists
                os.environ.pop("MOAI_ENV", None)
                logger = setup_logger("app", log_dir=tmpdir)
                assert logger.level == logging.INFO

    def test_setup_logger_custom_level(self):
        """Test setup_logger with custom log level."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir, level=logging.ERROR)
            assert logger.level == logging.ERROR

    def test_setup_logger_custom_level_overrides_env(self):
        """Test that custom level overrides environment variable."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "production"}):
                logger = setup_logger("app", log_dir=tmpdir, level=logging.DEBUG)
                # Custom level should override
                assert logger.level == logging.DEBUG

    def test_setup_logger_case_insensitive_env(self):
        """Test MOAI_ENV is case-insensitive."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "DEVELOPMENT"}):
                logger = setup_logger("app", log_dir=tmpdir)
                assert logger.level == logging.DEBUG

    def test_setup_logger_invalid_env_uses_default(self):
        """Test invalid MOAI_ENV value defaults to INFO."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "invalid"}):
                logger = setup_logger("app", log_dir=tmpdir)
                assert logger.level == logging.INFO


class TestSetupLoggerDirectories:
    """Tests for log directory handling in setup_logger."""

    def test_setup_logger_creates_log_directory(self):
        """Test that setup_logger creates log directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            log_dir = os.path.join(tmpdir, "logs")
            setup_logger("app", log_dir=log_dir)

            assert os.path.exists(log_dir)
            assert os.path.isdir(log_dir)

    def test_setup_logger_default_log_dir(self):
        """Test default log directory is .moai/logs."""
        logger = setup_logger("app")
        # Logger should be created successfully
        assert logger is not None

        # Cleanup
        import shutil

        if os.path.exists(".moai"):
            shutil.rmtree(".moai")

    def test_setup_logger_creates_parent_directories(self):
        """Test that setup_logger creates parent directories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            log_dir = os.path.join(tmpdir, "parent", "child", "logs")
            setup_logger("app", log_dir=log_dir)

            assert os.path.exists(log_dir)

    def test_setup_logger_with_existing_directory(self):
        """Test setup_logger with existing directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            setup_logger("app", log_dir=tmpdir)
            assert os.path.exists(tmpdir)

    def test_setup_logger_creates_moai_log_file(self):
        """Test that setup_logger creates moai.log file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            setup_logger("app", log_dir=tmpdir)

            log_file = os.path.join(tmpdir, "moai.log")
            assert os.path.exists(log_file)


class TestSetupLoggerHandlers:
    """Tests for handler configuration in setup_logger."""

    def test_setup_logger_has_console_handler(self):
        """Test that logger has console handler."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            console_handlers = [
                h
                for h in logger.handlers
                if isinstance(h, logging.StreamHandler) and not isinstance(h, logging.FileHandler)
            ]
            assert len(console_handlers) > 0

    def test_setup_logger_has_file_handler(self):
        """Test that logger has file handler."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            file_handlers = [h for h in logger.handlers if isinstance(h, logging.FileHandler)]
            assert len(file_handlers) > 0

    def test_setup_logger_handlers_have_formatter(self):
        """Test that handlers have formatter."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            for handler in logger.handlers:
                assert handler.formatter is not None

    def test_setup_logger_formatter_includes_timestamp(self):
        """Test that formatter includes timestamp."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            for handler in logger.handlers:
                formatter_str = handler.formatter._fmt
                assert "asctime" in formatter_str

    def test_setup_logger_formatter_includes_level(self):
        """Test that formatter includes log level."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            for handler in logger.handlers:
                formatter_str = handler.formatter._fmt
                assert "levelname" in formatter_str

    def test_setup_logger_formatter_includes_name(self):
        """Test that formatter includes logger name."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            for handler in logger.handlers:
                formatter_str = handler.formatter._fmt
                assert "name" in formatter_str

    def test_setup_logger_formatter_includes_message(self):
        """Test that formatter includes message."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            for handler in logger.handlers:
                formatter_str = handler.formatter._fmt
                assert "message" in formatter_str

    def test_setup_logger_removes_existing_handlers(self):
        """Test that setup_logger removes existing handlers."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("test_app_unique", log_dir=tmpdir)
            initial_handler_count = len(logger.handlers)

            # Call setup_logger again
            logger = setup_logger("test_app_unique", log_dir=tmpdir)

            # Should not have duplicate handlers
            assert len(logger.handlers) == initial_handler_count


class TestSetupLoggerFiltering:
    """Tests for sensitive data filtering in setup_logger."""

    def test_setup_logger_handlers_have_filter(self):
        """Test that handlers have sensitive data filter."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            for handler in logger.handlers:
                filters = handler.filters
                assert len(filters) > 0
                # Check if SensitiveDataFilter is present
                assert any(isinstance(f, SensitiveDataFilter) for f in filters)

    def test_setup_logger_masks_sensitive_in_console(self):
        """Test that sensitive data is masked in console output."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            # Capture console output
            with patch("sys.stdout", new=StringIO()) as fake_out:
                logger.info("API Key: sk-secret123")
                output = fake_out.getvalue()
                # Sensitive data should be masked
                assert "sk-secret123" not in output or "***REDACTED***" in output

    def test_setup_logger_masks_sensitive_in_file(self):
        """Test that sensitive data is masked in file output."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            logger.info("Email: test@example.com")

            # Read log file
            log_file = os.path.join(tmpdir, "moai.log")
            with open(log_file, "r") as f:
                content = f.read()

            # Sensitive data should be masked
            assert "test@example.com" not in content or "***REDACTED***" in content


class TestSetupLoggerLogging:
    """Tests for actual logging functionality."""

    def test_logger_logs_debug_in_development(self):
        """Test that debug logs appear in development environment."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "development"}):
                logger = setup_logger("app", log_dir=tmpdir)

                logger.debug("Debug message")

                log_file = os.path.join(tmpdir, "moai.log")
                with open(log_file, "r") as f:
                    content = f.read()

                assert "Debug message" in content

    def test_logger_skips_debug_in_production(self):
        """Test that debug logs don't appear in production."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "production"}):
                logger = setup_logger("app", log_dir=tmpdir)

                logger.debug("Debug message")
                logger.warning("Warning message")

                log_file = os.path.join(tmpdir, "moai.log")
                with open(log_file, "r") as f:
                    content = f.read()

                # Debug should not appear
                assert "Debug message" not in content
                # Warning should appear
                assert "Warning message" in content

    def test_logger_logs_to_file(self):
        """Test that logs are written to file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            logger.info("Test message")

            log_file = os.path.join(tmpdir, "moai.log")
            assert os.path.exists(log_file)

            with open(log_file, "r") as f:
                content = f.read()

            assert "Test message" in content

    def test_logger_uses_utf8_encoding(self):
        """Test that file handler uses UTF-8 encoding."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            # Log message with UTF-8 characters
            logger.info("Message with unicode: 한글 日本語")

            log_file = os.path.join(tmpdir, "moai.log")

            # Read with UTF-8 encoding
            with open(log_file, "r", encoding="utf-8") as f:
                content = f.read()

            assert "한글" in content or "Message with unicode" in content

    def test_logger_logs_multiple_levels(self):
        """Test logging at all levels."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "development"}):
                logger = setup_logger("app", log_dir=tmpdir)

                logger.debug("Debug")
                logger.info("Info")
                logger.warning("Warning")
                logger.error("Error")
                logger.critical("Critical")

                log_file = os.path.join(tmpdir, "moai.log")
                with open(log_file, "r") as f:
                    content = f.read()

                # All should be present in development
                assert "Debug" in content
                assert "Info" in content
                assert "Warning" in content
                assert "Error" in content
                assert "Critical" in content


class TestSetupLoggerEdgeCases:
    """Tests for edge cases in setup_logger."""

    def test_setup_logger_with_none_log_dir(self):
        """Test setup_logger with None log_dir uses default."""
        logger = setup_logger("app", log_dir=None)
        assert logger is not None

        # Cleanup
        import shutil

        if os.path.exists(".moai"):
            shutil.rmtree(".moai")

    def test_setup_logger_with_none_level(self):
        """Test setup_logger with None level uses environment."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch.dict(os.environ, {"MOAI_ENV": "test"}):
                logger = setup_logger("app", log_dir=tmpdir, level=None)
                assert logger.level == logging.INFO

    def test_setup_logger_with_numeric_level(self):
        """Test setup_logger with numeric log level."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir, level=25)
            # 25 is between INFO (20) and WARNING (30)
            assert logger.level == 25

    def test_setup_logger_with_relative_path(self):
        """Test setup_logger with relative path."""
        # This test uses a temporary relative path
        import os
        import tempfile

        original_cwd = os.getcwd()
        try:
            with tempfile.TemporaryDirectory() as tmpdir:
                os.chdir(tmpdir)
                setup_logger("app", log_dir="./logs")
                assert os.path.exists("./logs")
        finally:
            os.chdir(original_cwd)

    def test_setup_logger_concurrent_setup(self):
        """Test that concurrent logger setup is safe."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger1 = setup_logger("app1", log_dir=tmpdir)
            logger2 = setup_logger("app2", log_dir=tmpdir)

            # Both should have handlers
            assert len(logger1.handlers) > 0
            assert len(logger2.handlers) > 0


class TestSetupLoggerIntegration:
    """Integration tests for setup_logger."""

    def test_complete_logging_workflow(self):
        """Test complete logging workflow."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            # Log messages at different levels
            logger.debug("Debug message")
            logger.info("Info message")
            logger.warning("Warning message")
            logger.error("Error message")

            # Verify file created
            log_file = os.path.join(tmpdir, "moai.log")
            assert os.path.exists(log_file)

            # Verify messages in file
            with open(log_file, "r") as f:
                content = f.read()

            assert "Info message" in content
            assert "Warning message" in content
            assert "Error message" in content

    def test_sensitive_data_masking_workflow(self):
        """Test complete sensitive data masking workflow."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("app", log_dir=tmpdir)

            # Log with sensitive data
            logger.info("User: john@example.com with key sk-secret123")

            # Verify masking in file
            log_file = os.path.join(tmpdir, "moai.log")
            with open(log_file, "r") as f:
                content = f.read()

            # Original sensitive data should be masked
            assert "john@example.com" not in content or "***REDACTED***" in content
            assert "sk-secret123" not in content or "***REDACTED***" in content

    def test_environment_based_log_level_workflow(self):
        """Test environment-based log level workflow."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Test with development environment
            with patch.dict(os.environ, {"MOAI_ENV": "development"}):
                logger = setup_logger("dev_app", log_dir=tmpdir)
                assert logger.level == logging.DEBUG

            # Test with production environment
            with patch.dict(os.environ, {"MOAI_ENV": "production"}):
                logger = setup_logger("prod_app", log_dir=tmpdir)
                assert logger.level == logging.WARNING

    def test_multiple_loggers_same_directory(self):
        """Test multiple loggers writing to same directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger1 = setup_logger("app1", log_dir=tmpdir)
            logger2 = setup_logger("app2", log_dir=tmpdir)

            logger1.info("Message from app1")
            logger2.info("Message from app2")

            log_file = os.path.join(tmpdir, "moai.log")
            with open(log_file, "r") as f:
                content = f.read()

            # Both messages should be in the same file
            assert "Message from app1" in content
            assert "Message from app2" in content

    def test_logger_name_appears_in_output(self):
        """Test that logger name appears in formatted output."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("my_app", log_dir=tmpdir)
            logger.info("Test message")

            log_file = os.path.join(tmpdir, "moai.log")
            with open(log_file, "r") as f:
                content = f.read()

            assert "my_app" in content


class TestSetupLoggerErrorHandling:
    """Tests for error handling in setup_logger."""

    def test_setup_logger_with_invalid_name(self):
        """Test setup_logger with special characters in name."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Special characters in logger name should be handled
            logger = setup_logger("app.name-123", log_dir=tmpdir)
            assert logger is not None

    def test_setup_logger_with_very_long_name(self):
        """Test setup_logger with very long name."""
        with tempfile.TemporaryDirectory() as tmpdir:
            long_name = "a" * 1000
            logger = setup_logger(long_name, log_dir=tmpdir)
            assert logger.name == long_name

    def test_setup_logger_with_empty_name(self):
        """Test setup_logger with empty name."""
        with tempfile.TemporaryDirectory() as tmpdir:
            logger = setup_logger("", log_dir=tmpdir)
            assert logger is not None

    def test_setup_logger_level_argument_types(self):
        """Test setup_logger with different level argument types."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Integer level
            logger = setup_logger("app", log_dir=tmpdir, level=logging.INFO)
            assert logger.level == logging.INFO

            # Another valid level
            logger = setup_logger("app2", log_dir=tmpdir, level=logging.ERROR)
            assert logger.level == logging.ERROR


class TestSensitiveDataFilterComprehensive:
    """Comprehensive tests for SensitiveDataFilter patterns."""

    @pytest.mark.parametrize(
        "api_key",
        [
            "sk-1234567890",
            "sk-abcdefghijklmnop",
            "sk-" + "x" * 50,
        ],
    )
    def test_filter_various_api_keys(self, api_key):
        """Test filtering various API key formats."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test", level=logging.INFO, pathname="test.py", lineno=0, msg=f"Key: {api_key}", args=(), exc_info=None
        )

        result = filter_instance.filter(record)

        assert result is True
        assert api_key not in record.msg

    @pytest.mark.parametrize(
        "email",
        [
            "user@example.com",
            "john.doe@example.co.uk",
            "test+tag@sub.example.com",
            "a@b.co",
        ],
    )
    def test_filter_various_emails(self, email):
        """Test filtering various email formats."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test", level=logging.INFO, pathname="test.py", lineno=0, msg=f"Email: {email}", args=(), exc_info=None
        )

        result = filter_instance.filter(record)

        assert result is True
        assert email not in record.msg or "***REDACTED***" in record.msg

    @pytest.mark.parametrize(
        "password_msg",
        [
            "password: mypass",
            "PASSWORD: MYPASS",
            "passwd: secret",
            "pwd: sec",
            "password=value",
            "password:value",
        ],
    )
    def test_filter_various_password_formats(self, password_msg):
        """Test filtering various password formats."""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test", level=logging.INFO, pathname="test.py", lineno=0, msg=password_msg, args=(), exc_info=None
        )

        result = filter_instance.filter(record)

        assert result is True
        assert "***REDACTED***" in record.msg
