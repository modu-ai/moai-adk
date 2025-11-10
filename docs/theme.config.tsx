import React from 'react'

export default {
  logo: (
    <span
      style={{
        fontWeight: 700,
        fontSize: '1.2rem',
        display: 'flex',
        alignItems: 'center',
        gap: '8px'
      }}
      title="MoAI-ADK 홈페이지로 이동"
    >
      🗿 MoAI-ADK
    </span>
  ),

  project: {
    link: 'https://github.com/modu-ai/moai-adk',
  },

  docsRepositoryBase: 'https://github.com/modu-ai/moai-adk/tree/main/docs',

  // 기본 Nextra 테마 설정 (공식 문서 기반)
  titleSuffix: ' – MoAI-ADK',

  // 내비게이션 설정
  navigation: {
    prev: true,
    next: true,
  },

  // 검색 설정
  search: {
    placeholder: '문서 검색...',
    loading: '로딩 중...',
    error: '검색 오류',
    emptyResult: '검색 결과 없음',
  },

  // TOC 설정
  toc: {
    float: true,
    backToTop: '맨 위로',
  },

  // 사이드바 설정
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
  },

  // 다크모드 설정
  darkMode: true,

  // 푸터 설정
  footer: {
    text: `MIT ${new Date().getFullYear()} © GoosLab`,
  },

  // 편집 링크 설정
  editLink: {
    text: '이 페이지 편집하기 →',
  },

  // 피드백 설정
  feedback: {
    content: '이 페이지가 도움이 되었나요?',
    labels: 'feedback',
    useLink: () => 'https://github.com/modu-ai/moai-adk/issues/new?template=feedback.md',
  },

  // 다국어 지원
  i18n: [
    { locale: 'ko', name: '한국어', direction: 'ltr' },
    { locale: 'en', name: 'English', direction: 'ltr' },
    { locale: 'ja', name: '日本語', direction: 'ltr' },
    { locale: 'zh', name: '中文', direction: 'ltr' },
  ],

  // SEO 및 메타 태그 설정
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta name="theme-color" content="#000000" />
      <meta name="description" content="MoAI-ADK - AI 기반 SPEC-First TDD 개발 프레임워크" />
      <meta property="og:title" content="MoAI-ADK - AI 기반 SPEC-First TDD 개발 프레임워크" />
      <meta property="og:description" content="신뢰할 수 있고 유지보수하기 쉬운 소프트웨어를 AI의 도움으로 빌드하세요." />
      <meta property="og:type" content="website" />
      <meta property="og:url" content="https://moai-adk.gooslab.ai/" />
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:title" content="MoAI-ADK - AI 기반 SPEC-First TDD 개발 프레임워크" />
      <meta name="twitter:description" content="신뢰할 수 있고 유지보수하기 쉬운 소프트웨어를 AI의 도움으로 빌드하세요." />
      <meta name="keywords" content="AI, TDD, SPEC-First, 개발 프레임워크, 자동화, 테스트, 문서화, MoAI-ADK" />
      <meta name="author" content="GoosLab" />
      <meta name="robots" content="index, follow" />
    </>
  ),
}