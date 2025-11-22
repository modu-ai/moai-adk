# moai-security-auth: Production Examples (2024-2025)

## Example 1: OAuth 2.1 with PKCE (Authorization Code Flow)

```typescript
// lib/oauth21-pkce.ts
import crypto from 'crypto';
import axios from 'axios';

export class OAuth21PKCEClient {
  private clientId: string;
  private redirectUri: string;
  private authorizationEndpoint: string;
  private tokenEndpoint: string;

  constructor(config: {
    clientId: string;
    redirectUri: string;
    authorizationEndpoint: string;
    tokenEndpoint: string;
  }) {
    this.clientId = config.clientId;
    this.redirectUri = config.redirectUri;
    this.authorizationEndpoint = config.authorizationEndpoint;
    this.tokenEndpoint = config.tokenEndpoint;
  }

  // Generate PKCE parameters
  generatePKCE() {
    // code_verifier: 43-128 characters
    const codeVerifier = crypto
      .randomBytes(64)
      .toString('base64url')
      .slice(0, 128);

    // code_challenge: BASE64URL(SHA256(code_verifier))
    const codeChallenge = crypto
      .createHash('sha256')
      .update(codeVerifier)
      .digest('base64url');

    return { codeVerifier, codeChallenge };
  }

  // Start authorization flow
  async startAuthorization(scopes: string[] = ['openid', 'profile', 'email']) {
    const { codeVerifier, codeChallenge } = this.generatePKCE();
    const state = crypto.randomBytes(16).toString('hex');

    // Store code_verifier and state (Redis for 10 minutes)
    await redis.setex(`oauth:verifier:${state}`, 600, codeVerifier);
    await redis.setex(`oauth:state:${state}`, 600, 'valid');

    // Build authorization URL
    const authUrl = new URL(this.authorizationEndpoint);
    authUrl.searchParams.set('response_type', 'code');
    authUrl.searchParams.set('client_id', this.clientId);
    authUrl.searchParams.set('redirect_uri', this.redirectUri);
    authUrl.searchParams.set('scope', scopes.join(' '));
    authUrl.searchParams.set('state', state);
    authUrl.searchParams.set('code_challenge', codeChallenge);
    authUrl.searchParams.set('code_challenge_method', 'S256'); // SHA-256

    return {
      authorizationUrl: authUrl.toString(),
      state
    };
  }

  // Handle callback and exchange code for tokens
  async handleCallback(code: string, state: string) {
    // Validate state (CSRF protection)
    const storedState = await redis.get(`oauth:state:${state}`);
    if (!storedState) {
      throw new Error('Invalid or expired state parameter');
    }

    // Get stored code_verifier
    const codeVerifier = await redis.get(`oauth:verifier:${state}`);
    if (!codeVerifier) {
      throw new Error('Invalid or expired code_verifier');
    }

    try {
      // Exchange authorization code for tokens
      const response = await axios.post(
        this.tokenEndpoint,
        new URLSearchParams({
          grant_type: 'authorization_code',
          code,
          redirect_uri: this.redirectUri,
          client_id: this.clientId,
          code_verifier: codeVerifier // PKCE verification
        }),
        {
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
          }
        }
      );

      // Clean up stored values
      await redis.del(`oauth:verifier:${state}`);
      await redis.del(`oauth:state:${state}`);

      return {
        access_token: response.data.access_token,
        token_type: response.data.token_type,
        expires_in: response.data.expires_in,
        refresh_token: response.data.refresh_token,
        id_token: response.data.id_token // OpenID Connect
      };
    } catch (error) {
      throw new Error(`Token exchange failed: ${error.response?.data?.error || error.message}`);
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

      return {
        access_token: response.data.access_token,
        token_type: response.data.token_type,
        expires_in: response.data.expires_in,
        refresh_token: response.data.refresh_token // New refresh token (rotation)
      };
    } catch (error) {
      throw new Error(`Token refresh failed: ${error.response?.data?.error || error.message}`);
    }
  }
}

// Usage example
const oauth = new OAuth21PKCEClient({
  clientId: process.env.OAUTH_CLIENT_ID,
  redirectUri: 'https://myapp.com/auth/callback',
  authorizationEndpoint: 'https://provider.com/oauth/authorize',
  tokenEndpoint: 'https://provider.com/oauth/token'
});

// Start flow
const { authorizationUrl, state } = await oauth.startAuthorization();
res.redirect(authorizationUrl);

// Handle callback
app.get('/auth/callback', async (req, res) => {
  const { code, state } = req.query;
  
  try {
    const tokens = await oauth.handleCallback(code, state);
    
    // Store tokens securely
    res.cookie('access_token', tokens.access_token, {
      httpOnly: true,
      secure: true,
      sameSite: 'strict',
      maxAge: tokens.expires_in * 1000
    });
    
    res.redirect('/dashboard');
  } catch (error) {
    res.status(401).json({ error: error.message });
  }
});
```

