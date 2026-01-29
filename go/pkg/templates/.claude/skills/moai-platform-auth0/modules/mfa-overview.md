# Multi-Factor Authentication Overview

Auth0 Multi-Factor Authentication (MFA) is a user verification method requiring more than one type of validation, preventing unauthorized access even when credentials are compromised.

## Why Use MFA

MFA reduces cyber-attack likelihood by adding verification layers beyond passwords. Even if an attacker obtains a password, they cannot access the account without the additional factor(s).

## Supported MFA Factors

Auth0 supports multiple authentication methods categorized as independent and dependent factors.

### Independent Factors

At least one independent factor must be enabled. These can be used alone for MFA.

WebAuthn with FIDO Security Keys:
- Physical security keys (YubiKey, etc.)
- FIDO2/U2F standard compliance
- Phishing-resistant authentication

One-Time Password (OTP/TOTP):
- Time-based one-time passwords
- Compatible with authenticator apps
- Google Authenticator, Authy, etc.

Push Notifications (Auth0 Guardian):
- Mobile app push notifications
- One-tap approval/denial
- Rich notification with context

Phone Message:
- SMS verification codes
- Voice call verification
- Fallback for non-smartphone users

Cisco Duo Security:
- Enterprise Duo integration
- Existing Duo infrastructure support
- Unified security platform

### Dependent Factors

Require an independent factor to be configured first.

WebAuthn with Device Biometrics:
- Face ID, Touch ID, Windows Hello
- Device-bound authentication
- Convenient for enrolled devices

Email Verification:
- One-time codes via email
- Backup verification method
- Works across all devices

Recovery Codes:
- Pre-generated backup codes
- Use when primary factors unavailable
- One-time use per code

## MFA Policies

Configure when MFA is required.

Never:
- MFA not required
- Users can optionally enroll
- Lowest security, highest convenience

Always:
- MFA required for every login
- All users must complete MFA
- Highest security, more friction

Use Adaptive MFA (Enterprise):
- Risk-based MFA challenges
- Only challenges when risk detected
- Balance of security and convenience

## Configuration

### Dashboard Navigation

Access: Dashboard > Security > Multi-factor Auth

### Basic Setup

1. Navigate to MFA settings
2. Enable desired factors in Factors section
3. Select MFA policy
4. Configure additional settings

### Additional Settings

Show Multi-factor Authentication Options:
- Lets users select from enabled factors during enrollment
- Provides factor choice flexibility

Customize MFA Factors using Actions:
- Create personalized MFA flows via post-login Actions
- Implement custom logic for factor selection
- Challenge with specific factor sequences

### Factor Enrollment

User Enrollment Flow:
- First-time MFA triggers enrollment
- User selects from available factors
- Completes factor-specific setup
- Future logins use enrolled factor

Administrative Enrollment:
- Pre-enroll users via Management API
- Import existing MFA enrollments
- Bulk enrollment for organizations

## Implementation Approaches

### Universal Login MFA

Recommended approach using Auth0-hosted login:
- Automatic MFA integration
- Consistent user experience
- No custom UI development required

### Custom MFA with Actions

Use post-login Actions for custom logic:
- Conditional MFA based on user attributes
- Geographic or device-based challenges
- Custom factor sequencing
- Integration with external risk systems

### Embedded MFA

For custom applications:
- Auth0.js SDK integration
- Custom UI implementation
- Direct API calls for MFA operations

## Plan Requirements

Professional Plan:
- Standard MFA factors
- Basic policy configuration

Enterprise Plan:
- All MFA factors
- Adaptive MFA
- Guardian customization
- Advanced Actions support

## Best Practices

Factor Selection:
- Enable multiple factors for user choice
- Provide recovery codes as backup
- Consider user device capabilities

User Experience:
- Clear enrollment instructions
- Factor-specific guidance
- Fallback options when factors fail

Security Balance:
- Use Adaptive MFA when possible
- Avoid MFA fatigue with smart policies
- Regularly audit enrolled factors

Deployment:
- Pilot with subset of users
- Gradual rollout with communication
- Monitor enrollment and success rates
- Provide support resources
