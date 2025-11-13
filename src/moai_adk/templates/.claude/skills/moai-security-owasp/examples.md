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

