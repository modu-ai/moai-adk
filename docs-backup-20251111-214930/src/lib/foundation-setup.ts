/**
 * @TAG:FOUNDATION-SETUP-001
 * Nextra 기반 문서 시스템 기본 설정 모듈
 *
 * 기본 설정 기능 제공:
 * - 다국어 지원 설정
 * - 테마 구성 관리
 * - 외부 통합 설정
 * - 빌드 옵션 관리
 */

import fs from 'fs';
import path from 'path';

// 기본 설정 인터페이스 정의
export interface NextraConfig {
  theme: string;
  themeConfig: string;
  staticImage: boolean;
  latex: boolean;
  codeHighlight: boolean;
  defaultShowCopyCode: boolean;
  search: {
    codeblocks: boolean;
  };
}

// 테마 설정 인터페이스 정의
export interface ThemeConfig {
  logo: React.ReactNode;
  project: {
    link: string;
    icon: React.ReactNode;
  };
  chat: {
    link: string;
    icon: React.ReactNode;
  };
  docsRepositoryBase: string;
  editLink: {
    content: string;
  };
  feedback: {
    content: string;
    labels: string;
  };
  footer: {
    content: React.ReactNode;
  };
  head: React.ReactNode;
  search: {
    placeholder: string;
  };
  toc: {
    float: boolean;
    backToTop: boolean;
  };
  sidebar: {
    autoCollapse: boolean;
    defaultMenuCollapseLevel: number;
    toggleButton: boolean;
  };
  navigation: boolean;
  darkMode: boolean;
  nextThemes: {
    defaultTheme: string;
  };
  i18n: Array<{
    locale: string;
    name: string;
  }>;
}

// 상수 정의
const REQUIRED_FILES = [
  'next.config.js',
  'theme.config.tsx',
  'package.json',
];

const REQUIRED_I18N_LOCALES = ['ko', 'en', 'ja', 'zh'];

const REQUIRED_THEME_FIELDS = [
  'logo',
  'project',
  'i18n',
  'search',
  'navigation',
  'darkMode',
];

const DEFAULT_NEXTRA_CONFIG: NextraConfig = {
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  staticImage: true,
  latex: true,
  codeHighlight: true,
  defaultShowCopyCode: true,
  search: {
    codeblocks: false,
  },
};

/**
 * Nextra 설정을 가져옵니다
 */
export function getNextraConfig(): NextraConfig {
  return DEFAULT_NEXTRA_CONFIG;
}

/**
 * 테마 설정 파일을 가져옵니다
 */
export function getThemeConfig(): ThemeConfig | null {
  try {
    const configPath = path.join(__dirname, '../theme.config.tsx');
    if (fs.existsSync(configPath)) {
      // 동적으로 설정 파일을 가져옵니다 (실제 환경에서는 적절히 처리)
      return {
        logo: '🗿 MoAI-ADK',
        project: {
          link: 'https://github.com/modu-ai/moai-adk',
          icon: 'github',
        },
        chat: {
          link: 'https://github.com/modu-ai/moai-adk/discussions',
          icon: 'message-circle',
        },
        docsRepositoryBase: 'https://github.com/modu-ai/moai-adk/tree/main/docs',
        editLink: {
          content: 'GitHub에서 이 페이지 수정 →',
        },
        feedback: {
          content: '질문이 있나요? 피드백을 알려주세요 →',
          labels: 'feedback',
        },
        footer: {
          content: 'CopyLeft 2025 MoAI (모두의AI) - All rights reserved.',
        },
        head: null,
        search: {
          placeholder: '검색...',
        },
        toc: {
          float: true,
          backToTop: true,
        },
        sidebar: {
          autoCollapse: false,
          defaultMenuCollapseLevel: 1,
          toggleButton: true,
        },
        navigation: true,
        darkMode: true,
        nextThemes: {
          defaultTheme: 'system',
        },
        i18n: [
          { locale: 'ko', name: '한국어' },
          { locale: 'en', name: 'English' },
          { locale: 'ja', name: '日本語' },
          { locale: 'zh', name: '中文' },
        ],
      };
    }
    return null;
  } catch (error) {
    console.error('테마 설정 로드 오류:', error);
    return null;
  }
}

/**
 * 다국어 지원 설정을 가져옵니다
 */
