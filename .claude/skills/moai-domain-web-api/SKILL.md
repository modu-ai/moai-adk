---
name: moai-domain-web-api
version: 4.0.0
created: 2025-11-12
updated: 2025-11-12
status: active
tier: domain
description: "Enterprise-grade Web API expertise with stable REST/GraphQL frameworks (FastAPI 0.115.x, Apollo Server 4.x, OpenAPI 3.2.0, GraphQL September 2025). Covers API design patterns, authentication, pagination, error handling, gateway architecture, and monitoring. Activates for API architecture, REST/GraphQL design, API gateway configuration, and comprehensive API lifecycle management."
allowed-tools: "Read, Bash, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "api-expert"
secondary-agents: [alfred, qa-validator, doc-syncer]
keywords: [domain, web, api, rest, graphql, fastapi, apollo, openapi, authentication, pagination]
tags: [domain-expert, web-api, rest, graphql]
orchestration:
  can_resume: true
  typical_chain_position: "middle"
  depends_on: []
---

# moai-domain-web-api

**Enterprise Web API Design & Implementation**

> **Primary Agent**: api-expert  
> **Secondary Agents**: alfred, qa-validator, doc-syncer  
> **Version**: 4.0.0 (November 2025 Stable Releases)
> **Keywords**: domain, web, api, rest, graphql, fastapi, apollo, openapi, authentication, pagination

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

#### 🚀 Stable Web API Stack (November 2025)

**REST Frameworks (Stable)**:
```
FastAPI 0.115.13 (Python, Pydantic v2)
├─ Async/await native support
├─ OpenAPI 3.2.0 automatic generation
├─ Pydantic v2 validation
└─ High performance (Starlette/Uvicorn)

Express.js 4.21.2 (Node.js LTS)
├─ Mature ecosystem
├─ Middleware architecture
├─ Stable LTS support
└─ Wide adoption

Django REST Framework 3.15.x (Python)
├─ Full-featured
├─ Admin interface
├─ ORM integration
└─ Enterprise-ready
```

**GraphQL Frameworks (Stable)**:
```
Apollo Server 4.x (Node.js)
├─ GraphQL September 2025 spec
├─ Type-safe resolvers
├─ Schema federation
└─ Performance monitoring

Strawberry GraphQL 0.243.x (Python)
├─ Type hints based
├─ FastAPI integration
├─ DataLoader support
└─ Modern Python patterns
```

**API Specifications (Latest Stable)**:
```
OpenAPI 3.2.0 (September 2025)
├─ JSON Schema 2020-12 compatibility
├─ Webhook support
├─ Enhanced security schemes
└─ Better documentation

GraphQL Specification (September 2025)
├─ Operation descriptions
├─ Enhanced type system
├─ Improved error handling
└─ Better introspection
```

---

#### Pattern 1: FastAPI 0.115.x RESTful CRUD

```python
# FastAPI 0.115.13 with Pydantic v2 and async/await
from fastapi import FastAPI, HTTPException, Query, Depends
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field, ConfigDict
from typing import List, Optional
from datetime import datetime
import asyncio

app = FastAPI(
    title="Modern REST API",
    description="FastAPI 0.115.x with OpenAPI 3.2.0",
    version="1.0.0",
    openapi_version="3.2.0"  # Latest stable
)

# Pydantic v2 models
class ProductBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    description: Optional[str] = Field(None, max_length=500)
    price: float = Field(..., gt=0)
    category_id: int = Field(..., gt=0)
    
    model_config = ConfigDict(
        json_schema_extra={
            "example": {
                "name": "Product Name",
                "description": "Product description",
                "price": 29.99,
                "category_id": 1
            }
        }
    )

class ProductCreate(ProductBase):
    pass

class Product(ProductBase):
    id: int
    created_at: datetime
    updated_at: datetime
    
    model_config = ConfigDict(from_attributes=True)

# Dependency injection for database
async def get_db():
    # Simulated async database connection
    db = {"connected": True}
    try:
        yield db
    finally:
        # Cleanup
        pass

# RESTful CRUD endpoints
@app.post(
    "/api/v1/products",
    response_model=Product,
    status_code=201,
    tags=["products"],
    summary="Create a new product"
)
async def create_product(
    product: ProductCreate,
    db = Depends(get_db)
) -> Product:
    """
    Create a new product with validation:
    
    - **name**: Product name (1-100 characters)
    - **description**: Optional description (max 500 characters)
    - **price**: Product price (must be > 0)
    - **category_id**: Category ID (must be > 0)
    """
    # Simulate async database operation
    created_product = Product(
        id=1,
        **product.model_dump(),
        created_at=datetime.now(),
        updated_at=datetime.now()
    )
    return created_product

@app.get(
    "/api/v1/products",
    response_model=List[Product],
    tags=["products"],
    summary="List products with pagination"
)
async def list_products(
    skip: int = Query(0, ge=0, description="Number of records to skip"),
    limit: int = Query(50, ge=1, le=100, description="Maximum records to return"),
    category_id: Optional[int] = Query(None, gt=0, description="Filter by category"),
    db = Depends(get_db)
) -> List[Product]:
    """
    Retrieve products with pagination and filtering.
    
    Supports:
    - **Offset-based pagination** (skip/limit)
    - **Category filtering**
    - **Max 100 items per request**
    """
    # Simulate async database query
    products = [
        Product(
            id=i,
            name=f"Product {i}",
            description=f"Description for product {i}",
            price=float(i * 10),
            category_id=category_id or 1,
            created_at=datetime.now(),
            updated_at=datetime.now()
        )
        for i in range(skip + 1, skip + limit + 1)
    ]
    return products

@app.get(
    "/api/v1/products/{product_id}",
    response_model=Product,
    tags=["products"],
    summary="Get product by ID"
)
async def get_product(
    product_id: int,
    db = Depends(get_db)
) -> Product:
    """Retrieve a single product by ID."""
    # Simulate database lookup
    if product_id <= 0:
        raise HTTPException(
            status_code=404,
            detail=f"Product with ID {product_id} not found"
        )
    
    return Product(
        id=product_id,
        name=f"Product {product_id}",
        description="Sample product",
        price=99.99,
        category_id=1,
        created_at=datetime.now(),
        updated_at=datetime.now()
    )

@app.put(
    "/api/v1/products/{product_id}",
    response_model=Product,
    tags=["products"]
)
async def update_product(
    product_id: int,
    product: ProductCreate,
    db = Depends(get_db)
) -> Product:
    """Update an existing product."""
    return Product(
        id=product_id,
        **product.model_dump(),
        created_at=datetime.now(),
        updated_at=datetime.now()
    )

@app.delete(
    "/api/v1/products/{product_id}",
    status_code=204,
    tags=["products"]
)
async def delete_product(
    product_id: int,
    db = Depends(get_db)
):
    """Delete a product."""
    # Simulate deletion
    return None

# Error handling with RFC 7807 Problem Details
@app.exception_handler(HTTPException)
async def http_exception_handler(request, exc: HTTPException):
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "type": f"https://example.com/errors/{exc.status_code}",
            "title": "HTTP Error",
            "status": exc.status_code,
            "detail": exc.detail,
            "instance": str(request.url)
        }
    )
```

