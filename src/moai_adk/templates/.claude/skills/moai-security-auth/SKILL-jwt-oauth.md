---
module: jwt-oauth-2-1
parent: moai-security-auth
description: JWT and OAuth 2.1 modern authentication patterns with PKCE
---

# JWT & OAuth 2.1 Authentication

**Modern authentication with OAuth 2.1 and secure JWT patterns**

## OAuth 2.1 Key Changes (2024-2025)

### What Changed from OAuth 2.0

**Removed/Deprecated**:
- ❌ Implicit Flow (security vulnerability)
- ❌ Resource Owner Password Credentials Grant
- ❌ Bearer token in query parameters

**Now Mandatory**:
- ✅ PKCE (Proof Key for Code Exchange) for ALL clients
- ✅ Exact redirect URI matching
- ✅ State parameter for CSRF protection

### PKCE Flow (Authorization Code + PKCE)

```
Client → code_verifier (random 43-128 chars)
      → code_challenge = BASE64URL(SHA256(code_verifier))

1. Authorization Request:
   GET /authorize?
     response_type=code
     &client_id=CLIENT_ID
     &redirect_uri=CALLBACK_URL
     &scope=openid profile email
     &state=RANDOM_STATE
     &code_challenge=CHALLENGE
     &code_challenge_method=S256

2. Authorization Server → User Login → Redirect:
   CALLBACK_URL?code=AUTH_CODE&state=RANDOM_STATE

3. Token Request:
   POST /token
   {
     grant_type: "authorization_code",
     code: "AUTH_CODE",
     redirect_uri: "CALLBACK_URL",
     client_id: "CLIENT_ID",
     code_verifier: "ORIGINAL_VERIFIER"
   }

4. Response:
   {
     access_token: "...",
     token_type: "Bearer",
     expires_in: 3600,
     refresh_token: "...",
     id_token: "..." // if OpenID Connect
   }
```

### OAuth 2.1 Implementation (Node.js)

```typescript
// lib/oauth21-client.ts
import crypto from 'crypto';
import axios from 'axios';

export class OAuth21Client {
  private clientId: string;
  private redirectUri: string;
  private authorizationEndpoint: string;
  private tokenEndpoint: string;

  constructor(config: OAuth21Config) {
    this.clientId = config.clientId;
    this.redirectUri = config.redirectUri;
    this.authorizationEndpoint = config.authorizationEndpoint;
    this.tokenEndpoint = config.tokenEndpoint;
  }

  // Generate PKCE challenge
  generatePKCE() {
    // 1. code_verifier: 43-128 characters
    const codeVerifier = crypto
      .randomBytes(64)
      .toString('base64url')
      .slice(0, 128);

    // 2. code_challenge: SHA256(verifier)
    const codeChallenge = crypto
      .createHash('sha256')
      .update(codeVerifier)
      .digest('base64url');

    return { codeVerifier, codeChallenge };
  }

  // Start authorization flow
  async startAuthorization(scopes: string[] = ['openid', 'profile']) {
    const { codeVerifier, codeChallenge } = this.generatePKCE();
    const state = crypto.randomBytes(16).toString('hex');

    // Store verifier and state (Redis, session, etc.)
    await redis.setex(`oauth:verifier:${state}`, 600, codeVerifier);

    const authUrl = new URL(this.authorizationEndpoint);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('client_id', this.clientId);
    authUrl.searchParams.set('redirect_uri', this.redirectUri);
    authUrl.searchParams.set('scope', scopes.join(' '));
    authUrl.searchParams.set('state', state);
    authUrl.searchParams.set('code_challenge', codeChallenge);
    authUrl.searchParams.set('code_challenge_method', 'S256');

    return {
      authorizationUrl: authUrl.toString(),
      state
    };
  }

  // Exchange authorization code for tokens
  async exchangeCodeForTokens(code: string, state: string) {
    // Retrieve stored code_verifier
    const codeVerifier = await redis.get(`oauth:verifier:${state}`);
    if (!codeVerifier) {
      throw new Error('Invalid or expired state');
    }

    try {
      const response = await axios.post(
        this.tokenEndpoint,
        new URLSearchParams({
          grant_type: 'authorization_code',
          code,
          redirect_uri: this.redirectUri,
          client_id: this.clientId,
          code_verifier: codeVerifier
        }),
        {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }
      );

      // Clean up stored verifier
      await redis.del(`oauth:verifier:${state}`);

      return response.data;
    } catch (error) {
      throw new Error(`Token exchange failed: ${error.message}`);
    }
  }

  // Refresh access token
  async refreshAccessToken(refreshToken: string) {
    try {
      const response = await axios.post(
        this.tokenEndpoint,
        new URLSearchParams({
          grant_type: 'refresh_token',
          refresh_token: refreshToken,
          client_id: this.clientId
        }),
        {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }
      );

      return response.data;
    } catch (error) {
      throw new Error(`Token refresh failed: ${error.message}`);
    }
  }
}
```

## JWT Best Practices (2025)

### Secure JWT Configuration

