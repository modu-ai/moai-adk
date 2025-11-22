---
name: moai-readme-expert/optimization
description: Performance optimization for README generation, caching, and distribution
---

# README Expert Optimization (v4.0.0)

## Generation Optimization

### 1. Incremental README Updates

```typescript
class IncrementalReadmeGenerator {
    private readmeCache: Map<string, ReadmeVersion> = new Map();

    async updateReadme(
        projectId: string,
        changes: ProjectChange[]
    ): Promise<string> {
        const currentReadme = this.readmeCache.get(projectId);

        if (!currentReadme) {
            // No cached version, generate from scratch
            return await this.generateFull(projectId);
        }

        // Identify which sections need updating
        const changedSections = this.identifyChangedSections(
            changes,
            currentReadme
        );

        if (changedSections.length === 0) {
            // No changes, return cached version
            return currentReadme.content;
        }

        // Incrementally update affected sections
        const updatedReadme = { ...currentReadme };

        for (const section of changedSections) {
            updatedReadme.sections[section] = await this.generateSection(
                projectId,
                section
            );
        }

        // Recalculate table of contents only if structure changed
        if (changedSections.includes('structure')) {
            updatedReadme.toc = this.generateTOC(updatedReadme.content);
        }

        return updatedReadme.content;
    }

    private identifyChangedSections(
        changes: ProjectChange[],
        currentReadme: ReadmeVersion
    ): string[] {
        const sections = new Set<string>();

        for (const change of changes) {
            if (change.type === 'version-update') {
                sections.add('installation');
                sections.add('badges');
            } else if (change.type === 'dependency-add') {
                sections.add('dependencies');
            } else if (change.type === 'test-coverage-change') {
                sections.add('testing');
                sections.add('badges');
            } else if (change.type === 'contributor-add') {
                sections.add('contributors');
            }
        }

        return Array.from(sections);
    }
}
```

### 2. Parallel Section Generation

```typescript
class ParallelReadmeGenerator {
    async generateParallel(projectId: string): Promise<string> {
        // Identify independent sections
        const sections = [
            'title',
            'badges',
            'installation',
            'quickstart',
            'features',
            'api',
            'examples',
            'testing',
            'contributing'
        ];

        // Generate sections in parallel
        const promises = sections.map(section =>
            this.generateSection(projectId, section)
        );

        const results = await Promise.all(promises);

        // Combine results
        return results.join('\n\n');
    }
}
```

## Caching Strategy

### 1. Multi-Layer Cache

```typescript
class ReadmeCache {
    private memoryCache = new Map<string, CacheEntry>();
    private redisClient: Redis;
    private s3Client: S3Client;

    async getReadme(projectId: string): Promise<string | null> {
        // Layer 1: Memory cache
        let cached = this.memoryCache.get(projectId);
        if (cached && !this.isExpired(cached)) {
            return cached.content;
        }

        // Layer 2: Redis cache
        try {
            cached = await this.redisClient.get(`readme:${projectId}`);
            if (cached) {
                this.memoryCache.set(projectId, cached);
                return cached.content;
            }
        } catch (error) {
            logger.warn('Redis cache miss', { projectId });
        }

        // Layer 3: S3 cache
        try {
            const s3Object = await this.s3Client.getObject({
                Bucket: 'readme-cache',
                Key: `${projectId}/README.md`
            });

            if (s3Object) {
                cached = JSON.parse(s3Object.Body.toString());
                this.redisClient.set(`readme:${projectId}`, JSON.stringify(cached), 'EX', 7200);
                this.memoryCache.set(projectId, cached);
                return cached.content;
            }
        } catch (error) {
            logger.warn('S3 cache miss', { projectId });
        }

        return null;
    }

    async cache(projectId: string, readme: string) {
        const entry = {
            content: readme,
            timestamp: Date.now(),
            hash: this.calculateHash(readme)
        };

        // Cache to all layers
        this.memoryCache.set(projectId, entry);
        await this.redisClient.set(
            `readme:${projectId}`,
            JSON.stringify(entry),
            'EX',
            7200 // 2 hours
        );

        await this.s3Client.putObject({
            Bucket: 'readme-cache',
            Key: `${projectId}/README.md`,
            Body: JSON.stringify(entry),
            Metadata: { generated: new Date().toISOString() }
        });
    }

    private calculateHash(content: string): string {
        return crypto.createHash('sha256').update(content).digest('hex');
    }
}
```

### 2. Cache Invalidation

```typescript
class ReadmeCacheInvalidation {
    async invalidateOnChange(projectId: string, changeType: string) {
        const cacheManager = this.getCacheManager(projectId);

        switch (changeType) {
            case 'version-update':
                // Invalidate version-dependent caches
                await cacheManager.invalidate('badges', 'installation');
                break;
            case 'documentation-update':
                // Invalidate API docs cache
                await cacheManager.invalidate('api', 'examples');
                break;
            case 'contributor-update':
                // Minimal invalidation
                await cacheManager.invalidate('contributors');
                break;
            default:
                // Full invalidation
                await cacheManager.invalidateAll(projectId);
        }
    }
}
```

## Distribution Optimization

### 1. CDN Distribution

```typescript
class ReadmeCDN {
    async publishReadme(projectId: string, content: string) {
        // Publish to multiple CDN edge locations
        const regions = [
            'us-east-1',
            'eu-west-1',
            'ap-southeast-1',
            'ap-northeast-1'
        ];

        const promises = regions.map(region =>
            this.publishToRegion(projectId, content, region)
        );

        await Promise.all(promises);
    }

    private async publishToRegion(
        projectId: string,
        content: string,
        region: string
    ) {
        // Minify markdown
        const minified = this.minifyMarkdown(content);

        // Compress with gzip
        const compressed = await gzip(minified);

        // Upload to regional CloudFront distribution
        await this.cloudfront[region].invalidate({
            DistributionId: this.distributionId,
            InvalidationBatch: {
                Paths: {
                    Quantity: 1,
                    Items: [`/${projectId}/README.md`]
                }
            }
        });
    }

    private minifyMarkdown(content: string): string {
        // Remove extra whitespace
        return content
            .split('\n')
            .map(line => line.trim())
            .filter(line => line.length > 0)
            .join('\n');
    }
}
```

## Performance Monitoring

### 1. Generation Metrics

```typescript
class ReadmeMetrics {
    async trackGeneration(projectId: string, sections: string[]) {
        const startTime = performance.now();

        try {
            const readme = await this.generateReadme(projectId, sections);

            const duration = performance.now() - startTime;
            const lines = readme.split('\n').length;

            logger.info('README generation', {
                project_id: projectId,
                sections_count: sections.length,
                output_lines: lines,
                duration_ms: duration,
                timestamp: new Date()
            });

            // Alert if slow
            if (duration > 5000) {
                metrics.increment('generation.slow');
                logger.warn('Slow README generation', { duration_ms: duration });
            }
        } catch (error) {
            metrics.increment('generation.error');
            logger.error('README generation failed', { projectId, error });
            throw error;
        }
    }
}
```

---

**Version**: 4.0.0 | **Last Updated**: 2025-11-22 | **Enterprise Ready**: ✓
