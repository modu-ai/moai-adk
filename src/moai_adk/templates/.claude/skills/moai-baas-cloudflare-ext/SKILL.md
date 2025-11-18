---
name: moai-baas-cloudflare-ext
version: 4.0.0
created: 2025-11-11
updated: '2025-11-18'
status: stable
description: Enterprise Cloudflare Edge Platform with AI-powered edge computing architecture,
  Context7 integration, and intelligent global orchestration for scalable modern applications
keywords:
- cloudflare
- edge-computing
- workers
- pages
- cdn
- security
- context7-integration
- ai-orchestration
- production-deployment
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
stability: stable
---


# Enterprise Cloudflare Edge Platform Expert 

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-baas-cloudflare-ext |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Edge Platform Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when Cloudflare keywords detected |

---

## What It Does

Enterprise Cloudflare Edge Platform expert with AI-powered edge computing architecture, Context7 integration, and intelligent global orchestration for scalable modern applications.

**Revolutionary  capabilities**:
- 🤖 **AI-Powered Edge Architecture** using Context7 MCP for latest edge patterns
- 📊 **Intelligent Global Orchestration** with automated edge resource optimization
- 🚀 **Advanced Workers Runtime** with AI-driven performance optimization
- 🔗 **Enterprise Edge Security** with zero-configuration DDoS protection
- 📈 **Predictive Performance Analytics** with global latency optimization

---

## When to Use

**Automatic triggers**:
- Cloudflare edge computing and Workers deployment discussions
- Global CDN configuration and performance optimization
- Edge security implementation and DDoS protection strategies
- Multi-region application deployment and latency optimization

**Manual invocation**:
- Designing enterprise Cloudflare architectures with optimal edge patterns
- Implementing advanced Workers and Pages with edge computing
- Planning global application deployment with Cloudflare
- Optimizing edge performance and security configurations

---

# Quick Reference (Level 1)

## Cloudflare Edge Platform (November 2025)

### Core Services Overview
- **Workers**: Serverless edge computing with 0ms cold starts
- **Pages**: Static site hosting with edge functions
- **Durable Objects**: Edge state management with coordination
- **D1**: Distributed SQL database at the edge
- **KV**: Global key-value storage with instant replication
- **R2**: Object storage alternative to AWS S3

### Latest Features (November 2025)
- **Workers VPC Services (Beta)**: Secure private network access
- **Enhanced Node.js Compatibility**: Improved runtime compatibility
- **WebSocket Messages**: Support for up to 32 MiB messages
- **Workers Builds**: pnpm 10 support and improved builds
- **remove_nodejs_compat_eol flag**: Modern Node.js API support

### Global Network Performance
- **Global Coverage**: 300+ cities worldwide
- **Edge Latency**: Sub-10ms response times
- **Throughput**: 40+ Tbps global network capacity
- **DDoS Protection**: Automated mitigation at layer 3-7

### Security Features
- **WAF**: Web Application Firewall with custom rules
- **Bot Management**: Advanced bot detection and mitigation
- **Certificate Management**: Automatic SSL/TLS certificate provisioning
- **Network Security**: DDoS protection and edge firewall rules

---

# Core Implementation (Level 2)

## Cloudflare Architecture Intelligence

```python
# AI-powered Cloudflare architecture optimization with Context7
class CloudflareArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.edge_analyzer = EdgeAnalyzer()
        self.security_optimizer = SecurityOptimizer()
    
    async def design_optimal_cloudflare_architecture(self, 
                                                   requirements: ApplicationRequirements) -> CloudflareArchitecture:
        """Design optimal Cloudflare architecture using AI analysis."""
        
        # Get latest Cloudflare and edge computing documentation via Context7
        cloudflare_docs = await self.context7_client.get_library_docs(
            context7_library_id='/cloudflare/docs',
            topic="workers edge computing security optimization 2025",
            tokens=3000
        )
        
        edge_docs = await self.context7_client.get_library_docs(
            context7_library_id='/edge-computing/docs',
            topic="performance optimization global deployment 2025",
            tokens=2000
        )
        
        # Optimize edge computing strategy
        edge_strategy = self.edge_analyzer.optimize_edge_strategy(
            requirements.global_needs,
            requirements.performance_requirements,
            cloudflare_docs
        )
        
        # Configure security framework
        security_config = self.security_optimizer.configure_security(
            requirements.security_level,
            requirements.threat_model,
            cloudflare_docs
        )
        
        return CloudflareArchitecture(
            workers_configuration=self._design_workers(requirements),
            pages_setup=self._configure_pages(requirements),
            storage_strategy=self._design_storage(requirements),
            database_configuration=self._configure_d1(requirements),
            security_framework=security_config,
            edge_optimization=edge_strategy,
            global_deployment=self._plan_global_deployment(requirements),
            performance_predictions=edge_strategy.predictions
        )
```

