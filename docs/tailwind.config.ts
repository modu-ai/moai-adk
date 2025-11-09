import type { Config } from 'tailwindcss'
import defaultTheme from 'tailwindcss/defaultTheme'

const config: Config = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './node_modules/nextra-theme-docs/dist/**/*.js',
  ],

  theme: {
    extend: {
      // Font families matching mkdocs setup
      fontFamily: {
        sans: ['var(--font-pretendard)', 'var(--font-inter)', ...defaultTheme.fontFamily.sans],
        mono: ['var(--font-code)', ...defaultTheme.fontFamily.mono],
      },

      // Colors extracted from mkdocs main.html
      colors: {
        // Light theme colors (default)
        light: {
          primary: {
            fg: '#000000',            // Black text
            'fg-light': '#333333',    // Dark gray variant
            'fg-dark': '#000000',     // Pure black
            bg: '#FFFFFF',            // White background
            'bg-light': '#F5F5F5',    // Very light gray
          },
          accent: {
            fg: '#666666',            // Medium gray
            'fg-transparent': 'rgba(102, 102, 102, 0.1)',
          },
          bg: {
            DEFAULT: '#FFFFFF',       // Main white background
            light: '#F9F9F9',         // Almost white
            dark: '#F0F0F0',          // Light gray
          },
          surface: {
            DEFAULT: '#F5F5F5',       // Light gray surface
            light: '#FBFBFB',         // Almost white surface
            dark: '#F0F0F0',          // Muted gray surface
          },
          text: {
            DEFAULT: '#000000',       // Pure black text
            secondary: '#666666',     // Medium gray text
            disabled: '#AAAAAA',      // Disabled text gray
          },
          border: {
            DEFAULT: '#DDDDDD',       // Light gray border
            light: '#EEEEEE',         // Very light gray border
            dark: '#CCCCCC',          // Medium gray border
          },
          code: {
            bg: '#F0F0F0',            // Light gray code background
            fg: '#000000',            // Black code text
            border: '#DDDDDD',        // Light gray code border
          },
        },

        // Dark theme colors (slate)
        dark: {
          primary: {
            fg: '#FFFFFF',            // White text
            'fg-light': '#EEEEEE',    // Light gray variant
            'fg-dark': '#FFFFFF',     // Pure white
            bg: '#121212',            // Dark background
            'bg-light': '#1E1E1E',    // Lighter dark background
          },
          accent: {
            fg: '#BBBBBB',            // Light gray accent
            'fg-transparent': 'rgba(187, 187, 187, 0.1)',
          },
          bg: {
            DEFAULT: '#121212',       // Main dark background
            light: '#1E1E1E',         // Slightly lighter dark
            dark: '#0A0A0A',          // Darker background
          },
          surface: {
            DEFAULT: '#1E1E1E',       // Dark surface
            light: '#2A2A2A',         // Lighter dark surface
            dark: '#0A0A0A',          // Darker surface
          },
          text: {
            DEFAULT: '#FFFFFF',       // White text
            secondary: '#BBBBBB',     // Light gray text
            disabled: '#777777',      // Disabled text gray
          },
          border: {
            DEFAULT: '#333333',       // Dark gray border
            light: '#444444',         // Lighter dark border
            dark: '#222222',          // Darker border
          },
          code: {
            bg: '#1E1E1E',            // Dark gray code background
            fg: '#FFFFFF',            // White code text
            border: '#333333',        // Dark code border
          },
        },
      },

      // Shadow definitions matching mkdocs
      boxShadow: {
        sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
        md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
        lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
        'dark-sm': '0 1px 2px 0 rgba(0, 0, 0, 0.3)',
        'dark-md': '0 4px 6px -1px rgba(0, 0, 0, 0.4), 0 2px 4px -1px rgba(0, 0, 0, 0.3)',
      },

      // Transitions matching mkdocs
      transitionDuration: {
        normal: '250ms',
        slow: '350ms',
      },

      // Border radius
      borderRadius: {
        md: '6px',
        lg: '8px',
        xl: '12px',
      },

      // Spacing adjustments for Korean typography
      lineHeight: {
        ko: '1.6',
        en: '1.5',
      },

      // Letter spacing for Korean
      letterSpacing: {
        ko: '-0.5px',
        en: '0',
      },
    },
  },

  darkMode: 'class',

  plugins: [],
}

export default config
