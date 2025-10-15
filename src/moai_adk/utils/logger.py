# @CODE:LOGGING-001 | SPEC: SPEC-LOGGING-001.md | TEST: tests/unit/test_logger.py
"""
Python logging 기반 로깅 시스템

SPEC 요구사항:
- 로그 저장: .moai/logs/moai.log
- 민감정보 마스킹: API Key, Email, Password
- 로그 레벨: development (DEBUG), test (INFO), production (WARNING)
"""

import logging
import os
import re
from pathlib import Path


class SensitiveDataFilter(logging.Filter):
    """
    민감정보 마스킹 필터

    로그 메시지에서 민감정보를 자동으로 탐지하고 마스킹합니다.

    지원 패턴:
        - API Key: sk-로 시작하는 문자열
        - Email: 표준 이메일 주소 형식
        - Password: password/passwd/pwd 키워드 뒤의 값

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

    # @CODE:LOGGING-001:DOMAIN - 민감정보 패턴 정의
    PATTERNS = [
        (r"sk-[a-zA-Z0-9]+", "***REDACTED***"),  # API Key
        (r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b", "***REDACTED***"),  # Email
        (r"(?i)(password|passwd|pwd)[\s:=]+\S+", r"\1: ***REDACTED***"),  # Password
    ]

    def filter(self, record: logging.LogRecord) -> bool:
        """
        로그 레코드의 메시지에서 민감정보를 마스킹

        Args:
            record: 로그 레코드

        Returns:
            True (로그를 통과시킴)
        """
        message = record.getMessage()
        for pattern, replacement in self.PATTERNS:
            message = re.sub(pattern, replacement, message)

        record.msg = message
        record.args = ()  # args를 비워서 getMessage()가 msg를 그대로 반환하도록

        return True


def setup_logger(
    name: str,
    log_dir: str | None = None,
    level: int | None = None,
) -> logging.Logger:
    """
    로거 설정 및 반환

    콘솔 출력과 파일 저장을 동시에 지원하며, 민감정보를 자동으로 마스킹합니다.

    Args:
        name: 로거 이름 (모듈명 또는 애플리케이션명)
        log_dir: 로그 디렉토리 경로
            기본값: .moai/logs (자동 생성)
        level: 로그 레벨 (logging.DEBUG, INFO, WARNING 등)
            기본값: 환경변수 MOAI_ENV 기반 자동 결정

    Returns:
        설정된 Logger 객체 (콘솔 + 파일 핸들러 포함)

    환경별 로그 레벨 (MOAI_ENV):
        - development: DEBUG (모든 로그 출력)
        - test: INFO (정보성 로그 이상)
        - production: WARNING (경고 이상만 출력)
        - default: INFO (환경변수 미설정 시)

    Example:
        >>> logger = setup_logger("my_app")
        >>> logger.info("Application started")
        >>> logger.debug("Detailed debug info")
        >>> logger.error("Error occurred")

        # 프로덕션 환경 (WARNING 이상만 출력)
        >>> import os
        >>> os.environ["MOAI_ENV"] = "production"
        >>> prod_logger = setup_logger("prod_app")
        >>> prod_logger.warning("This will be logged")
        >>> prod_logger.info("This will NOT be logged")

    Notes:
        - 로그 파일은 UTF-8 인코딩으로 저장됩니다
        - 민감정보(API Key, Email, Password)는 자동 마스킹됩니다
        - 중복 핸들러 방지를 위해 기존 핸들러를 제거합니다
    """
    # @CODE:LOGGING-001:DOMAIN - 로그 레벨 결정
    if level is None:
        env = os.getenv("MOAI_ENV", "").lower()
        level_map = {
            "development": logging.DEBUG,
            "test": logging.INFO,
            "production": logging.WARNING,
        }
        level = level_map.get(env, logging.INFO)

    # @CODE:LOGGING-001:INFRA - 로거 생성 및 설정
    logger = logging.getLogger(name)
    logger.setLevel(level)
    logger.handlers.clear()  # 기존 핸들러 제거 (중복 방지)

    # @CODE:LOGGING-001:INFRA - 로그 디렉토리 생성
    if log_dir is None:
        log_dir = ".moai/logs"
    log_path = Path(log_dir)
    log_path.mkdir(parents=True, exist_ok=True)

    # @CODE:LOGGING-001:INFRA - 로그 포맷 정의
    formatter = logging.Formatter(
        fmt="[%(asctime)s] [%(levelname)s] [%(name)s] %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    # @CODE:LOGGING-001:INFRA - 콘솔 핸들러
    console_handler = logging.StreamHandler()
    console_handler.setLevel(level)
    console_handler.setFormatter(formatter)
    console_handler.addFilter(SensitiveDataFilter())
    logger.addHandler(console_handler)

    # @CODE:LOGGING-001:INFRA - 파일 핸들러
    log_file = log_path / "moai.log"
    file_handler = logging.FileHandler(log_file, encoding="utf-8")
    file_handler.setLevel(level)
    file_handler.setFormatter(formatter)
    file_handler.addFilter(SensitiveDataFilter())
    logger.addHandler(file_handler)

    return logger