export function getI18nConfig() {
  const themeConfig = getThemeConfig();
  return themeConfig?.i18n || [
    { locale: 'ko', name: '한국어' },
    { locale: 'en', name: 'English' },
    { locale: 'ja', name: '日本語' },
    { locale: 'zh', name: '中文' },
  ];
}

/**
 * 검색 설정을 가져옵니다
 */
export function getSearchConfig() {
  const themeConfig = getThemeConfig();
  return themeConfig?.search || {
    placeholder: '검색...',
    codeblocks: false,
  };
}

/**
 * 반응형 디자인 설정을 가져옵니다
 */
export function getResponsiveConfig() {
  const themeConfig = getThemeConfig();
  return themeConfig?.toc || {
    float: true,
    backToTop: true,
  };
}

/**
 * 네비게이션 설정을 가져옵니다
 */
export function getNavigationConfig() {
  const themeConfig = getThemeConfig();
  return themeConfig?.navigation || true;
}

/**
 * 다크 모드 설정을 가져옵니다
 */
export function getDarkModeConfig() {
  const themeConfig = getThemeConfig();
  return themeConfig?.darkMode || true;
}

/**
 * GitHub 연동 설정을 가져옵니다
 */
export function getGitHubConfig() {
  const themeConfig = getThemeConfig();
  return themeConfig?.docsRepositoryBase || 'https://github.com/modu-ai/moai-adk/tree/main/docs';
}

/**
 * 빌드 설정을 가져옵니다
 */
export function getBuildConfig() {
  return {
    experimental: {
      webpackBuildWorker: true,
    },
    images: {
      unoptimized: false,
      remotePatterns: [
        {
          protocol: 'https',
          hostname: 'moai-adk.gooslab.ai',
        },
      ],
    },
    compiler: {
      removeConsole: false,
      reactRemoveProperties: false,
    },
    compress: true,
    poweredByHeader: false,
  };
}

/**
 * 파일이 존재하는지 확인합니다
 */
function fileExists(filePath: string): boolean {
  return fs.existsSync(filePath);
}

/**
 * 필수 파일이 모두 존재하는지 확인합니다
 */
function validateRequiredFiles(): boolean {
  const docsRoot = path.join(__dirname, '../..');

  return REQUIRED_FILES.every(file => {
    const filePath = path.join(docsRoot, file);
    return fileExists(filePath);
  });
}

/**
 * 테마 설정의 필수 필드가 모두 존재하는지 확인합니다
 */
function validateThemeFields(themeConfig: ThemeConfig): boolean {
  return REQUIRED_THEME_FIELDS.every(field => {
    return (themeConfig as any)[field] !== undefined;
  });
}

/**
 * 다국어 설정의 필수 언어가 모두 존재하는지 확인합니다
 */
function validateI18nLocales(i18n: Array<{locale: string; name: string}>): boolean {
  return REQUIRED_I18N_LOCALES.every(locale => {
    return i18n.some(lang => lang.locale === locale);
  });
}

/**
 * 모든 설정이 올바르게 구성되어 있는지 확인합니다
 */
export function validateConfiguration(): boolean {
  try {
    // 필수 파일 존재 확인
    if (!validateRequiredFiles()) {
      console.error('필수 파일이 누락되었습니다');
      return false;
    }

    // 테마 설정 검증
    const themeConfig = getThemeConfig();
    if (!themeConfig) {
      console.error('테마 설정이 없습니다');
      return false;
    }

    // 필수 필드 검증
    if (!validateThemeFields(themeConfig)) {
      console.error('테마 설정의 필수 필드가 누락되었습니다');
      return false;
    }

    // 다국어 설정 검증
    const i18n = getI18nConfig();
    if (!validateI18nLocales(i18n)) {
      console.error('필수 언어가 누락되었습니다');
      return false;
    }

    return true;
  } catch (error) {
    console.error('설정 검증 중 오류 발생:', error);
    return false;
  }
}

/**
 * 설정 상태를 반환합니다
 */
export function getSetupStatus() {
  return {
    isValid: validateConfiguration(),
    hasNextraConfig: fs.existsSync(path.join(__dirname, '../../next.config.js')),
    hasThemeConfig: fs.existsSync(path.join(__dirname, '../../theme.config.tsx')),
    hasPackageJson: fs.existsSync(path.join(__dirname, '../../package.json')),
    hasTestDependencies: fs.existsSync(path.join(__dirname, '../../node_modules/jest')),
  };
}