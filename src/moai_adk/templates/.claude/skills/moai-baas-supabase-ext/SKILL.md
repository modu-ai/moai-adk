---
name: "moai-baas-supabase-ext"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-12
status: stable
description: Enterprise Supabase PostgreSQL Platform with AI-powered architecture, Context7 integration, and intelligent open-source BaaS orchestration for scalable production applications
keywords: ['supabase', 'postgresql', 'rls', 'row-level-security', 'realtime', 'edge-functions', 'enterprise-integration', 'context7-integration', 'ai-orchestration', 'production-deployment']
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

# Enterprise Supabase Platform Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-supabase-ext |
| **Version** | 4.0.0 (2025-11-11) |
| **Tier** | Enterprise Platform Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when Supabase keywords detected |

---

## What It Does

Enterprise Supabase PostgreSQL Platform expert with AI-powered architecture optimization, Context7 integration, and intelligent open-source BaaS orchestration for scalable production applications.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Supabase Architecture** using Context7 MCP for latest PostgreSQL documentation
- 📊 **Intelligent PostgreSQL Optimization** with automated performance tuning recommendations
- 🚀 **Real-time Performance Analytics** with AI-driven database optimization insights
- 🔗 **Enterprise Open-Source Integration** with zero-configuration PostgreSQL extensions
- 📈 **Predictive Cost Analysis** with usage forecasting and resource optimization
- 🔍 **Advanced Security Implementation** with automated RLS policy generation and compliance
- 🌐 **Multi-Region PostgreSQL Deployment** with intelligent replication and failover
- 🎯 **Intelligent Migration Planning** with PostgreSQL migration strategies and automation
- 📱 **Real-time Database Monitoring** with AI-powered performance alerting
- ⚡ **Zero-Configuration Setup** with intelligent template matching and deployment

---

## When to Use

**Automatic triggers**:
- Supabase project architecture and optimization discussions
- PostgreSQL database design and performance optimization
- Row Level Security (RLS) policy implementation and auditing
- Real-time subscription architecture and scaling strategies
- Edge Functions development and deployment patterns

**Manual invocation**:
- Designing enterprise Supabase architectures with PostgreSQL
- Optimizing PostgreSQL performance and RLS policies
- Planning PostgreSQL to Supabase migrations
- Implementing advanced real-time features and subscriptions
- Scaling Supabase applications for enterprise use
- Integrating Supabase with PostgreSQL extensions and tools

---

## Enterprise Supabase Architecture Intelligence

### AI-Enhanced Platform Analysis

#### 1. **Supabase PostgreSQL Enterprise** (Database with AI Optimization)
```yaml
supabase_postgresql_enterprise:
  context7_integration: true
  latest_features:
    - "PostgreSQL 16+ with latest extensions"
    - "pgvector for vector similarity search"
    - "Advanced RLS with policy conditions"
    - "Real-time subscriptions with WebSockets"
    - "Database branching for development workflows"
    - "Point-in-time recovery with 30-day retention"
    - "Connection pooling with PgBouncer"
    - "Read replicas for read scaling"
  
  ai_recommendations:
    best_for: ["PostgreSQL-centric apps", "Real-time features", "Open-source stack"]
    use_cases: ["SaaS platforms", "Real-time dashboards", "API-first applications"]
    performance_metrics:
      transaction_throughput: "50k+ TPS"
      query_latency: "P95 < 50ms"
      concurrent_connections: "10k+ with pooling"
      real_time_latency: "P95 < 100ms"
    
  enterprise_features:
    security: ["Row-Level Security", "Database encryption", "Audit logging"]
    monitoring: ["Query performance insights", "Real-time metrics", "Connection analytics"]
    deployment: ["Multi-region deployment", "Automated backups", "Database branching"]
```

#### 2. **Supabase Edge Functions Enterprise** (Serverless TypeScript Functions)
```yaml
edge_functions_enterprise:
  context7_integration: true
  latest_features:
    - "Deno runtime with TypeScript 5.7+"
    - "Edge deployment across 28+ regions"
    - "Automatic CORS and security headers"
    - "Environment variable management"
    - "Function versioning and rollback"
    - "Integrated authentication middleware"
    - "Database connection pooling"
    - "Real-time WebSocket support"
  
  ai_recommendations:
    best_for: ["API endpoints", "Data processing", "Real-time features"]
    use_cases: ["Webhook handlers", "Data transformation", "Authentication flows"]
    performance_metrics:
      cold_start: "< 50ms average"
      throughput: "10k+ requests/second"
      scalability: "Auto-scaling to 1000+ instances"
      timeout: "Up to 25 seconds"
    
  enterprise_features:
    security: ["JWT authentication", "Request validation", "Rate limiting"]
    monitoring: ["Real-time logs", "Error tracking", "Performance metrics"]
    deployment: ["Git integration", "Environment promotion", "A/B testing"]
```

