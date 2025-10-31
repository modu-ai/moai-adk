---
name: moai-domain-database
type: domain
description: Database design, schema optimization, and migration management
tier: domain
---

# Database Design

## Quick Start (30 seconds)
Database design impacts application performance, data integrity, and scalability. Master schema design, indexing, migrations, transactions, and query optimization to build robust data layers.

## Core Patterns

### Pattern 1: Schema Design with SQLAlchemy
Define database schema using Python ORM for type safety and maintainability.

```python
from sqlalchemy import Column, Integer, String, ForeignKey, DateTime, Index
from sqlalchemy.orm import relationship, declarative_base
from datetime import datetime

Base = declarative_base()

class User(Base):
    __tablename__ = 'users'

    id = Column(Integer, primary_key=True, index=True)
    username = Column(String(50), unique=True, nullable=False, index=True)
    email = Column(String(100), unique=True, nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow)

    # Relationships
    posts = relationship("Post", back_populates="author", cascade="all, delete-orphan")

    # Composite index for common queries
    __table_args__ = (
        Index('idx_user_email_created', 'email', 'created_at'),
    )

class Post(Base):
    __tablename__ = 'posts'

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String(200), nullable=False)
    content = Column(String, nullable=False)
    author_id = Column(Integer, ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    created_at = Column(DateTime, default=datetime.utcnow, index=True)

    # Relationships
    author = relationship("User", back_populates="posts")

    # Indexes for frequent queries
    __table_args__ = (
        Index('idx_post_author_created', 'author_id', 'created_at'),
    )
```

**Best practices**:
- Use foreign keys for referential integrity
- Add indexes on columns used in WHERE, JOIN, ORDER BY
- Use `ondelete='CASCADE'` to maintain data consistency
- Add composite indexes for multi-column queries
- Use `nullable=False` for required fields

### Pattern 2: Database Migrations with Alembic
Manage schema changes safely across environments.

```python
# alembic/env.py configuration
from alembic import context
from sqlalchemy import engine_from_config, pool
from myapp.models import Base

config = context.config
target_metadata = Base.metadata

def run_migrations_online():
    connectable = engine_from_config(
        config.get_section(config.config_ini_section),
        prefix='sqlalchemy.',
        poolclass=pool.NullPool,
    )

    with connectable.connect() as connection:
        context.configure(
            connection=connection,
            target_metadata=target_metadata
        )

        with context.begin_transaction():
            context.run_migrations()
```

**Migration workflow**:

```bash
# Initialize Alembic
alembic init alembic

# Generate migration from model changes
alembic revision --autogenerate -m "Add user posts relationship"

# Review generated migration file
# alembic/versions/xxxx_add_user_posts_relationship.py

# Apply migrations
alembic upgrade head

# Rollback migration
alembic downgrade -1
```

**Example migration file**:

```python
"""Add user posts relationship

Revision ID: abc123
Revises: def456
Create Date: 2024-01-15 10:30:00
"""
from alembic import op
import sqlalchemy as sa

def upgrade():
    op.create_table(
        'posts',
        sa.Column('id', sa.Integer(), nullable=False),
        sa.Column('title', sa.String(200), nullable=False),
        sa.Column('content', sa.String(), nullable=False),
        sa.Column('author_id', sa.Integer(), nullable=False),
        sa.Column('created_at', sa.DateTime(), nullable=True),
        sa.ForeignKeyConstraint(['author_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id')
    )
    op.create_index('idx_post_author_created', 'posts', ['author_id', 'created_at'])

def downgrade():
    op.drop_index('idx_post_author_created', table_name='posts')
    op.drop_table('posts')
```

### Pattern 3: Transactions and Data Integrity
Ensure data consistency with proper transaction management.

