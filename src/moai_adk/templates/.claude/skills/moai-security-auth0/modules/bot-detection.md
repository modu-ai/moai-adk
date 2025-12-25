# Bot Detection

Auth0 Bot Detection mitigates scripted attacks by identifying requests likely originating from bots using traffic pattern analysis and IP reputation data.

## How It Works

Auth0 uses large amounts of data and statistical models to identify patterns signaling when bursts of login, signup, or password reset traffic are likely from a bot or script. When detected, the system triggers authentication challenges.

## Configuration

### Dashboard Navigation

Access: Dashboard > Security > Attack Protection > Bot Detection

### Detection Sensitivity

Three risk levels available:

Low Sensitivity:
- Fewer users challenged
- May miss sophisticated bots
- Best for high-friction tolerance applications

Medium Sensitivity (Default):
- Balanced detection approach
- Recommended for most applications
- Good trade-off between security and UX

High Sensitivity:
- Maximum bot detection
- More legitimate users may be challenged
- Best for high-security applications

### Response Types

Auth Challenge (Recommended):
- CAPTCHA-free verification requiring JavaScript execution
- Minimal user friction
- Detects non-browser clients automatically

Simple CAPTCHA:
- Traditional CAPTCHA interface
- Works in non-JavaScript environments
- Higher user friction but more accessible

Third-Party Integration:
- reCAPTCHA integration available
- Other CAPTCHA providers supported
- Configure via Authentication API

### CAPTCHA Trigger Modes

Configure when CAPTCHA displays:
- Never: Disable challenges entirely
- When Risky: Challenge based on detection level (recommended)
- Always: Challenge every request

### IP AllowList

Supports up to 100 entries:
- Discrete IP addresses
- CIDR range notation
- Useful for trusted office networks
- Prevents blocking of known-good sources

## Supported Flows

Fully Supported:
- Auth0 Universal Login (recommended)
- Classic Login
- Lock.js v12.4.0 and later
- Auth0.swift 1.28.0 and later (iOS)
- Auth0.Android 1.25.0 and later

Not Supported:
- Enterprise connections (SAML, OIDC, AD/LDAP)
- Social login providers
- Cross-origin authentication flows

## Signup vs Login Detection

Auth0 uses distinct detection models for signup and login flows:

Login Detection:
- Focuses on credential stuffing patterns
- Analyzes failed authentication velocity
- Considers account targeting patterns

Signup Detection:
- Addresses automated account creation
- Analyzes registration velocity
- Requires updated library versions:
  - Auth0.js 9.28.0+
  - Lock 13.0+

## Monitoring Mode

Enable bot detection without response settings to record risk assessment details in tenant logs without enforcing actions. Useful for:
- Baseline traffic analysis
- False positive assessment
- Threshold calibration
- Pre-deployment validation

## Tenant Log Events

Bot detection events appear in tenant logs with:
- Risk assessment scores
- IP reputation data
- Device fingerprinting results
- Challenge outcomes

## Implementation Considerations

User Experience:
- Auth Challenge has minimal friction
- Consider fallback for JavaScript-disabled users
- Test challenge flows thoroughly

Performance:
- Minimal latency impact
- Client-side challenge execution
- No server-side processing delay

Integration:
- Works with Universal Login out-of-box
- Custom UI requires Auth0.js integration
- Native apps need SDK updates

## Troubleshooting

False Positives:
- Add IP to AllowList if consistent
- Lower sensitivity level
- Review user agent patterns

Bots Not Detected:
- Increase sensitivity level
- Enable Always challenge mode temporarily
- Review traffic patterns in logs

Challenge Failures:
- Verify JavaScript execution environment
- Check third-party CAPTCHA configuration
- Test network connectivity to Auth0