#### 3. **Supabase Auth Enterprise** (Authentication & Authorization)
```yaml
supabase_auth_enterprise:
  context7_integration: true
  latest_features:
    - "Multi-provider authentication (50+ providers)"
    - "Advanced JWT with custom claims"
    - "Row Level Security integration"
    - "Multi-factor authentication (MFA)"
    - "SSO with SAML 2.0 and OIDC"
    - "Passwordless authentication options"
    - "Advanced user management"
    - "Custom auth flows with hooks"
  
  ai_recommendations:
    best_for: ["User authentication", "Authorization", "Security compliance"]
    use_cases: ["SaaS platforms", "Enterprise applications", "Multi-tenant systems"]
    performance_metrics:
      authentication_latency: "P95 < 300ms"
      concurrent_sessions: "1M+ active users"
      mfa_methods: ["TOTP", "SMS", "Email", "WebAuthn"]
      sso_protocols: ["SAML 2.0", "OpenID Connect", "CAS"]
    
  enterprise_features:
    security: ["MFA enforcement", "SSO integration", "Advanced RLS policies"]
    monitoring: ["Authentication events", "Security alerts", "User activity analytics"]
    deployment: ["Multi-tenant support", "Custom domains", "Branding options"]
```

---

## AI-Powered Supabase Intelligence

### Intelligent PostgreSQL Optimization
```python
# AI-powered Supabase PostgreSQL optimization with Context7
class EnterpriseSupabaseOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.postgresql_analyzer = PostgreSQLAnalyzer()
        self.rls_optimizer = RLSOptimizer()
    
    async def optimize_supabase_architecture(self, 
                                           current_config: SupabaseConfig,
                                           performance_goals: PerformanceGoals) -> OptimizationPlan:
        """Optimize Supabase architecture using AI analysis."""
        
        # Get latest Supabase and PostgreSQL best practices via Context7
        supabase_docs = {}
        services = ['postgresql', 'auth', 'storage', 'realtime', 'edge-functions']
        
        for service in services:
            docs = await self.context7_client.get_library_docs(
                context7_library_id=await self._resolve_supabase_library(service),
                topic=f"enterprise optimization best practices 2025",
                tokens=3000
            )
            supabase_docs[service] = docs
        
        # Get PostgreSQL specific documentation
        pg_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_postgresql_library(),
            topic="performance tuning optimization indexes 2025",
            tokens=3000
        )
        
        # Analyze current configuration
        config_analysis = self._analyze_current_config(current_config, supabase_docs, pg_docs)
        
        # PostgreSQL optimization recommendations
        postgresql_recommendations = self.postgresql_analyzer.generate_recommendations(
            current_config.database,
            performance_goals,
            supabase_docs['postgresql'],
            pg_docs
        )
        
        # RLS optimization recommendations
        rls_recommendations = self.rls_optimizer.optimize_policies(
            current_config.auth,
            supabase_docs['auth']
        )
        
        # Generate comprehensive optimization plan
        return OptimizationPlan(
            database_optimizations=postgresql_recommendations,
            rls_improvements=rls_recommendations,
            auth_enhancements=self._optimize_auth_configuration(current_config.auth, supabase_docs['auth']),
            storage_optimizations=self._optimize_storage_configuration(current_config.storage, supabase_docs['storage']),
            realtime_improvements=self._optimize_realtime_configuration(current_config.realtime, supabase_docs['realtime']),
            expected_improvements=self._calculate_expected_improvements(
                postgresql_recommendations,
                rls_recommendations
            ),
            implementation_complexity=self._assess_implementation_complexity(
                postgresql_recommendations,
                rls_recommendations
            ),
            roi_projection=self._calculate_roi_projection(
                postgresql_recommendations,
                performance_goals
            )
        )
    
    def _optimize_postgresql_performance(self, 
                                       config: DatabaseConfig,
                                       supabase_docs: dict,
                                       pg_docs: dict) -> List[DatabaseOptimization]:
        """Generate PostgreSQL-specific performance optimizations."""
        optimizations = []
        
        # Query optimization
        query_recommendations = self._analyze_query_patterns(config, pg_docs)
        optimizations.extend(query_recommendations)
        
        # Index optimization
        index_recommendations = self._optimize_database_indexes(config, pg_docs)
        optimizations.extend(index_recommendations)
        
        # Connection and pooling optimization
        connection_recommendations = self._optimize_connections(config, supabase_docs['postgresql'])
        optimizations.extend(connection_recommendations)
        
        # Extension utilization
        extension_recommendations = self._optimize_extensions(config, pg_docs)
        optimizations.extend(extension_recommendations)
        
        return optimizations
    
    def _generate_rls_policies(self, 
                              schema: DatabaseSchema,
                              requirements: SecurityRequirements) -> List[RLSPolicy]:
        """Generate Row Level Security policies using AI analysis."""
        policies = []
        
        for table in schema.tables:
            # Analyze table access patterns
            access_patterns = self._analyze_table_access_patterns(table)
            
            # Generate base RLS policies
            base_policies = self._generate_base_rls_policies(table, access_patterns)
            policies.extend(base_policies)
            
            # Generate role-specific policies
            role_policies = self._generate_role_specific_policies(table, requirements)
            policies.extend(role_policies)
            
            # Generate time-based policies if needed
            if requirements.temporal_access:
                temporal_policies = self._generate_temporal_policies(table, requirements)
                policies.extend(temporal_policies)
        
        return policies
```

