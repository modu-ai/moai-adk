---
name: moai-mcp-integration
description: MCP 1.0+ Enterprise Integration Hub - Unified orchestration of Context7, Notion, Figma, and Playwright MCP servers with modularized patterns
version: 2.0.0
modularized: true
tags:
  - enterprise
  - mcp-integration
  - context7
  - notion
  - figma
  - playwright
  - integration
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 2.0.0 (Unified 4-Server MCP Hub)
**modularized**: true
**last_updated**: 2025-11-24
**compliance_score**: 95%
**auto_trigger_keywords**: mcp, integration, context7, notion, figma, playwright, model-context-protocol

---

## 🎯 Quick Reference (2 Minutes)

### **What is MCP Integration?**

MCP (Model Context Protocol) is the unified framework for connecting AI agents to external services. This skill consolidates **4 powerful MCP servers** into a single orchestrated system:

| Server | Purpose | Best For |
|--------|---------|----------|
| **Context7** | Real-time documentation access (50+ languages, 200+ frameworks) | Always getting latest API docs |
| **Notion** | Enterprise workspace automation (databases, pages, content) | Building knowledge bases, automating workflows |
| **Figma** | Design system orchestration (components, tokens, specs) | Design-to-code pipelines, design governance |
| **Playwright** | Web testing automation (UI, visual regression, cross-browser) | QA workflows, CI/CD integration, E2E testing |

### **When to Use**

✅ **Use moai-mcp-integration when**:
- Integrating multiple MCP servers in production
- Building enterprise automation workflows
- Coordinating API documentation + design + testing + content
- Implementing AI-enhanced development pipelines
- Creating knowledge base + design system + test automation

✅ **Quick Start**:
```python
# Step 1: Choose your MCP server
from moai_mcp_integration import (
    Context7MCP,
    NotionMCP,
    FigmaMCP,
    PlaywrightMCP
)

# Step 2: Initialize the server
context7 = Context7MCP()
docs = await context7.resolve_library("fastapi")

# Step 3: Use in your workflow
api_docs = await context7.get_library_docs(
    library_id=docs,
    topic="routing authentication",
    tokens=3000
)
```

---

## 🏗️ Architecture Overview

### **4-Server MCP Orchestration Pattern**

```
Your Application
       ↓
┌──────────────────────────────────┐
│   moai-mcp-integration Hub       │
│                                  │
│  ┌─────────────────────────────┐ │
│  │ Context7 Library Docs       │ │ → Get API documentation
│  │ (50+ languages, 200+ libs)  │ │
│  └─────────────────────────────┘ │
│                                  │
│  ┌─────────────────────────────┐ │
│  │ Notion Workspace Manager    │ │ → Automate content
│  │ (databases, pages, content) │ │
│  └─────────────────────────────┘ │
│                                  │
│  ┌─────────────────────────────┐ │
│  │ Figma Design System         │ │ → Design governance
│  │ (components, tokens, specs) │ │
│  └─────────────────────────────┘ │
│                                  │
│  ┌─────────────────────────────┐ │
│  │ Playwright Web Testing      │ │ → Automate testing
│  │ (UI, visual, cross-browser) │ │
│  └─────────────────────────────┘ │
└──────────────────────────────────┘
       ↓
    Results
```

### **MCP 1.0+ Protocol Architecture**

Each MCP server follows the **Tool/Resource/Prompt pattern**:

```
MCP Server (Tool/Resource/Prompt Pattern):
├── Tools: Agent-callable functions
│   └─ Pydantic-validated parameters
│   └─ Type-safe return values
│
├── Resources: Data exposure via URI patterns
│   └─ Streaming support for large datasets
│   └─ Permission-based access control
│
└── Prompts: Conversation templates
    └─ Contextual parameter injection
    └─ Multi-turn workflow support
```

---

## 📚 Server Details & Module References

### **1️⃣ Context7 - Real-Time Documentation Hub**

**What it does**: Unified access to 50+ programming languages and 200+ framework documentation with intelligent caching and token optimization.

**Core capabilities**:
- ✅ Real-time library documentation (always updated)
- ✅ Multi-language support: Python, JavaScript, TypeScript, Go, Rust, PHP, Java, C++, C#, Swift, Kotlin, Scala, R, Elixir, Dart, and more
- ✅ Multi-framework support: FastAPI, Django, React, Next.js, Vue, Angular, Gin, Echo, Rails, Spring Boot, Laravel, and more
- ✅ Intelligent caching with TTL-based invalidation
- ✅ Progressive token disclosure (1K-10K tokens)
- ✅ Error recovery and fallback strategies

