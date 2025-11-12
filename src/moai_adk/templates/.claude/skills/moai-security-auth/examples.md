# moai-security-auth: Production Examples

## Example 1: NextAuth.js 5 with MFA (TOTP + Backup Codes)

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
            ip: null,
            success: false
          });
          
          // Lock after 5 attempts
          const attempts = await db.loginAttempts.count({
            email: credentials.email,
            timestamp: { $gt: new Date(Date.now() - 15 * 60000) }
          });
          
          if (attempts >= 5) {
            throw new Error('Too many failed attempts. Try again in 15 minutes.');
          }
          
          return null;
        }
        
        // 3. Check MFA if enabled
        if (user.mfaEnabled) {
          if (!credentials.mfaCode) {
            throw new Error('MFA required');
          }
          
          // Try TOTP first
          const totpValid = authenticator.check(
            credentials.mfaCode,
            user.mfaSecret
          );
          
          if (totpValid) {
            return user;
          }
          
          // Try backup codes
          const backupCodeIndex = user.mfaBackupCodes.findIndex(code =>
            bcrypt.compareSync(credentials.mfaCode, code)
          );
          
          if (backupCodeIndex === -1) {
            throw new Error('Invalid MFA code');
          }
          
          // Mark backup code as used
          user.mfaBackupCodes.splice(backupCodeIndex, 1);
          await db.users.update(user.id, {
            mfaBackupCodes: user.mfaBackupCodes
          });
        }
        
        return user;
      }
    })
  ],
  
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.mfaEnabled = user.mfaEnabled;
        token.role = user.role;
      }
      return token;
    },
    
    async session({ session, token }) {
      session.user.mfaEnabled = token.mfaEnabled;
      session.user.role = token.role;
      return session;
    }
  }
} satisfies NextAuthConfig;

export const { handlers, auth, signIn, signOut } = NextAuth(config);
```

## Example 2: WebAuthn Registration & Authentication

```typescript
// lib/webauthn-service.ts
import { 
  generateRegistrationOptions, 
  verifyRegistrationResponse,
  generateAuthenticationOptions,
  verifyAuthenticationResponse
} from '@simplewebauthn/server';
import { isoBase64URL } from '@simplewebauthn/server/helpers/iso';

export class WebAuthnService {
  // Register new security key
  async startRegistration(user: User) {
    const options = generateRegistrationOptions({
      rpID: process.env.WEBAUTHN_RP_ID,
      rpName: 'MyApp',
      userID: isoBase64URL.fromBuffer(Buffer.from(user.id)),
      userName: user.email,
      userDisplayName: user.name,
      
      // Require platform authenticator (biometric/PIN)
      authenticatorSelection: {
        authenticatorAttachment: 'platform',
        residentKey: 'required',
        userVerification: 'preferred'
      },
      
      attestationType: 'direct',
      supportedAlgos: [-7, -257]
    });
    
    // Store challenge in Redis
    await redis.setex(
      `webauthn:challenge:${user.id}`,
      900,
      Buffer.from(options.challenge).toString('hex')
    );
    
    return options;
  }
  
  // Complete registration
  async completeRegistration(
    user: User,
    attestationResponse: any
  ) {
    const challenge = Buffer.from(
      await redis.get(`webauthn:challenge:${user.id}`),
      'hex'
    );
    
    const verification = await verifyRegistrationResponse({
      response: attestationResponse,
      expectedChallenge: challenge,
      expectedRPID: process.env.WEBAUTHN_RP_ID,
      expectedOrigin: process.env.WEBAUTHN_ORIGIN,
      requireUserVerification: true
    });
    
    if (!verification.verified) {
      throw new Error('Registration failed');
    }
    
    // Store credential
    await db.webauthnCredentials.create({
      user_id: user.id,
      credential_id: verification.registrationInfo!.credentialID,
      public_key: verification.registrationInfo!.credentialPublicKey,
      counter: verification.registrationInfo!.counter,
      transports: attestationResponse.response.getTransports?.(),
      created_at: new Date()
    });
    
    // Clean up
    await redis.del(`webauthn:challenge:${user.id}`);
    
    return true;
  }
  
