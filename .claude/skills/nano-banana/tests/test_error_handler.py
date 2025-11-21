"""
ErrorHandler 테스트

API 에러 처리 및 재시도 전략 검증
"""

import pytest
from modules.error_handler import ErrorHandler, FinishReasonHandler


class TestErrorHandler:
    """ErrorHandler 테스트 클래스"""

    def test_extract_error_code_dict(self):
        """dict 형식 에러 코드 추출"""
        error_response = {
            "error": {
                "code": "RESOURCE_EXHAUSTED",
                "message": "Too many requests"
            }
        }
        handler = ErrorHandler(error_response)
        assert handler.get_error_code() == "RESOURCE_EXHAUSTED"

    def test_extract_error_code_string(self):
        """string 형식 에러 코드 추출"""
        error_response = {"error": "INVALID_ARGUMENT"}
        handler = ErrorHandler(error_response)
        assert handler.get_error_code() == "INVALID_ARGUMENT"

    def test_is_retryable_resource_exhausted(self):
        """RESOURCE_EXHAUSTED 재시도 가능 확인"""
        error_response = {
            "error": {"code": "RESOURCE_EXHAUSTED", "message": "Rate limit"}
        }
        handler = ErrorHandler(error_response)
        assert handler.is_retryable() is True

    def test_is_retryable_internal_error(self):
        """INTERNAL 에러 재시도 가능 확인"""
        error_response = {
            "error": {"code": "INTERNAL", "message": "Server error"}
        }
        handler = ErrorHandler(error_response)
        assert handler.is_retryable() is True

    def test_is_not_retryable_safety(self):
        """SAFETY 에러 재시도 불가 확인"""
        error_response = {
            "error": {"code": "SAFETY", "message": "Content blocked"}
        }
        handler = ErrorHandler(error_response)
        assert handler.is_retryable() is False

    def test_is_not_retryable_recitation(self):
        """RECITATION 에러 재시도 불가 확인"""
        error_response = {
            "error": {"code": "RECITATION", "message": "Similar to training data"}
        }
        handler = ErrorHandler(error_response)
        assert handler.is_retryable() is False

    def test_is_not_retryable_invalid_argument(self):
        """INVALID_ARGUMENT 에러 재시도 불가 확인"""
        error_response = {
            "error": {"code": "INVALID_ARGUMENT", "message": "Bad input"}
        }
        handler = ErrorHandler(error_response)
        assert handler.is_retryable() is False

    def test_get_retry_delay_exponential_backoff(self):
        """지수 백오프 재시도 대기 시간"""
        error_response = {
            "error": {"code": "RESOURCE_EXHAUSTED", "message": "Rate limit"}
        }
        handler = ErrorHandler(error_response)

        # 첫 시도
        delay1 = handler.get_retry_delay()
        assert delay1 > 0

        # 두 번째 시도 (지수 증가)
        handler.increment_retry_count()
        delay2 = handler.get_retry_delay()
        assert delay2 > delay1

    def test_get_retry_delay_max_limit(self):
        """최대 재시도 대기 시간 제한"""
        error_response = {
            "error": {"code": "RESOURCE_EXHAUSTED", "message": "Rate limit"}
        }
        handler = ErrorHandler(error_response)

        # 매우 많은 재시도 후
        for _ in range(20):
            handler.increment_retry_count()

        delay = handler.get_retry_delay()
        assert delay <= 1800  # 최대 30분

    def test_get_max_retries(self):
        """최대 재시도 횟수"""
        error_response = {
            "error": {"code": "RESOURCE_EXHAUSTED", "message": "Rate limit"}
        }
        handler = ErrorHandler(error_response)

        max_retries = handler.get_max_retries()
        assert max_retries > 0

    def test_get_message_retryable_error(self):
        """Retryable error message"""
        error_response = {
            "error": {"code": "RESOURCE_EXHAUSTED", "message": "Rate limited"}
        }
        handler = ErrorHandler(error_response)

        message = handler.get_message()
        assert "temporary" in message or "Retrying" in message

    def test_get_message_safety_error(self):
        """Safety error message"""
        error_response = {
            "error": {"code": "SAFETY", "message": "Blocked"}
        }
        handler = ErrorHandler(error_response)

        message = handler.get_message()
        assert "safety" in message or "policy" in message

    def test_get_message_recitation_error(self):
        """Data similarity error message"""
        error_response = {
            "error": {"code": "RECITATION", "message": "Similar data"}
        }
        handler = ErrorHandler(error_response)

        message = handler.get_message()
        assert "training" in message or "data" in message or "similar" in message

    def test_get_resolution_action(self):
        """에러 해결 방법"""
        error_response = {
            "error": {"code": "SAFETY", "message": "Blocked"}
        }
        handler = ErrorHandler(error_response)

        action = handler.get_resolution_action()
        assert len(action) > 0

    def test_error_details(self):
        """에러 상세 정보"""
        error_response = {
            "error": {
                "code": "RESOURCE_EXHAUSTED",
                "message": "Rate limited",
                "details": {"retry_after": 60}
            }
        }
        handler = ErrorHandler(error_response)

        details = handler.get_error_details()
        assert details["code"] == "RESOURCE_EXHAUSTED"
        assert details["retryable"] is True

    def test_increment_retry_count(self):
        """재시도 횟수 증가"""
        error_response = {
            "error": {"code": "INTERNAL", "message": "Error"}
        }
        handler = ErrorHandler(error_response)

        initial_count = handler.retry_count
        handler.increment_retry_count()
        assert handler.retry_count == initial_count + 1


