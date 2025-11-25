# BaaS Foundation - Practical Examples

## Example 1: Evaluating BaaS Services - Decision Matrix

```javascript
// Framework for evaluating BaaS providers
const evaluateProvider = (provider, requirements) => {
    const criteria = {
        scalability: {
            weight: 0.2,
            evaluate: (p) => p.autoScaling ? 100 : 0
        },
        cost: {
            weight: 0.15,
            evaluate: (p) => p.payAsYouGo ? 100 : 50
        },
        development_speed: {
            weight: 0.2,
            evaluate: (p) => p.sdkQuality * 100
        },
        security: {
            weight: 0.25,
            evaluate: (p) => p.soc2 && p.encryption ? 100 : 0
        },
        ecosystem: {
            weight: 0.2,
            evaluate: (p) => p.communitySize + p.integrations
        }
    };

    let score = 0;
    for (const [criterion, config] of Object.entries(criteria)) {
        score += config.evaluate(provider) * config.weight;
    }
    return score;
};
```

## Example 2: Implementing Multi-Provider Strategy

```python
# Multi-provider authentication abstraction
from abc import ABC, abstractmethod
from enum import Enum

class AuthProvider(Enum):
    AUTH0 = "auth0"
    CLERK = "clerk"
    FIREBASE = "firebase"

class BaaSAuth(ABC):
    @abstractmethod
    async def authenticate(self, credentials):
        pass

    @abstractmethod
    async def create_user(self, email, password):
        pass

    @abstractmethod
    async def validate_token(self, token):
        pass

class Auth0Adapter(BaaSAuth):
    def __init__(self, domain, client_id, client_secret):
        self.domain = domain
        self.client_id = client_id
        self.client_secret = client_secret

    async def authenticate(self, credentials):
        # Auth0-specific implementation
        pass

class ClerkAdapter(BaaSAuth):
    def __init__(self, api_key):
        self.api_key = api_key

    async def authenticate(self, credentials):
        # Clerk-specific implementation
        pass

# Factory pattern for provider selection
class AuthFactory:
    @staticmethod
    def create_auth(provider: AuthProvider, **config):
        if provider == AuthProvider.AUTH0:
            return Auth0Adapter(**config)
        elif provider == AuthProvider.CLERK:
            return ClerkAdapter(**config)
        else:
            raise ValueError(f"Unknown provider: {provider}")
```

## Example 3: Database Service Selection Pattern

```typescript
// TypeScript interface for database service selection
interface DatabaseProvider {
    name: string;
    type: 'relational' | 'document' | 'realtime';
    scalability: 'horizontal' | 'vertical';
    realtimeCapabilities: boolean;
    pricing: 'fixed' | 'variable' | 'hybrid';
}

const providers: Record<string, DatabaseProvider> = {
    supabase: {
        name: 'Supabase',
        type: 'relational',
        scalability: 'horizontal',
        realtimeCapabilities: true,
        pricing: 'variable'
    },
    firebase: {
        name: 'Firebase Firestore',
        type: 'document',
        scalability: 'horizontal',
        realtimeCapabilities: true,
        pricing: 'variable'
    },
    neon: {
        name: 'Neon PostgreSQL',
        type: 'relational',
        scalability: 'horizontal',
        realtimeCapabilities: false,
        pricing: 'variable'
    }
};

// Selection logic based on requirements
const selectDatabase = (requirements: {
    dataModel: string;
    scaleExpectation: string;
    realtimeNeeded: boolean;
}) => {
    if (requirements.realtimeNeeded && requirements.dataModel === 'document') {
        return providers.firebase;
    }
    if (requirements.dataModel === 'relational') {
        return providers.supabase;
    }
    return providers.neon;
};
```

## Example 4: Deployment Platform Strategy

```bash
#!/bin/bash
# Deployment strategy selection based on application type

select_deployment_platform() {
    local app_type=$1
    local scale_needs=$2

    case "$app_type" in
        nextjs_app)
            if [[ "$scale_needs" == "serverless" ]]; then
                echo "vercel"  # Best for serverless Next.js
            else
                echo "railway"  # Good for full-stack
            fi
            ;;
        static_site)
            echo "vercel"  # Optimal for static + edge functions
            ;;
        full_stack)
            echo "railway"  # Complete environment management
            ;;
        edge_first)
            echo "cloudflare"  # Global edge network
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Usage
platform=$(select_deployment_platform "nextjs_app" "serverless")
echo "Selected platform: $platform"
```

