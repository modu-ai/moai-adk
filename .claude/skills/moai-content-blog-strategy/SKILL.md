---
name: moai-content-blog-strategy
description: Planning blog content strategy with calendars, audience targeting, and topic research. Create publishing schedules, define goals, and track performance. Use when planning content, building editorial calendars, or strategizing blog growth.
allowed-tools: Read, Bash
version: 1.0.0
tier: content
created: 2025-10-31
---

# Content: Blog Strategy and Planning

## What it does

Provides frameworks for creating data-driven blog content strategies including audience targeting, topic research, editorial calendar planning, and goal setting to achieve consistent engagement and growth.

## When to use

- Starting a new blog or content initiative
- Planning quarterly/annual content calendars
- Defining content goals and KPIs
- Researching topics and keywords
- Building editorial workflows
- Analyzing content performance

## Key Patterns

### 1. Goal Definition Framework

**Pattern**: SMART goals for blog content

\`\`\`markdown
## SMART Goal Template

Specific: "Increase organic traffic to documentation pages"
Measurable: "by 50% (from 10K to 15K monthly visitors)"
Achievable: "through SEO optimization and weekly publishing"
Relevant: "to drive product adoption and reduce support tickets"
Time-bound: "within 6 months (by April 2026)"

## Blog Goal Categories

### Traffic Goals
- Target: 50K monthly organic visitors
- Method: SEO-optimized tutorials, guides
- KPI: Google Analytics pageviews, unique visitors

### Engagement Goals
- Target: 10% increase in avg. time on page
- Method: Interactive content, clear structure
- KPI: Bounce rate, session duration

### Conversion Goals
- Target: 500 newsletter signups/month
- Method: Lead magnets, CTAs in posts
- KPI: Conversion rate, signup forms

### Authority Goals
- Target: 20 backlinks from DA 50+ sites
- Method: High-quality research posts
- KPI: Domain authority, referring domains
\`\`\`

### 2. Audience Persona Development

**Pattern**: Define target reader segments

\`\`\`markdown
## Persona Template

### Persona 1: Junior Developer
**Demographics**: 22-28 years, 0-2 years experience
**Pain Points**: 
- Learning modern frameworks
- Understanding best practices
- Building portfolio projects
**Content Needs**:
- Step-by-step tutorials
- Code examples with explanations
- Common error solutions
**Topics**: "Getting Started with...", "Beginner's Guide to..."

### Persona 2: Senior Engineer
**Demographics**: 30-45 years, 5+ years experience
**Pain Points**:
- Scaling systems
- Architecture decisions
- Team leadership
**Content Needs**:
- Case studies
- Architecture patterns
- Performance optimization
**Topics**: "Scaling...", "Advanced Patterns", "Best Practices"
\`\`\`

### 3. Topic Research Process

**Pattern**: Keyword research + competitor analysis + audience questions

\`\`\`bash
## Step 1: Keyword Research

Tools: Google Keyword Planner, Ahrefs, SEMrush

Primary Keywords:
- "Next.js tutorial" (8.1K searches/month, Medium difficulty)
- "React Server Components" (3.2K searches/month, Low difficulty)
- "SSR vs SSG" (1.5K searches/month, Medium difficulty)

Long-tail Keywords:
- "how to deploy Next.js on Vercel" (500 searches/month)
- "Next.js 15 new features" (300 searches/month)

## Step 2: Competitor Analysis

Top-ranking posts:
1. Vercel Docs - "Next.js Documentation" (DA 92)
2. LogRocket - "Next.js Tutorial" (DA 75)
3. Dev.to - "Getting Started with Next.js" (DA 85)

Content Gaps:
- No comprehensive migration guide from Create React App
- Limited Next.js 15 App Router tutorials
- Missing performance optimization deep dives

## Step 3: Audience Questions

Sources: Reddit, Stack Overflow, Twitter, Product Hunt

Common Questions:
- "How to handle authentication in Next.js?"
- "Next.js vs Remix: which to choose?"
- "Best practices for Next.js file structure?"

## Step 4: Topic Prioritization

| Topic | Search Volume | Difficulty | Business Value | Priority |
|-------|--------------|------------|----------------|----------|
| Next.js 15 Guide | 8.1K | Medium | High | 🔥 High |
| Auth in Next.js | 2.3K | Low | High | 🔥 High |
| Next.js vs Remix | 1.8K | Medium | Medium | 🟡 Medium |
\`\`\`

### 4. Editorial Calendar Structure

**Pattern**: Monthly calendar with themes and milestones

\`\`\`markdown
## Q1 2026 Editorial Calendar

### January Theme: "Getting Started with Modern React"

Week 1 (Jan 1-7):
- Mon: "React Server Components Explained" (Tutorial)
- Wed: "Setting Up Next.js 15 Project" (Guide)
- Fri: "Common React Mistakes to Avoid" (Listicle)

Week 2 (Jan 8-14):
- Mon: "Building Your First API Route" (Tutorial)
- Wed: "Next.js Routing Deep Dive" (Technical)
- Fri: "Interview: Next.js Core Team" (Case Study)

Week 3 (Jan 15-21):
- Mon: "Authentication with NextAuth.js" (Tutorial)
- Wed: "Optimizing Images in Next.js" (Performance)
- Fri: "5 Next.js Projects for Your Portfolio" (Inspiration)

Week 4 (Jan 22-28):
- Mon: "Deploying to Vercel vs Netlify" (Comparison)
- Wed: "Next.js SEO Best Practices" (SEO)
- Fri: "Monthly Roundup: React Ecosystem News"

### Milestones:
- Jan 15: Launch newsletter (goal: 100 subscribers)
- Jan 31: Publish 12 posts, target 5K visitors
\`\`\`

### 5. Content Types Mix

**Pattern**: Balance educational, promotional, and evergreen content

\`\`\`markdown
## Content Type Distribution (Monthly)

40% Educational (Tutorials, Guides)
- Step-by-step instructions
- Code examples
- Learning paths

30% Evergreen (Best Practices, Patterns)
- Architecture guides
- Design patterns
- Reference materials

20% Timely (News, Updates, Releases)
- Framework updates
- Industry news
- Tool releases

10% Promotional (Case Studies, Product)
- Customer success stories
- Feature announcements
- Company updates
\`\`\`

### 6. Publishing Frequency

**Pattern**: Consistent cadence based on resources

\`\`\`markdown
## Publishing Schedule Options

### Option 1: High Frequency (5x/week)
- Best for: Established blogs with dedicated writers
- Resource: 3+ writers, editor
- Benefit: High SEO momentum, audience engagement

### Option 2: Medium Frequency (3x/week)
- Best for: Growing blogs, small teams
- Resource: 1-2 writers
- Benefit: Sustainable, quality-focused

### Option 3: Low Frequency (1x/week)
- Best for: Solo bloggers, limited time
- Resource: 1 writer
- Benefit: Maintainable, deep content

## Recommended: Start with 2x/week
- Monday: Educational/Tutorial
- Thursday: Evergreen/Reference
- Adjust based on analytics and capacity
\`\`\`

## Best Practices

### Strategy Development
- Define clear, measurable goals before creating content
- Research target audience thoroughly (surveys, interviews)
- Analyze competitors for content gaps
- Balance SEO optimization with user value
- Plan 3 months ahead minimum

### Content Planning
- Build around business milestones (product launches, seasons)
- Mix content types (tutorials, guides, case studies)
- Account for evergreen vs timely content ratio
- Leave 20% buffer for trending topics

### Calendar Management
- Use tools: Notion, Airtable, Google Sheets, Asana
- Include all metadata: author, keywords, publish date, status
- Track performance: views, engagement, conversions
- Review and adjust monthly

### Performance Tracking
- Set up Google Analytics 4 (GA4)
- Track: pageviews, bounce rate, time on page, conversions
- Use Google Search Console for search performance
- Monitor social shares and backlinks
- Review monthly, adjust strategy quarterly

### Team Collaboration
- Define clear roles: writer, editor, SEO specialist
- Create style guide and templates
- Use version control for content (Git, Google Docs)
- Weekly sync meetings for alignment

## Resources

- Content Calendar Templates: https://coda.io/@atc/content-calendar-template-updated-2025
- Keyword Research: https://www.semrush.com/blog/content-calendar/
- SEMrush Topic Research: https://www.semrush.com/topic-research/
- Productive Blogging Guide: https://www.productiveblogging.com/blog-content-plan/

## Examples

**Example 1: Quarterly Strategy Document**

\`\`\`markdown
# Q1 2026 Blog Strategy

## Goals
1. Increase organic traffic from 10K to 15K/month (+50%)
2. Achieve 500 newsletter subscribers
3. Rank #1 for "Next.js 15 tutorial"

## Target Audience
- Junior to mid-level React developers
- Learning Next.js for first time or upgrading

## Content Pillars
1. Next.js 15 Features (30%)
2. React Best Practices (25%)
3. Performance Optimization (25%)
4. Deployment & DevOps (20%)

## Publishing Schedule
- 2 posts/week (Tuesdays & Thursdays)
- 24 total posts in Q1

## Success Metrics
- Pageviews: 15K/month
- Avg time on page: >4 minutes
- Newsletter signups: 500
- Backlinks acquired: 10
\`\`\`

**Example 2: Topic Ideation Matrix**

\`\`\`markdown
| Topic Idea | Keyword | Volume | Difficulty | Business Value | Priority |
|------------|---------|--------|------------|----------------|----------|
| Next.js 15 migration guide | "migrate to nextjs 15" | 800 | Low | High | 🔥 |
| Server Actions tutorial | "nextjs server actions" | 1.2K | Medium | High | 🔥 |
| Image optimization | "nextjs image optimization" | 900 | Low | Medium | 🟡 |
| Auth patterns | "nextjs authentication" | 2.3K | Medium | High | 🔥 |
| Performance tips | "nextjs performance" | 1.5K | High | Medium | 🟡 |
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with goal frameworks, editorial calendar, audience personas

## Works well with
- \`moai-content-seo-optimization\` (Optimize planned content)
- \`moai-content-image-generation\` (Create visuals for calendar)
- \`moai-saas-wordpress-publishing\` (Execute publishing schedule)
