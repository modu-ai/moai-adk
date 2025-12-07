# Authentication Platform Reference Documentation

## Auth0 Comprehensive Reference

### Context7 Integration

Context7 ID: /auth0/auth0-docs

Recommended Topics for Documentation Retrieval:

- Enterprise SSO: "enterprise sso saml oidc adfs connections"
- Organizations: "organizations b2b saas multi-tenant rbac"
- Security: "anomaly detection brute-force mfa adaptive"
- APIs: "management api authentication api"
- SDKs: "auth0-python auth0-nextjs sdk integration"

### Enterprise SSO Deep Dive

SAML Configuration Details:

Identity Provider Metadata Requirements:
- Entity ID: Unique identifier for the IdP
- SSO URL: Single Sign-On service endpoint
- Certificate: X.509 certificate for signature verification
- Name ID Format: User identifier format (email, persistent, transient)

Service Provider Configuration (Auth0 as SP):
- Entity ID: urn:auth0:{tenant}:{connection}
- ACS URL: https://{tenant}.auth0.com/login/callback
- Single Logout URL: https://{tenant}.auth0.com/logout

Attribute Mappings:
- email: NameID or email attribute
- given_name: First name attribute
- family_name: Last name attribute
- groups: Group membership for role assignment
- custom attributes: Map to user_metadata or app_metadata

OIDC Federation Configuration:

Discovery Document Requirements:
- issuer: Identity provider issuer URL
- authorization_endpoint: User authorization URL
- token_endpoint: Token exchange URL
- userinfo_endpoint: User profile URL
- jwks_uri: JSON Web Key Set URL

Client Configuration:
- client_id: Application identifier at IdP
- client_secret: Application secret (for confidential clients)
- scope: openid profile email (minimum required)
- response_type: code (for authorization code flow)

ADFS Specific Configuration:

Relying Party Trust Setup:
- Add Auth0 as relying party using federation metadata
- Configure claims rules for user attributes
- Enable access control policies
- Configure logout URLs for single logout

Claims Rules Examples:
- Send LDAP Attributes: Map AD attributes to outgoing claims
- Transform Claims: Convert group membership to roles
- Send Group Membership: Include security groups

### Organizations API Reference

Organization Object Structure:

Core Fields:
- id: Unique organization identifier (org_xxx)
- name: Machine-readable name (lowercase, no spaces)
- display_name: Human-readable display name
- branding: Logo, colors, and theme customization
- metadata: Custom key-value pairs
- enabled_connections: Active identity providers

Organization Roles:
- Admin: Full organization management
- Member: Standard organization access
- Custom roles: Define per organization type

Organization Invitations:
- inviter: User who created invitation
- invitee: Email address of invited user
- roles: Roles to assign upon acceptance
- expires_at: Invitation expiration timestamp
- organization_id: Target organization

Management API Operations:

Create Organization:
- Endpoint: POST /api/v2/organizations
- Required: name, display_name
- Optional: branding, metadata, enabled_connections

List Organizations:
- Endpoint: GET /api/v2/organizations
- Pagination: page, per_page (max 100)
- Filtering: name, metadata values

Add Organization Members:
- Endpoint: POST /api/v2/organizations/{id}/members
- Required: members array with user IDs
- Assign roles during membership creation

Organization Connections:
- Endpoint: POST /api/v2/organizations/{id}/enabled_connections
- Enable specific identity providers per organization
- Configure auto-membership and assignment rules

### Custom Database Connections

Database Action Scripts:

Login Script:
- Validates credentials against external database
- Returns user object with email, user_id, name
- Called during every login attempt

Create Script:
- Creates new user in external database
- Receives email, password (hashed), and profile data
- Called during user registration

Verify Script:
- Marks user email as verified
- Called after email verification flow

Change Password Script:
- Updates password in external database
- Receives user_id and new password
- Called during password reset

Get User Script:
- Retrieves user by email or user_id
- Used for forgot password and user lookup
- Must return consistent user object

Delete Script:
- Removes user from external database
- Called when user deletion requested

Migration Configuration:

Lazy Migration:
- Users migrate on first successful login
- Original database queried only if user not in Auth0
- Gradual migration over time
- Fallback to Auth0 database after migration

Bulk Migration:
- Export users with password hashes
- Import via Management API
- Supported hash formats: bcrypt, pbkdf2, md5, sha256

### Security Configuration

Anomaly Detection Settings:

Brute Force Protection:
- max_attempts: Failed attempts before blocking (default: 10)
- block_duration: Block duration in seconds (default: 3600)
- notification_email: Alert recipient