**Key Features (FastAPI 0.115.x)**:
- ✅ Pydantic v2 with `model_config` and `ConfigDict`
- ✅ OpenAPI 3.2.0 automatic generation
- ✅ Async/await throughout
- ✅ Dependency injection with `Depends()`
- ✅ Type hints and validation
- ✅ RFC 7807 Problem Details error format

---

#### Pattern 2: Apollo Server 4.x GraphQL API

```typescript
// Apollo Server 4.x with GraphQL September 2025 spec
import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { GraphQLError } from 'graphql';

// GraphQL Schema with September 2025 features
const typeDefs = `#graphql
  """
  Product type representing a sellable item
  """
  type Product {
    id: ID!
    name: String!
    description: String
    price: Float!
    category: Category!
    createdAt: String!
    updatedAt: String!
  }

  """
  Category type for product classification
  """
  type Category {
    id: ID!
    name: String!
    products: [Product!]!
  }

  """
  Pagination metadata
  """
  type PageInfo {
    hasNextPage: Boolean!
    hasPreviousPage: Boolean!
    startCursor: String
    endCursor: String
    totalCount: Int!
  }

  """
  Paginated product results
  """
  type ProductConnection {
    edges: [ProductEdge!]!
    pageInfo: PageInfo!
  }

  type ProductEdge {
    node: Product!
    cursor: String!
  }

  """
  Input for creating a new product
  """
  input CreateProductInput {
    name: String!
    description: String
    price: Float!
    categoryId: ID!
  }

  """
  Standard mutation response
  """
  interface MutationResponse {
    code: String!
    success: Boolean!
    message: String!
  }

  """
  Create product mutation response
  """
  type CreateProductResponse implements MutationResponse {
    code: String!
    success: Boolean!
    message: String!
    product: Product
  }

  type Query {
    """
    Get products with cursor-based pagination
    """
    products(
      first: Int = 50
      after: String
      categoryId: ID
    ): ProductConnection!

    """
    Get a single product by ID
    """
    product(id: ID!): Product
  }

  type Mutation {
    """
    Create a new product
    """
    createProduct(input: CreateProductInput!): CreateProductResponse!
  }
`;

// Type-safe resolvers
interface Product {
  id: string;
  name: string;
  description?: string;
  price: number;
  categoryId: string;
  createdAt: string;
  updatedAt: string;
}

interface Category {
  id: string;
  name: string;
}

interface Context {
  dataSources: {
    products: Product[];
    categories: Category[];
  };
}

const resolvers = {
  Query: {
    products: async (
      _parent: unknown,
      args: { first?: number; after?: string; categoryId?: string },
      context: Context
    ) => {
      const { first = 50, after, categoryId } = args;
      
      let filteredProducts = context.dataSources.products;
      
      // Filter by category
      if (categoryId) {
        filteredProducts = filteredProducts.filter(
          p => p.categoryId === categoryId
        );
      }
      
      // Cursor-based pagination
      const startIndex = after 
        ? filteredProducts.findIndex(p => p.id === after) + 1 
        : 0;
      
      const paginatedProducts = filteredProducts.slice(
        startIndex,
        startIndex + first
      );
      
      const edges = paginatedProducts.map(product => ({
        node: product,
        cursor: product.id
      }));
      
      return {
        edges,
        pageInfo: {
          hasNextPage: startIndex + first < filteredProducts.length,
          hasPreviousPage: startIndex > 0,
          startCursor: edges[0]?.cursor,
          endCursor: edges[edges.length - 1]?.cursor,
          totalCount: filteredProducts.length
        }
      };
    },

    product: async (
      _parent: unknown,
      args: { id: string },
      context: Context
    ) => {
      const product = context.dataSources.products.find(
        p => p.id === args.id
      );
      
      if (!product) {
        throw new GraphQLError('Product not found', {
          extensions: {
            code: 'NOT_FOUND',
            http: { status: 404 }
          }
        });
      }
      
      return product;
    }
  },

  Mutation: {
    createProduct: async (
      _parent: unknown,
      args: { input: Omit<Product, 'id' | 'createdAt' | 'updatedAt'> },
      context: Context
    ) => {
      // Validate price
      if (args.input.price <= 0) {
        throw new GraphQLError('Price must be greater than 0', {
          extensions: {
            code: 'BAD_USER_INPUT',
            argumentName: 'price'
          }
        });
      }
      
      const newProduct: Product = {
        id: String(context.dataSources.products.length + 1),
        ...args.input,
        categoryId: args.input.categoryId,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      };
      
      context.dataSources.products.push(newProduct);
      
      return {
        code: 'SUCCESS',
        success: true,
        message: 'Product created successfully',
        product: newProduct
      };
    }
  },

  Product: {
    category: (
      parent: Product,
      _args: unknown,
      context: Context
    ) => {
      return context.dataSources.categories.find(
        c => c.id === parent.categoryId
      );
    }
  },

  Category: {
    products: (
      parent: Category,
      _args: unknown,
      context: Context
    ) => {
      return context.dataSources.products.filter(
        p => p.categoryId === parent.id
      );
    }
  }
};

// Initialize Apollo Server 4.x
const server = new ApolloServer<Context>({
  typeDefs,
  resolvers,
  formatError: (formattedError, error) => {
    // Custom error formatting
    return {
      ...formattedError,
      extensions: {
        ...formattedError.extensions,
        timestamp: new Date().toISOString()
      }
    };
  }
});

// Start server
const { url } = await startStandaloneServer(server, {
  context: async () => ({
    dataSources: {
      products: [],
      categories: []
    }
  }),
  listen: { port: 4000 }
});

console.log(`🚀 Server ready at ${url}`);
```

