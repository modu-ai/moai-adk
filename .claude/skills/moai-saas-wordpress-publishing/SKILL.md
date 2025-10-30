---
name: moai-saas-wordpress-publishing
description: Publishing blog posts via WordPress REST API. Authenticate, create posts, manage media, and update content. Use when publishing to WordPress, automating blog workflows, or syncing content management systems.
allowed-tools: Read, Bash, WebFetch
version: 1.0.0
tier: saas
created: 2025-10-31
---

# SaaS: WordPress REST API Publishing

## What it does

Automates blog post publishing to WordPress sites using the REST API v2+ with Application Passwords authentication (WordPress 5.6+), supporting post creation, media uploads, category management, and content updates.

## When to use

- Publishing blog posts programmatically
- Automating content syndication to WordPress
- Bulk importing posts from external sources
- Updating multiple posts at once
- Managing WordPress content via CI/CD pipelines

## Key Patterns

### 1. Application Password Authentication (2025 Standard)

**Pattern**: Use Application Passwords for secure API access

\`\`\`bash
# Generate Application Password
# WordPress Admin → Users → Profile → Application Passwords
# Name: "API Access" → Add New

# Store credentials securely
export WP_USERNAME="admin"
export WP_APP_PASSWORD="xxxx xxxx xxxx xxxx xxxx xxxx"
export WP_SITE="https://yourdomain.com"
\`\`\`

### 2. Create Post with Categories and Tags

**Pattern**: POST request with authentication headers

\`\`\`typescript
const createPost = async (title: string, content: string) => {
  const auth = Buffer.from(
    \`\${WP_USERNAME}:\${WP_APP_PASSWORD}\`
  ).toString('base64');

  const response = await fetch(\`\${WP_SITE}/wp-json/wp/v2/posts\`, {
    method: 'POST',
    headers: {
      'Authorization': \`Basic \${auth}\`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      title,
      content,
      status: 'publish', // or 'draft'
      categories: [1, 5], // category IDs
      tags: [10, 20], // tag IDs
      meta: {
        seo_title: title,
        seo_description: 'Post description'
      }
    })
  });

  return await response.json();
};
\`\`\`

### 3. Upload Featured Image

**Pattern**: Multi-step upload and attachment

\`\`\`typescript
// Step 1: Upload image
const formData = new FormData();
formData.append('file', imageBuffer, 'featured.jpg');

const imageResponse = await fetch(\`\${WP_SITE}/wp-json/wp/v2/media\`, {
  method: 'POST',
  headers: {
    'Authorization': \`Basic \${auth}\`
  },
  body: formData
});

const { id: mediaId } = await imageResponse.json();

// Step 2: Attach to post
await fetch(\`\${WP_SITE}/wp-json/wp/v2/posts/\${postId}\`, {
  method: 'POST',
  headers: {
    'Authorization': \`Basic \${auth}\`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    featured_media: mediaId
  })
});
\`\`\`

### 4. Batch Update Posts

**Pattern**: Iterate and update multiple posts

\`\`\`typescript
const postIds = [123, 456, 789];

for (const postId of postIds) {
  await fetch(\`\${WP_SITE}/wp-json/wp/v2/posts/\${postId}\`, {
    method: 'POST',
    headers: {
      'Authorization': \`Basic \${auth}\`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      content: updatedContent,
      modified: new Date().toISOString()
    })
  });
}
\`\`\`

## Best Practices

- Always use HTTPS for API requests (required for Application Passwords)
- Store credentials in environment variables, never in code
- Handle rate limiting with exponential backoff
- Validate HTML content before posting
- Use post status 'draft' for review before publishing
- Set proper categories and tags for SEO
- Include custom fields for Yoast SEO or similar plugins
- Implement error handling for network failures
- Log all API responses for debugging

## Resources

- WordPress REST API Handbook: https://developer.wordpress.org/rest-api/
- Application Passwords Guide: https://make.wordpress.org/core/2020/11/05/application-passwords-integration-guide/
- REST API Authentication: https://developer.wordpress.org/rest-api/using-the-rest-api/authentication/

## Examples

**Example 1: Simple Post Creation**

\`\`\`bash
curl -X POST "https://yourdomain.com/wp-json/wp/v2/posts" \
  -u "username:xxxx xxxx xxxx xxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My New Post",
    "content": "<p>Post content here</p>",
    "status": "publish"
  }'
\`\`\`

**Example 2: Publish with SEO Metadata**

\`\`\`typescript
await createPost({
  title: 'Ultimate Guide to WordPress REST API',
  content: '<h2>Introduction</h2><p>...</p>',
  status: 'publish',
  categories: [2], // "Tutorials" category
  tags: [5, 10], // "API", "WordPress"
  meta: {
    _yoast_wpseo_title: 'Ultimate Guide to WordPress REST API - 2025',
    _yoast_wpseo_metadesc: 'Learn WordPress REST API...',
    _yoast_wpseo_focuskw: 'WordPress REST API'
  }
});
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with Application Passwords support, media upload, batch updates

## Works well with
- \`moai-content-markdown-to-blog\` (Convert Markdown to WordPress HTML)
- \`moai-content-seo-optimization\` (Optimize posts for SEO)
- \`moai-content-image-generation\` (Generate featured images)