Breached Password Detection:
- block_on_breach: Block sign-in with breached passwords
- password_policy: Require password change
- notification: Alert user of breach detection

Suspicious IP Throttling:
- stage_1_threshold: First stage throttle limit
- stage_1_delay: Delay after first threshold
- stage_2_threshold: Second stage throttle limit
- stage_2_delay: Delay after second threshold

Bot Detection:
- captcha_provider: hCaptcha, reCAPTCHA, or Auth0
- mode: always, suspicious, or never
- suspicious_ip_threshold: Confidence score threshold

Multi-Factor Authentication:

MFA Providers:
- OTP: Time-based one-time passwords (TOTP)
- SMS: SMS-based verification codes
- Push: Push notifications via Guardian
- WebAuthn: Security keys and biometrics
- Email: Email-based verification codes

Adaptive MFA Configuration:
- Enable based on risk score
- Geographic anomaly detection
- Device fingerprint changes
- Impossible travel detection

---

## Clerk Comprehensive Reference

### Context7 Integration

Context7 ID: /clerk/clerk-docs

Recommended Topics for Documentation Retrieval:

- WebAuthn: "webauthn passkeys passwordless biometric"
- Organizations: "organizations teams invitations roles"
- SDKs: "react nextjs vue expo sdk components"
- Backend: "backend api user management webhooks"
- Security: "session tokens jwt claims security"

### WebAuthn Implementation Details

Passkey Configuration Options:

Verification Requirements:
- required: User must verify identity (biometric/PIN)
- preferred: Request verification but allow skip
- discouraged: Skip verification when possible

Attestation Options:
- none: No attestation required (recommended)
- indirect: Anonymized attestation
- direct: Full attestation from authenticator

Authenticator Selection:
- platform: Built-in authenticator (Touch ID, Face ID)
- cross-platform: External security key (YubiKey)
- any: Allow both platform and cross-platform

Passkey Registration Flow:

Step 1: User initiates passkey setup from profile
Step 2: Clerk generates challenge from backend
Step 3: Browser creates credential with authenticator
Step 4: Credential public key stored in Clerk
Step 5: User can now authenticate with passkey

Passkey Authentication Flow:

Step 1: User initiates sign-in
Step 2: Clerk sends challenge with allowed credentials
Step 3: Authenticator signs challenge with private key
Step 4: Clerk verifies signature against stored public key
Step 5: Session created upon successful verification

### Organizations Deep Dive

Organization Object Structure:

Core Fields:
- id: Unique identifier (org_xxx)
- name: Unique slug for URL routing
- slug: URL-friendly identifier
- image_url: Organization avatar
- max_allowed_memberships: Membership limit
- members_count: Current member count
- created_at: Creation timestamp

Organization Metadata:
- public_metadata: Accessible from frontend
- private_metadata: Server-side only

Membership Object:

Fields:
- id: Membership identifier
- organization: Parent organization
- public_user_data: User information
- role: Assigned role
- created_at: Join timestamp
- updated_at: Last modification

Role Types:
- admin: Full organization management
- basic_member: Standard member access
- Custom roles: Define via Dashboard or API

Invitation Management:

Invitation Fields:
- id: Invitation identifier
- email_address: Invitee email
- organization_id: Target organization
- role: Assigned role upon acceptance
- status: pending, accepted, revoked
- created_at: Invitation creation time
- expires_at: Expiration timestamp

Invitation Flow:
- Admin creates invitation via API or Dashboard
- Clerk sends email with acceptance link
- Invitee clicks link and creates account or signs in
- Membership created with specified role

### SDK Reference

React SDK (@clerk/clerk-react):

Core Hooks:
- useUser(): Current user object and loading state
- useAuth(): Authentication state and methods
- useSession(): Active session information
- useOrganization(): Current organization context
- useOrganizationList(): User's organization memberships
- useSignIn(): Sign-in flow control
- useSignUp(): Sign-up flow control

Provider Configuration:
- ClerkProvider: Root provider with publishableKey
- afterSignInUrl: Redirect after sign-in
- afterSignUpUrl: Redirect after sign-up
- signInUrl: Custom sign-in page URL
- signUpUrl: Custom sign-up page URL

Next.js SDK (@clerk/nextjs):

Middleware Configuration:
- clerkMiddleware: Protect routes automatically
- createRouteMatcher: Define public/protected patterns
- auth(): Server-side authentication helper
- currentUser(): Get full user object server-side

App Router Support:
- Server Components: Use auth() and currentUser()
- Client Components: Use hooks from clerk-react
- API Routes: Use auth() helper function
- Middleware: Protect routes before rendering

