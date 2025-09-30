import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'MoAI-ADK',
  description: 'TypeScript 기반 SPEC 우선 TDD 개발 도구',
  lang: 'ko-KR',

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: '가이드', link: '/guide/workflow' },
      { text: '시작하기', link: '/getting-started/installation' },
      { text: 'CLI 명령어', link: '/cli/init' },
      {
        text: '링크',
        items: [
          { text: '공식 문서', link: 'https://adk.mo.ai.kr' },
          { text: '커뮤니티 (오픈 예정)', link: 'https://mo.ai.kr' },
          { text: 'GitHub', link: 'https://github.com/modu-ai/moai-adk' },
          { text: 'NPM', link: 'https://www.npmjs.com/package/moai-adk' }
        ]
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: '가이드',
          items: [
            { text: '3단계 워크플로우', link: '/guide/workflow' },
            { text: 'SPEC 우선 TDD', link: '/guide/spec-first-tdd' },
            { text: 'TAG 시스템', link: '/guide/tag-system' }
          ]
        }
      ],
      '/getting-started/': [
        {
          text: '시작하기',
          items: [
            { text: '설치', link: '/getting-started/installation' },
            { text: '빠른 시작', link: '/getting-started/quick-start' }
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
      pattern: 'https://github.com/modu-ai/moai-adk/edit/main/moai-adk-ts/docs/:path',
      text: 'GitHub에서 이 페이지 편집하기'
    }
  },

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
    ['meta', { name: 'theme-color', content: '#5f67ee' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:locale', content: 'ko' }],
    ['meta', { property: 'og:title', content: 'MoAI-ADK | SPEC-First TDD Framework' }],
    ['meta', { property: 'og:site_name', content: 'MoAI-ADK' }],
    ['meta', { property: 'og:url', content: 'https://adk.mo.ai.kr/' }]
  ]
})
