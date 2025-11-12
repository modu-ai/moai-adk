---
name: moai-baas-vercel-ext
version: 4.0.0
created: 2025-11-11
updated: 2025-11-12
status: active
description: Enterprise Vercel Edge Platform with AI-powered deployment architecture, Context7 integration, and intelligent edge computing orchestration for scalable global applications
keywords: ['vercel', 'edge-functions', 'next.js', 'deployment', 'serverless', 'edge-computing', 'global-cdn', 'context7-integration', 'ai-orchestration', 'production-deployment']
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

# Enterprise Vercel Edge Platform Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-vercel-ext |
| **Version** | 4.0.0 (2025-11-11) |
| **Tier** | Enterprise Edge Platform Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when Vercel keywords detected |

---

## What It Does

Enterprise Vercel Edge Platform expert with AI-powered deployment architecture, Context7 integration, and intelligent edge computing orchestration for scalable global applications.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Vercel Architecture** using Context7 MCP for latest Vercel documentation
- 📊 **Intelligent Edge Optimization** with automated global deployment strategies
- 🚀 **Real-time Performance Analytics** with AI-driven edge computing insights
- 🔗 **Enterprise Next.js Integration** with optimized framework patterns and deployment
- 📈 **Predictive Cost Analysis** with usage forecasting and bandwidth optimization
- 🔍 **Advanced Edge Security** with automated DDoS protection and compliance patterns
- 🌐 **Global CDN Intelligence** with intelligent latency and cache optimization
- 🎯 **Intelligent Migration Planning** with modern deployment strategies and automation
- 📱 **Real-time Edge Monitoring** with AI-powered performance alerting
- ⚡ **Zero-Configuration Deployment** with intelligent template matching and CI/CD

---

## When to Use

**Automatic triggers**:
- Vercel deployment architecture and optimization discussions
- Next.js application deployment and edge computing strategies
- Global CDN and edge function performance optimization
- Serverless architecture and scalability planning
- Modern web application deployment and CI/CD optimization

**Manual invocation**:
- Designing enterprise Vercel architectures with edge computing
- Optimizing Next.js performance and edge deployment strategies
- Planning global application deployment and CDN optimization
- Implementing advanced edge functions and serverless patterns
- Migrating applications to Vercel with zero-downtime strategies
- Optimizing Vercel costs and performance monitoring

---

## Enterprise Vercel Architecture Intelligence

### AI-Enhanced Platform Analysis

#### 1. **Vercel Edge Network** (Global CDN and Edge Computing)
```yaml
vercel_edge_network:
  context7_integration: true
  latest_features:
    - "Global edge network with 28+ POPs worldwide"
    - "Edge Functions with V8 isolate runtime"
    - "Incremental Static Regeneration (ISR) with smart invalidation"
    - "Edge Middleware for request/response transformation"
    - "Advanced caching strategies with stale-while-revalidate"
    - "Image optimization with WebP/AVIF format support"
    - "Edge analytics with real-time performance insights"
    - "Geographic routing and intelligent load balancing"
  
  ai_recommendations:
    best_for: ["Global applications", "Performance-critical sites", "Next.js apps"]
    use_cases: ["E-commerce platforms", "Content websites", "API gateways", "Global SaaS"]
    performance_metrics:
      edge_latency: "P95 < 100ms globally"
      cache_hit_rate: "95%+ for static assets"
      throughput: "10Tbps+ global network capacity"
      availability: "99.99% uptime SLA"
    
  enterprise_features:
    performance: ["Edge caching", "Smart invalidation", "Image optimization", "Compression"]
    security: ["DDoS protection", "WAF", "Edge authentication", "SSL/TLS"]
    monitoring: ["Real-time analytics", "Performance insights", "Error tracking"]
```