### Context7-Enhanced Supabase Intelligence
```python
# Real-time Supabase intelligence with Context7
class Context7SupabaseIntelligence:
    def __init__(self):
        self.context7_client = Context7Client()
        self.supabase_monitor = SupabaseMonitor()
        self.update_scheduler = UpdateScheduler()
    
    async def get_real_time_supabase_updates(self, services: List[str]) -> SupabaseUpdates:
        """Get real-time Supabase updates via Context7."""
        updates = {}
        
        for service in services:
            # Get latest Supabase documentation updates
            latest_docs = await self.context7_client.get_library_docs(
                context7_library_id=await self._resolve_supabase_library(service),
                topic="latest features updates deprecation warnings 2025",
                tokens=2500
            )
            
            # Analyze updates for impact
            impact_analysis = self._analyze_update_impact(latest_docs)
            
            updates[service] = SupabaseUpdate(
                new_features=self._extract_new_features(latest_docs),
                breaking_changes=self._extract_breaking_changes(latest_docs),
                performance_improvements=self._extract_performance_improvements(latest_docs),
                security_updates=self._extract_security_updates(latest_docs),
                deprecation_warnings=self._extract_deprecation_warnings(latest_docs),
                impact_assessment=impact_analysis,
                recommended_actions=self._generate_recommendations(latest_docs)
            )
        
        return SupabaseUpdates(updates)
    
    async def optimize_database_configuration(self, 
                                             current_config: DatabaseConfig,
                                             performance_goals: PerformanceGoals) -> DatabaseOptimization:
        """Optimize Supabase database configuration using AI analysis."""
        
        # Get PostgreSQL optimization best practices
        pg_best_practices = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_postgresql_library(),
            topic="performance tuning configuration optimization 2025",
            tokens=3000
        )
        
        # Get Supabase specific optimizations
        supabase_best_practices = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_supabase_library('postgresql'),
            topic="database optimization tuning connection pooling 2025",
            tokens=2500
        )
        
        # Analyze current configuration
        config_analysis = self._analyze_database_configuration(
            current_config,
            pg_best_practices,
            supabase_best_practices
        )
        
        # Generate optimization recommendations
        optimizations = self._generate_database_optimizations(
            config_analysis,
            performance_goals,
            pg_best_practices,
            supabase_best_practices
        )
        
        return DatabaseOptimization(
            configuration_changes=optimizations.config_changes,
            index_recommendations=optimizations.index_improvements,
            query_optimizations=optimizations.query_improvements,
            connection_tuning=optimizations.connection_optimizations,
            extension_recommendations=optimizations.extension_suggestions,
            expected_improvements=optimizations.performance_gains,
            implementation_complexity=optimizations.complexity_score,
            rollback_plan=optimizations.rollback_strategy
        )
```