Server Actions:
- auth(): Get authentication context
- currentUser(): Full user object
- clerkClient: Backend API access

### Backend API Reference

User Management Endpoints:

Create User:
- Method: POST
- Path: /users
- Body: email_address, phone_number, first_name, last_name
- Optional: password, external_id, public_metadata

Get User:
- Method: GET
- Path: /users/{user_id}
- Returns: Complete user object

Update User:
- Method: PATCH
- Path: /users/{user_id}
- Body: Fields to update

Delete User:
- Method: DELETE
- Path: /users/{user_id}
- Returns: Deleted user object

List Users:
- Method: GET
- Path: /users
- Query: email_address, phone_number, limit, offset

Session Management:

Get Sessions:
- Method: GET
- Path: /sessions
- Query: user_id, client_id, status

Revoke Session:
- Method: POST
- Path: /sessions/{session_id}/revoke

Verify Session Token:
- Method: POST
- Path: /sessions/{session_id}/verify
- Body: token

### Webhook Events

User Lifecycle Events:
- user.created: New user registered
- user.updated: User profile modified
- user.deleted: User account removed

Session Events:
- session.created: New session started
- session.ended: Session terminated
- session.removed: Session revoked

Organization Events:
- organization.created: New organization
- organization.updated: Organization modified
- organization.deleted: Organization removed
- organizationMembership.created: Member added
- organizationMembership.deleted: Member removed

Webhook Security:
- Verify svix-signature header
- Check svix-timestamp for replay attacks
- Use svix-id for idempotency

---

## Firebase Auth Comprehensive Reference

### Context7 Integration

Context7 ID: /firebase/firebase-docs

Recommended Topics for Documentation Retrieval:

- Authentication: "authentication google facebook apple twitter"
- Admin SDK: "admin sdk user management custom claims"
- Security Rules: "security rules firestore storage"
- Cloud Functions: "cloud functions auth triggers"
- Mobile: "ios android flutter react-native"

### Social Provider Configuration

Google Sign-In:

Console Configuration:
- Enable Google provider in Firebase Console
- Configure OAuth consent screen
- Add authorized domains
- Configure iOS/Android client IDs

OAuth Scopes:
- openid: OpenID Connect identifier
- email: User email address
- profile: Name and profile picture

Facebook Login:

App Configuration:
- Create Facebook App in Developer Console
- Enable Facebook Login product
- Configure OAuth redirect URLs
- Add App ID and Secret to Firebase

Permissions:
- email: User email (required)
- public_profile: Name and picture
- Additional permissions require App Review

Apple Sign-In:

Configuration Steps:
- Enable Sign in with Apple in Apple Developer
- Create Service ID for web
- Configure domains and return URLs
- Generate private key for Firebase

Requirements:
- Paid Apple Developer account
- App must implement Apple Sign In on iOS

### Admin SDK Reference

User Management:

Create User:
- createUser(): Create with email/password
- createCustomToken(): Generate custom auth token
- setCustomUserClaims(): Add custom claims

Retrieve Users:
- getUser(): By UID
- getUserByEmail(): By email address
- getUserByPhoneNumber(): By phone
- listUsers(): Paginated list

Modify Users:
- updateUser(): Update profile fields
- setCustomUserClaims(): Modify claims
- revokeRefreshTokens(): Force re-authentication

Delete Users:
- deleteUser(): Remove single user
- deleteUsers(): Bulk delete (up to 1000)

Custom Claims:

Purpose:
- Role-based access control
- Feature flags per user
- Subscription tier identification

Limitations:
- Maximum 1000 bytes per user
- Propagate with token refresh
- Available in security rules

### Cloud Functions Auth Triggers

onCreate Trigger:

Event Data:
- uid: User identifier
- email: User email address
- displayName: User display name
- photoURL: Profile photo URL
- phoneNumber: Phone number
- metadata: Creation and sign-in times
- providerData: Linked providers

Use Cases:
- Initialize user document in Firestore
- Send welcome email
- Set default custom claims
- Sync to external systems

onDelete Trigger:

Event Data:
- uid: Deleted user identifier
- email: User email at deletion
- displayName: Display name

Use Cases:
- Clean up user data
- Remove from external systems
- Archive user content
- Send exit survey

beforeCreate Trigger (Blocking):

Function Return:
- Return undefined to allow creation
- Throw HttpsError to block
- Modify user data before saving

Use Cases:
- Validate email domain
- Block disposable emails
- Require invitation code
- Check against blocklist

beforeSignIn Trigger (Blocking):

