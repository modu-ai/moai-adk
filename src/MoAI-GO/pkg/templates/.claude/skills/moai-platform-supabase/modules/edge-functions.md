---
name: edge-functions
description: Serverless Deno functions at the edge with authentication and rate limiting
parent-skill: moai-platform-supabase
version: 1.0.0
updated: 2026-01-06
---

# Edge Functions Module

## Overview

Supabase Edge Functions are serverless functions running on the Deno runtime at the edge, providing low-latency responses globally.

## Basic Edge Function

### Function Structure

```typescript
// supabase/functions/api/index.ts
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Headers': 'authorization, x-client-info, apikey, content-type'
}

serve(async (req) => {
  // Handle CORS preflight
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  try {
    const supabase = createClient(
      Deno.env.get('SUPABASE_URL')!,
      Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
    )

    // Process request
    const body = await req.json()

    return new Response(
      JSON.stringify({ success: true, data: body }),
      { headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  } catch (error) {
    return new Response(
      JSON.stringify({ error: error.message }),
      { status: 500, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  }
})
```

## Authentication

### JWT Token Verification

```typescript
serve(async (req) => {
  if (req.method === 'OPTIONS') {
    return new Response('ok', { headers: corsHeaders })
  }

  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
  )

  // Verify JWT token
  const authHeader = req.headers.get('authorization')
  if (!authHeader) {
    return new Response(
      JSON.stringify({ error: 'Unauthorized' }),
      { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  }

  const { data: { user }, error } = await supabase.auth.getUser(
    authHeader.replace('Bearer ', '')
  )

  if (error || !user) {
    return new Response(
      JSON.stringify({ error: 'Invalid token' }),
      { status: 401, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
    )
  }

  // User is authenticated, proceed with request
  return new Response(
    JSON.stringify({ success: true, user_id: user.id }),
    { headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
  )
})
```

### Using User Context Client

Create a client that inherits user permissions:

```typescript
serve(async (req) => {
  const authHeader = req.headers.get('authorization')!

  // Client with user's permissions (respects RLS)
  const supabaseUser = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_ANON_KEY')!,
    { global: { headers: { Authorization: authHeader } } }
  )

  // This query respects RLS policies
  const { data, error } = await supabaseUser
    .from('projects')
    .select('*')

  return new Response(JSON.stringify({ data }))
})
```

## Rate Limiting

### Database-Based Rate Limiting

```sql
-- Rate limits table
CREATE TABLE rate_limits (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  identifier TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_rate_limits_lookup ON rate_limits(identifier, created_at);
```

### Rate Limit Function

```typescript
async function checkRateLimit(
  supabase: SupabaseClient,
  identifier: string,
  limit: number,
  windowSeconds: number
): Promise<boolean> {
  const windowStart = new Date(Date.now() - windowSeconds * 1000).toISOString()

  const { count } = await supabase
    .from('rate_limits')
    .select('*', { count: 'exact', head: true })
    .eq('identifier', identifier)
    .gte('created_at', windowStart)

  if (count && count >= limit) {
    return false
  }

  await supabase.from('rate_limits').insert({ identifier })
  return true
}
```

### Usage in Edge Function

```typescript
serve(async (req) => {
  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
  )

  // Get client identifier (IP or user ID)
  const identifier = req.headers.get('x-forwarded-for') || 'anonymous'

  // 100 requests per minute
  const allowed = await checkRateLimit(supabase, identifier, 100, 60)

  if (!allowed) {
    return new Response(
      JSON.stringify({ error: 'Rate limit exceeded' }),
      { status: 429, headers: corsHeaders }
    )
  }

  // Process request...
})
```

## External API Integration

### Webhook Handler

```typescript
serve(async (req) => {
  const supabase = createClient(
    Deno.env.get('SUPABASE_URL')!,
    Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!
  )

  // Verify webhook signature
  const signature = req.headers.get('x-webhook-signature')
  const body = await req.text()

  const expectedSignature = await crypto.subtle.digest(
    'SHA-256',
    new TextEncoder().encode(body + Deno.env.get('WEBHOOK_SECRET'))
  )

  if (!verifySignature(signature, expectedSignature)) {
    return new Response('Invalid signature', { status: 401 })
  }

  const payload = JSON.parse(body)

  // Process webhook
  await supabase.from('webhook_events').insert({
    type: payload.type,
    data: payload.data,
    processed: false
  })

  return new Response('OK', { status: 200 })
})
```

### External API Call

```typescript
serve(async (req) => {
  const { query } = await req.json()

  // Call external API
  const response = await fetch('https://api.openai.com/v1/embeddings', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${Deno.env.get('OPENAI_API_KEY')}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model: 'text-embedding-ada-002',
      input: query
    })
  })

  const data = await response.json()

  return new Response(
    JSON.stringify({ embedding: data.data[0].embedding }),
    { headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
  )
})
```

## Error Handling

### Structured Error Response

```typescript
interface ErrorResponse {
  error: string
  code: string
  details?: unknown
}

function errorResponse(
  message: string,
  code: string,
  status: number,
  details?: unknown
): Response {
  const body: ErrorResponse = { error: message, code, details }
  return new Response(
    JSON.stringify(body),
    { status, headers: { ...corsHeaders, 'Content-Type': 'application/json' } }
  )
}

serve(async (req) => {
  try {
    // ... processing
  } catch (error) {
    if (error instanceof AuthError) {
      return errorResponse('Authentication failed', 'AUTH_ERROR', 401)
    }
    if (error instanceof ValidationError) {
      return errorResponse('Invalid input', 'VALIDATION_ERROR', 400, error.details)
    }
    console.error('Unexpected error:', error)
    return errorResponse('Internal server error', 'INTERNAL_ERROR', 500)
  }
})
```

## Deployment

### Local Development

```bash
supabase functions serve api --env-file .env.local
```

### Deploy Function

```bash
supabase functions deploy api
```

### Set Secrets

```bash
supabase secrets set OPENAI_API_KEY=sk-xxx
supabase secrets set WEBHOOK_SECRET=whsec-xxx
```

### List Functions

```bash
supabase functions list
```

## Best Practices

### Cold Start Optimization

Keep imports minimal at the top level:

```typescript
// Good: Import only what's needed
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'

// Bad: Heavy imports at top level increase cold start
// import { everything } from 'large-library'
```

### Response Streaming

Stream large responses:

```typescript
serve(async (req) => {
  const stream = new ReadableStream({
    async start(controller) {
      for (let i = 0; i < 10; i++) {
        await new Promise(r => setTimeout(r, 100))
        controller.enqueue(new TextEncoder().encode(`data: ${i}\n\n`))
      }
      controller.close()
    }
  })

  return new Response(stream, {
    headers: { 'Content-Type': 'text/event-stream' }
  })
})
```

## Context7 Query Examples

For latest Edge Functions documentation:

Topic: "edge functions deno runtime"
Topic: "supabase functions deploy secrets"
Topic: "edge functions cors authentication"

---

Related Modules:
- auth-integration.md - Authentication patterns
- typescript-patterns.md - Client invocation
