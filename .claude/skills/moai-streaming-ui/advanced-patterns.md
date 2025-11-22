---
name: moai-streaming-ui/advanced-patterns
description: Advanced streaming UI patterns, real-time updates, progressive rendering, and live data patterns
---

# Advanced Streaming UI Patterns (v5.0.0)

## Real-Time Data Streaming

### 1. Server-Sent Events (SSE) Architecture

```typescript
// Server side
import { EventEmitter } from 'events';

class StreamingDataServer {
    private emitters: Map<string, EventEmitter> = new Map();

    setupSSEEndpoint(app: Express) {
        app.get('/api/stream/data/:streamId', (req, res) => {
            const streamId = req.params.streamId;

            // Set headers for SSE
            res.setHeader('Content-Type', 'text/event-stream');
            res.setHeader('Cache-Control', 'no-cache');
            res.setHeader('Connection', 'keep-alive');
            res.setHeader('Access-Control-Allow-Origin', '*');

            // Create stream emitter
            const emitter = new EventEmitter();
            this.emitters.set(streamId, emitter);

            // Send initial connection message
            res.write('data: {"type":"connected","streamId":"' + streamId + '"}\n\n');

            // Listen for data events
            emitter.on('data', (data) => {
                res.write(`data: ${JSON.stringify(data)}\n\n`);
            });

            emitter.on('error', (error) => {
                res.write(`data: ${JSON.stringify({ type: 'error', error: error.message })}\n\n`);
            });

            // Cleanup on disconnect
            res.on('close', () => {
                emitter.removeAllListeners();
                this.emitters.delete(streamId);
            });
        });
    }

    // Broadcast data to all connected clients
    broadcastData(streamId: string, data: any) {
        const emitter = this.emitters.get(streamId);
        if (emitter) {
            emitter.emit('data', data);
        }
    }
}

// Client side React component
interface StreamData {
    type: string;
    value: any;
    timestamp: number;
}

export function StreamingDataComponent({ streamId }: { streamId: string }) {
    const [data, setData] = useState<StreamData[]>([]);
    const [connected, setConnected] = useState(false);

    useEffect(() => {
        const eventSource = new EventSource(`/api/stream/data/${streamId}`);

        eventSource.onopen = () => {
            setConnected(true);
        };

        eventSource.onmessage = (event) => {
            const message = JSON.parse(event.data);

            if (message.type === 'connected') {
                console.log('Connected to stream');
            } else {
                setData(prev => [...prev, {
                    type: message.type,
                    value: message.data,
                    timestamp: Date.now()
                }]);
            }
        };

        eventSource.onerror = () => {
            setConnected(false);
            eventSource.close();
        };

        return () => {
            eventSource.close();
        };
    }, [streamId]);

    return (
        <div>
            <div className="status">
                {connected ? '✓ Connected' : '✗ Disconnected'}
            </div>
            <div className="stream-data">
                {data.map((item, i) => (
                    <div key={i} className="data-item">
                        {item.type}: {JSON.stringify(item.value)}
                    </div>
                ))}
            </div>
        </div>
    );
}
```

## WebSocket Streaming

### 1. Bidirectional Streaming with WebSocket

