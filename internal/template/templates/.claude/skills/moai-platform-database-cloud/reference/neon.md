# Neon Serverless PostgreSQL Specialist

## Core Capabilities

Serverless Compute provides auto-scaling PostgreSQL with scale-to-zero for cost optimization. Compute automatically scales based on load and shuts down during idle periods.

Database Branching enables instant copy-on-write branches for development, staging, and preview environments. Branches are created instantly without duplicating storage.

Point-in-Time Recovery offers 30-day PITR with instant restore to any timestamp within the retention window.

Connection Pooling provides built-in connection pooler (PgBouncer) for serverless and edge runtime compatibility. Supports transaction and session pooling modes.

PostgreSQL 16 Support ensures full PostgreSQL 16 compatibility including extensions, JSONB, and advanced SQL features.

## Setup and Configuration

Package Installation requires @neondatabase/serverless from npm. Optionally install drizzle-orm for Drizzle ORM integration or @prisma/client and prisma for Prisma ORM integration.

Environment Configuration requires DATABASE_URL for direct connection used for migrations, formatted as postgresql://user:pass@ep-xxx.region.neon.tech/dbname?sslmode=require. DATABASE_URL_POOLED provides pooled connection for serverless and edge, formatted as postgresql://user:pass@ep-xxx-pooler.region.neon.tech/dbname?sslmode=require. NEON_API_KEY provides the Neon API key for branching operations. NEON_PROJECT_ID provides the project identifier.

## Serverless Driver Usage

Basic Query Execution requires importing neon from @neondatabase/serverless. Create the sql function by calling neon with the DATABASE_URL environment variable. Execute simple queries using tagged template literals with the sql function. For parameterized queries that are SQL injection safe, include variables inside the template literal. The driver supports transaction operations using sql.transaction with an array of SQL statements.

Drizzle ORM Integration requires importing table and column types from drizzle-orm/pg-core. Define tables using pgTable with column definitions. Use uuid for UUIDs with primaryKey and defaultRandom, text for strings with notNull and unique modifiers, timestamp for dates with defaultNow, and jsonb for JSON columns. For Drizzle client setup, import neon from @neondatabase/serverless, drizzle from drizzle-orm/neon-http, and the schema module. Create the sql function with neon and the DATABASE_URL. Export the db instance created with drizzle passing sql and the schema. Execute queries using db.select().from(schema.tableName).

Prisma ORM Integration requires importing Pool and neonConfig from @neondatabase/serverless, PrismaNeon from @prisma/adapter-neon, and PrismaClient from @prisma/client. Set neonConfig.webSocketConstructor to the ws module. Create a Pool with the DATABASE_URL connection string. Create an adapter with PrismaNeon passing the pool. Export the prisma instance created with PrismaClient passing the adapter.

## Database Branching Workflows

Branch Creation uses neonctl branches create command with parent branch name and branch name. Each branch is a copy-on-write fork that shares storage with parent until divergence.

Branch per Pattern enables creating a new branch for each pull request in CI/CD pipeline. Branch inherits parent data for realistic testing environment.

Branch Reset allows restoring branch to parent state using neonctl branches reset command. Useful for cleaning up test data between test runs.

Branch Delete removes branches using neonctl branches delete command when PR is merged or workflow completes.

## Auto-Scaling and Compute

Compute Units scale horizontally based on load. Active compute units handle connections and queries while inactive units are suspended.

Scale-to-Zero configuration enables automatic suspension during idle periods. No charges incurred during suspended state. Cold start time approximately 1-3 seconds on activation.

Cost Optimization involves monitoring compute hours usage, setting appropriate autoscaling limits, and leveraging scale-to-zero for development environments.

## Connection Pooling

Pooled Connection uses DATABASE_URL_POOLED endpoint with PgBouncer in transaction mode. Recommended for serverless functions and edge runtimes.

Pool Mode Selection depends on use case. Transaction mode for serverless and edge (default). Session mode for long-lived connections in traditional servers.

Connection Limits are enforced by pooler configuration. Configure max_client_conn in pooler settings based on expected concurrency.

## PITR and Backups

Point-in-Time Recovery enables restoring to any timestamp within 30-day retention window. Use neonctl time-travel inspect to view available restore points.

Branch Restoration creates new branch from historical timestamp using neonctl branches restore --timestamp command. Useful for data recovery and auditing.

Backup Strategy includes automated continuous backups with 30-day retention. Additional manual backups can be created using neonctl backups create.

## Edge Deployment

Edge Runtime Compatibility requires using pooled connection endpoint (DATABASE_URL_POOLED) with serverless driver. Edge functions have cold starts and connection limits.

Vercel Integration uses Vercel Postgres with Neon backend. Configure POSTGRES_URL and POSTGRES_URL_NON_POOLING environment variables.

Cloudflare Workers requires using pooled connection with fetch API compatibility. Consider connection reuse patterns for performance.

## When to Use Neon

Serverless Applications benefit from auto-scaling and scale-to-zero that reduce costs significantly compared to always-on databases.

Preview Environments benefit from instant branching that enables per-PR databases with production data snapshots.

Edge Deployment benefits from connection pooling that provides edge runtime compatibility without managing connection pools.

Development Workflow benefits from branching from production for realistic development data without affecting production.

Cost Optimization benefits from paying only for active compute time with scale-to-zero during idle periods.

## When to Consider Alternatives

Need Vector Search: Consider Supabase with pgvector or dedicated vector database for native vector similarity search.

Need Real-time Subscriptions: Consider Supabase or Convex for built-in real-time features and change listeners.

Need NoSQL Flexibility: Consider Firestore or Convex for document storage with flexible schemas.

Need Built-in Auth: Consider Supabase for integrated authentication and user management.

Need Integrated Storage: Consider Supabase for built-in file storage with transformations.

## Pricing Reference

Free Tier provides 3GB storage and 100 compute hours per month with scale-to-zero idle free.

Pro Tier provides usage-based pricing with additional storage and compute. No charge for scale-to-zero idle time.

Storage pricing is per GB-month. Compute pricing per compute hour-month. Data transfer may apply for egress.

---

Status: Production Ready
Platform: Neon Serverless PostgreSQL
Version: 2.1.0
