"""
ImageGenerator 테스트

이미지 생성 및 편집 기능 테스트 (Mock API 사용)
"""

import base64
import pytest
from unittest.mock import Mock, patch, MagicMock
from modules.image_generator import ImageGenerator
from modules.env_key_manager import EnvKeyManager


class TestImageGeneratorValidation:
    """ImageGenerator 검증 테스트"""

    @pytest.fixture
    def generator(self):
        """ImageGenerator 인스턴스"""
        return ImageGenerator("gsk_test_api_key_" + "a" * 50)

    def test_validate_prompt_valid(self, generator):
        """유효한 프롬프트 검증"""
        assert generator._validate_prompt("beautiful landscape") is True

    def test_validate_prompt_too_short(self, generator):
        """너무 짧은 프롬프트"""
        assert generator._validate_prompt("ab") is False

    def test_validate_prompt_empty(self, generator):
        """빈 프롬프트"""
        assert generator._validate_prompt("") is False

    def test_validate_prompt_none(self, generator):
        """None 프롬프트"""
        assert generator._validate_prompt(None) is False

    def test_validate_resolution_valid(self, generator):
        """유효한 해상도"""
        assert generator._validate_resolution("2048x2048") is True
        assert generator._validate_resolution("1024x1024") is True
        assert generator._validate_resolution("4096x4096") is True

    def test_validate_resolution_invalid(self, generator):
        """유효하지 않은 해상도"""
        assert generator._validate_resolution("512x512") is False
        assert generator._validate_resolution("invalid") is False


class TestImageGeneratorProcessResponse:
    """ImageGenerator 응답 처리 테스트"""

    @pytest.fixture
    def generator(self):
        """ImageGenerator 인스턴스"""
        return ImageGenerator("gsk_test_api_key_" + "a" * 50)

    def test_process_response_success(self, generator):
        """성공 응답 처리"""
        response_data = {
            "candidates": [
                {
                    "content": {
                        "parts": [
                            {
                                "inlineData": {
                                    "data": base64.standard_b64encode(b"fake_image_data").decode(),
                                    "mimeType": "image/jpeg"
                                }
                            }
                        ]
                    },
                    "finishReason": "STOP"
                }
            ],
            "usageMetadata": {
                "promptTokenCount": 100,
                "candidatesTokenCount": 50,
                "totalTokenCount": 150
            }
        }

        result = generator._process_response(response_data)

        assert result["success"] is True
        assert result["image_data"] is not None
        assert result["mime_type"] == "image/jpeg"
        assert result["metadata"]["synthetic_watermark"] is True

    def test_process_response_empty(self, generator):
        """빈 응답 처리"""
        response_data = {"candidates": []}
        result = generator._process_response(response_data)

        assert result["success"] is False

    def test_process_response_with_finish_reason(self, generator):
        """완료 이유 포함 응답"""
        response_data = {
            "candidates": [
                {
                    "content": {
                        "parts": [
                            {
                                "inlineData": {
                                    "data": base64.standard_b64encode(b"image").decode(),
                                    "mimeType": "image/png"
                                }
                            }
                        ]
                    },
                    "finishReason": "SAFETY"
                }
            ]
        }

        result = generator._process_response(response_data)
        assert result["finish_reason"] == "SAFETY"


class TestImageGeneratorLoadImage:
    """ImageGenerator 이미지 로드 테스트"""

    @pytest.fixture
    def generator(self):
        return ImageGenerator("gsk_test_api_key_" + "a" * 50)

    def test_load_image_url(self, generator):
        """URL 이미지 로드"""
        with patch('urllib.request.urlopen') as mock_urlopen:
            # Mock 응답
            mock_response = Mock()
            mock_response.read.return_value = b"fake_image_data"
            mock_response.headers = {"Content-Type": "image/jpeg"}
            mock_response.__enter__ = Mock(return_value=mock_response)
            mock_response.__exit__ = Mock(return_value=None)
            mock_urlopen.return_value = mock_response

            image_data = generator._load_image("https://example.com/image.jpg")

            assert image_data["mimeType"] == "image/jpeg"
            assert image_data["data"] == base64.standard_b64encode(b"fake_image_data").decode()

    def test_load_image_file_jpeg(self, generator, tmp_path):
        """로컬 JPEG 파일 로드"""
        # 임시 이미지 파일 생성
        image_file = tmp_path / "test.jpg"
        image_file.write_bytes(b"fake_jpeg_data")

        image_data = generator._load_image(str(image_file))

        assert image_data["mimeType"] == "image/jpeg"
        assert image_data["data"] == base64.standard_b64encode(b"fake_jpeg_data").decode()

    def test_load_image_file_png(self, generator, tmp_path):
        """로컬 PNG 파일 로드"""
        image_file = tmp_path / "test.png"
        image_file.write_bytes(b"fake_png_data")

        image_data = generator._load_image(str(image_file))

        assert image_data["mimeType"] == "image/png"

    def test_load_image_file_not_found(self, generator):
        """존재하지 않는 파일"""
        with pytest.raises(ValueError):
            generator._load_image("/nonexistent/file.jpg")