#### 2. **Vercel Edge Functions** (Serverless Computing at the Edge)
```yaml
vercel_edge_functions:
  context7_integration: true
  latest_features:
    - "V8 isolate runtime with 50ms cold starts"
    - "Node.js, TypeScript, and Go support"
    - "Edge Middleware for request/response processing"
    - "Environment variable management with edge secrets"
    - "Function versioning and gradual rollouts"
    - "Real-time logs and performance monitoring"
    - "Edge-to-origin connectivity optimization"
    - "Function bundling with tree-shaking optimization"
  
  ai_recommendations:
    best_for: ["API endpoints", "Dynamic content", "Request routing", "A/B testing"]
    use_cases: ["API gateways", "Dynamic routing", "Content personalization", "Geolocation"]
    performance_metrics:
      cold_start: "P95 < 50ms"
      execution_time: "P95 < 500ms"
      throughput: "1000+ requests/second per function"
      memory_limit: "Up to 1GB per function"
    
  enterprise_features:
    performance: ["Cold start optimization", "Memory management", "Function bundling"]
    security: ["Edge secrets", "Request validation", "CORS handling", "Rate limiting"]
    monitoring: ["Real-time logs", "Performance metrics", "Error analytics"]
```

#### 3. **Vercel Frontend Cloud** (Modern Frontend Deployment)
```yaml
vercel_frontend_cloud:
  context7_integration: true
  latest_features:
    - "Git-based deployments with automatic previews"
    - "Preview deployments for every branch/PR"
    - "Environment promotion with deployment aliases"
    - "Automatic dependency optimization and caching"
    - "Smart builds with incremental builds"
    - "Build failure analysis and debugging"
    - "Deployment analytics and insights"
    - "Custom domains with automatic SSL"
  
  ai_recommendations:
    best_for: ["React/Next.js apps", "Static sites", "JAMstack applications"]
    use_cases: ["Modern web apps", "Marketing sites", "Documentation portals", "Dashboards"]
    performance_metrics:
      build_time: "50% faster than traditional CI/CD"
      deployment_time: "Under 30 seconds for most apps"
      preview_generation: "Automatic for every push"
      rollback_time: "Instant rollback to previous deployments"
    
  enterprise_features:
    development: ["Preview deployments", "Hot reloading", "Build debugging"]
    deployment: ["Git integration", "Environment promotion", "Custom domains"]
    collaboration: ["Team permissions", "Deployment comments", "Slack notifications"]
```

---

## AI-Powered Vercel Intelligence

### Intelligent Edge Optimization
```python
# AI-powered Vercel edge optimization with Context7
class EnterpriseVercelOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.vercel_analyzer = VercelAnalyzer()
        self.edge_optimizer = EdgeOptimizer()
    
    async def optimize_vercel_architecture(self, 
                                         current_config: VercelConfig,
                                         performance_goals: PerformanceGoals) -> OptimizationPlan:
        """Optimize Vercel architecture using AI analysis."""
        
        # Get latest Vercel best practices via Context7
        vercel_docs = {}
        services = ['edge-network', 'edge-functions', 'frontend-cloud', 'analytics', 'security']
        
        for service in services:
            docs = await self.context7_client.get_library_docs(
                context7_library_id=await self._resolve_vercel_library(service),
                topic=f"enterprise optimization best practices 2025",
                tokens=3000
            )
            vercel_docs[service] = docs
        
        # Get Next.js optimization documentation
        nextjs_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_nextjs_library(),
            topic="performance optimization deployment patterns 2025",
            tokens=3000
        )
        
        # Analyze current configuration
        config_analysis = self._analyze_current_config(current_config, vercel_docs, nextjs_docs)
        
        # Edge optimization recommendations
        edge_recommendations = self.edge_optimizer.optimize_edge_performance(
            current_config,
            performance_goals,
            vercel_docs['edge-network']
        )
        
        # Function optimization recommendations
        function_recommendations = self.vercel_analyzer.optimize_functions(
            current_config.functions,
            vercel_docs['edge-functions']
        )
        
        # Generate comprehensive optimization plan
        return OptimizationPlan(
            edge_optimizations=edge_recommendations,
            function_optimizations=function_recommendations,
            frontend_optimizations=self._optimize_frontend_deployment(
                current_config.frontend,
                vercel_docs['frontend-cloud'],
                nextjs_docs
            ),
            caching_strategies=self._optimize_caching_configuration(
                current_config.caching,
                vercel_docs['edge-network']
            ),
            security_enhancements=self._optimize_security_configuration(
                current_config.security,
                vercel_docs['security']
            ),
            expected_improvements=self._calculate_expected_improvements(
                edge_recommendations,
                function_recommendations
            ),
            implementation_complexity=self._assess_implementation_complexity(
                edge_recommendations,
                function_recommendations
            ),
            roi_projection=self._calculate_roi_projection(
                edge_recommendations,
                performance_goals
            )
        )
    
    def _optimize_edge_caching(self, 
                             config: EdgeConfig,
                             performance_requirements: PerformanceRequirements) -> List[CachingOptimization]:
        """Generate edge caching optimizations."""
        optimizations = []
        
        # Static asset caching
        static_optimizations = self._optimize_static_caching(config, performance_requirements)
        optimizations.extend(static_optimizations)
        
        # API response caching
        api_optimizations = self._optimize_api_caching(config, performance_requirements)
        optimizations.extend(api_optimizations)
        
        # Image optimization caching
        image_optimizations = self._optimize_image_caching(config, performance_requirements)
        optimizations.extend(image_optimizations)
        
        # Dynamic content caching
        dynamic_optimizations = self._optimize_dynamic_caching(config, performance_requirements)
        optimizations.extend(dynamic_optimizations)
        
        return optimizations
    
    def _optimize_edge_functions(self, 
                               config: FunctionConfig,
                                performance_requirements: PerformanceRequirements) -> List[FunctionOptimization]:
        """Generate edge function optimizations."""
        optimizations = []
        
        # Cold start optimization
        coldstart_optimizations = self._optimize_cold_starts(config, performance_requirements)
        optimizations.extend(coldstart_optimizations)
        
        # Memory and performance optimization
        performance_optimizations = self._optimize_function_performance(config, performance_requirements)
        optimizations.extend(performance_optimizations)
        
        # Geographic optimization
        geo_optimizations = self._optimize_geographic_distribution(config, performance_requirements)
        optimizations.extend(geo_optimizations)
        
        return optimizations
```

