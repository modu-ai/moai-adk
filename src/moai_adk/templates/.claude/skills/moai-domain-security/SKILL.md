---
name: moai-domain-security
description: Enterprise-grade security architecture expertise with OAuth 2.1, zero-trust implementation, Argon2id password hashing, OWASP Top 10 2021 compliance, automated security testing with Snyk and OWASP ZAP; activates for security design, threat modeling, vulnerability management, authentication flows, API security, secure session management, CSRF/XSS protection, compliance automation (ISO 27001:2022, SOC 2, GDPR), and comprehensive security strategy development with NextAuth.js 5.x, Passport.js 0.7.x, and Helmet.js 7.x.
allowed-tools:
  - Read
  - Bash
  - WebSearch
  - WebFetch
  - mcp__context7__get-library-docs
  - mcp__context7__resolve-library-id
---

# 🛡️ Enterprise Security Architecture & Zero-Trust Implementation

## 🎯 Skill Metadata
| Field | Value |
| ----- | ----- |
| **Version** | **4.0.0 Enterprise** |
| **Created** | 2025-11-12 |
| **Updated** | 2025-11-12 |
| **Allowed tools** | Read, Bash, WebSearch, WebFetch, Context7 MCP |
| **Auto-load** | On-demand for security architecture requests |
| **Trigger cues** | Security design, authentication, authorization, threat modeling, vulnerability assessment, OWASP compliance, zero-trust architecture, password hashing, session management, API security, CSRF protection, XSS prevention, security headers, compliance automation, penetration testing |
| **Tier** | **4 (Enterprise)** |
| **Lines** | 930 lines |
| **Size** | ~30KB |
| **Stable Versions** | NextAuth.js 5.x, Passport.js 0.7.x, Helmet.js 7.x, Argon2id, OWASP Top 10 2021, ISO 27001:2022 |

## 🚀 Enterprise Security Capabilities

**Core Security Focus** (Stable Technologies - November 2025):
- OAuth 2.1 + PKCE authorization flows with NextAuth.js 5.x
- Zero-trust architecture with continuous verification
- Argon2id password hashing (OWASP recommended parameters: m=19456, t=2, p=1)
- OWASP Top 10 2021 compliance and mitigation (A01-A10 covered)
- Secure session management with JWT RS256 (asymmetric cryptography)
- CSRF protection (double-submit cookie pattern)
- XSS prevention (Content Security Policy, DOMPurify)
- SQL Injection prevention (parameterized queries, ORMs)
- API rate limiting (sliding window, token bucket)
- Security headers with Helmet.js 7.x (HSTS, CSP, COOP, CORP)
- Secret management (HashiCorp Vault, AWS Secrets Manager)
- Automated security testing (Snyk 1.x, OWASP ZAP 2.15.x, Trivy 0.58.x, Semgrep 1.x)
- Compliance automation (ISO 27001:2022, SOC 2 Type II, GDPR, PCI DSS 4.0)

## 🏗️ Architecture Overview

### **Security Architecture Layers**

```
🛡️ Enterprise Security Stack (Stable Versions):
├── Authentication & Authorization Layer
│   ├── NextAuth.js 5.x (OAuth 2.1 + PKCE)
│   ├── Passport.js 0.7.x (Strategy-based auth)
│   ├── JWT RS256 Signatures (asymmetric keys)
│   └── Argon2id Password Hashing (OWASP params)
├── Network Security Layer
│   ├── Zero-trust micro-segmentation
│   ├── API Gateway with rate limiting
│   ├── WAF (Web Application Firewall)
│   └── DDoS protection
├── Application Security Layer
│   ├── CSRF Protection (double-submit cookie)
│   ├── XSS Prevention (Content Security Policy)
│   ├── SQL Injection Prevention (ORMs, parameterized queries)
│   ├── Input Validation (Zod, Pydantic)
│   └── Security Headers (Helmet.js 7.x)
├── Data Security Layer
│   ├── Encryption at Rest (AES-256-GCM)
│   ├── Encryption in Transit (TLS 1.3)
│   ├── Secret Management (Vault, AWS Secrets)
│   └── Data Loss Prevention (DLP)
├── Infrastructure Security Layer
│   ├── Container Scanning (Trivy 0.58.x)
│   ├── Dependency Scanning (Snyk 1.x)
│   ├── SAST (Semgrep 1.x)
│   └── DAST (OWASP ZAP 2.15.x)
└── Compliance & Governance Layer
    ├── ISO 27001:2022 (current standard)
    ├── OWASP Top 10 2021 (current version)
    ├── OWASP ASVS 4.0.3 (stable version)
    ├── SOC 2 Type II
    ├── GDPR (EU compliance)
    └── PCI DSS 4.0 (payment security)
```

## 🔐 Pattern 1: OAuth 2.1 + PKCE Authorization with NextAuth.js 5.x

### **High Freedom: Architecture Decision**

**Use OAuth 2.1 with PKCE** for modern, secure authorization flows:
- Eliminates implicit flow vulnerabilities
- PKCE protects against authorization code interception
- Refresh token rotation for enhanced security
- Supports SPA, mobile apps, and server-side apps

**Technology Stack** (Stable Versions):
- NextAuth.js 5.x (formerly Auth.js)
- OAuth 2.1 (RFC standard)
- PKCE (Proof Key for Code Exchange)

### **Medium Freedom: Implementation Pattern**

