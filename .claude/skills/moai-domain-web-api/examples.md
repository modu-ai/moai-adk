# Web API Design Examples

## Example 1: RESTful API with FastAPI

**Complete CRUD API implementation**:

```python
from fastapi import FastAPI, HTTPException, Query, Path
from pydantic import BaseModel, EmailStr, Field
from typing import List, Optional
from datetime import datetime
import uvicorn

app = FastAPI(
    title="User Management API",
    version="1.0.0",
    description="RESTful API for user management"
)

# Data models
class UserBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    email: EmailStr
    age: int = Field(..., ge=18, le=120)

class UserCreate(UserBase):
    password: str = Field(..., min_length=8)

class UserUpdate(BaseModel):
    name: Optional[str] = None
    email: Optional[EmailStr] = None
    age: Optional[int] = None

class User(UserBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

# In-memory storage (replace with database)
users_db: dict = {}
next_user_id = 1

# API Endpoints
@app.post("/users", response_model=User, status_code=201)
async def create_user(user: UserCreate):
    """Create a new user"""
    global next_user_id
    user_id = next_user_id
    next_user_id += 1

    db_user = {
        "id": user_id,
        **user.dict(),
        "created_at": datetime.utcnow(),
        "updated_at": datetime.utcnow()
    }
    users_db[user_id] = db_user
    return db_user

@app.get("/users", response_model=List[User])
async def list_users(
    skip: int = Query(0, ge=0),
    limit: int = Query(10, ge=1, le=100)
):
    """List all users with pagination"""
    user_list = list(users_db.values())
    return user_list[skip:skip + limit]

@app.get("/users/{user_id}", response_model=User)
async def get_user(user_id: int = Path(..., gt=0)):
    """Get user by ID"""
    if user_id not in users_db:
        raise HTTPException(status_code=404, detail="User not found")
    return users_db[user_id]

@app.put("/users/{user_id}", response_model=User)
async def update_user(user_id: int, user: UserUpdate):
    """Update user information"""
    if user_id not in users_db:
        raise HTTPException(status_code=404, detail="User not found")

    db_user = users_db[user_id]
    update_data = user.dict(exclude_unset=True)
    db_user.update({**update_data, "updated_at": datetime.utcnow()})
    return db_user

@app.delete("/users/{user_id}", status_code=204)
async def delete_user(user_id: int):
    """Delete user"""
    if user_id not in users_db:
        raise HTTPException(status_code=404, detail="User not found")
    del users_db[user_id]

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

---

## Example 2: GraphQL API with Apollo Server

**Full GraphQL schema and resolvers**:

```typescript
import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';

const typeDefs = `
  type Query {
    user(id: ID!): User
    users(first: Int = 10, after: String): UserConnection
    search(query: String!): [SearchResult]
  }

  type Mutation {
    createUser(input: CreateUserInput!): User
    updateUser(id: ID!, input: UpdateUserInput!): User
    deleteUser(id: ID!): Boolean
  }

  type User {
    id: ID!
    name: String!
    email: String!
    createdAt: String!
    posts(first: Int = 5): [Post]
  }

  type Post {
    id: ID!
    title: String!
    content: String!
    author: User!
    createdAt: String!
  }

  type UserConnection {
    edges: [UserEdge!]!
    pageInfo: PageInfo!
    totalCount: Int!
  }

  type UserEdge {
    node: User!
    cursor: String!
  }

  type PageInfo {
    hasNextPage: Boolean!
    endCursor: String
  }

  union SearchResult = User | Post

  input CreateUserInput {
    name: String!
    email: String!
  }

  input UpdateUserInput {
    name: String
    email: String
  }
`;

// Resolvers
const resolvers = {
  Query: {
    user: async (parent, args) => {
      return await db.users.findById(args.id);
    },
    users: async (parent, args) => {
      const users = await db.users.find();
      return {
        edges: users.map(user => ({
          node: user,
          cursor: Buffer.from(user.id).toString('base64')
        })),
        pageInfo: {
          hasNextPage: users.length >= args.first,
          endCursor: Buffer.from(users[users.length - 1].id).toString('base64')
        },
        totalCount: users.length
      };
    },
    search: async (parent, args) => {
      const results = await db.search(args.query);
      return results;
    }
  },
  Mutation: {
    createUser: async (parent, args) => {
      return await db.users.create(args.input);
    },
    updateUser: async (parent, args) => {
      return await db.users.update(args.id, args.input);
    },
    deleteUser: async (parent, args) => {
      await db.users.delete(args.id);
      return true;
    }
  },
  User: {
    posts: async (user) => {
      return await db.posts.find({ author_id: user.id });
    }
  }
};

