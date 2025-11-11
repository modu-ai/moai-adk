# moai-domain-database - Working Examples

_Last updated: 2025-11-12_

## Example 1: PostgreSQL 17 Full-Text Search with GIN Index

```sql
-- Create table with tsvector column for full-text search
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2),
    search_vector tsvector GENERATED ALWAYS AS (
        to_tsvector('english', name || ' ' || COALESCE(description, ''))
    ) STORED
);

-- Create GIN index for fast full-text search
CREATE INDEX idx_products_search ON products USING GIN (search_vector);

-- Insert sample data
INSERT INTO products (name, description, price) VALUES
('Laptop Pro 15', 'High-performance laptop with 16GB RAM', 1299.99),
('Wireless Mouse', 'Ergonomic wireless mouse with USB receiver', 29.99),
('Mechanical Keyboard', 'RGB mechanical keyboard for gamers', 89.99);

-- Full-text search query
SELECT id, name, description,
       ts_rank(search_vector, query) AS rank
FROM products,
     to_tsquery('english', 'laptop | keyboard') AS query
WHERE search_vector @@ query
ORDER BY rank DESC;
```

---

## Example 2: SQLAlchemy 2.0 Async CRUD with Connection Pooling

```python
# models.py
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship
from sqlalchemy import String, Integer, ForeignKey, DateTime
from datetime import datetime
from typing import List

class Base(DeclarativeBase):
    pass

class User(Base):
    __tablename__ = 'users'
    
    id: Mapped[int] = mapped_column(primary_key=True)
    email: Mapped[str] = mapped_column(String(255), unique=True, index=True)
    name: Mapped[str] = mapped_column(String(100))
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
    
    posts: Mapped[List["Post"]] = relationship(back_populates="author")

class Post(Base):
    __tablename__ = 'posts'
    
    id: Mapped[int] = mapped_column(primary_key=True)
    title: Mapped[str] = mapped_column(String(255))
    content: Mapped[str] = mapped_column(String)
    author_id: Mapped[int] = mapped_column(ForeignKey('users.id'), index=True)
    published: Mapped[bool] = mapped_column(default=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
    
    author: Mapped["User"] = relationship(back_populates="posts")

# database.py
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy import select
from typing import Optional

engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    pool_size=20,
    max_overflow=10,
    pool_pre_ping=True,
    echo=False
)

AsyncSessionLocal = sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False
)

# crud.py
from sqlalchemy.orm import selectinload

async def create_user(email: str, name: str) -> User:
    async with AsyncSessionLocal() as session:
        user = User(email=email, name=name)
        session.add(user)
        await session.commit()
        await session.refresh(user)
        return user

async def get_user_with_posts(user_id: int) -> Optional[User]:
    async with AsyncSessionLocal() as session:
        result = await session.execute(
            select(User)
            .options(selectinload(User.posts))
            .where(User.id == user_id)
        )
        return result.scalar_one_or_none()

async def update_user(user_id: int, **kwargs) -> Optional[User]:
    async with AsyncSessionLocal() as session:
        user = await session.get(User, user_id)
        if not user:
            return None
        for key, value in kwargs.items():
            setattr(user, key, value)
        await session.commit()
        await session.refresh(user)
        return user

async def delete_user(user_id: int) -> bool:
    async with AsyncSessionLocal() as session:
        user = await session.get(User, user_id)
        if not user:
            return False
        await session.delete(user)
        await session.commit()
        return True

# Usage
async def main():
    # Create
    user = await create_user("alice@example.com", "Alice")
    print(f"Created user: {user.id}")
    
    # Read with relations
    user_with_posts = await get_user_with_posts(user.id)
    print(f"User has {len(user_with_posts.posts)} posts")
    
    # Update
    updated = await update_user(user.id, name="Alice Smith")
    print(f"Updated user: {updated.name}")
    
    # Delete
    deleted = await delete_user(user.id)
    print(f"Deleted: {deleted}")
```

---

## Example 3: Prisma 5 E-Commerce Schema with Relations

```prisma
// schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id        Int      @id @default(autoincrement())
  email     String   @unique
  name      String
  createdAt DateTime @default(now())
  updatedAt DateTime @updatedAt

  orders    Order[]
  cart      Cart?

  @@index([email])
}

model Product {
  id          Int      @id @default(autoincrement())
  name        String
  description String?
  price       Decimal  @db.Decimal(10, 2)
  stock       Int      @default(0)
  categoryId  Int
  createdAt   DateTime @default(now())
  updatedAt   DateTime @updatedAt

  category    Category @relation(fields: [categoryId], references: [id])
  orderItems  OrderItem[]
  cartItems   CartItem[]

  @@index([categoryId])
  @@index([name])
}

model Category {
  id       Int      @id @default(autoincrement())
  name     String   @unique
  products Product[]
}

model Order {
  id        Int      @id @default(autoincrement())
  userId    Int
  status    String   @default("pending")
  total     Decimal  @db.Decimal(10, 2)
  createdAt DateTime @default(now())

  user      User       @relation(fields: [userId], references: [id])
  items     OrderItem[]

  @@index([userId, status])
}

model OrderItem {
  id        Int     @id @default(autoincrement())
  orderId   Int
  productId Int
  quantity  Int
  price     Decimal @db.Decimal(10, 2)

  order     Order   @relation(fields: [orderId], references: [id])
  product   Product @relation(fields: [productId], references: [id])

  @@index([orderId])
  @@index([productId])
}

model Cart {
  id     Int        @id @default(autoincrement())
  userId Int        @unique
  user   User       @relation(fields: [userId], references: [id])
  items  CartItem[]
}

model CartItem {
  id        Int     @id @default(autoincrement())
  cartId    Int
  productId Int
  quantity  Int

  cart      Cart    @relation(fields: [cartId], references: [id])
  product   Product @relation(fields: [productId], references: [id])

  @@unique([cartId, productId])
}
```

