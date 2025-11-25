# Auth0 - Advanced Patterns

## Enterprise Pattern 1: Custom Database Connection with Migration

```python
# Auth0 custom database with user migration
class Auth0DBMigration:
    def __init__(self, auth0_domain, auth0_mgmt_token, legacy_db):
        self.domain = auth0_domain
        self.mgmt_token = auth0_mgmt_token
        self.legacy_db = legacy_db

    async def lazy_migrate_user(self, username, password):
        """Lazy migration: Migrate user on first Auth0 login"""
        # Check if user exists in Auth0
        auth0_user = await self.get_user_from_auth0(username)

        if auth0_user:
            return auth0_user

        # User doesn't exist in Auth0, check legacy DB
        legacy_user = await self.legacy_db.find_user(username)

        if not legacy_user:
            return None

        # Verify password against legacy system
        if not self.verify_legacy_password(password, legacy_user):
            return None

        # Migrate user to Auth0
        auth0_user = await self.create_user_in_auth0({
            'email': legacy_user['email'],
            'username': username,
            'user_metadata': {
                'legacy_id': legacy_user['id'],
                'migrated_from': 'legacy_db',
                'migrated_at': datetime.now().isoformat()
            }
        })

        return auth0_user

    async def bulk_migrate_users(self, batch_size=100):
        """Bulk migration of users from legacy system"""
        offset = 0
        migrated_count = 0
        failed_count = 0

        while True:
            users = await self.legacy_db.get_users(
                offset=offset,
                limit=batch_size
            )

            if not users:
                break

            for user in users:
                try:
                    await self.create_user_in_auth0({
                        'email': user['email'],
                        'email_verified': True,
                        'user_metadata': {
                            'legacy_id': user['id']
                        }
                    })
                    migrated_count += 1
                except Exception as e:
                    failed_count += 1
                    print(f"Failed to migrate user {user['id']}: {e}")

            offset += batch_size

        return {
            'migrated': migrated_count,
            'failed': failed_count,
            'total': migrated_count + failed_count
        }
```

## Enterprise Pattern 2: Multi-Tenant Authorization

```typescript
// Auth0 multi-tenant authorization with organizations
interface TenantContext {
    tenantId: string;
    userId: string;
    roles: string[];
    permissions: string[];
}

class Auth0MultiTenantAuth {
    async extractTenantContext(token: string): Promise<TenantContext> {
        const decoded = jwt_decode(token);

        // Extract tenant from token organization claim
        const org = decoded['org_id'];
        const userId = decoded['sub'];

        // Get tenant-specific roles and permissions
        const roles = await this.getOrgRoles(org, userId);
        const permissions = await this.getOrgPermissions(org, roles);

        return {
            tenantId: org,
            userId,
            roles,
            permissions
        };
    }

    async getOrgRoles(orgId: string, userId: string): Promise<string[]> {
        // Fetch user roles within organization
        const response = await fetch(
            `https://${this.domain}/api/v2/organizations/${orgId}/members/${userId}/roles`,
            {
                headers: { 'Authorization': `Bearer ${this.mgmt_token}` }
            }
        );
        const roles = await response.json();
        return roles.map(r => r.name);
    }

    async getOrgPermissions(orgId: string, roles: string[]): Promise<string[]> {
        const permissions = new Set<string>();

        for (const role of roles) {
            const perms = await this.getRolePermissions(orgId, role);
            perms.forEach(p => permissions.add(p));
        }

        return Array.from(permissions);
    }

    middlewareCheckPermission(requiredPermission: string) {
        return async (req: any, res: any, next: any) => {
            const token = req.headers.authorization?.split(' ')[1];
            const context = await this.extractTenantContext(token);

            if (!context.permissions.includes(requiredPermission)) {
                return res.status(403).json({ error: 'Insufficient permissions' });
            }

            req.tenantContext = context;
            next();
        };
    }
}
```

## Pattern 1: Guardian MFA Implementation

```python
# Auth0 Guardian MFA setup and enforcement
class Auth0Guardian:
    def __init__(self, auth0_domain, mgmt_token):
        self.domain = auth0_domain
        self.mgmt_token = mgmt_token

    async def enable_mfa(self, user_id: str, factors: list):
        """Enable MFA for user with specified factors"""
        response = await self.update_user_mfa({
            'multifactor': factors,  # ['sms', 'totp', 'guardian']
            'user_id': user_id
        })
        return response

    async def get_enrollment_tickets(self, user_id: str):
        """Get Guardian enrollment tickets for user"""
        response = await fetch(
            f'https://{self.domain}/api/v2/users/{user_id}/guardian-enrollments',
            headers={'Authorization': f'Bearer {self.mgmt_token}'}
        )
        return response.json()

    async def verify_mfa(self, user_id: str, code: str, factor: str):
        """Verify MFA code"""
        response = await fetch(
            f'https://{self.domain}/api/v2/users/{user_id}/guardian/enrollments/verify',
            method='POST',
            json={
                'code': code,
                'factor_name': factor
            },
            headers={'Authorization': f'Bearer {self.mgmt_token}'}
        )
        return response.json()
