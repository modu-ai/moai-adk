---
title: Getting Started with AI Agency
weight: 20
draft: false
---

{{< callout type="info" >}}
Complete your first project in 15 minutes by following this step-by-step guide. No prior AI knowledge required.
{{< /callout >}}

## Prerequisites

Before starting, ensure you have:

- Claude Code v2.1.50+ installed
- moai-adk-go installed locally
- A brief description of your project (2-3 sentences minimum)
- 15-30 minutes of uninterrupted time

## Step 1: Create Your First Brief

A brief is Agency's starting point. It captures your vision and business goals.

Execute the following command in Claude Code:

```
/agency brief "AI startup landing page showcasing research capabilities, 
target audience is academic institutions and research labs, emphasize 
collaboration and data insights"
```

Agency creates a BRIEF document in `.agency/briefs/BRIEF-001/` containing:

- **Goal**: What success looks like
- **Outcome**: Specific deliverables and metrics
- **Context**: Brand voice, target audience, constraints
- **Requirements**: Feature list and acceptance criteria

The planner agent expands your 2-3 sentence input into a comprehensive 500-line specification with goals, acceptance criteria, and technical requirements.

### What Happens During Brief Creation

When you submit a brief, the system:

1. **Client Interview**: Agent asks clarifying questions about brand, audience, and success metrics
2. **Context Assembly**: Builds brand context from your responses (stored in `.agency/context/`)
3. **BRIEF Generation**: Creates structured specification with EARS-format requirements
4. **Validation**: Checks specification completeness and feasibility

Brand context is stored in 5 files that guide all future decisions:

- `brand.md` - Voice, tone, visual identity
- `audience.md` - Target customer profile and goals
- `business.md` - Success metrics and constraints
- `technical.md` - Stack preferences and integrations
- `governance.md` - Quality standards and safety guidelines

## Step 2: Build Your Project

Once you have a BRIEF document, execute the build command:

```
/agency build BRIEF-001
```

This launches the full pipeline: planner expands requirements, copywriter creates marketing content, designer builds UI specifications, builder implements code with tests, and evaluator ensures quality.

### What the Build Pipeline Does

The pipeline executes 5 sequential phases:

**Phase 1 - Requirement Expansion (2 min)**
Planner analyzes your BRIEF and generates detailed specifications with acceptance criteria, technical requirements, and component breakdown.

**Phase 2 - Parallel Execution (5 min)**
Copywriter and Designer work simultaneously. Copywriter generates marketing copy with JSON-structured messaging hierarchy. Designer creates UI specifications with component designs and layout constraints.

**Phase 3 - Implementation (8 min)**
Builder implements code using Test-Driven Development. Tests are written first, then implementation follows. All code includes comments and docstrings.

**Phase 4 - Evaluation (2 min)**
Evaluator runs automated quality tests:
- Design Quality (alignment with brand)
- Code Quality (test coverage > 85%)
- Performance (load time < 2s)
- Accessibility (WCAG 2.1 AA compliance)

**Phase 5 - GAN Loop (conditional)**
If evaluator rejects output, learner agent analyzes failures and improves. Generator retries until quality threshold is met (max 5 iterations).

### Step-by-Step Mode

For more control, use step-by-step mode:

```
/agency build --step BRIEF-001
```

This executes each phase separately, allowing you to review output before proceeding. Useful for learning how the system works or making adjustments between phases.

### Team Parallel Mode

For faster execution with multiple agents:

```
/agency build --team BRIEF-001
```

This spawns parallel agents for copywriting, design, and code review. Team mode is 30-40% faster than sequential execution.

## Step 3: Review and Provide Feedback

After build completes, Agency outputs a live website URL. Review the result and collect feedback:

```
/agency learn
```

This command launches an interactive feedback collection session where you specify:

**What worked well** - Features and design choices to keep
**What needs improvement** - Specific changes or refinements
**Design preferences** - Brand alignment, visual style improvements
**Copy feedback** - Messaging clarity, tone adjustments
**Technical issues** - Bugs or missing functionality

Feedback is stored in `.agency/feedback/` and feeds the learning loop.

## Step 4: Evolve Your Skills

After 3+ projects with feedback, Agency identifies patterns and evolves its decision-making:

```
/agency evolve
```

This command triggers the learner agent to:

1. **Analyze feedback patterns** across all projects
2. **Identify successful patterns** (what works consistently)
3. **Generate rule candidates** (new decision heuristics)
4. **Validate candidates** against quality thresholds
5. **Apply improvements** to relevant skill modules

Evolution is incremental. Each skill generation (gen-001, gen-002, etc.) represents improvements based on accumulated feedback.

## Brand Context Setup Guide

Agency requires brand context in `.agency/context/` to operate effectively. Create these 5 files:

### 1. Brand Context (brand.md)

Define your brand voice and visual identity:

```markdown
# Brand Context

## Voice & Tone
- Professional yet approachable
- Authoritative but friendly
- Technical when necessary, accessible always

## Visual Identity
- Primary Color: #0066CC (Professional Blue)
- Secondary: #FF6B35 (Accent Orange)
- Font: Inter (sans-serif), Merriweather (serif)
- Spacing: 16px base grid
- Radius: 8px default border radius

## Key Phrases
- Always: "empowering innovation"
- Never: clichés about "disruption"
- Brand personality: innovative, trustworthy, human-centered
```

### 2. Audience Context (audience.md)

Describe your target customers:

```markdown
# Audience Profile

## Primary Persona
- Name: Research Director, PhD
- Age: 35-55
- Goals: Improve research collaboration, reduce time on data management
- Pain Points: Fragmented tools, data silos, learning curves
- Tech Savviness: Moderate (can use software but prefers simplicity)

## Secondary Persona
- Name: Lab Manager
- Age: 28-40
- Goals: Streamline workflows, ensure data quality
- Pain Points: Manual processes, version control issues
```

### 3. Business Context (business.md)

Define success metrics:

```markdown
# Business Goals

## Success Metrics
- Homepage load < 2 seconds
- Mobile usability: 90+ score
- Conversion rate: 5%+ from visitor to trial signup
- SEO: Top 10 for target keywords within 90 days

## Constraints
- Budget: $5K (automated, no manual design changes)
- Timeline: 2 weeks to launch
- Stack: Next.js, React, TypeScript, TailwindCSS
```

### 4. Technical Context (technical.md)

Specify technology preferences:

```markdown
# Technical Preferences

## Frontend Stack
- Framework: Next.js 15+
- UI Library: shadcn/ui (Radix + Tailwind)
- Animations: Framer Motion
- Forms: React Hook Form + Zod validation

## Backend Requirements
- API: RESTful (not GraphQL)
- Database: PostgreSQL
- Hosting: Vercel (frontend), Railway (backend)
- Auth: Clerk

## Integrations
- Analytics: Posthog
- Email: Resend
- Monitoring: Sentry
```

### 5. Governance Context (governance.md)

Set safety and compliance guidelines:

```markdown
# Governance & Safety

## Content Guidelines
- No claims without data backing
- Avoid competitor FUD (fear, uncertainty, doubt)
- Accessibility: WCAG 2.1 AA minimum
- Privacy: No third-party trackers, self-hosted analytics only

## Brand Safety
- Approved color palette (see brand.md)
- No arbitrary content changes
- All copy changes reviewed before publish
- Data retention: 30 days maximum

## Compliance
- GDPR: Cookie consent, privacy policy
- CCPA: Data deletion mechanism
- Terms of Service: Posted and accessible
```

## Directory Structure

Agency creates the following structure in `.agency/`:

```
.agency/
├── context/              # Brand and business context (immutable)
│   ├── brand.md
│   ├── audience.md
│   ├── business.md
│   ├── technical.md
│   └── governance.md
├── briefs/              # Project specifications
│   ├── BRIEF-001/
│   │   └── spec.md
│   └── BRIEF-002/
├── feedback/            # User feedback for learning
│   └── BRIEF-001/
├── learnings/           # Extracted patterns and insights
│   ├── copywriting.md
│   ├── design.md
│   └── patterns.md
├── skills/              # Evolved skill modules
│   ├── copywriter/
│   ├── designer/
│   └── builder/
└── config.yaml          # Configuration and settings
```

## Just Do It Mode vs Step-by-Step Mode

### Just Do It Mode (Default)

```
/agency build BRIEF-001
```

Recommended for experienced users or simple projects. Agency executes the full pipeline automatically and delivers a complete website.

**Advantages**: Fast (15 minutes), minimal interaction, trusted algorithm

**Disadvantages**: Less control, harder to debug if something goes wrong

### Step-by-Step Mode

```
/agency build --step BRIEF-001
```

Recommended for learning or complex projects requiring adjustments. Execute each phase separately and review output before proceeding.

**Advantages**: Full control, educational, ability to fix issues at each stage

**Disadvantages**: Slower (45+ minutes), requires more decision-making

## What Happens Next?

After your first project:

1. **Review the live site** at the provided URL
2. **Provide feedback** with `/agency learn` to teach the system
3. **Build a second project** to see learning in action
4. **Run evolution** after 3+ projects to improve skills

The more feedback you provide, the better Agency becomes at understanding your brand and producing results that require less revision.

{{< callout type="info" >}}
You're ready! Start with `/agency brief "your project description"` and watch as AI Agency transforms your vision into a live website.
{{< /callout >}}
