# Technical Blog Writing Plugin

**Technical Writing Excellence with Single Command** — Template-based blog creation, SEO optimization, markdown best practices, code examples, and AI discoverability.

## 🎯 What It Does

Write production-ready blog posts with a single natural language command:

```bash
/blog-write "Next.js 15 초보자 튜토리얼 작성"
/blog-write "마이그레이션으로 20% 성능 향상한 사례 작성"
/blog-write "React vs Vue 비교 분석"
```

**Automatically**:
- 🎨 Selects the right template (Tutorial, Case Study, How-to, Announcement, Comparison)
- 📝 Generates content with 7 specialist agents working in parallel
- 🔍 Optimizes for SEO (meta tags, hashtags, llms.txt)
- 💻 Creates runnable code examples
- 🖼️ Generates image prompts and diagrams
- ✅ Validates markdown quality and fixes issues

## 🏗️ Architecture

### 7 Specialist Agents

| Agent | Model | Role |
|-------|-------|------|
| **Technical Content Strategist** | Sonnet | Content strategy, audience profiling |
| **Technical Writer** | Haiku | Blog post writing with template structure |
| **SEO & Discoverability Specialist** | Haiku | Meta tags, hashtags, AI discovery optimization |
| **Code Example Curator** | Haiku | Runnable code examples generation |
| **Visual Content Designer** | Haiku | Images, diagrams, OG prompts |
| **Markdown Formatter** | Haiku | Quality assurance, linting, auto-fixes |
| **Template Workflow Coordinator** | Sonnet | Parsing, orchestration, final assembly |

### 5 Blog Templates

1. **Tutorial** — Step-by-step learning guides
2. **Case Study** — Problem → Approach → Results → Learnings
3. **How-to** — Task-oriented practical guides
4. **Announcement** — Feature/project introductions
5. **Comparison** — Tool/framework analysis

### 12 Skills

Design, content, SEO, code, templates, and markdown best practices.

## ⚡ Quick Start

### Installation

```bash
/plugin install moai-plugin-technical-blog
```

### Create Blog Post

```bash
# Auto-selects Tutorial template
/blog-write "Next.js 15 초보자 튜토리얼 작성"

# Auto-selects Case Study template
/blog-write "마이그레이션 성공 사례 작성"

# Auto-selects Comparison template
/blog-write "Next.js vs Remix 비교"
```

### Optimize Existing Post

```bash
/blog-write "./posts/nextjs-tutorial.md 최적화"
```

### List Templates

```bash
/blog-write "템플릿 목록"
```

## 📊 Output Example

When you use `/blog-write`, the plugin:

1. **Parses** your directive (auto-detects template and settings)
2. **Plans** content strategy (audience, structure, objectives)
3. **Creates** content in parallel (4 agents work simultaneously)
4. **Validates** markdown quality (linting, auto-fixes)
5. **Generates** output file with all sections complete

**Output File**: `./posts/[auto-slug].md`

**Includes**:
- ✅ YAML front matter (title, description, tags, difficulty)
- ✅ Structured content (Introduction, Steps, Resources)
- ✅ 5+ code examples (runnable, validated)
- ✅ Meta tags (60-char title, 160-char description, OG, Twitter)
- ✅ Hashtags (3-5 per platform: Twitter/X, LinkedIn, Dev.to)
- ✅ Image prompts (AI generation, diagrams, alt text)
- ✅ Markdown validated (markdownlint, H4 max depth)

## 🧠 Automatic Template Selection

The plugin intelligently selects templates based on keywords:

| Your Input | → | Template |
|-----------|---|----------|
| "...튜토리얼..." or "...가이드..." | → | Tutorial |
| "...케이스 스터디..." or "...성공 사례..." | → | Case Study |
| "...방법..." or "...어떻게..." | → | How-to |
| "...발표..." or "...릴리즈..." | → | Announcement |
| "...비교..." or "...vs..." | → | Comparison |

## 📝 Template Details

### Tutorial

Perfect for step-by-step learning guides.

**Structure**:
- Introduction (problem, learning goals)
- Prerequisites
- Step 1-5 (with code examples)
- Conclusion
- Resources

**Keywords**: "튜토리얼", "가이드", "배우기", "시작하기"

### Case Study

Great for sharing success stories with metrics.

**Structure**:
- Executive Summary
- The Problem
- The Approach (solution design, implementation)
- Results & Metrics
- Key Learnings
- Resources

**Keywords**: "케이스 스터디", "사례", "성공", "마이그레이션"

### How-to

Task-oriented practical guides.

**Structure**:
- Goal (clear statement)
- Prerequisites
- Step 1-4 (instructions, code, verification)
- Final Verification
- Troubleshooting
- Next Steps

**Keywords**: "방법", "어떻게", "구현"

### Announcement

Feature and project announcements.

**Structure**:
- What's New
- Problem Solved
- Key Features
- Getting Started
- Roadmap
- Call to Action

