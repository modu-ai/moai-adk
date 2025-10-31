# moai-content-blog-strategy

Planning blog content strategy with calendars, audience targeting, editorial planning, and performance tracking.

## Quick Start

A successful blog requires strategic planning, consistent publishing, and audience engagement. Use this skill when starting a blog, planning content calendar, defining audience personas, or measuring blog performance.

## Core Patterns

### Pattern 1: Editorial Calendar & Content Planning

**Pattern**: Create a structured editorial calendar with topic planning and publishing schedule.

```markdown
# 2024 Q4 Editorial Calendar

## Content Strategy Overview
- **Primary Focus**: Technical depth and practical tutorials
- **Publishing Schedule**: 2 posts per week (Tuesday, Friday)
- **Target Audience**: Mid-level developers (2-5 years experience)
- **Goals**: 50K monthly organic traffic, establish thought leadership

## Monthly Themes
- November: "Scaling Node.js Applications" (4 posts)
- December: "Security Best Practices" (4 posts)

## November Editorial Plan

### Week 1 (Nov 1-8)
| Date | Title | Topic | Type | Word Count | Priority |
|------|-------|-------|------|------------|----------|
| Tue, Nov 5 | Horizontal vs Vertical Scaling | Architecture | Tutorial | 2,500 | High |
| Fri, Nov 8 | Node.js Clustering Guide | Performance | Guide | 2,000 | High |

### Week 2 (Nov 9-15)
| Tue, Nov 12 | Database Connection Pooling | Performance | Deep Dive | 3,000 | High |
| Fri, Nov 15 | Load Balancing Strategies | Infrastructure | Tutorial | 2,200 | Medium |

### Week 3 (Nov 16-22)
| Tue, Nov 19 | Caching with Redis | Performance | Guide | 2,500 | High |
| Fri, Nov 22 | Monitoring Scalable Apps | Operations | How-To | 2,000 | Medium |

### Week 4 (Nov 23-29)
| Tue, Nov 26 | Case Study: Scaling to 1M RPS | Real-World | Case Study | 3,500 | Medium |
| Fri, Nov 29 | Future of Node.js Scaling | Trends | Opinion | 1,800 | Low |

## Content Types Mix
- Tutorials: 40% (practical, step-by-step)
- Guides: 30% (comprehensive, authoritative)
- Case Studies: 15% (real-world examples)
- Opinion/Trends: 15% (thought leadership)

## Content Themes Rotation
- Week 1: Architecture patterns
- Week 2: Performance optimization
- Week 3: Monitoring & observability
- Week 4: Real-world examples
```

**When to use**:
- Planning blog content for next quarter
- Coordinating team publishing schedule
- Ensuring topic variety
- Managing editorial workflow

**Key benefits**:
- Consistent publishing schedule
- Balanced content mix
- Team alignment on topics
- Better planning and resource allocation

### Pattern 2: Audience Personas & Target Definition

**Pattern**: Define target audience personas to guide content creation.

```markdown
# Audience Personas & Targeting

## Primary Persona: Mid-Level Full-Stack Developer

### Demographics
- Age: 25-35 years old
- Experience: 2-5 years of development experience
- Education: Computer Science degree or bootcamp
- Location: Primarily US/EU, some Asia
- Income: $80K-130K annually

### Psychographics
- Values: Learning, career growth, efficiency
- Frustrated by: Vague documentation, outdated tutorials
- Goals: Build better systems, improve performance, advance career

### Technical Depth
- Understands: JavaScript, Node.js, databases, APIs
- Learning: System design, DevOps, cloud architecture
- Interested in: Best practices, optimization, scalability

### Content Preferences
- Format: Detailed tutorials with code examples
- Length: 2,000-3,500 words
- Examples: Real-world problems, performance metrics
- Frequency: 2-3x per week

### Where They Are
- Twitter/LinkedIn for news
- HackerNews for in-depth articles
- Dev.to for tutorials
- Discord/Slack communities
- GitHub for code examples

## Secondary Persona: Engineering Manager

### Demographics
- Age: 30-45 years old
- Experience: 5-10+ years, managing teams
- Education: CS degree + MBA
- Location: Urban areas, major tech hubs

### Content Preferences
- Format: High-level overviews, business impact
- Length: 1,500-2,500 words
- Focus: Team productivity, scalability planning
- Frequency: 2-3x per week

## Content Gaps Analysis
- Underserved: Security in Node.js (opportunity)
- Saturated: "Getting started with Node.js" (avoid)
- Growing: Edge computing, serverless (priority)

## Audience Growth Strategy
- SEO optimization: Target 20+ keywords per month
- Community engagement: Answer questions, share insights
- Social amplification: LinkedIn, Twitter, communities
- Email newsletter: Nurture readers with updates
```

