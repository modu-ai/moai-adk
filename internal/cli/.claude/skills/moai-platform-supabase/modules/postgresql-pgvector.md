---
name: postgresql-pgvector
description: PostgreSQL 16 with pgvector extension for AI embeddings and semantic search
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# PostgreSQL 16 + pgvector Module

## Overview

PostgreSQL 16 with pgvector extension enables AI-powered semantic search through vector embeddings storage and similarity search operations.

## Extension Setup

Enable required extensions for vector operations:

```sql
-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;
```

## Embeddings Table Schema

Create a table optimized for storing AI embeddings:

```sql
CREATE TABLE documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  content TEXT NOT NULL,
  embedding vector(1536),  -- OpenAI ada-002 dimensions
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Common Embedding Dimensions

- OpenAI ada-002: 1536 dimensions
- OpenAI text-embedding-3-small: 1536 dimensions
- OpenAI text-embedding-3-large: 3072 dimensions
- Cohere embed-english-v3.0: 1024 dimensions
- Google PaLM: 768 dimensions

## Index Strategies

### HNSW Index (Recommended)

HNSW (Hierarchical Navigable Small World) provides fast approximate nearest neighbor search:

```sql
CREATE INDEX idx_documents_embedding ON documents
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

Parameters:
- m: Maximum number of connections per layer (default 16, higher = more accurate but slower)
- ef_construction: Size of dynamic candidate list during construction (default 64)

### IVFFlat Index (Large Datasets)

IVFFlat is better for datasets with millions of rows:

```sql
CREATE INDEX idx_documents_ivf ON documents
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);
```

Guidelines for lists parameter:
- Less than 1M rows: lists = rows / 1000
- More than 1M rows: lists = sqrt(rows)

## Distance Operations

Available distance operators:

- `<->` - Euclidean distance (L2)
- `<#>` - Negative inner product
- `<=>` - Cosine distance

For normalized embeddings, cosine distance is recommended.

## Semantic Search Function

Basic semantic search with threshold and limit:

```sql
CREATE OR REPLACE FUNCTION search_documents(
  query_embedding vector(1536),
  match_threshold FLOAT DEFAULT 0.8,
  match_count INT DEFAULT 10
) RETURNS TABLE (id UUID, content TEXT, similarity FLOAT)
LANGUAGE plpgsql AS $$
BEGIN
  RETURN QUERY SELECT d.id, d.content,
    1 - (d.embedding <=> query_embedding) AS similarity
  FROM documents d
  WHERE 1 - (d.embedding <=> query_embedding) > match_threshold
  ORDER BY d.embedding <=> query_embedding
  LIMIT match_count;
END; $$;
```

### Usage Example

```sql
SELECT * FROM search_documents(
  '[0.1, 0.2, ...]'::vector(1536),
  0.75,
  20
);
```

## Hybrid Search

Combine vector similarity with full-text search for better results:

```sql
CREATE OR REPLACE FUNCTION hybrid_search(
  query_text TEXT,
  query_embedding vector(1536),
  match_count INT DEFAULT 10,
  full_text_weight FLOAT DEFAULT 0.3,
  semantic_weight FLOAT DEFAULT 0.7
) RETURNS TABLE (id UUID, content TEXT, score FLOAT) AS $$
BEGIN
  RETURN QUERY
  WITH semantic AS (
    SELECT e.id, e.content, 1 - (e.embedding <=> query_embedding) AS similarity
    FROM documents e ORDER BY e.embedding <=> query_embedding LIMIT match_count * 2
  ),
  full_text AS (
    SELECT e.id, e.content,
      ts_rank(to_tsvector('english', e.content), plainto_tsquery('english', query_text)) AS rank
    FROM documents e
    WHERE to_tsvector('english', e.content) @@ plainto_tsquery('english', query_text)
    LIMIT match_count * 2
  )
  SELECT COALESCE(s.id, f.id), COALESCE(s.content, f.content),
    (COALESCE(s.similarity, 0) * semantic_weight + COALESCE(f.rank, 0) * full_text_weight)
  FROM semantic s FULL OUTER JOIN full_text f ON s.id = f.id
  ORDER BY 3 DESC LIMIT match_count;
END; $$ LANGUAGE plpgsql;
```

## Full-Text Search Index

Add GIN index for efficient full-text search:

```sql
CREATE INDEX idx_documents_content_fts ON documents
USING gin(to_tsvector('english', content));
```

## Performance Optimization

### Query Performance

Set appropriate ef_search for HNSW queries:

```sql
SET hnsw.ef_search = 100;  -- Higher = more accurate, slower
```

### Batch Insertions

Use COPY or multi-row INSERT for bulk embeddings:

```sql
INSERT INTO documents (content, embedding, metadata)
VALUES
  ('Content 1', '[...]'::vector(1536), '{"source": "doc1"}'),
  ('Content 2', '[...]'::vector(1536), '{"source": "doc2"}'),
  ('Content 3', '[...]'::vector(1536), '{"source": "doc3"}');
```

### Index Maintenance

Reindex after large bulk insertions:

```sql
REINDEX INDEX CONCURRENTLY idx_documents_embedding;
```

## Metadata Filtering

Combine vector search with JSONB metadata filters:

```sql
CREATE OR REPLACE FUNCTION search_with_filters(
  query_embedding vector(1536),
  filter_metadata JSONB,
  match_count INT DEFAULT 10
) RETURNS TABLE (id UUID, content TEXT, similarity FLOAT) AS $$
BEGIN
  RETURN QUERY SELECT d.id, d.content,
    1 - (d.embedding <=> query_embedding) AS similarity
  FROM documents d
  WHERE d.metadata @> filter_metadata
  ORDER BY d.embedding <=> query_embedding
  LIMIT match_count;
END; $$;
```

### Usage with Filters

```sql
SELECT * FROM search_with_filters(
  '[0.1, 0.2, ...]'::vector(1536),
  '{"category": "technical", "language": "en"}'::jsonb,
  10
);
```

## Context7 Query Examples

For latest pgvector documentation:

Topic: "pgvector extension indexes hnsw ivfflat"
Topic: "vector similarity search operators"
Topic: "postgresql full-text search tsvector"

---

Related Modules:
- row-level-security.md - Secure vector data access
- typescript-patterns.md - Client-side search implementation