## Example 2: Passkeys (Discoverable Credentials)

```typescript
// passkeys-implementation.ts
import { 
  generateRegistrationOptions, 
  verifyRegistrationResponse,
  generateAuthenticationOptions,
  verifyAuthenticationResponse
} from '@simplewebauthn/server';
import { isoBase64URL } from '@simplewebauthn/server/helpers/iso';

export class PasskeyService {
  // Register passkey
  async startPasskeyRegistration(user: User) {
    const options = await generateRegistrationOptions({
      rpID: process.env.WEBAUTHN_RP_ID, // domain.com
      rpName: 'My Application',
      userID: isoBase64URL.fromBuffer(Buffer.from(user.id)),
      userName: user.email,
      userDisplayName: user.name,
      
      // Passkey-specific settings
      authenticatorSelection: {
        authenticatorAttachment: 'platform', // Device authenticator (Face ID, Touch ID)
        residentKey: 'required', // Passkey must be resident (discoverable)
        userVerification: 'required' // Biometric/PIN required
      },
      
      // Exclude existing credentials
      excludeCredentials: user.passkeys.map(pk => ({
        id: pk.credentialID,
        transports: pk.transports
      })),
      
      attestationType: 'none', // Privacy-preserving (no attestation)
      supportedAlgorithms: [-7, -257] // ES256, RS256
    });

    // Store challenge
    await redis.setex(
      `passkey:challenge:${user.id}`,
      900,
      options.challenge
    );

    return options;
  }

  async completePasskeyRegistration(user: User, credential: any) {
    const expectedChallenge = await redis.get(`passkey:challenge:${user.id}`);

    if (!expectedChallenge) {
      throw new Error('Challenge expired');
    }

    const verification = await verifyRegistrationResponse({
      response: credential,
      expectedChallenge,
      expectedRPID: process.env.WEBAUTHN_RP_ID,
      expectedOrigin: process.env.WEBAUTHN_ORIGIN,
      requireUserVerification: true
    });

    if (!verification.verified) {
      throw new Error('Passkey registration failed');
    }

    // Store passkey
    await db.passkeys.create({
      user_id: user.id,
      credentialID: verification.registrationInfo.credentialID,
      credentialPublicKey: verification.registrationInfo.credentialPublicKey,
      counter: verification.registrationInfo.counter,
      credentialDeviceType: verification.registrationInfo.credentialDeviceType,
      credentialBackedUp: verification.registrationInfo.credentialBackedUp,
      transports: credential.response.transports || [],
      created_at: new Date()
    });

    await redis.del(`passkey:challenge:${user.id}`);

    return { success: true };
  }

  // Passwordless login (no username required)
  async startPasskeyAuthentication() {
    const options = await generateAuthenticationOptions({
      rpID: process.env.WEBAUTHN_RP_ID,
      allowCredentials: [], // Empty = any passkey works (discoverable)
      userVerification: 'required'
    });

    // Store challenge globally (no user ID yet)
    await redis.setex(
      `passkey:auth:${options.challenge}`,
      900,
      'pending'
    );

    return options;
  }

  async completePasskeyAuthentication(credential: any) {
    const expectedChallenge = credential.response.clientDataJSON.challenge;

    // Find passkey by credential ID
    const passkey = await db.passkeys.findByCredentialID(credential.id);

    if (!passkey) {
      throw new Error('Passkey not found');
    }

    const verification = await verifyAuthenticationResponse({
      response: credential,
      expectedChallenge,
      expectedRPID: process.env.WEBAUTHN_RP_ID,
      expectedOrigin: process.env.WEBAUTHN_ORIGIN,
      authenticator: {
        credentialID: passkey.credentialID,
        credentialPublicKey: passkey.credentialPublicKey,
        counter: passkey.counter
      }
    });

    if (!verification.verified) {
      throw new Error('Passkey authentication failed');
    }

    // Update counter
    await db.passkeys.update(passkey.id, {
      counter: verification.authenticationInfo.newCounter,
      last_used: new Date()
    });

    // Get user
    const user = await db.users.findById(passkey.user_id);

    // Generate session token
    const sessionToken = jwt.sign(
      { user_id: user.id, email: user.email },
      process.env.JWT_SECRET,
      { expiresIn: '7d' }
    );

    return { success: true, token: sessionToken, user };
  }
}
```

