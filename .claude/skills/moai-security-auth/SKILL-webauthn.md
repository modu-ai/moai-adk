---
module: webauthn-passkeys
parent: moai-security-auth
description: WebAuthn, FIDO2, and Passkeys passwordless authentication
---

# WebAuthn & Passkeys

**Passwordless authentication with WebAuthn (W3C) and FIDO2 standards**

## WebAuthn Overview

### What is WebAuthn?

**WebAuthn** (Web Authentication API) is a W3C standard for passwordless authentication using:
- **Platform authenticators**: Face ID, Touch ID, Windows Hello
- **Roaming authenticators**: USB security keys (YubiKey, Titan Security Key)
- **Passkeys**: Synced credentials across devices (iCloud Keychain, Google Password Manager)

### Key Benefits

- ✅ **Phishing-resistant**: Public key cryptography prevents credential theft
- ✅ **No passwords**: Biometric or PIN verification
- ✅ **Multi-device**: Passkeys sync across user's devices
- ✅ **Privacy-preserving**: No tracking across sites
- ✅ **Fast**: Touch/Face ID faster than typing passwords

## WebAuthn Registration Flow

### Client-Side (Browser)

```typescript
// registration-flow.ts
export async function registerWebAuthn(user: { id: string; name: string; email: string }) {
  try {
    // 1. Get registration options from server
    const optionsResponse = await fetch('/auth/webauthn/register/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ userId: user.id })
    });

    const options = await optionsResponse.json();

    // 2. Create credential with browser API
    const credential = await navigator.credentials.create({
      publicKey: {
        challenge: base64urlDecode(options.challenge),
        rp: {
          name: options.rp.name,
          id: options.rp.id // domain.com
        },
        user: {
          id: base64urlDecode(options.user.id),
          name: options.user.name,
          displayName: options.user.displayName
        },
        pubKeyCredParams: options.pubKeyCredParams,
        timeout: 60000,
        authenticatorSelection: {
          authenticatorAttachment: 'platform', // Face ID, Touch ID
          residentKey: 'required', // Passkey support
          userVerification: 'required' // Biometric required
        },
        attestation: 'direct'
      }
    });

    // 3. Send credential to server
    const verifyResponse = await fetch('/auth/webauthn/register/complete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        credential: {
          id: credential.id,
          rawId: base64urlEncode(credential.rawId),
          response: {
            attestationObject: base64urlEncode(credential.response.attestationObject),
            clientDataJSON: base64urlEncode(credential.response.clientDataJSON)
          },
          type: credential.type
        }
      })
    });

    const result = await verifyResponse.json();
    
    if (result.success) {
      console.log('WebAuthn registration successful');
    }
  } catch (error) {
    if (error.name === 'NotAllowedError') {
      console.error('User canceled registration');
    } else if (error.name === 'InvalidStateError') {
      console.error('Authenticator already registered');
    } else {
      console.error('WebAuthn error:', error);
    }
  }
}

function base64urlDecode(str: string): ArrayBuffer {
  const base64 = str.replace(/-/g, '+').replace(/_/g, '/');
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes.buffer;
}

function base64urlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}
```

### Server-Side (Node.js + SimpleWebAuthn)