```python
from sqlalchemy.orm import Session
from contextlib import contextmanager

@contextmanager
def get_db_session():
    """Context manager for database sessions"""
    session = SessionLocal()
    try:
        yield session
        session.commit()
    except Exception:
        session.rollback()
        raise
    finally:
        session.close()

# Using transactions
async def transfer_funds(from_user_id: int, to_user_id: int, amount: float):
    """Transfer funds between users atomically"""
    async with AsyncSession(engine) as session:
        try:
            # Start transaction
            async with session.begin():
                # Lock rows to prevent race conditions
                from_user = await session.execute(
                    select(User).where(User.id == from_user_id).with_for_update()
                )
                from_user = from_user.scalar_one()

                to_user = await session.execute(
                    select(User).where(User.id == to_user_id).with_for_update()
                )
                to_user = to_user.scalar_one()

                # Validate business logic
                if from_user.balance < amount:
                    raise ValueError("Insufficient funds")

                # Update balances
                from_user.balance -= amount
                to_user.balance += amount

                # Transaction commits automatically if no exception
        except Exception as e:
            # Transaction rolls back automatically
            raise
```

**Transaction isolation levels**:

```python
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

# Read Committed (default)
engine = create_engine(
    "postgresql://user:password@localhost/db",
    isolation_level="READ COMMITTED"
)

# Serializable (strictest)
engine = create_engine(
    "postgresql://user:password@localhost/db",
    isolation_level="SERIALIZABLE"
)
```

## Progressive Disclosure

### Query Optimization

```python
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload, joinedload

# N+1 Query Problem (BAD)
users = session.execute(select(User)).scalars().all()
for user in users:
    print(user.posts)  # Separate query for each user!

# Eager Loading (GOOD)
users = session.execute(
    select(User).options(selectinload(User.posts))
).scalars().all()
for user in users:
    print(user.posts)  # Already loaded

# Joined Load for one-to-one
users = session.execute(
    select(User).options(joinedload(User.profile))
).scalars().unique().all()

# Pagination
def get_users_paginated(page: int = 1, per_page: int = 20):
    offset = (page - 1) * per_page
    query = select(User).offset(offset).limit(per_page)
    return session.execute(query).scalars().all()
```

### Indexes and Performance

```python
# Single column index
class User(Base):
    __tablename__ = 'users'
    email = Column(String(100), unique=True, index=True)

# Composite index for multi-column queries
class Post(Base):
    __tablename__ = 'posts'
    __table_args__ = (
        Index('idx_author_created', 'author_id', 'created_at'),
    )

# Partial index (PostgreSQL)
from sqlalchemy import Index, text

Index(
    'idx_active_users',
    User.email,
    postgresql_where=text('is_active = true')
)

# Full-text search index (PostgreSQL)
from sqlalchemy.dialects.postgresql import TSVECTOR

class Article(Base):
    __tablename__ = 'articles'
    title = Column(String(200))
    content = Column(String)
    search_vector = Column(TSVECTOR)

    __table_args__ = (
        Index('idx_article_search', 'search_vector', postgresql_using='gin'),
    )
```

### Connection Pooling

```python
from sqlalchemy import create_engine
from sqlalchemy.pool import QueuePool

# Configure connection pool
engine = create_engine(
    "postgresql://user:password@localhost/db",
    poolclass=QueuePool,
    pool_size=20,        # Number of permanent connections
    max_overflow=10,     # Additional connections when busy
    pool_timeout=30,     # Seconds to wait for connection
    pool_recycle=3600,   # Recycle connections after 1 hour
    pool_pre_ping=True,  # Verify connection before use
)
```

## Works Well With
- **moai-lang-fastapi-patterns** - API integration with database
- **moai-lang-python** - Python ORM and type hints
- **moai-domain-backend** - Backend architecture patterns

## Common Pitfalls

1. **Missing indexes**
   - ❌ Wrong: No indexes on foreign keys or WHERE clauses
   - ✅ Right: Add indexes on columns used in queries

2. **N+1 query problem**
   - ❌ Wrong: Loading relationships in loops
   - ✅ Right: Use `selectinload()` or `joinedload()`

3. **Long transactions**
   - ❌ Wrong: Keeping transactions open during external API calls
   - ✅ Right: Keep transactions short and focused

4. **No migration strategy**
   - ❌ Wrong: Modifying schema directly in production
   - ✅ Right: Use Alembic for version-controlled migrations

## References
- [SQLAlchemy Documentation](https://docs.sqlalchemy.org/)
- [Alembic Documentation](https://alembic.sqlalchemy.org/)
- [PostgreSQL Performance Tips](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Database Indexing Strategies](https://use-the-index-luke.com/)
