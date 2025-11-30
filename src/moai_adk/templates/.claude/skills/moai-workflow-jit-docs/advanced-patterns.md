---
name: moai-workflow-jit-docs/advanced-patterns
description: Advanced JIT documentation patterns, intelligent doc discovery, lazy loading, and content optimization
---

# Advanced JIT Documentation Patterns (v5.0.0)

## Intelligent Document Discovery

### 1. Smart Doc Inference Engine

```typescript
interface DocumentPattern {
 name: string;
 filePattern: RegExp;
 contentPattern: RegExp;
 priority: number;
 extractor: (content: string) => string;
}

class IntelligentDocDiscovery {
 private patterns: DocumentPattern[] = [
 {
 name: 'README',
 filePattern: /README\.(md|mdx|txt)$/i,
 contentPattern: /^#\s+/,
 priority: 100,
 extractor: (content) => content.split('\n').slice(0, 20).join('\n')
 },
 {
 name: 'API Documentation',
 filePattern: /\/api\//i,
 contentPattern: /\/\*\*[\s\S]*?\*\/|\/\/\/.*$/m,
 priority: 90,
 extractor: (content) => this.extractJSDocComments(content)
 },
 {
 name: 'Getting Started',
 filePattern: /(getting-started|quick-start|introduction)\.(md|mdx)/i,
 contentPattern: /##\s+(Getting Started|Quick Start|Introduction)/i,
 priority: 80,
 extractor: (content) => content.substring(0, 2000)
 },
 {
 name: 'Configuration',
 filePattern: /(config|configure|settings)\.(md|mdx|json|yaml|yml)/i,
 contentPattern: /^#+\s+(Configuration|Settings|Config)/im,
 priority: 70,
 extractor: (content) => this.extractConfiguration(content)
 },
 {
 name: 'Examples',
 filePattern: /(examples?|samples?)\/(.*)\.(js|ts|jsx|tsx|py)/i,
 contentPattern: /\/\*\*|\/\/|'''|"""/,
 priority: 60,
 extractor: (content) => content.substring(0, 1000)
 }
 ];

 async discoverDocs(projectRoot: string): Promise<DiscoveredDoc[]> {
 const docs: DiscoveredDoc[] = [];

 // Scan entire project
 const allFiles = await this.scanProject(projectRoot);

 for (const file of allFiles) {
 const content = await fs.readFile(file, 'utf-8');

 // Try to match against patterns
 for (const pattern of this.patterns) {
 if (pattern.filePattern.test(file) && pattern.contentPattern.test(content)) {
 docs.push({
 type: pattern.name,
 path: file,
 content: pattern.extractor(content),
 priority: pattern.priority,
 relevance: this.calculateRelevance(file, content)
 });

 break; // Move to next file
 }
 }
 }

 // Sort by priority and relevance
 return docs.sort((a, b) => {
 const priorityDiff = b.priority - a.priority;
 return priorityDiff !== 0 ? priorityDiff : b.relevance - a.relevance;
 });
 }

 private calculateRelevance(filePath: string, content: string): number {
 let score = 0;

 // Filename relevance
 const filename = path.basename(filePath);
 if (filename.includes('readme')) score += 30;
 if (filename.includes('guide')) score += 20;
 if (filename.includes('tutorial')) score += 15;

 // Content metrics
 const lines = content.split('\n').length;
 if (lines > 100) score += 10; // Substantial content
 if (lines < 5) score -= 10; // Too brief

 // Freshness (modified recently)
 const stats = fs.statSync(filePath);
 const daysSinceModified = (Date.now() - stats.mtime.getTime()) / (1000 * 60 * 60 * 24);
 if (daysSinceModified < 7) score += 15;
 if (daysSinceModified < 30) score += 10;

 return Math.max(0, score);
 }

 private extractJSDocComments(content: string): string {
 const comments = content.match(/\/\*\*[\s\S]*?\*\//g) || [];
 return comments.slice(0, 10).join('\n\n');
 }

 private extractConfiguration(content: string): string {
 // Extract key config sections
 const lines = content.split('\n');
 const sections: string[] = [];

 let inConfig = false;
 let currentSection = '';

 for (const line of lines) {
 if (line.match(/^#+\s+(Configuration|Settings|Config)/i)) {
 inConfig = true;
 } else if (inConfig && line.match(/^#+\s+(?!Configuration|Settings|Config)/)) {
 break;
 }

 if (inConfig) {
 currentSection += line + '\n';
 if (currentSection.split('\n').length > 50) {
 sections.push(currentSection);
 currentSection = '';
 }
 }
 }

 return sections.join('\n\n');
 }
}
```

## Lazy Loading Strategies

### 1. On-Demand Documentation Loading