**Key Features (Apollo Server 4.x)**:
- ✅ GraphQL September 2025 spec compliance
- ✅ Type-safe resolvers with TypeScript
- ✅ Cursor-based pagination (GraphQL best practice)
- ✅ MutationResponse interface pattern
- ✅ Custom error formatting
- ✅ Context typing

---

### Level 2: Practical Implementation (Common Patterns)

#### Pattern 3: OpenAPI 3.2.0 Schema Generation

```python
# FastAPI generates OpenAPI 3.2.0 automatically
from fastapi import FastAPI
from fastapi.openapi.utils import get_openapi

app = FastAPI()

def custom_openapi():
    if app.openapi_schema:
        return app.openapi_schema
    
    openapi_schema = get_openapi(
        title="My API",
        version="1.0.0",
        description="API with OpenAPI 3.2.0",
        routes=app.routes,
        # OpenAPI 3.2.0 features
        webhooks={
            "newProduct": {
                "post": {
                    "summary": "New product webhook",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {"$ref": "#/components/schemas/Product"}
                            }
                        }
                    },
                    "responses": {
                        "200": {"description": "Webhook received"}
                    }
                }
            }
        }
    )
    
    # Add security schemes
    openapi_schema["components"]["securitySchemes"] = {
        "BearerAuth": {
            "type": "http",
            "scheme": "bearer",
            "bearerFormat": "JWT"
        },
        "OAuth2": {
            "type": "oauth2",
            "flows": {
                "authorizationCode": {
                    "authorizationUrl": "https://example.com/oauth/authorize",
                    "tokenUrl": "https://example.com/oauth/token",
                    "scopes": {
                        "read": "Read access",
                        "write": "Write access"
                    }
                }
            }
        }
    }
    
    app.openapi_schema = openapi_schema
    return app.openapi_schema

app.openapi = custom_openapi
```

---

#### Pattern 4: API Versioning Strategies

```python
# Strategy 1: URI Versioning (Most common)
from fastapi import FastAPI, APIRouter

app = FastAPI()

# Version 1 router
v1_router = APIRouter(prefix="/api/v1", tags=["v1"])

@v1_router.get("/products")
async def get_products_v1():
    return {"version": "1.0", "products": []}

# Version 2 router with breaking changes
v2_router = APIRouter(prefix="/api/v2", tags=["v2"])

@v2_router.get("/products")
async def get_products_v2():
    return {"version": "2.0", "items": [], "pagination": {}}

app.include_router(v1_router)
app.include_router(v2_router)

# Strategy 2: Header Versioning
from fastapi import Header, HTTPException

@app.get("/products")
async def get_products(
    api_version: str = Header(default="1.0", alias="API-Version")
):
    if api_version == "1.0":
        return {"version": "1.0", "products": []}
    elif api_version == "2.0":
        return {"version": "2.0", "items": []}
    else:
        raise HTTPException(
            status_code=400,
            detail=f"Unsupported API version: {api_version}"
        )

# Strategy 3: Content Negotiation
from fastapi.responses import JSONResponse

@app.get("/products")
async def get_products(accept: str = Header(default="application/json")):
    if accept == "application/vnd.api.v1+json":
        return JSONResponse({"version": "1.0", "products": []})
    elif accept == "application/vnd.api.v2+json":
        return JSONResponse({"version": "2.0", "items": []})
    else:
        return JSONResponse({"version": "1.0", "products": []})
```

**Versioning Best Practices**:
- ✅ **URI versioning** (e.g., `/api/v1/`) - Easiest for clients
- ✅ **Header versioning** - Cleaner URLs, more complex
- ✅ **Content negotiation** - RESTful, but harder to test
- ✅ **Deprecation policy** - Announce 6-12 months in advance
- ✅ **Version sunset** - Remove old versions gradually

---

#### Pattern 5: Cursor-Based Pagination

```python
# Cursor-based pagination (better for large datasets)
from typing import Optional, List
from pydantic import BaseModel
from fastapi import Query
import base64

class PageInfo(BaseModel):
    has_next_page: bool
    has_previous_page: bool
    start_cursor: Optional[str] = None
    end_cursor: Optional[str] = None
    total_count: int

class ProductEdge(BaseModel):
    node: Product
    cursor: str

class ProductConnection(BaseModel):
    edges: List[ProductEdge]
    page_info: PageInfo

def encode_cursor(product_id: int) -> str:
    """Encode product ID as opaque cursor."""
    return base64.b64encode(f"product:{product_id}".encode()).decode()

def decode_cursor(cursor: str) -> int:
    """Decode cursor to product ID."""
    decoded = base64.b64decode(cursor.encode()).decode()
    return int(decoded.split(':')[1])

@app.get("/api/v1/products/cursor", response_model=ProductConnection)
async def get_products_cursor(
    first: int = Query(50, ge=1, le=100),
    after: Optional[str] = Query(None)
):
    """
    Cursor-based pagination (GraphQL Relay spec).
    
    Benefits:
    - Handles insertions/deletions gracefully
    - No page drift
    - Better for infinite scroll
    """
    # Simulate database query
    all_products = [...]  # Your product list
    
    # Find starting position
    start_index = 0
    if after:
        try:
            last_id = decode_cursor(after)
            start_index = next(
                i for i, p in enumerate(all_products) 
                if p.id > last_id
            )
        except (ValueError, StopIteration):
            pass
    
    # Get page of products
    page_products = all_products[start_index:start_index + first]
    
    # Build edges with cursors
    edges = [
        ProductEdge(
            node=product,
            cursor=encode_cursor(product.id)
        )
        for product in page_products
    ]
    
    # Build page info
    page_info = PageInfo(
        has_next_page=start_index + first < len(all_products),
        has_previous_page=start_index > 0,
        start_cursor=edges[0].cursor if edges else None,
        end_cursor=edges[-1].cursor if edges else None,
        total_count=len(all_products)
    )
    
    return ProductConnection(edges=edges, page_info=page_info)
```

