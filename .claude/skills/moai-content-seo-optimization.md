# moai-content-seo-optimization

Optimizing blog posts for search engines with 2025 best practices, meta tags, schema markup, and keyword strategy.

## Quick Start

SEO optimization ensures your content is discoverable by search engines and ranks well for relevant keywords. Use this skill when writing blog posts, optimizing meta descriptions, implementing structured data, or improving search rankings.

## Core Patterns

### Pattern 1: Technical SEO & Meta Tags

**Pattern**: Implement proper meta tags, structured data, and SEO metadata for blog posts.

```html
<!-- HTML Head SEO Tags -->
<!DOCTYPE html>
<html lang="en">
<head>
  <!-- Primary Meta Tags -->
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>How to Build Scalable Node.js Applications | My Blog</title>
  <meta name="description" content="Learn best practices for building scalable Node.js applications with async/await, clustering, and database optimization.">
  <meta name="keywords" content="Node.js, scalability, performance, async/await, clustering">
  <meta name="author" content="Your Name">

  <!-- Open Graph / Facebook -->
  <meta property="og:type" content="article">
  <meta property="og:url" content="https://example.com/blog/nodejs-scalability">
  <meta property="og:title" content="How to Build Scalable Node.js Applications">
  <meta property="og:description" content="Learn best practices for building scalable Node.js applications">
  <meta property="og:image" content="https://example.com/images/nodejs-article.jpg">
  <meta property="og:site_name" content="My Blog">

  <!-- Twitter -->
  <meta property="twitter:card" content="summary_large_image">
  <meta property="twitter:url" content="https://example.com/blog/nodejs-scalability">
  <meta property="twitter:title" content="How to Build Scalable Node.js Applications">
  <meta property="twitter:description" content="Learn best practices for building scalable Node.js applications">
  <meta property="twitter:image" content="https://example.com/images/nodejs-article.jpg">
  <meta name="twitter:creator" content="@yourhandle">

  <!-- Canonical URL -->
  <link rel="canonical" href="https://example.com/blog/nodejs-scalability">

  <!-- Structured Data (JSON-LD) -->
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    "headline": "How to Build Scalable Node.js Applications",
    "description": "Learn best practices for building scalable Node.js applications with async/await, clustering, and database optimization.",
    "image": "https://example.com/images/nodejs-article.jpg",
    "datePublished": "2024-10-31T12:00:00+00:00",
    "dateModified": "2024-10-31T12:00:00+00:00",
    "author": {
      "@type": "Person",
      "name": "Your Name",
      "url": "https://example.com/about"
    },
    "publisher": {
      "@type": "Organization",
      "name": "My Blog",
      "logo": {
        "@type": "ImageObject",
        "url": "https://example.com/logo.png"
      }
    }
  }
  </script>
</head>
<body>
  <!-- Content -->
</body>
</html>
```

**When to use**:
- Writing new blog posts
- Optimizing existing articles
- Ensuring proper schema markup
- Improving social media sharing

**Key benefits**:
- Improved search engine visibility
- Better click-through rates from search results
- Rich snippets in search results
- Proper attribution and content sharing

### Pattern 2: Keyword Research & Content Strategy

**Pattern**: Identify target keywords and structure content for SEO.

```markdown
# Keyword Research & Content Strategy

## Target Keywords
- Primary: "Node.js scalability" (monthly searches: 2,100, competition: medium)
- Secondary: "scalable Node.js applications" (890 searches)
- Long-tail: "how to build scalable Node.js apps" (340 searches)
- LSI Keywords: clustering, horizontal scaling, load balancing

## Content Structure

### H1: How to Build Scalable Node.js Applications (Primary keyword)
- Use only once at top
- 50-60 characters for title tag

### H2 Sections (Include secondary keywords)
- Understanding Node.js Single-Threaded Architecture
- Horizontal vs Vertical Scaling
- Using Node.js Cluster Module
- Load Balancing Strategies
- Database Connection Pooling
- Caching Strategies (Redis)

### H3 Subsections (Related long-tail keywords)
- Setting up a Node.js cluster for multi-core systems
- Implementing round-robin load balancing
- Configuring connection pools for production

## Content Guidelines
- Target word count: 2,500-3,500 words
- Include 3-5 images with alt text
- Add code examples with syntax highlighting
- Include internal links to related articles
- Add external links to authoritative sources

## Meta Tags
- Title: "How to Build Scalable Node.js Applications | 2024 Guide"
- Description: "Learn best practices for building scalable Node.js applications with clustering, load balancing, and database optimization."
- Image: High-quality (1200x630px for social sharing)
```