## Example 5: Cost Optimization Framework

```python
# BaaS Cost Calculator and Optimizer
class BaaSCostCalculator:
    def __init__(self):
        self.monthly_costs = {}

    def calculate_auth_costs(self, provider, monthly_active_users):
        """Calculate authentication service costs"""
        costs = {
            'auth0': max(monthly_active_users * 0.001, 50),  # Min $50
            'clerk': max(monthly_active_users * 0.0015, 25),  # Min $25
            'firebase': monthly_active_users * 0.0005,  # Pay-per-use
        }
        return costs.get(provider, 0)

    def calculate_database_costs(self, provider, gb_stored, monthly_reads):
        """Calculate database service costs"""
        costs = {
            'supabase': (gb_stored * 0.25) + (monthly_reads * 0.000001),
            'firebase': (gb_stored * 0.18) + (monthly_reads * 0.000006),
            'neon': (gb_stored * 0.3) + (monthly_reads * 0.0000005),
        }
        return costs.get(provider, 0)

    def recommend_provider(self, workload_profile):
        """Recommend most cost-effective provider"""
        cheapest = None
        lowest_cost = float('inf')

        for provider in ['auth0', 'clerk', 'firebase']:
            cost = self.calculate_auth_costs(
                provider,
                workload_profile['monthly_users']
            )
            if cost < lowest_cost:
                lowest_cost = cost
                cheapest = provider

        return cheapest, lowest_cost
```

## Example 6: Migration Between BaaS Providers

```python
# BaaS Provider Migration Template
class BaasMigration:
    def __init__(self, source_provider, target_provider):
        self.source = source_provider
        self.target = target_provider
        self.migration_log = []

    async def migrate_users(self, user_batch_size=100):
        """Migrate users from source to target provider"""
        users = await self.source.get_all_users()

        for i in range(0, len(users), user_batch_size):
            batch = users[i:i + user_batch_size]

            for user in batch:
                try:
                    await self.target.create_user(
                        email=user['email'],
                        metadata=user.get('metadata', {})
                    )
                    self.migration_log.append({
                        'user_id': user['id'],
                        'status': 'success'
                    })
                except Exception as e:
                    self.migration_log.append({
                        'user_id': user['id'],
                        'status': 'failed',
                        'error': str(e)
                    })

    async def verify_migration(self):
        """Verify successful migration"""
        source_count = await self.source.count_users()
        target_count = await self.target.count_users()

        success_rate = (
            sum(1 for log in self.migration_log if log['status'] == 'success')
            / len(self.migration_log)
        )

        return {
            'source_count': source_count,
            'target_count': target_count,
            'success_rate': success_rate,
            'issues': [log for log in self.migration_log if log['status'] == 'failed']
        }
```

## Example 7: Multi-Tenant Architecture Pattern

```typescript
// Multi-tenant BaaS configuration
interface TenantConfig {
    tenantId: string;
    authProvider: 'auth0' | 'clerk';
    databaseProvider: 'supabase' | 'firebase';
    deploymentPlatform: 'vercel' | 'railway';
}

class MultiTenantBaaS {
    private tenants: Map<string, TenantConfig> = new Map();

    registerTenant(config: TenantConfig) {
        this.tenants.set(config.tenantId, config);
    }

    getTenantConfig(tenantId: string): TenantConfig {
        const config = this.tenants.get(tenantId);
        if (!config) {
            throw new Error(`Tenant ${tenantId} not found`);
        }
        return config;
    }

    async authenticateForTenant(
        tenantId: string,
        credentials: any
    ) {
        const config = this.getTenantConfig(tenantId);
        const authService = this.getAuthService(config.authProvider);
        return authService.authenticate(credentials);
    }

    private getAuthService(provider: string) {
        // Provider-specific authentication service
        // Implementation details...
        return null;
    }
}
```

## Example 8: Real-Time Data Synchronization

