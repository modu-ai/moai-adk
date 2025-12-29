# Application Credentials

Auth0 supports multiple authentication methods for confidential applications to securely authenticate with authorization servers when requesting tokens.

## Credential Types

### Client Secret (Default)

Symmetric key authentication.

How It Works:
- Auth0 generates high-entropy secret
- Shared between application and Auth0
- Included in token requests

Transmission:
- Secret sent over network
- Included in request body or Basic auth header
- HTTPS required for security

Risks:
- Man-in-the-middle vulnerability
- Secret compromise = complete breach
- Must protect on both ends

Best For:
- Simple implementations
- Trusted environments
- Initial development

### Private Key JWT

Asymmetric key authentication.

How It Works:
- Application generates key pair
- Public key registered with Auth0
- Private key creates signed assertions
- Auth0 verifies with public key

Transmission:
- Private key never transmitted
- Only signed JWT sent
- Assertion has short expiry

Benefits:
- Private key stays private
- Limited replay window
- No shared secret

Requirements: Enterprise plan

### mTLS for OAuth

Certificate-based mutual TLS authentication.

How It Works:
- Application obtains X.509 certificate
- Certificate registered with Auth0
- mTLS connection established
- Certificate validates client identity

Transmission:
- Certificate private key never transmitted
- TLS handshake authenticates client
- Transport-layer security

Benefits:
- Strongest authentication
- No application-layer credentials
- Certificate-based identity

Requirements: Enterprise plan with HRI add-on

## Credential Comparison

### Security Ranking

Most Secure to Least:
1. mTLS for OAuth (certificate-based)
2. Private Key JWT (asymmetric)
3. Client Secret (symmetric)

### Complexity

Simplest to Most Complex:
1. Client Secret (minimal setup)
2. Private Key JWT (key management)
3. mTLS for OAuth (PKI infrastructure)

### Recommendations

Upgrade Path:
1. Start with Client Secret for development
2. Move to Private Key JWT for production
3. Use mTLS for highest security needs

## Private Key JWT Details

### Key Generation

Supported Algorithms:
- RS256, RS384, RS512 (RSA)
- PS256, PS384, PS512 (RSA-PSS)

Key Requirements:
- Minimum key size per algorithm
- Secure key generation
- Protected storage

### Client Assertion

JWT Structure:
- iss: Client ID
- sub: Client ID
- aud: Token endpoint URL
- iat: Issue time
- exp: Expiration time
- jti: Unique identifier

Assertion Lifetime:
- Short expiry recommended
- Limits replay window
- Typically seconds to minutes

### Token Request

Parameters:
- client_assertion_type: urn:ietf:params:oauth:client-assertion-type:jwt-bearer
- client_assertion: Signed JWT

### Key Registration

Dashboard Steps:
1. Navigate to Applications
2. Select application
3. Go to Credentials tab
4. Upload public key
5. Save configuration

API Registration:
- Use Management API
- Provide JWKS or JWK
- Associate with application

### Key Rotation

Zero-Downtime:
- Register up to two keys
- Deploy new key to application
- Remove old key after transition

Process:
1. Generate new key pair
2. Register new public key
3. Update application
4. Verify new key works
5. Remove old public key

## mTLS for OAuth Details

### Certificate Requirements

Valid X.509 Certificate:
- RSA or ECDSA key
- Appropriate validity period
- Proper extensions

Certificate Chain:
- Complete chain available
- Trusted CA or registered self-signed
- Proper intermediate certificates

### Certificate Registration

Dashboard Steps:
1. Navigate to Applications
2. Select application
3. Go to Credentials tab
4. Upload certificate
5. Save configuration

Multiple Certificates:
- Up to two certificates
- Enables rotation
- Remove old before adding third

### Token Request

Connection:
- Establish mTLS to token endpoint
- Present registered certificate
- Complete mutual authentication

## JWT-Secured Authorization Requests (JAR)

### Overview

Protect authorization request parameters:
- Sign request as JWT
- Optionally encrypt
- Ensure integrity and confidentiality

### Benefits

Integrity:
- Detect parameter tampering
- Verify request source
- Prevent manipulation

Confidentiality (with encryption):
- Hide sensitive parameters
- Protect from intermediaries
- Enhanced privacy

### Implementation

Create Request JWT:
- Include all authorization parameters
- Sign with registered key
- Send as request parameter

## Best Practices

### Secret Management

For Client Secrets:
- Secure storage
- Environment variables
- Secret management service
- Regular rotation

For Private Keys:
- HSM when possible
- Encrypted storage
- Access controls
- Regular rotation

For Certificates:
- Proper CA hierarchy
- Lifecycle management
- Rotation procedures
- Revocation capability

### Rotation

Regular Rotation:
- Schedule periodic rotation
- Automate when possible
- Test rotation procedures

Emergency Rotation:
- Immediate capability
- Documented procedures
- Tested regularly

### Monitoring

Track:
- Credential usage
- Failed authentications
- Rotation events
- Expiration dates

Alert On:
- Failed authentications
- Approaching expiration
- Unusual patterns

### Security

Principle of Least Privilege:
- Minimum required scopes
- Appropriate credential type
- Regular review

Audit:
- Credential access
- Configuration changes
- Token requests
