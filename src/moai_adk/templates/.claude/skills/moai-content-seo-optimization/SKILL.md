---
name: moai-content-seo-optimization
description: Optimizing blog posts for search engines with 2025 best practices. Meta tags, schema markup, heading structure, and keyword strategy. Use when optimizing SEO, improving search rankings, or implementing structured data.
allowed-tools: Read, Bash, WebFetch
version: 1.0.0
tier: content
created: 2025-10-31
---

# Content: SEO Optimization for Blog Posts

## What it does

Applies modern SEO optimization techniques (2025 best practices) to blog posts, including meta tags, schema markup, heading structure, keyword optimization, and AI-system compatibility to improve search rankings and AI-powered recommendations.

## When to use

- Optimizing blog posts for search engines
- Implementing structured data (schema.org)
- Improving click-through rates from search results
- Preparing content for AI system indexing
- Auditing existing content for SEO issues

## Key Patterns

### 1. Essential Meta Tags (2025)

**Pattern**: Title and meta description optimization

\`\`\`html
<!-- Title Tag (50-60 chars) -->
<title>Ultimate Guide to Next.js 15 - Performance & SSR</title>

<!-- Meta Description (150-160 chars desktop, 120 mobile) -->
<meta name="description" content="Learn Next.js 15 with practical examples: App Router, Server Components, streaming, and performance optimization for production apps.">

<!-- Open Graph (Social Media) -->
<meta property="og:title" content="Ultimate Guide to Next.js 15">
<meta property="og:description" content="Master Next.js 15...">
<meta property="og:image" content="https://example.com/og-image.jpg">
<meta property="og:type" content="article">

<!-- Twitter Card -->
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="Ultimate Guide to Next.js 15">
<meta name="twitter:description" content="Master Next.js 15...">
<meta name="twitter:image" content="https://example.com/twitter-image.jpg">

<!-- Canonical URL -->
<link rel="canonical" href="https://example.com/nextjs-15-guide">
\`\`\`

### 2. Schema Markup (Article + FAQ)

**Pattern**: Structured data for search engines and AI systems

\`\`\`html
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "BlogPosting",
  "headline": "Ultimate Guide to Next.js 15",
  "image": "https://example.com/featured-image.jpg",
  "datePublished": "2025-10-31T09:00:00Z",
  "dateModified": "2025-10-31T09:00:00Z",
  "author": {
    "@type": "Person",
    "name": "John Doe",
    "url": "https://example.com/author/john-doe"
  },
  "publisher": {
    "@type": "Organization",
    "name": "Example Blog",
    "logo": {
      "@type": "ImageObject",
      "url": "https://example.com/logo.png"
    }
  },
  "description": "Comprehensive guide to Next.js 15...",
  "mainEntityOfPage": {
    "@type": "WebPage",
    "@id": "https://example.com/nextjs-15-guide"
  }
}
</script>

<!-- FAQ Schema -->
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "FAQPage",
  "mainEntity": [{
    "@type": "Question",
    "name": "What is Next.js 15?",
    "acceptedAnswer": {
      "@type": "Answer",
      "text": "Next.js 15 is the latest version..."
    }
  }, {
    "@type": "Question",
    "name": "What are the key features?",
    "acceptedAnswer": {
      "@type": "Answer",
      "text": "Key features include App Router..."
    }
  }]
}
</script>
\`\`\`

### 3. Heading Structure (H1-H6 Hierarchy)

**Pattern**: Semantic heading hierarchy for readability and SEO

\`\`\`html
<!-- Single H1 per page -->
<h1>Ultimate Guide to Next.js 15</h1>

<!-- H2 for major sections -->
<h2>Getting Started with Next.js 15</h2>
  <h3>Installation and Setup</h3>
  <h3>Project Structure</h3>

<h2>Core Features</h2>
  <h3>App Router</h3>
    <h4>Layouts and Pages</h4>
    <h4>Server Components</h4>
  <h3>Data Fetching</h3>

<h2>Advanced Patterns</h2>
  <h3>Streaming and Suspense</h3>
  <h3>Middleware</h3>
\`\`\`

### 4. Keyword Strategy

**Pattern**: Primary and secondary keyword placement

\`\`\`markdown
Primary Keyword: "Next.js 15 guide"
Secondary Keywords: "App Router", "Server Components", "SSR performance"

Placement Strategy:
- Title Tag: Primary keyword at start
- H1: Primary keyword
- First 100 words: Primary keyword + 1 secondary
- H2/H3 headings: Mix of primary + secondary keywords
- Image alt text: Descriptive with keywords
- URL slug: /nextjs-15-guide (primary keyword)
- Internal links: Anchor text with keywords
\`\`\`

### 5. Image Optimization

**Pattern**: Alt text, file names, and modern formats

\`\`\`html
<!-- Descriptive file name -->
<img 
  src="/images/nextjs-15-app-router-diagram.webp"
  alt="Next.js 15 App Router architecture showing routing hierarchy"
  width="800"
  height="400"
  loading="lazy"
>

<!-- Responsive images -->
<picture>
  <source srcset="/images/hero.webp" type="image/webp">
  <source srcset="/images/hero.jpg" type="image/jpeg">
  <img src="/images/hero.jpg" alt="Next.js 15 features overview">
</picture>
\`\`\`

### 6. Internal Linking

**Pattern**: Strategic internal links with descriptive anchors

\`\`\`html
<!-- Good internal link -->
<a href="/nextjs-14-migration-guide">
  Learn how to migrate from Next.js 14 to 15
</a>

<!-- Related content links -->
<section class="related-posts">
  <h2>Related Articles</h2>
  <ul>
    <li><a href="/react-server-components">Understanding React Server Components</a></li>
    <li><a href="/nextjs-performance">Next.js Performance Optimization</a></li>
  </ul>
</section>
\`\`\`

## Best Practices (2025)

### Meta Tags
- Title tag: 50-60 characters (desktop), include primary keyword
- Meta description: 150-160 characters (desktop), 120 (mobile)
- Use unique titles and descriptions per page
- Include target keywords naturally (avoid keyword stuffing)

### Schema Markup
- Implement Article schema for blog posts (increases AI trust)
- Add FAQ schema for Q&A sections (rich snippet opportunity)
- Use BreadcrumbList for navigation hierarchy
- Validate schema with Google Rich Results Test

### Content Structure
- Single H1 per page (primary keyword)
- H2-H6 for logical hierarchy
- Use short paragraphs (2-3 sentences)
- Include table of contents for long posts
- Add jump links for easy navigation

### Keyword Optimization
- Primary keyword density: 1-2%
- Use LSI keywords (related terms)
- Include keywords in first 100 words
- Natural language (write for humans, not algorithms)

### AI System Optimization (2025 Focus)
- Create llms.txt file for AI indexing
- Use clear, concise language
- Provide factual, well-sourced information
- Include code examples and practical guidance
- Avoid clickbait or misleading content

### Technical SEO
- Canonical URLs to avoid duplicate content
- XML sitemap for search engine crawling
- Robots.txt for crawl control
- HTTPS everywhere (security requirement)
- Mobile-first design (responsive)

## Resources

- Google Search Central: https://developers.google.com/search
- Schema.org Documentation: https://schema.org/
- Rank Math Schema Markup Guide: https://rankmath.com/blog/schema-markup/
- Google Rich Results Test: https://search.google.com/test/rich-results

## Examples

**Example 1: Complete SEO-Optimized HTML Head**

\`\`\`html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  
  <title>Next.js 15 Guide: App Router, SSR & Performance | 2025</title>
  <meta name="description" content="Master Next.js 15 with this comprehensive guide covering App Router, Server Components, data fetching, and performance optimization.">
  
  <link rel="canonical" href="https://example.com/nextjs-15-guide">
  
  <meta property="og:title" content="Next.js 15 Guide: App Router & SSR">
  <meta property="og:description" content="Comprehensive Next.js 15 tutorial...">
  <meta property="og:image" content="https://example.com/og-nextjs-15.jpg">
  <meta property="og:url" content="https://example.com/nextjs-15-guide">
  <meta property="og:type" content="article">
  
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="Next.js 15 Guide">
  <meta name="twitter:description" content="Master Next.js 15...">
  <meta name="twitter:image" content="https://example.com/twitter-card.jpg">
  
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    "headline": "Next.js 15 Guide: App Router, SSR & Performance",
    "datePublished": "2025-10-31T09:00:00Z",
    "author": {"@type": "Person", "name": "John Doe"}
  }
  </script>
</head>
<body>
  <!-- Content here -->
</body>
</html>
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with 2025 SEO best practices, AI system optimization, schema markup

## Works well with
- \`moai-content-markdown-to-blog\` (Generate SEO-optimized HTML)
- \`moai-content-llms-txt-management\` (AI system indexing)
- \`moai-saas-wordpress-publishing\` (Apply SEO to WordPress posts)