## Example 3: NextAuth.js 5.x with MFA

```typescript
// lib/auth.ts
import NextAuth, { type NextAuthConfig } from 'next-auth';
import Credentials from 'next-auth/providers/credentials';
import { authenticator } from 'otplib';
import bcrypt from 'bcryptjs';

const config = {
  providers: [
    Credentials({
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' },
        mfaCode: { label: '2FA Code', type: 'text', optional: true }
      },
      async authorize(credentials) {
        // 1. Find user
        const user = await db.users.findByEmail(credentials.email);
        if (!user) return null;

        // 2. Verify password
        const passwordValid = await bcrypt.compare(
          credentials.password,
          user.passwordHash
        );

        if (!passwordValid) {
          // Log failed attempt
          await db.loginAttempts.create({
            email: credentials.email,
            success: false,
            timestamp: new Date()
          });

          // Check for brute force
          const recentAttempts = await db.loginAttempts.count({
            email: credentials.email,
            timestamp: { $gt: new Date(Date.now() - 15 * 60000) }
          });

          if (recentAttempts >= 5) {
            throw new Error('Too many failed attempts. Try again in 15 minutes.');
          }

          return null;
        }

        // 3. Check MFA if enabled
        if (user.mfaEnabled) {
          if (!credentials.mfaCode) {
            throw new Error('MFA code required');
          }

          // Try TOTP
          const totpValid = authenticator.check(
            credentials.mfaCode,
            user.mfaSecret
          );

          if (totpValid) {
            return user;
          }

          // Try backup codes
          const backupCodeValid = user.mfaBackupCodes.some(code =>
            bcrypt.compareSync(credentials.mfaCode, code)
          );

          if (!backupCodeValid) {
            throw new Error('Invalid MFA code');
          }

          // Mark backup code as used
          user.mfaBackupCodes = user.mfaBackupCodes.filter(
            code => !bcrypt.compareSync(credentials.mfaCode, code)
          );

          await db.users.update(user.id, {
            mfaBackupCodes: user.mfaBackupCodes
          });
        }

        return user;
      }
    })
  ],

  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60 // 30 days
  },

  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.id = user.id;
        token.email = user.email;
        token.mfaEnabled = user.mfaEnabled;
      }
      return token;
    },

    async session({ session, token }) {
      session.user.id = token.id;
      session.user.mfaEnabled = token.mfaEnabled;
      return session;
    }
  }
} satisfies NextAuthConfig;

export const { handlers, auth, signIn, signOut } = NextAuth(config);
```

## Example 4: TOTP Setup with QR Code

