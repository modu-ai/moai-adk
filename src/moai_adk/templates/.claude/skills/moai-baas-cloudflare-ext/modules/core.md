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


# Advanced Implementation (Level 3)




## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.



## Context7 Integration

### Related Libraries & Tools
- [Cloudflare Workers](/cloudflare/workers-sdk): Edge computing platform
- [Cloudflare Pages](/cloudflare/pages): JAMstack deployment

### Official Documentation
- [Documentation](https://developers.cloudflare.com/)
- [API Reference](https://developers.cloudflare.com/api/)

### Version-Specific Guides
Latest stable version: Latest
- [Release Notes](https://developers.cloudflare.com/workers/platform/changelog/)
- [Migration Guide](https://developers.cloudflare.com/workers/configuration/compatibility-dates/)
