---
name: moai-library-mermaid/optimization
description: Performance optimization, rendering strategies, and SVG generation for Mermaid diagrams
---

# Mermaid Diagram Optimization (v6.0.0)

## Rendering Performance

### 1. Mermaid Rendering Strategies

Strategy 1: Server-Side Rendering (SSR)
```javascript
// Node.js server with mermaid-cli
const { execSync } = require('child_process');
const fs = require('fs');

function generateSVG(mermaidCode, outputPath) {
 const tempFile = '/tmp/diagram.mmd';
 fs.writeFileSync(tempFile, mermaidCode);

 try {
 // Use mermaid-cli for server-side rendering
 execSync(`mmdc -i ${tempFile} -o ${outputPath} -t dark -s 4`);

 return {
 status: 'success',
 path: outputPath,
 size: fs.statSync(outputPath).size
 };
 } catch (error) {
 return {
 status: 'error',
 message: error.message
 };
 }
}

// Cache generated diagrams
const diagrams = new Map();

app.get('/diagram/:id.svg', (req, res) => {
 const cacheKey = `diagram-${req.params.id}`;

 if (diagrams.has(cacheKey)) {
 return res.sendFile(diagrams.get(cacheKey));
 }

 const mermaidCode = getMermaidCode(req.params.id);
 const outputPath = generateSVG(mermaidCode, `/tmp/${cacheKey}.svg`);

 diagrams.set(cacheKey, outputPath);
 res.sendFile(outputPath);
});
```

Strategy 2: Client-Side Rendering (CSR)
```typescript
// React component with lazy loading
import mermaid from 'mermaid';

export const MermaidDiagram = ({ code, theme = 'default' }: Props) => {
 const ref = useRef<HTMLDivElement>(null);

 useEffect(() => {
 // Initialize mermaid with optimized settings
 mermaid.initialize({
 startOnLoad: false,
 theme: theme,
 securityLevel: 'antiscript', // Prevent XSS
 flowchart: { useMaxWidth: true },
 // Disable animations for better performance
 gantt: { useWidth: undefined }
 });

 // Render only visible diagrams
 if (ref.current && isInViewport(ref.current)) {
 mermaid.contentLoaderAsync();
 }
 }, [code, theme]);

 return <div ref={ref} className="mermaid">{code}</div>;
};

// Intersection Observer for lazy loading
const useVisibilityObserver = (ref) => {
 const [isVisible, setIsVisible] = useState(false);

 useEffect(() => {
 const observer = new IntersectionObserver(([entry]) => {
 if (entry.isIntersecting) {
 setIsVisible(true);
 observer.unobserve(entry.target);
 }
 }, { threshold: 0.1 });

 if (ref.current) observer.observe(ref.current);
 return () => observer.disconnect();
 }, [ref]);

 return isVisible;
};
```

Strategy 3: Hybrid Rendering
```typescript
// Use SSR for complex diagrams, CSR for simple ones
interface DiagramConfig {
 complexity: 'simple' | 'medium' | 'complex';
 data: string;
 format: 'svg' | 'png';
}

async function renderDiagram(config: DiagramConfig) {
 if (config.complexity === 'simple') {
 // Client-side for fast rendering
 return renderClientSide(config.data);
 } else if (config.complexity === 'complex') {
 // Server-side for reliability
 return await renderServerSide(config.data, config.format);
 } else {
 // Hybrid: SSR with caching
 const cached = await checkCache(config.data);
 if (cached) return cached;

 const svg = await renderServerSide(config.data, 'svg');
 await cache(config.data, svg);
 return svg;
 }
}
```

## Caching Strategies

### 1. Multi-Layer Caching