  // Start authentication
  async startAuthentication(email: string) {
    const user = await db.users.findByEmail(email);
    if (!user) throw new Error('User not found');
    
    const credentials = await db.webauthnCredentials.findByUserId(user.id);
    
    const options = generateAuthenticationOptions({
      rpID: process.env.WEBAUTHN_RP_ID,
      allowCredentials: credentials.map(c => ({
        id: c.credential_id,
        type: 'public-key',
        transports: c.transports
      })),
      userVerification: 'required'
    });
    
    await redis.setex(
      `webauthn:auth:${user.id}`,
      900,
      Buffer.from(options.challenge).toString('hex')
    );
    
    return options;
  }
  
  // Complete authentication
  async completeAuthentication(
    email: string,
    assertionResponse: any
  ) {
    const user = await db.users.findByEmail(email);
    const credential = await db.webauthnCredentials.findByCredentialID(
      assertionResponse.id
    );
    const challenge = Buffer.from(
      await redis.get(`webauthn:auth:${user.id}`),
      'hex'
    );
    
    const verification = await verifyAuthenticationResponse({
      response: assertionResponse,
      expectedChallenge: challenge,
      expectedRPID: process.env.WEBAUTHN_RP_ID,
      expectedOrigin: process.env.WEBAUTHN_ORIGIN,
      authenticator: {
        credentialID: credential.credential_id,
        credentialPublicKey: credential.public_key,
        counter: credential.counter
      }
    });
    
    if (!verification.verified) {
      throw new Error('Authentication failed');
    }
    
    // Update counter (prevents cloning)
    await db.webauthnCredentials.update(credential.id, {
      counter: verification.authenticationInfo!.newCounter
    });
    
    return user;
  }
}
```

## Example 3: TOTP Setup & Verification

```typescript
// lib/totp-service.ts
import { authenticator } from 'otplib';
import QRCode from 'qrcode';
import bcrypt from 'bcryptjs';

export class TOTPService {
  // Setup TOTP
  async setupTOTP(user: User) {
    const secret = authenticator.generateSecret({
      name: `MyApp (${user.email})`
    });
    
    // Generate QR code
    const qrCode = await QRCode.toDataURL(secret);
    
    // Store temporarily (not verified yet)
    await redis.setex(
      `totp:pending:${user.id}`,
      600,
      secret
    );
    
    return { secret, qrCode };
  }
  
  // Verify TOTP setup
  async verifyTOTPSetup(user: User, token: string) {
    const secret = await redis.get(`totp:pending:${user.id}`);
    if (!secret) throw new Error('No pending TOTP');
    
    // Verify token
    const isValid = authenticator.check(token, secret);
    if (!isValid) throw new Error('Invalid token');
    
    // Generate backup codes
    const backupCodes = Array.from({ length: 10 })
      .map(() => crypto.randomBytes(4).toString('hex').toUpperCase());
    
    // Store permanently
    await db.users.update(user.id, {
      mfaEnabled: true,
      mfaSecret: secret,
      mfaBackupCodes: backupCodes.map(code => bcrypt.hashSync(code, 10))
    });
    
    await redis.del(`totp:pending:${user.id}`);
    
    return backupCodes;
  }
  
  // Verify TOTP token during login
  async verifyToken(user: User, token: string) {
    // Check backup codes first
    const isBackup = user.mfaBackupCodes.some(code =>
      bcrypt.compareSync(token, code)
    );
    
    if (isBackup) {
      // Mark as used
      const codes = user.mfaBackupCodes.filter(
        code => !bcrypt.compareSync(token, code)
      );
      await db.users.update(user.id, { mfaBackupCodes: codes });
      return true;
    }
    
    // Check TOTP
    return authenticator.check(token, user.mfaSecret);
  }
}
```

## Example 4: Passport.js Local + JWT Strategy

```typescript
// strategies/passport-config.ts
import { Strategy as LocalStrategy } from 'passport-local';
import { Strategy as JwtStrategy, ExtractJwt } from 'passport-jwt';
import bcrypt from 'bcryptjs';
import passport from 'passport';