**Keywords**: "발표", "공지", "소개", "릴리즈"

### Comparison

Tool and framework comparisons.

**Structure**:
- Overview
- [Option A] Introduction + Strengths/Weaknesses
- [Option B] Introduction + Strengths/Weaknesses
- Detailed Comparison (6+ dimensions)
- Decision Matrix
- Conclusion

**Keywords**: "비교", "vs", "차이점"

## 🚀 Workflow Overview

```
User Input: /blog-write "directive"
           ↓
Template Coordinator (Parse & Plan)
├─ Detect template from keywords
├─ Infer difficulty level
├─ Extract topic
└─ Create execution plan
           ↓
[Phase 1] Content Strategist
├─ Analyze target audience
├─ Define learning objectives
└─ Create content structure outline
           ↓
[Phase 2-3] 4 Agents Work in Parallel
├─ Technical Writer (write content)
├─ Code Curator (generate code)
├─ SEO Specialist (optimize)
└─ Visual Designer (images/diagrams)
           ↓
[Phase 4] Markdown Formatter
├─ Run markdownlint
├─ Validate heading depth
├─ Check paragraph length
└─ Auto-fix issues
           ↓
Final Assembly & Verification
└─ Output: ./posts/[slug].md
```

## 🎨 Features

### Content Creation
- Template-based generation (5 templates)
- Paragraph structure enforcement (3-5 sentences)
- Code example generation and validation
- Image/diagram prompts

### SEO Optimization
- Keyword research and analysis
- Meta title (60 chars max)
- Meta description (150-160 chars)
- Open Graph tags (social media)
- Twitter Card tags
- Schema markup (BlogPosting)
- llms.txt knowledge base updates

### Quality Assurance
- Markdown linting (markdownlint rules)
- Heading hierarchy validation (H4 max)
- Link validation (internal/external)
- Automatic formatting fixes
- Code syntax highlighting

### Developer Experience
- Single command entry point
- Natural language understanding
- Automatic template selection
- Parallel agent execution (4x faster)
- Clear progress reporting

## 🔧 Configuration

Edit plugin settings in `plugin.json`:

```json
{
  "settings": {
    "output_directory": "./posts",
    "auto_create_frontmatter": true,
    "enable_seo_optimization": true,
    "markdown_linting": true
  }
}
```

## 📚 Examples

### Example 1: Tutorial

```bash
/blog-write "Next.js 15 초보자 튜토리얼 작성"
```

→ Creates: `./posts/nextjs-15-tutorial.md`
- Template: Tutorial
- Difficulty: beginner (inferred from "초보자")
- Content: Introduction → Prerequisites → 5 Steps → Conclusion
- Code: 5 examples with TypeScript validation
- SEO: Optimized for "next.js 15 tutorial"

### Example 2: Case Study

```bash
/blog-write "마이그레이션으로 50% 성능 향상 달성"
```

→ Creates: `./posts/migration-performance-case-study.md`
- Template: Case Study
- Focus: Business metrics and results
- Structure: Problem → Approach → Results → Learnings

### Example 3: Comparison

```bash
/blog-write "React vs Vue 상세 비교"
```

→ Creates: `./posts/react-vs-vue-comparison.md`
- Template: Comparison
- Structure: Overview → Features → Performance → Ecosystem → Decision Matrix

## 🎯 Best Practices

### For Tutorial
- Target junior to mid-level developers
- Include prerequisites upfront
- Use 5-7 clear steps
- Provide copy-paste-ready code
- Add troubleshooting section

### For Case Study
- Lead with key metric/achievement
- Quantify business impact
- Explain "why" behind decisions
- Share learnings and mistakes
- Include timeline and team size

### For How-to
- Start with clear goal statement
- Number your steps
- Include verification for each step
- Add troubleshooting section
- Link to related how-tos

### For Announcement
- Hook with "what" and "why"
- Highlight top 3-5 features
- Include "getting started" section
- Share roadmap or vision
- Add multiple CTAs

### For Comparison
- Use structured comparison tables
- Analyze 6+ dimensions
- Provide migration paths
- Include real-world examples
- End with clear decision matrix

## 🔗 Related Plugins

- **UI/UX Plugin**: Design automation, design-to-code
- **Frontend Plugin**: Next.js scaffolding, React patterns
- **Backend Plugin**: FastAPI scaffolding, database setup
- **DevOps Plugin**: Multi-cloud deployment (Vercel, Supabase, Render)

## 📖 Documentation

- Full docs: [GitHub Repository](https://github.com/moai-adk/moai-alfred-marketplace)
- Issue tracker: [GitHub Issues](https://github.com/moai-adk/moai-alfred-marketplace/issues)
- Discussions: [GitHub Discussions](https://github.com/moai-adk/moai-alfred-marketplace/discussions)

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - See LICENSE file for details

---

**Created by**: GOOS 🪿
**Version**: 2.0.0-dev
**Status**: Development
**Updated**: 2025-10-31