```javascript
// Real-time data sync between BaaS and local cache
class RealtimeBaaSSyncManager {
    constructor(baasProvider) {
        this.baasProvider = baasProvider;
        this.localCache = new Map();
        this.syncQueue = [];
    }

    startRealtimeSync(collectionName) {
        this.baasProvider.onCollectionChange(
            collectionName,
            (changes) => {
                this.handleRealtimeChanges(collectionName, changes);
            }
        );
    }

    handleRealtimeChanges(collectionName, changes) {
        changes.forEach(change => {
            if (change.type === 'add') {
                this.localCache.set(change.doc.id, change.doc.data());
            } else if (change.type === 'modify') {
                const existing = this.localCache.get(change.doc.id);
                this.localCache.set(change.doc.id, {
                    ...existing,
                    ...change.doc.data()
                });
            } else if (change.type === 'remove') {
                this.localCache.delete(change.doc.id);
            }
        });
    }

    queueOfflineChange(doc) {
        this.syncQueue.push({
            timestamp: Date.now(),
            doc: doc
        });
    }

    async syncOfflineChanges() {
        while (this.syncQueue.length > 0) {
            const change = this.syncQueue.shift();
            try {
                await this.baasProvider.update(change.doc);
            } catch (error) {
                this.syncQueue.unshift(change);
                throw error;
            }
        }
    }
}
```

## Example 9: Serverless Function Integration

```python
# Serverless functions across multiple BaaS platforms
from typing import Callable, Any

class ServerlessFunctionManager:
    def __init__(self, deployment_provider: str):
        self.deployment_provider = deployment_provider
        self.functions = {}

    def register_function(
        self,
        name: str,
        handler: Callable,
        runtime: str = 'node18',
        memory: int = 512,
        timeout: int = 30
    ):
        """Register a serverless function"""
        self.functions[name] = {
            'handler': handler,
            'runtime': runtime,
            'memory': memory,
            'timeout': timeout,
            'platform': self.deployment_provider
        }

    async def deploy_function(self, name: str):
        """Deploy function to target platform"""
        if name not in self.functions:
            raise ValueError(f"Function {name} not registered")

        func_config = self.functions[name]

        if self.deployment_provider == 'vercel':
            return await self._deploy_to_vercel(func_config)
        elif self.deployment_provider == 'cloudflare':
            return await self._deploy_to_cloudflare(func_config)
        else:
            raise ValueError(f"Unknown platform: {self.deployment_provider}")

    async def _deploy_to_vercel(self, config):
        # Vercel-specific deployment logic
        pass

    async def _deploy_to_cloudflare(self, config):
        # Cloudflare Workers deployment logic
        pass
```

## Example 10: Error Handling and Fallback Strategy

```typescript
// Resilient BaaS error handling with fallbacks
class ResilientBaaSClient {
    constructor(
        primaryProvider: string,
        fallbackProvider: string
    ) {
        this.primaryProvider = primaryProvider;
        this.fallbackProvider = fallbackProvider;
        this.failureCount = 0;
        this.failureThreshold = 5;
    }

    async executeWithFallback<T>(
        operation: () => Promise<T>,
        fallbackOperation: () => Promise<T>
    ): Promise<T> {
        try {
            const result = await operation();
            this.failureCount = 0;  // Reset on success
            return result;
        } catch (error) {
            this.failureCount++;

            if (this.failureCount >= this.failureThreshold) {
                console.log('Switching to fallback provider');
                return fallbackOperation();
            }

            throw error;
        }
    }

    async getUser(userId: string): Promise<any> {
        return this.executeWithFallback(
            () => this.primaryProvider.getUser(userId),
            () => this.fallbackProvider.getUser(userId)
        );
    }
}
```

## Example 11: Configuration Management

```yaml
# BaaS Configuration File Template
services:
  auth:
    provider: auth0
    config:
      domain: ${AUTH0_DOMAIN}
      clientId: ${AUTH0_CLIENT_ID}
      clientSecret: ${AUTH0_CLIENT_SECRET}
    features:
      - social_login
      - mfa
      - passwordless

  database:
    primary:
      provider: supabase
      config:
        projectUrl: ${SUPABASE_URL}
        apiKey: ${SUPABASE_KEY}
      features:
        - realtime
        - auth
        - storage

    fallback:
      provider: firebase
      config:
        projectId: ${FIREBASE_PROJECT_ID}
        apiKey: ${FIREBASE_API_KEY}

  deployment:
    platform: vercel
    config:
      projectId: ${VERCEL_PROJECT_ID}
      token: ${VERCEL_TOKEN}
    features:
      - edge_functions
      - preview_deployments
```

## Example 12: Security Best Practices

