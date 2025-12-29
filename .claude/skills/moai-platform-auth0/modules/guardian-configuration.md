# Auth0 Guardian Configuration

Auth0 Guardian is a mobile application for iOS and Android enabling multi-factor authentication through push notifications and one-time passwords.

## Overview

Guardian provides two primary authentication methods:
- Push Notifications: Users receive push notifications and approve with one tap
- One-Time Passwords (OTP): 6-digit codes valid for 30 seconds

## Mobile Application

### Supported Platforms

iOS:
- Available on App Store
- Supports Face ID and Touch ID
- Passcode protection available

Android:
- Available on Google Play Store
- Fingerprint authentication support
- PIN protection available

### App Features

Security Settings:
- 6-digit PIN protection
- Biometric authentication (Face ID, Touch ID, fingerprint)
- Secure storage of enrollment data

Language Support:
- 50+ languages and dialects
- User-selectable language preference
- Localized notifications

Customization:
- Custom application name display
- Branded accent colors
- Custom app icons

## Push Notification Configuration

### Notification Delivery Services

AWS Simple Notification Service (SNS):
- Configure AWS credentials
- Set up SNS topics
- Handle delivery to both platforms

Firebase Cloud Messaging (FCM):
- Android notification delivery
- Configure Firebase project
- Server key configuration

Apple Push Notification (APN):
- iOS notification delivery
- Certificate or key-based authentication
- Configure bundle identifier

### Enrollment Flow

1. User authenticates with primary credentials
2. Universal Login presents QR code
3. User scans QR code with Guardian app
4. App registers with Auth0
5. Enrollment confirmed

### Authentication Flow

1. User authenticates with primary credentials
2. Push notification sent to enrolled device
3. User sees transaction context
4. User approves or denies
5. Auth0 receives response
6. Authentication completes

### Push Notification Content

Notification includes:
- Application name
- Transaction context
- User location (if available)
- Browser/device information
- Approve/Deny options

## One-Time Password (OTP)

### When to Use OTP

OTP serves as:
- Fallback when push fails
- Primary factor when configured
- Offline authentication option

### OTP Characteristics

Code Format:
- 6 digits
- 30-second validity
- Automatic regeneration

Compatibility:
- Standard TOTP algorithm
- RFC 6238 compliant
- Works offline

### Enabling OTP

1. Enable One-Time Password as MFA factor
2. Users enroll via Guardian app
3. OTP available alongside push notifications

## Guardian SDK

### Purpose

Embed Guardian capabilities in custom branded applications instead of using Auth0 Guardian app.

### Available SDKs

iOS SDK:
- Swift/Objective-C support
- Full Guardian functionality
- Custom UI implementation

Android SDK:
- Java/Kotlin support
- Full Guardian functionality
- Custom UI implementation

### SDK Features

Enrollment:
- Programmatic QR code scanning
- Manual enrollment entry
- Custom enrollment UI

Authentication:
- Push notification handling
- OTP generation
- Custom approval UI

Management:
- Device management
- Enrollment deletion
- Token refresh

## Configuration in Dashboard

### Enable Guardian

1. Navigate to Dashboard > Security > Multi-factor Auth
2. Enable Push Notifications factor
3. Configure notification service

### Customize Guardian

Application Appearance:
- Set application name
- Configure accent color
- Upload custom icon

Notification Settings:
- Configure push service credentials
- Set notification timeouts
- Enable/disable features

## Best Practices

### Deployment

Start with Auth0 Guardian App:
- Fastest deployment
- No development required
- Automatic updates

Consider Guardian SDK When:
- Brand consistency critical
- Custom UX requirements
- Integration with existing app

### User Enrollment

Clear Instructions:
- Provide visual enrollment guide
- Support multiple languages
- Include troubleshooting steps

Multiple Device Support:
- Allow enrollment of multiple devices
- Provide device management
- Handle device replacement

### Fallback Options

Always Configure Fallbacks:
- OTP for offline scenarios
- SMS for non-smartphone users
- Recovery codes for emergencies

### Security

Enforce App Security:
- Require PIN or biometric on app
- Educate users on settings
- Monitor enrollment patterns

## Troubleshooting

### Push Notifications Not Received

Possible Causes:
- Network connectivity issues
- Notification permissions disabled
- Incorrect push service configuration
- Device in Do Not Disturb mode

Solutions:
- Verify network connection
- Check notification permissions
- Review push service configuration
- Test with OTP fallback

### Enrollment Failures

Possible Causes:
- QR code expired
- Camera permission denied
- Network timeout
- Clock synchronization issues

Solutions:
- Generate new QR code
- Grant camera permissions
- Check network connectivity
- Sync device time

### OTP Not Working

Possible Causes:
- Device clock out of sync
- Wrong account selected
- Code expired during entry
- Enrollment corrupted

Solutions:
- Sync device time automatically
- Verify correct account
- Enter code immediately after generation
- Re-enroll if persistent

## Metrics and Monitoring

### Available Metrics

Track in Auth0 Dashboard:
- Push notification delivery rates
- Approval/denial rates
- OTP usage frequency
- Enrollment counts

### Log Events

Guardian events logged:
- Enrollment success/failure
- Push sent/delivered
- Authentication approve/deny
- OTP verification attempts
