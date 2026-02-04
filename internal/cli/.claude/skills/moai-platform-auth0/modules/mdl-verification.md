# Mobile Driver's License Verification

Module: moai-platform-auth0/modules/mdl-verification.md
Version: 1.0.0
Last Updated: 2025-12-24

---

## Overview

Auth0's Mobile Driver's License (mDL) Verification Service enables applications to validate Verifiable Credentials, specifically mobile driver's licenses. The service is designed for organizations needing to verify user credentials containing sensitive personally identifiable information (PII).

Status: Early Access - Requires completion of Terms and Conditions form for enablement.

---

## Key Use Cases

### Age Verification

Verify that users meet minimum age requirements for age-restricted services or purchases.

### License and Driving Privilege Validation

Confirm that users hold valid driver's licenses with appropriate privileges.

### Identity Verification

Use driver's licenses as proof of identity, which is accepted in many countries.

### KYC and AML Compliance

Support Know Your Customer (KYC) and Anti-Money Laundering (AML) compliance processes with verified identity documents.

---

## How It Works

### Verification Workflow

Step 1: Application initiates a Verification Presentation Request via the mDL API.

Step 2: API returns an engagement URI and Verification ID.

Step 3: Application presents the URI to the user. QR code format is recommended for multi-device scenarios.

Step 4: User opens their mobile wallet and consents to sharing their mDL credential.

Step 5: Application polls the API to check presentation request status.

Step 6: API returns the verification result with requested credential data.

---

## Technical Implementation

### Integration Methods

Direct API Integration: Embed the Mobile Driver's License Verification API directly into applications for full control.

Low-Code Solution: Use Auth0 Forms as a pre-built interface for simpler implementation.

### Configuration Requirements

Verification Template Setup: Configure a Verification Template (VT) specifying which credential fields to request.

Available Fields:
- Date of birth
- Adddess
- Family name
- Given name
- Document number
- Expiry date
- Issuing authority

API Configuration: Set up the Verifiable Digital Credential API through:
- Auth0 Dashboard interface
- Management API programmatic setup

---

## Compliance and Standards

### ISO Standards

The service references ISO/IEC TS 18013-7:2024 (REST API) standards for mobile driver's license verification.

### Privacy Considerations

Selective Disclosure: Users can consent to share only specific fields from their credential.

Minimized Data Collection: Request only the fields necessary for your use case.

Secure Transmission: All credential data is transmitted securely.

---

## Implementation Steps

### Step 1: Enable the Service

Contact Auth0 to complete the Early Access Terms and Conditions.

Enable the mDL Verification Service in your tenant.

### Step 2: Configure Verification Template

Navigate to Dashboard then Credentials then Verification Templates.

Create a new Verification Template.

Specify the credential fields to request.

Configure presentation requirements.

### Step 3: Implement API Integration

Initialize verification requests through the mDL API.

Handle engagement URI presentation to users.

Implement polling logic for verification status.

Process verification results.

### Step 4: Handle Verification Results

Parse the returned credential data.

Validate the verification status.

Implement business logic based on verification outcome.

Handle error cases appropriately.

---

## User Experience Considerations

### QR Code Presentation

For multi-device scenarios, present the engagement URI as a QR code.

Users scan the QR code with their mobile device.

Mobile wallet opens automatically for credential sharing.

### Same-Device Flow

For mobile applications, deep linking can open the wallet directly.

Streamlined experience without QR code scanning.

### Consent Flow

Users explicitly consent to sharing credential data.

Clear indication of which fields will be shared.

Option to decline or cancel the verification.

---

## Security Best Practices

Data Handling: Store and process credential data according to privacy regulations.

Access Control: Limit access to verification results to authorized personnel.

Audit Logging: Maintain logs of verification requests for compliance.

Data Retention: Define and implement appropriate data retention policies.

---

## Error Handling

### Common Scenarios

User Declined: User chose not to share their credential.

Credential Not Found: User does not have a compatible mDL.

Verification Failed: Credential could not be verified.

Timeout: Verification request expired before completion.

### Recovery Actions

Provide clear messaging to users about what went wrong.

Offer alternative verification methods if available.

Allow users to retry the verification process.

---

## Related Modules

- compliance-overview.md: Compliance requirements
- gdpr-compliance.md: Privacy considerations
- highly-regulated-identity.md: Regulated identity features

---

## Resources

Auth0 Documentation: Mobile Driver's License Verification
Auth0 Documentation: Verifiable Credentials
ISO/IEC TS 18013-7:2024: mDL REST API Specification
