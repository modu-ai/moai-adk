"""
Gemini API Key Management Module

This module manages Gemini API keys by loading them from .env files,
setting new keys when needed, and validating API key formats.

Usage:
    from modules.env_key_manager import EnvKeyManager

    # Load API key
    api_key = EnvKeyManager.get_api_key()

    # Set API key
    EnvKeyManager.set_api_key("gsk_...")

    # Validate API key
    is_valid = EnvKeyManager.validate_api_key("gsk_...")
"""

import os
from pathlib import Path
from typing import Optional


class EnvKeyManager:
    """Gemini API Key Management Class"""

    ENV_FILE = ".env"
    ENV_KEY_NAME = "GEMINI_API_KEY"
    API_KEY_PREFIX = "gsk_"

    @classmethod
    def get_api_key(cls) -> Optional[str]:
        """
        Load Gemini API key from .env file.

        Returns:
            str: API key string
            None: If key is not configured
        """
        # Check environment variable first
        if env_key := os.environ.get(cls.ENV_KEY_NAME):
            return env_key

        # Check .env file
        if Path(cls.ENV_FILE).exists():
            with open(cls.ENV_FILE, 'r') as f:
                for line in f:
                    if line.startswith(f"{cls.ENV_KEY_NAME}="):
                        return line.split('=', 1)[1].strip()

        return None

    @classmethod
    def set_api_key(cls, api_key: str) -> bool:
        """
        Set Gemini API key in .env file.

        Args:
            api_key: API key to set

        Returns:
            bool: Success status

        Raises:
            ValueError: Invalid API key format
        """
        if not cls.validate_api_key(api_key):
            raise ValueError(
                f"Invalid API key format. Must start with '{cls.API_KEY_PREFIX}'"
            )

        # Read existing .env file
        env_content = ""
        key_found = False

        if Path(cls.ENV_FILE).exists():
            with open(cls.ENV_FILE, 'r') as f:
                for line in f:
                    if line.startswith(f"{cls.ENV_KEY_NAME}="):
                        env_content += f"{cls.ENV_KEY_NAME}={api_key}\n"
                        key_found = True
                    else:
                        env_content += line

        # Add new key if not found
        if not key_found:
            env_content += f"{cls.ENV_KEY_NAME}={api_key}\n"

        # Write to .env file
        with open(cls.ENV_FILE, 'w') as f:
            f.write(env_content)

        # Set environment variable
        os.environ[cls.ENV_KEY_NAME] = api_key

        return True

    @classmethod
    def validate_api_key(cls, api_key: str) -> bool:
        """
        Validate API key format.

        Args:
            api_key: API key to validate

        Returns:
            bool: Validation status
        """
        if not api_key:
            return False

        # Basic format validation (starts with gsk_ and min 50 chars)
        return (
            api_key.startswith(cls.API_KEY_PREFIX) and
            len(api_key) >= 50
        )

    @classmethod
    def is_configured(cls) -> bool:
        """
        Check if API key is configured.

        Returns:
            bool: Configuration status
        """
        api_key = cls.get_api_key()
        return cls.validate_api_key(api_key) if api_key else False


if __name__ == "__main__":
    # Test execution
    print(f"API Key Configured: {EnvKeyManager.is_configured()}")
    api_key = EnvKeyManager.get_api_key()
    if api_key:
        print(f"API Key (masked): {api_key[:10]}...{api_key[-10:]}")
    else:
        print("API Key not configured")
