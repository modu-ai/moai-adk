# moai-content-markdown-to-blog

Converting Markdown to WordPress/Naver Blog/Tistory HTML formats with platform-specific styling and metadata.

## Quick Start

Multi-platform publishing allows you to reach audiences across different blogging platforms without rewriting content. Use this skill when converting blog posts to multiple platforms, automating publishing workflows, or maintaining content across platforms.

## Core Patterns

### Pattern 1: Markdown to WordPress HTML Conversion

**Pattern**: Convert Markdown blog posts to WordPress-compatible HTML with custom styling.

```typescript
// lib/markdown-to-wordpress.ts
import { marked } from 'marked';
import { sanitizeHtml } from 'sanitize-html';

interface WordPressPost {
  title: string;
  content: string;
  excerpt: string;
  tags: string[];
  categories: string[];
  featured_image_url: string;
}

export async function convertMarkdownToWordPress(
  markdownContent: string,
  frontmatter: Record<string, any>
): Promise<WordPressPost> {
  // Parse markdown to HTML
  let html = marked(markdownContent);

  // Sanitize and WordPress-safe HTML
  html = sanitizeHtml(html, {
    allowedTags: [
      'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'p', 'br', 'strong', 'em', 'u', 'a',
      'ul', 'ol', 'li', 'blockquote', 'code', 'pre',
      'img', 'figure', 'figcaption', 'table', 'tr', 'td', 'th', 'tbody', 'thead'
    ],
    allowedAttributes: {
      'a': ['href', 'title'],
      'img': ['src', 'alt', 'width', 'height'],
      'code': ['class']
    }
  });

  // Add WordPress-specific classes
  html = html
    .replace(/<p>/g, '<p style="line-height: 1.6;">')
    .replace(/<code>/g, '<code style="background: #f4f4f4; padding: 2px 6px; border-radius: 3px;">')
    .replace(/<pre>/g, '<pre style="background: #282c34; color: #abb2bf; padding: 15px; border-radius: 5px; overflow-x: auto;">')
    .replace(/<blockquote>/g, '<blockquote style="border-left: 4px solid #007cba; margin-left: 0; padding-left: 15px; color: #666;">');

  // Extract excerpt
  const textContent = html.replace(/<[^>]*>/g, '').substring(0, 200);

  return {
    title: frontmatter.title || 'Untitled',
    content: html,
    excerpt: textContent + '...',
    tags: frontmatter.tags || [],
    categories: frontmatter.categories || [],
    featured_image_url: frontmatter.image || '',
  };
}

// WordPress REST API publishing
export async function publishToWordPress(
  post: WordPressPost,
  wpCredentials: {
    siteUrl: string;
    username: string;
    appPassword: string;
  }
) {
  const auth = Buffer.from(
    `${wpCredentials.username}:${wpCredentials.appPassword}`
  ).toString('base64');

  const response = await fetch(
    `${wpCredentials.siteUrl}/wp-json/wp/v2/posts`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Basic ${auth}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        title: post.title,
        content: post.content,
        excerpt: post.excerpt,
        tags: post.tags,
        categories: post.categories,
        status: 'draft', // Publish as draft for review
        featured_media: post.featured_image_url,
      }),
    }
  );

  return response.json();
}
```

**When to use**:
- Automating WordPress publishing
- Batch publishing multiple posts
- Converting blog content format
- Maintaining WordPress REST API integration

**Key benefits**:
- Automate publishing workflow
- Consistent HTML formatting
- Proper WordPress styling
- Easy batch operations

### Pattern 2: Naver Blog (Korean Blog Platform) Publishing

**Pattern**: Convert Markdown to Naver Blog HTML format with platform-specific features.

```typescript
// lib/markdown-to-naver-blog.ts
export interface NaverBlogPost {
  title: string;
  content: string;
  tags: string[];
}

export async function convertMarkdownToNaverBlog(
  markdownContent: string,
  frontmatter: Record<string, any>
): Promise<NaverBlogPost> {
  // Parse markdown
  let html = marked(markdownContent);

  // Naver Blog specific styling
  html = html
    .replace(/<h1>(.*?)<\/h1>/g, '<h3 style="color: #333;">$1</h3>')
    .replace(/<h2>(.*?)<\/h2>/g, '<h4 style="color: #555;">$1</h4>')
    .replace(/<p>(.*?)<\/p>/g, '<p style="line-height: 1.8; color: #333;">$1</p>')
    .replace(/<code>/g, '<code style="background: #f0f0f0; padding: 3px 5px; border-radius: 3px; font-family: \'Courier New\';">')
    .replace(/<pre>/g, '<pre style="background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; border-left: 4px solid #0066cc;">')
    .replace(/<img src="([^"]*)" alt="([^"]*)">/g,
      '<img src="$1" alt="$2" style="max-width: 100%; height: auto; margin: 10px 0;">');

  // Naver Blog wrapping structure
  const wrappedContent = `