**When to use**:
- Planning blog content
- Choosing article topics
- Structuring content for search
- Targeting competitive keywords

**Key benefits**:
- Data-driven content selection
- Better ranking for target keywords
- Higher organic traffic
- Content that matches user intent

### Pattern 3: On-Page Optimization Checklist

**Pattern**: Verify all on-page SEO elements are properly implemented.

```markdown
# SEO Optimization Checklist

## Title & Meta
- [ ] Title tag: 50-60 characters, includes primary keyword
- [ ] Meta description: 150-160 characters, compelling call-to-action
- [ ] H1 tag: Single, includes primary keyword
- [ ] URL slug: Descriptive, hyphenated, lowercase

## Content Structure
- [ ] Multiple H2/H3 headers with relevant keywords
- [ ] Intro paragraph summarizes main topic
- [ ] Content is well-organized and scannable
- [ ] Lists (ul/ol) used for readability
- [ ] Consistent formatting throughout

## Images & Media
- [ ] All images have descriptive alt text
- [ ] Image file names are descriptive (not "image1.jpg")
- [ ] Images are optimized for web (<200KB each)
- [ ] Video embeds include captions/transcripts

## Internal Linking
- [ ] 3-5 internal links to related articles
- [ ] Link anchor text is descriptive
- [ ] Links point to relevant, high-authority pages
- [ ] No broken links

## External Linking
- [ ] 2-3 links to authoritative external sources
- [ ] Links open in new tab (target="_blank")
- [ ] Linked domains are relevant and trustworthy

## Technical SEO
- [ ] Mobile-responsive design verified
- [ ] Page load time < 3 seconds
- [ ] Structured data (JSON-LD) implemented
- [ ] Canonical URL specified
- [ ] Robots meta tags correct

## Social & Sharing
- [ ] Open Graph tags present
- [ ] Twitter Card tags present
- [ ] Social sharing buttons available
- [ ] Preview looks good on Facebook/Twitter

## Performance
- [ ] Core Web Vitals optimized
- [ ] LCP < 2.5s, FID < 100ms, CLS < 0.1
- [ ] Mobile usability verified
- [ ] No crawl errors in Search Console
```

**When to use**:
- Before publishing blog posts
- Auditing existing content
- Improving rankings for old articles
- Ensuring consistent quality

**Key benefits**:
- Better search rankings
- Improved user experience
- Higher conversion rates
- Reduced bounce rates

## Progressive Disclosure

### Level 1: Basic SEO
- Meta tags and descriptions
- Proper heading structure
- Image alt text
- Mobile responsiveness

### Level 2: Advanced SEO
- Keyword research and targeting
- Internal linking strategy
- Schema markup implementation
- Core Web Vitals optimization

### Level 3: Expert SEO
- Advanced technical SEO
- Content cluster strategy
- Link building strategies
- International SEO

## Works Well With

- **Next.js**: SEO-friendly framework with metadata support
- **Markdown**: Format for writing blog content
- **Search Console**: Google's tool for monitoring SEO
- **Analytics**: Track organic traffic and keywords
- **Lighthouse**: Audit Core Web Vitals
- **Yoast**: SEO analysis tool

## References

- **Google SEO Starter Guide**: https://developers.google.com/search/docs/beginner/seo-starter-guide
- **Search Console**: https://search.google.com/search-console/about
- **Schema.org**: https://schema.org/
- **Core Web Vitals**: https://web.dev/vitals/
- **Keyword Tool**: https://ahrefs.com/keyword-generator