```typescript
// NextAuth.js 5.x OAuth 2.1 + PKCE Configuration
import NextAuth, { type NextAuthConfig } from "next-auth"
import Google from "next-auth/providers/google"
import GitHub from "next-auth/providers/github"

export const authConfig: NextAuthConfig = {
  providers: [
    Google({
      // OAuth 2.1 with PKCE - Google requires "offline" access_type for refresh tokens
      authorization: {
        params: {
          access_type: "offline",
          prompt: "consent",
          response_type: "code", // Authorization Code Flow
        },
      },
    }),
    GitHub({
      authorization: {
        params: {
          scope: "read:user user:email",
        },
      },
    }),
  ],
  session: {
    strategy: "jwt", // Stateless JWT sessions
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },
  callbacks: {
    // JWT callback: Handle access token refresh
    async jwt({ token, account, user, trigger }) {
      // Initial sign-in: Store tokens
      if (account && user) {
        return {
          ...token,
          access_token: account.access_token,
          refresh_token: account.refresh_token,
          expires_at: account.expires_at,
          user_id: user.id,
        }
      }

      // Token still valid
      if (token.expires_at && Date.now() < token.expires_at * 1000) {
        return token
      }

      // Refresh expired token
      if (!token.refresh_token) {
        throw new Error("Missing refresh token")
      }

      try {
        const response = await fetch("https://oauth2.googleapis.com/token", {
          method: "POST",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: new URLSearchParams({
            client_id: process.env.GOOGLE_CLIENT_ID!,
            client_secret: process.env.GOOGLE_CLIENT_SECRET!,
            grant_type: "refresh_token",
            refresh_token: token.refresh_token!,
          }),
        })

        const tokens = await response.json()

        if (!response.ok) throw tokens

        return {
          ...token,
          access_token: tokens.access_token,
          expires_at: Math.floor(Date.now() / 1000 + tokens.expires_in),
          // Preserve refresh token if not rotated
          refresh_token: tokens.refresh_token ?? token.refresh_token,
        }
      } catch (error) {
        console.error("Error refreshing access token", error)
        return { ...token, error: "RefreshTokenError" }
      }
    },

    // Session callback: Expose necessary data to client
    async session({ session, token }) {
      return {
        ...session,
        user: {
          ...session.user,
          id: token.user_id,
        },
        error: token.error,
      }
    },
  },
  pages: {
    signIn: "/auth/signin",
    error: "/auth/error",
  },
}

export const { handlers, auth, signIn, signOut } = NextAuth(authConfig)
```

### **Low Freedom: Security Configuration**

```typescript
// Environment variables (.env.local)
// CRITICAL: Never commit to version control
AUTH_SECRET="generate-with-openssl-rand-base64-32"
AUTH_URL="https://yourdomain.com"

// OAuth Provider Credentials
GOOGLE_CLIENT_ID="your-google-client-id"
GOOGLE_CLIENT_SECRET="your-google-client-secret"
GITHUB_CLIENT_ID="your-github-client-id"
GITHUB_CLIENT_SECRET="your-github-client-secret"

// TypeScript type extensions
declare module "next-auth" {
  interface Session {
    user: {
      id: string
    }
    error?: "RefreshTokenError"
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    access_token: string
    refresh_token: string
    expires_at: number
    user_id: string
    error?: "RefreshTokenError"
  }
}
```

## 🔒 Pattern 2: Argon2id Password Hashing (OWASP Recommended)

### **High Freedom: Password Security Strategy**

**Use Argon2id** for password hashing - OWASP recommended algorithm:
- Winner of Password Hashing Competition (2015)
- Resistant to GPU cracking attacks
- Resistant to side-channel attacks
- Configurable memory-hard function

**OWASP Recommended Parameters** (2025):
- Memory: 19 MiB minimum (m=19456 KiB)
- Iterations: 2 minimum (t=2)
- Parallelism: 1 (p=1) - prevents DoS in web contexts

### **Medium Freedom: Implementation Pattern**

```typescript
// Node.js - Argon2id with OWASP parameters
import argon2 from "argon2"

export async function hashPassword(password: string): Promise<string> {
  return argon2.hash(password, {
    type: argon2.argon2id, // Argon2id variant
    memoryCost: 19456, // 19 MiB memory (OWASP minimum)
    timeCost: 2, // 2 iterations (OWASP minimum)
    parallelism: 1, // 1 thread (OWASP recommended for web)
  })
}

export async function verifyPassword(
  hash: string,
  password: string
): Promise<boolean> {
  try {
    return await argon2.verify(hash, password)
  } catch (error) {
    console.error("Password verification error:", error)
    return false
  }
}

// Usage example
const userPassword = "user-input-password"
const hashedPassword = await hashPassword(userPassword)
// Store hashedPassword in database

// During login
const isValid = await verifyPassword(storedHash, userInputPassword)
```

```python
# Python - Argon2id with cryptography library 43.x + argon2-cffi
from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError

# Initialize with OWASP parameters
ph = PasswordHasher(
    time_cost=2,        # 2 iterations (OWASP minimum)
    memory_cost=19456,  # 19 MiB memory (OWASP minimum)
    parallelism=1,      # 1 thread (OWASP recommended)
    hash_len=32,        # 32-byte hash output
    salt_len=16,        # 16-byte random salt
)

def hash_password(password: str) -> str:
    """Hash password with Argon2id (OWASP parameters)"""
    return ph.hash(password)

def verify_password(password_hash: str, password: str) -> bool:
    """Verify password against Argon2id hash"""
    try:
        ph.verify(password_hash, password)
        
        # Check if hash needs rehashing (parameters upgraded)
        if ph.check_needs_rehash(password_hash):
            return True  # Signal to rehash with new parameters
        
        return True
    except VerifyMismatchError:
        return False
    except Exception as e:
        print(f"Password verification error: {e}")
        return False

# Usage
user_password = "user-input-password"
hashed = hash_password(user_password)
# Store hashed in database

# During login
is_valid = verify_password(stored_hash, user_input_password)
```

### **Low Freedom: Security Requirements**

