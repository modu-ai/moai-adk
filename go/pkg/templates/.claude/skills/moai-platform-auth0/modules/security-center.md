# Security Center

Auth0 Security Center provides real-time observability and threat response capabilities within the Management Dashboard for monitoring CIAM anomalies and configuring attack mitigation.

## Overview

Security Center is a centralized dashboard for:
- Real-time threat monitoring
- Attack protection configuration
- Security metrics visualization
- Threat pattern analysis

Access: Dashboard > Security > Security Center

## Key Monitoring Features

### Real-Time Visibility

Total Traffic Monitoring:
- Authentication attempts
- Signup requests
- Password reset requests
- Overall request volume

Threat Counting:
- Detected threats by category
- Threat trends over time
- Comparative analysis

Application-Specific:
- Per-application metrics
- Identify targeted applications
- Focus security efforts

### Filtering and Aggregation

Time Period Selection:
- Up to 14 days of data
- Custom date ranges
- Trend analysis

Filter Options:
- By application
- By connection
- By threat type
- Combined filters

Auto-Aggregation:
- Per minute (short periods)
- Per hour (medium periods)
- Per day (longer periods)

## Threat Detection Categories

### Credential Stuffing

Detection: Behavioral patterns indicating machine-driven credential submission attempts.

Characteristics:
- Automated login attempts
- Multiple credentials tested
- Bot-like behavior patterns
- Account compromise attempts

Metrics Shown:
- Detection count
- Trend over time
- Affected applications
- Source analysis

### Signup Attacks

Detection: Patterns indicating automated account creation attempts.

Characteristics:
- High-velocity signups
- Bot-driven registration
- Fake account creation
- Spam account patterns

Metrics Shown:
- Attack frequency
- Volume trends
- Target applications
- Success/failure rates

### MFA Bypass

Detection: Patterns indicating attempts to circumvent multi-factor authentication.

Characteristics:
- Repeated MFA failures
- Unusual MFA patterns
- Potential social engineering
- Session manipulation attempts

Metrics Shown:
- Bypass attempt count
- Challenge/response patterns
- Affected users
- Method analysis

## Attack Protection Integration

### Bot Detection

Monitoring:
- Detection counts
- 7-day rolling view
- Trigger frequency
- Response effectiveness

Configuration Access:
- Quick settings access
- Enable/disable
- Sensitivity adjustment

### Suspicious IP Throttling

Monitoring:
- Throttling events
- Blocked IP count
- Rate limit triggers
- Geographic distribution

Configuration Access:
- Threshold adjustment
- AllowList management
- Response settings

### Brute Force Protection

Monitoring:
- Account protection triggers
- Blocked login attempts
- Notification frequency
- Unblock events

Configuration Access:
- Threshold settings
- Notification options
- AllowList management

### Breached Password Detection

Monitoring:
- Detection alerts
- Blocked credential usage
- User notifications sent
- Admin alert frequency

Configuration Access:
- Detection settings
- Response configuration
- Notification settings

### Multi-Factor Authentication

Monitoring:
- MFA challenge count
- Success rates
- Failure rates
- Factor usage distribution

Configuration Access:
- Factor management
- Policy settings
- Enrollment status

## Authentication Event Tracking

### Login Metrics

Over Past 7 Days:
- Total login attempts
- Successful logins
- Failed logins
- Success rate percentage

Breakdown Available:
- By application
- By connection
- By time period

### Signup Metrics

Tracking:
- Total signup attempts
- Successful signups
- Failed signups
- Conversion rates

Analysis:
- Growth trends
- Failure patterns
- Spam detection

## Using Security Center

### Daily Monitoring

Recommended Checks:
- Review threat counts
- Check for unusual patterns
- Verify protection triggers
- Monitor success rates

Quick Actions:
- Adjust thresholds
- Enable/disable features
- Investigate anomalies

### Threat Investigation

When Threat Detected:
1. Review threat category
2. Analyze affected scope
3. Check source patterns
4. Verify protection response
5. Adjust settings if needed

### Configuration Workflow

From Security Center:
1. Identify protection needs
2. Click through to settings
3. Configure appropriate protection
4. Return to monitor effect

## Metrics Interpretation

### Understanding Counts

Threat Counts:
- Number of detected threats
- Not necessarily attacks stopped
- Indicates protection activity

Traffic Counts:
- Total authentication requests
- Context for threat ratio
- Baseline for comparison

### Trend Analysis

Increasing Trends:
- May indicate attack increase
- Could be traffic growth
- Compare with baseline

Decreasing Trends:
- Protection effectiveness
- Attack subsiding
- Configuration changes

Spikes:
- Sudden attack activity
- System issues
- Investigate promptly

## Best Practices

### Regular Review

Daily:
- Quick overview of threats
- Check for spikes
- Verify normal patterns

Weekly:
- Trend analysis
- Configuration review
- Threshold assessment

Monthly:
- Comprehensive review
- Policy adjustments
- Documentation update

### Alert Response

When Threats Spike:
1. Assess threat type
2. Check protection status
3. Verify blocking active
4. Adjust if needed
5. Monitor resolution

### Threshold Optimization

Balance:
- Security vs. user experience
- False positive rate
- Attack prevention
- Business requirements

Process:
1. Monitor current performance
2. Identify issues
3. Adjust incrementally
4. Verify improvement

## Integration

### Log Export

For Advanced Analysis:
- Export to SIEM
- Custom dashboards
- Long-term storage

### Alerting

Configure Notifications:
- Threshold-based alerts
- Email notifications
- Integration with monitoring

### Reporting

Generate Reports:
- Security posture
- Threat trends
- Protection effectiveness
- Compliance evidence