**Pagination Comparison**:

| Type | Best For | Pros | Cons |
|------|----------|------|------|
| **Offset/Limit** | Small datasets | Simple, direct page access | Page drift, inefficient for large datasets |
| **Cursor-based** | Large datasets, real-time | Stable, no drift | Cannot jump to arbitrary page |
| **Keyset** | Time-series data | Efficient, stable | Requires indexed column |

---

#### Pattern 6: Rate Limiting (Token Bucket)

```python
# Rate limiting with Redis (token bucket algorithm)
from fastapi import FastAPI, Request, HTTPException
from datetime import datetime, timedelta
import redis.asyncio as redis
import json

app = FastAPI()

class TokenBucketRateLimiter:
    def __init__(
        self,
        redis_client: redis.Redis,
        capacity: int = 100,
        refill_rate: int = 10  # tokens per second
    ):
        self.redis = redis_client
        self.capacity = capacity
        self.refill_rate = refill_rate
    
    async def allow_request(self, key: str) -> tuple[bool, dict]:
        """
        Check if request is allowed under rate limit.
        
        Returns:
            (allowed, info) where info contains:
            - remaining: tokens remaining
            - reset_at: when bucket refills
            - retry_after: seconds to wait (if blocked)
        """
        now = datetime.now().timestamp()
        bucket_key = f"rate_limit:{key}"
        
        # Get current bucket state
        bucket_data = await self.redis.get(bucket_key)
        
        if bucket_data:
            bucket = json.loads(bucket_data)
            tokens = bucket['tokens']
            last_refill = bucket['last_refill']
        else:
            tokens = self.capacity
            last_refill = now
        
        # Refill tokens based on time elapsed
        time_elapsed = now - last_refill
        tokens_to_add = time_elapsed * self.refill_rate
        tokens = min(self.capacity, tokens + tokens_to_add)
        
        # Check if request allowed
        if tokens >= 1:
            tokens -= 1
            allowed = True
            retry_after = 0
        else:
            allowed = False
            retry_after = int((1 - tokens) / self.refill_rate)
        
        # Save bucket state
        bucket = {
            'tokens': tokens,
            'last_refill': now
        }
        await self.redis.setex(
            bucket_key,
            timedelta(seconds=3600),  # 1 hour TTL
            json.dumps(bucket)
        )
        
        info = {
            'remaining': int(tokens),
            'reset_at': int(now + (self.capacity - tokens) / self.refill_rate),
            'retry_after': retry_after
        }
        
        return allowed, info

# Initialize Redis
redis_client = redis.from_url("redis://localhost")
rate_limiter = TokenBucketRateLimiter(redis_client)

# Middleware for rate limiting
@app.middleware("http")
async def rate_limit_middleware(request: Request, call_next):
    # Use client IP as rate limit key
    client_ip = request.client.host
    
    allowed, info = await rate_limiter.allow_request(client_ip)
    
    if not allowed:
        raise HTTPException(
            status_code=429,
            detail="Rate limit exceeded",
            headers={
                "X-RateLimit-Limit": str(rate_limiter.capacity),
                "X-RateLimit-Remaining": str(info['remaining']),
                "X-RateLimit-Reset": str(info['reset_at']),
                "Retry-After": str(info['retry_after'])
            }
        )
    
    # Add rate limit headers to response
    response = await call_next(request)
    response.headers["X-RateLimit-Limit"] = str(rate_limiter.capacity)
    response.headers["X-RateLimit-Remaining"] = str(info['remaining'])
    response.headers["X-RateLimit-Reset"] = str(info['reset_at'])
    
    return response
```

**Rate Limiting Algorithms**:

| Algorithm | Use Case | Characteristics |
|-----------|----------|-----------------|
| **Token Bucket** | Smooth traffic, burst allowed | Allows bursts up to capacity |
| **Leaky Bucket** | Strict rate, no bursts | Constant output rate |
| **Fixed Window** | Simple, low overhead | Can allow 2x rate at boundary |
| **Sliding Window** | Accurate, no boundary issue | More complex, higher memory |

---

#### Pattern 7: JWT Authentication

