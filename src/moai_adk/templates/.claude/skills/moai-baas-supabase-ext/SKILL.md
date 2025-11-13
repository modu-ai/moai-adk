---
name: "moai-baas-supabase-ext"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise Supabase PostgreSQL Platform with AI-powered open-source BaaS architecture, Context7 integration, and intelligent PostgreSQL orchestration for scalable modern applications
keywords: ['supabase', 'postgresql', 'open-source', 'realtime', 'row-level-security', 'edge-functions', 'context7-integration', 'ai-orchestration', 'production-deployment']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Supabase PostgreSQL Platform Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-supabase-ext |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Database Platform Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when Supabase keywords detected |

---

## What It Does

Enterprise Supabase PostgreSQL Platform expert with AI-powered open-source BaaS architecture, Context7 integration, and intelligent PostgreSQL orchestration for scalable modern applications.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Supabase Architecture** using Context7 MCP for latest PostgreSQL documentation
- 📊 **Intelligent Real-time Orchestration** with automated subscription optimization
- 🚀 **Advanced PostgreSQL Optimization** with AI-driven query tuning and indexing
- 🔗 **Enterprise Row-Level Security** with zero-configuration data protection
- 📈 **Predictive Performance Analytics** with usage forecasting and optimization

---

## When to Use

**Automatic triggers**:
- Supabase PostgreSQL architecture and real-time feature discussions
- Row-level security implementation and optimization
- Edge Functions and serverless backend development
- Open-source BaaS platform evaluation and integration

**Manual invocation**:
- Designing enterprise Supabase architectures with PostgreSQL optimization
- Implementing real-time subscriptions and data synchronization
- Planning PostgreSQL to Supabase migrations
- Configuring advanced security with Row-Level Security

---

# Quick Reference (Level 1)

## Supabase PostgreSQL Platform (November 2025)

### Core Features Overview
- **PostgreSQL 16+**: Latest version with extensions and performance improvements
- **Real-time Subscriptions**: WebSocket-based real-time data synchronization
- **Row-Level Security (RLS)**: Fine-grained data access control
- **Edge Functions**: Deno runtime with 28+ global regions
- **File Storage**: Object storage with CDN integration
- **Authentication**: Multi-provider auth with PostgreSQL integration

### Key PostgreSQL Extensions
- **pgvector**: Vector similarity search for AI applications
- **pg_cron**: Scheduled database jobs and maintenance
- **PostgREST**: Automatic REST API generation from PostgreSQL schemas
- **pg_graphql**: GraphQL API generation from PostgreSQL

### Latest Features (November 2025)
- **Database Branching**: Development workflows with isolated environments
- **Improved Auth Section**: User bans and authenticated logs
- **Official Vercel Integration**: Seamless deployment and management
- **Enhanced Edge Functions**: Improved cold start performance

### Performance Characteristics
- **Query Performance**: P95 < 50ms latency
- **Real-time Latency**: WebSocket connections < 100ms
- **Throughput**: 50k+ TPS with connection pooling
- **Global Coverage**: 28+ edge regions worldwide

---

# Core Implementation (Level 2)

## Supabase Architecture Intelligence

```python
# AI-powered Supabase architecture optimization with Context7
class SupabaseArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.postgresql_analyzer = PostgreSQLAnalyzer()
        self.realtime_optimizer = RealtimeOptimizer()
    
    async def design_optimal_supabase_architecture(self, 
                                                 requirements: ApplicationRequirements) -> SupabaseArchitecture:
        """Design optimal Supabase architecture using AI analysis."""
        
        # Get latest Supabase and PostgreSQL documentation via Context7
        supabase_docs = await self.context7_client.get_library_docs(
            context7_library_id='/supabase/docs',
            topic="postgresql rls realtime edge-functions optimization 2025",
            tokens=3000
        )
        
        postgresql_docs = await self.context7_client.get_library_docs(
            context7_library_id='/postgresql/docs',
            topic="performance optimization indexing extensions 2025",
            tokens=2000
        )
        
        # Optimize PostgreSQL configuration
        db_optimization = self.postgresql_analyzer.optimize_configuration(
            requirements.database_workload,
            postgresql_docs
        )
        
        # Design real-time architecture
        realtime_design = self.realtime_optimizer.design_realtime_system(
            requirements.realtime_needs,
            supabase_docs
        )
        
        return SupabaseArchitecture(
            database_configuration=db_optimization,
            realtime_system=realtime_design,
            edge_functions=self._design_edge_functions(requirements),
            storage_configuration=self._optimize_storage(requirements),
            auth_integration=self._integrate_authentication(requirements),
            rls_policies=self._design_row_level_security(requirements),
            performance_predictions=db_optimization.predictions
        )
```

