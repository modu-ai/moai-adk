# Claude Code Plugin Examples

## Example 1: Minimal Plugin (Metadata Only)

**Use Case**: Simple configuration plugin with no custom components

**Directory Structure**:
```
my-config-plugin/
├── .claude-plugin/
│   └── plugin.json
└── README.md
```

**plugin.json**:
```json
{
  "name": "my-config-plugin",
  "description": "Project configuration and settings management",
  "version": "1.0.0",
  "author": {
    "name": "John Doe"
  }
}
```

**When to Use**:
- Configuration-only plugins
- Metadata packages
- Placeholder for future development

---

## Example 2: Skills-Only Plugin

**Use Case**: Knowledge base plugin providing reusable patterns without automation

**Directory Structure**:
```
knowledge-base/
├── .claude-plugin/
│   └── plugin.json
├── skills/
│   ├── backend-patterns/
│   │   ├── SKILL.md
│   │   ├── reference.md
│   │   └── examples.md
│   ├── frontend-patterns/
│   │   ├── SKILL.md
│   │   └── examples.md
│   └── api-design.md
└── README.md
```

**plugin.json**:
```json
{
  "name": "knowledge-base",
  "description": "Curated development patterns and best practices",
  "version": "2.1.0",
  "author": {
    "name": "Development Team",
    "email": "dev@example.com"
  },
  "skills": [
    "./skills/backend-patterns/SKILL.md",
    "./skills/frontend-patterns/SKILL.md",
    "./skills/api-design.md"
  ],
  "category": "documentation",
  "tags": ["patterns", "best-practices", "architecture"]
}
```

**When to Use**:
- Documentation plugins
- Pattern libraries
- Educational resources
- Domain-specific knowledge capsules

---

## Example 3: Command-Driven Plugin

**Use Case**: Workflow automation with slash commands

**Directory Structure**:
```
workflow-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   ├── init-project.md
│   ├── run-tests.md
│   ├── deploy.md
│   └── create-spec.md
└── README.md
```

**plugin.json**:
```json
{
  "name": "workflow-plugin",
  "description": "Automate common development workflows and project tasks",
  "version": "1.5.0",
  "author": {
    "name": "DevTools Team"
  },
  "commands": [
    {
      "name": "init-project",
      "description": "Initialize new project structure with templates"
    },
    {
      "name": "run-tests",
      "description": "Execute full test suite with coverage reporting"
    },
    {
      "name": "deploy",
      "description": "Deploy application to staging or production"
    },
    {
      "name": "create-spec",
      "description": "Generate SPEC document from requirements"
    }
  ],
  "category": "workflow",
  "tags": ["automation", "productivity", "workflow"]
}
```

**When to Use**:
- Task automation
- Project scaffolding
- Deployment workflows
- Testing automation

---

## Example 4: Agent-Powered Plugin

**Use Case**: Specialized AI agents for domain expertise

**Directory Structure**:
```
backend-plugin/
├── .claude-plugin/
│   └── plugin.json
├── agents/
│   ├── api-designer.md
│   ├── db-optimizer.md
│   ├── security-auditor.md
│   └── performance-analyst.md
└── README.md
```

**plugin.json**:
```json
{
  "name": "backend-plugin",
  "description": "FastAPI 0.120.2 + SQLAlchemy 2.0 backend development specialists",
  "version": "1.0.0",
  "author": {
    "name": "Backend Team"
  },
  "agents": [
    {
      "name": "api-designer",
      "description": "REST API design and OpenAPI specification specialist"
    },
    {
      "name": "db-optimizer",
      "description": "Database query optimization and schema design expert"
    },
    {
      "name": "security-auditor",
      "description": "Security best practices and vulnerability analysis"
    },
    {
      "name": "performance-analyst",
      "description": "Application performance profiling and optimization"
    }
  ],
  "category": "backend",
  "tags": ["fastapi", "python", "sqlalchemy", "backend", "rest-api"],
  "repository": "https://github.com/example/backend-plugin"
}
```