```python
# JWT authentication with FastAPI
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from datetime import datetime, timedelta
import jwt
from typing import Optional

# Configuration
SECRET_KEY = "your-secret-key-keep-it-secret"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

security = HTTPBearer()

class User(BaseModel):
    username: str
    email: str
    full_name: Optional[str] = None
    disabled: bool = False

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None):
    """Create JWT access token."""
    to_encode = data.copy()
    
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=15)
    
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    
    return encoded_jwt

async def get_current_user(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> User:
    """
    Validate JWT token and return current user.
    
    Raises HTTPException if:
    - Token is missing
    - Token is expired
    - Token is invalid
    - User not found
    """
    token = credentials.credentials
    
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        
        if username is None:
            raise credentials_exception
            
    except jwt.ExpiredSignatureError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token has expired",
            headers={"WWW-Authenticate": "Bearer"},
        )
    except jwt.JWTError:
        raise credentials_exception
    
    # Fetch user from database
    user = get_user_from_db(username)
    if user is None:
        raise credentials_exception
    
    return user

async def get_current_active_user(
    current_user: User = Depends(get_current_user)
) -> User:
    """Check if user is active."""
    if current_user.disabled:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Inactive user"
        )
    return current_user

# Login endpoint
@app.post("/token")
async def login(username: str, password: str):
    """
    OAuth2 compatible token login.
    Returns JWT access token.
    """
    user = authenticate_user(username, password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Bearer"},
        )
    
    access_token_expires = timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
    access_token = create_access_token(
        data={"sub": user.username},
        expires_delta=access_token_expires
    )
    
    return {
        "access_token": access_token,
        "token_type": "bearer",
        "expires_in": ACCESS_TOKEN_EXPIRE_MINUTES * 60
    }

# Protected endpoint
@app.get("/users/me", response_model=User)
async def read_users_me(
    current_user: User = Depends(get_current_active_user)
):
    """Get current authenticated user."""
    return current_user
```

**JWT Best Practices**:
- ✅ Use strong secret key (256-bit minimum)
- ✅ Set reasonable expiration (15-30 minutes for access tokens)
- ✅ Use refresh tokens for long sessions
- ✅ Store tokens securely (HttpOnly cookies preferred)
- ✅ Implement token revocation (blacklist or database)
- ✅ Use RS256 for distributed systems

---

#### Pattern 8: Error Handling (RFC 7807)

```python
# RFC 7807 Problem Details for HTTP APIs
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from starlette.exceptions import HTTPException as StarletteHTTPException
from typing import Any, Dict, Optional

class ProblemDetail(BaseModel):
    """RFC 7807 Problem Details model."""
    type: str
    title: str
    status: int
    detail: str
    instance: str
    extensions: Optional[Dict[str, Any]] = None

@app.exception_handler(StarletteHTTPException)
async def http_exception_handler(
    request: Request,
    exc: StarletteHTTPException
) -> JSONResponse:
    """
    Convert HTTP exceptions to RFC 7807 format.
    
    Example response:
    {
        "type": "https://example.com/probs/not-found",
        "title": "Resource Not Found",
        "status": 404,
        "detail": "Product with ID 123 not found",
        "instance": "/api/v1/products/123"
    }
    """
    problem = ProblemDetail(
        type=f"https://example.com/errors/{exc.status_code}",
        title=get_title_for_status(exc.status_code),
        status=exc.status_code,
        detail=str(exc.detail),
        instance=str(request.url.path)
    )
    
    return JSONResponse(
        status_code=exc.status_code,
        content=problem.model_dump(exclude_none=True),
        headers={"Content-Type": "application/problem+json"}
    )

@app.exception_handler(RequestValidationError)
async def validation_exception_handler(
    request: Request,
    exc: RequestValidationError
) -> JSONResponse:
    """
    Convert validation errors to RFC 7807 format.
    
    Example response:
    {
        "type": "https://example.com/probs/validation-error",
        "title": "Validation Error",
        "status": 422,
        "detail": "Request validation failed",
        "instance": "/api/v1/products",
        "errors": [
            {
                "loc": ["body", "price"],
                "msg": "ensure this value is greater than 0",
                "type": "value_error.number.not_gt"
            }
        ]
    }
    """
    problem = ProblemDetail(
        type="https://example.com/errors/validation-error",
        title="Validation Error",
        status=422,
        detail="Request validation failed",
        instance=str(request.url.path),
        extensions={"errors": exc.errors()}
    )
    
    return JSONResponse(
        status_code=422,
        content=problem.model_dump(exclude_none=True),
        headers={"Content-Type": "application/problem+json"}
    )

def get_title_for_status(status_code: int) -> str:
    """Get human-readable title for HTTP status code."""
    titles = {
        400: "Bad Request",
        401: "Unauthorized",
        403: "Forbidden",
        404: "Not Found",
        422: "Unprocessable Entity",
        429: "Too Many Requests",
        500: "Internal Server Error",
        503: "Service Unavailable"
    }
    return titles.get(status_code, "HTTP Error")
```

**Error Response Best Practices**:
- ✅ Use RFC 7807 Problem Details format
- ✅ Include `type` URL pointing to error documentation
- ✅ Provide clear, actionable `detail` messages
- ✅ Set correct HTTP status codes
- ✅ Include `instance` for request traceability
- ✅ Add custom fields in `extensions`

---

#### Pattern 9: API Caching Strategies