## Row-Level Security Implementation

```sql
-- Comprehensive RLS policies for enterprise applications
-- Enable RLS on sensitive tables
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

-- User can only access their own profile
CREATE POLICY "Users can view own profile" ON profiles
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can update own profile" ON profiles
    FOR UPDATE USING (auth.uid() = user_id);

-- Project access based on team membership
CREATE POLICY "Team members can view team projects" ON projects
    FOR SELECT USING (
        user_id IN (
            SELECT user_id FROM team_members 
            WHERE team_id = projects.team_id
        )
    );

CREATE POLICY "Team admins can manage team projects" ON projects
    FOR ALL USING (
        user_id IN (
            SELECT user_id FROM team_members 
            WHERE team_id = projects.team_id AND role = 'admin'
        )
    );

-- Document access with role-based permissions
CREATE POLICY "Document access control" ON documents
    FOR ALL USING (
        CASE
            WHEN is_public THEN true
            WHEN owner_id = auth.uid() THEN true
            WHEN user_id IN (
                SELECT user_id FROM document_shares
                WHERE document_id = documents.id AND permission IN ('read', 'write', 'admin')
            ) THEN true
            ELSE false
        END
    );
```

## Real-time Subscription Patterns

```typescript
// Advanced real-time subscription management
import { RealtimeChannel } from '@supabase/supabase-js';

class RealtimeManager {
  private subscriptions: Map<string, RealtimeChannel> = new Map();
  
  subscribeToDataChanges(
    table: string, 
    filter: object, 
    callback: (payload: any) => void
  ): string {
    const subscriptionId = `${table}_${Date.now()}`;
    
    const channel = supabase
      .channel(subscriptionId)
      .on(
        'postgres_changes',
        { 
          event: '*', 
          schema: 'public', 
          table: table,
          filter: filter
        },
        callback
      )
      .subscribe();
    
    this.subscriptions.set(subscriptionId, channel);
    return subscriptionId;
  }
  
  subscribeToUserSpecificData(userId: string) {
    return {
      profile: this.subscribeToDataChanges(
        'profiles',
        { user_id: `eq.${userId}` },
        (payload) => this.handleProfileUpdate(payload)
      ),
      projects: this.subscribeToDataChanges(
        'projects',
        { team_id: `in.(${this.getUserTeamIds(userId).join(',')})` },
        (payload) => this.handleProjectUpdate(payload)
      ),
      notifications: this.subscribeToDataChanges(
        'notifications',
        { user_id: `eq.${userId}` },
        (payload) => this.handleNotification(payload)
      )
    };
  }
  
  unsubscribe(subscriptionId: string) {
    const channel = this.subscriptions.get(subscriptionId);
    if (channel) {
      supabase.removeChannel(channel);
      this.subscriptions.delete(subscriptionId);
    }
  }
}
```

---

# Advanced Implementation (Level 3)

## Edge Functions Development

```typescript
// Advanced Edge Function with TypeScript
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts';
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2';

interface NotificationPayload {
  userId: string;
  type: 'message' | 'mention' | 'project_update';
  title: string;
  body: string;
  data?: Record<string, any>;
}

serve(async (req) => {
  try {
    const { userId, type, title, body, data }: NotificationPayload = await req.json();
    
    // Initialize Supabase client
    const supabaseUrl = Deno.env.get('SUPABASE_URL')!;
    const supabaseKey = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!;
    const supabase = createClient(supabaseUrl, supabaseKey);
    
    // Validate user exists and has notifications enabled
    const { data: user, error: userError } = await supabase
      .from('profiles')
      .select('id, notification_preferences')
      .eq('id', userId)
      .single();
    
    if (userError || !user?.notification_preferences?.[type]) {
      return new Response(JSON.stringify({ error: 'Invalid user or preferences' }), {
        status: 400,
        headers: { 'Content-Type': 'application/json' }
      });
    }
    
    // Store notification in database
    const { error: insertError } = await supabase
      .from('notifications')
      .insert({
        user_id: userId,
        type,
        title,
        body,
        data,
        created_at: new Date().toISOString()
      });
    
    if (insertError) throw insertError;
    
    // Trigger real-time notification
    const { error: realtimeError } = await supabase
      .channel('notifications')
      .send({
        type: 'broadcast',
        event: 'new_notification',
        payload: { userId, type, title, body, data }
      });
    
    return new Response(JSON.stringify({ success: true }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' }
    });
    
  } catch (error) {
    console.error('Notification function error:', error);
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    });
  }
});
```

