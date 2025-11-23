# Security Logging & Monitoring Failures (A09)

## Overview

Insufficient logging and monitoring allows attacks to go undetected, delaying incident response and enabling attackers to maintain persistence.

## Critical Gaps

### Missing Security Events
- Failed login attempts
- Authorization failures
- Input validation failures
- Security exceptions

### Poor Log Management
- Missing centralized logging
- No log retention policy
- Logs not monitored
- Missing alerting

## Remediation Patterns

### Security Event Logging

**Comprehensive Login Logging**:
```javascript
const winston = require('winston');

const securityLogger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  defaultMeta: { service: 'auth' },
  transports: [
    new winston.transports.File({ filename: 'security.log' }),
    new winston.transports.Console()
  ]
});

app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  const ip = req.ip;
  const userAgent = req.headers['user-agent'];

  const user = await db.users.findByUsername(username);
  const valid = user && await bcrypt.compare(password, user.passwordHash);

  if (valid) {
    securityLogger.info('Login successful', {
      event: 'login_success',
      username,
      userId: user.id,
      ip,
      userAgent,
      timestamp: Date.now()
    });

    return res.json({ token: createToken(user) });
  } else {
    securityLogger.warn('Login failed', {
      event: 'login_failure',
      username,
      ip,
      userAgent,
      timestamp: Date.now()
    });

    // Check for brute force
    const attempts = await countFailedAttempts(username, ip);
    if (attempts > 5) {
      securityLogger.critical('Possible brute force detected', {
        event: 'brute_force_attempt',
        username,
        ip,
        attempts,
        timestamp: Date.now()
      });

      await triggerSecurityAlert('brute_force', { username, ip, attempts });
    }

    return res.status(401).json({ error: 'Invalid credentials' });
  }
});
```

### Authorization Logging

```python
import logging

security_logger = logging.getLogger('security')

def check_authorization(user, resource, action):
    """Check authorization with comprehensive logging."""
    allowed = has_permission(user, resource, action)

    if allowed:
        security_logger.info(
            'Authorization granted',
            extra={
                'event': 'authz_success',
                'user_id': user.id,
                'resource': resource.id,
                'action': action,
                'timestamp': time.time()
            }
        )
    else:
        security_logger.warning(
            'Authorization denied',
            extra={
                'event': 'authz_failure',
                'user_id': user.id,
                'resource': resource.id,
                'action': action,
                'timestamp': time.time()
            }
        )

        # Check for privilege escalation attempts
        if is_privilege_escalation(user, resource, action):
            security_logger.critical(
                'Privilege escalation attempt',
                extra={
                    'event': 'privilege_escalation',
                    'user_id': user.id,
                    'resource': resource.id,
                    'action': action
                }
            )

    return allowed
```

### Centralized Logging

**ELK Stack (Elasticsearch, Logstash, Kibana)**:
```javascript
const elasticsearch = require('@elastic/elasticsearch');

const client = new elasticsearch.Client({
  node: 'https://elasticsearch:9200'
});

async function logSecurityEvent(event) {
  await client.index({
    index: 'security-events',
    document: {
      '@timestamp': new Date(),
      event_type: event.type,
      user_id: event.userId,
      ip_address: event.ip,
      action: event.action,
      result: event.result,
      details: event.details
    }
  });
}
```

**Structured Logging (JSON)**:
```python
import json
import logging

class JSONFormatter(logging.Formatter):
    def format(self, record):
        log_data = {
            'timestamp': self.formatTime(record),
            'level': record.levelname,
            'logger': record.name,
            'message': record.getMessage(),
        }

        if hasattr(record, 'event'):
            log_data['event'] = record.event

        if hasattr(record, 'user_id'):
            log_data['user_id'] = record.user_id

        return json.dumps(log_data)

handler = logging.StreamHandler()
handler.setFormatter(JSONFormatter())

logger = logging.getLogger('security')
logger.addHandler(handler)
```

### Real-Time Monitoring