```typescript
// app.ts - Optimized queries
import { PrismaClient } from '@prisma/client'

const prisma = new PrismaClient()

// Get user with orders and items
async function getUserOrders(userId: number) {
  return await prisma.user.findUnique({
    where: { id: userId },
    include: {
      orders: {
        include: {
          items: {
            include: {
              product: {
                select: {
                  name: true,
                  price: true
                }
              }
            }
          }
        },
        orderBy: { createdAt: 'desc' },
        take: 10
      }
    }
  })
}

// Create order with transaction
async function createOrder(userId: number, items: Array<{productId: number, quantity: number}>) {
  return await prisma.$transaction(async (tx) => {
    // Calculate total
    const products = await tx.product.findMany({
      where: { id: { in: items.map(i => i.productId) } }
    })
    
    const total = items.reduce((sum, item) => {
      const product = products.find(p => p.id === item.productId)
      return sum + (product ? Number(product.price) * item.quantity : 0)
    }, 0)
    
    // Create order
    const order = await tx.order.create({
      data: {
        userId,
        total,
        status: 'pending',
        items: {
          create: items.map(item => {
            const product = products.find(p => p.id === item.productId)
            return {
              productId: item.productId,
              quantity: item.quantity,
              price: product?.price || 0
            }
          })
        }
      },
      include: {
        items: {
          include: {
            product: true
          }
        }
      }
    })
    
    // Update stock
    for (const item of items) {
      await tx.product.update({
        where: { id: item.productId },
        data: {
          stock: {
            decrement: item.quantity
          }
        }
      })
    }
    
    return order
  })
}
```

---

## Example 4: Redis 7.4 Distributed Caching Layer

```python
# cache.py - Production-ready Redis caching
import redis.asyncio as redis
import json
import hashlib
from typing import Any, Optional, Callable
from functools import wraps
import asyncio

class RedisCache:
    def __init__(self, redis_url: str = "redis://localhost:6379/0"):
        self.client = redis.from_url(
            redis_url,
            encoding="utf-8",
            decode_responses=True,
            max_connections=50,
            socket_keepalive=True,
            socket_connect_timeout=5,
            retry_on_timeout=True
        )
    
    async def get(self, key: str) -> Optional[Any]:
        """Get value from cache"""
        value = await self.client.get(key)
        return json.loads(value) if value else None
    
    async def set(self, key: str, value: Any, ttl: int = 3600):
        """Set value in cache with TTL"""
        await self.client.setex(key, ttl, json.dumps(value))
    
    async def delete(self, key: str):
        """Delete key from cache"""
        await self.client.delete(key)
    
    async def invalidate_pattern(self, pattern: str):
        """Invalidate all keys matching pattern"""
        keys = []
        async for key in self.client.scan_iter(match=pattern):
            keys.append(key)
        if keys:
            await self.client.delete(*keys)
    
    def cached(self, ttl: int = 3600, key_prefix: str = ""):
        """Decorator for caching function results"""
        def decorator(func: Callable):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                # Generate cache key from function name and args
                key_parts = [key_prefix or func.__name__]
                key_parts.extend(str(arg) for arg in args)
                key_parts.extend(f"{k}:{v}" for k, v in sorted(kwargs.items()))
                cache_key = hashlib.md5(":".join(key_parts).encode()).hexdigest()
                
                # Try cache first
                cached_value = await self.get(cache_key)
                if cached_value is not None:
                    return cached_value
                
                # Cache miss: execute function
                result = await func(*args, **kwargs)
                
                # Store in cache
                await self.set(cache_key, result, ttl)
                return result
            
            return wrapper
        return decorator

# Usage example
cache = RedisCache()

@cache.cached(ttl=300, key_prefix="user")
async def get_user_profile(user_id: int):
    # Expensive database query
    await asyncio.sleep(1)  # Simulate DB query
    return {
        "id": user_id,
        "name": f"User {user_id}",
        "email": f"user{user_id}@example.com"
    }

# Session storage example
async def store_session(session_id: str, data: dict):
    await cache.set(f"session:{session_id}", data, ttl=1800)  # 30 min

async def get_session(session_id: str):
    return await cache.get(f"session:{session_id}")

# Rate limiting example
async def rate_limit(user_id: int, max_requests: int = 100, window: int = 60):
    key = f"ratelimit:{user_id}"
    count = await cache.client.incr(key)
    if count == 1:
        await cache.client.expire(key, window)
    return count <= max_requests
```

