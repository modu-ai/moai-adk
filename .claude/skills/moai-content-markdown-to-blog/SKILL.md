---
name: moai-content-markdown-to-blog
description: Converting Markdown to WordPress/Naver Blog/Tistory HTML formats. Parse Markdown, transform syntax, handle images, and add platform-specific metadata. Use when converting content between formats or automating multi-platform publishing.
allowed-tools: Read, Bash
version: 1.0.0
tier: content
created: 2025-10-31
---

# Content: Markdown to Blog HTML Conversion

## What it does

Converts Markdown content to platform-specific HTML formats (WordPress, Naver Blog, Tistory) while handling images, code blocks, metadata, and platform-specific requirements to enable automated multi-platform publishing.

## When to use

- Converting Markdown blog posts to WordPress HTML
- Formatting content for Naver Blog or Tistory
- Automating multi-platform content syndication
- Building Markdown-based publishing workflows
- Migrating content between platforms

## Key Patterns

### 1. Basic Markdown to HTML Conversion

**Pattern**: Use markdown parsers for standard transformation

\`\`\`typescript
import { marked } from 'marked';

// Basic conversion
const markdownContent = \`
# Hello World

This is a **bold** text and *italic* text.

## Code Example

\\\`\\\`\\\`typescript
console.log("Hello");
\\\`\\\`\\\`
\`;

const htmlContent = marked.parse(markdownContent);
console.log(htmlContent);
// Output: <h1>Hello World</h1><p>This is a <strong>bold</strong>...
\`\`\`

### 2. WordPress-Specific Conversion

**Pattern**: Add WordPress blocks and metadata

\`\`\`typescript
const convertToWordPress = (markdown: string): string => {
  let html = marked.parse(markdown);
  
  // Add WordPress block comments
  html = html.replace(
    /<h2>(.*?)<\/h2>/g,
    '<!-- wp:heading --><h2>$1</h2><!-- /wp:heading -->'
  );
  
  html = html.replace(
    /<p>(.*?)<\/p>/g,
    '<!-- wp:paragraph --><p>$1</p><!-- /wp:paragraph -->'
  );
  
  // Handle code blocks with syntax highlighting
  html = html.replace(
    /<pre><code class="language-(\w+)">(.*?)<\/code><\/pre>/gs,
    '<!-- wp:code {"language":"$1"} --><pre class="wp-block-code"><code>$2</code></pre><!-- /wp:code -->'
  );
  
  return html;
};
\`\`\`

### 3. Naver Blog HTML Format

**Pattern**: Korean-optimized HTML structure

\`\`\`typescript
const convertToNaverBlog = (markdown: string): string => {
  // Naver Blog prefers simpler HTML without block comments
  let html = marked.parse(markdown, {
    headerIds: false, // Naver ignores IDs
    mangle: false
  });
  
  // Convert headings to Naver style
  html = html.replace(
    /<h1>(.*?)<\/h1>/g,
    '<p><span style="font-size: 24px; font-weight: bold;">$1</span></p>'
  );
  
  html = html.replace(
    /<h2>(.*?)<\/h2>/g,
    '<p><span style="font-size: 20px; font-weight: bold;">$1</span></p>'
  );
  
  // Ensure proper Korean UTF-8 encoding
  html = Buffer.from(html, 'utf8').toString('utf8');
  
  return html;
};
\`\`\`

### 4. Image Path Handling

**Pattern**: Transform local paths to CDN URLs

\`\`\`typescript
const processImages = (
  html: string, 
  imageMap: Map<string, string>
): string => {
  // Replace local image paths with uploaded CDN URLs
  html = html.replace(
    /src="\.\/images\/(.+?)"/g,
    (match, filename) => {
      const cdnUrl = imageMap.get(filename);
      return cdnUrl ? \`src="\${cdnUrl}"\` : match;
    }
  );
  
  // Add responsive image attributes
  html = html.replace(
    /<img src="([^"]+)"/g,
    '<img loading="lazy" src="$1" style="max-width: 100%; height: auto;"'
  );
  
  return html;
};

// Usage
const imageMap = new Map([
  ['screenshot.png', 'https://cdn.example.com/images/screenshot.png'],
  ['diagram.jpg', 'https://cdn.example.com/images/diagram.jpg']
]);

const htmlWithCdn = processImages(htmlContent, imageMap);
\`\`\`

### 5. Front Matter Extraction

**Pattern**: Parse YAML front matter for metadata

\`\`\`typescript
import matter from 'gray-matter';

const parseMarkdownWithMeta = (markdownFile: string) => {
  const { data, content } = matter(markdownFile);
  
  return {
    meta: {
      title: data.title,
      description: data.description,
      tags: data.tags || [],
      categories: data.categories || [],
      publishedAt: data.date,
      author: data.author
    },
    content: marked.parse(content)
  };
};

// Example Markdown with front matter
const markdownWithMeta = \`---
title: "Next.js 15 Guide"
description: "Comprehensive Next.js tutorial"
tags: ["nextjs", "react", "tutorial"]
categories: ["Web Development"]
date: 2025-10-31
author: "John Doe"
---

# Introduction

Welcome to Next.js 15...
\`;

const { meta, content } = parseMarkdownWithMeta(markdownWithMeta);
\`\`\`

### 6. Code Block Syntax Highlighting

**Pattern**: Add language classes for Prism/Highlight.js

\`\`\`typescript
const enhanceCodeBlocks = (html: string): string => {
  // Add language class for syntax highlighting
  html = html.replace(
    /<pre><code class="language-(\w+)">(.*?)<\/code><\/pre>/gs,
    (match, lang, code) => {
      // Escape HTML entities in code
      const escapedCode = code
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;');
      
      return \`<pre class="language-\${lang}"><code class="language-\${lang}">\${escapedCode}</code></pre>\`;
    }
  );
  
  return html;
};
\`\`\`

## Best Practices

### Markdown Writing
- Use standard Markdown syntax for portability
- Store images in dedicated directory (\`/images/\`)
- Include alt text for all images
- Use front matter for metadata (YAML format)
- Test Markdown rendering before conversion

### Conversion Process
- Validate Markdown syntax before conversion
- Handle platform-specific HTML requirements
- Upload images first, then update paths
- Add syntax highlighting classes for code blocks
- Test rendered HTML on target platform

### Image Handling
- Optimize images before uploading (WebP format, compress)
- Use descriptive file names (next-js-architecture.png)
- Upload to CDN or platform storage first
- Replace local paths with absolute URLs
- Add responsive image attributes

### Multi-Platform Publishing
- Create platform-specific conversion functions
- Store converted HTML separately (don't overwrite)
- Maintain mapping between Markdown source and published URLs
- Version control Markdown sources
- Automate conversion in CI/CD pipeline

## Resources

- Marked.js Documentation: https://marked.js.org/
- Gray-matter (Front Matter): https://github.com/jonschlinkert/gray-matter
- Unified/Remark: https://unifiedjs.com/
- Markdown Guide: https://www.markdownguide.org/

## Examples

**Example 1: Complete Conversion Workflow**

\`\`\`typescript
import { marked } from 'marked';
import matter from 'gray-matter';
import { uploadImage } from './storage';

const convertMarkdownToWordPress = async (
  markdownFile: string
): Promise<WordPressPost> => {
  // Step 1: Parse front matter
  const { data: meta, content: markdown } = matter(markdownFile);
  
  // Step 2: Extract image paths
  const imageMatches = markdown.matchAll(/!\[([^\]]+)\]\(([^)]+)\)/g);
  const imageMap = new Map();
  
  // Step 3: Upload images
  for (const [, alt, localPath] of imageMatches) {
    const cdnUrl = await uploadImage(localPath);
    imageMap.set(localPath, cdnUrl);
  }
  
  // Step 4: Convert Markdown to HTML
  let html = marked.parse(markdown);
  
  // Step 5: Replace image paths
  for (const [localPath, cdnUrl] of imageMap) {
    html = html.replace(
      new RegExp(\`src="\${localPath.replace(/[.*+?^${}()|[\]\\\\]/g, '\\\\$&')}"\`, 'g'),
      \`src="\${cdnUrl}"\`
    );
  }
  
  // Step 6: Add WordPress blocks
  html = convertToWordPress(html);
  
  return {
    title: meta.title,
    content: html,
    status: 'draft',
    categories: meta.categories,
    tags: meta.tags
  };
};
\`\`\`

**Example 2: Automation Script**

\`\`\`bash
#!/bin/bash
# convert-and-publish.sh

MARKDOWN_FILE="posts/nextjs-15-guide.md"
OUTPUT_DIR="output"

# Convert to WordPress HTML
node scripts/convert-to-wordpress.js "\$MARKDOWN_FILE" > "\$OUTPUT_DIR/wordpress.html"

# Convert to Naver Blog HTML
node scripts/convert-to-naver.js "\$MARKDOWN_FILE" > "\$OUTPUT_DIR/naver.html"

# Publish to WordPress
curl -X POST "https://yourdomain.com/wp-json/wp/v2/posts" \\
  -H "Authorization: Bearer \$WP_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d @<(jq -n \\
    --arg title "\$(head -n 1 \$MARKDOWN_FILE)" \\
    --arg content "\$(cat \$OUTPUT_DIR/wordpress.html)" \\
    '{title: \$title, content: \$content, status: "publish"}')

echo "Published to WordPress!"
\`\`\`

**Example 3: Platform-Specific Renderer**

\`\`\`typescript
class MultiPlatformConverter {
  private markdown: string;
  
  constructor(markdown: string) {
    this.markdown = markdown;
  }
  
  toWordPress(): string {
    return convertToWordPress(this.markdown);
  }
  
  toNaverBlog(): string {
    return convertToNaverBlog(this.markdown);
  }
  
  toMedium(): string {
    // Medium supports basic HTML
    return marked.parse(this.markdown);
  }
  
  toDevTo(): string {
    // Dev.to prefers Markdown
    return this.markdown;
  }
}

// Usage
const converter = new MultiPlatformConverter(markdownContent);
const wpHtml = converter.toWordPress();
const naverHtml = converter.toNaverBlog();
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with WordPress, Naver Blog, Tistory conversions, image handling

## Works well with
- \`moai-saas-wordpress-publishing\` (Publish converted HTML)
- \`moai-saas-naver-blog-publishing\` (Naver Blog publishing)
- \`moai-content-seo-optimization\` (Add SEO metadata)
- \`moai-content-image-generation\` (Generate featured images)
