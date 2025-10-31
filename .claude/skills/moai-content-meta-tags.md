# moai-content-meta-tags

Crafting effective meta tags, Open Graph tags, and Twitter cards for content sharing.

## Quick Start

Meta tags improve how content appears in search results and social media. Use this skill when publishing blog posts and content.

## Core Patterns

### Pattern 1: Essential Meta Tags

```html
<!-- Meta Tags for SEO and Social Sharing -->

<!-- Basic Meta Tags -->
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="description" content="Learn how to build scalable Node.js applications with async/await patterns, clustering, and database optimization.">
<meta name="keywords" content="Node.js, scalability, performance, backend, async">
<meta name="author" content="Author Name">

<!-- Open Graph (Facebook, LinkedIn, etc.) -->
<meta property="og:type" content="article">
<meta property="og:url" content="https://example.com/blog/nodejs-scalability">
<meta property="og:title" content="Building Scalable Node.js Applications">
<meta property="og:description" content="Learn best practices for building scalable Node.js applications with async/await, clustering, and database optimization.">
<meta property="og:image" content="https://example.com/images/nodejs-article.jpg">
<meta property="og:site_name" content="My Blog">
<meta property="article:published_time" content="2024-10-31">
<meta property="article:author" content="Author Name">

<!-- Twitter Card -->
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:url" content="https://example.com/blog/nodejs-scalability">
<meta name="twitter:title" content="Building Scalable Node.js Applications">
<meta name="twitter:description" content="Learn best practices for building scalable Node.js applications.">
<meta name="twitter:image" content="https://example.com/images/nodejs-article.jpg">
<meta name="twitter:creator" content="@yourhandle">

<!-- Canonical URL -->
<link rel="canonical" href="https://example.com/blog/nodejs-scalability">
```

### Pattern 2: Schema Markup

```json
{
  "@context": "https://schema.org",
  "@type": "BlogPosting",
  "headline": "Building Scalable Node.js Applications",
  "description": "Learn best practices for building scalable Node.js applications.",
  "image": "https://example.com/images/nodejs-article.jpg",
  "datePublished": "2024-10-31",
  "dateModified": "2024-10-31",
  "author": {
    "@type": "Person",
    "name": "Author Name",
    "url": "https://example.com/about"
  },
  "publisher": {
    "@type": "Organization",
    "name": "My Blog",
    "logo": {
      "@type": "ImageObject",
      "url": "https://example.com/logo.png"
    }
  },
  "mainEntityOfPage": {
    "@type": "WebPage",
    "@id": "https://example.com/blog/nodejs-scalability"
  }
}
```

### Pattern 3: Meta Tag Best Practices

**Meta Description**:
- Length: 150-160 characters
- Include primary keyword
- Call to action: "Learn", "Discover", "Master"
- Unique for each page

**OG Image**:
- Dimensions: 1200x630px (16:9 ratio)
- File size: <200KB
- High contrast and readable text
- Brand consistency

**Title Tags**:
- Length: 50-60 characters
- Primary keyword first
- Brand name at end
- Example: "Node.js Scalability Guide | My Blog"

**When to use**:
- Publishing blog posts
- Optimizing social sharing
- Improving search visibility
- A/B testing card designs

**Key benefits**:
- Better click-through rates
- Improved social sharing
- Better search rankings
- Professional appearance

## Progressive Disclosure

### Level 1: Basic Tags
- Meta description
- OG title and description
- Twitter card

### Level 2: Advanced Tags
- Schema markup
- Multiple OG images for platforms
- Canonical URLs
- Author and publisher tags

### Level 3: Expert Optimization
- A/B testing variants
- Platform-specific optimization
- Dynamic meta generation
- Performance optimization

## Works Well With

- **SEO Tools**: Check meta tag coverage
- **Social Media**: Preview before sharing
- **Analytics**: Track traffic by meta content
- **CMS**: Auto-generate meta tags
- **Testing**: Validate schema markup

## References

- **Meta Tags Guide**: https://ogp.me/
- **Twitter Cards**: https://developer.twitter.com/en/docs/twitter-for-websites/cards
- **Schema.org**: https://schema.org/
- **Facebook Debugger**: https://developers.facebook.com/tools/debug/
