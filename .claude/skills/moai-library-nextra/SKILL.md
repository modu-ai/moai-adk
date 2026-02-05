---
name: moai-library-nextra
description: >
  Enterprise Nextra documentation framework with Next.js. Use when building documentation
  sites, knowledge bases, or API reference documentation.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Write Edit Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.2.0"
  category: "library"
  modularized: "true"
  status: "active"
  updated: "2026-01-11"
  tags: "library, nextra, nextjs, documentation, mdx, static-site"
  aliases: "moai-library-nextra"

# MoAI Extension: Triggers
triggers:
  keywords: ["documentation", "nextra", "docs site", "knowledge base", "api reference", "mdx", "static site"]
---

## Quick Reference

Purpose: Build professional documentation sites with Nextra and Next.js.

Nextra Advantages:

- Zero config MDX with seamless Markdown and JSX integration
- File-system routing with automatic route generation
- Performance optimized with code splitting and prefetching
- Theme system with pluggable and customizable themes
- Built-in internationalization support

Core Files:

- The pages directory contains documentation pages in MDX format
- The theme.config.tsx file contains site configuration
- The _meta.js files control navigation structure

## Implementation Guide

### Features

This skill covers Nextra 3.x and 4.x documentation framework architecture patterns, Next.js 14 and 15 integration with optimal configuration, theme customization via theme.config.tsx or Layout props, advanced search with FlexSearch integration, internationalization support, MDX-powered content with React components, and App Router support in Nextra 4.x with Turbopack compatibility.

### When to Use

Use this skill when building documentation sites with modern React features, creating knowledge bases with advanced search capabilities, developing multi-language documentation portals, implementing custom documentation themes, or integrating interactive examples in technical docs.

### Project Setup

To initialize a Nextra documentation site, use the create-nextra-app command with npx specifying the docs template. The resulting project structure includes a pages directory containing the custom App component file, the index MDX file for the home page, and subdirectories for documentation sections. Each section contains MDX files for content and a _meta.json file for navigation configuration.

### Theme Configuration

The theme.config.tsx file exports a configuration object with several key properties. The logo property defines the site branding element. The project property contains a link to the project repository. The docsRepositoryBase property specifies the base URL for the edit link feature. The useNextSeoProps function returns SEO configuration including the title template.

Essential configuration options include branding settings for logo and logoLink, navigation settings for project links and repository base URLs, sidebar settings for default collapse level and toggle button visibility, table of contents settings including the back-to-top feature, and footer settings for custom footer text.

### Navigation Structure

The _meta.js files control sidebar menu ordering and display names. Each file exports a default object where keys represent file or directory names and values represent display labels. Special entries include separator lines using triple dashes as keys with empty string values, and external links can be configured with nested objects containing title, href, and newWindow properties.

### MDX Content and JSX Integration

Nextra supports mixing Markdown with React components directly in MDX files. Components can be imported at the top of the file and used inline with the Markdown content. Custom components can be defined and exported within the MDX file itself. The Callout component from nextra/components provides styled callout boxes for notes, warnings, and tips.

### Search and SEO Optimization

The theme configuration supports built-in search with customizable placeholder text. SEO metadata can be configured through the head property which accepts JSX for meta tags including Open Graph title, description, and image. The useNextSeoProps function provides dynamic title template configuration.

---

## Advanced Documentation

This skill uses Progressive Disclosure. For detailed patterns see the modules directory:

- modules/configuration.md provides complete theme.config reference
- modules/mdx-components.md covers the MDX component library
- modules/i18n-setup.md contains the internationalization guide
- modules/deployment.md covers hosting and deployment

---

## Theme Options

Built-in themes include nextra-theme-docs which is recommended for documentation sites, and nextra-theme-blog for blog implementations.

Customization options include CSS variables for colors, custom sidebar components, footer customization, and navigation layout modifications.

---

## Deployment

Popular deployment platforms include Vercel with zero-config recommended setup, GitHub Pages for free self-hosted options, Netlify for flexible CI/CD integration, and custom servers for full control.

For Vercel deployment, install the Vercel CLI globally using npm, then run the vercel command to select the project and deploy.

---

## Integration with Other Skills

Complementary skills include moai-docs-generation for auto-generating docs from code, moai-workflow-docs for validating documentation quality, and moai-cc-claude-md for Markdown formatting.

---

## Version History

Version 2.2.0 released 2026-01-11 removes code blocks to comply with CLAUDE.md Documentation Standards and converts all examples to narrative descriptions.

Version 2.1.0 released 2025-12-30 updated configuration.md with complete Nextra-specific theme.config.tsx patterns, added Nextra 4.x App Router configuration patterns, updated version compatibility for Next.js 14 and 15, and added Turbopack support documentation.

Version 2.0.0 released 2025-11-23 refactored with Progressive Disclosure, highlighted configuration patterns, and added MDX integration guide.

Version 1.0.0 released 2025-11-12 provided initial Nextra architecture guide, theme configuration, and i18n support.

---

Maintained by: MoAI-ADK Team
Domain: Documentation Architecture
Generated with: MoAI-ADK Skill Factory

---

## Works Well With

Agents:

- workflow-docs for documentation generation
- code-frontend for Nextra implementation
- workflow-spec for architecture documentation

Skills:

- moai-docs-generation for content generation
- moai-workflow-docs for documentation validation
- moai-library-mermaid for diagram integration

Commands:

- moai:3-sync for documentation deployment
- moai:0-project for Nextra project initialization
