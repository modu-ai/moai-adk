---
name: moai-saas-naver-blog-publishing
description: Publishing posts to Naver Blog via OpenAPI. Authenticate with Client ID/Secret, create posts, and manage content. Use when publishing to Naver Blog, automating Korean blog workflows, or syncing multi-platform content.
allowed-tools: Read, Bash, WebFetch
version: 1.0.0
tier: saas
created: 2025-10-31
---

# SaaS: Naver Blog OpenAPI Publishing

## What it does

Automates blog post publishing to Naver Blog using the Naver OpenAPI with Client ID/Secret authentication, supporting post creation, category management, and content updates for Korean-language blogging platforms.

## When to use

- Publishing to Naver Blog programmatically
- Automating Korean blog content distribution
- Syncing content across multiple Korean platforms
- Building Naver Blog management tools
- Integrating Naver Blog with content workflows

## Key Patterns

### 1. OpenAPI Authentication

**Pattern**: Use Client ID and Client Secret in HTTP headers

\`\`\`bash
# Register at: https://developers.naver.com/apps/#/register
# Get Client ID and Client Secret

export NAVER_CLIENT_ID="your_client_id"
export NAVER_CLIENT_SECRET="your_client_secret"
\`\`\`

**Authentication Headers**:
\`\`\`typescript
const headers = {
  'X-Naver-Client-Id': process.env.NAVER_CLIENT_ID,
  'X-Naver-Client-Secret': process.env.NAVER_CLIENT_SECRET,
  'Content-Type': 'application/json'
};
\`\`\`

### 2. Search Blog Posts (Read Access)

**Pattern**: Query Naver Blog posts via search API

\`\`\`typescript
const searchBlogPosts = async (query: string) => {
  const response = await fetch(
    \`https://openapi.naver.com/v1/search/blog.json?query=\${encodeURIComponent(query)}&display=10&sort=date\`,
    { headers }
  );
  
  const data = await response.json();
  return data.items; // Array of blog posts
};
\`\`\`

### 3. Post Structure

**Pattern**: Naver Blog post format requirements

\`\`\`typescript
interface NaverBlogPost {
  title: string;
  contents: string; // HTML content
  category?: string;
  tags?: string[]; // ["태그1", "태그2"]
  isPublic?: boolean; // true: public, false: private
}

const createPost = async (post: NaverBlogPost) => {
  // Note: Direct posting API requires blog ownership verification
  // Most common approach: Use Naver Blog RSS or manual posting
  
  const postData = {
    title: post.title,
    contents: post.contents,
    category: post.category || "기본",
    tags: post.tags?.join(','),
    open: post.isPublic ? "Y" : "N"
  };
  
  // Implementation depends on Naver Blog API access level
  return postData;
};
\`\`\`

### 4. Image Upload

**Pattern**: Upload images before including in post

\`\`\`typescript
// Upload image to Naver storage
const uploadImage = async (imageBuffer: Buffer, filename: string) => {
  const formData = new FormData();
  formData.append('image', imageBuffer, filename);
  
  const response = await fetch('https://openapi.naver.com/blog/image_upload', {
    method: 'POST',
    headers: {
      'X-Naver-Client-Id': process.env.NAVER_CLIENT_ID,
      'X-Naver-Client-Secret': process.env.NAVER_CLIENT_SECRET
    },
    body: formData
  });
  
  const { image_url } = await response.json();
  return image_url;
};
\`\`\`

## Best Practices

- Store credentials securely (environment variables, secret managers)
- Handle Korean UTF-8 encoding properly
- Use proper HTML tags for content formatting
- Include relevant tags for Naver search optimization
- Respect API rate limits (implement exponential backoff)
- Validate HTML content before posting
- Use public/private flags appropriately
- Test with draft posts before publishing
- Handle errors gracefully with Korean messages
- Log all API responses for debugging

## Important Notes

**API Limitations (2025)**:
- Naver Blog OpenAPI primarily provides **search/read access**
- Direct post creation via API requires special partnership or blog ownership verification
- Most common approach: Generate content programmatically, then use Naver Blog UI or RSS for publishing
- Alternative: Use web automation tools for posting

**Recommended Workflow**:
1. Generate content programmatically
2. Format as Naver Blog-compatible HTML
3. Use Naver Blog mobile app API or web interface for final publishing
4. Verify post via search API

## Resources

- Naver Developers: https://developers.naver.com/
- Naver OpenAPI Guide: https://github.com/naver/naver-openapi-guide
- Naver Blog Search API: https://developers.naver.com/docs/search/blog/

## Examples

**Example 1: Search Naver Blog Posts**

\`\`\`bash
curl -X GET "https://openapi.naver.com/v1/search/blog.json?query=워드프레스&display=5" \
  -H "X-Naver-Client-Id: YOUR_CLIENT_ID" \
  -H "X-Naver-Client-Secret: YOUR_CLIENT_SECRET"
\`\`\`

**Example 2: Format Post for Naver Blog**

\`\`\`typescript
const formatForNaverBlog = (markdown: string): string => {
  // Convert Markdown to Naver Blog HTML
  let html = markdown
    .replace(/^# (.+)$/gm, '<h2>$1</h2>')
    .replace(/^## (.+)$/gm, '<h3>$1</h3>')
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.+?)\*/g, '<em>$1</em>')
    .replace(/\n\n/g, '</p><p>')
    .replace(/\n/g, '<br>');
  
  return \`<p>\${html}</p>\`;
};

const post = {
  title: "네이버 블로그 API 활용하기",
  contents: formatForNaverBlog(markdownContent),
  tags: ["API", "네이버", "블로그"],
  isPublic: true
};
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with OpenAPI authentication, search API, post formatting

## Works well with
- \`moai-content-markdown-to-blog\` (Convert Markdown to Naver HTML)
- \`moai-content-seo-optimization\` (Optimize for Naver search)
- \`moai-saas-tistory-publishing\` (Multi-platform Korean blog publishing)