export function setupPassport() {
  // 1. Local Strategy (login)
  passport.use(new LocalStrategy(
    {
      usernameField: 'email',
      passwordField: 'password'
    },
    async (email, password, done) => {
      try {
        const user = await db.users.findByEmail(email);
        
        if (!user) {
          return done(null, false, { message: 'User not found' });
        }
        
        // Check account lock
        if (user.lockedUntil && user.lockedUntil > new Date()) {
          return done(null, false, { message: 'Account locked' });
        }
        
        // Verify password
        const isValid = await bcrypt.compare(password, user.passwordHash);
        
        if (!isValid) {
          // Increment failed attempts
          const attempts = (user.loginAttempts || 0) + 1;
          const lockUntil = attempts >= 5 
            ? new Date(Date.now() + 30 * 60000)
            : null;
          
          await db.users.update(user.id, {
            loginAttempts: attempts,
            lockedUntil: lockUntil
          });
          
          return done(null, false, { message: 'Invalid password' });
        }
        
        // Reset attempts on success
        await db.users.update(user.id, {
          loginAttempts: 0,
          lockedUntil: null
        });
        
        return done(null, user);
      } catch (err) {
        return done(err);
      }
    }
  ));
  
  // 2. JWT Strategy (verification)
  passport.use(new JwtStrategy(
    {
      jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
      secretOrKey: process.env.JWT_SECRET,
      algorithms: ['HS256']
    },
    async (payload, done) => {
      try {
        // Check token blacklist
        const isBlacklisted = await redis.get(`jwt:blacklist:${payload.jti}`);
        if (isBlacklisted) {
          return done(null, false);
        }
        
        const user = await db.users.findById(payload.id);
        if (!user) return done(null, false);
        
        return done(null, user, payload);
      } catch (err) {
        return done(err);
      }
    }
  ));
  
  // 3. Serialization
  passport.serializeUser((user, done) => {
    done(null, user.id);
  });
  
  passport.deserializeUser(async (id, done) => {
    try {
      const user = await db.users.findById(id);
      done(null, user);
    } catch (err) {
      done(err);
    }
  });
}

// Usage in routes
app.post('/login', passport.authenticate('local'), (req, res) => {
  const token = jwt.sign(
    { id: req.user.id, jti: uuid() },
    process.env.JWT_SECRET,
    { expiresIn: '7d' }
  );
  
  res.json({ token });
});

app.post('/logout', (req, res) => {
  // Add token to blacklist
  redis.setex(`jwt:blacklist:${req.user.jti}`, 604800, '1'); // 7 days
  res.json({ message: 'Logged out' });
});

app.get('/profile', 
  passport.authenticate('jwt'),
  (req, res) => {
    res.json(req.user);
  }
);
```

## Example 5: Session Hijacking Prevention

```typescript
// middleware/session-security.ts
import crypto from 'crypto';

export function sessionSecurityMiddleware() {
  return async (req, res, next) => {
    const session = req.session;
    const sessionId = req.sessionID;
    
    if (!session.user) {
      return next();
    }
    
    // 1. Fingerprint validation (browser + IP + user agent)
    const currentFingerprint = crypto
      .createHash('sha256')
      .update(req.ip + req.headers['user-agent'])
      .digest('hex');
    
    if (session.fingerprint && session.fingerprint !== currentFingerprint) {
      // Different browser/IP - possible hijack
      delete req.session;
      return res.status(401).json({ error: 'Session invalid' });
    }
    
    // Set fingerprint on first request
    if (!session.fingerprint) {
      session.fingerprint = currentFingerprint;
    }
    
    // 2. Regenerate session ID every hour
    const now = Date.now();
    const lastRegenerated = session.regeneratedAt || now;
    
    if (now - lastRegenerated > 3600000) {
      req.session.regenerate((err) => {
        if (err) return next(err);
        req.session.regeneratedAt = now;
        next();
      });
    } else {
      next();
    }
  };
}
```

