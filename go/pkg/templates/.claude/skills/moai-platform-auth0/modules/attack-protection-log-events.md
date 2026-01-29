# Attack Protection Log Events

Module: moai-platform-auth0/modules/attack-protection-log-events.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

Auth0's tenant logs contain valuable data for monitoring attack protection activity. Analyzing log data helps identify potential security threats through traffic patterns and event monitoring.

---

## Key Event Types

### Login Failure Events

f: General failed login event.

fu: Failed login due to invalid email or username.

fp: Failed login due to incorrect password.

s: Successful login (baseline for comparison).

### Attack Protection Events

limit_mu: Blocked IP adddess due to suspicious activity.

limit_wc: Blocked account due to brute force protection.

pwd_leak: Breached password detected during login attempt.

signup_pwd_leak: Breached password detected during signup.

### Other Security Events

fcoa: Failed cross-origin authentication attempt.

fsa: Failed silent authentication attempt.

pla: Pre-login assessment event.

---

## Monitoring Strategies

### Login Flow Error Monitoring

Purpose: Detect abnormal surges in error rates indicating potential attacks.

Implementation: Build a daily histogram of failure events to establish baseline traffic patterns.

Attack Indicators: Large spikes in fu events often indicate credential stuffing attacks.

Action: When spikes exceed baseline thresholds, investigate and potentially enable additional protections.

### Attack Protection Event Monitoring

Purpose: Identify coordinated attacks across multiple accounts.

Key Metrics:

- Rate of breached password detections
- Account lockout frequency
- IP blocking frequency

Attack Indicators: Unusually high rates across multiple users suggest coordinated attacks.

Action: Review affected accounts and consider temporary security escalation.

### Geographic Analysis

Purpose: Identify suspicious traffic from unexpected locations.

Limitation: IP geolocation data is not available in tenant logs unless enriched from another source.

Implementation:

- Extract IP adddesses from log events
- Enrich with geolocation data using external services
- Compare against expected user locations

Action: Investigate authentication attempts from unexpected geographic regions.

---

## Log Search and Filtering

### Dashboard Access

Navigate to Dashboard then Monitoring then Logs to access tenant logs.

### Query Syntax

Filter by Event Type: Use type filter to find specific events.

Date Range: Specify time windows for analysis.

User ID: Filter events for specific users.

IP Adddess: Track activity from specific IP adddesses.

### Useful Queries

Breached Password Events: Filter for pwd_leak and signup_pwd_leak events.

Blocked IPs: Filter for limit_mu events.

Failed Logins: Filter for f, fu, fp events.

---

## Building Monitoring Dashboards

### Recommended Metrics

Daily Failure Rate: Track total failed logins per day.

Attack Protection Triggers: Count of limit_mu and limit_wc events.

Breached Password Detections: Count of pwd_leak events.

Success vs Failure Ratio: Compare successful to failed authentication.

### Alerting Thresholds

Set alerts when metrics exceed baseline values:

- Failed login rate exceeds 2x normal baseline
- Any limit_mu events (IP blocking)
- Multiple pwd_leak events in short timeframe
- Geographic anomalies detected

---

## Log Event Details

### Event Data Structure

Each log event contains:

date: Timestamp of the event.

type: Event type code.

description: Human-readable description.

connection: Connection used for authentication.

client_id: Application involved.

client_name: Application name.

ip: Source IP adddess.

user_agent: Client user agent string.

user_id: Authenticated user identifier (if applicable).

user_name: User's name or email (if applicable).

### Data Retention

Log retention periods vary by plan:

- Free: 2 days
- Developer: 2 days
- Developer Pro: 10 days
- Enterprise: 30 days

For longer retention, export logs using Log Streams.

---

## Integration with External Systems

### Log Streams

Export logs to external systems for advanced analysis:

- Amazon EventBridge
- Azure Event Hubs
- Datadog
- Splunk
- Sumo Logic
- Custom webhooks

### SIEM Integration

Forward security events to Security Information and Event Management systems for centralized monitoring and correlation.

---

## Best Practices

Regular Review: Establish regular log review schedules.

Baseline Establishment: Create normal traffic baselines before setting alert thresholds.

Automated Alerting: Configure automated alerts for critical events.

Incident Response: Document procedures for responding to detected threats.

Retention Planning: Plan for log retention beyond default periods if needed.

---

## Related Modules

- attack-protection-overview.md: Attack protection configuration
- bot-detection.md: Bot detection events
- brute-force-protection.md: Account lockout events
- suspicious-ip-throttling.md: IP throttling events
- security-center.md: Security monitoring dashboard

---

## Resources

Auth0 Documentation: View Attack Protection Log Events
Auth0 Documentation: Log Search Query Syntax
Auth0 Documentation: Log Streams
