# Suspicious IP Throttling

Auth0 Suspicious IP Throttling automatically blocks traffic from IP addresses exhibiting high-velocity login or signup attempts, protecting against large-scale automated attacks.

## How It Works

The system tracks login and signup attempt velocity per IP address. When an address exceeds configured thresholds, Auth0 throttles subsequent attempts by responding with HTTP 429 (Too Many Requests) status codes.

This protection is enabled by default on new tenants.

## Velocity Detection Mechanisms

### Login Attempt Tracking

Monitoring Period: Daily (24-hour rolling window)

Detection Logic:
- Counts failed login attempts per IP address
- Threshold based on total failures across all accounts
- Does not require targeting specific account

Throttling Behavior:
- Once threshold exceeded, throttling activates
- Allowed attempts distributed evenly across 24 hours
- Example: Throttling rate of 100 grants approximately one attempt every 15 minutes

### Signup Attempt Tracking

Monitoring Period: Per minute

Detection Logic:
- Counts all signup attempts (successful and failed)
- Threshold based on attempts within one-minute window
- Triggers on high-velocity account creation

Throttling Behavior:
- When limit exceeded, further signups blocked
- Throttling rate distributes attempts over 24 hours
- Example: Rate of 72,000 allows roughly one attempt per second

## Configuration

### Dashboard Navigation

Access: Dashboard > Security > Attack Protection > Suspicious IP Throttling

### Threshold Settings

Login Thresholds:
- Maximum failed login attempts per day
- Throttling rate (attempts allowed per 24 hours after blocking)

Signup Thresholds:
- Maximum signup attempts per minute
- Throttling rate for signup after blocking

### IP AllowList

Add trusted IP sources to exempt from throttling:
- Up to 100 IP addresses or CIDR ranges
- Useful for automated testing systems
- Protects known-good high-volume sources

### Response Configuration

Enable Traffic Limiting:
- Activates HTTP 429 responses
- Required for active protection

Administrator Notifications:
- Email alerts when thresholds exceeded
- Configurable notification settings

Monitoring Mode:
- Disable all response actions
- Events still logged
- Useful for threshold calibration

## HTTP 429 Response

When throttled, requests receive:

Status: 429 Too Many Requests

Response includes:
- Error description
- Retry-after guidance
- Rate limit information

Client Handling:
- Implement exponential backoff
- Display user-friendly message
- Avoid immediate retries

## Important Considerations

### What Does Not Count

These request types do not increment thresholds:
- Malformed requests
- Schema validation errors
- Requests from AllowListed IPs
- Successful authentications (for login tracking)

### Backend Applications

For Resource Owner Password Grant:
- Auth0 sees application server IP, not user IP
- Must manually pass client IP via auth0-forwarded-for header
- Without this, all users appear from same IP

Implementation:
- Extract client IP from X-Forwarded-For or similar
- Include in auth0-forwarded-for header
- Ensure proper IP extraction behind proxies

### Shared IP Environments

Organizations behind NAT or proxies:
- All users share same public IP
- More likely to trigger throttling
- Consider higher thresholds or AllowList

Mobile Networks:
- Carrier NAT shares IP across subscribers
- Geographic IP pools may appear suspicious
- Consider mobile-specific thresholds

## Monitoring and Metrics

### Security Center

Access: Dashboard > Security > Security Center

Available Metrics:
- Throttling events over time
- Top throttled IPs
- Geographic distribution
- Attack pattern analysis

### Tenant Logs

Event Types:
- Rate limit exceeded events
- Throttling trigger events
- IP blocking/unblocking

Log Details:
- Source IP address
- Attempt counts
- Threshold exceeded
- Action taken

## Integration with Attack Protection

Layered with Bot Detection:
- Bot detection evaluates request patterns
- Suspicious IP throttling evaluates velocity
- Both can trigger on same request

Layered with Brute Force Protection:
- Suspicious IP tracks across all accounts
- Brute force tracks per account per IP
- Different protection scopes

## Best Practices

### Initial Configuration

1. Enable in monitoring mode
2. Analyze baseline traffic patterns
3. Identify high-volume legitimate sources
4. Configure AllowList for trusted IPs
5. Enable throttling with conservative thresholds
6. Monitor false positive rate
7. Adjust thresholds based on data

### Threshold Selection

Conservative (More Protection):
- Lower thresholds
- Faster throttling response
- May impact legitimate high-volume users

Permissive (Better UX):
- Higher thresholds
- Allow more attempts before throttling
- Less protection against sophisticated attacks

### For Different Application Types

Consumer Applications:
- Moderate login threshold
- Higher signup threshold (organic growth periods)
- Monitor for registration spam

Enterprise Applications:
- Lower thresholds acceptable
- AllowList corporate IP ranges
- Integrate with enterprise identity providers

API-Heavy Applications:
- Higher thresholds for legitimate API usage
- AllowList application server IPs
- Ensure auth0-forwarded-for header implementation

### Ongoing Management

Regular Reviews:
- Check throttling events weekly
- Identify new legitimate high-volume sources
- Update AllowList as needed

Attack Response:
- Review attack patterns
- Adjust thresholds temporarily if needed
- Document attack characteristics

Threshold Tuning:
- Balance security with user experience
- Consider seasonal traffic variations
- Account for growth in user base

## Troubleshooting

Legitimate Traffic Throttled:
- Add IP to AllowList
- Increase thresholds
- Check for auth0-forwarded-for header issues

Throttling Not Triggering:
- Verify feature is enabled
- Check if IP is AllowListed
- Confirm threshold configuration
- Review request patterns

429 Errors Not Handled:
- Implement proper error handling in client
- Add retry logic with backoff
- Display appropriate user message
