# moai-saas-supabase-mcp

Master Supabase for PostgreSQL database, authentication, real-time subscriptions, and vector search.

## Quick Start

Supabase is an open-source Firebase alternative with a PostgreSQL database, real-time subscriptions, and built-in authentication. Use this skill when setting up databases, implementing authentication, creating real-time features, or working with vector embeddings.

## Core Patterns

### Pattern 1: Database Schema & Row Level Security

**Pattern**: Design PostgreSQL schema with row-level security (RLS) policies for multi-tenant applications.

```sql
-- Create users table with RLS enabled
CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text UNIQUE NOT NULL,
  username text UNIQUE NOT NULL,
  avatar_url text,
  created_at timestamptz DEFAULT now()
);

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can only read their own profile
CREATE POLICY "Users can read own profile"
  ON users FOR SELECT
  USING (auth.uid() = id);

-- Create posts table with foreign key
CREATE TABLE posts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users(id) ON DELETE CASCADE NOT NULL,
  title text NOT NULL,
  content text NOT NULL,
  published boolean DEFAULT false,
  created_at timestamptz DEFAULT now()
);

ALTER TABLE posts ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can read published posts + their own posts
CREATE POLICY "Users can read published or own posts"
  ON posts FOR SELECT
  USING (published = true OR auth.uid() = user_id);

-- RLS Policy: Users can only create posts for themselves
CREATE POLICY "Users can create own posts"
  ON posts FOR INSERT
  WITH CHECK (auth.uid() = user_id);

-- Index for performance
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_published ON posts(published) WHERE published = true;
```

**When to use**:
- Multi-tenant SaaS applications
- Protecting sensitive user data
- Implementing role-based access control
- Ensuring data isolation between users

**Key benefits**:
- Database-level security (not application-level)
- Zero-trust architecture
- GDPR-compliant data protection
- Scalable authorization logic

### Pattern 2: Real-time Subscriptions with Broadcast

**Pattern**: Implement real-time features using Supabase's WebSocket subscriptions.

```typescript
// React component with real-time updates
import { useEffect, useState } from 'react';
import { createClient } from '@supabase/supabase-js';

const supabase = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL,
  process.env.NEXT_PUBLIC_SUPABASE_KEY
);

export function LivePostList() {
  const [posts, setPosts] = useState([]);

  useEffect(() => {
    // Fetch initial data
    const fetchPosts = async () => {
      const { data } = await supabase
        .from('posts')
        .select('*')
        .eq('published', true);
      setPosts(data || []);
    };

    fetchPosts();

    // Subscribe to changes (INSERT, UPDATE, DELETE)
    const subscription = supabase
      .from('posts')
      .on('*', (payload) => {
        console.log('Change received!', payload);

        if (payload.eventType === 'INSERT') {
          setPosts((prev) => [...prev, payload.new]);
        } else if (payload.eventType === 'UPDATE') {
          setPosts((prev) =>
            prev.map((p) => (p.id === payload.new.id ? payload.new : p))
          );
        } else if (payload.eventType === 'DELETE') {
          setPosts((prev) => prev.filter((p) => p.id !== payload.old.id));
        }
      })
      .subscribe();

    return () => {
      subscription.unsubscribe();
    };
  }, []);

  return (
    <div>
      {posts.map((post) => (
        <article key={post.id}>
          <h2>{post.title}</h2>
          <p>{post.content}</p>
        </article>
      ))}
    </div>
  );
}
```

**When to use**:
- Collaborative editing features
- Live notifications and feeds
- Real-time dashboards
- Chat and messaging applications

**Key benefits**:
- Sub-100ms latency updates
- No polling required (efficient)
- Automatic reconnection handling
- Typed events with payload data

### Pattern 3: Authentication with OAuth & MFA

**Pattern**: Implement authentication with multiple providers and multi-factor authentication.

```typescript
// Next.js authentication flow
import { createClient } from '@supabase/supabase-js';

const supabase = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL,
  process.env.NEXT_PUBLIC_SUPABASE_KEY
);

// Sign up with email
export async function signUp(email: string, password: string) {
  const { data, error } = await supabase.auth.signUp({
    email,
    password,
    options: {
      emailRedirectTo: `${process.env.NEXT_PUBLIC_SITE_URL}/auth/callback`,
    },
  });

  return { data, error };
}

// Sign in with OAuth provider
export async function signInWithGoogle() {
  const { data, error } = await supabase.auth.signInWithOAuth({
    provider: 'google',
    options: {
      redirectTo: `${process.env.NEXT_PUBLIC_SITE_URL}/auth/callback`,
    },
  });

  return { data, error };
}

// Enable MFA for user
export async function enableMFA(userId: string) {
  const { data, error } = await supabase.auth.admin.updateUserById(userId, {
    factors: [
      {
        factorType: 'totp',
        friendlyName: 'Authenticator App',
      },
    ],
  });

  return { data, error };
}

// Get session and verify user
export async function getSession() {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  return session;
}
```

**When to use**:
- User authentication in web/mobile apps
- Social login (Google, GitHub, Facebook)
- Enterprise SSO (Okta, SAML)
- Security-critical applications requiring MFA

**Key benefits**:
- Built-in password hashing (bcrypt)
- OAuth provider integration
- Email verification flows
- Session management and JWT tokens

## Progressive Disclosure

### Level 1: Basic Setup
- Create Supabase project and database
- Create tables with simple schema
- Enable basic authentication
- Write data with simple queries

### Level 2: Advanced Features
- Implement row-level security (RLS)
- Set up real-time subscriptions
- Configure OAuth providers
- Use vector search for embeddings

### Level 3: Expert Optimization
- Design complex multi-tenant schemas
- Optimize query performance with indexes
- Implement custom authentication flows
- Use pgvector for AI search features

## Works Well With

- **PostgreSQL 15+**: Full-featured open-source database
- **Next.js 16**: Server-side queries and streaming
- **React 19**: Real-time updates with use() hook
- **Render**: Deploy FastAPI backends connected to Supabase
- **Vercel**: Deploy Next.js frontends with Supabase backend
- **pgvector**: Vector embeddings for AI applications

## References

- **Official Documentation**: https://supabase.com/docs
- **PostgreSQL Guide**: https://www.postgresql.org/docs/
- **Real-time Subscriptions**: https://supabase.com/docs/guides/realtime
- **RLS Policies**: https://supabase.com/docs/guides/auth/row-level-security
- **Authentication**: https://supabase.com/docs/guides/auth
