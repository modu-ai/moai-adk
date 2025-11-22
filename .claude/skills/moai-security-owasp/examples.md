# moai-security-owasp: Production Examples

## Example 1: Complete Input Validation & Sanitization

```javascript
const { body, validationResult } = require('express-validator');
const sanitizeHtml = require('sanitize-html');

// Validation middleware
const commentValidator = [
  body('comment')
    .trim()
    .isLength({ min: 1, max: 500 })
    .withMessage('Comment must be 1-500 characters'),
  body('rating')
    .isInt({ min: 1, max: 5 })
    .withMessage('Rating must be 1-5'),
  body('email')
    .isEmail()
    .normalizeEmail()
];

app.post('/comments', commentValidator, (req, res) => {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    return res.status(400).json({ errors });
  }
  
  // Additional sanitization
  const sanitized = sanitizeHtml(req.body.comment, {
    allowedTags: ['b', 'i', 'em', 'strong'],
    allowedAttributes: {}
  });
  
  db.comments.create({
    content: sanitized,
    rating: req.body.rating,
    email: req.body.email
  });
  
  res.json({ success: true });
});
```

## Example 2: SQL Injection Prevention

```javascript
// VULNERABLE
app.get('/users/:id', (req, res) => {
  const query = `SELECT * FROM users WHERE id = ${req.params.id}`;
  db.query(query); // SQL Injection!
});

// SECURE: Parameterized query
app.get('/users/:id', (req, res) => {
  const query = 'SELECT * FROM users WHERE id = ?';
  db.query(query, [req.params.id]);
});

// SECURE: ORM (Type-safe)
const user = await User.findByPk(req.params.id);
```


## Example 3: XSS Prevention with Content Security Policy

```javascript
// Configure CSP headers
app.use((req, res, next) => {
  res.setHeader(
    'Content-Security-Policy',
    "default-src 'self'; script-src 'self' 'nonce-xyz'; style-src 'self'"
  );
  next();
});

// Escape output
const escaped = escapeHtml(userInput);
res.send(`<p>${escaped}</p>`);
```

## Example 4: Authentication & Authorization

```python
from flask import Flask, session, abort
from functools import wraps

app = Flask(__name__)

def require_auth(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        if 'user_id' not in session:
            abort(401)
        return f(*args, **kwargs)
    return decorated

@app.route('/admin')
@require_auth
def admin_panel():
    user = get_user(session['user_id'])
    if user.role != 'admin':
        abort(403)
    return 'Admin Panel'
```

## Example 5: CSRF Protection

```python
from flask_wtf.csrf import CSRFProtect

csrf = CSRFProtect()
csrf.init_app(app)

@app.route('/update', methods=['POST'])
@csrf.protect
def update_profile():
    # CSRF token automatically validated
    return update_user(request.form)
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
**Enterprise Ready**: Yes
