# mTLS Sender Constraining

Mutual TLS (mTLS) Sender Constraining cryptographically binds access and refresh tokens to client applications using X.509 certificates, preventing token theft and misuse at the transport layer.

## Overview

mTLS sender constraining uses mutual TLS connections where both client and server present certificates. The client certificate thumbprint is embedded in issued tokens, ensuring only the certificate holder can use those tokens.

## Requirements

Plan: Enterprise Plan with Highly Regulated Identity add-on

Client Type: Confidential clients only (not SPAs or mobile apps)

Infrastructure:
- PKI for certificate management
- mTLS termination capability
- Certificate rotation procedures

## How It Works

### Token Binding Process

1. Client Establishes mTLS Connection:
   - Client presents X.509 certificate
   - Auth0 validates certificate
   - Mutual authentication completed

2. Certificate Thumbprint Extraction:
   - Auth0 extracts client certificate
   - Computes SHA-256 hash of certificate
   - Creates base64url-encoded thumbprint

3. Token Issuance:
   - Thumbprint embedded in access token
   - Stored in cnf (confirmation) claim
   - Field name: x5t#S256

4. Token Usage:
   - Client establishes mTLS to resource server
   - Presents same certificate
   - Resource server validates binding

### Token Structure

Access token contains:

Confirmation Claim:
- cnf: Object with certificate binding
- x5t#S256: Base64url SHA-256 of certificate

Token Type:
- token_type: "DPoP" (indicates sender constraining)

Example Structure:
```
{
  "cnf": {
    "x5t#S256": "bwcK0esc3ACC3DB2Y5_lESsXE8o9ltc05O89jdN-dg2"
  }
}
```

### Resource Server Validation

When client calls API:

1. Establish mTLS Connection:
   - Client presents certificate
   - Server terminates TLS
   - Extracts client certificate

2. Extract Token:
   - Get access token from Authorization header
   - Format: Authorization: DPoP {token}

3. Validate Binding:
   - Compute SHA-256 of presented certificate
   - Base64url encode the hash
   - Compare with token's x5t#S256 value

4. Authorize if Match:
   - Thumbprints must match exactly
   - Reject if mismatch
   - Proceed with normal authorization

## Configuration

### Prerequisites

1. Highly Regulated Identity Add-on:
   - Contact Auth0 sales
   - Enterprise plan required
   - Enable HRI features

2. Client Certificate:
   - Valid X.509 certificate
   - Trusted CA or self-signed with registration
   - Proper key usage extensions

3. Application Configuration:
   - Configure as confidential client
   - Register certificate with Auth0
   - Enable mTLS authentication

### Certificate Registration

Register client certificate with Auth0:
- Upload certificate (public part)
- Associate with application
- Can register up to two certificates for rotation

### Token Request

Configure token requests:
- Establish mTLS connection to token endpoint
- Present registered certificate
- Auth0 binds token to certificate

## Certificate Management

### Certificate Requirements

Valid X.509 Certificate:
- RSA or ECDSA key
- Appropriate validity period
- Proper chain to trusted CA (or self-signed registered)

Key Usage:
- Digital signature
- Client authentication

### Certificate Rotation

Zero-Downtime Rotation:
1. Generate new certificate
2. Register new certificate (can have two active)
3. Deploy new certificate to clients
4. Remove old certificate after transition

Two Certificate Limit:
- Maximum two certificates per application
- Enables seamless rotation
- Remove old before adding third

### Certificate Storage

Private Key Protection:
- Never transmit private key
- Use HSM when possible
- Secure key storage

Certificate Distribution:
- Securely provision to clients
- Consider certificate management solution
- Audit certificate access

## Security Benefits

### Token Theft Prevention

Without Certificate:
- Attacker cannot use stolen token
- Certificate private key required
- Transport-layer binding

Compared to Bearer Tokens:
- Bearer tokens usable by anyone
- mTLS tokens bound to specific client
- Significantly stronger security

### Mutual Authentication

Both Parties Verified:
- Server proves identity via TLS
- Client proves identity via certificate
- Full mutual authentication

Trust Establishment:
- Certificate authority trust
- Explicit certificate registration
- Clear identity binding

## Comparison with DPoP

### mTLS Advantages

Transport Layer:
- Binding at TLS level
- Established PKI infrastructure
- No application-layer changes

Simpler Client Implementation:
- Certificate handling in TLS library
- No proof JWT generation
- Less application code

### mTLS Limitations

Confidential Clients Only:
- Not suitable for SPAs
- Not suitable for mobile apps
- Requires secure certificate storage

Infrastructure Requirements:
- PKI infrastructure needed
- Certificate management overhead
- mTLS termination capability

### When to Use mTLS

Choose mTLS When:
- Backend-to-backend communication
- Existing PKI infrastructure
- Confidential clients only
- Enterprise environment

Choose DPoP When:
- Public clients needed
- No PKI available
- Flexibility required

## Implementation

### Token Endpoint

Establish mTLS to Auth0:
- Configure TLS client with certificate
- Connect to token endpoint
- Auth0 extracts and binds certificate

### Resource Server

Configure mTLS termination:
- Accept client certificates
- Extract certificate from TLS session
- Validate token binding

Validation Code Logic:
1. Get client certificate from TLS context
2. Compute SHA-256 hash
3. Base64url encode
4. Extract x5t#S256 from token
5. Compare values
6. Accept or reject

### Multiple Resource Servers

Consistent Certificate:
- Use same certificate for all servers
- All tokens bound to same thumbprint
- Simplified certificate management

## Best Practices

### Certificate Management

Lifecycle Management:
- Track certificate expiration
- Automate renewal process
- Monitor certificate status

Rotation Schedule:
- Regular rotation (annual minimum)
- Emergency rotation capability
- Test rotation procedures

### Security

Private Key Protection:
- HSM when possible
- Encrypted storage
- Access controls

Certificate Validation:
- Validate certificate chain
- Check revocation status
- Verify key usage

### Operations

Monitoring:
- Track certificate usage
- Alert on expiration
- Log binding failures

Testing:
- Test mTLS connectivity
- Verify binding validation
- Test rotation procedures

## Troubleshooting

Connection Issues:

Certificate Not Presented:
- Verify TLS client configuration
- Check certificate path
- Confirm private key accessible

Certificate Rejected:
- Verify certificate registered
- Check certificate validity
- Confirm CA trust

Binding Issues:

Thumbprint Mismatch:
- Verify same certificate used
- Check certificate rotation
- Confirm computation correct

Token Rejected:
- Verify mTLS to resource server
- Check Authorization header format
- Confirm token not expired