---

## Example 5: MongoDB 8.0 Analytics Pipeline

```javascript
// analytics.js - Advanced aggregation pipeline
const { MongoClient } = require('mongodb')

const client = new MongoClient('mongodb://localhost:27017')

async function getUserEngagementStats(startDate, endDate) {
  const db = client.db('analytics')
  
  return await db.collection('events').aggregate([
    // Stage 1: Filter by date range
    {
      $match: {
        timestamp: {
          $gte: new Date(startDate),
          $lte: new Date(endDate)
        },
        eventType: { $in: ['page_view', 'click', 'purchase'] }
      }
    },
    
    // Stage 2: Group by user and event type
    {
      $group: {
        _id: {
          userId: '$userId',
          eventType: '$eventType'
        },
        count: { $sum: 1 },
        totalValue: { $sum: '$value' }
      }
    },
    
    // Stage 3: Reshape data
    {
      $group: {
        _id: '$_id.userId',
        events: {
          $push: {
            type: '$_id.eventType',
            count: '$count',
            value: '$totalValue'
          }
        },
        totalEvents: { $sum: '$count' }
      }
    },
    
    // Stage 4: Lookup user details
    {
      $lookup: {
        from: 'users',
        localField: '_id',
        foreignField: 'userId',
        as: 'userInfo'
      }
    },
    
    // Stage 5: Unwind user info
    {
      $unwind: {
        path: '$userInfo',
        preserveNullAndEmptyArrays: true
      }
    },
    
    // Stage 6: Calculate engagement score
    {
      $addFields: {
        engagementScore: {
          $add: [
            { $multiply: [
              { $size: { $filter: { 
                input: '$events',
                cond: { $eq: ['$$this.type', 'page_view'] }
              }}},
              1
            ]},
            { $multiply: [
              { $size: { $filter: {
                input: '$events',
                cond: { $eq: ['$$this.type', 'click'] }
              }}},
              5
            ]},
            { $multiply: [
              { $size: { $filter: {
                input: '$events',
                cond: { $eq: ['$$this.type', 'purchase'] }
              }}},
              10
            ]}
          ]
        }
      }
    },
    
    // Stage 7: Project final fields
    {
      $project: {
        userId: '$_id',
        userName: '$userInfo.name',
        userEmail: '$userInfo.email',
        events: 1,
        totalEvents: 1,
        engagementScore: 1
      }
    },
    
    // Stage 8: Sort by engagement score
    {
      $sort: { engagementScore: -1 }
    },
    
    // Stage 9: Limit results
    {
      $limit: 100
    }
  ]).toArray()
}

// Real-time analytics with $merge
async function updateDailyStats(date) {
  const db = client.db('analytics')
  
  await db.collection('events').aggregate([
    {
      $match: {
        timestamp: {
          $gte: new Date(date),
          $lt: new Date(date.getTime() + 86400000)
        }
      }
    },
    {
      $group: {
        _id: {
          date: { $dateToString: { format: '%Y-%m-%d', date: '$timestamp' } },
          userId: '$userId',
          eventType: '$eventType'
        },
        count: { $sum: 1 },
        totalValue: { $sum: '$value' }
      }
    },
    {
      $merge: {
        into: 'daily_stats',
        on: '_id',
        whenMatched: 'replace',
        whenNotMatched: 'insert'
      }
    }
  ])
}
```

---

## Example 6: Zero-Downtime Database Migration

```python
# migrations/versions/001_add_verified_column.py
"""Add verified column to users table

Revision ID: 001
Create Date: 2025-11-12
"""
from alembic import op
import sqlalchemy as sa

def upgrade():
    # Step 1: Add nullable column
    print("Step 1: Adding nullable verified column...")
    op.add_column('users',
        sa.Column('verified', sa.Boolean(), nullable=True)
    )
    
    # Step 2: Backfill existing rows in batches
    print("Step 2: Backfilling existing rows...")
    connection = op.get_bind()
    
    # Get total count
    result = connection.execute(
        sa.text("SELECT COUNT(*) FROM users WHERE verified IS NULL")
    )
    total = result.scalar()
    print(f"Total rows to backfill: {total}")
    
    # Backfill in batches of 1000
    batch_size = 1000
    for offset in range(0, total, batch_size):
        connection.execute(
            sa.text(
                "UPDATE users SET verified = false "
                "WHERE verified IS NULL "
                f"LIMIT {batch_size}"
            )
        )
        print(f"Backfilled {min(offset + batch_size, total)}/{total} rows")
    
    # Step 3: Make column NOT NULL
    print("Step 3: Making column NOT NULL...")
    op.alter_column('users', 'verified',
        existing_type=sa.Boolean(),
        nullable=False,
        server_default=sa.false()
    )
    
    print("Migration completed successfully!")

def downgrade():
    op.drop_column('users', 'verified')
```

---

_For more examples, see SKILL.md Level 2 patterns_