### Context7-Enhanced Vercel Intelligence
```python
# Real-time Vercel intelligence with Context7
class Context7VercelIntelligence:
    def __init__(self):
        self.context7_client = Context7Client()
        self.vercel_monitor = VercelMonitor()
        self.update_scheduler = UpdateScheduler()
    
    async def get_real_time_vercel_updates(self, services: List[str]) -> VercelUpdates:
        """Get real-time Vercel updates via Context7."""
        updates = {}
        
        for service in services:
            # Get latest Vercel documentation updates
            latest_docs = await self.context7_client.get_library_docs(
                context7_library_id=await self._resolve_vercel_library(service),
                topic="latest features updates deprecation warnings 2025",
                tokens=2500
            )
            
            # Analyze updates for impact
            impact_analysis = self._analyze_update_impact(latest_docs)
            
            updates[service] = VercelUpdate(
                new_features=self._extract_new_features(latest_docs),
                breaking_changes=self._extract_breaking_changes(latest_docs),
                performance_improvements=self._extract_performance_improvements(latest_docs),
                security_updates=self._extract_security_updates(latest_docs),
                deprecation_warnings=self._extract_deprecation_warnings(latest_docs),
                impact_assessment=impact_analysis,
                recommended_actions=self._generate_recommendations(latest_docs)
            )
        
        return VercelUpdates(updates)
    
    async def optimize_global_performance(self, 
                                        current_config: GlobalConfig,
                                        traffic_patterns: TrafficPatterns) -> GlobalOptimization:
        """Optimize global Vercel performance using AI analysis."""
        
        # Get global optimization documentation
        global_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_vercel_library('edge-network'),
            topic="global optimization geographic distribution latency 2025",
            tokens=3000
        )
        
        # Get CDN optimization strategies
        cdn_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_vercel_library('edge-network'),
            topic="CDN optimization caching strategies edge computing 2025",
            tokens=2500
        )
        
        # Analyze current global performance
        global_analysis = self._analyze_global_performance(
            current_config,
            traffic_patterns,
            global_docs,
            cdn_docs
        )
        
        # Generate global optimization recommendations
        optimizations = self._generate_global_optimizations(
            global_analysis,
            traffic_patterns,
            global_docs,
            cdn_docs
        )
        
        return GlobalOptimization(
            geographic_optimizations=optimizations.geo_improvements,
            cdn_strategies=optimizations.cdn_improvements,
            edge_function_distribution=optimizations.function_distribution,
            cache_configurations=optimizations.cache_optimizations,
            routing_optimizations=optimizations.routing_improvements,
            expected_improvements=self._calculate_global_improvements(optimizations),
            implementation_complexity=optimizations.complexity_score,
            rollout_strategy=optimizations.rollout_plan
        )
```