class TestFinishReasonHandler:
    """FinishReasonHandler 테스트"""

    def test_is_successful_stop(self):
        """STOP 완료 이유 성공 확인"""
        assert FinishReasonHandler.is_successful("STOP") is True

    def test_is_successful_max_tokens(self):
        """MAX_TOKENS 완료 이유 성공 여부"""
        assert FinishReasonHandler.is_successful("MAX_TOKENS") is False

    def test_is_successful_safety(self):
        """SAFETY 완료 이유 성공 여부"""
        assert FinishReasonHandler.is_successful("SAFETY") is False

    def test_get_description(self):
        """완료 이유 설명"""
        desc = FinishReasonHandler.get_description("STOP")
        assert len(desc) > 0

    def test_is_retryable_stop(self):
        """STOP 재시도 여부"""
        assert FinishReasonHandler.is_retryable("STOP") is True

    def test_is_retryable_safety(self):
        """SAFETY 재시도 여부"""
        assert FinishReasonHandler.is_retryable("SAFETY") is False

    def test_is_retryable_recitation(self):
        """RECITATION 재시도 여부"""
        assert FinishReasonHandler.is_retryable("RECITATION") is False

    def test_is_retryable_max_tokens(self):
        """MAX_TOKENS 재시도 여부"""
        assert FinishReasonHandler.is_retryable("MAX_TOKENS") is True


class TestErrorHandlerEdgeCases:
    """엣지 케이스 테스트"""

    def test_unknown_error_code(self):
        """알 수 없는 에러 코드"""
        error_response = {
            "error": {"code": "UNKNOWN_CODE", "message": "Unknown"}
        }
        handler = ErrorHandler(error_response)

        assert handler.get_error_code() == "UNKNOWN_CODE"
        assert handler.is_retryable() is False

    def test_missing_error_message(self):
        """에러 메시지 없음"""
        error_response = {"error": {"code": "INTERNAL"}}
        handler = ErrorHandler(error_response)

        message = handler.get_message()
        assert len(message) > 0

    def test_empty_error_response(self):
        """빈 에러 응답"""
        error_response = {}
        handler = ErrorHandler(error_response)

        assert handler.get_error_code() == "UNKNOWN"

    def test_malformed_error_response(self):
        """잘못된 형식의 에러 응답"""
        error_response = {"error": 123}  # 숫자로 전달
        handler = ErrorHandler(error_response)

        # 처리 가능해야 함
        assert handler.get_error_code() is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
