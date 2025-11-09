# Skill: moai-baas-clerk-ext

## Metadata

```yaml
skill_id: moai-baas-clerk-ext
skill_name: Clerk Authentication & User Management
version: 1.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["Clerk", "Authentication", "MFA", "User Management", "SSO", "Modern Auth"]
  - contexts: ["clerk-detected", "pattern-b", "auth-modern"]
agents:
  - security-expert
  - backend-expert
  - frontend-expert
freedom_level: high
word_count: 1000
context7_references:
  - url: "https://clerk.com/docs/quickstarts/setup-clerk"
    topic: "Clerk Quick Start Setup"
  - url: "https://clerk.com/docs/reference/backend-api"
    topic: "Backend API Reference"
  - url: "https://clerk.com/docs/custom-flows/overview"
    topic: "Custom Authentication Flows"
  - url: "https://clerk.com/docs/users/multi-tenancy"
    topic: "Multi-Tenancy & Organizations"
  - url: "https://clerk.com/docs/deployments/clerk-managed"
    topic: "Deployment & Hosting"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Clerk Architecture & Advantages (150 words)

**Clerk** is a modern authentication platform optimized for web and mobile with built-in multi-factor authentication and user management.

**Clerk vs Alternatives**:

```
                  Clerk        Auth0        Supabase Auth
Developer UX      ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐      ⭐⭐⭐
MFA/SSO          ⭐⭐⭐⭐⭐    ⭐⭐⭐⭐⭐    ⭐⭐
UI Components    ⭐⭐⭐⭐⭐    ⭐⭐         ⭐⭐
Pricing          Medium      High        Low
Setup Time       5 minutes   30 minutes  10 minutes
Multi-tenancy    ⭐⭐⭐⭐⭐    ⭐⭐⭐        ⭐⭐
```

**Core Components**:
```
┌──────────────────────────────┐
│ Clerk (Modern Auth Platform) │
├──────────────────────────────┤
│ 1. Frontend Components       │
│    └─ Pre-built UI (Sign in, Sign up)
│                              │
│ 2. Backend Verification      │
│    └─ JWT validation         │
│                              │
│ 3. Multi-Factor Auth         │
│    └─ TOTP, SMS, Backup codes
│                              │
│ 4. Organizations             │
│    └─ Multi-tenancy support  │
│                              │
│ 5. User Management           │
│    └─ Roles, permissions     │
└──────────────────────────────┘
```

---

### 2. Frontend Integration with React (250 words)

**Setup with Clerk React SDK**:

```typescript
// app.tsx
import { ClerkProvider } from "@clerk/clerk-react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";

const clerkPubKey = process.env.REACT_APP_CLERK_PUBLISHABLE_KEY;

const router = createBrowserRouter([
  {
    path: "/",
    element: <Dashboard />,
    loader: requireAuth,
  },
  {
    path: "/sign-in",
    element: <SignIn />,
  },
]);

export default function App() {
  return (
    <ClerkProvider publishableKey={clerkPubKey}>
      <RouterProvider router={router} />
    </ClerkProvider>
  );
}
```

**Using Clerk Hooks**:

```typescript
import { useAuth, useUser, useSignUp, useSignIn } from "@clerk/clerk-react";

export function UserProfile() {
  const { user, isLoaded } = useUser();
  const { userId } = useAuth();

  if (!isLoaded) return <Skeleton />;
  if (!user) return <SignIn />;

  return (
    <div>
      <h1>{user.firstName} {user.lastName}</h1>
      <p>{user.primaryEmailAddress?.emailAddress}</p>
      <img src={user.profileImageUrl} alt="Profile" />
    </div>
  );
}

export function CustomSignIn() {
  const { signIn, setActive, isLoaded } = useSignIn();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const result = await signIn?.create({
        identifier: email,
        password,
      });

      if (result?.status === "complete") {
        await setActive?.({ session: result.createdSessionId });
      }
    } catch (err: any) {
      console.error(err.errors[0].message);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email address"
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
      />
      <button type="submit">Sign In</button>
    </form>
  );
}
```

**Protected Routes**:

```typescript
import { ProtectedLayout } from "@clerk/clerk-react";

function RequireAuth({ children }) {
  return (
    <ProtectedLayout>
      {children}
    </ProtectedLayout>
  );
}

// Usage
<Routes>
  <Route path="/dashboard" element={<RequireAuth><Dashboard /></RequireAuth>} />
</Routes>
```

---

### 3. Backend Integration & JWT Verification (200 words)

**Express/Node.js Backend**:

```typescript
import { ClerkExpressRequireAuth } from "@clerk/clerk-sdk-node";

const app = express();

// Middleware: Verify JWT token
app.use(ClerkExpressRequireAuth());

// Protected endpoint
app.get("/api/profile", (req, res) => {
  const userId = req.auth.userId;
  const sessionId = req.auth.sessionId;

  res.json({ userId, sessionId });
});

// Get user details
app.get("/api/user", async (req, res) => {
  try {
    const user = await clerkClient.users.getUser(req.auth.userId);
    res.json(user);
  } catch (error) {
    res.status(401).json({ error: "Unauthorized" });
  }
});

