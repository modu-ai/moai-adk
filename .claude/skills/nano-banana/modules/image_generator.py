"""
Nano Banana Pro Image Generation Module

Generate and edit images using Gemini 3 Pro Image Preview API.

API Reference: https://ai.google.dev/api/rest/v1beta/models/generateContent

Usage:
    from modules.image_generator import ImageGenerator
    from modules.env_key_manager import EnvKeyManager

    # Set up API key
    api_key = EnvKeyManager.get_api_key()

    # Text-to-Image generation
    image_url = ImageGenerator.generate_image(
        prompt="beautiful mountain landscape",
        resolution="2048x2048"
    )

    # Image-to-Image editing
    edited_url = ImageGenerator.edit_image(
        image_url="https://...",
        instruction="change sky to sunset colors",
        resolution="2048x2048"
    )
"""

import base64
import json
import time
import urllib.request
import urllib.parse
from pathlib import Path
from typing import Optional, Dict, Any

from .error_handler import ErrorHandler


class ImageGenerator:
    """Nano Banana Pro Image Generation Class"""

    # Gemini API configuration
    API_ENDPOINT = "https://generativelanguage.googleapis.com/v1beta/models"
    TEXT_TO_IMAGE_MODEL = "gemini-3-pro-image-preview"
    IMAGE_TO_IMAGE_MODEL = "gemini-3-pro-image-preview"

    # Resolution options
    SUPPORTED_RESOLUTIONS = {
        "1024x1024": "1k",
        "2048x2048": "2k",
        "4096x4096": "4k",
    }

    # Generation configuration
    DEFAULT_CONFIG = {
        "temperature": 1.0,  # Creativity level (0.0-2.0)
        "top_p": 0.95,
        "top_k": 64,
        "max_output_tokens": 1000,
    }

    def __init__(self, api_key: str):
        """
        Initialize ImageGenerator.

        Args:
            api_key: Gemini API key
        """
        self.api_key = api_key

    def generate_image(
        self,
        prompt: str,
        resolution: str = "2048x2048",
        max_retries: int = 3,
    ) -> Dict[str, Any]:
        """
        Generate image from text prompt.

        Args:
            prompt: Image generation prompt
            resolution: Resolution (1024x1024, 2048x2048, 4096x4096)
            max_retries: Maximum retry attempts

        Returns:
            dict: Generation result {image_data, mime_type, metadata}

        Raises:
            ValueError: Invalid input
            Exception: API error
        """
        if not self._validate_prompt(prompt):
            raise ValueError(f"Invalid prompt: {prompt}")

        if not self._validate_resolution(resolution):
            raise ValueError(
                f"Invalid resolution: {resolution}. "
                f"Supported: {list(self.SUPPORTED_RESOLUTIONS.keys())}"
            )

        # Configure API request
        request_body = {
            "contents": [
                {
                    "role": "user",
                    "parts": [
                        {"text": prompt}
                    ]
                }
            ],
            "generationConfig": {
                **self.DEFAULT_CONFIG,
                "imageResolution": resolution,
            }
        }

        # API call with retry logic
        for attempt in range(max_retries):
            try:
                response = self._call_api(
                    self.TEXT_TO_IMAGE_MODEL,
                    request_body
                )

                # 응답 처리
                if response["status_code"] == 200:
                    return self._process_response(response["data"])
                else:
                    error_data = response["data"]
                    error_handler = ErrorHandler(error_data)

                    if error_handler.is_retryable() and attempt < max_retries - 1:
                        wait_time = error_handler.get_retry_delay()
                        time.sleep(wait_time)
                        continue
                    else:
                        raise Exception(
                            f"API Error: {error_handler.get_message()}"
                        )

            except Exception as e:
                if attempt < max_retries - 1:
                    time.sleep(2 ** attempt)  # Exponential backoff
                    continue
                else:
                    raise Exception(f"Request failed: {str(e)}")

    def edit_image(
        self,
        image_input: str,
        instruction: str,
        resolution: str = "2048x2048",
        max_retries: int = 3,
    ) -> Dict[str, Any]:
        """
        Edit image using image-to-image transformation.

        Args:
            image_input: Image file path or URL
            instruction: Editing instruction
            resolution: Resolution (1024x1024, 2048x2048, 4096x4096)
            max_retries: Maximum retry attempts

        Returns:
            dict: Edited image result

        Raises:
            ValueError: Invalid input
            Exception: API error
        """
        if not self._validate_resolution(resolution):
            raise ValueError(f"Invalid resolution: {resolution}")

        # Load image from file or URL
        image_data = self._load_image(image_input)

        # Configure API request
        request_body = {
            "contents": [
                {
                    "role": "user",
                    "parts": [
                        {"inlineData": image_data},
                        {"text": instruction}
                    ]
                }
            ],
            "generationConfig": {
                **self.DEFAULT_CONFIG,
                "imageResolution": resolution,
            }
        }

        # Call API with retry logic
        for attempt in range(max_retries):
            try:
                response = self._call_api(
                    self.IMAGE_TO_IMAGE_MODEL,
                    request_body
                )

                if response["status_code"] == 200:
                    return self._process_response(response["data"])
                else:
                    error_data = response["data"]
                    error_handler = ErrorHandler(error_data)

                    if error_handler.is_retryable() and attempt < max_retries - 1:
                        wait_time = error_handler.get_retry_delay()
                        time.sleep(wait_time)
                        continue
                    else:
                        raise Exception(
                            f"API Error: {error_handler.get_message()}"
                        )

            except Exception as e:
                if attempt < max_retries - 1:
                    time.sleep(2 ** attempt)
                    continue
                else:
                    raise Exception(f"Request failed: {str(e)}")

    def _call_api(self, model: str, request_body: Dict) -> Dict[str, Any]:
        """
        Call Gemini API.

        Args:
            model: Model name
            request_body: Request body

        Returns:
            Dict: API response data with status code
        """
        url = f"{self.API_ENDPOINT}/{model}:generateContent"
        params = {"key": self.api_key}
        query_string = urllib.parse.urlencode(params)
        final_url = f"{url}?{query_string}"

        request_json = json.dumps(request_body).encode('utf-8')
        request_obj = urllib.request.Request(
            final_url,
            data=request_json,
            headers={"Content-Type": "application/json"}
        )

        try:
            with urllib.request.urlopen(request_obj, timeout=60) as response:
                status_code = response.status
                response_data = json.loads(response.read().decode('utf-8'))
                return {"status_code": status_code, "data": response_data}
        except urllib.error.HTTPError as e:
            status_code = e.code
            try:
                response_data = json.loads(e.read().decode('utf-8'))
            except Exception:
                response_data = {"error": {"message": str(e)}}
            return {"status_code": status_code, "data": response_data}

    def _process_response(self, response_data: Dict) -> Dict[str, Any]:
        """
        Process API response.

        Args:
            response_data: API response data

        Returns:
            dict: Processed result
        """
        result = {
            "success": False,
            "image_url": None,
            "metadata": {},
        }

        # Parse response structure
        if "candidates" in response_data and response_data["candidates"]:
            candidate = response_data["candidates"][0]

            # Extract image data
            if "content" in candidate and "parts" in candidate["content"]:
                for part in candidate["content"]["parts"]:
                    if "inlineData" in part:
                        # Base64-encoded image data
                        inline_data = part["inlineData"]
                        result["image_data"] = inline_data.get("data")
                        result["mime_type"] = inline_data.get("mimeType")
                        result["success"] = True

            # Extract metadata
            if "finishReason" in candidate:
                result["finish_reason"] = candidate["finishReason"]

        # SynthID watermark information
        if "usageMetadata" in response_data:
            result["metadata"] = {
                "input_tokens": response_data["usageMetadata"].get("promptTokenCount"),
                "output_tokens": response_data["usageMetadata"].get(
                    "candidatesTokenCount"
                ),
                "total_tokens": response_data["usageMetadata"].get("totalTokenCount"),
                "synthetic_watermark": True,  # Nano Banana automatically applies SynthID
            }

        return result

    def _load_image(self, image_input: str) -> Dict:
        """
        Load image from file or URL.

        Args:
            image_input: Image file path or URL

        Returns:
            dict: Base64-encoded image data

        Raises:
            ValueError: Unsupported format
        """
        if image_input.startswith("http"):
            # Download from URL
            try:
                with urllib.request.urlopen(image_input, timeout=10) as response:
                    image_bytes = response.read()
                    mime_type = response.headers.get(
                        "Content-Type",
                        "image/jpeg"
                    )
            except urllib.error.URLError as e:
                raise ValueError(f"Failed to download image from URL: {str(e)}")
        else:
            # Load local file
            path = Path(image_input)
            if not path.exists():
                raise ValueError(f"File not found: {image_input}")

            with open(path, "rb") as f:
                image_bytes = f.read()

            # Infer MIME type from file extension
            suffix = path.suffix.lower()
            mime_map = {
                ".jpg": "image/jpeg",
                ".jpeg": "image/jpeg",
                ".png": "image/png",
                ".gif": "image/gif",
                ".webp": "image/webp",
            }
            mime_type = mime_map.get(suffix, "image/jpeg")

        return {
            "mimeType": mime_type,
            "data": base64.standard_b64encode(image_bytes).decode(),
        }

    def _validate_prompt(self, prompt: str) -> bool:
        """Validate prompt format."""
        return bool(prompt and isinstance(prompt, str) and len(prompt) >= 3)

    def _validate_resolution(self, resolution: str) -> bool:
        """Validate resolution format."""
        return resolution in self.SUPPORTED_RESOLUTIONS

    def save_image(self, image_data: Dict, output_path: str) -> bool:
        """
        Save generated image to file.

        Args:
            image_data: Generated image data
            output_path: Output file path

        Returns:
            bool: Whether save was successful
        """
        if not image_data.get("image_data"):
            return False

        try:
            image_bytes = base64.standard_b64decode(image_data["image_data"])
            output_file = Path(output_path)
            output_file.parent.mkdir(parents=True, exist_ok=True)

            with open(output_file, "wb") as f:
                f.write(image_bytes)

            return True
        except Exception:
            return False


if __name__ == "__main__":
    # Test execution
    from .env_key_manager import EnvKeyManager

    api_key = EnvKeyManager.get_api_key()
    if not api_key:
        print("Error: GEMINI_API_KEY not configured")
        exit(1)

    generator = ImageGenerator(api_key)

    try:
        # Text-to-Image generation
        result = generator.generate_image(
            "beautiful sunset over mountain",
            resolution="2048x2048"
        )
        print(f"Generation result: {result}")
    except Exception as e:
        print(f"Error: {e}")