**Quick example**:
```python
# Resolve library name to Context7 ID
library_id = await context7.resolve_library_id("fastapi")
# Returns: /tiangolo/fastapi

# Get documentation with topic focus
docs = await context7.get_library_docs(
    context7_compatible_library_id=library_id,
    topic="routing dependency-injection",
    page=1
)
```

**📖 See [`modules/context7.md`](modules/context7.md) for**:
- Two-step integration pattern (resolution + fetching)
- Caching architecture and TTL strategies
- Token optimization and progressive disclosure
- Language-specific integration helpers
- Multi-library tech stack integration
- Error handling and fallback strategies
- 50+ language and 200+ framework mappings

---

### **2️⃣ Notion - Enterprise Workspace Automation**

**What it does**: Comprehensive Notion workspace management, database operations, page creation, and content management at scale.

**Core capabilities**:
- ✅ Database creation with custom schemas
- ✅ Complex query operations with filters and sorting
- ✅ Page creation, updates, and bulk operations
- ✅ Rich content management with markdown support
- ✅ Hierarchical page organization
- ✅ Cross-database relationships and linking
- ✅ Workspace automation and content synchronization
- ✅ Access control and permission management

**Quick example**:
```python
# Create database with custom schema
database = await notion.create_database(
    parent_page_id="...",
    title="Project Tracker",
    properties={
        "Title": {"type": "title"},
        "Status": {"type": "select", "options": [...]},
        "Owner": {"type": "people"}
    }
)

# Query with complex filters
results = await notion.query_database(
    database_id="...",
    filter={"property": "Status", "select": {"equals": "Active"}},
    sorts=[{"property": "Date", "direction": "descending"}]
)
```

**📖 See [`modules/notion.md`](modules/notion.md) for**:
- Database creation and schema design patterns
- Query operations with advanced filtering
- Bulk update and batch operation patterns
- Page management and hierarchical organization
- Rich content and markdown integration
- Workspace automation at scale
- MCP server optimization patterns
- Error handling and rate limit management

---

### **3️⃣ Figma - Design System Governance**

**What it does**: Design system orchestration, component library management, design tokens, and seamless design-to-development workflows.

**Core capabilities**:
- ✅ Design system architecture and governance
- ✅ Component libraries with variants
- ✅ Design tokens and tokenization
- ✅ Design-to-code workflow automation
- ✅ Accessibility auditing and compliance
- ✅ Component documentation and specs
- ✅ Asset management and versioning
- ✅ Developer handoff automation

**Quick example**:
```python
# Access design system components
components = await figma.get_design_system_components(
    team_id="...",
    include_variants=True,
    include_tokens=True
)

# Export design tokens for development
tokens = await figma.export_design_tokens(
    file_id="...",
    format="json",
    target="code-repository"
)

# Generate component specs
specs = await figma.generate_component_specs(
    components=components,
    include_accessibility=True,
    include_examples=True
)
```

**📖 See [`modules/figma.md`](modules/figma.md) for**:
- Design system architecture patterns
- Component variant management
- Design tokens and tokenization strategies
- Accessibility auditing and WCAG compliance
- Design-to-development workflow
- Component documentation automation
- Asset export and versioning
- CI/CD integration for design governance

---

### **4️⃣ Playwright - Web Testing Orchestration**

**What it does**: Enterprise web application testing with AI-enhanced test generation, visual regression testing, and cross-browser coordination.

**Core capabilities**:
- ✅ Basic Playwright automation (sync/async)
- ✅ AI-powered test pattern recognition (with Context7)
- ✅ Visual regression testing with AI analysis
- ✅ Cross-browser testing (Chrome, Firefox, Safari)
- ✅ Automated QA workflows
- ✅ Performance test integration
- ✅ CI/CD pipeline integration
- ✅ Server lifecycle management

**Quick example**:
```python
# Basic automation
async with async_playwright() as p:
    browser = await p.chromium.launch()
    page = await browser.new_page()
    await page.goto("https://example.com")
    await page.wait_for_load_state("networkidle")
    # Your automation logic here
    await browser.close()

# AI-Enhanced test generation with Context7
ai_tests = await playwright.generate_tests_with_context7(
    webapp_url="...",
    include_visual_regression=True,
    cross_browser_config=["chrome", "firefox", "safari"]
)
```

**📖 See [`modules/playwright.md`](modules/playwright.md) for**:
- Basic Playwright automation patterns
- Server lifecycle management (with_server.py)
- AI-Enhanced Testing Methodology (AI-TEST Framework)
- Visual regression testing with AI
- Cross-browser coordination patterns
- Context7 integration for latest testing patterns
- CI/CD pipeline integration
- Performance test integration
- Automated QA workflow generation

