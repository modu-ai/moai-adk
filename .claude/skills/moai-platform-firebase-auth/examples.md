# Firebase Authentication Examples

Working code examples for common Firebase Authentication patterns across multiple platforms.

---

## Complete Authentication Service (TypeScript/Web)

```typescript
import {
  getAuth,
  signInWithEmailAndPassword,
  createUserWithEmailAndPassword,
  signInWithPopup,
  signOut,
  sendPasswordResetEmail,
  sendEmailVerification,
  updateProfile,
  onAuthStateChanged,
  GoogleAuthProvider,
  FacebookAuthProvider,
  User,
  UserCredential
} from 'firebase/auth';

class AuthService {
  private auth = getAuth();
  private googleProvider = new GoogleAuthProvider();
  private facebookProvider = new FacebookAuthProvider();

  // Email/Password Authentication
  async signUp(email: string, password: string, displayName: string): Promise<User> {
    const credential = await createUserWithEmailAndPassword(this.auth, email, password);
    await updateProfile(credential.user, { displayName });
    await sendEmailVerification(credential.user);
    return credential.user;
  }

  async signIn(email: string, password: string): Promise<User> {
    const credential = await signInWithEmailAndPassword(this.auth, email, password);
    return credential.user;
  }

  // Social Authentication
  async signInWithGoogle(): Promise<User> {
    this.googleProvider.addScope('profile');
    this.googleProvider.addScope('email');
    const credential = await signInWithPopup(this.auth, this.googleProvider);
    return credential.user;
  }

  async signInWithFacebook(): Promise<User> {
    this.facebookProvider.addScope('email');
    this.facebookProvider.addScope('public_profile');
    const credential = await signInWithPopup(this.auth, this.facebookProvider);
    return credential.user;
  }

  // Session Management
  async signOut(): Promise<void> {
    await signOut(this.auth);
  }

  async resetPassword(email: string): Promise<void> {
    await sendPasswordResetEmail(this.auth, email);
  }

  // Auth State Observer
  onAuthStateChange(callback: (user: User | null) => void): () => void {
    return onAuthStateChanged(this.auth, callback);
  }

  // Current User
  getCurrentUser(): User | null {
    return this.auth.currentUser;
  }

  // Claims
  async getUserClaims(): Promise<Record<string, any> | undefined> {
    const user = this.auth.currentUser;
    if (!user) return undefined;
    const tokenResult = await user.getIdTokenResult();
    return tokenResult.claims;
  }

  async isAdmin(): Promise<boolean> {
    const claims = await this.getUserClaims();
    return claims?.admin === true;
  }
}

export const authService = new AuthService();
```

---

## React Authentication Hook

```typescript
import { useState, useEffect, createContext, useContext, ReactNode } from 'react';
import { User, onAuthStateChanged, getAuth } from 'firebase/auth';
import { authService } from './authService';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  isAdmin: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signUp: (email: string, password: string, name: string) => Promise<void>;
  signOut: () => Promise<void>;
  signInWithGoogle: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    const unsubscribe = authService.onAuthStateChange(async (user) => {
      setUser(user);
      if (user) {
        const adminStatus = await authService.isAdmin();
        setIsAdmin(adminStatus);
      } else {
        setIsAdmin(false);
      }
      setLoading(false);
    });

    return unsubscribe;
  }, []);

  const value: AuthContextType = {
    user,
    loading,
    isAdmin,
    signIn: async (email, password) => {
      await authService.signIn(email, password);
    },
    signUp: async (email, password, name) => {
      await authService.signUp(email, password, name);
    },
    signOut: async () => {
      await authService.signOut();
    },
    signInWithGoogle: async () => {
      await authService.signInWithGoogle();
    }
  };

  return (
    <AuthContext.Provider value={value}>
      {!loading && children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
```

---

## Flutter Authentication Provider