## Advanced Workers Implementation

```typescript
// High-performance Cloudflare Worker with TypeScript
export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    
    // Route handling based on path
    switch (url.pathname) {
      case '/api/auth':
        return handleAuthentication(request, env);
      case '/api/data':
        return handleDataRequest(request, env);
      case '/api/webhook':
        return handleWebhook(request, env);
      default:
        return handleStaticRequest(request, env);
    }
  },

  async scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext): Promise<void> {
    // Scheduled tasks for data cleanup and analytics
    await runScheduledTasks(env);
  }
};

// Authentication handler with edge optimization
async function handleAuthentication(request: Request, env: Env): Promise<Response> {
  try {
    const { email, password } = await request.json() as LoginRequest;
    
    // Validate credentials at edge
    const user = await validateCredentials(email, password, env);
    if (!user) {
      return new Response(JSON.stringify({ error: 'Invalid credentials' }), {
        status: 401,
        headers: { 'Content-Type': 'application/json' }
      });
    }

    // Generate JWT token at edge
    const token = await generateEdgeToken(user, env);
    
    // Set secure cookies with edge storage
    const response = new Response(JSON.stringify({ user: { id: user.id, email: user.email } }), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Set-Cookie': `auth_token=${token}; Path=/; HttpOnly; Secure; SameSite=Strict; Max-Age=3600`,
        'Access-Control-Allow-Origin': env.ALLOWED_ORIGIN,
        'Access-Control-Allow-Credentials': 'true'
      }
    });

    return response;
  } catch (error) {
    return new Response(JSON.stringify({ error: 'Authentication failed' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    });
  }
}

// Edge-optimized data handler with KV storage
async function handleDataRequest(request: Request, env: Env): Promise<Response> {
  const cacheKey = new Request(request.url, request);
  const cache = caches.default;
  
  // Check edge cache first
  const cached = await cache.match(cacheKey);
  if (cached) {
    return cached;
  }

  try {
    const data = await fetchDataFromKV(request.url, env);
    
    const response = new Response(JSON.stringify(data), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 'public, max-age=300', // 5 minutes edge cache
        'Edge-Cache-Tag': 'api-data',
        'Access-Control-Allow-Origin': env.ALLOWED_ORIGIN
      }
    });

    // Store in edge cache
    ctx.waitUntil(cache.put(cacheKey, response.clone()));
    return response;
    
  } catch (error) {
    return new Response(JSON.stringify({ error: 'Data fetch failed' }), {
      status: 500,
      headers: { 'Content-Type': 'application/json' }
    });
  }
}
```

## Durable Objects for Real-Time Features

```typescript
// Durable Object for real-time collaboration
export class RealtimeRoom {
  private state: DurableObjectState;
  private env: Env;
  private sessions: Map<string, WebSocket>;
  private roomData: RoomData;

  constructor(state: DurableObjectState, env: Env) {
    this.state = state;
    this.env = env;
    this.sessions = new Map();
    this.roomData = { participants: [], messages: [], lastActivity: Date.now() };
  }

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);
    
    switch (url.pathname) {
      case '/websocket':
        return this.handleWebSocket(request);
      case '/join':
        return this.handleJoin(request);
      case '/leave':
        return this.handleLeave(request);
      default:
        return new Response('Not Found', { status: 404 });
    }
  }

  async handleWebSocket(request: Request): Promise<Response> {
    const upgradeHeader = request.headers.get('Upgrade');
    if (upgradeHeader !== 'websocket') {
      return new Response('Expected websocket', { status: 426 });
    }

    const [client, server] = Object.values(new WebSocketPair());
    
    // Accept WebSocket connection
    server.accept();
    
    const sessionId = crypto.randomUUID();
    this.sessions.set(sessionId, server);
    
    // Handle WebSocket messages
    server.addEventListener('message', (event) => {
      this.handleMessage(sessionId, event.data as string);
    });

    // Handle connection close
    server.addEventListener('close', () => {
      this.handleDisconnect(sessionId);
    });

    // Send current room state to new participant
    server.send(JSON.stringify({
      type: 'initial_state',
      data: this.roomData
    }));

    return new Response(null, { status: 101, webSocket: client });
  }

  private handleMessage(sessionId: string, data: string): void {
    try {
      const message = JSON.parse(data);
      
      switch (message.type) {
        case 'cursor_move':
          this.broadcastToOthers(sessionId, {
            type: 'cursor_update',
            sessionId,
            position: message.position
          });
          break;
          
        case 'text_edit':
          this.roomData.messages.push({
            sessionId,
            content: message.content,
            timestamp: Date.now()
          });
          this.broadcastToAll({
            type: 'text_update',
            sessionId,
            content: message.content
          });
          break;
          
        case 'user_presence':
          this.updateUserPresence(sessionId, message.presence);
          break;
      }
      
      // Persist to Durable Object storage
      this.state.storage.put('roomData', this.roomData);
      
    } catch (error) {
      console.error('Error handling message:', error);
    }
  }

  private broadcastToOthers(senderSessionId: string, message: any): void {
    for (const [sessionId, websocket] of this.sessions) {
      if (sessionId !== senderSessionId) {
        websocket.send(JSON.stringify(message));
      }
    }
  }

  private broadcastToAll(message: any): void {
    for (const websocket of this.sessions.values()) {
      websocket.send(JSON.stringify(message));
    }
  }
}
```

