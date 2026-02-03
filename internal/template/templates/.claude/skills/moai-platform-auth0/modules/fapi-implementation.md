# FAPI Implementation

Financial-grade API (FAPI) is a set of security and privacy specifications developed by the OpenID Foundation for robust authentication and authorization in financial services and other high-security scenarios.

## Overview

FAPI provides enhanced OAuth 2.0 and OpenID Connect profiles designed for:
- Financial services
- Open banking
- Healthcare
- Government services
- Any high-security application

## Auth0 FAPI Certification

Auth0 is certified for:
- FAPI 1.0 Advanced OP (OpenID Provider)
- mTLS client authentication profile
- Private Key JWT client authentication profile

## FAPI Security Profiles

### FAPI 1.0 Baseline

Minimum security requirements for read-only access:
- OAuth 2.0 authorization code flow
- PKCE required
- State parameter required
- Confidential clients

### FAPI 1.0 Advanced

Enhanced security for read-write access:
- All baseline requirements
- Pushed Authorization Requests (PAR)
- JWT-secured Authorization Requests (JAR)
- mTLS or Private Key JWT client authentication
- Sender-constrained tokens

## Core Security Features

### Strong Customer Authentication (SCA)

Requirement: At least two independent authentication factors.

Factor Categories:
- Something known (password, PIN)
- Something possessed (device, token)
- Something inherent (biometric)

Auth0 Implementation:
- Multi-factor authentication
- WebAuthn support
- Push notifications
- SMS/Voice verification

### Dynamic Linking

Purpose: Bind authorization to specific transaction details.

Implementation:
- Rich Authorization Requests (RAR)
- Transaction details in authorization request
- User verifies transaction during authorization
- Authorization uniquely linked to transaction

User Experience:
- See transaction details
- Confirm specific action
- Authentication bound to transaction

### Pushed Authorization Requests (PAR)

Purpose: Secure transmission of authorization parameters.

How It Works:
1. Client sends parameters to PAR endpoint
2. Auth0 returns request_uri
3. Client redirects with request_uri only
4. Sensitive parameters never in browser

Benefits:
- Parameters not exposed in URL
- Reduced risk of manipulation
- Signed request verification

### JWT-Secured Authorization Requests (JAR)

Purpose: Protect authorization request integrity and confidentiality.

How It Works:
- Authorization parameters in signed JWT
- Optionally encrypted
- Prevents tampering

Benefits:
- Request integrity
- Optional confidentiality
- Signed by client

### JSON Web Encryption (JWE)

Purpose: Encrypt access token payloads containing sensitive authorization details.

Use Cases:
- Rich authorization data
- Sensitive permissions
- Transaction details in tokens

## Client Authentication

### Private Key JWT

Asymmetric authentication using signed JWTs.

Features:
- Private key never transmitted
- Short-lived signed assertions
- No shared secret

Requirements:
- Enterprise plan
- Register public key with Auth0
- Sign client_assertion with private key

### mTLS for OAuth

Mutual TLS client authentication.

Features:
- Certificate-based authentication
- Transport-layer security
- Strong client identity

Requirements:
- Enterprise plan with HRI add-on
- Register client certificate
- mTLS infrastructure

## Token Security

### Sender Constraining

Bind tokens to client:
- DPoP for application-layer binding
- mTLS for transport-layer binding
- Prevents token theft

### Token Binding

Certificate thumbprint in tokens:
- cnf claim with x5t#S256
- Validates client certificate
- Ensures token only used by legitimate client

## Implementation Requirements

### Plan Requirements

Minimum: Enterprise Plan

For Full FAPI:
- Highly Regulated Identity add-on
- Required for mTLS
- Required for advanced features

### Configuration Steps

1. Enable HRI Features:
   - Contact Auth0
   - Enable add-on
   - Configure tenant

2. Configure Client Authentication:
   - Choose Private Key JWT or mTLS
   - Register credentials
   - Configure application

3. Enable PAR:
   - Configure PAR endpoint
   - Update client to use PAR
   - Test request flow

4. Configure Token Binding:
   - Enable sender constraining
   - Choose DPoP or mTLS
   - Configure resource servers

### Application Changes

Client Requirements:
- Support PAR flow
- Implement JAR if required
- Handle sender-constrained tokens
- Proper error handling

Resource Server Requirements:
- Validate sender-constrained tokens
- Verify token binding
- Handle FAPI token types

## Best Practices

### Credential Management

Private Keys:
- Secure generation
- Protected storage
- Regular rotation
- Zero-downtime rotation support

Certificates:
- Proper CA hierarchy
- Certificate lifecycle management
- Revocation handling

### Security Configuration

Enable All Features:
- PAR for authorization
- JAR for request signing
- Sender constraining for tokens
- Strong client authentication

Monitor and Audit:
- Log all FAPI transactions
- Monitor for anomalies
- Regular security review

### Testing

Conformance Testing:
- Use FAPI conformance suite
- Test all flows
- Verify error handling
- Document test results

## Regulatory Context

### Open Banking

UK/EU Open Banking:
- FAPI profiles mandated
- PSD2 alignment
- Strong customer authentication

### Financial Services

Global Standards:
- Financial sector requirements
- Regulatory compliance
- Security best practices

### Healthcare

SMART on FHIR:
- FAPI-aligned security
- Healthcare data protection
- Patient consent management
