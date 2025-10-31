# moai-content-llms-txt-management

Creating and managing llms.txt knowledge base files for AI system indexing, structure content, and optimize for LLM consumption.

## Quick Start

llms.txt files help AI models understand your project structure, documentation, and key concepts. Use this skill when creating knowledge bases for AI training, documenting code repositories, or improving AI discoverability.

## Core Patterns

### Pattern 1: Creating llms.txt for Projects

**Pattern**: Structure project information in llms.txt format for AI indexing.

```txt
# Project Name
Technical Blog Platform with Design System

## What is this?
A comprehensive platform for publishing technical content with integrated design systems, analytics, and multi-platform publishing capabilities.

## Overview
- **Purpose**: Enable technical writers to publish content across multiple platforms with consistent branding and design
- **Technology Stack**: Next.js 16, React 19, TypeScript, Tailwind CSS 4, shadcn/ui, Supabase, Vercel
- **Target Audience**: Software engineers, technical writers, and developers

## Key Concepts

### Architecture
- **Frontend**: Next.js 16 with App Router, Server Components, and streaming
- **Design System**: shadcn/ui components with Tailwind CSS v4 for theming
- **Database**: Supabase PostgreSQL with row-level security
- **Deployment**: Vercel for frontend, Render for backend services

### Core Features
1. **Multi-platform Publishing**
   - WordPress integration via REST API
   - Naver Blog publishing for Korean audiences
   - Tistory integration for Korean market
   - Medium and Dev.to cross-posting

2. **Design System**
   - Component library based on shadcn/ui
   - Design tokens and CSS variables
   - Dark mode support with system preference detection
   - Responsive design mobile-first approach

3. **Content Management**
   - Markdown-based content authoring
   - SEO optimization with meta tags and structured data
   - Image generation with DALL-E
   - Content scheduling and analytics

### Technical Highlights
- **Type Safety**: Full TypeScript with strict mode enabled
- **Performance**: Core Web Vitals optimization, image optimization
- **Security**: Row-level security in Postgres, CORS configuration
- **Scalability**: Horizontal scaling with load balancing, caching with Redis

## File Structure

### Key Directories
```
.
├── app/                      # Next.js App Router
│   ├── blog/                # Blog pages
│   ├── admin/               # Admin dashboard
│   └── api/                 # API routes
├── components/              # React components
│   ├── ui/                  # shadcn/ui components
│   ├── blog/                # Blog-specific components
│   └── shared/              # Shared components
├── lib/                     # Utility functions
│   ├── db.ts                # Database functions
│   ├── markdown-parser.ts   # Markdown processing
│   └── seo.ts               # SEO utilities
├── styles/                  # Global styles
├── types/                   # TypeScript types
└── public/                  # Static assets
```

### Important Files
- `lib/db.ts` - Database queries and functions (100+ lines)
- `lib/markdown-parser.ts` - Parse and convert markdown (50+ lines)
- `components/ui/button.tsx` - Custom button component (30+ lines)
- `app/blog/[slug]/page.tsx` - Blog post page (40+ lines)

## Technology Decisions

### Why Next.js 16?
- Server Components reduce JavaScript sent to browser
- Built-in image optimization
- API routes for backend functionality
- Official hosting on Vercel with zero-config deploys

### Why Tailwind CSS + shadcn/ui?
- Utility-first CSS for rapid development
- Pre-built accessible components
- Easy customization with CSS variables
- Dark mode support out of the box

### Why Supabase?
- PostgreSQL with full-featured database
- Built-in authentication with OAuth
- Real-time subscriptions
- Row-level security for multi-tenancy

## How to Use This Project

### For Content Authors
1. Write content in Markdown format
2. Include frontmatter with metadata (title, date, tags)
3. Images referenced in markdown are optimized automatically
4. Click publish to deploy across all platforms

### For Developers
1. Clone repository: `git clone <repo>`
2. Install dependencies: `npm install`
3. Setup environment variables in `.env.local`
4. Run development server: `npm run dev`
5. Open http://localhost:3000

### For Designers
1. Review Figma design file for component specs
2. All components in `components/ui/` match design system
3. Update Tailwind config in `tailwind.config.ts` for theming
4. Dark mode variants defined in `globals.css`

## Key Documentation

### Component Library
- `docs/COMPONENTS.md` - Component usage guide
- `docs/DESIGN_TOKENS.md` - Color, typography, spacing system
- `docs/THEMING.md` - Dark mode and custom themes

### API Reference
- `docs/API.md` - REST API endpoints
- `docs/DATABASE.md` - Schema and relationships
- `docs/AUTHENTICATION.md` - Auth flows and JWT

### Deployment
- `docs/DEPLOYMENT.md` - Vercel and Render setup
- `docs/ENVIRONMENT_VARIABLES.md` - Configuration reference
- `docs/CI_CD.md` - GitHub Actions workflows

## Common Questions

### How do I add a new blog post?
1. Create markdown file in `content/blog/`
2. Include frontmatter with title, date, tags
3. Run `npm run publish` to sync across platforms

### How do I customize colors?
1. Edit `globals.css` CSS variables
2. Update `tailwind.config.ts` theme section
3. Components automatically use new colors

### How do I add a new component?
1. Copy existing component from `components/ui/`
2. Customize for your needs
3. Export from `components/index.ts`
4. Use in pages or other components

### How do I deploy?
1. Push to main branch on GitHub
2. Vercel automatically deploys frontend
3. GitHub Actions trigger backend deployment to Render
4. Check deployment status in GitHub Actions tab

## Related Projects
- Design System: See `docs/DESIGN_SYSTEM.md`
- Content Pipeline: See `lib/content-pipeline/`
- Analytics: See `lib/analytics.ts`
- Email Templates: See `emails/`

## Learning Resources
- Next.js: https://nextjs.org/docs
- React 19: https://react.dev
- TypeScript: https://www.typescriptlang.org/docs
- Tailwind CSS: https://tailwindcss.com/docs
- shadcn/ui: https://ui.shadcn.com

## Contact & Support
- Issues: GitHub Issues
- Discussions: GitHub Discussions
- Email: support@example.com
```

