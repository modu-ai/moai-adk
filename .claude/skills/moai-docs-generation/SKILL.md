---
name: moai-docs-generation
description: >
  Documentation generation patterns for technical specs, API docs, user guides,
  and knowledge bases using real tools like Sphinx, MkDocs, TypeDoc, and Nextra.
  Use when creating docs from code, building doc sites, or automating
  documentation workflows.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob Bash(npm:*) Bash(npx:*) Bash(git:*) Bash(sphinx-build:*) Bash(mkdocs:*) Bash(typedoc:*) mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-01-08"
  modularized: "true"
  tags: "workflow, documentation, sphinx, mkdocs, typedoc, api-docs, static-sites"
  context: "fork"
  agent: "general-purpose"
---

# Documentation Generation Patterns

## Quick Reference (30 seconds)

Purpose: Generate professional documentation using established tools and frameworks.

Core Documentation Tools:
- Python: Sphinx with autodoc, MkDocs with Material theme, pydoc
- TypeScript/JavaScript: TypeDoc, JSDoc, TSDoc
- API Documentation: OpenAPI/Swagger from FastAPI/Express, Redoc, Stoplight
- Static Sites: Nextra (Next.js), Docusaurus (React), VitePress (Vue)
- Universal: Markdown, MDX, reStructuredText

When to Use This Skill:
- Generating API documentation from code annotations
- Building documentation sites with search and navigation
- Creating user guides and technical specifications
- Automating documentation updates in CI/CD pipelines
- Converting between documentation formats

---

## Implementation Guide (5 minutes)

### Python Documentation with Sphinx

Sphinx Setup and Configuration:

Install Sphinx and extensions with pip install sphinx sphinx-autodoc-typehints sphinx-rtd-theme myst-parser

Initialize a Sphinx project by running sphinx-quickstart docs which creates the basic structure.

Configure conf.py with the following key settings:
- Set extensions to include autodoc, napoleon, typehints, and myst_parser
- Configure html_theme to sphinx_rtd_theme for a professional look
- Add autodoc_typehints set to description for inline type hints

Generate API documentation by running sphinx-apidoc with the source directory, outputting to docs/api, then run make html in the docs directory.

### Python Documentation with MkDocs

MkDocs Material Setup:

Install with pip install mkdocs mkdocs-material mkdocstrings mkdocstrings-python

Create mkdocs.yml configuration:
- Set site_name and site_url
- Configure theme with name material and desired color palette
- Add plugins including search and mkdocstrings
- Define nav structure with sections and pages

Use mkdocstrings syntax in Markdown files with ::: module.path to auto-generate API docs from docstrings.

Serve locally with mkdocs serve, build with mkdocs build, deploy with mkdocs gh-deploy.

### TypeScript Documentation with TypeDoc

TypeDoc Setup:

Install with npm install typedoc --save-dev

Add to package.json scripts: typedoc --out docs/api src/index.ts

Configure with typedoc.json:
- Set entryPoints to source files
- Configure out to docs/api
- Enable includeVersion and categorizeByGroup
- Set theme to default or install custom themes

Generate documentation by running npm run docs:generate

### JavaScript Documentation with JSDoc

JSDoc Setup:

Install with npm install jsdoc --save-dev

Create jsdoc.json configuration:
- Set source include paths and includePattern
- Configure templates and output destination
- Enable markdown plugin for rich formatting

Document functions with JSDoc comments using tags:
- @param for parameters with type and description
- @returns for return value documentation
- @example for usage examples
- @throws for error documentation

### OpenAPI/Swagger Documentation

FastAPI Auto-Documentation:

FastAPI provides automatic OpenAPI docs. Access Swagger UI at /docs and ReDoc at /redoc.

Enhance documentation by:
- Adding docstrings to route handlers
- Using response_model for typed responses
- Defining examples in Pydantic model Config class
- Setting tags for endpoint grouping
- Adding detailed descriptions in route decorators

Export OpenAPI spec programmatically with app.openapi() and save to openapi.json.

Express with Swagger:

Install swagger-jsdoc and swagger-ui-express.

Configure swagger-jsdoc with OpenAPI definition and API file paths.

Add @openapi comments to route handlers documenting paths, parameters, and responses.

Serve Swagger UI at /api-docs endpoint.

### Static Documentation Sites

Nextra (Next.js):

Reference Skill("moai-library-nextra") for comprehensive Nextra patterns.