```python
# Multi-layer caching with Redis and HTTP headers
from fastapi import FastAPI, Response, Request
from fastapi_cache import FastAPICache
from fastapi_cache.backends.redis import RedisBackend
from fastapi_cache.decorator import cache
import redis.asyncio as redis
from datetime import timedelta
from typing import Optional

app = FastAPI()

# Initialize Redis cache
@app.on_event("startup")
async def startup():
    redis_client = redis.from_url("redis://localhost")
    FastAPICache.init(RedisBackend(redis_client), prefix="fastapi-cache")

# Pattern 1: Response caching with TTL
@app.get("/products/{product_id}")
@cache(expire=300)  # Cache for 5 minutes
async def get_product(product_id: int, response: Response):
    """
    Cached endpoint with HTTP cache headers.
    
    Caching layers:
    1. Redis cache (5 min TTL)
    2. HTTP Cache-Control headers
    3. ETag for validation
    """
    product = fetch_product_from_db(product_id)
    
    # Set HTTP cache headers
    response.headers["Cache-Control"] = "public, max-age=300"
    response.headers["ETag"] = generate_etag(product)
    
    return product

# Pattern 2: Conditional requests (ETag validation)
@app.get("/products/{product_id}/conditional")
async def get_product_conditional(
    product_id: int,
    request: Request,
    response: Response
):
    """
    Support conditional requests with ETag.
    Returns 304 Not Modified if content hasn't changed.
    """
    product = fetch_product_from_db(product_id)
    current_etag = generate_etag(product)
    
    # Check If-None-Match header
    if_none_match = request.headers.get("If-None-Match")
    
    if if_none_match == current_etag:
        response.status_code = 304  # Not Modified
        return Response(status_code=304)
    
    # Set cache headers
    response.headers["Cache-Control"] = "public, max-age=300"
    response.headers["ETag"] = current_etag
    response.headers["Last-Modified"] = product.updated_at.strftime(
        "%a, %d %b %Y %H:%M:%S GMT"
    )
    
    return product

# Pattern 3: Cache invalidation
@app.put("/products/{product_id}")
async def update_product(product_id: int, product: ProductUpdate):
    """
    Update product and invalidate cache.
    """
    updated_product = update_product_in_db(product_id, product)
    
    # Invalidate cache
    cache_key = f"fastapi-cache:get_product:{product_id}"
    await FastAPICache.clear(namespace=cache_key)
    
    return updated_product

# Pattern 4: Cache-aside pattern with fallback
async def get_product_cached(product_id: int) -> Product:
    """
    Cache-aside pattern:
    1. Try cache first
    2. On miss, fetch from database
    3. Store in cache
    """
    cache_key = f"product:{product_id}"
    
    # Try cache
    cached = await redis_client.get(cache_key)
    if cached:
        return Product.parse_raw(cached)
    
    # Cache miss - fetch from database
    product = fetch_product_from_db(product_id)
    
    # Store in cache
    await redis_client.setex(
        cache_key,
        timedelta(minutes=5),
        product.json()
    )
    
    return product

def generate_etag(content: Any) -> str:
    """Generate ETag from content hash."""
    import hashlib
    content_str = str(content)
    return f'"{hashlib.md5(content_str.encode()).hexdigest()}"'
```

**Caching Strategy Comparison**:

| Strategy | When to Use | TTL | Invalidation |
|----------|-------------|-----|--------------|
| **HTTP Cache-Control** | Public, static content | Long (hours-days) | Time-based |
| **ETag/Conditional** | Frequently changing | Short (minutes) | Content-based |
| **Redis cache** | Computed results | Medium (minutes-hours) | Explicit |
| **CDN** | Global distribution | Long (days-weeks) | Purge API |

---

#### Pattern 10: API Gateway (Kong 3.9.x)

```yaml
# Kong Gateway 3.9.x configuration with declarative format
_format_version: "3.0"
_transform: true

services:
  - name: product-service
    url: http://product-api:8000
    protocol: http
    connect_timeout: 60000
    write_timeout: 60000
    read_timeout: 60000
    
    routes:
      - name: product-route
        paths:
          - /api/products
        methods:
          - GET
          - POST
          - PUT
          - DELETE
        strip_path: false
        preserve_host: false
    
    plugins:
      # Rate limiting (Kong 3.9.x)
      - name: rate-limiting
        config:
          minute: 100
          hour: 5000
          policy: redis
          redis_host: redis
          redis_port: 6379
          fault_tolerant: true
      
      # JWT authentication
      - name: jwt
        config:
          claims_to_verify:
            - exp
          key_claim_name: iss
          secret_is_base64: false
      
      # CORS
      - name: cors
        config:
          origins:
            - "*"
          methods:
            - GET
            - POST
            - PUT
            - DELETE
            - PATCH
          headers:
            - Accept
            - Authorization
            - Content-Type
          exposed_headers:
            - X-RateLimit-Limit
            - X-RateLimit-Remaining
          credentials: true
          max_age: 3600
      
      # Request/Response transformation
      - name: request-transformer
        config:
          add:
            headers:
              - X-Gateway-Version:3.9
              - X-Request-ID:$(uuid)
      
      # Prometheus metrics
      - name: prometheus
        config:
          per_consumer: true
      
      # IP restriction
      - name: ip-restriction
        config:
          allow:
            - 10.0.0.0/8
            - 172.16.0.0/12
            - 192.168.0.0/16
      
      # Circuit breaker
      - name: request-termination
        config:
          status_code: 503
          message: "Service temporarily unavailable"
        enabled: false  # Enable during incidents

consumers:
  - username: api-client-1
    custom_id: client-1
    
    jwt_secrets:
      - key: issuer-key
        secret: your-256-bit-secret
        algorithm: HS256
    
    plugins:
      - name: rate-limiting
        config:
          minute: 200  # Higher limit for premium clients
          hour: 10000
```

**Kong 3.9.x Key Features**:
- ✅ Declarative configuration (GitOps-ready)
- ✅ Incremental sync (reduced memory/CPU)
- ✅ Built-in observability (Prometheus, OpenTelemetry)
- ✅ Flexible rate limiting (per-consumer, per-route)
- ✅ JWT/OAuth2 authentication
- ✅ Request/response transformation
- ✅ Circuit breaker support

---

### Level 3: Advanced Patterns (Expert Reference)

#### Pattern 11: GraphQL Subscriptions (Real-time)

```typescript
// GraphQL Subscriptions with Apollo Server 4.x and PubSub
import { ApolloServer } from '@apollo/server';
import { expressMiddleware } from '@apollo/server/express4';
import { ApolloServerPluginDrainHttpServer } from '@apollo/server/plugin/drainHttpServer';
import { createServer } from 'http';
import { WebSocketServer } from 'ws';
import { useServer } from 'graphql-ws/lib/use/ws';
import { makeExecutableSchema } from '@graphql-tools/schema';
import { PubSub } from 'graphql-subscriptions';
import express from 'express';

const pubsub = new PubSub();

const typeDefs = `#graphql
  type Product {
    id: ID!
    name: String!
    price: Float!
  }

  type Query {
    products: [Product!]!
  }

  type Mutation {
    createProduct(name: String!, price: Float!): Product!
  }

  type Subscription {
    productCreated: Product!
    productPriceChanged(productId: ID!): Product!
  }
