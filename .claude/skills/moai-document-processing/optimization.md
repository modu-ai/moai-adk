---
name: moai-document-processing/optimization
description: Performance optimization for document processing, streaming, and batch operations
---

# Document Processing Optimization (v5.0.0)

## Streaming Document Processing

### 1. Memory-Efficient Stream Processing

```typescript
import { PassThrough, Transform } from 'stream';
import fs from 'fs';

class StreamingDocumentProcessor {
    // Process large PDFs without loading into memory
    async processPDFStream(
        filePath: string
    ): Promise<ProcessedDocument> {
        const readStream = fs.createReadStream(filePath, {
            highWaterMark: 64 * 1024 // 64KB chunks
        });

        const processStream = new Transform({
            transform: (chunk: Buffer, encoding: string, callback: Function) => {
                try {
                    const processed = this.processChunk(chunk);
                    callback(null, processed);
                } catch (error) {
                    callback(error);
                }
            }
        });

        const outputStream = new PassThrough();
        let content = '';

        return new Promise((resolve, reject) => {
            readStream
                .pipe(processStream)
                .pipe(outputStream)
                .on('data', (chunk: Buffer) => {
                    content += chunk.toString();
                })
                .on('end', () => {
                    resolve(this.finalizeDocument(content));
                })
                .on('error', reject);
        });
    }

    private processChunk(chunk: Buffer): Buffer {
        // Process PDF chunk without loading entire file
        const text = chunk.toString();

        // Extract text, tables, etc. from chunk
        const processed = {
            text: this.extractText(text),
            metadata: this.extractMetadata(text)
        };

        return Buffer.from(JSON.stringify(processed));
    }
}
```

### 2. Batch Processing Optimization

```typescript
class BatchDocumentProcessor {
    async processBatch(
        documents: DocumentInput[],
        options: { workers: number; batchSize: number } = { workers: 4, batchSize: 10 }
    ): Promise<ProcessedDocument[]> {
        // Create worker pool
        const workers = this.createWorkerPool(options.workers);

        // Split into chunks
        const chunks = this.chunkArray(documents, options.batchSize);
        const results: ProcessedDocument[] = [];

        for (const chunk of chunks) {
            // Distribute to workers
            const promises = workers.map((worker, i) => {
                const workerDocs = chunk.slice(
                    i * Math.ceil(chunk.length / options.workers),
                    (i + 1) * Math.ceil(chunk.length / options.workers)
                );
                return worker.process(workerDocs);
            });

            const batchResults = await Promise.all(promises);
            results.push(...batchResults.flat());
        }

        return results;
    }

    // Parallel processing with connection pooling
    async processParallel(
        documents: DocumentInput[],
        maxConcurrent: number = 5
    ): Promise<ProcessedDocument[]> {
        const results: ProcessedDocument[] = [];
        const queue = [...documents];
        const processing = new Set<Promise<void>>();

        while (queue.length > 0 || processing.size > 0) {
            // Keep up to maxConcurrent processes running
            while (processing.size < maxConcurrent && queue.length > 0) {
                const doc = queue.shift()!;
                const promise = this.processDocument(doc)
                    .then(result => results.push(result))
                    .finally(() => processing.delete(promise));

                processing.add(promise);
            }

            // Wait for one to complete before starting next
            if (processing.size > 0) {
                await Promise.race(processing);
            }
        }

        return results;
    }
}
```

## Caching and Memoization

### 1. Multi-Layer Caching

```typescript
class DocumentCache {
    private memoryCache = new Map<string, CacheEntry>();
    private redisClient: Redis;
    private s3Client: S3Client;

    async getProcessed(documentId: string): Promise<ProcessedDocument | null> {
        // Layer 1: Memory cache (milliseconds)
        let cached = this.memoryCache.get(documentId);
        if (cached && !this.isExpired(cached)) {
            metrics.increment('cache.memory.hit');
            return cached.document;
        }

        // Layer 2: Redis cache (seconds)
        try {
            cached = await this.redisClient.get(`doc:${documentId}`);
            if (cached) {
                metrics.increment('cache.redis.hit');
                this.memoryCache.set(documentId, cached);
                return cached.document;
            }
        } catch (error) {
            logger.warn('Redis cache miss', { documentId });
        }

        // Layer 3: S3 cache (permanent)
        try {
            cached = await this.s3Client.getObject({
                Bucket: 'document-cache',
                Key: `processed/${documentId}.json`
            });

            if (cached) {
                metrics.increment('cache.s3.hit');
                this.redisClient.set(`doc:${documentId}`, cached, 'EX', 3600);
                this.memoryCache.set(documentId, cached);
                return cached.document;
            }
        } catch (error) {
            logger.warn('S3 cache miss', { documentId });
        }

        return null;
    }

    async cache(documentId: string, document: ProcessedDocument) {
        const entry = { document, timestamp: Date.now() };

        // Cache to all layers
        this.memoryCache.set(documentId, entry);
        await this.redisClient.set(`doc:${documentId}`, JSON.stringify(entry), 'EX', 3600);
        await this.s3Client.putObject({
            Bucket: 'document-cache',
            Key: `processed/${documentId}.json`,
            Body: JSON.stringify(document)
        });
    }

    invalidate(documentId: string) {
        this.memoryCache.delete(documentId);
        this.redisClient.del(`doc:${documentId}`);
        // Keep S3 for history
    }
}
```

