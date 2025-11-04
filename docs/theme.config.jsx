/** @jsxImportSource react */
import React from 'react'

export default {
  // MoAI-ADK 브랜딩
  titleSuffix: ' – MoAI-ADK 문서',
  logo: <span style={{ fontWeight: 'bold', color: '#374151' }}>🎩 MoAI-ADK</span>,
  logoLink: 'https://github.com/modu-ai/moai-adk',

  // GitHub 연동
  project: {
    link: 'https://github.com/modu-ai/moai-adk',
  },
  docsRepositoryBase: 'https://github.com/modu-ai/moai-adk/tree/main/docs',

  // 기능 설정
  search: true,
  darkMode: true,
  nextThemes: true,
  defaultMenuCollapsed: false,

  // Mermaid 다이어그램 설정
  mermaid: {
    theme: 'neutral',
    themeVariables: {
      primaryColor: '#f3f4f6',
      primaryTextColor: '#1f2937',
      primaryBorderColor: '#9ca3af',
      lineColor: '#6b7280',
      secondaryColor: '#ffffff',
      tertiaryColor: '#f9fafb',
      background: 'transparent',
      mainBkg: '#ffffff',
      secondBkg: '#f9fafb',
      tertiaryBkg: '#f3f4f6',
    },
    flowchart: {
      useMaxWidth: true,
      htmlLabels: true,
      curve: 'basis'
    },
    gantt: {
      titleTopMargin: 25,
      barHeight: 20,
      fontSize: 11,
      fontFamily: 'Inter, sans-serif'
    }
  },

  // 네비게이션
  toc: {
    backToTop: true,
    extraContent: (
      <a
        href="https://github.com/modu-ai/moai-adk/discussions"
        target="_blank"
        rel="noopener noreferrer"
        style={{
          fontSize: '0.875rem',
          color: '#6b7280',
          textDecoration: 'none'
        }}
      >
        💬 피드백
      </a>
    )
  },

  // 편집 및 푸터
  editLink: {
    text: '이 페이지 수정하기 →'
  },

  footer: {
    text: (
      <div>
        <span>
          {`MIT License ${new Date().getFullYear()} © `}
          <a
            href="https://mo.ai.kr"
            target="_blank"
            rel="noopener noreferrer"
          >
            MoAI
          </a>
        </span>
        <div style={{ marginTop: '0.5rem', fontSize: '0.875rem', color: '#6b7280' }}>
          {'🤖 Generated with '}
          <a
            href="https://claude.com/claude-code"
            target="_blank"
            rel="noopener noreferrer"
          >
            Claude Code
          </a>
          {' & 🎩 Alfred'}
        </div>
      </div>
    )
  },

  // 색상 테마 (무채색)
  color: {
    hue: 220,  // 회색 계열
    saturation: 0,  // 채도 제거
    lightness: 0.5
  },

  // 사용자 정의 스타일
  head: (
    <>
      <meta
        name="viewport"
        content="width=device-width, initial-scale=1.0"
      />
      <meta
        name="description"
        content="MoAI-ADK - SPEC-First TDD 개발 프레임워크 (Alfred 슈퍼에이전트 포함)"
      />
      <meta
        name="keywords"
        content="MoAI, ADK, TDD, Alfred, 슈퍼에이전트, 개발 프레임워크"
      />
      <link
        rel="icon"
        href="/favicon.ico"
      />
    </>
  ),

  // i18n 설정 (한국어 기본)
  i18n: [
    { locale: 'ko', text: '한국어' },
    { locale: 'en', text: 'English' }
  ],

  // 기타 설정
  useNextSeoProps() {
    return {
      titleTemplate: '%s – MoAI-ADK 문서'
    }
  }
}