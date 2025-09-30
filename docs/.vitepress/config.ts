import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

export default withMermaid(
  defineConfig({
    title: 'MoAI-ADK',
    description: 'TypeScript 기반 SPEC 우선 TDD 개발 도구',
    lang: 'ko-KR',
    ignoreDeadLinks: true,

    // Mermaid configuration
    mermaid: {
      theme: 'default',
      themeVariables: {
        primaryColor: '#5f67ee',
        primaryTextColor: '#fff',
        primaryBorderColor: '#5f67ee',
        lineColor: '#5f67ee',
        secondaryColor: '#339af0',
        tertiaryColor: '#51cf66'
      }
    },

    // Mermaid 플러그인 옵션
    mermaidPlugin: {
      class: "mermaid my-class", // Mermaid 다이어그램에 적용할 CSS 클래스
    },

    // Markdown configuration
    markdown: {
      config: (md) => {
        // Mermaid plugin will handle the rendering
      }
    },

  themeConfig: {
    logo: {
      light: '/moai-logo-light.png',
      dark: '/moai-logo-dark.png'
    },

    nav: [
      { text: '홈', link: '/' },
      {
        text: '소개',
        items: [
          { text: 'MoAI-ADK란?', link: '/introduction' },
          { text: '주요 특징', link: '/features' }
        ]
      },
      {
        text: '시작하기',
        items: [
          { text: '설치', link: '/getting-started/installation' },
          { text: '빠른 시작', link: '/getting-started/quick-start' },
          { text: '프로젝트 설정', link: '/getting-started/project-setup' }
        ]
      },
      {
        text: '핵심 개념',
        items: [
          { text: '3단계 워크플로우', link: '/guide/workflow' },
          { text: 'SPEC-First TDD', link: '/guide/spec-first-tdd' },
          { text: 'TAG 시스템', link: '/guide/tag-system' },
          { text: 'TRUST 5원칙', link: '/concepts/trust-principles' }
        ]
      },
      {
        text: 'CLI',
        items: [
          { text: 'moai init', link: '/cli/init' },
          { text: 'moai doctor', link: '/cli/doctor' },
          { text: 'moai status', link: '/cli/status' },
          { text: 'moai update', link: '/cli/update' },
          { text: 'moai restore', link: '/cli/restore' }
        ]
      },
      {
        text: 'Claude Code',
        items: [
          { text: '에이전트', link: '/claude/agents' },
          { text: '명령어', link: '/claude/commands' },
          { text: '훅', link: '/claude/hooks' }
        ]
      },
      {
        text: '레퍼런스',
        items: [
          { text: '설정 파일', link: '/reference/configuration' },
          { text: 'CLI 치트시트', link: '/reference/cli-cheatsheet' },
          { text: 'API 레퍼런스', link: '/reference/api-reference' },
          { text: 'Core 모듈', link: '/reference/core-modules' }
        ]
      },
      {
        text: '고급',
        items: [
          { text: 'doctor 고급 진단', link: '/advanced/doctor-advanced' },
          { text: '템플릿 커스터마이징', link: '/advanced/template-customization' }
        ]
      },
      {
        text: '도움말',
        items: [
          { text: 'FAQ', link: '/help/faq' },
          { text: 'GitHub', link: 'https://github.com/modu-ai/moai-adk' },
          { text: 'NPM', link: 'https://www.npmjs.com/package/moai-adk' }
        ]
      }
    ],

    sidebar: {
      '/introduction': [
        {
          text: '소개',
          items: [
            { text: 'MoAI-ADK란?', link: '/introduction' },
            { text: '주요 특징', link: '/features' }
          ]
        }
      ],
      '/features': [
        {
          text: '소개',
          items: [
            { text: 'MoAI-ADK란?', link: '/introduction' },
            { text: '주요 특징', link: '/features' }
          ]
        }
      ],
      '/getting-started/': [
        {
          text: '시작하기',
          items: [
            { text: '설치', link: '/getting-started/installation' },
            { text: '빠른 시작', link: '/getting-started/quick-start' },
            { text: '프로젝트 설정', link: '/getting-started/project-setup' }
          ]
        }
      ],
      '/guide/': [
        {
          text: '핵심 개념',
          items: [
            { text: '3단계 워크플로우', link: '/guide/workflow' },
            { text: 'SPEC-First TDD', link: '/guide/spec-first-tdd' },
            { text: 'TAG 시스템', link: '/guide/tag-system' }
          ]
        }
      ],
      '/concepts/': [
        {
          text: '핵심 개념',
          items: [
            { text: '3단계 워크플로우', link: '/guide/workflow' },
            { text: 'SPEC-First TDD', link: '/guide/spec-first-tdd' },
            { text: 'TAG 시스템', link: '/guide/tag-system' },
            { text: 'TRUST 5원칙', link: '/concepts/trust-principles' }
          ]
        }
      ],
      '/cli/': [
        {
          text: 'CLI 명령어',
          items: [
            { text: 'moai init', link: '/cli/init' },
            { text: 'moai doctor', link: '/cli/doctor' },
            { text: 'moai status', link: '/cli/status' },
            { text: 'moai update', link: '/cli/update' },
            { text: 'moai restore', link: '/cli/restore' }
          ]
        }
      ],
      '/advanced/': [
        {
          text: '고급 가이드',
          items: [
            { text: 'doctor 고급 진단', link: '/advanced/doctor-advanced' },
            { text: '템플릿 커스터마이징', link: '/advanced/template-customization' }
          ]
        }
      ],
      '/claude/': [
        {
          text: 'Claude Code 통합',
          items: [
            { text: '에이전트 개요', link: '/claude/agents' },
            { text: '워크플로우 명령어', link: '/claude/commands' },
            { text: '이벤트 훅', link: '/claude/hooks' },
            { text: '훅 시스템 상세', link: '/claude/hooks-detailed' }
          ]
        },
        {
          text: '에이전트 가이드',
          items: [
            { text: 'spec-builder', link: '/claude/agents/spec-builder' },
            { text: 'code-builder', link: '/claude/agents/code-builder' },
            { text: 'doc-syncer', link: '/claude/agents/doc-syncer' },
            { text: 'git-manager', link: '/claude/agents/git-manager' },
            { text: 'debug-helper', link: '/claude/agents/debug-helper' },
            { text: 'cc-manager', link: '/claude/agents/cc-manager' },
            { text: 'trust-checker', link: '/claude/agents/trust-checker' },
            { text: 'tag-agent', link: '/claude/agents/tag-agent' }
          ]
        }
      ],
      '/reference/': [
        {
          text: '레퍼런스',
          items: [
            { text: '설정 파일', link: '/reference/configuration' },
            { text: 'CLI 치트시트', link: '/reference/cli-cheatsheet' },
            { text: 'API 레퍼런스', link: '/reference/api-reference' },
            { text: 'Core 모듈', link: '/reference/core-modules' }
          ]
        }
      ],
      '/help/': [
        {
          text: '도움말',
          items: [
            { text: 'FAQ', link: '/help/faq' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/modu-ai/moai-adk' }
    ],

    footer: {
      message: 'MIT 라이선스로 배포됩니다.',
      copyright: 'Copyright © 2024-present MoAI Team | 문서: https://adk.mo.ai.kr | 커뮤니티: https://mo.ai.kr (오픈 예정)'
    },

    search: {
      provider: 'local'
    },

    editLink: {
      pattern: 'https://github.com/modu-ai/moai-adk/edit/main/docs/:path',
      text: 'GitHub에서 이 페이지 편집하기'
    }
  },

  head: [
    ['link', { rel: 'icon', type: 'image/png', href: '/moai-logo-dark.png' }],
    ['meta', { name: 'theme-color', content: '#5f67ee' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:locale', content: 'ko' }],
    ['meta', { property: 'og:title', content: 'MoAI-ADK | SPEC-First TDD Framework' }],
    ['meta', { property: 'og:site_name', content: 'MoAI-ADK' }],
    ['meta', { property: 'og:url', content: 'https://adk.mo.ai.kr/' }],
    ['meta', { property: 'og:description', content: 'TypeScript 기반 범용 언어 지원 SPEC-First TDD 개발 도구' }]
  ]
  })
)