const server = new ApolloServer({ typeDefs, resolvers });
const { url } = await startStandaloneServer(server, {
  listen: { port: 4000 }
});

console.log(`GraphQL server running at ${url}`);
```

---

## Example 3: API Versioning Strategy

**Multiple API versions with deprecation**:

```typescript
import express from 'express';

const app = express();

// Version 1 (Deprecated)
app.get('/v1/users/:id', (req, res) => {
    // Old response format
    res.json({
        id: 123,
        name: "John Doe",
        email: "john@example.com"
        // Missing: created_at, updated_at
    });
    res.setHeader('Deprecation', 'true');
    res.setHeader('Sunset', 'Wed, 31 Dec 2025 23:59:59 GMT');
    res.setHeader('Link', '</v2/users/123>; rel="successor-version"');
});

// Version 2 (Current)
app.get('/v2/users/:id', (req, res) => {
    // Enhanced response format
    res.json({
        id: 123,
        name: "John Doe",
        email: "john@example.com",
        created_at: "2024-01-15T10:30:00Z",
        updated_at: "2025-11-22T14:00:00Z",
        status: "active"
    });
});

// Header-based versioning alternative
app.get('/users/:id', (req, res) => {
    const version = req.get('Accept').match(/version=(\d+)/)?.[1] || '2';

    if (version === '1') {
        return res.json({
            id: 123,
            name: "John Doe",
            email: "john@example.com"
        });
    }

    // v2
    res.json({
        id: 123,
        name: "John Doe",
        email: "john@example.com",
        created_at: "2024-01-15T10:30:00Z",
        updated_at: "2025-11-22T14:00:00Z",
        status: "active"
    });
});
```

---

## Example 4: Authentication with JWT

**Secure JWT authentication flow**:

```python
from fastapi import FastAPI, HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthCredentials
import jwt
from datetime import datetime, timedelta
from typing import Optional

app = FastAPI()
security = HTTPBearer()

SECRET_KEY = "your-secret-key-change-in-production"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None):
    """Create JWT token"""
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=15)

    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

def verify_token(credentials: HTTPAuthCredentials = Depends(security)):
    """Verify JWT token"""
    token = credentials.credentials
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        user_id: int = payload.get("sub")
        if user_id is None:
            raise HTTPException(status_code=401, detail="Invalid token")
    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=401, detail="Token expired")
    except jwt.InvalidTokenError:
        raise HTTPException(status_code=401, detail="Invalid token")

    return user_id

@app.post("/login")
async def login(username: str, password: str):
    """Login endpoint - return JWT token"""
    # Verify credentials (simplified)
    if not await verify_password(username, password):
        raise HTTPException(status_code=401, detail="Invalid credentials")

    access_token_expires = timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    access_token = create_access_token(
        data={"sub": username},
        expires_delta=access_token_expires
    )
    return {"access_token": access_token, "token_type": "bearer"}

@app.get("/protected")
async def protected_route(user_id: int = Depends(verify_token)):
    """Protected endpoint requiring valid JWT"""
    return {"message": f"Hello user {user_id}"}
```

---

## Example 5: Error Handling Standards

**RFC 7807 Problem Details error responses**:

```typescript
import express from 'express';
import { v4 as uuidv4 } from 'uuid';

const app = express();

// Error handler middleware
app.use((err, req, res, next) => {
    const problemDetail = {
        type: `https://api.example.com/errors/${err.code || 'internal-error'}`,
        title: err.title || "An error occurred",
        status: err.statusCode || 500,
        detail: err.message,
        instance: req.path,
        timestamp: new Date().toISOString(),
        request_id: uuidv4(),
        ...(err.errors && { errors: err.errors })
    };

    res.status(problemDetail.status).json(problemDetail);
});