```typescript
// Password policy validation (before hashing)
import { z } from "zod"

export const passwordSchema = z.string()
  .min(12, "Password must be at least 12 characters")
  .regex(/[A-Z]/, "Password must contain uppercase letter")
  .regex(/[a-z]/, "Password must contain lowercase letter")
  .regex(/[0-9]/, "Password must contain number")
  .regex(/[^A-Za-z0-9]/, "Password must contain special character")
  .refine(
    (password) => {
      // Check against common passwords (use zxcvbn or similar)
      return !commonPasswords.includes(password.toLowerCase())
    },
    { message: "Password is too common" }
  )

// CRITICAL: Rate limit password attempts
// - 5 attempts per 15 minutes per IP
// - 10 attempts per hour per username
// - Exponential backoff after failures
// - CAPTCHA after 3 failures
```

## 🛡️ Pattern 3: CSRF Protection (Double-Submit Cookie)

### **High Freedom: CSRF Strategy**

**Use double-submit cookie pattern** for stateless CSRF protection:
- No server-side session storage required
- Works with JWT authentication
- Cryptographically secure random tokens

**When to apply**:
- All state-changing operations (POST, PUT, PATCH, DELETE)
- Cookie-based authentication systems
- APIs consumed by same-origin web apps

### **Medium Freedom: Implementation Pattern**

```typescript
// Next.js API Route with CSRF Protection
import { NextRequest, NextResponse } from "next/server"
import crypto from "crypto"

// Middleware: Generate and validate CSRF tokens
export function csrfMiddleware(handler: Function) {
  return async (req: NextRequest) => {
    const method = req.method

    // Generate CSRF token for safe methods
    if (method === "GET" || method === "HEAD") {
      const csrfToken = crypto.randomBytes(32).toString("hex")
      
      const response = await handler(req)
      
      // Set CSRF token in cookie
      response.cookies.set("csrf_token", csrfToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "strict",
        maxAge: 60 * 60, // 1 hour
      })
      
      return response
    }

    // Validate CSRF token for state-changing methods
    if (["POST", "PUT", "PATCH", "DELETE"].includes(method)) {
      const cookieToken = req.cookies.get("csrf_token")?.value
      const headerToken = req.headers.get("x-csrf-token")

      if (!cookieToken || !headerToken || cookieToken !== headerToken) {
        return NextResponse.json(
          { error: "CSRF token validation failed" },
          { status: 403 }
        )
      }
    }

    return handler(req)
  }
}

// Usage in API route
export const POST = csrfMiddleware(async (req: NextRequest) => {
  // Protected endpoint logic
  const data = await req.json()
  // Process request...
  return NextResponse.json({ success: true })
})
```

```typescript
// Client-side: Send CSRF token with requests
async function makeApiRequest(url: string, data: any) {
  // Get CSRF token from cookie (accessible via meta tag or API)
  const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content')

  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrfToken || "", // Send token in header
    },
    credentials: "include", // Include cookies
    body: JSON.stringify(data),
  })

  return response.json()
}
```

### **Low Freedom: Configuration Standards**

```typescript
// CSRF token requirements
const CSRF_CONFIG = {
  tokenLength: 32, // 32 bytes (256 bits)
  tokenExpiry: 3600, // 1 hour
  cookieName: "csrf_token",
  headerName: "x-csrf-token",
  cookieOptions: {
    httpOnly: true, // Prevent JavaScript access
    secure: true, // HTTPS only
    sameSite: "strict" as const, // Strict same-site policy
  },
}

// CRITICAL: Always use secure random generation
// - crypto.randomBytes() in Node.js
// - window.crypto.getRandomValues() in browsers
// - Never use Math.random() for security tokens
```

## 🚨 Pattern 4: XSS Prevention (Content Security Policy)

### **High Freedom: XSS Defense Strategy**

**Implement layered XSS protection**:
1. Content Security Policy (CSP) headers
2. Input sanitization (DOMPurify, validator.js)
3. Output encoding (framework defaults)
4. HttpOnly cookies (prevent XSS cookie theft)

**CSP Strategy**:
- Start strict, relax as needed
- Use nonces for inline scripts
- Report violations to monitoring service

### **Medium Freedom: Implementation Pattern**

```typescript
// Helmet.js 7.x - Security Headers Configuration
import helmet from "helmet"
import express from "express"

const app = express()

// Configure Helmet with strict CSP
app.use(
  helmet({
    contentSecurityPolicy: {
      directives: {
        defaultSrc: ["'self'"], // Default: same-origin only
        scriptSrc: [
          "'self'",
          "'nonce-{NONCE}'", // Replace with dynamic nonce
          "https://trusted-cdn.com",
        ],
        styleSrc: ["'self'", "'nonce-{NONCE}'", "https://fonts.googleapis.com"],
        imgSrc: ["'self'", "data:", "https:"],
        fontSrc: ["'self'", "https://fonts.gstatic.com"],
        connectSrc: ["'self'", "https://api.yourdomain.com"],
        frameSrc: ["'none'"], // Prevent clickjacking
        objectSrc: ["'none'"], // Disable plugins
        upgradeInsecureRequests: [], // Force HTTPS
      },
      reportOnly: false, // Set to true during testing
    },
    hsts: {
      maxAge: 31536000, // 1 year
      includeSubDomains: true,
      preload: true,
    },
    frameguard: {
      action: "deny", // Prevent clickjacking
    },
    referrerPolicy: {
      policy: "strict-origin-when-cross-origin",
    },
    crossOriginEmbedderPolicy: true,
    crossOriginOpenerPolicy: { policy: "same-origin" },
    crossOriginResourcePolicy: { policy: "same-origin" },
  })
)

// Generate CSP nonce per request
app.use((req, res, next) => {
  res.locals.cspNonce = crypto.randomBytes(16).toString("base64")
  next()
})
```