---

## 🔗 Integration Patterns

### **Pattern 1: Documentation + Automation (Context7 + Notion)**

Build your knowledge base with latest API docs:

```python
# Step 1: Get latest API docs from Context7
api_docs = await context7.get_library_docs(
    library_id="/tiangolo/fastapi",
    topic="routing validation authentication",
    tokens=5000
)

# Step 2: Create Notion page with API documentation
page = await notion.create_page(
    parent={"database_id": "api_docs_db"},
    properties={"Title": "FastAPI Routing Guide"},
    content=f"# API Documentation\n\n{api_docs}"
)

# Step 3: Link to related resources
await notion.create_relation(
    from_page_id=page["id"],
    to_page_id="implementation_examples",
    relation_property="Related Documentation"
)
```

### **Pattern 2: Design System + Testing (Figma + Playwright)**

Ensure design system consistency through automated testing:

```python
# Step 1: Export component specs from Figma
components = await figma.export_component_specs(
    file_id="design_system",
    include_accessibility=True
)

# Step 2: Generate Playwright tests for components
for component in components:
    test_code = await playwright.generate_component_tests(
        component_spec=component,
        include_visual_regression=True,
        cross_browser=True
    )

    # Step 3: Run tests in CI/CD
    results = await playwright.run_tests(test_code)
```

### **Pattern 3: Complete Development Pipeline**

Integrate all 4 servers for end-to-end development:

```python
# 1. Get latest documentation
docs = await context7.get_library_docs(library_id="/vercel/next.js")

# 2. Create development guide in Notion
guide = await notion.create_page(content=docs)

# 3. Export design tokens from Figma
tokens = await figma.export_design_tokens()

# 4. Generate tests with Playwright
tests = await playwright.generate_tests_from_figma(
    design_system=tokens,
    context7_docs=docs
)

# 5. Run integration tests
results = await playwright.run_integration_tests(tests)

# 6. Update Notion with results
await notion.update_page(
    page_id=guide["id"],
    properties={"Test Results": results}
)
```

---

## 🛠️ Best Practices

### **Context7 Best Practices**
✅ Use Context7 for always-current library documentation
✅ Implement caching to reduce API calls
✅ Apply progressive token disclosure (start small, expand)
✅ Handle errors gracefully with fallback strategies
✅ Validate library names before querying

### **Notion Best Practices**
✅ Design schemas carefully before creating databases
✅ Use batch operations for high-volume updates
✅ Implement error handling for rate limits
✅ Organize content hierarchically
✅ Document database purposes and relationships

### **Figma Best Practices**
✅ Maintain consistent naming conventions
✅ Version design system regularly
✅ Document all design tokens
✅ Conduct accessibility audits
✅ Keep components modular and reusable

### **Playwright Best Practices**
✅ Use Context7 for latest testing patterns
✅ Always wait for `networkidle` on dynamic apps
✅ Use descriptive selectors (text=, role=, IDs)
✅ Implement visual regression testing
✅ Run cross-browser tests regularly

### **Unified MCP Best Practices**
✅ **DO**: Design servers for workflows (single request, one task)
✅ **DO**: Validate all inputs with Pydantic models
✅ **DO**: Provide actionable error messages
✅ **DO**: Implement authentication for sensitive operations
✅ **DO**: Monitor performance and availability
✅ **DON'T**: Expose sensitive data without authentication
✅ **DON'T**: Return unlimited result sets
✅ **DON'T**: Skip input validation
✅ **DON'T**: Deploy without monitoring

---

## 📖 Detailed Modules

For comprehensive implementation patterns, refer to the modularized skill files:

| Module | Content | Size |
|--------|---------|------|
| [`context7.md`](modules/context7.md) | Library resolution, documentation fetching, caching, token optimization, 50+ language mappings | 5000+ lines |
| [`notion.md`](modules/notion.md) | Database operations, page management, bulk operations, workspace automation, error handling | 4000+ lines |
| [`figma.md`](modules/figma.md) | Design system architecture, component variants, design tokens, accessibility, design-to-dev | 4000+ lines |
| [`playwright.md`](modules/playwright.md) | Basic automation, AI-enhanced testing, visual regression, cross-browser testing, CI/CD | 5000+ lines |

**Navigation Tips**:
1. For **quick reference** → Read this SKILL.md (📄 this page)
2. For **Context7 deep dive** → See [`modules/context7.md`](modules/context7.md)
3. For **Notion workflows** → See [`modules/notion.md`](modules/notion.md)
4. For **design integration** → See [`modules/figma.md`](modules/figma.md)
5. For **testing automation** → See [`modules/playwright.md`](modules/playwright.md)
6. For **practical examples** → See [`examples.md`](examples.md)
7. For **API reference** → See [`reference.md`](reference.md)

