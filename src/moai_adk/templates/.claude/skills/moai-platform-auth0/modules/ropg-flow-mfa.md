# ROPG Flow with MFA

Module: moai-platform-auth0/modules/ropg-flow-mfa.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

The Resource Owner Password Grant (ROPG) flow allows authentication using username and password credentials. When MFA is enabled, the flow requires additional steps to enroll or challenge authenticators through the MFA API.

---

## Prerequisites

### Enable MFA Grant Type

Navigate to Dashboard then Applications then select your application.

Open Advanced Settings then Grant Types.

Enable MFA grant type and save changes.

### Supported Factors

The ROPG flow with MFA supports:

- SMS verification
- Voice call verification
- One-time passwords (TOTP)
- Push notifications (Guardian)
- Email verification
- Recovery codes

---

## Authentication Flow

### Step 1: Initial Authentication

Call the /oauth/token endpoint with user credentials.

When MFA is enabled, instead of receiving tokens, you receive an MFA challenge response:

Response includes:
- error: mfa_required
- error_description: Multifactor authentication required
- mfa_token: Token for MFA operations (valid for 10 minutes)

Token Expiry: Access tokens with the MFA audience expire in 10 minutes. This is non-configurable.

### Step 2: Retrieve Enrolled Authenticators

Use the MFA token to call the MFA Authenticators endpoint.

This returns an array of factors the user has enrolled.

If the array is empty, the user needs to enroll a factor.

Response Example Structure:
- id: Unique identifier for the authenticator
- authenticator_type: Type of factor (recovery-code, oob, etc.)
- active: Whether the factor is currently active
- oob_channel: For OOB factors, the channel (email, sms, etc.)

### Step 3: Enrollment or Challenge

If No Factors Enrolled: Use the MFA token with the MFA Associate endpoint to enroll a new factor.

If Factors Exist: Use the authenticator_id with the MFA Challenge endpoint to initiate a challenge.

### Step 4: Complete Challenge

User receives the challenge (OTP code, push notification, etc.).

Application submits the challenge response to Auth0.

### Step 5: Obtain Final Tokens

After successful challenge completion, call /oauth/token again using the MFA token.

Receive the final access token, ID token, and refresh token.

---

## MFA OTP Code Limitations

Expiry: OTP codes expire after 5 minutes. This is non-configurable.

One-Time Use: Validated codes cannot be reused.

Rate Limiting: Bucket algorithm with 10 attempts, refreshing at 1 attempt per 6 minutes.

---

## Customizable MFA Requirements

### Challenge Type Specification

The mfa_required error response includes mfa_requirements parameter.

This specifies which challenge types are supported:

- otp: One-time password from authenticator app
- push-notification: Push notification to Guardian app
- phone: SMS or voice verification
- recovery-code: Backup recovery codes

### Factor Selection

Use the mfa/authenticator endpoint to list enrolled factors matching your application's supported types.

Call request/mfa/challenge endpoint to enforce challenges for specific factors.

---

## Implementation Steps

### Initial Request

Endpoint: POST /oauth/token

Parameters:
- grant_type: password
- username: User's email or username
- password: User's password
- client_id: Application client ID
- client_secret: Application client secret (for confidential apps)
- scope: Requested scopes

### Handle MFA Required Response

Check for error code mfa_required.

Store the mfa_token for subsequent requests.

Determine next action based on user's enrolled factors.

### Enroll New Factor

Endpoint: POST /mfa/associate

Headers: Authorization Bearer with mfa_token

Body: authenticator_types array with desired factor type

### Challenge Existing Factor

Endpoint: POST /mfa/challenge

Headers: Authorization Bearer with mfa_token

Body:
- client_id: Application client ID
- challenge_type: Challenge type (otp, oob)
- authenticator_id: ID of the factor to challenge

### Complete Authentication

Endpoint: POST /oauth/token

Parameters:
- grant_type: mfa-otp or mfa-oob depending on factor
- mfa_token: The MFA token from initial response
- otp: The OTP code (for OTP challenges)
- oob_code: The OOB code (for OOB challenges)

---

## Error Handling

### Common Errors

invalid_grant: OTP code is invalid or expired.

too_many_attempts: Rate limit exceeded.

mfa_enrollment_required: User must enroll a factor before authenticating.

### Recovery

Provide clear error messages to users.

Offer alternative factors if available.

Implement retry logic with appropriate delays.

---

## Security Considerations

Secure Credential Handling: Never log or store user passwords.

Token Storage: Store MFA tokens securely during the flow.

Rate Limiting: Implement client-side rate limiting to avoid lockouts.

Timeout Handling: Handle MFA token expiration gracefully.

---

## Related Modules

- mfa-overview.md: MFA configuration
- mfa-factors.md: Factor types
- mfa-api-management.md: API operations
- customize-mfa.md: MFA customization

---

## Resources

Auth0 Documentation: ROPG Flow with MFA
Auth0 Documentation: MFA API
Auth0 Documentation: Multi-factor Authentication
