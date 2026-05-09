---
name: moai-library-nextra/optimization
description: Performance optimization for Nextra sites, caching, static generation, and deployment
---

# Nextra Architecture Optimization (v5.0.0)

## Static Generation and Caching

### 1. Incremental Static Regeneration (ISR)

```typescript
// pages/docs/[...slug].tsx
import { GetStaticProps, GetStaticPaths } from 'next';

interface DocsPageProps {
 content: string;
 meta: PageMetadata;
}

export const getStaticPaths: GetStaticPaths = async () => {
 // Generate paths for all documentation pages
 const pages = await getAllDocumentationPages();

 return {
 paths: pages.map(page => ({
 params: { slug: page.path.split('/').filter(Boolean) }
 })),
 fallback: 'blocking' // Generate on-demand
 };
};

export const getStaticProps: GetStaticProps<DocsPageProps> = async ({ params }) => {
 const slug = Array.isArray(params?.slug) ? params.slug.join('/') : params?.slug;

 try {
 const { content, meta } = await getDocumentationPage(slug);

 return {
 props: { content, meta },
 revalidate: 3600 // ISR: revalidate every hour
 };
 } catch (error) {
 return {
 notFound: true,
 revalidate: 60 // Retry failed pages more frequently
 };
 }
};

export default function DocsPage({ content, meta }: DocsPageProps) {
 return (
 <article>
 <h1>{meta.title}</h1>
 <div dangerouslySetInnerHTML={{ __html: content }} />
 </article>
 );
}
```

### 2. Build-Time Optimization

```typescript
// next.config.js
module.exports = {
 // Optimize images
 images: {
 formats: ['image/avif', 'image/webp'],
 unoptimized: false
 },

 // Compression
 compress: true,
 poweredByHeader: false,

 // SWC minification (faster than Terser)
 swcMinify: true,

 // React optimization
 reactStrictMode: true,

 // Font optimization
 optimizeFonts: true,

 // Experimental features for better performance
 experimental: {
 optimizePackageImports: [
 '@mui/material',
 '@mui/icons-material'
 ]
 },

 // Custom webpack config
 webpack: (config, { isServer }) => {
 // Tree shaking optimization
 config.optimization.usedExports = true;

 return config;
 }
};
```

## Content Caching Strategy

### 1. Multi-Layer Caching

```typescript
class DocumentationCache {
 private fileCache = new Map<string, CacheEntry>();
 private renderCache = new Map<string, CacheEntry>();

 async getDocumentationPage(slug: string): Promise<PageContent> {
 // Layer 1: In-memory render cache (milliseconds)
 const renderCached = this.renderCache.get(slug);
 if (renderCached && !this.isExpired(renderCached)) {
 return renderCached.data;
 }

 // Layer 2: File system cache (seconds)
 const fileCached = this.fileCache.get(slug);
 if (fileCached && !this.isExpired(fileCached)) {
 return fileCached.data;
 }

 // Layer 3: Read from disk
 const content = await this.readFromDisk(slug);

 // Cache to all layers
 this.fileCache.set(slug, {
 data: content,
 timestamp: Date.now(),
 ttl: 3600000 // 1 hour
 });

 this.renderCache.set(slug, {
 data: content,
 timestamp: Date.now(),
 ttl: 60000 // 1 minute
 });

 return content;
 }

 private isExpired(entry: CacheEntry): boolean {
 return Date.now() - entry.timestamp > entry.ttl;
 }
}
```

## Search Performance Optimization

### 1. Index Compression and Caching

```typescript
class OptimizedSearchIndex {
 async buildCompressedIndex(documents: SearchableContent[]): Promise<Buffer> {
 // Create optimized index structure
 const index = {
 documents: documents.map(doc => ({
 id: doc.id,
 title: doc.title,
 tokens: this.tokenize(doc.content)
 })),
 invertedIndex: this.buildInvertedIndex(documents),
 metadata: {
 version: 1,
 generated: Date.now(),
 count: documents.length
 }
 };

 // Compress with brotli (better compression than gzip)
 const json = JSON.stringify(index);
 const compressed = await this.brotliCompress(json);

 // Cache compressed index
 await this.cacheIndex(compressed);

 return compressed;
 }

 private buildInvertedIndex(
 documents: SearchableContent[]
 ): Map<string, number[]> {
 const invertedIndex = new Map<string, number[]>();

 documents.forEach((doc, idx) => {
 const tokens = this.tokenize(doc.content);

 for (const token of tokens) {
 if (!invertedIndex.has(token)) {
 invertedIndex.set(token, []);
 }

 invertedIndex.get(token)!.push(idx);
 }
 });

 return invertedIndex;
 }

 private tokenize(text: string): string[] {
 return text
 .toLowerCase()
 .match(/\b\w+\b/g) || [];
 }

 private async brotliCompress(data: string): Promise<Buffer> {
 return new Promise((resolve, reject) => {
 require('brotli').compress(Buffer.from(data), (err: any, output: Buffer) => {
 if (err) reject(err);
 else resolve(output);
 });
 });
 }
}
```

## Deployment Optimization

### 1. Vercel Deployment Configuration

```typescript
// vercel.json
{
 "buildCommand": "next build",
 "outputDirectory": ".next",
 "devCommand": "next dev",
 "env": {
 "SEARCH_ENABLED": "@search_enabled"
 },
 "functions": {
 "pages/api/": {
 "memory": 256,
 "maxDuration": 30
 }
 },
 "redirects": [
 {
 "source": "/docs/:slug*",
 "destination": "/:slug*"
 }
 ],
 "headers": [
 {
 "source": "/(_next|public)/:path*",
 "headers": [
 {
 "key": "Cache-Control",
 "value": "public, max-age=31536000, immutable"
 }
 ]
 },
 {
 "source": "/:path*",
 "headers": [
 {
 "key": "Cache-Control",
 "value": "public, max-age=3600, s-maxage=86400"
 }
 ]
 }
 ]
}
```

### 2. Performance Monitoring

```typescript
class NextraPerformanceMonitor {
 async trackPageLoad(slug: string, duration: number) {
 metrics.observe('page.load.duration', duration, {
 page: slug,
 bucket: this.getDurationBucket(duration)
 });

 // Alert if exceeds budget
 if (duration > 3000) {
 metrics.increment('page.load.slow');
 logger.warn('Slow page load', { slug, duration });
 }
 }

 async trackBuildTime(buildDuration: number) {
 metrics.observe('build.duration', buildDuration);

 if (buildDuration > 600000) { // 10 minutes
 metrics.increment('build.slow');
 logger.warn('Slow build detected', { buildDuration });
 }
 }

 private getDurationBucket(duration: number): string {
 if (duration < 1000) return 'fast';
 if (duration < 3000) return 'normal';
 return 'slow';
 }
}
```

---

Version: 5.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