```typescript
// lib/jwt-service.ts
import jwt from 'jsonwebtoken';
import crypto from 'crypto';

export class JWTService {
  private secret: string;
  private algorithm: 'HS256' | 'RS256';
  
  constructor(secret: string, algorithm: 'HS256' | 'RS256' = 'HS256') {
    if (secret.length < 32) {
      throw new Error('JWT secret must be at least 32 characters');
    }
    this.secret = secret;
    this.algorithm = algorithm;
  }

  // Generate access token (short-lived)
  generateAccessToken(payload: TokenPayload): string {
    return jwt.sign(
      {
        ...payload,
        type: 'access',
        jti: crypto.randomUUID() // JWT ID for revocation
      },
      this.secret,
      {
        algorithm: this.algorithm,
        expiresIn: '15m', // 15 minutes
        issuer: process.env.JWT_ISSUER,
        audience: process.env.JWT_AUDIENCE
      }
    );
  }

  // Generate refresh token (long-lived)
  generateRefreshToken(payload: TokenPayload): string {
    return jwt.sign(
      {
        user_id: payload.user_id,
        type: 'refresh',
        jti: crypto.randomUUID()
      },
      this.secret,
      {
        algorithm: this.algorithm,
        expiresIn: '7d', // 7 days
        issuer: process.env.JWT_ISSUER,
        audience: process.env.JWT_AUDIENCE
      }
    );
  }

  // Verify token
  verifyToken(token: string): TokenPayload {
    try {
      const decoded = jwt.verify(token, this.secret, {
        algorithms: [this.algorithm],
        issuer: process.env.JWT_ISSUER,
        audience: process.env.JWT_AUDIENCE
      }) as TokenPayload;

      // Check token blacklist
      if (this.isTokenBlacklisted(decoded.jti)) {
        throw new Error('Token has been revoked');
      }

      return decoded;
    } catch (error) {
      if (error.name === 'TokenExpiredError') {
        throw new Error('Token expired');
      } else if (error.name === 'JsonWebTokenError') {
        throw new Error('Invalid token');
      }
      throw error;
    }
  }

  // Revoke token (add to blacklist)
  async revokeToken(jti: string, expiresIn: number): Promise<void> {
    await redis.setex(`jwt:blacklist:${jti}`, expiresIn, '1');
  }

  // Check if token is blacklisted
  async isTokenBlacklisted(jti: string): Promise<boolean> {
    const result = await redis.get(`jwt:blacklist:${jti}`);
    return result === '1';
  }
}

interface TokenPayload {
  user_id: string;
  email: string;
  role: string;
  type: 'access' | 'refresh';
  jti: string;
  iat?: number;
  exp?: number;
}
```

### Token Rotation Strategy

```typescript
// middleware/token-refresh.ts
export async function tokenRefreshMiddleware(req, res, next) {
  const accessToken = req.headers.authorization?.split(' ')[1];
  
  if (!accessToken) {
    return res.status(401).json({ error: 'No token provided' });
  }

  try {
    // Verify access token
    const decoded = jwtService.verifyToken(accessToken);
    req.user = decoded;
    next();
  } catch (error) {
    if (error.message === 'Token expired') {
      // Check if refresh token exists
      const refreshToken = req.cookies.refresh_token;
      
      if (!refreshToken) {
        return res.status(401).json({ error: 'Session expired' });
      }

      try {
        // Verify refresh token
        const refreshDecoded = jwtService.verifyToken(refreshToken);
        
        if (refreshDecoded.type !== 'refresh') {
          throw new Error('Invalid token type');
        }

        // Generate new access token
        const newAccessToken = jwtService.generateAccessToken({
          user_id: refreshDecoded.user_id,
          email: refreshDecoded.email,
          role: refreshDecoded.role
        });

        // Set new access token in header
        res.setHeader('X-New-Access-Token', newAccessToken);
        
        req.user = refreshDecoded;
        next();
      } catch (refreshError) {
        return res.status(401).json({ error: 'Invalid refresh token' });
      }
    } else {
      return res.status(401).json({ error: error.message });
    }
  }
}
```

## Security Best Practices

### DO ✅

1. **Use PKCE for all OAuth flows** (mandatory in OAuth 2.1)
2. **Short-lived access tokens** (15 minutes or less)
3. **Secure refresh token storage** (httpOnly cookies, encrypted)
4. **Token rotation** (new refresh token on use)
5. **Validate iss, aud, exp claims** (prevent token reuse)
6. **Use strong secrets** (minimum 32 characters)
7. **Implement token blacklisting** (for logout)

### DON'T ❌

1. ❌ Don't use Implicit Flow (removed in OAuth 2.1)
2. ❌ Don't store tokens in localStorage (XSS vulnerability)
3. ❌ Don't use weak signing algorithms (HS256 < 256 bits)
4. ❌ Don't skip CSRF protection (state parameter)
5. ❌ Don't use long-lived access tokens (> 1 hour)
6. ❌ Don't expose tokens in URLs (query parameters)

## Related Modules

- [WebAuthn & Passkeys](SKILL-webauthn.md) - Passwordless authentication
- [MFA Patterns](SKILL.md#multi-factor-authentication-totp) - TOTP, SMS, backup codes

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready (OAuth 2.1 compliance)
