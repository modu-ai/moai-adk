---
name: moai-domain-web-api
description: REST API and GraphQL design with OpenAPI 3.1, authentication, versioning, and rate limiting.
version: 1.0.0
modularized: false
tags:
  - enterprise
  - patterns
  - web
  - architecture
  - api
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: moai, authentication, web, api, domain  


## Quick Reference

### Core API Design Patterns (November 2025)

**REST API Standards:**
- OpenAPI 3.1.0 specification format
- RESTful resource naming conventions
- HTTP verb semantics (GET, POST, PUT, DELETE, PATCH)
- Status code standards (2xx success, 4xx client errors, 5xx server errors)
- JSON response format consistency

**GraphQL Standards:**
- Schema-first design approach
- Query optimization and N+1 prevention
- Mutation naming conventions
- Subscription patterns for real-time updates
- Error handling with extensions

**Authentication & Authorization:**
- JWT tokens with RS256 signing
- OAuth 2.0 with PKCE flow
- API key management for service-to-service
- Role-based access control (RBAC)
- Rate limiting per authentication tier

**Versioning Strategies:**
- URL versioning (`/v1/users`, `/v2/users`)
- Header versioning (`Accept: application/vnd.api+json; version=1`)
- Backward compatibility requirements
- Deprecation timeline communication


## Implementation Guide

### Tool Version Matrix (2025-10-22)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **OpenAPI** | 3.1.0 | API specification | ✅ Current |
| **Postman** | 11.21.0 | API testing | ✅ Current |
| **Swagger UI** | 5.18.2 | Documentation | ✅ Current |
| **GraphQL** | 16.8.1 | Query language | ✅ Current |
| **Apollo Server** | 4.10.0 | GraphQL server | ✅ Current |

### REST API Implementation

```typescript
// OpenAPI 3.1 specification example
{
  "openapi": "3.1.0",
  "info": {
    "title": "User Management API",
    "version": "1.0.0",
    "description": "API for managing user accounts"
  },
  "servers": [
    {
      "url": "https://api.example.com/v1",
      "description": "Production server"
    }
  ],
  "paths": {
    "/users": {
      "get": {
        "summary": "List all users",
        "parameters": [
          {
            "name": "limit",
            "in": "query",
            "schema": { "type": "integer", "default": 20 }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "$ref": "#/components/schemas/User" }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "User": {
        "type": "object",
        "required": ["id", "email"],
        "properties": {
          "id": { "type": "string", "format": "uuid" },
          "email": { "type": "string", "format": "email" },
          "name": { "type": "string" }
        }
      }
    }
  }
}
```

### GraphQL Schema Design

```graphql
type Query {
  user(id: ID!): User
  users(limit: Int = 20, offset: Int = 0): UserConnection!
}

type Mutation {
  createUser(input: CreateUserInput!): User!
  updateUser(id: ID!, input: UpdateUserInput!): User!
  deleteUser(id: ID!): Boolean!
}

type Subscription {
  userCreated: User!
  userUpdated(id: ID!): User!
}

type User {
  id: ID!
  email: String!
  name: String
  createdAt: DateTime!
  updatedAt: DateTime!
}

input CreateUserInput {
  email: String!
  name: String
}

input UpdateUserInput {
  email: String
  name: String
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
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}
```

### Authentication Implementation

```typescript
// JWT authentication middleware
import jwt from 'jsonwebtoken';

interface JWTPayload {
  userId: string;
  email: string;
  roles: string[];
}

export function authenticateRequest(req, res, next) {
  const authHeader = req.headers.authorization;
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Missing or invalid token' });
  }

  const token = authHeader.substring(7);
  
  try {
    const decoded = jwt.verify(token, process.env.JWT_PUBLIC_KEY, {
      algorithms: ['RS256']
    }) as JWTPayload;
    
    req.user = decoded;
    next();
  } catch (error) {
    return res.status(401).json({ error: 'Invalid or expired token' });
  }
}

// Rate limiting middleware
import rateLimit from 'express-rate-limit';

export const apiLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // limit each IP to 100 requests per windowMs
  message: 'Too many requests from this IP, please try again later',
  standardHeaders: true,
  legacyHeaders: false,
});
```

### API Versioning Pattern

```typescript
// URL-based versioning
app.use('/v1', v1Router);
app.use('/v2', v2Router);

// Header-based versioning
app.use((req, res, next) => {
  const acceptHeader = req.headers['accept'];
  const versionMatch = acceptHeader?.match(/version=(\d+)/);
  req.apiVersion = versionMatch ? parseInt(versionMatch[1]) : 1;
  next();
});

// Version-specific handler
function getUserHandler(req, res) {
  if (req.apiVersion === 1) {
    return getUserV1(req, res);
  } else if (req.apiVersion === 2) {
    return getUserV2(req, res);
  } else {
    return res.status(400).json({ error: 'Unsupported API version' });
  }
}
```