```

## Pattern 2: Risk-Based Authentication

```javascript
// Auth0 Risk-Based Authentication (RBA)
class RiskBasedAuth {
    async evaluateRisk(context) {
        const riskFactors = {
            unknownLocation: await this.checkNewLocation(context),
            unusualTime: this.checkUnusualLoginTime(context),
            impossibleTravel: await this.checkImpossibleTravel(context),
            newDevice: await this.checkNewDevice(context),
            suspiciousActivity: await this.checkSuspiciousActivity(context)
        };

        const riskScore = this.calculateRiskScore(riskFactors);

        return {
            riskLevel: this.getRiskLevel(riskScore),
            riskScore,
            factors: riskFactors,
            action: this.determineAction(riskScore)
        };
    }

    determineAction(riskScore) {
        if (riskScore >= 80) {
            return 'deny';  // Block authentication
        } else if (riskScore >= 50) {
            return 'mfa_required';  // Require MFA
        } else if (riskScore >= 20) {
            return 'prompt_verification';  // Ask for verification
        }
        return 'allow';  // Allow authentication
    }

    async applyRiskAction(action, context) {
        switch (action) {
            case 'deny':
                throw new SecurityError('Authentication denied due to high risk');
            case 'mfa_required':
                return await this.requireMFA(context.user_id);
            case 'prompt_verification':
                return await this.sendVerificationEmail(context.email);
            default:
                return true;
        }
    }
}
```

## Pattern 3: Custom Rules for Business Logic

```javascript
// Auth0 Rules for complex business logic
module.exports = async function(user, context, callback) {
    // Add custom claims
    const customClaims = {
        'http://example.com/app_metadata': user.app_metadata,
        'http://example.com/user_segment': await getUserSegment(user.user_id)
    };

    Object.assign(context.idToken, customClaims);
    Object.assign(context.accessToken, customClaims);

    // Check account status
    if (user.app_metadata?.account_status === 'suspended') {
        return callback(new UnauthorizedError('Account suspended'));
    }

    // Track login for analytics
    await trackLoginEvent({
        user_id: user.user_id,
        email: user.email,
        timestamp: new Date(),
        ip: context.request.ip,
        user_agent: context.request.user_agent
    });

    // Add device information
    context.idToken['http://example.com/device'] = {
        user_agent: context.request.user_agent,
        ip_address: context.request.ip
    };

    callback(null, user, context);
};

async function getUserSegment(userId) {
    // Segment user based on analytics
    const events = await getRecentUserEvents(userId);
    if (events.length > 100) return 'power_user';
    if (events.length > 10) return 'active_user';
    return 'casual_user';
}
```

## Pattern 4: Session Management

```python
# Auth0 session management and timeout
class Auth0SessionManager:
    def __init__(self, auth0_domain):
        self.domain = auth0_domain
        self.sessions = {}
        self.session_timeout = 3600  # 1 hour

    async def create_session(self, user_id: str, token: str):
        """Create user session"""
        session_id = secrets.token_urlsafe(32)
        self.sessions[session_id] = {
            'user_id': user_id,
            'token': token,
            'created_at': datetime.now(),
            'last_activity': datetime.now()
        }
        return session_id

    async def validate_session(self, session_id: str) -> bool:
        """Validate active session"""
        if session_id not in self.sessions:
            return False

        session = self.sessions[session_id]
        age = (datetime.now() - session['last_activity']).total_seconds()

        if age > self.session_timeout:
            del self.sessions[session_id]
            return False

        # Update last activity
        session['last_activity'] = datetime.now()
        return True

    async def revoke_session(self, session_id: str):
        """Revoke session"""
        if session_id in self.sessions:
            del self.sessions[session_id]