**When to use**:
- Starting a new blog
- Redefining content strategy
- Choosing which topics to cover
- Planning marketing strategy

**Key benefits**:
- Content that resonates with audience
- Better engagement and retention
- Focused resource allocation
- Measurable audience growth

### Pattern 3: Performance Metrics & Analytics

**Pattern**: Track blog performance with key metrics and KPIs.

```markdown
# Blog Performance Dashboard & Analytics

## Key Metrics to Track

### Traffic Metrics
- Monthly Organic Visits: 50,000+ (goal)
- Page Views per Session: 1.8+
- Bounce Rate: < 45%
- Avg. Session Duration: 3+ minutes

### Engagement Metrics
- Comments per Post: 5+ average
- Social Shares: 50+ per post
- Email Newsletter Subscribers: 5,000+ (goal)
- Return Visitor Rate: 30%+

### Business Metrics
- Leads Generated: 10+ per month
- Email Signups: 50+ per week
- Product Trial Signups: 5+ per month
- Customer Acquisition Cost: < $100

## Analytics Setup

### Google Analytics 4 Events
```javascript
// Track article engagement
gtag('event', 'article_read', {
  'article_id': 'nodejs-scalability',
  'article_title': 'How to Build Scalable Node.js Applications',
  'time_on_page': duration,
  'scroll_depth': percentage,
});

// Track conversions
gtag('event', 'newsletter_signup', {
  'signup_source': 'blog',
  'article_id': 'nodejs-scalability',
});
```

### Monthly Performance Report

| Metric | Oct | Nov | Target | Status |
|--------|-----|-----|--------|--------|
| Organic Visits | 32,000 | 42,000 | 50,000 | 84% |
| Avg Session Duration | 2:45 | 3:12 | 3:00 | ✅ |
| Bounce Rate | 48% | 44% | <45% | ✅ |
| Email Subscribers | 3,200 | 4,100 | 5,000 | On Track |
| Social Shares | 35 | 52 | 50 | ✅ |

## Content Performance by Topic

| Topic | Posts | Total Views | Avg Views | Top Post |
|-------|-------|-------------|-----------|----------|
| Node.js | 8 | 24,500 | 3,062 | "Clustering Guide" (5,200) |
| React | 6 | 18,200 | 3,033 | "Hooks Deep Dive" (4,100) |
| DevOps | 4 | 8,900 | 2,225 | "Docker Best Practices" (3,100) |
| Security | 3 | 6,800 | 2,267 | "JWT Security" (2,800) |

## Optimization Actions
- Topics performing well: Increase publishing frequency
- Underperforming topics: Improve SEO, revise angle
- High engagement posts: Expand into series
- Low bounce rate articles: Use as lead magnets
```

**When to use**:
- Measuring blog ROI
- Identifying top-performing content
- Planning content improvements
- Reporting to stakeholders

**Key benefits**:
- Data-driven content decisions
- Clear ROI demonstration
- Continuous improvement
- Audience insight

## Progressive Disclosure

### Level 1: Basic Strategy
- Editorial calendar planning
- Content type selection
- Publishing schedule
- Basic metrics tracking

### Level 2: Advanced Planning
- Audience persona definition
- Keyword strategy
- Email marketing integration
- Social amplification

### Level 3: Expert Growth
- Growth hacking strategies
- Advanced analytics
- SEO optimization at scale
- Thought leadership positioning

## Works Well With

- **Google Analytics**: Traffic and behavior tracking
- **Search Console**: Keyword and ranking monitoring
- **Email Platform**: Newsletter and lead nurturing
- **Social Media**: Content promotion and engagement
- **CMS**: Content management and publishing
- **GitHub**: Code examples and tutorials

## References

- **Google Analytics 4**: https://analytics.google.com/
- **Search Console**: https://search.google.com/search-console/
- **Editorial Calendar Template**: https://www.hubspot.com/
- **Content Strategy Guide**: https://www.contentmarketinginstitute.com/
- **Audience Research**: https://www.semrush.com/blog/