## Advanced Patterns

### API Gateway Pattern

```typescript
// API Gateway with service routing
import { createProxyMiddleware } from 'http-proxy-middleware';

const gateway = express();

// Route to user service
gateway.use('/api/users', createProxyMiddleware({
  target: 'http://user-service:3001',
  changeOrigin: true,
  pathRewrite: { '^/api/users': '' },
}));

// Route to order service
gateway.use('/api/orders', createProxyMiddleware({
  target: 'http://order-service:3002',
  changeOrigin: true,
  pathRewrite: { '^/api/orders': '' },
}));

// Centralized authentication
gateway.use(authenticateRequest);

// Centralized rate limiting
gateway.use(apiLimiter);
```

### GraphQL DataLoader Pattern

```typescript
import DataLoader from 'dataloader';

// Batch loading function to prevent N+1 queries
async function batchGetUsers(userIds: string[]) {
  const users = await db.user.findMany({
    where: { id: { in: userIds } }
  });
  
  const userMap = new Map(users.map(u => [u.id, u]));
  return userIds.map(id => userMap.get(id) || null);
}

// Create DataLoader instance
const userLoader = new DataLoader(batchGetUsers);

// Use in GraphQL resolver
const resolvers = {
  Query: {
    user: (_, { id }, { loaders }) => loaders.user.load(id),
  },
  Post: {
    author: (post, _, { loaders }) => loaders.user.load(post.authorId),
  },
};
```

### API Documentation Generation

```typescript
// Generate OpenAPI docs from code
import swaggerJsdoc from 'swagger-jsdoc';
import swaggerUi from 'swagger-ui-express';

const swaggerOptions = {
  definition: {
    openapi: '3.1.0',
    info: {
      title: 'User API',
      version: '1.0.0',
    },
  },
  apis: ['./routes/*.ts'], // Path to API docs
};

const swaggerSpec = swaggerJsdoc(swaggerOptions);

app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerSpec));
```

### Error Handling Standard

```typescript
// Standardized error response format
interface APIError {
  error: {
    code: string;
    message: string;
    details?: any;
    timestamp: string;
    path: string;
  };
}

// Error handling middleware
app.use((err, req, res, next) => {
  const statusCode = err.statusCode || 500;
  const errorResponse: APIError = {
    error: {
      code: err.code || 'INTERNAL_ERROR',
      message: err.message || 'An unexpected error occurred',
      details: process.env.NODE_ENV === 'development' ? err.stack : undefined,
      timestamp: new Date().toISOString(),
      path: req.path,
    },
  };
  
  res.status(statusCode).json(errorResponse);
});
```


## Best Practices

### DO
- Use OpenAPI 3.1 for REST API documentation
- Implement comprehensive input validation
- Apply rate limiting per authentication tier
- Version APIs with clear deprecation timelines
- Use HTTPS for all API endpoints
- Implement proper error handling with meaningful messages
- Cache responses where appropriate
- Use pagination for list endpoints
- Document all API changes in CHANGELOG

### DON'T
- Expose internal error details in production
- Use GET requests for state-changing operations
- Skip authentication/authorization checks
- Return sensitive data in error messages
- Ignore API versioning strategy
- Forget to validate and sanitize inputs
- Use HTTP for production APIs
- Implement custom authentication without security review


## Works Well With

- `moai-foundation-trust` (TRUST 5 quality principles)
- `moai-security-api` (API security patterns)
- `moai-essentials-perf` (Performance optimization)
- `moai-domain-backend` (Backend integration)


## Changelog

- **v2.0.0** (2025-11-21): 3-level structure with comprehensive API patterns
- **v1.0.0** (2025-03-29): Initial release


**End of Skill** | Updated 2025-11-21



## Context7 Integration

### Related Libraries & Tools
- [OpenAPI](/OAI/OpenAPI-Specification): API specification standard
- [Swagger](/swagger-api/swagger-ui): API documentation tool

### Official Documentation
- [Documentation](https://swagger.io/docs/)
- [API Reference](https://swagger.io/docs/specification/about/)

### Version-Specific Guides
Latest stable version: 3.1
- [Release Notes](https://github.com/OAI/OpenAPI-Specification/releases)
- [Migration Guide](https://swagger.io/docs/specification/v3_0_to_v3_1/)
