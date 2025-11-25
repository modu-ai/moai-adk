# Advanced Configuration Patterns

Enterprise-grade configuration strategies for feature management, dynamic reloading, and real-time monitoring.

## Feature Flag Management

### Pattern 1: Environment-Specific Feature Flags

```typescript
interface FeatureFlags {
  [featureName: string]: {
    development?: boolean;
    staging?: boolean;
    production?: boolean;
  } | boolean;
}

class FeatureFlagManager {
  private flags: FeatureFlags;
  private environment: string;

  constructor(flags: FeatureFlags, environment: string) {
    this.flags = flags;
    this.environment = environment;
  }

  isEnabled(featureName: string): boolean {
    const flag = this.flags[featureName];

    // Simple boolean flag
    if (typeof flag === 'boolean') {
      return flag;
    }

    // Environment-specific flag
    if (typeof flag === 'object' && flag !== null) {
      const envFlag = flag[this.environment as keyof typeof flag];
      return envFlag ?? false;
    }

    return false;
  }

  getAllEnabled(): string[] {
    return Object.entries(this.flags)
      .filter(([_, flag]) => this.isEnabled(_))
      .map(([name]) => name);
  }
}
```

### Pattern 2: Gradual Rollout with Percentage-Based Flags

```python
import hashlib

class PercentageBasedRollout:
    def __init__(self, config: dict):
        self.feature_rollouts = config.get('feature_rollouts', {})

    def should_enable_for_user(self, feature_name: str, user_id: str) -> bool:
        if feature_name not in self.feature_rollouts:
            return False

        rollout = self.feature_rollouts[feature_name]
        percentage = rollout.get('percentage', 0)

        # Deterministic hash-based selection
        hash_input = f"{feature_name}:{user_id}".encode()
        hash_value = int(hashlib.md5(hash_input).hexdigest(), 16)

        # Map hash to 0-100 percentage
        user_percentage = (hash_value % 100) + 1

        return user_percentage <= percentage
```

## Configuration Caching Strategies

### Pattern 3: TTL-Based Cache with Invalidation

```typescript
interface CacheEntry<T> {
  value: T;
  timestamp: number;
  ttl: number;
}

class ConfigurationCache {
  private cache = new Map<string, CacheEntry<any>>();
  private readonly defaultTTL = 3600000; // 1 hour

  set<T>(key: string, value: T, ttl?: number): void {
    this.cache.set(key, {
      value,
      timestamp: Date.now(),
      ttl: ttl ?? this.defaultTTL
    });
  }

  get<T>(key: string): T | null {
    const entry = this.cache.get(key);

    if (!entry) {
      return null;
    }

    // Check if expired
    const elapsed = Date.now() - entry.timestamp;
    if (elapsed > entry.ttl) {
      this.cache.delete(key);
      return null;
    }

    return entry.value as T;
  }

  invalidate(key: string): void {
    this.cache.delete(key);
  }

  invalidatePattern(pattern: RegExp): void {
    const keysToDelete = Array.from(this.cache.keys()).filter(key =>
      pattern.test(key)
    );
    keysToDelete.forEach(key => this.cache.delete(key));
  }
}
```

## Dynamic Configuration Reloading

### Pattern 4: Hot Reload Without Restart

```python
import threading
import time
import json

class HotReloadableConfiguration:
    def __init__(self, config_path: str):
        self.config_path = config_path
        self._config = self._load_config()
        self._version = self._get_file_version()
        self._listeners = []
        self._running = False

    def _load_config(self) -> dict:
        with open(self.config_path, 'r') as f:
            return json.load(f)

    def get(self, key: str, default=None):
        keys = key.split('.')
        value = self._config

        for k in keys:
            if isinstance(value, dict):
                value = value.get(k)
            else:
                return default

        return value if value is not None else default

    def subscribe(self, callback) -> None:
        self._listeners.append(callback)

    def start_watching(self, check_interval: int = 5) -> None:
        self._running = True
        self._reload_thread = threading.Thread(
            target=self._watch_for_changes,
            args=(check_interval,),
            daemon=True
        )
        self._reload_thread.start()

    def _get_file_version(self) -> str:
        import os
        return str(os.path.getmtime(self.config_path))
```

