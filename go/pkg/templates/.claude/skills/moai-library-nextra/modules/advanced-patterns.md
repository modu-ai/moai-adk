---
name: moai-library-nextra/advanced-patterns
description: Advanced Nextra architecture patterns, MDX integration, dynamic routing, and site optimization
---

# Advanced Nextra Architecture Patterns (v5.0.0)

## Advanced MDX Integration

### 1. Custom MDX Components System

```typescript
import { MDXProvider } from '@mdx-js/react';
import { useMemo } from 'react';

interface MDXComponentConfig {
 name: string;
 component: React.ComponentType<any>;
 props?: Record<string, any>;
 metadata?: ComponentMetadata;
}

interface ComponentMetadata {
 version: string;
 description: string;
 deprecated?: boolean;
 replacedBy?: string;
}

class MDXComponentRegistry {
 private components: Map<string, MDXComponentConfig> = new Map();
 private componentCache: Map<string, React.ComponentType> = new Map();

 registerComponent(config: MDXComponentConfig) {
 this.components.set(config.name, config);
 // Clear cache when registering new component
 this.componentCache.delete(config.name);
 }

 getComponentForMDX(name: string) {
 if (this.componentCache.has(name)) {
 return this.componentCache.get(name)!;
 }

 const config = this.components.get(name);
 if (!config) {
 throw new Error(`Component not registered: ${name}`);
 }

 // Wrap with metadata warning if deprecated
 let component = config.component;

 if (config.metadata?.deprecated) {
 component = (props) => {
 console.warn(
 `Deprecated component: ${name}. Use ${config.metadata?.replacedBy} instead`
 );
 return <config.component {...props} />;
 };
 }

 this.componentCache.set(name, component);
 return component;
 }

 getMDXProviderValue() {
 const components: Record<string, React.ComponentType<any>> = {
 // Override HTML elements
 h1: CustomHeading1,
 h2: CustomHeading2,
 h3: CustomHeading3,
 p: CustomParagraph,
 a: CustomLink,
 code: CustomCode,
 pre: CustomPre,

 // Custom components
 ...Object.fromEntries(this.components.entries())
 };

 return components;
 }
}

// Usage in Nextra pages
export function AdvancedMDXPage({ children }: { children: React.ReactNode }) {
 const registry = useMemo(() => new MDXComponentRegistry(), []);

 registry.registerComponent({
 name: 'Alert',
 component: AlertComponent,
 metadata: { version: '1.0.0', description: 'Alert notification' }
 });

 const mdxComponents = registry.getMDXProviderValue();

 return (
 <MDXProvider components={mdxComponents}>
 {children}
 </MDXProvider>
 );
}
```

## Dynamic Routing and Nested Structure

### 1. Hierarchical Site Structure

```typescript
interface SiteNode {
 title: string;
 path: string;
 children?: SiteNode[];
 frontmatter?: Record<string, any>;
 meta?: {
 display?: 'hidden' | 'normal';
 sort?: number;
 href?: string;
 target?: string;
 search?: boolean;
 };
}

class NextraArchitecture {
 private siteTree: SiteNode[] = [];

 // Build site structure from file system
 async buildSiteTree(pagesDir: string): Promise<SiteNode[]> {
 const tree: SiteNode[] = [];

 const files = await this.scanDirectory(pagesDir);

 for (const file of files) {
 const node = await this.createNode(file, pagesDir);
 tree.push(node);
 }

 // Sort by metadata
 return this.sortNodes(tree);
 }

 private async createNode(filePath: string, baseDir: string): Promise<SiteNode> {
 const relativePath = path.relative(baseDir, filePath);
 const urlPath = '/' + relativePath.replace(/\.[^/.]+$/, '').replace(/\\g, '/');

 // Extract frontmatter
 const content = await fs.readFile(filePath, 'utf-8');
 const frontmatter = this.extractFrontmatter(content);
 const title = frontmatter.title || this.inferTitle(relativePath);

 return {
 title,
 path: urlPath,
 frontmatter,
 meta: {
 display: frontmatter.display || 'normal',
 sort: frontmatter.sort || 0,
 search: frontmatter.search !== false
 }
 };
 }

 private sortNodes(nodes: SiteNode[]): SiteNode[] {
 // Sort by display priority and custom sort
 return nodes.sort((a, b) => {
 // Hidden items last
 if (a.meta?.display === 'hidden') return 1;
 if (b.meta?.display === 'hidden') return -1;

 // Custom sort
 const aSort = a.meta?.sort || 0;
 const bSort = b.meta?.sort || 0;

 return bSort - aSort;
 });
 }

 private extractFrontmatter(content: string): Record<string, any> {
 const match = content.match(/^---\n([\s\S]*?)\n---/);
 if (!match) return {};

 // Parse YAML frontmatter
 return this.parseYAML(match[1]);
 }
}
```

