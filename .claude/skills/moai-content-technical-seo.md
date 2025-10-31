# moai-content-technical-seo

Technical SEO optimization for blog platforms, site structure, and crawlability.

## Quick Start

Technical SEO ensures search engines can effectively crawl, index, and rank your content. Use this skill when optimizing site performance and structure for SEO.

## Core Patterns

### Pattern 1: Site Structure & Crawlability

**Pattern**: Organize content for optimal crawling.

```
Ideal Site Structure:

/
├── /blog (main blog page)
├── /blog/category1/
│   ├── post-1.html (URL structure)
│   └── post-2.html (short slugs)
├── /blog/category2/
│   ├── post-3.html
│   └── post-4.html
├── /about
├── /contact
└── /rss.xml (XML sitemap)

URL Best Practices:
✅ Good:
- /blog/nodejs-scalability (descriptive, short)
- /guides/getting-started (hierarchical)
- /tutorials/build-api (lowercase, hyphenated)

❌ Bad:
- /blog/2024-10-31-article (date-based)
- /blog/article-about-nodejs (too long)
- /blog/Article-About-NodeJS (inconsistent casing)

Sitemap (sitemap.xml):
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/blog/nodejs-scalability</loc>
    <lastmod>2024-10-31</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>

Robots.txt:
User-agent: *
Allow: /
Disallow: /admin
Disallow: /private

Sitemap: https://example.com/sitemap.xml
```

### Pattern 2: Core Web Vitals Optimization

**Pattern**: Optimize for Core Web Vitals metrics.

```
Largest Contentful Paint (LCP) - Target: < 2.5s
- Optimize images (use WebP, modern formats)
- Lazy load below-the-fold content
- Minimize CSS/JavaScript
- Use CDN for static assets

First Input Delay (FID) - Target: < 100ms
- Minimize JavaScript execution
- Code split large bundles
- Use Web Workers for heavy processing
- Defer non-critical JavaScript

Cumulative Layout Shift (CLS) - Target: < 0.1
- Reserve space for images/videos
- Avoid inserting content dynamically
- Use transforms instead of dimensions
- Font-display: swap for web fonts

Measurement:
- Use PageSpeed Insights
- Monitor with Google Search Console
- Test with Lighthouse
- Track with Web Vitals API

```typescript
// Monitor Core Web Vitals
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

getCLS(console.log); // { name: 'CLS', value: 0.08, ... }
getFID(console.log); // { name: 'FID', value: 75, ... }
getLCP(console.log); // { name: 'LCP', value: 2100, ... }
```

### Pattern 3: Rich Results & Featured Snippets

**Pattern**: Optimize for featured snippets and rich results.

```markdown
# Question: How do I optimize for featured snippets?

Featured snippets appear as special search results.
There are 4 types: paragraph, list, table, video.

## Paragraph Snippet
Keep key answer in 40-60 words at top of section.

"Featured snippets are selected search results that
appear at the top of Google results in a special
format. They include a preview of the answer from
the website, the page title, and URL."

## List Snippet
Use clear, numbered or bulleted lists.

How to optimize for featured snippets:
1. Target long-tail keywords with "how to"
2. Structure content with clear headers
3. Keep answers concise (40-60 words)
4. Use lists and tables
5. Include schema markup

## Table Snippet
Use proper HTML tables with headers.

| Method | Difficulty | Time |
|--------|-----------|------|
| Method A | Easy | 5 min |
| Method B | Medium | 15 min |
| Method C | Hard | 30 min |

## Schema Markup for Rich Results
\`\`\`json
{
  "@context": "https://schema.org",
  "@type": "HowTo",
  "name": "How to Optimize for Featured Snippets",
  "step": [
    {
      "@type": "HowToStep",
      "name": "Research keywords",
      "text": "Find questions your audience asks"
    }
  ]
}
\`\`\`
```

**When to use**:
- Publishing technical content
- Optimizing site structure
- Improving search rankings
- Monitoring performance

**Key benefits**:
- Better search visibility
- Faster page loads
- Rich snippets in results
- Higher click-through rates

## Progressive Disclosure

### Level 1: Basic Technical SEO
- XML sitemap
- Robots.txt
- Mobile responsiveness
- Page speed basics

### Level 2: Advanced Technical SEO
- Core Web Vitals optimization
- Schema markup
- Structured data
- Canonical URLs

### Level 3: Expert Technical SEO
- Advanced crawl analysis
- Log file analysis
- Redirect chains
- JavaScript SEO

## Works Well With

- **Google Search Console**: Indexing and performance
- **Lighthouse**: Performance auditing
- **Screaming Frog**: Site crawling
- **Semrush**: Technical SEO audit
- **Schema.org**: Structured data

## References

- **Google SEO Starter Guide**: https://developers.google.com/search/docs/beginner/seo-starter-guide
- **Web Vitals**: https://web.dev/vitals/
- **Schema.org**: https://schema.org/
- **Search Console Help**: https://support.google.com/webmasters/