**When to Use**:
- Domain-specific expertise
- Code review automation
- Design consultation
- Performance analysis

---

## Example 5: Full-Featured Plugin (Commands + Agents + Skills)

**Use Case**: Complete development environment with all components

**Directory Structure**:
```
fullstack-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   ├── init-stack.md
│   ├── scaffold-api.md
│   └── deploy-fullstack.md
├── agents/
│   ├── frontend-agent.md
│   ├── backend-agent.md
│   └── database-agent.md
├── skills/
│   ├── frontend-patterns/
│   │   └── SKILL.md
│   ├── backend-patterns/
│   │   └── SKILL.md
│   └── deployment-strategies.md
└── README.md
```

**plugin.json**:
```json
{
  "name": "fullstack-plugin",
  "description": "Complete full-stack development toolkit with Next.js 15 and FastAPI 0.120.2",
  "version": "2.0.0",
  "author": {
    "name": "FullStack Team",
    "email": "team@fullstack.dev",
    "url": "https://fullstack.dev"
  },
  "commands": [
    {
      "name": "init-stack",
      "description": "Initialize full-stack project with Next.js frontend and FastAPI backend"
    },
    {
      "name": "scaffold-api",
      "description": "Generate API endpoints with OpenAPI specs"
    },
    {
      "name": "deploy-fullstack",
      "description": "Deploy frontend to Vercel and backend to Railway"
    }
  ],
  "agents": [
    {
      "name": "frontend-agent",
      "description": "Next.js 15, React 19, and TypeScript specialist"
    },
    {
      "name": "backend-agent",
      "description": "FastAPI, SQLAlchemy, and Pydantic expert"
    },
    {
      "name": "database-agent",
      "description": "PostgreSQL schema design and migration specialist"
    }
  ],
  "skills": [
    "./skills/frontend-patterns/SKILL.md",
    "./skills/backend-patterns/SKILL.md",
    "./skills/deployment-strategies.md"
  ],
  "category": "fullstack",
  "tags": ["nextjs", "fastapi", "react", "typescript", "python", "fullstack"],
  "repository": "https://github.com/example/fullstack-plugin",
  "documentation": "https://fullstack-plugin.dev",
  "permissions": {
    "filesystem": ["read", "write"],
    "network": ["https"]
  }
}
```

**When to Use**:
- Complete development environments
- Multi-component projects
- End-to-end workflows
- Team standardization

---

## Example 6: MCP Integration Plugin (Playwright)

**Use Case**: External tool integration via Model Context Protocol

**Directory Structure**:
```
playwright-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   ├── playwright-setup.md
│   └── run-e2e-tests.md
├── skills/
│   └── playwright-patterns/
│       ├── SKILL.md
│       └── examples.md
├── .mcp.json
└── README.md
```

**plugin.json**:
```json
{
  "name": "playwright-plugin",
  "description": "E2E testing automation with Playwright MCP integration",
  "version": "1.0.0",
  "author": {
    "name": "QA Team"
  },
  "commands": [
    {
      "name": "playwright-setup",
      "description": "Initialize Playwright-MCP for E2E testing automation"
    },
    {
      "name": "run-e2e-tests",
      "description": "Execute Playwright tests with MCP server"
    }
  ],
  "skills": [
    "./skills/playwright-patterns/SKILL.md"
  ],
  "mcpServers": [
    {
      "name": "playwright-mcp",
      "type": "optional",
      "configPath": ".mcp.json"
    }
  ],
  "category": "testing",
  "tags": ["playwright", "e2e", "testing", "automation"],
  "documentation": "https://playwright.dev"
}
```

**.mcp.json** (MCP Server Configuration):
```json
{
  "mcpServers": {
    "playwright-mcp": {
      "command": "npx",
      "args": ["-y", "@executeautomation/playwright-mcp-server"]
    }
  }
}
```

