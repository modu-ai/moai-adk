---
name: moai-streaming-ui/optimization
description: Performance optimization for streaming UIs, message batching, backpressure handling
---

# Streaming UI Optimization (v5.0.0)

## Message Batching and Debouncing

### 1. Intelligent Message Batching

```typescript
class StreamingMessageBatcher {
    private queue: Map<string, any[]> = new Map();
    private timers: Map<string, NodeJS.Timeout> = new Map();
    private batchSize = 10;
    private batchInterval = 100; // ms

    addMessage(streamId: string, data: any) {
        if (!this.queue.has(streamId)) {
            this.queue.set(streamId, []);
        }

        const queue = this.queue.get(streamId)!;
        queue.push(data);

        // Send immediately if batch is full
        if (queue.length >= this.batchSize) {
            this.flushBatch(streamId);
        } else if (!this.timers.has(streamId)) {
            // Otherwise schedule flush
            const timer = setTimeout(() => {
                this.flushBatch(streamId);
            }, this.batchInterval);

            this.timers.set(streamId, timer);
        }
    }

    private flushBatch(streamId: string) {
        const queue = this.queue.get(streamId);
        if (!queue || queue.length === 0) return;

        // Send batched messages
        this.sendBatch(streamId, queue);

        // Reset queue
        this.queue.set(streamId, []);

        // Clear timer
        const timer = this.timers.get(streamId);
        if (timer) {
            clearTimeout(timer);
            this.timers.delete(streamId);
        }
    }

    private sendBatch(streamId: string, messages: any[]) {
        const client = this.getClient(streamId);
        if (!client) return;

        client.send(JSON.stringify({
            type: 'batch',
            streamId,
            messages,
            count: messages.length,
            timestamp: Date.now()
        }));
    }
}

// Client-side handling
export function useBatchedStream(streamId: string) {
    const [data, setData] = useState<any[]>([]);

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8080');

        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);

            if (message.type === 'batch') {
                setData(prev => {
                    const newData = [...prev, ...message.messages];
                    // Keep last 100 items
                    return newData.slice(-100);
                });
            }
        };

        return () => ws.close();
    }, [streamId]);

    return data;
}
```

## Backpressure Handling

### 1. Flow Control with Backpressure

```typescript
class BackpressureManager {
    private clientBuffers: Map<string, Buffer> = new Map();
    private maxBufferSize = 1024 * 1024; // 1MB
    private highWaterMark = 0.8; // 80%
    private lowWaterMark = 0.2;  // 20%

    sendData(streamId: string, data: any, client: WebSocket) {
        const buffer = this.clientBuffers.get(streamId) || Buffer.alloc(0);
        const newData = Buffer.from(JSON.stringify(data) + '\n');

        const newBuffer = Buffer.concat([buffer, newData]);
        const usage = newBuffer.length / this.maxBufferSize;

        if (usage > this.highWaterMark) {
            // Backpressure - pause sending
            this.pauseProducer(streamId);
            logger.warn(`Backpressure triggered for ${streamId}`);
        }

        this.clientBuffers.set(streamId, newBuffer);

        // Attempt to send what we can
        const sent = client.write(newBuffer);

        if (sent) {
            // Successfully written, clear buffer
            this.clientBuffers.delete(streamId);

            const currentUsage = 0;
            if (currentUsage < this.lowWaterMark) {
                this.resumeProducer(streamId);
            }
        }
    }

    private pauseProducer(streamId: string) {
        const producer = this.getProducer(streamId);
        if (producer) {
            producer.pause();
        }
    }

    private resumeProducer(streamId: string) {
        const producer = this.getProducer(streamId);
        if (producer) {
            producer.resume();
        }
    }
}
```

## Compression and Bandwidth Optimization

### 1. Message Compression

```typescript
class CompressedStreamer {
    async sendCompressed(
        streamId: string,
        data: any,
        client: WebSocket,
        compressionLevel: 'low' | 'medium' | 'high' = 'medium'
    ) {
        const json = JSON.stringify(data);
        const compressed = await this.compress(json, compressionLevel);

        const message = {
            type: 'compressed',
            data: compressed.toString('base64'),
            originalSize: json.length,
            compressedSize: compressed.length,
            ratio: (compressed.length / json.length * 100).toFixed(1) + '%'
        };

        client.send(JSON.stringify(message));

        // Track metrics
        metrics.recordCompression(compressed.length, json.length);
    }

    private async compress(
        data: string,
        level: 'low' | 'medium' | 'high'
    ): Promise<Buffer> {
        const zlib = require('zlib');

        const options = {
            low: { level: zlib.Z_DEFAULT_COMPRESSION },
            medium: { level: 6 },
            high: { level: zlib.Z_BEST_COMPRESSION }
        };

        return new Promise((resolve, reject) => {
            zlib.gzip(data, options[level], (err, compressed) => {
                if (err) reject(err);
                else resolve(compressed);
            });
        });
    }
}

// Client decompression
function useCompressedStream(streamId: string) {
    const [data, setData] = useState<any[]>([]);

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8080');

        ws.onmessage = async (event) => {
            const message = JSON.parse(event.data);

            if (message.type === 'compressed') {
                const decompressed = await decompressData(message.data);
                setData(prev => [...prev, decompressed]);
            }
        };

        return () => ws.close();
    }, [streamId]);

    return data;
}

async function decompressData(base64: string): Promise<any> {
    const buffer = Buffer.from(base64, 'base64');
    const decompressed = await new Promise<Buffer>((resolve, reject) => {
        require('zlib').gunzip(buffer, (err, data) => {
            if (err) reject(err);
            else resolve(data);
        });
    });

    return JSON.parse(decompressed.toString());
}
```

## Performance Monitoring

### 1. Streaming Performance Metrics

```typescript
class StreamingMetrics {
    trackMessageLatency(streamId: string, sendTime: number) {
        const latency = Date.now() - sendTime;

        metrics.observe('stream.message.latency', latency, {
            streamId,
            bucket: latency < 100 ? 'fast' : latency < 500 ? 'normal' : 'slow'
        });

        if (latency > 1000) {
            logger.warn('High latency detected', { streamId, latency });
        }
    }

    trackThroughput(streamId: string, bytesTransferred: number, durationMs: number) {
        const throughputMbps = (bytesTransferred * 8 / 1024 / 1024) / (durationMs / 1000);

        metrics.observe('stream.throughput', throughputMbps, { streamId });
    }

    trackClientHealth(streamId: string) {
        const buffer = this.getClientBuffer(streamId);
        const memoryUsage = process.memoryUsage();

        metrics.gauge('stream.client_buffer', buffer?.length || 0, { streamId });
        metrics.gauge('process.memory.heap_used', memoryUsage.heapUsed);
    }
}
```

---

**Version**: 5.0.0 | **Last Updated**: 2025-11-22 | **Enterprise Ready**: ✓
