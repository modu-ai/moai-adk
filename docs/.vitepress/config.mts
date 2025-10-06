// @CODE:DOCS-001 | SPEC: .moai/specs/SPEC-DOCS-001/spec.md | TEST: moai-adk-ts/__tests__/docs/
// VitePress Configuration - Phase 1 (5개 핵심 페이지)
// ESM 모듈 형식 (.mts) - VitePress 호환성 보장

import { defineConfig } from 'vitepress';

// https://vitepress.dev/reference/site-config
export default defineConfig({
  // 사이트 메타데이터
  title: 'MoAI-ADK',
  description: 'SPEC-First TDD Development Framework with Universal Language Support',
  lang: 'ko-KR',

  // 테마 설정
  themeConfig: {
    // 로고
    logo: '/alfred_logo.png',

    // 네비게이션 바
    nav: [
      { text: '가이드', link: '/guide/getting-started' },
      { text: '핵심 개념', link: '/concepts/spec-first-tdd' },
      { text: 'FAQ', link: '/guide/faq' },
    ],

    // Sidebar 구조 (Phase 1: 5개 핵심 페이지)
    sidebar: [
      {
        text: '시작하기',
        collapsed: false,
        items: [
          { text: 'Quick Start', link: '/guide/getting-started' },
          { text: 'MoAI-ADK란?', link: '/guide/what-is-moai-adk' },
          { text: 'FAQ', link: '/guide/faq' },
        ],
      },
      {
        text: '핵심 개념',
        collapsed: false,
        items: [
          { text: 'SPEC 우선 TDD', link: '/concepts/spec-first-tdd' },
        ],
      },
    ],

    // 검색 기능 (로컬 검색)
    search: {
      provider: 'local',
      options: {
        locales: {
          ko: {
            translations: {
              button: {
                buttonText: '검색',
                buttonAriaLabel: '검색',
              },
              modal: {
                noResultsText: '결과를 찾을 수 없습니다',
                resetButtonTitle: '초기화',
                footer: {
                  selectText: '선택',
                  navigateText: '이동',
                  closeText: '닫기',
                },
              },
            },
          },
        },
      },
    },

    // 소셜 링크
    socialLinks: [
      { icon: 'github', link: 'https://github.com/modu-ai/moai-adk' },
    ],

    // Footer
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025 MoAI Team',
    },

    // Edit Link
    editLink: {
      pattern: 'https://github.com/modu-ai/moai-adk/edit/main/docs/:path',
      text: 'GitHub에서 이 페이지 편집하기',
    },

    // Last Updated
    lastUpdated: {
      text: '최종 업데이트',
      formatOptions: {
        dateStyle: 'short',
        timeStyle: 'short',
      },
    },
  },

  // Markdown 설정
  markdown: {
    lineNumbers: true,
    theme: {
      light: 'github-light',
      dark: 'github-dark',
    },
  },

  // 빌드 설정
  outDir: '.vitepress/dist',
  cacheDir: '.vitepress/cache',
});
