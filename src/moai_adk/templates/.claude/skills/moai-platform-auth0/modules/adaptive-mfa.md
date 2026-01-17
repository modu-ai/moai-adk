# Adaptive MFA

Adaptive MFA is a flexible, extensible multi-factor authentication policy that assesses potential risk during every login transaction and prompts for additional verification when appropriate.

## Requirements

Plan: Enterprise Plan with Adaptive MFA add-on

## How It Works

Adaptive MFA evaluates multiple risk signals for each login attempt, generates a confidence score, and triggers MFA challenges when the risk exceeds acceptable thresholds.

Key Characteristic: Adaptive MFA ignores any existing MFA sessions and cannot be bypassed by previous authentication.

## Risk Signals

### NewDevice

Detection: Login attempt from a device not used in the past 30 days.

Identification Method:
- User agent analysis
- Browser cookies
- Device fingerprinting

Risk Assessment:
- Unknown device increases risk
- Known device reduces risk
- Compares against historical account access

### ImpossibleTravel

Detection: Geographically suspicious login attempts.

Calculation:
- Distance between last valid location and current location
- Time elapsed between logins
- Hypothetical travel velocity

Threshold:
- Compares calculated velocity against reasonable travel speed
- Triggers when physically impossible to travel between locations
- Accounts for VPN and proxy usage patterns

### UntrustedIP

Detection: Logins from IP adddesses with suspicious activity history.

Intelligence Sources:
- Auth0 traffic intelligence
- IP reputation databases
- Historical attack association

Assessment:
- High-velocity attack association
- Known malicious IP ranges
- Tor exit nodes and proxies

## Confidence Scoring

The system combines all three risk factors to generate an overall confidence score:

High Confidence (Low Risk):
- Known device
- Normal location
- Trusted IP
- MFA not required

Low Confidence (High Risk):
- New device
- Impossible travel detected
- Untrusted IP
- MFA required

Risk Combination:
- Multiple risk factors compound
- Single high-risk factor can trigger MFA
- Weighted scoring based on signal strength

## Supported Authentication Flows

Fully Supported:
- OIDC/OAuth2 Authorization Code Flow
- SAML SP-initiated authentication
- WS-Federation
- AD/LDAP authentication

Not Supported:
- Resource Owner Password Grant (ROPG)
- Device Authorization Flow
- Refresh token flows

## Configuration

### Dashboard Settings

1. Navigate to Dashboard > Security > Multi-factor Auth
2. Set policy to Use Adaptive MFA
3. Ensure at least one MFA factor is enabled

### Customization with Actions

Create post-login Actions to customize Adaptive MFA behavior:

Custom Risk Logic:
- Add custom risk signals
- Integrate external risk services
- Implement business-specific rules

Conditional Challenges:
- Challenge based on user attributes
- Organization-specific policies
- Transaction-based challenges

Factor Selection:
- Enforce specific factors for high-risk
- Allow factor choice for medium-risk
- Skip MFA for trusted scenarios

## Integration with Custom Risk Assessment

External Risk Services:
- Pass transaction context to external API
- Receive risk score
- Combine with Auth0 Adaptive signals

Custom Signals:
- Geographic restrictions
- Time-based policies
- Device trust levels
- User behavior analytics

## User Experience

Low-Risk Login:
- Normal authentication flow
- No MFA prompt
- Seamless access

High-Risk Login:
- MFA challenge presented
- User completes additional factor
- Access granted after verification

Transparent to Users:
- No explanation of risk assessment
- Consistent MFA experience
- Standard factor enrollment

## Monitoring

### Adaptive MFA Logs

Events logged include:
- Risk assessment results
- Individual signal evaluations
- MFA challenge decisions
- Authentication outcomes

### Security Center

View Adaptive MFA metrics:
- Challenge rates over time
- Risk signal distribution
- Geographic patterns
- Device type analysis

## Best Practices

Factor Configuration:
- Enable multiple factors for user choice
- Include recovery codes
- Configure factors before enabling Adaptive

Gradual Rollout:
- Enable for subset of users first
- Monitor challenge rates
- Adjust based on feedback

Threshold Tuning:
- Review false positive rate
- Adjust risk thresholds via Actions
- Balance security with user friction

User Communication:
- Explain why MFA may be required
- Provide factor enrollment guidance
- Clear support procedures

## Comparison with Always MFA

Always MFA:
- Every login requires MFA
- Maximum security
- Highest user friction
- Simpler to understand

Adaptive MFA:
- Risk-based challenges
- Good security with less friction
- More sophisticated
- Requires Enterprise plan

Recommendation:
- Use Adaptive for consumer applications
- Consider Always for high-security admin access
- Combine with step-up for sensitive operations

## Troubleshooting

Frequent False Positives:

New Device Triggers:
- Users clearing cookies
- Private browsing mode
- Multiple browsers
- Solution: User education on device recognition

Impossible Travel:
- VPN usage
- Multiple location employees
- Solution: Custom Actions to handle known patterns

Untrusted IP:
- Corporate proxies
- Cloud-based VPNs
- Solution: IP AllowList or custom logic

MFA Not Triggering When Expected:
- Check if flow is supported
- Verify policy is set to Adaptive
- Review risk signal evaluations in logs
- Ensure MFA factors are enabled
