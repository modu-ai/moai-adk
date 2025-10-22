# Application Security - Working Examples

> Real-world security implementations

---

## Example 1: SQL Injection Prevention

### Bad (Vulnerable)
```python
# ❌ VULNERABLE
query = f"SELECT * FROM users WHERE email = '{user_input}'"
cursor.execute(query)
```

### Good (Parameterized)
```python
# ✅ SECURE
query = "SELECT * FROM users WHERE email = %s"
cursor.execute(query, (user_input,))
```

---

## Example 2: Password Hashing

### bcrypt
```python
import bcrypt

# Hash password
password = b"my_password"
hashed = bcrypt.hashpw(password, bcrypt.gensalt())

# Verify password
if bcrypt.checkpw(password, hashed):
    print("Password correct")
```

---

**Last Updated**: 2025-10-22
**Standards**: OWASP Top 10 2025
