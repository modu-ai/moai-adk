# Supabase Platform Reference (2025)

## Official Documentation
- [Supabase Documentation](https://supabase.com/docs)
- [PostgreSQL Extensions](https://supabase.com/docs/guides/database/extensions)
- [Edge Functions](https://supabase.com/docs/guides/functions)
- [Realtime Subscriptions](https://supabase.com/docs/guides/realtime)
- [Row Level Security](https://supabase.com/docs/guides/database/postgres/row-level-security)

## Latest Features (2025)

### Database Branching
- Development workflows with isolated environments
- Branch-per-PR for testing
- Automated branch cleanup
- Branch merging with conflict resolution

### Improved Auth Section
- User bans and suspension
- Authenticated logs and audit trails
- Enhanced session management
- Multi-factor authentication improvements

### Official Vercel Integration
- Seamless deployment from Supabase dashboard
- Automatic environment variable sync
- Preview deployments with database branches
- Production-ready configurations

### Enhanced Edge Functions
- Improved cold start performance (< 50ms)
- Deno 1.40+ runtime
- TypeScript support with type inference
- Global deployment across 28+ regions

## Context7 Integration

Access latest Supabase documentation:
```python
docs = await context7.get_library_docs(
    context7_library_id="/supabase/docs",
    topic="edge-functions rls realtime 2025"
)
```

## Client SDK Examples

### JavaScript/TypeScript Client
```typescript
import { createClient } from '@supabase/supabase-js';

const supabase = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL!,
  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!
);

// Query with RLS
const { data, error } = await supabase
  .from('posts')
  .select('*')
  .eq('published', true)
  .order('created_at', { ascending: false });
```

### Python Client
```python
from supabase import create_client, Client

supabase: Client = create_client(
    os.environ["SUPABASE_URL"],
    os.environ["SUPABASE_KEY"]
)

# Query with RLS
response = supabase.table('posts')\
    .select('*')\
    .eq('published', True)\
    .order('created_at', desc=True)\
    .execute()
```

## PostgreSQL Extensions

### pgvector for AI Applications
```sql
-- Enable vector extension
CREATE EXTENSION vector;

-- Create table with vector column
CREATE TABLE documents (
  id bigserial PRIMARY KEY,
  content text,
  embedding vector(1536)
);

-- Create index for similarity search
CREATE INDEX ON documents USING ivfflat (embedding vector_cosine_ops);

-- Similarity search
SELECT * FROM documents
ORDER BY embedding <=> '[...]'::vector
LIMIT 5;
```

### pg_cron for Scheduled Jobs
```sql
-- Enable pg_cron
CREATE EXTENSION pg_cron;

-- Schedule daily cleanup
SELECT cron.schedule(
  'daily-cleanup',
  '0 2 * * *',
  $$DELETE FROM logs WHERE created_at < NOW() - INTERVAL '30 days'$$
);
```

## Best Practices

1. **Use Row-Level Security** - Protect data at the database level
2. **Optimize Edge Functions** - Minimize cold starts with proper bundling
3. **Implement connection pooling** - Use Supavisor for high concurrency
4. **Monitor real-time subscriptions** - Track active connections and costs
5. **Use database branching** - Isolated development and testing

---

**Last Updated**: 2025-11-22