---

## Advanced Supabase Integration Patterns

### Enterprise Multi-Service Architecture
```yaml
# Enterprise Supabase multi-service architecture
enterprise_supabase_architecture:
  microservices_pattern:
    - name: "User Authentication Service"
      services: ["Supabase Auth", "PostgreSQL RLS"]
      features: ["Multi-provider auth", "MFA", "SSO", "User roles"]
      integration: "Supabase Auth SDK + RLS policies"
    
    - name: "Real-time Data Service"
      services: ["PostgreSQL", "Supabase Realtime"]
      features: ["Real-time subscriptions", "Change data capture", "Live queries"]
      integration: "PostgreSQL + Realtime subscriptions"
    
    - name: "Serverless Processing Service"
      services: ["Edge Functions", "PostgreSQL"]
      features: ["API endpoints", "Data processing", "Webhook handlers"]
      integration: "Edge Functions + Database access"
    
    - name: "File Storage Service"
      services: ["Supabase Storage", "PostgreSQL"]
      features: ["Object storage", "Image transformation", "CDN delivery"]
      integration: "Storage SDK + Database metadata"
    
    - name: "API Management Service"
      services: ["PostgreSQL", "PostgREST", "Edge Functions"]
      features: ["REST API generation", "GraphQL support", "Rate limiting"]
      integration: "PostgREST + Custom middleware"

  hybrid_strategy:
    development: "Local Docker + Supabase CLI"
    staging: "Supabase branch with production-like data"
    production: "Multi-region Supabase enterprise deployment"
    
  disaster_recovery:
    backup_strategy: "Automated continuous backups + Point-in-time recovery"
    failover_mechanism: "Read replicas + Automated failover"
    data_replication: "Streaming replication across regions"
```

### AI-Driven PostgreSQL Migration
```python
# Intelligent PostgreSQL to Supabase migration planning
class PostgreSQLToSupabaseMigrationPlanner:
    def __init__(self):
        self.context7_client = Context7Client()
        self.postgresql_analyzer = PostgreSQLAnalyzer()
        self.migration_engine = MigrationEngine()
    
    async def plan_postgresql_to_supabase_migration(self, 
                                                   postgresql_db: PostgreSQLDatabase,
                                                   target_supabase_features: List[str]) -> MigrationPlan:
        """Plan PostgreSQL to Supabase migration with AI-driven analysis."""
        
        # Get Supabase migration documentation
        supabase_docs = {}
        for feature in target_supabase_features:
            docs = await self.context7_client.get_library_docs(
                context7_library_id=await self._resolve_supabase_library(feature),
                topic=f"migration best practices enterprise patterns 2025",
                tokens=3000
            )
            supabase_docs[feature] = docs
        
        # Analyze current PostgreSQL database
        postgresql_analysis = self.postgresql_analyzer.analyze_database(postgresql_db)
        
        # Assess migration complexity
        complexity_analysis = self._assess_migration_complexity(
            postgresql_analysis,
            target_supabase_features,
            supabase_docs
        )
        
        # Generate migration roadmap
        roadmap = self._generate_migration_roadmap(
            postgresql_analysis,
            target_supabase_features,
            complexity_analysis
        )
        
        return MigrationPlan(
            phases=roadmap.phases,
            migration_strategies={
                'schema_migration': self._plan_schema_migration(postgresql_analysis, supabase_docs['postgresql']),
                'data_migration': self._plan_data_migration(postgresql_analysis, supabase_docs['postgresql']),
                'rls_migration': self._plan_rls_migration(postgresql_analysis, supabase_docs['auth']),
                'auth_integration': self._plan_auth_integration(postgresql_analysis, supabase_docs['auth']),
                'realtime_setup': self._plan_realtime_setup(postgresql_analysis, supabase_docs['realtime'])
            },
            timeline_estimate=self._calculate_timeline(roadmap),
            cost_projection=self._calculate_migration_costs(roadmap),
            risk_mitigation=complexity_analysis.mitigation_strategies,
            rollback_strategy=self._generate_rollback_plan(roadmap)
        )
    
    def _generate_migration_phases(self, 
                                 postgresql_analysis: PostgreSQLAnalysis,
                                 supabase_features: List[str],
                                 complexity: ComplexityAnalysis) -> List[MigrationPhase]:
        """Generate detailed migration phases."""
        phases = []
        
        # Phase 1: Assessment and Preparation
        phases.append(MigrationPhase(
            name="Assessment and Preparation",
            duration="2-3 weeks",
            activities=[
                "PostgreSQL database schema analysis",
                "Dependency mapping and compatibility check",
                "Performance baseline establishment",
                "Security and compliance assessment"
            ],
            deliverables=[
                "Database compatibility report",
                "Migration strategy document",
                "Risk assessment matrix"
            ],
            success_criteria=[
                "All database objects catalogued",
                "Compatibility issues identified",
                "Migration approach defined"
            ]
        ))
        
        # Phase 2: Supabase Project Setup
        phases.append(MigrationPhase(
            name="Supabase Project Setup",
            duration="1-2 weeks",
            activities=[
                "Create Supabase project with appropriate configuration",
                "Setup database extensions and configurations",
                "Implement RLS policies structure",
                "Configure authentication providers"
            ],
            deliverables=[
                "Configured Supabase project",
                "Database schema structure",
                "Authentication configuration"
            ],
            success_criteria=[
                "Supabase project provisioned",
                "Extensions installed",
                "Auth providers configured"
            ]
        ))
        
        # Phase 3: Schema and Data Migration
        phases.append(MigrationPhase(
            name="Schema and Data Migration",
            duration="3-6 weeks",
            activities=[
                "Migrate database schema to Supabase",
                "Implement Row Level Security policies",
                "Migrate data with validation",
                "Setup real-time subscriptions"
            ],
            deliverables=[
                "Migrated database schema",
                "RLS policies implemented",
                "Migrated and validated data"
            ],
            success_criteria=[
                "Schema successfully migrated",
                "RLS policies working correctly",
                "Data integrity verified"
            ]
        ))
        
        # Phase 4: Integration and Testing
        phases.append(MigrationPhase(
            name="Integration and Testing",
            duration="2-4 weeks",
            activities=[
                "Update application connection strings",
                "Implement Supabase client SDKs",
            ],
            deliverables=[
                "Updated application code",
                "Comprehensive test results",
                "Performance validation report"
            ],
            success_criteria=[
                "All tests passing",
                "Performance targets met",
                "Security controls validated"
            ]
        ))
        
        return phases
```