```typescript
// routes/webauthn-register.ts
import { 
  generateRegistrationOptions, 
  verifyRegistrationResponse 
} from '@simplewebauthn/server';
import { isoBase64URL } from '@simplewebauthn/server/helpers/iso';

export async function startRegistration(req, res) {
  const { userId } = req.body;
  const user = await db.users.findById(userId);

  if (!user) {
    return res.status(404).json({ error: 'User not found' });
  }

  // Generate registration options
  const options = await generateRegistrationOptions({
    rpID: process.env.WEBAUTHN_RP_ID, // domain.com
    rpName: 'My Application',
    userID: isoBase64URL.fromBuffer(Buffer.from(user.id)),
    userName: user.email,
    userDisplayName: user.name,
    
    // Authenticator selection
    authenticatorSelection: {
      authenticatorAttachment: 'platform', // Platform authenticator (Face ID, Touch ID)
      residentKey: 'required', // Passkey support
      userVerification: 'required' // Require biometric/PIN
    },
    
    // Exclude already registered credentials
    excludeCredentials: user.webauthnCredentials.map(cred => ({
      id: cred.credentialID,
      transports: cred.transports
    })),
    
    attestationType: 'direct',
    supportedAlgorithms: [-7, -257] // ES256, RS256
  });

  // Store challenge in Redis (15 minutes TTL)
  await redis.setex(
    `webauthn:challenge:${user.id}`,
    900,
    options.challenge
  );

  res.json(options);
}

export async function completeRegistration(req, res) {
  const { credential } = req.body;
  const user = req.user; // Assumes authenticated session

  // Get stored challenge
  const expectedChallenge = await redis.get(`webauthn:challenge:${user.id}`);

  if (!expectedChallenge) {
    return res.status(400).json({ error: 'Challenge expired' });
  }

  try {
    // Verify registration response
    const verification = await verifyRegistrationResponse({
      response: credential,
      expectedChallenge,
      expectedRPID: process.env.WEBAUTHN_RP_ID,
      expectedOrigin: process.env.WEBAUTHN_ORIGIN,
      requireUserVerification: true
    });

    if (!verification.verified) {
      return res.status(400).json({ error: 'Verification failed' });
    }

    // Store credential in database
    await db.webauthnCredentials.create({
      user_id: user.id,
      credentialID: verification.registrationInfo.credentialID,
      credentialPublicKey: verification.registrationInfo.credentialPublicKey,
      counter: verification.registrationInfo.counter,
      credentialDeviceType: verification.registrationInfo.credentialDeviceType,
      credentialBackedUp: verification.registrationInfo.credentialBackedUp,
      transports: credential.response.transports || [],
      created_at: new Date()
    });

    // Clean up challenge
    await redis.del(`webauthn:challenge:${user.id}`);

    res.json({ success: true });
  } catch (error) {
    console.error('Registration verification error:', error);
    res.status(500).json({ error: 'Registration failed' });
  }
}
```

## WebAuthn Authentication Flow

### Client-Side (Browser)

```typescript
// authentication-flow.ts
export async function authenticateWebAuthn(email: string) {
  try {
    // 1. Get authentication options from server
    const optionsResponse = await fetch('/auth/webauthn/authenticate/start', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email })
    });

    const options = await optionsResponse.json();

    // 2. Get credential (user verification)
    const credential = await navigator.credentials.get({
      publicKey: {
        challenge: base64urlDecode(options.challenge),
        rpId: options.rpId,
        allowCredentials: options.allowCredentials.map(cred => ({
          id: base64urlDecode(cred.id),
          type: 'public-key',
          transports: cred.transports
        })),
        timeout: 60000,
        userVerification: 'required'
      }
    });

    // 3. Send assertion to server
    const verifyResponse = await fetch('/auth/webauthn/authenticate/complete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        credential: {
          id: credential.id,
          rawId: base64urlEncode(credential.rawId),
          response: {
            authenticatorData: base64urlEncode(credential.response.authenticatorData),
            clientDataJSON: base64urlEncode(credential.response.clientDataJSON),
            signature: base64urlEncode(credential.response.signature),
            userHandle: credential.response.userHandle 
              ? base64urlEncode(credential.response.userHandle)
              : null
          },
          type: credential.type
        }
      })
    });

    const result = await verifyResponse.json();

    if (result.success) {
      console.log('Authentication successful');
      // Store session token
      localStorage.setItem('session_token', result.token);
    }
  } catch (error) {
    if (error.name === 'NotAllowedError') {
      console.error('User canceled authentication');
    } else {
      console.error('WebAuthn authentication error:', error);
    }
  }
}
```

### Server-Side (Node.js + SimpleWebAuthn)

