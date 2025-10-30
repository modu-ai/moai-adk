// @CODE:NEXTRA-THEME-001 - Nextra theme configuration
// @CODE:NEXTRA-I18N-010 - Language switcher integration
import React from 'react';
import LanguageSwitcher from './components/LanguageSwitcher';

const config = {
  logo: <span>MoAI-ADK Documentation</span>,
  project: {
    link: 'https://github.com/GoosLab/moai-adk'
  },
  docsRepositoryBase: 'https://github.com/GoosLab/moai-adk/tree/main/docs-site',
  footer: {
    text: 'MoAI-ADK © 2025 GoosLab'
  },
  useNextSeoProps() {
    return {
      titleTemplate: '%s – MoAI-ADK'
    };
  },
  navbar: {
    extraContent: () => <LanguageSwitcher />
  }
};

export default config;