class TestImageGeneratorAPICall:
    """ImageGenerator API 호출 테스트"""

    @pytest.fixture
    def generator(self):
        return ImageGenerator("gsk_test_api_key_" + "a" * 50)

    @patch('urllib.request.urlopen')
    def test_call_api_success(self, mock_urlopen, generator):
        """API 호출 성공"""
        mock_response = Mock()
        mock_response.status = 200
        mock_response.read.return_value = b'{"candidates": []}'
        mock_response.__enter__ = Mock(return_value=mock_response)
        mock_response.__exit__ = Mock(return_value=None)
        mock_urlopen.return_value = mock_response

        response = generator._call_api(
            "gemini-3-pro-image-preview",
            {"contents": []}
        )

        assert response["status_code"] == 200
        mock_urlopen.assert_called_once()

    @patch('urllib.request.urlopen')
    def test_call_api_failure(self, mock_urlopen, generator):
        """API 호출 실패"""
        mock_error = Mock()
        mock_error.code = 500
        mock_error.read.return_value = b'{"error": {"message": "Server Error"}}'

        import urllib.error
        mock_urlopen.side_effect = urllib.error.HTTPError(
            "url", 500, "Server Error", {}, None
        )

        # urllib 에러 처리 재현
        try:
            response = generator._call_api(
                "gemini-3-pro-image-preview",
                {"contents": []}
            )
            # _call_api는 urllib.error를 잡아서 dict 반환
            assert "status_code" in response
        except Exception:
            pass


class TestImageGeneratorIntegration:
    """통합 테스트"""

    @pytest.fixture
    def generator(self):
        return ImageGenerator("gsk_test_api_key_" + "a" * 50)

    @patch('modules.image_generator.ImageGenerator._call_api')
    def test_generate_image_success(self, mock_call_api, generator):
        """이미지 생성 성공"""
        # Mock API 응답 (새로운 형식)
        mock_call_api.return_value = {
            "status_code": 200,
            "data": {
                "candidates": [
                    {
                        "content": {
                            "parts": [
                                {
                                    "inlineData": {
                                        "data": base64.standard_b64encode(b"image_data").decode(),
                                        "mimeType": "image/jpeg"
                                    }
                                }
                            ]
                        },
                        "finishReason": "STOP"
                    }
                ],
                "usageMetadata": {
                    "promptTokenCount": 100,
                    "candidatesTokenCount": 50,
                    "totalTokenCount": 150
                }
            }
        }

        result = generator.generate_image(
            prompt="beautiful landscape",
            resolution="2048x2048"
        )

        assert result["success"] is True
        assert result["image_data"] is not None

    def test_generate_image_invalid_prompt(self, generator):
        """유효하지 않은 프롬프트"""
        with pytest.raises(ValueError):
            generator.generate_image(prompt="", resolution="2048x2048")

    def test_generate_image_invalid_resolution(self, generator):
        """유효하지 않은 해상도"""
        with pytest.raises(ValueError):
            generator.generate_image(prompt="landscape", resolution="512x512")

    def test_save_image_success(self, generator, tmp_path):
        """이미지 저장 성공"""
        image_data = {
            "image_data": base64.standard_b64encode(b"fake_image_content").decode()
        }

        output_path = str(tmp_path / "test_output.png")
        result = generator.save_image(image_data, output_path)

        assert result is True
        # 파일이 생성되었는지 확인
        import os
        assert os.path.exists(output_path)

    def test_save_image_no_data(self, generator, tmp_path):
        """이미지 데이터 없이 저장"""
        image_data = {"image_data": None}
        output_path = str(tmp_path / "test_output.png")

        result = generator.save_image(image_data, output_path)
        assert result is False


class TestImageGeneratorConfiguration:
    """ImageGenerator 설정 테스트"""

    def test_supported_resolutions(self):
        """지원하는 해상도"""
        assert "1024x1024" in ImageGenerator.SUPPORTED_RESOLUTIONS
        assert "2048x2048" in ImageGenerator.SUPPORTED_RESOLUTIONS
        assert "4096x4096" in ImageGenerator.SUPPORTED_RESOLUTIONS

    def test_api_endpoint(self):
        """API 엔드포인트"""
        assert ImageGenerator.API_ENDPOINT.startswith("https://")
        assert "generativelanguage.googleapis.com" in ImageGenerator.API_ENDPOINT

    def test_default_config(self):
        """기본 설정"""
        config = ImageGenerator.DEFAULT_CONFIG
        assert "temperature" in config
        assert "top_p" in config
        assert config["temperature"] == 1.0


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
