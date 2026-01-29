# DPoP Implementation

Demonstrating Proof-of-Possession (DPoP) is an OAuth 2.0 extension that cryptographically binds access and refresh tokens to the client application, preventing token theft and misuse.

## Overview

DPoP ensures only the client possessing the private key can use issued tokens. Unlike traditional bearer tokens where anyone with the token can use it, DPoP tokens are bound to a specific client.

## Current Status

DPoP at Auth0 is in Early Access. Contact your Auth0 representative to request access.

## How DPoP Works

### Key Pair Generation

Client generates asymmetric key pair:
- Private key: Kept secret, never transmitted
- Public key: Included in DPoP proof JWT

Recommended Algorithms:
- ES256 (Elliptic Curve): Recommended for modern applications
- RS256 (RSA): Broader compatibility

Key Storage:
- Hardware-backed keystores when available
- Secure enclave on mobile devices
- Encrypted storage for web applications

### DPoP Proof JWT

A signed JWT proving possession of the private key.

Header Claims:
- typ: Must be "dpop+jwt"
- alg: Signing algorithm (ES256, RS256)
- jwk: Public key representation

Payload Claims:
- jti: Unique identifier (prevents replay)
- htm: HTTP method of request (POST, GET)
- htu: HTTP URI of request (without fragments)
- iat: Issue timestamp
- ath: Access token hash (for API calls)
- nonce: Server-provided value (for public clients)

### Token Request

When requesting tokens:
1. Generate new DPoP proof JWT
2. Set htm to POST
3. Set htu to token endpoint
4. Sign with private key
5. Send in DPoP header

Request Headers:
- DPoP: {dpop_proof_jwt}
- Content-Type: application/x-www-form-urlencoded

### Nonce Handling (Public Clients)

For SPAs and mobile apps:

Initial Request:
- Send without nonce
- May receive use_dpop_nonce error
- Response includes DPoP-Nonce header

Retry with Nonce:
- Generate new DPoP proof
- Include nonce in payload
- Auth0 validates nonce freshness

### Token Binding

Auth0 binds token to public key:

Access Token Contains:
- cnf (confirmation) claim
- jkt (JWK Thumbprint) value
- SHA-256 hash of public key

This binding ensures only the key holder can use the token.

### API Request

When calling resource server:

Generate New DPoP Proof:
- htm: HTTP method (GET, POST, etc.)
- htu: API endpoint URL
- ath: Base64url SHA-256 hash of access token
- Same key pair as token request

Request Headers:
- Authorization: DPoP {access_token}
- DPoP: {dpop_proof_jwt}

### Server Validation

Resource server validates:
1. Extract DPoP proof from header
2. Verify proof signature
3. Check jti uniqueness (prevent replay)
4. Validate htm matches request method
5. Validate htu matches request URI
6. Compute access token hash
7. Compare ath with computed hash
8. Compare proof jwk thumbprint with token cnf.jkt

## Implementation Steps

### Step 1: Key Generation

Generate asymmetric key pair:
- Use cryptographic library
- Store private key securely
- Persist across token refresh

Key Considerations:
- Generate once per installation
- Reuse for token refresh
- May rotate periodically

### Step 2: Initial Token Request

Create DPoP proof for token endpoint:
- Include required claims
- Sign with private key
- No ath claim (no access token yet)

### Step 3: Handle Nonce (If Required)

If use_dpop_nonce error:
- Extract nonce from DPoP-Nonce header
- Create new proof with nonce
- Retry request

### Step 4: Store Tokens

After successful request:
- Store access token
- Store refresh token (if applicable)
- Associate with key pair

### Step 5: API Calls

For each API request:
- Generate fresh DPoP proof
- Include ath claim with token hash
- Send both token and proof

### Step 6: Token Refresh

When refreshing tokens:
- Generate new DPoP proof
- Use same key pair
- New tokens bound to same key

## Security Considerations

### Key Protection

Private Key Security:
- Never transmit private key
- Use secure storage
- Consider hardware backing

Key Compromise:
- If key compromised, token useless without new key
- Revoke tokens and regenerate key pair
- Better than bearer token compromise

### Replay Prevention

jti Uniqueness:
- Generate unique jti for each proof
- Resource server tracks seen jti values
- Rejects duplicates

Time Binding:
- iat limits proof validity
- Short acceptance window
- Clock synchronization important

### Token Binding

Benefits:
- Stolen token unusable without private key
- Attacker cannot forge valid proofs
- Significantly reduces token theft risk

## Client Types

### Confidential Clients

Server-side applications:
- Secure key storage available
- Combine with client authentication
- Strongest security configuration

### Public Clients

SPAs and mobile apps:
- Must handle nonce flow
- Use secure platform storage
- More complex but valuable

### Native Applications

Mobile apps:
- Use platform secure storage
- Hardware-backed keys when available
- iOS Keychain, Android Keystore

## Comparison with mTLS

DPoP Advantages:
- Application-layer (no TLS changes)
- Works with public clients
- No PKI infrastructure required
- Easier deployment

mTLS Advantages:
- Transport-layer binding
- Established infrastructure
- Simpler for confidential clients

Choose DPoP When:
- Public clients (SPA, mobile)
- No PKI available
- Flexibility needed

Choose mTLS When:
- Confidential clients only
- PKI exists
- Transport-layer binding preferred

## Troubleshooting

Common Issues:

Invalid Signature:
- Verify key pair consistency
- Check algorithm matches
- Confirm JWT format

Nonce Required:
- Public clients need nonce
- Extract from error response
- Include in retry

Token Binding Failure:
- Use same key for proof and token
- Verify ath calculation
- Check thumbprint computation

Clock Skew:
- Synchronize client clock
- Allow reasonable iat window
- Consider server time

## Best Practices

Key Management:
- Secure key generation
- Protected storage
- Consistent key usage

Proof Generation:
- Fresh proof per request
- Unique jti always
- Correct claim values

Error Handling:
- Handle nonce errors
- Retry logic
- Clear error messaging

Testing:
- Validate full flow
- Test error scenarios
- Verify binding works