Function Return:
- Return undefined to allow sign-in
- Throw HttpsError to block
- Modify session claims

Use Cases:
- Check subscription status
- Enforce 2FA requirement
- Block compromised accounts
- Rate limit sign-in attempts

### Security Rules Integration

Firestore Rules with Auth:

Authentication Check:
- request.auth != null: User is authenticated
- request.auth.uid: Current user ID
- request.auth.token: JWT claims

Custom Claims Access:
- request.auth.token.admin: Admin claim
- request.auth.token.role: Role claim
- request.auth.token.tier: Subscription tier

Common Patterns:
- Owner only: request.auth.uid == resource.data.userId
- Admin access: request.auth.token.admin == true
- Team access: request.auth.uid in resource.data.members

Cloud Storage Rules:

User-Scoped Access:
- Path: users/{userId}/files/{file}
- Rule: request.auth.uid == userId

Public Read, Auth Write:
- Read: true
- Write: request.auth != null

File Size Limits:
- request.resource.size < 5MB

### Mobile SDK Reference

iOS Integration:

Authentication Methods:
- signIn(withEmail:password:): Email/password
- signIn(with:): OAuth provider
- signInAnonymously(): Anonymous auth
- signInWithCustomToken(): Custom token

State Persistence:
- Auth.auth().currentUser: Current user
- Auth.auth().addStateDidChangeListener: State observer

Phone Authentication:
- verifyPhoneNumber(): Send SMS code
- signIn(with:verificationID:verificationCode:): Verify code

Android Integration:

Authentication Methods:
- signInWithEmailAndPassword(): Email/password
- signInWithCredential(): OAuth provider
- signInAnonymously(): Anonymous auth
- signInWithCustomToken(): Custom token

State Observation:
- FirebaseAuth.getInstance().currentUser: Current user
- addAuthStateListener(): State changes

Phone Authentication:
- PhoneAuthProvider.verifyPhoneNumber(): Send SMS
- signInWithCredential(): Verify with code

Flutter Integration (firebase_auth):

Core Methods:
- signInWithEmailAndPassword(): Email/password
- signInWithCredential(): OAuth provider
- signInAnonymously(): Anonymous auth
- signInWithCustomToken(): Custom token

State Stream:
- authStateChanges(): Authentication state
- idTokenChanges(): Token refresh events
- userChanges(): User profile changes

---

## Migration Scripts and Tools

### Auth0 to Clerk Migration

User Export from Auth0:
- Use Management API to export users
- Include user_metadata and app_metadata
- Handle pagination (100 users per request)
- Export organization memberships separately

Data Transformation:
- Map Auth0 user_id to Clerk external_id
- Transform metadata structures
- Convert role assignments
- Handle password hash formats

Import to Clerk:
- Use Backend API for user creation
- Create organizations first
- Add memberships after user creation
- Verify email addresses post-import

### Clerk to Firebase Migration

Export Clerk Users:
- Use Backend API to list users
- Export organization memberships
- Include custom metadata

Import to Firebase:
- Use Admin SDK importUsers()
- Set password hashes with algorithm
- Create custom claims from metadata
- Sync to Firestore for additional data

### Firebase to Auth0 Migration

Export Firebase Users:
- Use Admin SDK listUsers()
- Export with password hashes
- Include custom claims

Import to Auth0:
- Create custom database connection
- Use bulk import with password hashes
- Configure hash algorithm in connection
- Migrate custom claims to metadata

---

## Security Compliance Reference

### GDPR Compliance

Data Subject Rights:
- Right to Access: Export user data on request
- Right to Erasure: Delete user and related data
- Right to Rectification: Update incorrect data
- Right to Portability: Provide data in machine-readable format

Data Processing Agreement:
- Auth0: DPA available for enterprise
- Clerk: DPA included in terms
- Firebase: Google Cloud DPA applies

### SOC 2 Compliance

Auth0:
- SOC 2 Type II certified
- Annual audit reports available
- Enterprise features for compliance

Clerk:
- SOC 2 Type II certified
- Security documentation available
- Enterprise compliance features

Firebase:
- Inherits Google Cloud SOC 2
- Compliance reports via Google

### HIPAA Compliance

Auth0 Enterprise:
- BAA available for healthcare customers
- PHI handling guidelines
- Audit logging requirements

Firebase:
- BAA available via Google Cloud
- Healthcare-specific configuration
- Audit logging required

Clerk:
- Contact for healthcare requirements
- Enterprise plan features
- Custom compliance arrangements

---

Last Updated: 2025-12-07
Context7 Integration: All providers mapped
Status: Production Ready
