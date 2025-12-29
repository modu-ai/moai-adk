# WebAuthn and FIDO Implementation

WebAuthn provides phishing-resistant authentication using security keys and device biometrics through the FIDO2 standard.

## Overview

WebAuthn (Web Authentication) is a W3C standard for passwordless and multi-factor authentication. Combined with FIDO2 (CTAP2), it enables secure, phishing-resistant authentication using cryptographic credentials.

## Authentication Types

### Security Keys

Physical devices providing hardware-backed authentication.

Types of Security Keys:
- USB keys (plug into computer)
- NFC keys (tap on device)
- Bluetooth keys (wireless connection)
- Hybrid (multiple connection types)

Popular Security Key Vendors:
- Yubico (YubiKey series)
- Google (Titan Security Key)
- Feitian
- SoloKeys
- Nitrokey

### Device Biometrics

Platform authenticators built into devices.

Supported Platforms:
- Face ID (iOS devices)
- Touch ID (Apple devices)
- Windows Hello (Windows 10/11)
- Android Fingerprint/Face
- Chrome profile with biometrics

## User Verification

WebAuthn supports different verification levels.

Without User Verification:
- Presence check only (touch the key)
- Faster authentication
- Lower security assurance

With User Verification:
- Requires PIN, biometric, or passcode
- Proves both possession and identity
- Stronger security assurance

Configuring PIN for Security Keys:
- Enable user verification requirement in Auth0
- User sets PIN during first use
- PIN stored only on security key

## Benefits

Phishing Resistance:
- Credentials bound to specific domain
- Cannot be tricked into authenticating to fake site
- Origin validation built into protocol

No Shared Secrets:
- Private key never leaves device
- Only public key stored on server
- No credential database to breach

Multi-Factor in Single Step:
- Combines possession (device) with user verification
- Achieves MFA without separate steps
- Improved user experience

Passwordless Potential:
- Can replace passwords entirely
- First-factor authentication support
- Future-ready authentication

## Configuration in Auth0

### Enabling WebAuthn

1. Navigate to Dashboard > Security > Multi-factor Auth
2. Enable WebAuthn with Security Keys for hardware keys
3. Enable WebAuthn with Device Biometrics for platform authenticators

### User Verification Settings

Configure verification requirements:
- Discouraged: No user verification
- Preferred: User verification if available
- Required: User verification mandatory

Recommendation: Use Required for high-security applications.

### Attestation Settings

Attestation provides device metadata:
- None: No attestation requested
- Indirect: Anonymized attestation
- Direct: Full device attestation

Enterprise Considerations:
- Direct attestation for device policy enforcement
- Validate specific device types
- Requires attestation certificate management

## Enrollment Flow

### Security Key Enrollment

User Steps:
1. Initiate enrollment in MFA settings
2. Browser prompts for security key
3. Insert or tap security key
4. Enter PIN if required
5. Complete biometric if supported
6. Enrollment confirmed

### Device Biometrics Enrollment

User Steps:
1. Initiate enrollment in MFA settings
2. Browser prompts for biometric
3. Perform Face ID, Touch ID, or Windows Hello
4. Enrollment confirmed

Note: Device biometrics requires independent factor enrolled first.

## Authentication Flow

User Steps:
1. Enter username/password (or passwordless)
2. Prompted for WebAuthn authentication
3. Insert security key or use biometric
4. Complete user verification if required
5. Authentication successful

Behind the Scenes:
1. Server sends challenge
2. Authenticator signs challenge with private key
3. Browser sends assertion to server
4. Server validates signature with stored public key

## Cross-Device Authentication

Hybrid Transport (FIDO2.1):
- Use phone as authenticator for computer login
- QR code or Bluetooth connection
- Supports roaming between devices

Passkeys:
- Syncable WebAuthn credentials
- Available across user's devices
- iCloud Keychain, Google Password Manager support

## Browser Support

Current Support:
- Chrome (desktop and mobile)
- Firefox
- Safari
- Edge
- Opera

Check webauthn.me for current browser compatibility.

## Implementation Considerations

Fallback Strategies:
- Always provide alternative factor
- Handle unsupported browsers gracefully
- Clear messaging for users

Account Recovery:
- Recovery codes essential
- Administrator reset capability
- Clear recovery procedures

Multiple Authenticators:
- Allow enrolling multiple keys
- Backup security key recommended
- Mix of security key and biometric

Enterprise Deployment:
- Consider key provisioning
- Establish replacement procedures
- Train users on proper usage

## Security Best Practices

Enable User Verification:
- Require PIN or biometric
- Prevents unauthorized use of stolen key
- Meets multi-factor requirement

Require Attestation (Enterprise):
- Validate device types
- Enforce security key policy
- Maintain approved device list

Monitor Enrollments:
- Track WebAuthn registrations
- Alert on suspicious patterns
- Regular enrollment audits

Key Lifecycle Management:
- User-initiated key removal
- Admin key revocation
- Regular key rotation encouragement

## Troubleshooting

Common Issues:

Browser Not Prompting:
- Check browser compatibility
- Verify HTTPS connection
- Review browser security settings

Security Key Not Recognized:
- Verify USB/NFC connection
- Check key firmware version
- Try different USB port

Biometric Failure:
- Re-enroll biometric on device
- Check platform authenticator settings
- Verify browser permissions

PIN Errors:
- Reset PIN on security key
- Check PIN retry counter
- Consider key replacement if locked