// Create user metadata
app.post("/api/user/metadata", async (req, res) => {
  await clerkClient.users.updateUser(req.auth.userId, {
    publicMetadata: {
      role: "premium",
      company: "Acme Corp",
    },
  });

  res.json({ success: true });
});
```

**User Management Operations**:

```typescript
import { clerkClient } from "@clerk/clerk-sdk-node";

// List users
const users = await clerkClient.users.getUserList();

// Get specific user
const user = await clerkClient.users.getUser(userId);

// Update user
await clerkClient.users.updateUser(userId, {
  firstName: "John",
  lastName: "Doe",
  metadata: {
    role: "admin",
  },
});

// Delete user
await clerkClient.users.deleteUser(userId);

// Verify token
const token = req.headers.authorization?.split(" ")[1];
const payload = await clerkClient.verifyToken(token);
```

---

### 4. Multi-Factor Authentication Setup (150 words)

**MFA Configuration**:

```typescript
import { useAuth } from "@clerk/clerk-react";

export function EnableMFA() {
  const { user } = useAuth();
  const [totpUri, setTotpUri] = useState<string>("");

  const handleSetupTOTP = async () => {
    // Step 1: Generate TOTP secret
    const { totpSecret } = await user?.createTOTPSecret();

    // Step 2: Show QR code (or secret)
    setTotpUri(totpSecret?.uri || "");
  };

  const handleVerifyTOTP = async (code: string) => {
    // Step 3: Verify code from authenticator app
    try {
      await user?.verifyTOTPSecret({
        code,
      });

      console.log("MFA enabled!");
    } catch (error) {
      console.error("Invalid code");
    }
  };

  return (
    <div>
      <button onClick={handleSetupTOTP}>Setup 2FA</button>
      {totpUri && <QRCode value={totpUri} />}
      <input
        type="text"
        placeholder="Enter 6-digit code"
        onBlur={(e) => handleVerifyTOTP(e.target.value)}
      />
    </div>
  );
}
```

**Backup Codes** (Recovery):

```typescript
// Generate backup codes
const backupCodes = await user?.createBackupCode();

// Display to user (store securely)
backupCodes?.codes.forEach((code) => console.log(code));

// User can use backup code instead of TOTP if phone lost
```

---

### 5. Organizations & Multi-Tenancy (150 words)

**Organization Setup**:

```typescript
// Create organization
const org = await clerkClient.organizations.createOrganization({
  name: "Acme Corp",
  createdBy: userId,
});

// Add member to organization
await clerkClient.organizations.createOrganizationMembership(org.id, {
  userId,
  role: "member", // or "admin"
});

// Organization roles in frontend
export function OrgDashboard() {
  const { user } = useAuth();
  const { userMemberships } = user;

  return (
    <div>
      {userMemberships.map((membership) => (
        <div key={membership.id}>
          <h3>{membership.organization.name}</h3>
          <p>Role: {membership.role}</p>
        </div>
      ))}
    </div>
  );
}
```

**Multi-Tenant Architecture**:
- ✅ Organization-scoped data isolation
- ✅ Role-based access control (admin, member, viewer)
- ✅ Invitation system for adding users
- ✅ Custom permissions per organization

---

### 6. Production Deployment & Pricing (150 words)

**Clerk Pricing Model**:

```
Free Tier:        10,000 MAU/month
Pro Tier:         $25/month + $0.02 per additional MAU
Enterprise:       Custom pricing

Includes:
✅ All authentication methods
✅ MFA & SSO
✅ Unlimited organizations
✅ Custom domains
```

**Deployment Configuration**:

```bash
# Environment variables
REACT_APP_CLERK_PUBLISHABLE_KEY=pk_test_xxx
CLERK_SECRET_KEY=sk_test_xxx

# Production domain setup
# Clerk Dashboard → Domains → Add Production Domain
# clerk.example.com → your-domain.example.com
```

**Monitoring & Troubleshooting**:

```typescript
// Monitor authentication events
import { useClerk } from "@clerk/clerk-react";

const { client } = useClerk();
client?.on("sessionEnd", () => {
  console.log("Session ended");
  // Handle logout
});

// Error handling
try {
  await signIn?.create({...});
} catch (err: any) {
  const errors = err.errors;
  errors.forEach((error: any) => {
    console.error(error.message);
    // Handle validation, rate limit, etc.
  });
}
```

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-clerk-ext")
# Load when Pattern B (Neon + Clerk + Vercel) detected
```

### Context7 Integration
When Clerk detected:
- Frontend sign-in/sign-up components
- Backend JWT verification
- Multi-factor authentication setup
- Organizations & multi-tenancy

---

## 📚 Reference Materials

- [Clerk Quick Start](https://clerk.com/docs/quickstarts/setup-clerk)
- [Backend API Reference](https://clerk.com/docs/reference/backend-api)
- [Custom Authentication Flows](https://clerk.com/docs/custom-flows/overview)
- [Multi-Tenancy & Organizations](https://clerk.com/docs/users/multi-tenancy)
- [Deployment & Hosting](https://clerk.com/docs/deployments/clerk-managed)

---

## ✅ Validation Checklist

- [x] Clerk architecture & advantages
- [x] Frontend integration with React
- [x] Backend verification & JWT
- [x] Multi-factor authentication setup
- [x] Organizations & multi-tenancy
- [x] Production deployment & pricing
- [x] 1000-word target
- [x] English language (policy compliant)
