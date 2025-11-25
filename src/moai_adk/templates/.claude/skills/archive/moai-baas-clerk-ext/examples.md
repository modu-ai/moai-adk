# Clerk - Practical Examples

## Example 1: Basic Clerk Setup with React

```jsx
// React + Clerk authentication
import { ClerkProvider, SignIn, SignUp, UserButton } from '@clerk/nextjs';

export default function App() {
    return (
        <ClerkProvider
            publishableKey={process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY}
        >
            <Navigation />
            <Routes>
                <Route path="/sign-in" element={<SignIn />} />
                <Route path="/sign-up" element={<SignUp />} />
                <Route path="/dashboard" element={<Dashboard />} />
            </Routes>
        </ClerkProvider>
    );
}

function Navigation() {
    const { isSignedIn, user } = useAuth();

    return (
        <nav>
            {isSignedIn ? (
                <>
                    <span>Welcome, {user.firstName}</span>
                    <UserButton />
                </>
            ) : (
                <a href="/sign-in">Sign In</a>
            )}
        </nav>
    );
}
```

## Example 2: Protected Routes and Middleware

```typescript
// Next.js middleware for Clerk protection
import { auth } from '@clerk/nextjs/server';
import { NextResponse } from 'next/server';

export async function middleware(request: Request) {
    const { userId } = await auth();

    // Redirect unauthenticated users
    if (!userId) {
        return NextResponse.redirect(new URL('/sign-in', request.url));
    }

    return NextResponse.next();
}

export const config = {
    matcher: ['/dashboard/:path*', '/api/protected/:path*']
};
```

## Example 3: User Data Access

```jsx
// Accessing user data with Clerk
import { useUser, useAuth } from '@clerk/nextjs';

function UserProfile() {
    const { user } = useUser();
    const { getToken } = useAuth();

    if (!user) return <div>Loading...</div>;

    return (
        <div>
            <h1>{user.fullName}</h1>
            <p>Email: {user.emailAddresses[0].emailAddress}</p>
            <img src={user.profileImageUrl} alt="Profile" />

            {user.unsafeMetadata && (
                <p>Preferences: {JSON.stringify(user.unsafeMetadata)}</p>
            )}
        </div>
    );
}
```

## Example 4: Backend API Authentication

```javascript
// Express + Clerk authentication
const { auth } = require('@clerk/express');
const express = require('express');

const app = express();

// Clerk auth middleware
app.use(auth());

app.get('/api/protected', (req, res) => {
    const { userId } = req.auth;

    if (!userId) {
        return res.status(401).json({ error: 'Unauthorized' });
    }

    res.json({
        message: 'Protected route accessed',
        userId: userId
    });
});

app.post('/api/user-data', (req, res) => {
    const { userId } = req.auth;

    // Fetch full user object from Clerk
    const user = req.user;

    res.json({
        user: {
            id: userId,
            email: user.emailAddresses[0].emailAddress,
            name: user.fullName
        }
    });
});
```

## Example 5: Custom User Metadata

```typescript
// Store and manage custom metadata in Clerk
import { useUser } from '@clerk/nextjs';

export function UserPreferences() {
    const { user } = useUser();

    const updatePreferences = async (preferences) => {
        await user.update({
            unsafeMetadata: {
                ...user.unsafeMetadata,
                preferences: {
                    theme: preferences.theme,
                    language: preferences.language,
                    notifications: preferences.notifications
                }
            }
        });
    };

    return (
        <div>
            <button onClick={() => updatePreferences({
                theme: 'dark',
                language: 'ko',
                notifications: true
            })}>
                Save Preferences
            </button>
        </div>
    );
}
```

## Example 6: Social Login Integration

```jsx
// Clerk with social providers
import { SignIn } from '@clerk/nextjs';

export function AuthPage() {
    return (
        <SignIn
            path="/sign-in"
            routing="path"
            signUpUrl="/sign-up"
            socialProviders={['google', 'github', 'github']}
        />
    );
}

// Or programmatically
import { useSignIn } from '@clerk/nextjs';

function SocialLogin() {
    const { signIn } = useSignIn();

    const handleGoogleSignIn = async () => {
        try {
            await signIn.authenticateWithRedirect({
                strategy: 'oauth_google',
                redirectUrl: '/dashboard',
                redirectUrlComplete: '/dashboard'
            });
        } catch (error) {
            console.error('Google sign-in failed:', error);
        }
    };

    return (
        <button onClick={handleGoogleSignIn}>
            Sign in with Google
        </button>
    );
}
```

## Example 7: Organization Management

