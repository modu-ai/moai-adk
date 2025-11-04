module.exports = {
  // MoAI-ADK 브랜딩
  title: 'MoAI-ADK 문서',
  logo: '🎩 MoAI-ADK',
  logoLink: 'https://github.com/modu-ai/moai-adk',

  // GitHub 연동
  project: {
    link: 'https://github.com/modu-ai/moai-adk',
  },
  docsRepositoryBase: 'https://github.com/modu-ai/moai-adk/tree/main/docs',

  // 기능 설정
  search: true,
  darkMode: true,
  defaultMenuCollapsed: false,

  // 네비게이션
  toc: {
    backToTop: true
  },

  // 편집 및 푸터
  editLink: {
    text: '이 페이지 수정하기 →'
  },

  footer: {
    text: `MIT License ${new Date().getFullYear()} © MoAI`
  },

  // 색상 테마 (무채색)
  color: {
    hue: 220,  // 회색 계열
    saturation: 0,  // 채도 제거
    lightness: 0.5
  }
}