### 2. Dynamic Import and Code Splitting

```typescript
import dynamic from 'next/dynamic';

// Split docs by section
const DocumentationRoutes = {
 getting_started: dynamic(() => import('./docs/getting-started.mdx')),
 api_reference: dynamic(() => import('./docs/api-reference.mdx')),
 examples: dynamic(() => import('./docs/examples.mdx')),
 faq: dynamic(() => import('./docs/faq.mdx'))
};

export function DynamicDocumentationPage({ section }: { section: string }) {
 const Component = DocumentationRoutes[section as keyof typeof DocumentationRoutes];

 if (!Component) {
 return <NotFound section={section} />;
 }

 return (
 <Suspense fallback={<LoadingSpinner />}>
 <Component />
 </Suspense>
 );
}
```

## Search and Navigation

### 1. Full-Text Search Integration

```typescript
import Fuse from 'fuse.js';

interface SearchableContent {
 id: string;
 title: string;
 content: string;
 path: string;
 section: string;
 keywords: string[];
}

class DocumentationSearch {
 private index: Fuse<SearchableContent>;
 private documents: SearchableContent[] = [];

 async indexDocumentation(pages: SiteNode[]): Promise<void> {
 this.documents = await this.extractSearchableContent(pages);

 // Create Fuse index for fast full-text search
 this.index = new Fuse(this.documents, {
 keys: ['title', 'content', 'keywords'],
 includeScore: true,
 threshold: 0.3, // Fuzzy matching
 minMatchCharLength: 2
 });
 }

 async search(query: string, limit: number = 10): Promise<SearchResult[]> {
 const results = this.index.search(query).slice(0, limit);

 return results.map(result => ({
 ...result.item,
 relevance: 1 - (result.score || 0)
 }));
 }

 private async extractSearchableContent(
 pages: SiteNode[]
 ): Promise<SearchableContent[]> {
 const content: SearchableContent[] = [];

 for (const page of pages) {
 const mdxContent = await this.readMDXContent(page.path);
 const extracted = this.extractFromMDX(mdxContent);

 content.push({
 id: page.path,
 title: page.title,
 content: extracted.text,
 path: page.path,
 section: this.inferSection(page.path),
 keywords: extracted.keywords
 });
 }

 return content;
 }

 private extractFromMDX(content: string) {
 // Remove code blocks and MDX components
 const text = content
 .replace(/```[\s\S]*?```/g, '')
 .replace(/<[^>]+>/g, '');

 // Extract keywords from headings
 const headings = content.match(/^#+\s+.+$/gm) || [];
 const keywords = headings.map(h => h.replace(/^#+\s+/, ''));

 return { text, keywords };
 }
}

// React search component
export function SearchBar() {
 const [query, setQuery] = useState('');
 const [results, setResults] = useState<SearchResult[]>([]);
 const search = useSearch(); // Hook to access search instance

 const handleSearch = async (q: string) => {
 setQuery(q);
 const searchResults = await search(q);
 setResults(searchResults);
 };

 return (
 <div className="search-container">
 <input
 type="search"
 placeholder="Search documentation..."
 value={query}
 onChange={(e) => handleSearch(e.target.value)}
 className="search-input"
 />
 {results.length > 0 && (
 <ul className="search-results">
 {results.map(result => (
 <li key={result.id} className="search-result">
 <Link href={result.path}>
 <h3>{result.title}</h3>
 <p className="excerpt">{result.content.substring(0, 100)}...</p>
 <span className="relevance">
 {(result.relevance * 100).toFixed(0)}% match
 </span>
 </Link>
 </li>
 ))}
 </ul>
 )}
 </div>
 );
}
```

---

Version: 5.0.0 | Last Updated: 2025-11-22 | Enterprise Ready:
