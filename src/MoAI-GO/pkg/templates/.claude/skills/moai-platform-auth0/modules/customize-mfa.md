# Customize MFA

Module: moai-platform-auth0/modules/customize-mfa.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

Auth0 provides several methods to customize Multi-Factor Authentication experiences for users through Universal Login branding options, programmatic controls, and the Actions framework.

---

## Universal Login Branding

### Dashboard Configuration

Navigate to Dashboard then Branding then Universal Login to access MFA page customization options.

Customizable Elements:

- Logo and company branding
- Color schemes and themes
- Font selections
- Button styles
- Page layouts

### HTML Customization

For complete control, customize the full HTML content of MFA pages.

MFA Widget Theme Options: Auth0 provides theming capabilities including language dictionaries and visual customization options.

When to Use: Organizations requiring branded MFA experiences that match their application design.

---

## API-Based Configuration

### Enable MFA Grant Type

To use the MFA API, enable the MFA grant type:

Step 1: Navigate to Dashboard then Applications.

Step 2: Select your application.

Step 3: Open Advanced Settings.

Step 4: Enable MFA under Grant Types.

Step 5: Save changes.

### Supported Scenarios

Authenticating Users: Use Resource Owner Password Grant flow with MFA challenges.

Factor Management: Allow users to manage their own authentication factors.

Custom Enrollment: Create enrollment tickets to invite users to set up MFA.

---

## Programmatic MFA Control with Actions

### Actions Framework

Use Auth0 Actions to customize MFA policy based on various conditions.

### Conditional MFA Triggers

Application-Specific MFA: Require MFA for specific applications only.

User Group Targeting: Apply MFA based on user metadata or group membership.

IP-Based Requirements: Enforce MFA for authentication attempts from specific IP ranges or unknown locations.

Risk-Based MFA: Trigger MFA based on risk signals and context.

### Remember Browser Configuration

The allowRememberBrowser property controls MFA prompt frequency:

Enabled: Users can choose to trust their browser and skip MFA for subsequent logins.

Disabled: Users must complete MFA on every authentication.

Configuration: Set through Actions or tenant settings.

---

## Provider Configuration

### Supported Factors

The system supports multiple factors through the "any" provider setting:

- Push notifications (Guardian app)
- SMS verification
- Voice call verification
- One-time passwords (TOTP)
- Email verification
- WebAuthn security keys

### Factor Selection

New Universal Login: Full support for all factor types.

Classic Login: Some limitations on factor combinations.

---

## MFA API Limitations

Supported Factors: The MFA API works with SMS, push notifications (Guardian), email, and OTP factors.

Not Supported: The MFA API does not support enrolling with Duo Security. Duo enrollment must be done through the Universal Login flow.

---

## Customization Best Practices

### Branding Consistency

Ensure MFA pages match your application's overall design.

Use consistent logos, colors, and fonts across all authentication screens.

Provide clear instructions in the user's language.

### User Experience

Minimize friction while maintaining security.

Provide clear error messages and recovery options.

Offer multiple factor options when possible.

Consider accessibility requirements.

### Security Considerations

Balance convenience features (like remember browser) with security requirements.

Implement appropriate session timeouts.

Log MFA events for security monitoring.

---

## Language Customization

### Language Dictionaries

Auth0 supports multiple languages for MFA prompts and messages.

Configuration: Set language preferences in Universal Login settings or dynamically based on user preferences.

Custom Text: Override default text with organization-specific wording.

---

## Related Modules

- mfa-overview.md: MFA configuration basics
- mfa-factors.md: Factor types and setup
- guardian-configuration.md: Guardian app customization
- adaptive-mfa.md: Risk-based MFA policies

---

## Resources

Auth0 Documentation: Customize MFA
Auth0 Documentation: Universal Login Branding
Auth0 Documentation: Actions
