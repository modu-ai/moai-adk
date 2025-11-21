# Advanced Patterns - Claude Code Configuration

## Pattern 1: Multi-Layer Configuration Merging

Merge configurations from multiple sources with priority resolution.

```python
from typing import Dict, Any, Optional
from dataclasses import dataclass
from enum import Enum

class ConfigurationPriority(Enum):
    """Configuration source priority levels."""
    ENVIRONMENT = 1
    COMMAND_LINE = 2
    USER = 3
    PROJECT = 4
    DEFAULT = 5

@dataclass
class ConfigSource:
    """Configuration source with metadata."""
    source_type: str
    priority: ConfigurationPriority
    data: Dict[str, Any]
    path: Optional[str] = None
    timestamp: Optional[float] = None

class MultiLayerConfigMerger:
    """Merge configurations from multiple sources."""

    def __init__(self):
        self.sources: List[ConfigSource] = []

    def add_source(
        self,
        source_type: str,
        data: Dict[str, Any],
        priority: ConfigurationPriority,
        path: Optional[str] = None
    ):
        """Add configuration source."""
        self.sources.append(ConfigSource(
            source_type=source_type,
            priority=priority,
            data=data,
            path=path,
            timestamp=time.time()
        ))

    def merge_all(self) -> Dict[str, Any]:
        """Merge all sources by priority."""

        # Sort by priority (lower number = higher priority)
        sorted_sources = sorted(
            self.sources,
            key=lambda x: x.priority.value
        )

        merged = {}

        for source in sorted_sources:
            merged = self._deep_merge(merged, source.data)

        return merged

    def _deep_merge(
        self,
        base: Dict[str, Any],
        override: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Deep merge configuration dictionaries."""

        result = base.copy()

        for key, value in override.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._deep_merge(result[key], value)
            else:
                result[key] = value

        return result

    def get_merged_value(
        self,
        key_path: str,
        default: Any = None
    ) -> Any:
        """Get merged value by dot-notation path."""

        merged = self.merge_all()
        keys = key_path.split('.')

        current = merged
        for key in keys:
            if isinstance(current, dict) and key in current:
                current = current[key]
            else:
                return default

        return current

# Usage
merger = MultiLayerConfigMerger()

# Add sources in order
merger.add_source(
    'defaults',
    {'debug': False, 'port': 8000, 'workers': 4},
    ConfigurationPriority.DEFAULT
)

merger.add_source(
    'project',
    {'debug': False, 'port': 3000},
    ConfigurationPriority.PROJECT
)

merger.add_source(
    'environment',
    {'debug': True},
    ConfigurationPriority.ENVIRONMENT
)

# Merged configuration
config = merger.merge_all()
# Result: {debug: True, port: 3000, workers: 4}
```

## Pattern 2: Configuration Profiles with Environment-Specific Overrides

Support multiple configuration profiles (dev, staging, prod) with cascading overrides.

```python
from abc import ABC, abstractmethod

class ConfigProfile(ABC):
    """Base configuration profile."""

    @abstractmethod
    def load_base_config(self) -> Dict[str, Any]:
        """Load base configuration."""
        pass

    @abstractmethod
    def load_environment_overrides(self) -> Dict[str, Any]:
        """Load environment-specific overrides."""
        pass

class DevelopmentProfile(ConfigProfile):
    """Development configuration profile."""

    def load_base_config(self) -> Dict[str, Any]:
        return {
            'debug': True,
            'log_level': 'DEBUG',
            'database': 'sqlite:///:memory:',
            'cache_enabled': False
        }

    def load_environment_overrides(self) -> Dict[str, Any]:
        return {
            'hot_reload': True,
            'api_timeout': 30
        }

class ProductionProfile(ConfigProfile):
    """Production configuration profile."""

    def load_base_config(self) -> Dict[str, Any]:
        return {
            'debug': False,
            'log_level': 'WARNING',
            'database': os.getenv('DATABASE_URL'),
            'cache_enabled': True
        }

    def load_environment_overrides(self) -> Dict[str, Any]:
        return {
            'workers': int(os.getenv('WORKERS', 4)),
            'api_timeout': int(os.getenv('API_TIMEOUT', 60))
        }

class ProfileManager:
    """Manage configuration profiles."""

    PROFILES = {
        'development': DevelopmentProfile,
        'staging': StagingProfile,
        'production': ProductionProfile
    }

    @staticmethod
    def load_profile(profile_name: str) -> Dict[str, Any]:
        """Load and merge profile configuration."""

        if profile_name not in ProfileManager.PROFILES:
            raise ValueError(f"Unknown profile: {profile_name}")

        profile_class = ProfileManager.PROFILES[profile_name]
        profile = profile_class()

        base = profile.load_base_config()
        overrides = profile.load_environment_overrides()

        return {**base, **overrides}

# Usage
env = os.getenv('ENV', 'development')
config = ProfileManager.load_profile(env)
```