```typescript
// Next.js - CSP with nonces (next.config.js)
const nextConfig = {
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          {
            key: "Content-Security-Policy",
            value: `
              default-src 'self';
              script-src 'self' 'nonce-{NONCE}';
              style-src 'self' 'nonce-{NONCE}' https://fonts.googleapis.com;
              img-src 'self' data: https:;
              font-src 'self' https://fonts.gstatic.com;
              connect-src 'self' https://api.yourdomain.com;
              frame-ancestors 'none';
              base-uri 'self';
              form-action 'self';
            `.replace(/\s{2,}/g, " ").trim(),
          },
          {
            key: "X-Content-Type-Options",
            value: "nosniff",
          },
          {
            key: "X-Frame-Options",
            value: "DENY",
          },
          {
            key: "Referrer-Policy",
            value: "strict-origin-when-cross-origin",
          },
        ],
      },
    ]
  },
}
```

### **Low Freedom: Input Sanitization**

```typescript
// Sanitize user input (DOMPurify for HTML, validator for strings)
import DOMPurify from "isomorphic-dompurify"
import validator from "validator"

export function sanitizeHtml(dirty: string): string {
  return DOMPurify.sanitize(dirty, {
    ALLOWED_TAGS: ["b", "i", "em", "strong", "a", "p"],
    ALLOWED_ATTR: ["href"],
    ALLOW_DATA_ATTR: false,
  })
}

export function sanitizeString(input: string): string {
  return validator.escape(input) // HTML entity encoding
}

// CRITICAL: Always sanitize before database storage
// - Sanitize on input (defense in depth)
// - Escape on output (framework defaults)
// - Never trust user input
```

## 🔐 Pattern 5: SQL Injection Prevention

### **High Freedom: SQL Security Strategy**

**Use ORMs and parameterized queries** - never construct SQL from strings:
- Prisma (Node.js)
- Drizzle ORM (Node.js)
- SQLAlchemy (Python)
- TypeORM (Node.js)

**Defense Layers**:
1. Parameterized queries (primary defense)
2. Input validation (schema validation)
3. Least privilege database accounts
4. Web Application Firewall (WAF)

### **Medium Freedom: Implementation Pattern**

```typescript
// Prisma ORM - Safe by default (parameterized)
import { PrismaClient } from "@prisma/client"

const prisma = new PrismaClient()

// SAFE: Parameterized query
async function getUserByEmail(email: string) {
  return prisma.user.findUnique({
    where: { email }, // Prisma parameterizes automatically
  })
}

// SAFE: Complex query with filtering
async function searchUsers(searchTerm: string, role: string) {
  return prisma.user.findMany({
    where: {
      AND: [
        { role }, // Parameterized
        {
          OR: [
            { name: { contains: searchTerm } }, // Parameterized
            { email: { contains: searchTerm } },
          ],
        },
      ],
    },
  })
}

// DANGEROUS: Raw SQL (only if absolutely necessary)
async function rawQueryExample(userId: string) {
  // CORRECT: Use parameterized raw queries
  return prisma.$queryRaw`
    SELECT * FROM users WHERE id = ${userId}
  ` // Template literal provides parameterization

  // NEVER DO THIS:
  // return prisma.$queryRawUnsafe(`SELECT * FROM users WHERE id = '${userId}'`)
  // ❌ SQL injection vulnerability
}
```

```python
# SQLAlchemy (Python) - Safe parameterized queries
from sqlalchemy import select, and_, or_
from sqlalchemy.orm import Session
from models import User

def get_user_by_email(db: Session, email: str):
    """Safe: SQLAlchemy parameterizes automatically"""
    return db.query(User).filter(User.email == email).first()

def search_users(db: Session, search_term: str, role: str):
    """Safe: Complex query with parameterization"""
    stmt = select(User).where(
        and_(
            User.role == role,  # Parameterized
            or_(
                User.name.contains(search_term),  # Parameterized
                User.email.contains(search_term)
            )
        )
    )
    return db.execute(stmt).scalars().all()

# DANGEROUS: Raw SQL execution
def raw_query_example(db: Session, user_id: str):
    # CORRECT: Use parameterized raw SQL
    result = db.execute(
        "SELECT * FROM users WHERE id = :user_id",
        {"user_id": user_id}  # Named parameter (safe)
    )
    
    # NEVER DO THIS:
    # result = db.execute(f"SELECT * FROM users WHERE id = '{user_id}'")
    # ❌ SQL injection vulnerability
```

### **Low Freedom: Database Security Configuration**

```sql
-- Least privilege principle: Create restricted database user
CREATE USER 'app_user'@'localhost' IDENTIFIED BY 'strong_password';

-- Grant minimum necessary privileges
GRANT SELECT, INSERT, UPDATE, DELETE ON app_database.users TO 'app_user'@'localhost';
GRANT SELECT, INSERT, UPDATE, DELETE ON app_database.posts TO 'app_user'@'localhost';

-- Deny dangerous privileges
-- NO GRANT CREATE, DROP, ALTER, GRANT OPTION

-- Stored procedures for sensitive operations (optional defense layer)
DELIMITER //
CREATE PROCEDURE GetUserById(IN userId INT)
BEGIN
    SELECT * FROM users WHERE id = userId;
END //
DELIMITER ;

GRANT EXECUTE ON PROCEDURE app_database.GetUserById TO 'app_user'@'localhost';
```

## ⚡ Pattern 6: API Rate Limiting & Throttling

### **High Freedom: Rate Limiting Strategy**

**Implement multi-layer rate limiting**:
- IP-based rate limiting (prevent brute force)
- User-based rate limiting (prevent abuse)
- Endpoint-specific limits (sensitive operations)
- Sliding window algorithm (smoother limits)