```typescript
class LazyDocumentationLoader {
 private docCache: Map<string, CachedDoc> = new Map();
 private loadingQueue: Set<string> = new Set();

 async loadDocLazy(docId: string, priority: 'high' | 'normal' | 'low' = 'normal'): Promise<DocumentContent> {
 // Check cache first
 const cached = this.docCache.get(docId);
 if (cached && !this.isExpired(cached)) {
 return cached.content;
 }

 // Prevent duplicate loads
 if (this.loadingQueue.has(docId)) {
 // Wait for ongoing load
 return this.waitForLoad(docId);
 }

 this.loadingQueue.add(docId);

 try {
 const content = await this.loadDocumentContent(docId);

 // Cache with priority-based TTL
 const ttl = priority === 'high' ? 3600000 : priority === 'normal' ? 1800000 : 600000;

 this.docCache.set(docId, {
 content,
 timestamp: Date.now(),
 ttl,
 size: Buffer.byteLength(JSON.stringify(content))
 });

 return content;
 } finally {
 this.loadingQueue.delete(docId);
 }
 }

 // Batch lazy loading for multiple docs
 async loadDocsBatch(docIds: string[]): Promise<Map<string, DocumentContent>> {
 const results = new Map<string, DocumentContent>();

 // Load high-priority items first
 const highPriority = docIds.slice(0, 3);
 const restPriority = docIds.slice(3);

 // Load top 3 in parallel
 const topResults = await Promise.all(
 highPriority.map(id => this.loadDocLazy(id, 'high'))
 );

 topResults.forEach((content, i) => {
 results.set(highPriority[i], content);
 });

 // Queue rest for background loading
 for (const docId of restPriority) {
 this.loadDocLazy(docId, 'low').then(content => {
 results.set(docId, content);
 });
 }

 return results;
 }

 private async loadDocumentContent(docId: string): Promise<DocumentContent> {
 // Load from disk/database
 const metadata = await this.getDocMetadata(docId);
 const content = await this.readDocument(metadata.path);

 return {
 id: docId,
 title: metadata.title,
 content,
 metadata: metadata
 };
 }

 private isExpired(cached: CachedDoc): boolean {
 return Date.now() - cached.timestamp > cached.ttl;
 }
}
```

### 2. Progressive Enhancement

```typescript
class ProgressiveDocumentEnhancer {
 async enhanceDocumentProgressive(docId: string): Promise<EnhancedDocument> {
 // Step 1: Load and parse immediately
 const basic = await this.loadBasicDocument(docId);

 // Step 2: Load metadata asynchronously
 const withMetadata = await this.addMetadata(basic);

 // Step 3: Process code examples asynchronously
 const withExamples = await this.processCodeExamples(withMetadata);

 // Step 4: Generate search index asynchronously
 const withSearch = await this.generateSearchIndex(withExamples);

 // Step 5: Generate related links asynchronously
 const enhanced = await this.generateRelatedLinks(withSearch);

 return enhanced;
 }

 // Return document as it loads (streaming approach)
 async* streamDocumentEnhancement(docId: string) {
 // Step 1: Basic content
 const basic = await this.loadBasicDocument(docId);
 yield { stage: 'basic', document: basic };

 // Step 2: Metadata
 const withMetadata = await this.addMetadata(basic);
 yield { stage: 'metadata', document: withMetadata };

 // Step 3: Examples
 const withExamples = await this.processCodeExamples(withMetadata);
 yield { stage: 'examples', document: withExamples };

 // Step 4: Search
 const withSearch = await this.generateSearchIndex(withExamples);
 yield { stage: 'search', document: withSearch };

 // Step 5: Related
 const enhanced = await this.generateRelatedLinks(withSearch);
 yield { stage: 'complete', document: enhanced };
 }
}

// React hook for progressive loading
export function useProgressiveDoc(docId: string) {
 const [document, setDocument] = useState<DocumentContent>();
 const [stage, setStage] = useState<'loading' | 'basic' | 'enhanced' | 'complete'>('loading');

 useEffect(() => {
 const enhancer = new ProgressiveDocumentEnhancer();
 const generator = enhancer.streamDocumentEnhancement(docId);

 (async () => {
 for await (const { stage, document } of generator) {
 setDocument(document);
 setStage(stage as any);
 }
 })();
 }, [docId]);

 return { document, stage };
}
```

## Intelligent Content Recommendation

### 1. Related Documentation Discovery

```typescript
class RelatedDocumentFinder {
 async findRelated(currentDocId: string, limit: number = 5): Promise<RelatedDoc[]> {
 const current = await this.getDocument(currentDocId);
 const allDocs = await this.getAllDocuments();

 // Extract features from current doc
 const features = this.extractFeatures(current);

 // Score all other docs
 const scored = allDocs
 .filter(doc => doc.id !== currentDocId)
 .map(doc => ({
 doc,
 score: this.calculateSimilarity(features, this.extractFeatures(doc))
 }));

 // Return top N
 return scored
 .sort((a, b) => b.score - a.score)
 .slice(0, limit)
 .map(s => ({
 ...s.doc,
 relevance: s.score
 }));
 }

 private extractFeatures(doc: DocumentContent): DocumentFeatures {
 return {
 keywords: this.extractKeywords(doc.content),
 tags: doc.metadata?.tags || [],
 section: doc.metadata?.section,
 length: doc.content.length,
 codeExamples: (doc.content.match(/```/g) || []).length / 2
 };
 }

 private calculateSimilarity(features1: DocumentFeatures, features2: DocumentFeatures): number {
 let score = 0;

 // Keyword overlap (40%)
 const keywordIntersection = features1.keywords.filter(k =>
 features2.keywords.includes(k)
 ).length;
 score += (keywordIntersection / Math.max(1, features1.keywords.length)) * 0.4;

 // Tag overlap (30%)
 const tagIntersection = features1.tags.filter(t =>
 features2.tags.includes(t)
 ).length;
 score += (tagIntersection / Math.max(1, features1.tags.length)) * 0.3;

 // Same section (20%)
 if (features1.section === features2.section) {
 score += 0.2;
 }

 // Similar complexity (10%)
 const lengthRatio = Math.min(features1.length, features2.length) /
 Math.max(features1.length, features2.length);
 score += (lengthRatio > 0.8 ? 0.1 : 0);

 return score;
 }
}
```

---

Version: 5.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