**When to use**:
- Creating project documentation
- Setting up AI training for your codebase
- Onboarding new team members
- Improving code discoverability

**Key benefits**:
- Structured knowledge for AI systems
- Better code understanding
- Easier onboarding
- AI-friendly documentation

### Pattern 2: Creating Knowledge Base for Monorepos

**Pattern**: Organize llms.txt for monorepo with multiple packages.

```txt
# Monorepo Structure
MoAI-ADK: Multi-Agent Development Kit

## Repository Overview
- **Type**: Python monorepo with multiple packages
- **Purpose**: Development toolkit for agentic workflows
- **Structure**: Single-version monorepo with shared dependencies

## Packages

### moai-adk (Core Package)
Location: `src/moai_adk/`
- Purpose: Main development kit implementation
- Key Modules:
  - `phase_executor.py` - Workflow orchestration
  - `config.py` - Configuration management
  - `cli.py` - Command-line interface
- Dependencies: Python 3.13+, pydantic 2.12+, click 8.0+

### moai-marketplace (Plugin Marketplace)
Location: `moai-marketplace/`
- Purpose: Plugin distribution and management
- Plugins:
  - moai-plugin-backend - FastAPI and database patterns
  - moai-plugin-devops - Multi-cloud deployment
  - moai-plugin-frontend - Next.js and React patterns
  - moai-plugin-uiux - Design system and Figma integration
  - moai-plugin-technical-blog - Content creation and publishing

### moai-templates (Project Templates)
Location: `templates/`
- Purpose: Scaffolding and project initialization
- Languages:
  - Python backend template
  - React frontend template
  - Full-stack template

## Development Workflow
1. Install: `pip install -e .` from root
2. Test: `pytest tests/`
3. Type Check: `mypy src/`
4. Format: `ruff format src/`

## Key Concepts
- **Phase System**: Four-phase workflow (SPEC → BUILD → SYNC → REPORT)
- **Sub-agents**: Specialized AI agents for different tasks
- **Skills**: Reusable knowledge capsules
- **Commands**: User-facing workflow entry points
```

**When to use**:
- Managing large, complex projects
- Documenting monorepos
- Creating AI-friendly project guides
- Improving code organization

**Key benefits**:
- Clear project structure
- Better AI understanding
- Easier navigation
- Improved maintainability