**Technologies** (Stable):
- express-rate-limit 7.x (Node.js)
- Redis for distributed rate limiting
- Nginx rate limiting (infrastructure layer)

### **Medium Freedom: Implementation Pattern**

```typescript
// Express Rate Limiting with Redis (distributed)
import rateLimit from "express-rate-limit"
import RedisStore from "rate-limit-redis"
import Redis from "ioredis"

const redisClient = new Redis({
  host: process.env.REDIS_HOST,
  port: 6379,
})

// General API rate limit
export const apiLimiter = rateLimit({
  store: new RedisStore({
    client: redisClient,
    prefix: "rl:api:",
  }),
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // 100 requests per window
  message: "Too many requests, please try again later",
  standardHeaders: true, // Return rate limit info in headers
  legacyHeaders: false,
})

// Strict limit for authentication endpoints
export const authLimiter = rateLimit({
  store: new RedisStore({
    client: redisClient,
    prefix: "rl:auth:",
  }),
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 5, // 5 attempts per window
  message: "Too many login attempts, please try again later",
  skipSuccessfulRequests: true, // Don't count successful logins
})

// Usage in Express app
import express from "express"

const app = express()

// Apply global rate limit
app.use("/api/", apiLimiter)

// Apply strict limit to auth routes
app.use("/api/auth/login", authLimiter)
app.use("/api/auth/register", authLimiter)
```

```typescript
// Next.js API Route with Rate Limiting
import { NextRequest, NextResponse } from "next/server"
import { Ratelimit } from "@upstash/ratelimit"
import { Redis } from "@upstash/redis"

const redis = Redis.fromEnv()

// Sliding window rate limiter
const ratelimit = new Ratelimit({
  redis,
  limiter: Ratelimit.slidingWindow(10, "10 s"), // 10 requests per 10 seconds
  analytics: true,
})

export async function POST(req: NextRequest) {
  const ip = req.ip ?? "127.0.0.1"
  
  const { success, limit, reset, remaining } = await ratelimit.limit(
    `api_${ip}`
  )

  if (!success) {
    return NextResponse.json(
      { error: "Rate limit exceeded" },
      {
        status: 429,
        headers: {
          "X-RateLimit-Limit": limit.toString(),
          "X-RateLimit-Remaining": remaining.toString(),
          "X-RateLimit-Reset": reset.toString(),
        },
      }
    )
  }

  // Process request
  return NextResponse.json({ success: true })
}
```

### **Low Freedom: Rate Limit Configuration**

```typescript
// Rate limit tiers by endpoint sensitivity
export const RATE_LIMITS = {
  // Public endpoints (lenient)
  public: {
    windowMs: 15 * 60 * 1000, // 15 minutes
    max: 100, // 100 requests
  },
  // Authenticated endpoints (moderate)
  authenticated: {
    windowMs: 15 * 60 * 1000,
    max: 1000, // 1000 requests
  },
  // Authentication endpoints (strict)
  auth: {
    windowMs: 15 * 60 * 1000,
    max: 5, // 5 attempts
  },
  // Password reset (very strict)
  passwordReset: {
    windowMs: 60 * 60 * 1000, // 1 hour
    max: 3, // 3 attempts
  },
}

// CRITICAL: Monitor rate limit violations
// - Log excessive rate limit hits
// - Alert on sustained violations
// - Consider IP blocking for severe abuse
```

## 🔐 Pattern 7: Secure Session Management

### **High Freedom: Session Strategy**

**Use JWT with RS256 signatures** for stateless, scalable sessions:
- Asymmetric keys (private key signs, public key verifies)
- Short-lived access tokens (15 minutes)
- Long-lived refresh tokens (30 days)
- Secure storage (httpOnly cookies)

**Session Requirements**:
- Session rotation after authentication
- Absolute timeout (30 days)
- Idle timeout (15 minutes)
- Multi-device session management

### **Medium Freedom: Implementation Pattern**

```typescript
// JWT Session Management with RS256
import jwt from "jsonwebtoken"
import fs from "fs"

// Load RSA keys (generate with: openssl genrsa -out private.pem 2048)
const privateKey = fs.readFileSync("private.pem", "utf8")
const publicKey = fs.readFileSync("public.pem", "utf8")

interface TokenPayload {
  userId: string
  email: string
  role: string
}

// Generate access token (short-lived)
export function generateAccessToken(payload: TokenPayload): string {
  return jwt.sign(payload, privateKey, {
    algorithm: "RS256",
    expiresIn: "15m", // 15 minutes
    issuer: "yourdomain.com",
    audience: "yourdomain.com",
  })
}

// Generate refresh token (long-lived)
export function generateRefreshToken(userId: string): string {
  return jwt.sign({ userId }, privateKey, {
    algorithm: "RS256",
    expiresIn: "30d", // 30 days
    issuer: "yourdomain.com",
    audience: "yourdomain.com",
  })
}

// Verify token
export function verifyToken(token: string): TokenPayload | null {
  try {
    return jwt.verify(token, publicKey, {
      algorithms: ["RS256"],
      issuer: "yourdomain.com",
      audience: "yourdomain.com",
    }) as TokenPayload
  } catch (error) {
    console.error("Token verification failed:", error)
    return null
  }
}

// Next.js API: Set secure cookies
export function setAuthCookies(
  res: NextResponse,
  accessToken: string,
  refreshToken: string
) {
  // Access token (httpOnly, secure)
  res.cookies.set("access_token", accessToken, {
    httpOnly: true, // Prevent XSS access
    secure: process.env.NODE_ENV === "production", // HTTPS only
    sameSite: "strict", // CSRF protection
    maxAge: 15 * 60, // 15 minutes
    path: "/",
  })

  // Refresh token (httpOnly, secure, separate path)
  res.cookies.set("refresh_token", refreshToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "strict",
    maxAge: 30 * 24 * 60 * 60, // 30 days
    path: "/api/auth/refresh", // Only sent to refresh endpoint
  })
}
```