`;

const resolvers = {
  Query: {
    products: () => {
      // Fetch products from database
      return [];
    }
  },

  Mutation: {
    createProduct: async (_: any, args: { name: string; price: number }) => {
      const product = {
        id: String(Date.now()),
        name: args.name,
        price: args.price
      };

      // Save to database
      // ...

      // Publish to subscribers
      await pubsub.publish('PRODUCT_CREATED', {
        productCreated: product
      });

      return product;
    }
  },

  Subscription: {
    productCreated: {
      subscribe: () => pubsub.asyncIterator(['PRODUCT_CREATED'])
    },
    
    productPriceChanged: {
      subscribe: (_: any, args: { productId: string }) => {
        return pubsub.asyncIterator([`PRICE_CHANGED_${args.productId}`]);
      }
    }
  }
};

// Create schema
const schema = makeExecutableSchema({ typeDefs, resolvers });

// Create Express app
const app = express();
const httpServer = createServer(app);

// WebSocket server for subscriptions
const wsServer = new WebSocketServer({
  server: httpServer,
  path: '/graphql'
});

const serverCleanup = useServer({ schema }, wsServer);

// Apollo Server
const server = new ApolloServer({
  schema,
  plugins: [
    ApolloServerPluginDrainHttpServer({ httpServer }),
    {
      async serverWillStart() {
        return {
          async drainServer() {
            await serverCleanup.dispose();
          }
        };
      }
    }
  ]
});

await server.start();

app.use(
  '/graphql',
  express.json(),
  expressMiddleware(server)
);

httpServer.listen(4000, () => {
  console.log(`🚀 Server ready at http://localhost:4000/graphql`);
  console.log(`🔌 Subscriptions ready at ws://localhost:4000/graphql`);
});
```

**Subscription Client Example**:

```typescript
// Client-side subscription
import { ApolloClient, InMemoryCache, split, HttpLink } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import { getMainDefinition } from '@apollo/client/utilities';
import { gql } from '@apollo/client';

// HTTP link for queries and mutations
const httpLink = new HttpLink({
  uri: 'http://localhost:4000/graphql'
});

// WebSocket link for subscriptions
const wsLink = new GraphQLWsLink(
  createClient({
    url: 'ws://localhost:4000/graphql'
  })
);

// Split based on operation type
const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return (
      definition.kind === 'OperationDefinition' &&
      definition.operation === 'subscription'
    );
  },
  wsLink,
  httpLink
);

const client = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache()
});

// Subscribe to product creation
const PRODUCT_CREATED_SUBSCRIPTION = gql`
  subscription OnProductCreated {
    productCreated {
      id
      name
      price
    }
  }
`;

const subscription = client.subscribe({
  query: PRODUCT_CREATED_SUBSCRIPTION
}).subscribe({
  next: ({ data }) => {
    console.log('New product:', data.productCreated);
  },
  error: (error) => {
    console.error('Subscription error:', error);
  }
});
```

---

#### Pattern 12: API Performance Monitoring

```python
# Prometheus metrics with FastAPI
from prometheus_client import Counter, Histogram, Gauge
from prometheus_fastapi_instrumentator import Instrumentator
from fastapi import FastAPI
import time

app = FastAPI()

# Custom metrics
REQUEST_COUNT = Counter(
    'api_requests_total',
    'Total API requests',
    ['method', 'endpoint', 'status']
)

REQUEST_LATENCY = Histogram(
    'api_request_duration_seconds',
    'API request latency',
    ['method', 'endpoint']
)

ACTIVE_REQUESTS = Gauge(
    'api_active_requests',
    'Number of active requests',
    ['method', 'endpoint']
)

# Instrumentator with custom metrics
instrumentator = Instrumentator(
    should_group_status_codes=False,
    should_ignore_untemplated=True,
    should_respect_env_var=True,
    should_instrument_requests_inprogress=True,
    excluded_handlers=["/metrics", "/health"],
    env_var_name="ENABLE_METRICS",
    inprogress_name="api_requests_inprogress",
    inprogress_labels=True
)

# Add custom middleware
@app.middleware("http")
async def metrics_middleware(request, call_next):
    method = request.method
    endpoint = request.url.path
    
    # Track active requests
    ACTIVE_REQUESTS.labels(method=method, endpoint=endpoint).inc()
    
    # Measure latency
    start_time = time.time()
    
    try:
        response = await call_next(request)
        status = response.status_code
    except Exception as e:
        status = 500
        raise
    finally:
        # Record metrics
        duration = time.time() - start_time
        
        REQUEST_COUNT.labels(
            method=method,
            endpoint=endpoint,
            status=status
        ).inc()
        
        REQUEST_LATENCY.labels(
            method=method,
            endpoint=endpoint
        ).observe(duration)
        
        ACTIVE_REQUESTS.labels(method=method, endpoint=endpoint).dec()
    
    return response

# Instrument app
instrumentator.instrument(app).expose(app)

# Health check endpoint
@app.get("/health")
async def health():
    return {"status": "healthy"}
```

**Grafana Dashboard Query Examples**:

```promql
# Request rate by endpoint
rate(api_requests_total[5m])

# p95 latency by endpoint
histogram_quantile(0.95, 
  rate(api_request_duration_seconds_bucket[5m])
)

# Error rate
rate(api_requests_total{status=~"5.."}[5m])

# Active requests
api_active_requests