---

## Performance Optimization Intelligence

### Real-Time PostgreSQL Monitoring
```python
# AI-powered PostgreSQL performance monitoring and optimization
class SupabasePerformanceIntelligence:
    def __init__(self):
        self.context7_client = Context7Client()
        self.postgresql_monitor = PostgreSQLMonitor()
        self.optimization_engine = OptimizationEngine()
    
    async def optimize_database_performance(self, 
                                          current_metrics: DatabaseMetrics) -> OptimizationPlan:
        """Optimize PostgreSQL database performance using AI analysis."""
        
        # Get latest PostgreSQL optimization strategies
        optimization_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_postgresql_library(),
            topic="performance optimization query tuning indexing 2025",
            tokens=3000
        )
        
        # Get Supabase specific optimizations
        supabase_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_supabase_library('postgresql'),
            topic="connection pooling performance tuning 2025",
            tokens=2000
        )
        
        # Analyze current performance
        performance_analysis = self.postgresql_monitor.analyze_performance(
            current_metrics,
            optimization_docs,
            supabase_docs
        )
        
        # Identify optimization opportunities
        optimization_opportunities = self.optimization_engine.identify_opportunities(
            performance_analysis,
            optimization_docs,
            supabase_docs
        )
        
        # Generate optimization plan
        return OptimizationPlan(
            query_optimizations=optimization_opportunities.query_improvements,
            index_recommendations=optimization_opportunities.index_improvements,
            configuration_tuning=optimization_opportunities.config_changes,
            connection_optimizations=optimization_opportunities.connection_improvements,
            expected_improvements=self._calculate_expected_improvements(
                optimization_opportunities
            ),
            implementation_complexity=optimization_opportunities.complexity_analysis,
            roi_projections=self._calculate_roi_projections(optimization_opportunities)
        )
    
    def setup_intelligent_monitoring(self) -> DatabaseMonitoringSetup:
        """Setup intelligent PostgreSQL monitoring."""
        return DatabaseMonitoringSetup(
            real_time_metrics=[
                "Query execution time distribution",
                "Connection pool utilization",
                "Index usage statistics",
                "Table and index bloat analysis",
                "Lock wait times and deadlocks",
                "Cache hit ratios (buffer cache, index cache)",
                "Autovacuum activity and effectiveness",
                "WAL generation and replication lag"
            ],
            ai_alerts=[
                "Slow query detection and analysis",
                "Connection pool exhaustion alerts",
                "Index usage optimization opportunities",
                "Table bloat warnings and recommendations",
                "Unusual lock wait patterns",
                "Performance regression detection",
                "Resource utilization anomalies"
            ],
            dashboards=self._create_intelligent_dashboards(),
            notification_channels=self._setup_notification_channels(),
            automated_responses=self._configure_automated_responses()
        )
    
    def _optimize_rls_policies(self, 
                              current_policies: List[RLSPolicy],
                              performance_requirements: PerformanceRequirements) -> RLSPolicyOptimization:
        """Optimize Row Level Security policies for performance."""
        return RLSPolicyOptimization(
            policy_simplification=self._simplify_complex_policies(current_policies),
            index_recommendations=self._recommend_rls_indexes(current_policies),
            query_optimization=self._optimize_rls_queries(current_policies),
            caching_strategies=self._implement_rls_caching(current_policies),
            expected_improvements=self._calculate_rls_improvements(current_policies)
        )
```

