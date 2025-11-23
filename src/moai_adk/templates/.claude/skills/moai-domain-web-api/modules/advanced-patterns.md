# Advanced Web API Patterns

## API Versioning Strategies

### Strategy 1: URL Path Versioning

**Most common and explicit approach**:
```yaml
openapi: 3.1.0
info:
  title: User Management API
  version: "2.0.0"
servers:
  - url: https://api.example.com/v1
    description: "Version 1 (Deprecated - End of Life 2025-12-31)"
  - url: https://api.example.com/v2
    description: "Version 2 (Current)"

paths:
  /v1/users:
    get:
      summary: List users (v1 - deprecated)
      deprecated: true
      responses:
        "200":
          description: List of users (old format)
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/UserV1"

  /v2/users:
    get:
      summary: List users (v2 - current)
      responses:
        "200":
          description: List of users with enhanced fields
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/UserV2"

components:
  schemas:
    UserV1:
      type: object
      properties:
        id: { type: integer }
        name: { type: string }
        email: { type: string }

    UserV2:
      type: object
      properties:
        id: { type: integer }
        name: { type: string }
        email: { type: string }
        created_at: { type: string, format: date-time }
        updated_at: { type: string, format: date-time }
        status: { type: string, enum: [active, inactive] }
```

### Strategy 2: Header-Based Versioning

**Client specifies version via Accept header**:
```javascript
// Client request
fetch('https://api.example.com/users', {
    headers: {
        'Accept': 'application/vnd.api+json;version=2'
    }
})

// Server implementation (Express.js)
app.get('/users', (req, res) => {
    const apiVersion = req.get('Accept').match(/version=(\d+)/)?.[1] || '1';

    if (apiVersion === '2') {
        return res.json(getUsersV2());
    }
    res.json(getUsersV1());
});
```

**OpenAPI 3.1 representation**:
```yaml
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: Accept
          in: header
          schema:
            type: string
            pattern: 'application/vnd\.api\+json;version=\d+'
          example: "application/vnd.api+json;version=2"
      responses:
        "200":
          description: Success (content varies by Accept header)
```

---

## GraphQL Advanced Patterns

### N+1 Query Prevention with DataLoader

**Problem**: Fetching related data causes N+1 queries.

```typescript
import DataLoader from 'dataloader';

// Create DataLoader for batch loading users
const userLoader = new DataLoader(async (userIds) => {
    // Load all users in a single query
    const users = await db.users.find({ id: { $in: userIds } });
    // Return in same order as requested
    return userIds.map(id => users.find(u => u.id === id));
});

// GraphQL resolver
const resolvers = {
    Post: {
        async author(post, args, context) {
            // DataLoader batches these calls
            return userLoader.load(post.author_id);
        }
    },
    Query: {
        async posts(root, args, context) {
            return db.posts.find();
        }
    }
};

// Query execution
const query = `
    query {
        posts {
            title
            author {
                name
                email
            }
        }
    }
`;

// Without DataLoader: 1 query for posts + N queries for authors
// With DataLoader: 1 query for posts + 1 query for all authors
```

### Connection-Based Pagination (Relay Spec)

```graphql
type Query {
    users(
        first: Int
        after: String
        last: Int
        before: String
    ): UserConnection!
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

type User {
    id: ID!
    name: String!
    email: String!
}
```

**Implementation**:
```typescript
function resolvePaginatedUsers(root, args, context) {
    const { first = 10, after = null, last = null, before = null } = args;

    // Decode cursors
    const afterIndex = after ? Buffer.from(after, 'base64').toString() : null;
    const beforeIndex = before ? Buffer.from(before, 'base64').toString() : null;

    // Get users from database
    let users = context.db.users.findAll();

    // Apply cursor-based filtering
    if (afterIndex !== null) {
        users = users.filter(u => u.id > afterIndex);
    }
    if (beforeIndex !== null) {
        users = users.filter(u => u.id < beforeIndex);
    }

    // Slice for pagination
    const edges = users.slice(0, first || last).map(user => ({
        node: user,
        cursor: Buffer.from(user.id.toString()).toString('base64')
    }));

    return {
        edges,
        pageInfo: {
            hasNextPage: users.length > (first || last),
            hasPreviousPage: afterIndex !== null || beforeIndex !== null,
            startCursor: edges[0]?.cursor || null,
            endCursor: edges[edges.length - 1]?.cursor || null
        },
        totalCount: context.db.users.count()
    };
}
```

---

## Authentication Patterns

### OAuth 2.0 with PKCE Flow