## Configuration Monitoring & Alerting

### Pattern 5: Change Detection & Alerts

```python
from enum import Enum
from typing import List, Callable

class AlertSeverity(Enum):
    INFO = "info"
    WARNING = "warning"
    CRITICAL = "critical"

class ConfigurationMonitor:
    def __init__(self):
        self._change_handlers = []
        self._alert_handlers = []

    def on_change(self, handler: Callable) -> None:
        self._change_handlers.append(handler)

    async def detect_changes(self, old_config: dict, new_config: dict) -> None:
        changed_fields = self._get_changed_fields(old_config, new_config)

        if not changed_fields:
            return

        # Notify change handlers
        for handler in self._change_handlers:
            await handler(changed_fields)

        # Determine alert severity
        alert = self._determine_alert(changed_fields, new_config)

    def _get_changed_fields(self, old: dict, new: dict, prefix: str = '') -> List[str]:
        changed = []

        for key in set(list(old.keys()) + list(new.keys())):
            path = f"{prefix}.{key}" if prefix else key

            if key not in old:
                changed.append(path)
            elif key not in new:
                changed.append(path)
            elif isinstance(old[key], dict) and isinstance(new[key], dict):
                changed.extend(
                    self._get_changed_fields(old[key], new[key], path)
                )
            elif old[key] != new[key]:
                changed.append(path)

        return changed

    def _determine_alert(self, changed_fields: List[str], config: dict):
        critical_fields = ['database.password', 'auth.jwt_secret', 'api.key']

        for field in changed_fields:
            if any(field.startswith(cf) for cf in critical_fields):
                return AlertSeverity.CRITICAL

        return AlertSeverity.INFO
```

## Configuration Versioning & Rollback

### Pattern 6: Version History Management

```python
import hashlib
import json
from datetime import datetime

class ConfigurationVersionManager:
    def __init__(self, max_versions: int = 10):
        self.versions = []
        self.max_versions = max_versions
        self.current_version_index = -1

    def save_version(self, config: dict, description: str = '') -> str:
        version_id = f"v{len(self.versions) + 1}"

        version = {
            'id': version_id,
            'config': config,
            'description': description,
            'timestamp': datetime.utcnow().isoformat(),
            'checksum': hashlib.sha256(
                json.dumps(config, sort_keys=True).encode()
            ).hexdigest()
        }

        self.versions.append(version)
        self.current_version_index = len(self.versions) - 1

        if len(self.versions) > self.max_versions:
            self.versions.pop(0)

        return version_id

    def get_version(self, version_id: str) -> dict:
        for version in self.versions:
            if version['id'] == version_id:
                return version['config']

        raise ValueError(f"Version {version_id} not found")

    def list_versions(self):
        return [
            {
                'id': v['id'],
                'description': v['description'],
                'timestamp': v['timestamp'],
                'checksum': v['checksum']
            }
            for v in self.versions
        ]

    def rollback_to(self, version_id: str) -> dict:
        config = self.get_version(version_id)

        for i, v in enumerate(self.versions):
            if v['id'] == version_id:
                self.current_version_index = i
                break

        return config
```

## Best Practices

### ✅ DO

- Implement TTL-based caching for external configurations
- Use percentage-based rollouts for new features
- Validate all configuration changes before applying
- Maintain configuration change history for auditing
- Use feature flags for gradual rollout
- Monitor critical configuration changes
- Implement automatic rollback on validation failure
- Use environment-specific feature flags

### ❌ DON'T

- Store feature flags in code
- Change production configuration without validation
- Lose configuration history
- Reload configuration without monitoring
- Use same feature flags across environments
- Skip validation for dynamic reloading
- Apply configuration changes without alerts
- Hard-code feature decision logic

---

**Tools**: Flags frameworks (Unleash, LaunchDarkly), Configuration as Code tools