```typescript
class DiagramCache {
 private memoryCache = new Map<string, CacheEntry>();
 private redisClient = createRedisClient();
 private S3Bucket = createS3Bucket();

 async get(diagramId: string): Promise<SVG | null> {
 // Layer 1: Memory cache (milliseconds)
 let cached = this.memoryCache.get(diagramId);
 if (cached && !this.isExpired(cached)) {
 metrics.increment('cache.memory.hit');
 return cached.svg;
 }

 // Layer 2: Redis cache (seconds/minutes)
 try {
 cached = await this.redisClient.get(diagramId);
 if (cached) {
 metrics.increment('cache.redis.hit');
 // Populate memory cache
 this.memoryCache.set(diagramId, cached);
 return cached.svg;
 }
 } catch (error) {
 logger.warn('Redis cache miss', { diagramId });
 }

 // Layer 3: S3 cache (permanent)
 try {
 cached = await this.S3Bucket.get(`diagrams/${diagramId}.svg`);
 if (cached) {
 metrics.increment('cache.s3.hit');
 // Populate Redis and memory
 await this.redisClient.set(diagramId, cached, 3600); // 1 hour TTL
 this.memoryCache.set(diagramId, cached);
 return cached.svg;
 }
 } catch (error) {
 logger.warn('S3 cache miss', { diagramId });
 }

 // Cache miss - generate new
 metrics.increment('cache.miss');
 return null;
 }

 async set(diagramId: string, svg: SVG, ttl: number = 3600) {
 const entry = { svg, timestamp: Date.now(), ttl };

 // All layers
 this.memoryCache.set(diagramId, entry);
 await this.redisClient.set(diagramId, entry, ttl);
 await this.S3Bucket.put(`diagrams/${diagramId}.svg`, svg);
 }

 invalidate(diagramId: string) {
 this.memoryCache.delete(diagramId);
 this.redisClient.del(diagramId);
 // Keep S3 for history
 }
}
```

### 2. Cache Invalidation Strategies

```typescript
// Event-based cache invalidation
class CacheInvalidator {
 async invalidateOnDiagramUpdate(diagramId: string) {
 // Publish invalidation event
 await eventBus.publish('diagram.updated', { diagramId });

 // Subscribe to updates
 eventBus.subscribe('diagram.updated', async (event) => {
 await this.cache.invalidate(event.diagramId);

 // Pre-generate for frequently accessed diagrams
 if (await this.isFavorite(event.diagramId)) {
 await this.preGenerateSVG(event.diagramId);
 }
 });
 }

 // Time-based TTL
 setTTL(diagramId: string, category: string) {
 const ttls = {
 'architecture': 86400, // 24 hours
 'sequence': 3600, // 1 hour
 'temporary': 600, // 10 minutes
 'frequently_accessed': 0 // Never expire
 };
 return ttls[category] || 3600;
 }
}
```

## SVG Optimization

### 1. SVG Size Reduction

```typescript
import { minify } from 'svgo';

class SVGOptimizer {
 async optimize(svg: string): Promise<string> {
 const optimized = minify(svg, {
 plugins: [
 {
 name: 'preset-default',
 params: {
 overrides: {
 cleanupIds: { minify: true },
 removeViewBox: false, // Keep for responsiveness
 removeHiddenElems: true,
 removeEmptyContainers: true,
 convertStyleToAttrs: true,
 removeEmptyAttrs: true,
 removeUnknownsAndDefaults: true,
 removeUselessDefs: true,
 removeUselessStrokeAndFill: true,
 }
 }
 },
 'convertColors',
 'removeDoctype'
 ]
 });

 return optimized.data;
 }

 // Measure compression ratio
 measureCompression(original: string, optimized: string) {
 const ratio = ((original.length - optimized.length) / original.length * 100).toFixed(2);
 return {
 original: `${(original.length / 1024).toFixed(2)}KB`,
 optimized: `${(optimized.length / 1024).toFixed(2)}KB`,
 saved: `${ratio}%`
 };
 }
}

// Usage
const optimizer = new SVGOptimizer();
const original = generateMermaidSVG(code);
const optimized = await optimizer.optimize(original);
const stats = optimizer.measureCompression(original, optimized);
console.log(`Optimized: ${stats.saved}% reduction`);
```

### 2. Responsive SVG

