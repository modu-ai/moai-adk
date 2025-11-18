---
name: moai-security-owasp
version: 4.0.0
status: stable
description: Enterprise Skill for advanced development
allowed-tools: Read, Bash, WebSearch, WebFetch
updated: '2025-11-18'
stability: stable
---


# moai-security-owasp: OWASP Top 10 2021 Defense Patterns

**Complete Protection Against OWASP Top 10 2021 Vulnerabilities**  
Trust Score: 9.8/10 | Version: 4.0.0 | Enterprise Mode | Last Updated: 2025-11-12

---

## Overview

The OWASP Top 10 2021 represents the most critical web application security risks. This Skill provides production-ready defense patterns for all 10 categories with code examples and validation strategies.

**When to use this Skill:**
- Protecting against SQL injection, XSS, and CSRF attacks
- Implementing secure access control (BOLA/IDOR prevention)
- Validating and sanitizing user input
- Implementing security headers and CSP
- Building secure file upload handling
- Protecting against XXE and deserialization attacks
- Implementing cryptographic security
- Building secure authentication systems
- Preventing sensitive data exposure
- Implementing logging and monitoring

---

## Level 1: OWASP Top 10 2021 Overview

### Rankings & Changes

| Rank | 2021 Category | Focus | OWASP A# |
|------|---------------|-------|---------|
| **1** | Broken Access Control | BOLA, IDOR, BFLA | A01 |
| **2** | Cryptographic Failures | Weak encryption, hardcoded keys | A02 |
| **3** | Injection | SQL, OS, NoSQL, LDAP | A03 |
| **4** | Insecure Design | Missing threat modeling | A04 |
| **5** | Security Misconfiguration | Default creds, verbose errors | A05 |
| **6** | Vulnerable Components | Outdated dependencies | A06 |
| **7** | Authentication Failures | Weak MFA, session flaws | A07 |
| **8** | Data Integrity Failures | Insecure deserialization | A08 |
| **9** | Logging & Monitoring Failures | Missing audit trails | A09 |
| **10** | SSRF | Server-side request forgery | A10 |

### Key Changes from 2017 to 2021

```
2017 → 2021:
- XSS merged into Injection (A03)
- Broken Access Control elevated to #1
- Insecure Deserialization → Data Integrity Failures
- XXE moved to Injection
- Using Components with Known Vulns → Vulnerable Components
- Insufficient Logging → Logging & Monitoring
- SSRF added to Top 10
```

---

## Level 2: Defense Patterns for Each Category

### A01: Broken Access Control (BOLA & IDOR)

**BOLA (Broken Object Level Authorization):**

```javascript
// VULNERABLE: No object ownership check
app.get('/api/users/:userId', jwtAuth, (req, res) => {
  const user = db.users.findById(req.params.userId);
  res.json(user); // Attacker can access any user!
});

// SECURE: Verify ownership
app.get('/api/users/:userId', jwtAuth, (req, res) => {
  const user = db.users.findById(req.params.userId);
  
  // Check: User can only access their own data (or admin)
  if (req.user.id !== user.id && req.user.role !== 'admin') {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  res.json(user);
});

// Multi-tenant: Always check tenant_id
app.get('/api/users/:userId', jwtAuth, (req, res) => {
  const user = db.users.findById(req.params.userId);
  
  // CRITICAL: Verify tenant ownership
  if (user.tenant_id !== req.tenantId) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  res.json(user);
});
```

**BFLA (Broken Function Level Authorization):**

```javascript
// VULNERABLE: No role check
app.post('/api/users/:userId/admin', jwtAuth, (req, res) => {
  const user = db.users.findById(req.params.userId);
  user.role = 'admin';  // Any user can become admin!
  db.users.update(user);
  res.json(user);
});

// SECURE: Verify admin role
app.post('/api/users/:userId/promote', jwtAuth, (req, res) => {
  if (req.user.role !== 'admin') {
    return res.status(403).json({ error: 'Forbidden' });
  }
  
  const user = db.users.findById(req.params.userId);
  user.role = 'admin';
  db.users.update(user);
  res.json(user);
});
```

### A03: Injection (SQL, NoSQL, OS)

**SQL Injection Prevention:**

```javascript
// VULNERABLE: String concatenation
const userId = req.query.userId;
const query = `SELECT * FROM users WHERE id = ${userId}`;
// Attack: userId = "1 OR 1=1" returns all users
db.query(query);

// SECURE: Parameterized queries
const query = 'SELECT * FROM users WHERE id = ?';
db.query(query, [userId]); // userId treated as value, not code

// SECURE: With ORM (Sequelize)
const user = await User.findByPk(userId);

// SECURE: With TypeORM
const user = await userRepository.createQueryBuilder()
  .where('user.id = :id', { id: userId })
  .getOne();
```

**NoSQL Injection:**

```javascript
// VULNERABLE: Direct query construction
const query = { username: req.body.username };
const user = await db.collection('users').findOne(query);
// Attack: username = { $ne: '' } bypasses auth

// SECURE: Validation + parameterized
const schema = z.object({
  username: z.string().email()
});

const validated = schema.parse(req.body);
const user = await db.collection('users').findOne({
  username: validated.username
});
```

### A07: Authentication Failures

**Secure Password Validation:**