```typescript
// routes/webauthn-authenticate.ts
import { 
  generateAuthenticationOptions, 
  verifyAuthenticationResponse 
} from '@simplewebauthn/server';

export async function startAuthentication(req, res) {
  const { email } = req.body;
  const user = await db.users.findByEmail(email);

  if (!user) {
    return res.status(404).json({ error: 'User not found' });
  }

  // Get user's registered credentials
  const credentials = await db.webauthnCredentials.findByUserId(user.id);

  if (credentials.length === 0) {
    return res.status(400).json({ error: 'No credentials registered' });
  }

  // Generate authentication options
  const options = await generateAuthenticationOptions({
    rpID: process.env.WEBAUTHN_RP_ID,
    allowCredentials: credentials.map(cred => ({
      id: cred.credentialID,
      transports: cred.transports
    })),
    userVerification: 'required'
  });

  // Store challenge
  await redis.setex(
    `webauthn:auth:${user.id}`,
    900,
    options.challenge
  );

  res.json(options);
}

export async function completeAuthentication(req, res) {
  const { credential } = req.body;

  // Find user by credential ID
  const dbCredential = await db.webauthnCredentials.findByCredentialID(
    credential.id
  );

  if (!dbCredential) {
    return res.status(404).json({ error: 'Credential not found' });
  }

  const user = await db.users.findById(dbCredential.user_id);

  // Get stored challenge
  const expectedChallenge = await redis.get(`webauthn:auth:${user.id}`);

  if (!expectedChallenge) {
    return res.status(400).json({ error: 'Challenge expired' });
  }

  try {
    // Verify authentication response
    const verification = await verifyAuthenticationResponse({
      response: credential,
      expectedChallenge,
      expectedRPID: process.env.WEBAUTHN_RP_ID,
      expectedOrigin: process.env.WEBAUTHN_ORIGIN,
      authenticator: {
        credentialID: dbCredential.credentialID,
        credentialPublicKey: dbCredential.credentialPublicKey,
        counter: dbCredential.counter
      }
    });

    if (!verification.verified) {
      return res.status(401).json({ error: 'Authentication failed' });
    }

    // Update counter (prevents credential cloning)
    await db.webauthnCredentials.update(dbCredential.id, {
      counter: verification.authenticationInfo.newCounter,
      last_used: new Date()
    });

    // Clean up challenge
    await redis.del(`webauthn:auth:${user.id}`);

    // Generate session token
    const sessionToken = jwt.sign(
      { user_id: user.id, email: user.email },
      process.env.JWT_SECRET,
      { expiresIn: '7d' }
    );

    res.json({ success: true, token: sessionToken });
  } catch (error) {
    console.error('Authentication verification error:', error);
    res.status(500).json({ error: 'Authentication failed' });
  }
}
```

## Passkeys (Discoverable Credentials)

### What are Passkeys?

**Passkeys** are WebAuthn credentials that:
- Sync across devices (iCloud Keychain, Google Password Manager, Microsoft)
- Work without username entry (discoverable credentials)
- Support cross-device authentication (scan QR code)

### Passkey Registration

```typescript
// passkey-registration.ts
const options = await generateRegistrationOptions({
  rpID: process.env.WEBAUTHN_RP_ID,
  rpName: 'My Application',
  userID: isoBase64URL.fromBuffer(Buffer.from(user.id)),
  userName: user.email,
  userDisplayName: user.name,
  
  // Passkey-specific settings
  authenticatorSelection: {
    authenticatorAttachment: 'platform', // Device-bound authenticator
    residentKey: 'required', // Make credential discoverable
    userVerification: 'required' // Biometric/PIN required
  },
  
  attestationType: 'none' // Privacy-preserving (no attestation)
});
```

### Passwordless Login (No Username)

```typescript
// passwordless-login.ts
export async function loginWithPasskey() {
  try {
    // 1. Get authentication options (no username required)
    const optionsResponse = await fetch('/auth/passkey/authenticate/start');
    const options = await optionsResponse.json();

    // 2. User selects passkey from device
    const credential = await navigator.credentials.get({
      publicKey: {
        challenge: base64urlDecode(options.challenge),
        rpId: options.rpId,
        allowCredentials: [], // Empty = any passkey works
        timeout: 60000,
        userVerification: 'required'
      }
    });

    // 3. Send assertion to server
    const verifyResponse = await fetch('/auth/passkey/authenticate/complete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ credential })
    });

    const result = await verifyResponse.json();

    if (result.success) {
      console.log('Logged in with passkey');
    }
  } catch (error) {
    console.error('Passkey authentication error:', error);
  }
}
```

## Security Best Practices

### DO ✅

1. **Require user verification** (`userVerification: 'required'`)
2. **Update counter after authentication** (prevents cloning)
3. **Store credentials securely** (encrypted at rest)
4. **Use HTTPS** (WebAuthn requires secure context)
5. **Validate RP ID and origin** (prevents phishing)
6. **Implement challenge timeout** (15 minutes max)

### DON'T ❌

1. ❌ Don't skip user verification (biometric/PIN)
2. ❌ Don't reuse challenges (replay attack)
3. ❌ Don't ignore counter validation (cloning detection)
4. ❌ Don't use weak attestation (use 'direct' or 'none')
5. ❌ Don't store credentials in plain text

## Related Modules

- [JWT & OAuth 2.1](SKILL-jwt-oauth.md) - Token-based authentication
- [MFA Patterns](SKILL.md#multi-factor-authentication-totp) - TOTP, SMS, backup codes

---

**Last Updated**: 2025-11-22  
**Status**: Production Ready (WebAuthn Level 2 compliance)
