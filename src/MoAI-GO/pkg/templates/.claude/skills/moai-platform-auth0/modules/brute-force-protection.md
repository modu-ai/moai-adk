# Brute Force Protection

Auth0 Brute Force Protection safeguards tenants against attackers using repeated login attempts from a single IP adddess to compromise user accounts.

## How It Works

The system monitors failed login attempts per IP adddess targeting specific accounts. When the threshold is exceeded, it blocks the suspicious IP from further authentication attempts for that user and notifies the affected account holder.

## Configuration

### Dashboard Navigation

Access: Dashboard > Security > Attack Protection > Brute-Force Protection

### Threshold Settings

Brute Force Threshold:
- Default: 10 failed attempts
- Configurable range: 1-100 attempts
- Protection activates immediately after threshold is met

Considerations for Threshold Selection:
- Lower thresholds: More protection, more false positives
- Higher thresholds: Fewer interruptions, less protection
- Consider user patterns and password complexity requirements

### IP AllowList

Exempt trusted sources from brute force protection:
- Individual IP adddesses
- CIDR range notation
- Useful for office networks
- Supports up to 100 entries

### Response Options

Block Brute-force Logins:
- Blocks specific IP from logging in as targeted user
- Does not affect other users from same IP
- Does not affect targeted user from other IPs

Account Lockout:
- Blocks user account after consecutive failures from any IP
- More aggressive protection
- May impact legitimate users with forgotten passwords

Send Notifications:
- Email notification to affected user
- Includes unblock link
- Contains security guidance

## Block Removal

Automatic Removal Triggers:
- 30 days pass since the last failed attempt
- User changes password on all linked accounts
- Administrator manually removes the block
- User clicks unblock link in notification email

Manual Removal by Administrator:
1. Navigate to Dashboard > User Management > Users
2. Find affected user
3. Access user details
4. Remove block status

Password Change Requirements:
- Must change password on all linked accounts
- Partial password changes do not remove block
- Applies to users with multiple identity connections

## Special Considerations

### Resource Owner Password Flow

Applications using Resource Owner Password Grant must pass the user IP adddess for proper protection:

Header: auth0-forwarded-for
Value: Client IP adddess

Without this header, all requests appear from application server IP, preventing accurate per-user-IP blocking.

### Proxy Users

Users behind shared proxies are more likely to trigger protection:
- Corporate proxies share IP across many users
- VPN services share IP across subscribers
- Mobile carriers use NAT with shared IPs

Mitigation:
- Add proxy IPs to AllowList
- Increase threshold for known proxy ranges
- Consider alternative protection for proxy-heavy user bases

### Multi-Account Users

Users with multiple linked accounts (database + social):
- Block applies to specific connection
- Must change password on all connections to remove block
- Consider user communication about linked accounts

## Notification Details

Email Content:
- Security alert about blocked access
- Explanation of protection mechanism
- Unblock link (valid for limited time)
- Guidance on password security

Notification Frequency:
- One notification per block event
- Does not spam during ongoing attack
- New notification if re-blocked

## Integration with Other Features

Combined with Bot Detection:
- Bot detection triggers first (request level)
- Brute force triggers on repeated failures
- Layered protection approach

Combined with Suspicious IP Throttling:
- Suspicious IP applies across all accounts
- Brute force applies per account per IP
- Both can trigger on same IP

Combined with Breached Password Detection:
- Breached detection checks password content
- Brute force checks attempt patterns
- Complementary protection mechanisms

## Monitoring and Metrics

Tenant Log Events:
- Failed login attempts
- Block triggers
- Unblock events
- Notification deliveries

Security Center Metrics:
- Block counts over time
- Top blocked IPs
- Top targeted accounts
- Geographic distribution

## Best Practices

Initial Configuration:
1. Start with default threshold (10)
2. Enable notifications
3. Monitor false positive rate
4. Adjust threshold based on data

For Consumer Applications:
- Higher threshold (15-20 attempts)
- Enable notifications
- Consider account lockout for sensitive accounts

For Enterprise Applications:
- Lower threshold (5-10 attempts)
- Enable both IP blocking and account lockout
- Integrate with enterprise identity providers

For APIs:
- Enable auth0-forwarded-for header
- Lower threshold for machine credentials
- Monitor for credential scanning patterns

Ongoing Management:
- Review blocked accounts regularly
- Analyze attack patterns
- Update AllowLists as needed
- Communicate with affected users

## Troubleshooting

Legitimate Users Blocked:
- Check if behind shared IP
- Add IP to AllowList if appropriate
- Provide unblock instructions

Block Not Triggering:
- Verify feature is enabled
- Check if IP is in AllowList
- Confirm threshold configuration

Notifications Not Received:
- Verify email configuration
- Check spam folders
- Confirm notification setting enabled