---

## Advanced Vercel Integration Patterns

### Enterprise Edge Computing Architecture
```yaml
# Enterprise Vercel edge computing architecture
enterprise_vercel_architecture:
  edge_patterns:
    - name: "Global E-commerce Platform"
      features: ["Edge functions", "ISR", "Image optimization", "Geolocation routing"]
      optimization: "Regional inventory + Dynamic pricing + Localized content"
      integration: "Next.js + Edge Functions + Edge Middleware"
    
    - name: "Content Delivery Platform"
      features: ["Static site generation", "CDN optimization", "Image transformation"]
      optimization: "Smart caching + Progressive loading + WebP/AVIF conversion"
      integration: "Next.js + Vercel Image API + Edge caching"
    
    - name: "API Gateway Platform"
      features: ["Edge functions", "Rate limiting", "Authentication", "Load balancing"]
      optimization: "Geographic routing + Request routing + Response caching"
      integration: "Edge Functions + Edge Middleware + Backend APIs"
    
    - name: "Real-time Application Platform"
      features: ["Edge functions", "WebSocket support", "Server-sent events"]
      optimization: "Edge connectivity + Real-time sync + Event streaming"
      integration: "Edge Functions + Real-time APIs + Client-side sync"
    
    - name: "Multi-tenant SaaS Platform"
      features: ["Edge routing", "Custom domains", "Tenant isolation", "Analytics"]
      optimization: "Tenant routing + Brand customization + Performance isolation"
      integration: "Edge Middleware + Edge Functions + Multi-tenant databases"

  deployment_strategy:
    development: "Local development + Preview deployments + Hot reloading"
    staging: "Production-like environment + Performance testing + Security validation"
    production: "Global edge deployment + Performance monitoring + Security hardening"
    
  performance_optimization:
    caching_strategy: "Multi-tier caching with smart invalidation"
    image_optimization: "Automatic format conversion + Responsive images + Lazy loading"
    bundle_optimization: "Tree-shaking + Code splitting + Minification + Compression"
    edge_computing: "Geographic distribution + Cold start optimization + Memory management"
```