```

## Pattern 5: Audit Logging

```typescript
// Auth0 audit logging for compliance
class Auth0AuditLog {
    async logAuthEvent(event: {
        type: string;
        userId: string;
        action: string;
        result: 'success' | 'failure';
        timestamp: Date;
        metadata?: any;
    }) {
        const auditEntry = {
            id: generateId(),
            ...event,
            timestamp: event.timestamp.toISOString(),
            loggedAt: new Date().toISOString()
        };

        // Store in audit database
        await this.auditDatabase.insert(auditEntry);

        // Send to compliance system
        await this.complianceSystem.report(auditEntry);
    }

    async queryAuditLog(filters: {
        userId?: string;
        type?: string;
        startDate?: Date;
        endDate?: Date;
    }) {
        return this.auditDatabase.query(filters);
    }

    async generateComplianceReport(period: {
        startDate: Date;
        endDate: Date;
    }) {
        const events = await this.queryAuditLog({
            startDate: period.startDate,
            endDate: period.endDate
        });

        return {
            period,
            totalEvents: events.length,
            successRate: this.calculateSuccessRate(events),
            topFailureReasons: this.analyzeFailures(events),
            userActivity: this.aggregateUserActivity(events)
        };
    }
}
```

## Pattern 6: Anomaly Detection

```python
# Auth0 anomaly detection for security
class Auth0AnomalyDetection:
    def __init__(self):
        self.user_profiles = {}
        self.anomaly_threshold = 0.8

    async def analyze_login_attempt(self, user_id: str, attempt: dict):
        """Analyze login attempt for anomalies"""
        if user_id not in self.user_profiles:
            self.user_profiles[user_id] = self._create_profile(user_id)

        profile = self.user_profiles[user_id]

        # Check for anomalies
        anomalies = []

        # Check location
        if self._is_unusual_location(profile, attempt['location']):
            anomalies.append('unusual_location')

        # Check time
        if self._is_unusual_time(profile, attempt['timestamp']):
            anomalies.append('unusual_time')

        # Check device
        if self._is_new_device(profile, attempt['device_id']):
            anomalies.append('new_device')

        # Calculate anomaly score
        anomaly_score = len(anomalies) / 3

        if anomaly_score > self.anomaly_threshold:
            return {
                'requires_verification': True,
                'anomalies': anomalies,
                'score': anomaly_score
            }

        return {
            'requires_verification': False,
            'anomalies': [],
            'score': anomaly_score
        }

    def _create_profile(self, user_id: str):
        return {
            'user_id': user_id,
            'typical_locations': set(),
            'typical_times': set(),
            'known_devices': set()
        }
```

## Pattern 7: Delegated Administration

```typescript
// Auth0 delegated administration for enterprises
class Auth0DelegatedAdmin {
    async grantAdminPermissions(userId: string, orgId: string, adminRole: string) {
        const response = await fetch(
            `https://${this.domain}/api/v2/users/${userId}/roles`,
            {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.mgmt_token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    roles: [adminRole]
                })
            }
        );

        return response.json();
    }

    async createDelegatedOrgAdmin(email: string, orgId: string) {
        // 1. Create user or link existing user
        const user = await this.createOrGetUser(email);

        // 2. Add organization admin role
        await this.grantAdminPermissions(user.user_id, orgId, 'organization_admin');

        // 3. Send welcome email
        await this.sendDelegationWelcomeEmail(email, orgId);

        return user;
    }

    async syncDelegatedAdmins(orgId: string, admins: string[]) {
        const current = await this.getOrgAdmins(orgId);
        const currentEmails = new Set(current.map(a => a.email));

        // Remove admins not in new list
        for (const admin of current) {
            if (!admins.includes(admin.email)) {
                await this.removeAdminPermissions(admin.user_id, orgId);
            }
        }

        // Add new admins
        for (const email of admins) {
            if (!currentEmails.has(email)) {
                await this.createDelegatedOrgAdmin(email, orgId);
            }
        }
    }
}
```

---

**Version**: 1.0.0 | **Last Updated**: 2025-11-22