### **Low Freedom: Session Security Requirements**

```typescript
// Session configuration standards
const SESSION_CONFIG = {
  accessToken: {
    algorithm: "RS256" as const, // Asymmetric signature
    expiresIn: "15m", // 15 minutes
    keySize: 2048, // RSA key size (bits)
  },
  refreshToken: {
    algorithm: "RS256" as const,
    expiresIn: "30d", // 30 days
    keySize: 2048,
  },
  cookies: {
    httpOnly: true, // MANDATORY: Prevent XSS
    secure: true, // MANDATORY: HTTPS only
    sameSite: "strict" as const, // MANDATORY: CSRF protection
  },
}

// CRITICAL: Session security checklist
// - Rotate refresh tokens after use
// - Invalidate sessions on logout
// - Track active sessions per user
// - Allow user to revoke sessions
// - Log session creation/termination
```

## 🛡️ Pattern 8: Secret Management (HashiCorp Vault)

### **High Freedom: Secret Management Strategy**

**Use dedicated secret management systems**:
- HashiCorp Vault (self-hosted)
- AWS Secrets Manager (AWS)
- Google Secret Manager (GCP)
- Azure Key Vault (Azure)

**Never store secrets in**:
- Git repositories (.env files)
- Application code (hardcoded)
- CI/CD logs (masked secrets)
- Container images (build-time secrets)

### **Medium Freedom: Implementation Pattern**

```typescript
// HashiCorp Vault Client (Node.js)
import vault from "node-vault"

const vaultClient = vault({
  apiVersion: "v1",
  endpoint: process.env.VAULT_ADDR || "http://localhost:8200",
  token: process.env.VAULT_TOKEN, // From secure source (k8s secret, IAM)
})

// Read secret from Vault
export async function getSecret(path: string): Promise<any> {
  try {
    const result = await vaultClient.read(path)
    return result.data
  } catch (error) {
    console.error(`Failed to read secret from ${path}:`, error)
    throw new Error("Secret retrieval failed")
  }
}

// Write secret to Vault
export async function setSecret(path: string, data: any): Promise<void> {
  try {
    await vaultClient.write(path, { data })
  } catch (error) {
    console.error(`Failed to write secret to ${path}:`, error)
    throw new Error("Secret storage failed")
  }
}

// Usage: Load database credentials from Vault
async function getDatabaseConfig() {
  const dbSecret = await getSecret("secret/data/database")
  
  return {
    host: dbSecret.host,
    port: dbSecret.port,
    username: dbSecret.username,
    password: dbSecret.password,
    database: dbSecret.database,
  }
}
```

```python
# HashiCorp Vault Client (Python)
import hvac
import os

vault_client = hvac.Client(
    url=os.getenv("VAULT_ADDR", "http://localhost:8200"),
    token=os.getenv("VAULT_TOKEN")  # From secure source
)

def get_secret(path: str) -> dict:
    """Read secret from Vault"""
    try:
        response = vault_client.secrets.kv.v2.read_secret_version(path=path)
        return response["data"]["data"]
    except Exception as e:
        print(f"Failed to read secret from {path}: {e}")
        raise RuntimeError("Secret retrieval failed")

def set_secret(path: str, data: dict) -> None:
    """Write secret to Vault"""
    try:
        vault_client.secrets.kv.v2.create_or_update_secret(
            path=path,
            secret=data
        )
    except Exception as e:
        print(f"Failed to write secret to {path}: {e}")
        raise RuntimeError("Secret storage failed")

# Usage
db_config = get_secret("database/production")
DATABASE_URL = f"postgresql://{db_config['username']}:{db_config['password']}@{db_config['host']}/{db_config['database']}"
```

### **Low Freedom: Secret Management Standards**

```bash
# Vault initialization (production)
vault operator init -key-shares=5 -key-threshold=3

# Enable KV secrets engine
vault secrets enable -path=secret kv-v2

# Create secret policy
vault policy write app-policy - <<EOF
path "secret/data/app/*" {
  capabilities = ["read"]
}
EOF

# Create token with policy
vault token create -policy=app-policy

# CRITICAL: Secret rotation schedule
# - Database credentials: 90 days
# - API keys: 30 days
# - TLS certificates: before expiry
# - Vault root token: NEVER use in production
```

## 🔍 Pattern 9: Automated Security Testing

### **High Freedom: Security Testing Strategy**

**Implement multi-layer automated testing**:
1. SAST (Static Application Security Testing)
2. DAST (Dynamic Application Security Testing)
3. Dependency scanning (known vulnerabilities)
4. Container scanning (base image vulnerabilities)
5. Infrastructure scanning (IaC security)

**Technologies** (Stable Versions):
- Snyk 1.x (dependency scanning)
- OWASP ZAP 2.15.x (DAST)
- Trivy 0.58.x (container scanning)
- Semgrep 1.x (SAST)

### **Medium Freedom: Implementation Pattern**

