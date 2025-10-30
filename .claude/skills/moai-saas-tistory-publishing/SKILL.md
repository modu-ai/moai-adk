---
name: moai-saas-tistory-publishing
description: Publishing posts to Tistory Blog via OAuth API. Note - Tistory API service discontinued as of February 2024. Use when working with legacy Tistory integrations or exploring alternative publishing methods.
allowed-tools: Read, Bash, WebFetch
version: 1.0.0
tier: saas
created: 2025-10-31
---

# SaaS: Tistory Blog API Publishing (Legacy)

## What it does

Documents the Tistory Blog API OAuth authentication and posting patterns for legacy integrations. **Important**: Tistory Open API service was discontinued as of February 2024.

## When to use

- Maintaining legacy Tistory integrations (pre-2024)
- Understanding Tistory API architecture for migration
- Exploring alternative Tistory publishing methods
- Documenting historical API patterns

## Key Patterns (Historical Reference)

### 1. OAuth 2.0 Authentication Flow

**Pattern**: Three-step OAuth flow (No longer available)

\`\`\`typescript
// Step 1: Authorization URL
const authUrl = \`https://www.tistory.com/oauth/authorize?\` +
  \`client_id=\${CLIENT_ID}&\` +
  \`redirect_uri=\${REDIRECT_URI}&\` +
  \`response_type=code&\` +
  \`state=\${STATE}\`;

// Step 2: Exchange code for access token
const tokenResponse = await fetch('https://www.tistory.com/oauth/access_token', {
  method: 'POST',
  body: new URLSearchParams({
    client_id: CLIENT_ID,
    client_secret: CLIENT_SECRET,
    redirect_uri: REDIRECT_URI,
    code: authorizationCode,
    grant_type: 'authorization_code'
  })
});

const { access_token } = await tokenResponse.json();
\`\`\`

### 2. Post Creation (Historical)

**Pattern**: POST request to write endpoint

\`\`\`typescript
const createPost = async (accessToken: string, blogName: string) => {
  const response = await fetch('https://www.tistory.com/apis/post/write', {
    method: 'POST',
    body: new URLSearchParams({
      access_token: accessToken,
      output: 'json',
      blogName: blogName,
      title: 'Post Title',
      content: '<p>Post content in HTML</p>',
      visibility: '2', // 0: private, 1: protected, 2: public
      category: '0', // category ID
      tag: 'tag1,tag2,tag3',
      acceptComment: '1', // 0: disable, 1: enable
      password: '' // for protected posts
    })
  });
  
  return await response.json();
};
\`\`\`

### 3. Post Structure (Historical)

**Pattern**: Required and optional fields

\`\`\`typescript
interface TistoryPost {
  access_token: string;
  blogName: string;
  title: string;
  content: string; // HTML format
  visibility: '0' | '1' | '2'; // private | protected | public
  category?: string; // category ID
  tag?: string; // comma-separated
  acceptComment?: '0' | '1';
  password?: string; // for protected posts
}
\`\`\`

## Alternative Publishing Methods (2025)

Since the Tistory API was discontinued, consider these alternatives:

### 1. Manual Publishing
- Use Tistory web interface for posting
- Prepare content programmatically, publish manually

### 2. RSS/Email Integration
- Some blogs support email-to-post functionality
- Generate formatted emails programmatically

### 3. Web Automation
- Use browser automation (Playwright, Puppeteer)
- Automate form filling in Tistory editor

### 4. Platform Migration
- Consider migrating to platforms with active APIs:
  - WordPress (REST API actively maintained)
  - Medium (API available)
  - Dev.to (API available)
  - Ghost (Content API)

## Migration Strategy

**From Tistory to WordPress**:

\`\`\`typescript
// Export Tistory content
const tistoryPosts = await fetchTistoryBackup();

// Transform to WordPress format
const wordpressPosts = tistoryPosts.map(post => ({
  title: post.title,
  content: transformContent(post.content),
  status: 'publish',
  categories: mapCategories(post.category),
  tags: post.tags?.split(',')
}));

// Import to WordPress
for (const post of wordpressPosts) {
  await createWordPressPost(post);
}
\`\`\`

## Best Practices (Historical Context)

- Document existing Tistory integrations before migration
- Export all Tistory content as backup
- Plan migration path to alternative platforms
- Test alternative publishing methods
- Update documentation to reflect API discontinuation
- Inform users of service changes

## Resources

- Tistory API Discontinuation Notice: https://tistory.github.io/document-tistory-apis/
- Historical API Documentation: https://www.tistory.com/guide/api/oauth
- Migration Guide: WordPress REST API documentation

## Examples

**Example 1: Historical OAuth Flow**

\`\`\`python
import requests

# Step 1: Get authorization code (user interaction required)
auth_url = f"https://www.tistory.com/oauth/authorize?client_id={CLIENT_ID}&redirect_uri={REDIRECT_URI}&response_type=code"

# Step 2: Exchange for access token
token_response = requests.post(
    'https://www.tistory.com/oauth/access_token',
    data={
        'client_id': CLIENT_ID,
        'client_secret': CLIENT_SECRET,
        'redirect_uri': REDIRECT_URI,
        'code': authorization_code,
        'grant_type': 'authorization_code'
    }
)

access_token = token_response.json()['access_token']
\`\`\`

**Example 2: Alternative - Web Automation**

\`\`\`typescript
import { chromium } from 'playwright';

const publishToTistory = async (title: string, content: string) => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  
  // Login to Tistory
  await page.goto('https://www.tistory.com/auth/login');
  await page.fill('#loginId', USERNAME);
  await page.fill('#loginPw', PASSWORD);
  await page.click('button[type="submit"]');
  
  // Navigate to post editor
  await page.goto('https://example.tistory.com/manage/newpost/');
  
  // Fill post details
  await page.fill('#post-title-inp', title);
  await page.fill('.editor-content', content);
  
  // Publish
  await page.click('.btn-publish');
  
  await browser.close();
};
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Documentation of historical API patterns, migration strategies, alternative methods
- 2024-02: Tistory Open API service discontinued

## Works well with
- \`moai-saas-wordpress-publishing\` (Alternative platform)
- \`moai-content-markdown-to-blog\` (Format conversion)
- \`moai-saas-naver-blog-publishing\` (Korean blog alternative)