Key advantages: MDX support, file-system routing, built-in search, theme customization.

Create with npx create-nextra-app, configure theme.config.tsx, organize pages in pages directory.

Docusaurus (React):

Initialize with npx create-docusaurus@latest my-docs classic

Configure in docusaurus.config.js:
- Set siteMetadata with title, tagline, url
- Configure presets with docs and blog settings
- Add themeConfig for navbar and footer
- Enable search with algolia plugin

Organize documentation in docs folder with category.json files for sidebar structure.

VitePress (Vue):

Initialize with npm init vitepress

Configure in .vitepress/config.js:
- Set title, description, base path
- Define themeConfig with nav and sidebar
- Configure search and social links

Use Markdown with Vue components, code highlighting, and frontmatter.

---

## Advanced Patterns (10+ minutes)

### Documentation from SPEC Files

Pattern for generating documentation from MoAI SPEC files:

Read SPEC file content and extract key sections: id, title, description, requirements, api_endpoints.

Generate structured Markdown documentation:
- Create overview section from description
- List requirements as feature bullets
- Document each API endpoint with method, path, and description
- Add usage examples based on endpoint definitions

Save generated docs to appropriate location in docs directory.

### CI/CD Documentation Pipeline

GitHub Actions Workflow:

Create .github/workflows/docs.yml that triggers on push to main branch when src or docs paths change.

Workflow steps:
- Checkout repository
- Setup language runtime (Python, Node.js)
- Install documentation dependencies
- Generate documentation using appropriate tool
- Deploy to GitHub Pages, Netlify, or Vercel

Example for Python/Sphinx:
- Install with pip install sphinx sphinx-rtd-theme
- Generate with sphinx-build -b html docs/source docs/build
- Deploy using actions-gh-pages action

Example for TypeScript/TypeDoc:
- Install with npm ci
- Generate with npm run docs:generate
- Deploy to Pages

### Documentation Validation

Link Checking:

Use linkchecker for local link validation in HTML output.

For Markdown, use markdown-link-check in pre-commit hooks.

Spell Checking:

Use pyspelling with Aspell for automated spell checking.

Configure .pyspelling.yml with matrix entries for different file types.

Documentation Coverage:

For Python, use interrogate to check docstring coverage.

Configure minimum coverage thresholds in pyproject.toml.

Fail CI builds if coverage drops below threshold.

### Multi-Language Documentation

Internationalization with Nextra:

Configure i18n in next.config.js with locales array and defaultLocale.

Create locale-specific pages in pages/[locale] directory.

Use next-intl or similar for translations.

Internationalization with Docusaurus:

Configure i18n in docusaurus.config.js with defaultLocale and locales.

Use docusaurus write-translations to generate translation files.

Organize translations in i18n/[locale] directory structure.

---

## Works Well With

Skills:
- moai-library-nextra - Comprehensive Nextra documentation framework patterns
- moai-lang-python - Python docstring conventions and typing
- moai-lang-typescript - TypeScript/JSDoc documentation patterns
- moai-domain-backend - API documentation for backend services
- moai-workflow-project - Project documentation integration

Agents:
- manager-docs - Documentation workflow orchestration
- expert-backend - API endpoint documentation
- expert-frontend - Component documentation

Commands:
- /moai:3-sync - Documentation synchronization with code changes

---

## Tool Reference

Python Documentation:
- Sphinx: https://www.sphinx-doc.org/
- MkDocs: https://www.mkdocs.org/
- MkDocs Material: https://squidfunk.github.io/mkdocs-material/
- mkdocstrings: https://mkdocstrings.github.io/

JavaScript/TypeScript Documentation:
- TypeDoc: https://typedoc.org/
- JSDoc: https://jsdoc.app/
- TSDoc: https://tsdoc.org/

API Documentation:
- OpenAPI Specification: https://spec.openapis.org/
- Swagger UI: https://swagger.io/tools/swagger-ui/
- Redoc: https://redocly.github.io/redoc/
- Stoplight: https://stoplight.io/

Static Site Generators:
- Nextra: https://nextra.site/
- Docusaurus: https://docusaurus.io/
- VitePress: https://vitepress.dev/

Style Guides:
- Google Developer Documentation Style Guide: https://developers.google.com/style
- Microsoft Writing Style Guide: https://learn.microsoft.com/style-guide/

---

Version: 2.0.0
Last Updated: 2025-12-30