**When to Use**:
- External tool integration
- Browser automation
- API testing
- Visual regression testing

---

## Example 7: UI/UX Plugin with Multiple MCP Servers

**Use Case**: Design automation with Figma MCP integration

**Directory Structure**:
```
uiux-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   ├── init-design-system.md
│   ├── sync-figma.md
│   └── generate-components.md
├── agents/
│   ├── design-strategist.md
│   ├── figma-specialist.md
│   ├── component-builder.md
│   └── accessibility-auditor.md
├── skills/
│   ├── design-figma-mcp/
│   │   └── SKILL.md
│   ├── design-tokens/
│   │   └── SKILL.md
│   └── shadcn-ui-patterns.md
├── .mcp-figma.json
└── README.md
```

**plugin.json**:
```json
{
  "name": "uiux-plugin",
  "description": "Design automation with Figma MCP, shadcn/ui, and design-to-code workflows",
  "version": "2.0.0",
  "author": {
    "name": "Design Team"
  },
  "commands": [
    {
      "name": "init-design-system",
      "description": "Initialize design system with shadcn/ui components"
    },
    {
      "name": "sync-figma",
      "description": "Synchronize Figma designs with code components"
    },
    {
      "name": "generate-components",
      "description": "Generate React components from Figma designs"
    }
  ],
  "agents": [
    {
      "name": "design-strategist",
      "description": "Design system architecture and strategy specialist"
    },
    {
      "name": "figma-specialist",
      "description": "Figma file analysis and component extraction expert"
    },
    {
      "name": "component-builder",
      "description": "React component generation from design specs"
    },
    {
      "name": "accessibility-auditor",
      "description": "WCAG compliance and accessibility testing"
    }
  ],
  "skills": [
    "./skills/design-figma-mcp/SKILL.md",
    "./skills/design-tokens/SKILL.md",
    "./skills/shadcn-ui-patterns.md"
  ],
  "mcpServers": [
    {
      "name": "figma-mcp",
      "type": "optional",
      "configPath": ".mcp-figma.json"
    }
  ],
  "category": "ui-ux",
  "tags": ["figma", "design", "ui-ux", "shadcn-ui", "accessibility"],
  "repository": "https://github.com/example/uiux-plugin",
  "documentation": "https://uiux-plugin.dev"
}
```

**When to Use**:
- Design system management
- Figma-to-code workflows
- Component generation
- Design token synchronization

---

## Example 8: DevOps Plugin with Dependencies

**Use Case**: Infrastructure automation with base plugin dependency

**Directory Structure**:
```
devops-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   ├── init-docker.md
│   ├── deploy-k8s.md
│   ├── setup-ci.md
│   └── monitor-infra.md
├── agents/
│   ├── docker-specialist.md
│   ├── kubernetes-expert.md
│   ├── ci-cd-architect.md
│   └── monitoring-agent.md
└── README.md
```

**plugin.json**:
```json
{
  "name": "devops-plugin",
  "description": "Infrastructure automation with Docker, Kubernetes, and GitHub Actions",
  "version": "1.2.0",
  "author": {
    "name": "DevOps Team"
  },
  "commands": [
    {
      "name": "init-docker",
      "description": "Initialize Docker configuration with multi-stage builds"
    },
    {
      "name": "deploy-k8s",
      "description": "Deploy to Kubernetes with Helm charts"
    },
    {
      "name": "setup-ci",
      "description": "Configure GitHub Actions CI/CD pipeline"
    },
    {
      "name": "monitor-infra",
      "description": "Setup Prometheus and Grafana monitoring"
    }
  ],
  "agents": [
    {
      "name": "docker-specialist",
      "description": "Docker containerization and optimization expert"
    },
    {
      "name": "kubernetes-expert",
      "description": "Kubernetes orchestration and management specialist"
    },
    {
      "name": "ci-cd-architect",
      "description": "CI/CD pipeline design and automation expert"
    },
    {
      "name": "monitoring-agent",
      "description": "Infrastructure monitoring and alerting specialist"
    }
  ],
  "category": "devops",
  "tags": ["docker", "kubernetes", "ci-cd", "github-actions", "devops"],
  "repository": "https://github.com/example/devops-plugin",
  "dependencies": ["base-plugin"],
  "permissions": {
    "filesystem": ["read", "write"],
    "network": ["https"],
    "process": ["spawn"]
  }
}
```

