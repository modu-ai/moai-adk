# Multi-Format Documentation Output

## Overview

Generate documentation in multiple formats using established tools: static site generators for HTML, WeasyPrint or Pandoc for PDF, and Markdown for portable content.

## Static Site Generators

### Nextra (Next.js)

Reference Skill("moai-library-nextra") for comprehensive patterns.

Key features:
- MDX support for React components in Markdown
- File-system based routing
- Built-in search with FlexSearch
- Theme customization
- Internationalization support

Deployment options:
- Vercel (recommended for Next.js)
- Netlify
- GitHub Pages with static export

### Docusaurus (React)

Initialize with npx create-docusaurus command.

Configure docusaurus.config.js:
- Set site metadata (title, tagline, url)
- Configure presets for docs and blog
- Add theme configuration for navbar, footer
- Enable plugins for search, analytics

Organize content:
- Place docs in docs directory
- Use category.json for sidebar configuration
- Add versioning for multiple doc versions
- Configure blog in blog directory

### VitePress (Vue)

Initialize with npm init vitepress.

Configure in .vitepress/config.js:
- Set site title and description
- Define navigation and sidebar
- Configure search (built-in or Algolia)
- Add social links

Features:
- Vue components in Markdown
- Automatic code syntax highlighting
- Dark mode support
- Fast hot module replacement

### MkDocs with Material

Install mkdocs-material theme.

Configure mkdocs.yml:
- Set theme to material
- Configure color palette and fonts
- Enable navigation features (tabs, sections)
- Add plugins for search, social cards

Deploy with mkdocs gh-deploy for GitHub Pages.

## PDF Generation

### Sphinx PDF Output

Use sphinx-build with latex builder:
- Requires LaTeX installation (texlive)
- Configure latex_documents in conf.py
- Customize styling with latex_elements

Alternative with rinohtype:
- Pure Python PDF generation
- No LaTeX dependency
- Configure with rinoh.ini

### MkDocs PDF Export

Install mkdocs-pdf-export-plugin.

Configure in mkdocs.yml:
- Set combined output option
- Configure styling
- Exclude specific pages if needed

Alternative with mkdocs-with-pdf plugin:
- Better styling options
- Cover page support
- Table of contents generation

### Pandoc Conversion

Convert Markdown to PDF with Pandoc:
- Requires Pandoc and LaTeX installation
- Use templates for consistent styling
- Configure with YAML metadata

Supports multiple input formats:
- Markdown, reStructuredText
- HTML, EPUB
- Word documents

### WeasyPrint for HTML to PDF

Convert HTML documentation to PDF:
- Renders CSS for print media
- Supports modern CSS features
- Generates bookmarks from headings

Configuration options:
- Custom stylesheets
- Page size and margins
- Headers and footers

## Markdown Output

### GitHub Flavored Markdown

Standard for repository documentation:
- README.md for project overview
- CONTRIBUTING.md for contribution guidelines
- CHANGELOG.md for version history
- docs/ directory for extended documentation

Features:
- Tables, task lists, strikethrough
- Syntax highlighted code blocks
- Automatic linking for URLs
- Emoji shortcodes

### MDX for Component Documentation

Combine Markdown with JSX:
- Import and use React components
- Interactive examples
- Code playgrounds

Supported by:
- Nextra, Docusaurus
- Storybook for component docs
- Custom MDX processors

## Hosting and Deployment

### GitHub Pages

Deploy static sites:
- Configure in repository settings
- Use GitHub Actions for automation
- Support custom domains

### Netlify

Features:
- Automatic builds from Git
- Preview deployments for PRs
- Form handling, serverless functions
- Split testing

### Vercel

Optimized for Next.js (Nextra):
- Zero configuration deployment
- Edge functions
- Analytics
- Incremental static regeneration

### Read the Docs

Specialized for Sphinx:
- Automatic versioning
- PDF generation
- Search integration
- Pull request previews

---

Version: 2.0.0
Last Updated: 2025-12-30
