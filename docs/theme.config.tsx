import React from 'react'
import { DocsThemeConfig } from 'nextra-theme-docs'

const config: DocsThemeConfig = {
  logo: (
    <span style={{ fontSize: '1.5rem', fontWeight: 'bold' }}>
      🗿 MoAI-ADK
    </span>
  ),
  project: {
    link: 'https://github.com/moai-ai/moai-adk',
  },
  chat: {
    link: 'https://discord.gg/moai-adk',
  },
  docsRepositoryBase: 'https://github.com/moai-ai/moai-adk/tree/main/docs',
  footer: {
    content: 'MIT License © 2025 MoAI-ADK Contributors',
  },
  toc: {
    backToTop: true,
    float: true,
    title: '이 페이지에서',
    extraContent: (
      <div style={{ marginTop: '1.5rem', paddingTop: '1rem', borderTop: '1px solid #ddd' }}>
        <p style={{ fontSize: '0.875rem', color: '#666' }}>
          📚 <a href="/worktree" style={{ color: '#0070f3', textDecoration: 'none' }}>
            Git Worktree CLI 문서 →
          </a>
        </p>
      </div>
    ),
  },
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
  },
  navbar: {
    extraContent: null,
  },
  search: {
    placeholder: '검색...',
  },
  editLink: {
    content: '이 페이지 수정 →',
  },
  feedback: {
    content: '질문이 있으신가요? 피드백을 남겨주세요!',
    useLink: () => 'https://github.com/moai-ai/moai-adk/issues/new',
  },
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta name="description" content="MoAI-ADK 온라인 문서 - 온라인 제작 인공지능 개발 도구" />
      <meta property="og:type" content="website" />
      <meta property="og:locale" content="ko_KR" />
      <meta property="og:url" content="https://docs.moai-ai.dev" />
      <meta property="og:site_name" content="MoAI-ADK Documentation" />
    </>
  ),
}

export default config
