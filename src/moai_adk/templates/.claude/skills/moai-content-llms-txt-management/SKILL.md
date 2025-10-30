---
name: moai-content-llms-txt-management
description: Creating and managing llms.txt knowledge base files for AI system indexing. Structure content, optimize for LLM consumption, and maintain documentation. Use when optimizing for AI search, building knowledge bases, or improving AI discoverability.
allowed-tools: Read, Bash
version: 1.0.0
tier: content
created: 2025-10-31
---

# Content: llms.txt Knowledge Base Management

## What it does

Creates and maintains llms.txt and llms-full.txt files following the 2024 standard for optimizing website content for Large Language Model (LLM) indexing, improving AI-powered search results, and enhancing content discoverability by AI systems.

## When to use

- Optimizing website for AI system indexing (ChatGPT, Perplexity, Claude)
- Building structured knowledge bases
- Improving AI search discoverability
- Creating developer documentation sites
- Automating content indexing for LLMs

## Key Patterns

### 1. llms.txt File Structure

**Pattern**: Markdown-formatted navigation document in site root

\`\`\`markdown
# Project Name

> Brief description of what this site/project is about

## Overview

Short introduction to the content and purpose.

## Documentation

- [Getting Started](/docs/getting-started): Quick start guide for new users
- [API Reference](/docs/api): Complete API documentation
- [Tutorials](/tutorials): Step-by-step learning resources

## Guides

- [Installation Guide](/guides/installation): How to install and configure
- [Best Practices](/guides/best-practices): Recommended patterns and approaches
- [Troubleshooting](/guides/troubleshooting): Common issues and solutions

## Reference

- [Configuration](/reference/config): All configuration options
- [CLI Commands](/reference/cli): Command-line interface reference
- [Architecture](/reference/architecture): System design and structure

## Resources

- [FAQ](/faq): Frequently asked questions
- [Changelog](/changelog): Version history and updates
- [Contributing](/contributing): How to contribute to the project
\`\`\`

**File Location**: `/llms.txt` (root directory)

### 2. llms-full.txt Companion File

**Pattern**: Complete content aggregation for single-request access

\`\`\`markdown
# Project Name - Complete Documentation

> This file contains the full content of all documents referenced in llms.txt

---

## Getting Started

### Installation

[Full content of getting-started.md]

### Quick Start

[Full content of quick-start.md]

---

## API Reference

### Authentication

[Full content of api-auth.md]

### Endpoints

[Full content of api-endpoints.md]

---

[Continue with all referenced documents...]
\`\`\`

**File Location**: `/llms-full.txt` (root directory)

### 3. Structured Content Organization

**Pattern**: Hierarchical, topic-based structure

\`\`\`markdown
# Next.js Documentation

## Core Concepts
- [Routing](/docs/routing): File-based routing system
  - [Dynamic Routes](/docs/routing/dynamic): [id] parameter syntax
  - [API Routes](/docs/routing/api): Backend endpoints
- [Data Fetching](/docs/data-fetching): SSR, SSG, ISR patterns
- [Rendering](/docs/rendering): Server vs Client components

## Features
- [Image Optimization](/docs/images): Next.js Image component
- [Font Optimization](/docs/fonts): Built-in font loading
- [Metadata](/docs/metadata): SEO and social sharing

## Deployment
- [Vercel](/docs/deploy/vercel): Deploy to Vercel platform
- [Self-Hosting](/docs/deploy/self-host): Docker and Node.js
\`\`\`

### 4. Auto-Generation Script

**Pattern**: Programmatically generate llms.txt from docs structure

\`\`\`typescript
import { readdirSync, readFileSync, writeFileSync } from 'fs';
import { join } from 'path';
import matter from 'gray-matter';

interface DocLink {
  title: string;
  path: string;
  description: string;
}

const generateLlmsTxt = (docsDir: string): string => {
  const categories = new Map<string, DocLink[]>();
  
  // Scan documentation directory
  const scanDocs = (dir: string, category: string) => {
    const files = readdirSync(dir, { withFileTypes: true });
    
    for (const file of files) {
      if (file.isDirectory()) {
        scanDocs(join(dir, file.name), file.name);
      } else if (file.name.endsWith('.md')) {
        const content = readFileSync(join(dir, file.name), 'utf8');
        const { data } = matter(content);
        
        if (!categories.has(category)) {
          categories.set(category, []);
        }
        
        categories.get(category)!.push({
          title: data.title || file.name,
          path: \`/docs/\${category}/\${file.name.replace('.md', '')}\`,
          description: data.description || ''
        });
      }
    }
  };
  
  scanDocs(docsDir, 'docs');
  
  // Generate llms.txt content
  let llmsTxt = '# Project Documentation\n\n';
  llmsTxt += '> Comprehensive documentation for AI-powered search\n\n';
  
  for (const [category, links] of categories) {
    llmsTxt += \`## \${category.charAt(0).toUpperCase() + category.slice(1)}\n\n\`;
    
    for (const link of links) {
      llmsTxt += \`- [\${link.title}](\${link.path})\`;
      if (link.description) {
        llmsTxt += \`: \${link.description}\`;
      }
      llmsTxt += '\n';
    }
    
    llmsTxt += '\n';
  }
  
  return llmsTxt;
};

// Usage
const llmsTxtContent = generateLlmsTxt('./docs');
writeFileSync('./public/llms.txt', llmsTxtContent);
\`\`\`

### 5. Yoast SEO Auto-Generation (WordPress)

**Pattern**: Use Yoast SEO plugin for automatic llms.txt

\`\`\`php
// Yoast SEO (June 2025+) auto-generates llms.txt

// Verify generation
curl https://yourdomain.com/llms.txt

// Expected output:
# Your Site Name

> Brief site description from Yoast settings

## Posts
- [Post Title 1](/post-1): Post excerpt
- [Post Title 2](/post-2): Post excerpt

## Pages
- [About](/about): About page description
- [Contact](/contact): Contact information
\`\`\`

### 6. Content Optimization for LLMs

**Pattern**: Structure content for AI consumption

\`\`\`markdown
## Best Practices for LLM-Friendly Content

### Use Clear Headings
# Main Title (H1)
## Section Title (H2)
### Subsection (H3)

### Provide Context
- Start each document with a brief overview
- Include purpose and target audience
- Add relevant keywords naturally

### Structure Information
- Use bullet points for lists
- Create tables for comparisons
- Include code examples with syntax highlighting
- Add links to related content

### Avoid
- Complex HTML/JavaScript (prefer Markdown)
- Excessive ads or popups
- Paywalled content for public docs
- Broken links or outdated information
\`\`\`

## Best Practices

### File Creation
- Place llms.txt in site root directory (\`/llms.txt\`)
- Use Markdown format for clarity
- Include brief descriptions for each link
- Organize by topic hierarchy
- Update regularly (monthly or when adding major content)

### Content Structure
- Start with project overview
- Group by logical categories
- Prioritize most important docs first
- Include internal links only (avoid external)
- Keep descriptions concise (1-2 sentences)

### Maintenance
- Regenerate when documentation structure changes
- Validate links regularly (no 404s)
- Remove outdated or deprecated content
- Version control llms.txt file
- Monitor AI system requests (if possible)

### Adoption Strategy
- Start with llms.txt (navigation structure)
- Add llms-full.txt for comprehensive access
- Test with AI systems (ask ChatGPT about your content)
- Monitor analytics for AI traffic
- Iterate based on performance

## Adoption Status (2025)

**Current Reality**:
- ✅ Emerging standard (introduced Sept 2024)
- 🟡 Moderate adoption in dev/AI tooling communities
- ❌ Not yet mainstream across all websites
- ❌ Google does not officially use llms.txt (per John Mueller)
- ⚠️ No confirmed major AI system support yet

**Recommended Approach**:
- Implement llms.txt as **supplementary** optimization
- Primary focus: traditional SEO + schema markup
- Monitor for official AI system support announcements
- Use as structured documentation index regardless of AI adoption

## Resources

- llms.txt Official Guide: https://llms-txt.io/
- WordLift Guide: https://wordlift.io/blog/en/mastering-llms-txt-for-genai-optimized-indexing/
- Akram Hossain Guide: https://akramhossain.com/ultimate-guide-to-llms-txt/
- Mintlify Integration: https://www.mintlify.com/blog/simplifying-docs-with-llms-txt

## Examples

**Example 1: Simple llms.txt**

\`\`\`markdown
# MoAI-ADK Documentation

> MoAI-Agentic Development Kit for SPEC-first TDD workflows

## Getting Started
- [Quick Start](/docs/quickstart): 5-minute setup guide
- [Installation](/docs/install): Installation instructions for all platforms

## Core Concepts
- [SPEC-First Development](/docs/spec-first): Understanding SPEC-driven workflows
- [TRUST Principles](/docs/trust): Quality assurance framework
- [TAG System](/docs/tags): Traceability and linking

## Guides
- [Creating Skills](/docs/skills): How to create custom Skills
- [Writing SPECs](/docs/writing-specs): EARS format guide
- [TDD Workflow](/docs/tdd): Test-driven development patterns

## Reference
- [Commands](/docs/commands): All available commands
- [Configuration](/docs/config): Configuration options
- [API](/docs/api): Programmatic API reference
\`\`\`

**Example 2: CI/CD Auto-Generation**

\`\`\`yaml
# .github/workflows/generate-llms-txt.yml
name: Generate llms.txt

on:
  push:
    paths:
      - 'docs/**'
      - 'content/**'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Generate llms.txt
        run: |
          node scripts/generate-llms-txt.js
      
      - name: Commit changes
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add public/llms.txt
          git commit -m "docs: Update llms.txt [skip ci]"
          git push
\`\`\`

**Example 3: Full Content Aggregation**

\`\`\`typescript
import { readdirSync, readFileSync, writeFileSync } from 'fs';
import { join } from 'path';

const generateLlmsFullTxt = (docsDir: string): string => {
  let fullContent = '# Complete Documentation\n\n';
  fullContent += '> All content aggregated for AI system access\n\n';
  
  const files = readdirSync(docsDir);
  
  for (const file of files) {
    if (file.endsWith('.md')) {
      const content = readFileSync(join(docsDir, file), 'utf8');
      fullContent += \`---\n\n## \${file.replace('.md', '')}\n\n\`;
      fullContent += content + '\n\n';
    }
  }
  
  return fullContent;
};

// Generate both files
const llmsTxt = generateLlmsTxt('./docs');
const llmsFullTxt = generateLlmsFullTxt('./docs');

writeFileSync('./public/llms.txt', llmsTxt);
writeFileSync('./public/llms-full.txt', llmsFullTxt);

console.log('Generated llms.txt and llms-full.txt');
\`\`\`

## Changelog
- 2025-10-31: v1.0.0 - Initial release with llms.txt standard, auto-generation, best practices

## Works well with
- \`moai-content-seo-optimization\` (Complement SEO strategy)
- \`moai-content-blog-strategy\` (Organize content structure)
- \`moai-saas-wordpress-publishing\` (WordPress Yoast integration)
