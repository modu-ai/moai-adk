# Highly Regulated Identity

Auth0 Highly Regulated Identity (HRI) is a Financial-Grade Identity solution designed to secure sensitive data operations and services in regulated industries.

## Overview

HRI provides enhanced security for:
- Financial services (banks, payment processors)
- Healthcare (patient data, prescriptions)
- Government services (identity verification)
- Any high-value transaction processing

Target Operations:
- Money transfers
- Digital payments
- Medical record access
- Contract signing
- High-value authorizations

## Requirements

Plan: Enterprise Plan with Highly Regulated Identity add-on

Contact: Auth0 sales for add-on enablement

## Core Security Features

### Strong Customer Authentication (SCA)

Definition: Authentication requiring at least two independent factors from different categories.

Factor Categories:
- Knowledge: Something known (password, PIN)
- Possession: Something possessed (device, token)
- Inherence: Something intrinsic (biometric)

Supported MFA Factors:
- Mobile push notifications
- SMS verification
- Email verification
- WebAuthn (security keys, biometrics)

Dynamic Application:
- Apply based on transaction risk
- Step-up for sensitive operations
- Context-aware enforcement

### Dynamic Linking

Purpose: Bind authorization to specific transaction details so users know exactly what they authorize.

Rich Authorization Requests (RAR):
- Include transaction details in authorization
- Display details to user for confirmation
- Authorization linked to specific transaction

User Experience:
- User sees transaction context
- Confirms specific action (e.g., transfer $100)
- Cannot be repurposed for other transactions

Step-Up with Context:
- Trigger MFA with transaction details
- User verifies both identity and transaction
- Strong binding between auth and action

### Data Protection

Pushed Authorization Requests (PAR):
- Send authorization parameters directly to Auth0
- Receive reference URI
- Avoid exposing parameters in browser
- Protect sensitive data from URL exposure

JWT-Secured Authorization Requests (JAR):
- Sign authorization request as JWT
- Protect request integrity
- Optional encryption for confidentiality
- Prevent request tampering

JSON Web Encryption (JWE):
- Encrypt access token payloads
- Protect authorization details
- Confidentiality for sensitive data

### Application Authentication

Private Key JWT:
- Asymmetric key authentication
- Private key never transmitted
- Register up to two public keys
- Zero-downtime credential rotation

mTLS for OAuth:
- X.509 certificate authentication
- Mutual TLS required
- Certificate-based identity
- Strongest client authentication

Supported Algorithms:
- RS256, RS384, RS512
- PS256, PS384, PS512

### Token Binding

Certificate Thumbprint Association:
- Token bound to client certificate
- cnf claim with x5t#S256
- Prevents token theft

Sender Constraining:
- Only legitimate client can use token
- Certificate required for token use
- Useless if stolen without certificate

## Implementation

### Prerequisites

1. Enterprise Plan
2. HRI add-on enabled
3. Application configured as confidential client
4. Certificate or key pair for authentication

### Configuration Steps

1. Enable HRI:
   - Contact Auth0
   - Activate add-on on tenant
   - Verify HRI features available

2. Configure Client Authentication:
   - Generate key pair or obtain certificate
   - Register with Auth0
   - Configure application settings

3. Enable Security Features:
   - Configure PAR endpoint
   - Set up JAR signing
   - Enable token binding

4. Configure MFA:
   - Enable required factors
   - Set up step-up policies
   - Configure SCA rules

5. Test Implementation:
   - Verify all flows work
   - Test error scenarios
   - Validate security properties

### Credential Management

Key Pair Management:
- Generate secure key pairs
- Store private key securely
- Register public key with Auth0

Certificate Management:
- Obtain from trusted CA
- Register with Auth0
- Plan for rotation

Rotation:
- Register new credential
- Deploy to clients
- Remove old credential
- Up to two credentials active

### Transaction Flow

For High-Value Operations:

1. Initiate Transaction:
   - User starts sensitive operation
   - Application determines step-up needed

2. Authorization with RAR:
   - Include transaction details
   - Use PAR for security
   - Sign request with JAR

3. User Authentication:
   - SCA enforced (two factors)
   - Transaction details displayed
   - User confirms authorization

4. Token Issuance:
   - Sender-constrained token issued
   - Bound to client credential
   - Contains authorization details

5. Execute Operation:
   - Present token to API
   - Validate sender constraining
   - Complete transaction

## Security Benefits

### Against Common Attacks

Token Theft:
- Tokens bound to client
- Useless without credential
- Sender constraining protection

Request Manipulation:
- JAR ensures integrity
- Signed by client
- Tampering detected

Parameter Exposure:
- PAR keeps parameters off URL
- Server-to-server transmission
- Reduced attack surface

Phishing:
- WebAuthn resists phishing
- Transaction details visible
- User confirms actual operation

### Regulatory Compliance

PSD2 (Europe):
- SCA requirements met
- Dynamic linking supported
- Transaction authentication

Open Banking:
- FAPI-aligned
- Financial-grade security
- Regulatory compliance

## Best Practices

### Security Configuration

Enable All Features:
- Use PAR for all requests
- Sign requests with JAR
- Enable sender constraining
- Enforce SCA appropriately

Credential Security:
- HSM for private keys
- Secure certificate storage
- Regular rotation
- Access controls

### User Experience

Clear Transaction Display:
- Show what user authorizes
- Specific amounts and details
- No ambiguous permissions

Efficient MFA:
- WebAuthn for seamless experience
- Push notifications for speed
- Fallback options available

### Operations

Monitoring:
- Log all HRI transactions
- Monitor for anomalies
- Alert on failures

Testing:
- Regular security testing
- Conformance verification
- Penetration testing
