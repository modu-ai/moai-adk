import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'
import pkg from '../../moai-adk-ts/package.json' with { type: 'json' }

const version = `v${pkg.version}`

// https://vitepress.dev/reference/site-config
export default withMermaid(
  defineConfig({
  title: 'MoAI-ADK',
  description: 'SPEC-First TDD Development Kit with Universal Language Support',

  // Base path (for GitHub Pages or subpath deployment)
  base: '/',

  // Language and locale
  lang: 'ko-KR',

  // Theme configuration
  themeConfig: {
    // Logo
    logo: '/logo.svg',

    // Site title in nav
    siteTitle: 'MoAI-ADK',

    // Navigation menu
    nav: [
      { text: 'Home', link: '/' },
      {
        text: 'Getting Started',
        items: [
          { text: 'Introduction', link: '/guides/getting-started' },
          { text: 'Installation', link: '/guides/installation' },
          { text: 'Quick Start', link: '/guides/quick-start' }
        ]
      },
      {
        text: 'Workflow',
        items: [
          { text: '0. Project Setup', link: '/guides/workflow/0-project' },
          { text: '1. SPEC Writing', link: '/guides/workflow/1-spec' },
          { text: '2. TDD Implementation', link: '/guides/workflow/2-build' },
          { text: '3. Document Sync', link: '/guides/workflow/3-sync' },
          { text: '9. Update & Upgrade', link: '/guides/workflow/9-update' }
        ]
      },
      {
        text: 'Core Concepts',
        items: [
          { text: 'SPEC-First TDD', link: '/guides/concepts/spec-first-tdd' },
          { text: 'EARS Requirements', link: '/guides/concepts/ears-guide' },
          { text: 'TAG System', link: '/guides/concepts/tag-system' },
          { text: 'TRUST Principles', link: '/guides/concepts/trust-principles' }
        ]
      },
      {
        text: 'Agents & Hooks',
        items: [
          { text: 'Alfred Agents', link: '/guides/agents/overview' },
          { text: 'Hooks System', link: '/guides/hooks/overview' }
        ]
      },
      {
        text: version,
        items: [
          { text: 'API Reference', link: '/api/index.html' },
          { text: 'Changelog', link: 'https://github.com/modu-ai/moai-adk/releases' },
          { text: 'GitHub', link: 'https://github.com/modu-ai/moai-adk' }
        ]
      }
    ],

    // Sidebar configuration
    sidebar: {
      '/guides/': [
        {
          text: 'Getting Started',
          collapsed: false,
          items: [
            { text: 'Introduction', link: '/guides/getting-started' },
            { text: 'Installation', link: '/guides/installation' },
            { text: 'Quick Start', link: '/guides/quick-start' }
          ]
        },
        {
          text: 'Core Concepts',
          collapsed: false,
          items: [
            { text: 'SPEC-First TDD', link: '/guides/concepts/spec-first-tdd' },
            { text: 'EARS Requirements', link: '/guides/concepts/ears-guide' },
            { text: 'TAG System', link: '/guides/concepts/tag-system' },
            { text: 'TRUST Principles', link: '/guides/concepts/trust-principles' }
          ]
        },
        {
          text: 'Workflow',
          collapsed: false,
          items: [
            { text: '0. Project Setup', link: '/guides/workflow/0-project' },
            { text: '1. SPEC Writing', link: '/guides/workflow/1-spec' },
            { text: '2. TDD Implementation', link: '/guides/workflow/2-build' },
            { text: '3. Document Sync', link: '/guides/workflow/3-sync' },
            { text: '9. Update & Upgrade', link: '/guides/workflow/9-update' }
          ]
        },
        {
          text: 'Alfred Agents',
          collapsed: true,
          items: [
            { text: 'Overview', link: '/guides/agents/overview' }
          ]
        },
        {
          text: 'Hooks',
          collapsed: true,
          items: [
            { text: 'Overview', link: '/guides/hooks/overview' }
          ]
        }
      ]
    },

    // Social links
    socialLinks: [
      { icon: 'github', link: 'https://github.com/modu-ai/moai-adk' }
    ],

    // Footer
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyleft © 2024 MoAI-Lab'
    },

    // Edit link
    editLink: {
      pattern: 'https://github.com/modu-ai/moai-adk/edit/main/docs/:path',
      text: 'Edit this page on GitHub'
    },

    // Last updated
    lastUpdated: {
      text: 'Updated at',
      formatOptions: {
        dateStyle: 'full',
        timeStyle: 'medium'
      }
    },

    // Search (Local search)
    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: '검색',
            buttonAriaLabel: '검색'
          },
          modal: {
            noResultsText: '결과를 찾을 수 없습니다',
            resetButtonTitle: '초기화',
            footer: {
              selectText: '선택',
              navigateText: '이동',
              closeText: '닫기'
            }
          }
        }
      }
    },

    // Outline (Table of Contents)
    outline: {
      level: [2, 3],
      label: 'On this page'
    }
  },

  // Markdown configuration
  markdown: {
    // Line numbers in code blocks
    lineNumbers: true,

    // Math support (optional)
    math: false,

    // Code theme
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    }
  },

  // Mermaid configuration
  mermaid: {
    // Mermaid theme
    theme: 'default'
  },

  // Mermaid plugin options
  mermaidPlugin: {
    class: 'mermaid'
  },

  // Head configuration (meta tags, scripts)
  head: [
    ['link', { rel: 'icon', href: '/logo.svg' }],
    ['meta', { name: 'theme-color', content: '#646cff' }],
    ['meta', { name: 'og:type', content: 'website' }],
    ['meta', { name: 'og:locale', content: 'ko_KR' }],
    ['meta', { name: 'og:site_name', content: 'MoAI-ADK' }],
    ['meta', { name: 'og:image', content: '/alfred_logo.png' }]
  ],

  // Build options
  ignoreDeadLinks: true,

  // Clean URLs (remove .html extension)
  cleanUrls: true,

  // Site map generation
  sitemap: {
    hostname: 'https://moai-adk.vercel.app'
  }
})
)