---

## API Reference

### Core Functions
- `optimize_supabase_architecture(config, goals)` - AI-powered Supabase optimization
- `plan_postgresql_to_supabase_migration(database, features)` - Intelligent migration planning
- `optimize_database_performance(metrics)` - Real-time PostgreSQL optimization
- `get_real_time_supabase_updates(services)` - Context7 update monitoring
- `setup_intelligent_monitoring()` - AI-powered database monitoring setup
- `generate_rls_policies(schema, requirements)` - Automated RLS policy generation

### Context7 Integration
- `get_latest_supabase_documentation(service)` - Official docs via Context7
- `analyze_supabase_updates(services)` - Real-time update analysis
- `optimize_with_postgresql_best_practices()` - Latest PostgreSQL optimization strategies

### Data Structures
- `SupabaseConfig` - Comprehensive Supabase service configuration
- `DatabaseOptimization` - PostgreSQL performance optimization recommendations
- `MigrationPlan` - Detailed PostgreSQL to Supabase migration strategy
- `RLSPolicy` - Row Level Security policy definition and optimization
- `PerformanceIntelligence` - Real-time PostgreSQL monitoring and alerting

---

## Changelog

- **v4.0.0** (2025-11-12): Complete Enterprise v4.0 rewrite with Context7 integration, Supabase v2.24.0, CLI v2.58.5, pgvector vector search, edge functions dashboard, dashboard improvements integration, AI-powered Supabase optimization, enterprise PostgreSQL patterns, and intelligent migration planning
- **v2.0.0** (2025-11-09): Supabase Advanced Guide with RLS and production best practices
- **v1.0.0** (2025-11-09): Initial Supabase platform integration

---

## Works Well With

- `moai-baas-foundation` (BaaS platform selection and architecture)
- `moai-lang-postgresql` (PostgreSQL database optimization and patterns)
- `moai-foundation-trust` (security and compliance validation)
- `moai-essentials-perf` (performance optimization and monitoring)
- `moai-essentials-migration` (PostgreSQL migration automation)
- Context7 MCP (real-time Supabase and PostgreSQL documentation)

---

## Best Practices

✅ **DO**:
- Use Context7 integration for latest Supabase and PostgreSQL documentation
- Implement comprehensive Row Level Security policies for data protection
- Plan PostgreSQL migrations with detailed dependency and compatibility analysis
- Use AI-powered optimization recommendations for database performance
- Monitor database performance with real-time query analysis
- Implement connection pooling and read replicas for scalability
- Use database branching for development and testing workflows
- Establish clear backup and disaster recovery strategies

❌ **DON'T**:
- Skip Row Level Security policy implementation for sensitive data
- Ignore PostgreSQL performance tuning and optimization opportunities
- Underestimate migration complexity from standalone PostgreSQL
- Skip real-time performance monitoring and alerting
- Use inappropriate data types or missing indexes
- Neglect connection pooling configuration for high-traffic applications
- Ignore backup and point-in-time recovery setup
- Skip security audit and compliance validation