```yaml
# GitHub Actions - Comprehensive Security Scanning
name: Security Scan

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 0 * * 0' # Weekly scan

jobs:
  dependency-scan:
    name: Dependency Vulnerability Scan (Snyk)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Snyk to check for vulnerabilities
        uses: snyk/actions/node@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high --fail-on=all

  sast-scan:
    name: Static Code Analysis (Semgrep)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Semgrep
        uses: returntocorp/semgrep-action@v1
        with:
          config: >-
            p/security-audit
            p/owasp-top-ten
            p/javascript
            p/typescript

  container-scan:
    name: Container Image Scan (Trivy)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Build Docker image
        run: docker build -t myapp:latest .
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: myapp:latest
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
      
      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: 'trivy-results.sarif'

  dast-scan:
    name: Dynamic Security Testing (OWASP ZAP)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: ZAP Baseline Scan
        uses: zaproxy/action-baseline@v0.11.0
        with:
          target: 'https://staging.yourdomain.com'
          rules_file_name: '.zap/rules.tsv'
          cmd_options: '-a'
```

### **Low Freedom: Security Testing Standards**

```typescript
// Security testing configuration (.securityrc.json)
{
  "snyk": {
    "severity_threshold": "high",
    "fail_on_issues": true,
    "exclude_dev_dependencies": false
  },
  "trivy": {
    "severity": ["CRITICAL", "HIGH"],
    "vuln_type": ["os", "library"],
    "ignore_unfixed": true
  },
  "zap": {
    "scan_type": "baseline",
    "exclude_urls": ["/api/docs", "/health"],
    "alert_threshold": "WARN"
  },
  "semgrep": {
    "config": ["p/owasp-top-ten", "p/security-audit"],
    "severity": ["ERROR", "WARNING"]
  }
}

// CRITICAL: Security gate requirements
// - No CRITICAL vulnerabilities allowed
// - HIGH vulnerabilities require justification
// - All scans must pass before deployment
// - Monthly full penetration testing
```

## 📊 Pattern 10: OWASP Top 10 2021 Compliance

### **High Freedom: Compliance Strategy**

**Address all OWASP Top 10 2021 categories**:

**A01: Broken Access Control**
- Implement RBAC (Role-Based Access Control)
- Enforce principle of least privilege
- Deny by default

**A02: Cryptographic Failures**
- Use TLS 1.3 for data in transit
- Use AES-256-GCM for data at rest
- Argon2id for password hashing

**A03: Injection**
- Parameterized queries (ORMs)
- Input validation (Zod, Pydantic)
- Output encoding

**A04: Insecure Design**
- Threat modeling (STRIDE, DREAD)
- Secure design patterns
- Security requirements in SDLC

**A05: Security Misconfiguration**
- Disable default accounts
- Remove unnecessary features
- Security headers (Helmet.js)

**A06: Vulnerable and Outdated Components**
- Automated dependency scanning (Snyk)
- Regular updates
- SCA (Software Composition Analysis)

**A07: Identification and Authentication Failures**
- MFA (Multi-Factor Authentication)
- Secure password policies
- Session management

**A08: Software and Data Integrity Failures**
- Code signing
- CI/CD pipeline security
- Dependency verification (lock files)

**A09: Security Logging and Monitoring Failures**
- Centralized logging
- Real-time alerting
- Audit trails

**A10: Server-Side Request Forgery (SSRF)**
- URL allowlisting
- Network segmentation
- Input validation

### **Medium Freedom: OWASP Top 10 Checklist**

```typescript
// OWASP Top 10 2021 Compliance Checklist
export const OWASP_TOP_10_CHECKLIST = {
  A01_BrokenAccessControl: {
    implemented: [
      "RBAC with role hierarchy",
      "Resource-based access control",
      "Deny by default authorization",
      "API endpoint authorization checks",
    ],
    validation: "Manual code review + automated tests",
  },
  A02_CryptographicFailures: {
    implemented: [
      "TLS 1.3 for all connections",
      "AES-256-GCM for data at rest",
      "Argon2id for passwords (OWASP params)",
      "Secure random number generation",
    ],
    validation: "SSL Labs + Trivy scanning",
  },
  A03_Injection: {
    implemented: [
      "Parameterized queries (Prisma ORM)",
      "Input validation (Zod schemas)",
      "Output encoding (framework defaults)",
      "NoSQL injection prevention",
    ],
    validation: "Semgrep SAST + manual review",
  },
  A04_InsecureDesign: {
    implemented: [
      "Threat modeling (STRIDE)",
      "Security requirements documentation",
      "Secure design patterns",
      "Security review in design phase",
    ],
    validation: "Architecture review board",
  },
  A05_SecurityMisconfiguration: {
    implemented: [
      "Helmet.js security headers",
      "Disable default credentials",
      "Remove debug endpoints in production",
      "Minimal attack surface",
    ],
    validation: "OWASP ZAP + manual audit",
  },
  A06_VulnerableComponents: {
    implemented: [
      "Snyk dependency scanning",
      "Automated updates (Dependabot)",
      "SCA in CI/CD pipeline",
      "SBOM generation",
    ],
    validation: "Daily Snyk scans",
  },
  A07_AuthenticationFailures: {
    implemented: [
      "NextAuth.js 5.x with OAuth 2.1",
      "MFA support",
      "Secure session management (JWT RS256)",
      "Rate limiting on auth endpoints",
    ],
    validation: "Penetration testing",
  },
  A08_IntegrityFailures: {
    implemented: [
      "Code signing (GPG)",
      "Dependency lock files",
      "CI/CD pipeline security",
      "Immutable infrastructure",
    ],
    validation: "Pipeline audit",
  },
  A09_LoggingMonitoringFailures: {
    implemented: [
      "Centralized logging (ELK Stack)",
      "Real-time alerting (PagerDuty)",
      "Audit trails (immutable logs)",
      "Security event monitoring",
    ],
    validation: "Log analysis + SIEM integration",
  },
  A10_SSRF: {
    implemented: [
      "URL allowlisting",
      "Network segmentation",
      "Input validation (URL schemes)",
      "Internal network isolation",
    ],
    validation: "Manual testing + ZAP",
  },
}
```

### **Low Freedom: Compliance Validation**

