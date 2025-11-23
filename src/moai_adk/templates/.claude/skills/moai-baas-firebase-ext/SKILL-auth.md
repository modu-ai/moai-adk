---
name: moai-baas-firebase-auth
description: Firebase Authentication patterns, custom claims, and security implementation
---

## Firebase Authentication & Security

### Advanced Authentication Implementation

```typescript
import { getAuth, Auth } from 'firebase-admin/auth';

export class FirebaseAuthManager {
  private auth: Auth;

  constructor() {
    this.auth = getAuth();
  }

  // User authentication with custom claims
  async authenticateUser(
    uid: string,
    customClaims: Record<string, any> = {}
  ): Promise<AuthResult> {
    try {
      // Set custom claims (roles, permissions)
      await this.auth.setCustomUserClaims(uid, customClaims);

      // Get user record
      const userRecord = await this.auth.getUser(uid);

      return {
        success: true,
        user: {
          uid: userRecord.uid,
          email: userRecord.email,
          displayName: userRecord.displayName,
          photoURL: userRecord.photoURL,
          emailVerified: userRecord.emailVerified,
          customClaims: userRecord.customClaims,
        },
      };
    } catch (error) {
      return {
        success: false,
        error: error.message,
      };
    }
  }

  // Create user with email and password
  async createUser(
    email: string,
    password: string,
    displayName?: string
  ): Promise<CreateUserResult> {
    try {
      const userRecord = await this.auth.createUser({
        email,
        password,
        displayName,
        emailVerified: false,
      });

      // Send verification email
      const link = await this.auth.generateEmailVerificationLink(email);

      return {
        success: true,
        uid: userRecord.uid,
        verificationLink: link,
      };
    } catch (error) {
      return {
        success: false,
        error: error.message,
      };
    }
  }

  // Verify ID token and extract claims
  async verifyToken(idToken: string): Promise<TokenVerificationResult> {
    try {
      const decodedToken = await this.auth.verifyIdToken(idToken);

      return {
        success: true,
        uid: decodedToken.uid,
        email: decodedToken.email,
        customClaims: decodedToken,
      };
    } catch (error) {
      return {
        success: false,
        error: error.message,
      };
    }
  }

  // Set custom user claims (roles)
  async setUserRole(uid: string, role: string): Promise<void> {
    const customClaims = {
      role,
      permissions: this.getPermissionsForRole(role),
    };

    await this.auth.setCustomUserClaims(uid, customClaims);
  }

  private getPermissionsForRole(role: string): string[] {
    const rolePermissions: Record<string, string[]> = {
      admin: ['read', 'write', 'delete', 'manage_users'],
      editor: ['read', 'write'],
      viewer: ['read'],
    };

    return rolePermissions[role] || [];
  }
}
```

---

### Multi-Provider Authentication

```typescript
// Email/Password
const emailProvider = new EmailAuthProvider();

// Google OAuth
const googleProvider = new GoogleAuthProvider();

// Facebook OAuth
const facebookProvider = new FacebookAuthProvider();

// Phone Authentication
const phoneProvider = new PhoneAuthProvider();

// Anonymous Authentication
async function signInAnonymously() {
  const userCredential = await signInAnonymously(auth);
  return userCredential.user;
}
```

---

### Security Best Practices

**Password Requirements**:
```typescript
function validatePassword(password: string): ValidationResult {
  const minLength = 8;
  const hasUpperCase = /[A-Z]/.test(password);
  const hasLowerCase = /[a-z]/.test(password);
  const hasNumber = /\d/.test(password);
  const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(password);

  if (password.length < minLength) {
    return { valid: false, error: 'Password must be at least 8 characters' };
  }

  if (!hasUpperCase) {
    return { valid: false, error: 'Password must contain uppercase letter' };
  }

  if (!hasLowerCase) {
    return { valid: false, error: 'Password must contain lowercase letter' };
  }

  if (!hasNumber) {
    return { valid: false, error: 'Password must contain number' };
  }

  if (!hasSpecialChar) {
    return { valid: false, error: 'Password must contain special character' };
  }

  return { valid: true };
}
```

---

### Token Management

**JWT Token Verification**:
```typescript
async function verifyAndRefreshToken(idToken: string): Promise<string> {
  try {
    // Verify current token
    const decodedToken = await auth.verifyIdToken(idToken);

    // Check expiration
    const now = Math.floor(Date.now() / 1000);
    const timeUntilExpiry = decodedToken.exp - now;

    // Refresh if expiring soon (< 5 minutes)
    if (timeUntilExpiry < 300) {
      const newToken = await auth.createCustomToken(decodedToken.uid);
      return newToken;
    }

    return idToken;
  } catch (error) {
    throw new Error('Token verification failed');
  }
}
```

---

**End of Module** | moai-baas-firebase-auth