```typescript
// Clerk Organizations for multi-tenant apps
import { useOrganization, useOrganizations } from '@clerk/nextjs';

function OrgDashboard() {
    const { organization, membership } = useOrganization();
    const { organizations } = useOrganizations();

    return (
        <div>
            <h1>{organization?.name}</h1>
            <p>Your role: {membership?.role}</p>

            <ul>
                {organizations.map(org => (
                    <li key={org.id}>{org.name}</li>
                ))}
            </ul>
        </div>
    );
}

// Create organization
async function createOrg() {
    const { createOrganization } = useOrganizationList();

    await createOrganization({
        name: 'My Team',
        slug: 'my-team'
    });
}
```

## Example 8: Custom Authentication Flow

```python
# Python + Clerk backend authentication
from clerk_sdk import Clerk
from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import HTTPBearer, HTTPAuthCredentials

app = FastAPI()
clerk = Clerk(api_key=os.getenv('CLERK_SECRET_KEY'))
security = HTTPBearer()

async def verify_token(credentials: HTTPAuthCredentials = Depends(security)):
    """Verify Clerk token"""
    try:
        decoded = clerk.decode_token(credentials.credentials)
        return decoded
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid token")

@app.get("/protected")
async def protected_route(user = Depends(verify_token)):
    return {"message": f"Hello {user['email']}"}

@app.post("/create-user")
async def create_user(email: str, password: str):
    """Create user via Clerk API"""
    user = clerk.users.create(
        email_address=email,
        password=password
    )
    return {"user_id": user.id}
```

## Example 9: JWT Verification

```javascript
// Verify Clerk JWT tokens
import jwt from 'jsonwebtoken';
import { createClerkClient } from '@clerk/backend';

const clerk = createClerkClient({
    secretKey: process.env.CLERK_SECRET_KEY
});

async function verifyToken(token) {
    try {
        // Get Clerk's JWKS (public keys)
        const jwks = await clerk.interstitial.getInterstitialResource();

        // Decode and verify token
        const decoded = jwt.verify(token, jwks, {
            algorithms: ['RS256']
        });

        return decoded;
    } catch (error) {
        console.error('Token verification failed:', error);
        return null;
    }
}
```

## Example 10: Session Management

```typescript
// Clerk session handling
import { useSession } from '@clerk/nextjs';

function SessionInfo() {
    const { session, isLoaded } = useSession();

    if (!isLoaded) return <div>Loading...</div>;

    if (!session) return <div>Not signed in</div>;

    return (
        <div>
            <p>Session ID: {session.id}</p>
            <p>Status: {session.status}</p>
            <p>Created: {new Date(session.createdAt).toLocaleString()}</p>
        </div>
    );
}
```

## Example 11: Email and SMS Configuration

```javascript
// Configure email and SMS in Clerk
const clerkConfig = {
    // Email configuration
    email: {
        from_email_address: 'noreply@myapp.com',
        from_email_name: 'My App',

        // Email templates
        templates: {
            sign_in_code: {
                subject: 'Your sign-in code: {{code}}',
                body: '<p>Use this code: {{code}}</p>'
            },
            sign_up_verification: {
                subject: 'Verify your email',
                body: '<a href="{{url}}">Verify email</a>'
            }
        }
    },

    // SMS configuration
    sms: {
        from_phone_number: '+12025551234',
        templates: {
            sign_in_code: 'Your sign-in code: {{code}}'
        }
    }
};
```

## Example 12: Webhook Integration

```javascript
// Clerk webhooks for user events
import { Webhook } from 'svix';
import express from 'express';

const app = express();

app.post('/webhooks/clerk', express.raw({type: 'application/json'}), async (req, res) => {
    const secret = process.env.WEBHOOK_SECRET;
    const wh = new Webhook(secret);

    const evt = wh.verify(req.body, req.headers);

    switch (evt.type) {
        case 'user.created':
            console.log('User created:', evt.data.id);
            // Sync user to database
            await createUserInDB(evt.data);
            break;

        case 'user.updated':
            console.log('User updated:', evt.data.id);
            await updateUserInDB(evt.data);
            break;

        case 'user.deleted':
            console.log('User deleted:', evt.data.id);
            await deleteUserFromDB(evt.data.id);
            break;

        default:
            console.log('Unhandled event type:', evt.type);
    }

    res.json({ received: true });
});
```

## Example 13: Rate Limiting

```python
# Clerk API rate limiting
from functools import wraps
import time

class ClerkRateLimiter:
    def __init__(self, calls_per_second=10):
        self.calls_per_second = calls_per_second
        self.last_call = time.time()
        self.call_count = 0

    def __call__(self, func):
        @wraps(func)
        async def wrapper(*args, **kwargs):
            now = time.time()
            time_passed = now - self.last_call

            if time_passed < 1:
                self.call_count += 1
                if self.call_count > self.calls_per_second:
                    wait_time = 1 - time_passed
                    await asyncio.sleep(wait_time)
            else:
                self.call_count = 1
                self.last_call = now

            return await func(*args, **kwargs)

        return wrapper
```

## Example 14: Error Handling and Recovery