```bash
# OWASP Top 10 Compliance Audit Script
#!/bin/bash

echo "Starting OWASP Top 10 2021 Compliance Audit..."

# A01: Access Control
echo "A01: Checking access control implementation..."
grep -r "authorize" src/ | wc -l

# A02: Cryptography
echo "A02: Checking cryptographic implementations..."
grep -r "argon2" src/ | wc -l
grep -r "AES-256-GCM" src/ | wc -l

# A03: Injection
echo "A03: Checking for raw SQL usage..."
grep -r "queryRawUnsafe" src/ && echo "WARNING: Raw SQL found"

# A05: Security Headers
echo "A05: Verifying security headers..."
curl -I https://yourdomain.com | grep -E "Content-Security-Policy|X-Frame-Options|Strict-Transport-Security"

# A06: Vulnerable Dependencies
echo "A06: Scanning dependencies..."
snyk test --severity-threshold=high

# A09: Logging
echo "A09: Checking logging implementation..."
grep -r "logger" src/ | wc -l

echo "Compliance audit complete."
```

## 📋 ISO 27001:2022 Compliance Summary

### **Key Control Areas**

**Organizational Controls** (37 controls):
- Information security policies
- Roles and responsibilities
- Segregation of duties
- Management oversight

**People Controls** (8 controls):
- Security awareness training
- Background screening
- Employment agreements
- Disciplinary process

**Physical Controls** (14 controls):
- Physical security perimeters
- Access control systems
- Environmental protection
- Equipment security

**Technological Controls** (34 controls):
- Privileged access management
- Secure authentication
- Cryptography
- Vulnerability management
- Secure development lifecycle
- Security testing

### **Implementation Progress Tracking**

```typescript
// ISO 27001:2022 Control Status
export interface ISO27001Control {
  id: string
  title: string
  category: "Organizational" | "People" | "Physical" | "Technological"
  status: "Implemented" | "In Progress" | "Not Started"
  evidence: string[]
}

// Example implemented controls
export const implementedControls: ISO27001Control[] = [
  {
    id: "A.8.2",
    title: "Privileged access rights",
    category: "Technological",
    status: "Implemented",
    evidence: ["RBAC system", "MFA for admins", "Audit logs"],
  },
  {
    id: "A.8.5",
    title: "Secure authentication",
    category: "Technological",
    status: "Implemented",
    evidence: ["NextAuth.js 5.x", "Argon2id hashing", "JWT RS256"],
  },
  {
    id: "A.8.24",
    title: "Use of cryptography",
    category: "Technological",
    status: "Implemented",
    evidence: ["TLS 1.3", "AES-256-GCM", "RSA 2048-bit keys"],
  },
]
```

## 🎯 Security Metrics & KPIs

```typescript
export interface SecurityMetrics {
  vulnerabilityManagement: {
    criticalVulnerabilities: number // Target: 0
    highVulnerabilities: number // Target: < 5
    meanTimeToRemediate: number // Target: < 7 days
    patchCompliance: number // Target: > 95%
  }
  authentication: {
    successRate: number // Target: > 99%
    mfaAdoption: number // Target: > 90%
    passwordStrength: number // Target: > 95% strong
    accountTakeoverAttempts: number // Target: 0
  }
  accessControl: {
    unauthorizedAccessAttempts: number // Target: 0
    privilegeEscalationAttempts: number // Target: 0
    rbacCoverage: number // Target: 100%
  }
  compliance: {
    owaspTop10Score: number // Target: 100%
    iso27001Score: number // Target: > 95%
    failedAudits: number // Target: 0
  }
  incidentResponse: {
    meanTimeToDetect: number // Target: < 1 hour
    meanTimeToRespond: number // Target: < 4 hours
    meanTimeToResolve: number // Target: < 24 hours
  }
}
```

## 📚 Official Documentation & Resources

**Standards & Frameworks**:
- OWASP Top 10 2021: https://owasp.org/www-project-top-ten/
- OWASP ASVS 4.0.3: https://owasp.org/www-project-application-security-verification-standard/
- ISO 27001:2022: https://www.iso.org/standard/27001
- NIST Cybersecurity Framework: https://www.nist.gov/cyberframework
- CIS Controls: https://www.cisecurity.org/controls/

**Technologies** (Stable Versions):
- NextAuth.js 5.x: https://next-auth.js.org/
- Passport.js 0.7.x: http://www.passportjs.org/
- Helmet.js 7.x: https://helmetjs.github.io/
- Argon2 OWASP Guide: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- Snyk 1.x: https://snyk.io/
- OWASP ZAP 2.15.x: https://www.zaproxy.org/
- Trivy 0.58.x: https://aquasecurity.github.io/trivy/

**Compliance Resources**:
- GDPR Compliance: https://gdpr.eu/
- SOC 2 Guide: https://www.aicpa.org/
- PCI DSS 4.0: https://www.pcisecuritystandards.org/

## 🤝 Works Seamlessly With

- **moai-domain-backend**: Backend security architecture, API security, authentication
- **moai-domain-frontend**: Client-side security, XSS prevention, CSP configuration
- **moai-domain-database**: Database security, encryption at rest, access control
- **moai-domain-devops**: Security automation, vulnerability scanning, compliance testing
- **moai-domain-api**: API gateway security, rate limiting, authentication middleware
- **moai-domain-infrastructure**: Infrastructure security, network policies, secret management

---

**Version**: 4.0.0 Enterprise  
**Last Updated**: 2025-11-12  
**Enterprise Ready**: ✅ Production-Grade with Stable Technologies  
**OWASP Compliance**: ✅ Top 10 2021 Full Coverage  
**ISO Compliance**: ✅ ISO 27001:2022 Ready  
**Stable Versions**: ✅ All technologies stable (November 2025)
