# SPEC Authoring Examples

## Example 1: Complete Feature SPEC with All 5 EARS Patterns

**File**: `.moai/specs/SPEC-105/spec.md`

```yaml
---
code: SPEC-105
title: Email Notification Service with Template Engine
status: active
created_at: 2025-11-12
updated_at: 2025-11-12
priority: high
effort: 8
version: 1.0.0
epic: NOTIFICATIONS-01
depends_on:
  - SPEC-104
  - SPEC-102
domains:
  - backend
  - infrastructure
acceptance_difficulty: high
rollback_risk: high
risks: |
  - Performance: 1000+ emails/second stress testing required
  - Reliability: SMTP provider failover must be tested
  - Security: Email address enumeration attack mitigation needed
tags:
  - notifications
  - email
  - async
  - templates
---

# SPEC-105: Email Notification Service with Template Engine

## Overview

Implement asynchronous email notification service with support for templated
messages, intelligent retry logic with exponential backoff, delivery tracking,
and deduplication to prevent duplicate emails within time windows.

## Requirements

### REQ-001: Universal Pattern - Basic Email Sending

```
SPEC-105-REQ-001: The notification service SHALL send emails asynchronously
without blocking the calling request, returning immediately to the caller.
```

**Related TEST**:
- `test_email_sent_asynchronously`
- `test_request_returns_immediately`
- `test_async_processing_in_background`


---

### REQ-002: Conditional Pattern - Retry Logic

```
SPEC-105-REQ-002: If email delivery fails on the first attempt,
the service SHALL retry up to 3 times with exponential backoff
(1 second, 2 seconds, 4 seconds) before marking the email as failed.
```

**Detailed Behavior**:
- Attempt 1: Send immediately
- Attempt 2 (if failed): Wait 1 second, retry
- Attempt 3 (if failed): Wait 2 seconds, retry
- Attempt 4 (if failed): Wait 4 seconds, retry
- Final failure: Mark as FAILED, trigger alert

**Related TEST**:
- `test_retry_on_first_failure`
- `test_exponential_backoff_timing`
- `test_max_retry_limit`
- `test_failed_status_after_max_retries`


---

### REQ-003: Unwanted Behavior Pattern - Duplicate Prevention

```
SPEC-105-REQ-003: The service SHALL NOT send duplicate emails
to the same recipient within a 5-minute deduplication window.
```

**Implementation Details**:
- Hash email recipient + template type + key parameters
- Store hash in Redis with 5-minute TTL
- Check hash before queuing new emails
- Return duplicate-prevented status to caller

**Unwanted Scenarios**:
- Duplicate email to same user within 5-minute window
- Different template versions sent to same user in rapid succession
- Race condition causing duplicate sends

**Related TEST**:
- `test_duplicate_email_prevented`
- `test_dedup_window_5_minutes`
- `test_different_templates_allowed`
- `test_race_condition_prevention`


---

### REQ-004: Stakeholder Pattern - Developer Experience

```
SPEC-105-REQ-004: As an application developer,
I want to send templated emails through a simple API
so that I don't have to manage HTML email formatting and rendering.
```

**Developer Experience**:
```python
# Simple, clean API
notify.send_email(
    recipient="user@example.com",
    template_name="welcome_email",
    context={
        "user_name": "John Doe",
        "activation_link": "https://..."
    }
)
# Returns immediately, email sent asynchronously
```

**Related TEST**:
- `test_simple_api_usage`
- `test_context_variables_interpolated`
- `test_template_not_found_error`


---

### REQ-005: Boundary Condition Pattern - Load Requirements

```
SPEC-105-REQ-005: The notification service SHALL process at least
1,000 emails per second and SHALL NOT exceed 500MB memory usage
under sustained load (10-minute duration) with standard templates.
```

**Performance Targets**:
- Throughput: ≥1,000 emails/second
- Memory: ≤500MB sustained
- CPU: ≤80% single core
- Latency: ≤50ms to queue (p99)

**Test Conditions**:
- Template size: 10KB (realistic HTML)
- Recipient list: 100K unique recipients
- Duration: 10 minutes sustained
- Concurrency: 10+ sender processes

**Related TEST**:
- `test_load_1000_emails_per_second` (benchmark)
- `test_memory_usage_sustained_load` (benchmark)
- `test_cpu_usage_under_load` (benchmark)


---

## Unwanted Behaviors

### Security Constraints

| Behavior | Rationale | Test |
|----------|-----------|------|
| The system SHALL NOT store plaintext email addresses in logs | Prevents PII exposure | `test_email_address_not_in_logs` |
| The system SHALL NOT expose SMTP credentials in error messages | Prevents credential leakage | `test_credentials_not_in_errors` |
| The system SHALL NOT send emails to unverified sender addresses | Prevents abuse/spoofing | `test_unverified_sender_rejected` |

### Performance Constraints

| Behavior | Rationale | Test |
|----------|-----------|------|
| The system SHALL NOT block on SMTP connection | Prevents slowdown for caller | `test_smtp_nonblocking` |
| The system SHALL NOT load entire email template into memory per recipient | Prevents memory explosion | `test_streaming_template_rendering` |
| The system SHALL NOT exceed network bandwidth quota | Prevents infrastructure overload | `test_bandwidth_throttling` |

### Reliability Constraints

| Behavior | Rationale | Test |
|----------|-----------|------|
| The system SHALL NOT lose email jobs if service restarts | Ensures durability | `test_job_persistence_on_restart` |
| The system SHALL NOT fail authentication if secondary cache is down | Ensures availability | `test_fallback_without_redis` |

---

## Acceptance Criteria

- [ ] **REQ-001**: Async email sending verified (non-blocking)
- [ ] **REQ-002**: Retry logic with exponential backoff implemented and tested
- [ ] **REQ-003**: Duplicate detection within 5-minute window working
- [ ] **REQ-004**: Simple API (recipient + template + context) functional
- [ ] **REQ-005**: Load test results: 1000+ emails/sec, ≤500MB memory
- [ ] **Code Coverage**: ≥85% (`src/notifications/email.py`)
- [ ] **Security Scan**: OWASP Top 10 compliance verified
- [ ] **Performance Baseline**: p99 latency ≤50ms for queue operation
- [ ] **Documentation**: API guide with 5+ usage examples
- [ ] **Integration**: Verified with production email provider (SendGrid/SES)

---

## Implementation Notes

### Architecture

```
Application Code
    ↓
