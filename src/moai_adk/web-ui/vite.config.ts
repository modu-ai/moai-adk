import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 9005,
    proxy: {
      '/api': {
        target: 'http://localhost:9595',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:9595',
        ws: true,
      },
    },
  },
})
