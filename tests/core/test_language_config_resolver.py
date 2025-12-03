"""
Integration tests for the Language Configuration Resolver system.

Tests the complete environment variable-based language configuration
functionality including priority handling, template variable export,
and integration with other MoAI-ADK components.
"""

import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from moai_adk.core.language_config_resolver import (
    LanguageConfigResolver,
    get_resolver,
    resolve_language_config
)


class TestLanguageConfigResolver(unittest.TestCase):
    """Test cases for the Language Configuration Resolver."""

    def setUp(self):
        """Set up test environment with temporary directory."""
        self.test_dir = Path(tempfile.mkdtemp())
        self.moai_dir = self.test_dir / ".moai" / "config"
        self.moai_dir.mkdir(parents=True)
        self.config_file = self.moai_dir / "config.json"

        # Clear all MOAI environment variables to ensure test isolation
        self._env_vars_to_clear = [
            "MOAI_USER_NAME",
            "MOAI_CONVERSATION_LANG",
            "MOAI_AGENT_PROMPT_LANG",
            "MOAI_GIT_COMMIT_LANG",
            "MOAI_CODE_COMMENTS_LANG",
            "MOAI_DOCUMENTATION_LANG",
            "MOAI_ERROR_MESSAGES_LANG",
        ]
        self._saved_env = {}
        for key in self._env_vars_to_clear:
            if key in os.environ:
                self._saved_env[key] = os.environ.pop(key)

        # Reset global resolver instance for each test
        import moai_adk.core.language_config_resolver
        moai_adk.core.language_config_resolver._resolver_instance = None

    def tearDown(self):
        """Clean up test environment."""
        import shutil
        shutil.rmtree(self.test_dir, ignore_errors=True)

        # Restore saved environment variables
        for key, value in self._saved_env.items():
            os.environ[key] = value

    def _write_config(self, config_data):
        """Write configuration data to test config file."""
        self.config_file.write_text(json.dumps(config_data, indent=2))

    def _set_env_vars(self, env_vars):
        """Set environment variables for testing."""
        for key, value in env_vars.items():
            os.environ[key] = value

    def _clear_env_vars(self, env_vars):
        """Clear environment variables after testing."""
        for key in env_vars:
            os.environ.pop(key, None)

    def test_default_configuration(self):
        """Test default configuration when no config file or env vars exist."""
        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()

        self.assertEqual(config['conversation_language'], 'en')
        self.assertEqual(config['conversation_language_name'], 'English')
        self.assertEqual(config['agent_prompt_language'], 'en')
        self.assertEqual(config['user_name'], '')
        self.assertEqual(config['config_source'], 'default')

    def test_config_file_only(self):
        """Test configuration from config file only."""
        config_data = {
            "user": {"name": "TestUser"},
            "language": {
                "conversation_language": "ko",
                "conversation_language_name": "Korean",
                "agent_prompt_language": "ko"
            }
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()

        self.assertEqual(config['conversation_language'], 'ko')
        self.assertEqual(config['conversation_language_name'], 'Korean')
        self.assertEqual(config['agent_prompt_language'], 'ko')
        self.assertEqual(config['user_name'], 'TestUser')
        self.assertEqual(config['config_source'], 'config_file')

    def test_environment_variable_priority(self):
        """Test that environment variables take priority over config file."""
        config_data = {
            "user": {"name": "ConfigUser"},
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English",
                "agent_prompt_language": "en"
            }
        }
        self._write_config(config_data)

        # Use direct os.environ manipulation for better pytest compatibility
        env_vars = {
            'MOAI_USER_NAME': 'EnvUser',
            'MOAI_CONVERSATION_LANG': 'ja',
        }
        self._set_env_vars(env_vars)
        try:
            resolver = LanguageConfigResolver(str(self.test_dir))
            config = resolver.resolve_config()

            # Environment variables should override config file
            self.assertEqual(config['conversation_language'], 'ja')
            # Language name is auto-generated from conversation_language
            self.assertEqual(config['conversation_language_name'], 'Japanese')
            self.assertEqual(config['user_name'], 'EnvUser')
            self.assertEqual(config['config_source'], 'environment')

            # Agent prompt language from config file is preserved (not auto-updated)
            # It only defaults to conversation_language if not explicitly set
            self.assertEqual(config['agent_prompt_language'], 'en')
        finally:
            self._clear_env_vars(env_vars)

    def test_auto_language_name_generation(self):
        """Test automatic language name generation when not provided."""
        config_data = {
            "language": {
                "conversation_language": "ko"
                # conversation_language_name missing
            }
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()

        self.assertEqual(config['conversation_language'], 'ko')
        self.assertEqual(config['conversation_language_name'], 'Korean')

    def test_invalid_language_code_uses_first_two_chars(self):
        """Test that unknown codes use first 2 chars for language lookup."""
        config_data = {
            "language": {
                "conversation_language": "invalid_code"
            }
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()

        # All language codes are accepted but standardized to first 2 chars
        self.assertEqual(config['conversation_language'], 'invalid_code')
        # First 2 chars "in" is not a known language code (Indonesian is "id")
        # So it gets title-cased to "In"
        self.assertEqual(config['conversation_language_name'], 'In')

    def test_truly_unknown_language_code(self):
        """Test that truly unknown 2-char codes get title-cased name."""
        config_data = {
            "language": {
                "conversation_language": "zz"  # Not a known language code
            }
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()

        self.assertEqual(config['conversation_language'], 'zz')
        # Unknown codes get title-cased
        self.assertEqual(config['conversation_language_name'], 'Zz')

    def test_project_owner_fallback(self):
        """Test fallback to project.owner when user.name is not available."""
        config_data = {
            "project": {"owner": "ProjectOwner"},
            "language": {"conversation_language": "ko"}
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()

        self.assertEqual(config['user_name'], 'ProjectOwner')

    def test_template_variables_export(self):
        """Test template variables export functionality."""
        config_data = {
            "user": {"name": "TestUser"},
            "language": {
                "conversation_language": "ko",
                "conversation_language_name": "Korean",
                "agent_prompt_language": "ko"
            }
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()
        template_vars = resolver.export_template_variables(config)

        expected_vars = {
            'CONVERSATION_LANGUAGE': 'ko',
            'CONVERSATION_LANGUAGE_NAME': 'Korean',
            'AGENT_PROMPT_LANGUAGE': 'ko',
            'GIT_COMMIT_MESSAGES_LANGUAGE': 'en',
            'CODE_COMMENTS_LANGUAGE': 'en',
            'DOCUMENTATION_LANGUAGE': 'en',
            'ERROR_MESSAGES_LANGUAGE': 'en',
            'USER_NAME': 'TestUser',
            'PERSONALIZED_GREETING': 'TestUser님',
            'CONFIG_SOURCE': 'config_file'
        }

        self.assertEqual(template_vars, expected_vars)

    def test_personalized_greeting_korean(self):
        """Test Korean personalized greeting generation."""
        config_data = {
            "user": {"name": "홍길동"},
            "language": {"conversation_language": "ko"}
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()
        greeting = resolver.get_personalized_greeting(config)

        self.assertEqual(greeting, "홍길동님")

    def test_personalized_greeting_english(self):
        """Test English personalized greeting generation."""
        config_data = {
            "user": {"name": "JohnDoe"},
            "language": {"conversation_language": "en"}
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()
        greeting = resolver.get_personalized_greeting(config)

        self.assertEqual(greeting, "JohnDoe")

    def test_personalized_greeting_no_user(self):
        """Test personalized greeting with no user name."""
        config_data = {
            "language": {"conversation_language": "en"}
        }
        self._write_config(config_data)

        resolver = LanguageConfigResolver(str(self.test_dir))
        config = resolver.resolve_config()
        greeting = resolver.get_personalized_greeting(config)

        self.assertEqual(greeting, "")

    def test_is_korean_language(self):
        """Test Korean language detection."""
        resolver = LanguageConfigResolver(str(self.test_dir))

        # Test Korean config
        korean_config = {'conversation_language': 'ko'}
        self.assertTrue(resolver.is_korean_language(korean_config))

        # Test English config
        english_config = {'conversation_language': 'en'}
        self.assertFalse(resolver.is_korean_language(english_config))

        # Test default (should be English)
        self.assertFalse(resolver.is_korean_language())

    def test_global_resolver_singleton(self):
        """Test global resolver instance management."""
        resolver1 = get_resolver(str(self.test_dir))
        resolver2 = get_resolver(str(self.test_dir))

        # Should return the same instance
        self.assertIs(resolver1, resolver2)

        # New project root should create new instance
        test_dir2 = Path(tempfile.mkdtemp())
        try:
            resolver3 = get_resolver(str(test_dir2))
            self.assertIsNot(resolver1, resolver3)
        finally:
            import shutil
            shutil.rmtree(test_dir2, ignore_errors=True)

    def test_config_cache_invalidation(self):
        """Test configuration cache invalidation."""
        resolver = LanguageConfigResolver(str(self.test_dir))

        # Initial config
        config1 = resolver.resolve_config()
        self.assertEqual(config1['conversation_language'], 'en')

        # Write new config
        config_data = {
            "language": {"conversation_language": "ja"}
        }
        self._write_config(config_data)

        # Without refresh, should return cached config
        config2 = resolver.resolve_config(force_refresh=False)
        self.assertEqual(config2['conversation_language'], 'en')

        # With refresh, should return new config
        config3 = resolver.resolve_config(force_refresh=True)
        self.assertEqual(config3['conversation_language'], 'ja')

    def test_complete_integration(self):
        """Test complete integration with environment variables and config file."""
        # Write initial config
        config_data = {
            "user": {"name": "ConfigUser"},
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English"
            },
            "project": {"owner": "ProjectOwner"}
        }
        self._write_config(config_data)

        # Set environment variables
        env_vars = {
            'MOAI_CONVERSATION_LANG': 'ko',
            'MOAI_CONVERSATION_LANG_NAME': 'Korean',
            'MOAI_AGENT_PROMPT_LANG': 'ko'
        }
        self._set_env_vars(env_vars)

        try:
            # Test global resolver function
            config = resolve_language_config(str(self.test_dir))

            # Verify environment variable priority
            self.assertEqual(config['conversation_language'], 'ko')
            self.assertEqual(config['conversation_language_name'], 'Korean')
            self.assertEqual(config['agent_prompt_language'], 'ko')

            # Verify user name from env (if set) or config
            if 'MOAI_USER_NAME' in env_vars:
                self.assertEqual(config['user_name'], env_vars['MOAI_USER_NAME'])
            else:
                self.assertEqual(config['user_name'], 'ConfigUser')

            # Verify config source
            self.assertEqual(config['config_source'], 'environment')

        finally:
            self._clear_env_vars(env_vars)

    def test_partial_environment_override(self):
        """Test partial environment variable override."""
        config_data = {
            "user": {"name": "ConfigUser"},
            "language": {
                "conversation_language": "en",
                "conversation_language_name": "English",
                "agent_prompt_language": "en"
            }
        }
        self._write_config(config_data)

        # Use direct os.environ manipulation for better pytest compatibility
        env_vars = {'MOAI_CONVERSATION_LANG': 'ko'}
        self._set_env_vars(env_vars)
        try:
            resolver = LanguageConfigResolver(str(self.test_dir))
            config = resolver.resolve_config()

            # Environment variable should override
            self.assertEqual(config['conversation_language'], 'ko')

            # Language name should be auto-generated based on new language
            self.assertEqual(config['conversation_language_name'], 'Korean')

            # Agent prompt language from config file is preserved (not auto-updated)
            # It only defaults to conversation_language if not explicitly set
            self.assertEqual(config['agent_prompt_language'], 'en')

            # User name should come from config
            self.assertEqual(config['user_name'], 'ConfigUser')
        finally:
            self._clear_env_vars(env_vars)


if __name__ == '__main__':
    unittest.main()