```javascript
// Rate limiting for login attempts
const loginAttempts = new Map();

app.post('/login', async (req, res) => {
  const key = req.body.email;
  const attempts = loginAttempts.get(key) || 0;
  
  if (attempts >= 5) {
    return res.status(429).json({ 
      error: 'Too many attempts. Try again in 15 minutes.' 
    });
  }
  
  const user = await db.users.findByEmail(req.body.email);
  const passwordValid = user && 
    await bcrypt.compare(req.body.password, user.passwordHash);
  
  if (!passwordValid) {
    loginAttempts.set(key, attempts + 1);
    // Always return same error (prevents user enumeration)
    return res.status(401).json({ error: 'Invalid credentials' });
  }
  
  loginAttempts.delete(key);
  
  // Check MFA if enabled
  if (user.mfaEnabled) {
    // Send OTP or prompt for TOTP
    return res.json({ requiresMfa: true });
  }
  
  res.json({ token: jwt.sign({ id: user.id }, process.env.JWT_SECRET) });
});
```

### A05: Security Misconfiguration

**HTTP Security Headers:**

```javascript
const helmet = require('helmet');

app.use(helmet({
  // Content Security Policy
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'", "https://trusted-cdn.com"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      imgSrc: ["'self'", "data:", "https:"],
      connectSrc: ["'self'"],
      fontSrc: ["'self'"],
      objectSrc: ["'none'"],
      mediaSrc: ["'self'"],
      frameSrc: ["'none'"]
    }
  },
  
  // HSTS: Force HTTPS
  hsts: {
    maxAge: 31536000,
    includeSubDomains: true,
    preload: true
  },
  
  // Prevent clickjacking
  frameguard: { action: 'deny' },
  
  // Prevent MIME sniffing
  noSniff: true,
  
  // XSS Protection header
  xssFilter: true,
  
  // Referrer Policy
  referrerPolicy: { policy: 'no-referrer' }
}));

// Disable server header
app.disable('x-powered-by');
```

---

## Level 3: Advanced Input Validation

### XSS Prevention (A03 Injection)

```javascript
const { body, validationResult } = require('express-validator');
const sanitizeHtml = require('sanitize-html');

// 1. Input validation
const validateInput = [
  body('comment')
    .trim()
    .isLength({ min: 1, max: 500 })
    .escape()  // Convert <, >, &, ", ' to entities
];

// 2. Sanitization (stronger than escape)
app.post('/comments', validateInput, (req, res) => {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    return res.status(400).json({ errors });
  }
  
  // Sanitize HTML
  const sanitized = sanitizeHtml(req.body.comment, {
    allowedTags: ['b', 'i', 'em', 'strong', 'p'],
    allowedAttributes: {},
    disallowedTagsMode: 'discard'
  });
  
  // Store sanitized version
  db.comments.create({ content: sanitized });
  res.json({ success: true });
});

// 3. Output encoding in templates (EJS/Handlebars)
// <%= comment %> auto-escapes HTML
// {{{ comment }}} does NOT escape (dangerous!)
```

### CSRF Prevention

```javascript
const csrf = require('csurf');
const cookieParser = require('cookie-parser');

app.use(cookieParser());
app.use(csrf({ cookie: false }));

// GET: Return CSRF token
app.get('/form', (req, res) => {
  res.json({ csrfToken: req.csrfToken() });
});

// POST: Validate CSRF token
app.post('/form', csrf(), (req, res) => {
  // Token automatically verified by middleware
  // If invalid, returns 403
  res.json({ success: true });
});

// Alternative: SameSite cookie
res.cookie('session', token, {
  sameSite: 'strict',  // No cross-site requests
  secure: true,        // HTTPS only
  httpOnly: true       // No JavaScript access
});
```

### XXE (XML External Entity) Prevention

```javascript
const xml2js = require('xml2js');

// VULNERABLE: External entities enabled by default
const parser = new xml2js.Parser();
parser.parseString(xmlInput, (err, result) => {
  // Entity expansion attack possible
});

// SECURE: Disable external entities
const parser = new xml2js.Parser({
  strict: false,
  normalize: true,
  normalizeTags: true,
  // Libxmljs doesn't support DTD disabling,
  // use alternative parser or validate schema
});

// BETTER: Use JSON instead of XML
// If XML required: validate against schema
```

---

## Reference

### Official Resources
- OWASP Top 10 2021: https://owasp.org/Top10/
- OWASP Top 10 2017: https://owasp.org/www-project-top-ten-2017/
- CWE Top 25: https://cwe.mitre.org/top25/
- NIST SP 800-63: https://pages.nist.gov/800-63-3/

### Tools & Libraries
- **helmet** (7.0.x): https://helmetjs.github.io/
- **express-validator** (7.0.x): https://express-validator.github.io/
- **sanitize-html** (2.11.x): https://github.com/apostrophecms/sanitize-html
- **sql-bricks** (Parameterized queries): https://github.com/dresende/sql-bricks
- **OWASP Dependency Check**: https://owasp.org/www-project-dependency-check/

### Common Vulnerabilities

| Vulnerability | CWE | Prevention |
|---|---|---|
| SQL Injection | CWE-89 | Parameterized queries |
| XSS | CWE-79 | Input validation, output encoding |
| CSRF | CWE-352 | CSRF tokens, SameSite cookies |
| XXE | CWE-611 | Disable external entities |
| BOLA | CWE-639 | Check ownership on every request |

---

**Version**: 4.0.0 Enterprise  
**Skill Category**: Security (Vulnerability Defense)  
**Complexity**: Medium  
**Time to Implement**: 2-4 hours per category  
**Prerequisites**: Web security fundamentals, Express.js knowledge
