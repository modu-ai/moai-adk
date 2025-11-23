# Insecure Design (A04)

## Overview

Insecure design refers to missing or ineffective security controls in the design phase, including missing threat modeling and secure design patterns.

## Critical Design Flaws

### Missing Threat Modeling
- No security requirements defined
- Missing abuse case analysis
- Unidentified attack surfaces
- No security reviews

### Design Weaknesses
- Missing rate limiting
- No input validation strategy
- Weak authentication design
- Missing audit logging

## Remediation Patterns

### Threat Modeling

**STRIDE Methodology**:
```
S - Spoofing: Can attacker impersonate another user?
T - Tampering: Can attacker modify data?
R - Repudiation: Can attacker deny actions?
I - Information Disclosure: Can attacker access sensitive data?
D - Denial of Service: Can attacker disrupt service?
E - Elevation of Privilege: Can attacker gain unauthorized access?
```

**Example Threat Model**:
```markdown
## Feature: Password Reset

### Threat 1: Email Enumeration (Information Disclosure)
**Attack**: Attacker tests emails to find valid accounts
**Mitigation**: Generic response for all emails
**Status**: Implemented

### Threat 2: Rate Limit Bypass (Denial of Service)
**Attack**: Attacker floods reset endpoint
**Mitigation**: Rate limiting per IP and per email
**Status**: Implemented

### Threat 3: Token Prediction (Spoofing)
**Attack**: Attacker guesses reset tokens
**Mitigation**: Cryptographically random 256-bit tokens
**Status**: Implemented
```

### Rate Limiting

**Express Rate Limit**:
```javascript
const rateLimit = require('express-rate-limit');

// Global rate limit
const globalLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // 100 requests per window
  message: 'Too many requests, please try again later'
});

// Strict rate limit for authentication
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5, // 5 attempts per 15 minutes
  skipSuccessfulRequests: true
});

app.use('/api/', globalLimiter);
app.use('/api/auth/', authLimiter);
```

**Redis-based Rate Limiting**:
```javascript
class RateLimiter {
  constructor(redis, options) {
    this.redis = redis;
    this.maxRequests = options.maxRequests || 100;
    this.windowMs = options.windowMs || 60000;
  }

  async isAllowed(key) {
    const current = await this.redis.incr(key);

    if (current === 1) {
      await this.redis.expire(key, Math.ceil(this.windowMs / 1000));
    }

    return current <= this.maxRequests;
  }
}
```

### Secure Password Reset

**Comprehensive Implementation**:
```javascript
const crypto = require('crypto');

app.post('/reset-password', authLimiter, async (req, res) => {
  const { email } = req.body;

  // Validate email format
  if (!isValidEmail(email)) {
    return res.status(400).json({ error: 'Invalid email' });
  }

  const user = await db.users.findByEmail(email);

  // Always return same response (prevent enumeration)
  const response = {
    message: 'If account exists, reset email was sent'
  };

  if (user) {
    // Generate cryptographically random token
    const token = crypto.randomBytes(32).toString('hex');

    // Store token with expiration
    await db.resetTokens.create({
      userId: user.id,
      token: await bcrypt.hash(token, 10),
      expiresAt: Date.now() + 3600000 // 1 hour
    });

    // Send email (non-blocking)
    sendResetEmail(user.email, token).catch(console.error);

    // Log event
    logger.info('Password reset requested', {
      userId: user.id,
      ip: req.ip
    });
  }

  res.json(response);
});
```

### Business Logic Validation

**Order Processing Security**:
```javascript
class OrderValidator {
  async validateOrder(order, user) {
    // Check order ownership
    if (order.userId !== user.id) {
      throw new Error('Unauthorized');
    }

    // Verify prices (server-side recalculation)
    const calculatedTotal = await this.calculateTotal(order.items);
    if (Math.abs(order.total - calculatedTotal) > 0.01) {
      throw new Error('Price mismatch');
    }

    // Check inventory
    for (const item of order.items) {
      const available = await this.checkInventory(item.productId);
      if (available < item.quantity) {
        throw new Error(`Insufficient inventory: ${item.productId}`);
      }
    }

    // Verify discount codes
    if (order.discountCode) {
      const valid = await this.validateDiscount(
        order.discountCode,
        user.id,
        order.total
      );
      if (!valid) {
        throw new Error('Invalid discount code');
      }
    }

    return true;
  }

  async calculateTotal(items) {
    let total = 0;
    for (const item of items) {
      const product = await db.products.findById(item.productId);
      total += product.price * item.quantity;
    }
    return total;
  }
}
```

### Secure State Machine

**Account State Transitions**:
```javascript
class AccountStateMachine {
  constructor() {
    this.validTransitions = {
      'pending': ['active', 'suspended'],
      'active': ['suspended', 'closed'],
      'suspended': ['active', 'closed'],
      'closed': [] // Terminal state
    };
  }

  canTransition(from, to) {
    return this.validTransitions[from]?.includes(to) || false;
  }

  async transitionState(accountId, toState, reason) {
    const account = await db.accounts.findById(accountId);

    if (!this.canTransition(account.state, toState)) {
      throw new Error(
        `Invalid state transition: ${account.state} -> ${toState}`
      );
    }

    // Atomic update with audit trail
    await db.transaction(async (tx) => {
      await tx.accounts.update(accountId, { state: toState });
      await tx.auditLog.create({
        accountId,
        fromState: account.state,
        toState,
        reason,
        timestamp: Date.now()
      });
    });

    return true;
  }
}
```

## Design Principles

### Defense in Depth
```
Layer 1: Network (Firewall, WAF)
Layer 2: Application (Input validation, authentication)
Layer 3: Database (Parameterized queries, least privilege)
Layer 4: Monitoring (Logging, alerting)
```

### Principle of Least Privilege
```javascript
// Role-based permissions
const permissions = {
  'user': ['read:own', 'update:own'],
  'moderator': ['read:all', 'update:any', 'delete:flagged'],
  'admin': ['read:all', 'update:all', 'delete:all']
};

function hasPermission(user, action, resource) {
  const userPerms = permissions[user.role] || [];
  return userPerms.includes(action);
}
```

### Fail Securely
```javascript
// Always default to deny
function checkAccess(user, resource) {
  try {
    if (!user) return false;
    if (!resource) return false;

    // Explicit checks
    if (user.id === resource.ownerId) return true;
    if (user.role === 'admin') return true;

    // Default deny
    return false;
  } catch (error) {
    // Fail securely on error
    logger.error('Access check failed', { user, resource, error });
    return false;
  }
}
```

## Validation Checklist

- [ ] Threat model created
- [ ] Security requirements defined
- [ ] Rate limiting implemented
- [ ] Business logic validated
- [ ] State transitions secured
- [ ] Defense in depth applied
- [ ] Least privilege enforced
- [ ] Fail securely principle applied

---

**Last Updated**: 2025-11-24
**OWASP Category**: A04:2021
**CWE**: CWE-840 (Business Logic Errors)