---

# Advanced Implementation (Level 3)

## Edge Security Implementation

```typescript
// Advanced WAF and security rules configuration
export class EdgeSecurityManager {
  private wafRules: WAFRule[] = [
    {
      id: 'sql_injection_protection',
      expression: '(http.request.uri.query contains "SELECT" or http.request.uri.query contains "INSERT" or http.request.uri.query contains "DELETE") and (http.request.uri.query contains "UNION" or http.request.uri.query contains "DROP")',
      action: 'block',
      description: 'Block SQL injection attempts'
    },
    {
      id: 'xss_protection',
      expression: 'http.request.body.raw contains "<script>" or http.request.uri.query contains "<script>"',
      action: 'block',
      description: 'Block XSS attempts'
    },
    {
      id: 'rate_limiting',
      expression: 'http.request.method in {"POST", "PUT", "DELETE"} and ip.src.neq.192.0.2.0/24',
      action: 'rate_limit',
      rate_limit: {
        requests_per_minute: 60,
        burst: 10
      },
      description: 'Rate limit non-admin requests'
    }
  ];

  configureWASRules(): WAFConfiguration {
    return {
      rules: this.wafRules,
      custom_response: {
        status: 403,
        content: '{"error": "Access denied"}',
        headers: { 'Content-Type': 'application/json' }
      },
      logging: {
        enabled: true,
        redact_sensitive_data: true
      }
    };
  }
}

// DDoS protection and rate limiting
export class DDoSProtection {
  private rateLimiters: Map<string, RateLimiter> = new Map();

  async handleRequest(request: Request, env: Env): Promise<Response | null> {
    const clientIP = request.headers.get('CF-Connecting-IP');
    const userAgent = request.headers.get('User-Agent');
    
    // Check IP reputation
    const reputation = await this.checkIPReputation(clientIP, env);
    if (reputation.isMalicious) {
      return new Response('Blocked', { status: 403 });
    }

    // Apply rate limiting
    if (!await this.checkRateLimit(clientIP, request.url, env)) {
      return new Response('Rate limit exceeded', { 
        status: 429,
        headers: { 'Retry-After': '60' }
      });
    }

    // Check for suspicious patterns
    const suspiciousScore = await this.analyzeRequestPatterns(request, env);
    if (suspiciousScore > 0.8) {
      // Enable additional verification
      return this.requireAdditionalVerification(request, env);
    }

    return null; // Allow request to proceed
  }

  private async checkRateLimit(ip: string, endpoint: string, env: Env): Promise<boolean> {
    const key = `rate_limit:${ip}:${endpoint}`;
    const now = Date.now();
    const window = 60000; // 1 minute
    const limit = 100; // requests per minute

    const current = await env.KV.get(key);
    const count = current ? parseInt(current) : 0;

    if (count >= limit) {
      return false;
    }

    // Increment counter with expiration
    await env.KV.put(key, (count + 1).toString(), { expirationTtl: 60 });
    return true;
  }
}
```

### Global Performance Optimization