**When to Use**:
- Infrastructure as code
- Container orchestration
- CI/CD automation
- Monitoring setup

---

## Example 9: Technical Blog Plugin (Content Creation)

**Use Case**: Technical writing and documentation generation

**Directory Structure**:
```
technical-blog-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   └── generate-blog-post.md
├── agents/
│   ├── content-strategist.md
│   ├── technical-writer.md
│   ├── code-explainer.md
│   ├── seo-optimizer.md
│   ├── diagram-generator.md
│   ├── markdown-stylist.md
│   └── publishing-assistant.md
├── skills/
│   ├── nextra-i18n-patterns/
│   │   └── SKILL.md
│   ├── markdown-advanced/
│   │   └── SKILL.md
│   └── technical-writing-guide.md
└── README.md
```

**plugin.json**:
```json
{
  "name": "technical-blog-plugin",
  "description": "Technical content creation with Nextra, i18n, and SEO optimization",
  "version": "1.0.0",
  "author": {
    "name": "Content Team"
  },
  "commands": [
    {
      "name": "generate-blog-post",
      "description": "Generate technical blog post with code examples and diagrams"
    }
  ],
  "agents": [
    {
      "name": "content-strategist",
      "description": "Content planning and audience targeting specialist"
    },
    {
      "name": "technical-writer",
      "description": "Technical documentation and tutorial expert"
    },
    {
      "name": "code-explainer",
      "description": "Code example creation and explanation specialist"
    },
    {
      "name": "seo-optimizer",
      "description": "SEO metadata and search optimization expert"
    },
    {
      "name": "diagram-generator",
      "description": "Technical diagram and flowchart creation"
    },
    {
      "name": "markdown-stylist",
      "description": "Markdown formatting and structure expert"
    },
    {
      "name": "publishing-assistant",
      "description": "Multi-platform publishing workflow coordinator"
    }
  ],
  "skills": [
    "./skills/nextra-i18n-patterns/SKILL.md",
    "./skills/markdown-advanced/SKILL.md",
    "./skills/technical-writing-guide.md"
  ],
  "category": "documentation",
  "tags": ["blogging", "nextra", "markdown", "i18n", "technical-writing"]
}
```

**When to Use**:
- Technical blog creation
- Documentation generation
- Tutorial writing
- Multi-language content

---

## Example 10: Frontend Plugin (Complete Reference)

**Use Case**: Modern frontend development with Next.js 15, React 19, and Playwright MCP

**Directory Structure**:
```
frontend-plugin/
├── .claude-plugin/
│   └── plugin.json
├── commands/
│   ├── init-next.md
│   ├── biome-setup.md
│   └── playwright-setup.md
├── agents/
│   ├── nextjs-architect.md
│   ├── react-specialist.md
│   ├── state-manager.md
│   └── performance-optimizer.md
├── skills/
│   ├── framework-nextjs-advanced/
│   │   └── SKILL.md
│   ├── framework-react-19/
│   │   └── SKILL.md
│   ├── design-shadcn-ui/
│   │   └── SKILL.md
│   ├── testing-playwright-mcp/
│   │   └── SKILL.md
│   └── domain-frontend.md
├── .mcp.json
└── README.md
```

