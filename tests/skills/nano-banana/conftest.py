"""
Pytest configuration and shared fixtures for Nano Banana Pro tests.

This module provides common fixtures for:
- API mocking (google.genai client)
- Environment variable management
- Temporary file handling
"""

import os
import sys
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

# Add scripts directory to path for imports
# Calculate project root from tests/skills/nano-banana/
project_root = Path(__file__).parent.parent.parent.parent
scripts_dir = (
    project_root
    / "src"
    / "moai_adk"
    / "templates"
    / ".claude"
    / "skills"
    / "moai-ai-nano-banana"
    / "scripts"
)
sys.path.insert(0, str(scripts_dir))


@pytest.fixture
def mock_api_key():
    """Provide a mock API key for testing."""
    return "test_api_key_12345"


@pytest.fixture
def env_with_api_key(mock_api_key):
    """Set up environment with API key."""
    with patch.dict(os.environ, {"GOOGLE_API_KEY": mock_api_key}):
        yield mock_api_key


@pytest.fixture
def env_without_api_key():
    """Set up environment without API key."""
    with patch.dict(os.environ, {}, clear=True):
        # Explicitly remove GOOGLE_API_KEY if it exists
        env = os.environ.copy()
        env.pop("GOOGLE_API_KEY", None)
        with patch.dict(os.environ, env, clear=True):
            yield


@pytest.fixture
def temp_output_dir(tmp_path):
    """Provide a temporary output directory for generated images."""
    output_dir = tmp_path / "output"
    output_dir.mkdir(exist_ok=True)
    return output_dir


@pytest.fixture
def mock_genai_client():
    """Create a mock google.genai client."""
    mock_client = MagicMock()
    return mock_client


@pytest.fixture
def mock_successful_response():
    """Create a mock successful API response with image data."""
    mock_response = MagicMock()
    mock_part = MagicMock()
    mock_part.inline_data = MagicMock()
    mock_part.inline_data.data = b"\x89PNG\r\n\x1a\n" + b"\x00" * 100  # Fake PNG header
    mock_part.text = None

    mock_content = MagicMock()
    mock_content.parts = [mock_part]

    mock_candidate = MagicMock()
    mock_candidate.content = mock_content

    mock_response.candidates = [mock_candidate]
    return mock_response


@pytest.fixture
def mock_text_only_response():
    """Create a mock response with text only (no image)."""
    mock_response = MagicMock()
    mock_part = MagicMock()
    mock_part.inline_data = None
    mock_part.text = "Unable to generate image for the given prompt."

    mock_content = MagicMock()
    mock_content.parts = [mock_part]

    mock_candidate = MagicMock()
    mock_candidate.content = mock_content

    mock_response.candidates = [mock_candidate]
    return mock_response


@pytest.fixture
def mock_rate_limit_error():
    """Create a mock rate limit exception."""
    return Exception("429 Resource exhausted: Too many requests")


@pytest.fixture
def mock_api_key_error():
    """Create a mock API key exception."""
    return Exception("Permission denied: Invalid API key")


@pytest.fixture
def sample_prompts():
    """Provide sample prompts for testing."""
    return [
        "A fluffy cat eating a banana",
        "Modern dashboard UI design",
        "Mountain landscape at sunset"
    ]


@pytest.fixture
def sample_config():
    """Provide a sample batch configuration dictionary."""
    return {
        "defaults": {
            "style": "photorealistic",
            "resolution": "2K",
            "aspect_ratio": "16:9"
        },
        "images": [
            "Simple prompt string",
            {
                "prompt": "Detailed prompt with options",
                "filename": "custom_image.png",
                "resolution": "4K",
                "style": "watercolor"
            },
            {
                "prompt": "Another image",
                "aspect_ratio": "1:1"
            }
        ]
    }
