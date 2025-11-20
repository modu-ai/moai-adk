# API Security Examples

Practical implementation examples for secure API development.

---

## JWT Authentication

### Python/FastAPI Implementation

```python
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer
from datetime import datetime, timedelta
import jwt

SECRET_KEY = "your-secret-key-from-env"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

security = HTTPBearer()

def create_access_token(data: dict) -> str:
    """Create JWT access token."""
    to_encode = data.copy()
    expire = datetime.utcnow() + timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

def verify_token(credentials = Depends(security)) -> str:
    """Verify JWT token and return user_id."""
    try:
        token = credentials.credentials
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        user_id: str = payload.get("sub")
        if user_id is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid authentication credentials"
            )
        return user_id
    except jwt.ExpiredSignatureError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token has expired"
        )
    except jwt.InvalidTokenError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token"
        )

# Usage in endpoint
from fastapi import FastAPI

app = FastAPI()

@app.post("/login")
async def login(username: str, password: str):
    # Verify credentials (simplified)
    if username == "admin" and password == "secret":
        access_token = create_access_token(data={"sub": username})
        return {"access_token": access_token, "token_type": "bearer"}
    raise HTTPException(status_code=401, detail="Invalid credentials")

@app.get("/protected")
async def protected_route(user_id: str = Depends(verify_token)):
    return {"message": f"Hello {user_id}"}
```

### Node.js/Express Implementation

```javascript
const jwt = require("jsonwebtoken");
const express = require("express");

const SECRET_KEY = process.env.JWT_SECRET;
const app = express();

// Create token
function createToken(userId) {
  return jwt.sign({ sub: userId }, SECRET_KEY, { expiresIn: "1h" });
}

// Verify middleware
function verifyToken(req, res, next) {
  const authHeader = req.headers["authorization"];
  const token = authHeader && authHeader.split(" ")[1];

  if (!token) {
    return res.status(401).json({ error: "No token provided" });
  }

  jwt.verify(token, SECRET_KEY, (err, decoded) => {
    if (err) {
      return res.status(401).json({ error: "Invalid token" });
    }
    req.userId = decoded.sub;
    next();
  });
}

// Routes
app.post("/login", (req, res) => {
  const { username, password } = req.body;
  // Verify credentials
  if (username === "admin" && password === "secret") {
    const token = createToken(username);
    res.json({ access_token: token });
  } else {
    res.status(401).json({ error: "Invalid credentials" });
  }
});

app.get("/protected", verifyToken, (req, res) => {
  res.json({ message: `Hello ${req.userId}` });
});
```

---

## RBAC Implementation

### Decorator Pattern (Python)

```python
from functools import wraps
from fastapi import HTTPException, status

def require_role(*required_roles):
    """Decorator to enforce role-based access control."""
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, user_id: str = Depends(verify_token), **kwargs):
            # Fetch user role from database
            user = await get_user(user_id)

            if user.role not in required_roles:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail=f"Requires one of roles: {required_roles}"
                )

            return await func(*args, user_id=user_id, **kwargs)
        return wrapper
    return decorator

# Usage
@app.delete("/users/{user_id}")
@require_role("admin", "moderator")
async def delete_user(user_id: int, current_user: str = Depends(verify_token)):
    # Only admins and moderators can delete users
    await User.delete(user_id)
    return {"status": "deleted"}

@app.get("/analytics")
@require_role("admin")
async def get_analytics(current_user: str = Depends(verify_token)):
    # Only admins can view analytics
    return await Analytics.get_all()
```

### Middleware Pattern (Node.js)

```javascript
function requireRole(...roles) {
  return (req, res, next) => {
    const userRole = req.user.role; // Set by verifyToken middleware

    if (!roles.includes(userRole)) {
      return res.status(403).json({
        error: `Requires one of roles: ${roles.join(", ")}`,
      });
    }

    next();
  };
}

// Usage
app.delete(
  "/users/:id",
  verifyToken,
  requireRole("admin", "moderator"),
  (req, res) => {
    // Delete user
    res.json({ status: "deleted" });
  }
);

app.get("/analytics", verifyToken, requireRole("admin"), (req, res) => {
  // Return analytics
  res.json({ data: [] });
});
```

---

## Rate Limiting

### Token Bucket Algorithm (Python)