```typescript
// Robust error handling with Clerk
class ClerkErrorHandler {
    async handleClerkError(error: any) {
        switch (error.code) {
            case 'validation_failed':
                return { status: 400, message: 'Invalid input data' };
            case 'not_found':
                return { status: 404, message: 'Resource not found' };
            case 'authentication_failed':
                return { status: 401, message: 'Authentication failed' };
            case 'rate_limit':
                return { status: 429, message: 'Rate limit exceeded' };
            case 'server_error':
                return { status: 500, message: 'Server error occurred' };
            default:
                return { status: 500, message: 'Unknown error' };
        }
    }

    async retryWithExponentialBackoff(operation, maxRetries = 3) {
        let lastError;

        for (let attempt = 0; attempt < maxRetries; attempt++) {
            try {
                return await operation();
            } catch (error) {
                lastError = error;

                // Don't retry validation errors
                if (error.code === 'validation_failed') {
                    throw error;
                }

                const backoffMs = Math.pow(2, attempt) * 1000;
                await new Promise(resolve =>
                    setTimeout(resolve, backoffMs)
                );
            }
        }

        throw lastError;
    }
}
```

## Example 15: Passwordless Authentication

```javascript
// Passwordless authentication with Clerk
class ClerkPasswordless {
    async signInWithEmail(email) {
        const { client } = window.Clerk;

        const response = await client.signUp.create({
            strategy: 'email_code',
            identifier: email
        });

        return response;
    }

    async verifyEmailCode(code) {
        const { client } = window.Clerk;

        const response = await client.signUp.attemptEmailAddressVerification({
            code
        });

        return response;
    }

    async signInWithPhoneNumber(phoneNumber) {
        const { client } = window.Clerk;

        const response = await client.signUp.create({
            strategy: 'phone_code',
            identifier: phoneNumber
        });

        return response;
    }

    async verifyPhoneCode(code) {
        const { client } = window.Clerk;

        const response = await client.signUp.attemptPhoneNumberVerification({
            code
        });

        return response;
    }
}
```

## Example 16: Profile Management

```jsx
// User profile management with Clerk
import { useUser } from '@clerk/nextjs';

export function ProfileManager() {
    const { user, isLoaded } = useUser();
    const [formData, setFormData] = useState({
        firstName: '',
        lastName: '',
        username: ''
    });

    useEffect(() => {
        if (user) {
            setFormData({
                firstName: user.firstName || '',
                lastName: user.lastName || '',
                username: user.username || ''
            });
        }
    }, [user]);

    const handleUpdate = async (e) => {
        e.preventDefault();

        try {
            await user.update({
                firstName: formData.firstName,
                lastName: formData.lastName,
                username: formData.username
            });

            alert('Profile updated successfully');
        } catch (error) {
            console.error('Profile update failed:', error);
        }
    };

    if (!isLoaded) return <div>Loading...</div>;

    return (
        <form onSubmit={handleUpdate}>
            <input
                type="text"
                placeholder="First Name"
                value={formData.firstName}
                onChange={(e) => setFormData({
                    ...formData,
                    firstName: e.target.value
                })}
            />

            <input
                type="text"
                placeholder="Last Name"
                value={formData.lastName}
                onChange={(e) => setFormData({
                    ...formData,
                    lastName: e.target.value
                })}
            />

            <button type="submit">Update Profile</button>
        </form>
    );
}
```

## Example 17: Image Upload and Management

```typescript
// User image upload with Clerk
class ClerkImageManagement {
    async uploadProfileImage(userId: string, imageFile: File) {
        const formData = new FormData();
        formData.append('file', imageFile);

        const response = await fetch(
            `https://api.clerk.dev/v1/users/${userId}/profile_image`,
            {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.apiKey}`
                },
                body: formData
            }
        );

        return response.json();
    }

    async deleteProfileImage(userId: string) {
        const response = await fetch(
            `https://api.clerk.dev/v1/users/${userId}/profile_image`,
            {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${this.apiKey}`
                }
            }
        );

        return response.ok;
    }
}
```

## Example 18: Two-Factor Authentication Setup

```python
# Clerk 2FA management
class Clerk2FA:
    async enable_backup_codes(user_id: str):
        """Enable backup codes for user"""
        response = await fetch(
            f'https://api.clerk.dev/v1/users/{user_id}/backup_codes',
            method='POST',
            headers={
                'Authorization': f'Bearer {self.api_key}'
            }
        )

        return response.json()

    async list_backup_codes(user_id: str):
        """List backup codes"""
        response = await fetch(
            f'https://api.clerk.dev/v1/users/{user_id}/backup_codes',
            headers={
                'Authorization': f'Bearer {self.api_key}'
            }
        )

        return response.json()
```

---

**Context7 Integration**: Use `/clerk/clerk-js`, `/clerk/clerk-react`, `/clerk/clerk-nextjs` for latest SDK documentation.

**Version**: 1.0.0 | **Last Updated**: 2025-11-22