notify.send_email() [Non-blocking]
    ↓
Async Job Queue (Redis/Celery)
    ↓
Email Worker Process
    ├─ Template Rendering
    ├─ Header/Footer Injection
    ├─ Variable Substitution
    ├─ SMTP Retry Logic
    └─ Delivery Tracking Database

Monitoring:
- Email queue depth
- Failed emails count
- Retry attempt distribution
- SMTP provider response times
```

### Database Schema

```sql
-- Email jobs table
CREATE TABLE email_jobs (
    id BIGSERIAL PRIMARY KEY,
    recipient_email VARCHAR(255),
    template_name VARCHAR(100),
    context JSONB,
    status ENUM ('pending', 'processing', 'sent', 'failed'),
    attempt_count INT DEFAULT 0,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    sent_at TIMESTAMP,
    error_message TEXT
);

CREATE INDEX idx_status_retry ON email_jobs(status, next_retry_at);
CREATE INDEX idx_recipient_dedup ON email_jobs(recipient_email, template_name);
```

### Configuration

```yaml
# notifications.yaml
email:
  provider: sendgrid  # sendgrid | aws-ses | mailgun
  api_key: ${SENDGRID_API_KEY}
  
  retry:
    max_attempts: 3
    backoff_base: 1000  # milliseconds
    backoff_multiplier: 2
  
  deduplication:
    enabled: true
    window_seconds: 300  # 5 minutes
    backend: redis
  
  performance:
    batch_size: 100  # emails per SMTP connection
    max_concurrent: 10  # concurrent SMTP connections
    worker_count: 5  # background worker processes
  
  monitoring:
    log_email_addresses: false  # never log PII
    log_templates: true  # log template names
    track_delivery: true  # store delivery events
```

---

## Testing Strategy

### Unit Tests (45% coverage)

```python
# tests/test_email_service.py

import pytest
from src.notifications.email import EmailService, DuplicateEmailError

@pytest.fixture
def email_service():
    """Create email service instance for testing."""
    return EmailService(config="test")

class TestBasicSending:
    """Test REQ-001: Async email sending."""
    
    def test_send_email_returns_immediately(self, email_service):
        """Test that send_email returns immediately without blocking."""
        start = time.time()
        result = email_service.send_email(
            recipient="user@example.com",
            template_name="welcome",
            context={"name": "John"}
        )
        duration = time.time() - start
        
        assert duration < 0.1  # <100ms
        assert result["job_id"] is not None
        assert result["status"] == "queued"

class TestRetryLogic:
    """Test REQ-002: Retry with exponential backoff."""
    
    def test_retry_after_first_failure(self, email_service, mock_smtp):
        """Test exponential backoff after first SMTP failure."""
        mock_smtp.fail_count = 1  # Fail first attempt
        
        email_service.send_email(...)
        email_service.process_jobs()  # First attempt (fails)
        
        # Check retry scheduled
        next_retry = email_service.get_next_retry_time()
        assert next_retry == 1  # 1 second

class TestDuplication:
    """Test REQ-003: Duplicate prevention."""
    
    def test_duplicate_prevented(self, email_service):
        """Test duplicate emails rejected within window."""
        # Send first email
        email_service.send_email(
            recipient="user@example.com",
            template_name="welcome",
            context={"key": "value"}
        )
        
        # Try to send duplicate
        with pytest.raises(DuplicateEmailError):
            email_service.send_email(
                recipient="user@example.com",
                template_name="welcome",
                context={"key": "value"}
            )
```

### Integration Tests (35% coverage)

```python
# tests/test_email_integration.py