// Validation error example
app.post('/users', (req, res, next) => {
    const errors = validateUser(req.body);
    if (errors.length > 0) {
        return next({
            statusCode: 422,
            code: 'validation-error',
            title: 'Validation Error',
            message: 'Request validation failed',
            errors: errors.map(e => ({
                field: e.field,
                message: e.message,
                code: e.code
            }))
        });
    }
});

// Conflict error example
app.put('/users/:id', (req, res, next) => {
    if (!isLatestVersion(req.body, req.headers['if-match'])) {
        return next({
            statusCode: 409,
            code: 'conflict-error',
            title: 'Conflict',
            message: 'Resource has been modified. Please refresh and try again.'
        });
    }
});
```

---

## Example 6: Rate Limiting

**API rate limiting with tiered access**:

```python
from fastapi import FastAPI, HTTPException
from slowapi import Limiter
from slowapi.util import get_remote_address
from slowapi.errors import RateLimitExceeded

app = FastAPI()
limiter = Limiter(key_func=get_remote_address)

# Rate limit configurations
RATE_LIMITS = {
    'free': "10/minute",
    'pro': "100/minute",
    'enterprise': "1000/minute"
}

@app.exception_handler(RateLimitExceeded)
async def rate_limit_handler(request, exc):
    return {
        "error": "rate_limit_exceeded",
        "message": f"Rate limit exceeded: {exc.detail}",
        "retry_after": 60
    }

@app.get("/api/free-endpoint")
@limiter.limit(RATE_LIMITS['free'])
async def free_endpoint(request):
    return {"message": "Free tier response"}

@app.get("/api/pro-endpoint")
@limiter.limit(RATE_LIMITS['pro'])
async def pro_endpoint(request):
    return {"message": "Pro tier response"}

# Custom rate limiting per user
def get_user_tier(user_id: int) -> str:
    # Fetch from database
    return "pro"

def rate_limit_by_user(request):
    user_id = extract_user_id(request)
    tier = get_user_tier(user_id)
    return RATE_LIMITS[tier]

@app.get("/api/user-data")
@limiter.limit(rate_limit_by_user)
async def user_data(request):
    return {"data": "user-specific-data"}
```

---

## Example 7: OpenAPI Documentation

**Comprehensive OpenAPI 3.1 specification**:

```yaml
openapi: 3.1.0
info:
  title: User Management API
  version: 2.0.0
  contact:
    name: API Support
    url: https://support.example.com
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html

servers:
  - url: https://api.example.com/v2
    description: Production
  - url: https://staging-api.example.com/v2
    description: Staging

paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      tags:
        - Users
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            minimum: 1
            maximum: 100
        - name: offset
          in: query
          schema:
            type: integer
            default: 0
      responses:
        '200':
          description: List of users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
        '401':
          $ref: '#/components/responses/Unauthorized'

    post:
      summary: Create user
      operationId: createUser
      tags:
        - Users
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '422':
          $ref: '#/components/responses/ValidationError'

components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
        - email
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
        created_at:
          type: string
          format: date-time

    CreateUserRequest:
      type: object
      required:
        - name
        - email
        - password
      properties:
        name:
          type: string
        email:
          type: string
          format: email
        password:
          type: string
          format: password
          minLength: 8

  responses:
    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'

    ValidationError:
      description: Validation failed
      content:
        application/json:
          schema:
            type: object
            properties:
              errors:
                type: array
                items:
                  $ref: '#/components/schemas/FieldError'

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

---

## Example 8: Content Negotiation

**Support multiple response formats**:

```typescript
import express from 'express';
import { json2csv } from 'json2csv';
import xml from 'fast-xml-parser';

const app = express();

@app.get('/data/:id', (req, res) => {
    const data = {
        id: req.params.id,
        name: 'John Doe',
        email: 'john@example.com',
        created_at: new Date().toISOString()
    };

    // Determine response format from Accept header
    const acceptHeader = req.get('Accept') || 'application/json';

    if (acceptHeader.includes('application/xml')) {
        // XML response
        const xmlData = xml.j2xParser().parse(data);
        return res.type('application/xml').send(xmlData);
    }

    if (acceptHeader.includes('text/csv')) {
        // CSV response
        const csvData = json2csv({ data: [data] });
        return res.type('text/csv').send(csvData);
    }

    // Default to JSON
    return res.json(data);
});
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
