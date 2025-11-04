/** @jsxImportSource react */
import React from 'react'

const currentYear = new Date().getFullYear()

export default {
  titleSuffix: ' – MoAI-ADK 문서',
  logo: React.createElement('span', { style: { fontWeight: 'bold', color: '#374151' } }, '🎩 MoAI-ADK'),
  logoLink: 'https://github.com/modu-ai/moai-adk',

  project: {
    link: 'https://github.com/modu-ai/moai-adk',
  },
  docsRepositoryBase: 'https://github.com/modu-ai/moai-adk/tree/main/docs',

  search: true,
  darkMode: true,
  nextThemes: true,
  defaultMenuCollapsed: false,

  toc: {
    backToTop: true
  },

  editLink: {
    text: '이 페이지 수정하기 →'
  },

  footer: {
    text: React.createElement('div', {}, [
      React.createElement('span', {}, [
        `MIT License ${currentYear} © `,
        React.createElement('a', {
          href: 'https://mo.ai.kr',
          target: '_blank',
          rel: 'noopener noreferrer'
        }, 'MoAI')
      ]),
      React.createElement('div', {
        style: { marginTop: '0.5rem', fontSize: '0.875rem', color: '#6b7280' }
      }, [
        '🤖 Generated with ',
        React.createElement('a', {
          href: 'https://claude.com/claude-code',
          target: '_blank',
          rel: 'noopener noreferrer'
        }, 'Claude Code'),
        ' & 🎩 Alfred'
      ])
    ])
  },

  head: React.createElement(React.Fragment, {}, [
    React.createElement('meta', {
      name: 'viewport',
      content: 'width=device-width, initial-scale=1.0'
    }),
    React.createElement('meta', {
      name: 'description',
      content: 'MoAI-ADK - SPEC-First TDD 개발 프레임워크 (Alfred 슈퍼에이전트 포함)'
    }),
    React.createElement('meta', {
      name: 'keywords',
      content: 'MoAI, ADK, TDD, Alfred, 슈퍼에이전트, 개발 프레임워크'
    }),
    React.createElement('link', {
      rel: 'icon',
      href: '/favicon.ico'
    })
  ]),

  useNextSeoProps() {
    return {
      titleTemplate: '%s – MoAI-ADK 문서'
    }
  }
}
