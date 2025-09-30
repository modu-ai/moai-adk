// Custom theme extending default VitePress theme
import DefaultTheme from 'vitepress/theme'
import './custom.css'
import type { Theme } from 'vitepress'

export default {
  ...DefaultTheme,
  enhanceApp({ app }) {
    // Custom app enhancements can be added here
  }
} as Theme