### AI-Driven Global Performance Optimization
```python
# Intelligent global performance optimization with AI analysis
class VercelGlobalOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.global_analyzer = GlobalAnalyzer()
        self.performance_predictor = PerformancePredictor()
    
    async def optimize_global_application_performance(self, 
                                                    application: VercelApplication,
                                                    global_requirements: GlobalRequirements) -> GlobalOptimizationPlan:
        """Optimize global application performance using AI analysis."""
        
        # Get global performance optimization documentation
        global_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_vercel_library('edge-network'),
            topic="global performance optimization geographic latency 2025",
            tokens=3000
        )
        
        # Get edge computing patterns
        edge_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_vercel_library('edge-functions'),
            topic="edge computing patterns geographic distribution 2025",
            tokens=2500
        )
        
        # Analyze current global performance
        global_analysis = self.global_analyzer.analyze_global_performance(
            application,
            global_requirements,
            global_docs,
            edge_docs
        )
        
        # Predict performance improvements
        performance_predictions = self.performance_predictor.predict_improvements(
            global_analysis,
            global_requirements
        )
        
        # Generate comprehensive optimization plan
        return GlobalOptimizationPlan(
            geographic_optimization={
                'edge_function_placement': self._optimize_edge_function_placement(
                    global_analysis, edge_docs
                ),
                'cache_distribution': self._optimize_cache_distribution(
                    global_analysis, global_docs
                ),
                'routing_strategies': self._optimize_routing_strategies(
                    global_analysis, global_docs
                ),
                'content_localization': self._optimize_content_localization(
                    global_analysis, global_docs
                )
            },
            performance_optimization={
                'cdn_configuration': self._optimize_cdn_configuration(global_analysis, global_docs),
                'image_optimization': self._optimize_image_strategy(global_analysis, global_docs),
                'bundle_optimization': self._optimize_bundle_strategy(global_analysis, edge_docs),
                'edge_computing': self._optimize_edge_computing(global_analysis, edge_docs)
            },
            cost_optimization=self._optimize_global_costs(
                global_analysis,
                performance_predictions
            ),
            expected_improvements=performance_predictions,
            implementation_roadmap=self._generate_implementation_roadmap(
                global_analysis,
                performance_predictions
            ),
            monitoring_strategy=self._design_monitoring_strategy(global_analysis)
        )
    
    def _generate_edge_function_strategies(self, 
                                         traffic_patterns: TrafficPatterns,
                                         performance_requirements: PerformanceRequirements) -> EdgeFunctionStrategy:
        """Generate optimal edge function deployment strategies."""
        return EdgeFunctionStrategy(
            geographic_distribution={
                'primary_regions': self._identify_primary_regions(traffic_patterns),
                'secondary_regions': self._identify_secondary_regions(traffic_patterns),
                'function_replication': self._design_replication_strategy(traffic_patterns),
                'load_balancing': self._design_load_balancing_strategy(traffic_patterns)
            },
            performance_optimization={
                'cold_start_minimization': self._minimize_cold_starts(performance_requirements),
                'memory_optimization': self._optimize_memory_usage(performance_requirements),
                'execution_optimization': self._optimize_execution_time(performance_requirements),
                'bundling_strategy': self._optimize_function_bundling(performance_requirements)
            },
            cost_optimization={
                'invocation_patterns': self._optimize_invocation_patterns(traffic_patterns),
                'compute_allocation': self._optimize_compute_allocation(performance_requirements),
                'data_transfer': self._optimize_data_transfer(traffic_patterns),
                'storage_optimization': self._optimize_storage_usage(traffic_patterns)
            },
            monitoring={
                'performance_metrics': self._define_performance_metrics(),
                'error_tracking': self._setup_error_tracking(),
                'usage_analytics': self._setup_usage_analytics(),
                'cost_monitoring': self._setup_cost_monitoring()
            }
        )
```

---

## Performance and Monitoring Intelligence

### Real-Time Global Performance Monitoring
```python
# AI-powered Vercel global performance monitoring and optimization
class VercelPerformanceIntelligence:
    def __init__(self):
        self.context7_client = Context7Client()
        self.performance_monitor = PerformanceMonitor()
        self.optimization_engine = OptimizationEngine()
    
    async def setup_global_performance_monitoring(self, 
                                                application: VercelApplication) -> GlobalMonitoringSetup:
        """Setup comprehensive Vercel global performance monitoring."""
        
        # Get latest performance monitoring best practices
        monitoring_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_vercel_library('analytics'),
            topic="performance monitoring real-time analytics edge metrics 2025",
            tokens=3000
        )
        
        return GlobalMonitoringSetup(
            real_time_metrics={
                'edge_performance': [
                    "Edge response times by region",
                    "Cache hit ratios and miss patterns",
                    "Edge function execution metrics",
                    "Bandwidth utilization and transfer",
                    "Geographic latency distribution"
                ],
                'user_experience': [
                    "Core Web Vitals (LCP, FID, CLS)",
                    "Time to First Byte (TTFB)",
                    "Largest Contentful Paint (LCP)",
                    "Cumulative Layout Shift (CLS)",
                    "First Input Delay (FID)"
                ],
                'infrastructure_performance': [
                    "Build times and success rates",
                    "Deployment frequency and duration",
                    "Function cold start frequency",
                    "Error rates and recovery times",
                    "API response times and availability"
                ]
            },
            ai_analytics=[
                "Performance anomaly detection",
                "Geographic performance optimization",
                "User experience prediction",
                "Cost optimization recommendations",
                "Traffic pattern analysis"
            ],
            global_insights={
                'regional_performance': "Performance breakdown by geographic region",
                'device_performance': "Performance analysis by device type and capabilities",
                'network_performance': "Performance analysis by network conditions",
                'user_satisfaction': "User satisfaction scores and correlation with performance"
            },
            alerting=self._setup_performance_alerting(),
            dashboards=self._create_performance_dashboards(application),
            reporting=self._configure_performance_reporting()
        )
    
    async def optimize_edge_performance(self, 
                                      current_metrics: EdgeMetrics) -> EdgeOptimization:
        """Optimize edge performance using AI analysis."""
        
        # Get edge optimization documentation
        edge_docs = await self.context7_client.get_library_docs(
            context7_library_id=await self._resolve_vercel_library('edge-network'),
            topic="edge performance optimization caching strategies 2025",
            tokens=2500
        )
        
        # Analyze current edge performance
        performance_analysis = self.performance_monitor.analyze_edge_performance(
            current_metrics,
            edge_docs
        )
        
        # Identify optimization opportunities
        optimization_opportunities = self.optimization_engine.identify_edge_opportunities(
            performance_analysis,
            edge_docs
        )
        
        # Generate optimization recommendations
        optimizations = self._generate_edge_optimizations(
            optimization_opportunities,
            edge_docs
        )
        
        return EdgeOptimization(
            caching_improvements=optimizations.cache_optimizations,
            function_optimizations=optimizations.function_improvements,
            routing_optimizations=optimizations.routing_improvements,
            content_optimizations=optimizations.content_improvements,
            expected_improvements=self._calculate_edge_improvements(optimizations),
            implementation_complexity=optimizations.complexity_analysis,
            rollback_strategy=optimizations.rollback_plan
        )
```