```typescript
function createResponsiveSVG(width: number, height: number) {
 return `
 <svg
 viewBox="0 0 ${width} ${height}"
 preserveAspectRatio="xMidYMid meet"
 class="mermaid-responsive"
 style="width: 100%; height: auto; max-width: 100%;"
 >
 <!-- diagram content -->
 </svg>
 `;
}

// CSS for responsive behavior
const styles = `
 .mermaid-responsive {
 display: block;
 margin: 0 auto;
 }

 @media (max-width: 768px) {
 .mermaid-responsive {
 font-size: 12px;
 }

 .mermaid-node text {
 font-size: 10px;
 }
 }
`;
```

## Performance Monitoring

### 1. Metrics Collection

```typescript
class DiagramMetrics {
 async trackRendering(diagramId: string, code: string) {
 const startTime = performance.now();

 const svg = await renderDiagram(code);

 const duration = performance.now() - startTime;
 const nodeCount = svg.match(/<g class="node"/g)?.length || 0;
 const edgeCount = svg.match(/<line/g)?.length || 0;
 const svgSize = new Blob([svg]).size;

 // Send to metrics service
 await metrics.record({
 type: 'diagram.render',
 diagramId,
 duration_ms: duration,
 node_count: nodeCount,
 edge_count: edgeCount,
 svg_size_bytes: svgSize,
 complexity: this.calculateComplexity(nodeCount, edgeCount),
 timestamp: new Date().toISOString()
 });

 return {
 duration_ms: duration,
 metrics: { nodeCount, edgeCount, svgSize }
 };
 }

 calculateComplexity(nodes: number, edges: number): string {
 const score = nodes + edges;
 if (score < 20) return 'simple';
 if (score < 50) return 'medium';
 if (score < 100) return 'complex';
 return 'very-complex';
 }
}
```

### 2. Performance Budgets

```typescript
interface PerformanceBudget {
 renderTime_ms: number;
 svgSize_bytes: number;
 nodeCount: number;
 memory_mb: number;
}

const budgets: Record<string, PerformanceBudget> = {
 'simple': {
 renderTime_ms: 100,
 svgSize_bytes: 10000,
 nodeCount: 20,
 memory_mb: 10
 },
 'medium': {
 renderTime_ms: 500,
 svgSize_bytes: 100000,
 nodeCount: 50,
 memory_mb: 50
 },
 'complex': {
 renderTime_ms: 2000,
 svgSize_bytes: 500000,
 nodeCount: 150,
 memory_mb: 200
 }
};

async function validatePerformance(diagramId: string, metrics: DiagramMetrics) {
 const complexity = metrics.complexity;
 const budget = budgets[complexity];

 const issues = [];
 if (metrics.duration_ms > budget.renderTime_ms) {
 issues.push(`Render time exceeded: ${metrics.duration_ms}ms > ${budget.renderTime_ms}ms`);
 }
 if (metrics.svgSize_bytes > budget.svgSize_bytes) {
 issues.push(`SVG size exceeded: ${metrics.svgSize_bytes}B > ${budget.svgSize_bytes}B`);
 }

 if (issues.length > 0) {
 logger.warn('Performance budget exceeded', { diagramId, issues });
 await metrics.alert('diagram.performance.budget_exceeded', { diagramId, issues });
 }

 return issues.length === 0;
}
```

## Optimization Checklist

- [ ] Caching: Implement multi-layer caching (memory → Redis → S3)
- [ ] SVG Optimization: Minify SVG using SVGO with proper configuration
- [ ] Lazy Loading: Use Intersection Observer for diagrams below fold
- [ ] Code Splitting: Load mermaid library asynchronously
- [ ] Performance Monitoring: Track render times and SVG sizes
- [ ] Responsive Design: Use viewBox and preserveAspectRatio for scaling
- [ ] Content Delivery: Use CDN for cached SVG files
- [ ] Browser Compatibility: Test with Chrome, Firefox, Safari, Edge

## Best Practices for Large-Scale Deployments

1. Pre-generate Common Diagrams: Generate popular diagrams during off-peak hours
2. CDN Distribution: Cache SVGs on CDN for global delivery
3. Rate Limiting: Limit diagram generation requests per user/IP
4. Queue Processing: Use job queues for complex diagram generation
5. Error Recovery: Implement graceful degradation (show placeholder if rendering fails)
6. Monitoring: Alert on rendering failures or performance degradation

---

Version: 6.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
