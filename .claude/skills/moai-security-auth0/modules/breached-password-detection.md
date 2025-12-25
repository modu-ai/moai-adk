# Breached Password Detection

Auth0 Breached Password Detection protects applications by identifying when user credentials appear in known security breaches and taking appropriate action.

## How It Works

Auth0 monitors third-party data breaches and compares user credentials against known compromised credential databases. When a match is found, the system can block access and notify affected parties.

## Detection Methods

### Standard Detection

Included with B2B/B2C Professional or Enterprise plans.

Characteristics:
- Tracks publicly released breach data
- Detection time: 7-13 months after breach disclosure
- Relies on public breach announcements and databases
- Comprehensive coverage of major breaches

### Credential Guard (Enterprise Add-on)

Enhanced detection with dedicated security team access.

Characteristics:
- Accesses non-public breach data
- Detection time: 12-36 hours
- Coverage: 200+ countries
- Proactive breach intelligence
- Dark web monitoring

## Configuration

### Dashboard Navigation

Access: Dashboard > Security > Attack Protection > Breached Password Detection

### Response Scenarios

Block on Signup:
- Prevents account creation with compromised credentials
- User must choose a different password
- Immediate protection for new accounts

Block on Login:
- Prevents authentication with breached passwords
- Existing users must reset password
- Protects against credential stuffing

Block on Password Reset:
- Prevents setting compromised passwords during reset
- Ensures clean password after breach detection
- Maintains protection through password changes

### Notification Configuration

User Notifications:
- Alert users when their credentials are found in breaches
- Includes password reset instructions
- Configurable messaging

Admin Notifications:
- Alerts for signup attempts with breached passwords
- Alerts for login attempts with compromised credentials
- Frequency options: Immediate, Daily, Weekly, Monthly

### Response Combinations

Recommended for Most Applications:
- Block on signup: Enabled
- Block on login: Enabled
- User notifications: Enabled
- Admin notifications: Weekly

High-Security Applications:
- All blocking options enabled
- User notifications: Enabled
- Admin notifications: Immediate

Monitoring Only:
- All blocking disabled
- User notifications: Disabled
- Admin notifications: Enabled (for analysis)

## Testing

Auth0 provides test credentials for verification:

Test Password Pattern: Any password starting with AUTH0-TEST-

Examples:
- AUTH0-TEST-password123
- AUTH0-TEST-breached
- AUTH0-TEST-anything

Testing Process:
1. Enable breached password detection
2. Attempt signup or login with test password
3. Verify blocking or notification behavior
4. Confirm expected user experience

Note: Test passwords trigger detection without affecting production breach databases.

## Library Requirements

Ensure SDK versions support breached password detection:

Lock.js:
- Version 11.33.3 or later
- Full feature support

Auth0.js:
- Latest version recommended
- Full feature support

Native SDKs:
- Auth0.swift: 1.28.0+
- Auth0.Android: 1.25.0+

## User Experience

When Blocked on Signup:
- User sees password requirement message
- Must choose different password
- Clear guidance on password selection

When Blocked on Login:
- User sees account security message
- Directed to password reset flow
- Email sent with reset instructions

Notification Content:
- Explains credentials found in breach
- Does not reveal which breach
- Provides password reset link
- Valid for 5 days

## Integration Considerations

Password Requirements:
- Combine with strong password policies
- Consider password strength meters
- Provide clear error messaging

User Communication:
- Customize breach notification templates
- Explain without causing panic
- Emphasize security benefits

Recovery Flows:
- Ensure password reset works smoothly
- Consider step-up authentication
- Monitor reset completion rates

## Metrics and Monitoring

Track in Tenant Logs:
- Blocked signup attempts
- Blocked login attempts
- Notification deliveries
- Password reset completions

Dashboard Metrics:
- Detection counts over time
- Block rates by scenario
- User compliance rates

## Best Practices

Deployment:
1. Enable in monitoring mode first
2. Review detection rates
3. Enable blocking on signup
4. Enable blocking on login after communication
5. Monitor user support requests

Communication:
- Announce security feature to users
- Explain why passwords may be rejected
- Provide password manager recommendations
- Set expectations for breach notifications

Ongoing Management:
- Review detection metrics regularly
- Adjust notification frequency based on volume
- Keep SDKs updated for latest breach data
- Consider Credential Guard for enhanced protection
