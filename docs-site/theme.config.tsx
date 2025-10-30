// @CODE:NEXTRA-THEME-001 - Nextra theme configuration
import { DocsThemeConfig } from 'nextra-theme-docs';

const config: DocsThemeConfig = {
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
  }
};

export default config;
