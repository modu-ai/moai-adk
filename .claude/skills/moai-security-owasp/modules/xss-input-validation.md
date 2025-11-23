# XSS & Input Validation

## Overview

Cross-Site Scripting (XSS) and input validation failures allow attackers to inject malicious scripts into web applications.

## XSS Attack Types

### Reflected XSS
Malicious script reflected from user input immediately:
```javascript
// VULNERABLE
app.get('/search', (req, res) => {
  const query = req.query.q;
  res.send(`<p>Search results for: ${query}</p>`);
  // Attack: ?q=<script>alert('XSS')</script>
});
```

### Stored XSS
Malicious script stored in database and executed later:
```javascript
// VULNERABLE
app.post('/comments', (req, res) => {
  const comment = req.body.comment;
  db.comments.create({ text: comment }); // Stored without sanitization
  res.json({ success: true });
});
```

### DOM-Based XSS
Client-side script manipulation:
```javascript
// VULNERABLE
const userInput = location.hash.substring(1);
document.getElementById('output').innerHTML = userInput;
// Attack: #<img src=x onerror=alert('XSS')>
```

## Remediation Patterns

### Input Validation

**Express Validator**:
```javascript
const { body, validationResult } = require('express-validator');

const commentValidator = [
  body('comment')
    .trim()
    .isLength({ min: 1, max: 500 })
    .withMessage('Comment must be 1-500 characters'),
  body('comment')
    .matches(/^[a-zA-Z0-9\s.,!?'-]+$/)
    .withMessage('Comment contains invalid characters')
];

app.post('/comments', commentValidator, (req, res) => {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    return res.status(400).json({ errors: errors.array() });
  }

  // Safe to process
  processComment(req.body.comment);
});
```

### Output Encoding

**Template Engines (Auto-escaping)**:
```javascript
// EJS (auto-escapes)
<p>Comment: <%= comment %></p>

// Handlebars (auto-escapes)
<p>Comment: {{ comment }}</p>

// React (auto-escapes)
<p>Comment: {comment}</p>
```

**Manual Encoding**:
```javascript
function escapeHtml(text) {
  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  };
  return text.replace(/[&<>"']/g, m => map[m]);
}
```

### HTML Sanitization

**Sanitize-HTML**:
```javascript
const sanitizeHtml = require('sanitize-html');

app.post('/comments', (req, res) => {
  const sanitized = sanitizeHtml(req.body.comment, {
    allowedTags: ['b', 'i', 'em', 'strong', 'p'],
    allowedAttributes: {},
    disallowedTagsMode: 'discard'
  });

  db.comments.create({ text: sanitized });
  res.json({ success: true });
});
```

**DOMPurify (Client-side)**:
```javascript
import DOMPurify from 'dompurify';

const clean = DOMPurify.sanitize(userInput);
document.getElementById('output').innerHTML = clean;
```

### Content Security Policy (CSP)

**Strict CSP Configuration**:
```javascript
app.use(helmet.contentSecurityPolicy({
  directives: {
    defaultSrc: ["'self'"],
    scriptSrc: ["'self'", "'nonce-{random}'"],
    styleSrc: ["'self'", "'nonce-{random}'"],
    imgSrc: ["'self'", "data:", "https:"],
    fontSrc: ["'self'"],
    objectSrc: ["'none'"],
    baseUri: ["'self'"],
    formAction: ["'self'"],
    frameAncestors: ["'none'"],
    upgradeInsecureRequests: []
  }
}));
```

**CSP with Nonce**:
```javascript
app.use((req, res, next) => {
  res.locals.nonce = crypto.randomBytes(16).toString('base64');
  next();
});

// In template
<script nonce="<%= nonce %>">
  // Inline script allowed
</script>
```

## CSRF Protection

### CSRF Token Pattern

**Express with csurf**:
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
  res.json({ success: true });
});
```

### SameSite Cookies

```javascript
res.cookie('session', token, {
  sameSite: 'strict',  // No cross-site requests
  secure: true,        // HTTPS only
  httpOnly: true,      // No JavaScript access
  maxAge: 3600000      // 1 hour
});
```

## Validation Best Practices

### Whitelist Validation

```javascript
// Email validation
const emailSchema = z.string().email();

// URL validation
const urlSchema = z.string().url().refine(url => {
  const allowed = ['example.com', 'cdn.example.com'];
  return allowed.some(domain => url.includes(domain));
});

// Phone number validation
const phoneSchema = z.string().regex(/^\+?[1-9]\d{1,14}$/);
```

### Length Limits

```javascript
const schema = z.object({
  username: z.string().min(3).max(20),
  email: z.string().email().max(255),
  bio: z.string().max(500),
  url: z.string().url().max(2048)
});
```

### File Upload Validation

```javascript
const multer = require('multer');

const upload = multer({
  limits: {
    fileSize: 5 * 1024 * 1024 // 5MB
  },
  fileFilter: (req, file, cb) => {
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];

    if (!allowedTypes.includes(file.mimetype)) {
      return cb(new Error('Invalid file type'));
    }

    cb(null, true);
  }
});

app.post('/upload', upload.single('image'), (req, res) => {
  // Additional validation
  const magic = req.file.buffer.slice(0, 4);
  if (!isValidImageMagic(magic)) {
    return res.status(400).json({ error: 'Invalid image' });
  }

  res.json({ success: true });
});
```

## Testing

```javascript
describe('XSS Protection', () => {
  test('should escape HTML entities', () => {
    const malicious = '<script>alert("XSS")</script>';
    const escaped = escapeHtml(malicious);
    expect(escaped).toBe('&lt;script&gt;alert(&quot;XSS&quot;)&lt;/script&gt;');
  });

  test('should sanitize HTML tags', () => {
    const input = '<p>Hello <script>alert("XSS")</script></p>';
    const sanitized = sanitizeHtml(input, { allowedTags: ['p'] });
    expect(sanitized).toBe('<p>Hello </p>');
  });

  test('should reject CSRF without token', async () => {
    const response = await request(app)
      .post('/form')
      .send({ data: 'test' });

    expect(response.status).toBe(403);
  });
});
```

## Validation Checklist

- [ ] All user input validated with whitelists
- [ ] Output encoded in templates
- [ ] CSP headers configured
- [ ] CSRF protection enabled
- [ ] SameSite cookies configured
- [ ] File upload validation implemented
- [ ] Length limits enforced
- [ ] Dangerous HTML tags sanitized

---

**Last Updated**: 2025-11-24
**OWASP Category**: A03:2021 (Injection)
**CWE**: CWE-79 (Cross-site Scripting), CWE-352 (CSRF)