---

## API Reference

### Core Functions
- `optimize_vercel_architecture(config, goals)` - AI-powered Vercel optimization
- `optimize_global_application_performance(app, requirements)` - Global performance optimization
- `setup_global_performance_monitoring(application)` - Comprehensive performance monitoring
- `optimize_edge_performance(metrics)` - Edge performance optimization
- `get_real_time_vercel_updates(services)` - Context7 update monitoring
- `generate_edge_function_strategies(patterns, requirements)` - Edge function deployment strategies

### Context7 Integration
- `get_latest_vercel_documentation(service)` - Official docs via Context7
- `analyze_vercel_updates(services)` - Real-time update analysis
- `optimize_with_vercel_best_practices()` - Latest edge computing strategies

### Data Structures
- `VercelConfig` - Comprehensive Vercel deployment configuration
- `GlobalOptimization` - Global performance optimization recommendations
- `EdgeFunctionStrategy` - Edge function deployment and optimization
- `PerformanceIntelligence` - Real-time performance monitoring and analytics
- `MonitoringSetup` - Comprehensive monitoring and alerting configuration

---

## Changelog

- **v4.0.0** (2025-11-12): Complete Enterprise v4.0 rewrite with Context7 integration, Next.js v16, Turbopack stable, React Compiler stable, Cache Components PPR, enhanced routing and caching integration, AI-powered Vercel optimization, enterprise edge computing patterns, and intelligent global performance optimization
- **v2.0.0** (2025-11-09): Vercel Deployment and Edge Functions with production best practices
- **v1.0.0** (2025-11-09): Initial Vercel deployment platform integration

---

## Works Well With

- `moai-baas-foundation` (BaaS platform selection and architecture)
- `moai-lang-nextjs` (Next.js optimization and deployment patterns)
- `moai-lang-react` (React component optimization and performance)
- `moai-essentials-perf` (Performance optimization and monitoring)
- `moai-essentials-cicd` (CI/CD pipeline optimization and automation)
- Context7 MCP (real-time Vercel and edge computing documentation)

---

## Best Practices

✅ **DO**:
- Use Context7 integration for latest Vercel documentation and edge computing patterns
- Implement comprehensive global performance monitoring with geographic insights
- Optimize edge caching strategies for different content types and regions
- Use edge functions for geographic request processing and personalization
- Monitor Core Web Vitals and user experience metrics continuously
- Implement proper image optimization and responsive strategies
- Use preview deployments for development and testing workflows
- Establish clear performance budgets and alerting thresholds

❌ **DON'T**:
- Skip geographic performance analysis and optimization
- Ignore edge function cold start optimization and performance
- Neglect proper caching strategies for different content types
- Skip image optimization and responsive implementation
- Ignore Core Web Vitals and user experience metrics
- Use inappropriate edge function patterns for use cases
- Neglect global CDN optimization and cache hit ratios
- Skip performance monitoring and alerting setup