```dart
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';
import 'package:google_sign_in/google_sign_in.dart';

class AuthProvider extends ChangeNotifier {
  final FirebaseAuth _auth = FirebaseAuth.instance;
  final GoogleSignIn _googleSignIn = GoogleSignIn();

  User? get currentUser => _auth.currentUser;
  bool get isAuthenticated => currentUser != null;

  Stream<User?> get authStateChanges => _auth.authStateChanges();

  // Email/Password Sign Up
  Future<User?> signUp({
    required String email,
    required String password,
    required String displayName,
  }) async {
    final credential = await _auth.createUserWithEmailAndPassword(
      email: email,
      password: password,
    );

    await credential.user?.updateDisplayName(displayName);
    await credential.user?.sendEmailVerification();

    notifyListeners();
    return credential.user;
  }

  // Email/Password Sign In
  Future<User?> signIn({
    required String email,
    required String password,
  }) async {
    final credential = await _auth.signInWithEmailAndPassword(
      email: email,
      password: password,
    );

    notifyListeners();
    return credential.user;
  }

  // Google Sign In
  Future<User?> signInWithGoogle() async {
    final googleUser = await _googleSignIn.signIn();
    if (googleUser == null) return null;

    final googleAuth = await googleUser.authentication;
    final credential = GoogleAuthProvider.credential(
      accessToken: googleAuth.accessToken,
      idToken: googleAuth.idToken,
    );

    final userCredential = await _auth.signInWithCredential(credential);
    notifyListeners();
    return userCredential.user;
  }

  // Sign Out
  Future<void> signOut() async {
    await Future.wait([
      _auth.signOut(),
      _googleSignIn.signOut(),
    ]);
    notifyListeners();
  }

  // Password Reset
  Future<void> resetPassword(String email) async {
    await _auth.sendPasswordResetEmail(email: email);
  }

  // Get User Claims
  Future<Map<String, dynamic>?> getUserClaims() async {
    final user = currentUser;
    if (user == null) return null;

    final tokenResult = await user.getIdTokenResult();
    return tokenResult.claims;
  }

  // Check Admin Status
  Future<bool> isAdmin() async {
    final claims = await getUserClaims();
    return claims?['admin'] == true;
  }
}
```

---

## Cloud Functions: Admin User Management

```typescript
import { onCall, HttpsError } from 'firebase-functions/v2/https';
import { getAuth } from 'firebase-admin/auth';
import { getFirestore, FieldValue } from 'firebase-admin/firestore';

// List all users (admin only)
export const listUsers = onCall(async (request) => {
  if (!request.auth?.token.admin) {
    throw new HttpsError('permission-denied', 'Admin required');
  }

  const { pageToken, pageSize = 100 } = request.data;

  const listResult = await getAuth().listUsers(pageSize, pageToken);

  return {
    users: listResult.users.map(user => ({
      uid: user.uid,
      email: user.email,
      displayName: user.displayName,
      disabled: user.disabled,
      emailVerified: user.emailVerified,
      customClaims: user.customClaims,
      createdAt: user.metadata.creationTime
    })),
    pageToken: listResult.pageToken
  };
});

// Set user role (admin only)
export const setUserRole = onCall(async (request) => {
  if (!request.auth?.token.admin) {
    throw new HttpsError('permission-denied', 'Admin required');
  }

  const { uid, role } = request.data;

  const validRoles = ['viewer', 'member', 'editor', 'admin'];
  if (!validRoles.includes(role)) {
    throw new HttpsError('invalid-argument', 'Invalid role');
  }

  // Get current claims and preserve them
  const user = await getAuth().getUser(uid);
  const currentClaims = user.customClaims || {};

  await getAuth().setCustomUserClaims(uid, {
    ...currentClaims,
    role
  });

  // Log the change
  await getFirestore().collection('auditLogs').add({
    action: 'SET_USER_ROLE',
    targetUid: uid,
    newRole: role,
    performedBy: request.auth.uid,
    timestamp: FieldValue.serverTimestamp()
  });

  return { success: true };
});

// Disable user (admin only)
export const disableUser = onCall(async (request) => {
  if (!request.auth?.token.admin) {
    throw new HttpsError('permission-denied', 'Admin required');
  }

  const { uid, disabled } = request.data;

  await getAuth().updateUser(uid, { disabled });

  await getFirestore().collection('auditLogs').add({
    action: disabled ? 'DISABLE_USER' : 'ENABLE_USER',
    targetUid: uid,
    performedBy: request.auth.uid,
    timestamp: FieldValue.serverTimestamp()
  });

  return { success: true };
});
```