```python
# BaaS Security Implementation
import secrets
from cryptography.fernet import Fernet

class BaaSSecurityManager:
    def __init__(self):
        self.encryption_key = Fernet.generate_key()
        self.cipher = Fernet(self.encryption_key)

    def encrypt_sensitive_data(self, data: str) -> str:
        """Encrypt sensitive data before storing"""
        return self.cipher.encrypt(data.encode()).decode()

    def decrypt_sensitive_data(self, encrypted_data: str) -> str:
        """Decrypt sensitive data"""
        return self.cipher.decrypt(encrypted_data.encode()).decode()

    def generate_secure_token(self, length: int = 32) -> str:
        """Generate cryptographically secure token"""
        return secrets.token_urlsafe(length)

    def validate_csrf_token(self, token: str, session_token: str) -> bool:
        """Validate CSRF protection token"""
        return secrets.compare_digest(token, session_token)

    async def validate_api_key(self, api_key: str) -> bool:
        """Validate API key against BaaS provider"""
        # Implementation for provider-specific validation
        pass
```

## Example 13: Monitoring and Observability

```javascript
// BaaS Monitoring and Metrics Collection
class BaaSMonitor {
    constructor(metricsProvider) {
        this.metricsProvider = metricsProvider;
        this.metrics = {
            apiCalls: 0,
            errors: 0,
            latency: []
        };
    }

    recordMetric(metricName, value, tags = {}) {
        this.metricsProvider.gauge(
            `baas.${metricName}`,
            value,
            { tags }
        );
    }

    async trackApiCall(provider, endpoint) {
        const startTime = Date.now();

        try {
            const result = await endpoint();
            const duration = Date.now() - startTime;

            this.recordMetric('api_call_duration', duration, {
                provider: provider,
                status: 'success'
            });

            return result;
        } catch (error) {
            const duration = Date.now() - startTime;

            this.recordMetric('api_call_duration', duration, {
                provider: provider,
                status: 'error'
            });

            throw error;
        }
    }

    getHealthReport() {
        return {
            total_calls: this.metrics.apiCalls,
            error_rate: this.metrics.errors / this.metrics.apiCalls,
            avg_latency: this.metrics.latency.reduce((a, b) => a + b) / this.metrics.latency.length
        };
    }
}
```

## Example 14: Testing BaaS Integrations

```python
# Unit testing BaaS integrations
import pytest
from unittest.mock import Mock, AsyncMock, patch

@pytest.fixture
def baas_client():
    return Mock()

@pytest.mark.asyncio
async def test_user_creation(baas_client):
    """Test user creation with mocked BaaS provider"""
    baas_client.create_user = AsyncMock(
        return_value={'id': '123', 'email': 'test@example.com'}
    )

    result = await baas_client.create_user(
        email='test@example.com',
        password='secure_password'
    )

    assert result['id'] == '123'
    assert result['email'] == 'test@example.com'
    baas_client.create_user.assert_called_once()

@pytest.mark.asyncio
async def test_provider_fallback():
    """Test fallback when primary provider fails"""
    primary = AsyncMock(side_effect=Exception("Provider down"))
    fallback = AsyncMock(return_value={'status': 'ok'})

    # Test logic for fallback behavior
    with pytest.raises(Exception):
        await primary()

    result = await fallback()
    assert result['status'] == 'ok'
```

## Example 15: Documentation and Integration Guide

```markdown
# BaaS Integration Checklist

## Phase 1: Planning and Evaluation
- [ ] Define application requirements
- [ ] Evaluate available BaaS providers
- [ ] Compare costs and features
- [ ] Make provider selection

## Phase 2: Setup and Configuration
- [ ] Create accounts with selected providers
- [ ] Configure authentication providers
- [ ] Set up database and storage
- [ ] Configure deployment platform

## Phase 3: Development Integration
- [ ] Install SDKs and client libraries
- [ ] Implement authentication flow
- [ ] Set up data models
- [ ] Create API endpoints

## Phase 4: Testing and Validation
- [ ] Unit test integrations
- [ ] Integration test full workflows
- [ ] Test error scenarios
- [ ] Validate security measures

## Phase 5: Deployment
- [ ] Deploy to staging environment
- [ ] Run smoke tests
- [ ] Deploy to production
- [ ] Monitor and observe
```

---

**Context7 Integration**: Use `/auth0/auth0-js`, `/clerk/clerk-js`, `/firebase/firebase-js` for latest SDK documentation.

**Version**: 1.0.0 | **Last Updated**: 2025-11-22