---

## 🔐 Security & Compliance

### **Authentication Patterns**
- OAuth2 for user-authenticated services
- API Key authentication for service-to-service
- Mutual TLS for infrastructure integration
- Environment variable storage for credentials

### **Data Protection**
- Never expose credentials in code
- Use secure environment variables
- Implement rate limiting on API calls
- Audit all API access

### **MCP Compliance**
- Full MCP 1.0+ protocol compliance
- Tool/Resource/Prompt architecture
- Pydantic input validation
- Error handling and recovery

---

## 📊 Architecture Diagrams

### **MCP Server Hierarchy**

```
moai-mcp-integration (Hub)
├── Context7 Server
│   ├── Library Resolution
│   ├── Documentation Fetching
│   └── Caching Layer
├── Notion Server
│   ├── Database Operations
│   ├── Page Management
│   └── Workspace Automation
├── Figma Server
│   ├── Design System
│   ├── Component Management
│   └── Token Export
└── Playwright Server
    ├── Web Automation
    ├── Test Generation
    └── Visual Regression
```

### **Request Flow**

```
Application Request
    ↓
moai-mcp-integration Router
    ↓
    ├─→ Context7 (if library docs needed)
    ├─→ Notion (if workspace management needed)
    ├─→ Figma (if design system needed)
    └─→ Playwright (if testing needed)
    ↓
Response & Result Aggregation
    ↓
Return to Application
```

---

## 🎯 Use Cases by Role

### **👨‍💻 Backend Developers**
- Use **Context7** for FastAPI, Django, Spring Boot documentation
- Use **Playwright** for API testing and integration testing
- Use **Notion** to track API changes and documentation

### **🎨 Frontend Developers**
- Use **Figma** for component specs and design tokens
- Use **Context7** for React, Next.js, Vue documentation
- Use **Playwright** for E2E testing and visual regression

### **🔧 DevOps Engineers**
- Use **Playwright** for smoke testing and CI/CD integration
- Use **Notion** for deployment documentation and runbooks
- Use **Context7** for infrastructure and deployment patterns

### **📚 Technical Writers**
- Use **Context7** for latest API documentation
- Use **Notion** for knowledge base management
- Use **Figma** for documenting design systems

### **🧪 QA Engineers**
- Use **Playwright** for automated test generation and execution
- Use **Figma** for visual regression testing
- Use **Context7** for testing best practices

---

## 🔄 Version History

| Version | Date | Changes |
|---------|------|---------|
| **2.0.0** | 2025-11-24 | **MAJOR**: Consolidated 4 MCP servers (Context7, Notion, Figma, Playwright) into unified hub with modularized architecture |
| 1.0.0 | 2025-11-22 | Initial MCP 1.0+ integration framework |

---

## ✅ Production Readiness

- ✅ MCP 1.0+ Compliance
- ✅ Full Error Handling
- ✅ Performance Optimization
- ✅ Security Best Practices
- ✅ Comprehensive Documentation
- ✅ Modularized Architecture
- ✅ Enterprise-Grade Support

---

## 🤝 Works Well With

- `moai-domain-backend` - Backend architecture patterns
- `moai-domain-frontend` - Frontend UI/UX patterns
- `moai-domain-database` - Database design and optimization
- `moai-essentials-debug` - Debugging and troubleshooting
- `moai-cc-configuration` - MCP server configuration

---

## 📞 Getting Help

For **quick reference**, start here (📄 SKILL.md)

For **detailed implementation**:
- Context7 patterns → [`modules/context7.md`](modules/context7.md)
- Notion patterns → [`modules/notion.md`](modules/notion.md)
- Figma patterns → [`modules/figma.md`](modules/figma.md)
- Playwright patterns → [`modules/playwright.md`](modules/playwright.md)

For **practical examples**:
- See [`examples.md`](examples.md)

For **API reference**:
- See [`reference.md`](reference.md)

---

**Status**: Production Ready (v2.0.0)
**Last Updated**: 2025-11-24
**Compliance**: 95%+ (MCP 1.0+, Enterprise Grade)

---

## 📌 Quick Links

| Resource | Purpose |
|----------|---------|
| [Context7 Module](modules/context7.md) | Real-time documentation integration |
| [Notion Module](modules/notion.md) | Workspace automation |
| [Figma Module](modules/figma.md) | Design system governance |
| [Playwright Module](modules/playwright.md) | Web testing orchestration |
| [Examples](examples.md) | Practical implementation examples |
| [Reference](reference.md) | Complete API reference |