---

## Protected Route Component (React)

```typescript
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from './useAuth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireAdmin?: boolean;
}

export function ProtectedRoute({ children, requireAdmin = false }: ProtectedRouteProps) {
  const { user, loading, isAdmin } = useAuth();
  const location = useLocation();

  if (loading) {
    return <div>Loading...</div>;
  }

  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (requireAdmin && !isAdmin) {
    return <Navigate to="/unauthorized" replace />;
  }

  return <>{children}</>;
}

// Usage
function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/dashboard" element={
        <ProtectedRoute>
          <DashboardPage />
        </ProtectedRoute>
      } />
      <Route path="/admin" element={
        <ProtectedRoute requireAdmin>
          <AdminPage />
        </ProtectedRoute>
      } />
    </Routes>
  );
}
```

---

## Server-Side Token Verification (Express)

```typescript
import express from 'express';
import { getAuth } from 'firebase-admin/auth';

const app = express();

// Middleware to verify Firebase ID token
async function verifyToken(
  req: express.Request,
  res: express.Response,
  next: express.NextFunction
) {
  const authHeader = req.headers.authorization;

  if (!authHeader?.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Unauthorized' });
  }

  const idToken = authHeader.split('Bearer ')[1];

  try {
    const decodedToken = await getAuth().verifyIdToken(idToken);
    req.user = decodedToken;
    next();
  } catch (error) {
    return res.status(401).json({ error: 'Invalid token' });
  }
}

// Middleware to require admin
function requireAdmin(
  req: express.Request,
  res: express.Response,
  next: express.NextFunction
) {
  if (!req.user?.admin) {
    return res.status(403).json({ error: 'Admin required' });
  }
  next();
}

// Protected routes
app.get('/api/profile', verifyToken, (req, res) => {
  res.json({ uid: req.user.uid, email: req.user.email });
});

app.get('/api/admin/users', verifyToken, requireAdmin, async (req, res) => {
  const users = await getAuth().listUsers(100);
  res.json(users);
});
```

---

## Anonymous to Permanent Account Upgrade

```typescript
import {
  signInAnonymously,
  linkWithCredential,
  EmailAuthProvider,
  GoogleAuthProvider,
  linkWithPopup
} from 'firebase/auth';

// Start as anonymous
const startAnonymousSession = async () => {
  const result = await signInAnonymously(auth);
  console.log('Anonymous UID:', result.user.uid);
  return result.user;
};

// Upgrade to email/password
const upgradeWithEmail = async (email: string, password: string) => {
  const user = auth.currentUser;
  if (!user?.isAnonymous) {
    throw new Error('User is not anonymous');
  }

  const credential = EmailAuthProvider.credential(email, password);
  const result = await linkWithCredential(user, credential);
  console.log('Account upgraded:', result.user.uid);
  return result.user;
};

// Upgrade with Google
const upgradeWithGoogle = async () => {
  const user = auth.currentUser;
  if (!user?.isAnonymous) {
    throw new Error('User is not anonymous');
  }

  const provider = new GoogleAuthProvider();
  const result = await linkWithPopup(user, provider);
  console.log('Account upgraded:', result.user.uid);
  return result.user;
};
```

---

Version: 1.0.0
Last Updated: 2025-12-07
Parent Skill: moai-platform-firebase-auth