@pytest.mark.integration
class TestEmailIntegration:
    """Integration tests with real Redis and database."""
    
    def test_end_to_end_email_flow(self, integration_db, redis):
        """Test complete flow: send → queue → process → deliver."""
        service = EmailService(db=integration_db, cache=redis)
        
        # Send email
        result = service.send_email(
            recipient="test@example.com",
            template_name="welcome"
        )
        
        # Process queue
        service.process_jobs()
        
        # Verify in database
        job = integration_db.get_job(result["job_id"])
        assert job.status == "sent"
```

### Load Tests (10% coverage)

```bash
# tests/load/load_test.py

import concurrent.futures
import time

def test_load_1000_emails_per_second():
    """Benchmark: 1000 emails/second for 10 minutes."""
    service = EmailService()
    
    start_memory = get_memory_usage()
    start_time = time.time()
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
        for i in range(600_000):  # 600k emails over 10 minutes
            executor.submit(
                service.send_email,
                recipient=f"user{i}@example.com",
                template_name="welcome"
            )
    
    duration = time.time() - start_time
    peak_memory = get_memory_usage()
    
    throughput = 600_000 / duration
    memory_used = peak_memory - start_memory
    
    assert throughput >= 1000  # emails/second
    assert memory_used <= 500  # MB
```

---

## References

### Official Documentation

- **Sendgrid API**: https://docs.sendgrid.com/
- **AWS SES**: https://docs.aws.amazon.com/ses/
- **Mailgun API**: https://documentation.mailgun.com/
- **Email RFC 5322**: https://tools.ietf.org/html/rfc5322
- **MIME Types RFC 2045**: https://tools.ietf.org/html/rfc2045

### Related SPECs

- **SPEC-104**: User Profile Service (prerequisite)
- **SPEC-102**: Email Template Schema Definition (prerequisite)
- **SPEC-106**: Email Delivery Tracking (follow-up)
- **SPEC-107**: Notification Preferences (optional dependency)

### TAGs


---

## Example 2: Simple Bug Fix SPEC

**File**: `.moai/specs/SPEC-012/spec.md`

```yaml
---
code: SPEC-012
title: Fix Race Condition in User Cache Invalidation
status: active
created_at: 2025-11-12
updated_at: 2025-11-12
priority: high
effort: 3
version: 1.0.0
domains:
  - backend
  - cache
tags:
  - bug-fix
  - concurrency
  - cache
---

# SPEC-012: Fix Race Condition in User Cache Invalidation

## Overview

Address race condition in user profile cache invalidation that allows stale
data to be served when multiple updates occur rapidly. Implementation uses
distributed locking to ensure sequential cache updates.

## Requirements

### REQ-001: Universal - Mutex Protection

```
SPEC-012-REQ-001: The cache invalidation mechanism SHALL use
distributed mutual exclusion (mutex) to prevent concurrent updates
to the same user profile cache entry.
```

### REQ-002: Unwanted Behavior - Race Condition Prevention

```
SPEC-012-REQ-002: The system SHALL NOT serve stale user profile data
when multiple concurrent updates occur to the same user record.
```

## Acceptance Criteria

- [ ] Race condition test case fails before fix
- [ ] Distributed lock implementation added
- [ ] Race condition test case passes after fix
- [ ] Performance impact <5% (lock acquisition time)
- [ ] Code coverage ≥85%

## Testing

```python
def test_race_condition_prevented():
    """Test that concurrent updates don't cause stale data."""
    user_id = 123
    
    # Simulate 10 concurrent updates
    def update_user(new_data):
        user_service.update(user_id, new_data)
        cached = cache.get(f"user:{user_id}")
        return cached
    
    results = run_concurrent(
        update_user,
        num_threads=10,
        args=[{"name": f"User {i}"} for i in range(10)]
    )
    
    # All threads should see their own update
    # (no stale/mixed data from other threads)
    assert len(set(results)) == 10  # 10 unique final states
```

---

## Example 3: Minimal SPEC for Quick Fix

**File**: `.moai/specs/SPEC-025/spec.md`

```yaml
---
code: SPEC-025
title: Update Dependency: TypeScript 5.9.0
status: active
created_at: 2025-11-12
updated_at: 2025-11-12
priority: medium
effort: 1
domains:
  - frontend
tags:
  - dependencies
  - typescript
---

# SPEC-025: Update Dependency: TypeScript 5.9.0

## Requirements

### REQ-001

```
SPEC-025-REQ-001: The project SHALL update TypeScript
from 5.8.x to 5.9.x in package.json.
```

### REQ-002

```
SPEC-025-REQ-002: The build system SHALL NOT show
any new TypeScript compiler errors after the update.
```

## Acceptance Criteria

- [ ] package.json: `"typescript": "^5.9.0"`
- [ ] package-lock.json regenerated
- [ ] `npm run build` passes
- [ ] All tests pass
- [ ] No new linting errors

---

# Summary

These examples demonstrate:
1. **Complete feature SPEC** (SPEC-105): All 5 EARS patterns, comprehensive testing
2. **Bug fix SPEC** (SPEC-012): Smaller scope, focused requirements
3. **Minimal SPEC** (SPEC-025): Quick updates, simple requirements

Use these as templates when authoring your own SPECs.
