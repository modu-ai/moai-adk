/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./docs/**/*.{md,vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Pretendard', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
        mono: ['CloudSansCode', 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'Consolas', 'Monaco', 'monospace'],
      },
    },
  },
  plugins: [require('daisyui')],
  daisyui: {
    themes: [
      {
        light: {
          // shadcn UI zinc 무채색 팔레트
          primary: "#18181b",      // zinc-900
          secondary: "#52525b",    // zinc-600
          accent: "#3f3f46",       // zinc-700
          neutral: "#71717a",      // zinc-500
          "base-100": "#fafafa",   // zinc-50
          "base-200": "#f4f4f5",   // zinc-100
          "base-300": "#e4e4e7",   // zinc-200
          info: "#52525b",         // zinc-600
          success: "#3f3f46",      // zinc-700
          warning: "#71717a",      // zinc-500
          error: "#27272a",        // zinc-800
        },
        dark: {
          // shadcn UI zinc 무채색 팔레트 (Dark)
          primary: "#e4e4e7",      // zinc-200
          secondary: "#a1a1aa",    // zinc-400
          accent: "#d4d4d8",       // zinc-300
          neutral: "#71717a",      // zinc-500
          "base-100": "#09090b",   // zinc-950
          "base-200": "#18181b",   // zinc-900
          "base-300": "#27272a",   // zinc-800
          info: "#a1a1aa",         // zinc-400
          success: "#d4d4d8",      // zinc-300
          warning: "#71717a",      // zinc-500
          error: "#52525b",        // zinc-600
        },
      },
    ],
    darkTheme: "dark",
    base: true,
    styled: true,
    utils: true,
  },
}
