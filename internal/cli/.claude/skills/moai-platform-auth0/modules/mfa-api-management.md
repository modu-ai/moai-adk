# MFA API Management

Manage multi-factor authentication programmatically using Auth0 Management API and Authentication API for enrollments, factors, and user MFA settings.

## Management API Operations

### List User Enrollments

Retrieve all MFA enrollments for a user.

Endpoint: GET /api/v2/users/{user_id}/enrollments

Response includes:
- Enrollment ID
- Factor type
- Status (confirmed, pending)
- Enrollment date
- Device information

### Delete User Enrollment

Remove a specific MFA enrollment.

Endpoint: DELETE /api/v2/users/{user_id}/enrollments/{enrollment_id}

Use cases:
- User lost device
- Factor replacement
- Security incident response

### Reset User MFA

Remove all MFA enrollments for a user.

Endpoint: DELETE /api/v2/users/{user_id}/authenticators

Effect:
- Removes all enrolled factors
- User must re-enroll on next MFA challenge
- Recovery codes invalidated

### Generate Recovery Codes

Create new recovery codes for a user.

Endpoint: POST /api/v2/users/{user_id}/recovery-code-regeneration

Response:
- New recovery codes array
- Previous codes invalidated
- User must store new codes

## Authentication API Operations

### MFA Challenge

Initiate MFA challenge during authentication.

Endpoint: POST /oauth/token (with mfa_token)

Parameters:
- mfa_token: Token from initial authentication
- challenge_type: Factor type to challenge
- authenticator_id: Specific enrollment (optional)

### MFA OOB Challenge

Initiate out-of-band challenge (push, SMS).

Endpoint: POST /mfa/challenge

Parameters:
- mfa_token: MFA session token
- challenge_type: oob
- authenticator_id: Enrollment ID
- oob_channel: push, sms, or voice

### Verify MFA

Complete MFA verification.

Endpoint: POST /oauth/token

Parameters:
- grant_type: mfa-oob or mfa-otp
- mfa_token: MFA session token
- otp: One-time password (for TOTP)
- oob_code: Out-of-band code (for SMS)
- binding_code: Push notification code

## Enrollment Management

### Programmatic Enrollment

Enroll user in MFA factor via API.

Steps:
1. Get enrollment ticket via Management API
2. Generate enrollment data (QR, secret)
3. Present to user for enrollment
4. Confirm enrollment

### Enrollment Verification

Confirm pending enrollment.

Endpoint: POST /mfa/associate

Parameters:
- mfa_token: Association token
- otp: Verification code from new factor

### List Available Factors

Get configured MFA factors for tenant.

Endpoint: GET /api/v2/guardian/factors

Response includes:
- Factor type
- Enabled status
- Configuration details

## Factor-Specific APIs

### Guardian (Push)

Send push notification:
- Endpoint handles notification delivery
- Response includes challenge ID
- Poll or webhook for response

Guardian enrollment:
- Generate enrollment ticket
- Create QR code for app scanning
- Confirm via app acknowledgment

### SMS/Voice

Send verification code:
- Endpoint triggers message delivery
- Code valid for limited time
- Rate limiting applies

### TOTP

Generate secret:
- Create TOTP secret for enrollment
- Encode as QR code or manual entry
- Verify initial OTP to confirm

### WebAuthn

Create credential options:
- Generate challenge
- Define allowed authenticators
- Set user verification requirement

Verify credential:
- Validate authenticator assertion
- Confirm credential binding
- Store public key

## User Self-Service

### Enable User MFA Management

Allow users to manage their own MFA:
- View enrolled factors
- Add new factors
- Remove factors
- Generate recovery codes

### Implementation

User MFA Portal:
- Build custom UI or use Auth0 dashboard
- Call Management API with user token
- Implement proper authorization
- Audit user actions

## Administrative Operations

### Bulk MFA Reset

Reset MFA for multiple users:
- Export affected user list
- Iterate with Management API
- Log reset actions
- Communicate to users

### MFA Enforcement

Require MFA enrollment:
- Use Rules or Actions
- Check enrollment status
- Redirect to enrollment if missing
- Allow grace period if needed

### Audit MFA Events

Track MFA-related activities:
- Enrollment events
- Authentication events
- Reset events
- Failure events

## Error Handling

### Common Errors

mfa_required:
- User needs to complete MFA
- Provide mfa_token for challenge flow
- Redirect to MFA flow

invalid_otp:
- OTP verification failed
- May be expired or incorrect
- Allow retry with rate limiting

enrollment_not_found:
- Requested enrollment does not exist
- May be deleted or invalid ID
- Handle gracefully

rate_limited:
- Too many MFA attempts
- Implement backoff
- Inform user of wait time

### Error Response Handling

Implement proper error handling:
- Parse error codes
- Display user-friendly messages
- Log for debugging
- Alert on suspicious patterns

## Security Considerations

API Access:
- Use appropriate API permissions
- Implement rate limiting
- Audit API usage
- Rotate API credentials

Token Handling:
- MFA tokens are short-lived
- Do not log sensitive tokens
- Secure token storage
- Implement proper expiration

User Authorization:
- Verify user identity before MFA changes
- Require current authentication
- Log administrative actions
- Alert on suspicious changes

## Best Practices

Implementation:
- Use official SDKs when available
- Implement proper error handling
- Test all edge cases
- Monitor API usage

User Experience:
- Clear error messages
- Helpful enrollment guidance
- Fallback options available
- Support documentation

Security:
- Audit all MFA operations
- Alert on mass resets
- Monitor for abuse patterns
- Regular security review
