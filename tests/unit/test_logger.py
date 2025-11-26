# # REMOVED_ORPHAN_TEST:LOGGING-001 | SPEC: SPEC-LOGGING-001/spec.md
"""
로깅 시스템 단위 테스트

SPEC 요구사항:
- 로그 저장: .moai/logs/moai.log
- 민감정보 마스킹: API Key, Email, Password
- 로그 레벨: development (DEBUG), test (INFO), production (WARNING)
"""

import logging
import sys
from pathlib import Path

import pytest

from moai_adk.utils.logger import SensitiveDataFilter, setup_logger


class TestLoggerSetup:
    """로거 기본 설정 테스트"""

    def test_setup_logger_creates_logger(self, tmp_path: Path):
        """TEST-LOGGING-001-01: setup_logger가 Logger 객체를 반환해야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        assert isinstance(logger, logging.Logger)
        assert logger.name == "test_logger"

    def test_setup_logger_creates_log_directory(self, tmp_path: Path):
        """TEST-LOGGING-001-02: 로그 디렉토리가 자동 생성되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        setup_logger("test_logger", log_dir=str(log_dir))

        assert log_dir.exists()
        assert log_dir.is_dir()

    def test_setup_logger_creates_log_file(self, tmp_path: Path):
        """TEST-LOGGING-001-03: moai.log 파일이 생성되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))
        logger.info("Test message")

        log_file = log_dir / "moai.log"
        assert log_file.exists()


class TestConsoleHandler:
    """콘솔 핸들러 테스트"""

    def test_console_handler_exists(self, tmp_path: Path):
        """TEST-LOGGING-001-04: 콘솔 핸들러가 추가되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        console_handlers = [h for h in logger.handlers if isinstance(h, logging.StreamHandler)]
        assert len(console_handlers) > 0

    def test_console_handler_format(self, tmp_path: Path, caplog):
        """TEST-LOGGING-001-05: 콘솔 출력 형식이 올바른지 확인"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        with caplog.at_level(logging.INFO):
            logger.info("Test message")

        # 형식: [시간] [레벨] [이름] 메시지
        assert "INFO" in caplog.text
        assert "test_logger" in caplog.text
        assert "Test message" in caplog.text


class TestFileHandler:
    """파일 핸들러 테스트"""

    def test_file_handler_exists(self, tmp_path: Path):
        """TEST-LOGGING-001-06: 파일 핸들러가 추가되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        file_handlers = [h for h in logger.handlers if isinstance(h, logging.FileHandler)]
        assert len(file_handlers) > 0

    def test_file_handler_writes_to_file(self, tmp_path: Path):
        """TEST-LOGGING-001-07: 로그가 파일에 기록되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))
        logger.info("File test message")

        log_file = log_dir / "moai.log"
        content = log_file.read_text()
        assert "File test message" in content


class TestSensitiveDataMasking:
    """민감정보 마스킹 테스트"""

    def test_api_key_masking(self, tmp_path: Path):
        """TEST-LOGGING-001-08: API Key가 마스킹되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))
        logger.info("API Key: sk-1234567890abcdef")

        log_file = log_dir / "moai.log"
        content = log_file.read_text()
        assert "sk-1234567890abcdef" not in content
        assert "***REDACTED***" in content

    def test_email_masking(self, tmp_path: Path):
        """TEST-LOGGING-001-09: 이메일이 마스킹되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))
        logger.info("User email: user@example.com")

        log_file = log_dir / "moai.log"
        content = log_file.read_text()
        assert "user@example.com" not in content
        assert "***REDACTED***" in content

    def test_password_masking(self, tmp_path: Path):
        """TEST-LOGGING-001-10: 비밀번호가 마스킹되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))
        logger.info("Password: secret123")

        log_file = log_dir / "moai.log"
        content = log_file.read_text()
        assert "secret123" not in content
        assert "***REDACTED***" in content

    def test_multiple_sensitive_data_masking(self, tmp_path: Path):
        """TEST-LOGGING-001-11: 여러 민감정보가 동시에 마스킹되어야 한다"""
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))
        logger.info("API: sk-abc, Email: test@test.com, Password: pass123")

        log_file = log_dir / "moai.log"
        content = log_file.read_text()
        assert "sk-abc" not in content
        assert "test@test.com" not in content
        assert "pass123" not in content


class TestLogLevelByEnvironment:
    """환경별 로그 레벨 테스트"""

    def test_development_mode_debug_level(self, tmp_path: Path, monkeypatch):
        """TEST-LOGGING-001-12: development 모드는 DEBUG 레벨이어야 한다"""
        monkeypatch.setenv("MOAI_ENV", "development")
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        assert logger.level == logging.DEBUG

    def test_test_mode_info_level(self, tmp_path: Path, monkeypatch):
        """TEST-LOGGING-001-13: test 모드는 INFO 레벨이어야 한다"""
        monkeypatch.setenv("MOAI_ENV", "test")
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        assert logger.level == logging.INFO

    def test_production_mode_warning_level(self, tmp_path: Path, monkeypatch):
        """TEST-LOGGING-001-14: production 모드는 WARNING 레벨이어야 한다"""
        monkeypatch.setenv("MOAI_ENV", "production")
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        assert logger.level == logging.WARNING

    def test_default_mode_info_level(self, tmp_path: Path, monkeypatch):
        """TEST-LOGGING-001-15: 환경변수 미설정 시 INFO 레벨이어야 한다"""
        monkeypatch.delenv("MOAI_ENV", raising=False)
        log_dir = tmp_path / ".moai" / "logs"
        logger = setup_logger("test_logger", log_dir=str(log_dir))

        assert logger.level == logging.INFO


class TestSensitiveDataFilterClass:
    """SensitiveDataFilter 클래스 테스트"""

    def test_filter_api_key_pattern(self):
        """TEST-LOGGING-001-16: API Key 패턴이 올바르게 필터링되어야 한다"""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="API Key: sk-1234567890",
            args=(),
            exc_info=None,
        )

        result = filter_instance.filter(record)
        assert result is True  # filter는 True를 반환 (로그 통과)
        assert "sk-1234567890" not in record.msg
        assert "***REDACTED***" in record.msg

    def test_filter_preserves_non_sensitive_data(self):
        """TEST-LOGGING-001-17: 민감정보가 아닌 데이터는 보존되어야 한다"""
        filter_instance = SensitiveDataFilter()
        record = logging.LogRecord(
            name="test",
            level=logging.INFO,
            pathname="",
            lineno=0,
            msg="Normal log message",
            args=(),
            exc_info=None,
        )

        filter_instance.filter(record)
        assert record.msg == "Normal log message"


class TestDefaultLogDirectory:
    """기본 로그 디렉토리 테스트"""

    @pytest.mark.skipif(sys.platform == "win32", reason="Windows file locking issue")
    def test_default_log_directory(self, monkeypatch):
        """TEST-LOGGING-001-18: log_dir 미지정 시 .moai/logs 사용"""
        import tempfile
        from pathlib import Path

        # 임시 디렉토리로 cwd 변경
        with tempfile.TemporaryDirectory() as tmp_dir:
            monkeypatch.chdir(tmp_dir)
            logger = setup_logger("test_logger")

            # .moai/logs 디렉토리 생성 확인
            default_log_dir = Path(tmp_dir) / ".moai" / "logs"
            assert default_log_dir.exists()

            # 로그 파일 생성 확인
            logger.info("Test default directory")
            log_file = default_log_dir / "moai.log"
            assert log_file.exists()