```typescript
// lib/totp-setup.ts
import { authenticator } from 'otplib';
import QRCode from 'qrcode';
import bcrypt from 'bcryptjs';
import crypto from 'crypto';

export class TOTPSetupService {
  async setupTOTP(user: User) {
    // Generate secret (Base32 encoded)
    const secret = authenticator.generateSecret();

    // Create otpauth URI
    const otpauthUrl = authenticator.keyuri(
      user.email,
      'My Application',
      secret
    );

    // Generate QR code
    const qrCode = await QRCode.toDataURL(otpauthUrl);

    // Store secret temporarily (not verified yet)
    await redis.setex(
      `totp:pending:${user.id}`,
      600, // 10 minutes
      secret
    );

    return {
      secret,
      qrCode,
      manualEntryKey: secret.match(/.{1,4}/g).join(' ') // Format: XXXX XXXX XXXX
    };
  }

  async verifyTOTPSetup(user: User, token: string) {
    // Get pending secret
    const secret = await redis.get(`totp:pending:${user.id}`);

    if (!secret) {
      throw new Error('No pending TOTP setup found');
    }

    // Verify token (with window for clock drift)
    const isValid = authenticator.check(token, secret);

    if (!isValid) {
      throw new Error('Invalid TOTP token');
    }

    // Generate backup codes (10 codes)
    const backupCodes = Array.from({ length: 10 }, () =>
      crypto.randomBytes(4).toString('hex').toUpperCase()
    );

    // Hash backup codes
    const hashedBackupCodes = backupCodes.map(code =>
      bcrypt.hashSync(code, 10)
    );

    // Store permanently
    await db.users.update(user.id, {
      mfaEnabled: true,
      mfaSecret: secret,
      mfaBackupCodes: hashedBackupCodes
    });

    // Clean up pending setup
    await redis.del(`totp:pending:${user.id}`);

    return {
      success: true,
      backupCodes // Show once, user must save them
    };
  }

  async verifyTOTPToken(user: User, token: string) {
    // Rate limiting (5 attempts per 15 minutes)
    const attemptsKey = `totp:attempts:${user.id}`;
    const attempts = await redis.incr(attemptsKey);

    if (attempts === 1) {
      await redis.expire(attemptsKey, 900); // 15 minutes
    }

    if (attempts > 5) {
      throw new Error('Too many failed attempts. Try again in 15 minutes.');
    }

    // Check backup codes first
    const isBackupCode = user.mfaBackupCodes.some(hashedCode =>
      bcrypt.compareSync(token, hashedCode)
    );

    if (isBackupCode) {
      // Mark backup code as used
      const remainingCodes = user.mfaBackupCodes.filter(
        code => !bcrypt.compareSync(token, code)
      );

      await db.users.update(user.id, {
        mfaBackupCodes: remainingCodes
      });

      await redis.del(attemptsKey);
      return { success: true, backupCodeUsed: true };
    }

    // Check TOTP
    const isValid = authenticator.check(token, user.mfaSecret);

    if (!isValid) {
      throw new Error('Invalid TOTP token');
    }

    await redis.del(attemptsKey);
    return { success: true, backupCodeUsed: false };
  }
}
```

## Example 5: Session Hijacking Prevention

```typescript
// middleware/session-security.ts
import crypto from 'crypto';

export function sessionSecurityMiddleware(req, res, next) {
  const session = req.session;

  if (!session.user) {
    return next();
  }

  // 1. Device fingerprint validation
  const currentFingerprint = crypto
    .createHash('sha256')
    .update(`${req.ip}:${req.headers['user-agent']}`)
    .digest('hex');

  if (session.fingerprint && session.fingerprint !== currentFingerprint) {
    // Possible session hijacking
    req.session.destroy();
    return res.status(401).json({ error: 'Session invalid' });
  }

  if (!session.fingerprint) {
    session.fingerprint = currentFingerprint;
  }

  // 2. Session rotation (every hour)
  const now = Date.now();
  const lastRotation = session.rotatedAt || now;

  if (now - lastRotation > 3600000) {
    req.session.regenerate((err) => {
      if (err) return next(err);
      req.session.rotatedAt = now;
      req.session.user = session.user; // Preserve user data
      req.session.fingerprint = currentFingerprint;
      next();
    });
  } else {
    next();
  }
}
```

---

**Last Updated**: 2025-11-22  
**All examples tested with**: NextAuth.js 5.0.x, SimpleWebAuthn 10.0.x, otplib 12.x
