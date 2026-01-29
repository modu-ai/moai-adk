# Step-Up Authentication

Step-up authentication requires users to authenticate with stronger credentials when accessing sensitive resources, adding security without impacting the entire user experience.

## Concept

Step-up authentication allows applications to:
- Grant initial access with standard authentication
- Require additional verification for sensitive operations
- Dynamically elevate authentication level
- Protect high-risk transactions

## Use Cases

Financial Applications:
- View account balance: Standard login
- Transfer funds: Require MFA
- Change beneficiary: Require MFA + verification

Healthcare Applications:
- View appointments: Standard login
- Access medical records: Require MFA
- Download prescriptions: Require MFA

E-commerce Applications:
- Browse products: No authentication
- View order history: Standard login
- Change payment method: Require MFA

Administrative Applications:
- View dashboard: Standard login
- Modify user permissions: Require MFA
- Access audit logs: Require MFA

## Implementation Approaches

### API-Based Step-Up (Scopes)

For applications with API backends:

Mechanism:
- Map sensitive operations to specific scopes
- Include scope in access token requests
- API validates scope presence
- Trigger MFA when scope requires elevation

Flow:
1. User performs standard login
2. Access token contains basic scopes
3. User attempts sensitive operation
4. Application requests elevated scope
5. Auth0 triggers MFA challenge
6. User completes MFA
7. New access token contains elevated scope
8. API authorizes sensitive operation

Scope Examples:
- read:balance (standard)
- transfer:funds (requires MFA)
- admin:users (requires MFA)

### Web Application Step-Up (Token Claims)

For traditional web applications:

Mechanism:
- Verify authentication level through ID token claims
- Check for MFA completion in token
- Redirect to MFA if not present
- Grant access after verification

Claims to Check:
- acr (Authentication Context Class Reference)
- amr (Authentication Methods Reference)
- Custom claims set by Actions

Flow:
1. User performs standard login
2. ID token contains authentication claims
3. User navigates to sensitive page
4. Application checks token claims
5. If MFA not present, redirect to re-authentication
6. Auth0 prompts for MFA
7. New ID token contains MFA claims
8. Application grants access

## Implementation with Actions

### Triggering Step-Up

Use post-login Actions to enforce MFA for specific conditions:

Condition Examples:
- Specific scope requested
- Sensitive application accessed
- High-risk operation detected
- Elevated privilege requested

Action Logic:
- Check requested scopes
- Evaluate risk context
- Challenge with MFA if needed
- Add custom claims to tokens

### Custom Claims

Add claims indicating authentication strength:
- mfa_completed: boolean
- auth_level: numeric
- auth_methods: array

These claims enable applications to verify authentication status without additional API calls.

## Token Validation

### Access Token Validation

For API step-up:
- Validate token signature
- Check scope claims
- Verify audience
- Confirm token freshness

Scope Verification:
- Extract scope claim
- Check for required scope
- Deny if scope missing
- Consider scope hierarchies

### ID Token Validation

For web app step-up:
- Validate token signature
- Check authentication claims
- Verify token freshness
- Confirm claim values

Claim Verification:
- Extract acr/amr claims
- Check for MFA indicators
- Verify claim currency
- Deny if requirements not met

## Freshness Requirements

Token Age Considerations:
- Step-up may require fresh authentication
- Stale tokens may not reflect current context
- Consider max_age parameter for re-authentication

Implementing Freshness:
- Check iat (issued at) claim
- Require token issued within threshold
- Force re-authentication if too old
- Balance security with user experience

## User Experience

Seamless Step-Up:
- Clear explanation of why additional verification needed
- Quick MFA completion
- Return to original context after verification
- Remember step-up for session duration

Error Handling:
- Clear messages for MFA failures
- Fallback factor options
- Support contact information
- Graceful degradation

Session Management:
- Track step-up status in session
- Appropriate timeout for elevated sessions
- Clear elevation on logout
- Optional elevation expiry

## Security Considerations

Transaction Binding:
- Bind MFA to specific transaction
- Display transaction details during approval
- Prevent transaction manipulation
- Log transaction context

Rate Limiting:
- Limit step-up attempts
- Prevent MFA fatigue attacks
- Monitor unusual patterns
- Alert on suspicious activity

Scope Escalation Prevention:
- Validate scope transitions
- Prevent unauthorized elevation
- Audit scope requests
- Monitor privilege changes

## Best Practices

Scope Design:
- Clear scope hierarchy
- Consistent naming convention
- Documented scope requirements
- Regular scope review

User Communication:
- Explain step-up requirement
- Provide context for verification
- Offer help resources
- Consistent messaging

Implementation:
- Server-side enforcement
- Never trust client-only checks
- Comprehensive logging
- Regular security review

Testing:
- Test all step-up scenarios
- Verify scope enforcement
- Check error handling
- Validate token claims

## Example Scenarios

### Banking Step-Up

Initial Login:
- User logs in with password
- Receives basic access token
- Can view balances and statements

Fund Transfer:
- User initiates transfer
- Application requests transfer:funds scope
- Auth0 challenges with MFA
- User approves via Guardian
- Transfer completes

### Admin Console Step-Up

Initial Access:
- Admin logs in with SSO
- Can view dashboard and reports
- Basic admin privileges

User Management:
- Admin accesses user management
- System checks ID token claims
- MFA not present, redirects to step-up
- Admin completes MFA
- User management access granted