**Security Monitoring Dashboard**:
```python
from prometheus_client import Counter, Gauge, Histogram

# Metrics
login_attempts = Counter('login_attempts_total', 'Total login attempts', ['result'])
active_sessions = Gauge('active_sessions', 'Number of active sessions')
request_duration = Histogram('request_duration_seconds', 'Request duration')

# Track metrics
@login_required
def protected_route():
    with request_duration.time():
        # Route logic
        pass

def on_login_success(user):
    login_attempts.labels(result='success').inc()
    active_sessions.inc()

def on_login_failure():
    login_attempts.labels(result='failure').inc()
```

**Alerting Rules**:
```yaml
# Prometheus alerting rules
groups:
  - name: security_alerts
    rules:
      - alert: BruteForceDetected
        expr: rate(login_attempts_total{result="failure"}[5m]) > 10
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Possible brute force attack"
          description: "More than 10 failed login attempts per minute"

      - alert: PrivilegeEscalationAttempt
        expr: increase(authz_failure_total{action="admin"}[5m]) > 5
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Privilege escalation detected"
          description: "Multiple authorization failures for admin actions"
```

### Audit Trail

**Comprehensive Audit Logging**:
```python
class AuditLogger:
    def __init__(self, db):
        self.db = db

    async def log_event(self, event_type, user_id, details):
        """Log security event to audit trail."""
        await self.db.audit_log.create({
            'event_type': event_type,
            'user_id': user_id,
            'timestamp': datetime.utcnow(),
            'ip_address': request.remote_addr,
            'user_agent': request.headers.get('User-Agent'),
            'details': details
        })

    async def get_user_activity(self, user_id, start_date, end_date):
        """Retrieve user activity for investigation."""
        return await self.db.audit_log.find({
            'user_id': user_id,
            'timestamp': {
                '$gte': start_date,
                '$lte': end_date
            }
        }).sort('timestamp', -1)
```

### Log Retention

**Retention Policy**:
```python
class LogRetentionPolicy:
    def __init__(self):
        self.retention_days = {
            'security': 365,      # 1 year
            'audit': 2555,        # 7 years (compliance)
            'application': 90,    # 3 months
            'debug': 7            # 1 week
        }

    async def cleanup_old_logs(self):
        """Remove logs older than retention period."""
        for log_type, days in self.retention_days.items():
            cutoff = datetime.utcnow() - timedelta(days=days)

            deleted = await self.db.logs.delete_many({
                'type': log_type,
                'timestamp': {'$lt': cutoff}
            })

            logger.info(f'Deleted {deleted.deleted_count} old {log_type} logs')
```

## Best Practices

### Security Event Coverage

**Events to Log**:
- ✅ Authentication (success/failure)
- ✅ Authorization decisions
- ✅ Input validation failures
- ✅ Administrative actions
- ✅ Data access (sensitive)
- ✅ Configuration changes
- ✅ Security exceptions

**Events NOT to Log**:
- ❌ Passwords (even hashed)
- ❌ Session tokens
- ❌ Credit card numbers
- ❌ Personal identifiers (raw)
- ❌ Encryption keys

### Log Security

```python
def sanitize_log_data(data):
    """Remove sensitive data before logging."""
    sensitive_fields = ['password', 'token', 'ssn', 'credit_card']

    if isinstance(data, dict):
        return {
            k: '***REDACTED***' if k in sensitive_fields else v
            for k, v in data.items()
        }

    return data

# Usage
logger.info('User updated', extra=sanitize_log_data(user_data))
```

## Validation Checklist

- [ ] Security events logged
- [ ] Centralized log management
- [ ] Real-time alerting configured
- [ ] Log retention policy
- [ ] Regular log reviews
- [ ] Sensitive data redacted
- [ ] Audit trail immutable
- [ ] Monitoring dashboard active

## Testing

```python
def test_security_logging():
    """Verify security events are logged."""
    # Failed login should be logged
    response = client.post('/login', json={
        'username': 'test',
        'password': 'wrong'
    })

    logs = read_security_logs()
    assert any(l['event'] == 'login_failure' for l in logs)

    # Successful login should be logged
    response = client.post('/login', json={
        'username': 'test',
        'password': 'correct'
    })

    logs = read_security_logs()
    assert any(l['event'] == 'login_success' for l in logs)
```

---

**Last Updated**: 2025-11-24
**OWASP Category**: A09:2021
**CWE**: CWE-778 (Insufficient Logging)
