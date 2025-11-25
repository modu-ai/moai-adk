#!/usr/bin/env python3
"""
Claude Code Configuration Manager

Manage enterprise configuration across environments:
- Environment-based configuration loading
- Secret management integration
- Configuration validation
- Git strategy management

Usage: python config-manager.py [--validate] [--switch-env ENV] [--git-mode MODE]
"""

import os
import json
import yaml
import argparse
from pathlib import Path
from typing import Dict, List, Optional, Any
from enum import Enum


class GitMode(Enum):
    MANUAL = "manual"
    PERSONAL = "personal"
    TEAM = "team"


class Environment(Enum):
    DEV = "development"
    STAGING = "staging"
    PROD = "production"


class ConfigManager:
    """Manage Claude Code configuration with enterprise patterns."""

    def __init__(self, project_root: str = "."):
        self.project_root = Path(project_root)
        self.moai_dir = self.project_root / ".moai"
        self.config_dir = self.moai_dir / "config"
        self.config_file = self.config_dir / "config.json"

    def load_configuration(self) -> Dict:
        """Load and merge configuration from all sources."""
        # Load base configuration
        base_config = self._load_base_config()

        # Load environment-specific configuration
        env_config = self._load_environment_config(base_config)

        # Load secrets
        secrets = self._load_secrets()

        # Merge configurations with precedence
        merged_config = self._merge_configurations(base_config, env_config, secrets)

        return merged_config

    def _load_base_config(self) -> Dict:
        """Load base configuration file."""
        if not self.config_file.exists():
            return self._get_default_config()

        try:
            with open(self.config_file, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            print(f"Error loading config.json: {e}")
            return self._get_default_config()

    def _load_environment_config(self, base_config: Dict) -> Dict:
        """Load environment-specific configuration."""
        env_name = os.getenv('NODE_ENV', 'development')
        env_config_file = self.config_dir / f"{env_name}.json"

        env_config = {}
        if env_config_file.exists():
            try:
                with open(env_config_file, 'r', encoding='utf-8') as f:
                    env_config = json.load(f)
            except Exception as e:
                print(f"Warning: Could not load {env_name}.json: {e}")

        return env_config

    def _load_secrets(self) -> Dict:
        """Load secrets from various sources."""
        secrets = {}

        # Load from environment variables (with MOAI_ prefix)
        env_secrets = self._load_env_secrets()
        secrets.update(env_secrets)

        # Load from Vault (if available)
        vault_secrets = self._load_vault_secrets()
        if vault_secrets:
            secrets.update(vault_secrets)

        # Load from .env file (if exists)
        env_file_secrets = self._load_env_file_secrets()
        if env_file_secrets:
            secrets.update(env_file_secrets)

        return secrets

    def _load_env_secrets(self) -> Dict:
        """Load secrets from environment variables."""
        secrets = {}
        prefix = "MOAI_"

        for key, value in os.environ.items():
            if key.startswith(prefix):
                # Remove prefix and convert to lowercase with underscores
                secret_key = key[len(prefix):].lower()
                secrets[secret_key] = value

        return secrets

    def _load_vault_secrets(self) -> Optional[Dict]:
        """Load secrets from HashiCorp Vault."""
        try:
            import hvac

            vault_url = os.getenv('VAULT_URL')
            vault_token = os.getenv('VAULT_TOKEN')

            if not vault_url or not vault_token:
                return None

            client = hvac.Client(url=vault_url, token=vault_token)

            # Read secrets from vault
            secret_path = "secret/moai-app"
            response = client.secrets.kv.v2.read_secret_version(path=secret_path)

            if response and 'data' in response and 'data' in response['data']:
                return response['data']['data']

        except ImportError:
            # hvac not installed
            pass
        except Exception as e:
            print(f"Warning: Could not load Vault secrets: {e}")

        return None

    def _load_env_file_secrets(self) -> Optional[Dict]:
        """Load secrets from .env file."""
        env_file = self.project_root / ".env"

        if not env_file.exists():
            return None

        secrets = {}
        try:
            with open(env_file, 'r', encoding='utf-8') as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith('#') and '=' in line:
                        key, value = line.split('=', 1)
                        secrets[key.strip()] = value.strip()
        except Exception as e:
            print(f"Warning: Could not read .env file: {e}")

        return secrets

    def _merge_configurations(self, base: Dict, env: Dict, secrets: Dict) -> Dict:
        """Merge configurations with precedence: base < env < secrets."""
        merged = json.loads(json.dumps(base))  # Deep copy

        # Merge environment config
        self._deep_merge(merged, env)

        # Merge secrets (place in special secrets section)
        if secrets:
            merged['secrets'] = secrets

        return merged

    def _deep_merge(self, base: Dict, update: Dict) -> None:
        """Deep merge update into base dictionary."""
        for key, value in update.items():
            if key in base and isinstance(base[key], dict) and isinstance(value, dict):
                self._deep_merge(base[key], value)
            else:
                base[key] = value

    def _get_default_config(self) -> Dict:
        """Get default configuration."""
        return {
            "user": {
                "name": "",
                "email": ""
            },
            "project": {
                "name": "Untitled Project",
                "description": "",
                "language": "python",
                "locale": "en_US"
            },
            "git_strategy": {
                "mode": "manual",
                "branch_creation": {
                    "prompt_always": True,
                    "auto_enabled": False
                }
            },
            "constitution": {
                "test_coverage_target": 85,
                "enforce_tdd": True
            },
            "language": {
                "conversation_language": "en",
                "agent_prompt_language": "en"
            }
        }

    def validate_configuration(self, config: Dict) -> List[str]:
        """Validate configuration and return list of issues."""
        issues = []

        # Validate required fields
        required_fields = [
            "project.name",
            "git_strategy.mode",
            "constitution.test_coverage_target"
        ]

        for field_path in required_fields:
            if not self._get_nested_value(config, field_path):
                issues.append(f"Missing required field: {field_path}")

        # Validate git strategy
        git_mode = self._get_nested_value(config, "git_strategy.mode")
        if git_mode:
            try:
                GitMode(git_mode)
            except ValueError:
                issues.append(f"Invalid git_strategy.mode: {git_mode}")

        # Validate test coverage target
        coverage_target = self._get_nested_value(config, "constitution.test_coverage_target")
        if coverage_target and (not isinstance(coverage_target, int) or coverage_target < 0 or coverage_target > 100):
            issues.append("constitution.test_coverage_target must be an integer between 0 and 100")

        # Validate language settings
        lang = self._get_nested_value(config, "language.conversation_language")
        if lang and lang not in ["en", "ko", "ja", "zh"]:
            issues.append(f"Unsupported conversation_language: {lang}")

        return issues

    def _get_nested_value(self, data: Dict, path: str) -> Any:
        """Get nested value using dot notation."""
        keys = path.split('.')
        current = data

        for key in keys:
            if isinstance(current, dict) and key in current:
                current = current[key]
            else:
                return None

        return current

    def switch_git_mode(self, mode: GitMode) -> Dict:
        """Switch Git strategy mode."""
        config = self.load_configuration()

        # Update git strategy
        if "git_strategy" not in config:
            config["git_strategy"] = {}

        config["git_strategy"]["mode"] = mode.value

        # Set default values for the mode
        if mode == GitMode.PERSONAL:
            config["git_strategy"]["branch_creation"] = {
                "prompt_always": True,
                "auto_enabled": False
            }
            config["git_strategy"].update({
                "github_integration": True,
                "auto_branch": True,
                "auto_commit": True,
                "auto_push": True
            })
        elif mode == GitMode.TEAM:
            config["git_strategy"]["branch_creation"] = {
                "prompt_always": True,
                "auto_enabled": False
            }
            config["git_strategy"].update({
                "github_integration": True,
                "auto_branch": True,
                "auto_commit": True,
                "auto_push": True,
                "auto_pr": True,
                "draft_pr": True,
                "required_reviews": 1
            })
        elif mode == GitMode.MANUAL:
            config["git_strategy"].update({
                "github_integration": False,
                "auto_branch": False,
                "auto_commit": True,
                "auto_push": False
            })

        # Save configuration
        self.save_configuration(config)

        return {
            "previous_mode": config.get("git_strategy", {}).get("mode"),
            "new_mode": mode.value,
            "config_updated": True
        }

    def save_configuration(self, config: Dict) -> None:
        """Save configuration to file."""
        # Ensure config directory exists
        self.config_dir.mkdir(parents=True, exist_ok=True)

        # Save main config (without secrets)
        config_to_save = config.copy()
        if 'secrets' in config_to_save:
            del config_to_save['secrets']

        with open(self.config_file, 'w', encoding='utf-8') as f:
            json.dump(config_to_save, f, indent=2)

    def create_environment_config(self, environment: Environment, base_config: Optional[Dict] = None) -> None:
        """Create environment-specific configuration file."""
        if base_config is None:
            base_config = self.load_configuration()

        env_config = {
            "environment": environment.value,
            "debug": environment == Environment.DEV,
            "log_level": "DEBUG" if environment == Environment.DEV else "INFO"
        }

        # Add environment-specific overrides
        if environment == Environment.PROD:
            env_config.update({
                "debug": False,
                "performance_monitoring": True,
                "security": {
                    "enforce_https": True,
                    "rate_limiting": True
                }
            })

        # Save environment config
        env_config_file = self.config_dir / f"{environment.value}.json"
        with open(env_config_file, 'w', encoding='utf-8') as f:
            json.dump(env_config, f, indent=2)

        print(f"Created environment config: {env_config_file}")

    def print_configuration(self, config: Dict, show_secrets: bool = False) -> None:
        """Print configuration in a readable format."""
        print("\n" + "="*50)
        print("CLAUDE CODE CONFIGURATION")
        print("="*50)

        # Basic info
        project_name = self._get_nested_value(config, "project.name") or "Unnamed"
        git_mode = self._get_nested_value(config, "git_strategy.mode") or "unknown"
        coverage_target = self._get_nested_value(config, "constitution.test_coverage_target") or 85

        print(f"\n📋 Project: {project_name}")
        print(f"🔄 Git Mode: {git_mode}")
        print(f"🎯 Test Coverage Target: {coverage_target}%")

        # User info
        user_name = self._get_nested_value(config, "user.name")
        if user_name:
            print(f"👤 User: {user_name}")

        # Language
        conv_lang = self._get_nested_value(config, "language.conversation_language") or "en"
        print(f"🌐 Language: {conv_lang}")

        # Git strategy details
        git_config = self._get_nested_value(config, "git_strategy")
        if git_config:
            print(f"\n🔧 Git Strategy:")
            if "branch_creation" in git_config:
                bc = git_config["branch_creation"]
                print(f"  Branch Creation: {'Prompt' if bc.get('prompt_always') else 'Auto'}")
                if bc.get("auto_enabled"):
                    print(f"  Auto-creation: Enabled")

            if "github_integration" in git_config:
                print(f"  GitHub Integration: {'Yes' if git_config['github_integration'] else 'No'}")

        # Secrets (hidden by default)
        if 'secrets' in config and show_secrets:
            print(f"\n🔐 Secrets ({len(config['secrets'])} loaded)")
            for key in config['secrets'].keys():
                print(f"  {key}: [HIDDEN]")
        elif 'secrets' in config:
            print(f"\n🔐 Secrets: {len(config['secrets'])} loaded (use --show-secrets to view)")

    def export_configuration_template(self, output_path: str) -> None:
        """Export a complete configuration template."""
        template = self._get_default_config()

        # Add detailed comments and structure
        full_template = {
            "_template_info": {
                "description": "Claude Code Configuration Template",
                "version": "1.0.0",
                "generated": "2025-11-24"
            },
            "user": {
                "name": "Your Name",
                "email": "your.email@example.com",
                "_description": "User information for personalization"
            },
            "project": {
                "name": "Your Project Name",
                "description": "Brief project description",
                "language": "python",
                "locale": "en_US",
                "_description": "Basic project metadata"
            },
            "git_strategy": {
                "mode": "manual|personal|team",
                "branch_creation": {
                    "prompt_always": True,
                    "auto_enabled": False,
                    "_description": "Control automatic branch creation"
                },
                "_modes": {
                    "manual": "Local Git only, no GitHub integration",
                    "personal": "GitHub personal project with automation",
                    "team": "GitHub team project with code review"
                }
            },
            "constitution": {
                "test_coverage_target": 85,
                "enforce_tdd": True,
                "_description": "Quality gates and development standards"
            },
            "language": {
                "conversation_language": "en|ko|ja|zh",
                "agent_prompt_language": "en|ko|ja|zh",
                "_description": "Language settings for Claude interaction"
            },
            "memory": {
                "working_memory_limit_mb": 50,
                "cache_retention_days": 7,
                "_description": "Memory management settings"
            }
        }

        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)

        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(full_template, f, indent=2)

        print(f"Configuration template exported to: {output_file}")


def main():
    parser = argparse.ArgumentParser(description="Manage Claude Code configuration")
    parser.add_argument("--validate", action="store_true", help="Validate current configuration")
    parser.add_argument("--show", action="store_true", help="Show current configuration")
    parser.add_argument("--show-secrets", action="store_true", help="Show secrets in configuration output")
    parser.add_argument("--git-mode", choices=["manual", "personal", "team"], help="Switch Git strategy mode")
    parser.add_argument("--create-env", choices=["development", "staging", "production"], help="Create environment config")
    parser.add_argument("--export-template", help="Export configuration template to file")
    parser.add_argument("--project-root", default=".", help="Project root directory")

    args = parser.parse_args()

    manager = ConfigManager(args.project_root)

    if args.export_template:
        manager.export_configuration_template(args.export_template)
        return

    config = manager.load_configuration()

    if args.validate:
        issues = manager.validate_configuration(config)
        if issues:
            print("❌ Configuration validation failed:")
            for issue in issues:
                print(f"  • {issue}")
        else:
            print("✅ Configuration validation passed")

    if args.git_mode:
        try:
            mode = GitMode(args.git_mode)
            result = manager.switch_git_mode(mode)
            print(f"✅ Git mode switched from '{result['previous_mode']}' to '{result['new_mode']}'")
        except ValueError as e:
            print(f"❌ Error switching git mode: {e}")

    if args.create_env:
        try:
            env = Environment(args.create_env)
            manager.create_environment_config(env)
            print(f"✅ Created {args.create_env} environment configuration")
        except ValueError as e:
            print(f"❌ Error creating environment config: {e}")

    if args.show or args.validate or args.git_mode or args.create_env:
        manager.print_configuration(config, show_secrets=args.show_secrets)
    else:
        # Default: show current configuration
        manager.print_configuration(config)


if __name__ == "__main__":
    main()