## Pattern 3: Configuration Validation with Schemas

Validate configuration against JSON schemas or Pydantic models.

```python
from pydantic import BaseModel, Field, validator
from typing import Optional, List

class DatabaseConfig(BaseModel):
    """Database configuration."""
    host: str = Field(..., description="Database host")
    port: int = Field(default=5432, ge=1, le=65535)
    database: str = Field(..., min_length=1)
    user: str = Field(...)
    password: str = Field(...)

    @validator('host')
    def validate_host(cls, v):
        if not v.replace('.', '').replace('-', '').isalnum():
            raise ValueError('Invalid host format')
        return v

class ServerConfig(BaseModel):
    """Server configuration."""
    host: str = Field(default='0.0.0.0')
    port: int = Field(default=8000, ge=1, le=65535)
    workers: int = Field(default=4, ge=1)
    debug: bool = Field(default=False)

class AppConfig(BaseModel):
    """Complete application configuration."""
    server: ServerConfig
    database: DatabaseConfig
    cache_ttl: int = Field(default=300, ge=0)
    max_connections: int = Field(default=100, ge=1)

class ConfigValidator:
    """Validate configuration against schema."""

    def __init__(self, schema_class: type[BaseModel]):
        self.schema = schema_class

    def validate(self, config_dict: Dict[str, Any]) -> Tuple[bool, Optional[str]]:
        """Validate configuration."""
        try:
            self.schema(**config_dict)
            return True, None
        except Exception as e:
            return False, str(e)

    def coerce(self, config_dict: Dict[str, Any]) -> Dict[str, Any]:
        """Coerce configuration to valid schema."""
        validated = self.schema(**config_dict)
        return validated.dict()

# Usage
validator = ConfigValidator(AppConfig)

raw_config = {
    'server': {'port': 3000},
    'database': {
        'host': 'localhost',
        'database': 'myapp',
        'user': 'admin',
        'password': 'secret'
    }
}

is_valid, error = validator.validate(raw_config)
if is_valid:
    coerced = validator.coerce(raw_config)
```

## Pattern 4: Dynamic Configuration with Hot Reload

Reload configuration without restart.

```python
import watchfiles
import asyncio

class DynamicConfigManager:
    """Dynamically reload configuration on file changes."""

    def __init__(self, config_path: str):
        self.config_path = config_path
        self.current_config: Dict[str, Any] = {}
        self.callbacks: List[Callable] = []
        self.file_watcher = None

    def register_callback(self, callback: Callable):
        """Register callback for config changes."""
        self.callbacks.append(callback)

    async def watch_and_reload(self):
        """Watch config file and reload on changes."""

        async for changes in watchfiles.awatch(self.config_path):
            print(f"Configuration changed: {changes}")

            # Load new configuration
            new_config = self._load_config()

            # Detect changes
            diff = self._compute_diff(self.current_config, new_config)

            # Update configuration
            self.current_config = new_config

            # Notify callbacks
            for callback in self.callbacks:
                await callback(self.current_config, diff)

    def _load_config(self) -> Dict[str, Any]:
        """Load configuration from file."""
        with open(self.config_path) as f:
            if self.config_path.endswith('.json'):
                return json.load(f)
            elif self.config_path.endswith('.yaml'):
                return yaml.safe_load(f)

    def _compute_diff(
        self,
        old: Dict,
        new: Dict
    ) -> Dict[str, tuple]:
        """Compute configuration diff."""

        diff = {}

        for key in set(list(old.keys()) + list(new.keys())):
            if key not in old:
                diff[key] = (None, new[key])
            elif key not in new:
                diff[key] = (old[key], None)
            elif old[key] != new[key]:
                diff[key] = (old[key], new[key])

        return diff

# Usage
config_manager = DynamicConfigManager('config.yaml')

async def on_config_change(config: Dict, diff: Dict):
    print(f"Config changed: {diff}")
    # React to changes
    if 'log_level' in diff:
        set_log_level(config['log_level'])

config_manager.register_callback(on_config_change)

# Start watching in background
asyncio.create_task(config_manager.watch_and_reload())
```

## Pattern 5: Configuration Encryption and Secrets Management

Safely store and manage sensitive configuration values.