**plugin.json**:
```json
{
  "name": "frontend-plugin",
  "description": "Next.js 15 + React 19 + Biome + Playwright-MCP - Modern frontend development",
  "version": "1.0.0",
  "author": {
    "name": "Frontend Team"
  },
  "commands": [
    {
      "name": "init-next",
      "description": "Initialize Next.js 15 project with App Router"
    },
    {
      "name": "biome-setup",
      "description": "Setup Biome for linting and formatting"
    },
    {
      "name": "playwright-setup",
      "description": "Initialize Playwright-MCP for E2E testing automation"
    }
  ],
  "agents": [
    {
      "name": "nextjs-architect",
      "description": "Next.js 15 architecture and routing specialist"
    },
    {
      "name": "react-specialist",
      "description": "React 19 hooks and patterns expert"
    },
    {
      "name": "state-manager",
      "description": "State management with Zustand and TanStack Query"
    },
    {
      "name": "performance-optimizer",
      "description": "Bundle optimization and Core Web Vitals expert"
    }
  ],
  "skills": [
    "./skills/framework-nextjs-advanced/SKILL.md",
    "./skills/framework-react-19/SKILL.md",
    "./skills/design-shadcn-ui/SKILL.md",
    "./skills/testing-playwright-mcp/SKILL.md",
    "./skills/domain-frontend.md"
  ],
  "mcpServers": [
    {
      "name": "playwright-mcp",
      "type": "optional",
      "configPath": ".mcp.json"
    }
  ],
  "category": "frontend",
  "tags": ["nextjs", "react", "typescript", "biome", "playwright"],
  "repository": "https://github.com/example/frontend-plugin",
  "documentation": "https://frontend-plugin.dev"
}
```

**When to Use**:
- Modern frontend projects
- Next.js applications
- React 19 features
- E2E testing setup

---

## Validation Examples

### Valid plugin.json (All Features)
```json
{
  "name": "example-plugin",
  "description": "Example plugin demonstrating all features",
  "version": "1.0.0",
  "author": {
    "name": "Example Team",
    "email": "team@example.com",
    "url": "https://example.com"
  },
  "commands": [
    {
      "name": "example-command",
      "description": "Example command description"
    }
  ],
  "agents": [
    {
      "name": "example-agent",
      "description": "Example agent description"
    }
  ],
  "skills": [
    "./skills/example-skill/SKILL.md"
  ],
  "mcpServers": [
    {
      "name": "example-mcp",
      "type": "optional",
      "configPath": ".mcp.json"
    }
  ],
  "category": "example",
  "tags": ["example", "demo"],
  "repository": "https://github.com/example/example-plugin",
  "documentation": "https://docs.example.com",
  "permissions": {
    "filesystem": ["read"],
    "network": ["https"]
  },
  "dependencies": []
}
```

### Invalid plugin.json (Common Errors)
```json
{
  // ❌ Missing required field: description
  "name": "broken-plugin",
  "version": "1.0.0",
  
  // ❌ Author as string (should be object)
  "author": "John Doe",
  
  "commands": [
    {
      "name": "test-command",
      // ❌ Has invalid 'path' field
      "path": "commands/test.md",
      "description": "Test command"
    }
  ],
  
  "skills": [
    // ❌ Missing './' prefix
    "skills/my-skill.md"
  ]
}
```

---

## Best Practices Summary

1. **Start Simple**: Begin with minimal plugin (Example 1), add features incrementally
2. **Progressive Disclosure**: Example 1 → 2 → 3 → 4 → 5 (complexity increases)
3. **MCP Integration**: Use Example 6 or 7 for external tool integration
4. **Full-Featured**: Reference Example 5 or 10 for complete environments
5. **Validation**: Always validate with tools before deployment

---

**Quick Command Reference**:
```bash
# Validate plugin
/plugin validate ./my-plugin

# Install from marketplace
/plugin install my-plugin@marketplace-name

# List installed plugins
/plugin list

# Uninstall plugin
/plugin uninstall my-plugin
```