<div style="font-family: 'Noto Sans KR', sans-serif;">
  ${html}
</div>
  `.trim();

  return {
    title: frontmatter.title || '제목 없음',
    content: wrappedContent,
    tags: frontmatter.tags || [],
  };
}

// Naver Blog OpenAPI publishing
export async function publishToNaverBlog(
  post: NaverBlogPost,
  credentials: {
    clientId: string;
    clientSecret: string;
    accessToken: string;
    blogId: string;
  }
) {
  const response = await fetch(
    `https://openapi.naver.com/blog/writePost.json`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${credentials.accessToken}`,
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: new URLSearchParams({
        title: post.title,
        contents: post.content,
        tags: post.tags.join(','),
        publish: '0', // Draft mode
      }).toString(),
    }
  );

  return response.json();
}
```

**When to use**:
- Publishing to Korean Naver Blog
- Multi-language blog strategy
- Reaching Korean-speaking audiences
- Automating Naver publishing workflow

**Key benefits**:
- Native support for Naver Blog
- Proper Korean typography handling
- OpenAPI integration
- Batch publishing support

### Pattern 3: Content Synchronization Across Platforms

**Pattern**: Maintain content across multiple platforms with sync automation.

```typescript
// lib/multi-platform-sync.ts
export interface SyncConfig {
  markdown: string;
  frontmatter: Record<string, any>;
  platforms: {
    wordpress?: { enabled: boolean; siteUrl: string };
    naverblog?: { enabled: boolean; blogId: string };
    tistory?: { enabled: boolean; blogId: string };
    medium?: { enabled: boolean; publicationId: string };
    devto?: { enabled: boolean; orgSlug: string };
  };
}

export async function syncPostToAllPlatforms(config: SyncConfig) {
  const results = {
    wordpress: null as any,
    naverblog: null as any,
    tistory: null as any,
    medium: null as any,
    devto: null as any,
  };

  // WordPress
  if (config.platforms.wordpress?.enabled) {
    const wpPost = await convertMarkdownToWordPress(
      config.markdown,
      config.frontmatter
    );
    results.wordpress = await publishToWordPress(wpPost, {
      siteUrl: config.platforms.wordpress.siteUrl,
      username: process.env.WORDPRESS_USERNAME!,
      appPassword: process.env.WORDPRESS_APP_PASSWORD!,
    });
  }

  // Naver Blog
  if (config.platforms.naverblog?.enabled) {
    const naverPost = await convertMarkdownToNaverBlog(
      config.markdown,
      config.frontmatter
    );
    results.naverblog = await publishToNaverBlog(naverPost, {
      clientId: process.env.NAVER_CLIENT_ID!,
      clientSecret: process.env.NAVER_CLIENT_SECRET!,
      accessToken: process.env.NAVER_ACCESS_TOKEN!,
      blogId: config.platforms.naverblog.blogId,
    });
  }

  // Log results
  console.log('Platform Sync Results:', results);

  return results;
}

// CLI usage
export async function publishViaCLI() {
  const markdownFile = process.argv[2];
  const markdown = await fs.readFile(markdownFile, 'utf-8');

  const frontmatter = extractFrontmatter(markdown);
  const content = markdown.replace(/^---[\s\S]*?---/, '').trim();

  await syncPostToAllPlatforms({
    markdown: content,
    frontmatter,
    platforms: {
      wordpress: { enabled: true, siteUrl: 'https://myblog.com' },
      naverblog: { enabled: true, blogId: '12345678' },
      medium: { enabled: false, publicationId: '' },
    },
  });
}
```

**When to use**:
- Publishing to multiple platforms simultaneously
- Maintaining content across platforms
- Automating multi-platform publishing
- Reaching diverse audiences

**Key benefits**:
- One-click multi-platform publishing
- Consistent content across platforms
- Reduced manual work
- Better content distribution

## Progressive Disclosure

### Level 1: Basic Conversion
- Markdown to HTML conversion
- Basic platform formatting
- Simple metadata handling
- Manual platform publishing

### Level 2: Advanced Automation
- Multi-platform conversion
- Platform-specific styling
- OpenAPI integration
- Batch publishing

### Level 3: Expert Scale
- Scheduled publishing
- Analytics tracking per platform
- A/B testing across platforms
- Advanced SEO optimization

## Works Well With

- **Marked.js**: Markdown parsing
- **WordPress REST API**: Publishing integration
- **Naver Blog OpenAPI**: Korean platform integration
- **Medium API**: Tech writing platform
- **Dev.to API**: Developer community
- **Zapier/IFTTT**: Workflow automation

## References

- **WordPress REST API**: https://developer.wordpress.org/rest-api/
- **Naver Blog OpenAPI**: https://developers.naver.com/docs/naver-login/overview/
- **Medium API**: https://github.com/Medium/medium-api-docs
- **Dev.to API**: https://developers.forem.com/api
- **Tistory OpenAPI**: https://tistory.github.io/document-api/