### Migration from Firebase to Supabase

```python
class FirebaseToSupabaseMigrator:
    def __init__(self):
        self.supabase_client = SupabaseClient()
        self.firebase_client = FirebaseClient()
        self.data_transformer = DataTransformer()
    
    async def migrate_firestore_to_supabase(self, 
                                          firebase_config: FirebaseConfig,
                                          supabase_config: SupabaseConfig) -> MigrationResult:
        """Migrate Firestore collections to Supabase PostgreSQL."""
        
        # Get all Firestore collections
        collections = await self.firebase_client.list_collections()
        
        migration_results = []
        
        for collection in collections:
            # Get all documents from collection
            documents = await self.firebase_client.get_collection_documents(collection)
            
            # Transform Firebase data to PostgreSQL schema
            transformed_data = await self.data_transformer.transform_firestore_to_postgresql(
                documents, collection
            )
            
            # Create PostgreSQL table if not exists
            await self.supabase_client.create_table_from_schema(
                collection, transformed_data.schema
            )
            
            # Insert transformed data
            insert_result = await self.supabase_client.bulk_insert(
                collection, transformed_data.data
            )
            
            migration_results.append({
                collection: collection,
                documents_migrated: len(documents),
                status: insert_result.status
            })
        
        return MigrationResult(
            collections_migrated=len(collections),
            total_documents=sum(r['documents_migrated'] for r in migration_results),
            results=migration_results,
            success=True
        )
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Supabase Operations
- `create_table(schema_definition)` - Create PostgreSQL table
- `enable_rls(table_name)` - Enable Row-Level Security
- `create_policy(table, name, definition)` - Create RLS policy
- `deploy_edge_function(name, code)` - Deploy Edge Function
- `create_realtime_subscription(channel, filter)` - Create real-time subscription

### Context7 Integration
- `get_latest_supabase_documentation()` - Official Supabase docs via Context7
- `analyze_postgresql_optimization()` - PostgreSQL performance via Context7
- `optimize_realtime_architecture()` - Real-time best practices via Context7

## Best Practices (November 2025)

### DO
- Use Row-Level Security for all data access control
- Implement comprehensive database indexing for performance
- Optimize Edge Functions for cold start performance
- Use database branching for development workflows
- Implement connection pooling for high-throughput applications
- Monitor real-time subscriptions for performance and cost
- Use PostgreSQL extensions for specialized functionality
- Implement proper backup and disaster recovery

### DON'T
- Skip RLS policies on sensitive data tables
- Overcomplicate real-time subscriptions with excessive filters
- Ignore Edge Function cold start optimization
- Use production database for development testing
- Neglect PostgreSQL query optimization
- Forget to monitor real-time subscription costs
- Skip database migration testing
- Ignore security best practices for Edge Functions

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS architecture patterns)
- `moai-domain-database` (Database design and optimization)
- `moai-baas-neon-ext` (PostgreSQL alternative comparison)
- `moai-essentials-perf` (Performance optimization)
- `moai-security-api` (API security and authorization)
- `moai-foundation-trust` (Security and compliance)
- `moai-baas-firebase-ext` (Firebase migration comparison)
- `moai-domain-backend` (Backend database integration)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 Supabase platform updates, and advanced PostgreSQL optimization
- **v2.0.0** (2025-11-11): Complete metadata structure, RLS patterns, Edge Functions
- **v1.0.0** (2025-11-11): Initial Supabase PostgreSQL platform

---

**End of Skill** | Updated 2025-11-13

## Security & Compliance

### PostgreSQL Security
- Row-Level Security for fine-grained access control
- PostgreSQL native encryption and authentication
- Comprehensive audit logging and monitoring
- Advanced user privilege management

### Data Protection
- End-to-end encryption for data in transit and at rest
- GDPR compliance with data portability and deletion
- HIPAA ready configuration for healthcare applications
- SOC2 Type II security controls implementation

---

**End of Enterprise Supabase PostgreSQL Platform Expert v4.0.0**
