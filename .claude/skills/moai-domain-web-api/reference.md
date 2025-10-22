# Web API Design Reference

> REST, GraphQL, authentication, and API best practices

---

## Official Documentation Links

| Technology | Documentation | Status |
|------------|--------------|--------|
| **REST API** | https://restfulapi.net/ | ✅ Current (2025) |
| **GraphQL** | https://graphql.org/learn/ | ✅ Current (2025) |
| **OpenAPI** | https://spec.openapis.org/oas/latest.html | ✅ Current (2025) |
| **OAuth 2.0** | https://oauth.net/2/ | ✅ Current (2025) |

---

## REST API Best Practices

### HTTP Methods
- **GET**: Retrieve resources
- **POST**: Create resources
- **PUT**: Update (replace) resources
- **PATCH**: Partial update
- **DELETE**: Remove resources

### Status Codes
- **200 OK**: Success
- **201 Created**: Resource created
- **400 Bad Request**: Invalid input
- **401 Unauthorized**: Authentication required
- **403 Forbidden**: Access denied
- **404 Not Found**: Resource not found
- **500 Internal Server Error**: Server error

### Example REST Endpoint
```typescript
// GET /api/users/:id
app.get('/api/users/:id', async (req, res) => {
  const user = await db.findUser(req.params.id);
  if (!user) {
    return res.status(404).json({ error: 'User not found' });
  }
  res.json(user);
});
```

---

## GraphQL Best Practices

### Schema Definition
```graphql
type User {
  id: ID!
  name: String!
  email: String!
  posts: [Post!]!
}

type Post {
  id: ID!
  title: String!
  content: String!
  author: User!
}

type Query {
  user(id: ID!): User
  users: [User!]!
}

type Mutation {
  createUser(name: String!, email: String!): User!
}
```

---

## Authentication

### JWT
```typescript
import jwt from 'jsonwebtoken';

// Generate token
const token = jwt.sign({ userId: 123 }, process.env.JWT_SECRET, {
  expiresIn: '1h'
});

// Verify token
const decoded = jwt.verify(token, process.env.JWT_SECRET);
```

---

**Last Updated**: 2025-10-22
**Standards**: REST, GraphQL, OpenAPI 3.1, OAuth 2.0