### 2. Smart Cache Invalidation

```typescript
class CacheInvalidationManager {
    async invalidateOnUpdate(
        documentId: string,
        changeType: 'content' | 'metadata' | 'full'
    ) {
        switch (changeType) {
            case 'content':
                // Invalidate text extraction cache
                await this.cache.invalidate(`text:${documentId}`);
                break;
            case 'metadata':
                // Invalidate metadata cache only
                await this.cache.invalidate(`meta:${documentId}`);
                break;
            case 'full':
                // Full invalidation
                await this.cache.invalidate(`doc:${documentId}`);
                // Pre-generate for frequently accessed documents
                if (await this.isFavorite(documentId)) {
                    this.preProcessDocument(documentId);
                }
                break;
        }
    }
}
```

## OCR Performance Optimization

### 1. OCR Quality vs Speed Tradeoff

```typescript
class OptimizedOCR {
    async processWithQualityLevel(
        imagePath: string,
        qualityLevel: 'fast' | 'balanced' | 'accurate' = 'balanced'
    ): Promise<OCRResult> {
        const config = {
            fast: {
                languages: ['eng'],        // Single language
                scale: 2,                  // Lower resolution
                timeout: 5000
            },
            balanced: {
                languages: ['eng', 'fra'], // Common languages
                scale: 3,
                timeout: 10000
            },
            accurate: {
                languages: ['eng', 'fra', 'deu', 'spa', 'ita'], // Multiple languages
                scale: 4,                  // Higher resolution
                timeout: 30000
            }
        };

        const selectedConfig = config[qualityLevel];

        return await this.runOCR(imagePath, selectedConfig);
    }

    // Preprocess image for better OCR
    private preprocessImage(imagePath: string) {
        // Resize for OCR efficiency
        // Increase contrast
        // Deskew if needed
        return sharp(imagePath)
            .resize({ width: 2400, height: 3000 })
            .threshold(150)
            .toBuffer();
    }
}
```

## Performance Monitoring

### 1. Document Processing Metrics

```typescript
class ProcessingMetrics {
    async track(documentId: string, format: string, operation: string) {
        const startTime = performance.now();
        let fileSize = 0;
        let processedSize = 0;

        try {
            fileSize = await this.getFileSize(documentId);

            const result = await this.processDocument(documentId);

            processedSize = JSON.stringify(result).length;

            const duration = performance.now() - startTime;
            const throughput = (fileSize / 1024) / (duration / 1000); // KB/s

            // Log metrics
            logger.info('Document processing', {
                document_id: documentId,
                format,
                operation,
                file_size_kb: fileSize / 1024,
                processed_size_kb: processedSize / 1024,
                duration_ms: duration,
                throughput_kbps: throughput,
                timestamp: new Date()
            });

            // Alert if slow
            if (duration > 10000) {
                metrics.increment('processing.slow');
                logger.warn('Slow document processing', { duration_ms: duration });
            }
        } catch (error) {
            metrics.increment('processing.error');
            logger.error('Document processing failed', { documentId, error });
            throw error;
        }
    }
}
```

### 2. Performance Budgets

```typescript
interface ProcessingBudget {
    maxDuration: number;      // milliseconds
    maxMemory: number;        // MB
    maxCPU: number;           // percentage
}

const budgets: Record<string, ProcessingBudget> = {
    'pdf-text-extraction': {
        maxDuration: 5000,
        maxMemory: 500,
        maxCPU: 80
    },
    'ocr-processing': {
        maxDuration: 30000,
        maxMemory: 1000,
        maxCPU: 95
    },
    'format-conversion': {
        maxDuration: 10000,
        maxMemory: 300,
        maxCPU: 60
    }
};

async function validatePerformanceBudget(
    operation: string,
    metrics: OperationMetrics
) {
    const budget = budgets[operation];
    if (!budget) return;

    const violations: string[] = [];

    if (metrics.duration > budget.maxDuration) {
        violations.push(`Duration ${metrics.duration}ms > ${budget.maxDuration}ms`);
    }

    if (metrics.memory > budget.maxMemory) {
        violations.push(`Memory ${metrics.memory}MB > ${budget.maxMemory}MB`);
    }

    if (violations.length > 0) {
        throw new Error(`Performance budget exceeded:\n${violations.join('\n')}`);
    }
}
```

---

**Version**: 5.0.0 | **Last Updated**: 2025-11-22 | **Enterprise Ready**: ✓
