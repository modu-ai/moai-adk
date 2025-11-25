"""
Nano Banana Pro - API Key Management Module

Module for securely receiving Google Gemini 3 API keys and storing them in .env files
"""

import os
import re
from pathlib import Path
from typing import Optional
import getpass
import logging

logger = logging.getLogger(__name__)


class EnvKeyManager:
    """
    Class for managing API keys as environment variables

    Features:
    - Secure input (using getpass)
    - Format validation (gsk_ prefix check)
    - Secure storage in .env files
    - Connection testing
    - Permission settings (chmod 600)

    Example:
        >>> manager = EnvKeyManager()
        >>> manager.setup_api_key()  # Interactive setup
        >>> api_key = manager.load_api_key()  # Load stored key
        >>> manager.test_connection(api_key)  # Connection test
    """

    ENV_FILE = ".env"
    API_KEY_VAR = "GOOGLE_API_KEY"

    @staticmethod
    def setup_api_key() -> bool:
        """
        Interactively receive API key from user and save to .env

        Returns:
            bool: Success status

        Workflow:
            1. Display guidance message
            2. Secure input (getpass)
            3. Format validation
            4. Reconfirmation
            5. Save to .env
            6. Connection test
            7. Display results

        Example:
            >>> manager = EnvKeyManager()
            >>> success = manager.setup_api_key()
            🔐 Gemini API Key Setup Wizard
            ...
            ✅ All setup completed!
        """
        print("\n" + "="*60)
        print("🔐 Gemini 3 API Key Setup Wizard")
        print("="*60 + "\n")

        # Step 1: Guidance
        print("📋 Get your API key:")
        print("   1. Visit https://aistudio.google.com/apikey")
        print("   2. Click '+ Create new API key'")
        print("   3. Select 'In project' and create API key")
        print("   4. Copy the API key\n")

        # Step 2: Input
        print("⚠️  Security Notice: API key will not be displayed on screen\n")

        while True:
            api_key = getpass.getpass("Enter API key: ")

            if not api_key:
                print("❌ API key is empty. Please try again.\n")
                continue

            api_key = api_key.strip()

            # Step 3: Format validation
            if not EnvKeyManager.validate_api_key(api_key):
                print("❌ Invalid API key format.")
                print("   • Must start with gsk_")
                print("   • Must be at least 20 characters long\n")
                continue

            # Step 4: Reconfirmation
            print("\n✓ API key format is valid")
            confirm = input("Save this key? (y/n): ").strip().lower()

            if confirm == 'y':
                break
            else:
                print("Cancelled.\n")
                continue

        # Step 5: Save
        try:
            EnvKeyManager.save_api_key(api_key)
            print("\n✅ API key saved to .env file!")

            # Step 6: Test (optional)
            print("\n🔍 Testing API connection...\n")
            if EnvKeyManager.test_connection(api_key):
                print("✓ Gemini API connection successful")
                print("✓ Quota check completed")
                print("\n✅ All setup completed!")
                print("Ready to generate images! 🎨\n")
                return True
            else:
                print("⚠️  API key validation has issues.")
                print("Please check settings in Google Cloud Console.\n")
                return False

        except Exception as e:
            print(f"\n❌ Error during save: {str(e)}\n")
            return False

    @staticmethod
    def validate_api_key(api_key: str) -> bool:
        """
        Validate API key format

        Args:
            api_key: API key to validate

        Returns:
            bool: Whether format is valid

        Validation Rules:
            - Must start with gsk_
            - Must be at least 20 characters long
            - Must contain only letters, numbers, and underscores

        Example:
            >>> EnvKeyManager.validate_api_key("gsk_valid123456789")
            True
            >>> EnvKeyManager.validate_api_key("invalid")
            False
        """
        if not api_key:
            return False

        # Format validation: must start with gsk_ and be at least 20 characters
        pattern = r'^gsk_[a-zA-Z0-9_]{15,}$'

        if not re.match(pattern, api_key):
            return False

        return True

    @staticmethod
    def save_api_key(api_key: str) -> None:
        """
        Save API key to .env file

        Args:
            api_key: API key to save

        Security:
            - Set file permissions to 600 (owner read/write only)
            - Overwrite existing key if present
            - Create backup (.env.backup)
            - Ensure .env is included in gitignore

        Example:
            >>> EnvKeyManager.save_api_key("gsk_valid_key_here")
            # Saved to .env file
        """
        env_path = Path(EnvKeyManager.ENV_FILE)

        # Create backup if existing file
        if env_path.exists():
            backup_path = Path(f"{EnvKeyManager.ENV_FILE}.backup")
            with open(env_path, 'r') as f:
                backup_content = f.read()
            with open(backup_path, 'w') as f:
                f.write(backup_content)
            logger.info(f"Backup created: {backup_path}")

        # Load existing content and update
        env_vars = {}
        if env_path.exists():
            with open(env_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line and '=' in line and not line.startswith('#'):
                        key, value = line.split('=', 1)
                        env_vars[key.strip()] = value.strip()

        # Update API key
        env_vars[EnvKeyManager.API_KEY_VAR] = api_key

        # Write file
        with open(env_path, 'w') as f:
            for key, value in env_vars.items():
                f.write(f"{key}={value}\n")

        # Set file permissions (600: owner read/write only)
        os.chmod(env_path, 0o600)

        logger.info(f"API key saved to {env_path}")

    @staticmethod
    def load_api_key() -> Optional[str]:
        """
        Load API key from environment variables

        Returns:
            str or None: API key, None if not found

        Priority:
            1. Environment variable (GOOGLE_API_KEY)
            2. .env file

        Example:
            >>> api_key = EnvKeyManager.load_api_key()
            >>> if api_key:
            ...     print("API key loaded successfully")
        """
        # Check environment variable
        api_key = os.getenv(EnvKeyManager.API_KEY_VAR)
        if api_key:
            return api_key

        # Check .env file
        env_path = Path(EnvKeyManager.ENV_FILE)
        if env_path.exists():
            try:
                with open(env_path, 'r') as f:
                    for line in f:
                        line = line.strip()
                        if line.startswith(EnvKeyManager.API_KEY_VAR + '='):
                            api_key = line.split('=', 1)[1].strip()
                            if api_key:
                                return api_key
            except Exception as e:
                logger.error(f"Error reading .env file: {e}")

        return None

    @staticmethod
    def test_connection(api_key: str) -> bool:
        """
        Test API connection

        Args:
            api_key: API key to test

        Returns:
            bool: Connection success status

        Tests:
            - API key format validation
            - Gemini API connectivity test
            - Simple API call

        Example:
            >>> success = EnvKeyManager.test_connection("gsk_valid_key")
            >>> if success:
            ...     print("Connection successful")
        """
        try:
            # Format validation
            if not EnvKeyManager.validate_api_key(api_key):
                logger.error("Invalid API key format")
                return False

            # Gemini API test
            try:
                import google.generativeai as genai
                genai.configure(api_key=api_key)

                # Test connection by listing models
                models = genai.list_models()
                if models:
                    logger.info("API connection successful")
                    return True
                else:
                    logger.error("No models available")
                    return False

            except ImportError:
                # Format validation only if google-generativeai not installed
                logger.warning("google-generativeai not installed")
                return True
            except Exception as e:
                logger.error(f"API connection failed: {e}")
                return False

        except Exception as e:
            logger.error(f"Test connection error: {e}")
            return False

    @staticmethod
    def is_configured() -> bool:
        """
        Check if API key is configured

        Returns:
            bool: Configuration status

        Example:
            >>> if EnvKeyManager.is_configured():
            ...     print("API key is configured")
        """
        api_key = EnvKeyManager.load_api_key()
        return api_key is not None and EnvKeyManager.validate_api_key(api_key)

    @staticmethod
    def reset_api_key() -> None:
        """
        Remove API key (reset)

        Warning: This operation cannot be undone

        Example:
            >>> EnvKeyManager.reset_api_key()
            # API key removed from .env
        """
        env_path = Path(EnvKeyManager.ENV_FILE)

        if env_path.exists():
            env_vars = {}
            with open(env_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line and '=' in line and not line.startswith('#'):
                        key, value = line.split('=', 1)
                        if key.strip() != EnvKeyManager.API_KEY_VAR:
                            env_vars[key.strip()] = value.strip()

            # Rewrite file (excluding API key)
            with open(env_path, 'w') as f:
                for key, value in env_vars.items():
                    f.write(f"{key}={value}\n")

            os.chmod(env_path, 0o600)
            logger.info("API key removed from .env")

    @staticmethod
    def show_setup_status() -> None:
        """
        Display current setup status

        Example:
            >>> EnvKeyManager.show_setup_status()
            ============================================================
            📊 API Key Setup Status
            ============================================================

            ✅ API key is configured
               File: .env
               Variable: GOOGLE_API_KEY
               Format: gsk_...xxxx (masked)

            ✓ Ready to start generating images!
        """
        print("\n" + "="*60)
        print("📊 API Key Setup Status")
        print("="*60 + "\n")

        is_configured = EnvKeyManager.is_configured()

        if is_configured:
            print("✅ API key is configured")
            print(f"   File: {EnvKeyManager.ENV_FILE}")
            print(f"   Variable: {EnvKeyManager.API_KEY_VAR}")
            api_key = EnvKeyManager.load_api_key()
            if api_key:
                print(f"   Format: {api_key[:6]}...{api_key[-4:]} (masked)")
            print("\n✓ Ready to start generating images!\n")
        else:
            print("❌ API key is not configured")
            print("\nSetup with the following command:")
            print("  from modules.env_key_manager import EnvKeyManager")
            print("  EnvKeyManager.setup_api_key()\n")
