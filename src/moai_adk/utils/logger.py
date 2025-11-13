"""
Logging system built on Python's logging module

SPEC requirements:
- Store logs at .moai/logs/moai.log
- Mask sensitive data: API Key, Email, Password
- Log levels: development (DEBUG), test (INFO), production (WARNING)
"""

import logging
import os
import re
from pathlib import Path


class SensitiveDataFilter(logging.Filter):
    """
    Filter that masks sensitive information.

    Automatically detects and obfuscates sensitive values in log messages.

    Supported patterns:
        - API Key: strings that start with sk-
        - Email: standard email address format
        - Password: values following password/passwd/pwd keywords

    Example:
        >>> filter_instance = SensitiveDataFilter()
        >>> record = logging.LogRecord(
        ...     name="app", level=logging.INFO, pathname="", lineno=0,
        ...     msg="API Key: sk-secret123", args=(), exc_info=None
        ... )
        >>> filter_instance.filter(record)
        >>> print(record.msg)
        API Key: ***REDACTED***
    """

    PATTERNS = [
        (r"sk-[a-zA-Z0-9]+", "***REDACTED***"),  # API Key
        (
            r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b",
            "***REDACTED***",
        ),  # Email
        (r"(?i)(password|passwd|pwd)[\s:=]+\S+", r"\1: ***REDACTED***"),  # Password
    ]

    def filter(self, record: logging.LogRecord) -> bool:
        """
        Mask sensitive data in the log record message.

        Args:
            record: Log record to inspect.

        Returns:
            True to keep the record.
        """
        message = record.getMessage()
        for pattern, replacement in self.PATTERNS:
            message = re.sub(pattern, replacement, message)

        record.msg = message
        record.args = ()  # Clear args so getMessage() returns msg unchanged

        return True


def setup_logger(
    name: str,
    log_dir: str | None = None,
    level: int | None = None,
) -> logging.Logger:
    """
    Configure and return a logger instance.

    Supports simultaneous console and file output while masking sensitive data.

    Args:
        name: Logger name (module or application).
        log_dir: Directory where logs are written.
            Default: .moai/logs (created automatically).
        level: Logging level (logging.DEBUG, INFO, WARNING, etc.).
            Default: derived from the MOAI_ENV environment variable.

    Returns:
        Configured Logger object with console and file handlers.

    Log level per environment (MOAI_ENV):
        - development: DEBUG (emit all logs)
        - test: INFO (informational and above)
        - production: WARNING (warnings and above)
        - default: INFO (when the environment variable is unset)

    Example:
        >>> logger = setup_logger("my_app")
        >>> logger.info("Application started")
        >>> logger.debug("Detailed debug info")
        >>> logger.error("Error occurred")

        # Production environment (only WARNING and above)
        >>> import os
        >>> os.environ["MOAI_ENV"] = "production"
        >>> prod_logger = setup_logger("prod_app")
        >>> prod_logger.warning("This will be logged")
        >>> prod_logger.info("This will NOT be logged")

    Notes:
        - Log files are written using UTF-8 encoding.
        - Sensitive data (API Key, Email, Password) is automatically masked.
        - Existing handlers are removed to prevent duplicates.
    """
    if level is None:
        env = os.getenv("MOAI_ENV", "").lower()
        level_map = {
            "development": logging.DEBUG,
            "test": logging.INFO,
            "production": logging.WARNING,
        }
        level = level_map.get(env, logging.INFO)

    logger = logging.getLogger(name)
    logger.setLevel(level)
    logger.handlers.clear()  # Remove existing handlers to avoid duplicates

    if log_dir is None:
        log_dir = ".moai/logs"
    log_path = Path(log_dir)
    log_path.mkdir(parents=True, exist_ok=True)

    formatter = logging.Formatter(
        fmt="[%(asctime)s] [%(levelname)s] [%(name)s] %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    console_handler = logging.StreamHandler()
    console_handler.setLevel(level)
    console_handler.setFormatter(formatter)
    console_handler.addFilter(SensitiveDataFilter())
    logger.addHandler(console_handler)

    log_file = log_path / "moai.log"
    file_handler = logging.FileHandler(log_file, encoding="utf-8")
    file_handler.setLevel(level)
    file_handler.setFormatter(formatter)
    file_handler.addFilter(SensitiveDataFilter())
    logger.addHandler(file_handler)

    return logger