```typescript
import WebSocket from 'ws';

class BidirectionalStreamServer {
    private wss: WebSocket.Server;
    private clients: Map<string, WebSocket.WebSocket> = new Map();

    constructor(port: number) {
        this.wss = new WebSocket.Server({ port });

        this.wss.on('connection', (ws, req) => {
            const clientId = req.headers['x-client-id'] as string;
            this.clients.set(clientId, ws);

            ws.on('message', (message: string) => {
                this.handleMessage(clientId, JSON.parse(message));
            });

            ws.on('close', () => {
                this.clients.delete(clientId);
            });

            ws.on('error', (error) => {
                console.error(`Client ${clientId} error:`, error);
            });
        });
    }

    private handleMessage(clientId: string, message: any) {
        switch (message.type) {
            case 'subscribe':
                this.subscribeToStream(clientId, message.streamId);
                break;
            case 'unsubscribe':
                this.unsubscribeFromStream(clientId, message.streamId);
                break;
            case 'command':
                this.handleCommand(clientId, message.command, message.data);
                break;
        }
    }

    private subscribeToStream(clientId: string, streamId: string) {
        const client = this.clients.get(clientId);
        if (!client) return;

        // Start streaming data for this subscription
        const interval = setInterval(() => {
            const data = this.generateStreamData(streamId);
            client.send(JSON.stringify({
                type: 'data',
                streamId,
                data,
                timestamp: Date.now()
            }));
        }, 1000);

        // Store interval for cleanup
        if (!this.subscriptions) this.subscriptions = new Map();
        const key = `${clientId}:${streamId}`;
        this.subscriptions.set(key, interval);
    }

    broadcastToAllClients(message: any) {
        for (const [, client] of this.clients) {
            if (client.readyState === WebSocket.OPEN) {
                client.send(JSON.stringify(message));
            }
        }
    }
}

// Client side
export function WebSocketStreamComponent({ streamId }: { streamId: string }) {
    const [data, setData] = useState<any[]>([]);
    const wsRef = useRef<WebSocket>();

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8080');
        wsRef.current = ws;

        ws.onopen = () => {
            ws.send(JSON.stringify({
                type: 'subscribe',
                streamId,
                clientId: 'client-' + Math.random()
            }));
        };

        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            if (message.type === 'data') {
                setData(prev => [...prev.slice(-49), message.data]);
            }
        };

        return () => {
            ws.close();
        };
    }, [streamId]);

    return (
        <div className="stream-container">
            {data.map((item, i) => (
                <StreamDataItem key={i} data={item} />
            ))}
        </div>
    );
}
```

## Progressive Rendering

### 1. Server-Side Streaming with React

```typescript
// Node.js server with React rendering
import { renderToString } from 'react-dom/server';
import { Readable } from 'stream';

app.get('/stream-page', (req, res) => {
    res.setHeader('Content-Type', 'text/html; charset=utf-8');

    const readable = new Readable({
        read() {}
    });

    readable.pipe(res);

    // Send HTML head
    readable.push(`<!DOCTYPE html>
<html>
<head>
    <title>Streaming Page</title>
</head>
<body>`);

    // Stream initial content
    readable.push('<div id="app">');

    // Generate and stream components progressively
    const components = [
        { name: 'header', component: Header },
        { name: 'main-content', component: MainContent },
        { name: 'sidebar', component: Sidebar },
        { name: 'footer', component: Footer }
    ];

    let index = 0;

    const streamNextComponent = () => {
        if (index >= components.length) {
            readable.push('</div>');
            readable.push('</body></html>');
            readable.push(null);
            return;
        }

        const { name, component: Component } = components[index];

        try {
            const html = renderToString(<Component />);
            readable.push(`<div id="${name}">${html}</div>`);

            // Flush to client
            res.flush?.();

            index++;
            // Continue with next component after a delay
            setTimeout(streamNextComponent, 100);
        } catch (error) {
            readable.destroy(error);
        }
    };

    streamNextComponent();
});
```

### 2. React Suspense with Streaming

```typescript
export function StreamingPage() {
    return (
        <html>
            <head>
                <title>Streaming Content</title>
            </head>
            <body>
                <Header />
                <Suspense fallback={<MainContentSkeleton />}>
                    <MainContent />
                </Suspense>
                <Suspense fallback={<SidebarSkeleton />}>
                    <Sidebar />
                </Suspense>
                <Footer />
            </body>
        </html>
    );
}

// On server (using Node.js streaming)
import { renderToStaticMarkup } from 'react-dom/server';

export async function renderStreamingPage() {
    const readable = await renderToReadableStream(
        <StreamingPage />,
        {
            onError(error) {
                console.error('Streaming error:', error);
            }
        }
    );

    return readable;
}
```

---

**Version**: 5.0.0 | **Last Updated**: 2025-11-22 | **Enterprise Ready**: ✓
