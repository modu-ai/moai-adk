# Documentation Generation Reference

## Tool Reference

### Python Documentation Tools

Sphinx:
- Official site: https://www.sphinx-doc.org/
- Configuration: conf.py in docs directory
- Extensions: autodoc, napoleon, viewcode, intersphinx
- Themes: sphinx_rtd_theme, furo, pydata-sphinx-theme
- Output formats: HTML, PDF (via LaTeX), ePub, man pages

MkDocs:
- Official site: https://www.mkdocs.org/
- Configuration: mkdocs.yml in project root
- Themes: material (recommended), readthedocs
- Plugins: search, mkdocstrings, macros, git-revision-date
- Output formats: HTML static site

mkdocstrings:
- Official site: https://mkdocstrings.github.io/
- Handlers: python, crystal, vba
- Features: Cross-references, signature rendering, source linking
- Styles: google, numpy, sphinx

interrogate:
- Purpose: Docstring coverage analysis
- Configuration: pyproject.toml or .interrogate.yaml
- Output: Coverage report, badge generation
- CI integration: Fail builds below threshold

### JavaScript/TypeScript Documentation Tools

TypeDoc:
- Official site: https://typedoc.org/
- Configuration: typedoc.json or CLI options
- Themes: default, or custom themes
- Plugins: markdown, missing-exports, mermaid
- Output formats: HTML, JSON, Markdown

JSDoc:
- Official site: https://jsdoc.app/
- Configuration: jsdoc.json or conf.json
- Templates: default, docdash, minami, better-docs
- Plugins: markdown, summarize
- Output format: HTML

TSDoc:
- Official site: https://tsdoc.org/
- Standard: Unified TypeScript doc comment format
- Tooling: eslint-plugin-tsdoc, api-extractor
- Purpose: Standardize TypeScript documentation

### API Documentation Tools

OpenAPI/Swagger:
- Specification: https://spec.openapis.org/
- Versions: 3.0.x, 3.1.x (latest)
- Editors: Swagger Editor, Stoplight Studio
- Validators: spectral, swagger-cli
- Code generators: openapi-generator, swagger-codegen

Swagger UI:
- Official site: https://swagger.io/tools/swagger-ui/
- Features: Interactive try-it-out, OAuth flows
- Hosting: Standalone, embedded, SwaggerHub
- Customization: CSS, plugins

Redoc:
- Official site: https://redocly.github.io/redoc/
- Features: Three-panel layout, search, code samples
- Hosting: Static HTML, React component
- CLI: redocly bundle, redocly preview

Stoplight:
- Official site: https://stoplight.io/
- Features: Design-first API development
- Tools: Studio (editor), Elements (docs), Prism (mock)
- Hosting: Cloud or self-hosted

### Static Site Generators

Nextra:
- Official site: https://nextra.site/
- Framework: Next.js
- Features: MDX, file routing, search, i18n
- Themes: docs, blog
- Deployment: Vercel (optimized)

Docusaurus:
- Official site: https://docusaurus.io/
- Framework: React
- Features: Versioning, i18n, search, blog
- Plugins: Search (Algolia, local), PWA, ideal-image
- Deployment: Any static host

VitePress:
- Official site: https://vitepress.dev/
- Framework: Vue 3
- Features: Fast HMR, Vue in Markdown, theming
- Built-in: Search, dark mode, carbon ads
- Deployment: Any static host

MkDocs Material:
- Official site: https://squidfunk.github.io/mkdocs-material/
- Framework: MkDocs
- Features: Search, dark mode, navigation tabs, social cards
- Plugins: Blog, tags, privacy
- Deployment: GitHub Pages, any static host

### PDF Generation Tools

WeasyPrint:
- Official site: https://weasyprint.org/
- Input: HTML + CSS
- Features: Print CSS, bookmarks, PDF/A
- Dependencies: System libraries (cairo, pango)

Pandoc:
- Official site: https://pandoc.org/
- Input formats: Markdown, HTML, RST, many more
- Output formats: PDF (via LaTeX), DOCX, EPUB
- Features: Templates, filters, citations

rinohtype:
- Purpose: Sphinx PDF without LaTeX
- Features: Pure Python, customizable styles
- Configuration: rinoh.ini

## Configuration Examples

### Sphinx conf.py Essential Settings

Project information:
- project: Project name string
- author: Author name string
- version: Short version string
- release: Full version string

Extensions:
- sphinx.ext.autodoc: Auto-generate from docstrings
- sphinx.ext.napoleon: Google/NumPy docstring support
- sphinx.ext.viewcode: Add source links
- sphinx.ext.intersphinx: Link to other projects

Theme configuration:
- html_theme: Theme name
- html_theme_options: Theme-specific options
- html_static_path: Static files directory

### MkDocs mkdocs.yml Essential Settings

Site information:
- site_name: Displayed site title
- site_url: Canonical URL
- repo_url: Source repository link

Theme configuration:
- theme.name: Theme name (material recommended)
- theme.palette: Color scheme
- theme.features: Enable features like tabs, search

Plugins:
- search: Built-in search
- mkdocstrings: API documentation

Navigation:
- nav: List of sections and pages

### TypeDoc typedoc.json Essential Settings

Input:
- entryPoints: Source files or directories
- entryPointStrategy: expand, packages, merge

Output:
- out: Output directory path
- name: Project name in docs

Behavior:
- excludePrivate: Hide private members
- excludeProtected: Hide protected members
- includeVersion: Show version in docs

## Hosting Options

GitHub Pages:
- URL pattern: username.github.io/repo
- Configuration: Repository settings or gh-pages branch
- Automation: GitHub Actions workflow
- Custom domains: CNAME file in root

Netlify:
- Features: Auto-deploy, previews, forms
- Configuration: netlify.toml or UI
- Build: Automatic from Git
- Custom domains: DNS configuration

Vercel:
- Features: Optimized for Next.js
- Configuration: vercel.json or UI
- Build: Automatic from Git
- Analytics: Built-in web analytics

Read the Docs:
- Features: Versioning, PDF, search
- Configuration: .readthedocs.yaml
- Build: Automatic from Git
- Hosting: Subdomain or custom domain

## Style Guides

Google Developer Documentation Style Guide:
- URL: https://developers.google.com/style
- Key principles: Clear, concise, accessible
- Voice: Active, second person
- Formatting: Sentence case headings

Microsoft Writing Style Guide:
- URL: https://learn.microsoft.com/style-guide/
- Key principles: Simple, direct, inclusive
- Voice: Conversational, friendly
- Formatting: Sentence case, bias-free language

Write the Docs:
- URL: https://www.writethedocs.org/
- Resources: Guides, conferences, community
- Topics: Documentation strategies, tools, careers

---

Version: 2.0.0
Last Updated: 2025-12-30