### Pattern 3: Maintenance & Updates

**Pattern**: Keep llms.txt synchronized with project changes.

```typescript
// lib/llms-txt-generator.ts - Auto-generate llms.txt from code
import fs from 'fs/promises';
import path from 'path';

interface ProjectMetadata {
  name: string;
  description: string;
  version: string;
  techStack: string[];
  mainFiles: Record<string, string>;
}

async function generateLlmsTxt(
  projectRoot: string,
  metadata: ProjectMetadata
): Promise<string> {
  const fileStructure = await analyzeProjectStructure(projectRoot);
  const keyFiles = await extractKeyFiles(projectRoot);

  return `
# ${metadata.name}
${metadata.description}

## Project Info
- **Version**: ${metadata.version}
- **Tech Stack**: ${metadata.techStack.join(', ')}

## File Structure
\`\`\`
${formatFileTree(fileStructure)}
\`\`\`

## Key Files
${Object.entries(keyFiles)
  .map(([path, description]) => `- \`${path}\` - ${description}`)
  .join('\n')}

## Getting Started
1. Clone repository
2. Install dependencies: \`${metadata.techStack.includes('python') ? 'pip install' : 'npm install'}\`
3. Configure environment variables
4. Run tests and development server

## Documentation
- See \`docs/\` directory for detailed documentation
- See \`README.md\` for overview

## Contact
- Issues: GitHub Issues
- Discussions: GitHub Discussions
  `.trim();
}

async function analyzeProjectStructure(
  projectRoot: string
): Promise<string> {
  // Analyze directory structure
  // Exclude node_modules, .git, build artifacts
  // Return formatted tree

  const tree = await buildDirectoryTree(projectRoot, {
    exclude: ['node_modules', '.git', 'dist', 'build', '.next'],
    maxDepth: 3,
  });

  return tree;
}

async function extractKeyFiles(
  projectRoot: string
): Promise<Record<string, string>> {
  const keyFiles: Record<string, string> = {};

  // Find important files
  const importantPatterns = [
    { pattern: /main\.(py|ts|js)$/, description: 'Entry point' },
    { pattern: /config\.(py|ts|js)$/, description: 'Configuration' },
    { pattern: /README\.md/, description: 'Project overview' },
    { pattern: /package\.json/, description: 'Dependencies' },
    { pattern: /pyproject\.toml/, description: 'Python project config' },
  ];

  // Scan and categorize files
  // Return with descriptions

  return keyFiles;
}

// CLI command to generate/update llms.txt
async function updateLlmsTxt(projectRoot: string) {
  const metadata: ProjectMetadata = {
    name: 'My Project',
    description: 'Project description',
    version: '1.0.0',
    techStack: ['TypeScript', 'React', 'Node.js'],
    mainFiles: {},
  };

  const content = await generateLlmsTxt(projectRoot, metadata);

  const llmsTxtPath = path.join(projectRoot, 'llms.txt');
  await fs.writeFile(llmsTxtPath, content);

  console.log(`Updated ${llmsTxtPath}`);
}
```

**When to use**:
- Keeping documentation synchronized
- Automating documentation generation
- Updating project guides
- Tracking structure changes

**Key benefits**:
- Always current documentation
- Less manual maintenance
- Better consistency
- Faster updates

## Progressive Disclosure

### Level 1: Basic Knowledge Base
- Project overview
- File structure
- Key concepts
- Getting started guide

### Level 2: Advanced Documentation
- Detailed architecture
- API reference
- Component guide
- Common patterns

### Level 3: Expert Knowledge Management
- Automated generation
- Continuous synchronization
- AI training optimization
- Version tracking

## Works Well With

- **AI Models**: Claude, GPT-4, local LLMs
- **Code Documentation**: README, API docs, guides
- **Project Management**: GitHub, Jira, Linear
- **Knowledge Systems**: Notion, Wiki, Obsidian
- **CI/CD**: GitHub Actions for automatic updates

## References

- **llms.txt Specification**: https://llms.txt
- **Knowledge Management**: https://github.com/mckaywrigley/llm-docs
- **Project Documentation**: https://www.writethedocs.org/
- **Knowledge Bases**: https://en.wikipedia.org/wiki/Knowledge_base