# Requests per second by status
sum(rate(api_requests_total[1m])) by (status)
```

---

## 🎯 Best Practices Checklist

### Must-Have

- ✅ **Use stable versions** (FastAPI 0.115.x, Apollo Server 4.x, OpenAPI 3.2.0)
- ✅ **Implement proper error handling** (RFC 7807 Problem Details)
- ✅ **Add authentication/authorization** (JWT, OAuth 2.1)
- ✅ **Use pagination** (cursor-based for large datasets)
- ✅ **Implement rate limiting** (prevent abuse)
- ✅ **Version your APIs** (URI versioning recommended)
- ✅ **Add monitoring** (Prometheus, OpenTelemetry)
- ✅ **Document with OpenAPI/GraphQL SDL**

### Recommended

- ✅ **Use API gateway** (Kong 3.9.x, Traefik 3.2.x)
- ✅ **Implement caching** (Redis, HTTP cache headers)
- ✅ **Add request validation** (Pydantic v2, GraphQL schema)
- ✅ **Use async/await** (non-blocking I/O)
- ✅ **Add CORS** (for web clients)
- ✅ **Implement health checks** (`/health`, `/readiness`)
- ✅ **Use structured logging** (JSON format)

### Security

- 🔒 **HTTPS only** (TLS 1.3 preferred)
- 🔒 **Validate all inputs** (never trust user input)
- 🔒 **Use parameterized queries** (prevent SQL injection)
- 🔒 **Implement CSRF protection** (for state-changing operations)
- 🔒 **Rate limit authentication endpoints** (prevent brute force)
- 🔒 **Use secure headers** (HSTS, CSP, X-Frame-Options)
- 🔒 **Audit sensitive operations** (log security events)

---

## 🔗 Context7 MCP Integration

**When to Use Context7**:

- Working with FastAPI, Apollo Server, GraphQL
- Need latest stable version documentation
- Verifying API design patterns
- Checking authentication/authorization best practices

**Relevant Libraries**:

| Library | Context7 ID | Stable Version | Use Case |
|---------|-------------|----------------|----------|
| FastAPI | `/fastapi/fastapi/0.115.13` | 0.115.13 | REST API development |
| Apollo Server | `/apollographql/apollo-server` | 4.x | GraphQL server |
| Strawberry GraphQL | `/kamilkisiela/strawberry-graphql` | 0.243.x | Python GraphQL |
| Pydantic | `/pydantic/pydantic` | v2 | Data validation |

**Example Usage**:

```python
# Fetch latest FastAPI documentation
from context7_helper import get_docs

docs = await get_docs(
    library_id="/fastapi/fastapi/0.115.13",
    topic="authentication JWT OAuth2",
    tokens=5000
)
```

---

## 📊 Decision Tree

**When to use moai-domain-web-api:**

```
Need to build web API?
  ├─ REST API?
  │   ├─ Python → Use FastAPI 0.115.x (Pattern 1)
  │   ├─ Node.js → Use Express 4.21.x
  │   └─ Full-featured → Django REST Framework 3.15.x
  │
  ├─ GraphQL API?
  │   ├─ Node.js → Apollo Server 4.x (Pattern 2)
  │   ├─ Python → Strawberry GraphQL 0.243.x
  │   └─ Real-time → Add Subscriptions (Pattern 11)
  │
  ├─ Need API Gateway?
  │   ├─ Enterprise → Kong Gateway 3.9.x (Pattern 10)
  │   ├─ Cloud-native → Traefik 3.2.x
  │   └─ Service mesh → Envoy Proxy 1.31.x
  │
  └─ Design patterns?
      ├─ Pagination → Cursor-based (Pattern 5)
      ├─ Auth → JWT (Pattern 7)
      ├─ Rate limit → Token bucket (Pattern 6)
      ├─ Errors → RFC 7807 (Pattern 8)
      └─ Caching → Multi-layer (Pattern 9)
```

---

## 🔄 Integration with Other Skills

**Prerequisite Skills**:
- Skill("moai-foundation-coding") – Python/TypeScript basics
- Skill("moai-domain-database") – Database integration
- Skill("moai-domain-security") – Authentication/authorization

**Complementary Skills**:
- Skill("moai-domain-backend") – Backend architecture
- Skill("moai-domain-frontend") – API consumption
- Skill("moai-alfred-git-workflow") – Version control

**Next Steps**:
- Skill("moai-domain-devops") – API deployment
- Skill("moai-alfred-testing") – API testing strategies

---

## 📚 Official References

### REST & OpenAPI

- [FastAPI Documentation](https://fastapi.tiangolo.com/) - Official FastAPI 0.115.x docs
- [OpenAPI 3.2.0 Specification](https://spec.openapis.org/oas/v3.2.0.html) - Latest stable spec
- [RFC 7807: Problem Details](https://www.rfc-editor.org/rfc/rfc7807) - Error response standard
- [RFC 7231: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc7231) - HTTP methods & status codes

### GraphQL

- [GraphQL September 2025 Spec](https://spec.graphql.org/September2025/) - Latest GraphQL spec
- [Apollo Server Documentation](https://www.apollographql.com/docs/apollo-server/) - Apollo Server 4.x
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/) - Official best practices
- [Relay Cursor Connections](https://relay.dev/graphql/connections.htm) - Pagination spec

### API Gateway

- [Kong Gateway 3.9.x Docs](https://docs.konghq.com/gateway/latest/) - Kong documentation
- [Traefik Documentation](https://doc.traefik.io/traefik/) - Traefik proxy
- [Envoy Proxy Docs](https://www.envoyproxy.io/docs) - Envoy documentation

### Authentication & Security

- [JWT.io](https://jwt.io/) - JSON Web Token resources
- [OAuth 2.1](https://oauth.net/2.1/) - OAuth 2.1 specification
- [OWASP API Security](https://owasp.org/www-project-api-security/) - API security top 10

---

## 📈 Version History

**v4.0.0** (2025-11-12)
- ✨ Updated to stable versions (FastAPI 0.115.x, Apollo Server 4.x, OpenAPI 3.2.0)
- ✨ 12 production-ready patterns with stable tech stack
- ✨ Context7 MCP integration for latest docs
- ✨ Progressive disclosure structure (3 levels)
- ✨ REST vs GraphQL comparison matrix
- ✨ Comprehensive best practices checklist
- ✨ Security-first approach
- ✨ 800+ lines of expert-level content

---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (api-expert)  
**Stable Stack**: FastAPI 0.115.13 | Apollo Server 4.x | OpenAPI 3.2.0 | GraphQL Sep 2025