```python
from cryptography.fernet import Fernet

class SecretsManager:
    """Manage encrypted secrets in configuration."""

    def __init__(self, encryption_key: Optional[str] = None):
        if encryption_key is None:
            encryption_key = os.getenv('CONFIG_ENCRYPTION_KEY')

        self.cipher = Fernet(encryption_key.encode() if encryption_key else Fernet.generate_key())

    def encrypt_value(self, value: str) -> str:
        """Encrypt configuration value."""
        return self.cipher.encrypt(value.encode()).decode()

    def decrypt_value(self, encrypted: str) -> str:
        """Decrypt configuration value."""
        return self.cipher.decrypt(encrypted.encode()).decode()

    def encrypt_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Encrypt sensitive values in config."""

        encrypted = config.copy()
        sensitive_keys = {'password', 'token', 'secret', 'api_key'}

        for key, value in config.items():
            if key.lower() in sensitive_keys and isinstance(value, str):
                encrypted[key] = self.encrypt_value(value)

        return encrypted

    def decrypt_config(self, config: Dict[str, Any]) -> Dict[str, Any]:
        """Decrypt sensitive values in config."""

        decrypted = config.copy()

        for key, value in config.items():
            if isinstance(value, str) and value.startswith('gAAAAAA'):
                try:
                    decrypted[key] = self.decrypt_value(value)
                except Exception:
                    pass

        return decrypted

# Usage
secrets_manager = SecretsManager()

config = {
    'database_password': 'secret123',
    'api_token': 'token_xyz'
}

# Encrypt for storage
encrypted_config = secrets_manager.encrypt_config(config)
# Store encrypted_config in file

# Decrypt for use
decrypted_config = secrets_manager.decrypt_config(encrypted_config)
```

## Pattern 6: Configuration Environment Variable Expansion

Expand environment variables in configuration values.

```python
import re
from typing import Optional

class EnvironmentExpander:
    """Expand environment variables in configuration."""

    PATTERN = re.compile(r'\$\{([^}]+)\}|\$(\w+)')

    def expand_config(
        self,
        config: Dict[str, Any],
        strict: bool = False
    ) -> Dict[str, Any]:
        """Expand environment variables in configuration."""

        expanded = {}

        for key, value in config.items():
            if isinstance(value, str):
                expanded[key] = self.expand_string(value, strict)
            elif isinstance(value, dict):
                expanded[key] = self.expand_config(value, strict)
            elif isinstance(value, list):
                expanded[key] = [
                    self.expand_string(item, strict) if isinstance(item, str) else item
                    for item in value
                ]
            else:
                expanded[key] = value

        return expanded

    def expand_string(self, value: str, strict: bool = False) -> str:
        """Expand environment variables in string."""

        def replace_var(match):
            var_name = match.group(1) or match.group(2)
            env_value = os.getenv(var_name)

            if env_value is None:
                if strict:
                    raise ValueError(f"Environment variable not set: {var_name}")
                return match.group(0)

            return env_value

        return self.PATTERN.sub(replace_var, value)

# Usage
config = {
    'database_url': '${DATABASE_URL}',
    'api_key': '$API_KEY',
    'debug': False
}

expander = EnvironmentExpander()
expanded = expander.expand_config(config)
```

## Pattern 7: Configuration Versioning and Migration

Support configuration format upgrades with migration scripts.

```python
from abc import ABC, abstractmethod

class ConfigMigration(ABC):
    """Base class for configuration migrations."""

    version_from: int
    version_to: int

    @abstractmethod
    def migrate(self, old_config: Dict[str, Any]) -> Dict[str, Any]:
        """Migrate configuration from old to new format."""
        pass

class Migration_1_to_2(ConfigMigration):
    """Migrate config v1 to v2."""

    version_from = 1
    version_to = 2

    def migrate(self, old_config: Dict[str, Any]) -> Dict[str, Any]:
        """Rename 'server_port' to 'server.port'."""

        migrated = old_config.copy()

        if 'server_port' in migrated:
            migrated['server'] = {'port': migrated.pop('server_port')}

        migrated['version'] = 2

        return migrated

class ConfigMigrationManager:
    """Manage configuration migrations."""

    def __init__(self):
        self.migrations: Dict[Tuple[int, int], ConfigMigration] = {}

    def register_migration(self, migration: ConfigMigration):
        """Register configuration migration."""
        key = (migration.version_from, migration.version_to)
        self.migrations[key] = migration

    def migrate_to_latest(
        self,
        config: Dict[str, Any],
        target_version: int
    ) -> Dict[str, Any]:
        """Migrate configuration to target version."""

        current_version = config.get('version', 1)

        while current_version < target_version:
            migration_key = (current_version, current_version + 1)

            if migration_key not in self.migrations:
                raise ValueError(
                    f"No migration from v{current_version} to v{current_version + 1}"
                )

            migration = self.migrations[migration_key]
            config = migration.migrate(config)
            current_version = migration.version_to

        return config

# Usage
manager = ConfigMigrationManager()
manager.register_migration(Migration_1_to_2())

old_config = {'version': 1, 'server_port': 8000}
new_config = manager.migrate_to_latest(old_config, target_version=2)
```

---

**Advanced Patterns Summary**: 7 enterprise patterns for multi-layer configuration merging, profiles, validation, hot reload, secrets management, environment expansion, and versioning.