```python
import time
from collections import defaultdict
from fastapi import Request, HTTPException

class RateLimiter:
    def __init__(self, requests_per_minute=60):
        self.rate = requests_per_minute / 60.0  # Tokens per second
        self.max_tokens = requests_per_minute
        self.tokens = defaultdict(lambda: self.max_tokens)
        self.last_check = defaultdict(time.time)

    def allow_request(self, identifier: str) -> bool:
        """Check if request is allowed under rate limit."""
        now = time.time()
        time_passed = now - self.last_check[identifier]

        # Refill tokens based on time passed
        self.tokens[identifier] += time_passed * self.rate
        self.tokens[identifier] = min(self.tokens[identifier], self.max_tokens)

        self.last_check[identifier] = now

        # Try to consume one token
        if self.tokens[identifier] >= 1:
            self.tokens[identifier] -= 1
            return True

        return False

# Create rate limiter
rate_limiter = RateLimiter(requests_per_minute=100)

# Middleware
from fastapi import Request

@app.middleware("http")
async def rate_limit_middleware(request: Request, call_next):
    # Use IP address as identifier
    client_ip = request.client.host

    if not rate_limiter.allow_request(client_ip):
        raise HTTPException(
            status_code=429,
            detail="Too many requests. Please try again later.",
            headers={"Retry-After": "60"}
        )

    response = await call_next(request)
    return response
```

### Redis-Based Rate Limiting (Node.js)

```javascript
const redis = require("redis");
const client = redis.createClient();

async function rateLimitMiddleware(req, res, next) {
  const identifier = req.ip;
  const key = `rate_limit:${identifier}`;
  const limit = 100; // requests per minute
  const window = 60; // seconds

  const current = await client.incr(key);

  if (current === 1) {
    // First request, set expiry
    await client.expire(key, window);
  }

  if (current > limit) {
    return res.status(429).json({
      error: "Too many requests",
      retry_after: window,
    });
  }

  next();
}

app.use(rateLimitMiddleware);
```

---

## CORS Setup

### Secure CORS Configuration (FastAPI)

```python
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI()

# Production configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "https://app.example.com",
        "https://admin.example.com"
    ],  # Specific domains only
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Authorization", "Content-Type"],
    max_age=3600,  # Cache preflight for 1 hour
)

# Development configuration (separate file)
if os.getenv("ENV") == "development":
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["http://localhost:3000"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
```

### Express CORS (Node.js)

```javascript
const cors = require("cors");

// Production
const corsOptions = {
  origin: ["https://app.example.com", "https://admin.example.com"],
  credentials: true,
  methods: ["GET", "POST", "PUT", "DELETE"],
  allowedHeaders: ["Authorization", "Content-Type"],
  maxAge: 3600,
};

app.use(cors(corsOptions));

// Development
if (process.env.NODE_ENV === "development") {
  app.use(
    cors({
      origin: "http://localhost:3000",
      credentials: true,
    })
  );
}
```

---

## Input Validation

### Pydantic Models (Python)

```python
from pydantic import BaseModel, EmailStr, Field, constr, validator

class UserCreate(BaseModel):
    email: EmailStr
    password: constr(min_length=8, max_length=100)
    age: int = Field(ge=18, le=120, description="User age")
    username: constr(regex=r'^[a-zA-Z0-9_-]{3,20}$')

    @validator('password')
    def password_strength(cls, v):
        if not any(c.isupper() for c in v):
            raise ValueError('Password must contain uppercase')
        if not any(c.isdigit() for c in v):
            raise ValueError('Password must contain digit')
        return v

@app.post("/users")
async def create_user(user: UserCreate):
    # Input is automatically validated
    hashed_password = hash_password(user.password)
    new_user = await User.create(
        email=user.email,
        password=hashed_password,
        age=user.age
    )
    return {"id": new_user.id}
```

### Zod Validation (TypeScript)

```typescript
import { z } from "zod";

const userCreateSchema = z.object({
  email: z.string().email(),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(100)
    .refine((pwd) => /[A-Z]/.test(pwd), "Must contain uppercase")
    .refine((pwd) => /[0-9]/.test(pwd), "Must contain digit"),
  age: z.number().int().min(18).max(120),
  username: z.string().regex(/^[a-zA-Z0-9_-]{3,20}$/),
});

type UserCreate = z.infer<typeof userCreateSchema>;

app.post("/users", async (req, res) => {
  try {
    const userData = userCreateSchema.parse(req.body);
    const hashedPassword = await hashPassword(userData.password);
    const newUser = await User.create({
      ...userData,
      password: hashedPassword,
    });
    res.json({ id: newUser.id });
  } catch (error) {
    if (error instanceof z.ZodError) {
      res.status(400).json({ errors: error.errors });
    } else {
      res.status(500).json({ error: "Internal server error" });
    }
  }
});
```

---

**See also**: [reference.md](./reference.md) for complete API reference and advanced patterns.