```typescript
// Smart routing and cache optimization
export class GlobalPerformanceOptimizer {
  async optimizeGlobalRouting(request: Request, env: Env): Promise<OptimalRoute> {
    const clientIP = request.headers.get('CF-Connecting-IP');
    const userLocation = request.cf?.colo; // Cloudflare data center location
    
    // Determine optimal edge location
    const optimalRegion = await this.findOptimalRegion(clientIP, userLocation);
    
    // Configure smart caching
    const cacheStrategy = this.determineCacheStrategy(request.url, request.method);
    
    return {
      region: optimalRegion,
      cacheStrategy,
      estimatedLatency: this.estimateLatency(userLocation, optimalRegion)
    };
  }

  private determineCacheStrategy(url: string, method: string): CacheStrategy {
    const isGET = method === 'GET';
    const isAPI = url.includes('/api/');
    const isStatic = url.includes('/static/') || url.match(/\.(css|js|png|jpg|webp)$/i);

    if (isStatic) {
      return {
        edgeTTL: 86400, // 1 day
        browserTTL: 3600, // 1 hour
        cacheKey: url,
        vary: ['Accept-Encoding']
      };
    }

    if (isAPI && isGET) {
      return {
        edgeTTL: 300, // 5 minutes
        browserTTL: 60, // 1 minute
        cacheKey: url,
        vary: ['Authorization', 'Accept']
      };
    }

    return {
      edgeTTL: 0, // No caching
      browserTTL: 0,
      cacheKey: null
    };
  }

  // Smart image optimization at the edge
  async optimizeImages(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    
    if (!url.pathname.match(/\.(jpg|jpeg|png|webp|gif)$/i)) {
      return fetch(request);
    }

    // Parse image transformation parameters
    const width = url.searchParams.get('w');
    const height = url.searchParams.get('h');
    const quality = url.searchParams.get('q') || '80';
    const format = url.searchParams.get('f');

    // Fetch original image
    const originalResponse = await fetch(request);
    if (!originalResponse.ok) {
      return originalResponse;
    }

    // Transform image using Cloudflare Image Resizing
    if (env.IMAGE_RESIZING_ENABLED) {
      const transformedUrl = `https://api.cloudflare.com/client/v4/accounts/${env.CLOUDFLARE_ACCOUNT_ID}/images/v2/transform`;
      
      const transformResponse = await fetch(transformedUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${env.CLOUDFLARE_API_TOKEN}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          input: { url: request.url },
          transforms: [
            { resize: { width: parseInt(width), height: parseInt(height) } },
            { quality: parseInt(quality) },
            { output: { format: format || 'auto' } }
          ]
        })
      });

      if (transformResponse.ok) {
        return transformResponse;
      }
    }

    return originalResponse;
  }
}
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Cloudflare Operations
- `create_worker(script, bindings)` - Deploy Workers script
- `create_pages_project(project_config)` - Create Pages project
- `create_durable_object(class_name, script_name)` - Deploy Durable Object
- `create_kv_namespace(name)` - Create KV storage namespace
- `create_d1_database(name)` - Create D1 database
- `configure_waf(rules)` - Configure WAF rules

### Context7 Integration
- `get_latest_cloudflare_documentation()` - Official Cloudflare docs via Context7
- `analyze_edge_performance_patterns()` - Edge optimization via Context7
- `optimize_workers_configuration()` - Workers best practices via Context7

## Best Practices (November 2025)

### DO
- Use Workers for compute-intensive tasks at the edge
- Implement comprehensive caching strategies
- Configure DDoS protection and WAF rules
- Use Durable Objects for real-time collaboration features
- Optimize global routing for minimum latency
- Monitor edge performance and costs
- Use appropriate storage (KV, D1, R2) based on use case
- Implement security headers and policies

### DON'T
- Deploy large applications to Workers (size limits apply)
- Skip security configuration for production
- Ignore edge computing costs and limits
- Use D1 for highly transactional workloads
- Forget to implement proper error handling
- Neglect monitoring and observability
- Overuse edge resources when origin is sufficient
- Skip compliance requirements for data residency

## Works Well With

- `moai-baas-foundation` (Enterprise BaaS architecture patterns)
- `moai-domain-frontend` (Frontend edge optimization)
- `moai-security-api` (API security implementation)
- `moai-essentials-perf` (Performance optimization)
- `moai-foundation-trust` (Security and compliance)
- `moai-baas-vercel-ext` (Edge deployment comparison)
- `moai-baas-railway-ext` (Full-stack deployment alternative)
- `moai-domain-backend` (Backend edge optimization)

## Changelog

- ** .0** (2025-11-13): Complete Enterprise   rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, November 2025 Cloudflare platform updates, and advanced edge security patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, edge patterns, Workers optimization
- **v1.0.0** (2025-11-11): Initial Cloudflare edge platform

---

**End of Skill** | Updated 2025-11-13

## Security & Compliance

### Edge Security Framework
- Zero-trust network architecture with Cloudflare Access
- Advanced DDoS protection with automatic mitigation
- Web Application Firewall with custom rules and ML protection
- Bot management and CAPTCHA integration

### Data Protection
- End-to-end encryption for all edge communications
- GDPR compliance with data processing at the edge
- Regional data residency with smart routing
- Comprehensive audit logging and monitoring

---

**End of Enterprise Cloudflare Edge Platform Expert **