**Frontend Flow** (most secure for single-page apps):
```javascript
// 1. Generate code challenge
const generateCodeChallenge = async () => {
    const code_verifier = generateRandomString(128);
    const encoder = new TextEncoder();
    const data = encoder.encode(code_verifier);
    const hash = await crypto.subtle.digest('SHA-256', data);
    const code_challenge = btoa(String.fromCharCode(...new Uint8Array(hash)))
        .replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    return { code_verifier, code_challenge };
};

// 2. Redirect to authorization server
const { code_verifier, code_challenge } = await generateCodeChallenge();
const authUrl = new URL('https://auth.example.com/authorize');
authUrl.searchParams.append('client_id', CLIENT_ID);
authUrl.searchParams.append('redirect_uri', window.location.origin + '/callback');
authUrl.searchParams.append('response_type', 'code');
authUrl.searchParams.append('scope', 'openid profile email');
authUrl.searchParams.append('code_challenge', code_challenge);
authUrl.searchParams.append('code_challenge_method', 'S256');

window.location.href = authUrl.toString();

// 3. Handle callback and exchange code for token
const code = new URLSearchParams(window.location.search).get('code');
const tokenResponse = await fetch('https://auth.example.com/token', {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
        client_id: CLIENT_ID,
        code: code,
        code_verifier: code_verifier,
        grant_type: 'authorization_code'
    })
});

const { access_token, id_token } = await tokenResponse.json();
localStorage.setItem('access_token', access_token);
```

### API Key Authentication with Rate Limiting

```yaml
openapi: 3.1.0
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
      description: API Key for service-to-service authentication

security:
  - ApiKeyAuth: []

paths:
  /data:
    get:
      summary: Get data
      security:
        - ApiKeyAuth: []
      x-rate-limit:
        limit: 1000
        window: 3600  # per hour
      responses:
        "200":
          description: Success
        "429":
          description: Rate limit exceeded
          headers:
            X-Rate-Limit-Limit:
              schema:
                type: integer
              description: Request limit per window
            X-Rate-Limit-Remaining:
              schema:
                type: integer
              description: Remaining requests in window
            X-Rate-Limit-Reset:
              schema:
                type: integer
              description: Unix timestamp of window reset
```

---

## Error Handling Standards

### Structured Error Responses

```typescript
// Error response format (RFC 7807 Problem Details)
{
    "type": "https://api.example.com/errors/validation-error",
    "title": "Validation Error",
    "status": 422,
    "detail": "Invalid email format",
    "instance": "/users",
    "errors": [
        {
            "field": "email",
            "message": "Invalid email format",
            "code": "INVALID_FORMAT"
        },
        {
            "field": "age",
            "message": "Must be >= 18",
            "code": "CONSTRAINT_VIOLATION"
        }
    ],
    "timestamp": "2025-11-22T10:30:00Z",
    "request_id": "req-abc-123"
}
```

**OpenAPI 3.1 Schema**:
```yaml
components:
  schemas:
    ProblemDetail:
      type: object
      properties:
        type:
          type: string
          format: uri
        title:
          type: string
        status:
          type: integer
        detail:
          type: string
        instance:
          type: string
        errors:
          type: array
          items:
            $ref: "#/components/schemas/FieldError"
        timestamp:
          type: string
          format: date-time
        request_id:
          type: string

    FieldError:
      type: object
      properties:
        field:
          type: string
        message:
          type: string
        code:
          type: string
```

---

## Content Negotiation

### Multiple Media Type Support

```typescript
// Express.js implementation
app.get('/api/users/:id', (req, res) => {
    const user = getUser(req.params.id);

    // Check Accept header
    if (req.accepts('json')) {
        res.json(user);
    } else if (req.accepts('xml')) {
        res.type('application/xml').send(convertToXml(user));
    } else if (req.accepts('csv')) {
        res.type('text/csv').send(convertToCsv(user));
    } else {
        res.status(406).send('Not Acceptable');
    }
});
```

**OpenAPI Documentation**:
```yaml
paths:
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: User details
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
            application/xml:
              schema:
                $ref: "#/components/schemas/User"
            text/csv:
              schema:
                type: string
```

---

## API Filtering and Searching

### Advanced Query Parameters

```typescript
// Support complex filtering
GET /api/products?
    filter[price][gte]=100&
    filter[price][lte]=500&
    filter[category]=electronics&
    filter[in_stock]=true&
    sort=-price,name&
    fields=name,price,category&
    page[number]=1&
    page[size]=20

// Implementation
app.get('/api/products', (req, res) => {
    let query = db.products;

    // Handle filters
    if (req.query.filter) {
        Object.entries(req.query.filter).forEach(([field, conditions]) => {
            if (typeof conditions === 'object') {
                // Range queries
                if (conditions.gte) query = query.where(field, '>=', conditions.gte);
                if (conditions.lte) query = query.where(field, '<=', conditions.lte);
            } else {
                // Exact match
                query = query.where(field, conditions);
            }
        });
    }

    // Handle sorting
    if (req.query.sort) {
        req.query.sort.split(',').forEach(field => {
            const order = field.startsWith('-') ? 'desc' : 'asc';
            const fieldName = field.replace('-', '');
            query = query.orderBy(fieldName, order);
        });
    }

    // Handle field selection
    if (req.query.fields) {
        query = query.select(req.query.fields.split(','));
    }

    // Handle pagination
    const page = req.query.page?.number || 1;
    const pageSize = req.query.page?.size || 20;
    query = query.offset((page - 1) * pageSize).limit(pageSize);

    res.json(query.get());
});
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
